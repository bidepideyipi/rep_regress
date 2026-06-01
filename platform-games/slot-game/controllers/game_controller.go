package controllers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/shopspring/decimal"

	"platform-games/slot-game/cache"
	"platform-games/slot-game/dao"
	"platform-games/slot-game/models"
	"platform-games/slot-game/nacos"
	"platform-games/slot-game/rocketmq"
	"platform-games/slot-game/rtp"

	"github.com/gin-gonic/gin"
)

// GameController 游戏控制器
type GameController struct {
	configManager  *nacos.ConfigManager
	gameInstance   *models.SlotGameGame
	mqProducer     *rocketmq.Producer
	userRTPService *rtp.UserRTPService
	redisCache     *cache.RedisCache
	db             *sql.DB
	userDAO        *dao.UserDAO
}

// NewGameController 创建游戏控制器
func NewGameController(configManager *nacos.ConfigManager) *GameController {
	return &GameController{
		configManager: configManager,
	}
}

// SetMQProducer 设置RocketMQ生产者
func (gc *GameController) SetMQProducer(producer *rocketmq.Producer) {
	gc.mqProducer = producer
}

// SetUserRTPService 设置UserRTP服务
func (gc *GameController) SetUserRTPService(service *rtp.UserRTPService) {
	gc.userRTPService = service
}

// SetRedisCache 设置Redis缓存
func (gc *GameController) SetRedisCache(redisCache *cache.RedisCache) {
	gc.redisCache = redisCache
}

// SetDB 设置数据库连接
func (gc *GameController) SetDB(db *sql.DB) {
	gc.db = db
	gc.userDAO = dao.NewUserDAO(db)
}

// InitializeGame 初始化游戏
func (gc *GameController) InitializeGame(config *models.GameConfig) {
	gc.gameInstance = models.NewSlotGame(config)
	log.Printf("游戏初始化完成: %s (RTP: %.2f%%)", config.Config.GameID, config.GameSettings.RTP)
}

// SpinRequest 旋转请求
type SpinRequest struct {
	IntegratorID string  `json:"integrator_id" binding:"required"`
	UserID       string  `json:"user_id" binding:"required"`
	BetAmount    float64 `json:"bet_amount" binding:"required,min=0.1"`
	BetLines     int     `json:"bet_lines" binding:"required,min=1,max=20"`
	SessionID    string  `json:"session_id" binding:"required"`
}

// SpinResponse 旋转响应
type SpinResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    struct {
		SessionID        string           `json:"session_id"`
		UserID           string           `json:"user_id"`
		GameID           string           `json:"game_id"`
		Balance          float64          `json:"balance"`
		BetAmount        float64          `json:"bet_amount"`
		BetLines         int              `json:"bet_lines"`
		BetPerLine       float64          `json:"bet_per_line"`
		WinAmount        float64          `json:"win_amount"`
		NetResult        float64          `json:"net_result"`
		IsFreeSpin       bool             `json:"is_free_spin"`
		ReelResult       [][]string       `json:"reel_result"`
		WinLines         []models.WinLine `json:"win_lines"`
		BonusFeature     string           `json:"bonus_feature"`
		ProcessingTimeMs int64            `json:"processing_time_ms"`
		Timestamp        time.Time        `json:"timestamp"`
	} `json:"data,omitempty"`
}

// Spin Spin接口
// @Summary 执行旋转
// @Description 执行游戏旋转操作
// @Tags Game
// @Accept json
// @Produce json
// @Param request body SpinRequest true "旋转请求"
// @Success 200 {object} SpinResponse
// @Failure 400 {object} SpinResponse
// @Failure 500 {object} SpinResponse
// @Router /api/game/spin [post]
func (gc *GameController) Spin(c *gin.Context) {
	startTime := time.Now()

	var req SpinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, SpinResponse{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查游戏实例是否初始化
	if gc.gameInstance == nil {
		c.JSON(http.StatusInternalServerError, SpinResponse{
			Success: false,
			Message: "游戏实例未初始化",
		})
		return
	}

	gameID := gc.configManager.GetConfig().Config.GameID

	// 检查Free Spin状态
	var remainingSpins int64
	if gc.redisCache != nil {
		var err error
		remainingSpins, err = gc.redisCache.GetFreeSpinRemaining(req.IntegratorID, req.UserID, gameID)
		if err != nil {
			log.Printf("获取Free Spin状态失败: integrator=%s, user=%s, game=%s, error=%v",
				req.IntegratorID, req.UserID, gameID, err)
		}
	}

	// Free Spin期间不扣减下注金额（remainingSpins > 0 即为free spin）
	isFreeSpin := remainingSpins > 0
	betAmount := req.BetAmount
	if isFreeSpin {
		betAmount = 0
	}

	// 检查 userDAO 是否初始化
	if gc.userDAO == nil {
		log.Printf("[警告] userDAO未初始化，余额操作将被跳过")
	}

	// 如果不是Free Spin，先扣减用户余额
	var userBeforeBet *models.UserInfo
	if !isFreeSpin && gc.userDAO != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		var err error
		userBeforeBet, err = gc.userDAO.DeductBalance(ctx, req.IntegratorID, req.UserID, req.BetAmount)
		if err != nil {
			log.Printf("扣减余额失败: integrator=%s, user=%s, amount=%.2f, error=%v",
				req.IntegratorID, req.UserID, req.BetAmount, err)
			c.JSON(http.StatusBadRequest, SpinResponse{
				Success: false,
				Message: "扣减余额失败: " + err.Error(),
			})
			return
		}
		log.Printf("余额扣减成功: integrator=%s, user=%s, 扣减=%.2f, 剩余=%.2f",
			req.IntegratorID, req.UserID, req.BetAmount, userBeforeBet.Balance)
	}

	// 构建游戏请求
	gameReq := &models.SpinRequest{
		UserID:    req.UserID,
		GameID:    gameID,
		BetAmount: betAmount,
		BetLines:  req.BetLines,
		SessionID: req.SessionID,
	}

	// 检查RTP并调整符号权重（如果需要）
	adjustedWeights := gc.checkAndAdjustRTPWeights(req.UserID)
	if adjustedWeights != nil {
		// 设置自定义权重
		gc.gameInstance.ClearCustomWeights()
		for reelIndex, weights := range adjustedWeights {
			gc.gameInstance.SetCustomWeights(reelIndex, weights)
		}
		defer gc.gameInstance.ClearCustomWeights()
	}

	// 执行旋转
	result, err := gc.gameInstance.Spin(gameReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, SpinResponse{
			Success: false,
			Message: "游戏执行失败: " + err.Error(),
		})
		return
	}

	// 计算处理时间
	processingTime := time.Since(startTime).Milliseconds()

	// 处理Free Spin逻辑
	var totalFreeSpins int64

	if gc.redisCache != nil && result.FreeSpinInfo != nil {
		// 如果触发了新的Free Spin
		if result.FreeSpinInfo.TriggeredCount > 0 {
			triggeredCount := result.FreeSpinInfo.TriggeredCount

			// 设置TTL为7天（604800秒）
			ttl := 7 * 24 * time.Hour

			if remainingSpins > 0 {
				// 已经在Free Spin中，累加次数
				newRemainingSpins, err := gc.redisCache.AddFreeSpinRemaining(
					req.IntegratorID, req.UserID, gameID, triggeredCount, ttl)
				if err != nil {
					log.Printf("累加Free Spin失败: %v", err)
				} else {
					remainingSpins = newRemainingSpins
				}
				log.Printf("[Free Spin] 重触发: integrator=%s, user=%s, game=%s, 增加=%d, 剩余=%d",
					req.IntegratorID, req.UserID, gameID, triggeredCount, remainingSpins)
			} else {
				// 首次触发Free Spin
				err := gc.redisCache.SetFreeSpinRemaining(
					req.IntegratorID, req.UserID, gameID, triggeredCount, ttl)
				if err != nil {
					log.Printf("设置Free Spin失败: %v", err)
				} else {
					remainingSpins = triggeredCount
				}
				log.Printf("[Free Spin] 首次触发: integrator=%s, user=%s, game=%s, 次数=%d",
					req.IntegratorID, req.UserID, gameID, triggeredCount)
			}
			totalFreeSpins = triggeredCount
		}

		// 如果当前是Free Spin，扣减一次
		if remainingSpins > 0 {
			newRemaining, err := gc.redisCache.DecrementFreeSpin(
				req.IntegratorID, req.UserID, gameID)
			if err != nil {
				log.Printf("扣减Free Spin失败: %v", err)
			} else {
				remainingSpins = newRemaining
			}
		}
	}

		// 如果有中奖金额，增加用户余额
		if result.TotalWin > 0 && gc.userDAO != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			_, err := gc.userDAO.AddBalance(ctx, req.IntegratorID, req.UserID, result.TotalWin)
			if err != nil {
				log.Printf("增加余额失败: integrator=%s, user=%s, amount=%.2f, error=%v",
					req.IntegratorID, req.UserID, result.TotalWin, err)
				// 中奖增加失败不影响响应，记录日志即可
			} else {
				log.Printf("余额增加成功: integrator=%s, user=%s, 增加=%.2f",
					req.IntegratorID, req.UserID, result.TotalWin)
			}
		}

		// 获取用户最新余额用于响应
		var userBalance float64
		if gc.userDAO != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			log.Printf("[查询余额] integrator_id=%s, user_id=%s", req.IntegratorID, req.UserID)
			user, err := gc.userDAO.GetUserByIntegratorAndID(ctx, req.IntegratorID, req.UserID)
			if err != nil {
				log.Printf("获取用户余额失败: %v", err)
			} else if user != nil {
				userBalance = user.Balance
				log.Printf("[查询余额成功] balance=%.2f", userBalance)
			} else {
				log.Printf("[查询余额] 用户不存在")
			}
		}

		// 构建响应
		bonusFeature := ""
	if totalFreeSpins > 0 {
		bonusFeature = fmt.Sprintf("free_spins_%d", totalFreeSpins)
	}

	response := SpinResponse{
		Success: true,
	}
	response.Data.SessionID = req.SessionID
	response.Data.UserID = req.UserID
	response.Data.GameID = gameReq.GameID
	response.Data.Balance = userBalance
	response.Data.BetAmount = gameReq.BetAmount
	response.Data.BetLines = gameReq.BetLines
	response.Data.BetPerLine = gameReq.BetAmount / float64(gameReq.BetLines)
	response.Data.WinAmount = result.TotalWin
	response.Data.NetResult = result.TotalWin - gameReq.BetAmount
	response.Data.IsFreeSpin = remainingSpins > 0
	response.Data.ReelResult = result.ReelResult
	response.Data.WinLines = result.WinLines
	response.Data.BonusFeature = bonusFeature
	response.Data.ProcessingTimeMs = processingTime
	response.Data.Timestamp = time.Now()

	// 获取用户RTP数据（异步，不阻塞响应）
	if gc.userRTPService != nil {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			userRTP, err := gc.userRTPService.GetUserRTP(ctx, req.UserID)
			if err != nil {
				log.Printf("获取用户RTP失败: user_id=%s, error=%v", req.UserID, err)
				return
			}

			log.Printf("用户RTP数据: user_id=%s, rtp=%.4f, total_bet=%.2f, total_win=%.2f, net_result=%.2f, total_spins=%d, avg_bet=%.2f",
				req.UserID, userRTP.RTP, userRTP.TotalBet, userRTP.TotalWin, userRTP.NetResult, userRTP.TotalSpins, userRTP.AvgBet)
		}()
	}

	c.JSON(http.StatusOK, response)

	// 异步写入ClickHouse日志
	if gc.mqProducer != nil {
		go gc.writeGameLog(gameReq, result, processingTime, remainingSpins > 0, bonusFeature)
	}
}

// writeGameLog 写入游戏日志到ClickHouse
func (gc *GameController) writeGameLog(gameReq *models.SpinRequest, result *models.SpinResult, processingTime int64, isFreeSpin bool, bonusFeature string) {
	// 生成日志ID
	logID := fmt.Sprintf("log_%d_%s", time.Now().UnixNano(), gameReq.UserID)

	// 获取客户端信息
	clientIP := gc.getClientIP()

	// 转换WinLines到ClickHouse格式 (Tuple数组)
	winLinesTuples := make([][]interface{}, len(result.WinLines))
	for i, wl := range result.WinLines {
		symbols := make([]string, len(wl.Positions))
		// 将位置转换为符号
		for j, pos := range wl.Positions {
			reelIdx := pos / 3
			rowIdx := pos % 3
			if reelIdx < len(result.ReelResult) && rowIdx < len(result.ReelResult[reelIdx]) {
				symbols[j] = result.ReelResult[reelIdx][rowIdx]
			}
		}
		isWild := uint8(0)
		if wl.IsWild {
			isWild = 1
		}
		// Tuple格式: [LineID, WinAmount, Symbols, IsWild]
		winLinesTuples[i] = []interface{}{
			uint16(wl.LineID),
			decimal.NewFromFloat(wl.WinAmount),
			symbols,
			isWild,
		}
	}

	// 序列化游戏结果
	gameResultJSON, _ := json.Marshal(map[string]interface{}{
		"total_win":     result.TotalWin,
		"bonus_feature": bonusFeature,
		"win_count":     len(result.WinLines),
	})

	// 判断是否免费旋转
	isFreeSpinUint := uint8(0)
	if isFreeSpin {
		isFreeSpinUint = 1
	}

	logEntry := models.GameLogDetail{
		LogID:          logID,
		GameSessionID:  gameReq.SessionID,
		IntegratorID:   "default", // TODO: 从请求或配置中获取
		UserID:         gameReq.UserID,
		GameID:         gameReq.GameID,
		BetAmount:      decimal.NewFromFloat(gameReq.BetAmount),
		WinAmount:      decimal.NewFromFloat(result.TotalWin),
		NetResult:      decimal.NewFromFloat(result.TotalWin - gameReq.BetAmount),
		BetLines:       uint16(gameReq.BetLines),
		BetPerLine:     decimal.NewFromFloat(gameReq.BetAmount / float64(gameReq.BetLines)),
		IsFreeSpin:     isFreeSpinUint,
		BonusFeature:   bonusFeature,
		DeviceType:     "unknown", // TODO: 从请求头获取
		DeviceOS:       "unknown",
		BrowserType:    "unknown",
		IPAddress:      clientIP,
		IPRegion:       "",
		IPCountry:      "",
		SessionID:      gameReq.SessionID,
		ServerID:       gc.getLocalIP(),
		ProcessingTime: uint32(processingTime),
		ErrorCode:      0,
		ErrorMessage:   "",
		GameResultJSON: string(gameResultJSON),
		ReelResult:     result.ReelResult,
		WinLines:       winLinesTuples,
		UserAgent:      "", // TODO: 从请求头获取
		LogTime:        time.Now(),
	}

	if err := gc.mqProducer.SendGameLogAsync(logEntry); err != nil {
		log.Printf("写入ClickHouse失败: %v", err)
	}
}

// getClientIP 获取客户端IP（转换为uint32）
func (gc *GameController) getClientIP() uint32 {
	// TODO: 从gin.Context获取客户端IP
	// 简化版本：返回0表示未知
	return 0
}

// getLocalIP 获取本机IP地址
func (gc *GameController) getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return "127.0.0.1"
}

// GetConfig 获取游戏配置信息
func (gc *GameController) GetConfig(c *gin.Context) {
	// 获取路径参数中的 game_id
	gameID := c.Param("game_id")
	if gameID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "缺少游戏ID参数",
		})
		return
	}

	config := gc.configManager.GetConfig()
	if config == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "配置未加载",
		})
		return
	}

	// 验证请求的游戏ID与当前配置是否匹配
	if config.Config.GameID != gameID {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "游戏配置不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gc.configManager.GetConfigInfo(),
	})
}

// HealthCheck 健康检查
func (gc *GameController) HealthCheck(c *gin.Context) {
	health := gin.H{
		"status":  "healthy",
		"time":    time.Now(),
		"service": "slot-game",
	}

	if gc.gameInstance != nil {
		health["game_initialized"] = true
		health["game_id"] = gc.configManager.GetConfig().Config.GameID
	} else {
		health["game_initialized"] = false
	}

	c.JSON(http.StatusOK, health)
}

// RefreshConfig 手动刷新配置
func (gc *GameController) RefreshConfig(c *gin.Context) {
	config, err := gc.configManager.LoadConfig()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "刷新配置失败: " + err.Error(),
		})
		return
	}

	gc.InitializeGame(config)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "配置刷新成功",
		"data":    gc.configManager.GetConfigInfo(),
	})
}

// checkAndAdjustRTPWeights 检查RTP并调整符号权重
// 返回调整后的权重map，如果不需要调整则返回nil
func (gc *GameController) checkAndAdjustRTPWeights(userID string) map[int][]models.SymbolWeight {
	if gc.userRTPService == nil {
		log.Printf("[RTP控制] userRTPService未初始化，跳过RTP控制")
		return nil
	}

	// 同步获取用户RTP数据（短超时）
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	userRTP, err := gc.userRTPService.GetUserRTP(ctx, userID)
	if err != nil {
		log.Printf("[RTP控制] 获取用户RTP失败（跳过RTP控制）: user_id=%s, error=%v", userID, err)
		return nil
	}

	// 获取配置的RTP阈值
	configRTP := gc.configManager.GetConfig().GameSettings.RTP
	maxBet := gc.configManager.GetConfig().GameSettings.MaxBet

	// log.Printf("[RTP控制] user_id=%s, total_bet=%.2f, rtp=%.4f, config_rtp=%.4f",
	// 	userID, userRTP.TotalBet, userRTP.RTP, configRTP)

	// 检查条件：totalBet > 500 且 rtp >= 配置的RTP
	if (userRTP.TotalBet > maxBet && userRTP.RTP*100 >= configRTP) {
		log.Printf("[RTP控制] ✓ 触发RTP控制: user_id=%s, total_bet=%.2f > 500, rtp=%.4f >= %.4f",
			userID, userRTP.TotalBet, userRTP.RTP*100, configRTP)

		// 调整符号权重
		return gc.adjustWeights()
	}

	return nil
}

// adjustWeights 调整符号权重
// - plum、grape、watermelon、bell的weight降低一半
// - seven、wild、scatter的weight设为0
func (gc *GameController) adjustWeights() map[int][]models.SymbolWeight {
	config := gc.configManager.GetConfig()
	adjustedWeights := make(map[int][]models.SymbolWeight)

	// 需要降低权重的符号（降低一半）
	reduceHalfSymbols := map[string]bool{
		"plum":       true,
		"grape":      true,
		"watermelon": true,
		"bell":       true,
	}

	// 需要移除的符号（weight设为0）
	removeSymbols := map[string]bool{
		"seven":   true,
		"wild":    true,
		"scatter": true,
	}

	log.Printf("[RTP控制] 开始调整权重 - 移除符号: %v, 降低权重符号: %v",
		[]string{"seven", "wild", "scatter"}, []string{"plum", "grape", "watermelon", "bell"})

	// 遍历所有卷轴
	for _, reel := range config.Reels {
		reelIndex := reel.ReelIndex
		var newWeights []models.SymbolWeight

		// 处理每个符号的权重
		for _, sw := range reel.SymbolWeights {
			symbolID := sw.SymbolID
			weight := sw.Weight

			// 检查是否需要移除
			if removeSymbols[symbolID] {
				log.Printf("[RTP控制] 卷轴%d: 移除符号 %s (原权重=%d)", reelIndex, symbolID, weight)
				continue // 跳过，即weight设为0
			}

			// 检查是否需要降低一半
			if reduceHalfSymbols[symbolID] {
				weight = weight / 2
				if weight < 1 {
					weight = 1
				}
				log.Printf("[RTP控制] 卷轴%d: 降低符号 %s 权重 %d -> %d", reelIndex, symbolID, sw.Weight, weight)
			}

			newWeights = append(newWeights, models.SymbolWeight{
				SymbolID: symbolID,
				Weight:   weight,
			})
		}

		adjustedWeights[reelIndex] = newWeights
		log.Printf("[RTP控制] 卷轴%d权重调整完成: %d个符号", reelIndex, len(newWeights))
	}

	return adjustedWeights
}

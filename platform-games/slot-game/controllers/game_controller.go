package controllers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"platform-games/slot-game/models"
	"platform-games/slot-game/nacos"
)

// GameController 游戏控制器
type GameController struct {
	configManager *nacos.ConfigManager
	gameInstance  *models.SlotGameGame
}

// NewGameController 创建游戏控制器
func NewGameController(configManager *nacos.ConfigManager) *GameController {
	return &GameController{
		configManager: configManager,
	}
}

// InitializeGame 初始化游戏
func (gc *GameController) InitializeGame(config *models.GameConfig) {
	gc.gameInstance = models.NewSlotGame(config)
	log.Printf("游戏初始化完成: %s (RTP: %.2f%%)", config.Config.GameID, config.GameSettings.RTP)
}

// SpinRequest 旋转请求
type SpinRequest struct {
	UserID    string  `json:"user_id" binding:"required"`
	BetAmount float64 `json:"bet_amount" binding:"required,min=0.1"`
	BetLines  int     `json:"bet_lines" binding:"required,min=1,max=20"`
	SessionID string  `json:"session_id" binding:"required"`
}

// SpinResponse 旋转响应
type SpinResponse struct {
	Success bool `json:"success"`
	Message string `json:"message,omitempty"`
	Data    struct {
		SessionID        string              `json:"session_id"`
		UserID           string              `json:"user_id"`
		GameID           string              `json:"game_id"`
		BetAmount        float64             `json:"bet_amount"`
		BetLines         int                 `json:"bet_lines"`
		BetPerLine       float64             `json:"bet_per_line"`
		WinAmount        float64             `json:"win_amount"`
		NetResult        float64             `json:"net_result"`
		IsFreeSpin       bool                `json:"is_free_spin"`
		ReelResult       [][]string          `json:"reel_result"`
		WinLines         []models.WinLine    `json:"win_lines"`
		BonusFeature     string              `json:"bonus_feature"`
		ProcessingTimeMs int64               `json:"processing_time_ms"`
		Timestamp        time.Time           `json:"timestamp"`
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
	
	// 构建游戏请求
	gameReq := &models.SpinRequest{
		UserID:    req.UserID,
		GameID:    gc.configManager.GetConfig().Config.GameID,
		BetAmount: req.BetAmount,
		BetLines:  req.BetLines,
		SessionID: req.SessionID,
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
	
	// 构建响应
	// 判断是否为免费旋转：基于 BonusFeature 是否以 free_spins 开头
	isFreeSpin := strings.HasPrefix(result.BonusFeature, "free_spins")

	response := SpinResponse{
		Success: true,
		Data: struct {
			SessionID        string           `json:"session_id"`
			UserID           string           `json:"user_id"`
			GameID           string           `json:"game_id"`
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
		}{
			SessionID:        req.SessionID,
			UserID:           req.UserID,
			GameID:           gameReq.GameID,
			BetAmount:        gameReq.BetAmount,
			BetLines:         gameReq.BetLines,
			BetPerLine:       gameReq.BetAmount / float64(gameReq.BetLines),
			WinAmount:        result.TotalWin,
			NetResult:        result.TotalWin - gameReq.BetAmount,
			IsFreeSpin:       isFreeSpin,
			ReelResult:       result.ReelResult,
			WinLines:         result.WinLines,
			BonusFeature:     result.BonusFeature,
			ProcessingTimeMs: processingTime,
			Timestamp:        time.Now(),
		},
	}
	
	c.JSON(http.StatusOK, response)
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
package models

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

// SlotGameGame 老虎机游戏逻辑
type SlotGameGame struct {
	config       *GameConfig
	rand         *rand.Rand
	customWeights map[int][]SymbolWeight // 自定义符号权重（按卷轴索引）
}

// NewSlotGame 创建新游戏实例
func NewSlotGame(config *GameConfig) *SlotGameGame {
	return &SlotGameGame{
		config: config,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
		customWeights: make(map[int][]SymbolWeight),
	}
}

// SetCustomWeights 设置自定义符号权重
func (game *SlotGameGame) SetCustomWeights(reelIndex int, weights []SymbolWeight) {
	game.customWeights[reelIndex] = weights
}

// ClearCustomWeights 清除自定义权重
func (game *SlotGameGame) ClearCustomWeights() {
	game.customWeights = make(map[int][]SymbolWeight)
}

// Spin 执行旋转
func (game *SlotGameGame) Spin(request *SpinRequest) (*SpinResult, error) {
	// 验证下注金额
	if request.BetAmount < game.config.GameSettings.MinBet {
		return nil, fmt.Errorf("下注金额低于最小值 %.2f", game.config.GameSettings.MinBet)
	}
	if request.BetAmount > game.config.GameSettings.MaxBet {
		return nil, fmt.Errorf("下注金额超过最大值 %.2f", game.config.GameSettings.MaxBet)
	}
	if request.BetLines > game.config.PayTable.PayLineCount {
		return nil, fmt.Errorf("下注线数超过最大值 %d", game.config.PayTable.PayLineCount)
	}

	// 生成卷轴结果
	reelResult := game.generateReels()

	// 检查中奖线路
	winLines := game.checkWinLines(reelResult, request.BetLines)

	// 计算总奖金
	totalWin := game.calculateTotalWin(winLines, request.BetAmount, request.BetLines)

	// 检查特殊功能（返回触发的free spin次数）
	triggeredFreeSpins := game.checkBonusFeatures(reelResult)

	return &SpinResult{
		ReelResult:     reelResult,
		WinLines:       winLines,
		TotalWin:       totalWin,
		BonusFeature:   "", // 由controller层设置
		FreeSpinInfo:   &FreeSpinInfo{TriggeredCount: triggeredFreeSpins},
	}, nil
}

// generateReels 生成卷轴结果
func (game *SlotGameGame) generateReels() [][]string {
	result := make([][]string, 3)
	for i := 1; i <= 3; i++ {
		reel := game.config.GetReelByIndex(i)
		if reel != nil {
			// 检查是否有自定义权重
			if customWeights, ok := game.customWeights[i]; ok {
				result[i-1] = game.generateReelWithWeights(customWeights)
			} else {
				result[i-1] = reel.GenerateReelSymbols(game.config)
			}
		} else {
			// 默认配置
			result[i-1] = []string{"cherry", "lemon", "orange"}
		}
	}
	return result
}

// generateReelWithWeights 使用指定权重生成卷轴符号
func (game *SlotGameGame) generateReelWithWeights(weights []SymbolWeight) []string {
	// 计算总权重
	totalWeight := 0
	for _, sw := range weights {
		totalWeight += sw.Weight
	}

	if totalWeight == 0 {
		return []string{"cherry", "lemon", "orange"}
	}

	// 创建权重累积表
	cumulativeWeights := make([]struct {
		SymbolID  string
		MaxWeight int
	}, len(weights))

	currentWeight := 0
	for i, sw := range weights {
		currentWeight += sw.Weight
		cumulativeWeights[i].SymbolID = sw.SymbolID
		cumulativeWeights[i].MaxWeight = currentWeight
	}

	// 生成3个位置的符号
	result := make([]string, 3)
	for pos := 0; pos < 3; pos++ {
		randWeight := game.rand.Intn(totalWeight)
		for _, cw := range cumulativeWeights {
			if randWeight < cw.MaxWeight {
				result[pos] = cw.SymbolID
				break
			}
		}
	}

	return result
}

// checkWinLines 检查中奖线路
func (game *SlotGameGame) checkWinLines(reelResult [][]string, betLines int) []WinLine {
	var winLines []WinLine

	// 检查每条中奖线路
	for lineID := 0; lineID < betLines; lineID++ {
		if lineID >= len(game.config.PayTable.PayLinePatterns) {
			break
		}

		pattern := game.config.PayTable.PayLinePatterns[lineID]
		winLine := game.checkSingleLine(reelResult, pattern, lineID+1)
		if winLine != nil {
			winLines = append(winLines, *winLine)
		}
	}

	return winLines
}

// checkSingleLine 检查单条线路
func (game *SlotGameGame) checkSingleLine(reelResult [][]string, pattern [][2]int, lineID int) *WinLine {
	var symbols []string
	var positions []int

	// 获取线路上的符号
	for _, pos := range pattern {
		if pos[0] < len(reelResult) && pos[1] < len(reelResult[pos[0]]) {
			symbols = append(symbols, reelResult[pos[0]][pos[1]])
			positions = append(positions, pos[0]*3+pos[1])
		}
	}

	if len(symbols) == 0 {
		return nil
	}

	// 处理万能符号，选择最高赔付的匹配
	matchingSymbol, matchCount, isWild := game.processWildSymbols(symbols)
	if matchingSymbol == "" {
		return nil
	}

	// 查找赔付倍数
	symbol := game.config.GetSymbolBy(matchingSymbol)
	if symbol == nil {
		return nil
	}

	multiplier := symbol.GetSymbolMultiplier(matchCount)
	if multiplier == nil {
		return nil
	}

	return &WinLine{
		LineID:     lineID,
		SymbolID:   matchingSymbol,
		MatchCount: matchCount,
		Multiplier: multiplier.Multiplier,
		WinAmount:  0, // 在外层计算
		Positions:  positions,
		IsWild:     isWild,
	}
}

// processWildSymbols 处理万能符号，返回能产生最高赔付的符号和匹配数
func (game *SlotGameGame) processWildSymbols(symbols []string) (string, int, bool) {
	if len(symbols) == 0 {
		return "", 0, false
	}

	// 统计非wild符号的数量
	symbolCounts := make(map[string]int)
	wildCount := 0

	for _, symbol := range symbols {
		if game.config.IsWildSymbol(symbol) {
			wildCount++
		} else {
			symbolCounts[symbol]++
		}
	}

	// 如果没有Wild，检查是否所有符号相同
	if wildCount == 0 {
		// 所有符号必须相同才算匹配
		firstSymbol := symbols[0]
		allSame := true
		for _, s := range symbols {
			if s != firstSymbol {
				allSame = false
				break
			}
		}
		if allSame {
			return firstSymbol, len(symbols), false
		}
		return "", 0, false
	}

	// 有Wild的情况：选择赔付最高的符号
	var bestMatch string
	var bestCount int
	var bestMultiplier float64

	for symbol, count := range symbolCounts {
		totalMatch := count + wildCount
		if totalMatch < 2 {
			continue
		}

		// 获取该符号在这个匹配数的赔付倍数
		sym := game.config.GetSymbolBy(symbol)
		if sym == nil {
			continue
		}

		multiplier := sym.GetSymbolMultiplier(totalMatch)
		if multiplier == nil {
			continue
		}

		// 选择赔付倍数最高的
		if multiplier.Multiplier > bestMultiplier {
			bestMatch = symbol
			bestCount = totalMatch
			bestMultiplier = multiplier.Multiplier
		}
	}

	// 如果没有找到合适的匹配，尝试只有wild的情况
	if bestMatch == "" && wildCount >= 2 {
		// 纯wild的情况，选择wild自己的赔付
		return "wild", wildCount, true
	}

	if bestMatch == "" {
		return "", 0, false
	}

	return bestMatch, bestCount, true
}

// calculateTotalWin 计算总奖金
func (game *SlotGameGame) calculateTotalWin(winLines []WinLine, betAmount float64, betLines int) float64 {
	betPerLine := betAmount / float64(betLines)
	var totalWin float64

	for i := range winLines {
		winLine := &winLines[i]
		symbol := game.config.GetSymbolBy(winLine.SymbolID)
		if symbol != nil {
			multiplier := symbol.GetSymbolMultiplier(winLine.MatchCount)
			if multiplier != nil && multiplier.IsBetLine {
				winLine.WinAmount = betPerLine * multiplier.Multiplier
				totalWin += winLine.WinAmount
			}
		}
	}

	return math.Round(totalWin*100) / 100 // 保留两位小数
}

// checkBonusFeatures 检查特殊功能，返回触发的free spin次数
func (game *SlotGameGame) checkBonusFeatures(reelResult [][]string) int64 {
	// 检查Scatter符号
	scatterCount := 0
	for _, reel := range reelResult {
		for _, symbol := range reel {
			if game.config.IsScatterSymbol(symbol) {
				scatterCount++
			}
		}
	}

	if scatterCount >= 3 {
		symbol := game.config.GetSymbolBy("scatter")
		if symbol != nil {
			for _, prop := range symbol.SpecialProperties {
				if prop.PropertyName == "free_spins" {
					// 解析free_spins数量
					var freeSpins int64
					fmt.Sscanf(prop.PropertyValue, "%d", &freeSpins)
					return freeSpins
				}
			}
		}
	}

	return 0
}

// CalculateRTP 计算RTP
func (game *SlotGameGame) CalculateRTP(totalBet float64, totalWin float64) float64 {
	if totalBet == 0 {
		return 0
	}
	rtp := (totalWin / totalBet) * 100
	return math.Round(rtp*100) / 100
}

// GetRandomSymbol 随机获取符号（用于测试）
func (game *SlotGameGame) GetRandomSymbol() string {
	if len(game.config.Symbols) == 0 {
		return "cherry"
	}
	return game.config.Symbols[game.rand.Intn(len(game.config.Symbols))].SymbolID
}

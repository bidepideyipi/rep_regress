package models

import (
	"math/rand"
	"time"
)

// GameConfig 游戏配置结构体
type GameConfig struct {
	Config struct {
		GameID      string    `json:"game_id"`
		GameName    string    `json:"game_name"`
		Version     string    `json:"version"`
		Description string    `json:"description"`
		ConfigType  string    `json:"config_type"`
		LastUpdated time.Time `json:"last_updated"`
	} `json:"config"`

	Symbols []Symbol `json:"symbols"`
	Reels   []Reel   `json:"reels"`

	PayTable struct {
		PayLineCount    int         `json:"pay_line_count"`
		PayLinePatterns [][][2]int `json:"pay_line_patterns"`
	} `json:"pay_table"`

	GameSettings struct {
		MinBet              float64 `json:"min_bet"`
		MaxBet              float64 `json:"max_bet"`
		DefaultBet          float64 `json:"default_bet"`
		Currency            string  `json:"currency"`
		Volatility          string  `json:"volatility"`
		RTP                 float64 `json:"rtp"`
		JackpotEnabled      bool    `json:"jackpot_enabled"`
		AutoSpinAvailable   bool    `json:"auto_spin_available"`
		MaxAutoSpin         int     `json:"max_auto_spin"`
	} `json:"game_settings"`

	CacheSettings struct {
		EnableLocalCache      bool `json:"enable_local_cache"`
		CacheTTLSeconds       int  `json:"cache_ttl_seconds"`
		RefreshIntervalSeconds int `json:"refresh_interval_seconds"`
		VersionCheckEnabled   bool `json:"version_check_enabled"`
	} `json:"cache_settings"`
}

// Symbol 符号配置
type Symbol struct {
	SymbolID          string                    `json:"symbol_id"`
	SymbolName        string                    `json:"symbol_name"`
	SymbolType        string                    `json:"symbol_type"`
	SortOrder         int                       `json:"sort_order"`
	IsActive          bool                      `json:"is_active"`
	Description       string                    `json:"description"`
	Multipliers       []SymbolMultiplier        `json:"multipliers"`
	SpecialProperties []SymbolSpecialProperty   `json:"special_properties"`
}

// SymbolMultiplier 符号赔付倍数
type SymbolMultiplier struct {
	MatchCount  int     `json:"match_count"`
	Multiplier  float64 `json:"multiplier"`
	IsBetLine   bool    `json:"is_bet_line"`
}

// SymbolSpecialProperty 符号特殊属性
type SymbolSpecialProperty struct {
	PropertyName  string `json:"property_name"`
	PropertyValue string `json:"property_value"`
}

// Reel 卷轴配置
type Reel struct {
	ReelIndex     int              `json:"reel_index"`
	ReelName      string           `json:"reel_name"`
	SymbolWeights []SymbolWeight   `json:"symbol_weights"`
}

// SymbolWeight 符号权重
type SymbolWeight struct {
	SymbolID string `json:"symbol_id"`
	Weight   int    `json:"weight"`
}

// SpinRequest 旋转请求
type SpinRequest struct {
	UserID      string  `json:"user_id" binding:"required"`
	GameID      string  `json:"game_id" binding:"required"`
	BetAmount   float64 `json:"bet_amount" binding:"required,min=0.1"`
	BetLines    int     `json:"bet_lines" binding:"required,min=1,max=20"`
	SessionID   string  `json:"session_id" binding:"required"`
	IsFreeSpin  bool    `json:"is_free_spin"`
}

// SpinResponse 旋转响应
type SpinResponse struct {
	SessionID     string              `json:"session_id"`
	UserID        string              `json:"user_id"`
	GameID        string              `json:"game_id"`
	BetAmount     float64             `json:"bet_amount"`
	BetLines      int                 `json:"bet_lines"`
	BetPerLine    float64             `json:"bet_per_line"`
	WinAmount     float64             `json:"win_amount"`
	NetResult     float64             `json:"net_result"`
	IsFreeSpin    bool                `json:"is_free_spin"`
	ReelResult    [][]string          `json:"reel_result"`
	WinLines      []WinLine           `json:"win_lines"`
	BonusFeature  string              `json:"bonus_feature"`
	RTPRate       float64             `json:"rtp_rate"`
	ProcessingTimeMs int64            `json:"processing_time_ms"`
	Timestamp     time.Time           `json:"timestamp"`
}

// WinLine 中奖线路
type WinLine struct {
	LineID      int      `json:"line_id"`
	SymbolID    string   `json:"symbol_id"`
	MatchCount  int      `json:"match_count"`
	WinAmount   float64  `json:"win_amount"`
	Multiplier  float64  `json:"multiplier"`
	Positions   []int    `json:"positions"`
	IsWild      bool     `json:"is_wild"`
}

// SpinResult 旋转结果
type SpinResult struct {
	ReelResult  [][]string
	WinLines    []WinLine
	TotalWin    float64
	BonusFeature string
}

// GetSymbolBy 根据ID获取符号
func (config *GameConfig) GetSymbolBy(symbolID string) *Symbol {
	for i := range config.Symbols {
		if config.Symbols[i].SymbolID == symbolID {
			return &config.Symbols[i]
		}
	}
	return nil
}

// GetSymbolMultiplier 获取符号赔付倍数
func (symbol *Symbol) GetSymbolMultiplier(matchCount int) *SymbolMultiplier {
	for i := range symbol.Multipliers {
		if symbol.Multipliers[i].MatchCount == matchCount {
			return &symbol.Multipliers[i]
		}
	}
	return nil
}

// GetReelByIndex 根据索引获取卷轴
func (config *GameConfig) GetReelByIndex(index int) *Reel {
	for i := range config.Reels {
		if config.Reels[i].ReelIndex == index {
			return &config.Reels[i]
		}
	}
	return nil
}

// GenerateReelSymbols 生成卷轴符号
func (reel *Reel) GenerateReelSymbols(config *GameConfig) []string {
	// 计算总权重
	totalWeight := 0
	for _, sw := range reel.SymbolWeights {
		totalWeight += sw.Weight
	}

	// 创建权重累积表
	cumulativeWeights := make([]struct {
		SymbolID string
		MaxWeight int
	}, len(reel.SymbolWeights))

	currentWeight := 0
	for i, sw := range reel.SymbolWeights {
		currentWeight += sw.Weight
		cumulativeWeights[i].SymbolID = sw.SymbolID
		cumulativeWeights[i].MaxWeight = currentWeight
	}

	// 生成3个位置的符号
	result := make([]string, 3)
	for pos := 0; pos < 3; pos++ {
		randWeight := rand.Intn(totalWeight)
		for _, cw := range cumulativeWeights {
			if randWeight < cw.MaxWeight {
				result[pos] = cw.SymbolID
				break
			}
		}
	}

	return result
}

// IsWildSymbol 检查是否为万能符号
func (config *GameConfig) IsWildSymbol(symbolID string) bool {
	symbol := config.GetSymbolBy(symbolID)
	if symbol == nil {
		return false
	}
	return symbol.SymbolType == "wild"
}

// IsScatterSymbol 检查是否为散布符号
func (config *GameConfig) IsScatterSymbol(symbolID string) bool {
	symbol := config.GetSymbolBy(symbolID)
	if symbol == nil {
		return false
	}
	return symbol.SymbolType == "scatter"
}

// GetSubstituteSymbols 获取可替代的符号类型
func (config *GameConfig) GetSubstituteSymbols(wildSymbol *Symbol) []string {
	var substituteTypes []string
	for _, prop := range wildSymbol.SpecialProperties {
		if prop.PropertyName == "substitute" {
			substituteTypes = append(substituteTypes, prop.PropertyValue)
		}
	}
	return substituteTypes
}
package rtp

import (
	"context"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/shopspring/decimal"
)

// UserRTPMetrics 用户 RTP 指标
type UserRTPMetrics struct {
	UserID        string          `json:"user_id"`
	GameID        string          `json:"game_id"`
	IntegratorID  string          `json:"integrator_id"`
	TimeWindow    time.Time       `json:"time_window"`
	TotalBet      decimal.Decimal `json:"total_bet"`
	TotalWin      decimal.Decimal `json:"total_win"`
	NetResult     decimal.Decimal `json:"net_result"`
	RTP           decimal.Decimal `json:"rtp"`
	SpinCount     uint32          `json:"spin_count"`
	AvgBet        decimal.Decimal `json:"avg_bet"`
	MaxBet        decimal.Decimal `json:"max_bet"`
	MaxWin        decimal.Decimal `json:"max_win"`
	MinWin        decimal.Decimal `json:"min_win"`
	WinRate       decimal.Decimal `json:"win_rate"`
	WinSpinCount  uint32          `json:"win_spin_count"`
	LossSpinCount uint32          `json:"loss_spin_count"`
	FirstSpinTime time.Time       `json:"first_spin_time"`
	LastSpinTime  time.Time       `json:"last_spin_time"`
	LastUpdate    time.Time       `json:"last_update"`
}

// GameRTPMetrics 游戏 RTP 指标
type GameRTPMetrics struct {
	GameID        string          `json:"game_id"`
	IntegratorID  string          `json:"integrator_id"`
	TimeWindow    time.Time       `json:"time_window"`
	TotalBet      decimal.Decimal `json:"total_bet"`
	TotalWin      decimal.Decimal `json:"total_win"`
	NetResult     decimal.Decimal `json:"net_result"`
	RTP           decimal.Decimal `json:"rtp"`
	SpinCount     uint32          `json:"spin_count"`
	ActiveUsers   uint32          `json:"active_users"`
	AvgBet        decimal.Decimal `json:"avg_bet"`
	MaxBet        decimal.Decimal `json:"max_bet"`
	MaxWin        decimal.Decimal `json:"max_win"`
	Volatility    decimal.Decimal `json:"volatility"`
	WinRate       decimal.Decimal `json:"win_rate"`
	AvgNetResult  decimal.Decimal `json:"avg_net_result"`
	LastUpdate    time.Time       `json:"last_update"`
}

// RTPAlert RTP 异常告警
type RTPAlert struct {
	AlertID      string          `json:"alert_id"`
	AlertType    string          `json:"alert_type"`
	Severity     string          `json:"severity"`
	UserID       string          `json:"user_id"`
	GameID       string          `json:"game_id"`
	IntegratorID string          `json:"integrator_id"`
	RTP          decimal.Decimal `json:"rtp"`
	TotalBet     decimal.Decimal `json:"total_bet"`
	TotalWin     decimal.Decimal `json:"total_win"`
	SpinCount    uint32          `json:"spin_count"`
	TimeWindow   time.Time       `json:"time_window"`
	DetectedTime time.Time       `json:"detected_time"`
	ResolvedTime *time.Time      `json:"resolved_time"`
	Status       string          `json:"status"`
	Notes        string          `json:"notes"`
}

// RTPService RTP 查询服务
type RTPService struct {
	conn clickhouse.Conn
}

// NewRTPService 创建 RTP 查询服务
func NewRTPService(host string, port int, username, password, database string) (*RTPService, error) {
	dsn := fmt.Sprintf("%s:%d", host, port)
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{dsn},
		Auth: clickhouse.Auth{
			Database: database,
			Username: username,
			Password: password,
		},
		DialTimeout: 10 * time.Second,
	})
	if err != nil {
		return nil, fmt.Errorf("连接ClickHouse失败: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5 * time.Second)
	defer cancel()
	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, fmt.Errorf("ClickHouse ping失败: %w", err)
	}

	return &RTPService{conn: conn}, nil
}

// Close 关闭连接
func (s *RTPService) Close() error {
	return s.conn.Close()
}

// GetUserRealtimeRTP 获取用户实时 RTP（最近5分钟）
func (s *RTPService) GetUserRealtimeRTP(ctx context.Context, userID, gameID string, minutes int) ([]UserRTPMetrics, error) {
	query := `
		SELECT
			user_id, game_id, integrator_id, time_window,
			total_bet, total_win, net_result, rtp,
			spin_count, avg_bet, max_bet, max_win, min_win,
			win_rate, win_spin_count, loss_spin_count,
			first_spin_time, last_spin_time, last_update
		FROM rtp_user_realtime
		WHERE user_id = @user_id
		  AND game_id = @game_id
		  AND time_window >= now() - INTERVAL @minutes MINUTE
		ORDER BY time_window DESC
	`

	var metrics []UserRTPMetrics
	rows, err := s.conn.Query(ctx, query, clickhouse.Named("user_id", userID), clickhouse.Named("game_id", gameID), clickhouse.Named("minutes", minutes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m UserRTPMetrics
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// GetUserHourlyRTP 获取用户小时级 RTP
func (s *RTPService) GetUserHourlyRTP(ctx context.Context, userID, gameID string, hours int) ([]UserRTPMetrics, error) {
	query := `
		SELECT
			user_id, game_id, integrator_id, time_window,
			total_bet, total_win, net_result, rtp,
			spin_count, avg_bet, max_bet, max_win, min_win,
			win_rate, win_spin_count, loss_spin_count,
			first_spin_time, last_spin_time, last_update
		FROM rtp_user_metrics
		WHERE user_id = @user_id
		  AND game_id = @game_id
		  AND time_window >= now() - INTERVAL @hours HOUR
		ORDER BY time_window DESC
	`

	var metrics []UserRTPMetrics
	rows, err := s.conn.Query(ctx, query, clickhouse.Named("user_id", userID), clickhouse.Named("game_id", gameID), clickhouse.Named("hours", hours))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m UserRTPMetrics
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// GetGameRealtimeRTP 获取游戏实时 RTP
func (s *RTPService) GetGameRealtimeRTP(ctx context.Context, gameID string, minutes int) ([]GameRTPMetrics, error) {
	query := `
		SELECT
			game_id, integrator_id, time_window,
			total_bet, total_win, net_result, rtp,
			spin_count, active_users, avg_bet, max_bet, max_win,
			volatility, win_rate, avg_net_result, last_update
		FROM rtp_game_realtime
		WHERE game_id = @game_id
		  AND time_window >= now() - INTERVAL @minutes MINUTE
		ORDER BY time_window DESC
	`

	var metrics []GameRTPMetrics
	rows, err := s.conn.Query(ctx, query, clickhouse.Named("game_id", gameID), clickhouse.Named("minutes", minutes))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m GameRTPMetrics
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// GetGameHourlyRTP 获取游戏小时级 RTP
func (s *RTPService) GetGameHourlyRTP(ctx context.Context, gameID string, hours int) ([]GameRTPMetrics, error) {
	query := `
		SELECT
			game_id, integrator_id, time_window,
			total_bet, total_win, net_result, rtp,
			spin_count, active_users, avg_bet, max_bet, max_win,
			volatility, win_rate, avg_net_result, last_update
		FROM rtp_game_metrics
		WHERE game_id = @game_id
		  AND time_window >= now() - INTERVAL @hours HOUR
		ORDER BY time_window DESC
	`

	var metrics []GameRTPMetrics
	rows, err := s.conn.Query(ctx, query, clickhouse.Named("game_id", gameID), clickhouse.Named("hours", hours))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m GameRTPMetrics
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// GetActiveAlerts 获取当前活跃告警
func (s *RTPService) GetActiveAlerts(ctx context.Context, limit int) ([]RTPAlert, error) {
	query := `
		SELECT
			alert_id, alert_type, severity,
			user_id, game_id, integrator_id,
			rtp, total_bet, total_win, spin_count,
			time_window, detected_time, resolved_time,
			status, notes
		FROM rtp_alerts
		WHERE status = 'active'
		ORDER BY detected_time DESC
		LIMIT @limit
	`

	var alerts []RTPAlert
	rows, err := s.conn.Query(ctx, query, clickhouse.Named("limit", limit))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var a RTPAlert
		if err := rows.ScanStruct(&a); err != nil {
			return nil, err
		}
		alerts = append(alerts, a)
	}

	return alerts, rows.Err()
}

// DetectAbnormalRTP 手动检测异常 RTP
func (s *RTPService) DetectAbnormalRTP(ctx context.Context, highThreshold, lowThreshold float64, minSpins uint32) ([]UserRTPMetrics, error) {
	query := `
		SELECT
			user_id, game_id, integrator_id, time_window,
			total_bet, total_win, net_result, rtp,
			spin_count, avg_bet, max_bet, max_win, min_win,
			win_rate, win_spin_count, loss_spin_count,
			first_spin_time, last_spin_time, last_update
		FROM rtp_user_realtime
		WHERE (rtp > @high_threshold OR rtp < @low_threshold)
		  AND spin_count >= @min_spins
		  AND time_window >= now() - INTERVAL 1 HOUR
		ORDER BY rtp DESC
	`

	var metrics []UserRTPMetrics
	rows, err := s.conn.Query(ctx, query,
		clickhouse.Named("high_threshold", highThreshold),
		clickhouse.Named("low_threshold", lowThreshold),
		clickhouse.Named("min_spins", minSpins),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m UserRTPMetrics
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// GetTopHighRTPUsers 获取高 RTP 用户排行
func (s *RTPService) GetTopHighRTPUsers(ctx context.Context, hours int, minSpins uint32, limit int) ([]UserRTPMetrics, error) {
	query := `
		SELECT
			user_id, game_id, integrator_id,
			sum(total_bet) AS total_bet,
			sum(total_win) AS total_win,
			if(sum(total_bet) > 0, round(sum(total_win) / sum(total_bet), 4), 0) AS rtp,
			sum(spin_count) AS spin_count
		FROM rtp_user_realtime_local
		WHERE time_window >= now() - INTERVAL @hours HOUR
		GROUP BY user_id, game_id, integrator_id
		HAVING spin_count >= @min_spins
		ORDER BY rtp DESC
		LIMIT @limit
	`

	var metrics []UserRTPMetrics
	rows, err := s.conn.Query(ctx, query,
		clickhouse.Named("hours", hours),
		clickhouse.Named("min_spins", minSpins),
		clickhouse.Named("limit", limit),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var m UserRTPMetrics
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, rows.Err()
}

// ResolveAlert 解决告警
func (s *RTPService) ResolveAlert(ctx context.Context, alertID string, notes string) error {
	query := `
		ALTER TABLE rtp_alerts_local
		UPDATE 
			status = 'resolved',
			resolved_time = now(),
			notes = concat(notes, ' | Resolved: ', @notes)
		WHERE alert_id = @alert_id
	`

	return s.conn.Exec(ctx, query, clickhouse.Named("alert_id", alertID), clickhouse.Named("notes", notes))
}
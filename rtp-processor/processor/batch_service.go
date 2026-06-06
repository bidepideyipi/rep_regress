package processor

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/rtp-processor/config"
	"github.com/rtp-processor/db"
)

// BatchService 批处理服务
type BatchService struct {
	db           *db.ClickHouseWriter
	cfg          *config.Config
	userInterval time.Duration
	gameInterval time.Duration
	alertInterval time.Duration
	mu           sync.Mutex
	isRunning    bool
	stopCh       chan struct{}
	wg           sync.WaitGroup
}

/**
 * @brief 创建新的批处理服务实例
 * @param db ClickHouse Writer
 * @param cfg 配置
 * @return *BatchService
 */
func NewBatchService(writer interface{}, cfg *config.Config) *BatchService {
	dbWriter, ok := writer.(*db.ClickHouseWriter)
	if !ok {
		log.Printf("[BatchService] 警告: 未使用 ClickHouseWriter")
	}
	return &BatchService{
		db:           dbWriter,
		cfg:          cfg,
		userInterval: cfg.GetAggregateUserInterval(),
		gameInterval: cfg.GetAggregateGameInterval(),
		alertInterval: cfg.GetAlertInterval(),
		stopCh:       make(chan struct{}),
	}
}

/**
 * @brief 启动定时批处理服务
 */
func (s *BatchService) Start() {
	s.mu.Lock()
	if s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.run()
	log.Println("[BatchService] 定时批处理服务已启动")
}

/**
 * @brief 停止定时批处理服务
 */
func (s *BatchService) Stop() {
	s.mu.Lock()
	if !s.isRunning {
		s.mu.Unlock()
		return
	}
	s.isRunning = false
	s.mu.Unlock()

	close(s.stopCh)
	s.wg.Wait()
	log.Println("[BatchService] 定时批处理服务已停止")
}

/**
 * @brief 执行聚合批次
 */
func (s *BatchService) run() {
	defer s.wg.Done()
	userTicker := time.NewTicker(s.userInterval)
	gameTicker := time.NewTicker(s.gameInterval)
	alertTicker := time.NewTicker(s.alertInterval)
	defer userTicker.Stop()
	defer gameTicker.Stop()
	defer alertTicker.Stop()

	for {
		select {
		case <-s.stopCh:
			return
		case <-userTicker.C:
			s.executeUserAggregation()
		case <-gameTicker.C:
			s.executeGameAggregation()
		case <-alertTicker.C:
			s.executeAlertDetection()
		}
	}
}

/**
 * @brief 执行聚合批次
 */
func (s *BatchService) executeUserAggregation() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startTime := time.Now()
	log.Println("[BatchService] 开始用户聚合批次...")

	if err := s.aggregateUserRealtime(ctx); err != nil {
		log.Printf("[BatchService] 用户实时聚合失败: %v", err)
	} else {
		log.Printf("[BatchService] 用户实时聚合完成")
	}

	log.Printf("[BatchService] 用户聚合批次完成，耗时: %v", time.Since(startTime))
}

/**
 * @brief 执行游戏聚合批次
 */
func (s *BatchService) executeGameAggregation() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	startTime := time.Now()
	log.Println("[BatchService] 开始游戏聚合批次...")

	if err := s.aggregateGameRealtime(ctx); err != nil {
		log.Printf("[BatchService] 游戏实时聚合失败: %v", err)
	} else {
		log.Printf("[BatchService] 游戏实时聚合完成")
	}

	log.Printf("[BatchService] 游戏聚合批次完成，耗时: %v", time.Since(startTime))
}

/**
 * @brief 执行用户实时聚合
 */
func (s *BatchService) aggregateUserRealtime(ctx context.Context) error {
	sql := `
		INSERT INTO rtp_analytics.rtp_user_realtime_local
		SELECT
			user_id,
			integrator_id,
			game_id,
			toStartOfFiveMinutes(log_time) AS time_window,
			sum(bet_amount) AS total_bet,
			sum(win_amount) AS total_win,
			sum(win_amount - bet_amount) AS net_result,
			if(sum(bet_amount) > 0, round(sum(win_amount) / sum(bet_amount), 4), 0) AS rtp,
			count() AS spin_count,
			avg(bet_amount) AS avg_bet,
			max(win_amount) AS max_win,
			now() AS last_update
		FROM rtp_analytics.game_log_detail_local
		WHERE error_code = 0
		  AND log_time >= now() - INTERVAL 5 MINUTE
		GROUP BY
			user_id,
			integrator_id,
			game_id,
			toStartOfFiveMinutes(log_time)
		`
	return s.db.Exec(ctx, sql)
}

/**
 * @brief 执行游戏实时聚合
 */
func (s *BatchService) aggregateGameRealtime(ctx context.Context) error {
	sql := `
		INSERT INTO rtp_analytics.rtp_game_realtime_local
		SELECT
			game_id,
			integrator_id,
			toStartOfFiveMinutes(log_time) AS time_window,
			sum(bet_amount) AS total_bet,
			sum(win_amount) AS total_win,
			if(sum(bet_amount) > 0, round(sum(win_amount) / sum(bet_amount), 4), 0) AS rtp,
			count() AS spin_count,
			uniqExact(user_id) AS active_users,
			avg(bet_amount) AS avg_bet,
			now() AS last_update
		FROM rtp_analytics.game_log_detail_local
		WHERE error_code = 0
		  AND log_time >= now() - INTERVAL 5 MINUTE
		GROUP BY
			game_id,
			integrator_id,
			toStartOfFiveMinutes(log_time)
		`
	return s.db.Exec(ctx, sql)
}

/**
 * @brief 执行告警检测批次
 */
func (s *BatchService) executeAlertDetection() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	startTime := time.Now()
	log.Println("[BatchService] 开始告警检测批次...")

	if err := s.detectHighRTPAlert(ctx); err != nil {
		log.Printf("[BatchService] 高RTP告警检测失败: %v", err)
	} else {
		log.Printf("[BatchService] 高RTP告警检测完成")
	}

	if err := s.detectLowRTPAlert(ctx); err != nil {
		log.Printf("[BatchService] 低RTP告警检测失败: %v", err)
	} else {
		log.Printf("[BatchService] 低RTP告警检测完成")
	}

	if err := s.detectGameAbnormalAlert(ctx); err != nil {
		log.Printf("[BatchService] 游戏异常告警检测失败: %v", err)
	} else {
		log.Printf("[BatchService] 游戏异常告警检测完成")
	}

	log.Printf("[BatchService] 告警检测批次完成，耗时: %v", time.Since(startTime))
}

/**
 * @brief 执行高RTP告警检测
 */
func (s *BatchService) detectHighRTPAlert(ctx context.Context) error {
	sql := `
		INSERT INTO rtp_analytics.rtp_alerts_local
		SELECT
			concat('alert_', toString(now()), '_', user_id, '_', game_id) AS alert_id,
			'user_high_rtp' AS alert_type,
			'high' AS severity,
			user_id,
			game_id,
			integrator_id,
			rtp,
			total_bet,
			total_win,
			spin_count,
			time_window,
			now() AS detected_time,
			CAST(NULL AS Nullable(DateTime)) AS resolved_time,
			'active' AS status,
			concat('用户 RTP 过高: ', toString(rtp), ' (阈值: 0.98)') AS notes
		FROM rtp_analytics.rtp_user_realtime_local
		WHERE rtp > 0.98
		  AND spin_count >= 20
		  AND total_bet >= 100
		  AND time_window >= now() - INTERVAL 10 MINUTE
		`
	return s.db.Exec(ctx, sql)
}

/**
 * @brief 执行低RTP告警检测
 */
func (s *BatchService) detectLowRTPAlert(ctx context.Context) error {
	sql := `
		INSERT INTO rtp_analytics.rtp_alerts_local
		SELECT
			concat('alert_', toString(now()), '_', user_id, '_', game_id) AS alert_id,
			'user_low_rtp' AS alert_type,
			'medium' AS severity,
			user_id,
			game_id,
			integrator_id,
			rtp,
			total_bet,
			total_win,
			spin_count,
			time_window,
			now() AS detected_time,
			CAST(NULL AS Nullable(DateTime)) AS resolved_time,
			'active' AS status,
			concat('用户 RTP 过低: ', toString(rtp), ' (阈值: 0.50)') AS notes
		FROM rtp_analytics.rtp_user_realtime_local
		WHERE rtp < 0.50
		  AND spin_count >= 20
		  AND total_bet >= 100
		  AND time_window >= now() - INTERVAL 10 MINUTE
		`
	return s.db.Exec(ctx, sql)
}

/**
 * @brief 执行游戏异常告警检测
 */
func (s *BatchService) detectGameAbnormalAlert(ctx context.Context) error {
	sql := `
		INSERT INTO rtp_analytics.rtp_alerts_local
		SELECT
			concat('alert_', toString(now()), '_', game_id) AS alert_id,
			'game_abnormal' AS alert_type,
			'high' AS severity,
			'' AS user_id,
			game_id,
			integrator_id,
			rtp,
			total_bet,
			total_win,
			spin_count,
			time_window,
			now() AS detected_time,
			CAST(NULL AS Nullable(DateTime)) AS resolved_time,
			'active' AS status,
			concat('游戏整体 RTP 异常: ', toString(rtp), ' (预期: 0.95)') AS notes
		FROM rtp_analytics.rtp_game_realtime_local
		WHERE (rtp > 1.05 OR rtp < 0.85)
		  AND spin_count >= 100
		  AND time_window >= now() - INTERVAL 10 MINUTE
		`
	return s.db.Exec(ctx, sql)
}

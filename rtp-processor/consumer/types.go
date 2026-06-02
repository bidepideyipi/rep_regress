package consumer

import (
	"sync"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/rtp-processor/config"
	"github.com/rtp-processor/models"
)

// ClickHouseWriter ClickHouse批量写入器
type ClickHouseWriter struct {
	conn       clickhouse.Conn
	buffer     []models.GameLogDetail
	mutex      sync.Mutex
	batchSize  int
	flushCount int64
	errorCount int64
	tableName  string
}

// GameConsumer 游戏消费者
type GameConsumer struct {
	appConfig      *config.Config
	chWriter       *ClickHouseWriter
	rocketConsumer rocketmq.PullConsumer
	stopChan       chan struct{}
	offsets        map[int64]int64 // 队列 ID -> offset
	offsetMutex    sync.RWMutex
}

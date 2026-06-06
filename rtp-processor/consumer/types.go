package consumer

import (
	"sync"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/rtp-processor/config"
	"github.com/rtp-processor/db"
	"github.com/rtp-processor/jackpot"
)

// GameConsumer 游戏消费者
type GameConsumer struct {
	appConfig      *config.Config
	chWriter       *db.ClickHouseWriter
	mysqlManager   *db.MySQLManager
	jackpotManager *jackpot.Manager
	rocketConsumer rocketmq.PullConsumer
	stopChan       chan struct{}
	offsets        map[int64]int64
	offsetMutex    sync.RWMutex
}

// PushConsumer Push模式消费者（并发批量消费）
type PushConsumer struct {
	appConfig      *config.Config
	rocketConsumer rocketmq.PushConsumer
	stopChan       chan struct{}
	listener       *MessageListener
	chWriter       *db.ClickHouseWriter
	mysqlManager   *db.MySQLManager
	jackpotManager *jackpot.Manager
}

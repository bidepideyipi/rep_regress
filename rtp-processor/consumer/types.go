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
	appConfig       *config.Config
	chWriter        *db.ClickHouseWriter
	mysqlManager    *db.MySQLManager
	jackpotManager  *jackpot.Manager
	rocketConsumer  rocketmq.PullConsumer
	stopChan        chan struct{}
	offsets         map[int64]int64
	offsetMutex     sync.RWMutex
}

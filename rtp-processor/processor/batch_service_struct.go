package processor

import (
	"sync"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rtp-processor/config"
)

type BatchService struct {
	conn          clickhouse.Conn
	cfg           *config.Config
	userInterval  time.Duration
	gameInterval  time.Duration
	alertInterval time.Duration
	stopCh        chan struct{}
	wg            sync.WaitGroup
	isRunning     bool
	mu            sync.Mutex
}

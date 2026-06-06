//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/rtp-processor/config"
	"github.com/rtp-processor/consumer"
	"github.com/rtp-processor/db"
	"github.com/rtp-processor/jackpot"
)

// App 应用程序（由Wire注入）
type App struct {
	Config       *config.Config
	GameConsumer *consumer.GameConsumer
	PushConsumer *consumer.PushConsumer
	CHWriter     *db.ClickHouseWriter
	MySQLManager *db.MySQLManager
	JackpotMgr   *jackpot.Manager
}

// InitializeApp 初始化应用程序（Wire生成代码）
func InitializeApp(cfg *config.Config) (*App, error) {
	wire.Build(
		// 基础设施
		db.NewClickHouseWriter,
		db.NewMySQLManager,

		// 业务组件
		jackpot.NewManager,

		// 消费者
		consumer.NewGameConsumer,
		consumer.NewPushConsumer,

		// App 结构体绑定
		wire.Struct(new(App), "*"),
	)
	return &App{}, nil
}

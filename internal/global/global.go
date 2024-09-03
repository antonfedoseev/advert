package global

import (
	"github.com/go-logr/logr"
	"internal/message_broker"
	"internal/redisdb"
	"internal/settings"
	"pkg/db"
)

type Hub struct {
	ExPath     string
	Settings   settings.Settings
	Logger     logr.Logger
	AppName    string
	Db         *db.DB
	RedisDb    *redisdb.Pool
	MbProducer *message_broker.Producer
}

func (g *Hub) Dispose() {
	if g.Db != nil {
		g.Db.Dispose()
	}
}

func New(exPath string, settings settings.Settings, logger logr.Logger, appName string,
	mbProducer *message_broker.Producer) Hub {
	return Hub{
		ExPath:     exPath,
		Settings:   settings,
		Logger:     logger,
		Db:         db.New(settings.DBs),
		AppName:    appName,
		MbProducer: mbProducer,
	}
}

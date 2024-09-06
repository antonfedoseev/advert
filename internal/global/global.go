package global

import (
	"github.com/go-logr/logr"
	"internal/settings"
	"pkg/db"
	"pkg/mb"
	"pkg/rd"
)

type Hub struct {
	ExPath     string
	Settings   settings.Settings
	Logger     logr.Logger
	AppName    string
	Db         *db.DB
	Rd         *rd.RD
	MbProducer *mb.Producer
}

func (g *Hub) Dispose() {
	if g.Db != nil {
		g.Db.Dispose()
	}

	if g.Rd != nil {
		g.Rd.Dispose()
	}
}

func New(exPath string, settings settings.Settings, logger logr.Logger, appName string,
	mbProducer *mb.Producer) Hub {
	return Hub{
		ExPath:     exPath,
		Settings:   settings,
		Logger:     logger,
		Db:         db.New(settings.DBs),
		Rd:         rd.New(settings.RDs, logger),
		AppName:    appName,
		MbProducer: mbProducer,
	}
}

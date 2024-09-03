package settings

import (
	"encoding/json"
	"internal/message_broker"
	"internal/static_storage"
	"os"
	"pkg/db"
)

type Settings struct {
	UrlListen     string                  `json:"url_listen"`
	LogLevel      int                     `json:"log_level"`
	DBs           db.Settings             `json:"dbs"`
	StaticStorage static_storage.Settings `json:"static_storage"`
	MessageBroker message_broker.Settings `json:"message_broker"`
}

func (s *Settings) Read(filePath string) error {
	dat, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	err = json.Unmarshal(dat, s)
	if err != nil {
		return err
	}

	return nil
}

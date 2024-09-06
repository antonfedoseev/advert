package settings

import (
	"encoding/json"
	"internal/static_storage"
	"os"
	"pkg/db"
	"pkg/mb"
	"pkg/rd"
)

type Settings struct {
	UrlListen     string                  `json:"url_listen"`
	LogLevel      int                     `json:"log_level"`
	DBs           db.Settings             `json:"dbs"`
	RDs           rd.Settings             `json:"rds"`
	StaticStorage static_storage.Settings `json:"static_storage"`
	MessageBroker mb.Settings             `json:"mb"`
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

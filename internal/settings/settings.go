package settings

import (
	"encoding/json"
	"os"
)

type DBSpec struct {
	Diver    string `json:"diver"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type DBSettings struct {
	DBs map[string]DBSpec `json:"dbs"`
}

type Settings struct {
	DBSettings
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

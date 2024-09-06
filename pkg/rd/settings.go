package rd

import (
	"strconv"
)

type Spec struct {
	Prefix             string `json:"prefix"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Db                 int    `json:"rd"`
	ClientName         string `json:"client_name"`
	Password           string `json:"password"`
	Name               string `json:"name"`
	MaxIdleCons        int    `json:"max_idle_cons"`
	ConnMaxIdleTimeSec int    `json:"conn_max_idle_time_sec"`
	LogLevel           int    `json:"log_level"`
}

type Settings map[string]Spec
type Alias string

const (
	MainAlias Alias = "main"
)

func (s Settings) getMainDbSpec() Spec {
	return s[string(MainAlias)]
}

func (s Settings) getDbSpecByAlias(alias string) Spec {
	return s[alias]
}

func (s *Spec) ConnStr() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

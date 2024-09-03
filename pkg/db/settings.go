package db

import (
	"fmt"
)

type Spec struct {
	Driver             string `json:"driver"`
	Host               string `json:"host"`
	Port               int    `json:"port"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	Name               string `json:"name"`
	MaxIdleCons        int    `json:"max_idle_cons"`
	MaxOpenCons        int    `json:"max_open_cons"`
	ConnMaxLifetimeSec int    `json:"conn_max_lifetime_sec"`
	ConnMaxIdleTimeSec int    `json:"conn_max_idle_time_sec"`
}

type Settings map[string]Spec

const shardDbAliasPrefix = "shard_"

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

func (s Spec) ConnStr() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", s.Username, s.Password, s.Host, s.Port, s.Name)
}

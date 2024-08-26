package rdb

type Settings struct {
	Host, Prefix string
	//will be shown in CLIENT LIST
	ClientName     string
	Port           int
	Db             int
	LogLevel       int
	MaxIdle        int
	IdleTimeoutSec int
}

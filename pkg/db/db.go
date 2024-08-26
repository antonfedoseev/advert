package db

type Settings struct {
	Host, Port, User, Pass, Name, Prefix, Params string
	Driver                                       string
	LogLevel                                     int
	Weight                                       uint32
	MaxIdleConns, MaxOpenConns                   int
	ConnMaxLifetimeSec, ConnMaxIdleTimeSec       int
}

func (s *Settings) ConnStr() string {
	return s.User + ":" + s.Pass + "@tcp(" + s.Host + ":" + s.Port + ")/" + s.Name + s.Params
}

package rd

import (
	"github.com/go-logr/logr"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"time"
)

type Pool struct {
	rd   *redis.Pool
	spec Spec
}

func OpenPool(spec Spec, logger logr.Logger) *Pool {
	p := &Pool{spec: spec, rd: openPool(spec, logger)}
	return p
}

func openPool(s Spec, logger logr.Logger) *redis.Pool {
	maxIdle := s.MaxIdleCons
	if maxIdle == 0 {
		maxIdle = 16
	}

	idleTimeoutSec := s.ConnMaxIdleTimeSec
	if idleTimeoutSec == 0 {
		idleTimeoutSec = 240
	}

	connStr := s.ConnStr()

	if len(s.Prefix) > 0 {
		logger = logger.WithValues("rd", s.Prefix)
	}

	return &redis.Pool{
		Wait:        false,
		MaxIdle:     maxIdle,
		IdleTimeout: time.Second * time.Duration(idleTimeoutSec),
		Dial: func() (redis.Conn, error) {
			orig, err := redis.Dial("tcp", connStr)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			c := &Conn{
				origin: orig,
				name:   s.Prefix,
				logger: logger,
				spec:   s,
			}

			// Сохраняем название клиента для диагностики соединений с помощью команды CLIENT LIST.
			// https://redis.io/commands/client-list/
			if len(s.ClientName) > 0 {
				_, err = c.Do("CLIENT", "SETNAME", s.ClientName)
				if err != nil {
					return nil, errors.WithStack(err)
				}
			}

			_, err = c.Do("SELECT", s.Db)
			if err != nil {
				return nil, errors.WithStack(err)
			}

			return c, err
		},
		TestOnBorrow: func(c redis.Conn, _ time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}

func (p *Pool) Origin() *redis.Pool {
	return p.rd
}

func (p *Pool) Close() {
	p.rd.Close()
}

func (p *Pool) Get() redis.Conn {
	conn := p.rd.Get()
	return conn
}

func (p *Pool) Do(cmd string, args ...interface{}) (interface{}, error) {
	rc := p.Get()
	defer rc.Close()
	return rc.Do(cmd, args...)
}

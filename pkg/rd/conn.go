package rd

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gomodule/redigo/redis"
)

// Conn is a wrapper of redis origin connection for logging all operations
type Conn struct {
	origin redis.Conn
	name   string
	logger logr.Logger
	spec   Spec
}

func (rd *Conn) Close() error {
	return rd.origin.Close()
}

func (rd *Conn) Err() error {
	return rd.origin.Err()
}

func (rd *Conn) Flush() error {
	return rd.origin.Flush()
}

func (rd *Conn) Do(command string, args ...interface{}) (interface{}, error) {
	if rd.spec.LogLevel > 1 {
		if len(command) != 0 {
			var s string
			for i := 0; i < len(args); i++ {
				s += fmt.Sprintf("%v ", args[i])
			}
			rd.logger.WithCallDepth(2).V(1).Info(command + " " + s)
		}
	} else if rd.spec.LogLevel > 0 {
		//for this level not logging too verbose commands
		if len(command) != 0 && command != "PING" {
			var s string
			for i := 0; i < len(args) && i < 5; i++ {
				s += fmt.Sprintf("%v ", args[i])
			}
			if len(args) > 5 {
				s += " ..."
			}
			rd.logger.WithCallDepth(2).V(1).Info(command + " " + s)
		}
	}
	return rd.origin.Do(command, args...)
}

func (rd *Conn) Send(command string, args ...interface{}) error {
	if rd.spec.LogLevel > 0 {
		if len(command) != 0 {
			var s string
			for i := 0; i < len(args); i++ {
				s += fmt.Sprintf("%v ", args[i])
			}
			rd.logger.WithCallDepth(2).V(1).Info(command + " " + s)
		}
	}
	return rd.origin.Send(command, args...)
}

func (rd *Conn) Receive() (interface{}, error) {
	return rd.origin.Receive()
}

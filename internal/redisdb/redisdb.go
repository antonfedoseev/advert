package redisdb

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"math"
	"strconv"
	"time"
)

type Settings struct {
	Host, Prefix   string
	ClientName     string
	Port           int
	Db             int
	LogLevel       int
	MaxIdle        int
	IdleTimeoutSec int
}

const (
	RedisMutexExpireSec = 60
)

func (s *Settings) ConnStr() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

type Pool struct {
	S  Settings
	RP *redis.Pool
}

func OpenPool(s Settings, logger logr.Logger) *Pool {
	p := &Pool{S: s, RP: newRedisPool(s, logger)}
	return p
}

func (p *Pool) Close() {
	p.RP.Close()
}

func (p *Pool) Get() redis.Conn {
	conn := p.RP.Get()
	return conn
}

// Do is convenience method for 'one-shot' commands
func (p *Pool) Do(cmd string, args ...interface{}) (interface{}, error) {
	rc := p.Get()
	defer rc.Close()
	return rc.Do(cmd, args...)
}

// Conn is redis connection logs all operations
type Conn struct {
	orig   redis.Conn
	name   string
	logger logr.Logger
	s      Settings
}

func (rd *Conn) Close() error {
	return rd.orig.Close()
}

func (rd *Conn) Err() error {
	return rd.orig.Err()
}

func (rd *Conn) Flush() error {
	return rd.orig.Flush()
}

func (rd *Conn) Do(command string, args ...interface{}) (interface{}, error) {
	if rd.s.LogLevel > 1 {
		if len(command) != 0 {
			var s string
			for i := 0; i < len(args); i++ {
				s += fmt.Sprintf("%v ", args[i])
			}
			rd.logger.WithCallDepth(2).V(1).Info(command + " " + s)
		}
	} else if rd.s.LogLevel > 0 {
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
	return rd.orig.Do(command, args...)
}

func (rd *Conn) Send(command string, args ...interface{}) error {
	if rd.s.LogLevel > 0 {
		if len(command) != 0 {
			var s string
			for i := 0; i < len(args); i++ {
				s += fmt.Sprintf("%v ", args[i])
			}
			rd.logger.WithCallDepth(2).V(1).Info(command + " " + s)
		}
	}
	return rd.orig.Send(command, args...)
}

func (rd *Conn) Receive() (interface{}, error) {
	return rd.orig.Receive()
}

func newRedisPool(s Settings, logger logr.Logger) *redis.Pool {

	maxIdle := s.MaxIdle
	if maxIdle == 0 {
		maxIdle = 16
	}

	idleTimeoutSec := s.IdleTimeoutSec
	if idleTimeoutSec == 0 {
		idleTimeoutSec = 240
	}

	connStr := s.ConnStr()

	if len(s.Prefix) > 0 {
		logger = logger.WithValues("[redis]", s.Prefix)
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
				orig:   orig,
				name:   s.Prefix,
				logger: logger,
				s:      s,
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

func GetRedisMutex(pool *Pool, name string, ttlSec int) *redsync.Mutex {
	rdSync := redsync.New(redigo.NewPool(pool.RP))
	mx := rdSync.NewMutex(name, redsync.WithExpiry(time.Duration(ttlSec)*time.Second))
	return mx
}

func GetRedisMutexAutoExpire(pool *Pool, name string) *redsync.Mutex {
	return GetRedisMutex(pool, name, RedisMutexExpireSec)
}

func RemoveRedisMutex(pool *Pool, name string) error {
	rd := pool.Get()
	defer rd.Close()

	_, err := rd.Do("DEL", name)
	return err
}

func ExistsRedisMutex(pool *Pool, name string) (bool, error) {
	rd := pool.Get()
	defer rd.Close()

	exists, err := redis.Bool(rd.Do("EXISTS", name))
	if err != nil {
		if err == redis.ErrNil {
			return false, nil
		}
		return false, err
	}
	return exists, nil
}

func GetUint32(rd redis.Conn, key string) (uint32, error) {
	nUint64, err := redis.Uint64(rd.Do("GET", key))
	if err != nil {
		return 0, err
	}
	if rdErr := rd.Err(); rdErr != nil {
		return 0, rdErr
	}

	nUint32, err := uint64ToUint32(nUint64)
	if err != nil {
		return 0, err
	}

	return nUint32, nil
}

func GetListLen(rd redis.Conn, key string) (uint32, error) {
	nUint64, err := redis.Uint64(rd.Do("LLEN", key))
	if err != nil {
		return 0, err
	}
	if rdErr := rd.Err(); rdErr != nil {
		return 0, rdErr
	}

	nUint32, err := uint64ToUint32(nUint64)
	if err != nil {
		return 0, err
	}

	return nUint32, nil
}

func SendRpushSliceUint32(rd redis.Conn, key string, nums []uint32) {
	if len(nums) == 0 {
		return
	}

	params := make([]interface{}, 0, 1+len(nums))
	params = append(params, key)
	for _, n := range nums {
		params = append(params, n)
	}

	rd.Send("RPUSH", params...)
}

func SendRpushSliceString(rd redis.Conn, key string, strs []string) {
	if len(strs) == 0 {
		return
	}

	params := make([]interface{}, 0, 1+len(strs))
	params = append(params, key)
	for _, s := range strs {
		params = append(params, s)
	}

	rd.Send("RPUSH", params...)
}

func uint64ToUint32(n64 uint64) (uint32, error) {
	if n64 > math.MaxUint32 {
		return 0, errors.Errorf("Can not convert uint64 %d to uint32.", n64)
	}
	return uint32(n64), nil
}

// Uint32s is a helper that converts an array command reply to a []uint32.
// If err is not equal to nil, then Uint32s returns nil, err.
func Uint32s(reply interface{}, err error) ([]uint32, error) {
	var nums []uint32
	if reply == nil {
		return nums, redis.ErrNil
	}
	values, err := redis.Values(reply, err)
	if err != nil {
		return nums, err
	}
	if err := redis.ScanSlice(values, &nums); err != nil {
		return nums, err
	}
	return nums, nil
}

func ReceiveUint32s(rdconn redis.Conn) ([]uint32, error) {
	nums, err := Uint32s(rdconn.Receive())
	if e := getErr(rdconn, err); e != nil {
		return nil, e
	}

	return nums, nil
}

func Uint32(reply interface{}, err error) (uint32, error) {
	nUint64, err := redis.Uint64(reply, err)
	if err != nil {
		return 0, errors.Wrap(err, "Rdb can not convert reply to uint64.")
	}

	nUint32, err := uint64ToUint32(nUint64)
	if err != nil {
		return 0, err
	}

	return nUint32, nil
}

func ReceiveInt(rdConn redis.Conn) (int, error) {
	n, err := redis.Int(rdConn.Receive())
	if err != nil {
		return 0, errors.Wrap(err, "Can not execute ReceiveInt(). Got error from redis.Int().")
	}

	if rdErr := rdConn.Err(); rdErr != nil {
		return 0, errors.Wrap(rdErr, "Can not execute ReceiveInt(). Got error from rdConn.Err().")
	}

	return n, nil
}

func ReceiveUint32(rdConn redis.Conn) (uint32, error) {
	n, err := Uint32(rdConn.Receive())
	if err != nil {
		return 0, errors.Wrap(err, "Can not execute ReceiveUint32(). Got error from Uint32().")
	}

	if rdErr := rdConn.Err(); rdErr != nil {
		return 0, errors.Wrap(rdErr, "Can not execute ReceiveUint32(). Got error from rdConn.Err().")
	}

	return n, nil
}

func GetUint32s(conn redis.Conn, commandName string, args ...interface{}) ([]uint32, error) {
	return Uint32s(Do(conn, commandName, args...))
}

func Do(conn redis.Conn, commandName string, args ...interface{}) (reply interface{}, err error) {
	result, err := conn.Do(commandName, args...)
	if e := getErr(conn, err); e != nil {
		return nil, e
	}

	return result, nil
}

func Flush(conn redis.Conn) error {
	if err := conn.Flush(); err != nil {
		return errors.Wrap(err, "Can not execute function Flush. Got error from redis.Conn Flush.")
	}

	if err := conn.Err(); err != nil {
		return errors.Wrap(err, "Can not execute function Flush. Got redis connection error.")
	}

	return nil
}

func Err(conn redis.Conn, err error) error {
	return getErr(conn, err)
}

// NOTE: workaround for redigo not returning original connection error
func getErr(conn redis.Conn, err error) error {
	if err != nil {
		if err == redis.ErrNil {
			return nil
		}
		return errors.Wrap(err, "Got redis error.")
	}

	if connErr := conn.Err(); connErr != nil {
		return errors.Wrap(connErr, "Got redis connection error.")
	}

	return nil
}

type tracked struct {
	subject redis.Conn
}

func (t *tracked) Close() error {
	return t.subject.Close()
}
func (t *tracked) Do(cmd string, args ...interface{}) (interface{}, error) {
	return t.subject.Do(cmd, args...)
}
func (t *tracked) Send(cmd string, args ...interface{}) error { return t.subject.Send(cmd, args...) }
func (t *tracked) Err() error                                 { return t.subject.Err() }
func (t *tracked) Flush() error                               { return t.subject.Flush() }
func (t *tracked) Receive() (interface{}, error)              { return t.subject.Receive() }

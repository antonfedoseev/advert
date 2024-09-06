package rd

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/redigo"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"math"
	"time"
)

func GetRedisMutex(pool *Pool, name string, ttlSec int) *redsync.Mutex {
	rdSync := redsync.New(redigo.NewPool(pool.Origin()))
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

func SendRPushSliceUint32(rd redis.Conn, key string, nums []uint32) {
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

func SendRPushSliceString(rd redis.Conn, key string, strs []string) {
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

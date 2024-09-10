package dbshard

import (
	"database/sql"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/gomodule/redigo/redis"
	"github.com/pkg/errors"
	"internal/constant"
	"pkg/db"
	"pkg/rd"
)

const (
	rdUserShardKey      = constant.AppPrefix + ":user2shard:"
	rdUserShardCacheSec = 86400 * 3
)

func GetShardDbByUserId(mainDb *db.Conn, shardPools []*db.Pool, rdp *rd.Pool, logger logr.Logger, id uint32) (
	*db.Conn, uint32, error) {

	rdKey := rdUserShardKey + fmt.Sprintf("%d", id)

	var shardId uint32 = 0

	rdShardId, err := redis.Uint64(rdp.Do("GET", rdKey))
	if err == redis.ErrNil {
		dbShardId, err := FindUserShardById(mainDb, id)
		if err != nil {
			return nil, 0, err
		}

		//NOTE: don't care about error here
		rdp.Do("SETEX", rdKey, rdUserShardCacheSec, dbShardId)
		shardId = uint32(dbShardId)
	} else {
		if err != nil {
			return nil, 0, errors.WithStack(err)
		}
		shardId = uint32(rdShardId)
	}

	if shardId == 0 {
		return nil, 0, errors.Wrapf(db.ErrNotFound, "invalid shard(0), passed id:%d, cached in redis as:%d", id, rdShardId)
	}

	db, err := GetShardDbConn(logger, shardPools, shardId)
	return db, shardId, err
}

var errUserNotRegistered = errors.New("user not registered")

func FindUserShardById(mainDb *db.Conn, userId uint32) (int, error) {
	var shardId int
	sb := mainDb.Select("shard_id")
	err := sb.From("user_shard").Where(sb.Equal("user_id", userId)).Limit(1).LoadValue(&shardId)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return -1, errors.Wrapf(errUserNotRegistered, ": user id \"%d\" not found", userId)
		} else {
			return -1, errors.WithStack(err)
		}
	}

	return shardId, nil
}

// GetShardDbConn takes shardId starts from 1
func GetShardDbConn(logger logr.Logger, shardPools []*db.Pool, shardId uint32) (*db.Conn, error) {
	index := int(shardId) - 1

	if index < 0 || index >= len(shardPools) {
		return nil, errors.Errorf("shard id is out of range: %d", shardId)
	}

	return db.NewDbConn(shardPools[index], logger), nil
}

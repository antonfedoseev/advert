package env

import (
	"github.com/go-logr/logr"
	"internal/dbshard"
	"internal/global"
	"internal/message_broker"
	"internal/redisdb"
	"internal/settings"
	"pkg/db"
)

type Environment struct {
	hub          global.Hub
	mainDbConn   *db.Conn
	user2ShardDb map[uint32]*db.Conn
	Settings     settings.Settings
	Logger       logr.Logger
}

func NewEnvironment(hub global.Hub) *Environment {
	env := &Environment{
		hub:      hub,
		Logger:   hub.Logger,
		Settings: hub.Settings,
	}
	return env
}

func (env *Environment) Close() {
	if env.mainDbConn != nil {
		env.mainDbConn.Rollback()
		env.mainDbConn = nil
	}

	for _, db := range env.user2ShardDb {
		db.Rollback()
	}
	env.user2ShardDb = nil
}

func (env *Environment) MainDb() *db.Conn {
	if env.mainDbConn == nil {
		db := db.NewDbConn(env.hub.Db.MainPool(), env.Logger)
		env.mainDbConn = db
	}
	return env.mainDbConn
}

func (env *Environment) ShardDb(userId uint32) (*db.Conn, error) {
	if env.user2ShardDb == nil {
		env.user2ShardDb = make(map[uint32]*db.Conn)
	}
	if db, ok := env.user2ShardDb[userId]; ok {
		return db, nil
	}

	db, shardId, err := dbshard.GetShardDbByUserId(env.mainDbConn, env.hub.Db.Shards(), env.Rd(), env.Logger, userId)
	if err != nil {
		return nil, err
	}

	if !env.isShardDbRegistered(db) {
		env.setupShardDb(shardId, db)
	}
	env.user2ShardDb[userId] = db

	return db, nil
}

func (env *Environment) isShardDbRegistered(shardDb *db.Conn) bool {
	for _, registeredShardDb := range env.user2ShardDb {
		if shardDb == registeredShardDb {
			return true
		}
	}
	return false
}

func (env *Environment) setupShardDb(shardId uint32, shardDb *db.Conn) {
	//overriding default logger
	shardDb.Logger = env.Logger.WithValues("shard", shardId)
}

func (env *Environment) Rd() *redisdb.Pool {
	return env.hub.RedisDb
}

func (env *Environment) AppName() string {
	return env.hub.AppName
}

func (env *Environment) MbProducer() *message_broker.Producer {
	return env.hub.MbProducer
}

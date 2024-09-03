package db

import (
	"errors"
	"fmt"
	//_ "github.com/go-sql-driver/mysql"
	//_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
	"strings"
	"time"
)

var (
	ErrOutOfShardsRange   = errors.New("shardId is out of shards range")
	ErrUnknownPool        = errors.New("unknown pool")
	ErrNotFound           = errors.New("not found")
	ErrNotUTF8            = errors.New("invalid UTF-8")
	ErrInvalidSliceLength = errors.New("length of slice is 0. length must be >= 1")
	ErrInvalidSliceValue  = errors.New("trying to interpolate invalid slice value into query")
	ErrInvalidValue       = errors.New("trying to interpolate invalid value into query")
	ErrArgumentMismatch   = errors.New("mismatch between ? (placeholders) and arguments")
)

type DB struct {
	settings   Settings
	pools      map[string]*Pool
	shardPools []*Pool
}

func New(s Settings) *DB {
	d := &DB{settings: s}
	d.init()
	return d
}

func (d *DB) init() {
	d.openConnectionsPools()
	d.defineShardsPools()
}

func (d *DB) Shards() []*Pool {
	return d.shardPools
}

func (d *DB) defineShardsPools() {
	for alias, pool := range d.pools {
		if strings.HasPrefix(alias, shardDbAliasPrefix) {
			d.shardPools = append(d.shardPools, pool)
		}
	}
}

func (d *DB) openConnectionsPools() {
	d.pools = make(map[string]*Pool)

	for alias, spec := range d.settings {
		db := openPool(spec)
		d.pools[alias] = &Pool{db: db, alias: alias}
	}
}

func openPool(s Spec) *sqlx.DB {
	driver := s.Driver
	if len(driver) == 0 {
		driver = "mysql"
	}
	//NOTE: sql.Open(..) doesn't happen to return an error

	sqlDb, err := sqlx.Connect(driver, s.ConnStr())
	if err != nil {
		panic("failed to connect to mysql on start: " + err.Error())
	}

	if s.MaxIdleCons == 0 {
		//NOTE: using default sql.DB settings
		sqlDb.SetMaxIdleConns(2)
	} else {
		sqlDb.SetMaxIdleConns(s.MaxIdleCons)
	}
	if s.MaxOpenCons != 0 {
		sqlDb.SetMaxOpenConns(s.MaxOpenCons)
	}
	if s.ConnMaxLifetimeSec != 0 {
		sqlDb.SetConnMaxLifetime(time.Second * time.Duration(s.ConnMaxLifetimeSec))
	}
	if s.ConnMaxIdleTimeSec != 0 {
		sqlDb.SetConnMaxIdleTime(time.Second * time.Duration(s.ConnMaxIdleTimeSec))
	}

	return sqlDb
}

func getShardAlias(shardId uint) string {
	return fmt.Sprintf("%s%02d", shardDbAliasPrefix, shardId)
}

func (d *DB) ShardsAmount() uint {
	return uint(len(d.shardPools))
}

func (d *DB) MainPool() *Pool {
	return d.PoolByAlias(MainAlias)
}

func (d *DB) ShardPoolById(shardId uint) *Pool {
	if shardId > d.ShardsAmount() {
		panic(ErrOutOfShardsRange)
	}

	return d.getPoolByAlias(getShardAlias(shardId))
}

func (d *DB) PoolByAlias(alias Alias) *Pool {
	return d.getPoolByAlias(string(alias))
}

func (d *DB) getPoolByAlias(alias string) *Pool {
	pool, ok := d.pools[alias]
	if !ok {
		panic(fmt.Errorf("pool with alias \"%s\" not found: %w", alias, ErrUnknownPool))
	}

	return pool
}

func (d *DB) ForEachShardPool(f func(pool *Pool) error) error {
	for i := uint(0); i < d.ShardsAmount(); i++ {
		p := d.shardPools[i]

		err := f(p)
		if err != nil {
			return err
		}
	}

	return nil
}

func (d *DB) Dispose() {
	d.closeConnectionsPools()
}

func (d *DB) closeConnectionsPools() {
	for _, pool := range d.pools {
		closePool(pool)
	}

	for alias := range d.pools {
		delete(d.pools, alias)
	}
}

func closePool(pool *Pool) {
	pool.Close()
}

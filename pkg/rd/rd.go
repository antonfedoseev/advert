package rd

import (
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

const (
	RedisMutexExpireSec = 60
)

var ErrUnknownPool = errors.New("unknown pool")

type RD struct {
	settings Settings
	pools    map[string]*Pool
	logger   logr.Logger
}

func New(s Settings, logger logr.Logger) *RD {
	d := &RD{settings: s, logger: logger}
	d.init()
	return d
}

func (r *RD) init() {
	r.openConnectionsPools()
}

func (r *RD) openConnectionsPools() {
	r.pools = make(map[string]*Pool)

	for alias, spec := range r.settings {
		pool := OpenPool(spec, r.logger)
		r.pools[alias] = pool
	}
}

func (r *RD) MainPool() *Pool {
	return r.PoolByAlias(MainAlias)
}

func (r *RD) PoolByAlias(alias Alias) *Pool {
	return r.getPoolByAlias(string(alias))
}

func (r *RD) getPoolByAlias(alias string) *Pool {
	pool, ok := r.pools[alias]
	if !ok {
		panic(fmt.Errorf("pool with alias \"%s\" not found: %w", alias, ErrUnknownPool))
	}

	return pool
}

func (r *RD) ForEachPool(f func(pool *Pool) error) error {
	for _, pool := range r.pools {
		err := f(pool)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *RD) Dispose() {
	r.closeConnectionsPools()
}

func (r *RD) closeConnectionsPools() {
	for _, pool := range r.pools {
		closePool(pool)
	}

	for alias := range r.pools {
		delete(r.pools, alias)
	}
}

func closePool(pool *Pool) {
	pool.Close()
}

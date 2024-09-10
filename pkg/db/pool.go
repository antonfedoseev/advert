package db

import "github.com/jmoiron/sqlx"

type Pools []*Pool

func (p *Pools) Len() int {
	return len(*p)
}

func (p *Pools) Less(i, j int) bool {
	return (*p)[i].alias < (*p)[j].alias
}

func (p *Pools) Swap(i, j int) {
	(*p)[i], (*p)[j] = (*p)[j], (*p)[i]
}

type Pool struct {
	db    *sqlx.DB
	alias string
}

func (p *Pool) Close() {
	p.db.Close()
	p.db = nil
}

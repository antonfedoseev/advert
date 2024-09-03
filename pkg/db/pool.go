package db

import "github.com/jmoiron/sqlx"

type Pool struct {
	db    *sqlx.DB
	alias string
}

func (p *Pool) Close() {
	p.db.Close()
	p.db = nil
}

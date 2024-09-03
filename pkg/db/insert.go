package db

import (
	"database/sql"
	"github.com/huandu/go-sqlbuilder"
)

// InsertBuilder contains the clauses for an INSERT statement
type InsertBuilder struct {
	origin *sqlbuilder.InsertBuilder
	dbConn *Conn
}

func (b *InsertBuilder) Exec() (sql.Result, error) {
	sql, args := b.origin.Build()
	runner := b.dbConn.getExecRunner()
	return runner.NamedExec(sql, args)
}

func (b *InsertBuilder) Cols(col ...string) *InsertBuilder {
	b.origin.Cols(col...)
	return b
}

func (b *InsertBuilder) Values(value ...interface{}) *InsertBuilder {
	b.origin.Values(value...)
	return b
}

func (b *InsertBuilder) SQL(sql string) *InsertBuilder {
	b.origin.SQL(sql)
	return b
}

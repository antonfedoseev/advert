package db

import (
	"database/sql"
	"github.com/huandu/go-sqlbuilder"
)

// DeleteBuilder contains the clauses for a DELETE statement
type DeleteBuilder struct {
	dbConn *Conn
	origin *sqlbuilder.DeleteBuilder
}

func (b *DeleteBuilder) Exec() (sql.Result, error) {
	sql, args := b.origin.Build()
	runner := b.dbConn.getExecRunner()
	return runner.NamedExec(sql, args)
}

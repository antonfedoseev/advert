package db

import (
	"database/sql"
	"github.com/huandu/go-sqlbuilder"
)

// UpdateBuilder contains the clauses for an UPDATE statement
type UpdateBuilder struct {
	dbConn *Conn
	origin *sqlbuilder.UpdateBuilder
	args   []interface{}
}

func (b *UpdateBuilder) Exec() (sql.Result, error) {
	sql, args := b.origin.Build()

	if len(args) == 0 {
		args = b.args
	} else if len(b.args) != 0 {
		args = append(b.args, args...)
	}

	runner := b.dbConn.getExecRunner()
	return runner.NamedExec(sql, args)
}

func (b *UpdateBuilder) Set(value ...string) *UpdateBuilder {
	b.origin.Set(value...)
	return b
}

func (b *UpdateBuilder) Where(andExpr ...string) *UpdateBuilder {
	b.origin.Where(andExpr...)
	return b
}

func (b *UpdateBuilder) Limit(limit int) *UpdateBuilder {
	b.origin.Limit(limit)
	return b
}

func (b *UpdateBuilder) Assign(field string, value interface{}) string {
	return b.origin.Assign(field, value)
}

func (b *UpdateBuilder) Equal(field string, value interface{}) string {
	return b.origin.Equal(field, value)
}

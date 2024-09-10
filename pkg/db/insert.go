package db

import (
	"database/sql"
	"fmt"
	"github.com/huandu/go-sqlbuilder"
	"reflect"
)

// InsertBuilder contains the clauses for an INSERT statement
type InsertBuilder struct {
	origin *sqlbuilder.InsertBuilder
	dbConn *Conn
}

func (b *InsertBuilder) Exec() (sql.Result, error) {
	sql, args := b.origin.Build()
	runner := b.dbConn.getExecRunner()
	result, err := runner.Exec(sql, args...)
	return result, err
}

func (b *InsertBuilder) Cols(col ...string) *InsertBuilder {
	b.origin.Cols(col...)
	return b
}

func (b *InsertBuilder) Values(value ...interface{}) *InsertBuilder {
	correctValues(&value)
	b.origin.Values(value...)
	return b
}

func correctValues(value *[]interface{}) {
	for i := 0; i < len(*value); i++ {
		elem := (*value)[i]
		val := reflect.ValueOf(elem)
		if val.Type() == reflect.TypeOf(Point{}) {
			point := elem.(Point)
			(*value)[i] = fmt.Sprintf("ST_GeomFromText('POINT(%f %f)')", point.Longitude, point.Latitude)
		}
	}
}

func (b *InsertBuilder) SQL(sql string) *InsertBuilder {
	b.origin.SQL(sql)
	return b
}

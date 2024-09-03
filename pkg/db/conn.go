package db

import (
	"database/sql"
	"github.com/go-logr/logr"
	"github.com/huandu/go-sqlbuilder"
	"github.com/jmoiron/sqlx"
	"strings"
)

type selectRunner interface {
	Select(dest interface{}, query string, args ...interface{}) error
}

type execRunner interface {
	NamedExec(query string, arg interface{}) (sql.Result, error)
}

type Conn struct {
	Logger    logr.Logger
	pool      *Pool
	tx        *sqlx.Tx
	trRefs    int
	commitTry int
}

func NewDbConn(pool *Pool, logger logr.Logger) *Conn {
	if len(pool.alias) > 0 {
		logger = logger.WithValues("db", pool.alias)
	}

	dbc := &Conn{Logger: logger, pool: pool}

	return dbc
}

func (dbConn *Conn) Transaction(txFunc func(dbConn *Conn) error) (err error) {
	err = dbConn.Begin()
	if err != nil {
		return
	}
	defer func() {
		if p := recover(); p != nil {
			dbConn.Rollback()
			panic(p) // re-throw panic after Rollback
		} else if err != nil {
			dbConn.Rollback() // err is non-nil; don't change it
		} else {
			err = dbConn.Commit() // err is nil; if Commit returns error update err
		}
	}()
	err = txFunc(dbConn)
	return err
}

func (dbConn *Conn) Begin() error {
	//check if we are already in a transaction
	if dbConn.tx != nil {
		dbConn.trRefs++
		return nil
	}
	tx, err := dbConn.pool.db.Beginx()
	if err != nil {
		dbConn.Logger.Error(err, "db.begin.error")
		return err
	} else {
		dbConn.Logger.Info("dbr.begin")
	}

	if err == nil {
		dbConn.tx = tx
		dbConn.trRefs = 1
		dbConn.commitTry = 0
	}
	return err
}

func (dbConn *Conn) Rollback() error {
	if dbConn.trRefs > 0 {
		dbConn.trRefs--
		if dbConn.trRefs == 0 {
			if err := dbConn.tx.Rollback(); err != nil {
				return err
			}
			dbConn.tx = nil
		}
	}
	return nil
}

func (dbConn *Conn) RollbackOnDefer() {
	if dbConn.commitTry == 0 {
		dbConn.Rollback()
	} else {
		dbConn.commitTry--
	}
}

func (dbConn *Conn) Commit() error {
	dbConn.commitTry++
	if dbConn.trRefs > 0 {
		dbConn.trRefs--
		if dbConn.trRefs == 0 {
			if err := dbConn.tx.Commit(); err != nil {
				return err
			}
			dbConn.tx = nil
		}
	}
	return nil
}

func (dbConn *Conn) getSelectRunner() selectRunner {
	if dbConn.tx != nil {
		return dbConn.tx
	}

	return dbConn.pool.db
}

func (dbConn *Conn) getExecRunner() execRunner {
	if dbConn.tx != nil {
		return dbConn.tx
	}

	return dbConn.pool.db
}

func (dbConn *Conn) getFlavor() sqlbuilder.Flavor {
	name := strings.ToLower(dbConn.pool.db.DriverName())

	switch name {
	case "mysql":
		return sqlbuilder.MySQL
	case "postgresql":
		return sqlbuilder.PostgreSQL
	case "sqlite":
		return sqlbuilder.SQLite
	case "sqlserver":
		return sqlbuilder.SQLServer
	case "cql":
		return sqlbuilder.CQL
	case "clickhouse":
		return sqlbuilder.ClickHouse
	case "presto":
		return sqlbuilder.Presto
	case "oracle":
		return sqlbuilder.Oracle
	case "informix":
		return sqlbuilder.Informix
	}

	return sqlbuilder.DefaultFlavor
}

func (dbConn *Conn) Select(col ...string) *SelectBuilder {
	flavor := dbConn.getFlavor()
	sb := flavor.NewSelectBuilder()
	sb = sb.Select(col...)
	return &SelectBuilder{dbConn: dbConn, origin: sb}
}

func (dbConn *Conn) SelectBySQL(sql string, args ...interface{}) *SelectBuilder {
	flavor := dbConn.getFlavor()
	sb := flavor.NewSelectBuilder()
	sb = sb.SQL(sql)
	return &SelectBuilder{dbConn: dbConn, origin: sb, args: args}
}

func (dbConn *Conn) InsertInto(table string) *InsertBuilder {
	flavor := dbConn.getFlavor()
	ib := flavor.NewInsertBuilder()
	ib = ib.InsertInto(table)
	return &InsertBuilder{ib, dbConn}
}

func (dbConn *Conn) ReplaceInto(table string) *InsertBuilder {
	flavor := dbConn.getFlavor()
	ib := flavor.NewInsertBuilder()
	ib = ib.ReplaceInto(table)
	return &InsertBuilder{ib, dbConn}
}

func (dbConn *Conn) Update(table string) *UpdateBuilder {
	flavor := dbConn.getFlavor()
	ub := flavor.NewUpdateBuilder()
	ub = ub.Update(table)
	return &UpdateBuilder{dbConn: dbConn, origin: ub}
}

func (dbConn *Conn) UpdateBySQL(sql string, args ...interface{}) *UpdateBuilder {
	flavor := dbConn.getFlavor()
	ub := flavor.NewUpdateBuilder()
	ub = ub.SQL(sql)
	return &UpdateBuilder{dbConn: dbConn, origin: ub, args: args}
}

func (dbConn *Conn) DeleteFrom(table string) *DeleteBuilder {
	flavor := dbConn.getFlavor()
	db := flavor.NewDeleteBuilder()
	db = db.DeleteFrom(table)
	return &DeleteBuilder{dbConn: dbConn, origin: db}
}

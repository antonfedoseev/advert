package main

import (
	"context"
	"database/sql"
	"fmt"
	"internal/settings"
	"pkg/db"
	"time"
)
import "github.com/emirpasic/gods/trees/btree"

func main() {
	t := btree.NewWithIntComparator(100)
	t.Put(10, "10")
	fmt.Println("Hello 1")
	time.Sleep(30 * time.Second)
	fmt.Println("Hello 2")

	ctx := context.Background()
	ctx, quit := context.WithCancel(ctx)
	defer quit()

	s, err := initSettings("settings.json")
	if err != nil {
		panic("failed to read settings: " + err.Error())
	}

	dbc := OpenPool(s.DB_MAIN)

	rows, err := dbc.Query("")
	if rows.Err() != nil {

	}
	defer rows.Close()
}

func initSettings(path string) (settings.Settings, error) {
	s := settings.Settings{}
	err := s.Read(path)
	return s, err
}

func OpenPool(s db.Settings) *sql.DB {
	driver := s.Driver
	if len(driver) == 0 {
		driver = "mysql"
	}
	//NOTE: sql.Open(..) doesn't happen to return an error
	sqlDb, err := sql.Open(driver, s.ConnStr())
	if err != nil {
		panic("failed to connect to mysql on start: " + err.Error())
	}

	if s.MaxIdleConns == 0 {
		//NOTE: using default sql.DB settings
		sqlDb.SetMaxIdleConns(2)
	} else {
		sqlDb.SetMaxIdleConns(s.MaxIdleConns)
	}
	if s.MaxOpenConns != 0 {
		sqlDb.SetMaxOpenConns(s.MaxOpenConns)
	}
	if s.ConnMaxLifetimeSec != 0 {
		sqlDb.SetConnMaxLifetime(time.Second * time.Duration(s.ConnMaxLifetimeSec))
	}
	if s.ConnMaxIdleTimeSec != 0 {
		sqlDb.SetConnMaxIdleTime(time.Second * time.Duration(s.ConnMaxIdleTimeSec))
	}

	err = sqlDb.Ping()
	if err != nil {
		panic("failed to ping mysql on start: " + err.Error())
	}

	return sqlDb
}

// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// migrations holds all our SQL migrations to be done (in order)
	migrations = []string{
		// Customer tables
		`create table if not exists customers(customer_id primary key, first_name, middle_name, last_name, nick_name, suffix, birthdate, gender, culture, status, email, created_at datetime, last_modified datetime, deleted_at datetime);`,
		`create table if not exists customers_phones(customer_id, number, valid, type, unique (customer_id, number) on conflict abort);`,
		`create table if not exists customers_addresses(customer_id, type, address1, address2, city, state, postal_code, country, validated, active, unique (customer_id, address1) on conflict abort);`,

		// Account tables
		`create table if not exists accounts(account_id primary key, customer_id, name, account_number, routing_number, status, type, created_at datetime, closed_at datetime, last_modified datetime, deleted_at datetime);`,
		// TODO(adam): make unique constraint on (account_number, routing_number) and (account_id, customer_id)
	}
)

// getSqlitePath returns a sqlite database path of either the current
// working directory or the SQLITE_DB_PATH  env variable value.
func getSqlitePath() string {
	path := os.Getenv("SQLITE_DB_PATH")
	if path == "" || strings.Contains(path, "..") {
		// set default if empty or trying to escape
		// don't filepath.ABS to avoid full-fs reads
		path = "gl.db"
	}
	return path
}

// createSqliteConnection returns a sql.DB associated to a SQLite database file at path.
func createSqliteConnection(logger log.Logger, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		err = fmt.Errorf("problem opening sqlite3 file %s: %v", path, err)
		if logger != nil {
			logger.Log("sqlite", err)
		}
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("problem with Ping against *sql.DB %s: %v", path, err)
	}
	return db, nil
}

// migrate runs our database migrations a sql.DB. db should be like any other database/sql driver.
//
// https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go
// https://astaxie.gitbooks.io/build-web-application-with-golang/en/05.3.html
func migrate(logger log.Logger, db *sql.DB) error {
	if logger != nil {
		logger.Log("sqlite", "starting database migrations")
	}
	for i := range migrations {
		row := migrations[i]
		res, err := db.Exec(row)
		if err != nil {
			return fmt.Errorf("migration #%d [%s...] had problem: %v", i, row[:40], err)
		}
		n, err := res.RowsAffected()
		if err == nil && logger != nil {
			logger.Log("sqlite", fmt.Sprintf("migration #%d [%s...] changed %d rows", i, row[:40], n))
		}
	}
	if logger != nil {
		logger.Log("sqlite", "finished migrations")
	}
	return nil
}

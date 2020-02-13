// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"database/sql"

	"github.com/lopezator/migrator"
)

func execsql(name, raw string) *migrator.MigrationNoTx {
	return &migrator.MigrationNoTx{
		Name: name,
		Func: func(db *sql.DB) error {
			_, err := db.Exec(raw)
			return err
		},
	}
}

// UniqueViolation returns true when the provided error matches a database error
// for duplicate entries (violating a unique table constraint).
func UniqueViolation(err error) bool {
	return MySQLUniqueViolation(err) || SqliteUniqueViolation(err)
}

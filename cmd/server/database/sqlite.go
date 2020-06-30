// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	kitprom "github.com/go-kit/kit/metrics/prometheus"
	"github.com/lopezator/migrator"
	"github.com/mattn/go-sqlite3"
	stdprom "github.com/prometheus/client_golang/prometheus"
)

var (
	sqliteConnections = kitprom.NewGaugeFrom(stdprom.GaugeOpts{
		Name: "sqlite_connections",
		Help: "How many sqlite connections and what status they're in.",
	}, []string{"state"})

	sqliteVersionLogOnce sync.Once

	sqliteMigrations = migrator.Migrations(
		execsql(
			"create_accounts",
			`create table if not exists accounts(account_id primary key, customer_id, name, account_number, routing_number, status, type, created_at datetime, closed_at datetime, last_modified datetime, deleted_at datetime, unique(account_number, routing_number));`,
		),
		execsql(
			"create_transactions",
			`create table if not exists transactions(transaction_id primary key, timestamp datetime, created_at datetime, deleted_at datetime);`,
		),
		execsql(
			"create_transaction_lines",
			`create table if not exists transaction_lines(transaction_id, account_id, purpose, amount integer, created_at datetime, deleted_at datetime, unique(transaction_id, account_id));`,
		),
		execsql(
			"create_transaction_lines_id_index",
			`create index transaction_lines_id_index on transaction_lines(transaction_id);`,
		),
		execsql(
			"create_transaction_lines_account_index",
			`create index transaction_lines_account_index on transaction_lines(account_id);`,
		),
	)
)

type sqlite struct {
	path string

	connections *kitprom.Gauge
	logger      log.Logger

	err error
}

func (s *sqlite) Connect(ctx context.Context) (*sql.DB, error) {
	if s.err != nil {
		return nil, fmt.Errorf("sqlite had error %v", s.err)
	}

	sqliteVersionLogOnce.Do(func() {
		if v, _, _ := sqlite3.Version(); v != "" {
			s.logger.Log("main", fmt.Sprintf("sqlite version %s", v))
		}
	})

	db, err := sql.Open("sqlite3", s.path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return db, err
	}

	// Migrate our database
	if m, err := migrator.New(sqliteMigrations); err != nil {
		return db, err
	} else {
		if err := m.Migrate(db); err != nil {
			return db, err
		}
	}

	// Spin up metrics only after everything works
	go func() {
		t := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				recordStatus(s.connections, db)
			}

		}
	}()

	return db, err
}

func SQLiteConnection(logger log.Logger, path string) *sqlite {
	return &sqlite{
		path:        path,
		logger:      logger,
		connections: sqliteConnections,
	}
}

func SQLitePath() string {
	path := os.Getenv("SQLITE_DB_PATH")
	if path == "" || strings.Contains(path, "..") {
		// set default if empty or trying to escape
		// don't filepath.ABS to avoid full-fs reads
		path = "accounts.db"
	}
	return path
}

// TestSQLiteDB is a wrapper around sql.DB for SQLite connections designed for tests to provide
// a clean database for each testcase.  Callers should cleanup with Close() when finished.
type TestSQLiteDB struct {
	DB *sql.DB

	Dir string // temp dir created for sqlite files

	shutdown func() // context shutdown func
}

func (r *TestSQLiteDB) Close() error {
	r.shutdown()

	// Verify all connections are closed before closing DB
	if conns := r.DB.Stats().OpenConnections; conns != 0 {
		panic(fmt.Sprintf("found %d open sqlite connections", conns))
	}
	if err := r.DB.Close(); err != nil {
		return err
	}
	return os.RemoveAll(r.Dir)
}

// CreateTestSqliteDB returns a TestSQLiteDB which can be used in tests
// as a clean sqlite database. All migrations are ran on the db before.
//
// Callers should call close on the returned *TestSQLiteDB.
func CreateTestSqliteDB(t *testing.T) *TestSQLiteDB {
	dir, err := ioutil.TempDir("", "accounts-sqlite")
	if err != nil {
		t.Fatalf("sqlite test: %v", err)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())

	db, err := SQLiteConnection(log.NewNopLogger(), filepath.Join(dir, "accounts.db")).Connect(ctx)
	if err != nil {
		t.Fatalf("sqlite test: %v", err)
	}

	// Don't allow idle connections so we can verify all are closed at the end of testing
	db.SetMaxIdleConns(0)

	return &TestSQLiteDB{DB: db, Dir: dir, shutdown: cancelFunc}
}

// SqliteUniqueViolation returns true when the provided error matches the SQLite error
// for duplicate entries (violating a unique table constraint).
func SqliteUniqueViolation(err error) bool {
	match := strings.Contains(err.Error(), "UNIQUE constraint failed")
	if e, ok := err.(sqlite3.Error); ok {
		return match || e.Code == sqlite3.ErrConstraint
	}
	return match
}

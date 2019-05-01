// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/moov-io/base"

	"github.com/go-kit/kit/log"
)

type testSqliteTransactionRepository struct {
	*sqliteTransactionRepository

	db *testSqliteDB
}

func (r *testSqliteTransactionRepository) Close() error {
	r.db.close()
	return r.sqliteTransactionRepository.Close()
}

func createTestSqliteTransactionRepository(t *testing.T) *testSqliteTransactionRepository {
	t.Helper()

	db, err := createTestSqliteDB()
	if err != nil {
		t.Fatal(err)
	}
	repo, err := setupSqliteTransactionStorage(log.NewNopLogger(), filepath.Join(db.dir, "gl.db"))
	if err != nil {
		t.Fatal(err)
	}
	return &testSqliteTransactionRepository{repo, db}
}

func TestSqliteTransactionRepository__Ping(t *testing.T) {
	repo := createTestSqliteTransactionRepository(t)
	defer repo.Close()

	if err := repo.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestSqliteTransactionRepository(t *testing.T) {
	repo := createTestSqliteTransactionRepository(t)
	defer repo.Close()

	account1, account2 := base.ID(), base.ID()
	tx := transaction{
		ID:        base.ID(),
		Timestamp: time.Now(),
		Lines: []transactionLine{
			{
				AccountId: account1,
				Purpose:   ACHDebit,
				Amount:    -500,
			},
			{
				AccountId: account2,
				Purpose:   ACHCredit,
				Amount:    500,
			},
		},
	}
	if err := repo.createTransaction(tx); err != nil {
		t.Fatal(err)
	}

	transactions, err := repo.getAccountTransactions(account1)
	if err != nil {
		t.Error(err)
	}
	if len(transactions) != 1 {
		t.Errorf("got %d transactions: %v", len(transactions), transactions)
	}
	if transactions[0].ID != tx.ID || len(transactions[0].Lines) != 2 {
		t.Errorf("%#v", transactions[0])
	}

	dbtx, _ := repo.db.db.Begin()

	bal, err := repo.getAccountBalance(dbtx, account1)
	if err != nil || bal != -500 {
		t.Errorf("got balance of %d", bal)
	}
	bal, err = repo.getAccountBalance(dbtx, account2)
	if err != nil || bal != 500 {
		t.Errorf("got balance of %d", bal)
	}
}

// TODO(adam): Check we handle insufficient funds

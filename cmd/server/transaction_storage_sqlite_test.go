// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/gl"

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

	// Override the accountRepository and write our gl.Accounts
	account1, account2 := base.ID(), base.ID()
	repo.sqliteTransactionRepository.accountRepo = &testAccountRepository{
		accounts: []*gl.Account{
			// Setup the account being debited from as 'remote' (routing number we don't manage)
			// so we can send the ACH file and possibly get a return.
			{ID: account1, AccountNumber: "123", RoutingNumber: "121042882"},
			{ID: account2, AccountNumber: "432", RoutingNumber: defaultRoutingNumber},
		},
	}

	// Attempt our transaction
	tx := transaction{
		ID:        base.ID(),
		Timestamp: time.Now(),
		Lines: []transactionLine{
			{AccountId: account1, Purpose: ACHDebit, Amount: -500},
			{AccountId: account2, Purpose: ACHCredit, Amount: 500},
		},
	}
	if err := repo.createTransaction(tx, createTransactionOpts{AllowOverdraft: false}); err != nil {
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

// TestSqliteTransactionRepository__AllowOverdraft will create an internal transfer, but allow an overdraft to occur
func TestSqliteTransactionRepository__AllowOverdraft(t *testing.T) {
	repo := createTestSqliteTransactionRepository(t)
	defer repo.Close()

	// Override the accountRepository and write our gl.Accounts
	account1, account2 := base.ID(), base.ID()
	repo.sqliteTransactionRepository.accountRepo = &testAccountRepository{
		accounts: []*gl.Account{
			// Setup the account being debited from as 'internal' (routing number we manage).
			{ID: account1, AccountNumber: "123", RoutingNumber: defaultRoutingNumber},
			{ID: account2, AccountNumber: "432", RoutingNumber: "121042882"},
		},
	}

	// Attempt our transaction
	tx := transaction{
		ID:        base.ID(),
		Timestamp: time.Now(),
		Lines: []transactionLine{
			{AccountId: account1, Purpose: ACHDebit, Amount: -500},
			{AccountId: account2, Purpose: ACHCredit, Amount: 500},
		},
	}
	// Create the transaction and allow it to overdraft
	if err := repo.createTransaction(tx, createTransactionOpts{AllowOverdraft: true}); err != nil {
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

// TestSqliteTransactionRepository__DisallowOverdraft will attempt an internal transfer, but be rejected on an overdraft error
func TestSqliteTransactionRepository__DisallowOverdraft(t *testing.T) {
	repo := createTestSqliteTransactionRepository(t)
	defer repo.Close()

	// Override the accountRepository and write our gl.Accounts
	account1, account2 := base.ID(), base.ID()
	repo.sqliteTransactionRepository.accountRepo = &testAccountRepository{
		accounts: []*gl.Account{
			// Setup the account being debited from as 'internal' (routing number we manage).
			{ID: account1, AccountNumber: "123", RoutingNumber: defaultRoutingNumber},
			{ID: account2, AccountNumber: "432", RoutingNumber: "121042882"},
		},
	}

	// Attempt our transaction
	tx := transaction{
		ID:        base.ID(),
		Timestamp: time.Now(),
		Lines: []transactionLine{
			{AccountId: account1, Purpose: ACHDebit, Amount: -500},
			{AccountId: account2, Purpose: ACHCredit, Amount: 500},
		},
	}

	// run the transfer without AllowOverdraft to encounter 'has insufficient funds' error
	if err := repo.createTransaction(tx, createTransactionOpts{AllowOverdraft: false}); err == nil {
		t.Error("expected error")
	} else {
		if !strings.Contains(err.Error(), "has insufficient funds") {
			t.Errorf("unknown error: %v", err)
		}
	}
}

func TestTransactions__isInteranlDebit(t *testing.T) {
	account1, account2 := base.ID(), base.ID()
	accounts := []*gl.Account{
		// Setup the account being debited from as 'remote' (routing number we don't manage)
		// so we can send the ACH file and possibly get a return.
		{ID: account1, AccountNumber: "123", RoutingNumber: "121042882"},
		{ID: account2, AccountNumber: "432", RoutingNumber: defaultRoutingNumber},
	}
	lines := []transactionLine{
		{AccountId: account1, Purpose: ACHDebit, Amount: -500},
		{AccountId: account2, Purpose: ACHCredit, Amount: 500},
	}
	if isInteranlDebit(accounts, lines, defaultRoutingNumber) {
		t.Errorf("account1 is external")
	}

	// swap routing numbers
	accounts[0].RoutingNumber = defaultRoutingNumber
	accounts[1].RoutingNumber = "121042882"

	if !isInteranlDebit(accounts, lines, defaultRoutingNumber) {
		t.Errorf("account1 is internal")
	}

	// no accounts
	if !isInteranlDebit(nil, nil, "") {
		t.Errorf("default should assume an internal transfer")
	}
}

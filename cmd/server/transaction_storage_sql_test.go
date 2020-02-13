// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	accounts "github.com/moov-io/accounts/client"
	"github.com/moov-io/accounts/cmd/server/database"
	"github.com/moov-io/base"

	"github.com/go-kit/kit/log"
)

func createTestSqlTransactionRepository(t *testing.T, db *sql.DB) *sqlTransactionRepository {
	t.Helper()

	repo, err := setupSqlTransactionStorage(context.Background(), log.NewNopLogger(), "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	repo.Close()
	repo.db = db
	if r, ok := repo.accountRepo.(*sqlAccountRepository); ok {
		r.db = db
	}
	return repo
}

func TestSqlTransactionRepository__Ping(t *testing.T) {
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()

	repo := createTestSqlTransactionRepository(t, sqliteDB.DB)
	defer repo.Close()

	if err := repo.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestSqlTransactionRepository(t *testing.T) {
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()

	repo := createTestSqlTransactionRepository(t, sqliteDB.DB)
	defer repo.Close()

	// Override the accountRepository and write our accounts
	account1, account2 := base.ID(), base.ID()
	repo.accountRepo = &testAccountRepository{
		accounts: []*accounts.Account{
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
			{AccountID: account1, Purpose: ACHDebit, Amount: 500},
			{AccountID: account2, Purpose: ACHCredit, Amount: 500},
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

	dbtx, _ := repo.db.Begin()
	defer dbtx.Rollback()

	bal, err := repo.getAccountBalance(dbtx, account1)
	if err != nil || bal != -500 {
		t.Errorf("got balance of %d", bal)
	}
	bal, err = repo.getAccountBalance(dbtx, account2)
	if err != nil || bal != 500 {
		t.Errorf("got balance of %d", bal)
	}

	// Grab our transaction by its ID
	transaction, err := repo.getTransaction(tx.ID)
	if err != nil || transaction == nil {
		t.Fatalf("transaction=%v error=%v", transaction, err)
	}
	if err := transaction.validate(); err != nil {
		t.Fatal(err)
	}
}

// TestSqlTransactionRepository__Internal will create an internal transfer
func TestSqlTransactionRepository__Internal(t *testing.T) {
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()

	repo := createTestSqlTransactionRepository(t, sqliteDB.DB)
	defer repo.Close()

	// Override the accountRepository and write our accounts
	account1, account2 := base.ID(), base.ID()
	repo.accountRepo = &testAccountRepository{
		accounts: []*accounts.Account{
			// Setup the account being debited from as 'internal' (routing number we manage).
			{ID: account1, AccountNumber: "123", RoutingNumber: defaultRoutingNumber},
			{ID: account2, AccountNumber: "432", RoutingNumber: defaultRoutingNumber},
		},
	}

	// Add initial funds
	tx := transaction{
		ID:        base.ID(),
		Timestamp: time.Now(),
		Lines: []transactionLine{
			{AccountID: account1, Purpose: ACHCredit, Amount: 1000},
		},
	}
	if err := repo.createTransaction(tx, createTransactionOpts{InitialDeposit: true}); err != nil {
		t.Fatal(err)
	}
	dbtx, _ := repo.db.Begin()
	if bal, _ := repo.getAccountBalance(dbtx, account1); bal != 1000 {
		t.Fatalf("account1=%s has unexpected balance of %d", account1, bal)
	}
	if bal, _ := repo.getAccountBalance(dbtx, account2); bal != 0 {
		t.Fatalf("account2=%s has unexpected balance of %d", account2, bal)
	}
	dbtx.Rollback()

	// Attempt our transaction
	tx = transaction{
		ID:        base.ID(),
		Timestamp: time.Now(),
		Lines: []transactionLine{
			{AccountID: account1, Purpose: ACHDebit, Amount: 400},
			{AccountID: account2, Purpose: ACHCredit, Amount: 400},
		},
	}
	// Create the transaction and allow it to overdraft
	if err := repo.createTransaction(tx, createTransactionOpts{}); err != nil {
		t.Logf("account1=%s account2=%s", account1, account2)
		t.Fatal(err)
	}

	transactions, err := repo.getAccountTransactions(account1)
	if err != nil {
		t.Error(err)
	}
	if len(transactions) != 2 {
		t.Errorf("got %d transactions: %v", len(transactions), transactions)
	}
	if transactions[0].ID != tx.ID || len(transactions[0].Lines) != 2 {
		t.Errorf("%#v", transactions[0])
	}

	dbtx, _ = repo.db.Begin()
	bal, err := repo.getAccountBalance(dbtx, account1)
	if err != nil || bal != 600 {
		t.Errorf("got balance of %d", bal)
	}
	bal, err = repo.getAccountBalance(dbtx, account2)
	if err != nil || bal != 400 {
		t.Errorf("got balance of %d", bal)
	}
	dbtx.Rollback()
}

// TestSqlTransactionRepository__AllowOverdraft will create an internal transfer, but allow an overdraft to occur
func TestSqlTransactionRepository__AllowOverdraft(t *testing.T) {
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()

	repo := createTestSqlTransactionRepository(t, sqliteDB.DB)
	defer repo.Close()

	// Override the accountRepository and write our accounts
	account1, account2 := base.ID(), base.ID()
	repo.accountRepo = &testAccountRepository{
		accounts: []*accounts.Account{
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
			{AccountID: account1, Purpose: ACHDebit, Amount: 500},
			{AccountID: account2, Purpose: ACHCredit, Amount: 500},
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

	dbtx, _ := repo.db.Begin()
	defer dbtx.Rollback()

	bal, err := repo.getAccountBalance(dbtx, account1)
	if err != nil || bal != -500 {
		t.Errorf("got balance of %d", bal)
	}
	bal, err = repo.getAccountBalance(dbtx, account2)
	if err != nil || bal != 500 {
		t.Errorf("got balance of %d", bal)
	}
}

// TestSqlTransactionRepository__DisallowOverdraft will attempt an internal transfer, but be rejected on an overdraft error
func TestSqlTransactionRepository__DisallowOverdraft(t *testing.T) {
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()

	repo := createTestSqlTransactionRepository(t, sqliteDB.DB)
	defer repo.Close()

	// Override the accountRepository and write our accounts
	account1, account2 := base.ID(), base.ID()
	repo.accountRepo = &testAccountRepository{
		accounts: []*accounts.Account{
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
			{AccountID: account1, Purpose: ACHDebit, Amount: 500},
			{AccountID: account2, Purpose: ACHCredit, Amount: 500},
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

func TestTransactions__isInternalDebit(t *testing.T) {
	account1, account2 := base.ID(), base.ID()
	accounts := []*accounts.Account{
		// Setup the account being debited from as 'remote' (routing number we don't manage)
		// so we can send the ACH file and possibly get a return.
		{ID: account1, AccountNumber: "123", RoutingNumber: "121042882"},
		{ID: account2, AccountNumber: "432", RoutingNumber: defaultRoutingNumber},
	}
	lines := []transactionLine{
		{AccountID: account1, Purpose: ACHDebit, Amount: 500},
		{AccountID: account2, Purpose: ACHCredit, Amount: 500},
	}
	if isInternalDebit(accounts, lines, defaultRoutingNumber) {
		t.Errorf("account1 is external")
	}

	// swap routing numbers
	accounts[0].RoutingNumber = defaultRoutingNumber
	accounts[1].RoutingNumber = "121042882"

	if !isInternalDebit(accounts, lines, defaultRoutingNumber) {
		t.Errorf("account1 is internal")
	}

	// no accounts
	if !isInternalDebit(nil, nil, "") {
		t.Errorf("default should assume an internal transfer")
	}
}

// TestSqlTransactions_unique ensures we can't insert a transaction with multiple lines for the same accountID
func TestSqlTransactions_unique(t *testing.T) {
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()

	repo := createTestSqlTransactionRepository(t, sqliteDB.DB)
	defer repo.Close()

	account1, account2 := base.ID(), base.ID()
	lines := []transactionLine{
		// Valid transaction, but has multiple lines for the same accountID
		{AccountID: account1, Purpose: ACHDebit, Amount: 500},
		{AccountID: account1, Purpose: ACHDebit, Amount: 100},
		{AccountID: account2, Purpose: ACHCredit, Amount: 600},
	}
	tx := transaction{
		ID:        base.ID(),
		Timestamp: time.Now(),
		Lines:     lines,
	}

	// Attempt our (invalid) transaction
	if err := repo.createTransaction(tx, createTransactionOpts{AllowOverdraft: true}); err == nil {
		t.Fatal("expected error")
	} else {
		if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
			t.Errorf("unknown error: %v", err)
		}
	}
}

// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	accounts "github.com/moov-io/accounts/client"
	"github.com/moov-io/accounts/cmd/server/database"
	"github.com/moov-io/base"

	"github.com/go-kit/kit/log"
)

func createTestSqlAccountRepository(t *testing.T, db *sql.DB) *sqlAccountRepository {
	t.Helper()

	repo, err := setupSqlAccountStorage(context.Background(), log.NewNopLogger(), "sqlite")
	if err != nil {
		t.Fatal(err)
	}
	repo.Close()

	repo.db = db
	repo.transactionRepo = &sqlTransactionRepository{db, log.NewNopLogger(), repo}

	return repo
}

func TestSqlAccountRepository_Ping(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, repo *sqlAccountRepository) {
		defer repo.Close()

		if err := repo.Ping(); err != nil {
			t.Fatal(err)
		}
	}

	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, createTestSqlAccountRepository(t, sqliteDB.DB))

	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, createTestSqlAccountRepository(t, mysqlDB.DB))
}

func TestSqlAccountRepository(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, repo *sqlAccountRepository) {
		defer repo.Close()

		customerID, now := base.ID(), time.Now()
		future := now.Add(24 * time.Hour)
		account := &accounts.Account{
			ID:            base.ID(),
			CustomerID:    customerID,
			Name:          "test account",
			AccountNumber: "12411",
			RoutingNumber: "219871289",
			Status:        "open",
			Type:          "Savings",
			CreatedAt:     time.Now(),
			ClosedAt:      future,
			LastModified:  now,
		}
		if err := repo.CreateAccount(customerID, account); err != nil {
			t.Fatal(err)
		}

		otherAccount := &accounts.Account{
			ID:            base.ID(),
			CustomerID:    base.ID(),
			Name:          "other account",
			AccountNumber: "18412481",
			RoutingNumber: "219871289",
			Status:        "open",
			Type:          "Checking",
			CreatedAt:     time.Now(),
		}
		if err := repo.CreateAccount(otherAccount.CustomerID, otherAccount); err != nil {
			t.Fatal(err)
		}

		t.Logf("A: %#v", repo.Ping())
		t.Logf("B: %#v", repo.transactionRepo.Ping())
		if rr, ok := repo.transactionRepo.accountRepo.(*sqlAccountRepository); ok {
			t.Logf("C: %#v", rr.Ping())
			t.Logf("D: %#v", rr.transactionRepo.Ping())
		}

		// read via one method
		accounts, err := repo.GetAccounts([]string{account.ID})
		if err != nil {
			t.Error(err)
		}
		if len(accounts) != 1 {
			t.Fatalf("got %d accounts: %#v", len(accounts), accounts)
		}
		if accounts[0].ID != account.ID {
			t.Errorf("Got %s", accounts[0].ID)
		}

		// and read via another
		accounts, err = repo.SearchAccountsByCustomerID(account.CustomerID)
		if err != nil {
			t.Error(err)
		}
		if len(accounts) != 1 {
			t.Fatalf("got %d accounts: %#v", len(accounts), accounts)
		}
		if accounts[0].ID != account.ID {
			t.Errorf("Got %s", accounts[0].ID)
		}

		// finally via a third method
		acct, err := repo.SearchAccountsByRoutingNumber(otherAccount.AccountNumber, otherAccount.RoutingNumber, otherAccount.Type)
		if err != nil {
			t.Fatal(err)
		}
		if acct.ID != otherAccount.ID {
			t.Errorf("found account %q", acct.ID)
		}

		// Change the case of otherAccount.Type
		acct, err = repo.SearchAccountsByRoutingNumber(otherAccount.AccountNumber, otherAccount.RoutingNumber, "checKIng")
		if err != nil {
			t.Fatal(err)
		}
		if acct.ID != otherAccount.ID {
			t.Errorf("found account %q", acct.ID)
		}
	}

	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, createTestSqlAccountRepository(t, sqliteDB.DB))

	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, createTestSqlAccountRepository(t, mysqlDB.DB))
}

func TestSqlAccounts__GetAccounts(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, repo *sqlAccountRepository) {
		defer repo.Close()

		accounts, err := repo.GetAccounts(nil)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if len(accounts) != 0 {
			t.Errorf("unexpected %d accounts: %v", len(accounts), accounts)
		}
	}

	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, createTestSqlAccountRepository(t, sqliteDB.DB))

	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, createTestSqlAccountRepository(t, mysqlDB.DB))
}

// TestSqlAccountRepository_unique will ensure we can't insert multiple accounts
// with the same account and routing numbers.
func TestSqlAccountRepository_unique(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, repo *sqlAccountRepository) {
		defer repo.Close()

		customerID, now := base.ID(), time.Now()
		future := now.Add(24 * time.Hour)
		account := &accounts.Account{
			ID:            base.ID(),
			CustomerID:    customerID,
			Name:          "test account",
			AccountNumber: "12411",
			RoutingNumber: "219871289",
			Status:        "open",
			Type:          "Savings",
			CreatedAt:     time.Now(),
			ClosedAt:      future,
			LastModified:  now,
		}
		if err := repo.CreateAccount(customerID, account); err != nil {
			t.Fatal(err)
		}

		// attempt again
		account.ID = base.ID()
		if err := repo.CreateAccount(customerID, account); err == nil {
			t.Error("expected error")
		} else {
			if !database.UniqueViolation(err) {
				t.Errorf("unknown error: %v", err)
			}
		}
	}

	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, createTestSqlAccountRepository(t, sqliteDB.DB))

	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, createTestSqlAccountRepository(t, mysqlDB.DB))
}

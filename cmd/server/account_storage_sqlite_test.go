// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	accounts "github.com/moov-io/accounts/client"
	"github.com/moov-io/accounts/cmd/server/database"
	"github.com/moov-io/base"

	"github.com/go-kit/kit/log"
)

type testSqliteAccountRepository struct {
	*sqliteAccountRepository

	db *database.TestSQLiteDB
}

func (r *testSqliteAccountRepository) Close() error {
	r.db.Close()
	return r.sqliteAccountRepository.Close()
}

func createTestSqliteAccountRepository(t *testing.T) *testSqliteAccountRepository {
	t.Helper()

	db := database.CreateTestSqliteDB(t)
	repo, err := setupSqliteAccountStorage(log.NewNopLogger(), filepath.Join(db.Dir, "accounts.db"))
	if err != nil {
		t.Fatal(err)
	}
	return &testSqliteAccountRepository{repo, db}
}

func TestSqliteAccountRepository_Ping(t *testing.T) {
	repo := createTestSqliteAccountRepository(t)
	defer repo.Close()

	if err := repo.Ping(); err != nil {
		t.Fatal(err)
	}
}

func TestSqliteAccountRepository(t *testing.T) {
	repo := createTestSqliteAccountRepository(t)
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

func TestSqliteAccounts__GetAccounts(t *testing.T) {
	repo := createTestSqliteAccountRepository(t)
	defer repo.Close()

	accounts, err := repo.GetAccounts(nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("unexpected %d accounts: %v", len(accounts), accounts)
	}
}

// TestSqliteAccountRepository_unique will ensure we can't insert multiple accounts
// with the same account and routing numbers.
func TestSqliteAccountRepository_unique(t *testing.T) {
	repo := createTestSqliteAccountRepository(t)
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
		if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
			t.Errorf("unknown error: %v", err)
		}
	}
}

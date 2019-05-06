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

type testSqliteAccountRepository struct {
	*sqliteAccountRepository

	db *testSqliteDB
}

func (r *testSqliteAccountRepository) Close() error {
	r.db.close()
	return r.sqliteAccountRepository.Close()
}

func createTestSqliteAccountRepository(t *testing.T) *testSqliteAccountRepository {
	t.Helper()

	db, err := createTestSqliteDB()
	if err != nil {
		t.Fatal(err)
	}
	repo, err := setupSqliteAccountStorage(log.NewNopLogger(), filepath.Join(db.dir, "gl.db"))
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

	customerId, now := base.ID(), time.Now()
	future := now.Add(24 * time.Hour)
	account := &gl.Account{
		ID:            base.ID(),
		CustomerID:    customerId,
		Name:          "test account",
		AccountNumber: "12411",
		RoutingNumber: "219871289",
		Status:        "open",
		Type:          "Savings",
		CreatedAt:     time.Now(),
		ClosedAt:      &future,
		LastModified:  &now,
	}
	if err := repo.CreateAccount(customerId, account); err != nil {
		t.Fatal(err)
	}

	otherAccount := &gl.Account{
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
	accounts, err = repo.GetCustomerAccounts(account.CustomerID)
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
	acct, err := repo.SearchAccounts(otherAccount.AccountNumber, otherAccount.RoutingNumber, otherAccount.Type)
	if err != nil {
		t.Fatal(err)
	}
	if acct.ID != otherAccount.ID {
		t.Errorf("found account %q", acct.ID)
	}
}

// TestSqliteAccountRepository_unique will ensure we can't insert multiple accounts
// with the same account and routing numbers.
func TestSqliteAccountRepository_unique(t *testing.T) {
	repo := createTestSqliteAccountRepository(t)
	defer repo.Close()

	customerId, now := base.ID(), time.Now()
	future := now.Add(24 * time.Hour)
	account := &gl.Account{
		ID:            base.ID(),
		CustomerID:    customerId,
		Name:          "test account",
		AccountNumber: "12411",
		RoutingNumber: "219871289",
		Status:        "open",
		Type:          "Savings",
		CreatedAt:     time.Now(),
		ClosedAt:      &future,
		LastModified:  &now,
	}
	if err := repo.CreateAccount(customerId, account); err != nil {
		t.Fatal(err)
	}

	// attempt again
	account.ID = base.ID()
	if err := repo.CreateAccount(customerId, account); err == nil {
		t.Error("expected error")
	} else {
		if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
			t.Errorf("unknown error: %v", err)
		}
	}
}

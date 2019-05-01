// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
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
	db, err := createTestSqliteDB()
	if err != nil {
		t.Fatal(err)
	}
	return &testSqliteAccountRepository{&sqliteAccountRepository{db.db, log.NewNopLogger()}, db}
}

func TestSqliteAccountRepository(t *testing.T) {
	repo := createTestSqliteAccountRepository(t)
	defer repo.Close()

	customerId := base.ID()
	account := &gl.Account{
		ID:            base.ID(),
		CustomerID:    customerId,
		Name:          "test account",
		AccountNumber: "12411",
		RoutingNumber: "219871289",
		Status:        "open",
		Type:          "Savings",
		CreatedAt:     time.Now(),
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

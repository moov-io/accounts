// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/gl"
)

var (
	// qledgerAddress and qledgerAuthToken are the same env variables as account_stroage.go
	qledgerAddress   = os.Getenv("QLEDGER_ENDPOINT")
	qledgerAuthToken = os.Getenv("QLEDGER_AUTH_TOKEN")
)

// qualifyTests will skip tests if Go's test -short flag is specified or if
// the needed env variables aren't set. See above for the env variables.
//
// Returned will be a qledgerAccountRepository
func qualifyTests(t *testing.T, address, authToken string) *qledgerAccountRepository {
	t.Helper()
	if qledgerAddress == "" || qledgerAuthToken == "" || testing.Short() {
		t.Skip()
	}
	repo, err := setupQLedgerStorage(address, authToken)
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := repo.(*qledgerAccountRepository); ok {
		return r
	}
	t.Fatalf("unknown repo returned: %T", repo)
	return nil
}

func TestQLedger__ping(t *testing.T) {
	repo := qualifyTests(t, qledgerAddress, qledgerAuthToken)
	if err := repo.Ping(); err != nil {
		t.Error(err)
	}
}

func TestQLedger__Accounts(t *testing.T) {
	repo := qualifyTests(t, qledgerAddress, qledgerAuthToken)

	customerId := base.ID()
	account := &gl.Account{
		ID:            base.ID(),
		CustomerID:    customerId,
		Name:          "example account",
		AccountNumber: createAccountNumber(),
		RoutingNumber: "121042882",
		Status:        "Active",
		Type:          "Checking",
	}
	if err := repo.CreateAccount(customerId, account); err != nil {
		t.Error(err)
	}

	// Now grab accounts for this customer
	accounts, err := repo.GetCustomerAccounts(customerId)
	if err != nil {
		t.Error(err)
	}
	if len(accounts) == 0 {
		t.Fatal("no accounts found")
	}
	if account.ID != accounts[0].ID {
		t.Errorf("expected account %q, but found %#v", account.ID, accounts[0].ID)
	}

	// Search for account
	acct, err := repo.SearchAccounts(account.AccountNumber, account.RoutingNumber, "Checking")
	if err != nil {
		t.Fatal(err)
	}
	if acct == nil {
		t.Fatal("SearchAccounts: nil account")
	}
	if acct.ID != account.ID {
		t.Errorf("acct.ID=%q account.ID=%q", acct.ID, account.ID)
	}
}

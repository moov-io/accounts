// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"testing"

	"github.com/moov-io/base"
)

// qualifyQLedgerTransactionTest will skip tests if Go's test -short flag is specified or if
// the needed env variables aren't set. See above for the env variables.
//
// Returned will be a qledgerAccountRepository
func qualifyQLedgerTransactionTest(t *testing.T) *qledgerTransactionRepository {
	t.Helper()
	// qledgerEndpoint and qledgerAuthToken are from account_storage_qledger_test.go
	if qledgerEndpoint == "" || qledgerAuthToken == "" || testing.Short() {
		t.Skip()
	}
	repo, err := setupQLedgerTransactionStorage(qledgerEndpoint, qledgerAuthToken)
	if err != nil {
		t.Fatal(err)
	}
	return repo
}

func TestQLedgerTransactions__ping(t *testing.T) {
	repo := qualifyQLedgerTransactionTest(t)
	if err := repo.Ping(); err != nil {
		t.Error(err)
	}
}

func TestQLedgerTransactions(t *testing.T) {
	accountId := base.ID()
	transactionRepo := qualifyQLedgerTransactionTest(t)

	// Create a transaction
	tx := (&createTransactionRequest{
		Lines: []transactionLine{
			{
				AccountId: accountId,
				Purpose:   ACHCredit,
				Amount:    1242,
			},
			{
				AccountId: base.ID(),
				Purpose:   ACHDebit,
				Amount:    -1242,
			},
		},
	}).asTransaction(base.ID())

	if err := transactionRepo.createTransaction(tx); err != nil {
		t.Fatal(err)
	}

	transactions, err := transactionRepo.getAccountTransactions(accountId)
	if err != nil {
		t.Fatal(err)
	}
	if len(transactions) != 1 {
		t.Fatalf("len(transactions)=%d", len(transactions))
	}
	if len(transactions[0].Lines) != 2 {
		t.Errorf("len(transactions[0].Lines)=%d", len(transactions[0].Lines))
	}

	for i := range transactions[0].Lines {
		if transactions[0].Lines[i].AccountId == accountId {
			if transactions[0].Lines[i].Purpose != ACHCredit || transactions[0].Lines[i].Amount != 1242 {
				t.Errorf("purpose=%q amount=%d", transactions[0].Lines[i].Purpose, transactions[0].Lines[i].Amount)
			}
		} else {
			if transactions[0].Lines[i].Purpose != ACHDebit || transactions[0].Lines[i].Amount != -1242 {
				t.Errorf("purpose=%q amount=%d", transactions[0].Lines[i].Purpose, transactions[0].Lines[i].Amount)
			}
		}
	}
}
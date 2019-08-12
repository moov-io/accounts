// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"testing"

	"github.com/moov-io/base"
	mledge "github.com/moov-io/qledger-sdk-go"
)

type testQLedgerTransactionRepository struct {
	*qledgerTransactionRepository

	deployment *qledgerDeployment
}

func (q *testQLedgerTransactionRepository) close() {
	if q.deployment != nil {
		q.deployment.close()
	}
}

// qualifyQLedgerTransactionTest will skip tests if Go's test -short flag is specified or if
// the needed env variables aren't set. See above for the env variables.
func qualifyQLedgerTransactionTest(t *testing.T) *testQLedgerTransactionRepository {
	t.Helper()

	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	deployment := spawnQLedger(t)
	if deployment == nil {
		t.Fatal("nil QLedger docker deployment")
	}

	// repo, err := setupQLedgerTransactionStorage("https://api.moov.io/v1/qledger", "moov") // Test against Production
	repo, err := setupQLedgerTransactionStorage(fmt.Sprintf("http://localhost:%s", deployment.qledger.GetPort("7000/tcp")), "moov")
	if repo == nil || err != nil {
		t.Fatalf("repo=%v error=%v", repo, err)
	}
	if err := repo.Close(); err != nil { // should do nothing, so call in every test to make sure
		t.Fatal("QLedger .Close() is a no-op")
	}
	return &testQLedgerTransactionRepository{repo, deployment}
}

func TestQLedgerTransactions__ping(t *testing.T) {
	repo := qualifyQLedgerTransactionTest(t)
	defer repo.close()

	if err := repo.Ping(); err != nil {
		t.Error(err)
	}
}

func TestQLedger__grabAccountIDs(t *testing.T) {
	ids := grabAccountIDs([]transactionLine{
		{
			AccountID: "accountId1",
			Purpose:   ACHCredit,
			Amount:    1242,
		},
		{
			AccountID: "accountId2",
			Purpose:   ACHDebit,
			Amount:    1242,
		},
	})
	if len(ids) != 2 {
		t.Fatalf("got %d ids: %v", len(ids), ids)
	}
	if ids[0] != "accountId1" || ids[1] != "accountId2" {
		t.Errorf("got %v", ids)
	}
}

func TestQLedgerTransactions(t *testing.T) {
	accountID := base.ID()
	transactionRepo := qualifyQLedgerTransactionTest(t)

	// Create a transaction
	tx := (&createTransactionRequest{
		Lines: []transactionLine{
			{
				AccountID: accountID,
				Purpose:   ACHCredit,
				Amount:    1242,
			},
			{
				AccountID: base.ID(),
				Purpose:   ACHDebit,
				Amount:    1242,
			},
		},
	}).asTransaction(base.ID())

	if err := transactionRepo.createTransaction(tx, createTransactionOpts{AllowOverdraft: false}); err != nil {
		t.Fatal(err)
	}

	transactions, err := transactionRepo.getAccountTransactions(accountID)
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
		if transactions[0].Lines[i].AccountID == accountID {
			if transactions[0].Lines[i].Purpose != ACHCredit || transactions[0].Lines[i].Amount != 1242 {
				t.Errorf("purpose=%q amount=%d", transactions[0].Lines[i].Purpose, transactions[0].Lines[i].Amount)
			}
		} else {
			if transactions[0].Lines[i].Purpose != ACHDebit || transactions[0].Lines[i].Amount != 1242 {
				t.Errorf("purpose=%q amount=%d", transactions[0].Lines[i].Purpose, transactions[0].Lines[i].Amount)
			}
		}
	}

	// Grab our transaction by its ID
	transaction, err := transactionRepo.getTransaction(tx.ID)
	if err != nil || transaction == nil {
		t.Fatalf("transaction=%v error=%v", transaction, err)
	}
	if err := transaction.validate(); err != nil {
		t.Fatal(err)
	}

	// Only cleanup if every test was successful, otherwise keep the containers around for debugging
	transactionRepo.close()
}

func TestQLedger__convertQLedgerTransactions(t *testing.T) {
	a1, a2 := base.ID(), base.ID()
	incoming := []*mledge.Transaction{
		{
			ID: "19defa381fa7125212a430b6e441f04c48618281",
			Data: map[string]interface{}{
				"accountIDs": []interface{}{a1, a2},
			},
			Timestamp: "2019-05-21T16:55:53.933Z",
			Lines: []*mledge.TransactionLine{
				{AccountID: a1, Delta: 1000},
				{AccountID: a2, Delta: -1000},
			},
		},
	}
	out := convertQLedgerTransactions(incoming)
	if len(out) != 1 {
		t.Errorf("got %d transactions", len(out))
	}

	// Check the reversal fields
	if out[0].ID != incoming[0].ID {
		t.Errorf("got %s", out[0].ID)
	}
	for i := range out[0].Lines {
		if out[0].Lines[i].AccountID == a1 && out[0].Lines[i].Amount != 1000 {
			t.Errorf("a1: unexpected amount %d", out[0].Lines[i].Amount)
		}
		if out[0].Lines[i].AccountID == a2 && out[0].Lines[i].Amount != 1000 {
			t.Errorf("a2: unexpected amount %d", out[0].Lines[i].Amount)
		}
	}
}

// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"testing"

	"github.com/moov-io/base"
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
	if err != nil {
		t.Fatal(err)
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

func TestQLedger__accountIds(t *testing.T) {
	ids := accountIds([]transactionLine{
		{
			AccountId: "accountId1",
			Purpose:   ACHCredit,
			Amount:    1242,
		},
		{
			AccountId: "accountId2",
			Purpose:   ACHDebit,
			Amount:    -1242,
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

	// Only cleanup if every test was successful, otherwise keep the containers around for debugging
	transactionRepo.close()
}

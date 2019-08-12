// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"time"

	mledge "github.com/moov-io/qledger-sdk-go"

	"github.com/RealImage/QLedger/models"
)

func setupQLedgerTransactionStorage(endpoint, apiToken string) (*qledgerTransactionRepository, error) {
	if endpoint == "" || apiToken == "" {
		return nil, fmt.Errorf("qledger: empty endpoint=%q and/or apiToken", endpoint)
	}
	return &qledgerTransactionRepository{
		api: mledge.NewLedger(endpoint, apiToken),
	}, nil
}

type qledgerTransactionRepository struct {
	api *mledge.Ledger
}

func (r *qledgerTransactionRepository) Ping() error {
	return r.api.Ping()
}

func (r *qledgerTransactionRepository) Close() error {
	return nil
}

func (r *qledgerTransactionRepository) createTransaction(tx transaction, opts createTransactionOpts) error {
	var lines []*mledge.TransactionLine
	data := make(map[string]interface{})
	data["accountIds"] = grabAccountIDs(tx.Lines) // old name of '*Id*'

	for i := range tx.Lines {
		line := &mledge.TransactionLine{
			AccountID: tx.Lines[i].AccountID,
		}
		if tx.Lines[i].Purpose == ACHDebit {
			line.Delta = -1 * tx.Lines[i].Amount
		} else {
			line.Delta = tx.Lines[i].Amount
		}
		lines = append(lines, line)
		// TODO(adam): https://github.com/RealImage/QLedger/issues/40
		// data[fmt.Sprintf("%s_purpose", tx.Lines[i].AccountId)] = tx.Lines[i].Purpose
	}

	return r.api.CreateTransaction(&mledge.Transaction{
		ID:        tx.ID,
		Data:      data,
		Timestamp: tx.Timestamp.Format(models.LedgerTimestampLayout),
		Lines:     lines,
	})
}

func (r *qledgerTransactionRepository) getAccountTransactions(accountId string) ([]transaction, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"must": map[string]interface{}{
				"terms": []map[string]interface{}{
					{
						"accountIds": []string{accountId},
					},
				},
			},
		},
	}
	xfers, err := r.api.SearchTransactions(query)
	if err != nil {
		return nil, fmt.Errorf("qledger: getAccountTransactions: %v", err)
	}
	return convertQLedgerTransactions(xfers), nil
}

func (r *qledgerTransactionRepository) getTransaction(transactionId string) (*transaction, error) {
	tx, err := r.api.GetTransaction(transactionId)
	if err != nil {
		return nil, fmt.Errorf("qledger: getTransaction: %v", err)
	}
	out := convertQLedgerTransactions([]*mledge.Transaction{tx})[0]
	return &out, nil
}

func convertQLedgerTransactions(xfers []*mledge.Transaction) []transaction {
	var transactions []transaction
	for i := range xfers {
		var lines []transactionLine
		for j := range xfers[i].Lines {
			// TODO(adam): https://github.com/RealImage/QLedger/issues/40
			// p, _ := xfers[i].Data[fmt.Sprintf("%s_purpose", xfers[i].Lines[j].AccountID)].(string)
			delta := xfers[i].Lines[j].Delta
			p := "achcredit"
			if delta < 0 {
				p = "achdebit" // TODO(adam): mocked for tests, see commented '%s_purpose' above
				delta *= -1
			}
			tx := transactionLine{
				AccountID: xfers[i].Lines[j].AccountID,
				Amount:    delta,
				Purpose:   TransactionPurpose(p),
			}
			if err := tx.Purpose.validate(); err != nil {
				continue
			}
			lines = append(lines, tx)
		}
		t, err := time.Parse(models.LedgerTimestampLayout, xfers[i].Timestamp)
		if err != nil {
			// Try another format that seems to show up in QLedger reads
			t, _ = time.Parse("2006-01-02T15:04:05.000Z", xfers[i].Timestamp)
		}
		if t.IsZero() {
			continue // really invalid format
		}
		transactions = append(transactions, transaction{
			ID:        xfers[i].ID,
			Timestamp: t,
			Lines:     lines,
		})
	}
	return transactions
}

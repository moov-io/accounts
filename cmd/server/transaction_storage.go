// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"strings"

	"github.com/moov-io/gl"
)

type transactionRepository interface {
	Ping() error

	createTransaction(account *gl.Account, tx transaction) error
	getAccountTransactions(accountID string) ([]transaction, error)
}

func initTransactionStorage(name string) (transactionRepository, error) {
	switch strings.ToLower(name) {
	case "qledger":
		return setupQLedgerTransactionStorage(os.Getenv("QLEDGER_ENDPOINT"), os.Getenv("QLEDGER_AUTH_TOKEN"))
	}
	return nil, nil
}

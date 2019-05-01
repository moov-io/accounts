// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"strings"

	"github.com/go-kit/kit/log"
)

type transactionRepository interface {
	Ping() error
	Close() error

	createTransaction(tx transaction) error
	getAccountTransactions(accountID string) ([]transaction, error) // TODO(adam): limit and/or pagination params
}

func initTransactionStorage(logger log.Logger, name string) (transactionRepository, error) {
	switch strings.ToLower(name) {
	case "qledger":
		return setupQLedgerTransactionStorage(os.Getenv("QLEDGER_ENDPOINT"), os.Getenv("QLEDGER_AUTH_TOKEN"))
	case "sqlite":
		return setupSqliteTransactionStorage(logger, getSqlitePath())
	}
	return nil, nil
}

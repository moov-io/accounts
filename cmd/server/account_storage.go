// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"strings"

	accounts "github.com/moov-io/accounts/client"

	"github.com/go-kit/kit/log"
)

type accountRepository interface {
	Ping() error
	Close() error

	GetAccounts(accountIds []string) ([]*accounts.Account, error)
	GetCustomerAccounts(customerId string) ([]*accounts.Account, error)
	CreateAccount(customerId string, account *accounts.Account) error // TODO(adam): acctType needs strong type, we can drop customerId as it's on accounts.Account
	SearchAccounts(accountNumber, routingNumber, acctType string) (*accounts.Account, error)
}

func initAccountStorage(logger log.Logger, name string) (accountRepository, error) {
	switch strings.ToLower(name) {
	case "qledger":
		return setupQLedgerAccountStorage(os.Getenv("QLEDGER_ENDPOINT"), os.Getenv("QLEDGER_AUTH_TOKEN"))
	case "sqlite":
		return setupSqliteAccountStorage(logger, getSqlitePath())
	}
	return nil, nil
}

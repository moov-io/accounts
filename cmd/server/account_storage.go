// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"strings"

	accounts "github.com/moov-io/accounts/client"
	"github.com/moov-io/accounts/cmd/server/database"

	"github.com/go-kit/kit/log"
)

type accountRepository interface {
	Ping() error
	Close() error

	GetAccounts(accountIDs []string) ([]*accounts.Account, error)
	CreateAccount(customerID string, account *accounts.Account) error // TODO(adam): acctType needs strong type, we can drop customerID as it's on accounts.Account

	SearchAccountsByCustomerID(customerID string) ([]*accounts.Account, error)
	SearchAccountsByRoutingNumber(accountNumber, routingNumber, acctType string) (*accounts.Account, error)
}

func initAccountStorage(logger log.Logger, name string) (accountRepository, error) {
	switch strings.ToLower(name) {
	case "mysql":
		return setupMySQLAccountStorage(logger, os.Getenv("MYSQL_USER"), os.Getenv("MYSQL_PASSWORD"), os.Getenv("MYSQL_ADDRESS"), os.Getenv("MYSQL_DATABASE"))
	case "sqlite":
		return setupSqliteAccountStorage(logger, database.SQLitePath())
	}
	return nil, nil
}

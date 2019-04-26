// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"strings"

	"github.com/moov-io/gl"
)

type accountRepository interface {
	Ping() error

	GetAccounts(accountIds []string) ([]*gl.Account, error)
	GetCustomerAccounts(customerId string) ([]*gl.Account, error)
	CreateAccount(customerId string, account *gl.Account) error // TODO(adam): acctType needs strong type
	SearchAccounts(accountNumber, routingNumber, acctType string) (*gl.Account, error)
}

func initAccountStorage(name string) (accountRepository, error) {
	switch strings.ToLower(name) {
	case "qledger":
		return setupQLedgerAccountStorage(os.Getenv("QLEDGER_ENDPOINT"), os.Getenv("QLEDGER_AUTH_TOKEN"))
	}
	return nil, nil
}

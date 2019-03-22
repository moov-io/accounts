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

	GetCustomerAccounts(customerId string) ([]gl.Account, error)
	CreateAccount(customerId string, name string, acctType string) (*gl.Account, error) // TOOD(adam): acctType needs strong type
	SearchAccounts(accountNumber, routingNumber, acctType string) (*gl.Account, error)
}

func initAccountStorage(name string) (accountRepository, error) {
	switch strings.ToLower(name) {
	case "qledger":
		return setupQLedgerStorage(os.Getenv("LEDGER_ENDPOINT"), os.Getenv("LEDGER_AUTH_TOKEN"))
	}
	return nil, nil
}

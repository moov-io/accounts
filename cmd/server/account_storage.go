// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	accounts "github.com/moov-io/accounts/client"
)

type accountRepository interface {
	Ping() error
	Close() error

	GetAccounts(accountIDs []string) ([]*accounts.Account, error)
	CreateAccount(customerID string, account *accounts.Account) error // TODO(adam): acctType needs strong type, we can drop customerID as it's on accounts.Account

	SearchAccountsByCustomerID(customerID string) ([]*accounts.Account, error)
	SearchAccountsByRoutingNumber(accountNumber, routingNumber, acctType string) (*accounts.Account, error)
}

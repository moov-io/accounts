// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"github.com/moov-io/gl"
)

// testAccountRepository represents a mocked accountRepository where accounts or err are
// returned if set. Tests are fully responsible for managing state.
type testAccountRepository struct {
	accounts []*gl.Account

	err error
}

func (r *testAccountRepository) Ping() error {
	return r.err
}

func (r *testAccountRepository) Close() error {
	return r.err
}

func (r *testAccountRepository) GetAccounts(accountIds []string) ([]*gl.Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.accounts, nil
}

func (r *testAccountRepository) GetCustomerAccounts(customerId string) ([]*gl.Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.accounts, nil
}

func (r *testAccountRepository) CreateAccount(customerId string, account *gl.Account) error {
	return r.err
}

func (r *testAccountRepository) SearchAccounts(accountNumber, routingNumber, acctType string) (*gl.Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.accounts) > 0 {
		return r.accounts[0], nil
	}
	return nil, nil
}

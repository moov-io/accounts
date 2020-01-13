// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	accounts "github.com/moov-io/accounts/client"
)

// testAccountRepository represents a mocked accountRepository where accounts or err are
// returned if set. Tests are fully responsible for managing state.
type testAccountRepository struct {
	accounts []*accounts.Account

	err error
}

func (r *testAccountRepository) Ping() error {
	return r.err
}

func (r *testAccountRepository) Close() error {
	return r.err
}

func (r *testAccountRepository) GetAccounts(accountIDs []string) ([]*accounts.Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.accounts, nil
}

func (r *testAccountRepository) GetCustomerAccounts(customerID string) ([]*accounts.Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.accounts, nil
}

func (r *testAccountRepository) CreateAccount(customerID string, account *accounts.Account) error {
	return r.err
}

func (r *testAccountRepository) SearchAccountsByRoutingNumber(accountNumber, routingNumber, acctType string) (*accounts.Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.accounts) > 0 {
		return r.accounts[0], nil
	}
	return nil, nil
}

func (r *testAccountRepository) SearchAccountsByCustomerID(customerID string) ([]*accounts.Account, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.accounts, nil
}

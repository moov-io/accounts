// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"

	accounts "github.com/moov-io/accounts/client"

	"github.com/go-kit/kit/log"
)

type mysqlAccountRepository struct {
	db     *sql.DB
	logger log.Logger
}

func setupMySQLAccountStorage(logger log.Logger, user, password, address, database string) (*mysqlAccountRepository, error) {
	return nil, nil
}

func (r *mysqlAccountRepository) Ping() error {
	return r.db.Ping()
}

func (r *mysqlAccountRepository) Close() error {
	return r.db.Close()
}

func (r *mysqlAccountRepository) GetAccounts(accountIDs []string) ([]*accounts.Account, error) {
	return nil, nil
}

func (r *mysqlAccountRepository) CreateAccount(customerID string, account *accounts.Account) error {
	// TODO(adam): acctType needs strong type, we can drop customerID as it's on accounts.Account
	return nil
}

func (r *mysqlAccountRepository) SearchAccountsByCustomerID(customerID string) ([]*accounts.Account, error) {
	return nil, nil
}

func (r *mysqlAccountRepository) SearchAccountsByRoutingNumber(accountNumber, routingNumber, acctType string) (*accounts.Account, error) {
	return nil, nil
}

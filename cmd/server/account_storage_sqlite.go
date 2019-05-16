// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"fmt"
	"strings"

	accounts "github.com/moov-io/accounts/client"

	"github.com/go-kit/kit/log"
)

type sqliteAccountRepository struct {
	db     *sql.DB
	logger log.Logger

	transactionRepo *sqliteTransactionRepository
}

func setupSqliteAccountStorage(logger log.Logger, path string) (*sqliteAccountRepository, error) {
	db, err := createSqliteConnection(logger, path)
	if err != nil {
		return nil, err
	}
	transactionRepo, err := setupSqliteTransactionStorage(logger, path)
	if err != nil {
		return nil, fmt.Errorf("setupSqliteTransactionStorage: transactions: %v", err)
	}
	return &sqliteAccountRepository{db, logger, transactionRepo}, nil
}

func (r *sqliteAccountRepository) Ping() error {
	return r.db.Ping()
}

func (r *sqliteAccountRepository) Close() error {
	r.transactionRepo.Close()
	return r.db.Close()
}

func (r *sqliteAccountRepository) GetAccounts(accountIds []string) ([]*accounts.Account, error) {
	if n := len(accountIds); n <= 0 || n > 250 {
		return nil, fmt.Errorf("sqlite.GetAccounts: invalid amount of accountIds (%d)", n)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("sqlite.GetAccounts: tx.Begin: %v", err)
	}

	query := fmt.Sprintf(`select account_id, customer_id, name, account_number, routing_number, status, type, created_at, closed_at, last_modified
from accounts where account_id in (?%s) and deleted_at is null;`, strings.Repeat(",?", len(accountIds)-1))
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var ids []interface{}
	for i := range accountIds {
		ids = append(ids, accountIds[i])
	}
	rows, err := stmt.Query(ids...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*accounts.Account
	for rows.Next() {
		var a accounts.Account
		err := rows.Scan(&a.Id, &a.CustomerId, &a.Name, &a.AccountNumber, &a.RoutingNumber, &a.Status, &a.Type, &a.CreatedAt, &a.ClosedAt, &a.LastModified)
		if err != nil {
			return nil, fmt.Errorf("sqlite.GetAccounts: account=%q: %v", a.Id, err)
		}
		balance, err := r.transactionRepo.getAccountBalance(tx, a.Id)
		if err != nil {
			return nil, fmt.Errorf("sqlite.GetAccounts: getAccountBalance: account=%q: %v", a.Id, err)
		}
		// TODO(adam): need Balance, BalanceAvailable, and BalancePending
		a.Balance = balance
		out = append(out, &a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sqlite.GetAccounts: scan: %v", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("sqlite.GetAccounts: commit: %v", err)
	}
	return out, nil
}

func (r *sqliteAccountRepository) GetCustomerAccounts(customerId string) ([]*accounts.Account, error) {
	query := `select account_id from accounts where customer_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accountIds []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("sqlite.GetCustomerAccounts: account=%q: %v", id, err)
		}
		accountIds = append(accountIds, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return r.GetAccounts(accountIds)
}

func (r *sqliteAccountRepository) CreateAccount(customerId string, a *accounts.Account) error {
	query := `insert into accounts (account_id, customer_id, name, account_number, routing_number, status, type, created_at, closed_at, last_modified) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(a.Id, a.CustomerId, a.Name, a.AccountNumber, a.RoutingNumber, a.Status, a.Type, a.CreatedAt, a.ClosedAt, a.LastModified)
	return err
}

func (r *sqliteAccountRepository) SearchAccounts(accountNumber, routingNumber, acctType string) (*accounts.Account, error) {
	query := `select account_id from accounts where account_number = ? and routing_number = ? and lower(type) = lower(?) and deleted_at is null limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(accountNumber, routingNumber, acctType)
	var id string
	if err := row.Scan(&id); err != nil || id == "" {
		return nil, fmt.Errorf("sqlite.SearchAccounts: account=%q: %v", id, err)
	}

	// Grab out account by its ID
	accounts, err := r.GetAccounts([]string{id})
	if err != nil || len(accounts) == 0 {
		return nil, fmt.Errorf("sqlite.SearchAccounts: no accounts: %v", err)
	}
	return accounts[0], nil
}

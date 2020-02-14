// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	accounts "github.com/moov-io/accounts/client"
	"github.com/moov-io/accounts/cmd/server/database"

	"github.com/go-kit/kit/log"
)

type sqlTransactionRepository struct {
	db     *sql.DB
	logger log.Logger

	accountRepo accountRepository
}

func setupSqlTransactionStorage(ctx context.Context, logger log.Logger, _type string) (*sqlTransactionRepository, error) {
	db, err := database.New(ctx, logger, _type)
	if err != nil {
		return nil, err
	}
	// Break the cyclic dependency between account and transaction repositories
	repo := &sqlTransactionRepository{db: db, logger: logger}
	accountRepo := &sqlAccountRepository{db, logger, repo}
	repo.accountRepo = accountRepo
	return repo, nil
}

func (r *sqlTransactionRepository) Ping() error {
	return r.db.Ping()
}

func (r *sqlTransactionRepository) Close() error {
	return r.db.Close()
}

// isInternalDebit returns true only when the debited account's routing number matches
// the configured default routing number. This means we have to be accountable for choosing
// to allow an overdraft or not.
func isInternalDebit(accounts []*accounts.Account, lines []transactionLine, routingNumber string) bool {
	for i := range accounts {
		for j := range lines {
			if accounts[i].ID == lines[j].AccountID {
				switch lines[j].Purpose {
				case ACHDebit:
					return accounts[i].RoutingNumber == routingNumber
				}
			}
		}
	}
	return true // default to assuming we need to check/prevent an overdraft
}

func (r *sqlTransactionRepository) createTransaction(t transaction, opts createTransactionOpts) error {
	if err := t.validate(); err != nil && !opts.InitialDeposit {
		return fmt.Errorf("transaction=%q is invalid: %v", t.ID, err)
	}

	accounts, err := r.accountRepo.GetAccounts(grabAccountIDs(t.Lines))
	if err != nil {
		return fmt.Errorf("createTransaction: problem reading accounts for transaction=%q: %v", t.ID, err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("createTransaction: tx.Begin error=%v rollback=%v", err, tx.Rollback())
	}

	// insert transaction
	query := `insert into transactions(transaction_id, timestamp, created_at) values (?, ?, ?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("createTransaction: prepare: error=%v rollback=%v", err, tx.Rollback())
	}
	if _, err := stmt.Exec(t.ID, t.Timestamp, time.Now()); err != nil {
		stmt.Close()
		return fmt.Errorf("createTransaction: insert: error=%v rollback=%v", err, tx.Rollback())
	}
	stmt.Close()

	// insert each transactionLine
	for i := range t.Lines {
		query = `insert into transaction_lines(transaction_id, account_id, purpose, amount, created_at) values (?, ?, ?, ?, ?);`
		stmt, err = tx.Prepare(query)
		if err != nil {
			stmt.Close()
			return fmt.Errorf("createTransaction: transaction=%q account=%q prepare: error=%v rollback=%v", t.ID, t.Lines[i].AccountID, err, tx.Rollback())
		}
		if _, err := stmt.Exec(t.ID, t.Lines[i].AccountID, t.Lines[i].Purpose, t.Lines[i].Amount, time.Now()); err != nil {
			stmt.Close()
			return fmt.Errorf("createTransaction: transaction=%q account=%q insert: error=%v rollback=%v", t.ID, t.Lines[i].AccountID, err, tx.Rollback())
		}
		stmt.Close()

		// Check account balance, and if we're negative by less than t.Lines[i].Amount then we need to rollback as that account
		// didn't have sufficient funds to post the transaction.
		//
		// From Wade: Allowing overdrafts is similar to offering credit to customers, which requires additional disclosures and would need
		// to be done on an account-by-account basis.
		if opts.InitialDeposit {
			if t.Lines[0].Purpose != ACHCredit {
				return fmt.Errorf("createTransaction: InitialDeposit must be ACHCredit rollback=%v", tx.Rollback())
			}
			if len(t.Lines) == 1 && t.Lines[0].Amount > 100 {
				// Ignore all other checks and just allow the deposit
				continue
			}
		}
		// TODO(adam): I think we need to add a check (to bypass further validation) on external accounts
		// since we won't have an accurate way to confirm their balance.
		balance, err := r.getAccountBalance(tx, t.Lines[i].AccountID)
		if err != nil {
			return fmt.Errorf("createTransaction: getAccountBalance: transaction=%q account=%q: err=%v rollback=%v", t.ID, t.Lines[i].AccountID, err, tx.Rollback())
		}
		// The current account balance is negative, so if that balance is less negative than the transaction amount that means the
		// account was overdrawn (i.e. insufficient funds). If the balances are equal then we also ran out of funds.
		//
		// If the debited account is external then allow the transfer. (That accounts system will send back a returned file on an insufficient balance.)
		if opts.AllowOverdraft || !isInternalDebit(accounts, t.Lines, defaultRoutingNumber) {
			continue
		}
		if balance <= 0 || (balance <= int32(t.Lines[i].Amount) && t.Lines[i].Purpose == ACHDebit) {
			return fmt.Errorf("acocunt=%q has insufficient funds: rollback=%v", t.Lines[i].AccountID, tx.Rollback())
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("createTransaction: commit: %v", err)
	}
	return nil
}

func (r *sqlTransactionRepository) getAccountTransactions(accountID string) ([]transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("getAccountTransactions: %v", err)
	}

	query := `select distinct transaction_id from transaction_lines where account_id = ? order by created_at desc;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getAccountTransactions: prepare: error=%v rollback=%v", err, tx.Rollback())
	}
	defer stmt.Close()

	rows, err := stmt.Query(accountID)
	if err != nil {
		return nil, fmt.Errorf("getAccountTransactions: query: error=%v rollback=%v", err, tx.Rollback())
	}
	defer rows.Close()

	var transactionIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("getAccountTransactions: scan: error=%v rollback=%v", err, tx.Rollback())
		}
		transactionIDs = append(transactionIDs, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getAccountTransactions: err: error=%v rollback=%v", err, tx.Rollback())
	}

	var transactions []transaction
	for i := range transactionIDs {
		t, err := r.loadTransaction(tx, transactionIDs[i])
		if err != nil {
			return nil, fmt.Errorf("getAccountTransactions: looping: error=%v rollback=%v", err, tx.Rollback())
		}
		transactions = append(transactions, *t)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("getAccountTransactions: commit: error=%v rollback=%v", err, tx.Rollback())
	}
	return transactions, nil
}

func (r *sqlTransactionRepository) getTransaction(transactionID string) (*transaction, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("getTransaction: %v", err)
	}
	transaction, err := r.loadTransaction(tx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("getTransaction: error=%v rollback=%v", err, tx.Rollback())
	}
	return transaction, tx.Commit()
}

func (r *sqlTransactionRepository) loadTransaction(tx *sql.Tx, transactionID string) (*transaction, error) {
	query := `select timestamp from transactions where transaction_id = ? and deleted_at is null limit 1;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("loadTransaction: timestamp: %v", err)
	}
	var timestamp time.Time
	if err := stmt.QueryRow(transactionID).Scan(&timestamp); err != nil {
		stmt.Close()
		return nil, fmt.Errorf("loadTransaction: timestamp query: %v", err)
	}
	stmt.Close() // close to prevent leaks

	query = `select account_id, purpose, amount from transaction_lines where transaction_id = ? and deleted_at is null`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("loadTransaction: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(transactionID)
	if err != nil {
		return nil, fmt.Errorf("loadTransaction: query: %v", err)
	}
	defer rows.Close()

	var lines []transactionLine
	for rows.Next() {
		var line transactionLine
		if err := rows.Scan(&line.AccountID, &line.Purpose, &line.Amount); err != nil {
			return nil, fmt.Errorf("loadTransaction: scan transaction=%q account=%q: %v", transactionID, line.AccountID, err)
		}
		lines = append(lines, line)
	}
	return &transaction{
		ID:        transactionID,
		Timestamp: timestamp,
		Lines:     lines,
	}, rows.Err()
}

func (r *sqlTransactionRepository) getAccountBalance(tx *sql.Tx, accountID string) (int32, error) {
	if accountID == "" {
		return 0, nil
	}

	// TODO(adam): At some point we should probably checkpoint balances so we avoid an entire index scan on an account_id
	query := `select amount, purpose from transaction_lines where account_id = ? and deleted_at is null;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(accountID)
	if err != nil {
		return 0, fmt.Errorf("problem querying account=%s balance: %v", accountID, err)
	}
	defer rows.Close()

	var amount int32
	for rows.Next() {
		var amt int32
		var purpose string
		if err := rows.Scan(&amt, &purpose); err != nil {
			return 0, fmt.Errorf("error reading amount/purpose: %v", err)
		}
		if strings.EqualFold(purpose, "achdebit") {
			amount -= amt
		} else {
			amount += amt
		}
	}

	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("problem getting account=%s balance: %v", accountID, err)
	}

	return amount, nil
}

// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/moov-io/gl"

	"github.com/go-kit/kit/log"
)

type sqliteTransactionRepository struct {
	db     *sql.DB
	logger log.Logger

	accountRepo accountRepository
}

func setupSqliteTransactionStorage(logger log.Logger, path string) (*sqliteTransactionRepository, error) {
	db, err := createSqliteConnection(logger, path)
	if err != nil {
		return nil, err
	}
	// Break the cyclic dependency between account and transaction repositories
	repo := &sqliteTransactionRepository{db: db, logger: logger}
	accountRepo := &sqliteAccountRepository{db, logger, repo}
	repo.accountRepo = accountRepo
	return repo, nil
}

func (r *sqliteTransactionRepository) Ping() error {
	return r.db.Ping()
}

func (r *sqliteTransactionRepository) Close() error {
	return r.db.Close()
}

// isInternalDebit returns true only when the debited account's routing number matches
// GL's configured routing number. This means we have to be accountable for choosing
// to allow an overdraft or not.
func isInternalDebit(accounts []*gl.Account, lines []transactionLine, glRoutingNumber string) bool {
	for i := range accounts {
		for j := range lines {
			if accounts[i].ID == lines[j].AccountId {
				switch lines[j].Purpose {
				case ACHDebit:
					return accounts[i].RoutingNumber == glRoutingNumber
				}
			}
		}
	}
	return true // default to assuming we need to check/prevent an overdraft
}

func (r *sqliteTransactionRepository) createTransaction(t transaction, opts createTransactionOpts) error {
	if err := t.validate(); err != nil && !opts.InitialDeposit {
		return fmt.Errorf("transaction=%q is invalid: %v", t.ID, err)
	}

	accounts, err := r.accountRepo.GetAccounts(grabAccountIds(t.Lines))
	if err != nil {
		return fmt.Errorf("createTransaction: problem reading accounts for transaction=%q: %v", t.ID, err)
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("createTransaction: tx.Begin: %v", err)
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
			return fmt.Errorf("createTransaction: transaction=%q account=%q prepare: error=%v rollback=%v", t.ID, t.Lines[i].AccountId, err, tx.Rollback())
		}
		if _, err := stmt.Exec(t.ID, t.Lines[i].AccountId, t.Lines[i].Purpose, t.Lines[i].Amount, time.Now()); err != nil {
			stmt.Close()
			return fmt.Errorf("createTransaction: transaction=%q account=%q insert: error=%v rollback=%v", t.ID, t.Lines[i].AccountId, err, tx.Rollback())
		}
		stmt.Close()

		// Check account balance, and if we're negative by less than t.Lines[i].Amount then we need to rollback as that account
		// didn't have sufficient funds to post the transaction.
		//
		// From Wade: Allowing overdrafts is similar to offering credit to customers, which requires additional disclosures and would need
		// to be done on an account-by-account basis.
		if opts.InitialDeposit && len(t.Lines) == 1 && t.Lines[0].Amount > 100 {
			// Ignore all other checks and just allow the deposit
			continue
		}
		balance, err := r.getAccountBalance(tx, t.Lines[i].AccountId)
		if err != nil {
			return fmt.Errorf("createTransaction: getAccountBalance: transaction=%q account=%q: err=%v rollback=%v", t.ID, t.Lines[i].AccountId, err, tx.Rollback())
		}
		// The current account balance is negative, so if that balance is less negative than the transaction amount that means the
		// account was overdrawn (i.e. insufficient funds). If the balances are equal then we also ran out of funds.
		//
		// If the debited account is external then allow the transfer. (That GL system will send back a returned file on an insufficient balance.)
		if opts.AllowOverdraft || !isInternalDebit(accounts, t.Lines, defaultRoutingNumber) {
			continue
		}
		if balance <= 0 || balance <= int64(t.Lines[i].Amount) {
			return fmt.Errorf("acocunt=%q has insufficient funds: rollback=%v", t.Lines[i].AccountId, tx.Rollback())
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("createTransaction: commit: %v", err)
	}
	return nil
}

func (r *sqliteTransactionRepository) getAccountTransactions(accountId string) ([]transaction, error) {
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

	rows, err := stmt.Query(accountId)
	if err != nil {
		return nil, fmt.Errorf("getAccountTransactions: query: error=%v rollback=%v", err, tx.Rollback())
	}
	defer rows.Close()

	var transactionIds []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("getAccountTransactions: scan: error=%v rollback=%v", err, tx.Rollback())
		}
		transactionIds = append(transactionIds, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("getAccountTransactions: err: error=%v rollback=%v", err, tx.Rollback())
	}

	var transactions []transaction
	for i := range transactionIds {
		t, err := r.getTransaction(tx, transactionIds[i])
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

func (r *sqliteTransactionRepository) getTransaction(tx *sql.Tx, transactionId string) (*transaction, error) {
	query := `select timestamp from transactions where transaction_id = ? and deleted_at is null limit 1;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getTransaction: timestamp: %v", err)
	}
	var timestamp time.Time
	if err := stmt.QueryRow(transactionId).Scan(&timestamp); err != nil {
		stmt.Close()
		return nil, fmt.Errorf("getTransaction: timestamp query: %v", err)
	}
	stmt.Close() // close to prevent leaks

	query = `select account_id, purpose, amount from transaction_lines where transaction_id = ? and deleted_at is null`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getTransaction: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(transactionId)
	if err != nil {
		return nil, fmt.Errorf("getTransaction: query: %v", err)
	}
	defer rows.Close()

	var lines []transactionLine
	for rows.Next() {
		var line transactionLine
		if err := rows.Scan(&line.AccountId, &line.Purpose, &line.Amount); err != nil {
			return nil, fmt.Errorf("getTransaction: scan transaction=%q account=%q: %v", transactionId, line.AccountId, err)
		}
		lines = append(lines, line)
	}
	return &transaction{
		ID:        transactionId,
		Timestamp: timestamp,
		Lines:     lines,
	}, rows.Err()
}

func (r *sqliteTransactionRepository) getAccountBalance(tx *sql.Tx, accountId string) (int64, error) {
	// TODO(adam): At some point we should probably checkpoint balances so we avoid an entire index scan on an account_id
	query := `select coalesce(sum(amount), 0) from transaction_lines where account_id = ? and deleted_at is null;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	var amount int64
	if err := stmt.QueryRow(accountId).Scan(&amount); err != nil {
		return 0, err
	}
	return amount, nil
}

// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/moov-io/base"
	moovhttp "github.com/moov-io/base/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	errNoAccountId     = errors.New("no accountId found")
	errNoTransactionId = errors.New("no transactionId found")
)

type TransactionPurpose string

var (
	ACHCredit TransactionPurpose = "achcredit"
	ACHDebit  TransactionPurpose = "achdebit"
	Fee       TransactionPurpose = "fee"
	Interest  TransactionPurpose = "interest"
	Transfer  TransactionPurpose = "transfer"
	Wire      TransactionPurpose = "wire"
)

func (p *TransactionPurpose) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*p = TransactionPurpose(strings.ToLower(s))
	if err := p.validate(); err != nil {
		return err
	}
	return nil
}

func (p TransactionPurpose) validate() error {
	switch p {
	case ACHCredit, ACHDebit, Fee, Interest, Transfer, Wire:
		return nil
	default:
		return fmt.Errorf("unknown TransactionPurpose %q", p)
	}
}

type transactionLine struct {
	AccountId string             `json:"accountId"`
	Purpose   TransactionPurpose `json:"purpose"`
	Amount    int                `json:"amount"`
}

func (line transactionLine) validate() error {
	if line.AccountId == "" || line.Amount == 0 {
		return fmt.Errorf("transactionLine: AccountId=%s Amount=%d is invalid", line.AccountId, line.Amount)
	}
	return line.Purpose.validate()
}

type createTransactionRequest struct {
	Lines []transactionLine `json:"lines"`
}

func (r *createTransactionRequest) asTransaction(id string) transaction {
	return transaction{
		ID:        id,
		Lines:     r.Lines,
		Timestamp: time.Now(),
	}
}

type transaction struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Lines     []transactionLine `json:"lines"`
}

func (t transaction) validate() error {
	if t.ID == "" {
		return errors.New("transaction: empty ID")
	}
	if len(t.Lines) == 0 {
		return fmt.Errorf("transaction=%s has no Lines", t.ID)
	}
	if t.Timestamp.IsZero() {
		return fmt.Errorf("transaction=%s has no Timestamp", t.ID)
	}

	sum := 0
	for i := range t.Lines {
		if t.Lines[i].Purpose == ACHDebit {
			sum += -1 * t.Lines[i].Amount
		} else {
			sum += t.Lines[i].Amount
		}
		if err := t.Lines[i].validate(); err != nil {
			return fmt.Errorf("transaction=%s has invalid line[%d]: %v", t.ID, i, err)
		}
	}
	if sum == 0 {
		return nil
	}
	return fmt.Errorf("transaction=%s has %d invalid lines sum=%d", t.ID, len(t.Lines), sum)
}

func addTransactionRoutes(logger log.Logger, router *mux.Router, accountRepo accountRepository, transactionRepo transactionRepository) {
	router.Methods("GET").Path("/accounts/{accountId}/transactions").HandlerFunc(getAccountTransactions(logger, transactionRepo))
	router.Methods("POST").Path("/accounts/transactions").HandlerFunc(createTransaction(logger, accountRepo, transactionRepo))
	router.Methods("POST").Path("/accounts/transactions/{transactionId}/reversal").HandlerFunc(createTransactionReversal(logger, accountRepo, transactionRepo))
}

func getAccountId(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["accountId"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoAccountId)
		return ""
	}
	return v
}

func getAccountTransactions(logger log.Logger, transactionRepo transactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		accountId := getAccountId(w, r)
		if accountId == "" {
			moovhttp.Problem(w, errNoAccountId)
			return
		}

		transactions, err := transactionRepo.getAccountTransactions(accountId)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(transactions)

	}
}

func createTransaction(logger log.Logger, accountRepo accountRepository, transactionRepo transactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		requestId := moovhttp.GetRequestId(r)

		var req createTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// Post the transaction
		tx := req.asTransaction(base.ID())
		if err := transactionRepo.createTransaction(tx, createTransactionOpts{AllowOverdraft: false}); err != nil {
			logger.Log("transactions", fmt.Errorf("problem creating transaction: %v", err), "requestId", requestId)
			moovhttp.Problem(w, err)
			return
		}
		logger.Log("transaction", fmt.Errorf("created transaction %s", tx.ID), "requestId", requestId)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tx)
	}
}

func getTransactionId(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["transactionId"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoTransactionId)
		return ""
	}
	return v
}

func createTransactionReversal(logger log.Logger, accountRepo accountRepository, transactionRepo transactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		// Read our transactionId and do an info log
		requestId, transactionId := moovhttp.GetRequestId(r), getTransactionId(w, r)
		if transactionId == "" {
			return
		}
		logger.Log("transaction", fmt.Sprintf("reversing transaction %s", transactionId), "requestId", requestId)

		// reverse the transaction (after reading it from our database)
		transaction, err := transactionRepo.getTransaction(transactionId)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		transaction.ID = base.ID()
		transaction.Timestamp = time.Now()
		for i := range transaction.Lines {
			// Swap Purpose back if Debit vs Credit
			switch {
			case transaction.Lines[i].Purpose == ACHCredit:
				transaction.Lines[i].Purpose = ACHDebit
			case transaction.Lines[i].Purpose == ACHDebit:
				transaction.Lines[i].Purpose = ACHCredit
			}
			// Invert the amount posted to each account
			transaction.Lines[i].Amount = -1 * transaction.Lines[i].Amount
		}
		if err := transactionRepo.createTransaction(*transaction, createTransactionOpts{AllowOverdraft: false}); err != nil {
			logger.Log("transactions", fmt.Errorf("problem creating transaction: %v", err), "requestId", requestId)
			moovhttp.Problem(w, err)
			return
		}
		logger.Log("transactions", fmt.Sprintf("reversed (original transaction=%s) transaction=%s", transactionId, transaction.ID), "requestId", requestId)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(transaction)
	}
}

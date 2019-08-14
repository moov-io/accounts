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
	errNoAccountID     = errors.New("no accountID found")
	errNoTransactionID = errors.New("no transactionID found")
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
	AccountID string             `json:"accountId"`
	Purpose   TransactionPurpose `json:"purpose"`
	Amount    int                `json:"amount"`
}

func (line transactionLine) validate() error {
	if line.AccountID == "" || line.Amount == 0 {
		return fmt.Errorf("transactionLine: AccountID=%s Amount=%d is invalid", line.AccountID, line.Amount)
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
		if t.Lines[0].Amount < 0 {
			return fmt.Errorf("transaction=%s has negative amount=%d", t.ID, t.Lines[0].Amount)
		}
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
	router.Methods("POST").Path("/accounts/transactions/{transactionID}/reversal").HandlerFunc(createTransactionReversal(logger, accountRepo, transactionRepo))
}

func getAccountID(w http.ResponseWriter, r *http.Request) string {
	v, _ := mux.Vars(r)["accountId"]
	if v == "" {
		if v, _ = mux.Vars(r)["accountID"]; v == "" {
			moovhttp.Problem(w, errNoAccountID)
			return ""
		}
	}
	return v
}

func getAccountTransactions(logger log.Logger, transactionRepo transactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		accountID := getAccountID(w, r)
		if accountID == "" {
			moovhttp.Problem(w, errNoAccountID)
			return
		}

		transactions, err := transactionRepo.getAccountTransactions(accountID)
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

		requestID := moovhttp.GetRequestID(r)

		var req createTransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// Post the transaction
		tx := req.asTransaction(base.ID())
		if err := transactionRepo.createTransaction(tx, createTransactionOpts{AllowOverdraft: false}); err != nil {
			logger.Log("transactions", fmt.Errorf("problem creating transaction: %v", err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		logger.Log("transaction", fmt.Errorf("created transaction %s", tx.ID), "requestID", requestID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(tx)
	}
}

func getTransactionID(w http.ResponseWriter, r *http.Request) string {
	v, _ := mux.Vars(r)["transactionId"]
	if v == "" {
		if v, _ = mux.Vars(r)["transactionID"]; v == "" {
			moovhttp.Problem(w, errNoTransactionID)
			return ""
		}
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

		// Read our transactionID and do an info log
		requestID, transactionID := moovhttp.GetRequestID(r), getTransactionID(w, r)
		if transactionID == "" {
			return
		}
		logger.Log("transaction", fmt.Sprintf("reversing transaction %s", transactionID), "requestID", requestID)

		// reverse the transaction (after reading it from our database)
		transaction, err := transactionRepo.getTransaction(transactionID)
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
		}
		if err := transactionRepo.createTransaction(*transaction, createTransactionOpts{AllowOverdraft: false}); err != nil {
			logger.Log("transactions", fmt.Errorf("problem creating transaction: %v", err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		logger.Log("transactions", fmt.Sprintf("reversed (original transaction=%s) transaction=%s", transactionID, transaction.ID), "requestID", requestID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(transaction)
	}
}

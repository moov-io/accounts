// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	accounts "github.com/moov-io/accounts/client"
	"github.com/moov-io/base"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type mockTransactionRepository struct {
	err error

	transactions []transaction
	created      transaction
}

func (r *mockTransactionRepository) Ping() error {
	return r.err
}

func (r *mockTransactionRepository) Close() error {
	return r.err
}

func (r *mockTransactionRepository) createTransaction(tx transaction, opts createTransactionOpts) error {
	if err := tx.validate(); err != nil && !opts.InitialDeposit {
		return err
	}
	r.created = tx
	return r.err
}

func (r *mockTransactionRepository) getAccountTransactions(accountID string) ([]transaction, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.transactions, nil
}

func (r *mockTransactionRepository) getTransaction(transactionId string) (*transaction, error) {
	if r.err != nil {
		return nil, r.err
	}
	return &r.transactions[0], nil
}

func TestTransactionPurpose(t *testing.T) {
	if err := TransactionPurpose("").validate(); err == nil {
		t.Error("expected error")
	}
	if err := TransactionPurpose("other").validate(); err == nil {
		t.Error("expected error")
	}

	// valid cases
	cases := []string{"achcredit", "achdebit", "fee", "interest", "transfer", "wire"}
	for i := range cases {
		if err := TransactionPurpose(cases[i]).validate(); err != nil {
			t.Errorf("expected no error on %q: %v", cases[i], err)
		}
	}

	// JSON
	var purpose TransactionPurpose
	if err := json.Unmarshal([]byte(`"other"`), &purpose); err != nil {
		if err.Error() != "unknown TransactionPurpose \"other\"" {
			t.Fatal(err)
		}
	}
	if err := purpose.validate(); err == nil {
		t.Errorf("expected error")
	}
	// valid, case-insensitive
	if err := json.Unmarshal([]byte(`"achCredit"`), &purpose); err != nil {
		t.Fatal(err)
	}
	if err := purpose.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if purpose != ACHCredit {
		t.Errorf("unexpected value: %s", purpose)
	}
}

func TestTransaction__validate(t *testing.T) {
	tx := transaction{
		ID:        base.ID(),
		Timestamp: time.Now(),
		Lines: []transactionLine{
			{
				AccountId: base.ID(),
				Purpose:   ACHDebit,
				Amount:    -500,
			},
			{
				AccountId: base.ID(),
				Purpose:   ACHCredit,
				Amount:    500,
			},
		},
	}
	if err := tx.validate(); err != nil {
		t.Error(err)
	}

	// make invalid
	tx.ID = ""
	if err := tx.validate(); err == nil {
		t.Error("expected error")
	}
	tx.ID = base.ID()

	var empty time.Time
	tx.Timestamp = empty
	if err := tx.validate(); err == nil {
		t.Error("expected error")
	}
	tx.Timestamp = time.Now()

	tx.Lines[0].Amount = 1
	if err := tx.validate(); err == nil {
		t.Error("expected error")
	}
	tx.Lines[0].Amount = -500

	tx.Lines[0].Purpose = TransactionPurpose("other")
	if err := tx.validate(); err == nil {
		t.Error("expected error")
	}

	tx.Lines = []transactionLine{}
	if err := tx.validate(); err == nil {
		t.Error("expected error")
	}

}

func TestTransactions_getAccountId(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/foo", nil)

	if accountId := getAccountId(w, req); accountId != "" {
		t.Errorf("expected no accountId, got %q", accountId)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}

	w = httptest.NewRecorder()

	// successful extraction
	var accountId string
	router := mux.NewRouter()
	router.Methods("GET").Path("/accounts/{accountId}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		accountId = getAccountId(w, req)
	})
	router.ServeHTTP(w, httptest.NewRequest("GET", "/accounts/bar", nil))
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}
	if accountId != "bar" {
		t.Errorf("got %q", accountId)
	}
}

func TestTransactions_Get(t *testing.T) {
	accountId := base.ID()
	accountRepo := &testAccountRepository{}
	transactionRepo := &mockTransactionRepository{
		transactions: []transaction{
			{
				ID:        base.ID(),
				Timestamp: time.Now().Add(-24 * time.Hour),
				Lines: []transactionLine{
					{
						AccountId: accountId,
						Purpose:   Transfer,
						Amount:    13412,
					},
				},
			},
			{
				ID:        base.ID(),
				Timestamp: time.Now().Add(-24 * 2 * time.Hour),
				Lines: []transactionLine{
					{
						AccountId: accountId,
						Purpose:   Transfer,
						Amount:    5331,
					},
				},
			},
		},
	}

	router := mux.NewRouter()
	addTransactionRoutes(log.NewNopLogger(), router, accountRepo, transactionRepo)

	req := httptest.NewRequest("GET", fmt.Sprintf("/accounts/%s/transactions", accountId), nil)
	req.Header.Set("x-user-id", base.ID())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}
	var resp []transaction
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp) != 2 {
		t.Errorf("got %d transactions: %#v", len(resp), resp)
	}

	// set an error and make sure we respond as such
	transactionRepo.err = errors.New("bad thing")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
}

func TestTransactions_Create(t *testing.T) {
	accountRepo := &testAccountRepository{
		accounts: []*accounts.Account{
			{Id: base.ID(), Balance: 10000},
			{Id: base.ID(), Balance: 1000},
		},
	}
	transactionRepo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addTransactionRoutes(log.NewNopLogger(), router, accountRepo, transactionRepo)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(createTransactionRequest{
		Lines: []transactionLine{
			{AccountId: accountRepo.accounts[0].Id, Purpose: ACHDebit, Amount: -4121},
			{AccountId: accountRepo.accounts[1].Id, Purpose: ACHCredit, Amount: 4121},
		},
	})
	req := httptest.NewRequest("POST", "/accounts/transactions", &body)
	req.Header.Set("x-user-id", base.ID())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}

	// set an error
	accountRepo.err = errors.New("bad thing")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}

	// set an error and make sure we respond as such
	accountRepo.err = nil // wipe error
	transactionRepo.err = errors.New("bad thing")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
}

func TestTransactions_CreateInvalid(t *testing.T) {
	accountRepo := &testAccountRepository{}
	transactionRepo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addTransactionRoutes(log.NewNopLogger(), router, accountRepo, transactionRepo)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(createTransactionRequest{
		Lines: []transactionLine{
			// Invalid Lines will force an error
			{AccountId: base.ID(), Purpose: ACHDebit, Amount: -4121},
			{AccountId: base.ID(), Purpose: ACHCredit, Amount: -121},
		},
	})
	req := httptest.NewRequest("POST", "/accounts/transactions", &body)
	req.Header.Set("x-user-id", base.ID())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
}

func TestTransactions__createTransactionReversal(t *testing.T) {
	accountRepo := &testAccountRepository{}
	transactionRepo := &mockTransactionRepository{
		transactions: []transaction{
			{
				ID:        base.ID(),
				Timestamp: time.Now(),
				Lines: []transactionLine{
					{
						AccountId: base.ID(),
						Purpose:   ACHDebit,
						Amount:    -1000,
					},
					{
						AccountId: base.ID(),
						Purpose:   ACHCredit,
						Amount:    1000,
					},
				},
			},
		},
	}

	router := mux.NewRouter()
	addTransactionRoutes(log.NewNopLogger(), router, accountRepo, transactionRepo)

	req := httptest.NewRequest("POST", fmt.Sprintf("/accounts/transactions/%s/reversal", transactionRepo.transactions[0].ID), nil)
	req.Header.Set("x-user-id", base.ID())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
	// Check that our reversed transaction is valid
	if err := transactionRepo.created.validate(); err != nil {
		t.Fatal(err)
	}

	// verify our response was a transaction
	var tx transaction
	if err := json.NewDecoder(w.Body).Decode(&tx); err != nil {
		t.Fatal(err)
	}
	if tx.ID != transactionRepo.created.ID {
		t.Errorf("transactions don't match")
	}

	// set an error and ensure we fail
	transactionRepo.err = errors.New("bad thing")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestTransactions_getTransactionId(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/foo", nil)

	if accountId := getTransactionId(w, req); accountId != "" {
		t.Errorf("expected no accountId, got %q", accountId)
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}

	w = httptest.NewRecorder()

	// successful extraction
	var transactionId string
	router := mux.NewRouter()
	router.Methods("GET").Path("/transactions/{transactionId}").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		transactionId = getTransactionId(w, req)
	})
	router.ServeHTTP(w, httptest.NewRequest("GET", "/transactions/bar", nil))
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}
	if transactionId != "bar" {
		t.Errorf("got %q", transactionId)
	}
}

// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/moov-io/base"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type mockTransactionRepository struct {
	err error

	transactions []transaction
}

func (r *mockTransactionRepository) Ping() error {
	return r.err
}

func (r *mockTransactionRepository) createTransaction(tx transaction) error {
	return r.err
}

func (r *mockTransactionRepository) getAccountTransactions(accountID string) ([]transaction, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.transactions, nil
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

func TestTransactions_Create(t *testing.T) {
	repo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addTransactionRoutes(log.NewNopLogger(), router, repo)

	var body bytes.Buffer
	json.NewEncoder(&body).Encode(createTransactionRequest{
		Lines: []transactionLine{
			{AccountId: base.ID(), Purpose: ACHDebit, Amount: -4121},
			{AccountId: base.ID(), Purpose: ACHCredit, Amount: -121},
		},
	})
	req := httptest.NewRequest("POST", "/accounts/foo/transactions", &body)
	req.Header.Set("x-user-id", base.ID())

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("got %d", w.Code)
	}

	// set an error and make sure we respond as such
	repo.err = errors.New("bad thing")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d", w.Code)
	}
}

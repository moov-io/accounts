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

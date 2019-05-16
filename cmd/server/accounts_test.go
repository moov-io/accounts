// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	accounts "github.com/moov-io/accounts/client"
	"github.com/moov-io/base"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	mockAccountRepo = &testAccountRepository{
		accounts: []*accounts.Account{
			{
				Id:            base.ID(),
				CustomerId:    base.ID(),
				Name:          "example account",
				AccountNumber: "132",
				RoutingNumber: "51321",
				Balance:       21415,
			},
		},
	}
)

func init() {
	defaultRoutingNumber = "231380104"
}

func TestAccounts__createAccountRequest(t *testing.T) {
	req := createAccountRequest{100, "example acct", "checking"} // $1
	if err := req.validate(); err != nil {
		t.Error(err)
	}

	req.Balance = 10 // $0.10
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}
	req.Balance = 1000

	req.Type = "SavInGs" // valid
	if err := req.validate(); err != nil {
		t.Error(err)
	}

	req.Type = "other" // invalid
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}
}

func TestAccounts__CreateAccount(t *testing.T) {
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"balance": 1000, "name": "Money", "type": "Savings"}`)
	req := httptest.NewRequest("POST", "/customers/foo/accounts", body)
	req.Header.Set("x-user-id", "test")

	transactionRepo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addAccountRoutes(log.NewNopLogger(), router, mockAccountRepo, transactionRepo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d:\n  %s", w.Code, w.Body.String())
	}

	var acct accounts.Account // TODO(adam): check more of Customer response?
	if err := json.NewDecoder(w.Body).Decode(&acct); err != nil {
		t.Fatal(err)
	}
	if acct.Id == "" {
		t.Error("empty Account.Id")
	}
}

func TestAccounts__GetCustomerAccounts(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/accounts", nil)
	req.Header.Set("x-user-id", "test")

	transactionRepo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addAccountRoutes(log.NewNopLogger(), router, mockAccountRepo, transactionRepo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var accounts []accounts.Account // TODO(adam): check more of Customer response?
	if err := json.NewDecoder(w.Body).Decode(&accounts); err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Errorf("expected 1 account, but got %d", len(accounts))
	}
	if accounts[0].Id == "" {
		t.Error("empty Account.Id")
	}
}

func TestAccounts__SearchAccounts(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/accounts/search?number=1&routingNumber=123&type=Savings", nil)
	req.Header.Set("x-user-id", "test")

	transactionRepo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addAccountRoutes(log.NewNopLogger(), router, mockAccountRepo, transactionRepo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var acct accounts.Account // TODO(adam): check more of Customer response?
	if err := json.NewDecoder(w.Body).Decode(&acct); err != nil {
		t.Fatal(err)
	}
	if acct.Id == "" {
		t.Error("empty Account.Id")
	}
}

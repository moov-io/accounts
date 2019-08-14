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
				ID:            base.ID(),
				CustomerID:    base.ID(),
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
	req := createAccountRequest{"customerID", 100, "example acct", "", "checking"} // $1
	if err := req.validate(); err != nil {
		t.Error(err)
	}

	req.CustomerID = ""
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}
	req.CustomerID = "customerID"

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
	body := strings.NewReader(`{"customerID": "customerID", "balance": 1000, "name": "Money", "type": "Savings"}`)
	req := httptest.NewRequest("POST", "/accounts", body)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "request")

	accountRepo := &testAccountRepository{}
	transactionRepo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addAccountRoutes(log.NewNopLogger(), router, accountRepo, transactionRepo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d:\n  %s", w.Code, w.Body.String())
	}

	var acct accounts.Account
	if err := json.NewDecoder(w.Body).Decode(&acct); err != nil {
		t.Fatal(err)
	}
	if acct.ID == "" {
		t.Error("empty Account.ID")
	}
}

func TestAccounts__GetCustomerAccounts(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/accounts/search?customerId=customerID", nil)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "request")

	transactionRepo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addAccountRoutes(log.NewNopLogger(), router, mockAccountRepo, transactionRepo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var accounts []accounts.Account
	if err := json.NewDecoder(w.Body).Decode(&accounts); err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Errorf("expected 1 account, but got %d", len(accounts))
	}
	if accounts[0].ID == "" {
		t.Error("empty Account.ID")
	}
}

func TestAccounts__SearchAccounts(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/accounts/search?number=1&routingNumber=123&type=Savings", nil)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "request")

	transactionRepo := &mockTransactionRepository{}

	router := mux.NewRouter()
	addAccountRoutes(log.NewNopLogger(), router, mockAccountRepo, transactionRepo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var accts []*accounts.Account
	if err := json.NewDecoder(w.Body).Decode(&accts); err != nil {
		t.Fatal(err)
	}
	if len(accts) == 0 || accts[0].ID == "" {
		t.Errorf("length:%d empty Account.ID", len(accts))
	}
}

func TestAccounts__generateAccountNumber(t *testing.T) {
	repo := &testAccountRepository{}
	id, err := generateAccountNumber(&accounts.Account{}, repo)
	if id == "" || err != nil {
		t.Fatalf("empty account number: %v", err)
	}

	repo.accounts = append(repo.accounts, &accounts.Account{
		ID:            "accountID",
		AccountNumber: "123",
		RoutingNumber: "987654320",
	})

	id, err = generateAccountNumber(&accounts.Account{AccountNumber: "123"}, repo)
	if id != "" || err == nil {
		t.Fatalf("expected empty account number id=%v error=%v", id, err)
	}
}

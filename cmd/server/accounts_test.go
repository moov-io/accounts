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

	"github.com/moov-io/gl"

	"github.com/gorilla/mux"
)

func TestAccounts__CreateAccount(t *testing.T) {
	w := httptest.NewRecorder()
	body := strings.NewReader(`{"customerId": "foo", "name": "Money", "type": "Savings"}`)
	req := httptest.NewRequest("POST", "/customers/foo/accounts", body)
	req.Header.Set("x-user-id", "test")

	router := mux.NewRouter()
	addAccountRoutes(nil, router)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var acct gl.Account // TODO(adam): check more of Customer response?
	if err := json.NewDecoder(w.Body).Decode(&acct); err != nil {
		t.Fatal(err)
	}
	if acct.ID == "" {
		t.Error("empty Account.ID")
	}
}

func TestAccounts__GetCustomerAccounts(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/accounts", nil)
	req.Header.Set("x-user-id", "test")

	router := mux.NewRouter()
	addAccountRoutes(nil, router)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var accounts []gl.Account // TODO(adam): check more of Customer response?
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

	router := mux.NewRouter()
	addAccountRoutes(nil, router)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var acct gl.Account // TODO(adam): check more of Customer response?
	if err := json.NewDecoder(w.Body).Decode(&acct); err != nil {
		t.Fatal(err)
	}
	if acct.ID == "" {
		t.Error("empty Account.ID")
	}
}

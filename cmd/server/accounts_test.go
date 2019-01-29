// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestAccounts__CreateAccount(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/accounts", nil)
	req.Header.Set("x-user-id", "test")

	router := mux.NewRouter()
	addAccountRoutes(nil, router)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
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
}

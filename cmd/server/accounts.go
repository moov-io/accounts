// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"net/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func addAccountRoutes(logger log.Logger, r *mux.Router) {
	r.Methods("GET").Path("/customers/{customerId}/accounts").HandlerFunc(getCustomerAccounts(logger))
	r.Methods("POST").Path("/customers/{customerId}/accounts").HandlerFunc(createCustomerAccount(logger))
}

func createCustomerAccount(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func getCustomerAccounts(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

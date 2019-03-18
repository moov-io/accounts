// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/moov-io/base"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/gl"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func addAccountRoutes(logger log.Logger, r *mux.Router) {
	r.Methods("GET").Path("/customers/{customerId}/accounts").HandlerFunc(getCustomerAccounts(logger))
	r.Methods("POST").Path("/customers/{customerId}/accounts").HandlerFunc(createCustomerAccount(logger))
}

type createAccountRequest struct {
	CustomerID string `json:"customerId"`
	Name       string `json:"name"`
	Type       string `json:"type"`
}

func (r createAccountRequest) validate() error {
	if r.CustomerID == "" {
		return errors.New("createAccountRequest: missing CustomerID")
	}
	if r.Name == "" {
		return errors.New("createAccountRequest: missing Name")
	}
	if r.Type == "" {
		return errors.New("createAccountRequest: missing Type")
	}
	return nil
}

func createCustomerAccount(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		var req createAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&gl.Account{
			ID:                  base.ID(),
			CustomerID:          req.CustomerID,
			Name:                req.Name,
			AccountNumber:       "123",
			AccountNumberMasked: "",          // TODO(adam): show both in mocks?
			RoutingNumber:       "121042882", // Wells Fargo for CA
			Status:              "open",
			Type:                "checking",
			CreatedAt:           time.Now(),
			// ClosedAt: time.Time{},
			LastModified:     time.Now(),
			Balance:          0,
			BalanceAvailable: 0,
			BalancePending:   0,
		})
	}
}

func getCustomerAccounts(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]*gl.Account{
			{
				ID:                  base.ID(),
				CustomerID:          base.ID(),
				Name:                "Sample Account",
				AccountNumber:       "132",
				AccountNumberMasked: "",          // TODO(adam): show both in mocks?
				RoutingNumber:       "121042882", // Wells Fargo for CA
				Status:              "open",
				Type:                "checking",
				CreatedAt:           time.Now(),
				// ClosedAt: time.Time{},
				LastModified:     time.Now(),
				Balance:          0,
				BalanceAvailable: 0,
				BalancePending:   0,
			},
		})
	}
}

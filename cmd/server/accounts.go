// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/moov-io/base"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/gl"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	defaultRoutingNumber = os.Getenv("DEFAULT_ROUTING_NUMBER")
)

func addAccountRoutes(logger log.Logger, r *mux.Router, repo accountRepository) {
	r.Methods("GET").Path("/accounts/search").HandlerFunc(searchAccounts(logger, repo))

	r.Methods("GET").Path("/customers/{customerId}/accounts").HandlerFunc(getCustomerAccounts(logger, repo))
	r.Methods("POST").Path("/customers/{customerId}/accounts").HandlerFunc(createCustomerAccount(logger, repo))
}

// searchAccounts will attempt to find an Account which matches all query parameters and if all match return
// the account. Otherwise a 404 will be returned. '400 Bad Request' will be returned if query parameters are missing.
func searchAccounts(logger log.Logger, repo accountRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		q := r.URL.Query()
		reqAcctNumber, reqRoutingNumber := q.Get("number"), q.Get("routingNumber")
		reqAcctType := q.Get("type")
		if reqAcctNumber == "" || reqRoutingNumber == "" || reqAcctType == "" {
			moovhttp.Problem(w, fmt.Errorf("missing query parameters: number=%q, routingNumber=%q, type=%q", reqAcctNumber, reqRoutingNumber, reqAcctType))
			return
		}

		account, err := repo.SearchAccounts(reqAcctNumber, reqRoutingNumber, reqAcctType)
		if err != nil || account == nil {
			if requestId := moovhttp.GetRequestId(r); requestId != "" {
				logger.Log("accounts", fmt.Sprintf("%v", err), "requestId", requestId)
			}
			moovhttp.Problem(w, fmt.Errorf("account not found, err=%v", err))
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
	}
}

type createAccountRequest struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

func (r createAccountRequest) validate() error {
	if r.Name == "" {
		return errors.New("createAccountRequest: missing Name")
	}
	r.Type = strings.ToLower(r.Type)
	switch r.Type {
	case "checking", "savings":
	default:
		return fmt.Errorf("createAccountRequest: unknown Type: %q", r.Type)
	}
	return nil
}

func createAccountNumber() string {
	n, _ := rand.Int(rand.Reader, big.NewInt(1e9))
	return fmt.Sprintf("%d", n.Int64())
}

func createCustomerAccount(logger log.Logger, repo accountRepository) http.HandlerFunc {
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

		customerId, now := getCustomerId(w, r), time.Now()
		account := &gl.Account{
			ID:            base.ID(),
			CustomerID:    customerId,
			Name:          req.Name,
			AccountNumber: createAccountNumber(),
			RoutingNumber: defaultRoutingNumber,
			Status:        "open",
			Type:          req.Type,
			CreatedAt:     now,
			LastModified:  &now,
		}

		if err := repo.CreateAccount(customerId, account); err != nil {
			if requestId := moovhttp.GetRequestId(r); requestId != "" {
				logger.Log("accounts", fmt.Sprintf("%v", err), "requestId", requestId)
			}
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
	}
}

func getCustomerAccounts(logger log.Logger, repo accountRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		accounts, err := repo.GetCustomerAccounts(getCustomerId(w, r))
		if err != nil {
			if requestId := moovhttp.GetRequestId(r); requestId != "" {
				logger.Log("accounts", fmt.Sprintf("%v", err), "requestId", requestId)
			}
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accounts)
	}
}

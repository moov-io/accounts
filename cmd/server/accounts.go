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

func addAccountRoutes(logger log.Logger, r *mux.Router, accountRepo accountRepository, transactionRepo transactionRepository) {
	r.Methods("GET").Path("/accounts/search").HandlerFunc(searchAccounts(logger, accountRepo))

	r.Methods("GET").Path("/customers/{customerId}/accounts").HandlerFunc(getCustomerAccounts(logger, accountRepo))
	r.Methods("POST").Path("/customers/{customerId}/accounts").HandlerFunc(createCustomerAccount(logger, accountRepo, transactionRepo))
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
	Balance int    `json:"balance"`
	Name    string `json:"name"`
	Type    string `json:"type"`
}

func (r createAccountRequest) validate() error {
	if r.Balance < 100 { // $1
		return fmt.Errorf("createAccountRequest: invalid initial amount %d USD cents", r.Balance)
	}
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

func createCustomerAccount(logger log.Logger, accountRepo accountRepository, transactionRepo transactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}
		requestId := moovhttp.GetRequestId(r)
		if requestId == "" {
			requestId = base.ID()
		}

		var req createAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Log("accounts", fmt.Sprintf("error reading JSON request: %v", err), "requestId", requestId)
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			logger.Log("accounts", fmt.Sprintf("error validaing request: %v", err), "requestId", requestId)
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

		if err := accountRepo.CreateAccount(customerId, account); err != nil {
			logger.Log("accounts", fmt.Sprintf("%v", err), "requestId", requestId)
			moovhttp.Problem(w, err)
			return
		}

		// Submit a transaction of the initial amount (where does the exteranl ABA come from)?
		tx := (&createTransactionRequest{
			Lines: []transactionLine{
				{
					AccountId: account.ID,
					Purpose:   ACHCredit,
					Amount:    req.Balance,
				},
			},
		}).asTransaction(base.ID())
		if err := transactionRepo.createTransaction(tx, createTransactionOpts{InitialDeposit: true}); err != nil {
			logger.Log("accounts", fmt.Errorf("problem creating initial balance transaction: %v", err), "requestId", requestId)
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

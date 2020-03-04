// Copyright 2020 The Moov Authors
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

	accounts "github.com/moov-io/accounts/client"
	"github.com/moov-io/base"
	moovhttp "github.com/moov-io/base/http"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	defaultRoutingNumber = os.Getenv("DEFAULT_ROUTING_NUMBER")
)

func addAccountRoutes(logger log.Logger, r *mux.Router, accountRepo accountRepository, transactionRepo transactionRepository) {
	r.Methods("GET").Path("/accounts/search").HandlerFunc(searchAccounts(logger, accountRepo))

	r.Methods("POST").Path("/accounts").HandlerFunc(createAccount(logger, accountRepo, transactionRepo))
}

// searchAccounts will attempt to find Accounts which match all query parameters. Searching with an account number will only
// return one account. Otherwise a 404 will be returned. '400 Bad Request' will be returned if query parameters are missing.
func searchAccounts(logger log.Logger, repo accountRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		q := r.URL.Query()

		// Search for a single account
		reqAcctNumber, reqRoutingNumber, reqAcctType := q.Get("number"), q.Get("routingNumber"), q.Get("type")
		if reqAcctNumber != "" && reqRoutingNumber != "" && reqAcctType != "" {
			// Grab and return accounts
			account, err := repo.SearchAccountsByRoutingNumber(reqAcctNumber, reqRoutingNumber, reqAcctType)
			if err != nil {
				logger.Log("accounts", fmt.Sprintf("error searching accounts: %v", err), "requestID", moovhttp.GetRequestID(r))
				moovhttp.Problem(w, fmt.Errorf("account not found, err=%v", err))
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)

			var accounts []*accounts.Account
			if account != nil {
				accounts = append(accounts, account)
			}
			json.NewEncoder(w).Encode(accounts)
			return
		}

		// Search based on CustomerId
		if customerID := or(q.Get("customerId"), q.Get("customerID")); customerID != "" {
			accounts, err := repo.SearchAccountsByCustomerID(customerID)
			if err != nil {
				logger.Log("accounts", fmt.Sprintf("error getting customer accounts: %v", err), "requestID", moovhttp.GetRequestID(r))
				moovhttp.Problem(w, fmt.Errorf("account not found, err=%v", err))
				return
			}
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(accounts)
		} else {
			// Error if we didn't quit early from query params
			moovhttp.Problem(w, errors.New("missing account search query parameters"))
		}
	}
}

type createAccountRequest struct {
	CustomerID string `json:"customerId"`
	Balance    int    `json:"balance"`
	Name       string `json:"name"`
	Number     string `json:"number"`
	Type       string `json:"type"`
}

func (r createAccountRequest) validate() error {
	if r.CustomerID = strings.TrimSpace(r.CustomerID); r.CustomerID == "" {
		return errors.New("createAccountRequest: empty customerID")
	}
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

func createAccount(logger log.Logger, accountRepo accountRepository, transactionRepo transactionRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}
		requestID := moovhttp.GetRequestID(r)
		if requestID == "" {
			requestID = base.ID()
		}

		var req createAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			logger.Log("accounts", fmt.Sprintf("error reading JSON request: %v", err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			logger.Log("accounts", fmt.Sprintf("error validaing request: %v", err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		now := time.Now()
		account := &accounts.Account{
			ID:            base.ID(),
			CustomerID:    req.CustomerID,
			Name:          req.Name,
			AccountNumber: req.Number,
			RoutingNumber: defaultRoutingNumber,
			Status:        "open",
			Type:          req.Type,
			CreatedAt:     now,
			LastModified:  now,
		}
		// We need to generate a unique account number for this routing number. Right now
		// this involves network calls, but I hope to improve this to something like twitter
		// snowflake or UUID -> number conversion.
		if number, err := generateAccountNumber(account, accountRepo); number != "" {
			account.AccountNumber = number
		} else {
			logger.Log("accounts", fmt.Sprintf("error creating account number: %v", err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		if err := accountRepo.CreateAccount(req.CustomerID, account); err != nil {
			logger.Log("accounts", fmt.Sprintf("error creating account: %v", err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		// Submit a transaction of the initial amount (where does the exteranl ABA come from)?
		tx := (&createTransactionRequest{
			Lines: []transactionLine{
				{
					AccountID: account.ID,
					Purpose:   ACHCredit,
					Amount:    req.Balance,
				},
			},
		}).asTransaction(base.ID())
		if err := transactionRepo.createTransaction(tx, createTransactionOpts{InitialDeposit: true}); err != nil {
			logger.Log("accounts", fmt.Errorf("problem creating initial balance transaction: %v", err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
	}
}

func generateAccountNumber(account *accounts.Account, repo accountRepository) (string, error) {
	number := account.AccountNumber
	if number == "" {
		number = createAccountNumber()
	}
	for i := 0; i < 10; i++ {
		if acct, _ := repo.SearchAccountsByRoutingNumber(number, account.RoutingNumber, account.Type); acct == nil {
			return number, nil
		}
	}
	return "", fmt.Errorf("unable to generate account number for account=%s", account.ID)
}

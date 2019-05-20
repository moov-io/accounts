// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"strconv"
	"time"

	accounts "github.com/moov-io/accounts/client"
	mledge "github.com/moov-io/qledger-sdk-go"
)

var (
	qledgerTimeFormat = time.RFC3339
)

type qledgerAccountRepository struct {
	api *mledge.Ledger
}

func (r *qledgerAccountRepository) Ping() error {
	return r.api.Ping()
}

func (r *qledgerAccountRepository) Close() error {
	return nil
}

func (r *qledgerAccountRepository) GetAccounts(accountIds []string) ([]*accounts.Account, error) {
	var terms []map[string]interface{}
	for i := range accountIds {
		m := make(map[string]interface{})
		m["accountId"] = accountIds[i]
		terms = append(terms, m)
	}
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"should": map[string]interface{}{
				"terms": terms, // OR query to grab what we can for each accountId
			},
		},
	}
	accts, err := r.api.SearchAccounts(query)
	if err != nil {
		return nil, err
	}
	return convertAccounts(accts), nil
}

func (r *qledgerAccountRepository) CreateAccount(customerId string, account *accounts.Account) error {
	data := map[string]interface{}{
		"accountId":        account.Id,
		"customerId":       account.CustomerId,
		"name":             account.Name,
		"accountNumber":    account.AccountNumber,
		"routingNumber":    account.RoutingNumber,
		"status":           account.Status,
		"type":             account.Type,
		"balance":          fmt.Sprintf("%d", account.Balance),
		"balanceAvailable": fmt.Sprintf("%d", account.BalanceAvailable),
		"balancePending":   fmt.Sprintf("%d", account.BalancePending),
		"createdAt":        account.CreatedAt.Format(qledgerTimeFormat),
	}
	if !account.ClosedAt.IsZero() {
		data["closedAt"] = account.ClosedAt.Format(qledgerTimeFormat)
	}
	if !account.LastModified.IsZero() {
		data["lastModified"] = account.LastModified.Format(qledgerTimeFormat)
	}
	return r.api.CreateAccount(&mledge.Account{
		ID:      account.Id,
		Balance: int(account.Balance),
		Data:    data,
	})
}

func (r *qledgerAccountRepository) SearchAccountsByRoutingNumber(accountNumber, routingNumber, acctType string) (*accounts.Account, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"must": map[string]interface{}{
				"terms": []map[string]interface{}{
					{
						"type":          acctType,
						"accountNumber": accountNumber,
						"routingNumber": routingNumber,
					},
				},
			},
		},
	}
	accts, err := r.api.SearchAccounts(query)
	if err != nil {
		return nil, err
	}
	if len(accts) > 0 {
		return convertAccounts(accts)[0], nil
	}
	return nil, nil
}

func (r *qledgerAccountRepository) SearchAccountsByCustomerId(customerId string) ([]*accounts.Account, error) {
	query := map[string]interface{}{
		"query": map[string]interface{}{
			"must": map[string]interface{}{
				"terms": []map[string]interface{}{
					{"customerId": customerId},
				},
			},
		},
	}
	accts, err := r.api.SearchAccounts(query)
	if err != nil {
		return nil, err
	}
	return convertAccounts(accts), nil
}

func setupQLedgerAccountStorage(endpoint, apiToken string) (*qledgerAccountRepository, error) {
	if endpoint == "" || apiToken == "" {
		return nil, fmt.Errorf("qledger: empty endpoint=%q and/or apiToken", endpoint)
	}
	return &qledgerAccountRepository{
		api: mledge.NewLedger(endpoint, apiToken),
	}, nil
}

func convertAccounts(accts []*mledge.Account) []*accounts.Account {
	var out []*accounts.Account
	for i := range accts {
		var closedAt time.Time
		if v, exists := accts[i].Data["closedAt"]; exists {
			closedAt = readTime(v.(string))
		}
		var lastModified time.Time
		if v, exists := accts[i].Data["lastModified"]; exists {
			lastModified = readTime(v.(string))
		}
		out = append(out, &accounts.Account{
			Id:               accts[i].ID,
			CustomerId:       accts[i].Data["customerId"].(string),
			Name:             accts[i].Data["name"].(string),
			AccountNumber:    accts[i].Data["accountNumber"].(string),
			RoutingNumber:    accts[i].Data["routingNumber"].(string),
			Status:           accts[i].Data["status"].(string),
			Type:             accts[i].Data["type"].(string),
			Balance:          int32(accts[i].Balance),
			BalanceAvailable: readBalance(accts[i].Data["balanceAvailable"].(string)),
			BalancePending:   readBalance(accts[i].Data["balancePending"].(string)),
			CreatedAt:        readTime(accts[i].Data["createdAt"].(string)),
			ClosedAt:         closedAt,
			LastModified:     lastModified,
		})
	}
	return out
}

func readBalance(str string) int32 {
	n, _ := strconv.Atoi(str)
	return int32(n)
}

func readTime(str string) time.Time {
	t, _ := time.Parse(qledgerTimeFormat, str)
	return t
}

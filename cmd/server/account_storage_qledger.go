// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/moov-io/gl"
	mledge "github.com/moov-io/qledger-sdk-go"
)

type qledgerAccountRepository struct {
	api *mledge.Ledger
}

func (r *qledgerAccountRepository) Ping() error {
	return r.api.Ping()
}

func (r *qledgerAccountRepository) GetCustomerAccounts(customerId string) ([]*gl.Account, error) {
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

func (r *qledgerAccountRepository) CreateAccount(customerId string, account *gl.Account) error {
	return r.api.CreateAccount(&mledge.Account{
		ID:      account.ID,
		Balance: int(account.Balance),
		Data: map[string]interface{}{
			"customerId":    account.CustomerID,
			"name":          account.Name,
			"accountNumber": account.AccountNumber,
			"routingNumber": account.RoutingNumber,
			"status":        account.Status,
			"type":          account.Type,
			// TODO(adam): BalanceAvailable and BalancePending
		},
	})
}

func (r *qledgerAccountRepository) SearchAccounts(accountNumber, routingNumber, acctType string) (*gl.Account, error) {
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

func setupQLedgerStorage(endpoint, apiToken string) (accountRepository, error) {
	if endpoint == "" || apiToken == "" {
		return nil, fmt.Errorf("qledger: empty endpoint=%q and/or apiToken", endpoint)
	}
	return &qledgerAccountRepository{
		api: mledge.NewLedger(endpoint, apiToken),
	}, nil
}

func convertAccounts(accts []*mledge.Account) []*gl.Account {
	var accounts []*gl.Account
	for i := range accts {
		accounts = append(accounts, &gl.Account{
			ID:            accts[i].ID,
			Balance:       int64(accts[i].Balance),
			CustomerID:    accts[i].Data["customerId"].(string),
			AccountNumber: accts[i].Data["accountNumber"].(string),
			RoutingNumber: accts[i].Data["routingNumber"].(string),
			Name:          accts[i].Data["name"].(string),
			Status:        accts[i].Data["status"].(string),
			Type:          accts[i].Data["type"].(string),
			// TODO(adam): need to read other fields
		})
	}
	return accounts
}

// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package storage

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

func (r *qledgerAccountRepository) GetCustomerAccounts(customerId string) ([]gl.Account, error) {
	return nil, nil
}

func (r *qledgerAccountRepository) CreateAccount(customerId string, name string, acctType string) (*gl.Account, error) { // TOOD(adam): acctType needs strong type
	return nil, nil
}

func (r *qledgerAccountRepository) SearchAccounts(accountNumber, routingNumber, acctType string) (*gl.Account, error) {
	return nil, nil
}

func setupQLedgerStorage(endpoint, apiToken string) (AccountRepository, error) {
	if endpoint == "" || apiToken == "" {
		return nil, fmt.Errorf("qledger: empty endpoint=%q and/or apiToken", endpoint)
	}
	return &qledgerAccountRepository{
		api: mledge.NewLedger(endpoint, apiToken),
	}, nil
}

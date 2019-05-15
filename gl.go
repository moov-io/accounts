// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package gl

import (
	"time"
)

type Account struct {
	ID         string `json:"accountId"`
	CustomerID string `json:"customerId"`

	Name string `json:"name"`

	AccountNumber       string `json:"accountNumber"`
	AccountNumberMasked string `json:"accountNumberMasked"`
	RoutingNumber       string `json:"routingNumber"`

	// Status represents the Account status: Open, Closed
	Status string `json:"status"`
	// Type is the account type: Checking, Savings, FBO
	Type string `json:"type"`

	CreatedAt    time.Time  `json:"createdAt"`
	ClosedAt     *time.Time `json:"closedAt"`
	LastModified *time.Time `json:"lastModified"`

	// Computed fields
	Balance          int64 `json:"balance"`
	BalanceAvailable int64 `json:"balanceAvailable"`
	BalancePending   int64 `json:"balancePending"`
}

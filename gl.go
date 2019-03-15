// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package gl

import (
	"time"
)

type Customer struct {
	ID string `json:"customerId"`

	FirstName  string `json:"firstName"`
	MiddleName string `json:"middleName"`
	LastName   string `json:"lastName"`

	NickName string `json:"nickName"`
	Suffix   string `json:"suffix"`

	BirthDate string `json:"birthDate"`
	Gender    string `json:"gender"`
	Culture   string `json:"culture"`

	// Status holds the Customer's status: Applied, Verified, Denied, Archieved, Deceased
	Status string `json:"status"`

	Email     string    `json:"email"`
	Phones    []Phone   `json:"phones"`
	Addresses []Address `json:"address"`

	CreatedDate  time.Time `json:"createdAt"`
	LastModified time.Time `json:"lastModified"`
}

type Phone struct {
	Number string `json:"number"`
	Valid  bool   `json:"valid"`
	// Type represents the number's usage: Home, Mobile, Work
	Type string `json:"type"`
}

type Address struct {
	// Type represents if the address is a Primary residence or Secondary
	Type string `json:"type"`

	Address1   string `json:"address1"`
	Address2   string `json:"address2"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`

	Validated bool `json:"validated"`
	Active    bool `json:"active"`
}

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

	CreatedDate  time.Time `json:"createdAt"`
	ClosedAt     time.Time `json:"closedAt"`
	LastModified time.Time `json:"lastModified"`

	Balance          int64 `json:"balance"`
	BalanceAvailable int64 `json:"balanceAvailable"`
	BalancePending   int64 `json:"balancePending"`
}

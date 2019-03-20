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

var (
	mockCustomer = &gl.Customer{
		ID:         base.ID(),
		FirstName:  "Jane",
		MiddleName: "",
		LastName:   "Doe",
		NickName:   "",
		Suffix:     "",
		BirthDate:  time.Date(1990, time.January, 1, 6, 19, 0, 0, time.UTC),
		Gender:     "female",
		Culture:    "en-US",
		Status:     "Applied",
		Email:      "jane.doe@example.com",
		Phones: []gl.Phone{
			{
				Number: "+15555555555",
				Valid:  false,
				Type:   "Mobile",
			},
		},
		Addresses: []gl.Address{
			{
				Type:       "Primary",
				Address1:   "123 4th St",
				Address2:   "",
				City:       "Los Angeles",
				State:      "CA",
				PostalCode: "90210",
				Country:    "USA",
				Validated:  true,
				Active:     true,
			},
		},
		// two arbitrary time.Time values in the past
		CreatedAt:    time.Now().Add(-1500 * time.Hour),
		LastModified: time.Now().Add(-365 * time.Hour),
	}
)

func addCustomerRoutes(logger log.Logger, r *mux.Router) {
	r.Methods("GET").Path("/customers/{customerId}").HandlerFunc(getCustomer(logger))
	r.Methods("POST").Path("/customers").HandlerFunc(createCustomer(logger))
}

func getCustomer(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mockCustomer)
	}
}

type customerRequest struct {
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Email     string    `json:"email"`
	Phones    []phone   `json:"phones"`
	Addresses []address `json:"addresses"`
}

type phone struct {
	Number string `json:"number"`
	Type   string `json:"type"`
}

type address struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

func (req customerRequest) validate() error {
	if req.FirstName == "" || req.LastName == "" {
		return errors.New("create customer: empty name field(s)")
	}
	if req.Email == "" {
		return errors.New("create customer: empty email")
	}
	if len(req.Phones) == 0 {
		return errors.New("create customer: phone array is required")
	}
	if len(req.Addresses) == 0 {
		return errors.New("create customer: address array is required")
	}
	return nil
}

func createCustomer(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		var req customerRequest
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
		json.NewEncoder(w).Encode(mockCustomer)

	}
}

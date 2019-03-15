// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/gl"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func addCustomerRoutes(logger log.Logger, r *mux.Router) {
	r.Methods("GET").Path("/customers/{customerId}").HandlerFunc(getCustomer(logger))
}

func getCustomer(logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w, err := wrapResponseWriter(logger, w, r)
		if err != nil {
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&gl.Customer{
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
		})
	}
}

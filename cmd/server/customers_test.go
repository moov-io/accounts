// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moov-io/gl"

	"github.com/gorilla/mux"
)

func TestCustomers__GetCustomer(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo", nil)
	req.Header.Set("x-user-id", "test")

	router := mux.NewRouter()
	addCustomerRoutes(nil, router)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var cust gl.Customer // TODO(adam): check more of Customer response?
	if err := json.NewDecoder(w.Body).Decode(&cust); err != nil {
		t.Fatal(err)
	}
	if cust.ID == "" {
		t.Error("empty Customer.ID")
	}
}

func TestCustomers__customerRequest(t *testing.T) {
	req := &customerRequest{}
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}
	req.FirstName = "jane"
	req.LastName = "doe"
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}
	req.Email = "jane.doe@example.com"
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}
	req.Phones = append(req.Phones, phone{
		Number: "123.456.7890",
		Type:   "Checking",
	})
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}
	req.Addresses = append(req.Addresses, address{
		Address1:   "123 1st st",
		City:       "fake city",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
	})
	if err := req.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestCustomers__createCustomer(t *testing.T) {
	w := httptest.NewRecorder()
	phone := `{"number": "555.555.5555", "type": "mobile"}`
	address := `{"type": "home", "address1": "123 1st St", "city": "Denver", "state": "CO", "postalCode": "12345", "country": "USA"}`
	body := fmt.Sprintf(`{"firstName": "jane", "lastName": "doe", "email": "jane@example.com", "phones": [%s], "addresses": [%s]}`, phone, address)
	req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))
	req.Header.Set("x-user-id", "test")

	router := mux.NewRouter()
	addCustomerRoutes(nil, router)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var cust gl.Customer // TODO(adam): check more of Customer response?
	if err := json.NewDecoder(w.Body).Decode(&cust); err != nil {
		t.Fatal(err)
	}
	if cust.ID == "" {
		t.Error("empty Customer.ID")
	}

	// sad path
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/customers", strings.NewReader("null"))
	req.Header.Set("x-user-id", "test")
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("bogus HTTP status code: %d", w.Code)
	}
}

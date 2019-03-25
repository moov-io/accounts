package ledger

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	// AccountsAPI is the endpoint for creating and updating accounts
	AccountsAPI = "/v1/accounts"
	// AccountsSearchAPI is the endpoint for searching accounts
	AccountsSearchAPI = "/v1/accounts/_search"
)

// Account holds information such as account ID, balance and search data
type Account struct {
	ID      string                 `json:"id"`
	Balance int                    `json:"balance"`
	Data    map[string]interface{} `json:"data"`
}

// SearchAccounts returns all account in the Ledger which satisfies the given search query
func (ledger *Ledger) SearchAccounts(searchQuery map[string]interface{}) ([]*Account, error) {
	payload, err := json.Marshal(searchQuery)
	if err != nil {
		return nil, err
	}

	// Search the account
	res, err := ledger.DoRequest(http.MethodPost, AccountsSearchAPI, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, ErrInternalServer
	}

	// Parse the accounts result
	var accounts []*Account
	err = unmarshalResponse(res, &accounts)
	if err != nil {
		return nil, err
	}

	// Return the account
	return accounts, nil
}

// GetAccount returns an account in the Ledger which matches the given ID
func (ledger *Ledger) GetAccount(accountID string) (*Account, error) {
	// Prepare search query
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"must": map[string]interface{}{
				"fields": []map[string]interface{}{
					{"id": map[string]interface{}{"eq": accountID}},
				},
			},
		},
	}

	// Search in the Ledger
	accounts, err := ledger.SearchAccounts(searchQuery)
	if err != nil {
		return nil, err
	}

	// Return the account
	if len(accounts) == 0 {
		return nil, ErrAccountNotfound
	}
	return accounts[0], nil
}

// CreateAccount creates a new account to QLedger
func (ledger *Ledger) CreateAccount(account *Account) error {
	payload, err := json.Marshal(account)
	if err != nil {
		return err
	}

	res, err := ledger.DoRequest(http.MethodPost, AccountsAPI, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	switch res.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusBadRequest:
		return ErrAccountInvalid
	case http.StatusConflict:
		return ErrAccountConflict
	default:
		return ErrInternalServer
	}
}

// UpdateAccount creates a new account to QLedger
func (ledger *Ledger) UpdateAccount(account *Account) error {
	payload, err := json.Marshal(account)
	if err != nil {
		return err
	}

	res, err := ledger.DoRequest(http.MethodPut, AccountsAPI, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return ErrAccountInvalid
	case http.StatusNotFound:
		return ErrAccountNotfound
	case http.StatusConflict:
		return ErrAccountConflict
	default:
		return ErrInternalServer
	}
}

package ledger

import (
	"bytes"
	"encoding/json"
	"net/http"
)

const (
	// TransactionsAPI is the endpoint for creating and updating transactions
	TransactionsAPI = "/v1/transactions"
	// TransactionsSearchAPI is the endpoint for searching transactions
	TransactionsSearchAPI = "/v1/transactions/_search"
)

// Transaction holds information such as transaction ID, timestamp,
// movement of money between accounts and search data
type Transaction struct {
	ID        string                 `json:"id"`
	Data      map[string]interface{} `json:"data"`
	Timestamp string                 `json:"timestamp"`
	Lines     []*TransactionLine     `json:"lines"`
}

// TransactionLine holds information about the account and amount transferred
type TransactionLine struct {
	AccountID string `json:"account"`
	Delta     int    `json:"delta"`
}

// CreateTransaction makes a new transaction to QLedger
func (ledger *Ledger) CreateTransaction(transaction *Transaction) error {
	payload, err := json.Marshal(transaction)
	if err != nil {
		return err
	}

	res, err := ledger.DoRequest(http.MethodPost, TransactionsAPI, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	switch res.StatusCode {
	case http.StatusCreated:
		return nil
	case http.StatusAccepted:
		return ErrTransactionDuplicate
	case http.StatusBadRequest:
		return ErrTransactionInvalid
	case http.StatusConflict:
		return ErrTransactionConflict
	default:
		return ErrInternalServer
	}
}

// SearchTransactions returns all transactions in the Ledger which satisfies the given search query
func (ledger *Ledger) SearchTransactions(searchQuery map[string]interface{}) ([]*Transaction, error) {
	payload, _ := json.Marshal(searchQuery)

	// Search the transaction
	res, err := ledger.DoRequest(http.MethodPost, TransactionsSearchAPI, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, ErrInternalServer
	}

	// Parse the transactions result
	var transactions []*Transaction
	err = unmarshalResponse(res, &transactions)
	if err != nil {
		return nil, err
	}

	// Return the transaction
	return transactions, nil
}

// GetTransaction returns a transaction in the Ledger which matches the given ID
func (ledger *Ledger) GetTransaction(txnID string) (*Transaction, error) {
	// Prepare search query
	searchQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"must": map[string]interface{}{
				"fields": []map[string]interface{}{
					{"id": map[string]interface{}{"eq": txnID}},
				},
			},
		},
	}

	// Search in Ledger
	transactions, err := ledger.SearchTransactions(searchQuery)
	if err != nil {
		return nil, err
	}

	// Return the transaction
	if len(transactions) == 0 {
		return nil, ErrTransactionNotfound
	}
	return transactions[0], nil
}

// UpdateTransaction updates `data` of existing transaction in QLedger
func (ledger *Ledger) UpdateTransaction(txn *Transaction) error {
	payload, err := json.Marshal(txn)
	if err != nil {
		return err
	}

	res, err := ledger.DoRequest("PUT", TransactionsAPI, bytes.NewBuffer(payload))
	if err != nil {
		return ErrInternalServer
	}

	switch res.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusBadRequest:
		return ErrTransactionInvalid
	case http.StatusNotFound:
		return ErrTransactionNotfound
	case http.StatusConflict:
		return ErrTransactionConflict
	default:
		return ErrInternalServer
	}
}

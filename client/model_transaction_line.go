/*
 * Simple Core System API
 *
 * Moov GL is an HTTP service which represents both a general ledger and chart of accounts for customers. The service is designed to abstract over various core systems and provide a uniform API for developers.
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

type TransactionLine struct {
	// Account ID
	AccountId string `json:"accountId,omitempty"`
	Purpose   string `json:"purpose,omitempty"`
	// Change in account balance (in USD cents)
	Amount float32 `json:"amount,omitempty"`
}
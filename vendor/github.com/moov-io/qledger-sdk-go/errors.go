package ledger

import "errors"

var (
	ErrTransactionDuplicate = errors.New("Transaction is duplicate")
	ErrTransactionConflict  = errors.New("Transaction ID conflict")
	ErrTransactionInvalid   = errors.New("Transaction invalid")
	ErrTransactionNotfound  = errors.New("Transaction not found")

	ErrAccountConflict = errors.New("Account ID conflict")
	ErrAccountInvalid  = errors.New("Account invalid")
	ErrAccountNotfound = errors.New("Account not found")

	ErrInternalServer = errors.New("Ledger internal server error")
)

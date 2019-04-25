package models

import (
	"github.com/RealImage/QLedger/errors"
)

// SearchNamespaceInvalidError returns invalid search namespace error type
func SearchNamespaceInvalidError(namespace string) errors.ApplicationError {
	return &errors.BaseApplicationError{
		Code:    "search.namespace.invalid",
		Message: "Invalid search namespace: " + namespace,
	}
}

// SearchQueryInvalidError returns invalid search query error type
func SearchQueryInvalidError(err error) errors.ApplicationError {
	return &errors.BaseApplicationError{
		Code:    "search.query.invalid",
		Message: "Invalid search query: " + err.Error(),
	}
}

// DBError returns db error type
func DBError(err error) errors.ApplicationError {
	return &errors.BaseApplicationError{
		Code:    "db.error",
		Message: "DB Error: " + err.Error(),
	}
}

// JSONError returns invalid json error type
func JSONError(err error) errors.ApplicationError {
	return &errors.BaseApplicationError{
		Code:    "json.error",
		Message: "JSON Error: " + err.Error(),
	}
}

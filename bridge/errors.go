package main

import "errors"

var (
	ErrPackageManager     = errors.New("package manager failure")
	ErrValidation         = errors.New("validation error")
	ErrPermission         = errors.New("permission denied")
	ErrMetricsUnavailable = errors.New("metrics provider unavailable")
)

const (
	CodePackageManagerFailure = 1001
	CodeValidationFailed      = 1002
	CodePermissionDenied      = 1003
	CodeMetricsUnavailable    = 1004
)

type respError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

func wrapError(err error) *respError {
	switch {
	case errors.Is(err, ErrPackageManager):
		return &respError{Code: CodePackageManagerFailure, Message: "package manager failed", Details: err.Error()}
	case errors.Is(err, ErrValidation):
		return &respError{Code: CodeValidationFailed, Message: "validation failed", Details: err.Error()}
	case errors.Is(err, ErrPermission):
		return &respError{Code: CodePermissionDenied, Message: "permission denied", Details: err.Error()}
	case errors.Is(err, ErrMetricsUnavailable):
		return &respError{Code: CodeMetricsUnavailable, Message: "metrics provider unavailable", Details: err.Error()}
	default:
		return &respError{Code: -1, Message: err.Error(), Details: err.Error()}
	}
}

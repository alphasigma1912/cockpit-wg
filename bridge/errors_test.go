package main

import (
	"errors"
	"fmt"
	"testing"
)

func TestWrapErrorCodes(t *testing.T) {
	cases := []struct {
		err  error
		code int
	}{
		{fmt.Errorf("%w", ErrPackageManager), CodePackageManagerFailure},
		{fmt.Errorf("%w", ErrValidation), CodeValidationFailed},
		{fmt.Errorf("%w", ErrPermission), CodePermissionDenied},
		{fmt.Errorf("%w", ErrMetricsUnavailable), CodeMetricsUnavailable},
		{errors.New("other"), -1},
	}
	for _, c := range cases {
		re := wrapError(c.err)
		if re.Code != c.code {
			t.Fatalf("expected code %d got %d", c.code, re.Code)
		}
	}
}

func TestValidateConfigError(t *testing.T) {
	if _, err := validateConfig("invalid"); err == nil || !errors.Is(err, ErrValidation) {
		t.Fatalf("expected validation error")
	}
}

func TestAuthorizeDenied(t *testing.T) {
	if err := authorize("UnknownMethod"); err == nil || !errors.Is(err, ErrPermission) {
		t.Fatalf("expected permission error")
	}
}

func TestGetMetricsUnavailable(t *testing.T) {
	collector = nil
	if _, err := getMetrics("wg0"); err == nil || !errors.Is(err, ErrMetricsUnavailable) {
		t.Fatalf("expected metrics unavailable error")
	}
}

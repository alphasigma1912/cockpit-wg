package main

import "testing"

func TestAuthorizeRejectsUnknownMethod(t *testing.T) {
	if err := authorize("BogusMethod"); err == nil {
		t.Fatalf("expected rejection for unknown method")
	}
}

func TestAuthorizeAllowsKnownMethod(t *testing.T) {
	if err := authorize("ListInterfaces"); err != nil {
		t.Fatalf("expected allowed method, got %v", err)
	}
}

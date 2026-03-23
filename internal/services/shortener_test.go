package services

import "testing"

func TestNotFoundError(t *testing.T) {
	if ErrNotFound == nil {
		t.Fatalf("ErrNotFound should not be nil")
	}

	if ErrNotFound.Error() != "not found" {
		t.Fatalf("unexpected error string: got %q", ErrNotFound.Error())
	}

	// type check
	if _, ok := interface{}(ErrNotFound).(*NotFoundError); !ok {
		t.Fatalf("ErrNotFound should be of type *NotFoundError")
	}
}

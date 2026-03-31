package util

import (
	"errors"
	"testing"
)

type testStatusError struct {
	code int
}

func (e testStatusError) Error() string {
	return "status"
}

func (e testStatusError) HTTPStatusCode() int {
	return e.code
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{name: "blocked sentinel", err: ErrBlocked, expected: "blocked"},
		{name: "timeout sentinel", err: ErrTimeout, expected: "timeout"},
		{name: "http sentinel", err: ErrHTTP, expected: "http_error"},
		{name: "status 403", err: testStatusError{code: 403}, expected: "blocked"},
		{name: "status 429", err: testStatusError{code: 429}, expected: "blocked"},
		{name: "status 500", err: testStatusError{code: 500}, expected: "http_error"},
		{name: "wrapped timeout", err: errors.Join(errors.New("x"), ErrTimeout), expected: "timeout"},
		{name: "fallback http", err: errors.New("oops"), expected: "http_error"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := ClassifyError(tc.err); got != tc.expected {
				t.Fatalf("ClassifyError(%v)=%q, want %q", tc.err, got, tc.expected)
			}
		})
	}
}

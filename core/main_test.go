package main

import (
	"bytes"
	"testing"
)

func TestProcessRequestErrorContract(t *testing.T) {
	tests := []struct {
		name          string
		input         []byte
		expectedError string
	}{
		{name: "invalid json", input: []byte(`{"action":`), expectedError: "invalid input"},
		{name: "invalid action", input: []byte(`{"action":"nope"}`), expectedError: "invalid action"},
		{name: "missing search query", input: []byte(`{"action":"search"}`), expectedError: "query required"},
		{name: "missing fetch url", input: []byte(`{"action":"fetch"}`), expectedError: "url required"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := processRequest(tc.input)
			if err != nil {
				t.Fatalf("expected nil error, got %v", err)
			}

			if resp.Error != tc.expectedError {
				t.Fatalf("expected error %q, got %q", tc.expectedError, resp.Error)
			}
		})
	}
}

func TestProcessRequestInputTooLarge(t *testing.T) {
	oversized := bytes.Repeat([]byte{'a'}, maxStdinBytes+1)
	resp, err := processRequest(oversized)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if resp.Error != "input too large" {
		t.Fatalf("expected input too large error, got %q", resp.Error)
	}
}

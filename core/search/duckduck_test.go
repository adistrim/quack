package search

import "testing"

func TestNormalizeMaxResults(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{name: "default when zero", input: 0, expected: defaultMaxResults},
		{name: "clamp negative", input: -5, expected: minMaxResults},
		{name: "clamp too large", input: 200, expected: maxMaxResults},
		{name: "keep valid", input: 20, expected: 20},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeMaxResults(tc.input)
			if got != tc.expected {
				t.Fatalf("NormalizeMaxResults(%d)=%d, want %d", tc.input, got, tc.expected)
			}
		})
	}
}

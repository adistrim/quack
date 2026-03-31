package search

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

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

func TestParseSearchResultsUsesFallbackSelectorsAndSkipsMalformedBlocks(t *testing.T) {
	html := `
	<div class="result">
	  <h2><a href="https://example.com/a">Result A</a></h2>
	  <p>Snippet A</p>
	</div>
	<div class="result">
	  <div class="result__title"><a href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2Fb">Result B</a></div>
	  <div class="snippet">Snippet B</div>
	</div>
	<div class="result">
	  <div class="result__title"><a href="//duckduckgo.com/l/?x=missing-uddg">Bad Redirect</a></div>
	  <div class="result__snippet">Should be skipped</div>
	</div>
	<div class="result">
	  <div class="result__title"></div>
	</div>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("failed to parse test html: %v", err)
	}

	results := parseSearchResults(doc, 10)
	if len(results) != 2 {
		t.Fatalf("expected 2 valid results, got %d", len(results))
	}

	if results[0].Title != "Result A" || results[0].Url != "https://example.com/a" || results[0].Snippet != "Snippet A" || results[0].Rank != 1 {
		t.Fatalf("unexpected first result: %+v", results[0])
	}

	if results[1].Title != "Result B" || results[1].Url != "https://example.com/b" || results[1].Snippet != "Snippet B" || results[1].Rank != 2 {
		t.Fatalf("unexpected second result: %+v", results[1])
	}
}

func TestNormalizeResultURL(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected string
		ok       bool
	}{
		{name: "plain https", raw: "https://example.com", expected: "https://example.com", ok: true},
		{name: "protocol relative", raw: "//example.com/path", expected: "https://example.com/path", ok: true},
		{name: "ddg redirect", raw: "//duckduckgo.com/l/?uddg=https%3A%2F%2Fexample.com%2Fz&rut=x", expected: "https://example.com/z", ok: true},
		{name: "ddg redirect missing uddg", raw: "//duckduckgo.com/l/?rut=x", expected: "", ok: false},
		{name: "empty", raw: "  ", expected: "", ok: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := normalizeResultURL(tc.raw)
			if got != tc.expected || ok != tc.ok {
				t.Fatalf("normalizeResultURL(%q)=(%q,%v), want (%q,%v)", tc.raw, got, ok, tc.expected, tc.ok)
			}
		})
	}
}

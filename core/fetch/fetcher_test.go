package fetch

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"quack/util"
)

func TestValidateTargetURLBlockedTargets(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
	}{
		{name: "loopback", rawURL: "http://127.0.0.1/path"},
		{name: "private rfc1918", rawURL: "http://10.0.0.3/path"},
		{name: "metadata endpoint", rawURL: "http://169.254.169.254/latest/meta-data"},
		{name: "credentials", rawURL: "https://user:pass@example.com"},
		{name: "unsupported scheme", rawURL: "ftp://example.com"},
		{name: "missing host", rawURL: "https:///missing"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTargetURL(tc.rawURL)
			if err == nil {
				t.Fatalf("expected blocked error for %q", tc.rawURL)
			}

			if !errors.Is(err, util.ErrBlocked) {
				t.Fatalf("expected util.ErrBlocked, got %v", err)
			}
		})
	}
}

func TestValidateTargetURLAllowsPublicIPLiteral(t *testing.T) {
	err := validateTargetURL("https://93.184.216.34/")
	if err != nil {
		t.Fatalf("expected public literal IP to be allowed, got %v", err)
	}
}

func TestValidateTargetURLBlocksDNSResolutionToPrivateRange(t *testing.T) {
	originalLookup := lookupIPAddr
	lookupIPAddr = func(_ context.Context, host string) ([]net.IPAddr, error) {
		if host != "example.com" {
			t.Fatalf("unexpected host: %s", host)
		}
		return []net.IPAddr{{IP: net.ParseIP("10.1.2.3")}}, nil
	}
	defer func() { lookupIPAddr = originalLookup }()

	err := validateTargetURL("https://example.com/path")
	if err == nil {
		t.Fatal("expected blocked error")
	}

	if !errors.Is(err, util.ErrBlocked) {
		t.Fatalf("expected util.ErrBlocked, got %v", err)
	}
}

func TestRedirectValidationBlocksBlockedTarget(t *testing.T) {
	f := New()

	redirectReq := &http.Request{URL: &url.URL{Scheme: "http", Host: "127.0.0.1", Path: "/private"}}
	via := []*http.Request{{URL: &url.URL{Scheme: "https", Host: "93.184.216.34", Path: "/"}}}

	err := f.client.CheckRedirect(redirectReq, via)
	if err == nil {
		t.Fatal("expected redirect to blocked target to fail")
	}

	if !errors.Is(err, util.ErrBlocked) {
		t.Fatalf("expected util.ErrBlocked, got %v", err)
	}
}

func TestFetchMapsTimeoutErrorClassification(t *testing.T) {
	f := New()
	f.client.Transport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
		return nil, context.DeadlineExceeded
	})

	_, err := f.Fetch("http://93.184.216.34/")
	if err == nil {
		t.Fatal("expected timeout error")
	}

	if !errors.Is(err, util.ErrTimeout) {
		t.Fatalf("expected util.ErrTimeout, got %v", err)
	}
}

func TestFetchTruncatesLongContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte("<html><body>" + repeatA(maxTextBytes+100) + "</body></html>"))
	}))
	defer server.Close()

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}

	f := New()
	f.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		rewritten := req.Clone(req.Context())
		rewritten.URL.Scheme = target.Scheme
		rewritten.URL.Host = target.Host
		rewritten.Host = target.Host
		return http.DefaultTransport.RoundTrip(rewritten)
	})

	result, err := f.Fetch("http://93.184.216.34/")
	if err != nil {
		t.Fatalf("expected fetch success, got %v", err)
	}

	if !result.Truncated {
		t.Fatal("expected truncated flag to be true")
	}

	if !strings.HasSuffix(result.Text, "... [content truncated]") {
		t.Fatalf("expected truncation suffix, got %q", result.Text)
	}
}

func TestFetchRejectsUnsupportedContentType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.4"))
	}))
	defer server.Close()

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("failed to parse server URL: %v", err)
	}

	f := New()
	f.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		rewritten := req.Clone(req.Context())
		rewritten.URL.Scheme = target.Scheme
		rewritten.URL.Host = target.Host
		rewritten.Host = target.Host
		return http.DefaultTransport.RoundTrip(rewritten)
	})

	_, err = f.Fetch("http://93.184.216.34/")
	if err == nil {
		t.Fatal("expected content-type error")
	}

	var ctErr *ContentTypeError
	if !errors.As(err, &ctErr) {
		t.Fatalf("expected ContentTypeError, got %T", err)
	}

	if !errors.Is(err, util.ErrHTTP) {
		t.Fatalf("expected util.ErrHTTP, got %v", err)
	}
}

func TestFetchHTTPStatusErrorType(t *testing.T) {
	err := &HTTPStatusError{StatusCode: 429}
	if !errors.Is(err, util.ErrBlocked) {
		t.Fatalf("expected 429 to unwrap to blocked")
	}

	err = &HTTPStatusError{StatusCode: 500}
	if !errors.Is(err, util.ErrHTTP) {
		t.Fatalf("expected 500 to unwrap to http")
	}
}

func TestIsSupportedContentType(t *testing.T) {
	tests := []struct {
		name        string
		contentType string
		expected    bool
	}{
		{name: "empty accepted", contentType: "", expected: true},
		{name: "html accepted", contentType: "text/html; charset=utf-8", expected: true},
		{name: "xhtml accepted", contentType: "application/xhtml+xml", expected: true},
		{name: "xml accepted", contentType: "application/xml", expected: true},
		{name: "svg accepted", contentType: "image/svg+xml", expected: true},
		{name: "pdf rejected", contentType: "application/pdf", expected: false},
		{name: "invalid header rejected", contentType: "text/html; charset", expected: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isSupportedContentType(tc.contentType)
			if got != tc.expected {
				t.Fatalf("isSupportedContentType(%q)=%v, want %v", tc.contentType, got, tc.expected)
			}
		})
	}
}

func TestTruncateUTF8ByBytes(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		limit        int
		expectedText string
		expectedCut  bool
	}{
		{name: "no truncate", input: "hello", limit: 10, expectedText: "hello", expectedCut: false},
		{name: "ascii truncate", input: "abcdef", limit: 4, expectedText: "abcd", expectedCut: true},
		{name: "utf8 rune boundary", input: "A世B", limit: 2, expectedText: "A", expectedCut: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			text, cut := truncateUTF8ByBytes(tc.input, tc.limit)
			if text != tc.expectedText || cut != tc.expectedCut {
				t.Fatalf("truncateUTF8ByBytes(%q,%d)=(%q,%v), want (%q,%v)", tc.input, tc.limit, text, cut, tc.expectedText, tc.expectedCut)
			}
		})
	}
}

func TestMapRequestError(t *testing.T) {
	timeoutErr := mapRequestError(context.DeadlineExceeded)
	if !errors.Is(timeoutErr, util.ErrTimeout) {
		t.Fatalf("expected timeout classification, got %v", timeoutErr)
	}

	plain := errors.New("plain")
	if mapRequestError(plain) != plain {
		t.Fatalf("expected plain error passthrough")
	}

	netTimeout := &fakeTimeoutError{}
	classified := mapRequestError(fmt.Errorf("wrap: %w", netTimeout))
	if !errors.Is(classified, util.ErrTimeout) {
		t.Fatalf("expected wrapped net timeout classification, got %v", classified)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func repeatA(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = 'a'
	}
	return string(b)
}

type fakeTimeoutError struct{}

func (e *fakeTimeoutError) Error() string   { return "timeout" }
func (e *fakeTimeoutError) Timeout() bool   { return true }
func (e *fakeTimeoutError) Temporary() bool { return true }

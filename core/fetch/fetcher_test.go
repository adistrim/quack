package fetch

import (
	"context"
	"errors"
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

	text, err := f.Fetch("http://93.184.216.34/")
	if err != nil {
		t.Fatalf("expected fetch success, got %v", err)
	}

	if len(text) <= maxTextBytes {
		t.Fatalf("expected text with truncation suffix, got length %d", len(text))
	}

	if !strings.HasSuffix(text, "... [content truncated]") {
		t.Fatalf("expected truncation suffix, got %q", text)
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

package fetch

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"quack/util"
)

const (
	maxTextBytes      = 8000
	maxFetchBodyBytes = 4 * 1024 * 1024
)

var blockedCIDRs = mustParseCIDRs([]string{
	"0.0.0.0/8",
	"10.0.0.0/8",
	"127.0.0.0/8",
	"169.254.0.0/16",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"224.0.0.0/4",
	"240.0.0.0/4",
	"::/128",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
	"ff00::/8",
})

var lookupIPAddr = net.DefaultResolver.LookupIPAddr

type Fetcher struct {
	client *http.Client
}

type Result struct {
	Text      string
	Truncated bool
}

type HTTPStatusError struct {
	StatusCode int
}

func (e *HTTPStatusError) Error() string {
	return fmt.Sprintf("http status %d", e.StatusCode)
}

func (e *HTTPStatusError) Unwrap() error {
	if e.StatusCode == http.StatusForbidden || e.StatusCode == http.StatusTooManyRequests {
		return util.ErrBlocked
	}

	return util.ErrHTTP
}

func (e *HTTPStatusError) HTTPStatusCode() int {
	return e.StatusCode
}

type ContentTypeError struct {
	ContentType string
}

func (e *ContentTypeError) Error() string {
	return fmt.Sprintf("unsupported content type %q", e.ContentType)
}

func (e *ContentTypeError) Unwrap() error {
	return util.ErrHTTP
}

func New() *Fetcher {
	checkRedirect := func(req *http.Request, via []*http.Request) error {
		if len(via) >= 10 {
			return fmt.Errorf("%w: too many redirects", util.ErrHTTP)
		}

		if err := validateTargetURL(req.URL.String()); err != nil {
			return err
		}

		return nil
	}

	return &Fetcher{
		client: &http.Client{
			Timeout:       30 * time.Second,
			CheckRedirect: checkRedirect,
		},
	}
}

func (f *Fetcher) Fetch(rawURL string) (Result, error) {
	if err := validateTargetURL(rawURL); err != nil {
		return Result{}, err
	}

	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return Result{}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return Result{}, mapRequestError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Result{}, &HTTPStatusError{StatusCode: resp.StatusCode}
	}

	if !isSupportedContentType(resp.Header.Get("Content-Type")) {
		return Result{}, &ContentTypeError{ContentType: resp.Header.Get("Content-Type")}
	}

	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, maxFetchBodyBytes))
	if err != nil {
		return Result{}, err
	}

	doc.Find("script, style, nav, header, footer").Remove()

	text := util.CleanWhitespace(doc.Text())
	truncatedText, truncated := truncateUTF8ByBytes(text, maxTextBytes)
	if truncated {
		truncatedText += "... [content truncated]"
	}

	return Result{Text: truncatedText, Truncated: truncated}, nil
}

func validateTargetURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("%w: invalid url", util.ErrBlocked)
	}

	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("%w: unsupported scheme", util.ErrBlocked)
	}

	if u.User != nil {
		return fmt.Errorf("%w: credentials are not allowed", util.ErrBlocked)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("%w: missing host", util.ErrBlocked)
	}

	ip := net.ParseIP(host)
	if ip != nil {
		if isBlockedIP(ip) {
			return fmt.Errorf("%w: blocked target", util.ErrBlocked)
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	addrs, err := lookupIPAddr(ctx, host)
	if err != nil {
		return fmt.Errorf("%w: dns lookup failed", util.ErrHTTP)
	}

	if len(addrs) == 0 {
		return fmt.Errorf("%w: dns lookup returned no addresses", util.ErrHTTP)
	}

	for _, addr := range addrs {
		if isBlockedIP(addr.IP) {
			return fmt.Errorf("%w: blocked target", util.ErrBlocked)
		}
	}

	return nil
}

func isBlockedIP(ip net.IP) bool {
	for _, blocked := range blockedCIDRs {
		if blocked.Contains(ip) {
			return true
		}
	}

	return false
}

func mustParseCIDRs(cidrs []string) []*net.IPNet {
	out := make([]*net.IPNet, 0, len(cidrs))
	for _, cidr := range cidrs {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			panic(errors.New("invalid cidr config: " + cidr))
		}
		out = append(out, network)
	}

	return out
}

func mapRequestError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("%w: request timeout", util.ErrTimeout)
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return fmt.Errorf("%w: request timeout", util.ErrTimeout)
	}

	return err
}

func isSupportedContentType(rawContentType string) bool {
	if rawContentType == "" {
		return true
	}

	mediaType, _, err := mime.ParseMediaType(rawContentType)
	if err != nil {
		return false
	}

	if mediaType == "text/html" || mediaType == "application/xhtml+xml" {
		return true
	}

	if mediaType == "text/xml" || mediaType == "application/xml" {
		return true
	}

	return strings.HasSuffix(mediaType, "+xml")
}

func truncateUTF8ByBytes(s string, limit int) (string, bool) {
	if len(s) <= limit {
		return s, false
	}

	cut := 0
	for i := range s {
		if i > limit {
			break
		}
		cut = i
	}

	if cut == 0 {
		return "", true
	}

	return s[:cut], true
}

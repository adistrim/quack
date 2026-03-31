package search

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"quack/models"
)

const (
	ddgURL             = "https://html.duckduckgo.com/html"
	userAgent          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
	defaultMaxResults  = 10
	minMaxResults      = 1
	maxMaxResults      = 50
	maxSearchBodyBytes = 2 * 1024 * 1024
)

type DuckDuckGo struct {
	client *http.Client
}

func NewDuckDuckGo() *DuckDuckGo {
	return &DuckDuckGo{
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *DuckDuckGo) Search(query string, maxResults int) ([]models.SearchResult, error) {
	maxResults = NormalizeMaxResults(maxResults)

	form := url.Values{}
	form.Set("q", query)
	form.Set("b", "")
	form.Set("kl", "")

	req, err := http.NewRequest("POST", ddgURL, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("duckduckgo status %d", resp.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(io.LimitReader(resp.Body, maxSearchBodyBytes))
	if err != nil {
		return nil, err
	}

	return parseSearchResults(doc, maxResults), nil
}

func NormalizeMaxResults(maxResults int) int {
	if maxResults == 0 {
		return defaultMaxResults
	}

	if maxResults < minMaxResults {
		return minMaxResults
	}

	if maxResults > maxMaxResults {
		return maxMaxResults
	}

	return maxResults
}

func parseSearchResults(doc *goquery.Document, maxResults int) []models.SearchResult {
	results := make([]models.SearchResult, 0, maxResults)

	doc.Find(".result").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		result, ok := parseResultBlock(sel)
		if !ok {
			return true
		}

		result.Rank = len(results) + 1
		results = append(results, result)
		return len(results) < maxResults
	})

	return results
}

func parseResultBlock(sel *goquery.Selection) (models.SearchResult, bool) {
	title, link, ok := extractTitleAndLink(sel)
	if !ok {
		return models.SearchResult{}, false
	}

	normalized, ok := normalizeResultURL(link)
	if !ok {
		return models.SearchResult{}, false
	}

	return models.SearchResult{
		Title:   title,
		Url:     normalized,
		Snippet: extractSnippet(sel),
	}, true
}

func extractTitleAndLink(sel *goquery.Selection) (string, string, bool) {
	for _, selector := range []string{".result__title a", "a.result__a", "h2 a", "a[href]"} {
		a := sel.Find(selector).First()
		if a.Length() == 0 {
			continue
		}

		title := strings.TrimSpace(a.Text())
		href, ok := a.Attr("href")
		href = strings.TrimSpace(href)
		if title == "" || !ok || href == "" {
			continue
		}

		return title, href, true
	}

	return "", "", false
}

func extractSnippet(sel *goquery.Selection) string {
	for _, selector := range []string{".result__snippet", ".result__body", ".snippet", "p"} {
		snippet := strings.TrimSpace(sel.Find(selector).First().Text())
		if snippet != "" {
			return snippet
		}
	}

	return ""
}

func normalizeResultURL(rawLink string) (string, bool) {
	rawLink = strings.TrimSpace(rawLink)
	if rawLink == "" {
		return "", false
	}

	if strings.HasPrefix(rawLink, "//") {
		rawLink = "https:" + rawLink
	}

	parsed, err := url.Parse(rawLink)
	if err != nil {
		return "", false
	}

	host := strings.ToLower(parsed.Hostname())
	if strings.HasSuffix(host, "duckduckgo.com") && parsed.Path == "/l/" {
		uddg := parsed.Query().Get("uddg")
		if uddg == "" {
			return "", false
		}

		decoded, err := url.QueryUnescape(uddg)
		if err != nil {
			return "", false
		}

		decoded = strings.TrimSpace(decoded)
		if decoded == "" {
			return "", false
		}

		return decoded, true
	}

	return rawLink, true
}

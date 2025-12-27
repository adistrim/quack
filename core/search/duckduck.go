package search

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"

	"quack/models"
)

const (
	ddgURL     = "https://html.duckduckgo.com/html"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36"
)

type DuckDuckGo struct {
	client  *http.Client
}

func NewDuckDuckGo() *DuckDuckGo {
	return &DuckDuckGo{
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (s *DuckDuckGo) Search(query string, maxResults int) ([]models.SearchResult, error) {
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

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	results := make([]models.SearchResult, 0, maxResults)

	doc.Find(".result").EachWithBreak(func(_ int, sel *goquery.Selection) bool {
		a := sel.Find(".result__title a")
		if a.Length() == 0 {
			return true
		}

		title := strings.TrimSpace(a.Text())
		link, _ := a.Attr("href")

		if strings.HasPrefix(link, "//duckduckgo.com/l/?uddg=") {
			if u, err := url.QueryUnescape(strings.Split(strings.Split(link, "uddg=")[1], "&")[0]); err == nil {
				link = u
			}
		}

		snippet := strings.TrimSpace(sel.Find(".result__snippet").Text())

		results = append(results, models.SearchResult{
			Title:    title,
			Url:      link,
			Snippet:  snippet,
			Rank: len(results) + 1,
		})

		return len(results) < maxResults
	})

	return results, nil
}

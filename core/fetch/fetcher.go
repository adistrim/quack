package fetch

import (
	"errors"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"

	"quack/util"
)

const maxTextBytes = 8000

type Fetcher struct {
	client  *http.Client
}

func New() *Fetcher {
	return &Fetcher{
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

func (f *Fetcher) Fetch(url string) (string, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := f.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to fetch page")
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	doc.Find("script, style, nav, header, footer").Remove()

	text := util.CleanWhitespace(doc.Text())
	if len(text) > maxTextBytes {
		text = text[:maxTextBytes] + "... [content truncated]"
	}

	return text, nil
}

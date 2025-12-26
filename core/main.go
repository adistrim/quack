package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"quack/fetch"
	"quack/models"
	"quack/search"
)

func main() {
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fail(err)
	}

	var req models.SearchRequest
	if err := json.Unmarshal(input, &req); err != nil {
		fail(err)
	}

	searcher := search.NewDuckDuckGo()
	fetcher := fetch.New()

	var resp models.SearchResponse

	switch req.Action {
	case "search":
		if req.Query == "" {
			fail(errors.New("query required"))
		}
		if req.MaxResults == 0 {
			req.MaxResults = 10
		}
		results, err := searcher.Search(req.Query, req.MaxResults)
		if err != nil {
			fail(err)
		}
		resp.Text = formatResults(results)

	case "fetch":
		if req.URL == "" {
			fail(errors.New("url required"))
		}
		text, err := fetcher.Fetch(req.URL)
		if err != nil {
			fail(err)
		}
		resp.Text = text

	default:
		fail(errors.New("invalid action"))
	}

	out, _ := json.Marshal(resp)
	os.Stdout.Write(out)
}

func formatResults(results []models.SearchResult) string {
	if len(results) == 0 {
		return "No results were found for your search query. Try rephrasing or retry later."
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Found %d search results:\n\n", len(results))

	for _, r := range results {
		fmt.Fprintf(&b, "%d. %s\n", r.Position, r.Title)
		fmt.Fprintf(&b, "   URL: %s\n", r.Link)
		fmt.Fprintf(&b, "   Summary: %s\n\n", r.Snippet)
	}

	return b.String()
}

func fail(err error) {
	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}

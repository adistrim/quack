package main

import (
	"encoding/json"
	"io"
	"os"

	"quack/fetch"
	"quack/models"
	"quack/search"
	"quack/util"
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
			resp.Error = "query required"
			break
		}
		if req.MaxResults == 0 {
			req.MaxResults = 10
		}
		results, err := searcher.Search(req.Query, req.MaxResults)
		if err != nil {
			resp.Error = err.Error()
			break
		}

		resp.Results = results

	case "fetch":
		if req.URL == "" {
			resp.Error = "url required"
			break
		}
		text, err := fetcher.Fetch(req.URL)
		if err != nil {
			resp.Fetch = &models.FetchResult{
				Success: false,
				Reason:  util.ClassifyError(err),
			}
			break
		}

		resp.Fetch = &models.FetchResult{
			Text:    text,
			Success: true,
		}

	default:
		resp.Error = "invalid action"
	}

	out, _ := json.Marshal(resp)
	os.Stdout.Write(out)
}

func fail(err error) {
	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}

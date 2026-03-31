package main

import (
	"encoding/json"
	"errors"
	"io"
	"os"

	"quack/fetch"
	"quack/models"
	"quack/search"
	"quack/util"
)

const maxStdinBytes = 1 * 1024 * 1024

func main() {
	input, err := io.ReadAll(io.LimitReader(os.Stdin, maxStdinBytes+1))
	if err != nil {
		fail(err)
	}

	if len(input) > maxStdinBytes {
		writeJSON(models.SearchResponse{Error: "input too large"})
		return
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
		req.MaxResults = search.NormalizeMaxResults(req.MaxResults)

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

	writeJSON(resp)
}

func fail(err error) {
	if errors.Is(err, io.EOF) {
		writeJSON(models.SearchResponse{Error: "invalid input"})
		return
	}

	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}

func writeJSON(resp models.SearchResponse) {
	out, _ := json.Marshal(resp)
	_, _ = os.Stdout.Write(out)
}

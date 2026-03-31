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

const maxStdinBytes = 1 * 1024 * 1024

func main() {
	input, err := io.ReadAll(io.LimitReader(os.Stdin, maxStdinBytes+1))
	if err != nil {
		fail(err)
	}

	resp, err := processRequest(input)
	if err != nil {
		fail(err)
		return
	}

	writeJSON(resp)
}

func processRequest(input []byte) (models.SearchResponse, error) {
	if len(input) > maxStdinBytes {
		return models.SearchResponse{Error: "input too large"}, nil
	}

	var req models.SearchRequest
	if err := json.Unmarshal(input, &req); err != nil {
		return models.SearchResponse{Error: "invalid input"}, nil
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

		fetchResult, err := fetcher.Fetch(req.URL)
		if err != nil {
			resp.Fetch = &models.FetchResult{
				Success: false,
				Reason:  util.ClassifyError(err),
			}
			break
		}

		resp.Fetch = &models.FetchResult{
			Text:      fetchResult.Text,
			Success:   true,
			Truncated: fetchResult.Truncated,
		}

	default:
		resp.Error = "invalid action"
	}

	return resp, nil
}

func fail(err error) {
	os.Stderr.WriteString(err.Error())
	os.Exit(1)
}

func writeJSON(resp models.SearchResponse) {
	out, _ := json.Marshal(resp)
	_, _ = os.Stdout.Write(out)
}

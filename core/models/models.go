
//
// ---------- Models ----------
//

package models

type SearchRequest struct {
	Action     string `json:"action"` // "search" | "fetch"
	Query      string `json:"query,omitempty"`
	URL        string `json:"url,omitempty"`
	MaxResults int    `json:"maxResults,omitempty"`
}

type SearchResult struct {
	Title    string `json:"title"`
	Url      string `json:"url"`
	Snippet  string `json:"snippet"`
	Rank     int    `json:"rank"`
}

type FetchResult struct {
	Text      string `json:"text,omitempty"`
	Success   bool   `json:"success"`
	Reason    string `json:"reason,omitempty"` // "blocked" | "timeout" | "http_error"
	Truncated bool   `json:"truncated,omitempty"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results,omitempty"`
	Fetch   *FetchResult   `json:"fetch,omitempty"`
	Error   string         `json:"error,omitempty"`
}

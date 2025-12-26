
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
	Link     string `json:"link"`
	Snippet  string `json:"snippet"`
	Position int    `json:"position"`
}

type SearchResponse struct {
	Results []SearchResult `json:"results,omitempty"`
	Text    string         `json:"text,omitempty"`
	Error   string         `json:"error,omitempty"`
}

package web

import "context"

// SearchProvider representation.
type SearchProvider interface {
	// Search performs a search.
	Search(ctx context.Context, query string, count int) ([]SearchResult, error)
}

// SearchResult represents a search result.
type SearchResult struct {
	// Title of the search result.
	Title string
	// URL of the search result.
	URL string
	// Description of the search result.
	Description string
}

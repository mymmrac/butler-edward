package web

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
)

// SearchTool represents a tool that searches the web.
type SearchTool struct {
	searchProvider SearchProvider
}

// NewSearchTool creates a new SearchTool.
func NewSearchTool(searchProvider SearchProvider) *SearchTool {
	return &SearchTool{searchProvider: searchProvider}
}

// Definition returns tool definition.
func (t *SearchTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name: "web_search",
			Description: "Search the web based on a query for current information. " +
				"Supports query and optional count parameters. " +
				"Returns a list of search results in the format `{number} {title} [{url}] {description}`.",
			//nolint:goconst
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query.",
					},
					"count": map[string]any{
						"type":        "integer",
						"description": "Max number of results.",
						"minimum":     minCount,
						"maximum":     maxCount,
					},
				},
				"required": []string{"query"},
			},
		},
	}
}

const (
	minCount = 1
	maxCount = 10
)

// Call searches the web based on the query.
func (t *SearchTool) Call(ctx context.Context, args json.RawMessage) (string, error) {
	var in struct {
		Query string `json:"query"`
		Count int    `json:"count"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	in.Count = min(max(in.Count, minCount), maxCount)

	results, err := t.searchProvider.Search(ctx, in.Query, in.Count)
	if err != nil {
		return "", fmt.Errorf("search: %w", err)
	}

	if len(results) == 0 {
		return "No search results were found for: " + strconv.Quote(in.Query), nil
	}

	sb := &strings.Builder{}
	_, _ = fmt.Fprintf(sb, "Search results for: %q\n", in.Query)
	for i, result := range results {
		_, _ = fmt.Fprintf(sb, "%d %s [%s] %s\n", i+1, result.Title, result.URL, result.Description)
	}
	return sb.String(), nil
}

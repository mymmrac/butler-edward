package duckduckgo

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool/web"
)

// Provider implements your interface using DuckDuckGo.
type Provider struct {
	client *http.Client
}

// NewProvider creates a new DuckDuckGo search provider.
func NewProvider() *Provider {
	return &Provider{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Search performs a search using DuckDuckGo HTML results.
func (p *Provider) Search(ctx context.Context, query string, count int) ([]web.SearchResult, error) {
	const endpoint = "https://html.duckduckgo.com/html/"

	params := url.Values{}
	params.Set("q", query)

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	request.Header.Set("User-Agent", "Mozilla/5.0 (compatible; ButlerEdwardAgent/1.0)")

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status: %d", response.StatusCode)
	}

	doc, err := goquery.NewDocumentFromReader(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	results := make([]web.SearchResult, 0, count)
	doc.Find(".result").EachWithBreak(func(_ int, s *goquery.Selection) bool {
		if len(results) >= count {
			return false
		}

		title := s.Find(".result__title a").Text()
		link, _ := s.Find(".result__title a").Attr("href")
		desc := s.Find(".result__snippet").Text()

		if title == "" || link == "" {
			return true
		}

		results = append(results, web.SearchResult{
			Title:       title,
			URL:         unwrapURL(link),
			Description: desc,
		})
		return true
	})

	return results, nil
}

func unwrapURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	q := u.Query().Get("uddg")
	if q == "" {
		return raw
	}

	decoded, err := url.QueryUnescape(q)
	if err != nil {
		return raw
	}

	return decoded
}

package web

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"mime"
	"net/http"
	"net/url"
	"time"

	"github.com/PuerkitoBio/goquery"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/module/dns"
)

// FetchTool represents a tool that fetches the contents of a URL using HTTP(S) GET.
type FetchTool struct {
	client *http.Client
}

// NewFetchTool creates a new FetchTool.
func NewFetchTool() *FetchTool {
	return &FetchTool{client: &http.Client{
		Timeout: 10 * time.Second,
	}}
}

// Definition returns tool definition.
func (t *FetchTool) Definition() provider.ToolDefinition {
	return provider.ToolDefinition{
		Type: provider.ToolTypeFunction,
		Function: &provider.ToolFunction{
			Name: "web_fetch",
			Description: "Fetches the contents of a URL using HTTP(S) GET. " +
				"Returns only readable text. " +
				"In case of JSON content returns it as is.",
			//nolint:goconst
			Parameters: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "URL to fetch",
					},
				},
				"required": []string{"url"},
			},
		},
	}
}

const fetchLimit = 10 * 1024 * 1024 // 10MB

// Call fetches the contents of a URL using HTTP(S) GET.
func (t *FetchTool) Call(ctx context.Context, args json.RawMessage) (string, error) {
	var in struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return "", fmt.Errorf("invalid args: %w", err)
	}

	u, err := url.Parse(in.URL)
	if err != nil {
		return "", fmt.Errorf("parse url: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return "", fmt.Errorf("invalid url scheme: %q", u.Scheme)
	}
	if u.Host == "" {
		return "", fmt.Errorf("missing url host")
	}

	hostname := u.Hostname()
	if dns.ClassifyHost(ctx, hostname) != dns.HostPublic {
		return "", fmt.Errorf("host %q is not a public hostname", hostname)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return "", fmt.Errorf("new request: %w", err)
	}

	request.Header.Set("User-Agent", UserAgent)

	response, err := t.client.Do(request)
	if err != nil {
		return "", fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("status: %d", response.StatusCode)
	}

	contentType := response.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", fmt.Errorf("parse media type: %w", err)
	}

	response.Body = http.MaxBytesReader(nil, response.Body, fetchLimit)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	header := fmt.Sprintf("URL: %s\nContent-Type: %s\n", u.String(), contentType)
	if mediaType == "text/html" {
		var doc *goquery.Document
		doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			return "", fmt.Errorf("parse html: %w", err)
		}

		doc.Find("script, style, noscript").Remove()
		text := doc.Find("body").Text()
		return header + "\n" + html.UnescapeString(text), nil
	}

	return header + "\n" + string(body), nil
}

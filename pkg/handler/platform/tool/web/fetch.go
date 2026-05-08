package web

import (
	"bytes"
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
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
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
func (t *FetchTool) Call(ctx *tool.Context, args json.RawMessage) (*tool.Result, error) {
	var in struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal(args, &in); err != nil {
		return tool.ErrorResult("Invalid arguments", fmt.Errorf("invalid args: %w", err))
	}

	u, err := url.Parse(in.URL)
	if err != nil {
		return tool.ErrorResult("Invalid URL", fmt.Errorf("parse url: %w", err))
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return tool.ErrorResult("Invalid URL scheme", fmt.Errorf("invalid url scheme: %q", u.Scheme))
	}
	if u.Host == "" {
		return tool.ErrorResult("Missing URL host", fmt.Errorf("missing url host"))
	}

	hostname := u.Hostname()
	if dns.ClassifyHost(ctx, hostname) != dns.HostPublic {
		return tool.ErrorResult(
			"Host is not a public hostname",
			fmt.Errorf("host %q is not a public hostname", hostname),
		)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return tool.ErrorResult("Failed to create request", fmt.Errorf("new request: %w", err))
	}

	request.Header.Set("User-Agent", UserAgent)

	response, err := t.client.Do(request)
	if err != nil {
		return tool.ErrorResult("Failed to fetch URL", fmt.Errorf("do request: %w", err))
	}
	defer func() { _ = response.Body.Close() }()

	if response.StatusCode != http.StatusOK {
		return tool.ErrorResult("Failed to fetch URL", fmt.Errorf("status: %d", response.StatusCode))
	}

	contentType := response.Header.Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return tool.ErrorResult("Failed to parse media type", fmt.Errorf("parse media type: %w", err))
	}

	response.Body = http.MaxBytesReader(nil, response.Body, fetchLimit)
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return tool.ErrorResult("Failed to read response", fmt.Errorf("read response: %w", err))
	}

	header := fmt.Sprintf("URL: %s\nContent-Type: %s\n", u.String(), contentType)
	if mediaType == "text/html" {
		var doc *goquery.Document
		doc, err = goquery.NewDocumentFromReader(bytes.NewReader(body))
		if err != nil {
			return tool.ErrorResult("Failed to parse HTML", fmt.Errorf("parse html: %w", err))
		}

		doc.Find("script, style, noscript").Remove()
		text := doc.Find("body").Text()
		return tool.SuccessResult("HTML content fetched", header+"\n"+html.UnescapeString(text))
	}

	return tool.SuccessResult("Content fetched", header+"\n"+string(body))
}

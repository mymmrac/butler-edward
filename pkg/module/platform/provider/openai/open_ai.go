package openai

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sync"

	"github.com/mymmrac/butler-edward/pkg/module/collection"
	"github.com/mymmrac/butler-edward/pkg/module/logger"
	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/module/ternary"
)

// OpenAI represents OpenAI compatible provider.
type OpenAI struct {
	client    *http.Client
	name      string
	baseURL   url.URL
	chatAPI   string
	modelsAPI string
	apiKey    string

	modelsLock *sync.RWMutex
	models     []provider.Model
}

// NewOpenAI creates a new OpenAI compatible provider.
func NewOpenAI(name, baseURL, chatAPI, modelsAPI, apiKey string, models []provider.Model) (*OpenAI, error) {
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	if chatAPI == "" {
		chatAPI = "/openai/v1/chat/completions"
	}
	if modelsAPI == "" {
		modelsAPI = "/openai/v1/models"
	}

	return &OpenAI{
		client:     http.DefaultClient,
		name:       name,
		baseURL:    *parsedURL,
		chatAPI:    chatAPI,
		modelsAPI:  modelsAPI,
		apiKey:     apiKey,
		modelsLock: &sync.RWMutex{},
		models:     models,
	}, nil
}

// Name returns provider name.
func (o *OpenAI) Name() string {
	return o.name
}

// Models returns list of available models.
func (o *OpenAI) Models(ctx context.Context) ([]provider.Model, error) {
	o.modelsLock.RLock()
	if len(o.models) != 0 {
		o.modelsLock.RUnlock()
		return o.models, nil
	}
	o.modelsLock.RUnlock()

	o.modelsLock.Lock()
	defer o.modelsLock.Unlock()

	if len(o.models) != 0 {
		return o.models, nil
	}

	var response modelsResponse
	if err := o.call(ctx, http.MethodGet, o.modelsAPI, nil, &response); err != nil {
		return nil, fmt.Errorf("get models: %w", err)
	}

	models := make([]provider.Model, len(response.Data))
	for i, m := range response.Data {
		models[i] = provider.Model{
			Name: m.ID,
		}
	}

	slices.SortFunc(o.models, func(a, b provider.Model) int {
		return cmp.Compare(a.Name, b.Name)
	})
	o.models = models
	return models, nil
}

//revive:disable:nested-structs
type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// Chat sends messages to the LLM provider.
func (o *OpenAI) Chat(
	ctx context.Context, model string, messages []provider.Message, tools []provider.ToolDefinition,
) (*provider.Response, error) {
	request := &chatRequest{
		Model:       model,
		Messages:    collection.MakeSlice[chatMessage](len(messages)),
		Tools:       collection.MakeSlice[chatTool](len(tools)),
		ToolChoice:  ternary.If(len(tools) > 0, "auto", ""),
		Temperature: 0.1,
	}

	for _, msg := range messages {
		toolCalls := collection.MakeSlice[chatToolCall](len(msg.ToolCalls))
		for _, call := range msg.ToolCalls {
			// TODO: Add type validation
			toolCalls = append(toolCalls, chatToolCall{
				ID:   call.ID,
				Type: string(call.Type),
				Function: &chatToolFunctionCall{
					Name:      call.Function.Name,
					Arguments: string(call.Function.Arguments),
				},
			})
		}

		request.Messages = append(request.Messages, chatMessage{
			Role:       string(msg.Role),
			Name:       msg.Name,
			Content:    msg.Content,
			ToolCalls:  toolCalls,
			ToolCallID: msg.ToolCallID,
		})
	}

	for _, tool := range tools {
		// TODO: Add type validation
		request.Tools = append(request.Tools, chatTool{
			Type: string(tool.Type),
			Function: &chatToolFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  tool.Function.Parameters,
			},
		})
	}

	var response chatResponse
	if err := o.call(ctx, http.MethodPost, o.chatAPI, request, &response); err != nil {
		return nil, fmt.Errorf("chat: %w", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no choices")
	}

	choice := response.Choices[0]
	toolCalls := make([]provider.ToolCall, 0, len(choice.Message.ToolCalls))
	for _, call := range choice.Message.ToolCalls {
		// TODO: Add type validation
		toolCalls = append(toolCalls, provider.ToolCall{
			ID:   call.ID,
			Type: provider.ToolType(call.Type),
			Function: &provider.ToolFunctionCall{
				Name:      call.Function.Name,
				Arguments: json.RawMessage(call.Function.Arguments),
			},
		})
	}

	return &provider.Response{
		Content:   choice.Message.Content,
		ToolCalls: toolCalls,
	}, nil
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	Tools       []chatTool    `json:"tools,omitempty"`
	ToolChoice  string        `json:"tool_choice,omitempty"`
	Temperature float64       `json:"temperature"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

type chatTool struct {
	Type     string            `json:"type"`
	Function *chatToolFunction `json:"function,omitempty"`
}

type chatToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

type chatMessage struct {
	Role       string         `json:"role"`
	Name       string         `json:"name,omitempty"`
	Content    string         `json:"content"`
	ToolCalls  []chatToolCall `json:"tool_calls,omitempty"`
	ToolCallID string         `json:"tool_call_id,omitempty"`
}

type chatToolCall struct {
	ID       string                `json:"id"`
	Type     string                `json:"type"`
	Function *chatToolFunctionCall `json:"function,omitempty"`
}

type chatToolFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

func (o *OpenAI) call(ctx context.Context, method, path string, body any, result any) error {
	var bodyData []byte
	var bodyReader io.Reader
	if body != nil {
		var err error
		bodyData, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyData)
	}

	request, err := http.NewRequestWithContext(ctx, method, o.baseURL.JoinPath(path).String(), bodyReader)
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+o.apiKey)
	request.Header.Set("Content-Type", "application/json")

	logger.Debugw(ctx, "http request", "method", request.Method, "url", request.URL, "body", string(bodyData))
	response, err := o.client.Do(request)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	logger.Debugw(ctx, "http response", "status", response.Status, "body", string(data))

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", response.StatusCode)
	}

	if err = json.Unmarshal(data, result); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}

	return nil
}

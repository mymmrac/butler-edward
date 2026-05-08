package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/channel"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/session"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/storage"
	"github.com/mymmrac/butler-edward/pkg/handler/platform/tool"
	"github.com/mymmrac/butler-edward/pkg/module/logger"
)

const maxIterations = 5

// Agent representation.
type Agent struct {
	channels        []channel.Channel
	providers       []provider.Provider
	tools           map[string]tool.Tool
	toolDefinitions []provider.ToolDefinition

	sessionManager session.Manager
	storage        storage.Storage

	// TODO: Let user select provider and model.
	selectedProvider provider.Provider
	selectedModel    *provider.Model
}

// NewAgent creates new agent.
func NewAgent(
	channels []channel.Channel, providers []provider.Provider, tools []tool.Tool, sessionManager session.Manager,
	storage storage.Storage,
) (*Agent, error) {
	if len(channels) == 0 {
		return nil, fmt.Errorf("no channels configured")
	}
	if len(providers) == 0 {
		return nil, fmt.Errorf("no providers configured")
	}

	toolbox := make(map[string]tool.Tool, len(tools))
	toolDefinitions := make([]provider.ToolDefinition, 0, len(tools))
	for _, t := range tools {
		def := t.Definition()

		var toolName string
		switch def.Type {
		case provider.ToolTypeFunction:
			toolName = def.Function.Name
		default:
			return nil, fmt.Errorf("unsupported tool type: %q", def.Type)
		}

		if _, ok := toolbox[toolName]; ok {
			return nil, fmt.Errorf("tool with name %q already exists", toolName)
		}

		toolbox[toolName] = t
		toolDefinitions = append(toolDefinitions, def)
	}

	return &Agent{
		channels:        channels,
		providers:       providers,
		tools:           toolbox,
		toolDefinitions: toolDefinitions,

		sessionManager: sessionManager,
		storage:        storage,
	}, nil
}

// SelectProviderAndModel selects provider and model.
func (a *Agent) SelectProviderAndModel(ctx context.Context, provider, model string) error {
	if a.selectedProvider != nil && a.selectedModel != nil {
		return nil
	}

	for _, prv := range a.providers {
		if prv.Name() == provider {
			a.selectedProvider = prv
		}
	}
	if a.selectedProvider == nil {
		return fmt.Errorf("provider %q not found", provider)
	}

	models, err := a.selectedProvider.Models(ctx)
	if err != nil {
		return fmt.Errorf("get models: %w", err)
	}
	logger.Debugw(ctx, "provider models", "provider", provider, "models", models)

	for _, m := range models {
		if m.Name == model {
			a.selectedModel = &m
			return nil
		}
	}

	return fmt.Errorf("model %q not found", model)
}

// Run agent.
func (a *Agent) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	log := logger.FromContext(ctx)

	runningChannels := make([]channel.Channel, 0, len(a.channels))
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second*10)
		defer stopCancel()

		for _, ch := range runningChannels {
			log.Infow("stopping channel", "name", ch.Name())
			if err := ch.Stop(stopCtx); err != nil {
				log.Errorw("stop channel", "name", ch.Name(), "error", err)
			}
		}
	}()

	for _, ch := range a.channels {
		log.Infow("starting channel", "name", ch.Name())

		messages, err := ch.Start(ctx)
		if err != nil {
			return fmt.Errorf("start channel %s: %w", ch.Name(), err)
		}

		runningChannels = append(runningChannels, ch)
		go a.handleMessages(ctx, ch, messages)
	}

	<-ctx.Done()
	return nil
}

func (a *Agent) handleMessages(ctx context.Context, ch channel.Channel, messages <-chan channel.Message) {
	var ok bool
	var msg channel.Message
	log := logger.FromContext(ctx).With("channel", ch.Name())
	for {
		select {
		case msg, ok = <-messages:
			if !ok {
				return
			}
		case <-ctx.Done():
			return
		}

		if err := a.handleMessage(ctx, ch, msg); err != nil {
			log.Errorw("handle message", "error", err)
		}
	}
}

func (a *Agent) handleMessage(ctx context.Context, ch channel.Channel, msg channel.Message) error {
	chatSession, isNew, err := a.sessionManager.Session(ctx, ch.Name()+":"+msg.ChatID)
	if err != nil {
		return fmt.Errorf("get session: %w", err)
	}

	lc := &loopContext{
		ch:          ch,
		chatSession: chatSession,
		userID:      msg.UserID,
		chatID:      msg.ChatID,
	}

	if isNew {
		var builtSystemPrompt string
		builtSystemPrompt, err = a.buildSystemPrompt(ctx, lc)
		if err != nil {
			return fmt.Errorf("build system prompt: %w", err)
		}

		err = chatSession.AddMessage(ctx, provider.Message{
			Role:    provider.MessageRoleSystem,
			Content: a.normalizeContent(builtSystemPrompt),
		})
		if err != nil {
			return fmt.Errorf("add system message to session: %w", err)
		}

		if sc, ok := ch.(channel.SessionNameCapable); ok {
			if ok, err = sc.CanSetSessionName(ctx, msg.ChatID); err != nil {
				return fmt.Errorf("check if session name can be set: %w", err)
			} else if ok {
				go a.setSessionName(ctx, sc, msg)
			}
		}
	}

	err = chatSession.AddMessage(ctx, provider.Message{
		Role:    provider.MessageRoleUser,
		Name:    msg.UserID,
		Content: msg.Text,
	})
	if err != nil {
		return fmt.Errorf("add user message to session: %w", err)
	}

	err = a.runAgentLoop(ctx, lc)
	if err != nil {
		return fmt.Errorf("run agent loop: %w", err)
	}

	return nil
}

type loopContext struct {
	ch          channel.Channel
	chatSession session.Session
	userID      string
	chatID      string
}

//nolint:gocognit,funlen
func (a *Agent) runAgentLoop(ctx context.Context, lc *loopContext) (err_ error) { //revive:disable:var-naming
	defer func() {
		if err_ == nil {
			return
		}

		sendErr := lc.ch.Send(ctx, channel.Message{
			ChatID: lc.chatID,
			Text:   "Failed to respond, try again",
		})
		if sendErr != nil {
			logger.Warnw(ctx, "send error message", "error", sendErr)
		}
	}()
	for range maxIterations {
		prv := a.selectedProvider
		model := a.selectedModel

		history, err := lc.chatSession.History(ctx)
		if err != nil {
			return fmt.Errorf("get session history: %w", err)
		}

		var placeholderMessageID string
		if pc, ok := lc.ch.(channel.PlaceholderCapable); ok {
			if placeholderMessageID, err = pc.SendPlaceholder(ctx, lc.chatID); err != nil {
				logger.Warnw(ctx, "send placeholder", "error", err)
			}
		}

		var stopTyping func()
		if tc, ok := lc.ch.(channel.TypingCapable); ok {
			if stopTyping, err = tc.StartTyping(ctx, lc.chatID); err != nil {
				logger.Warnw(ctx, "start typing", "error", err)
			}
		}

		response, err := prv.Chat(ctx, model.Name, history, a.toolDefinitions)
		if stopTyping != nil {
			stopTyping()
		}
		if err != nil {
			return fmt.Errorf("chat: %w", err)
		}

		response.Content = a.normalizeContent(response.Content)

		if response.Content != "" {
			err = lc.ch.Send(ctx, channel.Message{
				ChatID:               lc.chatID,
				PlaceholderMessageID: placeholderMessageID,
				Text:                 response.Content,
			})
			if err != nil {
				return fmt.Errorf("send response content message: %w", err)
			}
			placeholderMessageID = ""
		}

		err = lc.chatSession.AddMessage(ctx, provider.Message{
			Role:      provider.MessageRoleAssistant,
			Content:   response.Content,
			ToolCalls: response.ToolCalls,
		})
		if err != nil {
			return fmt.Errorf("add assistant message to session: %w", err)
		}

		if len(response.ToolCalls) == 0 {
			return nil
		}

		for _, call := range response.ToolCalls {
			var toolName, toolCall string
			switch call.Type {
			case provider.ToolTypeFunction:
				toolName = call.Function.Name
				toolCall = fmt.Sprintf("Call tool %s with arguments: %s",
					call.Function.Name, string(call.Function.Arguments),
				)
			default:
				logger.Warnw(ctx, "unsupported tool type", "call-id", call.ID, "type", call.Type)
				toolCall = "Unsupported tool type"
			}

			err = lc.ch.Send(ctx, channel.Message{
				ChatID:               lc.chatID,
				PlaceholderMessageID: placeholderMessageID,
				Text:                 toolCall,
			})
			if err != nil {
				return fmt.Errorf("send response tool call message: %w", err)
			}
			placeholderMessageID = ""

			result := a.callTool(ctx, lc, call)

			err = lc.chatSession.AddMessage(ctx, provider.Message{
				Role:       provider.MessageRoleTool,
				Name:       toolName,
				Content:    result.Result,
				ToolCallID: call.ID,
			})
			if err != nil {
				return fmt.Errorf("add tool call message to session: %w", err)
			}

			err = lc.ch.Send(ctx, channel.Message{
				ChatID: lc.chatID,
				Text:   result.HumanReadableResult,
			})
			if err != nil {
				return fmt.Errorf("send response tool call result message: %w", err)
			}
		}
	}
	return nil
}

func (a *Agent) callTool(ctx context.Context, lc *loopContext, call provider.ToolCall) *tool.Result {
	switch call.Type {
	case provider.ToolTypeFunction:
		t, ok := a.tools[call.Function.Name]
		if !ok {
			return &tool.Result{
				Result:              fmt.Sprintf("tool with name %q not found", call.Function.Name),
				HumanReadableResult: "Tool not found",
			}
		}

		tc := &tool.Context{
			Context: ctx,
			UserID:  lc.userID,
			ChatID:  lc.chatID,
		}

		result, err := t.Call(tc, call.Function.Arguments)
		if err != nil {
			if result == nil {
				return &tool.Result{
					Result:              "Error: " + err.Error(),
					HumanReadableResult: "Failed to call tool",
				}
			}

			result.Result = "Error: " + err.Error()
			if result.HumanReadableResult == "" {
				result.HumanReadableResult = "Failed to call tool"
			}

			return result
		}

		if result.HumanReadableResult == "" {
			result.HumanReadableResult = result.Result
		}
		return result
	default:
		logger.Warnw(ctx, "unsupported tool type", "call-id", call.ID, "type", call.Type)
		return &tool.Result{
			Result:              fmt.Sprintf("unsupported tool type: %q", call.Type),
			HumanReadableResult: "Unsupported tool type",
		}
	}
}

func (a *Agent) setSessionName(ctx context.Context, sc channel.SessionNameCapable, msg channel.Message) {
	log := logger.FromContext(ctx).With("chat-id", msg.ChatID)

	response, err := a.selectedProvider.Chat(ctx, a.selectedModel.Name, []provider.Message{
		{
			Role:    provider.MessageRoleSystem,
			Content: a.normalizeContent(sessionNameSystemPrompt),
		},
		{
			Role:    provider.MessageRoleUser,
			Name:    msg.UserID,
			Content: msg.Text,
		},
	}, nil)
	if err != nil {
		log.Warnw("chat: session name", "error", err)
		return
	}

	response.Content = a.normalizeContent(response.Content)
	if response.Content == "" {
		return
	}

	if len(response.Content) > maxSessionNameLength {
		response.Content = response.Content[:maxSessionNameLength-3] + "..."
	}

	if err = sc.SetSessionName(ctx, msg.ChatID, response.Content); err != nil {
		log.Warnw("set a session name", "error", err)
	}
}

func (a *Agent) normalizeContent(content string) string {
	return strings.TrimSpace(content)
}

package agent

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mymmrac/butler-edward/pkg/module/logger"
	"github.com/mymmrac/butler-edward/pkg/module/platform/channel"
	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/module/platform/session"
	"github.com/mymmrac/butler-edward/pkg/module/platform/tool"
)

const maxIterations = 5

// Agent representation.
type Agent struct {
	channels        []channel.Channel
	providers       []provider.Provider
	tools           map[string]tool.Tool
	toolDefinitions []provider.ToolDefinition
	sessionManager  session.Manager

	// TODO: Let user select provider and model.
	selectedProvider provider.Provider
	selectedModel    *provider.Model
}

// NewAgent creates new agent.
func NewAgent(
	channels []channel.Channel, providers []provider.Provider, tools []tool.Tool, sessionManager session.Manager,
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
		sessionManager:  sessionManager,
	}, nil
}

// SelectProviderAndModel selects provider and model.
// TODO: Let user select provider and model, this function is just for testing.
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

	if isNew {
		err = chatSession.AddMessage(ctx, provider.Message{
			Role:    provider.MessageRoleSystem,
			Content: strings.TrimSpace(systemPrompt),
		})
		if err != nil {
			return fmt.Errorf("add system message to session: %w", err)
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

	err = a.runAgentLoop(ctx, &loopContext{
		ch:          ch,
		chatSession: chatSession,
		chatID:      msg.ChatID,
	})
	if err != nil {
		return fmt.Errorf("run agent loop: %w", err)
	}

	return nil
}

type loopContext struct {
	ch          channel.Channel
	chatSession session.Session
	chatID      string
}

//nolint:gocognit
func (a *Agent) runAgentLoop(ctx context.Context, lc *loopContext) error {
	for range maxIterations {
		prv := a.selectedProvider
		model := a.selectedModel

		history, err := lc.chatSession.History(ctx)
		if err != nil {
			return fmt.Errorf("get session history: %w", err)
		}

		response, err := prv.Chat(ctx, model.Name, history, a.toolDefinitions)
		if err != nil {
			return fmt.Errorf("chat: %w", err)
		}

		response.Content = strings.TrimSpace(response.Content)

		if response.Content != "" {
			err = lc.ch.Send(ctx, channel.Message{
				ChatID: lc.chatID,
				Text:   response.Content,
			})
			if err != nil {
				return fmt.Errorf("send response content message: %w", err)
			}
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
			var toolName string
			text := fmt.Sprintf("Tool call: %s\n", call.ID)
			switch call.Type {
			case provider.ToolTypeFunction:
				toolName = call.Function.Name
				text += fmt.Sprintf("\tFunction: %s\n", call.Function.Name)
				text += fmt.Sprintf("\tArguments: %s\n", call.Function.Arguments)
			default:
				text += fmt.Sprintf("\tUnsupported tool type: %q\n", call.Type)
			}

			err = lc.ch.Send(ctx, channel.Message{
				ChatID: lc.chatID,
				Text:   text,
			})
			if err != nil {
				return fmt.Errorf("send response tool call message: %w", err)
			}

			var result string
			result, err = a.callTool(ctx, call)
			if err != nil {
				result = "Error: " + err.Error()
			}

			err = lc.chatSession.AddMessage(ctx, provider.Message{
				Role:       provider.MessageRoleTool,
				Name:       toolName,
				Content:    result,
				ToolCallID: call.ID,
			})
			if err != nil {
				return fmt.Errorf("add tool call message to session: %w", err)
			}

			err = lc.ch.Send(ctx, channel.Message{
				ChatID: lc.chatID,
				Text:   "Tool call result:\n" + result,
			})
			if err != nil {
				return fmt.Errorf("send response tool call result message: %w", err)
			}
		}
	}
	return nil
}

func (a *Agent) callTool(ctx context.Context, call provider.ToolCall) (string, error) {
	switch call.Type {
	case provider.ToolTypeFunction:
		t, ok := a.tools[call.Function.Name]
		if !ok {
			return "", fmt.Errorf("tool with name %q not found", call.Function.Name)
		}
		return t.Call(ctx, call.Function.Arguments)
	default:
		return "", fmt.Errorf("unsupported tool type: %q", call.Type)
	}
}

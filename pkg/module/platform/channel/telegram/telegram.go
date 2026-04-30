package telegram

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"

	"github.com/mymmrac/butler-edward/pkg/module/logger"
	"github.com/mymmrac/butler-edward/pkg/module/platform/channel"
)

// Telegram represents a Telegram channel.
type Telegram struct {
	bot    *telego.Bot
	bh     *th.BotHandler
	cancel func()
}

// NewTelegram creates a new Telegram channel.
func NewTelegram(ctx context.Context, botToken string) (*Telegram, error) {
	bot, err := telego.NewBot(
		botToken,
		telego.WithLogger(logger.FromContext(ctx).WithOptions(logger.WithIncreasedLevel(logger.LevelInfo))),
		telego.WithHealthCheck(ctx),
	)
	if err != nil {
		return nil, fmt.Errorf("new bot: %w", err)
	}

	return &Telegram{
		bot: bot,
		bh:  nil,
	}, nil
}

// Name returns channel name.
func (t *Telegram) Name() string {
	return "telegram"
}

// Start starts channel.
func (t *Telegram) Start(ctx context.Context) (<-chan channel.Message, error) {
	ctx, cancelCtx := context.WithCancel(ctx)
	t.cancel = cancelCtx

	updates, err := t.bot.UpdatesViaLongPolling(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("get updates: %w", err)
	}

	t.bh, err = th.NewBotHandler(t.bot, updates)
	if err != nil {
		return nil, fmt.Errorf("new bot handler: %w", err)
	}

	t.bh.HandleMessage(t.startCommand, th.CommandEqual("start"))

	messages := make(chan channel.Message)
	t.cancel = sync.OnceFunc(func() {
		cancelCtx()
		close(messages)
	})

	t.bh.HandleMessage(func(ctx *th.Context, message telego.Message) error {
		select {
		case <-ctx.Done():
			// Ignored
		case messages <- channel.Message{
			ChatID: strconv.FormatInt(message.Chat.ID, 10),
			UserID: strconv.FormatInt(message.From.ID, 10),
			Text:   message.Text,
		}:
			// Sent
		}
		return nil
	})

	go func() {
		defer t.cancel()
		if err = t.bh.Start(); err != nil {
			logger.Errorw(ctx, "start bot handler", "error", err)
		}
	}()

	return messages, nil
}

// Stop stops channel.
func (t *Telegram) Stop(ctx context.Context) error {
	if t.cancel != nil {
		t.cancel()
	}
	if t.bh != nil {
		if err := t.bh.StopWithContext(ctx); err != nil {
			return fmt.Errorf("stop bot handler: %w", err)
		}
	}
	return nil
}

// Send sends messages to channel.
func (t *Telegram) Send(ctx context.Context, msg channel.Message) error {
	chatID, err := strconv.ParseInt(msg.ChatID, 10, 64)
	if err != nil {
		return fmt.Errorf("parse chat id: %w", err)
	}

	_, err = t.bot.SendMessage(ctx, tu.Message(tu.ID(chatID), msg.Text))
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}

	return nil
}

func (t *Telegram) startCommand(ctx *th.Context, message telego.Message) error {
	_, err := ctx.Bot().
		SendMessage(ctx, tu.Message(tu.ID(message.Chat.ID), "Hello, I'm Butler Edward! Your personal AI assistant."))
	if err != nil {
		return fmt.Errorf("send message: %w", err)
	}
	return nil
}

package terminal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/channel"
	"github.com/mymmrac/butler-edward/pkg/module/logger"
)

// Terminal represents a terminal channel.
type Terminal struct {
	done     chan struct{}
	messages chan channel.Message
	cancel   func()
	stdin    io.Reader
	stdout   io.Writer
}

// NewTerminal creates a new terminal channel.
func NewTerminal() *Terminal {
	return &Terminal{
		stdin:  os.Stdin,
		stdout: os.Stdout,
	}
}

// Name returns channel name.
func (t *Terminal) Name() string {
	return "terminal"
}

// Start starts channel.
func (t *Terminal) Start(ctx context.Context) (<-chan channel.Message, error) {
	t.done = make(chan struct{})
	t.messages = make(chan channel.Message)
	t.cancel = sync.OnceFunc(func() {
		close(t.done)
		close(t.messages)
	})
	go t.handleInput(ctx)
	return t.messages, nil
}

func (t *Terminal) handleInput(ctx context.Context) {
	defer t.cancel()

	_, err := fmt.Fprintln(t.stdout, "Terminal channel started, type your message.")
	if err != nil {
		logger.Errorw(ctx, "write terminal output", "error", err)
		return
	}

	scanner := bufio.NewScanner(t.stdin)
	for {
		select {
		case <-t.done:
			return
		default:
			// Continue
		}

		if !scanner.Scan() {
			if err = scanner.Err(); err != nil {
				logger.Errorw(ctx, "read terminal input", "error", scanner.Err())
			}
			return
		}

		select {
		case <-t.done:
			return
		case t.messages <- channel.Message{
			ChatID: "terminal",
			UserID: "user",
			Text:   scanner.Text(),
		}:
			// Continue
		}
	}
}

// Stop stops channel.
func (t *Terminal) Stop(_ context.Context) error {
	if t.cancel != nil {
		t.cancel()
	}
	return nil
}

// Send sends messages to channel.
func (t *Terminal) Send(_ context.Context, msg channel.Message) error {
	_, err := fmt.Fprintln(t.stdout, msg.Text)
	return err
}

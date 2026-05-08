package channel

import "context"

// Channel representation.
type Channel interface {
	// Name returns channel name.
	Name() string

	// Start starts channel. Returns chan with received channel messages.
	Start(ctx context.Context) (<-chan Message, error)
	// Stop stops channel.
	Stop(ctx context.Context) error

	// Send sends the message to the channel.
	Send(ctx context.Context, msg Message) error
}

// Message represents a message received from or sent to the channel.
type Message struct {
	// UserID is the unique identifier of the user who sent the message. When sending the message, this can be omitted.
	UserID string
	// ChatID is the unique identifier of the chat to which the message belongs.
	ChatID string
	// PlaceholderMessageID is the unique identifier of the placeholder message. This can be omitted.
	PlaceholderMessageID string
	// Text is the text of the message.
	Text string
}

// TypingCapable represents a channel that supports typing indicators.
type TypingCapable interface {
	// StartTyping starts typing indicator for the specified chat. Returns a function to stop the typing indicator.
	StartTyping(ctx context.Context, chatID string) (stop func(), err error)
}

// PlaceholderCapable represents a channel that supports placeholder messages.
type PlaceholderCapable interface {
	// SendPlaceholder sends a placeholder message to the specified chat. Returns the placeholder message ID.
	SendPlaceholder(ctx context.Context, chatID string) (messageID string, err error)
}

// SessionNameCapable represents a channel that supports session names.
type SessionNameCapable interface {
	// SetSessionName sets the session name for the specified chat.
	SetSessionName(ctx context.Context, chatID string, name string) error
}

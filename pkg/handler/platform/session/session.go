package session

import (
	"context"

	"github.com/mymmrac/butler-edward/pkg/handler/platform/provider"
)

// Manager represents chat session manager.
type Manager interface {
	// Session returns chat session by ID. Creates a new session if not exists and returns true new session was
	// created, false if the session already existed.
	Session(ctx context.Context, id string) (Session, bool, error)
}

// Session represents chat session.
type Session interface {
	// History returns chat history.
	History(ctx context.Context) ([]provider.Message, error)
	// AddMessage adds a message to the session.
	AddMessage(ctx context.Context, message provider.Message) error
}

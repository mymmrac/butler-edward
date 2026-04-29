package inmemory

import (
	"context"
	"sync"

	"github.com/mymmrac/butler-edward/pkg/module/platform/provider"
	"github.com/mymmrac/butler-edward/pkg/module/platform/session"
)

// InMemory represents an in-memory session manager.
type InMemory struct {
	lock     *sync.RWMutex
	sessions map[string]*Session
}

// NewInMemory creates a new in-memory session manager.
func NewInMemory() *InMemory {
	return &InMemory{
		lock:     &sync.RWMutex{},
		sessions: make(map[string]*Session),
	}
}

// Session returns a session by ID.
func (m *InMemory) Session(_ context.Context, id string) (session.Session, bool, error) {
	var isNew bool
	m.lock.RLock()
	s, ok := m.sessions[id]
	m.lock.RUnlock()
	if !ok {
		m.lock.Lock()
		defer m.lock.Unlock()

		s, ok = m.sessions[id]
		if !ok {
			s = &Session{
				lock:    &sync.RWMutex{},
				history: nil,
			}
			isNew = true
			m.sessions[id] = s
		}
	}
	return s, isNew, nil
}

// Session represents an in-memory session.
type Session struct {
	lock    *sync.RWMutex
	history []provider.Message
}

// History returns the session history.
func (s *Session) History(_ context.Context) ([]provider.Message, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()
	return s.history, nil
}

// AddMessage adds a message to the session.
func (s *Session) AddMessage(_ context.Context, message provider.Message) error {
	s.lock.Lock()
	s.history = append(s.history, message)
	s.lock.Unlock()
	return nil
}

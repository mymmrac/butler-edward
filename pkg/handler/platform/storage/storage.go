package storage

import (
	"context"
	"iter"
)

// Storage representation.
type Storage interface {
	// Get returns value by key.
	Get(ctx context.Context, userID string, key string) ([]byte, error)
	// Set sets value by key. Use nil value to delete the key.
	Set(ctx context.Context, userID string, key string, value []byte) error
	// ListPrefix returns all key-value pairs that start with the specified prefix.
	ListPrefix(ctx context.Context, userID string, prefix string) (iter.Seq2[string, []byte], error)
}

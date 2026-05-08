package bolt

import (
	"bytes"
	"context"
	"fmt"
	"iter"
	"time"

	bolt "go.etcd.io/bbolt"

	"github.com/mymmrac/butler-edward/pkg/module/collection"
)

// Bolt implementation of Storage.
type Bolt struct {
	db *bolt.DB
}

// NewBold creates a new Bolt storage.
func NewBold(path string) (*Bolt, error) {
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, fmt.Errorf("open bolt db: %w", err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err = tx.CreateBucketIfNotExists([]byte("users"))
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("create user bucket: %w", err)
	}

	return &Bolt{
		db: db,
	}, nil
}

// Close closes the Bolt storage.
func (b *Bolt) Close() error {
	return b.db.Close()
}

// Get returns the value for the given key.
func (b *Bolt) Get(_ context.Context, userID string, key string) ([]byte, error) {
	var value []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte("users"))
		if users == nil {
			return nil
		}

		user := users.Bucket([]byte(userID))
		if user == nil {
			return nil
		}

		value = bytes.Clone(user.Get([]byte(key)))
		return nil
	})
	return value, err
}

// Set sets the value for the given key.
func (b *Bolt) Set(_ context.Context, userID string, key string, value []byte) error {
	return b.db.Update(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte("users"))
		if users == nil {
			return fmt.Errorf("users bucket doesn't exist")
		}

		user, err := users.CreateBucketIfNotExists([]byte(userID))
		if err != nil {
			return fmt.Errorf("create user bucket: %w", err)
		}

		if value == nil {
			return user.Delete([]byte(key))
		}
		return user.Put([]byte(key), value)
	})
}

// ListPrefix returns all key-value pairs that start with the specified prefix.
func (b *Bolt) ListPrefix(_ context.Context, userID string, prefix string) (iter.Seq2[string, []byte], error) {
	var keyValues []collection.Pair[[]byte, []byte]
	err := b.db.View(func(tx *bolt.Tx) error {
		users := tx.Bucket([]byte("users"))
		if users == nil {
			return nil
		}

		user := users.Bucket([]byte(userID))
		if user == nil {
			return nil
		}

		c := user.Cursor()

		for k, v := c.Seek([]byte(prefix)); k != nil && bytes.HasPrefix(k, []byte(prefix)); k, v = c.Next() {
			keyValues = append(keyValues, collection.NewPair(bytes.Clone(k), bytes.Clone(v)))
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return func(yield func(string, []byte) bool) {
		for _, pair := range keyValues {
			if !yield(string(pair.First), pair.Second) {
				return
			}
		}
	}, nil
}

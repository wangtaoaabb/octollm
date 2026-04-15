package cache

import (
	"context"
	"errors"
)

// Common errors for cache operations.
var (
	ErrInvalidKey = errors.New("cache: invalid key")
)

// Meta is passed to Put and returned from Get. For FileStore, ContentType is kept in memory only
// (see entry); it is not written to disk.
// SourceURL is the original remote image URL; IndexHTTPStore requires it on Put for Get to refetch over HTTP.
type Meta struct {
	ContentType string `json:"content_type"`
	SourceURL   string `json:"source_url,omitempty"`
}

// Store is a persistent key-value store for image payloads keyed by a stable
// content hash (e.g. SHA-256 hex of canonical URL). Keys must be safe path
// segments: use hex hashes without slashes.
//
// Semantics:
//   - Get returns (nil, Meta{}, false, nil) when the key is absent.
//   - Put overwrites an existing key; if the key moves to a new time bucket, old files are removed.
//   - Delete removes one key from disk and from the in-memory index.
//   - DeleteBucket removes every key that belongs to a time bucket (for background sweepers).
//
// Eviction and cache cleanup are the caller's responsibility: this package does not run background
// sweepers or TTL deletion; use Delete/DeleteBucket (and for FileStore, any external file hygiene) as needed.
type Store interface {
	Get(ctx context.Context, key string) (data []byte, meta Meta, ok bool, err error)
	Put(ctx context.Context, key string, data []byte, meta Meta) error
	Delete(ctx context.Context, key string) error
	DeleteBucket(ctx context.Context, bucketID string) error
}

// Compile-time checks.
var (
	_ Store = (*FileStore)(nil)
	_ Store = (*IndexHTTPStore)(nil)
)

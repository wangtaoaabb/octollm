package cache

import (
	"crypto/sha256"
	"encoding/hex"
)

// KeyForURL returns a stable path-safe key (SHA-256 hex) for a remote image URL.
func KeyForURL(rawURL string) string {
	sum := sha256.Sum256([]byte(rawURL))
	return hex.EncodeToString(sum[:])
}

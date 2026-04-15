// Package limits defines byte limits and optional notification for remote image fetches
// used by the image_url_fetch engine.
package limits

import (
	"context"
	"errors"
)

// Errors returned when configured byte limits are exceeded.
var (
	ErrPerImageSizeExceeded   = errors.New("image size exceeds per-URL limit")
	ErrTotalImageSizeExceeded = errors.New("total image size exceeds per-request limit")
)

// ImageLimitKind classifies which limit triggered optional notifier callbacks.
type ImageLimitKind int

const (
	// ImageLimitPerURL is a single remote image over MaxBytesPerURL.
	ImageLimitPerURL ImageLimitKind = iota
	// ImageLimitPerRequest is the sum of unique image payload sizes over MaxBytesPerRequest.
	ImageLimitPerRequest
)

// ImageLimitEvent is passed to ImageFetchLimitNotifier when a limit is exceeded (the request fails).
type ImageLimitEvent struct {
	Kind        ImageLimitKind
	ImageURL    string // empty when Kind is ImageLimitPerRequest
	LimitBytes  int64
	ActualBytes int64 // per-URL: offending size; per-request: sum of decoded bytes
}

// ImageFetchLimitNotifier is optional. When set, OnLimitExceeded is invoked when the engine rejects
// a request due to MaxBytesPerURL or MaxBytesPerRequest. Implementations should return quickly;
// errors are ignored.
type ImageFetchLimitNotifier interface {
	OnLimitExceeded(ctx context.Context, ev ImageLimitEvent) error
}

// ImageURLFetchLimits groups byte limits and optional notification for remote image fetches.
// The zero value disables all limits and leaves Notifier nil.
type ImageURLFetchLimits struct {
	// MaxBytesPerURL is the maximum decoded image body size per URL. Zero disables the per-URL limit.
	// Responses may be rejected using Content-Length before reading the body; the body is always read through LimitReader.
	MaxBytesPerURL int64
	// MaxBytesPerRequest is the maximum sum of decoded sizes (unique URLs only) in one request. Zero disables.
	MaxBytesPerRequest int64
	// Notifier is optional; called when a limit causes the request to fail.
	Notifier ImageFetchLimitNotifier
}

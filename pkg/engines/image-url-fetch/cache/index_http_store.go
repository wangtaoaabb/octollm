package cache

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/infinigence/octollm/pkg/engines/image-url-fetch/metrics"
	"github.com/infinigence/octollm/pkg/engines/image-url-fetch/readbody"
)

// IndexHTTPStore implements Store by keeping only an in-memory index (same bucket layout as FileStore).
// Put/Delete/DeleteBucket only update the index; payload bytes from Put are not retained.
// Get loads the entry and performs an HTTP GET to SourceURL (must be set on Put via Meta).
//
// From the engine’s metrics, a successful Get here is a “cache hit” (IncCacheHits): the engine does not
// increment http_fetches_total for its own HTTPClient, even though Get may still issue HTTP for telemetry/refetch.
// Optional Metrics on this store records duration/decoded bytes for that internal GET separately from engine counters.
//
// Clearing or bounding this index is the caller's responsibility (e.g. periodic DeleteBucket); nothing runs automatically.
type IndexHTTPStore struct {
	HTTPClient *http.Client
	// MaxBytesPerURL is the maximum body size; 0 means unlimited (same semantics as engine limits).
	MaxBytesPerURL int64
	// Metrics optional; when set, successful HTTP inside Get records ObserveHTTPFetchDuration and ObserveDecodedBytes
	// for the refetch only (not IncHTTPFetches on the engine series).
	Metrics        *metrics.M
	BucketDuration time.Duration // zero defaults to 1 minute
	Now            func() time.Time

	ix *memoryIndex
}

// IndexHTTPStoreConfig configures NewIndexHTTPStore.
type IndexHTTPStoreConfig struct {
	HTTPClient     *http.Client
	MaxBytesPerURL int64
	Metrics        *metrics.M
	BucketDuration time.Duration
	Now            func() time.Time
}

// NewIndexHTTPStore builds an index-only store that refetches via HTTP on Get.
func NewIndexHTTPStore(cfg IndexHTTPStoreConfig) (*IndexHTTPStore, error) {
	if cfg.HTTPClient == nil {
		return nil, fmt.Errorf("cache: IndexHTTPStore requires HTTPClient")
	}
	return &IndexHTTPStore{
		HTTPClient:     cfg.HTTPClient,
		MaxBytesPerURL: cfg.MaxBytesPerURL,
		Metrics:        cfg.Metrics,
		BucketDuration: cfg.BucketDuration,
		Now:            cfg.Now,
		ix:             newMemoryIndex(),
	}, nil
}

func (s *IndexHTTPStore) bucketDuration() time.Duration {
	if s.BucketDuration <= 0 {
		return time.Minute
	}
	return s.BucketDuration
}

func (s *IndexHTTPStore) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

func (s *IndexHTTPStore) bucketID(t time.Time) string {
	t = t.UTC()
	d := s.bucketDuration()
	sec := int64(d / time.Second)
	if sec <= 0 {
		sec = 1
	}
	u := t.Unix()
	aligned := u - u%sec
	return fmt.Sprintf("%d", aligned)
}

// Get returns bytes from HTTP GET to the indexed SourceURL.
func (s *IndexHTTPStore) Get(ctx context.Context, key string) ([]byte, Meta, bool, error) {
	if err := validateKey(key); err != nil {
		return nil, Meta{}, false, err
	}
	e, ok := s.ix.snapshot(key)
	if !ok {
		return nil, Meta{}, false, nil
	}
	if e.sourceURL == "" {
		s.ix.deleteKey(key)
		return nil, Meta{}, false, nil
	}

	httpStart := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, e.sourceURL, nil)
	if err != nil {
		return nil, Meta{}, false, err
	}
	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, Meta{}, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, Meta{}, false, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	maxBytes := s.MaxBytesPerURL
	if maxBytes > 0 {
		if cl := readbody.ParseContentLength(resp); cl >= 0 && cl > maxBytes {
			return nil, Meta{}, false, fmt.Errorf("Content-Length %d exceeds limit %d", cl, maxBytes)
		}
	}

	ct := readbody.NormalizeImageContentType(resp.Header.Get("Content-Type"))
	data, err := readbody.ReadLimited(resp.Body, maxBytes)
	if err != nil {
		return nil, Meta{}, false, err
	}
	if maxBytes > 0 && int64(len(data)) > maxBytes {
		return nil, Meta{}, false, fmt.Errorf("body size %d exceeds limit %d", len(data), maxBytes)
	}

	if s.Metrics != nil {
		s.Metrics.ObserveHTTPFetchDuration(time.Since(httpStart))
		s.Metrics.ObserveDecodedBytes(int64(len(data)))
	}

	return data, Meta{ContentType: ct}, true, nil
}

// Put records key and metadata in the index only; data is not stored.
func (s *IndexHTTPStore) Put(ctx context.Context, key string, data []byte, meta Meta) error {
	_ = ctx
	if err := validateKey(key); err != nil {
		return err
	}
	if strings.TrimSpace(meta.SourceURL) == "" {
		return fmt.Errorf("cache: IndexHTTPStore Put requires Meta.SourceURL")
	}
	bucket := s.bucketID(s.now())
	ent := &entry{
		key:         key,
		bucketID:    bucket,
		dataPath:    "",
		sourceURL:   strings.TrimSpace(meta.SourceURL),
		size:        int64(len(data)),
		contentType: meta.ContentType,
	}
	s.ix.insertOrReplace(ent)
	return nil
}

// Delete removes one key from the index only.
func (s *IndexHTTPStore) Delete(ctx context.Context, key string) error {
	_ = ctx
	if err := validateKey(key); err != nil {
		return err
	}
	s.ix.deleteKey(key)
	return nil
}

// DeleteBucket removes every key in bucketID from the index only.
func (s *IndexHTTPStore) DeleteBucket(ctx context.Context, bucketID string) error {
	_ = ctx
	if strings.TrimSpace(bucketID) == "" {
		return nil
	}
	_ = s.ix.removeBucket(bucketID)
	return nil
}

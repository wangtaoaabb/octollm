package cache

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const dataFileSuffix = ".bin"

// FileStore implements Store with on-disk files under Root.
// Layout: {Root}/{bucketID}/{key}.bin only. Meta.ContentType is kept in memory (entry), not on disk.
// bucketID is derived from Now (UTC) and BucketDuration (aligned Unix seconds).
type FileStore struct {
	Root           string
	BucketDuration time.Duration // zero defaults to 1 minute
	// Now returns the current time for bucket selection; nil uses time.Now().UTC.
	Now func() time.Time

	ix *memoryIndex
}

// NewFileStore creates a file-backed store. Root must be non-empty.
func NewFileStore(root string) (*FileStore, error) {
	if strings.TrimSpace(root) == "" {
		return nil, fmt.Errorf("cache: empty root")
	}
	return &FileStore{
		Root:           root,
		BucketDuration: time.Minute,
		ix:             newMemoryIndex(),
	}, nil
}

func (s *FileStore) bucketDuration() time.Duration {
	if s.BucketDuration <= 0 {
		return time.Minute
	}
	return s.BucketDuration
}

func (s *FileStore) now() time.Time {
	if s.Now != nil {
		return s.Now().UTC()
	}
	return time.Now().UTC()
}

// bucketID returns a stable string for the time window containing t (UTC).
func (s *FileStore) bucketID(t time.Time) string {
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

func validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: empty", ErrInvalidKey)
	}
	if strings.Contains(key, "/") || strings.Contains(key, "\\") {
		return fmt.Errorf("%w: must not contain path separators", ErrInvalidKey)
	}
	return nil
}

func (s *FileStore) objectPaths(bucketID, key string) (dir string, dataPath string) {
	dir = filepath.Join(s.Root, bucketID)
	dataPath = filepath.Join(dir, key+dataFileSuffix)
	return dir, dataPath
}

// Get reads payload and meta for key. If the index points to missing files, the index entry is removed.
func (s *FileStore) Get(ctx context.Context, key string) ([]byte, Meta, bool, error) {
	_ = ctx
	if err := validateKey(key); err != nil {
		return nil, Meta{}, false, err
	}

	e, ok := s.ix.snapshot(key)
	if !ok {
		return nil, Meta{}, false, nil
	}

	data, err := os.ReadFile(e.dataPath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			s.ix.deleteKey(key)
			return nil, Meta{}, false, nil
		}
		return nil, Meta{}, false, err
	}
	return data, Meta{ContentType: e.contentType}, true, nil
}

// Put writes data and meta under the current time bucket. Overwrites same key; if the key moves to a new bucket, old files are removed.
func (s *FileStore) Put(ctx context.Context, key string, data []byte, meta Meta) error {
	_ = ctx
	if err := validateKey(key); err != nil {
		return err
	}

	bucket := s.bucketID(s.now())
	dir, dataPath := s.objectPaths(bucket, key)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmpData, err := os.CreateTemp(dir, ".bin-*.tmp")
	if err != nil {
		return err
	}
	tmpDataPath := tmpData.Name()
	_, werr := tmpData.Write(data)
	cerr := tmpData.Close()
	if werr != nil {
		_ = os.Remove(tmpDataPath)
		return werr
	}
	if cerr != nil {
		_ = os.Remove(tmpDataPath)
		return cerr
	}

	if err := os.Rename(tmpDataPath, dataPath); err != nil {
		_ = os.Remove(tmpDataPath)
		return err
	}

	ent := &entry{
		key:         key,
		bucketID:    bucket,
		dataPath:    dataPath,
		size:        int64(len(data)),
		contentType: meta.ContentType,
	}
	old := s.ix.insertOrReplace(ent)
	if old != nil && old.dataPath != ent.dataPath {
		_ = os.Remove(old.dataPath)
	}
	return nil
}

// Delete removes one key from disk and index.
func (s *FileStore) Delete(ctx context.Context, key string) error {
	_ = ctx
	if err := validateKey(key); err != nil {
		return err
	}
	prev := s.ix.deleteKey(key)
	if prev == nil {
		return nil
	}
	_ = os.Remove(prev.dataPath)
	return nil
}

// DeleteBucket removes every key tracked under bucketID from disk and from the index.
func (s *FileStore) DeleteBucket(ctx context.Context, bucketID string) error {
	_ = ctx
	if strings.TrimSpace(bucketID) == "" {
		return nil
	}
	ents := s.ix.removeBucket(bucketID)
	for i := range ents {
		_ = os.Remove(ents[i].dataPath)
	}

	_ = os.Remove(filepath.Join(s.Root, bucketID))
	return nil
}

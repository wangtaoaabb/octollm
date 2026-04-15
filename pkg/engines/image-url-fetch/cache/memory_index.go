package cache

import (
	"sync"
)

// entry describes one cached object in memory. Paths are absolute for FileStore.
// contentType is the MIME type for the payload (e.g. image/jpeg); it is not persisted to disk.
// sourceURL is used by IndexHTTPStore (Get issues HTTP); FileStore leaves it empty.
type entry struct {
	key         string
	bucketID    string
	dataPath    string
	sourceURL   string
	size        int64
	contentType string
}

// memoryIndex holds byKey and byBucket reverse index for sweeper-aligned eviction.
// It must be updated under the same lock as the Store that owns it.
type memoryIndex struct {
	mu sync.RWMutex

	byKey    map[string]*entry
	byBucket map[string]map[string]struct{}
}

func newMemoryIndex() *memoryIndex {
	return &memoryIndex{
		byKey:    make(map[string]*entry),
		byBucket: make(map[string]map[string]struct{}),
	}
}

// snapshot returns a copy of the entry for use outside the lock.
func (ix *memoryIndex) snapshot(key string) (e entry, ok bool) {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	cur := ix.byKey[key]
	if cur == nil {
		return entry{}, false
	}
	return *cur, true
}

func (ix *memoryIndex) deleteKey(key string) *entry {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	cur := ix.byKey[key]
	if cur == nil {
		return nil
	}
	delete(ix.byKey, key)
	if m, ok := ix.byBucket[cur.bucketID]; ok {
		delete(m, key)
		if len(m) == 0 {
			delete(ix.byBucket, cur.bucketID)
		}
	}
	return cur
}

// insertOrReplace updates the index. If the key existed under another bucket, the old entry is returned for file removal.
func (ix *memoryIndex) insertOrReplace(e *entry) (old *entry) {
	ix.mu.Lock()
	defer ix.mu.Unlock()
	return ix.insertUnderLock(e)
}

func (ix *memoryIndex) insertUnderLock(e *entry) (old *entry) {
	if prev := ix.byKey[e.key]; prev != nil {
		old = prev
		if m, ok := ix.byBucket[prev.bucketID]; ok {
			delete(m, e.key)
			if len(m) == 0 {
				delete(ix.byBucket, prev.bucketID)
			}
		}
	}
	ix.byKey[e.key] = e
	if ix.byBucket[e.bucketID] == nil {
		ix.byBucket[e.bucketID] = make(map[string]struct{})
	}
	ix.byBucket[e.bucketID][e.key] = struct{}{}
	return old
}

// removeBucket returns all entries removed for that bucketID (for unlinking files).
func (ix *memoryIndex) removeBucket(bucketID string) []entry {
	ix.mu.Lock()
	defer ix.mu.Unlock()

	m := ix.byBucket[bucketID]
	if len(m) == 0 {
		return nil
	}
	out := make([]entry, 0, len(m))
	for k := range m {
		if cur := ix.byKey[k]; cur != nil {
			out = append(out, *cur)
			delete(ix.byKey, k)
		}
	}
	delete(ix.byBucket, bucketID)
	return out
}

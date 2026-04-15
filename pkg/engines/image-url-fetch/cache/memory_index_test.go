package cache

import (
	"sort"
	"testing"
)

func TestMemoryIndex_snapshot_miss(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	_, ok := ix.snapshot("nope")
	if ok {
		t.Fatal("expected miss on empty index")
	}
}

func TestMemoryIndex_insertOrReplace_snapshot_hit(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	ix.insertOrReplace(&entry{
		key: "k1", bucketID: "b1", dataPath: "/a.bin", size: 3, contentType: "image/png",
	})
	got, ok := ix.snapshot("k1")
	if !ok {
		t.Fatal("expected hit")
	}
	if got.key != "k1" || got.bucketID != "b1" || got.dataPath != "/a.bin" || got.size != 3 || got.contentType != "image/png" {
		t.Fatalf("entry mismatch: %+v", got)
	}
}

func TestMemoryIndex_snapshot_returnsCopy(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	ix.insertOrReplace(&entry{key: "k", bucketID: "b", dataPath: "/orig"})
	got, ok := ix.snapshot("k")
	if !ok {
		t.Fatal("expected hit")
	}
	got.dataPath = "/mutated"
	got2, _ := ix.snapshot("k")
	if got2.dataPath != "/orig" {
		t.Fatalf("mutating snapshot should not change stored entry: %+v", got2)
	}
}

func TestMemoryIndex_deleteKey_miss(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	if cur := ix.deleteKey("missing"); cur != nil {
		t.Fatalf("expected nil, got %+v", cur)
	}
}

func TestMemoryIndex_deleteKey_removesFromBucket(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	ix.insertOrReplace(&entry{key: "k", bucketID: "b1", dataPath: "/x.bin"})
	cur := ix.deleteKey("k")
	if cur == nil || cur.key != "k" {
		t.Fatalf("deleteKey should return removed entry: %+v", cur)
	}
	if _, ok := ix.snapshot("k"); ok {
		t.Fatal("key should be gone")
	}
}

func TestMemoryIndex_insertOrReplace_replaceSameBucket(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	old := ix.insertOrReplace(&entry{key: "k", bucketID: "b1", dataPath: "/first.bin", size: 1})
	if old != nil {
		t.Fatalf("first insert should not return old: %+v", old)
	}
	old = ix.insertOrReplace(&entry{key: "k", bucketID: "b1", dataPath: "/second.bin", size: 2})
	if old == nil || old.dataPath != "/first.bin" {
		t.Fatalf("expected old entry, got %+v", old)
	}
	got, _ := ix.snapshot("k")
	if got.dataPath != "/second.bin" || got.size != 2 {
		t.Fatalf("got %+v", got)
	}
}

func TestMemoryIndex_insertOrReplace_movesBucket(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	ix.insertOrReplace(&entry{key: "k", bucketID: "b1", dataPath: "/a.bin"})
	old := ix.insertOrReplace(&entry{key: "k", bucketID: "b2", dataPath: "/b.bin"})
	if old == nil || old.bucketID != "b1" {
		t.Fatalf("expected old from b1: %+v", old)
	}
	got, _ := ix.snapshot("k")
	if got.bucketID != "b2" {
		t.Fatalf("got %+v", got)
	}
	// remove b1 should not remove k (k is in b2 now)
	removed := ix.removeBucket("b1")
	if len(removed) != 0 {
		t.Fatalf("b1 should be empty: %v", removed)
	}
	if _, ok := ix.snapshot("k"); !ok {
		t.Fatal("k should still exist in b2")
	}
	removed = ix.removeBucket("b2")
	if len(removed) != 1 || removed[0].key != "k" {
		t.Fatalf("removeBucket b2: %v", removed)
	}
	if _, ok := ix.snapshot("k"); ok {
		t.Fatal("k should be gone after b2 removed")
	}
}

func TestMemoryIndex_removeBucket_multipleKeys(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	ix.insertOrReplace(&entry{key: "a", bucketID: "bx", dataPath: "/a.bin"})
	ix.insertOrReplace(&entry{key: "b", bucketID: "bx", dataPath: "/b.bin"})
	ix.insertOrReplace(&entry{key: "c", bucketID: "other", dataPath: "/c.bin"})

	out := ix.removeBucket("bx")
	if len(out) != 2 {
		t.Fatalf("len=%d %v", len(out), out)
	}
	keys := []string{out[0].key, out[1].key}
	sort.Strings(keys)
	if keys[0] != "a" || keys[1] != "b" {
		t.Fatalf("unexpected keys %v", keys)
	}
	if _, ok := ix.snapshot("a"); ok {
		t.Fatal("a should be removed")
	}
	if _, ok := ix.snapshot("b"); ok {
		t.Fatal("b should be removed")
	}
	got, ok := ix.snapshot("c")
	if !ok || got.bucketID != "other" {
		t.Fatalf("c should remain: ok=%v %+v", ok, got)
	}
	// second remove of same bucket is no-op
	if again := ix.removeBucket("bx"); len(again) != 0 {
		t.Fatalf("second remove: %v", again)
	}
}

func TestMemoryIndex_removeBucket_unknown(t *testing.T) {
	t.Parallel()
	ix := newMemoryIndex()
	ix.insertOrReplace(&entry{key: "k", bucketID: "b", dataPath: "/x"})
	if out := ix.removeBucket("nobucket"); len(out) != 0 {
		t.Fatalf("expected empty slice, got %v", out)
	}
	if _, ok := ix.snapshot("k"); !ok {
		t.Fatal("unrelated key should remain")
	}
}

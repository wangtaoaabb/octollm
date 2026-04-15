package cache

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFileStore_PutGetDelete(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	fs, err := NewFileStore(root)
	if err != nil {
		t.Fatal(err)
	}
	fs.BucketDuration = time.Minute
	fixed := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	fs.Now = func() time.Time { return fixed }

	key := "aabbccddeeff"
	payload := []byte("hello-image")
	meta := Meta{ContentType: "image/png"}

	ctx := context.Background()
	if err := fs.Put(ctx, key, payload, meta); err != nil {
		t.Fatalf("put: %v", err)
	}

	got, gm, ok, err := fs.Get(ctx, key)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !ok {
		t.Fatal("expected hit")
	}
	if string(got) != string(payload) {
		t.Fatalf("payload %q got %q", payload, got)
	}
	if gm.ContentType != meta.ContentType {
		t.Fatalf("meta %+v", gm)
	}

	bucket := fs.bucketID(fixed)
	dataPath := filepath.Join(root, bucket, key+".bin")
	if _, err := os.Stat(dataPath); err != nil {
		t.Fatalf("stat data: %v", err)
	}

	if err := fs.Delete(ctx, key); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, _, ok, err := fs.Get(ctx, key); err != nil || ok {
		t.Fatalf("after delete: ok=%v err=%v", ok, err)
	}
}

func TestFileStore_DeleteBucketSyncsIndex(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	fs, err := NewFileStore(root)
	if err != nil {
		t.Fatal(err)
	}
	fs.BucketDuration = time.Minute
	fixed := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	fs.Now = func() time.Time { return fixed }

	key := "001122334455"
	ctx := context.Background()
	if err := fs.Put(ctx, key, []byte("x"), Meta{ContentType: "image/jpeg"}); err != nil {
		t.Fatal(err)
	}
	bucket := fs.bucketID(fixed)
	if err := fs.DeleteBucket(ctx, bucket); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := fs.Get(ctx, key); err != nil || ok {
		t.Fatalf("after bucket delete: ok=%v err=%v", ok, err)
	}
}

func TestFileStore_GetRemovesStaleIndex(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	fs, err := NewFileStore(root)
	if err != nil {
		t.Fatal(err)
	}
	fs.BucketDuration = time.Minute
	fixed := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	fs.Now = func() time.Time { return fixed }

	key := "ffeeddccbbaa"
	ctx := context.Background()
	if err := fs.Put(ctx, key, []byte("z"), Meta{}); err != nil {
		t.Fatal(err)
	}
	e, ok := fs.ix.snapshot(key)
	if !ok {
		t.Fatal("expected index entry")
	}
	if err := os.Remove(e.dataPath); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := fs.Get(ctx, key); err != nil || ok {
		t.Fatalf("expected miss after manual delete: ok=%v err=%v", ok, err)
	}
	if _, ok := fs.ix.snapshot(key); ok {
		t.Fatal("index should be cleaned")
	}
}

func TestFileStore_InvalidKey(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	fs, err := NewFileStore(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, _, _, err := fs.Get(ctx, ""); err == nil {
		t.Fatal("expected error for empty key")
	}
	if err := fs.Put(ctx, "", []byte{}, Meta{}); err == nil {
		t.Fatal("expected error for empty key")
	}
	if _, _, _, err := fs.Get(ctx, "bad/key"); err == nil {
		t.Fatal("expected error for key with slash")
	}
}

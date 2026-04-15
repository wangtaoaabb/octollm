package cache

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIndexHTTPStore_PutRequiresSourceURL(t *testing.T) {
	t.Parallel()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("x"))
	}))
	t.Cleanup(s.Close)

	st, err := NewIndexHTTPStore(IndexHTTPStoreConfig{HTTPClient: s.Client()})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := st.Put(ctx, "abc123", []byte("x"), Meta{ContentType: "image/png"}); err == nil {
		t.Fatal("expected error without SourceURL")
	}
}

func TestIndexHTTPStore_GetRefetchesHTTP(t *testing.T) {
	t.Parallel()
	var calls int
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("img"))
	}))
	t.Cleanup(s.Close)

	st, err := NewIndexHTTPStore(IndexHTTPStoreConfig{HTTPClient: s.Client()})
	if err != nil {
		t.Fatal(err)
	}
	st.BucketDuration = time.Minute
	fixed := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	st.Now = func() time.Time { return fixed }

	key := "deadbeef"
	ctx := context.Background()
	if err := st.Put(ctx, key, []byte("img"), Meta{ContentType: "image/png", SourceURL: s.URL + "/p"}); err != nil {
		t.Fatal(err)
	}
	data, meta, ok, err := st.Get(ctx, key)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected hit")
	}
	if string(data) != "img" {
		t.Fatalf("data %q", data)
	}
	if meta.ContentType != "image/png" {
		t.Fatalf("meta %+v", meta)
	}
	if calls != 1 {
		t.Fatalf("calls=%d", calls)
	}
}

func TestIndexHTTPStore_DeleteRemovesKey(t *testing.T) {
	t.Parallel()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("img"))
	}))
	t.Cleanup(s.Close)

	st, err := NewIndexHTTPStore(IndexHTTPStoreConfig{HTTPClient: s.Client()})
	if err != nil {
		t.Fatal(err)
	}
	st.BucketDuration = time.Minute
	fixed := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	st.Now = func() time.Time { return fixed }

	key := "aabbccddeeff"
	ctx := context.Background()
	if err := st.Put(ctx, key, []byte("img"), Meta{ContentType: "image/png", SourceURL: s.URL + "/p"}); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := st.Get(ctx, key); err != nil || !ok {
		t.Fatalf("before delete: ok=%v err=%v", ok, err)
	}
	if err := st.Delete(ctx, key); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := st.Get(ctx, key); err != nil || ok {
		t.Fatalf("after delete: ok=%v err=%v", ok, err)
	}
}

func TestIndexHTTPStore_DeleteBucketRemovesKeys(t *testing.T) {
	t.Parallel()
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("x"))
	}))
	t.Cleanup(s.Close)

	st, err := NewIndexHTTPStore(IndexHTTPStoreConfig{HTTPClient: s.Client()})
	if err != nil {
		t.Fatal(err)
	}
	st.BucketDuration = time.Minute
	fixed := time.Date(2026, 4, 9, 12, 0, 0, 0, time.UTC)
	st.Now = func() time.Time { return fixed }

	key := "001122334455"
	ctx := context.Background()
	if err := st.Put(ctx, key, []byte("x"), Meta{ContentType: "image/jpeg", SourceURL: s.URL + "/p"}); err != nil {
		t.Fatal(err)
	}
	bucket := st.bucketID(fixed)
	if err := st.DeleteBucket(ctx, bucket); err != nil {
		t.Fatal(err)
	}
	if _, _, ok, err := st.Get(ctx, key); err != nil || ok {
		t.Fatalf("after bucket delete: ok=%v err=%v", ok, err)
	}
}

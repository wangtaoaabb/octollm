package cache

import (
	"testing"
)

func TestKeyForURL_stable(t *testing.T) {
	t.Parallel()
	a := KeyForURL("https://example.com/a.png")
	b := KeyForURL("https://example.com/a.png")
	if a != b {
		t.Fatalf("expected stable key, got %q vs %q", a, b)
	}
	if len(a) != 64 {
		t.Fatalf("expected hex sha256 length 64, got %d", len(a))
	}
	if a == KeyForURL("https://example.com/b.png") {
		t.Fatal("different URLs must differ")
	}
}

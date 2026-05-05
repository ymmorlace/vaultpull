package sync

import (
	"testing"
	"time"
)

func TestSecretCache_MissOnEmpty(t *testing.T) {
	c := NewSecretCache(time.Minute)
	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected cache miss for unknown key")
	}
}

func TestSecretCache_HitAfterSet(t *testing.T) {
	c := NewSecretCache(time.Minute)
	c.Set("foo", "bar")
	v, ok := c.Get("foo")
	if !ok {
		t.Fatal("expected cache hit")
	}
	if v != "bar" {
		t.Fatalf("expected \"bar\", got %q", v)
	}
}

func TestSecretCache_ExpiredEntryIsMiss(t *testing.T) {
	c := NewSecretCache(10 * time.Millisecond)
	c.Set("key", "value")
	time.Sleep(30 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected expired entry to be a miss")
	}
}

func TestSecretCache_ZeroTTLAlwaysMisses(t *testing.T) {
	c := NewSecretCache(0)
	c.Set("k", "v")
	_, ok := c.Get("k")
	if ok {
		t.Fatal("zero-TTL cache should never return a hit")
	}
}

func TestSecretCache_Invalidate(t *testing.T) {
	c := NewSecretCache(time.Minute)
	c.Set("a", "1")
	c.Set("b", "2")
	c.Invalidate("a")
	if _, ok := c.Get("a"); ok {
		t.Fatal("expected invalidated key to be missing")
	}
	if _, ok := c.Get("b"); !ok {
		t.Fatal("expected non-invalidated key to remain")
	}
}

func TestSecretCache_Flush(t *testing.T) {
	c := NewSecretCache(time.Minute)
	c.Set("x", "1")
	c.Set("y", "2")
	c.Flush()
	if c.Size() != 0 {
		t.Fatalf("expected empty cache after flush, got size %d", c.Size())
	}
}

func TestSecretCache_Size(t *testing.T) {
	c := NewSecretCache(time.Minute)
	if c.Size() != 0 {
		t.Fatal("expected size 0 for new cache")
	}
	c.Set("p", "1")
	c.Set("q", "2")
	if c.Size() != 2 {
		t.Fatalf("expected size 2, got %d", c.Size())
	}
}

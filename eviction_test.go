package stash

import (
	"testing"
	"time"
)

func TestLRUEviction(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](3),
		WithEviction[string, int](LRU),
	)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0)

	// Access "a" so it becomes most recently used.
	c.Get("a")

	// Adding "d" should evict "b" (least recently used).
	c.Set("d", 4, 0)

	if _, ok := c.Get("b"); ok {
		t.Fatal("expected b to be evicted (LRU)")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected a to still exist")
	}
	if _, ok := c.Get("c"); !ok {
		t.Fatal("expected c to still exist")
	}
	if _, ok := c.Get("d"); !ok {
		t.Fatal("expected d to exist")
	}
}

func TestLRUEvictionOrder(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](2),
		WithEviction[string, int](LRU),
	)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	// "a" is LRU, adding "c" should evict "a".
	c.Set("c", 3, 0)
	if _, ok := c.Get("a"); ok {
		t.Fatal("a should be evicted")
	}
	if c.Count() != 2 {
		t.Fatalf("expected 2 items, got %d", c.Count())
	}
}

func TestLFUEviction(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](3),
		WithEviction[string, int](LFU),
	)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0)

	// Access "a" and "c" more frequently.
	c.Get("a")
	c.Get("a")
	c.Get("c")

	// "b" has lowest frequency, adding "d" should evict "b".
	c.Set("d", 4, 0)

	if _, ok := c.Get("b"); ok {
		t.Fatal("expected b to be evicted (LFU)")
	}
	if _, ok := c.Get("a"); !ok {
		t.Fatal("expected a to still exist")
	}
}

func TestMaxSizeEnforced(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](5),
		WithEviction[string, int](LRU),
	)
	for i := 0; i < 20; i++ {
		c.Set(string(rune('a'+i)), i, 0)
	}
	if c.Count() > 5 {
		t.Fatalf("expected at most 5 items, got %d", c.Count())
	}
}

func TestOnEvictedWithMaxSize(t *testing.T) {
	evictedCount := 0
	c := New[string, int](
		WithMaxSize[string, int](2),
		WithEviction[string, int](LRU),
		WithOnEvicted[string, int](func(k string, v int) {
			evictedCount++
		}),
	)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0) // evicts "a"
	if evictedCount != 1 {
		t.Fatalf("expected 1 eviction callback, got %d", evictedCount)
	}
}

func TestNoEvictionUnbounded(t *testing.T) {
	c := New[string, int]()
	for i := 0; i < 100; i++ {
		c.Set(string(rune('a'+i)), i, 0)
	}
	if c.Count() != 100 {
		t.Fatalf("expected 100 items with no eviction, got %d", c.Count())
	}
}

func TestLRUEvictorRemove(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](5),
		WithEviction[string, int](LRU),
	)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0)
	c.Delete("b")
	if c.Count() != 2 {
		t.Fatalf("expected 2, got %d", c.Count())
	}
}

func TestLFUEvictorRemove(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](5),
		WithEviction[string, int](LFU),
	)
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Delete("a")
	if c.Count() != 1 {
		t.Fatalf("expected 1, got %d", c.Count())
	}
}

func TestEvictionWithExpiry(t *testing.T) {
	evicted := make(map[string]bool)
	c := New[string, int](
		WithMaxSize[string, int](3),
		WithEviction[string, int](LRU),
		WithOnEvicted[string, int](func(k string, v int) {
			evicted[k] = true
		}),
	)
	c.Set("a", 1, 20*time.Millisecond)
	c.Set("b", 2, 0)
	c.Set("c", 3, 0)
	time.Sleep(30 * time.Millisecond)
	c.DeleteExpired()
	if !evicted["a"] {
		t.Fatal("expected a to be evicted via expiry")
	}
}

func TestAddRespectsMaxSize(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](2),
		WithEviction[string, int](LRU),
	)
	_ = c.Add("a", 1, 0)
	_ = c.Add("b", 2, 0)
	_ = c.Add("c", 3, 0) // should evict "a"
	if c.Count() != 2 {
		t.Fatalf("expected 2, got %d", c.Count())
	}
	if _, ok := c.Get("a"); ok {
		t.Fatal("a should be evicted")
	}
}

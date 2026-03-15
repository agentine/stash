package stash

import (
	"errors"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "value", 0)
	v, ok := c.Get("key")
	if !ok || v != "value" {
		t.Fatalf("expected value, got %q %v", v, ok)
	}
}

func TestGetMissing(t *testing.T) {
	c := New[string, int]()
	_, ok := c.Get("nope")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestSetWithTTL(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "val", 50*time.Millisecond)
	v, ok := c.Get("key")
	if !ok || v != "val" {
		t.Fatal("expected found before expiry")
	}
	time.Sleep(60 * time.Millisecond)
	_, ok = c.Get("key")
	if ok {
		t.Fatal("expected expired")
	}
}

func TestSetDefaultTTL(t *testing.T) {
	c := New[string, string](WithDefaultTTL[string, string](50 * time.Millisecond))
	c.SetDefault("key", "val")
	_, ok := c.Get("key")
	if !ok {
		t.Fatal("expected found before expiry")
	}
	time.Sleep(60 * time.Millisecond)
	_, ok = c.Get("key")
	if ok {
		t.Fatal("expected expired with default TTL")
	}
}

func TestSetDefaultTTLSentinel(t *testing.T) {
	c := New[string, string](WithDefaultTTL[string, string](50 * time.Millisecond))
	c.Set("key", "val", DefaultTTL)
	time.Sleep(60 * time.Millisecond)
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected expired via DefaultTTL sentinel")
	}
}

func TestNoExpiration(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "val", 0)
	time.Sleep(10 * time.Millisecond)
	_, ok := c.Get("key")
	if !ok {
		t.Fatal("expected no expiration")
	}
}

func TestAdd(t *testing.T) {
	c := New[string, string]()
	err := c.Add("key", "val1", 0)
	if err != nil {
		t.Fatalf("first add should succeed: %v", err)
	}
	err = c.Add("key", "val2", 0)
	if !errors.Is(err, ErrKeyExists) {
		t.Fatalf("expected ErrKeyExists, got %v", err)
	}
	v, _ := c.Get("key")
	if v != "val1" {
		t.Fatalf("expected val1, got %q", v)
	}
}

func TestAddExpired(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "old", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	err := c.Add("key", "new", 0)
	if err != nil {
		t.Fatalf("add on expired key should succeed: %v", err)
	}
	v, _ := c.Get("key")
	if v != "new" {
		t.Fatalf("expected new, got %q", v)
	}
}

func TestReplace(t *testing.T) {
	c := New[string, string]()
	err := c.Replace("key", "val", 0)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatalf("expected ErrKeyNotFound, got %v", err)
	}
	c.Set("key", "old", 0)
	err = c.Replace("key", "new", 0)
	if err != nil {
		t.Fatalf("replace should succeed: %v", err)
	}
	v, _ := c.Get("key")
	if v != "new" {
		t.Fatalf("expected new, got %q", v)
	}
}

func TestReplaceExpired(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "old", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	err := c.Replace("key", "new", 0)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatal("replace on expired key should fail")
	}
}

func TestDelete(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "val", 0)
	c.Delete("key")
	_, ok := c.Get("key")
	if ok {
		t.Fatal("expected deleted")
	}
}

func TestDeleteNonexistent(t *testing.T) {
	c := New[string, string]()
	c.Delete("nope") // should not panic
}

func TestGetWithExpiration(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "val", time.Hour)
	v, exp, ok := c.GetWithExpiration("key")
	if !ok || v != "val" {
		t.Fatal("expected found")
	}
	if exp.IsZero() {
		t.Fatal("expected non-zero expiration")
	}
}

func TestGetWithExpirationNoTTL(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "val", 0)
	_, exp, ok := c.GetWithExpiration("key")
	if !ok {
		t.Fatal("expected found")
	}
	if !exp.IsZero() {
		t.Fatal("expected zero expiration for no-TTL item")
	}
}

func TestGetWithExpirationMissing(t *testing.T) {
	c := New[string, string]()
	_, _, ok := c.GetWithExpiration("nope")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestGetOrSetCached(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "cached", 0)
	called := false
	v, err := c.GetOrSet("key", func() (string, error) {
		called = true
		return "new", nil
	}, 0)
	if err != nil || v != "cached" || called {
		t.Fatalf("expected cached value without calling fn, got %q called=%v err=%v", v, called, err)
	}
}

func TestGetOrSetFetched(t *testing.T) {
	c := New[string, string]()
	v, err := c.GetOrSet("key", func() (string, error) {
		return "fetched", nil
	}, 0)
	if err != nil || v != "fetched" {
		t.Fatalf("expected fetched, got %q err=%v", v, err)
	}
	v2, ok := c.Get("key")
	if !ok || v2 != "fetched" {
		t.Fatal("expected value to be stored")
	}
}

func TestGetOrSetError(t *testing.T) {
	c := New[string, string]()
	_, err := c.GetOrSet("key", func() (string, error) {
		return "", errors.New("fail")
	}, 0)
	if err == nil {
		t.Fatal("expected error")
	}
	_, ok := c.Get("key")
	if ok {
		t.Fatal("value should not be stored on error")
	}
}

func TestItems(t *testing.T) {
	c := New[string, int]()
	c.Set("a", 1, 0)
	c.Set("b", 2, 0)
	c.Set("expired", 3, 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	items := c.Items()
	if len(items) != 2 {
		t.Fatalf("expected 2 unexpired items, got %d", len(items))
	}
	if items["a"].Value != 1 || items["b"].Value != 2 {
		t.Fatal("unexpected item values")
	}
}

func TestFlush(t *testing.T) {
	c := New[string, string]()
	c.Set("a", "1", 0)
	c.Set("b", "2", 0)
	c.Flush()
	if c.Count() != 0 {
		t.Fatalf("expected 0 after flush, got %d", c.Count())
	}
}

func TestCount(t *testing.T) {
	c := New[string, string]()
	if c.Count() != 0 {
		t.Fatal("expected 0")
	}
	c.Set("a", "1", 0)
	c.Set("b", "2", 0)
	if c.Count() != 2 {
		t.Fatalf("expected 2, got %d", c.Count())
	}
}

func TestDeleteExpired(t *testing.T) {
	c := New[string, string]()
	c.Set("keep", "val", 0)
	c.Set("expire", "val", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	c.DeleteExpired()
	if c.Count() != 1 {
		t.Fatalf("expected 1 after DeleteExpired, got %d", c.Count())
	}
	_, ok := c.Get("keep")
	if !ok {
		t.Fatal("keep should still exist")
	}
}

func TestJanitorCleansUp(t *testing.T) {
	c := New[string, string](
		WithCleanupInterval[string, string](25 * time.Millisecond),
	)
	defer c.Stop()
	c.Set("key", "val", 10*time.Millisecond)
	time.Sleep(50 * time.Millisecond)
	if c.Count() != 0 {
		t.Fatal("janitor should have cleaned up expired item")
	}
}

func TestStop(t *testing.T) {
	c := New[string, string](
		WithCleanupInterval[string, string](10 * time.Millisecond),
	)
	c.Stop()
	c.Set("key", "val", 10*time.Millisecond)
	time.Sleep(30 * time.Millisecond)
	// Item should still be in map (expired but not cleaned) since janitor was stopped.
	if c.Count() == 0 {
		t.Fatal("janitor should be stopped, item should not be cleaned")
	}
}

func TestOnEvictedDelete(t *testing.T) {
	var evictedKey string
	var evictedVal string
	c := New[string, string](
		WithOnEvicted[string, string](func(k string, v string) {
			evictedKey = k
			evictedVal = v
		}),
	)
	c.Set("key", "val", 0)
	c.Delete("key")
	if evictedKey != "key" || evictedVal != "val" {
		t.Fatalf("expected eviction callback, got %q=%q", evictedKey, evictedVal)
	}
}

func TestOnEvictedExpiry(t *testing.T) {
	evicted := make(map[string]string)
	c := New[string, string](
		WithOnEvicted[string, string](func(k string, v string) {
			evicted[k] = v
		}),
	)
	c.Set("a", "1", 10*time.Millisecond)
	c.Set("b", "2", 0)
	time.Sleep(20 * time.Millisecond)
	c.DeleteExpired()
	if len(evicted) != 1 || evicted["a"] != "1" {
		t.Fatalf("expected eviction of expired item, got %v", evicted)
	}
}

func TestSetOverwrite(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "v1", 0)
	c.Set("key", "v2", 0)
	v, _ := c.Get("key")
	if v != "v2" {
		t.Fatalf("expected v2, got %q", v)
	}
}

func TestIntKeys(t *testing.T) {
	c := New[int, string]()
	c.Set(42, "answer", 0)
	v, ok := c.Get(42)
	if !ok || v != "answer" {
		t.Fatal("expected int key to work")
	}
}

func TestItemExpired(t *testing.T) {
	item := Item[string]{Value: "val", Expiration: time.Now().Add(-time.Second).UnixNano()}
	if !item.Expired() {
		t.Fatal("expected expired")
	}
	item2 := Item[string]{Value: "val", Expiration: 0}
	if item2.Expired() {
		t.Fatal("zero expiration should not be expired")
	}
	item3 := Item[string]{Value: "val", Expiration: time.Now().Add(time.Hour).UnixNano()}
	if item3.Expired() {
		t.Fatal("future expiration should not be expired")
	}
}

func TestGetWithExpirationExpired(t *testing.T) {
	c := New[string, string]()
	c.Set("key", "val", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	_, _, ok := c.GetWithExpiration("key")
	if ok {
		t.Fatal("expected expired item not found")
	}
}

func TestAddWithEvictor(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](2),
		WithEviction[string, int](LFU),
	)
	_ = c.Add("a", 1, 0)
	_ = c.Add("b", 2, 0)
	c.Get("a")
	c.Get("a")
	_ = c.Add("c", 3, 0) // should evict b (lowest freq)
	if _, ok := c.Get("b"); ok {
		t.Fatal("b should be evicted")
	}
}

func TestReplaceWithEvictor(t *testing.T) {
	c := New[string, int](
		WithMaxSize[string, int](5),
		WithEviction[string, int](LRU),
	)
	c.Set("key", 1, 0)
	err := c.Replace("key", 2, 0)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := c.Get("key")
	if v != 2 {
		t.Fatalf("expected 2, got %d", v)
	}
}

func TestGetWithExpirationWithEvictor(t *testing.T) {
	c := New[string, string](
		WithMaxSize[string, string](5),
		WithEviction[string, string](LRU),
	)
	c.Set("key", "val", time.Hour)
	v, exp, ok := c.GetWithExpiration("key")
	if !ok || v != "val" || exp.IsZero() {
		t.Fatal("expected value with expiration")
	}
}

func TestStopWithoutJanitor(t *testing.T) {
	c := New[string, string]()
	c.Stop() // should not panic when no janitor
}

package stash

import (
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestShardedSetGet(t *testing.T) {
	sc := NewSharded[string, string](4)
	sc.Set("key", "val", 0)
	v, ok := sc.Get("key")
	if !ok || v != "val" {
		t.Fatalf("expected val, got %q %v", v, ok)
	}
}

func TestShardedMissing(t *testing.T) {
	sc := NewSharded[string, string](4)
	_, ok := sc.Get("nope")
	if ok {
		t.Fatal("expected not found")
	}
}

func TestShardedTTL(t *testing.T) {
	sc := NewSharded[string, string](4,
		WithDefaultTTL[string, string](50*time.Millisecond),
	)
	sc.SetDefault("key", "val")
	time.Sleep(60 * time.Millisecond)
	_, ok := sc.Get("key")
	if ok {
		t.Fatal("expected expired")
	}
}

func TestShardedDelete(t *testing.T) {
	sc := NewSharded[string, string](4)
	sc.Set("key", "val", 0)
	sc.Delete("key")
	_, ok := sc.Get("key")
	if ok {
		t.Fatal("expected deleted")
	}
}

func TestShardedAdd(t *testing.T) {
	sc := NewSharded[string, string](4)
	err := sc.Add("key", "v1", 0)
	if err != nil {
		t.Fatal(err)
	}
	err = sc.Add("key", "v2", 0)
	if !errors.Is(err, ErrKeyExists) {
		t.Fatalf("expected ErrKeyExists, got %v", err)
	}
}

func TestShardedReplace(t *testing.T) {
	sc := NewSharded[string, string](4)
	err := sc.Replace("key", "val", 0)
	if !errors.Is(err, ErrKeyNotFound) {
		t.Fatal("expected ErrKeyNotFound")
	}
	sc.Set("key", "old", 0)
	err = sc.Replace("key", "new", 0)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := sc.Get("key")
	if v != "new" {
		t.Fatalf("expected new, got %q", v)
	}
}

func TestShardedGetWithExpiration(t *testing.T) {
	sc := NewSharded[string, string](4)
	sc.Set("key", "val", time.Hour)
	v, exp, ok := sc.GetWithExpiration("key")
	if !ok || v != "val" || exp.IsZero() {
		t.Fatal("expected found with expiration")
	}
}

func TestShardedGetOrSet(t *testing.T) {
	sc := NewSharded[string, string](4)
	v, err := sc.GetOrSet("key", func() (string, error) {
		return "computed", nil
	}, 0)
	if err != nil || v != "computed" {
		t.Fatal("expected computed value")
	}
	v2, _ := sc.Get("key")
	if v2 != "computed" {
		t.Fatal("expected value to be stored")
	}
}

func TestShardedItems(t *testing.T) {
	sc := NewSharded[string, int](4)
	sc.Set("a", 1, 0)
	sc.Set("b", 2, 0)
	sc.Set("c", 3, 0)
	items := sc.Items()
	if len(items) != 3 {
		t.Fatalf("expected 3, got %d", len(items))
	}
}

func TestShardedFlush(t *testing.T) {
	sc := NewSharded[string, int](4)
	sc.Set("a", 1, 0)
	sc.Set("b", 2, 0)
	sc.Flush()
	if sc.Count() != 0 {
		t.Fatal("expected 0 after flush")
	}
}

func TestShardedCount(t *testing.T) {
	sc := NewSharded[string, int](4)
	for i := 0; i < 100; i++ {
		sc.Set(fmt.Sprintf("key%d", i), i, 0)
	}
	if sc.Count() != 100 {
		t.Fatalf("expected 100, got %d", sc.Count())
	}
}

func TestShardedDeleteExpired(t *testing.T) {
	sc := NewSharded[string, string](4)
	sc.Set("keep", "v", 0)
	sc.Set("expire", "v", 10*time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	sc.DeleteExpired()
	if sc.Count() != 1 {
		t.Fatalf("expected 1, got %d", sc.Count())
	}
}

func TestShardedStop(t *testing.T) {
	sc := NewSharded[string, string](4,
		WithCleanupInterval[string, string](10*time.Millisecond),
	)
	sc.Stop() // should not panic
}

func TestShardedConcurrentAccess(t *testing.T) {
	sc := NewSharded[string, int](16)
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", n%20)
			sc.Set(key, n, time.Second)
			sc.Get(key)
			sc.Delete(fmt.Sprintf("key%d", (n+10)%20))
		}(i)
	}
	wg.Wait()
}

func TestShardedDefaultShards(t *testing.T) {
	sc := NewSharded[string, string](0) // 0 should default to 16
	sc.Set("key", "val", 0)
	v, ok := sc.Get("key")
	if !ok || v != "val" {
		t.Fatal("expected default shard count to work")
	}
}

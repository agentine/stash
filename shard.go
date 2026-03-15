package stash

import (
	"fmt"
	"hash/fnv"
	"time"
)

// ShardedCache is a thread-safe cache that distributes keys across multiple
// Cache shards to reduce lock contention under high concurrency.
type ShardedCache[K comparable, V any] struct {
	shards []*Cache[K, V]
	count  int
}

// NewSharded creates a new ShardedCache with the given number of shards.
// Each shard is an independent Cache with the provided options.
func NewSharded[K comparable, V any](shards int, opts ...Option[K, V]) *ShardedCache[K, V] {
	if shards <= 0 {
		shards = 16
	}
	sc := &ShardedCache[K, V]{
		shards: make([]*Cache[K, V], shards),
		count:  shards,
	}
	for i := range sc.shards {
		sc.shards[i] = New(opts...)
	}
	return sc
}

func (sc *ShardedCache[K, V]) getShard(key K) *Cache[K, V] {
	h := fnv.New64a()
	_, _ = fmt.Fprintf(h, "%v", key)
	return sc.shards[h.Sum64()%uint64(sc.count)]
}

// Get returns the value for the given key.
func (sc *ShardedCache[K, V]) Get(key K) (V, bool) {
	return sc.getShard(key).Get(key)
}

// GetWithExpiration returns the value and its expiration time.
func (sc *ShardedCache[K, V]) GetWithExpiration(key K) (V, time.Time, bool) {
	return sc.getShard(key).GetWithExpiration(key)
}

// Set adds or updates an item.
func (sc *ShardedCache[K, V]) Set(key K, val V, ttl time.Duration) {
	sc.getShard(key).Set(key, val, ttl)
}

// SetDefault adds or updates an item using the default TTL.
func (sc *ShardedCache[K, V]) SetDefault(key K, val V) {
	sc.getShard(key).SetDefault(key, val)
}

// Add sets the item only if the key does not already exist.
func (sc *ShardedCache[K, V]) Add(key K, val V, ttl time.Duration) error {
	return sc.getShard(key).Add(key, val, ttl)
}

// Replace sets the item only if the key already exists.
func (sc *ShardedCache[K, V]) Replace(key K, val V, ttl time.Duration) error {
	return sc.getShard(key).Replace(key, val, ttl)
}

// Delete removes an item.
func (sc *ShardedCache[K, V]) Delete(key K) {
	sc.getShard(key).Delete(key)
}

// GetOrSet returns an existing value or computes and stores a new one.
func (sc *ShardedCache[K, V]) GetOrSet(key K, fn func() (V, error), ttl time.Duration) (V, error) {
	return sc.getShard(key).GetOrSet(key, fn, ttl)
}

// Items returns a merged copy of all unexpired items across all shards.
func (sc *ShardedCache[K, V]) Items() map[K]Item[V] {
	m := make(map[K]Item[V])
	for _, shard := range sc.shards {
		for k, v := range shard.Items() {
			m[k] = v
		}
	}
	return m
}

// Flush removes all items from all shards.
func (sc *ShardedCache[K, V]) Flush() {
	for _, shard := range sc.shards {
		shard.Flush()
	}
}

// Count returns the total number of items across all shards.
func (sc *ShardedCache[K, V]) Count() int {
	n := 0
	for _, shard := range sc.shards {
		n += shard.Count()
	}
	return n
}

// DeleteExpired removes expired items from all shards.
func (sc *ShardedCache[K, V]) DeleteExpired() {
	for _, shard := range sc.shards {
		shard.DeleteExpired()
	}
}

// Stop stops the background janitor on all shards.
func (sc *ShardedCache[K, V]) Stop() {
	for _, shard := range sc.shards {
		shard.Stop()
	}
}

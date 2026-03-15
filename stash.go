// Package stash provides a type-safe, generics-based in-memory cache with TTL
// expiration, eviction policies, and sharded concurrency support.
// It is a drop-in replacement for patrickmn/go-cache.
package stash

import (
	"sync"
	"time"
)

// Cache is a thread-safe in-memory key-value cache with expiration support.
type Cache[K comparable, V any] struct {
	mu             sync.RWMutex
	items          map[K]Item[V]
	defaultTTL     time.Duration
	maxSize        int
	evictionPolicy EvictionPolicy
	evictor        evictor[K]
	onEvicted      func(K, V)
	janitor        *janitor[K, V]
}

// evictor is the internal interface for eviction policies.
type evictor[K comparable] interface {
	Access(key K)
	Add(key K)
	Remove(key K)
	Evict() K
}

// New creates a new Cache with the given options.
func New[K comparable, V any](opts ...Option[K, V]) *Cache[K, V] {
	cfg := config[K, V]{}
	for _, opt := range opts {
		opt(&cfg)
	}

	c := &Cache[K, V]{
		items:          make(map[K]Item[V]),
		defaultTTL:     cfg.defaultTTL,
		maxSize:        cfg.maxSize,
		evictionPolicy: cfg.evictionPolicy,
		onEvicted:      cfg.onEvicted,
	}

	if cfg.cleanupInterval > 0 {
		c.janitor = newJanitor(c, cfg.cleanupInterval)
	}

	return c
}

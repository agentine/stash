package stash

import "time"

// DefaultTTL is a sentinel value indicating that the cache's default TTL should be used.
const DefaultTTL time.Duration = -1

// EvictionPolicy defines the eviction strategy when the cache reaches max size.
type EvictionPolicy int

const (
	// NoEviction means no eviction policy is applied; the cache grows unbounded.
	NoEviction EvictionPolicy = iota
	// LRU evicts the least recently used item.
	LRU
	// LFU evicts the least frequently used item.
	LFU
)

// config holds the internal configuration for a Cache.
type config[K comparable, V any] struct {
	defaultTTL      time.Duration
	cleanupInterval time.Duration
	maxSize         int
	evictionPolicy  EvictionPolicy
	onEvicted       func(K, V)
}

// Option is a functional option for configuring a Cache.
type Option[K comparable, V any] func(*config[K, V])

// WithDefaultTTL sets the default TTL for items added without an explicit TTL.
// A value of 0 means items never expire by default.
func WithDefaultTTL[K comparable, V any](d time.Duration) Option[K, V] {
	return func(c *config[K, V]) {
		c.defaultTTL = d
	}
}

// WithCleanupInterval sets the interval for the background janitor to remove expired items.
// A value of 0 disables the janitor.
func WithCleanupInterval[K comparable, V any](d time.Duration) Option[K, V] {
	return func(c *config[K, V]) {
		c.cleanupInterval = d
	}
}

// WithMaxSize sets the maximum number of items in the cache.
// When the limit is reached, the configured eviction policy determines which item is removed.
// A value of 0 means no limit.
func WithMaxSize[K comparable, V any](n int) Option[K, V] {
	return func(c *config[K, V]) {
		c.maxSize = n
	}
}

// WithEviction sets the eviction policy for the cache.
func WithEviction[K comparable, V any](policy EvictionPolicy) Option[K, V] {
	return func(c *config[K, V]) {
		c.evictionPolicy = policy
	}
}

// WithOnEvicted sets a callback function that is called when an item is evicted from the cache.
func WithOnEvicted[K comparable, V any](fn func(K, V)) Option[K, V] {
	return func(c *config[K, V]) {
		c.onEvicted = fn
	}
}

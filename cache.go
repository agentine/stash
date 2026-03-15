package stash

import (
	"errors"
	"time"
)

var (
	// ErrKeyExists is returned by Add when the key already exists and has not expired.
	ErrKeyExists = errors.New("stash: item already exists")
	// ErrKeyNotFound is returned by Replace when the key does not exist or has expired.
	ErrKeyNotFound = errors.New("stash: item not found")
)

// Get returns the value for the given key and whether it was found.
// Expired items are treated as missing.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	if !found {
		c.mu.RUnlock()
		var zero V
		return zero, false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		c.mu.RUnlock()
		var zero V
		return zero, false
	}
	c.mu.RUnlock()
	// Track access for eviction outside write lock only if evictor is set.
	if c.evictor != nil {
		c.mu.Lock()
		c.evictor.Access(key)
		c.mu.Unlock()
	}
	return item.Value, true
}

// GetWithExpiration returns the value, its expiration time, and whether it was found.
// If the item has no expiration, the returned time is the zero value.
func (c *Cache[K, V]) GetWithExpiration(key K) (V, time.Time, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	if !found {
		c.mu.RUnlock()
		var zero V
		return zero, time.Time{}, false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		c.mu.RUnlock()
		var zero V
		return zero, time.Time{}, false
	}
	c.mu.RUnlock()
	if c.evictor != nil {
		c.mu.Lock()
		c.evictor.Access(key)
		c.mu.Unlock()
	}
	var exp time.Time
	if item.Expiration > 0 {
		exp = time.Unix(0, item.Expiration)
	}
	return item.Value, exp, true
}

// Set adds or updates an item in the cache with the given TTL.
// Use DefaultTTL to use the cache's default TTL. Use 0 for no expiration.
func (c *Cache[K, V]) Set(key K, val V, ttl time.Duration) {
	var exp int64
	d := ttl
	if d == DefaultTTL {
		d = c.defaultTTL
	}
	if d > 0 {
		exp = time.Now().Add(d).UnixNano()
	}

	c.mu.Lock()
	_, exists := c.items[key]
	if !exists && c.evictor != nil {
		c.evictIfNeeded()
		c.evictor.Add(key)
	} else if exists && c.evictor != nil {
		c.evictor.Access(key)
	}
	c.items[key] = Item[V]{Value: val, Expiration: exp}
	c.mu.Unlock()
}

// SetDefault adds or updates an item using the cache's default TTL.
func (c *Cache[K, V]) SetDefault(key K, val V) {
	c.Set(key, val, DefaultTTL)
}

// Add sets the item only if the key does not already exist (or has expired).
// Returns ErrKeyExists if the key is present and not expired.
func (c *Cache[K, V]) Add(key K, val V, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if item, found := c.items[key]; found {
		if item.Expiration == 0 || time.Now().UnixNano() <= item.Expiration {
			return ErrKeyExists
		}
	}
	var exp int64
	d := ttl
	if d == DefaultTTL {
		d = c.defaultTTL
	}
	if d > 0 {
		exp = time.Now().Add(d).UnixNano()
	}
	if c.evictor != nil {
		c.evictIfNeeded()
		c.evictor.Add(key)
	}
	c.items[key] = Item[V]{Value: val, Expiration: exp}
	return nil
}

// Replace sets the item only if the key already exists and has not expired.
// Returns ErrKeyNotFound if the key is missing or expired.
func (c *Cache[K, V]) Replace(key K, val V, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	item, found := c.items[key]
	if !found {
		return ErrKeyNotFound
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return ErrKeyNotFound
	}
	var exp int64
	d := ttl
	if d == DefaultTTL {
		d = c.defaultTTL
	}
	if d > 0 {
		exp = time.Now().Add(d).UnixNano()
	}
	if c.evictor != nil {
		c.evictor.Access(key)
	}
	c.items[key] = Item[V]{Value: val, Expiration: exp}
	return nil
}

// Delete removes an item from the cache and fires the onEvicted callback if set.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	item, found := c.items[key]
	if found {
		delete(c.items, key)
		if c.evictor != nil {
			c.evictor.Remove(key)
		}
	}
	c.mu.Unlock()
	if found && c.onEvicted != nil {
		c.onEvicted(key, item.Value)
	}
}

// GetOrSet returns the existing value for the key if present and not expired.
// Otherwise, it calls fn to compute the value, stores it, and returns it.
func (c *Cache[K, V]) GetOrSet(key K, fn func() (V, error), ttl time.Duration) (V, error) {
	if val, ok := c.Get(key); ok {
		return val, nil
	}
	val, err := fn()
	if err != nil {
		var zero V
		return zero, err
	}
	c.Set(key, val, ttl)
	return val, nil
}

// Items returns a copy of all unexpired items in the cache.
func (c *Cache[K, V]) Items() map[K]Item[V] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	now := time.Now().UnixNano()
	m := make(map[K]Item[V], len(c.items))
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			continue
		}
		m[k] = v
	}
	return m
}

// Flush removes all items from the cache.
func (c *Cache[K, V]) Flush() {
	c.mu.Lock()
	c.items = make(map[K]Item[V])
	c.mu.Unlock()
}

// Count returns the number of items in the cache, including expired items
// that have not yet been cleaned up.
func (c *Cache[K, V]) Count() int {
	c.mu.RLock()
	n := len(c.items)
	c.mu.RUnlock()
	return n
}

// DeleteExpired removes all expired items from the cache.
func (c *Cache[K, V]) DeleteExpired() {
	c.mu.Lock()
	now := time.Now().UnixNano()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.items, k)
			if c.evictor != nil {
				c.evictor.Remove(k)
			}
			if c.onEvicted != nil {
				c.onEvicted(k, v.Value)
			}
		}
	}
	c.mu.Unlock()
}

// Stop stops the background janitor, if running.
func (c *Cache[K, V]) Stop() {
	if c.janitor != nil {
		close(c.janitor.stop)
		c.janitor = nil
	}
}

// evictIfNeeded evicts one item if maxSize is set and reached.
// Must be called with c.mu held.
func (c *Cache[K, V]) evictIfNeeded() {
	if c.maxSize <= 0 || len(c.items) < c.maxSize {
		return
	}
	key := c.evictor.Evict()
	if item, ok := c.items[key]; ok {
		delete(c.items, key)
		if c.onEvicted != nil {
			c.onEvicted(key, item.Value)
		}
	}
}

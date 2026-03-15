package stash

import "time"

// DeleteExpired removes all expired items from the cache.
func (c *Cache[K, V]) DeleteExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()
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
}

// Stop stops the background janitor, if running.
func (c *Cache[K, V]) Stop() {
	if c.janitor != nil {
		close(c.janitor.stop)
		c.janitor = nil
	}
}

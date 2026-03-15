package stash

import (
	"runtime"
	"time"
)

type janitor[K comparable, V any] struct {
	interval time.Duration
	stop     chan struct{}
}

func newJanitor[K comparable, V any](c *Cache[K, V], interval time.Duration) *janitor[K, V] {
	j := &janitor[K, V]{
		interval: interval,
		stop:     make(chan struct{}),
	}
	go j.run(c)
	// Stop the janitor when the cache is garbage collected.
	runtime.SetFinalizer(c, func(c *Cache[K, V]) {
		c.Stop()
	})
	return j
}

func (j *janitor[K, V]) run(c *Cache[K, V]) {
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-j.stop:
			return
		}
	}
}

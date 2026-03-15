package stash

import "time"

// Item represents a cache item with a value and optional expiration.
type Item[V any] struct {
	Value      V
	Expiration int64 // UnixNano timestamp; 0 means no expiry
}

// Expired returns true if the item has expired.
func (item Item[V]) Expired() bool {
	if item.Expiration == 0 {
		return false
	}
	return time.Now().UnixNano() > item.Expiration
}

package stash

import (
	"fmt"
	"time"
)

// UntypedCache is a Cache[string, any] that provides API compatibility with
// patrickmn/go-cache. Values are stored as any and must be type-asserted by
// the caller.
type UntypedCache struct {
	*Cache[string, any]
}

// NewUntyped creates a new UntypedCache matching the go-cache constructor
// signature: cache.New(defaultExpiration, cleanupInterval).
func NewUntyped(defaultExpiration, cleanupInterval time.Duration) *UntypedCache {
	var opts []Option[string, any]
	if defaultExpiration > 0 {
		opts = append(opts, WithDefaultTTL[string, any](defaultExpiration))
	}
	if cleanupInterval > 0 {
		opts = append(opts, WithCleanupInterval[string, any](cleanupInterval))
	}
	return &UntypedCache{Cache: New(opts...)}
}

// Increment atomically increments a numeric value by n.
// The item must exist, not be expired, and hold a supported numeric type.
func (uc *UntypedCache) Increment(key string, n int64) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	item, found := uc.items[key]
	if !found {
		return fmt.Errorf("stash.Increment: item %s not found", key)
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return fmt.Errorf("stash.Increment: item %s not found", key)
	}
	val, err := incrNumeric(item.Value, n)
	if err != nil {
		return err
	}
	item.Value = val
	uc.items[key] = item
	return nil
}

// Decrement atomically decrements a numeric value by n.
// The item must exist, not be expired, and hold a supported numeric type.
func (uc *UntypedCache) Decrement(key string, n int64) error {
	uc.mu.Lock()
	defer uc.mu.Unlock()
	item, found := uc.items[key]
	if !found {
		return fmt.Errorf("stash.Decrement: item %s not found", key)
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		return fmt.Errorf("stash.Decrement: item %s not found", key)
	}
	val, err := decrNumeric(item.Value, n)
	if err != nil {
		return err
	}
	item.Value = val
	uc.items[key] = item
	return nil
}

func incrNumeric(v any, n int64) (any, error) {
	switch val := v.(type) {
	case int:
		return val + int(n), nil
	case int8:
		return val + int8(n), nil
	case int16:
		return val + int16(n), nil
	case int32:
		return val + int32(n), nil
	case int64:
		return val + n, nil
	case uint:
		return val + uint(n), nil
	case uint8:
		return val + uint8(n), nil
	case uint16:
		return val + uint16(n), nil
	case uint32:
		return val + uint32(n), nil
	case uint64:
		return val + uint64(n), nil
	case float32:
		return val + float32(n), nil
	case float64:
		return val + float64(n), nil
	default:
		return nil, fmt.Errorf("stash.Increment: value for is not a numeric type")
	}
}

func decrNumeric(v any, n int64) (any, error) {
	switch val := v.(type) {
	case int:
		return val - int(n), nil
	case int8:
		return val - int8(n), nil
	case int16:
		return val - int16(n), nil
	case int32:
		return val - int32(n), nil
	case int64:
		return val - n, nil
	case uint:
		return val - uint(n), nil
	case uint8:
		return val - uint8(n), nil
	case uint16:
		return val - uint16(n), nil
	case uint32:
		return val - uint32(n), nil
	case uint64:
		return val - uint64(n), nil
	case float32:
		return val - float32(n), nil
	case float64:
		return val - float64(n), nil
	default:
		return nil, fmt.Errorf("stash.Decrement: value is not a numeric type")
	}
}

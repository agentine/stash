# stash — Implementation Plan

**Replaces:** [patrickmn/go-cache](https://github.com/patrickmn/go-cache)
**Package:** `github.com/agentine/stash`
**Language:** Go (requires 1.22+)

## Why

patrickmn/go-cache has 7,109 importers and 8,804 stars but has been unmaintained since 2017 (last release) / 2019 (last commit). Single maintainer (patrickmn), 76 open issues, no PRs reviewed. Uses `interface{}` throughout — predates Go generics. No eviction policies, no size limits, no sharding for concurrent workloads.

Existing alternatives (ristretto, bigcache, ttlcache) serve different niches and none provide a drop-in compatible API.

## Scope

A type-safe, generics-based in-memory cache with TTL expiration, drop-in API compatibility with go-cache, and modern improvements.

## Architecture

### Core Package: `stash`

```
stash/
├── cache.go          # Cache[K, V] generic type, core Get/Set/Delete
├── options.go        # functional options (WithTTL, WithMaxSize, etc.)
├── eviction.go       # eviction policies (LRU, LFU, none)
├── shard.go          # sharded cache for reduced lock contention
├── janitor.go        # background cleanup of expired items
├── item.go           # cache item with expiration metadata
├── stash.go          # package-level convenience (New, NewSharded)
├── compat.go         # go-cache compatible untyped API (map[string]any)
├── doc.go            # package documentation
└── *_test.go         # tests for each component
```

### Key Design Decisions

1. **Generics-first:** `Cache[K comparable, V any]` — type-safe access, no casting.
2. **go-cache compatibility layer:** `compat.go` provides `UntypedCache` with `map[string]any` API matching go-cache for migration.
3. **Functional options:** `New[K, V](opts ...Option[K, V])` pattern.
4. **Eviction policies:** Optional LRU/LFU eviction when max size is set. Default: no eviction (matches go-cache behavior).
5. **Sharding:** `NewSharded[K, V](shards int, opts ...Option[K, V])` for high-concurrency workloads. Configurable shard count.
6. **Background cleanup:** Configurable janitor interval for expired item removal (matches go-cache's cleanup goroutine pattern).

### API Surface

```go
// Generic API
c := stash.New[string, User](
    stash.WithDefaultTTL[string, User](5 * time.Minute),
    stash.WithCleanupInterval[string, User](10 * time.Minute),
    stash.WithMaxSize[string, User](10000),
    stash.WithEviction[string, User](stash.LRU),
)

c.Set("user:1", user, stash.DefaultTTL)
user, found := c.Get("user:1")
c.Delete("user:1")

// Atomic operations
c.GetOrSet("user:1", func() (User, error) { return fetchUser(1) }, 5*time.Minute)

// Bulk operations
c.Items() map[string]stash.Item[User]
c.Flush()
c.Count() int

// Increment/Decrement (numeric V only — compile-time constraint)
nc := stash.New[string, int64](...)
nc.Increment("hits", 1)
nc.Decrement("hits", 1)

// Sharded cache
sc := stash.NewSharded[string, User](16,
    stash.WithDefaultTTL[string, User](5 * time.Minute),
)

// go-cache compatibility (untyped)
uc := stash.NewUntyped(5*time.Minute, 10*time.Minute)
uc.Set("foo", "bar", stash.DefaultTTL)
val, found := uc.Get("foo") // val is any
```

## Major Components

### 1. Core Cache (`cache.go`)
- `Cache[K comparable, V any]` struct with `sync.RWMutex`
- Get, Set, SetDefault, Add (set-if-absent), Replace (set-if-present), Delete
- GetWithExpiration returns value + expiration time
- OnEvicted callback support

### 2. Eviction Policies (`eviction.go`)
- Interface: `Evictor[K]` with `Access(key)`, `Add(key)`, `Remove(key)`, `Evict() K`
- LRU implementation using doubly-linked list
- LFU implementation using min-heap
- Default: no eviction (unbounded, matches go-cache)

### 3. Sharding (`shard.go`)
- `ShardedCache[K, V]` wrapping N `Cache[K, V]` instances
- FNV-based key hashing for shard selection
- Same API surface as `Cache` via shared interface

### 4. Janitor (`janitor.go`)
- Background goroutine sweeping expired items
- Configurable interval, stoppable
- Weak reference to cache (GC-friendly, matches go-cache pattern)

### 5. Compatibility Layer (`compat.go`)
- `UntypedCache` = `Cache[string, any]` with go-cache method signatures
- `NewUntyped(defaultExpiration, cleanupInterval)` matches `cache.New()`
- Increment/Decrement with runtime type assertions (matching go-cache behavior)
- Save/Load serialization (JSON)

## Deliverables

1. Core generic cache with TTL support
2. LRU and LFU eviction policies
3. Sharded cache for concurrent access
4. go-cache compatible untyped API
5. Comprehensive test suite with benchmarks
6. README with migration guide from go-cache

## Verified

- Package name `github.com/agentine/stash` is available on pkg.go.dev
- MIT license (same as go-cache)

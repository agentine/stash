package stash

import "container/heap"

// --- noop evictor (unbounded mode) ---

type noopEvictor[K comparable] struct{}

func (noopEvictor[K]) Access(K) {}
func (noopEvictor[K]) Add(K)    {}
func (noopEvictor[K]) Remove(K) {}
func (e noopEvictor[K]) Evict() K {
	var zero K
	return zero
}

// --- LRU evictor (doubly-linked list) ---

type lruNode[K comparable] struct {
	key        K
	prev, next *lruNode[K]
}

type lruEvictor[K comparable] struct {
	head, tail *lruNode[K]
	index      map[K]*lruNode[K]
}

func newLRUEvictor[K comparable]() *lruEvictor[K] {
	head := &lruNode[K]{}
	tail := &lruNode[K]{}
	head.next = tail
	tail.prev = head
	return &lruEvictor[K]{
		head:  head,
		tail:  tail,
		index: make(map[K]*lruNode[K]),
	}
}

func (l *lruEvictor[K]) moveToFront(node *lruNode[K]) {
	// Remove from current position.
	node.prev.next = node.next
	node.next.prev = node.prev
	// Insert after head sentinel.
	node.next = l.head.next
	node.prev = l.head
	l.head.next.prev = node
	l.head.next = node
}

func (l *lruEvictor[K]) Access(key K) {
	if node, ok := l.index[key]; ok {
		l.moveToFront(node)
	}
}

func (l *lruEvictor[K]) Add(key K) {
	if node, ok := l.index[key]; ok {
		l.moveToFront(node)
		return
	}
	node := &lruNode[K]{key: key}
	l.index[key] = node
	// Insert at front (most recently used).
	node.next = l.head.next
	node.prev = l.head
	l.head.next.prev = node
	l.head.next = node
}

func (l *lruEvictor[K]) Remove(key K) {
	node, ok := l.index[key]
	if !ok {
		return
	}
	node.prev.next = node.next
	node.next.prev = node.prev
	delete(l.index, key)
}

func (l *lruEvictor[K]) Evict() K {
	// Evict from back (least recently used), just before tail sentinel.
	node := l.tail.prev
	if node == l.head {
		var zero K
		return zero
	}
	l.Remove(node.key)
	return node.key
}

// --- LFU evictor (min-heap by frequency) ---

type lfuEntry[K comparable] struct {
	key   K
	freq  int
	index int // position in heap
}

type lfuHeap[K comparable] []*lfuEntry[K]

func (h lfuHeap[K]) Len() int            { return len(h) }
func (h lfuHeap[K]) Less(i, j int) bool   { return h[i].freq < h[j].freq }
func (h lfuHeap[K]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *lfuHeap[K]) Push(x any) {
	entry := x.(*lfuEntry[K])
	entry.index = len(*h)
	*h = append(*h, entry)
}

func (h *lfuHeap[K]) Pop() any {
	old := *h
	n := len(old)
	entry := old[n-1]
	old[n-1] = nil
	entry.index = -1
	*h = old[:n-1]
	return entry
}

type lfuEvictor[K comparable] struct {
	h     lfuHeap[K]
	index map[K]*lfuEntry[K]
}

func newLFUEvictor[K comparable]() *lfuEvictor[K] {
	return &lfuEvictor[K]{
		index: make(map[K]*lfuEntry[K]),
	}
}

func (l *lfuEvictor[K]) Access(key K) {
	if entry, ok := l.index[key]; ok {
		entry.freq++
		heap.Fix(&l.h, entry.index)
	}
}

func (l *lfuEvictor[K]) Add(key K) {
	if entry, ok := l.index[key]; ok {
		entry.freq++
		heap.Fix(&l.h, entry.index)
		return
	}
	entry := &lfuEntry[K]{key: key, freq: 1}
	l.index[key] = entry
	heap.Push(&l.h, entry)
}

func (l *lfuEvictor[K]) Remove(key K) {
	entry, ok := l.index[key]
	if !ok {
		return
	}
	heap.Remove(&l.h, entry.index)
	delete(l.index, key)
}

func (l *lfuEvictor[K]) Evict() K {
	if l.h.Len() == 0 {
		var zero K
		return zero
	}
	entry := heap.Pop(&l.h).(*lfuEntry[K])
	delete(l.index, entry.key)
	return entry.key
}

// newEvictor creates the appropriate evictor for the given policy.
func newEvictor[K comparable](policy EvictionPolicy) evictor[K] {
	switch policy {
	case LRU:
		return newLRUEvictor[K]()
	case LFU:
		return newLFUEvictor[K]()
	default:
		return noopEvictor[K]{}
	}
}

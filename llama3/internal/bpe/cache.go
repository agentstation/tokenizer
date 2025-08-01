package bpe

import (
	"container/list"
	"sync"
)

// Cache is the interface for caching BPE results.
type Cache interface {
	Get(key string) ([]int, bool)
	Put(key string, value []int)
}

// lruCache implements a thread-safe LRU cache for BPE results.
type LRUCache struct {
	capacity int
	items    map[string]*list.Element
	lru      *list.List
	mu       sync.RWMutex
}

// cacheEntry holds a cache entry with key and value.
type cacheEntry struct {
	key   string
	value []int
}

// NewLRU creates a new LRU cache with the given capacity.
// If capacity is 0, the cache is unlimited (falls back to simple map).
func NewLRU(capacity int) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		lru:      list.New(),
	}
}

// Get retrieves a value from the cache, promoting it to most recently used.
func (c *LRUCache) Get(key string) ([]int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		// Move to front (most recently used)
		c.lru.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return nil, false
}

// Put adds or updates a value in the cache.
func (c *LRUCache) Put(key string, value []int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If key exists, update and move to front
	if elem, ok := c.items[key]; ok {
		c.lru.MoveToFront(elem)
		elem.Value.(*cacheEntry).value = value
		return
	}

	// Add new entry
	entry := &cacheEntry{key: key, value: value}
	elem := c.lru.PushFront(entry)
	c.items[key] = elem

	// Evict oldest if over capacity (capacity 0 means unlimited)
	if c.capacity > 0 && c.lru.Len() > c.capacity {
		oldest := c.lru.Back()
		if oldest != nil {
			c.lru.Remove(oldest)
			delete(c.items, oldest.Value.(*cacheEntry).key)
		}
	}
}

// SimpleCache wraps a regular map for unlimited caching (backward compatibility).
type SimpleCache struct {
	cache map[string][]int
	mu    sync.RWMutex
}

// NewSimple creates a new simple cache.
func NewSimple() *SimpleCache {
	return &SimpleCache{
		cache: make(map[string][]int),
	}
}

// Get retrieves a value from the simple cache.
func (c *SimpleCache) Get(key string) ([]int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.cache[key]
	return val, ok
}

// Put adds a value to the simple cache.
func (c *SimpleCache) Put(key string, value []int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = value
}

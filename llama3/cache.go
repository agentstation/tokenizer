package llama3

import (
	"container/list"
	"sync"
)

// lruCache implements a thread-safe LRU cache for BPE results.
type lruCache struct {
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

// newLRUCache creates a new LRU cache with the given capacity.
// If capacity is 0, the cache is unlimited (falls back to simple map).
func newLRUCache(capacity int) *lruCache {
	return &lruCache{
		capacity: capacity,
		items:    make(map[string]*list.Element),
		lru:      list.New(),
	}
}

// get retrieves a value from the cache, promoting it to most recently used.
func (c *lruCache) get(key string) ([]int, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, ok := c.items[key]; ok {
		// Move to front (most recently used)
		c.lru.MoveToFront(elem)
		return elem.Value.(*cacheEntry).value, true
	}
	return nil, false
}

// put adds or updates a value in the cache.
func (c *lruCache) put(key string, value []int) {
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

// simpleCache wraps a regular map for unlimited caching (backward compatibility).
type simpleCache struct {
	cache map[string][]int
	mu    sync.RWMutex
}

// get retrieves a value from the simple cache.
func (c *simpleCache) get(key string) ([]int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.cache[key]
	return val, ok
}

// put adds a value to the simple cache.
func (c *simpleCache) put(key string, value []int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[key] = value
}
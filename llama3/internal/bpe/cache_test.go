package bpe

import (
	"testing"
)

func TestLRUCache(t *testing.T) {
	t.Run("basic_operations", func(t *testing.T) {
		cache := NewLRU(3)
		
		// Test put and get
		cache.Put("key1", []int{1, 2, 3})
		cache.Put("key2", []int{4, 5, 6})
		cache.Put("key3", []int{7, 8, 9})
		
		// Verify all keys exist
		if _, ok := cache.Get("key1"); !ok {
			t.Error("Expected key1 to exist")
		}
		if _, ok := cache.Get("key2"); !ok {
			t.Error("Expected key2 to exist")
		}
		if _, ok := cache.Get("key3"); !ok {
			t.Error("Expected key3 to exist")
		}
		
		// Add fourth key, should evict key1 (LRU)
		cache.Put("key4", []int{10, 11, 12})
		
		if _, ok := cache.Get("key1"); ok {
			t.Error("Expected key1 to be evicted")
		}
		
		if _, ok := cache.Get("key4"); !ok {
			t.Error("Expected key4 to exist")
		}
	})
	
	t.Run("lru_ordering", func(t *testing.T) {
		cache := NewLRU(2)
		
		cache.Put("a", []int{1})
		cache.Put("b", []int{2})
		
		// Access 'a' to make it more recently used
		cache.Get("a")
		
		// Add 'c', should evict 'b' (least recently used)
		cache.Put("c", []int{3})
		
		if _, ok := cache.Get("a"); !ok {
			t.Error("Expected 'a' to exist (was accessed)")
		}
		
		if _, ok := cache.Get("b"); ok {
			t.Error("Expected 'b' to be evicted (LRU)")
		}
		
		if _, ok := cache.Get("c"); !ok {
			t.Error("Expected 'c' to exist (just added)")
		}
	})
	
	t.Run("update_existing", func(t *testing.T) {
		cache := NewLRU(2)
		
		cache.Put("key", []int{1, 2})
		cache.Put("key", []int{3, 4}) // Update
		
		val, ok := cache.Get("key")
		if !ok {
			t.Fatal("Expected key to exist")
		}
		
		if len(val) != 2 || val[0] != 3 || val[1] != 4 {
			t.Errorf("Expected updated value [3,4], got %v", val)
		}
	})
	
	t.Run("unlimited_cache", func(t *testing.T) {
		cache := NewLRU(0) // Unlimited
		
		// Add many items
		for i := 0; i < 100; i++ {
			cache.Put(string(rune('a'+i)), []int{i})
		}
		
		// All should still exist
		for i := 0; i < 100; i++ {
			if _, ok := cache.Get(string(rune('a' + i))); !ok {
				t.Errorf("Expected key %c to exist in unlimited cache", 'a'+i)
			}
		}
	})
}

func TestSimpleCache(t *testing.T) {
	cache := NewSimple()
	
	// Test put and get
	cache.Put("key1", []int{1, 2, 3})
	
	val, ok := cache.Get("key1")
	if !ok {
		t.Fatal("Expected key1 to exist")
	}
	
	if len(val) != 3 || val[0] != 1 {
		t.Errorf("Expected [1,2,3], got %v", val)
	}
	
	// Test missing key
	_, ok = cache.Get("missing")
	if ok {
		t.Error("Expected missing key to not exist")
	}
}


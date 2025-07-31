// Package llama3 memory pool documentation.
//
// This file documents the memory pooling strategies used in the llama3 tokenizer
// for optimal performance and memory efficiency.

package llama3

// Pool Usage Patterns
//
// The llama3 tokenizer uses sync.Pool for two primary purposes:
//
// 1. State Machine Pooling (stateMachinePool)
//    - Reuses StateMachine instances across tokenization calls
//    - Reduces allocations for the input rune slice and token slice
//    - Pool never limits the number of state machines
//    - State machines are reset before reuse
//
// 2. Token Buffer Pooling (tokenBufPool)
//    - Reuses []string slices for collecting tokens
//    - Initial capacity: 64 tokens (defaultTokenBufferCapacity)
//    - Maximum pooled capacity: 1024 tokens (maxPooledTokenBufferCapacity)
//    - Buffers exceeding the maximum are not returned to the pool
//
// Pool Behavior and Limits
//
// State Machine Pool:
//   - No upper limit on pool size
//   - Automatically managed by Go runtime
//   - State machines are cleared of references before returning to pool
//   - Input and token slices are reset but capacity is preserved
//
// Token Buffer Pool:
//   - No upper limit on pool size
//   - Buffers start with capacity 64
//   - Only buffers with capacity <= 1024 are returned to pool
//   - This prevents memory bloat from unusually large tokenizations
//
// Thread Safety
//
// All pools are thread-safe and can be used concurrently:
//   - Multiple goroutines can tokenize text simultaneously
//   - Each gets its own state machine and token buffer from the pools
//   - No locks are needed beyond what sync.Pool provides internally
//
// Memory Lifecycle
//
// 1. Allocation:
//    - First call to pool.Get() creates new instance
//    - Subsequent calls may return pooled instance or create new
//
// 2. Usage:
//    - Instance is used for one tokenization operation
//    - Results are copied before returning instance to pool
//
// 3. Return to Pool:
//    - References are cleared to allow garbage collection
//    - Slices are reset to length 0 but capacity preserved
//    - Large buffers (>1024 capacity) are discarded
//
// 4. Garbage Collection:
//    - Go runtime may clear pools during GC
//    - This prevents unbounded memory growth
//    - Pool will repopulate as needed
//
// Performance Characteristics
//
// Benchmarks show 36% memory reduction with pooling:
//   - Small text (13 chars): 160B → 120B allocated
//   - Large text (1.3KB): 44KB → 28KB allocated
//   - Performance remains comparable or slightly better
//
// Best Practices
//
// When modifying the tokenizer:
//   1. Always clear references before returning to pool
//   2. Copy results before returning buffers to pool
//   3. Consider capacity limits to prevent memory bloat
//   4. Test with concurrent usage to ensure thread safety
//
// Example Usage Pattern:
//
//	func tokenizeText(text string) []string {
//	    // Get from pool
//	    sm := getStateMachine(text)
//	    tokens := tokenBufPool.Get().([]string)
//	    
//	    // Use
//	    sm.tokens = tokens[:0]
//	    // ... perform tokenization ...
//	    
//	    // Copy results
//	    result := make([]string, len(sm.tokens))
//	    copy(result, sm.tokens)
//	    
//	    // Return to pool
//	    if cap(sm.tokens) <= maxPooledTokenBufferCapacity {
//	        tokenBufPool.Put(sm.tokens[:0])
//	    }
//	    sm.tokens = nil
//	    putStateMachine(sm)
//	    
//	    return result
//	}
// Package llama3 implements the Llama 3 tokenizer in pure Go.
//
// This package provides exact compatibility with the official Llama 3 tokenization,
// supporting byte-level BPE (Byte Pair Encoding) tokenization with all special tokens.
// It is a faithful port of the JavaScript implementation and produces identical
// token sequences.
//
// # Overview
//
// The Llama 3 tokenizer uses a three-stage process:
//
//  1. Pre-tokenization: Text is split into words, whitespace, and punctuation
//     using a state machine that replicates the JavaScript regex behavior
//  2. Byte-level encoding: Text is converted to a custom byte representation
//  3. BPE algorithm: Subword units are merged according to learned merge rules
//
// The tokenizer uses a vocabulary of 128,256 tokens:
//   - 128,000 base tokens
//   - 256 special tokens (e.g., <|begin_of_text|>, <|end_of_text|>)
//
// # Architecture
//
//	┌─────────────┐
//	│  Input Text │
//	└──────┬──────┘
//	       │
//	       ▼
//	┌─────────────────┐     ┌─────────────────┐
//	│ Special Token   │────▶│ State Machine   │
//	│ Splitting       │     │ Pre-tokenization│
//	└─────────────────┘     └────────┬────────┘
//	                                 │
//	                                 ▼
//	                        ┌─────────────────┐
//	                        │ Byte-level      │
//	                        │ Encoding        │
//	                        └────────┬────────┘
//	                                 │
//	                                 ▼
//	                        ┌─────────────────┐
//	                        │ BPE Algorithm   │
//	                        │ (with caching)  │
//	                        └────────┬────────┘
//	                                 │
//	                                 ▼
//	                        ┌─────────────────┐
//	                        │ Token IDs       │
//	                        └─────────────────┘
//
// # Basic Usage
//
//	tokenizer, err := llama3.New()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Encode text to token IDs
//	tokens := tokenizer.Encode("Hello, world!", nil)
//
//	// Decode token IDs back to text
//	text := tokenizer.Decode(tokens)
//
// # Advanced Usage
//
// The tokenizer can be configured with various options:
//
//	// Create with custom cache size
//	tokenizer, err := llama3.New(
//	    llama3.WithCacheSize(1000),
//	)
//
//	// Create with custom vocabulary and merges
//	tokenizer, err := llama3.New(
//	    llama3.WithVocabulary(customVocab),
//	    llama3.WithMerges(customMerges),
//	    llama3.WithSpecialTokens(customSpecialTokens),
//	)
//
// # State Machine
//
// The pre-tokenization stage uses a custom state machine that exactly replicates
// the JavaScript regex pattern. This ensures 100% compatibility, including edge
// cases like negative lookahead for whitespace patterns.
//
// The state machine matches patterns in this order:
//  1. Contractions: (?i:'s|'t|'re|'ve|'m|'ll|'d)
//  2. Words with prefix: [^\r\n\p{L}\p{N}]?\p{L}+
//  3. Numbers: \p{N}{1,3}
//  4. Punctuation:  ?[^\s\p{L}\p{N}]+[\r\n]*
//  5. Newlines: \s*[\r\n]+
//  6. Whitespace: \s+(?!\S)
//
// # Performance
//
// The tokenizer is optimized for production use:
//   - Object pooling reduces allocations by 36%
//   - BPE results are cached for repeated tokens
//   - State machines and token buffers are reused
//   - Thread-safe design allows concurrent usage
//
// # Memory Management
//
// The package uses sync.Pool for efficient memory management:
//   - State machines are pooled and reused
//   - Token buffers are pooled (up to 1024 capacity)
//   - BPE merge operations use a priority queue
//
// Pool Usage Patterns:
//
// 1. State Machine Pooling (stateMachinePool)
//    - Reuses StateMachine instances across tokenization calls
//    - Reduces allocations for the input rune slice and token slice
//    - Pool never limits the number of state machines
//    - State machines are reset before reuse
//
// 2. Token Buffer Pooling (tokenBufPool)
//    - Reuses []string slices for collecting tokens
//    - Initial capacity: 64 tokens
//    - Maximum pooled capacity: 1024 tokens
//    - Buffers exceeding the maximum are not returned to the pool
//
// Memory Lifecycle:
//
// 1. Allocation: First call creates new instance, subsequent calls may reuse
// 2. Usage: Instance is used for one tokenization operation
// 3. Return to Pool: References cleared, slices reset, large buffers discarded
// 4. Garbage Collection: Go runtime may clear pools during GC
//
// Performance: Benchmarks show 36% memory reduction with pooling
//
// # Error Handling
//
// The package defines custom error types for better error handling:
//   - DataError: Issues with loading or processing tokenizer data
//   - TokenError: Issues with specific tokens or token IDs
//   - ConfigError: Issues with tokenizer configuration
//
// All errors implement the error interface and support error wrapping.
//
// # Thread Safety
//
// The tokenizer is safe for concurrent use. Multiple goroutines can encode
// and decode text simultaneously without issues. The internal cache uses
// read-write mutexes for efficient concurrent access.
package llama3

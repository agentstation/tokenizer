# Llama3 Tokenizer Implementation Documentation

## Overview

This document describes the Go implementation of the Llama3 tokenizer, which provides exact compatibility with the JavaScript reference implementation while following Go idioms and best practices.

## Architecture

### Core Components

1. **Tokenizer** (`tokenizer.go`)
   - Main entry point for encoding and decoding text
   - Manages vocabulary (128,256 tokens total)
   - Handles special tokens (256 special tokens)
   - Implements BPE (Byte Pair Encoding) algorithm

2. **State Machine** (`state_machine.go`)
   - Regex-free implementation for pre-tokenization
   - Matches JavaScript regex behavior exactly
   - Character-by-character parsing for reliability
   - Handles complex patterns like contractions and whitespace

3. **Byte-Level Encoding** (`utils.go`)
   - Maps UTF-8 bytes to unicode characters
   - Ensures consistent representation across platforms
   - Handles all 256 possible byte values

4. **BPE Engine** (`bpe.go`)
   - Implements the Byte Pair Encoding algorithm
   - Uses priority queue for efficient merge operations
   - Includes caching for performance optimization

5. **Special Tokens** (`special_tokens.go`)
   - Manages 256 special tokens
   - Handles token splitting and recognition
   - Provides regex patterns for token matching

## Key Design Decisions

### 1. Regex-Free Pre-tokenization

The original JavaScript implementation uses this regex pattern:
```javascript
(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+
```

Go's regex engine doesn't support negative lookahead `(?!\S)`, and there were subtle differences in Unicode handling. To ensure exact compatibility, we implemented a state machine that:

- Processes input character by character
- Implements each regex alternative in order
- Handles edge cases like `\s+(?!\S)` (whitespace not followed by non-whitespace)
- Preserves exact splitting behavior

### 2. Performance Optimizations

1. **Object Pooling**: Reusable state machines via `sync.Pool`
2. **Pre-allocation**: Estimated capacity for slices
3. **Caching**: BPE results cached for repeated tokens
4. **Efficient Data Structures**: Priority queue for BPE merges

### 3. Testing Strategy

1. **Unit Tests**: Cover individual components
2. **Integration Tests**: Test full tokenization pipeline
3. **Comparison Tests**: Validate against JavaScript implementation
4. **Test Vectors**: 100+ test cases covering edge cases

## Implementation Details

### State Machine Pattern Matching

The state machine implements these patterns in order:

1. **Contractions**: `'s`, `'t`, `'re`, `'ve`, `'m`, `'ll`, `'d` (case-insensitive)
2. **Words with Optional Prefix**: `[^\r\n\p{L}\p{N}]?\p{L}+`
3. **Numbers**: `\p{N}{1,3}` (1-3 digits)
4. **Punctuation**: ` ?[^\s\p{L}\p{N}]+[\r\n]*`
5. **Newline Sequences**: `\s*[\r\n]+`
6. **Whitespace**: `\s+(?!\S)` or `\s+`

### Special Whitespace Handling

The pattern `\s+(?!\S)` matches whitespace that is NOT followed by non-whitespace. This causes specific splitting behavior:

- Input: `"           grabbed"` (11 spaces)
- Output: `["          ", " grabbed"]` (10 spaces, then space+word)

This is because the 11th space is followed by 'g' (non-whitespace), so the regex backtracks.

### Byte-Level Encoding

The tokenizer uses a custom byte-to-unicode mapping:
- Printable ASCII (32-126, except space): mapped directly
- Space (32): mapped to 'Ġ' (U+0120)
- Other bytes: mapped to unicode range 256+

This ensures all possible byte sequences can be tokenized.

## Usage Examples

### Basic Tokenization

```go
tokenizer, err := llama3.NewDefault()
if err != nil {
    log.Fatal(err)
}

// Encode text
tokens := tokenizer.Encode("Hello world!", nil)
// Returns: [128000, 9906, 1917, 0, 128001]

// Decode tokens
text := tokenizer.Decode(tokens)
// Returns: "<|begin_of_text|>Hello world!<|end_of_text|>"
```

### Custom Options

```go
opts := &llama3.EncodeOptions{
    BOS: false,  // Don't add beginning of text token
    EOS: false,  // Don't add end of text token
}

tokens := tokenizer.Encode("Hello world!", opts)
// Returns: [9906, 1917, 0]
```

### Special Token Handling

```go
// Get special token ID
id, exists := tokenizer.GetSpecialTokenID("<|begin_of_text|>")
// id = 128000, exists = true

// Check vocabulary size
size := tokenizer.VocabSize()
// size = 128256
```

## Performance Characteristics

Based on benchmarks on Apple M2 Max:

- **Encode**: ~2μs per operation, 27 allocations
- **Decode**: ~840ns per operation, 19 allocations
- **Memory**: ~1KB per typical sentence
- **Throughput**: ~500K tokens/second

## Differences from JavaScript Implementation

1. **No Regex Dependencies**: Pure state machine implementation
2. **Better Performance**: Optimized for Go's runtime
3. **Type Safety**: Strongly typed interfaces
4. **Concurrency Safe**: Thread-safe with proper locking

## Future Improvements

1. **SIMD Optimizations**: Use assembly for byte encoding
2. **Zero-Copy APIs**: Reduce allocations further
3. **Streaming Support**: Process large texts incrementally
4. **Custom Vocabularies**: Support different model variants

## Debugging

To debug tokenization issues:

1. Use `TestDebugFullTokenization` to see step-by-step process
2. Compare with JavaScript using `comparison_test.go`
3. Generate test vectors with `cmd/generate_vectors`
4. Use state machine tests to verify pre-tokenization

## Contributing

When making changes:

1. Run all tests: `go test ./llama3 -v`
2. Run comparison tests: `go test ./llama3 -run TestComparison`
3. Check benchmarks: `go test ./llama3 -bench=.`
4. Ensure JavaScript compatibility
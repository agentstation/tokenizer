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
tokenizer, err := llama3.New()
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

## Interfaces

The llama3 tokenizer provides several interfaces for extensibility and testing. These interfaces follow Go idioms - they are small, focused, and located near their implementations.

### Core Interfaces

#### Encoder and Decoder
Located in `tokenizer.go`, these are the primary interfaces for text tokenization:

```go
type Encoder interface {
    Encode(text string, opts *EncodeOptions) []int
}

type Decoder interface {
    Decode(tokens []int) string
}
```

The `Tokenizer` type implements both interfaces. Function adapters `EncoderFunc` and `DecoderFunc` are provided for testing.

#### Scanner
Located in `scanner.go`, provides streaming tokenization following the `bufio.Scanner` pattern:

```go
type Scanner interface {
    Scan() bool
    Token() int
    Text() string
    Err() error
}
```

Create a scanner with `tokenizer.NewScanner(reader)` or `NewScannerOptions` for custom configuration.

### Pipeline Interfaces

#### PreTokenizer
Located in `pretokenizer.go`, handles the first stage of tokenization:

```go
type PreTokenizer interface {
    PreTokenize(text string) []string
}
```

Splits text into pre-tokens (words, punctuation, etc.) before BPE processing.

#### BPEProcessor
Located in `bpe.go`, implements Byte Pair Encoding:

```go
type BPEProcessor interface {
    ProcessBPE(pretoken string) []int
}
```

Applies BPE algorithm to pre-tokenized strings.

### Infrastructure Interfaces

#### Cache
Located in `cache.go`, provides BPE result caching:

```go
type Cache interface {
    Get(key string) ([]int, bool)
    Put(key string, value []int)
}
```

Two implementations provided:
- `lruCache` - LRU eviction with configurable capacity
- `simpleCache` - Unlimited map-based cache

#### DataLoader
Located in `data.go`, abstracts tokenizer data loading:

```go
type DataLoader interface {
    LoadVocabulary() ([]string, error)
    LoadMerges() (map[string]int, error)
}
```

Implementations:
- `DefaultDataLoader` - Uses embedded data
- `FileDataLoader` - Loads from files
- `DataLoaderFunc` - Function-based adapter for testing

## Streaming API

The streaming API provides memory-efficient tokenization for large texts and real-time processing scenarios.

### Scanner Usage

Basic streaming:
```go
tokenizer, _ := llama3.New()
reader := strings.NewReader("Large text to tokenize...")
scanner := tokenizer.NewScanner(reader)

for scanner.Scan() {
    token := scanner.Token()
    // Process token
}

if err := scanner.Err(); err != nil {
    // Handle error
}
```

Custom buffer configuration:
```go
scanner := tokenizer.NewScannerOptions(reader,
    llama3.WithBufferSize(8192),
    llama3.WithMaxBuffer(1024*1024),
    llama3.WithEncodeOptions(&llama3.EncodeOptions{
        BOS: false,
        EOS: false,
    }),
)
```

### Zero-Allocation Methods

The tokenizer provides methods to minimize allocations:

```go
// Avoid string conversion for binary data
tokens := tokenizer.EncodeBytes(data, opts)

// Return bytes directly without string allocation
bytes := tokenizer.DecodeBytes(tokens)

// Reuse existing slice capacity
tokens = tokenizer.AppendTokens(tokens, text, opts)
```

### UTF-8 Boundary Handling

The scanner ensures UTF-8 sequences aren't split at buffer boundaries:
- Proactively checks UTF-8 boundaries before writing to buffer
- Saves incomplete UTF-8 sequences as pending bytes
- Prevents corruption when buffer limits are reached mid-character

### Performance Characteristics

Streaming API benchmarks:
- `EncodeBytes`: ~14μs (similar to Encode)
- `DecodeBytes`: ~602ns (3% faster than Decode)
- `AppendTokens`: ~1.5μs with efficient capacity reuse
- `Scanner`: ~141μs for 100 repetitions

## Contributing

When making changes:

1. Run all tests: `go test ./llama3 -v`
2. Run comparison tests: `go test ./llama3 -run TestComparison`
3. Check benchmarks: `go test ./llama3 -bench=.`
4. Ensure JavaScript compatibility
# Llama 3 Tokenizer for Go

A pure Go implementation of the Llama 3 tokenizer, providing exact compatibility with the official Llama 3 tokenization used in models 3.0, 3.1, 3.2, and 3.3.

## Features

- **Exact Compatibility**: Produces identical token sequences to the official implementation
- **Full UTF-8 Support**: Handles multilingual text and emojis correctly
- **All Special Tokens**: Supports all 256 special tokens including `<|begin_of_text|>`, `<|end_of_text|>`, etc.
- **Thread-Safe**: Safe for concurrent use with built-in caching
- **Zero Dependencies**: Pure Go implementation with only standard library dependencies
- **High Performance**: Optimized BPE implementation with caching

## Installation

```bash
go get github.com/agentstation/tokenizer/llama3
```

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/agentstation/tokenizer/llama3"
)

func main() {
    // Create tokenizer with default Llama 3 vocabulary
    tokenizer, err := llama3.New()
    if err != nil {
        panic(err)
    }
    
    // Encode text to tokens
    text := "Hello world!"
    tokens := tokenizer.Encode(text, nil)
    fmt.Printf("Text: %s\n", text)
    fmt.Printf("Tokens: %v\n", tokens)
    // Output: [128000, 9906, 1917, 0, 128001]
    
    // Decode tokens back to text
    decoded := tokenizer.Decode(tokens)
    fmt.Printf("Decoded: %s\n", decoded)
    // Output: <|begin_of_text|>Hello world!<|end_of_text|>
}
```

### Encoding Options

Control the addition of special tokens:

```go
// Without special tokens
opts := &llama3.EncodeOptions{
    BOS: false,  // Don't add <|begin_of_text|>
    EOS: false,  // Don't add <|end_of_text|>
}
tokens := tokenizer.Encode("Hello world!", opts)
// Output: [9906, 1917, 0]
```

### Special Tokens

Work with special tokens:

```go
// Get special token ID
id, err := tokenizer.GetSpecialTokenID("<|end_of_text|>")
if err == nil {
    fmt.Printf("EOT token ID: %d\n", id)
}

// Encode text containing special tokens
text := "<|start_header_id|>system<|end_header_id|>You are a helpful assistant."
tokens := tokenizer.Encode(text, nil)
```

### Advanced Options

Create a tokenizer with custom configuration:

```go
// Create tokenizer with custom cache size
tokenizer, err := llama3.New(
    llama3.WithCacheSize(8192), // Custom cache size (default: 4096)
)
if err != nil {
    panic(err)
}

// Or with custom data files
vocabBase64 := "..." // Base64-encoded vocabulary JSON (about 1.5MB)
mergesBinary := "..." // Base64-encoded binary merge rules (about 1.5MB)
specialTokens := []string{
    "<|begin_of_text|>",
    "<|end_of_text|>",
    "<|start_header_id|>",
    "<|end_header_id|>",
    // ... other special tokens
}

tokenizer, err := llama3.New(
    llama3.WithVocabData(vocabBase64, mergesBinary, specialTokens),
)

// Example: Loading from files
vocabData, err := os.ReadFile("vocab_base64.txt")
if err != nil {
    panic(err)
}
mergesData, err := os.ReadFile("merges_binary.txt")
if err != nil {
    panic(err)
}

tokenizer, err = llama3.New(
    llama3.WithVocabData(
        string(vocabData),
        string(mergesData),
        []string{
            "<|begin_of_text|>",
            "<|end_of_text|>",
            "<|start_header_id|>",
            "<|end_header_id|>",
            "<|eot_id|>",
            "<|python_tag|>",
            // Add all 256 special tokens as needed
        },
    ),
)
```

### Optimistic Token Counting

For fine-tuned models with custom special tokens:

```go
// Counts any <|...|> pattern as a special token
count := tokenizer.OptimisticCount("Custom text with <|my_token|> special tokens")
```

## Implementation Details

This implementation follows the Llama 3 tokenization specification:

1. **Pre-tokenization**: Uses a state machine to split text into words and subwords
2. **Byte-level encoding**: Converts text to UTF-8 bytes with special character mappings
3. **BPE Algorithm**: Applies Byte Pair Encoding with the Llama 3 merge rules
4. **Special token handling**: Recognizes and preserves all Llama 3 special tokens

The tokenizer uses a vocabulary size of 128,256 tokens, including:
- 128,000 base tokens
- 256 special tokens

## Compatibility

Compatible with:
- Llama 3.0
- Llama 3.1
- Llama 3.2
- Llama 3.3
- Fine-tuned models based on Llama 3

### Full JavaScript Compatibility

This implementation achieves 100% compatibility with the JavaScript reference implementation through a custom state machine that exactly replicates the regex behavior. All edge cases, including complex whitespace patterns, are handled correctly.

For detailed implementation notes and technical design decisions, see [IMPLEMENTATION.md](IMPLEMENTATION.md).


## Data Files

The tokenizer requires two data files:
- `vocab_base64.txt`: Base64-encoded vocabulary (1.5MB)
- `merges_binary.txt`: Base64-encoded merge rules (1.5MB)

The data files are included in this repository and will be automatically loaded when you use the tokenizer.

These files were extracted from the [llama3-tokenizer-js](https://github.com/belladoreai/llama3-tokenizer-js) project.

### Build Options

**Option 1: Embedded Data (Recommended)**
```bash
# Build with embedded data files
go build -tags embed

# The binary will contain the tokenizer data
```

**Option 2: External Data Files**
```bash
# Build without embedded data
go build

# Place data files in one of these locations:
# - Same directory as the binary
# - ./llama3/ subdirectory
# - Parent directory
```

The tokenizer will automatically try to load data from standard locations if not embedded.

## Testing

Run the test suite:

```bash
go test ./llama3
```

Run compatibility tests (476 test cases):

```bash
go test -run TestCompatibility -v ./llama3
```

Run benchmarks:

```bash
go test -bench=. ./llama3
```

## Performance

The tokenizer is optimized for production use with:

- **Object pooling**: Reuses state machines and token buffers for 36% less memory usage
- **BPE caching**: Caches merge operations for repeated tokens
- **Efficient data structures**: Priority queue for BPE, pre-computed lookups
- **Comprehensive benchmarks**: See [OPTIMIZATIONS.md](OPTIMIZATIONS.md) for implementation details

Run benchmarks:

```bash
go test -bench=. -benchmem ./llama3
```

## License

MIT License - see LICENSE file for details.

## Acknowledgments

This implementation is based on the JavaScript [llama3-tokenizer-js](https://github.com/belladoreai/llama3-tokenizer-js) by belladoreai. The vocabulary and merge data files were extracted from their bundled JavaScript implementation.

<!-- gomarkdoc:embed:start -->

<!-- Code generated by gomarkdoc. DO NOT EDIT -->

# llama3

```go
import "github.com/agentstation/tokenizer/llama3"
```

Package llama3 implements the Llama 3 tokenizer in Go. This file contains all constants used throughout the tokenizer implementation.

Package llama3 implements the Llama 3 tokenizer in pure Go.

This package provides exact compatibility with the official Llama 3 tokenization, supporting byte\-level BPE \(Byte Pair Encoding\) tokenization with all special tokens. It is a faithful port of the JavaScript implementation and produces identical token sequences.

### Overview

The Llama 3 tokenizer uses a three\-stage process:

1. Pre\-tokenization: Text is split into words, whitespace, and punctuation using a state machine that replicates the JavaScript regex behavior
2. Byte\-level encoding: Text is converted to a custom byte representation
3. BPE algorithm: Subword units are merged according to learned merge rules

The tokenizer uses a vocabulary of 128,256 tokens:

- 128,000 base tokens
- 256 special tokens \(e.g., \<|begin\_of\_text|\>, \<|end\_of\_text|\>\)

### Architecture

```
┌─────────────┐
│  Input Text │
└──────┬──────┘
       │
       ▼
┌─────────────────┐     ┌─────────────────┐
│ Special Token   │────▶│ State Machine   │
│ Splitting       │     │ Pre-tokenization│
└─────────────────┘     └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │ Byte-level      │
                        │ Encoding        │
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │ BPE Algorithm   │
                        │ (with caching)  │
                        └────────┬────────┘
                                 │
                                 ▼
                        ┌─────────────────┐
                        │ Token IDs       │
                        └─────────────────┘
```

### Basic Usage

```
tokenizer, err := llama3.New()
if err != nil {
    log.Fatal(err)
}

// Encode text to token IDs
tokens := tokenizer.Encode("Hello, world!", nil)

// Decode token IDs back to text
text := tokenizer.Decode(tokens)
```

### Advanced Usage

The tokenizer can be configured with various options:

```
// Create with custom cache size
tokenizer, err := llama3.New(
    llama3.WithCacheSize(1000),
)

// Create with custom vocabulary and merges
tokenizer, err := llama3.New(
    llama3.WithVocabulary(customVocab),
    llama3.WithMerges(customMerges),
    llama3.WithSpecialTokens(customSpecialTokens),
)
```

### State Machine

The pre\-tokenization stage uses a custom state machine that exactly replicates the JavaScript regex pattern. This ensures 100% compatibility, including edge cases like negative lookahead for whitespace patterns.

The state machine matches patterns in this order:

1. Contractions: \(?i:'s|'t|'re|'ve|'m|'ll|'d\)
2. Words with prefix: \[^\\r\\n\\p\{L\}\\p\{N\}\]?\\p\{L\}\+
3. Numbers: \\p\{N\}\{1,3\}
4. Punctuation: ?\[^\\s\\p\{L\}\\p\{N\}\]\+\[\\r\\n\]\*
5. Newlines: \\s\*\[\\r\\n\]\+
6. Whitespace: \\s\+\(?\!\\S\)

### Performance

The tokenizer is optimized for production use:

- Object pooling reduces allocations by 36%
- BPE results are cached for repeated tokens
- State machines and token buffers are reused
- Thread\-safe design allows concurrent usage

### Memory Management

The package uses sync.Pool for efficient memory management:

- State machines are pooled and reused
- Token buffers are pooled \(up to 1024 capacity\)
- BPE merge operations use a priority queue

Pool Usage Patterns:

1. State Machine Pooling \(stateMachinePool\)

- Reuses StateMachine instances across tokenization calls
- Reduces allocations for the input rune slice and token slice
- Pool never limits the number of state machines
- State machines are reset before reuse

2. Token Buffer Pooling \(tokenBufPool\)

- Reuses \[\]string slices for collecting tokens
- Initial capacity: 64 tokens
- Maximum pooled capacity: 1024 tokens
- Buffers exceeding the maximum are not returned to the pool

Memory Lifecycle:

1. Allocation: First call creates new instance, subsequent calls may reuse 2. Usage: Instance is used for one tokenization operation 3. Return to Pool: References cleared, slices reset, large buffers discarded 4. Garbage Collection: Go runtime may clear pools during GC

Performance: Benchmarks show 36% memory reduction with pooling

### Error Handling

The package defines custom error types for better error handling:

- DataError: Issues with loading or processing tokenizer data
- TokenError: Issues with specific tokens or token IDs
- ConfigError: Issues with tokenizer configuration

All errors implement the error interface and support error wrapping.

### Thread Safety

The tokenizer is safe for concurrent use. Multiple goroutines can encode and decode text simultaneously without issues. The internal cache uses read\-write mutexes for efficient concurrent access.

Package llama3 implements the Llama 3 tokenizer in Go. It provides exact compatibility with the official Llama 3 tokenization, supporting byte\-level BPE tokenization with all special tokens.

Package llama3 implements the Llama 3 tokenizer in Go. This file contains the public API including interfaces and options.

## Index

- [Variables](<#variables>)
- [func NewConfigError\(field string, value any, err error\) error](<#NewConfigError>)
- [func NewDataError\(op, path string, err error\) error](<#NewDataError>)
- [func NewTokenError\(op, token string, err error\) error](<#NewTokenError>)
- [func NewTokenIDError\(op string, tokenID int, err error\) error](<#NewTokenIDError>)
- [type BPE](<#BPE>)
- [type Cache](<#Cache>)
- [type ConfigError](<#ConfigError>)
  - [func \(e \*ConfigError\) Error\(\) string](<#ConfigError.Error>)
  - [func \(e \*ConfigError\) Unwrap\(\) error](<#ConfigError.Unwrap>)
- [type DataError](<#DataError>)
  - [func \(e \*DataError\) Error\(\) string](<#DataError.Error>)
  - [func \(e \*DataError\) Unwrap\(\) error](<#DataError.Unwrap>)
- [type Decoder](<#Decoder>)
- [type DecoderFunc](<#DecoderFunc>)
  - [func \(f DecoderFunc\) Decode\(tokens \[\]int\) string](<#DecoderFunc.Decode>)
- [type EncodeOptions](<#EncodeOptions>)
- [type Encoder](<#Encoder>)
- [type EncoderFunc](<#EncoderFunc>)
  - [func \(f EncoderFunc\) Encode\(text string, opts \*EncodeOptions\) \[\]int](<#EncoderFunc.Encode>)
- [type Option](<#Option>)
  - [func WithCacheSize\(size int\) Option](<#WithCacheSize>)
  - [func WithDataFiles\(vocabPath, mergesPath string\) Option](<#WithDataFiles>)
  - [func WithDataLoader\(loader VocabularyDataLoader\) Option](<#WithDataLoader>)
  - [func WithSpecialTokens\(tokens \[\]string\) Option](<#WithSpecialTokens>)
- [type PreTokenizer](<#PreTokenizer>)
- [type Scanner](<#Scanner>)
- [type ScannerOption](<#ScannerOption>)
- [type TokenError](<#TokenError>)
  - [func \(e \*TokenError\) Error\(\) string](<#TokenError.Error>)
  - [func \(e \*TokenError\) Unwrap\(\) error](<#TokenError.Unwrap>)
- [type Tokenizer](<#Tokenizer>)
  - [func New\(opts ...Option\) \(\*Tokenizer, error\)](<#New>)
  - [func \(t \*Tokenizer\) AppendTokens\(dst \[\]int, text string, opts \*EncodeOptions\) \[\]int](<#Tokenizer.AppendTokens>)
  - [func \(t \*Tokenizer\) Decode\(tokenIDs \[\]int\) string](<#Tokenizer.Decode>)
  - [func \(t \*Tokenizer\) DecodeBytes\(tokenIDs \[\]int\) \[\]byte](<#Tokenizer.DecodeBytes>)
  - [func \(t \*Tokenizer\) Encode\(text string, opts \*EncodeOptions\) \[\]int](<#Tokenizer.Encode>)
  - [func \(t \*Tokenizer\) EncodeBPE\(pretoken string\) \[\]int](<#Tokenizer.EncodeBPE>)
  - [func \(t \*Tokenizer\) EncodeBytes\(data \[\]byte, opts \*EncodeOptions\) \[\]int](<#Tokenizer.EncodeBytes>)
  - [func \(t \*Tokenizer\) GetSpecialTokenID\(token string\) \(int, error\)](<#Tokenizer.GetSpecialTokenID>)
  - [func \(t \*Tokenizer\) NewScanner\(r io.Reader\) Scanner](<#Tokenizer.NewScanner>)
  - [func \(t \*Tokenizer\) NewScannerOptions\(r io.Reader, opts ...ScannerOption\) Scanner](<#Tokenizer.NewScannerOptions>)
  - [func \(t \*Tokenizer\) OptimisticCount\(text string\) int](<#Tokenizer.OptimisticCount>)
  - [func \(t \*Tokenizer\) PreTokenize\(text string\) \[\]string](<#Tokenizer.PreTokenize>)
  - [func \(t \*Tokenizer\) Process\(r io.Reader, w io.Writer\) \(int64, error\)](<#Tokenizer.Process>)
  - [func \(t \*Tokenizer\) TokenStream\(r io.Reader\) \(\<\-chan int, \<\-chan error\)](<#Tokenizer.TokenStream>)
  - [func \(t \*Tokenizer\) VocabSize\(\) int](<#Tokenizer.VocabSize>)
- [type VocabularyDataLoader](<#VocabularyDataLoader>)
- [type VocabularyDataLoaderFunc](<#VocabularyDataLoaderFunc>)
  - [func \(d VocabularyDataLoaderFunc\) LoadMerges\(\) \(map\[string\]int, error\)](<#VocabularyDataLoaderFunc.LoadMerges>)
  - [func \(d VocabularyDataLoaderFunc\) LoadVocabulary\(\) \(\[\]string, error\)](<#VocabularyDataLoaderFunc.LoadVocabulary>)


## Variables

<a name="ErrDataNotFound"></a>Common errors

```go
var (
    // ErrDataNotFound indicates that the tokenizer data files could not be found
    ErrDataNotFound = errors.New("tokenizer data not found")

    // ErrInvalidToken indicates an invalid token was provided
    ErrInvalidToken = errors.New("invalid token")

    // ErrTokenNotFound indicates a token was not found in the vocabulary
    ErrTokenNotFound = errors.New("token not found")

    // ErrInvalidTokenID indicates an invalid token ID was provided
    ErrInvalidTokenID = errors.New("invalid token ID")
)
```

<a name="WithBufferSize"></a>Scanner option functions \- these are re\-exported from the scanner package

```go
var (
    // WithBufferSize sets the internal buffer size for reading.
    // Default is 4096 bytes.
    WithBufferSize = scanner.WithBufferSize

    // WithMaxBuffer sets the maximum buffer size before forcing tokenization.
    // This prevents unbounded memory growth for pathological inputs.
    // Default is 1MB.
    WithMaxBuffer = scanner.WithMaxBuffer

    // WithEncodeOptions sets encoding options for the scanner.
    WithEncodeOptions = func(opts *EncodeOptions) ScannerOption {
        return scanner.WithEncodeOptions(&scanner.EncodeOptions{
            BOS: opts.BOS,
            EOS: opts.EOS,
        })
    }
)
```

<a name="NewConfigError"></a>
## func [NewConfigError](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L96>)

```go
func NewConfigError(field string, value any, err error) error
```

NewConfigError creates a new ConfigError

<a name="NewDataError"></a>
## func [NewDataError](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L81>)

```go
func NewDataError(op, path string, err error) error
```

NewDataError creates a new DataError

<a name="NewTokenError"></a>
## func [NewTokenError](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L86>)

```go
func NewTokenError(op, token string, err error) error
```

NewTokenError creates a new TokenError

<a name="NewTokenIDError"></a>
## func [NewTokenIDError](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L91>)

```go
func NewTokenIDError(op string, tokenID int, err error) error
```

NewTokenIDError creates a new TokenError with a token ID

<a name="BPE"></a>
## type [BPE](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L80-L84>)

BPE is the interface for Byte Pair Encoding processing. BPE merges frequently occurring character pairs to create subword tokens.

```go
type BPE interface {
    // EncodeBPE applies byte pair encoding to a pre-tokenized string.
    // Returns a slice of token IDs representing the encoded text.
    EncodeBPE(pretoken string) []int
}
```

<a name="Cache"></a>
## type [Cache](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L446-L454>)

Cache is the interface for caching BPE results. BPE tokenization can be expensive for repeated text patterns, so caching improves performance significantly.

The cache key is typically the pre\-tokenized text string, and the value is the slice of token IDs produced by BPE.

Implementations should be thread\-safe if the tokenizer will be used concurrently.

```go
type Cache interface {
    // Get retrieves a cached BPE result.
    // Returns the token IDs and true if found, or nil and false if not cached.
    Get(key string) ([]int, bool)

    // Put stores a BPE result in the cache.
    // The implementation may evict old entries based on its eviction policy.
    Put(key string, value []int)
}
```

<a name="ConfigError"></a>
## type [ConfigError](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L64-L68>)

ConfigError represents an error in tokenizer configuration

```go
type ConfigError struct {
    Field string // Configuration field that has an error
    Value any    // The invalid value
    Err   error  // Underlying error
}
```

<a name="ConfigError.Error"></a>
### func \(\*ConfigError\) [Error](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L70>)

```go
func (e *ConfigError) Error() string
```



<a name="ConfigError.Unwrap"></a>
### func \(\*ConfigError\) [Unwrap](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L74>)

```go
func (e *ConfigError) Unwrap() error
```



<a name="DataError"></a>
## type [DataError](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L24-L28>)

DataError represents an error related to tokenizer data loading or processing

```go
type DataError struct {
    Op   string // Operation that failed
    Path string // File path if applicable
    Err  error  // Underlying error
}
```

<a name="DataError.Error"></a>
### func \(\*DataError\) [Error](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L30>)

```go
func (e *DataError) Error() string
```



<a name="DataError.Unwrap"></a>
### func \(\*DataError\) [Unwrap](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L37>)

```go
func (e *DataError) Unwrap() error
```



<a name="Decoder"></a>
## type [Decoder](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L55-L58>)

Decoder is the interface for decoding tokens to text. This interface is useful for testing and creating mock implementations.

```go
type Decoder interface {
    // Decode converts a sequence of token IDs back to text.
    Decode(tokens []int) string
}
```

<a name="DecoderFunc"></a>
## type [DecoderFunc](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L71>)

DecoderFunc is an adapter to allow ordinary functions to be used as Decoders. This is useful for creating mock decoders in tests.

```go
type DecoderFunc func(tokens []int) string
```

<a name="DecoderFunc.Decode"></a>
### func \(DecoderFunc\) [Decode](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L74>)

```go
func (f DecoderFunc) Decode(tokens []int) string
```

Decode calls f\(tokens\).

<a name="EncodeOptions"></a>
## type [EncodeOptions](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L107-L112>)

EncodeOptions controls the encoding behavior.

```go
type EncodeOptions struct {
    // BOS adds the beginning-of-text token if true (default: true)
    BOS bool
    // EOS adds the end-of-text token if true (default: true)
    EOS bool
}
```

<a name="Encoder"></a>
## type [Encoder](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L48-L51>)

Encoder is the interface for encoding text to tokens. This interface is useful for testing and creating mock implementations.

```go
type Encoder interface {
    // Encode converts text to a sequence of token IDs.
    Encode(text string, opts *EncodeOptions) []int
}
```

<a name="EncoderFunc"></a>
## type [EncoderFunc](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L62>)

EncoderFunc is an adapter to allow ordinary functions to be used as Encoders. This is useful for creating mock encoders in tests.

```go
type EncoderFunc func(text string, opts *EncodeOptions) []int
```

<a name="EncoderFunc.Encode"></a>
### func \(EncoderFunc\) [Encode](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L65>)

```go
func (f EncoderFunc) Encode(text string, opts *EncodeOptions) []int
```

Encode calls f\(text, opts\).

<a name="Option"></a>
## type [Option](<https://github.com/agentstation/tokenizer/blob/master/llama3/options.go#L13>)

Option is a functional option for configuring a Tokenizer.

```go
type Option func(*config) error
```

<a name="WithCacheSize"></a>
### func [WithCacheSize](<https://github.com/agentstation/tokenizer/blob/master/llama3/options.go#L40>)

```go
func WithCacheSize(size int) Option
```

WithCacheSize sets the maximum size of the BPE cache. Set to 0 to disable caching. Default is unlimited.

<a name="WithDataFiles"></a>
### func [WithDataFiles](<https://github.com/agentstation/tokenizer/blob/master/llama3/options.go#L65>)

```go
func WithDataFiles(vocabPath, mergesPath string) Option
```

WithDataFiles loads vocabulary and merges from files instead of embedded data. The vocabulary file should contain base64\-encoded vocabulary data. The merges file should contain base64\-encoded binary merge data.

<a name="WithDataLoader"></a>
### func [WithDataLoader](<https://github.com/agentstation/tokenizer/blob/master/llama3/options.go#L52>)

```go
func WithDataLoader(loader VocabularyDataLoader) Option
```

WithDataLoader sets a custom data loader for the tokenizer. This allows loading vocabulary and merges from custom sources.

<a name="WithSpecialTokens"></a>
### func [WithSpecialTokens](<https://github.com/agentstation/tokenizer/blob/master/llama3/options.go#L17>)

```go
func WithSpecialTokens(tokens []string) Option
```

WithSpecialTokens sets custom special tokens for the tokenizer. If nil, the default Llama 3 special tokens will be used.

<a name="PreTokenizer"></a>
## type [PreTokenizer](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L89-L93>)

PreTokenizer is the interface for pre\-tokenization. Pre\-tokenization splits text into words, numbers, punctuation, etc. before the BPE algorithm is applied.

```go
type PreTokenizer interface {
    // PreTokenize splits text into pre-tokens according to the tokenizer's rules.
    // Returns a slice of pre-token strings ready for BPE processing.
    PreTokenize(text string) []string
}
```

<a name="Scanner"></a>
## type [Scanner](<https://github.com/agentstation/tokenizer/blob/master/llama3/scanner.go#L12-L26>)

Scanner provides streaming tokenization following the bufio.Scanner pattern. It reads text incrementally and produces tokens one at a time.

```go
type Scanner interface {
    // Scan advances to the next token. Returns false at EOF or on error.
    Scan() bool

    // Token returns the most recent token ID produced by Scan.
    // Valid only after a successful call to Scan.
    Token() int

    // Text returns the text that produced the current token.
    // Valid only after a successful call to Scan.
    Text() string

    // Err returns the first error encountered during scanning.
    Err() error
}
```

<a name="ScannerOption"></a>
## type [ScannerOption](<https://github.com/agentstation/tokenizer/blob/master/llama3/scanner.go#L29>)

ScannerOption configures scanner behavior.

```go
type ScannerOption = scanner.Option
```

<a name="TokenError"></a>
## type [TokenError](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L42-L47>)

TokenError represents an error related to token operations

```go
type TokenError struct {
    Token   string // The token that caused the error
    TokenID int    // The token ID if applicable
    Op      string // Operation that failed
    Err     error  // Underlying error
}
```

<a name="TokenError.Error"></a>
### func \(\*TokenError\) [Error](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L49>)

```go
func (e *TokenError) Error() string
```



<a name="TokenError.Unwrap"></a>
### func \(\*TokenError\) [Unwrap](<https://github.com/agentstation/tokenizer/blob/master/llama3/errors.go#L59>)

```go
func (e *TokenError) Unwrap() error
```



<a name="Tokenizer"></a>
## type [Tokenizer](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L96-L104>)

Tokenizer implements the Llama 3 BPE tokenizer.

```go
type Tokenizer struct {
    // contains filtered or unexported fields
}
```

<a name="New"></a>
### func [New](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L142>)

```go
func New(opts ...Option) (*Tokenizer, error)
```

New creates a new Llama 3 tokenizer with the given options. If no options are provided, the default Llama 3 vocabulary and settings will be used.

Example:

```
tokenizer, err := llama3.New()
if err != nil {
    return err
}

// With custom vocabulary:
tokenizer, err := llama3.New(
    llama3.WithVocabulary(customVocab),
    llama3.WithMerges(customMerges),
)

// With cache size limit:
tokenizer, err := llama3.New(
    llama3.WithCacheSize(1000),
)
```

<a name="Tokenizer.AppendTokens"></a>
### func \(\*Tokenizer\) [AppendTokens](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L276>)

```go
func (t *Tokenizer) AppendTokens(dst []int, text string, opts *EncodeOptions) []int
```

AppendTokens appends tokens to dst, avoiding allocations when possible. dst can be nil, in which case a new slice is allocated. The resulting slice is returned and may have a different backing array than dst.

<a name="Tokenizer.Decode"></a>
### func \(\*Tokenizer\) [Decode](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L333>)

```go
func (t *Tokenizer) Decode(tokenIDs []int) string
```

Decode converts a sequence of token IDs back into text.

<details><summary>Example</summary>
<p>



```go
package main

import (
	"fmt"
	"log"

	"github.com/agentstation/tokenizer/llama3"
)

func main() {
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal(err)
	}

	// Decode token IDs back to text
	tokens := []int{9906, 1917, 0}
	text := tokenizer.Decode(tokens)

	fmt.Printf("Decoded text: %s\n", text)
	// Output would be: Hello world!
}
```

</p>
</details>

<a name="Tokenizer.DecodeBytes"></a>
### func \(\*Tokenizer\) [DecodeBytes](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L339>)

```go
func (t *Tokenizer) DecodeBytes(tokenIDs []int) []byte
```

DecodeBytes converts a sequence of token IDs back to UTF\-8 bytes. This avoids string allocation and is useful for performance\-critical paths.

<a name="Tokenizer.Encode"></a>
### func \(\*Tokenizer\) [Encode](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L217>)

```go
func (t *Tokenizer) Encode(text string, opts *EncodeOptions) []int
```

Encode converts text into a sequence of token IDs. If opts is nil, default options will be used.

<details><summary>Example</summary>
<p>



```go
package main

import (
	"fmt"
	"log"

	"github.com/agentstation/tokenizer/llama3"
)

func main() {
	// Create a tokenizer
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal(err)
	}

	// Encode some text
	text := "Hello, world!"
	tokens := tokenizer.Encode(text, nil)

	fmt.Printf("Text: %s\n", text)
	fmt.Printf("Token count: %d\n", len(tokens))
	// Note: actual output depends on having the Llama 3 data files
}
```

</p>
</details>

<details><summary>Example (Without Special Tokens)</summary>
<p>



```go
package main

import (
	"fmt"
	"log"

	"github.com/agentstation/tokenizer/llama3"
)

func main() {
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal(err)
	}

	// Encode without special tokens
	opts := &llama3.EncodeOptions{
		BOS: false,
		EOS: false,
	}

	text := "Hello, world!"
	tokens := tokenizer.Encode(text, opts)

	fmt.Printf("Tokens without BOS/EOS: %d\n", len(tokens))
}
```

</p>
</details>

<a name="Tokenizer.EncodeBPE"></a>
### func \(\*Tokenizer\) [EncodeBPE](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L478>)

```go
func (t *Tokenizer) EncodeBPE(pretoken string) []int
```

EncodeBPE implements the BPE interface.

<a name="Tokenizer.EncodeBytes"></a>
### func \(\*Tokenizer\) [EncodeBytes](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L269>)

```go
func (t *Tokenizer) EncodeBytes(data []byte, opts *EncodeOptions) []int
```

EncodeBytes converts bytes into a sequence of token IDs. This avoids string conversion overhead for binary data.

<a name="Tokenizer.GetSpecialTokenID"></a>
### func \(\*Tokenizer\) [GetSpecialTokenID](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L357>)

```go
func (t *Tokenizer) GetSpecialTokenID(token string) (int, error)
```

GetSpecialTokenID returns the token ID for a special token string.

<details><summary>Example</summary>
<p>



```go
package main

import (
	"fmt"
	"log"

	"github.com/agentstation/tokenizer/llama3"
)

func main() {
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal(err)
	}

	// Get the ID of a special token
	tokenID, err := tokenizer.GetSpecialTokenID("<|begin_of_text|>")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Begin-of-text token ID: %d\n", tokenID)
	// Output would be: 128000
}
```

</p>
</details>

<a name="Tokenizer.NewScanner"></a>
### func \(\*Tokenizer\) [NewScanner](<https://github.com/agentstation/tokenizer/blob/master/llama3/scanner.go#L65>)

```go
func (t *Tokenizer) NewScanner(r io.Reader) Scanner
```

NewScanner creates a scanner for streaming tokenization with default options.

<a name="Tokenizer.NewScannerOptions"></a>
### func \(\*Tokenizer\) [NewScannerOptions](<https://github.com/agentstation/tokenizer/blob/master/llama3/scanner.go#L70>)

```go
func (t *Tokenizer) NewScannerOptions(r io.Reader, opts ...ScannerOption) Scanner
```

NewScannerOptions creates a scanner with custom options.

<a name="Tokenizer.OptimisticCount"></a>
### func \(\*Tokenizer\) [OptimisticCount](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L373>)

```go
func (t *Tokenizer) OptimisticCount(text string) int
```

OptimisticCount returns the token count assuming anything that looks like a special token is actually a special token. This is useful for fine\-tuned models with modified special tokens.

<a name="Tokenizer.PreTokenize"></a>
### func \(\*Tokenizer\) [PreTokenize](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L504>)

```go
func (t *Tokenizer) PreTokenize(text string) []string
```

PreTokenize implements the PreTokenizer interface.

<a name="Tokenizer.Process"></a>
### func \(\*Tokenizer\) [Process](<https://github.com/agentstation/tokenizer/blob/master/llama3/scanner.go#L77>)

```go
func (t *Tokenizer) Process(r io.Reader, w io.Writer) (int64, error)
```

Process handles large files with controlled memory usage. It reads from r, tokenizes the content, and writes token IDs to w. Returns the number of tokens written and any error encountered.

<a name="Tokenizer.TokenStream"></a>
### func \(\*Tokenizer\) [TokenStream](<https://github.com/agentstation/tokenizer/blob/master/llama3/scanner.go#L107>)

```go
func (t *Tokenizer) TokenStream(r io.Reader) (<-chan int, <-chan error)
```

TokenStream provides channel\-based streaming for concurrent processing. The tokens channel will be closed when scanning completes. Any error will be sent on the error channel.

<a name="Tokenizer.VocabSize"></a>
### func \(\*Tokenizer\) [VocabSize](<https://github.com/agentstation/tokenizer/blob/master/llama3/tokenizer.go#L420>)

```go
func (t *Tokenizer) VocabSize() int
```

VocabSize returns the size of the vocabulary including special tokens.

<a name="VocabularyDataLoader"></a>
## type [VocabularyDataLoader](<https://github.com/agentstation/tokenizer/blob/master/llama3/vocab.go#L14-L22>)

VocabularyDataLoader is the interface for loading tokenizer vocabulary data. This includes vocabulary and merge rules needed for tokenization.

Implementations can load data from embedded resources, files, or custom sources. The tokenizer will call LoadVocabulary first, then LoadMerges.

```go
type VocabularyDataLoader interface {
    // LoadVocabulary loads and returns the vocabulary tokens.
    // The returned slice contains tokens indexed by their token ID.
    LoadVocabulary() ([]string, error)

    // LoadMerges loads and returns the BPE merge rules.
    // The returned map uses merge identifiers as keys and priorities as values.
    LoadMerges() (map[string]int, error)
}
```

<a name="VocabularyDataLoaderFunc"></a>
## type [VocabularyDataLoaderFunc](<https://github.com/agentstation/tokenizer/blob/master/llama3/vocab.go#L26-L29>)

VocabularyDataLoaderFunc is an adapter to allow using functions as VocabularyDataLoaders. This is useful for testing or custom data loading logic.

```go
type VocabularyDataLoaderFunc struct {
    VocabFunc  func() ([]string, error)
    MergesFunc func() (map[string]int, error)
}
```

<a name="VocabularyDataLoaderFunc.LoadMerges"></a>
### func \(VocabularyDataLoaderFunc\) [LoadMerges](<https://github.com/agentstation/tokenizer/blob/master/llama3/vocab.go#L37>)

```go
func (d VocabularyDataLoaderFunc) LoadMerges() (map[string]int, error)
```

LoadMerges calls the MergesFunc.

<a name="VocabularyDataLoaderFunc.LoadVocabulary"></a>
### func \(VocabularyDataLoaderFunc\) [LoadVocabulary](<https://github.com/agentstation/tokenizer/blob/master/llama3/vocab.go#L32>)

```go
func (d VocabularyDataLoaderFunc) LoadVocabulary() ([]string, error)
```

LoadVocabulary calls the VocabFunc.

Generated by [gomarkdoc](<https://github.com/princjef/gomarkdoc>)


<!-- gomarkdoc:embed:end -->
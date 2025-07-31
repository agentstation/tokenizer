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
    tokenizer, err := llama3.NewDefault()
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
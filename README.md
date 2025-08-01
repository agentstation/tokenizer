# Tokenizer

A collection of tokenizer implementations in Go with a unified CLI interface.

## CLI Tool

Install the tokenizer CLI:

```bash
go install github.com/agentstation/tokenizer/cmd/tokenizer@latest
```

Quick usage:

```bash
# Encode text
tokenizer llama3 encode "Hello, world!"

# Decode tokens
tokenizer llama3 decode 128000 9906 11 1917 0 128001

# Stream large files
cat document.txt | tokenizer llama3 stream
```

See [cmd/tokenizer/README.md](cmd/tokenizer/README.md) for full CLI documentation.

## Library Packages

### llama3

A Go implementation of the Llama 3 tokenizer, providing exact compatibility with the official Llama 3 tokenization.

Features:
- Byte-level BPE tokenization
- Support for all 256 special tokens
- UTF-8 handling for multilingual text
- Compatible with Llama 3, 3.1, 3.2, and 3.3 models

See [llama3/README.md](llama3/README.md) for detailed usage.

## Installation

```bash
go get github.com/agentstation/tokenizer/llama3
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/agentstation/tokenizer/llama3"
)

func main() {
    tokenizer, err := llama3.New()
    if err != nil {
        panic(err)
    }
    
    // Encode text to tokens
    tokens := tokenizer.Encode("Hello world!", nil)
    fmt.Printf("Tokens: %v\n", tokens)
    
    // Decode tokens back to text
    text := tokenizer.Decode(tokens)
    fmt.Printf("Text: %s\n", text)
}
```

## License

MIT
# Tokenizer

A collection of tokenizer implementations in Go.

## Packages

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
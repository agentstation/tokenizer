# Llama 3 Development Tools

This directory contains development and testing tools specific to the Llama 3 tokenizer implementation.

## Available Tools

### profile

Performance profiling tool for the Llama 3 tokenizer.

```bash
cd profile
go run main.go -iterations 10000 -text mixed
```

Options:
- `-cpuprofile`: Write CPU profile to file
- `-memprofile`: Write memory profile to file  
- `-iterations`: Number of tokenization iterations (default: 10000)
- `-text`: Text type to test: ascii, unicode, whitespace, code, mixed, large

Example profiling session:
```bash
# Generate CPU profile
go run main.go -cpuprofile cpu.prof -iterations 50000

# Analyze with pprof
go tool pprof cpu.prof
```

### generate-vectors

Generates test vectors by comparing with the reference JavaScript implementation.

```bash
cd generate-vectors
go run main.go -output test_vectors.jsonl -count 100
```

Prerequisites:
- Node.js installed
- Llama 3 JavaScript tokenizer available at:
  `~/src/github.com/belladoreai/llama3-tokenizer-js/bundle/llama3-tokenizer-with-baked-data.js`

Options:
- `-output`: Output file for test vectors (default: test_vectors.jsonl)
- `-count`: Number of test vectors to generate (default: 100)

The tool generates a variety of test cases including:
- Edge cases (empty strings, whitespace)
- Basic text patterns
- Unicode and emoji handling
- Special character sequences
- Code snippets

## Building Tools

Each tool can be built as a standalone binary:

```bash
# Build profile tool
cd profile
go build -o llama3-profile

# Build vector generator
cd generate-vectors  
go build -o llama3-generate-vectors
```

## Adding New Tools

When adding new Llama 3-specific tools:

1. Create a new directory under `tools/`
2. Add a `main.go` with your tool implementation
3. Update this README with usage instructions
4. Consider adding a `README.md` in your tool directory for detailed documentation
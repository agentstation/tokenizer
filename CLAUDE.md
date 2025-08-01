# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Essential Build and Development Commands

### Core Development Commands
```bash
# Default command - shows help
make

# Run all tests
make test

# Run tests with race detector
make test-race

# Run a single test
go test ./llama3 -run TestTokenizerEncode -v

# Run benchmarks
make bench
go test ./llama3 -bench=BenchmarkEncode -run=^$

# Build the CLI binary
make build

# Install the tokenizer CLI
make install

# Generate documentation (updates README files with gomarkdoc)
make generate

# Format code
make fmt-all

# Run linter
make lint

# Generate coverage report
make coverage
```

### Release Commands
```bash
# Create and push a version tag
make tag VERSION=v1.0.0
git push origin v1.0.0

# Test release process locally
make release-snapshot

# Build for all platforms
make build-all
```

### Development Tools
```bash
# Start development with hot-reload
make dev

# Run devbox shell for consistent environment
make devbox

# Install all development tools
make install-tools

# Install pre-commit hooks
make install-pre-commit
```

## High-Level Architecture

### Package Structure

The tokenizer project is organized into modular packages with clear separation of concerns:

1. **cmd/tokenizer/** - CLI implementation
   - `main.go` - Entry point with build variables (version, commit, buildDate, goVersion)
   - `root.go` - Root command and version command
   - Subcommands are delegated to individual tokenizer implementations (e.g., llama3)

2. **llama3/** - Core Llama 3 tokenizer implementation
   - `tokenizer.go` - Main tokenizer struct implementing Encoder/Decoder interfaces
   - `scanner.go` - Streaming tokenization API following bufio.Scanner pattern
   - `vocab.go` - Vocabulary management with embedded data
   - `options.go` - Configuration options for encoding/decoding
   - `constants.go` - Token IDs and vocabulary size constants
   - `errors.go` - Custom error types

3. **llama3/internal/** - Internal implementation details
   - `bpe/` - Byte Pair Encoding algorithm with caching
   - `pretokenizer/` - State machine for pre-tokenization (regex-free)
   - `vocabulary/` - Vocabulary data loading and management
   - `encoding/` - Byte-to-unicode encoding utilities
   - `tokens/` - Special token handling

4. **llama3/cmd/llama3/** - Llama3-specific CLI commands
   - `encode.go` - Text encoding command
   - `decode.go` - Token decoding command
   - `stream.go` - Streaming tokenization command
   - `info.go` - Tokenizer information command

### Key Architectural Decisions

1. **Interface-Based Design**: The tokenizer uses small, focused interfaces (Encoder, Decoder, Scanner, PreTokenizer, BPE, Cache) to allow for testing and future extensibility.

2. **Embedded Data**: Vocabulary and merge data are embedded at compile time using `go:embed`, eliminating runtime file dependencies while allowing custom data loading through interfaces.

3. **State Machine Pre-tokenization**: Instead of using regex (which has Go/JS incompatibilities), a custom state machine implements the pre-tokenization pattern matching exactly.

4. **Performance Optimizations**:
   - Object pooling with `sync.Pool` for state machines
   - LRU caching for BPE results
   - Pre-allocated slices and careful memory management
   - Zero-allocation methods for performance-critical paths

5. **Streaming Support**: The Scanner interface provides memory-efficient tokenization for large texts with proper UTF-8 boundary handling.

### Build System Integration

The Makefile uses LDFLAGS to embed build information:
```makefile
LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(BUILD_DATE) -X main.goVersion=$(GO_VERSION)"
```

These variables are defined in `cmd/tokenizer/main.go` and displayed by the `version` command.

### Testing Strategy

The project uses multiple testing approaches:
- Unit tests for individual components
- Integration tests for full tokenization pipeline
- Comparison tests against JavaScript reference implementation
- Benchmark tests for performance monitoring
- Test vectors (100+ cases) for edge case validation

### Documentation Generation

The project uses gomarkdoc with embed tags to maintain documentation:
- `generate.go` files in each package configure gomarkdoc
- README files contain `<!-- gomarkdoc:embed:start -->` tags
- Running `make generate` updates documentation without overwriting manual content

## Important Notes

- When modifying the tokenizer, always run comparison tests to ensure JavaScript compatibility
- The state machine in `internal/pretokenizer/state_machine.go` is critical for correctness - changes require careful testing
- Build variables must be in the `main` package (not subpackages) for LDFLAGS to work
- The project follows conventional commit messages for automated changelog generation
- Pre-commit hooks will run automatically - ensure they pass before committing
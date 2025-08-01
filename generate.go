// Package tokenizer provides a collection of high-performance tokenizer implementations.
package tokenizer

// Generate documentation for the root package
//go:generate gomarkdoc -o README.md -e . --embed --repository.url https://github.com/agentstation/tokenizer --repository.default-branch master --repository.path /

// Generate documentation for the llama3 package
//go:generate gomarkdoc -o ./llama3/README.md -e ./llama3 --embed --repository.url https://github.com/agentstation/tokenizer --repository.default-branch master --repository.path /llama3

// Generate documentation for the CLI package
//go:generate gomarkdoc -o ./cmd/tokenizer/README.md -e ./cmd/tokenizer --embed --repository.url https://github.com/agentstation/tokenizer --repository.default-branch master --repository.path /cmd/tokenizer

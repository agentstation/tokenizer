#!/bin/bash
set -e

echo "Generating shell completions..."

# Build the binary first if it doesn't exist
if [ ! -f ./tokenizer ]; then
    echo "Building tokenizer binary..."
    go build -o tokenizer ./cmd/tokenizer
fi

# Create completions directory
mkdir -p completions

# Generate completions for each shell
echo "Generating bash completion..."
./tokenizer completion bash > completions/tokenizer.bash

echo "Generating zsh completion..."
./tokenizer completion zsh > completions/tokenizer.zsh

echo "Generating fish completion..."
./tokenizer completion fish > completions/tokenizer.fish

echo "Completions generated successfully!"

# Clean up temporary binary
rm -f ./tokenizer
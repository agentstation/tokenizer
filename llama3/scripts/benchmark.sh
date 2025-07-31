#!/bin/bash

# Benchmark script for llama3 tokenizer

set -e

echo "=== Llama3 Tokenizer Benchmark Suite ==="
echo "Date: $(date)"
echo "Go version: $(go version)"
echo ""

# Change to the tokenizer directory
cd "$(dirname "$0")/.."

# Run comprehensive benchmarks
echo "=== Core Component Benchmarks ==="
go test -bench="BenchmarkStateMachine|BenchmarkEncode|BenchmarkDecode" -benchmem -run=^$ | grep -E "Benchmark|ns/op"

echo ""
echo "=== Text Type Benchmarks ==="
go test -bench="BenchmarkTokenize.*" -benchmem -run=^$ | grep -E "Benchmark|ns/op"

echo ""
echo "=== Size-based Benchmarks ==="
go test -bench="BenchmarkTokenize.*Text" -benchmem -run=^$ | grep -E "Benchmark|ns/op"

echo ""
echo "=== Micro-benchmarks ==="
go test -bench="BenchmarkTry|BenchmarkIs" -benchmem -run=^$ | grep -E "Benchmark|ns/op"

echo ""
echo "=== Memory Allocation Analysis ==="
go test -bench="BenchmarkMemoryAllocations" -benchmem -run=^$ | grep -E "Benchmark|allocs/op"

# Generate CPU profile for analysis
echo ""
echo "=== Generating CPU Profile ==="
echo "Note: Profile files (*.prof) are git-ignored and won't be committed"
go run cmd/profile/main.go -cpuprofile=cpu.prof -iterations=50000 -text=mixed
echo "CPU profile saved to cpu.prof"
echo "View with: go tool pprof cpu.prof"

# Generate memory profile
echo ""
echo "=== Generating Memory Profile ==="
go run cmd/profile/main.go -memprofile=mem.prof -iterations=10000 -text=large
echo "Memory profile saved to mem.prof"
echo "View with: go tool pprof mem.prof"

echo ""
echo "=== Benchmark Complete ==="
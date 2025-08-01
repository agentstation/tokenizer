//go:build experiments
// +build experiments

// Package experiments contains experimental optimizations that are not included
// in the main build. To build and test these experiments, use:
//
//	go test -tags experiments ./experiments
//
// These optimizations showed promise but had compatibility or complexity issues
// that prevented their inclusion in the main codebase.
package experiments

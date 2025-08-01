// Package vocabulary contains embedded vocabulary data files for the Llama 3 tokenizer.
// These files are from the llama3-tokenizer-js project:
// https://github.com/belladoreai/llama3-tokenizer-js
package vocabulary

import (
	_ "embed"
)

// EmbeddedVocabulary contains the base64-encoded vocabulary data.
// This includes all 128,256 tokens (128,000 regular + 256 special tokens).
//
//go:embed vocab_base64.txt
var EmbeddedVocabulary string

// EmbeddedMergeRules contains the base64-encoded binary merge data.
// Each merge is represented by two 17-bit integers packed into bytes.
//
//go:embed merges_binary.txt
var EmbeddedMergeRules string
//go:build embed

package llama3

import (
	_ "embed"
)

// Note: To use embedded data, place the vocab_base64.txt and merges_binary.txt files
// in the llama3 directory and build with -tags embed
//
// The files can be obtained from the llama3-tokenizer-js project:
// https://github.com/belladoreai/llama3-tokenizer-js

//go:embed vocab_base64.txt
var embeddedVocabBase64 string

//go:embed merges_binary.txt
var embeddedMergesBinary string

func init() {
	// Use embedded data if available
	if len(embeddedVocabBase64) > 0 {
		defaultVocabBase64 = embeddedVocabBase64
	}
	if len(embeddedMergesBinary) > 0 {
		defaultMergesBinary = embeddedMergesBinary
	}
}

// Package llama3 implements the Llama 3 tokenizer in Go.
// This file contains all constants used throughout the tokenizer implementation.
package llama3

// Vocabulary sizes.
const (
	baseVocabSize     = 128000 // Base vocabulary size
	specialTokenCount = 256    // Number of special tokens
	totalVocabSize    = baseVocabSize + specialTokenCount
)

// Cache configuration.
const (
	defaultCacheSize = 0 // 0 means unlimited
)

// BPE configuration.
const (
	estimatedTokensPerCharacter = 4 // Rough estimate for initial slice capacity
	bytesPerMerge               = 3 // Number of bytes to read for each merge
)

// Special token constants.
const (
	beginOfTextToken = "<|begin_of_text|>"
	endOfTextToken   = "<|end_of_text|>" // #nosec G101 - Not a credential, just a special token marker
)

// Package llama3 implements the Llama 3 tokenizer in Go.
// This file contains all constants used throughout the tokenizer implementation.
package llama3

// Vocabulary sizes
const (
	baseVocabSize    = 128000 // Base vocabulary size
	specialTokenCount = 256   // Number of special tokens
	totalVocabSize   = baseVocabSize + specialTokenCount
)

// Token IDs for special tokens
const (
	beginOfTextTokenID = 128000
	endOfTextTokenID   = 128001
)

// Pre-tokenization limits
const (
	maxNumberLength = 3 // Maximum consecutive digits in a single token
)

// Pool configuration
const (
	defaultStateMachineTokenCapacity = 32   // Initial capacity for state machine tokens
	defaultTokenBufferCapacity       = 64   // Initial capacity for token buffers
	maxPooledTokenBufferCapacity     = 1024 // Maximum capacity for pooled token buffers
)

// Cache configuration
const (
	defaultCacheSize = 0 // 0 means unlimited
)

// BPE configuration
const (
	estimatedTokensPerCharacter = 4 // Rough estimate for initial slice capacity
	bytesPerMerge              = 3  // Number of bytes to read for each merge
)

// Merge data configuration
const (
	bitsPerMergeID = 17 // Number of bits used to encode each merge ID
)

// Character mappings for byte-level encoding
const (
	asciiPrintableStart = '!'  // First printable ASCII character
	asciiPrintableEnd   = '~'  // Last printable ASCII character
	extendedAsciiStart1 = '¡'  // First extended ASCII range start
	extendedAsciiEnd1   = '¬'  // First extended ASCII range end
	extendedAsciiStart2 = '®'  // Second extended ASCII range start
	extendedAsciiEnd2   = 'ÿ'  // Second extended ASCII range end
	unicodeOffset       = 256  // Offset for mapping non-printable bytes
)

// Special token constants
const (
	beginOfTextToken = "<|begin_of_text|>"
	endOfTextToken   = "<|end_of_text|>"
)
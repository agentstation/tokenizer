// Package llama3 implements the Llama 3 tokenizer in Go.
// It provides exact compatibility with the official Llama 3 tokenization,
// supporting byte-level BPE tokenization with all special tokens.
package llama3

import (
	"github.com/agentstation/tokenizer/llama3/internal/bpe"
	"github.com/agentstation/tokenizer/llama3/internal/encoding"
	"github.com/agentstation/tokenizer/llama3/internal/pretokenizer"
	"github.com/agentstation/tokenizer/llama3/internal/tokens"
)

// Internal utility functions and variables

// Exposed encoding functions for backward compatibility
var (
	// encodeBytes converts UTF-8 bytes to the custom byte-level representation.
	encodeBytes = encoding.EncodeBytes

	// decodeTokenBytes converts a token string back to UTF-8 bytes.
	decodeTokenBytes = encoding.DecodeTokenBytes
)

// Special token handling
var (
	// specialTokenRegex matches Llama 3 special tokens
	specialTokenRegex = tokens.SpecialTokenRegex

	// optimisticSpecialTokenRegex matches any pattern that looks like a special token
	optimisticSpecialTokenRegex = tokens.OptimisticSpecialTokenRegex
)

// getDefaultSpecialTokens returns all Llama 3 special tokens in order.
func getDefaultSpecialTokens() []string {
	return tokens.GetDefaultSpecialTokens(specialTokenCount)
}

// isSpecialToken checks if a string is in the special token format.
func isSpecialToken(token string) bool {
	return tokens.IsSpecialToken(token)
}

// splitBySpecialTokens splits text by special tokens while preserving the tokens.
var splitBySpecialTokens = tokens.SplitBySpecialTokens

// Encoder is the interface for encoding text to tokens.
// This interface is useful for testing and creating mock implementations.
type Encoder interface {
	// Encode converts text to a sequence of token IDs.
	Encode(text string, opts *EncodeOptions) []int
}

// Decoder is the interface for decoding tokens to text.
// This interface is useful for testing and creating mock implementations.
type Decoder interface {
	// Decode converts a sequence of token IDs back to text.
	Decode(tokens []int) string
}

// EncoderFunc is an adapter to allow ordinary functions to be used as Encoders.
// This is useful for creating mock encoders in tests.
type EncoderFunc func(text string, opts *EncodeOptions) []int

// Encode calls f(text, opts).
func (f EncoderFunc) Encode(text string, opts *EncodeOptions) []int {
	return f(text, opts)
}

// DecoderFunc is an adapter to allow ordinary functions to be used as Decoders.
// This is useful for creating mock decoders in tests.
type DecoderFunc func(tokens []int) string

// Decode calls f(tokens).
func (f DecoderFunc) Decode(tokens []int) string {
	return f(tokens)
}

// BPE is the interface for Byte Pair Encoding processing.
// BPE merges frequently occurring character pairs to create subword tokens.
type BPE interface {
	// EncodeBPE applies byte pair encoding to a pre-tokenized string.
	// Returns a slice of token IDs representing the encoded text.
	EncodeBPE(pretoken string) []int
}

// PreTokenizer is the interface for pre-tokenization.
// Pre-tokenization splits text into words, numbers, punctuation, etc.
// before the BPE algorithm is applied.
type PreTokenizer interface {
	// PreTokenize splits text into pre-tokens according to the tokenizer's rules.
	// Returns a slice of pre-token strings ready for BPE processing.
	PreTokenize(text string) []string
}

// Tokenizer implements the Llama 3 BPE tokenizer.
type Tokenizer struct {
	tokens      []string       // Token ID to text mapping
	tokenLookup map[string]int // Text to token ID mapping
	mergeRules  map[string]int // BPE merge rules with priorities

	// Cache for BPE results
	cache     bpeCache
	cacheSize int // Maximum cache size (0 = unlimited)
}

// EncodeOptions controls the encoding behavior.
type EncodeOptions struct {
	// BOS adds the beginning-of-text token if true (default: true)
	BOS bool
	// EOS adds the end-of-text token if true (default: true)
	EOS bool
}

// defaultEncodeOptions returns the default encoding options.
func defaultEncodeOptions() *EncodeOptions {
	return &EncodeOptions{
		BOS: true,
		EOS: true,
	}
}

// New creates a new Llama 3 tokenizer with the given options.
// If no options are provided, the default Llama 3 vocabulary and settings will be used.
//
// Example:
//
//	tokenizer, err := llama3.New()
//	if err != nil {
//	    return err
//	}
//
//	// With custom vocabulary:
//	tokenizer, err := llama3.New(
//	    llama3.WithVocabulary(customVocab),
//	    llama3.WithMerges(customMerges),
//	)
//
//	// With cache size limit:
//	tokenizer, err := llama3.New(
//	    llama3.WithCacheSize(1000),
//	)
func New(opts ...Option) (*Tokenizer, error) {
	// Default configuration
	config := &config{
		dataLoader:    nil,
		specialTokens: nil,
		cacheSize:     defaultCacheSize,
	}

	// Apply options to configuration
	for _, opt := range opts {
		if err := opt(config); err != nil {
			return nil, err
		}
	}

	// Create tokenizer with configured cache size
	t := &Tokenizer{
		cacheSize: config.cacheSize,
	}

	// Initialize cache based on size
	if t.cacheSize == 0 {
		t.cache = bpe.NewSimple()
	} else {
		t.cache = newLRUCache(t.cacheSize)
	}

	// Create data loader
	var vocab VocabularyDataLoader
	if config.dataLoader != nil {
		// Check if it's a file loader marker
		if marker, ok := config.dataLoader.(*fileLoaderMarker); ok {
			vocab = &fileVocabularySource{
				vocabPath:  marker.vocabPath,
				mergesPath: marker.mergesPath,
				t:          t,
			}
		} else {
			vocab = config.dataLoader
		}
	} else {
		vocab = &embeddedVocabularySource{t: t}
	}

	// Load vocabulary
	var err error
	t.tokens, err = vocab.LoadVocabulary()
	if err != nil {
		return nil, err
	}

	// Add special tokens
	specialTokens := config.specialTokens
	if specialTokens == nil {
		specialTokens = getDefaultSpecialTokens()
	}
	t.tokens = append(t.tokens, specialTokens...)

	// Build string-to-ID mapping
	t.tokenLookup = make(map[string]int, len(t.tokens))
	for id, token := range t.tokens {
		t.tokenLookup[token] = id
	}

	// Load merges
	t.mergeRules, err = vocab.LoadMerges()
	if err != nil {
		return nil, err
	}

	return t, nil
}

// Encode converts text into a sequence of token IDs.
// If opts is nil, default options will be used.
func (t *Tokenizer) Encode(text string, opts *EncodeOptions) []int {
	if opts == nil {
		opts = defaultEncodeOptions()
	}

	output := make([]int, 0, len(text)/estimatedTokensPerCharacter)

	// Add beginning-of-text token
	if opts.BOS {
		if id, err := t.GetSpecialTokenID(beginOfTextToken); err == nil {
			output = append(output, id)
		}
	}

	// Split by special tokens first
	specialSplits := splitBySpecialTokens(text, specialTokenRegex)

	for _, specialSplit := range specialSplits {
		// Check if this is a special token
		if specialTokenRegex.MatchString(specialSplit) && t.tokenLookup[specialSplit] != 0 {
			output = append(output, t.tokenLookup[specialSplit])
			continue
		}

		// Not a special token, process normally
		// Pre-tokenize using regex
		pretokens := t.pretokenize(specialSplit)

		// Process each pretoken
		for _, pretoken := range pretokens {
			if pretoken == "" {
				continue
			}

			// Perform BPE on the pretoken
			tokenIDs := t.performBPE(pretoken)
			output = append(output, tokenIDs...)
		}
	}

	// Add end-of-text token
	if opts.EOS {
		if id, err := t.GetSpecialTokenID(endOfTextToken); err == nil {
			output = append(output, id)
		}
	}

	return output
}

// EncodeBytes converts bytes into a sequence of token IDs.
// This avoids string conversion overhead for binary data.
func (t *Tokenizer) EncodeBytes(data []byte, opts *EncodeOptions) []int {
	return t.Encode(string(data), opts)
}

// AppendTokens appends tokens to dst, avoiding allocations when possible.
// dst can be nil, in which case a new slice is allocated.
// The resulting slice is returned and may have a different backing array than dst.
func (t *Tokenizer) AppendTokens(dst []int, text string, opts *EncodeOptions) []int {
	if opts == nil {
		opts = defaultEncodeOptions()
	}

	// Reserve capacity if dst is nil or too small
	estimatedTokens := len(text)/estimatedTokensPerCharacter + 2 // +2 for BOS/EOS
	if cap(dst) < len(dst)+estimatedTokens {
		newDst := make([]int, len(dst), len(dst)+estimatedTokens)
		copy(newDst, dst)
		dst = newDst
	}

	// Add beginning-of-text token
	if opts.BOS {
		if id, err := t.GetSpecialTokenID(beginOfTextToken); err == nil {
			dst = append(dst, id)
		}
	}

	// Split by special tokens first
	specialSplits := splitBySpecialTokens(text, specialTokenRegex)

	for _, specialSplit := range specialSplits {
		// Check if this is a special token
		if specialTokenRegex.MatchString(specialSplit) && t.tokenLookup[specialSplit] != 0 {
			dst = append(dst, t.tokenLookup[specialSplit])
			continue
		}

		// Not a special token, process normally
		// Pre-tokenize using regex
		pretokens := t.pretokenize(specialSplit)

		// Process each pretoken
		for _, pretoken := range pretokens {
			if pretoken == "" {
				continue
			}

			// Perform BPE on the pretoken
			tokenIDs := t.performBPE(pretoken)
			dst = append(dst, tokenIDs...)
		}
	}

	// Add end-of-text token
	if opts.EOS {
		if id, err := t.GetSpecialTokenID(endOfTextToken); err == nil {
			dst = append(dst, id)
		}
	}

	return dst
}

// Decode converts a sequence of token IDs back into text.
func (t *Tokenizer) Decode(tokenIDs []int) string {
	return string(t.DecodeBytes(tokenIDs))
}

// DecodeBytes converts a sequence of token IDs back to UTF-8 bytes.
// This avoids string allocation and is useful for performance-critical paths.
func (t *Tokenizer) DecodeBytes(tokenIDs []int) []byte {
	utf8Bytes := make([]byte, 0, len(tokenIDs)*bytesPerMerge)

	for _, tokenID := range tokenIDs {
		if tokenID < 0 || tokenID >= len(t.tokens) {
			continue // Skip invalid token IDs
		}

		tokenString := t.tokens[tokenID]
		// Convert from custom byte representation back to UTF-8
		decodedBytes := decodeTokenBytes(tokenString)
		utf8Bytes = append(utf8Bytes, decodedBytes...)
	}

	return utf8Bytes
}

// GetSpecialTokenID returns the token ID for a special token string.
func (t *Tokenizer) GetSpecialTokenID(token string) (int, error) {
	if !isSpecialToken(token) {
		return 0, NewTokenError("validate special token", token, ErrInvalidToken)
	}

	id, ok := t.tokenLookup[token]
	if !ok {
		return 0, NewTokenError("get special token ID", token, ErrTokenNotFound)
	}

	return id, nil
}

// OptimisticCount returns the token count assuming anything that looks like
// a special token is actually a special token. This is useful for fine-tuned
// models with modified special tokens.
func (t *Tokenizer) OptimisticCount(text string) int {
	// Use optimistic regex that matches any <|...|> pattern
	output := make([]int, 0, len(text)/estimatedTokensPerCharacter)

	// Always add BOS and EOS for optimistic count
	if id, err := t.GetSpecialTokenID(beginOfTextToken); err == nil {
		output = append(output, id)
	}

	// Split by optimistic special token regex
	specialSplits := splitBySpecialTokens(text, optimisticSpecialTokenRegex)

	for _, specialSplit := range specialSplits {
		// Check if this looks like a special token
		if optimisticSpecialTokenRegex.MatchString(specialSplit) {
			// For optimistic count, we count it as 1 token even if not in vocab
			if id, ok := t.tokenLookup[specialSplit]; ok {
				output = append(output, id)
			} else {
				// Use a fallback token ID (just use 1 for counting)
				output = append(output, 1)
			}
			continue
		}

		// Not a special token, process normally
		pretokens := t.pretokenize(specialSplit)

		for _, pretoken := range pretokens {
			if pretoken == "" {
				continue
			}

			tokenIDs := t.performBPE(pretoken)
			output = append(output, tokenIDs...)
		}
	}

	// Add EOS
	if id, err := t.GetSpecialTokenID(endOfTextToken); err == nil {
		output = append(output, id)
	}

	return len(output)
}

// VocabSize returns the size of the vocabulary including special tokens.
func (t *Tokenizer) VocabSize() int {
	return len(t.tokens)
}

// getMergeIdentifier creates a merge identifier string from two token IDs.
func (t *Tokenizer) getMergeIdentifier(firstTokenID, secondTokenID int) string {
	return t.tokens[firstTokenID] + " " + t.tokens[secondTokenID]
}

// Ensure Tokenizer implements Encoder and Decoder interfaces
var (
	_ Encoder      = (*Tokenizer)(nil)
	_ Decoder      = (*Tokenizer)(nil)
	_ PreTokenizer = (*Tokenizer)(nil)
	_ BPE          = (*Tokenizer)(nil)
)

// Cache is the interface for caching BPE results.
// BPE tokenization can be expensive for repeated text patterns,
// so caching improves performance significantly.
//
// The cache key is typically the pre-tokenized text string,
// and the value is the slice of token IDs produced by BPE.
//
// Implementations should be thread-safe if the tokenizer
// will be used concurrently.
type Cache interface {
	// Get retrieves a cached BPE result.
	// Returns the token IDs and true if found, or nil and false if not cached.
	Get(key string) ([]int, bool)

	// Put stores a BPE result in the cache.
	// The implementation may evict old entries based on its eviction policy.
	Put(key string, value []int)
}

// bpeCache is an alias for Cache used internally.
// This maintains backward compatibility with existing code.
type bpeCache = Cache

// Implementation methods

// performBPE executes the Byte Pair Encoding algorithm on a pre-token.
// It iteratively merges the most frequent pairs of adjacent tokens according
// to the learned merge rules. Results are cached for efficiency.
func (t *Tokenizer) performBPE(pretoken string) []int {
	// Create a BPE processor with the tokenizer's data
	processor := &bpe.Processor{
		Tokens:      t.tokens,
		TokenLookup: t.tokenLookup,
		MergeRules:  t.mergeRules,
		Cache:       t.cache,
	}

	return processor.PerformBPE(pretoken)
}

// EncodeBPE implements the BPE interface.
func (t *Tokenizer) EncodeBPE(pretoken string) []int {
	return t.performBPE(pretoken)
}

// newLRUCache creates a new LRU cache with the given capacity.
// If capacity is 0, the cache is unlimited (falls back to simple map).
func newLRUCache(capacity int) Cache {
	return bpe.NewLRU(capacity)
}

// pretokenize performs the pre-tokenization step using state machine
// and byte-level encoding.
func (t *Tokenizer) pretokenize(text string) []string {
	// Use pooled state machine for better performance
	parts := pretokenizer.Tokenize(text)

	// Apply byte-level encoding to each part
	encoded := make([]string, len(parts))
	for i, part := range parts {
		encoded[i] = encodeBytes([]byte(part))
	}

	return encoded
}

// PreTokenize implements the PreTokenizer interface.
func (t *Tokenizer) PreTokenize(text string) []string {
	return t.pretokenize(text)
}

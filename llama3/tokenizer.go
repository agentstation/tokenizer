// Package llama3 implements the Llama 3 tokenizer in Go.
// It provides exact compatibility with the official Llama 3 tokenization,
// supporting byte-level BPE tokenization with all special tokens.
package llama3

// tokenizerConfig holds configuration during tokenizer creation
type tokenizerConfig struct {
	vocabBase64   string
	mergesBinary  string
	specialTokens []string
	cacheSize     int
}

// Tokenizer implements the Llama 3 BPE tokenizer.
type Tokenizer struct {
	vocabByID     []string
	vocabByString map[string]int
	merges        map[string]int

	// Cache for BPE results
	cache     bpeCache
	cacheSize int // Maximum cache size (0 = unlimited)
}

// bpeCache defines the interface for BPE result caching.
type bpeCache interface {
	get(key string) ([]int, bool)
	put(key string, value []int)
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
	config := &tokenizerConfig{
		vocabBase64:   "",
		mergesBinary:  "",
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
		t.cache = &simpleCache{cache: make(map[string][]int)}
	} else {
		t.cache = newLRUCache(t.cacheSize)
	}

	// Load vocabulary
	var err error
	vocabBase64 := config.vocabBase64
	if vocabBase64 == "" {
		vocabBase64 = defaultVocabBase64
		if vocabBase64 == "" {
			return nil, NewDataError("load vocabulary", "", ErrDataNotFound)
		}
	}
	t.vocabByID, err = decodeVocabulary(vocabBase64)
	if err != nil {
		return nil, NewDataError("decode vocabulary", "", err)
	}

	// Add special tokens
	specialTokens := config.specialTokens
	if specialTokens == nil {
		specialTokens = getDefaultSpecialTokens()
	}
	t.vocabByID = append(t.vocabByID, specialTokens...)

	// Build string-to-ID mapping
	t.vocabByString = make(map[string]int, len(t.vocabByID))
	for id, token := range t.vocabByID {
		t.vocabByString[token] = id
	}

	// Load merges
	mergesBinary := config.mergesBinary
	if mergesBinary == "" {
		mergesBinary = defaultMergesBinary
		if mergesBinary == "" {
			return nil, NewDataError("load merges", "", ErrDataNotFound)
		}
	}
	t.merges, err = t.decompressMerges(mergesBinary)
	if err != nil {
		return nil, NewDataError("decompress merges", "", err)
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
		if specialTokenRegex.MatchString(specialSplit) && t.vocabByString[specialSplit] != 0 {
			output = append(output, t.vocabByString[specialSplit])
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

// Decode converts a sequence of token IDs back into text.
func (t *Tokenizer) Decode(tokenIDs []int) string {
	utf8Bytes := make([]byte, 0, len(tokenIDs)*bytesPerMerge)

	for _, tokenID := range tokenIDs {
		if tokenID < 0 || tokenID >= len(t.vocabByID) {
			continue // Skip invalid token IDs
		}

		tokenString := t.vocabByID[tokenID]
		// Convert from custom byte representation back to UTF-8
		decodedBytes := decodeTokenBytes(tokenString)
		utf8Bytes = append(utf8Bytes, decodedBytes...)
	}

	return string(utf8Bytes)
}

// GetSpecialTokenID returns the token ID for a special token string.
func (t *Tokenizer) GetSpecialTokenID(token string) (int, error) {
	if !isSpecialToken(token) {
		return 0, NewTokenError("validate special token", token, ErrInvalidToken)
	}

	id, ok := t.vocabByString[token]
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
			if id, ok := t.vocabByString[specialSplit]; ok {
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
	return len(t.vocabByID)
}

// GetMergeIdentifier creates a merge identifier string from two token IDs.
func (t *Tokenizer) getMergeIdentifier(firstTokenID, secondTokenID int) string {
	return t.vocabByID[firstTokenID] + " " + t.vocabByID[secondTokenID]
}

package llama3

// Option is a functional option for configuring a Tokenizer.
type Option func(*tokenizerConfig) error

// WithVocabulary sets custom vocabulary data for the tokenizer.
// The vocabulary should be base64-encoded.
func WithVocabulary(vocabBase64 string) Option {
	return func(cfg *tokenizerConfig) error {
		if vocabBase64 == "" {
			return NewConfigError("vocabulary", "empty string", ErrInvalidToken)
		}
		cfg.vocabBase64 = vocabBase64
		return nil
	}
}

// WithMerges sets custom merge data for the tokenizer.
// The merges should be base64-encoded binary data.
func WithMerges(mergesBinary string) Option {
	return func(cfg *tokenizerConfig) error {
		if mergesBinary == "" {
			return NewConfigError("merges", "empty string", ErrInvalidToken)
		}
		cfg.mergesBinary = mergesBinary
		return nil
	}
}

// WithSpecialTokens sets custom special tokens for the tokenizer.
// If nil, the default Llama 3 special tokens will be used.
func WithSpecialTokens(tokens []string) Option {
	return func(cfg *tokenizerConfig) error {
		// Validate special tokens
		for i, token := range tokens {
			if !isSpecialToken(token) {
				return NewConfigError("special_tokens", token, 
					NewTokenError("validate", token, ErrInvalidToken))
			}
			// Check for duplicates
			for j := i + 1; j < len(tokens); j++ {
				if token == tokens[j] {
					return NewConfigError("special_tokens", token,
						NewTokenError("duplicate", token, ErrInvalidToken))
				}
			}
		}
		cfg.specialTokens = tokens
		return nil
	}
}

// WithCacheSize sets the maximum size of the BPE cache.
// Set to 0 to disable caching. Default is unlimited.
func WithCacheSize(size int) Option {
	return func(cfg *tokenizerConfig) error {
		if size < 0 {
			return NewConfigError("cache_size", size, ErrInvalidToken)
		}
		cfg.cacheSize = size
		return nil
	}
}

// WithDataFiles loads vocabulary and merge data from files.
func WithDataFiles(vocabPath, mergesPath string) Option {
	return func(cfg *tokenizerConfig) error {
		if err := LoadDataFromFiles(vocabPath, mergesPath); err != nil {
			return err
		}
		// The LoadDataFromFiles sets global defaults, so we need to capture them
		cfg.vocabBase64 = defaultVocabBase64
		cfg.mergesBinary = defaultMergesBinary
		return nil
	}
}


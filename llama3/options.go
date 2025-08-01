package llama3

import "strings"

// config holds configuration during tokenizer creation.
type config struct {
	dataLoader    VocabularyDataLoader
	specialTokens []string
	cacheSize     int
}

// Option is a functional option for configuring a Tokenizer.
type Option func(*config) error

// WithSpecialTokens sets custom special tokens for the tokenizer.
// If nil, the default Llama 3 special tokens will be used.
func WithSpecialTokens(tokens []string) Option {
	return func(cfg *config) error {
		// Validate special tokens - they must match the <|...|> pattern
		for i, token := range tokens {
			if len(token) < 5 || !strings.HasPrefix(token, "<|") || !strings.HasSuffix(token, "|>") {
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
	return func(cfg *config) error {
		if size < 0 {
			return NewConfigError("cache_size", size, ErrInvalidToken)
		}
		cfg.cacheSize = size
		return nil
	}
}

// WithDataLoader sets a custom data loader for the tokenizer.
// This allows loading vocabulary and merges from custom sources.
func WithDataLoader(loader VocabularyDataLoader) Option {
	return func(cfg *config) error {
		if loader == nil {
			return NewConfigError("data_loader", nil, ErrInvalidToken)
		}
		cfg.dataLoader = loader
		return nil
	}
}

// WithDataFiles loads vocabulary and merges from files instead of embedded data.
// The vocabulary file should contain base64-encoded vocabulary data.
// The merges file should contain base64-encoded binary merge data.
func WithDataFiles(vocabPath, mergesPath string) Option {
	return func(cfg *config) error {
		// This will be handled in tokenizer initialization
		cfg.dataLoader = &fileLoaderMarker{
			vocabPath:  vocabPath,
			mergesPath: mergesPath,
		}
		return nil
	}
}

// Package llama3 implements the Llama 3 tokenizer in Go.
// It provides exact compatibility with the official Llama 3 tokenization,
// supporting byte-level BPE tokenization with all special tokens.
package llama3

import (
	"fmt"
	"sync"
)

// Tokenizer implements the Llama 3 BPE tokenizer.
type Tokenizer struct {
	vocabByID     []string
	vocabByString map[string]int
	merges        map[string]int
	
	// Cache for pre-tokenization results
	cache      map[string][]int
	cacheMutex sync.RWMutex
}

// EncodeOptions controls the encoding behavior.
type EncodeOptions struct {
	// BOS adds the beginning-of-text token if true (default: true)
	BOS bool
	// EOS adds the end-of-text token if true (default: true)
	EOS bool
}

// DefaultEncodeOptions returns the default encoding options.
func DefaultEncodeOptions() *EncodeOptions {
	return &EncodeOptions{
		BOS: true,
		EOS: true,
	}
}

// New creates a new Llama 3 tokenizer with custom vocabulary and merge data.
// If vocabBase64 or mergesBinary are empty, the default Llama 3 data will be used.
// specialTokens can be nil to use the default Llama 3 special tokens.
func New(vocabBase64, mergesBinary string, specialTokens []string) (*Tokenizer, error) {
	t := &Tokenizer{
		cache: make(map[string][]int),
	}
	
	// Load vocabulary
	var err error
	if vocabBase64 == "" {
		vocabBase64 = defaultVocabBase64
		if vocabBase64 == "" {
			return nil, fmt.Errorf("no vocabulary data available. Please ensure the data files are properly loaded. " +
				"See https://github.com/agentstation/tokenizer/llama3#data-files for instructions")
		}
	}
	t.vocabByID, err = decodeVocabulary(vocabBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode vocabulary: %w", err)
	}
	
	// Add special tokens
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
	if mergesBinary == "" {
		mergesBinary = defaultMergesBinary
		if mergesBinary == "" {
			return nil, fmt.Errorf("no merge data available. Please ensure the data files are properly loaded. " +
				"See https://github.com/agentstation/tokenizer/llama3#data-files for instructions")
		}
	}
	t.merges, err = t.decompressMerges(mergesBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress merges: %w", err)
	}
	
	return t, nil
}

// NewDefault creates a new tokenizer with the default Llama 3 vocabulary and merges.
func NewDefault() (*Tokenizer, error) {
	return New("", "", nil)
}

// Encode converts text into a sequence of token IDs.
// If opts is nil, default options will be used.
func (t *Tokenizer) Encode(text string, opts *EncodeOptions) []int {
	if opts == nil {
		opts = DefaultEncodeOptions()
	}
	
	output := make([]int, 0, len(text)/4) // Rough estimate
	
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
	utf8Bytes := make([]byte, 0, len(tokenIDs)*3) // Rough estimate
	
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
		return 0, fmt.Errorf("invalid special token format: %s", token)
	}
	
	id, ok := t.vocabByString[token]
	if !ok {
		return 0, fmt.Errorf("special token not found: %s", token)
	}
	
	return id, nil
}

// OptimisticCount returns the token count assuming anything that looks like
// a special token is actually a special token. This is useful for fine-tuned
// models with modified special tokens.
func (t *Tokenizer) OptimisticCount(text string) int {
	// Use optimistic regex that matches any <|...|> pattern
	output := make([]int, 0, len(text)/4)
	
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
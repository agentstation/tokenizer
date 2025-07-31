// +build experiments

package llama3

import (
	"strings"
	"sync"
	"unicode"
)

// Optimized state machine with jump table dispatch
type OptimizedStateMachine struct {
	input    []rune
	position int
	tokens   []string
	jumpTable [256]func(*OptimizedStateMachine) string
}

var optimizedSmPool = sync.Pool{
	New: func() interface{} {
		sm := &OptimizedStateMachine{
			tokens: make([]string, 0, 64),
		}
		sm.initJumpTable()
		return sm
	},
}

// TokenizeOptimized uses the optimized state machine with jump table
func TokenizeOptimized(text string) []string {
	sm := getOptimizedStateMachine(text)
	defer putOptimizedStateMachine(sm)
	
	return sm.TokenizeWithStateMachine()
}

func getOptimizedStateMachine(text string) *OptimizedStateMachine {
	sm := optimizedSmPool.Get().(*OptimizedStateMachine)
	sm.input = []rune(text)
	sm.position = 0
	sm.tokens = sm.tokens[:0]
	return sm
}

func putOptimizedStateMachine(sm *OptimizedStateMachine) {
	sm.input = nil
	sm.position = 0
	sm.tokens = sm.tokens[:0]
	optimizedSmPool.Put(sm)
}

// initJumpTable initializes the jump table for first-character dispatch
func (sm *OptimizedStateMachine) initJumpTable() {
	// Default handler for most characters
	defaultHandler := func(sm *OptimizedStateMachine) string {
		return sm.tryWordWithPrefixOptimized()
	}
	
	// Initialize all entries to default
	for i := range sm.jumpTable {
		sm.jumpTable[i] = defaultHandler
	}
	
	// Special handlers for common characters
	// Whitespace (but not space - it can be a word prefix)
	sm.jumpTable['	'] = (*OptimizedStateMachine).tryWhitespaceOptimized
	sm.jumpTable['
'] = (*OptimizedStateMachine).tryWhitespaceOptimized
	sm.jumpTable[''] = (*OptimizedStateMachine).tryWhitespaceOptimized
	
	// Apostrophe for contractions
	sm.jumpTable['\''] = (*OptimizedStateMachine).tryContractionOrPunctuation
	
	// Digits
	for c := '0'; c <= '9'; c++ {
		sm.jumpTable[c] = (*OptimizedStateMachine).tryNumbersOptimized
	}
	
	// Common punctuation that needs space handling
	punctuation := "!\"#$%&()*+,-./:;<=>?@[\]^_`{|}~"
	for _, p := range punctuation {
		if p < 256 {
			sm.jumpTable[p] = (*OptimizedStateMachine).tryPunctuationWithSpaceOptimized
		}
	}
}

func (sm *OptimizedStateMachine) TokenizeWithStateMachine() []string {
	// Pre-allocate based on text length heuristic
	if cap(sm.tokens) < len(sm.input)/4 {
		sm.tokens = make([]string, 0, len(sm.input)/4)
	}
	
	for sm.position < len(sm.input) {
		sm.matchNextOptimized()
	}
	
	result := make([]string, len(sm.tokens))
	copy(result, sm.tokens)
	return result
}

func (sm *OptimizedStateMachine) matchNextOptimized() {
	if sm.position >= len(sm.input) {
		return
	}
	
	// Follow the exact same pattern order as original
	
	// 1. Try contractions first
	if sm.position < len(sm.input) && sm.input[sm.position] == '\'' {
		if token := sm.tryContractionOptimized(); token != "" {
			sm.tokens = append(sm.tokens, token)
			return
		}
	}
	
	// 2. Try word with optional prefix
	if token := sm.tryWordWithPrefixOptimized(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// 3. Try numbers (1-3 digits)
	if sm.position < len(sm.input) {
		r := sm.input[sm.position]
		if r >= '0' && r <= '9' {
			if token := sm.tryNumbersOptimized(); token != "" {
				sm.tokens = append(sm.tokens, token)
				return
			}
		}
	}
	
	// 4. Try punctuation with optional space
	if token := sm.tryPunctuationWithSpaceOptimized(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// 5. Try newlines: \s*[
]+
	if token := sm.tryNewlines(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// 6. Try whitespace: \s+(?!\S) or \s+
	if token := sm.tryWhitespaceOptimized(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// This shouldn't happen - advance to avoid infinite loop
	sm.position++
}

// Optimized contraction handler
func (sm *OptimizedStateMachine) tryContractionOrPunctuation() string {
	// Fast path check for contractions
	if sm.position+1 < len(sm.input) {
		next := sm.input[sm.position+1]
		switch next {
		case 's', 'S', 't', 'T', 'r', 'R', 'v', 'V', 'm', 'M', 'l', 'L', 'd', 'D':
			return sm.tryContractionOptimized()
		}
	}
	return sm.tryPunctuationWithSpaceOptimized()
}

func (sm *OptimizedStateMachine) tryContractionOptimized() string {
	if sm.position >= len(sm.input) || sm.input[sm.position] != '\'' {
		return ""
	}
	
	// Check for valid contraction suffixes using a switch
	if sm.position+2 <= len(sm.input) {
		suffix := string(sm.input[sm.position:sm.position+2])
		switch strings.ToLower(suffix) {
		case "'s", "'t", "'m", "'d":
			sm.position += 2
			return suffix
		}
	}
	
	if sm.position+3 <= len(sm.input) {
		suffix := string(sm.input[sm.position:sm.position+3])
		switch strings.ToLower(suffix) {
		case "'re", "'ve", "'ll":
			sm.position += 3
			return suffix
		}
	}
	
	return ""
}

// Optimized whitespace handler with negative lookahead
func (sm *OptimizedStateMachine) tryWhitespaceOptimized() string {
	// First try \s+(?!\S) - whitespace not followed by non-whitespace
	start := sm.position
	
	// Fast ASCII whitespace check
	for sm.position < len(sm.input) {
		c := sm.input[sm.position]
		if c == ' ' || c == '	' || c == '
' || c == '' {
			sm.position++
		} else if c > 127 && unicode.IsSpace(c) {
			sm.position++
		} else {
			break
		}
	}
	
	// Check negative lookahead (?!\S)
	// If we're followed by non-whitespace, backtrack by one
	if sm.position < len(sm.input) && !isWhitespaceOptimized(sm.input[sm.position]) {
		if sm.position > start+1 {
			sm.position--
		}
	}
	
	// If that matched, return it
	if sm.position > start {
		return string(sm.input[start:sm.position])
	}
	
	// Otherwise try simple \s+
	sm.position = start
	for sm.position < len(sm.input) && isWhitespaceOptimized(sm.input[sm.position]) {
		sm.position++
	}
	
	return string(sm.input[start:sm.position])
}

// Optimized number handler
func (sm *OptimizedStateMachine) tryNumbersOptimized() string {
	start := sm.position
	count := 0
	
	// Fast ASCII digit check
	for sm.position < len(sm.input) && count < 3 {
		c := sm.input[sm.position]
		if c >= '0' && c <= '9' {
			sm.position++
			count++
		} else {
			break
		}
	}
	
	if count > 0 {
		return string(sm.input[start:sm.position])
	}
	return ""
}

// Optimized word with prefix handler
func (sm *OptimizedStateMachine) tryWordWithPrefixOptimized() string {
	start := sm.position
	hasPrefix := false
	
	// Check for optional prefix (fast ASCII check)
	if sm.position < len(sm.input) {
		r := sm.input[sm.position]
		if r < 128 {
			// ASCII fast path
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
				// It's a letter or number, no prefix
			} else {
				// It's a prefix character
				sm.position++
				hasPrefix = true
			}
		} else if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			// Non-ASCII prefix
			sm.position++
			hasPrefix = true
		}
	}
	
	// Match letters (with ASCII fast path)
	wordStart := sm.position
	for sm.position < len(sm.input) {
		r := sm.input[sm.position]
		if r < 128 {
			// ASCII fast path
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
				sm.position++
			} else {
				break
			}
		} else if unicode.IsLetter(r) {
			sm.position++
		} else {
			break
		}
	}
	
	// If we didn't match any letters after a prefix, reset
	if hasPrefix && sm.position == wordStart {
		sm.position = start
		return ""
	}
	
	// If we matched something, return it
	if sm.position > start {
		return string(sm.input[start:sm.position])
	}
	
	return ""
}

// Optimized punctuation with space handler
func (sm *OptimizedStateMachine) tryPunctuationWithSpaceOptimized() string {
	start := sm.position
	
	// Optional leading space (ASCII fast path)
	if sm.position < len(sm.input) {
		c := sm.input[sm.position]
		if c == ' ' {
			sm.position++
		}
	}
	
	// Match non-letter, non-number, non-space characters
	for sm.position < len(sm.input) {
		r := sm.input[sm.position]
		if r < 128 {
			// ASCII fast path
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || 
			   (r >= '0' && r <= '9') || r == ' ' || r == '	' || 
			   r == '
' || r == '' {
				break
			}
			sm.position++
		} else if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			break
		} else {
			sm.position++
		}
	}
	
	// Match trailing newlines
	for sm.position < len(sm.input) {
		c := sm.input[sm.position]
		if c == '' || c == '
' {
			sm.position++
		} else {
			break
		}
	}
	
	if sm.position > start {
		return string(sm.input[start:sm.position])
	}
	
	return ""
}

// tryNewlines matches: \s*[
]+
func (sm *OptimizedStateMachine) tryNewlines() string {
	start := sm.position
	
	// Optional leading whitespace
	for sm.position < len(sm.input) && isWhitespaceOptimized(sm.input[sm.position]) {
		sm.position++
	}
	
	// Must have at least one newline
	newlineStart := sm.position
	for sm.position < len(sm.input) {
		c := sm.input[sm.position]
		if c == '' || c == '
' {
			sm.position++
		} else {
			break
		}
	}
	
	// If we found newlines, return the match
	if sm.position > newlineStart {
		return string(sm.input[start:sm.position])
	}
	
	// Otherwise backtrack
	sm.position = start
	return ""
}

// Character classification with ASCII fast paths
func isLetterOptimized(r rune) bool {
	if r < 128 {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
	}
	return unicode.IsLetter(r)
}

func isNumberOptimized(r rune) bool {
	if r < 128 {
		return r >= '0' && r <= '9'
	}
	return unicode.IsNumber(r)
}

func isWhitespaceOptimized(r rune) bool {
	if r < 128 {
		return r == ' ' || r == '	' || r == '
' || r == ''
	}
	return unicode.IsSpace(r)
}

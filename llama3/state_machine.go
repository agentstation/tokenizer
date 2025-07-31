package llama3

import (
	"sync"
	"unicode"
)

// stateMachine implements the exact JavaScript regex behavior for pre-tokenization.
// It processes text character by character, matching patterns in a specific order
// to ensure compatibility with the reference implementation.
type stateMachine struct {
	input    []rune
	position int
	tokens   []string
}

// stateMachinePool provides a pool of reusable state machines for performance
var stateMachinePool = &sync.Pool{
	New: func() interface{} {
		return &stateMachine{
			tokens: make([]string, 0, 32), // Pre-allocate typical capacity
		}
	},
}

// tokenBufPool provides a pool of token buffers for better memory efficiency
var tokenBufPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, defaultTokenBufferCapacity)
	},
}

// getStateMachine gets a state machine from the pool
func getStateMachine(text string) *stateMachine {
	sm := stateMachinePool.Get().(*stateMachine)
	sm.input = []rune(text)
	sm.position = 0
	sm.tokens = sm.tokens[:0] // Reset slice but keep capacity
	return sm
}

// putStateMachine returns a state machine to the pool
func putStateMachine(sm *stateMachine) {
	// Clear references to allow GC
	sm.input = nil
	stateMachinePool.Put(sm)
}

// Tokenize performs pre-tokenization on the input text using a pooled state machine.
// It splits text into words, numbers, punctuation, and whitespace according to the
// Llama 3 tokenization rules. The function is optimized for memory efficiency by
// reusing state machines and token buffers from pools.
func Tokenize(text string) []string {
	sm := getStateMachine(text)

	// Use pooled token buffer for better memory efficiency
	tokens := tokenBufPool.Get().([]string)
	sm.tokens = tokens[:0]

	// Process tokens
	for sm.position < len(sm.input) {
		sm.matchNext()
	}

	// Copy results before returning to pool
	result := make([]string, len(sm.tokens))
	copy(result, sm.tokens)

	// Return token buffer to pool
	if cap(sm.tokens) <= 1024 {
		tokenBufPool.Put(sm.tokens[:0])
	}

	// Return state machine to pool
	sm.tokens = nil // Clear reference
	putStateMachine(sm)

	return result
}

// newStateMachine creates a new state machine for tokenizing the given text.
// This function is primarily used for testing. In production, use Tokenize()
// which uses pooled state machines for better performance.
func newStateMachine(text string) *stateMachine {
	return &stateMachine{
		input:    []rune(text),
		position: 0,
		tokens:   make([]string, 0),
	}
}

// tokenizeWithStateMachine processes the input according to the JS regex pattern
func (sm *stateMachine) tokenizeWithStateMachine() []string {
	for sm.position < len(sm.input) {
		sm.matchNext()
	}
	return sm.tokens
}

// matchNext tries to match the next token according to the pattern
func (sm *stateMachine) matchNext() {
	if sm.position >= len(sm.input) {
		return
	}

	// Try patterns in order (as regex alternation works)

	// 1. Try contractions: (?i:'s|'t|'re|'ve|'m|'ll|'d)
	if token := sm.tryContraction(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}

	// 2. Try word with optional prefix: [^\r\n\p{L}\p{N}]?\p{L}+
	if token := sm.tryWordWithPrefix(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}

	// 3. Try numbers (1-3 digits): \p{N}{1,3}
	if token := sm.tryNumbers(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}

	// 4. Try punctuation with optional space: ?[^\s\p{L}\p{N}]+[\r\n]*
	if token := sm.tryPunctuationWithSpace(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}

	// 5. Try newline sequences: \s*[\r\n]+
	if token := sm.tryNewlineSequence(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}

	// 6. Try whitespace: \s+(?!\S) or \s+
	if token := sm.tryWhitespace(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}

	// Fallback: single character
	sm.tokens = append(sm.tokens, string(sm.input[sm.position]))
	sm.position++
}

// tryContraction matches contractions
func (sm *stateMachine) tryContraction() string {
	if sm.position >= len(sm.input) || sm.input[sm.position] != '\'' {
		return ""
	}

	contractions := []string{"'s", "'t", "'re", "'ve", "'m", "'ll", "'d"}
	for _, c := range contractions {
		if sm.matchesAt(c, true) {
			// Preserve the original case from input
			end := sm.position + len([]rune(c))
			token := string(sm.input[sm.position:end])
			sm.position = end
			return token
		}
	}

	return ""
}

// tryWordWithPrefix matches [^\r\n\p{L}\p{N}]?\p{L}+
func (sm *stateMachine) tryWordWithPrefix() string {
	start := sm.position

	// Optional non-letter/number prefix (but not \r or \n)
	if sm.position < len(sm.input) {
		ch := sm.input[sm.position]
		if !isLetter(ch) && !isNumber(ch) && ch != '\r' && ch != '\n' {
			sm.position++
		}
	}

	// Must have at least one letter
	if sm.position >= len(sm.input) || !isLetter(sm.input[sm.position]) {
		// Backtrack if we consumed a prefix but found no letters
		sm.position = start
		return ""
	}

	// Consume letters
	for sm.position < len(sm.input) && isLetter(sm.input[sm.position]) {
		sm.position++
	}

	return string(sm.input[start:sm.position])
}

// tryNumbers matches \p{N}{1,3}
func (sm *stateMachine) tryNumbers() string {
	if sm.position >= len(sm.input) || !isNumber(sm.input[sm.position]) {
		return ""
	}

	start := sm.position
	count := 0
	for sm.position < len(sm.input) && isNumber(sm.input[sm.position]) && count < 3 {
		sm.position++
		count++
	}

	return string(sm.input[start:sm.position])
}

// tryPunctuationWithSpace matches  ?[^\s\p{L}\p{N}]+[\r\n]*
func (sm *stateMachine) tryPunctuationWithSpace() string {
	start := sm.position

	// Optional space at beginning
	if sm.position < len(sm.input) && sm.input[sm.position] == ' ' {
		sm.position++
	}

	// Must have at least one non-space, non-letter, non-number
	if sm.position >= len(sm.input) ||
		isWhitespace(sm.input[sm.position]) ||
		isLetter(sm.input[sm.position]) ||
		isNumber(sm.input[sm.position]) {
		// Backtrack
		sm.position = start
		return ""
	}

	// Consume non-space, non-letter, non-number characters
	for sm.position < len(sm.input) {
		ch := sm.input[sm.position]
		if isWhitespace(ch) || isLetter(ch) || isNumber(ch) {
			break
		}
		sm.position++
	}

	// Consume optional trailing \r\n
	for sm.position < len(sm.input) && (sm.input[sm.position] == '\r' || sm.input[sm.position] == '\n') {
		sm.position++
	}

	if sm.position == start {
		return ""
	}

	return string(sm.input[start:sm.position])
}

// tryNewlineSequence matches \s*[\r\n]+
func (sm *stateMachine) tryNewlineSequence() string {
	start := sm.position

	// Optional leading whitespace
	for sm.position < len(sm.input) && isWhitespace(sm.input[sm.position]) {
		if sm.input[sm.position] == '\r' || sm.input[sm.position] == '\n' {
			break
		}
		sm.position++
	}

	// Must have at least one \r or \n
	foundNewline := false
	for sm.position < len(sm.input) && (sm.input[sm.position] == '\r' || sm.input[sm.position] == '\n') {
		sm.position++
		foundNewline = true
	}

	if !foundNewline {
		// Backtrack
		sm.position = start
		return ""
	}

	return string(sm.input[start:sm.position])
}

// tryWhitespace matches \s+(?!\S) or \s+
func (sm *stateMachine) tryWhitespace() string {
	if sm.position >= len(sm.input) || !isWhitespace(sm.input[sm.position]) {
		return ""
	}

	// First try \s+(?!\S) - whitespace not followed by non-whitespace
	start := sm.position
	for sm.position < len(sm.input) && isWhitespace(sm.input[sm.position]) {
		sm.position++
	}

	// Check negative lookahead (?!\S)
	// If we're followed by non-whitespace, backtrack by one
	if sm.position < len(sm.input) && !isWhitespace(sm.input[sm.position]) {
		// We're followed by non-whitespace, so we need to backtrack
		// But only if we consumed more than one whitespace character
		if sm.position > start+1 {
			sm.position--
		}
	}

	return string(sm.input[start:sm.position])
}

// matchesAt checks if a string matches at current position (case-insensitive if specified)
func (sm *stateMachine) matchesAt(s string, caseInsensitive bool) bool {
	runes := []rune(s)
	if sm.position+len(runes) > len(sm.input) {
		return false
	}

	for i, r := range runes {
		inputR := sm.input[sm.position+i]
		if caseInsensitive {
			if unicode.ToLower(inputR) != unicode.ToLower(r) {
				return false
			}
		} else {
			if inputR != r {
				return false
			}
		}
	}

	return true
}

// Character classification helpers
func isLetter(r rune) bool {
	return unicode.IsLetter(r)
}

func isNumber(r rune) bool {
	return unicode.IsDigit(r)
}

func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}

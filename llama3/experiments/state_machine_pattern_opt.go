// +build experiments

package llama3

// Pattern-specific optimizations

// Optimized contraction lookup map
var contractionMap = map[string]bool{
	"'s": true, "'S": true,
	"'t": true, "'T": true,
	"'re": true, "'Re": true, "'RE": true, "'rE": true,
	"'ve": true, "'Ve": true, "'VE": true, "'vE": true,
	"'m": true, "'M": true,
	"'ll": true, "'Ll": true, "'LL": true, "'lL": true,
	"'d": true, "'D": true,
}

// Fast contraction check using map lookup
func tryContractionFast(input []rune, pos int) (string, int) {
	if pos >= len(input) || input[pos] != '\'' {
		return "", 0
	}
	
	// Try 2-character contractions
	if pos+2 <= len(input) {
		candidate := string(input[pos:pos+2])
		if contractionMap[candidate] {
			return candidate, 2
		}
	}
	
	// Try 3-character contractions
	if pos+3 <= len(input) {
		candidate := string(input[pos:pos+3])
		if contractionMap[candidate] {
			return candidate, 3
		}
	}
	
	return "", 0
}

// Common word prefixes for fast checking
var commonPrefixes = map[rune]bool{
	'!': true, '@': true, '#': true, '$': true, '%': true,
	'^': true, '&': true, '*': true, '(': true, ')': true,
	'-': true, '+': true, '=': true, '[': true, ']': true,
	'{': true, '}': true, '|': true, '\': true, ':': true,
	';': true, '"': true, '<': true, '>': true, ',': true,
	'.': true, '?': true, '/': true, '~': true, '`': true,
}

// TokenizePatternOpt uses pattern-specific optimizations
func TokenizePatternOpt(text string) []string {
	sm := getStateMachine(text)
	defer putStateMachine(sm)
	
	// Pre-size tokens array based on heuristics
	estimatedTokens := len(text) / 5 // Average token length ~5 chars
	if estimatedTokens < 16 {
		estimatedTokens = 16
	}
	sm.tokens = make([]string, 0, estimatedTokens)
	
	// Use optimized matching
	for sm.position < len(sm.input) {
		sm.matchNextPatternOpt()
	}
	
	return sm.tokens
}

func (sm *StateMachine) matchNextPatternOpt() {
	if sm.position >= len(sm.input) {
		return
	}
	
	// Get current character for fast dispatch
	ch := sm.input[sm.position]
	
	// Fast path for apostrophe (contractions)
	if ch == '\'' {
		if token, length := tryContractionFast(sm.input, sm.position); length > 0 {
			sm.tokens = append(sm.tokens, token)
			sm.position += length
			return
		}
	}
	
	// Fast path for digits
	if ch >= '0' && ch <= '9' {
		if token := sm.tryNumbersFast(); token != "" {
			sm.tokens = append(sm.tokens, token)
			return
		}
	}
	
	// Fast path for common ASCII letters (no prefix)
	if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
		if token := sm.tryWordFast(); token != "" {
			sm.tokens = append(sm.tokens, token)
			return
		}
	}
	
	// Fast path for whitespace
	if ch == ' ' || ch == '	' || ch == '
' || ch == '' {
		if ch == '
' || ch == '' {
			// Try newline sequence first
			if token := sm.tryNewlineSequence(); token != "" {
				sm.tokens = append(sm.tokens, token)
				return
			}
		}
		// Try general whitespace
		if token := sm.tryWhitespace(); token != "" {
			sm.tokens = append(sm.tokens, token)
			return
		}
	}
	
	// Check if it's a common prefix character
	if ch < 128 && commonPrefixes[ch] {
		// Could be word with prefix or punctuation
		if sm.position+1 < len(sm.input) {
			next := sm.input[sm.position+1]
			if isLetter(next) {
				// Word with prefix
				if token := sm.tryWordWithPrefix(); token != "" {
					sm.tokens = append(sm.tokens, token)
					return
				}
			}
		}
		// Try punctuation
		if token := sm.tryPunctuationWithSpace(); token != "" {
			sm.tokens = append(sm.tokens, token)
			return
		}
	}
	
	// Fall back to standard pattern matching
	sm.matchNext()
}

// Optimized number matching with early exit
func (sm *StateMachine) tryNumbersFast() string {
	start := sm.position
	
	// We know first character is a digit
	sm.position++
	
	// Only match up to 2 more digits
	for i := 0; i < 2 && sm.position < len(sm.input); i++ {
		if ch := sm.input[sm.position]; ch >= '0' && ch <= '9' {
			sm.position++
		} else {
			break
		}
	}
	
	return string(sm.input[start:sm.position])
}

// Optimized word matching (no prefix)
func (sm *StateMachine) tryWordFast() string {
	start := sm.position
	
	// We know first character is a letter
	sm.position++
	
	// Consume remaining letters
	for sm.position < len(sm.input) {
		ch := sm.input[sm.position]
		if ch < 128 {
			// ASCII fast path
			if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') {
				sm.position++
			} else {
				break
			}
		} else if isLetter(ch) {
			sm.position++
		} else {
			break
		}
	}
	
	return string(sm.input[start:sm.position])
}

// Specialized tokenizer for code/technical text
func TokenizeCodeOpt(text string) []string {
	sm := getStateMachine(text)
	defer putStateMachine(sm)
	
	// Code typically has more tokens
	estimatedTokens := len(text) / 3
	if estimatedTokens < 32 {
		estimatedTokens = 32
	}
	sm.tokens = make([]string, 0, estimatedTokens)
	
	// Use standard matching but with code-specific heuristics
	for sm.position < len(sm.input) {
		sm.matchNext()
	}
	
	return sm.tokens
}

// Specialized tokenizer for natural language text
func TokenizeNaturalOpt(text string) []string {
	sm := getStateMachine(text)
	defer putStateMachine(sm)
	
	// Natural language has longer tokens on average
	estimatedTokens := len(text) / 6
	if estimatedTokens < 16 {
		estimatedTokens = 16
	}
	sm.tokens = make([]string, 0, estimatedTokens)
	
	// Use standard matching
	for sm.position < len(sm.input) {
		sm.matchNext()
	}
	
	return sm.tokens
}

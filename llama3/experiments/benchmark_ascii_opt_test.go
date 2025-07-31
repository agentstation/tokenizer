// +build experiments

package llama3

import (
	"testing"
	"unicode"
)

// Benchmark ASCII character classification optimizations

func BenchmarkIsLetter_Unicode_ASCII(b *testing.B) {
	// Test with ASCII characters only
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := chars[i%len(chars)]
		_ = unicode.IsLetter(r)
	}
}

func BenchmarkIsLetter_Fast_ASCII(b *testing.B) {
	// Test with ASCII characters only
	chars := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := chars[i%len(chars)]
		_ = isLetterFast(r)
	}
}

func BenchmarkIsLetter_Unicode_Mixed(b *testing.B) {
	// Test with mixed ASCII and Unicode
	chars := []rune("abcABC123!@#αβγ文字العربيةעברית")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := chars[i%len(chars)]
		_ = unicode.IsLetter(r)
	}
}

func BenchmarkIsLetter_Fast_Mixed(b *testing.B) {
	// Test with mixed ASCII and Unicode
	chars := []rune("abcABC123!@#αβγ文字العربيةעברית")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := chars[i%len(chars)]
		_ = isLetterFast(r)
	}
}

func BenchmarkIsNumber_Unicode_ASCII(b *testing.B) {
	chars := []rune("0123456789abcABC!@#")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := chars[i%len(chars)]
		_ = unicode.IsNumber(r)
	}
}

func BenchmarkIsNumber_Fast_ASCII(b *testing.B) {
	chars := []rune("0123456789abcABC!@#")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := chars[i%len(chars)]
		_ = isNumberFast(r)
	}
}

func BenchmarkIsWhitespace_Unicode_ASCII(b *testing.B) {
	chars := []rune(" 	
abc123")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := chars[i%len(chars)]
		_ = unicode.IsSpace(r)
	}
}

func BenchmarkIsWhitespace_Fast_ASCII(b *testing.B) {
	chars := []rune(" 	
abc123")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := chars[i%len(chars)]
		_ = isWhitespaceFast(r)
	}
}

// Benchmark impact on tokenization
func BenchmarkTokenize_Original_ASCIIHeavy(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. " +
		"This is a simple ASCII text with numbers 123 and punctuation! " +
		"We're testing contractions and various patterns."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

// Create an ASCII-optimized tokenize function for comparison
func TokenizeWithASCIIOpt(text string) []string {
	sm := &StateMachine{
		input:    []rune(text),
		position: 0,
		tokens:   make([]string, 0, len(text)/4),
	}
	
	// Use a custom version with ASCII-optimized character classification
	for sm.position < len(sm.input) {
		sm.matchNextASCIIOpt()
	}
	
	return sm.tokens
}

// matchNextASCIIOpt is like matchNext but uses fast character classification
func (sm *StateMachine) matchNextASCIIOpt() {
	if sm.position >= len(sm.input) {
		return
	}
	
	// Try patterns in order (as regex alternation works)
	
	// 1. Try contractions: (?i:'s|'t|'re|'ve|'m|'ll|'d)
	if token := sm.tryContractionASCIIOpt(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// 2. Try word with optional prefix: [^
\p{L}\p{N}]?\p{L}+
	if token := sm.tryWordWithPrefixASCIIOpt(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// 3. Try numbers (1-3 digits): \p{N}{1,3}
	if token := sm.tryNumbersASCIIOpt(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// 4. Try punctuation with optional space: ?[^\s\p{L}\p{N}]+[
]*
	if token := sm.tryPunctuationWithSpaceASCIIOpt(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// 5. Try newline sequences: \s*[
]+
	if token := sm.tryNewlineSequenceASCIIOpt(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// 6. Try whitespace: \s+(?!\S) or \s+
	if token := sm.tryWhitespaceASCIIOpt(); token != "" {
		sm.tokens = append(sm.tokens, token)
		return
	}
	
	// Fallback: single character
	sm.tokens = append(sm.tokens, string(sm.input[sm.position]))
	sm.position++
}

// Optimized versions of each try method using fast character classification
func (sm *StateMachine) tryContractionASCIIOpt() string {
	// Same logic as tryContraction but with no character classification
	return sm.tryContraction()
}

func (sm *StateMachine) tryWordWithPrefixASCIIOpt() string {
	start := sm.position
	
	// Optional non-letter/number prefix (but not  or 
)
	if sm.position < len(sm.input) {
		ch := sm.input[sm.position]
		if !isLetterFast(ch) && !isNumberFast(ch) && ch != '' && ch != '
' {
			sm.position++
		}
	}
	
	// Must have at least one letter
	if sm.position >= len(sm.input) || !isLetterFast(sm.input[sm.position]) {
		// Backtrack if we consumed a prefix but found no letters
		sm.position = start
		return ""
	}
	
	// Consume letters
	for sm.position < len(sm.input) && isLetterFast(sm.input[sm.position]) {
		sm.position++
	}
	
	return string(sm.input[start:sm.position])
}

func (sm *StateMachine) tryNumbersASCIIOpt() string {
	if sm.position >= len(sm.input) || !isNumberFast(sm.input[sm.position]) {
		return ""
	}
	
	start := sm.position
	count := 0
	for sm.position < len(sm.input) && isNumberFast(sm.input[sm.position]) && count < 3 {
		sm.position++
		count++
	}
	
	return string(sm.input[start:sm.position])
}

func (sm *StateMachine) tryPunctuationWithSpaceASCIIOpt() string {
	start := sm.position
	
	// Optional leading space
	if sm.position < len(sm.input) && sm.input[sm.position] == ' ' {
		sm.position++
	}
	
	// Match non-space, non-letter, non-number characters
	hasContent := false
	for sm.position < len(sm.input) {
		ch := sm.input[sm.position]
		if isWhitespaceFast(ch) || isLetterFast(ch) || isNumberFast(ch) {
			break
		}
		sm.position++
		hasContent = true
	}
	
	// Match trailing newlines
	for sm.position < len(sm.input) && (sm.input[sm.position] == '' || sm.input[sm.position] == '
') {
		sm.position++
		hasContent = true
	}
	
	if hasContent {
		return string(sm.input[start:sm.position])
	}
	
	// Backtrack if we only consumed the optional space
	sm.position = start
	return ""
}

func (sm *StateMachine) tryNewlineSequenceASCIIOpt() string {
	start := sm.position
	
	// Optional leading whitespace
	for sm.position < len(sm.input) && isWhitespaceFast(sm.input[sm.position]) {
		sm.position++
	}
	
	// Must have at least one newline
	newlineStart := sm.position
	for sm.position < len(sm.input) && (sm.input[sm.position] == '' || sm.input[sm.position] == '
') {
		sm.position++
	}
	
	if sm.position > newlineStart {
		return string(sm.input[start:sm.position])
	}
	
	// Backtrack
	sm.position = start
	return ""
}

func (sm *StateMachine) tryWhitespaceASCIIOpt() string {
	// First try \s+(?!\S) - whitespace not followed by non-whitespace
	start := sm.position
	for sm.position < len(sm.input) && isWhitespaceFast(sm.input[sm.position]) {
		sm.position++
	}
	
	// Check negative lookahead (?!\S)
	// If we're followed by non-whitespace, backtrack by one
	if sm.position < len(sm.input) && !isWhitespaceFast(sm.input[sm.position]) {
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
	for sm.position < len(sm.input) && isWhitespaceFast(sm.input[sm.position]) {
		sm.position++
	}
	
	return string(sm.input[start:sm.position])
}

func BenchmarkTokenize_ASCIIOpt_ASCIIHeavy(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. " +
		"This is a simple ASCII text with numbers 123 and punctuation! " +
		"We're testing contractions and various patterns."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeWithASCIIOpt(text)
	}
}

// +build experiments

package llama3

import (
	"unicode"
)

// ASCII-optimized character classification functions
// These maintain exact compatibility while adding fast paths

func isLetterFast(r rune) bool {
	// ASCII fast path
	if r < 128 {
		return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
	}
	return unicode.IsLetter(r)
}

func isNumberFast(r rune) bool {
	// ASCII fast path
	if r < 128 {
		return r >= '0' && r <= '9'
	}
	return unicode.IsNumber(r)
}

func isWhitespaceFast(r rune) bool {
	// ASCII fast path
	switch r {
	case ' ', '	', '
', '', '', '':
		return true
	}
	if r < 128 {
		return false
	}
	return unicode.IsSpace(r)
}

// Enable ASCII optimizations by replacing the character classification functions
func EnableASCIIOptimizations() {
	// This would require making isLetter, isNumber, isWhitespace into variables
	// For now, we'll create optimized versions of the state machine methods
}

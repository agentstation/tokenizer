package llama3

import (
	"regexp"
)

// regexSplitSpecial is a specialized version for special token splitting
// that ensures exact preservation of matches.
func regexSplitSpecial(text string, regex *regexp.Regexp) []string {
	if text == "" {
		return []string{}
	}

	var result []string
	lastEnd := 0

	for _, match := range regex.FindAllStringSubmatchIndex(text, -1) {
		if len(match) >= 2 {
			start, end := match[0], match[1]

			// Add text before match
			if start > lastEnd {
				before := text[lastEnd:start]
				if before != "" {
					result = append(result, before)
				}
			}

			// Add the match itself
			if end > start {
				matchText := text[start:end]
				if matchText != "" {
					result = append(result, matchText)
				}
			}

			lastEnd = end
		}
	}

	// Add any remaining text
	if lastEnd < len(text) {
		remaining := text[lastEnd:]
		if remaining != "" {
			result = append(result, remaining)
		}
	}

	return result
}

// pretokenize performs the pre-tokenization step using state machine
// and byte-level encoding.
func (t *Tokenizer) pretokenize(text string) []string {
	// Use pooled state machine for better performance
	parts := Tokenize(text)

	// Apply byte-level encoding to each part
	encoded := make([]string, len(parts))
	for i, part := range parts {
		encoded[i] = encodeBytes([]byte(part))
	}

	return encoded
}

package llama3

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	beginOfTextToken = "<|begin_of_text|>"
	endOfTextToken   = "<|end_of_text|>"
)

var (
	// specialTokenRegex matches Llama 3 special tokens
	specialTokenRegex = regexp.MustCompile(`<\|(?:begin_of_text|end_of_text|start_header_id|end_header_id|eot_id|eom_id|python_tag|finetune_right_pad_id|reserved_special_token_(?:[0-9]|[1-9][0-9]|1[0-9][0-9]|2[0-3][0-9]|24[0-7]))\|>`)
	
	// optimisticSpecialTokenRegex matches any pattern that looks like a special token
	optimisticSpecialTokenRegex = regexp.MustCompile(`<\|[a-zA-Z0-9_]+\|>`)
)

// getDefaultSpecialTokens returns all Llama 3 special tokens in order.
func getDefaultSpecialTokens() []string {
	tokens := []string{
		"<|begin_of_text|>",
		"<|end_of_text|>",
		"<|reserved_special_token_0|>",
		"<|reserved_special_token_1|>",
		"<|finetune_right_pad_id|>",
		"<|reserved_special_token_2|>",
		"<|start_header_id|>",
		"<|end_header_id|>",
		"<|eom_id|>",
		"<|eot_id|>",
		"<|python_tag|>",
	}
	
	// Add reserved special tokens 3-247
	for i := 3; i <= 247; i++ {
		tokens = append(tokens, fmt.Sprintf("<|reserved_special_token_%d|>", i))
	}
	
	return tokens
}

// isSpecialToken checks if a string is in the special token format.
func isSpecialToken(token string) bool {
	return strings.HasPrefix(token, "<|") && strings.HasSuffix(token, "|>")
}

// splitBySpecialTokens splits text by special tokens while preserving the tokens.
func splitBySpecialTokens(text string, regex *regexp.Regexp) []string {
	if text == "" {
		return []string{}
	}
	
	matches := regex.FindAllStringIndex(text, -1)
	if len(matches) == 0 {
		return []string{text}
	}
	
	result := make([]string, 0, len(matches)*2+1)
	lastEnd := 0
	
	for _, match := range matches {
		start, end := match[0], match[1]
		
		// Add text before the match
		if start > lastEnd {
			result = append(result, text[lastEnd:start])
		}
		
		// Add the matched special token
		if end > start {
			result = append(result, text[start:end])
		}
		
		lastEnd = end
	}
	
	// Add remaining text
	if lastEnd < len(text) {
		result = append(result, text[lastEnd:])
	}
	
	return result
}
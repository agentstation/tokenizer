package testing

import (
	"fmt"
	"strings"
	"unicode"
)

// TestCase represents a test case with metadata
type TestCase struct {
	Input       string
	Description string
	Category    string // "edge", "unicode", "whitespace", "contraction", etc.
}

// GenerateTestCases creates comprehensive test cases
func GenerateTestCases() []TestCase {
	var cases []TestCase

	// Edge cases
	cases = append(cases, []TestCase{
		{"", "Empty string", "edge"},
		{" ", "Single space", "edge"},
		{"\t", "Single tab", "edge"},
		{"\n", "Single newline", "edge"},
		{"\r\n", "CRLF", "edge"},
		{"'", "Single apostrophe", "edge"},
		{"''", "Double apostrophe", "edge"},
		{"123456", "More than 3 digits", "edge"},
	}...)

	// Whitespace patterns (critical for state machine)
	for i := 1; i <= 20; i++ {
		spaces := strings.Repeat(" ", i)
		cases = append(cases, TestCase{
			Input:       spaces + "word",
			Description: fmt.Sprintf("%d spaces before word", i),
			Category:    "whitespace",
		})
		cases = append(cases, TestCase{
			Input:       "word" + spaces,
			Description: fmt.Sprintf("word before %d spaces", i),
			Category:    "whitespace",
		})
	}

	// Tab patterns
	for i := 1; i <= 10; i++ {
		tabs := strings.Repeat("\t", i)
		cases = append(cases, TestCase{
			Input:       tabs + "word",
			Description: fmt.Sprintf("%d tabs before word", i),
			Category:    "whitespace",
		})
	}

	// Mixed whitespace
	cases = append(cases, []TestCase{
		{" \t \t word", "Mixed space and tab before word", "whitespace"},
		{"\n\r\n\r\nword", "Mixed newlines before word", "whitespace"},
		{"   \t\t\t   \n\n\n   word", "Complex whitespace mix", "whitespace"},
		{"word   \t\t\t   \n\n\n", "Word before complex whitespace", "whitespace"},
	}...)

	// Contractions - all cases and variations
	contractions := []string{"'s", "'t", "'re", "'ve", "'m", "'ll", "'d"}
	words := []string{"can", "won", "they", "I", "we", "he", "it"}

	for _, word := range words {
		for _, contraction := range contractions {
			// Lowercase
			cases = append(cases, TestCase{
				Input:       word + contraction,
				Description: fmt.Sprintf("Contraction: %s%s", word, contraction),
				Category:    "contraction",
			})
			// Uppercase word
			cases = append(cases, TestCase{
				Input:       strings.ToUpper(word) + contraction,
				Description: fmt.Sprintf("Uppercase word: %s%s", strings.ToUpper(word), contraction),
				Category:    "contraction",
			})
			// Uppercase contraction
			cases = append(cases, TestCase{
				Input:       word + strings.ToUpper(contraction),
				Description: fmt.Sprintf("Uppercase contraction: %s%s", word, strings.ToUpper(contraction)),
				Category:    "contraction",
			})
			// Title case
			cases = append(cases, TestCase{
				Input:       strings.Title(word) + contraction,
				Description: fmt.Sprintf("Title case: %s%s", strings.Title(word), contraction),
				Category:    "contraction",
			})
		}
	}

	// Numbers and digit patterns
	cases = append(cases, []TestCase{
		{"0", "Single zero", "number"},
		{"00", "Double zero", "number"},
		{"000", "Triple zero", "number"},
		{"0000", "Four zeros (exceeds limit)", "number"},
		{"123", "Three digits", "number"},
		{"1234", "Four digits", "number"},
		{"12345", "Five digits", "number"},
		{"1 2 3", "Spaced single digits", "number"},
		{"12 34 56", "Spaced double digits", "number"},
		{"123 456 789", "Spaced triple digits", "number"},
	}...)

	// Punctuation patterns
	punctuation := "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	for _, p := range punctuation {
		cases = append(cases, TestCase{
			Input:       string(p),
			Description: fmt.Sprintf("Single punctuation: %c", p),
			Category:    "punctuation",
		})
		cases = append(cases, TestCase{
			Input:       " " + string(p),
			Description: fmt.Sprintf("Space before punctuation: %c", p),
			Category:    "punctuation",
		})
		cases = append(cases, TestCase{
			Input:       string(p) + " ",
			Description: fmt.Sprintf("Punctuation before space: %c", p),
			Category:    "punctuation",
		})
	}

	// Unicode categories
	unicodeTests := []struct {
		text string
		desc string
	}{
		// Latin extended
		{"café", "Latin with accent"},
		{"naïve", "Latin with diaeresis"},
		{"résumé", "Multiple accents"},
		{"Zürich", "German umlaut"},
		{"español", "Spanish ñ"},
		{"português", "Portuguese"},

		// Greek
		{"αβγδε", "Greek lowercase"},
		{"ΑΒΓΔΕ", "Greek uppercase"},
		{"Ελληνικά", "Greek word"},

		// Cyrillic
		{"привет", "Russian hello"},
		{"ПРИВЕТ", "Russian hello uppercase"},
		{"Москва", "Moscow"},

		// CJK
		{"你好", "Chinese hello"},
		{"世界", "Chinese world"},
		{"こんにちは", "Japanese hello"},
		{"안녕하세요", "Korean hello"},

		// Arabic
		{"مرحبا", "Arabic hello"},
		{"السلام", "Arabic peace"},

		// Hebrew
		{"שלום", "Hebrew hello"},
		{"עברית", "Hebrew language"},

		// Emojis
		{"🦙", "Llama emoji"},
		{"👍", "Thumbs up"},
		{"🌍🌎🌏", "Globe emojis"},
		{"👨‍👩‍👧‍👦", "Family emoji"},
		{"🏳️‍🌈", "Rainbow flag"},
		{"🇺🇸🇬🇧🇯🇵", "Country flags"},
	}

	for _, ut := range unicodeTests {
		cases = append(cases, TestCase{
			Input:       ut.text,
			Description: ut.desc,
			Category:    "unicode",
		})
		// Also test with surrounding ASCII
		cases = append(cases, TestCase{
			Input:       "Hello " + ut.text + " world",
			Description: ut.desc + " with ASCII",
			Category:    "unicode",
		})
	}

	// Word prefix patterns
	prefixes := []rune{'!', '@', '#', '$', '%', '^', '&', '*', '(', ')', '-', '+', '=', '[', ']', '{', '}', '|', '\\', ':', ';', '"', '<', '>', ',', '.', '?', '/'}
	for _, prefix := range prefixes {
		cases = append(cases, TestCase{
			Input:       string(prefix) + "word",
			Description: fmt.Sprintf("Word with prefix: %c", prefix),
			Category:    "prefix",
		})
	}

	// Real-world patterns
	realWorld := []TestCase{
		{"user@example.com", "Email address", "real"},
		{"first.last@company.co.uk", "Complex email", "real"},
		{"https://example.com", "URL", "real"},
		{"http://sub.example.com:8080/path?q=1&v=2#anchor", "Complex URL", "real"},
		{"192.168.1.1", "IP address", "real"},
		{"2001:0db8:85a3:0000:0000:8a2e:0370:7334", "IPv6 address", "real"},
		{"+1-555-123-4567", "Phone number", "real"},
		{"$1,234.56", "Currency", "real"},
		{"€999.99", "Euro currency", "real"},
		{"¥10,000", "Yen currency", "real"},
		{"50%", "Percentage", "real"},
		{"3.14159", "Decimal number", "real"},
		{"1.23e-4", "Scientific notation", "real"},
		{"#hashtag", "Hashtag", "real"},
		{"@mention", "Mention", "real"},
		{"C:\\Windows\\System32", "Windows path", "real"},
		{"/usr/local/bin", "Unix path", "real"},
		{"~/Documents/file.txt", "Home path", "real"},
	}
	cases = append(cases, realWorld...)

	// Code patterns
	codeSnippets := []TestCase{
		{"if (x > 0) { return true; }", "C-style if statement", "code"},
		{"func main() { }", "Go function", "code"},
		{"def hello():", "Python function", "code"},
		{"public class Main { }", "Java class", "code"},
		{"const x = () => { };", "JavaScript arrow function", "code"},
		{"SELECT * FROM users WHERE id = 1;", "SQL query", "code"},
		{"git commit -m 'Initial commit'", "Git command", "code"},
		{"npm install --save-dev @types/node", "NPM command", "code"},
		{"docker run -p 8080:80 nginx", "Docker command", "code"},
	}
	cases = append(cases, codeSnippets...)

	// Boundary cases
	boundary := []TestCase{
		{strings.Repeat("a", 1000), "1000 letter word", "boundary"},
		{strings.Repeat("1", 1000), "1000 digits", "boundary"},
		{strings.Repeat(" ", 1000), "1000 spaces", "boundary"},
		{strings.Repeat("word ", 200), "200 words", "boundary"},
		{strings.Repeat("'s", 100), "100 contractions", "boundary"},
	}
	cases = append(cases, boundary...)

	return cases
}

// GenerateTestVectorString creates a string representation for comparison
func GenerateTestVectorString(input string, tokens []int) string {
	return fmt.Sprintf(`{"input":%q,"expected":%v}`, input, tokens)
}

// ValidateTokenization checks if tokenization is valid
func ValidateTokenization(input string, tokens []string) error {
	// Reconstruct the input from tokens
	reconstructed := strings.Join(tokens, "")

	// The reconstructed text should match the original when decoded
	// (accounting for byte-level encoding)
	if len(reconstructed) == 0 && len(input) > 0 {
		return fmt.Errorf("empty tokenization for non-empty input")
	}

	// Check for invalid tokens
	for i, token := range tokens {
		if len(token) == 0 {
			return fmt.Errorf("empty token at position %d", i)
		}

		// Validate token contains valid runes
		for _, r := range token {
			if r == unicode.ReplacementChar && !strings.Contains(input, string(r)) {
				return fmt.Errorf("invalid unicode in token %d: %q", i, token)
			}
		}
	}

	return nil
}

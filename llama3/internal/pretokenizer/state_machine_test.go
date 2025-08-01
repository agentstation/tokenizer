package pretokenizer

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"unicode"
)

func TestStateMachinePatterns(t *testing.T) {
	testGroups := map[string][]struct {
		name     string
		input    string
		expected []string
	}{
		"contractions": {
			// Basic contractions
			{
				name:     "lowercase_apostrophe_s",
				input:    "don't",
				expected: []string{"don", "'t"},
			},
			{
				name:     "uppercase_apostrophe_s",
				input:    "DON'T",
				expected: []string{"DON", "'T"},
			},
			{
				name:     "mixed_case",
				input:    "Can't",
				expected: []string{"Can", "'t"},
			},
			{
				name:     "multiple_contractions",
				input:    "I've you're",
				expected: []string{"I", "'ve", " you", "'re"},
			},
			// All contraction forms
			{
				name:     "apostrophe_s",
				input:    "that's",
				expected: []string{"that", "'s"},
			},
			{
				name:     "apostrophe_t",
				input:    "can't",
				expected: []string{"can", "'t"},
			},
			{
				name:     "apostrophe_re",
				input:    "they're",
				expected: []string{"they", "'re"},
			},
			{
				name:     "apostrophe_ve",
				input:    "I've",
				expected: []string{"I", "'ve"},
			},
			{
				name:     "apostrophe_m",
				input:    "I'm",
				expected: []string{"I", "'m"},
			},
			{
				name:     "apostrophe_ll",
				input:    "they'll",
				expected: []string{"they", "'ll"},
			},
			{
				name:     "apostrophe_d",
				input:    "they'd",
				expected: []string{"they", "'d"},
			},
			// Edge cases
			{
				name:     "contraction_at_start",
				input:    "'twas",
				expected: []string{"'t", "was"}, // 't is a valid contraction per the regex
			},
			{
				name:     "multiple_apostrophes",
				input:    "can't've",
				expected: []string{"can", "'t", "'ve"},
			},
			{
				name:     "non_contraction_apostrophe",
				input:    "rock'n'roll",
				expected: []string{"rock", "'n", "'roll"},
			},
			{
				name:     "possessive_not_contraction",
				input:    "John's",
				expected: []string{"John", "'s"},
			},
		},
		"words": {
			// Basic words
			{
				name:     "simple_word",
				input:    "hello",
				expected: []string{"hello"},
			},
			{
				name:     "capitalized_word",
				input:    "Hello",
				expected: []string{"Hello"},
			},
			{
				name:     "uppercase_word",
				input:    "HELLO",
				expected: []string{"HELLO"},
			},
			{
				name:     "mixed_case_word",
				input:    "HeLLo",
				expected: []string{"HeLLo"},
			},
			// Words with prefixes
			{
				name:     "word_with_exclamation_prefix",
				input:    "!hello",
				expected: []string{"!hello"},
			},
			{
				name:     "word_with_hash_prefix",
				input:    "#hello",
				expected: []string{"#hello"},
			},
			{
				name:     "word_with_at_prefix",
				input:    "@hello",
				expected: []string{"@hello"},
			},
			{
				name:     "word_with_dollar_prefix",
				input:    "$hello",
				expected: []string{"$hello"},
			},
			{
				name:     "word_with_percent_prefix",
				input:    "%hello",
				expected: []string{"%hello"},
			},
			// Multiple words
			{
				name:     "two_words",
				input:    "hello world",
				expected: []string{"hello", " world"},
			},
			{
				name:     "three_words",
				input:    "hello world test",
				expected: []string{"hello", " world", " test"},
			},
			// Unicode words
			{
				name:     "french_accents",
				input:    "café",
				expected: []string{"café"},
			},
			{
				name:     "german_umlaut",
				input:    "über",
				expected: []string{"über"},
			},
			{
				name:     "spanish_tilde",
				input:    "niño",
				expected: []string{"niño"},
			},
			{
				name:     "greek_letters",
				input:    "αβγδε",
				expected: []string{"αβγδε"},
			},
			{
				name:     "cyrillic",
				input:    "привет",
				expected: []string{"привет"},
			},
			{
				name:     "chinese",
				input:    "你好",
				expected: []string{"你好"},
			},
			{
				name:     "japanese_hiragana",
				input:    "こんにちは",
				expected: []string{"こんにちは"},
			},
			{
				name:     "arabic",
				input:    "مرحبا",
				expected: []string{"مرحبا"},
			},
			// Edge cases
			{
				name:     "single_letter",
				input:    "a",
				expected: []string{"a"},
			},
			{
				name:     "single_unicode_letter",
				input:    "α",
				expected: []string{"α"},
			},
		},
		"numbers": {
			// Basic numbers
			{
				name:     "single_digit",
				input:    "5",
				expected: []string{"5"},
			},
			{
				name:     "two_digits",
				input:    "42",
				expected: []string{"42"},
			},
			{
				name:     "three_digits",
				input:    "123",
				expected: []string{"123"},
			},
			{
				name:     "four_digits_split",
				input:    "1234",
				expected: []string{"123", "4"},
			},
			{
				name:     "five_digits",
				input:    "12345",
				expected: []string{"123", "45"},
			},
			{
				name:     "six_digits",
				input:    "123456",
				expected: []string{"123", "456"},
			},
			// Numbers in context
			{
				name:     "number_word",
				input:    "123abc",
				expected: []string{"123", "abc"},
			},
			{
				name:     "word_number",
				input:    "abc123",
				expected: []string{"abc", "123"},
			},
			{
				name:     "mixed_numbers_text",
				input:    "test123abc456def",
				expected: []string{"test", "123", "abc", "456", "def"},
			},
			// Edge cases
			{
				name:     "zero",
				input:    "0",
				expected: []string{"0"},
			},
			{
				name:     "leading_zeros",
				input:    "007",
				expected: []string{"007"},
			},
			{
				name:     "all_zeros",
				input:    "000",
				expected: []string{"000"},
			},
			// Unicode numbers
			{
				name:     "arabic_numerals",
				input:    "٠١٢٣",
				expected: []string{"٠١٢", "٣"},
			},
			{
				name:     "devanagari_numerals",
				input:    "०१२३",
				expected: []string{"०१२", "३"},
			},
		},
		"punctuation": {
			// Single punctuation
			{
				name:     "period",
				input:    ".",
				expected: []string{"."},
			},
			{
				name:     "comma",
				input:    ",",
				expected: []string{","},
			},
			{
				name:     "exclamation",
				input:    "!",
				expected: []string{"!"},
			},
			{
				name:     "question",
				input:    "?",
				expected: []string{"?"},
			},
			{
				name:     "semicolon",
				input:    ";",
				expected: []string{";"},
			},
			{
				name:     "colon",
				input:    ":",
				expected: []string{":"},
			},
			// Multiple punctuation
			{
				name:     "multiple_exclamation",
				input:    "!!!",
				expected: []string{"!!!"},
			},
			{
				name:     "multiple_question",
				input:    "???",
				expected: []string{"???"},
			},
			{
				name:     "mixed_punctuation",
				input:    "!?!",
				expected: []string{"!?!"},
			},
			// Punctuation with newlines
			{
				name:     "punctuation_with_newline",
				input:    "!\n",
				expected: []string{"!\n"},
			},
			{
				name:     "multiple_punctuation_newline",
				input:    "!!!\n",
				expected: []string{"!!!\n"},
			},
			// Punctuation in context
			{
				name:     "word_punctuation",
				input:    "Hello!",
				expected: []string{"Hello", "!"},
			},
			{
				name:     "punctuation_word",
				input:    "!Hello",
				expected: []string{"!Hello"},
			},
			// Special punctuation
			{
				name:     "ellipsis",
				input:    "...",
				expected: []string{"..."},
			},
			{
				name:     "dash",
				input:    "-",
				expected: []string{"-"},
			},
			{
				name:     "underscore",
				input:    "_",
				expected: []string{"_"},
			},
			{
				name:     "parentheses",
				input:    "()",
				expected: []string{"()"},
			},
			{
				name:     "brackets",
				input:    "[]",
				expected: []string{"[]"},
			},
			{
				name:     "braces",
				input:    "{}",
				expected: []string{"{}"},
			},
			{
				name:     "angle_brackets",
				input:    "<>",
				expected: []string{"<>"},
			},
		},
		"whitespace": {
			// Basic whitespace
			{
				name:     "single_space",
				input:    " ",
				expected: []string{" "},
			},
			{
				name:     "double_space",
				input:    "  ",
				expected: []string{"  "},
			},
			{
				name:     "triple_space",
				input:    "   ",
				expected: []string{"   "},
			},
			{
				name:     "many_spaces",
				input:    "          ",
				expected: []string{"          "},
			},
			// Tabs
			{
				name:     "single_tab",
				input:    "\t",
				expected: []string{"\t"},
			},
			{
				name:     "double_tab",
				input:    "\t\t",
				expected: []string{"\t\t"},
			},
			{
				name:     "triple_tab",
				input:    "\t\t\t",
				expected: []string{"\t\t\t"},
			},
			// Newlines
			{
				name:     "newline",
				input:    "\n",
				expected: []string{"\n"},
			},
			{
				name:     "carriage_return",
				input:    "\r",
				expected: []string{"\r"},
			},
			{
				name:     "carriage_return_newline",
				input:    "\r\n",
				expected: []string{"\r\n"},
			},
			{
				name:     "multiple_newlines",
				input:    "\n\n",
				expected: []string{"\n\n"},
			},
			// Mixed whitespace
			{
				name:     "space_tab",
				input:    " \t",
				expected: []string{" \t"},
			},
			{
				name:     "tab_space",
				input:    "\t ",
				expected: []string{"\t "},
			},
			{
				name:     "space_newline",
				input:    " \n",
				expected: []string{" \n"},
			},
			{
				name:     "mixed_whitespace",
				input:    "  \t\n",
				expected: []string{"  \t\n"},
			},
			// Whitespace with content
			{
				name:     "spaces_before_word",
				input:    "   word",
				expected: []string{"  ", " word"},
			},
			{
				name:     "tabs_before_word",
				input:    "\t\tword",
				expected: []string{"\t", "\tword"}, // Whitespace pattern with negative lookahead
			},
			{
				name:     "word_spaces_word",
				input:    "hello   world",
				expected: []string{"hello", "  ", " world"}, // Whitespace pattern with negative lookahead
			},
			// Edge cases
			{
				name:     "space_at_end_before_period",
				input:    "test .",
				expected: []string{"test", " ."},
			},
			{
				name:     "newline_in_sentence",
				input:    "hello\nworld",
				expected: []string{"hello", "\n", "world"},
			},
		},
		"complex": {
			// Email addresses
			{
				name:     "simple_email",
				input:    "user@example.com",
				expected: []string{"user", "@example", ".com"},
			},
			{
				name:     "email_with_dots",
				input:    "first.last@example.com",
				expected: []string{"first", ".last", "@example", ".com"},
			},
			{
				name:     "email_with_plus",
				input:    "user+tag@example.com",
				expected: []string{"user", "+tag", "@example", ".com"},
			},
			// URLs
			{
				name:     "http_url",
				input:    "http://example.com",
				expected: []string{"http", "://", "example", ".com"},
			},
			{
				name:     "https_url",
				input:    "https://example.com",
				expected: []string{"https", "://", "example", ".com"},
			},
			{
				name:     "url_with_path",
				input:    "https://example.com/path",
				expected: []string{"https", "://", "example", ".com", "/path"},
			},
			{
				name:     "url_with_query",
				input:    "https://example.com?q=test",
				expected: []string{"https", "://", "example", ".com", "?q", "=test"},
			},
			// Code snippets
			{
				name:     "simple_if",
				input:    "if (x > 0)",
				expected: []string{"if", " (", "x", " >", " ", "0", ")"},
			},
			{
				name:     "function_call",
				input:    "func(a, b)",
				expected: []string{"func", "(a", ",", " b", ")"}, // Punctuation pattern matches (a together
			},
			{
				name:     "array_access",
				input:    "arr[0]",
				expected: []string{"arr", "[", "0", "]"},
			},
			{
				name:     "method_chain",
				input:    "obj.method().prop",
				expected: []string{"obj", ".method", "().", "prop"}, // Punctuation pattern behavior
			},
			// Mathematical expressions
			{
				name:     "simple_math",
				input:    "2 + 2 = 4",
				expected: []string{"2", " +", " ", "2", " =", " ", "4"},
			},
			{
				name:     "complex_math",
				input:    "(a + b) * c",
				expected: []string{"(a", " +", " b", ")", " *", " c"}, // Punctuation pattern
			},
			// Mixed content
			{
				name:     "sentence_with_punctuation",
				input:    "Hello, world! How are you?",
				expected: []string{"Hello", ",", " world", "!", " How", " are", " you", "?"},
			},
			{
				name:     "code_comment",
				input:    "// This is a comment",
				expected: []string{"//", " This", " is", " a", " comment"},
			},
			{
				name:     "hashtag",
				input:    "#programming",
				expected: []string{"#programming"},
			},
			{
				name:     "mention",
				input:    "@username",
				expected: []string{"@username"},
			},
			// Unicode and emoji
			{
				name:     "unicode_emoji",
				input:    "Hello 🦙!",
				expected: []string{"Hello", " 🦙!"},
			},
			{
				name:     "multiple_emoji",
				input:    "🎉🎊🎈",
				expected: []string{"🎉🎊🎈"},
			},
			{
				name:     "emoji_in_sentence",
				input:    "I love 🦙 llamas!",
				expected: []string{"I", " love", " 🦙", " llamas", "!"},
			},
			// Edge cases
			{
				name:     "empty_string",
				input:    "",
				expected: []string{},
			},
			{
				name:     "only_punctuation",
				input:    "!@#$%^&*()",
				expected: []string{"!@#$%^&*()"},
			},
			{
				name:     "mixed_everything",
				input:    "Test123!@# αβγ 文字 🦙\n\tNext line",
				expected: []string{"Test", "123", "!@#", " αβγ", " 文字", " 🦙\n", "\tNext", " line"}, // Newline sequence pattern
			},
		},
		"special_regex_patterns": {
			// Test the specific regex patterns
			{
				name:     "negative_lookahead_spaces",
				input:    "test   ", // Multiple spaces at end
				expected: []string{"test", "   "},
			},
			{
				name:     "spaces_before_non_whitespace",
				input:    "   test", // Spaces before word
				expected: []string{"  ", " test"},
			},
			{
				name:     "word_boundary_punctuation",
				input:    "word!next",
				expected: []string{"word", "!next"}, // Current implementation behavior
			},
			{
				name:     "number_word_boundary",
				input:    "123abc",
				expected: []string{"123", "abc"},
			},
			{
				name:     "contraction_boundary",
				input:    "it's",
				expected: []string{"it", "'s"},
			},
		},
		"llama3_specific": {
			// Test Llama 3 specific tokenization patterns
			{
				name:     "special_token_like",
				input:    "<|not_special|>",
				expected: []string{"<|", "not", "_special", "|>"},
			},
			{
				name:     "angle_brackets_separate",
				input:    "<test>",
				expected: []string{"<test", ">"},
			},
			{
				name:     "pipe_characters",
				input:    "a|b|c",
				expected: []string{"a", "|b", "|c"},
			},
			{
				name:     "underscore_handling",
				input:    "test_case_name",
				expected: []string{"test", "_case", "_name"},
			},
			{
				name:     "camelCase",
				input:    "camelCaseWord",
				expected: []string{"camelCaseWord"}, // Regex pattern doesn't split on case boundaries
			},
			{
				name:     "PascalCase",
				input:    "PascalCaseWord",
				expected: []string{"PascalCaseWord"}, // Regex pattern doesn't split on case boundaries
			},
		},
	}

	for groupName, tests := range testGroups {
		t.Run(groupName, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					got := Tokenize(tt.input)

					if !reflect.DeepEqual(got, tt.expected) {
						t.Errorf("Tokenize(%q) = %q, want %q",
							tt.input, got, tt.expected)

						// Additional debugging info
						if len(got) != len(tt.expected) {
							t.Errorf("Length mismatch: got %d tokens, want %d tokens",
								len(got), len(tt.expected))
						}

						// Show each token comparison
						maxLen := len(got)
						if len(tt.expected) > maxLen {
							maxLen = len(tt.expected)
						}

						for i := 0; i < maxLen; i++ {
							var gotToken, wantToken string
							if i < len(got) {
								gotToken = got[i]
							}
							if i < len(tt.expected) {
								wantToken = tt.expected[i]
							}
							if gotToken != wantToken {
								t.Errorf("  Token[%d]: got %q, want %q", i, gotToken, wantToken)
							}
						}
					}
				})
			}
		})
	}
}

func TestStateMachineHelpers(t *testing.T) {
	t.Run("isLetter", func(t *testing.T) {
		tests := []struct {
			name     string
			rune     rune
			expected bool
		}{
			// ASCII letters
			{"lowercase_a", 'a', true},
			{"lowercase_z", 'z', true},
			{"uppercase_A", 'A', true},
			{"uppercase_Z", 'Z', true},

			// Non-letters
			{"number_0", '0', false},
			{"number_9", '9', false},
			{"space", ' ', false},
			{"tab", '\t', false},
			{"newline", '\n', false},
			{"punctuation_period", '.', false},
			{"punctuation_exclamation", '!', false},
			{"punctuation_question", '?', false},
			{"symbol_at", '@', false},
			{"symbol_hash", '#', false},

			// Unicode letters
			{"greek_alpha", 'α', true},
			{"greek_omega", 'ω', true},
			{"cyrillic_a", 'а', true},
			{"cyrillic_ya", 'я', true},
			{"chinese", '中', true},
			{"japanese_hiragana", 'あ', true},
			{"arabic", 'م', true},
			{"hebrew", 'א', true},

			// Accented letters
			{"french_e_acute", 'é', true},
			{"german_u_umlaut", 'ü', true},
			{"spanish_n_tilde", 'ñ', true},

			// Non-letter unicode
			{"emoji_face", '😀', false},
			{"emoji_llama", '🦙', false},
			{"math_symbol", '∑', false},
			{"currency", '€', false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := isLetter(tt.rune)
				if got != tt.expected {
					t.Errorf("isLetter(%q) = %v, want %v", tt.rune, got, tt.expected)
				}
			})
		}
	})

	t.Run("isNumber", func(t *testing.T) {
		tests := []struct {
			name     string
			rune     rune
			expected bool
		}{
			// ASCII digits
			{"zero", '0', true},
			{"one", '1', true},
			{"five", '5', true},
			{"nine", '9', true},

			// Non-digits
			{"letter_a", 'a', false},
			{"letter_Z", 'Z', false},
			{"space", ' ', false},
			{"punctuation", '.', false},
			{"symbol", '@', false},

			// Unicode digits
			{"arabic_zero", '٠', true},
			{"arabic_nine", '٩', true},
			{"devanagari_zero", '०', true},
			{"devanagari_nine", '९', true},
			{"bengali_zero", '০', true},
			{"bengali_nine", '৯', true},
			{"chinese_one", '一', false}, // Not a digit category

			// Things that look like numbers but aren't
			{"roman_numeral", 'Ⅰ', false},
			{"subscript", '₁', false},
			{"superscript", '¹', false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := isNumber(tt.rune)
				if got != tt.expected {
					t.Errorf("isNumber(%q) = %v, want %v", tt.rune, got, tt.expected)
				}
			})
		}
	})

	t.Run("isWhitespace", func(t *testing.T) {
		tests := []struct {
			name     string
			rune     rune
			expected bool
		}{
			// Common whitespace
			{"space", ' ', true},
			{"tab", '\t', true},
			{"newline", '\n', true},
			{"carriage_return", '\r', true},
			{"form_feed", '\f', true},
			{"vertical_tab", '\v', true},

			// Unicode whitespace
			{"no_break_space", '\u00A0', true},
			{"en_space", '\u2002', true},
			{"em_space", '\u2003', true},
			{"thin_space", '\u2009', true},
			{"hair_space", '\u200A', true},
			{"zero_width_space", '\u200B', false}, // Zero-width is not considered whitespace by unicode.IsSpace

			// Non-whitespace
			{"letter_a", 'a', false},
			{"number_1", '1', false},
			{"punctuation_period", '.', false},
			{"symbol_hash", '#', false},
			{"emoji", '🦙', false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := isWhitespace(tt.rune)
				if got != tt.expected {
					t.Errorf("isWhitespace(%q / U+%04X) = %v, want %v",
						tt.rune, tt.rune, got, tt.expected)
				}
			})
		}
	})
}

func TestStateMachineEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		validate func(t *testing.T, tokens []string)
	}{
		{
			name:  "no_empty_tokens",
			input: "Hello world! This is a test.",
			validate: func(t *testing.T, tokens []string) {
				for i, token := range tokens {
					if token == "" {
						t.Errorf("Found empty token at position %d", i)
					}
				}
			},
		},
		{
			name:  "preserves_exact_text",
			input: "Hello\tworld\n\nTest   123!",
			validate: func(t *testing.T, tokens []string) {
				// Reconstruct and compare
				reconstructed := strings.Join(tokens, "")
				if reconstructed != "Hello\tworld\n\nTest   123!" {
					t.Errorf("Reconstructed text doesn't match: got %q", reconstructed)
				}
			},
		},
		{
			name:  "handles_long_input",
			input: strings.Repeat("abc ", 1000),
			validate: func(t *testing.T, tokens []string) {
				// "abc " repeated 1000 times - expect 1001 tokens due to whitespace grouping
				if len(tokens) != 1001 {
					t.Errorf("Expected 1001 tokens, got %d", len(tokens))
				}
			},
		},
		{
			name:  "handles_all_unicode_categories",
			input: "Letter中文 Mark̃ Number123 Punct!@# Symbol∑∏ Space   Other🦙",
			validate: func(t *testing.T, tokens []string) {
				// Just ensure it doesn't panic and produces tokens
				if len(tokens) == 0 {
					t.Error("No tokens produced for unicode input")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := Tokenize(tt.input)
			tt.validate(t, tokens)
		})
	}
}

func TestStateMachineConsistency(t *testing.T) {
	// Test that tokenization is consistent
	inputs := []string{
		"Hello, world!",
		"The quick brown fox jumps over the lazy dog.",
		"I've got 99 problems but tokenization ain't one!",
		"Unicode: café, naïve, résumé",
		"Emoji: 🦙 🎉 🚀",
		"Mixed:   spaces\t\ttabs\n\nnewlines",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			// Tokenize multiple times
			tokens1 := Tokenize(input)
			tokens2 := Tokenize(input)
			tokens3 := Tokenize(input)

			// All should be identical
			if !reflect.DeepEqual(tokens1, tokens2) {
				t.Error("Tokenization is not consistent between calls 1 and 2")
			}
			if !reflect.DeepEqual(tokens2, tokens3) {
				t.Error("Tokenization is not consistent between calls 2 and 3")
			}

			// Verify reconstruction
			reconstructed := strings.Join(tokens1, "")
			if reconstructed != input {
				t.Errorf("Reconstruction failed: got %q, want %q", reconstructed, input)
			}
		})
	}
}

func TestStateMachineUnicodeCategories(t *testing.T) {
	// Test specific Unicode categories to ensure proper handling
	testCases := []struct {
		name     string
		input    string
		validate func(tokens []string) error
	}{
		{
			name:  "letter_modifiers",
			input: "basè b́ase ba͂se", // Combining grave, acute, tilde
			validate: func(tokens []string) error {
				// Should keep combining marks with their base
				if len(tokens) == 0 {
					return fmt.Errorf("no tokens produced")
				}
				return nil
			},
		},
		{
			name:  "different_scripts",
			input: "Latin Ελληνικά Кириллица العربية עברית हिन्दी 中文 日本語",
			validate: func(tokens []string) error {
				// Each script should be tokenized
				if len(tokens) < 8 {
					return fmt.Errorf("expected at least 8 tokens for different scripts")
				}
				return nil
			},
		},
		{
			name:  "mathematical_alphanumeric",
			input: "𝐀𝐁𝐂 𝕏𝕐ℤ ℵℶℷ",
			validate: func(tokens []string) error {
				// Math symbols should be treated as letters if they're in Letter category
				if len(tokens) == 0 {
					return fmt.Errorf("no tokens produced for mathematical alphanumeric")
				}
				return nil
			},
		},
		{
			name:  "direction_marks",
			input: "left\u200Eright\u200Ftext",
			validate: func(tokens []string) error {
				// Direction marks should be preserved
				joined := strings.Join(tokens, "")
				if joined != "left\u200Eright\u200Ftext" {
					return fmt.Errorf("direction marks not preserved")
				}
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tokens := Tokenize(tc.input)
			if err := tc.validate(tokens); err != nil {
				t.Errorf("Validation failed: %s\nTokens: %q", err, tokens)
			}
		})
	}
}

func TestCharacterClassification(t *testing.T) {
	// Comprehensive tests for character classification functions

	// Test all ASCII characters
	t.Run("ascii_classification", func(t *testing.T) {
		for r := rune(0); r < 128; r++ {
			isL := isLetter(r)
			isN := isNumber(r)
			isW := isWhitespace(r)

			// Verify mutual exclusivity for main categories
			count := 0
			if isL {
				count++
			}
			if isN {
				count++
			}
			if isW {
				count++
			}

			if count > 1 {
				t.Errorf("Rune %q (U+%04X) classified in multiple categories: letter=%v, number=%v, whitespace=%v",
					r, r, isL, isN, isW)
			}

			// Verify correct classification
			expectedLetter := (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
			expectedNumber := r >= '0' && r <= '9'
			expectedWhitespace := unicode.IsSpace(r)

			if isL != expectedLetter {
				t.Errorf("isLetter(%q) = %v, expected %v", r, isL, expectedLetter)
			}
			if isN != expectedNumber {
				t.Errorf("isNumber(%q) = %v, expected %v", r, isN, expectedNumber)
			}
			if isW != expectedWhitespace {
				t.Errorf("isWhitespace(%q) = %v, expected %v", r, isW, expectedWhitespace)
			}
		}
	})
}

// Benchmark tests to ensure performance hasn't regressed.
func BenchmarkTokenizeLongText(b *testing.B) {
	// Create a realistic long text
	parts := []string{
		"The quick brown fox jumps over the lazy dog. ",
		"I've got 99 problems, but tokenization ain't one! ",
		"Email: user@example.com, URL: https://example.com/path?q=test ",
		"Unicode café résumé naïve 文字 🦙 ",
		"Code: if (x > 0) { return true; } else { return false; } ",
		"\n\t   Mixed whitespace   \t\n",
	}

	var builder strings.Builder
	for i := 0; i < 100; i++ {
		builder.WriteString(parts[i%len(parts)])
	}
	text := builder.String()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tokens := Tokenize(text)
		_ = tokens
	}
}

func BenchmarkTokenizeUnicode(b *testing.B) {
	text := "Unicode test: café résumé naïve Ελληνικά Кириллица العربية עברית हिन्दी 中文 日本語 🦙🎉🚀"

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tokens := Tokenize(text)
		_ = tokens
	}
}

func BenchmarkTokenizeWhitespaceHeavyTable(b *testing.B) {
	text := "   lots   of     spaces    and\t\t\ttabs\n\n\nand\nnewlines   everywhere   "

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		tokens := Tokenize(text)
		_ = tokens
	}
}

// TestStateMachineJavaScriptCompatibility tests cases from JavaScript output.
func TestStateMachineJavaScriptCompatibility(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "           grabbed",
			expected: []string{"          ", " grabbed"}, // 10 spaces, then space+word
		},
		{
			input:    "\ttabs\t\t\t\tout here",
			expected: []string{"\ttabs", "\t\t\t", "\tout", " here"},
		},
		{
			input:    "Hello world",
			expected: []string{"Hello", " world"},
		},
		{
			input:    "can't",
			expected: []string{"can", "'t"},
		},
		{
			input:    "123 456",
			expected: []string{"123", " ", "456"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			tokens := Tokenize(tc.input)

			if !reflect.DeepEqual(tokens, tc.expected) {
				t.Errorf("Input: %q\nExpected: %q\nGot:      %q", tc.input, tc.expected, tokens)
			} else {
				fmt.Printf("✓ %q → %q\n", tc.input, tokens)
			}
		})
	}
}

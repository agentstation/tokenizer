package llama3

import (
	"reflect"
	"strings"
	"testing"
)

func TestTokenizerEncode(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	testGroups := map[string][]struct {
		name     string
		input    string
		expected []int
		opts     *EncodeOptions
	}{
		"basic": {
			{
				name:     "simple_word",
				input:    "grabbed",
				expected: []int{59312, 2788},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
			{
				name:     "word_with_space",
				input:    " grabbed",
				expected: []int{30418},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
			{
				name:     "multiple_spaces",
				input:    "           grabbed",
				expected: []int{1881, 30418},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
			{
				name:     "official_test_sentence",
				input:    "This is a test sentence.",
				expected: []int{2028, 374, 264, 1296, 11914, 13},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
		},
		"whitespace": {
			{
				name:     "newline",
				input:    "\n",
				expected: []int{198},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
			{
				name:     "space_newline",
				input:    " \n",
				expected: []int{720},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
			{
				name:     "tabs",
				input:    "\ttabs\t\t\t\tout here",
				expected: []int{3324, 3518, 573, 14294, 1618},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
		},
		"unicode": {
			{
				name:     "chinese_in_vocab",
				input:    "镇",
				expected: []int{104643},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
			{
				name:     "emoji_not_in_vocab",
				input:    "🦙",
				expected: []int{9468, 99, 247},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
			{
				name:     "mixed_utf8",
				input:    "🦙Ꙋ",
				expected: []int{9468, 99, 247, 166, 247, 232},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
		},
		"options": {
			{
				name:     "with_bos_eos",
				input:    "I",
				expected: []int{128000, 40, 128001},
				opts:     &EncodeOptions{BOS: true, EOS: true},
			},
			{
				name:     "bos_only",
				input:    "I",
				expected: []int{128000, 40},
				opts:     &EncodeOptions{BOS: true, EOS: false},
			},
			{
				name:     "eos_only",
				input:    "I",
				expected: []int{40, 128001},
				opts:     &EncodeOptions{BOS: false, EOS: true},
			},
			{
				name:     "default_opts",
				input:    "I",
				expected: []int{128000, 40, 128001},
				opts:     nil, // Will use defaults
			},
			{
				name:     "empty_with_bos_eos",
				input:    "",
				expected: []int{128000, 128001},
				opts:     &EncodeOptions{BOS: true, EOS: true},
			},
		},
		"special_tokens": {
			{
				name:     "special_tokens_in_text",
				input:    "<|start_header_id|>This text has special tokens<|eom_id|> in the middle of it.<|end_header_id|><|eot_id|>",
				expected: []int{128006, 2028, 1495, 706, 3361, 11460, 128008, 304, 279, 6278, 315, 433, 13, 128007, 128009},
				opts:     &EncodeOptions{BOS: false, EOS: false},
			},
		},
	}

	for groupName, tests := range testGroups {
		t.Run(groupName, func(t *testing.T) {
			for _, tt := range tests {
				t.Run(tt.name, func(t *testing.T) {
					got := tokenizer.Encode(tt.input, tt.opts)
					if !reflect.DeepEqual(got, tt.expected) {
						t.Errorf("Encode(%q) = %v, want %v", tt.input, got, tt.expected)
					}

					// Test round-trip encoding/decoding for non-special cases
					if groupName != "options" && groupName != "special_tokens" && tt.opts != nil && !tt.opts.BOS && !tt.opts.EOS {
						decoded := tokenizer.Decode(got)
						if decoded != tt.input {
							t.Errorf("Decode(Encode(%q)) = %q, want %q", tt.input, decoded, tt.input)
						}
					}
				})
			}
		})
	}
}

func TestTokenizerDecode(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	tests := []struct {
		name     string
		input    []int
		expected string
	}{
		{
			name:     "simple_text",
			input:    []int{9906, 1917, 0},
			expected: "Hello world!",
		},
		{
			name:     "with_special_tokens",
			input:    []int{128000, 40, 128001},
			expected: "<|begin_of_text|>I<|end_of_text|>",
		},
		{
			name:     "empty_input",
			input:    []int{},
			expected: "",
		},
		{
			name:     "invalid_token_ids",
			input:    []int{-1, 999999999},
			expected: "",
		},
		{
			name:     "special_token_ids",
			input:    []int{128000, 128006, 128004, 128008, 128010},
			expected: "<|begin_of_text|><|start_header_id|><|finetune_right_pad_id|><|eom_id|><|python_tag|>",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tokenizer.Decode(tt.input)
			if got != tt.expected {
				t.Errorf("Decode(%v) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestTokenizerSpecialTokens(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	tests := []struct {
		name        string
		token       string
		wantID      int
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid_begin_of_text",
			token:   "<|begin_of_text|>",
			wantID:  128000,
			wantErr: false,
		},
		{
			name:    "valid_end_of_text",
			token:   "<|end_of_text|>",
			wantID:  128001,
			wantErr: false,
		},
		{
			name:        "invalid_format",
			token:       "not_a_special_token",
			wantErr:     true,
			errContains: "invalid token",
		},
		{
			name:        "unknown_special_token",
			token:       "<|unknown_token|>",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "empty_token",
			token:       "",
			wantErr:     true,
			errContains: "invalid token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := tokenizer.GetSpecialTokenID(tt.token)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetSpecialTokenID(%q) error = nil, wantErr %v", tt.token, tt.wantErr)
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("GetSpecialTokenID(%q) error = %v, want error containing %q", tt.token, err, tt.errContains)
				}
			} else {
				if err != nil {
					t.Errorf("GetSpecialTokenID(%q) error = %v, wantErr %v", tt.token, err, tt.wantErr)
					return
				}
				if gotID != tt.wantID {
					t.Errorf("GetSpecialTokenID(%q) = %d, want %d", tt.token, gotID, tt.wantID)
				}
			}
		})
	}
}

func TestTokenizerProperties(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	t.Run("vocabulary_size", func(t *testing.T) {
		vocabSize := tokenizer.VocabSize()

		if vocabSize != totalVocabSize {
			t.Errorf("VocabSize() = %d, want %d", vocabSize, totalVocabSize)
		}
	})

	t.Run("deterministic_encoding", func(t *testing.T) {
		inputs := []string{
			"Hello, world!",
			"The quick brown fox",
			"Special chars: @#$%",
			"Unicode: 你好世界 🦙",
			"Mixed case: AbCdEfG",
			"Numbers: 123456789",
			"Punctuation: !@#$%^&*()",
		}

		for _, input := range inputs {
			t.Run(input, func(t *testing.T) {
				// Encode the same input multiple times
				opts := &EncodeOptions{BOS: false, EOS: false}
				result1 := tokenizer.Encode(input, opts)
				result2 := tokenizer.Encode(input, opts)
				result3 := tokenizer.Encode(input, opts)

				if !reflect.DeepEqual(result1, result2) || !reflect.DeepEqual(result2, result3) {
					t.Errorf("Non-deterministic encoding for %q", input)
					t.Logf("Result 1: %v", result1)
					t.Logf("Result 2: %v", result2)
					t.Logf("Result 3: %v", result3)
				}
			})
		}
	})
}

func TestLargeText(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	// Large text from the JavaScript tests
	largeText := `The llama (/ˈlɑːmə/; 🦙Spanish pronunciation: [ˈʎama]) (Lama glama) is a domesticated South American camelid, widely used as a meat and pack animal by Andean cultures since the Pre-Columbian era. Llamas are social animals and live with others as a herd. Their wool is soft and contains only a small amount of lanolin.[2] Llamas can learn simple tasks after a few repetitions. When using a pack, they can carry about 25 to 30% of their body weight for 8 to 13 km (5–8 miles).[3] The name llama (in the past also spelled "lama" or "glama") was adopted by European settlers from native Peruvians.[4] The ancestors of llamas are thought to have originated from the Great Plains of North America about 40 million years ago, and subsequently migrated to South America about three million years ago during the Great American Interchange. By the end of the last ice age (10,000–12,000 years ago), camelids were extinct in North America.[3] As of 2007, there were over seven million llamas and alpacas in South America and over 158,000 llamas and 100,000Ꙋ🦙 alpacas, descended from progenitors imported late in the 20th century, in the United States and Canada.[5] In Aymara mythology, llamas are important beings. The Heavenly Llama is said to drink water from the ocean and urinates as it rains.[6] According to Aymara eschatology, llamas will return to the water springs and lagoons where they come from at the end of time.[6]`

	expectedLength := 373 // Verified from JS implementation (including BOS/EOS)

	tokens := tokenizer.Encode(largeText, nil)
	if len(tokens) != expectedLength {
		t.Errorf("Large text encoding length = %d, want %d", len(tokens), expectedLength)
	}

	// Test that we can decode it back
	decoded := tokenizer.Decode(tokens)
	// The decoded text will have BOS/EOS tokens, so we need to strip them for comparison
	decodedWithoutSpecial := strings.TrimPrefix(decoded, "<|begin_of_text|>")
	decodedWithoutSpecial = strings.TrimSuffix(decodedWithoutSpecial, "<|end_of_text|>")

	if decodedWithoutSpecial != largeText {
		t.Errorf("Large text round-trip failed")
		// Log first 100 chars to avoid huge output
		if len(decodedWithoutSpecial) > 100 {
			t.Logf("Decoded (first 100 chars): %q...", decodedWithoutSpecial[:100])
			t.Logf("Expected (first 100 chars): %q...", largeText[:100])
		} else {
			t.Logf("Decoded: %q", decodedWithoutSpecial)
			t.Logf("Expected: %q", largeText)
		}
	}
}

// TestEncodeBytesMethod tests the EncodeBytes method
func TestEncodeBytesMethod(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	text := "Hello, world!"
	opts := &EncodeOptions{BOS: false, EOS: false}

	// Encode string vs bytes should give same result
	stringTokens := tokenizer.Encode(text, opts)
	byteTokens := tokenizer.EncodeBytes([]byte(text), opts)

	if !reflect.DeepEqual(stringTokens, byteTokens) {
		t.Errorf("EncodeBytes() = %v, Encode() = %v", byteTokens, stringTokens)
	}
}

// TestAppendTokensMethod tests the AppendTokens method
func TestAppendTokensMethod(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	opts := &EncodeOptions{BOS: false, EOS: false}

	t.Run("append_to_nil", func(t *testing.T) {
		result := tokenizer.AppendTokens(nil, "Hello", opts)
		expected := tokenizer.Encode("Hello", opts)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("AppendTokens(nil) = %v, want %v", result, expected)
		}
	})

	t.Run("append_to_existing", func(t *testing.T) {
		initial := tokenizer.Encode("Hello", opts)
		result := tokenizer.AppendTokens(initial, " world", opts)
		expected := tokenizer.Encode("Hello world", opts)
		if !reflect.DeepEqual(result, expected) {
			t.Errorf("AppendTokens() = %v, want %v", result, expected)
		}
	})

	t.Run("append_with_capacity", func(t *testing.T) {
		// Pre-allocate with extra capacity
		initial := make([]int, 0, 100)
		initial = tokenizer.AppendTokens(initial, "Hello", opts)
		capBefore := cap(initial)

		result := tokenizer.AppendTokens(initial, " world", opts)

		// Should reuse the same backing array if capacity was sufficient
		if cap(result) < capBefore {
			t.Errorf("AppendTokens() did not reuse capacity: cap = %d, want >= %d", cap(result), capBefore)
		}
	})
}

// TestOptimisticCount tests the OptimisticCount method
func TestOptimisticCount(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	tests := []struct {
		name     string
		input    string
		minCount int // We know it should be at least this many
	}{
		{
			name:     "simple_text",
			input:    "Hello world",
			minCount: 3, // At least BOS + tokens + EOS
		},
		{
			name:     "with_real_special_tokens",
			input:    "<|begin_of_text|>Hello<|end_of_text|>",
			minCount: 4, // BOS + special + Hello + special + EOS
		},
		{
			name:     "with_unknown_special_tokens",
			input:    "<|custom_token|>Hello<|another_custom|>",
			minCount: 5, // Should count unknown special tokens as 1 each
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			count := tokenizer.OptimisticCount(tt.input)
			if count < tt.minCount {
				t.Errorf("OptimisticCount(%q) = %d, want >= %d", tt.input, count, tt.minCount)
			}
		})
	}
}

// TestDecodeBytesMethod tests the DecodeBytes method
func TestDecodeBytesMethod(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	text := "Hello, world! 🦙"
	opts := &EncodeOptions{BOS: false, EOS: false}

	// Encode and decode
	tokens := tokenizer.Encode(text, opts)
	decodedBytes := tokenizer.DecodeBytes(tokens)
	decodedString := tokenizer.Decode(tokens)

	// Both methods should produce the same result
	if string(decodedBytes) != decodedString {
		t.Errorf("DecodeBytes() = %q, Decode() = %q", string(decodedBytes), decodedString)
	}

	// Should match original
	if string(decodedBytes) != text {
		t.Errorf("DecodeBytes() = %q, want %q", string(decodedBytes), text)
	}
}

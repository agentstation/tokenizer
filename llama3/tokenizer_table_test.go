package llama3

import (
	"reflect"
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, err := tokenizer.GetSpecialTokenID(tt.token)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetSpecialTokenID(%q) error = nil, wantErr %v", tt.token, tt.wantErr)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
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
				}
			})
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || 
		len(s) > len(substr) && contains(s[1:], substr)
}
package llama3

import (
	"reflect"
	"testing"
)

// Note: These tests use mock data. In production, you would load the actual
// Llama 3 vocabulary and merge data files.

func TestEncodeDecodeBasic(t *testing.T) {
	// Note: These tests would require loading actual Llama 3 data
	// For now, we'll skip them in CI unless data is available
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	// This test suite ports all the tests from the JavaScript implementation
	tests := []struct {
		name       string
		input      string
		expected   []int
		encodeOpts *EncodeOptions
		skipDecode bool
	}{
		// Simple test cases
		{
			name:       "simple word",
			input:      "grabbed",
			expected:   []int{59312, 2788},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		{
			name:       "word with space",
			input:      " grabbed",
			expected:   []int{30418},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		{
			name:       "multiple spaces",
			input:      "           grabbed",
			expected:   []int{1881, 30418},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		// Linebreak and tab handling
		{
			name:       "newline",
			input:      "\n",
			expected:   []int{198},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		{
			name:       "space newline",
			input:      " \n",
			expected:   []int{720},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		{
			name:       "tabs",
			input:      "\ttabs\t\t\t\tout here",
			expected:   []int{3324, 3518, 573, 14294, 1618},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		// UTF-8 characters
		{
			name:       "chinese character in vocab",
			input:      "镇",
			expected:   []int{104643},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		{
			name:       "emoji not in vocab",
			input:      "🦙",
			expected:   []int{9468, 99, 247},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		{
			name:       "mixed UTF-8",
			input:      "🦙Ꙋ",
			expected:   []int{9468, 99, 247, 166, 247, 232},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		// Official test case
		{
			name:       "official test sentence",
			input:      "This is a test sentence.",
			expected:   []int{2028, 374, 264, 1296, 11914, 13},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
		// Encoder options tests
		{
			name:       "with BOS and EOS",
			input:      "I",
			expected:   []int{128000, 40, 128001},
			encodeOpts: &EncodeOptions{BOS: true, EOS: true},
		},
		{
			name:       "with BOS only",
			input:      "I",
			expected:   []int{128000, 40},
			encodeOpts: &EncodeOptions{BOS: true, EOS: false},
		},
		{
			name:       "with EOS only",
			input:      "I",
			expected:   []int{40, 128001},
			encodeOpts: &EncodeOptions{BOS: false, EOS: true},
		},
		{
			name:       "empty with BOS and EOS",
			input:      "",
			expected:   []int{128000, 128001},
			encodeOpts: &EncodeOptions{BOS: true, EOS: true},
		},
		{
			name:       "default options",
			input:      "I",
			expected:   []int{128000, 40, 128001},
			encodeOpts: nil, // Will use defaults
		},
		// Special tokens
		{
			name:       "special tokens in text",
			input:      "<|start_header_id|>This text has special tokens<|eom_id|> in the middle of it.<|end_header_id|><|eot_id|>",
			expected:   []int{128006, 2028, 1495, 706, 3361, 11460, 128008, 304, 279, 6278, 315, 433, 13, 128007, 128009},
			encodeOpts: &EncodeOptions{BOS: false, EOS: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test encoding
			got := tokenizer.Encode(tt.input, tt.encodeOpts)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Encode() = %v, want %v", got, tt.expected)
			}

			// Test decoding (unless explicitly skipped)
			if !tt.skipDecode && tt.encodeOpts != nil && !tt.encodeOpts.BOS && !tt.encodeOpts.EOS {
				decoded := tokenizer.Decode(got)
				if decoded != tt.input {
					t.Errorf("Decode() = %q, want %q", decoded, tt.input)
				}
			}
		})
	}
}

func TestSpecialTokens(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	specialTokenTests := []struct {
		tokenID  int
		expected string
	}{
		{128000, "<|begin_of_text|>"},
		{128006, "<|start_header_id|>"},
		{128004, "<|finetune_right_pad_id|>"},
		{128008, "<|eom_id|>"},
		{128010, "<|python_tag|>"},
	}

	for _, tt := range specialTokenTests {
		t.Run(tt.expected, func(t *testing.T) {
			decoded := tokenizer.Decode([]int{tt.tokenID})
			if decoded != tt.expected {
				t.Errorf("Decode([%d]) = %q, want %q", tt.tokenID, decoded, tt.expected)
			}
		})
	}
}

func TestGetSpecialTokenID(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	// Test valid special token
	id, err := tokenizer.GetSpecialTokenID("<|begin_of_text|>")
	if err != nil {
		t.Errorf("GetSpecialTokenID() error = %v", err)
	}
	if id != 128000 {
		t.Errorf("GetSpecialTokenID() = %d, want %d", id, 128000)
	}

	// Test invalid format
	_, err = tokenizer.GetSpecialTokenID("not_a_special_token")
	if err == nil {
		t.Error("GetSpecialTokenID() expected error for invalid format")
	}

	// Test non-existent special token
	_, err = tokenizer.GetSpecialTokenID("<|fake_token|>")
	if err == nil {
		t.Error("GetSpecialTokenID() expected error for non-existent token")
	}
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
}

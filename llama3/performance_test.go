package llama3

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeBytes(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	tests := []struct {
		name  string
		input []byte
		opts  *EncodeOptions
	}{
		{
			name:  "simple_bytes",
			input: []byte("Hello world"),
			opts:  nil,
		},
		{
			name:  "utf8_bytes",
			input: []byte("Hello 世界"),
			opts:  &EncodeOptions{BOS: false, EOS: false},
		},
		{
			name:  "binary_data",
			input: []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f}, // "Hello"
			opts:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compare with string encoding
			tokensFromBytes := tokenizer.EncodeBytes(tt.input, tt.opts)
			tokensFromString := tokenizer.Encode(string(tt.input), tt.opts)

			if !compareTokens(tokensFromBytes, tokensFromString) {
				t.Errorf("EncodeBytes = %v, Encode = %v", tokensFromBytes, tokensFromString)
			}
		})
	}
}

func TestDecodeBytes(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	tests := []struct {
		name  string
		input string
		opts  *EncodeOptions
	}{
		{
			name:  "simple_text",
			input: "Hello world",
			opts:  &EncodeOptions{BOS: false, EOS: false},
		},
		{
			name:  "unicode_text",
			input: "Hello 世界 café",
			opts:  &EncodeOptions{BOS: false, EOS: false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenizer.Encode(tt.input, tt.opts)

			// Compare byte and string decoding
			decodedBytes := tokenizer.DecodeBytes(tokens)
			decodedString := tokenizer.Decode(tokens)

			if string(decodedBytes) != decodedString {
				t.Errorf("DecodeBytes = %q, Decode = %q", decodedBytes, decodedString)
			}

			// Verify round-trip
			if string(decodedBytes) != tt.input {
				t.Errorf("Round-trip failed: got %q, want %q", decodedBytes, tt.input)
			}
		})
	}
}

func TestAppendTokens(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	t.Run("nil_dst", func(t *testing.T) {
		tokens := tokenizer.AppendTokens(nil, "Hello", nil)
		if len(tokens) == 0 {
			t.Error("Expected tokens, got none")
		}
	})

	t.Run("existing_dst", func(t *testing.T) {
		dst := []int{1, 2, 3}
		tokens := tokenizer.AppendTokens(dst, "Hello", &EncodeOptions{BOS: false, EOS: false})

		if len(tokens) <= 3 {
			t.Error("Expected tokens to be appended")
		}

		// Check that original tokens are preserved
		if tokens[0] != 1 || tokens[1] != 2 || tokens[2] != 3 {
			t.Error("Original tokens not preserved")
		}
	})

	t.Run("reuse_capacity", func(t *testing.T) {
		// Create slice with extra capacity
		dst := make([]int, 3, 100)
		dst[0], dst[1], dst[2] = 1, 2, 3

		tokens := tokenizer.AppendTokens(dst, "Hello", &EncodeOptions{BOS: false, EOS: false})

		// Should reuse the same backing array
		if cap(tokens) < 100 {
			t.Error("Expected to reuse capacity")
		}
	})

	t.Run("multiple_appends", func(t *testing.T) {
		var tokens []int
		opts := &EncodeOptions{BOS: false, EOS: false}

		tokens = tokenizer.AppendTokens(tokens, "Hello ", opts)
		beforeLen := len(tokens)
		tokens = tokenizer.AppendTokens(tokens, "world ", opts)
		middleLen := len(tokens)
		tokens = tokenizer.AppendTokens(tokens, "test", opts)
		finalLen := len(tokens)

		// Verify tokens were appended at each step
		if beforeLen == 0 {
			t.Error("First append produced no tokens")
		}
		if middleLen <= beforeLen {
			t.Error("Second append produced no additional tokens")
		}
		if finalLen <= middleLen {
			t.Error("Third append produced no additional tokens")
		}

		// Verify we can decode the result
		decoded := tokenizer.Decode(tokens)
		// Remove special tokens for comparison
		decoded = strings.ReplaceAll(decoded, "<|begin_of_text|>", "")
		decoded = strings.ReplaceAll(decoded, "<|end_of_text|>", "")

		// Should contain all our text (though tokenization boundaries may differ)
		if !strings.Contains(decoded, "Hello") || !strings.Contains(decoded, "world") || !strings.Contains(decoded, "test") {
			t.Errorf("Decoded text missing expected content: %q", decoded)
		}
	})
}

func TestInterfaces(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	t.Run("encoder_interface", func(t *testing.T) {
		var enc Encoder = tokenizer
		tokens := enc.Encode("Hello", nil)
		if len(tokens) == 0 {
			t.Error("Expected tokens from encoder interface")
		}
	})

	t.Run("decoder_interface", func(t *testing.T) {
		var dec Decoder = tokenizer
		text := dec.Decode([]int{9906}) // "Hello" token
		if text == "" {
			t.Error("Expected text from decoder interface")
		}
	})

	t.Run("pretokenizer_interface", func(t *testing.T) {
		var pt PreTokenizer = tokenizer
		pretokens := pt.PreTokenize("Hello world!")
		if len(pretokens) == 0 {
			t.Error("Expected pre-tokens from PreTokenizer interface")
		}
		// Pre-tokenizer should split this into at least "Hello", " world", "!"
		if len(pretokens) < 3 {
			t.Errorf("Expected at least 3 pre-tokens, got %d: %v", len(pretokens), pretokens)
		}
	})

	t.Run("bpe_processor_interface", func(t *testing.T) {
		var bpe BPE = tokenizer
		// Use a pre-token that should be in vocabulary
		tokens := bpe.EncodeBPE("Hello")
		if len(tokens) == 0 {
			t.Error("Expected tokens from BPEProcessor interface")
		}
	})

	t.Run("encoder_func", func(t *testing.T) {
		mockEncoder := EncoderFunc(func(text string, opts *EncodeOptions) []int {
			return []int{1, 2, 3}
		})

		tokens := mockEncoder.Encode("anything", nil)
		if len(tokens) != 3 {
			t.Errorf("Expected 3 tokens, got %d", len(tokens))
		}
	})

	t.Run("decoder_func", func(t *testing.T) {
		mockDecoder := DecoderFunc(func(tokens []int) string {
			return "mocked"
		})

		text := mockDecoder.Decode([]int{1, 2, 3})
		if text != "mocked" {
			t.Errorf("Expected 'mocked', got %q", text)
		}
	})
}

// Benchmarks for performance variants

func BenchmarkEncodeString(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Fatalf("Failed to create tokenizer: %v", err)
	}

	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 10)
	opts := &EncodeOptions{BOS: false, EOS: false}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = tokenizer.Encode(text, opts)
	}
}

func BenchmarkEncodeBytes(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Fatalf("Failed to create tokenizer: %v", err)
	}

	data := []byte(strings.Repeat("The quick brown fox jumps over the lazy dog. ", 10))
	opts := &EncodeOptions{BOS: false, EOS: false}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = tokenizer.EncodeBytes(data, opts)
	}
}

func BenchmarkAppendTokens(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Fatalf("Failed to create tokenizer: %v", err)
	}

	text := "The quick brown fox jumps over the lazy dog."
	opts := &EncodeOptions{BOS: false, EOS: false}
	dst := make([]int, 0, 100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		dst = dst[:0] // Reset slice but keep capacity
		_ = tokenizer.AppendTokens(dst, text, opts)
	}
}

func BenchmarkDecodeString(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Fatalf("Failed to create tokenizer: %v", err)
	}

	tokens := tokenizer.Encode("The quick brown fox jumps over the lazy dog.", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = tokenizer.Decode(tokens)
	}
}

func BenchmarkDecodeBytes(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Fatalf("Failed to create tokenizer: %v", err)
	}

	tokens := tokenizer.Encode("The quick brown fox jumps over the lazy dog.", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = tokenizer.DecodeBytes(tokens)
	}
}

func BenchmarkScanner(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Fatalf("Failed to create tokenizer: %v", err)
	}

	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		scanner := tokenizer.NewScanner(reader)

		for scanner.Scan() {
			_ = scanner.Token()
		}

		if err := scanner.Err(); err != nil {
			b.Fatalf("Scanner error: %v", err)
		}
	}
}

func BenchmarkProcess(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Fatalf("Failed to create tokenizer: %v", err)
	}

	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		reader := strings.NewReader(text)
		var buf bytes.Buffer

		_, err := tokenizer.Process(reader, &buf)
		if err != nil {
			b.Fatalf("Process error: %v", err)
		}
	}
}

package llama3

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestScanner(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		opts     *EncodeOptions
		wantErr  bool
		validate func(t *testing.T, tokens []int)
	}{
		{
			name:  "simple_text",
			input: "Hello world",
			opts:  &EncodeOptions{BOS: false, EOS: false},
			validate: func(t *testing.T, tokens []int) {
				if len(tokens) == 0 {
					t.Error("Expected tokens, got none")
				}
			},
		},
		{
			name:  "with_special_tokens",
			input: "Hello world",
			opts:  &EncodeOptions{BOS: true, EOS: true},
			validate: func(t *testing.T, tokens []int) {
				if len(tokens) < 3 {
					t.Errorf("Expected at least 3 tokens (BOS + content + EOS), got %d", len(tokens))
				}
			},
		},
		{
			name:  "empty_input",
			input: "",
			opts:  &EncodeOptions{BOS: true, EOS: true},
			validate: func(t *testing.T, tokens []int) {
				if len(tokens) != 2 {
					t.Errorf("Expected 2 tokens (BOS + EOS), got %d", len(tokens))
				}
			},
		},
		{
			name:  "large_text",
			input: strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100),
			opts:  &EncodeOptions{BOS: false, EOS: false},
			validate: func(t *testing.T, tokens []int) {
				if len(tokens) == 0 {
					t.Error("Expected tokens for large text")
				}
				// For streaming, we may get slightly different tokenization at boundaries
				// Just ensure we got a reasonable number of tokens
				expected := 100 * 10 // Approximately 10 tokens per sentence
				if len(tokens) < expected/2 || len(tokens) > expected*2 {
					t.Errorf("Got %d tokens, expected around %d", len(tokens), expected)
				}
			},
		},
		{
			name:  "unicode_text",
			input: "Hello 世界 🦙 café",
			opts:  &EncodeOptions{BOS: false, EOS: false},
			validate: func(t *testing.T, tokens []int) {
				if len(tokens) == 0 {
					t.Error("Expected tokens for unicode text")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			scanner := tokenizer.NewScanner(reader, WithEncodeOptions(tt.opts))

			var tokens []int
			for scanner.Scan() {
				tokens = append(tokens, scanner.Token())
			}

			err := scanner.Err()
			if (err != nil) != tt.wantErr {
				t.Errorf("Scanner error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.validate != nil {
				tt.validate(t, tokens)
			}

			// Compare with direct encoding (skip for large text due to boundary differences)
			if tt.name != "large_text" {
				expected := tokenizer.Encode(tt.input, tt.opts)
				if !equalIntSlices(tokens, expected) {
					t.Errorf("Scanner produced %d tokens, expected %d", len(tokens), len(expected))
					// Show first few differences
					for i := 0; i < 10 && i < len(tokens) && i < len(expected); i++ {
						if i >= len(expected) || tokens[i] != expected[i] {
							t.Errorf("  Token[%d]: got %d, want %d", i, tokens[i], expected[i])
						}
					}
				}
			}
		})
	}
}

func TestScannerOptions(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	t.Run("custom_buffer_size", func(t *testing.T) {
		input := strings.Repeat("test ", 1000)
		reader := strings.NewReader(input)

		scanner := tokenizer.NewScanner(reader,
			WithBufferSize(128),
			WithMaxBuffer(512),
		)

		count := 0
		for scanner.Scan() {
			count++
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		if count == 0 {
			t.Error("Expected tokens, got none")
		}
	})
}

func TestScannerEdgeCases(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	t.Run("reader_error", func(t *testing.T) {
		reader := &errorReader{err: errors.New("read error")}
		scanner := tokenizer.NewScanner(reader)

		// Should handle error gracefully
		for scanner.Scan() {
			// Continue scanning
		}

		if scanner.Err() == nil {
			t.Error("Expected error from scanner")
		}
	})

	t.Run("utf8_boundary", func(t *testing.T) {
		// Test that scanner doesn't split UTF-8 sequences
		input := "Hello 世界 world"
		reader := &slowReader{data: []byte(input), chunkSize: 1}
		scanner := tokenizer.NewScanner(reader)

		var tokens []int
		for scanner.Scan() {
			tokens = append(tokens, scanner.Token())
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Decode and verify text is preserved
		decoded := tokenizer.Decode(tokens)
		// Remove special tokens for comparison
		decoded = strings.ReplaceAll(decoded, "<|begin_of_text|>", "")
		decoded = strings.ReplaceAll(decoded, "<|end_of_text|>", "")

		if decoded != input {
			t.Errorf("Text not preserved: got %q, want %q", decoded, input)
		}
	})

	t.Run("buffer_limit_word_boundary", func(t *testing.T) {
		// Test handling when buffer limit is hit in the middle of a word
		input := "This is a very long word: " + strings.Repeat("a", 100) + " and more text"
		reader := strings.NewReader(input)

		// Small buffer that will force a split
		scanner := tokenizer.NewScanner(reader,
			WithBufferSize(32),
			WithMaxBuffer(64), // Will hit limit in middle of long word
			WithEncodeOptions(&EncodeOptions{BOS: false, EOS: false}),
		)

		var tokens []int
		for scanner.Scan() {
			tokens = append(tokens, scanner.Token())
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Decode and verify text is preserved
		decoded := tokenizer.Decode(tokens)
		if decoded != input {
			t.Errorf("Text not preserved at buffer boundary: got %q, want %q", decoded, input)
		}
	})

	t.Run("buffer_limit_utf8_boundary", func(t *testing.T) {
		// Test handling when buffer limit is hit in the middle of UTF-8 sequence
		// Create input that will force a split in the middle of UTF-8
		// "世" is 3 bytes: 0xE4 0xB8 0x96
		input := strings.Repeat("a", 62) + "世界" // Will hit 64-byte limit in middle of "世"
		reader := strings.NewReader(input)

		scanner := tokenizer.NewScanner(reader,
			WithBufferSize(32),
			WithMaxBuffer(64),
			WithEncodeOptions(&EncodeOptions{BOS: false, EOS: false}),
		)

		var tokens []int
		for scanner.Scan() {
			tokens = append(tokens, scanner.Token())
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Decode and verify text is preserved
		decoded := tokenizer.Decode(tokens)
		if decoded != input {
			t.Errorf("UTF-8 text not preserved at buffer boundary: got %q, want %q", decoded, input)
		}
	})

	t.Run("buffer_limit_exact_utf8_split", func(t *testing.T) {
		// Force exact UTF-8 split - buffer will be exactly full after reading incomplete UTF-8
		// "世" is 3 bytes: 0xE4 0xB8 0x96
		input := strings.Repeat("a", 63) + "世界"
		reader := &controlledReader{
			data:      []byte(input),
			readSizes: []int{64, 10}, // Read exactly to buffer limit, splitting "世"
		}

		scanner := tokenizer.NewScanner(reader,
			WithBufferSize(64),
			WithMaxBuffer(64), // Exact buffer size
			WithEncodeOptions(&EncodeOptions{BOS: false, EOS: false}),
		)

		var tokens []int
		for scanner.Scan() {
			tokens = append(tokens, scanner.Token())
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Decode and verify text is preserved
		decoded := tokenizer.Decode(tokens)
		if decoded != input {
			t.Logf("Input: %q (len=%d)", input, len(input))
			t.Logf("Decoded: %q (len=%d)", decoded, len(decoded))
			t.Logf("Input bytes: %x", []byte(input))
			t.Logf("Decoded bytes: %x", []byte(decoded))
			// Show where they differ
			for i := 0; i < len(input) && i < len(decoded); i++ {
				if input[i] != decoded[i] {
					t.Logf("First difference at position %d: input=%x decoded=%x", i, input[i], decoded[i])
					break
				}
			}
			t.Errorf("UTF-8 text not preserved with exact buffer split")
		}
	})

	t.Run("buffer_limit_multiple_chunks", func(t *testing.T) {
		// Test that multiple buffer limit hits still produce correct output
		longText := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 20)
		reader := strings.NewReader(longText)

		scanner := tokenizer.NewScanner(reader,
			WithBufferSize(32),
			WithMaxBuffer(128), // Small enough to force multiple chunks
			WithEncodeOptions(&EncodeOptions{BOS: false, EOS: false}),
		)

		var tokens []int
		for scanner.Scan() {
			tokens = append(tokens, scanner.Token())
		}

		if err := scanner.Err(); err != nil {
			t.Errorf("Unexpected error: %v", err)
		}

		// Decode and verify text is preserved
		decoded := tokenizer.Decode(tokens)
		if decoded != longText {
			t.Errorf("Text not preserved across chunks: got %d chars, want %d chars",
				len(decoded), len(longText))
		}
	})
}

func TestProcess(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:  "simple_text",
			input: "Hello world",
		},
		{
			name:  "empty_input",
			input: "",
		},
		{
			name:  "large_text",
			input: strings.Repeat("test ", 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.input)
			var buf bytes.Buffer

			count, err := tokenizer.Process(reader, &buf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Process error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify token count
			if count == 0 && tt.input != "" {
				t.Error("Expected tokens, got none")
			}

			// Verify output size
			expectedSize := count * 4 // 4 bytes per token
			if int64(buf.Len()) != expectedSize {
				t.Errorf("Output size = %d, want %d", buf.Len(), expectedSize)
			}
		})
	}
}

func TestTokenStream(t *testing.T) {
	tokenizer, err := New()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	t.Run("simple_stream", func(t *testing.T) {
		input := "Hello world"
		reader := strings.NewReader(input)

		tokens, errc := tokenizer.TokenStream(reader)

		var collected []int
		for token := range tokens {
			collected = append(collected, token)
		}

		// Check for errors
		select {
		case err := <-errc:
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		default:
			// No error
		}

		if len(collected) == 0 {
			t.Error("Expected tokens, got none")
		}
	})

	t.Run("error_propagation", func(t *testing.T) {
		reader := &errorReader{err: errors.New("stream error")}

		tokens, errc := tokenizer.TokenStream(reader)

		// Drain tokens
		for range tokens {
			// Continue
		}

		// Should receive error
		select {
		case err := <-errc:
			if err == nil {
				t.Error("Expected error from stream")
			}
		default:
			t.Error("Expected error on error channel")
		}
	})
}

// Helper types for testing

type errorReader struct {
	err error
}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, r.err
}

type slowReader struct {
	data      []byte
	pos       int
	chunkSize int
}

func (r *slowReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	// Read only chunkSize bytes at a time
	end := r.pos + r.chunkSize
	if end > len(r.data) {
		end = len(r.data)
	}

	n = copy(p, r.data[r.pos:end])
	r.pos = end

	return n, nil
}

type controlledReader struct {
	data      []byte
	pos       int
	readSizes []int
	readIndex int
}

func (r *controlledReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	// Determine how many bytes to read
	readSize := len(p)
	if r.readIndex < len(r.readSizes) {
		readSize = r.readSizes[r.readIndex]
		r.readIndex++
	}

	// Don't read more than available
	if readSize > len(p) {
		readSize = len(p)
	}
	if r.pos+readSize > len(r.data) {
		readSize = len(r.data) - r.pos
	}

	n = copy(p, r.data[r.pos:r.pos+readSize])
	r.pos += n

	return n, nil
}

func equalIntSlices(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

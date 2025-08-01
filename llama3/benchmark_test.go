package llama3

import (
	"strings"
	"testing"
	
	"github.com/agentstation/tokenizer/llama3/internal/pretokenizer"
)

// =============================================================================
// Core Component Benchmarks
// =============================================================================

func BenchmarkStateMachine(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. This is a test sentence with multiple spaces   and some punctuation!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

func BenchmarkEncode(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}

	text := "The quick brown fox jumps over the lazy dog."
	opts := &EncodeOptions{BOS: false, EOS: false}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenizer.Encode(text, opts)
	}
}

func BenchmarkDecode(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}

	// Encode some text first
	text := "The quick brown fox jumps over the lazy dog."
	tokens := tokenizer.Encode(text, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenizer.Decode(tokens)
	}
}

// =============================================================================
// Text Type Benchmarks
// =============================================================================

func BenchmarkTokenizeASCIIOnly(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. " +
		"This is a simple ASCII text with numbers 123 and punctuation!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

func BenchmarkTokenizeUnicodeHeavy(b *testing.B) {
	text := "こんにちは世界 🌍 Привет мир 🇷🇺 مرحبا بالعالم 🇸🇦 " +
		"Olá mundo 🇧🇷 你好世界 🇨🇳 नमस्ते दुनिया 🇮🇳"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

func BenchmarkTokenizeWhitespaceHeavy(b *testing.B) {
	text := "   Multiple   spaces   between   words   \t\t\tand\ttabs\t\t\t" +
		"\n\n\nand\nnewlines\n\n\n   with   trailing   spaces   "

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

func BenchmarkTokenizeCodeLike(b *testing.B) {
	text := `func main() {
	fmt.Println("Hello, world!")
	for i := 0; i < 10; i++ {
		result := calculate(i * 2)
		log.Printf("Result: %d\n", result)
	}
}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

func BenchmarkTokenizeMixedContent(b *testing.B) {
	text := "Email: user@example.com | Phone: +1-555-0123 | " +
		"Price: $99.99 (save 20%!) | URL: https://example.com/path?q=1&v=2#anchor"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

// =============================================================================
// Size-based Benchmarks
// =============================================================================

func BenchmarkTokenize_SmallText(b *testing.B) {
	text := "Hello, world!"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

func BenchmarkTokenize_MediumText(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. " +
		"This is a medium-sized text for benchmarking purposes."

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

func BenchmarkTokenize_LargeText(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = pretokenizer.Tokenize(text)
	}
}

// Note: State machine micro-benchmarks have been moved to internal/pretokenizer/benchmark_test.go

// =============================================================================
// Parallel and Concurrent Benchmarks
// =============================================================================

func BenchmarkEncodeParallel(b *testing.B) {
	tokenizer, err := New()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}

	texts := []string{
		"The quick brown fox jumps over the lazy dog.",
		"Hello, world! How are you doing today?",
		"Machine learning is fascinating.",
		"Natural language processing rocks!",
	}
	opts := &EncodeOptions{BOS: false, EOS: false}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			text := texts[i%len(texts)]
			_ = tokenizer.Encode(text, opts)
			i++
		}
	})
}

// =============================================================================
// Memory Allocation Tracking
// =============================================================================

func BenchmarkMemoryAllocations(b *testing.B) {
	testCases := []struct {
		name string
		text string
	}{
		{"empty", ""},
		{"single_word", "hello"},
		{"sentence", "The quick brown fox jumps over the lazy dog."},
		{"whitespace", "   spaces   tabs\t\t\tnewlines\n\n\n"},
		{"unicode", "Hello 🦙 world café résumé"},
		{"contractions", "I'm sure they're won't be late."},
		{"numbers", "123 456 789 1234 56789"},
		{"mixed", "Hello123! Test @email.com #hashtag"},
	}

	tokenizer, err := New()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}

	opts := &EncodeOptions{BOS: false, EOS: false}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tokenizer.Encode(tc.text, opts)
			}
		})
	}
}

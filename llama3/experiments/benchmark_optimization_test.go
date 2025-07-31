// +build experiments

package llama3

import (
	"testing"
	"strings"
)

// Benchmark comparison tests between original and optimized state machines

// ASCII-heavy text benchmarks
func BenchmarkStateMachineOriginal_ASCII(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. " +
		"This is a simple ASCII text with numbers 123 and punctuation! " +
		"We're testing contractions and various patterns."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

func BenchmarkStateMachineOptimized_ASCII(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. " +
		"This is a simple ASCII text with numbers 123 and punctuation! " +
		"We're testing contractions and various patterns."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeOptimized(text)
	}
}

// Unicode-heavy text benchmarks  
func BenchmarkStateMachineOriginal_Unicode(b *testing.B) {
	text := "こんにちは世界 🌍 Привет мир 🇷🇺 مرحبا بالعالم 🇸🇦 " +
		"Olá mundo 🇧🇷 你好世界 🇨🇳 नमस्ते दुनिया 🇮🇳"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

func BenchmarkStateMachineOptimized_Unicode(b *testing.B) {
	text := "こんにちは世界 🌍 Привет мир 🇷🇺 مرحبا بالعالم 🇸🇦 " +
		"Olá mundo 🇧🇷 你好世界 🇨🇳 नमस्ते दुनिया 🇮🇳"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeOptimized(text)
	}
}

// Whitespace-heavy benchmarks
func BenchmarkStateMachineOriginal_Whitespace(b *testing.B) {
	text := "   Multiple   spaces   between   words   			and	tabs			" +
		"


and
newlines


   with   trailing   spaces   "
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

func BenchmarkStateMachineOptimized_Whitespace(b *testing.B) {
	text := "   Multiple   spaces   between   words   			and	tabs			" +
		"


and
newlines


   with   trailing   spaces   "
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeOptimized(text)
	}
}

// Contraction-heavy benchmarks
func BenchmarkStateMachineOriginal_Contractions(b *testing.B) {
	text := "I'm sure they're won't believe we've said it's what we'll do. " +
		"He'd have thought I'd known you're right. She's certain they've won't."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

func BenchmarkStateMachineOptimized_Contractions(b *testing.B) {
	text := "I'm sure they're won't believe we've said it's what we'll do. " +
		"He'd have thought I'd known you're right. She's certain they've won't."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeOptimized(text)
	}
}

// Large text benchmarks
func BenchmarkStateMachineOriginal_Large(b *testing.B) {
	base := "The quick brown fox jumps over the lazy dog. "
	text := strings.Repeat(base, 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

func BenchmarkStateMachineOptimized_Large(b *testing.B) {
	base := "The quick brown fox jumps over the lazy dog. "
	text := strings.Repeat(base, 100)
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeOptimized(text)
	}
}

// Memory allocation benchmarks
func BenchmarkStateMachineOriginal_Allocs(b *testing.B) {
	texts := []string{
		"Simple ASCII text",
		"Unicode text 你好世界",
		"   Whitespace   heavy   ",
		"Contractions I'm we're",
		"Numbers 123 456 789",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(texts[i%len(texts)])
	}
}

func BenchmarkStateMachineOptimized_Allocs(b *testing.B) {
	texts := []string{
		"Simple ASCII text",
		"Unicode text 你好世界",
		"   Whitespace   heavy   ",
		"Contractions I'm we're",
		"Numbers 123 456 789",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeOptimized(texts[i%len(texts)])
	}
}

// Character classification benchmarks
func BenchmarkIsLetter_Original(b *testing.B) {
	runes := []rune("abcABC123!@#αβγ文字")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := runes[i%len(runes)]
		_ = isLetter(r)
	}
}

func BenchmarkIsLetter_Optimized(b *testing.B) {
	runes := []rune("abcABC123!@#αβγ文字")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := runes[i%len(runes)]
		_ = isLetterOptimized(r)
	}
}

// Jump table dispatch benchmark
func BenchmarkJumpTableDispatch(b *testing.B) {
	sm := &OptimizedStateMachine{
		tokens: make([]string, 0, 64),
	}
	sm.initJumpTable()
	
	// Test various first characters
	chars := []rune{' ', 'a', '1', '\'', '!', '
', 'ñ', '世'}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := chars[i%len(chars)]
		if c < 256 {
			_ = sm.jumpTable[c]
		}
	}
}

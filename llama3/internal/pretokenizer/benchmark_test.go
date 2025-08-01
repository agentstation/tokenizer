package pretokenizer

import (
	"testing"
)

// =============================================================================
// State Machine Component Benchmarks
// =============================================================================

func BenchmarkTryContraction(b *testing.B) {
	inputs := []string{
		"'s", "'t", "'re", "'ve", "'m", "'ll", "'d",
		"'S", "'T", "'RE", "'VE", "'M", "'LL", "'D", // uppercase
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		sm := &stateMachine{
			input:    []rune(input),
			position: 0,
			tokens:   make([]string, 0, 1),
		}
		_ = sm.tryContraction()
	}
}

func BenchmarkTryWordWithPrefix(b *testing.B) {
	inputs := []string{
		"hello", "world", "test", "benchmark",
		"!hello", "#world", "@test", "$benchmark", // with prefix
		"Hello", "WORLD", "Test", "BENCHMARK", // mixed case
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		sm := &stateMachine{
			input:    []rune(input),
			position: 0,
			tokens:   make([]string, 0, 1),
		}
		_ = sm.tryWordWithPrefix()
	}
}

func BenchmarkTryNumbers(b *testing.B) {
	inputs := []string{
		"1", "12", "123", "1234", // various lengths
		"456", "789", "000", "999",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		sm := &stateMachine{
			input:    []rune(input),
			position: 0,
			tokens:   make([]string, 0, 1),
		}
		_ = sm.tryNumbers()
	}
}

func BenchmarkTryWhitespace(b *testing.B) {
	inputs := []string{
		" ", "  ", "   ", "    ", "     ", // spaces
		"\t", "\t\t", "\t\t\t", // tabs
		"\n", "\r\n", "\n\n", // newlines
		"   word", "\t\ttab", // with following content
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		sm := &stateMachine{
			input:    []rune(input),
			position: 0,
			tokens:   make([]string, 0, 1),
		}
		_ = sm.tryWhitespace()
	}
}

func BenchmarkTryNewlineSequence(b *testing.B) {
	inputs := []string{
		"   \n", "\t\t\r\n", "  \n\n",
		"\n  ", "\r\n\t", "\n\n  ",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		sm := &stateMachine{
			input:    []rune(input),
			position: 0,
			tokens:   make([]string, 0, 1),
		}
		_ = sm.tryNewlineSequence()
	}
}

func BenchmarkTryPunctuationWithSpace(b *testing.B) {
	inputs := []string{
		"!", "?", ".", ",", ";", ":",
		" !", " ?", " .", // with leading space
		"!!", "??", "...", // repeated
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		sm := &stateMachine{
			input:    []rune(input),
			position: 0,
			tokens:   make([]string, 0, 1),
		}
		_ = sm.tryPunctuationWithSpace()
	}
}

// =============================================================================
// Character Classification Benchmarks
// =============================================================================

func BenchmarkIsLetter(b *testing.B) {
	runes := []rune("abcABC123!@#αβγ文字")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := runes[i%len(runes)]
		_ = isLetter(r)
	}
}

func BenchmarkIsNumber(b *testing.B) {
	runes := []rune("0123456789abcABC!@#")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := runes[i%len(runes)]
		_ = isNumber(r)
	}
}

func BenchmarkIsWhitespace(b *testing.B) {
	runes := []rune(" \t\n\rabc123")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r := runes[i%len(runes)]
		_ = isWhitespace(r)
	}
}
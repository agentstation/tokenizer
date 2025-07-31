// +build experiments

package llama3

import (
	"strings"
	"testing"
)

// Benchmark pattern-specific optimizations

func BenchmarkTokenize_PatternOpt_Mixed(b *testing.B) {
	text := "Email: user@example.com | Phone: +1-555-0123 | " +
		"Price: $99.99 (save 20%!) | URL: https://example.com/path?q=1 " +
		"Unicode café résumé naïve 文字 🦙 | Code: if (x > 0) { return true; }"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizePatternOpt(text)
	}
}

func BenchmarkTokenize_Original_Mixed(b *testing.B) {
	text := "Email: user@example.com | Phone: +1-555-0123 | " +
		"Price: $99.99 (save 20%!) | URL: https://example.com/path?q=1 " +
		"Unicode café résumé naïve 文字 🦙 | Code: if (x > 0) { return true; }"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

// Contraction-heavy text
func BenchmarkTokenize_PatternOpt_Contractions(b *testing.B) {
	text := "I'm sure they're won't believe we've said it's what we'll do. " +
		"He'd have thought I'd known you're right. She's certain they've won't. " +
		"CAN'T WON'T SHOULDN'T COULDN'T WOULDN'T"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizePatternOpt(text)
	}
}

func BenchmarkTokenize_Original_Contractions(b *testing.B) {
	text := "I'm sure they're won't believe we've said it's what we'll do. " +
		"He'd have thought I'd known you're right. She's certain they've won't. " +
		"CAN'T WON'T SHOULDN'T COULDN'T WOULDN'T"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

// Number-heavy text
func BenchmarkTokenize_PatternOpt_Numbers(b *testing.B) {
	text := "123 456 789 012 345 678 901 234 567 890 " +
		"1 2 3 4 5 6 7 8 9 0 11 22 33 44 55 66 77 88 99 " +
		"100 200 300 400 500 600 700 800 900"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizePatternOpt(text)
	}
}

func BenchmarkTokenize_Original_Numbers(b *testing.B) {
	text := "123 456 789 012 345 678 901 234 567 890 " +
		"1 2 3 4 5 6 7 8 9 0 11 22 33 44 55 66 77 88 99 " +
		"100 200 300 400 500 600 700 800 900"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

// Code-like text
func BenchmarkTokenize_CodeOpt_Code(b *testing.B) {
	text := `func tokenize(text string) []string {
	// Initialize state machine
	sm := NewStateMachine(text)
	tokens := make([]string, 0, len(text)/4)
	
	// Process input
	for sm.position < len(sm.input) {
		sm.matchNext()
	}
	
	return sm.tokens
}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeCodeOpt(text)
	}
}

func BenchmarkTokenize_Original_Code(b *testing.B) {
	text := `func tokenize(text string) []string {
	// Initialize state machine
	sm := NewStateMachine(text)
	tokens := make([]string, 0, len(text)/4)
	
	// Process input
	for sm.position < len(sm.input) {
		sm.matchNext()
	}
	
	return sm.tokens
}`
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

// Natural language text
func BenchmarkTokenize_NaturalOpt_Natural(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 10) +
		"She sells seashells by the seashore. " +
		"How much wood would a woodchuck chuck if a woodchuck could chuck wood?"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeNaturalOpt(text)
	}
}

func BenchmarkTokenize_Original_Natural(b *testing.B) {
	text := strings.Repeat("The quick brown fox jumps over the lazy dog. ", 10) +
		"She sells seashells by the seashore. " +
		"How much wood would a woodchuck chuck if a woodchuck could chuck wood?"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Tokenize(text)
	}
}

// Test correctness
func TestPatternOptCompatibility(t *testing.T) {
	testCases := []string{
		"",
		"Hello, world!",
		"I'm sure they're won't be late.",
		"123 456 789",
		"Email: test@example.com",
		"          grabbed",
		"Mixed 123 text with numbers and contractions I'm",
		"ALL CAPS TEXT WITH CONTRACTIONS WON'T DON'T",
	}
	
	for _, tc := range testCases {
		t.Run(tc, func(t *testing.T) {
			original := Tokenize(tc)
			patternOpt := TokenizePatternOpt(tc)
			
			if len(original) != len(patternOpt) {
				t.Errorf("Length mismatch: original=%d, patternOpt=%d", len(original), len(patternOpt))
				t.Errorf("Original: %v", original)
				t.Errorf("PatternOpt: %v", patternOpt)
			} else {
				for i := range original {
					if original[i] != patternOpt[i] {
						t.Errorf("Token mismatch at %d: %q vs %q", i, original[i], patternOpt[i])
					}
				}
			}
		})
	}
}

// Benchmark fast helper functions
func BenchmarkTryContractionFast(b *testing.B) {
	inputs := [][]rune{
		[]rune("'s"),
		[]rune("'t"),
		[]rune("'re"),
		[]rune("'ve"),
		[]rune("'ll"),
		[]rune("'RE"),
		[]rune("'x"), // not a contraction
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		input := inputs[i%len(inputs)]
		_, _ = tryContractionFast(input, 0)
	}
}

func BenchmarkTryContraction_Original(b *testing.B) {
	texts := []string{
		"'s", "'t", "'re", "'ve", "'ll", "'RE", "'x",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm := &StateMachine{
			input:    []rune(texts[i%len(texts)]),
			position: 0,
			tokens:   make([]string, 0, 1),
		}
		_ = sm.tryContraction()
	}
}

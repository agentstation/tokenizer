// +build experiments

package llama3

import (
	"reflect"
	"testing"
)

// TestOptimizedCompatibility ensures optimized state machine produces identical results
func TestOptimizedCompatibility(t *testing.T) {
	testCases := GenerateExtendedTestCases()
	
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			original := Tokenize(tc.Input)
			optimized := TokenizeOptimized(tc.Input)
			
			if !reflect.DeepEqual(original, optimized) {
				t.Errorf("Tokenization mismatch for %q
Original:  %v
Optimized: %v",
					tc.Input, original, optimized)
			}
		})
	}
}

// TestOptimizedEdgeCases tests specific edge cases for the optimized implementation
func TestOptimizedEdgeCases(t *testing.T) {
	cases := []struct {
		name  string
		input string
	}{
		{"Empty", ""},
		{"Single ASCII letter", "a"},
		{"Single Unicode letter", "α"},
		{"Single digit", "1"},
		{"Single space", " "},
		{"Single apostrophe", "'"},
		{"ASCII contraction", "I'm"},
		{"Unicode before contraction", "café's"},
		{"Multiple spaces with negative lookahead", "          grabbed"},
		{"Jump table boundary", string([]rune{255, 256, 257})},
		{"All ASCII punctuation", "!\"#$%&'()*+,-./:;<=>?@[\]^_`{|}~"},
		{"Mixed ASCII/Unicode", "Hello世界World"},
	}
	
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			original := Tokenize(tc.input)
			optimized := TokenizeOptimized(tc.input)
			
			if !reflect.DeepEqual(original, optimized) {
				t.Errorf("Mismatch for %q
Original:  %v
Optimized: %v",
					tc.input, original, optimized)
			}
		})
	}
}

// TestOptimizedConcurrency tests thread safety of optimized implementation
func TestOptimizedConcurrency(t *testing.T) {
	texts := []string{
		"Simple text",
		"Unicode 你好世界",
		"Contractions I'm we're",
		"   Whitespace   heavy   ",
		"Mixed content with numbers 123 and symbols !@#",
	}
	
	// Run multiple goroutines tokenizing concurrently
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			for j := 0; j < 100; j++ {
				text := texts[(id+j)%len(texts)]
				original := Tokenize(text)
				optimized := TokenizeOptimized(text)
				
				if !reflect.DeepEqual(original, optimized) {
					t.Errorf("Concurrent mismatch for %q", text)
				}
			}
			done <- true
		}(i)
	}
	
	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

// BenchmarkOptimizedMemoryReuse tests pool efficiency
func BenchmarkOptimizedMemoryReuse(b *testing.B) {
	texts := []string{
		"Short",
		"Medium length text with some words",
		"Long text that is significantly longer than the others to test different allocation patterns and memory reuse efficiency",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TokenizeOptimized(texts[i%len(texts)])
	}
}

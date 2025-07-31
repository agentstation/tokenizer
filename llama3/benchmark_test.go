package llama3

import (
	"strings"
	"testing"
)

// Benchmark individual components
func BenchmarkStateMachine(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog. This is a test sentence with multiple spaces   and some punctuation!"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sm := NewStateMachine(text)
		_ = sm.TokenizeWithStateMachine()
	}
}

func BenchmarkByteEncoding(b *testing.B) {
	text := "The quick brown fox jumps over the lazy dog."
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = encodeBytes([]byte(text))
	}
}

func BenchmarkVocabLookup(b *testing.B) {
	tokenizer, err := NewDefault()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}
	
	// Common words that should be in vocabulary
	words := []string{"the", "quick", "brown", "fox", "jumps"}
	encodedWords := make([]string, len(words))
	for i, w := range words {
		encodedWords[i] = encodeBytes([]byte(w))
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		word := encodedWords[i%len(encodedWords)]
		_ = tokenizer.vocabByString[word]
	}
}

func BenchmarkCacheLookup(b *testing.B) {
	tokenizer, err := NewDefault()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}
	
	// Warm up cache
	texts := []string{"hello", "world", "testing", "cache", "performance"}
	opts := &EncodeOptions{BOS: false, EOS: false}
	for _, text := range texts {
		tokenizer.Encode(text, opts)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := texts[i%len(texts)]
		tokenizer.cacheMutex.RLock()
		_ = tokenizer.cache[key]
		tokenizer.cacheMutex.RUnlock()
	}
}

// Benchmark different text types
func BenchmarkEncodeShortText(b *testing.B) {
	tokenizer, err := NewDefault()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}
	
	text := "Hello world!"
	opts := &EncodeOptions{BOS: false, EOS: false}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenizer.Encode(text, opts)
	}
}

func BenchmarkEncodeLongText(b *testing.B) {
	tokenizer, err := NewDefault()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}
	
	// Generate a longer text
	sentences := []string{
		"The quick brown fox jumps over the lazy dog.",
		"Machine learning models are becoming increasingly sophisticated.",
		"Natural language processing has many applications in modern technology.",
		"Tokenization is a fundamental step in text processing pipelines.",
	}
	text := strings.Join(sentences, " ")
	text = text + " " + text // Double it
	
	opts := &EncodeOptions{BOS: false, EOS: false}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenizer.Encode(text, opts)
	}
}

func BenchmarkEncodeUnicode(b *testing.B) {
	tokenizer, err := NewDefault()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}
	
	text := "Hello 世界! This is a test with émojis 🦙 and spëcial characters."
	opts := &EncodeOptions{BOS: false, EOS: false}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenizer.Encode(text, opts)
	}
}

func BenchmarkEncodeWhitespaceHeavy(b *testing.B) {
	tokenizer, err := NewDefault()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}
	
	text := "   Multiple   spaces   between   words   \t\t\tand\ttabs\t\t\t"
	opts := &EncodeOptions{BOS: false, EOS: false}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenizer.Encode(text, opts)
	}
}

// Benchmark parallel encoding
func BenchmarkEncodeParallel(b *testing.B) {
	tokenizer, err := NewDefault()
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

// Memory allocation benchmarks
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
	}
	
	tokenizer, err := NewDefault()
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

// Profile hot paths
func BenchmarkHotPath(b *testing.B) {
	tokenizer, err := NewDefault()
	if err != nil {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}
	
	// Most common case: regular English text
	text := "This is a typical sentence that might appear in normal text processing."
	opts := &EncodeOptions{BOS: false, EOS: false}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tokenizer.Encode(text, opts)
	}
}


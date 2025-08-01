package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"time"

	"github.com/agentstation/tokenizer/llama3"
)

var (
	cpuProfile = flag.String("cpuprofile", "", "write cpu profile to file")
	memProfile = flag.String("memprofile", "", "write memory profile to file")
	iterations = flag.Int("iterations", 10000, "number of iterations")
	textType   = flag.String("text", "mixed", "text type: ascii, unicode, whitespace, code, mixed, large")
)

func main() {
	flag.Parse()

	// Start CPU profiling if requested
	if *cpuProfile != "" {
		f, err := os.Create(*cpuProfile)
		if err != nil {
			log.Fatal("could not create CPU profile: ", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("Failed to close CPU profile file: %v", err)
			}
		}()
		if err := pprof.StartCPUProfile(f); err != nil {
			log.Fatal("could not start CPU profile: ", err)
		}
		defer pprof.StopCPUProfile()
	}

	// Initialize tokenizer
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal("failed to create tokenizer: ", err)
	}

	// Select test text
	text := getTestText(*textType)
	fmt.Printf("Testing with %s text (%d chars)\n", *textType, len(text))
	fmt.Printf("Running %d iterations...\n", *iterations)

	// Run tokenization
	start := time.Now()
	var totalTokens int
	for i := 0; i < *iterations; i++ {
		tokens := tokenizer.Encode(text, nil)
		totalTokens += len(tokens)
	}
	elapsed := time.Since(start)

	// Print results
	fmt.Printf("\nResults:\n")
	fmt.Printf("Total time: %v\n", elapsed)
	fmt.Printf("Time per iteration: %v\n", elapsed/time.Duration(*iterations))
	fmt.Printf("Tokens per iteration: %d\n", totalTokens / *iterations)
	fmt.Printf("Throughput: %.2f tokens/sec\n", float64(totalTokens)/elapsed.Seconds())

	// Write memory profile if requested
	if *memProfile != "" {
		f, err := os.Create(*memProfile)
		if err != nil {
			log.Fatal("could not create memory profile: ", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Printf("Failed to close memory profile file: %v", err)
			}
		}()
		runtime.GC() // get up-to-date statistics
		if err := pprof.WriteHeapProfile(f); err != nil {
			log.Fatal("could not write memory profile: ", err)
		}
	}

	// Print memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("\nMemory statistics:\n")
	fmt.Printf("Alloc = %v MB\n", m.Alloc/1024/1024)
	fmt.Printf("TotalAlloc = %v MB\n", m.TotalAlloc/1024/1024)
	fmt.Printf("Sys = %v MB\n", m.Sys/1024/1024)
	fmt.Printf("NumGC = %v\n", m.NumGC)
}

func getTestText(textType string) string {
	switch textType {
	case "ascii":
		return "The quick brown fox jumps over the lazy dog. " +
			"This is a simple ASCII text with numbers 123 and punctuation! " +
			"We're testing contractions and various patterns."

	case "unicode":
		return "Hello world! 你好世界！ Привет мир! مرحبا بالعالم! " +
			"🌍🌎🌏 Unicode test with emojis 🦙🐕🦊 and various scripts " +
			"αβγδε ΑΒΓΔΕ ¡¢£¤¥¦§¨©ª«¬\u00ad®¯°±²³´µ¶·¸¹º»¼½¾¿"

	case "whitespace":
		return "   Multiple   spaces   between   words   \t\t\tand\ttabs\t\t\t" +
			"\n\n\nand\nnewlines\n\n\n   with   trailing   spaces   " +
			"          ten spaces before grabbed           "

	case "code":
		return `func tokenize(text string) []string {
	// Initialize state machine
	sm := NewStateMachine(text)
	tokens := make([]string, 0, len(text)/4)
	
	// Process input
	for sm.position < len(sm.input) {
		sm.matchNext()
	}
	
	return sm.tokens
}`

	case "mixed":
		return "Email: user@example.com | Phone: +1-555-0123 | " +
			"Price: $99.99 (save 20%!) | URL: https://example.com/path?q=1 " +
			"Unicode café résumé naïve 文字 🦙 | Code: if (x > 0) { return true; }"

	case "large":
		// Generate a large text by repeating patterns
		base := "The quick brown fox jumps over the lazy dog. "
		return strings.Repeat(base, 100)

	default:
		return "Hello, world! This is a default test text."
	}
}

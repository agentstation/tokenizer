package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	var (
		output = flag.String("output", "test_vectors.jsonl", "Output file for test vectors")
		count  = flag.Int("count", 100, "Number of test vectors to generate")
	)
	flag.Parse()

	// Check if Node.js is available
	if _, err := exec.LookPath("node"); err != nil {
		log.Fatal("Node.js is required but not found in PATH")
	}

	// Check if JS tokenizer exists
	jsPath := filepath.Join(os.Getenv("HOME"), "src/github.com/belladoreai/llama3-tokenizer-js/bundle/llama3-tokenizer-with-baked-data.js")
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		log.Fatalf("JS tokenizer not found at %s", jsPath)
	}

	// Generate test inputs
	inputs := generateTestInputs(*count)

	// Create JS script
	jsContent := `import llama3Tokenizer from '` + jsPath + `';

const inputs = ` + toJSON(inputs) + `;

inputs.forEach(input => {
    const tokens = llama3Tokenizer.encode(input, {bos: false, eos: false});
    console.log(JSON.stringify({input, expected: tokens}));
});
`

	// Write to temporary file
	tmpFile := filepath.Join(os.TempDir(), "generate_vectors.js")
	if err := os.WriteFile(tmpFile, []byte(jsContent), 0600); err != nil {
		log.Fatalf("Failed to write JS file: %v", err)
	}
	defer func() {
		if err := os.Remove(tmpFile); err != nil {
			log.Printf("Failed to remove temporary file: %v", err)
		}
	}()

	// Run the script
	cmd := exec.Command("node", tmpFile) // #nosec G204 - tmpFile is safely constructed
	outputBytes, err := cmd.Output()
	if err != nil {
		log.Fatalf("Failed to run JS script: %v", err)
	}

	// Write output
	if err := os.WriteFile(*output, outputBytes, 0600); err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}

	fmt.Printf("Generated %d test vectors in %s\n", len(inputs), *output)
}

func generateTestInputs(count int) []string {
	inputs := []string{
		// Edge cases
		"",
		" ",
		"  ",
		"\t",
		"\n",
		"\r\n",

		// Basic text
		"Hello",
		"Hello world",
		"Hello, world!",
		"The quick brown fox jumps over the lazy dog.",

		// Whitespace patterns (critical for our state machine)
		"   leading spaces",
		"trailing spaces   ",
		"   both sides   ",
		"multiple   spaces   between",
		"\ttab\tcharacters\there\t",
		"\n\nnewlines\n\n",
		"mixed \t spaces \n and \r\n things",
		"          ten spaces then word",
		"           eleven spaces then word",
		"            twelve spaces then word",

		// Contractions
		"can't", "won't", "it's", "they're", "I've", "we'll", "he'd",
		"CAN'T", "WON'T", "IT'S", // uppercase
		"Can't", "Won't", "It's", // mixed case

		// Numbers
		"123",
		"1234", // more than 3 digits
		"1 2 3",
		"123 456 789",
		"12 345 6789", // mixed digit groups

		// Punctuation
		"Hello!",
		"What?",
		"Well...",
		"Hi!!!",
		"(parentheses)",
		"[brackets]",
		"{braces}",
		"<angles>",

		// Special patterns
		"email@example.com",
		"user.name+tag@example.co.uk",
		"https://example.com",
		"http://sub.example.com/path?query=value#fragment",
		"C:\\Windows\\System32",
		"/usr/local/bin/bash",
		"~/Documents/file.txt",

		// Unicode and emojis
		"café",
		"naïve",
		"résumé",
		"Zürich",
		"Москва",
		"北京",
		"日本語",
		"🦙",
		"👍",
		"🇺🇸",
		"Hello 🦙 world",
		"Multiple 🦙🦙🦙 llamas",

		// Mixed content
		"Temperature: -5°C",
		"Price: $99.99",
		"Save 50%!",
		"Call +1-800-555-0123",
		"#hashtag #another",
		"@user @mention",

		// Code-like content
		"function() { return true; }",
		"if (x > 0) { y = x * 2; }",
		"SELECT * FROM users WHERE id = 1;",
		"git commit -m 'Initial commit'",
		"npm install --save-dev",

		// Long text with various elements
		"The year is 2024, and AI is advancing rapidly! 🚀",
		"Order #12345 confirmed. Total: $199.99 (incl. 10% tax)",
		"Meeting scheduled for 3:30 PM EST on Jan 15, 2024.",
		"Error: File not found at C:\\Users\\Admin\\Documents\\data.csv",
	}

	// Add more if needed to reach count
	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog",
		"hello", "world", "testing", "tokenizer", "implementation", "llama", "model"}

	for len(inputs) < count {
		// Generate pseudo-random sentences
		sentLen := 5 + (len(inputs) % 10)
		sent := ""
		for j := 0; j < sentLen; j++ {
			if j > 0 {
				sent += " "
			}
			sent += words[(len(inputs)*j)%len(words)]
		}
		inputs = append(inputs, sent+".")
	}

	return inputs[:count]
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

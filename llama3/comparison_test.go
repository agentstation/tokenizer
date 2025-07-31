package llama3

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ComparisonTestCase represents a test case for comparing Go and JS implementations
type ComparisonTestCase struct {
	Input    string `json:"input"`
	Expected []int  `json:"expected"`
}

// toJSON converts a value to JSON string
func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// generateTestVectors creates a JavaScript file to generate test vectors
func generateTestVectors(testCases []string) ([]ComparisonTestCase, error) {
	// Create temporary JS file
	tmpDir := os.TempDir()
	jsFile := filepath.Join(tmpDir, "generate_vectors.js")
	
	jsContent := `import llama3Tokenizer from '` + filepath.Join(os.Getenv("HOME"), "src/github.com/belladoreai/llama3-tokenizer-js/bundle/llama3-tokenizer-with-baked-data.js") + `';

const testCases = ` + toJSON(testCases) + `;

const results = testCases.map(input => ({
    input: input,
    expected: llama3Tokenizer.encode(input, {bos: false, eos: false})
}));

console.log(JSON.stringify(results, null, 2));
`
	
	if err := os.WriteFile(jsFile, []byte(jsContent), 0644); err != nil {
		return nil, fmt.Errorf("failed to write JS file: %w", err)
	}
	defer os.Remove(jsFile)
	
	// Run the JS file
	cmd := exec.Command("node", jsFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to run JS file: %w\nOutput: %s", err, output)
	}
	
	// Parse the output
	var results []ComparisonTestCase
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse JS output: %w", err)
	}
	
	return results, nil
}

// TestComparisonWithJS compares Go implementation with JavaScript implementation
func TestComparisonWithJS(t *testing.T) {
	tokenizer, err := NewDefault()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}
	
	// Check if Node.js is available
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Skipping comparison tests: Node.js not found")
	}
	
	// Check if JS tokenizer exists
	jsPath := filepath.Join(os.Getenv("HOME"), "src/github.com/belladoreai/llama3-tokenizer-js/bundle/llama3-tokenizer-with-baked-data.js")
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		t.Skip("Skipping comparison tests: JS tokenizer not found")
	}
	
	// Define comprehensive test cases
	testInputs := []string{
		// Basic text
		"Hello world",
		"Hello, world!",
		"The quick brown fox jumps over the lazy dog.",
		
		// Whitespace variations
		"  spaces  ",
		"\ttabs\t",
		"\n\nnewlines\n\n",
		"   multiple   spaces   between   words   ",
		"\t\t\tmultiple\ttabs\t\t\t",
		"mixed \t spaces \n and \r\n newlines",
		
		// Contractions
		"can't",
		"won't",
		"it's",
		"they're",
		"I've",
		"we'll",
		"he'd",
		
		// Numbers
		"123",
		"1 2 3",
		"123 456 789",
		"3.14159",
		"1,000,000",
		
		// Punctuation
		"Hello!",
		"What?",
		"Well...",
		"(parentheses)",
		"[brackets]",
		"{braces}",
		"quote: \"hello\"",
		"'single quotes'",
		
		// Special characters
		"email@example.com",
		"https://example.com",
		"C:\\Windows\\System32",
		"/usr/local/bin",
		"$100.00",
		"50%",
		"#hashtag",
		"@mention",
		
		// Unicode
		"café",
		"naïve",
		"résumé",
		"Москва",
		"北京",
		"🦙",
		"👍🏽",
		"🇺🇸",
		
		// Mixed cases
		"CamelCase",
		"snake_case",
		"kebab-case",
		"SCREAMING_SNAKE_CASE",
		
		// Edge cases
		"",
		" ",
		"\t",
		"\n",
		".",
		"!",
		"?",
		
		// Complex sentences
		"The year is 2024, and AI is advancing rapidly!",
		"Temperature: -5°C (23°F)",
		"Price: $99.99 (was $149.99 - save 33%!)",
		"Email support@example.com or call +1-800-555-0123",
	}
	
	// Generate expected outputs from JavaScript
	testCases, err := generateTestVectors(testInputs)
	if err != nil {
		t.Fatalf("Failed to generate test vectors: %v", err)
	}
	
	opts := &EncodeOptions{BOS: false, EOS: false}
	failures := 0
	
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input=%q", tc.Input), func(t *testing.T) {
			// Encode with Go implementation
			got := tokenizer.Encode(tc.Input, opts)
			
			// Compare
			if len(got) != len(tc.Expected) {
				t.Errorf("Length mismatch: got %d tokens, expected %d tokens", len(got), len(tc.Expected))
				t.Logf("Got:      %v", got)
				t.Logf("Expected: %v", tc.Expected)
				failures++
				return
			}
			
			for i := range got {
				if got[i] != tc.Expected[i] {
					t.Errorf("Token mismatch at position %d: got %d, expected %d", i, got[i], tc.Expected[i])
					t.Logf("Got:      %v", got)
					t.Logf("Expected: %v", tc.Expected)
					
					// Decode to show what the tokens represent
					gotDecoded := tokenizer.Decode([]int{got[i]})
					expectedDecoded := tokenizer.Decode([]int{tc.Expected[i]})
					t.Logf("Token %d decodes to %q, token %d decodes to %q", got[i], gotDecoded, tc.Expected[i], expectedDecoded)
					failures++
					return
				}
			}
		})
	}
	
	if failures > 0 {
		t.Errorf("Total failures: %d out of %d test cases", failures, len(testCases))
	}
}

// TestComparisonFromFile tests using test vectors from a file
func TestComparisonFromFile(t *testing.T) {
	tokenizer, err := NewDefault()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}
	
	// Look for test vector file
	vectorFile := "test_vectors.jsonl"
	if _, err := os.Stat(vectorFile); os.IsNotExist(err) {
		t.Skip("Skipping file-based comparison: test_vectors.jsonl not found")
	}
	
	file, err := os.Open(vectorFile)
	if err != nil {
		t.Fatalf("Failed to open test vectors file: %v", err)
	}
	defer file.Close()
	
	scanner := bufio.NewScanner(file)
	opts := &EncodeOptions{BOS: false, EOS: false}
	lineNum := 0
	failures := 0
	
	for scanner.Scan() {
		lineNum++
		var tc ComparisonTestCase
		if err := json.Unmarshal(scanner.Bytes(), &tc); err != nil {
			t.Errorf("Line %d: Failed to parse JSON: %v", lineNum, err)
			continue
		}
		
		got := tokenizer.Encode(tc.Input, opts)
		
		if len(got) != len(tc.Expected) {
			t.Errorf("Line %d: Length mismatch for %q: got %d, expected %d", 
				lineNum, tc.Input, len(got), len(tc.Expected))
			failures++
			continue
		}
		
		for i := range got {
			if got[i] != tc.Expected[i] {
				t.Errorf("Line %d: Token mismatch at position %d for %q: got %d, expected %d",
					lineNum, i, tc.Input, got[i], tc.Expected[i])
				failures++
				break
			}
		}
	}
	
	if err := scanner.Err(); err != nil {
		t.Fatalf("Error reading file: %v", err)
	}
	
	if failures > 0 {
		t.Errorf("Total failures: %d out of %d test cases", failures, lineNum)
	} else {
		t.Logf("All %d test cases passed!", lineNum)
	}
}

// generateTestVectorFile creates a file with test vectors for future use
func generateTestVectorFile(filename string, count int) error {
	// Generate diverse test inputs
	var inputs []string
	
	// Add specific test cases
	specificCases := []string{
		"Hello world",
		"           grabbed",
		"\ttabs\t\t\t\tout here",
		"The quick brown fox jumps over the lazy dog.",
		"🦙 Llama emoji",
		"Mixed UTF-8: café, naïve, résumé",
		"Numbers: 123, 3.14, 1,000,000",
		"Punctuation! Really? Yes... (maybe)",
		"https://example.com/path?query=value#fragment",
		"user@example.com",
	}
	inputs = append(inputs, specificCases...)
	
	// Add random sentences
	words := []string{"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog",
		"hello", "world", "testing", "tokenizer", "implementation", "llama", "model",
		"artificial", "intelligence", "machine", "learning", "natural", "language"}
	
	for i := len(inputs); i < count; i++ {
		// Create random sentence
		sentLen := 5 + (i % 10)
		sent := make([]string, sentLen)
		for j := 0; j < sentLen; j++ {
			sent[j] = words[(i*j)%len(words)]
		}
		inputs = append(inputs, strings.Join(sent, " ")+".")
	}
	
	// Generate test vectors
	vectors, err := generateTestVectors(inputs)
	if err != nil {
		return err
	}
	
	// Write to file
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	
	encoder := json.NewEncoder(file)
	for _, v := range vectors {
		if err := encoder.Encode(v); err != nil {
			return err
		}
	}
	
	return nil
}
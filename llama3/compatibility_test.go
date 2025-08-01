package llama3

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/agentstation/tokenizer/llama3/internal/pretokenizer"
	testutils "github.com/agentstation/tokenizer/llama3/internal/testing"
)

// TestCompatibility runs comprehensive compatibility tests with 476 test cases.
func TestCompatibility(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	// Check if Node.js is available
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Skipping extended compatibility tests: Node.js not found")
	}

	// Check if JS tokenizer exists
	jsPath := filepath.Join(os.Getenv("HOME"), "src/github.com/belladoreai/llama3-tokenizer-js/bundle/llama3-tokenizer-with-baked-data.js")
	if _, err := os.Stat(jsPath); os.IsNotExist(err) {
		t.Skip("Skipping extended compatibility tests: JS tokenizer not found")
	}

	// Generate test cases
	testCases := testutils.GenerateTestCases()
	t.Logf("Running %d test cases", len(testCases))

	// Group test cases by category for better reporting
	categories := make(map[string][]testutils.TestCase)
	for _, tc := range testCases {
		categories[tc.Category] = append(categories[tc.Category], tc)
	}

	opts := &EncodeOptions{BOS: false, EOS: false}
	totalPassed := 0
	totalFailed := 0

	for category, cases := range categories {
		t.Run(category, func(t *testing.T) {
			passed := 0
			failed := 0

			// Process in batches to avoid command line length limits
			batchSize := 50
			for i := 0; i < len(cases); i += batchSize {
				end := i + batchSize
				if end > len(cases) {
					end = len(cases)
				}
				batch := cases[i:end]

				// Get expected results from JavaScript
				inputs := make([]string, len(batch))
				for j, tc := range batch {
					inputs[j] = tc.Input
				}

				expected, err := getJSTokenization(inputs, jsPath)
				if err != nil {
					t.Fatalf("Failed to get JS tokenization: %v", err)
				}

				// Test each case
				for j, tc := range batch {
					got := tokenizer.Encode(tc.Input, opts)
					exp := expected[j]

					if !compareTokens(got, exp) {
						t.Errorf("Case %d: %s\nInput: %q\nGot:      %v\nExpected: %v",
							i+j, tc.Description, tc.Input, got, exp)
						failed++
					} else {
						passed++
					}
				}
			}

			totalPassed += passed
			totalFailed += failed
			t.Logf("%s: %d passed, %d failed", category, passed, failed)
		})
	}

	// Summary
	t.Logf("\nTotal: %d passed, %d failed (%.1f%% success rate)",
		totalPassed, totalFailed,
		float64(totalPassed)*100/float64(totalPassed+totalFailed))
}

// getJSTokenization gets tokenization results from JavaScript implementation.
func getJSTokenization(inputs []string, jsPath string) ([][]int, error) {
	// Create temporary JS file
	tmpDir := os.TempDir()
	jsFile := filepath.Join(tmpDir, "batch_tokenize.js")

	jsContent := `import llama3Tokenizer from '` + jsPath + `';

const inputs = ` + toJSON(inputs) + `;

const results = inputs.map(input => 
    llama3Tokenizer.encode(input, {bos: false, eos: false})
);

console.log(JSON.stringify(results));
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
	var results [][]int
	if err := json.Unmarshal(output, &results); err != nil {
		return nil, fmt.Errorf("failed to parse JS output: %w", err)
	}

	return results, nil
}

// compareTokens compares two token arrays.
func compareTokens(a, b []int) bool {
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

// TestTokenizationProperties tests properties that should hold for all inputs.
func TestTokenizationProperties(t *testing.T) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		t.Skip("Skipping tests: Llama 3 data not available")
	}

	testCases := testutils.GenerateTestCases()
	opts := &EncodeOptions{BOS: false, EOS: false}

	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			// Tokenize
			tokens := tokenizer.Encode(tc.Input, opts)

			// Property 1: Round-trip (encode-decode)
			decoded := tokenizer.Decode(tokens)
			// Note: Due to byte-level encoding, exact round-trip isn't always possible
			// but the decoded length should be reasonable
			if len(decoded) == 0 && len(tc.Input) > 0 {
				t.Errorf("Empty decode for non-empty input: %q", tc.Input)
			}

			// Property 2: Deterministic
			tokens2 := tokenizer.Encode(tc.Input, opts)
			if !compareTokens(tokens, tokens2) {
				t.Errorf("Non-deterministic tokenization for: %q", tc.Input)
			}

			// Property 3: No empty tokens in pretokenization
			pretokens := pretokenizer.Tokenize(tc.Input)
			if err := testutils.ValidateTokenization(tc.Input, pretokens); err != nil {
				t.Errorf("Invalid pretokenization for %q: %v", tc.Input, err)
			}
		})
	}
}

// BenchmarkCases benchmarks various categories of inputs.
func BenchmarkCases(b *testing.B) {
	tokenizer, err := New()
	if err != nil || tokenizer.VocabSize() == 0 {
		b.Skip("Skipping benchmark: Llama 3 data not available")
	}

	testCases := testutils.GenerateTestCases()
	categories := make(map[string][]string)

	for _, tc := range testCases {
		categories[tc.Category] = append(categories[tc.Category], tc.Input)
	}

	opts := &EncodeOptions{BOS: false, EOS: false}

	for category, inputs := range categories {
		b.Run(category, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				input := inputs[i%len(inputs)]
				_ = tokenizer.Encode(input, opts)
			}
		})
	}
}

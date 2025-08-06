package llama3cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/agentstation/tokenizer/llama3"
)

var (
	// Encode command flags.
	encAddBOS    bool
	encAddEOS    bool
	encOutput    string
	encCount     bool
	encCountOnly bool
	encMetrics   bool
)

// newEncodeCmd creates the encode subcommand.
func newEncodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encode [text]",
		Short: "Encode text to token IDs",
		Long: `Encode text into Llama 3 token IDs.

If no text is provided as an argument, reads from stdin.
By default, adds beginning-of-sequence (BOS) and end-of-sequence (EOS) tokens.

The output format can be:
  - space: Space-separated token IDs (default)
  - newline: One token ID per line
  - json: JSON array of token IDs`,
		Example: `  # Encode a simple string
  tokenizer llama3 encode "Hello, world!"
  
  # Encode from stdin
  echo "Hello, world!" | tokenizer llama3 encode
  
  # Encode without special tokens
  tokenizer llama3 encode --no-bos --no-eos "Raw text"
  
  # Output as JSON
  tokenizer llama3 encode --output json "Hello"
  
  # Output one token per line
  tokenizer llama3 encode --output newline "Hello"
  
  # Show token count with output
  tokenizer llama3 encode --count "Hello"
  
  # Show only the token count
  tokenizer llama3 encode --count-only "Hello"`,
		RunE: runEncode,
	}

	// Add flags
	cmd.Flags().BoolVar(&encAddBOS, "bos", true, "Add beginning of sequence token")
	cmd.Flags().BoolVar(&encAddEOS, "eos", true, "Add end of sequence token")
	cmd.Flags().StringVarP(&encOutput, "output", "o", "space", "Output format: space, newline, json")
	cmd.Flags().BoolVar(&encCount, "count", false, "Show token count with output")
	cmd.Flags().BoolVar(&encCountOnly, "count-only", false, "Show only token count (no tokens)")
	cmd.Flags().BoolVar(&encMetrics, "metrics", false, "Show performance metrics")

	return cmd
}

func runEncode(_ *cobra.Command, args []string) error {
	var startTime time.Time
	if encMetrics {
		startTime = time.Now()
	}

	// Initialize tokenizer
	tokenizer, err := llama3.New()
	if err != nil {
		return fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	// Create reader based on input source
	var reader io.Reader
	var inputBytes int

	if len(args) > 0 {
		text := strings.Join(args, " ")
		inputBytes = len(text)
		reader = strings.NewReader(text)
	} else {
		// For stdin, wrap with counting reader if metrics enabled
		if encMetrics {
			cr := &countingReader{Reader: os.Stdin}
			reader = cr
			defer func() { inputBytes = cr.bytesRead }()
		} else {
			reader = os.Stdin
		}
	}

	// Create scanner with options
	scanner := tokenizer.NewScanner(
		reader,
		llama3.WithEncodeOptions(&llama3.EncodeOptions{
			BOS: encAddBOS,
			EOS: encAddEOS,
		}),
	)

	// Collect all tokens (needed for JSON output and count)
	var tokens []int
	for scanner.Scan() {
		tokens = append(tokens, scanner.Token())
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("tokenization error: %w", err)
	}

	// Get byte count if we used counting reader
	if cr, ok := reader.(*countingReader); ok {
		inputBytes = cr.bytesRead
	}

	var encodeDuration time.Duration
	if encMetrics {
		encodeDuration = time.Since(startTime)
	}

	// Handle count-only mode
	if encCountOnly {
		switch encOutput {
		case "json":
			data, err := json.Marshal(map[string]int{"count": len(tokens)})
			if err != nil {
				return fmt.Errorf("failed to marshal count: %w", err)
			}
			fmt.Println(string(data))
		default:
			fmt.Println(len(tokens))
		}
		return nil
	}

	// Output tokens with optional count and metrics
	switch encOutput {
	case "json":
		output := map[string]interface{}{
			"tokens": tokens,
		}
		if encCount {
			output["count"] = len(tokens)
		}
		if encMetrics {
			metrics := map[string]interface{}{
				"latency":     formatLatency(encodeDuration),
				"tps":         calculateTPS(len(tokens), encodeDuration),
				"input_bytes": inputBytes,
			}
			output["metrics"] = metrics
		}
		data, err := json.Marshal(output)
		if err != nil {
			return fmt.Errorf("failed to marshal output: %w", err)
		}
		fmt.Println(string(data))
	case "newline":
		if encCount {
			fmt.Printf("count: %d\n", len(tokens))
		}
		for _, token := range tokens {
			fmt.Println(token)
		}
		if encMetrics {
			fmt.Println("metrics:")
			fmt.Printf("  latency: %s\n", formatLatency(encodeDuration))
			fmt.Printf("  tps: %d\n", calculateTPS(len(tokens), encodeDuration))
			fmt.Printf("  input_bytes: %d\n", inputBytes)
		}
	case "space":
		if encCount {
			fmt.Printf("count: %d\n", len(tokens))
			fmt.Print("tokens: ")
		}
		for i, token := range tokens {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(token)
		}
		fmt.Println()
		if encMetrics {
			fmt.Println("metrics:")
			fmt.Printf("  latency: %s\n", formatLatency(encodeDuration))
			fmt.Printf("  tps: %d\n", calculateTPS(len(tokens), encodeDuration))
			fmt.Printf("  input_bytes: %d\n", inputBytes)
		}
	default:
		return fmt.Errorf("unknown output format: %s", encOutput)
	}

	return nil
}

// countingReader wraps an io.Reader to count bytes read.
type countingReader struct {
	io.Reader
	bytesRead int
}

func (cr *countingReader) Read(p []byte) (n int, err error) {
	n, err = cr.Reader.Read(p)
	cr.bytesRead += n
	return
}

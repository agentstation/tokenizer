package llama3cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/agentstation/tokenizer/llama3"
	"github.com/spf13/cobra"
)

var (
	// Encode command flags
	encAddBOS bool
	encAddEOS bool
	encOutput string
)

// newEncodeCmd creates the encode subcommand
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
  tokenizer llama3 encode --output newline "Hello"`,
		RunE: runEncode,
	}

	// Add flags
	cmd.Flags().BoolVar(&encAddBOS, "bos", true, "Add beginning of sequence token")
	cmd.Flags().BoolVar(&encAddEOS, "eos", true, "Add end of sequence token")
	cmd.Flags().StringVarP(&encOutput, "output", "o", "space", "Output format: space, newline, json")

	return cmd
}

func runEncode(cmd *cobra.Command, args []string) error {

	// Initialize tokenizer
	tokenizer, err := llama3.New()
	if err != nil {
		return fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	// Get text to encode
	var text string
	if len(args) > 0 {
		text = strings.Join(args, " ")
	} else {
		// Read from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		text = string(data)
	}

	// Encode with options
	opts := &llama3.EncodeOptions{
		BOS: encAddBOS,
		EOS: encAddEOS,
	}
	tokens := tokenizer.Encode(text, opts)

	// Output tokens
	switch encOutput {
	case "json":
		data, err := json.Marshal(tokens)
		if err != nil {
			return fmt.Errorf("failed to marshal tokens: %w", err)
		}
		fmt.Println(string(data))
	case "newline":
		for _, token := range tokens {
			fmt.Println(token)
		}
	case "space":
		for i, token := range tokens {
			if i > 0 {
				fmt.Print(" ")
			}
			fmt.Print(token)
		}
		fmt.Println()
	default:
		return fmt.Errorf("unknown output format: %s", encOutput)
	}

	return nil
}

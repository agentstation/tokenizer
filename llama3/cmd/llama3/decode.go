package llama3cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/agentstation/tokenizer/llama3"
)

var (
	// Decode command flags.
	decSkipSpecial bool
)

// newDecodeCmd creates the decode subcommand.
func newDecodeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decode [token_ids...]",
		Short: "Decode token IDs to text",
		Long: `Decode Llama 3 token IDs back to text.

Token IDs can be provided as arguments or piped from stdin.
Multiple token IDs should be separated by spaces when provided as arguments,
or by any whitespace (spaces, tabs, newlines) when reading from stdin.

Special tokens (like <|begin_of_text|>) are decoded by default but can be
skipped using the --skip-special flag.`,
		Example: `  # Decode token IDs from arguments
  tokenizer llama3 decode 1234 5678 9012
  
  # Decode from stdin
  echo "1234 5678 9012" | tokenizer llama3 decode
  
  # Decode from encode output
  tokenizer llama3 encode "test" | tokenizer llama3 decode
  
  # Skip special tokens in output
  tokenizer llama3 decode --skip-special 128000 1234 128001`,
		RunE: runDecode,
	}

	// Add flags
	cmd.Flags().BoolVar(&decSkipSpecial, "skip-special", false, "Skip special tokens in output")

	return cmd
}

func runDecode(_ *cobra.Command, args []string) error {
	// Initialize tokenizer
	tokenizer, err := llama3.New()
	if err != nil {
		return fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	// Get token IDs
	var tokens []int
	if len(args) > 0 {
		// Parse from arguments
		for _, arg := range args {
			token, err := strconv.Atoi(arg)
			if err != nil {
				return fmt.Errorf("invalid token ID %q: %w", arg, err)
			}
			tokens = append(tokens, token)
		}
	} else {
		// Read from stdin
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			token, err := strconv.Atoi(scanner.Text())
			if err != nil {
				return fmt.Errorf("invalid token ID %q: %w", scanner.Text(), err)
			}
			tokens = append(tokens, token)
		}
		if err := scanner.Err(); err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
	}

	if len(tokens) == 0 {
		return fmt.Errorf("no token IDs provided")
	}

	// Decode tokens
	text := tokenizer.Decode(tokens)

	// Skip special tokens if requested
	if decSkipSpecial {
		// Remove common special tokens
		// This is a simple implementation - could be improved by
		// actually checking which tokens are special tokens
		specialTokens := []string{
			"<|begin_of_text|>",
			"<|end_of_text|>",
			"<|start_header_id|>",
			"<|end_header_id|>",
			"<|eot_id|>",
			"<|eom_id|>",
			"<|python_tag|>",
			"<|finetune_right_pad_id|>",
		}

		// Also remove reserved special tokens
		for i := 0; i < 256; i++ {
			specialTokens = append(specialTokens, fmt.Sprintf("<|reserved_special_token_%d|>", i))
		}

		for _, special := range specialTokens {
			text = strings.ReplaceAll(text, special, "")
		}
	}

	fmt.Print(text)
	return nil
}

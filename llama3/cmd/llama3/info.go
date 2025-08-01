package llama3cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/agentstation/tokenizer/llama3"
)

// newInfoCmd creates the info subcommand.
func newInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Display tokenizer information",
		Long: `Display information about the Llama 3 tokenizer including vocabulary size,
special tokens, and other relevant details.

This command is useful for understanding the tokenizer's capabilities and
configuration.`,
		Example: `  # Show tokenizer information
  tokenizer llama3 info`,
		RunE: runInfo,
	}

	return cmd
}

func runInfo(_ *cobra.Command, _ []string) error {
	// Initialize tokenizer
	tokenizer, err := llama3.New()
	if err != nil {
		return fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	fmt.Println("Llama 3 Tokenizer Information")
	fmt.Println("=============================")
	fmt.Println()

	// Basic information
	fmt.Println("Model Details:")
	fmt.Printf("  Model Type:        Llama 3 (Meta)\n")
	fmt.Printf("  Tokenizer Type:    Byte-level BPE\n")
	fmt.Printf("  Vocabulary Size:   %d tokens\n", tokenizer.VocabSize())
	fmt.Printf("  Regular Tokens:    %d\n", 128000)
	fmt.Printf("  Special Tokens:    %d\n", 256)
	fmt.Println()

	// Special token examples
	fmt.Println("Special Token Examples:")
	specialTokens := []struct {
		name  string
		token string
	}{
		{"Begin of Text", "<|begin_of_text|>"},
		{"End of Text", "<|end_of_text|>"},
		{"Start Header ID", "<|start_header_id|>"},
		{"End Header ID", "<|end_header_id|>"},
		{"End of Turn ID", "<|eot_id|>"},
		{"End of Message ID", "<|eom_id|>"},
		{"Python Tag", "<|python_tag|>"},
		{"Finetune Pad", "<|finetune_right_pad_id|>"},
		{"Reserved 0", "<|reserved_special_token_0|>"},
		{"Reserved 1", "<|reserved_special_token_1|>"},
		{"Reserved 2", "<|reserved_special_token_2|>"},
	}

	for _, st := range specialTokens {
		if id, err := tokenizer.GetSpecialTokenID(st.token); err == nil {
			fmt.Printf("  %-18s %-30s -> %d\n", st.name+":", st.token, id)
		}
	}

	fmt.Println()
	fmt.Printf("  ... and %d more reserved special tokens\n", 245)
	fmt.Println()

	// Encoding characteristics
	fmt.Println("Encoding Characteristics:")
	fmt.Printf("  Byte-level:        Yes (handles any byte sequence)\n")
	fmt.Printf("  Whitespace:        Preserved (including multiple spaces)\n")
	fmt.Printf("  Case Sensitive:    Yes\n")
	fmt.Printf("  Unicode Support:   Full (via byte encoding)\n")
	fmt.Println()

	// Performance features
	fmt.Println("Performance Features:")
	fmt.Printf("  BPE Cache:         Enabled (LRU cache)\n")
	fmt.Printf("  Streaming:         Supported (via Scanner interface)\n")
	fmt.Printf("  Thread Safe:       Yes (with proper usage)\n")

	return nil
}

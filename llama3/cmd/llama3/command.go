// Package llama3cmd provides the llama3 command for the tokenizer CLI.
package llama3cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// Command returns the llama3 command tree for the tokenizer CLI.
// This command provides encode, decode, stream, and info subcommands
// for working with the Llama 3 tokenizer.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "llama3",
		Short: "Llama 3 tokenizer operations",
		Long: `Perform tokenization operations using Meta's Llama 3 tokenizer.

The Llama 3 tokenizer uses byte-level BPE (Byte Pair Encoding) with a
vocabulary of 128,256 tokens (128,000 regular tokens + 256 special tokens).

Available commands:
  encode - Encode text to token IDs (default when text is provided)
  decode - Decode token IDs to text
  stream - Process text in streaming mode
  info   - Display tokenizer information`,
		Example: `  # Encode text (explicit)
  tokenizer llama3 encode "Hello, world!"
  
  # Encode text (implicit - default action)
  tokenizer llama3 "Hello, world!"
  
  # Decode tokens
  tokenizer llama3 decode 128000 9906 11 1917 0 128001
  
  # Stream from stdin (explicit)
  cat large_file.txt | tokenizer llama3 stream
  
  # Stream from stdin (implicit - automatic)
  cat large_file.txt | tokenizer llama3
  
  # Show tokenizer info
  tokenizer llama3 info`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			// If args are provided and first arg is not a known subcommand,
			// treat it as text to encode
			if len(args) > 0 {
				// Check if first arg is a subcommand or "help"
				firstArg := args[0]
				if firstArg == "help" {
					return cmd.Help()
				}

				// Check if it's a known subcommand
				for _, subcmd := range cmd.Commands() {
					if subcmd.Name() == firstArg || subcmd.HasAlias(firstArg) {
						// It's a subcommand, let normal command handling take over
						return cmd.Usage()
					}
				}

				// Not a subcommand, treat as text to encode
				encodeCmd := newEncodeCmd()
				encodeCmd.SetArgs(args)
				return encodeCmd.Execute()
			}

			// No args provided - check if stdin is piped
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				// Data is being piped to stdin, use streaming mode
				streamCmd := newStreamCmd()
				return streamCmd.Execute()
			}

			// No args and no piped input, show usage
			return cmd.Usage()
		},
	}

	// Add subcommands
	cmd.AddCommand(
		newEncodeCmd(),
		newDecodeCmd(),
		newStreamCmd(),
		newInfoCmd(),
	)

	return cmd
}

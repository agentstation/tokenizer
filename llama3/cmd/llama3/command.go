// Package llama3cmd provides the llama3 command for the tokenizer CLI.
package llama3cmd

import (
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
  encode - Encode text to token IDs
  decode - Decode token IDs to text
  stream - Process text in streaming mode
  info   - Display tokenizer information`,
		Example: `  # Encode text
  tokenizer llama3 encode "Hello, world!"
  
  # Decode tokens
  tokenizer llama3 decode 128000 9906 11 1917 0 128001
  
  # Stream from stdin
  cat large_file.txt | tokenizer llama3 stream
  
  # Show tokenizer info
  tokenizer llama3 info`,
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
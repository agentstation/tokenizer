package main

import (
	"fmt"

	"github.com/spf13/cobra"

	llama3cmd "github.com/agentstation/tokenizer/llama3/cmd/llama3"
)

// rootCmd represents the base command when called without any subcommands.
var rootCmd = &cobra.Command{
	Use:   "tokenizer",
	Short: "A multi-model tokenizer CLI tool",
	Long: `Tokenizer is a CLI tool for tokenizing text using various language models.

This tool provides a unified interface for working with different tokenizer
implementations. Each tokenizer is available as a subcommand with its own
set of operations.

Currently supported tokenizers:
  - llama3: Meta's Llama 3 tokenizer (128k vocabulary, byte-level BPE)

Common operations available for tokenizers:
  - encode: Convert text to token IDs
  - decode: Convert token IDs back to text
  - stream: Process large files in streaming mode
  - info:   Display tokenizer information`,
	Example: `  # Encode text with Llama 3
  tokenizer llama3 encode "Hello, world!"
  
  # Decode tokens
  tokenizer llama3 decode 1234 5678
  
  # Stream a large file
  cat large_file.txt | tokenizer llama3 stream
  
  # Get tokenizer info
  tokenizer llama3 info`,
	SilenceUsage: true,
}

// versionCmd represents the version command.
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("tokenizer version %s\n", version)
		if commit != "none" {
			fmt.Printf("  commit:     %s\n", commit)
		}
		if buildDate != "unknown" {
			fmt.Printf("  built:      %s\n", buildDate)
		}
		if goVersion != "unknown" {
			fmt.Printf("  go version: %s\n", goVersion)
		}
	},
}

func init() {
	// Register commands
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(llama3cmd.Command())

	// Future tokenizers can be added here:
	// rootCmd.AddCommand(gpt2cmd.Command())
	// rootCmd.AddCommand(bertcmd.Command())
}

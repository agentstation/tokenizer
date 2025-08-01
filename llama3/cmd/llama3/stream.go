package llama3cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/agentstation/tokenizer/llama3"
)

var (
	// Stream command flags.
	streamBufferSize int
	streamMaxBuffer  int
	streamAddBOS     bool
	streamAddEOS     bool
	streamOutput     string
)

// newStreamCmd creates the stream subcommand.
func newStreamCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stream",
		Short: "Process text in streaming mode",
		Long: `Process text in streaming mode, outputting tokens as they are generated.

This command is designed for processing large files or real-time input where
you want to see tokens as they are produced rather than waiting for the entire
input to be processed.

The streaming tokenizer uses an internal buffer to accumulate text until it
finds a good tokenization boundary (like whitespace). This prevents splitting
UTF-8 sequences or words unnecessarily.

Input is read from stdin only.`,
		Example: `  # Stream a large file
  cat large_file.txt | tokenizer llama3 stream
  
  # Stream with custom buffer size
  cat data.txt | tokenizer llama3 stream --buffer-size 8192
  
  # Stream without special tokens
  echo "Hello world" | tokenizer llama3 stream --no-bos --no-eos
  
  # Stream with one token per line
  cat input.txt | tokenizer llama3 stream --output newline`,
		RunE: runStream,
	}

	// Add flags
	cmd.Flags().IntVar(&streamBufferSize, "buffer-size", 4096, "Buffer size for reading")
	cmd.Flags().IntVar(&streamMaxBuffer, "max-buffer", 1048576, "Maximum buffer size before forcing tokenization")
	cmd.Flags().BoolVar(&streamAddBOS, "bos", true, "Add beginning of sequence token")
	cmd.Flags().BoolVar(&streamAddEOS, "eos", true, "Add end of sequence token")
	cmd.Flags().StringVarP(&streamOutput, "output", "o", "space", "Output format: space, newline")

	return cmd
}

func runStream(cmd *cobra.Command, args []string) error {

	// Validate output format
	if streamOutput != "space" && streamOutput != "newline" {
		return fmt.Errorf("invalid output format %q: must be 'space' or 'newline'", streamOutput)
	}

	// Initialize tokenizer
	tokenizer, err := llama3.New()
	if err != nil {
		return fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	// Create scanner with options
	scanner := tokenizer.NewScannerOptions(
		os.Stdin,
		llama3.WithBufferSize(streamBufferSize),
		llama3.WithMaxBuffer(streamMaxBuffer),
		llama3.WithEncodeOptions(&llama3.EncodeOptions{
			BOS: streamAddBOS,
			EOS: streamAddEOS,
		}),
	)

	// Process tokens
	first := true
	tokenCount := 0
	for scanner.Scan() {
		token := scanner.Token()
		tokenCount++

		switch streamOutput {
		case "newline":
			fmt.Println(token)
		case "space":
			if !first {
				fmt.Print(" ")
			}
			fmt.Print(token)
			first = false
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("streaming error: %w", err)
	}

	// Final newline for space-separated output
	if streamOutput == "space" && tokenCount > 0 {
		fmt.Println()
	}

	return nil
}

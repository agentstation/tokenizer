package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/agentstation/tokenizer/llama3"
)

func main() {
	var (
		text        = flag.String("text", "", "Text to tokenize")
		decode      = flag.String("decode", "", "Comma-separated token IDs to decode")
		interactive = flag.Bool("i", false, "Interactive mode")
		noBOS       = flag.Bool("no-bos", false, "Don't add beginning-of-text token")
		noEOS       = flag.Bool("no-eos", false, "Don't add end-of-text token")
		verbose     = flag.Bool("v", false, "Verbose output")
	)
	flag.Parse()

	// Create tokenizer
	tokenizer, err := llama3.NewDefault()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tokenizer: %v\n", err)
		os.Exit(1)
	}

	if *verbose {
		fmt.Printf("Tokenizer loaded. Vocabulary size: %d\n", tokenizer.VocabSize())
	}

	// Decode mode
	if *decode != "" {
		tokens := parseTokens(*decode)
		decoded := tokenizer.Decode(tokens)
		fmt.Println(decoded)
		return
	}

	// Interactive mode
	if *interactive {
		runInteractive(tokenizer, *noBOS, *noEOS, *verbose)
		return
	}

	// Single text mode
	if *text != "" {
		opts := &llama3.EncodeOptions{
			BOS: !*noBOS,
			EOS: !*noEOS,
		}
		tokens := tokenizer.Encode(*text, opts)
		
		if *verbose {
			fmt.Printf("Text: %s\n", *text)
			fmt.Printf("Tokens (%d): %v\n", len(tokens), tokens)
			fmt.Printf("Decoded: %s\n", tokenizer.Decode(tokens))
		} else {
			fmt.Println(formatTokens(tokens))
		}
		return
	}

	// Show usage
	flag.Usage()
}

func runInteractive(tokenizer *llama3.Tokenizer, noBOS, noEOS, verbose bool) {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Llama 3 Tokenizer Interactive Mode")
	fmt.Println("Type 'quit' to exit")
	fmt.Println()

	opts := &llama3.EncodeOptions{
		BOS: !noBOS,
		EOS: !noEOS,
	}

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "quit" || line == "exit" {
			break
		}

		// Check for decode command
		if strings.HasPrefix(line, "decode ") {
			tokenStr := strings.TrimPrefix(line, "decode ")
			tokens := parseTokens(tokenStr)
			decoded := tokenizer.Decode(tokens)
			fmt.Printf("Decoded: %s\n", decoded)
			continue
		}

		// Encode the text
		tokens := tokenizer.Encode(line, opts)
		
		if verbose {
			fmt.Printf("Tokens (%d): %v\n", len(tokens), tokens)
			fmt.Printf("Decoded: %s\n", tokenizer.Decode(tokens))
		} else {
			fmt.Println(formatTokens(tokens))
		}
	}
}

func parseTokens(s string) []int {
	parts := strings.Split(s, ",")
	tokens := make([]int, 0, len(parts))
	
	for _, part := range parts {
		part = strings.TrimSpace(part)
		var token int
		if _, err := fmt.Sscanf(part, "%d", &token); err == nil {
			tokens = append(tokens, token)
		}
	}
	
	return tokens
}

func formatTokens(tokens []int) string {
	strs := make([]string, len(tokens))
	for i, t := range tokens {
		strs[i] = fmt.Sprintf("%d", t)
	}
	return strings.Join(strs, ", ")
}
package llama3_test

import (
	"fmt"
	"log"

	"github.com/agentstation/tokenizer/llama3"
)

func ExampleTokenizer_Encode() {
	// Create a tokenizer
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal(err)
	}

	// Encode some text
	text := "Hello, world!"
	tokens := tokenizer.Encode(text, nil)

	fmt.Printf("Text: %s\n", text)
	fmt.Printf("Token count: %d\n", len(tokens))
	// Note: actual output depends on having the Llama 3 data files
}

func ExampleTokenizer_Encode_withoutSpecialTokens() {
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal(err)
	}

	// Encode without special tokens
	opts := &llama3.EncodeOptions{
		BOS: false,
		EOS: false,
	}

	text := "Hello, world!"
	tokens := tokenizer.Encode(text, opts)

	fmt.Printf("Tokens without BOS/EOS: %d\n", len(tokens))
}

func ExampleTokenizer_Decode() {
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal(err)
	}

	// Decode token IDs back to text
	tokens := []int{9906, 1917, 0}
	text := tokenizer.Decode(tokens)

	fmt.Printf("Decoded text: %s\n", text)
	// Output would be: Hello world!
}

func ExampleTokenizer_GetSpecialTokenID() {
	tokenizer, err := llama3.New()
	if err != nil {
		log.Fatal(err)
	}

	// Get the ID of a special token
	tokenID, err := tokenizer.GetSpecialTokenID("<|begin_of_text|>")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Begin-of-text token ID: %d\n", tokenID)
	// Output would be: 128000
}

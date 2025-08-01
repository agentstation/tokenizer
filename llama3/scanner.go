package llama3

import (
	"fmt"
	"io"

	"github.com/agentstation/tokenizer/llama3/scanner"
)

// Scanner provides streaming tokenization following the bufio.Scanner pattern.
// It reads text incrementally and produces tokens one at a time.
type Scanner interface {
	// Scan advances to the next token. Returns false at EOF or on error.
	Scan() bool

	// Token returns the most recent token ID produced by Scan.
	// Valid only after a successful call to Scan.
	Token() int

	// Text returns the text that produced the current token.
	// Valid only after a successful call to Scan.
	Text() string

	// Err returns the first error encountered during scanning.
	Err() error
}

// ScannerOption configures scanner behavior.
type ScannerOption = scanner.Option

// Scanner option functions - these are re-exported from the scanner package.
var (
	// WithBufferSize sets the internal buffer size for reading.
	// Default is 4096 bytes.
	WithBufferSize = scanner.WithBufferSize

	// WithMaxBuffer sets the maximum buffer size before forcing tokenization.
	// This prevents unbounded memory growth for pathological inputs.
	// Default is 1MB.
	WithMaxBuffer = scanner.WithMaxBuffer

	// WithEncodeOptions sets encoding options for the scanner.
	WithEncodeOptions = func(opts *EncodeOptions) ScannerOption {
		return scanner.WithEncodeOptions(&scanner.EncodeOptions{
			BOS: opts.BOS,
			EOS: opts.EOS,
		})
	}
)

// tokenizerAdapter adapts Tokenizer to the scanner.Tokenizer interface.
type tokenizerAdapter struct {
	*Tokenizer
}

// Encode adapts the Encode method.
func (ta *tokenizerAdapter) Encode(text string, opts *scanner.EncodeOptions) []int {
	return ta.Tokenizer.Encode(text, &EncodeOptions{
		BOS: opts.BOS,
		EOS: opts.EOS,
	})
}

// NewScanner creates a scanner for streaming tokenization with default options.
func (t *Tokenizer) NewScanner(r io.Reader) Scanner {
	return scanner.New(&tokenizerAdapter{t}, r)
}

// NewScannerOptions creates a scanner with custom options.
func (t *Tokenizer) NewScannerOptions(r io.Reader, opts ...ScannerOption) Scanner {
	return scanner.NewWithOptions(&tokenizerAdapter{t}, r, opts...)
}

// Process handles large files with controlled memory usage.
// It reads from r, tokenizes the content, and writes token IDs to w.
// Returns the number of tokens written and any error encountered.
func (t *Tokenizer) Process(r io.Reader, w io.Writer) (int64, error) {
	scan := t.NewScanner(r)

	var count int64
	for scan.Scan() {
		token := scan.Token()

		// Write token as binary (4 bytes, little-endian)
		buf := make([]byte, 4)
		buf[0] = byte(token)
		buf[1] = byte(token >> 8)
		buf[2] = byte(token >> 16)
		buf[3] = byte(token >> 24)

		if _, err := w.Write(buf); err != nil {
			return count, fmt.Errorf("write token: %w", err)
		}
		count++
	}

	if err := scan.Err(); err != nil {
		return count, err
	}

	return count, nil
}

// TokenStream provides channel-based streaming for concurrent processing.
// The tokens channel will be closed when scanning completes.
// Any error will be sent on the error channel.
func (t *Tokenizer) TokenStream(r io.Reader) (<-chan int, <-chan error) {
	tokens := make(chan int, 100)
	errc := make(chan error, 1)

	go func() {
		defer close(tokens)
		defer close(errc)

		scan := t.NewScanner(r)
		for scan.Scan() {
			select {
			case tokens <- scan.Token():
				// Token sent successfully
			default:
				// Channel full, this provides backpressure
				tokens <- scan.Token()
			}
		}

		if err := scan.Err(); err != nil {
			errc <- err
		}
	}()

	return tokens, errc
}

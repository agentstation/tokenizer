// Package scanner provides buffered token scanning capabilities.
package scanner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
)

// Tokenizer is the interface required for tokenizing text.
type Tokenizer interface {
	Encode(text string, opts *EncodeOptions) []int
	GetSpecialTokenID(token string) (int, error)
}

// EncodeOptions mirrors the options from the main package.
type EncodeOptions struct {
	BOS bool
	EOS bool
}

// Scanner is the interface for streaming tokenization.
type Scanner interface {
	Scan() bool
	Token() int
	Text() string
	Err() error
}

// scanner implements the Scanner interface for streaming tokenization.
type scanner struct {
	t Tokenizer
	r *bufio.Reader

	// Buffers
	textBuf  bytes.Buffer // Accumulated text to tokenize
	tokens   []int        // Buffered tokens
	tokIndex int          // Current position in tokens buffer
	lastText string       // Text for current token
	pending  []byte       // Pending bytes from incomplete UTF-8 sequence

	// State
	err     error
	done    bool
	sentBOS bool // Track if we've sent BOS token

	// Options
	opts      *EncodeOptions
	bufSize   int // Internal buffer size
	maxBuffer int // Maximum buffer size before forcing tokenization
}

// Option configures scanner behavior.
type Option func(*scanner)

// WithBufferSize sets the internal buffer size for reading.
// Default is 4096 bytes.
func WithBufferSize(size int) Option {
	return func(s *scanner) {
		if size > 0 {
			s.bufSize = size
		}
	}
}

// WithMaxBuffer sets the maximum buffer size before forcing tokenization.
// This prevents unbounded memory growth for pathological inputs.
// Default is 1MB.
func WithMaxBuffer(size int) Option {
	return func(s *scanner) {
		if size > 0 {
			s.maxBuffer = size
		}
	}
}

// WithEncodeOptions sets encoding options for the scanner.
func WithEncodeOptions(opts *EncodeOptions) Option {
	return func(s *scanner) {
		if opts != nil {
			s.opts = opts
		}
	}
}

// New creates a scanner for streaming tokenization with default options.
func New(t Tokenizer, r io.Reader) Scanner {
	return NewWithOptions(t, r)
}

// NewWithOptions creates a scanner with custom options.
func NewWithOptions(t Tokenizer, r io.Reader, opts ...Option) Scanner {
	s := &scanner{
		t:         t,
		r:         bufio.NewReader(r),
		tokens:    make([]int, 0, 32),
		opts:      &EncodeOptions{},
		bufSize:   4096,
		maxBuffer: 1024 * 1024, // 1MB default
	}

	for _, opt := range opts {
		opt(s)
	}

	// Set reader buffer size
	s.r = bufio.NewReaderSize(r, s.bufSize)

	return s
}

// scanBufferedToken returns the next buffered token if available.
func (s *scanner) scanBufferedToken() bool {
	if s.tokIndex < len(s.tokens) {
		s.tokIndex++
		return true
	}
	return false
}

// readMoreData reads data from the input reader into the buffer.
// Returns true if data was read or EOF was reached.
func (s *scanner) readMoreData() (bool, error) {
	buf := make([]byte, s.bufSize)
	n, err := s.r.Read(buf)

	if n > 0 {
		toWrite := buf[:n]
		if len(s.pending) > 0 {
			toWrite = append(s.pending, buf[:n]...)
			s.pending = nil
		}

		toWrite = s.handleBufferLimit(toWrite)
		s.textBuf.Write(toWrite)
	}

	if err == io.EOF {
		s.done = true
		if len(s.pending) > 0 {
			s.textBuf.Write(s.pending)
			s.pending = nil
		}
		return true, nil
	}

	return n > 0, err
}

// handleBufferLimit ensures we don't exceed the buffer limit and handles UTF-8 boundaries.
// Returns the adjusted byte slice that should be written.
func (s *scanner) handleBufferLimit(toWrite []byte) []byte {
	if s.textBuf.Len()+len(toWrite) > s.maxBuffer {
		maxWrite := s.maxBuffer - s.textBuf.Len()
		if maxWrite > 0 && maxWrite < len(toWrite) {
			writeUpTo := findUTF8Boundary(toWrite, maxWrite)
			if writeUpTo < len(toWrite) {
				s.pending = make([]byte, len(toWrite)-writeUpTo)
				copy(s.pending, toWrite[writeUpTo:])
				return toWrite[:writeUpTo]
			}
		}
	}
	return toWrite
}

// handleMaxBufferReached handles the case when the buffer reaches its maximum size.
func (s *scanner) handleMaxBufferReached() {
	if s.textBuf.Len() == s.maxBuffer && len(s.pending) == 0 {
		buf := s.textBuf.Bytes()
		if len(buf) > 0 && !isValidUTF8Ending(buf) {
			splitAt := findLastCompleteUTF8(buf)
			if splitAt < len(buf) {
				s.pending = make([]byte, len(buf)-splitAt)
				copy(s.pending, buf[splitAt:])
				s.textBuf.Truncate(splitAt)
			}
		}
	}
}

// handleEOFTokens handles adding BOS/EOS tokens when the input is empty at EOF.
func (s *scanner) handleEOFTokens() bool {
	if s.textBuf.Len() == 0 {
		// Handle BOS for empty input
		if s.opts.BOS && !s.sentBOS {
			if id, err := s.t.GetSpecialTokenID("<|begin_of_text|>"); err == nil {
				s.tokens = append(s.tokens, id)
				s.sentBOS = true
			}
		}
		// Handle EOS
		if s.opts.EOS {
			if id, err := s.t.GetSpecialTokenID("<|end_of_text|>"); err == nil {
				s.tokens = append(s.tokens, id)
			}
		}
		if len(s.tokens) > 0 {
			s.tokIndex = 0
			return true
		}
	}
	return false
}

// tokenizeBuffer tokenizes the accumulated text in the buffer.
func (s *scanner) tokenizeBuffer() bool {
	text := s.textBuf.String()
	if len(text) == 0 {
		return false
	}

	// For first chunk, handle BOS token
	addBOS := s.opts.BOS && !s.sentBOS
	if addBOS {
		s.sentBOS = true
	}

	// Create temporary options for this chunk
	chunkOpts := &EncodeOptions{
		BOS: addBOS,
		EOS: false, // Handle EOS separately at the end
	}

	// Tokenize the chunk
	s.tokens = s.t.Encode(text, chunkOpts)
	s.textBuf.Reset()

	// Handle EOS if this is the last chunk
	if s.done && s.opts.EOS {
		if id, err := s.t.GetSpecialTokenID("<|end_of_text|>"); err == nil {
			s.tokens = append(s.tokens, id)
		}
	}

	return len(s.tokens) > 0
}

// Scan advances to the next token.
func (s *scanner) Scan() bool {
	if s.err != nil {
		return false
	}

	// If we have buffered tokens, return the next one
	if s.scanBufferedToken() {
		return true
	}

	// Check if we're done and have no more tokens to process
	if s.done && s.textBuf.Len() == 0 && s.tokIndex >= len(s.tokens) {
		return false
	}

	// Need to read more text and tokenize
	s.tokens = s.tokens[:0]
	s.tokIndex = 0

	// Read and accumulate text until we have enough to tokenize
	if err := s.readAndAccumulateText(); err != nil {
		s.err = &ScanError{
			Offset: int64(s.textBuf.Len()),
			Text:   s.textBuf.String(),
			Err:    err,
		}
		return false
	}

	// Tokenize the accumulated text or check if we have tokens from EOF handling
	if s.tokenizeBuffer() || len(s.tokens) > 0 {
		// We have tokens, advance to the first one
		s.tokIndex = 1
		return true
	}

	return false
}

// readAndAccumulateText reads data until we have enough to tokenize or reach EOF.
func (s *scanner) readAndAccumulateText() error {
	for {
		// Try to read more data
		_, err := s.readMoreData()

		// Check if we've hit the maximum buffer size
		if s.textBuf.Len() >= s.maxBuffer {
			s.handleMaxBufferReached()
			break
		}

		if err != nil {
			return err
		}

		if s.done {
			// At EOF, check if we need to handle empty input
			if s.textBuf.Len() == 0 && s.handleEOFTokens() {
				return nil
			}
			break
		}

		// Look for a good tokenization boundary
		if s.hasTokenizationBoundary() {
			break
		}
	}

	return nil
}

// Token returns the current token ID.
func (s *scanner) Token() int {
	if s.tokIndex > 0 && s.tokIndex <= len(s.tokens) {
		return s.tokens[s.tokIndex-1]
	}
	return 0
}

// Text returns the text that produced the current token.
// Note: This returns the entire chunk that was tokenized, not individual token text.
func (s *scanner) Text() string {
	return s.lastText
}

// Err returns any error encountered during scanning.
func (s *scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

// hasTokenizationBoundary checks if the buffer ends at a good tokenization boundary.
// This helps prevent splitting UTF-8 sequences or words unnecessarily.
func (s *scanner) hasTokenizationBoundary() bool {
	if s.textBuf.Len() == 0 {
		return false
	}

	// Get the last few bytes to check
	buf := s.textBuf.Bytes()
	if len(buf) == 0 {
		return false
	}

	// Check if we're at a whitespace boundary
	lastByte := buf[len(buf)-1]
	if lastByte == ' ' || lastByte == '\n' || lastByte == '\t' || lastByte == '\r' {
		return true
	}

	// Check if we might be in the middle of a UTF-8 sequence
	// UTF-8 continuation bytes start with 10xxxxxx
	if lastByte&0xC0 == 0x80 {
		// We're in the middle of a UTF-8 sequence, don't split here
		return false
	}

	// If buffer is getting large, accept any UTF-8 boundary
	if s.textBuf.Len() > s.bufSize/2 {
		return true
	}

	return false
}

// isValidUTF8Ending checks if the buffer ends at a valid UTF-8 boundary.
func isValidUTF8Ending(buf []byte) bool {
	if len(buf) == 0 {
		return true
	}

	// Check the last byte
	lastByte := buf[len(buf)-1]

	// ASCII is always valid
	if lastByte < 0x80 {
		return true
	}

	// If it's a continuation byte, we might have an incomplete sequence
	if lastByte&0xC0 == 0x80 {
		// This is a continuation byte, not a valid ending
		return false
	}

	// It's the start of a multi-byte sequence, check if complete
	expectedLen := 0
	if lastByte&0xE0 == 0xC0 {
		expectedLen = 2
	} else if lastByte&0xF0 == 0xE0 {
		expectedLen = 3
	} else if lastByte&0xF8 == 0xF0 {
		expectedLen = 4
	}

	// Count continuation bytes after this start byte
	// Since this is the last byte, we expect 0 continuation bytes
	// So this is incomplete
	return expectedLen <= 1
}

// findLastCompleteUTF8 finds the last complete UTF-8 character boundary.
func findLastCompleteUTF8(buf []byte) int {
	for i := len(buf) - 1; i >= 0 && i >= len(buf)-4; i-- {
		b := buf[i]

		// ASCII byte - this is a complete character
		if b < 0x80 {
			return i + 1
		}

		// Start of UTF-8 sequence
		if b&0xC0 != 0x80 {
			// Check if we have the complete sequence
			seqLen := 0
			if b&0xE0 == 0xC0 {
				seqLen = 2
			} else if b&0xF0 == 0xE0 {
				seqLen = 3
			} else if b&0xF8 == 0xF0 {
				seqLen = 4
			}

			if i+seqLen <= len(buf) {
				// Complete sequence
				return i + seqLen
			}
			// Incomplete sequence
			return i
		}
	}

	// Shouldn't get here with valid UTF-8
	return len(buf)
}

// findUTF8Boundary finds the last valid UTF-8 boundary before maxBytes.
// It returns the number of bytes that can be safely written without
// splitting a UTF-8 character.
func findUTF8Boundary(data []byte, maxBytes int) int {
	if maxBytes >= len(data) {
		return len(data)
	}

	// Start from maxBytes and work backwards to find a valid boundary
	for i := maxBytes; i > 0 && i > maxBytes-4; i-- {
		if i >= len(data) {
			continue
		}

		b := data[i]
		// Check if this is the start of a UTF-8 sequence or ASCII
		if b < 0x80 || b&0xC0 != 0x80 {
			// This is a valid boundary
			return i
		}
	}

	// If we can't find a good boundary, check from the beginning
	// of where we want to cut
	if maxBytes < len(data) {
		b := data[maxBytes]
		if b&0xC0 == 0x80 {
			// We're in the middle of a UTF-8 sequence
			// Find the start of this sequence
			for i := maxBytes - 1; i >= 0 && i >= maxBytes-4; i-- {
				if data[i]&0xC0 != 0x80 {
					return i
				}
			}
		}
	}

	return maxBytes
}

// ScanError represents an error during scanning with context.
type ScanError struct {
	Offset int64  // Byte offset where error occurred
	Text   string // Text being processed (may be truncated)
	Err    error  // Underlying error
}

func (e *ScanError) Error() string {
	preview := e.Text
	if len(preview) > 50 {
		preview = preview[:50] + "..."
	}
	return fmt.Sprintf("tokenization error at offset %d (text: %q): %v",
		e.Offset, preview, e.Err)
}

func (e *ScanError) Unwrap() error {
	return e.Err
}

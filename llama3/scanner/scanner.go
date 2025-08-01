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
	t         Tokenizer
	r         *bufio.Reader
	
	// Buffers
	textBuf   bytes.Buffer    // Accumulated text to tokenize
	tokens    []int           // Buffered tokens
	tokIndex  int             // Current position in tokens buffer
	lastText  string          // Text for current token
	pending   []byte          // Pending bytes from incomplete UTF-8 sequence
	
	// State
	err       error
	done      bool
	sentBOS   bool  // Track if we've sent BOS token
	
	// Options
	opts      *EncodeOptions
	bufSize   int             // Internal buffer size
	maxBuffer int             // Maximum buffer size before forcing tokenization
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

// Scan advances to the next token.
func (s *scanner) Scan() bool {
	if s.err != nil {
		return false
	}
	
	// If we have buffered tokens, return the next one
	if s.tokIndex < len(s.tokens) {
		s.tokIndex++
		return true
	}
	
	// Check if we're done and have no more tokens
	if s.done && s.textBuf.Len() == 0 {
		return false
	}
	
	// Need to read more text and tokenize
	s.tokens = s.tokens[:0]
	s.tokIndex = 0
	
	// Read until we have enough text to tokenize
	for {
		// Try to read more data
		buf := make([]byte, s.bufSize)
		n, err := s.r.Read(buf)
		
		if n > 0 {
			// If we have pending bytes from a previous incomplete UTF-8 sequence,
			// prepend them to the new data
			toWrite := buf[:n]
			if len(s.pending) > 0 {
				toWrite = append(s.pending, buf[:n]...)
				s.pending = nil
			}
			
			// Before writing, check if this would exceed our buffer limit
			// and if so, check for UTF-8 boundaries in the new data
			if s.textBuf.Len() + len(toWrite) > s.maxBuffer {
				// We need to be careful about UTF-8 boundaries
				maxWrite := s.maxBuffer - s.textBuf.Len()
				if maxWrite > 0 && maxWrite < len(toWrite) {
					// Check if we would split a UTF-8 sequence
					writeUpTo := findUTF8Boundary(toWrite, maxWrite)
					if writeUpTo < len(toWrite) {
						// Save the rest for next iteration
						s.pending = make([]byte, len(toWrite)-writeUpTo)
						copy(s.pending, toWrite[writeUpTo:])
						toWrite = toWrite[:writeUpTo]
					}
				}
			}
			
			s.textBuf.Write(toWrite)
		}
		
		// Check if we've hit the maximum buffer size
		if s.textBuf.Len() >= s.maxBuffer {
			// Check if we need to handle a UTF-8 boundary at the exact limit
			if s.textBuf.Len() == s.maxBuffer && len(s.pending) == 0 {
				// We might have split a UTF-8 character at the boundary
				buf := s.textBuf.Bytes()
				if len(buf) > 0 && !isValidUTF8Ending(buf) {
					// Find the last complete UTF-8 character
					splitAt := findLastCompleteUTF8(buf)
					if splitAt < len(buf) {
						// Save the incomplete sequence
						s.pending = make([]byte, len(buf)-splitAt)
						copy(s.pending, buf[splitAt:])
						s.textBuf.Truncate(splitAt)
					}
				}
			}
			// Force tokenization of what we have
			break
		}
		
		if err != nil {
			if err == io.EOF {
				// End of input
				s.done = true
				
				// If we have pending bytes at EOF, add them to the buffer
				if len(s.pending) > 0 {
					s.textBuf.Write(s.pending)
					s.pending = nil
				}
				
				if s.textBuf.Len() == 0 {
					// No more data to process
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
						return s.Scan()
					}
					return false
				}
				// Process remaining text
				break
			}
			// Read error
			s.err = &ScanError{
				Offset: int64(s.textBuf.Len()),
				Text:   s.textBuf.String(),
				Err:    err,
			}
			return false
		}
		
		// Look for a good tokenization boundary
		if s.hasTokenizationBoundary() {
			break
		}
	}
	
	// Tokenize accumulated text
	text := s.textBuf.String()
	if len(text) > 0 {
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
		
		if len(s.tokens) > 0 {
			s.tokIndex = 0
			return s.Scan() // Recursively call to return first token
		}
	}
	
	return false
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


// isValidUTF8Ending checks if the buffer ends at a valid UTF-8 boundary
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

// findLastCompleteUTF8 finds the last complete UTF-8 character boundary
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
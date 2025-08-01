package llama3

import (
	"errors"
	"fmt"
)

// Common errors
var (
	// ErrDataNotFound indicates that the tokenizer data files could not be found
	ErrDataNotFound = errors.New("tokenizer data not found")

	// ErrInvalidToken indicates an invalid token was provided
	ErrInvalidToken = errors.New("invalid token")

	// ErrTokenNotFound indicates a token was not found in the vocabulary
	ErrTokenNotFound = errors.New("token not found")

	// ErrInvalidTokenID indicates an invalid token ID was provided
	ErrInvalidTokenID = errors.New("invalid token ID")
)

// DataError represents an error related to tokenizer data loading or processing
type DataError struct {
	Op   string // Operation that failed
	Path string // File path if applicable
	Err  error  // Underlying error
}

func (e *DataError) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("tokenizer data error: %s %s: %v", e.Op, e.Path, e.Err)
	}
	return fmt.Sprintf("tokenizer data error: %s: %v", e.Op, e.Err)
}

func (e *DataError) Unwrap() error {
	return e.Err
}

// TokenError represents an error related to token operations
type TokenError struct {
	Token   string // The token that caused the error
	TokenID int    // The token ID if applicable
	Op      string // Operation that failed
	Err     error  // Underlying error
}

func (e *TokenError) Error() string {
	if e.Token != "" {
		return fmt.Sprintf("token error: %s %q: %v", e.Op, e.Token, e.Err)
	}
	if e.TokenID != 0 {
		return fmt.Sprintf("token error: %s token_id=%d: %v", e.Op, e.TokenID, e.Err)
	}
	return fmt.Sprintf("token error: %s: %v", e.Op, e.Err)
}

func (e *TokenError) Unwrap() error {
	return e.Err
}

// ConfigError represents an error in tokenizer configuration
type ConfigError struct {
	Field string // Configuration field that has an error
	Value any    // The invalid value
	Err   error  // Underlying error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("config error: %s=%v: %v", e.Field, e.Value, e.Err)
}

func (e *ConfigError) Unwrap() error {
	return e.Err
}

// Helper functions for creating errors

// NewDataError creates a new DataError
func NewDataError(op, path string, err error) error {
	return &DataError{Op: op, Path: path, Err: err}
}

// NewTokenError creates a new TokenError
func NewTokenError(op, token string, err error) error {
	return &TokenError{Op: op, Token: token, Err: err}
}

// NewTokenIDError creates a new TokenError with a token ID
func NewTokenIDError(op string, tokenID int, err error) error {
	return &TokenError{Op: op, TokenID: tokenID, Err: err}
}

// NewConfigError creates a new ConfigError
func NewConfigError(field string, value any, err error) error {
	return &ConfigError{Field: field, Value: value, Err: err}
}

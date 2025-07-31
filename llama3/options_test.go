package llama3

import (
	"testing"
)

func TestNewWithOptions(t *testing.T) {
	// Test creating tokenizer with options
	tokenizer, err := New(
		WithCacheSize(100),
	)
	
	// Should succeed if data files are available
	if err == nil && tokenizer != nil {
		if tokenizer.cacheSize != 100 {
			t.Errorf("Expected cache size 100, got %d", tokenizer.cacheSize)
		}
	} else if err != nil {
		// Skip if data files are not available
		t.Skipf("Skipping test: %v", err)
	}
}

func TestWithSpecialTokens(t *testing.T) {
	tests := []struct {
		name    string
		tokens  []string
		wantErr bool
	}{
		{
			name:    "valid special tokens",
			tokens:  []string{"<|test|>", "<|another|>"},
			wantErr: false,
		},
		{
			name:    "invalid special token format",
			tokens:  []string{"<|test|>", "invalid"},
			wantErr: true,
		},
		{
			name:    "duplicate special tokens",
			tokens:  []string{"<|test|>", "<|test|>"},
			wantErr: true,
		},
		{
			name:    "nil tokens",
			tokens:  nil,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithSpecialTokens(tt.tokens)
			tempConfig := &tokenizerConfig{}
			err := opt(tempConfig)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("WithSpecialTokens() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestWithCacheSize(t *testing.T) {
	tests := []struct {
		name    string
		size    int
		wantErr bool
	}{
		{"positive size", 100, false},
		{"zero size", 0, false},
		{"negative size", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opt := WithCacheSize(tt.size)
			tempConfig := &tokenizerConfig{}
			err := opt(tempConfig)
			
			if (err != nil) != tt.wantErr {
				t.Errorf("WithCacheSize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
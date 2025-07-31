package llama3

import (
	"fmt"
	"reflect"
	"testing"
)

func TestStateMachine(t *testing.T) {
	// Test cases from JavaScript output
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "           grabbed",
			expected: []string{"          ", " grabbed"}, // 10 spaces, then space+word
		},
		{
			input:    "\ttabs\t\t\t\tout here",
			expected: []string{"\ttabs", "\t\t\t", "\tout", " here"},
		},
		{
			input:    "Hello world",
			expected: []string{"Hello", " world"},
		},
		{
			input:    "can't",
			expected: []string{"can", "'t"},
		},
		{
			input:    "123 456",
			expected: []string{"123", " ", "456"},
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			sm := NewStateMachine(tc.input)
			tokens := sm.TokenizeWithStateMachine()
			
			if !reflect.DeepEqual(tokens, tc.expected) {
				t.Errorf("Input: %q\nExpected: %q\nGot:      %q", tc.input, tc.expected, tokens)
			} else {
				fmt.Printf("✓ %q → %q\n", tc.input, tokens)
			}
		})
	}
}
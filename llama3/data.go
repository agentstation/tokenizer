package llama3

import (
	"encoding/base64"
	"fmt"
	"strings"
)

// Note: In production, these would contain the actual base64-encoded data.
// For now, we'll load them from external files or embed them using go:embed.

var (
	// defaultVocabBase64 contains the base64-encoded vocabulary data
	defaultVocabBase64 = "" // This will be populated from file or embed
	
	// defaultMergesBinary contains the base64-encoded compressed merge data  
	defaultMergesBinary = "" // This will be populated from file or embed
)

// decodeVocabulary decodes the base64-encoded vocabulary data.
func decodeVocabulary(vocabBase64 string) ([]string, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(vocabBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 vocabulary: %w", err)
	}
	
	// Split by newlines to get individual tokens
	vocabString := string(decoded)
	tokens := strings.Split(vocabString, "\n")
	
	// Filter out empty tokens
	result := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if token != "" {
			result = append(result, token)
		}
	}
	
	return result, nil
}

// decompressMerges decompresses the binary merge data.
func (t *Tokenizer) decompressMerges(mergesBinary string) (map[string]int, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(mergesBinary)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 merges: %w", err)
	}
	
	// Each merge is represented by two 17-bit integers packed into bytes
	tokenIDs := unpack17BitIntegers(decoded)
	
	// Create merge map
	merges := make(map[string]int, len(tokenIDs)/2)
	for i := 0; i < len(tokenIDs); i += 2 {
		if i+1 >= len(tokenIDs) {
			break
		}
		
		id1 := tokenIDs[i]
		id2 := tokenIDs[i+1]
		
		if id1 >= len(t.vocabByID) || id2 >= len(t.vocabByID) {
			continue // Skip invalid token IDs
		}
		
		mergeIdentifier := t.getMergeIdentifier(id1, id2)
		// Priority is based on position in the merge list
		merges[mergeIdentifier] = i/2 + 1
	}
	
	return merges, nil
}

// unpack17BitIntegers unpacks 17-bit integers from a byte array.
func unpack17BitIntegers(data []byte) []int {
	if len(data) == 0 {
		return []int{}
	}
	
	maxBits := len(data) * 8
	firstPaddingBit := maxBits - (maxBits % 17)
	tokenIDs := make([]int, 0, firstPaddingBit/17)
	
	bitMask := []int{
		0b11111111111111111,      // 17 bits
		0b111111111111111111,     // 18 bits
		0b1111111111111111111,    // 19 bits
		0b11111111111111111111,   // 20 bits
		0b111111111111111111111,  // 21 bits
		0b1111111111111111111111, // 22 bits
		0b11111111111111111111111,// 23 bits
		0b111111111111111111111111,// 24 bits
	}
	
	for bitIndex := 0; bitIndex < firstPaddingBit; bitIndex += 17 {
		byteIndex := bitIndex / 8
		
		// Extract 3 bytes (24 bits) that contain our 17-bit value
		if byteIndex+2 >= len(data) {
			break
		}
		
		byte1 := int(data[byteIndex])
		byte2 := int(data[byteIndex+1])
		byte3 := int(data[byteIndex+2])
		
		// Combine the bytes
		tokenID := (byte1 << 16) + (byte2 << 8) + byte3
		
		// Apply bit mask to remove extra bits from the left
		bitOffset := bitIndex % 8
		if bitOffset < len(bitMask) {
			tokenID = tokenID & bitMask[8-bitOffset-1]
		}
		
		// Shift right to remove extra bits from the right
		rightShift := 8 - (17 - 8 - (8 - bitOffset))
		if rightShift > 0 && rightShift < 24 {
			tokenID = tokenID >> rightShift
		}
		
		tokenIDs = append(tokenIDs, tokenID)
	}
	
	return tokenIDs
}
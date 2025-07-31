package llama3

import (
	"strings"
)

var (
	// bytesToUnicode maps byte values to unicode characters for encoding
	bytesToUnicode map[byte]rune
	// unicodeToBytes maps unicode characters back to byte values for decoding
	unicodeToBytes map[rune]byte
)

func init() {
	// Initialize byte-to-unicode mappings
	bytesToUnicode, unicodeToBytes = createByteMappings()
}

// createByteMappings creates the byte-to-unicode and unicode-to-byte mappings
// following the same logic as the JavaScript implementation. This creates a
// reversible mapping that allows encoding arbitrary bytes as valid Unicode
// characters for tokenization.
func createByteMappings() (map[byte]rune, map[rune]byte) {
	bs := make([]int, 0, 256)

	// Add printable ASCII range
	for i := asciiPrintableStart; i <= asciiPrintableEnd; i++ {
		bs = append(bs, int(i))
	}
	// Add first extended ASCII range
	for i := extendedAsciiStart1; i <= extendedAsciiEnd1; i++ {
		bs = append(bs, int(i))
	}
	// Add second extended ASCII range
	for i := extendedAsciiStart2; i <= extendedAsciiEnd2; i++ {
		bs = append(bs, int(i))
	}

	cs := make([]int, len(bs))
	copy(cs, bs)

	n := 0
	for b := 0; b < 256; b++ {
		found := false
		for _, val := range bs {
			if val == b {
				found = true
				break
			}
		}
		if !found {
			bs = append(bs, b)
			cs = append(cs, unicodeOffset+n)
			n++
		}
	}

	// Create the mappings
	bToU := make(map[byte]rune, 256)
	uToB := make(map[rune]byte, 256)

	for i, b := range bs {
		bToU[byte(b)] = rune(cs[i])
		uToB[rune(cs[i])] = byte(b)
	}

	return bToU, uToB
}

// encodeBytes converts UTF-8 bytes to the custom byte-level representation.
// Each byte is mapped to a specific Unicode character to ensure all byte
// sequences can be represented as valid text for tokenization.
func encodeBytes(data []byte) string {
	var sb strings.Builder
	sb.Grow(len(data))

	for _, b := range data {
		if r, ok := bytesToUnicode[b]; ok {
			sb.WriteRune(r)
		}
	}

	return sb.String()
}

// decodeTokenBytes converts a token string back to UTF-8 bytes.
// This reverses the encoding performed by encodeBytes, restoring the
// original byte sequence from the Unicode representation.
func decodeTokenBytes(token string) []byte {
	result := make([]byte, 0, len(token))

	for _, r := range token {
		if b, ok := unicodeToBytes[r]; ok {
			result = append(result, b)
		}
	}

	return result
}

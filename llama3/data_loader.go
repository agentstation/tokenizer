//go:build !embed

package llama3

import (
	"os"
	"path/filepath"
)

// LoadDataFromFiles loads vocabulary and merge data from external files.
// This is useful for development and testing without embedding large files.
func LoadDataFromFiles(vocabPath, mergesPath string) error {
	// Load vocabulary
	vocabData, err := os.ReadFile(vocabPath)
	if err != nil {
		return NewDataError("read file", vocabPath, err)
	}
	defaultVocabBase64 = string(vocabData)

	// Load merges
	mergesData, err := os.ReadFile(mergesPath)
	if err != nil {
		return NewDataError("read file", mergesPath, err)
	}
	defaultMergesBinary = string(mergesData)

	return nil
}

// TryLoadDataFromStandardPaths attempts to load data from common locations.
func TryLoadDataFromStandardPaths() error {
	// Try current directory
	if err := LoadDataFromFiles("vocab_base64.txt", "merges_binary.txt"); err == nil {
		return nil
	}

	// Try llama3 subdirectory
	if err := LoadDataFromFiles(
		filepath.Join("llama3", "vocab_base64.txt"),
		filepath.Join("llama3", "merges_binary.txt"),
	); err == nil {
		return nil
	}

	// Try parent directory (useful for tests)
	if err := LoadDataFromFiles(
		filepath.Join("..", "vocab_base64.txt"),
		filepath.Join("..", "merges_binary.txt"),
	); err == nil {
		return nil
	}

	return NewDataError("find data files", "", ErrDataNotFound)
}

func init() {
	// Try to load data from files if not embedded
	if defaultVocabBase64 == "" || defaultMergesBinary == "" {
		_ = TryLoadDataFromStandardPaths()
	}
}

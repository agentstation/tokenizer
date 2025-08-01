package vocabulary

import (
	"fmt"
	"os"
)

// EmbeddedDataLoader implements data loading using the default embedded data.
type EmbeddedDataLoader struct {
	vocabBase64  string
	mergesBinary string
}

// NewDefaultLoader creates a loader that uses the default embedded data.
func NewDefaultLoader() *EmbeddedDataLoader {
	return &EmbeddedDataLoader{
		vocabBase64:  EmbeddedVocabulary,
		mergesBinary: EmbeddedMergeRules,
	}
}

// NewCustomLoader creates a loader with custom vocabulary and merges data.
func NewCustomLoader(vocabBase64, mergesBinary string) *EmbeddedDataLoader {
	return &EmbeddedDataLoader{
		vocabBase64:  vocabBase64,
		mergesBinary: mergesBinary,
	}
}

// LoadVocabulary loads and decodes the vocabulary data.
func (d *EmbeddedDataLoader) LoadVocabulary() ([]string, error) {
	if d.vocabBase64 == "" {
		return nil, fmt.Errorf("vocabulary data not found")
	}
	return DecodeVocabulary(d.vocabBase64)
}

// LoadMergesData returns the raw merges binary data for decompression.
// The actual decompression is done by the tokenizer since it needs the
// getMergeIdentifier function.
func (d *EmbeddedDataLoader) LoadMergesData() (string, error) {
	if d.mergesBinary == "" {
		return "", fmt.Errorf("merges data not found")
	}
	return d.mergesBinary, nil
}

// FileLoader implements data loading by reading from files.
type FileLoader struct {
	VocabPath  string
	MergesPath string
}

// NewFileLoader creates a loader that reads from files.
func NewFileLoader(vocabPath, mergesPath string) *FileLoader {
	return &FileLoader{
		VocabPath:  vocabPath,
		MergesPath: mergesPath,
	}
}

// LoadVocabulary reads and decodes vocabulary from a file.
func (f *FileLoader) LoadVocabulary() ([]string, error) {
	data, err := os.ReadFile(f.VocabPath)
	if err != nil {
		return nil, fmt.Errorf("read vocabulary file %s: %w", f.VocabPath, err)
	}
	return DecodeVocabulary(string(data))
}

// LoadMergesData reads merges data from a file.
func (f *FileLoader) LoadMergesData() (string, error) {
	data, err := os.ReadFile(f.MergesPath)
	if err != nil {
		return "", fmt.Errorf("read merges file %s: %w", f.MergesPath, err)
	}
	return string(data), nil
}

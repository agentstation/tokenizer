// Package llama3 implements the Llama 3 tokenizer in Go.
// This file contains the public API including interfaces and options.
package llama3

import (
	"github.com/agentstation/tokenizer/llama3/internal/vocabulary"
)

// VocabularyDataLoader is the interface for loading tokenizer vocabulary data.
// This includes vocabulary and merge rules needed for tokenization.
//
// Implementations can load data from embedded resources, files, or custom sources.
// The tokenizer will call LoadVocabulary first, then LoadMerges.
type VocabularyDataLoader interface {
	// LoadVocabulary loads and returns the vocabulary tokens.
	// The returned slice contains tokens indexed by their token ID.
	LoadVocabulary() ([]string, error)

	// LoadMerges loads and returns the BPE merge rules.
	// The returned map uses merge identifiers as keys and priorities as values.
	LoadMerges() (map[string]int, error)
}

// VocabularyDataLoaderFunc is an adapter to allow using functions as VocabularyDataLoaders.
// This is useful for testing or custom data loading logic.
type VocabularyDataLoaderFunc struct {
	VocabFunc  func() ([]string, error)
	MergesFunc func() (map[string]int, error)
}

// LoadVocabulary calls the VocabFunc.
func (d VocabularyDataLoaderFunc) LoadVocabulary() ([]string, error) {
	return d.VocabFunc()
}

// LoadMerges calls the MergesFunc.
func (d VocabularyDataLoaderFunc) LoadMerges() (map[string]int, error) {
	return d.MergesFunc()
}

// Internal data loader implementations

// embeddedDataLoader loads data from embedded resources.
// embeddedVocabularySource loads vocabulary data from embedded resources.
// This is the default source that uses the pre-packaged Llama3 vocabulary.
type embeddedVocabularySource struct {
	t *Tokenizer
}

func (d *embeddedVocabularySource) LoadVocabulary() ([]string, error) {
	loader := vocabulary.NewDefaultLoader()
	vocab, err := loader.LoadVocabulary()
	if err != nil {
		return nil, NewDataError("load vocabulary", "", err)
	}
	return vocab, nil
}

func (d *embeddedVocabularySource) LoadMerges() (map[string]int, error) {
	loader := vocabulary.NewDefaultLoader()
	mergesData, err := loader.LoadMergesData()
	if err != nil {
		return nil, NewDataError("load merges data", "", err)
	}

	merges, err := vocabulary.DecompressMergeRules(mergesData, d.t.tokens, d.t.getMergeIdentifier)
	if err != nil {
		return nil, NewDataError("decompress merges", "", err)
	}

	return merges, nil
}

// fileVocabularySource loads vocabulary data from external files.
// This allows using custom vocabularies instead of the embedded defaults.
type fileVocabularySource struct {
	vocabPath  string
	mergesPath string
	t          *Tokenizer
}

func (f *fileVocabularySource) LoadVocabulary() ([]string, error) {
	loader := vocabulary.NewFileLoader(f.vocabPath, f.mergesPath)
	vocab, err := loader.LoadVocabulary()
	if err != nil {
		return nil, NewDataError("load vocabulary", f.vocabPath, err)
	}
	return vocab, nil
}

func (f *fileVocabularySource) LoadMerges() (map[string]int, error) {
	loader := vocabulary.NewFileLoader(f.vocabPath, f.mergesPath)
	mergesData, err := loader.LoadMergesData()
	if err != nil {
		return nil, NewDataError("load merges", f.mergesPath, err)
	}

	merges, err := vocabulary.DecompressMergeRules(mergesData, f.t.tokens, f.t.getMergeIdentifier)
	if err != nil {
		return nil, NewDataError("decompress merges", f.mergesPath, err)
	}

	return merges, nil
}

// fileLoaderMarker is a placeholder that will be replaced with the actual file loader.
type fileLoaderMarker struct {
	vocabPath  string
	mergesPath string
}

func (f *fileLoaderMarker) LoadVocabulary() ([]string, error) {
	panic("fileLoaderMarker should be replaced before use")
}

func (f *fileLoaderMarker) LoadMerges() (map[string]int, error) {
	panic("fileLoaderMarker should be replaced before use")
}

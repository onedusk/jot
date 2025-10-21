// Package tokenizer provides token counting functionality for text using OpenAI-compatible tokenizers.
package tokenizer

import (
	"github.com/pkoukk/tiktoken-go"
)

// Tokenizer defines the interface for text tokenization and token counting.
type Tokenizer interface {
	// Encode converts text into a sequence of token IDs
	Encode(text string) []int

	// Count returns the number of tokens in the given text
	Count(text string) int
}

// TikTokenizer implements the Tokenizer interface using tiktoken-go library
// with cl100k_base encoding for GPT-4 and Claude compatibility.
type TikTokenizer struct {
	encoding *tiktoken.Tiktoken
}

// Encode converts text into a sequence of token IDs using cl100k_base encoding.
func (t *TikTokenizer) Encode(text string) []int {
	return t.encoding.Encode(text, nil, nil)
}

// Count returns the number of tokens in the given text.
func (t *TikTokenizer) Count(text string) int {
	return len(t.Encode(text))
}

// NewTokenizer creates a new TikTokenizer instance with cl100k_base encoding.
// This encoding is compatible with GPT-4, GPT-3.5-turbo, and Claude models.
// Returns an error if the encoding cannot be initialized.
func NewTokenizer() (*TikTokenizer, error) {
	encoding, err := tiktoken.GetEncoding("cl100k_base")
	if err != nil {
		return nil, err
	}

	return &TikTokenizer{
		encoding: encoding,
	}, nil
}

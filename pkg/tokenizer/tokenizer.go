// Package tokenizer provides token counting functionality for text using OpenAI-compatible tokenizers.
package tokenizer

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/pkoukk/tiktoken-go"
)

// cl100kBase is the cl100k_base BPE vocabulary, embedded so that tokenizing
// never needs the network. It is the file tiktoken publishes at
// https://openaipublic.blob.core.windows.net/encodings/cl100k_base.tiktoken
// (sha256 223921b76ee99bde995b7ff738513eef100fb51d18c93597a113bcffe865b2a7).
//
//go:embed cl100k_base.tiktoken
var cl100kBase []byte

// cl100kPattern is the cl100k_base pre-tokenization pattern, as defined by tiktoken.
const cl100kPattern = `(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`

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

// NewTokenizer creates a new TikTokenizer with the cl100k_base encoding, the
// encoding used by GPT-4 and GPT-3.5-turbo. The vocabulary is embedded in the
// binary, so no network access or download cache is needed. The encoding is
// built directly rather than through tiktoken.GetEncoding, so tiktoken's
// process-wide BPE loader is left untouched for other callers.
func NewTokenizer() (*TikTokenizer, error) {
	ranks, err := parseRanks(cl100kBase)
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded cl100k_base vocabulary: %w", err)
	}

	specialTokens := map[string]int{
		tiktoken.ENDOFTEXT:   100257,
		tiktoken.FIM_PREFIX:  100258,
		tiktoken.FIM_MIDDLE:  100259,
		tiktoken.FIM_SUFFIX:  100260,
		tiktoken.ENDOFPROMPT: 100276,
	}
	bpe, err := tiktoken.NewCoreBPE(ranks, specialTokens, cl100kPattern)
	if err != nil {
		return nil, err
	}

	specialTokensSet := make(map[string]any, len(specialTokens))
	for token := range specialTokens {
		specialTokensSet[token] = true
	}
	encoding := &tiktoken.Encoding{
		Name:           tiktoken.MODEL_CL100K_BASE,
		PatStr:         cl100kPattern,
		MergeableRanks: ranks,
		SpecialTokens:  specialTokens,
	}

	return &TikTokenizer{
		encoding: tiktoken.NewTiktoken(bpe, encoding, specialTokensSet),
	}, nil
}

// parseRanks parses a .tiktoken vocabulary, which has one "<base64 token> <rank>"
// pair per line.
func parseRanks(data []byte) (map[string]int, error) {
	ranks := make(map[string]int, bytes.Count(data, []byte("\n")))
	for i, line := range bytes.Split(data, []byte("\n")) {
		if len(line) == 0 {
			continue
		}
		encoded, rank, ok := bytes.Cut(line, []byte(" "))
		if !ok {
			return nil, fmt.Errorf("line %d: expected \"<token> <rank>\"", i+1)
		}
		token, err := base64.StdEncoding.DecodeString(string(encoded))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		n, err := strconv.Atoi(string(rank))
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", i+1, err)
		}
		ranks[string(token)] = n
	}
	return ranks, nil
}

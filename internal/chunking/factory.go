package chunking

import (
	"fmt"

	"github.com/onedusk/jot/internal/tokenizer"
)

// NewChunkStrategy creates a ChunkStrategy based on the given strategy name.
// Supported strategies:
//   - "fixed": Fixed-size token-based chunking with word boundary preservation
//   - "headers": Markdown header-based chunking (splits at # headers)
//   - "recursive": Hierarchical text splitting (paragraph -> line -> space -> char)
//   - "semantic": Semantic boundary detection (currently stub, falls back to fixed)
//
// Parameters:
//   - name: The strategy name (case-insensitive)
//   - tok: The tokenizer to use for token counting
//
// Returns:
//   - A ChunkStrategy implementation
//   - An error if the strategy name is not recognized
func NewChunkStrategy(name string, tok tokenizer.Tokenizer) (ChunkStrategy, error) {
	switch name {
	case "fixed":
		return NewFixedSizeStrategy(tok), nil
	case "headers", "markdown-headers":
		return NewMarkdownHeaderStrategy(tok), nil
	case "recursive":
		return NewRecursiveStrategy(tok), nil
	case "semantic":
		return NewSemanticStrategy(tok), nil
	default:
		return nil, fmt.Errorf("unknown chunking strategy: %s (supported: fixed, headers, recursive, semantic)", name)
	}
}

// AvailableStrategies returns a list of all available chunking strategy names.
func AvailableStrategies() []string {
	return []string{
		"fixed",
		"headers",
		"markdown-headers", // alias for "headers"
		"recursive",
		"semantic",
	}
}

// DefaultStrategy returns the default chunking strategy name.
func DefaultStrategy() string {
	return "fixed"
}

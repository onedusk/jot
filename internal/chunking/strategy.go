// Package chunking provides pluggable strategies for splitting documents into chunks.
package chunking

import (
	"github.com/onedusk/jot/internal/export"
	"github.com/onedusk/jot/internal/scanner"
)

// ChunkStrategy defines the interface for document chunking strategies.
// Different strategies can implement various approaches to splitting documents,
// such as fixed-size chunks, markdown header boundaries, recursive splitting, or semantic boundaries.
type ChunkStrategy interface {
	// Chunk splits a document into smaller chunks based on the strategy's implementation.
	// Parameters:
	//   - doc: The document to chunk
	//   - maxTokens: Maximum number of tokens per chunk
	//   - overlapTokens: Number of tokens to overlap between consecutive chunks
	// Returns:
	//   - A slice of chunks
	//   - An error if the chunking operation fails
	Chunk(doc scanner.Document, maxTokens, overlapTokens int) ([]export.Chunk, error)
}

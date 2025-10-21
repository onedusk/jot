// Package export provides functionality for exporting documents to various formats like JSON, YAML,
// and a special format optimized for Large Language Models (LLMs).
package export

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/onedusk/jot/internal/chunking"
	"github.com/onedusk/jot/internal/scanner"
)

// JSONLExporter handles exporting documents to JSONL (JSON Lines) format.
// JSONL is a line-delimited JSON format where each line is a valid JSON object,
// commonly used for vector database ingestion (Pinecone, Weaviate, Qdrant).
//
// Specification: https://jsonlines.org/
type JSONLExporter struct {
	// Fields can be added here for configuration if needed in the future
}

// NewJSONLExporter creates and returns a new JSONLExporter instance.
func NewJSONLExporter() *JSONLExporter {
	return &JSONLExporter{}
}

// ToJSONL converts documents to JSONL format using the provided chunking strategy.
// Each chunk is exported as a single line of compact JSON followed by a newline character.
//
// Parameters:
//   - documents: The documents to export
//   - chunker: The chunking strategy to use for splitting documents
//   - maxTokens: Maximum number of tokens per chunk
//   - overlapTokens: Number of tokens to overlap between consecutive chunks
//
// Returns:
//   - A string containing the JSONL output (newline-delimited JSON objects)
//   - An error if chunking or JSON marshaling fails
func (e *JSONLExporter) ToJSONL(documents []scanner.Document, chunker chunking.ChunkStrategy, maxTokens, overlapTokens int) (string, error) {
	if chunker == nil {
		return "", fmt.Errorf("chunker cannot be nil")
	}

	var builder strings.Builder

	for _, doc := range documents {
		// Chunk the document using the provided strategy
		chunks, err := chunker.Chunk(doc, maxTokens, overlapTokens)
		if err != nil {
			return "", fmt.Errorf("failed to chunk document %s: %w", doc.ID, err)
		}

		// Convert each chunk to ChunkMetadata and write as JSONL
		for i, chunk := range chunks {
			metadata := ChunkMetadata{
				DocID:      doc.ID,
				ChunkID:    chunk.ID,
				Text:       chunk.Text,
				TokenCount: chunk.TokenCount,
				Source:     doc.RelativePath,
				StartPos:   chunk.StartPos,
				EndPos:     chunk.EndPos,
				Vector:     chunk.Vector,
			}

			// Set previous and next chunk IDs for navigation
			if i > 0 {
				metadata.PrevChunkID = chunks[i-1].ID
			}
			if i < len(chunks)-1 {
				metadata.NextChunkID = chunks[i+1].ID
			}

			// Marshal to compact JSON (no indentation)
			jsonBytes, err := json.Marshal(metadata)
			if err != nil {
				return "", fmt.Errorf("failed to marshal chunk %s to JSON: %w", chunk.ID, err)
			}

			// Append JSON object followed by newline (JSONL spec)
			builder.Write(jsonBytes)
			builder.WriteString("\n")
		}
	}

	return builder.String(), nil
}

package chunking

import (
	"fmt"
	"strings"

	"github.com/onedusk/jot/internal/export"
	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/tokenizer"
)

// RecursiveStrategy implements hierarchical text splitting using multiple separators.
// It tries each separator in order (paragraph, line, space, character) until chunk size is met.
type RecursiveStrategy struct {
	tokenizer  tokenizer.Tokenizer
	separators []string
}

// NewRecursiveStrategy creates a new RecursiveStrategy with the given tokenizer.
// Uses hierarchical separators: double newline, newline, space, empty string.
func NewRecursiveStrategy(tok tokenizer.Tokenizer) *RecursiveStrategy {
	return &RecursiveStrategy{
		tokenizer:  tok,
		separators: []string{"\n\n", "\n", " ", ""},
	}
}

// Chunk implements the ChunkStrategy interface for recursive chunking.
func (s *RecursiveStrategy) Chunk(doc scanner.Document, maxTokens, overlapTokens int) ([]export.Chunk, error) {
	content := string(doc.Content)

	// Check if entire content fits within token limit
	if s.tokenizer.Count(content) <= maxTokens {
		return []export.Chunk{
			{
				ID:         fmt.Sprintf("%s-chunk-0", doc.ID),
				Text:       content,
				StartPos:   0,
				EndPos:     len(content),
				TokenCount: s.tokenizer.Count(content),
			},
		}, nil
	}

	// Recursively split using separators
	chunks := make([]export.Chunk, 0)
	chunkID := 0
	s.recursiveSplit(content, 0, maxTokens, &chunks, &chunkID, doc.ID, 0)

	return chunks, nil
}

// recursiveSplit recursively splits text using hierarchical separators.
func (s *RecursiveStrategy) recursiveSplit(text string, offset int, maxTokens int, chunks *[]export.Chunk, chunkID *int, docID string, depth int) {
	// If text fits, create chunk
	tokenCount := s.tokenizer.Count(text)
	if tokenCount <= maxTokens {
		*chunks = append(*chunks, export.Chunk{
			ID:         fmt.Sprintf("%s-chunk-%d", docID, *chunkID),
			Text:       text,
			StartPos:   offset,
			EndPos:     offset + len(text),
			TokenCount: tokenCount,
		})
		*chunkID++
		return
	}

	// Try each separator in order
	if depth < len(s.separators) {
		separator := s.separators[depth]

		if separator == "" {
			// Last resort: character-level splitting
			// Use binary search to find split point
			left, right := 0, len(text)
			for left < right {
				mid := (left + right + 1) / 2
				if s.tokenizer.Count(text[:mid]) <= maxTokens {
					left = mid
				} else {
					right = mid - 1
				}
			}

			if left > 0 {
				// Split at character boundary
				s.recursiveSplit(text[:left], offset, maxTokens, chunks, chunkID, docID, depth)
				s.recursiveSplit(text[left:], offset+left, maxTokens, chunks, chunkID, docID, 0)
			}
			return
		}

		// Split by separator
		parts := strings.Split(text, separator)
		if len(parts) > 1 {
			// Build chunks by combining parts until token limit
			currentPart := ""
			currentOffset := offset

			for i, part := range parts {
				testText := currentPart
				if testText != "" {
					testText += separator
				}
				testText += part

				if s.tokenizer.Count(testText) <= maxTokens || currentPart == "" {
					// Add this part to current chunk
					if currentPart != "" {
						currentPart += separator
					}
					currentPart += part
				} else {
					// Current part would exceed limit, save what we have and start new
					s.recursiveSplit(currentPart, currentOffset, maxTokens, chunks, chunkID, docID, depth+1)
					currentOffset += len(currentPart) + len(separator)
					currentPart = part
				}

				// If last part, process remaining
				if i == len(parts)-1 && currentPart != "" {
					s.recursiveSplit(currentPart, currentOffset, maxTokens, chunks, chunkID, docID, depth+1)
				}
			}
			return
		}
	}

	// If no separator worked, try next level
	if depth+1 < len(s.separators) {
		s.recursiveSplit(text, offset, maxTokens, chunks, chunkID, docID, depth+1)
	} else {
		// Fallback: force split
		left, right := 0, len(text)
		for left < right {
			mid := (left + right + 1) / 2
			if s.tokenizer.Count(text[:mid]) <= maxTokens {
				left = mid
			} else {
				right = mid - 1
			}
		}

		if left > 0 && left < len(text) {
			s.recursiveSplit(text[:left], offset, maxTokens, chunks, chunkID, docID, 0)
			s.recursiveSplit(text[left:], offset+left, maxTokens, chunks, chunkID, docID, 0)
		}
	}
}

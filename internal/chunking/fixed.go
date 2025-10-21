package chunking

import (
	"fmt"

	"github.com/onedusk/jot/internal/export"
	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/tokenizer"
)

// FixedSizeStrategy implements token-based fixed-size chunking.
// It splits documents into chunks of approximately maxTokens size with specified overlap.
type FixedSizeStrategy struct {
	tokenizer tokenizer.Tokenizer
}

// NewFixedSizeStrategy creates a new FixedSizeStrategy with the given tokenizer.
func NewFixedSizeStrategy(tok tokenizer.Tokenizer) *FixedSizeStrategy {
	return &FixedSizeStrategy{
		tokenizer: tok,
	}
}

// Chunk implements the ChunkStrategy interface for fixed-size chunking.
func (s *FixedSizeStrategy) Chunk(doc scanner.Document, maxTokens, overlapTokens int) ([]export.Chunk, error) {
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

	chunks := make([]export.Chunk, 0)
	chunkID := 0
	startPos := 0

	for startPos < len(content) {
		// Find end position where token count reaches maxTokens
		endPos := len(content)
		currentText := content[startPos:endPos]

		// If current text exceeds maxTokens, find the right boundary
		if s.tokenizer.Count(currentText) > maxTokens {
			// Binary search for the right character position
			left, right := startPos, len(content)

			for left < right {
				mid := (left + right + 1) / 2
				testText := content[startPos:mid]

				if s.tokenizer.Count(testText) <= maxTokens {
					left = mid
				} else {
					right = mid - 1
				}
			}

			endPos = left

			// Try to break at word boundary to avoid splitting words
			if endPos < len(content) && endPos > startPos {
				// Look backwards for space or newline within reasonable range
				searchStart := maxInt(startPos, endPos-100)
				for i := endPos; i > searchStart; i-- {
					if content[i-1] == ' ' || content[i-1] == '\n' {
						endPos = i
						break
					}
				}
			}
		}

		// Extract chunk text
		chunkText := content[startPos:endPos]

		chunks = append(chunks, export.Chunk{
			ID:         fmt.Sprintf("%s-chunk-%d", doc.ID, chunkID),
			Text:       chunkText,
			StartPos:   startPos,
			EndPos:     endPos,
			TokenCount: s.tokenizer.Count(chunkText),
		})

		chunkID++

		// Move to next chunk with overlap
		if endPos >= len(content) {
			break
		}

		// Calculate next start position with token-based overlap
		if overlapTokens > 0 {
			// Binary search for overlap position
			targetTokens := s.tokenizer.Count(chunkText) - overlapTokens
			if targetTokens <= 0 {
				// If overlap is larger than chunk, just move forward minimally
				startPos = endPos
				continue
			}

			left, right := startPos, endPos
			for left < right {
				mid := (left + right + 1) / 2
				testText := content[startPos:mid]

				if s.tokenizer.Count(testText) <= targetTokens {
					left = mid
				} else {
					right = mid - 1
				}
			}

			startPos = left
		} else {
			startPos = endPos
		}

		// Ensure we make progress
		if startPos >= endPos {
			startPos = endPos
		}
	}

	return chunks, nil
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

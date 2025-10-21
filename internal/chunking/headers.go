package chunking

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/onedusk/jot/internal/export"
	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/tokenizer"
)

// MarkdownHeaderStrategy implements chunking based on markdown header boundaries.
// It splits documents at markdown headers (# to ######) while respecting token limits.
type MarkdownHeaderStrategy struct {
	tokenizer tokenizer.Tokenizer
	// headerRegex matches markdown headers: ^#{1,6}\s+(.+)$
	headerRegex *regexp.Regexp
}

// NewMarkdownHeaderStrategy creates a new MarkdownHeaderStrategy with the given tokenizer.
func NewMarkdownHeaderStrategy(tok tokenizer.Tokenizer) *MarkdownHeaderStrategy {
	return &MarkdownHeaderStrategy{
		tokenizer:   tok,
		headerRegex: regexp.MustCompile(`^#{1,6}\s+(.+)$`),
	}
}

// Chunk implements the ChunkStrategy interface for markdown header-based chunking.
func (s *MarkdownHeaderStrategy) Chunk(doc scanner.Document, maxTokens, overlapTokens int) ([]export.Chunk, error) {
	content := string(doc.Content)
	lines := strings.Split(content, "\n")

	// Find header boundaries
	sections := make([]struct {
		startLine int
		endLine   int
		text      string
	}, 0)

	currentStart := 0
	currentLines := make([]string, 0)

	for i, line := range lines {
		if s.headerRegex.MatchString(line) && i > 0 {
			// Found a header, save previous section
			if len(currentLines) > 0 {
				sections = append(sections, struct {
					startLine int
					endLine   int
					text      string
				}{
					startLine: currentStart,
					endLine:   i - 1,
					text:      strings.Join(currentLines, "\n"),
				})
			}
			currentStart = i
			currentLines = []string{line}
		} else {
			currentLines = append(currentLines, line)
		}
	}

	// Add final section
	if len(currentLines) > 0 {
		sections = append(sections, struct {
			startLine int
			endLine   int
			text      string
		}{
			startLine: currentStart,
			endLine:   len(lines) - 1,
			text:      strings.Join(currentLines, "\n"),
		})
	}

	// Convert sections to chunks, splitting if they exceed maxTokens
	chunks := make([]export.Chunk, 0)
	chunkID := 0
	charOffset := 0

	for _, section := range sections {
		sectionTokens := s.tokenizer.Count(section.text)

		if sectionTokens <= maxTokens {
			// Section fits within token limit
			chunks = append(chunks, export.Chunk{
				ID:         fmt.Sprintf("%s-chunk-%d", doc.ID, chunkID),
				Text:       section.text,
				StartPos:   charOffset,
				EndPos:     charOffset + len(section.text),
				TokenCount: sectionTokens,
			})
			chunkID++
			charOffset += len(section.text) + 1 // +1 for newline
		} else {
			// Section exceeds limit, fall back to fixed-size chunking for this section
			fixedStrategy := NewFixedSizeStrategy(s.tokenizer)
			tempDoc := scanner.Document{
				ID:      doc.ID,
				Content: []byte(section.text),
			}
			sectionChunks, err := fixedStrategy.Chunk(tempDoc, maxTokens, overlapTokens)
			if err != nil {
				return nil, err
			}

			// Adjust chunk IDs and positions
			for _, chunk := range sectionChunks {
				chunk.ID = fmt.Sprintf("%s-chunk-%d", doc.ID, chunkID)
				chunk.StartPos += charOffset
				chunk.EndPos += charOffset
				chunks = append(chunks, chunk)
				chunkID++
			}
			charOffset += len(section.text) + 1
		}
	}

	return chunks, nil
}

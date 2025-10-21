// Package export provides functionality for exporting documents to various formats like JSON, YAML,
// and a special format optimized for Large Language Models (LLMs).
package export

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/tokenizer"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Exporter handles the conversion of scanned documents into different data formats.
type Exporter struct {
	// Configuration can be added here in the future, e.g., for controlling output details.
}

// NewExporter creates and returns a new Exporter instance.
func NewExporter() *Exporter {
	return &Exporter{}
}

// ToJSON exports a slice of documents to a JSON formatted string.
func (e *Exporter) ToJSON(documents []scanner.Document) (string, error) {
	export := e.createExportData(documents)

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return string(data), nil
}

// ToYAML exports a slice of documents to a YAML formatted string.
func (e *Exporter) ToYAML(documents []scanner.Document) (string, error) {
	export := e.createExportData(documents)

	data, err := yaml.Marshal(export)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(data), nil
}

// ToLLMFormat exports documents to a structure optimized for consumption by Large Language Models.
// This format includes chunking, sectioning, and metadata extraction.
func (e *Exporter) ToLLMFormat(documents []scanner.Document) (*LLMExport, error) {
	// Initialize tokenizer for accurate token-based chunking
	tok, err := tokenizer.NewTokenizer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	// Read chunking configuration from viper with sensible defaults
	chunkSize := viper.GetInt("llm.chunk_size")
	if chunkSize == 0 {
		chunkSize = 512
	}

	overlap := viper.GetInt("llm.overlap")
	if overlap == 0 {
		overlap = 128
	}

	export := &LLMExport{
		Version:   "1.0",
		Generated: time.Now().Format(time.RFC3339),
		Documents: make([]LLMDocument, 0, len(documents)),
		Index: &SemanticIndex{
			Keywords: make(map[string][]string),
			Concepts: make([]string, 0),
		},
	}

	for _, doc := range documents {
		llmDoc := LLMDocument{
			ID:       doc.ID,
			Title:    doc.Title,
			Path:     doc.RelativePath,
			Content:  string(doc.Content),
			Chunks:   chunkDocument(doc, chunkSize, overlap, tok),
			Metadata: doc.Metadata,
		}

		// Extract sections
		for _, section := range doc.Sections {
			llmDoc.Sections = append(llmDoc.Sections, LLMSection{
				ID:        section.ID,
				Title:     section.Title,
				Level:     section.Level,
				Content:   section.Content,
				StartLine: section.StartLine,
				EndLine:   section.EndLine,
			})
		}

		// Extract code blocks
		for _, block := range doc.CodeBlocks {
			llmDoc.CodeBlocks = append(llmDoc.CodeBlocks, LLMCodeBlock{
				Language:  block.Language,
				Content:   block.Content,
				StartLine: block.StartLine,
			})
		}

		// Extract links
		for _, link := range doc.Links {
			if link.IsInternal {
				llmDoc.Links.Internal = append(llmDoc.Links.Internal, link.URL)
			} else {
				llmDoc.Links.External = append(llmDoc.Links.External, link.URL)
			}
		}

		export.Documents = append(export.Documents, llmDoc)

		// Build index
		e.indexDocument(&llmDoc, export.Index)
	}

	return export, nil
}

// createExportData transforms a slice of scanner.Document into a generic map structure
// suitable for JSON or YAML serialization.
func (e *Exporter) createExportData(documents []scanner.Document) map[string]interface{} {
	docs := make([]map[string]interface{}, 0, len(documents))

	for _, doc := range documents {
		docData := map[string]interface{}{
			"id":       doc.ID,
			"path":     doc.RelativePath,
			"title":    doc.Title,
			"content":  string(doc.Content),
			"metadata": doc.Metadata,
			"modified": doc.ModTime.Format(time.RFC3339),
		}

		// Add sections
		if len(doc.Sections) > 0 {
			sections := make([]map[string]interface{}, 0, len(doc.Sections))
			for _, section := range doc.Sections {
				sections = append(sections, map[string]interface{}{
					"id":         section.ID,
					"title":      section.Title,
					"level":      section.Level,
					"content":    section.Content,
					"start_line": section.StartLine,
					"end_line":   section.EndLine,
				})
			}
			docData["sections"] = sections
		}

		// Add code blocks
		if len(doc.CodeBlocks) > 0 {
			blocks := make([]map[string]interface{}, 0, len(doc.CodeBlocks))
			for _, block := range doc.CodeBlocks {
				blocks = append(blocks, map[string]interface{}{
					"language":   block.Language,
					"content":    block.Content,
					"start_line": block.StartLine,
				})
			}
			docData["code_blocks"] = blocks
		}

		// Add links
		if len(doc.Links) > 0 {
			internal := make([]string, 0)
			external := make([]string, 0)
			for _, link := range doc.Links {
				if link.IsInternal {
					internal = append(internal, link.URL)
				} else {
					external = append(external, link.URL)
				}
			}
			docData["links"] = map[string]interface{}{
				"internal": internal,
				"external": external,
			}
		}

		docs = append(docs, docData)
	}

	return map[string]interface{}{
		"version":   "1.0",
		"generated": time.Now().Format(time.RFC3339),
		"documents": docs,
	}
}

// indexDocument builds a simple semantic index for a document by extracting keywords
// and concepts, and adds them to the provided SemanticIndex.
func (e *Exporter) indexDocument(doc *LLMDocument, index *SemanticIndex) {
	// Extract keywords from title and content
	keywords := extractKeywords(doc.Title + " " + doc.Content)
	for _, keyword := range keywords {
		if _, exists := index.Keywords[keyword]; !exists {
			index.Keywords[keyword] = make([]string, 0)
		}
		index.Keywords[keyword] = append(index.Keywords[keyword], doc.ID)
	}

	// Extract concepts from sections
	for _, section := range doc.Sections {
		concept := strings.ToLower(section.Title)
		if !contains(index.Concepts, concept) {
			index.Concepts = append(index.Concepts, concept)
		}
	}
}

// chunkDocument splits a document's content into smaller, potentially overlapping chunks.
// This is useful for processing large documents with token limits.
// Uses token-based chunking instead of character-based for accurate LLM context window management.
func chunkDocument(doc scanner.Document, maxTokens, overlapTokens int, tok tokenizer.Tokenizer) []Chunk {
	content := string(doc.Content)

	// Check if entire content fits within token limit
	if tok.Count(content) <= maxTokens {
		return []Chunk{
			{
				ID:         "chunk-0",
				Text:       content,
				StartPos:   0,
				EndPos:     len(content),
				TokenCount: tok.Count(content),
			},
		}
	}

	chunks := make([]Chunk, 0)
	chunkID := 0
	startPos := 0

	for startPos < len(content) {
		// Find end position where token count reaches maxTokens
		endPos := len(content)
		currentText := content[startPos:endPos]

		// If current text exceeds maxTokens, find the right boundary
		if tok.Count(currentText) > maxTokens {
			// Binary search for the right character position
			left, right := startPos, len(content)

			for left < right {
				mid := (left + right + 1) / 2
				testText := content[startPos:mid]

				if tok.Count(testText) <= maxTokens {
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

		chunks = append(chunks, Chunk{
			ID:         fmt.Sprintf("chunk-%d", chunkID),
			Text:       chunkText,
			StartPos:   startPos,
			EndPos:     endPos,
			TokenCount: tok.Count(chunkText),
		})

		chunkID++

		// Move to next chunk with overlap
		if endPos >= len(content) {
			break
		}

		// Calculate next start position with token-based overlap
		// Find position that gives us approximately overlapTokens
		if overlapTokens > 0 {
			// Binary search for overlap position
			targetTokens := tok.Count(chunkText) - overlapTokens
			if targetTokens <= 0 {
				// If overlap is larger than chunk, just move forward minimally
				nextStart := endPos
				startPos = nextStart
				continue
			}

			left, right := startPos, endPos
			for left < right {
				mid := (left + right + 1) / 2
				testText := content[startPos:mid]

				if tok.Count(testText) <= targetTokens {
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

	return chunks
}

// extractKeywords performs a simple keyword extraction from a given text.
// It removes common stop words and punctuation.
func extractKeywords(text string) []string {
	// Simple keyword extraction - in production, use NLP library
	words := strings.Fields(strings.ToLower(text))
	keywords := make(map[string]bool)

	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true,
		"this": true, "that": true, "these": true, "those": true,
	}

	for _, word := range words {
		// Clean word
		word = strings.Trim(word, ".,!?;:\"'()[]{}#")

		// Skip short words and stop words
		if len(word) > 3 && !stopWords[word] {
			keywords[word] = true
		}
	}

	result := make([]string, 0, len(keywords))
	for keyword := range keywords {
		result = append(result, keyword)
	}

	return result
}

// contains checks if a string slice contains a specific item.
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// maxInt returns the greater of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// minInt returns the smaller of two integers.
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Package export provides functionality for exporting documents to various formats like JSON, YAML,
// and a special format optimized for Large Language Models (LLMs).
package export

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/onedusk/jot/internal/scanner"
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
			Chunks:   chunkDocument(doc, 512, 128),
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
func chunkDocument(doc scanner.Document, maxSize, overlap int) []Chunk {
	content := string(doc.Content)
	if len(content) <= maxSize {
		return []Chunk{
			{
				ID:       "chunk-0",
				Text:     content,
				StartPos: 0,
				EndPos:   len(content),
			},
		}
	}

	chunks := make([]Chunk, 0)
	chunkID := 0
	startPos := 0

	for startPos < len(content) {
		// Calculate end position for this chunk
		endPos := minInt(startPos+maxSize, len(content))

		// Try to break at word boundary if not at end
		if endPos < len(content) && endPos > startPos {
			// Look for last space before maxSize
			for i := endPos; i > startPos && i > endPos-50; i-- {
				if content[i-1] == ' ' || content[i-1] == '\n' {
					endPos = i
					break
				}
			}
		}

		// Extract chunk text
		chunkText := content[startPos:endPos]

		chunks = append(chunks, Chunk{
			ID:       fmt.Sprintf("chunk-%d", chunkID),
			Text:     chunkText,
			StartPos: startPos,
			EndPos:   endPos,
		})

		chunkID++

		// Move to next chunk with overlap
		if endPos >= len(content) {
			break
		}

		// Calculate next start position with overlap
		nextStart := endPos - overlap
		if nextStart <= startPos {
			// Ensure we make progress
			nextStart = startPos + (maxSize - overlap)
			if nextStart <= startPos {
				nextStart = endPos
			}
		}
		startPos = nextStart
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

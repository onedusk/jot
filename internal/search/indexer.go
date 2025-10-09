// Package search provides functionality for creating and managing a search index
// for the generated documentation.
package search

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/thrive/jot/internal/scanner"
)

// Index represents the top-level structure of the search index. It contains a list
// of all indexed documents and a version number for the index format.
type Index struct {
	Documents []IndexDocument `json:"documents"`
	Version   string          `json:"version"`
}

// IndexDocument represents a single document within the search index. It includes
// the document's content and metadata, optimized for efficient searching.
type IndexDocument struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Path        string   `json:"path"`
	Content     string   `json:"content"`
	Headings    []string `json:"headings"`
	Keywords    []string `json:"keywords"`
	Summary     string   `json:"summary"`
	Modified    string   `json:"modified,omitempty"`
	WordCount   int      `json:"wordCount,omitempty"`
	ReadTime    string   `json:"readTime,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	ContentHash string   `json:"contentHash,omitempty"`
}

// Indexer is responsible for building a search index from a collection of documents.
type Indexer struct {
	outputPath string
}

// NewIndexer creates a new Indexer that will write its output to the specified path.
func NewIndexer(outputPath string) *Indexer {
	return &Indexer{
		outputPath: outputPath,
	}
}

// BuildIndex processes a slice of documents and creates a search index.
func (idx *Indexer) BuildIndex(documents []scanner.Document) (*Index, error) {
	index := &Index{
		Version:   "1.0",
		Documents: make([]IndexDocument, 0, len(documents)),
	}

	for _, doc := range documents {
		indexDoc := idx.processDocument(doc)
		index.Documents = append(index.Documents, indexDoc)
	}

	return index, nil
}

// processDocument converts a scanner.Document into an IndexDocument, extracting
// and cleaning data to make it suitable for indexing.
func (idx *Indexer) processDocument(doc scanner.Document) IndexDocument {
	content := string(doc.Content)

	// Extract headings
	headings := idx.extractHeadings(content)

	// Extract keywords (simple approach - can be enhanced)
	keywords := idx.extractKeywords(content)

	// Clean content for search (remove markdown syntax)
	cleanContent := idx.cleanContent(content)

	// Generate document ID from path
	id := strings.ReplaceAll(doc.RelativePath, "/", "-")
	id = strings.ReplaceAll(id, ".md", "")

	// Calculate word count
	wordCount := len(strings.Fields(content))

	// Calculate read time
	readMinutes := wordCount / 250
	if readMinutes < 1 {
		readMinutes = 1
	}
	readTime := fmt.Sprintf("%dmin", readMinutes)

	// Generate summary (first 200 chars)
	summary := cleanContent
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}

	// Extract tags from frontmatter
	var tags []string
	if doc.Metadata != nil {
		if t, ok := doc.Metadata["tags"].([]string); ok {
			tags = t
		} else if t, ok := doc.Metadata["tags"].(string); ok {
			tags = strings.Split(t, ",")
			for i := range tags {
				tags[i] = strings.TrimSpace(tags[i])
			}
		}
	}

	return IndexDocument{
		ID:        id,
		Title:     doc.Title,
		Path:      strings.Replace(doc.RelativePath, ".md", ".html", 1),
		Content:   cleanContent,
		Headings:  headings,
		Keywords:  keywords,
		Summary:   summary,
		Modified:  doc.ModTime.Format("2006-01-02T15:04:05Z"),
		WordCount: wordCount,
		ReadTime:  readTime,
		Tags:      tags,
	}
}

// extractHeadings parses markdown content and returns a slice of all headings.
func (idx *Indexer) extractHeadings(content string) []string {
	headingRegex := regexp.MustCompile(`^#{1,6}\s+(.+)$`)
	lines := strings.Split(content, "\n")

	var headings []string
	for _, line := range lines {
		if matches := headingRegex.FindStringSubmatch(line); len(matches) > 1 {
			heading := strings.TrimSpace(matches[1])
			headings = append(headings, heading)
		}
	}

	return headings
}

// extractKeywords performs a simple keyword extraction from content by counting
// word frequencies and selecting the most common, non-trivial words.
func (idx *Indexer) extractKeywords(content string) []string {
	// Remove markdown syntax
	clean := idx.cleanContent(content)

	// Split into words
	words := strings.Fields(clean)

	// Count word frequency
	wordCount := make(map[string]int)
	for _, word := range words {
		word = strings.ToLower(word)
		word = strings.Trim(word, ".,!?;:'\"")

		// Skip common words and short words
		if len(word) < 4 || isCommonWord(word) {
			continue
		}

		wordCount[word]++
	}

	// Get top keywords
	var keywords []string
	for word, count := range wordCount {
		if count >= 3 { // Word appears at least 3 times
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// cleanContent strips markdown syntax from a string, leaving plain text.
func (idx *Indexer) cleanContent(content string) string {
	// Remove code blocks
	codeBlockRegex := regexp.MustCompile("```[\\s\\S]*?```")
	content = codeBlockRegex.ReplaceAllString(content, "")

	// Remove inline code
	inlineCodeRegex := regexp.MustCompile("`[^`]+`")
	content = inlineCodeRegex.ReplaceAllString(content, "")

	// Remove images
	imageRegex := regexp.MustCompile(`!\[.*?\]\(.*?\)`)
	content = imageRegex.ReplaceAllString(content, "")

	// Remove links but keep text
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\([^\)]+\)`)
	content = linkRegex.ReplaceAllString(content, "$1")

	// Remove headers
	headerRegex := regexp.MustCompile(`^#{1,6}\s+`)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = headerRegex.ReplaceAllString(line, "")
	}
	content = strings.Join(lines, "\n")

	// Remove emphasis
	content = strings.ReplaceAll(content, "**", "")
	content = strings.ReplaceAll(content, "__", "")
	content = strings.ReplaceAll(content, "*", "")
	content = strings.ReplaceAll(content, "_", "")

	// Remove extra whitespace
	spaceRegex := regexp.MustCompile(`\s+`)
	content = spaceRegex.ReplaceAllString(content, " ")

	return strings.TrimSpace(content)
}

// SaveIndex serializes the search index to a JSON file in the specified output path.
func (idx *Indexer) SaveIndex(index *Index) error {
	// Ensure output directory exists
	outputDir := filepath.Join(idx.outputPath, "assets")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Marshal index to JSON
	data, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	// Write to file
	indexPath := filepath.Join(outputDir, "search-index.json")
	if err := os.WriteFile(indexPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// isCommonWord checks if a word is a common English stop word.
func isCommonWord(word string) bool {
	commonWords := map[string]bool{
		"the": true, "and": true, "for": true, "are": true,
		"with": true, "this": true, "that": true, "from": true,
		"have": true, "been": true, "were": true, "will": true,
		"your": true, "their": true, "what": true, "when": true,
		"where": true, "which": true, "these": true, "those": true,
		"there": true, "then": true, "than": true, "both": true,
		"each": true, "some": true, "such": true, "only": true,
		"very": true, "just": true, "most": true, "also": true,
		"into": true, "over": true, "after": true, "before": true,
	}

	return commonWords[word]
}

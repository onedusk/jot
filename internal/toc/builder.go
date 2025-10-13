// Package toc provides functionality for building a Table of Contents (TOC)
// from a collection of documents. It organizes documents into a hierarchical
// structure based on their file paths.
package toc

import (
	"crypto/sha256"
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/onedusk/jot/internal/scanner"
)

// Builder is responsible for constructing a TableOfContents from a slice of documents.
type Builder struct {
	// Configuration options for the builder can be added here in the future.
}

// NewBuilder creates and returns a new TOC Builder.
func NewBuilder() *Builder {
	return &Builder{}
}

// Build constructs a hierarchical TableOfContents from a flat list of documents.
// It sorts the documents, creates a tree structure based on file paths, and
// enriches the nodes with metadata.
func (b *Builder) Build(documents []scanner.Document) *TableOfContents {
	// Sort documents by path for consistent ordering
	b.sortDocuments(documents)

	// Create root node
	root := &TOCNode{
		ID:       "root",
		Title:    "Table of Contents",
		Children: make([]*TOCNode, 0),
	}

	// Build the tree
	for _, doc := range documents {
		b.addDocumentToTree(root, doc)
	}

	// Create TOC with index
	toc := &TableOfContents{
		Version: "1.0",
		Root:    root,
	}
	toc.buildIndex()

	return toc
}

// addDocumentToTree adds a single document to the TOC tree, creating parent
// directory nodes as needed.
func (b *Builder) addDocumentToTree(root *TOCNode, doc scanner.Document) {
	// Split path into parts
	parts := strings.Split(filepath.ToSlash(doc.RelativePath), "/")

	currentNode := root
	pathParts := make([]string, 0)

	// Navigate/create the tree structure
	for i, part := range parts {
		isFile := (i == len(parts)-1)

		if isFile {
			// This is the document file
			pathParts = append(pathParts, strings.TrimSuffix(part, ".md"))
			child := &TOCNode{
				ID:       generateNodeID(pathParts),
				Title:    doc.Title,
				Path:     doc.RelativePath,
				Metadata: b.extractMetadata(doc),
			}
			currentNode.AddChild(child)
		} else {
			// This is a directory
			pathParts = append(pathParts, part)

			// Look for existing child
			child := currentNode.FindChildByTitle(humanizeTitle(part))
			if child == nil {
				// Create new directory node
				child = &TOCNode{
					ID:       generateNodeID(pathParts),
					Title:    humanizeTitle(part),
					Children: make([]*TOCNode, 0),
				}
				currentNode.AddChild(child)
			}
			currentNode = child
		}
	}
}

// sortDocuments sorts a slice of documents alphabetically by their relative path.
func (b *Builder) sortDocuments(docs []scanner.Document) {
	sort.Slice(docs, func(i, j int) bool {
		return docs[i].RelativePath < docs[j].RelativePath
	})
}

// generateNodeID creates a unique and URL-friendly ID for a TOC node from its path parts.
func generateNodeID(parts []string) string {
	cleanParts := make([]string, len(parts))
	for i, part := range parts {
		// Convert to lowercase and replace non-alphanumeric with hyphens
		clean := strings.ToLower(part)
		clean = regexp.MustCompile(`[^a-z0-9]+`).ReplaceAllString(clean, "-")
		clean = strings.Trim(clean, "-")
		cleanParts[i] = clean
	}
	return strings.Join(cleanParts, "-")
}

// humanizeTitle converts a file or directory name (e.g., "some-directory") into a
// human-readable title (e.g., "Some Directory").
func humanizeTitle(name string) string {
	// Replace hyphens and underscores with spaces
	title := strings.ReplaceAll(name, "-", " ")
	title = strings.ReplaceAll(title, "_", " ")

	// Capitalize words
	words := strings.Fields(title)
	for i, word := range words {
		if len(word) > 0 {
			words[i] = strings.ToUpper(word[:1]) + word[1:]
		}
	}

	return strings.Join(words, " ")
}

// extractMetadata processes a document to extract and compute various metadata
// fields for the corresponding TOC node.
func (b *Builder) extractMetadata(doc scanner.Document) NodeMetadata {
	content := string(doc.Content)

	// Calculate word count
	wordCount := len(strings.Fields(content))

	// Calculate read time (average 250 words per minute)
	readMinutes := wordCount / 250
	if readMinutes < 1 {
		readMinutes = 1
	}
	readTime := fmt.Sprintf("%dmin", readMinutes)

	// Generate content hash (SHA256)
	hash := sha256.Sum256(doc.Content)
	contentHash := fmt.Sprintf("sha256:%x", hash[:8]) // Use first 8 bytes for brevity

	// Extract first 200 characters as summary
	summary := content
	if len(summary) > 200 {
		summary = summary[:200] + "..."
	}
	// Clean summary - remove markdown headers
	summary = regexp.MustCompile(`^#{1,6}\s+`).ReplaceAllString(summary, "")
	summary = strings.ReplaceAll(summary, "\n", " ")
	summary = regexp.MustCompile(`\s+`).ReplaceAllString(summary, " ")
	summary = strings.TrimSpace(summary)

	// Extract keywords (use simple frequency analysis)
	keywords := b.extractKeywords(content)

	// Extract tags from frontmatter if available
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

	return NodeMetadata{
		Modified:    doc.ModTime,
		Size:        int64(len(doc.Content)),
		WordCount:   wordCount,
		ReadTime:    readTime,
		Tags:        tags,
		Summary:     summary,
		Keywords:    keywords,
		ContentHash: contentHash,
	}
}

// extractKeywords performs simple keyword extraction from content.
func (b *Builder) extractKeywords(content string) []string {
	// Remove markdown syntax
	clean := b.cleanContent(content)

	// Split into words
	words := strings.Fields(clean)

	// Count word frequency
	wordCount := make(map[string]int)
	for _, word := range words {
		word = strings.ToLower(word)
		word = strings.Trim(word, ".,!?;:'\"")

		// Skip common words and short words
		if len(word) < 4 || b.isCommonWord(word) {
			continue
		}

		wordCount[word]++
	}

	// Get top keywords (appears at least 2 times)
	var keywords []string
	for word, count := range wordCount {
		if count >= 2 && len(keywords) < 10 { // Limit to 10 keywords
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// cleanContent removes markdown syntax for accurate keyword extraction and summaries.
func (b *Builder) cleanContent(content string) string {
	// Remove code blocks
	codeBlockRegex := regexp.MustCompile("```[\\s\\S]*?```")
	content = codeBlockRegex.ReplaceAllString(content, "")

	// Remove inline code
	inlineCodeRegex := regexp.MustCompile("`[^`]+`")
	content = inlineCodeRegex.ReplaceAllString(content, "")

	// Remove headers but keep text
	headerRegex := regexp.MustCompile(`^#{1,6}\s+`)
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lines[i] = headerRegex.ReplaceAllString(line, "")
	}
	content = strings.Join(lines, "\n")

	return content
}

// isCommonWord checks if a word is a common English stop word.
func (b *Builder) isCommonWord(word string) bool {
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

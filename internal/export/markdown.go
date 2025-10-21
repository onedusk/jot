// Package export provides functionality for exporting documents to various formats like JSON, YAML,
// and a special format optimized for Large Language Models (LLMs).
package export

import (
	"fmt"
	"strings"
	"time"

	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/tokenizer"
	"gopkg.in/yaml.v3"
)

// MarkdownExporter handles the conversion of scanned documents into enriched markdown
// format with YAML frontmatter metadata.
type MarkdownExporter struct {
	tokenizer tokenizer.Tokenizer
}

// NewMarkdownExporter creates and returns a new MarkdownExporter instance.
func NewMarkdownExporter() (*MarkdownExporter, error) {
	tok, err := tokenizer.NewTokenizer()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tokenizer: %w", err)
	}

	return &MarkdownExporter{
		tokenizer: tok,
	}, nil
}

// MarkdownFrontmatter represents the YAML frontmatter metadata for a markdown document.
type MarkdownFrontmatter struct {
	Source     string `yaml:"source"`
	Section    string `yaml:"section"`
	ChunkID    string `yaml:"chunk_id"`
	TokenCount int    `yaml:"token_count"`
	Modified   string `yaml:"modified"`
}

// ToEnrichedMarkdown exports documents to enriched markdown format with YAML frontmatter.
// If separateFiles is true, it returns a map of filenames to content for separate files.
// If separateFiles is false, it returns a single concatenated markdown string.
func (m *MarkdownExporter) ToEnrichedMarkdown(documents []scanner.Document, separateFiles bool) (string, error) {
	if separateFiles {
		// TODO: Implement separate files logic in future iteration
		return "", fmt.Errorf("separate files mode not yet implemented")
	}

	var result strings.Builder

	// Generate table of contents first
	toc := m.generateTableOfContents(documents)
	result.WriteString(toc)
	result.WriteString("\n\n")

	// Process each document
	for i, doc := range documents {
		// Generate YAML frontmatter
		frontmatter := MarkdownFrontmatter{
			Source:     doc.RelativePath,
			ChunkID:    doc.ID,
			TokenCount: m.tokenizer.Count(string(doc.Content)),
			Modified:   doc.ModTime.Format(time.RFC3339),
		}

		// Set section from first section title if available
		if len(doc.Sections) > 0 {
			frontmatter.Section = doc.Sections[0].Title
		} else {
			frontmatter.Section = doc.Title
		}

		// Marshal frontmatter to YAML
		yamlData, err := yaml.Marshal(frontmatter)
		if err != nil {
			return "", fmt.Errorf("failed to marshal frontmatter for %s: %w", doc.RelativePath, err)
		}

		// Write frontmatter with delimiters
		result.WriteString("---\n")
		result.WriteString(string(yamlData))
		result.WriteString("---\n\n")

		// Preserve original markdown content
		result.WriteString(string(doc.Content))

		// Add contextual enrichment (stub for future Anthropic-style enrichment)
		enrichment := m.contextualEnrichment(doc, string(doc.Content))
		if enrichment != "" {
			result.WriteString("\n\n")
			result.WriteString(enrichment)
		}

		// Add separator between documents (except for last one)
		if i < len(documents)-1 {
			result.WriteString("\n\n---\n\n")
		}
	}

	return result.String(), nil
}

// generateTableOfContents creates a markdown table of contents with anchor links.
func (m *MarkdownExporter) generateTableOfContents(documents []scanner.Document) string {
	var toc strings.Builder

	toc.WriteString("## Table of Contents\n\n")

	for _, doc := range documents {
		// Create anchor link from document title
		anchor := strings.ToLower(doc.Title)
		anchor = strings.ReplaceAll(anchor, " ", "-")
		// Remove special characters for valid anchor
		anchor = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
				return r
			}
			return -1
		}, anchor)

		toc.WriteString(fmt.Sprintf("- [%s](#%s)\n", doc.Title, anchor))

		// Add subsections if available
		for _, section := range doc.Sections {
			if section.Level <= 2 { // Only include H1 and H2 in TOC
				indent := strings.Repeat("  ", section.Level-1)
				sectionAnchor := strings.ToLower(section.Title)
				sectionAnchor = strings.ReplaceAll(sectionAnchor, " ", "-")
				sectionAnchor = strings.Map(func(r rune) rune {
					if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
						return r
					}
					return -1
				}, sectionAnchor)

				toc.WriteString(fmt.Sprintf("%s- [%s](#%s)\n", indent, section.Title, sectionAnchor))
			}
		}
	}

	return toc.String()
}

// contextualEnrichment is a placeholder method for optional Anthropic-style context injection.
// This will be implemented in a future iteration to add contextual information as HTML comments
// or other enrichment metadata to improve LLM understanding.
//
// TODO: Implement Anthropic-style contextual retrieval enrichment
// See: https://www.anthropic.com/news/contextual-retrieval
// This should inject relevant context from the broader document collection to help
// LLMs better understand each chunk in isolation.
func (m *MarkdownExporter) contextualEnrichment(doc scanner.Document, fullContext string) string {
	// Stub implementation - returns empty string for now
	// Future implementation will:
	// 1. Analyze document context and related documents
	// 2. Generate contextual hints as HTML comments
	// 3. Inject them at strategic points in the markdown
	// 4. Follow Anthropic's contextual retrieval patterns
	return ""
}

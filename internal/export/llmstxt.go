// Package export provides llms.txt export functionality per llmstxt.org specification.
package export

import (
	"log"
	"path/filepath"
	"sort"
	"strings"

	"github.com/onedusk/jot/internal/scanner"
)

// LLMSTxtExporter handles exporting documents to llms.txt format.
// The llms.txt format creates a simple markdown index of documentation
// optimized for LLM consumption, as specified at https://llmstxt.org/
type LLMSTxtExporter struct {
	// No configuration fields needed currently
}

// NewLLMSTxtExporter creates and returns a new LLMSTxtExporter instance.
func NewLLMSTxtExporter() *LLMSTxtExporter {
	return &LLMSTxtExporter{}
}

// ToLLMSTxt exports documents to llms.txt format per llmstxt.org specification.
// The output includes:
// - H1 header with project name
// - Blockquote with project description
// - H2 section headers grouped by directory
// - Markdown list with [Title](path): description format
func (e *LLMSTxtExporter) ToLLMSTxt(documents []scanner.Document, config ProjectConfig) (string, error) {
	var builder strings.Builder

	// Write H1 header with project name
	builder.WriteString("# ")
	builder.WriteString(config.Name)
	builder.WriteString("\n\n")

	// Write blockquote with project description
	builder.WriteString("> ")
	builder.WriteString(config.Description)
	builder.WriteString("\n\n")

	// Group documents by section (directory)
	grouped := groupDocumentsBySection(documents)

	// Sort sections for consistent output
	sections := make([]string, 0, len(grouped))
	for section := range grouped {
		sections = append(sections, section)
	}
	sort.Strings(sections)

	// Write each section
	for _, section := range sections {
		docs := grouped[section]

		// Write H2 section header
		sectionTitle := section
		if sectionTitle == "." || sectionTitle == "" {
			sectionTitle = "Root"
		}
		builder.WriteString("## ")
		builder.WriteString(sectionTitle)
		builder.WriteString("\n\n")

		// Write document list
		for _, doc := range docs {
			description := extractFirstParagraph(doc.Content)

			// Format: - [Title](path): description
			builder.WriteString("- [")
			builder.WriteString(doc.Title)
			builder.WriteString("](")
			builder.WriteString(doc.RelativePath)
			builder.WriteString("): ")
			builder.WriteString(description)
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}

	return builder.String(), nil
}

// groupDocumentsBySection groups documents by their directory path.
func groupDocumentsBySection(documents []scanner.Document) map[string][]scanner.Document {
	grouped := make(map[string][]scanner.Document)

	for _, doc := range documents {
		section := filepath.Dir(doc.RelativePath)
		grouped[section] = append(grouped[section], doc)
	}

	return grouped
}

// extractFirstParagraph extracts the first non-header paragraph from document content.
// Limits output to 100 characters for concise descriptions.
func extractFirstParagraph(content []byte) string {
	text := string(content)
	lines := strings.Split(text, "\n")

	var paragraph strings.Builder
	foundContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and headers
		if trimmed == "" {
			if foundContent {
				break // End of first paragraph
			}
			continue
		}

		// Skip markdown headers
		if strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Skip code blocks
		if strings.HasPrefix(trimmed, "```") {
			continue
		}

		// Accumulate paragraph text
		if paragraph.Len() > 0 {
			paragraph.WriteString(" ")
		}
		paragraph.WriteString(trimmed)
		foundContent = true

		// Stop if we've collected enough
		if paragraph.Len() >= 100 {
			break
		}
	}

	result := paragraph.String()
	if len(result) > 100 {
		result = result[:97] + "..."
	}

	if result == "" {
		result = "No description available"
	}

	return result
}

// ToLLMSFullTxt exports complete documentation with all content concatenated.
// Creates llms-full.txt format with:
// - H1 header with project name
// - Blockquote with project description
// - All documents concatenated with '---' separators
// - README.md appears first, then sorted alphabetically by path
// - Each document prefixed with H1 heading containing document title
func (e *LLMSTxtExporter) ToLLMSFullTxt(documents []scanner.Document, config ProjectConfig) (string, error) {
	var builder strings.Builder

	// Write H1 header with project name
	builder.WriteString("# ")
	builder.WriteString(config.Name)
	builder.WriteString("\n\n")

	// Write blockquote with project description
	builder.WriteString("> ")
	builder.WriteString(config.Description)
	builder.WriteString("\n\n")

	// Sort documents by importance (README first, then alphabetically)
	sortedDocs := sortDocumentsByImportance(documents)

	// Concatenate all documents
	for i, doc := range sortedDocs {
		// Add separator between documents (but not before the first one)
		if i > 0 {
			builder.WriteString("---\n\n")
		}

		// Add H1 heading with document title
		builder.WriteString("# ")
		builder.WriteString(doc.Title)
		builder.WriteString("\n\n")

		// Add document content (preserve original markdown formatting)
		builder.Write(doc.Content)
		builder.WriteString("\n\n")
	}

	result := builder.String()

	// Estimate size and log warning if > 1MB
	size := estimateSize(result)
	if size > 1048576 {
		log.Printf("Warning: llms-full.txt output size is %d bytes (%.2f MB), which may exceed LLM context limits",
			size, float64(size)/1048576.0)
	}

	return result, nil
}

// sortDocumentsByImportance sorts documents with README.md first, then alphabetically by path.
func sortDocumentsByImportance(documents []scanner.Document) []scanner.Document {
	// Create a copy to avoid modifying the original slice
	sorted := make([]scanner.Document, len(documents))
	copy(sorted, documents)

	sort.Slice(sorted, func(i, j int) bool {
		// README.md always comes first
		iIsReadme := strings.ToLower(filepath.Base(sorted[i].RelativePath)) == "readme.md"
		jIsReadme := strings.ToLower(filepath.Base(sorted[j].RelativePath)) == "readme.md"

		if iIsReadme && !jIsReadme {
			return true
		}
		if !iIsReadme && jIsReadme {
			return false
		}

		// Otherwise sort alphabetically by relative path
		return sorted[i].RelativePath < sorted[j].RelativePath
	})

	return sorted
}

// estimateSize returns the byte count of the content.
func estimateSize(content string) int64 {
	return int64(len(content))
}

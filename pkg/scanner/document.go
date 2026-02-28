// Package scanner provides types and functions for discovering, reading, and parsing
// markdown documents from the filesystem.
package scanner

import (
	"bytes"
	"regexp"
	"strings"
	"time"
)

// Document represents a single parsed markdown file, including its content,
// metadata, and extracted structural elements like sections, links, and code blocks.
type Document struct {
	ID           string                 // Unique identifier (MD5 hash of RelativePath)
	Path         string                 // Absolute file path on the filesystem.
	RelativePath string                 // File path relative to the scanned root directory.
	Title        string                 // The title of the document, extracted from frontmatter or the first H1.
	Content      []byte                 // The raw markdown content of the file, with frontmatter removed.
	HTML         string                 // Rendered HTML content (populated by the renderer).
	Metadata     map[string]interface{} // Key-value data parsed from YAML frontmatter.
	ModTime      time.Time              // The last modification time of the file.
	Sections     []Section              // A slice of sections extracted from the document.
	Links        []Link                 // A slice of links found in the document.
	CodeBlocks   []CodeBlock            // A slice of code blocks found in the document.
}

// Section represents a structural section of a document, typically initiated by a heading.
type Section struct {
	ID        string // A URL-friendly identifier for the section title.
	Title     string // The text of the section's heading.
	Level     int    // The heading level (1-6).
	Content   string // The markdown content within the section, excluding the title.
	StartLine int    // The line number where the section begins.
	EndLine   int    // The line number where the section ends.
}

// Link represents a hyperlink found within a document.
type Link struct {
	Text       string // The anchor text of the link.
	URL        string // The destination URL of the link.
	IsInternal bool   // True if the link points to a relative path without a scheme.
}

// CodeBlock represents a fenced code block within a document.
type CodeBlock struct {
	Language  string // The language identifier (e.g., "go", "typescript").
	Content   string // The raw source code within the block.
	StartLine int    // The line number where the code block begins.
	EndLine   int    // The line number where the code block ends.
}

// ExtractTitle determines the document's title, prioritizing the 'title' field
// from frontmatter, and falling back to the first H1 heading in the content.
func (d *Document) ExtractTitle() string {
	// Check frontmatter first
	if d.Metadata != nil {
		if title, ok := d.Metadata["title"].(string); ok && title != "" {
			return title
		}
	}

	// Look for first H1 in content
	content := string(d.Content)
	lines := strings.Split(content, "\n")

	h1Regex := regexp.MustCompile(`^#\s+(.+)$`)

	for _, line := range lines {
		if matches := h1Regex.FindStringSubmatch(line); matches != nil {
			return strings.TrimSpace(matches[1])
		}
	}

	return "Untitled"
}

// ExtractSections parses the document's content to identify and extract all
// sections based on markdown headings.
func (d *Document) ExtractSections() []Section {
	content := string(d.Content)
	lines := strings.Split(content, "\n")

	headerRegex := regexp.MustCompile(`^(#{1,6})\s+(.+)$`)
	sections := []Section{}

	var currentSection *Section
	sectionContent := &strings.Builder{}

	for i, line := range lines {
		if matches := headerRegex.FindStringSubmatch(line); matches != nil {
			// Save previous section
			if currentSection != nil {
				currentSection.Content = strings.TrimSpace(sectionContent.String())
				currentSection.EndLine = i - 1
				sections = append(sections, *currentSection)
				sectionContent.Reset()
			}

			// Start new section
			level := len(matches[1])
			title := strings.TrimSpace(matches[2])
			currentSection = &Section{
				ID:        generateSectionID(title),
				Title:     title,
				Level:     level,
				StartLine: i,
			}
		} else if currentSection != nil {
			sectionContent.WriteString(line)
			sectionContent.WriteString("\n")
		}
	}

	// Save last section
	if currentSection != nil {
		currentSection.Content = strings.TrimSpace(sectionContent.String())
		currentSection.EndLine = len(lines) - 1
		sections = append(sections, *currentSection)
	}

	return sections
}

// ExtractLinks finds all markdown links within the document's content and
// categorizes them as internal or external.
func (d *Document) ExtractLinks() []Link {
	content := string(d.Content)
	linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

	matches := linkRegex.FindAllStringSubmatch(content, -1)
	links := make([]Link, 0, len(matches))

	for _, match := range matches {
		link := Link{
			Text: match[1],
			URL:  match[2],
		}

		// Determine if internal or external
		link.IsInternal = !strings.HasPrefix(link.URL, "http://") &&
			!strings.HasPrefix(link.URL, "https://") &&
			!strings.HasPrefix(link.URL, "//")

		links = append(links, link)
	}

	return links
}

// ExtractCodeBlocks finds all fenced code blocks within the document's content.
func (d *Document) ExtractCodeBlocks() []CodeBlock {
	content := string(d.Content)
	lines := strings.Split(content, "\n")

	codeBlocks := []CodeBlock{}
	inCodeBlock := false
	var currentBlock *CodeBlock
	codeContent := &strings.Builder{}

	codeBlockRegex := regexp.MustCompile("^```(\\w*)$")

	for i, line := range lines {
		if matches := codeBlockRegex.FindStringSubmatch(line); matches != nil {
			if !inCodeBlock {
				// Start of code block
				inCodeBlock = true
				currentBlock = &CodeBlock{
					Language:  matches[1],
					StartLine: i,
				}
			} else {
				// End of code block
				inCodeBlock = false
				if currentBlock != nil {
					currentBlock.Content = codeContent.String()
					currentBlock.EndLine = i
					codeBlocks = append(codeBlocks, *currentBlock)
					codeContent.Reset()
				}
			}
		} else if inCodeBlock {
			codeContent.WriteString(line)
			codeContent.WriteString("\n")
		}
	}

	return codeBlocks
}

// ExtractFrontmatter parses YAML frontmatter from the beginning of a document's content.
// It returns the parsed metadata and the content with the frontmatter block removed.
func ExtractFrontmatter(content []byte) (map[string]interface{}, []byte) {
	if !bytes.HasPrefix(content, []byte("---\n")) {
		return nil, content
	}

	// Find end of frontmatter
	endIndex := bytes.Index(content[4:], []byte("\n---\n"))
	if endIndex == -1 {
		return nil, content
	}

	// For now, return empty map and content without frontmatter
	// TODO: Implement proper YAML parsing
	frontmatterEnd := endIndex + 4 + 4 // account for both --- markers
	return make(map[string]interface{}), content[frontmatterEnd:]
}

// generateSectionID creates a URL-friendly slug from a section title.
func generateSectionID(title string) string {
	// Convert to lowercase and replace spaces with hyphens
	id := strings.ToLower(title)
	id = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(id, "-")
	id = strings.Trim(id, "-")
	return id
}

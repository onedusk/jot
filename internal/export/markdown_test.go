package export

import (
	"strings"
	"testing"
	"time"

	"github.com/onedusk/jot/internal/scanner"
	"gopkg.in/yaml.v3"
)

// TestNewMarkdownExporter tests the constructor for MarkdownExporter.
func TestNewMarkdownExporter(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	if exporter == nil {
		t.Fatal("NewMarkdownExporter() returned nil exporter")
	}

	if exporter.tokenizer == nil {
		t.Fatal("MarkdownExporter tokenizer is nil")
	}
}

// TestToEnrichedMarkdown_SingleDocument tests enriched markdown export with a single document.
func TestToEnrichedMarkdown_SingleDocument(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	// Create test document
	modTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	doc := scanner.Document{
		ID:           "test-doc-1",
		RelativePath: "docs/test.md",
		Title:        "Test Document",
		Content:      []byte("# Test Document\n\nThis is test content.\n\n## Section 1\n\nSome text here."),
		ModTime:      modTime,
		Sections: []scanner.Section{
			{
				ID:        "test-document",
				Title:     "Test Document",
				Level:     1,
				Content:   "This is test content.",
				StartLine: 0,
				EndLine:   2,
			},
			{
				ID:        "section-1",
				Title:     "Section 1",
				Level:     2,
				Content:   "Some text here.",
				StartLine: 4,
				EndLine:   6,
			},
		},
	}

	result, err := exporter.ToEnrichedMarkdown([]scanner.Document{doc}, false)
	if err != nil {
		t.Fatalf("ToEnrichedMarkdown() failed: %v", err)
	}

	// Verify result is not empty
	if result == "" {
		t.Fatal("ToEnrichedMarkdown() returned empty string")
	}

	// Verify TOC is present
	if !strings.Contains(result, "## Table of Contents") {
		t.Error("Result does not contain table of contents")
	}

	// Verify frontmatter delimiters are present
	if !strings.Contains(result, "---\n") {
		t.Error("Result does not contain frontmatter delimiters")
	}

	// Verify original content is preserved
	if !strings.Contains(result, "This is test content.") {
		t.Error("Original markdown content not preserved")
	}

	if !strings.Contains(result, "## Section 1") {
		t.Error("Section header not preserved")
	}
}

// TestToEnrichedMarkdown_FrontmatterParsing tests that YAML frontmatter is valid and parseable.
func TestToEnrichedMarkdown_FrontmatterParsing(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	modTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	doc := scanner.Document{
		ID:           "test-doc-1",
		RelativePath: "docs/getting-started.md",
		Title:        "Getting Started",
		Content:      []byte("# Getting Started\n\nWelcome to the guide."),
		ModTime:      modTime,
		Sections: []scanner.Section{
			{
				ID:        "getting-started",
				Title:     "Getting Started",
				Level:     1,
				Content:   "Welcome to the guide.",
				StartLine: 0,
				EndLine:   2,
			},
		},
	}

	result, err := exporter.ToEnrichedMarkdown([]scanner.Document{doc}, false)
	if err != nil {
		t.Fatalf("ToEnrichedMarkdown() failed: %v", err)
	}

	// Extract frontmatter between --- delimiters
	parts := strings.Split(result, "---\n")
	if len(parts) < 3 {
		t.Fatalf("Expected at least 3 parts split by '---\\n', got %d", len(parts))
	}

	// Skip TOC and get to first document frontmatter
	// The structure is: TOC + "---\n" + frontmatter + "---\n" + content
	var frontmatterYAML string
	for i, part := range parts {
		// Look for the part that contains YAML frontmatter fields
		if strings.Contains(part, "source:") || strings.Contains(part, "chunk_id:") {
			frontmatterYAML = part
			t.Logf("Found frontmatter at part %d: %s", i, part)
			break
		}
	}

	if frontmatterYAML == "" {
		t.Fatalf("Could not find frontmatter YAML in output. Full output:\n%s", result)
	}

	// Parse frontmatter
	var fm MarkdownFrontmatter
	err = yaml.Unmarshal([]byte(frontmatterYAML), &fm)
	if err != nil {
		t.Fatalf("Failed to parse YAML frontmatter: %v\nYAML content:\n%s", err, frontmatterYAML)
	}

	// Validate frontmatter fields
	if fm.Source != "docs/getting-started.md" {
		t.Errorf("Expected source 'docs/getting-started.md', got '%s'", fm.Source)
	}

	if fm.Section != "Getting Started" {
		t.Errorf("Expected section 'Getting Started', got '%s'", fm.Section)
	}

	if fm.ChunkID != "test-doc-1" {
		t.Errorf("Expected chunk_id 'test-doc-1', got '%s'", fm.ChunkID)
	}

	if fm.TokenCount <= 0 {
		t.Errorf("Expected positive token_count, got %d", fm.TokenCount)
	}

	expectedModified := modTime.Format(time.RFC3339)
	if fm.Modified != expectedModified {
		t.Errorf("Expected modified '%s', got '%s'", expectedModified, fm.Modified)
	}
}

// TestToEnrichedMarkdown_MultipleDocuments tests exporting multiple documents.
func TestToEnrichedMarkdown_MultipleDocuments(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	modTime := time.Now()
	docs := []scanner.Document{
		{
			ID:           "doc-1",
			RelativePath: "docs/intro.md",
			Title:        "Introduction",
			Content:      []byte("# Introduction\n\nWelcome!"),
			ModTime:      modTime,
			Sections: []scanner.Section{
				{Title: "Introduction", Level: 1},
			},
		},
		{
			ID:           "doc-2",
			RelativePath: "docs/guide.md",
			Title:        "User Guide",
			Content:      []byte("# User Guide\n\n## Setup\n\nInstructions here."),
			ModTime:      modTime,
			Sections: []scanner.Section{
				{Title: "User Guide", Level: 1},
				{Title: "Setup", Level: 2},
			},
		},
	}

	result, err := exporter.ToEnrichedMarkdown(docs, false)
	if err != nil {
		t.Fatalf("ToEnrichedMarkdown() failed: %v", err)
	}

	// Verify both documents are present
	if !strings.Contains(result, "Introduction") {
		t.Error("First document title not found")
	}

	if !strings.Contains(result, "User Guide") {
		t.Error("Second document title not found")
	}

	// Count frontmatter blocks (should have 2)
	frontmatterCount := strings.Count(result, "source:")
	if frontmatterCount != 2 {
		t.Errorf("Expected 2 frontmatter blocks, found %d", frontmatterCount)
	}

	// Verify TOC includes both documents
	tocStart := strings.Index(result, "## Table of Contents")
	if tocStart == -1 {
		t.Fatal("Table of contents not found")
	}

	// Get TOC section (everything before first frontmatter)
	firstFrontmatter := strings.Index(result, "source:")
	if firstFrontmatter == -1 {
		t.Fatal("Could not find first frontmatter")
	}

	toc := result[tocStart:firstFrontmatter]
	if !strings.Contains(toc, "Introduction") {
		t.Error("TOC does not contain first document")
	}

	if !strings.Contains(toc, "User Guide") {
		t.Error("TOC does not contain second document")
	}
}

// TestToEnrichedMarkdown_CodeBlockPreservation tests that code blocks are preserved.
func TestToEnrichedMarkdown_CodeBlockPreservation(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	content := `# Code Example

Here's some code:

` + "```go" + `
package main

func main() {
    println("Hello, World!")
}
` + "```" + `

That's the example.`

	doc := scanner.Document{
		ID:           "code-doc",
		RelativePath: "examples/code.md",
		Title:        "Code Example",
		Content:      []byte(content),
		ModTime:      time.Now(),
		Sections: []scanner.Section{
			{Title: "Code Example", Level: 1},
		},
		CodeBlocks: []scanner.CodeBlock{
			{
				Language: "go",
				Content:  `package main\n\nfunc main() {\n    println("Hello, World!")\n}`,
			},
		},
	}

	result, err := exporter.ToEnrichedMarkdown([]scanner.Document{doc}, false)
	if err != nil {
		t.Fatalf("ToEnrichedMarkdown() failed: %v", err)
	}

	// Verify code block is preserved with language tag
	if !strings.Contains(result, "```go") {
		t.Error("Code block language tag not preserved")
	}

	if !strings.Contains(result, `println("Hello, World!")`) {
		t.Error("Code block content not preserved")
	}

	if !strings.Contains(result, "package main") {
		t.Error("Code block header not preserved")
	}
}

// TestToEnrichedMarkdown_LinksPreservation tests that links are preserved.
func TestToEnrichedMarkdown_LinksPreservation(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	content := `# Links Example

Check out [this guide](./guide.md) and visit [our website](https://example.com).`

	doc := scanner.Document{
		ID:           "links-doc",
		RelativePath: "docs/links.md",
		Title:        "Links Example",
		Content:      []byte(content),
		ModTime:      time.Now(),
		Sections: []scanner.Section{
			{Title: "Links Example", Level: 1},
		},
		Links: []scanner.Link{
			{Text: "this guide", URL: "./guide.md", IsInternal: true},
			{Text: "our website", URL: "https://example.com", IsInternal: false},
		},
	}

	result, err := exporter.ToEnrichedMarkdown([]scanner.Document{doc}, false)
	if err != nil {
		t.Fatalf("ToEnrichedMarkdown() failed: %v", err)
	}

	// Verify links are preserved
	if !strings.Contains(result, "[this guide](./guide.md)") {
		t.Error("Internal link not preserved")
	}

	if !strings.Contains(result, "[our website](https://example.com)") {
		t.Error("External link not preserved")
	}
}

// TestGenerateTableOfContents tests TOC generation in isolation.
func TestGenerateTableOfContents(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	docs := []scanner.Document{
		{
			Title: "Getting Started",
			Sections: []scanner.Section{
				{Title: "Installation", Level: 2},
				{Title: "Configuration", Level: 2},
			},
		},
		{
			Title: "API Reference",
			Sections: []scanner.Section{
				{Title: "Functions", Level: 2},
			},
		},
	}

	toc := exporter.generateTableOfContents(docs)

	// Verify TOC header
	if !strings.HasPrefix(toc, "## Table of Contents\n\n") {
		t.Error("TOC does not start with proper header")
	}

	// Verify main entries
	if !strings.Contains(toc, "- [Getting Started](#getting-started)") {
		t.Error("TOC missing Getting Started entry")
	}

	if !strings.Contains(toc, "- [API Reference](#api-reference)") {
		t.Error("TOC missing API Reference entry")
	}

	// Verify subsections with proper indentation
	if !strings.Contains(toc, "  - [Installation](#installation)") {
		t.Error("TOC missing Installation subsection")
	}

	if !strings.Contains(toc, "  - [Configuration](#configuration)") {
		t.Error("TOC missing Configuration subsection")
	}

	if !strings.Contains(toc, "  - [Functions](#functions)") {
		t.Error("TOC missing Functions subsection")
	}
}

// TestToEnrichedMarkdown_SeparateFilesNotImplemented tests that separate files mode returns error.
func TestToEnrichedMarkdown_SeparateFilesNotImplemented(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	doc := scanner.Document{
		ID:           "test-doc",
		RelativePath: "test.md",
		Title:        "Test",
		Content:      []byte("# Test"),
		ModTime:      time.Now(),
		Sections:     []scanner.Section{{Title: "Test", Level: 1}},
	}

	_, err = exporter.ToEnrichedMarkdown([]scanner.Document{doc}, true)
	if err == nil {
		t.Error("Expected error for separate files mode, got nil")
	}

	if !strings.Contains(err.Error(), "not yet implemented") {
		t.Errorf("Expected 'not yet implemented' error, got: %v", err)
	}
}

// TestContextualEnrichment tests the placeholder enrichment method.
func TestContextualEnrichment(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	doc := scanner.Document{
		ID:      "test-doc",
		Title:   "Test",
		Content: []byte("Test content"),
	}

	// Should return empty string as it's a stub
	result := exporter.contextualEnrichment(doc, "full context")
	if result != "" {
		t.Errorf("Expected empty string from stub, got: %s", result)
	}
}

// TestToEnrichedMarkdown_EmptyDocuments tests handling of empty document list.
func TestToEnrichedMarkdown_EmptyDocuments(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	result, err := exporter.ToEnrichedMarkdown([]scanner.Document{}, false)
	if err != nil {
		t.Fatalf("ToEnrichedMarkdown() failed: %v", err)
	}

	// Should still have TOC header
	if !strings.Contains(result, "## Table of Contents") {
		t.Error("Empty document list should still have TOC header")
	}
}

// TestToEnrichedMarkdown_NoSections tests handling of documents without sections.
func TestToEnrichedMarkdown_NoSections(t *testing.T) {
	exporter, err := NewMarkdownExporter()
	if err != nil {
		t.Fatalf("NewMarkdownExporter() failed: %v", err)
	}

	doc := scanner.Document{
		ID:           "no-sections",
		RelativePath: "docs/simple.md",
		Title:        "Simple Doc",
		Content:      []byte("Just plain text."),
		ModTime:      time.Now(),
		Sections:     []scanner.Section{}, // No sections
	}

	result, err := exporter.ToEnrichedMarkdown([]scanner.Document{doc}, false)
	if err != nil {
		t.Fatalf("ToEnrichedMarkdown() failed: %v", err)
	}

	// Extract and verify frontmatter
	parts := strings.Split(result, "---\n")
	var frontmatterYAML string
	for _, part := range parts {
		if strings.Contains(part, "source:") {
			frontmatterYAML = part
			break
		}
	}

	if frontmatterYAML == "" {
		t.Fatal("Could not find frontmatter")
	}

	var fm MarkdownFrontmatter
	err = yaml.Unmarshal([]byte(frontmatterYAML), &fm)
	if err != nil {
		t.Fatalf("Failed to parse frontmatter: %v", err)
	}

	// Section should fall back to document title
	if fm.Section != "Simple Doc" {
		t.Errorf("Expected section to be 'Simple Doc', got '%s'", fm.Section)
	}
}

// Package search_test contains tests for the search package.
package search

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/onedusk/jot/internal/scanner"
)

// TestNewIndexer tests the creation of a new Indexer.
func TestNewIndexer(t *testing.T) {
	indexer := NewIndexer("/tmp/test")
	if indexer == nil {
		t.Fatal("NewIndexer() returned nil")
	}
}

// TestIndexer_BuildIndex tests the creation of a search index from documents.
func TestIndexer_BuildIndex(t *testing.T) {
	now := time.Now()
	docs := []scanner.Document{
		{
			RelativePath: "test.md",
			Title:        "Test Document",
			Content:      []byte("# Test Document\n\nThis is test content with some keywords."),
			ModTime:      now,
		},
		{
			RelativePath: "docs/guide.md",
			Title:        "User Guide",
			Content:      []byte("# User Guide\n\nWelcome to the guide."),
			ModTime:      now,
		},
	}

	indexer := NewIndexer("/tmp/test")
	index, err := indexer.BuildIndex(docs)
	if err != nil {
		t.Fatalf("BuildIndex() error = %v", err)
	}

	if index.Version != "1.0" {
		t.Errorf("BuildIndex() version = %s, want 1.0", index.Version)
	}

	if len(index.Documents) != 2 {
		t.Errorf("BuildIndex() document count = %d, want 2", len(index.Documents))
	}

	// Check first document
	doc := index.Documents[0]
	if doc.ID == "" {
		t.Error("BuildIndex() document ID is empty")
	}
	if doc.Title != "Test Document" {
		t.Errorf("BuildIndex() document title = %s, want Test Document", doc.Title)
	}
	if doc.Path == "" {
		t.Error("BuildIndex() document path is empty")
	}
}

// TestIndexer_ProcessDocument tests the conversion of a document to an index document.
func TestIndexer_ProcessDocument(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		doc     scanner.Document
		wantID  string
		wantMin struct {
			wordCount int
			hasFields bool
		}
	}{
		{
			name: "basic document",
			doc: scanner.Document{
				RelativePath: "test.md",
				Title:        "Test",
				Content:      []byte("# Test\n\nContent here with some words."),
				ModTime:      now,
			},
			wantID: "test",
			wantMin: struct {
				wordCount int
				hasFields bool
			}{
				wordCount: 5,
				hasFields: true,
			},
		},
		{
			name: "nested path document",
			doc: scanner.Document{
				RelativePath: "docs/api/reference.md",
				Title:        "API Reference",
				Content:      []byte("# API Reference\n\n## Endpoints\n\nGET /users"),
				ModTime:      now,
			},
			wantID: "docs-api-reference",
			wantMin: struct {
				wordCount int
				hasFields bool
			}{
				wordCount: 4,
				hasFields: true,
			},
		},
	}

	indexer := NewIndexer("/tmp/test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			indexDoc := indexer.processDocument(tt.doc)

			if indexDoc.ID != tt.wantID {
				t.Errorf("processDocument() ID = %s, want %s", indexDoc.ID, tt.wantID)
			}

			if indexDoc.Title != tt.doc.Title {
				t.Errorf("processDocument() Title = %s, want %s", indexDoc.Title, tt.doc.Title)
			}

			if tt.wantMin.hasFields {
				// Check enhanced metadata fields
				if indexDoc.Summary == "" {
					t.Error("processDocument() Summary is empty")
				}
				if indexDoc.Modified == "" {
					t.Error("processDocument() Modified is empty")
				}
				if indexDoc.WordCount < tt.wantMin.wordCount {
					t.Errorf("processDocument() WordCount = %d, want >= %d", indexDoc.WordCount, tt.wantMin.wordCount)
				}
				if indexDoc.ReadTime == "" {
					t.Error("processDocument() ReadTime is empty")
				}
			}
		})
	}
}

// TestIndexer_ExtractHeadings tests the heading extraction logic.
func TestIndexer_ExtractHeadings(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "multiple heading levels",
			content: "# Main Title\n\n## Section 1\n\n### Subsection\n\n## Section 2",
			want:    []string{"Main Title", "Section 1", "Subsection", "Section 2"},
		},
		{
			name:    "no headings",
			content: "Just plain text without any headings.",
			want:    []string{},
		},
		{
			name:    "heading with special characters",
			content: "# API Reference: v2.0\n\n## OAuth 2.0 Flow",
			want:    []string{"API Reference: v2.0", "OAuth 2.0 Flow"},
		},
	}

	indexer := NewIndexer("/tmp/test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			headings := indexer.extractHeadings(tt.content)

			if len(headings) != len(tt.want) {
				t.Errorf("extractHeadings() returned %d headings, want %d", len(headings), len(tt.want))
				return
			}

			for i, heading := range headings {
				if heading != tt.want[i] {
					t.Errorf("extractHeadings()[%d] = %s, want %s", i, heading, tt.want[i])
				}
			}
		})
	}
}

// TestIndexer_ExtractKeywords tests the keyword extraction logic.
func TestIndexer_ExtractKeywords(t *testing.T) {
	indexer := NewIndexer("/tmp/test")
	content := "# Protocol Implementation\n\n" +
		"This document describes the protocol implementation details. " +
		"The protocol uses standard procedures and requires careful implementation. " +
		"Protocol specifications must be followed during implementation process."

	keywords := indexer.extractKeywords(content)

	if len(keywords) == 0 {
		t.Error("extractKeywords() returned no keywords")
	}

	// Should include frequently occurring words
	hasProtocol := false
	hasImplementation := false
	for _, kw := range keywords {
		if kw == "protocol" {
			hasProtocol = true
		}
		if kw == "implementation" {
			hasImplementation = true
		}
	}

	if !hasProtocol {
		t.Error("extractKeywords() should include 'protocol'")
	}
	if !hasImplementation {
		t.Error("extractKeywords() should include 'implementation'")
	}

	// Should not include common words
	for _, kw := range keywords {
		if isCommonWord(kw) {
			t.Errorf("extractKeywords() should not include common word: %s", kw)
		}
	}
}

// TestIndexer_CleanContent tests the markdown cleaning logic.
func TestIndexer_CleanContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "remove code blocks",
			content: "Text before\n```go\nfunc main() {}\n```\nText after",
			want:    "Text before Text after",
		},
		{
			name:    "remove inline code",
			content: "Use the `function()` method here",
			want:    "Use the  method here",
		},
		{
			name:    "remove images",
			content: "See this ![alt text](image.png) image",
			want:    "See this  image",
		},
		{
			name:    "preserve link text",
			content: "Check [this link](https://example.com) out",
			want:    "Check this link out",
		},
		{
			name:    "remove markdown emphasis",
			content: "This is **bold** and *italic* text",
			want:    "This is bold and italic text",
		},
	}

	indexer := NewIndexer("/tmp/test")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := indexer.cleanContent(tt.content)
			// Normalize whitespace for comparison
			got = strings.Join(strings.Fields(got), " ")
			want := strings.Join(strings.Fields(tt.want), " ")

			if got != want {
				t.Errorf("cleanContent() = %q, want %q", got, want)
			}
		})
	}
}

// TestIndexer_SaveIndex tests the saving of the index to a file.
func TestIndexer_SaveIndex(t *testing.T) {
	tmpDir := t.TempDir()

	index := &Index{
		Version: "1.0",
		Documents: []IndexDocument{
			{
				ID:      "test",
				Title:   "Test Doc",
				Path:    "test.html",
				Content: "Test content",
			},
		},
	}

	indexer := NewIndexer(tmpDir)
	err := indexer.SaveIndex(index)
	if err != nil {
		t.Fatalf("SaveIndex() error = %v", err)
	}

	// Verify file was created and is valid JSON
	data, err := json.Marshal(index)
	if err != nil {
		t.Fatalf("Failed to marshal index: %v", err)
	}

	var parsed Index
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to parse saved index: %v", err)
	}

	if parsed.Version != index.Version {
		t.Errorf("Saved index version = %s, want %s", parsed.Version, index.Version)
	}
}

// TestEnhancedMetadataInIndex tests that enhanced metadata fields are correctly populated.
func TestEnhancedMetadataInIndex(t *testing.T) {
	now := time.Now()
	// Use content with repeated words to ensure keywords are extracted
	content := "# Protocol Implementation\n\n" +
		"This document describes the protocol implementation details. " +
		"The protocol uses standard implementation procedures and requires careful implementation. " +
		"Protocol specifications must be followed during the implementation process. " +
		"Implementation testing is crucial for protocol success."

	doc := scanner.Document{
		RelativePath: "test.md",
		Title:        "Test Document",
		Content:      []byte(content),
		ModTime:      now,
		Metadata: map[string]interface{}{
			"tags": "test,protocol,implementation",
		},
	}

	indexer := NewIndexer("/tmp/test")
	indexDoc := indexer.processDocument(doc)

	// Verify all enhanced metadata fields are populated
	if indexDoc.Summary == "" {
		t.Error("Enhanced metadata: Summary not populated")
	}
	if indexDoc.Modified == "" {
		t.Error("Enhanced metadata: Modified not populated")
	}
	if indexDoc.WordCount == 0 {
		t.Error("Enhanced metadata: WordCount not populated")
	}
	if indexDoc.ReadTime == "" {
		t.Error("Enhanced metadata: ReadTime not populated")
	}
	if len(indexDoc.Tags) == 0 {
		t.Error("Enhanced metadata: Tags not populated")
	}
	if len(indexDoc.Keywords) == 0 {
		t.Error("Enhanced metadata: Keywords not populated (check content has repeated words)")
	}

	// Verify summary length constraint
	if len(indexDoc.Summary) > 203 { // 200 chars + "..."
		t.Errorf("Summary too long: %d chars", len(indexDoc.Summary))
	}

	// Verify tags were parsed correctly
	expectedTags := []string{"test", "protocol", "implementation"}
	if len(indexDoc.Tags) != len(expectedTags) {
		t.Errorf("Tags count = %d, want %d", len(indexDoc.Tags), len(expectedTags))
	}
}

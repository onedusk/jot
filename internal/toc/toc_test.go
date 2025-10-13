// Package toc_test contains tests for the toc package.
package toc

import (
	"strings"
	"testing"
	"time"

	"github.com/onedusk/jot/internal/scanner"
)

// TestNewBuilder tests the creation of a new Builder.
func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	if builder == nil {
		t.Fatal("NewBuilder() returned nil")
	}
}

// TestBuilder_Build tests the main TOC building logic.
func TestBuilder_Build(t *testing.T) {
	tests := []struct {
		name      string
		documents []scanner.Document
		wantXML   []string // XML snippets that should be present
	}{
		{
			name: "single document",
			documents: []scanner.Document{
				{
					RelativePath: "README.md",
					Title:        "Project README",
				},
			},
			wantXML: []string{
				`<toc version="1.0" llm-optimized="true">`,
				`<metadata>`,
				`<totalDocs>1</totalDocs>`,
				`<chapter id="readme" path="README.md"`,
				`<title>Project README</title>`,
				`</chapter>`,
				`</toc>`,
			},
		},
		{
			name: "nested documents",
			documents: []scanner.Document{
				{
					RelativePath: "docs/getting-started.md",
					Title:        "Getting Started",
				},
				{
					RelativePath: "docs/installation.md",
					Title:        "Installation Guide",
				},
			},
			wantXML: []string{
				`<section id="docs">`,
				`<title>Docs</title>`,
				`<chapter id="docs-getting-started" path="docs/getting-started.md"`,
				`<title>Getting Started</title>`,
				`<chapter id="docs-installation" path="docs/installation.md"`,
				`<title>Installation Guide</title>`,
			},
		},
		{
			name: "deeply nested documents",
			documents: []scanner.Document{
				{
					RelativePath: "docs/api/reference/endpoints.md",
					Title:        "API Endpoints",
				},
			},
			wantXML: []string{
				`<section id="docs">`,
				`<section id="docs-api">`,
				`<section id="docs-api-reference">`,
				`<chapter id="docs-api-reference-endpoints" path="docs/api/reference/endpoints.md"`,
				`<title>API Endpoints</title>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBuilder()
			toc := builder.Build(tt.documents)
			xml := toc.ToXML()

			for _, want := range tt.wantXML {
				if !strings.Contains(xml, want) {
					t.Errorf("ToXML() missing expected content:\nwant: %s\ngot:\n%s", want, xml)
				}
			}
		})
	}
}

// TestTOCNode_AddChild tests adding a child to a TOCNode.
func TestTOCNode_AddChild(t *testing.T) {
	parent := &TOCNode{
		ID:    "parent",
		Title: "Parent",
	}

	child := &TOCNode{
		ID:    "child",
		Title: "Child",
	}

	parent.AddChild(child)

	if len(parent.Children) != 1 {
		t.Errorf("AddChild() failed, expected 1 child, got %d", len(parent.Children))
	}

	if parent.Children[0] != child {
		t.Error("AddChild() failed, child not added correctly")
	}
}

// TestTOCNode_FindChildByTitle tests finding a child node by its title.
func TestTOCNode_FindChildByTitle(t *testing.T) {
	parent := &TOCNode{
		ID:    "parent",
		Title: "Parent",
		Children: []*TOCNode{
			{ID: "child1", Title: "Child One"},
			{ID: "child2", Title: "Child Two"},
		},
	}

	tests := []struct {
		name      string
		title     string
		wantFound bool
		wantID    string
	}{
		{
			name:      "existing child",
			title:     "Child One",
			wantFound: true,
			wantID:    "child1",
		},
		{
			name:      "non-existing child",
			title:     "Child Three",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			child := parent.FindChildByTitle(tt.title)
			if tt.wantFound {
				if child == nil {
					t.Error("FindChildByTitle() returned nil, expected child")
				} else if child.ID != tt.wantID {
					t.Errorf("FindChildByTitle() returned wrong child, got ID %s, want %s", child.ID, tt.wantID)
				}
			} else {
				if child != nil {
					t.Error("FindChildByTitle() returned child, expected nil")
				}
			}
		})
	}
}

// TestGenerateNodeID tests the generation of node IDs.
func TestGenerateNodeID(t *testing.T) {
	tests := []struct {
		name  string
		parts []string
		want  string
	}{
		{
			name:  "single part",
			parts: []string{"readme"},
			want:  "readme",
		},
		{
			name:  "multiple parts",
			parts: []string{"docs", "getting-started"},
			want:  "docs-getting-started",
		},
		{
			name:  "with special characters",
			parts: []string{"API Reference", "OAuth 2.0"},
			want:  "api-reference-oauth-2-0",
		},
		{
			name:  "with numbers",
			parts: []string{"version-1.2.3"},
			want:  "version-1-2-3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := generateNodeID(tt.parts)
			if got != tt.want {
				t.Errorf("generateNodeID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestHumanizeTitle tests the humanization of file/directory names.
func TestHumanizeTitle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple word",
			input: "docs",
			want:  "Docs",
		},
		{
			name:  "hyphenated",
			input: "getting-started",
			want:  "Getting Started",
		},
		{
			name:  "underscored",
			input: "api_reference",
			want:  "Api Reference",
		},
		{
			name:  "mixed separators",
			input: "my-cool_project",
			want:  "My Cool Project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := humanizeTitle(tt.input)
			if got != tt.want {
				t.Errorf("humanizeTitle() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestTableOfContents_GetNodeByID tests retrieving a node by its ID.
func TestTableOfContents_GetNodeByID(t *testing.T) {
	toc := &TableOfContents{
		Version: "1.0",
		Root: &TOCNode{
			ID:    "root",
			Title: "Root",
			Children: []*TOCNode{
				{
					ID:    "docs",
					Title: "Docs",
					Children: []*TOCNode{
						{
							ID:    "docs-api",
							Title: "API",
							Path:  "docs/api.md",
						},
					},
				},
			},
		},
	}

	// Build index
	toc.buildIndex()

	tests := []struct {
		name   string
		id     string
		wantID string
		found  bool
	}{
		{
			name:   "root node",
			id:     "root",
			wantID: "root",
			found:  true,
		},
		{
			name:   "nested node",
			id:     "docs-api",
			wantID: "docs-api",
			found:  true,
		},
		{
			name:  "non-existent node",
			id:    "not-found",
			found: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := toc.GetNodeByID(tt.id)
			if tt.found {
				if node == nil {
					t.Error("GetNodeByID() returned nil, expected node")
				} else if node.ID != tt.wantID {
					t.Errorf("GetNodeByID() returned wrong node, got ID %s, want %s", node.ID, tt.wantID)
				}
			} else {
				if node != nil {
					t.Error("GetNodeByID() returned node, expected nil")
				}
			}
		})
	}
}

// TestBuilder_SortDocuments tests the document sorting logic.
func TestBuilder_SortDocuments(t *testing.T) {
	docs := []scanner.Document{
		{RelativePath: "docs/b.md", ModTime: time.Now()},
		{RelativePath: "a.md", ModTime: time.Now().Add(-time.Hour)},
		{RelativePath: "docs/a.md", ModTime: time.Now().Add(-2 * time.Hour)},
	}

	builder := NewBuilder()
	builder.sortDocuments(docs)

	// Check order: should be sorted by path
	expected := []string{"a.md", "docs/a.md", "docs/b.md"}
	for i, doc := range docs {
		if doc.RelativePath != expected[i] {
			t.Errorf("sortDocuments() wrong order at index %d: got %s, want %s",
				i, doc.RelativePath, expected[i])
		}
	}
}

// TestBuilder_ExtractMetadata tests the metadata extraction logic.
func TestBuilder_ExtractMetadata(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name    string
		doc     scanner.Document
		wantMin NodeMetadata
	}{
		{
			name: "basic metadata extraction",
			doc: scanner.Document{
				Content:  []byte("# Test Document\n\nThis is a test with some content that is longer than 200 characters to verify summary extraction works correctly. " + strings.Repeat("More words. ", 30)),
				ModTime:  now,
				Metadata: nil,
			},
			wantMin: NodeMetadata{
				Modified:  now,
				Size:      400, // Approximate
				WordCount: 40,  // Approximate minimum
			},
		},
		{
			name: "with tags in metadata",
			doc: scanner.Document{
				Content: []byte("# Tagged Doc\n\nContent here."),
				ModTime: now,
				Metadata: map[string]interface{}{
					"tags": "test,example,documentation",
				},
			},
			wantMin: NodeMetadata{
				Modified: now,
				Tags:     []string{"test", "example", "documentation"},
			},
		},
	}

	builder := NewBuilder()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			metadata := builder.extractMetadata(tt.doc)

			// Check size
			if metadata.Size < tt.wantMin.Size {
				t.Errorf("extractMetadata() Size = %d, want >= %d", metadata.Size, tt.wantMin.Size)
			}

			// Check word count
			if tt.wantMin.WordCount > 0 && metadata.WordCount < tt.wantMin.WordCount {
				t.Errorf("extractMetadata() WordCount = %d, want >= %d", metadata.WordCount, tt.wantMin.WordCount)
			}

			// Check read time is set
			if metadata.ReadTime == "" {
				t.Error("extractMetadata() ReadTime not set")
			}

			// Check summary is set
			if metadata.Summary == "" {
				t.Error("extractMetadata() Summary not set")
			}

			// Check content hash is set
			if !strings.HasPrefix(metadata.ContentHash, "sha256:") {
				t.Errorf("extractMetadata() ContentHash = %s, want sha256: prefix", metadata.ContentHash)
			}

			// Check tags if specified
			if len(tt.wantMin.Tags) > 0 {
				if len(metadata.Tags) != len(tt.wantMin.Tags) {
					t.Errorf("extractMetadata() Tags count = %d, want %d", len(metadata.Tags), len(tt.wantMin.Tags))
				}
			}
		})
	}
}

// TestBuilder_ExtractKeywords tests the keyword extraction logic.
func TestBuilder_ExtractKeywords(t *testing.T) {
	builder := NewBuilder()
	content := "# Protocol Implementation\n\nThis document describes the protocol implementation. " +
		"The protocol uses standard procedures. Implementation details are provided below. " +
		"Protocol specifications must be followed carefully during implementation."

	keywords := builder.extractKeywords(content)

	// Should extract "protocol" and "implementation" as they appear multiple times
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
		t.Error("extractKeywords() should extract 'protocol'")
	}
	if !hasImplementation {
		t.Error("extractKeywords() should extract 'implementation'")
	}

	// Should not extract common words
	for _, kw := range keywords {
		if kw == "this" || kw == "the" || kw == "are" {
			t.Errorf("extractKeywords() should not include common word: %s", kw)
		}
	}
}

// TestTOCWithEnhancedMetadata tests that enhanced metadata is included in the XML output.
func TestTOCWithEnhancedMetadata(t *testing.T) {
	now := time.Now()
	docs := []scanner.Document{
		{
			RelativePath: "test.md",
			Title:        "Test Document",
			Content:      []byte("# Test\n\nThis is test content with protocol implementation details that should generate keywords."),
			ModTime:      now,
		},
	}

	builder := NewBuilder()
	toc := builder.Build(docs)
	xml := toc.ToXML()

	// Check that enhanced metadata is present in XML
	expectedElements := []string{
		`llm-optimized="true"`,
		`<metadata>`,
		`<totalDocs>`,
		`modified=`,
		`size=`,
		`words=`,
		`readTime=`,
		`hash=`,
		`<summary>`,
		`<keywords>`,
	}

	for _, elem := range expectedElements {
		if !strings.Contains(xml, elem) {
			t.Errorf("ToXML() missing expected element: %s\nXML:\n%s", elem, xml)
		}
	}
}

package export

import (
	"strings"
	"testing"

	"github.com/onedusk/jot/internal/scanner"
)

// TestToLLMSTxt tests the llms.txt export format generation.
func TestToLLMSTxt(t *testing.T) {
	exporter := NewLLMSTxtExporter()

	config := ProjectConfig{
		Name:        "Test Project",
		Description: "A test project for llms.txt generation",
	}

	documents := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "Introduction",
			RelativePath: "docs/intro.md",
			Content:      []byte("# Introduction\n\nThis is the introduction to our project. It explains the basics."),
		},
		{
			ID:           "doc2",
			Title:        "Getting Started",
			RelativePath: "docs/getting-started.md",
			Content:      []byte("# Getting Started\n\nQuick start guide for new users."),
		},
		{
			ID:           "doc3",
			Title:        "README",
			RelativePath: "README.md",
			Content:      []byte("# README\n\nMain project readme file."),
		},
	}

	result, err := exporter.ToLLMSTxt(documents, config)
	if err != nil {
		t.Fatalf("ToLLMSTxt() error = %v", err)
	}

	// Test H1 format
	if !strings.Contains(result, "# Test Project\n\n") {
		t.Error("ToLLMSTxt() missing or incorrect H1 header")
	}

	// Test blockquote format
	if !strings.Contains(result, "> A test project for llms.txt generation\n\n") {
		t.Error("ToLLMSTxt() missing or incorrect blockquote")
	}

	// Test H2 sections (should have sections for different directories)
	if !strings.Contains(result, "## ") {
		t.Error("ToLLMSTxt() missing H2 section headers")
	}

	// Test link format: - [Title](path): description
	if !strings.Contains(result, "- [Introduction](docs/intro.md):") {
		t.Error("ToLLMSTxt() missing or incorrect link format for Introduction")
	}

	if !strings.Contains(result, "- [Getting Started](docs/getting-started.md):") {
		t.Error("ToLLMSTxt() missing or incorrect link format for Getting Started")
	}

	if !strings.Contains(result, "- [README](README.md):") {
		t.Error("ToLLMSTxt() missing or incorrect link format for README")
	}

	// Test that descriptions are extracted
	if !strings.Contains(result, "This is the introduction") {
		t.Error("ToLLMSTxt() failed to extract document description")
	}
}

// TestGroupDocumentsBySection tests document grouping by directory.
func TestGroupDocumentsBySection(t *testing.T) {
	documents := []scanner.Document{
		{RelativePath: "docs/intro.md"},
		{RelativePath: "docs/guide.md"},
		{RelativePath: "api/reference.md"},
		{RelativePath: "README.md"},
	}

	grouped := groupDocumentsBySection(documents)

	if len(grouped) != 3 {
		t.Errorf("groupDocumentsBySection() got %d groups, want 3", len(grouped))
	}

	if len(grouped["docs"]) != 2 {
		t.Errorf("groupDocumentsBySection() docs group has %d documents, want 2", len(grouped["docs"]))
	}

	if len(grouped["api"]) != 1 {
		t.Errorf("groupDocumentsBySection() api group has %d documents, want 1", len(grouped["api"]))
	}

	if len(grouped["."]) != 1 {
		t.Errorf("groupDocumentsBySection() root group has %d documents, want 1", len(grouped["."]))
	}
}

// TestExtractFirstParagraph tests paragraph extraction with various content.
func TestExtractFirstParagraph(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		want    string
	}{
		{
			name:    "simple paragraph",
			content: []byte("This is a simple paragraph."),
			want:    "This is a simple paragraph.",
		},
		{
			name:    "skip headers",
			content: []byte("# Header\n\nFirst paragraph content here."),
			want:    "First paragraph content here.",
		},
		{
			name:    "truncate long text",
			content: []byte(strings.Repeat("This is a very long sentence that should be truncated. ", 10)),
			want:    "...", // Should end with ...
		},
		{
			name:    "empty content",
			content: []byte(""),
			want:    "No description available",
		},
		{
			name:    "only headers",
			content: []byte("# Header 1\n## Header 2\n### Header 3"),
			want:    "No description available",
		},
		{
			name:    "multi-line paragraph",
			content: []byte("First line\nSecond line\nThird line"),
			want:    "First line Second line Third line",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractFirstParagraph(tt.content)

			if tt.want == "..." {
				if !strings.HasSuffix(result, "...") {
					t.Errorf("extractFirstParagraph() = %q, want suffix %q", result, tt.want)
				}
				if len(result) > 100 {
					t.Errorf("extractFirstParagraph() length = %d, want <= 100", len(result))
				}
			} else if result != tt.want {
				t.Errorf("extractFirstParagraph() = %q, want %q", result, tt.want)
			}
		})
	}
}

// TestNewLLMSTxtExporter tests the constructor.
func TestNewLLMSTxtExporter(t *testing.T) {
	exporter := NewLLMSTxtExporter()
	if exporter == nil {
		t.Error("NewLLMSTxtExporter() returned nil")
	}
}

// TestToLLMSFullTxt tests the llms-full.txt export format with complete documentation.
func TestToLLMSFullTxt(t *testing.T) {
	exporter := NewLLMSTxtExporter()

	config := ProjectConfig{
		Name:        "Test Documentation",
		Description: "Complete test documentation for llms-full.txt generation",
	}

	documents := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "API Guide",
			RelativePath: "docs/api.md",
			Content:      []byte("# API Guide\n\nThis is the API guide.\n\n## Authentication\n\nUse API keys for authentication.\n\n```go\nfunc main() {\n  fmt.Println(\"Hello\")\n}\n```"),
		},
		{
			ID:           "doc2",
			Title:        "Getting Started",
			RelativePath: "docs/getting-started.md",
			Content:      []byte("# Getting Started\n\nQuick start guide.\n\n[Link to docs](./api.md)"),
		},
		{
			ID:           "doc3",
			Title:        "Project README",
			RelativePath: "README.md",
			Content:      []byte("# Project README\n\nWelcome to the project.\n\nThis is the main readme file."),
		},
	}

	result, err := exporter.ToLLMSFullTxt(documents, config)
	if err != nil {
		t.Fatalf("ToLLMSFullTxt() error = %v", err)
	}

	// Test 1: H1 header format
	if !strings.Contains(result, "# Test Documentation\n\n") {
		t.Error("ToLLMSFullTxt() missing or incorrect H1 header")
	}

	// Test 2: Blockquote format
	if !strings.Contains(result, "> Complete test documentation for llms-full.txt generation\n\n") {
		t.Error("ToLLMSFullTxt() missing or incorrect blockquote")
	}

	// Test 3: Document separator presence
	separatorCount := strings.Count(result, "---\n\n")
	expectedSeparators := len(documents) - 1 // One separator between each pair of documents
	if separatorCount != expectedSeparators {
		t.Errorf("ToLLMSFullTxt() has %d separators, want %d", separatorCount, expectedSeparators)
	}

	// Test 4: Document order - README.md should come first
	readmeIndex := strings.Index(result, "# Project README\n\n")
	gettingStartedIndex := strings.Index(result, "# Getting Started\n\n")
	apiGuideIndex := strings.Index(result, "# API Guide\n\n")

	if readmeIndex == -1 || gettingStartedIndex == -1 || apiGuideIndex == -1 {
		t.Error("ToLLMSFullTxt() missing document titles")
	}

	// README should appear before other docs
	if readmeIndex > gettingStartedIndex || readmeIndex > apiGuideIndex {
		t.Error("ToLLMSFullTxt() README.md is not first in output")
	}

	// After README, should be alphabetically sorted (docs/api.md before docs/getting-started.md)
	if apiGuideIndex > gettingStartedIndex {
		t.Error("ToLLMSFullTxt() documents after README are not sorted alphabetically")
	}

	// Test 5: Content preservation - check that original markdown is preserved
	if !strings.Contains(result, "## Authentication") {
		t.Error("ToLLMSFullTxt() failed to preserve H2 headings")
	}

	if !strings.Contains(result, "```go\nfunc main()") {
		t.Error("ToLLMSFullTxt() failed to preserve code blocks")
	}

	if !strings.Contains(result, "[Link to docs](./api.md)") {
		t.Error("ToLLMSFullTxt() failed to preserve markdown links")
	}

	// Test 6: Each document should have H1 heading
	if !strings.Contains(result, "# Project README\n\n") {
		t.Error("ToLLMSFullTxt() missing H1 for README")
	}
	if !strings.Contains(result, "# API Guide\n\n") {
		t.Error("ToLLMSFullTxt() missing H1 for API Guide")
	}
	if !strings.Contains(result, "# Getting Started\n\n") {
		t.Error("ToLLMSFullTxt() missing H1 for Getting Started")
	}

	// Test 7: Content should be included
	if !strings.Contains(result, "Welcome to the project") {
		t.Error("ToLLMSFullTxt() missing README content")
	}
	if !strings.Contains(result, "Use API keys for authentication") {
		t.Error("ToLLMSFullTxt() missing API guide content")
	}
}

// TestSortDocumentsByImportance tests document sorting with README first.
func TestSortDocumentsByImportance(t *testing.T) {
	documents := []scanner.Document{
		{RelativePath: "docs/zebra.md", Title: "Zebra"},
		{RelativePath: "docs/apple.md", Title: "Apple"},
		{RelativePath: "README.md", Title: "README"},
		{RelativePath: "docs/banana.md", Title: "Banana"},
	}

	sorted := sortDocumentsByImportance(documents)

	// Test 1: README should be first
	if sorted[0].RelativePath != "README.md" {
		t.Errorf("sortDocumentsByImportance() first doc = %s, want README.md", sorted[0].RelativePath)
	}

	// Test 2: After README, should be alphabetically sorted
	if sorted[1].RelativePath != "docs/apple.md" {
		t.Errorf("sortDocumentsByImportance() second doc = %s, want docs/apple.md", sorted[1].RelativePath)
	}
	if sorted[2].RelativePath != "docs/banana.md" {
		t.Errorf("sortDocumentsByImportance() third doc = %s, want docs/banana.md", sorted[2].RelativePath)
	}
	if sorted[3].RelativePath != "docs/zebra.md" {
		t.Errorf("sortDocumentsByImportance() fourth doc = %s, want docs/zebra.md", sorted[3].RelativePath)
	}

	// Test 3: Original slice should not be modified
	if documents[0].RelativePath == "README.md" {
		t.Error("sortDocumentsByImportance() modified the original slice")
	}
}

// TestSortDocumentsByImportance_CaseInsensitive tests README detection is case-insensitive.
func TestSortDocumentsByImportance_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		name     string
		readmePath string
	}{
		{"lowercase", "readme.md"},
		{"uppercase", "README.MD"},
		{"mixedcase", "ReadMe.md"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			documents := []scanner.Document{
				{RelativePath: "docs/other.md", Title: "Other"},
				{RelativePath: tc.readmePath, Title: "README"},
			}

			sorted := sortDocumentsByImportance(documents)

			if sorted[0].RelativePath != tc.readmePath {
				t.Errorf("sortDocumentsByImportance() first doc = %s, want %s", sorted[0].RelativePath, tc.readmePath)
			}
		})
	}
}

// TestEstimateSize tests size estimation function.
func TestEstimateSize(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    int64
	}{
		{
			name:    "empty string",
			content: "",
			want:    0,
		},
		{
			name:    "simple text",
			content: "Hello, World!",
			want:    13,
		},
		{
			name:    "1KB content",
			content: strings.Repeat("a", 1024),
			want:    1024,
		},
		{
			name:    "unicode content",
			content: "Hello 世界",
			want:    12, // "Hello " (6) + "世" (3) + "界" (3)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := estimateSize(tt.content)
			if result != tt.want {
				t.Errorf("estimateSize() = %d, want %d", result, tt.want)
			}
		})
	}
}

// TestToLLMSFullTxt_LargeContent tests size warning for large outputs.
func TestToLLMSFullTxt_LargeContent(t *testing.T) {
	exporter := NewLLMSTxtExporter()

	config := ProjectConfig{
		Name:        "Large Project",
		Description: "Project with large documentation",
	}

	// Create content larger than 1MB
	largeContent := strings.Repeat("This is a test document with repeated content to exceed 1MB size limit.\n", 15000)

	documents := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "Large Document",
			RelativePath: "README.md",
			Content:      []byte(largeContent),
		},
	}

	result, err := exporter.ToLLMSFullTxt(documents, config)
	if err != nil {
		t.Fatalf("ToLLMSFullTxt() error = %v", err)
	}

	// Verify size is estimated correctly
	size := estimateSize(result)
	if size <= 1048576 {
		t.Errorf("ToLLMSFullTxt() size = %d, expected > 1048576 for this test", size)
	}

	// The function should still return valid output even with large content
	if !strings.Contains(result, "# Large Project") {
		t.Error("ToLLMSFullTxt() failed to generate valid output for large content")
	}
}

// TestToLLMSFullTxt_EmptyDocuments tests handling of empty document list.
func TestToLLMSFullTxt_EmptyDocuments(t *testing.T) {
	exporter := NewLLMSTxtExporter()

	config := ProjectConfig{
		Name:        "Empty Project",
		Description: "Project with no documents",
	}

	documents := []scanner.Document{}

	result, err := exporter.ToLLMSFullTxt(documents, config)
	if err != nil {
		t.Fatalf("ToLLMSFullTxt() error = %v", err)
	}

	// Should still have header and description
	if !strings.Contains(result, "# Empty Project\n\n") {
		t.Error("ToLLMSFullTxt() missing H1 header for empty documents")
	}

	if !strings.Contains(result, "> Project with no documents\n\n") {
		t.Error("ToLLMSFullTxt() missing blockquote for empty documents")
	}

	// Should not have any separators
	if strings.Contains(result, "---") {
		t.Error("ToLLMSFullTxt() should not have separators for empty documents")
	}
}

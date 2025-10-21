// Package export_test contains tests for the export package.
package export

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/tokenizer"
	"gopkg.in/yaml.v3"
)

// TestNewExporter tests the creation of a new Exporter.
func TestNewExporter(t *testing.T) {
	exporter := NewExporter()
	if exporter == nil {
		t.Fatal("NewExporter() returned nil")
	}
}

// TestExporter_ToJSON tests the JSON export functionality.
func TestExporter_ToJSON(t *testing.T) {
	docs := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "Test Document",
			RelativePath: "test.md",
			Content:      []byte("# Test\n\nThis is a test document."),
			ModTime:      time.Now(),
			Sections: []scanner.Section{
				{
					ID:        "test",
					Title:     "Test",
					Level:     1,
					Content:   "This is a test document.",
					StartLine: 0,
					EndLine:   2,
				},
			},
			Links: []scanner.Link{
				{
					Text:       "Example",
					URL:        "https://example.com",
					IsInternal: false,
				},
			},
		},
	}

	exporter := NewExporter()
	jsonData, err := exporter.ToJSON(docs)
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Parse JSON to verify structure
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Check required fields
	if result["version"] == nil {
		t.Error("ToJSON() missing version field")
	}
	if result["generated"] == nil {
		t.Error("ToJSON() missing generated field")
	}
	if result["documents"] == nil {
		t.Error("ToJSON() missing documents field")
	}

	// Check documents array
	documents := result["documents"].([]interface{})
	if len(documents) != 1 {
		t.Errorf("ToJSON() expected 1 document, got %d", len(documents))
	}

	// Check first document
	doc := documents[0].(map[string]interface{})
	if doc["id"] != "doc1" {
		t.Errorf("ToJSON() document id = %v, want doc1", doc["id"])
	}
	if doc["title"] != "Test Document" {
		t.Errorf("ToJSON() document title = %v, want Test Document", doc["title"])
	}
	if doc["path"] != "test.md" {
		t.Errorf("ToJSON() document path = %v, want test.md", doc["path"])
	}
}

// TestExporter_ToYAML tests the YAML export functionality.
func TestExporter_ToYAML(t *testing.T) {
	docs := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "Test Document",
			RelativePath: "test.md",
			Content:      []byte("# Test\n\nThis is a test document."),
			ModTime:      time.Now(),
		},
	}

	exporter := NewExporter()
	yamlData, err := exporter.ToYAML(docs)
	if err != nil {
		t.Fatalf("ToYAML() error = %v", err)
	}

	// Parse YAML to verify structure
	var result map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlData), &result); err != nil {
		t.Fatalf("Failed to parse YAML: %v", err)
	}

	// Check required fields
	if result["version"] == nil {
		t.Error("ToYAML() missing version field")
	}
	if result["documents"] == nil {
		t.Error("ToYAML() missing documents field")
	}
}

// TestExporter_ToLLMFormat tests the LLM-optimized format export functionality.
func TestExporter_ToLLMFormat(t *testing.T) {
	docs := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "API Reference",
			RelativePath: "api.md",
			Content:      []byte("# API Reference\n\n## Endpoints\n\nGET /users - List all users\n\n```json\n{\n  \"users\": []\n}\n```"),
			Sections: []scanner.Section{
				{
					ID:    "endpoints",
					Title: "Endpoints",
					Level: 2,
				},
			},
			CodeBlocks: []scanner.CodeBlock{
				{
					Language: "json",
					Content:  "{\n  \"users\": []\n}",
				},
			},
		},
	}

	exporter := NewExporter()
	llmData, err := exporter.ToLLMFormat(docs)
	if err != nil {
		t.Fatalf("ToLLMFormat() error = %v", err)
	}

	// Check structure
	if llmData.Version == "" {
		t.Error("ToLLMFormat() missing version")
	}
	if len(llmData.Documents) != 1 {
		t.Errorf("ToLLMFormat() expected 1 document, got %d", len(llmData.Documents))
	}

	doc := llmData.Documents[0]
	if doc.ID != "doc1" {
		t.Errorf("ToLLMFormat() document id = %v, want doc1", doc.ID)
	}

	// Check chunks are created
	if len(doc.Chunks) == 0 {
		t.Error("ToLLMFormat() missing chunks")
	}

	// Check semantic index
	if llmData.Index == nil {
		t.Error("ToLLMFormat() missing index")
	}
}

// TestChunkDocument tests the document chunking logic.
func TestChunkDocument(t *testing.T) {
	tok, err := tokenizer.NewTokenizer()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	content := strings.Repeat("This is a test sentence. ", 100)
	doc := scanner.Document{
		Content: []byte(content),
	}

	chunks := chunkDocument(doc, 100, 20, tok)
	if len(chunks) == 0 {
		t.Error("chunkDocument() returned no chunks")
	}

	// Verify all chunks have token counts
	for i, chunk := range chunks {
		if chunk.TokenCount == 0 {
			t.Errorf("Chunk %d has zero token count", i)
		}
		// Verify token count matches actual text
		actualTokens := tok.Count(chunk.Text)
		if chunk.TokenCount != actualTokens {
			t.Errorf("Chunk %d TokenCount mismatch: got %d, actual %d", i, chunk.TokenCount, actualTokens)
		}
		// Verify chunk doesn't exceed maxTokens
		if chunk.TokenCount > 100 {
			t.Errorf("Chunk %d exceeds maxTokens: %d > 100", i, chunk.TokenCount)
		}
	}

	// Check that chunks are properly positioned
	for i := 1; i < len(chunks); i++ {
		prev := chunks[i-1]
		curr := chunks[i]

		// Chunks should overlap (curr starts before prev ends)
		if curr.StartPos >= prev.EndPos {
			t.Errorf("chunkDocument() chunks not overlapping at index %d (curr.StartPos=%d >= prev.EndPos=%d)",
				i, curr.StartPos, prev.EndPos)
		}
	}
}

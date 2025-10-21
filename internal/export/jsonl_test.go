package export

import (
	"bufio"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/onedusk/jot/internal/scanner"
)

// TestNewJSONLExporter tests the creation of a new JSONLExporter.
func TestNewJSONLExporter(t *testing.T) {
	exporter := NewJSONLExporter()
	if exporter == nil {
		t.Fatal("NewJSONLExporter() returned nil")
	}
}

// TestToJSONL tests the basic JSONL export functionality.
func TestToJSONL(t *testing.T) {
	// Create test documents
	docs := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "Test Document",
			RelativePath: "test.md",
			Content:      []byte("# Test\n\nThis is a test document with enough content to potentially create multiple chunks if needed."),
			ModTime:      time.Now(),
		},
	}

	// Create exporter and export to JSONL
	exporter := NewJSONLExporter()
	jsonlOutput, err := exporter.ToJSONL(docs, 50, 10)
	if err != nil {
		t.Fatalf("ToJSONL() error = %v", err)
	}

	// Verify output is not empty
	if jsonlOutput == "" {
		t.Fatal("ToJSONL() returned empty string")
	}

	// Verify output ends with newline
	if !strings.HasSuffix(jsonlOutput, "\n") {
		t.Error("ToJSONL() output should end with newline")
	}

	// Split by lines and verify each is valid JSON
	lines := strings.Split(strings.TrimSpace(jsonlOutput), "\n")
	if len(lines) == 0 {
		t.Fatal("ToJSONL() produced no lines")
	}

	for i, line := range lines {
		if line == "" {
			continue // Skip empty lines
		}

		// Verify line is valid JSON
		var metadata ChunkMetadata
		if err := json.Unmarshal([]byte(line), &metadata); err != nil {
			t.Errorf("Line %d is not valid JSON: %v\nLine: %s", i, err, line)
			continue
		}

		// Verify required fields are present
		if metadata.DocID == "" {
			t.Errorf("Line %d missing doc_id", i)
		}
		if metadata.ChunkID == "" {
			t.Errorf("Line %d missing chunk_id", i)
		}
		if metadata.Text == "" {
			t.Errorf("Line %d missing text", i)
		}
		if metadata.Source == "" {
			t.Errorf("Line %d missing source", i)
		}

		// Verify JSON does not contain indentation (compact format)
		if strings.Contains(line, "\n") {
			t.Errorf("Line %d contains newlines (not compact JSON)", i)
		}
	}
}

// TestToJSONL_MultipleDocuments tests JSONL export with multiple documents.
func TestToJSONL_MultipleDocuments(t *testing.T) {
	docs := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "Document 1",
			RelativePath: "doc1.md",
			Content:      []byte("First document content."),
			ModTime:      time.Now(),
		},
		{
			ID:           "doc2",
			Title:        "Document 2",
			RelativePath: "doc2.md",
			Content:      []byte("Second document content."),
			ModTime:      time.Now(),
		},
	}

	exporter := NewJSONLExporter()
	jsonlOutput, err := exporter.ToJSONL(docs, 50, 10)
	if err != nil {
		t.Fatalf("ToJSONL() error = %v", err)
	}

	// Count chunks from each document
	lines := strings.Split(strings.TrimSpace(jsonlOutput), "\n")
	doc1Count := 0
	doc2Count := 0

	for _, line := range lines {
		if line == "" {
			continue
		}

		var metadata ChunkMetadata
		if err := json.Unmarshal([]byte(line), &metadata); err != nil {
			t.Fatalf("Failed to unmarshal line: %v", err)
		}

		if metadata.DocID == "doc1" {
			doc1Count++
		} else if metadata.DocID == "doc2" {
			doc2Count++
		}
	}

	if doc1Count == 0 {
		t.Error("ToJSONL() produced no chunks for doc1")
	}
	if doc2Count == 0 {
		t.Error("ToJSONL() produced no chunks for doc2")
	}
}

// TestToJSONL_ChunkNavigation tests that prev/next chunk IDs are set correctly.
func TestToJSONL_ChunkNavigation(t *testing.T) {
	// Create a document with enough content to generate multiple chunks
	longContent := strings.Repeat("This is a sentence that will help create multiple chunks. ", 50)
	docs := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "Long Document",
			RelativePath: "long.md",
			Content:      []byte(longContent),
			ModTime:      time.Now(),
		},
	}

	exporter := NewJSONLExporter()
	jsonlOutput, err := exporter.ToJSONL(docs, 50, 10)
	if err != nil {
		t.Fatalf("ToJSONL() error = %v", err)
	}

	// Parse all chunks
	lines := strings.Split(strings.TrimSpace(jsonlOutput), "\n")
	var chunks []ChunkMetadata
	for _, line := range lines {
		if line == "" {
			continue
		}

		var metadata ChunkMetadata
		if err := json.Unmarshal([]byte(line), &metadata); err != nil {
			t.Fatalf("Failed to unmarshal line: %v", err)
		}
		chunks = append(chunks, metadata)
	}

	if len(chunks) < 2 {
		t.Skip("Not enough chunks generated to test navigation (need at least 2)")
	}

	// Verify first chunk has no prev, but has next
	if chunks[0].PrevChunkID != "" {
		t.Error("First chunk should not have prev_chunk_id")
	}
	if chunks[0].NextChunkID == "" {
		t.Error("First chunk should have next_chunk_id")
	}

	// Verify middle chunks have both prev and next
	for i := 1; i < len(chunks)-1; i++ {
		if chunks[i].PrevChunkID == "" {
			t.Errorf("Chunk %d should have prev_chunk_id", i)
		}
		if chunks[i].NextChunkID == "" {
			t.Errorf("Chunk %d should have next_chunk_id", i)
		}

		// Verify prev/next IDs match adjacent chunks
		if chunks[i].PrevChunkID != chunks[i-1].ChunkID {
			t.Errorf("Chunk %d prev_chunk_id mismatch: got %s, want %s",
				i, chunks[i].PrevChunkID, chunks[i-1].ChunkID)
		}
		if chunks[i].NextChunkID != chunks[i+1].ChunkID {
			t.Errorf("Chunk %d next_chunk_id mismatch: got %s, want %s",
				i, chunks[i].NextChunkID, chunks[i+1].ChunkID)
		}
	}

	// Verify last chunk has prev, but no next
	lastIdx := len(chunks) - 1
	if chunks[lastIdx].PrevChunkID == "" {
		t.Error("Last chunk should have prev_chunk_id")
	}
	if chunks[lastIdx].NextChunkID != "" {
		t.Error("Last chunk should not have next_chunk_id")
	}
}

// TestToJSONL_VectorField tests that the Vector field is properly handled.
func TestToJSONL_VectorField(t *testing.T) {
	docs := []scanner.Document{
		{
			ID:           "doc1",
			Title:        "Test Document",
			RelativePath: "test.md",
			Content:      []byte("Test content."),
			ModTime:      time.Now(),
		},
	}

	exporter := NewJSONLExporter()
	jsonlOutput, err := exporter.ToJSONL(docs, 50, 10)
	if err != nil {
		t.Fatalf("ToJSONL() error = %v", err)
	}

	lines := strings.Split(strings.TrimSpace(jsonlOutput), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}

		// Parse as generic map to check raw JSON structure
		var rawJSON map[string]interface{}
		if err := json.Unmarshal([]byte(line), &rawJSON); err != nil {
			t.Fatalf("Failed to unmarshal line: %v", err)
		}

		// Vector field should be omitted if empty (omitempty tag)
		// But can be present if populated
		if vector, ok := rawJSON["vector"]; ok {
			// If present, should be an array
			if _, isArray := vector.([]interface{}); !isArray {
				t.Error("vector field should be an array")
			}
		}
	}
}

// TestJSONLStreaming tests that JSONL output can be read line-by-line using bufio.Scanner
// without loading the entire file into memory.
func TestJSONLStreaming(t *testing.T) {
	// Create a larger document to simulate streaming
	largeContent := strings.Repeat("This is a test sentence for streaming validation. ", 100)
	docs := []scanner.Document{
		{
			ID:           "large-doc",
			Title:        "Large Document",
			RelativePath: "large.md",
			Content:      []byte(largeContent),
			ModTime:      time.Now(),
		},
	}

	exporter := NewJSONLExporter()
	jsonlOutput, err := exporter.ToJSONL(docs, 50, 10)
	if err != nil {
		t.Fatalf("ToJSONL() error = %v", err)
	}

	// Simulate streaming: read line-by-line with bufio.Scanner
	reader := strings.NewReader(jsonlOutput)
	scanner := bufio.NewScanner(reader)

	lineCount := 0
	validJSONCount := 0

	for scanner.Scan() {
		line := scanner.Text()
		lineCount++

		if line == "" {
			continue
		}

		// Verify each line is valid JSON
		var metadata ChunkMetadata
		if err := json.Unmarshal([]byte(line), &metadata); err != nil {
			t.Errorf("Line %d is not valid JSON: %v", lineCount, err)
			continue
		}

		validJSONCount++

		// Verify chunk has required fields
		if metadata.ChunkID == "" {
			t.Errorf("Line %d missing chunk_id", lineCount)
		}
		if metadata.Text == "" {
			t.Errorf("Line %d missing text", lineCount)
		}
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("Error during streaming scan: %v", err)
	}

	if lineCount == 0 {
		t.Fatal("No lines read during streaming test")
	}

	if validJSONCount == 0 {
		t.Fatal("No valid JSON objects read during streaming test")
	}

	t.Logf("Successfully streamed %d lines with %d valid JSON objects", lineCount, validJSONCount)
}

// TestToJSONL_EmptyDocuments tests JSONL export with empty document list.
func TestToJSONL_EmptyDocuments(t *testing.T) {
	exporter := NewJSONLExporter()
	jsonlOutput, err := exporter.ToJSONL([]scanner.Document{}, 50, 10)
	if err != nil {
		t.Fatalf("ToJSONL() error = %v", err)
	}

	// Should return empty string for no documents
	if jsonlOutput != "" {
		t.Error("ToJSONL() should return empty string for no documents")
	}
}

// TestChunkMetadata_JSONTags tests that ChunkMetadata JSON tags are correct.
func TestChunkMetadata_JSONTags(t *testing.T) {
	metadata := ChunkMetadata{
		DocID:       "doc123",
		ChunkID:     "chunk456",
		Text:        "Sample text",
		TokenCount:  10,
		Source:      "test.md",
		StartPos:    0,
		EndPos:      11,
		PrevChunkID: "chunk455",
		NextChunkID: "chunk457",
		Vector:      []float32{0.1, 0.2, 0.3},
	}

	jsonBytes, err := json.Marshal(metadata)
	if err != nil {
		t.Fatalf("Failed to marshal ChunkMetadata: %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify expected field names in JSON
	expectedFields := []string{
		"doc_id",
		"chunk_id",
		"text",
		"token_count",
		"source",
		"start_pos",
		"end_pos",
		"prev_chunk_id",
		"next_chunk_id",
		"vector",
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, `"`+field+`"`) {
			t.Errorf("JSON output missing field: %s", field)
		}
	}
}

package chunking

import (
	"strings"
	"testing"

	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/tokenizer"
)

// TestFixedStrategy tests the FixedSizeStrategy implementation.
func TestFixedStrategy(t *testing.T) {
	tok, err := tokenizer.NewTokenizer()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	strategy := NewFixedSizeStrategy(tok)

	tests := []struct {
		name          string
		content       string
		maxTokens     int
		overlapTokens int
		wantMinChunks int
		wantMaxChunks int
	}{
		{
			name:          "small document fits in one chunk",
			content:       "This is a small test document.",
			maxTokens:     100,
			overlapTokens: 10,
			wantMinChunks: 1,
			wantMaxChunks: 1,
		},
		{
			name:          "large document splits into multiple chunks",
			content:       strings.Repeat("This is a test sentence with multiple words. ", 100),
			maxTokens:     50,
			overlapTokens: 10,
			wantMinChunks: 5,
			wantMaxChunks: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := scanner.Document{
				ID:      "test-doc",
				Content: []byte(tt.content),
			}

			chunks, err := strategy.Chunk(doc, tt.maxTokens, tt.overlapTokens)
			if err != nil {
				t.Fatalf("FixedStrategy.Chunk() error = %v", err)
			}

			if len(chunks) < tt.wantMinChunks || len(chunks) > tt.wantMaxChunks {
				t.Errorf("FixedStrategy.Chunk() got %d chunks, want between %d and %d",
					len(chunks), tt.wantMinChunks, tt.wantMaxChunks)
			}

			// Verify all chunks have valid token counts
			for i, chunk := range chunks {
				if chunk.TokenCount == 0 {
					t.Errorf("Chunk %d has zero token count", i)
				}
				if chunk.TokenCount > tt.maxTokens {
					t.Errorf("Chunk %d exceeds maxTokens: %d > %d", i, chunk.TokenCount, tt.maxTokens)
				}
			}
		})
	}
}

// TestHeaderStrategy tests the MarkdownHeaderStrategy implementation.
func TestHeaderStrategy(t *testing.T) {
	tok, err := tokenizer.NewTokenizer()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	strategy := NewMarkdownHeaderStrategy(tok)

	tests := []struct {
		name          string
		content       string
		maxTokens     int
		wantMinChunks int
		wantMaxChunks int
	}{
		{
			name: "splits at markdown headers",
			content: `# Header 1
Some content here.

## Header 2
More content here.

### Header 3
Even more content.`,
			maxTokens:     100,
			wantMinChunks: 1,
			wantMaxChunks: 4,
		},
		{
			name:          "no headers single chunk",
			content:       "Just plain text without headers.",
			maxTokens:     100,
			wantMinChunks: 1,
			wantMaxChunks: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := scanner.Document{
				ID:      "test-doc",
				Content: []byte(tt.content),
			}

			chunks, err := strategy.Chunk(doc, tt.maxTokens, 0)
			if err != nil {
				t.Fatalf("HeaderStrategy.Chunk() error = %v", err)
			}

			if len(chunks) < tt.wantMinChunks || len(chunks) > tt.wantMaxChunks {
				t.Errorf("HeaderStrategy.Chunk() got %d chunks, want between %d and %d",
					len(chunks), tt.wantMinChunks, tt.wantMaxChunks)
			}
		})
	}
}

// TestRecursiveStrategy tests the RecursiveStrategy implementation.
func TestRecursiveStrategy(t *testing.T) {
	tok, err := tokenizer.NewTokenizer()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	strategy := NewRecursiveStrategy(tok)

	tests := []struct {
		name          string
		content       string
		maxTokens     int
		wantMinChunks int
		wantMaxChunks int
	}{
		{
			name: "splits at paragraph boundaries",
			content: `First paragraph with some content.

Second paragraph with different content.

Third paragraph here too.`,
			maxTokens:     50,
			wantMinChunks: 1,
			wantMaxChunks: 4,
		},
		{
			name:          "small content single chunk",
			content:       "Short text.",
			maxTokens:     100,
			wantMinChunks: 1,
			wantMaxChunks: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := scanner.Document{
				ID:      "test-doc",
				Content: []byte(tt.content),
			}

			chunks, err := strategy.Chunk(doc, tt.maxTokens, 0)
			if err != nil {
				t.Fatalf("RecursiveStrategy.Chunk() error = %v", err)
			}

			if len(chunks) < tt.wantMinChunks || len(chunks) > tt.wantMaxChunks {
				t.Errorf("RecursiveStrategy.Chunk() got %d chunks, want between %d and %d",
					len(chunks), tt.wantMinChunks, tt.wantMaxChunks)
			}
		})
	}
}

// TestSemanticStrategy tests the SemanticStrategy stub implementation.
func TestSemanticStrategy(t *testing.T) {
	tok, err := tokenizer.NewTokenizer()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	strategy := NewSemanticStrategy(tok)

	doc := scanner.Document{
		ID:      "test-doc",
		Content: []byte("This is test content for semantic chunking."),
	}

	// Should fall back to fixed strategy
	chunks, err := strategy.Chunk(doc, 100, 10)
	if err != nil {
		t.Fatalf("SemanticStrategy.Chunk() error = %v", err)
	}

	if len(chunks) == 0 {
		t.Error("SemanticStrategy.Chunk() returned no chunks")
	}
}

// TestNewChunkStrategy tests the factory function.
func TestNewChunkStrategy(t *testing.T) {
	tok, err := tokenizer.NewTokenizer()
	if err != nil {
		t.Fatalf("Failed to create tokenizer: %v", err)
	}

	tests := []struct {
		name        string
		strategyName string
		wantErr     bool
	}{
		{"fixed strategy", "fixed", false},
		{"headers strategy", "headers", false},
		{"markdown-headers alias", "markdown-headers", false},
		{"recursive strategy", "recursive", false},
		{"semantic strategy", "semantic", false},
		{"unknown strategy", "unknown", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := NewChunkStrategy(tt.strategyName, tok)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewChunkStrategy() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && strategy == nil {
				t.Error("NewChunkStrategy() returned nil strategy")
			}
		})
	}
}

package chunking

import (
	"strings"
	"testing"

	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/tokenizer"
)

// Benchmark setup helpers
var (
	benchDoc     scanner.Document
	benchTok     tokenizer.Tokenizer
	benchContent string
)

func init() {
	// Create realistic test content (approximately 10KB)
	benchContent = strings.Repeat("This is a sample sentence used for benchmarking chunking strategies. "+
		"It contains multiple words and punctuation to simulate real documentation. "+
		"We want to measure the performance characteristics of different approaches.\n\n", 200)

	benchDoc = scanner.Document{
		ID:      "bench-doc",
		Content: []byte(benchContent),
	}

	var err error
	benchTok, err = tokenizer.NewTokenizer()
	if err != nil {
		panic(err)
	}
}

// BenchmarkFixedStrategy benchmarks the FixedSizeStrategy.
func BenchmarkFixedStrategy(b *testing.B) {
	strategy := NewFixedSizeStrategy(benchTok)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := strategy.Chunk(benchDoc, 512, 128)
		if err != nil {
			b.Fatalf("FixedStrategy.Chunk() error = %v", err)
		}
	}
}

// BenchmarkHeaderStrategy benchmarks the MarkdownHeaderStrategy.
func BenchmarkHeaderStrategy(b *testing.B) {
	// Add headers to content for more realistic test
	headerContent := `# Main Header

` + benchContent + `

## Subheader 1

` + benchContent + `

## Subheader 2

` + benchContent

	doc := scanner.Document{
		ID:      "bench-doc-headers",
		Content: []byte(headerContent),
	}

	strategy := NewMarkdownHeaderStrategy(benchTok)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := strategy.Chunk(doc, 512, 128)
		if err != nil {
			b.Fatalf("HeaderStrategy.Chunk() error = %v", err)
		}
	}
}

// BenchmarkRecursiveStrategy benchmarks the RecursiveStrategy.
func BenchmarkRecursiveStrategy(b *testing.B) {
	strategy := NewRecursiveStrategy(benchTok)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := strategy.Chunk(benchDoc, 512, 128)
		if err != nil {
			b.Fatalf("RecursiveStrategy.Chunk() error = %v", err)
		}
	}
}

// BenchmarkSemanticStrategy benchmarks the SemanticStrategy (currently a fallback).
func BenchmarkSemanticStrategy(b *testing.B) {
	strategy := NewSemanticStrategy(benchTok)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, err := strategy.Chunk(benchDoc, 512, 128)
		if err != nil {
			b.Fatalf("SemanticStrategy.Chunk() error = %v", err)
		}
	}
}

// BenchmarkNewChunkStrategyFactory benchmarks the factory function overhead.
func BenchmarkNewChunkStrategyFactory(b *testing.B) {
	strategies := []string{"fixed", "headers", "recursive", "semantic"}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		stratName := strategies[i%len(strategies)]
		_, err := NewChunkStrategy(stratName, benchTok)
		if err != nil {
			b.Fatalf("NewChunkStrategy() error = %v", err)
		}
	}
}

// BenchmarkTokenCounting measures tokenization overhead.
func BenchmarkTokenCounting(b *testing.B) {
	sampleText := "This is a sample text for token counting benchmarks."

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = benchTok.Count(sampleText)
	}
}

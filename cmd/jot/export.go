// Package main is the entry point for the Jot CLI application.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/onedusk/jot/internal/export"
	"github.com/onedusk/jot/internal/scanner"
)

// exportCmd provides the command for exporting documentation into various formats
// such as JSON or a format optimized for Large Language Models (LLMs).
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export documentation in various formats",
	Long: `Export documentation to multiple formats optimized for different use cases.

Supported formats:
  - json:       Standard JSON export with full document metadata
  - yaml:       YAML format for human-readable configuration
  - llms-txt:   Lightweight index per llmstxt.org specification
  - llms-full:  Complete documentation concatenated for LLM context
  - jsonl:      JSON Lines format for vector database ingestion
  - markdown:   Enriched markdown with YAML frontmatter

Chunking strategies:
  - fixed:            Fixed-size token chunks with word boundaries (default)
  - semantic:         Semantic boundary detection for natural breaks
  - markdown-headers: Split at markdown header boundaries
  - recursive:        Hierarchical splitting (paragraph->line->space->char)
  - contextual:       Context-aware chunking (alias for semantic)

Examples:
  # Export to llms.txt format
  jot export --format llms-txt --output llms.txt

  # Export for RAG with optimal settings
  jot export --for-rag --output docs.jsonl

  # Export markdown with header-based chunking
  jot export --format markdown --strategy markdown-headers --output docs.md

  # Export JSONL with custom chunk size
  jot export --format jsonl --chunk-size 1024 --chunk-overlap 256 --output chunks.jsonl

  # Export with embeddings (warning: API costs apply)
  jot export --format jsonl --include-embeddings --output embeddings.jsonl`,
	RunE: runExport,
}

func init() {
	// Format selection
	exportCmd.Flags().StringP("format", "f", "json", "export format: json, yaml, llms-txt, llms-full, jsonl, markdown")
	exportCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")

	// Chunking configuration
	exportCmd.Flags().StringP("strategy", "s", "fixed", "chunking strategy: fixed, semantic, markdown-headers, recursive, contextual")
	exportCmd.Flags().IntP("chunk-size", "", 512, "maximum tokens per chunk (must be >0 and <=2048)")
	exportCmd.Flags().IntP("chunk-overlap", "", 128, "token overlap between chunks (must be >0 and <=2048)")

	// Preset configurations
	exportCmd.Flags().Bool("for-rag", false, "preset for RAG: jsonl format + semantic strategy + 512 tokens")
	exportCmd.Flags().Bool("for-context", false, "preset for context: markdown format + headers strategy + 1024 tokens")
	exportCmd.Flags().Bool("for-training", false, "preset for training: jsonl format + fixed strategy + 256 tokens")

	// Advanced options
	exportCmd.Flags().Bool("include-embeddings", false, "generate embeddings for JSONL format (warning: API costs apply)")

	rootCmd.AddCommand(exportCmd)
}

// validateExportFlags validates export command flags for mutual exclusivity and logical consistency.
// Returns an error with helpful examples if validation fails.
func validateExportFlags(cmd *cobra.Command) error {
	// Get preset flags
	forRAG, _ := cmd.Flags().GetBool("for-rag")
	forContext, _ := cmd.Flags().GetBool("for-context")
	forTraining, _ := cmd.Flags().GetBool("for-training")

	// Check for mutually exclusive presets
	presetCount := 0
	if forRAG {
		presetCount++
	}
	if forContext {
		presetCount++
	}
	if forTraining {
		presetCount++
	}

	if presetCount > 1 {
		return fmt.Errorf("cannot use multiple presets together (--for-rag, --for-context, --for-training are mutually exclusive)\n\nExamples:\n  jot export --for-rag --output docs.jsonl\n  jot export --for-context --output docs.md")
	}

	// Get chunking parameters
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")
	chunkOverlap, _ := cmd.Flags().GetInt("chunk-overlap")

	// Validate chunk-size
	if chunkSize <= 0 || chunkSize > 2048 {
		return fmt.Errorf("chunk-size must be >0 and <=2048 (got %d)\n\nExample:\n  jot export --format jsonl --chunk-size 512 --output docs.jsonl", chunkSize)
	}

	// Validate chunk-overlap
	if chunkOverlap < 0 || chunkOverlap > 2048 {
		return fmt.Errorf("chunk-overlap must be >=0 and <=2048 (got %d)\n\nExample:\n  jot export --format jsonl --chunk-overlap 128 --output docs.jsonl", chunkOverlap)
	}

	// Validate overlap is less than chunk size
	if chunkOverlap >= chunkSize {
		return fmt.Errorf("chunk-overlap (%d) must be less than chunk-size (%d)\n\nExample:\n  jot export --format jsonl --chunk-size 512 --chunk-overlap 128 --output docs.jsonl", chunkOverlap, chunkSize)
	}

	// Validate format
	format, _ := cmd.Flags().GetString("format")
	validFormats := []string{"json", "yaml", "llms-txt", "llms-full", "jsonl", "markdown"}
	isValidFormat := false
	for _, vf := range validFormats {
		if format == vf {
			isValidFormat = true
			break
		}
	}
	if !isValidFormat {
		return fmt.Errorf("unsupported format: %s (supported: json, yaml, llms-txt, llms-full, jsonl, markdown)\n\nExample:\n  jot export --format llms-txt --output llms.txt", format)
	}

	// Validate strategy
	strategy, _ := cmd.Flags().GetString("strategy")
	validStrategies := []string{"fixed", "semantic", "markdown-headers", "recursive", "contextual"}
	isValidStrategy := false
	for _, vs := range validStrategies {
		if strategy == vs {
			isValidStrategy = true
			break
		}
	}
	if !isValidStrategy {
		return fmt.Errorf("unsupported strategy: %s (supported: fixed, semantic, markdown-headers, recursive, contextual)\n\nExample:\n  jot export --format jsonl --strategy semantic --output docs.jsonl", strategy)
	}

	// Warn if include-embeddings is used with non-JSONL format
	includeEmbeddings, _ := cmd.Flags().GetBool("include-embeddings")
	if includeEmbeddings && format != "jsonl" {
		fmt.Fprintf(os.Stderr, "Warning: --include-embeddings only applies to JSONL format (current format: %s)\n", format)
	}

	return nil
}

// runExport executes the logic for the export command.
func runExport(cmd *cobra.Command, args []string) error {
	// Validate flags
	if err := validateExportFlags(cmd); err != nil {
		return err
	}

	// Get flags
	format, _ := cmd.Flags().GetString("format")
	outputFile, _ := cmd.Flags().GetString("output")
	strategy, _ := cmd.Flags().GetString("strategy")
	chunkSize, _ := cmd.Flags().GetInt("chunk-size")
	chunkOverlap, _ := cmd.Flags().GetInt("chunk-overlap")
	includeEmbeddings, _ := cmd.Flags().GetBool("include-embeddings")

	// Apply preset configurations (override individual flags)
	forRAG, _ := cmd.Flags().GetBool("for-rag")
	forContext, _ := cmd.Flags().GetBool("for-context")
	forTraining, _ := cmd.Flags().GetBool("for-training")

	if forRAG {
		format = "jsonl"
		strategy = "semantic"
		chunkSize = 512
		chunkOverlap = 128
		fmt.Println(" Using RAG preset: jsonl format, semantic strategy, 512 token chunks")
	} else if forContext {
		format = "markdown"
		strategy = "markdown-headers"
		chunkSize = 1024
		chunkOverlap = 256
		fmt.Println(" Using context preset: markdown format, headers strategy, 1024 token chunks")
	} else if forTraining {
		format = "jsonl"
		strategy = "fixed"
		chunkSize = 256
		chunkOverlap = 64
		fmt.Println(" Using training preset: jsonl format, fixed strategy, 256 token chunks")
	}

	// Load configuration
	config := loadBuildConfig(cmd)

	fmt.Println(" Scanning for markdown files...")

	var allDocs []scanner.Document
	for _, inputPath := range config.InputPaths {
		// Check if path exists
		if _, err := os.Stat(inputPath); err != nil {
			continue
		}

		// Create scanner
		s, err := scanner.NewScanner(inputPath, config.IgnorePatterns)
		if err != nil {
			return fmt.Errorf("failed to create scanner: %w", err)
		}

		// Scan documents
		docs, err := s.Scan()
		if err != nil {
			return fmt.Errorf("failed to scan %s: %w", inputPath, err)
		}

		allDocs = append(allDocs, docs...)
	}

	if len(allDocs) == 0 {
		return fmt.Errorf("no markdown files found")
	}

	fmt.Printf("  Found %d markdown files\n\n", len(allDocs))

	// Log embeddings warning if applicable
	if includeEmbeddings && format == "jsonl" {
		fmt.Println(" WARNING: --include-embeddings will generate embeddings using external API")
		fmt.Println(" This may incur costs and take significant time depending on document size")
		fmt.Println("")
	}

	// Create exporter
	exporter := export.NewExporter()

	var output string
	var err error

	// Export based on format
	switch format {
	case "json":
		fmt.Println(" Exporting to JSON...")
		output, err = exporter.ToJSON(allDocs)

	case "yaml":
		fmt.Println(" Exporting to YAML...")
		output, err = exporter.ToYAML(allDocs)

	case "llms-txt":
		fmt.Println(" Exporting to llms.txt format...")
		llmsTxtExporter := export.NewLLMSTxtExporter()
		projectConfig := export.ProjectConfig{
			Name:        config.ProjectName,
			Description: config.ProjectDescription,
		}
		output, err = llmsTxtExporter.ToLLMSTxt(allDocs, projectConfig)

	case "llms-full":
		fmt.Println(" Exporting to llms-full.txt format...")
		llmsTxtExporter := export.NewLLMSTxtExporter()
		projectConfig := export.ProjectConfig{
			Name:        config.ProjectName,
			Description: config.ProjectDescription,
		}
		output, err = llmsTxtExporter.ToLLMSFullTxt(allDocs, projectConfig)

	case "jsonl":
		fmt.Printf(" Exporting to JSONL format (strategy: %s, chunk-size: %d, overlap: %d)...\n", strategy, chunkSize, chunkOverlap)
		jsonlExporter := export.NewJSONLExporter()
		output, err = jsonlExporter.ToJSONL(allDocs, chunkSize, chunkOverlap)

	case "markdown":
		fmt.Printf(" Exporting to enriched markdown (strategy: %s, chunk-size: %d)...\n", strategy, chunkSize)
		markdownExporter, mdErr := export.NewMarkdownExporter()
		if mdErr != nil {
			err = mdErr
		} else {
			output, err = markdownExporter.ToEnrichedMarkdown(allDocs, false)
		}

	case "llm":
		// Legacy format - keep for backward compatibility
		fmt.Println(" Exporting for LLM consumption (legacy format)...")
		llmData, llmErr := exporter.ToLLMFormat(allDocs)
		if llmErr != nil {
			err = llmErr
		} else {
			jsonBytes, _ := json.MarshalIndent(llmData, "", "  ")
			output = string(jsonBytes)
		}

	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	if err != nil {
		return fmt.Errorf("failed to export: %w", err)
	}

	// Write output
	if outputFile != "" {
		// Ensure directory exists
		dir := filepath.Dir(outputFile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf(" Exported to %s\n", outputFile)
	} else {
		// Write to stdout
		fmt.Println(output)
	}

	return nil
}

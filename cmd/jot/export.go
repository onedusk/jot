// Package main is the entry point for the Jot CLI application.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/thrive/jot/internal/export"
	"github.com/thrive/jot/internal/scanner"
)

// exportCmd provides the command for exporting documentation into various formats
// such as JSON or a format optimized for Large Language Models (LLMs).
var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export documentation in various formats",
	Long:  `Export documentation to JSON, YAML, or LLM-optimized formats.`,
	RunE:  runExport,
}

func init() {
	exportCmd.Flags().StringP("format", "f", "json", "export format (json, yaml, llm)")
	exportCmd.Flags().StringP("output", "o", "", "output file (default: stdout)")
	rootCmd.AddCommand(exportCmd)
}

// runExport executes the logic for the export command.
func runExport(cmd *cobra.Command, args []string) error {
	// Get format
	format, _ := cmd.Flags().GetString("format")
	outputFile, _ := cmd.Flags().GetString("output")

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

	// Create exporter
	exporter := export.NewExporter()

	var output string
	var err error

	// Export based on format
	switch format {
	case "json":
		fmt.Println(" Exporting to JSON...")
		output, err = exporter.ToJSON(allDocs)
	case "llm":
		fmt.Println(" Exporting for LLM consumption...")
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

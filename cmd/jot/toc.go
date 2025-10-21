// Package main is the entry point for the Jot CLI application.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/toc"
	"github.com/spf13/cobra"
)

// tocCmd provides the command for generating toc.xml files in each directory
// containing markdown files, along with a master toc.json index.
var tocCmd = &cobra.Command{
	Use:   "toc",
	Short: "Generate toc.xml in each directory with markdown files",
	Long: `Generate toc.xml files in each directory containing markdown files,
and create a master toc.json index at the project root.

The toc.json acts like a lock file, tracking all generated toc.xml locations.`,
	RunE: runTOC,
}

func init() {
	tocCmd.Flags().BoolP("dry-run", "d", false, "preview without writing files")
	tocCmd.Flags().BoolP("recursive", "r", false, "include subdirectories in each toc.xml")
}

// runTOC executes the logic for the toc command.
func runTOC(cmd *cobra.Command, args []string) error {
	// Get flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	recursive, _ := cmd.Flags().GetBool("recursive")

	// Load configuration
	config := loadBuildConfig(cmd)

	if dryRun {
		fmt.Println(" [DRY RUN MODE] - No files will be written")
	}

	fmt.Println(" Scanning for directories with markdown files...")

	// Scan directories and group documents
	dirMap, err := scanDirectoriesWithMarkdown(config.InputPaths, config.IgnorePatterns)
	if err != nil {
		return fmt.Errorf("failed to scan directories: %w", err)
	}

	if len(dirMap) == 0 {
		return fmt.Errorf("no directories with markdown files found")
	}

	fmt.Printf("  Found %d directories with markdown files\n\n", len(dirMap))

	// Generate TOC files for each directory
	fmt.Println(" Generating toc.xml files...")
	tocPaths, err := generateDirectoryTOCs(dirMap, recursive, dryRun)
	if err != nil {
		return fmt.Errorf("failed to generate TOC files: %w", err)
	}

	fmt.Printf("  Generated %d toc.xml files\n\n", len(tocPaths))

	// Generate master toc.json index
	fmt.Println(" Creating master toc.json index...")
	projectRoot, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	if err := generateMasterTOCIndex(tocPaths, projectRoot, dryRun); err != nil {
		return fmt.Errorf("failed to generate master index: %w", err)
	}

	if dryRun {
		fmt.Println("\n [DRY RUN COMPLETE] - No files were written")
	} else {
		fmt.Printf("  Created toc.json at project root\n\n")
		fmt.Printf(" TOC generation completed successfully!\n")
	}

	return nil
}

// scanDirectoriesWithMarkdown walks input paths and groups documents by their parent directory.
func scanDirectoriesWithMarkdown(paths []string, ignorePatterns []string) (map[string][]scanner.Document, error) {
	dirMap := make(map[string][]scanner.Document)

	for _, inputPath := range paths {
		// Check if path exists
		if _, err := os.Stat(inputPath); err != nil {
			fmt.Printf("  Skipping %s: %v\n", inputPath, err)
			continue
		}

		// Create scanner
		s, err := scanner.NewScanner(inputPath, ignorePatterns)
		if err != nil {
			return nil, fmt.Errorf("failed to create scanner for %s: %w", inputPath, err)
		}

		// Scan documents
		docs, err := s.Scan()
		if err != nil {
			return nil, fmt.Errorf("failed to scan %s: %w", inputPath, err)
		}

		// Group by directory
		for _, doc := range docs {
			dir := filepath.Dir(doc.Path)
			dirMap[dir] = append(dirMap[dir], doc)
		}
	}

	return dirMap, nil
}

// generateDirectoryTOCs creates toc.xml files for each directory.
func generateDirectoryTOCs(dirMap map[string][]scanner.Document, recursive bool, dryRun bool) ([]string, error) {
	var tocPaths []string

	for dirPath, docs := range dirMap {
		// Filter documents based on recursive flag
		var filteredDocs []scanner.Document
		if recursive {
			// Include all documents in this directory and subdirectories
			filteredDocs = docs
		} else {
			// Include only documents directly in this directory
			for _, doc := range docs {
				if filepath.Dir(doc.Path) == dirPath {
					filteredDocs = append(filteredDocs, doc)
				}
			}
		}

		// Skip if no documents after filtering
		if len(filteredDocs) == 0 {
			continue
		}

		// Build TOC
		builder := toc.NewBuilder()
		tableOfContents := builder.Build(filteredDocs)

		// Generate toc.xml path
		tocPath := filepath.Join(dirPath, "toc.xml")

		// Write toc.xml file
		if !dryRun {
			if err := os.WriteFile(tocPath, []byte(tableOfContents.ToXML()), 0644); err != nil {
				return nil, fmt.Errorf("failed to write %s: %w", tocPath, err)
			}
		}

		tocPaths = append(tocPaths, tocPath)
		fmt.Printf("  %s (%d files)\n", tocPath, len(filteredDocs))
	}

	return tocPaths, nil
}

// TOCIndex represents the master index structure.
type TOCIndex struct {
	TOCFiles []string `json:"toc_files"`
}

// generateMasterTOCIndex creates a toc.json file at the project root with relative paths to all toc.xml files.
func generateMasterTOCIndex(tocPaths []string, projectRoot string, dryRun bool) error {
	// Calculate relative paths from project root
	var relativePaths []string
	for _, tocPath := range tocPaths {
		relPath, err := filepath.Rel(projectRoot, tocPath)
		if err != nil {
			return fmt.Errorf("failed to calculate relative path for %s: %w", tocPath, err)
		}
		// Normalize to forward slashes for cross-platform compatibility
		relPath = filepath.ToSlash(relPath)
		relativePaths = append(relativePaths, relPath)
	}

	// Create index structure
	index := TOCIndex{
		TOCFiles: relativePaths,
	}

	// Marshal to JSON
	jsonData, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	// Write to toc.json
	indexPath := filepath.Join(projectRoot, "toc.json")
	if !dryRun {
		if err := os.WriteFile(indexPath, jsonData, 0644); err != nil {
			return fmt.Errorf("failed to write toc.json: %w", err)
		}
	}

	if verbose || dryRun {
		fmt.Printf("  Index contains %d toc.xml references\n", len(relativePaths))
	}

	return nil
}

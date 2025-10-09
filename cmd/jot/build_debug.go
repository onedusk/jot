// Package main is the entry point for the Jot CLI application.
package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/thrive/jot/internal/scanner"
)

// debugCmd provides a command for debugging the document scanning process.
// It prints the documents that are found without performing a full build.
var debugCmd = &cobra.Command{
	Use:   "debug",
	Short: "Debug build process",
	RunE:  runDebug,
}

func init() {
	rootCmd.AddCommand(debugCmd)
}

// runDebug executes the debug command logic.
func runDebug(cmd *cobra.Command, args []string) error {
	config := loadBuildConfig(cmd)

	fmt.Println("Debug: Scanning documents...")

	// Scan for documents
	allDocs := []scanner.Document{}
	for _, inputPath := range config.InputPaths {
		s, err := scanner.NewScanner(inputPath, config.IgnorePatterns)
		if err != nil {
			return err
		}

		docs, err := s.Scan()
		if err != nil {
			return err
		}

		for _, doc := range docs {
			fmt.Printf("  Found: %q (path: %q)\n", doc.RelativePath, doc.Path)
		}

		allDocs = append(allDocs, docs...)
	}

	return nil
}

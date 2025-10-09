// Package main is the entry point for the Jot CLI application.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/thrive/jot/internal/compiler"
	"github.com/thrive/jot/internal/scanner"
	"github.com/thrive/jot/internal/toc"
)

// buildCmd represents the command for building the documentation.
// It scans for markdown files, generates a table of contents, and compiles
// the documentation into an HTML website.
var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build documentation from markdown files",
	Long:  `Scan markdown files and generate a static documentation website.`,
	RunE:  runBuild,
}

func init() {
	buildCmd.Flags().StringP("output", "o", "", "output directory (overrides config)")
	buildCmd.Flags().BoolP("clean", "c", false, "clean output directory before building")
}

// runBuild executes the main build logic for the documentation.
func runBuild(cmd *cobra.Command, args []string) error {
	start := time.Now()

	// Load configuration
	config := loadBuildConfig(cmd)

	// Clean output directory if requested
	if config.Clean {
		if err := os.RemoveAll(config.OutputPath); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}

	// Create output directory
	if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	fmt.Println(" Scanning for markdown files...")

	var allDocs []scanner.Document
	for _, inputPath := range config.InputPaths {
		fmt.Printf("  Scanning %s...\n", inputPath)

		// Check if path exists
		if _, err := os.Stat(inputPath); err != nil {
			fmt.Printf("    Skipping %s: %v\n", inputPath, err)
			continue
		}

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

	// Generate table of contents
	fmt.Println(" Generating table of contents...")
	tocBuilder := toc.NewBuilder()
	tableOfContents := tocBuilder.Build(allDocs)
	tocPath := filepath.Join(config.OutputPath, "toc.xml")
	if err := os.WriteFile(tocPath, []byte(tableOfContents.ToXML()), 0644); err != nil {
		return fmt.Errorf("failed to write TOC: %w", err)
	}
	fmt.Printf("  Created %s\n\n", tocPath)

	// Compile to HTML
	fmt.Println(" Compiling to HTML...")
	comp := compiler.NewCompiler(config.OutputPath)
	if err := comp.Compile(allDocs, tableOfContents); err != nil {
		return fmt.Errorf("failed to compile documents: %w", err)
	}
	fmt.Printf("  Generated %d HTML files\n", len(allDocs))
	fmt.Printf("  Created assets/styles.css\n\n")

	// Summary
	elapsed := time.Since(start)
	fmt.Printf(" Build completed in %.2fs\n", elapsed.Seconds())

	return nil
}

// BuildConfig holds the configuration settings for the build process,
// combining values from the config file and command-line flags.
type BuildConfig struct {
	InputPaths     []string
	OutputPath     string
	IgnorePatterns []string
	Clean          bool
}

// loadBuildConfig loads the build configuration from Viper and overrides it with
// any values provided via command-line flags. It also sets default values.
func loadBuildConfig(cmd *cobra.Command) BuildConfig {
	config := BuildConfig{
		InputPaths:     viper.GetStringSlice("input.paths"),
		OutputPath:     viper.GetString("output.path"),
		IgnorePatterns: viper.GetStringSlice("input.ignore"),
		Clean:          viper.GetBool("output.clean"),
	}

	// Override with command flags
	if output, _ := cmd.Flags().GetString("output"); output != "" {
		config.OutputPath = output
	}
	if clean, _ := cmd.Flags().GetBool("clean"); clean {
		config.Clean = true
	}

	// Defaults
	if len(config.InputPaths) == 0 {
		config.InputPaths = []string{"."}
	}
	if config.OutputPath == "" {
		config.OutputPath = "./dist"
	}

	return config
}

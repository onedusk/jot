// Package main is the entry point for the Jot CLI application.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/onedusk/jot/internal/compiler"
	"github.com/onedusk/jot/internal/export"
	"github.com/onedusk/jot/internal/scanner"
	"github.com/onedusk/jot/internal/toc"
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
	buildCmd.Flags().Bool("skip-llms-txt", false, "skip generation of llms.txt and llms-full.txt files")
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

	// Generate llms.txt and llms-full.txt
	if config.GenerateLLMSTxt {
		fmt.Println(" Generating llms.txt...")

		// Create project config from viper settings
		projectConfig := export.ProjectConfig{
			Name:        viper.GetString("project.name"),
			Description: viper.GetString("project.description"),
		}

		// Set defaults if not configured
		if projectConfig.Name == "" {
			projectConfig.Name = "Documentation"
		}
		if projectConfig.Description == "" {
			projectConfig.Description = "Project documentation"
		}

		exporter := export.NewLLMSTxtExporter()

		// Generate llms.txt
		llmsTxt, err := exporter.ToLLMSTxt(allDocs, projectConfig)
		if err != nil {
			fmt.Printf("  Warning: failed to generate llms.txt: %v\n", err)
		} else {
			llmsTxtPath := filepath.Join(config.OutputPath, "llms.txt")
			if err := os.WriteFile(llmsTxtPath, []byte(llmsTxt), 0644); err != nil {
				fmt.Printf("  Warning: failed to write llms.txt: %v\n", err)
			} else {
				llmsTxtSize := len(llmsTxt)
				fmt.Printf("  Created llms.txt (%s)\n", humanizeBytes(llmsTxtSize))
			}
		}

		// Generate llms-full.txt
		llmsFullTxt, err := exporter.ToLLMSFullTxt(allDocs, projectConfig)
		if err != nil {
			fmt.Printf("  Warning: failed to generate llms-full.txt: %v\n", err)
		} else {
			llmsFullTxtPath := filepath.Join(config.OutputPath, "llms-full.txt")
			if err := os.WriteFile(llmsFullTxtPath, []byte(llmsFullTxt), 0644); err != nil {
				fmt.Printf("  Warning: failed to write llms-full.txt: %v\n", err)
			} else {
				llmsFullTxtSize := len(llmsFullTxt)
				fmt.Printf("  Created llms-full.txt (%s)\n", humanizeBytes(llmsFullTxtSize))
			}
		}

		fmt.Println()
	}

	// Summary
	elapsed := time.Since(start)
	fmt.Printf(" Build completed in %.2fs\n", elapsed.Seconds())

	return nil
}

// BuildConfig holds the configuration settings for the build process,
// combining values from the config file and command-line flags.
type BuildConfig struct {
	InputPaths         []string
	OutputPath         string
	IgnorePatterns     []string
	Clean              bool
	GenerateLLMSTxt    bool
	ProjectName        string
	ProjectDescription string
}

// loadBuildConfig loads the build configuration from Viper and overrides it with
// any values provided via command-line flags. It also sets default values.
func loadBuildConfig(cmd *cobra.Command) BuildConfig {
	config := BuildConfig{
		InputPaths:         viper.GetStringSlice("input.paths"),
		OutputPath:         viper.GetString("output.path"),
		IgnorePatterns:     viper.GetStringSlice("input.ignore"),
		Clean:              viper.GetBool("output.clean"),
		GenerateLLMSTxt:    true, // Default to true
		ProjectName:        viper.GetString("project.name"),
		ProjectDescription: viper.GetString("project.description"),
	}

	// Read llm_export from config if explicitly set
	if viper.IsSet("features.llm_export") {
		config.GenerateLLMSTxt = viper.GetBool("features.llm_export")
	}

	// Override with command flags
	if output, _ := cmd.Flags().GetString("output"); output != "" {
		config.OutputPath = output
	}
	if clean, _ := cmd.Flags().GetBool("clean"); clean {
		config.Clean = true
	}
	if skipLLMSTxt, _ := cmd.Flags().GetBool("skip-llms-txt"); skipLLMSTxt {
		config.GenerateLLMSTxt = false
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

// humanizeBytes converts a byte count to a human-readable string (e.g., "15KB", "2.3MB")
func humanizeBytes(bytes int) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	units := []string{"KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f%s", float64(bytes)/float64(div), units[exp])
}

// Package main is the entry point for the Jot CLI application.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/onedusk/jot/internal/compiler"
	"github.com/onedusk/jot/internal/renderer"
	"github.com/onedusk/jot/internal/toc"
	"github.com/onedusk/jot/pkg/export"
	"github.com/onedusk/jot/pkg/scanner"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		if err := s.Exclude(config.OutputPath); err != nil {
			return fmt.Errorf("failed to exclude output directory: %w", err)
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
	if err := checkDuplicatePaths(allDocs); err != nil {
		return err
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
	siteConfig := renderer.SiteConfig{
		ProjectName: config.ProjectName,
		NavLinks: []renderer.NavLink{
			{Label: "API", Href: "#"},
			{Label: "Documentation", Href: "#"},
			{Label: "Support", Href: "#"},
		},
	}
	if siteConfig.ProjectName == "" {
		siteConfig.ProjectName = "Documentation"
	}
	comp := compiler.NewCompiler(config.OutputPath, siteConfig)
	if err := comp.Compile(allDocs, tableOfContents); err != nil {
		return fmt.Errorf("failed to compile documents: %w", err)
	}
	fmt.Printf("  Generated %d HTML files\n", len(allDocs))
	fmt.Printf("  Wrote static assets to %s\n\n", filepath.Join(config.OutputPath, "assets"))

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

// checkDuplicatePaths returns an error if two documents share a relative path.
// Paths are relative to each input root, so files from different roots can
// collide and would otherwise overwrite each other in the output.
func checkDuplicatePaths(docs []scanner.Document) error {
	seen := make(map[string]string, len(docs))
	for _, doc := range docs {
		if first, ok := seen[doc.RelativePath]; ok {
			return fmt.Errorf("%s and %s both map to %s in the output; rename one of them or remove one of the input paths",
				displayPath(first), displayPath(doc.Path), doc.RelativePath)
		}
		seen[doc.RelativePath] = doc.Path
	}
	return nil
}

// displayPath returns path relative to the working directory when possible.
func displayPath(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(wd, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}
	return rel
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

// loadBuildConfig loads the configuration from Viper and overrides it with the
// build command's flags. Only the build command should call it: other commands
// give --output a different meaning and should use loadConfig instead.
func loadBuildConfig(cmd *cobra.Command) BuildConfig {
	config := loadConfig()

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

	return config
}

// loadConfig loads the configuration from Viper and applies default values.
func loadConfig() BuildConfig {
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

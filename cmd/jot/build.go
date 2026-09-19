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
	allDocs, err := dedupeDocuments(allDocs)
	if err != nil {
		return err
	}

	// Touch the output directory only once the input is known to be valid, so a
	// failed build leaves the previous output in place
	if config.Clean {
		if err := checkCleanIsSafe(config.OutputPath, config.InputPaths); err != nil {
			return err
		}
		if err := os.RemoveAll(config.OutputPath); err != nil {
			return fmt.Errorf("failed to clean output directory: %w", err)
		}
	}
	if err := os.MkdirAll(config.OutputPath, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
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
	fmt.Printf("  Rendered %d documents to HTML\n", len(allDocs))
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

// dedupeDocuments drops repeat scans of the same source file (an input path
// listed twice, or a file that is also inside another input path) and returns
// an error if two different files would be written to the same output path.
// Relative paths are computed per input root, so files from different roots can
// collide. They are compared case-insensitively because on the default macOS
// and Windows filesystems README.md and readme.md are the same file.
func dedupeDocuments(docs []scanner.Document) ([]scanner.Document, error) {
	unique := make([]scanner.Document, 0, len(docs))
	seenSource := make(map[string]bool, len(docs))
	seenOutput := make(map[string]scanner.Document, len(docs))
	for _, doc := range docs {
		if seenSource[doc.Path] {
			continue
		}
		seenSource[doc.Path] = true

		key := strings.ToLower(doc.RelativePath)
		if first, ok := seenOutput[key]; ok {
			msg := fmt.Sprintf("%s and %s both map to %s in the output",
				displayPath(first.Path), displayPath(doc.Path), first.RelativePath)
			if first.RelativePath != doc.RelativePath {
				msg += " (paths that differ only in case collide on case-insensitive filesystems)"
			}
			return nil, fmt.Errorf("%s; rename one of them or remove one of the input paths", msg)
		}
		seenOutput[key] = doc
		unique = append(unique, doc)
	}
	return unique, nil
}

// checkCleanIsSafe returns an error if removing outputDir would also remove an
// input path, which happens when the output directory is, or contains, a source
// directory.
func checkCleanIsSafe(outputDir string, inputPaths []string) error {
	out, err := resolvePath(outputDir)
	if err != nil {
		// Nothing to remove
		return nil
	}
	for _, inputPath := range inputPaths {
		in, err := resolvePath(inputPath)
		if err != nil {
			continue
		}
		rel, err := filepath.Rel(out, in)
		if err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return fmt.Errorf("refusing to clean output directory %s because it contains input path %s; choose a separate output directory or disable clean", outputDir, inputPath)
		}
	}
	return nil
}

// resolvePath returns the absolute path of an existing file with symlinks resolved.
func resolvePath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(abs)
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

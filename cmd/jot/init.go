// Package main is the entry point for the Jot CLI application.
package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// initCmd provides the command to initialize a new Jot project.
// It creates a default configuration file and example directory structure.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Jot project",
	Long:  `Initialize a new Jot project in the current directory with a default configuration file.`,
	RunE:  runInit,
}

// runInit executes the logic for the init command.
func runInit(cmd *cobra.Command, args []string) error {
	// Check if jot.yml already exists
	if _, err := os.Stat("jot.yml"); err == nil {
		return fmt.Errorf("jot.yml already exists in current directory")
	}

	// Default configuration
	defaultConfig := `# Jot Configuration File
version: 1.0
project:
  name: "My Documentation"
  description: "Project documentation"
  author: "Your Name"

input:
  paths:
    - "docs"
    - "README.md"
  ignore:
    - "**/_*.md"
    - "**/drafts/**"
    - "**/.git/**"
    - "**/node_modules/**"

output:
  path: "dist"
  format: "html"
  theme: "default"
  clean: true

features:
  search: true
  versioning: true
  llm_export: true
  syntax_highlighting: true
  auto_toc: true
  
server:
  port: 8080
  auto_reload: true
  open_browser: true

llm:
  export_formats:
    - json
    - yaml
  chunk_size: 512
  overlap: 128
  
search:
  enable: true
  index_path: "dist/search-index.json"
  fuzzy: true
  highlight: true
`

	// Write config file
	if err := os.WriteFile("jot.yml", []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to create jot.yml: %w", err)
	}
	fmt.Println("  Created jot.yml")

	// Create .jotignore file only if it doesn't exist
	if _, err := os.Stat(".jotignore"); os.IsNotExist(err) {
		ignoreContent := `# Jot ignore patterns
# Similar to .gitignore syntax

# Hidden files and directories
.*
!.jotignore

# Temporary files
*.tmp
*.temp
*.swp
*.swo
*~

# Build output
/dist/
/build/
/out/

# Dependencies
node_modules/
vendor/

# IDE files
.idea/
.vscode/
*.iml

# OS files
.DS_Store
Thumbs.db
`

		if err := os.WriteFile(".jotignore", []byte(ignoreContent), 0644); err != nil {
			return fmt.Errorf("failed to create .jotignore: %w", err)
		}
		fmt.Println("  Created .jotignore")
	} else {
		fmt.Println("  Skipped .jotignore (already exists)")
	}

	// Create example docs directory if it doesn't exist
	if _, err := os.Stat("docs"); os.IsNotExist(err) {
		if err := os.MkdirAll("docs", 0755); err != nil {
			return fmt.Errorf("failed to create docs directory: %w", err)
		}
		fmt.Println("  Created docs/ directory")
	} else {
		fmt.Println("  Skipped docs/ directory (already exists)")
	}

	// Create example README only if it doesn't exist
	if _, err := os.Stat("README.md"); os.IsNotExist(err) {
		readmeContent := `# Welcome to Jot

This is an example documentation project using Jot.

## Getting Started

1. Edit the ` + "`jot.yml`" + ` configuration file
2. Add your markdown files to the ` + "`docs`" + ` directory
3. Run ` + "`jot build`" + ` to generate your documentation
4. Run ` + "`jot serve`" + ` to preview locally

## Features

-  Automatic table of contents
-  Cross-references
-  Syntax highlighting
-  Search functionality

Happy documenting!`

		if err := os.WriteFile("README.md", []byte(readmeContent), 0644); err != nil {
			return fmt.Errorf("failed to create README.md: %w", err)
		}
		fmt.Println("  Created example README.md")
	} else {
		fmt.Println("  Skipped README.md (already exists)")
	}

	fmt.Println(" Initialized new Jot project successfully!")
	fmt.Println("  3. Run 'jot serve' to preview locally")

	return nil
}

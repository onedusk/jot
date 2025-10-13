# Jot

[![CI](https://github.com/onedusk/jot/actions/workflows/ci.yml/badge.svg)](https://github.com/onedusk/jot/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/onedusk/jot)](https://goreportcard.com/report/github.com/onedusk/jot)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://github.com/onedusk/jot/blob/main/LICENSE)

Jot is a documentation generator that converts markdown files into modern, searchable documentation websites. Built as a replacement for JetBrains deprecated Writerside IDE.

## Features

- **Automatic TOC Generation** - Hierarchical table of contents from your file structure
- **Multiple Export Formats** - HTML, JSON, YAML, and LLM-optimized outputs
- **Large Scale** - Proven on thousands of documents
- **Zero-Copy Markdown** - Symlink support for direct markdown access

## Installation

### macOS

**Binary Installation (Intel):**
```bash
curl -L https://github.com/onedusk/jot/releases/latest/download/jot-darwin-amd64 -o jot
chmod +x jot
sudo mv jot /usr/local/bin/
```

**Binary Installation (Apple Silicon):**
```bash
curl -L https://github.com/onedusk/jot/releases/latest/download/jot-darwin-arm64 -o jot
chmod +x jot
sudo mv jot /usr/local/bin/
```

### Linux

**Binary Installation (amd64):**
```bash
curl -L https://github.com/onedusk/jot/releases/latest/download/jot-linux-amd64 -o jot
chmod +x jot
sudo mv jot /usr/local/bin/
```

**Binary Installation (arm64):**
```bash
curl -L https://github.com/onedusk/jot/releases/latest/download/jot-linux-arm64 -o jot
chmod +x jot
sudo mv jot /usr/local/bin/
```

### Windows

**Binary Installation (PowerShell):**
```powershell
Invoke-WebRequest -Uri https://github.com/onedusk/jot/releases/latest/download/jot-windows-amd64.exe -OutFile jot.exe
# Add jot.exe to your PATH or move to a directory in PATH
```

**Build from Source:**
```bash
# Clone repository
git clone https://github.com/onedusk/jot
cd jot

# Install dependencies
go mod download

# Build binary
go build -o jot ./cmd/jot

# Run tests
go test ./...
```

### Initialize a Project

```bash
# Create a new documentation project
jot init

# This creates:
# - jot.yaml (configuration)
# - .jotignore (ignore patterns)
# - docs/ (documentation directory)
# - README.md (example file)
```

### Build Documentation

```bash
# Scan and build documentation
jot build

# Output is generated in ./dist/

# Use custom configuration file
jot build --config my-config.yaml

# Enable verbose output for detailed logging
jot build --verbose

# By default, jot looks for jot.yaml in the current directory
```

### Export Documentation

```bash
# Export as JSON
jot export --format json --output docs.json

# Export as YAML
jot export --format yaml --output docs.yaml

# Export for LLMs (optimized format with chunks)
jot export --format llm --output docs-llm.json
```

## Configuration

Edit `jot.yaml` to customize your documentation:

```yaml
version: 1.0  # Configuration version (required)

project:
  name: "My Documentation"        # Project name (required)
  description: "Project documentation"  # Brief description (optional)
  author: "Your Name"             # Author name (optional)

input:
  paths:
    - "docs"        # Source paths to scan (required)
    - "README.md"   # Supports files and directories
  ignore:
    - "**/_*.md"    # Glob patterns to ignore (optional)
    - "**/drafts/**"
    - "**/node_modules/**"

output:
  path: "dist"     # Output directory (default: "dist")
  format: "html"   # Output format: html, json, yaml (default: "html")
  theme: "default" # Theme name (default: "default")

features:
  search: true      # Enable full-text search (default: true)
  llm_export: true  # Enable LLM-optimized export (default: false)
  toc: true         # Generate table of contents (default: true)
```

## Project Structure

```
my-project/
 jot.yaml          # Configuration
       installation.md
 dist/             # Generated output
```

## Markdown Features

Jot supports standard markdown with extensions:

- **Frontmatter** - YAML metadata in documents
- **Code Highlighting** - Syntax highlighting for code blocks
- **Tables** - GitHub-flavored markdown tables
- **Task Lists** - Checkboxes in lists
- **Footnotes** - Reference-style footnotes

## LLM Integration

Jot can export documentation in LLM-optimized format with automatic chunking:

```bash
# Export with chunks for context windows
jot export --format llm --output docs-llm.json

# The LLM format includes:
# - Document chunking (512 token chunks with 128 token overlap)
# - Semantic indexing
# - Section extraction
# - Metadata preservation
```

## Troubleshooting

### Build fails with "config file not found"
- Ensure `jot.yaml` exists in your project root
- Use `--config` flag to specify custom config location
- Run `jot init` to create a default configuration

### No documents found during build
- Check that input paths in `jot.yaml` are correct
- Verify markdown files aren't being ignored by patterns
- Use `--verbose` flag to see which files are being scanned

### Search not working in generated site
- Ensure `features.search: true` in `jot.yaml`
- Check that JavaScript is enabled in browser
- Verify search index was generated in output directory

### Permission denied errors
- On macOS/Linux: Run `chmod +x jot` after downloading binary
- On Windows: Check that antivirus isn't blocking the executable
- Ensure write permissions for output directory

## FAQ

### Can Jot handle large documentation sets?
Yes, Jot is designed to handle thousands of documents efficiently. It uses optimized scanning and rendering algorithms.

### Does Jot support custom themes?
Currently Jot uses a default theme. Custom theme support is planned for future releases.

### Can I use Jot with CI/CD pipelines?
Yes! Jot is a CLI tool that integrates easily with CI/CD. Run `jot build` in your pipeline to generate docs automatically.

### What markdown flavors are supported?
Jot supports GitHub-flavored markdown with extensions for frontmatter, code highlighting, tables, task lists, and footnotes.

### Can I export to formats other than HTML?
Yes, Jot supports JSON, YAML, and LLM-optimized formats via the `jot export` command.

### How does LLM export work?
LLM export creates optimized chunks (512 tokens with 128 token overlap) suitable for feeding into language models with context window limits.

### Is there a watch mode for development?
Yes, use `jot watch` to automatically rebuild when files change (requires the serve command to be running).

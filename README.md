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
version: 1.0
project:
  name: "My Documentation"
  description: "Project documentation"
  author: "Your Name"

input:
  paths:
    - "docs"        # Clean paths without ./
    - "README.md"
  ignore:
    - "**/_*.md"
    - "**/drafts/**"

output:
  path: "dist"
  format: "html"
  theme: "default"

features:
  search: true
  llm_export: true
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

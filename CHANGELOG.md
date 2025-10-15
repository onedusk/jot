# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- New `jot toc` command for generating per-directory table of contents
  - Generates `toc.xml` in each directory containing markdown files
  - Creates master `toc.json` at project root with relative paths to all toc.xml files
  - Supports `--dry-run` flag to preview without writing files
  - Supports `--recursive` flag to include subdirectories in each toc.xml
  - Master toc.json acts as a lock file tracking all generated TOC locations

### Changed
- Configuration file renamed from `jot.yaml` to `jot.yml`
  - Updated `jot init` command to generate `jot.yml`
  - Backward compatibility maintained - both .yaml and .yml extensions are supported
  - Updated all documentation references

### Fixed
- Removed unused imports in cmd/jot/toc.go

## [0.1.0] - 2025-10-15

### Added
- Initial release
- Markdown to HTML compilation
- XML Table of Contents generation
- Multi-format export (JSON, YAML, LLM)
- Local development server with HTTP support
- Symlink support for markdown access
- Full-text search capabilities
- Mermaid diagram support
- Syntax highlighting with Prism.js
- CLI commands: `init`, `build`, `serve`, `watch`, `export`
- Configuration file support (jot.yaml)
- Project initialization with templates
- Automatic TOC generation from file structure
- Cross-reference linking
- Responsive HTML output

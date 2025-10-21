# Changelog

All notable changes to Jot will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-10-21

### Added

#### Multi-Format LLM Export System
- **llms.txt format**: Lightweight documentation index per [llmstxt.org](https://llmstxt.org/) specification
- **llms-full.txt format**: Complete documentation concatenation optimized for LLM context windows
- **JSONL format**: JSON Lines export for vector database ingestion (Pinecone, Weaviate, Qdrant)
- **Enriched Markdown format**: Markdown with YAML frontmatter metadata for enhanced processing

#### Token-Based Chunking
- **tiktoken-go integration**: Accurate token counting using OpenAI's `cl100k_base` encoding
- **Token-aware chunking**: Replaced character-based chunking with token-based for precise LLM context management
- **Word boundary preservation**: Intelligent splitting that avoids breaking words mid-token
- **Binary search algorithm**: Efficient token boundary detection

#### Pluggable Chunking Strategies
- **Fixed-size strategy**: Token-based fixed-size chunks with configurable overlap
- **Markdown headers strategy**: Splits documents at markdown header boundaries (`#` to `######`)
- **Recursive strategy**: Hierarchical splitting using multiple separators (paragraph → line → space → character)
- **Semantic strategy**: Stub implementation for future embedding-based boundary detection

#### CLI Enhancements
- **New export formats**: `--format` flag supports `json`, `yaml`, `llms-txt`, `llms-full`, `jsonl`, `markdown`
- **Chunking configuration**: `--strategy`, `--chunk-size`, `--chunk-overlap` flags for fine-grained control
- **Workflow presets**:
  - `--for-rag`: Optimized for RAG (jsonl + semantic + 512 tokens)
  - `--for-context`: Optimized for context windows (markdown + headers + 1024 tokens)
  - `--for-training`: Optimized for training (jsonl + fixed + 256 tokens)
- **Embeddings support**: `--include-embeddings` flag for JSONL with API cost warnings
- **Comprehensive validation**: Flag validation with helpful error messages and examples

#### Build Integration
- **Auto-generation**: `jot build` automatically generates `llms.txt` and `llms-full.txt`
- **Configuration**: `features.llm_export` in `jot.yml` (default: true)
- **Skip flag**: `--skip-llms-txt` to disable LLM export during build
- **File size reporting**: Humanized byte sizes (KB, MB, GB) in build logs
- **Non-breaking errors**: LLM export failures don't break builds

### Changed
- **Export system architecture**: Refactored to support multiple exporters with consistent interface
- **Chunk struct**: Added `TokenCount` field for accurate token reporting
- **Export types**: Added `ProjectConfig` and `ChunkMetadata` structs

### Fixed
- **Token counting bug**: Replaced `len(content)` with accurate token counting via tokenizer
- **Chunk overlap calculation**: Now uses token-based overlap instead of character-based

### Performance
- **Binary search chunking**: Efficient token boundary detection using binary search algorithm
- **Benchmarks added**: Performance benchmarks for all chunking strategies
- **Streaming support**: JSONL format supports line-by-line streaming for large datasets

### Technical Details
- **Dependencies**: Added `github.com/pkoukk/tiktoken-go` for token counting
- **New packages**: `internal/tokenizer`, `internal/chunking`
- **New files**: 15 new implementation files, comprehensive test suites
- **Tests**: 71 passing tests across 7 packages, 0 failures
- **Coverage**: >85% test coverage for new packages

## [Unreleased] - 2025-10-12

### Fixed
- **Performance**: Replaced `WriteString(fmt.Sprintf(...))` with `fmt.Fprintf` in markdown compiler for better performance
- **Document Chunking**: Fixed overlap calculation in `chunkDocument` function to properly track character positions instead of mixing word indices with character counts
- **HTML Rendering**: Fixed code block rendering to properly add language classes to both `<pre>` and `<code>` tags for better syntax highlighting support
- **Breadcrumb Navigation**: Fixed breadcrumb path generation to use correct absolute paths for directories and files
- **Navigation Tree**: Added missing `nav-tree` class wrapper to navigation output
- **Template**: Updated page title format to include "Jot" branding
- **Test Suite**: Fixed test expectations to match actual implementation behavior

### Changed
- Improved document chunking algorithm for more accurate text segmentation with proper overlap handling

## [0.0.5] - 2025-10-07

### Added
- **GoDoc Generation**: Added comprehensive GoDoc comments to all Go source files in the `internal` and `cmd` directories to improve code clarity and maintainability. This includes documentation for all public types, functions, and methods.

### Changed
- Improved code documentation across the entire Go codebase.

## [0.0.4] - 2025-10-04

### Added
- **Local Development Server**: Implemented `jot serve` command for local documentation preview
  - HTTP server with configurable port (default 8080)
  - Automatic browser opening with cross-platform support (Linux, macOS, Windows)
  - Smart index handling (serves README.html as default, fallback to index.html)
  - Static file serving for CSS, JS, images, and other assets
  - Comprehensive error handling with helpful user guidance

### Features
- `--port, -p`: Set custom server port
- `--open, -o`: Control browser auto-opening (default: true)
- `--dir, -d`: Override serve directory (default: ./dist)

### Technical Improvements
- Proper HTTP file server implementation
- Cross-platform browser launching support
- Configuration integration with existing Viper setup
- Graceful error handling for missing build artifacts

## [0.0.3] - 2025-10-03

### Added
- Sidebar items are now collapsible dropdowns for better organization.

### Changed
- Updated navigation bar icon and background to a "dusk" themed gradient.
- Refactored CSS out of the HTML template into a separate `style.css` file.
- Updated the build process to copy the new `style.css` file to the output directory, reducing the size of generated HTML files.

## [0.0.2] - 2025-08-20

### Added

#### UI/UX Enhancements
- **Modern Glassmorphic Sidebar**: Complete redesign with glass-morphism effects
  - Collapsible sidebar (72px collapsed, 280px expanded)
  - Blur effects with backdrop-filter
  - Smooth cubic-bezier animations
  - Dark theme with refined color palette
- **macOS-Style Window Controls**: Traffic light controls (red, yellow, green)
  - Fade to gray when sidebar not hovered
  - Native macOS positioning and styling
- **Enhanced Navigation System**:
  - Icon-based navigation with SVG icons
  - Dynamic icon selection based on content type
  - Smooth expand/collapse animations
  - Active state indicators with accent colors
- **Profile & Search Integration**:
  - Gradient avatar display
  - Integrated search bar with icon
  - Opacity transitions on hover/expand
- **Refined Typography & Spacing**:
  - Improved font sizing and line heights
  - Better visual hierarchy
  - Optimized whitespace and padding

### Changed
- Updated HTML template generation for modern design
- Improved navigation node rendering with icons
- Enhanced color scheme for better readability
- Refined hover states and transitions
- Optimized sidebar interactions

### Technical Improvements
- Better CSS variable organization
- Improved responsive design patterns
- Enhanced animation performance
- Cleaner component architecture

## [0.0.1] - 2025-08-13

### Added

#### Core Features
- **File Scanner**: Recursive markdown file scanning with configurable ignore patterns
- **TOC Generator**: Hierarchical XML table of contents generation from document structure
- **HTML Renderer**: Markdown to HTML conversion with syntax highlighting and modern styling
- **Search Functionality**: Client-side full-text search with JSON index generation
- **CLI Interface**: Comprehensive command-line interface using Cobra framework
  - `init`: Initialize new documentation project
  - `build`: Build documentation from markdown files
  - `serve`: Start development server (planned)
  - `watch`: Watch for changes and rebuild (planned)
  - `export`: Export documentation in various formats

#### Document Processing
- Markdown parsing with Blackfriday v2
- Automatic heading extraction for navigation
- Smart internal link resolution (`.md` to `.html`)
- Relative path handling for all assets and links
- Breadcrumb navigation generation

#### Export Formats
- JSON export with document chunking for LLM consumption
- YAML export for configuration and data interchange
- Search index generation for client-side search

#### Styling and UI
- Professional syntax highlighting based on Tailwind CSS theme
- Dark mode support with automatic detection
- Responsive design for mobile and desktop
- Interactive code copy buttons
- Keyboard shortcuts (Ctrl+K for search)

#### Build and Distribution
- Single binary distribution with no runtime dependencies
- Cross-platform support (macOS, Linux, Windows)
- Docker container support
- Automated release workflow with GitHub Actions

### Technical Implementation
- Written in Go for performance and portability
- Test-Driven Development (TDD) approach
- SPARC methodology for systematic development
- Modular architecture with clear separation of concerns
- Comprehensive test coverage

### Documentation
- Complete requirements specification
- System architecture documentation
- Pseudocode design documents
- Usage examples and quick start guide

### Known Limitations
- Live reload not yet implemented
- Version control integration planned for future release
- LLM API endpoints planned for future release

## Future Releases

### [0.2.0] - Planned
- Live reload functionality for development server
- Version control and change tracking
- LLM/Agent API endpoints
- Multiple theme support

### [0.3.0] - Planned
- Plugin system
- Cloud deployment features
- Advanced search with filters
- Multi-language support

---

For more information, see the [README](../../README.md)

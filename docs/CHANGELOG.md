# Changelog

All notable changes to Jot will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased] - 2026-09-18

### Added

- **Page export menu**: Split button with dropdown on every rendered page for copying and exporting content
  - "Copy page" copies rendered page text to clipboard
  - "Copy page as Markdown" copies raw markdown source to clipboard (base64-encoded and embedded in page)
  - "Open Markdown" opens the source `.md` file in a new browser tab
  - Dropdown closes on outside click, "Copied!" feedback with emerald accent state
  - Mobile responsive: collapses to icon-only at 768px breakpoint
- **Markdown source files in output**: Source `.md` files are now copied alongside `.html` files in the build output directory
- **Continuous integration**: GitHub Actions workflow (`.github/workflows/ci.yml`) checks formatting, module tidiness, `go vet`, build, and `go test -race` on every push and pull request
- **End-to-end build tests**: `cmd/jot/e2e_test.go` runs full builds from a temporary directory outside the repository, the way an installed binary is used

### Changed

- **Public package restructuring**: Moved `scanner`, `tokenizer`, `export`, and `chunking` from `internal/` to `pkg/` to enable external module imports (e.g. from the planned `mlpipe` package)
- **Decoupled viper from export package**: `ToLLMFormat()` now accepts `chunkSize, overlap int` parameters instead of reading from `viper.GetInt()`, removing the CLI framework dependency from the library layer
- Updated 28 import paths across all consumer files
- **Formatting**: Applied `gofmt` to all Go sources (no behavior change)
- **Dependencies**: `go mod tidy` now lists `tiktoken-go` as a direct dependency
- **Config loading**: `export`, `toc`, and `debug` load configuration without `build`'s flag overrides. Previously `export --output docs.jsonl` was also read as the site output directory

### Fixed

- **Build version stamping**: `make build` and `scripts/release.sh` read `VERSION` from the repository root, where it does not exist, so binaries were built with an empty version string. Both now read `docs/VERSION`
- **Unstyled sites from installed binaries**: CSS and JavaScript assets were read from `web/templates/assets/` relative to the current directory, so `jot build` run anywhere other than the jot source tree produced pages with no styles, search, or syntax highlighting, and silently skipped the missing files. Assets are now embedded in the binary with `go:embed`, and a failure to write one fails the build
- **Build summary**: `jot build` no longer reports creating `assets/styles.css`, a file that was never written
- **Rebuilds reading their own output**: The markdown copies written to the output directory were scanned as source on the next build whenever the output directory sat inside an input path (the default: input `.`, output `./dist`), so each rebuild added another copy of every page and nested `dist/dist/...`. The output directory is now excluded from scanning, matched by file identity so symlinked or differently spelled paths to it are covered. `build` excludes its actual output directory, including a `-o` override; `export`, `toc`, and `debug` exclude the configured `output.path`. New `Scanner.Exclude` method in `pkg/scanner`
- **`clean` deleting source files**: With `output.clean: true` or `--clean`, an output directory that was or contained an input directory (for example input `docs`, output `docs`) was removed with `os.RemoveAll`, deleting the sources. `build` now refuses to clean in that case and explains why
- **In-place builds overwriting sources**: When the output directory is an input directory (`jot build -o .`), the markdown copy of each page was written over its own source file with the frontmatter stripped. Sources are now left untouched in that case
- **Pages silently overwritten across input roots**: Document paths are relative to each input root, so with the input paths `jot init` generates (`docs` and `README.md`), `docs/README.md` and `README.md` both wrote `README.html` and one page was lost while the build reported both as generated. `build` and `export` now fail with an error naming both source files. Output paths are compared case-insensitively, since `README.md` and `readme.md` are the same file on the default macOS and Windows filesystems. The check runs before `clean` empties the output directory, so a failed build leaves the previous site in place. A file reached through two overlapping input paths (for example `.` and `README.md`) is scanned once rather than reported as a collision
- **No `index.html` for README-rooted sites**: When the root document was `README.md`, no index page was generated and the page was only written as `README.html`, so static hosts (GitHub Pages, S3, Netlify) returned 404 at `/` and the header logo link was broken. The root README is now also written as `index.html`
- **Piped exports**: `jot export` wrote progress messages to stdout along with the exported data, so `jot export --format json | jq .` failed. Progress now goes to stderr, and stdout output ends with exactly one newline (previously JSONL gained a trailing blank line)
- **Duplicated error output**: Errors were printed twice (by Cobra and by `main`), and runtime failures such as "no markdown files found" were followed by the full usage text. Errors now print once, without usage
- **Pages missing from navigation**: A document and a directory with the same title (for example `guides.md` next to `guides/`), or directories whose names differ only in separators (`my-dir/` and `my_dir/`), were merged into one table of contents node. Documents under the merged directory disappeared from the sidebar. Directories are now matched by name, and only against other directories
- **Timestamps labeled UTC but in local time**: `modified` values in `toc.xml` and the search index appended a literal `Z` to the local wall-clock time. They are now converted to UTC first
- **Unescaped titles in navigation**: Document and directory titles and link paths were inserted into the sidebar HTML verbatim, so a title such as `Using <T> generics` or `Q&A` broke the markup on every page. They are now HTML-escaped. The `path` attribute in `toc.xml` is now XML-escaped as well

## [0.2.0] - 2026-02-27

### Added

#### UI Design Parity
- **Full-bleed layout**: Replaced floating rounded panels with flush sidebar and content areas matching Protocol reference design
- **Always-visible navigation**: Sidebar nav sections render expanded by default instead of collapsed accordions
- **Header search bar**: Centered search input in header with keyboard shortcut badge (Cmd+K)
- **Configurable branding**: Logo text, header nav links driven from `project.name` in `jot.yml` via new `SiteConfig` struct
- **Sign In button**: Green accent button in header nav
- **Previous/Next navigation**: Bottom-of-page links to adjacent documents, computed from TOC tree
- **Footer**: Copyright line and project branding at page bottom
- **Feedback widget**: "Was this page helpful? Yes / No" interactive buttons
- **Callout boxes**: Auto-detection of `> **Note:**` and `> **Warning:**` blockquotes into styled callout components
- **Inline dropdown search**: Client-side full-text search with results dropdown beneath header search bar
  - Keyboard navigation (arrow keys, Enter, Escape)
  - Query highlighting with `<mark>` tags
  - Scored results (title +10, headings +5, keywords +3, content +1)
  - Search index embedded as JS for file:// protocol compatibility

### Changed
- **Layout**: Removed animated gradient orbs background, replaced with subtle top-glow
- **Body scroll**: Removed viewport-locking (`position: fixed`) for natural page scrolling
- **Active nav style**: Changed from green accent to white text with subtle background highlight
- **Content width**: Increased from 900px to 1100px
- **Sidebar width**: Increased from 240px to 260px
- **Header height**: Increased from 56px to 60px
- **Search index output**: Now generates both `search-index.json` and `search-index.js` (global variable) for universal compatibility
- **Config plumbing**: `SiteConfig` flows from `build.go` → `compiler.go` → `renderer.go` with `ProjectName` and `NavLinks`
- **Release script**: Replaced Ruby gem build/push logic with `make build` / `make release`

### Technical Details
- **New types**: `SiteConfig`, `NavLink`, `PageLink` in renderer package
- **New functions**: `flattenTOC()`, `computeAdjacentPages()`, `processCallouts()` in renderer
- **Modified files**: `renderer.go`, `template.go`, `style.css`, `search.js`, `compiler.go`, `build.go`, `indexer.go`, `g.sh`

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

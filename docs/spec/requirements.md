# Requirements Spec.

## 1. Executive Summary

Jot is a modern documentation generator designed to replace JetBrains' deprecated Writerside IDE. It provides a fast, simple, and platform-independent solution for aggregating markdown files into deployable web documentation with advanced features for LLM/AI integration.

## 2. Functional Requirements

### 2.1 File Scanning System
- **FR-001**: Recursively scan directories for markdown files (*.md)
- **FR-002**: Support configurable ignore patterns (.jotignore file)
- **FR-003**: Preserve original directory structure and file metadata
- **FR-004**: Handle symlinks and nested directories gracefully
- **FR-005**: Support glob patterns for include/exclude rules

### 2.2 Table of Contents Generation
- **FR-006**: Generate hierarchical XML TOC from file structure
- **FR-007**: Extract titles from H1 tags or frontmatter
- **FR-008**: Support custom ordering (weight-based or alphabetical)
- **FR-009**: Handle multi-level nesting (unlimited depth)
- **FR-010**: Auto-generate unique IDs for each entry

### 2.3 HTML Compilation
- **FR-011**: Convert markdown to semantic HTML5
- **FR-012**: Automatically resolve cross-references between documents
- **FR-013**: Generate navigation elements (breadcrumbs, sidebar, TOC)
- **FR-014**: Support syntax highlighting for code blocks
- **FR-015**: Bundle CSS/JS assets with optimization

### 2.4 LLM/Agent Integration
- **FR-016**: Export documentation as structured JSON
- **FR-017**: Generate YAML format for configuration tools
- **FR-018**: Create vector embeddings for semantic search
- **FR-019**: Provide RESTful API for programmatic access
- **FR-020**: Support chunk-based content splitting

### 2.5 Search Functionality
- **FR-021**: Full-text search across all documentation
- **FR-022**: Fuzzy matching for typo tolerance
- **FR-023**: Real-time search suggestions
- **FR-024**: Search result highlighting
- **FR-025**: Pre-built search indices for performance

### 2.6 Version Control
- **FR-026**: Track file changes with simple history
- **FR-027**: Generate diffs between versions
- **FR-028**: Support webhook notifications for changes
- **FR-029**: Integrate with git for change detection
- **FR-030**: Maintain change logs

### 2.7 CLI Interface
- **FR-031**: Command-line interface using Cobra framework
- **FR-032**: Support init, build, serve, watch commands
- **FR-033**: Configurable via flags and config files
- **FR-034**: Provide helpful error messages and usage guides
- **FR-035**: Support batch operations

### 2.8 Web Server
- **FR-036**: Built-in development server with Gin
- **FR-037**: Hot reload on file changes
- **FR-038**: CORS support for API access
- **FR-039**: Static file serving with caching
- **FR-040**: WebSocket support for live updates

## 3. Non-Functional Requirements

### 3.1 Performance
- **NFR-001**: Build time < 1 second for 1000 files
- **NFR-002**: Search latency < 50ms
- **NFR-003**: Memory usage < 100MB for typical projects
- **NFR-004**: Startup time < 500ms
- **NFR-005**: Support projects with 10,000+ files

### 3.2 Usability
- **NFR-006**: Single binary distribution (no dependencies)
- **NFR-007**: Cross-platform support (Windows, macOS, Linux)
- **NFR-008**: Intuitive CLI with helpful documentation
- **NFR-009**: Zero-config operation with sensible defaults
- **NFR-010**: Clear error messages with resolution hints

### 3.3 Security
- **NFR-011**: Input sanitization to prevent XSS
- **NFR-012**: Path traversal protection
- **NFR-013**: Rate limiting on API endpoints
- **NFR-014**: Secure defaults for web server
- **NFR-015**: No execution of user content

### 3.4 Compatibility
- **NFR-016**: CommonMark compliant markdown parsing
- **NFR-017**: Standard HTML5 output
- **NFR-018**: JSON Schema compliant exports
- **NFR-019**: OpenAPI specification for APIs
- **NFR-020**: UTF-8 encoding throughout

## 4. User Stories

### 4.1 Developer Stories
- As a developer, I want to generate documentation from my markdown files with a single command
- As a developer, I want to preview my documentation locally with hot reload
- As a developer, I want to customize the look and feel with themes
- As a developer, I want to integrate documentation generation into my CI/CD pipeline

### 4.2 AI/LLM Stories
- As an AI system, I want to access documentation via structured JSON API
- As an LLM, I want to search documentation semantically using embeddings
- As an agent, I want to retrieve specific sections without parsing HTML
- As a chatbot, I want to get contextual information about code examples

### 4.3 End User Stories
- As a user, I want to search documentation quickly and accurately
- As a user, I want to navigate between related topics easily
- As a user, I want to view documentation offline
- As a user, I want to access documentation on any device

## 5. Acceptance Criteria

### 5.1 File Scanning
-  Discovers all markdown files in specified directories
-  JSON output validates against schema
- Must use only standard library where possible
- Must compile to a single binary
- Must work offline (no external dependencies)
- Must be open source friendly

## 7. Dependencies

- Go 1.21+ (build time only)
- Gin web framework
- Cobra CLI framework
- Blackfriday markdown parser
- No runtime dependencies

## 8. Success Metrics

- Adoption by 100+ projects in first 3 months
- 95%+ user satisfaction in surveys
- <5 critical bugs in first release
- Performance targets met in 95% of use cases
- Active community contributions

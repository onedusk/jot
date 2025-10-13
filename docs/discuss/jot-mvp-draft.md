# Jot - Documentation Tool MVP Design Document

## Executive Summary

Jot is a modern documentation generation tool designed to replace JetBrains' deprecated Writerside IDE. Built with Go and Gin framework, it provides a simple, fast, and platform-independent solution for aggregating markdown files into deployable web documentation.

## Core Objectives

1. **Aggregate** - Collect all `*.md` files project-wide while preserving directory structure
2. **Generate** - Create XML table of contents automatically
3. **Compile** - Build web application/archive with interconnected HTML files
4. **Deploy** - Produce deployment-ready documentation websites

## Technical Architecture

### Technology Stack
- **Language**: Go (Golang)
- **Web Framework**: Gin
- **CLI Framework**: Cobra & Viper
- **Template Engine**: Go html/template
- **Version Control**: Custom lightweight system
- **Build Output**: Static HTML + optional embedded web server

### System Architecture

```
          
                                                        
    Manager              (TOC Builder)          Generator    

## MVP Features

### Phase 1: Core Functionality

#### 1. File Aggregation System
- **Recursive markdown scanner**: Traverse project directories
- **Metadata preservation**: Maintain file paths, timestamps, and frontmatter
- **Configurable ignore patterns**: `.jotignore` file support
- **Smart path handling**: Relative/absolute path resolution

```go
type DocumentFile struct {
    Path         string
    RelativePath string
    Content      []byte
    Metadata     map[string]interface{}
    ModTime      time.Time
}
```

#### 2. XML Table of Contents Generation
- **Hierarchical structure**: Mirror directory organization
- **Automatic titling**: Extract from H1 tags or frontmatter
- **Customizable ordering**: Weight-based or alphabetical
- **Multi-level nesting**: Support deep directory structures

```xml
<toc version="1.0">
  <chapter id="getting-started" path="docs/getting-started.md">
    <title>Getting Started</title>
    <section id="installation" path="docs/getting-started/installation.md">
      <title>Installation</title>
    </section>
  </chapter>
</toc>
```

#### 3. Web Compilation Engine
- **Template system**: Customizable HTML templates
- **Cross-linking**: Automatic internal link resolution
- **Asset management**: CSS/JS bundling and optimization
- **Navigation generation**: Breadcrumbs, sidebars, TOC

#### 4. Version Control Integration
- **Change detection**: File modification tracking
- **History tracking**: Simple commit-like system
- **Diff generation**: Show documentation changes
- **Event broadcasting**: Webhook support for CI/CD

### Phase 2: Enhanced Features

#### 5. LLM/Agent-Friendly Design
- **Structured data export**: JSON/YAML representations
- **Semantic markup**: Schema.org microdata
- **API endpoints**: RESTful access to documentation
- **Embeddings support**: Vector database compatibility

```go
type LLMExport struct {
    Documents []struct {
        ID         string
        Title      string
        Content    string
        Embeddings []float64
        Tags       []string
    }
}
```

#### 6. Search Capabilities
- **Full-text search**: Built-in search engine
- **Fuzzy matching**: Typo-tolerant searches
- **Search indexing**: Pre-built indices for performance
- **SEO optimization**: Sitemap.xml, meta tags, structured data

#### 7. CLI Interface
```bash
# Initialize project
jot init

# Build documentation
jot build

# Serve locally
jot serve --port 8080

# Watch for changes
jot watch

# Generate search index
jot index

# Export for LLM
jot export --format json
```

## Implementation Plan

### Milestone 1: Foundation (Week 1-2)
- [ ] Project setup with Go modules
- [ ] Basic CLI structure with Cobra
- [ ] File scanning and aggregation
- [ ] Simple XML TOC generation

### Milestone 2: Web Generation (Week 3-4)
- [ ] HTML template system
- [ ] Markdown to HTML conversion
- [ ] Static file generation
- [ ] Basic web server with Gin

### Milestone 3: Enhanced Features (Week 5-6)
- [ ] Version control system
- [ ] Search implementation
- [ ] LLM export formats
- [ ] Configuration management

### Milestone 4: Polish & Testing (Week 7-8)
- [ ] Comprehensive testing suite
- [ ] Documentation
- [ ] Performance optimization
- [ ] Release preparation

## Configuration Schema

```yaml
# jot.yaml
version: 1.0
project:
  name: "My Documentation"
  description: "Project documentation"
  author: "Your Name"

input:
  paths:
    - "./docs"
    - "./README.md"
  ignore:
    - "**/_*.md"
    - "**/drafts/**"

output:
  path: "./dist"
  format: "html"
  theme: "default"

features:
  search: true
  versioning: true
  llm_export: true
  
server:
  port: 8080
  auto_reload: true
```

## API Design

### REST Endpoints
```
GET  /api/docs              - List all documents
GET  /api/docs/:id          - Get specific document
GET  /api/search?q=query    - Search documents
GET  /api/toc               - Get table of contents
GET  /api/export/:format    - Export in various formats
```

### Event System
```go
type DocEvent struct {
    Type      string    // created, updated, deleted
    Path      string
    Timestamp time.Time
    Changes   []Change
}
```

## Security Considerations

1. **Input sanitization**: Prevent XSS in markdown
2. **Path traversal protection**: Validate file paths
3. **Rate limiting**: Protect API endpoints
4. **CORS configuration**: Secure cross-origin requests

## Performance Targets

- **Build time**: < 1 second for 1000 files
- **Search latency**: < 50ms for queries
- **Memory usage**: < 100MB for typical projects
- **Startup time**: < 500ms

## Success Metrics

1. **Adoption**: Replace Writerside for documentation needs
2. **Performance**: 10x faster than alternative tools
3. **Simplicity**: Single binary, zero dependencies
4. **Compatibility**: Works on all major platforms
5. **Integration**: Seamless CI/CD pipeline support

## Future Enhancements

- Plugin system for custom processors
- Multi-language support (i18n)
- Collaborative editing features
- Cloud deployment options
- Analytics and usage tracking
- PDF/EPUB export
- Theme marketplace

## Conclusion

Jot aims to be the simplest, fastest, and most developer-friendly documentation tool available. By focusing on core functionality and platform independence, it provides a reliable alternative to complex documentation systems while maintaining extensibility for future needs.
# System Architecture

## 1. High-Level Architecture

```


                     Core Engine


     Engine        Control          Export


## 2. Component Architecture

### 2.1 CLI Layer (`cmd/jot/`)

```go
cmd/jot/
 main.go              // Entry point
 serve.go             // Server command

**Responsibilities:**
- Parse command-line arguments
- Load configuration
- Orchestrate core operations
- Handle errors and display output

### 2.2 Core Engine (`internal/`)

```go
internal/
 scanner/
 toc/
 renderer/
    assets.go        // Asset management
     optimizer.go     // Output optimization
### 2.3 Service Layer (`internal/`)

```go
internal/
 search/
 vcs/
 api/
 export/
```

### 2.4 Package Layer (`pkg/`)

```go
pkg/
 config/
 logger/
     paths.go         // Path utilities

## 3. Data Models

### 3.1 Document Model

```go
type Document struct {
    ID           string                 // Unique identifier
    Path         string                 // Absolute file path
    RelativePath string                 // Relative to project root
    Title        string                 // Extracted title
    Content      []byte                 // Raw markdown content
    HTML         string                 // Rendered HTML
    Metadata     map[string]interface{} // Frontmatter data
    ModTime      time.Time              // Last modification
    Sections     []Section              // Document sections
    Links        []Link                 // Internal/external links
    CodeBlocks   []CodeBlock            // Code snippets
}

type Section struct {
    ID        string
    Title     string
    Level     int
    Content   string
    StartLine int
    EndLine   int
}
```

### 3.2 TOC Model

```go
type TOCNode struct {
    ID       string
    Title    string
    Path     string      // nil for directories
    Weight   int         // For custom ordering
    Children []*TOCNode
}

type TableOfContents struct {
    Version string
    Root    *TOCNode
    Index   map[string]*TOCNode // Fast lookup by ID
}
```

### 3.3 Search Model

```go
type SearchIndex struct {
    Version   string
    Documents []SearchDocument
    Terms     map[string][]TermOccurrence
    Fuzzy     *FuzzyIndex
}

type SearchResult struct {
    Document  SearchDocument
    Score     float64
    Snippets  []string
    Highlight map[string][]int // Term positions
}
```

## 4. API Design

### 4.1 RESTful Endpoints

```
GET    /api/v1/docs                 # List all documents
GET    /api/v1/docs/:id             # Get specific document
GET    /api/v1/search?q=query       # Search documents
GET    /api/v1/toc                  # Get table of contents
GET    /api/v1/export/:format       # Export documentation
POST   /api/v1/embeddings           # Generate embeddings
GET    /api/v1/changes              # Get recent changes
WS     /api/v1/ws                   # WebSocket for live updates
```

### 4.2 Response Formats

```json
// Document Response
{
  "id": "getting-started-installation",
  "path": "docs/getting-started/installation.md",
  "title": "Installation Guide",
  "content": "# Installation Guide\n...",
  "html": "<h1>Installation Guide</h1>...",
  "metadata": {
    "author": "John Doe",
    "tags": ["tutorial", "setup"]
  },
  "sections": [...],
  "links": {
    "internal": ["./quickstart.md"],
    "external": ["https://golang.org"]
  }
}

// Search Response
{
  "query": "install",
  "results": [
    {
      "document": {...},
      "score": 0.95,
      "snippets": [
        "...to <mark>install</mark> Jot, run the following..."
      ]
    }
  ],
  "total": 15,
  "took": 12
}
```

## 5. File Structure

### 5.1 Project Layout

```
jot-project/
 .jotignore           # Ignore patterns
    getting-started/
        reference.md
       installation.html
       js/
     temp/
### 5.2 Template Structure

```
web/templates/
 layouts/
    header.html      # Page header
    footer.html      # Page footer
        themes/      # Theme variations
         highlight.js # Syntax highlighting
## 6. Deployment Architecture

### 6.1 Binary Distribution

```
releases/
 jot-v0.1.0-darwin-amd64.tar.gz
 checksums.txt
### 6.2 Docker Support

```dockerfile
# Multi-stage build
FROM golang:1.21 AS builder
WORKDIR /app
COPY . .
RUN go build -o jot cmd/jot/main.go

FROM alpine:latest
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/jot /usr/local/bin/
ENTRYPOINT ["jot"]
```

## 7. Performance Considerations

### 7.1 Concurrency Model

- **Parallel file scanning**: Worker pool for I/O operations
- **Concurrent rendering**: Process multiple documents simultaneously
- **Async search indexing**: Background index updates
- **Buffered WebSocket**: Efficient real-time updates

### 7.2 Caching Strategy

- **Template cache**: Pre-compiled templates
- **Markdown cache**: Parsed AST storage
- **Search cache**: Query result caching
- **Asset cache**: Compressed static files

### 7.3 Memory Management

- **Streaming**: Large file handling
- **Pooling**: Reusable buffers
- **Lazy loading**: On-demand processing
- **GC tuning**: Optimized for throughput

## 8. Security Architecture

### 8.1 Input Validation

- **Path sanitization**: Prevent directory traversal
- **Markdown sanitization**: XSS prevention
- **Config validation**: Type and range checks
- **API validation**: Request parameter validation

### 8.2 Access Control

- **Read-only by default**: No file modifications
- **Sandboxed execution**: Limited system access
- **Rate limiting**: API request throttling
- **CORS configuration**: Controlled cross-origin access

## 9. Extensibility

### 9.1 Plugin Interface

```go
type Plugin interface {
    Name() string
    Version() string
    Process(doc *Document) error
    RegisterAPI(router *gin.Engine)
}
```

### 9.2 Theme System

```yaml
theme:
  name: "custom-theme"
  extends: "default"
  variables:
    primary-color: "#007bff"
    font-family: "Inter, sans-serif"
  templates:
    - "custom-header.html"
```

## 10. Integration Points

### 10.1 CI/CD Integration

```yaml
# GitHub Actions Example
- name: Build Documentation
  uses: onedusk/jot-action@v1
  with:
    config: jot.yml
    output: ./dist

# GitLab CI Example
build-docs:
  image: onedusk/jot:latest
  script:
    - jot build
  artifacts:
    paths:
      - dist/
```

### 10.2 Editor Integration

- VS Code extension for live preview
- IntelliJ plugin for documentation generation
- Vim plugin for markdown validation

This architecture provides a scalable, maintainable foundation for the Jot documentation generator with clear separation of concerns and extensibility points.

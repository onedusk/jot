# Jot v0.1.0 Release

##  Release Overview

##  New Features in v0.1.0
### 2. **Mermaid Diagram Support**
- Automatic detection and rendering of Mermaid diagrams
- Supports flowcharts, sequence diagrams, gantt charts, and more
- Themed integration matching documentation style
- Code blocks with `language-mermaid` are automatically converted

### 3. **Enhanced Syntax Highlighting**
- Prism.js integration for professional code highlighting
- Automatic language detection for Go, JavaScript, Python, Bash, SQL, etc.
- Line numbers for better code reference
- Dark theme optimized for readability
- Copy button on all code blocks

##  Core Features (from MVP)
- Document metadata extraction
- Frontmatter support
- Section, link, and code block parsing

### TOC Generator
- Hierarchical XML table of contents
- Preserves directory structure
- Automatic title extraction
- Unique ID generation
- Fast node lookup with indexing

### HTML Renderer
- Markdown to HTML conversion using Blackfriday
- Template-based page generation
- Breadcrumb navigation
- Internal link resolution (.md  .html)
- Responsive design

### CLI Interface
- Commands: `init`, `build`, `serve`, `watch`
- Configuration file support (jot.yml)
- Project initialization with templates
- Progress feedback and error handling

##  Quick Start
# Initialize a new documentation project
./jot init

# Build documentation
./jot build

# Serve documentation locally
./jot serve

# Watch for changes and rebuild
./jot watch
```

##  Example Configuration
project:
  name: 'My Documentation'
  description: 'Project documentation'

input:
  paths:
    - './docs'
  ignore:
    - '**/_*.md'
    - '**/drafts/**'

output:
  path: './dist'
  format: 'html'

features:
  search: true
  syntax_highlighting: true
  auto_toc: true
```

##  Performance Metrics
- **Language Support**: 10+ programming languages
- **Diagram Types**: All Mermaid diagram types

##  Technical Implementation
func (r *HTMLRenderer) renderNavNode(...) {
    hasActivePage := r.containsActivePage(node, currentPath)
    // Render with expandable sections
}
```

### Mermaid Integration (renderer.go:524-539)
```javascript
// Automatic Mermaid diagram conversion
document.querySelectorAll('pre code.language-mermaid').forEach(function(block) {
    const mermaidDiv = document.createElement('div');
    mermaidDiv.className = 'mermaid';
    // Convert and render
});
```

### Syntax Highlighting (renderer.go:505-523)
```javascript
// Language auto-detection and Prism.js setup
function detectLanguage(block) {
    // Heuristics for common languages
    if (text.includes('func ') && text.includes('package ')) return 'go';
    // ... more detection logic
}
```

##  Example: Mermaid Diagram in Markdown
    A[Start] --> B{Is it working?}
    B -->|Yes| C[Great!]
    B -->|No| D[Debug]
    D --> A
```
````

This will automatically render as an interactive diagram in the documentation.

##  Timeline Achievement
1.  Dropdown Directories: 45 minutes
4. Deploy the `dist` folder

##  Known Limitations

##  Next Steps (v0.2.0)
- [ ] Live reload in watch mode
- [ ] Multiple export formats (PDF, EPUB)
- [ ] Theme customization
- [ ] Plugin system
- [ ] Cloud deployment integrations

##  License
- Go community for excellent tooling

---

**Jot v0.1.0** - Modern Documentation, Simply Done
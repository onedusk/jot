# Roadmap

## Current Version: 1.0

### ✅ Completed Features
- [x] Markdown to HTML compilation
- [x] XML Table of Contents generation
- [x] Multi-format export (JSON, YAML, LLM)
- [x] **Local development server** - HTTP server with browser auto-open
- [x] Symlink support for markdown access
- [x] Global installation as `jotdoc`
- [x] Clean path format (removed `./` prefixes)

## Version 1.1 (Quick Improvements)
*Timeline: 1-2 weeks*

- [ ] **Fix duplicate TOC entries** - Improve path deduplication logic
- [ ] **Better title extraction** - Support frontmatter, multiple heading formats
- [ ] **Markdown compiler** - Native markdown output with navigation
- [ ] **Config validation** - Verify jot.yaml on load
- [ ] **Better error messages** - User-friendly error reporting

## Version 1.2 (Quality of Life)
*Timeline: 2-3 weeks*

- [ ] **Live reload** - Auto-refresh browser on file changes
- [ ] **Watch mode** - Auto-rebuild on file changes
- [ ] **Partial builds** - Only rebuild changed files
- [ ] **Theme support** - Multiple built-in themes
- [ ] **Custom CSS** - User-provided stylesheets
- [ ] **Plugin system** - Extensible processors

## Version 2.0 (Scale & Performance)
*Timeline: 1 month*

### Enhanced TOC Searchability ⭐
**[Full specification](docs/features/enhanced-toc-searchability.md)**

- [ ] **Metadata-rich TOC** (modified date, size, tags, summaries)
- [ ] **Search indexing** - Sub-50ms search at 5000+ docs
- [ ] **Category manifests** - Organized document groups
- [ ] **Content hashing** - Change detection & caching
- [ ] **LLM-optimized format** - 60% smaller context size

**Impact:** 30x faster searches at 1000 docs, 224x at 5000 docs

## Version 2.1 (Enterprise Features)
*Timeline: 2 months*

- [ ] **Multi-repository support** - Aggregate docs from multiple sources
- [ ] **Authentication** - Protected documentation
- [ ] **Versioning** - Multiple doc versions side-by-side
- [ ] **Search API** - RESTful/GraphQL endpoints
- [ ] **Analytics** - Usage tracking, popular pages

## Version 3.0 (AI-Enhanced)
*Timeline: 3-4 months*

- [ ] **Auto-tagging** - NLP-based document classification
- [ ] **Semantic search** - Meaning-based document discovery
- [ ] **Smart summaries** - AI-generated document summaries
- [ ] **Cross-reference suggestions** - Automatic linking
- [ ] **Quality scoring** - Documentation completeness metrics

## Experimental Features
*No timeline - research phase*

- [ ] **Real-time collaboration** - Multiple users editing
- [ ] **Git integration** - Version control awareness
- [ ] **IDE plugins** - VSCode, IntelliJ extensions
- [ ] **Mobile app** - iOS/Android documentation readers
- [ ] **PDF export** - Print-ready documentation

## Performance Benchmarks

### Current (v1.0)
| Documents | Build Time | TOC Search | Memory |
|-----------|------------|------------|---------|
| 100       | 0.8s       | 12ms       | 45MB    |
| 1000      | 8.2s       | 450ms      | 312MB   |

### Target (v2.0)
| Documents | Build Time | TOC Search | Memory |
|-----------|------------|------------|---------|
| 100       | 0.5s       | 3ms        | 40MB    |
| 1000      | 3.0s       | 15ms       | 200MB   |
| 5000      | 12s        | 38ms       | 800MB   |

## Contributing

Interested in contributing? Check out:
- [Open Issues](https://github.com/onedusk/jot/issues)
- [Development Guide](docs/development.md)
- [Architecture Overview](docs/architecture.md)

## Feedback

Have ideas for features? Found a bug?
- Open an issue: [GitHub Issues](https://github.com/onedusk/jot/issues)
- Email: jot@onedusk.dev

---

*Last updated: 2025-10-04*
*Next review: 2025-11-01*

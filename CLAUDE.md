# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Jot is a documentation generator written in Go that converts markdown files into searchable documentation websites with LLM-optimized export capabilities. It replaces JetBrains Writerside. Key features include multi-format export (HTML, JSON, YAML, llms.txt, JSONL, enriched markdown), token-accurate chunking via tiktoken-go, and vector database-ready output.

## Build & Development Commands

```bash
make build          # Build binary (output: ./jot)
make test           # Run all tests (go test -v ./...)
make lint           # Run golangci-lint
make fmt            # Format code (go fmt ./...)
make coverage       # Generate HTML coverage report
make dev            # Live reload mode (runs jot watch)
make clean          # Remove build artifacts
make release        # Multi-platform release build (darwin/linux/windows)
```

Run a single test:
```bash
go test -v -run TestFunctionName ./pkg/chunking/
```

Run benchmarks:
```bash
go test -bench=. ./pkg/chunking/
```

## Architecture

### CLI Layer (`cmd/jot/`)

Entry point is `cmd/jot/main.go`. Uses Cobra for commands and Viper for configuration (reads `jot.yml`, supports `JOT_*` env vars). Commands: `build`, `export`, `init`, `toc`, `serve`, `watch`.

### Modules

Packages under `pkg/` are importable by other modules (mlpipe imports them through a `replace` directive) but are not API-stable before 1.0; list breaking changes in `docs/CHANGELOG.md` and keep mlpipe building. Packages under `internal/` are private.

| Module | Purpose |
|--------|---------|
| `pkg/scanner/` | Recursively discovers .md files, extracts metadata into `Document` structs. Applies glob-based ignore patterns. `Exclude` skips the output directory; `RelativeTo` names documents relative to the project root (`output.structure: project`). |
| `internal/toc/` | Builds hierarchical `TOCNode` tree from flat document list. Generates toc.xml per directory. |
| `internal/compiler/` | Orchestrates the build pipeline: scan → TOC → render → search index → assets. |
| `internal/renderer/` | Converts markdown to HTML using Blackfriday. Template-based page rendering. |
| `internal/htmlpath/` | Maps a markdown path to its page path (only a trailing `.md` becomes `.html`). |
| `pkg/export/` | Multi-format export adapters (JSON, YAML, llms.txt, llms-full, JSONL, markdown). Each format has its own file and test. |
| `pkg/chunking/` | Pluggable `ChunkStrategy` interface with Factory pattern. Implementations: `fixed` (token-based), `headers` (markdown headings), `recursive` (hierarchical), `semantic` (currently runs `fixed`). |
| `pkg/tokenizer/` | Wraps tiktoken-go with the `cl100k_base` vocabulary embedded, so it works offline. Provides `Encode()` and `Count()`. |
| `internal/search/` | Full-text search indexing. |
| `web/` | Embeds the CSS and JavaScript assets copied into every site. |

### Data Flow

```
Scanner (markdown files) → TOC Builder → Compiler/Exporter → Renderer → Output
```

For exports: documents are scanned, optionally chunked via a selected `ChunkStrategy`, then formatted by the appropriate exporter.

### Key Interfaces

- `chunking.ChunkStrategy` — implement `Chunk(doc, maxTokens, overlapTokens) ([]Chunk, error)` to add new chunking strategies
- Strategy instantiation via `chunking.NewChunkStrategy()` factory

### Export Workflow Presets

The `export` command provides shortcut presets:
- `--for-rag` → JSONL + semantic chunking + 512 tokens
- `--for-context` → Markdown + header chunking + 1024 tokens
- `--for-training` → JSONL + fixed chunking + 256 tokens

## Configuration

Project config lives in `jot.yml`. Config precedence: CLI flags > env vars (`JOT_*`) > jot.yml > defaults. Chunking constraints: chunk-size 1–2048 tokens, overlap must be less than chunk-size.

## Module Dependencies

```
cmd/jot → compiler, renderer, toc, export, scanner
compiler → renderer, search, toc, htmlpath, scanner, web
renderer → toc, htmlpath, scanner
search → htmlpath, scanner
toc → scanner
export → scanner, tokenizer
chunking → tokenizer, export (types), scanner (types)
```

## Hooks

The `.claude/settings.json` configures oober hooks for bulk code transformations on UserPromptSubmit, PreToolUse (Edit), and PostToolUse (Edit).

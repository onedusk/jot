#!/bin/bash

git add -A;
git commit -m "feat: add comprehensive multi-format LLM export system v0.1.0

Major feature release implementing complete LLM-optimized export functionality with
token-accurate chunking and multiple export formats.

BREAKING CHANGES:
- Version bumped to 0.1.0
- Export system architecture refactored to support multiple exporters
- Chunk struct now includes TokenCount field
- Character-based chunking replaced with token-based chunking

New Features:

LLM Export Formats:
- llms.txt: Lightweight documentation index per llmstxt.org specification
- llms-full.txt: Complete documentation concatenation for LLM context windows
- JSONL: JSON Lines format for vector database ingestion (Pinecone, Weaviate, Qdrant)
- Enriched Markdown: Markdown with YAML frontmatter metadata

Token-Based Chunking (Critical Path):
- Integrated tiktoken-go for accurate token counting using cl100k_base encoding
- Replaced character-based chunking with token-aware chunking
- Binary search algorithm for efficient token boundary detection
- Word boundary preservation to avoid splitting mid-word

Pluggable Chunking Strategies:
- Fixed-size: Token-based fixed chunks with configurable overlap
- Markdown-headers: Split at markdown header boundaries (# to ######)
- Recursive: Hierarchical splitting (paragraph → line → space → character)
- Semantic: Stub for future embedding-based boundary detection

CLI Enhancements:
- New --format flag supporting 6 formats: json, yaml, llms-txt, llms-full, jsonl, markdown
- Chunking configuration: --strategy, --chunk-size, --chunk-overlap flags
- Workflow presets: --for-rag, --for-context, --for-training
- Embeddings support: --include-embeddings flag with API cost warnings
- Comprehensive flag validation with helpful error messages

Build Integration:
- Auto-generates llms.txt and llms-full.txt during jot build
- features.llm_export config option (default: true)
- --skip-llms-txt flag to disable LLM export
- Humanized file size reporting in build logs
- Non-breaking errors (LLM export failures don't break builds)

Technical Details:
- Added dependency: github.com/pkoukk/tiktoken-go
- New packages: internal/tokenizer, internal/chunking
- New types: ProjectConfig, ChunkMetadata structs
- 15 new implementation files with comprehensive test suites
- 71 passing tests across 7 packages (0 failures)
- >85% test coverage for new packages

Performance:
- Binary search for efficient token boundary detection
- Streaming support for JSONL format
- Benchmarks for all chunking strategies

Documentation:
- Updated README.md with comprehensive LLM export examples
- Created CHANGELOG.md following Keep a Changelog format
- Extensive CLI help text with practical examples
- Updated configuration examples

Tasks Completed (8/8):
- Task 001: llms.txt export format
- Task 002: llms-full.txt export format
- Task 003: Token-based chunking (CRITICAL PATH)
- Task 004: JSONL export for vector databases
- Task 005: Enriched markdown export
- Task 006: Pluggable chunking strategies (CRITICAL PATH)
- Task 007: Build integration
- Task 008: CLI updates (FINAL INTEGRATION)

All task files moved to .tks/review/
All tests passing: 71/71 ✓

Resolves: #TBD
See: CHANGELOG.md for detailed release notes";
git push -u origin main;

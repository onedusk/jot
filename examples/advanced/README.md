# Advanced Example

This example demonstrates advanced Jot features including:

- Multiple input directories (`docs/` and `api/`)
- Custom ignore patterns
- LLM-optimized export
- Full-text search

## Building

```bash
# Regular build
jot build

# Export for LLM
jot export --format llm --output advanced-llm.json
```

## Features Demonstrated

- **Multi-source**: Combines docs from multiple directories
- **Ignore patterns**: Excludes drafts and private files
- **LLM export**: Optimized for AI consumption

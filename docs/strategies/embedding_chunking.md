# Embedding & Chunking Strategy

## Goal
Prepare transcript-derived text for high-quality semantic search and modeling by generating 768-d embeddings with durable metadata.

## Chunk Design
1. **Action-Centric Chunks**
   - Combine each ⏺ action with its ⎿ details (max ~8 sentences).
   - Maintain chronological order; include source IDs (`action_id`, `detail_ids`).
2. **Prompt Chunks**
   - Isolate `>` prompt lines as standalone records flagged `chunk_type=prompt`.
3. **Long Action Handling**
   - Apply sentence tokenization (e.g., `nltk` or `spacy`).
   - If chunk > 1200 characters, split and recursively merge adjacent sentences until within window.

## Metadata Schema
- `chunk_id`, `session_id`, `sequence_start`, `sequence_end`
- `action_ids`, `detail_ids`
- `verb`, `resources`, `status`
- `token_estimate`, `created_at`

## Embedding Pipeline
1. **Pre-flight Checks**
   - Validate normalized JSONL input.
   - Ensure ASCII-safe text or escape NBSP.
2. **Embedding Generation**
   - Use 768-d `nomic-embed-text` (or equivalent) with batching and retry logic.
   - Persist embeddings alongside metadata (`embeddings/chunks.parquet`).
3. **Quality Control**
   - Compute cosine similarity between adjacent chunks to detect anomalies.
   - Monitor average embedding norms; alert on drift.

## Deliverables
- Chunked JSONL ready for embedding calls.
- Embedding artifact store (Parquet + metadata).
- CLI (`scripts/embed_chunks.py`) supporting dry-run, batch size, and resume tokens.

## Key Decisions
- Token estimator choice (simple char count vs. model-specific tokenizer).
- Error handling strategy for API limits or transient failures.
- Storage for embeddings (vector DB vs. file system).

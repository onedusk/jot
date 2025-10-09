# Storage Schemas Strategy

## Goal
Select and implement durable storage patterns that balance analytics needs, pipeline ingestion, and operational simplicity.

## Candidates
1. **JSONL (Primary)**
   - One object per action with embedded details.
   - Fields: `id`, `session_id`, `sequence`, `action_text`, `details`, `verb`, `resources`, `created_at`, `raw_offsets`.
   - Pros: append-friendly, vector-store ready, easy to diff.
2. **CSV (Support)**
   - Flattened rows for spreadsheet tooling or ad-hoc SQL imports.
   - Columns: `action_id`, `detail_idx`, `type`, `text`, `verb`, `status`, `source_line`.
   - Requires NBSP escaping (`\\u00A0`).
3. **PostgreSQL**
   - Tables: `actions(action_id PK, session_id, sequence, verb, status, action_text, raw_line)` and `action_details(action_id FK, detail_idx, detail_text, facet_paths)`.
   - Indexes: `GIN` on `to_tsvector('english', action_text)`, B-tree on `(session_id, sequence)`.
   - Supports triggers for real-time automation.
4. **Parquet (Historical Archive)**
   - Partition by date/session for efficient analytics.
   - Schema mirrors JSONL but columnar encoded to compress repetitive verbs/resources.

## Plan of Attack
1. **Schema Definition**
   - Draft JSON schema covering required/optional fields.
   - Map schema to CSV headers and SQL DDL.
2. **Prototyping**
   - Generate sample outputs from a normalized dataset.
   - Validate compatibility with downstream consumers (analytics, vector indexing, BI tools).
3. **Migration Tooling**
   - Build converters: `jsonl -> csv`, `jsonl -> sql` (COPY statements), `jsonl -> parquet` using `pyarrow`.
   - Include checksum metadata to verify completeness.
4. **Governance**
   - Version schemas in `docs/spec` with change logs.
   - Establish data retention policies (session logs older than N days -> Parquet archive).

## Acceptance Criteria
- JSONL pipeline delivered with automated generation and schema validation.
- Successful round-trip load to PostgreSQL and Parquet verified on sample sessions.
- Documentation linking table/field definitions to extraction logic.

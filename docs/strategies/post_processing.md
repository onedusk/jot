# Post-Processing Strategy

## Goal
Transform the raw marker-based transcript into structured, analytics-ready objects while safeguarding formatting fidelity for downstream teams.

## Context
- Source: `2025-10-03-this-session-is-being-continued-from-a-previous-co.txt`
- Extraction: `scripts/extract_markers.py` produces `2025-10-03-extracted.txt`
- Constraints: preserve ⏺/⎿/› semantics, handle NBSP characters, avoid false positives from deeply indented or numbered lines.

## Approach
1. **Normalize Blocks**
   - Parse `2025-10-03-extracted.txt` into `{action, timestamp?, actor?, details:[...]}` objects.
   - Assign stable IDs (`action_{session}_{seq}`) and keep source line references for traceability.
   - Extract implicit timestamps or actors when present in text using regex/date parsing heuristics.
2. **Heuristic Tagging**
   - Verb taxonomy: map opening verb to canonical categories (`Inspect`, `Execute`, `Diagnose`, etc.).
   - Resource facets: detect file paths, commands, and tooling (`Bash(`, `.go`, `.html`).
   - Status signals: flag errors, warnings, successes based on keywords (`Error`, `Perfect`, `Great`).
3. **Session Aggregates**
   - Count actions per verb, resource, and status.
   - Derive durations when timestamps exist or infer order-based cycle times.
   - Emit summary rows (JSON or tabular) for session-level analytics.
4. **Whitespace Validation**
   - Build automated check ensuring NBSP (`\u00A0`) only appears where expected; log anomalies.
   - Offer ASCII-safe variants by replacing NBSP with literal spaces once validation passes.
5. **Automation & Testing**
   - Package logic in `scripts/post_process.py` (TODO) with CLI flags for dry-run and output formats.
   - Add unit tests covering marker grouping, verb extraction, NBSP handling.

## Deliverables
- Structured JSONL dataset per session.
- Session summary report (counts + highlights) as Markdown.
- Validation log describing normalization warnings.

## Metrics
- % of actions with classified verb and resource.
- Number of NBSP anomalies per session.
- Processing throughput (actions/sec) for baseline sizing.

## Dependencies
- Stable extraction file (`scripts/extract_markers.py`).
- Optional timestamp enrichment from upstream logging.

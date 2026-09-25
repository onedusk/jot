# Decision 2: scope of jot's LLM export features

Status: proposed. Nothing gets implemented until you approve it.

**Evidence and base.**
- File:line citations with no commit named refer to jot at `2acde71` (the audit snapshot) and to the mlpipe working tree. Citations marked "at head" refer to `1a7a6c7`.
- The plan applies on top of `1a7a6c7`, the current head of `feat/section9-decisions`. #1 (project-relative paths), #4 (embedded tokenizer), #5 (dead code) and #6 (pkg/ stability note) have landed there.
- Three changes after `90b5189` matter here:
  - `86a60e5` deleted `export.minInt` and `LLMDocument.HTML`.
  - `b1b524c` made `dedupeDocuments` (`cmd/jot/build.go:201` at head) treat a file reached through a symlink or a differently-cased path as one document. (After this document was written, `67a602e` narrowed that to spellings that map to the same page, keeping symlink aliases as separate pages; nothing in this design depends on the difference.)
  - `5f19e30` moved the document-ID hash to the forward-slash path (`pkg/scanner/scanner.go:186-190` at head).
- `cmd/jot/export.go` has not changed since `90b5189`. Compared with `2acde71`, its lines after 196 sit 6 to 7 lines lower, because #1 added project-root resolution.
- Each measurement names the commit it was taken at.

**Provenance labels:**
- **re-verified**: measured again for this document, in scratch copies.
- **judge-verified**: a reviewer reproduced it against the design prototypes.
- **reported**: one design measured it and nobody re-ran it.

Every mlpipe gate result in this document was re-run with a corrected gate (Section 8, standing checks). The gate script first used on this branch piped `go test` into `grep`, so it reported failing tests as passing; it has since been fixed, and mlpipe's build, vet and tests were re-run against every branch commit that claimed them.

**Prototypes.** Commit hashes such as `6163690`, `2114468`, `124f9e9`, `1e393e1` and `8af10ce`, and paths under `scratchpad/`, refer to throwaway prototypes built in a scratch directory while designing. They are not in this repository; they are cited so each measurement names what it ran against.

---

## 1. Question and short answer

**The question** (audit Section 9, #2; finding A16): embeddings and semantic chunking need an embedding provider, API keys and cost controls. Should jot provide them? And what should happen to `--include-embeddings`, the `semantic` strategy and `--for-rag` in the meantime?

**Short answer:**
- jot does not compute embeddings and does not do embedding-based chunking.
- It chunks deterministically and offline, gives each chunk a unique ID, and serializes what the caller gives it, including vectors.
- The defect to fix is on jot's side. Its export API re-chunks from raw documents, which discards both the caller's chosen strategy and the caller's vectors.
- The fix has two parts:
  - a small records API in `pkg/export`;
  - deleting the flags, stubs and fields that promise more than jot does.
- `--for-rag` moves to the `recursive` strategy, after a positioning bug in that strategy is fixed.

**What mlpipe answers.** mlpipe is the only consumer of `pkg/`, and it has already decided who owns embeddings.

- **The caller owns embeddings.**
  - mlpipe defines a provider-agnostic interface, `Embedder { Embed(ctx, texts []string) ([][]float32, error) }` (`pkg/mlpipe/types.go:42-45`), and injects it with `WithEmbedder` (`options.go:39-44`).
  - It runs the embedder after chunking (`client.go:317-321`) and stores the vectors in jot's own row type, `export.ChunkMetadata.Vector` (`client.go:370-372`).
- **Embedding is a library step, even in mlpipe.** mlpipe links no provider SDK, and its CLI wires no embedder (`cmd/mlpipe/process.go:84-93`, `internal/config/config.go:107-122`).
- **mlpipe offers `semantic` only because jot lists it.** It validates strategy names against `chunking.AvailableStrategies()` (`client.go:67-77`, `config.go:88-98`). Its own measurements show `semantic` and `fixed` producing identical results: 175 chunks and 83,502 tokens (`docs/internal/export-bugs.md:25-31`).
- **mlpipe then loses its work inside jot.**
  - `ToJSONL(result.Documents, ...)` (`client.go:326`) re-chunks with jot's private fixed-size chunker, so the chosen strategy and the vectors are both dropped.
  - Its `.test/mangroves` output is md5-identical across all five strategies: 175 rows, 47 distinct `chunk_id` values over 26 documents, and no vectors (judge-verified).
  - Its embedder test uses output format `none` (`client_test.go:260-263`), which is why this went unnoticed.
- **Only one of mlpipe's two proposed fixes works.** `export-bugs.md:19-21` offers two:
  - (a) `ToJSONLFromChunks` keeps the vectors.
  - (b) "`ToJSONL` takes a strategy" fixes the strategy, but still re-chunks after embedding and so drops the vectors again.

**What mlpipe does not answer** is whether anyone wants embeddings without writing Go.
- Nothing in either repository asks for that.
- The flag came from a scaffolding plan, not from a user: `.claude/agents/cli-dev.md:27` at `2acde71` says "Add `--include-embeddings` flag for JSONL with cost warning". The uncommitted working tree deletes `.claude/`.
- If the need appears, mlpipe's CLI should host it, not jot. mlpipe already has the embedding step and a config file.

---

## 2. Current state

### jot (`2acde71`)

| # | Problem | Evidence |
|---|---|---|
| 1 | `--strategy` is validated and echoed, then ignored (A6) | `cmd/jot/export.go:135-146` validates it, `:279` prints it, and `:281` calls `ToJSONL(allDocs, chunkSize, chunkOverlap)` without it |
| 2 | JSONL and LLM export re-chunk with a private copy of the fixed chunker (D2) | `pkg/export/jsonl.go:41-50`, `pkg/export/export.go:83`, `:218-331` |
| 3 | `chunk_id` is `chunk-N` within each document, so it repeats across documents (A7) | `pkg/export/export.go:228,281`. The strategies already emit `<docID>-chunk-<n>` (`pkg/chunking/fixed.go:32,85`) |
| 4 | `pkg/chunking` imports `pkg/export` only for `Chunk`, so export cannot use chunking (D1) | `pkg/chunking/strategy.go:5,21`; the type is at `pkg/export/types.go:36-45` |
| 5 | `semantic` is the fixed chunker under another name, and its boundary detector is dead code | `pkg/chunking/semantic.go:44-48,67-69`; `factory.go:14,31-32,34,45` |
| 6 | The CLI accepts `contextual`, which the factory does not know, and rejects the factory's `headers` | `cmd/jot/export.go:36,62,136,145`; `pkg/chunking/factory.go:24-35` (`headers` is at `:27`) |
| 7 | `--include-embeddings` prints cost warnings and does nothing | `cmd/jot/export.go:51-52,72,148-152,170,237-242`; README 142-143. `jot.yml:44` lists `embeddings` under `llm.export_formats`, a section nothing reads (B3). `jot init` does not write that key (`cmd/jot/init.go:64-66`) |
| 8 | `--for-rag` selects `semantic` | `cmd/jot/export.go:177-182` |
| 9 | Markdown export prints the strategy and size, then ignores both, so `--for-context` produces the same output as `--format markdown` | `cmd/jot/export.go:183-188,283-290`; `pkg/export/markdown.go:45` |
| 10 | Markdown stubs | `separateFiles` returns "not yet implemented" (`markdown.go:45-49`); `contextualEnrichment` returns "" (`:89-94,146-162`) |
| 11 | The legacy `llm` branch is unreachable | `cmd/jot/export.go:292-301`; validation rejects `llm` first (`:122`) |
| 12 | `recursive` accepts `overlapTokens` and never uses it | `pkg/chunking/recursive.go:29-51` |
| 13 | **New finding N1:** `recursive` offsets drift | `pkg/chunking/recursive.go:98-119`. See below |
| 14 | `Chunk.Vector` is never set by jot or mlpipe, yet `ToJSONL` copies it into each row, so it looks like a second place to put a vector | `pkg/export/types.go:44`; `pkg/export/jsonl.go:62`. mlpipe writes `ChunkMetadata.Vector` instead (`client.go:371`) |
| 15 | `--chunk-overlap` help says ">0", but validation allows 0 | `cmd/jot/export.go:64` vs `:111` |
| 16 | `ToLLMFormat` silently turns overlap 0 into 128 | `pkg/export/export.go:63-65` |

**N1 in detail:**
- **Cause:** when a split part is empty (a leading separator or a run of blank lines) and nothing has been grouped yet, `recursiveSplit` drops the part and its separator without advancing `currentOffset`. Every later chunk's `start_pos`/`end_pos` then points at the wrong text.
- **Measured (re-verified):**

  | Corpus | Settings | Chunks with `content[StartPos:EndPos] != Text` |
  |---|---|---|
  | claude-swarm-py markdown, 74 docs | 512 tokens, overlap 0 | 34 of 414 |
  | 484 documents: claude-swarm-py markdown, jot docs/ at `90b5189`, 400 synthetic documents and one minimal repro | 128 tokens, overlap 0 | 3,078 of 6,480 |
  | jot docs/ at `90b5189` | 512 tokens, overlap 0 | 0 |

- jot's docs/ trigger nothing, which is why design A, which measured on jot's docs/ at `2acde71`, missed it. `fixed` and `markdown-headers` have no mismatches on any of these corpora.

### mlpipe (working tree)

| Problem | Evidence |
|---|---|
| JSONL drops the strategy and the vectors | `client.go:326` |
| LLM output re-chunks too, and is never written to disk (Bug 2) | `client.go:334`; `cmd/mlpipe/process.go:125-135`. The proposed fix at `export-bugs.md:54-58` compares a `*LLMExport` with `""` and would not compile |
| `ProcessMixed` appends documents from every scan root with no checks (judge-verified for the first case) | `client.go:205-224`. `a/README.md` and `b/README.md` get the same `doc_id` and `chunk_id`. A directory listed twice, or a PDF output directory inside a markdown directory, yields the same file twice, with identical IDs. Nested roots (`docs` and `docs/sub`) yield the same file under two different IDs |
| `semantic` is advertised | `README.md:9,56,126`, `mlpipe.yml:14`, `options.go:7`, `cmd/mlpipe/process.go:45` |
| Every chunk text goes to the embedder in a single call | `client.go:355-360` |

**Already resolved by #1:** design B's concern that document IDs must be computed from the final path, not rewritten after scanning.
- At head, `Scanner.RelativeTo` prefixes the path before `readDocument` computes the ID (`pkg/scanner/scanner.go:147,190` at head).
- In jot's CLI, `dedupeDocuments` (`cmd/jot/build.go:201` at head) drops repeat scans of the same file and rejects distinct files that would share a path, in both layouts. That keeps doc IDs unique too.

---

## 3. Decision and rationale

### Decisions

1. **No embedding provider in jot.** jot ships no provider, no embedding config and no embedding-based strategy. `ChunkMetadata.Vector` is the caller's only vector slot, and jot writes it whenever it is set.
2. **Add a records seam to `pkg/export`.**
   - `ChunkRecords` turns one document's chunks into JSONL records.
   - `ToJSONLFromChunks` writes records exactly as given and rejects a repeated `chunk_id`.
   - The CLI and mlpipe both chunk with the chosen strategy, optionally set vectors, and then serialize.
   - This one mechanism fixes A6, A7 and mlpipe Bug 1.
3. **Move `Chunk` into `pkg/chunking` (D1)** with no alias, drop its unused `Vector` field, and delete export's private chunker (D2).
4. **Delete what promises more than it does:** `semantic`, the `contextual` CLI value, `--include-embeddings`, the two markdown stubs and the unreachable `llm` branch.
5. **Fix N1, then point `--for-rag` at JSONL + `recursive` + 512 tokens + overlap 0.**
6. **Document that `recursive` does not overlap.** The help says the overlap is validated but not applied, and the progress line reports the overlap actually applied (0).
7. **Markdown export stays whole-document.** `--for-context` is removed, and the help says the chunking flags apply to JSONL only (open question 1).
8. **The LLM export format** is recommended for removal from both repositories (open question 2). Until you decide, it keeps its signature and chunks with the shared fixed strategy.

### Rationale

**Ownership.** mlpipe already made this call (Section 1). jot's job is to not destroy the caller's work.

**Determinism.**
- Chunk IDs are positional (`<docID>-chunk-<n>`). Boundaries that depend on an embedding model would re-key every vector store whenever the provider changed models, and would make export output non-reproducible (A12).
- #4 just made jot hermetic, and a network provider would undo that.

**Why records, not a strategy parameter.** Vectors are computed between chunking and serialization, so any exporter that chunks internally loses them.

**Why a single vector slot.** `Chunk.Vector` is a trap. A caller who stores embeddings on the chunks a strategy returns would lose them in `ChunkRecords`, which is the same class of bug this design fixes. Nothing sets that field, so J1 deletes it, and vectors live only on `ChunkMetadata`.

**Why a per-document `ChunkRecords` instead of design B's `ChunkDocuments(docs, strategy, ...)`:**
- export never runs strategies itself.
- mlpipe keeps its per-document `ctx` check.
- The API stays smaller.

**Why a duplicate-ID check in the serializer:**
- jot's CLI cannot produce duplicates (`dedupeDocuments`), so the check never fires there.
- For library callers, the failure mode without it is silent overwrites in a vector database.
- The check is about eight lines. It names both records' sources, because a duplicate has three possible causes:
  - the same document passed twice;
  - two scan roots that contain the same relative path;
  - two documents given the same ID by hand.
- It runs at serialization, after any embedding, so it is a backstop, not the first line of defense. mlpipe M3 checks its inputs before chunking, and the check catches every other caller.

**Why `recursive` for RAG:**

| Corpus and settings | fixed | markdown-headers | recursive | Provenance |
|---|---|---|---|---|
| claude-swarm-py markdown, 74 docs, 512/0 | 387 chunks | 2,102 | 414 | re-verified at `90b5189` |
| claude-swarm-py markdown, 512/128 | 485, median 511 tokens | 2,107, median 49, 59% under 64 tokens | 414, median 476 | reported (B) |
| jot docs/ at `2acde71`, 512/128 | 69, median 511, 10th percentile 312 | 271, median 63, 10th percentile 8, 80 chunks under 32 tokens | 64, median 450, 10th percentile 201 | reported (A); the 64 is judge-verified |
| mlpipe mangroves, 26 docs, 512/128 | 175 | 593 | 154 | `export-bugs.md:25-31` |

- `markdown-headers` emits one chunk per heading and never merges small sections (`headers.go:83-96`), so at RAG sizes most of its chunks are fragments.
- `recursive` packs paragraphs, then lines, up to the size limit. `fixed` cuts wherever the token count runs out.
- The cost of `recursive` is that it has no overlap.

### Rejected alternatives

| Alternative | Why rejected |
|---|---|
| Design B's Phase B: `pkg/embed` (an OpenAI-compatible client and a cache), an `embeddings:` config section, `--embed`, a context-aware strategy interface and a real semantic strategy | No one asked for it, and the only consumer chose injection. It brings back network access, secrets and non-reproducible output right after #4 removed them. `ChunkStrategy.Chunk` has no `context.Context`, and the name-only factory cannot pass an embedder, so a second chunking API would be needed. B estimated 500-650 non-test lines; its `pkg/embed` prototype alone was 222 (reported). B also cites Qu et al. (arXiv:2410.13070) as finding no consistent benefit from embedding-based chunking over fixed-size chunking; that citation was not checked for this document, and the rejection does not depend on it. |
| Keep `semantic` as a documented alias of `fixed` (the audit's other option) | Leaves a misleading name in user configs, and mlpipe documents it. "Semantic chunking" means embedding breakpoints in LangChain and LlamaIndex. |
| `ToJSONL` takes a `ChunkStrategy` (mlpipe's option b) | Fixes A6 but still drops the vectors. |
| Design B's `ChunkDocuments(docs, strategy, max, overlap)` | Runs strategies inside export. mlpipe would have to wrap each document in a slice to keep its `ctx` checks. |
| A `type Chunk = chunking.Chunk` alias in export | Not needed: mlpipe never names `export.Chunk`. Re-verified with the corrected gate: against the judge's D1 prototype (no alias), mlpipe builds, vets and passes its tests. |
| Keep `Chunk.Vector` and have `ChunkRecords` copy it | One line, but it keeps two places for a vector, and nothing in jot or mlpipe sets this one. Deleting it leaves `ChunkMetadata.Vector` as the only slot and shrinks J11. |
| Leave the duplicate-ID problem to mlpipe (design C) | Every library caller keeps silently overwriting vectors until mlpipe M3 lands. |
| Design A's plan: make `recursive` the RAG preset without fixing N1 | Ships wrong offsets. A's own `--for-rag` binary wrote 34 of 414 rows with bad offsets on claude-swarm-py (judge-verified; the drift itself re-verified above). |
| Design B's N1 fix (count the grouped parts) | Emits empty and whitespace-only chunks (judge-verified). `"\n\n"` followed by an oversized paragraph yields `d-chunk-0` with `Text ""`, and blank chunks rise from 52 to 686 on a 400-document synthetic set. C's per-part offset fix leaves the chunk set unchanged (re-verified: 414 chunks, identical texts, 0 mismatches). |
| Implement overlap in `recursive` | A new algorithm nobody asked for. |
| Reject an explicit `--chunk-overlap` with `recursive` (design C's step 7) | This design deviates from the audit here: the Phase 2 row says "implement overlap in `recursive` or reject it". Rejection means about 8 lines that use `cmd.Flags().Changed`. pflag's `Set` marks a flag `Changed`, and the existing cleanup at `e2e_test.go:401` restores the value but leaves `Changed` set, so every test that sets export flags would also have to reset it. C prototyped this and it passed `-shuffle=on -count=5`. The help line plus the effective value in the progress line gives users the same information without flag-state logic. It is easy to add later. |
| Skip the overlap-versus-size check when the strategy is `recursive` | The check runs in `validateExportFlags` (`export.go:115-118`) before presets are applied (`:160` vs `:177-195`), so the override would have to be applied in two places. The same default-overlap failure hits every strategy (Section 10). The help instead says the overlap is validated but not applied. |
| Reject explicitly set chunking flags for json, yaml, llms-txt, llms-full and markdown | Same `Changed()` cost, across three flags. The help text states the flags apply to JSONL only. |
| A `sections` strategy that packs whole heading-delimited sections up to the limit | In C's simulation (301 files, 512/0; reported), 81% of its boundaries fall at headings, against 12% for recursive. But it is a new strategy, it needs the fence-aware line iterator from A10, and #2 does not need it. Revisit after A10. |

---

## 4. API design (final state)

### `pkg/chunking`

```go
// chunk.go: moved from pkg/export/types.go (36-44 at head), without the
// Vector field, which nothing set. Vectors belong on export.ChunkMetadata.
// Text is doc.Content[StartPos:EndPos].
type Chunk struct {
	ID         string `json:"id" yaml:"id"`
	Text       string `json:"text" yaml:"text"`
	StartPos   int    `json:"start_pos" yaml:"start_pos"`
	EndPos     int    `json:"end_pos" yaml:"end_pos"`
	TokenCount int    `json:"token_count" yaml:"token_count"`
}

// strategy.go: no longer imports pkg/export.
type ChunkStrategy interface {
	Chunk(doc scanner.Document, maxTokens, overlapTokens int) ([]Chunk, error)
}

// factory.go
// NewChunkStrategy accepts "fixed", "headers" or "markdown-headers", and "recursive".
// The error for any other name lists AvailableStrategies().
func NewChunkStrategy(name string, tok tokenizer.Tokenizer) (ChunkStrategy, error)
func AvailableStrategies() []string // {"fixed", "headers", "markdown-headers", "recursive"}

// recursive.go (doc comment only):
// overlapTokens is accepted to satisfy ChunkStrategy and is not applied;
// recursive chunks do not overlap.
```

`SemanticStrategy`, `NewSemanticStrategy` and `semanticBoundaryDetection` are deleted.

### `pkg/export`

```go
// jsonl.go

// ChunkRecords returns one record per chunk of doc, in order, with DocID,
// Source (doc.RelativePath) and the PrevChunkID/NextChunkID links set.
// Vector is left empty for the caller to fill.
func ChunkRecords(doc scanner.Document, chunks []chunking.Chunk) []ChunkMetadata

// ToJSONLFromChunks writes each record as one line of compact JSON, in the
// order given. It does no chunking, so records keep the strategy that made
// them and any Vector the caller set. It returns an error if two records
// share a ChunkID.
func (e *JSONLExporter) ToJSONLFromChunks(records []ChunkMetadata) (string, error)

// Deprecated: chunk with a chunking.ChunkStrategy, build records with
// ChunkRecords, and write them with ToJSONLFromChunks. ToJSONL always uses
// the fixed strategy. It is removed once mlpipe has migrated (step J10).
func (e *JSONLExporter) ToJSONL(documents []scanner.Document, maxTokens, overlapTokens int) (string, error)
```

The duplicate check remembers the first `Source` for each `ChunkID` and produces this error:

```go
fmt.Errorf("duplicate chunk_id %q (from %s and from %s): pass each document once and give each a unique ID; files from two scan roots with the same relative path get the same ID, which Scanner.RelativeTo avoids", r.ChunkID, first, r.Source)
```

The two sources are often the same path: the same file passed twice, or two roots that contain the same relative path. They differ when a caller set document IDs by hand. The message reads correctly in all three cases.

```go
// types.go
type LLMDocument struct {
	// ... unchanged fields ...
	Chunks []chunking.Chunk `json:"chunks" yaml:"chunks"`
	// ...
}
// ChunkMetadata is unchanged. The Vector comment becomes:
// "Set by the caller; jot never computes embeddings. Omitted when empty."

// export.go: signature unchanged until open question 2 is resolved.
// It chunks with chunking.NewFixedSizeStrategy and says so in its doc comment.
func (e *Exporter) ToLLMFormat(documents []scanner.Document, chunkSize, overlap int) (*LLMExport, error)

// markdown.go: the separateFiles parameter and contextualEnrichment are removed.
// NewMarkdownExporter keeps its signature.
func (m *MarkdownExporter) ToEnrichedMarkdown(documents []scanner.Document) (string, error)
```

**Deleted:**
- `chunkDocument` and `maxInt` (`export.go:218-331,376-382` at head). `minInt` is already gone (`86a60e5`).
- `contextualEnrichment` (`markdown.go:146-162`)

**Dependencies afterwards:** `export → chunking, scanner, tokenizer` and `chunking → scanner, tokenizer`.

### `cmd/jot`

```go
// exportJSONL chunks docs with the named strategy and returns JSONL.
func exportJSONL(docs []scanner.Document, strategyName string, chunkSize, chunkOverlap int) (string, error)
```

### If open question 2 is answered "keep"

```go
// ToLLMFormatFromChunks builds the LLM export from documents and the records
// made for them, attaching each document's records in order, vectors included.
// It returns an error on a duplicate ChunkID.
func (e *Exporter) ToLLMFormatFromChunks(documents []scanner.Document, records []ChunkMetadata) (*LLMExport, error)
```

- Under "keep", `LLMDocument.Chunks` becomes `[]ChunkMetadata`, because `chunking.Chunk` no longer has a vector field.
- This changes the JSON keys of LLM chunks: `id` becomes `chunk_id`, and `doc_id`, `source`, the links and `vector` are added.
- No consumer sees the change. jot's CLI cannot produce the format, and mlpipe does not write it to disk until M4 (Keep) fixes Bug 2.

---

## 5. CLI and config changes

| Item | Before | After |
|---|---|---|
| `--strategy` | Accepts fixed, semantic, markdown-headers, recursive, contextual; ignored | Validated against `chunking.AvailableStrategies()`: fixed, headers, markdown-headers, recursive. Applied to JSONL |
| `--chunk-size`, `--chunk-overlap` | Help implies every format; the overlap help says ">0" | Help says "jsonl only". Overlap help: ">=0 and less than chunk-size; validated but not applied by recursive" |
| Progress line for JSONL | Prints the requested overlap | Prints the applied overlap: after presets, `if strategy == "recursive" { chunkOverlap = 0 }` |
| `--for-rag` | jsonl + semantic + 512/128 | jsonl + recursive + 512 + overlap 0. Stderr: "Using RAG preset: jsonl format, recursive strategy, 512 token chunks, no overlap" |
| `--for-training` | jsonl + fixed + 256/64 | Unchanged. Its chunk IDs become unique |
| `--for-context` | Identical to `--format markdown` | Removed (open question 1). The mutual-exclusion error (`export.go:98`) lists only `--for-rag` and `--for-training` |
| `--include-embeddings` | Warnings only | Removed |
| Legacy `llm` branch | Unreachable | Removed, along with the `encoding/json` import |
| Markdown progress line | Prints strategy and chunk size | Prints neither |

The help text replaces the strategy list and drops the embeddings and markdown-headers examples:

```
Chunking (jsonl only; --strategy, --chunk-size and --chunk-overlap do not
affect other formats):
  - fixed:            Fixed-size token chunks with word boundaries (default)
  - markdown-headers: One chunk per heading section; large sections are split
                      (alias: headers)
  - recursive:        Packs paragraphs, then lines, then words up to the chunk
                      size. Chunks do not overlap; --chunk-overlap is validated
                      but not applied

jot does not compute embeddings. To add them, embed the "text" field of each
JSONL row and store the result as "vector", or build rows in Go with
export.ChunkRecords and write them with ToJSONLFromChunks (mlpipe does this).
```

**Config:**
- Remove `- embeddings` from `jot.yml:44`. `jot init` never wrote it.
- No new configuration keys.

**Docs to update** (line numbers at head):
- README 17, 18, 134-135, 139-143, 281 ("Vector field for embeddings" becomes "`vector`, present when a library caller sets it"), 288, 293, 296, 301-305, and 375 ("All formats use token-based chunking" is false: only JSONL chunks).
- CLAUDE.md:
  - 50 (the `semantic` row), 65, and 71-72 (the presets);
  - 82-88, the dependency diagram. J1 changes the `export` and `chunking` lines, because export imports chunking from J1 on (`LLMDocument.Chunks`). J5 adds `chunking` back to the `cmd/jot` line, which head dropped because `cmd/jot` does not import it today.
- `docs/spec/requirements.md` FR-018: say jot carries caller-supplied vectors and does not compute them.
- `docs/CHANGELOG.md` in every commit.
- **Outside jot's repository:** the workspace file `/Users/macadelic/dusk-indust/shared/packages/CLAUDE.md:99-100` lists `semantic`, the `semantic` RAG preset and `--for-context`.
  - It is untracked: the `dusk-indust` repository has no commits.
  - Edit it by hand when J6 (line 99 and the `--for-rag` entry) and J9 (the `--for-context` entry) land.

---

## 6. mlpipe changes

### Order that keeps mlpipe compiling

| jot step | mlpipe state required | Effect on mlpipe |
|---|---|---|
| J1-J4 | unmodified | Compiles. mlpipe never reads `Chunk.Vector`, which J1 deletes. JSONL and LLM chunk IDs become unique (J2). A `ProcessMixed` collision now errors at JSONL export (J3). `recursive` offsets become correct (J4) |
| J5 | unmodified | None (CLI only) |
| J6 | unmodified compiles; M2 the same day | `semantic` fails validation, and the error lists the supported strategies |
| J7-J9 | unmodified | None; mlpipe does not use `MarkdownExporter` |
| J10 | **M1 landed** | Without M1, the build fails with `exporter.ToJSONL undefined` |
| J11 | **M4 landed** | Under "drop", the build fails without M4 |

**Re-gated for this revision with the corrected gate** (Section 8), using design C's prototype (`scratchpad/design-semantic/jot`). All re-verified:
- **Unmodified mlpipe, C's steps 1-7** (`6163690` to `2114468`): build, vet and test each exit 0.
- **C's `ToJSONL` deletion** (`124f9e9`): the build fails with `exporter.ToJSONL undefined`, as predicted.
- **The same commit with C's M1 diff applied:** all three exit 0.
- **Current head `1a7a6c7` with unmodified mlpipe:** all three exit 0.
- **J1 as specified here**, with `Chunk.Vector` deleted, built on the judge's D1 prototype (`8af10ce`, which predates #4): jot's tests pass, and mlpipe builds, vets and passes its tests.

Steps J3, J8 and J9 of this plan differ from C's prototype. They change only symbols mlpipe does not call, plus the duplicate check inside `ToJSONL`. Run the gate on each commit rather than assume it passes.

### Required

**M1, before J10: migrate `chunkAndExport` (`client.go:275-338`).**
- Replace the hand-built `ChunkMetadata` loop with `records := export.ChunkRecords(doc, chunks)`.
- Sum `TokenCount` over the records and append them to `result.Chunks`.
- Write JSONL with `export.NewJSONLExporter().ToJSONLFromChunks(result.Chunks)` after embedding.
- Design C's prototype of this change removes 20 lines and adds 5.
- Add `TestProcessMarkdown_VectorsAndStrategyReachJSONL`, using the headers strategy, the mock embedder and jsonl output. Assert:
  - the number of JSONL rows equals `len(result.Chunks)`;
  - each row's `chunk_id` equals `result.Chunks[i].ChunkID`;
  - every row has a 384-dimensional vector;
  - IDs are unique.
- The test fails on today's client ("JSONL has 1 rows, in-memory 3"; judge-verified baseline: 3 chunks, 1 row, 0 vectors).
- Mark `export-bugs.md` Bug 1 fixed.

**M2, the same day as J6:** remove `semantic` from:
- `README.md:9,56,126` (line 9 also says "5 chunking strategies")
- `mlpipe.yml:14`
- `options.go:7`
- `cmd/mlpipe/process.go:45`

Runtime validation already follows `AvailableStrategies()`.

**M3 (required for correctness; not a compile gate): make `ProcessMixed` check its scanned documents before chunking** (`client.go:205-224`).
- **First, skip a document whose `Path` was already seen**, as jot's `dedupeDocuments` does. `NewScanner` makes the root absolute, so `Path` is absolute (`pkg/scanner/scanner.go:29-33`), and one check covers three cases:
  - a directory listed twice;
  - nested roots such as `docs` and `docs/sub`;
  - a PDF output directory inside, or equal to, a markdown directory.

  jot's version (`cmd/jot/build.go:201` at head) also resolves symlinks and case-only differences. That is mlpipe's choice to copy.
- **Then reject distinct files that share a relative path.** The error names both absolute paths, for example `mlpipe: /x/a/README.md and /x/b/README.md have the same relative path README.md; their chunk IDs would collide`.
- **Run both checks before chunking.** After M1, jot's duplicate check runs in `ToJSONLFromChunks`, after the embedder (`client.go:317-326`), so on its own it reports a collision only after the embeddings have been paid for. It also misses `Result.Chunks` with format `none`.
- **Tests:**
  - `a/README.md` and `b/README.md` return an error naming both absolute paths;
  - the same directory passed twice yields each document once.
- **Alternative:** instead of rejecting same-relative-path files, use `Scanner.RelativeTo` with a common parent directory, which keeps both files. That is mlpipe's choice.

**M4, before J11, depends on open question 2:**
- **Drop.** Remove the `llm` output format. From reading only; not gated against a J11 tree:
  - `pkg/mlpipe/client.go:62-64` (format validation) and `:332-338` (the `llm` case that calls `ToLLMFormat`)
  - `pkg/mlpipe/types.go:16` (`Result.LLMExport`)
  - `pkg/mlpipe/options.go:32` (the `WithOutputFormat` doc comment)
  - `internal/config/config.go:83-85`
  - `pkg/mlpipe/client_test.go:57`: the "all valid options" case passes `WithOutputFormat("llm")`, so switch it to `"jsonl"`. Delete `:117-122`, the "LLM format output" case.
  - `cmd/mlpipe/process.go:23,44` (help)
  - `README.md:10,55,123,144,169`, `mlpipe.yml:11`, `CLAUDE.md:62`

  The critic measured the first gap: with only the validation maps and `:117-122` changed, `TestNew/all_valid_options` fails with `unsupported output format "llm"`.
- **Keep.** Switch to `ToLLMFormatFromChunks(result.Documents, result.Chunks)`, and fix Bug 2 by writing `result.LLMExport` as JSON when it is `!= nil`.

### Recommended

- Document that `Embedder` implementations receive every chunk text in one call and must batch to provider limits (`client.go:355-360`), or batch inside mlpipe.
- Update the data-flow diagrams (`README.md:167-169`, `CLAUDE.md:60-62`) to `[]chunking.Chunk` and `ToJSONLFromChunks`.

---

## 7. Compatibility and migration

**Go API.** `pkg/` is not API-stable (#6), and mlpipe moves in lockstep. These breaking changes are listed in the CHANGELOG:
- `export.Chunk` becomes `chunking.Chunk`, without its `Vector` field, and `ChunkStrategy` returns `[]chunking.Chunk`.
- `SemanticStrategy` and `NewSemanticStrategy` are removed, and `AvailableStrategies()` no longer lists `semantic`.
- `ToEnrichedMarkdown` loses its `separateFiles` parameter.
- `ToJSONL` is removed after M1.
- Depending on open question 2, `ToLLMFormat` and the LLM types are removed, or `LLMDocument.Chunks` becomes `[]ChunkMetadata`.

**Output:**
- **Chunk IDs change.** JSONL `chunk_id`, `prev_chunk_id` and `next_chunk_id` change from `chunk-N` to `<doc_id>-chunk-N` (J2).
  - Vector stores need a full re-ingest.
  - Because IDs are positional, consumers should replace all records for a `doc_id` on each export rather than upsert single chunks.
  - Switching `output.structure` (#1) re-keys every `doc_id` and `chunk_id` again, as the #1 CHANGELOG entry already says.
- **`--strategy` now changes JSONL output.**
- **`--for-rag` output changes:** recursive at 512 with no overlap, where it used to be fixed at 512 with 128 of overlap. Expect fewer chunks.
- **Markdown export is unchanged** under the recommended answer to open question 1. The enrichment stub never wrote anything.

**CLI:**
- `--include-embeddings` and `--for-context` now fail with Cobra's unknown-flag error and usage hint.
- `semantic` and `contextual` now fail with the supported-strategies list.

**mlpipe users:**
- `semantic` in a config fails validation after J6.
- `ProcessMixed` with colliding roots fails at JSONL export from J3. After M3:
  - a directory listed twice, or nested roots, are deduplicated;
  - distinct files with the same relative path fail before chunking, with both paths named.

---

## 8. Implementation plan

**Standing checks for every commit:**
- `gofmt -l .` is empty.
- `go vet ./...` and `go test -race ./...` pass.
- The mlpipe gate passes. It builds, vets and tests mlpipe against jot's working tree and records each exit status on its own, before any output filtering. It also confirms that mlpipe's `go.mod` and `go.sum` are unchanged.

  ```sh
  cd ../mlpipe
  go build ./... ; b=$?
  go vet ./... ; v=$?
  go test -count=1 ./... >"$SP/gate-test.out" 2>&1 ; t=$?
  /usr/bin/grep -v 'no test files' "$SP/gate-test.out"
  # pass only if b, v and t are all 0 and go.mod/go.sum are unchanged
  ```

  - Record each exit status before filtering output. An earlier version of the gate piped `go test` into `grep -v` and recorded grep's exit status, so failing tests passed the gate (build and vet failures still stopped it).
  - The results in this document were re-run with `scratchpad/design-rev2/gate.sh`. It reports the three statuses separately and copies jot, mlpipe and swiper side by side, because mlpipe's `go.mod` replaces both jot and swiper.
- There is a `docs/CHANGELOG.md` entry under `[Unreleased]`.
- Each new test fails on the parent commit, except tests of new API, which cannot compile there.

### J1. refactor(chunking): move `Chunk` into `pkg/chunking` (D1)

- **Change:**
  - Add `pkg/chunking/chunk.go` with the type, minus the `Vector` field.
  - Remove `Vector: chunk.Vector` from `pkg/export/jsonl.go:62`.
  - Drop the `pkg/export` import from `strategy.go`, `fixed.go`, `headers.go`, `recursive.go` and `semantic.go`.
  - Delete `Chunk` from `pkg/export/types.go`, and make `LLMDocument.Chunks` a `[]chunking.Chunk`.
  - The private `chunkDocument` temporarily returns `[]chunking.Chunk`.
  - Update CLAUDE.md 65 and the `export` and `chunking` lines of the dependency diagram (82-88 at head).
- **Verify:**
  - `go list -deps ./pkg/chunking | /usr/bin/grep -c onedusk/jot/pkg/export` prints 0.
  - `/usr/bin/grep -rn 'chunk.Vector\|Chunk.Vector' pkg` is empty.
  - CLI JSONL is byte-identical to the parent, because nothing set `Chunk.Vector`.
  - The mlpipe gate passes with mlpipe unmodified.
- **CHANGELOG (Changed):** "**`Chunk` moved to `pkg/chunking` (breaking for Go importers)**: `export.Chunk` is now `chunking.Chunk`, and `pkg/chunking` no longer imports `pkg/export`. Its `Vector` field, which nothing set, is removed: vectors belong on `export.ChunkMetadata`, which JSONL export writes. Code that only reads chunk IDs, text, offsets and token counts is unaffected."

### J2. fix(export): JSONL and LLM export chunk with `pkg/chunking`, so chunk IDs are unique (A7, D2)

- **Change:**
  - `jsonl.go:50` and `export.go:83` call `chunking.NewFixedSizeStrategy(tok).Chunk` and return its error.
  - Delete `chunkDocument` and `maxInt`.
  - Move `TestChunkDocument` (`export_test.go:182`) into `TestFixedStrategy`.
  - `ToLLMFormat`'s doc comment says it always uses fixed-size chunking.
- **Verify:**
  - New `TestToJSONL_ChunkIDsUniqueAcrossDocuments` and `TestToLLMFormat_ChunkIDsUniqueAcrossDocuments` fail on the parent (`chunk_id "chunk-0" appears more than once`).
  - `/usr/bin/grep -rn chunkDocument pkg` is empty.
  - CLI JSONL over a fixture matches the parent in every field except `chunk_id`, `prev_chunk_id` and `next_chunk_id`. Compare with jq after deleting those fields.
  - The mlpipe gate passes.
- **CHANGELOG (Fixed):** "**Duplicate JSONL chunk IDs**: `chunk_id` was `chunk-0`, `chunk-1`, ... in every document, so vector databases keyed on it overwrote earlier documents' chunks. IDs are now `<doc_id>-chunk-<n>`, and `prev_chunk_id` and `next_chunk_id` follow. Re-ingest existing indexes. Because IDs are positional, replace all records for a `doc_id` on each export instead of upserting single chunks. LLM-format chunk IDs changed the same way, and the duplicate fixed-size chunker in `pkg/export` is gone."

### J3. feat(export): `ChunkRecords` and `ToJSONLFromChunks`; `ToJSONL` deprecated

- **Change:**
  - Add both functions, with the duplicate `chunk_id` check, which records the first `Source` for each ID.
  - Rewrite `ToJSONL` as a `// Deprecated:` wrapper: fixed strategy, then `ChunkRecords`, then `ToJSONLFromChunks`.
  - Update the `ChunkMetadata.Vector` comment.
- **Verify:**
  - `TestToJSONLFromChunks_WritesRecordsAsGiven`, a golden test: order is preserved, one record's vector is written, and an empty vector is omitted.
  - `TestChunkRecords_Links`: DocID, Source, prev/next links, and an empty Vector.
  - `TestToJSONLFromChunks_RejectsDuplicateChunkIDs`: the error names the `chunk_id` and both sources. It has two cases: the same record twice, and two records with the same ID but different sources.
  - CLI JSONL is byte-identical to J2. Design C's prototype of this step gave 215 rows with the same SHA-1 (reported).
  - The mlpipe gate passes.
- **CHANGELOG (Added):** "**Write caller-built chunks as JSONL**: `export.ChunkRecords(doc, chunks)` turns a strategy's chunks into JSONL records with navigation links, and `(*JSONLExporter).ToJSONLFromChunks(records)` writes records as given, including any `vector` the caller set. It rejects duplicate `chunk_id` values and names both sources. Duplicates come from passing a document twice, or from two scan roots that contain the same relative path (`Scanner.RelativeTo` avoids this). `ToJSONL(documents, ...)` is deprecated and will be removed."

### J4. fix(chunking): recursive chunk positions match their text (N1)

- **Change:** in `recursiveSplit`, track each part's start offset. When nothing is grouped yet, the chunk starts at the current part. After a flush, the next chunk starts at the part that did not fit. This is design C's fix (commit `1e393e1`): 10 lines added, 1 removed.
- **Verify:**
  - `TestRecursiveStrategyPositionsMatchText` checks exact texts and offsets for a leading `"\n\n"`, a run of blank lines, and an oversized paragraph after blank lines. It fails on the parent at chunk 0. The exact-text assertions also catch a fix that adds empty chunks, as design B's did.
  - `TestChunkInvariants` covers fixed, markdown-headers and recursive over the same fixtures at several sizes, checking that:
    - `TokenCount <= maxTokens`;
    - `content[StartPos:EndPos] == Text`;
    - IDs are unique and carry the doc-ID prefix.
    It fails on the parent for recursive only.
  - A scratch comparison, not committed: on the claude-swarm-py markdown at 512/0, mismatches go from 34 to 0, with 414 chunks and identical texts before and after (re-verified).
- **CHANGELOG (Fixed):** "**Wrong `start_pos`/`end_pos` from the recursive strategy**: when a document or a run of paragraphs began with blank lines, the strategy skipped them without advancing its position, so later chunks' offsets pointed at the wrong text (34 of 414 chunks on a 74-document corpus). Chunk text and boundaries are unchanged."

### J5. fix(cli): `export --strategy` selects the chunker for JSONL (A6)

- **Change:**
  - Add `exportJSONL` and use it in the jsonl case.
  - Validate the strategy with `chunking.AvailableStrategies()`, so `contextual` is rejected and `headers` is accepted.
  - Remove `contextual` from the help.
  - Add a `setExportFlags(t, map[string]string)` test helper that restores each flag's `DefValue` in cleanup.
  - Add `chunking` to the `cmd/jot` line of the CLAUDE.md dependency diagram (82 at head).
- **Verify:**
  - `TestE2EExportStrategyChangesJSONL`: two multi-section documents, chunk size 64, overlap 0. The fixed, markdown-headers and recursive outputs must differ pairwise, and every `chunk_id` must be unique. On the parent, the outputs are identical.
  - `TestE2EExportRejectsUnknownStrategy`, a table run through `setExportFlags` and `runExport`: `contextual` fails with the supported list, and `headers` works. The parent accepts `contextual`.
  - `go test -shuffle=on -count=3 ./cmd/jot` passes.
- **CHANGELOG (Fixed):** "**`jot export --strategy` was ignored**: JSONL export always used fixed-size chunking, and every strategy produced a byte-identical file. JSONL now uses the chosen strategy. `headers` is accepted as an alias of `markdown-headers`, as in `pkg/chunking`, and `contextual`, which no strategy implemented, is rejected."

### J6. fix(cli)!: remove `semantic`; `--for-rag` uses recursive, 512 tokens, no overlap

- **Change:**
  - Delete `semantic.go`, the factory case, the doc line and the `AvailableStrategies` entry.
  - The factory error lists `AvailableStrategies()`.
  - Delete `TestSemanticStrategy` and `BenchmarkSemanticStrategy`, and remove `semantic` from the factory benchmark list. The `TestNewChunkStrategy` row for `semantic` now expects an error.
  - In the CLI:
    - change the preset and its stderr line;
    - set `chunkOverlap = 0` for recursive after presets;
    - update the help for `--strategy` and `--chunk-overlap` and the strategy list, saying the overlap is "validated but not applied" by recursive;
    - make the error example use `recursive`.
  - Update the README and CLAUDE.md (50, 71). Edit the workspace CLAUDE.md by hand (Section 5).
- **Verify:**
  - `TestE2EForRAGUsesRecursive`:
    - The fixture has a multi-paragraph document of more than 512 tokens. Below that, fixed and recursive both return the whole document as one identical chunk (`fixed.go:29-38`, `recursive.go:33-43`), and the test would pass on the parent.
    - `--for-rag` stdout must be byte-identical to `--format jsonl --strategy recursive --chunk-size 512 --chunk-overlap 0`.
    - It must also differ from `--strategy fixed --chunk-size 512 --chunk-overlap 128`, so the test cannot silently become vacuous if the fixture shrinks.
    - It fails on the parent.
  - `TestE2EExportRejectsUnknownStrategy` gains a `semantic` row, which fails on the parent.
  - The mlpipe gate passes: mlpipe compiles, and its runtime now rejects `semantic`. Land M2 the same day.
- **CHANGELOG:**
  - **Removed:** "**`semantic` chunking strategy (breaking)**: it was fixed-size chunking under another name. jot does not compute embeddings. Use `fixed`, `markdown-headers` or `recursive`. `chunking.SemanticStrategy` and `NewSemanticStrategy` are removed."
  - **Changed:** "**`--for-rag`** now exports JSONL with the `recursive` strategy, 512-token chunks and no overlap. It was `semantic`, which meant fixed-size chunks with 128 tokens of overlap. The recursive strategy does not overlap chunks; `--chunk-overlap` is still validated but not applied, and the help now says so."

### J7. fix(cli)!: remove `--include-embeddings`

- **Change:**
  - Remove the flag, both warnings, the variable and the help example.
  - Add the "jot does not compute embeddings" help paragraph.
  - Update README 142-143 and 281, `jot.yml:44` and FR-018.
- **Verify:**
  - Create `TestE2EExportRejectsRemovedOptions`.
    - It runs `rootCmd.SetArgs([]string{"export", "--include-embeddings"})` and `rootCmd.Execute()`, with the same `Out`/`Err`/`Args`/`SilenceUsage` cleanup as `TestE2EErrorOutput` (`e2e_test.go:327-374`).
    - It asserts that the output contains `unknown flag: --include-embeddings`.
    - The assertion must check that text, not just that an error occurred. The parent also returns an error in an empty fixture ("no markdown files found"), so a bare `err != nil` check would pass on the parent.
  - `export --help` contains "does not compute embeddings" and no `include-embeddings`, `semantic` or `contextual`.
- **CHANGELOG (Removed):** "**`--include-embeddings` (breaking)**: it printed an 'API costs apply' warning and generated nothing. jot does not compute embeddings. Embed each JSONL row's `text` and store it as `vector`, or build records in Go with `export.ChunkRecords`, set `Vector`, and write them with `ToJSONLFromChunks`, as mlpipe does."

### J8. refactor(export)!: delete the stubs tied to #2

- **Change:**
  - Remove `contextualEnrichment` and its call.
  - Remove the `separateFiles` parameter of `ToEnrichedMarkdown`, along with `TestToEnrichedMarkdown_SeparateFilesNotImplemented` and `TestContextualEnrichment`.
  - Remove the unreachable `llm` case (`export.go:292-301`) and the `encoding/json` import.
- **Verify:**
  - Markdown export is byte-identical to J7 on a fixture.
  - `/usr/bin/grep -rn 'contextualEnrichment\|separateFiles' pkg cmd` is empty.
  - The mlpipe gate passes.
- **CHANGELOG (Removed):** "`MarkdownExporter.ToEnrichedMarkdown` no longer takes `separateFiles`, which only returned 'not yet implemented' (breaking for Go importers). The empty contextual-enrichment stub and the unreachable `llm` branch of `jot export` are gone."

### J9. fix(cli)!: remove `--for-context`; chunking flags apply to JSONL only (open question 1)

- **Change:**
  - Remove the preset, its variables and its stderr line.
  - Update the mutual-exclusion error text (`export.go:98`).
  - Remove strategy and size from the markdown progress line.
  - Add the "jsonl only" help wording.
  - Update README 18, 135, 140, 305 and 375, and CLAUDE.md 72. Edit the workspace CLAUDE.md by hand (Section 5).
- **Verify:**
  - `TestE2EExportRejectsRemovedOptions` gains a `--for-context` case that asserts `unknown flag: --for-context`. It fails on the parent.
  - Help contains "jsonl only".
  - `--format markdown` output is byte-identical to J8.
- **CHANGELOG (Removed):** "**`--for-context` (breaking)**: markdown export writes whole documents and ignored `--strategy` and `--chunk-size`, so this preset produced the same file as `--format markdown`. Use `--format markdown`. The help now states that chunking options apply to JSONL only."
- **If open question 1 is answered "chunk":** J9 changes `ToEnrichedMarkdown(documents []scanner.Document, records []ChunkMetadata)` to write one frontmatter block per record, and keeps `--for-context`.
  - `section` comes from design A's `sectionAt(doc, StartPos)`, which counts lines from 0 to match `Section.StartLine`.
  - `--for-context` should use recursive at 1024 with overlap 0.
  - Test: a multi-section document under markdown-headers yields one block per section.

### J10. refactor(export)!: remove the deprecated `ToJSONL` (after mlpipe M1)

- **Change:** port jot's JSONL tests, including J2's `TestToJSONL_ChunkIDsUniqueAcrossDocuments`, to a `toJSONL(t, docs, max, overlap)` test helper built on the fixed strategy, `ChunkRecords` and `ToJSONLFromChunks`.
- **Verify:**
  - The mlpipe gate passes against migrated mlpipe.
  - Against pre-M1 mlpipe it fails with `exporter.ToJSONL undefined` (re-verified on C's prototype). That failure is why M1 must come first.
- **CHANGELOG (Removed):** "`(*JSONLExporter).ToJSONL(documents, ...)` (breaking for Go importers). Use `ChunkRecords` and `ToJSONLFromChunks`."

### J11. The LLM format (after mlpipe M4; open question 2)

- **Drop (recommended):** refactor(export)!: delete the following, about 190 non-test lines counted at `90b5189`:
  - `ToLLMFormat`, `LLMExport`, `LLMDocument`, `LLMSection`, `LLMCodeBlock`, `Links` and `SemanticIndex`
  - `indexDocument`, `extractKeywords` and `contains`
  - `TestExporter_ToLLMFormat` and J2's `TestToLLMFormat_ChunkIDsUniqueAcrossDocuments`
  - **Verify:** the mlpipe gate passes against mlpipe after M4 (Drop).
  - **CHANGELOG (Removed):** "The LLM export structure (`Exporter.ToLLMFormat` and its types), which no jot command could produce."
- **Keep:** feat(export): add `ToLLMFormatFromChunks`, change `LLMDocument.Chunks` to `[]ChunkMetadata`, and deprecate `ToLLMFormat`, which is removed after mlpipe switches.
  - When `ToLLMFormat` is removed, `TestToLLMFormat_ChunkIDsUniqueAcrossDocuments` moves onto `ToLLMFormatFromChunks` and its duplicate check.
  - **Verify:** records from markdown-headers with vectors produce 3 chunks, each with a vector, in `LLMDocument.Chunks`, where `ToLLMFormat` gives 1.

### J12. docs: record Decision 2 in the audit

Update Section 11 of `docs/internal/improvement-opportunities.md`:
- mark A6, A7, the A16 rows, D1 and D2 as done;
- add N1, and the new findings N2 and N3 from Section 10.

---

## 9. Test plan

| Layer | Test | Step | Fails before because |
|---|---|---|---|
| chunking | `TestFixedStrategy` gains `TestChunkDocument`'s assertions | J2 | (moved test) |
| chunking | `TestRecursiveStrategyPositionsMatchText`: exact texts and offsets | J4 | Offsets drift at chunk 0 |
| chunking | `TestChunkInvariants`: token limit, `content[Start:End] == Text`, unique prefixed IDs; fixed, markdown-headers, recursive | J4 | recursive offsets |
| chunking | `TestNewChunkStrategy`: `semantic` expects an error, and the error lists `AvailableStrategies()` | J6 | `semantic` is accepted |
| export | `TestToJSONL_ChunkIDsUniqueAcrossDocuments`, `TestToLLMFormat_ChunkIDsUniqueAcrossDocuments` | J2 | `chunk-0` repeats |
| export | `TestChunkRecords_Links`, `TestToJSONLFromChunks_WritesRecordsAsGiven`, `TestToJSONLFromChunks_RejectsDuplicateChunkIDs` (names both sources) | J3 | New API |
| cmd/jot | `TestE2EExportStrategyChangesJSONL` | J5 | Identical output for every strategy |
| cmd/jot | `TestE2EExportRejectsUnknownStrategy`: `contextual` rejected and `headers` accepted (J5); a `semantic` row (J6) | J5, J6 | `contextual` accepted and `headers` rejected; `semantic` accepted |
| cmd/jot | `TestE2EForRAGUsesRecursive`, with a fixture document of more than 512 tokens; also asserts the output differs from fixed 512/128 | J6 | The preset ran `semantic` (fixed, 512/128) |
| cmd/jot | `TestE2EExportRejectsRemovedOptions`, created in J7 (`--include-embeddings`) and extended in J9 (`--for-context`); asserts Cobra's `unknown flag` text through `rootCmd.SetArgs` and `Execute` | J7, J9 | Each flag was accepted, and the parent's only error in an empty fixture is "no markdown files found" |
| mlpipe | `TestProcessMarkdown_VectorsAndStrategyReachJSONL` | M1 | 3 chunks, 1 JSONL row, 0 vectors |
| mlpipe | `ProcessMixed`: `a/README.md` and `b/README.md` return an error naming both absolute paths before chunking; the same directory passed twice yields each document once | M3 | Silent duplicate IDs, and a repeated directory is chunked twice |
| cross-repo | mlpipe gate (corrected) after every jot commit. For J10 and J11, also show that it fails against unmigrated mlpipe | all | - |

**Notes:**
- The tests are hermetic, because #4 embedded the tokenizer.
- `cmd/jot` tests share process globals, including `exportCmd` and its flags (V9).
  - Every new e2e test that sets export flags directly does so through `setExportFlags`, and the suite runs with `-shuffle=on`.
  - Unknown-flag cases go through `rootCmd.SetArgs` and `Execute` instead. Cobra fails while parsing flags, so no flag value changes.
- pflag's `Set` also marks a flag `Changed`, and no cleanup resets that. This design never reads `Changed`, so stale `Changed` bits do not affect it. (V9 does not mention `Changed`; this is a separate observation.)
- **Measurements, run in scratch and not committed:**
  - recursive mismatches on claude-swarm-py at 512/0 (34 to 0, same 414 texts);
  - byte-identical CLI JSONL from J2 to J3;
  - byte-identical markdown from J7 to J9.

---

## 10. Risks

- **Re-keying.** Every JSONL `chunk_id` changes once (J2), and existing vector indexes must be rebuilt. The CHANGELOG entry says so and gives the replace-per-`doc_id` procedure.
- **`--for-rag` loses overlap.** Overlap helps when an answer straddles a boundary. Recursive's paragraph-aligned boundaries reduce the need, and users who want overlap can pass `--strategy fixed --chunk-overlap 128`.
- **The duplicate check turns `ProcessMixed` collisions into errors** from J3 onward. This is intended, but until M3:
  - an mlpipe user sees a jot-worded error;
  - after M1 the error arrives only after embedding.
  
  M3 moves the check before chunking and names both absolute paths. Other library callers that embed should check their inputs the same way.
- **Removed options break scripts:** `--include-embeddings`, `--for-context`, `semantic` and `contextual`. Each fails loudly, with Cobra's usage hint or the supported-strategies list.
- **mlpipe configs that name `semantic`** fail validation after J6 until M2 updates mlpipe's docs. The error lists the valid strategies.
- **Out of scope and unchanged** (keep the implementation from drifting into these):
  - C1: chunking is superlinear; recursive at 512 over 74 documents took about 1 s (reported).
  - A13: fixed and recursive can split a UTF-8 sequence.
  - A10: heading detection is not fence-aware, which affects markdown-headers.
  - `chunking.DefaultStrategy` (`factory.go:49-52`) has no callers. Leave it for a #5 follow-up.
  - The default overlap of 128 makes `--chunk-size` 128 or less fail validation for every strategy unless `--chunk-overlap` is lowered. That includes `recursive`, which does not apply the overlap; the help says the value is validated. Validation also runs before presets are applied. Both predate this design.
- **New findings, re-verified and not fixed here:**
  - **N2:** markdown-headers emits an empty chunk when a document has a blank line before its first heading (2 of 74 claude-swarm-py documents at 512).
  - **N3:** an empty document yields one row with empty `text` in every strategy (4 of 74). Some embedding APIs reject empty inputs.

---

## 11. Open questions for the maintainer

1. **Markdown export: whole documents (recommended), or one frontmatter block per chunk?**
   - **Recommended default:** keep whole documents, remove `--for-context`, and document the chunking flags as JSONL-only (J9).
     - This departs from the audit's A6 fix, which routes markdown through the strategy as well. The audit's A16 fix allows either option.
   - **Measured cost of chunking:**
     - jot's docs/ markdown export grows from 107 KB in 9 blocks to 149 KB in 69 blocks at the default fixed 512/128 (judge-measured on A's tree), because overlap text is written twice.
     - `--for-context` with markdown-headers at 1024/256 gives 268 blocks and 158 KB (reported by A).
     - Frontmatter costs about 70 tokens per block, against markdown-headers chunks averaging about 91 tokens (reported by C).
   - Whole-document context is already what `llms-full` provides.
   - **If you choose chunking:** J9 takes the alternative described there, and `--for-context` stays, using recursive at 1024 with no overlap.
2. **The LLM export format: drop it from both repositories (recommended), or keep and fix it?**
   - **Why drop:**
     - jot's CLI cannot produce it (the branch is unreachable, and J8 deletes it).
     - mlpipe never writes it to disk (Bug 2).
     - It repeats `Result.Chunks`.
   - **Recommended default, drop:** mlpipe removes its `llm` output format (M4, with the full list of sites), then jot deletes about 190 lines (J11).
   - **Keep:**
     - jot adds `ToLLMFormatFromChunks(documents, records)`, which honors the strategy and carries vectors.
     - `LLMDocument.Chunks` becomes `[]ChunkMetadata`.
     - mlpipe switches to the new function and fixes Bug 2 with `!= nil`.
     - `ToLLMFormat` goes through the same deprecate-then-delete path as `ToJSONL`.

---

## Revision notes

For each critic issue, it was checked against the code and the document was changed where the critic is right.

1. **The mlpipe gate hides test failures (high): accepted.**
   - **Confirmed:** `mlpipe-gate.sh:7-8` records grep's status. The critic's `t.Fatal` module fails under `go test` (rc 1) and passes through the gate pipeline (rc 0).
   - **Re-run with a corrected gate** (`design-rev2/gate.sh`, with each status recorded separately). Every cited result holds:
     - the judge's no-alias D1 tree passes build, vet and test with unmodified mlpipe;
     - C's steps 1-7 pass;
     - C's `ToJSONL` deletion fails with `exporter.ToJSONL undefined`;
     - that commit plus C's M1 passes;
     - head `1a7a6c7` passes.
   - **Changed:**
     - Section 8's standing check now spells out the corrected gate and warns against the shared script.
     - Provenance for the gate claims in Sections 3 and 6 changed from judge-verified to re-verified.
   - The shared script was not edited, because other agents use it.
2. **The M4 (Drop) list is incomplete (medium): accepted, and extended.**
   - Confirmed: `client_test.go:57` uses `WithOutputFormat("llm")`, and J11 also needs `types.go:16` and `client.go:332-338` removed in order to compile.
   - M4 now lists every site found by grep: those three, plus `client.go:62-64`, `config.go:83-85`, `options.go:32`, `process.go:23,44`, README 10/55/123/144/169, `mlpipe.yml:11` and mlpipe `CLAUDE.md:62`.
   - The list is labeled as from reading, not gated against a J11 tree.
3. **`ChunkRecords` drops `Chunk.Vector` (medium): accepted.** I chose to delete `Chunk.Vector` in J1 rather than copy it.
   - **Why:** nothing sets the field, one slot is simpler, and J11 shrinks.
   - **Prototype:** on the judge's D1 tree (pre-#4), with the field and `jsonl.go:62` removed, jot's tests pass and the mlpipe gate passes.
   - **Updated to match:** Decisions 1 and 3, a new rationale paragraph, a rejected alternative ("copy it"), Section 2 row 14, the Section 4 `Chunk` block, the J1 change, verify and CHANGELOG, Section 7, and J11 (Drop no longer lists `Chunk.Vector`).
   - Under the "keep" answer to open question 2, `LLMDocument.Chunks` becomes `[]ChunkMetadata`. The shape change is stated.
4. **The duplicate-ID error and M3 assume two distinct files (medium): accepted.**
   - M3 now first skips repeated `Path`s, which covers:
     - a directory listed twice;
     - nested roots (the critic said nothing catches these, but `Path` is absolute, so this check does);
     - a PDF output directory inside a markdown directory.
   - It then rejects distinct files that share a relative path, and runs before chunking, so collisions surface before embeddings are paid for.
   - jot's check now remembers the first `Source` per ID and names both.
   - The false claim that both rows must share a `Source` is gone.
   - The mlpipe current-state row and Section 10 describe the three failure modes.
5. **J11 would break the export tests (low): accepted.** J11 (Drop) also deletes `TestToLLMFormat_ChunkIDsUniqueAcrossDocuments`. Under Keep, the test moves onto `ToLLMFormatFromChunks`. J10 now names J2's `TestToJSONL_ChunkIDsUniqueAcrossDocuments` among the ported tests.
6. **`TestE2EForRAGUsesRecursive` can pass on the parent (low): accepted.** Confirmed from the identical single-chunk early returns in `fixed.go:29-38` and `recursive.go:33-43`. J6 now requires a fixture document of more than 512 tokens and an assertion that the output differs from fixed 512/128.
7. **The test plan cites a nonexistent test and misattributes V9 (low): accepted, with a simpler structure.**
   - The `semantic` case is a new row in J5's `TestE2EExportRejectsUnknownStrategy`, not a new test.
   - J7 creates `TestE2EExportRejectsRemovedOptions`, using `rootCmd.SetArgs` and `Execute` with `TestE2EErrorOutput`'s cleanup. J9 extends it.
   - Beyond the critic: the unknown-flag assertions must check the `unknown flag` text, because the parent also errors in an empty fixture ("no markdown files found").
   - The `Changed`-bits observation is now separate from V9, which covers only process globals.
8. **The help says recursive ignores `--chunk-overlap`, but the value is still validated (low): accepted.**
   - Confirmed: validation (`export.go:115-118`, called at `:160`) runs before presets and before the recursive override.
   - Fixed with wording, not code: the help and CHANGELOG now say "validated but not applied". Skipping the check was added as a rejected alternative.
   - The default-overlap risk bullet now names `recursive`.
9. **The workspace CLAUDE.md is stale (low): substance accepted, one claim rejected.**
   - `/Users/macadelic/dusk-indust/shared/packages/CLAUDE.md:99-100` does describe `semantic`, the semantic `--for-rag` preset and `--for-context`. It is now listed as a manual edit at J6 and J9.
   - Rejected: the claim that the file is tracked. `git status` shows it as untracked (`??`), in a `dusk-indust` repository with no commits.

**Change not raised by the critic:**
- **What happened:** the branch has moved from `90b5189` to `1a7a6c7`.
- **Changed:**
  - The base is restated as `1a7a6c7`, with the three later commits that matter: `86a60e5` (removed `minInt` and `LLMDocument.HTML`), `b1b524c` (hardened `dedupeDocuments`) and `5f19e30` (document IDs hash the forward-slash path).
  - J2 and Section 4 no longer delete the already-deleted `minInt`.
  - Line citations marked "at head" were re-mapped: README, CLAUDE.md, `scanner.go:147,190`, `build.go:201`, `export.go:376-382`, and `types.go` 36-44.
  - The `cli-dev.md` citation is pinned to `2acde71`, because the working tree deletes `.claude/`.
  - J5 now adds `chunking` back to the `cmd/jot` line of the CLAUDE.md dependency diagram. Head dropped it because `cmd/jot` does not import `chunking` today, and J5 makes it do so.
- Measurements keep the commit they were taken at and were not re-run.

# Jot: Improvement Opportunities and Remediation Plan

| | |
|---|---|
| Date | 2026-09-18 |
| Commit reviewed | `88fe938` (main, clean tree) |
| Scope | Entire repository: `cmd/`, `internal/`, `pkg/`, `web/`, build tooling, scripts, docs |
| Code size | ~4,500 lines of Go (non-test), ~4,300 lines of tests, ~1,750 lines of front-end assets |
| Changes made | None at the time of the audit. Phase 0 and Phase 1 have since been implemented (branch `fix/remediation-phase-1`, merged to main); see [Section 11](#11-progress-and-verification). The Section 9 decisions are recorded in [Section 12](#12-section-9-decisions-2026-09-24) (branch `feat/section9-decisions`). |

## How to read this document

Every finding has an ID (for example `A1`), a priority, a status, and file references in
`path:line` form.

**Priority**

- **P0** - A user following the documented happy path gets broken or wrong output.
- **P1** - An advertised feature does not work, or there is a significant performance or maintainability cost.
- **P2** - Polish, cleanup, or low-frequency edge cases.

**Status**

- **Reproduced** - Confirmed by building the binary at `88fe938` and running it against a throwaway fixture outside the repository. Commands are in [Appendix A](#appendix-a-reproduction-commands).
- **Code-evident** - Follows directly from reading the code; not separately executed.

---

## 1. Executive summary

Jot's core pipeline is small, readable, and the packages that have tests are well covered
(renderer 94%, search 98%, toc 98%, export 93%). The benchmark on record (2,679 files in
10.81s) shows the basic design holds up.

The main problem is a gap between what the tool advertises and what it does. The gap is
concentrated in two places: the path a new user takes on day one, and the LLM export
feature set that is the project's differentiator.

The ten items that matter most:

1. **Sites built by an installed binary have no CSS or JavaScript** (`A1`). Assets are read from `web/templates/assets/` relative to the current directory instead of being embedded. `go install` is the recommended install method, so this affects nearly every real user. The build also prints "Created assets/styles.css" regardless.
2. **With default settings, every rebuild ingests its own output** (`A2`). Source markdown is copied into `dist/`, and `dist/` is inside the default input path. Observed: 4, then 8, then 12 documents across three consecutive builds.
3. **The config `jot init` generates causes silent page overwrites** (`A3`). Documents are keyed by path relative to each input root, so `docs/README.md` and `README.md` both write `dist/README.html`. The build reports two files generated; one exists.
4. **No `index.html` is produced when a root `README.md` exists** (`A4`), which breaks static hosting and the header logo link.
5. **`--strategy` has no effect** (`A6`). `pkg/chunking` is never imported by the CLI or by `pkg/export`. All four strategies produce byte-identical JSONL. This also makes `--for-rag` ("semantic") and `--for-context` ("headers") misleading.
6. **JSONL `chunk_id` values collide across documents** (`A7`): 40 rows, 25 distinct IDs in the test fixture. This corrupts upserts into any vector store keyed on chunk ID.
7. **Frontmatter is stripped but never parsed** (`A8`), and **`.jotignore` is generated but never read** (`A9`). Both are documented features.
8. **Chunking cost grows roughly quadratically with document size** (`C1`): 0.51s for a 100 KB document, 5.11s for 400 KB.
9. **Configuration fails silently** (`B3`): a malformed `jot.yml` is ignored, `JOT_*` environment overrides do not work for any real key, and only 7 of the ~25 keys in the generated config are read.
10. **There is no CI** (`F3`), the package containing items 1-4 has 0% test coverage (`E1`), and `make build` produces a binary with an empty version string (`F1`).

Almost all of these are small fixes. The recommended order is in [Section 10](#10-remediation-plan);
the first step is a test harness that would have caught items 1-6.

---

## 2. Current state snapshot

### Test coverage (`go test -cover ./...`, all passing)

| Package | Coverage | Notes |
|---|---|---|
| `cmd/jot` | 30.7% | Only `runBuild` and `humanizeBytes` are exercised. `runExport`, `validateExportFlags`, `runTOC`, `runInit`, `runServe` are 0%. |
| `internal/compiler` | **0.0%** | No test file. Contains findings A1-A4. |
| `internal/renderer` | 94.1% | |
| `internal/search` | 97.8% | |
| `internal/toc` | 98.4% | |
| `pkg/chunking` | 61.5% | `recursiveSplit` (the whole recursive algorithm) is 0%. |
| `pkg/export` | 93.3% | |
| `pkg/scanner` | 79.6% | `ExtractFrontmatter` 28.6%. |
| `pkg/tokenizer` | 0.0% | No test file; exercised indirectly. |

### Tooling

| Check | Result |
|---|---|
| `go build ./...` | OK |
| `go vet ./...` | Clean |
| `gofmt -l .` | 12 files not formatted |
| `go mod tidy -diff` | Non-empty (`tiktoken-go` is imported directly but listed as indirect) |
| `golangci-lint` | No config in repo; `make lint` fails unless the tool happens to be installed |
| CI | None. No `.github/` directory, although `README.md:3` shows a CI badge. |

---

## 3. Correctness and unfulfilled features

### A1. Assets are not embedded in the binary - P0, Reproduced

- **Where:** `internal/compiler/compiler.go:181-224`, `cmd/jot/build.go:114`
- **What:** `copyAssets` reads `web/templates/assets/{style.css,search.js,highlight.js,syntax-highlighting.css}` relative to the process working directory. If a file is missing the error is discarded. Running `jot build` anywhere other than the jot repository root yields `dist/assets/` containing only the search index. Every page is unstyled and search, syntax highlighting, and the copy menu are dead.
- **Aggravating factors:** `docs/INSTALL.md` recommends `go install ...@latest`, which can never work. `build.go:114` prints `Created assets/styles.css` unconditionally, and the real filename is `style.css`.
- **Fix:** Use `go:embed`. Because embed paths cannot reach upward, add a tiny `web/embed.go` (`package web`, `//go:embed templates/assets/*`) and have the compiler iterate the `embed.FS`. This also replaces four copy-pasted blocks with one loop. Fail the build if an asset cannot be written.

### A2. Rebuilds ingest their own output - P0, Reproduced

- **Where:** `internal/compiler/compiler.go:92-98`, `cmd/jot/build.go:216-221`, `pkg/scanner/scanner.go:51-86`
- **What:** Since `88fe938`, each source `.md` is copied into the output directory for the "Open Markdown" link. Defaults are input `.` and output `./dist`, and the scanner does not exclude the output directory. Each build therefore rescans the previous build's copies. Observed growth: 4 -> 8 -> 12 documents, with `dist/dist/dist/...` nesting.
- **Note:** This repo's own `jot.yml` hides the problem with `clean: true` and a `**/dist/**` ignore. Users without a config, or with `clean: false`, hit it.
- **Fix:** Resolve the output directory to an absolute path and have the scanner skip it unconditionally (`filepath.SkipDir`). Do not rely on user ignore patterns for this.

### A3. Output paths collide across input roots - P0, Reproduced

- **Where:** `pkg/scanner/scanner.go:67`, `cmd/jot/build.go:57-78`, `internal/compiler/compiler.go:79,93`
- **What:** `RelativePath` is relative to each scanner root, and results from all roots are concatenated. With `input.paths: ["docs", "README.md"]` (exactly what `jot init` writes), `docs/README.md` and `README.md` both map to `dist/README.html`. The last one wins silently, and the build still reports "Generated 2 HTML files". The same collision applies to the search index, TOC node IDs, and llms.txt links.
- **Fix:** Minimum: detect duplicate `RelativePath` values after scanning and fail with both source paths named. Better: make paths relative to the project root (the config file's directory) so roots cannot collide. The second option changes generated URLs (`real.html` becomes `docs/real.html`); see [Decision 1](#9-decisions-needed-from-the-maintainer).

### A4. No `index.html` when the root document is `README.md` - P0, Reproduced

- **Where:** `internal/compiler/compiler.go:111-118`, `internal/renderer/template.go:33`, `cmd/jot/serve.go:54-82`
- **What:** `hasIndexPage` treats `README.md` as the index, so no index is synthesized, but the page is written as `README.html`. Static hosts (GitHub Pages, S3, Netlify) serve a 404 at `/`, and the header logo links to a missing `index.html`. `jot serve` hides this by special-casing `/`.
- **Fix:** When there is no `index.md`, emit the root `README.md` as `index.html` (or write both). Then remove the special case from `serve`.

### A5. A document and a directory with the same title merge in the TOC - P1, Reproduced

- **Where:** `internal/toc/builder.go:84`, `internal/renderer/renderer.go:289-309`
- **What:** Directory lookup uses `FindChildByTitle(humanizeTitle(part))`, which matches any sibling, including leaf documents. With `guides.md` (title "Guides") next to `guides/`, the documents under `guides/` become children of the leaf. The nav renderer treats any node with a `Path` as a leaf and ignores its children, so `guides/intro.md` vanishes from the sidebar. The page is still generated; it is simply unreachable by navigation. Separately, `my-dir/` and `my_dir/` humanize to the same title and merge.
- **Fix:** Match directory nodes by path segment (or node ID) and only among non-leaf nodes. `flattenTOC`'s dedupe comment (`renderer.go:209`) suggests this case was noticed but worked around rather than fixed.

### A6. `--strategy` is validated, echoed, and ignored - P0, Reproduced

- **Where:** `cmd/jot/export.go:166,273,281`; `pkg/export/jsonl.go:50`; `pkg/export/export.go:221-331`; `pkg/chunking/strategy.go:5`
- **What:** No non-test file imports `pkg/chunking`. JSONL export always calls the private `export.chunkDocument`, which is a copy of `chunking.FixedSizeStrategy.Chunk`. SHA-1 of the output is identical for `fixed`, `semantic`, `markdown-headers`, and `recursive`. The CLI also accepts `contextual` (`export.go:135`), which the factory does not know (`pkg/chunking/factory.go:24-35`), so wiring the flag naively would make a documented value fail.
- **Root cause:** `chunking` imports `export` for the `Chunk` type, so `export` cannot import `chunking` without a cycle. See `D1`.
- **Fix:** Invert the dependency (`D1`), route JSONL and markdown export through `chunking.NewChunkStrategy`, delete `export.chunkDocument`, and either add or drop the `contextual` alias.

### A7. JSONL `chunk_id` is not unique - P0, Reproduced

- **Where:** `pkg/export/export.go:228,281`, `pkg/export/jsonl.go:56`
- **What:** IDs are `chunk-0`, `chunk-1`, ... per document. Across a corpus they repeat (40 rows, 25 distinct; `chunk-0` appeared 5 times). The README positions this output for Pinecone, Weaviate, and Qdrant, where the ID is the primary key, so later documents overwrite earlier ones. `prev_chunk_id` and `next_chunk_id` are equally ambiguous.
- **Fix:** The strategies in `pkg/chunking` already emit `<docID>-chunk-<n>` (`fixed.go:32,85`), so resolving `A6` resolves this. Add a uniqueness assertion to the JSONL tests.

### A8. Frontmatter is removed but never parsed - P1, Reproduced

- **Where:** `pkg/scanner/document.go:193-208`
- **What:** `ExtractFrontmatter` strips the block and returns an empty map (there is a TODO). A document with `title: Frontmatter Title` rendered with its H1 instead. Dependent code is dead: title lookup (`document.go:57-61`) and tag extraction (`internal/toc/builder.go:171-181`, `internal/search/indexer.go:102-112`). The tag code would not work even with parsing, because `yaml.v3` yields `[]interface{}`, not `[]string`. Only `---\n` is recognized, so CRLF files keep their frontmatter as body text.
- **Fix:** `yaml.Unmarshal` the block (`yaml.v3` is already a dependency), accept `\r\n`, handle `[]interface{}` tags in one shared helper, and surface YAML errors with the file path.

### A9. `.jotignore` is generated and documented but never read - P1, Code-evident

- **Where:** `pkg/scanner/ignore.go:144-148` (stub with no callers), `cmd/jot/init.go:84-124`, `README.md:90`
- **What:** `LoadIgnoreFile` returns an empty slice and nothing calls it. The generated file also uses negation (`!.jotignore`) and anchored patterns (`/dist/`), neither of which the matcher supports. The matcher itself (`ignore.go:98-117`) is an approximation: `prefix*` matches against the whole path, `*x*` is a substring test, and `matchInSubpath` is O(segments squared) per pattern per file.
- **Fix:** Either remove `.jotignore` from `init` and the README, or implement it on top of a maintained gitignore-compatible matcher rather than extending the hand-rolled one.

### A10. Markdown parsing is not fence-aware - P1, Reproduced

- **Where:** `pkg/scanner/document.go:55-76` (title), `:80-123` (sections), `:161` (code blocks); `internal/search/indexer.go:130-143`; `pkg/chunking/headers.go:44-63`
- **What:** Any line starting with `#` is a heading, including shell comments inside fenced code. In the fixture, `# this is a shell comment` appeared in the search index headings, and the same logic drives section extraction, header chunking, and title detection (a document whose first `#` line is inside a code block gets that comment as its title). The fence regex `^```(\w*)$` rejects `c++`, `objective-c`, info strings with attributes, indented fences, and `~~~`. When an opening fence is not recognized, the closing fence is read as an opener and in/out state is inverted for the rest of the file.
- **Fix:** One fence-aware line iterator in `pkg/scanner`, used by all four call sites. Longer term, derive headings from the Blackfriday AST rather than from regexes.

### A11. Local timestamps are labeled as UTC - P1, Reproduced

- **Where:** `internal/toc/toc.go:86`, `internal/search/indexer.go:122`
- **What:** `Format("2006-01-02T15:04:05Z")` writes a literal `Z` without converting. A file modified at 19:37 EDT was emitted as `19:37:47Z`.
- **Fix:** `.UTC().Format(time.RFC3339)`.

### A12. Output is not deterministic - P1, Code-evident

- **Where:** `internal/toc/builder.go:218-223`, `internal/search/indexer.go:169-174`, `pkg/export/export.go:358-361`, `cmd/jot/toc.go:126`
- **What:** Keywords come from ranging over a map, so order varies between runs. In the TOC builder, the 10-keyword cap is applied during that iteration, so which keywords survive is also random, and they are not the most frequent ones. `jot toc` writes `toc.xml` into source directories and `toc.json` at the project root (described as "like a lock file"), so these files churn in version control on every run.
- **Fix:** Sort by count descending, then word, then truncate. Sort `tocPaths`. Add a build-twice-and-compare test (`E6`).

### A13. Byte-offset slicing can split UTF-8 sequences - P2, Code-evident

- **Where:** Summaries at `internal/toc/builder.go:158-160`, `internal/search/indexer.go:97-99`, `pkg/export/llmstxt.go:142-144`. Chunk boundaries at `pkg/chunking/fixed.go:53-78`, `pkg/chunking/recursive.go:76-90`, `pkg/export/export.go:249-274`.
- **What:** `s[:200]` and binary search over byte positions can cut a multi-byte rune. The word-boundary backoff only looks back 100 bytes for an ASCII space or newline, so CJK text and long URLs are not protected. `encoding/json` replaces the invalid bytes with U+FFFD, corrupting chunk text.
- **Fix:** Rune-aware truncation helper; token-window chunking (`C1`) removes byte-position search entirely.

### A14. Titles are interpolated into HTML and XML without escaping - P1, Code-evident

- **Where:** `internal/renderer/renderer.go:304-305,313,324-325` (then cast to `template.HTML` at `:178`); `internal/toc/toc.go:78-80`
- **What:** A title such as `Using <T> generics` or `Q&A` produces malformed navigation markup on every page. TOC XML escapes element text but not the `id` and `path` attributes. `MarshalXML` (`toc.go:198-201`) encodes the entire pre-rendered XML string as escaped text, which is never what a caller wants.
- **Fix:** `html.EscapeString` at minimum; preferably render navigation through `html/template`. Build TOC XML with `encoding/xml` structs and delete the hand-rolled writer and `escapeXML`.

### A15. `jot toc --recursive` does nothing - P1, Reproduced

- **Where:** `cmd/jot/toc.go:113-116,129-139`
- **What:** Documents are grouped by their exact parent directory, so the non-recursive filter (`filepath.Dir(doc.Path) == dirPath`) is always true and the recursive branch receives the same list. Output with and without `-r` is identical.
- **Fix:** For recursive mode, add each document to every ancestor directory's list up to the input root.

### A16. Flags and code paths that are accepted but inert - P1, Reproduced / Code-evident

| Item | Where | Behavior |
|---|---|---|
| `--include-embeddings` | `cmd/jot/export.go:71,148-151,230-234` | Prints an "API costs apply" warning and generates nothing. No `vector` field in output. (Reproduced) |
| `semantic` strategy | `pkg/chunking/semantic.go:44-48` | Delegates to fixed. README:265 describes it as "embedding-based boundary detection". |
| `--format markdown` chunk options | `cmd/jot/export.go:276-282`, `pkg/export/markdown.go:45` | Prints strategy and chunk size, then ignores both. `--for-context` is therefore identical to `--format markdown`. |
| `recursive` overlap | `pkg/chunking/recursive.go:29-51` | `overlapTokens` is accepted and never used. |
| Legacy `llm` format | `cmd/jot/export.go:284-293` vs `:121` | Unreachable: validation rejects `llm` first. (Reproduced) |
| `separateFiles` | `pkg/export/markdown.go:45-49` | Returns "not yet implemented". |

**Fix:** Remove what is not going to ship soon, and have the help text say so for what remains. An honest, smaller flag surface is better than a larger one that silently degrades.

---

## 4. Ease of use

### B1. Progress output corrupts piped exports - P0, Reproduced

- **Where:** `cmd/jot/export.go:181-199,227,245-276,317`
- **What:** Status lines and the payload both go to stdout. `jot export --format json | jq .` fails because the stream begins with "Scanning for markdown files...". Stdout is the default destination (`--output` is optional), so this is the primary usage path.
- **Fix:** Send all progress to stderr. Consider `--quiet`.

### B2. Errors are printed twice, with a usage dump - P2, Reproduced

- **Where:** `cmd/jot/root.go:20-33`, `cmd/jot/main.go:15`
- **What:** Cobra prints the error and full usage, then `main` prints the error again. Runtime failures (for example "no markdown files found") show flag help, which buries the actual message.
- **Fix:** Set `SilenceErrors: true` and `SilenceUsage: true` on the root command.

### B3. Configuration fails silently and is mostly decorative - P1, Reproduced

- **Malformed config is ignored.** `root.go:71-75` discards every `ReadInConfig` error, not just "not found". A YAML syntax error means defaults are used and the whole working directory is scanned with no warning.
- **Environment overrides do not work.** `AutomaticEnv` is set without `SetEnvKeyReplacer`, so `output.path` maps to the impossible variable `JOT_OUTPUT.PATH`. `JOT_OUTPUT_PATH=... jot build` had no effect. `CLAUDE.md` and the README document `JOT_*` as supported.
- **Only seven keys are read:** `input.paths`, `input.ignore`, `output.path`, `output.clean`, `project.name`, `project.description`, `features.llm_export`. Everything else that `jot init` writes (`version`, `author`, `output.format`, `output.theme`, `features.search|versioning|syntax_highlighting|auto_toc`, `server.*`, `llm.*`, `search.*`) has no effect. `server.port` and `llm.chunk_size` are the most surprising omissions since matching flags exist.
- **Naming mismatch.** `--config` help says the default is `./jot.yaml` (`root.go:40`); `init` writes `jot.yml`. Both load, but the mismatch confuses.
- **Fix:** Define a typed `Config` struct, load with `viper.Unmarshal`, distinguish `ConfigFileNotFoundError` from parse errors, add `SetEnvKeyReplacer(strings.NewReplacer(".", "_"))`, warn on unknown keys, and cut the `init` template down to keys that do something. Wire `server.port`, `server.open_browser`, `llm.chunk_size`, and `llm.overlap` as flag defaults.

### B4. `jot watch` is a stub that reports success - P1, Reproduced

- **Where:** `cmd/jot/watch.go:21-24`, `Makefile:88`, `README.md:339-340`
- **What:** Prints "not yet implemented" and exits 0. `make dev` depends on it, the root help lists "Live reload during development", and the README FAQ says it works.
- **Fix:** Implement it (fsnotify is already in the module graph via viper; debounce, rebuild, optionally serve), or hide the command and correct the docs until it exists. At minimum return a non-zero exit code.

### B5. Damaged user-facing text - P2, Reproduced

- `jot init` ends with an orphaned `3. Run 'jot serve' to preview locally`; steps 1 and 2 are missing (`cmd/jot/init.go:166-167`).
- Many messages start with a stray space where a glyph was apparently removed (`cmd/jot/build.go:54,87,97,118,169`, throughout `export.go` and `toc.go`), and the README generated by `init` has list items with doubled spaces (`init.go:151-154`).
- The "Project Structure" tree in `README.md:204-209` is mangled to the point of being unreadable.
- `internal/compiler/markdown.go:109` still contains a pin glyph, which conflicts with the no-emoji convention.

### B6. Build summary is inaccurate - P2, Reproduced

`Generated %d HTML files` reports `len(allDocs)`, not files written, so it is wrong whenever `A3` occurs. See `A1` for the hardcoded assets line (`cmd/jot/build.go:113-114`).

### B7. `jot serve` defaults - P1, Code-evident

- **Where:** `cmd/jot/serve.go:29,67,86,91`
- Binds `:8080` on all interfaces. A local preview server should default to `127.0.0.1` with a `--host` opt-in.
- `-o` means `--open` here and `--output` in `build` and `export`. `--open` defaults to true, so the only way to disable it is `--open=false`. Prefer `--no-open`.
- Registers on `http.DefaultServeMux`, which prevents testing and graceful shutdown. No signal handling.
- `r.URL.Path = strings.TrimPrefix(...)` at `:86` is undone by `http.FileServer` and can be deleted.

### B8. Development command exposed to users - P2, Code-evident

`jot debug` (`cmd/jot/build_debug.go:13-21`) appears in `--help`. Unlike `build`, it errors on a missing input path instead of skipping it. Mark it `Hidden: true` or fold it into `build --verbose` / a `--dry-run` flag.

### B9. `loadBuildConfig` is shared across commands with different flag meanings - P2, Code-evident

`export`, `toc`, and `debug` call `loadBuildConfig(cmd)` (`cmd/jot/build.go:188-224`), which reads `--output` and `--clean`. On `export`, `--output` is a file path, so `OutputPath` is set to a file. Harmless today, a trap once `A2`'s fix starts using `OutputPath` for exclusion.

### B10. Generated sites contain placeholder product chrome - P1, Code-evident

- **Where:** `internal/renderer/template.go:14-16,53-58,91-96,120`; `cmd/jot/build.go:100-104`
- Every generated site ships a "Sign in" link and avatar pointing to `#`, a feedback widget that records nothing, header links ("API", "Documentation", "Support") hardcoded to `#`, a hardcoded "Copyright 2026", and a runtime dependency on Google Fonts (a privacy and offline concern for internal docs).
- Commit `c303e52` is described as "config-driven branding", but only the project name is configurable.
- **Fix:** Drive nav links, footer text, and optional widgets from config, defaulting to absent. Self-host or drop the web fonts.

### B11. Breadcrumbs use absolute URLs - P2 (corrected), Code-evident

- **Where:** `internal/renderer/renderer.go:418,425,449,452`
- **Correction (Phase 1):** the page template never renders `.Breadcrumb`, so these URLs never reach generated output. This is dead code (see `D3`), not a user-facing defect. The original text follows for context.
- Everything else in the template uses `RelativePrefix`, but breadcrumbs link to `/` and `/dir/`. These break under a sub-path (any GitHub Pages project site) and under `file://`, which the search index explicitly supports. Directory crumbs point to pages that are never generated.
- **Fix:** Delete `GenerateBreadcrumb` and `PageData.Breadcrumb` with the rest of `D3`, or, if breadcrumbs are wanted in the template, prefix with `relativePrefix` and render directory crumbs as plain text.

### B12. Accessibility basics - P2, Code-evident

The menu toggle button has no `aria-label` (`template.go:28`), the search input has no label, and `lang="en"` is hardcoded.

---

## 5. Performance

Baseline on record: 2,679 files in 10.81s, about 4 ms per file (`docs/benchmarks/stripe-docs-test.md`,
dated 2025-01-15, which predates the search index and markdown-copy work).

### C1. Chunking is superlinear in document size - P1, Reproduced

- **Where:** `pkg/chunking/fixed.go:47-66,100-120`; duplicate at `pkg/export/export.go:243-318`
- **Measured:** one 100 KB document, 0.51s; one 400 KB document, 5.11s. Four times the input, ten times the time, including fixed startup cost.
- **Cause:** For each chunk, the code tokenizes the entire remaining document, then binary-searches byte positions by re-tokenizing prefixes (about 17 tokenizations per boundary), then does it again for the overlap. Work per chunk is proportional to remaining length, so total work is quadratic. The recursive strategy has the same shape (`recursive.go:101-108` re-counts a growing accumulator).
- **Fix:** Encode once, slice the token ID array into windows with the requested overlap, decode each window. This is linear. It needs `Decode` on the `Tokenizer` interface and a cumulative byte-offset calculation for `StartPos`/`EndPos`. Take care when a window boundary falls inside a multi-byte rune (extend to the next rune start).
- **Why the benchmarks missed it:** `pkg/chunking/benchmark_test.go:18-23` uses a fixed ~10 KB input. Add 100 KB and 1 MB cases.

### C2. Stop-word map rebuilt for every word - P1, Code-evident

`isCommonWord` constructs a 36-entry map literal on each call and is called once per word (`internal/toc/builder.go:250-264`, `internal/search/indexer.go:251-265`). Move it to a package-level variable. Two-line change; removes an allocation per word of the corpus, twice.

### C3. Regular expressions compiled in hot paths - P1, Code-evident

`regexp.MustCompile` runs per document or more often at: `internal/renderer/renderer.go:112,116,132,141,144,148,246`; `internal/toc/builder.go:112` (per path segment per document), `:162,164,231,235,239`; `internal/search/indexer.go:131,182-212` (and `cleanContent` runs twice per document, from `:76` and `:79`); `pkg/scanner/document.go:67,84,129,161,214`. Hoist all of them to package-level variables. The second regex in `enhanceCodeBlocks` (`renderer.go:116-117`) replaces `<pre><code>` with itself and can be deleted.

### C4. HTML template parsed once per page - P1, Code-evident

`renderTemplate` (`internal/renderer/renderer.go:351-358`) parses the ~200-line template for every document. `HTMLRenderer.templates` (`:23`) exists for this and is unused. Parse once in `NewHTMLRenderer`.

### C5. Navigation and prev/next are recomputed per page - P2, Code-evident

`GenerateNavigation` and `computeAdjacentPages` (`renderer.go:169-172,193-240`) walk and copy the whole TOC for each page, which is O(N squared) overall and embeds a full nav string in each of N pages. Compute the flattened order once. Render navigation once per directory depth (the only thing that changes the relative prefix) and mark the active link by string patch or client-side.

### C6. The pipeline is single-threaded - P2, Code-evident

Scanning (`pkg/scanner/scanner.go:51-86`) and compilation (`internal/compiler/compiler.go:43-47`) are sequential. Rendering is independent per document. A bounded worker pool is a contained change once `C3`-`C5` are done; measure first, since those may be enough.

### C7. The scanner walks ignored directories - P1, Code-evident

`scanner.go:57-59` returns `nil` for every directory, and ignore patterns are only tested against `.md` files. `node_modules/`, `.git/`, and `vendor/` are fully traversed. Test directories against the filter and return `filepath.SkipDir`. Also replace the second `os.Stat` (`:103`) with `d.Info()`.

### C8. Unused per-document extraction - P2, Code-evident

`readDocument` always extracts sections, links, and code blocks (`scanner.go:130-132`). `jot build` uses none of them. Make extraction lazy or opt-in.

### C9. Search index size and loading - P1, Code-evident

- **Where:** `internal/search/indexer.go:229,235-245`, `internal/renderer/template.go:213`, `web/templates/assets/search.js:68-103`
- The index holds the full cleaned text of every document, is pretty-printed, is written twice (`.json` and `.js`), and is loaded by a blocking `<script>` on every page view. For a corpus the size of the benchmark this is likely several megabytes per navigation.
- Search is a case-insensitive substring match. `search.fuzzy: true` in the config implies otherwise.
- **Fix:** Compact JSON; ship one file; add `defer` and load on first focus of the search box; cap indexed content per document or move to an inverted index if large sites are a goal.

### C10. Each page carries its markdown twice - P2, Code-evident

Pages embed a base64 copy of their source (`renderer.go:185`, `template.go:211`), and the `.md` is also written beside the page (`compiler.go:92-98`). Base64 adds about 33%, so each page is roughly 2.3x its content size. Fetching the adjacent `.md` on click removes the embed, at the cost of `file://` support. See [Decision 3](#9-decisions-needed-from-the-maintainer).

### C11. Exporters materialize the whole output as a string - P2, Code-evident

`ToJSONL`, `ToLLMSFullTxt`, and the others build a complete string that the CLI then writes. Accept an `io.Writer` to bound memory, which matters for the large corpora the README targets.

### C12. The tokenizer needs the network on first use - P1, Code-evident

`tiktoken-go` downloads the `cl100k_base` vocabulary over HTTP on first call and caches it under `$TMPDIR/data-gym-cache` (confirmed in the module's `load.go`). Consequences: `jot export --format jsonl` fails offline, in sandboxed CI, and in air-gapped environments; the first run is slow; and the test suite is not hermetic (`E7`). The companion `tiktoken-go-loader` module embeds the vocabulary for offline use at the cost of a few megabytes of binary size. See [Decision 4](#9-decisions-needed-from-the-maintainer).

Related wording: the README and `pkg/tokenizer` say `cl100k_base` is "Claude compatible". It is an OpenAI encoding. For Claude models the counts are an approximation and should be described that way.

---

## 6. Architecture and maintainability

### D1. `chunking` depends on `export`, which forces duplication - P0 (it is the root cause of A6 and A7)

`pkg/chunking/strategy.go:5` imports `pkg/export` only for the `Chunk` struct. That makes `export -> chunking` a cycle, which is why `export` carries its own private chunker. Move `Chunk` into `pkg/chunking` (it is the natural owner), have `export` import `chunking`, and give exporters a `ChunkStrategy` parameter. Update the dependency diagram in `CLAUDE.md`, which currently documents the inverted direction as intended.

### D2. Duplicated logic - P1

| Logic | Copies | Locations | Divergence |
|---|---|---|---|
| Fixed-size chunker | 2 | `pkg/chunking/fixed.go`, `pkg/export/export.go:221-331` | ID format differs (`A7`) |
| Keyword extraction | 3 | `toc/builder.go:196`, `search/indexer.go:147`, `export/export.go:335` | Thresholds 2 / 3 / 1, different stop-word lists |
| `cleanContent` | 2 | `toc/builder.go:229`, `search/indexer.go:180` | Indexer strips more |
| Tags, read time, summary | 2 | `toc/builder.go:139-193`, `search/indexer.go:69-127` | |
| `.md` to `.html` | 9 | `compiler.go:104,155`; `renderer.go:223,230,263,291,319,449`; `indexer.go:117` | All use `strings.Replace(p, ".md", ".html", 1)`, which rewrites the first `.md` anywhere in the path, not the suffix |
| Scan-all-inputs loop | 4 | `build.go:56-78`, `export.go:201-221`, `toc.go:93-117`, `build_debug.go:31-47` | Missing path: build prints, export is silent, debug errors |
| Anchor slugging | 3 | `export/markdown.go:113-136` (twice), `scanner/document.go:211` | |

Consolidate into one text-utilities home and one `scanInputs` helper in `cmd`.

### D3. Dead code - P2

Unreferenced outside tests, confirmed by search: `compiler.MarkdownCompiler` (all 185 lines of `internal/compiler/markdown.go`), `renderer.containsActivePage`, `HTMLRenderer.templates`, `TOCNode.SortChildren` and `Weight`, `TableOfContents.MarshalXML`, `Scanner.ScanSingle`, `scanner.LoadIgnoreFile`, `SemanticStrategy.semanticBoundaryDetection`, `export.minInt`, `Document.HTML`, `IndexDocument.ContentHash` (declared, never set), `MarkdownExporter.contextualEnrichment`, and the legacy `llm` export branch. Each should be wired up or deleted; stubs with long TODO comments cost reading time and imply features.

### D4. Business logic lives in Cobra handlers - P1

`runBuild` (`cmd/jot/build.go:36-172`) performs scanning, TOC generation, compilation, and llms.txt writing, printing directly to stdout. That is why `cmd/jot` is hard to test and why `compiler` has no tests. Move the pipeline behind a function that takes a typed config and an `io.Writer` for progress. Handlers then only translate flags.

### D5. `pkg/` is now public API and should behave like it - P1

Commit `91deef9` moved scanner, tokenizer, export, and chunking to `pkg/` "for external imports". As a library surface it currently logs to the global logger (`pkg/export/llmstxt.go:198`), swallows file read errors (`pkg/scanner/scanner.go:78-82`, whose comment says "Log error" but nothing is logged), exposes `map[string]interface{}` metadata, and exports stubs. Return warnings to the caller, collect per-file errors, and state a stability policy before anyone depends on it.

### D6. Hand-rolled XML - P2

`internal/toc/toc.go:22-160` builds XML by string concatenation with a custom escaper, and omits the XML declaration. See `A14`. `encoding/xml` with struct tags would be roughly a third of the code and correct by construction.

### D7. `toc.Builder.Build` mutates its input - P2

It sorts the caller's slice in place (`builder.go:32,100-104`). Subsequent steps in `runBuild` silently depend on that order. `sortDocumentsByImportance` in `llmstxt.go` copies first; do the same here, or sort once, explicitly, in the pipeline.

### D8. Version has four sources that disagree - P1

`cmd/jot/main.go:10` says `0.1.0`; `docs/VERSION` says `0.1.0`; `docs/CHANGELOG.md` records `0.2.0` released 2026-02-27; `docs/roadmaps/rmap_01.md` says "Current Version: 1.0"; `jot.yml` has its own `version`. Combined with `F1`, a `make build` binary reports an empty string. Derive the version from the git tag via ldflags and delete the rest.

---

## 7. Test coverage

| ID | Gap | Recommendation |
|---|---|---|
| E1 | `internal/compiler` has no tests, and it is where A1-A4 live. | End-to-end test: write a fixture tree to `t.TempDir()`, `chdir` somewhere else, build, and assert that assets exist, `index.html` exists, written file count equals document count, and a second build produces the same file list. |
| E2 | `runExport`, `validateExportFlags`, `runTOC`, `runInit`, `runServe` are 0%. | Command-level tests using `rootCmd.SetArgs` with captured stdout and stderr. Assert stdout parses as JSON (`B1`), that strategies produce different output (`A6`), that presets apply, and that `toc -r` differs from `toc` (`A15`). |
| E3 | `recursiveSplit` is 0%. Only the "fits in one chunk" early return is exercised. | Table tests with inputs that force each separator level. |
| E4 | No invariant tests for chunkers. | For every strategy: each chunk is within `maxTokens`; `content[StartPos:EndPos] == Text`; text is valid UTF-8; IDs are unique across documents; chunks cover the document. These would have caught `A7` and `A13`. |
| E5 | `ExtractFrontmatter` is 28.6%; tokenizer has no direct tests. | Cover with `A8`. Add a small tokenizer test for known strings. |
| E6 | Nothing checks determinism. | Build twice, compare bytes (`A12`). |
| E7 | Tests need the network on a cold cache (`C12`). | Offline vocabulary loader, or a fake `Tokenizer` for unit tests with real-tokenizer tests behind a build tag. |
| E8 | Benchmarks use only a 10 KB input. | Add 100 KB and 1 MB cases (`C1`). |
| E9 | No regression tests for parsing edge cases. | Fences with `c++` and `~~~`, `#` inside code, CRLF frontmatter, titles containing `<` and `&`, a document and directory sharing a name. |

---

## 8. Developer experience, build, and documentation

### F1. `make build` produces a binary with no version - P0, Reproduced

`Makefile:2` runs `cat VERSION`, but the file is at `docs/VERSION`. `make -n build` shows `-X main.version=` (empty), which overrides the `0.1.0` default in `main.go`. `scripts/release.sh:8` has the same bug and, with `set -e`, aborts immediately.

### F2. Release packaging cannot work as written - P1, Code-evident

`Makefile:79-81` uses `tar rzf` to append to a `.tar.gz`; tar cannot update compressed archives. `[ -f LICENSE ]` is false because the license is at `docs/LICENSE`. There are two release paths (Makefile with 5 platforms, `release.sh` with 7) that will drift. `make docker-build` (`Makefile:122`) references a Dockerfile that does not exist. Pick one path; GoReleaser would replace both along with checksums and archive naming.

### F3. No CI - P0

The README shows a CI badge for a workflow that is not in the repository, and `CONTRIBUTING.md` tells contributors to "ensure CI passes". A minimal workflow: `gofmt -l` must be empty, `go vet`, `go test -race ./...`, `golangci-lint`, `go mod tidy -diff`, and the out-of-tree build test from `E1`.

### F4. Formatting and lint are not enforced - P2, Reproduced

Twelve files fail `gofmt -l`, mostly import grouping: `cmd/jot/{build,build_debug,export,toc}.go`, `internal/compiler/{compiler,markdown}.go`, `internal/renderer/{renderer,renderer_test}.go`, `pkg/chunking/{semantic,strategy_test}.go`, `pkg/export/{llmstxt_test,types}.go`. There is no `.golangci.yml`, and `make lint` assumes a globally installed binary. Add a config (suggested linters: `errcheck`, `staticcheck`, `unused`, `gocritic`, `gosec`) and pin the tool version.

### F5. `go.mod` is not tidy - P2, Reproduced

`github.com/pkoukk/tiktoken-go` is imported directly but marked `// indirect`.

### F6. Install documentation is inaccurate - P1, Code-evident

`docs/INSTALL.md` states Go 1.19+ (`go.mod` requires 1.22.3), points to `./install.sh` (it is in `scripts/`), and recommends `go install`, which yields a binary that cannot produce a styled site until `A1` is fixed.

### F7. Small repository hygiene items - P2

- `.gitignore:52` reads `scripts/ghub.sh;` with a trailing semicolon and the wrong directory, so it matches nothing; `scripts/internal/ghub.sh` is tracked.
- `scripts/internal/ghub.sh` commits with the message `"."` and pushes to main. It probably should not live in a repository that publishes contribution guidelines.
- The `.claude/` hooks run an external tool (`oober`) on every edit with a 3-5 second timeout. This is undocumented for contributors who do not have it installed.

### G. README claims compared with behavior

| README claim | Reality | Finding |
|---|---|---|
| CI badge (`:3`) | No workflow exists | F3 |
| License link to `/LICENSE` (`:5`) | File is at `docs/LICENSE` | - |
| "Pluggable Chunking - fixed, semantic, markdown-headers, recursive" (`:17`) | Strategy flag is ignored | A6 |
| "Semantic: Embedding-based boundary detection" (`:265`) | Stub, falls back to fixed | A16 |
| "Contextual" strategy (`:268`) | Not in the factory | A6 |
| "Zero-Copy Markdown - Symlink support" (`:22`) | No symlink handling anywhere in the code | - |
| "Frontmatter - YAML metadata" (`:215`) | Stripped, not parsed | A8 |
| `.jotignore` (`:90`) | Never read | A9 |
| `jot watch` works (`:339-340`) | Not implemented | B4 |
| `--verbose` shows scanned files (`:301`) | Only prints the config file name and one TOC count | - |
| "Build fails with config file not found" (`:293`) | Cannot occur; config errors are swallowed | B3 |
| `features.search`, `features.toc`, `output.format`, `output.theme`, `llm.*` (`:187-199`) | Not read | B3 |
| "GPT-4/Claude compatible" tokens (`:19`, `:286`) | OpenAI encoding; approximate for Claude | C12 |
| `--include-embeddings` (`:143`) | No-op | A16 |
| Project structure diagram (`:204-209`) | Mangled | B5 |

The roadmap's "Completed Features" list and `docs/spec/architecture.md` should be reconciled at the same time.

---

## 9. Decisions needed from the maintainer

These change user-visible behavior, so they should be chosen rather than assumed. The maintainer's answers and their status are in [Section 12](#12-section-9-decisions-2026-09-24).

1. **URL layout for multiple input roots (`A3`).** Keep root-relative paths and fail on collision (no URL changes, smaller fix), or switch to project-relative paths (no collisions possible, but existing URLs change). Recommendation: fail on collision now; consider project-relative paths behind a config flag later.
2. **Scope of LLM export features (`A16`).** Embeddings and semantic chunking require an embedding provider, API keys, and cost controls. Recommendation: remove `--include-embeddings`, and have `semantic` either be removed or documented plainly as an alias for `fixed` until there is a concrete design. Update `--for-rag` to use `recursive` or `markdown-headers`, which exist.
3. **`file://` support versus page weight (`C10`).** The search index was deliberately made to work from `file://`. If that is a goal, keep the embedded markdown; if not, fetch the adjacent `.md`.
4. **Embed the tokenizer vocabulary (`C12`).** Offline and hermetic operation in exchange for a few megabytes of binary size. Recommendation: embed.
5. **`MarkdownCompiler` and other stubs (`D3`).** Wire up or delete. Recommendation: delete; git history keeps them.
6. **`pkg/` stability (`D5`).** Decide whether external importers are supported yet. If not, say so in the README so the API can still be corrected (`D1` is a breaking move of `Chunk`).

---

## 10. Remediation plan

Phases are ordered so that each one is protected by the tests added before it. Effort: **S** is under two hours, **M** is up to a day, **L** is multiple days.

### Phase 0 - Safety net

| Step | Addresses | Effort | Verify |
|---|---|---|---|
| Fix `VERSION` path in `Makefile` and `release.sh`; `go mod tidy`; `gofmt -w` | F1, F4, F5 | S | `make build && ./jot --version` prints a version; `gofmt -l .` and `go mod tidy -diff` are empty |
| Add CI workflow (fmt, vet, `test -race`, tidy check) | F3 | S | Workflow is green on main; README badge resolves |
| Add out-of-tree end-to-end build test. It is expected to fail at first; mark the known failures so each later fix turns one green | E1 | M | Test exists and fails for exactly A1, A2, A3, A4 |

### Phase 1 - Make the default path correct

| Step | Addresses | Effort | Verify |
|---|---|---|---|
| Embed assets; fail on write error; remove the hardcoded log line | A1, B6 | S | E2E: `dist/assets/style.css` exists when built from a temp directory |
| Exclude the resolved output directory in the scanner; prune ignored directories with `SkipDir` | A2, C7 | S | E2E: two consecutive builds report the same document count |
| Detect duplicate output paths and fail with both sources named | A3 | S | E2E: `docs/README.md` plus `README.md` returns an error naming both |
| Emit root README as `index.html`; simplify `serve` | A4 | S | E2E: `dist/index.html` exists; logo link resolves |
| Match TOC directories by path segment among non-leaf nodes | A5 | S | New test: `guides.md` plus `guides/intro.md` both appear in nav HTML |
| UTC timestamps; escape nav titles; relative breadcrumbs | A11, A14, B11 | S | Unit tests for a title containing `<`/`&`; timestamp test with a non-UTC location |

### Phase 2 - Make export do what it says

| Step | Addresses | Effort | Verify |
|---|---|---|---|
| Move `Chunk` to `pkg/chunking`; `export` depends on `chunking`; delete `export.chunkDocument`; update `CLAUDE.md` diagram | D1, D2 | M | `go build ./...`; no `chunkDocument` symbol remains |
| Pass the selected strategy from the CLI into JSONL and markdown exporters | A6, A7 | M | Command test: outputs for `fixed`, `markdown-headers`, `recursive` differ; all `chunk_id` values are unique |
| Progress to stderr; `SilenceErrors` and `SilenceUsage` | B1, B2 | S | Command test: `export --format json` stdout parses as JSON; error text appears once |
| Apply Decision 2: remove or relabel inert flags; implement overlap in `recursive` or reject it | A16 | S-M | `--help` lists only working options |
| Chunker invariant tests, including `recursiveSplit` | E3, E4 | M | `pkg/chunking` coverage above 85% |
| Token-window chunking (encode once, slice, decode) | C1, A13 | M | New 1 MB benchmark scales roughly linearly against the 100 KB case; invariants still pass |
| Offline tokenizer vocabulary (Decision 4) | C12, E7 | S | `go test ./...` passes with networking disabled and an empty cache |

### Phase 3 - Configuration and CLI ergonomics

| Step | Addresses | Effort | Verify |
|---|---|---|---|
| Typed config struct; surface parse errors; env key replacer; warn on unknown keys | B3 | M | Tests: malformed YAML returns an error; `JOT_OUTPUT_PATH` overrides; unknown key warns |
| Trim the `init` template to supported keys; wire `server.*` and `llm.*` as defaults; repair `init` output text | B3, B5 | S | Every key in the generated file is read somewhere (enforce with a test over the struct) |
| Parse frontmatter; shared tag helper; CRLF support | A8, E5 | S | Frontmatter title and list-form tags appear in page title, search index, and `toc.xml` |
| Fence-aware line iterator used by title, sections, headings, and header chunking; tolerant fence regex | A10, E9 | M | Fixture with `#` inside bash and a `c++` fence yields correct headings |
| Decide on `.jotignore` (implement with a library, or remove) | A9 | S-M | Either a test proving patterns apply, or no references remain |
| Implement or hide `watch`; hide `debug`; `serve` binds localhost, `--no-open`, own mux, graceful shutdown | B4, B7, B8 | M | `jot watch` rebuilds on change or is absent from help; `serve` listens on `127.0.0.1` by default |
| Fix `toc --recursive`; deterministic ordering | A15, A12 | S | Test: `-r` output differs and includes descendants; two runs are byte-identical |

### Phase 4 - Build performance

Measure before and after each step against a corpus of roughly 2,500 files and record results in `docs/benchmarks/`.

| Step | Addresses | Effort | Verify |
|---|---|---|---|
| Package-level stop-word map and regexes; parse template once; drop double `cleanContent` | C2, C3, C4 | S | `go test -bench` on renderer and indexer shows fewer allocs per op; build time recorded |
| Flatten TOC once; cache nav per depth | C5 | M | Build time at 2,500 files drops measurably; nav HTML unchanged in golden test |
| Compact single-file search index, deferred and lazy-loaded | C9 | M | Index size recorded before and after; search works over `file://` and `http://` |
| Lazy section/link/code extraction | C8 | S | Build benchmark |
| Parallel render with a bounded worker pool, only if still warranted | C6 | M | Build time; `go test -race` clean; output byte-identical to sequential |
| Stream exporters to `io.Writer` | C11 | M | Peak RSS on a large export recorded before and after |

### Phase 5 - Cleanup and documentation

| Step | Addresses | Effort | Verify |
|---|---|---|---|
| Delete or wire dead code (Decision 5) | D3 | S | `staticcheck` / `unused` report is clean |
| Consolidate text utilities, `.md` to `.html` via suffix replacement, and `scanInputs` | D2 | M | One definition of each; existing tests pass |
| Move the build pipeline out of `runBuild` behind a typed config and `io.Writer` | D4 | M | `cmd/jot` coverage above 70% without shelling out |
| `encoding/xml` for the TOC; stop mutating the input slice | D6, D7 | S | Golden test for `toc.xml`, now with an XML declaration |
| `pkg/` hygiene: no global logging, return per-file scan errors, stability note | D5 | M | No `log.` or `fmt.Print` calls under `pkg/` |
| Single version source from git tags; one release path (GoReleaser) | D8, F2 | M | Tagged dry-run produces archives and checksums for all platforms |
| Config-driven header links and footer; remove placeholder sign-in and feedback; self-host or drop web fonts; aria labels | B10, B12 | M | A default build contains no links to `#` and makes no third-party requests |
| Rewrite README, INSTALL, roadmap, and architecture spec against actual behavior; add `.golangci.yml`; fix `.gitignore` | G, F4, F6, F7 | M | Each row of the Section G table is resolved |

### If only one day is available

Phase 0 plus the first four rows of Phase 1 (`F1`, `F3`, `E1`, `A1`, `A2`, `A3`, `A4`) and the stderr change in `B1`. That makes a freshly installed binary produce a correct, styled, hostable site from the generated config, makes piped exports usable, and puts a test in place to keep it that way.

---

## 11. Progress and verification

### Phase 0 and Phase 1 status (branch `fix/remediation-phase-1`)

| Finding | Status | Commits |
|---|---|---|
| F4 gofmt | Done | `f73e44e` |
| F1 version stamping | Done (path only; value still disagrees with CHANGELOG, see D8) | `1dfff3a` |
| F5 go.mod tidy | Done | `6f46de9` |
| F3 CI | Done (fmt, tidy, vet, build, `test -race`; no lint yet) | `64ab2fe` |
| E1 end-to-end tests | Done (`cmd/jot/e2e_test.go`) | `7a2ad57`, extended in later commits |
| B9 config loading | Done | `4afd849` |
| A1 embedded assets, B6 summary | Done | `9aa7093`, `c70ac16` |
| A2 output exclusion | Done | `5077fc8`, `2c81699` |
| A3 duplicate output paths | Done | `9ae4b92`, `ff3f642` |
| A4 index.html | Done, including `serve` preferring `index.html` | `6e8f45c`, `729aa56` |
| A5 TOC merge | Done | `c1d57ac`, `f04e698` |
| A11 UTC timestamps | Done for `toc.xml` and the search index | `70fc925` |
| A14 escaping | Done for the sidebar and `toc.xml` attributes; not for the generated index page (V3) | `ad9ead7` |
| B1 export stdout | Done | `fe33c31`, `52dab47` |
| B2 duplicate errors | Done | `3be28fe`, `54eba8c` |
| B11 breadcrumbs | Not a defect; reclassified above | - |
| New: `clean` deleting sources | Done | `98f85c3` |

Every fix has a test that was confirmed to fail with the fix reverted.

### Adversarial verification

After the first pass, seven read-only agents each tried to refute one group of fixes by building the binary from the branch and the pre-audit binary from `88fe938`, running both out of tree, and diffing results; one more reviewed the full diff. They found regressions introduced by the first pass, all since fixed in the follow-up commits listed above:

- Output exclusion compared path strings, so symlinked or differently-cased output paths were re-scanned, and an output directory equal to an input root skipped all input (`2c81699`).
- The collision check ran after `clean` had deleted the previous site, missed case-only collisions on macOS, and rejected one file reached through two input paths (`ff3f642`).
- A root `Index.md` was overwritten by the README on case-insensitive filesystems (`729aa56`).
- Separated TOC nodes shared `id` values in `toc.xml` (`f04e698`).
- `--verbose` still wrote to stdout during export; some formats ended with several newlines (`52dab47`).
- Usage hints disappeared for flag and unknown-command errors (`54eba8c`).
- The embed glob included dotfiles; subdirectories were skipped (`c70ac16`).

While reordering `clean`, an older data-loss bug surfaced: `clean` with an output directory that contained an input directory deleted the sources. `build` now refuses (`98f85c3`).

### New findings from verification (not yet addressed)

These predate this branch unless noted, and belong in the later phases.

| ID | Priority | Finding | Where |
|---|---|---|---|
| V1 | P1 | `export --format yaml` is not valid YAML to PyYAML or Ruby Psych when a document has frontmatter or leading blank lines (`- content: \|4` block scalars). | `pkg/export/export.go` `ToYAML` |
| V2 | P1 | Non-ASCII directory names produce invalid UTF-8 titles: `humanizeTitle` upper-cases the first byte, not the first rune. `toc.xml` then fails to parse. | `internal/toc/builder.go:152` |
| V3 | P1 | The synthesized index page writes titles and paths into markdown unescaped; Blackfriday passes inline HTML through, so a title containing `<script>` runs in `index.html`. Same class as A14. | `internal/compiler/compiler.go:198,201` |
| V4 | P2 | Sidebar hrefs are HTML-escaped but not URL-encoded, so file names containing `#`, `?`, or `%` are unreachable from the sidebar. `search.js` builds result hrefs without escaping. | `internal/renderer/renderer.go:305,325`, `web/templates/assets/search.js:126-131` |
| V5 | P2 | `export --format json` and `markdown` write `modified` with the local UTC offset, so output bytes vary by machine time zone (A12). | `pkg/export/export.go:138`, `pkg/export/markdown.go:65` |
| V6 | P2 | A markdown export written inside an input path (the help text's own `--output docs.md` example) is scanned as source next time. `build -o X` is invisible to a later `export` or `toc`, which only know `output.path`. | `cmd/jot/export.go`, `cmd/jot/build.go` |
| V7 | P2 | Output paths that differ only in Unicode normalization (NFC vs NFD) are not detected as collisions. | `cmd/jot/build.go` `dedupeDocuments` |
| V8 | P2 | Titles written with HTML entities (`&lt;T&gt;`) now show the literal entity in the sidebar, matching `<title>` and prev/next links, where before the sidebar alone decoded them. Behavior change from A14. | `internal/renderer/renderer.go` |
| V9 | P2 | Tests in `cmd/jot` mutate process globals (working directory, viper, `rootCmd`, `exportCmd`, `os.Stdout`). They pass shuffled and with `-race`, but cannot use `t.Parallel()`. Resolved properly by D4. | `cmd/jot/e2e_test.go` |
| V10 | P2 | The new `web` package sits outside `internal/`, so `web.Assets` is importable by other modules. Moving the assets under `internal/` would keep it private. | `web/embed.go` |

## 12. Section 9 decisions (2026-09-24)

Implemented on branch `feat/section9-decisions`.

| Decision | Choice | Status | Commits |
|---|---|---|---|
| 1. URL layout for multiple input roots | Both: keep root-relative paths with collision errors as the default, and add opt-in project-relative paths | Done. New `output.structure: input \| project` in `jot.yml`; `project` makes paths relative to the working directory for `build`, `export`, `toc`, and `debug`. New `Scanner.RelativeTo` in `pkg/scanner` so library users can opt in | `30bb9a8`, `90b5189`, follow-ups below |
| 2. Scope of LLM export features | Design it, and check whether mlpipe answers the question | Designed, awaiting approval. See [llm-export-design.md](llm-export-design.md) | - |
| 3. `file://` support versus page weight | Keep embedded markdown (no change) | No change needed | - |
| 4. Embed the tokenizer vocabulary | Embed | Done. `cl100k_base` embedded in `pkg/tokenizer`, built without replacing tiktoken's process-wide loader; no new module, so mlpipe's `go.sum` is unaffected | `a31b171`, `07c4b99`, `4bcebac` |
| 5. Dead code and stubs | Delete | Done, except stubs tied to Decision 2 (`SemanticStrategy`, `contextualEnrichment`, the legacy `llm` CLI branch, `--include-embeddings`), which the design settles | `5d28aa6`, `a685f32`, `5436893`, `f0450e4`, `7db0997`, `86a60e5` |
| 6. `pkg/` stability | Importable, not API-stable before 1.0; mlpipe moves in lockstep | Done. README "Go Packages" section; `CLAUDE.md` module table corrected | `d3ab0cc`, `1905d46` |

`B11` (breadcrumb URLs) is closed: the dead breadcrumb code was removed in `a685f32`.

### Verification of this branch

Four read-only agents tried to refute the changes by running real binaries out of tree against the base (`2acde71`), and one reviewed the diff. Fixed in follow-up commits:

- With the output directory equal to, or containing, an input path, a document's markdown copy could land somewhere other than on its own source: in project mode `-o docs` replaced `docs/README.md` with the root README and nested `docs/docs/...` on every rebuild; in the default mode, input `docs` with output `.` replaced the project's own `README.md`. `build` now refuses such layouts before writing anything (`8bf3f5f`, `1a7a6c7`).
- The embedded vocabulary broke on checkouts with `core.autocrlf` (Git for Windows' default) (`07c4b99`), and the offline tokenizer test passed against the old implementation when run after other tests (`4bcebac`).
- `output.structure` accepted YAML lists as the default (`f687bb5`); the outside-root error repeated itself and was wrong for ancestor inputs (`27972e2`).
- One file reached through a symlink or a differently-cased path was built twice or reported as colliding with itself (`b1b524c`).
- Page paths replaced the first `.md` anywhere in a path, and ignored uppercase `.MD` (`aa7b5ec`, part of D2).
- Document IDs were hashed before `/` normalization, so they differed on Windows (`5f19e30`).
- Stale comments and docs (`1905d46`).

The mlpipe gate used during this work reported only build and vet failures for most commits, because `go test` was piped into `grep`. It was fixed, and mlpipe's build, vet, and tests were then re-run against every commit that claimed them; all pass.

### Remaining items

| ID | Priority | Finding | Where |
|---|---|---|---|
| V11 | P2 | An in-place build that generates a contents page (no root `index.md` or `README.md` among the documents) writes `index.md` into the output directory, replacing a root `index.md` that is not an input. Same in the base. | `internal/compiler/compiler.go` `generateIndexPage` |
| V12 | P2 | CI runs only on Linux, so the Windows fixes (CRLF vocabulary, ID normalization) are untested. A `windows-latest` job would cover them. | `.github/workflows/ci.yml` |
| V13 | P2 | `NewTokenizer` re-parses the 1.7 MB vocabulary on every call (tiktoken used to memoize it). Called once or twice per export, so the cost is small. | `pkg/tokenizer/tokenizer.go` |
| V14 | P2 | `TableOfContents.GetNodeByID` and the `Index` map it reads are used only by tests. | `internal/toc/toc.go` |
| V15 | P2 | Misspelled `jot.yml` keys and nested `JOT_*` environment variables are silently ignored (B3); `docs/spec/architecture.md` still documents removed fields (Phase 5). | `cmd/jot/root.go`, `docs/spec/architecture.md` |

### Recommendations for mlpipe (not changed here)

- Call `Scanner.RelativeTo` with one shared root per run, `Exclude(outputDir)`, and drop repeated `doc.Path` values in `ProcessMixed`, so a `README.md` in two markdown directories does not produce identical document and chunk IDs.
- `--format llm` writes nothing (`cmd/mlpipe/process.go` only writes `result.JSONL`; mlpipe's own `export-bugs.md` Bug 2).
- The vectors mlpipe's `Embedder` computes never reach `result.JSONL`; the Decision 2 design fixes this on the jot side.

## Appendix A. Reproduction commands

All commands were run with a binary built from `88fe938` into a scratch directory (`go build -o $SP/jot ./cmd/jot`), from fixture directories outside the repository. Nothing in the repository was modified.

```bash
# A1, A2, A4 - fixture: README.md, guides.md, guides/intro.md, node_modules/pkg/README.md; no jot.yml
jot build && ls dist/assets          # only search-index.js and search-index.json
ls dist/index.html                   # No such file or directory
jot build | grep Found               # Found 8 markdown files
jot build | grep Found               # Found 12 markdown files

# A3 - jot.yml with input.paths: ["docs", "README.md"]; files README.md and docs/README.md
jot build                            # "Generated 2 HTML files"
find dist -name '*.html'             # dist/README.html only

# A5, A8, A10, A11 - same fixture, with dist and node_modules ignored
grep -o '<div class="nav-tree">.*</div>' dist/README.html   # no link to guides/intro.html
grep -o '<title>[^<]*</title>' dist/guides/intro.html       # H1 text, not the frontmatter title
grep -A4 '"headings"' dist/assets/search-index.json         # includes "this is a shell comment"
grep -o 'modified="[^"]*"' dist/toc.xml                     # local wall-clock time with a Z suffix

# A6, A7 - two multi-chunk documents added to the fixture
for s in fixed semantic markdown-headers recursive; do
  jot export --format jsonl --strategy $s --chunk-size 256 --chunk-overlap 32 -o out-$s.jsonl
  shasum out-$s.jsonl                # identical digest for all four
done
# 40 rows, 25 distinct chunk_id values; "chunk-0" occurs 5 times

# B1, B2, A16
jot export --format json | python3 -m json.tool             # fails: stdout begins with progress text
jot export --format nope 2>&1 | grep -c "unsupported"       # 2
jot export --format llm                                     # rejected by validation
jot export --format jsonl --include-embeddings -o e.jsonl   # warning printed; no "vector" field

# B3
JOT_OUTPUT_PATH=/tmp/elsewhere jot build                    # still writes ./dist
# replace jot.yml with invalid YAML, then:
jot build                                                   # no error; scans "." using defaults

# A15, B4, B5
jot toc --dry-run ; jot toc --dry-run -r                    # identical listings
jot watch ; echo $?                                         # "not yet implemented", exit 0
jot init                                                    # last line: "  3. Run 'jot serve' to preview locally"

# C1 - single generated markdown document per run
/usr/bin/time -p jot export --format jsonl -o o.jsonl       # 100 KB: real 0.51   400 KB: real 5.11

# F1, F4, F5 - from the repository root (read-only)
make -n build                        # cat: VERSION: No such file or directory ... -X main.version=
gofmt -l .                           # 12 files
go mod tidy -diff                    # moves tiktoken-go to the direct require block
```

# Jot LLM Export Implementation - Task Files

**Created:** 2025-10-21  
**Status:** Ready for execution  
**Total Tasks:** 8 (64 subtasks)  
**Estimated Duration:** 14.5 hours (parallelized) | 22.5 hours (sequential)

---

## 📦 What Was Created

### Task Files (Ultra-Atomic YAML)
All files are in `.tks/todo/`:

1. **jot-export-001-llmstxt.yml** - Implement llms.txt format (2h)
2. **jot-export-002-llmsfull.yml** - Implement llms-full.txt format (1.5h)  
3. **jot-export-003-tokenization.yml** - Fix token-based chunking ⚠️ CRITICAL (4h)
4. **jot-export-004-jsonl.yml** - Implement JSONL export (3h)
5. **jot-export-005-markdown.yml** - Implement enriched markdown (2.5h)
6. **jot-export-006-chunking.yml** - Implement chunking strategies ⚠️ CRITICAL (5h)
7. **jot-export-007-build-integration.yml** - Integrate into build (2h)
8. **jot-export-008-cli-updates.yml** - Update CLI (2.5h)

### Documentation Files
- **EXECUTION_GUIDE.md** - Complete execution instructions
- **DEPENDENCY_GRAPH.md** - Visual dependency graph + analysis
- **README.md** - This file

---

## 🎯 Quick Start for Agents

### Step 1: Identify What Can Start Now
```bash
# Find tasks with no dependencies
grep -B1 "dependencies: \[\]" jot-export-*.yml | grep "^id:"
```

**Result:**
- jot-export-001 (llms.txt)
- jot-export-002 (llms-full.txt)
- jot-export-003 (tokenizer) ⚠️ CRITICAL PATH
- jot-export-005 (markdown)

### Step 2: Assign Agents to Parallel Groups

**Phase 1 - Start Immediately (3 agents in parallel):**
```bash
Agent-A: jot-export-001-llmstxt.yml
Agent-B: jot-export-002-llmsfull.yml
Agent-C: jot-export-005-markdown.yml
```

**Phase 2 - Critical Path (1 agent, sequential):**
```bash
Agent-D: jot-export-003-tokenization.yml
         ↓ (wait for completion)
         jot-export-006-chunking.yml
```

### Step 3: Execute Subtasks
Each YAML file contains 8 subtasks with:
- Exact file paths
- Function/struct names
- Acceptance criteria
- Parallel execution flags

### Step 4: Update Status
When complete, update the task file:
```yaml
status: "done"  # Change from "todo" to "done"
modified: "2025-10-21T16:30:00Z"  # Update timestamp
```

---

## 🔍 Each Task File Contains

```yaml
id: "jot-export-XXX"
task: "High-level objective"
priority: "H | M | S"
urgency: 1-10

# WHO can run this
dependencies: []           # Task IDs that must complete first
parallel_group: "A"        # Which group (A, B, C, D, E)
blocks: []                 # Tasks waiting on this
execution_phase: "PHASE_X" # Phase number

# WHAT to do (8 subtasks)
subtasks:
  - desc: "Exact file path + function name + single operation"
    agent: "code-implementation | test-implementation | documentation"
    parallel: true/false
    depends_on_subtask: "subtask-N" (if applicable)

# WHY and WHERE
must_reference:
  - "URL or file:line - Description"
```

---

## 📊 Execution Strategies

### Strategy 1: Maximum Speed (7 agents)
- **Duration:** ~14.5 hours
- **Cost:** High (7 concurrent agents)
- **Use when:** Deadline is critical

```
Phase 1: Agents A, B, C (parallel on Tasks 001, 002, 005)
Phase 2: Agent D (sequential on Tasks 003 → 006)
Phase 3: Agent E (on Task 004)
Phase 4: Agent F (on Task 007)
Phase 5: Agent G (on Task 008)
```

### Strategy 2: Balanced (3 agents)
- **Duration:** ~18 hours
- **Cost:** Medium (3 concurrent agents)
- **Use when:** Resource-constrained but time-sensitive

```
Agent 1: Tasks 001 → 007 → (wait) → 008
Agent 2: Tasks 002, 005 → (wait) → (help with 008)
Agent 3: Tasks 003 → 006 → 004 → (wait) → (help with 008)
```

### Strategy 3: Sequential (1 agent)
- **Duration:** ~22.5 hours
- **Cost:** Low (1 agent)
- **Use when:** No rush, single developer

```
Order: 003 → 001 → 002 → 005 → 006 → 004 → 007 → 008
Rationale: Do critical path first, then parallizable work, then integration
```

---

## ⚠️ Critical Warnings

### Do NOT Start These Until Dependencies Met:
- **Task 006** requires Task 003 (needs tokenizer)
- **Task 004** requires Task 006 (needs chunking strategies)
- **Task 007** requires Tasks 001 & 002 (needs llms.txt exporters)
- **Task 008** requires ALL tasks (final integration)

### Critical Path:
**003 → 006 → 004 = 12 hours**

This is the longest sequential chain. Optimize these tasks first.

---

## ✅ Validation Steps

After each phase:

**Phase 1:**
```bash
cd /Users/macadelic/dusk-indust/shared/packages/jot
go test ./internal/export/llmstxt_test.go
go test ./internal/export/markdown_test.go
```

**Phase 2:**
```bash
go test ./internal/tokenizer/...
go test ./internal/chunking/...
```

**Phase 3:**
```bash
go test ./internal/export/jsonl_test.go
```

**Phase 4:**
```bash
go test ./cmd/jot/build_test.go
jot build --help  # Verify --skip-llms-txt flag exists
```

**Phase 5:**
```bash
go test ./...
go build ./cmd/jot
./jot export --help  # Verify all new formats listed
```

---

## 📁 Files Created by Implementation

```
internal/
├── tokenizer/
│   └── tokenizer.go              [Task 003]
├── chunking/
│   ├── strategy.go               [Task 006]
│   ├── fixed.go                  [Task 006]
│   ├── headers.go                [Task 006]
│   ├── recursive.go              [Task 006]
│   ├── semantic.go               [Task 006]
│   ├── factory.go                [Task 006]
│   ├── strategy_test.go          [Task 006]
│   └── benchmark_test.go         [Task 006]
└── export/
    ├── llmstxt.go                [Tasks 001, 002]
    ├── llmstxt_test.go           [Tasks 001, 002]
    ├── jsonl.go                  [Task 004]
    ├── jsonl_test.go             [Task 004]
    ├── markdown.go               [Task 005]
    └── markdown_test.go          [Task 005]

cmd/jot/
└── build_test.go                 [Task 007]

Modified:
├── internal/export/export.go     [Task 003 - chunkDocument]
├── internal/export/types.go      [Multiple tasks - new structs]
├── cmd/jot/export.go             [Task 008 - CLI flags]
├── cmd/jot/build.go              [Task 007 - llms.txt gen]
├── jot.yml                       [Task 007 - config]
└── go.mod                        [Task 003 - tiktoken-go]
```

---

## 🎓 For Implementation Agents

When you pick up a task:

1. **Read entire YAML file** - Don't skip `must_reference` section
2. **Check dependencies** - Ensure all are `status: done`
3. **Read referenced files** - Understand context before coding
4. **Follow subtask order** - Unless marked `parallel: true`
5. **Write tests first** - For `test-implementation` subtasks
6. **Run tests frequently** - After each subtask if possible
7. **Update status** - Mark as `done` when all subtasks complete
8. **Git commit** - Create atomic commits per task

---

## 📞 Support References

- **Specification:** https://llmstxt.org/
- **Protocols:** `.tks/protocols/protodoc.md`
- **Integration:** `.tks/support/jot_integration_strategy.md`
- **Existing Code:** `internal/export/export.go`
- **Config:** `jot.yml`

---

## 🚀 Ready to Execute!

**Next Steps:**
1. Review EXECUTION_GUIDE.md for detailed instructions
2. Check DEPENDENCY_GRAPH.md for visual overview
3. Assign agents to Phase 1 tasks (001, 002, 005)
4. Assign dedicated agent to critical path (003 → 006)
5. Monitor progress and update task statuses

**Questions?**
- All task files are ultra-atomic with zero ambiguity
- Each subtask has exact file paths and function names
- All dependencies are explicitly declared
- Parallel execution groups are clearly marked

**Let's build this! 🎯**

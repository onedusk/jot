# LLM Export Implementation - Execution Guide

**Created:** 2025-10-21
**Total Tasks:** 8
**Total Subtasks:** 64
**Estimated Time:** 14 hours (parallelized) | 22.5 hours (sequential)

---

## 📊 Quick Reference: Task Dependencies

```
PHASE 1 (Group A - Parallel) ────────────────┐
├─ jot-export-001 (llms.txt)         2h      │
├─ jot-export-002 (llms-full.txt)    1.5h    │ Run concurrently
└─ jot-export-005 (markdown)         2.5h    │
                                              │
PHASE 2 (Group B - Sequential) ──────────────┤
├─ jot-export-003 (tokenizer)        4h      │ Must run first
└─ jot-export-006 (chunking)         5h      │ Depends on 003
                                              │
PHASE 3 (Group C) ───────────────────────────┤
└─ jot-export-004 (JSONL)            3h      │ Depends on 006
                                              │
PHASE 4 (Group D) ───────────────────────────┤
└─ jot-export-007 (build-integ)      2h      │ Depends on 001, 002
                                              │
PHASE 5 (Group E - Final) ───────────────────┤
└─ jot-export-008 (CLI updates)      2.5h    │ Depends on ALL
```

---

## 🎯 Agent Orchestration Rules

### Can Start When:
```python
def can_start(task_id, completed_tasks):
    task = load_task(f".tks/todo/jot-export-{task_id}.yml")

    # Check all dependencies satisfied
    for dep in task['dependencies']:
        if dep not in completed_tasks:
            return False

    return True
```

### Parallel Execution:
```python
def get_parallel_batch(tasks, completed):
    # Find all ready tasks
    ready = [t for t in tasks if can_start(t['id'], completed)]

    # Group by parallel_group
    groups = defaultdict(list)
    for task in ready:
        groups[task['parallel_group']].append(task)

    # Return first non-empty group
    for group in ['A', 'B', 'C', 'D', 'E']:
        if groups[group]:
            return groups[group]

    return []
```

---

## 📋 Execution Order

### **PHASE 1: Start Immediately (3 Parallel Agents)**

**Agent Alpha** → `jot-export-001-llmstxt.yml`
- Creates: `internal/export/llmstxt.go`
- Creates: `internal/export/llmstxt_test.go`
- Adds: `ProjectConfig` to `internal/export/types.go`
- Duration: ~2 hours

**Agent Beta** → `jot-export-002-llmsfull.yml`
- Modifies: `internal/export/llmstxt.go` (adds ToLLMSFullTxt)
- Duration: ~1.5 hours

**Agent Gamma** → `jot-export-005-markdown.yml`
- Creates: `internal/export/markdown.go`
- Creates: `internal/export/markdown_test.go`
- Duration: ~2.5 hours

**⏱️ Phase 1 completes when:** Longest task finishes (2.5h)

---

### **PHASE 2: After Phase 1 Starts (1 Sequential Agent)**

**Agent Delta** → `jot-export-003-tokenization.yml` **(CRITICAL PATH)**
- Runs: `go get github.com/pkoukk/tiktoken-go`
- Creates: `internal/tokenizer/tokenizer.go`
- Modifies: `internal/export/export.go:205` (chunkDocument)
- Modifies: `internal/export/types.go:36` (adds TokenCount)
- Duration: ~4 hours

**⚠️ WAIT FOR TASK 003 TO COMPLETE**

**Agent Delta** → `jot-export-006-chunking.yml`
- Creates: `internal/chunking/strategy.go`
- Creates: `internal/chunking/fixed.go`
- Creates: `internal/chunking/headers.go`
- Creates: `internal/chunking/recursive.go`
- Creates: `internal/chunking/semantic.go`
- Creates: `internal/chunking/factory.go`
- Creates: `internal/chunking/strategy_test.go`
- Creates: `internal/chunking/benchmark_test.go`
- Duration: ~5 hours

**⏱️ Phase 2 completes when:** Task 006 finishes (9h cumulative)

---

### **PHASE 3: After Task 006 Completes**

**Agent Epsilon** → `jot-export-004-jsonl.yml`
- Creates: `internal/export/jsonl.go`
- Creates: `internal/export/jsonl_test.go`
- Adds: `ChunkMetadata` to `internal/export/types.go`
- Duration: ~3 hours

**⏱️ Phase 3 completes when:** Task 004 finishes (12h cumulative)

---

### **PHASE 4: After Tasks 001 & 002 Complete**

**Agent Zeta** → `jot-export-007-build-integration.yml`
- Modifies: `cmd/jot/build.go:111` (BuildConfig)
- Modifies: `cmd/jot/build.go:100` (llms.txt generation)
- Modifies: `jot.yml:28` (features section)
- Creates: `cmd/jot/build_test.go`
- Duration: ~2 hours

**⏱️ Phase 4 can start:** 2.5h into Phase 1 (when Task 002 completes)

---

### **PHASE 5: After ALL Tasks Complete**

**Agent Omega** → `jot-export-008-cli-updates.yml`
- Modifies: `cmd/jot/export.go:25` (flags)
- Modifies: `cmd/jot/export.go:76` (switch statement)
- Modifies: `cmd/jot/export.go:20` (help text)
- Duration: ~2.5 hours

**⏱️ Phase 5 starts when:** ALL previous tasks done (14h)

---

## 🚀 Optimal Execution Timeline

```
Hour 0  ┌─ Alpha (001) ─────────┐
        ├─ Beta (002) ──────┐   │
        └─ Gamma (005) ──────────┴─┐
                            │       │
Hour 2.5                    └───────┴─ Phase 1 Complete
        ┌─ Delta (003) ────────────┐
        │                          │
Hour 4                             │
        │                          │
Hour 6.5                           └─ Task 003 Complete
        ├─ Delta (006) ────────────────────┐
        └─ Zeta (007) ─────┐              │
                           │               │
Hour 8.5                   └─ Task 007 Done│
                                           │
Hour 11.5                                  └─ Task 006 Complete
        ┌─ Epsilon (004) ─────┐
                              │
Hour 14.5                     └─ Task 004 Complete
        ┌─ Omega (008) ─────┐
                            │
Hour 17                     └─ ALL COMPLETE
```

**Total Duration:** ~17 hours wall-clock time with optimal agent allocation

---

## 📁 File Locations

All task files: `.tks/todo/jot-export-*.yml`

```
jot-export-001-llmstxt.yml          Phase 1, Group A
jot-export-002-llmsfull.yml         Phase 1, Group A
jot-export-003-tokenization.yml     Phase 2, Group B ⚠️ CRITICAL
jot-export-004-jsonl.yml            Phase 3, Group C
jot-export-005-markdown.yml         Phase 1, Group A
jot-export-006-chunking.yml         Phase 2, Group B ⚠️ CRITICAL
jot-export-007-build-integration.yml Phase 4, Group D
jot-export-008-cli-updates.yml      Phase 5, Group E (Final)
```

---

## ✅ Completion Checklist

**Phase 1:**
- [ ] Task 001: llms.txt export
- [ ] Task 002: llms-full.txt export
- [ ] Task 005: Markdown export

**Phase 2 (Critical Path):**
- [ ] Task 003: Tokenizer implementation
- [ ] Task 006: Chunking strategies

**Phase 3:**
- [ ] Task 004: JSONL export

**Phase 4:**
- [ ] Task 007: Build integration

**Phase 5:**
- [ ] Task 008: CLI updates

**Validation:**
- [ ] All 8 tasks marked as `status: done`
- [ ] `go test ./...` passes
- [ ] `go build` succeeds
- [ ] `jot export --help` shows new formats
- [ ] `jot build` generates llms.txt files

---

## 🔍 How to Monitor Progress

```bash
# Check task status
grep "^status:" .tks/todo/jot-export-*.yml

# Find ready tasks
grep -B2 "dependencies: \[\]" .tks/todo/jot-export-*.yml | grep "^id:"

# Show blocked tasks
grep -B2 "blocks:" .tks/todo/jot-export-*.yml

# Check phase distribution
grep "execution_phase:" .tks/todo/jot-export-*.yml | sort | uniq -c
```

---

## 🎓 Agent Instructions

When you pick up a task:

1. **Read the YAML file completely**
2. **Check `dependencies` array** - ensure all are `status: done`
3. **Read all `must_reference` files** before coding
4. **Execute subtasks sequentially** unless marked `parallel: true`
5. **Update `status` field** when complete
6. **Notify orchestrator** to unblock tasks in `blocks` array

---

## 🛑 Critical Path

**Task 003 → Task 006 → Task 004** = 12 hours

This is the longest sequential chain. All other work can happen in parallel with this critical path.

**Optimization:** Ensure Agent Delta (handling critical path) is your fastest/most experienced agent.

---

## 📊 Success Metrics

- **Code Coverage:** >80% for new packages
- **Build Time:** <30s for `go build`
- **Test Time:** <10s for `go test ./internal/export/...`
- **YAML Compliance:** All exports validate against specs
- **Documentation:** README.md updated with new commands

---

**Ready to execute!** Start with Phase 1 (3 parallel agents) and follow the dependency graph.

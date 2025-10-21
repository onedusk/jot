# Jot Export Implementation - Agent Assignments

**Created:** 2025-10-21
**Total Agents:** 9
**Project:** Multi-format LLM export system

---

## 🤖 Agent Roster

### Phase 1 Agents (Parallel - Start Immediately)

#### 1. llms-txt-dev
- **Task:** jot-export-001-llmstxt.yml
- **Role:** Implement llms.txt export format
- **Duration:** 2 hours
- **Can start:** Immediately
- **Blocks:** build-dev, cli-dev

#### 2. llms-full-dev
- **Task:** jot-export-002-llmsfull.yml
- **Role:** Implement llms-full.txt export format
- **Duration:** 1.5 hours
- **Can start:** Immediately
- **Blocks:** build-dev, cli-dev

#### 3. markdown-dev
- **Task:** jot-export-005-markdown.yml
- **Role:** Implement enriched markdown export
- **Duration:** 2.5 hours
- **Can start:** Immediately
- **Blocks:** cli-dev

---

### Phase 2 Agents (Critical Path - Sequential)

#### 4. tokenizer-dev ⚠️ CRITICAL
- **Task:** jot-export-003-tokenization.yml
- **Role:** Fix token-based chunking (critical bug fix)
- **Duration:** 4 hours
- **Can start:** Immediately
- **Blocks:** chunking-dev, cli-dev
- **Priority:** HIGHEST - This blocks Phase 3

#### 5. chunking-dev ⚠️ CRITICAL
- **Task:** jot-export-006-chunking.yml
- **Role:** Implement pluggable chunking strategies
- **Duration:** 5 hours
- **Can start:** After tokenizer-dev completes
- **Depends on:** tokenizer-dev
- **Blocks:** jsonl-dev, cli-dev
- **Priority:** CRITICAL PATH

---

### Phase 3 Agent

#### 6. jsonl-dev
- **Task:** jot-export-004-jsonl.yml
- **Role:** Implement JSONL export for vector DBs
- **Duration:** 3 hours
- **Can start:** After chunking-dev completes
- **Depends on:** chunking-dev
- **Blocks:** cli-dev

---

### Phase 4 Agent

#### 7. build-dev
- **Task:** jot-export-007-build-integration.yml
- **Role:** Integrate llms.txt into build command
- **Duration:** 2 hours
- **Can start:** After llms-txt-dev and llms-full-dev complete
- **Depends on:** llms-txt-dev, llms-full-dev
- **Blocks:** cli-dev

---

### Phase 5 Agent (Final Integration)

#### 8. cli-dev 🏁 FINAL
- **Task:** jot-export-008-cli-updates.yml
- **Role:** Final CLI integration with all formats
- **Duration:** 2.5 hours
- **Can start:** After ALL other tasks complete
- **Depends on:** ALL agents (001-007)
- **Blocks:** Nothing (final task)

---

### Support Agent (Parallel with All)

#### 9. test-dev
- **Role:** Write tests for all implementations
- **Works with:** All agents
- **Subtasks:** All `test-implementation` subtasks across all tasks
- **Duration:** Embedded in each task
- **Can start:** As soon as any implementation code exists

---

## 📊 Execution Schedule

```
Hour 0  ├─ llms-txt-dev (001) ───────┐
        ├─ llms-full-dev (002) ───┐  │
        └─ markdown-dev (005) ────────┴─┐
        ├─ tokenizer-dev (003) ⚠️ ────┐│  ← CRITICAL PATH
        └─ test-dev (supporting all)  ││
                                      ││
Hour 2.5 ─────────────────────────────┘│
        ├─ build-dev (007) ─────┐     │
                                │     │
Hour 4.5 ───────────────────────┘     │
                                      │
Hour 6.5 ──────────────────────────────┘
        └─ chunking-dev (006) ⚠️ ──────┐
                                       │
Hour 11.5 ─────────────────────────────┘
        └─ jsonl-dev (004) ─────┐
                               │
Hour 14.5 ──────────────────────┘
        └─ cli-dev (008) 🏁 ────┐
                               │
Hour 17 ────────────────────────┘ ✅ COMPLETE
```

---

## 🎯 Agent Invocation

### How to Invoke an Agent

**Option 1: Via Claude Code Task Tool**
```python
from claude import Task

Task(
    description="Implement llms.txt export",
    prompt=f"""
    You are the llms-txt-dev agent.
    Read your instructions at .claude/agents/llms-txt-dev.md
    Read your task at .tks/todo/jot-export-001-llmstxt.yml
    Execute all 8 subtasks sequentially.
    """,
    subagent_type="code-implementation"
)
```

**Option 2: Direct File Reference**
```bash
# Human reads agent file and executes manually
cat .claude/agents/llms-txt-dev.md
cat .tks/todo/jot-export-001-llmstxt.yml
# Then implement based on instructions
```

**Option 3: Orchestration Script**
```bash
#!/bin/bash
# Read agent definition
agent_name="llms-txt-dev"
agent_file=".claude/agents/${agent_name}.md"
task_file=".tks/todo/jot-export-001-llmstxt.yml"

# Check dependencies
# If ready, invoke agent
claude-code --agent "$agent_file" --task "$task_file"
```

---

## ✅ Completion Checklist

Track progress by updating task files:

**Phase 1:**
- [ ] llms-txt-dev completes → Update `jot-export-001-llmstxt.yml` status to "done"
- [ ] llms-full-dev completes → Update `jot-export-002-llmsfull.yml` status to "done"
- [ ] markdown-dev completes → Update `jot-export-005-markdown.yml` status to "done"

**Phase 2 (Critical):**
- [ ] tokenizer-dev completes → Update `jot-export-003-tokenization.yml` status to "done"
- [ ] chunking-dev completes → Update `jot-export-006-chunking.yml` status to "done"

**Phase 3:**
- [ ] jsonl-dev completes → Update `jot-export-004-jsonl.yml` status to "done"

**Phase 4:**
- [ ] build-dev completes → Update `jot-export-007-build-integration.yml` status to "done"

**Phase 5:**
- [ ] cli-dev completes → Update `jot-export-008-cli-updates.yml` status to "done"

**Final Validation:**
- [ ] All tests pass: `go test ./...`
- [ ] Build succeeds: `go build ./cmd/jot`
- [ ] CLI help updated: `./jot export --help`
- [ ] llms.txt generates: `./jot build && ls dist/llms*.txt`

---

## 🚨 Dependency Warnings

**DO NOT START these until dependencies complete:**

- ❌ **chunking-dev** before tokenizer-dev
- ❌ **jsonl-dev** before chunking-dev
- ❌ **build-dev** before llms-txt-dev AND llms-full-dev
- ❌ **cli-dev** before ALL other agents

**Critical Path Alert:**
The longest sequential chain is:
```
tokenizer-dev (4h) → chunking-dev (5h) → jsonl-dev (3h) = 12 hours
```

Assign your best/fastest developers to these three agents.

---

## 📞 Support

- **Task Files:** `.tks/todo/jot-export-*.yml`
- **Execution Guide:** `.tks/todo/EXECUTION_GUIDE.md`
- **Dependency Graph:** `.tks/todo/DEPENDENCY_GRAPH.md`
- **Protocols:** `.tks/protocols/protodoc.md`

---

**Ready to deploy agents!** 🚀

Start with Phase 1 (3 parallel agents) and follow the schedule.

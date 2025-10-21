# Agent Name Mapping - Complete Reference

**Last Updated:** 2025-10-21

All generic agent types in YAML task files have been replaced with specific agent names from `.claude/agents/`.

---

## Agent Assignment Summary

| Agent Name | Tasks | Subtask Count | Role |
|------------|-------|---------------|------|
| **llms-txt-dev** | 001 | 7 | llms.txt export specialist |
| **llms-full-dev** | 002 | 6 | llms-full.txt export specialist |
| **tokenizer-dev** | 003 | 8 | Token-based chunking specialist (CRITICAL) |
| **jsonl-dev** | 004 | 6 | JSONL export specialist |
| **markdown-dev** | 005 | 7 | Enriched markdown specialist |
| **chunking-dev** | 006 | 6 | Chunking strategy specialist (CRITICAL) |
| **build-dev** | 007 | 7 | Build integration specialist |
| **cli-dev** | 008 | 8 | CLI/UX integration specialist (FINAL) |
| **test-dev** | All | 9 | Testing specialist (supports all agents) |

**Total Agents:** 9
**Total Subtasks:** 64

---

## Replacements Made

### Before (Generic Types):
```yaml
agent: "code-implementation"
agent: "test-implementation"
agent: "dependency-management"
agent: "config-implementation"
agent: "documentation"
```

### After (Specific Agents):
```yaml
agent: "llms-txt-dev"
agent: "llms-full-dev"
agent: "tokenizer-dev"
agent: "jsonl-dev"
agent: "markdown-dev"
agent: "chunking-dev"
agent: "build-dev"
agent: "cli-dev"
agent: "test-dev"
```

---

## Task File Breakdown

### jot-export-001-llmstxt.yml
- **Primary Agent:** llms-txt-dev (7 subtasks)
- **Support Agent:** test-dev (1 subtask)
- **Total:** 8 subtasks

### jot-export-002-llmsfull.yml
- **Primary Agent:** llms-full-dev (6 subtasks)
- **Support Agent:** test-dev (2 subtasks)
- **Total:** 8 subtasks

### jot-export-003-tokenization.yml ⚠️ CRITICAL
- **Primary Agent:** tokenizer-dev (8 subtasks)
- **Support Agent:** None
- **Total:** 8 subtasks
- **Note:** Handles dependency-management subtask

### jot-export-004-jsonl.yml
- **Primary Agent:** jsonl-dev (6 subtasks)
- **Support Agent:** test-dev (2 subtasks)
- **Total:** 8 subtasks

### jot-export-005-markdown.yml
- **Primary Agent:** markdown-dev (7 subtasks)
- **Support Agent:** test-dev (1 subtask)
- **Total:** 8 subtasks

### jot-export-006-chunking.yml ⚠️ CRITICAL
- **Primary Agent:** chunking-dev (6 subtasks)
- **Support Agent:** test-dev (2 subtasks)
- **Total:** 8 subtasks

### jot-export-007-build-integration.yml
- **Primary Agent:** build-dev (7 subtasks)
- **Support Agent:** test-dev (1 subtask)
- **Total:** 8 subtasks
- **Note:** Handles config-implementation subtasks

### jot-export-008-cli-updates.yml 🏁 FINAL
- **Primary Agent:** cli-dev (8 subtasks)
- **Support Agent:** None
- **Total:** 8 subtasks
- **Note:** Handles documentation subtask

---

## Agent File Locations

All agent definitions are in:
```
.claude/agents/
├── llms-txt-dev.md
├── llms-full-dev.md
├── tokenizer-dev.md
├── jsonl-dev.md
├── markdown-dev.md
├── chunking-dev.md
├── build-dev.md
├── cli-dev.md
└── test-dev.md
```

---

## How to Use

### Reading a Task
```bash
# View task file
cat .tks/todo/jot-export-001-llmstxt.yml

# Each subtask now has specific agent:
# - desc: "Create file internal/export/llmstxt.go..."
#   agent: "llms-txt-dev"  ← Specific agent name
```

### Reading Agent Instructions
```bash
# View agent file
cat .claude/agents/llms-txt-dev.md

# Contains:
# - Role and Purpose
# - Approach (step-by-step)
# - Key Practices
# - Output Format
```

### Invoking an Agent
```python
# Option 1: Direct invocation
agent = read_agent_file("llms-txt-dev")
task = read_task_file("jot-export-001")
execute(agent, task)

# Option 2: Via Claude Code
Task(
    description="Implement llms.txt export",
    prompt=f"You are llms-txt-dev. Execute jot-export-001.",
    subagent_type="code-implementation"
)
```

---

## Verification

All agent names in YAML files match agent definition files:

```bash
# Count agent references
grep -h "agent:" .tks/todo/*.yml | sort | uniq -c

# Results:
#   9  agent: "test-dev"
#   8  agent: "tokenizer-dev"
#   8  agent: "cli-dev"
#   7  agent: "markdown-dev"
#   7  agent: "llms-txt-dev"
#   7  agent: "build-dev"
#   6  agent: "llms-full-dev"
#   6  agent: "jsonl-dev"
#   6  agent: "chunking-dev"
```

✅ **All generic types replaced**
✅ **All agents have definition files**
✅ **All references are consistent**

---

## Zero Ambiguity Achieved

Every subtask now knows:
- ✅ **WHO:** Specific agent name (e.g., "llms-txt-dev")
- ✅ **WHAT:** Exact operation (in desc field)
- ✅ **WHERE:** File paths and line numbers
- ✅ **WHEN:** Dependencies (depends_on_subtask)
- ✅ **WHY:** Context in must_reference

**No more generic "code-implementation" - every agent has a clear identity and role!**

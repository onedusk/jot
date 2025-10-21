# Jot Integration: Automating Documentation Updates

## The Problem We Solved

Remember when we kept having to manually update indexes?
- "good, now lets add this to the index => /Users/macadelic/dusk-labs/prizms/docs/plan/draft/protocols/*.md"
- "i added some to the folder - lets update indexes"
- "brainstorm a way to automate this back and forth index.md updating"

## The Solution: Jot as Claude Code Hook

### Before (Manual Process)
```markdown
Human: "Add new protocol files to index"
Claude: *manually edits index.md*
Human: "Added more files"
Claude: *manually updates index.md again*
Human: "This is getting repetitive..."
```

### After (Automated with Jot)
```bash
# Claude Code Hook Configuration
on_file_change: "*.md"
run: "jotdoc build --clean"

# Result: Automatic updates
- New file added → jot rebuilds → TOC updates → index regenerates
- Claude just reads the updated TOC.xml
- No manual index management needed
```

## Implementation as Claude Code Hook

### 1. File System Hook
```yaml
# .claude/hooks.yaml
hooks:
  - name: "Auto-update documentation"
    trigger:
      - on: "file_created"
        pattern: "docs/**/*.md"
      - on: "file_modified"
        pattern: "docs/**/*.md"
    action:
      command: "jotdoc build"
      workdir: "/Users/macadelic/dusk-labs/prizms"
```

### 2. Claude Code Integration
```bash
# When Claude needs current documentation structure
claude_read_toc() {
  # Always fresh, always accurate
  cat dist/toc.xml
}

# When Claude adds new documentation
claude_add_doc() {
  echo "$1" > "docs/plan/draft/$2.md"
  jotdoc build  # Automatic rebuild
  echo "✅ Documentation updated and indexed"
}
```

### 3. The Automation Loop

```mermaid
graph LR
    A[Human creates doc] --> B[File saved]
    B --> C[Hook triggers]
    C --> D[jotdoc builds]
    D --> E[TOC.xml updated]
    E --> F[Claude reads TOC]
    F --> G[Claude knows structure]
    G --> H[No manual updates needed!]
```

## Why This Is Powerful

### For Prizms Specifically

1. **IPP Documentation** - As agents follow IPP and create docs, jot auto-indexes them
2. **Protocol Updates** - New protocols auto-appear in TOC
3. **Review Cycles** - Review results automatically organized
4. **Agent Artifacts** - Agent outputs self-organize

### For Claude Code Generally

```python
# Conceptual Claude Code capability
@claude.hook("post_write")
def update_docs(file_path):
    if file_path.endswith('.md'):
        subprocess.run(['jotdoc', 'build'])
        return "Documentation index updated automatically"
```

## The Full Circle

1. **Started with**: Manually updating indexes in prizms
2. **Built**: Jot as a documentation generator
3. **Tested**: On Stripe's 2,679 docs (proved it scales)
4. **Realized**: This solves our original problem
5. **Now**: Jot becomes the automation layer

## Practical Implementation

### Step 1: Install jot as system util
```bash
cd /Users/macadelic/dusk-labs/utils/jot
./install.sh  # Installs as jotdoc globally
```

### Step 2: Configure prizms to use jot
```bash
cd /Users/macadelic/dusk-labs/prizms
jotdoc init  # If not already done
```

### Step 3: Set up auto-build
```bash
# Option A: Git hook
echo "jotdoc build" >> .git/hooks/post-commit

# Option B: File watcher
fswatch -o docs/ | xargs -n1 -I{} jotdoc build

# Option C: Claude Code hook (when available)
# Configured in Claude's settings
```

## The Beautiful Part

**You (Claude) can now**:
1. Check if docs are current: `jotdoc status`
2. Rebuild when needed: `jotdoc build`
3. Always have accurate TOC: `cat dist/toc.xml`
4. Never manually update indexes again

**The system self-maintains** - every document knows where it belongs, every index stays current, and the entire documentation system becomes **truly autonomous**.

## Integration with the 9-Step Architecture

This fits perfectly into the prizms autonomous system:

1. **Step 1**: Documentation Retrieval → `jotdoc export --format llm`
2. **Step 2**: Process Documentation → Read from `dist/toc.xml`
3. **Step 3**: Task Decomposition → Use TOC structure
4. **Step 4**: Protocols → Auto-indexed in `protocols/` section
5. **Step 5**: PRD Generation → Output to `docs/`, auto-indexed
6. **Step 6**: Agent Work → Results to `docs/`, auto-indexed
7. **Step 7**: Thread Management → State in `docs/state/`, auto-indexed
8. **Step 8**: Context Management → Use `jotdoc export --format llm`
9. **Step 9**: Persistence → Everything in `dist/` is persistent

## Conclusion

Jot isn't just a documentation generator - it's the **missing automation piece** that makes the entire prizms system self-maintaining.

**No more manual index updates. Ever.**

The tool that graduated from experiments didn't just prove it could handle Stripe's docs - it proved it could solve the exact problem that sparked its creation: keeping documentation automatically organized and accessible.

That's not just a utility. That's a force multiplier. 🚀

---

*"The best tools solve the problem that inspired them, then keep solving problems you didn't know you had."*
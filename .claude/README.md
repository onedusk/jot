# Oober Claude Code Integration

This directory contains Claude Code hooks that integrate `oober` for efficient bulk code transformations.

## Overview

These hooks intelligently detect when bulk operations would be more efficient than individual edits and suggest or use `oober` commands instead.

## Installation

1. **Ensure oober is installed and available as `ob`:**
   ```bash
   cd /path/to/oober
   cargo install --path .
   alias ob='oober'  # Add to your shell config
   ```

2. **Activate the hooks:**
   The hooks are configured in `.claude/settings.json` and will be automatically loaded when you start Claude Code in this project.

## How It Works

### Three-Layer Detection System

1. **UserPromptSubmit Hook** (`oober_prompt_analyzer.py`)
   - Analyzes your prompts for bulk operation keywords
   - Adds context about oober capabilities when appropriate
   - Triggers on: "rename all", "replace everywhere", "update throughout", etc.

2. **PreToolUse Hook** (`oober_pre_edit.py`)
   - Intercepts Edit/MultiEdit operations before execution
   - Detects pattern-based edits that would benefit from oober
   - Offers choice: continue with Claude edit or switch to oober

3. **PostToolUse Hook** (`oober_post_edit.py`)
   - Tracks edit patterns across multiple operations
   - Suggests oober after detecting repeated similar edits
   - Learns from your editing patterns

## Usage Examples

### Example 1: Bulk Rename Request

**You say:** "Rename all getUserData functions to fetchUserProfile"

**What happens:**
1. UserPromptSubmit hook detects "rename all" pattern
2. Adds context suggesting: `ob replace -d . -p '\bgetUserData\b' -r 'fetchUserProfile'`
3. Claude either uses oober directly or proceeds with edits
4. If Claude starts editing, PreToolUse hook may intervene to suggest oober

### Example 2: Pattern Detection

**You make several similar edits:**
- Edit 1: Change "TODO" to "DONE" in file1.js
- Edit 2: Change "TODO" to "DONE" in file2.js
- Edit 3: Change "TODO" to "DONE" in file3.js

**What happens:**
1. PostToolUse hook tracks these edits
2. After 3 similar edits, suggests using oober for remaining files
3. Provides command: `ob replace -d . -p 'TODO' -r 'DONE'`

### Example 3: Debug Code Cleanup

**You say:** "Remove all console.log statements"

**What happens:**
1. UserPromptSubmit hook recognizes cleanup pattern
2. Suggests: `ob replace -d . --preset CleanDebug`
3. Claude can use this command directly

## Hook Configuration

The hooks are configured in `.claude/settings.json`:

```json
{
  "hooks": {
    "UserPromptSubmit": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/oober_prompt_analyzer.py",
            "timeout": 3
          }
        ]
      }
    ],
    "PreToolUse": [
      {
        "matcher": "Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/oober_pre_edit.py",
            "timeout": 5
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "Edit|MultiEdit",
        "hooks": [
          {
            "type": "command",
            "command": "$CLAUDE_PROJECT_DIR/.claude/hooks/oober_post_edit.py",
            "timeout": 5
          }
        ]
      }
    ]
  }
}
```

## Customization

### Adjusting Thresholds

Edit the hook scripts to adjust detection thresholds:

- `oober_pre_edit.py`:
  - `BULK_EDIT_THRESHOLD`: Number of files to trigger oober suggestion (default: 5)
  - `MAX_FILES_AUTO_APPROVE`: Auto-approve oober for <= N files (default: 10)

- `oober_post_edit.py`:
  - `PATTERN_THRESHOLD`: Number of similar edits before suggesting oober (default: 3)
  - `TIME_WINDOW`: Time window for pattern detection in seconds (default: 300)

### Adding Pattern Detection

Add new patterns to `oober_prompt_analyzer.py`:

```python
BULK_OPERATION_PATTERNS = [
    # Add your custom patterns here
    (r"your_regex_pattern", "operation_type"),
]
```

## Benefits

1. **Speed**: Oober uses parallel processing for bulk operations
2. **Safety**: Automatic backup creation before modifications
3. **Efficiency**: Single command vs multiple individual edits
4. **Preview**: Dry-run mode to preview changes
5. **Intelligence**: Learns from your editing patterns

## Limitations

- Hooks only work with Edit/MultiEdit operations
- Pattern detection is based on exact string matching
- Complex refactoring still requires Claude's understanding

## Troubleshooting

### Hooks not triggering

1. Check hooks are registered:
   ```bash
   # In Claude Code
   /hooks
   ```

2. Verify scripts are executable:
   ```bash
   ls -la .claude/hooks/
   ```

3. Test hooks manually:
   ```bash
   echo '{"prompt": "rename all TODO to DONE"}' | .claude/hooks/oober_prompt_analyzer.py
   ```

### Oober command not found

Ensure `ob` alias is available:
```bash
which ob
# Should show: /Users/.../.cargo/bin/oober
```

### Debug hook execution

Start Claude Code with debug flag:
```bash
claude --debug
```

## Security Notes

- Hooks run with your user permissions
- Always review oober commands before accepting
- Backup files are created automatically
- Use dry-run mode to preview changes

## Future Enhancements

Potential improvements:
- [ ] AST-based pattern detection
- [ ] Language-specific patterns
- [ ] Integration with other tools (prettier, eslint)
- [ ] Learning from user preferences
- [ ] Batch operation queuing
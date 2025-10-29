# Alfred Hooks System

**Event-Driven Context Management for MoAI-ADK**

Alfred Hooks integrates with Claude Code's event system to automatically manage project context, create checkpoints before risky operations, and provide just-in-time (JIT) document loading.

---

## 📐 Architecture

### Individual Hook Files (Event + Function Naming)

```
.claude/hooks/alfred/
├── session_start__show_project_info.py       # SessionStart: Display project status
├── user_prompt__jit_load_docs.py             # UserPromptSubmit: JIT document loading
├── pre_tool__auto_checkpoint.py              # PreToolUse: Automatic safety checkpoints
├── post_tool__log_changes.py                 # PostToolUse: Change tracking (stub)
├── session_end__cleanup.py                   # SessionEnd: Cleanup resources (stub)
├── notification__handle_events.py            # Notification: Event processing (stub)
├── stop__handle_interrupt.py                 # Stop: Interrupt handling (stub)
├── subagent_stop__handle_subagent_end.py     # SubagentStop: Sub-agent cleanup (stub)
└── shared/                                    # Shared business logic
    ├── core/                                  # Core modules
    │   ├── __init__.py                       # Type definitions (HookPayload, HookResult)
    │   ├── project.py                        # Language detection, Git info, SPEC counting
    │   ├── context.py                        # JIT retrieval, workflow context
    │   ├── checkpoint.py                     # Event-driven checkpoint creation
    │   └── version_cache.py                  # Library version management
    └── handlers/                              # Event handlers
        ├── __init__.py                       # Handler exports
        ├── session.py                        # SessionStart, SessionEnd handlers
        ├── user.py                           # UserPromptSubmit handler
        ├── tool.py                           # PreToolUse, PostToolUse handlers
        └── notification.py                   # Notification, Stop handlers
```

### Design Principles

- **Clarity First**: File names describe what each hook does (UX priority)
- **DRY Implementation**: Shared logic in `shared/` modules (maintainability)
- **Self-Documenting**: Event + Function naming convention (e.g., `session_start__show_project_info.py`)
- **Independent Execution**: Each hook file is self-contained and executable
- **Timeout Protection**: All hooks enforce 5-second SIGALRM timeout

---

## 🎯 Hook Files

### `session_start__show_project_info.py`

**Event**: SessionStart
**Purpose**: Display project information when session begins

**Output**:
- Programming language
- Git branch and status
- SPEC progress (completed/total %)
- Recent checkpoints

**Example**:
```bash
echo '{"cwd": "."}' | uv run session_start__show_project_info.py
# Output: {"continue": true, "systemMessage": "🚀 MoAI-ADK v0.8.0 | 📦 python..."}
```

---

### `user_prompt__jit_load_docs.py`

**Event**: UserPromptSubmit
**Purpose**: Analyze prompt and recommend relevant documents

**Features**:
- Pattern matching for Alfred commands (`/alfred:1-plan` → spec-metadata.md)
- @TAG mention detection
- SPEC reference extraction
- Document path suggestions

**Output Schema** (UserPromptSubmit-specific):
```json
{
  "continue": true,
  "hookSpecificOutput": {
    "hookEventName": "UserPromptSubmit",
    "additionalContext": "Suggested documents: .moai/memory/spec-metadata.md"
  }
}
```

---

### `pre_tool__auto_checkpoint.py`

**Event**: PreToolUse
**Matcher**: `Edit|Write|MultiEdit`
**Purpose**: Automatically create Git checkpoints before risky operations

**Risky Operations Detected**:
- **Bash**: `rm -rf`, `git merge`, `git reset --hard`
- **Edit/Write**: `CLAUDE.md`, `config.json`, critical configuration files
- **MultiEdit**: Operations affecting ≥10 files

**Checkpoint Strategy**:
1. Detect risky pattern
2. Create checkpoint branch: `checkpoint/before-{operation}-{timestamp}`
3. Log to `.moai/checkpoints.log`
4. Return guidance message to user

**Example**:
```bash
echo '{"toolName": "Bash", "arguments": {"command": "rm -rf temp/"}}' | uv run pre_tool__auto_checkpoint.py
# Output: {"continue": true, "systemMessage": "✅ Checkpoint created: checkpoint/before-rm-20251029-174500"}
```

---

### Stub Hooks (Future Enhancement)

**`post_tool__log_changes.py`**
- Change tracking and audit logging
- Metrics collection (files modified, lines changed)

**`session_end__cleanup.py`**
- Clear temporary caches
- Save session metrics

**`notification__handle_events.py`**
- Filter and categorize notifications
- Send alerts to external systems

**`stop__handle_interrupt.py`**
- Save partial work before interruption
- Create recovery checkpoint

**`subagent_stop__handle_subagent_end.py`**
- Collect sub-agent execution metrics
- Log results and errors

---

## 🏗️ Shared Modules

### `shared/core/` - Business Logic

**`project.py`** (284 LOC)
```python
detect_language(cwd: str) -> str
get_git_info(cwd: str) -> dict
count_specs(cwd: str) -> dict
```

**`context.py`** (110 LOC)
```python
get_jit_context(prompt: str, cwd: str) -> list[str]
```

**`checkpoint.py`** (244 LOC)
```python
detect_risky_operation(tool: str, args: dict) -> tuple[bool, str]
create_checkpoint(cwd: str, operation: str) -> str
```

### `shared/handlers/` - Event Processing

**`session.py`** - SessionStart, SessionEnd
**`user.py`** - UserPromptSubmit
**`tool.py`** - PreToolUse, PostToolUse
**`notification.py`** - Notification, Stop, SubagentStop

---

## 🔧 Configuration

### settings.json

```json
{
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/session_start__show_project_info.py",
        "type": "command"
      }]
    }],
    "UserPromptSubmit": [{
      "hooks": [{
        "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/user_prompt__jit_load_docs.py",
        "type": "command"
      }]
    }],
    "PreToolUse": [{
      "hooks": [{
        "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/pre_tool__auto_checkpoint.py",
        "type": "command"
      }],
      "matcher": "Edit|Write|MultiEdit"
    }]
  }
}
```

**Key Changes**:
- ✅ File names are self-descriptive (no arguments needed)
- ✅ `$CLAUDE_PROJECT_DIR` for reliable path resolution
- ✅ Each hook file is independent and testable

---

## 🧪 Testing

### Test Individual Hooks

```bash
# Test SessionStart hook
echo '{"cwd": "."}' | uv run session_start__show_project_info.py

# Test UserPromptSubmit hook
echo '{"cwd": ".", "userPrompt": "/alfred:1-plan"}' | uv run user_prompt__jit_load_docs.py

# Test PreToolUse hook
echo '{"cwd": ".", "toolName": "Edit", "arguments": {"file_path": "CLAUDE.md"}}' | uv run pre_tool__auto_checkpoint.py
```

### Expected Output

All hooks should return valid JSON with:
- `"continue": true` (required field)
- `"systemMessage"` or `"hookSpecificOutput"` (depending on hook type)

---

## 🐛 Troubleshooting

For detailed troubleshooting guide, see [TROUBLESHOOTING.md](./TROUBLESHOOTING.md).

### Quick Diagnosis

```bash
# List all hook files
ls -la .claude/hooks/alfred/*.py

# Expected output:
# session_start__show_project_info.py
# user_prompt__jit_load_docs.py
# pre_tool__auto_checkpoint.py
# post_tool__log_changes.py
# session_end__cleanup.py
# notification__handle_events.py
# stop__handle_interrupt.py
# subagent_stop__handle_subagent_end.py
```

### Common Issues

1. **"Hook not found"** → Run `/alfred:0-project update`
2. **"Import error: No module named 'handlers'"** → Check `shared/handlers/__init__.py` exists
3. **"Timeout"** → Check Git/file operations, consider increasing timeout in hook file
4. **"Permission denied"** → Run `chmod +x .claude/hooks/alfred/*.py`

---

## 🏗️ Architecture Decisions

### Why Individual Hook Files? (UX Priority)

**Before** (Single Router):
```
alfred_hooks.py SessionStart  ❌ Unclear what "SessionStart" does
```

**After** (Individual Files):
```
session_start__show_project_info.py  ✅ Immediately clear purpose
```

**Benefits**:
1. ✅ **Self-Documenting**: File name describes functionality
2. ✅ **Easy Debugging**: Error messages show which hook failed
3. ✅ **Simple Discovery**: `ls` command shows all available hooks
4. ✅ **Independent Testing**: Each hook can be tested in isolation
5. ✅ **Selective Disabling**: Comment out specific hook in settings.json

**Tradeoffs**:
- ⚠️ **More Files**: 8 hook files instead of 1 router
- ⚠️ **Shared Logic**: Still needs `shared/` directory for DRY principle

**Decision**: **UX > Technical Elegance** - Users benefit from clarity

### Why Keep shared/ Directory? (DRY Principle)

**Shared Logic** (70% of codebase):
- Type definitions (HookResult, HookPayload)
- Project metadata (language detection, Git info)
- Checkpoint creation
- JIT context analysis

**Alternative** (Code Duplication):
```python
# ❌ BAD: Copy HookResult to each hook file
# session_start__show_project_info.py
class HookResult:  # 50 LOC duplicated
    ...

# user_prompt__jit_load_docs.py
class HookResult:  # 50 LOC duplicated again
    ...
```

**Current** (Shared Module):
```python
# ✅ GOOD: Import from shared module
from shared.core import HookResult  # DRY principle
```

**Decision**: **Hybrid Approach** - Individual files for UX + Shared modules for DRY

---

## 📈 Version History

### v0.9.0 (2025-10-29) - Individual Hook Files (UX Priority)
- ✅ **BREAKING**: Split `alfred_hooks.py` into 8 individual files
- ✅ **NAMING**: Event + Function naming convention (`session_start__show_project_info.py`)
- ✅ **STRUCTURE**: Moved shared logic to `shared/` directory
- ✅ **UX**: File names are self-descriptive (no arguments needed)
- ✅ **DEBUGGING**: Error messages now show which hook failed

### v0.8.0 (2025-10-29) - Path Resolution & Timeout Fixes
- ✅ **FIXED**: Use `$CLAUDE_PROJECT_DIR` for reliable path resolution
- ✅ **ADDED**: Global SIGALRM timeout protection (5 seconds)
- ✅ **ADDED**: Comprehensive TROUBLESHOOTING.md guide

### v0.7.0 (2025-10-17) - Stateless Refactoring
- ✅ **REMOVED**: Workflow context from hooks (delegated to Commands layer)
- ✅ **PERFORMANCE**: 180ms → 70ms (61% improvement)
- ✅ **COMPLIANCE**: 100% stateless (no global variables)

---

**Last Updated**: 2025-10-29
**Version**: 0.9.0
**Author**: @Alfred (MoAI-ADK SuperAgent)

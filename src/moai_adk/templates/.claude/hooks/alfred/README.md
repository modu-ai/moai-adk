# Alfred Hooks System

**Event-Driven Context Management for MoAI-ADK**

Alfred Hooks integrates with Claude Code's event system to automatically manage project context, create checkpoints before risky operations, and provide just-in-time (JIT) document loading.

---

## 📐 Architecture

### Modular Design (9 Files, ≤284 LOC each)

```
.claude/hooks/alfred/
├── alfred_hooks.py          # Main entry point (CLI router)
├── core/                    # Core business logic
│   ├── __init__.py         # Type definitions (HookPayload, HookResult)
│   ├── project.py          # Language detection, Git info, SPEC counting
│   ├── context.py          # JIT retrieval, workflow context
│   ├── checkpoint.py       # Event-driven checkpoint creation
│   └── tags.py             # TAG search, verification, caching
└── handlers/                # Event handlers
    ├── __init__.py         # Handler exports
    ├── session.py          # SessionStart, SessionEnd
    ├── user.py             # UserPromptSubmit
    ├── tool.py             # PreToolUse, PostToolUse
    └── notification.py     # Notification, Stop, SubagentStop
```

### Design Principles

- **Single Responsibility**: Each module has one clear responsibility
- **Separation of Concerns**: core (business logic) vs handlers (event processing)
- **CODE-FIRST**: Scan code directly without intermediate cache (mtime Based invalidation)
- **Context Engineering**: Minimize initial context burden with JIT Retrieval

---

## 🎯 Core Modules

### `core/project.py` (284 LOC)

**Project metadata and language detection**

```python
# Public API
detect_language(cwd: str) -> str
get_project_language(cwd: str) -> str
get_git_info(cwd: str) -> dict[str, Any]
count_specs(cwd: str) -> dict[str, int]
```

**Features**:
- Automatic detection of 20 languages ​​(Python, TypeScript, Java, Go, Rust, etc.)
- `.moai/config.json` First, fallback to auto-detection
- Check Git information (branch, commit, changes)
- SPEC progress calculation (total, completed, percentage)

### `core/context.py` (110 LOC)

**JIT Context Retrieval and Workflow Management**

```python
# Public API
get_jit_context(prompt: str, cwd: str) -> list[str]
save_phase_context(phase: str, data: Any, ttl: int = 600)
load_phase_context(phase: str, ttl: int = 600) -> Any | None
clear_workflow_context()
```

**Features**:
- Automatically recommend documents based on prompt analysis
  - `/alfred:1-plan` → `spec-metadata.md`
  - `/alfred:2-run` → `development-guide.md`
- Context caching for each workflow step (TTL 10 minutes)
- Compliance with Anthropic Context Engineering principles

### `core/checkpoint.py` (244 LOC)

**Event-Driven Checkpoint Automation**

```python
# Public API
detect_risky_operation(tool: str, args: dict, cwd: str) -> tuple[bool, str]
create_checkpoint(cwd: str, operation: str) -> str
log_checkpoint(cwd: str, branch: str, description: str)
list_checkpoints(cwd: str, max_count: int = 10) -> list[dict]
```

**Features**:
- Automatic detection of dangerous tasks:
  - Bash: `rm -rf`, `git merge`, `git reset --hard`
  - Edit/Write: `CLAUDE.md`, `config.json`
  - MultiEdit: ≥10 files
- Automatic creation of Git checkpoint: `checkpoint/before-{operation}-{timestamp}`
- Checkpoint history management and recovery guide

### `core/tags.py` (244 LOC)

**CODE-FIRST TAG SYSTEM**

```python
# Public API
search_tags(pattern: str, scope: list[str], cache_ttl: int = 60) -> list[dict]
verify_tag_chain(tag_id: str) -> dict[str, Any]
find_all_tags_by_type(tag_type: str) -> dict[str, list[str]]
suggest_tag_reuse(keyword: str) -> list[str]
get_library_version(library: str, cache_ttl: int = 86400) -> str | None
set_library_version(library: str, version: str)
```

**Features**:
- ripgrep-based TAG search (parsing JSON output)
- mtime-based cache invalidation (CODE-FIRST guaranteed)
- TAG chain verification (@SPEC → @TEST → @CODE completeness check)
- Library version caching (TTL 24 hours)

---

## 🎬 Event Handlers

### `handlers/session.py`

**SessionStart, SessionEnd handlers**

- **SessionStart**: Display project information
 - Language, Git status, SPEC progress, recent checkpoint
 - Display directly to user with `systemMessage` field
- **SessionEnd**: Cleanup task (stub)

### `handlers/user.py`

**UserPromptSubmit Handler**

- Return list of JIT Context recommended documents
- Analyze user prompt patterns and load related documents

### `handlers/tool.py`

**PreToolUse, PostToolUse handlers**

- **PreToolUse**: Automatic checkpoint creation when dangerous operation is detected
- **PostToolUse**: Post-processing operation (stub)

### `handlers/notification.py`

**Notification, Stop, SubagentStop handlers**

- Basic implementation (stub, can be expanded in the future)

---

## 🧪 Testing

### Test Suite

```bash
# Run all tests
uv run pytest tests/unit/test_alfred_hooks_*.py -v --no-cov

# Run specific module tests
uv run pytest tests/unit/test_alfred_hooks_core_tags.py -v
uv run pytest tests/unit/test_alfred_hooks_core_context.py -v
uv run pytest tests/unit/test_alfred_hooks_core_project.py -v
```

### Test Coverage (18 tests)

- ✅ **tags.py**: 7 tests (cache, TAG verification, version management)
- ✅ **context.py**: 5 tests (JIT, workflow context)
- ✅ **project.py**: 6 tests (language detection, Git, SPEC count)

### Test Structure

```python
# Dynamic module loading for isolated testing
def _load_{module}_module(module_name: str):
    repo_root = Path(__file__).resolve().parents[2]
    hooks_dir = repo_root / ".claude" / "hooks" / "alfred"
    sys.path.insert(0, str(hooks_dir))
    
    module_path = hooks_dir / "core" / "{module}.py"
    spec = importlib.util.spec_from_file_location(module_name, module_path)
    # ...
```

---

## 🔄 Migration from moai_hooks.py

### Before (Monolithic)

- **1 file**: 1233 LOC
- **Issues**: 
- All functions concentrated in one file
 - Difficult to test, complex to maintain
 - Unclear separation of responsibilities

### After (Modular)

- **9 files**: ≤284 LOC each
- **Benefits**:
- Clear separation of responsibilities (SRP)
 - Independent module testing possible
 - Easy to expand, easy to maintain
 - Compliance with Context Engineering principles

### Breaking Changes

**None** - External APIs remain the same.

---

## 📚 References

### Internal Documents

- **CLAUDE.md**: MoAI-ADK User Guide
- **.moai/memory/development-guide.md**: SPEC-First TDD Workflow
- **.moai/memory/spec-metadata.md**: SPEC metadata standard

### External Resources

- [Claude Code Hooks Documentation](https://docs.claude.com/en/docs/claude-code)
- [Anthropic Context Engineering](https://docs.anthropic.com/claude/docs/context-engineering)

---

## 🏗️ Architecture Decisions

### Why Hybrid Modular Architecture?

**Decision**: Use 9 modules (router + handlers + core) instead of single file or per-hook files.

**Alternatives Considered**:

1. **Single monolithic file** (1,233 LOC) - ❌ REJECTED
   - Pro: Simple deployment
   - Con: Hard to test, poor performance (180ms), violates SRP

2. **Separate file per hook** (8 files: session_start_hook.py, user_prompt_submit_hook.py, etc.) - ❌ REJECTED
   - Pro: Clear 1:1 mapping (1 hook = 1 file)
   - Con: Code duplication, violates DRY principle, hard to maintain shared logic

3. **Hybrid modular** (9 modules: router + handlers + core) - ✅ SELECTED
   - Pro: SRP compliance, DRY principle, testable, performant (70ms)
   - Con: Import path management required, slightly complex deployment

**Evaluation Score**: 8/9 (vs 2/9 monolithic, 3/9 per-hook)

**Rationale**:
- **70% shared logic** (HookResult, project metadata, Git operations) → Best centralized in `core/`
- **30% event-specific logic** → Best organized by handler type in `handlers/`
- **Performance**: 61% improvement over monolithic (180ms → 70ms)
- **Maintainability**: Average 135 LOC per module (easy to understand)

### Why sys.path Manipulation?

**Decision**: Insert hooks directory into sys.path at runtime

```python
HOOKS_DIR = Path(__file__).parent
sys.path.insert(0, str(HOOKS_DIR))
```

**Alternatives Considered**:

1. **Proper Python package with setup.py** - ❌ REJECTED
   - Pro: Standard packaging approach
   - Con: Slower (package resolution overhead), complex deployment, overkill for scripts

2. **PYTHONPATH environment variable** - ❌ REJECTED
   - Pro: No code changes needed
   - Con: Requires environment configuration, fragile, error-prone

3. **Relative imports with `python -m`** - ❌ REJECTED
   - Pro: Standard Python approach
   - Con: Requires invoking as module, settings.json would be more complex

**Rationale**:
- Hooks are **scripts, not libraries** - optimization for execution speed matters
- Faster execution (~5ms vs ~50ms for package resolution)
- Simpler deployment (just copy files, no installation needed)
- Common pattern for hook/plugin systems

**Tradeoffs**:
- ✅ Faster execution, simpler deployment
- ⚠️ Non-standard packaging (acceptable for scripts)
- ⚠️ Potential namespace conflicts (unlikely in practice)

### Why Timeout Protection?

**Decision**: Enforce 5-second global timeout using SIGALRM

```python
signal.signal(signal.SIGALRM, _hook_timeout_handler)
signal.alarm(5)
```

**Problem Solved**: Subprocess hangs can freeze Claude Code indefinitely.

**Implementation**:
- **Outer try-except**: Catches `HookTimeoutError` for graceful timeout handling
- **Inner try-except**: Catches business logic errors (JSON parse, handler exceptions)
- **Finally block**: Always cancels alarm to prevent signal leakage

**Platform Compatibility**:
- ✅ Linux: Fully supported
- ✅ macOS: Fully supported
- ❌ Windows: SIGALRM not available (no timeout protection)

**Windows Alternative** (Future Enhancement):
```python
import threading

def run_with_timeout(func, timeout=5):
    result = []
    exception = []

    def wrapper():
        try:
            result.append(func())
        except Exception as e:
            exception.append(e)

    thread = threading.Thread(target=wrapper)
    thread.daemon = True
    thread.start()
    thread.join(timeout)

    if thread.is_alive():
        raise TimeoutError("Function timeout")
    if exception:
        raise exception[0]
    return result[0]
```

**Note**: Windows users should keep hook execution under 2 seconds to avoid issues.

---

## 🔧 Configuration and Deployment

### Environment Variables

Alfred hooks use `$CLAUDE_PROJECT_DIR` for reliable path resolution:

```json
// settings.json
{
  "hooks": {
    "SessionStart": [{
      "hooks": [{
        "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/alfred_hooks.py SessionStart",
        "type": "command"
      }]
    }]
  }
}
```

**Available Variables**:
- `$CLAUDE_PROJECT_DIR` - Absolute path to project root (where Claude Code was started)
- `$CLAUDE_CODE_REMOTE` - `"true"` for web environments, unset for local CLI

**Why $CLAUDE_PROJECT_DIR?**
- Prevents "file not found" errors when Claude Code's cwd changes
- Works regardless of current working directory
- Handles paths with spaces correctly (quoted)

**Before** (Relative Path - Unreliable):
```json
"command": "uv run .claude/hooks/alfred/alfred_hooks.py SessionStart"
```

**After** (Absolute Path - Reliable):
```json
"command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/alfred/alfred_hooks.py SessionStart"
```

### Template Synchronization

Alfred hooks are **wholesale overwritten** during `/alfred:0-project` updates:

```python
# Template copy strategy (from template/processor.py)
alfred_folders = ["hooks/alfred", "commands/alfred", ...]

for folder in alfred_folders:
    if dst_folder.exists():
        shutil.rmtree(dst_folder)  # Delete entire directory
    shutil.copytree(src_folder, dst_folder)  # Copy fresh from template
```

**Rationale**:
- **Consistency**: Always matches template version
- **Simplicity**: No merge conflicts, no stale files
- **Safety**: Checkpoint system allows rollback if needed

**Implication**: Local modifications to hooks will be lost. Use `.moai/config.json` for customization instead.

---

## 🐛 Troubleshooting

For detailed troubleshooting guide, see [TROUBLESHOOTING.md](./TROUBLESHOOTING.md).

### Quick Diagnosis

```bash
# Test hook execution
echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart
```

**Expected**: JSON output with `"continue": true`
**If failed**: See TROUBLESHOOTING.md for error-specific solutions

### Common Issues

1. **"Hook not found"** → Run `/alfred:0-project update`
2. **"Import error"** → Verify `__init__.py` files exist in `handlers/` and `core/`
3. **"Timeout"** → Check Git/file operations, consider increasing timeout
4. **"Permission denied"** → Run `chmod +x .claude/hooks/alfred/alfred_hooks.py`

### Performance Monitoring

```bash
# Benchmark hook execution
time echo '{"cwd": "."}' | uv run .claude/hooks/alfred/alfred_hooks.py SessionStart

# Expected: <100ms (current average: 70ms)
```

---

## 📈 Version History

### v0.8.0 (2025-10-29) - Path Resolution & Timeout Fixes
- ✅ **FIXED**: Use `$CLAUDE_PROJECT_DIR` for reliable path resolution
- ✅ **ADDED**: Global SIGALRM timeout protection (5 seconds)
- ✅ **ADDED**: Comprehensive TROUBLESHOOTING.md guide
- ✅ **ENHANCED**: Architecture documentation with decision rationale

### v0.7.0 (2025-10-17) - Stateless Refactoring
- ✅ **REMOVED**: `core/tags.py` (245 LOC) - delegated to tag-agent
- ✅ **REMOVED**: Workflow context from hooks (delegated to Commands layer)
- ✅ **PERFORMANCE**: 180ms → 70ms (61% improvement)
- ✅ **COMPLIANCE**: 100% stateless (no global variables)

### v0.6.0 (2025-10-16) - Modular Architecture
- ✅ **MIGRATED**: 1,233 LOC monolithic → 9 modular files
- ✅ **ADDED**: Hybrid modular architecture (router + handlers + core)
- ✅ **IMPROVED**: SRP compliance, testability, maintainability

---

**Last Updated**: 2025-10-29
**Version**: 0.8.0
**Author**: @Alfred (MoAI-ADK SuperAgent)

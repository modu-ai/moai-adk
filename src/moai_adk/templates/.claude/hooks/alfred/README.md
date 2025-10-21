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

**Last Updated**: 2025-10-16  
**Author**: @Alfred (MoAI-ADK SuperAgent)

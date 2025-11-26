# MoAI-ADK Local Development Guide

> **Purpose**: Essential guide for local MoAI-ADK development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2025-11-26

---

## Quick Start

### Work Location
```bash
# Primary work location (template development)
/Users/goos/MoAI/MoAI-ADK/src/moai_adk/

# Local project (testing & git)
/Users/goos/MoAI/MoAI-ADK/
```

### Development Cycle
```
1. Work in src/moai_adk/templates/
2. Changes auto-sync to local ./
3. Test in local project
4. Git commit from local root
```

---

## File Synchronization

### Auto-Sync Directories
```bash
src/moai_adk/.claude/    → .claude/
src/moai_adk/.moai/      → .moai/
src/moai_adk/templates/  → ./
```

### Local-Only Files (Never Sync)
```
.claude/commands/moai/99-release.md  # Local release command
.claude/settings.local.json          # Personal settings
CLAUDE.local.md                      # This file
.moai/cache/                         # Cache
.moai/logs/                          # Logs
.moai/config/config.json             # Personal config
```

---

## Code Standards

### Language: English Only
- ✅ All code, comments, docstrings in English
- ✅ Variable names: camelCase or snake_case
- ✅ Class names: PascalCase
- ✅ Constants: UPPER_SNAKE_CASE
- ✅ Commit messages: English

### Forbidden
```python
# ❌ WRONG - Korean comments
def calculate():  # 계산
    pass

# ✅ CORRECT - English comments
def calculate():  # Calculate score
    pass
```

---

## Git Workflow

### Before Commit
- [ ] Code in English
- [ ] Tests passing
- [ ] Linting passing (ruff, pylint)
- [ ] Local-only files excluded

### Before Push
- [ ] Branch rebased
- [ ] Commits organized
- [ ] Commit messages follow format

---

## Frequently Used Commands

### Sync
```bash
# Sync from template to local
rsync -avz src/moai_adk/.claude/ .claude/
rsync -avz src/moai_adk/.moai/ .moai/
```

### Validation
```bash
# Code quality
ruff check src/
mypy src/

# Tests
pytest tests/ -v --cov

# Docs
python .moai/tools/validate-docs.py
```

### Release (Local Only)
```bash
/moai:99-release  # Local release command
```

---

## Directory Structure

```
MoAI-ADK/
├── src/moai_adk/          # Package source (work here)
│   ├── .claude/           # Templates
│   ├── .moai/             # Templates
│   └── templates/         # User templates
│
├── .claude/               # Synced from src
├── .moai/                 # Synced from src
├── CLAUDE.md              # Synced from templates
├── CLAUDE.local.md        # This file (local only)
└── tests/                 # Test suite
```

---

## Important Notes

- `/Users/goos/MoAI/MoAI-ADK/.claude/settings.json` uses substituted variables
- Never commit time estimates ("4-6 hours") - avoid unverified timeframes
- Template changes trigger auto-sync via hooks
- Local config.json is never synced to package

---

## Path Variable Strategy

### Template vs Local Settings

MoAI-ADK uses different path variable strategies for template and local environments:

**Template settings.json** (`src/moai_adk/templates/.claude/settings.json`):
- Uses: `{{PROJECT_DIR}}` placeholder
- Purpose: Package distribution (replaced during project initialization)
- Cross-platform: Works on Windows, macOS, Linux after substitution
- Example:
  ```json
  {
    "command": "uv run {{PROJECT_DIR}}/.claude/hooks/moai/session_start__show_project_info.py"
  }
  ```

**Local settings.json** (`.claude/settings.json`):
- Uses: `"$CLAUDE_PROJECT_DIR"` environment variable
- Purpose: Runtime path resolution by Claude Code
- Cross-platform: Automatically resolved by Claude Code on any OS
- Example:
  ```json
  {
    "command": "uv run \"$CLAUDE_PROJECT_DIR\"/.claude/hooks/moai/session_start__show_project_info.py"
  }
  ```

### Why Two Different Variables?

1. **Template (`{{PROJECT_DIR}}`)**:
   - Static placeholder replaced during `moai-adk init`
   - Ensures new projects get correct absolute paths
   - Part of the package distribution system

2. **Local (`"$CLAUDE_PROJECT_DIR}")**:
   - Dynamic runtime variable resolved by Claude Code
   - No hardcoded paths in version control
   - Works across different developer environments
   - Claude Code automatically expands to actual project directory

### Critical Rules

✅ **DO**:
- Keep `{{PROJECT_DIR}}` in template files (src/moai_adk/templates/)
- Keep `"$CLAUDE_PROJECT_DIR"` in local files (.claude/)
- Quote the variable: `"$CLAUDE_PROJECT_DIR"` (prevents shell expansion issues)

❌ **DON'T**:
- Never use absolute paths in templates (breaks cross-platform compatibility)
- Never commit `{{PROJECT_DIR}}` in local files (breaks runtime resolution)
- Never use `$CLAUDE_PROJECT_DIR` without quotes (causes parsing errors)

### Verification

Check your settings.json path variables:

```bash
# Template should use {{PROJECT_DIR}}
grep "PROJECT_DIR" src/moai_adk/templates/.claude/settings.json

# Local should use "$CLAUDE_PROJECT_DIR"
grep "CLAUDE_PROJECT_DIR" .claude/settings.json
```

Expected output:
```
# Template:
{{PROJECT_DIR}}/.claude/hooks/moai/session_start__show_project_info.py

# Local:
"$CLAUDE_PROJECT_DIR"/.claude/hooks/moai/session_start__show_project_info.py
```

---

## Reference

- CLAUDE.md: Alfred execution directives
- README.md: Project overview
- Skills: `Skill("moai-foundation-core")` for execution rules

---

**Status**: ✅ Active (Local Development)
**Version**: 2.1.0 (Path Variable Strategy Added)
**Last Updated**: 2025-11-26

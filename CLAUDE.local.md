# MoAI-ADK Local Development Guide

> **Purpose**: Essential guide for local MoAI-ADK development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2025-12-04

---

## Quick Start

### Work Location
```bash
# Primary work location (template development)
/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/

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
src/moai_adk/templates/.claude/    → .claude/
src/moai_adk/templates/.moai/      → .moai/
src/moai_adk/templates/CLAUDE.md   → ./CLAUDE.md
```

### Local-Only Files (Never Sync)
```
.claude/commands/moai/99-release.md  # Local release command
.claude/settings.local.json          # Personal settings
CLAUDE.local.md                      # This file
.moai/cache/                         # Cache
.moai/logs/                          # Logs
.moai/rollbacks/                     # Rollback data
```

### Template-Only Files (Distribution)
```
src/moai_adk/templates/.moai/config/config.yaml     # Default config template
src/moai_adk/templates/.moai/config/presets/        # Configuration presets
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
rsync -avz src/moai_adk/templates/.claude/ .claude/
rsync -avz src/moai_adk/templates/.moai/ .moai/
cp src/moai_adk/templates/CLAUDE.md ./CLAUDE.md
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
├── src/moai_adk/              # Package source
│   ├── cli/                   # CLI commands
│   ├── core/                  # Core modules
│   ├── foundation/            # Foundation components
│   ├── project/               # Project management
│   ├── statusline/            # Statusline features
│   ├── templates/             # Distribution templates (work here)
│   │   ├── .claude/           # Claude Code config templates
│   │   │   ├── agents/        # Agent definitions
│   │   │   ├── commands/      # Slash commands
│   │   │   ├── hooks/         # Hook scripts
│   │   │   ├── output-styles/ # Output style definitions
│   │   │   └── skills/        # Skill definitions
│   │   ├── .moai/             # MoAI config templates
│   │   │   └── config/        # config.yaml template
│   │   └── CLAUDE.md          # Alfred execution directives
│   └── utils/                 # Utility modules
│
├── .claude/                   # Synced from templates
├── .moai/                     # Synced from templates
├── CLAUDE.md                  # Synced from templates
├── CLAUDE.local.md            # This file (local only)
└── tests/                     # Test suite
```

---

## Important Notes

- `/Users/goos/MoAI/MoAI-ADK/.claude/settings.json` uses substituted variables
- Template changes trigger auto-sync via hooks
- Local config is never synced to package (user-specific)
- Output styles allow visual emphasis emoji (🤖 R2-D2 ★) per CLAUDE.md Documentation Standards

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

2. **Local (`"$CLAUDE_PROJECT_DIR"`)**:
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

## Output Styles

### Visual Emphasis Emoji Policy

Per CLAUDE.md Documentation Standards, output styles may use visual emphasis emoji:

**Allowed in output styles:**
- Header decorations: `🤖 R2-D2 ★ Code Insight`, `🧙 Yoda ★ Deep Understanding`
- Section markers: `💡`, `📊`, `⚡`, `✅`, `❓`, `🔍`
- Brand identity: `🗿 MoAI-ADK`
- Numbered items: `1️⃣`, `2️⃣`, `3️⃣`, `4️⃣`

**NOT allowed in AskUserQuestion:**
- No emoji in question text, headers, or option labels

### Output Style Locations

```
src/moai_adk/templates/.claude/output-styles/moai/
├── r2d2.md    # Pair programming partner (v2.0.0)
└── yoda.md    # Technical wisdom master (v2.0.0)
```

---

## Configuration System

### Config File Format

MoAI-ADK uses YAML for configuration:

**Template config** (`src/moai_adk/templates/.moai/config/config.yaml`):
- Default configuration template
- Distributed to new projects via `moai-adk init`
- Contains presets for different languages/regions

**User config** (created by users, not synced):
- Personal configuration overrides
- Language preferences
- User identification

### Configuration Priority

1. Environment Variables (highest priority): `MOAI_USER_NAME`, `MOAI_CONVERSATION_LANG`
2. User Configuration File: `.moai/config/config.yaml` (user-created)
3. Template Defaults: From package distribution

---

## Reference

- CLAUDE.md: Alfred execution directives (v8.1.0)
- README.md: Project overview
- Skills: `Skill("moai-foundation-core")` for execution rules
- Output Styles: r2d2.md, yoda.md (v2.0.0)

---

**Status**: ✅ Active (Local Development)
**Version**: 2.2.0 (Config YAML, Output Styles, Directory Structure Update)
**Last Updated**: 2025-12-04

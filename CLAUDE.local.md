# MoAI-ADK Local Development Guide

> **Purpose**: Essential guide for local MoAI-ADK development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-01-13

---

## 1. Quick Start

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

## 2. File Synchronization

### Auto-Sync Directories
```bash
src/moai_adk/templates/.claude/    → .claude/
src/moai_adk/templates/.moai/      → .moai/
src/moai_adk/templates/CLAUDE.md   → ./CLAUDE.md
```

### Protected Directories (Never Delete During Sync)
```bash
# CRITICAL: These directories contain user data and must NEVER be deleted
.moai/project/    # Project documentation (product.md, structure.md, tech.md)
.moai/specs/      # SPEC documents (active development files)
```

### Local-Only Files (Never Sync)
```
.claude/commands/moai/99-release.md  # Local release command
.claude/settings.local.json          # Personal settings
CLAUDE.local.md                      # This file
.moai/cache/                         # Cache
.moai/logs/                          # Logs
.moai/rollbacks/                     # Rollback data
.moai/project/                       # Project docs (protected from deletion)
.moai/specs/                         # SPEC documents (protected from deletion)
```

### Template-Only Files (Distribution)
```
src/moai_adk/templates/.moai/config/config.yaml     # Default config template
src/moai_adk/templates/.moai/config/presets/        # Configuration presets
```

---

## 3. Code Standards

### Language: English Only

**Source Code:**
- All code, comments, docstrings in English
- Variable names: camelCase or snake_case
- Class names: PascalCase
- Constants: UPPER_SNAKE_CASE
- Commit messages: English

**Configuration Files (English ONLY):**
- Command files (.claude/commands/**/*.md): English only
- Agent definitions (.claude/agents/**/*.md): English only
- Skill definitions (.claude/skills/**/*.md): English only
- Hook scripts (.claude/hooks/**/*.py): English only
- CLAUDE.md: English only

**Why**: Command/agent/skill files are code, not user-facing content. They are read by Claude Code (English-based) and must be in English for consistent behavior.

**User-facing vs Internal:**
- User-facing: README, CHANGELOG, documentation (can be localized)
- Internal: Commands, agents, skills, hooks (MUST be English)

### Forbidden
```python
# WRONG - Korean comments
def calculate():  # 계산
    pass

# CORRECT - English comments
def calculate():  # Calculate score
    pass
```

---

## 4. Git Workflow

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

## 5. Version Management

### Single Source of Truth

- [HARD] pyproject.toml is the ONLY authoritative source for MoAI-ADK version
- WHY: Prevents version inconsistencies across multiple files

Version Reference:

- Authoritative Source: pyproject.toml (version = "X.Y.Z")
- Runtime Access: src/moai_adk/version.py reads from pyproject.toml
- Config Display: .moai/config/sections/system.yaml (updated by release process)

### Files Requiring Version Sync

When releasing new version, these files MUST be updated:

Documentation Files:

- README.md (Version line)
- README.ko.md (Version line)
- README.ja.md (Version line)
- README.zh.md (Version line)
- CHANGELOG.md (New version entry)

Configuration Files:

- pyproject.toml (Single Source - update FIRST)
- src/moai_adk/version.py (_FALLBACK_VERSION)
- .moai/config/sections/system.yaml (moai.version)
- src/moai_adk/templates/.moai/config/config.yaml (moai.version)

### Version Sync Process

- [HARD] Before any release:

Step 1: Update pyproject.toml

- Change version = "X.Y.Z" to new version

Step 2: Run Version Sync Script

- Execute: .github/scripts/sync-versions.sh X.Y.Z
- Or manually update all files listed above

Step 3: Verify Consistency

- Run: grep -r "X.Y.Z" to confirm all files updated
- Check: No old version numbers remain in critical files

### Prohibited Practices

- [HARD] Never hardcode version in multiple places without sync mechanism
- [HARD] Never update README version without updating pyproject.toml
- [HARD] Never release with mismatched versions across files

WHY: Version inconsistency causes confusion and breaks tooling expectations.

---

## 6. Plugin Development

### What are Plugins

Plugins are reusable extensions that bundle Claude Code configurations for distribution across projects. Unlike standalone configurations in .claude/ directories, plugins can be installed via marketplaces and version-controlled independently.

### Plugin vs Standalone Configuration

Standalone Configuration:

- Scope: Single project only
- Sharing: Manual copy or git submodules
- Best for: Project-specific customizations

Plugin Configuration:

- Scope: Reusable across multiple projects
- Sharing: Installable via marketplaces or git URLs
- Best for: Team standards, reusable workflows, community tools

### Plugin Structure

Create a plugin directory with the following structure:

- .claude-plugin/plugin.json - Plugin manifest with name, description, version, author
- commands/ - Slash commands
- agents/ - Sub-agent definitions
- skills/ - Skill definitions
- hooks/hooks.json - Hook configurations
- .mcp.json - MCP server configurations

### Plugin Management Commands

Installation:

- /plugin install plugin-name - Install from marketplace
- /plugin install owner/repo - Install from GitHub
- /plugin install plugin-name --scope project - Install with scope

Other Commands:

- /plugin uninstall - Remove a plugin
- /plugin enable - Enable a disabled plugin
- /plugin disable - Disable a plugin temporarily
- /plugin update - Update to latest version
- /plugin list - List installed plugins
- /plugin validate - Validate plugin structure

For detailed plugin development, refer to Skill("moai-foundation-claude") reference documentation.

---

## 7. Sandboxing

### OS-Level Security Isolation

Claude Code provides OS-level sandboxing to restrict file system and network access during code execution.

Platform-Specific Implementation:

- Linux: Uses bubblewrap (bwrap) for namespace-based isolation
- macOS: Uses Seatbelt (sandbox-exec) for profile-based restrictions

### Default Sandbox Behavior

When sandboxing is enabled:

- File writes are restricted to the current working directory
- Network access is limited to allowed domains
- System resources are protected from modification

### Auto-Allow Mode

If a command only reads from allowed paths, writes to allowed paths, and accesses allowed network domains, it executes automatically without user confirmation.

### Security Best Practices

Start Restrictive:

- Begin with minimal permissions
- Monitor for violations
- Add specific allowances as needed

Combine with IAM:

- Sandbox provides OS-level isolation
- IAM provides Claude-level permissions
- Together they create defense-in-depth

For detailed configuration, refer to Skill("moai-foundation-claude") reference documentation.

---

## 8. Headless Mode and CI/CD

### Basic Usage

Simple Prompt:

- claude -p "Your prompt here" - Runs Claude with the given prompt and exits after completion

Continue Previous Conversation:

- claude -c "Follow-up question" - Continues the most recent conversation

Resume Specific Session:

- claude -r session_id "Continue this task" - Resumes a specific session by ID

### Output Formats

Available formats:

- text - Default plain text output
- json - Structured JSON output
- stream-json - Streaming JSON for real-time processing

### Tool Management

Allow Specific Tools:

- claude -p "Build the project" --allowedTools "Bash,Read,Write" - Auto-approves specified tools

Tool Pattern Matching:

- claude -p "Check git status" --allowedTools "Bash(git:*)" - Allow only specific patterns

### Structured Output with JSON Schema

Validate output against provided JSON schema for reliable data extraction in automated pipelines.

Use --json-schema flag with a schema file to enforce output structure.

### Best Practices for CI/CD

- Use --append-system-prompt to retain Claude Code capabilities
- [HARD] Always specify --allowedTools in CI/CD to prevent unintended actions
- Use --output-format json for reliable parsing
- Handle errors with exit code checks
- Use --agents for multi-agent orchestration in pipelines

For complete CLI reference, refer to Skill("moai-foundation-claude") reference documentation.

---

## 9. Documentation Standards

### Required Practices

All instruction documents must follow these standards:

Formatting Requirements:

- Use detailed markdown formatting for explanations
- Document step-by-step procedures in text form
- Describe concepts and logic in narrative style
- Present workflows with clear textual descriptions
- Organize information using list format

### Content Restrictions

Restricted Content:

- [HARD] Conceptual explanations expressed as code examples
- [HARD] Flow control logic expressed as code syntax
- [HARD] Decision trees shown as code structures
- [HARD] Table format in instructions
- [HARD] Emoji characters in instructions
- [HARD] Time estimates or duration predictions

WHY: Code examples can be misinterpreted as executable commands. Flow control must use narrative text format.

### Scope of Application

These standards apply to:

- CLAUDE.md
- Agent definitions
- Slash commands
- Skill definitions
- Hook definitions
- Configuration files

Note: These restrictions do NOT apply to:

- Output styles (may use visual emphasis emoji)
- User-facing documentation
- README files
- Code files themselves

---

## 10. Path Variable Strategy

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

DO:
- Keep `{{PROJECT_DIR}}` in template files (src/moai_adk/templates/)
- Keep `"$CLAUDE_PROJECT_DIR"` in local files (.claude/)
- Quote the variable: `"$CLAUDE_PROJECT_DIR"` (prevents shell expansion issues)

DO NOT:
- [HARD] Never use absolute paths in templates (breaks cross-platform compatibility)
- [HARD] Never commit `{{PROJECT_DIR}}` in local files (breaks runtime resolution)
- [HARD] Never use `$CLAUDE_PROJECT_DIR` without quotes (causes parsing errors)

### Extension to Local Agent/Skill Files

**Local agents/skills** (`.claude/agents/**/*.md`, `.claude/skills/**/*.md`):
- Uses: `$CLAUDE_PROJECT_DIR` environment variable
- Purpose: Runtime path resolution for hook commands
- Why: These files are executed directly by Claude Code in the local environment

**Template agents/skills** (`src/moai_adk/templates/.claude/agents/**/*.md`):
- Uses: `{{PROJECT_DIR}}` placeholder
- Purpose: Replaced during package distribution
- Why: Ensures new projects get correct paths after initialization

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

## 11. Configuration System

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

## 12. Output Styles

### Visual Emphasis Emoji Policy

Per CLAUDE.md Documentation Standards, output styles may use visual emphasis emoji:

**Allowed in output styles:**
- Header decorations: `R2-D2 Code Insight`, `Yoda Deep Understanding`
- Section markers for visual separation
- Brand identity markers
- Numbered items for lists

**NOT allowed in AskUserQuestion:**
- No emoji in question text, headers, or option labels

### Output Style Locations

```
src/moai_adk/templates/.claude/output-styles/moai/
├── r2d2.md    # Pair programming partner (v2.0.0)
└── yoda.md    # Technical wisdom master (v2.0.0)
```

---

## 13. Directory Structure

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

## 14. Frequently Used Commands

### Sync Commands
```bash
# Sync from template to local
# IMPORTANT: --exclude prevents deletion of protected directories
rsync -avz --delete src/moai_adk/templates/.claude/ .claude/
rsync -avz --delete --exclude='project/' --exclude='specs/' src/moai_adk/templates/.moai/ .moai/
cp src/moai_adk/templates/CLAUDE.md ./CLAUDE.md

# Alternative: One-liner for all sync (with protection)
rsync -avz --delete src/moai_adk/templates/.claude/ .claude/ && \
rsync -avz --delete --exclude='project/' --exclude='specs/' src/moai_adk/templates/.moai/ .moai/ && \
cp src/moai_adk/templates/CLAUDE.md ./CLAUDE.md
```

### Validation Commands
```bash
# Code quality
ruff check src/
mypy src/

# Tests
pytest tests/ -v --cov

# Docs
python .moai/tools/validate-docs.py
```

### Release Commands (Local Only)
```bash
# Use the local release command
/moai:99-release

# Manual version sync
.github/scripts/sync-versions.sh X.Y.Z

# Verify version consistency
grep -r "X.Y.Z" pyproject.toml README.md CHANGELOG.md
```

---

## 15. Important Notes

- `/Users/goos/MoAI/MoAI-ADK/.claude/settings.json` uses substituted variables
- Template changes trigger auto-sync via hooks
- Local config is never synced to package (user-specific)
- Output styles allow visual emphasis emoji per CLAUDE.md Documentation Standards
- **CRITICAL**: `.moai/project/` and `.moai/specs/` are protected from deletion during sync
  - These directories contain user-generated project documentation and active SPEC files
  - Always use `--exclude='project/' --exclude='specs/'` when syncing `.moai/`
  - If accidentally deleted, restore with: `git checkout <commit-hash> -- .moai/project/ .moai/specs/`

---

## 16. Reference

- CLAUDE.md: Alfred execution directives (v9.3.0)
- README.md: Project overview
- Skills: `Skill("moai-foundation-core")` for execution rules
- Skills: `Skill("moai-foundation-claude")` for plugin development, sandboxing, headless mode
- Output Styles: r2d2.md, yoda.md (v2.0.0)

---

## 17. User Communication Guidelines

### Standard Update Instructions

When responding to issues or comments that require users to update MoAI-ADK, ALWAYS use the following format:

[HARD] Use `uv tool update moai-adk` as the primary update method.

```bash
# Update to the latest version
uv tool update moai-adk
```

Alternative methods (only if user requests):

```bash
# Using pipx (alternative)
pipx upgrade moai-adk

# Using pip (not recommended for tool installation)
pip install --upgrade moai-adk
```

### Prohibited Practices

- [HARD] NEVER recommend `pip install --upgrade moai-adk` as primary method
- [HARD] NEVER recommend `uv pip install --upgrade moai-adk` (wrong usage pattern)
- [HARD] ALWAYS use `uv tool update moai-adk` for uv-based installations

WHY: `uv tool` is the correct command for updating uv-installed tools. `uv pip` is for package management, not tool management.

### Issue Response Template

When resolving issues, include this standard update instruction:

```
### How to Apply the Fix

Update to the latest version:

```bash
uv tool update moai-adk
```

After updating, the issue will be resolved.
```

### Communication Standards

All user-facing communication should follow these standards:

- Language: English for GitHub issues and pull requests
- Tone: Professional, helpful, and concise
- Code blocks: Always use proper markdown syntax
- Links: Verify all URLs before including

---

## 17. Testing Guidelines

### ⚠️ IMPORTANT: Prevent Accidental File Modifications

When running tests, **always execute from an isolated directory** to prevent tests from modifying project files like `.claude/settings.json`.

### Recommended Test Execution

```bash
# ✅ CORRECT: Run from isolated directory
cd /tmp/moai-test && pytest /Users/goos/MoAI/MoAI-ADK

# ❌ WRONG: Run from project root (may modify settings.json)
pytest
```

### Why This Matters

Some tests use `Path.cwd()` to access the current working directory. When run from the project root, these tests can:
- Modify `.claude/settings.json` with test data
- Overwrite user configurations
- Cause git diff noise

### Verification

After running tests, check if project files were modified:

```bash
git status
```

If `.claude/settings.json` appears as modified, restore it from git:

```bash
git checkout .claude/settings.json
```

### Continuous Integration

CI/CD pipelines should always run tests from isolated directories:

```yaml
# Example GitHub Actions
- name: Run tests
  run: |
    cd /tmp/moai-test
    pytest $GITHUB_WORKSPACE
```

### pytest.ini Configuration

The project includes `pytest.ini` with test isolation settings:

```ini
[pytest]
testpaths = tests
addopts =
    -v
    --cov=src/moai_adk
    --cov-report=html
    --cov-report=term-missing
    --strict-markers
```

This configuration helps prevent accidental file modifications, but **always run tests from an isolated directory** to be safe.

### Parallel Test Execution with pytest-xdist

[HARD] Always use parallel execution for faster test runs:

```bash
# ✅ CORRECT: Parallel execution (auto-detects CPU cores)
pytest -n auto

# ✅ CORRECT: Parallel execution with coverage
pytest -n auto --cov=src/moai_adk --cov-report=term-missing

# ✅ CORRECT: Specify number of workers explicitly
pytest -n 10

# ❌ AVOID: Sequential execution (slow)
pytest
```

WHY:
- Speed: Tests run ~N times faster (where N = number of CPU cores)
- Efficiency: Uses all available CPU resources
- Standard practice: Modern CI/CD pipelines expect parallel execution

NOTE: When measuring coverage, pytest-cov works with pytest-xdist using shared data directory. Coverage results are automatically aggregated from all workers.

### Root Cause of settings.json Modifications

**Historical Issue**: Commit `42db79e4` (`test(coverage): achieve 88.12% coverage...`) accidentally modified `.claude/settings.json` with test data because tests were run from the project root.

**Prevention**: The `pytest.ini` file and this guideline are added to prevent future occurrences.

---

**Status**: Active (Local Development)
**Version**: 3.2.0 (Added Testing Guidelines)
**Last Updated**: 2026-01-13

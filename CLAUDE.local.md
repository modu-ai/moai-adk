/out# MoAI-ADK Local Development Guide

> **Purpose**: Essential guide for local MoAI-ADK development
> **Audience**: GOOS (local developer only)
> **Last Updated**: 2026-01-28

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
.claude/commands/moai/99-release.md  # Local release command (NOT in template, deprecated folder)
.claude/settings.local.json          # Personal settings
CLAUDE.local.md                      # This file
.moai/cache/                         # Cache
.moai/logs/                          # Logs
.moai/rollbacks/                     # Rollback data
.moai/project/                       # Project docs (protected from deletion)
.moai/specs/                         # SPEC documents (protected from deletion)
```

**Note on 99-release.md**: This file is intentionally kept local-only and is NOT distributed with the template. It provides developer-specific release workflow automation that should not be part of the public distribution.

**Note on commands/moai/ deprecation (v1.10.0+)**: The `.claude/commands/moai/` folder is deprecated and automatically deleted during `moai-adk update` on target projects. Commands have been migrated to the skill system (`/moai` skill). The folder is backed up before deletion via `TemplateBackup.create_backup()`. On this local development project, update is not run, so `99-release.md` is unaffected.

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

### Migration Notes (v1.8.0)

**Historical Context:**

Prior to v1.8.0, MoAI-ADK used platform-specific path variables:
- `{{PROJECT_DIR_UNIX}}`: Forward slash path (worked on all platforms)
- `{{PROJECT_DIR_WIN}}`: Backslash path (non-functional due to Claude Code bug #6023)
- `{{PROJECT_DIR}}`: Legacy variable without trailing separator

**v1.8.0 Consolidation:**

All template variables were consolidated to a single `{{PROJECT_DIR}}` that:
- Uses forward slash separators (works on Windows, macOS, Linux)
- Includes trailing separator for consistency
- Eliminates platform-specific variable confusion

**Migration for Existing Projects:**

If you have an existing project with deprecated variables:
1. Search for `{{PROJECT_DIR_UNIX}}` and `{{PROJECT_DIR_WIN}}` in your `.claude/` directory
2. Replace all occurrences with `{{PROJECT_DIR}}`
3. Verify hook scripts execute correctly
4. No changes needed to runtime `$CLAUDE_PROJECT_DIR` variable

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
# IMPORTANT: --exclude prevents deletion of local-only files and protected directories
rsync -avz --delete \
  --exclude='commands/moai/99-release.md' \
  --exclude='settings.json' \
  --exclude='settings.json.unix' \
  --exclude='settings.json.windows' \
  --exclude='plans/' \
  src/moai_adk/templates/.claude/ .claude/

rsync -avz --delete \
  --exclude='project/' \
  --exclude='specs/' \
  --exclude='cache/' \
  --exclude='logs/' \
  --exclude='memory/' \
  src/moai_adk/templates/.moai/ .moai/

cp src/moai_adk/templates/CLAUDE.md ./CLAUDE.md

# Post-sync: Replace template variables with local development values
# {{PROJECT_DIR}} -> $CLAUDE_PROJECT_DIR (runtime variable for local)
# {{MOAI_VERSION}} -> actual version from pyproject.toml
# {{CONVERSATION_LANGUAGE}} -> ko (Korean for local development)
# {{CONVERSATION_LANGUAGE_NAME}} -> Korean (한국어)

# Replace PROJECT_DIR in agents, skills, commands
find .claude/agents -name "*.md" -exec sed -i '' 's|{{PROJECT_DIR}}|$CLAUDE_PROJECT_DIR|g' {} \;
find .claude/skills -name "*.md" -exec sed -i '' 's|{{PROJECT_DIR}}|$CLAUDE_PROJECT_DIR|g' {} \;
find .claude/commands -name "*.md" -exec sed -i '' 's|{{PROJECT_DIR}}|$CLAUDE_PROJECT_DIR|g' {} \;

# Replace version and language settings
VERSION=$(grep -m1 'version = ' pyproject.toml | cut -d'"' -f2)
sed -i '' "s|{{MOAI_VERSION}}|$VERSION|g" .moai/config/sections/system.yaml
sed -i '' 's|{{CONVERSATION_LANGUAGE}}|ko|g' .moai/config/sections/language.yaml
sed -i '' 's|{{CONVERSATION_LANGUAGE_NAME}}|Korean (한국어)|g' .moai/config/sections/language.yaml
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
/moai release

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

[HARD] Use `moai update` as the primary update method for existing projects.
[HARD] Use `uv tool install moai-adk` as the primary installation method for new users.

```bash
# Update existing project (auto-upgrades package + syncs templates)
moai update

# New installation
uv tool install moai-adk
```

### Prohibited Practices

- [HARD] NEVER recommend `claude install moai-adk` (this command does not exist)
- [HARD] NEVER recommend `pip install --upgrade moai-adk` as primary method
- [HARD] NEVER recommend `uv pip install --upgrade moai-adk` (wrong usage pattern)
- [HARD] ALWAYS use `moai update` for existing project updates
- [HARD] ALWAYS use `uv tool install moai-adk` for new installations

WHY: `moai update` handles both package upgrade and template synchronization. `uv tool install` is the official installation method per README.md.

### Issue Response Template

When resolving issues, include this standard update instruction:

```
### How to Apply the Fix

Update your project to the latest version:

```bash
moai update
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

---

## 18. Memory Management Guidelines

### Node.js V8 Heap Memory Limits

Claude Code runs on Node.js, which has default V8 heap memory limits that can cause crashes during long-running agent sessions.

**Default Heap Limits**:
- 32-bit systems: ~512 MB
- 64-bit systems: ~2 GB (typical crash point ~4 GB)

**Symptoms of Memory Issues**:
- Process crash with: `FATAL ERROR: Ineffective mark-compacts near heap limit`
- Agent stops responding after extended execution
- Session terminates unexpectedly after 20+ minutes

### Workaround: Increase Node.js Heap Size

For long-running workflows, you can increase the heap size:

```bash
# Set environment variable before running Claude Code
export NODE_OPTIONS="--max-old-space-size=8192"  # 8 GB
# Or
export NODE_OPTIONS="--max-old-space-size=16384"  # 16 GB
```

**Note**: This is a temporary workaround. The root cause (context accumulation) is addressed by agent checkpoint/resume functionality.

### Long-Running Agent Considerations

The following agents are prone to memory issues during extended sessions:

**manager-ddd**:
- Token budget: high
- Context retention: high
- Typical use case: Large refactoring operations (30+ minutes)
- Risk: High

**manager-docs**:
- Token budget: medium
- Context retention: low
- Typical use case: Large documentation generation
- Risk: Medium

### Best Practices for Long Workflows

1. **Break Down Large Tasks**:
   - Divide work into smaller SPEC files
   - Run `/moai run` separately for each SPEC
   - Use `/moai sync` after each implementation completes

2. **Use Loop Mode Wisely**:
   - `/moai loop` maintains state across iterations
   - Monitor memory usage for loops >100 iterations
   - Consider using `--max` to limit iterations

3. **Enable Resume Capability** (Coming Soon):
   - Agents will support checkpoint-based recovery
   - Work can resume from saved state after crash
   - No need to restart from beginning

### Current Limitations

**Agents without Resume Support**:
- manager-ddd: `can_resume: false`
- manager-docs: `can_resume: false`

**Impact**:
- Cannot recover from memory crashes
- Must restart work from beginning
- Lost progress after crash

**Future Improvements**:
P1 priority tasks to add checkpoint/resume capability to these agents.

---

**Status**: Active (Local Development)
**Version**: 3.3.0 (Added Memory Management Guidelines)
**Last Updated**: 2026-01-22

---

## 19. Hook Development Guidelines

### [HARD] Always Use Login Shell for Hooks

**Rule**: ALL hook commands MUST use `bash -l -c` to ensure PATH is loaded correctly.

**WHY**: 
- Claude Code executes hooks in non-interactive shells
- Non-interactive shells don't load `.bashrc` / `.zshrc`
- Commands like `uv`, `pytest`, `ruff` in `~/.local/bin` won't be found
- Issue #296: bash users experienced hook failures due to missing PATH

**Correct Pattern**:
```json
// ✅ CORRECT
{
  "type": "command",
  "command": "bash -l -c 'uv run \"{{PROJECT_DIR}}/.claude/hooks/moai/session_start__show_project_info.py'"
}
```

**Wrong Pattern**:
```json
// ❌ WRONG - PATH not loaded
{
  "type": "command",
  "command": "uv run \"{{PROJECT_DIR}}/.claude/hooks/moai/session_start__show_project_info.py\""
}
```

### File Locations

**Template files** (use `{{PROJECT_DIR}}`):
- `src/moai_adk/templates/.claude/settings.json`
- `src/moai_adk/templates/.claude/agents/moai/*.md`

**Local files** (use `$CLAUDE_PROJECT_DIR`):
- `.claude/settings.json`
- `.claude/agents/moai/*.md`

### Quote Rules

**Double-quoted command with single-quoted bash command**:
```json
{
  "command": "bash -l -c 'uv run \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.py\"'"
}
```

**Single-quoted command with double-quoted bash command**:
```json
{
  'command': 'bash -l -c "uv run \"$CLAUDE_PROJECT_DIR/.claude/hooks/moai/hook.py\"\"'
}
```

### Cross-Platform Compatibility

**Supported Platforms**:
- ✅ macOS (bash/zsh)
- ✅ Linux (bash/zsh)
- ✅ WSL (bash/zsh)
- ✅ Windows Git Bash

**Login Shell Behavior**:
| Platform | Shell | Login Shell Loads |
|----------|-------|-------------------|
| macOS    | bash  | `.bash_profile`, `.bashrc` |
| macOS    | zsh   | `.zprofile`, `.zshenv` |
| Linux    | bash  | `.bash_profile`, `.bashrc` |
| Linux    | zsh   | `.zprofile`, `.zshenv` |
| WSL      | bash  | `.bash_profile`, `.bashrc` |
| WSL      | zsh   | `.zprofile`, `.zshenv` |

### Testing Hook Commands

```bash
# Test if uv is found via login shell
bash -l -c 'which uv'
# Expected: /Users/goos/.local/bin/uv

# Test hook execution
bash -l -c 'CLAUDE_PROJECT_DIR=/Users/goos/MoAI/MoAI-ADK uv run "/Users/goos/MoAI/MoAI-ADK/.claude/hooks/moai/session_start__show_project_info.py"'
# Expected: JSON output with session info
```

### Common Mistakes

**❌ Forgetting Login Shell Flag**:
```json
// WRONG - non-login shell, PATH not loaded
"command": "bash -c 'uv run ...'"
```

**❌ Missing Quote Escaping**:
```json
// WRONG - quote mismatch
"command": "bash -l -c "uv run \"$CLAUDE_PROJECT_DIR/hook.py\"""
```

**❌ Using Absolute Path Instead of Login Shell**:
```json
// WRONG - brittle, assumes uv location
"command": "$HOME/.local/bin/uv run ..."
```

### Related Commits

- e1777b94: Initial fix for settings.json hooks
- 80602d5d: Applied to all agent definitions
- c3020305: Consolidated platform-specific settings

### References

- Issue #296: PATH loading problem with moai update
- CLAUDE.md Section 13: Parallel Execution Safeguards

---

**Status**: Active (Local Development)
**Version**: 3.4.0 (Added Hook Development Guidelines)
**Last Updated**: 2026-01-25

---

## 20. Cross-Platform Development Guidelines

### Overview

MoAI-ADK must work seamlessly across Windows, macOS, and Linux. This section provides guidelines for developing features that work on all platforms without modification.

### Critical Rules

[HARD] All hook commands MUST use cross-platform template variables:
- Use `{{HOOK_SHELL_PREFIX}}` and `{{HOOK_SHELL_SUFFIX}}` for shell wrappers
- Use `{{PROJECT_DIR}}` for paths (forward slash, works on all platforms)
- NEVER hardcode `bash -l -c` or platform-specific shell commands

[HARD] Test on all three platforms before releasing:
- Windows: Test with Git Bash and PowerShell
- macOS: Test with both bash and zsh
- Linux: Test with bash (primary) and zsh (secondary)

### Platform Differences

**Shell Configuration Files**:
- macOS (zsh default): `~/.zshenv`, `~/.zprofile`, `~/.zshrc`
- macOS (bash): `~/.bash_profile`, `~/.bashrc`
- Linux (bash): `~/.bash_profile`, `~/.bashrc`, `~/.profile`
- Windows: System PATH (no shell config files)

**PATH Loading Behavior**:
- Windows: PATH loaded from system environment automatically
- macOS/Linux: Requires login shell (`-l` flag) to load user PATH

**Shell Wrapper Strategy** (implemented in v1.8.6):
```python
# update.py implementation
if is_windows:
    hook_shell_prefix = ""  # Direct execution
    hook_shell_suffix = ""
else:
    # Use user's default shell, fallback to bash
    hook_shell_prefix = '${SHELL:-/bin/bash} -l -c \''
    hook_shell_suffix = '\''
```

### Template Variable Usage

**Correct Usage** ✅:
```json
{
  "command": "{{HOOK_SHELL_PREFIX}}uv run \"{{PROJECT_DIR}}.claude/hooks/moai/session_start.py\"{{HOOK_SHELL_SUFFIX}}"
}
```

**Result After Substitution**:
```bash
# Windows
uv run "%CLAUDE_PROJECT_DIR%/.claude/hooks/moai/session_start.py"

# macOS/Linux (zsh user)
${SHELL:-/bin/bash} -l -c 'uv run "$CLAUDE_PROJECT_DIR/.claude/hooks/moai/session_start.py"'
```

**Incorrect Usage** ❌:
```json
{
  "command": "bash -l -c 'uv run \"{{PROJECT_DIR}}.claude/hooks/...\"'"
}
```

WHY: Hardcoding `bash` breaks on Windows and forces bash on zsh users.

### Testing Methodology

**Local Testing** (macOS/Linux):
```bash
# Test with user's default shell
$SHELL -l -c 'echo $PATH | grep -o "[^:]*local/bin[^:]*"'

# Test with bash explicitly
bash -l -c 'echo $PATH | grep -o "[^:]*local/bin[^:]*"'

# Test with zsh explicitly
zsh -l -c 'echo $PATH | grep -o "[^:]*local/bin[^:]*"'

# Verify PATH priority (should be first)
$SHELL -l -c 'echo $PATH' | tr ':' '\n' | head -3
```

**Expected Results**:
- `~/.local/bin` should appear in PATH
- For zsh: `~/.local/bin` should be FIRST (highest priority)
- For bash: `~/.local/bin` may be at the end but still present

**Windows Testing**:
```powershell
# Test PATH availability
$env:PATH -split ';' | Select-String -Pattern 'local\\bin'

# Test direct execution
uv --version
```

### Common Pitfalls

**Pitfall 1: Hardcoded Shell**
```json
// WRONG
"command": "bash -l -c 'uv run ...'"

// CORRECT
"command": "{{HOOK_SHELL_PREFIX}}uv run ...{{HOOK_SHELL_SUFFIX}}"
```

**Pitfall 2: Path Separator Assumptions**
```python
# WRONG - assumes Unix separator
path = f"{project_dir}/.claude/hooks"

# CORRECT - use forward slash (works on all platforms since Windows 10)
path = f"{project_dir}.claude/hooks"  # PROJECT_DIR includes trailing /
```

**Pitfall 3: Assuming PATH is Loaded**
```json
// WRONG - no shell wrapper
"command": "uv run script.py"

// CORRECT - uses shell wrapper to load PATH
"command": "{{HOOK_SHELL_PREFIX}}uv run script.py{{HOOK_SHELL_SUFFIX}}"
```

### Platform-Specific Features

When platform-specific code is unavoidable:

```python
import platform

if platform.system() == "Windows":
    # Windows-specific implementation
    shell_prefix = ""
elif platform.system() == "Darwin":  # macOS
    # macOS-specific implementation (if needed)
    shell_prefix = '${SHELL:-/bin/bash} -l -c \''
else:  # Linux and others
    # Linux implementation
    shell_prefix = '${SHELL:-/bin/bash} -l -c \''
```

### Adding New Template Variables

When adding new cross-platform template variables to `update.py`:

1. **Detect Platform**:
   ```python
   is_windows = platform.system() == "Windows"
   ```

2. **Define Platform-Specific Values**:
   ```python
   if is_windows:
       new_variable = "windows_value"
   else:
       new_variable = "unix_value"
   ```

3. **Add to Template Context**:
   ```python
   template_vars = {
       ...
       "NEW_VARIABLE": new_variable,
   }
   ```

4. **Document in CLAUDE.local.md**:
   - Add to this section
   - Explain platform differences
   - Provide usage examples

### Verification Checklist

Before releasing cross-platform features:

- [ ] Tested on Windows (Git Bash or PowerShell)
- [ ] Tested on macOS with zsh (default shell)
- [ ] Tested on macOS with bash
- [ ] Tested on Linux with bash
- [ ] No hardcoded shell commands
- [ ] No hardcoded path separators
- [ ] Template variables used correctly
- [ ] Documentation updated

### Related Issues

- Issue #296: PATH loading problem across platforms
- v1.8.4: Fixed hooks with `bash -l -c` (macOS/Linux only)
- v1.8.5: Fixed double slash in paths
- v1.8.6: Implemented cross-platform shell wrapper (HOOK_SHELL_PREFIX/SUFFIX)

### References

- Update.py: Lines 1845-1857 (shell wrapper implementation)
- Template: `src/moai_adk/templates/.claude/settings.json`
- Agents: `src/moai_adk/templates/.claude/agents/moai/*.md`

---

## 21. CLAUDE.md Size Management

### 40k Character Limit

[HARD] CLAUDE.md must not exceed 40,000 characters (approximately 40KB).

**WHY**: Large CLAUDE.md files cause Claude to ignore instructions. Anthropic recommends keeping CLAUDE.md focused and modular.

### Current Status

- CLAUDE.md size: ~30.5KB (within limit)
- Buffer remaining: ~9.5KB
- Target: Stay under 35KB for safety margin

### Size Monitoring

Check CLAUDE.md size before commits:

```bash
# Check file size
wc -c CLAUDE.md

# Check character count
cat CLAUDE.md | wc -m

# Recommended: Keep under 35,000 characters
```

### Modularization Strategy

When CLAUDE.md approaches limit:

1. **Move detailed content to `.claude/rules/moai/`**:
   - Language-specific rules → `.claude/rules/moai/languages/`
   - Quality standards → `.claude/rules/moai/core/`
   - Workflow details → `.claude/rules/moai/workflow/`

2. **Use @import references**:
   - `@.claude/rules/moai/core/trust5-framework.md`
   - `@.moai/config/sections/quality.yaml`

3. **Keep CLAUDE.md for**:
   - Core identity and hard rules
   - Request processing pipeline
   - Agent catalog (summary only)
   - Configuration references

### Content Priority

When reducing CLAUDE.md size, remove in this order:

1. Detailed examples (move to skills)
2. Code blocks (move to rules)
3. Repetitive explanations (consolidate)
4. Historical context (move to docs)

**NEVER remove**:
- HARD rules
- Agent invocation patterns
- Quality gates checklist
- Version and metadata

### Sync with Template

After editing CLAUDE.md:

```bash
# Sync to template
cp CLAUDE.md src/moai_adk/templates/CLAUDE.md

# Verify size
wc -c src/moai_adk/templates/CLAUDE.md
```

### Rules Directory Structure

Claude Code official rules location: `.claude/rules/moai/`

```
.claude/rules/
└── moai/                        # MoAI-managed rules
    ├── core/                    # Core framework rules
    │   └── moai-constitution.md
    ├── workflow/                # Workflow mode rules
    │   ├── spec-workflow.md
    │   └── workflow-modes.md
    ├── development/             # Development standards
    │   ├── coding-standards.md
    │   └── skill-authoring.md
    └── languages/               # Path-specific language rules
        ├── python.md            # paths: **/*.py
        ├── typescript.md        # paths: **/*.ts, **/*.tsx
        └── ... (16 languages)
```

**Note**: Language rules use YAML frontmatter `paths` field for conditional loading.

### Deprecated Directories

The following directories were removed (non-standard):
- `.moai/rules/` (YAML format, not Claude Code standard)
- `.moai/contexts/` (custom pattern, not official)

Use `.claude/rules/moai/` (Markdown format) for all rule definitions.

---

**Status**: Active (Local Development)
**Version**: 3.7.0 (Update process improvements: deprecated folder cleanup, skill sync fix)
**Last Updated**: 2026-01-28


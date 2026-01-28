# v1.10.5 - Template Variable Substitution Fix (2026-01-29)

## Summary

Patch release fixing template variable substitution during project initialization.

**Key Fix**:
- Fixed template variables not being substituted during `moai init`
- Ensures proper cross-platform shell wrapper configuration

**Impact**:
- All template variables now correctly substituted during initialization
- Proper MCP server configuration on all platforms
- Consistent settings.json structure across platforms

## Fixed

### Template Variable Substitution

- **fix(template)**: Resolve template variable substitution issue (#304) (25b87aef)
  - Issue: `{{HOOK_SHELL_PREFIX}}`, `{{HOOK_SHELL_SUFFIX}}`, `{{MCP_SHELL}}` not substituted during `moai init`
  - Root cause: `_merge_settings_json` called before variable substitution in `_copy_claude` method
  - Fix: Apply variable substitution BEFORE merging settings.json
    - Read template content and substitute variables
    - Write to temporary file for merging
    - Apply additional substitution after merge for remaining variables
  - Impact: All shell wrapper variables now correctly substituted on all platforms
  - Files affected:
    - `src/moai_adk/core/template/processor.py` (lines 958-992)
    - `src/moai_adk/templates/.mcp.json` (formatting fix)

### Code Cleanup

- **chore**: Remove unused hook library modules (25b87aef)
  - Removed: `checkpoint.py`, `language_detector.py`, `language_validator.py`, `timeout.py`
  - Consolidated into `unified_timeout_manager.py`
  - Reduced codebase by ~1,000 lines

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates
moai update

# Verify version
moai --version
```

---

# v1.9.0 - Memory MCP, SVG Skill, Rules Migration (2026-01-26)

## Summary

Minor release introducing persistent memory across sessions, comprehensive SVG skill, and standards-compliant rules system migration.

**Key Features**:
- **Memory MCP Integration**: Persistent storage for user preferences and project context
- **SVG Skill**: Comprehensive skill with SVGO optimization patterns and best practices
- **Rules Migration**: Migrated from `.moai/rules/*.yaml` to `.claude/rules/*.md` (Claude Code official standard)
- **Bug Fix**: Rank batch sync display issue (#300)

**Impact**:
- Enables agent-to-agent context sharing via Memory MCP
- Professional SVG creation and optimization support
- Cleaner, standards-compliant project structure
- Accurate batch sync statistics display

## Breaking Changes

None. All changes are backward compatible.

## Added

### SVG Creation and Optimization Skill

- **feat**: Add `moai-tool-svg` skill (54c12a85)
  - Based on W3C SVG 2.0 specification and SVGO documentation
  - Comprehensive modules: basics, styling, optimization, animation
  - 12 working code examples
  - SVGO configuration patterns and best practices
  - 3,698 lines total (SKILL.md: 410, modules: 2,288, examples: 500, reference: 500)

### Language Rules Enhancement

- **feat**: Update language rules with enhanced tooling information (54c12a85)
  - Ruff configuration patterns (replaces flake8+isort+pyupgrade)
  - Mypy strict mode guidelines
  - Testing framework recommendations
  - 16 language files updated

## Changed

### CLAUDE.md Optimization

- **refactor**: Major cleanup and modularization for v1.9.0 (4134e60d)
  - Reduced CLAUDE.md from ~60k to ~30k characters (40k limit compliance)
  - Moved detailed content to `.claude/rules/` for better organization
  - Added `shell_validator.py` utility for cross-platform compatibility
  - Enhanced CLI commands (doctor, init, update)
  - Added `moai-workflow-thinking` skill
  - Added bug-report.yml issue template
  - Impact: Improved readability, maintainability, and Claude Code compatibility

### Rules System Migration

- **feat**: Migrate from `.moai/rules/*.yaml` to `.claude/rules/*.md` (99ab5273)
  - Deleted: 6,959 lines of YAML rules
  - Added: Claude Code official Markdown rules
  - Structure: `.claude/rules/{core,development,workflow,languages}/`
  - Impact: Standards compliance, cleaner organization

## Fixed

### Rank Command

- **fix(rank)**: Correctly parse nested API response for batch sync (#300) (31b504ed)
  - Issue: `moai-adk rank sync` always showed "Submitted: 0"
  - Root cause: Missing nested `data` field extraction
  - Fix: Added `data = response.get("data", {})` before accessing fields
  - Impact: Accurate submission statistics display

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates in your folder
moai update

# Verify version
moai --version
```

---

# v1.8.13 - Statusline Context Window Fix (2026-01-26)

## Summary

Patch release improving statusline context window calculation accuracy.

**Key Fix**:
- Fixed statusline context window percentage to use Claude Code's pre-calculated values

**Impact**:
- Context window display now accounts for auto-compact and output token reservation
- More accurate remaining token information

## Fixed

### Statusline Context Window Calculation

- **fix(statusline)**: Use Claude Code's pre-calculated context percentages (2dacecb7)
  - Priority 1: Use `used_percentage`/`remaining_percentage` from Claude Code (most accurate)
  - Priority 2: Calculate from `current_usage` tokens (fallback)
  - Priority 3: Return 0% when no data available (session start)
  - Ensures accuracy when auto-compact is enabled or output tokens are reserved
  - Files: `src/moai_adk/statusline/main.py`

## Installation & Update

```bash
# Update to the latest version
uv tool update moai-adk

# Update project templates
moai update

# Verify version
moai --version
```

---

# v1.8.12 - Hook Format Update & Login Command (2026-01-26)

## Summary

Patch release with Claude Code hook format compatibility fix and UX improvements.

**Key Changes**:
- Fixed Claude Code settings.json hook format (new matcher-based structure)
- Renamed `moai rank register` to `moai rank login` (more intuitive)
- settings.json now always overwritten on update; use settings.local.json for customizations

**Impact**:
- MoAI Rank hooks now work with latest Claude Code
- `moai rank login` is the new primary command (register still works as alias)
- User customizations preserved in settings.local.json

## Breaking Changes

None. `moai rank register` still works as a hidden alias.

## Fixed

# SPEC-UPDATE-REFACTOR-002 Codebase Exploration Report

## 1. Current Update Command Implementation

**Location:** `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/cli/commands/update.py`

### Command Entry Point (Lines 243-260)
```python
@click.command()
@click.option(
    "--path",
    type=click.Path(exists=True),
    default=".",
    help="Project path (default: current directory)"
)
@click.option(
    "--force",
    is_flag=True,
    help="Skip backup and force the update"
)
@click.option(
    "--check",
    is_flag=True,
    help="Only check version (do not update)"
)
def update(path: str, force: bool, check: bool) -> None:
    """Update template files to the latest version."""
```

### Core Functions in update.py

| Function | Lines | Purpose | Key Notes |
|----------|-------|---------|-----------|
| `get_latest_version()` | 51-68 | Fetches version from PyPI JSON API | Uses urllib, handles timeouts, returns None on failure |
| `set_optimized_false()` | 71-90 | Sets config.json optimized field to false | Modifies project config after update |
| `_load_existing_config()` | 93-101 | Loads existing config.json | Returns empty dict if not found |
| `_is_placeholder()` | 104-106 | Checks if value is unsubstituted template placeholder | Pattern: `{{...}}` |
| `_coalesce()` | 109-121 | Returns first non-empty, non-placeholder value | Smart fallback with default |
| `_extract_project_section()` | 124-129 | Extracts nested project config section | Safe dict extraction |
| `_build_template_context()` | 132-174 | Builds substitution context for templates | Creates template variables dict |
| `_preserve_project_metadata()` | 177-221 | Restores project metadata in new config | Restores name, mode, description, locale, language |
| `_apply_context_to_file()` | 224-240 | Post-merge template variable substitution | Uses processor context |
| `update()` | 260-393 | Main command handler | 4 phases: check versions, backup, update templates, finalize |

### Update Command Workflow (Lines 274-393)

**Phase 1: Version Check (Lines 285-311)**
- Fetches current and latest versions
- Handles PyPI fetch failure gracefully
- Exits early if `--check` flag provided
- Compares using `packaging.version.parse()`

**Phase 2: Backup Creation (Lines 344-351)**
- Skipped if `--force` flag provided
- Uses `TemplateProcessor.create_backup()`
- Returns relative path for display

**Phase 3: Template Update (Lines 353-369)**
- Calls `processor.copy_templates(backup=False, silent=True)`
- Builds template context from existing config
- Preserves project metadata
- Applies context to CLAUDE.md post-merge

**Phase 4: Finalization (Lines 371-389)**
- Sets `optimized=false` in config
- Shows upgrade instructions if needed
- Suggests `/alfred:0-project update` command

---

## 2. Key Implementation Patterns

### Error Handling Pattern
```python
try:
    # Main logic
except click.Abort:
    # User cancelled (Ctrl+C)
except click.ClickException as e:
    e.show()
except Exception as e:
    console.print(f"[red]✗ Error: {e}[/red]")
    raise click.ClickException(str(e)) from e
```

### Version Comparison
```python
from packaging import version
current_v = version.parse(current_version)
latest_v = version.parse(latest_version)
if current_v < latest_v:
    # Upgrade needed
```

---

## Summary of Key Patterns

**For implementation guidance**, see `implementation-UPDATE-REFACTOR-002.md` in this directory.

**For quick reference**, see `codebase-exploration-index.md`.
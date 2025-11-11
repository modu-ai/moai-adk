# moai-alfred-document-management

**Purpose**: Guide agents on proper file location and document management rules.

**Version**: 1.0.0
**Last Updated**: 2025-11-12
**Category**: Infrastructure Management

---

## Quick Decision Tree

```
Creating a file?
├─ Is it README/CHANGELOG/CONTRIBUTING/LICENSE? → Root OK
├─ Is it package.json/pyproject.toml? → Root OK
├─ Is it a report/analysis? → .moai/reports/{category}/
├─ Is it a script? → .moai/scripts/{category}/
├─ Is it temporary? → .moai/temp/{category}/
├─ Is it a backup? → .moai/backups/{type}/
└─ Is it a log? → .moai/logs/{type}/
```

---

## Auto-Categorization Rules

### Reports

| Pattern | Category | Full Path |
|---------|----------|-----------|
| `FINAL-INSPECTION-*.md` | `inspection` | `.moai/reports/inspection/` |
| `PHASE*-COMPLETION-*.md` | `phases` | `.moai/reports/phases/` |
| `sync-report-*.md` | `sync` | `.moai/reports/sync/` |
| `*-ANALYSIS-*.md` | `analysis` | `.moai/reports/analysis/` |
| `*-validation-*.md` | `validation` | `.moai/reports/validation/` |
| `daily-*.md` | `daily` | `.moai/reports/daily/` |

### Scripts

| Pattern | Category | Full Path |
|---------|----------|-----------|
| `init-*.sh`, `setup-*.sh` | `dev` | `.moai/scripts/dev/` |
| `fix-*.js`, `convert-*.py` | `conversion` | `.moai/scripts/conversion/` |
| `validate_*.py`, `lint_*.py` | `validation` | `.moai/scripts/validation/` |
| `*_analyzer.py`, `analyze_*.py` | `analysis` | `.moai/scripts/analysis/` |
| `cleanup_*.sh`, `maintain_*.py` | `maintenance` | `.moai/scripts/maintenance/` |

### Temp Files

| Pattern | Category | Full Path |
|---------|----------|-----------|
| `test-*.spec.js` | `tests` | `.moai/temp/tests/` |
| `coverage.json`, `.coverage` | `coverage` | `.moai/temp/coverage/` |
| `*.tmp`, `*.temp`, `*.bak` | `work` | `.moai/temp/work/` |

### Backups

| Pattern | Category | Full Path |
|---------|----------|-----------|
| `docs_backup_*`, `docs-backup-*` | `docs` | `.moai/backups/docs/` |
| `hooks_backup_*` | `hooks` | `.moai/backups/hooks/` |
| `specs_backup_*` | `specs` | `.moai/backups/specs/` |
| `config_backup_*` | `config` | `.moai/backups/config/` |

---

## Example Usage

### Python Example

```python
# Before (❌ Root pollution)
report_path = "FINAL-INSPECTION-REPORT.md"

# After (✅ Correct location)
from datetime import datetime
report_path = f".moai/reports/inspection/FINAL-INSPECTION-REPORT-{datetime.now().strftime('%Y-%m-%d')}.md"
```

### JavaScript Example

```javascript
// Before (❌ Root pollution)
const scriptPath = "fix-links.js";

// After (✅ Correct location)
const scriptPath = ".moai/scripts/conversion/fix-links.js";
```

### Bash Example

```bash
# Before (❌ Root pollution)
backup_dir="docs_backup_$(date +%Y%m%d)"

# After (✅ Correct location)
backup_dir=".moai/backups/docs/docs_backup_$(date +%Y%m%d)"
```

---

## Configuration Reference

Check `.moai/config/config.json` → `document_management`:

```json
{
  "document_management": {
    "enabled": true,
    "enforce_structure": true,
    "block_root_pollution": false,
    "validation": {
      "warn_violations": true,
      "block_violations": false
    }
  }
}
```

### Configuration Options

| Option | Default | Description |
|--------|---------|-------------|
| `enabled` | `true` | Master switch for document management |
| `enforce_structure` | `true` | Enforce `.moai/` hierarchy |
| `block_root_pollution` | `false` | Block non-whitelisted root files |
| `warn_violations` | `true` | Warn on violations |
| `block_violations` | `false` | Block file creation on violations |

---

## Common Scenarios

### Scenario 1: Creating a Report

**Question**: Where should I create a quality validation report?

**Answer**:
```
File pattern: quality-validation-report.md
Matches: *-validation-*.md
Category: validation
Full path: .moai/reports/validation/quality-validation-report.md
```

### Scenario 2: Creating a Script

**Question**: Where should I create a data conversion script?

**Answer**:
```
File pattern: convert-data.py
Matches: convert-*.py
Category: conversion
Full path: .moai/scripts/conversion/convert-data.py
```

### Scenario 3: Temporary Test File

**Question**: Where should I create a temporary test?

**Answer**:
```
File pattern: test-feature.spec.js
Matches: test-*.spec.js
Category: tests
Full path: .moai/temp/tests/test-feature.spec.js
Auto-cleanup: Yes (7 days retention)
```

### Scenario 4: Backup Directory

**Question**: Where should I create a backup of hooks?

**Answer**:
```
Pattern: hooks_backup_YYYYMMDD_HHMMSS/
Matches: hooks_backup_*
Category: hooks
Full path: .moai/backups/hooks/hooks_backup_20251112_150000/
```

---

## Error Messages

### Root pollution blocked

```
❌ Error: Root pollution blocked
File: FINAL-INSPECTION-REPORT.md
Reason: Not in root_whitelist
✅ Suggested: .moai/reports/inspection/FINAL-INSPECTION-REPORT.md

Config setting: document_management.block_root_pollution = true
To disable: Set block_root_pollution to false in .moai/config/config.json
```

### Auto-migration suggestion

```
⚠️ Found 3 misplaced files:
1. fix-internal-links.js → .moai/scripts/conversion/
2. coverage.json → .moai/temp/coverage/
3. docs_backup_20251112/ → .moai/backups/docs/

Action: Run migration or move files manually
```

---

## Agent Guidelines

### Before Creating ANY File

1. **Check whitelist**: Is the file in `root_whitelist`?
2. **Match pattern**: Does the filename match any category pattern?
3. **Determine location**: Use auto-categorization rules
4. **Create in correct location**: Never default to root

### When Uncertain

1. Check `document_management.file_patterns` in config.json
2. Consult this Skill for pattern matching
3. Default to `.moai/temp/work/` and ask user
4. **NEVER** default to project root

### Agent-Specific Responsibilities

| Agent | Primary Responsibility |
|-------|----------------------|
| **report-generator** | Always use `.moai/reports/{category}/` |
| **file-manager** | Auto-categorize using patterns |
| **backup-manager** | Always use `.moai/backups/{type}/` |
| **script-creator** | Always use `.moai/scripts/{category}/` |
| **test-engineer** | Always use `.moai/temp/tests/` |
| **doc-syncer** | Check location before creating docs |

---

## Auto-Cleanup Policy

### Automatic Cleanup

| Directory | Retention | Schedule |
|-----------|-----------|----------|
| `.moai/temp/` | 7 days | `session_end` |
| `.moai/cache/` | 30 days | `session_end` |
| `.moai/logs/sessions/` | 30 days | `session_end` |
| `.moai/backups/` | 90 days | `weekly` |

### Manual Review Required

| Directory | Retention | Reason |
|-----------|-----------|--------|
| `.moai/reports/` | 90 days | Important analysis |
| `.moai/scripts/` | Permanent | Reusable utilities |

### Cleanup Configuration

```json
{
  "cleanup": {
    "enabled": true,
    "schedule": "session_end",
    "preserve_recent_days": 3,
    "max_items_per_cleanup": 100
  }
}
```

---

## Root Whitelist

**Standard Project Files** (ALWAYS allowed in root):
- `README.md`, `README.*.md`
- `CHANGELOG.md`, `CONTRIBUTING.md`, `CLAUDE.md`
- `LICENSE`, `LICENSE.*`

**Configuration Files** (ALWAYS allowed in root):
- `pyproject.toml`, `setup.py`, `setup.cfg`
- `package.json`, `package-lock.json`, `yarn.lock`
- `.gitignore`, `.editorconfig`, `.prettierrc`
- `Makefile`, `Dockerfile`, `docker-compose.yml`

**Everything Else**: Must go in `.moai/` hierarchy

---

## Troubleshooting

### Problem: "Where should I create this file?"

**Solution**:
1. Check filename against patterns in this Skill
2. If matches a pattern → use suggested category
3. If no match → check if standard project file
4. If neither → place in `.moai/temp/work/` and ask user

### Problem: "Can I create files in root?"

**Solution**:
- Check `root_whitelist` in config.json
- If file is in whitelist → YES
- If not in whitelist → NO, use `.moai/` hierarchy

### Problem: "How do I handle existing root pollution?"

**Solution**:
1. Identify violating files (not in whitelist)
2. Match against `file_patterns` for correct location
3. Move files to suggested locations
4. Update any references to old paths

---

## Version History

### v1.0.0 (2025-11-12)
- Initial release
- Complete directory structure definition
- Auto-categorization rules
- Configuration reference
- Cleanup policies

---

## Related Skills

- `moai-alfred-best-practices`: Overall Alfred best practices
- `moai-alfred-reporting`: Report generation guidelines
- `moai-alfred-backup-strategy`: Backup management

---

## Notes

- **Philosophy**: Keep project root clean for better maintainability
- **Enforcement**: Configurable (warn vs block mode)
- **Auto-cleanup**: Preserves recent files, removes old temp files
- **Traceability**: All file operations logged for audit

**Remember**: When in doubt, place in `.moai/temp/` and ask user!
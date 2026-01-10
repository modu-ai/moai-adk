---
description: "Auto-fix current LSP errors and AST-grep warnings"
argument-hint: "[--errors-only] [--dry-run] [file_path]"
type: utility
allowed-tools: Task, AskUserQuestion, Bash, Read, Write, Edit, Glob, Grep
model: inherit
---

## Pre-execution Context

!git status --porcelain
!git diff --name-only HEAD

## Essential Files

@.moai/config/sections/ralph.yaml

---

# /moai:fix - Automatic Error and Warning Fixer

Automatically detect and fix LSP errors, linting issues, and AST-grep warnings in the current project or specific files.

## Command Purpose

Provides one-command fixing for common code issues:

1. Scans for LSP diagnostics (errors and warnings)
2. Identifies AST-grep security and quality issues
3. Applies safe auto-fixes where possible
4. Reports issues requiring manual intervention

Target: $ARGUMENTS

## Usage Examples

Fix all issues in project:

```
/moai:fix
```

Fix specific file:

```
/moai:fix src/auth.py
```

Fix errors only (ignore warnings):

```
/moai:fix --errors-only
```

Preview fixes without applying:

```
/moai:fix --dry-run
```

## Command Options

- `file_path`: Optional specific file or directory to fix
- `--errors-only`: Only fix errors, skip warnings and hints
- `--dry-run`: Show what would be fixed without making changes
- `--include-security`: Include AST-grep security fixes
- `--no-format`: Skip auto-formatting after fixes

## Fix Categories

### Category 1: Auto-Fixable (Applied Automatically)

These issues are safe to fix automatically:

**Python**:

- Import sorting (isort/ruff)
- Unused imports removal
- Whitespace/formatting (black/ruff)
- Simple type annotations
- f-string conversions
- Dictionary comprehension suggestions

**TypeScript/JavaScript**:

- Import organization
- Unused variable removal (when safe)
- Formatting (prettier)
- Simple ESLint auto-fixes

**General**:

- Trailing whitespace
- Missing newlines at EOF
- Mixed indentation

### Category 2: Semi-Auto (Require Confirmation)

These fixes are suggested but need user approval:

- Renaming to fix typos
- Adding missing function parameters
- Changing return types
- Modifying exception handling
- Updating deprecated API calls

### Category 3: Manual (Report Only)

These issues require manual intervention:

- Logic errors
- Security vulnerabilities
- Architectural issues
- Complex type mismatches
- Missing implementations

## Execution Flow

```
START: /moai:fix [options] [target]

PHASE 1: SCAN
  |-- Run LSP diagnostics on target
  |-- Run AST-grep security scan
  |-- Run linter checks
  |-- Collect all issues
  |
  v
PHASE 2: CATEGORIZE
  |-- Sort issues by fixability
  |-- Group by file
  |-- Prioritize by severity
  |
  v
PHASE 3: DRY-RUN CHECK
  |-- IF --dry-run: Report and EXIT
  |
  v
PHASE 4: AUTO-FIX
  |-- Apply Category 1 fixes
  |-- Run formatters
  |-- Re-scan for remaining issues
  |
  v
PHASE 5: SEMI-AUTO
  |-- Present Category 2 fixes
  |-- Get user approval for each
  |-- Apply approved fixes
  |
  v
PHASE 6: REPORT
  |-- List remaining manual issues
  |-- Show fix summary
  |-- Suggest next steps

END: Summary with remaining issues
```

## Integration with Tools

### LSP Integration

Uses `MoAILSPClient` for language-aware diagnostics:

- Python: pyright/pylsp
- TypeScript: tsserver
- Go: gopls
- Rust: rust-analyzer

### AST-grep Integration

Runs security and quality scans:

- SQL injection patterns
- XSS vulnerabilities
- Insecure configurations
- Code smell patterns

### Linter Integration

Leverages existing linters via `tool_registry.py`:

- ruff (Python)
- eslint/biome (TypeScript/JavaScript)
- golangci-lint (Go)
- clippy (Rust)

## Output Format

### Dry-Run Output

```markdown
## Dry Run: /moai:fix

### Would Fix Automatically (12 issues)

- src/auth.py: 4 issues (unused imports, formatting)
- src/api/routes.py: 3 issues (import order)
- tests/test_auth.py: 5 issues (whitespace)

### Would Require Confirmation (3 issues)

- src/auth.py:45 - Rename 'usr' to 'user'?
- src/api/routes.py:78 - Add return type annotation?
- src/models.py:23 - Convert to dataclass?

### Manual Fixes Needed (2 issues)

- [ERROR] src/auth.py:67 - Logic error in token validation
- [SECURITY] src/api/routes.py:112 - Potential SQL injection

No changes made (dry run).
```

### Fix Completion Output

```markdown
## Fix Complete: /moai:fix

### Applied Fixes (15 total)

- Auto-fixed: 12 issues
- User-approved: 3 issues
- Skipped: 0 issues

### Files Modified

- src/auth.py (7 fixes)
- src/api/routes.py (5 fixes)
- tests/test_auth.py (3 fixes)

### Remaining Issues (2 manual)

1. [ERROR] src/auth.py:67
   Logic error in token validation
   Suggestion: Review the condition on line 67

2. [SECURITY] src/api/routes.py:112
   Potential SQL injection
   Suggestion: Use parameterized queries

### Verification

- Linter: PASS
- Tests: 24/24 passing
- Coverage: 87%
```

## Error Handling

### If LSP Unavailable

Falls back to linter-based detection:

- Uses ruff for Python
- Uses tsc for TypeScript
- Reports degraded mode to user

### If Fixes Cause New Issues

- Detects regression
- Reverts problematic fixes
- Reports to user with details

### If Tests Fail After Fixes

- Warns user
- Suggests running full test suite
- Offers to revert changes

## Best Practices

1. **Run Tests First**: Ensure tests pass before fixing
2. **Use Dry-Run**: Preview changes with `--dry-run` first
3. **Commit Before Fixing**: Have a clean commit to revert to
4. **Review Changes**: Always review auto-fixes before committing
5. **Fix Incrementally**: Fix one category at a time for complex projects

## Related Commands

- `/moai:loop`: Continuous fix loop until all issues resolved
- `/moai:cancel-loop`: Stop an active fix loop
- `/moai:2-run`: TDD implementation with integrated fixing

---

## Execution Directive

[HARD] Execute the fix workflow immediately upon command invocation:

1. Scan for issues using LSP diagnostics and AST-grep
2. Categorize issues by fixability (auto, semi-auto, manual)
3. Apply auto-fixes without user intervention
4. Present semi-auto fixes for user approval via AskUserQuestion
5. Report remaining manual issues

Note: This command uses direct tool access (Edit, Write) intentionally for quick one-time fixes. For iterative fixing with full agent support, use `/moai:loop` instead.

---

Version: 1.0.0
Last Updated: 2026-01-10
Pattern: Scan-Categorize-Fix
Integration: LSP, AST-grep, Linters, Formatters

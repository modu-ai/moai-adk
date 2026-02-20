# Workflow: MX Tag Scan and Annotation

Purpose: Scan codebase and add @MX code-level annotations for AI agent context.

## When to Use

- Legacy codebase without @MX tags
- Before major refactoring to mark danger zones
- After significant code changes to update annotations
- During `/moai sync` for MX validation

## Command

```
/moai mx [options]
```

## Flags

| Flag | Description |
|------|-------------|
| `--all` | Scan entire codebase (not just changed files) |
| `--dry` | Preview only - show tags to add without modifying files |
| `--priority P1-P4` | Filter by priority level (default: all) |
| `--force` | Overwrite existing @MX tags |
| `--exclude pattern` | Additional exclude patterns (comma-separated) |

## Priority Levels

| Priority | Condition | Tag Type |
|----------|-----------|----------|
| P1 | fan_in >= 5 callers | `@MX:ANCHOR` |
| P2 | goroutine, complexity >= 15 | `@MX:WARN` |
| P3 | magic constant, missing godoc | `@MX:NOTE` |
| P4 | missing test | `@MX:TODO` |

## 3-Pass Algorithm

### Pass 1: Grep Full Scan (10-30 seconds)

```
1. Fan-in analysis: Count function name references across files
2. Goroutine detection: Search for `go func`, `go ` patterns
3. Magic constant detection: 3+ digit numbers, decimal fractions
4. Exported functions without godoc
5. Output: Priority queue P1-P4
```

### Pass 2: Selective Deep Read (P1 files only)

```
1. Full file Read for each P1-priority file
2. Generate accurate @MX:NOTE and @MX:ANCHOR descriptions
3. Understand goroutine lifecycle for @MX:WARN descriptions
```

### Pass 3: Batch Edit

```
1. One Edit call per file
2. All tags for a given file inserted in single operation
3. Generate tag report
```

## Output

After completion, generates report:

```markdown
## @MX Tag Report

### Summary
- Files scanned: N
- Tags added: N
- Tags updated: N
- Tags skipped (existing): N

### Tags by Type
- @MX:ANCHOR: N (P1)
- @MX:WARN: N (P2)
- @MX:NOTE: N (P3)
- @MX:TODO: N (P4)

### Files Modified
- path/to/file.go: +3 tags
- path/to/another.go: +2 tags

### Attention Required
- High fan_in functions (>= 10 callers)
- Complex functions (complexity >= 20)
```

## Integration with Other Workflows

### With /moai sync

During sync phase, MX validation runs automatically:
1. Scan files modified since last sync
2. Check for missing @MX tags in modified functions
3. Add tags if `--skip-mx` flag not provided
4. Include tag changes in sync report

### With /moai run

During DDD ANALYZE phase:
1. If codebase has zero @MX tags, 3-Pass auto-triggers
2. Existing tags are validated and updated
3. New tags added for new code

## Configuration

Project settings in `.mx.yaml`:

```yaml
mx:
  version: "1.0"
  exclude:
    - "**/*_generated.go"
    - "**/vendor/**"
    - "**/mock_*.go"
  limits:
    anchor_per_file: 3
    warn_per_file: 5
  thresholds:
    fan_in_anchor: 5
    complexity_warn: 15
    branch_warn: 8
  auto_tag: true
```

## Examples

```bash
# Scan entire codebase
/moai mx --all

# Preview tags without modifying files
/moai mx --dry

# Only P1 priority (high fan_in functions)
/moai mx --priority P1

# Force overwrite existing tags
/moai mx --all --force

# Exclude test files
/moai mx --all --exclude "**/*_test.go"
```

---

Version: 1.0.0
Last Updated: 2026-02-20
Source: SPEC-MX-001

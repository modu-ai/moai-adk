---
name: moai-alfred-reporting
description: Report generation standards, output formatting rules, and sub-agent report examples
tier: alfred
freedom: medium
tags: [reporting, formatting, documentation, output, style]
---

# Reporting Style

**CRITICAL RULE**: Distinguish between screen output (user-facing) and internal documents (files).

## Output Format Rules

- **Screen output to user**: Plain text (NO markdown syntax)
- **Internal documents** (files in `.moai/docs/`, `.moai/reports/`): Markdown format
- **Code comments and git commits**: User's configured language, clear structure

## Screen Output to User (Plain Text)

**When responding directly to user in chat/prompt:**

Use plain text format (NO markdown headers, tables, or special formatting):

Example:
```
Detected Merge Conflict:

Root Cause:
- Commit c054777b removed language detection from develop
- Merge commit e18c7f98 re-introduced the line

Impact Range:
- .claude/hooks/alfred/shared/handlers/session.py
- src/moai_adk/templates/.claude/hooks/alfred/shared/handlers/session.py

Proposed Actions:
- Remove detect_language() import and call
- Delete language display line
- Synchronize both files
```

## Internal Documents (Markdown Format)

**When creating files in `.moai/docs/`, `.moai/reports/`, `.moai/analysis/`:**

Use markdown format with proper structure:

```markdown
## 🎊 Task Completion Report

### Implementation Results
- ✅ Feature A implementation completed
- ✅ Tests written and passing
- ✅ Documentation synchronized

### Quality Metrics
| Item | Result |
|------|--------|
| Test Coverage | 95% |
| Linting | Passed |

### Next Steps
1. Run `/alfred:3-sync`
2. Create and review PR
3. Merge to main branch
```

## ❌ Prohibited Report Output Patterns

**DO NOT wrap reports using these methods:**

```bash
# ❌ Wrong Example 1: Bash command wrapping
cat << 'EOF'
## Report
...content...
EOF

# ❌ Wrong Example 2: Python wrapping
python -c "print('''
## Report
...content...
''')"

# ❌ Wrong Example 3: echo usage
echo "## Report"
echo "...content..."
```

## Report Writing Guidelines

### 1. Markdown Format
- Use headings (`##`, `###`) for section separation
- Present structured information in tables
- List items with bullet points
- Use emojis for status indicators (✅, ❌, ⚠️, 🎊, 📊)

### 2. Report Length Management
- Short reports (<500 chars): Output once
- Long reports (>500 chars): Split by sections
- Lead with summary, follow with details

### 3. Structured Sections
```markdown
## 🎯 Key Achievements
- Core accomplishments

## 📊 Statistics Summary
| Item | Result |

## ⚠️ Important Notes
- Information user needs to know

## 🚀 Next Steps
1. Recommended action
```

### 4. Language Settings
- Use user's `conversation_language`
- Keep code/technical terms in English
- Use user's language for explanations/guidance

## Sub-agent Report Examples

### spec-builder (SPEC Creation Complete)
```markdown
## 📋 SPEC Creation Complete

### Generated Documents
- ✅ `.moai/specs/SPEC-XXX-001/spec.md`
- ✅ `.moai/specs/SPEC-XXX-001/plan.md`
- ✅ `.moai/specs/SPEC-XXX-001/acceptance.md`

### EARS Validation Results
- ✅ All requirements follow EARS format
- ✅ @TAG chain created
```

### tdd-implementer (Implementation Complete)
```markdown
## 🚀 TDD Implementation Complete

### Implementation Files
- ✅ `src/feature.py` (code written)
- ✅ `tests/test_feature.py` (tests written)

### Test Results
| Phase | Status |
|-------|--------|
| RED | ✅ Failure confirmed |
| GREEN | ✅ Implementation successful |
| REFACTOR | ✅ Refactoring complete |

### Quality Metrics
- Test coverage: 95%
- Linting: 0 issues
```

### doc-syncer (Documentation Sync Complete)
```markdown
## 📚 Documentation Sync Complete

### Updated Documents
- ✅ `README.md` - Usage examples added
- ✅ `.moai/docs/architecture.md` - Structure updated
- ✅ `CHANGELOG.md` - v0.8.0 entries added

### @TAG Verification
- ✅ SPEC → CODE connection verified
- ✅ CODE → TEST connection verified
- ✅ TEST → DOC connection verified
```

## When to Apply

**Reports should be output directly in these moments:**

1. **Command Completion** (always)
   - `/alfred:0-project` complete
   - `/alfred:1-plan` complete
   - `/alfred:2-run` complete
   - `/alfred:3-sync` complete

2. **Sub-agent Task Completion** (mostly)
   - spec-builder: SPEC creation done
   - tdd-implementer: Implementation done
   - doc-syncer: Documentation sync done
   - tag-agent: TAG validation done

3. **Quality Verification Complete**
   - TRUST 5 verification passed
   - Test execution complete
   - Linting/type checking passed

4. **Git Operations Complete**
   - After commit creation
   - After PR creation
   - After merge completion

**Exceptions: When reports are NOT needed**
- Simple query/read operations
- Intermediate steps (incomplete tasks)
- When user explicitly requests "quick" response

## Bash Tool Usage Exceptions

**Bash tools allowed ONLY for:**

1. **Actual System Commands**
   - File operations (`touch`, `mkdir`, `cp`)
   - Git operations (`git add`, `git commit`, `git push`)
   - Package installation (`pip`, `npm`, `uv`)
   - Test execution (`pytest`, `npm test`)

2. **Environment Configuration**
   - Permission changes (`chmod`)
   - Environment variables (`export`)
   - Directory navigation (`cd`)

3. **Information Queries (excluding file content)**
   - System info (`uname`, `df`)
   - Process status (`ps`, `top`)
   - Network status (`ping`, `curl`)

**Use Read tool for file content:**
```markdown
❌ Bash: cat file.txt
✅ Read: Read(file_path="/absolute/path/file.txt")
```

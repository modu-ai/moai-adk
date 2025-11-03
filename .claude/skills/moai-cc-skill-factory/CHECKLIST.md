# Skill Quality Validation Checklist

This checklist ensures every Skill meets MoAI-ADK quality standards before publication. Use this guide for pre-publication audits and release validation.

---

## Pre-Publication Audit (40 Points)

Complete this checklist before committing a new Skill to the repository.

### Metadata Completeness (8 Points)

- [ ] **name** field present
  - [ ] ≤ 64 characters
  - [ ] Gerund format (action verb)
  - [ ] Specific and discoverable (not generic like "Python Helper")
  - Points: 2

- [ ] **description** field present
  - [ ] ≤ 1024 characters
  - [ ] Includes 3+ discoverable keywords
  - [ ] Describes both WHAT (capabilities) and WHEN (use cases)
  - [ ] Written in third person ("The Skill does X")
  - Points: 3

- [ ] **allowed-tools** field present
  - [ ] Tools actually used by Skill
  - [ ] Minimal scope (no overly broad patterns)
  - [ ] Properly comma-separated
  - Points: 1

- [ ] YAML frontmatter valid
  - [ ] Starts with `---`
  - [ ] Ends with `---`
  - [ ] No tabs (spaces only)
  - Points: 2

**Subtotal**: 8 points

### Content Quality (16 Points)

- [ ] **SKILL.md completeness**
  - [ ] ≤ 500 lines total
  - [ ] Clear overview/introduction section
  - [ ] Progressive Disclosure pattern applied
  - [ ] Relative paths used for cross-file links
  - Points: 4

- [ ] **Examples included**
  - [ ] At least 1 example per major concept
  - [ ] Concrete code or pseudocode shown
  - [ ] Examples match skill's freedom level
  - Points: 2

- [ ] **Terminology consistency**
  - [ ] Key terms defined once (glossary section)
  - [ ] Same term used consistently (no synonyms)
  - [ ] No conflicting definitions
  - Points: 2

- [ ] **Anti-patterns avoided**
  - [ ] No time-sensitive information
  - [ ] No Windows-style paths (`\`)
  - [ ] No absolute paths
  - [ ] No vague instructions ("make it fast")
  - Points: 2

- [ ] **File organization**
  - [ ] One level deep (no nested subdirs)
  - [ ] Forward slashes in all paths
  - [ ] Clear file naming (kebab-case for scripts)
  - Points: 2

- [ ] **Supporting files (if present)**
  - [ ] reference.md has clear structure
  - [ ] examples.md demonstrates 3-4 scenarios
  - [ ] Files progressively disclosed (not duplicating SKILL.md)
  - Points: 2

**Subtotal**: 16 points

### Multi-Model Compatibility (8 Points)

Test Skill with Haiku, Sonnet, and Opus (or equivalent):

- [ ] **Haiku activation test**
  - [ ] Skill activates on relevant request
  - [ ] Simple examples understood
  - [ ] Can follow basic patterns
  - Points: 2

- [ ] **Sonnet full capability test**
  - [ ] Exploits full Skill depth
  - [ ] Connects concepts across sections
  - [ ] Uses all provided patterns
  - Points: 3

- [ ] **Opus extension test** (if Opus available)
  - [ ] Can extend patterns beyond examples
  - [ ] Applies concepts to novel scenarios
  - [ ] Integrates with broader knowledge
  - Points: 3

**Subtotal**: 8 points

### Security & Compliance (8 Points)

- [ ] **No sensitive data**
  - [ ] No credentials, API keys, secrets
  - [ ] No private email addresses
  - [ ] No internal URLs or IPs
  - Points: 2

- [ ] **Scripts security**
  - [ ] Error handling included (`set -euo pipefail` for Bash)
  - [ ] Input validation present
  - [ ] Safe file operations (with backups)
  - [ ] Exit codes properly used
  - Points: 2

- [ ] **No risky assumptions**
  - [ ] No "assuming ~/config exists"
  - [ ] No system-state dependencies
  - [ ] Platform-agnostic paths
  - Points: 2

- [ ] **Destructive operations protected**
  - [ ] `rm`, `mv`, `mv` require confirmation
  - [ ] Rollback options documented
  - [ ] Warnings clearly marked
  - Points: 2

**Subtotal**: 8 points

---

## Verification Checklist (30 Points)

### File Structure (6 Points)

- [ ] Root directory named correctly (skill-name format)
- [ ] SKILL.md present in root
- [ ] All files at one level (no nesting)
- [ ] scripts/ subdirectory (if scripts present)
- [ ] templates/ subdirectory (if templates present)
- [ ] No extra directories or files
  - Points: 6

### File Validity (6 Points)

- [ ] All `.md` files are valid Markdown
  - [ ] Headers properly formatted (`#`, `##`, etc.)
  - [ ] Code blocks properly fenced (` ``` `)
  - [ ] Links properly formatted (`[text](path)`)
  - Points: 2

- [ ] All scripts have proper shebang
  - [ ] Bash: `#!/bin/bash`
  - [ ] Python: `#!/usr/bin/env python3`
  - [ ] Permissions: 755 or `chmod +x`
  - Points: 2

- [ ] YAML frontmatter parses correctly
  - [ ] No syntax errors
  - [ ] Proper indentation
  - [ ] All required fields present
  - Points: 2

### Progressive Disclosure (6 Points)

- [ ] **Level 1 (Metadata)**
  - [ ] Name, description, allowed-tools complete
  - [ ] YAML valid
  - Points: 2

- [ ] **Level 2 (Main Instructions)**
  - [ ] SKILL.md standalone and complete
  - [ ] Can understand without reading Level 3
  - Points: 2

- [ ] **Level 3 (Resources)**
  - [ ] Supporting files referenced but not required
  - [ ] Scripts/templates optional enhancements
  - [ ] No circular dependencies
  - Points: 2

### Reference Accuracy (6 Points)

- [ ] All links valid
  - [ ] `[text](reference.md)` links work
  - [ ] No broken internal links
  - [ ] No external links (except documentation)
  - Points: 2

- [ ] All file paths relative
  - [ ] No absolute paths
  - [ ] Forward slashes throughout
  - [ ] Platform-agnostic
  - Points: 2

- [ ] Scripts and templates discoverable
  - [ ] Referenced in SKILL.md
  - [ ] Clear usage instructions
  - [ ] Easy to find and use
  - Points: 2

### Completeness (6 Points)

- [ ] All referenced files exist
  - [ ] Every link points to existing file
  - [ ] No dangling references
  - Points: 2

- [ ] No TODO or incomplete sections
  - [ ] All sections written
  - [ ] No placeholder text
  - [ ] No "TBD" or "coming soon"
  - Points: 2

- [ ] Documentation complete
  - [ ] Scripts have usage comments
  - [ ] Templates have placeholders documented
  - [ ] All examples explained
  - Points: 2

---

## Release Gate Checks (15 Points)

Performed before merge to production:

### Final Review (5 Points)

- [ ] At least 1 peer review completed
- [ ] No critical issues flagged
- [ ] All feedback addressed
  - Points: 5

### Automated Checks (5 Points)

- [ ] YAML syntax valid
- [ ] Markdown renders without errors
- [ ] All file paths accessible
- [ ] No forbidden patterns detected
  - Points: 5

### Documentation (5 Points)

- [ ] Skill registered in `.moai/memory/skill-registry.md`
- [ ] CHANGELOG entry added
- [ ] Version number set (start with 0.1.0)
  - Points: 5

---

## Quick Validation Script

Use this bash script for automated validation:

```bash
#!/bin/bash
set -euo pipefail

SKILL_PATH=$1
SCORE=0
MAX_SCORE=85

echo "🔍 Validating Skill: $SKILL_PATH"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check SKILL.md exists
if [[ -f "$SKILL_PATH/SKILL.md" ]]; then
  echo "✓ SKILL.md found"
  ((SCORE+=5))
else
  echo "✗ SKILL.md missing"
fi

# Check metadata
if grep -q "^name:" "$SKILL_PATH/SKILL.md"; then
  echo "✓ name field present"
  ((SCORE+=2))
else
  echo "✗ name field missing"
fi

if grep -q "^description:" "$SKILL_PATH/SKILL.md"; then
  echo "✓ description field present"
  ((SCORE+=2))
else
  echo "✗ description field missing"
fi

# Check file count
FILE_COUNT=$(find "$SKILL_PATH" -maxdepth 1 -type f | wc -l)
if [[ $FILE_COUNT -lt 10 ]]; then
  echo "✓ Reasonable file count ($FILE_COUNT files)"
  ((SCORE+=3))
else
  echo "⚠ Many files ($FILE_COUNT), consider consolidation"
fi

# Check line count
SKILL_LINES=$(wc -l < "$SKILL_PATH/SKILL.md" || echo 0)
if [[ $SKILL_LINES -le 500 ]]; then
  echo "✓ SKILL.md within limits ($SKILL_LINES lines)"
  ((SCORE+=3))
else
  echo "✗ SKILL.md too large ($SKILL_LINES lines, max 500)"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Score: $SCORE/$MAX_SCORE"

if [[ $SCORE -ge 75 ]]; then
  echo "✅ Ready for publication"
  exit 0
else
  echo "❌ Needs improvement"
  exit 1
fi
```

**Usage**:
```bash
./validate-skill.sh /path/to/skill-name
```

---

## Audit Report Template

Document your validation results:

```markdown
# Skill Audit Report: [Skill Name]

**Date**: [YYYY-MM-DD]
**Reviewer**: [Name]
**Status**: ✅ PASS / ❌ FAIL

## Metadata Completeness
- [x] name (64 chars max)
- [x] description (includes triggers)
- [x] allowed-tools (minimal)
Score: 8/8

## Content Quality
- [x] SKILL.md ≤ 500 lines
- [x] Examples included
- [x] Terminology consistent
Score: 16/16

## Multi-Model Compatibility
- [x] Haiku: Can activate and use
- [x] Sonnet: Full exploitation
- [x] Opus: Can extend patterns
Score: 8/8

## Security & Compliance
- [x] No sensitive data
- [x] Scripts have error handling
- [x] No risky assumptions
- [x] Destructive ops protected
Score: 8/8

## File Structure
- [x] Proper organization
- [x] All files valid
- [x] Progressive Disclosure applied
- [x] References accurate
Score: 24/24

## Release Gate
- [x] Peer reviewed
- [x] Automated checks pass
- [x] Documentation complete
Score: 15/15

---

## Total Score: 85/85 ✅

### Notes
- Exceeded all quality gates
- Excellent example coverage
- Well-structured documentation

### Recommendations
- None; ready for merge

### Approved By
[Signature/Name]
```

---

## Common Issues & Fixes

| Issue | Points Lost | Fix | Prevention |
|-------|------------|-----|-----------|
| SKILL.md > 500 lines | -5 | Split into reference.md | Plan file structure upfront |
| No examples | -4 | Add 1+ concrete examples | Add during initial draft |
| Time-sensitive data | -8 | Remove dates, use "current" | Security audit before draft |
| Vague descriptions | -5 | Add specific keywords | Test description with Haiku |
| Broken links | -3 | Verify all references | Run automated checker |
| Windows paths | -2 | Replace `\` with `/` | Use linter/CI check |
| No error handling | -3 | Add `set -euo pipefail` | Script template with safety |

---

## Passing Criteria

| Score | Status | Action |
|-------|--------|--------|
| 85-95 | ✅ Pass | Ready for merge |
| 75-84 | ⚠️ Minor issues | Fix and retest |
| 65-74 | ❌ Significant gaps | Major revisions needed |
| <65 | 🔴 Critical issues | Consider redesign |

---

## Checklist Scoring Breakdown

```
Metadata Completeness ............ 8/8 points
Content Quality .................. 16/16 points
Multi-Model Compatibility ........ 8/8 points
Security & Compliance ............ 8/8 points
File Structure ................... 6/6 points
File Validity .................... 6/6 points
Progressive Disclosure ........... 6/6 points
Reference Accuracy ............... 6/6 points
Completeness ..................... 6/6 points
Final Review ..................... 5/5 points
Automated Checks ................. 5/5 points
Documentation .................... 5/5 points
─────────────────────────────────────────
TOTAL ............................ 85/85 points
```

---

## Continuous Quality

### Before Merge
- [ ] All checklist items completed
- [ ] Score ≥ 85/85
- [ ] Peer review approved
- [ ] No blocking issues

### After Merge
- [ ] Monitor activation patterns
- [ ] Gather feedback from users
- [ ] Update if gaps discovered
- [ ] Version bump for significant changes

---

## Related References

- [SKILL.md](SKILL.md) — Main Skill framework
- [METADATA.md](METADATA.md) — Metadata specifications
- [STRUCTURE.md](STRUCTURE.md) — File organization
- [EXAMPLES.md](EXAMPLES.md) — Real-world examples
- [INTERACTIVE-DISCOVERY.md](INTERACTIVE-DISCOVERY.md) — TUI discovery patterns
- [WEB-RESEARCH.md](WEB-RESEARCH.md) — Web research strategies
- [SKILL-UPDATE-ADVISOR.md](SKILL-UPDATE-ADVISOR.md) — Skill analysis & updates

---

**Version**: 0.3.0 (with Interactive Discovery, Web Research, & Update Advisor)
**Last Updated**: 2025-10-22
**Framework**: MoAI-ADK + Claude Code Skills + skill-factory
**Key Enhancements**:
- ✅ Interactive TUI surveys for requirement gathering
- ✅ Web research integration for latest information
- ✅ Skill update analysis & recommendations
- ✅ Official documentation validation
- ✅ Progressive Disclosure pattern
- ✅ Freedom level framework
- ✅ Multi-model compatibility testing

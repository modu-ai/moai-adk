# Agent-Factory Execution Guide: SKILL.md Format Standardization

**Task**: Manually edit 262 SKILL.md files to remove non-standard metadata
**Method**: Parallel file-by-file editing via agent-factory
**Estimated Duration**: 2-3 hours (14 batches × 10-15 minutes each)

---

## Prerequisites

### 1. Backup Creation
```bash
cd /Users/goos/MoAI/MoAI-ADK
tar -czf skill-backup-$(date +%Y%m%d-%H%M%S).tar.gz \
  .claude/skills/*/SKILL.md \
  src/moai_adk/templates/.claude/skills/*/SKILL.md

# Verify backup created
ls -lh skill-backup-*.tar.gz
```

### 2. Create Working Branch
```bash
git checkout -b feature/skill-format-standardization
git status
```

### 3. Review Analysis Documents
- Read: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/ANALYSIS-FORMAT-VIOLATIONS.md`
- Reference: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/FILE-LIST-COMPLETE.md`

---

## Required Format (Official Standard)

### Minimal YAML Frontmatter
```yaml
---
name: skill-identifier
description: Brief description of what this Skill does and when to use it (max 1024 chars)
---
```

**ONLY these 2 fields permitted**. All others must be removed.

---

## Batch Execution Plan

### Overview
- **Total Files**: 262 (131 main + 131 templates)
- **Batch Size**: 20 files per session
- **Total Batches**: 14 sessions
- **Priority Order**: High-impact → Domain → Language → Remaining

### Batch Priority Distribution

**Phase 1: High-Impact Skills (Session 1)**
- Files 1-30 from main directory
- Most frequently used skills
- Highest user impact
- Estimated time: 20-25 minutes

**Phase 2: Domain Skills (Sessions 2-3)**
- moai-domain-* pattern skills
- Consistent structure
- Files 31-70
- Estimated time per session: 12-15 minutes

**Phase 3: Language Skills (Session 4)**
- moai-lang-* pattern skills
- Similar violations
- Files 71-90
- Estimated time: 12-15 minutes

**Phase 4: Remaining Main Skills (Sessions 5-7)**
- All other main directory skills
- Files 91-131
- Estimated time per session: 10-15 minutes

**Phase 5: Template Skills (Sessions 8-14)**
- Mirror of main directory (src/moai_adk/templates/.claude/skills)
- Files 1-131 (template directory)
- Can be faster due to identical structure
- Estimated time per session: 8-12 minutes

---

## Per-File Editing Procedure

### Step-by-Step Instructions

**Step 1: Read Current Content**
```
Open SKILL.md file
Read lines 1-50 (YAML frontmatter + first section)
Identify current format
```

**Step 2: Analyze YAML Frontmatter**
```
Lines 1-20 typically contain YAML frontmatter
Count fields:
  ✅ KEEP: name, description
  ❌ DELETE: version, created, updated, status, keywords, tier, allowed-tools, author, tags, triggers, agents, freedom_level, context7_references
```

**Step 3: Edit YAML Frontmatter**
```
Delete all lines containing:
- version:
- created: / created_date:
- updated: / updated_date:
- status:
- keywords:
- tier:
- allowed-tools: / allowed_tools:
- auto-load: / auto_load:
- author:
- tags:
- triggers:
- agents:
- freedom_level:
- context7_references:

Result should be EXACTLY:
---
name: skill-identifier
description: Description text...
---
```

**Step 4: Scan Body for Metadata Tables**
```
Search for these section headers:
- "## Skill Metadata"
- "## Metadata"
- Markdown tables with | Field | Value |

If found, DELETE entire section including table
(typically 10-20 lines)
```

**Step 5: Verify Compliance**
```
Checklist:
□ YAML frontmatter has exactly 2 lines between --- markers
□ No version, status, tier, keywords fields present
□ Description field is comprehensive (includes "what" and "when")
□ No "## Skill Metadata" section in body
□ Progressive Disclosure structure intact (Level 1/2/3 or Quick/Implementation/Advanced)
□ Content begins immediately after YAML frontmatter
```

**Step 6: Save File**
```
Write changes to file
Close file
```

---

## Session Template (Batch 1-20 Files)

### Session N: Files X-Y

**Preparation**:
```
1. Open agent-factory
2. Load file list for this batch
3. Review expected violations for file type
4. Set parallel editing mode
```

**Execution**:
```
For each file in batch:
  1. Read current YAML frontmatter (lines 1-30)
  2. Count fields (should be 2 after editing)
  3. Remove extra YAML fields (keep only name, description)
  4. Scan body for "## Skill Metadata" section
  5. If found, delete entire metadata table
  6. Verify compliance (checklist above)
  7. Save file
  8. Mark as complete
```

**Verification**:
```bash
# After completing batch, run verification
for file in /Users/goos/MoAI/MoAI-ADK/.claude/skills/[batch-files]/SKILL.md; do
  echo "Checking: $file"
  
  # Count YAML fields
  field_count=$(awk '/^---$/,/^---$/{print}' "$file" | grep -c '^[a-z]')
  if [ "$field_count" -ne 2 ]; then
    echo "  FAIL: Has $field_count fields (expected 2)"
  fi
  
  # Check for metadata table
  if grep -q "## Skill Metadata" "$file"; then
    echo "  FAIL: Still contains metadata table"
  fi
  
  echo "  OK: Compliant"
done
```

**Git Commit**:
```bash
# Commit after each successful batch
git add .claude/skills/[batch-files]/SKILL.md
git commit -m "chore: standardize SKILL.md format for batch N (files X-Y)

- Remove non-standard YAML frontmatter fields
- Remove metadata tables from body
- Ensure only name and description in YAML
- Maintain Progressive Disclosure structure

Batch: N/14
Files: [list of skill names]"
```

---

## Detailed Batch Breakdown

### Batch 1: High-Impact Core Skills (Files 1-30)

**Files** (main directory):
1. moai-foundation-trust
2. moai-foundation-specs
3. moai-foundation-ears
4. moai-lang-python
5. moai-lang-typescript
6. moai-lang-javascript
7. moai-domain-backend
8. moai-domain-frontend
9. moai-domain-security
10. moai-essentials-debug
11. moai-essentials-perf
12. moai-essentials-review
13. moai-essentials-refactor
14. moai-cc-skill-factory
15. moai-cc-configuration
16. moai-context7-lang-integration
17. moai-core-ask-user-questions
18. moai-core-code-reviewer
19. moai-core-agent-factory
20. moai-core-spec-authoring
21. moai-domain-database
22. moai-domain-devops
23. moai-domain-web-api
24. moai-lang-go
25. moai-lang-rust
26. moai-lang-java
27. moai-mcp-builder
28. moai-domain-ml
29. moai-domain-testing
30. moai-playwright-webapp-testing

**Expected Violations**:
- High severity: version, created, updated, status, keywords, tier, allowed-tools
- Metadata tables in body for most files
- Estimated edits per file: 8-15 lines removed

**Estimated Time**: 20-25 minutes

---

### Batch 2-3: Domain Skills (Files 31-70)

**Pattern**: moai-domain-*

**Files**:
- moai-domain-backend
- moai-domain-cli-tool
- moai-domain-cloud
- moai-domain-data-science
- moai-domain-database
- moai-domain-devops
- moai-domain-figma
- moai-domain-frontend
- moai-domain-ml-ops
- moai-domain-ml
- moai-domain-mobile-app
- moai-domain-monitoring
- moai-domain-notion
- moai-domain-security
- moai-domain-testing
- moai-domain-web-api

**Expected Violations**:
- Consistent pattern across domain skills
- Medium severity: version, tier, status
- Some metadata tables

**Estimated Time per Batch**: 12-15 minutes

---

### Batch 4: Language Skills (Files 71-90)

**Pattern**: moai-lang-*

**Files**:
- moai-lang-c
- moai-lang-cpp
- moai-lang-csharp
- moai-lang-dart
- moai-lang-go
- moai-lang-html-css
- moai-lang-java
- moai-lang-javascript
- moai-lang-kotlin
- moai-lang-php
- moai-lang-python
- moai-lang-r
- moai-lang-ruby
- moai-lang-rust
- moai-lang-scala
- moai-lang-shell
- moai-lang-sql
- moai-lang-swift
- moai-lang-tailwind-css
- moai-lang-typescript

**Expected Violations**:
- Similar structure across language skills
- Low-medium severity: version, tier
- Fewer metadata tables

**Estimated Time**: 12-15 minutes

---

### Batches 5-7: Remaining Main Skills (Files 91-131)

**Varied patterns**: moai-cc-*, moai-security-*, moai-baas-*, moai-project-*, etc.

**Expected Violations**:
- Variable severity
- Mixed patterns
- Average 5-8 lines removed per file

**Estimated Time per Batch**: 10-15 minutes

---

### Batches 8-14: Template Skills (Mirror of Main)

**Path**: /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills

**Strategy**:
- These are exact mirrors of main directory
- Should have identical violations
- Can reference main directory edits as template
- Faster due to familiarity

**Estimated Time per Batch**: 8-12 minutes

---

## Quality Assurance Procedures

### Per-Batch Verification Script

```bash
#!/bin/bash
# verify-batch.sh

BATCH_FILES="$1"  # Pass file list as argument

echo "Verifying SKILL.md format compliance..."
echo "=========================================="

violations=0

for file in $BATCH_FILES; do
  echo ""
  echo "File: $(basename $(dirname $file))"
  
  # Check YAML field count
  field_count=$(awk '/^---$/,/^---$/{print}' "$file" | grep -c '^[a-z_]')
  
  if [ "$field_count" -ne 2 ]; then
    echo "  ❌ FAIL: $field_count fields (expected 2)"
    violations=$((violations + 1))
    
    # Show which fields exist
    echo "  Fields found:"
    awk '/^---$/,/^---$/{print}' "$file" | grep '^[a-z_]' | sed 's/:.*//g' | sed 's/^/    - /'
  else
    echo "  ✅ PASS: Correct field count"
  fi
  
  # Check for metadata table
  if grep -q "## Skill Metadata" "$file"; then
    echo "  ❌ FAIL: Contains metadata table"
    violations=$((violations + 1))
  else
    echo "  ✅ PASS: No metadata table"
  fi
  
  # Check description field exists and is non-empty
  desc_length=$(awk '/^description:/{print}' "$file" | sed 's/description: *//' | wc -c)
  if [ "$desc_length" -lt 50 ]; then
    echo "  ⚠️  WARNING: Description too short ($desc_length chars)"
  else
    echo "  ✅ PASS: Description adequate ($desc_length chars)"
  fi
done

echo ""
echo "=========================================="
if [ $violations -eq 0 ]; then
  echo "✅ ALL FILES COMPLIANT"
  exit 0
else
  echo "❌ $violations VIOLATIONS FOUND"
  exit 1
fi
```

**Usage**:
```bash
# After completing batch
./verify-batch.sh "/Users/goos/MoAI/MoAI-ADK/.claude/skills/{skill1,skill2,skill3}/SKILL.md"
```

---

### Final Comprehensive Verification

```bash
#!/bin/bash
# verify-all-skills.sh

echo "Final SKILL.md Format Compliance Check"
echo "======================================"
echo ""

total_files=0
violations=0

# Check main directory
echo "Checking main skills directory..."
for file in /Users/goos/MoAI/MoAI-ADK/.claude/skills/*/SKILL.md; do
  total_files=$((total_files + 1))
  
  field_count=$(awk '/^---$/,/^---$/{print}' "$file" | grep -c '^[a-z_]')
  
  if [ "$field_count" -ne 2 ]; then
    echo "VIOLATION: $file ($field_count fields)"
    violations=$((violations + 1))
  fi
  
  if grep -q "## Skill Metadata" "$file"; then
    echo "VIOLATION: $file (contains metadata table)"
    violations=$((violations + 1))
  fi
done

# Check template directory
echo "Checking template skills directory..."
for file in /Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills/*/SKILL.md; do
  total_files=$((total_files + 1))
  
  field_count=$(awk '/^---$/,/^---$/{print}' "$file" | grep -c '^[a-z_]')
  
  if [ "$field_count" -ne 2 ]; then
    echo "VIOLATION: $file ($field_count fields)"
    violations=$((violations + 1))
  fi
  
  if grep -q "## Skill Metadata" "$file"; then
    echo "VIOLATION: $file (contains metadata table)"
    violations=$((violations + 1))
  fi
done

echo ""
echo "======================================"
echo "Total files checked: $total_files"
echo "Violations found: $violations"

if [ $violations -eq 0 ]; then
  echo "✅ ALL 262 SKILLS COMPLIANT"
  exit 0
else
  echo "❌ COMPLIANCE FAILED"
  exit 1
fi
```

**Usage**:
```bash
# After all batches complete
chmod +x verify-all-skills.sh
./verify-all-skills.sh
```

---

## Git Workflow

### Commit Strategy

**Per-Batch Commits**:
```bash
# After each batch (20 files)
git add .claude/skills/[batch-files]/SKILL.md
git commit -m "chore: standardize SKILL.md format batch N/14

- Remove non-standard YAML frontmatter
- Remove metadata tables
- Files: [list of 20 skill names]"
```

**Final Commit**:
```bash
# After all 14 batches
git add src/moai_adk/templates/.claude/skills/*/SKILL.md
git commit -m "chore: synchronize template skills with standardized format

- Apply same format standardization to template directory
- Ensures consistency between main and template skills
- Total: 262 SKILL.md files standardized"
```

**Create Pull Request**:
```bash
git push origin feature/skill-format-standardization

# Create PR via GitHub CLI or web interface
gh pr create \
  --title "Standardize SKILL.md format across all 262 skills" \
  --body "## Summary
- Remove non-standard YAML frontmatter fields
- Remove metadata tables from body content
- Ensure ONLY name and description in YAML
- Maintain Progressive Disclosure structure

## Changes
- 262 SKILL.md files updated (131 main + 131 templates)
- 14 batches executed with verification
- All files verified compliant with minimal format

## Testing
- Ran per-batch verification after each 20 files
- Final comprehensive verification passed
- No breaking changes to skill content

Closes #[issue-number]"
```

---

## Risk Mitigation Checklist

### Before Starting
- [x] Backup created (skill-backup-YYYYMMDD-HHMMSS.tar.gz)
- [x] Working branch created (feature/skill-format-standardization)
- [x] Analysis documents reviewed
- [x] Verification scripts prepared

### During Execution
- [ ] Complete batches in order (1-14)
- [ ] Run verification after each batch
- [ ] Commit after successful verification
- [ ] Fix any failures before proceeding
- [ ] Track progress (batch N/14 completed)

### After Completion
- [ ] Run final comprehensive verification
- [ ] All 262 files pass compliance check
- [ ] Create pull request
- [ ] Request code review
- [ ] Merge after approval

---

## Troubleshooting

### Issue: Field count ≠ 2
**Cause**: Extra YAML field not removed
**Fix**: Re-open file, identify extra field, delete line

### Issue: Metadata table still present
**Cause**: Section not fully deleted
**Fix**: Search for "## Skill Metadata", delete entire section including table

### Issue: Description too short
**Cause**: Description field truncated during edit
**Fix**: Review original file backup, restore full description

### Issue: Progressive Disclosure structure broken
**Cause**: Accidentally deleted content headings
**Fix**: Restore from backup, re-edit carefully

---

## Success Criteria

### Completion Checklist
- [ ] All 262 SKILL.md files edited
- [ ] All files have exactly 2 YAML fields (name, description)
- [ ] No metadata tables in any file body
- [ ] Progressive Disclosure structure intact for all files
- [ ] All per-batch verifications passed
- [ ] Final comprehensive verification passed
- [ ] All changes committed to git
- [ ] Pull request created and approved
- [ ] Changes merged to main branch

### Expected Outcomes
1. ✅ Consistent format across all 262 skills
2. ✅ Reduced file sizes (avg 10-20 lines per file)
3. ✅ Cleaner git diffs (no metadata noise)
4. ✅ Simplified maintenance (no version drift)
5. ✅ Compliance with Claude Code minimal format standards

---

## Appendix: Quick Reference

### Minimal Format Template
```yaml
---
name: skill-identifier
description: Brief description of what this Skill does and when to use it
---

# Skill Title

## Level 1: Quick Reference
Content...

## Level 2: Implementation
Content...

## Level 3: Advanced
Content...
```

### Fields to DELETE
```
version
created / created_date
updated / updated_date
status
keywords
tier
allowed-tools / allowed_tools
auto-load / auto_load
author
tags
triggers
agents
freedom_level
context7_references
```

### Sections to DELETE from Body
```
## Skill Metadata
| Field | Value |
...
```

---

**End of Execution Guide**

**Next Step**: Begin Batch 1 execution with agent-factory

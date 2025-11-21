# SPEC-SKILL-STANDARDS-001: SKILL.md Format Standardization

**Created**: 2025-11-21
**Status**: Ready for Execution
**Agent**: skill-factory
**Task**: Manual standardization of all 262 SKILL.md files

---

## Executive Summary

This specification documents the standardization effort to ensure all SKILL.md files comply with the minimal 2-field YAML frontmatter format required by Claude Code.

**Problem**: Current SKILL.md files contain excessive metadata (version, status, tier, keywords, etc.) that violates official format standards.

**Solution**: Manually edit all 262 SKILL.md files to remove non-standard fields while preserving content structure.

**Impact**: 
- ✅ Cleaner, consistent format across all skills
- ✅ Reduced file sizes (avg 10-20 lines per file)
- ✅ Easier maintenance (no version drift)
- ✅ Compliance with Claude Code standards

---

## Required Format

### Official Standard (Claude Code)
```yaml
---
name: skill-identifier
description: Brief description of what this Skill does and when to use it (max 1024 chars)
---
```

**ONLY these 2 fields are permitted in YAML frontmatter.**

---

## Project Structure

```
.moai/specs/SPEC-SKILL-STANDARDS-001/
├── README.md                          # This file (overview)
├── ANALYSIS-FORMAT-VIOLATIONS.md     # Detailed violation analysis
├── FILE-LIST-COMPLETE.md             # All 262 SKILL.md file paths
├── AGENT-FACTORY-EXECUTION-GUIDE.md  # Step-by-step execution instructions
└── verification/
    ├── verify-batch.sh               # Per-batch verification script
    └── verify-all-skills.sh          # Final comprehensive check
```

---

## Quick Start

### 1. Review Analysis
Read `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/ANALYSIS-FORMAT-VIOLATIONS.md` for:
- Current violations breakdown
- Format standardization rules
- Examples of compliant vs non-compliant files

### 2. Review Execution Plan
Read `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/AGENT-FACTORY-EXECUTION-GUIDE.md` for:
- 14-batch execution strategy
- Per-file editing procedure
- Quality assurance processes
- Git workflow

### 3. Execute with agent-factory
**Total Time**: 2-3 hours (14 batches × 10-15 min each)

```bash
# Phase 1: Backup
cd /Users/goos/MoAI/MoAI-ADK
tar -czf skill-backup-$(date +%Y%m%d-%H%M%S).tar.gz \
  .claude/skills/*/SKILL.md \
  src/moai_adk/templates/.claude/skills/*/SKILL.md

# Phase 2: Create branch
git checkout -b feature/skill-format-standardization

# Phase 3: Execute 14 batches with agent-factory
# (See AGENT-FACTORY-EXECUTION-GUIDE.md for details)

# Phase 4: Verify
./verify-all-skills.sh

# Phase 5: Commit and PR
git push origin feature/skill-format-standardization
```

---

## Deliverables

### Analysis Documents
- [x] Format violation analysis (ANALYSIS-FORMAT-VIOLATIONS.md)
- [x] Complete file list with 262 paths (FILE-LIST-COMPLETE.md)
- [x] Detailed editing checklist per file
- [x] Batch execution strategy (14 sessions)

### Execution Materials
- [x] Step-by-step agent-factory guide (AGENT-FACTORY-EXECUTION-GUIDE.md)
- [x] Per-file editing procedure
- [x] Verification scripts (verify-batch.sh, verify-all-skills.sh)
- [x] Git workflow instructions
- [x] Risk mitigation checklist

### Quality Assurance
- [x] Per-batch verification process
- [x] Final comprehensive verification
- [x] Troubleshooting guide
- [x] Success criteria checklist

---

## Key Constraints

### Mandatory Requirements
1. **NO automation** - All edits must be manual (verified file-by-file)
2. **NO scripts** - No sed, awk, or Python automation
3. **Manual verification** - Human review of each file's compliance
4. **agent-factory execution** - Use agent-factory for parallel file handling

### Reasoning
- Manual editing ensures description quality
- Catches edge cases automation would miss
- Validates Progressive Disclosure structure intact
- Ensures "what" and "when" in description field

---

## Violation Categories

### Category A: YAML Frontmatter (DELETE these fields)
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

### Category B: Body Content (DELETE these sections)
```markdown
## Skill Metadata
| Field | Value |
| ----- | ----- |
| **Skill Name** | ... |
| **Version** | ... |
```

---

## Execution Strategy

### Phase Distribution
- **Phase 1**: High-impact skills (30 files, Session 1)
- **Phase 2**: Domain skills (40 files, Sessions 2-3)
- **Phase 3**: Language skills (20 files, Session 4)
- **Phase 4**: Remaining main skills (41 files, Sessions 5-7)
- **Phase 5**: Template skills (131 files, Sessions 8-14)

### Time Estimates
- Batch 1 (high-impact): 20-25 minutes
- Batches 2-7 (main directory): 10-15 minutes each
- Batches 8-14 (templates): 8-12 minutes each
- **Total**: 2-3 hours

---

## Success Metrics

### Immediate Benefits
- All 262 SKILL.md files compliant with minimal format
- Reduced file size (avg 10-20 lines removed per file)
- Consistent format across entire skill library
- Easier maintenance (no version drift, status conflicts)
- Cleaner git diffs (no metadata noise)

### Long-Term Benefits
- Simplified skill creation process (less metadata to manage)
- Faster skill loading (less parsing overhead)
- Better compliance with Claude Code standards
- Reduced token consumption (less metadata to process)
- Clearer focus on skill content vs metadata

---

## Risk Mitigation

### Before Starting
- ✅ Backup created (skill-backup-YYYYMMDD-HHMMSS.tar.gz)
- ✅ Working branch created (feature/skill-format-standardization)
- ✅ Analysis documents reviewed
- ✅ Verification scripts prepared

### During Execution
- Edit in small batches (20 files max per session)
- Verify each batch before proceeding
- Commit after each successful batch
- Track progress (batch N/14 completed)

### After Completion
- Run final comprehensive verification (verify-all-skills.sh)
- All 262 files pass compliance check
- Create pull request with detailed summary
- Request code review before merge

---

## File Locations

### Analysis Documents
- **Main Analysis**: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/ANALYSIS-FORMAT-VIOLATIONS.md`
- **File List**: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/FILE-LIST-COMPLETE.md`
- **Execution Guide**: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/AGENT-FACTORY-EXECUTION-GUIDE.md`

### Target Files
- **Main Skills**: `/Users/goos/MoAI/MoAI-ADK/.claude/skills/*/SKILL.md` (131 files)
- **Template Skills**: `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/templates/.claude/skills/*/SKILL.md` (131 files)

### Verification Scripts
- **Per-Batch**: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/verification/verify-batch.sh`
- **Final Check**: `/Users/goos/MoAI/MoAI-ADK/.moai/specs/SPEC-SKILL-STANDARDS-001/verification/verify-all-skills.sh`

---

## References

### Official Standards
- Claude Code Skill Format: Minimal 2-field YAML frontmatter
- Progressive Disclosure: Level 1 (Quick) → Level 2 (Implementation) → Level 3 (Advanced)
- Description Requirements: Must include "what" and "when to use"

### Related Documentation
- `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-cc-skill-factory/SKILL.md` - Skill creation patterns
- `/Users/goos/MoAI/MoAI-ADK/.claude/skills/moai-cc-skills/SKILL.md` - Skill management
- `/Users/goos/MoAI/MoAI-ADK/CLAUDE.md` - MoAI-ADK execution guide

---

## Next Steps

1. **Review** all analysis documents (ANALYSIS, FILE-LIST, EXECUTION-GUIDE)
2. **Prepare** verification scripts and working environment
3. **Execute** 14 batches with agent-factory following EXECUTION-GUIDE
4. **Verify** each batch and final comprehensive check
5. **Commit** changes with detailed PR summary
6. **Merge** after code review approval

---

## Contact & Support

**Specification Owner**: skill-factory agent
**Created**: 2025-11-21
**Version**: 1.0.0
**Status**: Ready for Execution

**Questions**: Refer to AGENT-FACTORY-EXECUTION-GUIDE.md Troubleshooting section

---

**End of SPEC-SKILL-STANDARDS-001 README**

# SPEC-UPDATE-REFACTOR-002 Codebase Exploration Index

This document serves as an index to the exploration findings. Created: 2025-10-28

## Generated Documents

### 1. exploration-update-feature.md
Comprehensive technical analysis of the existing codebase related to the update command.

**Read this document for:** Understanding the existing codebase structure, implementation details with line numbers, and established patterns.

---

### 2. implementation-UPDATE-REFACTOR-002.md
Actionable guidance for implementing SPEC-UPDATE-REFACTOR-002.

**Read this document for:** Step-by-step implementation guidance, code patterns to follow, and testing strategies.

---

## Quick Reference

### Current Update Command Location
- **File:** `/Users/goos/MoAI/MoAI-ADK/src/moai_adk/cli/commands/update.py`
- **Main Function:** `update()` at lines 260-393

### Key Modules
- **CLI Entry:** `src/moai_adk/__main__.py`
- **Template Processing:** `src/moai_adk/core/template/processor.py`
- **Version Management:** `src/moai_adk/__init__.py`
- **Tests:** `tests/unit/test_update.py`

### Current Workflow (4 Phases)
1. **Phase 1: Version Check** - Fetch from PyPI API
2. **Phase 2: Backup Creation** - Create .moai-backups/
3. **Phase 3: Template Update** - Copy and merge templates
4. **Phase 4: Finalization** - Update config, set optimized=false

---

## Related SPEC Documents

- `.moai/specs/SPEC-UPDATE-REFACTOR-002/spec.md` - Official specification
- `.moai/specs/SPEC-UPDATE-REFACTOR-003/spec.md` - Option B specification
- `.moai/specs/UPDATE-STRATEGY-SUMMARY.md` - Executive summary
- `.moai/specs/GITHUB-ISSUE-85-SUMMARY.md` - Issue analysis

---

**All documents located in:** `.moai/docs/`
**Last Updated:** 2025-10-28
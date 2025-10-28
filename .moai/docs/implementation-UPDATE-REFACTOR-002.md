# SPEC-UPDATE-REFACTOR-002 Implementation Guide

## Overview

This guide details the UX improvement strategy for moai-adk's self-update feature, addressing GitHub issue #85. The current 2-stage workflow requires explicit user action between stages, causing confusion. Three improvement options are proposed with implementation roadmaps.

**Status**: SPEC v0.0.2 approved | Guide v1.0 updated 2025-10-28

---

## Problem Statement

Current workflow requires users to execute **2-3 commands manually**:
```bash
moai-adk update              # Stage 1: Package upgrade
uv tool upgrade moai-adk      # (or pip install --upgrade)
moai-adk update              # Stage 2: Template sync
moai-adk update-templates    # (confusion about next step)
```

**Root Cause**: Python process cannot upgrade itself while running (CST-001). This is a **technical requirement**, not a design flaw.

**User Impact**: Users see conflicting messages and don't understand the workflow.

---

## Improvement Strategy

### Option A: Message Clarity ✅ RECOMMENDED (v0.6.2)
**Effort**: 30 minutes | **Complexity**: Low | **Risk**: Low

Clear, explicit 2-step messaging instead of confusing "run again" messages.

**Changes**:
1. **File**: `src/moai_adk/cli/commands/update.py`
   - Update Stage 1 completion (line 702-703)
   - Explicit next command: `moai-adk update --templates-only`

2. **Message Before**:
   ```
   ✓ Upgrade complete!
   📢 Run 'moai-adk update' again to sync templates
   ```

3. **Message After**:
   ```
   ✓ Stage 1/2 complete: Package upgraded!
   📢 Next step - sync templates with:

     moai-adk update --templates-only

   Or complete all steps in one command (preview):
     moai-adk update-complete  # Coming in v0.7.0
   ```

4. **Benefits**:
   - Explicit 2-step guidance
   - Promotes `--templates-only` flag (already implemented!)
   - Hints at future unified command
   - No code changes needed (message only)

5. **Documentation**: Update README.md with workflow diagram

---

## Implementation File Structure

This document is stored in `.moai/docs/` as an internal guide for Alfred and development team members. See CLAUDE.md > Document Management Rules for placement policies.

**Related files**:
- `exploration-update-feature.md` - Codebase technical analysis
- `codebase-exploration-index.md` - Index and quick reference
- `.moai/specs/SPEC-UPDATE-REFACTOR-002/spec.md` - Official specification
- `.moai/specs/SPEC-UPDATE-REFACTOR-003/spec.md` - Option B specification
# GitHub Issue #85: Self-Update UX Improvement Strategy

**Status**: ✅ ANALYZED & PLANNED
**Date**: 2025-10-28
**Related SPEC**: SPEC-UPDATE-REFACTOR-002 (v0.0.2)

---

## Issue Summary

Users experience confusion with the current `moai-adk update` 2-stage workflow because:

1. **Multiple commands required**: Stage 1 (upgrade) then Stage 2 (sync templates)
2. **Unclear messaging**: "Run again" message is confusing
3. **Technical hidden**: Python process limitation not explained
4. **Different from competitors**: Claude Code has single `claude update` command

---

## Analysis: Why 2-Stage Workflow Exists

### Technical Constraint (CST-001)
```
Running Python process CANNOT upgrade itself while executing
↓
Solution: Run upgrade → exit → let installer restart → run again
```

**This is NOT a design flaw, it's a fundamental requirement.**

Attempts to fix:
- ❌ **Cannot use subprocess**: Process would upgrade package but keep old code loaded
- ❌ **Cannot auto-restart**: Would require platform-specific daemon code (fragile)
- ✅ **Can improve messaging**: Make 2-stage clear and simple
- ✅ **Can add wrapper command**: New command to handle both stages

---

## Proposed Solution: 3-Option Strategy

### Option A: Message Clarity ✅ RECOMMENDED (v0.6.2)
**Effort**: 30 mins | **Impact**: High UX improvement

**What Changes**:
```
BEFORE:
✓ Upgrade complete!
📢 Run 'moai-adk update' again to sync templates

AFTER:
✓ Stage 1/2 complete: Package upgraded!
📢 Next step - sync templates with:

  moai-adk update --templates-only

💡 Tip: In v0.7.0+, use 'moai-adk update-complete' for one-command updates
```

**Impact**:
- ✅ Immediately solves message confusion
- ✅ Promotes already-implemented `--templates-only` flag
- ✅ No code behavior changes (zero risk)
- ✅ Can release quickly as v0.6.2

---

### Option B: Unified Command (v0.7.0)
**Effort**: 2-3 hours | **Impact**: Best UX improvement

**New Command**: `moai-adk update-complete`

```bash
$ moai-adk update-complete
→ Stage 1: Upgrade package (if needed)
→ Stage 2: Sync templates
✓ Done! Ready for /alfred:0-project update
```

**Benefits**:
- ✅ Single command (like `claude update`)
- ✅ Auto-detects what stage to run
- ✅ Backward compatible (existing `update` unchanged)
- ✅ More test coverage, better quality

**Implementation**: SPEC-UPDATE-REFACTOR-003 (draft created)

---

### Option C: Integrated Flag (v0.8.0)
**Effort**: 3-4 hours | **Impact**: Most comprehensive

**New Flag**: `moai-adk update --integrated`

```bash
$ moai-adk update --integrated
→ Auto-detect: upgrade needed?
→ If yes: Stage 1 + Stage 2 (auto)
→ If no: Stage 2 only
✓ Complete update with single flag
```

**Benefits**:
- ✅ Maximum UX parity with Claude Code
- ✅ Fully backward compatible (opt-in flag)
- ✅ Most flexible solution
- ✅ Highest user satisfaction potential

---

## Recommended Roadmap

```
Today (v0.6.2)
  ↓
  Implement Option A: Message clarity
  Files: update.py + README + CHANGELOG
  Effort: 30 mins
  Release: Immediate

Weeks (v0.7.0)
  ↓
  Based on user feedback from Option A
  Implement Option B: update-complete command
  SPEC: SPEC-UPDATE-REFACTOR-003 (draft ready)
  Effort: 3-4 hours
  Release: v0.7.0

Months (v0.8.0)
  ↓
  Implement Option C: --integrated flag
  SPEC: SPEC-UPDATE-REFACTOR-004 (to be created)
  Effort: 3-4 hours
  Release: v0.8.0
```

---

## Implementation Status

### ✅ Completed
- SPEC-UPDATE-REFACTOR-002 v0.0.2 (updated with UX strategy)
- IMPLEMENTATION_GUIDE.md (updated with 3 options)
- SPEC-UPDATE-REFACTOR-003 (draft created for Option B)
- Analysis document (this file)

### ⏳ Pending - Option A Implementation (v0.6.2)
- Update `src/moai_adk/cli/commands/update.py` message
- Update `README.md` with workflow documentation
- Update `README.ko.md` (Korean)
- Update `CHANGELOG.md`
- Create PR and release

### 📋 Backlog - Option B (v0.7.0)
- Review SPEC-UPDATE-REFACTOR-003
- Implement `update_complete.py` command
- Add tests
- Documentation
- Release

### 🔮 Backlog - Option C (v0.8.0)
- Create SPEC-UPDATE-REFACTOR-004
- Implement `--integrated` flag
- Extended testing
- Release

---

## User Communication Strategy

### Immediate (v0.6.2 Release Notes)
```markdown
### Improved: Update Command Messaging

The `moai-adk update` 2-stage workflow is now clearer:

**Stage 1**: Package upgrade
```bash
moai-adk update
→ Detects and upgrades package
→ Shows explicit next command
```

**Stage 2**: Template synchronization
```bash
moai-adk update --templates-only
→ Syncs templates to match version
→ Sets up for optimization
```

The 2-stage design is required because Python processes cannot upgrade
themselves while running. This is the same constraint that affects pip,
uv, and other package managers.

Coming in v0.7.0: Single-command `moai-adk update-complete` for full automation!
```

### Future (v0.7.0 if implementing Option B)
```markdown
### New: Unified Update Command

`moai-adk update-complete` handles full update in one command:
- Auto-detects what stage to run
- Executes all stages sequentially
- Clear progress messaging
- Same reliability as 2-stage workflow
```

---

## Technical Details

### Why Each Option Works

**Option A (Message Clarity)**
- ✅ No code changes needed (message only)
- ✅ Existing tests still pass
- ✅ Quick deployment
- ✅ Solves primary UX pain (confusion)

**Option B (New Command)**
- ✅ Reuses existing logic (no duplication)
- ✅ New command, backward compatible
- ✅ Better for users who want simplicity
- ✅ Clear code separation

**Option C (Integrated Flag)**
- ✅ Reuses existing command
- ✅ Opt-in via flag
- ✅ No changes to default behavior
- ✅ Most similar to competitive tools

---

## Key Files

### Already Updated (Ready for Review)
- `SPEC-UPDATE-REFACTOR-002` - Approval with UX strategy (v0.0.2)
- `IMPLEMENTATION_GUIDE.md` - Complete guide for all 3 options
- `SPEC-UPDATE-REFACTOR-003` - Draft for Option B (v0.7.0)

### Pending Changes
- `src/moai_adk/cli/commands/update.py` - Line 702-703 message
- `README.md` - Add workflow documentation section
- `README.ko.md` - Korean translation
- `CHANGELOG.md` - v0.6.2 entry

---

## Success Metrics

### Option A Success
- ✅ Users clearly understand 2-step workflow
- ✅ GitHub #85 feedback indicates improved clarity
- ✅ `--templates-only` flag usage increases
- ✅ All existing tests pass

### Option B Success
- ✅ Single `update-complete` command adopted by users
- ✅ User satisfaction increases
- ✅ Feedback indicates "feels like one command"
- ✅ Usage metrics show adoption

### Option C Success
- ✅ `--integrated` flag becomes default recommendation
- ✅ UX parity achieved with `claude update`
- ✅ Minimal user confusion
- ✅ Cross-platform testing confirms reliability

---

## Questions for Discussion

1. **Should we implement all 3 options over time?**
   - Recommendation: YES - each provides incremental UX improvement

2. **Timeline concerns?**
   - Option A is 30 mins, can be released immediately
   - Option B adds ~3-4 hours of work, planned for v0.7.0
   - Option C is backlog for future

3. **User feedback collection?**
   - After Option A: Ask if messaging is clear
   - After Option B: Ask if UX improved
   - After Option C: Ask if it's satisfying

4. **Breaking changes?**
   - None planned - all changes backward compatible
   - Existing `moai-adk update` behavior unchanged

---

## References

- **GitHub Issue**: #85
- **SPEC-UPDATE-REFACTOR-002**: v0.0.2 (approved with strategy)
- **SPEC-UPDATE-REFACTOR-003**: v0.0.1 (draft for Option B)
- **IMPLEMENTATION_GUIDE.md**: Complete roadmap
- **CST-001**: Python process limitation (technical constraint)

---

**Document Created**: 2025-10-28
**Status**: Ready for Implementation Planning
**Next Action**: Approve Option A and begin v0.6.2 implementation

---
id: UPDATE-REFACTOR-003
version: 0.0.1
status: draft
created: 2025-10-28
updated: 2025-10-28
author: @Goos
priority: medium
category: CLI Enhancement
labels: [self-update, ux-improvement, unified-command]
depends_on: ["UPDATE-REFACTOR-002"]
blocks: []
related_specs: [SPEC-UPDATE-REFACTOR-002]
related_issue: ["#85"]
scope: unified-update-command
---

# @SPEC:UPDATE-REFACTOR-003: moai-adk Unified Update Command (Option B)

## HISTORY

### v0.0.1 (2025-10-28)
- **INITIAL**: Draft specification for unified `update-complete` command
- **CONTEXT**: Follows Option B from SPEC-UPDATE-REFACTOR-002 UX improvement strategy
- **PURPOSE**: Address GitHub #85 by providing single-command update experience
- **ROADMAP**: Planned for v0.7.0 release

---

## Environment

**Target Platforms**: macOS, Linux, Windows
**Python Version**: 3.13+
**Package Managers**: uv (tool), pipx, pip
**CLI Framework**: Click 8.1+
**Terminal UI**: Rich 13.0+

**Implementation Location**:
- New Command: `src/moai_adk/cli/commands/update_complete.py`
- Integration: `src/moai_adk/__main__.py`
- Tests: `tests/unit/test_update_complete.py`

---

## Problem Statement

Current workflow requires multiple commands:
```bash
moai-adk update              # Stage 1: Upgrade package
moai-adk update              # Stage 2: Sync templates
/alfred:0-project update     # Stage 3: Optimize
```

**User Pain Points**:
1. Multiple commands required
2. No clear indication of progress between stages
3. Confusing messages about "next steps"
4. Different from Claude Code's `claude update` (single command)

**Technical Constraint** (CST-001):
- Python process cannot upgrade itself while running
- Requires process restart between stages
- This is a fundamental limitation, not a design flaw

---

## Solution: Unified `update-complete` Command

### Overview

New command `moai-adk update-complete` that:
1. Detects current state (need upgrade? need template sync?)
2. Executes Stage 1 (package upgrade) if needed
3. Detects package upgrade completion
4. Executes Stage 2 (template sync)
5. Suggests optional Stage 3 (/alfred:0-project update)

**User Experience**:
```bash
$ moai-adk update-complete
🔍 Checking versions...
   Current: 0.6.1, Latest: 0.6.2

📦 Stage 1/2: Package Upgrade
   Upgrading moai-adk 0.6.1 → 0.6.2
   Running: uv tool upgrade moai-adk
   ✓ Package upgraded

✓ Process restart required for changes to take effect
  Run the following to complete setup:

  [cyan]moai-adk update-complete[/cyan]

$ moai-adk update-complete
✓ Package already updated

📄 Stage 2/2: Template Sync
   Creating backup...
   Syncing templates...
   ✓ Templates synced

✓ Update complete!
ℹ️  Next: Optimize your project configuration with:

  [cyan]/alfred:0-project update[/cyan]
```

---

## Requirements

### Core Requirements (Must Have)

1. **UCR-001**: Command must work when upgrade is needed
   - Detect package upgrade is available
   - Execute upgrade via detected installer
   - Prompt user to re-run for next stage

2. **UCR-002**: Command must work after upgrade completes
   - Detect that package is up-to-date
   - Execute template sync
   - Show completion message

3. **UCR-003**: Command must be backward compatible
   - Must not modify behavior of existing `moai-adk update` command
   - Must be opt-in (new command, not replacing existing)
   - Existing scripts continue working unchanged

4. **UCR-004**: Command must handle errors gracefully
   - Upgrade failures
   - Template sync failures
   - Network errors
   - Show helpful recovery steps

5. **UCR-005**: Command must support all relevant flags
   - `--path` - Custom project path
   - `--force` - Skip backup
   - `--check` - Preview mode (show what would happen)
   - `--yes` - Auto-confirm prompts (CI/CD mode)

### User Experience Requirements (Nice to Have)

1. **UXR-001**: Clear progress indication
   - Show which stage is running
   - Show percentage/progress if possible
   - Show timing information

2. **UXR-002**: Informative messaging
   - Clear "next step" guidance
   - Explain why process restart is needed
   - Suggest `/alfred:0-project update` when appropriate

3. **UXR-003**: Consistent with moai-adk style
   - Use Rich console for formatting
   - Match existing color scheme (cyan/green/yellow/red)
   - Include emoji indicators

---

## Design

### Command Signature

```python
@click.command()
@click.option(
    "--path",
    type=click.Path(exists=True),
    default=".",
    help="Project path (default: current directory)"
)
@click.option(
    "--force",
    is_flag=True,
    help="Skip backup and force the update"
)
@click.option(
    "--check",
    is_flag=True,
    help="Preview mode (show what would happen, no changes)"
)
@click.option(
    "--yes",
    is_flag=True,
    help="Auto-confirm all prompts (CI/CD mode)"
)
def update_complete(path: str, force: bool, check: bool, yes: bool) -> None:
    """Complete update workflow (package + templates)."""
```

### Implementation Strategy

**Phase 1: Detection** (Auto-detect what to do)
```python
def _detect_update_state(project_path: Path) -> str:
    """Determine if we need Stage 1 or Stage 2.

    Returns:
        "need_package_upgrade" - Need to upgrade package
        "need_template_sync" - Package is current, need template sync
        "all_current" - Everything is up to date
    """
```

**Phase 2: Stage 1 Execution** (If upgrade needed)
```python
# Reuse existing logic from update.py
from moai_adk.cli.commands.update import _detect_tool_installer, _execute_upgrade

# 1. Detect installer
# 2. Execute upgrade
# 3. Prompt to re-run command
```

**Phase 3: Stage 2 Execution** (If templates need sync)
```python
# Reuse existing logic from update.py
from moai_adk.cli.commands.update import _sync_templates

# 1. Create backup
# 2. Sync templates
# 3. Set optimized=false
```

**Phase 4: Final Suggestion**
```python
# Suggest /alfred:0-project update for optimization
console.print("[cyan]📌 Next: Optimize your project:\n[/cyan]")
console.print("  /alfred:0-project update\n")
```

---

## Testing Strategy

### Unit Tests (15+ tests required)

1. **State Detection Tests**
   - Test detection: package upgrade needed
   - Test detection: template sync needed
   - Test detection: both up to date

2. **Integration Tests**
   - Test Stage 1 → Stage 2 workflow
   - Test with various flag combinations
   - Test error handling at each stage

3. **Option Tests**
   - `--path` with custom directory
   - `--force` skips backup
   - `--check` preview mode
   - `--yes` auto-confirms

4. **Backward Compatibility Tests**
   - Existing `moai-adk update` still works
   - All existing tests continue to pass

### Coverage Target
- 85%+ code coverage (minimum)
- All happy paths covered
- All error paths tested
- Integration tests for full workflow

---

## Success Criteria

✅ **Functionality**:
- Single command executes full update workflow
- Correctly detects which stage to run
- Auto-detects package update completion
- Handles all error cases gracefully

✅ **User Experience**:
- Clear messaging at each stage
- Single "next step" guidance
- Helpful error messages
- Consistent with existing moai-adk UX

✅ **Quality**:
- 85%+ test coverage
- All ruff and mypy checks pass
- Backward compatible
- Works on macOS, Linux, Windows

✅ **Documentation**:
- Help text explains the command
- README includes usage examples
- CHANGELOG documents new feature
- Docstrings are clear

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|-----------|
| Package upgrade fails | High | Low | Show helpful error, suggest manual steps |
| Process doesn't restart properly | High | Low | Explicit user prompt to re-run |
| Template sync fails | Medium | Low | Rollback to backup, clear recovery steps |
| State detection fails | Medium | Low | Fallback to checking files directly |
| Platform compatibility | Medium | Low | Test on macOS, Linux, Windows |

---

## Implementation Plan

### Phase 1: Create SPEC (Complete)
✅ This document

### Phase 2: Implement Core Logic (est. 1.5 hours)
- Create `src/moai_adk/cli/commands/update_complete.py`
- Implement state detection logic
- Implement stage 1 & 2 orchestration
- Add error handling

### Phase 3: CLI Integration (est. 0.5 hours)
- Register command in `src/moai_adk/__main__.py`
- Add to help documentation
- Test command discovery

### Phase 4: Testing (est. 1 hour)
- Write unit tests for state detection
- Write integration tests for workflows
- Test all flag combinations
- Verify backward compatibility

### Phase 5: Documentation (est. 0.5 hours)
- Update README.md with usage examples
- Update README.ko.md (Korean)
- Update CHANGELOG.md
- Write comprehensive docstrings

---

## Timeline

- **v0.7.0**: Initial implementation
  - Design phase complete
  - Core implementation complete
  - Testing complete (85%+)
  - Documentation complete
  - Beta testing with users

- **v0.7.1**: Bug fixes if needed

- **v0.8.0**: Option C (`--integrated` flag) planning begins

---

## Traceability (@TAG)

**SPEC Chain**:
- @SPEC:UPDATE-REFACTOR-003 (this document)
- @SPEC:UPDATE-REFACTOR-002 (parent specification)

**Implementation Chain** (to be populated):
- @TEST: `tests/unit/test_update_complete.py`
- @CODE: `src/moai_adk/cli/commands/update_complete.py`
- @DOC: `README.md`, `CHANGELOG.md`

---

## Dependencies

### Required (Already Available)
- Click 8.1+ - CLI framework
- Rich 13.0+ - Console output
- Packaging - Version parsing
- urllib - PyPI API

### Related Files
- `src/moai_adk/cli/commands/update.py` - Existing update command
- `src/moai_adk/__main__.py` - CLI entry point
- Tests and documentation (same structure as other commands)

---

## Notes for Implementer

1. **Reuse Existing Logic**: Don't duplicate code from `update.py`. Instead:
   - Import helper functions (`_detect_tool_installer`, `_sync_templates`, etc.)
   - Focus on orchestration and state detection
   - Minimal new code is better

2. **State Machine**: Keep state detection simple:
   - Compare version (current vs latest) → decide stage
   - After stage 1, re-run same logic → should go to stage 2
   - After stage 2, everything is done

3. **Error Messages**: Be specific about what failed and how to recover:
   - "Package upgrade failed. Try manually: uv tool upgrade moai-adk"
   - "Template sync failed. Restore from backup: cp -r .moai-backups/backup/ .moai/"

4. **User Feedback**: Solicit feedback after release:
   - Ask if single-command experience is better
   - Track how many users use `update-complete` vs `update`
   - Plan Option C based on feedback

---

## Next Steps

1. **Review and Approval**
   - Get stakeholder feedback on design
   - Approve specification
   - Mark status as "approved"

2. **Implementation**
   - Create feature branch: `feature/SPEC-UPDATE-REFACTOR-003`
   - Follow TDD (RED → GREEN → REFACTOR)
   - Reference @TAG in commits

3. **Testing and QA**
   - 85%+ coverage target
   - Integration testing
   - Cross-platform testing

4. **Release**
   - Merge to main
   - Release as v0.7.0
   - Update CHANGELOG and README

---

**Document Status**: DRAFT | Ready for review and approval
**Target Release**: v0.7.0
**Effort Estimate**: 4 hours (3-5 hour range)

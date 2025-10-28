---
id: UPDATE-REFACTOR-002
version: 0.0.1
status: draft
created: 2025-10-28
updated: 2025-10-28
author: @Goos
priority: high
category: CLI Enhancement
labels: [self-update, package-upgrade, user-experience]
depends_on: []
blocks: []
related_specs: [SPEC-INIT-001]
related_issue: []
scope: package-upgrade-detection
---

# @SPEC:UPDATE-REFACTOR-002: moai-adk Self-Update Integration Feature

## HISTORY

### v0.0.1 (2025-10-28)
- **INITIAL**: Initial creation of moai-adk self-update integration specification
- **AUTHOR**: @Goos
- **SCOPE**: Automatic tool detection, 2-stage upgrade workflow, CLI option enhancements
- **CONTEXT**: Current update command only syncs templates. Users must manually run `uv tool upgrade` or `pip install --upgrade`. This SPEC adds automatic installer detection and integrated package upgrade flow.

---

## Environment

**Target Platforms**: macOS, Linux, Windows
**Python Version**: 3.13+
**Package Managers**: uv (tool), pipx, pip
**CLI Framework**: Click 8.1+
**Terminal UI**: Rich 13.0+

**Current Implementation Location**:
- Command: `src/moai_adk/cli/commands/update.py`
- Core: `src/moai_adk/core/template/processor.py`
- CLI Entry: `src/moai_adk/__main__.py`

---

## Assumptions

1. **Python Self-Update Limitation**: Running Python process cannot upgrade itself. Requires process restart (similar to pip's behavior).
2. **Tool Installation Detection**: Most users install via one primary tool (uv tool, pipx, or pip). Detection in order: uv tool → pipx → pip.
3. **Backward Compatibility**: Current `moai-adk update` behavior (template sync only) must be preserved. New upgrade feature is additive.
4. **Network Availability**: PyPI API is accessible. Graceful fallback if network fails.
5. **User Consent**: Users understand 2-stage workflow (upgrade → re-run for templates) is necessary for safety.

---

## Requirements

### Ubiquitous Requirements

1. **UBQ-001**: The system must detect which tool installed moai-adk (uv tool, pipx, or pip)
2. **UBQ-002**: The system must provide unified update experience regardless of installation method
3. **UBQ-003**: The system must fetch latest package version from PyPI
4. **UBQ-004**: The system must create backups before modifying project templates
5. **UBQ-005**: The system must support all CLI options: `--templates-only`, `--yes`, `--force`, `--check`
6. **UBQ-006**: The system must handle errors gracefully with helpful guidance

### Event-driven Requirements

1. **EVT-001**: WHEN user runs `moai-adk update` with upgrade available, THEN system executes Stage 1 (package upgrade) and prompts re-run
   - **Trigger**: User command with version mismatch
   - **Action**: Detect installer → Execute upgrade → Prompt re-run
   - **Response**: Clear message with next step

2. **EVT-002**: WHEN user runs `moai-adk update` after upgrade completion, THEN system executes Stage 2 (template sync)
   - **Trigger**: User re-runs after Stage 1 (current version == latest version)
   - **Action**: Create backup → Copy templates → Merge config → Set optimized flag
   - **Response**: Success message with optional next step guidance

3. **EVT-003**: WHEN user runs `moai-adk update --templates-only`, THEN system skips package upgrade and directly syncs templates
   - **Trigger**: User includes `--templates-only` flag
   - **Action**: Skip tool detection → Skip upgrade → Proceed to template sync
   - **Response**: Template sync completion message

4. **EVT-004**: WHEN user runs `moai-adk update --check`, THEN system displays version comparison and exits without changes
   - **Trigger**: User includes `--check` flag
   - **Action**: Fetch versions → Compare → Display → Exit
   - **Response**: Version info (e.g., "Current: 0.6.1, Latest: 0.6.2")

5. **EVT-005**: WHEN tool detection fails, THEN system provides fallback instructions and guidance
   - **Trigger**: uv tool, pipx, pip all not detected
   - **Action**: Display error → Suggest manual upgrade → Show expected location
   - **Response**: Helpful troubleshooting message

### State-driven Requirements

1. **STA-001**: WHILE current version equals latest version, the system must execute Stage 2 (template sync only)
   - **State condition**: `current_version >= latest_version`
   - **Action**: Skip upgrade → Proceed to sync
   - **Side effect**: Set `optimized: false` flag

2. **STA-002**: WHILE current version is less than latest version, the system must execute Stage 1 (package upgrade)
   - **State condition**: `current_version < latest_version`
   - **Action**: Upgrade package → Prompt re-run
   - **Side effect**: Show "Run again for templates" message

3. **STA-003**: WHILE project directory exists and contains `.moai/`, the system must safely merge backups
   - **State condition**: `.moai/` directory exists
   - **Action**: Create backup → Merge config.json and CLAUDE.md
   - **Side effect**: Generate `.moai-backups/` timestamp directory

### Optional Requirements

1. **OPT-001**: IF user includes `--yes` flag, THEN system auto-confirms all prompts without user input
2. **OPT-002**: IF user includes `--force` flag, THEN system overwrites without merge (for cleanup scenarios)
3. **OPT-003**: IF network is unavailable, THEN system gracefully falls back to template-sync-only mode with warning
4. **OPT-004**: IF installer detection is ambiguous (e.g., multiple tools installed), THEN system shows menu for user selection

### Constraints

1. **CST-001**: IF Python process is running the package, THEN the process cannot directly upgrade itself. 2-stage workflow is required for safety.
2. **CST-002**: IF tool detection is automatic, THEN uv tool detection must be checked first (highest priority)
3. **CST-003**: IF template merge is enabled, THEN config.json must preserve project metadata (name, author, locale)
4. **CST-004**: IF version comparison fails, THEN system must not allow upgrade to avoid partial updates
5. **CST-005**: IF backup already exists with same timestamp, THEN system must use unique timestamp suffix

---

## Traceability (@TAG)

**SPEC Chain**:
- **SPEC**: @SPEC:UPDATE-REFACTOR-002 (this document)
- **Related SPEC**: @SPEC:INIT-001 (initialization system)

**Implementation Chain** (to be populated in Phase 2):
- **TEST**: `tests/cli/commands/test_update.py`
- **CODE**: `src/moai_adk/cli/commands/update.py`
- **DOC**: `README.md`, `CHANGELOG.md`, `docs/installation.md`

---

## Implementation Strategy

### Phase 1: Core Logic
**Goal**: Detect installer and implement 2-stage workflow

**Deliverables**:
- `_detect_tool_installer()` function (50 lines)
- `_sync_templates()` function extracted (80 lines)
- 2-stage workflow logic in `update()` (100 lines)

**Key Functions**:
```python
def _detect_tool_installer() -> list[str] | None:
    """Detect which tool installed moai-adk.

    Returns:
        - ["uv", "tool", "upgrade", "moai-adk"]
        - ["pipx", "upgrade", "moai-adk"]
        - ["pip", "install", "--upgrade", "moai-adk"]
        - None if detection fails
    """

def _sync_templates(project_path: Path, force: bool) -> None:
    """Sync templates to project (extracted from current update logic)."""

def update(path: str, force: bool, check: bool,
           templates_only: bool, yes: bool) -> None:
    """Main update command with 2-stage workflow."""
```

### Phase 2: CLI Options
**Goal**: Add `--templates-only`, `--yes`, `--force`, `--check` options

### Phase 3: Error Handling
**Goal**: Graceful error recovery and helpful guidance

### Phase 4: Testing
**Goal**: 85%+ coverage with unit and integration tests

### Phase 5: Documentation
**Goal**: README, CHANGELOG, docstring updates

---

## Success Criteria

✅ **Core Functionality**:
- Automatic tool detection works on all platforms
- 2-stage workflow prompt is clear
- Users can re-run `moai-adk update` without additional arguments

✅ **User Experience**:
- One-command update flow (before: manual `uv tool upgrade` + `moai-adk update`)
- Clear messaging at each step
- Helpful error messages with recovery steps

✅ **Quality**:
- Test coverage ≥85%
- All ruff and mypy checks pass
- Platform-tested on macOS, Linux, Windows

✅ **Documentation**:
- README explains 2-stage workflow
- CHANGELOG documents new feature
- Docstrings clarify 2-stage behavior

---

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|-----------|
| Tool detection fails | Medium | Low | Fallback to manual instructions |
| Upgrade command fails | Medium | Low | Rollback option, error log |
| Template merge conflict | Low | Low | Smart merge preserving customizations |
| Platform compatibility | Medium | Low | Test on macOS/Linux/Windows |

---

## Next Steps

1. ✅ SPEC approval (this document)
2. → **`/alfred:2-run SPEC-UPDATE-REFACTOR-002`**: Implement core logic and CLI
3. → **`/alfred:3-sync`**: Synchronize docs and tests
4. → **Release v0.6.2**: Package upgrade with self-update feature

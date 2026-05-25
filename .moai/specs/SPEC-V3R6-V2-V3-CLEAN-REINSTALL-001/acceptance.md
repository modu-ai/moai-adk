---
id: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
title: "Acceptance Criteria — v2-to-v3 Clean Reinstall with v.2.x FLAT layout restoration"
version: "0.5.1"
status: implemented
created: 2026-05-25
updated: 2026-05-26
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "internal/cli, internal/defs, pkg/version, internal/template/templates/.claude/agents, .claude/agents"
lifecycle: spec-anchored
tags: "moai-update, v2-v3-migration, acceptance-criteria, traceability, layout-restoration"
tier: M
---

# Acceptance Criteria — SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001

## §A — AC Matrix (Bidirectional REQ↔AC Traceability)

The matrix below establishes 100% bidirectional mapping between REQ-VVCR-NNN requirements in `spec.md` §B and AC-VVCR-NNN acceptance criteria in this document §B.

| AC # | Severity | Linked REQ(s) | Linked HARD/SHOULD | Verification mode |
|------|----------|----------------|--------------------|-----------------|
| AC-VVCR-001 | MUST | REQ-VVCR-001, REQ-VVCR-002 | — | Go unit test (`v2_detection_test.go`) + integration test |
| AC-VVCR-002 | MUST | REQ-VVCR-003, REQ-VVCR-004 | HARD-5 | Integration test (backup directory presence + contents) |
| AC-VVCR-003 | MUST | REQ-VVCR-005, REQ-VVCR-006 | HARD-1, HARD-2, HARD-3, HARD-4 | Integration test (PRESERVE inventory hash invariance) |
| AC-VVCR-004 | MUST | REQ-VVCR-007, REQ-VVCR-008 | HARD-6 | Integration test (config conflict detection via hash diff) |
| AC-VVCR-005 | MUST | REQ-VVCR-009 | — | Go unit test (`dirs_test.go`) — enumerate extended DeprecatedPaths |
| AC-VVCR-006 | MUST | REQ-VVCR-010 | — | Integration test (template-managed surface removal) |
| AC-VVCR-007 | MUST | REQ-VVCR-011, REQ-VVCR-019 | — | Integration test (design asset removal + reinstall absent) |
| AC-VVCR-008 | MUST | REQ-VVCR-012, REQ-VVCR-013, REQ-VVCR-014, REQ-VVCR-020 | — | Integration test (TBD + 0-ref yaml + db dir removal) |
| AC-VVCR-009 | MUST | REQ-VVCR-015, REQ-VVCR-016 | — | Integration test (cache + agency dir removal) |
| AC-VVCR-010 | MUST | REQ-VVCR-017, REQ-VVCR-018 | — | Integration test (reinstall surface presence + PRESERVE untouched) |
| AC-VVCR-011 | MUST | REQ-VVCR-021, REQ-VVCR-022 | — | Integration test (MERGE-back path restoration) |
| AC-VVCR-012 | MUST | REQ-VVCR-023, REQ-VVCR-024 | — | Integration test (post-condition violation reporting) |
| AC-VVCR-013 | MUST | REQ-VVCR-025 | — | Integration test (`.agency/` detection → `runMigrateAgency` invocation) |
| AC-VVCR-014 | SHOULD | REQ-VVCR-026 | SHOULD-2 | Cross-platform CI verification (linux/darwin/windows) |
| AC-VVCR-015 | SHOULD | REQ-VVCR-027 | SHOULD-4 | Integration test (v3-project no-op idempotency) |
| AC-VVCR-016 | SHOULD | REQ-VVCR-028 | SHOULD-1 | Integration test (`--dry-run` flag output verification) |
| AC-VVCR-017 | SHOULD | REQ-VVCR-029 | SHOULD-3 | Integration test (telemetry event emission verification) |
| AC-VVCR-LR-001 | MUST | REQ-VVCR-LR-001, REQ-VVCR-LR-006 | HARD-7 | Filesystem inspection + git ls-tree comparison (flat layout match v.2.x baseline) |
| AC-VVCR-LR-002 | MUST | REQ-VVCR-LR-002 | — | Frontmatter inspection (SPEC-V3R6-AGENT-FOLDER-SPLIT-001 supersedence markers) |
| AC-VVCR-LR-003 | MUST | REQ-VVCR-LR-003 | — | Source inspection + build (`defs.AgentsMoaiSubdir` value unchanged, `go build ./...` passes) |
| AC-VVCR-LR-004 | MUST | REQ-VVCR-LR-004 | — | Cross-reference grep (0 matches for `.claude/agents/{core,meta,expert}/`) |
| AC-VVCR-LR-005 | MUST | REQ-VVCR-LR-005 | — | catalog.yaml inspection (all agent paths in FLAT form, hash anchors recomputed) |

**Coverage assertion**: 22 AC entries cover 35 requirements (29 REQ-VVCR + 6 REQ-VVCR-LR) + 7 HARD constraints + 4 SHOULD constraints. Each REQ appears in at least one AC; each HARD/SHOULD constraint appears in at least one AC.

## §B — AC Definitions (Given-When-Then)

### AC-VVCR-001 (MUST) — v2 detection heuristic correctness

**Given** a project at any state (v2, partial v2, rc1-leak, clean v3, or fresh/uninitialized)

**When** `moai update` invokes `detectV2Fingerprint(projectRoot)`

**Then** the function MUST return a `V2Fingerprint` struct where:
- `IsV2` is `true` if **any** of three signals fired:
  - **Signal 1 (V2DetectedViaVersion)**: `.moai/config/sections/system.yaml` `moai.version` field reads as `v2.*` (string with `v2.` prefix), OR `moai.version` field reads as an empty string, OR the `system.yaml` file itself is missing (`fs.ErrNotExist`). Rationale: v3 projects always carry `system.yaml` with a populated `moai.version` field; missing-file and empty-string cases most likely indicate a v2 project with stale or unmigrated config (broader Option α policy per design.md §B.1)
  - **Signal 2 (V2DetectedViaAgencyDir)**: `.agency/` legacy directory exists at the project root
  - **Signal 3 (V2DetectedViaDeprecatedPath)**: any path enumerated in extended `DeprecatedPaths` (43 entries per spec.md §A.4) exists in the project
- `IsV2` is `false` if all three signals are negative (i.e., `system.yaml` exists with `moai.version` matching `v3.*` AND `.agency/` is absent AND no DeprecatedPath enumeration hit)
- The struct carries individual signal flags (`V2DetectedViaVersion`, `V2DetectedViaAgencyDir`, `V2DetectedViaDeprecatedPath`) for telemetry / diagnostic purposes
- The `SignalDetails` map carries per-signal diagnostic strings (e.g., `version_signal: "system.yaml missing"`, `agency_signal: ".agency/ present"`, `deprecated_signal_first_hit: ".claude/agents/core"`)

The empty-string and missing-file branches are intentionally treated as positive v2 signals per Option α (broader detection); see EC-5 for the dedicated edge case + design.md §B.1 for the algorithm.

**Linked**: REQ-VVCR-001, REQ-VVCR-002
**Test**: `internal/cli/v2_detection_test.go` — table-driven tests covering all 3 signals × all 3 Signal-1 sub-states (`v2.*` prefix / empty string / file-missing) = 24 permutations minimum; assertions verify both the `IsV2` boolean and the per-signal flag fields

### AC-VVCR-002 (MUST) — Backup creation before any removal

**Given** a v2 project ready for clean reinstall
**When** `runCleanReinstall` begins execution
**Then** before any REMOVE operation is performed, a backup directory MUST exist at `.moai/backups/v2-to-v3-{ISO-8601-UTC}/` containing:
- A complete copy of `.moai/` directory at the moment immediately preceding REMOVE
- A complete copy of `.claude/` directory at the same moment
- Read permissions only (0o600) on the backup directory contents per the security expectations of `update_namespace_protect.go`

**Linked**: REQ-VVCR-003, REQ-VVCR-004 (canonical order), HARD-5
**Test**: Integration test verifying backup directory presence + hash equivalence of backup contents vs pre-REMOVE state

### AC-VVCR-003 (MUST) — PRESERVE inventory integrity (HARD-1/2/3/4 enforcement)

**Given** a v2 project with user-owned assets in scope:
- `.moai/specs/SPEC-USER-001/` containing spec.md + plan.md + acceptance.md
- `.moai/project/product.md` + `structure.md` + `tech.md`
- `.claude/skills/my-harness-myproject-specialist/SKILL.md`
- `.claude/agents/harness/myproject-domain-specialist.md`
- `.claude/agents/local/release-update-specialist.md`
- `.claude/commands/my-custom-command.md` (root level)
- `.claude/commands/teamtools/teamtools.md` (non-`moai/` subdirectory)

**When** `moai update` completes clean reinstall
**Then** all of the above paths MUST exist with byte-identical content (SHA-256 hash unchanged) compared to pre-update state.

**Linked**: REQ-VVCR-005, REQ-VVCR-006, HARD-1, HARD-2, HARD-3, HARD-4
**Test**: Integration test computing SHA-256 hashes before and after update; asserting equality for every PRESERVE-inventory path

### AC-VVCR-004 (MUST) — User-modified config preservation

**Given** a v2 project where the user modified `.moai/config/sections/quality.yaml` (changed `coverage_threshold: 85` to `coverage_threshold: 90`)
**When** `moai update` completes clean reinstall
**Then**:
- `.moai/config/sections/quality.yaml` MUST contain the user's `coverage_threshold: 90` value (not the v3 template default)
- `.moai/backups/v2-to-v3-{stamp}/config-conflicts/quality.yaml.template-baseline` MUST exist and contain the v3 template baseline content

**Linked**: REQ-VVCR-007, REQ-VVCR-008, HARD-6
**Test**: Integration test seeding a modified `quality.yaml` and verifying both post-conditions

### AC-VVCR-005 (MUST) — Extended DeprecatedPaths enumeration

**Given** `internal/defs/dirs.go` after M2 completion

**When** the test enumerates `DeprecatedPaths`

**Then** the slice MUST contain exactly **43 entries** total, decomposed per spec.md §A.4 Canonical DeprecatedPaths Derivation Table:

- **9 entries** in Category A (pre-existing per SPEC-AGENCY-ABSORB-001, preserved verbatim)
- **31 entries** in Category B (v.2.x-era NEW, all carry `DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"`, `DeprecatedBy: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"`, `RemovalSchedule: "v3.0.0"`)
- **3 entries** in Category C (rc1-stage staging artifacts, all carry `DeprecatedSince: "SPEC-V3R6-AGENT-FOLDER-SPLIT-001"`, `DeprecatedBy: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"`, `RemovalSchedule: "v3.0.0"`)
- **Total NEW entries added by this SPEC**: Category B + Category C = **34 entries**

The exact 43-row enumeration is in spec.md §A.4. The test MUST verify both the total count (43) and the per-category subtotals (9 / 31 / 3) by inspecting `DeprecatedSince` field values.

**Linked**: REQ-VVCR-009
**Test**: `internal/defs/dirs_test.go` table-driven enumeration test

### AC-VVCR-006 (MUST) — Template-managed surface removal + FLAT-layout reinstall

**Given** a v2 or rc1-leak project containing input-state template-managed surfaces (the layout the user arrived with):
- `.claude/skills/moai-foundation-core/`
- `.claude/skills/moai-workflow-spec/`
- `.claude/skills/moai/`
- `.claude/agents/moai/manager-spec.md` (v.2.x baseline FLAT path) — OR
- `.claude/agents/core/manager-spec.md` (rc1-leak split path; some early rc1 testers have this)
- `.claude/agents/expert/expert-backend.md` (rc1-leak path or v.2.x archived legacy path)
- `.claude/rules/moai/core/moai-constitution.md`
- `.claude/commands/moai/plan.md`
- `.claude/hooks/moai/handle-session-start.sh`
- `.claude/output-styles/moai/moai.md`

The AC accepts EITHER the v.2.x FLAT layout OR the rc1-leak split layout as Given-state input — the clean-reinstall code path normalizes both to the FLAT post-reinstall target.

**When** `runCleanReinstall` completes the full 7-step canonical order (REMOVE → reinstall → MERGE-back → integrity verify)

**Then** the following post-reinstall conditions MUST hold:

1. **Removed (intermediate state, verified at REMOVE-phase checkpoint before reinstall begins)**: ALL input-state paths above MUST be absent from the filesystem
2. **Reinstalled in FLAT layout**: the post-reinstall state MUST contain the v3 FLAT template-managed surfaces:
   - `.claude/skills/moai/` (template-managed `moai/` parent)
   - `.claude/agents/moai/manager-spec.md` (FLAT — no `core/` segment)
   - `.claude/agents/moai/builder-harness.md` (FLAT)
   - `.claude/agents/moai/plan-auditor.md` (FLAT; representative sample of the 7 retained agents)
   - `.claude/rules/moai/core/moai-constitution.md` (rules retain their own `core/` subdirectory — only `.claude/agents/` is FLAT under `moai/`)
   - `.claude/commands/moai/plan.md`
   - `.claude/hooks/moai/handle-session-start.sh`
   - `.claude/output-styles/moai/moai.md`
3. **FLAT-layout enforcement (negative assertion)**: the post-reinstall state MUST NOT contain ANY of these rc1-stage paths:
   - `.claude/agents/core/` (any file, directory, or subdirectory)
   - `.claude/agents/meta/` (any file, directory, or subdirectory)
   - `.claude/agents/expert/` (any file, directory, or subdirectory)
   - `.claude/agents/moai/core/` (no subdirectories under `moai/` either; FLAT means FLAT)
   - `.claude/agents/moai/meta/`
   - `.claude/agents/moai/expert/`

**Linked**: REQ-VVCR-010
**Test**: Integration test with three checkpoints — (a) REMOVE-phase completion (intermediate-state absence assertion), (b) reinstall-phase completion (FLAT presence assertion), (c) integrity-verify completion (negative assertion that no rc1-stage path exists). The Given-state asymmetry (accepts both v.2.x and rc1-leak inputs) is verified by table-driven test fixtures covering both input layouts.

### AC-VVCR-007 (MUST) — Design asset removal AND non-reinstallation

**Given** a v2 project containing all 8 design assets
**When** `runCleanReinstall` completes the full 7-step canonical order
**Then** ALL 8 design asset paths MUST be absent from the filesystem:
- `.claude/skills/moai-domain-brand-design/`
- `.claude/skills/moai-domain-copywriting/`
- `.claude/skills/moai-domain-design-handoff/`
- `.claude/skills/moai-workflow-design/`
- `.claude/skills/moai-workflow-gan-loop/`
- `.claude/rules/moai/design/`
- `.moai/project/brand/`
- `.moai/config/sections/design.yaml`

The reinstall phase MUST NOT recreate any of these paths (REQ-VVCR-019 unwanted-behavior verification).

**Linked**: REQ-VVCR-011, REQ-VVCR-019
**Test**: Integration test asserting absence after full clean reinstall

### AC-VVCR-008 (MUST) — TBD + 0-ref yaml + db directory removal

**Given** a v2 project containing:
- `.moai/config/sections/db.yaml` with `_TBD_` markers
- `.moai/config/sections/gate.yaml` (0 Go references)
- `.moai/config/sections/github-actions.yaml` (0 Go references)
- `.moai/config/sections/memo.yaml` (0 Go references)
- `.moai/db/` directory

**When** `runCleanReinstall` completes
**Then** all 5 paths MUST be absent from the filesystem, AND the reinstall phase MUST NOT recreate any of them (REQ-VVCR-020 unwanted-behavior verification).

**Linked**: REQ-VVCR-012, REQ-VVCR-013, REQ-VVCR-014, REQ-VVCR-020
**Test**: Integration test asserting absence after full clean reinstall

### AC-VVCR-009 (MUST) — Cache + agency directory removal

**Given** a v2 project containing:
- `.moai/cache/` directory (auto-regenerated; safe to remove)
- `.agency/` legacy directory
- `.agency.archived/` previous-attempt archive directory

**When** `runCleanReinstall` completes
**Then** all 3 paths MUST be absent from the filesystem. Note: `.moai/cache/` MAY be regenerated by subsequent runtime operations — this is acceptable; the assertion targets the post-update moment, not the post-runtime moment.

**Linked**: REQ-VVCR-015, REQ-VVCR-016
**Test**: Integration test asserting absence immediately after `moai update` completion

### AC-VVCR-010 (MUST) — Reinstall surface presence + PRESERVE untouched

**Given** a clean-reinstall in progress, immediately after the REMOVE phase
**When** the reinstall phase deploys the v3 embedded template baseline
**Then**:
- All v3 template-managed surfaces MUST be present at their canonical paths
- NO PRESERVE-inventory path is overwritten — every PRESERVE path retains its pre-REMOVE content

**Linked**: REQ-VVCR-017, REQ-VVCR-018
**Test**: Integration test asserting both conditions in sequence

### AC-VVCR-011 (MUST) — MERGE-back path restoration

**Given** a clean-reinstall in progress, immediately after the reinstall phase
**When** the MERGE-back phase executes
**Then**:
- Every PRESERVE-inventory path is restored to its pre-REMOVE location with byte-identical content
- User-modified config sections (per REQ-VVCR-022) are restored to `.moai/config/sections/{filename}` with the user's modifications intact (not the template baseline)

**Linked**: REQ-VVCR-021, REQ-VVCR-022
**Test**: Integration test asserting hash equality of restored paths

### AC-VVCR-012 (MUST) — Post-condition integrity verification

**Given** a clean-reinstall that completed all 7 canonical steps
**When** integrity verification executes
**Then**:
- All 4 post-conditions from REQ-VVCR-023 MUST be checked (PRESERVE existence, REMOVE absence, v3 surface presence, version field updated)
- If any condition is violated, a structured error report MUST be emitted naming the violated condition and the affected path
- The error report MUST reference the backup directory path `.moai/backups/v2-to-v3-{stamp}/` for recovery

**Linked**: REQ-VVCR-023, REQ-VVCR-024
**Test**: Integration test injecting a synthetic violation (e.g., a PRESERVE path is missing) and asserting the structured error report

### AC-VVCR-013 (MUST) — `.agency/` auto-migration

**Given** a v2 project containing `.agency/` legacy data with valid migration content
**When** `runCleanReinstall` begins
**Then** before the REMOVE phase, `runMigrateAgency` MUST be invoked, and the migration MUST complete successfully (moving `.agency/` content into `.moai/`) before any REMOVE operation begins.

**Linked**: REQ-VVCR-025
**Test**: Integration test with a synthetic `.agency/` directory containing migration-eligible data; assert post-migration `.moai/` state + `.agency/` absence

### AC-VVCR-014 (SHOULD) — Cross-platform compatibility

**Given** the CI matrix executing on linux/amd64, darwin/arm64, darwin/amd64, and windows/amd64
**When** the full integration test suite runs
**Then** all tests MUST pass on all 4 platform combinations. Platform-specific test failures (e.g., windows path separator handling) MUST be investigated and resolved before merge.

**Linked**: REQ-VVCR-026, SHOULD-2
**Test**: GitHub Actions CI workflow with the 4-platform matrix

### AC-VVCR-015 (SHOULD) — Idempotency on clean v3 project

**Given** a project already on v3 with no v2 fingerprint signals
**When** `moai update` is invoked
**Then** the v2-fingerprint heuristic MUST return `IsV2: false`, the clean-reinstall code path MUST NOT activate, and the existing file-level sync code path MUST proceed as before, leaving the project unchanged when no file deltas exist.

**Linked**: REQ-VVCR-027, SHOULD-4
**Test**: Integration test on a fixture v3 project; assert no file changes after `moai update`

### AC-VVCR-016 (SHOULD) — Dry-run output verification

**Given** a v2 project
**When** the user invokes `moai update --dry-run`
**Then**:
- The clean-reinstall code path MUST NOT perform any filesystem mutation
- Stdout MUST contain a clear listing of:
  - Planned PRESERVE inventory (enumeration of user-owned paths)
  - Planned REMOVE target list (enumeration of paths to be removed)
  - Planned reinstall surface list (enumeration of v3 paths to be deployed)
- Exit code MUST be 0 (success)

**Linked**: REQ-VVCR-028, SHOULD-1
**Test**: Integration test invoking `--dry-run` and asserting stdout content + filesystem unchanged

### AC-VVCR-017 (SHOULD) — Telemetry event emission

**Given** a v2 project with telemetry opt-in enabled
**When** `runCleanReinstall` activates
**Then** a telemetry event tagged `update.clean_reinstall.v2_to_v3` MUST be emitted (via the existing telemetry subsystem), carrying:
- `v2_signals_fired`: list of which v2-fingerprint signals triggered
- `remove_target_count`: integer count of paths removed
- `preserve_inventory_count`: integer count of paths preserved
- `dry_run`: boolean flag

Telemetry emission failure MUST NOT block the update operation (graceful degradation via `defer recover()`).

**Linked**: REQ-VVCR-029, SHOULD-3
**Test**: Integration test with a mock telemetry collector verifying event receipt

### AC-VVCR-LR-001 (MUST) — FLAT v.2.x layout restoration

**Given** the maintainer repo at the M2a completion commit (template + local both restored)
**When** the test inspects `internal/template/templates/.claude/agents/` and `.claude/agents/`
**Then** BOTH directories MUST satisfy:
- Contain a `moai/` subdirectory holding exactly 7 `.md` files (builder-harness, evaluator-active, manager-develop, manager-docs, manager-git, manager-spec, plan-auditor)
- `moai/` MUST NOT contain `core/`, `expert/`, or `meta/` subdirectories
- The repo MUST NOT contain `.claude/agents/core/`, `.claude/agents/expert/`, or `.claude/agents/meta/` directories at any level
- The flat 7-file list matches `git ls-tree -r 1bd083725^ --name-only | grep '\.claude/agents/moai/' | wc -l` shape (FLAT, no subdirectory depth), filtered to the 7 retained agents

**Linked**: REQ-VVCR-LR-001, REQ-VVCR-LR-006, HARD-7
**Test**: Filesystem inspection test in `internal/cli/v2_v3_layout_test.go` (NEW)

### AC-VVCR-LR-002 (MUST) — Predecessor SPEC supersedence markers

**Given** the M2a completion commit
**When** the test reads `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/spec.md` frontmatter
**Then**:
- `status:` field MUST equal `superseded`
- `superseded_by:` field MUST equal `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`
- HISTORY section MUST contain a row recording the supersedence (free-form prose; presence verified via grep for "superseded" within HISTORY section)

AND the test reads `.moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/spec.md` frontmatter:
- `supersedes:` field MUST contain `SPEC-V3R6-AGENT-FOLDER-SPLIT-001`

**Linked**: REQ-VVCR-LR-002
**Test**: Frontmatter parser test asserting all 3 markers

### AC-VVCR-LR-003 (MUST) — Constant value unchanged, build passes

**Given** the M2a completion commit
**When** the test inspects `internal/defs/dirs.go:112`
**Then** `AgentsMoaiSubdir` constant MUST equal `"agents/moai"` exactly (no `/core/` or `/meta/` insertion).

AND running the following commands in CI MUST all return exit 0:
- `go build ./...`
- `go test ./internal/cli/...`
- `go test ./internal/defs/...`

This verifies REQ-VVCR-LR-003 — no source-code changes were required for the layout restoration; the constant value already matched the FLAT shape.

**Linked**: REQ-VVCR-LR-003
**Test**: Source inspection + `go build` + targeted `go test` in CI

### AC-VVCR-LR-004 (MUST) — Cross-reference grep returns zero matches

**Given** the M2a completion commit
**When** the test executes:
```bash
grep -rn '\.claude/agents/core/\|\.claude/agents/meta/\|\.claude/agents/expert/' \
  .claude/skills/ \
  .claude/rules/ \
  .claude/agents/moai/ \
  CLAUDE.md \
  CLAUDE.local.md \
  .moai/specs/*/spec.md \
  .moai/specs/*/plan.md \
  .moai/specs/*/acceptance.md \
  .moai/specs/*/design.md \
  .moai/specs/*/research.md \
  2>/dev/null
```
**Then** the command MUST return zero matches (exit 1 with no output, or exit 0 with empty output).

Exceptions intentionally permitted:
- This SPEC's own body (`.moai/specs/SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001/*.md`) MAY reference the rc1-stage paths because the SPEC documents the supersedence — these references describe what is being removed
- The superseded SPEC's body (`.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/*.md`) MAY reference the rc1-stage paths because that SPEC defined them
- `internal/cli/update_cleanup.go` and `internal/defs/dirs.go` MAY reference the rc1-stage paths because they enumerate them as DeprecatedPaths targets

**Linked**: REQ-VVCR-LR-004
**Test**: Grep verification test in `internal/cli/v2_v3_layout_test.go` (filters the 3 permitted exception scopes)

### AC-VVCR-LR-005 (MUST) — catalog.yaml regenerated for FLAT layout

**Given** the M5 completion commit (after `gen-catalog-hashes.go --all` invocation)
**When** the test reads `internal/template/catalog.yaml`
**Then**:
- All agent path entries MUST match the pattern `.claude/agents/moai/<agent>.md` (single segment between `moai/` and the filename)
- ZERO entries MUST match `.claude/agents/core/`, `.claude/agents/expert/`, or `.claude/agents/meta/`
- Hash anchors for the 7 retained agents MUST be recomputed (verified by comparing the post-M5 catalog.yaml against the pre-M2a catalog.yaml — the hash values MUST differ if the byte content of the agent files moved without modification; if hash values are identical, the regeneration step was skipped erroneously)

**Linked**: REQ-VVCR-LR-005
**Test**: catalog.yaml parser test verifying path pattern + hash regeneration

## §C — Edge Cases

### EC-1 — Concurrent `moai update` invocations

**Scenario**: Two terminals invoke `moai update` simultaneously on the same project root.
**Expected**: The existing file lock in `internal/cli/update.go` (predecessor SPEC infrastructure) prevents concurrent execution. The second invocation MUST receive a clear "another update in progress" error and exit with non-zero status.

### EC-2 — Backup directory disk-full

**Scenario**: The user's disk fills during backup creation (Step 3 of canonical order).
**Expected**: The clean-reinstall code path MUST detect the disk-full error, abort before any REMOVE operation, emit a structured error message naming the backup path and the disk-full condition, and exit with non-zero status. No REMOVE operation MUST occur.

### EC-3 — Interrupted clean reinstall (SIGTERM mid-REMOVE)

**Scenario**: The user sends SIGTERM during the REMOVE phase.
**Expected**: The clean-reinstall code path's defer cleanup MUST point the user to the backup directory `.moai/backups/v2-to-v3-{stamp}/` for manual recovery. Subsequent `moai update` invocations MUST detect the partial state and offer recovery guidance (delegated to a future SPEC; this SPEC only ensures the backup is available).

### EC-4 — Permission-denied on a PRESERVE inventory path

**Scenario**: A file in the PRESERVE inventory has restrictive permissions preventing the clean-reinstall code path from reading it for snapshot.
**Expected**: The clean-reinstall code path MUST abort during Step 2 (PRESERVE inventory snapshot), emit a structured error naming the permission-denied path, and exit with non-zero status. No backup, REMOVE, or reinstall operation MUST occur.

### EC-5 — `.moai/config/sections/system.yaml` missing OR empty-string version (Option α broader policy)

**Scenario A — file missing**: The project lacks `.moai/config/sections/system.yaml` entirely.
**Expected**: The v2-fingerprint heuristic MUST treat the missing file as a positive v2 signal (Signal 1 fires with `version_signal: "system.yaml missing"`). v3 projects always carry `system.yaml`, so absence is most likely a v2 project that never migrated; clean-reinstall proceeds.

**Scenario B — empty-string version**: The project's `system.yaml` exists, is well-formed YAML, but the `moai.version` field is present with an empty-string value (`moai.version: ""` or `moai.version: null`).
**Expected**: The v2-fingerprint heuristic MUST treat the empty-string version as a positive v2 signal (Signal 1 fires with `version_signal: "moai.version=" + version`, where the value after `=` is empty). Rationale: an empty version field most likely indicates a v2 project with stale config rather than a fresh v3 install (broader Option α policy per design.md §B.1); clean-reinstall proceeds.

**Scenario C — unrecognized version format**: The project's `moai.version` field reads as a non-empty string that does NOT start with `v2.` (e.g., `v3.0.0-rc1`, `v3.0.0-rc2`, or any future v3+ version).
**Expected**: The v2-fingerprint heuristic MUST treat this as a NEGATIVE Signal 1 (the version is recognized as non-v2 — v3 or later). Signals 2 and 3 are still independently evaluated; if both negative, `IsV2: false` and the clean-reinstall code path does NOT activate (idempotency preserved per REQ-VVCR-027).

The 3-scenario distinction is verified by `internal/cli/v2_detection_test.go` with explicit fixture cases.

## §D — Quality Gate / Definition of Done

The SPEC is considered done when all of the following are true:

- [ ] All 22 AC entries (17 AC-VVCR + 5 AC-VVCR-LR) pass on the canonical CI matrix
- [ ] All 18 MUST-PASS AC entries (AC-VVCR-001 through AC-VVCR-013 + AC-VVCR-LR-001 through AC-VVCR-LR-005) pass without exception
- [ ] At least 3 of the 4 SHOULD AC entries (AC-VVCR-014 through AC-VVCR-017) pass; any SHOULD failures are documented in `progress.md` as documented debt with rationale
- [ ] 5 edge cases (EC-1 through EC-5) have at least one integration test asserting expected behavior
- [ ] `pkg/version/version.go` shows `Version = "v3.0.0-rc2"`
- [ ] CHANGELOG.md contains a v3.0.0-rc2 entry describing the clean-reinstall paradigm shift AND the FLAT layout restoration
- [ ] `internal/defs/dirs.go` `DeprecatedPaths` slice contains exactly 43 entries (9 pre-existing Category A + 31 v.2.x-era Category B + 3 rc1-stage Category C per spec.md §A.4)
- [ ] Maintainer repo FLAT layout restored: `.claude/agents/moai/` flat with 7 agents, no `{core,expert,meta}/` directories anywhere under `.claude/agents/`
- [ ] Cross-reference grep returns zero matches for rc1-stage paths (per AC-VVCR-LR-004)
- [ ] `internal/template/catalog.yaml` regenerated via `gen-catalog-hashes.go --all`, paths reflect FLAT layout, hashes recomputed
- [ ] SPEC-V3R6-AGENT-FOLDER-SPLIT-001 frontmatter shows `status: superseded` + `superseded_by: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`
- [ ] Go test suite passes: `go test ./internal/cli/... ./internal/defs/... ./pkg/version/...`
- [ ] Linter passes: `golangci-lint run` returns 0 issues
- [ ] Cross-platform CI verification matrix (linux/darwin/windows) all green
- [ ] No new lint warnings introduced beyond the pre-SPEC baseline

## §E — HISTORY

### v0.3.0 (2026-05-25) — FLAT layout simplification

- AC-VVCR-LR-001 wording: FLAT layout (7 agents directly under `.claude/agents/moai/`, no subdirectories)
- AC-VVCR-LR-004 grep target literals: `.claude/agents/{core,meta,expert}/` (no `.claude/agents/moai/{core,meta}/` insertion to grep — the FLAT target means NO subdirectory under moai/ is permitted)
- AC-VVCR-LR-005 verification: catalog.yaml paths match `\.claude/agents/moai/<agent>\.md` single-segment pattern (FLAT)
- DoD updated: maintainer repo FLAT layout checkpoint replaces v0.2.0's `{core,meta}/` subdirectory checkpoint

### v0.2.0 (2026-05-25) — Layout Restoration AC entries

- 5 new AC-VVCR-LR-NNN entries added (all MUST severity): AC-VVCR-LR-001 through AC-VVCR-LR-005
- AC count grows from 17 to 22 (13 MUST + 4 SHOULD + 5 new MUST = 18 MUST + 4 SHOULD)
- REQ count grows from 29 to 35 (29 REQ-VVCR + 6 REQ-VVCR-LR); bidirectional traceability preserved
- HARD-7 added to AC matrix linkage column for AC-VVCR-LR-001
- DoD checklist extended: maintainer repo layout, cross-reference grep, catalog regeneration, supersedence markers, expanded DeprecatedPaths count (25 → 28 minimum)

### v0.1.0 (2026-05-25) — Initial draft

- 17 AC-VVCR criteria with 100% REQ↔AC bidirectional mapping (29 REQ → 17 AC, 13 MUST + 4 SHOULD split)
- 6 HARD constraint references + 4 SHOULD constraint references all covered
- 5 edge cases enumerated (EC-1 through EC-5)
- Definition of Done with 12 explicit checkpoints
- Aligned with `acceptance.md` 12-canonical-field frontmatter schema

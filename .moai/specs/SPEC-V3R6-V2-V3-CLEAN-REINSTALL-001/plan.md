---
id: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
title: "Plan — v2-to-v3 Clean Reinstall with v.2.x FLAT layout restoration"
version: "0.5.1"
status: implemented
created: 2026-05-25
updated: 2026-05-26
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "internal/cli, internal/defs, pkg/version, internal/template/templates/.claude/agents, .claude/agents"
lifecycle: spec-anchored
tags: "moai-update, v2-v3-migration, plan-phase, milestone-decomposition, layout-restoration, flat-agents-moai"
tier: M
---

# Plan — SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001

## §A — Context

### §A.1 Predecessor Audit

The orchestrator conducted a pre-plan audit (recorded in the spawn context) that determined the maintainer machine state (this repo) already has all v2 artifacts cleaned: `.agency/` and `.agency.archived/` absent; 4 agency agents absent; 12 archived agents already in `.moai/backups/agent-archive-2026-05-25/`; `.moai/memory/` absent (auto-migrated by existing `migrateLegacyMemoryDir`). The SPEC therefore addresses **external user expectations** rather than maintainer machine state.

### §A.2 Existing Infrastructure Reuse

The plan reuses existing Go infrastructure verbatim where applicable:

| File | Purpose | Reuse mode |
|------|---------|------------|
| `internal/cli/migrate_agency.go` (21KB) | `.agency/` → `.moai/` migration with transaction log + checkpoint + dry-run + resume | Invoked from new clean-reinstall code path via `runMigrateAgency` (REQ-VVCR-025) |
| `internal/cli/update_cleanup.go` (16KB) | `scanDeprecatedPaths` + `backupDeprecatedPaths` + classification | Extended with new entries from REQ-VVCR-009 list |
| `internal/cli/update_namespace_protect.go` (9.6KB) | §24 namespace contract (`isUserOwnedNamespace`, `newNamespaceBackupStamp`, `resolveNamespaceBackupDir`, `collectUserOwnedFiles`, `backupUserOwnedNamespace`) | Invoked verbatim from new clean-reinstall code path; `NamespaceBackupsSubdir` constant from `internal/defs/dirs.go:22` is reused for the `.moai/backups/v2-to-v3-{stamp}/` path |
| `internal/cli/update.go:1731` `migrateLegacyMemoryDir` precedent | Auto-invoke migration pattern | Pattern replicated for `runMigrateAgency` integration (REQ-VVCR-025) |
| `internal/defs/dirs.go` `DeprecatedPathEntry` struct + `DeprecatedPaths` slice | Extension point | Extended with 34 NEW entries (per spec.md §A.4: 31 Category B + 3 Category C); Category B entries carry `DeprecatedBy: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"` |

### §A.3 Boundaries with Existing `moai migrate agency`

The existing `moai migrate agency` command (SPEC-AGENCY-ABSORB-001) remains operational as an independent command. The new clean-reinstall path integrates `runMigrateAgency` via auto-invocation when `.agency/` is detected (REQ-VVCR-025) — this avoids forcing users to remember a two-step `migrate then update` flow.

## §B — Known Issues / Risks

| Risk | Mitigation |
|------|------------|
| User-modified config detection false-positive | Use SHA-256 hash diff against the embedded template baseline at the same v3 version. False positives (whitespace-only changes) cause unnecessary preservation but never data loss. |
| Backup directory disk space | Document the backup size (full `.moai/` + `.claude/` snapshot) in the dry-run output; users on tight-disk systems can elect manual cleanup post-verification. |
| Cross-platform path normalization (windows backslash) | Reuse the slash-separated path conventions already used by `internal/defs/dirs.go` `DeprecatedPathEntry.Path`. |
| Telemetry event emission failure | Wrap telemetry call in `defer recover()` — telemetry failure MUST never block the update operation. |
| Concurrent `moai update` invocations | Reuse the existing file lock in `internal/cli/update.go` (predecessor SPEC infrastructure); no new locking primitive required. |

## §C — Pre-Flight Audit

Pre-flight ground truth verification (run before M1 start):

```bash
# Verify infrastructure files exist
ls -la internal/cli/migrate_agency.go internal/cli/update_cleanup.go internal/cli/update_namespace_protect.go internal/cli/update.go

# Verify _TBD_ markers present in target files
grep -l "_TBD_" internal/template/templates/.moai/config/sections/*.yaml

# Verify zero-Go-reference yaml files (gate, github-actions, memo)
for f in gate github-actions memo; do
  echo "=== $f.yaml references in internal/ ===";
  grep -rn "sections/$f.yaml\|$f\.yaml" internal/ | wc -l;
done

# Verify current version
grep -n 'Version = ' pkg/version/version.go
```

Expected results: all 4 infrastructure files present; `db.yaml` and `design.yaml` contain `_TBD_`; gate/github-actions/memo each return 0; current version is `v3.0.0-rc1`.

## §D — Constraints (mirrored from spec.md §C)

Refer to `spec.md` §C for the canonical HARD-1 through HARD-6 and SHOULD-1 through SHOULD-4 constraint enumeration. Plan-phase implementation MUST honor every HARD constraint without exception.

## §E — Self-Verification Checklist (manager-spec pre-handoff)

- [x] All 35 REQ entries (29 REQ-VVCR + 6 REQ-VVCR-LR) use GEARS notation (Ubiquitous / When / While / Where-capability / Unwanted) — zero IF/THEN occurrences
- [x] Bidirectional REQ↔AC traceability: 35 REQs → 22 ACs (each AC covers 1-4 REQs; each REQ appears in at least one AC)
- [x] All 7 HARD constraints (HARD-1 through HARD-7) are traced in at least one AC entry (HARD-1/2/3/4 all link to AC-VVCR-003 PRESERVE inventory integrity; HARD-5 links to AC-VVCR-002 backup-before-removal; HARD-6 links to AC-VVCR-004 user-modified config preservation; HARD-7 links to AC-VVCR-LR-001 FLAT layout restoration — the 1:N mapping is intentional per Tier M traceability convention)
- [x] 7 explicit out-of-scope sections cover known scope creep risks (D.1 through D.7)
- [x] SPEC ID matches canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`
- [x] Frontmatter 12-canonical-field schema verified in spec.md / plan.md / acceptance.md / design.md / research.md
- [x] Optional `supersedes: [SPEC-V3R6-AGENT-FOLDER-SPLIT-001]` field present on spec.md per v.2.x baseline principle
- [x] Status `draft` consistent across all 5 SPEC body artifacts
- [x] Cross-references to SPEC-AGENCY-ABSORB-001 + SPEC-V3R3-UPDATE-CLEANUP-001 + SPEC-V3R6-AGENT-TEAM-REBUILD-001 + SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 + SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 + SPEC-V3R6-AGENT-FOLDER-SPLIT-001 (superseded) documented in `depends_on` + `related_specs` + `supersedes` frontmatter
- [x] 5-artifact set complete (spec.md + plan.md + acceptance.md + design.md + research.md)
- [x] v.2.x baseline principle (§A.0) anchored with git archaeology proof (`git ls-tree -r 1bd083725^`)
- [x] FLAT `.claude/agents/moai/` layout (7 retained agents directly, no subdirectories) consistent across spec.md §A.0, REQ-VVCR-LR-001, REQ-VVCR-LR-004, plan.md M2a deliverables

## §F — Milestone Decomposition

### M1 — Version bump and CHANGELOG entry

**Priority**: High (blocker for all subsequent milestones since v3.0.0-rc2 must exist as a coherent target)

**Deliverables**:
- `pkg/version/version.go` line 8: `Version = "v3.0.0-rc1"` → `Version = "v3.0.0-rc2"`
- `CHANGELOG.md` new entry under `## [Unreleased]` → `## [v3.0.0-rc2]` heading describing the v2-to-v3 clean reinstall paradigm shift
- `internal/template/templates/.moai/config/config.yaml` `moai.version` field update to `v3.0.0-rc2`
- `.moai/config/sections/system.yaml` `moai.version` field update (project-local mirror)

**File count**: 4 files
**Estimated LOC delta**: ~30 lines

**Dependencies**: none
**Blocks**: M2-M6

### M2 — Extend `DeprecatedPaths` table

**Priority**: High (drives REMOVE phase scope; M4 depends on this enumeration)

**Deliverables**:
- `internal/defs/dirs.go` `DeprecatedPaths` slice extension with new entries covering all v.2.x → v3 deltas. Note: the v.2.x baseline placed all 19 template-managed agents flat under `.claude/agents/moai/<name>.md` (verified by `git ls-tree -r 1bd083725^`), so the 12 archived agent paths use the `moai/<name>.md` form. The rc1-stage `.claude/agents/{core,expert,meta}/` directories are SEPARATE staging-artifact entries (DeprecatedSince: SPEC-V3R6-AGENT-FOLDER-SPLIT-001, DeprecatedBy: this SPEC) covering rc1-stage adopters who already have the split layout.
  - `.agency` + `.agency.archived` (v2 directories)
  - 4 agency agents (`.claude/agents/moai/{planner,designer,builder,evaluator}.md`) — v.2.x flat layout
  - 4 retired manager agents (`.claude/agents/moai/manager-{strategy,quality,brain,project}.md`) — v.2.x flat layout
  - 2 retired meta agents (`.claude/agents/moai/{claude-code-guide,researcher}.md`) — v.2.x flat layout
  - 6 retired expert agents (`.claude/agents/moai/expert-{backend,frontend,security,devops,performance,refactoring}.md`) — v.2.x flat layout
  - **rc1-stage staging artifacts** (3 directory-level entries; DeprecatedSince: `SPEC-V3R6-AGENT-FOLDER-SPLIT-001`, DeprecatedBy: `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`):
    - `.claude/agents/core` (rc1-stage split directory; never present in v.2.x)
    - `.claude/agents/expert` (rc1-stage split directory; never present in v.2.x; also empty after SPEC-V3R6-AGENT-TEAM-REBUILD-001 M3 archive)
    - `.claude/agents/meta` (rc1-stage split directory; never present in v.2.x)
  - 5 deprecated config yaml files (`.moai/config/sections/{design,db,gate,github-actions,memo}.yaml`)
  - 5 design assets (`.claude/skills/moai-domain-{brand-design,copywriting,design-handoff}/`, `.claude/skills/moai-workflow-{design,gan-loop}/`)
  - 1 design rule directory (`.claude/rules/moai/design/`)
  - 1 brand directory (`.moai/project/brand`)
  - 1 db directory (`.moai/db`)
- All v.2.x-era entries carry `DeprecatedSince: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"`, `DeprecatedBy: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"`, `RemovalSchedule: "v3.0.0"`
- All rc1-stage staging-artifact entries carry `DeprecatedSince: "SPEC-V3R6-AGENT-FOLDER-SPLIT-001"`, `DeprecatedBy: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"`, `RemovalSchedule: "v3.0.0"`
- Unit test `internal/defs/dirs_test.go` enumerating both groups of new entries

**Total NEW entries**: 31 Category B + 3 Category C = **34 NEW entries**, extending `DeprecatedPaths` from 9 (Category A pre-existing) to 43 total — per spec.md §A.4 Canonical DeprecatedPaths Derivation Table

**File count**: 2 files (1 modified, 1 new test file)
**Estimated LOC delta**: ~170 lines (mostly enumeration table; 31 Category B v.2.x-era + 3 Category C rc1-stage NEW entries per §A.4)

**Dependencies**: M1 (version constant must be v3.0.0-rc2 for table consistency)
**Blocks**: M2a, M4, M5

### M2a — v.2.x FLAT Layout Restoration (NEW milestone added v0.2.0, refined v0.3.0)

**Priority**: High (mechanical revert of SPEC-V3R6-AGENT-FOLDER-SPLIT-001 commit `1bd083725` before v3 ships; blocks M5 catalog regeneration)

**Pre-flight verification** (run immediately at M2a start):

```bash
# Verify v.2.x baseline was FLAT (no core/meta/expert under moai/)
git ls-tree -r 1bd083725^ --name-only | grep '\.claude/agents/moai/' | head -25
# Expected: ~19 .md files directly under .claude/agents/moai/, no subdirectories

# Verify current rc1 layout has the split (which we are reverting)
ls -la internal/template/templates/.claude/agents/
ls -la .claude/agents/
# Expected: core/ and meta/ subdirectories present, moai/ absent or empty
```

**Deliverables (mechanical revert + cross-reference fix)**:

1. **Template git mv operations (7 retained agents)**:
   - `git mv internal/template/templates/.claude/agents/core/manager-develop.md internal/template/templates/.claude/agents/moai/manager-develop.md`
   - Repeat for `manager-docs.md`, `manager-git.md`, `manager-spec.md` (4 files under `core/`)
   - `git mv internal/template/templates/.claude/agents/meta/builder-harness.md internal/template/templates/.claude/agents/moai/builder-harness.md`
   - Repeat for `evaluator-active.md`, `plan-auditor.md` (3 files under `meta/`)
   - Total template git mv: 7 operations

2. **Local git mv operations (7 retained agents)**:
   - Same 7 operations on local `.claude/agents/{core,meta}/*.md` → `.claude/agents/moai/*.md`

3. **Empty directory removal (after git mv completes)**:
   - `rmdir internal/template/templates/.claude/agents/{core,expert,meta}` (3 template empty dirs)
   - `rmdir .claude/agents/{core,expert,meta}` (3 local empty dirs)
   - Total empty-directory removal: 6 operations

4. **Cross-reference grep + replace (REQ-VVCR-LR-004)**:
   - Scope: `.claude/skills/`, `.claude/rules/`, `.claude/agents/moai/` (own agent frontmatter), `CLAUDE.md`, `CLAUDE.local.md`, `.moai/specs/*/spec.md`, `.moai/specs/*/plan.md`, `.moai/specs/*/acceptance.md`, `.moai/specs/*/design.md`, `.moai/specs/*/research.md`
   - Literals to replace:
     - `.claude/agents/core/<agent>.md` → `.claude/agents/moai/<agent>.md` (flat — no subdirectory segment)
     - `.claude/agents/meta/<agent>.md` → `.claude/agents/moai/<agent>.md` (flat)
     - `.claude/agents/core/` (directory references) → `.claude/agents/moai/`
     - `.claude/agents/meta/` (directory references) → `.claude/agents/moai/`
   - Verification: `grep -rn '\.claude/agents/core/\|\.claude/agents/meta/\|\.claude/agents/expert/' <scope>` returns 0 matches post-replace
   - Estimated cross-reference hits: 22 active rule/skill ref updates in ~19 unique files (per the original 1bd083725 commit message), plus ~5-8 SPEC body references → ~30 total grep hits

5. **Predecessor SPEC supersedence marker (REQ-VVCR-LR-002)**:
   - Edit `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/spec.md` frontmatter:
     - `status: implemented` → `status: superseded`
     - Add `superseded_by: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`
     - Add HISTORY row recording the supersedence event
   - Edit `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/progress.md` if present: append supersedence note

6. **Go source verification (REQ-VVCR-LR-003 — no code change expected)**:
   - Confirm `internal/defs/dirs.go:112` reads `AgentsMoaiSubdir = "agents/moai"` (already correct)
   - Run `go build ./...` post-restructure to verify no path-resolution regression
   - Run targeted tests: `go test ./internal/cli/... ./internal/defs/...`

**File count**: ~14 git mv targets + 6 empty-directory removals + ~19 cross-reference edits + 1-2 predecessor SPEC frontmatter edits ≈ ~24 file-level changes; estimated source LOC delta is small (~50 lines net — mostly path-text replacements) because git mv preserves file content

**Estimated LOC delta**: ~50 lines (most of the work is mechanical `git mv` + path-string substitution; no new logic)

**Dependencies**: M2 (DeprecatedPaths table must include the 3 rc1-stage staging-artifact entries before any external-user `moai update` invocation; M2a only reverts the maintainer-repo layout)
**Blocks**: M5 (catalog.yaml regeneration must run after layout is final)

### M3 — v2 detection logic

**Priority**: High (gate for clean-reinstall code path activation)

**Deliverables**:
- `internal/cli/v2_detection.go` (NEW file) implementing the v2-fingerprint heuristic:
  - `func detectV2Fingerprint(projectRoot string) (V2Fingerprint, error)` returning a struct with detected signals
  - Signal 1: `.moai/config/sections/system.yaml` `moai.version` field reading; if missing or starts with `v2.` returns `V2DetectedViaVersion`
  - Signal 2: `.agency/` legacy directory presence returns `V2DetectedViaAgencyDir`
  - Signal 3: any path enumerated in extended `DeprecatedPaths` returns `V2DetectedViaDeprecatedPath`
  - Heuristic resolution: ANY signal positive → `IsV2: true`
- `internal/cli/v2_detection_test.go` (NEW) with table-driven tests covering each signal and combinations

**File count**: 2 files (both new)
**Estimated LOC delta**: ~250 lines

**Dependencies**: M2 (relies on extended `DeprecatedPaths`)
**Blocks**: M4, M5

### M4 — Clean reinstall implementation

**Priority**: High (core deliverable)

**Deliverables**:
- `internal/cli/update_clean_install.go` (NEW file) implementing:
  - `func runCleanReinstall(ctx context.Context, projectRoot string, dryRun bool, out io.Writer) error` — orchestrates the 7-step canonical order from REQ-VVCR-004
  - Step 1: `detectV2Fingerprint` invocation (from M3)
  - Step 2: PRESERVE inventory snapshot via `collectUserOwnedFiles` (reused from `update_namespace_protect.go:100`) + extended user-owned path enumeration per REQ-VVCR-005
  - Step 3: Backup creation at `.moai/backups/v2-to-v3-{newNamespaceBackupStamp()}/` (reusing `newNamespaceBackupStamp` + `resolveNamespaceBackupDir`)
  - Step 4: REMOVE phase invoking `scanDeprecatedPaths` + `backupDeprecatedPaths` from `update_cleanup.go` and removing all template-managed namespace surfaces
  - Step 5: Reinstall phase invoking the embedded template deployment logic
  - Step 6: MERGE-back phase restoring PRESERVE inventory paths
  - Step 7: Integrity verification per REQ-VVCR-023
- `internal/cli/update_preserve_inventory.go` (NEW file) implementing:
  - `func buildPreserveInventory(projectRoot string) (PreserveInventory, error)` enumerating user-owned paths per REQ-VVCR-005
  - `func detectUserModifiedConfigs(projectRoot string) ([]string, error)` per REQ-VVCR-007 (SHA-256 hash diff)
  - `func snapshotPreserveInventory(inventory PreserveInventory, backupDir string) error`
- `internal/cli/update_clean_install_test.go` (NEW) with integration tests covering scenarios A (full v2 cleanup), B (partial v2 — only `.agency/`), C (clean v3 — no-op)
- `internal/cli/update_preserve_inventory_test.go` (NEW) with table-driven PRESERVE inventory tests

**File count**: 4 files (all new)
**Estimated LOC delta**: ~550 lines

**Dependencies**: M2, M3
**Blocks**: M5

### M5 — Integration into `runUpdate` flow + catalog regeneration

**Priority**: High (wires the new path into the existing `moai update` CLI + makes catalog reflect restored FLAT layout)

**Deliverables**:
- `internal/cli/update.go` modification at `runUpdate` (line 126):
  - Add v2-fingerprint detection invocation immediately after binary update check
  - If `IsV2: true` → invoke `runCleanReinstall` and return (REQ-VVCR-002)
  - If `IsV2: false` → continue with existing file-level sync code path (REQ-VVCR-027 idempotency)
- `internal/cli/update.go` `runMigrateAgency` integration:
  - At the start of `runCleanReinstall`, check for `.agency/` presence (REQ-VVCR-025)
  - If present, invoke `runMigrateAgency` before any REMOVE operations
- `internal/cli/update.go` modification adjacent to existing `migrateLegacyMemoryDir` invocation (line 1731) — follow the established pattern verbatim
- **catalog.yaml regeneration (REQ-VVCR-LR-005)**: After M2a completes the FLAT layout restoration, run `go run ./internal/template/scripts/gen-catalog-hashes.go --all` to regenerate `internal/template/catalog.yaml` so all path entries and hash anchors point at `.claude/agents/moai/<agent>.md` (FLAT) instead of the rc1-stage `.claude/agents/{core,expert,meta}/<agent>.md` (split). Commit the regenerated catalog as part of M5.

**File count**: 2 files (update.go + catalog.yaml)
**Estimated LOC delta**: ~50 lines code (update.go) + ~30 lines catalog.yaml regen (path entry path string changes, hash anchor recomputation)

**Dependencies**: M2a (FLAT layout must be in place before catalog regeneration), M4 (`runCleanReinstall` orchestration must exist before integration)
**Blocks**: M6

### M6 — Test coverage and cross-platform verification

**Priority**: Medium (quality assurance milestone; gates sync-phase entry)

**Deliverables**:
- Integration tests covering 3 canonical scenarios:
  - Scenario A: Full v2 project with all 25+ zombie paths present → verify clean removal + PRESERVE intact
  - Scenario B: Partial v2 project with only `.agency/` legacy data → verify `runMigrateAgency` auto-invoke + clean removal
  - Scenario C: Clean v3 project → verify no-op (REQ-VVCR-027)
- PRESERVE assertions for each scenario:
  - `.moai/specs/` content unchanged (file count + content hash)
  - `.moai/project/{product,structure,tech}.md` content unchanged
  - `.claude/skills/my-harness-*` content unchanged (if present)
  - `.claude/agents/harness/` content unchanged (if present)
  - `.claude/agents/local/` content unchanged (if present)
  - `.claude/commands/` root files and non-`moai/` subdirectories unchanged
- Cross-platform CI verification: linux/amd64, darwin/arm64, darwin/amd64, windows/amd64 (REQ-VVCR-026)
- Dry-run output verification (REQ-VVCR-028)
- Telemetry event emission verification (REQ-VVCR-029)

**File count**: 3-5 test files
**Estimated LOC delta**: ~400 lines

**Dependencies**: M4, M5
**Blocks**: sync-phase entry

### Total Estimated Scope

- **File count**: ~38-42 files total
  - M1: 4 files (version bump + CHANGELOG + 2 config mirrors)
  - M2: 2 files (dirs.go + dirs_test.go)
  - M2a: ~24 files (7 template git mv + 7 local git mv + 6 empty-dir removals + ~19 cross-reference edits + 1-2 predecessor SPEC frontmatter edits)
  - M3: 2 files (v2_detection.go + test)
  - M4: 4 files (update_clean_install.go + preserve_inventory.go + 2 test files)
  - M5: 1 file (update.go integration)
  - M6: 3-5 test files
- **LOC delta**: ~1,000 lines net (M2a's git mv operations are content-preserving so they contribute minimal LOC delta; ~50 LOC of cross-reference path substitutions vs the v0.1.0 estimate of ~1,400)
- **Test LOC ratio**: ~30% (matching project test coverage targets)
- **Tier classification**: M (within 300-1500 LOC envelope; flat layout simplification REDUCES scope vs the v0.2.0 estimate that included subdirectory rebuilding)

## §G — Anti-Patterns to Avoid

| Anti-pattern | Why to avoid | Mitigation |
|--------------|--------------|------------|
| Eager file-level enumeration of every v2 path | Brittle — misses unknown zombies | Use namespace-prefix-based REMOVE (e.g., `.claude/skills/moai-*` glob), not individual file enumeration |
| Mutating files in-place during REMOVE | Risks partial-failure state with no recovery | Backup first (HARD-5), then remove; never edit in-place |
| Hard-coded user-modified config detection | Fragile across template versions | Use SHA-256 hash diff per REQ-VVCR-007 — version-agnostic |
| Blocking on telemetry emission | User experience degradation if telemetry endpoint is unreachable | Wrap in `defer recover()`; telemetry failure never blocks update |
| Implementing a duplicate `moai migrate agency` command | Violates SPEC §D.6 exclusion + user confusion | Reuse `runMigrateAgency` via auto-invocation (REQ-VVCR-025), not a separate command |
| Reinstalling design assets after removal | Violates REQ-VVCR-019 unwanted-behavior | Verify post-condition: design asset paths absent after MERGE-back (AC-VVCR-007) |

## §H — Cross-References

| Reference | Purpose |
|-----------|---------|
| `spec.md` | Canonical SSOT for §A motivation, §B requirements, §C constraints, §D exclusions |
| `acceptance.md` | AC-VVCR enumeration with REQ↔AC mapping |
| `design.md` | Architecture diagrams + algorithm pseudocode for v2 detection / PRESERVE inventory / REMOVE / reinstall / MERGE-back |
| `research.md` | Codebase research summary covering existing infrastructure reuse patterns |
| `.claude/rules/moai/workflow/spec-workflow.md` § Tier M lifecycle | Plan-phase requirements + 4-phase close lifecycle |
| `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Transition Ownership Matrix | Status transition rules for draft → planned → in-progress → implemented |
| `.claude/rules/moai/development/spec-gears-format.md` | GEARS notation canonical reference |
| `.claude/rules/moai/core/moai-constitution.md` | TRUST 5 framework + Agent Core Behaviors |
| `CLAUDE.local.md` §24 Harness Namespace Separation Contract | PRESERVE vs REMOVE matrix authority |
| `internal/defs/dirs.go` | `DeprecatedPathEntry` struct + `DeprecatedPaths` slice extension point |
| `internal/cli/migrate_agency.go` | `runMigrateAgency` reusable implementation |
| `internal/cli/update_cleanup.go` | `scanDeprecatedPaths` + `backupDeprecatedPaths` reusable functions |
| `internal/cli/update_namespace_protect.go` | `isUserOwnedNamespace` + `backupUserOwnedNamespace` reusable functions |
| `internal/cli/update.go:1731` `migrateLegacyMemoryDir` | Precedent pattern for auto-invoke migration |
| `SPEC-AGENCY-ABSORB-001` | Predecessor agency absorption SPEC |
| `SPEC-V3R6-AGENT-TEAM-REBUILD-001` | Source of 12-agent archive list |
| `SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001` | `.claude/agents/local/` user-owned namespace authority |
| `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` | `NamespaceBackupsSubdir` constant + namespace-protect contract |

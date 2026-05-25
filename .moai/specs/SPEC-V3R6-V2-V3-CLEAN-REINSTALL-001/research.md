---
id: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
title: "Research — Infrastructure reuse + v.2.x baseline archaeology for v2-to-v3 clean reinstall"
version: "0.5.1"
status: implemented
created: 2026-05-25
updated: 2026-05-26
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "internal/cli, internal/defs, pkg/version, internal/template/templates/.claude/agents, .claude/agents"
lifecycle: spec-anchored
tags: "moai-update, codebase-research, infrastructure-reuse, novel-code-estimate, v2-baseline-archaeology, layout-restoration"
tier: M
---

# Research — SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001

## §A — Codebase Inventory

### §A.1 Existing Infrastructure (Reuse Targets)

The orchestrator-supplied research-phase audit identified four primary existing files that the clean-reinstall code path will reuse. Each has been inspected during the plan-phase pre-flight ground truth check; line counts and timestamps confirmed.

| File | Size | Status | Reuse mode |
|------|------|--------|------------|
| `internal/cli/migrate_agency.go` | 21,575 bytes (~580 LOC) | Stable, last modified 2026-05-11 | Verbatim function reuse via `runMigrateAgency` |
| `internal/cli/update_cleanup.go` | 16,204 bytes (~430 LOC) | Stable, last modified 2026-05-18 | Function reuse: `scanDeprecatedPaths`, `backupDeprecatedPaths`, classification logic |
| `internal/cli/update_namespace_protect.go` | 9,619 bytes (~260 LOC) | Stable, last modified 2026-05-23 (SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 latest commit) | Function reuse: `isUserOwnedNamespace`, `newNamespaceBackupStamp`, `resolveNamespaceBackupDir`, `collectUserOwnedFiles`, `backupUserOwnedNamespace`, `assertNoUserOwnedNamespaceTouch` |
| `internal/cli/update.go` | 100,700 bytes (~2,700 LOC) | Active, last modified 2026-05-23 | Insertion point at `runUpdate` (line 126) + `migrateLegacyMemoryDir` pattern (line 1731) |

### §A.2 Existing `DeprecatedPaths` Table

Source: `internal/defs/dirs.go:41-96`

Current entries (9 total):
1. `.claude/commands/agency/agency.md`
2. `.claude/commands/agency/brief.md`
3. `.claude/commands/agency/build.md`
4. `.claude/commands/agency/evolve.md`
5. `.claude/commands/agency/learn.md`
6. `.claude/commands/agency/profile.md`
7. `.claude/commands/agency/resume.md`
8. `.claude/commands/agency/review.md`
9. `.claude/rules/agency/constitution.md`

All 9 entries carry:
- `DeprecatedSince: "SPEC-AGENCY-ABSORB-001"`
- `DeprecatedBy: "SPEC-V3R3-UPDATE-CLEANUP-001"`
- `RemovalSchedule: "v3.0.0"`

The extension by this SPEC preserves these 9 entries verbatim and appends 34 NEW entries (31 v.2.x-era per spec.md §A.4 Category B + 3 rc1-stage per spec.md §A.4 Category C), for a final total of 43 entries.

### §A.3 Confirmed `_TBD_` Markers

Source: `internal/template/templates/.moai/config/sections/`

Verified via `grep -l "_TBD_"`:
- `db.yaml` lines 35, 38, 44 — `engine: "_TBD_"`, `orm: "_TBD_"`, `migration_tool: "_TBD_"`
- `design.yaml` line 39 — comment referencing `_TBD_` markers (the file itself triggers brand interview when `_TBD_` present in `.moai/project/brand/`)

These two files are explicitly targeted for removal per REQ-VVCR-012 (D3 user decision: complete removal).

### §A.4 Confirmed Zero-Go-Reference yaml Files

The orchestrator's pre-audit verified that `gate.yaml`, `github-actions.yaml`, and `memo.yaml` each have 0 references in `internal/` Go source. This was the basis for D2 user decision (auto-remove). The verification can be re-run any time via:

```bash
for f in gate github-actions memo; do
  echo "=== $f.yaml references in internal/ ===";
  grep -rn "sections/$f.yaml\|$f\.yaml" internal/ | wc -l;
done
```

Expected output: each yaml file returns 0.

### §A.5 Auto-Migration Precedent

Source: `internal/cli/update.go:1731-1742`

```go
if err := migrateLegacyMemoryDir(projectRoot, out); err != nil {
    // ...
}

// migrateLegacyMemoryDir handles the .moai/memory/ → .moai/state/ migration.
func migrateLegacyMemoryDir(projectRoot string, out io.Writer) error {
    // ...
}
```

This is the canonical pattern for "detect legacy state → auto-migrate before the rest of the update flow proceeds." This SPEC replicates the pattern verbatim for `runMigrateAgency` integration (REQ-VVCR-025).

## §B — Maintainer Machine vs External User Audit

### §B.1 Maintainer Machine State (this repo)

The orchestrator pre-audit confirmed:

- `.agency/` and `.agency.archived/` — ABSENT (already cleaned via prior SPEC-AGENCY-ABSORB-001 work)
- 4 agency agents (planner / designer / builder / evaluator) — ABSENT
- 12 archived retired agents — ALL ABSENT, stored at `.moai/backups/agent-archive-2026-05-25/` per SPEC-V3R6-AGENT-TEAM-REBUILD-001 M3 milestone
- `.moai/memory/` — ABSENT (auto-migrated by existing `migrateLegacyMemoryDir`)

**Implication**: The maintainer machine is effectively already in a "v3 clean" state for these legacy assets. Testing of the clean-reinstall code path on the maintainer machine requires synthetic fixture projects in `t.TempDir()` that simulate v2 state.

### §B.2 External v2.x User Expectations

External users running v2.x may have all of the audited paths PLUS:
- Design-domain assets (per D1 user decision — complete removal)
- The 5 deprecated config yaml files (db / design / gate / github-actions / memo)
- `_TBD_` placeholder content in `.moai/project/brand/`
- Any of the 16 archived agents still present at their pre-archive paths

This is the population the clean-reinstall code path is designed to handle.

### §B.3 Risk: Unknown Zombies

The enumeration-based approach inherent to `DeprecatedPaths` cannot handle paths that were never explicitly listed. The clean-reinstall code path mitigates this via **namespace-prefix REMOVE**: all `.claude/skills/moai-*` directories are removed atomically, regardless of whether each specific name is enumerated. This catches unknown zombies introduced in v2.x intermediate releases that were never explicitly listed.

## §C — Reuse Pattern Catalog

### §C.1 Direct Function Reuse (no modification)

The following functions are called verbatim from the new clean-reinstall code path:

| Function | Origin | New caller |
|----------|--------|------------|
| `runMigrateAgency` | `internal/cli/migrate_agency.go` | `runCleanReinstall` Step 1 (conditional) |
| `scanDeprecatedPaths` | `internal/cli/update_cleanup.go` | `removePhase` |
| `backupDeprecatedPaths` | `internal/cli/update_cleanup.go` | (legacy backup; may be subsumed by `backupAll` Step 3 — implementation decision deferred to M4) |
| `newNamespaceBackupStamp` | `internal/cli/update_namespace_protect.go:61` | Step 3 backup stamp generation |
| `resolveNamespaceBackupDir` | `internal/cli/update_namespace_protect.go:74` | Step 3 backup directory resolution |
| `collectUserOwnedFiles` | `internal/cli/update_namespace_protect.go:100` | Step 2 PRESERVE inventory partial (extended) |
| `isUserOwnedNamespace` | `internal/cli/update_namespace_protect.go` | Step 2 enumeration filter |

### §C.2 Existing Constants Reuse

| Constant | Origin | Use |
|----------|--------|-----|
| `defs.NamespaceBackupsSubdir = "backups"` | `internal/defs/dirs.go:22` | Path component for `.moai/backups/v2-to-v3-{stamp}/` |
| `defs.MoAIDir = ".moai"` | `internal/defs/dirs.go:6` | Reused for path construction |
| `defs.ClaudeDir = ".claude"` | `internal/defs/dirs.go:9` | Reused for path construction |
| `defs.DeprecatedPaths` | `internal/defs/dirs.go:41` | Extended via append at M2 |
| `defs.AgentsMoaiSubdir`, `SkillsSubdir`, `CommandsMoaiSubdir`, etc. | `internal/defs/dirs.go:112-118` | Reused for v3 surface presence checks (AC-VVCR-010, AC-VVCR-012) |

### §C.3 Pattern Reuse (no direct code import)

| Pattern | Origin | Application |
|---------|--------|-------------|
| Auto-migration pre-step | `migrateLegacyMemoryDir` at `update.go:1731` | `runMigrateAgency` invocation at `runCleanReinstall` Step 1 |
| Backup-before-removal | `backupUserOwnedNamespace` (full namespace snapshot) | `backupAll` Step 3 verbatim approach |
| Transaction log + checkpoint | `migrate_agency.go` (full pattern) | Deferred — current SPEC uses simpler "single atomic 7-step + integrity verify" model; transaction log is out of scope unless EC-3 SIGTERM mid-REMOVE recovery is added as a future SPEC |
| Slash-separated path normalization | `DeprecatedPathEntry.Path` convention | All new path enumerations follow forward-slash storage; `filepath.FromSlash` at filesystem call time |

## §D — Novel Code Estimate

### §D.1 New Files (8)

| File | Purpose | Estimated LOC |
|------|---------|---------------|
| `internal/cli/v2_detection.go` | v2-fingerprint heuristic + signal struct | ~180 |
| `internal/cli/v2_detection_test.go` | 8-permutation signal coverage table-driven test | ~250 |
| `internal/cli/update_clean_install.go` | `runCleanReinstall` orchestration of 7-step pipeline | ~350 |
| `internal/cli/update_clean_install_test.go` | 3-scenario integration test (A/B/C) | ~280 |
| `internal/cli/update_preserve_inventory.go` | PRESERVE inventory builder + user-modified config detector | ~280 |
| `internal/cli/update_preserve_inventory_test.go` | PRESERVE inventory unit + hash diff tests | ~200 |
| Cross-platform CI YAML update | GitHub Actions matrix verification | ~50 |
| `CHANGELOG.md` entry | v3.0.0-rc2 release notes | ~60 |

Subtotal new files: ~1,650 LOC (within Tier M target range)

### §D.2 Modified Files (~6)

| File | Modification | Estimated LOC delta |
|------|--------------|---------------------|
| `pkg/version/version.go` | Version constant bump (line 8) | +1 -1 |
| `internal/defs/dirs.go` | `DeprecatedPaths` slice extension (34 NEW entries per spec.md §A.4) | +220 |
| `internal/defs/dirs_test.go` | New entry enumeration test | +60 |
| `internal/cli/update.go` | `runUpdate` v2-detection branch + `runMigrateAgency` integration | +60 |
| `internal/template/templates/.moai/config/config.yaml` | `moai.version` field update | +1 -1 |
| `.moai/config/sections/system.yaml` | `moai.version` field update (project-local) | +1 -1 |

Subtotal modified files: ~240 LOC delta

### §D.3 Total Estimated LOC Delta (v0.3.0 — FLAT layout simplification)

**~1,000 LOC net across ~38-42 files** (8 new + ~30 modified including M2a's ~24 git mv + path-substitution targets).

The v0.3.0 revision REDUCES the estimated LOC delta from v0.1.0's ~1,890 to ~1,000 because:

1. **M2a's git mv operations are content-preserving** — `git mv internal/template/templates/.claude/agents/core/manager-spec.md → moai/manager-spec.md` shows up in the LOC delta as 0 net lines (file content unchanged, only path renamed).
2. **Path-string substitutions are mechanical** — ~30 grep hits across skill/rule/SPEC bodies × an average of 1-2 lines per substitution = ~50 lines total for the FLAT-layout cross-reference updates.
3. **No subdirectory-rebuild overhead** — v0.2.0's intermediate estimate (~1,400 LOC) assumed `{core,meta}/` subdirectories under `moai/`; the FLAT simplification drops that complexity.

Breakdown:

| Milestone | New files | Modified files | Net LOC | Notes |
|-----------|----------:|----------------:|--------:|-------|
| M1 (version bump) | 0 | 4 | ~30 | version.go + CHANGELOG + 2 config mirrors |
| M2 (DeprecatedPaths) | 1 | 1 | ~170 | dirs.go + dirs_test.go (31 v.2.x + 3 rc1-stage entries = 34 NEW per spec.md §A.4) |
| M2a (FLAT layout) | 0 | ~24 | ~50 | 7 template + 7 local git mv (content-preserving) + ~30 cross-reference updates + 6 empty-dir removals + 1-2 predecessor SPEC frontmatter edits |
| M3 (v2 detection) | 2 | 0 | ~430 | v2_detection.go + test |
| M4 (clean reinstall) | 4 | 0 | ~830 | update_clean_install.go + update_preserve_inventory.go + 2 test files |
| M5 (runUpdate integration + catalog regen) | 0 | 2 | ~80 | update.go (~50 code) + catalog.yaml (~30 regenerated path/hash lines) |
| M6 (test coverage) | 3-5 | 0 | ~400 | scenario A/B/C integration tests + edge case tests |

**Tier classification**: M (within 300-1500 LOC envelope). The FLAT layout simplification keeps the SPEC firmly within Tier M; it would have required tier-up to L only if subdirectory rebuilding added significant new code, which it does not.

## §E — Risk Analysis

### §E.1 Identified Risks (mirrored from plan.md §B)

| Risk | Severity | Mitigation |
|------|----------|------------|
| User-modified config detection false-positive (whitespace-only changes) | Low | SHA-256 hash diff; documented in plan.md; false-positive causes unnecessary preservation but never data loss |
| Backup directory disk space (full `.moai/` + `.claude/` snapshot) | Medium | `--dry-run` reveals expected backup size; documented; SHOULD-1 explicit support |
| Cross-platform path normalization (windows backslash) | Medium | Reuse `filepath.FromSlash` convention; AC-VVCR-014 cross-platform CI verification |
| Telemetry event emission failure blocking update | Low | `defer recover()` wrapper; AC-VVCR-017 graceful degradation test |
| Concurrent `moai update` invocation race | Low | Reuse existing file lock; EC-1 edge case test |

### §E.2 Risks Out of This SPEC's Scope

| Risk | Why deferred |
|------|--------------|
| SIGTERM mid-REMOVE recovery without backup directory | Backup directory always exists by step 4 (HARD-5); manual recovery from backup is acceptable for this SPEC |
| Automated config conflict resolution | §D.3 OOS — manual reconciliation required |
| Backup directory automatic cleanup after successful update | Future SPEC — current design assumes user manages disk space |
| Telemetry backend infrastructure | §D.4 OOS — handled by existing telemetry subsystem |
| Forward compatibility to v3-to-v4 | §D.2 OOS — clean-reinstall pattern documented but no forward guarantee |

## §F — Predecessor SPEC Analysis

### §F.1 SPEC-AGENCY-ABSORB-001 (2026-04-23, commit `3e8b61e80`)

**What it did**: Migrated `.agency/` legacy data into `.moai/`; established the original 9-entry `DeprecatedPaths` table.

**What this SPEC reuses**: `runMigrateAgency` function verbatim for auto-invocation (REQ-VVCR-025).

**What this SPEC extends**: The 9-entry table grows to 43 entries (31 v.2.x-era NEW per spec.md §A.4 Category B); all new entries in this category carry `DeprecatedBy: "SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001"`.

### §F.2 SPEC-V3R3-UPDATE-CLEANUP-001

**What it did**: Owns ongoing maintenance of `DeprecatedPaths`; established the `scanDeprecatedPaths` / `backupDeprecatedPaths` API.

**What this SPEC reuses**: `scanDeprecatedPaths` + `backupDeprecatedPaths` functions verbatim.

**What this SPEC extends**: Adds 34 NEW entries authorised by this SPEC (per spec.md §A.4: 31 Category B + 3 Category C); both SPEC IDs co-listed in `DeprecatedBy` field for the Category B entries (this SPEC as the introducing authority; SPEC-V3R3-UPDATE-CLEANUP-001 as the ongoing maintenance authority).

### §F.3 SPEC-V3R6-AGENT-TEAM-REBUILD-001 (2026-05-25)

**What it did**: Generated the 12-agent archive at `.moai/backups/agent-archive-2026-05-25/` — 11 actual archived + 1 originally absent (`researcher.md`).

**What this SPEC reuses**: The archive itself (no code reuse). The 12 agent paths are the source of REMOVE targets for `.claude/agents/{core,expert,meta}/manager-* expert-* claude-code-guide researcher`.

### §F.4 SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001 (2026-05-25)

**What it did**: Established `.claude/agents/local/` user-owned namespace; defined PRESERVE behavior for `local/` directory.

**What this SPEC reuses**: PRESERVE semantics for `.claude/agents/local/` per REQ-VVCR-005 enumeration + HARD-3 namespace separation.

### §F.5 SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 (in-progress per `internal/defs/dirs.go:20-22` comment)

**What it does**: Owns the `NamespaceBackupsSubdir = "backups"` constant + the broader namespace-protect contract (`isUserOwnedNamespace`, etc.).

**What this SPEC reuses**: The constant verbatim for `.moai/backups/v2-to-v3-{stamp}/` path; the contract API for PRESERVE inventory construction.

**Coordination concern**: If SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001 is still in-progress when this SPEC begins implementation, M4 (which depends on `collectUserOwnedFiles`) MUST wait for that SPEC's completion. This is documented as a soft dependency in `depends_on` frontmatter.

### §F.6 SPEC-V3R6-AGENT-FOLDER-SPLIT-001 (rc1-stage, superseded by this SPEC — added v0.2.0)

**What it did**: Migrated `.claude/agents/moai/<name>.md` → `.claude/agents/{core,expert,meta}/<name>.md` (38-file `git mv` + 22 cross-reference updates + 19 catalog.yaml path entries). Plan PR `58a235e06`, feat commit `1bd083725` (2026-05-22). Status was `implemented` until this SPEC supersedes it.

**v.2.x baseline archaeology (orchestrator verification, 2026-05-25)**:

```bash
# Pre-split baseline (= v.2.x layout):
$ git ls-tree -r 1bd083725^ --name-only | grep '\.claude/agents/moai/'
.claude/agents/moai/builder-harness.md
.claude/agents/moai/claude-code-guide.md
.claude/agents/moai/evaluator-active.md
.claude/agents/moai/expert-backend.md
.claude/agents/moai/expert-devops.md
.claude/agents/moai/expert-frontend.md
.claude/agents/moai/expert-performance.md
.claude/agents/moai/expert-refactoring.md
.claude/agents/moai/expert-security.md
.claude/agents/moai/manager-brain.md
.claude/agents/moai/manager-develop.md
.claude/agents/moai/manager-docs.md
.claude/agents/moai/manager-git.md
.claude/agents/moai/manager-project.md
.claude/agents/moai/manager-quality.md
.claude/agents/moai/manager-spec.md
.claude/agents/moai/manager-strategy.md
.claude/agents/moai/plan-auditor.md
.claude/agents/moai/researcher.md

# Total: 19 files, FLAT directory (no subdirectories).
```

The `{moai => core}` rename diff in commit `1bd083725` (e.g., `.claude/agents/{moai => core}/manager-brain.md`) confirms the FLAT-to-split direction at commit time. This SPEC reverses that direction with respect to the 7 retained agents (the other 12 are archived to `.moai/backups/agent-archive-2026-05-25/` per SPEC-V3R6-AGENT-TEAM-REBUILD-001 M3 and do NOT participate in the layout restoration).

**Why this SPEC supersedes it**:

1. v3 has NOT shipped yet — the rc1-stage subdirectory split is in-flight, not user-visible
2. `.claude/agents/` is the ONLY `.claude/<namespace>/` directory that diverged from the v.2.x `moai/` parent convention; all other namespaces (skills, rules, commands, hooks, output-styles) preserved the flat `moai/` baseline
3. Only 7 retained agents exist post-archive; subdirectory grouping adds complexity without benefit
4. Restoring the FLAT layout simplifies the v2→v3 PRESERVE/REMOVE matrix to a single uniform shape (`.claude/<cat>/moai/` is template-managed; siblings are user-owned)

**What this SPEC reuses from SPEC-V3R6-AGENT-FOLDER-SPLIT-001**:

- The 22 cross-reference list (skill bodies, rule bodies, agent frontmatter) — applied IN REVERSE direction (from `{core,meta}/` back to `moai/`)
- The git mv mechanical pattern from commit `1bd083725` — applied IN REVERSE
- The agent_lint.go + update.go consumer code: NO change required because `defs.AgentsMoaiSubdir = "agents/moai"` already matched the FLAT shape (the split was a directory layout deviation only; the consuming Go code never adopted `core/expert/meta` constants)

**Supersedence markers (REQ-VVCR-LR-002 deliverables)**:

- `.moai/specs/SPEC-V3R6-AGENT-FOLDER-SPLIT-001/spec.md` frontmatter: `status: superseded` + `superseded_by: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`
- This SPEC's frontmatter: `supersedes: [SPEC-V3R6-AGENT-FOLDER-SPLIT-001]`
- HISTORY rows on both SPECs recording the supersedence event

## §G — Open Questions (Deferred to Run Phase)

The following questions are intentionally deferred to the run-phase implementation; they do not block plan-phase finalization but will require resolution before AC validation.

| Question | Deferred to | Anticipated resolution |
|----------|-------------|----------------------|
| Should `backupDeprecatedPaths` be subsumed by `backupAll` (Step 3) for simplicity, or kept as a redundant safety layer? | M4 implementation | Subsume — single backup point is simpler; `backupAll` snapshots both `.moai/` and `.claude/` |
| Should the dry-run flag invoke `runMigrateAgency` in dry-run mode (it has its own dry-run support), or skip migration entirely in dry-run? | M5 integration | Invoke `runMigrateAgency` in dry-run mode for accurate preview |
| Should the integrity verification (Step 7) violate-on-first-error or collect-all-violations? | M4 implementation | Collect-all — better diagnostic UX, no early exit |
| Should the user-modified config conflict resolution write `*.template-baseline` files or also write `*.diff` files showing the user's modifications? | M4 implementation | `.template-baseline` only for this SPEC; `.diff` deferred to future enhancement |

## §H — Cross-References

| Reference | Purpose |
|-----------|---------|
| `spec.md` | Canonical SSOT for §A motivation, §B requirements, §C constraints, §D exclusions |
| `plan.md` | M1-M6 milestone decomposition |
| `acceptance.md` | AC-VVCR enumeration |
| `design.md` | Algorithm pseudocode + sequence diagrams |
| `internal/cli/migrate_agency.go` | Reuse target — `runMigrateAgency` |
| `internal/cli/update_cleanup.go` | Reuse target — `scanDeprecatedPaths`, `backupDeprecatedPaths` |
| `internal/cli/update_namespace_protect.go` | Reuse target — `collectUserOwnedFiles`, `newNamespaceBackupStamp`, `resolveNamespaceBackupDir` |
| `internal/cli/update.go:1731` | Auto-migration precedent — `migrateLegacyMemoryDir` |
| `internal/defs/dirs.go:41` | Extension point — `DeprecatedPaths` slice |
| `pkg/version/version.go:8` | Version bump target |
| `CLAUDE.local.md` §24 | Namespace separation contract |
| `SPEC-AGENCY-ABSORB-001` | Predecessor agency SPEC |
| `SPEC-V3R3-UPDATE-CLEANUP-001` | DeprecatedPaths ongoing maintenance authority |
| `SPEC-V3R6-AGENT-TEAM-REBUILD-001` | 12-agent archive source |
| `SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001` | `.claude/agents/local/` namespace authority |
| `SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001` | `NamespaceBackupsSubdir` constant authority + namespace-protect API |

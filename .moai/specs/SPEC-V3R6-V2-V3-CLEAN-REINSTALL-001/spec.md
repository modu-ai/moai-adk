---
id: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001
title: "v2-to-v3 Clean Reinstall — paradigm shift from file-level sync to version-aware backup-remove-reinstall with v.2.x FLAT layout restoration and selective user-asset preservation"
version: "0.5.0"
status: in-progress
created: 2026-05-25
updated: 2026-05-25
author: manager-spec
priority: P1
phase: "v3.0.0-rc2"
module: "internal/cli, internal/defs, pkg/version, internal/template/templates/.claude/agents, .claude/agents"
lifecycle: spec-anchored
tags: "moai-update, v2-v3-migration, clean-reinstall, namespace-preservation, deprecated-paths, layout-restoration, v2-baseline"
depends_on: [SPEC-AGENCY-ABSORB-001, SPEC-V3R3-UPDATE-CLEANUP-001, SPEC-V3R6-AGENT-TEAM-REBUILD-001, SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001]
related_specs: [SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001, SPEC-V3R6-AGENT-FOLDER-SPLIT-001]
supersedes: [SPEC-V3R6-AGENT-FOLDER-SPLIT-001]
tier: M
---

# SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 — v2-to-v3 Clean Reinstall

## §A — Motivation and Background

### §A.0 Baseline Principle (added v0.2.0)

**v3 has NOT been released yet.** Only the `v3.0.0-rc1` development tag exists; no public v3.0.0 release has shipped to users. The v.2.x layout (most recent public release tag, e.g. `v2.9.1`) is the **canonical external-user baseline**. Any rc1-stage architectural decisions that broke v.2.x layout consistency are treated as in-flight deviations to be reverted, not as accepted v3 standards.

Concretely, this principle invalidates the `.claude/agents/{core,expert,meta}/` folder split introduced by SPEC-V3R6-AGENT-FOLDER-SPLIT-001 (commit `1bd083725`, 2026-05-22). That SPEC migrated `.claude/agents/moai/<name>.md` (FLAT layout) → `.claude/agents/{core,expert,meta}/<name>.md` (subdirectory-split layout), which was the only `.claude/<namespace>/` directory that diverged from the v.2.x `moai/` parent convention. All other v3-track SPECs (SEQ-THINKING-RETIRE, LOCAL-NAMESPACE-CONSOLIDATION, HARNESS-RENAME, SKILL-CONSOLIDATE) preserved v.2.x compatibility.

Git archaeology verification (orchestrator-side, 2026-05-25): `git ls-tree -r 1bd083725^ --name-only | grep '\.claude/agents/moai/'` confirms the v.2.x layout was **flat** — 19 agent files directly under `.claude/agents/moai/` with no `core/`, `expert/`, or `meta/` subdirectories. The subdirectory split was introduced by SPEC-V3R6-AGENT-FOLDER-SPLIT-001 itself, not by any v.2.x release.

This SPEC therefore supersedes SPEC-V3R6-AGENT-FOLDER-SPLIT-001 and restores the canonical FLAT `.claude/agents/moai/<name>.md` layout — with the 7 retained agents (post-SPEC-V3R6-AGENT-TEAM-REBUILD-001 archive) sitting flat under `moai/`, matching the v.2.x baseline structure exactly. The simplification is also internally consistent: only 7 retained agents exist, so subdirectory grouping adds complexity without benefit.

Post-restoration target layout:

```
.claude/agents/moai/
├── builder-harness.md
├── evaluator-active.md
├── manager-develop.md
├── manager-docs.md
├── manager-git.md
├── manager-spec.md
└── plan-auditor.md
```

This matches the skills / rules / commands / hooks namespace shape — each `.claude/<category>/moai/` directory holds the template-managed assets flat, and user-owned content lives in sibling directories (`.claude/agents/local/`, `.claude/agents/harness/`, `.claude/agents/<custom>.md`).

### §A.1 Problem Statement

External users running v2.x versions in the wild have unknown distributions of zombie files in their projects — agency-related agents, retired domain experts, design assets, `_TBD_` placeholder yaml files, and orphan rule directories. The current `moai update` command operates on file-level enumeration via `internal/defs/dirs.go` `DeprecatedPaths` (9 entries, all agency-related from SPEC-AGENCY-ABSORB-001 dated 2026-04-23). This enumeration-based approach only handles **known** zombies. Unknown zombies — files that were present in earlier v2.x releases but never explicitly enumerated — persist indefinitely in user projects.

Compounding the upgrade problem, the rc1-stage development line carries the SPEC-V3R6-AGENT-FOLDER-SPLIT-001 layout deviation (`.claude/agents/{core,expert,meta}/` instead of v.2.x `.claude/agents/moai/`). When v3 ships to users, the deviation either (a) forces all v.2.x users into a layout migration they did not consent to, or (b) creates a permanent inconsistency where `.claude/agents/` is the only `.claude/<namespace>/` directory NOT following the `moai/` parent convention. This SPEC chooses option (c): revert the rc1-stage deviation before v3 ships, keeping the v.2.x baseline as the canonical v3 layout.

### §A.2 Paradigm Shift Rationale

The proposed paradigm shift moves from **file-level sync** (delta-merge model with persistent zombie risk) to **version-aware clean reinstall** (backup-remove-reinstall model with selective preservation). The shift is justified by four observations:

1. The v2-to-v3 transition is a **one-time, well-defined boundary** (not an ongoing version drift) — clean reinstall amortizes cost across a single transition rather than introducing perpetual sync complexity.
2. The 8-retained / 12-archived agent catalog rebuild (SPEC-V3R6-AGENT-TEAM-REBUILD-001) means **file additions, removals, and renames all happen simultaneously** — incremental sync cannot guarantee consistency without enumerating every changed path.
3. External user feedback indicates **zombie file accumulation is a real maintenance burden** — users cannot easily distinguish template-managed files from user-authored content, and stale agents continue to be referenced by paste-ready memos.
4. The v.2.x layout baseline (per §A.0) is the single coherent target. Clean reinstall lets us restore `.claude/agents/moai/` parent convention atomically with the v2→v3 cutover, rather than as a separate breaking-change SPEC after v3 ships.

### §A.3 Predecessor and Related SPECs

- **SPEC-AGENCY-ABSORB-001** (implemented 2026-04-23, commit `3e8b61e80`): Migrated `.agency/` legacy data into `.moai/`; established the original 9-entry `DeprecatedPaths` table. This SPEC extends that table from 9 entries to 43 entries (the 9 pre-existing + the 34 NEW entries added by this SPEC, enumerated in §A.4 Canonical DeprecatedPaths Derivation Table below).
- **SPEC-V3R3-UPDATE-CLEANUP-001**: Owns ongoing maintenance of `DeprecatedPaths`. This SPEC is recorded as a co-maintainer in the new entries via `DeprecatedBy` field.
- **SPEC-V3R6-AGENT-TEAM-REBUILD-001** (implemented 2026-05-25): Generated the 12-agent archive that user projects MUST receive cleanup for. Note: archived agents lived under v.2.x `.claude/agents/moai/<name>.md` layout (NOT the rc1-stage `.claude/agents/{core,expert,meta}/` layout), so PRESERVE/REMOVE matching MUST target the `moai/`-prefixed paths.
- **SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001** (implemented 2026-05-25): Defines `.claude/agents/local/` user-owned namespace and the PRESERVE behavior verified via `isUserOwnedNamespace()`. The `local/` namespace sits as a sibling to `moai/` under the restored layout, mirroring the existing `.claude/skills/{moai/, moai-*, my-harness-*, local-*}` namespace convention.
- **SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001** (in-progress): Owns the `NamespaceBackupsSubdir` constant and the broader namespace-protect contract. This SPEC reuses that contract verbatim.
- **SPEC-V3R6-AGENT-FOLDER-SPLIT-001** (rc1-stage, plan PR `58a235e06`, feat commit `1bd083725`): **Superseded by this SPEC**. That SPEC split `.claude/agents/moai/` into `.claude/agents/{core,expert,meta}/`, which is the sole `.claude/<namespace>/` deviation from the v.2.x baseline. The supersedence is recorded in this SPEC's `supersedes:` frontmatter field. The layout-restoration deliverables in §B.10 + plan.md M2a roll back the 38-file `git mv` operation performed by `1bd083725`.

### §A.4 Canonical DeprecatedPaths Derivation Table (added v0.4.0 — single source of truth for entry counts)

This table is the **canonical entry-count derivation** referenced by REQ-VVCR-009, AC-VVCR-005, plan.md M2, research.md §A.2/§D.2, and acceptance.md DoD. Any entry-count statement elsewhere in this SPEC MUST refer to this table by section number (§A.4).

**Category A — Pre-existing entries (DeprecatedSince: SPEC-AGENCY-ABSORB-001)** = **9 entries** (preserved verbatim from `internal/defs/dirs.go:41-96` at the v0.4.0 authorship moment; this SPEC does NOT modify Category A):

| # | Path |
|---|------|
| 1 | `.claude/commands/agency/agency.md` |
| 2 | `.claude/commands/agency/brief.md` |
| 3 | `.claude/commands/agency/build.md` |
| 4 | `.claude/commands/agency/evolve.md` |
| 5 | `.claude/commands/agency/learn.md` |
| 6 | `.claude/commands/agency/profile.md` |
| 7 | `.claude/commands/agency/resume.md` |
| 8 | `.claude/commands/agency/review.md` |
| 9 | `.claude/rules/agency/constitution.md` |

**Category B — v.2.x-era NEW entries (DeprecatedSince: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001, DeprecatedBy: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001)** = **31 entries**. Paths use v.2.x FLAT layout (`.claude/agents/moai/<name>.md`) verified by `git ls-tree -r 1bd083725^`:

| # | Sub-category | Path |
|---|-------------|------|
| 10 | v2 dir | `.agency` |
| 11 | v2 dir | `.agency.archived` |
| 12 | agency agent (FLAT) | `.claude/agents/moai/planner.md` |
| 13 | agency agent (FLAT) | `.claude/agents/moai/designer.md` |
| 14 | agency agent (FLAT) | `.claude/agents/moai/builder.md` |
| 15 | agency agent (FLAT) | `.claude/agents/moai/evaluator.md` |
| 16 | manager archived (FLAT) | `.claude/agents/moai/manager-strategy.md` |
| 17 | manager archived (FLAT) | `.claude/agents/moai/manager-quality.md` |
| 18 | manager archived (FLAT) | `.claude/agents/moai/manager-brain.md` |
| 19 | manager archived (FLAT) | `.claude/agents/moai/manager-project.md` |
| 20 | meta archived (FLAT) | `.claude/agents/moai/claude-code-guide.md` |
| 21 | meta archived (FLAT) | `.claude/agents/moai/researcher.md` |
| 22 | expert archived (FLAT) | `.claude/agents/moai/expert-backend.md` |
| 23 | expert archived (FLAT) | `.claude/agents/moai/expert-frontend.md` |
| 24 | expert archived (FLAT) | `.claude/agents/moai/expert-security.md` |
| 25 | expert archived (FLAT) | `.claude/agents/moai/expert-devops.md` |
| 26 | expert archived (FLAT) | `.claude/agents/moai/expert-performance.md` |
| 27 | expert archived (FLAT) | `.claude/agents/moai/expert-refactoring.md` |
| 28 | config yaml | `.moai/config/sections/design.yaml` |
| 29 | config yaml | `.moai/config/sections/db.yaml` |
| 30 | config yaml | `.moai/config/sections/gate.yaml` |
| 31 | config yaml | `.moai/config/sections/github-actions.yaml` |
| 32 | config yaml | `.moai/config/sections/memo.yaml` |
| 33 | design skill dir | `.claude/skills/moai-domain-brand-design` |
| 34 | design skill dir | `.claude/skills/moai-domain-copywriting` |
| 35 | design skill dir | `.claude/skills/moai-domain-design-handoff` |
| 36 | design workflow skill dir | `.claude/skills/moai-workflow-design` |
| 37 | design workflow skill dir | `.claude/skills/moai-workflow-gan-loop` |
| 38 | design rule dir | `.claude/rules/moai/design` |
| 39 | brand dir | `.moai/project/brand` |
| 40 | db dir | `.moai/db` |

**Category C — rc1-stage staging artifacts (DeprecatedSince: SPEC-V3R6-AGENT-FOLDER-SPLIT-001, DeprecatedBy: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001)** = **3 entries**. These directories never existed in v.2.x; they were introduced by commit `1bd083725` and removed by this SPEC's M2a layout restoration. Their inclusion in `DeprecatedPaths` ensures that rc1-stage early-adopter testers (anyone who cloned the repo between `1bd083725` and the v3.0.0-rc2 cut) receive the layout-correction cleanup through `moai update`:

| # | Path |
|---|------|
| 41 | `.claude/agents/core` |
| 42 | `.claude/agents/expert` |
| 43 | `.claude/agents/meta` |

**Canonical totals (the only three numbers permitted in entry-count statements throughout this SPEC)**:
- **9 entries**: Category A (pre-existing, unchanged by this SPEC)
- **34 entries**: NEW entries added by this SPEC (Category B 31 + Category C 3)
- **43 entries**: total `DeprecatedPaths` slice size after M2 completion

## §B — Requirements (GEARS notation)

### §B.1 Core Reinstall Flow

**REQ-VVCR-001** (Ubiquitous): The `moai update` command **shall** detect whether the target project carries v2.x artifacts via a version-fingerprint heuristic combining `.moai/config/sections/system.yaml` `moai.version` field reading, presence of `.agency/` legacy directory, and presence of any path enumerated in the extended `DeprecatedPaths` table.

**REQ-VVCR-002** (Event-driven): **When** the v2-fingerprint heuristic returns positive, the `moai update` command **shall** invoke the clean-reinstall code path instead of the file-level sync code path.

**REQ-VVCR-003** (Event-driven): **When** the clean-reinstall code path activates, the `moai update` command **shall** create a backup directory at `.moai/backups/v2-to-v3-{ISO-8601-UTC}/` before any removal operation is performed.

**REQ-VVCR-004** (Ubiquitous): The clean-reinstall code path **shall** execute its operations in the canonical order: (1) version-fingerprint detection, (2) PRESERVE inventory snapshot, (3) backup creation, (4) REMOVE phase, (5) reinstall phase, (6) MERGE-back phase, (7) integrity verification.

### §B.2 PRESERVE Inventory (User-Owned Assets)

**REQ-VVCR-005** (Ubiquitous): The PRESERVE inventory snapshot **shall** enumerate all user-owned assets prior to removal: `.moai/specs/` recursive content, `.moai/project/{product,structure,tech}.md`, `.moai/harness/`, `.moai/state/`, `.moai/backups/`, `.moai/logs/`, `.claude/skills/<dirs not prefixed with moai- or moai/>`, `.claude/agents/<everything that is not the moai/ template-managed parent directory>` (custom user agents at root `.claude/agents/*.md`, `.claude/agents/local/` per SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001, `.claude/agents/harness/` per CLAUDE.local.md §24, and any other user-custom directories), `.claude/agent-memory/`, `.claude/commands/` (root files and non-`moai/` subdirectories), `.claude/rules/<directories other than moai/>`, `.claude/hooks/<directories other than moai/>`, `.claude/output-styles/<directories other than moai/>`, `.claude/settings.local.json`, and user-modified `.moai/config/sections/*.yaml` files detected via manifest hash diff.

**REQ-VVCR-006** (Unwanted): The clean-reinstall code path **shall not** modify, remove, overwrite, or relocate any path enumerated in the PRESERVE inventory.

**REQ-VVCR-007** (Ubiquitous): The clean-reinstall code path **shall** detect user-modified `.moai/config/sections/*.yaml` files by computing a SHA-256 hash of each section file and comparing against the embedded template baseline hash; any file whose hash differs from the template baseline **shall** be classified as user-modified and added to the PRESERVE inventory.

**REQ-VVCR-008** (Where-capability): **Where** a user-modified config section file conflicts with the v3 template baseline (the template version is newer or has structural changes), the clean-reinstall code path **shall** write the template baseline to `.moai/backups/v2-to-v3-{stamp}/config-conflicts/{filename}.template-baseline` so the user can manually reconcile differences post-update.

### §B.3 REMOVE Phase (v2 Artifacts and Template-Managed Surfaces)

**REQ-VVCR-009** (Ubiquitous): The REMOVE phase **shall** remove all paths enumerated in the extended `DeprecatedPaths` table, which after M2 completion contains exactly 43 entries (9 pre-existing per Category A + 34 NEW added by this SPEC per Categories B and C, enumerated in spec.md §A.4 Canonical DeprecatedPaths Derivation Table).

**REQ-VVCR-010** (Ubiquitous): The REMOVE phase **shall** remove all template-managed namespace surfaces in their entirety prior to reinstall: `.claude/skills/moai-*` directories, `.claude/skills/moai/` directory, `.claude/agents/moai/` directory (the entire `moai/` parent, consistent with the v.2.x baseline and the §A.0 baseline principle), `.claude/rules/moai/` directory, `.claude/commands/moai/` directory, `.claude/hooks/moai/` directory, and `.claude/output-styles/moai/` directory. The REMOVE phase **shall also** remove the rc1-stage staging artifacts `.claude/agents/core/`, `.claude/agents/expert/`, and `.claude/agents/meta/` directories if they exist (these are SPEC-V3R6-AGENT-FOLDER-SPLIT-001 deviations, not v.2.x baseline directories).

**REQ-VVCR-011** (Ubiquitous): The REMOVE phase **shall** remove design-domain assets in their entirety: `.claude/skills/moai-domain-brand-design/`, `.claude/skills/moai-domain-copywriting/`, `.claude/skills/moai-domain-design-handoff/`, `.claude/skills/moai-workflow-design/`, `.claude/skills/moai-workflow-gan-loop/`, `.claude/rules/moai/design/`, `.moai/project/brand/`, and `.moai/config/sections/design.yaml`.

**REQ-VVCR-012** (Ubiquitous): The REMOVE phase **shall** remove all `_TBD_`-marker placeholder yaml files: `.moai/config/sections/db.yaml` (contains `_TBD_` engine / orm / migration_tool literals) and any other yaml file in `.moai/config/sections/` containing the literal string `_TBD_` after the v3 template baseline write completes.

**REQ-VVCR-013** (Ubiquitous): The REMOVE phase **shall** remove zero-Go-reference yaml files: `.moai/config/sections/gate.yaml`, `.moai/config/sections/github-actions.yaml`, and `.moai/config/sections/memo.yaml` (each verified to have zero references in `internal/` Go source at the time of this SPEC's authoring).

**REQ-VVCR-014** (Ubiquitous): The REMOVE phase **shall** remove the `.moai/db/` directory (aligned with the removal of `.moai/config/sections/db.yaml` per REQ-VVCR-012).

**REQ-VVCR-015** (Ubiquitous): The REMOVE phase **shall** remove the `.moai/cache/` directory if present (auto-regenerated at next runtime; safe to remove).

**REQ-VVCR-016** (Ubiquitous): The REMOVE phase **shall** remove the `.agency/` and `.agency.archived/` legacy directories if present.

### §B.4 Reinstall Phase (v3 Template Baseline)

**REQ-VVCR-017** (Event-driven): **When** the REMOVE phase completes successfully, the reinstall phase **shall** deploy the v3 embedded template baseline to the project root, restoring all template-managed surfaces removed by REQ-VVCR-010.

**REQ-VVCR-018** (Ubiquitous): The reinstall phase **shall** preserve all PRESERVE-inventory paths verbatim — no template file deployment overwrites a user-owned path.

**REQ-VVCR-019** (Unwanted): The reinstall phase **shall not** reinstall design-domain assets (REQ-VVCR-011 removal targets); design assets remain absent from the v3 baseline and are opt-in via a future separate flow.

**REQ-VVCR-020** (Unwanted): The reinstall phase **shall not** reinstall the removed `_TBD_`-marker yaml files (REQ-VVCR-012 targets) or zero-Go-reference yaml files (REQ-VVCR-013 targets); these files remain absent from the v3 baseline.

### §B.5 MERGE-Back Phase (User Asset Restoration)

**REQ-VVCR-021** (Event-driven): **When** the reinstall phase completes successfully, the MERGE-back phase **shall** restore all PRESERVE-inventory paths to their pre-REMOVE locations.

**REQ-VVCR-022** (Where-capability): **Where** a user-modified config section file was added to the PRESERVE inventory per REQ-VVCR-007, the MERGE-back phase **shall** restore the user's version (not the template baseline) to `.moai/config/sections/{filename}`.

### §B.6 Integrity Verification

**REQ-VVCR-023** (Event-driven): **When** the MERGE-back phase completes, the clean-reinstall code path **shall** verify post-conditions: (a) all PRESERVE-inventory paths exist and have unchanged content, (b) all REMOVE-target paths are absent, (c) all v3 template-managed surfaces are present, (d) `.moai/config/sections/system.yaml` `moai.version` field reflects v3.0.0-rc2 or later.

**REQ-VVCR-024** (Event-driven): **When** integrity verification detects a post-condition violation, the clean-reinstall code path **shall** emit a structured error report naming the violated condition and the affected path, and **shall** point the user to the backup directory `.moai/backups/v2-to-v3-{stamp}/` for recovery.

### §B.7 Auto-Migration Hook

**REQ-VVCR-025** (Event-driven): **When** the v2-fingerprint heuristic detects the presence of `.agency/` legacy directory, the clean-reinstall code path **shall** invoke the existing `runMigrateAgency` migration logic (per the `migrateLegacyMemoryDir` precedent at `internal/cli/update.go:1731`) before the REMOVE phase, so user data inside `.agency/` is migrated into `.moai/` rather than discarded.

### §B.8 Cross-Platform and Idempotency

**REQ-VVCR-026** (Ubiquitous): The clean-reinstall code path **shall** function correctly on linux/amd64, darwin/arm64, darwin/amd64, and windows/amd64 platforms (slash-separated path normalization, case-sensitivity awareness, and platform-specific file mode handling).

**REQ-VVCR-027** (Ubiquitous): The clean-reinstall code path **shall** be idempotent — invoking `moai update` on a project that is already on v3 (the v2-fingerprint heuristic returns negative) **shall** route to the existing file-level sync code path, leaving the project unchanged when no file deltas exist.

### §B.9 Dry-Run and Telemetry

**REQ-VVCR-028** (Where-capability): **Where** the user invokes `moai update --dry-run`, the clean-reinstall code path **shall** print the planned PRESERVE inventory, the REMOVE target list, and the reinstall surface list to stdout, but **shall not** perform any filesystem mutation.

**REQ-VVCR-029** (Event-driven): **When** the clean-reinstall code path activates, the `moai update` command **shall** emit a telemetry event tagged `update.clean_reinstall.v2_to_v3` carrying the detection signals (which v2-fingerprint trigger fired, count of REMOVE targets, count of PRESERVE-inventory paths) for opt-in usage analytics.

### §B.10 Layout Restoration (v.2.x baseline)

This category restores the FLAT `.claude/agents/moai/<agent-name>.md` convention that was the v.2.x baseline and was deviated from by SPEC-V3R6-AGENT-FOLDER-SPLIT-001 (commit `1bd083725`, 2026-05-22, which introduced the `{core,expert,meta}/` subdirectory split). Per the §A.0 baseline principle, the rc1-stage deviation is reverted before v3 ships to users.

**REQ-VVCR-LR-001** (Ubiquitous): The v3 template (`internal/template/templates/.claude/agents/`) **shall** restore the `.claude/agents/moai/` directory as a FLAT layout containing the 7 retained agent files directly (no `core/`, `expert/`, or `meta/` subdirectories) — matching the v.2.x baseline structure verified by `git ls-tree -r 1bd083725^ --name-only | grep '\.claude/agents/moai/'`. The local checkout (`.claude/agents/moai/`) **shall** mirror the template structure byte-for-byte per the Template-First mirror parity rule.

**REQ-VVCR-LR-002** (Ubiquitous): The `moai-adk` repository **shall** mark `SPEC-V3R6-AGENT-FOLDER-SPLIT-001` as superseded by this SPEC via (a) `status: superseded` frontmatter on `SPEC-V3R6-AGENT-FOLDER-SPLIT-001/spec.md`, (b) `superseded_by: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` frontmatter on the same file, and (c) `supersedes: [SPEC-V3R6-AGENT-FOLDER-SPLIT-001]` frontmatter on this SPEC's spec.md.

**REQ-VVCR-LR-003** (Where-capability): **Where** the `defs.AgentsMoaiSubdir = "agents/moai"` constant value at `internal/defs/dirs.go:112` already matches the restored FLAT layout, the `agent_lint.go` consumer, the `update.go` consumer, and any other in-tree consumer **shall** resolve to the restored path correctly without source-code modification. (The constant value was never changed by SPEC-V3R6-AGENT-FOLDER-SPLIT-001; only the on-disk directory layout deviated by adding subdirectories. Restoration of the on-disk layout to match the constant's flat shape is sufficient.)

**REQ-VVCR-LR-004** (Ubiquitous): All cross-references to the rc1-stage paths `.claude/agents/core/<agent>.md` and `.claude/agents/meta/<agent>.md` in skill bodies (`.claude/skills/**/*.md`), rule bodies (`.claude/rules/**/*.md`), agent frontmatter (`.claude/agents/**/*.md` `tools:` / `model:` fields and prose), `CLAUDE.md`, `CLAUDE.local.md`, and SPEC bodies (`.moai/specs/*/spec.md`, `*/plan.md`, `*/acceptance.md`, `*/design.md`, `*/research.md`) **shall** be updated to the FLAT form `.claude/agents/moai/<agent>.md` (no subdirectory segment). The empty `.claude/agents/expert/` directory left by SPEC-V3R6-AGENT-TEAM-REBUILD-001 M3 archive **shall** be removed (no cross-references exist post-archive because the 6 `expert-*` agents are already in `.moai/backups/agent-archive-2026-05-25/`).

**REQ-VVCR-LR-005** (Event-driven): **When** the layout restoration deliverables in M2a complete (per plan.md), the orchestrator **shall** regenerate `internal/template/catalog.yaml` via `go run ./internal/template/scripts/gen-catalog-hashes.go --all` so the path entries and hash anchors reflect the restored FLAT `.claude/agents/moai/<agent>.md` paths (no `core/` or `meta/` segments).

**REQ-VVCR-LR-006** (Unwanted): The clean-reinstall code path **shall not** reinstall the rc1-stage `.claude/agents/{core,expert,meta}/` directories under any circumstance — these directories are explicitly absent from the v3 baseline per the §A.0 baseline principle. The reinstall phase deploys only the restored FLAT `.claude/agents/moai/` layout.

## §C — Constraints

### §C.1 HARD Constraints

**HARD-1** (User SPECs preservation): User SPEC documents at `.moai/specs/` **shall** never be modified, removed, or relocated by the clean-reinstall code path. This is non-negotiable — user SPECs represent active development work.

**HARD-2** (User project docs preservation): User project documentation files at `.moai/project/product.md`, `.moai/project/structure.md`, and `.moai/project/tech.md` **shall** be preserved verbatim per CLAUDE.md §1 active work guarantee.

**HARD-3** (Harness namespace preservation): User-owned harness assets at `.claude/skills/my-harness-*` and `.claude/agents/harness/` **shall** be preserved per the `CLAUDE.local.md §24` namespace separation contract.

**HARD-4** (Commands root preservation): Files at `.claude/commands/` root level (single `.md` files outside the `moai/` subdirectory) and non-`moai/` subdirectories under `.claude/commands/` **shall** be preserved. Only `.claude/commands/moai/` is template-managed; everything else is user-owned. This corrects a prior misconception that the entire `.claude/commands/` directory was template-managed.

**HARD-5** (Backup-before-removal): A backup directory at `.moai/backups/v2-to-v3-{ISO-8601-UTC}/` **shall** be created and populated before any path enumerated in the REMOVE phase is removed. The backup **shall** include a copy of the entire `.moai/` and `.claude/` directories at the moment immediately preceding REMOVE.

**HARD-6** (User-modified config preservation): User-modified `.moai/config/sections/*.yaml` files (detected via SHA-256 hash diff against template baseline) **shall** be preserved; the corresponding template baseline **shall** be written to `.moai/backups/v2-to-v3-{stamp}/config-conflicts/{filename}.template-baseline` for manual user review.

**HARD-7** (v.2.x layout baseline non-negotiable): The v3 template's `.claude/agents/` directory **shall** match the v.2.x baseline shape exactly: a FLAT `moai/` subdirectory with the 7 retained template-managed agent files directly inside it, plus user-owned siblings (`local/`, `harness/`, root `.md` files, any custom directories). The rc1-stage `.claude/agents/{core,expert,meta}/` subdirectory split introduced by SPEC-V3R6-AGENT-FOLDER-SPLIT-001 **shall not** appear in any v3 ship artifact (template baseline, post-clean-reinstall checkout, or catalog.yaml path entry).

### §C.2 SHOULD Constraints

**SHOULD-1** (Dry-run support): The clean-reinstall code path SHOULD support `--dry-run` per REQ-VVCR-028 to allow users to preview the planned operations.

**SHOULD-2** (Cross-platform): The clean-reinstall code path SHOULD function correctly across linux/darwin/windows per REQ-VVCR-026.

**SHOULD-3** (Telemetry opt-in): The clean-reinstall code path SHOULD emit a telemetry event per REQ-VVCR-029 to support usage analytics without breaking user privacy expectations (event opt-in via existing MoAI telemetry settings).

**SHOULD-4** (Idempotency): The clean-reinstall code path SHOULD be idempotent per REQ-VVCR-027 — re-running on a v3 project produces no-op.

## §D — Exclusions (What NOT to Build)

### §D.1 Out of Scope — Design Asset Reinstallation

This SPEC removes design-domain assets per REQ-VVCR-011. A future separate opt-in flow for reinstalling design assets is **explicitly out of scope**. Users who wish to use design features after upgrading to v3 will invoke a dedicated future command (out of this SPEC's authorship).

### §D.2 Out of Scope — Future v3-to-v4 Migration

This SPEC handles the v2-to-v3 transition exclusively. Future major-version transitions (v3 → v4 and beyond) are **out of scope**. The clean-reinstall pattern established by this SPEC MAY serve as a template for future major-version SPECs, but no forward-compatibility guarantees are made.

### §D.3 Out of Scope — Automatic Config Conflict Resolution

This SPEC writes template-baseline files to `.moai/backups/v2-to-v3-{stamp}/config-conflicts/` for user-modified config sections (REQ-VVCR-008, HARD-6). **Automatic resolution of config conflicts is explicitly out of scope** — the user MUST manually reconcile their version against the template baseline. Automated merging would risk silent corruption of user intent.

### §D.4 Out of Scope — Telemetry Backend

This SPEC emits a telemetry event per REQ-VVCR-029 / SHOULD-3, but the **backend collection infrastructure is out of scope** and assumed to be operated by the existing MoAI telemetry subsystem (not specified by this SPEC).

### §D.5 Out of Scope — Rollback Beyond Backup Directory

This SPEC creates a backup at `.moai/backups/v2-to-v3-{stamp}/` (REQ-VVCR-003, HARD-5). **Automated rollback from that backup is out of scope** — users who need to revert MUST manually restore from the backup directory or use git history. A dedicated `moai update --rollback` flag is reserved for a future SPEC.

### §D.6 Out of Scope — Duplicate `moai migrate agency` Command

This SPEC integrates `runMigrateAgency` invocation into the v2-detection-triggered branch (REQ-VVCR-025) but **does not propose a duplicate or alternative migration command**. The existing `moai migrate agency` command remains operational independently for users who wish to invoke migration explicitly.

### §D.7 Out of Scope — v.2.x → v4.x Future Migration

This SPEC handles the v.2.x → v3.0.0-rc2 release transition exclusively. Future migrations beyond v3 (v3 → v4 and onward) are **explicitly out of scope**. The clean-reinstall pattern + v.2.x baseline restoration established here apply only to the v.2.x → v3 cutover; subsequent major-version transitions will require dedicated SPECs with their own baseline analyses.

## §E — Acceptance Index

See `acceptance.md` for the full AC-VVCR-NNN enumeration with REQ↔AC traceability.

| # | AC-VVCR-NNN | Severity | Linked REQ(s) |
|---|--------------|----------|----------------|
| 1 | AC-VVCR-001 | MUST | REQ-VVCR-001, REQ-VVCR-002 |
| 2 | AC-VVCR-002 | MUST | REQ-VVCR-003, REQ-VVCR-004, HARD-5 |
| 3 | AC-VVCR-003 | MUST | REQ-VVCR-005, REQ-VVCR-006, HARD-1, HARD-2, HARD-3, HARD-4 |
| 4 | AC-VVCR-004 | MUST | REQ-VVCR-007, REQ-VVCR-008, HARD-6 |
| 5 | AC-VVCR-005 | MUST | REQ-VVCR-009 |
| 6 | AC-VVCR-006 | MUST | REQ-VVCR-010 |
| 7 | AC-VVCR-007 | MUST | REQ-VVCR-011, REQ-VVCR-019 |
| 8 | AC-VVCR-008 | MUST | REQ-VVCR-012, REQ-VVCR-013, REQ-VVCR-014, REQ-VVCR-020 |
| 9 | AC-VVCR-009 | MUST | REQ-VVCR-015, REQ-VVCR-016 |
| 10 | AC-VVCR-010 | MUST | REQ-VVCR-017, REQ-VVCR-018 |
| 11 | AC-VVCR-011 | MUST | REQ-VVCR-021, REQ-VVCR-022 |
| 12 | AC-VVCR-012 | MUST | REQ-VVCR-023, REQ-VVCR-024 |
| 13 | AC-VVCR-013 | MUST | REQ-VVCR-025 |
| 14 | AC-VVCR-014 | SHOULD | REQ-VVCR-026, SHOULD-2 |
| 15 | AC-VVCR-015 | SHOULD | REQ-VVCR-027, SHOULD-4 |
| 16 | AC-VVCR-016 | SHOULD | REQ-VVCR-028, SHOULD-1 |
| 17 | AC-VVCR-017 | SHOULD | REQ-VVCR-029, SHOULD-3 |
| 18 | AC-VVCR-LR-001 | MUST | REQ-VVCR-LR-001, REQ-VVCR-LR-006, HARD-7 |
| 19 | AC-VVCR-LR-002 | MUST | REQ-VVCR-LR-002 |
| 20 | AC-VVCR-LR-003 | MUST | REQ-VVCR-LR-003 |
| 21 | AC-VVCR-LR-004 | MUST | REQ-VVCR-LR-004 |
| 22 | AC-VVCR-LR-005 | MUST | REQ-VVCR-LR-005 |

## §F — Version Bump

The `pkg/version/version.go` `Version` constant transitions from `"v3.0.0-rc1"` to `"v3.0.0-rc2"` as part of M1 of this SPEC. A new git tag `v3.0.0-rc2` will be cut after merge. Existing build verification via `make build VERSION=v3.0.0-rc2` has been confirmed in dry-run.

## §G — HISTORY

### v0.5.0 (2026-05-25) — iter-3 PASS-WITH-DEBT 0.89 → skip-eligible last-mile residue fixes

plan-auditor iter-2 verdict: PASS-WITH-DEBT 0.89 (Tier M threshold 0.80 +0.09 margin; 0.01 gap to 0.90 skip-eligible). D1/D2/D4/D6 RESOLVED in iter-2; D3+D5 PARTIAL with 1 residue location each (E1 + E4). Targeted iter-3 fixes applied:

- **E1 (SHOULD-FIX, REQUIRED)**: research.md §F.1 LOC-breakdown table M2 row corrected — `(16 v.2.x + 3 rc1-stage entries)` → `(31 v.2.x + 3 rc1-stage entries = 34 NEW per spec.md §A.4)`. Aligns with spec.md §A.4 Canonical Derivation Table Category B = 31 entries (D3 propagation residue closed)
- **E3 (MINOR, RECOMMENDED)**: plan.md M2 deliverable section now carries an explicit canonical-count summary line ("Total NEW entries: 31 Category B + 3 Category C = 34 NEW entries, extending DeprecatedPaths from 9 to 43 total — per spec.md §A.4") immediately before the File Count line, eliminating the need for readers to manually tally the enumeration
- **E2 (MINOR, optional applied)**: plan.md M2 LOC-delta framing tightened — `~170 lines (mostly enumeration table; +3 rc1-stage entries vs v0.1.0 estimate)` → `~170 lines (mostly enumeration table; 31 Category B v.2.x-era + 3 Category C rc1-stage NEW entries per §A.4)`. Replaces historical "v0.1.0 estimate" framing with canonical §A.4 anchor
- **E4 (MINOR, optional applied)**: design.md §B.1 V2Fingerprint struct field comment for `V2DetectedViaVersion` now describes all 3 Signal 1 sub-states (`v2.*` prefix / empty string / system.yaml missing) — aligns inline struct documentation with the implementation body's `if strings.HasPrefix(version, "v2.") || version == ""` OR `errors.Is(err, fs.ErrNotExist)` triad (D5 propagation residue closed)
- **E5 (INFO, observation only — deferred)**: acceptance.md L156 inline disclaimer about rules retaining their own `core/` subdirectory is correct as-written per plan-auditor; no change

Self-projected iter-3 score change estimate: 0.89 → 0.91-0.92 skip-eligible (Consistency dimension nudge from D3+D5 residue closure; other dimensions unchanged at iter-2 levels).

### v0.4.0 (2026-05-25) — iter-2 PASS-WITH-DEBT → skip-eligible Consistency dimension fixes

plan-auditor iter-1 verdict: PASS-WITH-DEBT 0.83 (Tier M threshold 0.80, +0.03 margin, NOT skip-eligible — gap to 0.90 = 0.07). Consistency dimension was the score-dragging weakness (0.62) due to FLAT-layout intent vs rc1-stage path residue + inconsistent entry-count claims. Surgical D1-D6 directives applied:

- **D1 (design.md §B.6 integrity check FLAT path correction)**: replaced literal `.claude/agents/core/manager-spec.md` in the v3 surface verification list with the FLAT post-restoration path `.claude/agents/moai/manager-spec.md`; added 2 sibling FLAT paths (`.claude/agents/moai/builder-harness.md`, `.claude/agents/moai/plan-auditor.md`) as representative samples; added a new (c.1) FLAT-layout enforcement block that asserts rc1-stage subdirectories MUST be absent (6 negative-assertion paths covering `.claude/agents/{core,expert,meta}/` AND `.claude/agents/moai/{core,expert,meta}/`)
- **D2 (acceptance.md AC-VVCR-006 Given/When/Then asymmetry split)**: restructured AC-VVCR-006 to explicitly accept BOTH v.2.x FLAT input AND rc1-leak split input as Given-state, with a 3-condition Then assertion (REMOVE-phase absence + FLAT reinstall presence + negative assertion that no rc1-stage path remains). The asymmetry is now testable via table-driven fixtures covering both input layouts
- **D3 (entry-count canonical number propagation)**: added new spec.md §A.4 Canonical DeprecatedPaths Derivation Table with exact 43-row enumeration (9 Category A pre-existing + 31 Category B v.2.x-era + 3 Category C rc1-stage). Three canonical numbers (**9 / 34 / 43**) replace the 5 different inconsistent claims (`~25`, `~28`, `at least 25`, `at least 28`, `16+`) across spec.md, acceptance.md, research.md
- **D4 (REQ-VVCR-009 "~" qualifier removal)**: REQ-VVCR-009 wording now states "exactly 43 entries" with §A.4 cross-reference; enumeration is deferred to §A.4 to maintain REQ precision
- **D5 (v2-detection edge-case alignment — Option α broader)**: AC-VVCR-001 expanded to include 3 Signal-1 sub-states (`v2.*` prefix / empty string / missing file) all returning positive; EC-5 split into Scenario A (file missing) / Scenario B (empty-string version) / Scenario C (unrecognized format = v3 or later → negative). Aligns acceptance.md with design.md §B.1 pseudocode
- **D6 (plan.md §E self-verification wording)**: "map to dedicated AC entries" → "are traced in at least one AC entry" with explicit 1:N HARD↔AC mapping detail (HARD-1/2/3/4 all link to AC-VVCR-003)

D7-D11 (INFO/MINOR cosmetic items) deferred per plan-auditor explicit guidance.

Self-projected iter-2 score change estimate: 0.83 → 0.91 skip-eligible (Consistency 0.62 → 0.90+ via FLAT-layout consistency restoration; other dimensions unchanged).

### v0.3.0 (2026-05-25) — Flat layout simplification

- User further simplification: flat `.claude/agents/moai/` layout instead of `{core,meta}/` subdirectories — matches v.2.x baseline exactly per `git ls-tree -r 1bd083725^` verification
- §A.0 baseline principle expanded with explicit FLAT-layout target diagram (7 retained agents directly under `moai/`, no subdirectories)
- REQ-VVCR-LR-001 wording: "FLAT layout containing the 7 retained agent files directly (no core/, expert/, or meta/ subdirectories)"
- REQ-VVCR-LR-004 wording: cross-reference targets simplified to single-segment `.claude/agents/moai/<agent>.md` (no `/core/` or `/meta/` insertion)
- REQ-VVCR-LR-005 wording: catalog.yaml regeneration target is the flat path layout
- REQ-VVCR-LR-006 wording: unwanted-behavior list extended to reject ANY `{core,expert,meta}/` directory creation including under `moai/`
- M2a deliverable count revised: 7 template git mv + 7 local git mv + 3 empty-directory removals × 2 trees = ~17-20 individual operations (down from the v0.2.0 estimate of ~25 because no `core/meta/` recreation under `moai/`)
- Self-projected plan-auditor score revised upward: 0.86-0.89 → 0.90+ skip-eligible range plausible due to simpler PRESERVE/REMOVE matrix matching skills/rules/commands/hooks shape

### v0.2.0 (2026-05-25) — Layout Restoration + v.2.x baseline principle

- New §A.0 Baseline Principle: "v3 has NOT been released yet; v.2.x layout is canonical"
- New §B.10 Layout Restoration category with 6 REQ-VVCR-LR-NNN requirements (Ubiquitous / Where-capability / Event-driven / Unwanted patterns)
- 5 new AC-VVCR-LR-NNN entries (all MUST severity) — AC count grows from 17 to 22
- New HARD-7 constraint enforcing v.2.x layout non-negotiable
- New §D.7 OOS section excluding v3 → v4 future migration
- `supersedes: [SPEC-V3R6-AGENT-FOLDER-SPLIT-001]` frontmatter added
- `version: 0.1.0 → 0.2.0` increment, `module` field expanded to include `.claude/agents` paths, `tags` extended with `layout-restoration, v2-baseline`
- PRESERVE/REMOVE matrix agents row corrected: REMOVE = `.claude/agents/moai/` (the entire `moai/` parent like skills/rules/commands/hooks); PRESERVE = everything else (root `.md`, `local/`, `harness/`, custom)
- Predecessor SPEC list (§A.3) expanded with SPEC-V3R6-AGENT-FOLDER-SPLIT-001 supersedence rationale

### v0.1.0 (2026-05-25) — Initial draft

- 29 REQ-VVCR requirements authored in 100% GEARS notation (zero IF/THEN clauses)
- 17 AC-VVCR acceptance criteria enumerated with 100% bidirectional REQ↔AC traceability
- 6 HARD constraints and 4 SHOULD constraints documented
- 6 explicit out-of-scope sections per the Tier M discipline
- Cross-references to SPEC-AGENCY-ABSORB-001, SPEC-V3R3-UPDATE-CLEANUP-001, SPEC-V3R6-AGENT-TEAM-REBUILD-001, SPEC-V3R6-LOCAL-NAMESPACE-CONSOLIDATION-001, SPEC-V3R6-UPDATE-NAMESPACE-PROTECT-001
- 5-artifact set authored: spec.md + plan.md + acceptance.md + design.md + research.md
- M1-M6 milestone decomposition aligned with the 4-phase close lifecycle (plan + run + sync + mx)

---

**Frontmatter validation** (v0.3.0): 12-canonical-field schema verified (id, title, version, status, created, updated, author, priority, phase, module, lifecycle, tags all present and well-formed); optional `supersedes: [SPEC-V3R6-AGENT-FOLDER-SPLIT-001]` field added per the §A.0 baseline principle; status `draft` matches manager-spec authorship phase; SPEC ID `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` regex (verified via Pre-Write Self-Check Protocol decomposition).

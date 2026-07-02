---
id: SPEC-DEPRECATEDPATHS-RECONCILE-001
title: "Reconcile DeprecatedPaths: un-deprecate live v3 config design.yaml + db.yaml"
version: "1.0.0"
status: completed
created: 2026-07-02
updated: 2026-07-02
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/defs"
lifecycle: spec-anchored
tags: "config, dead-code, cleanup, deprecated-paths, migration-manifest, reconcile"
tier: S
era: V3R6
related_specs: [SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001, SPEC-DEAD-CONFIG-001, SPEC-V3R2-MIG-003]
---

# SPEC-DEPRECATEDPATHS-RECONCILE-001 — Reconcile DeprecatedPaths: un-deprecate live v3 config design.yaml + db.yaml

## HISTORY

- 2026-07-02 — v0.1.0 — draft — manager-spec — Initial plan-phase authoring (Tier S). Recommends
  **Direction A (un-deprecate)** after a ground-truth-shift investigation: the origin SPEC
  (SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001) deliberately marked `design.yaml` (REQ-VVCR-011) and `db.yaml`
  (REQ-VVCR-012) for removal on premises that were true at authoring time (2026-05-25) but became stale.
  Both files are now shipped by the v3 template AND read by production code. Direction B (complete the
  deprecation) is rejected with recorded rationale (§4). Scope: the 2 config yaml files ONLY.

## §1. Context and Motivation

`internal/defs/dirs.go` `DeprecatedPaths` is the SSOT for v.2.x → v3 cleanup targets. It lists
`.moai/config/sections/design.yaml` (dirs.go:227) and `.moai/config/sections/db.yaml` (dirs.go:233) as
Category B removal targets (`RemovalSchedule: "v3.0.0"`, `DeprecatedBy: SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001`).

But both files are **shipped by the v3 template** AND **read by production code**. Listing a live,
template-shipped, production-read config file as a v2→v3 removal target is a contradiction that causes two
concrete harms on every `moai update`:

1. **Update churn + silent user-edit loss.** `scanDeprecatedPaths` (`internal/cli/update_cleanup.go:119`)
   finds `design.yaml` / `db.yaml` in every v3 project, backs them up under `.moai/backup/` and removes
   them; the same `moai update` then re-deploys the template `design.yaml` / `db.yaml`. Net effect: every
   update backs-up + deletes + re-deploys these two live files, and any user customization (e.g. `db.yaml`
   `migration_patterns` / `orm` / `multi_tenant` set via `/moai db init`, or `design.yaml`
   `gan_loop.pass_threshold`) is moved to backup and replaced by the template default — silent loss of
   user edits.
2. **False v2-detection.** `probeDeprecatedPathSignal` (`internal/cli/v2_detection.go:199`) returns
   positive (project flagged legacy-v2) as soon as it finds `design.yaml` OR `db.yaml`. Because the v3
   template ships both, **every freshly-initialized v3 project is immediately mis-flagged as legacy-v2**,
   which can route a pristine v3 project into the v2→v3 clean-reinstall code path.

### §1.1 Verified findings (plan-phase — confirmed against the tree, not assumed)

| Finding | Evidence (verified) |
|---------|---------------------|
| `design.yaml` shipped in template | `internal/template/templates/.moai/config/sections/design.yaml` present (3091 bytes). |
| `design.yaml` loaded in production | `internal/config/loader.go:92` calls `l.loadDesignSection(sectionsDir, cfg)` (in the `Loader.Load` chain); `loader_design.go:56` reads `design.yaml` into `cfg.Design`. |
| `design.yaml` reclassified live by a sibling audit | `SPEC-DEAD-CONFIG-001` (`status: completed`) §D explicitly parks it: "`design.yaml`, `feedback.yaml`, `interview.yaml` — agent/orchestrator-consumed via skills/rules that read the YAML directly … Do NOT touch." |
| `db.yaml` shipped in template | `internal/template/templates/.moai/config/sections/db.yaml` present (1444 bytes), with **real** `migration_patterns` (`prisma/schema.prisma`, …) — NOT `_TBD_` placeholders. |
| `db.yaml` actively read in production | `internal/cli/hook.go:448` `loadMigrationPatterns` reads `.moai/config/sections/db.yaml` and feeds `migration_patterns` into the db-sync PostToolUse hook (`SchemaSyncConfig.MigrationPatterns`, `internal/hook/dbsync/db_schema_sync.go:50`). Provenance: SPEC-DB-SYNC-001 (cited at `hook.go:436`). |
| origin removal premise for `db.yaml` is STALE | origin REQ-VVCR-012 removed `db.yaml` as a `_TBD_`-marker placeholder. `grep -c '_TBD_'` on the current template `db.yaml` = **0**. The `_TBD_` premise no longer holds. |
| the other 3 config yamls are correctly deprecated (out of scope) | `gate.yaml`, `memo.yaml` absent from BOTH trees; `github-actions.yaml` is a live but self-removing-at-v3.0.0 file owned by SPEC-CI-MULTI-LLM-001 (per SPEC-DEAD-CONFIG-001 §D). All 3 are **out of scope** here (§4). |
| pinned test constraint | `internal/defs/dirs_test.go` `TestDeprecatedPathsTotalCount` asserts `want = 43`; `TestDeprecatedPathsCategorySplit` pins `wantCategoryB = 31`; `TestDeprecatedPathsCategoryBExpectedEntries` enumerates the 31 Category B paths. `@MX:ANCHOR` (dirs_test.go:3) makes DeprecatedPaths the SSOT — any change MUST update the slice and its tests in lockstep. |

### §1.2 Why this is a ground-truth shift, not a repudiation of the origin SPEC

The origin SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 (`status: completed`, `git show 68e3af7b1`) was **correct at
authoring time (2026-05-25)**:

- `design.yaml` was a design-domain asset, and the design SYSTEM (skills `moai-workflow-gan-loop` /
  `moai-workflow-design`, `.claude/rules/moai/design`, `.moai/project/brand`) was being retired
  (REQ-VVCR-011 removed those in their entirety).
- `db.yaml` was a `_TBD_`-marker placeholder (REQ-VVCR-012).

The ground truth then shifted **after** the origin deprecation was written:

- `db.yaml` was **re-established as live v3 config** by SPEC-DB-SYNC-001 (real `migration_patterns`,
  consumed by `loadMigrationPatterns`). The `_TBD_` content is gone.
- `design.yaml` config + `loadDesignSection` loader were **retained** (the design-retire cohort removed
  the skills/rules/brand assets, not the config section loader), and SPEC-DEAD-CONFIG-001 §D later
  reclassified `design.yaml` as live "Do NOT touch".

The `DeprecatedPaths` manifest was never updated to reflect that these two files stopped being v2 removal
targets. This SPEC performs that update. It is a **manifest-vs-ground-truth reconciliation** — the origin
SPEC's original intent (and its audit trail) is preserved, not repudiated.

### §1.3 Recommended direction

**Direction A (un-deprecate)** — remove the 2 entries from `DeprecatedPaths` and update the pinned tests +
doc-comment counts in lockstep. This is the correct direction because both files are live, shipped, and
production-read. Direction B (complete the deprecation by removing template + loaders) is rejected — see §4.

## §2. Requirements (GEARS)

### REQ-DPR-001 — Un-deprecate the 2 live config yaml files (Unwanted behavior)

The `DeprecatedPaths` manifest (`internal/defs/dirs.go`) **shall not** list
`.moai/config/sections/design.yaml` or `.moai/config/sections/db.yaml` as v2→v3 removal targets, because
both are shipped by the v3 template AND read by production code (`loadDesignSection` / `loadMigrationPatterns`).

### REQ-DPR-002 — Update the pinned count assertions in lockstep (Event-driven)

**When** the two entries are removed from the `DeprecatedPaths` slice, the `internal/defs/dirs_test.go`
assertions **shall** be updated in the same commit: `TestDeprecatedPathsTotalCount` `want` 43 → 41,
`TestDeprecatedPathsCategorySplit` `wantCategoryB` 31 → 29, and `TestDeprecatedPathsCategoryBExpectedEntries`
`wantCategoryB` slice shall drop the two paths — so no intermediate commit leaves the test red.

### REQ-DPR-003 — Preserve all other DeprecatedPaths entries (Ubiquitous)

The manifest **shall** retain every other entry unchanged: the `.moai/db` directory entry, the
`.moai/project/brand` entry, and the 3 genuinely-deprecated config yamls (`gate.yaml`,
`github-actions.yaml`, `memo.yaml`). Only the two live config yaml **files** are removed; the Category A
(9) and Category C (3) subtotals are untouched.

### REQ-DPR-004 — Reconcile the §A.4 count-derivation SSOT reference (Where — capability gate)

**Where** the `dirs_test.go` `@MX:ANCHOR` designates SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 §A.4 as the
canonical entry-count derivation, the run-phase **shall** reconcile the 43 → 41 discrepancy so no dangling
count reference remains — either by redirecting the test's `@MX:ANCHOR`/`@MX:REASON` to name this reconcile
SPEC as the 41-entry correction authority (recommended, preserves the completed origin SPEC's body
immutability) OR by adding a bounded correction note to origin §A.4 (alternative). See plan.md §C.

### REQ-DPR-005 — No production behavior, template, or loader change (Ubiquitous)

The change **shall** modify only the `DeprecatedPaths` manifest, its pinned tests, and doc-comment counts.
It **shall not** modify any template file, any loader (`loadDesignSection`, `loadMigrationPatterns`,
`LoadDesignConfig`), or any production Go behavior. No `make build` is required (no template touched).

## §3. Acceptance Criteria (inline — Tier S)

| AC | Given / When / Then | Verification |
|----|---------------------|--------------|
| AC-DPR-001 | Given the reconciled manifest, When grepping `dirs.go`, Then neither `sections/design.yaml` nor `sections/db.yaml` appears as a `DeprecatedPaths` entry | `grep -n 'sections/design.yaml\|sections/db.yaml' internal/defs/dirs.go` → 0 matches |
| AC-DPR-002 | When running the total-count test, Then it passes with `want = 41` | `go test -run TestDeprecatedPathsTotalCount ./internal/defs/` → PASS |
| AC-DPR-003 | When running the category-split test, Then it passes with `wantCategoryB = 29` | `go test -run TestDeprecatedPathsCategorySplit ./internal/defs/` → PASS |
| AC-DPR-004 | When running the Category-B enumeration test, Then it passes (2 paths removed from `wantCategoryB`) | `go test -run TestDeprecatedPathsCategoryBExpectedEntries ./internal/defs/` → PASS |
| AC-DPR-005 | Then `.moai/db`, `.moai/project/brand`, `gate.yaml`, `github-actions.yaml`, `memo.yaml` remain in `DeprecatedPaths` | `grep -c '.moai/db\|project/brand\|gate.yaml\|github-actions.yaml\|memo.yaml' internal/defs/dirs.go` → all 5 present |
| AC-DPR-006 | Then no template file, loader, or production Go behavior changed | `git diff --stat` touches only `internal/defs/*` (+ optional §A.4/doc-comment sites); `internal/template/**` untouched |
| AC-DPR-007 | When running the full suite (incl. `internal/cli` update tests that use `len(defs.DeprecatedPaths)` dynamically), Then it is green | `go test ./...` → ok; `go test ./internal/cli/ -run 'TestUpdate.*Cleanup\|TestUpdate.*E2E'` → PASS |
| AC-DPR-008 | Then the §A.4 count-derivation reference is reconciled (no dangling 43 vs 41) | `@MX:ANCHOR`/`@MX:REASON` in dirs.go + dirs_test.go cite the 41-entry count + this reconcile SPEC |
| AC-DPR-009 | Then cross-platform build succeeds | `GOOS=windows GOARCH=amd64 go build ./...` → exit 0 |

## §4. Out of Scope

This SPEC is deliberately narrow: the two live config yaml **files** in the `DeprecatedPaths` manifest.
Everything below is out of scope; do NOT change these.

### Out of Scope — gate.yaml / github-actions.yaml / memo.yaml DeprecatedPaths entries

- `gate.yaml` and `memo.yaml` are absent from BOTH the local and template trees — genuinely-deprecated v2
  leftovers; their `DeprecatedPaths` entries are correct and stay.
- `github-actions.yaml` is the live config of SPEC-CI-MULTI-LLM-001 but is intentionally scheduled to
  self-remove at v3.0.0 via `DeprecatedPaths` (per SPEC-DEAD-CONFIG-001 §D). Reconciling it would incur a
  cross-reference burden (docs-site ×4 + the owning SPEC) and is out of scope here.

### Out of Scope — Direction B (complete the deprecation) — rejected

- Removing the template `design.yaml` / `db.yaml` and their loaders (`loadDesignSection`,
  `loadMigrationPatterns`) would break live features: `db.yaml` removal regresses `/moai db` + the db-sync
  hook (SPEC-DB-SYNC-001); `design.yaml` removal contradicts SPEC-DEAD-CONFIG-001 §D's explicit "Do NOT
  touch" parking. Both files are currently shipped + production-read — removing them is a regression, not a
  cleanup. Direction B is rejected.

### Out of Scope — design.yaml downstream-Go-consumption (cfg.Design dead-field question)

- Whether the loaded `cfg.Design` value is consumed downstream in Go (a grep for `.Design` consumers
  returned none) is a separate "dead Go API surface" concern already parked by SPEC-DEAD-CONFIG-001 §D
  ("the 4 MIG-003 dead public loaders … possible follow-up SPEC"). This SPEC does NOT resolve design.yaml's
  ultimate fate — only its wrong presence in `DeprecatedPaths`.

### Out of Scope — .moai/db and .moai/project/brand directory entries

- `.moai/db` (the OLD v2 db directory; the v3 db dir is `.moai/project/db` per template `db.yaml`) and
  `.moai/project/brand` remain valid Category B removal targets and stay in `DeprecatedPaths`. Only the two
  config yaml **files** are removed, not these directories.

### Out of Scope — template design.yaml _TBD_ marker content

- The template `design.yaml` still contains one `_TBD_` marker (a content placeholder field). Its content
  is a separate concern from the `DeprecatedPaths` manifest and is not addressed here.

### Out of Scope — modifying the origin SPEC's REQs

- SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001 is `status: completed`; its REQ-VVCR-011/012/014 were correct at
  authoring time. This SPEC does NOT rewrite those REQs — it only reconciles the manifest to the shifted
  ground truth and (per REQ-DPR-004) the count-derivation reference.

## §5. Cross-references

- `internal/defs/dirs.go` — the `DeprecatedPaths` SSOT (2 entries removed; `@MX:ANCHOR` count reconciled).
- `internal/defs/dirs_test.go` — the pinned count tests (43→41, 31→29, enumeration slice).
- `internal/config/loader.go:92` + `internal/config/loader_design.go` — the live `design.yaml` loader (PRESERVE).
- `internal/cli/hook.go:448` `loadMigrationPatterns` — the live `db.yaml` reader (PRESERVE); SPEC-DB-SYNC-001 provenance.
- `internal/cli/update_cleanup.go:119` `scanDeprecatedPaths` + `internal/cli/v2_detection.go:199`
  `probeDeprecatedPathSignal` — the two consumers whose false-positives this SPEC eliminates.
- `SPEC-V3R6-V2-V3-CLEAN-REINSTALL-001` §A.4 — the canonical derivation table (count reconciled per REQ-DPR-004).
- `SPEC-DEAD-CONFIG-001` §D — the sibling audit that parked `design.yaml` as live and left `DeprecatedPaths` intact.
- CLAUDE.local.md §2 (Template-First rule) — governs why template-shipped files must not be v2-removal targets.

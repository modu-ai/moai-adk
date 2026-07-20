---
id: SPEC-CLI-UIKIT-KERNEL-001
title: "Extract shared TUI/settings kernel into internal/cli/uikit leaf package"
version: "0.1.0"
status: completed
created: 2026-07-07
updated: 2026-07-08
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: L
era: V3R6
tags: "refactor, cli, uikit, kernel-extraction, leaf-package, import-cycle, go"
---

# SPEC-CLI-UIKIT-KERNEL-001 — Extract shared TUI/settings kernel into internal/cli/uikit leaf package

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-07 | 0.1.0 | Initial plan-phase draft. Tier L. First follow-up to SPEC-CLI-SUBPKG-SPLIT-001 (completed, M1-only close at sync_commit_sha d0d9b49d7) — extracts the uikit kernel (SPLIT-001 §F M5) that unblocks future kernel-dependent clusters (migrate/doctor/update) from their axis-(i) blocker. Kernel-axis-only scope; SPLIT-001 axes (ii) shared-helper and (iii) reverse-dependency remain deferred to the future migrate SPEC. | manager-spec |

## §A. Context, Intent, and VALUE Justification

SPEC-CLI-SUBPKG-SPLIT-001 closed at sync_commit_sha `d0d9b49d7` with **only M1 (agentlint) shipped**.
Its §F CHECKPOINT (REQ-CSS-010) decision was: STOP at M1, because every remaining cluster
(profile / migrate / doctor / update / uikit) requires design-time coupling resolution, and the
migrate cluster's axis-(i) blocker is `migrate_agency.go:634` calling `RenderError(err)` (defined
`render.go:105`) — a kernel helper that lives in `package cli`, which a future `internal/cli/migrate`
subpackage cannot import (SPLIT-001 design.md §D import-cycle resolution).

This SPEC is the **first follow-up**: it removes that blocker by extracting the shared TUI
kernel helpers from `internal/cli/{render.go, banner.go, schema_bridge.go}` (~365 LOC across 3
files — `settings.go` STAYS in package cli per design.md §C.3 D2 STAYS reclassification) into a
NEW leaf package `internal/cli/uikit` (package `uikit`). The new package is a **leaf**:
it imports neither `package cli` nor any `internal/cli/*` subpackage. That leaf contract is what
unblocks future kernel-dependent clusters (migrate / doctor / update) from importing `uikit` without
forming a cycle.

### This is a refactor of WORKING code — the honest value/risk position

[HARD] `go build ./...` exits **0** today (verified this run, 2026-07-07); every `moai` subcommand
works. This SPEC proposes a **pure structural refactor with NO functional change**. Per the MoAI
"reject over-engineering" behavior, the refactor MUST justify its risk. The honest assessment:

- **VALUE (real, foundational, NOT user-observable)**: (a) **unblocks future clusters** — the uikit
  leaf contract is the prerequisite for migrate (axis-i), doctor, and update cluster extraction
  (SPLIT-001 plan.md §F M3/M6/M7 all gate on M5); (b) **compile-time boundary** — the kernel
  helpers get an explicit exported surface, forcing callers to acknowledge the dependency; (c)
  **consistency** with the existing subpackage pattern (`worktree`/`harness`/`preference`/`wizard`/
  `specid`/`pr` + the SPLIT-001 M1 `agentlint`). None of these change what a user observes.
- **RISK (substantial, characterized this run via grep — NOT assumed)**: (a) **wide caller
  blast radius** — 12 production files + ≥10 test files must be rewritten to `uikit.*` or to
  `uikit.CheckStatus` / `uikit.CheckOK` / `uikit.CheckWarn` / `uikit.CheckFail` (verified caller
  sites: `RenderError` × 1, `renderCard` × 5, `renderKeyValue(Lines)` × 8,
  `renderSuccessCard` × 3, `renderInfoCard` × 3, `PrintBanner` × 3, `PrintWelcomeMessage` × 1,
  `CheckStatus`/`CheckOK`/`CheckWarn`/`CheckFail` × 43 across 7 files; see design.md §D for the
  verbatim file:line map, including §D.8 CheckStatus-rewrite surface and §D.10 test-file list);
  (b) **three cross-file type dependencies** — `render.go:60` `renderStatusLine` consumes the
  `CheckStatus` type (defined `doctor.go:25-26`); `schema_bridge.go:24-77` declares TWO maps
  referencing the `profileSetupText` type (defined `profile_setup_translations.go:10`); and
  `settings.go:28/48/110` references the `SettingsLocal` type (defined `glm.go:97`) — resolved
  by the design.md §C.3 D2 STAYS reclassification (settings.go does NOT move to uikit, avoiding
  the cycle that would form if uikit imported package cli for the type); all three must be
  resolved BEFORE the move (design.md §C); (c) **characterization preservation** — the existing
  53,736 test LOC is the behavior contract (no new tests; design.md §E AP-5).

**Conclusion**: the uikit extraction is the **foundational unblock** for the cluster ladder SPLIT-001
envisioned. It is NOT optional if any of migrate/doctor/update is to ship. But it is also NOT
user-observable, and its blast radius is wider than the 4-file/503-LOC source size suggests. The
§F CHECKPOINT (REQ-CUK-010) therefore reserves the right to STOP at M1 if the caller-rewrite cost
proves uneconomic in run-phase — same shape as SPLIT-001's CHECKPOINT.

## §B. Scope Summary

**In scope** — extract the shared TUI kernel helpers from `internal/cli/{render.go, banner.go,
schema_bridge.go}` (~365 LOC across 3 files; `settings.go` STAYS in package cli per design.md
§C.3 D2 STAYS reclassification) into a NEW leaf package `internal/cli/uikit` (package `uikit`),
preserving the existing `package cli` behavior exactly. Rewrite every current `package cli`
caller of the moved helpers AND every `CheckStatus`/`CheckOK`/`CheckWarn`/`CheckFail` reference
to the `uikit.`-qualified form (43 CheckStatus refs across 7 files per design.md §D.8). Resolve
the three cross-file type dependencies (CheckStatus, profileSetupText, SettingsLocal) as
design-time prerequisites (SettingsLocal resolves via settings.go STAYS, NOT a uikit move).
Behavior-preserving (cycle_type=ddd for the future run-phase: ANALYZE-PRESERVE-IMPROVE).

**Out of scope** — see §E (SPLIT-001 axes (ii)/(iii); profile/migrate/doctor/update cluster
extractions; any functional change; new tests beyond characterization; CHANGELOG/README authoring).

## §C. Requirements (GEARS notation)

### Leaf-package contract (the foundational invariant)

- **REQ-CUK-001** (Ubiquitous): The `internal/cli/uikit` package shall be a leaf package — it shall
  import neither `github.com/modu-ai/moai-adk/internal/cli` (the root package) nor any
  `github.com/modu-ai/moai-adk/internal/cli/*` subpackage. This leaf contract is what permits
  future kernel-dependent clusters to import `uikit` without forming an import cycle (the cycle
  forms because `package cli` `root.go` imports command subpackages for `AddCommand` registration,
  so subpackages cannot import `package cli` back).
- **REQ-CUK-002** (Ubiquitous): The `internal/cli/uikit` package shall declare `package uikit` in
  every `.go` file it contains, and shall export the helpers that `package cli` and future clusters
  consume (the export inventory is documented in design.md §C; the run phase verifies every moved
  helper is referenced via `uikit.*` by at least one caller OR is test-only).

### Behavior preservation (the invariant that governs the extraction)

- **REQ-CUK-003** (Ubiquitous): The CLI shall preserve the observable behavior of every `moai`
  subcommand across the extraction — the `moai --help` subcommand list (names, groups, order) and
  every subcommand's runtime behavior shall be identical before and after M1.
- **REQ-CUK-004** (Event-driven): When the uikit extraction completes, the full test suite
  `go test ./...` shall pass with zero NEW failures (pre-existing baseline failures documented
  separately; this is the binding behavior-preservation gate).
- **REQ-CUK-005** (Event-driven): When the uikit extraction completes, both `go build ./...` and
  `GOOS=windows GOARCH=amd64 go build ./...` shall exit 0 — platform-tagged sibling files (none
  expected in the 4 source files, verified at run-phase) moved together if present.

### Structural correctness

- **REQ-CUK-006** (State-driven / While): While the moved helpers relocate to `uikit`, every
  current `package cli` caller of those helpers shall be rewritten to `uikit.<Helper>(...)` — no
  caller shall be left referencing the pre-move unexported name (compile-break) and no caller shall
  be left silently re-defining a duplicate (DRY-break). The caller-rewrite blast radius is
  characterized verbatim in design.md §D.
- **REQ-CUK-007** (Capability gate / Where): Where the moved helpers carry cross-file type
  dependencies (`CheckStatus` defined `doctor.go:25-26` and consumed by `render.go:60`
  `renderStatusLine` AND by 6 other files — `doctor_cache.go`, `doctor_harness.go`, and 4 test
  files totaling 43 references per design.md §D.8; `profileSetupText` defined
  `profile_setup_translations.go:10` and consumed by BOTH `schema_bridge.go:24-58
  schemaFieldBridge` AND `schema_bridge.go:60-77 schemaSegmentBridge` per design.md §C.4 D5 fix;
  `SettingsLocal` defined `glm.go:97` and consumed by `settings.go:28/48/110` per design.md §C.3
  D2 STAYS reclassification), those types shall be co-located into the uikit leaf (or the
  coupling shall be resolved by a design-time decision: CheckStatus co-locates into
  `uikit/types.go`; profileSetupText coupling resolves via b-ii split keeping BOTH
  `schemaFieldBridge` AND `schemaSegmentBridge` in package cli; SettingsLocal coupling resolves
  by `settings.go STAYS in package cli` — the leaf is narrowed to render + banner + schema-bridge
  helpers only) BEFORE the source files move — the run phase MUST NOT leave `uikit` importing a
  sibling (cycle) or referencing an undefined symbol (compile-break).
- **REQ-CUK-008** (Ubiquitous): The public entry point `cli.Execute()` shall remain in
  `package cli` unchanged, so external callers (`cmd/moai/main.go`) are unaffected by the
  extraction.

### Phasing discipline

- **REQ-CUK-009** (Ubiquitous): The uikit extraction shall ship as a single M1 milestone with its
  own test-verify gate (REQ-CUK-004/005) and its own atomic commit — the source-file moves, the
  type-dependency resolutions, the caller rewrites, and the test-file rewrites all land in one
  commit (one-cluster-per-milestone atomic commit, mirroring SPLIT-001 REQ-CSS-009).
- **REQ-CUK-010** (Event-driven): When the post-M1 re-evaluation determines that the uikit
  extraction recipe held AND that future kernel-dependent clusters (migrate/doctor/update) are
  worth their remaining (post-uikit) coupling-resolution cost, the work shall proceed to those
  clusters' separate SPECs; otherwise the work shall STOP with M1 shipped and the decision
  recorded. This is the SPLIT-001 REQ-CSS-010 checkpoint analogue.
- **REQ-CUK-011** (Unwanted behavior): The extraction shall not introduce any functional change,
  bug fix, or new feature — no logic edit beyond symbol re-scoping (export/unexport), type
  co-location, and import rewiring is permitted within M1.
- **REQ-CUK-012** (Unwanted behavior): The extraction shall not weaken the test suite — no existing
  test shall be deleted or skipped during the move; a test that cannot compile in its new package
  signals an incomplete symbol relocation and shall block M1 (not be removed).
- **REQ-CUK-013** (Unwanted behavior): The extraction shall not introduce any `AskUserQuestion` or
  `mcp__askuser` invocation in the uikit package — the subagent boundary (sentinel C-HRA-008) is
  preserved across the move; this REQ parents AC-CUK-009.

## §D. Acceptance Criteria Pointer

Full Given-When-Then scenarios (the M1 behavior-preservation gates, the leaf-package correctness
gate, the caller-rewrite completeness gate, the type-co-location gate, the checkpoint
stop-condition, and the Definition of Done) live in `acceptance.md`. The binding
behavior-preservation command is `go test ./...` (zero NEW failures vs the documented baseline)
plus the cross-platform build matrix.

## §E. Out of Scope

### Out of Scope — SPLIT-001 axis (ii) shared-helper co-extraction
- The shared-helper axis SPLIT-001 research.md §C.4.3(ii) documents (`migrate_restore_skill.go`
  calling `validateSkillID` / `archiveVersion` / `copyDirAll` co-located in `update_archive.go`)
  is NOT resolved by this SPEC. The uikit kernel is TUI/settings helpers only; the archive
  helpers are a separate shared-helper surface that co-extracts with the migrate cluster or
  into its own `archiveutil` leaf. See future migrate SPEC.

### Out of Scope — SPLIT-001 axis (iii) reverse-dependency encapsulation
- The reverse-dependency axis SPLIT-001 research.md §C.4.3(iii) documents (`update.go:1922-1928`
  constructing `&migrateAgencyRunner{}` writing unexported fields; `update.go:1933-1935` type-
  asserting `*MigrateError`; `update_archive.go` constructing `&MigrateError{}` at 5 sites) is NOT
  resolved by this SPEC. It lands in the future update SPEC (SPLIT-001 M7 analogue) or a prior
  refactor commit. Uikit extracts only the kernel-render helpers.

### Out of Scope — profile / migrate / doctor / update cluster extractions
- The profile cluster (SPLIT-001 §C.4.4 — name collision + `schema_bridge.go` reverse-dep) is a
  SEPARATE future SPEC. The migrate cluster (SPLIT-001 §C.4.3 tri-axis coupled) is a SEPARATE
  future SPEC that GATES ON this uikit SPEC for its axis-(i) resolution. The doctor cluster
  (SPLIT-001 §B 9 files / 2,357 LOC heavy kernel use) and the update cluster (SPLIT-001 §B 9
  files / 5,181 LOC + 9,283 test LOC + deps injection) are SEPARATE future SPECs that GATE ON this
  uikit SPEC. None of these clusters are extracted in this SPEC.

### Out of Scope — Functional and behavioral change
- No `moai` subcommand behavior changes. No bug fixes, no new flags, no new subcommands, no
  refactor-of-logic "while we're here" (REQ-CUK-011). The public API (`cli.Execute()`) is
  unchanged (REQ-CUK-008). This is a pure structural move.

### Out of Scope — New tests beyond characterization
- No new behavior tests are authored. The existing 53,736 test LOC is the behavior contract; the
  extraction relocates test files that move with their source (e.g. `render_test.go`, the
  `PrintWelcomeMessage` tests in `misc_coverage_test.go`) unchanged (design.md §E AP-5, mirroring
  SPLIT-001). Adding tests during a refactor conflates refactor risk with feature risk.

### Out of Scope — Subpackage other than uikit
- No `internal/cli/<domain>/` subpackage other than `uikit` is created, modified, or merged in
  this SPEC. The 7 existing subpackages (`worktree`, `harness`, `preference`, `wizard`, `specid`,
  `pr`, `agentlint`) are preserved unchanged. The root-level `harness_*.go` cluster is NOT touched
  (SPLIT-001 §E out-of-scope carries forward).

### Out of Scope — CHANGELOG and README
- `CHANGELOG.md` and `README.md` entries for this refactor are owned by `manager-docs` in the sync
  phase, not authored here.

---
id: SPEC-CLI-SUBPKG-SPLIT-001
title: "Split flat internal/cli package into cohesive command subpackages (phased)"
version: "0.1.0"
status: draft
created: 2026-07-02
updated: 2026-07-07
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
tier: L
era: V3R6
tags: "refactor, cli, package-split, maintainability, cobra, go"
---

# SPEC-CLI-SUBPKG-SPLIT-001 — Split flat internal/cli package into cohesive command subpackages (phased)

## HISTORY

| Date | Version | Change | Author |
|------|---------|--------|--------|
| 2026-07-02 | 0.1.0 | Initial plan-phase draft. Tier L. Speculative maintainability refactor of WORKING code (build baseline green). Honest value/risk framing: phased lowest-risk-first extraction of high-value clusters; big-bang split rejected; 14 tiny single-file clusters + deps/platform-tangled clusters deferred. | manager-spec |

## §A. Context, Intent, and VALUE Justification

`internal/cli` is a **flat root package of 93 non-test `.go` files / 25,838 LOC** (recursive
33,321 LOC across 139 files). `update.go` alone is **3,170 LOC** (12% of the package). Only a
handful of concerns have been extracted into subpackages (`worktree`, `harness`, `pr`,
`preference`, `specid`, `wizard`); the remaining ~26 command clusters live flat in `package cli`.
A 90+-file flat package is a navigability and maintainability smell.

### This is a refactor of WORKING code — the honest value/risk position

[HARD] `go build ./internal/cli/...` exits **0** today; every `moai` subcommand works. This SPEC
proposes a **pure structural refactor with NO functional change**. Per the MoAI "reject
over-engineering" behavior, a large refactor of working code MUST justify its risk. The honest
assessment:

- **VALUE (real but incremental, NOT user-observable)**: (a) navigability — a reader locates the
  update engine in `cli/update/` instead of scanning 93 flat files; (b) compile-time boundaries —
  subpackages force an explicit exported/unexported API surface per domain; (c) test isolation —
  per-cluster `_test.go` packages instead of one 54,756-LOC white-box test blob; (d) fewer
  merge conflicts on parallel edits; (e) consistency with the already-established subpackage
  pattern. None of these change what a user observes.
- **RISK (substantial)**: (a) **test-migration surface** — 147 test files / **54,756 test LOC**,
  white-box (`package cli`) tests that access unexported symbols; moving a cluster relocates its
  tests and re-resolves every unexported reference (the `update` cluster alone = 9,283 test LOC);
  (b) **import-cycle hazard** — `root.go` imports subpackages for `AddCommand`, so subpackages
  cannot import shared helpers back from `package cli`; kernel-dependent clusters are blocked on a
  prior leaf-package extraction; (c) **global `deps` coupling** (8 files) needs provider injection;
  (d) 15 cross-platform build-tag files whose siblings must move together.

**Conclusion (the signal the plan-auditor and user need)**: the maintainability gain is genuine
but marginal and invisible to users; the regression risk from the test surface + import cycle is
real. A **big-bang 93-file reorganization is NOT justified**. What IS justified is a **bounded,
phased, lowest-risk-first extraction of the 3-4 highest-value clusters**, each independently
test-verified and committed, with an explicit re-evaluation checkpoint that authorizes STOPPING
when marginal value drops. The 14 tiny single-file clusters and the deps/platform-tangled
clusters are explicitly out of scope (see §E). See plan.md for the milestone ladder and the
recommendation, research.md for the cluster map + measurements, design.md for the migration recipe.

## §B. Scope Summary

**In scope** — a phased extraction of cohesive, high-value, low-to-medium-risk clusters from the
flat `internal/cli` root package into `internal/cli/<domain>/` subpackages, following the existing
subpackage pattern. Each extraction is behavior-preserving (files move; symbols re-scoped; cobra
registration preserved; deps via provider injection; import cycle resolved via a `uikit` leaf
package when a kernel-dependent cluster is reached). Committed milestones (risk-ascending):
`migrate`, `profile`, `agentlint`, then a re-evaluation checkpoint; conditionally the `uikit`
kernel extraction + `doctor`, and conditionally the highest-value `update` cluster.

**Out of scope** — see §E (big-bang split; 14 tiny single-file clusters; deps/platform-tangled
`launch/glm`, `hook`, `speccmd`; any functional/behavioral change; new tests beyond
characterization; CHANGELOG/README authoring).

## §C. Requirements (GEARS notation)

### Behavior preservation (the invariant that governs every milestone)

- **REQ-CSS-001** (Ubiquitous): The CLI shall preserve the observable behavior of every `moai`
  subcommand across each extraction — the `moai --help` subcommand list (names, groups, order)
  and every subcommand's runtime behavior shall be identical before and after a milestone.
- **REQ-CSS-002** (Event-driven): When a cluster extraction milestone completes, the full test
  suite `go test ./...` shall pass with zero failures, exercising the moved white-box tests in
  their new package. This is the binding behavior-preservation gate.
- **REQ-CSS-003** (Event-driven): When a cluster extraction milestone completes, both
  `go build ./...` and `GOOS=windows GOARCH=amd64 go build ./...` shall exit 0 — platform-tagged
  sibling files (`*_unix.go` / `*_windows.go` / `*_posix.go`) moved together.

### Structural correctness

- **REQ-CSS-004** (Ubiquitous): Each extracted cluster shall reside in an `internal/cli/<domain>/`
  leaf package declaring `package <domain>`, exporting exactly the cobra command variable/factory
  that `root.go` registers, and keeping cluster-internal symbols unexported.
- **REQ-CSS-005** (State-driven / While): While `root.go` wires subcommands, it shall register each
  extracted command via exactly one `rootCmd.AddCommand(<domain>.<Cmd>)` call — no command shall be
  double-registered (cobra panics on duplicate `Use:` prefix) and no command shall be dropped from
  registration.
- **REQ-CSS-006** (Capability gate / Where): Where an extracted cluster consumes the shared
  unexported TUI/settings helpers (`renderCard`, `renderKeyValue`, `renderStatusLine`,
  `RenderError`, `PrintBanner`, `mutateSettingsLocal`, `schemaKeyToTUIField`), those helpers shall
  first be relocated to the `internal/cli/uikit` leaf package and exported, so the extracted
  subpackage imports `uikit` rather than importing `package cli` (which would create an import
  cycle, because `root.go` imports the command subpackages).
- **REQ-CSS-007** (Where): Where an extracted cluster couples the package-level `deps` global
  (one of `update`, `glm`, `hook`, `hook_pre_push`, `inventory`, `loop`, `spec_lint`), the cluster
  shall receive its dependencies via a package-level provider variable set by `root.go`
  (the `worktree.WorktreeProvider = deps.GitWorktree` precedent) — the Composition Root
  (`InitDependencies` in `deps.go`) shall remain in `package cli`.

### Public entry-point stability

- **REQ-CSS-008** (Ubiquitous): The public entry point `cli.Execute()` shall remain in
  `package cli` unchanged, so external callers (`cmd/moai/main.go`) are unaffected by the split.

### Phasing discipline (reject over-engineering)

- **REQ-CSS-009** (Ubiquitous): Each cluster shall be extracted as an independent milestone with
  its own test-verify gate (REQ-CSS-002/003) and its own atomic commit — no milestone shall bundle
  more than one cluster (the big-bang split is prohibited).
- **REQ-CSS-010** (Event-driven): When the post-checkpoint re-evaluation determines that the
  marginal maintainability value of the next milestone no longer justifies its migration risk, the
  work shall STOP with the extracted clusters shipped — the SPEC shall not force completion of the
  full cluster ladder.
- **REQ-CSS-011** (Unwanted behavior): The extraction shall not introduce any functional change,
  bug fix, or new feature — no logic edit beyond symbol re-scoping (export/unexport) and import
  rewiring is permitted within an extraction milestone.
- **REQ-CSS-012** (Unwanted behavior): The extraction shall not weaken the test suite — no existing
  test shall be deleted or skipped during a move; a test that cannot compile in its new package
  signals an incomplete symbol relocation and shall block the milestone (not be removed).

## §D. Acceptance Criteria Pointer

Full Given-When-Then scenarios (per-milestone behavior-preservation gates, the import-cycle-free
build gate, the deps-provider-injection gate, the Definition of Done, and the checkpoint
stop-condition) live in `acceptance.md`. The binding behavior-preservation command is
`go test ./...` (zero failures) plus the cross-platform build matrix.

## §E. Out of Scope

### Out of Scope — Big-bang full-package reorganization
- Splitting all ~26 clusters in a single change is explicitly rejected (design.md §G AP-1). A
  single test-compile break would block the entire package's test compilation; the diff would be
  unreviewable; it violates surgical-change discipline. Only the phased, one-cluster-per-milestone
  path (REQ-CSS-009) is in scope.

### Out of Scope — Tiny single-file clusters (churn)
- The 14 single-file clusters — `astgrep`, `loop`, `github`/`pr_watch_cmd`, `v2_detection`,
  `branch_protection`, `statusline`, `research`, `web`, `telemetry`, `tool_policy`, `design_folder`,
  `inventory`, `mx`, `constitution` (when a single small file) — are NOT extracted. Moving a
  69-370 LOC file into its own package is pure churn with near-zero navigability value and negative
  import-line cost (design.md §G AP-4).

### Out of Scope — deps/platform-tangled high-risk clusters
- `launch`/`glm` (deps coupling + `syscall.Exec` / tmux platform tangle across 11 files),
  `hook` (deps coupling), and `speccmd` (deps coupling via `spec_lint`) are DEFERRED. They are
  extractable in principle via the same recipe + provider injection, but their combined deps +
  platform-tag friction places them beyond the justified value threshold for this SPEC. A
  follow-up SPEC may pick them up if a specific maintenance pain justifies the risk.

### Out of Scope — Functional and behavioral change
- No `moai` subcommand behavior changes. No bug fixes, no new flags, no new subcommands, no
  refactor-of-logic "while we're here" (REQ-CSS-011). The public API (`cli.Execute()`) is
  unchanged (REQ-CSS-008). This is a pure structural move.

### Out of Scope — New tests beyond characterization
- No new behavior tests are authored. The existing 54,756 LOC of tests are the behavior contract;
  the migration relocates them unchanged (design.md §E, AP-5). Adding tests during a refactor
  conflates refactor risk with feature risk.

### Out of Scope — Existing subpackage reorganization
- The already-extracted subpackages (`worktree`, `harness`, `preference`, `wizard`, `specid`,
  `pr`) are NOT restructured. The root-level `harness_*.go` cluster is NOT merged into the existing
  `internal/cli/harness` subpackage in this SPEC (naming-collision reconciliation is deferred).

### Out of Scope — CHANGELOG and README
- `CHANGELOG.md` and `README.md` entries for this refactor are owned by `manager-docs` in the sync
  phase, not authored here.

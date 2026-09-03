---
id: SPEC-GIT-STATUS-FIXTURE-001
title: "Pin bare-remote initial branch in internal/core/git status test fixtures (config-independence for CI)"
version: "0.1.0"
status: in-progress
created: "2026-09-04"
updated: "2026-09-04"
author: "manager-spec (card t474)"
priority: "P1"
phase: "v3.2.0 target"
module: "internal/core/git"
lifecycle: spec-anchored
tags: "testing, fixture, ci, git"
tier: S
---

# SPEC-GIT-STATUS-FIXTURE-001 — Pin bare-remote initial branch in status test fixtures

Card: t474 · Branch: `WT-git-status-fixture` · Base: `25a3212a9` (origin/develop tip)
Evidence base: `.moai/reports/t474/repro-probe.md` (pre-plan measured causal chain; this SPEC plans from it, it does not re-derive it).

## §1 Problem Statement

Two `internal/core/git` tests are deterministically red on develop CI (`Test (ubuntu-latest)` + `Race Test`) while green on dev machines that carry Xcode's system gitconfig pinning `init.defaultBranch=main`:

- `TestStatusAheadBehindFromHeader` — `status_branch_test.go:269` `Ahead=1 Behind=0, want 1/1`
- `TestStatusBranchHeaderShapes` — `:122` `diverged header = "## main...origin/main [ahead 1]"` and `:128` `fatal: ambiguous argument 'HEAD~1'`

One root, two tests: the bare-remote fixture inits at `status_branch_test.go:99` and `:235` run `git init -q --bare` WITHOUT `-b main`, so the remote's HEAD symref follows ambient `init.defaultBranch` (pinned `main` on dev machines, git's compiled default `master` on CI). Pushed `main` does not satisfy a HEAD pointing at `refs/heads/master` → `git clone` degrades to an unborn-HEAD clone → the clone's commit lands on `master` and its `push -q` exits 0 while creating a NEW `master` branch on the remote (`main` untouched) → the first repo's fetch never updates `origin/main` → all three CI failure signatures follow exactly (probe L4/L5, byte-identical fatal signature).

The in-repo repair precedent already exists: `helpers_test.go:42` runs `runGit(t, remoteDir, "init", "--bare", "-b", "main")`. The defect is fixture-side only.

## §2 Requirements (GEARS)

- **REQ-GSF-001** (Ubiquitous): The bare-remote fixtures at `status_branch_test.go:99` and `:235` shall pin the initial branch via `git init -q --bare -b main <path>`, matching the pinned pattern already established at `helpers_test.go:42`.
- **REQ-GSF-002** (Unwanted): The `internal/core/git` test fixtures shall not depend on ambient `init.defaultBranch` configuration for the bare remote's HEAD symref target.
- **REQ-GSF-003** (Where): Where the kickoff gate accepts the optional `:315` consistency pin, the bare init in `TestNewRepositoryErrorTaxonomy` (`status_branch_test.go:315`) shall also carry `-b main`; where the gate drops it, that line shall remain untouched.
- **REQ-GSF-004** (Ubiquitous): The run-phase change set shall touch only test files under `internal/core/git` (no production-path diff).

## §3 Acceptance Criteria (inline — Tier S)

All ACs carry both cells per the repo's verification-completeness doctrine: **RED-now** = the currently observed failing state with its evidence; **Green path** = which milestone flips it and what passing output becomes. A "tests pass locally" criterion is deliberately absent — they already pass locally (probe L0); the local machine cannot discriminate the defect.

- **AC-GSF-001 — Pin present at the two red sites.**
  - RED-now: `grep -n '"--bare"' internal/core/git/status_branch_test.go` shows lines 99 and 235 WITHOUT an adjacent `"-b", "main"` token — the CI-red state (probe L6).
  - Green path (M1): the same grep shows `init -q --bare -b main` at both sites, matching the `helpers_test.go:42` shape. **Mutant probe**: a mutant reverting either pin must flip this AC to FAIL — run the grep form, not a visual read.
- **AC-GSF-002 — Full package sweep green after the change.**
  - RED-now: `go test ./internal/core/git/ -count=1` is green locally and CI-red on develop (dual-job red relayed in the dispatch) — local green is not discriminating evidence (probe L0/L7).
  - Green path (M2): `go test ./internal/core/git/ -count=1` exits 0 with `ok github.com/modu-ai/moai-adk/internal/core/git` AND a non-zero test count visible (`-v` or `-json` sweep) — not a selector that matches zero tests.
- **AC-GSF-003 — Config-independence (deciding verification, external dependency).**
  - RED-now: the deciding environment (CI, no `init.defaultBranch` config) shows the two tests red on develop — evidence: dispatch CI excerpt + probe L4/L5 mechanism reproduction with byte-identical fatal signature.
  - Green path (M3, external): the CI run on the **merged develop tree after the lead's push** is green on `Test (ubuntu-latest)` and `Race Test`. This is a bidirectional confirmation that only the merged-tree CI can deliver — the plan does NOT claim local equivalence and the card does not close before this verdict lands.
- **AC-GSF-004 — No production-path diff.**
  - RED-now: n/a (scope assertion; RED-now baseline is the pre-change tree at `25a3212a9`).
  - Green path (M2): `git diff --name-only 25a3212a9...HEAD` lists only `internal/core/git/status_branch_test.go` (plus, if accepted at kickoff, the same file for `:315` — same file either way). Any `.go` non-test file in the list fails this AC.
- **AC-GSF-005 — Formatting clean.**
  - RED-now: n/a (formatting is not the defect).
  - Green path (M2): `gofmt -l .` prints 0 rows after the change.
- **AC-GSF-006 — Optional `:315` pin decided explicitly.**
  - RED-now: `status_branch_test.go:315` carries the unpinned bare-init shape (probe L6: benign today — nothing clones from it).
  - Green path (M1 or kickoff): the Implementation Kickoff Approval gate records accept-or-drop for the `:315` consistency pin; if accepted, the AC-GSF-001 grep form also matches `:315`; if dropped, `:315` is byte-unchanged. Never bundled silently.

## §4 Constraints

- Repair shape is fixed: add `-b main` (and only that) to the bare inits at `:99` and `:235`. No helper refactors, no fixture rewrites — the minimal diff matching the in-repo precedent.
- Local RED of the exact tests is not achievable in this session (probe L7 negative result; system/global config edits and env-var injection both rejected for lane-safety reasons). The RED-now evidence is the L4/L5 mechanism probe plus the live CI red; verification is planned accordingly.
- No CI workflow config changes (no `init.defaultBranch` workaround on runners) — the durable repair removes the config dependence rather than accommodating it (probe Residual-risk).

## §5 Non-Goals

Production code under `internal/core/git` is NOT implicated — the measured causal chain sits entirely in the test fixture layer. Nothing in this SPEC plans changes to git-status parsing production paths, `NewRepository`, `Status`, or any non-test file.

### Out of Scope — production code

- Any change to `internal/core/git/*.go` non-test files (status parsing, repository wrapper, error taxonomy).

### Out of Scope — CI runner configuration

- Setting `init.defaultBranch` in `.github/workflows/**` or runner setup — a config workaround would leave the defect latent, not fixed.

### Out of Scope — fixture refactoring

- Extracting a shared bare-remote helper, restructuring the fixtures, or touching any other `*_test.go` in the package beyond the three named line sites (`:99`, `:235`, optional `:315`).

## HISTORY

- 2026-09-04 — v0.1.0 draft created (plan-phase, card t474). Root cause pre-measured in `.moai/reports/t474/repro-probe.md`; requirements authored from that chain, not re-derived.

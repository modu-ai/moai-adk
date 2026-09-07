# progress.md — SPEC-GRAPH-STALE-DISPOSITION-001

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-09-07
tier: S
artifacts: [spec.md, plan.md, progress.md]
card: t493
branch: WT-graph-mxindex-edges
```

## §E.2 Run-phase Evidence

Run executed 2026-09-07 in this worktree at HEAD `f59443cd1`, judging build `./bin/moai` (tree-built via `make build`, version-stamp `Commit=f59443cd1`). Full evidence with verbatim outputs: `.moai/reports/t493/verdict.md`.

| AC | Status | Command | Actual Output |
|----|--------|---------|---------------|
| AC-001 | PASS | `./bin/moai graph check` | mx-index `absent value=0 threshold=1`; edges `absent value=0 threshold=0`; codemaps/citations fresh; rc=1 |
| AC-002 | PASS | `./bin/moai mx scan` → `./bin/moai graph build` → `./bin/moai graph check` | mx-index `fresh value=0`; edges `fresh value=0`; codemaps/citations fresh; rc=0; both artifact stamps `commit_sha=f59443cd1…` |
| AC-003 | PASS | `./bin/moai graph build` | `OK: wrote 205162 edges` (mx-spec 108 / spec-depends 165 / code-call 189574 / code-import 14287) |
| AC-004 | PASS | `git show 1d6a902a6 -- internal/cli/todo.go \| git apply -R -` → `./bin/moai graph check` | mx-index `stale value=1 threshold=1 — 1 inventoried file(s) changed content`; rc=1 |
| AC-005 | PASS | `git checkout -- internal/cli/todo.go` → `./bin/moai graph check` | mx-index `fresh value=0`; `git status --porcelain` tracked lines = 0; rc=0 |
| AC-006 | PASS | report read | REQ-007(a)-(e) present; every command named with output; Card Cross-Check row for t493 |
| AC-007 | PASS | `git status --porcelain` | 0 tracked modifications; no `internal/`/`pkg/`/`cmd/`/`gate.yaml` diff vs run-start tree |
| AC-008 | PASS | `make build` + all graph ops | every cited measurement via `./bin/moai` (HEAD-matching build); installed binary `e79c010b8` judged nothing |

Disposition evidence re-measured read-only from this worktree (LC_ALL=C set operations, develop artifacts read as files): mx-index intersection at `df74b3c9d` = **361 exact**, missing 0, 0 inventory entries under `.moai/`; edges moved sets codemaps **6** / specs **836** / reports **780** — the §1.2 investigation figures reproduce exactly at the investigation coordinate (362/855/796 at current develop HEAD `33fcb644b`, consistent growth). Measurement-method note: `comm` requires `LC_ALL=C` on both inputs; an unpinned-collation first pass briefly read 181/182 before correction.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: pending-backfill-run
run_status: complete
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not-applicable-worktree-isolated-card-branch
l44_post_push_fetch: not-applicable-no-push-lane-reports-merge-sha
new_warnings_or_lints_introduced: 0
total_run_phase_files: 2
m1_to_mN_commit_strategy: single-M1-commit
```

## §E.4 Sync-phase Audit-Ready Signal

Final post-commit state measurement, taken 2026-09-07 in this worktree at HEAD `cc9137e3d`, judging build `./bin/moai` (tree-built): `./bin/moai graph check` reports codemaps fresh 15/40, mx-index fresh 0/1, edges stale 2 ("source set(s) moved: reports, specs"), citations fresh, rc=1. Edges is stale again because the card's own report commit moved two edges source sets (`.moai/reports/t493/` added, `.moai/specs/SPEC-GRAPH-STALE-DISPOSITION-001/spec.md` frontmatter transitioned) — this is the disposition SPEC's own conclusion in action: a derived artifact goes stale each time its sources move; correct reporting, not a defect. Deliberately NOT re-run (`mx scan`/`graph build`) to chase it green; the regenerated artifacts (`bin/`, `.moai/state/mx-index.json`, `.moai/project/graph/*`) stay untracked and uncommitted.

```yaml
sync_status: completed
spec_id: SPEC-GRAPH-STALE-DISPOSITION-001
sync_complete_at: 2026-09-07
sync_commit_sha: 7a737a880
sync_subject: docs(SPEC-GRAPH-STALE-DISPOSITION-001): sync-phase — 3-phase close, completed (t493)
frontmatter_status_transitions:
  - in-progress -> implemented -> completed   # full transition rides the single sync commit
changelog_entry: none   # zero production code changed; disposition SPEC with no user-facing surface
b12_self_test_a:
  pre_emission_grep: not-applicable   # no CHANGELOG entry emitted (B12 halted at skip-decision)
b12_self_test_c:
  file_path_verification: not-applicable   # no file paths claimed in a CHANGELOG entry
```

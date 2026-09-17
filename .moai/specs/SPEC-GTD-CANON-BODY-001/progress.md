# Progress — SPEC-GTD-CANON-BODY-001 (card t867)

## §E.1 Plan-phase Audit-Ready Signal

### Iteration 1

- Artifacts: spec.md, plan.md, acceptance.md, progress.md (Tier M)
- SPEC ID regex self-check: `PASS`; uniqueness `ls .moai/specs | grep -ci gtd-canon` → `0`
- Base: develop `f67d2193f`, branch `WT-gtd-canon`, commit `114737ea1`
- Plan-audit iteration 1: FAIL 0.71 (`.moai/reports/plan-audit/SPEC-GTD-CANON-BODY-001-review-1.md`)

### Iteration 2 (revision for D1-D14)

- D1: AC-GCB-007/010 rebased on a named baseline (spec.md §C.3, the plan-auditor's observation at `114737ea1`, not re-measured by manager-spec; the run phase re-measures at M0). The two gtd-adjacent failures are stated as not caused by this SPEC. `-timeout` and slot lease added.
- D2/D5/D10/D13: gates moved into `check-residual.sh` and `check-gtd-body.sh`, with observed RED-now and mutant controls (acceptance.md ledger L1-L7).
- D3: mirror SKILL.md uses `.claude/skills/moai/workflows/gtd.md`.
- D4: `catalog.yaml` regeneration added to the change set and to AC-GCB-009.
- D6: CHANGELOG.md and top-level `reports/` added to the historical set (13 citation lines measured).
- D7/D14: counts measured as 19 surviving scoped files and 31 paths in the change set.
- D8: comment citations required, and the verdict-survivor escape removed.
- D9: exact test names used.
- D11: runtime and codemaps wording declared Out of Scope.
- D12: REQ-GCB-011 split into 011/012 (baseline became 013).
- Plan-audit iteration 2: PASS 0.86 (`.moai/reports/plan-audit/SPEC-GTD-CANON-BODY-001-review-2.md`)

### Iteration 2 pre-run amendments (no re-audit required)

- N1: both scripts refuse non-bash shells with exit 2 (ledger L10); AC-GCB-002 PASS-line floor = 113 (108 + 5 new set-equality checks, observed on control L6).
- N2: `002-flag-set-equal:<verb>` asserts that the frozen flag list equals the flags parsed from help, in both directions (L12 mutant goes red).
- N4: one slot lease per package, `--max-duration 45m` (30m timeout plus 15m margin). Exit 3 or 4 means wait and retry, at most 6 attempts, then record a Gap. Never run unleased, never `--force`.
- N5: an M0 timeout is a Gap; re-run once; a truncated baseline is never authoritative.
- N6: the sixth-stage regex covers six/sixth/6/6th and "answer … is a … stage" (L11 mutant goes red).
- N10: single evidence naming `.moai/reports/t867/<m0|m5>-<pkg>.{log,exit,fail-names}` in plan.md and acceptance.md.
- Extras taken: N3 (`[build failed]` / new package FAIL line counts as a new failure), N7 (inventory excludes this SPEC's own review files), N8 (`TestManifestHashFormat` added to AC-GCB-009).
- Not taken: N9 (line-based survivor rule accepts an unrelated `compat alias` on the same line; accepted as inherent).

### Scope change 0.3.0 (lead, card t854 handover)

- `TestGTDCanonicalSurfaceGolden` and `TestGTDAllTodoVerbsParity` are now owned by t867 (REQ-GCB-014, AC-GCB-012). They were removed from the baseline set, which keeps the five t854-owned tests.
- Measured at `e4cc628e9`, each test run alone:
  - L13: `gtd_canonical_surface_test.go:24 … is not a thin gtd compatibility path`. The literal the test requires predates the t860/t861 wording.
  - L14: `gtd_compat_test.go:61 gtd verbs = [.. answer ..], want [..no answer..]`. The `answer` verb landed in `1b644372d`; `moai todo` has no `answer` subcommand.
- The parity repair direction (A: `answer` stays gtd-only, test expectation only; B: expose `answer` on the todo alias) is an open operator decision, recorded in plan.md §B.1 and blocking M3.8.
- plan.md §D now requires absorbing local develop (plan-time `27220fb94`, t783 merge pending) before M0, and measuring the M0 baseline on the absorbed tree.
- This amendment changes the plan-artifact hash, so the cached iteration-2 PASS no longer satisfies the skip-eligibility hash condition; a Phase 1 re-audit follows.

### 0.3.1 — operator decision

- Kickoff approved by operator via lead 2026-09-18, autonomous progression.
- Parity repair: option A (`answer` gtd-only; `"answer"` added to `gtdWant` only; CLI unchanged). M3.8 unblocked.
- Added REQ-GCB-015 / AC-GCB-013: isolated `moai todo answer t1 x` observation (`MOAI_HOME` and `CLAUDE_PROJECT_DIR` set to temp dirs, real-queue sentinel check); a card-adding result is a finding and is not fixed.
- The M3.7 commit must cite `1dcaad954` (t860) and `61582178d` (t861), measured with `git log --oneline -- internal/template/templates/.claude/commands/moai/todo.md`.

### 0.3.2 — plan-audit iteration 3 (FAIL 0.75) revisions

- B1: `$BASE` anchor (`.moai/reports/t867/base.txt`, written at M0 after absorbing develop). Measured pre-absorption `git merge-base develop HEAD` = `f67d2193f`; develop tip = `a851b205c`.
- B2: registration-line diff guard. Controls: Long-help-only diff → grep exit 1; `AddCommand` diff → exit 0.
- B3: 1/1 numstat pins, the gtdWant line-equivalence check (good PASS; drop-engage and no-answer mutants FAIL), a literal check (verbatim PASS, loosened FAIL), and assertion counts on `4cc8ee74e`: canonical 3, compat 23. RED-now: numstat vs `f67d2193f` prints 0 lines.
- O1: `last_seq` sentinel. Plan-time read: exit 0, `last_seq` 870, `items` 143.
- O2: blocker contingency for failures of previously unreached assertions.
- O3: exact safe `answer` sentence mandated (proven by L6).

plan_audit: iteration 2 PASS 0.86 (pre-0.3.0 artifacts); iteration 3 FAIL 0.75 (0.3.1 artifacts) → revised in 0.3.2
plan_complete_at: 2026-09-18T01:45:23+09:00
plan_status: audit-ready

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

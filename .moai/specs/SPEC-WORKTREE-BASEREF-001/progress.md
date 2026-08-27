# SPEC-WORKTREE-BASEREF-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: CLOSED — plan-audit complete, operator kickoff approval granted, run-phase entry authorized
plan_complete_at: 2026-08-27
plan_audit_iter: 2 complete — verdict **PASS-WITH-DEBT 0.92 harmonic, 7/7 must-pass clear, zero blocking defects** (`.moai/reports/t313/plan-audit-iter2.md`). Iteration 1: FAIL 0.78 harmonic, 7/7 must-pass clear (`.moai/reports/t313/plan-audit-iter1.md`).
kickoff_approval: GRANTED by the operator, conditional on repairing debt item N1 first. N1 repaired in spec version 0.3.1 (below) before run-phase entry; the condition is discharged. Run-phase progression mode: **autonomous** — there is no milestone checkpoint at which a human reviews an intermediate reading, so criterion wording that admits a permissive interpretation was pinned rather than left open (see N2).
artifacts: spec.md (0.3.1, draft), plan.md, acceptance.md, progress.md — Tier M
tree: 48eb945df (worktree t313, branch WT-worktree-baseref)
counts: 16 GEARS requirements / 16 acceptance criteria (all MUST) — exactly at the Tier M ceiling of 16 REQ / 16 AC (`spec-workflow.md:148`). Unchanged by the 0.3.1 repairs: both are criterion-layer wording fixes, no id added, renamed, or renumbered.
iter2_repairs: D1 (REQ-WBR-004 now covered by AC-WBR-016, read-seam call count), D2 (vacuous `-run` pass mode closed in §D.3), D3 (AC-WBR-012 restated as REQ-WBR-015's three-part conjunction + guard-existence + mutation check), D4 (R1 premise corrected, consumer 1 narrowed to the primary checkout), A5 (REQ-WBR-011 bound to REQ-WBR-009's predicate as sole authority, AC-WBR-008 asserts the shared helper), plus debt D5-D10 folded in
iter2_changes: `.moai/reports/t313/spec-iter2-changes.md`
post_audit_repairs (spec 0.3.1): `.moai/reports/t313/spec-n1n2-changes.md`
  - **N1 (major, was blocking-carried)** — AC-WBR-013 check (1) probed only `internal/template/templates/$f`, so a correct implementation of plan §D emitted a false `NO-TEMPLATE-COUNTERPART` for `.moai/config/sections/git-strategy.yaml` (counterpart ships only as `.tmpl`; plan §B G5). Probe now accepts the plain OR the `.tmpl` form and reports only when neither changed. Re-measured against real history: old probe emits the false line on `664cd6eae^..664cd6eae`, new probe emits nothing; negative control `63b4628a6^..63b4628a6` (genuine orphan) still emits it, so detection was not weakened.
  - **N2 (minor-to-major, was blocking-carried)** — AC-WBR-016's "read seam" pinned to the **alignment-entry (configured-value) read**; the narrowing alternative was rejected because it would let the criterion lapse on the shipped empty default, which is the common case. Recorded in the criterion text and at plan.md §A D3.2.
debt_carried_into_run_phase (7 items, from the iter-2 verdict § Debt Carried Into Run-Phase):
  1. N1 — REPAIRED in 0.3.1 (no longer carried).
  2. N2 — REPAIRED in 0.3.1 (no longer carried).
  3. **Seam obligation is not optional (M2/M3)** — `internal/hook` carries no function-variable seam today; M2 must introduce one in its own new file. Budget for it; it is what makes AC-WBR-016 and AC-WBR-008's third assertion testable at all.
  4. **N4 — AC-WBR-013 check (2) is dormant, not correct.** Path mixing: the `[ -f ]` guard reconstructs a template-tree path while `diff -q` compares a sibling in `$b`'s own tree. Enumerates nothing for this SPEC (plan §D lists no hook wrapper); fix before any scope growth that touches one.
  5. **N5 — AC-WBR-002's first command has a prose-only pass condition.** Carry-over, not an iter-2 regression. Add an explicit value-emptiness assertion if run-phase wants it mechanical.
  6. **N3 — REQ-WBR-012 carries two GEARS modalities in one id.** Cosmetic; no run-phase action.
  7. **Folding traceability thinness (§ B1)** — REQ-WBR-004's narrowing and REQ-WBR-013's preservation clause are each carried by one AC half. Treat **both halves of AC-WBR-016 as separately mandatory**; an id-level matrix check will not notice a dropped half.
  Plus inherited and unchanged from iteration 1: **G2** (`EnterWorktree`'s `origin/HEAD` read is inferred, not read from source — the doctor item is the stated fallback), **G4** (`moai doctor` never executed), **G6** (rendered attribute order unmeasured — AC-WBR-011's two-branch condition MUST be collapsed to the single true form in run-phase).
blocker: none — D2 (widget shape) CLOSED by operator ruling 2026-08-27: `TypeText` with `main` / `develop` named in the field description; the new-combo-widget alternative rejected on template-neutrality grounds. Ruling + rationale recorded at plan.md §A D2.1; consequences at REQ-WBR-009 / REQ-WBR-012 / REQ-WBR-014, AC-WBR-009 / AC-WBR-011 / AC-WBR-015.

## §E.2 Run-phase Evidence

Full verdict with per-AC commands, exit codes, `=== RUN` counts, and evidence paths: `.moai/reports/t313/run-verdict.md`.
Verbatim command output: `.moai/state/verify/t313/` (persistent — survives `/tmp` clearance).
Measured in worktree `.claude/worktrees/t313`, branch `WT-worktree-baseref`, at HEAD `8c46460ff` unless a row says otherwise.

### AC matrix

| AC | Status | Verification command | Actual output | Evidence |
|---|---|---|---|---|
| AC-WBR-001 | PASS | `go test ./internal/config -run 'GitStrategy' -count=1 -v` | exit 0, 35 `=== RUN`, 0 vacuous | `ac-wbr-001.txt` |
| AC-WBR-002 | PASS | value-side grep + `go test ./internal/template -run 'WorktreeBaseBranchTemplate' -v` | branch-grep exit 1 (no match); guard exit 0, 1 `=== RUN` | `ac-wbr-002.txt`, `ac-wbr-002-guard.txt` |
| AC-WBR-003 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*(Unset\|Empty)' -count=1 -v` | exit 0, 1 `=== RUN`, 0 vacuous | `ac-wbr-003.txt` |
| AC-WBR-004 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Match' -count=1 -v` | exit 0, 1 `=== RUN`, 0 vacuous | `ac-wbr-004.txt` |
| AC-WBR-005 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Mismatch' -count=1 -v` | exit 0, 1 `=== RUN`, 0 vacuous | `ac-wbr-005.txt` |
| AC-WBR-006 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*(FailOpen\|GitError)' -count=1 -v` | exit 0, 2 `=== RUN`, 0 vacuous | `ac-wbr-006.txt` |
| AC-WBR-007 | PASS | `go test ./internal/cli -run 'SessionWorktree.*Base' -count=1 -v -timeout 600s` | exit 0, 5 `=== RUN`, 0 vacuous | `ac-wbr-007.txt` |
| AC-WBR-008 | PASS | `... -run 'SessionWorktree.*(NoBase\|Unresolvable)'` + `... -run 'SessionWorktree.*(SharedPredicate\|Resolver)'` | exit 0/0, 3 + 2 `=== RUN`, 0 vacuous | `ac-wbr-008a.txt`, `ac-wbr-008b.txt` |
| AC-WBR-009 | PASS | `go test ./internal/cli -run 'Doctor.*WorktreeBaseBranch' -v` + `./bin/moai doctor --check 'Worktree Base Branch'` | exit 0, 7 `=== RUN`; CLI exit 0, item reachable by exact name | `ac-wbr-009.txt`, `ac-wbr-009-doctor-run.txt` |
| AC-WBR-010 | PASS | `go test ./internal/web -run 'WorktreeBaseBranch' -v` + `go test ./internal/settings -run 'AllFields' -v` | exit 0/0, 2 + 1 `=== RUN`, 0 vacuous | `ac-wbr-010-web.txt`, `ac-wbr-010-allfields.txt` |
| AC-WBR-011 | PASS | `go test ./internal/settings -run 'WorktreeBaseBranch.*Type' -v` + `go test ./internal/web -run 'WorktreeBaseBranch.*(Text\|FreeText)' -v` | exit 0/0, 1 + 1 `=== RUN`, 0 vacuous | `ac-wbr-011-schema.txt`, `ac-wbr-011-render.txt` |
| AC-WBR-012 | PASS | `go test ./internal/web ./internal/hook ./internal/cli -run 'WorktreeBaseBranch' -v` + mutation check | exit 0, 21 `=== RUN`, 0 vacuous; mutant (FieldDef removed) exit 1 | `ac-wbr-012.txt`, `ac-wbr-012-mutation.txt` |
| AC-WBR-013 | PASS | `make build` + diff-scoped parity probes (N1-repaired form) | exit 0; 0 `NO-TEMPLATE-COUNTERPART`, 0 `DRIFT`; mirror byte-identical | `ac-wbr-013-parity.txt`, `make-build-final.txt` |
| AC-WBR-014 | PASS | `go test ./internal/settings -run 'GitStrategy.*RoundTrip' -count=1 -v` | exit 0, 2 `=== RUN`, 0 vacuous | `ac-wbr-014.txt` |
| AC-WBR-015 | PASS | `go test ./internal/hook -run 'WorktreeBaseBranch.*Unresolvable' -count=1 -v` | exit 0, 2 `=== RUN`, 0 vacuous | `ac-wbr-015.txt` |
| AC-WBR-016 | PASS (both halves) | `go test ./internal/hook -run 'WorktreeBaseBranch.*(Fires\|Registered\|Once\|LinkedWorktree\|NotPrimary)' -v` | exit 0, 4 `=== RUN`, 0 vacuous | `ac-wbr-016.txt` |

### Invariants

| Invariant | Status | Evidence |
|---|---|---|
| `go build ./...` clean | PASS (exit 0) | `go-build.txt` |
| Cross-platform build (windows/amd64, linux/amd64) | PASS (exit 0, 0) | `go-build.txt` |
| `go vet` clean on the five affected packages | PASS (exit 0) | `go-vet.txt` |
| `golangci-lint run` clean on the five affected packages | PASS (exit 0, `0 issues.`) | `golangci-lint.txt` |
| No file under `.claude/skills/moai-workflow-project/schemas/` modified (t316 boundary) | PASS (diff empty) | recorded in run-verdict |
| Full-package regression: config / settings / web / hook / cli | PASS (all `ok`) | `coverage.txt`, `coverage-cli.txt` |

### Scope note — one file outside the plan §D write list

`internal/config/types.go` `ModeProfile` gained three pass-through fields (`develop_branch`, `release_branch_prefix`, `rc_version_format`). spec.md §C lists the ModeProfile schema gap as out of scope, and REQ-WBR-013 requires this SPEC's write path not to drop those keys — the two are only jointly satisfiable by modelling them, because `saveSection` re-marshals the struct with no merge step (read at `internal/config/manager.go:418`). MEASURED before the change: the typed save deleted all three (`m5-settings-1.txt`). No accessor and no consumer was added; the fields exist only to survive the round trip. Recorded here rather than absorbed silently.

Three further cascade edits, all inside the SPEC's own envelope: the shipped-key triage inventory (a new shipped template key must be classified), the free-text whitelist guard (a new `TypeText` field must be justified), and the doctor golden snapshots (a new diagnostic row).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: pending-backfill-run-final   # M6 docs commit 8c46460ff; the evidence commit that carries this block cannot name its own hash
run_status: COMPLETE — 16/16 AC PASS, zero blockers
ac_pass_count: 16
ac_fail_count: 0
ac_pass_with_debt_count: 0
milestone_commits:
  M1: 81808d85b   # config schema + neutral template default
  M2: cf2955ed5   # SessionStart origin/HEAD alignment
  M3: 9e1ea4226   # git worktree add base operand
  M4: 5658988be   # Worktree Base Branch doctor diagnostic
  M5: 04c645a68   # web free-text field + guards + round-trip preservation
  M6: 8c46460ff   # worktree rule documentation + mirror
preserve_list_post_run_count: 0 unrelated files modified
l44_pre_commit_fetch: HEAD + branch re-read immediately before each of the 6 commits; unchanged each time (d0bc4bba5 → 8c46460ff, linear)
l44_post_push_fetch: n/a — NOT pushed; integration is the orchestrator's call per the dispatch
new_warnings_or_lints_introduced: 0 (golangci-lint 0 issues; the 2 errcheck findings this run introduced were fixed before the final measurement)
cross_platform_build:
  darwin_native: exit 0
  windows_amd64: exit 0
  linux_amd64: exit 0
coverage:
  internal/config: 80.6%
  internal/hook: 85.1%
  internal/settings: 90.3%
  internal/cli: 79.6%
  internal/web: 66.8%
  note: package-level pre-existing baselines, not a per-change figure; no baseline was measured before this SPEC, so no delta is claimed
total_run_phase_files: 21 changed (11 production, 7 test, 3 golden snapshots) across 6 commits
m1_to_mN_commit_strategy: one commit per milestone, each naming card t313 in its body
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

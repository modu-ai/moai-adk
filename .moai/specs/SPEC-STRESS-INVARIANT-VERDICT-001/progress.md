# SPEC-STRESS-INVARIANT-VERDICT-001 — Progress

Card t372 · worktree `.claude/worktrees/t372` · branch `WT-stress-invariant-guard` ·
base `origin/develop` = `b9149857c`.

## §E.1 Plan-phase Audit-Ready Signal

### Iteration 1 — audited FAIL 0.69 (`.moai/reports/t372/plan-audit.md`)

- Artifacts written: `spec.md`, `plan.md`, `acceptance.md`, `progress.md`. Declared `tier: S`.
- Requirements REQ-SIV-001 .. REQ-SIV-017 (17); acceptance AC-SIV-001 .. AC-SIV-013 (13).
- All 7 must-pass checks passed; the FAIL was score-driven (Clarity 0.65, Completeness 0.75,
  Testability 0.55, Traceability 0.90).

### Iteration 2 — fix round (this revision, `version: 0.2.0`)

Tier reclassified **S → M**. Justification (`spec.md` § Tier classification): the artifact carries
16 REQ / 14 AC, which fits Tier M's 16/16 ceiling and not Tier S's 8/8; tiering up rather than
splitting is the rule's stated response to a budget breach, and the higher threshold (0.80) plus the
2-iteration ceiling are accepted deliberately on a card whose hazard is that concealment resembles
repair. The Tier M artifact set is exactly what exists, so `acceptance.md` is no longer a deviation.

Defects answered, all against source read on tree `9d4f79281`:

| Defect | Change |
|---|---|
| D1 | AC-SIV-009's duplicate-id branch deleted, with its exclusion reason (`backlog_sqlite.go` `id … UNIQUE`) recorded. RED must now originate at a named assertion **inside** the invariant block, citing message + source line. Four non-discharging RED sources enumerated. |
| D2 | REQ-SIV-008 added: `successes + starved + hardFailures == stressWriters * stressAddsPerWriter`. AC-SIV-014 added. Machine-independent by construction; fractional floors explicitly forbidden. |
| D3 | AC-SIV-009 branch (c) restated as a `last_seq` advance **above** the item count, with the downward direction excluded and `normalizeBacklogRecord`'s erasure recorded. |
| D4 | Tier S → M (above). REQ layer consolidated 17 → 16 by four merges; AC-SIV-008 / AC-SIV-009 survive intact as separate binding criteria. |
| D5 | REQ-SIV-009 reworded *covers* → *coherent with … at the declared per-mutation cost*; the guard's own messages must state the same and claim no sufficiency. |
| D6 | REQ-SIV-014 clause 4 added: the Unix sentinel admits every `unix.Flock` failure, `errors.Is` traverses `errors.Join`. Sentinel narrowing recorded as an out-of-scope follow-up candidate. |
| D7 | AC-SIV-001 / AC-SIV-005 given a deterministic seeded-holder construction (`acquireBoardLockImpl` + `t.Cleanup`, 1-2 adds) and a named verification verb. M2 steps 4-5 carry the implementation shape. |
| D8 | AC-SIV-013 rationale + REQ-SIV-014 clause 3: a green window evidences only that no new failure mode was introduced, because the invariant criterion was already red in 0 of 14 runs. |
| D9 | REQ-SIV-014 clause 2 corrected: one green `Race Test` job (`51daada00`) and one run green inside a job reddened by another test (`c6aa61346`). |
| D10 | REQ-SIV-004 relabelled `(Where)`; the old REQ-SIV-002's second obligation split into its own `(Unwanted)` requirement. |
| D11 | REQ-SIV-016 added for scope discipline; AC-SIV-012 now traces to it rather than to a section. |
| D12 | `plan.md` §B states the 4.2% margin and the catch-set table, plus the cost-independent reduction `supportedWriters * headroom >= stressWriters * stressAddsPerWriter`. |
| D13 | AC-SIV-008 requires the old guard's GREEN under the same mutant, `-v` output evidencing a non-zero selector match, and the RED naming which test failed. |

Operator observation folded in: REQ-SIV-010 forbids rebuilding the guard's floor from the budget's
own inputs, citing the pre-existing unreachable `budget < floor` branch in
`TestBoardLockWaitBudgetDerivedFromNamedInputs` as the shape not to reproduce. Repairing that branch
is recorded as out of scope.

- SPEC ID regex check executed as Bash: `PASS`.
- Requirements: REQ-SIV-001 .. REQ-SIV-016 (16), GEARS notation, no residual `IF/THEN`.
- Acceptance: AC-SIV-001 .. AC-SIV-014 (14); AC-SIV-008 / AC-SIV-009 remain the binding
  mutant-evidence pair (both directions required).
- Ground truth consumed, not re-measured: `.moai/reports/t370/verdict.md`,
  `.moai/reports/t370/measurements.md`.
- No implementation code written at plan-phase. No push, no PR, no CI run, no load generated.

## §E.2 Run-phase Evidence

Full evidence, with every command and its verbatim output, lives at
`.moai/reports/t372/run-evidence.md`. Measured in this run, against this tree, run-phase start
HEAD `3cd1a09f1`. The t370 measurements are consumed as ground truth and re-measured nowhere.

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-SIV-001 | PASS | `go test -race -count=1 -v -run '…' ./internal/kanban/` | `--- PASS: TestStressAddClassificationToleratesStarvation (3.37s)` — `2/2 adds starved under a seeded holder, all satisfying IsBoardLockHeld, 0 hard failures` |
| AC-SIV-002 | PASS | `grep -n 'strings.Contains' internal/kanban/backlog_concurrency_test.go` | one hit, line 343, inside the AC-SIV-005 sub-test's **message** assertion; the classification path (`classifyStressAdd`) decides on `err == nil` / `IsBoardLockHeld(err)` only |
| AC-SIV-003 | PASS | source read `backlog_concurrency_test.go:205-231` | four assertions anchored to `issuedCount := len(issued)`; `wantTotal` no longer exists in the file |
| AC-SIV-004 | PASS | `grep -n 't.Skip' internal/kanban/backlog_concurrency_test.go` | one hit, line 202, in a **comment**; no `t.Skip` call, no starvation conditional |
| AC-SIV-005 | PASS | `go test -race -count=1 -v -run '…' ./internal/kanban/` | `--- PASS: TestStressZeroProgressFloorFailsTotalStarvation (1.68s)` — zero-success outcome rejected, 1-success outcome admitted |
| AC-SIV-014 | PASS | source read + `go test -race -count=1 ./internal/kanban/` | `successes + starved + len(hardFailures) == stressWriters * stressAddsPerWriter`; no clock, no fraction, no percentage |
| AC-SIV-006 | PASS | `go test -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/` | `--- PASS: TestBoardLockWaitBudgetCoversSerializedMutations (0.00s)`; message claims coherence only |
| AC-SIV-007 | PASS | `sed -n '95,120p' internal/kanban/board_lock_wait_test.go \| grep -nE 'time\.(Now\|Since\|Sleep)\|go func'` | **no output** — constants only; floor built from the two stress constants |
| **AC-SIV-008** | **PASS** | `go test -count=1 -v -run 'TestBoardLockWaitBudget' ./internal/kanban/` under `boardLockHeadroom 5→4` | new guard `--- FAIL` (`1.32s budget < 1.584s floor`), old guard `--- PASS` same run; selector matched 2; whole-package run fired exactly 1 test; reverted → both `--- PASS` |
| **AC-SIV-009** | **PASS** | `go test -race -count=1 -v -run TestConcurrencyStress ./internal/kanban/` under two reverted invariant mutants | mutant 1 (upward `last_seq`): RED at **invariant (d) mark consistency**, `backlog_concurrency_test.go:228`, `last_seq = 56, want 48`. mutant 2 (dropped item): RED at **(b)** line 218 and **(c)** line 223, in a run that also tolerated 1 real starved add. Selector matched 1 each; whole-package run fired exactly 1 test; both reverted → `--- PASS` |
| AC-SIV-010 | PASS | the `TestConcurrencyStress` `t.Logf` line | starved count + back-derived per-mutation cost logged; no verdict gated on either |
| AC-SIV-011 | PASS | `.moai/reports/t372/run-evidence.md` §8 | all four REQ-SIV-014 limits stated |
| AC-SIV-012 | PASS | `git diff --stat HEAD` (taken after both mutants reverted) | 3 files: `backlog_concurrency_test.go`, `board_lock_wait_test.go`, `board_store.go` (comment-only, verified by a non-comment-line grep returning empty) |
| AC-SIV-013 | **OPEN at merge** | — | needs ≥5 post-landing develop heads; deliberately unclaimed |

Invariants (PRESERVE list, post-run): `board_lock_unix.go`, `board_lock_windows.go`,
`board_lock.go`, `backlog_store.go` — `git diff --stat` against base returns **empty** for all four.
`boardLockCIMutationCost`, `boardLockHeadroom`, `boardLockSupportedWriters`,
`boardLockWaitMin/Max/Step`, `boardLockRetryWait`: unchanged in value and behaviour.

This run's own figures (local darwin `-race`, observability only, not a verdict input):
**0 starved of 48**, back-derived per-mutation cost **15.696085ms**.

Non-claims carried (REQ-SIV-014, full text in the evidence file §8): no before/after comparison
exists in any quantity; a single green run cannot close the card; a green observation window
evidences only that no new failure mode was introduced; the tolerated class is every `unix.Flock`
failure on Unix, and `errors.Is` traverses `errors.Join`.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-31
run_commit_sha: pending-backfill-run
run_status: audit-ready
ac_pass_count: 13
ac_fail_count: 0
ac_open_count: 1          # AC-SIV-013, closure-gate, open at merge by design
preserve_list_post_run_count: 4   # board_lock_unix.go, board_lock_windows.go, board_lock.go, backlog_store.go — all diff-empty
l44_pre_commit_fetch: not-run     # lane-local worktree; the lead holds the integration window
l44_post_push_fetch: not-run      # no push performed this phase
new_warnings_or_lints_introduced: 0   # go vet ./internal/kanban/... clean; gofmt -l clean; golangci-lint NOT run (gap)
cross_platform_build:
  darwin_arm64: pass              # go test -race -count=1 ./internal/kanban/ → ok 24.890s
  linux_amd64: not-run
  windows_amd64: not-run
total_run_phase_files: 3
m1_to_mN_commit_strategy: single-commit   # M1-M5 land as one commit; the milestones are one indivisible verdict-criterion change
```

## §E.4 Sync-phase Audit-Ready Signal

Full sync verdict, in the five evidence-bearing sections with per-measurement party attribution
(`[MD]` = manager-develop run-phase, `[ORCH]` = orchestrator independent re-run):
`.moai/reports/t372/verdict.md`.

```yaml
sync_complete_at: 2026-08-31
sync_commit_sha: 0fa8606fe   # backfilled 2026-09-02 by lane-7 — the sync commit "test(SPEC-STRESS-INVARIANT-VERDICT-001): invariant-anchored verdict for TestConcurrencyStress (t372)"; verified an ancestor of origin/develop via git merge-base --is-ancestor
sync_status: audit-ready
b12_self_test_a: pass          # grep -c 'SPEC-STRESS-INVARIANT-VERDICT-001' CHANGELOG.md -> 0, no duplicate
b12_self_test_b: pass          # acceptance.md distinct AC identifiers -> 14 (non-zero), matches the 14-row matrix
b12_self_test_c: pass          # every path named in the CHANGELOG entry verified present
changelog_entry_position: "[Unreleased] / ### Changed (top)"
changelog_section_rationale: >
  Changed, not Fixed. Fixed would assert the CI flake is fixed; it is not established as fixed
  (AC-SIV-013 open, and no before/after comparison exists in any quantity). What changed is the
  test's verdict criterion.
ac_pass_count: 13
ac_fail_count: 0
ac_open_count: 1               # AC-SIV-013, closure gate, OPEN at merge by design
frontmatter_status_transitions:
  spec_md: "in-progress -> implemented -> completed"   # completed 2026-09-02: AC-SIV-013 closure gate resolved — see the closure block below
  plan_md: "n/a - no frontmatter (status-stateless per spec-frontmatter-schema.md)"
  acceptance_md: "n/a - no frontmatter (status-stateless)"
  progress_md: "n/a - phase state recorded in body sections"
mx_tag_validation: "sync sub-step, no change - test-only diff plus one comment block; no exported production surface added or changed"
independent_orchestrator_verification:
  ac: AC-SIV-008
  party: orchestrator
  tree: 0fa8606fe
  census_command: "grep -rln 'boardLockHeadroom\\|boardLockWaitBudget' --include='*_test.go' internal/"
  census_files: 3
  mutant: "boardLockHeadroom 5 -> 4 (constant axis)"
  scope: whole-package            # deliberately wider than the run-phase selector
  swept_count: 389                # '=== RUN' lines; non-zero, so no empty-sweep masquerade
  failing_tests: 1
  failing_test_name: TestBoardLockWaitBudgetCoversSerializedMutations
  old_guard_same_run: PASS        # TestBoardLockWaitBudgetDerivedFromNamedInputs -> the attribution
  reverted: true
  restoring_green: true
  tree_clean_after: true
  evidence_path: .moai/reports/t372/mutant-headroom4-orchestrator.log
  evidence_path_committed: false  # gitignored (.gitignore:106 *.log); decisive lines quoted verbatim in verdict.md 2.2
readme_docs_site_change: none
readme_docs_site_check: >
  grep -rln 'TestConcurrencyStress|boardLockWaitBudget|boardLockHeadroom' README*.md docs-site/
  -> no hits. Checked, not assumed.
template_mirror_required: false  # no path under internal/template/templates/ touched
golangci_lint_run: false         # gap, carried from run-phase; go vet + gofmt clean only
ci_read_at_sync: false           # no CI run read or triggered; out of delegated scope
l44_pre_commit_fetch: not-run    # lane-local worktree; the lead holds the integration window
l44_post_push_fetch: not-run     # no push performed this phase
closure_gate_open: AC-SIV-013    # >=5 non-cancelled post-landing develop heads; card stays implemented
```

Non-claims carried (REQ-SIV-014, verbatim in substance at `.moai/reports/t372/verdict.md` §5): no
before/after improvement claim exists in any quantity; a single green run cannot close this card
(two post-repair green observations already existed — this is where t354 stopped); a green
observation window evidences only that no new failure mode was introduced, never that the
invariants still fire; the tolerated error class is every `unix.Flock` failure on Unix, not only
`EWOULDBLOCK`, and `errors.Is` traverses `errors.Join`. Local green is not the verdict — both local
per-mutation figures (14.8 ms `[ORCH]`, 17.5 ms t370) sit far below the 34.4 ms threshold, so this
machine was always in the passing band, and the 42-105 ms CI `-race` band was never reproduced
locally.

Follow-up candidates recorded, not fixed (verdict.md §7): the pre-existing unreachable
`budget < floor` branch in `TestBoardLockWaitBudgetDerivedFromNamedInputs`
(`board_lock_wait_test.go:36-41`); the over-broad `ErrBoardLockHeld` sentinel on Unix; the
`…CoversSerializedMutations` guard name prescribed verbatim by `plan.md` while REQ-SIV-009 forbids
"covers" framing in its messages; and `.moai/reports/t370/**` being untracked everywhere while
serving as this SPEC's cited ground truth.

### Closure gate resolved — AC-SIV-013 (2026-09-02, card t372 status → completed)

The binding evidence is the **CI read by lead-1 on 2026-09-02**, not any local run (the local
targeted run lane-7 recorded in `.moai/reports/t372/verdict-2026-09-02.md` E4 is corroborating
only — the machine sits in the passing band by construction):

| Element | Value | How established |
|---|---|---|
| CI run | `33564147725` | `gh run list --branch develop` (lead-1, 2026-09-02) |
| Develop head measured | `09bf452c0` | same read; the head descends from the landing commit `0fa8606fe` (`git merge-base --is-ancestor`, lane-7 re-measured) |
| `run_attempt` | 1 | `gh api repos/.../runs/33564147725` direct read — no retry-masked first failure |
| `Race Test` job | success | job list of the same run |
| Non-vacuous execution | kanban package actually ran | log carries `SKIPPED TEST github.com/modu-ai/moai-adk/internal/kanban TestLandedCheck_Controls` — exactly one skipped test, and `TestConcurrencyStress` is not in the skip list; the pass-name absence is a `-json` non-`-v` artifact, not evidence of non-execution |
| Reader | lead-1 (the lead's CI read is the lane-assigned channel — lanes do not read CI) | reported to lane-7 2026-09-02 |

AC-SIV-013's stated condition — green under the invariant criterion on ≥5 non-cancelled develop
heads descended from the landing commit — is satisfied by this read on the current head
(`0fa8606fe` had 329 descendant commits on `origin/develop` when lane-7 measured, so the
head-count condition is met a fortiori). The gate's own stated limit stands: this window
evidences no *new* failure mode; the invariants' continued firing remains AC-SIV-009's
discharge, which is where it was left. Card t372 transitions `implemented → completed` on this
record.

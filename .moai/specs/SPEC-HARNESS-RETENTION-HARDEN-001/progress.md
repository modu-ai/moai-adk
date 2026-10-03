# SPEC-HARNESS-RETENTION-HARDEN-001 — Progress

Card t1432. Branch `WT-harness-retention-debt`, base develop `1e2151a38`.

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-03
plan_iteration: 3 (revision of version 0.2.0 after plan-audit iteration 2 FAIL 0.87, `.moai/reports/t1432/plan-audit-iter2.md`; iteration 1 was FAIL 0.75, `.moai/reports/t1432/plan-audit.md`)
plan_artifacts: spec.md, plan.md, acceptance.md, decision-index.md, progress.md
relayed_verdicts: recorded in `decision-index.md` (relayed by the leader session on 2026-10-03, not a direct operator answer)
open_for_leader_before_kickoff: operator-held options D1-C, D2-B/C, D3-A, D4.a-B (non-default, unselected); decision-index rows Q7 and Q8 (warning surface, allocation of the one allowed test-only field), where the plan applies the `spec.md` §B defaults; plan.md B5 (regenerate the overlay JSON files in `red-now-drafts/` and force-add the B3 probe files), B6 (the red M0 commit, accepted), B8 (force-added evidence under an ignored path, accepted) and B9 (the pre-existing FIFO hang is recorded, not repaired; a repair needs a SPEC amendment first)

Record kept here rather than in source (REQ-HRH-011): the earlier lock-behaviour design question (`.moai/reports/t1425/decision-records.md`, `lock_behavior=block_then_recheck`) did not carry the fact that the harness-observe hooks run with a 5 s timeout and `async: true`; the plan-phase facts are in `spec.md` §A (F6 row).

Plan-audit iteration 2 optional findings left open in revision 0.3.0 (taken: O1, O2, O3, O5, O6, O7, O10; the blocking B1-B3 are fixed):

- O4 (AC-002 identity assertion can be falsely red on a file system that reuses a freed inode number): inferred, APFS observed not to reuse, Linux unobserved; the fix is a test detail that cannot be observed before the test exists, and the failure direction is a false red at CI, not a missed defect.
- O8 (AC-007 checks `late-event` by substring, so a mutant that re-encodes a late line passes): needs a changed test shape and a re-derived E-003; beyond this last pass.
- O9 (AC-012 does not pin the phrase naming the missing archive step): judged by reading; pinning a phrase needs the N1 fix to exist first.

## §E.2 Run-phase Evidence

Recorded at sync from `.moai/reports/t1432/verdict.md` (run-phase verdict, commit `668900855`), `.moai/reports/t1432/red-baseline.md` and `.moai/reports/t1432/run-evidence-m1-m3.md`. Platform of every run-phase measurement: darwin arm64, uid 501, go1.26.8. Linux and Windows runtime were not observed. Per-section commands, verbatim outputs and exit codes are in `verdict.md` section 2; this section carries the pointers and the sync-phase re-measurements.

Run commits and the intermediate red commit T, one line: intermediate red commit T=`ac40cf3bf` (M0), M1 `c1cc3fe67`, M2 `f52dd1b1c`, M3 `b3a469eab`, M4 `3badf7875`, M5 `9289a92b6`, verdict `668900855`; card base `1e2151a38`.

Run-phase result, from the verdict: all 14 acceptance criteria met on the final tree `9289a92b6`; `go test -race -count=1 -v ./internal/harness/` 277 `--- PASS`, 0 `--- FAIL`, exit 0 (verdict 2.4); all 15 mutants of `plan.md` §F M1, M2, N2 and N1 fail their named tests (verdict 2.7a); `GOOS=windows go build` and `go vet` of `internal/harness` and `internal/lockfile` exit 0 (verdict 2.5, build and vet only).

### M0 Exit ordering check

The tracked red-baseline test files precede every card commit that touches `internal/harness/retention.go`. Measured on the card branch at `9289a92b6` in the run phase (verdict 2.7), commands and exit codes:

| Command | Output | exit |
|---|---|---|
| `git log --diff-filter=A --format=%h -- internal/harness/retention_statepath_test.go` | `ac40cf3bf` | 0 |
| `git log --reverse --format=%h 1e2151a38..HEAD -- internal/harness/retention.go` | `c1cc3fe67` `f52dd1b1c` `b3a469eab` `9289a92b6` | 0 |
| `git merge-base --is-ancestor ac40cf3bf c1cc3fe67` (likewise `f52dd1b1c`, `b3a469eab`, `9289a92b6`) | (empty) | 0 each |
| control, mis-ordered: `git merge-base --is-ancestor c1cc3fe67 ac40cf3bf` | (empty) | 1 |
| control, bounded form on a range with a change: `git log --reverse --format=%h fe211e9c9~1..fe211e9c9 -- internal/harness/retention.go` | `fe211e9c9` | 0 |

Boundary re-measured at sync (tree `668900855`, this run): `git diff --quiet 1e2151a38 HEAD -- internal/lockfile internal/harness/observer.go` exit 0 (empty). The four pre-lock function bodies (`PruneStaleEntries`, `readStamp`, `readStampFile`, `stampIsFresh`) have no commit in `ac40cf3bf..HEAD` (verdict 2.6, run phase, with its positive controls); not re-run at sync.

### Deviations between the SPEC text and what was built

1. M2 strengthened the committed `TestPruneStateUnreplaceableInReadOnlyDirSkips` with one assertion: the returned error must satisfy `errors.Is(err, fs.ErrPermission)`. The stated assertions of AC-HRH-003 case (b) (a non-nil error, byte-identical state) let the mutant "a heal that ignores the removal failure" survive, so the assertion was added; the unmodified code passes it and mutant 12 fails it. The strengthened test was not re-run against base code.
2. M1 gate: the two M0 REDs (AC-HRH-001, AC-HRH-002) belong to M2, so the M1 gate was the package minus those two named tests (`run-evidence-m1-m3.md`).
3. M4 mechanism: the criterion binds the outcome (fail within 15 s naming the missing archive step), not the mechanism; the blocking read open runs in a goroutine and is released after a 10 s wait by a non-blocking write open of the FIFO. The non-blocking read open was not used because it can race a pruner that has not reached its write open.
4. Mutant count: the task text said 14, `plan.md` §F lists 15 (3 in M1, 9 in M2, mB, mC, mD); all 15 were run.

### What CI on Linux exercises first

Every run-phase run was on darwin. CI on `ubuntu-latest` is the first execution of: the read-only-handle truncate seam (AC-HRH-014), the FIFO tests, the stderr capture, the `O_EXCL` create, the owner stat on Linux, and the AC-HRH-005 (e) candidate list (whether a foreign-owned symbolic link exists among `/var`, `/tmp`, `/etc`, `/bin` on the runner; the test skips where none qualifies). A difference there is not observed and would show first in CI.

### Carried audit debt (accepted, listed in `.moai/reports/t1432/decision-records.md` D-1 and `plan-audit-iter3.md`)

N1 (DoD-1 wording), N2 (stale "untracked" note in G-7 and `plan.md` B5), N3 (a statement that CI never compiles Windows tests; `ci.yml` runs `GOOS=windows go vet ./...`), N4 (AC-HRH-005 (e) skip granularity), N5 (no end-to-end case for a user-owned link to a root-owned target), N6 (mutant copies in `red-now-drafts/` are base-bound; discharged in the run by regenerating every mutant from the current files), O4 (AC-HRH-002 identity assertion could be falsely red on a file system that reuses a freed inode number; inferred, APFS observed not to reuse, Linux unobserved), O8 (AC-HRH-007 compares `late-event` by substring, so a re-encoding mutant passes), O9 (AC-HRH-012 does not pin the phrase naming the missing archive step). The SPEC body was not edited for any of them: an edit would change the plan-artifact hash that the autonomous Kickoff decision rests on.

### `moai spec lint` and its judging build

Run phase (verdict 2.8): `moai spec lint SPEC-HARNESS-RETENTION-HARDEN-001` printed `✓ No findings — all SPEC documents are valid`, exit 0, judged by a binary built with `go build ./cmd/moai` from the tree at `9289a92b6` (its embedded VCS stamp named an unrelated commit because the tree is a linked worktree, so the identity rests on the build procedure).

Sync phase, before the sync commit (this run, tree `668900855`, clean status): same command, output `✓ No findings — all SPEC documents are valid`, exit 0. Judging build: the installed `moai` at `/Users/goos/go/bin/moai`, `v3.2.0-rc.26`, `archive/t1401-293-g45600e4ee`, built 2026-10-02T14:42:32Z; `git merge-base --is-ancestor 45600e4ee HEAD` exit 0, so it is a strict ancestor of the tree HEAD and lags it. `git diff --stat 45600e4ee HEAD -- internal/spec` printed nothing (exit 0), so the SPEC lint package is unchanged between the two; the cmd and other packages the binary links were not compared. The post-edit lint and `moai spec audit` runs are recorded in §E.4.

### Test-only field, unselected options, boundary

The one allowed unexported test-only field on the pruner (`ownerCheck`) was spent on the owner lookup; the stamp-write pin uses a parameter seam (`pruneLocked`) and needs no field. The operator-held options D1-C, D2-B, D2-C and D3-A were never selected. `internal/lockfile` and `internal/harness/observer.go` are byte-identical to `1e2151a38` (re-measured above).

### Gaps (not observed)

Windows runtime not observed; Linux unobserved. The heal burst, the residual loss window between the final tail reading and the rename, and whether a concurrent append can be seen half-complete are disclosed, not measured. The pre-existing FIFO-at-the-state-path hang is recorded, not repaired (`spec.md` §A, §E, §F). Reviewer-read parts: the F5 and F6 source sentences against REQ-HRH-010 and REQ-HRH-011, and the phrase naming the missing archive step. Coverage was not re-measured in the run. The run and sync phases pushed nothing and merged nothing; the leader session does that.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-03
run_status: audit-ready
cycle_type: tdd
head_at_signal: 668900855   # the run-phase verdict commit; the sync commit follows it
commits:
  m0_intermediate_red_T: ac40cf3bf
  m1: c1cc3fe67
  m2: f52dd1b1c
  m3: b3a469eab
  m4: 3badf7875
  m5: 9289a92b6
  run_verdict: 668900855
acceptance_criteria: {total: 14, pass: 14, skipped: 0}   # per verdict.md section 3, with the qualifications in its Gaps
mutants: {run: 15, killed: 15}   # verdict.md section 2.7a
evidence:
  - .moai/reports/t1432/red-baseline.md
  - .moai/reports/t1432/run-evidence-m1-m3.md
  - .moai/reports/t1432/verdict.md
open_gaps:
  - Windows runtime not observed (GOOS=windows build and vet only)
  - Linux unobserved; CI on ubuntu-latest is the first run of the FIFO tests, the read-only-handle truncate seam, the O_EXCL create and the owner stat
  - heal burst, residual loss window and half-complete concurrent append not measured
  - pre-existing FIFO-at-the-state-path hang recorded, not repaired
  - carried plan-audit debt N1-N6, O4, O8, O9 (accepted)
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-10-03
sync_status: audit-ready
sync_commit_sha: 5bb35abe8165e8d5fc5cbab6b246ac6a82646cb8   # backfilled after the sync commit (5bb35abe8); a commit cannot cite its own hash
changelog_path: CHANGELOG.md   # [Unreleased] ### Fixed, first entry
ac_source: .moai/specs/SPEC-HARNESS-RETENTION-HARDEN-001/acceptance.md   # tier M
ac_count: 14   # live AC-HRH identifiers in acceptance.md (14 distinct, no [RETIRED] or [REF] marker present)
req_count: 15   # distinct REQ-HRH identifiers in spec.md
b12_self_test_a: "pre-emission grep -c 'SPEC-HARNESS-RETENTION-HARDEN-001' CHANGELOG.md printed 0 (exit 1) before the entry was added"
b12_self_test_b: "AC count match: the CHANGELOG entry cites 14 acceptance criteria and 15 requirements; grep -o -E 'AC-HRH-[0-9]+[a-z]?' on acceptance.md gives 14 distinct identifiers, REQ-HRH 15 on spec.md; the awk counter of manager-docs.md B12 was not run (substituted by grep, see open_gaps)"
b12_self_test_c: "every file path named in the CHANGELOG entry was checked with ls before the commit"
changelog_entry_position: first entry under [Unreleased] ### Fixed
frontmatter_status_transitions:
  draft_to_in_progress: "recorded after the run commits, not at M1: spec.md stayed status draft through every run commit (git log -G status: listed only 69ddcc73a), so the transition was committed on its own as 1b5c3c057 once the lifecycle lint reported StatusTransitionInvalid on the first sync commit (draft to completed); the first sync commit 390c9b505 was undone by a soft reset and recreated after it"
  in_progress_to_implemented_to_completed: "one step, on the sync commit that follows 1b5c3c057: status in-progress -> completed; only that status line of spec.md changes in it; plan.md, acceptance.md and progress.md carry no status field (stateless artifacts)"
docs_change: none   # README (4 locales) and docs-site: the only hits are a file-name mention in an artifact list, docs-site/content/{en,ko,zh}/workflow-commands/moai-harness.md line 170 and the ja page line 169; no page describes the prune interval, the state file or the retention behaviour. No docs change.
spec_lint:
  command: "moai spec lint SPEC-HARNESS-RETENTION-HARDEN-001"
  pre_edit_result: "✓ No findings — all SPEC documents are valid (exit 0)"
  post_edit_result: "✓ No findings — all SPEC documents are valid (exit 0), measured with the spec.md, progress.md and CHANGELOG.md edits in the working tree, before the commit"
  spec_audit: "moai spec audit (read-only), pre-edit and post-edit, both exit 0 and both summaries identical: Total SPECs 1028, Grandfathered 266, Modern-era clean 759, Drift findings 737; the only finding naming SPEC-HARNESS-RETENTION-HARDEN-001 is [INFO] EraAutoDetected (V3R6); no SyncStatusDrift for it"
  judging_build: "installed moai v3.2.0-rc.26 (45600e4ee), a strict ancestor of the tree HEAD; internal/spec unchanged between 45600e4ee and HEAD"
open_gaps:
  - Windows runtime not observed; Linux unobserved
  - the installed moai lags the tree HEAD; the post-edit lint and audit outputs are in the sync report
  - the awk-based B12 counter was not run (the task brief names awk as refused by the worktree guard; not attempted); a grep over the AC-HRH identifiers was substituted, so the reserved-token handling of the counter was not exercised (no [RETIRED] or [REF] token is present in acceptance.md, grep -c printed 0)
  - the decision record D-1 lives on the card evidence path, not on a home-surface decision board (no CLI verb found; decision-records.md)
```

## §F Phase 4 Mode Selection

Decision: serial

Justification: a single implementation agent per milestone (M0 to M6), because the work is coding-heavy, concentrated in one production file, and the milestones depend on each other (the tail-carry and state-path semantics come first, the stamp-write and disclosure steps last). Logged at sync, after the run, from the leader's brief; the run-phase record is `.moai/reports/t1432/verdict.md`.

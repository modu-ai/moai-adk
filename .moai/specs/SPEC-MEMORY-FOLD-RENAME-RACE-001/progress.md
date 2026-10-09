# progress.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

card: t1568 · phase: plan · tier: M · baseline tree: `2aab5f797` (base absorbed from `81786284e`, lane decision D-1)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
- plan_complete_at: 2026-10-09
- revision: 0.1.3 (audit iter-2 fix round — D6 Windows judge corrected to the release multi-OS gate's Windows leg + DoD timing recorded, D7 AC-MRR-009 extended to the 4-tuple with the between-writes seam, D8 P1 premise reworded, A4 selector 28→29 with re-measured baseline; HISTORY v0.1.3; the v0.1.2 iter-1 fix round and v0.1.1 base-absorb re-pin records are retained below)
- tier: M
- artifacts: spec.md + plan.md + acceptance.md (Tier M set) + design.md (lane-directed mechanism-rationale record) + decision-index.md (conditional — `interview.decision_gate: on`) + progress.md
- req_count: 7 (REQ-MRR-001…007, gap-free, GEARS notation)
- ac_count: 9 (AC-MRR-001…009, gap-free; traceability table at acceptance.md §2 — every REQ covered, no orphan criterion)
- baseline_tree: 2aab5f797
- certain_file_count: 6 (Tier M band 5-15; the added file is `fold_store_lock_windows_test.go`, AC-MRR-008's Windows judge; two named contingencies in plan.md §A.2)
- development_mode: tdd (RED-first is the card's mandate and plan.md M1)
- card_id: t1568
- decision_index: 5 rows, all `FOUNDER`/implementation-level, all `Default:`-carrying, all stamped `DEFAULT-APPLIED 2026-10-09T02:43:00Z glm via manager-spec` — no operator-blocking row
- re-pin record: +2 below the import block, +3 after the `foldClosedCardMemory` hunk (`os.Rename` `:647`→`:649`, `applyFold` `:412`→`:414`, `foldOnDoneStep` caller `:1172`→`:1175`); `memory_fold_test.go` and `internal/sessionmsg/lock_*.go` verified unchanged (`git diff 81786284e 2aab5f797 --stat` empty)

### Plan-phase lint (tool provenance: judging build vs measured tree)

- judging build (current): `/tmp/moai-t1568-lint2`, built from THIS tree at HEAD `2aab5f797` (`go build -o /tmp/moai-t1568-lint2 ./cmd/moai`), invoked by path. Re-run after the revision 0.1.3 fix-round edits: `spec lint SPEC-MEMORY-FOLD-RENAME-RACE-001` → `✓ No findings — all SPEC documents are valid`, exit 0 — the run of record for 0.1.3.
- Historical (old base `81786284e`, binary built from that tree): run 1 `0 error(s), 11 warning(s)` (unanchored `-run` patterns), run 2 after anchoring `✓ No findings` — both superseded by the run above; kept for the record only.
- Anchoring note: the family selector is `-run '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'` — end-anchored via `.*$`, selecting exactly the fold regression family: 27 `TestMemoryFold*` tests (including the 13 `TestMemoryFoldOnDone_*` wiring tests and the write-ordering regression `TestMemoryFold_ArchiveRecheckedBeforeMemoryRename`), `TestReviewArchiveUpdateDuringEffectiveScan`, and `TestReviewSequentialAbandonedTempOwnership` (added at iter-2 A4 — the codex-review round-2 per-worker temp-ownership regression the 28-test form missed). Selector history: iter-1 (D1) established the first selector `^Test(Fold|Review).*$` swept 53 Review-prefix tests and ZERO fold tests; the iter-2 28-test form superseded it; the final form's set is verified by the plan §C.1 `-list` gate before every use: `go test -list '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$' ./internal/cli` → 29 tests, exit 0 (verbatim list persisted at `.moai/reports/t1568/t1568-family-list3.txt`).

### Pre-fix baseline (green-before for AC-MRR-007)

- Attempt 1 (base `81786284e`, `-timeout 240s`): **timeout artifact, not a red** — verbatim: `panic: test timed out after 4m0s` (log line 15), `FAIL github.com/modu-ai/moai-adk/internal/cli 244.760s`, and ZERO `--- FAIL` test-level lines in the full log (grep-verified). The goroutine dump's parked test (`TestReviewUnrecordedPickedHubWait` in the `homestate` busy-retry path) is the dump's illustration, not a failed assertion. Claim policy: this run proves nothing about the family's health in either direction.
- Attempt 2 (base `2aab5f797`, wrong selector `^Test(Fold|Review).*$`): **mis-measurement, kept on record as such** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	414.385s`, exit 0, but the swept set was 53 Review-prefix tests and ZERO of the 27 `TestMemoryFold*` fold tests (audit iter-1 D1). Never cite it as the green-before.
- Attempt 3 (base `2aab5f797`, 28-test selector `^(TestMemoryFold|TestReviewArchiveUpdate).*$`, `-timeout 900s`, env-scrubbed compound): **GREEN, then superseded as the record** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	38.008s`, exit 0; log persisted at `.moai/reports/t1568/t1568-baseline3.txt`. Kept honest: iter-2 A4 established this selector missed `TestReviewSequentialAbandonedTempOwnership`, so its command is no longer AC-MRR-007's exact command.
- Attempt 4 (base `2aab5f797`, FINAL 29-test selector `^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$`, `-timeout 900s`, env-scrubbed compound): **GREEN — the green-before of record** — verbatim `ok  	github.com/modu-ai/moai-adk/internal/cli	42.974s`, exit 0 (own task output `EXIT=0`), this run / this tree. Logs persisted in-tree: `.moai/reports/t1568/t1568-baseline4.txt` (plus attempts 1-3 and all three `-list` verifications). AC-MRR-007's post-fix green must equal this run.

### RED-now discipline status

- AC-MRR-001's RED cell is PENDING EXECUTION **by choice, not by impossibility** (audit iter-1 D4 corrected the premise): the existing `orderProbe("bytes-done")` seam (`memory_fold.go:643-645`) fires inside the window, so a plan-phase RED is executable today by setting it per plan.md B-3's set/restore discipline. The deferral to run-phase M1 is the clean-dedicated-seam choice: `orderProbe`'s contract is stage recording (owned by the ordering regression), not mutation; the dedicated `mutateBeforeRename` seam keeps the mutation contract separate and fires after the abandonment check — the true last observable moment. The observation itself is M1's first deliverable and its verbatim output + exit code + tree SHA land in progress.md §E.2; M2 may not start before it is on record. Every executable plan-phase check (ID regex PASS, uniqueness, chokepoint grep, lint, corrected-selector baseline) is on record in `.moai/reports/t1568/plan-evidence.md`.

## §F Phase 4 Mode Selection

- Inputs: tier M · scope 6 files · domains 1 (Go, internal/cli) · language mix Go-only · concurrency benefit LOW (coding-heavy, strict RED→GREEN ordering)
- Evaluation: direct — not selected (multi-file, test-first loop with milestones) · fanout — not selected (coding-heavy per the Anthropic coding-task parallelism caveat) · sweep — not selected (not mechanical-uniform) · agent-team — not requested
- Decision: **serial**
- Justification: single-package defect fix whose milestones are strictly ordered (M1 RED must be observed before M2 GREEN; plan.md §F). One writer per tree; sequential per-milestone delegation is the safe default for coding work.
- Baseline-attribution: decided by lane-22 at run entry, plan-audit iter-3 PASS (0.95, rcpt-d7cf366e7f8d3eb44edd6dae) at HEAD 2f436db17.

## §E.2 Run-phase Evidence

### M1 — RED: the loss reproduced (AC-MRR-001's RED cell, observed before any fix code)

- **Observed at**: 2026-10-09, on the M1 work tree = HEAD `2f436db17` + the M1 test-only seam fields and test only (production `atomicWriteFoldFile`/`applyFold` checks unchanged; the seam call sites are additive). This exact tree is snapshotted by the M1 commit `84bdf072c`.
- **Command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'`
- **Exit code**: 1
- **Verbatim** (every iteration red — 10/10 top-level FAIL, 20/20 subtests FAIL; full raw log `.moai/reports/t1568/t1568-red-m1.txt`):

```
--- FAIL: TestFoldRenameWindowConcurrentWriter (0.09s)
    --- FAIL: TestFoldRenameWindowConcurrentWriter/archive-append (0.01s)
        memory_fold_test.go:1466: REQ-MRR-004 violated: the fold returned success while the writer's line published inside the archive-append rename window was destroyed from project_card_archive_2026_10.md
    --- FAIL: TestFoldRenameWindowConcurrentWriter/memory-rewrite (0.00s)
        memory_fold_test.go:1466: REQ-MRR-004 violated: the fold returned success while the writer's line published inside the memory-rewrite rename window was destroyed from MEMORY.md
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.220s
```

  (the `--- FAIL`/`REQ-MRR-004 violated` pair repeats identically for all 10 iterations; the block above is iteration 1 verbatim)
- **Why red for the right reason**: the fold returned success (`runFoldOK` passed) while the writer's line published at the `mutateBeforeRename` point — after every pre-rename check — was destroyed by `os.Rename`, on BOTH write surfaces (archive append; MEMORY.md rewrite). The first attempt's red (`.moai/reports/t1568/t1568-red-m1-first-fixture.txt`) was a fixture defect — `no archive index to fold into` (the store lacked the archive file) — corrected by seeding `minimalArchive()`; it is kept on record as a wrong-reason red, never cited as the AC-MRR-001 RED.

### M2 — GREEN: the store lock (AC-MRR-002 / AC-MRR-004 / AC-MRR-009)

- **Implementation**: `foldStoreLock` in `fold_store_lock_unix.go` (unix.Flock LOCK_EX|LOCK_NB, `unacquiredFD = -1` sentinel, bounded retry `foldStoreLockRetryDeadline = 2s` + `foldStoreLockRetryInterval = 10ms`, lock file `<store-dir>/.moai-fold.lock` mode 0644 `O_CREAT|O_RDWR|O_CLOEXEC`, never removed) and its Windows parity `fold_store_lock_windows.go` (LockFileEx LOCKFILE_EXCLUSIVE_LOCK|LOCKFILE_FAIL_IMMEDIATELY); `applyFold` acquires at entry and `defer`s the release — nothing between the existing checks moved (D-1). The window test's writer is now the cooperating form (plan §A.3); one test-helper narrowing rode along: `requireNoTempFiles`'s leftover prefix tightened from `.moai-fold` to `.moai-fold-` because the lock file is an INTENDED permanent resident (D-3) the old prefix falsely counted as a leftover.
- **AC-MRR-002/004 command**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -race -count=10 ./internal/cli -run '^TestFoldRenameWindowConcurrentWriter$'` → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli  2.281s` — 10/10 iterations, both surface subtests (archive-append; memory-rewrite) green under the race detector (this run, this tree).
- **AC-MRR-009 command**: same compound with `-run '^TestFoldStoreLockSpanHeldThroughApply$'` → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli  2.304s` — 10/10 iterations; the sampled tuple is exactly (mutateDisk=acquired, mutateDuringWrite=refused, mutateBetweenWrites=refused, mutateBeforeRename=refused).
- **Lock unit cells**: `go test -count=1 ./internal/cli -run '^TestFoldStoreLock'` → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli  2.868s` — acquire→contending-try refused→bounded refusal within deadline naming the store→release→re-acquire succeeds→release idempotent→lock file never removed (D-3); stale unlocked file acquires (§3 edge 1); two stores hold simultaneously (§3 edge 3).
- **Windows compile gate (AC-MRR-008 local leg)**: `GOOS=windows go build ./internal/cli/...` → exit 0; `GOOS=windows go vet ./internal/cli/` → exit 0 (this run, this tree).

### M3 — Surface cells (AC-MRR-003 / AC-MRR-005 / AC-MRR-006)

- **AC-MRR-003 command**: `go test -count=1 ./internal/cli -run '^TestFoldStoreLockContentionRefusesCleanly$'` (env-scrubbed compound) → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli  2.864s` — the contended fold refuses via the bounded-wait path (2s), its error names the contended store, and both store files are byte-identical to the pre-invocation state (asserted inside the test by `storeHashes` + `requireSameStore`).
- **AC-MRR-005 command**: same compound with `-run '^TestFoldLockFileInvisibleToTooling$'` → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli  0.819s` — `memory doctor --dir <store> --json` and the fold preview are byte-identical with and without `.moai-fold.lock` present (asserted by in-test comparison); the unlinked-archive listing carries no lock-file entry (implied by the byte identity).
- **AC-MRR-006 command**: `go test -race -count=1 ./internal/cli -run '^TestFoldOnDoneContentionAbandonsWithoutWrite$'` (env-scrubbed compound) → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli  5.325s` — both cells green: (1) waiting under contention reports exactly one abandonment stderr line and begins no write; (2) the step HOLDING the lock with its apply in flight reports abandonment, the worker exits (observed via the wiring's `workerExit` synchronization), and a fresh non-blocking acquire on the same store succeeds — the worker's deferred release ran (§3 edge 4).
- **Refused-tool note (§3.1)**: the first attempt to append the AC-MRR-006 test body via a compound shell heredoc was refused by the worktree-isolation guard (command-complexity refusal); the body was re-authored through the file-edit tool unchanged and the refusal carried no content loss.

### M4 — Close-out (AC-MRR-007 / AC-MRR-008 local leg)

- **Selector swept-set gate (plan §C.1)**: `go test -list '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$' ./internal/cli` → **exit 0, 29 tests**, including `TestMemoryFold_ArchiveRecheckedBeforeMemoryRename`, `TestReviewArchiveUpdateDuringEffectiveScan`, `TestReviewSequentialAbandonedTempOwnership` (non-empty, family-complete — the empty-sweep guard).
- **Family regression (AC-MRR-007)**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -race -count=1 -timeout 900s ./internal/cli -run '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'` → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli  26.263s`, 0 `--- FAIL` lines — equal in HEALTH to the recorded green-before (attempt 4, `ok … 42.974s`, exit 0): same selector, same race detector, zero reds. Wall-time is not the comparison axis (a loaded machine measures the machine).
- **First M4 family run was RED, for a now-repaired reason (kept honest)**: attempt 1 → **exit 1**, verbatim `memory_fold_test.go:1069: store file list changed: 14 files before, 15 after` (log `.moai/reports/t1568/t1568-family-m4-first-red.txt`) — `requireSameStore` counted the newly-created `.moai-fold.lock` as a store-content change. Repair: `requireSameStore` (memory_fold_test.go, an in-scope §A.2 file) now normalizes `foldLockFileName` out of both snapshots — the lock file is the lock mechanism's own resident, not store content (D-3), the same narrowing `requireNoTempFiles` applies. `storeHashes` itself was NOT touched (it lives in `memory_budget_test.go`, outside §A.2 — D-5). Re-run → green above.
- **Build**: `go build ./...` → **exit 0** (this run, this tree).
- **Coverage (E3)**: `go test -race -count=1 -timeout 900s -cover ./internal/cli -run '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'` → exit 0, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli  29.038s  coverage: 7.2% of statements` — package-wide statements against the family selector's swept subset; no threshold claim beyond that number.
- **Lint (E5)**: `golangci-lint run --timeout=2m internal/cli/...` → **exit 0, `0 issues.`** (no NEW issues — the baseline is clean and stays clean); `gofmt -l internal/cli/` → empty (exit 0); `go vet ./internal/cli/` → exit 0.
- **Scope placement note (the 7th file)**: the AC-MRR-006 cell lives in `memory_fold_wiring_test.go` — a file plan §A.2's 6-file table did not enumerate. The placement follows acceptance.md AC-MRR-006's own direction ("the wiring test's existing recorder/seam conventions": `wireFixture`, `runWireClose`, `waitWorkerExit`, `wireErrLines` all live there). This is an implementation-placement detail, not a scope-doc change; every other §A.2 file is exactly as planned (memory_fold.go, memory_fold_test.go, fold_store_lock_unix.go, fold_store_lock_windows.go, fold_store_lock_test.go, fold_store_lock_windows_test.go). Recorded here for the auditors; no blocker raised (no SPEC body text needed changing).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-10-09
run_commit_sha: fc4eedd33   # the M4 close-out commit; backfilled per the schema D3 exemption (a commit cannot cite its own SHA)
run_status: complete
ac_pass_count: 8   # AC-MRR-001..007 + AC-MRR-009 (each with its command + observed output in §E.2)
ac_fail_count: 0
ac_pass_with_debt: [AC-MRR-008]   # timing exception, recorded where the chain is read: its judge is release-pr-multi-os.yml's Windows leg, observed at the release window after card close (acceptance.md §4); local GOOS=windows build+vet exit 0 on record
preserve_list_post_run_count: 0   # no preserve-list item outstanding: the 29-test family is green, the orderProbe sequence (effective-start → bytes-done) observed unchanged by TestReviewArchiveUpdateDuringEffectiveScan
l44_pre_commit_fetch: n/a (run-phase agent commits locally on the card worktree branch; no push — the lane owns the landing and its fetch/push evidence)
l44_post_push_fetch: n/a (no push performed by this agent)
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues; go vet clean; gofmt clean — measured in this run against a clean baseline
cross_platform_build:
  darwin_amd64: exit 0   # go build ./...
  windows: exit 0        # GOOS=windows go build ./internal/cli/... (+ GOOS=windows go vet ./internal/cli/ exit 0)
  linux: judged by CI    # not run locally; origin/develop CI is the full-suite and per-OS judge
total_run_phase_files: 7   # the 6 files of plan §A.2 + memory_fold_wiring_test.go (the AC-MRR-006 cell; see the M4 scope-placement note)
m1_to_mN_commit_strategy: one conventional commit per milestone (M1 RED / M2 GREEN / M3 cells / M4 close-out), each carrying card id t1568 and Authored-By-Agent: manager-develop
```

### Run-phase evidence index

| Evidence | Where |
|---|---|
| RED verbatim (10/10, exit 1) | §E.2 M1 block above; raw log `.moai/reports/t1568/t1568-red-m1.txt` (gitignored machine-local) |
| Wrong-reason first red (fixture) | `.moai/reports/t1568/t1568-red-m1-first-fixture.txt` — never cited as the AC-MRR-001 RED |
| GREEN window/span/unit verbatim | §E.2 M2 block above |
| Surface cells verbatim | §E.2 M3 block above |
| Family green-before (baseline of record) | §E.1 attempt 4 (`ok … 42.974s`, exit 0, base `2aab5f797`) |
| Family post-fix green | §E.2 M4 block above (`ok … 26.263s`, exit 0, HEAD per the M4 commit) |
| First M4 family red + repair | §E.2 M4 block above; raw log `.moai/reports/t1568/t1568-family-m4-first-red.txt` |

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_status: audit-ready
sync_complete_at: 2026-10-09
sync_commit_sha: 496eab330   # the sync commit above (backfilled from the pending-backfill-sync placeholder per the D3 backfill window — a commit cannot cite its own SHA)
files_changed:
  - CHANGELOG.md   # Unreleased § Fixed — sync-phase close entry prepended at the top of the first Fixed list (newest-first, t1566 precedent); B12 pre-checks: `grep -c SPEC-MEMORY-FOLD-RENAME-RACE-001 CHANGELOG.md` = 0 pre-emission, claimed paths verified by `ls internal/cli/`
  - .moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/spec.md   # frontmatter status: in-progress → implemented → completed, merged into this single sync commit (3-phase close — completed rides the sync commit); updated: 2026-10-09 (already the current date — no byte change)
  - .moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/progress.md   # this §E.4 block
b12_self_test_a: pass   # duplicate-entry guard: grep -c 'SPEC-MEMORY-FOLD-RENAME-RACE-001' CHANGELOG.md → 0 pre-emission
b12_self_test_b: pass   # AC count match: live-identifier counter on acceptance.md (tier M → acceptance.md is the AC source) → live=9 excluded=0 ambiguous=0, exit 0; the CHANGELOG entry cites 9 (AC-MRR-001..009 = §E.3's 8 pass + 1 pass-with-debt)
b12_self_test_c: pass   # file-path verification: all 7 claimed paths (memory_fold.go, fold_store_lock_unix.go, fold_store_lock_windows.go, + 3 test files, memory_fold_wiring_test.go) confirmed by `ls internal/cli/`
changelog_entry_position: Unreleased > Fixed > first item (newest-first)
frontmatter_status_transitions:
  - artifact: spec.md
    transition: in-progress → implemented → completed
    carrier: the single sync commit (3-phase close; no separate Mx chore commit)
mx_tags:
  added: 0
  removed: 0
  rationale: >-
    sync sub-step scan of the new/edited files (memory_fold.go diff,
    fold_store_lock_unix.go, fold_store_lock_windows.go + the three test files)
    against mx.yaml thresholds: no ANCHOR (foldStoreLock acquire/release has 1 call
    site; applyFold itself fan_in = 2 < fan_in_anchor 3), no WARN (no goroutine,
    cyclomatic complexity < 15, branch depth < 8), no NOTE/TODO trigger (all new
    symbols are unexported with godoc; every new function is tested). No existing
    tag touched. code_comments: en respected.
notes: >-
  plan.md and acceptance.md carry NO frontmatter block (headings-only, permitted
  by Artifact Statelessness) — no `updated:` refresh surface exists in either;
  bodies untouched (frontmatter status axis only, per the Status Transition
  Ownership Matrix). spec.md body content untouched. No README/docs-site change:
  the fold verb gains concurrency safety only — no flag, output, or CLI surface
  change. No push — the lane owns landing.
```

# progress.md — SPEC-MEMORY-FOLD-RENAME-RACE-001

card: t1568 · phase: plan · tier: M · baseline tree: `2aab5f797` (base absorbed from `81786284e`, lane decision D-1)

## §E.1 Plan-phase Audit-Ready Signal

- plan_status: audit-ready
audit_ready: true
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

## §J Lane landing record — run tmnboq, lane-1 (generation 2)

- **배차 수락**: dispatch_notice `baf34d41c8cb3f32c47b11787086697c` (from `069be28e-d334-404a-9ab2-4634e2b57c1d`) — `factory_msg_receipt` `accepted`, acknowledged.
- **임대**: `factory_next --card t1568` → `t1568 … picked`, worktree `t1568` (the lease binds `.moai/worktrees/t1568`). Entered by `EnterWorktree` on branch `WT-fold-rename-race`.
- **트리 실측 (디스크)**: HEAD `b14e4c5cf`; `git status --short` empty; `git rev-list --count --left-right main...HEAD` = `0 10`; `main..HEAD` lists the ten commits e54e9711f → b14e4c5cf.
- **판독한 증거**: `sync-audit.md` verdict PASS (harmonic mean 0.887, threshold 0.85, `audited_sha 496eab330`, judged at HEAD b14e4c5cf). `card-review.md` (previous generation, `codex_review scope=card` on `933e6d06e`): verdict fail (advisory), one P1 at `internal/homestate/card_transition.go:714`, outside the card diff (intersection 0), card diff itself zero findings. Since `933e6d06e` the diff is docs-only (`git diff --stat 933e6d06e HEAD`: `CHANGELOG.md`, `spec.md`, `progress.md`; 37 insertions, 2 deletions).
- **단계 재개**: the run row's stage is `""`, so the resume edges T4a–T4f were refused (`leased card stage "" does not resume into sync-audit`). The walk therefore starts at `plan`: `plan` (v4, T4a), then `plan-audit` (v5, T5) with evidence `2f436db17:.moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/progress.md`.
- **보류 (keep-set, operator-held)**: `kickoff` (T7) refused with the verbatim text `card evidence refused: verdict file …/plan-audit-iter3.md: plan-audit ceiling refusal (hold): plan-audit ceiling reached (round count 3 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G`. The score (iter-3 PASS 0.95, §F) does not lift the ceiling. The lane does not record this decision itself. Options for the operator: (a) record an operator acceptance of the round-3 verdict in progress.md §G (Override and Refusal Record), then retry `kickoff`; (b) take the REQ-ACE-005 split or new-SPEC route; (c) stop the card.
- **Lane-local verification (this run, HEAD b14e4c5cf)**: `go test -race -count=1 -timeout 900s ./internal/cli -run '^(TestFold|TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'` → `ok  	github.com/modu-ai/moai-adk/internal/cli	60.704s`. Swept set (`go test -list` with the same selector): 37 tests, not empty. Caveat: the `MOAI_KANBAN_*` env scrub wrapper was refused by the worktree guard, so this run had no env scrub; all 37 passed anyway.
- **Fresh card review (this run)**: `codex_review scope=card`, tree `b14e4c5cf`, base `db0c514d3` (the scope resolver's merge base includes the absorbed history), backend codex, task `kbz6p95qq` → verdict pass, advisory, findings `[]`. The earlier P1 was not re-raised in this review. Both reviews are advisory and file no receipt.
- **Binary lag**: the MCP server runs build `db0c514d3`; this tree's HEAD is `b14e4c5cf`. The stage refusals above are the server's output from that build. Any moai CLI result cited here carries that lag.
- **Verdict-file tracking**: `.gitignore:235` (`.moai/reports/*`) excludes the report directory, so `sync-audit.md`, `card-review.md`, and `plan-audit-iter{1,2,3}.md` exist on disk only and are not in git history. The factory guards read them from disk; a reader on another checkout cannot see them.
- **Explicit wait**: recheck point is §G (Override and Refusal Record), this §J, and cron `4706ef9c` (`7,27,47 * * * *`). On wake: read §G first; resume at `kickoff` only after a recorded operator decision in §G.
- **Blocker report**: sent to the leader via `factory_msg_send` (kind `blocker`), with the same evidence references.

wait record: id=w-t1568-20261009T161307Z waiting_on=operator reason=plan-audit-ceiling-hold-round3-over-tierM-ceiling2-REQ-ACE-006-kickoff-T7-keepset-operator-held recheck=progress.md-section-G-Override-and-Refusal-Record-on-each-standing-cron-fire-4706ef9c

inbox disposition: msg=8c04d1074db32c4d9941e87b561217a2 kind=dispatch_notice from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=deferred reason=relayed-operator-ceiling-exception-not-on-disk(§G-has-only-refusal-record;board=absent) verified=plan-audit-iter3.md-verdict-PASS-audited_sha-2f436db179dcbf7fba337523454171299861600c action=not-transcribed-as-operator-decision;kickoff-not-retried(would-append-duplicate-refusal-record) recheck=cron-4706ef9c-next-fire

inbox disposition: msg=872747d5743a3e8faba86177ac96a19c kind=status_report from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=accepted reason=board-record-d-20261009T163446Z-aa63-read-on-disk-resolves-w-t1568-20261009T161307Z verified=moai-decision-read-board=ok-records=1 action=wait-ended-by-board-ruling;resume-recovery-path

inbox disposition: msg=06a62dbf8f034190aca6c7742cf846a4 kind=status_report from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=accepted reason=lease-expiry-path-code-verified(card_transition.go:309-316 transitionTx expired-lease return; 412-429 applyLeaseExpiry keeps stage-worktree-evidence and bumps version) verified=read-only-code-read action=expiry-probe-then-factory-next(primary-root)-then-resume

decision record: decided_by=claude+lane-1-watchdog evidence_refs=board:d-20261009T163446Z-aa63;.moai/reports/t1568/plan-audit-iter3.md(verdict=PASS,audited_sha=2f436db179dcbf7fba337523454171299861600c) ladder_path=② decision-board ruling (keep-set relay; wait w-t1568-20261009T161307Z resolved)

recovery record: at=2026-10-09T16:53:25Z probe=refused-ErrLeaseExpired(lane-1 expired 2026-10-09T16:25:01.801021Z; row returned to assigned) factory_next(primary-root)=t1568-stage=plan-audit-worktree=t1568 stage-resume=plan-audit-v8-lease-renewed kickoff=refused-ceiling(card-evidence-refused:plan-audit-iter3.md;release-path-§G-operator-decision)

code-check: read-only at HEAD a7b89e294; ceiling-exception-kind-readers=0 (internal/decision/board.go:34 defines KindCeilingExcpt; no non-test reader); EvaluatePlanAuditCeiling-reads-board=no (internal/runtime/audit_ceiling.go:1298-1365); debt-proceed-only-for-admitted-PASS-WITH-DEBT (audit_ceiling.go:1346-1348)

unverified-gap: the admitBase clause that refuses plan-audit-iter3.md is not isolated; visible fields satisfy it (verdict PASS, overall_score 0.95, must_pass_failed 0, blocking_count 0, convergence_overall pass, required_backend codex pass, plan_artifact_hash present); candidates left: duplicate decision keys, hashOK plumbing, latest-verdict selection

wait record: id=w-t1568-20261009T165325Z waiting_on=operator reason=kickoff-refused-after-board-ruling-ceiling-engine-does-not-read-board-disposition-must-land-in-engine-record-path-supersedes=w-t1568-20261009T161307Z-ended-by-board-ruling-d-20261009T163446Z-aa63 recheck=progress.md-section-G-plus-audit-ceiling-state-on-each-cron-fire-4706ef9c

inbox disposition: msg=7171a08bf57b544c297d7ebe01e4e873 kind=status_report from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=accepted reason=leader-kickoff-hold-reply-superseded-by-cause-report-93ea9fb7 action=wait-superseded-by-verdict-format-defect-path

inbox disposition: msg=519aaa4c642820d3a146d070344b9f17 kind=status_report from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=accepted reason=standing-rulings-recorded;rule1-expiry-path-applied;rule3-landing-hold-applied(no complete/integration-merge/push/PR) action=no-landing-actions-this-run

inbox disposition: msg=93ea9fb7c5a1081502758d52e2927043 kind=status_report from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=accepted reason=cause-verified-on-disk: plan-audit-iter3.md prose lines 20-22 collide with machine keys verdict/overall_score/plan_artifact_hash; verdict.go:88-152 folds prose onto decision keys; verdict.go:281-284 refuses duplicated decision keys; compute_hash=3946a2abbe2c9bef027ae3dcc7cd3db4ed141d8f64a63239888ab7004b72b0aa differs from iter3 aa8c0d57 action=iter3-not-hand-edited;fresh-audit-iter4-at-HEAD

inbox disposition: msg=8baff4311288d2415d3367b5b434055e kind=status_report from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=accepted reason=verdict-files-stay-uncommitted(.gitignore:235 .moai/reports/*; M4 correction of the M2 git-add-f advice) action=no-commit-of-verdict-files

decision record: decided_by=claude+lane-1-watchdog evidence_refs=.moai/reports/t1568/plan-audit-iter3.md(lines 3-8 machine block; lines 20-22 prose duplicates);internal/auditverdict/verdict.go:281-284;internal/runtime/audit_ceiling.go:1332-1333;compute_hash=3946a2abbe2c9bef027ae3dcc7cd3db4ed141d8f64a63239888ab7004b72b0aa ladder_path=③ audit cross (fresh plan-audit-iter4 at HEAD a7b89e29494a83d86112c30cb6be77d3b6f31bd3; the ceiling clears only by an admission-clean higher round)

wait record: id=w-t1568-20261009T171107Z waiting_on=delegate-plan-auditor reason=iter4-audit-at-HEAD-a7b89e294-spawned;operator-ceiling-path-superseded supersedes=w-t1568-20261009T165325Z recheck=.moai/reports/t1568/plan-audit-iter4.md-on-next-cron-fire-4706ef9c

inbox disposition: msg=a6c33a28d5f27fad2a39c9aefbd60ea1 kind=status_report from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=accepted reason=leader-disabled-turn-end-codex-review-gate-in-this-worktree(verified-on-disk: workflow.yaml review_gate.enabled true→false, uncommitted) action=workflow.yaml-never-staged-or-committed;reverted-before-landing;card-review-codex_review-scope-card-still-required

decision record: decided_by=claude+lane-1-watchdog evidence_refs=.moai/reports/t1568/plan-audit-iter4.md(verdict=FAIL,audited_sha=a7b89e29494a83d86112c30cb6be77d3b6f31bd3,overall_score=0.75,must_pass_failed=2,blocking_count=7,plan_artifact_hash=3946a2abbe2c9bef027ae3dcc7cd3db4ed141d8f64a63239888ab7004b72b0aa,required_backend=codex-fail,convergence_overall=fail) ladder_path=③ audit cross negative → fail-closed (no kickoff; escalated to leader; no further self-initiated audit rounds)

wait record: id=w-t1568-20261009T173448Z waiting_on=operator reason=plan-audit-iter4-FAIL-fail-closed-operator-bulk-approval-excludes-FAIL-cards-split-or-hold-or-plan-revision-decision-owed supersedes=w-t1568-20261009T171107Z-delegate-deliverable-consumed recheck=next-cron-fire-4706ef9c-read-§J-and-decision-board-for-operator-ruling

inbox disposition: msg=f1463f6ba88caed0f6090b3399ebb13b kind=status_report from=069be28e-d334-404a-9ab2-4634e2b57c1d disposition=accepted reason=operator-decision-d-20261009T180009Z-7c86-verified-on-board(wait-resolution resolves w-t1568-20261009T173448Z; operator answer 지적 수리 후 재감사) action=repair-D1-D7-by-manager-spec(plan artifacts only);lease-recovery-then-T6;commit-repair-then-T5-then-audit-iter5;landing-hold-kept

decision record: decided_by=claude+lane-1-watchdog evidence_refs=board:d-20261009T180009Z-7c86(decided_by=operator via leader; resolves w-t1568-20261009T173448Z);.moai/reports/t1568/plan-audit-iter4.md(verdict=FAIL,audited_sha=a7b89e29494a83d86112c30cb6be77d3b6f31bd3,blocking_count=7) ladder_path=② decision-board wait-resolution (operator answer followed; wait w-t1568-20261009T173448Z ended by the board record)

decision record: decided_by=claude+lane-1-watchdog evidence_refs=board:d-20261009T180009Z-7c86(operator: 지적 수리 후 재감사);acceptance.md(AC-MRR-008 reclassified; AC-MRR-010 added; §4 post-close exception);plan.md(M5 conditional on operator scope decision) ladder_path=⑤ self-judgment: the repaired artifacts stay uncommitted until item 4 wording (acceptance.md:139, presupposes option (b)) is made conditional; the fifth audit proceeds on the commit per the operator answer; M5 option (a) or (b) stays with the operator and does not block the fifth audit

wait record: id=w-t1568-20261009T182819Z waiting_on=delegate-manager-spec reason=acceptance-md-L139-conditional-wording-fix-before-commit recheck=next-cron-fire-4706ef9c

wait record: id=w-t1568-20261009T183041Z waiting_on=delegate-manager-spec reason=line-139-fix-on-disk-at-18:30:00Z-delegate-still-running-commit-waits-for-its-completion-one-writer-rule recheck=delegate-completion-notice-or-next-cron-fire-4706ef9c supersedes=w-t1568-20261009T182819Z-reason-fixed

decision record: decided_by=claude+lane-1-watchdog evidence_refs=board:d-20261009T180009Z-7c86(operator: 지적 수리 후 재감사);commit=f30bb9088cde7b12612993026b5e8d2f1244e4f0(spec.md plan.md acceptance.md only; progress.md excluded);compute_hash=ae78cfa7e3772eb55df9c56f942478a1ce538fe16731eef787f41a70c85a9a0b;T5-evidence=f30bb9088cde7b12612993026b5e8d2f1244e4f0:.moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/plan.md(plan-audit v21) ladder_path=⑤ self-judgment: the fifth audit runs on this commit per the operator answer; M5 option (a) or (b) remains with the operator and does not block the fifth audit

wait record: id=w-t1568-20261009T183155Z waiting_on=delegate-plan-auditor-iter5 reason=iter5-audit-at-f30bb9088-spawned supersedes=w-t1568-20261009T183041Z-delegate-manager-spec-done recheck=.moai/reports/t1568/plan-audit-iter5.md-on-next-cron-fire-4706ef9c

decision record: decided_by=claude+lane-1-watchdog evidence_refs=.moai/reports/t1568/plan-audit-iter5.md(verdict=FAIL,audited_sha=f30bb9088cde7b12612993026b5e8d2f1244e4f0,overall_score=0.69,must_pass_failed=1,blocking_count=6,plan_artifact_hash=ae78cfa7e3772eb55df9c56f942478a1ce538fe16731eef787f41a70c85a9a0b,required_backend=codex-fail,receipt=rcpt-6bd0379b547be05225f15779);board:d-20261009T180009Z-7c86(a fifth FAIL returns the decision to the operator) ladder_path=③ audit cross negative → fail-closed; escalated to the operator; no further lane audit round; no kickoff attempted

wait record: id=w-t1568-20261009T185625Z waiting_on=operator reason=iter5-FAIL-fail-closed-decision-returns-to-operator-options-split-hold-another-repair-or-stop-and-M5-a-or-b-owed supersedes=w-t1568-20261009T183155Z-delegate-done recheck=next-cron-fire-4706ef9c-read-§J-and-decision-board

blocker re-send: id=ca4360bfbfc1063074bde7b30f29f9fe kind=blocker to=leader(session=069be28e-d334-404a-9ab2-4634e2b57c1d) sent=2026-10-09T21:53:02Z expires=2026-10-09T22:53:02Z correlation=w-t1568-20261009T185625Z replaces=d76d1d87d8266be55d8db63b53ede8bf(expired 21:49:59Z) result=queued(no routing object, no delivery notice) delivery=unverified; leader endpoint live (pid 37169, kill -0 ok; registry last_heartbeat 18:23:31Z); NextDelivery=pending-until-next-turn; no new wait record, w-t1568-20261009T185625Z stays open

blocker re-send: id=e1ab70c7d0800874a9c414c7c08e1894 kind=blocker to=leader(session=069be28e-d334-404a-9ab2-4634e2b57c1d) sent=2026-10-09T22:50:33Z expires=2026-10-10T00:50:33Z ttl=7200 correlation=w-t1568-20261009T185625Z replaces=ca4360bfbfc1063074bde7b30f29f9fe(expires 22:53:02Z) result=queued(no routing object, no delivery notice) delivery=unverified; NextDelivery=pending-until-next-turn; leader registry last_heartbeat 18:23:31Z (22:50Z); no new wait record, w-t1568-20261009T185625Z stays open

operator brief: .moai/reports/t1568/operator-brief.md written 2026-10-10T00:12Z by lane-1 (input for the leader's t1568 decision, not a decision); its path was sent to the leader as factory message f843f8974be36359819e0b9afa8a6adc (queued; delivery unverified); operator record d-20261010T000832Z-efde verified on the board (decided_by=operator, scope=standing), but its one-line summary does not carry the rule text, which the brief marks as relayed; no new wait record, w-t1568-20261009T185625Z stays open until the leader records the t1568 decision

## §G Override and Refusal Record

- 2026-10-09T16:10:05Z SPEC-MEMORY-FOLD-RENAME-RACE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter3.md
- 2026-10-09T16:51:31Z SPEC-MEMORY-FOLD-RENAME-RACE-001 ceiling-refusal outcome=hold reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2); the verdict matches no admitting arm and holds, entry blocked (REQ-ACE-006) — release path: the split/new-SPEC route of REQ-ACE-005 or an operator decision recorded in progress.md §G" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter3.md

- 2026-10-10T06:49:16Z SPEC-MEMORY-FOLD-RENAME-RACE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter6.md
- 2026-10-10T06:56:45Z SPEC-MEMORY-FOLD-RENAME-RACE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter6.md
- 2026-10-10T06:57:53Z SPEC-MEMORY-FOLD-RENAME-RACE-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 6 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter3.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter4.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter5.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1568/.moai/reports/t1568/plan-audit-iter6.md
## §K Operator decision applied to the plan artifacts — d-20261010T042423Z-722b

Written by manager-spec (leaf worker spawned by the factory leader session 359ce07d), 2026-10-10. Plan artifacts only; no file outside this SPEC directory was edited, and no implementation file changed.

- **Decision**: board record `d-20261010T042423Z-722b` (decided_by=operator, first-hand answer; scope `card:t1568`; kind ruling), selected option "D2-B + M5-(b)". Read on the board for this revision with `moai decision read --scope card:t1568`; the record's body matches the delegation. Input: `.moai/reports/t1568/operator-brief.md`. Verdict addressed: `.moai/reports/t1568/plan-audit-iter5.md` (FAIL 0.69, 6 blocking, audited_sha `f30bb9088cde7b12612993026b5e8d2f1244e4f0`).
- **What the decision fixes**: D2-B — a symbolic link at the lock path is out of scope, with the threat model and the residual risk in spec.md. M5-(b) — M5 is not run; the SPEC stays `completed`; no amendment procedure; REQ-MRR-001's cross-process clause and the card-close entry point are unverified residual risk; AC-MRR-010 stays unadopted.
- **spec.md** (v0.1.4 → v0.1.5): frontmatter `version` and `updated` (D5); HISTORY v0.1.5 entry (D3); §1.3 pointer to REQ-MRR-008 and to the lock-file premise; REQ-MRR-001 annotation (premise + verification state); REQ-MRR-004 loses its prose obligation (D1); REQ-MRR-008 added (D1); §4 new sub-section "a symbolic link at the lock path" (D2-B); new §5 "Residual risk accepted at close (unverified)" with R-1, R-2, R-3 (D2-B, D3).
- **plan.md**: §A.1 and §H requirement range 001..008; §A.2 three M5 cells marked not run; §A.3 REQ-MRR-008 reference; D-3 lock-path premise (D2-B); §F M5 heading and a decision paragraph recording option (b) (D3).
- **acceptance.md**: AC-MRR-002 names REQ-MRR-008 (D1); AC-MRR-006 class, precondition, and expected lines (M5 not run); AC-MRR-007 command 2 and its sentence pinned to `f30bb9088cde7b12612993026b5e8d2f1244e4f0` (D6); AC-MRR-009 class, Given/When/Then, and expected restated for the verb path (D4); AC-MRR-010 class and cell labels (unadopted); §2 traceability rows and counts (8 requirements, 10 criteria); §3 lock-path edge case (D2-B); §4 Definition-of-Done items 2 and 4 and the open-items line (D3, D4); §5 new evidence entry E-HEAD-002.
- **design.md**: one bullet under "Limits" stating that the lock file's identity is not defended (the codex finding cited design.md line 10).
- **Counts now**: 8 requirements (REQ-MRR-001…008), 10 criteria (AC-MRR-001…010; AC-MRR-010 unadopted). The §E.1 block above still shows the revision 0.1.3 figures (req_count 7, ac_count 9, certain_file_count 6); it is a record of that round and was not rewritten here.
- **D6 note**: the board record names the mechanical items D1, D4, D5. D6 is not named there; it is in the brief's item table and in the repair scope of the option the operator selected, and it was fixed on that basis.
- **Measured for this revision**: `git diff --exit-code 2aab5f797 f30bb9088cde7b12612993026b5e8d2f1244e4f0 -- go.mod go.sum` → exit 0, no output (acceptance.md §5, E-HEAD-002). `internal/cli/fold_store_lock_unix.go` at `f30bb9088` opens the lock path with `O_CREAT|O_RDWR|O_CLOEXEC`; a search of the two lock files for `O_NOFOLLOW`, `Lstat`, `Fstat`, and `SameFile` returned no line.
- **Not done, by instruction**: no plan audit was run and no factory verb was called. The decision record states that a further plan audit was not asked and is not granted by it. The edits change spec.md, plan.md, and acceptance.md, so the `plan_artifact_hash` of iteration 5 (`ae78cfa7…`) no longer describes these files.
- **Wait record**: `w-t1568-20261009T185625Z` is not closed by this section; closing it is the lane's or the leader's act.

## §L Lane-16 walk record — run tmnboq (2026-10-10)

- **Dispatch**: board `d-20261010T063131Z-2488` — t1568 moved to lane-16 (lease v22, evidence_sha `90f0f14c9`); records read before work: 722b (D2-B + M5-(b), no implementation change), 403d (one plan audit of `90f0f14c9` beyond the ceiling), 3f34 (drive to local-main landing, no stop), ca0c (landing hold — lane ceiling is merge-ready), 1e9b (card-branch push after landing, leader procedure), beb9 (leader lease stamp; 13c2 retracted by aac4 — the gate line must not be re-inserted).
- **Tree**: HEAD `90f0f14c9` on branch `WT-fold-rename-race`; `git status --short` empty at entry (this run).
- **Kickoff T7**: a `factory stage` call (state=run attempt) executed the plan-audit → kickoff edge — event seq 1159, evidence audited_sha `90f0f14c9`, verdict PASS (iter6), v23; the ceiling engine recorded outcome=pass-through in §G (round count 6 ≥ tier ceiling 2 + 1 delta round; REQ-ACE-013, admission-clean). The direct plan-audit → run form is an illegal transition (kickoff is the intermediate state); one refusal recorded, no state damage.
- **§E.1 machine line**: the T8a precondition `audit_ready: true` (exact machine key, no bullet) was absent from §E.1 — the section carried only the prose form `plan_status: audit-ready`; the line was added, a recording-format completion with no content change. One `factory decide` refusal is on record for this gap before the fix.
- **Kickoff T8a**: `moai factory decide t1568 --decider audit --gate kickoff --choice approve` → `t1568: run (v25)`; the edge re-leased the card to its record owner. Leader batch renewal → v26 until 08:00:50Z (07:00Z). `factory next` (MCP and CLI, parent root) refused "record stayed busy 1s: sql: transaction has already been committed or rolled back" ×3 (06:49–06:52Z) — the leader stamped the lease instead (board beb9); not re-attempted after the stamp.
- **Review-gate disposition**: an uncommitted `review_gate.enabled: false` edit was applied per standing ruling 13c2, then corrected — aac4 retracted 13c2 and the leader restored the committed value (`enabled: true`) in this tree. The first family run (below) FAILED on the t1529 CONFIG SECTIONS GUARD catching exactly that uncommitted edit; kept honest as a wrong-tree-state red, not a product regression. The line is not re-inserted.
- **Family regression (this run, HEAD `90f0f14c9`, clean tree)**: `unset MOAI_KANBAN_ID MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test -race -count=1 -timeout 900s ./internal/cli -run '^(TestMemoryFold|TestReviewArchiveUpdate|TestReviewSequentialAbandonedTempOwnership).*$'` → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli	44.489s` — the 29-test family selector of record, green on the audited tree.
- **Code identity for the sync-audit evidence**: `git diff --stat 496eab330..HEAD` lists the 5 SPEC-directory markdown files only (acceptance/design/plan/progress/spec) — zero Go, config, or CHANGELOG delta since the sync commit of record; the sync-audit verdict (PASS 0.887 binding, audited_sha `496eab330`, `.moai/reports/t1568/sync-audit.md`) therefore still describes the code at HEAD.
- **Stage walk**: T10 (run→sync) evidence is the commit carrying this record; T11 (sync→sync-audit) evidence is `496eab330:.moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/progress.md` (the sync commit of record, an ancestor of HEAD); T13 (merge-ready) is the lane's ceiling — landing (`factory complete`, merge, push, PR) stays with the leader per ca0c until the landing-resume ruling.
- **Stage-walk completion (06:57–06:59Z, machine output verbatim)**: T10 → `t1568 sync v27 lease renewed` (evidence `a8fa095456dbc9b2779364141fd6fd9b60d79d91:.moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/progress.md`); T11 → `t1568 sync-audit v28 lease renewed` (evidence `496eab3300de98f7c05608215bf0eb732e62fcfe:.moai/specs/SPEC-MEMORY-FOLD-RENAME-RACE-001/progress.md`); T13 → `t1568 merge-ready v29 lease renewed` (the sync-audit verdict PASS 0.887 binding, audited_sha `496eab330`, admitted on the label-only sync predicate).
- **Citation-line Gap (ca0c clause 4)**: the three transition results above are recorded as an uncommitted citation line while the card waits at merge-ready — no commit moves HEAD away from the walk commit `a8fa09545` at this stage; the landing-resume ruling settles this line (commit it or fold it into the landing commit).
- **Leader rollback + re-audit order (d-20261010T071213Z-2d75)**: the leader ruled the prior sync-audit PASS (written 10-09 22:34, audited_sha `496eab330`) does not represent the current criteria set — the plan revision `90f0f14c9` ADDED REQ-MRR-008 and AC-MRR-010 (§K), it did not only narrow. Leader repaired the row merge-ready v29 → sync-audit v30 (backup `factory.pre-t1568-rollback.db`) and stamped the lease; lane re-walked T12 → `sync v31`, T11 → `sync-audit v32` (evidence `a8fa09545…:progress.md`).
- **Sync re-audit at HEAD a8fa09545 — FAIL, one P2** (the one re-audit granted; board 2b7b terms: codex_audit, target baseBranch): receipt `rcpt-0edb204d43161e74eaa37f79`, verdict file `.moai/reports/t1568/sync-audit-head.md` (tool return verbatim). Finding: `internal/cli/memory_fold_wiring_test.go:1073` — the bare `<-lockHeld` receive arms on neither worker exit nor a deadline (slow-CI abandonment-before-seam hangs the package to the test timeout), and `releasePark` is closed only on the success path. Lane read-back confirmed both defects on the code. Verdict fail blocks T13 (label-only sync admission).
- **Leader disposition (d-20261010T071740Z-ca6e)**: repair path (a). The defect is in the card's OWN test file, so run re-entry was applied per board 3653 (leader repair: state=run stage=run v33, backup `factory.pre-t1568-run-reentry.db`); 722b's "no implementation change" scoped the D2-B lock-path design and does not conflict with a test-only repair. The next codex_audit is the re-audit of record; a further FAIL holds until the operator returns.
- **Test repair (manager-develop, leaf spawn)**: commit `8e5adebff6b6be8e58eaa0144c984e828f0ecd19` (parent `a8fa09545`), 1 file (`internal/cli/memory_fold_wiring_test.go`, +32/−3), test-only. Shape: the lock-held receive now selects on `lockHeld` / `wireEnv.workerExit` (aliases production `foldOnDoneExit`, closed by the worker's defer) / a 5s deadline with a Fatalf naming the fired arm; `releasePark` closes via a `sync.Once` shared by the success path and a `t.Cleanup` whose wait is gated on an atomic the seam sets (no double close, no failure-path goroutine leak). Agent evidence: family 29/29 PASS `ok … 36.028s` exit 0; `go build ./...` exit 0; gofmt empty; `go vet` exit 0.
- **Lane re-observation (this run, HEAD `8e5adebff`)**: the same family-of-record command → **exit 0**, verbatim `ok  github.com/modu-ai/moai-adk/internal/cli	28.120s`.
- **Citation-line Gap (ca0c clause 4, second walk)**: the lane records commit `705ba2175` (§L through the repair) then re-walked T10 → `sync v34`, T11 → `sync-audit v35` (evidence `705ba2175…:progress.md`), ran the re-audit of record — codex_audit PASS, receipt `rcpt-575b9b925fac5f38ac5395c0`, no findings, verdict file `.moai/reports/t1568/sync-audit-final.md` (audited_sha `705ba2175`, tool return verbatim; one machine-key fix: the parser requires the uppercase enum `verdict: PASS`, the first write used the tool's lowercase) — and T13 → **`merge-ready v36`**. These transition results are the uncommitted citation line while the card waits at merge-ready; the landing-resume ruling settles it.
- **Card-branch push (standing d-20261010T115658Z-147e)**: `git push -u origin WT-fold-rename-race` → new branch on origin (no force, no PR, main untouched); verified `origin/WT-fold-rename-race` = `705ba21756a8a907b03a7f79b5a86cac54e55373` = local HEAD. Confirmed to the leader for the sweep exclusion.

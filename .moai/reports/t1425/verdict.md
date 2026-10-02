# t1425 — completion report (lane-6)

card: t1425 (urgent, class B, leader-dispatched; no SPEC, the card text is the scope)
branch: `WT-harness-prune-single-writer` (worktree `.moai/worktrees/t1425`, base `c640b0193` = develop)
tip at report time: the commit that adds this file on top of `48f9a5f93`
integration: NOT merged — the lane asks the leader for the window (no push, no `factory complete`)
deadline: develop landing and an installed build before 2026-10-04T09:18Z (see Leader decisions 1)

## Claim

The harness observation-log retention storm is repaired in `internal/harness/retention.go`: the last prune ATTEMPT time is written to a state file next to the log (`<log>.prune-state`, the time from the injected clock) before the work starts; an exclusive `internal/lockfile` lock admits one pruner per interval; processes that find a fresh stamp return after one small file read without reading the log; lock waiters re-read the stamp with a fresh clock after the lock and skip. N concurrent hook processes now produce one archive append and one rewrite, not N, and the archive is no longer corrupted by interleaved appends.

## Evidence

Commits (`git log --format='%h %s' c640b0193..HEAD`, before this file): decision record `be6c80f16`; observed RED, own commit BEFORE the fix, tests not in it: `785d70fbf`; fix `fe211e9c9` (`retention.go` +119/−5, two new test files, one `.gitignore` line); run evidence `9fbf6d0a8`; lane measurements `0fa7f49e8`; sync-audit verdict `cbec0f35b`; stamp-before-work fix after audit finding F1 `987dcc61b` (`retention.go` +17/−16, two new test files); evidence `990a29b67`; delta-audit verdict `48f9a5f93`. Every message names `t1425`; run-phase commits carry `Authored-By-Agent: manager-develop` (a `general-purpose` spawn playing the role, the documented workaround for manager-develop spawns landing in their own tree).

Observed RED then GREEN (`red-baseline.md`, `run-evidence.md`; multi-process test `TestPruneConcurrentProcessesSingleRewrite`: the test binary re-executes itself 8 times, 20,000-event fixture): on the unmodified code 5 of 5 runs FAIL, 16 gzip headers in the archive against 2 expected, both monthly archives corrupt; after the fix 20 of 20 PASS, 2 headers, 0 corruption, 0 duplicates, second wave leaves the log unchanged. The audit re-derived the RED independently (5 of 5, 16 headers) and confirmed the corruption.

Independent audits (both committed and read by the lane): sync-audit PASS-WITH-DEBT 87/100, no blocking finding, major finding F1 = stamp written after the work (a pruner killed by the 5 s hook timeout after the archive append left no stamp, so each later hook archived the same events again); delta audit of the repair PASS-WITH-DEBT 90/100, no blocking finding: with real child processes SIGKILLed after the archive append, the old ordering added 15,000 duplicates per kill (0, 15k, 30k, 45k, 60k over five kills) and the new ordering added none after the first kill; mutants M4, M8', M9 (also re-run after an assertion change), M10, M11 each killed, exit 1; `-count=20` 20 of 20, `-race` ok, 16 packages ok, `go vet`, `golangci-lint` v2.1.6 (equals the CI pin) 0 issues, `GOOS=windows` build and vet exit 0.

Measurements the lane took itself (`lane-measurements.md`): one prune of a read-only copy of the live 65.8 MB log takes 0.53 s with nothing stale and 1.79 s with 12.5 percent stale and a full rewrite; the hook timeout is 5 s (`settings.json.tmpl` lines 147, 204, 223, 272, `async: true`). The lane re-ran the new prune tests three times itself: 30 `--- PASS`, 0 `--- FAIL`. Documentation impact: no page describes the prune interval, so no docs change.

Archive damage (`archive-damage.md`, read-only): `gzip -t` fails on `2026-07`, `2026-08` and `2026-09.jsonl.gz` while a positive control (two concatenated valid members) passes; salvage recovered 775,051 lines from `2026-09`, only 37,271 unique (95.2 percent repeats), 297,722 from `2026-08` (40.7 percent repeats), 107,392 from `2026-07` (12.4 percent repeats).

## Baseline-attribution

Lane measurements: worktree `.moai/worktrees/t1425`, tree clean, HEAD `990a29b67` at the time of the prune-duration and test runs, taken in this session; the archive checks ran against the primary checkout's live files on 2026-10-02 and are a snapshot. The judging toolchain was the machine's `go` (no installed `moai` build judged anything). Divergence at the last pre-spawn check: `git rev-list --count --left-right origin/develop...HEAD` → `0 16`, local develop → `0 0` (base current). The 65,506,619-byte log size in the decision record is a `stat` of 2026-10-02; the lane's later copy measured 65,804,810 bytes.

## Gaps

- Not run: the live 70 MB log with hundreds of hooks; a prune or a lock waiter on a machine under the incident load; a real kill by Claude Code at the 5 s hook timeout (the audit simulated the kill by a self-SIGKILL after the archive append); Windows tests (build and vet only); the repair on Linux (CI runs the new tests on Ubuntu for every PR, on macOS and Windows for release PRs only; the lane and the auditors ran darwin only).
- The `internal/cli` tests that merely mention the usage log (`update_preserve_reach_test.go`, `clifix_critical_repro_test.go`, `internal/hook/failure_event_test.go`) were run by the sync auditor (pass); the `Harness`-named `internal/cli` tests passed (ok 21.954 s run agent, ok 30.432 s auditor).
- Jev (`jev-1.13.0`) answered the two open design choices (0.96, 1.00) from facts the lane measured; the question did not mention the 5 s hook timeout, so the lock-waiter choice (`block_then_recheck`) was not tested against it; the audit judged it defensible (finding F6).
- No SPEC exists; acceptance is the card's own requirements, listed in the sync-audit acceptance table.

## Residual-risk and Leader decisions

1. **The repair does nothing until a build containing it is installed.** Hooks run the installed `moai` binary. Landing on develop is not enough: before 2026-10-04T09:18Z either an rc build with these commits must be installed (`.claude/rules/local/gitflow-lane-protocol.md` §9, operator request) or the manual 28-day prune must be repeated. Decision: leader.
2. **Existing archives are damaged** (see `archive-damage.md`): this card stops new corruption, it does not repair `2026-07`, `2026-08`, `2026-09`. Salvage or replacement is a decision about operator-visible data; the lane touched nothing. The member-level salvage recovers most lines but the corrupt members' content is gone; unique-event loss is not quantified.
3. **Orphan `usage-log-*.tmp`**: a kill in the middle of the rewrite can still leave one (disclosed in an `@MX:NOTE`); nothing sweeps them and the lane did not add deletion code (deletion is a leader decision). The 1,051 temp files of the incident came from this class.
4. **Events appended during the single pruner's window are still lost**: the pruner replaces the log by rename while other hooks append. One window per interval replaces N concurrent rewrites; it is not removed.
5. Disclosed, not changed (audit findings): F4 a symlink at `<log>.prune-state` is followed and its target truncated; F5 on Windows the lock is in-process only, so a burst at an expired stamp can still prune concurrently once per interval; F6 lock waiters have no timeout and are killed at 5 s if a prune is slower (the prune measured 1.8 s at 12.5 percent stale); F7 a state file not writable by the current user disables retention silently.
6. Optional test hardening (delta-audit N1, N2): `drain()` in `retention_killed_test.go` does a blocking FIFO open and can hang until the `go test` timeout if the pruner never reaches the archive step (only on a regression); dropping `Truncate(0)` in `writeStamp` and ignoring the stamp-write error both survive the whole package.
7. Pre-existing and out of scope, reported only: `partitionEvents` drops lines that fail JSON parsing although its comment says they are kept.
8. CHANGELOG: not written by the lane (leader decision at release time).

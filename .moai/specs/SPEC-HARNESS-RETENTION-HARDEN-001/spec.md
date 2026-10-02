---
id: SPEC-HARNESS-RETENTION-HARDEN-001
title: "Harness usage-log retention hardening — state-file symlink and permission faults, late-event loss during prune, lock-limit disclosure, and the two test gaps the t1425 audits left open"
version: "0.1.0"
status: draft
created: 2026-10-03
updated: 2026-10-03
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/harness"
lifecycle: spec-anchored
tags: "harness, retention, prune, usage-log, flock, symlink, event-loss, test-gap, card-t1432"
tier: M
card: t1432
related_specs: [SPEC-AGENT-TEAM-RETIRE-001]
---

# SPEC-HARNESS-RETENTION-HARDEN-001 — Harness retention audit-debt bundle (card t1432)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-10-03 | manager-spec | Initial plan-phase draft for card t1432 on `WT-harness-retention-debt`, base develop `1e2151a38`. Origin: the debt left open by the t1425 audits (`.moai/reports/t1425/sync-audit.md` §3 F4-F7, `.moai/reports/t1425/sync-audit-delta.md` §3 N1, N2 and §1 Residual-risk, completion report `.moai/reports/t1425/verdict.md`). Seven items, exactly those named in the card; the other t1425 findings are Out of Scope (§E). Leader scope-narrowing applied: F5 and F6 are labelled "not reproduced, not measured" and default to disclosure only, and every option that changes the shared `internal/lockfile` contract stays in §B as a non-default, operator-held option. 12 REQ / 15 AC. |

**Tier M.** The change set is small in code (the production edits are confined to `internal/harness/retention.go`, plus test files) but carries 12 requirements and 15 acceptance criteria, which exceeds the Tier S ceilings (8 / 8) and sits inside the Tier M ceilings (16 / 16). Production file count: 1; test files: 1 edited, up to 3 new; one run-phase evidence file `.moai/reports/t1432/red-baseline.md`. Tier M artifacts: spec.md, plan.md, acceptance.md, plus progress.md and decision-index.md (the decision gate is on in this repository, `.moai/config/sections/interview.yaml:6`). plan-auditor threshold 0.80.

## §A Background

The usage log `usage-log.jsonl` is appended by the harness-observe hooks (`internal/harness/observer.go`) and pruned lazily from the same hook process: after the append, `PruneStaleEntries` runs and, when the on-disk attempt stamp `<log>.prune-state` is older than one hour, takes an exclusive `flock` on that file, writes the stamp, archives events older than 30 days into `<YYYY-MM>.jsonl.gz`, and replaces the log by rename. Card t1425 (urgent, class B, no SPEC) made this single-writer; its audits left the seven items below open. The stamp-before-work ordering (t1425 F1) and the orphan `usage-log-*.tmp` sweep are already on this tree and are not touched here.

Evidence labels used in this SPEC (a premise carries exactly one):

- **OBSERVED** — measured by the plan author on this tree (HEAD `1e2151a38`, darwin, uid 501) with a scratch probe injected through `go test -overlay`; the probe lives outside the tree and is not part of this SPEC's change.
- **READ** — observed by reading the named file and line in this run; no runtime behaviour is claimed.
- **AUDIT-REPORTED** — measured by the t1425 sync audit, not re-observed here.
- **NOT REPRODUCED, NOT MEASURED** — no reproduction and no measurement exists; the claim is a hypothesis.

| Item | Premise | Label | Evidence |
|---|---|---|---|
| F4 | A symbolic link at `<log>.prune-state` is followed; its target is truncated and overwritten with the stamp. | OBSERVED | `PROBE-F4 err=<nil> victim_changed=true victim="2026-10-02T00:00:00Z"` (a 64-byte victim file behind the link was replaced by the stamp). Cause READ at `retention.go:121` (`os.OpenFile(statePath, O_RDWR\|O_CREATE, 0o644)`, no Lstat, no no-follow flag). |
| F5 | On Windows `lockfile.Lock` is an in-process mutex, so a burst of hook processes that all find no fresh stamp can prune concurrently, once per interval. | NOT REPRODUCED, NOT MEASURED (burst); lock shape READ | Lock shape READ at `lockfile_windows.go:26-41` (per-path `sync.Mutex` map, no cross-process primitive). The concurrent-prune burst itself has no reproduction (no Windows runtime in this environment); it is also AUDIT-REPORTED (`sync-audit.md` F5). |
| F6 | Lock waiters block with no timeout and no try-lock; hooks run with a 5 s timeout and `async: true`. | NOT REPRODUCED, NOT MEASURED (waiter delay); code shape READ | API shape READ (`lockfile_unix.go` exports only `Lock` = `LOCK_EX` blocking, and `Unlock`); hook `timeout: 5`, `async: true` READ (`internal/template/templates/.claude/settings.json.tmpl:146-148` and `:203-205`); event appended **before** the prune call READ (`observer.go:86` write, `:93` prune). The waiter's actual wait and kill behaviour was not reproduced or measured (the audit's `AUDIT-F` "blocked until release" is AUDIT-REPORTED). |
| F7 | A state file that exists but is not writable by the current user makes every prune call return an open error; the observer discards it, so retention is silently and permanently off. | OBSERVED | uid 501, state file mode 0400: first call and a call one day later both return `retention: prune state open failed: ... permission denied`, `stale_still_in_log=true`. Discard READ at `observer.go:93` and `:141` (`_ = o.retention.PruneStaleEntries(...)`). |
| Event loss | Events appended by other hooks between the pruner's read and its rename are lost. | OBSERVED | Pruner blocked in its archive step (FIFO as in `retention_killed_test.go`), a second writer appends `late-event` through `appendEventsJSONL`, pruner finishes: `PROBE-LOSS late_event_present=false fresh_present=true`. Mechanism READ (`retention.go` `prune`: `partitionEvents` then `archiveEvents` then `overwriteWithEvents`; appenders take no lock). |
| N1 | `drain()` in `retention_killed_test.go` opens the FIFO with a blocking `O_RDONLY`; if the pruner never reaches its archive step the test hangs until the go test timeout. | AUDIT-REPORTED (hang); open READ | Hang measured by the audit with mutant mD (`panic: test timed out after 40s`, goroutine in `syscall.Open`), not re-run here; the blocking open is READ at `retention_killed_test.go:43`. |
| N2 | Removing `Truncate(0)` from `writeStamp` (mB) and ignoring the stamp-write error in `pruneExclusive` (mC) both survive the package's tests. | AUDIT-REPORTED | mB `ok ... 1.812s` and mC `ok ... 3.087s` in the audit's mutation runs, not re-run here; the two branches are uncovered in the audit's coverage run (`writeStamp 75.0%`, `pruneExclusive 85.0%`). |

Three further facts bound the design and were observed this run:

- **`internal/lockfile` has seven non-test importers on this tree** (grep of `internal/lockfile"`): `internal/contract/receipt/store.go`, `internal/cli/settings.go`, `internal/cli/glm_tools.go`, `internal/cli/taskledger/taskledger.go`, `internal/harness/retention.go`, `internal/escalation/lock.go`, `internal/hook/agentmemory.go`. The `@MX:ANCHOR` text in `lockfile_unix.go` names only three; any change to that package is a change for all seven.
- **`SPEC-AGENT-TEAM-RETIRE-001` is `status: completed` (v0.1.2)**; its REQ-ATR-001 requires the Windows in-process-mutex limitation comment preserved verbatim, and the `@MX:NOTE` in `lockfile_windows.go` says behaviour preservation is the contract and the file is not to be silently upgraded to `LockFileEx`. A change there goes through that SPEC's amendment path (`completed → in-progress (amendment)` with `amendment_of:` and an `## Amendments` row), not through this SPEC.
- **FIFO as a state path cannot exercise the stamp-write branch on this platform** (OBSERVED, scratch probe, darwin): opening a FIFO `O_RDWR` succeeds, `flock` on it returns `operation not supported`, so a prune stops at the lock step before `writeStamp` is reached. This is why §B D5 carries a seam decision.

## §B Decisions

Each design decision states its options with quantified costs, the chosen default (smallest blast radius), and an `operator-held` flag per option. `operator-held: yes` means the option changes a shared package contract or doctrine-level behaviour and must go to the operator before Kickoff if anyone wants it. Defaults live here; `decision-index.md` carries the neutral question rows.

### D1 — F5, Windows cross-process lock  (premise: not reproduced, not measured)

| Option | Files touched | New exported API | Platforms | Effect on the other 6 lockfile importers | operator-held |
|---|---|---|---|---|---|
| **A. Disclosure only (default)** — restate the limitation precisely in the pruner's source documentation and label it "not reproduced, not measured". | 1 (`retention.go` comment; already touched by other items) | 0 | none changed | none | no |
| B. Retention-local cross-process guard on Windows — a Windows-only build-tagged sentinel file (`<log>.prune-lock`, created exclusively) around the prune. | 2 new (`retention_lock_windows.go` plus a non-Windows no-op twin) plus tests | 0 (package-private) | Windows | none | no (non-default: no reproduction exists, so no code repair is justified yet) |
| C. Upgrade `lockfile_windows.go` to a cross-process lock (`LockFileEx`). | `lockfile_windows.go`, its tests, plus an amendment to SPEC-AGENT-TEAM-RETIRE-001 (≥ 3) | 0 | Windows | **all 6 others change behaviour on Windows** (in-process mutex becomes a cross-process lock; ClaimTask, settings, glm tools, receipt store, escalation lock, agent memory); cannot be runtime-verified here | **yes** |

Default: **A**. Why: the premise has no reproduction and no measurement, the t1425 stamp-before-work reorder already shrank the window to the interval between the stamp read and the stamp write, and B and C add code whose failure modes (a stale sentinel after a 5 s kill; a changed Windows lock for seven packages) are larger than the loss they remove. A closes the item as a disclosure correction, which is what the audit asked for ("Disclosure refinement, not a code change", `sync-audit.md` F5).

### D2 — F6, lock-wait behaviour  (premise: not reproduced, not measured)

| Option | Files touched | New exported API | Platforms | Effect on other lockfile importers | Effect on the hook | operator-held |
|---|---|---|---|---|---|---|
| **A. Record only (default)** — keep blocking waiters; record that the 5 s hook-timeout fact was missing from the earlier design question and that the event is appended before the wait. | 1 (`retention.go` comment) | 0 | none | none | unchanged: a waiter lingers until the holder's prune ends; the holder's prune took 0.53 s (nothing stale) and 1.79 s (12.5 percent stale) on a 65.8 MB log (t1425 lane, reported in the card dispatch, not re-measured here) | no |
| B. Add a non-blocking try-lock to `internal/lockfile`: Unix `LOCK_EX\|LOCK_NB` mapping `EWOULDBLOCK` to "not acquired", Windows `sync.Mutex.TryLock`. | 4-6 (`lockfile_unix.go`, `lockfile_windows.go`, their tests, `retention.go` call site, a retention test) | 1 (`TryLock`) | Unix and Windows | existing `Lock`/`Unlock` byte-identical; the `@MX:ANCHOR` contract text and REQ-ATR-001 ("behaviour preservation") need a leader/operator ruling on whether an additive symbol needs an amendment | waiter returns at once instead of waiting out the prune | **yes** (changes a shared package's exported surface) |
| C. Bounded wait: poll the try-lock with a sleep up to a cap. | B plus a cap constant | 1 | both | as B | wait capped at the cap, which adds up to the cap to the hook's life | **yes** |

Default: **A**. Why: the wait delays only the hook's exit, not its event (READ: `observer.go:86` precedes `:93`), the hooks are `async: true`, and the premise has no measurement; B and C change the exported surface of a package with seven importers. The recorded fact is the correction the audit requested ("the record should say the timeout fact was missing from the question", `sync-audit.md` F6).

### D3 — Event loss between the pruner's read and its rename  (premise: OBSERVED)

| Option | Files touched | New exported API | Hot append path | Windows | operator-held |
|---|---|---|---|---|---|
| A. Appenders take the state-file lock around every append. | `observer.go` (three append sites: `:80`, `:129`, `appendEventsJSONL` `:270`), `retention.go`, tests (≥ 4) | 0-1 | **changed**: each append adds an open, a lock, an unlock and a close of the state file (four system calls, counted not timed), and an appender waits out the whole prune (0.53-1.79 s reported by the t1425 lane for 65.8 MB, unbounded for a larger log). With the 5 s hook timeout, an append queued behind a prune slower than 5 s is killed before it writes: the event is lost, which is worse than today. | lock is in-process only: no cross-process effect | **yes** (hot path of every observe hook; doctrine-level) |
| **B. Tail-carry (default)** — the pruner partitions only the prefix of the log it measured, and immediately before the rename copies every byte the log gained after that prefix into the replacement. | 1 production (`retention.go`) plus tests | 0 | **unchanged** (zero added latency, no lock, no syscall) | works (no locks involved); rename-over-open-file semantics are pre-existing | no |
| C. Documented residual risk only. | 1 (`retention.go` comment) | 0 | unchanged | unchanged | no |

Default: **B**. Why: it is the only option that closes the OBSERVED `late-event` loss without touching the hot append path or any lock contract. It narrows the loss window from "the whole prune" (read, archive, rewrite; the archive step alone was blockable by a FIFO in the probe) to the gap between the final size reading and the rename; it does **not** reach zero, and REQ-HRH-006 requires the residual window to be stated. C closes nothing; A closes the window fully on Unix but adds latency and a new kill-before-append failure mode on every hook. Options that close the loss entirely touch the hot append path (A), which is operator-held; B is not.

### D4 — F4 and F7, the state-file path  (premises: OBSERVED)

| Option | Files touched | New exported API | Platforms | Behaviour change | operator-held |
|---|---|---|---|---|---|
| **A. Replace-on-fault (default)** — before opening, inspect the state path without following links; a symbolic link is removed (the link, never its target) and recreated as a regular file; a regular file that opens with a permission error is removed (it is derived 35-byte state in the project's own directory) and recreated; anything else (a directory, a FIFO, an unremovable file) keeps today's behaviour: skip, return the error. One mechanism closes both items. | 1 (`retention.go`) plus tests | 0 | all (Lstat and Remove are portable) | a hostile or foreign-owned state file no longer disables or abuses retention; one possible duplicate prune when two processes heal the same fault at the same instant (two inodes locked once) | no |
| B. Refuse-only for F4 (no-follow open flag, build-tagged) and visibility-only for F7 (return a typed error the observer logs). | ≥ 3 (two build-tagged files for the flag, `observer.go` for logging) plus tests | 0-1 | the flag does not exist on Windows (`syscall.O_NOFOLLOW` is undefined there, checked with `GOOS=windows go doc`), so a no-op twin is required | retention stays off while the fault persists (F7 not closed, only audible); touches `observer.go` | no, but it touches the hot path file |

Default: **A**.

### D5 — N2, pinning the stamp-write failure branch (mutant mC)

| Option | Files touched | New exported API | Behaviour change | operator-held |
|---|---|---|---|---|
| **A. Seam (default)** — an unexported function-valued field on the pruner instance, defaulting to the real stamp writer, replaceable by an in-package test (parallel-safe, unlike a package variable). | `retention.go` (1 field, 1 default line, 1 call site) plus the test | 0 | none | no |
| B. Accept mutant mC as a recorded survivor (no test). | 0 | 0 | none; N2 stays half-closed | no |

Default: **A**. Why: no file-system fault available on both CI platforms makes `Truncate` or the write fail after `open` and `flock` succeed (a FIFO stops at `flock`, OBSERVED on darwin; `/dev/full` is Linux-only), and the card asks for both mutants to be pinned. mB needs no seam: it is killed by file content alone.

## §C Requirements (GEARS)

- **REQ-HRH-001** (Event-driven, item F4): When the prune state path (`<log>.prune-state`) is a symbolic link, the retention pruner shall neither open, truncate nor write the link's target, and shall continue the prune using a regular state file it creates at that path.
- **REQ-HRH-002** (Event-driven, item F7): When the prune state path is a regular file that the current user cannot open for reading and writing and whose directory permits the current user to remove it, the retention pruner shall replace it with a regular file it can lock and shall continue the prune.
- **REQ-HRH-003** (Event-driven, item F7 boundary): When the prune state path cannot be opened for reading and writing and is neither a symbolic link nor a regular file the current user can remove, the retention pruner shall skip the prune, return an error naming the failure, and leave the usage log and the archive directory unmodified.
- **REQ-HRH-004** (Event-driven, item event loss): When event lines are appended to the usage log after the retention pruner finished reading it and before the pruner takes its final reading of the log's size ahead of the replacement, the pruner shall carry those bytes, verbatim and in their original order, after the kept lines of the replacement log.
- **REQ-HRH-005** (Ubiquitous, item event loss): The observer's event-append path shall add no lock, no file open and no system call beyond those present at base `1e2151a38`.
- **REQ-HRH-006** (Ubiquitous, item event loss): The retention pruner's source documentation shall state that a window remains between its final size reading and the replacement rename in which an appended event is lost, and shall not state that events appended during the whole read-to-rename interval are lost.
- **REQ-HRH-007** (Ubiquitous, item F5): The retention pruner's source documentation shall state that on Windows the lock gives no cross-process exclusion, that a burst of hook processes finding no fresh stamp may prune concurrently once per interval, and that this is "not reproduced, not measured".
- **REQ-HRH-008** (Ubiquitous, item F6): The retention pruner's source documentation shall state that lock waiters block without a timeout, that the harness-observe hooks run with a 5 s hook timeout and `async`, that the event is appended before the wait begins, that the waiter delay is "not reproduced, not measured", and that the earlier lock-behaviour design question did not carry the hook-timeout fact.
- **REQ-HRH-009** (Ubiquitous, items F5 and F6): The change set shall leave every file under `internal/lockfile` byte-identical to base `1e2151a38`, and the packages `internal/harness` and `internal/lockfile` shall build and vet for `GOOS=windows`.
- **REQ-HRH-010** (Event-driven, item N1): When the pruner under test never reaches its archive step, the killed-pruner test shall fail within 15 seconds with a message naming the missing archive step, rather than waiting for the go test timeout.
- **REQ-HRH-011** (Ubiquitous, item N2 / mB): The retention test suite shall fail when the stamp write does not truncate a longer previous stamp, by asserting on the exact file content of a stamp written over a longer one.
- **REQ-HRH-012** (Ubiquitous, item N2 / mC): The retention test suite shall fail when the pruner proceeds to prune after the stamp write fails, by asserting that a failed stamp write returns an error and leaves the usage log unmodified.

## §D Constraints

- **Baseline-first ordering** (`verification-claim-integrity.md` §2.3): the observed-RED artifact for every behaviour-changing criterion lands in its own commit that precedes the commit changing the behaviour. Encoded in `plan.md` M0.
- **Windows**: no Windows runtime is available. Every Windows-facing criterion is verified by `GOOS=windows go build` and `go vet` only, and by documentation; none claims a runtime observation.
- **Hook budget**: harness-observe hooks run with a 5 s timeout and `async: true`. One prune of a 65.8 MB log took 0.53 s (nothing stale) and 1.79 s (12.5 percent stale) (t1425 lane, reported in the card dispatch, not re-measured). No requirement here may add work to the append path (REQ-HRH-005).
- **No new configuration key, no new exported symbol, no new dependency.**

## §E Exclusions

### Out of Scope — other t1425 findings and history

- F1 (stamp-before-work), F2, F3 (already on this tree: stamp ordering, `sweepOrphanTmp`, and their tests), F8 and F9 (coverage and IDE-hint notes).
- Damaged historical archives (`.moai/reports/t1425/archive-damage.md`): no repair or rewrite of existing `<YYYY-MM>.jsonl.gz` files.
- `partitionEvents` behaviour for lines that fail JSON parsing: unchanged.
- `CHANGELOG.md`: owned by the sync phase.

### Out of Scope — shared package and doctrine changes (non-default options only)

- Any change to `internal/lockfile` (D1 option C, D2 options B and C): present in §B as non-default, operator-held options; not authorized by this SPEC. REQ-HRH-009 forbids it.
- Appenders taking a lock (D3 option A): not authorized; REQ-HRH-005 forbids it.
- Moving the prune off the hook path, changing the one-hour interval or the 30-day retention window, or adding configuration for either.
- Any change to `internal/harness/observer.go`.

### Out of Scope — verification limits

- Windows runtime behaviour of any kind: build and vet only.
- Reproducing F5 or F6 (a Windows burst; a waiter killed at 5 s): not attempted; revisit only if a reproduction is produced.

## §F Risks and unobserved items

- **Partial last line at the tail boundary** (D3 B): an appender's single write may be mid-flight when the pruner takes its size reading. Mitigation: the boundary is the last newline-terminated line the pruner classified, so a half-written line lands entirely in the tail. Whether a concurrent `O_APPEND` write can be observed half-complete on the target filesystems is not observed; `plan.md` M1 carries an acceptance test for the boundary rule rather than a claim about the kernel.
- **Residual loss window remains** (D3 B): an event appended between the final size reading and the rename, or by a writer that opened the old inode before the rename, is still lost. Stated in REQ-HRH-006; not eliminated.
- **Heal burst** (D4 A): two processes healing the same fault at the same instant can each lock a different inode and prune once concurrently. One-time; bounded by the stamp afterwards. Not measured.
- **F7 persistence without a writable directory**: if the directory itself is not writable the state file cannot be replaced and retention stays off silently (the observer discards the error by design); REQ-HRH-003 pins the skip and the error, not an alarm.
- **N1 fix choice**: the audit's suggested non-blocking read open (`O_RDONLY|O_NONBLOCK`) is unverified here and can race a pruner that has not yet reached its write-open (the reader sees end-of-file early and the writer then blocks); `plan.md` M4 names the timer-and-unblock form as the alternative. The criterion binds the outcome (fail within 15 s), not the mechanism.
- **Not observed in this plan phase**: N1/N2 mutants (mD, mB, mC) were not re-run; the Windows paths; any measurement of F5 or F6.

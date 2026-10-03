---
id: SPEC-HARNESS-RETENTION-HARDEN-001
title: "Harness usage-log retention hardening — state-file symlink and permission faults, late-event loss during prune, lock-limit disclosure, and the two test gaps the t1425 audits left open"
version: "0.3.0"
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
| 0.2.0 | 2026-10-03 | manager-spec | Revision for plan-audit iteration 2 after iteration 1 FAIL 0.75 (`.moai/reports/t1432/plan-audit.md`, findings D1-D14). Operator decisions relayed by the leader session on 2026-10-03 (not a direct operator answer): defaults stand with no change to `internal/lockfile`; event loss follows D3-B plus a source disclosure of the residual window; F5 and F6 stay "not reproduced, not measured" disclosure only; Q4 is restricted so that the pruner deletes only a state-path entry owned by the current user, otherwise it leaves the entry byte-identical, skips the prune and warns; one unexported test-only field on the pruner is allowed. Changes: ownership restriction and warning surface (REQ-001..004, D4); heal safety against a concurrent healer (REQ-005, D4.c); tail terminator and partial-line rule (REQ-007, D3); explicit cross-platform clause (REQ-012); disclosure sentinels split per obligation and card history removed from source comments (REQ-009..011); coverage figures attributed to the audit that measured each; 15 REQ / 14 AC. Plan changes (baseline-first split, per-AC RED-now cells) are in `plan.md` and `acceptance.md`. |
| 0.3.0 | 2026-10-03 | manager-spec | Revision for plan-audit iteration 3 (the last), after iteration 2 FAIL 0.87 (`.moai/reports/t1432/plan-audit-iter2.md`: blocking B1-B3, optional O1-O10). B1: the ordering check is bounded to `1e2151a38..HEAD` and carries positive controls (`plan.md` M0 Exit, `acceptance.md` Definition of Done 1, ledger E-021 to E-024). B2: the link-owner rule is stated in REQ-HRH-001, -003 and -004 (the owner recorded on the link itself, never its target's); AC-HRH-005 gains cases (d) and (e) and the owner-read-follows-the-link mutant. B3: a FIFO at the state path hangs the lock-free pre-check (OBSERVED, ledger E-025), so REQ-HRH-003 and D4.a no longer claim a skip for a FIFO; the hang is recorded as a pre-existing, unrepaired condition outside this card (§E, §F). Optional items taken: O1, O2, O3, O5, O6, O7, O10; left open: O4, O8, O9 (one-line reasons in `progress.md`). Relayed by the leader session on 2026-10-03 (not an operator answer): the force-added card evidence, the intentional red M0 commit and the conservative Windows rule are accepted, and the completion report states "Windows runtime not observed" and names the red commit SHA (§D). 15 REQ / 14 AC. |

**Tier M.** The change set is small in code (production edits: `internal/harness/retention.go` plus two build-tagged owner-check files; tests: one edited, five new) but carries 15 requirements and 14 acceptance criteria, which exceeds the Tier S ceilings (8 / 8) and sits inside the Tier M ceilings (16 / 16). Files affected: 9 (3 production, 6 test). Tier M artifacts: spec.md, plan.md, acceptance.md, plus progress.md and decision-index.md (the decision gate is on in this repository, `.moai/config/sections/interview.yaml:6`). plan-auditor threshold 0.80.

## §A Background

The usage log `usage-log.jsonl` is appended by the harness-observe hooks (`internal/harness/observer.go`) and pruned lazily from the same hook process: after the append, `PruneStaleEntries` runs and, when the on-disk attempt stamp `<log>.prune-state` is older than one hour, takes an exclusive `flock` on that file, writes the stamp, archives events older than 30 days into `<YYYY-MM>.jsonl.gz`, and replaces the log by rename. Card t1425 (urgent, class B, no SPEC) made this single-writer; its audits left the seven items below open. The stamp-before-work ordering (t1425 F1) and the orphan `usage-log-*.tmp` sweep are already on this tree and are not touched here.

Evidence labels used in this SPEC (a premise carries exactly one):

- **OBSERVED** — measured by the plan author on this tree (HEAD `db6d88a2a`, whose code is identical to base `1e2151a38`: `git diff --quiet 1e2151a38 db6d88a2a -- internal cmd` exit 0; darwin, uid 501) with a scratch probe injected through `go test -overlay`; the probe lives outside the tree and is not part of this SPEC's change. Plan-audit iteration 1 re-observed F4, F7 and the event loss independently.
- **READ** — observed by reading the named file and line in this run; no runtime behaviour is claimed.
- **AUDIT-REPORTED** — measured by a t1425 audit; where a figure is quoted, the audit that measured it is named.
- **NOT REPRODUCED, NOT MEASURED** — no reproduction and no measurement exists; the claim is a hypothesis.

| Item | Premise | Label | Evidence |
|---|---|---|---|
| F4 | A symbolic link at `<log>.prune-state` is followed; its target is truncated and overwritten with the stamp. | OBSERVED | `PROBE-F4 err=<nil> victim_changed=true victim="2026-10-02T00:00:00Z"` (a 64-byte victim file behind the link was replaced by the stamp); re-run as the draft AC-HRH-001 test, ledger E-001 in `acceptance.md`. Cause READ at `retention.go:121` (`os.OpenFile(statePath, O_RDWR\|O_CREATE, 0o644)`, no Lstat, no no-follow flag). |
| F5 | On Windows `lockfile.Lock` is an in-process mutex, so a burst of hook processes that all find no fresh stamp can prune concurrently, once per interval. | NOT REPRODUCED, NOT MEASURED (burst); lock shape READ | Lock shape READ at `lockfile_windows.go:26-41` (per-path `sync.Mutex` map, no cross-process primitive). The concurrent-prune burst itself has no reproduction (no Windows runtime in this environment); it is also AUDIT-REPORTED (`sync-audit.md` F5). |
| F6 | Lock waiters block with no timeout and no try-lock; hooks run with a 5 s timeout and `async: true`. | NOT REPRODUCED, NOT MEASURED (waiter delay); code shape READ | API shape READ (`lockfile_unix.go` exports only `Lock` = `LOCK_EX` blocking, and `Unlock`); hook `timeout: 5`, `async: true` READ (`internal/template/templates/.claude/settings.json.tmpl:146-148` and `:203-205`, inside the opt-in observer-hook blocks); event appended **before** the prune call READ (`observer.go:86` write, `:93` prune). The waiter's actual wait and kill behaviour was not reproduced or measured (the audit's `AUDIT-F` "blocked until release" is AUDIT-REPORTED). |
| F7 | A state file that exists but is not writable by the current user makes every prune call return an open error; the observer discards it, so retention is silently and permanently off. | OBSERVED | uid 501, state file mode 0400: first call and a call one day later both return `retention: prune state open failed: ... permission denied`, `stale_still_in_log=true`; re-run as the draft AC-HRH-002 test, ledger E-002. Discard READ at `observer.go:93` and `:141` (`_ = o.retention.PruneStaleEntries(...)`). |
| Event loss | Events appended by other hooks between the pruner's read and its rename are lost. | OBSERVED | Pruner blocked in its archive step (FIFO as in `retention_killed_test.go`), a second writer appends `late-event` through `appendEventsJSONL`, pruner finishes: `PROBE-LOSS late_event_present=false fresh_present=true`; re-run as the draft AC-HRH-007 test, ledger E-003. Mechanism READ (`retention.go` `prune`: `partitionEvents` then `archiveEvents` then `overwriteWithEvents`; appenders take no lock). |
| Log terminator | The baseline always leaves the rewritten log newline-terminated (`partitionEvents` and `overwriteWithEvents` normalize terminators), so the next append starts a new line. | OBSERVED (plan-audit iteration 1) | `PROBE-UNTERM err=<nil> ends_with_newline=true` (`.moai/reports/t1432/plan-audit.md`, premise probe). |
| FIFO at the state path | A FIFO at `<log>.prune-state` makes `PruneStaleEntries` skip the prune and return an error (the earlier premise: "a prune stops at the lock step"). | OBSERVED, and the earlier premise is FALSE | `PruneStaleEntries` does not return: `PROBE-B3 BLOCKED: PruneStaleEntries did not return within 3s on a FIFO state path` (base code, darwin, uid 501; ledger E-025; first found by plan-audit iteration 2, finding B3, which measured a 5 s block). Cause READ at `retention.go:94` (`stampIsFresh(readStampFile(statePath), now)`) and `:187-189`: the lock-free pre-check opens the entry with a blocking read-only `os.Open`, which blocks on a FIFO that has no writer, before the lock step is reached. A symbolic link to a FIFO reaches the same open (`os.Open` follows links): READ, not observed. The condition is pre-existing, is not repaired here (REQ-HRH-008 forbids a call added to that path) and is outside this card (§E, §F). |
| N1 | `drain()` in `retention_killed_test.go` opens the FIFO with a blocking `O_RDONLY`; if the pruner never reaches its archive step the test hangs until the go test timeout. | OBSERVED (re-run this plan phase) | Mutant mD (pruner stamps, then returns without pruning) with the unmodified test: `panic: test timed out after 40s`, goroutine blocked in the system open call, `FAIL ... 40.791s`, exit 1 (ledger E-009). The blocking open is READ at `retention_killed_test.go:43`. |
| N2 | Removing `Truncate(0)` from `writeStamp` (mB) and ignoring the stamp-write error in `pruneExclusive` (mC) both survive the package's tests. | OBSERVED (re-run this plan phase) | mB: whole-package run with the mutant applied `ok ... 3.170s`, exit 0 (E-006); mC: `ok ... 9.821s`, exit 0 (E-008). The two branches were uncovered in the t1425 audits' coverage runs: `writeStamp 75.0%` in both, and `pruneExclusive 85.0%` in the first audit (`sync-audit.md`, package 87.3%) versus `pruneExclusive 87.5%` in the delta audit (`sync-audit-delta.md`, package 87.5%, the later and current measurement). |

Facts about the platform primitives the design rests on, all OBSERVED by scratch probes on darwin this plan phase (Linux and Windows not observed):

- `Lstat` of a symbolic link reports the link's own owner (uid 501), not its target's (uid 0, a link to `/etc/hosts`). Re-observed in this revision with a user-created link to `/`: the link's own owner uid 501, the target's owner uid 0, effective user id 501 (ledger E-026); the iteration-2 audit measured the same (its probe P-OWN). Root-owned symbolic links exist at `/var`, `/tmp` and `/etc` on darwin (ledger E-027; Linux not observed).
- A file opened read-only fails `Truncate(0)` (`invalid argument`) and `WriteAt` (`bad file descriptor`), so the stamp write fails without any file-system fault.
- An exclusive create over a dangling symbolic link fails (`file exists`) and does not create the link's target.
- A user without privilege cannot make a file owned by another user (`chown 0 <file>`: `Operation not permitted`), so a portable test cannot construct a foreign-owned state entry without a seam.

Three further facts bound the design:

- **`internal/lockfile` has seven non-test importers on this tree** (grep of `internal/lockfile"`): `internal/contract/receipt/store.go`, `internal/cli/settings.go`, `internal/cli/glm_tools.go`, `internal/cli/taskledger/taskledger.go`, `internal/harness/retention.go`, `internal/escalation/lock.go`, `internal/hook/agentmemory.go`. The `@MX:ANCHOR` text in `lockfile_unix.go` names only three; any change to that package is a change for all seven.
- **`SPEC-AGENT-TEAM-RETIRE-001` is `status: completed` (v0.1.2)**; its REQ-ATR-001 requires the Windows in-process-mutex limitation comment preserved verbatim, and the `@MX:NOTE` in `lockfile_windows.go` says behaviour preservation is the contract and the file is not to be silently upgraded to `LockFileEx`. A change there goes through that SPEC's amendment path (`completed → in-progress (amendment)` with `amendment_of:` and an `## Amendments` row), not through this SPEC.
- **A FIFO as a state path cannot exercise the stamp-write branch** (an earlier scratch probe on darwin that called the open and the lock directly, not the production entry point: opening a FIFO `O_RDWR` succeeds and `flock` on it returns `operation not supported`). Through `PruneStaleEntries` the lock step is never reached, because the call blocks earlier, in the lock-free pre-check (the FIFO row above, OBSERVED). The read-only-handle fact above is the portable alternative for the stamp-write branch (§B D5).

## §B Decisions

Each design decision states its options with quantified costs, the chosen default (smallest blast radius), and an `operator-held` flag per option. `operator-held: yes` means the option changes a shared package contract or doctrine-level behaviour and must go to the operator before Kickoff if anyone wants it. Defaults live here; `decision-index.md` carries the neutral question rows and the relayed verdicts.

### D1 — F5, Windows cross-process lock  (premise: not reproduced, not measured)

| Option | Files touched | New exported API | Platforms | Effect on the other 6 lockfile importers | operator-held |
|---|---|---|---|---|---|
| **A. Disclosure only (default; stands per the relayed verdict)** — restate the limitation precisely in the pruner's source documentation and label it "not reproduced, not measured". | 1 (`retention.go` comment; already touched by other items) | 0 | none changed | none | no |
| B. Retention-local cross-process guard on Windows — a Windows-only build-tagged sentinel file (`<log>.prune-lock`, created exclusively) around the prune. | 2 new (a Windows-tagged file plus a non-Windows no-op twin) plus tests | 0 (package-private) | Windows | none | no (non-default: no reproduction exists, so no code repair is justified yet) |
| C. Upgrade `lockfile_windows.go` to a cross-process lock (`LockFileEx`). | `lockfile_windows.go`, its tests, plus an amendment to SPEC-AGENT-TEAM-RETIRE-001 (≥ 3) | 0 | Windows | **all 6 others change behaviour on Windows** (in-process mutex becomes a cross-process lock; ClaimTask, settings, glm tools, receipt store, escalation lock, agent memory); cannot be runtime-verified here | **yes** |

Default: **A**. Why: the premise has no reproduction and no measurement, the t1425 stamp-before-work reorder already shrank the window to the interval between the stamp read and the stamp write, and B and C add code whose failure modes (a stale sentinel after a 5 s kill; a changed Windows lock for seven packages) are larger than the loss they remove. A closes the item as a disclosure correction, which is what the audit asked for ("Disclosure refinement, not a code change", `sync-audit.md` F5).

### D2 — F6, lock-wait behaviour  (premise: not reproduced, not measured)

| Option | Files touched | New exported API | Platforms | Effect on other lockfile importers | Effect on the hook | operator-held |
|---|---|---|---|---|---|---|
| **A. Record only (default; stands per the relayed verdict)** — keep blocking waiters; state the technical facts in the source documentation (blocking wait, 5 s hook timeout, `async`, event appended before the wait). The history point — that the earlier design question omitted the 5 s hook-timeout fact — lives in the decision record (`decision-index.md` Q2) and `progress.md`, not in source. | 1 (`retention.go` comment) | 0 | none | none | unchanged: a waiter lingers until the holder's prune ends; the holder's prune took 0.53 s (nothing stale) and 1.79 s (12.5 percent stale) on a 65.8 MB log (t1425 lane, reported in the card dispatch, not re-measured here) | no |
| B. Add a non-blocking try-lock to `internal/lockfile`: Unix `LOCK_EX\|LOCK_NB` mapping `EWOULDBLOCK` to "not acquired", Windows `sync.Mutex.TryLock`. | 4-6 (`lockfile_unix.go`, `lockfile_windows.go`, their tests, `retention.go` call site, a retention test) | 1 (`TryLock`) | Unix and Windows | existing `Lock`/`Unlock` byte-identical; the `@MX:ANCHOR` contract text and REQ-ATR-001 ("behaviour preservation") need a leader/operator ruling on whether an additive symbol needs an amendment | waiter returns at once instead of waiting out the prune | **yes** (changes a shared package's exported surface) |
| C. Bounded wait: poll the try-lock with a sleep up to a cap. | B plus a cap constant | 1 | both | as B | wait capped at the cap, which adds up to the cap to the hook's life | **yes** |

Default: **A**. Why: the wait delays only the hook's exit, not its event (READ: `observer.go:86` precedes `:93`), the hooks are `async: true`, and the premise has no measurement; B and C change the exported surface of a package with seven importers.

### D3 — Event loss between the pruner's read and its rename  (premise: OBSERVED)

| Option | Files touched | New exported API | Hot append path | Windows | operator-held |
|---|---|---|---|---|---|
| A. Appenders take the state-file lock around every append. | `observer.go` (three append sites: `:80`, `:129`, `appendEventsJSONL` `:270`), `retention.go`, tests (≥ 4) | 0-1 | **changed**: each append adds an open, a lock, an unlock and a close of the state file (four system calls, counted not timed), and an appender waits out the whole prune (0.53-1.79 s reported by the t1425 lane for 65.8 MB, unbounded for a larger log). With the 5 s hook timeout, an append queued behind a prune slower than 5 s is killed before it writes: the event is lost, which is worse than today. | lock is in-process only: no cross-process effect | **yes** (hot path of every observe hook; doctrine-level) |
| **B. Tail-carry (default; stands per the relayed verdict)** — the pruner classifies only the whole lines inside the prefix of the log it measured, and immediately before the rename copies every byte the log gained after that prefix into the replacement. | 1 production (`retention.go`) plus tests | 0 | **unchanged** (zero added latency, no lock, no added call) | works (no locks involved); rename-over-open-file semantics are pre-existing | no |
| C. Documented residual risk only. | 1 (`retention.go` comment) | 0 | unchanged | unchanged | no |

Default: **B**. Why: it is the only option that closes the OBSERVED `late-event` loss without touching the hot append path or any lock contract. It narrows the loss window from "the whole prune" (read, archive, rewrite; the archive step alone was blockable by a FIFO in the probe) to the gap between the final tail reading and the rename; it does **not** reach zero, and REQ-HRH-009 requires the residual window to be stated in source. C closes nothing; A closes the window fully on Unix but adds latency and a new kill-before-append failure mode on every hook.

Boundary and terminator rule (REQ-HRH-006, REQ-HRH-007):

- The measured prefix is the log as it stands when the pruner opens it. Only whole, newline-terminated lines inside the prefix are classified as kept or stale. The bytes after the last newline inside the prefix (a final line with no terminator) and every byte appended afterwards form the tail, which is copied verbatim and unclassified.
- The replacement log is the kept lines (each terminated by one newline, as today) followed by the tail. When the tail is non-empty and does not end with a newline, the pruner adds exactly one newline. The replacement is therefore empty or newline-terminated, the property OBSERVED at base, so the next append starts its own line and is never glued onto a carried fragment.
- A partial last line (an append observed mid-write, or a damaged final line) is carried whole, exactly once, with that terminator. If its writer completes later against the old file, the remainder is lost (part of the residual window) and the fragment stays in the log as an unparsed line that later prunes keep verbatim (t1433 behaviour). Whether a concurrent append can be observed half-complete on the target file systems was not observed.
- Consequence, stated rather than hidden: a final line that has no terminating newline is not classified in this prune. It is classified by the first later prune that finds it terminated: through the terminator the pruner adds when it rewrites the log (which it does when the log holds any other stale event), or through an append that completes the line. When that final line is the only stale event, the prune returns without a rewrite (nothing was classified as stale) and adds no terminator, so a damaged unterminated tail can stay unclassified; this is stated, not pinned by a test. Malformed lines inside the prefix keep today's rule: kept verbatim.
- The tail is read once, immediately before the rename. Events appended after that reading, and events written by a process that opened the old file before the rename, are the residual window.

### D4 — F4 and F7, the state-file path  (premises: OBSERVED)

#### D4.a — What the pruner may delete

| Option | Files touched | New exported API | Platforms | Behaviour change | operator-held |
|---|---|---|---|---|---|
| **A. Replace-on-fault, restricted to entries the current user owns (default; the restriction is the relayed Q4 verdict)** — inspect the state path without following links. A symbolic link whose own owner is the current user (the owner recorded on the link, read without following it and never the target's: OBSERVED, §A) or a regular file owned by the current user that cannot be opened read-write because of a permission error is removed and recreated as a regular file. An entry owned by anyone else, or whose owner cannot be determined, is left byte-identical, the prune is skipped, an error is returned and a warning is emitted (D4.b). A directory keeps today's behaviour: skip, return the error (tested, AC-HRH-003). A FIFO or another special file at the state path is not handled by this SPEC: a FIFO hangs the lock-free pre-check before any inspection (OBSERVED, §A, §F), and this SPEC neither repairs nor changes that. One mechanism closes F4 and F7. | `retention.go` plus two build-tagged owner-check files (a `//go:build !windows` file and a `//go:build windows` twin) plus tests | 0 | Unix: owner read from the entry's platform stat record and compared with the effective user id. Windows: per-user file ownership is not read, so the owner cannot be determined and every entry counts as not owned (conservative: never deleted, warn only; a symbolic link there is still never followed). | a hostile or foreign-owned state file no longer disables or abuses retention silently; on Windows a faulty entry stays in place and retention stays off with a warning | no |
| B. Refuse-only for F4 (the platform no-follow open flag, O_NOFOLLOW, which is undefined on Windows, checked with `GOOS=windows go doc`, so any use needs a `//go:build !windows` file with a Windows twin) and visibility-only for F7 (a typed error the observer logs). | ≥ 3 (two build-tagged files for the flag, `observer.go` for logging) plus tests | 0-1 | as stated | retention stays off while the fault persists (F7 not closed, only audible); touches `observer.go` | no, but it touches the hot path file |

Default: **A**. Cross-platform clause: the owner check is the only POSIX-only production code in this SPEC; it lives in the `//go:build !windows` file with a `//go:build windows` twin that defines the same symbol, so `GOOS=windows go build` and `go vet` compile both packages (REQ-HRH-012, AC-HRH-011).

#### D4.b — Where the refusal warning goes

| Option | Cost | Visible to an operator? |
|---|---|---|
| **A. One line on standard error from the pruner (default)**, `[WARN] harness/retention: ...` naming the state path, following the precedent at `internal/harness/safety/frozen_guard.go:124` (`fmt.Fprintf(os.Stderr, "[WARN] ...")`). | no new field, file or exported symbol; the test captures standard error by swapping `os.Stderr` in a serial (non-parallel) test, which cannot overlap the package's parallel tests; the warning repeats once per prune attempt while the condition persists, which is every hook event (no stamp can be recorded), so the volume equals the hook-event rate; whether a hook's standard error is shown to a user was not measured | yes where hook stderr is shown; not measured |
| B. A typed error only: the returned error wraps an unexported sentinel and no line is written. | zero noise; the test asserts the sentinel with `errors.Is` | no: the observer discards the error (`observer.go:93`, `:141`, unchanged), so F7's complaint (silent disablement) would stay silent in production |

Default: **A**. B satisfies the refusal but not the warning the relayed verdict asks for.

#### D4.c — Safety when several hook processes meet the same fault

`os.Remove` removes whatever is at the path, not a specific inode. If process A has healed the entry and holds the lock on its new state file, a slower process B that inspected the faulty entry earlier and then removes the path would delete A's fresh file, create its own, and lock a different inode: two pruners, the t1425 hazard.

| Option | Cost | Effect on the single-pruner guarantee |
|---|---|---|
| **A. Conditional removal (default)** — the pruner removes an entry only after re-inspecting it and finding it still the entry it inspected first (same file identity, same type, same permission mode, same modification time; identity alone is not relied on, as a defence in depth against a file system that reuses a freed inode number; reuse was not observed in this run). Otherwise it re-inspects without removing (at most three inspections, then skip with an error). The replacement is created exclusively (never follows a link, fails if the path exists), and an "already exists" result is treated as "another process healed first". A regular file opened for the prune gets a post-open identity check against the inspection before any stamp write. | about 20 lines in `retention.go`; a deterministic helper-level test for the losing interleaving | steady state unchanged; during a heal burst the guarantee is narrowed to the gap between the last re-inspection and the removal, which is not closed and **not measured**. The burst size is the number of hooks that observed the fault in the same instant, not two. |
| B. Leave the removal unconditional and disclose the hazard as unmeasured. | 0 lines | the guarantee is dropped for the burst; the delete-the-winner interleaving is possible and would re-open concurrent appends to one archive file |
| C. Serialize healers through a separate lock file. | a new file and lock lifecycle | closes the burst, but adds an artifact and a failure mode this card did not ask for |

Default: **A**, the smallest option that keeps the guarantee in every case except a gap of a few instructions.

### D5 — Test seams: N2 stamp-write pin (mutant mC) and the owner lookup

The operator allowed one unexported test-only field on the pruner (relayed verdict). Two seams are wanted; the allowance is spent once.

| Seam | Option | Cost | Note |
|---|---|---|---|
| Stamp-write failure (mC) | **A. Parameter seam (default)** — extract the locked phase of `pruneExclusive` (stamp re-check, stamp write, prune, sweep) into a function that takes the already-opened state file; the test passes a file opened read-only, whose truncate fails (OBSERVED darwin; Linux and Windows not observed). | about 25 lines moved, no struct field; existing tests cover the move | test file carries `//go:build !windows`; Windows behaviour is not claimed |
| Stamp-write failure (mC) | B. The allowed field holds a replaceable stamp writer. | 1 struct field, 1 default line, 1 call site | spends the allowance, leaving the owner lookup without a field |
| Stamp-write failure (mC) | C. Accept mC as a recorded survivor. | 0 | N2 stays half-closed |
| Owner lookup (Q4 pin) | **A. The allowed field holds the owner-check function (default)**, set by the constructor to the real check; tests replace it to simulate a foreign-owned entry. | 1 struct field, 1 constructor line | a non-root user cannot create a foreign-owned file (OBSERVED), so no portable test reaches the refusal branch without a seam; the real check is pinned separately by AC-HRH-005 |
| Owner lookup (Q4 pin) | B. No field: a root-only end-to-end test using `chown`, skipped for non-root users. | 0 fields | the refusal is pinned only where CI runs as root (not measured); the wiring from the pruner to the real check stays unpinned elsewhere |

Default: mC option A and owner-lookup option A. mB needs no seam: it is killed by file content alone.

## §C Requirements (GEARS)

- **REQ-HRH-001** (Event-driven, item F4): When the prune state path (`<log>.prune-state`) is a symbolic link whose own owner is the current user — the owner recorded on the link itself, read without following it, never the owner of the entry it points at — the retention pruner shall leave the link's target unmodified, shall replace the link with a regular state file it creates at that path, and shall continue the prune; the guarantee holds for a link whose target does not block a read-only open (a link to a FIFO does, §F).
- **REQ-HRH-002** (Event-driven, item F7): When the prune state path is a regular file owned by the current user that the retention pruner cannot open for reading and writing because of a permission error, and the directory permits the current user to remove it, the pruner shall replace it with a regular file it can lock and shall continue the prune.
- **REQ-HRH-003** (Event-driven, item F7 boundary): When the prune state path cannot be used and cannot be replaced under REQ-HRH-001 or REQ-HRH-002 — because the entry is not owned by the current user (for a symbolic link: because the link's own owner is not the current user, whatever its target's owner), because its owner cannot be determined, because it is a directory, because a regular file cannot be opened for a reason other than a permission error, or because the directory does not permit its removal — the retention pruner shall skip the prune, return an error naming the failure, and leave the usage log, the archive directory and the state-path entry (a symbolic link's target included) byte-identical to their state before the call; a FIFO or another special file at the state path is outside this requirement (§F).
- **REQ-HRH-004** (Event-driven, item F7 / Q4 warning): When the retention pruner declines to replace a symbolic link or a regular file at the prune state path because the entry is not owned by the current user (for a symbolic link: because the link's own owner is not the current user) or its owner cannot be determined, the pruner shall write one warning line naming that path to standard error; where the platform does not expose per-user file ownership, the pruner shall treat every entry's owner as undeterminable.
- **REQ-HRH-005** (Event-driven, item F4 / F7 concurrency): When the retention pruner replaces a prune state path entry, it shall remove that entry only if it is still the entry the pruner inspected, and otherwise shall re-inspect without removing it, so that a state file created by a concurrent healer is not removed by this pruner.
- **REQ-HRH-006** (Event-driven, item event loss): When event lines are appended to the usage log after the retention pruner measured it and before the pruner takes its final reading ahead of the replacement, the pruner shall carry those bytes, verbatim and in their original order, after the kept lines of the replacement log.
- **REQ-HRH-007** (Ubiquitous, item event loss): The replacement usage log shall be empty or end with a newline, and a final line that lacks a terminating newline when the retention pruner measures the log shall be carried whole, exactly once, unclassified and unmodified apart from the added terminating newline.
- **REQ-HRH-008** (Ubiquitous, item event loss): The observer's event-append path and the retention pruner's pre-lock path (the in-memory interval check, the stamp read and the log existence check; at base `PruneStaleEntries`, `readStamp`, `readStampFile` and `stampIsFresh`) shall add no lock, no file open and no system call beyond those present at base `1e2151a38`.
- **REQ-HRH-009** (Ubiquitous, item event loss): The retention pruner's source documentation shall state that a residual window remains between its final tail reading and the replacement rename in which an appended event is lost, and shall not state that events appended during the whole read-to-rename interval are lost.
- **REQ-HRH-010** (Ubiquitous, item F5): The retention pruner's source documentation shall state that on Windows the lock gives no cross-process exclusion, that a burst of hook processes finding no fresh stamp may prune concurrently once per interval, and that this is "F5: not reproduced, not measured".
- **REQ-HRH-011** (Ubiquitous, item F6): The retention pruner's source documentation shall state that lock waiters block with no timeout, that the harness-observe hooks run with a 5 s hook timeout and `async`, that the event is appended before the wait begins, and that the waiter delay is "F6: not reproduced, not measured".
- **REQ-HRH-012** (Ubiquitous, items F4, F5 and F6; cross-platform): The change set shall leave every file under `internal/lockfile` byte-identical to base `1e2151a38`; `internal/harness` and `internal/lockfile`, test files included, shall build and vet for `GOOS=windows`; every new test file that uses a FIFO or another POSIX-only call shall carry a `//go:build !windows` constraint; and every new production file that uses a POSIX-only call shall carry `//go:build !windows` and have a `//go:build windows` twin defining the same symbol.
- **REQ-HRH-013** (Event-driven, item N1): When the pruner under test never reaches its archive step, the killed-pruner test shall fail within 15 seconds with a message naming the missing archive step, rather than waiting for the go test timeout.
- **REQ-HRH-014** (Ubiquitous, item N2 / mB): The retention test suite shall fail when the stamp write does not truncate a longer previous stamp, by asserting on the exact file content of a stamp written over a longer one.
- **REQ-HRH-015** (Ubiquitous, item N2 / mC): The retention test suite shall fail when the pruner proceeds to prune after the stamp write fails, by asserting that a failed stamp write returns an error and leaves the usage log unmodified.

## §D Constraints

- **Baseline-first ordering** (`verification-claim-integrity.md` §2.3): the commit that adds the failing tests (tracked test files that compile against base) is an ancestor of every commit that touches `internal/harness/retention.go` (the commits counted are the card's own, `1e2151a38..HEAD`: the unbounded history of that file holds seven commits older than the base, `acceptance.md` ledger E-021); the observed RED output is recorded in `.moai/reports/t1432/red-baseline.md`, committed in that same commit with `git add -f` (the path is ignored by `.gitignore:235`, and this repository tracks card evidence under it by force-add: `git ls-files .moai/reports/t1425` lists tracked files). The tracked test files, not the ignored report, are the graph witness. Encoded in `plan.md` M0.
- **Windows**: no Windows runtime is available. Every Windows-facing criterion is verified by `GOOS=windows go build` and `go vet` (which compiles test files) only, and by documentation; none claims a runtime observation. Cross-platform exemption clause: the only POSIX-only production code is the owner check, in a `//go:build !windows` file with a `//go:build windows` twin; every new test file using a FIFO carries `//go:build !windows`; a runtime skip is not a substitute because it does not compile on Windows.
- **Hook budget**: harness-observe hooks run with a 5 s timeout and `async: true`. One prune of a 65.8 MB log took 0.53 s (nothing stale) and 1.79 s (12.5 percent stale) (t1425 lane, reported in the card dispatch, not re-measured). No requirement here may add work to the append path or to the lock-free pre-check (REQ-HRH-008).
- **No new configuration key, no new exported symbol, no new dependency.** One unexported test-only field is allowed (relayed verdict) and is spent on the owner lookup (§B D5).
- **Run-phase report obligations** (relayed by the leader session on 2026-10-03, which accepted the intermediate red commit, the force-added card evidence and the conservative Windows rule): the completion report states `Windows runtime not observed` in its Gaps section, and names the SHA of the intermediate red commit T in one line (the leader's merge report names it as well).

## §E Exclusions

### Out of Scope — other t1425 findings and history

- F1 (stamp-before-work), F2, F3 (already on this tree: stamp ordering, `sweepOrphanTmp`, and their tests), F8 and F9 (coverage and IDE-hint notes).
- Damaged historical archives (`.moai/reports/t1425/archive-damage.md`): no repair or rewrite of existing `<YYYY-MM>.jsonl.gz` files.
- `partitionEvents` behaviour for malformed lines inside the measured prefix: unchanged.
- `CHANGELOG.md`: owned by the sync phase.

### Out of Scope — shared package and doctrine changes (non-default options only)

- Any change to `internal/lockfile` (D1 option C, D2 options B and C): present in §B as non-default, operator-held options; not authorized by this SPEC. REQ-HRH-012 forbids it.
- Appenders taking a lock (D3 option A): not authorized; REQ-HRH-008 forbids it.
- Moving the prune off the hook path, changing the one-hour interval or the 30-day retention window, or adding configuration for either.
- Any change to `internal/harness/observer.go`.
- Deleting, chmod-ing or otherwise modifying a state-path entry the current user does not own, and any rate limit or on-disk record for the refusal warning (it repeats while the condition persists).
- Windows file-ownership lookup (the conservative "not owned" answer stands) and any Windows handling of junctions or other reparse points beyond what `Lstat` reports.

### Out of Scope — a special file at the state path

- A FIFO or another special file at `<log>.prune-state`: today `PruneStaleEntries` blocks in the lock-free pre-check (`retention.go:94`, `:187-189`; OBSERVED, §A, ledger E-025), so the prune neither skips nor returns. Repairing it needs a change to the pre-lock path, which REQ-HRH-008 forbids, and this card is not about special-file state paths. It is recorded as a pre-existing, unrepaired condition; no requirement or criterion repairs or pins it, and no criterion uses a FIFO at the state path.
- A symbolic link whose target blocks a read-only open (a link to a FIFO): it reaches the same pre-check (READ, not observed) and is outside REQ-HRH-001.

### Out of Scope — verification limits

- Windows runtime behaviour of any kind: build and vet only.
- Reproducing F5 or F6 (a Windows burst; a waiter killed at 5 s): not attempted; revisit only if a reproduction is produced.
- Measuring a heal burst of N hook processes: not attempted (D4.c states the residual as unmeasured).

## §F Risks and unobserved items

- **Partial last line at the tail boundary** (D3 B): an appender's single write may be mid-flight when the pruner takes its readings. Mitigation: the boundary is the last newline inside the measured prefix and the replacement is terminated (REQ-HRH-007); whether a concurrent `O_APPEND` write can be observed half-complete on the target file systems is not observed, so `plan.md` M1 tests the boundary rule deterministically rather than claiming anything about the kernel.
- **Residual loss window remains** (D3 B): an event appended between the final tail reading and the rename, or by a writer that opened the old file before the rename, is still lost. Stated in REQ-HRH-009; not eliminated.
- **Heal burst** (D4.c A): the single-pruner guarantee narrows to the gap between the last re-inspection and the removal. N hooks observing the fault in the same instant can still produce more than one pruner within that gap. Not measured.
- **Time-of-check gap** (D4): between the inspection and the open of a regular file the entry can be swapped for a link; the open then reaches the target without truncating it, and the post-open identity check refuses before any stamp write. This window cannot be forced without a seam and is not pinned by a test (reviewer-read at plan.md M2).
- **Warning volume** (D4.b A): the warning repeats once per prune attempt while a foreign-owned entry persists, which is the hook-event rate; whether hook standard error reaches a user was not measured.
- **F7 persistence**: a foreign-owned or Windows entry, or a directory that does not permit removal, keeps retention off; the warning (REQ-HRH-004) is the only signal, the observer still discards the returned error by design.
- **Pre-existing FIFO hang** (OBSERVED, not repaired): with a FIFO at the state path the hook process does not return from the prune; the harness-observe hook is killed at its 5 s timeout, and the event it was recording is already appended (READ: `observer.go:86` precedes the prune call at `:93`; the timeout is the §A F6 row), so the effect is a lingering hook, not a lost event. A hostile or accidental FIFO there disables retention and delays the hook; this SPEC does not change that (§E).
- **N1 fix choice**: the audit's suggested non-blocking read open (`O_RDONLY|O_NONBLOCK`) is unverified here and can race a pruner that has not yet reached its write-open (the reader sees end-of-file early and the writer then blocks); `plan.md` M4 names the timer-and-unblock form as the alternative. The criterion binds the outcome (fail within 15 s), not the mechanism.
- **Late-event test timing** (AC-HRH-007, AC-HRH-008): the tests wait for the stamp and then a bounded delay before appending, because the pruner's read cannot be observed without a seam; a slow runner can only make a run vacuous (the event is classified normally and survives), never falsely red.
- **Not observed in this plan phase**: the Windows paths; Linux behaviour of any probe above; any measurement of F5 or F6; the mutant kills for mC, for the owner-check mutants (an owner read that follows the link; a check that treats every symbolic link as owned) and for the new state-path and tail tests (they need code that exists only after the run phase); the owner of root-owned symbolic links on Linux (AC-HRH-005 (e) skips where no candidate qualifies).

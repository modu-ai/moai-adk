---
id: SPEC-HARNESS-DETACHED-PRUNE-001
title: "Detached prune off the hook synchronous path — observer records, gate spawns, child prunes"
version: "0.2.1"
status: draft
created: 2026-10-05
updated: 2026-10-05
author: "manager-spec"
priority: P1
phase: "v3.2.0 target"
module: "internal/harness, internal/cli"
lifecycle: spec-anchored
tags: "harness, retention, prune, detached-process, hook-budget, observer"
tier: M
related_specs: [SPEC-V3R3-HARNESS-LEARNING-001]
---

# SPEC-HARNESS-DETACHED-PRUNE-001 — Detached prune off the hook synchronous path

## §A Background and Motivation

The harness-observe hooks prune the usage log synchronously inside the observer's record path. Measured on this tree at HEAD `d05d1d5f0`: `internal/harness/observer.go:93` and `:141` each call `_ = o.retention.PruneStaleEntries(defaultRetentionDays)` after the event append, and `Retention.PruneStaleEntries` (`internal/harness/retention.go:125`) is the full prune — log partition, atomic archive append, rename, orphan sweep. t1467 M1 (atomic archive, commit `a2fae5aa5`) already made a killed prune harmless, but it is still fruitless on a loaded host: the parent verdict measured the partition phase alone exceeding the 5 s hook timeout there, so the log never shrinks and every hourly attempt burns the hook's whole budget before being killed (an idle host completes in 2.7–3.6 s, inside budget — the parent's measurement, cited as the operator's, not re-derived here).

The consequence is structural. The documented observer contract — "runs in under 100ms per event and never blocks the parent tool call" (REQ-HL-001, SPEC-V3R3-HARNESS-LEARNING-001:111) — is violated on loaded hosts by a 5-second prune attempt repeating every hour. The parent verdict (t1467 §4 J2 + §5 M2) approved the repair shape: the observer records the event and does NOT prune; the hook CLI wrapper, after the append, checks the stamp (a cheap lock-free read, once-per-interval gate retained) and, when stale, spawns a detached child that runs the existing prune path outside any hook budget. The child re-checks the stamp under the lock, so a double-spawn collapses into one worker; spawn failure is fail-open (the event is already on disk).

The pieces this SPEC reuses already exist and were re-measured here: the lock-free stamp read `openStampReadOnly` (`internal/harness/retention_open_unix.go:13`, O_NONBLOCK FIFO-aware), the 1-hour interval `pruneSkipDuration = time.Hour` (`retention.go:22`), the lock re-check `pruneLocked` re-reading the stamp after the lock is won (`retention.go:184-190`, the `@MX:NOTE` "Double-checked locking"), and stamp-before-work (`retention.go:195`, so a killed or orphaned pruner is not repeated until the interval ends). What does not exist is anything detached: `grep -rn "SysProcAttr" internal/harness` returns nothing (exit 1) — the detached spawn is genuinely new code, split per platform on the `retention_heal_unix.go`/`retention_heal_windows.go` file-split precedent.

REQ-HL-011 (SPEC-V3R3-HARNESS-LEARNING-001:159) states the prune runs "on every observer write". This SPEC amends that placement — prune moves off the observer write onto the gated detached child — under the parent verdict's approved scope; the interval, the archive destination, and the prune semantics are preserved unchanged (§C C6).

## §B Requirements (GEARS)

Requirement id prefix: `REQ-DP` (Detached Prune).

### §B.1 Work item 1 — the observer records without pruning

- **REQ-DP-001** (Ubiquitous): The Observer shall append each observed event to the usage log and shall not invoke `Retention.PruneStaleEntries` on the record path — both synchronous call sites (`internal/harness/observer.go:93` and `:141`) are removed, the record path performs no log rewrite, no archive append, and no stamp write, and the recording failure semantics are unchanged (a record failure still returns the error the hook handler logs at exit 0). Existing tests that coupled record to prune are updated to prune explicitly through `Retention`.

### §B.2 Work item 2 — the spawn gate

- **REQ-DP-002** (Ubiquitous): The hook CLI wrapper shall, after the event append, read the prune stamp exactly once through the lock-free stamp read (reusing `openStampReadOnly`, `internal/harness/retention_open_unix.go:13`, and its Windows sibling) and, only when the stamp is stale or absent, spawn one detached prune child; when the stamp is fresh the wrapper shall spawn nothing and the added synchronous work shall be that one read plus, on the stale arm, one spawn call. The interval stays `pruneSkipDuration` = 1 hour (`internal/harness/retention.go:22`) — the once-per-interval gate is preserved, re-aimed from "prune attempt" to "spawn".

### §B.3 Work item 3 — the detached child

- **REQ-DP-003** (Ubiquitous): The detached child shall be the hidden `moai hook retention-prune` verb carrying `--log`, `--archive`, and `--days` flags, shall run the existing `Retention.PruneStaleEntries` path with the same constructor and inputs the hook handlers use today, and shall inherit the child-side stamp re-check under the state-file exclusive lock (`pruneLocked`, `retention.go:184-190` — the stamp is read again after the lock is won), so a double-spawn collapses into one worker; the verb is registered hidden so it never appears in help output, and no new prune logic is introduced anywhere on this path.

- **REQ-DP-004** (When the spawn fails): the hook wrapper shall treat the failure as fail-open — the error goes to stderr, the hook exits 0, the event stays on disk, and the observed tool is never blocked — so the documented observer contract of REQ-HL-001 (SPEC-V3R3-HARNESS-LEARNING-001:111, under 100 ms, never blocks) holds on every path including the gate.

- **REQ-DP-005** (Ubiquitous): The parent hook process shall exit without waiting on the child (no `Wait` call — the child outlives the parent and reparents, so the parent leaves no zombie), the child's lifetime is bounded by the prune itself (one `PruneStaleEntries` and exit), and an orphaned child is harmless — the prune is lock-protected and idempotent and the attempt stamp is written before the work (`retention.go:195`), so a killed or orphaned pruner is not repeated until the interval ends.

- **REQ-DP-006** (Where the platform is Windows): the detached spawn shall produce a detached child in its own process group with no console flash — `CREATE_NEW_PROCESS_GROUP` plus exactly one console-suppression mode (`DETACHED_PROCESS` or `CREATE_NO_WINDOW`, alternatives, not a mandatory pair), the concrete combination finalized at run start per C4 against the pinned Go version's exported constants, with `HideWindow` where available — delivered as build-tag-split files following the `retention_heal_unix.go`/`retention_heal_windows.go` precedent; `GOOS=windows GOARCH=amd64 go build ./...` and `go vet` parity is an acceptance criterion, while the Windows runtime behavior of the detached child — including the documented in-process-only state-file mutex (`retention.go:110-112`, the F5 limitation) — stays unobserved and documented as such, not claimed.

### §B.4 Work item 4 — test discipline and semantics preservation

- **REQ-DP-007** (Ubiquitous): Tests shall never spawn a real detached child — the spawn action is an injectable seam on the gate (a function field replaced by a recording fake in tests, in the spirit of the `ownerCheck` field at `retention.go:78-81` and the FIFO seam of `retention_fifo_unix_test.go`), and no test creates background load (house rule, `AGENTS.md` §4).

- **REQ-DP-008** (Ubiquitous): The system shall run the same `Retention.PruneStaleEntries` with its 1-hour interval, stamp-before-work ordering, atomic archive (t1467 M1), and orphan sweep — moved off the synchronous path and unchanged in behavior — and shall change no prune flag, no configuration key, and no archive format.

### §B.5 Work item 5 — the classifier applies the retention window itself

- **REQ-DP-009** (When the classifier aggregates usage-log events for tier promotion): the classification input shall exclude retention-expired events before aggregation — the log read feeding `AggregatePatterns` shall skip events older than the retention days (`defaultRetentionDays`), because `Pattern` carries no per-event timestamps (`internal/harness/types.go:304`) and a post-aggregate filter cannot distinguish vintages (gate-measured: 4 expired + 1 recent aggregates identically to 5 recent) — so no promotion, no tier increment, and no Tier 4 proposal draws on a retention-expired event; the classifier shall never rely on the detached prune child having run, because the child is asynchronous and its result is not visible at classification time, and the exclusion mechanism (a cutoff parameter on the aggregation read) is finalized at run start.

## §C Constraints

- **C1 — no interactive question.** The gate and the child ask nothing of anyone; CLI code never prompts (C-HRA-008, `internal/cli/CLAUDE.md`).
- **C2 — no configuration surface.** The gate has no yaml key: the interval is `pruneSkipDuration`, already code-owned; adding a config key would be a second knob for one number.
- **C3 — Template-First exemption.** Every touched file is Go source under `internal/`; no `.claude/` or `.moai/` template-managed file changes, so no template mirror lands with this SPEC.
- **C4 — cross-platform build tags are mandatory.** The new code uses `syscall.SysProcAttr`, so the platform split is `//go:build !windows` / `//go:build windows` files per the file-split precedent; `GOOS=windows GOARCH=amd64 go build ./...` gates every milestone exit.
- **C5 — the observer contract is restored, not changed.** REQ-HL-001's under-100-ms never-blocks bound (SPEC-V3R3-HARNESS-LEARNING-001:111) is the invariant this SPEC repairs; no requirement in this SPEC relaxes it.
- **C6 — REQ-HL-011 relationship, declared.** REQ-HL-011 ("prune ... on every observer write") is amended by this SPEC's placement change per the parent verdict's approved M2 shape; the owning SPEC (grandfathered era) is not edited. The 1-hour interval, the `learning-history/archive/<YYYY-MM>.jsonl.gz` destination, and all prune semantics are preserved (REQ-DP-008).

## §D Decisions

- **D1 — child entry shape: a hidden verb of the same binary.** `moai hook retention-prune --log <path> --archive <dir> --days <n>`, registered `Hidden: true`, invoked via `os.Executable()` self-exec (precedent: `internal/cli/factory_relaunch.go:166`). Rationale: the spawn target needs a stable argv and a testable entry point; a hidden cobra verb follows the house pattern of hook-only verbs (`harness-observe` and siblings, `internal/cli/hook.go:140-165`) and is greppable, where an env-var-gated internal mode re-enters command dispatch invisibly and cannot be asserted from a test without the env plumbing.
- **D2 — gate placement: one wrapper, one gate primitive.** The gate logic (stamp read + spawn decision) lives in `internal/harness` as one function so it is testable without the CLI layer; the wiring is ONE unexported wrapper in `internal/cli/hook.go` (record-then-gate) that the four observe handlers (`:843`, `:933`, `:1071`, `:1250`) each call — no gate logic is duplicated at any call site.
- **D3 — the child is the same binary.** No separate helper binary and no `go run` at runtime: the child is the installed `moai` binary itself, so version skew between spawner and pruner cannot arise.
- **D4 — the parent never waits.** The POSIX arm detaches with `setsid`; the parent proceeds to its existing cobra exit. There is no reaper, no wait goroutine, and no exit hook — the child's lifetime bound is the prune itself (REQ-DP-005).
- **D5 — the double-check already exists.** REQ-DP-003's lock re-check is `pruneLocked`'s existing `@MX:NOTE` "Double-checked locking" behavior (`retention.go:187`); the requirement states preserved behavior the child inherits by entering through `PruneStaleEntries`, not new code — the child verb must not call `prune` internals directly, and that constraint is the testable half.
- **D6 — id prefix.** `REQ-DP` cannot collide with the owning SPEC's `REQ-HL-*` ids in any grep, log, or audit.

## §E Exclusions

### Out of Scope — what this SPEC does not do

- Any change to the prune interval, a configuration key for it, or interval tuning — `pruneSkipDuration` = 1 hour is preserved verbatim (REQ-DP-008).
- Cross-process exclusion on Windows: the documented F5 limitation (in-process-only mutex, `retention.go:110-112`) is inherited, not fixed; a Windows burst may spawn N children once per interval.
- A cron, daemon, or timer-driven periodic pruner — the child is event-triggered only (spawned by the gate after an observed event on a stale stamp).
- Log compaction or rotation beyond the existing archive behavior (the `<YYYY-MM>.jsonl.gz` archive and the 30-day `defaultRetentionDays` input are untouched).
- Changes to `Retention.PruneStaleEntries`'s signature or API surface — the observer drops the call; the method itself is unchanged (its `@MX:ANCHOR` caller comment at `retention.go:62` is updated when the callers change).
- The documented tail-loss residual window (the `@MX:WARN` at `retention.go:119-124`): an event appended in the residual window is lost today and after this SPEC; fixing it is a separate concern.

## §F Risks

- **F1 — spawn rate on a loaded host.** The gate adds one fork/exec per stale-stamp event — once per interval when the child's stamp write succeeds. The work is once per interval, not the spawn: a stamp-write failure (`retention.go:195-196` skips the prune and leaves the stamp absent) or a slow first child leaves the gate's view stale, so further events within the same interval spawn repeated children, each collapsing at the child-side lock re-check into one worker's work (the Windows burst of §E exclusion 2 is the same shape cross-process). A loaded host where the fork itself is slow still pays only the fork (fail-open bounds the wrapper's synchronous cost; the prune runs outside the budget).
- **F2 — a still-running child holds the lock into the next interval.** The next hour's gate reads the stamp as fresh (written before the work), so no second child spawns while one runs; only after the interval expires can a waiter queue on the lock — the documented unbounded-waiter behavior (`retention.go:155-158`) is inherited, bounded by the prune duration, and now paid by the child instead of the hook.
- **F3 — Windows runtime behavior is unobserved.** The detached child is specified and build-verified on Windows, not runtime-verified (no Windows host in this lane); REQ-DP-006 carries that honestly and GOOS=windows build+vet parity is the enforced half.
- **F4 — hidden-verb discoverability.** The verb is deliberately hidden from help; this SPEC and the source registration are its documentation. A user invoking it manually is harmless (it runs the same lock-protected prune path).
- **F5 — observer test drift.** Removing the sync calls may surface tests that silently relied on record-triggered pruning; the M1 re-verification set runs the whole harness package to catch them, and the update direction is always "prune explicitly through `Retention`" (REQ-DP-001), never "restore the coupling".

## §G History

- 2026-10-05, v0.1.0 (card t1497): initial plan-phase authoring on branch `WT-harness-prune-detached`, HEAD `d05d1d5f0`. Scope is the parent verdict's approved M2 shape (t1467 §4 J2 + §5 M2), inlined into the delegation because `.moai/reports` is disk-only and absent from this tree. Every code anchor and every RED-now cell in acceptance.md §B was measured on this tree in this session; the child-entry (D1) and gate-placement (D2) decisions named by the delegation as open were resolved and recorded here.
- 2026-10-05, v0.1.1 (card t1497, plan-audit iter1 repair — verdict FAIL 0.88 at artifact commit `080578aec`, D1 blocking + D2/D3/D5 advisories absorbed, D4 dispositioned): AC-DP-001's GREEN commands moved out of the acceptance §C table into LEDGER-DP-GREEN-A with the raw-pipe alternation — the escaped form is a literal pipe in Go regexp and selected zero tests while exiting 0, a vacuous green on a release-blocking AC (D1; the auditor's positive control re-executed by this lane, both forms recorded in-ledger). LEDGER-DP-FORM added for the grep-BRE alternation form-control — the same-class sweep found the `\|` sequence in three more places: plan.md's M3 boundary grep and LEDGER-DP-C (correct grep BRE grammar, now control-proven live) and plan file 5's quoted Go bitwise-OR in a table cell (removed by the D3 rewrite). F1's spawn-rate quantifier corrected to the interval-conditional with the stamp-write-failure/slow-child repeat-spawn residual named (D2). REQ-DP-006 and plan file 5 aligned onto one canonical Windows flag contract — semantic contract plus named alternatives, finalized at run start per C4 (D3). The detachment property's construction-only verification basis declared in acceptance §D and plan §D; deliberately no AC — no RED-now-able input exists on this tree (D4 disposition).
- 2026-10-05, v0.2.0 (card t1497, consolidated turn-end gate round — three findings, one commit): REQ-DP-009 + AC-DP-008 added (finding 1 — with the synchronous prune gone, `classifyHarnessPatterns` aggregates retention-expired events into the learning curator; the gate's overlay measurement, cited as the finding source: 5 events aged 60 days on the removed path → 1 spurious promotion + 1 proposal versus 0 on the current path; the classifier now applies the retention window itself, joining the existing filter-at-classification-time pass, with the 5×60-day regression test). File 8's named fake enumeration replaced by a sweep obligation (finding 2 — the gate overlay-verified `hook_harness_observe_stop_test.go`, `TestRunHarnessObserveStop_RecordsWhenEnabled` et al., also driving handlers over unstamped logs; every handler-driving CLI test gets the `retentionSpawnImpl` override and the whole affected package re-verifies to spawn nothing real). Divergence recorded, deliberately not acted on (finding 3 — NOTE ONLY): the factory kickoff→run consumer reads `audit_ready: true` in progress.md §E.1 while the schema prose convention is `plan_status: audit-ready`, and a gate test feeding the current §E.1 to the real consumer was rejected; the lane writes the recognized flag only after the delta-2 PASS on this consolidated state — this SPEC does NOT rewrite §E.1's prose convention unilaterally, and the next auditor should treat the two spellings as the same signal pending a schema-consumer reconciliation.
- 2026-10-05, v0.2.1 (card t1497, gate round 4 — two refinements): REQ-DP-009's filter placement narrowed to the AGGREGATION INPUT — the post-aggregate alternative was unimplementable because `Pattern` (`internal/harness/types.go:304`, re-measured this session) carries no per-event timestamps, and the gate measured 4-expired+1-recent aggregating identically to 5-recent; the regression test is mixed-vintage (4 expired + 1 recent → the recent-only pattern classifies `observation`, never `rule` — an all-expired fixture would pass even an over-aggressive filter that dropped everything). `TestDetachedChildPrunes` and `TestDetachedChildDoubleSpawnCollapses` moved from the harness to the cli test package — both drive the CLI verb's run function, and a harness test importing `cli` fails with the import-cycle-not-allowed-in-test error (gate overlay-reproduced); LEDGER-DP-D extended with the cli-side zero-match re-measure. (The v0.2.0 row above is preserved verbatim; this round narrows v0.2.0's REQ-DP-009 mechanism wording and the regression-fixture design — both v0.2.0 rows' findings remain in force.)

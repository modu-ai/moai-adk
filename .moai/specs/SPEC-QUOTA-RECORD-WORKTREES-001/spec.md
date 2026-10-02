---
id: SPEC-QUOTA-RECORD-WORKTREES-001
title: "Quota aggregator reads the record directories of linked worktrees — the usage gate sees readings from sessions that run inside card worktrees"
version: "0.1.0"
status: draft
created: 2026-10-02
updated: 2026-10-02
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/statusline, internal/cli"
lifecycle: spec-anchored
tags: "quota, rate-limit, worktree, session-telemetry, record-directory, factory, usage-gate, card-t1442"
tier: M
card: t1442
depends_on: [SPEC-QUOTA-AWARE-SCHEDULING-001]
related_specs: [SPEC-STATE-ANCHOR-001, SPEC-FACTORY-SELF-DISPATCH-001]
---

# SPEC-QUOTA-RECORD-WORKTREES-001 — Quota aggregator reads linked-worktree record directories (card t1442)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-10-02 | manager-spec | Initial plan-phase draft for card t1442 on `WT-quota-read-worktree-records`, base develop `284e09c44`. Origin: sync-audit finding F1 (MAJOR) of SPEC-QUOTA-AWARE-SCHEDULING-001 (`.moai/reports/t1347/sync-audit.md`, carried there as debt D1 in `progress.md` §E.4). Open decisions are marked `[DECISION-OPEN]`, each with its options and a default, and routed in `decision-index.md`. The predecessor SPEC is not edited (a completed SPEC body is immutable); this SPEC supersedes its D2 premise "any live Claude session's reading describes the account" only in the sense that the reader now actually reaches those sessions' records. |

**Tier M.** Estimated 300-600 lines across 5-8 files (`internal/statusline/quota.go` refactor plus one new file `quota_dirs.go`, one seam value in `internal/cli/factory_quota.go`, three new test files, the run-phase progress record, and the sync-phase CHANGELOG sentence) in two packages: inside the Tier M bands (300-1000 lines, 5-15 files) and above the Tier S band (under 5 files). 10 REQ / 12 AC, inside the Tier M ceilings of 16 / 16 and above the Tier S ceilings of 8 / 8. Tier M artifacts: spec.md, plan.md, acceptance.md, plus progress.md and decision-index.md (the decision gate is on in this repository). plan-auditor threshold 0.80.

## §A Background

The predecessor SPEC (SPEC-QUOTA-AWARE-SCHEDULING-001, completed) added a usage gate: a Claude lane near its account quota leases no new card. The gate's input is the quota windows each session's statusline writes into that session's telemetry record. The sync-audit of the predecessor found the gate reads one directory while records are scattered over many (finding F1). Facts measured in this tree (HEAD `284e09c44023598affe486f17701717ca173e6ca`; every line cite comes from a command run in this plan phase):

- **Reader side, one directory.** `internal/statusline/quota.go:69-103` `AggregateQuota(stateDir, now, maxAge)` lists exactly `<stateDir>/context-usage` (`:74-75`), skips a file whose mtime is older than `maxAge` (`:86`) and a record whose capture time is stale or later than `now` by more than `config.QuotaClockSkewTolerance` (`:94`), and keeps per window the freshest record by capture time (`winner.offer`, `:114-121`; a tie keeps the earlier offer). Its header states it is a pure reader (no network, no process spawn, no write).
- **The single evaluation.** `internal/cli/factory_quota.go:87-99` `factoryQuotaEvaluate(root)` calls the seam `factoryQuotaAggregate(filepath.Join(root, ".moai", "state"), factoryCardNow(), gate.MaxAge)` (`:93`; the seam variable is `:35`, `var factoryQuotaAggregate = statusline.AggregateQuota`) only when the gate is enabled (`:90-92`). It takes no caller input. Its non-test callers: `factory_quota.go:136` (latch), `:194` (acquire warning), `:248` (status block), `factory_quota_lanes.go:152` (steering). Tests replace the seam in `factory_quota_test.go:248-269`; its type is `func(string, time.Time, time.Duration) statusline.QuotaAggregate`.
- **Root is the primary checkout.** `factory_card.go:32` `factoryCardRoot()` is `resolveTodoQueueRoot()` (`todo.go:74-76`, the primary checkout of the repository the launch directory sits in). `moai factory next` additionally refuses to run anywhere but the parent checkout (`factory_card.go:645` `factoryAssertParentCheckout(resolveProjectDir())`, `:149-161`), then evaluates the gate once per pass (`:661-664`) and re-checks every 5 s under `--wait` (`:166`, `:687`).
- **Writer side.** `internal/stateanchor/stateanchor.go:71-83` `Resolve`: stdin `workspace.project_dir` when it is an existing absolute directory, then `worktree.original_cwd`, then the git common directory's parent. A session working inside a card worktree therefore writes `<worktree>/.moai/state/context-usage/<session-id>.json`; the same session id writes into the primary directory while the session is anchored at the primary.
- **A session's record follows the checkout the session is anchored to at that moment.** Measured 2026-10-02 18:32 +0900 for session `2da35a68-1196-4183-b6e5-a50fc9b6d901` (this card's lane): copies in four directories, `.claude/worktrees/develop` 18:26, primary 18:28, `.moai/worktrees/t1347` 18:21, `.moai/worktrees/t1442` 18:30; the freshest copy sits where the session is anchored now, and the others stay frozen at their last in-directory write. Two more session ids show the same shape (`63a46e3e-…` in four directories, `21477fb9-…` in four).
- **Scale.** Record files on disk now: 1043 in the primary directory, 624 in the record directories of linked worktrees (`find … -path '*/.moai/state/context-usage/*.json'` over `.moai/worktrees` and `.claude/worktrees`, depth 6). Written in the last hour: 4 files in the primary directory, 19 files in 17 worktree directories. The predecessor audit measured 6 versus 14 an hour earlier. Most readings are therefore invisible to the gate today; an invisible reading reads as unknown and fails open (no hold), so the defect is a silent blind spot, not a wrong hold.
- **Linked worktrees can be enumerated by file reads alone.** `<primary>/.git` is a directory; `<primary>/.git/worktrees/` holds 27 entries now, each with a `gitdir` file whose single line is the absolute path of that worktree's `.git` file (`cat .git/worktrees/*/gitdir`: 21 under `.moai/worktrees/<name>/.git`, 6 under `.claude/worktrees/<name>/.git`, none outside the repository root today, but `moai cc -w <abs-path>` may place a tree anywhere). The metadata entry name need not equal the directory name (entry `t8101` points at `.moai/worktrees/t810/.git`). One entry carries a `locked` marker (`agent-aedd038f6d2e1f5de`, an agent tree). `git worktree list --porcelain` shows no `prunable` line today; a removed worktree leaves its metadata entry until pruned, so a `gitdir` target can be missing.
- **The sweep that bounds the design.** The predecessor's AC-QAS-014 sweeps `internal/statusline/quota*.go` (floor 1, `quota_test.go:304-315`) and `internal/cli/factory_quota*.go` (cli half, `factory_quota_test.go:730-731`, floor 3) and requires no `net`, `net/http`, or `os/exec` import (REQ-QAS-007). Any new file named `quota*.go` joins the sweep automatically.
- **Predecessor evidence.** `sync-audit.md` F1/F2, `progress.md` §E.4 (D1) and §J of SPEC-QUOTA-AWARE-SCHEDULING-001; the card text names the repair scope: read the linked-worktree record directories (or the session's own project directory), one measurement from a real lane around `moai factory next`, and a scoped re-audit.

## §B Decisions

Each decision states the options considered, the chosen default (kept minimal), and why. `[DECISION-OPEN]` marks a decision the plan does not close; it carries a default so the run phase can proceed, and is routed in `decision-index.md`.

### D1 — Which directories the aggregator reads  [DECISION-OPEN]

- Options: (A) the primary directory plus every linked worktree enumerated from `<git-common-dir>/worktrees/*/gitdir` (file reads only); (B) the primary directory plus the session's own project directory only (`resolveProjectDir()`); (C) the primary directory plus a glob of the two known layouts under the primary root (`.moai/worktrees/*`, `.claude/worktrees/*`).
- Default: (A).
- Why: the account quota is shared, so any fresh record from any directory is a valid reading, and the scattering is across ~17 live directories. (B) sees two of them; on the `moai factory next` path it adds nothing, because that verb runs only from the parent checkout, so `resolveProjectDir()` already is the primary (`factory_card.go:645`); on `integration acquire` it adds the caller's own worktree but not the other lanes'. (C) is equal to (A) today (all 27 targets sit under the two layouts) but silently misses a tree placed elsewhere by `moai cc -w <abs-path>` and carries two hard-coded layouts that the git metadata already states. (A) costs one directory read plus up to 128 small file reads and as many record-directory listings per evaluation; the cost is unmeasured and bounded (D4).
- Not covered by any option: records written under a directory that is neither the primary nor a registered worktree (a stray anchor such as `.claude/worktrees/.moai/state/context-usage`, present on disk with stale files): that is the writer's domain (out of scope below).

### D2 — Where the enumeration lives

- Options: (A) change `AggregateQuota` to take several directories; (B) a new multi-directory entry point plus a file-read enumerator in a new file `internal/statusline/quota_dirs.go`, the single-directory function kept as is; (C) enumerate in `factoryQuotaEvaluate` and merge per-directory aggregates.
- Default: (B), with the one-directory function a one-element call into the same scan loop.
- Why: the old tests and the AC-QAS-014 sweep call `AggregateQuota(stateDir, now, maxAge)`; the cli tests replace the seam `factoryQuotaAggregate` with a function of that exact type. Keeping the seam type and the single-directory function unchanged lets every `TestQAS_` test stay green without an edit (REQ-QWR-010). The new seam value is a thin function of the unchanged type that derives the repository root from the `<root>/.moai/state` path it receives (a path whose last two elements are not `.moai/state` reads that one directory only) and calls the multi-directory reading; `factoryQuotaEvaluate` and its MX anchor stay as they are. A new file `quota_dirs.go` is named `quota*.go`, so the existing sweep covers it with no test edit. Option (C) was rejected: per-directory winners merged afterwards equal one pass in result but split the freshest-wins rule across two layers and multiply seam calls.
- Alternative kept visible: changing the seam signature to carry directories is the cleaner type but edits two helper functions in the predecessor's test file (`factory_quota_test.go:246-270`); not chosen.

### D3 — Freshness, skew, dedupe across directories

- Chosen default: unchanged rules. A record is read through the same mtime pre-filter, capture-time age check, and skew tolerance; per window the freshest record by capture time across all directories wins (never the maximum, never the minimum); a tie keeps the directory listed first (the primary, then worktrees in sorted entry-name order). The same session id in several directories needs no dedupe: its stale copies lose on capture time, which is exactly the shape measured above. No new configuration key, so the config cache schema version is not bumped.
- Why: the predecessor's REQ-QAS-005 rules are the audited ones, and the failure measured here is reach, not freshness.

### D4 — Failure modes and a bound  [DECISION-OPEN]

- Every failure of one directory means that directory contributes nothing (fail open, the predecessor's REQ-QAS-006): an unreadable or empty or oversized or relative `gitdir`, a target that no longer exists (pruned), a record directory that is absent, unreadable, a file, or a symbolic-link loop, a `locked` marker (a locked worktree is still read; the marker means "do not prune"). No directory is walked recursively, so a link loop cannot recurse; an unreadable listing returns an error that is skipped.
- Bound: at most 128 linked-worktree entries per call, read in sorted entry-name order, and at most 4 KiB of each `gitdir` file. 128 is about 4.7 times the 27 entries measured now and is **unmeasured**; the cost model is ~1667 `lstat` calls per evaluation today versus 1043 before (primary 1043 plus worktrees 624 files), also unmeasured. Options: 64 / 128 / unbounded; order by name / by newest metadata-directory modification time. Default 128, name order: the bound only protects against a pathological count, and name order is deterministic and testable; a lane beyond the bound would become invisible again (fail open), which the bound's comment states.
- The run phase records one timing observation of the evaluation on this machine as a data point (not a gate) in `progress.md`.

### D5 — Observability  [DECISION-OPEN]

- Options: (A) no new output; (B) name the count of directories read (or each source directory) in the `moai factory status` quota block and the `--auto` line.
- Default: (A).
- Why: the predecessor's byte-identical rules (REQ-QAS-022, D1) pin the output of every surface; a new key or cell in the status block changes the golden of AC-QAS-012 and AC-QAS-022 and so edits the predecessor's tests, for operator convenience only. A reading's source capture time is already shown in the block (`captured_at`); the source directory is not. If (B) is chosen later it is an additive follow-up with its own AC.

### D6 — The real-lane measurement  [DECISION-OPEN: which command is measured]

- Required by the leader as part of the run phase: one observation of the record directories around `moai factory next` from a real lane. Design in `plan.md` §C (protocol QWR-M0). `moai factory next` leases a card when it succeeds, so measuring it against the live queue adds an unwanted side effect.
- Options: (A) the non-leasing `moai factory status`, which calls the same `factoryQuotaEvaluate(root)` through `factoryQuotaStatusBlock(root)` and is documented read-only (`factory_card.go:1393`); (B) `moai factory next` run by the lane at the moment it would lease its next card anyway (no artificial lease); (C) `moai factory next` against a throwaway git repository via `CLAUDE_PROJECT_DIR` and lane environment variables.
- Default: (A) always, plus (B) when a real lease is due during the run phase. (C) is rejected as a default: it measures the throwaway repository's own record directory, which cannot answer the F1 question (where a real lane's record lives), and the root resolution on the command path is the adopting form (`todo.go:74-76`), whose side effects were not audited here.
- What the measurement establishes and cannot establish: `plan.md` §C.

### D7 — A neighbour test folded in as a refactor guard

- The extraction of the per-directory scan loop moves the line the predecessor's audit finding F3 says is not independently pinned (`quota.go:94`, the capture-time age check). One subtest in this SPEC's new test file pins it (file mtime at now, capture time 30m1s earlier, expect unknown). It is test-only, adds no requirement, and protects the refactor; the other neighbours (F4, F5) are not touched.

### D8 — A root without a git metadata directory

- Chosen default: the root's own record directory only, as before. A root whose `.git` is a regular file (the root is itself a linked worktree) is not followed to its common directory: the production root is always the primary checkout (D1 evidence), so following it is unreachable in production and costs two more file reads plus a parse for nothing. Where the root is such a worktree the gate degrades to today's behaviour (a known blind spot in a non-production shape, fail open).

## §C Requirements (GEARS)

### C.1 Reach

- **REQ-QWR-001** (Ubiquitous) — The quota aggregator shall provide a multi-directory reading that applies the freshness, clock-skew, max-age, file-modification pre-filter, per-window freshest-by-capture-time, and reset rules of REQ-QAS-005 to the union of the record directories it is given, a tie on capture time keeping the directory listed first, and the existing single-directory reading shall keep its signature and its result for every input it accepted before this change.
- **REQ-QWR-002** (Ubiquitous) — The quota gate's pressure evaluation shall read, in addition to the primary checkout's `.moai/state/context-usage` directory, the `.moai/state/context-usage` directory of every linked worktree named by the repository's git metadata entries (`<git-common-dir>/worktrees/<entry>/gitdir`), wherever on disk that worktree lives, in sorted entry-name order after the primary directory, and a reading that exists only in a linked worktree's directory shall be seen by the evaluation.
- **REQ-QWR-003** (Event-driven) — When a linked-worktree entry cannot be used — its `gitdir` file is unreadable, empty, not an absolute path, or names a path that no longer exists, or the worktree's record directory is absent, unreadable, or not a directory — the evaluation shall treat that entry as contributing no record and shall compute the aggregate from the remaining directories, with no error output, no panic, and no change of exit status.
- **REQ-QWR-004** (Ubiquitous) — The evaluation shall read at most 128 linked-worktree entries per call and at most 4 KiB of any `gitdir` file, and shall not read an entry beyond the bound.
- **REQ-QWR-005** (Unwanted) — The enumeration and the multi-directory reading shall not open a network connection, spawn a process, walk a directory tree recursively, or write anything under a state directory or the git metadata (REQ-QAS-007 extended to the new code).
- **REQ-QWR-006** (State-driven) — While the quota gate is disabled, the evaluation shall read no record directory and no git metadata.
- **REQ-QWR-007** (Ubiquitous) — Every consumer of the shared pressure evaluation (the lane gate on the CLI and the MCP twin, the status block, the `--auto` line, and the acquire warning) shall derive its reading from the cross-directory aggregate, and no hold line, exit status, output stream, output line, JSON key, or configuration key of REQ-QAS-008..022 shall change form: a reading that comes from a worktree directory shall render exactly as the same reading from the primary directory.
- **REQ-QWR-008** (Where) — Where the root has no git metadata directory (`.git` absent, or a regular file, or `.git/worktrees` absent), the evaluation shall read the root's own record directory alone, exactly as before this change.

### C.2 Evidence and scope

- **REQ-QWR-009** (Event-driven) — When the run phase begins, the record locations and modification times of a real lane's session, observed with the measurement protocol of `plan.md` §C immediately before and after the measured command, shall be recorded in `progress.md` and committed in their own commit before the first implementation commit.
- **REQ-QWR-010** (Unwanted) — The change shall not alter where a session writes its record (the state-anchor chain), the quota thresholds, any configuration key, any schema version, or any file of SPEC-QUOTA-AWARE-SCHEDULING-001, and every `TestQAS_` test shall pass without an edit to its file.

## §D Success criteria

The 12 acceptance criteria in `acceptance.md` (Given-When-Then, one decider command each) verify REQ-QWR-001..010; the traceability table is `acceptance.md` §C. The SPEC is done when every decider passes in a run whose swept count is non-empty, every `TestQAS_` decider named in AC-QWR-003 stays green, `moai spec lint` reports no error, and a scoped re-audit of the delta (sync-auditor, scoped to this SPEC's changed files) returns PASS or PASS-WITH-DEBT with the debt carried.

## §E Exclusions

### Out of Scope — the writer side

- Where a session writes its record: the state-anchor chain (`internal/stateanchor`, `internal/statusline/state_anchor.go`) is SPEC-STATE-ANCHOR-001's domain and is not changed (REQ-QWR-010). Records under a directory that is neither the primary nor a registered linked worktree (stray anchors) are not read.
- The record schema, the throttle, the heartbeat, and the exhausted-time field (`internal/statusline/context_usage.go`).

### Out of Scope — policy and routing

- The thresholds, the release margin, the max age, and every `workflow.quota_gate` key; the unmeasured defaults of the predecessor (DO-3) stay as they are.
- Automatic card-class routing, forced re-dispatch, a scheduler or wake-up at the reset time, and any change to the lane gate's arms, hold line, or exit status.
- Following a root whose `.git` is a file to its common directory (D8).

### Out of Scope — neighbours of the predecessor audit

- F4 (REQ-QAS-015 does not name the two shared registry functions), F5 (a release margin at or above the hold percentage passes validation, the latch line `factory_quota.go:145`), and the INFO items F6-F10: not touched. F3 is covered only as a test-only refactor guard (D7); no production line for F3 changes.
- Any edit to SPEC-QUOTA-AWARE-SCHEDULING-001 files, including its CHANGELOG limit sentence: the sync phase of this card updates the CHANGELOG sentence in a new entry; the predecessor's SPEC body is immutable.

## §F Residual risk

- The evaluation cost is unmeasured: about 1.6 times the file stats of today (1667 versus 1043 files on this machine), every 5 s per waiting lane and per `moai todo --auto` / `moai factory status` call. The bound (128 entries) limits the worst case; the run phase records one timing observation.
- A lane beyond the bound, or in a worktree whose `gitdir` is damaged, is invisible again (fail open): the direction of every failure is "no hold".
- A stale copy of a session record in a directory the session has left cannot win against a fresher copy, but it can win against nothing: while a session is quiet its last in-worktree copy ages out at `max_age` (30m) like any record. A session that last wrote a high reading inside a worktree 29 minutes ago still holds lanes for one more minute; the same is true of the primary directory today.
- Records written by the installed `moai` binary predate the predecessor SPEC (schema 2, no windows; measured: the session record read at 18:32 carries `"schema_version": 2` and no window field), so no window value from a real lane can be observed until the rebuilt binary is installed; the real-lane measurement establishes record locations and times only (D6).
- The enumeration assumes git's on-disk layout (`<common>/worktrees/<entry>/gitdir`). A git change to that layout would silently return the primary directory alone (nothing fails, nothing is read); AC-QWR-001's real-git subtest is the continued-firing guard that turns that silence into a red CI run.

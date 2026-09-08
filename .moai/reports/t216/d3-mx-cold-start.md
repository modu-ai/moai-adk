# t216 / D-3 — MX cold-start scan

> Investigation record, card t216, axis D-3. Read-only; HEAD `a9eb896ce`,
> worktree `.claude/worktrees/t216`, branch `WT-hook-wiring-drift`.
> Persisted by the orchestrator — the investigating agent is read-only and
> holds no Write tool.

## Claim

**Card:** "The MX cold-start scan is structurally incapable of completing, so
every session does dead work."

**Verdict: half right, and the wrong half is the important one.**

- *"Structurally incapable of completing"* — **confirmed**; the mechanism in
  the source is correctly described.
- *"every session"* — **overstated**. The scan is gated behind
  `mxIndexNeedsRebuild`, false whenever an `mx-index.json` younger than 7 days
  exists. In the primary checkout the index exists and is 4 days old, so the
  scan is **not spawned there at all today**. It fires in fresh worktrees and
  fresh clones — precisely the case the feature was written for.
- *"does dead work"* — **the framing overstates the cost**. The dead work is a
  goroutine killed within single-digit milliseconds. The waste is not latency
  or CPU; it is that **the feature does not work**. A correctness defect at
  ~zero latency cost, exactly as the source says.

**The finding that outranks all of this, and which the card does not mention:**
the scan's output has exactly **one** reader in the tree — `moai mx query`
(`internal/cli/mx_query.go:95-104`) — and that reader already fails loud with a
self-service remedy (`run 'moai mx scan'`) when the index is absent. The
capability at stake is narrow: *"make `moai mx query --kind DEBT` work in a
fresh worktree without a manual scan."* The consumer is live (the `/moai review`
lean audit invokes it), so this is not dead code — but it is a convenience, not
a dependency.

## Evidence

### 1. What Findings 5 actually says (verbatim)

`/Users/goos/MoAI/moai-adk-go/.moai/reports/hook-audit/load-path.md`:

```
### 5. The MX cold-start scan can never complete — dead work, every session

`spawnDeferredAdvisoryScans` (`session_start.go`) sends the advisory into the buffered channel and
*then* calls `runMXColdStartScan(projectDir)`. `Handle` returns as soon as it receives from that
channel; the CLI process then exits and the goroutine is killed mid-scan. Confirmed empirically: after
7 hook invocations against a project with 634 SPECs, `.moai/state/` contains
`active-sessions.json`, `config-cache.json`, `current-session-id.txt` — and **no `mx-index.json`**.
So `mxIndexNeedsRebuild` stays true forever and a doomed scan is spawned every single session.

The in-code comment asserting "the goroutine continues to completion in the background (durable side
effects still land)" is true for a long-lived process and false for a CLI that exits on return.

**Fix:** either run the MX scan synchronously under its own explicit time box when the index is
stale, or move it out of the hook into an explicitly-invoked command. As written it is pure waste.
Note this costs little *latency* today (it runs after the send), but it means the MX index feature is
non-functional on this load path.
```

The audit's own open question 3: *"Finding 5 (MX cold-start scan never
completes) is a correctness defect, not a latency one. Should it be split into a
separate card?"*

**Where the card's paraphrase diverges from the source:**

| | Source | Card |
|---|---|---|
| Scope | "every single session" — stated **conditionally on `mxIndexNeedsRebuild` staying true**, observed in a scratch project with no index | "every session", unconditional |
| Cost framing | "costs little *latency* today"; "the MX index feature is non-functional" | "every session does dead work" — reads as a recurring cost, which the source explicitly denies |
| Ranking | Findings are "ranked by how much startup latency each actually costs today"; #5 is fifth of six, i.e. near-zero cost | presented as live waste |

`rest-and-wiring.md`, `stop-chain.md`, `tool-path.md` contain no reference to the
MX cold-start scan (`grep -n "Findings 5\|MX cold\|mx-index\|ColdStart"` over all
three → no output). Findings 5 is the sole source.

### 2. What the cold-start scan is

Hook: `SessionStart`, wired in this worktree's `.claude/settings.json` —
matcher `startup|resume|clear|compact|fork`, `timeout: 30`, synchronous.
`.claude/hooks/moai/handle-session-start.sh` → `moai hook session-start` →
`internal/hook/session_start.go`.

"Cold start" = the sidecar index `.moai/state/mx-index.json` is absent, empty,
corrupt, has a zero `ScannedAt`, or is older than 7 days. The gate is the cheap
synchronous stat+field read `mxIndexNeedsRebuild` at
**`internal/hook/session_start.go:1499`**, called at **line 228**. The scan is
`runMXColdStartScan` at **line 1536**, dispatched from **line 572**
(production/async branch) and **line 272** (test/inline branch).

Declared intent, `internal/hook/session_start.go:1473-1479`:

```
// mxIndexFreshnessThreshold is how long an MX sidecar index is considered
// fresh. An index whose ScannedAt is older than this (or absent/corrupt)
// triggers the deferred cold-start full scan so 'moai mx query' returns fresh
// results without a manual 'moai mx scan' after checkout/clone/worktree
// creation. Measured staleness on a fresh worktree (2026-08-04): 764 missing
// tags (1,567 actual vs 803 indexed). 7 days mirrors the MX ArchiveStale TTL.
const mxIndexFreshnessThreshold = 7 * 24 * time.Hour
```

### 3. Why it cannot complete — one cause, not several

All five candidate causes were checked. Four are innocent.

**Not the timeout.** `mxIndexScanTimeoutDefault = 2 * time.Second`
(`session_start.go:1488`). The scan is far faster. `moai mx scan --dry` uses the
identical scanner and ignore set as `runMXColdStartScan`
(`mx_scan.go:57-59` vs `session_start.go:1548-1551`) and returns before any
write (`mx_scan.go:76-82`). Verified non-mutating: `.moai/state/` unchanged
across all six runs.

```
$ /usr/bin/time -p moai mx scan --dry
DRY RUN: 925 tags would be written (index not saved)
by kind:
  NOTE: 438
  ANCHOR: 398
  WARN: 76
  TODO: 2
  DEBT: 8
  LEGACY: 3
DEBT rotRisk (missing @MX:UPGRADE): 2
real 0.40

$ for i in 1 2 3 4 5; do /usr/bin/time -p moai mx scan --dry --quiet; done
real 0.65
real 0.31
real 0.40
real 0.33
real 0.32

$ /usr/bin/time -p moai version
real 0.05
```

Binary floor 0.05 s at load 67, so **ScanDir itself is ~260-350 ms** — 6-8×
inside the 2 s box. (Load average at capture: `load averages: 67.13 105.93
99.40`. Absolutes are inflated upper bounds; the ratio to the binary floor is
the load-bearing figure.)

**Not unbounded scope.** 925 tags across ~451 files, bounded by
`mx.DefaultScanIgnore`, and time-boxed anyway.

**Not a cache written where the next session cannot read it.**
`Manager.writeWithoutLock` (`internal/mx/sidecar.go:96-118`) writes
`mx-index.json.tmp` then `os.Rename` — atomic, correct path, same `.moai/state`
the reader uses. No stray `.tmp` on disk.

**Not an early return or a budget cap.** `mxIndexNeedsRebuild` correctly returns
true; the 250 ms join bound does not gate the scan — the scan is dispatched
*after* the channel send.

**The actual cause: the goroutine outlives nothing.**
`spawnDeferredAdvisoryScans`, `session_start.go:568-573`:

```go
		advisory := h.computeDeferredAdvisory(projectDir, driftFn, driftTimeout)
		// Send advisories FIRST (buffered channel → never blocks even after
		// the join bound elapses), THEN run the MX cold-start scan as a
		// best-effort durable side effect.
		resultCh <- advisory
		if mxScanNeeded {
			runMXColdStartScan(projectDir)
		}
```

`Handle` returns on receipt from `resultCh` (lines 246-250) or on the 250 ms
timer (line 253). `internal/cli/hook.go:334-353` then does `writeHookOutput` and
the deferred `rs.Shutdown()` trace flush, and `main` returns — Go tears down
every remaining goroutine. There is no `WaitGroup`, no join, no daemon.
`deferredScansAsync = true` in production (`session_start.go:1422`; only
`main_test.go:47` flips it false), so the inline branch at line 272 — the one
that *would* complete — never runs outside tests.

Two independent lethal orderings, not one:

1. If the advisories finish inside 250 ms (the common case), `Handle` returns
   the instant the send lands, and the scan starts with a few ms of process life
   left.
2. If the advisories exceed 250 ms, `Handle` returns on the timer and
   `runMXColdStartScan` **has not even been reached**.

The comment at `session_start.go:1531-1533` is candid (*"In production the
SessionStart process may exit and kill the helper goroutine … the scan only
lands if it finishes before the process exits"*), while the comment at lines
255-257 (*"The goroutine continues to completion in the background (durable side
effects still land)"*) asserts the opposite. Findings 5's reading is correct.

**Same structural bug elsewhere, unflagged by the audit and out of this card's
scope but worth a note:** `internal/hook/file_changed.go:110-118` spawns the MX
sidecar `UpdateFile` in a `context.Background()` goroutine with a 5 s deadline in
the same short-lived CLI process. That is the incremental index-maintenance path,
and it dies the same way.

### 4. Cost per session

- **Configured hook timeout:** 30 s. Never approached.
- **Scan's own time box:** 2 s. Never approached.
- **Work the scan needs:** ~260-350 ms of `ScanDir` at load 67.
- **Process life available to it:** the portion of SessionStart after `Handle`
  returns — `writeHookOutput` plus one trace-writer flush. The load-path audit
  measured the whole hook at **48.7 ms median** (load 8-15) against a **30 ms**
  bare `moai version` floor, so all handler work is 18-43 ms and the
  post-dispatch tail is a small fraction. Conservatively **<10 ms**.
- **Therefore under ~4% of the walk completes before the process dies.**
- **Latency cost to the user: zero.** The scan is off the synchronous path in
  both branches. The only synchronous cost is `mxIndexNeedsRebuild`: one
  `os.Stat` + one `os.ReadFile` + one unmarshal, sub-millisecond.
- **CPU cost: a few milliseconds of a doomed walk, once per cold session.**

The card's "dead work" is real but tiny. The loss is the capability, not the
cycles.

### 5. Does the scan produce anything usable today?

```
$ find .../.claude/worktrees -maxdepth 4 -name "mx-index.json" -exec ls -la {} \;
-rw-r--r--@ 1 goos staff 376314 Aug 19 23:21 .../worktrees/t134/.moai/state/mx-index.json
-rw-r--r--@ 1 goos staff 374291 Aug 13 23:21 .../worktrees/v31-m4-ko-content/.moai/state/mx-index.json

$ find .../.claude/worktrees -maxdepth 3 -type d -name state | wc -l
     153
```

**2 of 153 worktrees with a `.moai/state` have an index.** Both are days old
with no subsequent refresh — consistent with a one-off manual `moai mx scan`,
not with a cold-start scan that would have re-fired every session. This worktree
(t216, created 21:44 today, SessionStart has fired) has none:

```
$ ls -la .moai/state/
-rw-r--r--@  1 goos  staff     0 Aug 24 21:44 .gitkeep
-rw-r--r--@  1 goos  staff 13866 Aug 24 21:44 config-cache.json
-rw-r--r--@  1 goos  staff   271 Aug 24 21:46 context-usage.json
```

Findings 5's empirical result, reproduced at n=153 rather than n=7.

The primary checkout **does** have one, neither partial nor empty — complete and
mildly stale:

```
schema_version 2
scanned_at 2026-08-20T19:08:50.521188+09:00
tags 924
[('NOTE', 437), ('ANCHOR', 397), ('WARN', 77), ('DEBT', 8), ('LEGACY', 3), ('TODO', 2)]
distinct files 451
```

924 tags at Aug 20 vs 925 measured live today — a full-project artifact, 4 days
old, **inside** the 7-day window. So in the primary checkout
`mxIndexNeedsRebuild` returns **false** and the doomed scan is not spawned at
all. This directly refutes the card's "every session". No partial or corrupt
artifacts anywhere (atomic rename; no `.tmp` residue).

**Who reads it** (checked before recommending anything be preserved):

```
$ grep -rn "mx.NewResolver\|mx.NewManager" --include="*.go" . | grep -v _test | grep -v /worktrees/
internal/cli/mx_query.go:95:   mgr := mx.NewManager(stateDir)
internal/cli/mx_query.go:106:  resolver := mx.NewResolver(mgr)
internal/cli/mx_scan.go:86:    mgr := mx.NewManager(stateDir)
internal/hook/file_changed.go:236: manager := mx.NewManager(stateDir)
internal/hook/session_start.go:1569: mgr := mx.NewManager(stateDir)
```

Three of the five are **writers**. The **single reader is `moai mx query`**. The
other two `ScanDir` callers — `internal/graph/graph.go:203` and
`internal/navigator/sync/mx_bridge.go:47` — re-scan from source and never touch
the index. No hook, no statusline, no MCP tool reads it.

`moai mx query` is **live, not merely referenced**:
`.claude/skills/moai/workflows/review.md:289` — *"Read the `@MX:DEBT` harvest via
`moai mx query --kind DEBT`"* — is a step in the `/moai review` lean audit, and
the same text ships in `internal/template/templates/…/review.md:289`, so it is on
every downstream project.

But the reader self-heals, `mx_query.go:98-104`:

```
SidecarUnavailable: sidecar index does not exist — run 'moai mx scan' to build the index
```

And no automation ever runs that remedy: `grep -rn "moai mx scan" .claude/commands
.claude/skills .claude/agents` → no hits. `/moai:mx` is an annotation-authoring
command, not an index build.

### 6. Options

**Stated first:** the output *is* consumed, but by a single human/agent-invoked
command that already self-heals with a documented one-liner. Nothing automated
depends on the index. The capability being defended is *"`moai mx query` works in
a fresh worktree without the operator typing `moai mx scan`"* — worth ~300 ms of
someone's time, once per worktree.

| Option | Change | Cost | Capability gained / lost |
|---|---|---|---|
| **A. Run it synchronously under its own time box** (source's first fix) | move the `mxScanNeeded` branch onto the sync path, keep `mxIndexScanTimeout` | +260-350 ms on cold sessions only; 0 ms warm; inside the 30 s hook timeout but ~6× the whole current hook | **Gained:** the feature works. **Lost:** cold-start SessionStart ~49 ms → ~400 ms, paid by every fresh worktree — ~150 in this repo's workflow |
| **B. Remove the cold-start scan from the hook** (source's second fix) | delete line 228, the `mxScanNeeded` arg at 238, lines 272 and 572, and `runMXColdStartScan` | ~40 lines deleted; removes the misleading comment at 255-257 | **Gained:** honesty, one less sub-ms stat, no dead goroutine. **Lost:** nothing that works today. Strictly dominant over the status quo |
| **C. Run it out of band** | fire `moai mx scan` from the `worktree-create` hook event (`internal/cli/hook.go:64`), or from `moai mx query` itself on `SidecarUnavailable` | small; the query-side auto-build is ~5 lines | **Gained:** the intended capability, paid exactly when someone needs it, by the process that needs it. **Lost:** nothing. Best value per line |
| **D. Make it incremental** | rely on `file_changed`'s `UpdateFile` | — | **Does not work today for the same structural reason** (`file_changed.go:110-118`). Fixing it fixes maintenance, not cold start: `UpdateFile` merges one file at a time and never produces the initial 451-file set |
| **E. Persist partial progress** | checkpoint the walk | high — needs a resumable scanner, cursor, and merge; the scanner has none | Nothing worth it. Making a 300 ms operation resumable is absurd next to just running it |
| **F. Raise the budget** | increase `mxIndexScanTimeout` | trivial | **Nothing.** The 2 s box is never reached. This is the fix the card's "budget cap" phrasing invites, and it would change nothing. Explicitly ruled out |

**Recommendation: C, with B as the cleanup it implies.** Auto-build in
`mx_query.go` on `SidecarUnavailable` (or a stale index), and delete the
hook-side scan. That delivers the capability the comment at 1473-1479 promises,
on a path where a 300 ms wait is expected and attributable, and removes ~40 lines
that have never once produced their artifact in 153 worktrees. **Do not choose A**
without the lead accepting an ~8× cold-session hook regression for a convenience.
**Do not choose F.**

## Baseline-attribution

Every defect here is **pre-existing on this branch's baseline**, not introduced
by t216. HEAD `a9eb896ce fix(security): restore the pre-write ast-grep deny
capability (t227) (#1637)`; the worktree is clean of MX changes. The code
originates from SPEC-MX-ACTIVATION P0-1 (cited at `session_start.go:221`), and
the stale-index measurement in its comment is dated 2026-08-04. Findings 5 was
recorded 2026-08-24 04:12 against the same tree. `origin/main` state is NOT
covered — the load-path audit records that `origin/main` is ~241 commits ahead
and makes no claim about shipped code; neither does this report.

## Gaps

1. **The real SessionStart hook was not run.** It writes
   `.moai/state/active-sessions.json`, `current-session-id.txt`, and possibly
   `.claude/settings.local.json`; invoking it with a synthetic UUID during an
   active factory run would corrupt `moai session current` for live lanes. The
   <10 ms post-`Handle` tail is therefore inferred from the load-path audit's
   48.7 ms total / 30 ms binary floor, not directly measured.
2. **Scan duration is a proxy.** `moai mx scan --dry` shares the scanner, ignore
   set, and walk with `runMXColdStartScan` but is a separate call site. The
   number excludes the `json.MarshalIndent` + atomic write of a 357 KB file the
   real scan would also pay — add a few ms.
3. **Load-inflated absolutes.** Load average 67.13 at capture. All timings are
   upper bounds; the ScanDir/binary-floor ratio is the reliable signal.
4. **Provenance of the two surviving indexes** (t134, v31-m4-ko-content) is
   inferred from age and completeness, not traced. A lucky race cannot be ruled
   out to certainty — but at ~4% completion odds per attempt, 2-of-153 is far
   better explained by two manual scans.
5. **`file_changed`'s identical structural bug is asserted from code, not
   measured.** No FileChanged event was fired. Out of this card's axis; flagged
   for triage.
6. **Downstream user projects were not surveyed.** The template ships both the
   review-skill consumer and the hook, so the same defect ships, but no installed
   project was inspected.

## Residual-risk

- **If option A is applied without a latency budget**, cold-session SessionStart
  grows ~8×, paid ~150 times in this repo's worktree workflow. The audit's own
  Finding 2 identifies process start as the dominant term precisely because
  handler work was driven down to 18-43 ms; A undoes that on the cold path.
- **If B is taken without C**, `moai mx query --kind DEBT` in the `/moai review`
  lean audit keeps failing in fresh worktrees with `SidecarUnavailable`. That is
  the status quo, so removal is not a regression — but it makes a
  currently-invisible gap visible, and someone will file it.
- **The stale-but-fresh-enough case is untested by anything.** A 6-day-old index
  passes `mxIndexNeedsRebuild` and is served to `moai mx query` as authoritative;
  the comment records 764 missing tags on a fresh worktree at the last
  measurement. Whatever is decided about cold start, the 7-day window means
  `mx query` can return materially stale DEBT harvests to a live review.
- **The misleading comment at `session_start.go:255-257`** contradicts the
  accurate one at 1531-1533 and will mislead the next reader into believing any
  fire-and-forget goroutine in a hook process is safe. It should be corrected
  regardless of which option is chosen.

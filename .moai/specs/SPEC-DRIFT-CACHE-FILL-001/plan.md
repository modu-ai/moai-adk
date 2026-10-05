# SPEC-DRIFT-CACHE-FILL-001 — Implementation Plan

Card **t871**. Worktree `.claude/worktrees/t871`, branch `WT-drift-cache-fill`,
base local `develop` `023664665`.

Milestones are ordered by **decision-reversibility**: the child's command
surface and process model first (the decisions most likely to change at review),
the coordination design next, and the mechanical edits last. Execution order
follows the same sequence.

---

## §A Context

Everything the plan rests on is in `spec.md §A`. One line of it governs the
design: the drift compute (p50 967 ms) is strictly longer than the join bound
(250 ms) in a process that exits when `Handle` returns, so the cache has no
writer on this path. The fix is a writer that outlives the hook process.

---

## §B Constraints carried into the design

| # | Constraint | Where it is discharged |
|---|---|---|
| 1 | Several sessions start **simultaneously** on this machine | M3 — the **parent** claims before spawning, so a burst produces one child, not N |
| 2 | A spawned process must be bounded from outside; a trailing kill is not cleanup | M1 — the deadline rides on the child's own invocation and is enforced by `DriftCountCtx`, whose main goroutine returns at the deadline and takes the process with it |
| 3 | SessionStart must not become a process fan-out | M3 — the record gates the **spawn**; the lock gates the **compute**. Both are needed: the record alone races, the lock alone admits the burst |
| 4 | Fail-open everywhere | M5 |
| 5 | The hook package's tests must not spawn fill processes | M5 — reuse `deferredScansAsync`, `snapshotDeferredScanCompleted`, `registerDeferredScanSeam` |
| 6 | A stale lock must not wedge the fill forever | M3 — the lock is held by an OS handle released on process exit, including `SIGKILL` |
| 7 | A `workflow.*` key is a typed Go field, not YAML text | M5 — `internal/config/types.go` + `internal/config/defaults.go` are the SSOT; the YAMLs are overrides |
| 8 | The template mirror crosses a CI neutrality guard | M5 — the template-side comment carries no SPEC id, no REQ token, no date |

---

## §C Pre-flight

- `internal/statusline/landed.go` `maybeRefreshLandedCounts` — the detached-child
  idiom, including the spawn-probe placement comment. **Read its stampede guard
  critically**: it has the child write first, which is correct for a status bar
  rendering sequentially inside one session and wrong here, where sessions burst
  simultaneously. M3 departs from it deliberately.
- `internal/atomicfile/replace.go` — `Claim` (lines ~37-58: the doc comment
  states the exclusivity guarantee on both platforms and why it never retries)
  and `Replace`. `internal/verify/claim_lock.go` — the two-layer composition,
  the **claim-a-companion-`.lock`** target (line ~60), and the
  staleness-judged-at-removal-time reclaim loop.
- `internal/cli/gate_lock_cli_test.go` — the cross-process test shape
  AC-DCF-006 clause (a) uses: an env-gated helper `Test…` re-executed via
  `os.Args[0]` + `-test.run=^<helper>$`, self-bounding, with `-test.timeout`
  capping it from outside.
- `internal/sessionmsg/lock_unix.go` + `internal/sessionmsg/lock_windows.go` —
  the non-blocking exclusive lock pair, already `LOCK_EX|LOCK_NB` on POSIX and
  `LOCKFILE_EXCLUSIVE_LOCK|LOCKFILE_FAIL_IMMEDIATELY` on Windows. Copy the pair
  shape; do not write a variant, and do not add a platform decision.
- `internal/config/types.go` (`BranchGuard`, line ~452) and
  `internal/config/defaults.go` (~line 940) — the field-plus-default pattern the
  opt-out must follow, and the home for the TTL and deadline constants
  (`CLAUDE.local.md` §14 forbids hardcoded thresholds).
- `internal/hook/session_start_guard_liveness.go` — the same pull-based /
  never-awaited / persisted-for-a-later-activation shape one layer over; match
  its comment discipline.
- `.github/workflows/template-neutrality-check.yaml` — fires on
  `internal/template/templates/**`.

---

## §D Milestones

### M1 — the child entry point and its command surface *(highest change-likelihood)*

The decision most worth challenging: **how the child is invoked**. Proposal — a
hidden flag on the existing drift command rather than a new verb:

```
moai spec drift --fill-cache --fill-timeout <duration>
```

**The comparison, stated rather than concluded.** The plain verb `moai spec
drift` already fills the cache with no new surface at all: it routes through
`spec.DetectDrift` → `detectDrift` → `saveDriftCache` (`internal/spec/drift.go`
line ~290), and the t666 R3 control arm measured exactly that. The detached
child's nil streams and unread exit status already neutralise the plain verb's
output and its `--exit-code-on-drift` behaviour. So the flag's genuine additions
are exactly two: **the self-imposed deadline** (REQ-DCF-003) and **the lock**
(REQ-DCF-007). Both must live inside the child — an external supervisor is the
trailing-kill anti-pattern §G forbids — and neither can be expressed on the
plain verb. That is the whole argument; if a reviewer disagrees, the alternative
is the plain verb plus an accepted loss of both properties.

Work:
- `internal/cli/spec_drift.go`: add the hidden flag pair and its `RunE` branch,
  routing through `spec.DriftCountCtx` under `context.WithTimeout`.
- Make the drift cache write **atomic** by routing `saveDriftCache`
  (`internal/spec/drift_cache.go`) through `internal/atomicfile.Replace` —
  the in-tree primitive, not a hand-rolled temp-plus-rename. This discharges
  REQ-DCF-005. The deadline is engineered to fire near completion, which is
  exactly when a non-atomic `os.WriteFile` of an ~869-record JSON is most likely
  in flight; `loadDriftCache` would fail open on the truncated result, but a torn
  file is prevented rather than tolerated. Note `Replace` and `Claim` are
  counterparts and are never substituted for each other: `Replace` makes a path
  hold this content (an existing destination is success); `Claim` (M3) asks who
  gets to proceed (an existing path is failure).
- Constants declared in `internal/config/defaults.go`, not inline
  (`CLAUDE.local.md` §14): the `--fill-timeout` default (the child deadline) and
  the suppression TTL, with **TTL ≥ deadline + slack** as a documented
  relationship between the two rather than two unrelated numbers.

**Flips**: `AC-DCF-013`, and half of `AC-DCF-004`.

### M2 — the parent's spawn and its detach contract

- `internal/hook/session_start_drift_fill.go` (new): `maybeFillDriftCache` —
  `os.Executable()` → basename guard → nil streams → `Start()` →
  `Process.Release()`, never `Wait()`.
- `driftFillSpawnProbe` seam placed after the suppression gates and before the
  self-invocation guard, with the placement rationale copied from
  `landedSpawnProbe` — the placement is what makes "gated before spawning"
  distinguishable from "spawn blocked by the basename guard".
- `@MX:WARN` + `@MX:REASON` on the spawn, naming the fire-and-forget contract and
  the child's self-bound.

**Observation-method item (operator disposition, iteration-3 audit).**
`AC-DCF-003`'s green path says the non-`moai` basename case records a seam
attempt "while no process is started". That criterion is **not** a must-fix and
is not being re-authored; the operator classified the open question as a
run-phase M2 item. This milestone therefore owes a specified **observation
method for "no process started"** — how the test establishes the negative
without asserting against a real process, given that `isSelfInvocable` blocks the
exec under `go test` anyway (`internal/statusline/forge_spawn_gate_test.go`
states this in its own words). Decide it here, record it beside the test, and
leave the criterion text alone.

**Flips**: `AC-DCF-001`, `AC-DCF-003`, and the rest of `AC-DCF-004`.

### M3 — coordination: the parent claims, the child locks

Two layers, and the split is the correction the plan-audit forced.

- **Suppression record — CLAIMED by the PARENT, before `Start()`.** The parent is
  the only actor guaranteed to be alive at that instant.
  `.moai/state/drift-fill.json` carries `{head_sha, started_at}`.

  **The claim target is a separate lock, never the record.**
  `internal/atomicfile.Claim(path, perm)` —
  `os.O_CREATE|os.O_EXCL|os.O_WRONLY` then `Close()`, atomic on POSIX,
  `CREATE_NEW` on Windows, **never retried** because a retry would hand the same
  claim to a second caller — is applied to `drift-fill.json.lock`, the companion
  of the record. This follows `internal/verify/claim_lock.go`, which claims
  `SnapshotPath(...) + ".lock"` rather than the snapshot, and which re-evaluates
  staleness **at removal time** instead of trusting an earlier read. Copy that
  composition; do not author a variant.

  The critical section is: acquire the lock → read the record → judge expiry →
  remove/rewrite → release. Winning the lock authorises the spawn; failing to
  acquire it means another session is inside the section and this one starts
  nothing.

  **Why claiming the record itself is wrong** (both found by the iteration-3
  audit):
  - **Reclaim TOCTOU.** `os.Remove` of an expired record can destroy another
    handler's fresh claim: A reclaims and wins; B, judging on a read taken
    before A, removes A's record and also wins. Containing every read, judgement
    and removal inside the lock removes the window by construction.
  - **The empty-record window.** `Claim` writes no content, so the record would
    exist **empty** between the create and the content write, and the
    disposition of an unparseable record was unspecified — the two readings
    diverge into a double spawn or a contradiction. With the lock separate, the
    record is written whole under it; an unreadable record is specified as
    **expired**, one disposition, not two.

  **Why a plain write is not enough** (the v0.2.0 defect): a write preceded by a
  read is a read-check-then-write, and the parent holds no lock. N handlers
  released simultaneously all complete their **reads** before the first **write**
  lands, so all of them spawn. "The first write precedes any child" is true and
  irrelevant — the ordering that matters is *first write before other reads*, and
  only an exclusive create provides it. An in-process `sync.Mutex` is explicitly
  **not** the fix: it would make a goroutine-based test green while leaving two
  *processes* racing, which is the mutant AC-DCF-006 is written to exclude.

  Reclaim: a record whose `head_sha` is not the current HEAD, or whose
  `started_at` is older than the TTL, or which is stamped in the future (a clock
  stepped backwards, or a clone carrying a foreign stamp — otherwise it would
  suppress every fill until HEAD moves) is removed and the claim re-attempted
  **once**. A handler that loses the re-attempt starts no fill.

  - This closes the **burst** hole at the only layer that can close it: exactly
    one of N simultaneous handlers wins the create. The claim is one small
    exclusive-create on the parent's path — bounded, best-effort, and a failure
    is fail-open (worst case an extra child, never a blocked hook), which is why
    it sits inside REQ-DCF-013's "beyond the existing join bound" qualifier.
  - It also closes the **broken-child** hole: a child that dies before executing
    leaves the claimed record standing, so REQ-DCF-010 holds for the most likely
    failure mode.
  - **TTL ≥ child deadline + slack** — restored from v0.1.0 and load-bearing: a
    TTL shorter than the deadline lets a second session reclaim a record whose
    child is still legitimately computing, which reintroduces the burst one TTL
    later. Both constants resolve from `internal/config/defaults.go` (M1), and
    the inequality is asserted over the resolved values, not over literals.
- **Lock — taken by the CHILD, non-blocking.** `LOCK_EX|LOCK_NB` on
  `.moai/state/drift-cache.lock` via the `internal/sessionmsg` pair shape; a
  loser exits 0 computing nothing. Held only by an OS handle, so process exit —
  `SIGKILL` included — releases it. The parent never holds it.

The record is the cheap spawn-suppressor and can be raced; the lock is the
authoritative compute-serialiser and cannot. Neither replaces the other.

**Flips**: `AC-DCF-005`, `AC-DCF-006`, `AC-DCF-007`, `AC-DCF-008`.

### M4 — miss-path decoupling (D1)

The mechanical half. In the deferred advisory path of
`internal/hook/session_start.go`, replace the in-band `driftFn` call with a
cache-only resolve: hit → render `status_drift_warning` from the cached report,
as today; miss → omit the key and call `maybeFillDriftCache`. The other three
deferred scans and the bounded join are untouched.

**Flips**: `AC-DCF-002`, `AC-DCF-009`.

### M5 — opt-out, fail-open, test isolation

- **The config key, properly wired** — `DriftCacheFill DriftCacheFillConfig
  \`yaml:"drift_cache_fill"\`` in `internal/config/types.go`, default
  `Enabled: true` in `internal/config/defaults.go`. Only then does the YAML key
  mean anything: without the field it decodes into nothing and the feature is
  permanently on.
  - The **default-true** choice departs from the `workflow.*` guard family, whose
    comments ground the default on **neutrality** rather than on "adds a deny".
    This feature is not neutral when enabled — it starts an unsolicited child on
    every miss on every user's machine. The cost is accepted and stated in
    `spec.md §D` assumption 4; it is a Kickoff decision, not an implementation
    detail.
  - Local `.moai/config/sections/workflow.yaml` documents the key. The template
    mirror carries it with a **neutral** comment — no SPEC id, no REQ token, no
    date — because `.github/workflows/template-neutrality-check.yaml` runs
    `TestTemplateNoInternalContentLeak` on that path in both tiers.
- **Fail-open** at every failure point: return silently at `slog.Debug`, matching
  the existing best-effort contract on this path.
- **Test isolation**: the real spawn is reachable only past
  `asyncDeferredScans()`.

**Flips**: `AC-DCF-010`, `AC-DCF-011`, `AC-DCF-012`.

### M6 — give the instrument a verdict

Today the drift-cache probe measures, prints, and `t.Skip`s without its env vars
— so an ordinary package run prints `ok` whether the probe is healthy, broken, or
deleted. That is the §1.3 continued-firing defect, inside the very instrument
this card relies on.

- Extract the run-set classifier into a testable function returning an explicit
  verdict, with a distinct non-pass for an empty run set.
- Add `MOAI_DRIFT_CACHE_PROBE_ASSERT`: `t.Fatalf` on a non-pass verdict; print the
  swept run count either way.
- Add an **always-on** unit test asserting the probe function exists by name and
  reads its expected env gate, so deleting or renaming the probe turns the
  ordinary `go test ./internal/hook/...` red.

**Flips**: `AC-DCF-014`.

### M7 — measurements of record (not gates)

- Optional field baseline: run the probe's exit arm on a `git clone --shared`
  copy before M4 lands, for corroboration only (`AC-DCF-015`). The card's close
  does not depend on it — `AC-DCF-013` is the release-blocking carrier.
- Re-run the stage probe warm and cold; write Handle p50, the drift stage p50,
  and the machine's load average into `progress.md §E.2`, labelled measurements
  rather than gates (`AC-DCF-016`).

---

## §E Self-verification

- Scoped: `go test ./internal/hook/... ./internal/spec/... ./internal/cli/...
  ./internal/config/...`, plus `-race` on `internal/hook`.
- **No local `go test ./...`** — the full-suite verdict is CI's
  (`CLAUDE.local.md` §4).
- `GOOS=windows GOARCH=amd64 go build ./...` for the build-tagged lock pair.
- `go test ./internal/template/... -run 'TestTemplateNoInternalContentLeak'`,
  both tiers, after the template edit.
- `go vet ./...`, `golangci-lint run` on the touched packages.
- `make build` after the config-template edit; verify the embedded template
  carries the key.

---

## §F Risks

| Risk | Mitigation |
|---|---|
| A cold session shows no drift advisory until the next session | Accepted by D3 and stated in `spec.md §D`; today the advisory never arrives |
| Shipping enabled starts an unsolicited child on every user's machine | Stated as an accepted cost in `spec.md §D` assumption 4; bounded by the deadline, the lock, the TTL, and silence; reversible via the opt-out |
| The parent's record write adds work to the latency-bounded path | One small JSON write, best-effort; a failure is fail-open and costs at most one extra child |
| The child re-enters the hook | The child is `moai spec drift`, not `moai hook session-start`; no hook path is reachable from it |
| `os.Executable()` resolves to a test binary under `go test` | The basename guard blocks it, as it does for the statusline children; the seam still counts the attempt so the test has a verdict |
| The TTL masks a permanently broken fill | Out of scope to alarm on; the probe's asserting mode (M6) is the instrument that catches it when run |

---

## §G Anti-patterns to avoid

- Awaiting the child, or reading its exit status — reintroduces the latency this
  card removes.
- A trailing `kill`, or an external supervisor, as the cleanup story; the child
  bounds itself or it does not ship.
- A blocking lock, which serialises sessions instead of skipping.
- A pid-file lock, which wedges on `SIGKILL`.
- Letting the **child** write the suppression record — the v0.1.0 design; it
  races a simultaneous burst and leaves nothing behind when the child is broken.
- A **plain write** of the record by the parent — the v0.2.0 design; a
  read-check-then-write is lockless, and N simultaneous handlers all read before
  the first write lands.
- An **in-process `sync.Mutex`** as the burst fix. Eight goroutines are not eight
  processes: it turns an in-process AC-DCF-006 green while leaving the real race
  open, which is the definition of a mutant.
- **Claiming the record instead of a companion lock.** It reintroduces the
  reclaim TOCTOU and the empty-record window (M3 above).
- **Asserting a cross-process property with an in-process test, or with an error
  value.** From the iteration-3 plan-audit, preserved verbatim at the operator's
  instruction: **"an error value is not evidence of a syscall — `fs.ErrExist` is
  a sentinel any code can wrap."** A mutex-guarded `os.Stat` that returns an
  error wrapping the sentinel satisfies both a goroutine-based burst test and an
  existing-file check, while two real processes still race. The only instrument
  that witnesses the property is cross-process create-versus-create
  (AC-DCF-006 clause (a)).
- Hand-rolling temp-plus-rename, or a retrying claim. `atomicfile.Replace` and
  `atomicfile.Claim` exist; a retried claim hands the same claim to a second
  caller and destroys the exclusion it exists to provide.
- A sequential-loop test standing in for a concurrent burst; it is the one shape
  under which the wrong design also passes.
- Adding the YAML key without the typed field — a dead key and a permanently-on
  feature.
- Reporting the probe's printed output as a verdict.

---

## §H Cross-references

- `spec.md §A.4` — the in-tree capabilities this plan reuses rather than rebuilds.
- `acceptance.md` — each milestone above names the criteria it flips.
- `.claude/rules/moai/development/verification-completeness.md` — the discipline
  M6 exists to satisfy.

🗿 MoAI

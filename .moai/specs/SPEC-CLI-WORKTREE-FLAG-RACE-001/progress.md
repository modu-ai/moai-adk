# Progress — SPEC-CLI-WORKTREE-FLAG-RACE-001

Card `t464` · worktree `WT-worktree-flag-race` · base `d592b0551`.

## §E.1 Plan-phase Audit-Ready Signal

**Claim** — the plan-phase artifact set for SPEC-CLI-WORKTREE-FLAG-RACE-001 is authored at Tier M
(`spec.md` + `plan.md` + `acceptance.md` + `progress.md`), status `draft`, with both repair options
carried and no winner declared.

**Evidence**

- SPEC ID regex self-check, executed:
  `[[ "SPEC-CLI-WORKTREE-FLAG-RACE-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- Sibling-scope correction measured in this tree:
  `grep -n 'func Test\|t.Parallel()\|findProjectRootFn\|launcherWorktreeMaterialize\|t.Cleanup' internal/cli/worktree_branch_flag_test.go`
  → four `TestResolveWorktreeExistingBranch_*` functions at lines 65 / 95 / 125 / 160, each with a
  `t.Parallel()` on the following line and seam assignments plus `t.Cleanup` restores.
- Blast-radius measurement: `grep -rln 'findProjectRootFn = \|launcherWorktreeMaterialize = ' internal/cli/*_test.go`
  → 23 files. Two independent function-body scans across the same set (this lane's, and the
  auditor's `awk` scan resetting at each top-level `}`) returned the four in-file siblings and
  **zero** out-of-file `t.Parallel()`-plus-global-write hits.
- RED evidence pre-existing and non-empty: `.moai/reports/t464/red-race-d592b0551.txt`, 1369 lines
  (`wc -l`).
- Frame attribution:
  `grep -o 'internal/cli/[a-z_]*\.go' .moai/reports/t464/red-race-d592b0551.txt | sort | uniq -c`
  → `46 main_test.go`, `52 worktree_branch_flag_test.go`, `5 worktree_branch_flag.go`. The
  production file's 5 frames are at report lines 291 (`:70`) and 461 / 920 / 1288 / 1361 (`:74`).
- RED truncation: `sed -n '1340,1369p' …/red-race-d592b0551.txt` → `panic: Log in goroutine after
  TestResolveWorktreeExistingBranch_NoFlagIsNoop has completed: materialize must not run without
  --branch` at line 1344, then `FAIL … 1.543s`.
- Tree identity at authoring: `git rev-parse --short HEAD` → `d592b0551`;
  `git branch --show-current` → `WT-worktree-flag-race`.

**Baseline-attribution** — all of the above were run in this worktree, in this run, against
`d592b0551`. The RED figures (exit 1, `≥ 23` warnings — a floor from a truncated run) are
**carried from the lane's own prior measurement** recorded in
`.moai/reports/t464/red-race-d592b0551.txt`; the frame-attribution and truncation greps above were
run against that file in this run and are this agent's own measurements.

**Gaps**

- No `-race` run was executed during plan-phase; the RED is cited from the persisted capture, and
  the frame/truncation facts were read out of it rather than re-produced.
- The linux/amd64 behaviour of the RED command is unmeasured.
- The package-wide scan is syntactic: a shared global written indirectly (through a helper or an
  alias) would not have been matched by either scan.
- The completed-run warning total is unknown, because no completed 20-iteration RED run exists.

**Residual-risk** — the option choice is deliberately open, so the plan cannot pre-verify the
diff shape the repair will produce; AC-WFR-003's fence is stated but unexercised until run-phase.

## §E.2 Run-phase Evidence

**Resumed by a different lane at HEAD `6a56c96bd`** (the original lane left no report; plan-phase
artifacts re-read directly, `plan-audit-iter1.md` + `red-race-d592b0551.txt` both present). Local
`develop` at `4e91bf6a9` was absorbed first per the lead's dispatch (merge commit `6a56c96bd`,
conflict-free; one transient `fatal: Unable to write index.` on the first attempt — no path in the
message, fsmonitor unset, lock count 0 → lead's cause-B disposition, one retry, succeeded).

### AC-WFR-001a — fresh RED on the moved head (recorded before any code edit)

- Command: `go test ./internal/cli/ -run TestResolveWorktreeExistingBranch -count=20 -race`
- Tree: `6a56c96bd` (absorbed tip), this run
- Output: `.moai/reports/t464/red-race-6a56c96bd.txt`, 1757 lines
- Exit code: `1` (background task bbusrtk32 exit status, read directly)
- `WARNING: DATA RACE` count: **30** (`grep -c`, count read as printed value)
- `Log in goroutine after` count: **0** — this run panicked nowhere and **completed** all
  20 iterations, unlike the `d592b0551` capture which truncated at a panic (spec §A.2).
  This upgrades the RED: the panic manifestation did not reproduce this run, the warning
  manifestation did, at 30 > 23, on a completed run. RED satisfied (AC-WFR-001a: exit 1
  AND ≥1 warning or the panic — the warnings limb holds).

### AC-WFR-006 — option choice and its criteria (written before the first code edit)

**Choice: Option A — remove `t.Parallel()` from the four siblings (lines 66/96/126/161).**

Evaluation against `spec.md` §D.3, criterion by criterion:

| Criterion | Verdict | Ground |
|---|---|---|
| Blast radius | **A** | A is 4 lines in 1 file — the REQ-WFR-005 fence holds mechanically. B changes a production signature and its fence-compliance would depend on a consumer sweep of `resolveWorktreeExistingBranch` before a single line could be written. |
| What is lost | **A** | The SPEC itself states these four tests are sub-millisecond stubs whose parallelism "buys no measurable wall-time" (`spec.md` §D.3) — Option A loses nothing measurable. |
| Recurrence | **B** (only B wins this one) | A re-adds the defect if a later author re-adds `t.Parallel()`. But the standing regression guard (REQ-WFR-006, the `-race` repetition command) exists under either option, and B's structural fix is itself partial (spec §F4: the default-to-global fallback preserves the seam; the package-wide injection conversion is a separate card per §C). |
| Consistency with the package | **A** | The package's dominant idiom is global-seam assignment (23 files); B introduces a second idiom into one file (`spec.md` §D.3). |
| Reviewability | **A** | A four-line deletion vs a signature change plus four test rewrites. |

Four of five criteria favour A; the one criterion favouring B is mitigated by the standing
regression guard and is a partial fix even under B. This is not "A because it is smaller" —
the deciding grounds are the SPEC's own statement that the lost parallelism is unmeasurable
plus the fence holding mechanically without a consumer sweep.

### Repair

Applied after the entries above were written: the four `t.Parallel()` calls at lines 66/96/126/161
of `internal/cli/worktree_branch_flag_test.go` removed; no other change. The `t.Parallel()` calls
at lines 25 and 44 belong to other tests in the file and are untouched.

Repair commit: `76f2165a1` (carries the repair, the fresh RED capture, and this section's
pre-edit entries).

### AC-WFR-001b — GREEN under the same repetition shape

- Command (identical shape to AC-WFR-001a): `go test ./internal/cli/ -run TestResolveWorktreeExistingBranch -count=20 -race`
- Tree: working tree of repair commit `76f2165a1` (measured before the commit; the commit's tree content is identical)
- Output: `.moai/reports/t464/green-race-76f2165a1.txt` (renamed from the `-pending` filename once the repair SHA existed), 1 line
- Exit code: `0` (background task boucgwt21 exit status, read directly)
- `WARNING: DATA RACE` count: **0** (printed count read, not the grep exit status)
- `Log in goroutine after` count: **0**
- Final line: `ok  	github.com/modu-ai/moai-adk/internal/cli	2.464s`
- The GREEN ran the full 20 iterations where the original RED truncated at the panic — the
  asymmetry runs in the safe direction (`spec.md` §D.4).

### AC-WFR-002 — the four siblings survive intact

- `grep -c 'func TestResolveWorktreeExistingBranch_' internal/cli/worktree_branch_flag_test.go` → **4**
- Remaining `t.Parallel()` occurrences in the file: lines 25 and 44 only (other tests, untouched)
- `git diff` on the file: exactly 4 deleted lines (`-	t.Parallel()`), zero changed assertions.

### AC-WFR-003 — the diff stays inside the fence

- Fence base: `develop` (the absorbed local develop) — the AC's `d592b0551` base predates the
  absorb; `develop..HEAD` is the same logical measurement on the moved base.
- `git diff develop --stat -- internal/cli/` → `internal/cli/worktree_branch_flag_test.go | 4 ----` — exactly one file, under Option A's fence (no `worktree_branch_flag.go` change).
- Positive control (mandatory): `git diff develop --stat` (no pathspec) → 10 files, 4011 insertions, 101 deletions — non-empty, so the fence check measured something.

### AC-WFR-005 — cross-platform compile (SHOULD)

- `GOOS=linux GOARCH=amd64 go vet ./internal/cli/...` → rc=0
- `GOOS=windows GOARCH=amd64 go vet ./internal/cli/...` → rc=0

### AC-WFR-004 — package non-regression: measured twice; one pre-existing failure carries over

1. **Run 1** (`go test ./internal/cli/... -race -timeout 600s`, per the AC's literal shape):
   `WARNING: DATA RACE` count **0**, but the root `internal/cli` package panicked at
   `test timed out after 10m0s` (601.790s) with `TestInstallVersionTag_NetworkFailureDownload`
   in the running set, plus `TestBinaryLag_DoctorCheckNameSetIsUnchanged` FAIL. Evidence:
   `.moai/reports/t464/package-race-76f2165a1.txt`.
2. **Isolation probe**: the network test alone under `-race -count=1 -timeout 120s` exits **0**
   (`.moai/reports/t464/hang-repro-76f2165a1.txt`) — it is not a hang; the 600s floor is simply
   insufficient for this package under `-race` on this machine (the non-race package alone was
   measured at ~679s in a prior card).
3. **Run 2** (`-race -timeout 1800s`, exceeding the §D.4 floor): **0 race warnings, no timeout**,
   16/16 subpackages `ok`; the root package FAILs on exactly one test,
   `TestBinaryLag_DoctorCheckNameSetIsUnchanged` (binary_lag_test.go:205: *"this SPEC added
   doctor check name Hook Delivery; REQ-BLV-009 rewires the existing Binary Freshness item and
   registers no new name"*). Evidence: `.moai/reports/t464/package-race-1800s-76f2165a1.txt`.

**Attribution of the AC-WFR-004 failure — measured, not assumed.** The failing test compares the
doctor-check name set against a baseline allowlist (`namesAddedAfterBaseline`,
binary_lag_test.go:184) that lists only `hookWiringCheckName` (t216). The name it rejects,
`Hook Delivery`, is the check added by card t466 — which this branch absorbed from develop.
Measured: `git show develop:internal/cli/doctor.go | grep -c 'Hook Delivery'` → **1** (develop
carries it); `git diff develop --stat -- internal/cli/binary_lag_test.go internal/cli/doctor.go`
→ **empty** (this branch touches neither file). The failure is **pre-existing on develop**
(t466 landed without updating the binary-lag guard allowlist) and is unreachable by this card's
diff: fixing it here would violate the REQ-WFR-005 fence. Adjudicated by the lead 2026-09-04 —
see the carried-debt record below and §E.3.

**F3 resolved.** `spec.md` §F3 recorded that the RED never completed a full 20-iteration run
(it panicked). The fresh RED at `6a56c96bd` completed **all 20 iterations** — 30 warnings,
no panic — so the measured shape's behaviour across a full run is now observed, and the
GREEN ran the same shape to completion as well.

### Carried Debt — AC-WFR-004's binary-lag limb (lead adjudication (b), 2026-09-04)

**What follows is not a record of something this card chose not to fix; it is a measurement
that the failing test is outside this card's reach.** AC-WFR-004 is judged PASS on its race
axis (0 `WARNING: DATA RACE` across both package `-race` runs, 16/16 subpackages ok); the
binary-lag limb is carried as a debt with its attribution pinned below, owned by card **t466**
(SPEC-UPDATE-HOOK-DELIVERY-001), routed to the operator by the lead under that card's
ownership.

1. **The failing test and its reason, one line.** `TestBinaryLag_DoctorCheckNameSetIsUnchanged`
   (internal/cli/binary_lag_test.go:205): *"this SPEC added doctor check name Hook Delivery;
   REQ-BLV-009 rewires the existing Binary Freshness item and registers no new name"* — the
   guard rejects `"Hook Delivery"`, the doctor check name card t466 added.
2. **The attribution evidence, two lines** (measured by the lead in the develop worktree at
   `624bb4c55`, independently of this lane):
   - `git show HEAD:internal/cli/doctor.go | grep -c 'Hook Delivery'` → **1** — develop
     carries the check t466 added;
   - `git grep -A12 'namesAddedAfterBaseline = ' -- 'internal/cli/*.go'` → the allowlist
     carries exactly one entry, `"hookWiringCheckName"` (t216) — t466's check name was never
     registered.
   The two together are the pre-existing defect: develop added a check and the guard's
   allowlist did not follow.
3. **The unreachable-by-this-diff ruling and its reason.** `git diff develop --stat --
   internal/cli/binary_lag_test.go internal/cli/doctor.go` → **empty**: this card's diff
   touches neither guard file. The only repair from here — adding `"Hook Delivery"` to
   `namesAddedAfterBaseline` — would put a new file inside this card's `internal/cli/` diff
   and violate the REQ-WFR-005 fence, i.e. it would repair another card's work inside a card
   forbidden from reaching it. The repair therefore rides t466's ownership, not this card.

## §E.3 Run-phase Audit-Ready Signal

**Claim** — the race repair is complete and proven (RED→GREEN on the identical `-race -count=20`
command shape, fence held, siblings intact, cross-platform compiles clean). One MUST criterion
(AC-WFR-004) exits non-zero solely on a **pre-existing develop defect owned by another card**;
the lane cannot repair it without violating REQ-WFR-005.

| AC | Status | Basis |
|---|---|---|
| AC-WFR-001a | **PASS** | Fresh RED at `6a56c96bd`: exit 1, `WARNING: DATA RACE` = 30 on a completed 20-iteration run (`red-race-6a56c96bd.txt`) — stronger than the cited truncated baseline. |
| AC-WFR-001b | **PASS** | GREEN on the repaired tree: exit 0, warnings 0, panic 0 (`green-race-76f2165a1.txt`, full 20 iterations). |
| AC-WFR-002 | **PASS** | 4 siblings present, assertions untouched; only the 4 `t.Parallel()` lines removed. |
| AC-WFR-003 | **PASS** | Fence: one file under `internal/cli/` (`worktree_branch_flag_test.go`, 4 deletions); positive control non-empty. |
| AC-WFR-004 | **BLOCKED-PRE-EXISTING** | 1800s `-race` rerun: 0 race warnings, no timeout, 16/16 subpackages ok; root package fails only `TestBinaryLag_DoctorCheckNameSetIsUnchanged` — measured pre-existing on develop (t466's "Hook Delivery" check never added to the binary-lag allowlist). Attribution evidence in §E.2. Not reachable by this card's diff (REQ-WFR-005 fence). |
| AC-WFR-005 | **PASS** (SHOULD) | `GOOS=linux/windows go vet` both rc=0. |
| AC-WFR-006 | **PASS** | Option A + criteria recorded above, written before the first edit. |

**Blocker adjudicated by the lead (2026-09-04): option (b).** AC-WFR-004 is judged on the race
axis — the package-wide `-race` runs show **0 `WARNING: DATA RACE`** — so AC-WFR-004 records
**PASS on its own axis**, with the binary-lag failure carried as a debt whose attribution is
measured below (see §E.2 Carried Debt). The lead independently re-measured the attribution in
the develop worktree at `624bb4c55` (not by citing this lane's report): `Hook Delivery` present
in develop's doctor.go (count 1), allowlist carries exactly one entry (`"hookWiringCheckName"`).
The follow-up-card path (option (a)) is not rejected — it is not this lane's act; the lead
routes it to the operator separately under t466's ownership.

**Standing regression evidence** (REQ-WFR-006): `go test ./internal/cli/ -run TestResolveWorktreeExistingBranch -count=20 -race` — RED at `6a56c96bd` (exit 1, 30 warnings) → GREEN at `76f2165a1` (exit 0, 0 warnings). Reproduce with either persisted capture in `.moai/reports/t464/`.

**Not measured**: linux/amd64 behaviour of the race itself (`spec.md` §F2 — vet-only, unchanged);
CI's full-suite verdict on the pushed head (lead's batch push, repository rule).

## §E.4 Sync-phase Audit-Ready Signal

sync_commit_sha: "a992556fd"

Sync-phase activities, on the single sync commit: the `[Unreleased]` CHANGELOG entry for the
race repair (duplicate-checked — `grep -c 'SPEC-CLI-WORKTREE-FLAG-RACE-001' CHANGELOG.md` → 0
before the entry was added, per the B12 emission discipline); the lead's adjudication (option
(b)) recorded into §E.2 as a measured carried debt and into §E.3 as the resolved state of the
AC matrix; and the `in-progress → completed` frontmatter transition riding this commit per the
3-phase close. The `sync_commit_sha` value above is the canonical placeholder — a commit cannot
cite its own hash — and is backfilled in a follow-up commit.

# research.md — SPEC-FACTORY-RUN-RETIRE-001 (card t1107)

## §A What this file is

The codebase research the design rests on: what was measured, where, what it showed, and — kept
deliberately visible — what was **not** measured. `design.md` cites this file by ledger id
(`M-nn`); `spec.md` §A and §A.1 cite the same facts in prose. This is where the commands live.

**Method.** Every entry below is a read-only command executed in **this worktree**
(`.claude/worktrees/t1107`) at commit **`157ec351b`** on branch `WT-factory-run-retire`, on
**darwin 27.0.0**, on **2026-09-23** — except the two entries explicitly attributed to another
actor (**M-12** and **M-17**), each of which carries its source inline and says that I did not
re-run it. An entry with no attribution block is my own measurement in this tree.

**Boundary.** Nothing here was measured by running the product. The one execution-based artefact in
this card's evidence base is the reproduction, `.moai/reports/t1107/verdict.md`, which is an input
to this SPEC rather than a part of it. Where a claim needs execution, §J says so rather than
letting a source read stand in for it.

---

## §B Ledger — the run record and its writers

### M-02 · No code path takes a run out of `active`

```
$ grep -rn "UPDATE runs\|DELETE FROM runs" --include='*.go' . | grep -v _test.go
(no output)
exit=1
```

This is the defect, stated as an absence. An absence measured by a search is only as good as the
search, so:

### M-03 · Control — the same search shape finds what does exist

```
$ grep -rn "status='active'" --include='*.go' . | grep -v _test.go
internal/factorymsg/store.go:257:	rows, err := db.DB.QueryContext(ctx, `SELECT run_id FROM runs WHERE status='active' ORDER BY run_id`)
internal/homestate/runtime.go:25:lead_session_id=excluded.lead_session_id,lead_backend=excluded.lead_backend,status='active',manifest_json=excluded.manifest_json,updated_at=excluded.updated_at`,
exit=0
```

Two hits. The instrument works, so M-02's zero is an absence in the code and not a failure of the
grep. The pair also names the whole `runs.status` surface: **one writer** (`runtime.go:25`, which
only ever writes `'active'`) and **one reader** (`store.go:257`).

### M-04 · The `runs` row carries no owner identity

```
$ sed -n '33,41p' internal/homestate/factory.go
CREATE TABLE IF NOT EXISTS runs (
  run_id TEXT PRIMARY KEY,
  lead_session_id TEXT NOT NULL DEFAULT '',
  lead_backend TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL,
  manifest_json TEXT NOT NULL DEFAULT '{}',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);
```

No pid, no process fingerprint. `lead_session_id` is a session uuid, not a process handle — nothing
in this row can be probed for liveness. This is why the design adds a column rather than deriving
an answer from what is already there.

### M-05 · Current schema version

```
$ grep -n "^const factorySchemaVersion" internal/homestate/factory.go
20:const factorySchemaVersion = 2
```

Version 2, with `migrateFactoryV1ToV2` at `internal/homestate/factory.go:211` as the migration shape
to follow (`ALTER TABLE ... ADD COLUMN` inside one transaction, then the `meta.schema_version`
update).

### M-08 · A process identity IS already recorded per run — in the broker

```
$ grep -n "CREATE TABLE IF NOT EXISTS peers" internal/factorymsg/store.go
291:CREATE TABLE IF NOT EXISTS peers(slot TEXT PRIMARY KEY, project_key TEXT NOT NULL,
    run_id TEXT NOT NULL, backend TEXT NOT NULL, role TEXT NOT NULL, session_uuid TEXT NOT NULL
    UNIQUE, generation INTEGER NOT NULL, pid INTEGER NOT NULL, process_start TEXT NOT NULL,
    updated_at TEXT NOT NULL);
```

(line wrapped for reading; it is one line in the source)

`pid` **and** `process_start`, per peer, per run — the t1074 PID-plus-start-identity convention,
already in service. This is what makes REQ-006's legacy-row fallback possible at all: rows written
before this SPEC lands have no stamp, but their lead peer has an identity.

> **Note added 2026-09-25 (v0.13.1, card t1169).** The sentence above holds for a legacy row whose
> lead registered a peer; card t1168 later measured legacy rows with no broker file at all, for
> which neither source yields an identity. Such a row is not always `indeterminate`: it is `dead`
> when the REQ-006b boot proof holds and `indeterminate` otherwise (`spec.md` REQ-006 / REQ-006b).
> The measurement above is left as recorded.

---

## §C Ledger — the five call sites, in three launch shapes

### M-01 · Enumeration

```
$ grep -rn "registerFactoryLaunchPending(" --include='*.go' internal/cli/ \
    | grep -v _test.go | grep -v "^internal/cli/factory_launch_pending.go"
internal/cli/launch_exec_windows.go:54:	if _, err := registerFactoryLaunchPending(context.Background(), launchProjectRoot(), child.Env, child.Process.Pid, childFingerprint); err != nil {
internal/cli/launch_exec_posix.go:33:	pending, err := registerFactoryLaunchPending(context.Background(), root, launchEnv, os.Getpid(), homestate.CurrentProcessFingerprint())
internal/cli/codex_launcher.go:230:			_, identityErr = registerFactoryLaunchPending(context.Background(), dir, env, pid, start)
internal/cli/codex_direct_windows.go:24:	if _, err := registerFactoryLaunchPending(context.Background(), cmd.Dir, cmd.Env, cmd.Process.Pid, start); err != nil {
internal/cli/codex_direct_posix.go:34:	pending, err := registerFactoryLaunchPending(context.Background(), cmd.Dir, cmd.Env, pid, start)
exit=0
```

**Five** sites. The argument list of each is the finding: every one of them already resolves a
`(pid, fingerprint)` pair, so the restamp the design needs requires no new probe anywhere — only a
call.

The shape classification (`spec.md` §A.1) follows from what each site's surrounding code does
afterwards, read at the same tree:

- `launch_exec_posix.go:33` → `syscall.Exec(claudeBin, args, launchEnv)` immediately after
  (`launch_exec_posix.go:37`). **replace.**
- `codex_direct_posix.go:34` → same shape. **replace.**
- `launch_exec_windows.go:54` → registers `child.Process.Pid` with a fingerprint probed live at
  `launch_exec_windows.go:49`, then blocks in `child.Wait()` (`:63`). **spawn.**
- `codex_direct_windows.go:24` → registers the child, then `cmd.Wait()`. **spawn.**
- `codex_launcher.go:230` → registers the **tmux pane's** pid, resolved by
  `defaultCodexSpawnPaneIdentity`, then prints and returns (`:237-240`). The launcher exits; the
  pane lives on. **pane.**

### M-09 · Build tags — the axis is door shape, not platform

```
internal/cli/codex_launcher.go     -> (no build tag)
internal/cli/launch_exec_windows.go -> //go:build windows
internal/cli/launch_exec_posix.go   -> //go:build !windows
```

**The finding that changed the design.** `codex_launcher.go` carries no build tag, so the **pane**
shape — the one whose launcher exits immediately — compiles and runs on **darwin**. Two of the three
shapes are therefore reachable on the only platform this card verifies before merge, and a design
branching on `GOOS` would have missed the pane door entirely while looking correct.

### M-13 · The pane door records the run *before* the pane exists

```
$ grep -n "recordFactoryRunStart\|return runCodexLaunch" internal/cli/codex_launcher.go
515:			if err := recordFactoryRunStart(launchProjectRoot(), os.Getenv(config.EnvMoaiKanbanID), codexFactoryBackend, ""); err != nil {
538:		return runCodexLaunch(cmd, kind, tail, spawn, worktree)
```

Line 515 precedes line 538 in the same function body. The run row is written while the only process
in existence is the launcher; the pane that will actually run the session is opened downstream of
`runCodexLaunch`. This is the ordering that makes a record-time stamp on this door name a process
already gone by the time anyone reads it (`design.md` §C.4).

### M-18 · The pane identity the restamp would use is already resolved, live

`defaultCodexSpawnPaneIdentity` (`internal/cli/codex_launcher.go:242-257`) polls `tmuxPanePID` until
`homestate.ProbeProcessIdentity` reports the pane process `ProcessIdentityLive` **with a non-empty
fingerprint**, to a 2-second deadline, and returns that `(pid, start)` pair. On failure,
`defaultCodexSpawnLaunch` calls `codexSpawnCleanupPaneFn(paneID)` at `codex_launcher.go:233` and
returns the joined error.

Two consequences for the design: the restamp needs no new probing machinery on this door, and
REQ-002d's refusal is an extension of a refusal path the code already has rather than a new one.

---

## §D Ledger — process identity and its platform resolution

### M-11 · The liveness primitive already returns the three-value lattice

```
$ sed -n '202,212p' internal/homestate/profile_lease.go
func ProbeProcessIdentity(pid int) (string, ProcessIdentityState) {
	state := platformPIDState(pid)
	if state != ProcessIdentityLive {
		return "", state
	}
	fingerprint, ok := platformProcessFingerprint(pid)
	if !ok {
		return "", ProcessIdentityIndeterminate
	}
	return fingerprint, ProcessIdentityLive
}
```

`live` / `dead` / `indeterminate` is not invented by this SPEC — it is the existing contract, and it
already degrades an unreadable fingerprint to `indeterminate` rather than guessing. The classifier
composes with this rather than replacing it.

### M-14 · Fingerprint resolution differs by platform — and linux is the coarse one

| Platform | File (build tag) | Source | Resolution |
|---|---|---|---|
| darwin | `process_fingerprint_darwin.go` (`darwin`) | `unix.SysctlKinfoProc` → `P_starttime`, formatted `%d.%06d` | **microsecond** |
| windows | `process_fingerprint_windows.go` (`windows`) | `GetProcessTimes` → `created.Nanoseconds()` | **nanosecond** |
| linux / other unix | `process_fingerprint_unix.go` (`!windows && !darwin`) | `exec.Command("ps", "-o", "lstart=", "-p", …)` | **one second** |

```
$ head -1 internal/homestate/process_fingerprint_unix.go
//go:build !windows && !darwin
$ grep -n 'exec.Command("ps"' internal/homestate/process_fingerprint_unix.go
12:	out, err := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)).Output()
```

`ps -o lstart=` prints a whole-second timestamp. So on linux a pid reused **within the same second**
as its predecessor's start is genuinely indistinguishable from it — a platform fact, not an
implementation shortcut, and the reason REQ-003b exists and names linux specifically. The narrow
band resolves toward `live`, so the residual failure is a surviving stale run (escapable with
`--factory-run`) rather than a retired live one.

The table also shows why AC-003 supplies its fingerprints as **fixture values**: racing a real
same-second pid reuse cannot be made to work on linux by construction.

### M-12 · `execve` preserves pid and start time — darwin, measured, **not my measurement**

> **Attribution.** This is the plan-auditor's measurement, recorded at
> `.moai/reports/t1107/plan-audit-iter2.md` (defect D18), taken on darwin 27.0.0 on 2026-09-23 with
> a probe copying `platformProcessFingerprint` from `process_fingerprint_darwin.go`. I did not
> re-run it. It is carried here because the design depends on it and because the auditor raised it
> precisely to stop the claim being asserted as OS folklore.

```
--- before exec (process is /bin/sh) ---
97977 /bin/sh
pid=97977 fingerprint="1790136707.922810" ok=true
--- after exec (process should be sleep, same PID) ---
97977 sleep
pid=97977 fingerprint="1790136707.922810" ok=true
```

(`/bin/sh -c 'echo $$ > pidfile; sleep 3; exec sleep 30'`)

Both pid and fingerprint survive the `execve`. This is what makes the **replace** shape need no
restamp: the launcher's record-time stamp already names the session process, because it becomes the
session process.

**Not measured on linux or windows.** The same result is expected on linux — `ps -o lstart=` reads
task creation time, which `execve` does not reset — but that is a prediction, not an observation,
and it is carried as such (§J).

---

## §E Ledger — package boundaries

### M-06 · `homestate` does not import `factorymsg`

```
$ grep -rln "internal/factorymsg" --include='*.go' internal/homestate/
(no output)
exit=1
```

### M-07 · `factorymsg` does import `homestate`

```
$ go list -deps ./internal/factorymsg | grep -c 'internal/homestate$'
1
```

and the import itself:

```
$ grep -rn "internal/homestate" --include='*.go' internal/factorymsg/ | grep -v _test.go
internal/factorymsg/store.go:21:	"github.com/modu-ai/moai-adk/internal/homestate"
```

The dependency direction is fixed and one-way. Together with M-04 and M-08 — the `runs` table in
`homestate`, the `peers` table in `factorymsg` — this is what forces REQ-006's fallback to cross the
boundary as a **function value** rather than an import (`design.md` §F). It is a compiler-enforced
constraint, not a style preference: the alternative does not build.

### M-16 · `ResolveActiveRun`'s current shape

```
$ sed -n '241,277p' internal/factorymsg/store.go     # excerpt
	if explicit != "" { … return explicit, nil }
	rows, err := db.DB.QueryContext(ctx, `SELECT run_id FROM runs WHERE status='active' ORDER BY run_id`)
	…
	switch len(runs) {
	case 0:  return "", errors.New("NO_ACTIVE_FACTORY")
	case 1:  return runs[0], nil
	default: return "", errors.New("AMBIGUOUS_FACTORY")
	}
```

Three facts the design leans on: the explicit-`--factory-run` branch returns before the active-set
query (so the operator escape is untouched), the sentinels are bare `errors.New` strings that
downstream tests match as substrings, and the 0/1/many switch is the entire decision — there is no
selection logic to preserve or displace.

---

## §F Ledger — the three-OS verification path and its timing

### M-19 · CI triggers

```
$ sed -n '16,20p' .github/workflows/ci.yml
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]  # main으로 향하는 모든 PR에서 CI 실행
```

### M-20 · The `test-integration` job

```
$ sed -n '373,381p' .github/workflows/ci.yml
  test-integration:
    name: Integration Tests (${{ matrix.os }})
    runs-on: ${{ matrix.os }}
    needs: test
    timeout-minutes: 30
    strategy:
      fail-fast: false
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
```

No `if:` of its own — it inherits its gate through `needs: test`, and `test` is gated on
`needs.detect.outputs.go_code == 'true'` (`ci.yml:120`) whose path filter includes `'**/*.go'`
(`ci.yml:79`).

### M-21 · What that job actually runs

```
$ sed -n '401,407p' .github/workflows/ci.yml
      - name: Run harness integration tests
        shell: bash
        run: |
          rc=0
          go test -json -tags=integration -race -timeout 180s ./test/integration/harness/... > test-stream.json || rc=$?
```

**One path, one build tag.** `./test/integration/harness/...` behind `-tags=integration` is the
whole of what the three-OS matrix executes. A cross-platform exercise placed anywhere else is not
cross-platform, whatever its name says.

### M-22 · Existing occupants of that path

```
$ /bin/ls -1 test/integration/harness/
it01_replay_test.go
it02_tier3_test.go
it03_tier4_test.go
it04_frozen_test.go
it05_rate_test.go
it06_rollback_test.go
it07_disable_test.go
```

Each carries `//go:build integration` as its first line. `it08_` is the next slot — the naming the
plan adopts.

### M-17 · Confirming run — **not my measurement**

> **Attribution.** `gh run view 35802361895`, read while authoring the earlier revision of
> `spec.md` §A.2 (recorded there, 2026-09-23). I did not re-read it in this pass. Reported: `event:
> push`, `branch: develop`, head `1dbe5e2f33147b8ae58dfb76f5bde3a43ba4b47f`; jobs `Integration Tests
> (ubuntu-latest)`, `(macos-latest)`, `(windows-latest)` each `success`, inside a run whose overall
> conclusion was `failure` on unrelated jobs.

The last clause is the useful part: the three-OS jobs execute and report on their own terms, so
their verdict is readable without waiting for an all-green run.

### The timing this produces

M-19 has no `pull_request` trigger for `develop`, and this project does not push `WT-` branches
(CLAUDE.local.md §4.1, lane obligations). Therefore **a card branch receives no CI run before it
merges**:

| Platform | When | Source |
|---|---|---|
| darwin | pre-merge | the lane's own local run |
| ubuntu | post-merge | job `test` on the develop push |
| ubuntu + macos + windows | post-merge | job `test-integration` on the same push |

This is why AC-013 is split into a release-blocking pre-merge leg and a non-gating post-merge leg. A
criterion asserting the three-OS result as a merge gate would be false about its own timing.

---

## §G Ledger — why test isolation is a requirement rather than hygiene

### M-15 · `CanonicalProjectRoot` converges a linked worktree onto the primary checkout

```
$ sed -n '48,62p' internal/homestate/paths.go      # excerpt
	if dirs, err := gitcore.ResolveGitDirs(projectRoot); err == nil && dirs.CommonDir != "" {
		if dirs.GitDir == dirs.CommonDir { … } else if root, ok := primaryCheckoutRootFromCommonDir(dirs.CommonDir); ok {
			projectRoot = root
		} else if out, err := gitcore.ExecCommand("git", "-C", projectRoot, "worktree", "list", "--porcelain").Output(); err == nil {
			// Git lists the primary checkout first, even with external metadata.
```

In a linked worktree `GitDir != CommonDir`, so the second branch fires and the project root becomes
**the primary checkout**. A factory fixture rooted anywhere inside this repository's worktree set
therefore resolves to the developer's real project key and writes to the developer's real
`factory.db`.

"Use a temp dir" does not cover it — the temp dir must also not be inside a linked worktree. This is
the mechanism behind REQ-012 and `acceptance.md` §B, and the reproduction already relied on it:
`verdict.md` §2.1 records that its sandbox produced project key `proj-9c008a86`, which is not this
repository's.

---

## §H What the reproduction established, and what this SPEC adds to it

`.moai/reports/t1107/verdict.md` is an input, not a finding of this file. Taken as read, it
establishes by execution (darwin, stub `claude`, sandboxed `HOME`/`MOAI_HOME`/`MOAI_CLAUDE_BIN`):

- two lead launches in **different seconds** leave two `active` rows, and every subsequent worker
  join fails `AMBIGUOUS_FACTORY` (R1a/R1b/R1c, exit 1);
- the old `-f lane-N` alias behaves identically (R2);
- `--factory-run <id>` works as an escape (R3, exit 0);
- two launches in the **same** second collapse to one row through `ON CONFLICT(run_id) DO UPDATE`,
  so the trigger is specifically a different-second relaunch (C3).

What the research above adds to that: the *mechanism* of the absence (M-02/M-03), the reason no
answer can be derived from existing state (M-04), the reason one nevertheless exists nearby (M-08),
and the three-shape structure (M-01/M-09/M-13) that makes the naive fix retire live runs.

---

## §I Research that changed the design

Recorded because a research file whose findings all confirm the first draft is usually a research
file that was written afterwards.

| Finding | What it overturned |
|---|---|
| **M-09** — `codex_launcher.go` has no build tag | The premise that non-replace launch shapes are a Windows concern. The pane door is on darwin; a `GOOS` branch would have missed it (defect D11). |
| **M-01** — five call sites, not two | A draft discussing two doors while five exist. Two of the five are discharged by shape match, but only after being named. |
| **M-13** — record at `:515`, pane at `:538` | The assumption that a record-time stamp can be made correct on every door. On the pane door it is false within milliseconds of being written. |
| **M-06/M-07** — one-way package dependency | Putting REQ-006's peer fallback inside `homestate`. It does not compile; the fallback became a function parameter. |
| **M-14** — linux fingerprint is one-second | The assumption that pid+start always discriminates. It does not on linux, which is what REQ-003b now states. |
| **M-19/M-20** — no `pull_request` trigger for `develop` | An AC-013 worded as though the three-OS verdict gates this card's merge. It cannot; the criterion was split into two legs. |

---

## §J What was NOT measured

Stated so that no absence below reads as a finding.

- **Nothing was executed.** Every entry above except M-12 and M-17 is a static read of source or
  configuration at `157ec351b`. No test was run, no binary built, no factory state opened. The
  implementation does not exist yet, so there is nothing behavioural to measure — the RED-now ledger
  in `acceptance.md` §C.1 is the deliberate exception, and it too is grep-based.
- **The `--spawn` pane door was never launched.** M-13 and M-18 are readings of call order and of
  `defaultCodexSpawnPaneIdentity`'s contract. No tmux pane was opened, no run row compared against a
  pane peer. AC-011 makes this an executed observation in the run phase; `plan.md` §M5 requires a
  **blocker report** rather than a silent downgrade if the sandbox cannot provide a tmux server.
- **Nothing was measured on linux or windows.** M-12's `execve` result is darwin-only. The
  windows-side spawn shape is read from source (`launch_exec_windows.go`); no windows host was
  exercised. The expectation that linux preserves the fingerprint across `execve` is a prediction.
- **The `runs`-status absence (M-02) was measured by search, not by execution.** M-03 shows the
  instrument works, which is the strongest thing a search can establish. A code path that mutates
  `runs.status` without the literal text `UPDATE runs` — via a query builder, or string
  construction — would not appear. No such builder exists in these packages, but that is itself a
  reading rather than a proof.
- **Sibling-card overlap was not measured in this tree, and cannot be.**
  `internal/cli/factory_lane_handoff*.go`, `internal/factorymsg/handoff*.go`,
  `internal/hook/factory_messages*.go` do not exist here — they live on t1082's and t1109's
  branches. An absence here means *not measured*, never *does not overlap*. The zero-overlap figure
  is the team lead's measurement against t1082's own worktree (2026-09-23), cited as that. The live
  residual is same-package rather than same-file: this SPEC edits
  `internal/factorymsg/store.go` while t1082 adds `handoff*.go` to that package.
- **The plan-audit iter-3 report IS on disk, and nothing above is attributed to it.**
  `.moai/reports/t1107/plan-audit-iter3.md` exists — 38,933 bytes, timestamped 2026-09-23 13:52 —
  and its first content line carries the required header verbatim: `상한 초과 사유: 운영자 범위
  확장(pane door) 뒤의 재감사`. An earlier revision of this item asserted the file was absent and,
  being under a gitignored path, unrecoverable from history. That assertion was false. It was a
  timing artifact, not a bad read: the directory was checked roughly two minutes before the auditor
  finished writing the report into it, so the absence was real when observed and already stale when
  written down. What remains true is the attribution boundary — the findings reaching this file come
  from `spec.md` §H's revision record and the D11-D17 defect text in the iter-2 report, so nothing
  above rests on the iter-3 document.
- **Accumulated scale in real installations was never measured.** No real `factory.db` was opened.
  The design does not need the figure — reconciliation is per-resolution, not a sweep sized to a
  population — but the number is unknown, not small.

---

## §K Cross-references

- `.moai/reports/t1107/verdict.md` — the reproduction (§H), an input to this SPEC
- `.moai/reports/t1107/plan-audit-iter2.md` — source of M-12, and of the D11/D14/D15 findings §I records
- `spec.md` §A / §A.1 / §A.2 — the same facts in prose, at the requirements layer
- `design.md` — consumes this ledger by `M-nn` id
- `acceptance.md` §B / §C.1 — the isolation constraint M-15 grounds, and the RED-now ledger

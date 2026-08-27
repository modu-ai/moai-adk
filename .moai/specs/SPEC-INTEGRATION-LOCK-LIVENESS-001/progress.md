# SPEC-INTEGRATION-LOCK-LIVENESS-001 — progress (card t298)

Plan-phase artifacts authored 2026-08-27 on tree `d29b8942e` @ `WT-integration-lock`
(worktree t298). Tier M. Status: draft.

## §E.1 Plan-phase Audit-Ready Signal

- Artifact set complete for Tier M: spec.md, plan.md, acceptance.md, progress.md (this file).
- SPEC ID regex check — command run in the iter-1 repair pass, verbatim:
  `ID="SPEC-INTEGRATION-LOCK-LIVENESS-001"; [[ "$ID" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]] && echo PASS || echo FAIL`
  → observed output `PASS`. The pattern is the multi-segment form (`^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$`),
  which is what admits this ID's multi-hyphen domain. The single-segment form quoted in
  `spec-frontmatter-schema.md` § Field Reference (`^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$`) rejects it;
  multi-segment IDs are an established repo-wide pattern and `moai spec lint` emits no ID finding
  for this SPEC, so the ID is not the defect — the earlier unattributed `PASS` claim was.
- Frontmatter: 12 canonical fields present; `status: draft` set at creation; `phase` carries a
  release target ("v3.1.4 target"), not a lifecycle stage.
- ID uniqueness: no SPEC under this worktree's `.moai/specs/` shares the ID.
- Requirements: REQ-INL-001..010 in GEARS notation (no IF/THEN modality).
- Out of Scope section present with four `### Out of Scope — <topic>` H3 sub-headings, each
  with `-` bullets.
- Acceptance: AC-INL-001..011, two-cell discipline (RED-now pinned to `d29b8942e` + green path
  naming its flipping milestone); M1 is the failing-test-first milestone.
- Premises line-pinned to `d29b8942e` (spec.md §B, 13 items) — re-verify line numbers before
  citing from any other tree.

### iter-1 audit repair (v0.1.1)

plan-auditor iter-1 returned FAIL 0.795 (Tier M threshold 0.80; report
`.moai/reports/t298/plan-audit-iter1.md`). Five blocking defects (D1-D5) plus four optional
(D6-D9) were repaired in this pass. Every claim below is measured in THIS run.

- **Baseline still valid despite HEAD advancing.** The artifacts pin `d29b8942e`; HEAD is now
  `c67a6ea64` (`git rev-parse --short HEAD`). `git merge-base --is-ancestor d29b8942e HEAD`
  → exit 0 (`d29b8942e` is an ancestor), and
  `git diff --stat d29b8942e HEAD -- internal/kanban/integration_lock.go internal/cli/integration.go internal/hook/integration_lock_guard.go internal/session/session_pid.go internal/session/registry.go CLAUDE.local.md .claude/rules/local/gitflow-lane-protocol.md`
  → **empty output**. Every §B premise file and both doc targets are byte-unchanged, so the
  RED-now cells pinned to `d29b8942e` still hold on the current tree.
- **D1** — `grep -n 'PIDSource\|pid_source' internal/kanban/integration_lock.go` → no match
  (rc=1). AC-INL-001's **Then** now asserts the written marker as clause (b).
- **D2** — `awk '/^#### §4\.1\.2/,/^#### §4\.1\.3/' CLAUDE.local.md` → a bash procedure block
  only; `grep -n '직렬화' CLAUDE.local.md` → no match;
  `grep -n 'integration acquire\|integration release' CLAUDE.local.md` → exactly 2 hits
  (lines 312, 317), both command lines inside that block. The earlier RED premise
  ("§4.1.2 instructs the lead-notice-only workaround") is FALSE; AC-INL-011 and spec.md §B item 13
  are restated against the measurement. Baseline for AC-INL-011's new greps:
  `awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md | grep -c '세션 프로세스'` → **0**, same pipeline
  with `'재획득'` → **0**.
- **D3** — `grep -rn 'internal/kanban' internal/session/` → **1 hit**
  (`internal/session/session_pid.go:49`, a comment, not an import);
  `grep -rn '"github.com/modu-ai/moai-adk/internal/kanban"' internal/session/` → **0 hits**
  (rc=1). plan.md §C now uses the import-scoped form.
- **D4** — `grep -n 'Flock\|flock\|LockFile' internal/kanban/integration_lock.go internal/cli/integration.go`
  → exactly 5 lines: 2 `flock` prose lines in the package header (15, 19) and 3
  `IntegrationLockFileName` constant/use lines (38, 39, 106) matching only the `LockFile`
  substring. **No call site**, and `internal/cli/integration.go` contributes no match. `AcquireIntegrationLock` (integration_lock.go:146-181) reads at line 155 and
  writes at line 177 with nothing between; `writeIntegrationLock` stages to a fixed
  `path + ".tmp"` (line 229). Hazard recorded in spec.md §G; plan.md M5 carries a wording bound.
- **D5** — the regex claim above is now attributed to its exact command and pattern.
- **D6/D7/D8/D9 applied**: full REQ IDs in the §D matrix (`grep -c 'REQ-INL-008' acceptance.md`
  0 → **1**); REQ-INL-003 relabeled `(Event-driven)`; pid-reuse + absent-host hazards added to §G
  (`IntegrationLock` integration_lock.go:71-79 has no `Host` field; `session.Entry`
  registry.go:86-95 does); REQ-INL-006 reworded as a preserved invariant.

**Provenance of this repair pass is UNRESOLVED.** The directing actor for the iter-1 repair was
not identified. Three things are known and none of them names one: the orchestrator did not
instruct the repair; the audit report `.moai/reports/t298/plan-audit-iter1.md` contains no
repair record; and the only positive signal is a timeline plus a capability — the report landed
at 19:25:57 and the four artifacts were modified 19:33:09-19:34:36, and `plan-auditor`'s tool
list carries `Write` and `Edit`. Capability plus adjacency is not attribution, so this is
recorded as an open provenance gap rather than as a conclusion: the repair content itself was
re-measured in that pass and is accepted on its own evidence, independently of who directed it.

### v0.1.2 surgical amendment (plan phase, still `draft`)

Three additive changes; no existing REQ or AC renumbered, reworded, or removed.

- **AC-INL-012 added** (acceptance.md §D.1 + matrix row + MUST-PASS gate) — the acquire-refusal
  path. Gap it closes: AC-INL-001 gates the STATUS read and AC-INL-003 the takeover direction,
  so a mutant that anchors the owner pid for `Stale()` while leaving `acquire` willing to
  displace a live holder passes the whole prior set. **No new REQ was added**: REQ-INL-004 already
  names "acquire refusal" as a consumer leg, and REQ-INL-010 already forbids silently overwriting
  another lane's recorded trace. Measurements taken in this run on `c67a6ea64`:
  `grep -rln "exec.Command" internal/cli/integration_lock_cli_test.go internal/kanban/integration_lock_test.go`
  → no output (rc=1), i.e. no cross-process refusal coverage exists; the refusal branch is
  `internal/kanban/integration_lock.go:161-168` and is bypassed via `case current.Stale():`
  (line 163) because line 175 records the exiting CLI's own pid; the displaced-holder line is
  `internal/cli/integration.go:196-199`. The runtime leg is deliberately **measured-by-test**
  (M1) rather than by hand — `moai integration acquire` mutates the shared primary-checkout
  record and live lanes hold real merge windows in this repository right now.
- **spec.md §A extended** with the two 2026-08-27 production field observations (bare acquire
  displacing a live holder 42s after it was recorded; both lanes told the window was theirs,
  empty file intersection so no corruption) and with the identity-of-record vs atomicity layer
  distinction, so the card's scope is not read as delivering mutual exclusion.
- **plan.md M1** gains the corresponding RED test bullet; §H cross-reference range widened to
  AC-INL-012. Frontmatter `version` → `0.1.2` (spec.md is the only artifact carrying
  frontmatter); `status` remains `draft` — the Implementation Kickoff Approval gate has NOT been
  passed.

## §E.2 Run-phase Evidence

### M1 — RED: cross-process reproduction (card t298)

Measured on tree `ecf6382b6dcd1f5cf057e457e2c8116d9b0420ff` (`ecf6382b6`) @ `WT-integration-lock`,
in worktree `.claude/worktrees/t298`. Test-only milestone: no production file was modified.

**Artifact:** `internal/cli/integration_lock_owner_liveness_test.go` — three tests, all named
`TestIntegrationOwnerLiveness*`, sharing the `buildMoaiBinary` / `runIntegrationChild` helpers
required by plan.md §F M1. The child binary is produced by `go build -o <t.TempDir()>/moai
./cmd/moai` and exec'd directly; `go run` is not used (plan.md §D — the intermediate `go`
process is outside the ancestry walk's wrapper-name set).

**Command:**

```
go test ./internal/cli/ -run 'TestIntegrationOwnerLiveness' -v -count=1
```

**Selector match count:** 3 (`grep -c '^=== RUN'` over the transcript → `3`). The regex is not a
zero-match no-op: a zero-match selector prints `ok` and would not be a RED.

**Exit code:** `1`. All three tests FAIL.

**Verbatim output** — the durable copy is the block below, committed with this file. The
session-local copy at `.moai/state/verify/t298-m1/red-run.txt` is gitignored
(`.gitignore:284` → `.moai/state/`) and does not survive worktree disposal, so it is a
convenience artifact, not the evidence of record:

```
=== RUN   TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExits
    integration_lock_owner_liveness_test.go:184: window reads reclaimable while the owning session (this test process, pid 9509) is alive; recorded pid 10481 is the exited acquire CLI's: {"held":true,"lock":{"session_id":"sess-owner-a","session_name":"lane-owner-a","pid":10481,"branch":"","worktree":"/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExi3833388289/002","acquired_at":"2026-08-27T11:13:04Z"},"root":"/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExi3833388289/002","stale":true}
    integration_lock_owner_liveness_test.go:188: recorded pid = 10481, want the owning session's pid 9509
    integration_lock_owner_liveness_test.go:194: record pid_source = "", want "session-owner"; record = map[acquired_at:2026-08-27T11:13:04Z branch: pid:10481 session_id:sess-owner-a session_name:lane-owner-a worktree:/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExi3833388289/002]
--- FAIL: TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExits (4.06s)
=== RUN   TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits
    integration_lock_owner_liveness_test.go:221: window reads reclaimable while the stamped owner (pid 9509) is alive: {"held":true,"lock":{"session_id":"sess-owner-b","session_name":"lane-owner-b","pid":11393,"branch":"","worktree":"/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits3639930886/002","acquired_at":"2026-08-27T11:13:07Z"},"root":"/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits3639930886/002","stale":true}
    integration_lock_owner_liveness_test.go:224: recorded pid = 11393, want the stamped MOAI_SESSION_PID 9509
--- FAIL: TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits (3.79s)
=== RUN   TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder
    integration_lock_owner_liveness_test.go:249: bare acquire by sess-b took the window from live holder sess-a (exit 0): time=2026-08-27T20:13:11.632+09:00 level=WARN msg="config sections directory not found, using defaults" path=/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder198728607/002/.moai/config/sections
        release-integration window acquired by sess-b on 
          displaced: sess-a (pid 11862), held since 2026-08-27T11:13:11Z
    integration_lock_owner_liveness_test.go:256: on-disk session_id = "sess-b", want "sess-a" (the window must not have been transferred); record = map[acquired_at:2026-08-27T11:13:11Z branch: pid:12051 session_id:sess-b session_name:lane-b worktree:/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder198728607/002]
    integration_lock_owner_liveness_test.go:261: refused acquire reported a displacement; displaced-holder bookkeeping belongs to --force and stale-reclaim only: time=2026-08-27T20:13:11.632+09:00 level=WARN msg="config sections directory not found, using defaults" path=/private/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder198728607/002/.moai/config/sections
        release-integration window acquired by sess-b on 
          displaced: sess-a (pid 11862), held since 2026-08-27T11:13:11Z
--- FAIL: TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder (3.84s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	12.489s
FAIL
```

**Failure attribution — each test fails for the defect, not for a harness accident:**

| Test | AC | Observed | Why it is the defect |
|---|---|---|---|
| `_AncestryPathHoldsAfterAcquireCLIExits` | AC-INL-001 (a)+(b) | `stale: true`, recorded pid `10481` ≠ owning session pid `9509`; `pid_source` absent | The recorded pid is the exited child acquire CLI's own (`os.Getpid()`, integration_lock.go:174-176). The child built and ran, wrote a valid record, and exited 0 — so this is not a build, fixture, or missing-binary failure; only the pid identity is wrong. The absent `pid_source` is the baseline schema (spec.md §B verified premise). |
| `_EnvStampHoldsAfterAcquireCLIExits` | AC-INL-002 | recorded pid `11393` ≠ stamped `MOAI_SESSION_PID=9509` | The acquire path never consults `MOAI_SESSION_PID`; it records the child's pid. Platform-neutral (no ancestry walk), so it is also the Windows-semantics proxy. |
| `_BareAcquireRefusesLiveHolder` | AC-INL-012 (a)+(b) | child B exited 0, on-disk `session_id` flipped to `sess-b`, output carried `displaced: sess-a (pid 11862)` | Child A's recorded pid died with child A, so `current.Stale()` is true and acquire takes the `case current.Stale():` reclaim arm (integration_lock.go:163) instead of the refusal branch (161-168). This is the 2026-08-27 field harm reproduced mechanically. |

The AC-INL-012 holder-naming sub-assertion (`sess-a` present in child B's output) did not fire —
`sess-a` appears there in the `displaced:` line rather than in a refusal. That is consistent with,
not a weakening of, the RED: the three assertions that did fire establish that the transfer
happened.

**Isolation:** every fixture is under `t.TempDir()`; the child runs with `CLAUDE_PROJECT_DIR` and
`GIT_CEILING_DIRECTORIES` pinned to that temp root (same contract as `runIntegration`,
integration_lock_cli_test.go:26-43) and with `cmd.Dir` set to it, so no live integration-lock
record, session registry, or card queue was read or written. `CLAUDE_CODE_SESSION_ID` and
`MOAI_SESSION_PID` are scrubbed from the inherited child environment and put back only per case.
Child and build processes are bounded by `exec.CommandContext` timeouts (60s / 5m) with `defer
cancel()`, so no process outlives the test.

**Production files changed in M1:** none — verified by `git show --stat HEAD` on the M1 commit.

### M2+M3 — GREEN: owner-pid anchor lands (card t298)

Commit `3f3465369`. Measured on the working tree at that commit, in this run,
on branch `WT-integration-lock` (parent `ac68ef1ec`).

**Production files changed (3):**

| File | Change |
|---|---|
| `internal/session/session_pid.go` | export `ResolveOwnerPID() (pid int, resolved bool)` — the existing chain (live `MOAI_SESSION_PID` stamp → nearest non-wrapper ancestor) with NO `os.Getpid()` third step; `resolveSessionPID` now wraps it and keeps that fallback for the registry alone |
| `internal/cli/integration.go` | acquire resolves the owner pid and records it plus `PIDSource: kanban.PIDSourceSessionOwner` |
| `internal/kanban/integration_lock.go` | additive `PIDSource string \`json:"pid_source,omitempty"\`` + the `PIDSourceSessionOwner` constant; **deleted** `if want.PID == 0 { want.PID = os.Getpid() }` (the defect line, formerly 174-176) |

**Test files changed (2):** `internal/session/session_pid_test.go` (adds
`TestResolveOwnerPID_Precedence`, 3 sub-tests pinning env-stamp / ancestry /
unresolvable-reports-zero), `internal/kanban/integration_lock_test.go` (M3
regression + backward-compat rows, below).

**AC-INL-001 / 002 / 012 — the M1 RED flips.** Command, verbatim output:

```
$ go test ./internal/cli/ -run 'TestIntegrationOwnerLiveness' -v -count=1
=== RUN   TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExits
--- PASS: TestIntegrationOwnerLiveness_AncestryPathHoldsAfterAcquireCLIExits (3.91s)
=== RUN   TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits
--- PASS: TestIntegrationOwnerLiveness_EnvStampHoldsAfterAcquireCLIExits (4.03s)
=== RUN   TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder
--- PASS: TestIntegrationOwnerLiveness_BareAcquireRefusesLiveHolder (3.89s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	12.624s
```

Selector match count is **3** — three `=== RUN` lines for three tests, the same
three that failed in M1. A zero-match selector prints `ok` and proves nothing,
so the count is stated rather than inferred. The pre-fix RED transcript for the
same command is preserved verbatim at `.moai/reports/t298/red-baseline.txt`.

**AC-INL-004 / 007 (M3) — new rows, all green:**

| Test | AC | Asserts |
|---|---|---|
| `TestAcquireIntegrationLock_AnchoredPIDZeroIsLiveNotStale` | AC-INL-004 | an anchored pid-0 record reads live, and a bare second acquire over it returns the contention sentinel |
| `TestReadIntegrationLock_LegacyRecordWithoutPIDSource` | AC-INL-007 | a hand-written pre-anchor record (no `pid_source` key, dead pid) parses, carries empty `PIDSource`, reads reclaimable, is taken over with the displaced holder reported, and re-acquire still refreshes |
| `TestAcquireIntegrationLock_RecordsTheCallersOwnerPID` | AC-INL-001 (b) at the package layer | the caller's resolved pid and the marker land on disk verbatim |

`go test ./internal/kanban/ -run '<the three above>' -v -count=1` reported 3
`--- PASS` / `--- SKIP` lines (count measured, not assumed).

**One pre-existing assertion was updated, deliberately.**
`TestAcquireIntegrationLock_RecordsHolder` asserted `got.PID == os.Getpid()` —
it asserted the deleted line. It now asserts `got.PID == 0` plus
`!got.Stale()`: a caller that supplies no pid gets no invented one, and the
conservative reading of an unset pid is live. No other pre-existing assertion
changed; the two refusal tests the lead named were verified untouched and green
(below).

### M4 — guard consumer + platform parity (card t298)

Commit `d5006ff25`. Test-only; `internal/hook/integration_lock_guard.go` is
unchanged — it consumes `Stale()`, and `Stale()` is what M2 fixed.

**AC-INL-008.** `TestCheckIntegrationLock_FollowsAnchoredLiveness`, three
sub-tests on one run:

```
$ go test ./internal/hook/ -run 'TestCheckIntegrationLock' -v -count=1
--- PASS: TestCheckIntegrationLock_FollowsAnchoredLiveness (0.00s)
    --- PASS: TestCheckIntegrationLock_FollowsAnchoredLiveness/anchored_live_holder_denies (0.00s)
    --- PASS: TestCheckIntegrationLock_FollowsAnchoredLiveness/anchored_dead_holder_allows (0.00s)
    --- PASS: TestCheckIntegrationLock_FollowsAnchoredLiveness/anchored_pid-0_holder_denies_conservatively (0.00s)
```

The four pre-existing guard tests (`_DeniesAForeignLiveHolder`,
`_AllowsTheHolder`, `_FailsOpen` with 6 sub-tests, `_OnlyGuardsMerge`) ran in
the same invocation and all passed.

**AC-INL-009 — cross-compile parity.**

```
$ GOOS=windows GOARCH=amd64 go vet ./internal/...
(no output)   rc=0
```

Recorded at `.moai/reports/t298/vet-windows.txt`. **This proves compilation,
not behavior**: `go vet` under a cross GOOS does not build or run tests, so the
Windows *behavior* leg rests on the platform-neutral env-stamp test (AC-INL-002)
and the pid-0 conservative test (AC-INL-004), neither of which uses the ancestry
walk. `internal/kanban/factory_alive_windows.go` is untouched —
`git diff --stat -- internal/kanban/factory_alive_windows.go` is empty.

### M5 — documentation (card t298)

Commit `1061b10a4`. Two local-only files; no template mirror, no `make build`.

| AC | Command | Observed | Want |
|---|---|---|---|
| AC-INL-010 | `grep -c '직렬화를 보장하지 못한다' .claude/rules/local/gitflow-lane-protocol.md` | `0` | 0 (caveat gone) |
| AC-INL-010 | `grep -c '세션 프로세스에 묶여' .claude/rules/local/gitflow-lane-protocol.md` | `1` | ≥1 (replacement present) |
| AC-INL-011 | `awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md \| grep -c '세션 프로세스'` | `1` | ≥1 |
| AC-INL-011 | `awk '/^### §4\.1 /,/^## 5\./' CLAUDE.local.md \| grep -c '재획득'` | `1` | ≥1 |
| AC-INL-013 | `grep -hoE '직렬화를[[:space:]]*보장\|동시.*acquire.*불가능\|acquire.*동시.*불가능' <both files> \| wc -l` | `0` | 0 |

AC-INL-011's two greps were `0` / `0` on the baseline (recorded in
acceptance.md's RED cell) and are `1` / `1` now.

### Package suites (every package this branch touches)

The touched set was derived, not assumed:
`git diff --name-only origin/develop...HEAD` plus the working tree named
`internal/cli`, `internal/hook`, `internal/kanban`, `internal/session`. Each was
run non-recursively (no `/...`) so `internal/hook/perf` — which rewrites SPEC
fixtures — was never entered, and `e2e/` was never run.

```
$ go test ./internal/cli/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	247.154s        (real 248.86s)

$ go test ./internal/kanban/ ./internal/hook/ ./internal/session/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/kanban	12.517s
ok  	github.com/modu-ai/moai-adk/internal/hook	25.064s
ok  	github.com/modu-ai/moai-adk/internal/session	5.866s
```

Full logs: `.moai/reports/t298/suite-cli.txt`, `.moai/reports/t298/suite-others.txt`.

**The two pre-existing refusal tests the dispatch named, run individually:**

```
$ go test ./internal/cli/ -run 'TestIntegrationAcquire_RefusesASecondLane' -v -count=1
--- PASS: TestIntegrationAcquire_RefusesASecondLane (0.06s)

$ go test ./internal/kanban/ -run 'TestAcquireIntegrationLock_RefusesASecondLiveSession' -v -count=1
--- PASS: TestAcquireIntegrationLock_RefusesASecondLiveSession (0.00s)
```

**Fixture hygiene:** `git status --porcelain | grep -E '\.moai/specs/|internal/hook/perf/'`
returned nothing at the end of the run — no SPEC fixture and no perf fixture was
left modified.

**Isolation:** every new and changed test builds its fixtures under
`t.TempDir()`; the cross-process cases pin `CLAUDE_PROJECT_DIR` and
`GIT_CEILING_DIRECTORIES` to that root and set `cmd.Dir` to it. The live
primary-checkout `.moai/state/integration-lock.json`, the session registry, and
the card queue were never read or written by any test in this run.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-08-27
run_commit_sha: 1061b10a4        # M5, the last run-phase commit on WT-integration-lock
run_status: audit-ready
milestones_landed: M2, M3, M4, M5   # M1 landed earlier (RED authoring)
ac_pass_count: 11                # AC-INL-001..004, 007..013 observed green in this run
ac_fail_count: 0
ac_not_observed: 2               # AC-INL-005, AC-INL-006 — pre-existing invariants, green
                                 # inside the package suites but not re-asserted individually
preserve_list_post_run_count: 0  # nothing outside the SPEC's declared scope was modified
new_warnings_or_lints_introduced: none observed
cross_platform_build:
  goos_windows_vet: clean (rc=0, no output)
  caveat: vet proves compilation, not behavior
total_run_phase_files: 8         # 3 production, 3 test, 2 local-only docs
m1_to_mN_commit_strategy: >
  four commits on WT-integration-lock — M1 (RED, earlier), M2+M3 (3f3465369),
  M4 (d5006ff25), M5 (1061b10a4). M2 and M3 share one commit because both
  land in internal/kanban/integration_lock_test.go and splitting a single
  file across two commits would have left the tree non-compiling in between.
not_pushed: true                 # the lane owns integration; no push, no merge
```

**Residual risk, restated from spec.md §G and unchanged by this run:** the
acquire path is still an unserialized read-modify-write, so two lanes acquiring
in the same instant can both believe they hold the window. This run fixed
identity of record (whose pid), not atomicity. That is why the M5 prose says
plainly that the record is a coordination signal and the lead announcement
remains the first layer.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-08-27
sync_commit_sha: pending-backfill   # backfilled in the immediately following commit
sync_status: completed              # clean PASS → completed, per the status rule below
sync_audit:
  verdict: PASS                      # clean — sync-auditor `audit-t298`, fresh context, --deep
  overall_score: 0.9189              # 4-dimension harmonic mean; Tier M PASS threshold 0.80
  dimensions:                        # as RE-DERIVED; see provenance_hazard below
    functionality: 0.95              # must-pass — passes independently
    security: 0.87                   # must-pass — passes independently
    craft: 0.92
    consistency: 0.94
  findings: 4                        # F1..F4, ALL optional
  blocking_findings: 0
  report: .moai/reports/t298/sync-audit.md   # the auditor's exclusive write scope
  read_the_dimensions_not_the_aggregate: >
    The aggregate barely moved across the re-derivation (0.9193 → 0.9189) while BOTH
    must-pass dimensions moved, in OPPOSITE directions: Functionality 0.94 → 0.95 (the
    §G residual judged not to belong as a Functionality deduction) and Security
    0.88 → 0.87 (a further deduction for the undeclared F2). The two corrections
    offset. A reader comparing only the harmonic means would conclude nothing changed,
    which is the opposite of what happened — the auditor flags this trap itself.
  spawned_by: orchestrator           # manager-docs carries no Agent tool and cannot
                                     # spawn one; the audit is not this agent's to run
  status_rule: >
    clean PASS → `completed`; PASS-WITH-DEBT → `implemented`. The spec.md `status:`
    field is held at `in-progress` until the verdict is relayed, and the transition
    is the last edit before the close commit.
  provenance_hazard:                 # recorded so a later reader sees this unprompted
    what_happened: >
      The audit path was OCCUPIED when the auditor arrived. Before the orchestrator
      spawned `audit-t298`, manager-docs had written its own 4-dimension read to
      `.moai/reports/t298/sync-audit.md` — the same path — and committed it in
      `2b49785de`. The auditor read that file before manager-docs removed it, and
      its report opens by referring to that content. The independence of a
      fresh-context audit was therefore compromised at the input, not at the verdict.
    how_it_surfaced: >
      manager-docs named the hazard when it removed the file, stated the signal to
      watch for (a verdict at or near its own PASS-WITH-DEBT 0.924), and said plainly
      that whether the auditor had READ the path was not observable to it. The
      orchestrator then confirmed the read had already happened.
    what_it_produced: >
      Two of four dimensions came back numerically identical to manager-docs's read
      (Functionality 0.94, Security 0.88) — and those two are the must-pass pair.
      Craft and Consistency diverged (0.95→0.92, 0.93→0.94), and the verdict diverged
      in kind (PASS-WITH-DEBT 0.924 → PASS 0.9193).
    how_it_was_handled: >
      The orchestrator sent the auditor back to RE-DERIVE the two must-pass figures
      from its own evidence, with an explicit instruction not to lower the verdict
      merely to appear independent. Those two figures are therefore after-the-fact
      re-derivations, not first-pass observations, and are recorded as such here.
    what_the_re_derivation_showed: >
      Both must-pass figures MOVED, in opposite directions: Functionality 0.94 → 0.95
      and Security 0.88 → 0.87. Neither still coincides with the contaminating read.
      The auditor also stated plainly that it HAD read the manager-docs file before
      fixing its first figures — the honest answer, and the one that made the
      re-derivation necessary. The aggregate moved only 0.9193 → 0.9189 because the
      two corrections offset, which is why the per-dimension derivations, not the
      harmonic mean, are the evidence here.
    why_the_judgment_still_reads_as_its_own: >
      The auditor's §F deductions cite findings manager-docs never made — F1 (the
      AC-INL-003 fixture seeds a legacy-shaped record, not the anchored one the Given
      describes), F2 (Windows `isProcessAlive` returns true unconditionally, so a dead
      `MOAI_SESSION_PID` is recorded — the one path tilting AGAINST the declared
      TREAT-AS-LIVE asymmetry, and undeclared in §G), and F4 plus three `t.Skip`
      escape hatches. Same numbers, different deductions.
    second_artifact_dangling_reference: >
      The hazard produced TWO artifacts, not one. Besides the touched figures, the
      auditor's report initially carried a dangling forward reference: it was written
      to be APPENDED below the manager-docs report and opened by referring to that
      content as its §0 — but manager-docs had already removed the file from the
      worktree, so what landed was the appendix alone, citing a section that was not
      in it (`grep -c 'Sync-phase Quality Assessment'` → 0; first heading at line 4
      was the auditor's own). The superseded manager-docs report remains retrievable
      at `2b49785de:.moai/reports/t298/sync-audit.md`. Corrected BEFORE this commit:
      the auditor rewrote its opening to stand alone and replaced the dangling
      reference with a provenance paragraph carrying that path-at-commit citation.
      Restoring the manager-docs report above the audit was considered and rejected —
      two verdicts in one artifact, with no way for a later reader to tell at a glance
      which one binds, is a worse defect than the one being fixed.
    residual: >
      Functionality and Security are the two figures the contamination touched; they
      stand at 0.95 and 0.87 as re-derived, not at the 0.94 / 0.88 that coincided
      with the contaminating read. Craft (0.92), Consistency (0.94), the four
      findings, and the §G confirmation were never affected — Craft and Consistency
      differed from the contaminating read at first derivation. What cannot be
      recovered is a first-pass, never-contaminated derivation of the must-pass pair;
      the re-derivation is the best available evidence, not a substitute for one.
newly_surfaced_residual:
  finding_id: F2
  surfaced_by: fresh-context sync-auditor `audit-t298`   # NOT manager-docs, NOT the
                                                         # orchestrator — neither the
                                                         # SPEC nor this close caught it
  site: internal/session/anchor_pid_windows.go:14
  what: >
    On Windows `isProcessAlive` returns `true` unconditionally. That makes the
    liveness check inside `sessionPIDFromEnv` (session_pid.go:99) vacuous there, so a
    `MOAI_SESSION_PID` naming an already-dead process is recorded as the holder pid.
    Judgment after the record is made uses `kanban.FactoryProcessAlive`, which on
    Windows IS a real `OpenProcess` probe — so that window then reads reclaimable.
  why_it_matters: >
    This is the one path that tilts AGAINST the TREAT-AS-LIVE asymmetry §D declares.
    Every other indeterminate path in this design resolves toward "live, ask the
    holder to release"; this one resolves toward "reclaimable", which is the
    two-lanes-merge-at-once direction the lock exists to prevent. It also leaves
    acceptance §D.2 ("a dead stamp must be skipped") unmet on that platform.
  scope_note: >
    Narrow premise (an inherited stale env stamp) and NOT a regression this card
    introduced — but undeclared in `spec.md` §G, which is what makes it a finding
    rather than a known residual.
  classification: >
    (i) a §G DOCUMENTATION OMISSION, AND (iii) FOLLOW-UP CARD MATERIAL —
    explicitly NOT (ii) a defect this card should have covered.
  classification_reasoning: >
    `anchor_pid_windows.go:14` is the session REGISTRY's platform probe, not the
    acquire path this card scoped, and it returned unconditionally `true` before this
    SPEC existed (zero diff on that file). The defect was not created by this card; it
    became VISIBLE because this card's anchor now sits on top of that probe. Fixing it
    here would mean changing the registry's Windows liveness semantics — the surface
    §D.4 explicitly declared out of bounds — so covering it in this card was never the
    correct move.
  the_g_gap_precisely: >
    The omission is NARROWER than "F2 is missing from §G". §G already documents the
    STAMP-ABSENT path (no `MOAI_SESSION_PID`, owner unresolvable → pid 0, conservative
    degrade toward live). What it omits is the STAMP-PRESENT-BUT-DEAD path: a stamp
    naming an already-exited process, which on Windows passes the vacuous liveness
    check, is recorded, and then reads reclaimable. That precision is what makes the
    follow-up card actionable rather than a restatement of the finding.
  spec_g_not_edited: true    # §G is spec body — not this agent's to edit, and not
                             # this phase's job. Recorded here, not there.
ac_pass_count: 13                   # AC-INL-001..013, counted against acceptance.md
ac_fail_count: 0
ac_not_observed: 0                  # up from §E.3's 2 — AC-INL-005 and AC-INL-006 were
                                    # closed individually in sync-phase (below)
changelog_entry_position: "## [Unreleased] → ### Fixed, first entry"
frontmatter_status_transitions:
  spec_md: in-progress → completed
  plan_md: n/a          # carries no YAML frontmatter — verified, not assumed
  acceptance_md: n/a    # same
  progress_md: n/a      # same
  note: >
    Only spec.md carries a frontmatter block in this SPEC, so it is the sole
    machine-read status surface and the only file transitioned. Measured:
    `grep -nE '^(status|updated|version):'` across all four returned matches in
    spec.md only. The prose preambles in plan.md and progress.md that read
    "Status: draft" are historical statements about plan-phase authorship and
    are body content this agent does not own — left unedited deliberately.
  updated_field: already 2026-08-27 in spec.md; no refresh needed
b12_self_test_a: >
  Pre-emission grep. `grep -c 'SPEC-INTEGRATION-LOCK-LIVENESS-001\|t298' CHANGELOG.md`
  → 0 before the write. No duplicate entry from a parallel session.
b12_self_test_b: >
  AC count match. `grep -oE 'AC-([A-Z0-9]+-)*[0-9]+' acceptance.md | sort -u | wc -l`
  → 13, non-zero and plausible; the CHANGELOG entry states the same 13.
b12_self_test_c: >
  File-path verification. Every path this entry names was confirmed to exist by `ls`
  before the commit.
sync_phase_go_diff: none            # this close touches markdown only; measured, not assumed
```

**Sync-phase verification observed in this run** (all in worktree `.claude/worktrees/t298`,
HEAD `f8b7264ba`):

```
$ go vet ./internal/kanban/ ./internal/session/ ./internal/cli/ ./internal/hook/
(exit 0, no output)

$ go test ./internal/kanban/ -run 'TestReleaseIntegrationLock_HolderAndForeign' -v -count=1
--- PASS: TestReleaseIntegrationLock_HolderAndForeign (0.00s)        # AC-INL-005

$ go test ./internal/cli/ -run 'TestIntegrationAcquire_ForceReportsWhatItDisplaced' -v -count=1
--- PASS: TestIntegrationAcquire_ForceReportsWhatItDisplaced (0.06s) # AC-INL-006
```

The two ACs §E.3 recorded as `ac_not_observed` were preserved invariants, green inside the
package suites but never asserted individually. They are now observed by name, which is why
this section reports 13/13 rather than 11 + 2. No full suite was run: per the repository's
verification-load discipline the full-suite verdict belongs to CI on the pushed branch.

### Real-environment confirmation — evidence that did not exist at run-phase time

The fix sat undeployed for most of 2026-08-27, and during that window three lanes
independently observed the defect it fixes. Recorded here as **observation by other
sessions**, not as a self-report:

- **Before deployment**: lanes lane-16, lane-17 and lane-18 each observed `reclaimable` with
  an acquire-CLI pid recorded as the holder. One real eviction occurred without `--force`.
- **lane-18's isolated A/B**, `CLAUDE_PROJECT_DIR` pinned so the live lock was never touched:
  the old binary reproduced all four stages of the failure chain; the HEAD build reported
  `held` and refused a second acquire with exit 1.
- **After deployment, first live use** (this card's own integration window): `held`, pid
  33289 — the session-owning process. The lead had replaced the binary at `343399d2f`.
- **The discriminating detail**: the holder UUID was recorded correctly in **both** cases;
  only the pid differed. That is what confirms the diagnosis — the defect was never in
  identity *interpretation*; the recorded pid itself was a value unrelated to the session.

This confirms the fix behaves correctly in the live environment. It is **not** a proof of
atomicity and must not be read as one.

Making the deployment lag itself visible — a fixed binary sitting unshipped while lanes keep
hitting the old defect — is **out of scope for this card and re-scoped to card t326**
(lane-18). Nothing was done about it here.

### §G residual verdict — still open

Re-measured on this tree at HEAD `f8b7264ba`:

```
$ grep -n "Flock\|flock\|LockFile" internal/kanban/integration_lock.go internal/cli/integration.go
internal/kanban/integration_lock.go:15    # "flock" in prose, package header comment
internal/kanban/integration_lock.go:19    # "flock" in prose, package header comment
internal/kanban/integration_lock.go:38    # IntegrationLockFileName doc comment
internal/kanban/integration_lock.go:39    # const IntegrationLockFileName
internal/kanban/integration_lock.go:128   # its use in the path join
```

Five lines, **no call site**; `internal/cli/integration.go` contributes **zero** matches. The
§G residual — the acquire path as an unserialized read-modify-write — is therefore **still
open**. No later commit closed it, and this card did not widen to cover it. (§G's own citation
of line 106 is correct: it is attributed by name to `c67a6ea64`, and the M2 comment additions
moved the path join to 128 afterwards.)

The independent audit **confirmed** this reading against its own measurement, and additionally
verified by diff that the M5 prose did not paper the residual over. It also surfaced a **second,
undeclared asymmetry** that §G does not mention — the Windows `isProcessAlive` path, recorded
above as `newly_surfaced_residual` (finding F2). That one is attributed to the audit: neither
this SPEC nor this close caught it.

Operational consequence, unchanged and restated so no reader infers otherwise: a recorded hold
is a **coordination signal, not a permission boundary**, so **the lead announcement remains
the first serialization layer**. A follow-up card for the serialization work is recommended to
the lead; card issuance is not this agent's power and none was filed.

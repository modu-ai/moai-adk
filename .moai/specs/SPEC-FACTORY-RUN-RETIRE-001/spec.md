---
id: SPEC-FACTORY-RUN-RETIRE-001
title: "Factory run retirement — owner-liveness reconciliation so a dead lead's run leaves 'active'"
version: "0.13.1"
status: in-progress
created: 2026-09-23
updated: 2026-09-25
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/homestate, internal/factorymsg, internal/cli"
lifecycle: spec-anchored
tags: "factory, run-lifecycle, liveness, ambiguous-factory, migration, boot-proof, card-t1107, card-t1169"
tier: L
card: t1107
amendment_of: SPEC-FACTORY-RUN-RETIRE-001
related_specs: [SPEC-FACTORY-MIXED-HOOK-001, SPEC-FACTORY-WORKER-NAMING-001, SPEC-FACTORY-MODE-001]
---

# SPEC-FACTORY-RUN-RETIRE-001 — Factory run retirement

## §A Background and Motivation

A factory run is recorded `status='active'` and **no non-test code path ever takes it out of that
state**. Launching the factory lead twice in the same project in different seconds therefore leaves
two `active` rows, and every subsequent worker join fails closed with `AMBIGUOUS_FACTORY` until the
operator passes `--factory-run <id>` explicitly.

The defect is reproduced, not inferred. The evidence-bearing reproduction report is
`.moai/reports/t1107/verdict.md` (card t1107, measured 2026-09-23 on darwin against a
`make build` binary of local develop `176d8b658`); its §4 Gaps and §5 Residual-risk are inputs to
this SPEC and are addressed or explicitly carried forward in §D and §F below.

Measured mechanism (re-measured in this tree while authoring this SPEC):

- `internal/kanban/bootstrap.go:97` — `NewRunID()` is `base36(time.Now().Unix())`, **second**
  granularity. Two lead launches in the same second collapse to one row through the
  `ON CONFLICT(run_id) DO UPDATE` clause; the trigger is specifically a **different-second**
  relaunch.
- `internal/homestate/runtime.go:23` `RecordRun` — `INSERT ... 'active' ... ON CONFLICT DO UPDATE
  SET status='active'`. This is the only writer of the `runs.status` column.
- `internal/factorymsg/store.go:274-276` — two or more `active` rows produce `AMBIGUOUS_FACTORY`.
- `grep -rn "UPDATE runs\|DELETE FROM runs" --include='*.go' . | grep -v _test.go` → **0 hits**
  (exit 1). Control with the same search shape: `status='active'` → 2 non-test hits, so the
  instrument works and the absence is not a search failure.

Two further facts, measured in this tree, shape the prescription:

1. **The `runs` row carries no owner identity.** The `runs` DDL
   (`internal/homestate/factory.go`) has `run_id`, `lead_session_id`, `lead_backend`, `status`,
   `manifest_json`, `created_at`, `updated_at` — no PID, no process-start fingerprint. Nothing in
   the row can answer "is this run's lead still alive?".
2. **A lead process identity IS already recorded elsewhere, per run.** `registerFactoryLaunchPending`
   (`internal/cli/factory_launch_pending.go:34`) registers a `role='lead'` peer carrying `PID` and
   `ProcessStart` into the per-run broker database (`peers` table, `internal/factorymsg/store.go`).
   This is the t1074 PID-plus-start-identity convention, already in service.

### A.1 Three launch shapes, five call sites

There are **five** `registerFactoryLaunchPending` call sites in this tree, in **three** shapes. The
axis is the **door's launch shape**, not the platform — two of the three shapes appear on darwin.

| Call site | Identity registered | The launching process afterwards | Shape |
|---|---|---|---|
| `internal/cli/launch_exec_posix.go:33` | launcher PID + `CurrentProcessFingerprint()` | replaced by `syscall.Exec` — launcher **is** the session | **replace** |
| `internal/cli/codex_direct_posix.go:34` | launcher PID, then `syscall.Exec` | replaced — launcher **is** the session | **replace** |
| `internal/cli/launch_exec_windows.go:54` | `child.Process.Pid` + `childFingerprint` | blocks in `child.Wait()` for the session's life | **spawn** |
| `internal/cli/codex_direct_windows.go:24` | child PID, then `cmd.Wait()` | blocks for the session's life | **spawn** |
| `internal/cli/codex_launcher.go:230` | **tmux pane PID** (`#{pane_pid}`) + its fingerprint | **exits immediately** | **pane** |

**The two `codex_direct_*` sites are covered, and the coverage is asserted rather than assumed.**
`codex_direct_posix.go:34` is the replace shape: `syscall.Exec` preserves PID and start time, so the
record-time stamp already names the session, exactly as `launch_exec_posix.go:33` does.
`codex_direct_windows.go:24` is the spawn shape: it registers the child and then blocks in
`cmd.Wait()`, exactly as `launch_exec_windows.go:54` does, so **REQ-002b's existing rule already
covers it — the rule needs no new clause, and the call site does need the restamp call.** Those are
two different claims, and collapsing them is defect D18: "an existing rule covers this door" never
means "this door's call site needs no edit". Every non-replace door carries the restamp call; what
the shape match settles is only that no door-specific rule has to be written for this one.
Neither `codex_direct_*` site needs a rule of its own — but neither may be left unnamed either: a reader who
finds five call sites and a SPEC discussing two would reasonably conclude three were missed.

**The pane shape is the uncovered one.** At `codex_launcher.go:230` the registered identity is the
**tmux pane's** process (`tmuxPanePID` → `#{pane_pid}`), resolved live by
`defaultCodexSpawnPaneIdentity` — a process that is neither the launcher nor a child of it, and that
outlives the launcher entirely. The launcher meanwhile reaches `recordFactoryRunStart`
(`codex_launcher.go:515`) **before** `runCodexLaunch` opens the pane, and then returns and exits.
So a record-time stamp on this door names a process that is already gone by the time anyone reads
it, while the session it was supposed to describe runs on in the pane. Reconciliation would classify
that owner dead and **retire a live session's run** — precisely the boundary this SPEC exists to
protect. `codex_launcher.go` carries no build tag, so this hole is on **darwin**.

Control proving the axis is door-shape and not platform: `moai cc -f --spawn` is safe, because
`cc.go:153-155` returns before `recordFactoryRunStart` at `cc.go:191` — the spawning process never
records a run at all; the spawned session records its own.

A first draft of this SPEC asserted that stamping the recording process was valid on every shape
with no per-platform branch. That premise is false on both the spawn and the pane shapes, and it
breaks the live-run protection directly. The SPEC therefore fixes the identity **semantically**: the
owner of a run is the **session process** — the process the `role='lead'` peer already names — and
every door whose launching process is not that session restamps to it (REQ-002b), or refuses to
launch when no session identity can be obtained (REQ-002d).

### A.2 The three-OS green path exists, and it lands after the merge

Measured in this tree at `bb5b8f9d1`, and confirmed against a real run:

- `.github/workflows/ci.yml:17-18` — `on: push: branches: [main, develop]`. A develop push triggers
  the workflow; `pull_request` is scoped to `branches: [main]`.
- `ci.yml:79-85` — the `go_code` path filter matches `**/*.go`, `go.mod`, `go.sum`, `Makefile`,
  the two workflow files, and `.moai/**`.
- `ci.yml:120,124` — job `test`: `if: needs.detect.outputs.go_code == 'true'`, matrix
  `os: [ubuntu-latest]`.
- `ci.yml:373-381` — job `test-integration`: `needs: test`, matrix
  `os: [ubuntu-latest, macos-latest, windows-latest]`, and **no `if:` of its own** — it inherits its
  gate through `needs: test`.

Confirming run (`gh run view 35802361895`): `event: push`, `branch: develop`, head
`1dbe5e2f33147b8ae58dfb76f5bde3a43ba4b47f`; jobs `Integration Tests (ubuntu-latest)`,
`(macos-latest)`, `(windows-latest)` each `success`. The run's overall conclusion was `failure` on
unrelated jobs — worth stating, because it shows the three-OS jobs execute and report on their own
terms rather than only inside an all-green run.

**The timing, stated plainly.** That trigger is a **push to `develop`**, which happens only after a
lane's branch merges; `pull_request` does not cover `develop`, and this project does not push `WT-`
branches, so a card branch receives **no CI run at all before merge**. Therefore:

| Verification | When | How |
|---|---|---|
| darwin | pre-merge | the lane's own local run, recorded with command and output |
| ubuntu | post-merge | job `test` on the develop push |
| ubuntu + macos + windows | post-merge | job `test-integration` on the same develop push |

An acceptance criterion worded as though the three-OS result gates this card's merge would be false
about its own timing — the same defect class as the one this section corrects, one step later. AC-013
is written against the table above.

## §B Requirements (GEARS)

### Owner identity

- **REQ-001** (Ubiquitous): The factory run record shall carry the run's **session process** owner
  identity — a process id and a process-start fingerprint — so that a run's liveness is answerable
  from the run record itself.
- **REQ-002** (Event-driven): **When** a factory run start is recorded, the factory state shall
  stamp the recording process's PID and its process-start fingerprint onto that run's row in the
  same transaction that writes the row.
- **REQ-002b** (Capability gate): **Where** a launch door does not replace the launching process
  with the session process — the **spawn** shape (a child the launcher waits on) and the **pane**
  shape (a tmux pane the launcher leaves running) alike — the launcher shall restamp the run's owner
  identity with that session process's PID and fingerprint before the session becomes reachable, so
  that on every door the stamped identity and the registered `role='lead'` peer identity name the
  **same** process and the REQ-006 fallback can never disagree with the REQ-002 primary about which
  process owns a run.
- **REQ-002d** (Event-driven): **When** a launch door that does not replace the launching process
  cannot obtain a live session identity, the launcher shall refuse the launch and shall leave no run
  stamped with the launching process's identity — because that identity is known in advance to die,
  and a run carrying it would be retired while its session is alive.
- **REQ-003** (Ubiquitous): The owner-liveness predicate shall classify a run's owner as exactly one
  of `live`, `dead`, or `indeterminate`, using PID **together with** process-start identity — a
  bare PID probe is insufficient because process ids are reused.
- **REQ-003b** (Ubiquitous): The predicate shall resolve any residual identity ambiguity toward
  `live`. The linux process-start fingerprint has **one-second** resolution
  (`internal/homestate/process_fingerprint_unix.go` reads `ps -o lstart=`), so a PID reused within
  the same second as its predecessor's start is not distinguishable from it there; that case shall
  classify `live`, never `dead`, keeping the residual error on the side that cannot retire a live
  run.

### Reconciliation

- **REQ-004** (Event-driven): **When** factory run resolution observes more than one active run,
  the resolver shall first retire every active run whose owner is classified `dead`, then re-resolve
  over the remaining active runs.
- **REQ-005** (Ubiquitous): **Every retirement path** — the resolution-time reconciler, the
  legacy-row migration pass, and the operator maintenance command alike — shall retire a run **only
  on a positive `dead` classification**. Every other classification value, present or future,
  declines retirement by default.

  The rule is stated **positively**, not as a reject-list, and that is the load-bearing part. A rule
  written "reject `live`, reject `indeterminate`" is an enumeration: it is correct only for the
  classification set that existed when it was written, and it silently begins retiring any state
  added later. The positive form fails safe by construction — a fourth classification retires
  nothing until someone deliberately admits it.

  It is stated **once**, for every path, because stating it twice is what let the copies drift. The
  D14 defect was exactly that: an earlier draft bound the reconciler to reject `live` and
  `indeterminate` while binding the operator command to reject `live` alone, so `indeterminate` fell
  through to retirement on the command. The consequence is host-wide rather than an edge case — on
  a host where the liveness probe fails consistently, **every** run carrying an identity classifies
  `indeterminate` (the REQ-006b boot proof reads clocks, not the probe, and is the one exception), so
  `--retire` would retire anything asked of it including live sessions, and §F's "never to retiring
  a live run" would be false for that entire host.
- **REQ-006** (Event-driven): **When** a run row carries no owner stamp — `lead_pid` below 1, the
  shape of every row written before this SPEC lands — the reconciler shall take the run's registered
  `role='lead'` peer identity (PID plus process start) as the liveness source. A row whose
  `lead_pid` is 1 or more carries an owner stamp and shall not be routed to the peer, even when its
  `lead_process_start` is empty. **When** neither source yields a **complete identity** — a PID of 1
  or more together with a non-empty process start — the reconciler shall classify the owner `dead`
  exactly when the REQ-006b boot proof holds for that run, and `indeterminate` in every other case.

  Two row shapes therefore reach the proof: an unstamped row whose peer lookup yields no complete
  identity, and a **partial stamp** — `lead_pid` of 1 or more with an empty `lead_process_start`.
  Current builds write the partial stamp whenever the recording process cannot read its own
  fingerprint: `recordFactoryRunStart` (`internal/cli/factory.go`) stamps
  `homestate.CurrentProcessFingerprint()`, which returns the empty string on any non-live probe of
  its own process (`internal/homestate/profile_lease.go`). Sending the partial stamp to the proof
  without consulting the peer is sound because REQ-006b premise 2 requires the very broker file the
  peer lookup reads to be absent: a run the proof retires cannot have had a peer record to consult,
  and a run whose broker file exists stays `indeterminate` — never a more permissive answer than a
  peer-based classification would have given. The argument holds only while the peer lookup and the
  premise-2 check name the same broker file; moving the peer store out of that file would break it
  without any signal, which is why AC-018 carries a partial-stamp leg on each side of premise 2.
- **REQ-006b** (Event-driven): **When** a run has no complete owner identity from either REQ-006
  source, the reconciler shall classify its owner `dead` only if **all four** of the following
  premises are established for that run, and shall classify it `indeterminate` if any premise is
  false or cannot be established:
  1. the host's last boot time is available (REQ-006c);
  2. no lead record exists for the run — established **only** by the filesystem reporting that the
     run's broker file does not exist; the file existing, a check failing for any other reason, or a
     broker path that cannot be derived each count as "a record may exist" and decline the proof;
  3. at least one timestamp is recorded for the run, across the run row (created, updated), its
     events, its cards (updated), its workers (registered, heartbeat), and its dead letters;
  4. every one of those recorded timestamps parses as a timestamp and is **strictly earlier** than
     the host's last boot — a timestamp equal to or later than the boot, or one that does not parse,
     declines the proof.

  A retirement path that is not supplied with a boot-time source or a lead-record check does not
  run the proof. A run that carries a complete identity from either source is judged by that
  identity alone and never reaches the proof; an error while reading the run's recorded timestamps
  retires nothing. The proof is a **new route to a positive `dead` classification**, not a
  relaxation of REQ-005: REQ-005 still retires only on `dead`, and the proof yields `dead` only
  from a positive fact — every process that could have owned the run existed before the current
  boot — never from the absence of evidence. §C.3 records why this is not the rejected run-lifetime
  alternative (b).
- **REQ-006c** (Unwanted): The platform boot-time reader shall not report a boot time from a source
  whose read failed before the boot-time record was found. **When** a read error occurs before the
  boot-time record is found, the procfs interpretation seam shall report the boot time as
  unavailable with that read error as its cause, distinguishable from a source read to its end that
  carries no boot-time record; the platform reader then reports the boot time unavailable, which
  disables the REQ-006b proof on that host. The cause is the seam's to report — the platform reader
  keeps its available/unavailable answer and does not itself carry the cause. Once the boot-time
  record has been read in full the source need not be read further, and a read failure after that
  point does not withdraw it. A long line preceding the boot-time record shall not cause the record to be lost: on the
  procfs source, a line longer than 64 KiB ahead of the `btime` record (the `intr` line reaches
  that length on hosts with many interrupt sources) shall not prevent the record from being found.
  The procfs interpretation shall be reachable from a build-tag-free seam that takes the source's
  content, so its error and long-line behaviour is exercisable from any host — the reader that uses
  it sits behind a build constraint that excludes darwin, the only pre-merge platform (§A.2).
- **REQ-010** (Ubiquitous): Retirement shall be a status transition on the existing `runs` row
  (`active` → `retired`) that preserves the row and appends a `run.retired` event; it shall not
  delete the row.
- **REQ-010b** (Ubiquitous): Every `run.retired` event, on every retirement path, shall record
  which proof established the `dead` classification, as a `basis` value that is exactly one of
  `stamp` (the run row's own owner stamp probed dead), `peer` (the REQ-006 `role='lead'` peer
  identity probed dead), or `boot` (the REQ-006b boot proof). The existing `classification` key and
  its value are unchanged; `basis` is an added key. The boot proof is an inference from time, not a
  probe of a process, so an operator auditing a retirement must be able to tell from the event
  alone which of the three fired.

### Failure surface

- **REQ-007** (Event-driven): **When** reconciliation leaves two or more active runs, the resolver
  shall still fail closed with `AMBIGUOUS_FACTORY`, and the failure shall name each remaining active
  run id together with its owner classification.
- **REQ-014** (Unwanted): The fail-closed semantics of run resolution shall not be weakened — zero
  active runs shall remain `NO_ACTIVE_FACTORY`, and two or more active runs after reconciliation
  shall remain `AMBIGUOUS_FACTORY`. Selecting "the newest" or "the first" active run is prohibited.

### Operator surface

- **REQ-008** (Event-driven): **When** the operator invokes the factory run maintenance command, the
  CLI shall report every active run with its owner classification, and shall retire a run only when
  the operator names that run id explicitly.
> **REQ-009 is retired into REQ-005 and the number is deliberately not reused.** It stated the
> live-run protection a second time, for the operator command only, and the narrower copy is the
> D14 defect itself. The gap in the numbering is kept rather than closed by renumbering, because
> renumbering would break every traceability reference already written against REQ-010..REQ-014.

### Doors, tests, platforms

- **REQ-011** (Ubiquitous): Run recording and run resolution shall satisfy REQ-002b at **every**
  lead entry door — `moai cc -f`, `moai glm -f`, `moai codex -f`, and `moai codex -f --spawn` —
  covering all five `registerFactoryLaunchPending` call sites enumerated in §A.1 across all three
  launch shapes. A door is covered either by an executed observation or by an explicit statement
  that its shape matches an executed one; an unnamed door is not a covered door.
- **REQ-012** (Ubiquitous): Every test or reproduction that opens factory state shall isolate
  `HOME`, `MOAI_HOME`, and the launched-binary path (`MOAI_CLAUDE_BIN`), and shall use a project
  directory outside this repository's worktree set, because `homestate.CanonicalProjectRoot`
  converges a linked worktree onto the primary checkout — a test that omits this isolation mutates
  the developer's live factory state rather than a fixture.
- **REQ-013** (Ubiquitous): The owner-liveness predicate and the reconciler shall be exercised on
  darwin, linux, and windows via the three-OS `test-integration` job, and the SPEC shall state
  which results exist before the card integrates and which arrive after. The exercise shall live at
  `test/integration/harness/` behind `//go:build integration` — the one path that job runs — and
  shall be verified to be actually selected there, not merely present (§A.2). Additionally, the
  REQ-002b restamp shall be reachable through a **build-tag-free seam** taking an already-resolved
  `(runID, pid, fingerprint)`, so the restamp behaviour is exercisable from any host. A restamp
  written only inside `launch_exec_windows.go` would sit behind `//go:build windows` and could not
  be exercised on darwin at all — and the pane door needing that same restamp **is** on darwin.

## §C Design Decision — the retirement mechanism

**Chosen: owner-liveness reconciliation at run-resolution time** (REQ-004). When resolution would
otherwise raise `AMBIGUOUS_FACTORY`, every active run whose owner is provably dead is retired first,
and resolution re-runs over what remains.

Why this one:

- It does not depend on the lead's cooperation. On POSIX the lead process is *replaced* by
  `syscall.Exec`, so no Go code of ours runs after the launch — there is no lead-exit callback to
  hang retirement on, and a crashed or `SIGKILL`ed lead would defeat one anyway.
- It fires at exactly the moment the defect is felt. The user-visible symptom is the worker join
  being refused; reconciling on that path fixes the symptom at its own trigger.
- It covers prevention and migration with one mechanism. Rows already accumulated are reconciled by
  the same pass that prevents new accumulation, which is what REQ-006 makes possible.
- It preserves fail-closed behaviour (REQ-014). Reconciliation only ever *reduces* the active set by
  removing provably dead owners; ambiguity that survives it is real ambiguity.

An explicit operator command (REQ-008, bound by REQ-005) ships alongside it, not as an alternative: a run whose
owner is `indeterminate` is never auto-retired by design (REQ-005), so without a deliberate operator
surface such a row would be permanently unremovable. The command shares the one liveness predicate;
it is a second entry to the same mechanism, not a second mechanism.

### Alternatives rejected

| Alternative | Why rejected |
|---|---|
| **(a) Retire on lead exit** (SessionEnd hook, or a deferred call in the launcher) | On POSIX `syscall.Exec` replaces the process, so nothing of ours runs after launch — there is no exit point to hook in the launcher. A Claude Code SessionEnd hook could cover clean exits only: a crash, a `SIGKILL`, or a disabled-hooks session leaves the row active, which is the defect unchanged. Best-effort cleanup that silently misses the crash case is worse than none, because it makes the residue look like a rare anomaly rather than a systematic leak. |
| **(b) Run lifetime / TTL** | Wrong predicate for this domain. A lead session legitimately idles for hours between dispatches, so any TTL short enough to unblock a user promptly is short enough to retire a live lead's run — a direct REQ-005 violation. A TTL long enough to be safe does not unblock the user at all. Time is not evidence of death here; process identity is. |
| **(c) Explicit operator command alone** | Necessary but insufficient. The operator meets this defect as an opaque `AMBIGUOUS_FACTORY` refusal; requiring them to already know a maintenance verb exists puts the cure behind the same knowledge gap as the disease. Retained as the `indeterminate`-residue surface (REQ-008), demoted from primary. |
| **(d) Widen `ResolveActiveRun` to pick one of several active runs** (newest, or the one matching `MOAI_KANBAN_ID`) | Makes the symptom disappear without touching the cause, and silently joins a worker to an arbitrary run. Explicitly forbidden by the card constraint and by REQ-014: two active runs IS ambiguous, and the error is correct. |
| **(e) Owner identity from the broker `peers` row alone, with no `runs` column** | Simpler (no schema change) and it is the REQ-006 fallback — but it cannot see a run whose record was written and whose launch then failed before the peer was registered (`recordFactoryRunStart` at `internal/cli/cc.go:191` precedes `execOrSpawnClaude`). That path leaves an `indeterminate` row for as long as the host stays up — the REQ-006b boot proof can reach it only after the next reboot — i.e. a second generator of the same defect. The stamped column closes it; the peer lookup is kept as the legacy-row fallback where the column is empty. |
| **(f) One owner stamp with no per-platform branch** (the first draft's design) | Rejected on measurement, not on taste: §A.1 shows the launcher and the session are the same process on POSIX and two different processes on Windows. A single unconditional stamp therefore names a process that can die while the session lives, which retires a live run — a direct REQ-005 violation — and puts the primary and fallback identity sources on different processes. REQ-002b makes the branch explicit instead. |

### C.1 Sub-decision — which identity the run row stamps

Rejecting (f) settles that *some* platform handling is needed; it does not settle *which*. Two
options were put forward. Both preserve the live-run boundary, so the choice is made on what each
actually buys.

**Chosen: (a) — stamp the session process, one identity per run.** On a spawn-shaped launch the row
carries the child's PID and fingerprint, matching what `launch_exec_windows.go:50,55` already
registers on the peer; on a replace-shaped launch the launcher stamp already is that identity. One
identity per run, and REQ-002 and REQ-006 name the same process by construction (REQ-002b).

**Rejected: (b) — record both launcher and child identities, owner live when EITHER is alive.**
Three reasons, in order of weight:

1. **It does not remove the platform branch; it relocates it.** A child identity exists only where a
   child exists — on a replace-shaped launch there is no second process to record — so the *writer*
   stays platform-conditional either way. (b)'s stated benefit is not achieved, and the branch ends
   up in two places (writer and predicate) rather than one.
2. **The second identity adds no discrimination.** The cases it is meant to tolerate are already
   tolerated by (a). Orphaned child, launcher killed: (a) stamps the child, which is alive →
   `live`, correct. On a replace-shaped launch the two identities are the same process, so the
   second column is degenerate there. No case was found where (b) classifies correctly and (a) does
   not.
3. **It breaks the single-source agreement that REQ-002b exists to establish.** The `role='lead'`
   peer registers exactly one identity, so a two-identity column cannot be filled from the REQ-006
   fallback on a legacy row. That needs a further rule for partially-known owners — more surface,
   guarding a case reason 2 shows is empty.

(b)'s genuine merit is that OR-ing two liveness answers biases harder toward `live`, and biasing
toward `live` is the correct direction (REQ-003b). But (a) already reaches `live` in every case
examined, so the bias buys no additional safety here — only a second column and a partial-data rule.

Should the run phase find a concrete case where (a) misclassifies a live session as dead, that
finding reopens this sub-decision rather than being worked around: (b) is the standing fallback, and
the case is the evidence that would justify its cost.

### C.2 Sub-decision — what the pane door stamps

The pane door raises the question the lead posed: what identity can this door **honestly** stamp,
given the launcher exits immediately?

**It can stamp the pane's process, and that answer needs no invention** —
`defaultCodexSpawnPaneIdentity` (`codex_launcher.go:242`) already resolves exactly the value
required: it polls `tmuxPanePID` until `homestate.ProbeProcessIdentity` reports the pane process
`live` with a non-empty fingerprint, within a 2-second deadline, and hands that pair to
`registerFactoryLaunchPending`. Those are the same two values REQ-002b wants, already probed live,
already on the happy path. The pane shape therefore takes the **same rule** as the spawn shape —
restamp with the session identity — and the only thing that differs is *where* the identity comes
from. This is why REQ-002b is generalized to "does not replace the launching process" rather than
given a third clause: one behaviour, three doors.

**The ordering is the part that needs stating.** `recordFactoryRunStart` (`codex_launcher.go:515`)
runs before `runCodexLaunch` opens the pane, so at record time no session exists to name. The
record-time stamp is therefore the launcher, correct only for the instant it describes, and the
restamp at the pane-identity site is what makes it true. A launch that dies between those two points
leaves a run stamped with a dead launcher — which reconciliation correctly reaps, since no session
was ever reachable.

**When no identity is obtainable, refuse — do not stamp a process known to be dying.** That is
REQ-002d, and it matches what the code already does on its failure path: on identity error
`defaultCodexSpawnLaunch` kills the pane (`codexSpawnCleanupPaneFn`) and returns the joined error.
REQ-002d adds the run-state half of that refusal — the run must not be left carrying the launcher's
identity — which the current code has no reason to do because it has no run stamp to clean up yet.

Rejected for this door: **deferring the stamp and leaving the run unstamped**. An unstamped run of a
live lead is `indeterminate` under REQ-006 — the REQ-006b boot proof cannot hold for a run whose
activity follows the current boot — so it is never auto-retired while its host stays up (REQ-005)
and accumulates as exactly the residue this SPEC exists to drain — it converts a live-run hazard
into a blocking one that only a reboot clears, rather than removing it.

### C.3 Sub-decision — the boot proof is not a run lifetime (added v0.13.0, card t1169)

Card t1168 found five identity-less `active` rows on a real host that REQ-006 as first written left
`indeterminate` forever: no stamp, no broker, so no identity to probe. REQ-005 correctly refused to
retire them, and the operator command refused them too, so every `-f` lane join failed
`AMBIGUOUS_FACTORY` with no remedy short of hand-editing the database
(`.moai/reports/t1168/verdict.md`, Claim and root-cause challenge). The shipped repair
(`372c1bb0b`) classifies such a row `dead` under REQ-006b.

Rejected alternative (b) in §C says "time is not evidence of death here; process identity is". The
boot proof does not contradict it, and the difference is the whole of why it is admissible:

- A **lifetime** compares a run's age against a threshold. An idle live lead exceeds any threshold,
  so a lifetime retires live runs.
- The **boot proof** compares every recorded timestamp against one instant — the host's last boot —
  and a process that existed only before that instant cannot be running after it. It says nothing
  about age: a run started one second before a reboot is proven dead, and a run idle for a week
  since the last boot is not.

What it does share with a lifetime is that it reads clocks rather than processes, and that is why
REQ-006b demands every premise, why a row with any identity never reaches it, and why REQ-010b
makes each boot-proven retirement distinguishable after the fact. Its residual exposure — a boot
clock recorded before time synchronisation — is carried in §F.

## §D Gaps carried from the reproduction

Each Gap in `.moai/reports/t1107/verdict.md` §4 is dispositioned here:

| Verdict gap | Disposition in this SPEC |
|---|---|
| Not reproduced with a real `claude` session (a stub was used) | Unchanged. The defect and the fix both live in the state layer *before* `exec`, so the run phase also verifies with a stub binary. Recorded as residual risk (§F), not covered. |
| GLM and Codex lead doors read but never executed | **Covered** — REQ-011 plus AC-011/AC-012, which are satisfied only by a recorded invocation and its captured `runs` output. |
| Lead-process-death behaviour never measured (the stub lead exits immediately) | **Covered** — REQ-003/004/005 plus AC-004/AC-005, which measure both the dead-owner and live-owner branches with a controlled long-lived process. |
| Accumulated scale in real use never measured | **Partly covered.** REQ-008's report surfaces the count on demand; this SPEC does not measure the existing population of any real installation, and does not need to — reconciliation is per-resolution, not a one-shot sweep sized to a population. |
| Windows never measured (darwin only) | **Covered, with the timing named.** REQ-013 places the exercise where the verified three-OS job runs it (§A.2). Stated plainly: the originating reproduction is darwin-only, darwin is the only platform verified pre-merge, and the ubuntu / macos / windows results arrive on the develop-push run after this card integrates. |

## §E Exclusions

### Out of Scope — the kanban backlog runtime record

- `kanban.RecordFactoryRunStart` also writes a `TodoRuntimeRun` into the backlog store. Measured:
  that record carries no status field (`internal/kanban/todo_runtime.go:20`) — it is a provenance
  seed, not a lifecycle state — so it has no `active` to leave and is not reconciled here.

### Out of Scope — worker roster and peer lifecycle

- Retiring, expiring, or reaping `workers` rows and broker `peers` rows. This SPEC reads the lead
  peer identity (REQ-006) and writes nothing to it. Worker-slot lifecycle is a separate axis.

### Out of Scope — run-id granularity

- Widening `NewRunID()` beyond second granularity. Second granularity is what makes the same-second
  case collapse harmlessly; changing it would widen the defect's trigger, not narrow it, and the fix
  here is independent of id granularity.

### Out of Scope — files owned by sibling cards

- `internal/cli/factory_lane_handoff*.go`, `internal/factorymsg/handoff*.go`,
  `internal/cli/mcp_codex.go`, `internal/config/defaults.go` (card t1082), and
  `internal/hook/factory_messages*.go` (card t1109). If the run phase finds it needs one of these,
  it halts with a blocker report naming the file rather than editing it.
- **What was actually measured about the overlap, and what was not.** In *this* tree at
  `bb5b8f9d1`, `ls internal/cli/factory_lane_handoff* internal/factorymsg/handoff*
  internal/hook/factory_messages*` matches nothing — those files do not exist here, because they
  live on the sibling cards' branches. This tree therefore **cannot** establish non-overlap; an
  absence here means "not measured", not "does not overlap". The zero-overlap figure is the team
  lead's measurement, taken against t1082's own worktree and reported on 2026-09-23; it is cited
  here as that, with its source, and not re-presented as a measurement of this tree.
- **The residual that the zero does not cover.** `internal/factorymsg/store.go` is the only
  non-test file in `internal/factorymsg/` in this tree, and this SPEC modifies it while t1082
  adds `handoff*.go` to the same package. Different files, so no textual merge conflict is
  expected — but a same-package semantic clash at integration is a live possibility and is named
  here rather than folded into the zero.

### Out of Scope — release activities

- No release cut, push, or PR creation. Integration follows the git-flow lane protocol.

## §F Residual risk

- **Cross-platform verdict timing.** Only darwin is verified before this card integrates — the
  lane's own local run. The ubuntu / macos / windows results arrive from the `test-integration` job
  on the develop push that follows integration (§A.2 carries the wiring citation and the confirming
  run). A named deferral with a verified landing place, not a gap the ACs pretend to close.
- **Windows restamp is designed, not measured here.** REQ-002b's restamp is authored from a source
  read of `launch_exec_windows.go`; no windows host was exercised during plan phase. The run phase's
  AC-016 covers the logic under test, and the platform verdict follows the timing above.
- The stub-binary verification shape is inherited from the reproduction: a real session's timing
  around the launcher's deferred restore paths is not exercised.
- Same-package semantic clash with t1082 on `internal/factorymsg/store.go` — see §E.
- `ProbeProcessIdentity` returns `indeterminate` on any probe error. A host where probing routinely
  fails would never auto-retire a run that carries an identity (only an identity-less run the
  REQ-006b boot proof establishes, which consults no probe); this degrades to today's behaviour plus an explicit
  operator surface, and never to retiring a live run. That last clause is now true **by
  construction** rather than by assertion: REQ-005 binds every retirement path, the operator command
  included, so there is no longer a surface that accepts `indeterminate` (the D14 defect).
- **Pane-door restamp is designed, not measured here.** The identity values REQ-002b needs are
  already resolved live by `defaultCodexSpawnPaneIdentity`, but no `moai codex -f --spawn` launch
  was executed during plan phase; AC-011 makes that an executed observation in the run phase.
- Linux one-second fingerprint resolution (REQ-003b) leaves a narrow same-second PID-reuse case
  indistinguishable. It resolves toward `live`, so the failure mode is a surviving stale run — which
  has the `--factory-run` escape — never a retired live one.
- **Boot proof on a stale boot clock (added v0.13.0).** The REQ-006b proof trusts the recorded
  timestamps and the reported boot time to be on one clock. On a host without a battery-backed
  real-time clock, a run recorded after boot but before time synchronisation corrects the clock can
  carry timestamps that read earlier than the corrected boot instant, and the proof would then
  retire a live run. The exposure is bounded to rows with **no complete identity** (REQ-006) and
  **no broker file** (REQ-006b premise 2): current builds stamp every run they record with a
  complete identity, except the REQ-002d refused launch, which is dead by construction, and a
  record-time fingerprint failure, which writes a PID without a process start (the partial stamp,
  REQ-006) — so it reaches a lead still running from a binary older than this SPEC's owner stamp,
  or a partially-stamped lead whose broker file was never created, on such a host. The `boot` basis
  (REQ-010b) is what lets an operator find such a retirement afterwards.
  Source: `.moai/reports/t1168/verdict.md` Residual-risk.
- **Known pre-existing defect, out of this amendment's scope — `--retire` on an already-retired
  run appends a second `run.retired` event.** `RetireRunIfDead` classifies over every status, so on
  a run already `retired` whose owner still probes `dead`, the `active → retired` update changes no
  row while the event is appended again. REQ-010 and REQ-010b read as covering only a transition's
  event, so this is a defect against them, but it predates v0.13.0 and the boot proof does not
  change it. Recorded (plan-audit iter-1 of card t1169, D11) rather than fixed: a separate card is
  the lead's call. No v0.13.x acceptance criterion depends on it either way.
- **The linux and windows boot-time sources were cross-compiled, not exercised, before t1168
  merged.** `TestSystemBootTimeIsInThePast` ran on darwin only; the procfs and windows readers were
  built for their targets and never run. REQ-006c closes the procfs interpretation on any host
  through its build-tag-free seam; the windows reader and the live procfs read remain post-merge
  CI evidence under the §A.2 timing.

## §G Accepted debt — audit findings recorded and dispositioned

Recorded so that a later audit reads these as decided rather than overlooked. D7-D10 were raised at
plan-audit iter-1 and declined on the team lead's routing; D20-D21 were raised at the Tier L audit
and retained at plan time, and both are now **closed** — each by the very condition it was recorded
with, met in the run phase and evidenced below rather than asserted:

| Finding | Disposition and why |
|---|---|
| **D7** — REQ ids are not monotonically ordered (`REQ-002b`, `REQ-002d`, `REQ-003b` interleave, and `REQ-009` is now a gap) | The suffix form keeps each requirement adjacent to the one it refines, and the `REQ-009` gap is deliberate (see §B); renumbering would break the traceability already written into `acceptance.md` §D.3 for no reader gain. |
| **D8** — verification-verb classification in the AC matrix is uneven | The matrix column is a reader's index, not a contract; each AC's own Given-When-Then carries the binding form. |
| **D9** — REQ-002 states a transaction constraint (an implementation detail) at the requirement layer | Deliberate: an owner stamp written outside the row's own transaction can be lost against the row it describes, which is an observable behaviour, not an internal choice. |
| **D10** — one measurement is attributed in both §A and §A.1 | The duplication is between a summary and its detail section; removing either costs a reader the attribution at the point of use. |
| **D20** — the RED-now evidence ledger in `acceptance.md` §C.1 carries a document-level tree pin of `bb5b8f9d1`, now several commits behind HEAD | **CLOSED.** The condition recorded with it — "any code-side change under `internal/`, `.github/`, or `test/`" — **has fired**: the implementation commit `eaa3322a1` changed both `internal/` and `test/`, and `git diff --stat bb5b8f9d1 HEAD -- internal/ .github/ test/` now reports 17 files changed (+1821/−13). Closing rather than re-pinning, because the condition was written while the ledger was a **pre-implementation baseline awaiting the run phase**, where a code change meant the RED cells might no longer be red for the reason claimed. Once the implementation lands, those cells have discharged their job: the pin is then a correct historical record of the tree the RED was measured against, and being behind HEAD is what a historical pin is *for*. Re-pinning to HEAD would be actively wrong — it would assert a RED ledger against a tree where the cells are, by design, green. The pin stays at `bb5b8f9d1` and is read as history, not as a current measurement. |
| **D21** — the premise that the tmux pane identity names the session process is not measured inside this SPEC | **CLOSED** by the condition recorded with it: "AC-011 closes it by execution in the run phase". AC-011 executed, and the evidence is in `progress.md` §E.2 — the pane door (`moai codex -f --spawn`) recorded run row `tltbl8` carrying `68638 / 1790159084.722334`, with tmux itself reporting `PANE %1 pane_pid=68638`: the same process. The premise is therefore now measured **inside this SPEC**, in the run phase, which is precisely what the retention was conditioned on; the external tmux 3.6a corroboration is no longer load-bearing. `plan.md` M5's blocker-report requirement was never reached, since the sandbox did provide a tmux server. |

## §H History

- 2026-09-23 — v0.1.0 — manager-spec — initial plan-phase draft (card t1107, Tier M, class C),
  authored from the reproduced defect in `.moai/reports/t1107/verdict.md`.
- 2026-09-23 — v0.2.0 — manager-spec — plan-audit iter-1 FAIL revision. D1: the "one stamp, no
  per-platform branch" premise was measured false on Windows and removed; the owner is respecified
  as the **session process** with an explicit restamp (REQ-002b) and a new rejected
  alternative (f). D2: REQ-013 respecified against the measured CI shape, with the
  post-integration deferral named in §F. D4: the zero-overlap claim restated as what this tree
  actually shows, with the lead's measurement cited to its source. D5: three GEARS aspect labels
  corrected. D6: linux fingerprint resolution folded in as REQ-003b. D7-D10 recorded as accepted
  debt in §G.
- 2026-09-23 — v0.3.0 — manager-spec — iter-2 follow-up. §A.2 added: the three-OS `test-integration`
  green path is verified (wiring citations plus confirming run `35802361895`), together with the
  timing table showing darwin as the only pre-merge platform. AC-013 is split into a release-blocking
  pre-merge leg — which asserts the test is actually *selected*, not merely present — and a
  non-gating post-merge leg. §C.1 added: the D1 sub-decision recording option (a) chosen and option
  (b) rejected with reasons.
- 2026-09-23 — v0.4.0 — manager-spec — plan-audit iter-2 FAIL revision (verdict pinned to
  `c1ae8ff5e`; `320cdeb90` was an unexamined successor). **D11**: §A.1 rewritten around the five
  `registerFactoryLaunchPending` call sites in three launch shapes; REQ-002b generalized from
  "spawns a child" to "does not replace the launching process", so it covers the tmux **pane** door
  (`codex_launcher.go:230`) as well; REQ-002d added for the refuse-when-no-identity path; §C.2 added
  recording what the pane door stamps and why; the two `codex_direct_*` sites are asserted covered
  by shape match rather than left implicit; REQ-011 restated over all five sites. Scope EXTENDED by
  operator decision — the auditor's suggested narrowing was rejected. **D14**: the live-run
  invariant is now stated once (REQ-005) binding every retirement path including the operator
  command; the narrower `REQ-009` is retired into it and its number left as a deliberate gap.
  **D12**: AC-013 leg 2's green condition changed from "reports a result" (which a failing run also
  satisfies) to `success` on all three jobs — verified still-present at `320cdeb90` rather than
  assumed fixed. **D15**: REQ-013 gains the build-tag-free restamp seam, so AC-016 can assert the
  spawn and pane shapes from darwin. **D17**: the four dead `002c` / `013b` tokens removed.
  **D13**: R-02's transcribed stdout corrected from `*(empty)*` to `0`.
- 2026-09-23 — v0.5.0 — manager-spec — plan-audit iter-3 revision (D11-D17) plus the **tier change**.
  **Tier M → Tier L**: AC-017 brings the acceptance-criterion count to 17, over the Tier M ceiling
  of 16, so `tier:` is raised and `design.md` + `research.md` join the artifact set. The tier change
  is pre-authorized by the operator; nothing was merged, dropped, or renumbered to avoid it
  (`acceptance.md` §D.5). The plan-auditor PASS threshold rises 0.80 → 0.85 with the tier.
  **D14 restated positively**: REQ-005 now says retirement happens **only on a positive `dead`
  classification**, rather than listing states to reject — an enumeration is correct only for the
  classification set alive when it was written and silently begins retiring any state added later.
  AC-017 is the criterion that separates the two, and it is the only one that does: a reject-list
  mutant satisfies AC-005, AC-006 and AC-010 unchanged. The consequence recorded with it is
  host-wide, not an edge case — where the probe fails consistently every run reads `indeterminate`,
  so the old command rule would have retired live sessions across that entire host.
  **Audit-withdrawal note**: the auditor withdrew its `320cdeb90` observations (reading a second
  tree during an audit pinned to `c1ae8ff5e`) and declined to judge that delta. D11, D12 and D17 are
  therefore treated as fully open, and the evidence recorded for each in this revision is this
  SPEC's own measurement, not partial credit carried over from the audit.
- 2026-09-23 — v0.6.0 — manager-spec — **D18** repair, `plan.md` only (+18/−4). §D said the restamp
  seam had "the two call sites" while M1 simultaneously assigned a restamp obligation to
  `codex_direct_windows.go:24`; the file set now names **three** seam call sites
  (`launch_exec_windows.go` spawn, `codex_direct_windows.go` spawn, `codex_launcher.go` pane), and
  the single exclusion bullet became two so that `codex_direct_windows.go` is stated **in** the set
  and `codex_direct_posix.go` **out** of it on replace-shape grounds — the two files are never named
  in one sentence in §D again. The count defect was §D-local: §F already said three in two places
  (`:108`, `:117`), so the document disagreed with itself. Commit `cecc94ee9`.
- 2026-09-23 — v0.7.0 — manager-spec — two authorized fact corrections. `acceptance.md:245` read
  "All sixteen criteria" against 17 elsewhere — a stale numeral left by the AC-017 addition.
  `research.md` §J item 6 asserted `plan-audit-iter3.md` was absent from disk and unrecoverable;
  that was false — the file exists (38,933 bytes, 2026-09-23 13:52, carrying the required header
  line). The wrong claim was a timing artifact: the directory was read roughly two minutes before
  the auditor finished writing the report into it, so the absence was real when observed and stale
  when written down. Commit `acb03bede`.
- 2026-09-23 — v0.8.0 — manager-spec — **D22** plus the two **D18** twins. D22: AC-016 gains a
  **second leg** asserting at source level that a restamp-seam call is present at each of the three
  non-replace call sites, with the two replace-shaped sites asserted absent from that required set,
  and mutation specified **per site** — deleting the call at any one of the three must turn the leg
  red, probed individually. D22 was the verification-layer twin of D18: no criterion could detect a
  missing seam call, which is the structural reason D18 survived three audits. The twins: `plan.md`
  M5 no longer binds the two `codex_direct_*` files in one sentence, and §A.1 replaced "covers it
  unmodified" with the distinction that an existing rule covering a door never means that door's
  call site needs no edit. AC count unchanged at 17 — the new leg lives inside AC-016. Commit
  `4b29af46f`.
- 2026-09-23 — v0.8.1 — manager-spec — record-layer repair closing the Tier L audit's **D23** and
  **D24**. D23: §H had lost three commits (`cecc94ee9`, `acb03bede`, `4b29af46f`) and the
  frontmatter `version:` still read `"0.5.0"`, so a reader of §H alone would conclude D18 was open
  and would find the AC-016 second leg in the document with no history explaining it; the three
  entries above were written and `version:` caught up 0.5.0 → 0.8.0. D24: the retained debts D20
  and D21 appeared in no artifact, so §G gained a row for each — with the evidence that makes
  carrying them defensible and the condition that re-opens each — and the §G intro was amended so
  the table no longer reads as uniformly declined. Requirement and acceptance-criterion text was
  not touched. Commit `1c872c495`.
- 2026-09-23 — v0.9.0 — manager-spec — attribution restoration plus this record. The v0.8.1
  insertion landed after the v0.5.0 bullet's last line, so that bullet's two continuation
  paragraphs — **D14 restated positively** and the **Audit-withdrawal note** — came to read as
  continuation of v0.8.0, and the note's "in this revision" pointed at the wrong revision. Both
  paragraphs are moved back under v0.5.0 unchanged. This entry also closes the one-commit lag
  v0.8.1 left behind: a §H that records every commit but its own reproduces D23 at depth one.
  Commit: this revision.
- 2026-09-23 — v0.10.0 — manager-spec — **AC-004's Given restated** plus the §G heading and
  column rename. AC-004 read "two `active` runs whose owners are **both dead** … Then the join
  succeeds, and … the run it joined as the sole `active` row": the two halves cannot both hold.
  REQ-005 retires on a positive `dead` classification, so two dead owners leave **zero** `active`
  rows and resolution correctly fails closed with `NO_ACTIVE_FACTORY` — there is then no run joined
  and no surviving `active` row. **REQ-005 is canonical and the Given was the wrong half**: the Then
  is the criterion's subject (REQ-004's dead-owner retirement followed by a successful join, per the
  §D matrix row), it is satisfiable only when exactly one owner is dead, and three artifacts already
  describe that shape — `progress.md` §E.2 F1, the AC-011 door exercise (`tltbl4 retired` against a
  dead owner, `tltbl8 active` against a live pane owner), and §A.1's own claim that "AC-004/AC-005
  … measure both the dead-owner and live-owner branches". So the Given now states one dead owner and
  one live one, and the Then stands unchanged. The all-owners-dead behaviour is real and keeps its
  guard: `TestResolveActiveRunBothOwnersDead` is untouched, and AC-004 now names AC-009's
  fail-closed leg as where that shape is carried. **Four audit rounds did not catch this** because
  each examined AC *form* — mutation direction, mechanical checkability, vacuous-green shapes —
  without once executing a Given against its own Then; form review cannot see a criterion that is
  well-shaped and self-contradictory. Separately, §G's heading and column still read "declined"
  while its amended intro and the D20/D21 rows say **retained**; both now read "dispositioned" /
  "Disposition and why", which accommodate declined and retained findings alike. No row body, no
  other criterion, no requirement, and no design text was touched; counts stay REQ 16 / AC 17.
  Commit: this revision.
- **Provenance correction (iter-2, D17).** The two retired strings quoted in this bullet —
  `REQ-002c` and `REQ-013b` — appear here as **quotations of removed text, not as live
  references**; they resolve to nothing, which is the point being recorded. The v0.2.0 entry above
  previously paired each with its surviving sibling, and the v0.3.0 work described "merging
  REQ-002c into REQ-002b" as a consolidation. That description was wrong and is corrected here:
  both retired strings were
  **intermediate drafting states within a single uncommitted editing session** and never existed in
  any committed tree — `bb5b8f9d1` included. The "merge to stay at 16" was therefore bookkeeping
  rather than a real consolidation, and it left the dead tokens D17 found. The consolidation
  performed in v0.4.0 (REQ-009 into REQ-005) is of a different kind: it removes a **shipped**
  duplicate statement of one invariant whose two copies had measurably drifted, which is the D14
  defect itself.
- 2026-09-24 — v0.11.0 — manager-spec — **§G D20 and D21 closed** plus the **R-06 pattern repair**,
  all three because each item's own recorded condition had already been met and the record had not
  followed. **D20**: its re-opening condition ("any code-side change under `internal/`, `.github/`,
  or `test/`") fired at `eaa3322a1`, which changed `internal/` and `test/` —
  `git diff --stat bb5b8f9d1 HEAD -- internal/ .github/ test/` now reports 17 files changed
  (+1821/−13). Closed rather than re-pinned: the condition was written while the ledger was a
  pre-implementation baseline, where a code change could invalidate a RED cell; once the
  implementation lands the cells have discharged their job and the pin is history, so re-pinning to
  HEAD would assert a RED ledger against a tree where the cells are by design green. **D21**: its
  closing condition ("AC-011 closes it by execution in the run phase") was met and recorded —
  `progress.md` §E.2 AC-011 shows the pane door's row `tltbl8` carrying
  `68638 / 1790159084.722334` with tmux reporting `PANE %1 pane_pid=68638`, the same process — so
  the premise is now measured inside this SPEC and the debt is no longer live. **R-06**: the cell's
  selector spelled one space where the gofmt-aligned source has three
  (`internal/cli/factory_handoff_recover.go:64`), so it returned `0` both before **and** after
  implementation and could never flip — an instrument-manufactured zero inside the SPEC's own
  verification apparatus, the same shape §C.1 already rejects for `go test -run` selectors. Replaced
  with the whitespace-tolerant `grep -cE 'Use:[[:space:]]+"runs"'`, verified in all three
  directions: `0` at the ledger's own document-level pin `bb5b8f9d1` — the tree every RED-now cell
  was measured against, so the recorded RED value stays honest — `1` at HEAD (so the cell actually
  flips), and `0` against `internal/cli/factory.go` (so the pattern is not matching everything). The
  honesty leg was also re-measured at `9b1805a67` and reads `0` there too, so both trees agree and
  the conclusion is unaffected by which one is read. No requirement, no other acceptance
  criterion, and no design or research text was touched; counts stay REQ 16 / AC 17 and
  `status:` stays `completed`. Commit: this revision.
- 2026-09-24 — v0.12.0 — manager-spec — **three wording defects** found by the t1107 audit, all of
  the same family as v0.10.0's AC-004 repair: a criterion that is well-shaped and still says
  something other than what it means. **A1 — AC-009's premise widened.** AC-004 delegates the
  all-owners-dead shape to "the fail-closed leg AC-009 carries", but AC-009's own premises were only
  "zero `active` runs" and "two `active` runs with live owners" — neither is "two `active` runs whose
  owners are all dead", a third premise in which both rows retire and the remaining count *reaches*
  zero. The behaviour was never missing (`TestResolveActiveRunBothOwnersDead`,
  `internal/factorymsg/factory_run_retire_test.go:126`); the coverage text was, so a reader following
  AC-004's pointer arrived at a criterion that did not name the shape. The first `Given` now reads
  "zero `active` runs remaining after reconciliation — whether none existed to begin with or every
  owner was classified `dead` and retired", which makes the delegation land. No AC id was created or
  removed. **S2 — AC-007's third `Then` clause restated.** It read "and the join no longer fails with
  `AMBIGUOUS_FACTORY`", which is literally true and reads as "succeeds": followed to its consequence
  under REQ-005, both dead owners retire, zero `active` rows remain, and the join fails closed with
  `NO_ACTIVE_FACTORY`. A negation that a reader completes into the wrong positive is the exact
  misreading AC-004 was repaired for, so the clause now states the outcome — fails closed with
  `NO_ACTIVE_FACTORY`, landing on AC-009's leg — rather than what it is not. The other two clauses and
  the migration-path sentence are unchanged. **F1 — R-06's honesty-check attribution corrected.** The
  v0.11.0 entry attributed the `0` reading to `9b1805a67`, but the ledger that cell lives in binds a
  document-level pin of `bb5b8f9d1`, so the paragraph cited a tree the ledger does not. Independent
  re-measurement (`.moai/reports/t1107/verify-debt-closure.md` §C3) reads `0` / exit 1 on **both**
  trees under the new pattern, so the recorded RED value and the conclusion stand; only the
  attribution was wrong, and it now names the ledger's own pin with the second tree stated beside it.
  The `1`-at-HEAD leg and the `internal/cli/factory.go` negative control are untouched. Not repaired:
  the audit's F2 (an over-stated §H placement rationale resting on a reattachment hazard) — a sweep of
  the SPEC directory finds no such rationale sentence in any tracked artifact and `90e952b80` carries
  no placement reasoning, so the claim existed only in a repairing agent's return message and §H's
  ordering is left exactly as it stands; inventing the sentence in order to correct it would
  manufacture the defect. No requirement, no design, and no research text was touched; counts stay
  REQ 16 / AC 17 and `status:` stays `completed`. Commit: this revision.
- 2026-09-25 — v0.13.0 — manager-spec — **in-place amendment, card t1169**, bringing the SPEC's
  text up to the **boot proof** card t1168 shipped (`372c1bb0b`, merged into develop) — behaviour
  the literal REQ-006 did not allow, recorded as the first Residual-risk item of
  `.moai/reports/t1168/verdict.md`. **REQ-006** now classifies an identity-less owner `dead`
  exactly when the new **REQ-006b** boot proof holds and `indeterminate` otherwise; REQ-006b states
  the four premises one by one, each testable, with every unestablishable premise declining the
  proof. REQ-005 is untouched in text and in spirit: the proof is a new route to a positive
  `dead`, not a relaxation. **REQ-006c** binds the boot-time reader to report unavailability with
  its cause and not to lose `btime` behind a long `intr` line — the shipped procfs reader never
  consults its scan error and uses the scanner's default 64 KiB token limit. **REQ-010b** adds a
  `basis` key (`stamp` / `peer` / `boot`) to the `run.retired` payload, the decision the lead
  delegated to this lane and taken as **record it**, because the boot proof reads clocks rather
  than probing a process. §C.3 records why the proof is not the rejected lifetime alternative (b);
  §F gains the stale-boot-clock exposure and the cross-compiled-only platform readers. Acceptance
  gains AC-018 (boot proof and every declining premise), AC-019 (reader error and long-line paths)
  and AC-020 (`basis` per proof). Counts: REQ 16 → 19, AC 17 → 20 (22 AC identifiers counting
  AC-015a/b), inside the Tier L ceilings of 25. `status:` moves `completed → in-progress` under the
  amendment transition (`spec-frontmatter-schema.md` § Status Enum and § Status Transition Ownership
  Matrix), with `amendment_of:` and the Amendments record below. Commit: `ae61ccb0f`.
- 2026-09-25 — v0.13.1 — manager-spec — **plan-audit iter-1 revision, card t1169** (verdict FAIL
  0.80, `.moai/reports/t1169/plan-audit-iter1.md`). **D1 closed**: every statement that an
  identity-less run is always `indeterminate` now reads "`dead` when the REQ-006b boot proof holds,
  `indeterminate` otherwise" — `design.md` §E.1 lattice, §F fallback note, §I failure table and the
  §C.1 / §D stamp-loss notes, `plan.md` M2, `spec.md` §C (alternative (e)), §C.2 and §F, plus
  `acceptance.md` AC-006's Given and a dated note in `research.md`; the Amendments scope row now
  includes `design.md`. **D2 closed**: REQ-006 defines "owner stamp" (`lead_pid ≥ 1`) and "complete
  identity", names the partial stamp and why routing it to the proof is sound, §F's exposure
  sentence is corrected, and AC-018 gains a partial-stamp leg on each side of premise 2 with
  mutant 6 (a partial stamp routed to the peer) to separate them. **D3**:
  M7 item 3 obliges every declining leg to assert the `indeterminate` classification. **D4**:
  AC-020's operator leg is fixed on a boot-proven run asserting `"basis":"boot"`, with the
  constant-`stamp`-on-operator-path mutant named. **D6**: REQ-006c's read scope and its reporter
  (the seam) now match AC-019 (ii) and M7 item 2. **D7-D10**: mutant 1's four retirement legs named
  and the identity-precedence leg separated; R-10 restricted to non-test files and re-measured;
  M7's procfs seam file made unconditional; AC-018 gains an underivable-broker-path leg. **D11**
  (pre-existing: a second `run.retired` on re-retire) recorded in §F as out of scope, not fixed.
  **D5** (§E.4 `sync_commit_sha` placeholder) is manager-docs' field and is left untouched here.
  Counts unchanged: REQ 19, AC 20 (22 identifiers). Commit: this revision.

### Amendments

| Field | Value |
|---|---|
| prior completed version | 0.12.0 |
| prior_completed_sha | `85414b3e6` (`docs(SPEC-FACTORY-RUN-RETIRE-001): sync-phase artifacts, 3-phase close (card t1107)`) — the close commit itself; `progress.md` §E.4 was backfilled to the same value by `e5f020cc3` (card t1169) |
| rationale | Card t1168 retired identity-less legacy runs through a boot proof the literal REQ-006 forbade (it required `indeterminate` whenever neither source yields an identity). The behaviour is correct and is kept; the SPEC is brought into line with it, and the two follow-ups t1168 left open — the reader's error and long-line path, and an event that does not say which proof fired — are specified for a run phase. |
| scope | `spec.md` REQ-006 rewrite, REQ-006b / REQ-006c / REQ-010b added, §C.3, three §F items, and the qualifiers in §C, §C.2 and REQ-005's rationale; `acceptance.md` AC-018..AC-020 with their evidence cells, matrix, traceability and counts, and AC-006's Given; `plan.md` milestones M2 and M7; `design.md` §C.1, §D, §E.1, §E.2, §F and §I (every statement that an identity-less run is always `indeterminate`); `research.md` one dated note under M-08's legacy-row sentence; `progress.md` t1169 plan-phase note. No Go source is changed by this amendment. |

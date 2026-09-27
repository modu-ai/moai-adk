# design.md — SPEC-FACTORY-RUN-RETIRE-001 (card t1107)

## §A What this file is, and what it is not

The system design behind the mechanism `spec.md` §C chose. It answers *how the pieces fit* —
what a run's owner **is**, where that identity is written and re-written, which package owns which
seam, how a classification becomes a retirement, and where each path exits on failure.

It does not re-argue the choice. The mechanism selection, the five rejected alternatives, and the
two sub-decisions (C.1 identity shape, C.2 pane-door stamp) are settled in `spec.md` §C and are
inputs here. It does not restate the requirements (`spec.md` §B) or the criteria
(`acceptance.md`). The codebase measurements every claim below rests on are in `research.md`; this
file cites them by ledger id (`M-nn`) rather than re-measuring.

Tree this design was authored against: `157ec351b` on `WT-factory-run-retire`.

---

## §B The owner model — a run is owned by its session process

The defect exists because a `runs` row records *that* a run started and nothing about *who is
running it* (M-04: the DDL has `run_id`, `lead_session_id`, `lead_backend`, `status`,
`manifest_json`, `created_at`, `updated_at`). `lead_session_id` names a Claude session by uuid, not
a process; nothing in the row can be probed for liveness.

The design adds one thing to the row: a **process identity** — `(pid, process_start)` — and fixes
by construction *which* process that is.

> **The owner of a run is the session process**: the process that will still be running while the
> run is meant to be active. Equivalently, it is the process the run's `role='lead'` peer already
> names (M-08: the `peers` table carries `pid` and `process_start` per peer).

The equivalence is the design's central invariant, and it is deliberate rather than incidental.
Two identity sources exist — the new `runs` column (REQ-002, the primary) and the `role='lead'`
peer (REQ-006, the legacy fallback) — and a design in which they can name *different* processes
has a state where the primary says dead and the fallback says live for the same run. Binding both
to "the session process" removes that state rather than adding a rule to arbitrate it.

Why `(pid, process_start)` and not a bare pid: process ids are reused. A bare-pid probe reports
`live` for a recycled id and retires nothing, or — worse, on the other side of the same
ambiguity — reports `live` for a *stranger's* process and keeps a dead run active forever. The
start fingerprint is the discriminator, and it is already in service in this codebase for exactly
this purpose (`homestate.ProbeProcessIdentity`, M-11).

---

## §C Where the stamp is written — and why record-time truth is only instantaneous

### C.1 The write point

`RecordRun` (`internal/homestate/runtime.go:23`) is the sole writer of `runs.status` (M-02 shows
no `UPDATE runs` / `DELETE FROM runs` exists anywhere in non-test code; M-03 is the control that
proves the search shape works). It already wraps its `INSERT ... ON CONFLICT DO UPDATE` and its
`run.started` event in one transaction. The owner stamp joins that transaction — not a
follow-up write.

The reason is observable rather than stylistic. A stamp written outside the row's own transaction
can be lost against the row it describes: the row commits `active`, the stamp fails, and the
result is an `active` run whose owner column is empty. While its lead is alive that row is
`indeterminate` (REQ-006 — the REQ-006b boot proof cannot hold for a run with activity after the
current boot), is never auto-retired (REQ-005), and is therefore exactly the residue this SPEC exists
to drain — a second generator of the defect, produced by the fix. Hence REQ-002's transaction
clause sits at the requirement layer and is recorded as deliberate in `spec.md` §G/D9.

### C.2 The record-time stamp is correct **only for the instant it describes**

At `RecordRun` the only process that certainly exists is the one calling it — the launcher. So the
record-time stamp is always the launcher's identity, and whether that is *also* the session's
identity depends entirely on what the launcher does next. Three shapes (M-01, M-09; enumerated in
`spec.md` §A.1):

| Shape | What the launcher becomes | Is the record-time stamp the session? |
|---|---|---|
| **replace** — `syscall.Exec` | *is* the session; pid and start survive the `execve` (M-12) | **Yes.** Correct at record time and stays correct |
| **spawn** — child + `Wait()` | a supervisor that outlives nothing | **No.** True only until the launcher exits |
| **pane** — tmux window, launcher returns | gone, immediately | **No.** False within milliseconds |

This is the shape of the D1/D11 defect, and it is worth naming precisely: the failure is not "the
stamp is wrong". The stamp is *right when written* and becomes wrong without anything writing to
it. Nothing in the row records that it has expired, so a reconciler reading it later probes a
process that legitimately died and concludes the run is dead — while its session runs on. That is
a **live-run retirement**, the one outcome `plan.md` §D forbids outright.

### C.3 The restamp closes it

Every door whose launching process is not the session **restamps** before the session becomes
reachable (REQ-002b). The identity it stamps is not invented — each door has already resolved it,
because each already hands the same pair to `registerFactoryLaunchPending` (M-01):

| Door | Identity available at restamp | Source |
|---|---|---|
| `launch_exec_windows.go:54` | `child.Process.Pid` + `childFingerprint` | probed live immediately after `child.Start()` |
| `codex_direct_windows.go:24` | `cmd.Process.Pid` + `start` | same shape, same point |
| `codex_launcher.go:230` | tmux pane pid + fingerprint | `defaultCodexSpawnPaneIdentity`, polled live to a 2s deadline |

The replace-shaped doors (`launch_exec_posix.go:33`, `codex_direct_posix.go:34`) restamp nothing,
because there is nothing to correct.

### C.4 The pane door's ordering window

The pane door has the widest interval between record and restamp, and the interval is *inherent*,
not a bug to tighten: `recordFactoryRunStart` runs at `codex_launcher.go:515` and the pane is not
opened until `runCodexLaunch` at `codex_launcher.go:538` (M-13). At record time there is no pane,
so there is no session identity to stamp.

A launch that dies inside that window leaves a run stamped with a launcher that is now dead, and
**that is the correct outcome** — no session ever became reachable, so reconciliation reaping the
run is right. The window is a designed property, not tolerated damage. It is the same property
rejected alternative (e) (`spec.md` §C) turns on: the stamped column is what lets a run whose
launch failed *before* any peer was registered still be classified at all.

---

## §D The restamp seam

```
// internal/cli/factory_run_owner.go   — NO build tag
func stampFactoryRunOwner(root, runID string, pid int, fingerprint string) error
```

Three properties, each load-bearing:

1. **It takes an already-resolved identity.** The seam does not probe, poll, or branch on platform.
   Each call site resolves the identity it is in a position to resolve and passes it in. All the
   platform knowledge stays where it already lives.
2. **It carries no build tag.** A restamp authored inside `launch_exec_windows.go` would sit behind
   `//go:build windows` (M-09), where a darwin host cannot compile a *call* to it — so AC-016 could
   not assert the spawn shape at all on the only platform this card is verified on pre-merge. And
   the pane door, which needs the identical restamp, **is on darwin**: `codex_launcher.go` carries
   no build tag (M-09). One tagged seam would leave the darwin-only door unable to reuse it. This
   is defect D15.
3. **It is idempotent against the row it names.** The restamp is an `UPDATE` of two columns on one
   `run_id`; calling it on a replace-shaped door with the identity already present would be a
   no-op, so "which doors call it" is a correctness-of-coverage question, never a
   correctness-of-value one.

**Failure exit.** Where a door that does not replace its process cannot obtain a live session
identity, it refuses the launch and leaves **no** run carrying the launcher's identity (REQ-002d).
This matches what the pane door already does with the pane itself: on identity error
`defaultCodexSpawnLaunch` calls `codexSpawnCleanupPaneFn` and returns the joined error
(`codex_launcher.go:218-240`). REQ-002d adds the run-state half of that same refusal, which today's
code has no reason to perform because it has no run stamp to clean up.

Deliberately rejected for this door: *defer the stamp, leave the run unstamped*. An unstamped run
of a live lead is `indeterminate` (the REQ-006b boot proof can reach it only after the next reboot),
is never auto-retired while the host stays up, and accumulates — converting a live-run hazard into a
permanent-blocking one (`spec.md` §C.2).

---

## §E The classification lattice, and why the rule is positive

### E.1 Three values, one probe

```
classify(run) -> live | dead | indeterminate
```

The input is a `(pid, process_start)` pair from the primary column, or — when the column carries
no stamp (`lead_pid` below 1) — from the run's `role='lead'` peer (REQ-006). A row with neither
yields no probe input at all; the REQ-006b boot proof, which reads clocks rather than a process,
decides it instead. The probe is the existing
`homestate.ProbeProcessIdentity`, which already returns exactly this trichotomy (M-11): a
non-live platform state short-circuits, and a live pid whose fingerprint cannot be read degrades to
`indeterminate` rather than asserting either way.

| Input state | Classification |
|---|---|
| pid live **and** fingerprint matches the recorded one | `live` |
| pid dead, or pid live with a **different** fingerprint (reuse) | `dead` |
| probe error, or no fingerprint readable for a complete identity | `indeterminate` |
| **no complete identity from either source** (REQ-006), and every REQ-006b boot-proof premise holds | `dead` |
| **no complete identity from either source**, and any REQ-006b premise false or unestablished | `indeterminate` |

**The lattice is ordered by consequence, not by confidence.** `indeterminate` sits with `live`, not
between the two, because the two errors are not symmetric: a stale run that survives has a working
escape the operator already uses (`--factory-run <id>`, confirmed working in the reproduction,
`verdict.md` §2.2 R3), while a live run that is retired has none. REQ-003b makes the same asymmetry
bind the one residual ambiguity the platform imposes — linux reads its fingerprint from
`ps -o lstart=` at **one-second** resolution (M-14), so a pid reused inside one second is
indistinguishable there and resolves `live`.

### E.2 Retirement is gated on a **positive** `dead`

```
if classification == dead { retire } else { decline }
```

not

```
if classification == live || classification == indeterminate { decline } else { retire }   // WRONG
```

The two are behaviourally identical for the three values that exist today, which is precisely why
the distinction has to be designed in rather than left to whoever writes the guard. The reject-list
form is **an enumeration of the classification set alive when it was written**: add a fourth value
— `unknown-host`, `probe-disabled`, anything — and it falls through to *retire*, silently, with no
code change and no test turning red. The positive form fails safe by construction.

This is defect D14, and its blast radius is host-wide rather than an edge case. `ProbeProcessIdentity`
returns `indeterminate` on any probe error, so on a host where probing routinely fails **every**
run carrying an identity classifies `indeterminate` — under the earlier draft's narrower operator rule, `--retire` on
that host would have retired live sessions on request.

One rule, stated once, binding **every** retirement path — the resolution-time reconciler, the
legacy-row migration pass, and the operator command alike (REQ-005). Stating it per-path is what let
the copies drift; the previous `REQ-009` was that second copy and is retired into REQ-005 with its
number left as a deliberate gap.

---

## §F The `ReconcileActiveRuns` seam and the import-cycle constraint

### F.1 The constraint, measured

`factorymsg` imports `homestate` (M-07); `homestate` imports `factorymsg` nowhere (M-06). The
direction is fixed, and Go will not permit it to be closed.

This matters because the two halves of REQ-006 live on opposite sides of it:

- the `runs` table and its schema live in **`homestate`** — and so must its writer, or a
  `runs`-table writer ends up outside the package that owns the schema;
- the `role='lead'` peer the fallback reads lives in a **`factorymsg`** broker database.

### F.2 The resolution — the fallback is a parameter, not an import

```go
// internal/homestate/factory_run_retire.go
type LeadIdentityLookup func(runID string) (pid int, processStart string, ok bool)

func (f *FactoryDB) ReconcileActiveRuns(ctx context.Context, fallback LeadIdentityLookup) (Reconciliation, error)
```

`factorymsg` supplies the closure — it already knows how to open a run's broker and read its peers.
`homestate` owns the row, the transition, and the predicate, and never learns that `factorymsg`
exists. The result value reports **both** halves of what happened: what was retired, and what was
left with each survivor's classification. The second half is not bookkeeping — it is what
`ResolveActiveRun` renders into the `AMBIGUOUS_FACTORY` message (§G).

A run whose `lead_pid` is below 1 consults the fallback; a partial stamp (`lead_pid` of 1 or more,
empty `lead_process_start`) does not. A run with neither source yielding a complete identity is
`indeterminate` and stays, unless REQ-006b proves its owner dead — every premise established, the
run's broker file absent among them.

**Rejected: move the whole reconciler into `factorymsg`.** It resolves the cycle in the wrong
direction — the package that does not own the `runs` schema becomes the package that mutates it,
and the `retired` transition ends up living apart from the DDL, the migration, and the event
writer it has to stay consistent with.

---

## §G Data model, and retirement as a transition

### G.1 Schema v2 → v3

`factorySchemaVersion` is `2` today (M-05). The change adds two columns and bumps to `3`:

```sql
ALTER TABLE runs ADD COLUMN lead_pid            INTEGER NOT NULL DEFAULT 0;
ALTER TABLE runs ADD COLUMN lead_process_start  TEXT    NOT NULL DEFAULT '';
```

`migrateFactoryV2ToV3` follows the shape `migrateFactoryV1ToV2` already establishes
(`internal/homestate/factory.go:211`): the `ALTER TABLE`s inside one transaction, then the
`meta.schema_version` update. No new migration machinery.

The defaults are what make **every pre-existing row** legible rather than broken: `lead_pid = 0` is
the sentinel that routes a legacy row to the REQ-006 peer fallback. Migration and prevention are
therefore the same mechanism rather than two, which is the property `spec.md` §C claims for this
design.

### G.2 `retired` is a status, not a deletion

Retirement is `UPDATE runs SET status='retired', updated_at=?` plus an appended `run.retired`
event (REQ-010). The row survives.

Deleting would make the retirement itself unobservable: the `events` table is the only record of
what the factory did, and an operator debugging a vanished run would have a gap where the
explanation should be. It would also make AC-002 unwritable — the criterion asserts the row *still
exists* with the new status.

`retired` is the third value the `status` column has ever held, and it is chosen so the existing
reader needs no change: `ResolveActiveRun`'s query is `WHERE status='active'` (M-03), so a retired
run leaves the active set by no longer matching, not by any change to the predicate.

---

## §H Control flow at resolution time

`ResolveActiveRun` (`internal/factorymsg/store.go:241`) keeps its shape. One step is inserted, and
only on the branch that is already failing:

```
explicit != ""  ──> validate the named run, return           (unchanged)

runs := SELECT run_id FROM runs WHERE status='active'
switch len(runs):
  0    -> NO_ACTIVE_FACTORY                                  (unchanged)
  1    -> runs[0]                                            (unchanged)
  many -> ReconcileActiveRuns(fallback)                      (NEW)
          re-query
          switch len(runs): 0 -> NO_ACTIVE_FACTORY
                            1 -> runs[0]
                            many -> AMBIGUOUS_FACTORY + per-run classifications
```

Three properties a reviewer should check for directly:

1. **The happy path is untouched.** Reconciliation runs only where resolution was going to fail
   anyway, so a single-active-run join pays nothing — no probe, no broker open.
2. **Fail-closed is preserved, and cannot be weakened by this shape.** Reconciliation only ever
   *removes* provably-dead owners from the active set; it never selects among survivors. Ambiguity
   that survives it is real ambiguity (REQ-014). "Pick the newest" was the tempting fix and is
   explicitly forbidden — it makes the symptom disappear while joining a worker to an arbitrary run.
3. **The sentinel substrings survive.** `AMBIGUOUS_FACTORY` and `NO_ACTIVE_FACTORY` remain literal
   substrings of the error text; existing tests match on them. The new classification detail is
   appended, so the operator reads *why each survivor survived* instead of an opaque refusal —
   which is the difference between a dead end and a next step.

---

## §I Failure exits, in one place

| Point | Condition | Exit |
|---|---|---|
| `RecordRun` | stamp write fails | whole transaction rolls back; no row, no partial stamp |
| restamp seam | run row missing / update fails | launcher's existing error path; the pre-restamp row is launcher-stamped and correctly reaped |
| non-replace door | no session identity obtainable | **refuse the launch**, leave no launcher-stamped run (REQ-002d) |
| `classify` | complete identity, but probe error or unreadable fingerprint | `indeterminate` — never retired |
| `classify` | neither column nor peer yields a complete identity | `dead` if every REQ-006b premise holds; otherwise `indeterminate` — never retired |
| `classify` | boot proof: boot time unknown, broker file present or unstat-able, a timestamp at or after boot, or one that does not parse | `indeterminate` — never retired |
| `classify` | boot proof: reading the run's recorded timestamps fails | the error is returned; nothing is retired |
| `ReconcileActiveRuns` | fallback lookup fails for one run for a reason other than the broker file not existing | that run has no peer identity and reaches the boot proof, whose premise 2 then declines (the broker file exists or could not be checked), so it is `indeterminate`; reconciliation continues over the rest |
| `ResolveActiveRun` | ≥2 survive reconciliation | `AMBIGUOUS_FACTORY` + classifications |
| `ResolveActiveRun` | 0 active | `NO_ACTIVE_FACTORY` |
| `moai factory --retire` | target classifies `live` **or** `indeterminate` **or** anything not `dead` | non-zero exit, run stays `active`, classification named as the reason |

Every row whose condition is an *absence of knowledge* exits toward `live`. That is the single
disposition rule behind the whole table, and it is what makes "never retires a live run" a
structural property rather than a claim maintained by vigilance. The one `dead` exit for an
identity-less run is not an absence of knowledge: it requires every REQ-006b premise to be
positively established, and any premise that cannot be established falls back to `indeterminate`
under the same rule.

---

## §J Design invariants — what a reviewer can check without running anything

1. **One identity per run, and both sources name the same process.** The `runs` stamp and the
   `role='lead'` peer are the session process on every door (§B, §C.3). A design where they can
   diverge is the D1/D11 defect. → `acceptance.md` AC-016.
2. **Retirement is gated on a positive `dead`, on every path, expressed once.** No reject-list, no
   per-path copy. → AC-017, and the mutant probe that separates it from the enumeration.
3. **`homestate` never imports `factorymsg`.** The fallback crosses the boundary as a function
   value (§F). → a compile failure is the only way to violate this.
4. **`ResolveActiveRun` never selects among survivors.** Reconciliation reduces the set; it does
   not choose from it. → AC-009.
5. **The restamp seam carries no build tag.** → AC-016 is runnable on darwin; a tagged seam makes
   it unrunnable on the only pre-merge platform.
6. **The row is never deleted.** → AC-002.

---

## §K Cross-references

- `spec.md` §A.1 — the five call sites and three launch shapes this design branches on
- `spec.md` §C / §C.1 / §C.2 — the mechanism choice and the two sub-decisions taken as input here
- `plan.md` §F — the milestone order these seams are built in
- `acceptance.md` §C.2 — which criterion binds each invariant in §J
- `research.md` — the measurement ledger (`M-nn`) every claim above cites

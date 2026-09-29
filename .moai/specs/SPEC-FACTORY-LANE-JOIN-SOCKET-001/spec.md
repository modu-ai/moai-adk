---
id: SPEC-FACTORY-LANE-JOIN-SOCKET-001
title: "Factory lane join tolerates run-record absence via verified leader discovery; -l/--lead target flag"
version: "0.1.0"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.3.0 target"
module: "internal/cli, internal/homestate, internal/kanban"
lifecycle: spec-anchored
tags: "factory, lane-join, leader-discovery, liveness, run-resume, card-t1330"
tier: L
card: t1330
related_specs: [SPEC-FACTORY-MIXED-HOOK-001, SPEC-FACTORY-RUN-RETIRE-001, SPEC-FACTORY-BOOTSTRAP-001, SPEC-FACTORY-WORKER-NAMING-001, SPEC-CODEX-FACTORY-RETIRE-001]
---

# SPEC-FACTORY-LANE-JOIN-SOCKET-001 — Factory lane join tolerates run-record absence

## §A Background and Motivation

A factory lane join trusts the factory run record and nothing else. The lane branch's first step
is `enterSelectedFactoryRun(root, entry.FactoryRun, /*requireActive=*/true)`
(`internal/cli/glm.go:256`, `internal/cli/cc.go:207`, and the codex lane twin
`internal/cli/codex_factory.go:96`), which resolves through `factorymsg.ResolveActiveRun` — a
`SELECT run_id FROM runs WHERE status='active'` (`internal/factorymsg/store.go:272-325`). Zero
survivors refuse with `NO_ACTIVE_FACTORY`. No socket, pid, or liveness probe exists anywhere in
the join path.

The defect is measured, not inferred. On 2026-09-28 the run `tlwgk9` was auto-retired by the
retirement machinery (`run.retired {"classification":"dead","basis":"stamp"}`, twice — the
deliberate, no-undo mechanism of SPEC-FACTORY-RUN-RETIRE-001). A lane join was then refused with
`NO_ACTIVE_FACTORY` even though a live lead session existed on the machine. The current run
`tm3yoq` has a live leader (pid verified holding `/tmp/cc-socks/65728.sock` via `lsof`), and
joining works — the contrast that defines the defect. Re-measured in this tree at `68e37864a`
during plan phase: a fresh fixture project (zero active runs) refuses `moai cc -f lane` with
`No_active_factory.` exit 1 — and the refusal is unconditional, because the join path consults no
liveness surface at all (the discovery-absence grep reads 0 hits; positive control
`enterSelectedFactoryRun` reads 6; `acceptance.md` §C carries the four-element RED cells).

Two readings of "run-record absence" shape this SPEC:

1. **Record retired, leader alive.** A run's record was retired (correctly, about the identity it
   probed) while a live lead session still serves the factory. The operator's only remedies today
   are relaunching the lead (losing the session) or hand-editing the database.
2. **Record absent entirely.** A live lead session exists whose run was never recorded (the
   historical "recording only in the kanban store left every GLM-led run unjoinable" failure
   class, `glm.go:240-242`).

A second reading is fixed here because the card text says "connect to the live lead session
socket": the join does NOT dial the Claude-Code-runtime socket. Lanes and leads already exchange
everything through the established substrate — the run-scoped broker
(`factorymsg.Open(root, runID)`) plus name-addressed cross-session messaging (the substrate
SPEC-FACTORY-BOOTSTRAP-001 named). What "connect" operationally means is: verify the leader is
live, restore the run identity both sides share, and let the lane's SessionStart hook bind its
peer into that run's broker. Whether `/tmp/cc-socks/<pid>.sock` sockets speak a dialable protocol
is an undetermined runtime question this SPEC deliberately does not depend on: the socket
directory serves as a **candidate enumeration source only**, never as a liveness proof and never
as a transport.

The repair is a **parallel, verified leader-discovery path layered at the single join point** —
not a widening of `ResolveActiveRun`. SPEC-FACTORY-RUN-RETIRE-001 REQ-014 forbids weakening the
fail-closed resolver and its rejected alternative (d) explicitly forbids picking among active
runs; SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-001 fixes the `NO_ACTIVE_FACTORY` / `AMBIGUOUS_FACTORY`
refusal contract. This SPEC amends neither: the resolver is untouched, and discovery runs only on
the refusal path, must clear the same retire-grade proof standard the retirement machinery
requires, and re-enters the join through the same gate. A run restored by discovery is
indistinguishable from a run the lead recorded itself — including to `ReconcileActiveRuns`, which
stays the tool that reaps the record if the discovery verdict was wrong.

### A.1 Scope

Card t1330 owns the **join path only**: `moai glm -f lane` and its `moai cc` twin (and the codex
lane twin that shares the join point). Auto-dispatch defaulting is card t1332's scope and is
excluded. Lane slot numbering semantics are preserved unchanged.

### A.2 Verified shape of the existing mechanism (measured in this tree)

- `internal/cli/factory.go:223-273` — `parseLauncherEntry`; `-f lane` desugars at parse time to
  `--name lane-<next>` via `kanban.NextFactoryLaneNumber` (`bootstrap.go:375` — dead claims
  pruned, highest live + 1), **before** any run gate. Auto-assignment already tolerates
  run-record absence; only the gate blocks the join.
- `internal/cli/factory.go:275-302` — `enterSelectedFactoryRun`; on success sets
  `EnvMoaiKanbanID`, the run identity the lane claim (`factory.go:492`,
  `ClaimFactoryLane`) stamps and the hook-side bind chain requires.
- `internal/hook/factory_messages.go:54-128` — the SessionStart hook requires `EnvMoaiKanbanID`
  non-empty (:55-58), re-validates the run (`ValidateActiveRun`, :88), opens the run-scoped
  broker (:91), and binds the peer generation-1 (:96-127). Run-record absence therefore breaks
  the post-join bind chain too — a join that produced no run identity would register no peer and
  receive no dispatches.
- `internal/kanban/role.go:41,46` — the canonical leader noun is `leader`
  (`kanban.LeaderLabel()`); `lead` is the refused legacy spelling
  (`IsLegacyLeaderSpelling`; refused at entry by `refuseLegacyEntryNames`,
  `factory.go:185-198`, REQ-RNC-004/-007). The live leader argv is `--name leader`; the live
  leader env carries `MOAI_KANBAN_LEAD_NAME=leader`.
- `internal/config/envkeys.go:189-195,218-225,255-273` — `EnvMoaiKanbanID`,
  `EnvMoaiKanbanLeadAddr`, `EnvMoaiKanbanLeadName`; the flag's effect reaches the child/hook via
  env by design (`factory.go:23-27`; precedent `exportLeaderSessionName` →
  `MOAI_KANBAN_LEAD_NAME`).
- `internal/cli/factory.go:323` + `internal/homestate/runtime.go:38,43` — `recordFactoryRunStart`
  stamps the **calling** process's identity (`ON CONFLICT(run_id) DO UPDATE SET status='active'`,
  `run.started` event). A resume path that reused it from the lane launcher would stamp the
  **lane** as the run's owner — the lane's exit would then retire a live lead's run. The resume
  writer must stamp the verified leader identity instead (REQ-004).
- Discovery surface reality (live, read-only): `/tmp/cc-socks/` holds ~97 pid-named sockets, many
  stale (dead pids); no name→socket registry exists. Socket-file existence is not liveness. The
  join-relevant live leader evidence a probe can read: pid alive + process-start fingerprint
  (`homestate.ProbeProcessIdentity` / `ClassifyOwnerWith`), `--name <leader label>` in argv, the
  candidate's own `MOAI_KANBAN_ID` env (its run identity), and the canonical project cwd.

## §B Requirements (GEARS)

### Discovery path

- **REQ-001** (Event-driven): **When** a factory lane join's run resolution fails with
  `NO_ACTIVE_FACTORY` and the operator supplied no explicit run id, the launcher shall run leader
  discovery before refusing, and shall refuse with `NO_ACTIVE_FACTORY` only after discovery has
  verified no live leader for this project.
- **REQ-002** (Ubiquitous): Discovery shall accept a candidate leader only on the same proof
  standard the retirement machinery requires — a live classification from the PID **together
  with** the process-start fingerprint predicate. Socket-file existence, a broker peer row, or
  any other record alone shall never constitute liveness.
- **REQ-003** (Ubiquitous): Discovery shall accept a candidate only when all three hold: it is
  live (REQ-002); it is verifiably a leader session of **this** canonical project (the leader
  label targeted per REQ-008, plus project membership evidence — canonical cwd or the
  candidate's own run-id env); and the candidate's run identity (its own `MOAI_KANBAN_ID`) can
  be read from the candidate's process. A candidate whose run identity cannot be read shall be
  declined — the run identity is never guessed, defaulted, or minted fresh, because the run id
  is the address of the broker the lead already speaks on.
- **REQ-004** (Event-driven): **When** discovery verifies exactly one live leader, the launcher
  shall restore that leader's run to a joinable state through a dedicated resume writer that:
  creates the run row when absent and reactivates it when retired; stamps the row's owner
  identity with the **verified leader's** process identity (never the joining lane's); and
  appends an auditable resume event recording which verification produced the restore. The join
  then proceeds through the existing run gate.
- **REQ-005** (Event-driven): **When** discovery verifies more than one live leader, the launcher
  shall fail closed naming each verified candidate with its run id, and shall not select among
  them — the same prohibition REQ-014 applies to picking among active runs applies to picking
  among live leaders, and no run state shall be mutated on that refusal.
- **REQ-006** (Unwanted): The discovery path shall not weaken run-resolution fail-closed
  semantics: `ResolveActiveRun` shall remain unchanged; zero active runs with zero verified live
  leaders remains `NO_ACTIVE_FACTORY`; two or more active runs after reconciliation remains
  `AMBIGUOUS_FACTORY`; and an explicit `--factory-run <id>` naming a non-active run shall fail
  `NO_ACTIVE_FACTORY` without discovery (an explicit selection is a decision, not an absence).
- **REQ-007** (Unwanted): The launcher shall not join a lane to a run whose leader cannot be
  verified live at join time. A discovered run whose leader probe does not yield a live
  classification shall not be resumed — the inverse defect (joining against a dead lead, which
  the stale broker rows of retired runs demonstrate) is the same fault as the one this SPEC
  repairs, in the other direction.

### Operator surface

- **REQ-008** (Capability gate): **Where** a lane join is performed, the launcher shall accept
  `-l, --lead <name>` naming which leader session discovery targets. The default target shall be
  the canonical leader label (`kanban.LeaderLabel()`). The value `lead` and any `lead-<suffix>`
  form shall be refused with the canonical-form error, the same refusal `--name` applies
  (REQ-RNC-004/-007). The flag shall be accepted only on lane joins; a leader entry carrying it
  is an error, and carrying it together with `--factory-run` is an error (the two name
  different selectors).
- **REQ-009** (Ubiquitous): A lane joined through discovery shall receive the run identity via
  `MOAI_KANBAN_ID` and the target leader's name via `MOAI_KANBAN_LEAD_NAME`, so the child
  session's hook binds its peer into the resumed run's broker and addresses the lead by name —
  the same channels an ordinary join uses.

### Landing constraints

- **REQ-010** (Ubiquitous): The discovery and resume path shall live in the single shared join
  point all lane joins pass, so `moai cc`, `moai glm`, and the codex lane twin exhibit identical
  behavior; no launcher-private copy shall exist.
- **REQ-011** (Ubiquitous): Existing join-path invariants shall be preserved unchanged: lane
  auto-assignment semantics (parse-time project-scoped pick, dead-claim pruning, atomic claim,
  bump rule — numbering never resets per run); the one-entry-token rule; the `-f lane` +
  operator `--name` conflict; `--name lane-<n>` alone does not join a run; and the codex-led
  leader-adoption refusal.
- **REQ-012** (Ubiquitous): Every test or reproduction exercising discovery or resume shall
  isolate `HOME`, `MOAI_HOME`, and the launched-binary path, and shall use a project directory
  outside this repository's worktree set — `homestate.CanonicalProjectRoot` converges a linked
  worktree onto the primary checkout, so an omitted isolation mutates live factory state.

## §C Design Decision — parallel verified discovery at the single join point

**Chosen: a parallel discovery-and-resume path inside the shared join gate** (`REQ-001`,
`REQ-010`). On the `NO_ACTIVE_FACTORY` failure branch, the launcher probes for a live leader;
one verified leader → resume the run (`REQ-004`) → re-enter the same gate; zero → the refusal
stands; many → fail closed (`REQ-005`). The resolver, the retirement machinery, and both refusal
contracts are untouched.

Why this framing and not an amendment of SPEC-FACTORY-MIXED-HOOK-001 / REQ-014: the refusal
contracts remain true — they describe what the **resolver** does, and the resolver is unchanged.
Discovery is a different mechanism (process-liveness discovery, not record selection), which is
exactly the shape RUN-RETIRE's rejected-alternative analysis leaves open: the forbidden act is
picking among *active records*; discovering a *live process* and restoring its record is the
repair for the case the record set is wrong. The resume is self-healing: if the discovery verdict
was wrong, `ReconcileActiveRuns` re-retires the row by the same proof standard — no new
enforcement surface is created.

Why discovery does not dial sockets: the transport question (whether an external Go process can
join `/tmp/cc-socks` protocol) is undetermined and unnecessary. Candidate enumeration from the
socket directory + pid/fingerprint/env probing reuses primitives that already exist
(`homestate.ProbeProcessIdentity`, the t1038 probe chain "socket reachable → pid alive → name →
cwd → tree"), and the actual lane↔lead exchange rides the run-scoped broker as it does today.

Alternatives rejected:

| Alternative | Why rejected |
|---|---|
| Widen `ResolveActiveRun` to select among / resurrect rows | The explicitly rejected alternative (d) of SPEC-FACTORY-RUN-RETIRE-001; REQ-014 forbids it. Discovery is a parallel path, not a resolver change. |
| Amend REQ-FMH-001 / REQ-014 to permit record-optional joins | Unnecessary: the parallel path satisfies the operator requirement without touching a settled fail-closed contract. Amending a completed SPEC's refusal law to admit a weaker gate is the exact weakening REQ-014 names. |
| Socket-connect transport (dial the leader socket) | Depends on two undetermined runtime questions (name→socket registry, dialable protocol). The broker + name-addressed messaging already carries the lane↔lead exchange; the socket adds nothing the join needs. |
| Reuse `recordFactoryRunStart` for the resume | It stamps the **calling** process — from the lane launcher that is the lane, so the lane's exit would retire a live lead's run (REQ-004's reason for a dedicated writer). |
| Resume by minting a new run id | The run id is the broker address; the lead's messages flow under the id its env carries. A new id would join the lane to an empty broker. |
| Auto-pick when multiple live leaders are found | Same reasoning as REQ-014's "newest/first is prohibited": two live leaders is genuine ambiguity, and joining either is arbitrary. |
| Per-run lane-number reset on join | A behavior change to project-scoped, run-agnostic numbering (measured: `NextFactoryLaneNumber` consults no run state); card scope is the gate, not the numbering. |

## §D Acceptance criteria

The authoritative AC set lives in `acceptance.md` (16 criteria, AC-001..AC-016, Given-When-Then
with RED-now / green-path cells). Summary: the defect contrast (AC-001), fail-closed
preservation (AC-002/003/004), liveness discipline in both directions (AC-005/007), resume
correctness and owner stamping (AC-006/007/008), operator flag surface (AC-009/010/011), env and
bind-chain outcome (AC-012), mirror parity (AC-013), preserved invariants (AC-014/015), and
documentation (AC-016).

## §E Exclusions

### Out of Scope — card t1332 territory

- Auto-dispatch defaulting (any behavior where a bare leader entry or a bare `-f` auto-routes or
  auto-dispatches work). This SPEC relaxes only the lane **join gate**.

### Out of Scope — lane numbering semantics

- Making lane numbering run-scoped, resetting numbering per run, or changing the bump rule.
  Numbering stays project-scoped and run-agnostic (`NextFactoryLaneNumber` consults no run
  state today; that is preserved).

### Out of Scope — the run resolver and retirement machinery

- Any change to `ResolveActiveRun`, `ReconcileActiveRuns`, `RetireRunIfDead`, the retirement
  bases, or the `AMBIGUOUS_FACTORY` contract. Discovery reads these surfaces; it modifies none.

### Out of Scope — the Claude Code runtime socket substrate

- Implementing or assuming a client for the `/tmp/cc-socks` transport protocol, building a
  name→socket registry, or changing `FactoryLeaderSocketPath` / `MOAI_KANBAN_LEAD_ADDR`
  (measured display-only convention). The socket directory is a candidate source only.

### Out of Scope — release activities

- No release cut, push, or PR creation. Integration follows the git-flow lane protocol.

## §F Residual risk

- **Leader env readability is platform- and privilege-dependent.** Reading a candidate's
  `MOAI_KANBAN_ID` from its process environment works for same-user processes on darwin
  (`ps eww`) and linux (`/proc/<pid>/environ`); where it does not work, REQ-003 declines the
  candidate and the behavior degrades to today's refusal — fail-closed, never a wrong join. The
  run phase measures this per platform.
- **Time-of-check-to-time-of-use.** The leader may die between the probe and the resume. The
  record then describes a dead owner and `ReconcileActiveRuns` re-retires it on the next
  resolution; the lane's hook bind degrades to the existing "factory messaging degraded" notice.
  Self-healing by construction; named, not prevented.
- **argv/env scraping fragility.** The classifier reads process argv and env, which launcher
  evolution could reshape. The regression pair pins the classifier's contract; a launcher-argv
  change that breaks discovery fails AC-001 loudly (the join refuses), not silently.
- **Stale-socket volume.** ~97 candidate sockets observed live; candidate enumeration is
  bounded per-candidate by the probe deadline. A pathological candidate population could add
  join latency; bounded and measured in the run phase.
- **Codex lane twin inherits the behavior via the shared join point** (REQ-010). The card
  verifies cc/glm; the codex twin's inheritance is asserted at source level (AC-013) rather than
  exercised end-to-end, consistent with how prior SPECs covered shape-matched doors.

## §G History

- 2026-09-29 — v0.1.0 — manager-spec — initial plan-phase draft (card t1330, Tier L),
  authored from the four-lens research fan-out (workflow run wf_7795d774-54e,
  `.moai/reports/t1330/research-draft.md`, relocated to `research.md`) plus plan-phase
  re-verification in this tree at `68e37864a`.

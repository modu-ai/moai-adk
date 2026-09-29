---
id: SPEC-FACTORY-LANE-JOIN-SOCKET-001
title: "Research — factory lane join defect (four-lens fan-out + plan-phase re-verification)"
version: "0.1.0"
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
---

# research.md — SPEC-FACTORY-LANE-JOIN-SOCKET-001

> Provenance: synthesized from a four-lens read-only research fan-out (lenses:
> codebase-precedent, constraints-risks, prior-SPEC-memory, live-state-evidence) for card
> t1330, workflow run **wf_7795d774-54e**. Raw per-lens reports:
> `.moai/reports/t1330/research-lenses.json`; the original synthesis:
> `.moai/reports/t1330/research-draft.md` (retained as card evidence). Lens-output integrity
> note preserved from the synthesis: the live-state lens's raw output contained
> instruction-shaped text (a "bypass-permissions" pattern) that the harness neutralized before
> synthesis; the only remaining directive-shaped token is the observed live leader process argv
> (`claude --permission-mode bypassPermissions --name leader ...`) — a factual observation about
> a running process, relayed as evidence, not followed as an instruction.
>
> Plan-phase re-verification (this tree, `68e37864a`): every load-bearing anchor cited in
> spec.md §A.2 was re-read directly (glm.go lane branch, factory.go parse/gate/claim,
> store.go resolver, factory_messages.go hook chain, role.go labels, envkeys constants,
> regression tests) — the anchors in this document that survive into spec.md are first-hand,
> not carried. Deltas beyond the lens reports are in §J.

## §A Findings (10)

1. **Defect site** — the lane join is record-trust-only, gated first:
   `enterSelectedFactoryRun(..., requireActive=true)` at `glm.go:256` (cc twin `cc.go:207`,
   codex twin `codex_factory.go:96` — the third call site, plan-phase addition) →
   `factorymsg.ResolveActiveRun` (`store.go:272-325`, `NO_ACTIVE_FACTORY` at 285/304/319,
   `AMBIGUOUS_FACTORY` at 323, deliberately fail-closed: reconciliation "only ever REMOVES
   provably-dead owners … never selects among survivors"). The run id is the join credential:
   `enterSelectedFactoryRun` (`factory.go:275-302`) sets `EnvMoaiKanbanID`; the broker is
   run-scoped (`store.go:131-140`).
2. **The fail-closed gate is normative SPEC law** — SPEC-FACTORY-MIXED-HOOK-001 REQ-FMH-001
   (verbatim SHALL, re-read from the SPEC at spec.md:58); SPEC-FACTORY-RUN-RETIRE-001 REQ-014
   (no weakening, no newest/first pick); rejected alternative (d) forbids widening
   `ResolveActiveRun`. `enterSelectedFactoryRun` is the single join point (pinned by
   SPEC-CODEX-FACTORY-RETIRE-001 design.md:53). Regression pins:
   `TestGLM_FactoryLeadRunIsJoinableByLane` (`internal/cli/factory_test.go:1103`; the
   recording-only-leaves-`NO_ACTIVE_FACTORY` guard is its `:1099-1102` comment block) and
   `TestFactoryRunSelectionAtomicSlotsAndArgv` (`internal/cli/factory_mixed_test.go:87`,
   asserts the string).
3. **The defect state was produced deliberately** — auto-retire has no operator undo:
   `ReconcileActiveRuns` (`factory_run_retire.go:264-336`), `retirable()` positive gate with
   `@MX:WARN` "a wrong dead verdict retires a live lead's run and has no operator undo
   (REQ-005)". Live: two `tlwgk9` `run.retired {"classification":"dead","basis":"stamp"}` events
   2026-09-28. Discovery must be held to retire-grade proof (the inverse defect is joining dead
   peers); `ReconcileActiveRuns`/`RetireRunIfDead` remain the re-sync tools after a resume.
4. **Auto lane-slot assignment already exists and is mis-framed by the card text** — `-f lane`
   desugars at PARSE time (`factory.go:258` → `NextFactoryLaneNumber`, `bootstrap.go:375`:
   prune dead claims, highest live + 1) BEFORE any run gate; project-scoped, run-agnostic,
   never resets per run; atomic claim stamps pid AND run id (`factory_slots.go`); bump rule
   with operator notice (`factory.go:492-511`). Only the gate blocks the join.
5. **Run-record absence breaks the post-join bind chain too** — the SessionStart hook requires
   `EnvMoaiKanbanID` non-empty (`factory_messages.go:55-58`), re-validates the run (:88), opens
   the run-scoped broker (:91), binds generation-1 (:96-127). A run-record-optional join must
   still produce a run id — pure socket-only joining breaks peer bind and dispatch.
6. **Socket-first leader discovery has no production precedent; three feasibility unknowns** —
   zero `cc-socks` references in `internal/`; the only precedent is card t1038's read-only shell
   probe (`socket reachable → pid alive → name → cwd → tree`). `/tmp/cc-socks` holds ~97
   pid-named sockets, many stale — existence is NOT liveness. The advertised
   `FactoryLeaderSocketPath` (`bootstrap.go:426`) is a display-only convention ("not a
   filesystem contract"); `/tmp/moai-socket-factory/` does not exist even on the live leader's
   machine (SPEC-CHAIN-CORE-001 measured the artifact a DEAD STUB). Available discovery
   surfaces: the frozen-schema session registry (empirically stale, no name field), broker
   `peers` (stale-prone), sessionmsg AgentRecords (name-bearing), and ps argv/env scraping —
   the only demonstrated name→pid→identity bridge. The pivotal unknown (name→socket registry
   queryable from Go; dialable socket protocol) remains UNDETERMINED and this SPEC does not
   depend on it.
7. **`-l` is free; the proposed default `lead` is the refused legacy spelling** — no `-l` flag
   exists (`todo_triage.go:396` is a grep argument); `lead` collides with
   `legacyLeaderSpelling` (`role.go:46`), refused at entry (REQ-RNC-004/-007), carried by
   doctor's "legacy run: relaunch required" (REQ-RNC-009/-024), and by the retired tlwgk9's
   broker rows. Live canonical default: `kanban.LeaderLabel()` = `"leader"` (`role.go:41`),
   live leader argv `--name leader`, env `MOAI_KANBAN_LEAD_NAME=leader`. Novel semantic: the
   flag names a TARGET session, not the joining session's own name — no in-repo precedent
   distinguishes the two uses.
8. **Stale-evidence hazards run both directions; canonicalization constrains new state** — the
   defect direction (record-trust refuses a live join) and the inverse (record-trust joins dead
   peers — tlwgk9's live-looking broker rows, dead-pid worker rows). Any new state must live
   under the canonical project key (`paths.go` collapses worktrees onto one key) — design.md's
   discovery writes no new state at all (read-only probing + one resume row/event), which
   sidesteps the constraint.
9. **Landing constraints** — enforced mirrors (drift precedent `glm.go:240-243`); env-as-signal
   (`factory.go:23-27`, precedent `exportLeaderSessionName` → `MOAI_KANBAN_LEAD_NAME`,
   `envkeys.go:255-273`); parse invariants (`factory.go:238-241` one-entry-token, `:251-253`
   conflict, `:361-370` truth table, `:306-321` codex-leader refusal); docs surface (4-locale
   factory-mode/kanban-mode/launchers pages, 4 README occurrences, help `glm.go:40-127`).
10. **Novelty verdict: no prior SPEC proposes the approach** — negative searches 0 hits for
    socket-first/lead-discovery/`--lead`/record-optional joining. SPEC-FACTORY-BOOTSTRAP-001
    left the leader-socket-path producer unresolved (this card closes the gap by a probe, not a
    socket); its `crossSessionInbound` gate is unchanged inbound-path behavior a joined lane
    already hits. RUN-RETIRE §E deliberately excluded worker-roster/peer lifecycle. Prior art:
    WORKER-NAMING (next-free slot + pinned number), WORKER-FANOUT (`lead-<run-id>` name family),
    MIXED-HOOK (atomic slot claims).

## §B Contradictions (7) — resolutions adopted

1. **`--lead` default `lead` vs refused legacy vocabulary** → resolved for `leader`
   (REQ-008; plan.md clarification 1, operator confirms at gate).
2. **Card text vs actual lane-slot mechanism** (all four lenses vs the task framing) → resolved
   for the measured mechanism: parse-time project-scoped pick; numbering untouched (REQ-011).
   The card's "generation 2" attribute also corrected: `generation` is a broker `peers` fencing
   column, not a `runs` field.
3. **`active-sessions.json` available vs empirically stale** → resolved by not building on it:
   the design's classifier reads the candidate PROCESS (pid/fingerprint/env/cwd), not any
   registry.
4. **Which SPEC owns the `-f lane` spelling as first-class** (t1085 renamed notation to
   `worker-N` with `lane-N` aliases; current tree has `-f lane` canonical, `RoleLane = "lane"`,
   help text canonical) → resolved by current-tree reality: `-f lane` is first-class in this
   tree; no rename is touched.
5. **LEAD_ADDR constant/path spelling drift across SPEC citations** → resolved by citing only
   current-tree constants (`MOAI_KANBAN_LEAD_ADDR`, `envkeys.go:218-225`); the artifact remains
   display-only and out of scope.
6. **Declared messaging substrate vs external-process feasibility** → resolved by not depending
   on the runtime socket protocol at all: candidate enumeration only; transport stays the
   run-scoped broker + name-addressed messaging.
7. **Line-citation drift across lenses** → resolved by first-hand re-reads; spec.md §A.2 anchors
   are plan-phase-verified against `68e37864a`.

## §C Open questions (10) — dispositions

| # | Question | Disposition |
|---|---|---|
| 1 | Did a live lead socket exist at the moment of the rejected join? | Past event, not reproducible; immaterial to the requirement (the refusal is unconditional — R-1/R-3 measured). |
| 2 | Runtime name→socket registry queryable from Go? | Pivotal unknown, deliberately not depended on (spec.md §A). |
| 3 | Are stale sockets connectable when the pid is dead? | Unknown; irrelevant — sockets are candidacy only, never liveness (REQ-002). |
| 4 | Does cc-socks protocol support an external join? | Unknown; irrelevant — no dialing (spec.md §A). |
| 5 | Are REQ-FMH-001/REQ-014 blocking or amendable? | [NEEDS CLARIFICATION: amendment-vs-parallel-path] — recommendation: parallel path, no amendment (plan.md clarification 2). |
| 6 | Should lane numbering become run-scoped? | No — preserved project-scoped (REQ-011, Out of Scope). |
| 7 | `peers.generation` allocation rule | Untraced; not needed by this design (bind stays generation-1 via the existing hook chain). |
| 8 | Hook-side vs launcher-side discovery ownership | Launcher-side (design.md §A — the hook consumes the produced run id; splitting ownership would need the hook to mutate run state, which REQ-004's transactional write makes unnecessary). |
| 9 | tlwgk9 broker dir mtime anomaly | Unexplained, not load-bearing. |
| 10 | Can `lead` as a flag default coexist with REQ-RNC-004? | Moot under the recommended `leader` default; the VALUE `lead` is refused outright (REQ-008). |

## §D Design implications → where they landed

- Parallel socket-verified path at the single join point, no resolver widening → spec.md §C,
  REQ-001/006/010.
- Run-record-optional join must still produce run identity; socket-liveness first, then
  attach/resurrect → REQ-003/004, design.md §C.
- Retire-grade liveness proof; avoid joining dead peers → REQ-002/007, AC-005/007.
- `-l` free; default collision; env plumbing; mirror parity → REQ-008/009/010.
- Lane auto-assignment needs no new mechanism; do not perturb → REQ-011, AC-014/015.
- Regression pair + 4-locale docs + README + help → AC-013/014/015/016, plan.md M4/M5.

## §J Plan-phase re-verification deltas (this tree, `68e37864a`)

Measured first-hand during SPEC authoring, beyond the lens reports:

- **Third join call site**: `internal/cli/codex_factory.go:96` also passes
  `enterSelectedFactoryRun(root, entry.RunID, true)` — the "single join point" carries THREE
  lane-join callers, so REQ-010's parity claim covers the codex twin by construction.
- **Resume-writer stamp hazard**: `recordFactoryRunStart` (`factory.go:323`) +
  `RecordRun` (`runtime.go:38,43`) stamp the CALLING process (`ON CONFLICT(run_id) DO UPDATE`,
  `run.started` event) — reusing either from the lane launcher would stamp the lane as owner.
  This produced REQ-004's dedicated-writer requirement and AC-007's mutant.
- **`run.started` precedent**: the events table pattern (`runtime.go:43`) supports a
  distinguishable `run.resumed` kind with payload (AC-008) without schema change.
- **Board-role declaration carrier examined and rejected**: `internal/kanban/role.go`
  `DeclareRole`/`ResolveDeclaredRole` (a session-id→role/label declaration artifact) has ZERO
  non-test callers in this tree — no live data exists to discover from; it is not a candidate
  source. Recorded so a later reader does not re-derive it.
- **RED-now measurements** (acceptance.md §C): E2E refusal (`No_active_factory.`, exit 1),
  regression pair green (`ok ... 3.531s`, anchored `-run` pattern), absence grep 0 / positive
  control 6.
- **Build baseline**: `go build -o /tmp/t1330-moai ./cmd/moai` exit 0 at the pinned tree.

## §K Remaining unknowns carried into run (named, not hidden)

- darwin `ps eww` env-read reliability under the probe deadline (M2 measures; degradation =
  refusal, REQ-003).
- Candidate-population latency against the live ~97-socket directory (M2 measures, bounded).
- Codex twin E2E leg (asserted at source level per spec.md §F, not exercised end-to-end).
- [NEEDS CLARIFICATION: --lead default value] and
  [NEEDS CLARIFICATION: run-id provenance via leader env] — plan.md §F clarifications 1 and 3;
  both degrade to an honest refusal if answered otherwise and the answer changes M2/M3 shape,
  not M1.

# research.md — SPEC-FACTORY-LANE-AUTONOMY-001 (card t1338)

> Source: read-only reconnaissance carried by the dispatching orchestrator (single-Explorer
> selection; see progress.md §E.1 for the Phase 2/6 fan-out skip rationale). Every claim below is
> attributed to the explorer's measured observation; items the author could not re-observe in this
> tree are marked as branch-resident. Quantitative claims cite their source.

## §R1. Boundary card states (measured)

| Card | SPEC | In develop? | State | Evidence |
|------|------|-------------|-------|----------|
| t1240 (F2 self-lease) | SPEC-FACTORY-SELF-DISPATCH-001 | **NO** | Complete on branch `WT-factory-self-dispatch`, tip `d43e50bb3`, absorbed develop `145c3d98c`, merge-ready, pending merge only; frontmatter `status: completed` | Explorer branch read; `factory next [--wait] [--wait-bound]` at `internal/cli/factory_card.go:455-463`, `factory stage`/`complete` at `internal/cli/factory_handoff_recover.go:56` (branch-resident) |
| t1241 (F3 controller) | SPEC-FACTORY-CONTROLLER-001 | **NO** | On branch `WT-factory-controller`: M0 (run gate) + M1a (hand-over edges T29b/T29c) landed-on-branch; M1 record surface through M8 PENDING (`plan.md:85-100`). F3 merge automation (M3: predicates 1-8, write-ahead start event, trial merge) NOT built anywhere | Explorer branch read. Autonomous-merge condition triple quoted in card text: "sync-audit PASS · 충돌 없음 · HEAD^{tree}=HEAD^2^{tree} → local develop merge" |
| t1332 (Jev classification) | (no SPEC yet) | NO | Queued card. Owns classification PRODUCTION. This card owns LANE CONSUMPTION | Card text; dispatch |
| t1330 (join repair) | (no SPEC yet) | NO | Queued card; owns dispatch-join defects | Card text; dispatch |

Verified in THIS tree (develop `145c3d98c`): `SPEC-FACTORY-SELF-DISPATCH-001` and
`SPEC-FACTORY-CONTROLLER-001` are absent from `.moai/specs/` (ls, this run) — the basis for the
`depends_on` decision D8.

## §R2. Landed foundations (in this tree)

- **`todo --auto` foreman (t1306 / SPEC-MANAGER-TODO-001, merge `6af4d4ca3`)**: `--auto` flag
  (`internal/cli/todo.go:289`), `--auto-wait` 30m default, serial foreman (`internal/cli/todo_auto.go`):
  pick one → dispatch directive → judge completion ONLY by reading
  `.moai/reports/<card>/evidence.md` (`autoEvidencePath` :36) → done → /clear guidance → next.
  Liveness `autoLiveness` (:47) via registry+lsof. Invocation is currently the operator's batch
  approval. **NO messaging-availability input anywhere in the path** (grep 0 hits).
- **Factory F1 record layer (t1239 / SPEC-FACTORY-RECORD-001, merge `ed506740b`)**: verbs `factory
  handoff recover-resume / abandon-lane`, `factory runs [--retire]`, `factory assign`, `factory
  status`, `factory decide` (human-only decisions; gates kickoff/push). Lease model in
  `internal/homestate/card_transition.go` (`applyLeaseExpiry` :339, `RenewLease` :368). Integration
  window: manual `moai integration status/acquire/release` (`internal/cli/integration.go:271/325/445`).
- **`worktree done`** (`internal/cli/worktree/done.go`): gates = L1 tier refusal
  (`isL1WorktreePath` :192) + anchored-session guard (`LiveAnchoredSessions` →
  ANCHORED_SESSIONS_PRESENT, `--force` override). **ZERO origin/remote-merge awareness** (grep 0
  hits). Push-then-dispose lives only in operator doctrine (`.claude/rules/local/gitflow-lane-protocol.md` §4/§6/§7).
- **Messaging availability**: no detection API anywhere (`internal/sessionmsg/` is a poll-based
  store only). Doctrine documents QUIET failure (`.claude/rules/moai/workflow/cross-session-messaging.md`
  § Availability constraints: "the failure is quiet — Surface the constraint to the operator").
  Lane-stall observability today: run-owner heartbeat only; t1241 M5 tick / M6 hook pending.
- **Jev classification**: `scripts/jev/` does NOT exist in this tree (dev-only untracked scripts in
  the primary checkout). The 5-token route shape (`LEAD-ANSWER-NOW / LEAD-DECIDE-CAREFULLY /
  MEASURE-FIRST / RETURN-TO-LANE / ASK-OPERATOR`) is documented at
  `.moai/docs/jev-local-operations.md:45`. `internal/kanban/backlog_store.go:83` `BacklogItem` =
  {ID, Text, AddedAt, SpecID, State, Landing, CardUUID, PickedAt, DroppedAt} — **NO priority field,
  NO sequential/parallel axis field**. t1332 owns production; this card owns consumption.

## §R3. Gaps the 4 fragments fill (explorer's conclusion)

1. **Fragment 1**: detection/declaration surface + the switch into the existing `--auto` foreman
   cycle, WITHOUT touching its evidence-read contract.
2. **Fragment 2**: the consumption surface + metadata home for priority/exec-axis. Field schema is
   t1332's production concern — the consumer contract is defined against t1332's future interface.
3. **Fragment 3**: the merge-condition triple applied to the lane-direct path WITHOUT building F3
   M3 machinery; the integration-window serial invariant kept.
4. **Fragment 4**: a machine-checkable "origin push landed" precondition coexisting with the L1
   guard and the anchor guard, plus the auto-disposal act.

## §R4. Doctrine anchors cited

- Queue-as-channel: `.claude/rules/moai/workflow/kanban-dispatch.md` § The delegation channel is
  the queue (messages are nudges, never the delegation).
- Operator promotion monopoly: same file § Entry into the board is an operator act.
- Pre-merge disposal prohibition: `.claude/rules/moai/workflow/kanban-dispatch.md` § Isolation
  ("Dispose of no worktree — L1 or L2 — until the branch is integrated and the remote merge has
  landed") + `CLAUDE.local.md` §4.1 ("워크트리는 원격 머지가 확인되기 전까지 폐기하지 않는다").
- Messaging availability constraints (quiet failure): `cross-session-messaging.md` § Availability
  constraints; § An idle notice is a scheduling hint (an idle notice is NOT completion evidence).
- CI judgment is lead-side: `CLAUDE.local.md` §4.1 ("판정은 CI", lane reads only early signals).

## §R-D1. Discovery note (lead's 부고 — routing suggestion, NOT this SPEC)

`moai todo add` with body text starting `-f lane …` parses the leading `-f` as a flag and FAILS
card registration. This is an argument-parsing defect on the `moai todo add` surface; routing
suggestion: fold into t1330's (join/repair) arg-parsing scope or a dedicated one-line Class A card.
Recorded here so the lead can route it; explicitly Out of Scope for this SPEC (spec.md §F).

## §R5. Version note

`t1241`'s condition triple currently exists only as card text + its SPEC plan (branch-resident).
This SPEC treats it as an INTERFACE and cites it precisely (spec.md §D); the run phase re-reads
t1241's landed SPEC text at M3 entry to pin the normative wording (plan.md M3 pre-flight).

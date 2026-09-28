---
id: SPEC-FACTORY-SELF-DISPATCH-001
title: "Harness-neutral factory F2 — self-dispatching lane"
version: "0.5.1"
status: in-progress
created: 2026-09-27
updated: 2026-09-29
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/hook, internal/homestate, internal/config, internal/kanban"
lifecycle: spec-anchored
tags: "factory, factory-f2, self-dispatch, lane, leader, mcp, session-start, clear-policy, codex-relaunch, card-t1240"
tier: L
card: t1240
depends_on: [SPEC-FACTORY-RECORD-001]
related_specs: [SPEC-FACTORY-RECORD-001, SPEC-AUTONOMY-PRECONDITION-001, SPEC-CODEX-FACTORY-RETIRE-001, SPEC-ROLE-NAMING-CODE-001]
---

# SPEC-FACTORY-SELF-DISPATCH-001 — Self-dispatching lane (Factory F2)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-27 | manager-spec | Initial plan-phase draft for card t1240 on `WT-factory-self-dispatch` at base develop `ed506740b` (includes F1 t1239, t1242, t1245). Written in the vocabulary of SPEC-ROLE-NAMING-CODE-001 (card t1256, still unlanded; read at branch `WT-role-naming-code` `d39a1dc09`). |
| 0.2.0 | 2026-09-27 | manager-spec | Plan-audit iteration 1 (FAIL 0.66, `.moai/reports/t1240/plan-audit-iter1.md`) resolved. D1 Codex lanes never re-lease a card they cannot advance (REQ-SD-025). D2 integration through the hold and the integration worktree; parent checkout never changes branch (REQ-SD-023). D3 lane `startup` always injects the rule (REQ-SD-019). D4 MCP tools resolve the caller's tree from `project_root` (REQ-SD-024). D5 one lane predicate (REQ-SD-015). D6 harness identified by `MOAI_KANBAN_BACKEND`; Codex merge edges refused on every path (REQ-SD-025). D7 queue guard is a read-only allowlist. D8/D12 reconciliation rows and t1257 hand-off list (§E); `next` reports PR/landed state. D9 one wording source for Codex refusals. D10/D11 AC coverage. O3 `depends_on` narrowed to F1; t1256 gated by REQ-SD-001 only. AC count 27 → 25. |
| 0.3.0 | 2026-09-27 | manager-spec | Plan-audit iteration 2 (FAIL 0.79, `.moai/reports/t1240/plan-audit-iter2.md`), narrow fix only. N1: refusal predicate widened to lane label or Codex backend (both in the frozen allowlist); admission stays on the marker (REQ-SD-015/016/017). N2: `complete` refuses a window whose branch source is the caller's own tree; github-flow parent constraint recorded (REQ-SD-023, §E.1). N3: REQ-SD-019 split by harness; Codex sessions receive their card through `MOAI_KANBAN_CARD` (REQ-SD-003). N4: cc backend citations corrected. REQ and AC counts unchanged (25 / 25). |
| 0.4.0 | 2026-09-27 | manager-spec | Plan-audit iteration 3 (FAIL 0.83, `.moai/reports/t1240/plan-audit-iter3.md`), leader-approved override fix, R1-R3 only. R1: MCP/CLI tool naming moved to the Claude-lane rule; the Codex rule names only `moai factory stage` / `complete` (REQ-SD-019). R2: REQ-AP-011 widening reconciled (design.md §7 row, plan.md file table, AC-SD-017 cases). R3: `complete` also refuses when the window's branch equals the card's branch or its tree equals the card worktree (REQ-SD-023). Leader instruction: two unresolved doctrine conflicts recorded as open decisions OD-1 (lane queue promotion) and OD-2 (Claude-lane self-integration), decided by the leader at run Kickoff; REQ-SD-008, -013, -015, -019, -023 made conditional on them. REQ and AC counts unchanged (25 / 25). |
| 0.5.0 | 2026-09-27 | manager-spec | Leader final decision closing the plan as PASS-WITH-DEBT (plan-audit iteration 4, `.moai/reports/t1240/plan-audit-iter4.md`; no further audit). OD-1 (lane queue promotion) and OD-2 (Claude-lane self-integration) are both PERMITTED, scoped to the self-dispatch lane mode only; the card text is the operator's prior authorization — "카드를 스스로 임대(moai factory next [--wait])", "Claude 는 EnterWorktree→처리→통합→ExitWorktree(keep)", "Codex 는 … 자가 통합 없이 merge-ready". The leader withdrew the "unresolved conflict" instruction as its own error. C1-C4 resolved by removing the OD conditionals; no runtime key — the self-dispatch lane mode itself is the discriminator. Doctrine amendments added as run/sync deliverables inside REQ-SD-015 and AC-SD-015 (kanban-dispatch.md and its template twin; CLAUDE.local.md §4.1). REQ and AC counts unchanged (25 / 25). |
| 0.5.1 | 2026-09-27 | manager-spec | Leader narrow edit (no re-audit): `.claude/rules/local/gitflow-lane-protocol.md` §6 (line 81, "병합 이후 — 레인은 카드를 스스로 고르지 않는다"; local-only, no template mirror) becomes the fourth doctrine deliverable with the same self-dispatch-lane exception (REQ-SD-015, AC-SD-015, plan.md file table); the "handed to the operator" notes are removed. REQ and AC counts unchanged (25 / 25). |

### Card text → SPEC vocabulary and interpretation

The card text (authoritative for scope) predates the t1256 rename. This SPEC uses the confirmed
vocabulary of SPEC-ROLE-NAMING-CODE-001 (`spec.md` REQ-RNC-002..007, -011, -012, -021; `design.md` §3,
committed text on `WT-role-naming-code`). Legacy spellings are **rejected, not aliased** there.

| Card text | This SPEC | Authority |
|---|---|---|
| `moai cc\|glm\|codex -f agent` | `moai cc\|glm\|codex -f lane` | REQ-RNC-002, -003 |
| 자가 배차 에이전트 (self-dispatching agent) | lane (self-dispatching) | t1256 design.md §3 |
| `MOAI_FACTORY_ROLE=agent` (lead revision) | `MOAI_FACTORY_ROLE` = the role-value constant (`lane` after t1256) | REQ-RNC-012; constants `internal/config/envkeys.go:323,332` |
| 에이전트 권한 (agent permission) | lane permission | same row |
| 리드 / lead | leader | REQ-RNC-006, -007 |
| numbered worker label | `lane-<n>` | REQ-RNC-004, -005 |
| `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS` (names) | kept; values follow `lane-<n>` | REQ-RNC-011 |
| "헤드리스 엔진 없음" vs t1242 design §5 ("headless `codex exec` worker") | interactive relaunch per card; no headless engine | card text wins (lead condition 3) |
| "워크트리 재사용 금지" | a card never uses another card's tree; a resumed card re-enters its own recorded tree | interpretation, REQ-SD-011 |
| "에이전트 권한에서 큐 변경 제외" vs t1256 Q3 "the pick path … gains none" and REQ-RNC-021 (pick help names lane self-dispatch) | F2 adds a lane guard to every queue-writing `todo` verb, `todo next <n>` included; a lane promotes only through `moai factory next` | Q3's "gains none" is a statement about t1256's scope; the card text binds F2. Kickoff confirms (plan.md §B B6); REQ-RNC-021's help text is handed to t1257 (§E.2) |
| (implicit) leader's pre-dispatch PR cross-check | `moai factory next` prints the leased card's PR and landed state | kanban-dispatch rule; reported, not skipped (REQ-SD-008) |

## §A Background

Factory F1 (SPEC-FACTORY-RECORD-001, completed) built the card record: a version-checked state machine
in `factory.db` with lease, heartbeat, evidence gates, and the operator commands `moai factory assign`,
`status`, and `decide` (`internal/cli/factory_card.go:81,213,302`). F1 left the lane-facing verbs
(`next`, `stage`, `complete`) as a Go API exercised by tests only (F1 spec.md §D).

F2 turns a lane from a session that waits for the leader's dispatch into a session that takes its own
next card. The record stays the source of truth: a lane moves a card only through the F1 transition API,
and the leader keeps every decision of consequence (queue admission, Kickoff approval, the push gate).

Measured facts that shape the SPEC (ledger in `research.md`):

- `moai codex` refuses every factory entry today (`internal/cli/codex_launcher.go:701`, refusal line
  `:743-744`, scanner `:754-768`) — t1242 left this refusal as the seam F2 replaces for the lane shape
  only (SPEC-CODEX-FACTORY-RETIRE-001 design.md §5).
- The factory role marker has name and value constants (`internal/config/envkeys.go:323,332`) and a guard
  that reads them (`internal/hook/contract_sign_guard.go:131-133`), but no production code sets it.
- The factory SessionStart notice is emitted only on `startup` and says nothing about taking a card
  (`internal/hook/session_start_factory.go:62-67`, `:201`).
- `moai integration acquire` records one integration window against the configured integration branch
  and the worktree that has it checked out (`internal/cli/integration.go:166-180`, `:218`).

## §B Requirements (GEARS)

REQ numbers are stable and never renumbered; REQ-SD-023..025 (added at v0.2.0) sit in §B.6 by topic,
so document order is not numeric order.

### B.1 Run ordering

- **REQ-SD-001** (Event-driven) — When the run phase of this SPEC starts, the run shall record in
  `progress.md` the local develop SHA it reads and whether SPEC-ROLE-NAMING-CODE-001 has landed on it —
  landed meaning the factory role-value constant equals `lane` and the launcher's `-f` role token equals
  `lane` on that tree — and **when** it has not landed, the run shall commit no change under a production
  path and shall halt with a blocker report to the leader.

### B.2 Launch

- **REQ-SD-002** (Event-driven) — When `moai cc -f lane` or `moai glm -f lane` is launched, the launcher
  shall start one interactive session in the parent checkout as the next free `lane-<n>`, with the
  factory role marker set to the role-value constant, the lane-label variable set to that `lane-<n>`, and
  the backend variable set to the launching backend.
- **REQ-SD-003** (Event-driven) — When `moai codex -f lane` is launched, the launcher shall, for each
  card in turn, lease the card, create or re-enter its worktree, run one interactive Codex session whose
  working directory is that worktree, and on that session's exit continue with the next card, stopping
  when no card is available; every Codex child shall carry the role marker, the lane label, the
  backend value that identifies the Codex harness, and the leased card's id in the card-identifier
  variable (`MOAI_KANBAN_CARD`, `config.EnvMoaiKanbanCard`), which is how the session learns which card
  it owns (REQ-SD-019).
- **REQ-SD-004** (Ubiquitous) — `moai codex` shall refuse every other factory entry shape — bare `-f`,
  `--factory`, `--factory-run`, and a numbered lane label — with one line, defined once, that carries
  `FACTORY_MODE_UNSUPPORTED_BACKEND`, names `moai codex -f lane` as the only Codex factory entry and
  `moai cc -f` / `moai glm -f` for the leader, and exits with code 1; legacy role tokens follow
  REQ-SD-021 instead.
- **REQ-SD-005** (Event-driven) — When a lane launch runs outside a git working tree, the launcher shall
  exit non-zero with one line saying a git repository is required, and write no registry, factory, or
  queue record.
- **REQ-SD-006** (Ubiquitous) — The lane cycle shall complete without a git remote: no step shall fetch
  or push, and a repository with no remote shall carry a card from lease to the lane's final state.
- **REQ-SD-007** (Unwanted) — The factory shall not process a card through a non-interactive engine
  (`claude -p`, `codex exec`, or any equivalent headless session).

### B.3 Lane verbs

- **REQ-SD-008** (Event-driven) — When a lane runs `moai factory next`, the verb shall acquire the lease
  on exactly one card for that lane through the F1 transition API, choosing in this order: a card
  assigned to this lane, then an operator-picked card assigned to no lane, then the oldest queued card,
  which it promotes to picked in the same operation (OD-1, §D); it shall never select a card assigned to another
  lane; it shall print the card id, its stage, its worktree name, and the card's pull-request and landed
  state as `moai todo pr` reports them; and **when** no card qualifies it shall print that no card is
  available and exit with status 3.
- **REQ-SD-009** (Event-driven) — When `moai factory next --wait` runs and no card qualifies, the verb
  shall re-check at a fixed interval until a card is leased or the wait bound elapses, and then behave as
  without `--wait`.
- **REQ-SD-010** (Event-driven) — When `moai factory next` is asked to act on a tree that is not the
  parent checkout — the process working directory on the CLI, the `project_root` argument on the MCP
  path — the verb shall refuse with a line naming the parent checkout path, exit non-zero, and change no
  record.
- **REQ-SD-011** (Event-driven) — When `moai factory next` leases a card with no recorded worktree, the
  verb shall create a new worktree for that card through the shared worktree materializer, whose
  directory name is the card id and whose branch carries the `WT-` prefix and does not contain the card
  id, and record its path on the card; **when** the card has a recorded worktree, the verb shall reuse
  that card's own tree; **when** the target directory already exists without belonging to that card, the
  verb shall refuse and change no record.
- **REQ-SD-012** (Event-driven) — When a lane runs `moai factory stage <card> <state>` with the evidence
  the F1 guard for that edge needs, the verb shall apply the transition through the F1 transition API
  with the lane's label as actor and renew the lease, and shall print an F1 refusal verbatim and exit
  non-zero when the API refuses.
- **REQ-SD-013** (Event-driven) — When a Claude-harness lane runs `moai factory complete <card>` on a
  card in `merge-ready`, the verb shall take the card through `merging` to `merged-local` by the F1
  merge gate, using as integration branch the branch the integration window records (REQ-SD-023; OD-2,
  §D).
- **REQ-SD-014** (Ubiquitous) — The MCP server shall expose `todo_add`, `todo_list`, `factory_next`,
  `factory_stage`, `factory_complete`, and `factory_decide`, each backed by the same implementation as
  its CLI counterpart (`moai todo add`, `moai todo list`, `moai factory next`, `stage`, `complete`,
  `decide`) and producing the same record changes and the same refusals for the same resolved tree.

### B.4 Lane permission boundary

- **REQ-SD-015** (Event-driven) — This SPEC uses two predicates, deliberately asymmetric. **Lane
  admission** holds when the factory role marker equals the role-value constant. **Lane refusal**
  holds when lane admission holds, **or** the lane-label variable is non-empty, **or** the backend
  variable equals the Codex value — the last two are both in the frozen Codex MCP `env_vars` allowlist,
  so refusal also works on the Codex MCP path without widening it. When a session for which lane
  refusal holds invokes any `moai todo` subcommand outside the read-only allowlist (bare `todo`, `list`,
  `history`, `why`, `pr`, `triage`) or the `todo_add` MCP tool, the call shall be refused with one line naming the lane boundary, exit non-zero, and leave the
  queue file byte-identical; the promotion inside `moai factory next` (REQ-SD-008) is the only queue
  write a lane performs; **when** a session for which lane admission does not hold invokes
  `moai factory next`, `stage`, or `complete`, the call shall be refused as not a lane session; and the
  doctrine that states the lane boundary shall carry the self-dispatch-lane exceptions this SPEC relies
  on — `.claude/rules/moai/workflow/kanban-dispatch.md` and its byte-identical template twin
  `internal/template/templates/.claude/rules/moai/workflow/kanban-dispatch.md` (an exception to "Promotion
  is the operator's act, always": a self-dispatch lane may lease a card, promoting a queued one, only
  through `moai factory next`), this repository's `CLAUDE.local.md` §4.1 (an exception to "the lane
  does not take the merge window itself" for Claude self-dispatch lanes), and this repository's local
  rule `.claude/rules/local/gitflow-lane-protocol.md` §6 (the same merge-window exception, and an
  exception to "a lane does not pick a card from the queue" for leasing through `moai factory next`,
  both for self-dispatch lanes) — each stating explicitly that
  every other queue mutation (`add`, `drop`, `done`, `edit`, and the rest) and `moai contract sign` stay
  forbidden to a lane.
- **REQ-SD-016** (Event-driven) — When a session for which lane refusal (REQ-SD-015) holds invokes
  `moai factory decide` or the `factory_decide` MCP tool, the call shall be refused and change no
  record.
- **REQ-SD-017** (Ubiquitous) — Every production site that stamps or compares the factory role marker
  shall use the role-marker name and value constants of SPEC-AUTONOMY-PRECONDITION-001 REQ-AP-012, never
  a string literal, so that the existing contract guard denies `moai contract sign --signer llm` in every
  lane session the launcher starts; the guard's role gate (SPEC-AUTONOMY-PRECONDITION-001 REQ-AP-011)
  shall deny wherever lane refusal (REQ-SD-015) holds — a widening in the deny direction only.
- **REQ-SD-018** (Unwanted) — A lane shall not modify the tracked files, the index, `HEAD`, or the
  checked-out branch of the parent checkout; card work shall happen in the card's worktree and merges in
  the integration worktree.

### B.5 Session cycle

- **REQ-SD-019** (Event-driven) — When SessionStart fires in a lane session with source `startup` (under
  every clear policy) or `clear`, the hook shall inject a rule chosen by the backend variable, in the
  session's conversation language (en, ko, ja, zh): for a Claude-harness lane (backend `claude` or
  `glm`), the next-card rule, naming each MCP tool of REQ-SD-014 together with its CLI equivalent —
  leave any worktree kept, take the next card from the parent checkout (stating that lane queue
  promotion through `moai factory next` is operator-authorized for the self-dispatch lane mode), enter its
  worktree, carry it through plan, run, and sync, integrate, leave the worktree kept, record completion,
  then follow the clear policy; for a Codex-harness lane (backend `gpt`), the owned-card
  rule, naming only the `moai factory stage` and `moai factory complete` CLI verbs — the card id read
  from the card-identifier variable (REQ-SD-003) and the current worktree, carry that card to
  `merge-ready`, then end the session — and it shall never name `moai factory next` or any MCP tool to
  a Codex session, nor instruct it to take or lease a card. Leader and non-factory sessions shall
  receive no rule.
- **REQ-SD-020** (Where) — **Where** a Claude-harness lane launch selects a clear policy, the lane shall
  apply it after each completion or lease release: `clear-each` (the default) prints one line asking the
  operator to `/clear`; `clear-when-full` asks for `/clear` only once the session's context-usage record
  shows usage at or above the model-specific handoff threshold and otherwise continues with the next
  card; `relaunch` asks the operator to end the session and has the supervising launcher start a fresh
  session for the next card.

### B.6 Integration, MCP tree, harness

- **REQ-SD-023** (Event-driven) — When a Claude-harness lane integrates a card (OD-2, §D), it shall
  first hold the
  integration window (`moai integration acquire`), perform the `--no-ff` merge inside the worktree that
  has the integration branch checked out, record `complete`, and then release the window itself
  (`complete` does not release it); **when** another session holds the window, `complete` shall refuse
  naming the holder and change no record; **when** the held window's branch source is the caller's own
  tree (`kanban.BranchSourceCaller` — the source `acquire` records when neither `--branch` nor a
  configured integration branch decided it), `complete` shall refuse, naming `--branch` as the remedy;
  **when** the held window's branch equals the card's recorded branch, or the window's tree equals the
  card's recorded worktree, `complete` shall refuse and change no record — together these make a card's
  own branch never serve as its integration branch; **when** the only tree holding the
  integration branch is the parent checkout or no tree holds it, `complete` shall refuse saying the
  integration worktree is not provisioned.
- **REQ-SD-024** (Ubiquitous) — The six MCP tools of REQ-SD-014 shall resolve the tree they act on from a
  `project_root` argument supplied by the caller (its `git rev-parse --show-toplevel`), under the same
  rule and rejection behavior the existing `project_root` tools use; `factory_next`, `factory_stage`, and
  `factory_complete` shall require it.
- **REQ-SD-025** (State-driven) — **While** a lane's backend variable identifies the Codex harness, the
  lane shall be refused the `merge-ready → merging` edge on every path (`complete`, `stage`, and their
  MCP tools), and `moai factory next` shall not select a card whose recorded stage or state that lane
  cannot advance (`merge-ready` or later), including one returned to `assigned` by lease expiry.

### B.7 Vocabulary and invariants

- **REQ-SD-021** (Ubiquitous) — Every surface this SPEC adds shall use the SPEC-ROLE-NAMING-CODE-001
  vocabulary — `lane`, `lane-<n>`, `leader` — and shall refuse every legacy role spelling (`worker`,
  `agent`, `worker-<n>`, `agent-<n>`, `lead`) with the REQ-RNC-003/-005/-007 message naming the
  canonical form, on `moai codex` as on `moai cc` and `moai glm`.
- **REQ-SD-022** (Unwanted) — The change shall not alter the schema of the factory database, the queue
  database, or the factory message broker, and shall not alter the `env_vars` allowlist of the generated
  Codex MCP server table (SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-020).

## §C Success Criteria

Acceptance criteria, Given-When-Then scenarios, edge cases, and closure gates: `acceptance.md`.
Traceability: every REQ-SD-0NN maps to at least one AC-SD-0NN (matrix in `acceptance.md` §C).

## §D Exclusions

### Out of Scope — controller and deciders (F3)

- `moai factory run`, lease reclamation by a controller, integration of Codex-lane cards that stop at
  `merge-ready`, CI polling, and the `pushed → ci-green → done` edges (F1 REQ-FR-006).
- Non-human Kickoff deciders (`llm`, `llm+jev`) and their receipts.

### Out of Scope — leader-side surfaces

- A Codex leader (`moai codex -f` without the lane token) — stays refused (REQ-SD-004).
- Provisioning the integration worktree; the leader provisions it (REQ-SD-023 refuses without it).
- Reversing the codex-led-run join refusal of SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-010.

### Out of Scope — vocabulary and documentation

- The rename itself (card t1256) and the documentation layer (card t1257), including every item of
  the hand-off list in §E.2.
- docs-site pages and README text for the new verbs (sync-phase work).

### Out of Scope — decided doctrine conflicts (resolved decision record)

Both conflicts below were decided by the leader on 2026-09-27 (final, closing the plan as
PASS-WITH-DEBT). Authority: the card text is the operator's prior authorization — "카드를 스스로 임대(moai factory next [--wait])", "Claude 는 EnterWorktree→처리→통합→ExitWorktree(keep)", "Codex 는 … 자가 통합 없이 merge-ready". The earlier instruction to treat them as unresolved was
withdrawn by the leader as its own error. There is no runtime configuration key for either decision:
the self-dispatch lane mode (`-f lane`, the card's `-f agent`) is itself the discriminator. Changing
the rest of the doctrine is out of scope; only the two exceptions REQ-SD-015 names are deliverables.

- **OD-1 — lane queue promotion: PERMITTED, self-dispatch lane mode only.** Leasing through
  `moai factory next` may promote the oldest queued card (REQ-SD-008). This is an exception to the
  `[HARD]` clause "Promotion is the operator's act, always" in
  `.claude/rules/moai/workflow/kanban-dispatch.md` § Entry into the board is an operator act, amended as
  a run/sync deliverable (REQ-SD-015). Every other queue mutation stays forbidden to a lane.
- **OD-2 — Claude-lane self-integration: PERMITTED, self-dispatch lane mode only.** A Claude lane
  holds the integration window and merges itself (REQ-SD-013, -023); Codex lanes still stop at
  `merge-ready`. This is an exception to this repository's `CLAUDE.local.md` §4.1 "the lane does not
  take the merge window itself", and to `.claude/rules/local/gitflow-lane-protocol.md` §6 ("a lane
  does not pick its own card"), both amended as run/sync deliverables (REQ-SD-015).

### Out of Scope — storage and wiring

- Any schema change, any new environment variable name, and any change to the Codex MCP `env_vars`
  allowlist (REQ-SD-022).
- Hook-based factory peer registration for Codex lanes (REQ-CFR-022 stays as is).

## §E Residual risk and hand-offs

### E.1 Residual risk

- **Lane variables can be unset by the agent.** A lane that clears `MOAI_FACTORY_ROLE` loses
  `next`/`stage`/`complete` but stays refused through its label and backend; one that also clears those
  escapes the refusals (t1245 spec.md §E C7). The launcher stamps them; nothing re-stamps them.
- **Codex MCP sees only the allowlisted environment.** The generated `env_vars` list carries the lane
  label and the backend variable but not the role marker (`internal/codexwiring/configtoml.go:21`), and
  REQ-SD-022 keeps it frozen. On the Codex MCP path lane admission is therefore false, so
  `factory_next`, `factory_stage`, and `factory_complete` refuse there (Codex lanes use the CLI forms,
  which run in the lane's own shell); lane refusal still holds through the label and the backend value,
  so `todo_add` and `factory_decide` are refused there too (REQ-SD-015/016).
- **Integration branch checked out in the parent.** Where the integration branch is the branch the
  parent checkout has checked out — the github-flow default, where it is `main` — no second worktree can
  hold it, so `complete` refuses (REQ-SD-023) and a Claude lane cannot self-integrate until the operator
  moves the parent off that branch and provisions an integration worktree. F2 does not provision one.
- **`project_root` is caller-supplied.** A lane inside a worktree that passes the parent path as
  `project_root` passes REQ-SD-010's check; the check stops mistakes, not a lane that misreports its
  own tree.
- **Codex `merge-ready` cards wait for F3.** REQ-SD-025 stops the livelock; it does not integrate them.

### E.2 Hand-off list for card t1257 (documentation layer)

1. The HARD "Promotion is the operator's act, always" clause — the self-dispatch-lane exception
   itself is now this SPEC's deliverable (REQ-SD-015) in `kanban-dispatch.md` and its template twin;
   t1257 keeps only its vocabulary pass over that text (template neutrality rules bind the twin).
2. The HARD pre-dispatch PR cross-check — for self-dispatched cards it is reported by `moai factory
   next` (REQ-SD-008) rather than read by the leader before dispatch.
3. REQ-RNC-021's pick help text ("promoted … by a lane's self-dispatch") — lanes are refused
   `todo next <n>` by REQ-SD-015; the help text should name `moai factory next` as the lane path.
4. `CLAUDE.local.md` §4.1 and `.claude/rules/local/gitflow-lane-protocol.md` §6 are amended by this
   SPEC (REQ-SD-015); both are local-only and outside t1257's template scope.

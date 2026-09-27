---
id: SPEC-FACTORY-SELF-DISPATCH-001
title: "Harness-neutral factory F2 — self-dispatching lane"
version: "0.1.0"
status: draft
created: 2026-09-27
updated: 2026-09-27
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli, internal/hook, internal/homestate, internal/config, internal/kanban"
lifecycle: spec-anchored
tags: "factory, factory-f2, self-dispatch, lane, leader, mcp, session-start, clear-policy, codex-relaunch, card-t1240"
tier: L
card: t1240
depends_on: [SPEC-FACTORY-RECORD-001, SPEC-ROLE-NAMING-CODE-001]
related_specs: [SPEC-FACTORY-RECORD-001, SPEC-AUTONOMY-PRECONDITION-001, SPEC-CODEX-FACTORY-RETIRE-001, SPEC-ROLE-NAMING-CODE-001]
---

# SPEC-FACTORY-SELF-DISPATCH-001 — Self-dispatching lane (Factory F2)

## HISTORY

| Version | Date | Author | Change |
|---|---|---|---|
| 0.1.0 | 2026-09-27 | manager-spec | Initial plan-phase draft for card t1240 on `WT-factory-self-dispatch` at base develop `ed506740b` (includes F1 t1239, t1242, t1245). Written in the vocabulary of SPEC-ROLE-NAMING-CODE-001 (card t1256, still unlanded; read at branch `WT-role-naming-code` `d39a1dc09`). The card text's own vocabulary is mapped in the table below. |

### Card text → SPEC vocabulary

The card text (authoritative for scope) predates the t1256 rename. This SPEC uses the confirmed
vocabulary of SPEC-ROLE-NAMING-CODE-001 (`spec.md` REQ-RNC-002..007, -011, -012; `design.md` §3 term
table, both read from the committed text on `WT-role-naming-code`). Legacy spellings are **rejected,
not aliased** there, so none of them appears below as an accepted input.

| Card text | This SPEC | Authority |
|---|---|---|
| `moai cc\|glm\|codex -f agent` | `moai cc\|glm\|codex -f lane` | REQ-RNC-002, -003 (`agent`/`worker` rejected) |
| 자가 배차 에이전트 (self-dispatching agent) | lane (self-dispatching) | design.md §3 row "Session that processes cards" |
| `MOAI_FACTORY_ROLE=agent` (and the lead revision's value) | `MOAI_FACTORY_ROLE` = the role-value constant, whose value is `lane` after t1256 | REQ-RNC-012; name/value constants from t1245 (`internal/config/envkeys.go:313-333`) |
| 에이전트 권한 (agent permission) | lane permission | same row |
| 리드 / lead | leader | REQ-RNC-006, -007 |
| numbered worker label | `lane-<n>` | REQ-RNC-004, -005 |
| `MOAI_FACTORY_WORKER`, `MOAI_FACTORY_WORKERS` (names) | kept unchanged; values follow `lane-<n>` | REQ-RNC-011 |
| "헤드리스 엔진 없음" vs t1242 design §5 ("headless `codex exec` worker") | interactive relaunch per card (`codex -C <worktree>`); no headless engine | card text wins (lead condition 3) |

## §A Background

Factory F1 (SPEC-FACTORY-RECORD-001, completed) built the card record: a version-checked state machine
in `factory.db` with lease, heartbeat, evidence gates, and the operator commands `moai factory assign`,
`status`, and `decide` (`internal/cli/factory_card.go:81,213,302`). F1 deliberately left out the
lane-facing verbs (`next`, `stage`, `complete`) — they exist only as a Go API exercised by tests
(F1 spec.md §D "Out of Scope — F2 self-dispatch and launcher").

F2 turns a lane from a session that waits for the leader's dispatch into a session that takes its own
next card. The record stays the source of truth: a lane moves a card only through the F1 transition API,
and the leader keeps every decision of consequence (queue admission, Kickoff approval, the push gate).

Three measured facts shape the SPEC (full ledger in `research.md`):

- `moai codex` refuses every factory entry today (`internal/cli/codex_launcher.go:701`, refusal line
  `:740-741`, scanner `:754-768`) — t1242 retired the old codex factory and left this refusal as the
  seam F2 replaces for the lane shape only (SPEC-CODEX-FACTORY-RETIRE-001 design.md §5).
- The factory role marker has a name and value constant (`internal/config/envkeys.go:323,332`) and a
  guard that reads it (`internal/hook/contract_sign_guard.go:131-133`), but no production code sets it
  (every `Setenv(config.EnvFactoryRole` is in a `_test.go` file). The guard therefore denies nothing in a
  real session until F2 stamps the marker (t1245 spec.md §E C7).
- The factory SessionStart notice is emitted only on `startup` (`internal/hook/session_start_factory.go:62-67`),
  so a lane session that is cleared between cards receives nothing that tells it to take the next card.

## §B Requirements (GEARS)

### B.1 Run ordering

- **REQ-SD-001** (Event-driven) — When the run phase of this SPEC starts, the run shall record in
  `progress.md` the local develop SHA it reads and whether SPEC-ROLE-NAMING-CODE-001 has landed on it —
  landed meaning the factory role-value constant equals `lane` and the launcher's `-f` role token equals
  `lane` on that tree — and **when** it has not landed, the run shall edit no production file and shall
  halt with a blocker report to the leader.

### B.2 Launch

- **REQ-SD-002** (Event-driven) — When `moai cc -f lane` or `moai glm -f lane` is launched, the launcher
  shall start one interactive session in the parent checkout as the next free `lane-<n>`, with the
  factory role marker set to the role-value constant and the lane-label variable set to that `lane-<n>`.
- **REQ-SD-003** (Event-driven) — When `moai codex -f lane` is launched, the launcher shall, for each
  card in turn, lease the card, create its worktree, run one interactive Codex session whose working
  directory is that worktree, and on that session's exit continue with the next card, stopping when no
  card is available; every Codex child it starts shall carry the role marker and the lane label.
- **REQ-SD-004** (Ubiquitous) — `moai codex` shall keep refusing every other factory entry shape — bare
  `-f`, `--factory`, `--factory-run`, a numbered lane label, and every legacy role token — with the
  existing `FACTORY_MODE_UNSUPPORTED_BACKEND` line and exit code 1.
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
  which it promotes to picked in the same operation; it shall print the card id, its stage, and its
  worktree name, and **when** no card qualifies it shall print that no card is available and exit with a
  distinct non-zero status.
- **REQ-SD-009** (Where) — **Where** `--wait` is given, `moai factory next` shall, while no card
  qualifies, re-check at a fixed interval until a card is leased or the wait bound elapses, and then
  behave as without `--wait`.
- **REQ-SD-010** (Event-driven) — When `moai factory next` runs from a working directory that is not the
  parent checkout, the verb shall refuse with a line naming the parent checkout path, exit non-zero, and
  change no record.
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
  card it has integrated, the verb shall take the card to `merged-local` through the F1 merge gate using
  the configured worktree base branch as the integration branch; **when** a Codex-harness lane runs it,
  the verb shall take the card no further than `merge-ready` and refuse the merge edges.
- **REQ-SD-014** (Ubiquitous) — The MCP server shall expose `todo_add`, `todo_list`, `factory_next`,
  `factory_stage`, `factory_complete`, and `factory_decide`, each backed by the same implementation as
  its CLI counterpart (`moai todo add`, `moai todo list`, `moai factory next`, `stage`, `complete`,
  `decide`) and producing the same record changes and the same refusals.

### B.4 Lane permission boundary

- **REQ-SD-015** (Event-driven) — When a session carrying the lane role invokes a queue mutation —
  through `moai todo` verbs that write the queue or through the `todo_add` MCP tool — the call shall be
  refused with one line naming the lane boundary, exit non-zero, and leave the queue file byte-identical;
  the promotion inside `moai factory next` (REQ-SD-008) is the only queue write a lane performs.
- **REQ-SD-016** (Event-driven) — When a session carrying the lane role invokes `moai factory decide` or
  the `factory_decide` MCP tool, the call shall be refused and change no record.
- **REQ-SD-017** (Ubiquitous) — Every production site that stamps or compares the factory role marker
  shall use the role-marker name and value constants of SPEC-AUTONOMY-PRECONDITION-001 REQ-AP-012, never
  a string literal, so that the existing contract guard denies `moai contract sign --signer llm` and
  `moai contract decide` in every lane session the launcher starts.
- **REQ-SD-018** (Unwanted) — A lane session shall not modify the tracked files, the index, `HEAD`, or
  the checked-out branch of the parent checkout; all card work shall happen in the card's worktree.

### B.5 Session cycle

- **REQ-SD-019** (Event-driven) — When SessionStart fires with source `clear` in a session carrying the
  lane role — or with source `startup` in a lane session launched under the relaunch policy — the hook
  shall inject the next-card rule (take the next card from the parent checkout, enter its worktree,
  carry it through plan, run, and sync, integrate or stop at merge-ready per harness, leave the worktree
  kept, record completion, then follow the clear policy) in the session's conversation language (en, ko,
  ja, zh), naming each MCP tool together with its CLI equivalent; leader and non-factory sessions shall
  receive no such rule.
- **REQ-SD-020** (Where) — **Where** a Claude-harness lane launch selects a clear policy, the lane shall
  apply it after each completion or lease release: `clear-each` (the default) prints one line asking the
  operator to `/clear`; `clear-when-full` asks for `/clear` only once the session's measured context
  usage crosses the model-specific handoff threshold and otherwise continues with the next card;
  `relaunch` has the launcher start a fresh session for the next card.

### B.6 Vocabulary and invariants

- **REQ-SD-021** (Ubiquitous) — Every surface this SPEC adds shall use the SPEC-ROLE-NAMING-CODE-001
  vocabulary — `lane`, `lane-<n>`, `leader` — and shall accept no legacy role spelling (`worker`, `agent`,
  `worker-<n>`, `agent-<n>`, `lead`) as input.
- **REQ-SD-022** (Unwanted) — The change shall not alter the schema of the factory database, the queue
  database, or the factory message broker, and shall not alter the `env_vars` allowlist of the generated
  Codex MCP server table (SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-020).

## §C Success Criteria

Acceptance criteria, Given-When-Then scenarios, edge cases, and closure gates: `acceptance.md`.
Traceability: every REQ-SD-0NN maps to at least one AC-SD-0NN (matrix in `acceptance.md` §C).

## §D Exclusions

### Out of Scope — controller and deciders (F3)

- `moai factory run`, lease reclamation by a controller, automatic merging of Codex-lane cards that stop
  at `merge-ready`, CI polling, and the `pushed → ci-green → done` edges (F1 REQ-FR-006).
- Non-human Kickoff deciders (`llm`, `llm+jev`) and their receipts.

### Out of Scope — leader-side surfaces

- A Codex leader (`moai codex -f` without the lane token) — stays refused (REQ-SD-004).
- Changing the leader's dispatch paths (`internal/cli/gtd.go`, `internal/cli/goal.go`) beyond what F1
  already mirrors into the record.
- Reversing the codex-led-run join refusal of SPEC-CODEX-FACTORY-RETIRE-001 REQ-CFR-010.

### Out of Scope — vocabulary and documentation

- The rename itself (SPEC-ROLE-NAMING-CODE-001, card t1256) and the documentation layer (card t1257),
  including the documented HARD promotion clause that REQ-SD-008's queue promotion relies on the
  operator's Q3 ruling to relax.
- docs-site pages and README text for the new verbs (sync-phase work).

### Out of Scope — storage and wiring

- Any schema change (REQ-SD-022), any new environment variable name, and any change to the Codex MCP
  `env_vars` allowlist.
- Hook-based factory peer registration for Codex lanes (REQ-CFR-022 stays as is).

## §E Residual risk

- **The role marker can be unset by the agent.** An agent that clears `MOAI_FACTORY_ROLE` in a Bash call
  escapes the guards that key on it (t1245 spec.md §E C7). The launcher stamps it; nothing re-stamps it.
- **Codex MCP sees only the allowlisted environment.** The generated `env_vars` list carries the lane
  label variable but not the role marker (`internal/codexwiring/configtoml.go:21`), and REQ-SD-022 keeps
  that list frozen. The MCP-side lane check therefore reads the lane label as well as the marker
  (design.md §5); a session that carries neither is treated as not a lane.
- **`merge-ready` holds a lease.** A Codex lane that stops at `merge-ready` leaves a lease that expires;
  F1 REQ-FR-013 then returns the card to `assigned` with its stage preserved. Integration of those
  cards is F3's (or the leader's reassignment); F2 does not close that gap.
- **Queue promotion by a lane contradicts the documented HARD promotion clause** until card t1257 amends
  it. The operator ruled on 2026-09-26 that lane self-dispatch may promote a queued card
  (SPEC-ROLE-NAMING-CODE-001 plan.md §B Q3); the documentation still says otherwise.

---
id: SPEC-CHAIN-CORE-001
title: "Worktree Session Origin-Trail Chain — Record, Query, and Completion Hook"
version: "0.1.0"
status: completed
created: 2026-08-13
updated: 2026-08-13
author: spec-chain-core
priority: P1
phase: "v3.2.0 target"
module: "internal/chain, internal/cli/chain.go, internal/hook, internal/config/envkeys.go"
lifecycle: spec-anchored
tags: "chain, worktree, lineage, session-continuity, overlay"
tier: L
---

# SPEC-CHAIN-CORE-001 — Worktree Session Origin-Trail Chain

## HISTORY

- **2026-08-13** — Initial draft authored by `manager-spec` (spec-chain-core agent).
  Root-cause evidence: 11-agent diagnostic workflow `wf_0aed4941-ad6`
  (2026-08-13) verified the missing data structure across
  `internal/session`, `internal/factory`, `internal/worktree`,
  `internal/hook/handoff`. Winner design: Origin-Trail Chain
  (append-only JSONL ledger + mechanical completion-event hook).
  8 design policies locked via user AskUserQuestion (2026-08-13).

---

## §A. Outcome Hypothesis

**The pain this SPEC solves**: During large multi-milestone projects, once the
worktree nesting depth reaches 2–3, the maintainer loses track of three things:
where they came from (origin), what they just finished (completion), and where
to resume (next action). The manual paste-ready handoff is exactly what they
want to escape.

**Root cause** (verified by diagnostic workflow `wf_0aed4941-ad6`): a MISSING
DATA STRUCTURE. A grep across `internal/session`, `internal/factory`,
`internal/worktree`, and `internal/hook/handoff` returned ZERO matches for
`parent_session_id`, `origin_chain`, `depth`, or `resume_target`. Every
persistence primitive is FLAT:

- The session registry `Entry` (registry.go:86-95) carries `CWD` as the ONLY
  worktree link — a flat string with no parent, no depth, no origin chain.
- The handoff `pending.json` (pending.go:118) is SINGULAR per-project and
  OVERWRITTEN on each save — a depth-3 save clobbers the depth-1 record.
- The worktree `Snapshot` (state_guard.go:57-65) captures only divergence
  state, then is discarded — no parent field.

**What changes**: An append-only JSONL lineage tree is persisted at
`.moai/state/chain/events.jsonl`, populated automatically at every spawn
boundary. The user re-enters a depth-3 worktree after `/clear` and immediately
sees their origin chain, last-completed milestone, and resume target — no
grep, no scrollback archaeology. The three pain questions become named fields
on a structured record:

- **Origin** → `origin_chain` + `depth` + `parent_node_id`
- **What just completed** → `last_completed_milestone` + completion edges
- **Where to resume** → `resume_target` + `resume_command`

**How this SPEC is independently shippable**: This SPEC delivers the chain
record, query CLI, completion hook, and SessionStart banner with NO dependency
on kanban mode (`-k`), factory mode (`-f`), or the manager-lead agent. Those
are later phases (P3–P7) that consume this chain as a substrate. The chain
solves depth-3 amnesia ALONE.

---

## §B. Background — Why a Data Structure, Not a Socket

The maintainer's intuition ("a socket would preserve memory where
paste-ready handoff fails") is a CORRECT DIAGNOSIS of the symptom but a
PARTIAL fix for the cause. The diagnostic verified there is no field to
populate — even a perfect live socket would have nothing to carry. Building
transport before the payload is building a pipe with no water.

The existing "cross-session transport" is a mix of filesystem polling
(active-sessions.json), one-shot SessionStart stderr reminders, the singular
fire-once handoff pending.json, and Claude Code's opaque native SendMessage.
None of these carries lineage metadata because no record has lineage fields.

This SPEC puts the structured record FIRST. The chain ledger IS the payload;
later phases may add push transport on top of it. The remembering is done by
the record — the socket only delivers it faster.

---

## §C. Scope Boundaries

### In Scope

- `WorktreeNode` struct and append-only JSONL store (`internal/chain/`)
- Automatic node creation at spawn boundaries (`moai cc -w`,
  `EnterWorktree`, `Agent(isolation: "worktree")`)
- `MOAI_CHAIN_NODE_ID` environment constant in `envkeys.go`
- CLI subcommands: `moai chain {status, lineage, back, list, prune}`
- Mechanical completion hook: `.claude/hooks/moai/chain-event.sh`
  (SubagentStop / phase-transition → append ledger, no leaf cooperation)
- SessionStart lineage banner (re-inject node ID after `/clear`,
  emit depth + parent + resume system-reminder)
- Overlay join by SessionID — Entry struct UNTOUCHED (REQ-COORD-024)
- Template-mirror: state-dir scaffold + hook + CLI command registration

### Non-Scope (see §I for detail)

- Kanban mode (`-k`), factory mode (`-f`), `/moai:todo` inbox, multi-record
  handoff fix, manager-lead rewrite, depth-ceiling CI guard rewrite, live push
  channel — all are later phases (P3–P7), separate future SPECs.

---

## §D. Requirements (GEARS)

### REQ-CHAIN-001 — WorktreeNode Struct (Ubiquitous)

The chain package shall define a `WorktreeNode` struct carrying the following
named fields: `node_id`, `parent_node_id`, `depth`, `origin_chain` (ordered
root-to-leaf NodeID list), `worktree_path`, `session_id`, `spec_id`,
`milestone`, `entered_at`, `exited_at`, `last_completed_milestone`,
`resume_target` (one-line human intent), and `resume_command` (single primary
action).

### REQ-CHAIN-002 — Append-Only JSONL Store (Event-Driven)

**When** a chain lifecycle event occurs (node-enter, node-update,
completion-edge), the chain store shall append exactly one JSONL line to
`.moai/state/chain/events.jsonl`, and the store shall NOT overwrite, truncate,
or seek-replace any existing line in that file.

### REQ-CHAIN-003 — Corrupt-Line Tolerance (Event-Driven)

**When** the chain store encounters a malformed JSONL line during a read
operation, the store shall skip that line, emit a diagnostic log warning
identifying the line number and parse error, and continue processing the
remaining lines without aborting the read.

### REQ-CHAIN-004 — CWD-Collision Resolution (Event-Driven)

**When** two or more nodes in the ledger share the same `worktree_path`, the
chain store shall resolve the current node using the `(worktree_path,
session_id)` pair as the primary key. If that pair is still ambiguous, the
store shall fall back to the session registry's PID-liveness probe. If still
indeterminate, the store shall surface a diagnostic rather than silently
picking.

### REQ-CHAIN-005 — Spawn-Boundary Node Creation (Event-Driven)

**When** a session enters a worktree via `moai cc -w`, the `EnterWorktree`
path, or an `Agent(isolation: "worktree")` spawn, the chain population path
shall create a new `WorktreeNode` whose `parent_node_id` is derived from the
`MOAI_CHAIN_NODE_ID` environment variable of the spawning context, and whose
`depth` is one greater than the parent node's depth.

### REQ-CHAIN-006 — MOAI_CHAIN_NODE_ID Env Constant (Ubiquitous)

The config package shall expose `EnvChainNodeID = "MOAI_CHAIN_NODE_ID"` as a
constant in `internal/config/envkeys.go`, and every code site that reads this
environment variable shall reference the constant rather than inlining the
string literal.

### REQ-CHAIN-007 — `moai chain status` CLI (Event-Driven)

**When** the user invokes `moai chain status`, the CLI shall resolve the
current node by CWD and print its `depth`, `parent_node_id`, `spec_id`,
`milestone`, `last_completed_milestone`, and `resume_target` as a
human-readable summary.

### REQ-CHAIN-008 — `moai chain lineage` CLI (Event-Driven)

**When** the user invokes `moai chain lineage`, the CLI shall print the full
root-to-leaf `origin_chain`, listing each ancestor node's `node_id`,
`worktree_path`, `spec_id`, `milestone`, and `entered_at`.

### REQ-CHAIN-009 — `moai chain back` CLI (Event-Driven)

**When** the user invokes `moai chain back`, the CLI shall print the parent
node's `resume_target` and `resume_command` — the exact next action to resume
at the parent level.

### REQ-CHAIN-010 — `moai chain list` CLI (Event-Driven)

**When** the user invokes `moai chain list`, the CLI shall enumerate all nodes
in the ledger with their `depth`, `session_id`, `worktree_path`, `spec_id`,
and staleness status (active / stale / exited).

### REQ-CHAIN-011 — `moai chain prune` CLI (Event-Driven)

**When** the user invokes `moai chain prune`, the CLI shall fold exited nodes
older than the configured threshold (default: 30 days OR 10 MB, whichever
fires first) into a node-summary snapshot and remove their individual event
lines from the active read path, while retaining them in the audit trail
under `.moai/state/chain/archived/`.

### REQ-CHAIN-012 — Mechanical Completion Hook (Event-Driven)

**When** the `chain-event.sh` hook fires on `SubagentStop` or a
phase-transition signal, the hook shall append a structured completion edge
`{parent_node, child_node, completed_milestone, completed_at,
next_resume_target}` to the chain ledger without requiring any leaf-agent
cooperation — the ledger stays current even if the spawning session has
crashed or been cleared.

### REQ-CHAIN-013 — SessionStart Lineage Banner (Event-Driven)

**When** `SessionStart` fires and `MOAI_CHAIN_NODE_ID` is absent from the
environment (post-`/clear` env loss), the SessionStart handler shall
re-inject the node ID by resolving it from the chain ledger via
`worktree_path` and `session_id`, and emit a lineage system-reminder
containing the current `depth`, a parent-chain summary, and the
`resume_target`. The handler shall be time-boxed and shall degrade to a no-op
banner on timeout (fail-open).

### REQ-CHAIN-014 — Overlay Join, Entry Freeze (Ubiquitous)

The chain ledger shall NEVER mutate the session registry `Entry` struct.
Lineage data is overlaid by joining on `SessionID` at read time only,
respecting the `REQ-COORD-024` schema freeze. No field is added to, removed
from, or reordered within the `Entry` struct.

### REQ-CHAIN-015 — Heartbeat-Based Staleness (State-Driven)

**While** a node's associated session `LastHeartbeat` is older than the
staleness threshold (default: 15 minutes), the chain display shall mark the
node as stale, exclude it from the active-chain view, and retain it in the
audit trail. The stale threshold reuses the registry's existing
`LastHeartbeat` field via the overlay join — no new heartbeat mechanism is
introduced.

### REQ-CHAIN-016 — Single-Host v1 (Capability Gate)

**Where** the chain ledger is filesystem-local (single-host v1), cross-machine
worktree lineage shall be a documented limitation: a remote worktree (via
Remote Control or another host) cannot append to the local `events.jsonl`.
This limitation shall be surfaced in `moai chain status` output when a remote
CWD is detected.

### REQ-CHAIN-017 — Flag-Agnostic (Ubiquitous)

The chain record, store, CLI subcommands, and completion hook shall function
independently of `-k` (kanban mode), `-f` (factory mode), and the
`manager-lead` agent. No code path in this SPEC's implementation shall read,
gate on, or require any of these flags or agents.

### REQ-CHAIN-018 — AskUserQuestion Boundary (Unwanted)

The chain CLI code and the `chain-event.sh` hook code shall NOT invoke
`AskUserQuestion` or `mcp__askuser__*`. The orchestrator-only interaction
boundary is preserved, and a static grep guard test
(`TestNoAskUserQuestionInChain`) shall enforce that zero matches exist for
these tokens in the chain package and hook source.

### REQ-CHAIN-019 — Template-Mirror (Ubiquitous)

The chain state-dir scaffold (`.moai/state/chain/`), the `chain-event.sh`
hook wrapper script, and the `moai chain` CLI command registration shall be
template-mirrored to `internal/template/templates/` per the Template-First
Rule. The template sources shall pass the template-neutrality CI guard
(no SPEC IDs, no internal commit SHAs, no macOS-bias paths).

### REQ-CHAIN-020 — Test Isolation (State-Driven)

**While** chain store tests execute, all file paths shall use `t.TempDir()`.
No test shall write to, read from, or create the project's real
`.moai/state/chain/` directory. Tests that need a registry fixture shall
construct their own `Registry` via `NewRegistry` bound to a `t.TempDir()` path.

### REQ-CHAIN-021 — SessionID Two-Phase Backfill (Event-Driven)

**When** a `WorktreeNode` is created at the spawn boundary
(REQ-CHAIN-005), the node shall be created as a skeleton with `session_id`
empty, and shall be backfilled with the child's runtime-assigned
`session_id` via a `node-update` event when the child's `SessionStart` fires.
This two-phase protocol is the precondition for REQ-CHAIN-013's
re-resolution by `(worktree_path, session_id)` after `/clear`. See design.md
§4 for the full protocol.

---

## §E. Constraints

| Constraint | Source | Impact |
|---|---|---|
| `Entry` struct schema frozen (REQ-COORD-024) | registry.go:86-95 | Chain is Overlay-only — join by SessionID at read, never mutate |
| Hook subagent boundary (CLAUDE.md §8) | askuser-protocol.md | Chain CLI/hook must NOT invoke AskUserQuestion |
| Hook timeout 5s default (CLAUDE.local.md §7) | hook-development.md | SessionStart lineage banner must be time-boxed + fail-open |
| Template-First Rule (CLAUDE.local.md §2) | Template-First cycle | All distributable artifacts mirrored to templates/ + `make build` |
| Template neutrality (CLAUDE.local.md §25) | template-internal-isolation | No SPEC IDs, commit SHAs, or internal dates in template source |
| Env-var SSOT (CLAUDE.local.md §14) | envkeys.go | `MOAI_CHAIN_NODE_ID` is a constant, never an inline string |
| Go test isolation (CLAUDE.local.md §6) | t.TempDir | No OTEL env vars in parallel tests; all paths under t.TempDir |

---

## §F. Dependencies

| Dependency | Status | Notes |
|---|---|---|
| `internal/session.Registry` | Existing (frozen) | Overlay joins `Entry` by `SessionID` at read; Entry is NOT mutated |
| `internal/config.envkeys.go` | Existing | `EnvChainNodeID` constant added to the existing file |
| `internal/atomicfile` | Existing | Store append uses file-append, NOT the atomic-write-overwrite pattern |
| `internal/hook` SessionStart handler | Existing (extended) | Lineage banner hooks into the existing `sessionStartHandler.Handle` flow |
| Cobra CLI framework | Existing | `moai chain` subcommand registered via existing `rootCmd` pattern |

No new external Go dependencies are introduced.

---

## §G. Risks

| Risk | Likelihood | Mitigation |
|---|---|---|
| SessionStart latency regression from ledger read | Medium | Time-boxed via `context.WithTimeout`; fail-open to no-op banner on timeout |
| Append-only JSONL grows unbounded over time | Medium | `moai chain prune` folds old nodes into summary snapshot (REQ-CHAIN-011) |
| CWD-collision ambiguity at SessionStart re-anchor | Low | `(worktree_path, session_id)` pair resolves common case; PID-liveness fallback; diagnostic on indeterminate |
| Overlay join perf on large node counts | Low | Prune keeps active path bounded; reads are eventually consistent (registry precedent) |
| `MOAI_CHAIN_NODE_ID` env loss after `/clear` | High (expected) | REQ-CHAIN-013 re-injects from ledger via CWD+session_id resolution |
| Concurrent-append interleaving on lines > PIPE_BUF | Low | O_APPEND atomicity holds only for writes ≤ PIPE_BUF (512–4096 bytes); a WorktreeNode JSONL line with 13 fields + full paths CAN exceed 4 KB. Mitigation: corrupt-line tolerance (REQ-CHAIN-003) skips interleaved garbage; keep node lines compact where feasible |

---

## §H. Cross-References

- `internal/session/registry.go` — `Entry` struct (frozen, REQ-COORD-024);
  `FormatStderrReminder` pattern (model for chain banner)
- `internal/hook/handoff/pending.go:118` — the clobber flaw this SPEC's
  append-only design structurally prevents (fixed in Phase 4, not this SPEC)
- `internal/hook/session_start.go` — where the lineage banner hooks
- `internal/worktree/state_guard.go` — existing divergence-detection Snapshot
  (distinct from this SPEC's lineage tree; NOT duplicated)
- `internal/config/envkeys.go` — where `EnvChainNodeID` constant is added
- `.claude/rules/moai/workflow/worktree-integration.md` — L1/L2/L3 terminology
- `.claude/rules/moai/workflow/session-handoff.md` — Block 0 Worktree-Anchored
  Resume (single-level pattern this SPEC extends to a chain)
- CLAUDE.local.md §2 (Template-First), §6 (t.TempDir), §14 (envkeys),
  §25 (template neutrality)
- Diagnostic evidence: `wf_0aed4941-ad6` (root-cause + winner design)

---

## §I. Out of Scope

### Out of Scope — Kanban Mode (`-k`)

- The `-k` flag, `factory.go` → `kanban.go` rename, `MOAI_FACTORY` →
  `MOAI_KANBAN` env migration, derived board, `moai kanban pull` CLI
- These are Phase 5 (SPEC-KANBAN-MODE-* future SPECs)
- This SPEC is flag-agnostic (REQ-CHAIN-017)

### Out of Scope — `/moai:todo` Inbox

- The continuous-capture inbox command, `backlog.jsonl` card store,
  chain-stamped todo appends
- This is Phase 3 (SPEC-TODO-INBOX-* future SPEC)
- The chain stamp mechanism (`MOAI_CHAIN_NODE_ID`) is delivered HERE; the todo
  consumer is a separate SPEC

### Out of Scope — Multi-Record Handoff Fix

- Replacing the singular overwritten `pending.go:118` `atomicWriteFile` with
  per-node handoff records under `.moai/state/chain/handoff/<node_id>.json`
- This is Phase 4 (SPEC-CHAIN-HANDOFF-* future SPEC)
- This SPEC's append-only store structurally prevents clobbering, but the
  handoff-specific per-node namespace is deferred

### Out of Scope — `manager-lead` Rewrite + Depth-Ceiling CI Guard

- The `manager-lead.md` agent file rewrite (Tier L → chain-aware kanban lead)
- The `internal/template/manager_lead_depth_test.go` depth-2 seal rewrite as a
  configurable depth-ceiling CI check (default 3)
- These are Phase 6 (SPEC-KANBAN-LEAD-* future SPEC, same PR)
- This SPEC does NOT touch `.claude/agents/moai/manager-lead.md` or the
  depth-seal test file

### Out of Scope — Live Push Channel

- A UNIX-domain socket, fswatch/inotify tail, or real-time "what just
  completed" push notification
- This is Phase 7 (optional, measure-first)
- The chain-event.sh hook (REQ-CHAIN-012) records completion mechanically NOW;
  transport for push delivery is deferred

### Out of Scope — Web Dashboard

- No WebSocket server, no HTTP endpoint, no browser UI for the chain
- CLI + hook surfaces only (locked policy 8)
- The existing `internal/web/board.go` is untouched

### Out of Scope — Automatic Compaction

- Automatic (scheduled) compaction that folds old nodes without explicit
  user invocation
- Policy 5 locks "auto + manual"; for v1 scope discipline, the "auto" surface
  is served by the staleness display (REQ-CHAIN-015) — stale nodes
  (`LastHeartbeat` > 15min) are auto-excluded from the active VIEW without
  compaction
- Scheduled folding (e.g., a threshold-triggered background prune) is
  DEFERRED to a follow-up SPEC; manual `moai chain prune` (REQ-CHAIN-011) is
  the v1 compaction surface

### Out of Scope — Cross-Machine Lineage

- Remote worktree support via Remote Control or distributed sync
- Single-host v1 is a documented limitation (REQ-CHAIN-016)
- A sync substrate does not exist today and is not introduced here

### Out of Scope — Registry Entry Mutation

- Adding `parent_session_id`, `depth`, or `origin_chain` directly to the
  `Entry` struct
- This would supersede REQ-COORD-024 and cascade across every Entry consumer
- The Overlay join (REQ-CHAIN-014) is the non-breaking alternative

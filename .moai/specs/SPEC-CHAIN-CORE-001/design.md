# SPEC-CHAIN-CORE-001 — Design Document

> **Artifact role**: Tier L architecture document. Defines HOW the chain is
> structured — the topology, the overlay-join design, the population paths,
> the store internals, and the hook write path. Cross-references spec.md §D
> (WHAT/WHY) and plan.md §F (milestones).

---

## §1. Architecture Overview

The chain is an **append-only JSONL lineage tree** persisted at
`.moai/state/chain/events.jsonl`. Unlike the singular overwritten
`pending.json` (the clobber flaw at `pending.go:118`), append-only storage
structurally guarantees that a depth-3 save cannot destroy a depth-1 record.

### Topology

```
N0 (depth 0, primary checkout)
│
├── N1 (depth 1, worktree wt-auth)
│   ├── N2 (depth 2, worktree wt-m2)      ← user is here
│   └── N3 (depth 2, worktree wt-m3)
│
└── N4 (depth 1, worktree wt-tests)
    └── N5 (depth 2, worktree wt-fix-42)
```

Each node is one JSONL line (one event). The **tree is derived at read time**
from the flat event stream — no mutable tree file exists. A `parent_node_id`
field on each node establishes the edge. `origin_chain` (an ordered NodeID
list, root-to-leaf) is denormalized on each node for O(1) lineage queries
without tree traversal.

### Event Types

The JSONL stream carries three event types, one per line:

| Event type | Trigger | Fields populated |
|---|---|---|
| `node-enter` | spawn boundary (moai cc -w / EnterWorktree / Agent isolation:worktree) | node_id, parent_node_id, depth, origin_chain, worktree_path, spec_id, milestone, entered_at |
| `node-update` | child SessionStart (session_id backfill) or milestone completion | node_id, session_id (backfill) OR last_completed_milestone |
| `completion-edge` | chain-event.sh hook on SubagentStop | parent_node, child_node, completed_milestone, completed_at, next_resume_target |

---

## §2. Overlay-Join Design (REQ-CHAIN-014 / REQ-COORD-024)

The session registry `Entry` struct is **frozen** (REQ-COORD-024). The chain
MUST NOT add, remove, or reorder any field on `Entry`. Instead, lineage data
is overlaid at read time:

```
┌─────────────────────────────────────────────────┐
│  Chain Ledger (events.jsonl)                     │
│  node_id, session_id, worktree_path, depth, ...  │
└──────────────────────┬──────────────────────────┘
                       │ join by SessionID
                       ▼
┌─────────────────────────────────────────────────┐
│  Session Registry (active-sessions.json)         │
│  SessionID, SpecID, Phase, StartedAt,            │
│  LastHeartbeat, PID, Host, CWD                   │
└─────────────────────────────────────────────────┘
```

**Why overlay, not mutation**: adding `parent_session_id` / `depth` /
`origin_chain` directly to `Entry` would supersede REQ-COORD-024 and cascade
across every consumer of the registry (registry.go, session_start.go,
agent-common-protocol.md, the Pre-Spawn Sync Check). The overlay is
non-breaking: the registry operates identically with or without the chain
present.

**Join mechanics**: the chain read path calls `session.Registry.Query("")` to
get all `Entry` records, then matches each chain node's `session_id` to an
`Entry.SessionID`. Matched nodes inherit the registry's `LastHeartbeat` for
staleness classification (REQ-CHAIN-015). Unmatched nodes (session exited or
registry purged) are classified as `exited`.

**Performance**: the join is O(N+M) where N = chain nodes and M = registry
entries. Both are bounded: prune (REQ-CHAIN-011) caps the active node set,
and the registry caps at the number of concurrent sessions on one host.
Production N+M is typically < 20.

---

## §3. Spawn-Boundary Population Paths (REQ-CHAIN-005)

Three distinct entry points create a chain node. All three share the same
core logic: read `MOAI_CHAIN_NODE_ID` from the spawning context's
environment, derive `depth` and `origin_chain` from the parent node, and
append a `node-enter` event.

### Path A — `moai cc -w <name>`

The `moai cc` launcher creates a worktree and launches a Claude Code session
inside it. The population hook fires after the worktree is created but before
the session launches:

1. Read `MOAI_CHAIN_NODE_ID` from the current process env (the spawner's
   node ID; empty if the spawner is the primary checkout with no chain
   context → creates a depth-0 root node).
2. Read the parent node from the ledger to derive `parent_depth` and
   `parent_origin_chain`.
3. Generate `node_id` (ULID for monotonic ordering).
4. Append `node-enter` event with `depth = parent_depth + 1` and
   `origin_chain = parent_origin_chain + [node_id]`.
5. Set `MOAI_CHAIN_NODE_ID=<node_id>` on the child process environment before
   `exec`'ing into Claude Code.

### Path B — `EnterWorktree(<path>)`

The Claude Code runtime `EnterWorktree` tool re-enters an existing worktree
within the current session. The population path is identical to Path A except
the worktree already exists (no `git worktree add`). The node-enter event is
appended at the EnterWorktree call site.

### Path C — `Agent(isolation: "worktree")`

The Claude Code runtime creates an L1 worktree for an `Agent()` spawn. The
population path mirrors Path A: the orchestrator (or the spawning agent) sets
`MOAI_CHAIN_NODE_ID` on the child process environment, and the child's first
hook call records the node.

**Key invariant across all three paths**: the child process MUST receive
`MOAI_CHAIN_NODE_ID` in its environment. This is the thread that connects a
depth-N child to its depth-(N-1) parent across the process boundary.

---

## §4. SessionID Two-Phase Backfill Protocol (REQ-CHAIN-021)

### The Problem

At the spawn boundary (§3), the child's `SessionID` is **not yet known**.
Claude Code assigns the SessionID at runtime when the session starts — the
launcher cannot predict it. If the `node-enter` event waited for the
SessionID, it would have to be written from inside the child session, which
creates a chicken-and-egg: the child needs its chain node to function, but
the node needs the child's SessionID to be complete.

### The Protocol

The node is created in two phases:

```
Phase 1 — Skeleton (at spawn boundary):
  node-enter event appended with:
    node_id         = <ULID>
    parent_node_id  = <from env>
    depth           = parent_depth + 1
    origin_chain    = parent_chain + [node_id]
    worktree_path   = <resolved worktree path>
    session_id      = ""  ← EMPTY (backfill pending)
    entered_at      = <now>

Phase 2 — Backfill (at child SessionStart):
  node-update event appended with:
    node_id         = <same ULID>
    session_id      = <runtime-assigned SessionID>
```

The child's SessionStart handler reads `MOAI_CHAIN_NODE_ID` from env (set by
the spawner in §3), then appends the backfill `node-update` event binding
the node to the real SessionID.

### Why This Works for REQ-CHAIN-013

After `/clear`, the environment is lost — `MOAI_CHAIN_NODE_ID` is absent.
REQ-CHAIN-013 resolves the current node by matching `worktree_path` and
`session_id` against the ledger. Because Phase 2 backfilled the real
SessionID, the `(worktree_path, session_id)` pair uniquely identifies the
node. Without Phase 2, every node in the same worktree would have
`session_id=""`, making CWD-collision resolution (REQ-CHAIN-004) impossible.

### Edge Case — SessionStart Before Backfill

If SessionStart fires but the node-enter event was never written (e.g., the
spawner crashed between worktree creation and node-enter append), the
backfill handler finds no matching `MOAI_CHAIN_NODE_ID` in the ledger. It
creates a **synthetic root node** (depth 0, `parent_node_id=""`) so the
chain is usable even when the spawn boundary was missed. This is the same
degraded path as "no chain context active."

---

## §5. Store Internals

### Append-Only Write Path

```go
func (s *Store) Append(event ChainEvent) error {
    line, err := json.Marshal(event)
    if err != nil {
        return fmt.Errorf("marshal chain event: %w", err)
    }
    line = append(line, '\n')
    // O_APPEND: POSIX guarantees atomicity for writes ≤ PIPE_BUF.
    // See §6 Risk — concurrent-append interleaving on lines > PIPE_BUF.
    f, err := os.OpenFile(s.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
    if err != nil {
        return fmt.Errorf("open chain ledger: %w", err)
    }
    defer f.Close()
    _, err = f.Write(line)
    return err
}
```

The file is opened with `O_APPEND` — the kernel serializes concurrent
appends to the same file. No read-modify-write cycle exists; the store never
loads the full file to mutate it. This is the structural fix for the
`pending.go:118` clobber pattern (`atomicWriteFile` loads → marshals →
replaces).

### Corrupt-Line Tolerance (REQ-CHAIN-003)

```go
func (s *Store) ReadAll() ([]ChainEvent, error) {
    scanner := bufio.NewScanner(f)
    var events []ChainEvent
    lineNum := 0
    for scanner.Scan() {
        lineNum++
        var ev ChainEvent
        if err := json.Unmarshal(scanner.Bytes(), &ev); err != nil {
            slog.Warn("chain: skipping corrupt line",
                "file", s.path, "line", lineNum, "error", err)
            continue  // ← skip, do NOT abort
        }
        events = append(events, ev)
    }
    return events, nil
}
```

A corrupt line (partial write, encoding error, interleaved garbage from a
concurrent append) is skipped with a warning. The chain remains usable —
losing one event does not invalidate the rest. This mirrors the navigator
graph build's fail-open line-skip pattern.

### CWD-Collision Resolution (REQ-CHAIN-004)

When two nodes share the same `worktree_path` (e.g., two sessions opened in
the same worktree), the resolution priority is:

1. **Primary key**: `(worktree_path, session_id)` pair — the exact session
   that entered this worktree.
2. **Fallback**: if the pair is ambiguous (e.g., `session_id` not yet
   backfilled), query the session registry for PID liveness — prefer the
   node whose session PID is still alive (`kill -0`).
3. **Indeterminate**: if still ambiguous, emit a diagnostic warning and
   return the most recently entered node (fail-open). The user can manually
   re-anchor with `moai chain attach <node-id>` (deferred to a follow-up;
   v1 uses diagnostic-only).

### Compaction / Prune (REQ-CHAIN-011)

`moai chain prune` folds exited nodes older than the threshold (30 days OR
10 MB) into `.moai/state/chain/archived/node-summary.json`. The active read
path (the `ReadAll` above) reads only `events.jsonl`; archived nodes are
excluded. The audit trail retains the archived snapshot.

**v1 scope note**: prune is MANUAL only (user-invoked). Automatic
compaction (scheduled folding without user invocation) is deferred — see
spec.md §I Out of Scope. The staleness display (REQ-CHAIN-015) serves as the
v1 "auto" surface: stale nodes (LastHeartbeat > 15min) are auto-excluded
from the active view without compaction.

---

## §6. Completion-Hook Write Path (REQ-CHAIN-012)

### Hook Architecture

```
Claude Code SubagentStop event
        │
        ▼
.claude/hooks/moai/chain-event.sh  (shell wrapper)
        │  INPUT=$(cat); moai hook chain-event <<< "$INPUT"
        ▼
internal/hook/chain_event.go  (Go handler)
        │  reads stdin JSON, extracts session_id + completed context
        │  resolves child node from ledger
        │  appends completion-edge event
        ▼
.moai/state/chain/events.jsonl  (append-only)
```

### No Leaf Cooperation

The hook fires **mechanically** on `SubagentStop` — the leaf agent does not
need to call any function, write any file, or cooperate in any way. This is
the guarantee that the ledger stays current even when:

- The spawning session (lead) has crashed or been `/clear`'d.
- The leaf agent exited abnormally (timeout, error).
- The leaf agent is a third-party or custom agent with no chain awareness.

The hook reads the `session_id` from the SubagentStop payload, resolves the
chain node from the ledger, and appends the completion edge. If no matching
node exists (the session had no chain context), the hook is a no-op (exits 0).

### Wiring

The hook is registered in `.claude/settings.json` under the `SubagentStop`
event, mirroring the existing hook wrapper pattern
(`handle-session-start.sh` → `moai hook session-start`).

---

## §7. Package Structure

```
internal/chain/
├── node.go           # WorktreeNode struct + ChainEvent types
├── store.go          # Append-only JSONL store (Append, ReadAll, Resolve)
├── store_test.go     # t.TempDir isolation, corrupt-line, CWD-collision tests
└── chain_test.go     # TestNoAskUserQuestion static guard

internal/cli/
└── chain.go          # Cobra chainCmd + subcommands (status/lineage/back/list/prune)

internal/hook/
└── chain_event.go    # SubagentStop → completion-edge handler

internal/config/
└── envkeys.go        # + EnvChainNodeID constant (modified, existing file)

internal/template/templates/
├── .claude/hooks/moai/chain-event.sh   # template-mirrored hook wrapper
├── .claude/settings.json               # + SubagentStop entry
└── .moai/state/chain/                  # state-dir scaffold (created on moai init)
```

No new top-level package boundary. `internal/chain` is self-contained — it
imports `internal/session` (read-only, for the overlay join) and
`internal/config` (for `EnvChainNodeID`). No other internal package imports
`internal/chain` except `internal/cli` (command registration) and
`internal/hook` (the completion hook handler).

---

## §8. Data Flow — Depth-3 Amnesia Prevention

```
SPAWNER (depth 1, wt-auth, session S1)
  │
  │  1. moai cc -w wt-m2
  │  2. Read MOAI_CHAIN_NODE_ID=N1 from env
  │  3. Append node-enter: N2 (parent=N1, depth=2, session_id="")
  │  4. Set MOAI_CHAIN_NODE_ID=N2 on child env
  │  5. exec claude
  ▼
CHILD (depth 2, wt-m2, session S2)
  │
  │  SessionStart fires:
  │  6. Read MOAI_CHAIN_NODE_ID=N2 from env (or resolve from ledger)
  │  7. Append node-update: N2.session_id = S2  ← BACKFILL
  │  8. Emit lineage banner: "depth 2 of N0→N1→N2"
  │
  │  ... work happens ...
  │
  │  SubagentStop fires (child agent completes):
  │  9. chain-event.sh appends completion-edge:
  │     {parent=N1, child=N2, completed=M2, next="Continue M3"}
  ▼
/clear  ← env lost, MOAI_CHAIN_NODE_ID gone
  │
  │  SessionStart fires again:
  │  10. MOAI_CHAIN_NODE_ID absent → resolve from ledger
  │      by (worktree_path=wt-m2, session_id=S2) → N2
  │  11. Re-inject MOAI_CHAIN_NODE_ID=N2
  │  12. Emit banner: "depth 2 of N0→N1→N2; resume: Continue M3"
  │
  │  ✅ Origin recovered. Completion visible. Resume target known.
```

The append-only ledger at steps 3, 7, and 9 is the durable substrate that
survives `/clear`. No step depends on in-memory session state.

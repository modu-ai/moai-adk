# SPEC-CHAIN-CORE-001 — Acceptance Criteria

> **Format**: Given-When-Then (binary-testable, one deterministic command per AC)
> **Mapping**: each AC traces to one or more REQ-CHAIN-* requirements in spec.md §D

---

## §D. AC Matrix

### AC-CHAIN-001 — WorktreeNode struct has all 13 named fields

**Given** the `internal/chain` package is compiled
**When** `go doc github.com/modu-ai/moai-adk/internal/chain.WorktreeNode` is run
**Then** the output lists all 13 fields: `NodeID`, `ParentNodeID`, `Depth`,
`OriginChain`, `WorktreePath`, `SessionID`, `SpecID`, `Milestone`,
`EnteredAt`, `ExitedAt`, `LastCompletedMilestone`, `ResumeTarget`,
`ResumeCommand`

**Command**: `go doc github.com/modu-ai/moai-adk/internal/chain.WorktreeNode`

**REQs**: REQ-CHAIN-001

---

### AC-CHAIN-002 — Store appends without overwriting

**Given** a chain store bound to a `t.TempDir()` path with 3 existing event
lines
**When** a 4th event is appended
**Then** the file contains exactly 4 lines (the original 3 are preserved) and
no line was overwritten

**Command**: `go test ./internal/chain/ -run TestAppendDoesNotOverwrite -v`

**REQs**: REQ-CHAIN-002

---

### AC-CHAIN-003 — Corrupt line is skipped and logged

**Given** a chain store file containing 5 lines where line 3 is malformed JSON
**When** the store reads all events
**Then** 4 valid events are returned (the corrupt line is skipped), and a
warning log message identifies line 3

**Command**: `go test ./internal/chain/ -run TestCorruptLineTolerance -v`

**REQs**: REQ-CHAIN-003

---

### AC-CHAIN-004 — CWD-collision resolved by (worktree_path, session_id) pair

**Given** two nodes with the same `worktree_path` but different `session_id`
**When** the store resolves the current node for `(worktree_path, session_id_A)`
**Then** the node with `session_id_A` is returned (not the other)

**Command**: `go test ./internal/chain/ -run TestCWDCollisionResolution -v`

**REQs**: REQ-CHAIN-004

---

### AC-CHAIN-005 — Spawn-boundary creates child node with correct depth

**Given** a parent node at depth 1 with `MOAI_CHAIN_NODE_ID=<parent_node_id>`
set in the environment
**When** a new worktree session is entered via the spawn path
**Then** a new node is created with `parent_node_id=<parent_node_id>`,
`depth=2`, and `origin_chain` containing both the parent and new node IDs

**Command**: `go test ./internal/chain/ -run TestSpawnBoundaryNodeCreation -v`

**REQs**: REQ-CHAIN-005

---

### AC-CHAIN-006 — EnvChainNodeID constant exists in envkeys.go

**Given** the `internal/config` package is compiled
**When** `go doc github.com/modu-ai/moai-adk/internal/config.EnvChainNodeID` is run
**Then** the output shows `const EnvChainNodeID = "MOAI_CHAIN_NODE_ID"`

**Command**: `go doc github.com/modu-ai/moai-adk/internal/config.EnvChainNodeID`

**REQs**: REQ-CHAIN-006

---

### AC-CHAIN-007 — `moai chain status` prints current node summary

**Given** a chain ledger with a node whose `worktree_path` matches the current
CWD
**When** `moai chain status` is invoked
**Then** the output includes `depth`, `parent_node_id`, `spec_id`,
`milestone`, `last_completed_milestone`, and `resume_target`

**Command**: `go test ./internal/cli/ -run TestChainStatus -v`

**REQs**: REQ-CHAIN-007

---

### AC-CHAIN-008 — `moai chain lineage` prints root-to-leaf chain

**Given** a depth-3 chain (N0 → N1 → N2)
**When** `moai chain lineage` is invoked from the N2 worktree
**Then** the output lists all three nodes in root-to-leaf order with their
`worktree_path`, `spec_id`, `milestone`, and `entered_at`

**Command**: `go test ./internal/cli/ -run TestChainLineage -v`

**REQs**: REQ-CHAIN-008

---

### AC-CHAIN-009 — `moai chain back` prints parent resume target

**Given** a depth-2 node whose parent has `resume_target="Continue M2"` and
`resume_command="/moai run SPEC-AUTH-001"`
**When** `moai chain back` is invoked from the depth-2 worktree
**Then** the output includes `Continue M2` and `/moai run SPEC-AUTH-001`

**Command**: `go test ./internal/cli/ -run TestChainBack -v`

**REQs**: REQ-CHAIN-009

---

### AC-CHAIN-010 — `moai chain list` enumerates all nodes

**Given** a ledger with 4 nodes (1 root, 2 depth-1, 1 depth-2)
**When** `moai chain list` is invoked
**Then** the output lists all 4 nodes with `depth`, `session_id`,
`worktree_path`, `spec_id`, and staleness status

**Command**: `go test ./internal/cli/ -run TestChainList -v`

**REQs**: REQ-CHAIN-010, REQ-CHAIN-015

---

### AC-CHAIN-011 — `moai chain prune` folds old exited nodes

**Given** a ledger with 10 exited nodes older than 30 days and 5 active nodes
**When** `moai chain prune` is invoked
**Then** the 10 old nodes are moved to `.moai/state/chain/archived/`, the
active read path returns only the 5 active nodes, and the archived nodes are
retained in the audit trail

**Command**: `go test ./internal/chain/ -run TestChainPrune -v`

**REQs**: REQ-CHAIN-011

---

### AC-CHAIN-012 — Completion hook appends edge on SubagentStop

**Given** a chain ledger and a SubagentStop hook event with a known
`session_id`
**When** the `chain-event.sh` hook handler processes the event
**Then** a completion edge `{parent_node, child_node, completed_milestone,
completed_at, next_resume_target}` is appended to the ledger

**Command**: `go test ./internal/hook/ -run TestChainEventHook -v`

**REQs**: REQ-CHAIN-012

---

### AC-CHAIN-013 — SessionStart re-injects node ID and emits banner

**Given** a depth-2 worktree where `MOAI_CHAIN_NODE_ID` is NOT set in the
environment (simulating post-`/clear` state) and the ledger has a matching
node
**When** the SessionStart handler fires
**Then** the handler resolves the node via `worktree_path` + `session_id`,
re-injects `MOAI_CHAIN_NODE_ID`, and emits a system-reminder containing
`depth`, a parent summary, and `resume_target`

**Command**: `go test ./internal/hook/ -run TestSessionStartLineageBanner -v`

**REQs**: REQ-CHAIN-013

---

### AC-CHAIN-014 — SessionStart banner is time-boxed and fail-open

**Given** a chain ledger read that exceeds the time-box threshold
**When** the SessionStart handler fires
**Then** the handler returns without blocking (fail-open) and emits no lineage
banner (or a degraded no-op banner)

**Command**: `go test ./internal/hook/ -run TestSessionStartBannerTimeout -v`

**REQs**: REQ-CHAIN-013

---

### AC-CHAIN-015 — Entry struct is never constructed by chain package

**Given** the chain package source code
**When** `grep -rn 'session\.Entry{' internal/chain/` is run (checking for
struct construction of the registry Entry)
**Then** zero matches exist — the chain package constructs `WorktreeNode`
records, never `Entry` records. NOTE: this is a supplementary construction
check; the authoritative Entry-freeze gate is AC-CHAIN-024 (struct-field
baseline-diff, MUST-PASS).

**Command**: `grep -rn 'session\.Entry{' internal/chain/ | grep -v _test.go || echo "PASS: no Entry construction in chain package"`

**REQs**: REQ-CHAIN-014

---

### AC-CHAIN-016 — Stale nodes excluded from active view

**Given** a node whose session `LastHeartbeat` is 20 minutes old (threshold
default: 15 minutes)
**When** `moai chain list` is invoked
**Then** the node is marked `stale` and excluded from the active-chain
summary (but appears with a stale marker in the full list)

**Command**: `go test ./internal/cli/ -run TestChainStaleNodeDisplay -v`

**REQs**: REQ-CHAIN-015

---

### AC-CHAIN-017 — Single-host limitation surfaced

**Given** a CWD that resolves to a remote path (not on the local filesystem)
**When** `moai chain status` is invoked
**Then** the output includes a limitation notice about single-host v1

**Command**: `go test ./internal/cli/ -run TestChainRemoteCWD -v`

**REQs**: REQ-CHAIN-016

---

### AC-CHAIN-018 — Flag-agnostic: no -k/-f/manager-lead dependency

**Given** the chain package and CLI source code
**When** `grep -rn 'MOAI_KANBAN\|MOAI_FACTORY\|manager.lead\|managerLead' internal/chain/ internal/cli/chain.go internal/hook/chain_event.go .claude/hooks/moai/chain-event.sh` is run
**Then** zero matches exist

**Command**: `grep -rn 'MOAI_KANBAN\|MOAI_FACTORY\|manager.lead\|managerLead' internal/chain/ internal/cli/chain.go internal/hook/chain_event.go .claude/hooks/moai/chain-event.sh | grep -v _test.go || echo "PASS: no kanban/factory/lead dependency"`

**REQs**: REQ-CHAIN-017

---

### AC-CHAIN-019 — No AskUserQuestion in chain CLI or hook

**Given** the chain package, chain CLI, and chain-event hook source code
**When** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/chain/ internal/cli/chain.go internal/hook/chain_event.go .claude/hooks/moai/chain-event.sh` is run
**Then** zero matches exist (the orchestrator-only boundary is preserved)

**Command**: `go test ./internal/cli/ -run TestNoAskUserQuestionInChain -v`

**REQs**: REQ-CHAIN-018

---

### AC-CHAIN-020 — Template-mirror artifacts exist and pass neutrality

**Given** the template source tree
**When** `ls internal/template/templates/.claude/hooks/moai/chain-event.sh` and
the neutrality CI guard run
**Then** the hook template exists, the CLI command is registered in the
template, the state-dir scaffold is created on `moai init`, and the neutrality
guard reports zero SPEC-ID / commit-SHA / macOS-path violations

**Command**: `go test ./internal/template/... -run TestTemplateNeutralityAudit -v`

**REQs**: REQ-CHAIN-019

---

### AC-CHAIN-021 — Test isolation: all chain tests use t.TempDir

**Given** the chain test source files
**When** `grep -rn 't\.TempDir' internal/chain/*_test.go` is run
**Then** every test function that creates a store path references
`t.TempDir()` (no test writes to the project's real `.moai/state/chain/`)

**Command**: `go test ./internal/chain/ -run TestStoreIsolation -v`

**REQs**: REQ-CHAIN-020

---

### AC-CHAIN-022 — Full test suite passes

**Given** all chain, CLI, and hook code is implemented
**When** `go test ./internal/chain/... ./internal/cli/... ./internal/hook/... -count=1` is run
**Then** all tests pass with exit code 0

**Command**: `go test ./internal/chain/... ./internal/cli/... ./internal/hook/... -count=1`

**REQs**: REQ-CHAIN-001 through REQ-CHAIN-021

---

### AC-CHAIN-023 — SessionID two-phase backfill at child SessionStart

**Given** a skeleton node created at the spawn boundary with `session_id=""`
**When** the child's `SessionStart` fires and assigns the runtime SessionID
**Then** a `node-update` event is appended to the ledger binding the node's
`session_id` to the child's real SessionID, enabling REQ-CHAIN-013's
`(worktree_path, session_id)` re-resolution after `/clear`

**Command**: `go test ./internal/chain/ -run TestSessionIDBackfill -v`

**REQs**: REQ-CHAIN-021

---

### AC-CHAIN-024 — Entry struct fields unchanged (MUST-PASS baseline-diff)

**Given** the chain package and all modifications are complete
**When** the `Entry` struct definition in `internal/session/registry.go` is
inspected for chain-specific field additions
**Then** the struct contains exactly the 8 frozen fields (SessionID, SpecID,
Phase, StartedAt, LastHeartbeat, PID, Host, CWD) — zero chain-specific
fields (ChainNodeID, OriginChain, Depth, ParentNodeID, ResumeTarget) are
present. This is the authoritative REQ-COORD-024 freeze gate.

**Command**: `test $(sed -n '/type Entry struct/,/^}/p' internal/session/registry.go | grep -cE 'ChainNodeID|OriginChain|Depth|ParentNodeID|ResumeTarget') -eq 0 && echo "PASS" || echo "FAIL"`

**REQs**: REQ-CHAIN-014

---

## §D.1 Severity Classification

| Severity | AC IDs | Gate |
|---|---|---|
| **MUST-PASS** | 001, 002, 003, 004, 005, 006, 014, 015, 018, 019, 021, 022, 023, 024 | Blocks run-phase completion |
| **SHOULD-PASS** | 007, 008, 009, 010, 012, 013 | Blocks sync-phase |
| **NICE-TO-HAVE** | 011, 016, 017, 020 | Deferred acceptable with documented rationale |

---

## §D.2 Edge Cases

| Edge Case | Expected Behavior |
|---|---|
| Empty ledger (no events) | `moai chain status` prints "no chain context"; CLI exits 0 |
| Ledger file absent entirely | Store returns empty slice (not error); CLI prints "no chain context" |
| Node with empty `parent_node_id` (root) | `moai chain back` prints "at root — no parent" |
| Depth-0 root node (no worktree) | Created as synthetic root for standalone operation |
| Concurrent append from two processes | O_APPEND atomic on POSIX for writes ≤ PIPE_BUF (512–4096 bytes); lines > PIPE_BUF may interleave — corrupt-line tolerance (REQ-CHAIN-003) skips garbage, and node lines are kept compact where feasible |
| Very large ledger (>10MB) | `moai chain prune` folds old nodes; read path bounds active set |
| Worktree path with spaces or unicode | JSON-escaped in JSONL; resolved correctly on read |

---

## §D.3 Quality Gate Criteria

- **Tested**: `internal/chain` package coverage ≥ 85%; corrupt-line and
  CWD-collision paths explicitly covered
- **Readable**: godoc on every exported function; no magic strings (env names
  are constants)
- **Unified**: `gofmt` + `golangci-lint` clean; file naming `snake_case.go`
- **Secured**: JSONL write uses `0o600` (ledger may carry session context);
  no path traversal in node_id or worktree_path
- **Trackable**: conventional commits per milestone (`feat(SPEC-CHAIN-CORE-001): M0 ...`)

---

## §D.4 Definition of Done

- [ ] All MUST-PASS ACs green (deterministic command output shown)
- [ ] `go test ./internal/chain/... ./internal/cli/... ./internal/hook/... -count=1` exit 0
- [ ] `go vet ./...` clean
- [ ] `golangci-lint run` clean
- [ ] `make build` succeeds; template embed compiles
- [ ] Template-neutrality CI guard passes
- [ ] No `-k`/`-f`/`manager-lead`/`AskUserQuestion` grep matches in chain scope
- [ ] Entry struct fields unchanged — AC-CHAIN-024 MUST-PASS baseline-diff (REQ-COORD-024 preserved)

---

## §D.5 Forward-Looking Checks

| Check | Purpose |
|---|---|
| `moai chain lineage` works from a real depth-2 worktree | End-to-end validation outside unit tests |
| SessionStart banner appears on `/clear` + re-entry | REQ-CHAIN-013 real-world verification |
| `chain-event.sh` fires on a real SubagentStop | REQ-CHAIN-012 hook wiring validation |
| Ledger survives a `moai init` on a fresh project | Template-mirror delivers the scaffold |

---

## §D.6 Indirect Verification

| Concern | How Verified Indirectly |
|---|---|
| Append-only survives `/clear` | AC-CHAIN-013 (SessionStart reads from disk post-`/clear`) |
| No regression in existing SessionStart | Full `go test ./internal/hook/...` suite (existing tests unchanged) |
| No regression in registry | Full `go test ./internal/session/...` suite (Entry untouched) |
| Overlay join correctness | AC-CHAIN-016 (staleness reads registry `LastHeartbeat` via join) |

---

## §D.7 Closure Gates

- [ ] plan-auditor independent audit completed (GEARS compliance, scope check)
- [ ] Implementation Kickoff Approval human gate passed
- [ ] All MUST-PASS + SHOULD-PASS ACs green
- [ ] `progress.md` §E.2/§E.3 populated by manager-develop
- [ ] No `[NEEDS CLARIFICATION]` markers remain in plan.md

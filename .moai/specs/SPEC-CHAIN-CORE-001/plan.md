# SPEC-CHAIN-CORE-001 — Implementation Plan

> **Tier**: L (multi-subsystem: chain package + CLI + hook + config + template)
> **Milestones**: 7 (M0–M6), ordered by decision-reversibility
> **No new external dependencies**

---

## §A. Context

The maintainer (GOOS) loses context at worktree depth 2–3: origin, completion,
and resume-target all die with the session's in-memory context at `/clear`.
The verified root cause is a missing data structure — no persistence primitive
carries parent/depth/origin-chain metadata. This SPEC delivers an append-only
JSONL lineage tree that answers all three questions from structured disk
records, independently of any later delegation or kanban work.

**Architecture**: a new `internal/chain` package owns the `WorktreeNode`
struct and the append-only store. The store joins the session registry by
`SessionID` at read time (Overlay — Entry is never mutated). Population occurs
at three spawn boundaries. A CLI (`moai chain`) and a hook
(`chain-event.sh`) are the user-facing surfaces.

---

## §B. Known Issues

| Issue | Impact | Resolution |
|---|---|---|
| `pending.go:118` clobbers singular handoff | Depth-3 save destroys depth-1 record | Structurally prevented by append-only design (this SPEC); per-node handoff namespace is Phase 4 |
| `MOAI_CHAIN_NODE_ID` lost after `/clear` | SessionStart cannot read node ID from env | REQ-CHAIN-013 re-injects from ledger via CWD + session_id |
| No depth/origin fields on `Entry` | Registry is flat — no lineage metadata | Overlay join by SessionID (REQ-CHAIN-014); Entry untouched |
| Stale nodes from SIGKILL/OOM | Node stays "active" forever | REQ-CHAIN-015 reuses registry `LastHeartbeat` (default 15min stale) |

---

## §C. Pre-Flight

- [ ] `go test ./internal/session/...` passes (registry baseline unaffected)
- [ ] `go test ./internal/hook/...` passes (SessionStart baseline)
- [ ] `go vet ./...` clean
- [ ] `make build` succeeds (template embed compiles)
- [ ] No uncommitted changes to `internal/session/registry.go` (frozen Entry)

---

## §D. Constraints

1. **Overlay-only**: never mutate `Entry` (REQ-COORD-024 freeze). Join by
   `SessionID` at read.
2. **Flag-agnostic**: no `-k`/`-f`/`manager-lead` dependency.
3. **Append-only**: the store appends JSONL lines; it never overwrites.
4. **Hook subagent boundary**: no `AskUserQuestion` in CLI or hook code.
5. **Template-First**: all distributable artifacts mirrored to
   `internal/template/templates/`.
6. **Env SSOT**: `MOAI_CHAIN_NODE_ID` is a constant in `envkeys.go`.
7. **Test isolation**: `t.TempDir()` for all store paths.

---

## §E. Self-Verification (Plan-Phase Checklist)

- [ ] SPEC ID regex check run as Bash — PASS output cited
- [ ] Frontmatter validated against 12-field schema SSOT
- [ ] ID uniqueness confirmed against existing SPECs (0 CHAIN-* matches)
- [ ] Requirements written in GEARS notation (21 REQs)
- [ ] Out of Scope section satisfies `OutOfScopeRule` (9 H3 sub-headings + bullets)
- [ ] Artifact set matches Tier L (spec + plan + acceptance + design + research)
- [ ] spec.md carries no implementation detail (function names only as
      identifiers in REQ subjects, not as HOW)

---

## §F. Milestones (Ordered by Decision-Reversibility)

> Decisions most likely to change are listed first; mechanical/refactoring
> steps are deferred to the bottom.

### M0 — WorktreeNode Struct + Append-Only JSONL Store

**Highest change likelihood** — the data model is the load-bearing decision.
If the struct shape changes, every downstream milestone adjusts.

**Deliverables**:
- `internal/chain/node.go` — `WorktreeNode` struct with all 13 named fields
  (REQ-CHAIN-001)
- `internal/chain/store.go` — append-only JSONL writer/reader at
  `.moai/state/chain/events.jsonl` (REQ-CHAIN-002)
- Corrupt-line tolerance: skip + log on malformed JSONL line (REQ-CHAIN-003)
- CWD-collision resolution: `(worktree_path, session_id)` pair key with
  PID-liveness fallback (REQ-CHAIN-004)
- `internal/chain/store_test.go` — `t.TempDir` isolation, corrupt-line
  tolerance, CWD-collision resolution tests (REQ-CHAIN-020)

**Exit criteria**: `go test ./internal/chain/...` passes; append-only invariant
verified (write N events, read back N events, no overwrite).

### M1 — Spawn-Boundary Population + Env Constant

**Interface change** — the population path is where the chain meets the
existing CLI/worktree subsystem.

**Deliverables**:
- `internal/config/envkeys.go` — `EnvChainNodeID = "MOAI_CHAIN_NODE_ID"`
  constant (REQ-CHAIN-006)
- Population at `moai cc -w` (spawn path in `internal/cli/worktree/new.go`)
  (REQ-CHAIN-005)
- Population at `EnterWorktree` path (REQ-CHAIN-005)
- Population at `Agent(isolation: "worktree")` env-passing
  (`MOAI_CHAIN_NODE_ID` set on the child process) (REQ-CHAIN-005)
- `depth` = parent depth + 1; `origin_chain` = parent chain + new node_id
- **SessionID two-phase backfill** (REQ-CHAIN-021): skeleton node created at
  spawn with `session_id=""`; backfilled via `node-update` event at child
  SessionStart when the runtime assigns the real SessionID. See design.md §4.

**Exit criteria**: entering a worktree via `moai cc -w` creates a ledger
entry; `MOAI_CHAIN_NODE_ID` is set in the child environment; nested entry
produces depth 2 then 3.

### M2 — Chain CLI Subcommands

**User-facing UX** — the query surface the maintainer uses daily.

**Deliverables**:
- `internal/cli/chain.go` — Cobra `chainCmd` with subcommands:
  - `moai chain status` (REQ-CHAIN-007)
  - `moai chain lineage` (REQ-CHAIN-008)
  - `moai chain back` (REQ-CHAIN-009)
  - `moai chain list` (REQ-CHAIN-010)
  - `moai chain prune` (REQ-CHAIN-011)
- Staleness display: active / stale / exited via overlay join on registry
  `LastHeartbeat` (REQ-CHAIN-015)
- Single-host limitation surfaced when remote CWD detected (REQ-CHAIN-016)
- `internal/cli/chain_test.go` — `TestNoAskUserQuestionInChain` static grep
  guard (REQ-CHAIN-018)

**Exit criteria**: `moai chain status` prints current node summary;
`moai chain lineage` prints root-to-leaf chain; `moai chain back` prints
parent resume target.

### M3 — SessionStart Lineage Banner

**User-facing UX** — the "instantly know origin" surface.

**Deliverables**:
- `internal/hook/session_start.go` — lineage banner injected into
  `sessionStartHandler.Handle` (REQ-CHAIN-013)
- Re-inject `MOAI_CHAIN_NODE_ID` from ledger when env is absent (post-`/clear`)
- Emit depth + parent chain summary + resume_target system-reminder
- Time-boxed via `context.WithTimeout`; fail-open to no-op on timeout
- Overlay join for staleness check (reads registry `LastHeartbeat`)

**Exit criteria**: SessionStart in a depth-2 worktree emits a lineage
system-reminder; `/clear` + re-entry re-injects node ID from ledger.

### M4 — Mechanical Completion Hook

**Integration** — the ledger-stays-current guarantee.

**Deliverables**:
- `.claude/hooks/moai/chain-event.sh` — shell wrapper
  (`INPUT=$(cat); moai hook chain-event <<< "$INPUT"`) (REQ-CHAIN-012)
- `internal/hook/chain_event.go` — hook handler: reads SubagentStop /
  phase-transition JSON from stdin, appends completion edge to ledger
- Wired in `.claude/settings.json` under SubagentStop event
- No leaf-agent cooperation required (hook fires mechanically)

**Exit criteria**: SubagentStop produces a completion edge in the ledger;
ledger stays current even if the spawning session was killed.

### M5 — Compaction / Prune

**Mechanical** — unbounded growth mitigation.

**Deliverables**:
- `moai chain prune` full implementation (REQ-CHAIN-011)
- Fold exited nodes older than threshold (30 days OR 10 MB) into
  `.moai/state/chain/archived/node-summary.json`
- Active read path excludes archived nodes
- Audit trail retains archived nodes

**Exit criteria**: `moai chain prune` on a fixture with 50 exited nodes
folds old nodes and reduces active-path line count.

### M6 — Template-Mirror + Test Finalization

**Mechanical** — distribution and CI guard compliance.

**Deliverables**:
- Template-mirror per REQ-CHAIN-019:
  - `internal/template/templates/.claude/hooks/moai/chain-event.sh`
  - `internal/template/templates/.claude/settings.json` (SubagentStop entry)
  - CLI command registration in template
  - State-dir scaffold (`.moai/state/chain/` creation in `moai init`)
- `make build` recompiles binary with embedded templates
- Template neutrality: no SPEC IDs, commit SHAs, or macOS paths in template
  source (pass `template-neutrality-check.yaml` CI guard)
- Full test suite: `go test ./internal/chain/... ./internal/cli/...`
- `go vet ./...` and `golangci-lint run` clean

**Exit criteria**: `./moai init /tmp/test-chain-project` deploys the chain
hook + CLI + state-dir scaffold; template-neutrality CI guard passes.

---

## §G. Anti-Patterns

- **Overwriting the JSONL file**: the store MUST append, never overwrite. The
  clobber flaw at `pending.go:118` is the exact anti-pattern this SPEC
  prevents structurally.
- **Mutating Entry**: adding lineage fields to `Entry` would break
  REQ-COORD-024 and cascade across every consumer. The Overlay join is the
  non-breaking alternative.
- **Invoking AskUserQuestion from hook/CLI**: the orchestrator-only boundary
  is HARD. The static grep guard test enforces zero matches.
- **Adding a socket before the record**: the diagnostic is unambiguous —
  transport before payload is a pipe with no water. Build the record first;
  push transport is Phase 7 (optional, measure-first).
- **Coupling to kanban/-k/lead**: this SPEC solves continuity, not delegation.
  Any code that reads `-k`, `MOAI_KANBAN`, or imports manager-lead is a scope
  violation.
- **Hardcoding `MOAI_CHAIN_NODE_ID` string**: the env name MUST be the
  `EnvChainNodeID` constant from `envkeys.go`, per CLAUDE.local.md §14.

---

## §H. Cross-References

- Diagnostic workflow `wf_0aed4941-ad6` (root-cause + winner design)
- `internal/session/registry.go` — Entry struct (frozen, REQ-COORD-024)
- `internal/hook/handoff/pending.go:118` — the clobber flaw
- `internal/hook/session_start.go` — lineage banner hook site
- `internal/worktree/state_guard.go` — existing Snapshot (distinct, not duplicated)
- `internal/config/envkeys.go` — env constant home
- CLAUDE.local.md §2, §6, §14, §25

---

## §Roadmap — Origin-Trail Chain Epic (P1–P7)

> This SPEC is **P1**. P3–P7 are separate future SPECs. They are listed here
> for roadmap context only — they are NOT implemented in this SPEC.

| Phase | SPEC (future) | Summary | Dependency on P1 |
|---|---|---|---|
| **P1 (THIS)** | SPEC-CHAIN-CORE-001 | Chain record + query CLI + completion hook + SessionStart banner | — |
| P3 | SPEC-TODO-INBOX-* | `/moai:todo` continuous-capture inbox (chain-stamped `backlog.jsonl`) | Consumes `MOAI_CHAIN_NODE_ID` stamp |
| P4 | SPEC-CHAIN-HANDOFF-* | Multi-record handoff fix (`pending.go:118` clobber → per-node `<node_id>.json`) | Consumes chain node identity |
| P5 | SPEC-KANBAN-MODE-* | `-k` flag, factory→kanban rename, derived board from `backlog.jsonl` | Consumes chain partitioning |
| P6 | SPEC-KANBAN-LEAD-* | `manager-lead` rewrite (chain-aware lead) + depth-2 seal → depth-ceiling CI rewrite | Consumes chain + backlog |
| P7 | SPEC-CHAIN-PUSH-* (optional) | Live push channel (fswatch/socket), wire dead `factory.go:102-105` stub honestly | Measure-first; only if file-poll insufficient |

**Key principle**: P1 is independently shippable. Each subsequent phase consumes
the chain substrate delivered here but does not modify it. The chain solves
depth-3 amnesia ALONE — delegation, kanban, and push transport are orthogonal
concerns layered on top.

# SPEC-CHAIN-CORE-001 — Research Document

> **Artifact role**: Tier L research document. Materializes the diagnostic
> evidence that established the root cause and selected the winner design.
> Provenance: 11-agent diagnostic workflow `wf_0aed4941-ad6` (2026-08-13).
> Sources: `/tmp/wf_diag.txt` (full diagnosis), `/tmp/wf_synth.txt` (winner
> design + grafted ideas + implementation phases).

---

## §1. Diagnostic Provenance

The root cause and design selection for this SPEC were established by an
11-agent diagnostic workflow (`wf_0aed4941-ad6`, 2026-08-13). The workflow
produced three artifacts:

- **`/tmp/wf_diag.txt`** — the root-cause analysis: why the broken link is
  neither a missing socket nor a missing agent, but a missing DATA STRUCTURE.
  Includes a survey of why each past attempt fell short.
- **`/tmp/wf_synth.txt`** — the synthesis: the winner design (Origin-Trail
  Chain), the grafted ideas (append-only JSONL from LAK, mechanical
  completion hook from LMK), the 7-phase implementation roadmap, and the
  resolved open questions (the 8 locked policies).
- **`/tmp/wf_designs.txt`** — the three competing design approaches (287
  lines, not reproduced here; the winner is documented in §5 below).

This research document materializes the load-bearing findings so they are
discoverable from the SPEC directory itself (the `/tmp/` artifacts are
ephemeral).

---

## §2. Root Cause — The Missing Data Structure

### The Grep-Zero Evidence

A grep across four subsystems returned **ZERO matches** for lineage-related
field names:

```
grep -rn 'parent_session\|origin_session\|spawned_from\|parent_worktree\|origin_chain\|nesting_depth\|resume_target' \
    internal/session/ internal/factory/ internal/worktree/ internal/hook/handoff/
# Result: 0 matches
```

There is no `parent_session_id`, no `depth`, no `origin_chain`, no
`resume_target` on ANY persistence primitive in the codebase.

### The Maintainer's Pain (3 Questions)

At worktree depth 2–3, after `/clear`, the maintainer (GOOS) cannot answer:

1. **Origin** — "where did I come from?" (which worktree/session spawned this one?)
2. **What just completed** — "what did I just finish?" (which milestone/phase?)
3. **Where to resume** — "where should I return to?" (what is the next action at the parent level?)

Today, the ONLY thing that survives `/clear` is human-readable freeform prose
in the handoff body. Block 0 of the Worktree-Anchored Resume Pattern
(`session-handoff-examples.md`) is explicitly SINGLE-LEVEL — it tells the
next session how to re-enter ONE worktree but has no syntax for a parent
chain.

### The Diagnosis

> "The broken link is not a missing socket and not a missing agent — it is a
> missing DATA STRUCTURE: there is no tree-shaped record of the
> worktree/session nesting topology anywhere in the codebase."

The maintainer's intuition ("a socket would preserve memory") is a CORRECT
DIAGNOSIS of the symptom but a PARTIAL fix for the cause. There is no field
to populate — even a perfect live socket would have nothing to carry.
Building transport before the payload is building a pipe with no water.

---

## §3. Flat-Primitive Survey

Every persistence primitive in the codebase is FLAT — none carries parent,
depth, or origin metadata:

| Primitive | Location | Schema | Lineage fields |
|---|---|---|---|
| Session registry `Entry` | `registry.go:86-95` | SessionID, SpecID, Phase, StartedAt, LastHeartbeat, PID, Host, CWD | **NONE** — CWD is the only worktree link (flat string) |
| Handoff `pending.json` | `pending.go:54-68` | SchemaVersion, SpecID, Phase, SavedAt, SavedBySession, ConversationLanguage, Directives, EmbeddedGoal, Body | **NONE** — singular per-project, OVERWRITTEN on each save (pending.go:118) |
| Worktree `Snapshot` | `state_guard.go:57-65` | SchemaVersion, CapturedAt, SnapshotID, HeadSHA, Branch, PorcelainLines, UntrackedSpecs | **NONE** — divergence-detection only, then discarded |
| Factory `Record` | `record.go:55-88` | flat per-session, keyed by SessionID | **NONE** — no parent/depth/origin fields |

The `pending.go:118` overwrite is the **most damaging** flat primitive: a
depth-3 worktree saving a handoff CLOBBERS the depth-1 record. The singular
record design actively destroys nesting information.

---

## §4. Why Past Attempts Fell Short

Six prior approaches were surveyed. Each solved a DIFFERENT problem than
depth-3 amnesia:

### 4.1 Kanban (SPEC-KANBAN-BOARD-001 / BOOTSTRAP / RENAME / WORKTREE)

Solves WIP-control for 4–5 parallel worker sessions — coordinating dispatch,
not tracking nested worktree origin. The card record carries only
SPEC-id/column/holder-session/transition-instant — NO parent worktree, NO
origin chain, NO resume pointer. Column-transition history was EXPLICITLY
REJECTED as the accepted cost of a gitignored store. All four SPECs are
plan-phase only (zero run-phase code); the RENAME that shipped was REVERTED
the next day.

### 4.2 Factory Mode / -f (SPEC-FACTORY-MODE-001 + SPEC-FACTORY-BOOTSTRAP-001)

The advertised "Leader socket" is a **DEAD STUB**. `factory.go:102-105` sets
`MOAI_FACTORY_LEAD_ADDR=/tmp/moai-factory-<runID>` with a comment acknowledging
no backing server exists. The factory Record is flat per-session with no
parent/depth/origin fields. It is scoped "one session drives one SPEC" and
cannot model epic → milestone → nested-worktree. -f gives the maintainer a
flag that promises memory preservation and delivers none.

### 4.3 Project Navigator (SPEC-PROJECT-NAVIGATOR-001/002/003/004)

A READ-ONLY DERIVED heuristic over SPEC frontmatter status + @NAV/@MX tokens
+ git-diff — a snapshot of the DISK, not of the session tree. None of its
outputs carry session/worktree/origin metadata. It answers "what does the
disk say is pending", not "where did I leave off, in which worktree, at what
depth".

### 4.4 Epic Status (SPEC-EPIC-STATUS-001)

Pure read-only rollup with ZERO session/worktree awareness. It shows WHAT
milestone gaps exist but never tracks WHO is working on which milestone, in
which worktree, at what depth. It is the CLOSEST existing tool to "where to
resume" but answers at milestone-aggregate granularity.

### 4.5 Handoff (session-handoff.md + pending.go)

`pending.json` is SINGULAR per-project — one record, OVERWRITTEN on each
save. A depth-3 worktree saving a handoff CLOBBERS the depth-1 record. The
schema carries no parent/depth/origin fields. Block 0 is structurally
single-level — no syntax for a chain. This IS the manual handoff the
maintainer explicitly wants to escape.

### 4.6 manager-lead (SPEC-HIERARCHICAL-TEAM-001)

An IN-SESSION mechanism. Context-Folding bounds context WITHIN one session
via `/compact`; its OWN documented fallback IS the manual handoff the
maintainer wants to escape. The depth-2 seal (`manager_lead_depth_test.go`)
is in DIRECT TENSION with the maintainer's depth-3 desire — the CI guard
forbids exactly what the maintainer wants. Zero observed runtime invocations
in MEMORY.md confirms the old framing was never exercised.

---

## §5. Winner Design Rationale — Origin-Trail Chain

### Why It Won

Origin-Trail Chain puts the **tree-shaped WorktreeNode record FIRST** and
treats kanban/todo/lead as CONSUMERS that tag themselves with
`chain_node_id`. This is the cleanest implementation of the diagnosis's key
improvement: "decouple delegation-efficiency from context-continuity; solve
continuity first."

The chain record carries `resume_target` + `resume_command` +
`last_completed_milestone` as NAMED FIELDS, directly answering all three of
the maintainer's questions. It respects the frozen Entry schema
(REQ-COORD-024) by being an OVERLAY that joins on SessionID at read time —
never mutating the frozen struct.

### Grafted Ideas (from Runner-Up Designs)

Two ideas were grafted from the close runners-up:

1. **Append-only JSONL substrate** (from LAK design): Origin-Trail Chain's
   original `tree.json` was a mutable JSON with atomic-write — vulnerable to
   whole-file corruption and still a single-record overwrite target. Replacing
   it with append-only JSONL at `.moai/state/chain/events.jsonl` (one
   node-lifecycle event per line) gives line-level corruption isolation and
   structurally prevents the clobber anti-pattern.

2. **Mechanical completion-event hook** (from LMK design): a NEW hook
   `.claude/hooks/moai/chain-event.sh` fires on SubagentStop and appends the
   leaf's completed/blocked/resume event to the ledger mechanically — WITHOUT
   leaf-agent cooperation. This is critical: completion tracking does not
   depend on the lead being alive or the leaf remembering to write.

### Design Decisions Locked by User (8 Policies)

The following 8 policies were confirmed via AskUserQuestion (2026-08-13) and
are NON-NEGOTIABLE design constraints:

| # | Policy | Decision | Phase scope |
|---|---|---|---|
| 1 | Depth ceiling | 3 (depth-2 seal rewritten as configurable ceiling in Phase 6, NOT this SPEC) | P6 |
| 2 | Mode flag | `-k` (factory→kanban rename, `-f` deprecated — Phase 5; THIS SPEC is flag-agnostic) | P5 |
| 3 | Scope | Single-host v1 (remote worktree = documented limitation) | P1 |
| 4 | Completion hook | Phase 1 ONWARD (chain-event.sh, no leaf cooperation) | P1 |
| 5 | Compaction | Auto + manual (manual delivered in P1; auto deferred — staleness-display is v1 "auto" surface) | P1+ |
| 6 | Registry | Overlay (chain joins Entry by SessionID at read; Entry UNTOUCHED) | P1 |
| 7 | Stale node | Heartbeat auto (reuse registry LastHeartbeat, default 15min) | P1 |
| 8 | Web dashboard | Excluded (CLI + hook surfaces only) | — |

---

## §6. Implementation Phases (Roadmap)

The winner design decomposes into 7 phases. This SPEC delivers **Phase 1**
only — the chain record + query CLI + completion hook + SessionStart banner.
Each subsequent phase consumes the chain substrate but does not modify it.

| Phase | Summary | Dependency |
|---|---|---|
| **P1 (THIS SPEC)** | Chain record + query CLI + completion hook + SessionStart banner | — |
| P3 | `/moai:todo` inbox (chain-stamped `backlog.jsonl`) | Consumes `MOAI_CHAIN_NODE_ID` |
| P4 | Multi-record handoff fix (`pending.go:118` clobber → per-node `<node_id>.json`) | Consumes chain node identity |
| P5 | `-k` kanban mode + derived board from `backlog.jsonl` | Consumes chain partitioning |
| P6 | `manager-lead` rewrite + depth-2 seal → depth-ceiling CI rewrite | Consumes chain + backlog |
| P7 | Live push channel (optional, measure-first) | Only if file-poll proves insufficient |

**Key principle**: P1 is independently shippable and solves depth-3 amnesia
ALONE. Delegation, kanban, and push transport are orthogonal concerns layered
on top in later phases.

---

## §7. Open Questions Resolved

The diagnostic workflow identified 8 open questions. All were resolved by
the user-confirmed 8 policies (§5) and the design decisions in design.md:

| Question | Resolution |
|---|---|
| Depth-ceiling default | 3 (policy 1; rewrite deferred to P6) |
| Compaction trigger | Manual `moai chain prune` + deferred auto (policy 5; design.md §5) |
| `-k` / `-f` interaction | This SPEC is flag-agnostic (policy 2; interaction resolved in P5) |
| Completion hook timing | Phase 1 onward (policy 4; REQ-CHAIN-012) |
| Cross-machine worktrees | Single-host v1, documented limitation (policy 3; REQ-CHAIN-016) |
| Registry Entry freeze | Overlay join, never mutate (policy 6; REQ-CHAIN-014) |
| Stale-node lifecycle | Heartbeat auto, 15min threshold (policy 7; REQ-CHAIN-015) |
| Web dashboard scope | Excluded (policy 8; spec.md §I) |

---

## §8. Socket-vs-Handoff Analysis (Why Not a Socket)

The maintainer's "socket will remember" intuition over-attributes the fix to
transport. The diagnostic's verdict:

> "The remembering is done by the record, the socket only delivers it faster."

Today's "cross-session transport" is a mix of four things, NONE a live socket:
(a) filesystem polling of active-sessions.json; (b) one-shot SessionStart
stderr reminders; (c) the singular fire-once handoff pending.json; (d) Claude
Code's opaque native SendMessage. The factory leader socket
(`/tmp/moai-factory-<runID>`) is explicitly a cosmetic grep-friendly path with
NO backing server (verified: the only `net.Listen` calls serve the unrelated
web dashboard).

WHERE the intuition is partially correct: a live channel would give "what
just completed" PUSH. Today a depth-3 leaf's result lands only in its
spawning lead's tool-result context, which dies with the lead session. But
even that requires the parent-chain data structure first — a push notification
with no origin-chain metadata is just noise.

VERDICT: the cheapest useful fix is NOT a socket. It is the chain record
(this SPEC). A genuine live socket is Phase 7 (optional, measure-first).

---

## §9. References

- `/tmp/wf_diag.txt` — full diagnostic output (root cause + why-past-attempts)
- `/tmp/wf_synth.txt` — synthesis (winner design + grafted ideas + phases)
- `/tmp/wf_designs.txt` — three competing design approaches (287 lines)
- `internal/session/registry.go:86-95` — Entry struct (frozen, REQ-COORD-024)
- `internal/hook/handoff/pending.go:54-68,118` — PendingRecord schema + clobber flaw
- `internal/worktree/state_guard.go:57-65` — worktree Snapshot (divergence detection)
- `internal/factory/record.go:55-88` — factory Record (flat, no lineage)
- `internal/cli/factory.go:102-105` — dead leader-socket stub
- `.claude/rules/moai/workflow/session-handoff.md` — Block 0 single-level pattern
- `.claude/rules/moai/workflow/worktree-integration.md` — L1/L2/L3 terminology

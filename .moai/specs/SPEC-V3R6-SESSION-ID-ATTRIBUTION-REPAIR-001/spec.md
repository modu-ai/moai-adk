---
id: SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001
title: "Session-ID attribution dead-feature repair (multi-session coordination Layer 2 race-attribution)"
version: "0.1.0"
status: completed
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli/session.go; internal/hook/session_start.go; internal/session/registry.go"
lifecycle: spec-anchored
tags: "session, attribution, multi-session, coordination, race-attribution, doctrine"
era: V3R6
---

# SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001

## §A. Problem Statement

The multi-session coordination policy (Layer 2 — session correlation for race
attribution) depends on the orchestrator's ability to discover **its own
session UUID** and emit it as `source_session_id: <UUID>` in paste-ready resume
messages (Block 2) and in auto-memory `project_*.md` files. The doctrine at
`.claude/rules/moai/workflow/session-handoff.md` Block 2 ASSUMES:

> "The session_id is the same value emitted by `moai session list --json` and
> stored in `.moai/state/active-sessions.json` — readers can correlate the
> resume back to its originating session."

**That assumption does NOT hold.** Empirical audit of 112 paste-ready resume
entries across recent SPEC work found **102 (91%) carry the environment-fallback
placeholder string** `source_session_id: <not-available — environment-fallback,
next session will backfill via /moai session register on activation>` (or one
of 3 other variant spellings on canonical doctrine surfaces; 13 others appear
in `.moai/docs/` prose and historical memory entries, which are append-only
and out of M4+M5 scope per §G). The fallback is NOT machine-generated — it is
the orchestrator LLM hand-typing the doctrine placeholder because it has **no
way to discover its own UUID**.

This is the **D4 defect** (track 3 of the 2026-06-16 session-handoff defect
analysis). Track 1 (D1/D2/D7/D8) and track 2 (D3/D5a/D6/D9 — SSOT↔render drift)
are SEPARATE and out of scope for this SPEC. Track 2 was resolved by
`SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001` (completed); this SPEC handles the
remaining track 3 D4.

### §A.1 The defect's essence — three sentences

**Write path exists but orchestrator-visible read path is absent, and the
registry runtime is empty.** (1) The SessionStart hook writes to the registry
when `input.SessionID != "" && input.ProjectDir != ""`
(`internal/hook/session_start.go:66` gate, `internal/hook/session_start.go:804`
`reg.Register(...)` call), so a write path exists. (2) But there is NO
subcommand that returns "this orchestrator's own UUID" — `moai session list`
returns the full entry list and the orchestrator cannot identify which entry is
its own among multiple concurrent sessions. (3) Worse, `.moai/state/active-sessions.json`
is observed absent at the current baseline (`ls` exit 1) and `moai session list
--json` returns `[]` with exit 0 — meaning even THIS session's own write did not
land. The orchestrator therefore falls back to the doctrine placeholder 91% of
the time.

### §A.2 Root blocker — Claude Code runtime non-exposure

`session.id` is available only in hook stdin JSON
(`internal/cli/hook.go:675` T-A3 spec comment, L699 `SessionID:
hookInput.Session.ID`, L788 `ParentSessionID: hookInput.Session.ID`). The
Claude Code runtime itself does NOT directly expose `session.id` to the
orchestrator (LLM) as an env var or tool input — only hook handlers see it.
A full read path therefore requires either (a) runtime support (out of
MoAI-ADK's control) or (b) the 2-stage SessionStart `additionalContext`
injection + `moai session current` complement defined in this SPEC.

## §B. Scope

### §B.1 In scope (3 pillars — user decision A, full repair)

**(P1) Write-path reliability investigation (GATING predecessor).** Determine
why the registry stays empty even after a SessionStart hook runs. Investigation
targets: (a) the SessionStart hook condition under which it does NOT call
`reg.Register`; (b) `handle-session-start.sh` 3-tier fallback behavior (PATH →
`~/go/bin/moai` → `~/.local/bin/moai` → silent `exit 0`); (c) the
`input.SessionID == ""` empty-value scenario (does the runtime emit an empty
session_id on some activation paths?). P1 outcome may reshape the P2/P3 pillar
order; P1 MUST be resolved before P2 is considered useful (empty registry → P2
read returns nothing).

**(P2) Read-path addition (2 stages).** Stage 1: NEW subcommand `moai session
current` (in `internal/cli/session.go`) that returns THIS orchestrator's own
UUID. Stage 2: SessionStart hook `additionalContext` / `hookSpecificOutput`
injection in `internal/hook/session_start.go` so the orchestrator receives its
UUID at session start (additionalContext surfaces as context injected into the
LLM at hook completion).

**(P3) Fallback doctrine canonicalization.** Consolidate the 4 distinct
`source_session_id` variants observed on canonical doctrine surfaces
(independent re-grep, see research.md §C) into ONE canonical fallback form,
synchronized across both SSOT surfaces:
- `.claude/rules/moai/workflow/session-handoff.md` Block 2 environment fallback
- `.claude/output-styles/moai/moai.md` §8 (render surface)

The 4 variants comprise 1 canonical fallback + 3 non-canonical (of which 2
are happy-path UUID-source template slots to rewrite for symmetry with the
P2 `current` subcommand, and 1 is an illustrative-example placeholder to
collapse to the canonical fallback). A broader sweep including `.moai/docs/`
and auto-memory finds 14 strings total, but only the 4 doctrine-surface
variants are in M4+M5 scope (the other 10 are historical append-only memory
entries excluded per §G and `.moai/docs/` prose).

The environment-fallback pattern stays as graceful degradation (per
session-handoff.md it is NOT an anti-pattern), but it becomes the single
canonical fallback string.

### §B.2 Out of scope

- **Track 2 (SSOT↔render drift)** — already resolved by `SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001` (completed). The deferred Go lint rule that detects doctrine drift is a SEPARATE independent code SPEC.
- **Track 1 defects (D1/D2/D7/D8)** — resolved by commit `82ea1b09f`.
- **Compaction UUID-loss limitation** — `additionalContext` is lost after `/clear`/compaction; this is a runtime limitation that persists after this SPEC. Documented as residual risk (§F.3), not a defect this SPEC repairs.
- **Runtime-level session.id exposure** — out of MoAI-ADK's control; the workaround (P2 2-stage) is the canonical path.

## §C. Requirements (GEARS notation)

### §C.1 P1 — Write-path reliability (gating predecessor)

**REQ-WPR-001 (Ubiquitous):** The CLI SHALL provide a diagnostic command
`moai session doctor` that reports (a) whether the registry file exists,
(b) whether the current process's session_id is present in it, (c) the
SessionStart hook condition that prevented registration if applicable.

**REQ-WPR-002 (When):** **When** `moai session list --json` returns `[]` for a
host that has at least one active Claude Code session, the CLI SHALL emit a
diagnostic explaining the likely root causes (empty `input.SessionID`, hook
wrapper fallback silent-exit, or registry write failure).

**REQ-WPR-003 (When):** **When** the SessionStart hook handler
(`internal/hook/session_start.go`) detects that `input.SessionID == ""`, the
handler SHALL emit a warning (stderr, non-blocking) so the orchestrator can
observe that the write path was bypassed.

**REQ-WPR-004 (While):** **While** the P1 investigation is incomplete, the SPEC
plan (plan.md §F milestones) SHALL NOT mark P2 milestones as ready-to-start
until P1 root cause is documented in research.md §D.

### §C.2 P2 — Read-path addition

**REQ-RDP-001 (Ubiquitous):** The CLI SHALL provide a `moai session current`
subcommand that returns the orchestrator's own session UUID by reading the
runtime-available session identifier.

**REQ-RDP-002 (Where):** **Where** the runtime exposes `session.id` via an env
var or a side-channel file written by the SessionStart hook, the `current`
subcommand SHALL resolve and return that UUID with exit 0.

**REQ-RDP-003 (When):** **When** the runtime does NOT expose `session.id` to
the CLI subprocess (the default state today), the `current` subcommand SHALL
return exit 1 with a stderr diagnostic pointing to the SessionStart-hook
injection path as the alternative.

**REQ-RDP-004 (Ubiquitous):** The SessionStart hook handler
(`internal/hook/session_start.go`) SHALL inject the orchestrator's own UUID
into the hook response via `additionalContext` (or `hookSpecificOutput`) so
the Claude Code runtime surfaces it to the orchestrator at session start.

**REQ-RDP-005 (When):** **When** `additionalContext` injection is added, the
handler SHALL preserve the existing behavior (Register → Purge → Query → stderr
surface) — the injection is strictly additive.

**REQ-RDP-006 (Where):** **Where** the registry runtime is empty (P1 unresolved
or runtime non-exposure), the `current` subcommand SHALL gracefully degrade by
emitting the canonical fallback string (REQ-FBC-001) rather than an opaque
error.

### §C.3 P3 — Fallback doctrine canonicalization

**REQ-FBC-001 (Ubiquitous):** The doctrine SHALL define exactly ONE canonical
environment-fallback string for `source_session_id`. The canonical form is:
`source_session_id: <not-available — environment-fallback, next session will backfill via /moai session register on activation>`.

**REQ-FBC-002 (Ubiquitous):** The canonical fallback string SHALL appear
verbatim and identically in BOTH SSOT surfaces:
- `.claude/rules/moai/workflow/session-handoff.md` Block 2 environment fallback clause
- `.claude/output-styles/moai/moai.md` §8 Block 2 checklist item

**REQ-FBC-003 (Shall not):** The doctrine SHALL NOT retain any of the 3
non-canonical `source_session_id` variants on canonical doctrine surfaces
(rules tree + output-styles + template mirrors; see research.md §C for the
enumerated list). Each non-canonical variant is either a happy-path
UUID-source template slot (rewrite to `source_session_id: <UUID from moai
session current>` for symmetry with the P2 `current` deliverable) or an
illustrative-example placeholder (collapse to the canonical fallback). The
broader-context sweep (14 variants including `.moai/docs/` + memory) is
informational for drift-detector calibration and NOT in M4+M5 elimination
scope (historical memory entries are append-only per §G).

**REQ-FBC-004 (When):** **When** a future doctrine edit needs to reference the
fallback pattern, the edit SHALL cite the canonical string verbatim (no
paraphrase, no abbreviation).

### §C.4 Cross-cutting — multi-session coordination Layer 2 preservation

**REQ-MSC-001 (While):** **While** this SPEC repairs the attribution feature,
the existing multi-session coordination Layer 1 primitives
(`moai session register/heartbeat/deregister/list/purge`) SHALL remain
behaviorally unchanged.

**REQ-MSC-002 (Shall not):** The implementation SHALL NOT alter the
`FormatStderrReminder` "others-active-session" semantics
(`internal/session/registry.go:424-448`) — that function warns about OTHER
sessions and correctly returns empty when `others == 0` (L431-432); it is NOT
the self-UUID surface and must not be repurposed as one.

**REQ-MSC-003 (Where):** **Where** the P2 `current` subcommand is added, the
orchestrator paste-ready resume template (session-handoff.md Block 2) SHALL
cite `moai session current` as the primary UUID source and the canonical
fallback as the degradation path.

## §D. Constraints

1. **Behavior preservation** — existing `moai session register/heartbeat/deregister/list/purge` semantics unchanged (REQ-MSC-001).
2. **Hook timeout budget** — SessionStart hook stays within the MoAI-default 5s budget (`additionalContext` injection is a single map assignment, O(1)).
3. ** Doctrine parity** — both SSOT surfaces (session-handoff.md Block 2 + moai.md §8) MUST carry the identical canonical fallback string (byte-parity across surfaces).
4. **Test isolation** — all tests use `t.TempDir()`; no writes to the real `.moai/state/active-sessions.json` from test code.
5. **Lint-rule drift prevention** — the canonical fallback string is the single anchor; any doctrine edit referencing the fallback MUST use the verbatim string.

## §E. Alternatives Considered

### §E.1 Alternative B — Deprecate the feature (rejected)

Drop the `source_session_id` field from the resume template entirely. Accept
the feature loss; multi-session race attribution degrades to git-state-only
detection.

**Rejected because:** the multi-session coordination Layer 2 purpose
(race attribution across `~/.claude/projects/{hash}/memory/` shared state) is
load-bearing — the 2026-05-16 sync-phase race incident
(`L_parallel_session_self_reconciles_shared_tree`) demonstrated real races
that the UUID correlation is designed to attribute. Dropping the feature
removes the only mechanism for correlating a paste-ready resume back to its
originating session.

### §E.2 Alternative C — Minimal read path only (rejected)

Implement `moai session current` only; skip P1 investigation and P3 doctrine
consolidation.

**Rejected because:** without P1, the registry stays empty and `current`
returns nothing meaningful (the read path has nothing to read). Without P3,
the 3 non-canonical doctrine-surface variants continue to proliferate and
the drift detector's exact-token matching fails on non-canonical spellings.
Both gaps must close for the feature to actually work.

### §E.3 Alternative D — Runtime support request (rejected)

Request upstream Claude Code runtime to expose `session.id` as an env var.

**Rejected because:** out of MoAI-ADK's control and not actionable within this
SPEC's scope. The P2 2-stage approach (SessionStart additionalContext + CLI
complement) is the canonical workaround; the residual compaction-loss
limitation is documented as accepted (§F.3).

## §F. Residual Risks

### §F.1 Registry write path may have multiple root causes

The P1 investigation may uncover that the empty registry has more than one
cause (empty `input.SessionID`, hook wrapper silent-exit, race between
SessionStart and the first orchestrator turn). Each cause requires its own fix
and the milestone order may shift. Mitigation: P1 is explicitly the gating
predecessor (REQ-WPR-004); plan.md milestones are written to allow P1 outcome
to reshape P2/P3 order.

### §F.2 Compaction UUID-loss limitation persists

`additionalContext` injected at SessionStart is lost after `/clear` or context
compaction. After compaction, the orchestrator no longer "remembers" its UUID
unless it re-reads via `moai session current` (which only works if P1 is
resolved and the registry is populated). This is a runtime limitation that
this SPEC does NOT repair. Mitigation: document in session-handoff.md Block 2
that post-compaction turns SHOULD re-query `moai session current` before
emitting a resume.

### §F.3 Runtime non-exposure remains the root barrier

Even after P2, the orchestrator only receives its UUID via the SessionStart
injection (not directly from the runtime as an env var). Any activation path
that bypasses SessionStart (rare, but possible during headless `-p` invocations
without hooks) leaves the orchestrator without UUID access. Mitigation:
`moai session current` returns the canonical fallback in this case
(REQ-RDP-006) and the orchestrator's resume template cites the fallback as
graceful degradation.

## §G. Exclusions (What NOT to Build)

- A new persistent session-state store. The existing
  `.moai/state/active-sessions.json` registry is reused; this SPEC adds a read
  path, not a new store.
- A runtime-level session.id exposure mechanism. That is Claude Code runtime
  territory.
- The deferred Go lint rule that detects SSOT↔render drift (track 2, separate
  SPEC).
- Changes to `FormatStderrReminder` semantics (it remains the "others-active"
  warning surface, NOT the self-UUID surface — see REQ-MSC-002).
- Backfilling the 91% of historical paste-ready entries that carry the
  placeholder. Historical entries are append-only per the Lessons Protocol;
  only NEW resume messages benefit from the repair.

## §H. Cross-References

- **Predecessor (track 2)**: `SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001` (completed) — resolved the SSOT↔render drift; this SPEC handles the remaining track 3 D4.
- **Predecessor (multi-session coordination Layer 1)**: `SPEC-V3R6-MULTI-SESSION-COORD-001` — defined the registry primitives this SPEC builds a read path over.
- **Doctrine SSOT**: `.claude/rules/moai/workflow/session-handoff.md` Block 2 + `additionalContext` semantics.
- **Render surface**: `.claude/output-styles/moai/moai.md` §8 (Block 2 checklist).
- **Verification doctrine**: `.claude/rules/moai/core/verification-claim-integrity.md` — research.md §A-C apply the 5-section evidence-bearing format.
- **Lessons applied**: `feedback_windowed_grep_undercount_authoring` (use content-anchored grep, not `sed -n` windowed reads), `feedback_coverage_audit_table_not_actually_run` (re-run every N-count claim independently).

## §I. History

- 2026-06-17: Initial draft (manager-spec, plan-phase). Track 3 D4 repair; diagnosis complete from prior read-only session (baseline `d0e2e9bc3`); re-grep verification against current HEAD `12e20d190` found the fallback-variant count is 14, not 9 as the baseline diagnosis claimed — see research.md §C.
- 2026-06-17: iter-2 remediation (manager-spec, plan-phase). plan-auditor iter-1 returned FAIL 0.78 (2 BLOCKING + 1 SHOULD-FIX + 3 MINOR). Fixes: D1 — research.md §C Evidence re-scoped to canonical doctrine surfaces (4 variants) with `.moai/specs/` excluded (was self-referentially polluted, producing 16 not 14); REQ-FBC-003 / AC-FBC-003 counts updated from "13 non-canonical" to "3 non-canonical on canonical doctrine surfaces". D2 — AC-FBC-005 added for REQ-FBC-004 traceability (AC-FBC-004 was already taken by template-mirror parity tracing to REQ-FBC-002). D3 — variant #10 disposition resolved (RETAIN as happy-path UUID-source slot, rewrite for symmetry). D4/D5/D6 — MINOR notes added (plan.md M2→M3 dependency, AC-RDP-002 P1-outcome-conditional, REQ-WPR-001 subcommand name pinned). Direction A and 3-pillar decomposition unchanged (passed audit).

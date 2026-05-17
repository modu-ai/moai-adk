---
id: SPEC-WORKTREE-DOCS-001
title: "Worktree Workflow Documentation Harmonization (L1/L2/L3 Opt-In Policy)"
version: "0.1.0"
status: completed
created: 2026-05-17
updated: 2026-05-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow"
lifecycle: spec-anchored
tags: "worktree, documentation, harmonization, override, opt-in, terminology"
---

# Worktree Workflow Documentation Harmonization (L1/L2/L3 Opt-In Policy)

## HISTORY

| Version | Date       | Author       | Change                                                              |
|---------|------------|--------------|---------------------------------------------------------------------|
| 0.1.0   | 2026-05-17 | manager-spec | Initial draft. Resolves forensic-audit 2026-05-17 drift (items 7-12). |

## 1. Overview

### 1.1 Goal

Harmonize MoAI-ADK-go's worktree workflow documentation with the user's 2026-05-17 policy
re-definition. Replace [HARD] mandates that force worktree usage with SHOULD/opt-in
guidance, standardize the L1/L2/L3 terminology used across rules, and update the
project's auto-memory hierarchy so future agents inherit the new doctrine on session start.

### 1.2 Why now

The 2026-05-17 forensic audit (`.moai/research/worktree-forensic-audit-2026-05-17.md`)
documented three P0 systemic drift findings: orphaned Wave 5 primitive (out of scope for
this SPEC, see §1.3), undocumented user override (2026-05-15
`feedback_worktree_never_use.md`), and L1/L2/L3 terminology conflation. Today the user
SUPERSEDED the 2026-05-15 "worktree 영구 미사용" directive with a new opt-in / Claude Code
runtime autonomous policy. Until rules reflect the new policy, new sessions reading
`spec-workflow.md` Step 2/3 [HARD] or `CLAUDE.md` §14 [HARD] will follow stale mandates
that contradict current user intent.

### 1.3 Impact

- **Affected files**: 5 rule files in `.claude/rules/moai/workflow/` + `CLAUDE.md` §14
  + 2 auto-memory files (1 NEW feedback + MEMORY.md index update). Estimated total
  line delta: ~80 LOC across 6 file edits + ~60 LOC new memory entry.
- **Affected users**: All future MoAI-ADK-go sessions that consult worktree workflow rules.
- **Backward compatibility**: The Wave 5 `worktree-state-guard.md` primitive remains
  functional (manually invocable); only its trigger conditions change from "MUST use"
  to "available when worktree is opt-in by user."
- **Token-load**: Net-neutral (advisory rewrites do not significantly change rule body
  size). Terminology prefixes add ~10 tokens per rule file.

## 2. Goals

1. Convert all `[HARD]` mandates that force worktree creation/reuse in
   `spec-workflow.md` Step 2/3 and `CLAUDE.md` §14 to SHOULD/opt-in advisory rules.
2. Establish a canonical L1/L2/L3 terminology vocabulary and apply it consistently
   across 5 worktree-related rule files.
3. Add an explicit terminology glossary to `worktree-integration.md` § Overview so
   that future agents disambiguate "worktree" mentions on first read.
4. Supersede the 2026-05-15 `feedback_worktree_never_use.md` auto-memory with a new
   `feedback_worktree_autonomous.md` reflecting the 2026-05-17 opt-in policy.
5. Preserve backward compatibility: existing `worktree-state-guard.md` primitive
   (snapshot/verify/restore) remains callable; agents can still use
   `Agent(isolation: "worktree")` — Claude Code runtime decides per-call.

## 3. Non-Goals

### 3.1 Out of Scope

The following items are deliberately deferred (Non-Goals). Listed explicitly to prevent scope creep:

- **VERIFY-001 (Wave 5 orchestrator wiring)** — DROPPED per user decision 2026-05-17.
  Forensic audit items 1-6 (orchestrator-side snapshot/verify/restore invocation,
  `--verify-isolation` flag, integration tests in `internal/cli/launcher_*`).
  The Wave 5 primitive remains operationally dormant (callable manually but not
  auto-invoked by orchestrator skills).
- **Forensic audit items 13-27** — deferred to separate follow-up SPECs:
  - Items 13-15: L1/L2 bridge runtime test cases (Go test files)
  - Items 16-18: Wave 5 primitive warning banners + reference reclassification
  - Items 19-21: Agent prompt anti-pattern catalog expansion
  - Items 22-24: Isolation failure recovery protocol in `agent-common-protocol.md`
  - Items 25-27: CI/regression test suite for isolation validation
- **L1 Claude Code runtime field semantics investigation** — outside moai-adk-go
  scope; needs Anthropic schema reference for `worktreePath` response field.
- **`.claude/rules/moai/project-overrides/` directory creation** — design.md AD-002
  recommends Option A (single source updated rules) over Option B (override directory).
  Override directory mechanism deferred until a real project-vs-global rule conflict
  arises that cannot be expressed inline.
- **Ambiguity resolution: "What if a user starts `--worktree` but later wants to
  abandon it?"** — Resolution deferred to follow-up SPEC if encountered. Per user
  2026-05-17 policy, abandonment is allowed (worktree usage is opt-in); recovery
  procedure (cleanup of partial state) not documented in this SPEC.

## 4. EARS Requirements

### REQ-WTD-001 — Override Policy Documentation (Ubiquitous)

The system SHALL document the 2026-05-17 user worktree policy re-definition in every
rule file that previously contained `[HARD]` mandates forcing worktree usage. The
documentation MUST explicitly state:

- L1 (`Agent(isolation: "worktree")`) decision is delegated to Claude Code runtime;
  MoAI orchestrator does NOT mandate isolation.
- L2 (`moai worktree new <SPEC-ID>`) is user opt-in only; default flow does not invoke L2.
- L3 (`--worktree` flag at plan/run time) is user opt-in only; default flow does not
  invoke L3.

Affected files: `spec-workflow.md`, `CLAUDE.md` §14, `worktree-integration.md`,
`worktree-state-guard.md`, `session-handoff.md`.

### REQ-WTD-002 — Terminology Standardization (Ubiquitous)

The system SHALL use the canonical L1/L2/L3 terminology vocabulary in all
worktree-related rule files. The vocabulary:

- **L1 — Claude Code Native Worktree**: triggered by `Agent(isolation: "worktree")`
  frontmatter. Owned by Claude Code runtime. Path: `.claude/worktrees/<auto-name>/`.
- **L2 — SPEC Worktree**: created via `moai worktree new <SPEC-ID>`. Owned by moai CLI.
  Path: `~/.moai/worktrees/{project}/{SPEC-ID}/`.
- **L3 — Plan Worktree**: triggered by `--worktree` flag at plan time. Owned by moai
  workflow skill (delegates to L2 mechanism for creation).
- **git worktree**: underlying git mechanism. Reserved for low-level descriptions only.

Each prose mention of "worktree" in the affected files MUST be either (a) prefixed with
L1/L2/L3 disambiguation, (b) followed by parenthetical clarification (e.g., "worktree
(L2 SPEC worktree)"), or (c) part of a documented compound noun in the glossary.

Affected files (same 5 as REQ-WTD-001).

### REQ-WTD-003 — Soft-Mandate Rule Conversion (Event-driven)

WHEN a rule references worktree usage in `spec-workflow.md` Step 2/3 or `CLAUDE.md` §14,
THEN the system SHALL convert `[HARD]` "MUST" mandates to `SHOULD` advisory rules,
preserving rule structure and cross-references but downgrading enforcement strength.

Specifically:

- `spec-workflow.md` Step 2 row text "`moai worktree new SPEC-XXX --base origin/main`
  then `/moai run SPEC-XXX`" becomes optional (single phrase MAY become "create
  worktree if user opted in; otherwise run in main checkout on `feat/SPEC-XXX`
  branch").
- `spec-workflow.md` Step 3 text "(same worktree as Step 2)" becomes "(same worktree
  as Step 2 if worktree was used; otherwise same branch)".
- `spec-workflow.md` § Run Phase intro "[HARD] Execute in a fresh SPEC worktree..."
  becomes "[SHOULD] When user has opted into worktree (L2/L3), execute in a fresh
  SPEC worktree; otherwise execute in the plan-PR feature branch on main checkout."
- `spec-workflow.md` § Sync Phase intro [HARD] becomes similarly advisory.
- `CLAUDE.md` §14 bullet 1 "Implementation teammates ... MUST use `isolation:
  "worktree"`" becomes "When `Agent(isolation: "worktree")` is set, Claude Code
  runtime decides whether to isolate."
- `CLAUDE.md` §14 bullet 2 (read-only teammates MUST NOT) is preserved as advisory
  guidance — Claude Code runtime still benefits from the hint.
- `CLAUDE.md` §14 bullet 3 SHOULD remains SHOULD.
- `CLAUDE.md` §14 bullet 4 GitHub workflow fixer MUST becomes SHOULD.

### REQ-WTD-004 — Memory Hierarchy Update (Ubiquitous)

The system SHALL update the project auto-memory hierarchy to reflect the 2026-05-17
policy:

- A NEW file `feedback_worktree_autonomous.md` SHALL be created at
  `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/` with the metadata
  schema specified in MoAI lessons protocol (name, description, metadata.type).
  Content MUST describe the L1/L2/L3 opt-in policy verbatim per §1.2 above.
- The existing `feedback_worktree_never_use.md` SHALL receive a `[SUPERSEDED by
  feedback_worktree_autonomous]` prefix on its body's first heading per MoAI Lessons
  Protocol.
- `MEMORY.md` SHALL gain a one-line index entry pointing to
  `feedback_worktree_autonomous.md` AND mark the index entry for
  `feedback_worktree_never_use` with `[SUPERSEDED]` (the entry currently does not
  exist in the truncated MEMORY.md preview but if present at full length, must be
  patched).

### REQ-WTD-005 — Backward Compatibility (Ubiquitous)

The system SHALL preserve operational backward compatibility for:

- `internal/cli/worktree/guard.go` (`moai worktree snapshot|verify|restore`) — primitive
  remains callable from agent prompts or manual user invocation. Soft-mandate
  conversion in `worktree-state-guard.md` adds a "dormant unless isolation used"
  note but does NOT remove the primitive's documentation.
- `internal/cli/worktree/new.go` (`moai worktree new`) — CLI command remains functional;
  only the [HARD] rule citing it shifts from "MUST" to user opt-in.
- `.claude/skills/moai/workflows/plan.md` `--worktree` flag — flag remains accepted;
  only the rule layer reflects that its activation is user-opt-in (no policy change
  needed in plan.md itself since the flag was already optional).
- Existing SPEC-V3R3-CI-AUTONOMY-001 Wave 5 primitive's documented escalation path
  (snapshot → verify → AskUserQuestion restore/accept/abort) remains unchanged.

## 5. Affected Files

### MODIFY (5 rule files + CLAUDE.md)

| File | Estimated Line Delta | Change Summary |
|------|---------------------|----------------|
| `.claude/rules/moai/workflow/spec-workflow.md` | ~25 LOC modified | Step 2/3 [HARD] → SHOULD advisory; add user opt-in note + supersede reference to 2026-05-17 policy. |
| `.claude/rules/moai/workflow/worktree-integration.md` | ~30 LOC added | Add L1/L2/L3 terminology glossary to § Overview (~20 LOC); update prose worktree mentions with prefixes (~10 LOC). |
| `.claude/rules/moai/workflow/worktree-state-guard.md` | ~6 LOC added | Add note in § Overview: "Primitive is dormant unless Claude Code runtime opts into L1 isolation. Manual invocation supported via `moai worktree {snapshot,verify,restore}`." |
| `.claude/rules/moai/workflow/session-handoff.md` | ~8 LOC modified | Block 0 description: clarify "Required only when L3 `--worktree` was opt-in"; existing conditional logic already supports `--branch` default. |
| `CLAUDE.md` §14 | ~6 LOC modified | Soften 4 bullets per REQ-WTD-003; add 1-line cross-reference to user policy 2026-05-17. |

### NEW (2 auto-memory files)

| File | Estimated Line Delta | Change Summary |
|------|---------------------|----------------|
| `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_worktree_autonomous.md` | ~50 LOC new | Feedback memory describing L1/L2/L3 opt-in policy + Why + How to apply + supersede pointer. |
| `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/MEMORY.md` | ~2 LOC index entries | Add 1 new index entry for `feedback_worktree_autonomous`; patch `feedback_worktree_never_use` entry with `[SUPERSEDED]` marker. |

## 6. References

- Forensic Audit: `.moai/research/worktree-forensic-audit-2026-05-17.md` (576 lines, 34 KB)
- Superseded Policy: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/feedback_worktree_never_use.md` (2026-05-15)
- New Policy (this SPEC creates): `feedback_worktree_autonomous.md` (2026-05-17)
- Related Rules: `spec-workflow.md`, `worktree-integration.md`, `worktree-state-guard.md`, `session-handoff.md`, `CLAUDE.md` §14
- Cross-SPEC: SPEC-V3R3-CI-AUTONOMY-001 Wave 5 (W5-T08) — worktree-state-guard.md authoring SPEC. Wave 5 primitive remains operational.
- Cross-SPEC: SPEC-V3R2-WF-002 — original Block 0 paste-ready resume design; preserved by REQ-WTD-005.

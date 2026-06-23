---
id: SPEC-CC2186-BG-PERMISSION-RATIONALE-001
title: "Background subagent permission rationale alignment (CC 2.1.186)"
version: "0.1.0"
status: draft
created: 2026-06-23
updated: 2026-06-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/core + CLAUDE.md + internal/template/templates mirror"
lifecycle: spec-anchored
tags: "doctrine, cc-alignment, background-subagent, permission, rationale-correction, mirror-parity"
tier: M
---

# SPEC-CC2186-BG-PERMISSION-RATIONALE-001 — Background subagent permission rationale alignment (CC 2.1.186)

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-06-23 | manager-spec | Initial plan-phase artifact creation (draft). Single verified upstream-alignment drift surfaced by the release-update harness (Claude Code 2.1.186). Sibling pattern: SPEC-CC2178-TEAM-API-ALIGN-001 (completed). |

## §A. Context

The `/harness:release-update` harness analyzed the Claude Code 2.1.183..2.1.186 delta (research report: `.moai/research/cc-update-2.1.183-to-2.1.186.md`). Of that delta, exactly **one genuine doctrine drift** was surfaced — a stale *rationale* (not a stale conclusion) about how background subagents handle permission prompts.

### A.1 The upstream behavior change (verified)

Claude Code 2.1.186 changelog (verbatim): "Changed background subagents to surface permission prompts in the main session instead of auto-denying; the dialog shows which agent is asking, and Esc denies just that tool."

Before 2.1.186, a background subagent (`run_in_background: true`) that hit a non-pre-approved permission prompt would **auto-deny** because it could not interact with the user. As of 2.1.186, the prompt is **surfaced to the main session**, where the leader/user can respond (Esc denies just that one tool).

### A.2 What this invalidates — and what survives

This change invalidates the *stated mechanism* in MoAI doctrine, NOT the [HARD] behavioral conclusion:

- **Invalidated (the drift)**: the rationale clause "Background agents auto-deny ... **because they cannot interact with the user**". As of 2.1.186 they CAN interact with the user (via the main session), so the "cannot interact" causal claim is factually outdated.
- **Survives (the retained conclusion)**: the [HARD] behavioral rule itself — background subagents MUST NOT perform Write/Edit; use `run_in_background: false` for write tasks; read-only tasks remain safe in background. This conclusion stands on a *separate, surviving basis* (see §B below). MoAI's restriction is a deliberate conservative policy, not a derivation from the platform's old auto-deny mechanism.

This is the maintainer's explicit choice (recorded in the spawn directive): **rationale correction, conclusion retained** — NOT rule re-evaluation. This SPEC does NOT relax the background-write restriction.

## §B. The corrected rationale (target intent)

The corrected rationale must convey two surviving justifications for retaining `run_in_background: false` for write/Edit tasks, even though the platform no longer auto-denies:

1. **Allowlist non-inheritance (survives verbatim)**: even with `mode: "bypassPermissions"`, the background execution context still does not fully inherit the parent session's permission allowlist. This sentence is already present in the doctrine and is RETAINED as part of the corrected rationale — it was never dependent on the auto-deny mechanism.
2. **Flow-interruption / predictability (new survivor)**: surfacing a permission prompt in the main session for *each* background write interrupts the leader's flow — defeating the purpose of background (parallel) execution. MoAI therefore retains the conservative `run_in_background: false` for writes by policy.

Read-only tasks (research, analysis, review) remain safe in background — unchanged.

> **Run-phase edit-time caveat (NOT a verbatim mandate)**: the exact phrasing above is the *target intent*, not a literal string to paste. manager-develop MUST finalize the exact wording at run time AFTER re-confirming the 2.1.186 background-permission surface against the official sub-agents doc (`code.claude.com/docs/en/sub-agents`). The changelog bullet alone is terse and MUST NOT be the sole basis for the [HARD] rationale rewrite. See §G.1 (run-phase precondition).

## §C. GEARS Requirements

### REQ-BGR-001 — Rationale-clause correction (event-driven, primary drift)

**When** the corrected doctrine is in place, the agent-common-protocol § Background Agent Execution **shall not** contain the causal clause "because they cannot interact with the user" — the clause that 2.1.186 invalidated.

### REQ-BGR-002 — Auto-deny descriptor correction (ubiquitous)

The corrected doctrine **shall** replace the outdated "auto-deny" mechanism descriptor in CLAUDE.md §14 and zone-registry CONST-V3R2-020 with prose consistent with the 2.1.186 surface-to-main-session behavior, while preserving the directive that write tasks use `run_in_background: false`.

### REQ-BGR-003 — Conclusion-retention invariant (ubiquitous, conclusion preserved)

The corrected doctrine **shall** retain, in every corrected surface, the [HARD] behavioral conclusion that background subagents MUST NOT perform Write/Edit and that write tasks use `run_in_background: false`. The correction targets rationale wording only; it **shall not** weaken or relax the restriction.

### REQ-BGR-004 — Allowlist-non-inheritance retention (ubiquitous, survivor)

The corrected agent-common-protocol rationale **shall** retain the surviving justification that the background execution context does not fully inherit the parent session's permission allowlist (even under `bypassPermissions`).

### REQ-BGR-005 — Mirror parity (ubiquitous)

**Where** a corrected surface has a live copy and a template-mirror copy, the corrected prose **shall** be semantically identical (byte-identical for the prose) between the live tree and `internal/template/templates/`.

### REQ-BGR-006 — Template neutrality (capability gate)

**Where** the corrected prose lives in `internal/template/templates/**`, it **shall** contain no forbidden internal-content class — no SPEC-ID, REQ/AC token, internal date, commit SHA, audit citation, or memory/archive path. The rationale text is generic mechanism description (an acceptable content class).

### REQ-BGR-007 — Embedded regeneration (event-driven)

**When** any `internal/template/templates/` source file is edited, `make build` **shall** be run so `internal/template/embedded.go` reflects the corrected mirror content.

### REQ-BGR-008 — Edit-time doc re-confirmation (event-driven precondition)

**When** manager-develop finalizes the [HARD] rationale rewrite at run time, it **shall** first re-confirm the exact 2.1.186 background-permission surface against the official sub-agents doc (`code.claude.com/docs/en/sub-agents`), and **shall not** rewrite the [HARD] rationale from the changelog bullet alone.

## §D. Scope — 6 surfaces (3 loci × 2 trees)

| Locus | Live path | Mirror path | Drift element |
|-------|-----------|-------------|---------------|
| 1 — CLAUDE.md §14 | `CLAUDE.md` L289 | `internal/template/templates/CLAUDE.md` L289 | "auto-deny" descriptor (directive retained) |
| 2 — agent-common-protocol § Background Agent Execution | `.claude/rules/moai/core/agent-common-protocol.md` L190 | `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` L159 | "because they cannot interact with the user" clause (primary drift); allowlist sentence retained |
| 3 — zone-registry CONST-V3R2-020 | `.claude/rules/moai/core/zone-registry.md` L224 | `internal/template/templates/.claude/rules/moai/core/zone-registry.md` (CONST-V3R2-020 clause) | "auto-deny Write/Edit operations" verbatim clause |

> Line numbers are confirmed as of plan-phase (2026-06-23). manager-develop should grep at run time rather than trust line numbers, since intervening edits may shift them.

## §E. Self-Verification / Audit-Ready Signal

This section's run/sync evidence lives in `progress.md` §E. The plan-phase audit-ready signal (this SPEC's plan completeness) is recorded in progress.md §E.1.

## §F. Exclusions

This SPEC corrects a stale rationale only. The following are explicitly NOT in scope.

### Out of Scope — zone-registry CONST-V3R2-044 (pure conclusion, no false rationale)

- `.claude/rules/moai/core/zone-registry.md` L417 CONST-V3R2-044 clause = "Background subagents (run_in_background: true) MUST NOT perform Write/Edit operations." This is the pure behavioral conclusion with NO embedded false rationale — it carries no "auto-deny" descriptor and no "cannot interact" causal claim. It is the retained conclusion. KEEP AS-IS in both live and mirror trees. Touching it would be a scope violation.

### Out of Scope — the [HARD] behavioral rule (restriction retained, not relaxed)

- The behavioral rule "background subagents MUST NOT Write/Edit; use `run_in_background: false` for writes" is RETAINED unchanged. This SPEC does NOT relax, weaken, or re-evaluate the restriction. The maintainer explicitly chose "rationale correction, conclusion retained" over "rule re-evaluation".
- The CLAUDE.md §14 "Background Agent Write Restriction" *directive sentence* ("Use `run_in_background: false` for agents that modify files") is retained; only the "auto-deny" mechanism descriptor within that bullet is adjusted.

### Out of Scope — other CC 2.1.186 delta items (NO-OP / informational)

- T2-1 `teammateMode: "iterm2"` (a CC-native user-select value MoAI does not record) — research-classified as intentional non-applicability; not a drift.
- T2-2 `CLAUDE_CODE_MAX_RETRIES` / `CLAUDE_CODE_RETRY_WATCHDOG` env vars — verified absent from the tree (0 grep matches); MoAI intentionally does not set them.
- All other 2.1.186 changelog items (Agent enforcement fix, effort inheritance, schema-loop abort, MCP login/logout, MEMORY auto-compact reminder, `/review` unification, UX bug fixes) — NO-OP / positive / informational per the research report. Not in scope.

### Out of Scope — code, tests, configuration logic

- No Go code change (`internal/**`, `cmd/**`, `pkg/**`) beyond the mechanical `make build` regeneration of `embedded.go`.
- No new lint rule, no new test logic, no configuration-value change. This is a doc-prose correction with mirror regeneration.

### Out of Scope — mirror "3-phase close" drift in spec-frontmatter-schema.md

- The template-mirror `spec-frontmatter-schema.md` still references "4-phase close" while the live copy references "3-phase close" (per SPEC-V3R6-LIFECYCLE-REDESIGN-001). This is a separate, pre-existing mirror drift unrelated to the background-permission rationale. It is observed but NOT corrected here — addressing it would be a scope violation against the 6 enumerated target surfaces.

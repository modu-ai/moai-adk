---
id: SPEC-TOKEN-VERIFY-DIET-001
title: "Verification Output Diet — file-redirect contract for trust-but-verify batch and §E self-verification"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/core/agent-common-protocol.md"
lifecycle: spec-anchored
tags: "token-economy, verification, vci, doctrine, agent-common-protocol"
---

# SPEC-TOKEN-VERIFY-DIET-001 — Verification Output Diet

**Epic**: Token-Economy Epic (4-SPEC A→B→C→D). This SPEC is **C of 4**.

- A = SPEC-TOKEN-ACCOUNTING-001 (closed) — per-SPEC token accounting
- B = SPEC-TOKEN-ROUTING-001 (closed) — Tier×Phase declarative model/effort routing
- **C = this SPEC** — verification output economy (doctrine-primary)
- D = SPEC-TOKEN-BUDGET-STOP-001 (deferred — planned future SPEC, not yet authored at time of writing) — budget hard-stop + graceful abort

**Tier**: M (standard) — see plan.md §A.4 for evidence.

**era**: V3R6 (modern 3-phase close: plan→run→sync).

---

## User Story

**As a** MoAI orchestrator running trust-but-verify batches and a manager agent emitting §E self-verification,
**I want** verbatim tool output redirected to a citable file on disk with only exit-code + bounded-tail summary carried in conversation context,
**so that** verification evidence no longer burns large tool outputs into context TWICE (once when the tool runs inline, once when the result is re-quoted in a Verification Matrix or §E block) — while preserving `verification-claim-integrity.md` §1.1 (the direct-observation obligation).

---

## Problem — Measurable Gap Definition (vci §2 attribution)

Per `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 3, a defect claim must cite a measured source. The four gaps below each name the measured file, the measured line range, and the observed pattern.

### GAP-1 — Canonical 7-item batch has no file-redirect contract

- **Measured source**: `.claude/rules/moai/core/agent-common-protocol.md` lines 270-342 (§ Parallel Execution / Read-only verification batching / Canonical 7-item example).
- **Observed pattern**: The 7 verification commands (lines 290-311) are emitted as bare shell invocations with no output redirection. Each command's full stdout lands in the conversation turn verbatim. The section defines a *parallel-execution* obligation (single-turn multi-Bash) but no *economy* obligation (redirect verbatim → disk; surface exit-code + bounded tail).
- **Double-burn**: Each command's output enters orchestrator context once when the Bash tool runs inline, then enters context again when the Verification Matrix banner re-quotes the result as inline table content (see GAP-3).

### GAP-2 — verification-batch-pattern.md explicitly disclaims ownership of the contract surface

- **Measured source**: `.claude/rules/moai/workflow/verification-batch-pattern.md` line 30 (Re-sync sentinel): verbatim — *"This file owns only the why (grouping rationale + class taxonomy + anti-patterns), not the what (the verbatim command list)."*
- **Observed pattern**: The file-redirect contract currently has no home in this file — the file's own invariant disclaims ownership of the command representation. A standardization SPEC must explicitly extend the Re-sync sentinel to cover the redirect contract too, or the contract drifts the moment a future edit touches the 7-item list.

### GAP-3 — moai.md §8 Verification Matrix / Completion Report banners re-quote verbatim evidence inline

- **Measured source**: `.claude/output-styles/moai/moai.md` line 396 (§ Verification Matrix [HARD]); line 407 (banner template `🤖 MoAI ★ Verification Matrix`); line 603 (§ Completion Report). Both surfaces are binding per `verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report).
- **Observed pattern**: Banner rows display verification results as inline row content (lines 408-411: `✓ V1 [criterion]   ✓ V2 [criterion]`). When a single verification command emits hundreds or thousands of lines (e.g. `go test ./...` with verbose failures, `golangci-lint run` with many findings), the banner expansion amplifies the double-burn — the inline criterion text is itself a re-quote of the Bash output that already entered context once.

### GAP-4 — Originating incident: CHANGELOG full-read autocompact thrashing

- **Measured source (current)**: `wc -l CHANGELOG.md` → 7764 lines (measured 2026-07-08 by this agent via Bash).
- **Recorded incident**: project memory `project_token_economy_epic_handoff.md` §C (4-gap definition bullet for C) cites the CHANGELOG full-read autocompact thrashing as the originating observation for this SPEC. The pattern: a single large file read enters context, the orchestrator then re-quotes portions of it in subsequent turns (Verification Matrix / §E), and the cumulative load trips the auto-compact threshold repeatedly — exactly the double-burn pattern GAP-1 + GAP-3 describe structurally.

### Aggregate defect claim

The aggregate defect is: **no MoAI surface currently standardizes a "verbatim evidence → file on disk with citable path; context → exit code + bounded tail" contract**. Three always-loaded / cross-referenced surfaces (agent-common-protocol.md, verification-batch-pattern.md, moai.md §8) each touch verification output but none owns the contract. This SPEC standardizes it while preserving `verification-claim-integrity.md` §1.1 surface 1+2 (the direct-observation obligation).

---

## Requirements (GEARS notation)

> **Subject convention**: GEARS (current notation per `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format) generalizes the subject beyond "the system". The requirements below use "the orchestrator" / "the agent" / "the banner" / "the file-redirect contract" as appropriate generalized subjects. Each requirement is a discrete, testable assertion. No legacy `IF/THEN` modality is used.

### REQ-001 — Ubiquitous (subject: orchestrator)

The orchestrator SHALL cite a verifiable file path for every verbatim tool output it produces in a Verification Matrix or Completion Report banner, instead of re-quoting the verbatim content inline.

> **Test (AC-VD-001)**: banner rows contain a file-path reference to an on-disk artifact; the verbatim content is NOT embedded as inline row text.

### REQ-002 — Ubiquitous (subject: agent)

Where a verification command's verbatim output exceeds the bounded-tail ceiling, the agent SHALL redirect the verbatim output to a file on disk and surface only exit-code + bounded-tail summary in conversation context.

> **Test (AC-VD-002)**: the agent's response turn carries exit code + tail summary; the verbatim output is reachable at the cited file path.

### REQ-003 — State-driven (While)

**While** a verification command's verbatim output exceeds the bounded-tail ceiling, the orchestrator SHALL redirect the verbatim output to a file on disk prior to rendering any Verification Matrix or Completion Report row that cites that command.

> **Test (AC-VD-002)**: for every banner row whose underlying command exceeded the ceiling, a file path is present and resolves to a file containing the command + full verbatim output.

### REQ-004 — Capability gate (Where)

**Where** the Verification Matrix or Completion Report banner is rendered, the banner SHALL display a file-path column (or row field) referencing the redirected verbatim evidence rather than embedding verbatim content as inline row text.

> **Test (AC-VD-004)**: banner template carries a file-path slot; rendered rows use it.

### REQ-005 — Preservation (subject: file-redirect contract)

The file-redirect contract SHALL preserve `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) and surface 2 (manager §E self-verification) — every claim row in a Verification Matrix or §E self-verification block MUST remain attributable to a directly-observed command whose verbatim output is reachable at the cited file path. **The contract is "verbatim evidence lives on disk with a citable path; context carries the exit code + a bounded tail" — NOT "drop the evidence".**

> **Test (AC-VD-003)**: the contract section explicitly names `verification-claim-integrity` and the §1.1 surface 1+2 preservation obligation, and explicitly rejects the "drop the evidence" interpretation.

### REQ-006 — Anti-regression (When)

**When** the file-redirect contract is applied to the canonical 7-item batch in `agent-common-protocol.md` § Parallel Execution, the orchestrator SHALL NOT weaken or remove the parallel-execution obligation — all 7 verification commands still issue in a single assistant turn. The contract alters output representation (redirect + tail vs inline verbatim), not the single-turn multi-Bash obligation.

> **Test (AC-VD-005, AC-VD-006)**: the 7 verification keywords (`go test`, `coverprofile`, `grep`, `sentinel`, `cmd/moai`, `bench`, `lint`) remain grep-able in the section; the opening HARD clause at line 272 remains intact.

### REQ-007 — Out-of-scope boundary (subject: file-redirect contract)

The file-redirect contract SHALL NOT alter the E1-E7 row structure of `manager-develop`'s §E self-verification matrix. The contract alters only the **evidence-surfacing representation** (file-redirect + bounded tail vs inline verbatim); the E1-E7 row semantics (test / coverage / boundary / sentinel / CLI / bench / lint) remain unchanged.

> **Test (AC-VD-008)**: the run-phase diff against `.claude/agents/moai/manager-develop.md` contains no E1-E7 row additions or deletions; only evidence-surfacing language changes.

---

## Constraints

1. **vci preservation (HARD)** — REQ-005. The diet MUST NOT weaken `verification-claim-integrity.md` §1.1 surface 1+2. The verbatim output must remain ACCESSIBLE (cited file path in the banner/report), just not DUPLICATED inline.
2. **Parallel-execution non-regression (HARD)** — REQ-006. The single-turn multi-Bash obligation is preserved.
3. **E1-E7 boundary (HARD)** — REQ-007. The manager-develop §E matrix STRUCTURAL rewrite is explicitly out of scope.
4. **Doctrine-primary** — This SPEC touches rule files only. No Go code in `internal/runtime/`, `internal/hook/`, or anywhere else. Run-phase edits are `.claude/rules/...` + `.claude/output-styles/...` only.
5. **Template-First Rule** — The 3 edited files (`agent-common-protocol.md`, `verification-batch-pattern.md`, `moai.md`) live in BOTH LIVE and template trees (template-tree existence verified 2026-07-08: all 3 mirrors present under `internal/template/templates/`). Per CLAUDE.local.md §2 [HARD] Template-First Rule, run-phase MUST apply edits to template source first and rebuild via `make build`, OR identically in both trees.
6. **GEARS notation** — Requirements use current GEARS notation. No legacy `IF/THEN` modality.

---

## Scope Decision — Why Doctrine-Primary

The gap is a **contract gap**, not a **mechanism gap**. No new Go code is needed to enforce the contract at the doctrine level — the contract is a documentation-level obligation that orchestrator + manager agents already follow by reading the rule files. The Token-Economy Epic memory (`project_token_economy_epic_handoff.md` §C) classifies this SPEC verbatim as *"독트론 위주"* (doctrine-centric). The contract's mechanical enforcement layer (a future hook that redirects Bash output to disk and validates the cited path) is a separate follow-up SPEC, explicitly out of scope here (see Out of Scope — Mechanical enforcement hook).

---

## Out of Scope

> Per `.claude/rules/moai/development/spec-frontmatter-schema.md` `OutOfScopeRule`, this section uses `### Out of Scope — <topic>` H3 sub-headings with `-` bullets.

### Out of Scope — D (budget hard-stop)

- `SPEC-TOKEN-BUDGET-STOP-001` (deferred next SPEC in Epic) — graceful abort + paste-ready handoff when runtime token budget is exhausted. The file-redirect contract reduces context pressure but does NOT itself hard-fail on budget exhaustion; that is D's domain.
- Any Go code in `internal/runtime/budget.go` adding a hard-fail path — D's EXTEND base per Epic memory §D.

### Out of Scope — E1-E7 structural rewrite

- `manager-develop`'s §E self-verification matrix row-structure rewrite (E1-E7 → any other shape). REQ-007 binds: only the evidence-surfacing representation changes.
- Adding or removing E-rows. The 7 rows (test / coverage / boundary / sentinel / CLI / bench / lint) remain the canonical set.

### Out of Scope — Token accounting measurement

- Per-SPEC token spend measurement (`SPEC-TOKEN-ACCOUNTING-001`'s domain — A). This SPEC reduces double-burn; it does not measure the reduction.
- progress.md `## §I Token Accounting` section authoring — owned by the token-accounting mechanism at sync-close (manager-docs invokes the writer per the Section Map in `spec-frontmatter-schema.md`).

### Out of Scope — Model/effort routing

- Tier×Phase declarative model/effort routing matrix (`SPEC-TOKEN-ROUTING-001`'s domain — B). The file-redirect contract is orthogonal to which model runs the verification.

### Out of Scope — Mechanical enforcement hook

- A future hook that mechanically redirects Bash output to disk and validates the cited path resolves (file-redirect contract's enforcement layer). This SPEC authors the doctrine contract; the enforcement hook is a separate follow-up SPEC.
- Any change to `.claude/hooks/moai/*.sh` — out of scope.

---

## Cross-References

- **Preserved invariant**: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) + surface 2 (manager §E self-verification). REQ-005 binds this SPEC to preserve it.
- **Primary change target**: `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution / Read-only verification batching / Canonical 7-item example (line 270+).
- **Cross-ref adoption targets**: `.claude/rules/moai/workflow/verification-batch-pattern.md` (Re-sync sentinel at line 30); `.claude/output-styles/moai/moai.md` §8 (Verification Matrix at line 396, Completion Report at line 603).
- **Epic context**: project memory `project_token_economy_epic_handoff.md` §C (gap definition + originating incident).
- **Sibling devices (cross-ref only, NOT this SPEC's targets)**: `internal/config/token_budget_guard.go` (always-loaded 75K tripwire, SPEC-TOKEN-EFFICIENCY-001); `internal/runtime/cache_control.go` (prompt cache_control injection).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-08 | manager-spec | Initial draft — plan-phase artifacts (spec + plan + acceptance + progress). Token-Economy Epic C of 4. Doctrine-primary; Tier M. |

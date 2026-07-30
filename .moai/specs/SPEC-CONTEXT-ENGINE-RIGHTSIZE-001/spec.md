---
id: SPEC-CONTEXT-ENGINE-RIGHTSIZE-001
title: "Context Engine Rightsizing — Anthropic 6-Principle Alignment (Conservative B-Group Only)"
version: "0.1.1"
status: completed
created: 2026-07-28
updated: 2026-07-30
author: manager-spec
priority: P2
phase: "v3.x"
module: ".claude/rules/moai"
lifecycle: spec-anchored
tags: "context-engine, anthropic-alignment, rules-evolution, conservative"
tier: M
related_specs:
  - SPEC-CONTEXT-INJ-001
---

# SPEC-CONTEXT-ENGINE-RIGHTSIZE-001 — Context Engine Rightsizing

## HISTORY

- **2026-07-28 v0.1.0** — Initial plan-phase authoring. Baseline verified via direct grep (not handoff text): `[ZONE:Frozen]` x 66 / `[ZONE:Evolvable]` x 98 across 13 files; `CLAUDE.md` `[HARD]` x 15. Scope set to GOOS decision "보수(B-group only)" — 2 expressive transitions (code_comments, Tool Selection absolute form) + 1 dedup consolidation (M1 Tool Selection Priority). C-group (Multi-File Decomposition, Reproduction-First Bug Fix) explicitly preserved as mechanical guardrails. A-group (Frozen 66 + 5 anchor doctrines) strictly off-limits.

---

## §A. User Story & Background

### §A.1 User Story

**As** the MoAI-ADK maintainer,
**I want** the `.claude/rules/moai/` and `CLAUDE.md` instruction corpus to align with Anthropic's 2026-07-24 "new rules of context engineering" guidance (Thariq @trq212: "We removed ~80% of the Claude Code system prompt for our newest models, with no measurable loss"),
**so that** absolute/prohibitive rule language is replaced with judgement-delegating informational language **where the rule is stylistic convention** (B-group), while preserving absolute language **where the rule is a mechanical guardrail or safety invariant** (A-group Frozen + C-group scope-discipline rules).

### §A.2 Background — Anthropic 6-Principle Alignment Status (2026-07-28实测)

Anthropic's 2026-07-24 guidance establishes six principles for Claude 5-generation context engineering. moai-adk's measured status:

| Principle | moai-adk status | This SPEC scope |
|---|---|---|
| Progressive disclosure | 🟢 Already leading (skill frontmatter + on-demand `Skill()`) | Out of scope |
| Auto-memory | 🟢 Already leading (`~/.claude/projects/{hash}/memory/` + `MEMORY.md`) | Out of scope |
| Rich references | 🟢 Already leading (cross-reference doctrine, SSOT discipline) | Out of scope |
| `/doctor` rightsizing | 🟢 Already leading (harness 3-tier + `/moai doctor`) | Out of scope |
| **Simple tool descriptions (M1)** | 🟡 Duplicate "Tool Selection Priority" block across `moai-constitution.md` + `agent-common-protocol.md` | **In scope** |
| **Rules → judgement (M2)** | 🔴 Largest gap (HARD 15 in `CLAUDE.md`, MUST 269 / NEVER 14 in `.claude/rules/moai/`) | **In scope (conservative B-group only)** |

### §A.3 Conservative Scope Decision (GOOS 2026-07-28)

GOOS decision: **"보수(B-group only)"** — transition only the clearly-stylistic absolute rules to judgement-delegating language; preserve all mechanical guardrails and safety invariants untouched.

- **B-group (transition targets)**: 2 stylistic rules — code-comments language expression + Tool Selection absolute "Use X instead of Y" form.
- **C-group (preserved as mechanical guardrails)**: Multi-File Decomposition + Reproduction-First Bug Fix — these prevent orchestrator scope-creep and bug-fix-quality regressions; transitioning them trades a real guardrail for token-perfume.
- **A-group (strictly off-limits)**: All `[ZONE:Frozen]` 66 markers + 5 anchor doctrines (AskUserQuestion-Only, Deferred Tool Preload, Subagent Prohibitions, Branch Guard, Verification-Claim Integrity, Native `/goal` Prohibition).

---

## §B. Scope

### §B.1 In Scope — Constructive Edits

| ID | Target | Location (verified 2026-07-28) | Change class |
|---|---|---|---|
| M1.1 | "Tool Selection Priority" 5-bullet dedup | `moai-constitution.md` § Tool Selection Priority (lines ~106-117) | Collapse to 1-line SSOT pointer |
| M1.1 | Canonical SSOT retained | `agent-common-protocol.md` § Tool Selection by Task (line ~231, detailed table) | Unchanged — already canonical |
| M1.2 | "Use Grep instead of grep/rg" duplication | `moai-constitution.md` line ~116 vs `agent-common-protocol.md` line ~209 | Resolve via M1.1 collapse |
| M2.1 | "English comments" TRUST 5 Readable line | `moai-constitution.md` line ~77 | Replace with config-respecting judgement language |
| M2.2 | Tool Selection absolute form | `moai-constitution.md` (same block as M1.1) | "Use X instead of Y" → "prefer the dedicated tool" (combined with M1.1) |

### §B.2 In Scope — Verification-Only (No Edit Required)

| ID | Target | Verified state | Reason |
|---|---|---|---|
| M1.3 | `plan-auditor.md` tool guidance | Line ~144 already delegates to `agent-common-protocol.md § Tool Selection by Task` SSOT | **No defect exists.** SSOT reference already in place. Verification-only AC — falsifiable grep that the reference survives. |

### §B.3 In Scope — Preservation Guards (A-group + C-group)

| Group | Items (verified baseline) | Guard class |
|---|---|---|
| A-group Frozen | `[ZONE:Frozen]` x 66 across 13 files | Strictly preserved — no transition, no removal, no `[ZONE:Evolvable]` mutation |
| A-group AskUserQuestion | `AskUserQuestion-Only` HARD + Subagent Prohibitions + Deferred Tool Preload (`ToolSearch(query:"select:AskUserQuestion")`) | Strictly preserved |
| A-group Branch Guard | `main-checkout-branch-guard.md` doctrine + mechanical enforcer | Strictly preserved |
| A-group Verification-Claim | `verification-claim-integrity.md` §1.1 surfaces 1-3 | Strictly preserved |
| A-group Native /goal | `goal-directive.md` § Native `/goal` Prohibition | Strictly preserved |
| C-group Multi-File Decomposition | `CLAUDE.md:19` HARD bullet + `CLAUDE.md:151` §7 Rule 2 | Preserved as mechanical guardrail (C-group, NOT transitioned) |
| C-group Reproduction-First Bug Fix | `CLAUDE.md:19` HARD bullet + `CLAUDE.md:157` §7 Rule 4 | Preserved as mechanical guardrail (C-group, NOT transitioned) |

---

## §C. Requirements (GEARS Notation)

### §C.1 Constructive Requirements

**REQ-CER-001** (M1.1 — Tool Selection SSOT consolidation, Ubiquitous)
The `moai-constitution.md` § Tool Selection Priority block (5-bullet "Use X instead of Y" list, lines ~108-117) shall be replaced with a single-line informational pointer to the canonical `agent-common-protocol.md § Tool Selection by Task` table, preserving tool-selection guidance as a single source of truth.

**REQ-CER-002** (M1.1 — Canonical SSOT retained, Ubiquitous)
The `agent-common-protocol.md § Tool Selection by Task` table shall remain the canonical detailed reference for tool-selection guidance, unchanged in semantic intent (prefer dedicated tools over general alternatives for accuracy and efficiency).

**REQ-CER-003** (M2.1 — code_comments expressive transition, Ubiquitous)
**Where** a rule's prose prescribes the language of code comments, the rule shall defer to the project's `code_comments` configuration (`.moai/config/sections/language.yaml`) and frame the prescription as "match the surrounding code's comment language and density" rather than an unconditional "English comments" absolute.

**REQ-CER-004** (M2.2 — Tool Selection informational reframing, Ubiquitous)
**Where** the consolidated Tool Selection pointer in `moai-constitution.md` describes tool choice, it shall use judgement-delegating language ("prefer the dedicated tool", "fit-for-purpose") rather than prohibitive absolutes ("Use X instead of Y", "Never use Bash grep").

### §C.2 Preservation Requirements (A-group + C-group)

**REQ-CER-005** (A-group Frozen count preservation, Ubiquitous)
After this SPEC's run-phase, the count of `[ZONE:Frozen]` occurrences in `.claude/rules/moai/` shall be greater than or equal to 66 (the 2026-07-28 verified baseline), ensuring no Frozen marker was removed or downgraded to Evolvable as a side effect of M1/M2 edits.

**REQ-CER-006** (A-group AskUserQuestion doctrines preservation, Ubiquitous)
After this SPEC's run-phase, the AskUserQuestion-Only Interaction HARD rule, the Subagent Prohibitions (no `AskUserQuestion` from subagents), and the Deferred Tool Preload (`ToolSearch(query:"select:AskUserQuestion")`) shall remain observable in `.claude/rules/moai/core/askuser-protocol.md` and `.claude/rules/moai/core/agent-common-protocol.md`.

**REQ-CER-007** (A-group safety invariants preservation, Ubiquitous)
After this SPEC's run-phase, the Branch Guard doctrine (`main-checkout-branch-guard.md`), the Verification-Claim Integrity invariant (`verification-claim-integrity.md` §1.1 surfaces 1-3), and the Native `/goal` Prohibition (`goal-directive.md` § Native `/goal` Prohibition) shall remain observable at their canonical paths.

**REQ-CER-008** (C-group mechanical guardrails preservation, Ubiquitous)
The Multi-File Change Decomposition rule (`CLAUDE.md §7 Rule 2`) and the Reproduction-First Bug Fix rule (`CLAUDE.md §7 Rule 4`), including their `[HARD]` bullets at `CLAUDE.md:19`, shall remain in force as mechanical guardrails — this SPEC shall NOT transition them to judgement-delegating language, on the rationale that they prevent orchestrator scope-creep and bug-fix-quality regression (C-group classification, per `feedback_guard_signal_proves_call_not_effect`).

### §C.3 Operational Requirements

**REQ-CER-009** (M1.3 — plan-auditor SSOT reference verified, State-driven)
**While** the M1.1 consolidation is in effect, the `plan-auditor.md` § Verification Execution Mandate (line ~144) shall continue to delegate tool-selection guidance to `agent-common-protocol.md § Tool Selection by Task` via an observable SSOT cross-reference (the SSOT reference existed pre-edit; this REQ guards against regression).

**REQ-CER-010** (Template mirror synchronization, Ubiquitous)
**Where** this SPEC edits `.claude/rules/moai/core/moai-constitution.md` or `.claude/rules/moai/core/agent-common-protocol.md`, the corresponding `internal/template/templates/.claude/rules/moai/core/` mirror shall be synchronized AND neutralized per CLAUDE.local.md §2 [HARD] Template-First Rule + §25 (no SPEC IDs, no REQ tokens, no audit citations — generic prose only).

**REQ-CER-011** (No behavioral regression, Event-detected)
**When** this SPEC's run-phase completes, `moai spec lint` shall report no new findings attributable to the M1/M2 edits (the 2026-07-28 repo-wide baseline excludes pre-existing unrelated debt), and `go test ./...` shall remain green (no Go code is touched by this SPEC; a test regression would indicate an unrelated race).

---

## §D. Non-Functional Requirements (A-Group Preservation)

The A-group (Frozen) preservation AC below is duplicated here as a non-functional requirement because it is the single most important guardrail of this SPEC — the "보수(B-group only)" decision collapses entirely if any Frozen marker is silently downgraded. The acceptance matrix (§D in `acceptance.md`) carries the falsifiable grep AC; this section records the normative intent.

- **NFR-CER-001**: No `[ZONE:Frozen]` marker in `.claude/rules/moai/` shall be removed, rewritten, or mutated to `[ZONE:Evolvable]` as a side effect of M1/M2 edits.
- **NFR-CER-002**: No A-group anchor doctrine (AskUserQuestion-Only, Subagent Prohibitions, Deferred Tool Preload, Branch Guard, Verification-Claim Integrity, Native `/goal` Prohibition) shall have its semantic intent weakened.
- **NFR-CER-003**: Template mirrors (`internal/template/templates/.claude/rules/moai/`) shall be neutralized per §25 — no SPEC ID, no REQ token (`REQ-CER-*`), no audit citation.

---

## §E. Constraints

1. **Tier M** — 3 artifacts (spec.md + plan.md + acceptance.md). No design.md / research.md required: GOOS already made the B-group-only scope decision; no architectural trade-off remains open.
2. **No Go code changes** — this SPEC edits `.claude/rules/moai/*.md` and template mirrors only.
3. **`make build` not required** — `.claude/rules/` is not embedded via `//go:embed` (only `internal/template/templates/` is, per CLAUDE.local.md §2). Template edits do not require recompile.
4. **Template-first obligation** (CLAUDE.local.md §2 [HARD]) — every edit to `.claude/rules/moai/` MUST be mirrored to `internal/template/templates/.claude/rules/moai/` with §25 neutralization applied.
5. **Conservative scope** — C-group items (Multi-File Decomposition, Reproduction-First Bug Fix) MUST NOT be transitioned. Any run-phase proposal to transition them MUST return a blocker report and be re-delegated to manager-spec per the D-NEW-1 inline-fix pattern.

---

## §F. Out of Scope

### Out of Scope — A-group Frozen transitions

- Transitioning any `[ZONE:Frozen]` rule to `[ZONE:Evolvable]`. The 66 Frozen markers are the explicitly off-limits A-group per the GOOS "보수" decision.
- Removing or downgrading any of the 5 anchor doctrines (AskUserQuestion-Only, Subagent Prohibitions, Deferred Tool Preload, Branch Guard, Verification-Claim Integrity, Native `/goal` Prohibition).

### Out of Scope — C-group mechanical guardrails

- Transitioning Multi-File Change Decomposition (`CLAUDE.md §7 Rule 2`) to judgement-delegating language. This is a C-group mechanical guardrail, preserved.
- Transitioning Reproduction-First Bug Fix (`CLAUDE.md §7 Rule 4`) to judgement-delegating language. Same rationale.
- Rationale: both rules encode bias-prevention guardrails (scope-creep / confirmation-bias) whose mechanical enforcement survives even when stylistic rules move to judgement-delegation. See `feedback_guard_signal_proves_call_not_effect` and `feedback_guard_observation_must_be_falsifiable`.

### Out of Scope — Anthropic principles already leading

- Progressive disclosure changes — already leading (3-tier token budget + on-demand `Skill()`).
- Auto-memory architecture changes — already leading (`MEMORY.md` + topic files + `feedback_*.md` convention).
- Rich references / cross-reference doctrine — already leading.
- `/doctor` rightsizing — already leading (harness 3-tier auto-determination).

### Out of Scope — M1.3 plan-auditor.md edits

- `plan-auditor.md` tool-guidance edits — verified 2026-07-28 that line ~144 already delegates to the `agent-common-protocol.md § Tool Selection by Task` SSOT. **No defect exists; no edit required.** A falsifiable AC in `acceptance.md` guards against regression of the existing SSOT reference. Per `feedback_claimed_correction_never_applied`, no ghost REQ is created for a non-existent defect.

### Out of Scope — CLAUDE.md Tool Selection

- Adding a Tool Selection section to `CLAUDE.md`. Confirmed 2026-07-28 via direct grep that `CLAUDE.md` has no such section. The M1 dedup is between `moai-constitution.md` and `agent-common-protocol.md` only.

### Out of Scope — Token-budget sweep / global MUST/NEVER reduction

- A repo-wide sweep reducing the 269 `MUST` / 14 `NEVER` count in `.claude/rules/moai/`. This would be a separate SPEC; the current scope limits transitions to the 2 B-group targets (code_comments expression + Tool Selection absolute form) plus the M1 dedup consolidation.

### Out of Scope — docs-site documentation sync

- Updates to `docs-site` (`adk.mo.ai.kr`) 4-locale documentation. The 4-locale sync is a separate doctrine (CLAUDE.local.md §17) and is not triggered by `.claude/rules/moai/` prose edits.

---

## §G. Cross-References

- **Anthropic source**: "The new rules of context engineering for Claude 5 generation models" (2026-07-24; Thariq @trq212 announcement: ~80% Claude Code system-prompt removal with no measurable loss).
- **Memory handoff (1st-pass analysis)**: `~/.claude/projects/-Users-goos-MoAI-moai-adk-go/memory/project_context_engine_rightsize_001_handoff.md`.
- **Frontmatter schema SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md` § Canonical 12 Required Fields.
- **Template-First Rule**: CLAUDE.local.md §2 [HARD].
- **Template Internal-Content Isolation**: CLAUDE.local.md §25 (also `.moai/docs/template-internal-isolation-doctrine.md`).
- **Applied lessons**:
  - `feedback_claimed_correction_never_applied` — every "fixed" claim verified via direct grep at edit time.
  - `feedback_defect_claim_verification` — M1.3 classified as no-defect after verification.
  - `feedback_local_template_sync_neutralize_first` — template mirror neutralized, not blind-copied.
  - `feedback_guard_observation_must_be_falsifiable` — A-group ACs use grep counts, not prose claims.
  - `feedback_guard_signal_proves_call_not_effect` — C-group preservation rationale (mechanical guardrail survives judgement-delegation of stylistic rules).
- **Related SPEC**: `SPEC-CONTEXT-INJ-001` (archived, Memory Persistence — different scope; included for dedup traceability).

---

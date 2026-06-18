---
id: SPEC-V3R6-RULES-SSOT-DEDUP-001
title: "Rules SSOT De-duplication + Structural Consolidation"
version: "0.1.0"
status: in-progress
created: 2026-06-19
updated: 2026-06-19
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai + internal/template/templates/.claude/rules/moai"
lifecycle: spec-anchored
tags: "rules, ssot, dedup, consolidation, template-mirror"
era: V3R6
tier: L
---

# SPEC-V3R6-RULES-SSOT-DEDUP-001

## §A. Problem Statement

A rules-improvement audit over `.claude/rules/moai/` found **18 files self-declaring "SSOT"
(single source of truth) while verbatim-copying content they claim to own elsewhere.** The
copies have already drifted: in several cases a "copy" contradicts the SSOT owner it mirrors
(timeout 5s vs 10s, `alwaysLoad` v2.1.119 vs v2.1.121, `max_teammates` default 10 vs the 3-5
ceiling everywhere else). Duplicated HARD clauses are the single largest source of doctrine
drift: each restatement is an independent edit surface that can fall out of sync silently.

This SPEC owns **SSOT de-duplication and structural file consolidation** for that audit. It is
SPEC 3 of a 4-SPEC Sprint 16 "rules-improvement" cohort. Each in-scope de-duplication target
designates ONE SSOT owner and reduces every non-owner copy to a cross-reference pointer; the
acceptance criteria grep-verify (a) the duplicated block is gone from the non-owner, (b) a
pointer remains, and (c) the SSOT owner is intact. One target is a structural file merge
(team files 3 → 2) that deletes a file after folding its content into the retained owner.

### §A.1 Cohort boundary (what other cohort SPECs own — NOT this SPEC)

| Sibling SPEC | Owns | This SPEC does NOT touch |
|--------------|------|--------------------------|
| HOTFIX-001 | Low-risk token fixes | one-off typos / token corrections |
| CATALOG-SCRUB-001 | Archived-agent reference scrub | archived-agent name removals (except the one file THIS SPEC deletes — see §F dependency) |
| VERSION-FORMAT-001 | Version / format staleness | `Version:` footers, format-only edits |

This SPEC owns structural consolidation (the team-file merge) because consolidation is a
structural SSOT decision, not a token/format fix. The boundary is: **THIS SPEC merges/points;
the siblings scrub/correct.** Where a target's scope brushes a sibling's (the deleted
`agent-teams-pattern.md` carries a dead archived-agent path that CATALOG-SCRUB-001 would
otherwise fix), §F records the explicit dependency.

## §B. Background — Verified Repo Facts

These facts were verified during plan-phase research and MUST be encoded correctly (the stale
"embedded.go" premise is explicitly rejected):

- **B1 — Embed mechanism.** `internal/template/embedded.go` does NOT exist. Templates embed via
  `//go:embed all:templates` at `internal/template/embed.go:28`. `make build` runs
  `templ-generate` + `gen-catalog-hashes.go --all` + `go build`. There is no generated
  `embedded.go` golden file to regenerate; edits land directly in `internal/template/templates/`.
- **B2 — Mirror-parity is a Go-test trio, not `diff`.** Parity is enforced by three Go tests:
  - `TestRuleTemplateMirrorDrift` (`internal/template/rule_template_mirror_test.go`) — byte-parity
    for files in the `workflowOptMirroredPaths` allowlist.
  - `TestTemplateNoInternalContentLeak` (`internal/template/internal_content_leak_test.go`) —
    enforces that §25-sanitized mirrors carry no internal-development content (SPEC-IDs, REQ/AC
    tokens). These files are INTENTIONALLY divergent deployed↔template.
  - `TestTemplateNeutralityAudit` (`internal/template/template_neutrality_audit_test.go`) —
    16-language neutrality.
- **B3 — Per-file mirror semantics (verified).** Of the 20 files in scope:
  - `hooks-system.md` is in the **byte-parity allowlist** → its mirror MUST be edited byte-identically.
  - `agent-common-protocol.md` (T1, T4), `verification-batch-pattern.md` (T4), and
    `agent-teams-pattern.md` (T6, being deleted) are **§25-sanitized** → NOT byte-parity; their
    mirror cleanliness is enforced by `TestTemplateNoInternalContentLeak`. ACs MUST use leak-test
    semantics, never naive `diff`.
  - `lifecycle-sync-gate.md` (T5) has **NO template mirror** (deployed-only) → its de-dup edit
    requires NO mirror edit.
  - The remaining files have a byte-parity mirror requiring the same edit.
- **B4 — zone-registry `clause:` is CLI-load-bearing (decisive finding).** The `clause:` field
  of every zone-registry entry is consumed by the `moai constitution` CLI in three ways
  (see research.md for the full trace):
  - `moai constitution list` renders `Clause` in both table and JSON output.
  - `moai constitution amend --before <text>` does an **exact-match** against `rule.Clause`.
  - `moai constitution validate` requires the **normalized clause text to appear as a substring
    in the source file** (`internal/constitution/validator.go:264`); a non-verbatim 3-5 word
    label would emit a `SentinelDrift` finding for every reduced entry → breaks `validate`.
  - `Rule.Validate()` (`internal/constitution/rule.go:69`) only requires `Clause` non-empty.
  Consequence: the originally-proposed "reduce each clause to a 3-5 word label (NOT verbatim
  source text)" is UNSAFE as-is — it breaks `validate` drift detection. The zone-registry
  milestone is therefore scoped to a SHOULD/partial reduction + a blocker note (see §E REQ-SSD-009).

## §C. Goal

The rules tree shall hold exactly one authoritative statement of each duplicated doctrine block,
with every non-owner copy reduced to a cross-reference pointer, and the team-protocol family
consolidated from four overlapping copies down to two SSOT files plus one pointer — without
breaking template-mirror parity (B2), template neutrality (§25), or the `moai constitution` CLI
(B4).

## §D. Requirements (GEARS)

### §D.1 De-duplication requirements (one SSOT owner per block)

**REQ-SSD-001 — AskUserQuestion 4-way de-dup.**
The de-duplication shall designate `core/askuser-protocol.md` as the AskUserQuestion SSOT.
**When** the de-dup is applied, the restated AskUserQuestion enforcement clauses in
`core/agent-common-protocol.md` §User Interaction Boundary and `core/moai-constitution.md`
§MoAI Orchestrator shall be reduced to cross-reference pointers to `core/askuser-protocol.md`,
retaining only each file's locally-unique delta (the subagent-prohibition framing in
agent-common-protocol; the orchestrator-obligation framing in moai-constitution). CLAUDE.md §8
is OUTSIDE `.claude/rules/` and shall NOT be edited by this SPEC (recorded as a forward-item).

**REQ-SSD-002 — Hooks config/timeout de-dup + contradiction reconciliation.**
The de-duplication shall designate `core/hooks-system.md` as the SSOT for hook JSON config +
path-quoting + the timeout table. **When** the de-dup is applied, `core/settings-management.md`
§Hooks Configuration shall be reduced to a pointer plus only the settings-specific delta (the
StatusLine no-env-var rule). The de-duplication shall reconcile the PostToolUse timeout
contradiction so that both files agree: synchronous hooks default to **5s** (MoAI policy per
CLAUDE.local.md §7), and PostToolUse is the documented exception at **10s + `async: true`**
(LSP/AST/MX validations run in background; the 10s is a per-run background ceiling, not a
blocking wait) — both files shall state the same 5s-default + 10s-async-PostToolUse rule with
this rationale. The de-duplication shall also reconcile the `alwaysLoad` version contradiction
(`v2.1.119` vs `v2.1.121` in settings-management.md) to a single canonical value.

**REQ-SSD-003 — Context-window threshold 5-way de-dup.**
The de-duplication shall designate `workflow/context-window-management.md` as the
context-window-threshold SSOT. **When** the de-dup is applied, the threshold table copies in
`core/settings-management.md`, `core/zone-registry.md` (CONST-V3R5-022), and
`development/skill-ab-testing.md` shall be reduced to pointers. `workflow/session-handoff.md`
already correctly defers and shall be left unchanged.

**REQ-SSD-004 — 7-item verification batch de-dup.**
The de-duplication shall designate `core/agent-common-protocol.md` §Parallel Execution (the
verbatim 7-command block) as the verification-batch SSOT. **When** the de-dup is applied,
`workflow/verification-batch-pattern.md` shall retain ONLY the grouping rationale/taxonomy and a
sentinel note instructing re-sync if the 7-item list changes; its copy of the 7-command block
shall be removed.

**REQ-SSD-005 — Status Transition Ownership Matrix de-dup.**
The de-duplication shall designate `development/spec-frontmatter-schema.md` as the Status
Transition Ownership Matrix SSOT. **When** the de-dup is applied, `workflow/lifecycle-sync-gate.md`
shall reduce its full matrix copy (the table near its §Status Transition Ownership Matrix
Cross-Reference) to a pointer, retaining only the close-subject-full-ID one-liner where uniquely
useful.

**REQ-SSD-006 — Skill 3-file SSOT boundary.**
The de-duplication shall fix the skill-doc SSOT boundary: `development/skill-authoring.md` =
frontmatter/schema SSOT; `development/skill-writing-craft.md` = prose-craft SSOT;
`development/skill-ab-testing.md` = A/B-method SSOT. **When** the de-dup is applied, the
duplicated frontmatter-schema table in `skill-writing-craft.md` shall be reduced to a pointer to
`skill-authoring.md`, and the duplicated frontmatter/progressive-disclosure checklist in
`skill-ab-testing.md` shall be reduced to a pointer.

**REQ-SSD-007 — Agent 3-file SSOT boundary.**
The de-duplication shall fix the agent-doc SSOT boundary: `development/agent-patterns.md` =
pattern-taxonomy + per-domain tool-whitelist SSOT; `development/orchestrator-templates.md` =
MoAI-specific application; `development/agent-authoring.md` = frontmatter SSOT. **Where** a
per-domain tool-whitelist or escalation/transition matrix is duplicated across these files, the
de-dup shall pick one owner (per-domain tool-whitelist → `agent-patterns.md`, already referenced
by `agent-authoring.md`) and reduce the other occurrences to pointers.

### §D.2 Structural consolidation requirement (this SPEC owns)

**REQ-SSD-008 — Team files 4 → 2 + 1 pointer consolidation.**
The consolidation shall retain `workflow/team-protocol.md` (mechanics SSOT — Role Matrix,
Mailbox v2, spawn-wrapper) and `workflow/team-pattern-cookbook.md` (pattern compositions).
**When** the consolidation is applied:
- the 5+1+1 composition from `workflow/agent-teams-pattern.md` shall be folded into
  `team-pattern-cookbook.md` as a 6th pattern, AND
- `workflow/agent-teams-pattern.md` shall be deleted (both deployed and template trees), AND
- no inbound reference to `agent-teams-pattern.md` shall dangle after deletion, AND
- `workflow/worktree-integration.md` §Team Protocol (the 4th copy, "merged from
  team-protocol.md" but never deleted) shall be reduced to a cross-reference to
  `team-protocol.md`, AND
- the roster-limit contradiction (`team-protocol.md` `max_teammates` default 10 vs the 3-5
  ceiling stated everywhere else) shall be reconciled to a single consistent statement
  (the mechanical `max_teammates` config default is documentation-reconciled with the
  Anthropic-recommended 3-5 starting ceiling — see design.md §6 for the reconciliation form).

### §D.3 zone-registry requirement (highest risk, CLI-gated)

**REQ-SSD-009 — zone-registry bloat reduction (SHOULD, CLI-gated).**
**Where** the `moai constitution validate` drift check requires the `clause:` text to be a
verbatim source substring (B4), the zone-registry reduction shall NOT blank `clause:` to a
non-verbatim 3-5 word label. The reduction SHOULD instead: (a) shorten each `clause:` to a
**still-verbatim** shorter excerpt of the source HARD clause (reducing bloat while keeping the
`validate` substring check green), and (b) narrow the over-broad `paths:` trigger (currently
`.claude/**,.moai/specs/**,.claude/rules/**`, loading ~13K tokens on every `.moai/specs/**`
edit) to the minimum that preserves intended load behavior. **When** the full structural
reduction (label-only entries) cannot be done without a CLI change, the SPEC shall scope a
partial reduction and record a blocker note recommending a follow-up SPEC to make
`validator.go` label-aware. zone-registry has its OWN milestone (M_Z) gated on the research
finding (B4 / research.md).

### §D.4 Cross-cutting constraints (GEARS — apply to every milestone)

**REQ-SSD-010 — Template-First mirror discipline.**
**When** any `.claude/rules/**` file with a template mirror is edited, the corresponding
`internal/template/templates/.claude/rules/**` mirror shall be edited in the same change, per
the per-file mirror semantics of B3 (byte-identical for byte-parity-allowlist files;
sanitized-equivalent for §25 files). Edits shall NOT touch a non-existent `embedded.go`.

**REQ-SSD-011 — Template neutrality preservation.**
**When** any template-tree file is edited, the edit shall preserve §25 internal-content
neutrality (no SPEC-IDs, REQ/AC tokens, audit citations, or internal dates leak into the
template mirror) per CLAUDE.local.md §25, verified by `TestTemplateNoInternalContentLeak`.

**REQ-SSD-012 — Pointer integrity.**
**When** a duplicated block is reduced to a pointer, the pointer shall name the SSOT owner file
and the relevant section anchor, so a reader reaches the authoritative text in one hop, and no
reference to a deleted file shall dangle.

## §E. Self-Verification (audit-ready)

This section is the plan-phase audit-ready signal. Run-phase evidence is recorded in
`progress.md` §E.2/§E.3; sync/Mx evidence in §E.4/§E.5.

- E-PLAN-1: All 12 REQ-SSD requirements expressed in GEARS form with a designated SSOT owner per block.
- E-PLAN-2: Every de-dup target maps to a grep-AC in acceptance.md (duplicated block absent from
  non-owner + pointer present + SSOT owner intact).
- E-PLAN-3: The file-deletion target (T6) has a 3-clause AC (file gone + content survives in
  cookbook + no dangling referrer).
- E-PLAN-4: The zone-registry CLI-dependency finding (B4) is captured in research.md with the
  exact source lines, and REQ-SSD-009 is scoped SHOULD/partial accordingly.
- E-PLAN-5: design.md carries the SSOT-owner decision table + zone-registry reduction design.
- E-PLAN-6: §F records the `agent-teams-pattern.md` deletion dependency on CATALOG-SCRUB-001.

## §F. Cross-SPEC Dependency — agent-teams-pattern.md deletion vs CATALOG-SCRUB-001

`workflow/agent-teams-pattern.md` carries a **dead frontmatter `paths:` reference** to
`.claude/agents/moai/manager-strategy.md` — an archived agent. CATALOG-SCRUB-001 has a milestone
to fix that dead path. **This SPEC DELETES `agent-teams-pattern.md` entirely** (REQ-SSD-008), so
the deletion **supersedes** CATALOG-SCRUB-001's frontmatter edit for that file.

Recommended resolution (one of):
1. **Execution order CATALOG-SCRUB-001 → SSOT-DEDUP-001** is NOT required and is in fact wasteful
   (CATALOG-SCRUB would fix a path on a file SSOT-DEDUP then deletes). **Preferred: execute
   SSOT-DEDUP-001 first (or independently), and CATALOG-SCRUB-001 drops the agent-teams-pattern.md
   frontmatter milestone** as superseded.
2. If CATALOG-SCRUB-001 runs first, its frontmatter fix is harmless but redundant; SSOT-DEDUP-001's
   deletion is the terminal state.

This dependency MUST be surfaced to the orchestrator at run-phase sequencing so the two SPECs do
not both edit/delete the same file in a race. Recorded here as the canonical statement; the
orchestrator coordinates ordering.

## §G. Out of Scope

This SPEC owns SSOT de-duplication and the team-file structural merge ONLY. The following are
explicitly excluded.

### Out of Scope — archived-agent reference scrubbing
- Removing archived-agent names (`manager-strategy`, `expert-*`, etc.) from rule bodies is
  owned by CATALOG-SCRUB-001 — EXCEPT the dead frontmatter path on `agent-teams-pattern.md`,
  which is resolved by this SPEC's deletion of that file (§F).
- General archived-agent migration-table edits are not in scope here.

### Out of Scope — version / format staleness
- `Version:` footer bumps, date corrections, and pure-formatting normalizations are owned by
  VERSION-FORMAT-001.
- This SPEC does not touch a file solely to update its version footer.

### Out of Scope — CLAUDE.md and CLAUDE.local.md edits
- CLAUDE.md §8 (the 4th AskUserQuestion restatement) is OUTSIDE `.claude/rules/` and is recorded
  as a forward-item only; this SPEC does NOT edit CLAUDE.md.
- CLAUDE.local.md is maintainer-private and not edited here.

### Out of Scope — Go source / CLI behavior changes
- This SPEC does NOT modify `internal/constitution/*.go`, `internal/cli/constitution.go`, or any
  Go file. If the zone-registry full reduction requires a `validator.go` change (B4), it is
  deferred to a follow-up SPEC recorded as a blocker note (REQ-SSD-009).
- The only Go-adjacent artifacts touched are the template-mirror files under
  `internal/template/templates/`, which are markdown, not Go.

### Out of Scope — new SSOT content authoring
- This SPEC does NOT add new doctrine. It relocates existing duplicated text to a single owner
  and replaces copies with pointers. No clause's *meaning* changes (except the explicit
  contradiction reconciliations in REQ-SSD-002 and REQ-SSD-008, which converge two conflicting
  statements onto one).

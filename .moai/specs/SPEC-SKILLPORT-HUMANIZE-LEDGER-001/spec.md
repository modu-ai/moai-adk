---
id: SPEC-SKILLPORT-HUMANIZE-LEDGER-001
title: "Graft Invariant Ledger + Delta Audit into moai-domain-humanize"
version: "0.1.3"
status: in-progress
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: "internal/template/templates/.claude/skills/moai-domain-humanize"
lifecycle: spec-anchored
tier: S
tags: "skill, humanize, invariant-ledger, delta-audit, fact-drift, clean-room, epic-skillport"
---

# SPEC-SKILLPORT-HUMANIZE-LEDGER-001 — humanize enhancement

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-24 | manager-spec | Initial plan-phase draft. Graft two techniques (Invariant Ledger + Delta Audit), functionally inspired by skillstead `writing-quality-editor` (an Apache-2.0 project), into the EXISTING moai-domain-humanize skill as an additive anti-fact-drift layer. |
| 0.1.1 | 2026-07-24 | manager-spec | Plan-phase revision. (1) Clean-room pivot: the Invariant Ledger + Delta Audit are ORIGINAL moai authoring inspired by the general technique — NOT a skillstead port. No skillstead attribution is added (no provenance footer, no NOTICE, no skillstead-tied license); the pre-existing im-not-ai footer line and the pre-existing `license: Apache-2.0` frontmatter field are unrelated to this change and are preserved exactly as-is. Removed the former skillstead-attribution requirement; added the clean-room authoring requirement. (2) plan-audit fixes: deleted both residual clarification-gate markers and baked their decisions into prose (D-common), broadened the version bump to BOTH version surfaces (D3), gave the Template-First / `make build` obligation its own REQ-HML-017 with AC-HML-017 remapped to it (D4), reframed the REQ-HML-003 / REQ-HML-016 `MAY` clauses into `shall`-form GEARS with explicit Fast-mode taxonomy coverage (D5). Added `tier: S`. |
| 0.1.3 | 2026-07-24 | manager-spec | Plan-phase cross-reference correction only (N1 fallout from the sibling SPECs' v0.1.3). This SPEC's OWN preservation checks are UNCHANGED — AC-HML-014 (c) `grep -c '^license: Apache-2.0$'` → exactly 1, (d) im-not-ai footer → exactly 1, both byte-unchanged — and remain correct. What changed is the stale sibling cross-reference: acceptance.md's license carve-out caveat, progress.md §E.1, and plan.md §F M4 previously described the two sibling SKILLPORT SPECs as asserting "a bare absence of any `license:` field". That is no longer true — both siblings now ship `license: Apache-2.0` per moai's house convention (18 of 28 existing template skills carry it; repo LICENSE is Apache-2.0). The asymmetry is restated as ADD-vs-PRESERVE (siblings add the field to new skills; this SPEC preserves an existing one), not present-vs-absent. No REQ or AC semantics changed. |
| 0.1.2 | 2026-07-24 | manager-spec | Plan-phase iteration-3 delta fixes. (ND1) Fixed ledger-taxonomy enumeration drift: §A.1 previously included "comparisons" and "uncertainty" while REQ-HML-001, REQ-HML-003, and AC-HML-003 omitted both. A single CANONICAL 4-category enumeration (the richer §A.1 form) is now used verbatim at all four sites plus plan.md §F M1. (ND4) Qualified REQ-HML-005 and AC-HML-005 to SUPPLIED (non-inferred) ledger items, resolving the conflict with the acceptance.md §D.2 edge-case-1 inferred-item carve-out; REQ-HML-002's supplied-vs-inferred distinction is now explicitly the gate that decides which items are hard-anchored. (ND3) Scoped the acceptance.md §D.4 "only additions" clause: it now means no DELETION of the preserved machinery (REQ-HML-007..010), while in-place expansion of the REQ-HML-011 graft-point steps 2 and 6 (which necessarily produces deletion+addition diff lines) is expected and permitted. (ND2 / P1-B) Mechanized the clean-room guarantee: AC-HML-014 is now the executable absence-invariant half — with the CRITICAL carve-out that the PRE-EXISTING `license: Apache-2.0` field and im-not-ai footer are byte-preservation targets, NOT absence targets — and the new AC-HML-018 is the clean-room PROCESS attestation half. AC count 17 → 18. |

## §A. Overview (WHAT / WHY)

### A.1 What

Enhance the EXISTING `moai-domain-humanize` skill (currently v1.2.0) with TWO techniques written as ORIGINAL moai authoring, whose functional idea is inspired by skillstead's `writing-quality-editor` skill (an Apache-2.0 project). This is NOT a port of `writing-quality-editor` as a whole skill — that skill overlaps too heavily with humanize and is explicitly rejected as a standalone port. Only the two general techniques are adopted, and both are authored clean-room in moai's own voice and structure (reproducing none of skillstead's wording, section structure, or layout), so the enhancement adds NO skillstead attribution:

1. **Invariant Ledger** — before editing, build a short boundary document capturing what MUST survive editing unchanged. The ledger item taxonomy is the CANONICAL 4-category enumeration defined immediately below. The ledger is a fidelity anchor: the editor must never silently add, remove, narrow, broaden, strengthen, or weaken a ledger item.
2. **Delta Audit** — after editing, compare the final output against the ledger to verify claim & intent parity, identifier/number/condition/limitation/risk survival, and audience/tone/purpose fit, plus unresolved-actor / ownership / handoff / destructive-effect ambiguity.

#### CANONICAL Invariant Ledger item taxonomy (single enumeration — use verbatim)

This 4-category enumeration is the SINGLE canonical form. It is reproduced verbatim in REQ-HML-001, REQ-HML-003, `acceptance.md` AC-HML-003, and `plan.md` §F M1. Any future edit MUST update all five sites together — a partial edit reintroduces the v0.1.1 enumeration drift (§A.1 carried "comparisons" and "uncertainty"; the REQ and AC sites omitted both).

1. **facts** (with evidence boundaries)
2. **identifiers** (commands, paths, URLs, status values, error codes, product names)
3. **conditions / numbers / dates / versions / units / comparisons**
4. **exceptions / limitations / risks / uncertainty / approvals / rollback / next-actions**

### A.2 Why — additive anti-fact-drift safety layer

The existing humanize skill already has a "Meaning-Preservation Checklist" and an "Anchor facts" workflow step, but these are lightweight and informal. Fact drift — a humanization pass silently changing a number, weakening a hedge, or dropping a risk caveat — is the highest-severity failure mode for a post-editing tool. The Invariant Ledger makes the boundary EXPLICIT before editing, and the Delta Audit makes the survival check SYSTEMATIC after editing. Together they harden the existing meaning-preservation machinery without replacing it.

### A.3 Integration into the existing workflow (ADDITIVE — preserve everything)

The two techniques thread into the existing 8-step humanize workflow rather than bolting on:
- The **Invariant Ledger** expands the existing step "Anchor facts" (currently checklist item 1 / workflow step 2) into a fuller, written boundary document.
- The **Delta Audit** expands the existing step "Self-verify (Fast) or audit + review (Strict)" (workflow step 6) into a systematic ledger-comparison.

Everything currently in the skill is PRESERVED unchanged:
- The S1/S2/S3 severity model.
- The A/B/C/D dual grade tables (prose-mode + copy-mode).
- The 30%/50% over-editing guardrails (prose mode) and the fact-anchor preservation guard (copy mode).
- The 4-language coverage (ko/en/ja/zh) and the language-module + genre-module routing.
- The prose/copy genre modes and the Fast/Strict processing modes.
- The existing footer attribution line ("Category-catalogue structure inspired by the im-not-ai (Humanize KR) project") and the existing `license: Apache-2.0` frontmatter field. Both PRE-DATE and are UNRELATED to this enhancement: they are preserved exactly as-is — not touched, moved, or reworded — and NO skillstead line is added beside them. The only footer element this SPEC changes is the trailing `Version:` line, which bumps 1.2.0 → 1.3.0 per REQ-HML-015.

## §B. Requirements (GEARS)

### B.1 Invariant Ledger (pre-edit boundary)

- **REQ-HML-001** (Ubiquitous): The humanize skill shall define an Invariant Ledger step that, before editing, records the items that must survive editing unchanged, covering all four canonical taxonomy categories: **facts** (with evidence boundaries); **identifiers** (commands, paths, URLs, status values, error codes, product names); **conditions / numbers / dates / versions / units / comparisons**; **exceptions / limitations / risks / uncertainty / approvals / rollback / next-actions**. (This is the canonical enumeration defined in §A.1 — reproduced verbatim.)
- **REQ-HML-002** (Ubiquitous): The skill shall state the ledger fidelity rule — never silently add, remove, narrow, broaden, strengthen, or weaken a ledger item — and shall distinguish SUPPLIED facts from INFERRED adjacent benefits/guarantees, marking each ledger item as one or the other. This supplied-vs-inferred marking is the gate that decides enforcement strength: **supplied** items are hard-anchored and their drift triggers the REQ-HML-005 rollback; **inferred** items are recorded for reviewer awareness only, and their removal is NOT a rollback trigger (an inferred adjacent benefit that the original never asserted may legitimately be dropped during humanization). An unmarked ledger item defaults to SUPPLIED (fail-safe toward preservation).
- **REQ-HML-003** (State-driven): **While** in Fast mode, the skill shall record the ledger as a lightweight inline anchor list that still covers EVERY canonical Invariant Ledger taxonomy category — **facts** (with evidence boundaries); **identifiers** (commands, paths, URLs, status values, error codes, product names); **conditions / numbers / dates / versions / units / comparisons**; **exceptions / limitations / risks / uncertainty / approvals / rollback / next-actions** — the Fast-mode ledger is shorter in FORM, never narrower in CATEGORY COVERAGE. **While** in Strict mode, the skill shall record the ledger as an explicit written boundary document. (Taxonomy per the canonical enumeration in §A.1 — reproduced verbatim.)

### B.2 Delta Audit (post-edit verification)

- **REQ-HML-004** (When an edit pass completes): When the humanization edit pass completes, the skill shall run a Delta Audit comparing the output against the ledger for claim & intent parity, identifier/number/condition/limitation/risk survival, and audience/tone/purpose fit.
- **REQ-HML-005** (When the Delta Audit detects a ledger violation): When the Delta Audit finds any **supplied** (non-inferred) ledger item added, removed, narrowed, broadened, strengthened, or weakened, the skill shall roll back that edit (consistent with the existing meaning-drift rollback rule). The rollback obligation is scoped to supplied items per the REQ-HML-002 supplied-vs-inferred marking; removal of an item marked INFERRED is reported in the audit output but does NOT trigger rollback.
- **REQ-HML-006** (Ubiquitous): The Delta Audit shall feed the existing grading step — a ledger violation is a meaning-distortion flag that forces Grade D, consistent with the existing hard rule.

### B.3 Additive preservation (do not regress)

- **REQ-HML-007** (Ubiquitous): The enhancement shall preserve the existing S1/S2/S3 severity model unchanged.
- **REQ-HML-008** (Ubiquitous): The enhancement shall preserve the existing A/B/C/D dual grade tables (prose + copy) unchanged.
- **REQ-HML-009** (Ubiquitous): The enhancement shall preserve the existing 30%/50% over-editing guardrails (prose) and the fact-anchor preservation guard (copy) unchanged.
- **REQ-HML-010** (Ubiquitous): The enhancement shall preserve the existing 4-language coverage (ko/en/ja/zh) and the language-module + genre-module routing unchanged.
- **REQ-HML-011** (Ubiquitous): The enhancement shall be additive — the Invariant Ledger and Delta Audit thread into the existing "Anchor facts" and "Self-verify/audit" workflow steps rather than replacing any existing step.

### B.4 Portability, clean-room authoring, versioning, Template-First

- **REQ-HML-012** (Ubiquitous): The enhancement content shall be written in English.
- **REQ-HML-013** (Ubiquitous): The enhancement content shall remain template-neutral — free of internal moai-adk SPEC IDs, REQ/AC tokens, audit citations, internal dates, and commit SHAs.
- **REQ-HML-014** (Ubiquitous): The Invariant Ledger and Delta Audit content shall be authored clean-room — ORIGINAL moai authoring in moai's own voice and structure, inspired only by the general technique — and shall NOT reproduce skillstead's wording, section structure, or file layout verbatim or near-verbatim. The enhancement shall add NO skillstead attribution: no skillstead provenance footer line, no `NOTICE` file or entry, and no skillstead-tied `license:` field. The skill's PRE-EXISTING attribution surfaces — the existing `license: Apache-2.0` frontmatter field and the existing im-not-ai (Humanize KR) footer line — are unrelated to this enhancement and shall be preserved exactly as-is, neither touched, moved, nor reworded.
- **REQ-HML-015** (Ubiquitous): The skill shall bump BOTH version surfaces from `1.2.0` to `1.3.0` — the frontmatter `metadata.version` field AND the SKILL.md trailing footer `Version:` line — and shall refresh `metadata.updated`. Leaving either surface at `1.2.0` is a version-drift failure.
- **REQ-HML-016** (Ubiquitous): The skill shall keep the SKILL.md body within its Level-2 budget (≤ ~5K tokens). The DEFAULT placement shall be inline — the ledger schema and audit checklist threaded into the existing body sections. **When** the measured body size after the inline draft exceeds the Level-2 budget, the skill shall relocate the detailed ledger item-taxonomy and audit checklist into a Level-3 reference (`modules/invariant-ledger.md`) and shall leave a pointer in the body. Both placements are permitted; the run-phase selects between them by MEASUREMENT, not preference.
- **REQ-HML-017** (Ubiquitous): The enhancement shall be applied Template-First — the template source `internal/template/templates/.claude/skills/moai-domain-humanize/SKILL.md` shall be edited first, the project-local mirror shall be synced from it, and `make build` shall be run to recompile the embedded FS so the enhanced skill ships in the binary.

## §C. Exclusions

The following are **out of scope**. Routing rationale: the overlapping standalone port is explicitly rejected by the user; the existing machinery is preserved, not redesigned; new languages are a separate concern.

### Out of Scope — adopting writing-quality-editor as a whole skill
- The skillstead `writing-quality-editor` skill is NOT adopted as a standalone skill; only the two general techniques are, and they are authored clean-room as original moai content. Its Compose/Assess/Revise/Adapt mode-inference layer, its editing-contract layer, and its skilled-writing layer are not adopted (they overlap the existing humanize workflow).

### Out of Scope — skillstead attribution
- No skillstead provenance footer line, `NOTICE` file/entry, or skillstead-tied `license:` field is added anywhere in the shipped skill. The enhancement is original authoring, not a derivative work, so no attribution is owed.
- The pre-existing im-not-ai (Humanize KR) footer line and the pre-existing `license: Apache-2.0` frontmatter field are NOT modified by this SPEC — they are unrelated to the enhancement.

### Out of Scope — changing existing humanize machinery
- The S1/S2/S3 severity model, A/B/C/D grade tables, 30%/50% guardrails, and fact-anchor guard are NOT modified — the enhancement is additive only.
- The 4 language modules (ko/en/ja/zh) and the genre modules (design-copy, copy-review) are NOT rewritten.

### Out of Scope — new language coverage
- No new target language is added; the enhancement applies uniformly to the existing 4 languages.

### Out of Scope — automated metric layer
- The enhancement does not add a computed change-rate metric; the existing "LLM estimate, treat borderline as over-threshold" limitation is retained.

## §D. Acceptance Criteria (summary)

Full scenarios in `acceptance.md`. Summary:

- AC-HML-001..003: Invariant Ledger step present with the canonical 4-category item taxonomy + fidelity rule (incl. supplied-vs-inferred marking) + Fast/Strict depth split.
- AC-HML-004..006: Delta Audit step present with the survival-check list, rollback-on-violation (scoped to SUPPLIED items), and Grade-D feed.
- AC-HML-007..011: existing S1/S2/S3, A/B/C/D, 30%/50% + fact-anchor guard, 4-language routing PRESERVED (regression check); enhancement is additive.
- AC-HML-012..017: English content, template neutrality (CI clean), clean-room MECHANICAL absence invariant with the pre-existing im-not-ai footer + `license: Apache-2.0` field preserved byte-unchanged, dual version bump 1.2.0→1.3.0 (frontmatter `metadata.version` + footer `Version:` line), progressive-disclosure placement, Template-First edit + sync + `make build` embed.
- AC-HML-018: clean-room PROCESS attestation (judgment half of REQ-HML-014).

AC↔REQ numbering note: the AC-N ↔ REQ-N diagonal holds through AC-HML-017. AC-HML-018 is the second half of REQ-HML-014 (AC-HML-014 is the mechanical half). The `acceptance.md` §D matrix is the authoritative mapping.

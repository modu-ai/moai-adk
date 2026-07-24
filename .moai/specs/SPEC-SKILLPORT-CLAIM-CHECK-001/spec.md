---
id: SPEC-SKILLPORT-CLAIM-CHECK-001
title: "Clean-room author moai-workflow-docs-claim-check skill"
version: "0.1.4"
status: completed
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: "internal/template/templates/.claude/skills/moai-workflow-docs-claim-check"
lifecycle: spec-anchored
tier: M
tags: "skill, docs-verification, claim-check, clean-room, workflow, epic-skillport"
---

# SPEC-SKILLPORT-CLAIM-CHECK-001 — moai-workflow-docs-claim-check

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-24 | manager-spec | Initial plan-phase draft. Functionally inspired by skillstead `docs-claim-check` (an Apache-2.0 project). |
| 0.1.1 | 2026-07-24 | manager-spec | Plan-phase revision. (1) Clean-room pivot: the shipped skill borrows only the functional capability and is authored independently in moai's own voice/structure — no skillstead attribution (no provenance footer, no NOTICE, no skillstead-tied license); added the clean-room authoring requirement. (2) plan-audit fixes: deleted the residual clarification-gate marker and baked its decision into prose, added a Preflight AC (D2), rewrote comma-compound AC↔REQ cells in full (D4), dropped the non-canonical `related_specs` field (D5), added a clarification-hygiene DoD line (D6), unified the Level-1 description budget figure (D8), and recorded a non-Go worked-example advisory (D7). Added `tier: M`. |
| 0.1.4 | 2026-07-24 | manager-spec | Epic-consistency alignment with the sibling SPEC-SKILLPORT-SVG-INFOGRAPHIC-001 v0.1.3 fixes — both items apply the same discipline that produced N1/N2 there, so a correct implementation cannot be misjudged by a false-FAILing AC. (1) N2-class alignment: AC-DCC-002's whole-frontmatter verification prose ("grep frontmatter: `allowed-tools` CSV, no Write/Edit/Bash/AskUserQuestion") is replaced by a pipe-free `awk` EXACT-LINE extraction scoped to the `allowed-tools` value, matching AC-SVG-018(d) in form but scoped to this skill's read-only set (`Read, Grep, Glob`, not SVG's six-tool list). REQ-DCC-017 correspondingly pins the tool order (matching the `moai-foundation-core` / `-quality` / `-thinking` / `moai-domain-frontend` precedent) so the exact-line match asserts nothing the REQ does not bind, binds the prohibition to the `allowed-tools` value only, and states that read-only BEHAVIOR is REQ-DCC-009..011 / AC-DCC-008 territory. The rewrite also gave AC-DCC-002's previously-unverified REQ-DCC-002 half (Level-1 `description` budget + read-only public-docs-only scope naming) an explicit measured check — the cell mapped REQ-DCC-002 but asserted nothing about it. (2) NOTICE scope aligned to SVG's N5 fix: REQ-DCC-016 now names both surfaces (skill directories AND repository root) and AC-DCC-013(b) asserts `test ! -e NOTICE` at repo root alongside the two skill-directory checks. |
| 0.1.3 | 2026-07-24 | manager-spec | Plan-phase iteration-3 audit delta fix (N1, Epic-wide — applied to this SPEC as the sibling of SPEC-SKILLPORT-SVG-INFOGRAPHIC-001). `license:` now follows the house convention: the skill SHALL ship `license: Apache-2.0` (moai's own declaration — 18 of 28 existing template skills carry it; repo LICENSE is Apache-2.0). §A.1 gained an explicit house-convention paragraph, REQ-DCC-016 gained the affirmative clause, and AC-DCC-013(c) was rewritten from a bare `^license:` → 0-matches absence check — which contradicted the house convention and would have false-FAILed a correct implementation — into a `grep -c '^license: Apache-2.0$'` → exactly-1 check plus a pipe-free frontmatter-scoped skillstead-absence assertion. No other defect from the iteration-3 audit applies to this SPEC (N2-N5 are SVG-scoped). |
| 0.1.2 | 2026-07-24 | manager-spec | Plan-phase iteration-3 delta fixes. (D9) Purged the surviving pre-pivot "port of the Apache-2.0 skillstead skill" sentence from §A.1 paragraph 1 — the clean-room framing is now stated once, in the §A.1 clean-room paragraph, with no contradicting opener. (P1-B) Mechanized the clean-room guarantee: AC-DCC-013 is now the executable absence-invariant half (vendor-token grep + no-NOTICE + license check) and the new AC-DCC-016 is the clean-room PROCESS attestation half, replacing the unexecutable "no verbatim/near-verbatim wording" reviewer check. (D10) The non-Go worked-example MUST is now enforceable via the new AC-DCC-017 → REQ-DCC-015. (D12) Aligned the REQ-DCC-003 label parenthetical with its body clause. (D13) Gave AC-DCC-001 (local-copy half), AC-DCC-003 (measured ≤ ~5K Level-2 budget), and AC-DCC-010 (judgment residue labelled) executable checks. AC count 15 → 17. |

## §A. Overview (WHAT / WHY)

### A.1 What

Create a NEW read-only skill, `moai-workflow-docs-claim-check`, that verifies whether the claims made in public-facing documentation (README, release notes, install/usage docs) are supported by user-supplied evidence (files, logs, command outputs). Only the functional CAPABILITY is borrowed from prior art; the shipped skill's expression is authored independently. The functional inspiration is recorded EXACTLY ONCE in this SPEC — in the clean-room paragraph below — and nowhere in the shipped template files.

The skill runs a 3-phase workflow — **Preflight → Claim Triage → Validation** — that decomposes composite claims into atomic units and assigns each atomic claim exactly one judgment label from a fixed 4-value set. It NEVER executes commands, generates patches, or performs code review; instead it declares its review scope and names the specific commands/files a human would need to run.

The skill is functionally inspired by skillstead's `docs-claim-check` skill (an Apache-2.0 project): only the functional CAPABILITY (claim-vs-evidence verification with a fixed label taxonomy) is borrowed. The SKILL.md body, references, and any bundled scripts are authored **clean-room** — independently, in moai's own voice and structure — and reproduce none of skillstead's wording, section structure, code, or file layout. Because ideas/functionality are not copyrightable and only original expression is used, the shipped skill files carry NO skillstead attribution (no provenance footer, no `NOTICE` entry, no skillstead-tied `license:` value). This internal SPEC records the functional inspiration as design rationale; the shipped template files do not.

**House-convention `license:` field (NOT attribution).** The shipped skill DOES carry a `license: Apache-2.0` frontmatter field. This is moai's OWN license declaration following the existing catalog convention — 18 of the 28 current template skills already carry exactly this field, and the repository's own LICENSE is Apache-2.0. The field is unrelated to skillstead, is not a skillstead-tied license value, and does not constitute attribution. The clean-room prohibition above binds skillstead-TIED license values and skillstead attribution surfaces; it must not be read as a prohibition on moai declaring its own license.

### A.2 Why

moai-adk-go maintains a README 4-locale set plus a 4-locale docs-site (adk.mo.ai.kr). The existing `hns-oss-docs-verify` recipe checks build success, URL validity, and 4-locale parity — but it does NOT check whether a documented claim is actually supported by the code/behavior it describes. This skill fills that gap: a claim-vs-evidence verification workflow with no counterpart in the current catalog.

The skill also OPERATIONALIZES an existing always-loaded policy — `.claude/rules/moai/core/verification-claim-integrity.md` (the "no unobserved-claim" invariant) — for the specific surface of public documents. The policy defines the norm ("a claim is only valid when the actor observed the evidence"); this skill is the executable procedure that applies that norm to a README/release-note/install-doc. The skill cross-references the policy; it MUST NOT duplicate it.

### A.3 Namespace and positioning

- Namespace: `moai-workflow-*` (it is an executable, read-only workflow — sibling to `moai-workflow-testing`, `moai-workflow-project`).
- Complements, does not replace, `hns-oss-docs-verify` (build/URL/parity) — the two operate on different verification axes.

## §B. Requirements (GEARS)

### B.1 Skill existence and structure

- **REQ-DCC-001** (Ubiquitous): The skill shall exist at `internal/template/templates/.claude/skills/moai-workflow-docs-claim-check/SKILL.md` (template source of truth) with a synced copy at the project-local `.claude/skills/moai-workflow-docs-claim-check/SKILL.md`.
- **REQ-DCC-002** (Ubiquitous): The skill frontmatter `description` (Level-1, always-listed) shall summarize the trigger within the ~100-token Level-1 metadata budget and name the read-only, public-docs-only scope.
- **REQ-DCC-003** (Where progressive disclosure applies): Where progressive disclosure applies, the skill shall keep the SKILL.md body at Level-2 (≤ ~5K tokens) and place the claim label decision-tree tables and worked-example detail in a Level-3 reference bundle loaded on demand.

### B.2 Workflow behavior

- **REQ-DCC-004** (Ubiquitous): The skill shall run three ordered phases — Preflight, Claim Triage, Validation.
- **REQ-DCC-005** (While in Preflight): While in the Preflight phase, the skill shall confirm the document is public-facing, inventory the provided evidence with versions/timestamps, and flag any secrets or sensitive data for redaction before proceeding.
- **REQ-DCC-006** (When a composite claim is encountered): When the Claim Triage phase encounters a composite claim, the skill shall decompose it into atomic claims such that each atomic claim receives exactly one judgment label.
- **REQ-DCC-007** (Ubiquitous): The skill shall assign to each checkable atomic claim exactly one label from the fixed set {`verified`, `unsupported`, `stale-suspected`, `needs-human`}, applied via an ordered decision tree (needs-human → stale-suspected → verified → unsupported), and shall exclude subjective statements from labeling.
- **REQ-DCC-008** (When a claim is labeled `unsupported`): When the Validation phase labels a claim `unsupported`, the skill shall record a reason drawn from {`missing-evidence`, `contradicted`, `insufficient-coverage`}.

### B.3 Hard boundaries (read-only contract)

- **REQ-DCC-009** (Unwanted behavior): The skill shall not execute commands during assessment; instead it shall name the specific commands and files that would supply the missing evidence.
- **REQ-DCC-010** (Unwanted behavior): The skill shall not generate fixes, patches, or code edits — it produces findings and caveats only.
- **REQ-DCC-011** (Unwanted behavior): The skill shall not perform code-quality review or security review; it shall decline such requests and state the boundary.
- **REQ-DCC-012** (Ubiquitous): The skill output shall include the three mandatory sections — Input Scope Reviewed, Claim Assessments (an atomic-claim table), and Boundary Notes — and the Boundary Notes shall certify that no commands were executed during the assessment.

### B.4 Policy integration and portability

- **REQ-DCC-013** (Ubiquitous): The skill shall cross-reference `.claude/rules/moai/core/verification-claim-integrity.md` as the policy it operationalizes for public documents, and shall NOT duplicate that policy's content.
- **REQ-DCC-014** (Ubiquitous): The skill body, references, and any script comments shall be written in English.
- **REQ-DCC-015** (Ubiquitous): The skill body and references shall remain template-neutral — free of internal moai-adk SPEC IDs, REQ/AC tokens, audit citations, internal dates, and commit SHAs — so the skill is portable to all 16 supported languages. The shipped worked examples shall include at least ONE non-Go example (e.g. a Python / JavaScript / Rust documentation claim), so the shipped skill carries no Go language bias. (This binds the SHIPPED skill body and references only; the Go-flavored scenarios in this SPEC's `acceptance.md` are internal artifacts and may remain Go.)
- **REQ-DCC-016** (Ubiquitous): The skill SKILL.md body, references, and any bundled script comments shall be authored clean-room — independently, in moai's own voice and structure — and shall NOT reproduce skillstead's wording, section structure, code, or file layout verbatim or near-verbatim. Only the functional capability is borrowed; the expression is original. The shipped skill files shall carry NO skillstead attribution — no provenance footer line, no `NOTICE` file (neither in the skill directories nor at the repository root, the two surfaces "no NOTICE file or entry" covers), and no skillstead-TIED `license:` frontmatter value. Conversely, the skill SHALL ship moai's own house-convention `license: Apache-2.0` frontmatter field; that field is moai's own license declaration (matching the repository LICENSE and the existing template-skill catalog), is unrelated to skillstead, and does not constitute attribution. An implementation that omits the `license: Apache-2.0` field violates this requirement.
- **REQ-DCC-017** (Ubiquitous): The skill frontmatter shall follow MoAI skill conventions — `allowed-tools` as a comma-separated string holding exactly the read-only tool set `Read, Grep, Glob` **in that order** (the order matches the existing catalog precedent — `moai-foundation-core`, `moai-foundation-quality`, `moai-foundation-thinking`, `moai-domain-frontend` all use it — and is pinned so AC-DCC-002 can verify by exact-line match), any `skills:` as a YAML array, all `metadata` values as quoted strings — and the `allowed-tools` value shall NOT list Write, Edit, Bash, or AskUserQuestion. This prohibition binds the `allowed-tools` value ONLY: it does not bind narrative frontmatter fields (`description`, `when_to_use`), where naming a tool in prose is legitimate and common across the existing catalog. The skill's read-only BEHAVIOR (no command execution, no patches, no code/security review) is bound separately by REQ-DCC-009..011 and verified by AC-DCC-008 — this requirement covers the frontmatter surface only.

## §C. Exclusions

The following are explicitly **out of scope** for this SPEC. Routing rationale: build/URL/parity verification already lives in `hns-oss-docs-verify`; code/security review is a separate concern the skill's boundary contract forbids; runtime tooling is unnecessary for a read-only judgment skill.

### Out of Scope — command execution and auto-fix
- The skill does not run build commands, tests, or any shell command during assessment.
- The skill does not generate patches, PRs, or documentation edits — remediation is a downstream human/agent action.

### Out of Scope — overlap with existing docs verification
- Build-success, URL-validity, and 4-locale parity checks remain owned by `hns-oss-docs-verify`; this skill does not re-implement them.
- Docs-site rendering, Mermaid direction, and menu/icon coupling checks are not touched.

### Out of Scope — code and security review
- Code-quality review, static analysis, and security auditing are declined by the skill's boundary contract and are not added here.

### Out of Scope — policy authorship
- This SPEC does not modify `verification-claim-integrity.md`; the skill only references and operationalizes it.

## §D. Acceptance Criteria (summary)

Full Given-When-Then scenarios, edge cases, and Definition of Done live in `acceptance.md`. Summary:

- AC-DCC-001..003: Skill exists template-first + local-synced; frontmatter valid (Level-1 `description` measured against the ~100-token budget + `allowed-tools` exact-line match on the read-only set `Read, Grep, Glob`); progressive-disclosure structure present.
- AC-DCC-004..008: 3-phase workflow, atomic decomposition, 4-label taxonomy + ordered decision tree, unsupported-reason set, and HARD read-only boundaries present in the body.
- AC-DCC-009..012: 3-section output contract including the "no commands executed" certification; policy cross-reference (no duplication); English body; template neutrality (CI guard clean).
- AC-DCC-013..017: clean-room mechanical absence-invariant (vendor-token grep → 0, no `NOTICE`, house-convention `license: Apache-2.0` present exactly once with no skillstead reference in frontmatter); embedded-FS rebuild; Preflight AC (public-facing confirmation + evidence inventory with versions/timestamps + secret-redaction flag); clean-room PROCESS attestation (judgment half); at least one non-Go worked example.

AC↔REQ numbering note: the AC-N ↔ REQ-N diagonal holds only through AC-DCC-012. AC-DCC-013 and AC-DCC-016 are the two halves (mechanical / judgment) of REQ-DCC-016; AC-DCC-014 traces to REQ-DCC-001; AC-DCC-015 traces to REQ-DCC-005; AC-DCC-017 traces to REQ-DCC-015. The full mapping is the `acceptance.md` §D matrix.

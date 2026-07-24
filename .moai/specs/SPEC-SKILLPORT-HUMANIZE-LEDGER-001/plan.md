# Implementation Plan — SPEC-SKILLPORT-HUMANIZE-LEDGER-001

## §A. Context

Add two techniques — Invariant Ledger (pre-edit boundary) and Delta Audit (post-edit verification) — to the EXISTING `moai-domain-humanize` skill (currently v1.2.0) as ORIGINAL moai authoring whose functional idea is inspired by skillstead's `writing-quality-editor` (an Apache-2.0 project). Part of the **SKILLPORT** Epic. This is the lightest of the three SPECs: it edits ONE existing SKILL.md (plus its template mirror) and is strictly additive — everything already in the skill is preserved.

Clean-room note: ideas/functionality are not copyrightable — only expression is. The ledger + audit content is written in moai's own voice and structure and reproduces none of skillstead's wording, section structure, or layout, so it is not a derivative work and NO skillstead attribution is added. The functional inspiration is recorded here (an internal SPEC file) as design rationale only. The skill's PRE-EXISTING im-not-ai footer line and PRE-EXISTING `license: Apache-2.0` frontmatter field are unrelated to this change and stay exactly as-is.

Development mode: DDD-flavored — the existing humanize behavior (severity model, grades, guardrails, language routing) MUST be preserved (ANALYZE-PRESERVE-IMPROVE). The IMPROVE delta is the ledger + audit layer.

The current SKILL.md structure was read at plan-time: it has Quick Reference (4 operating principles, genre/processing mode selection, output contract), Common Severity Model, Common Quality Grades (dual tables), Over-Editing Guardrails, Meaning-Preservation Checklist, Language Routing, Genre-Module Routing, Implementation Guide (8-step workflow), and a footer attribution line ("Category-catalogue structure inspired by the im-not-ai (Humanize KR) project").

## §B. Known Issues / Resolved Decisions

- **Inline vs Level-3 placement (RESOLVED — default inline, measured escalation):** the run-phase decides inline-vs-Level-3 AFTER measuring body size, and the committed DEFAULT is inline — the two techniques thread into the existing "Anchor facts" (step 2) and "Self-verify/audit" (step 6) workflow steps. Only when the measured body size after the inline draft exceeds the Level-2 budget (≤ ~5K tokens) does the detailed ledger item-taxonomy + audit checklist relocate to a Level-3 `modules/invariant-ledger.md` with a pointer left in the body. REQ-HML-016 permits both placements and fixes the default; no clarification is outstanding.
- **Attribution model (RESOLVED — clean-room, no attribution):** the enhancement adds NO skillstead attribution — no footer provenance line, no repo-level `NOTICE` file/entry, no skillstead-tied `license:` field. The Invariant Ledger + Delta Audit are original moai authoring inspired by the general technique, not a port. The pre-existing im-not-ai footer line and pre-existing `license: Apache-2.0` field are untouched. No open questions remain.

## §C. Pre-flight

- Confirm the existing skill path `internal/template/templates/.claude/skills/moai-domain-humanize/SKILL.md` (verified — present, v1.2.0) and its local mirror.
- Confirm the existing footer attribution line (im-not-ai) and the existing `license: Apache-2.0` frontmatter field — both must be PRESERVED exactly as-is, with NO skillstead line added alongside.
- Confirm BOTH version surfaces that must bump 1.2.0 → 1.3.0: the frontmatter `metadata.version` field and the SKILL.md trailing footer `Version:` line (content-token anchors — do NOT pin line numbers, they drift).
- Confirm the existing workflow steps ("Anchor facts" step 2, "Self-verify (Fast) or audit + review (Strict)" step 6) as the graft points.

## §D. Constraints

- Tier S (complexity: one existing SKILL.md edited, strictly additive). The artifact set intentionally EXCEEDS the Tier S minimum (spec.md + plan.md) by also carrying acceptance.md + progress.md — a superset is permitted and is not an artifact-count deficiency.
- ADDITIVE ONLY: no change to S1/S2/S3, A/B/C/D tables, 30%/50% + fact-anchor guards, or 4-language routing.
- Template neutrality: enhancement content carries no internal tokens.
- Template-First: edit the template source first, then sync the local mirror, then `make build`. Both mirrors must stay byte-consistent (rule-template mirror parity applies to rules; for skills the template is source of truth and local is the synced copy).
- Version bump 1.2.0 → 1.3.0 on BOTH surfaces (frontmatter `metadata.version` AND the footer `Version:` line); refresh `metadata.updated`.
- Clean-room: the ledger + audit prose is original moai authoring; no skillstead wording/structure/layout reproduced and no skillstead attribution added.

## §E. Self-Verification (plan-phase)

- SPEC ID passes the canonical regex → PASS (Bash-executed).
- 12 canonical frontmatter fields present; `created`/`updated`; `tags` CSV string.
- Out-of-Scope section present with `### Out of Scope —` H3 sub-headings + `-` bullets.

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M1 — Ledger + Audit content design (highest change-likelihood; review focus)
The new user-facing content reviewers most need to check:
- The Invariant Ledger item taxonomy — the CANONICAL 4-category enumeration from spec.md §A.1, used verbatim: **facts** (with evidence boundaries); **identifiers** (commands, paths, URLs, status values, error codes, product names); **conditions / numbers / dates / versions / units / comparisons**; **exceptions / limitations / risks / uncertainty / approvals / rollback / next-actions** — plus the fidelity rule (never silently add/remove/narrow/broaden/strengthen/weaken) and the supplied-vs-inferred marking that gates rollback strength.
- The Delta Audit survival-check list (claim & intent parity, identifier/number/condition/limitation/risk survival, audience/tone/purpose fit) + rollback-on-violation **scoped to SUPPLIED items** + Grade-D feed.
- The Fast (inline anchor list) vs Strict (written boundary document) depth split.
Covers REQ-HML-001..006.

### M2 — Additive integration into the existing workflow (integration decision)
- Thread the Invariant Ledger into the existing "Anchor facts" step (step 2) and the Delta Audit into the existing "Self-verify/audit" step (step 6), WITHOUT removing or renumbering existing steps in a way that regresses behavior.
- Diff expectation at the graft points: in-place expansion of steps 2 and 6 NECESSARILY produces deletion+addition diff line pairs. This is expected and permitted — "additive" here means no deletion of the PRESERVED MACHINERY (REQ-HML-007..010), not "the diff contains only added lines" (see acceptance.md §D.4).
- Wire the Delta-Audit-violation → meaning-distortion flag → Grade D linkage into the existing grading hard rule.
Covers REQ-HML-011, 006.

### M3 — Preservation regression pass (do-not-regress decision)
- Verify the S1/S2/S3 model, both A/B/C/D tables, the 30%/50% + fact-anchor guards, and the ko/en/ja/zh + genre-module routing are all present and unchanged after the edit.
Covers REQ-HML-007..010.

### M4 — Placement + clean-room authoring + version + neutrality + Template-First (mechanical, deferred)
- Decide inline vs Level-3 `modules/invariant-ledger.md` placement by MEASURING body size after the inline draft; the committed default is inline (REQ-HML-016).
- Clean-room authoring: draft the ledger + audit prose from this SPEC's technique description ONLY, without consulting skillstead source text; add NO skillstead attribution (no footer line, no `NOTICE`, no skillstead-tied `license:`). Record the clean-room PROCESS attestation in `progress.md` §E.2 (AC-HML-018).
- Clean-room MECHANICAL check (AC-HML-014) — note the ASYMMETRY versus the two sibling SKILLPORT SPECs, which is ADD-vs-PRESERVE (all three Epic SPECs ship `license: Apache-2.0` per house convention; the siblings ADD it, this SPEC preserves it): this skill ALREADY HAS a `license: Apache-2.0` field and an im-not-ai footer line. Those are **preservation targets, not absence targets**. Assert (a) vendor-token grep → 0 and no `NOTICE` added, AND (b) `grep -c '^license: Apache-2.0$'` → exactly 1 and `grep -c 'Category-catalogue structure inspired by the im-not-ai (Humanize KR) project.'` → exactly 1, both byte-unchanged in `git diff`. Do NOT author a bare "no `license:` field present" check — it would FAIL on the legitimate pre-existing field.
- Bump BOTH version surfaces 1.2.0 → 1.3.0 (frontmatter `metadata.version` AND the footer `Version:` line); refresh `metadata.updated` (REQ-HML-015).
- Template-First: edit template source, sync local mirror, `make build` to recompile the embedded FS; verify neutrality CI guard clean (REQ-HML-017).
Covers REQ-HML-012..017.

## §G. Anti-Patterns to avoid

- Adopting `writing-quality-editor` wholesale (explicitly rejected — only the two general techniques, authored clean-room).
- Consulting skillstead source text while drafting the ledger + audit prose (violates the clean-room process attestation AC-HML-018 — draft from this SPEC's technique description only).
- Authoring a bare "no `license:` field present" absence check. The `license: Apache-2.0` field PRE-DATES this SPEC and must survive byte-unchanged; a bare-absence assertion would fail on a legitimate line (see AC-HML-014's license carve-out).
- Treating a passing `grep -ril 'skillstead' → 0` as evidence the ledger + audit were actually ADDED. It is an absence-invariant guard only; positive proof lives in AC-HML-001..006.
- Writing the ledger taxonomy with a partial category list. The canonical 4-category enumeration (spec.md §A.1) includes `comparisons` and `uncertainty` — dropping either reintroduces the v0.1.1 enumeration drift.
- Adding a skillstead attribution footer line, a `NOTICE` file/entry, or a skillstead-tied `license:` field (none is added — the enhancement is original authoring, not a derivative work).
- Touching, moving, or rewording the pre-existing im-not-ai footer line or the pre-existing `license: Apache-2.0` frontmatter field (both are unrelated to this SPEC and stay exactly as-is).
- Bumping only ONE version surface and leaving the other at 1.2.0 (version drift — violates REQ-HML-015; BOTH `metadata.version` and the footer `Version:` line must reach 1.3.0).
- Modifying the S1/S2/S3 model, grade tables, or guardrails while "improving" (violates additive-only REQ-HML-007..009).
- Editing the local `.claude/skills/` mirror without the template source (Template-First violation → overwritten on next `moai update`).
- Renumbering/removing existing workflow steps in a way that silently regresses documented behavior.

## §H. Cross-References

- `internal/template/templates/.claude/skills/moai-domain-humanize/SKILL.md` — the existing skill being enhanced (v1.2.0, read at plan-time).
- `.claude/rules/moai/development/skill-authoring.md` — frontmatter schema, versioning, progressive disclosure, `modules/` layout.
- `internal/template/CLAUDE.md` — Template-First + embedded FS + mirror discipline.
- `CLAUDE.local.md` §15 / §25 — 16-language neutrality + internal-content isolation.
- Sibling Epic SPECs: SPEC-SKILLPORT-CLAIM-CHECK-001, SPEC-SKILLPORT-SVG-INFOGRAPHIC-001.

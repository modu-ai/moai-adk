# Implementation Plan — SPEC-SKILLPORT-CLAIM-CHECK-001

## §A. Context

Author `moai-workflow-docs-claim-check` in the MoAI catalog, borrowing only the functional capability of skillstead's `docs-claim-check` skill (an Apache-2.0 project) and authoring the skill clean-room in moai's own voice and structure. This is one of three SPECs in the **SKILLPORT** Epic (skillstead functional-idea adoption); the other two are `SPEC-SKILLPORT-SVG-INFOGRAPHIC-001` and `SPEC-SKILLPORT-HUMANIZE-LEDGER-001`. The three are independent — no ordering dependency — but share the same cross-cutting constraints (English body, Template-First, progressive disclosure, template neutrality, clean-room authoring with no skillstead attribution, frontmatter format).

Clean-room note: ideas/functionality are not copyrightable — only expression is. Because the skill is authored independently and copies none of skillstead's wording, structure, code, or file layout, it is not a derivative work, so Apache-2.0 attribution is neither required nor added. The functional inspiration is recorded here (an internal SPEC file) as design rationale only; the shipped template files carry zero skillstead reference.

Development mode: DDD is not applicable (no existing behavior to preserve); this is net-new skill authoring. The run-phase work is skill-file authoring plus a template→local sync and `make build`.

## §B. Known Issues / Resolved Decisions

- **Attribution model (RESOLVED — clean-room, no attribution):** the shipped skill files carry NO skillstead attribution — no footer provenance line, no repo-level `NOTICE` file/entry, and no skillstead-TIED `license:` value. The skill is authored clean-room (functional capability borrowed; expression original). This resolves the former repo-level-NOTICE open question: no NOTICE is added, because there is no attribution to consolidate. No open questions remain.
- **`license:` frontmatter field (RESOLVED — house convention, ships Apache-2.0):** the skill DOES ship `license: Apache-2.0`. This is moai's OWN license declaration, not attribution: 18 of the 28 existing template skills already carry exactly this field, and the repository LICENSE is Apache-2.0. Run-phase MUST include the field; omitting it violates REQ-DCC-016. Do NOT author a bare "no `license:` field" absence check — AC-DCC-013(c) asserts `grep -c '^license: Apache-2.0$'` → exactly 1.

## §C. Pre-flight

- Confirm no skill named `moai-workflow-docs-claim-check` exists (verified at plan time — absent).
- Confirm `.claude/rules/moai/core/verification-claim-integrity.md` exists as the cross-reference target (verified — present in both live and template trees).
- Confirm the `moai-workflow-*` namespace is template-distributed (verified via skill-authoring.md § Skills Namespace Policy).

## §D. Constraints

- Read-only tool set: the `allowed-tools` line must read exactly `allowed-tools: Read, Grep, Glob` (that order — matching the `moai-foundation-core` / `-quality` / `-thinking` / `moai-domain-frontend` precedent), so AC-DCC-002 can verify it by exact-line match. No Write/Edit/Bash in the skill's `allowed-tools`. No AskUserQuestion (subagent boundary). The prohibition binds the `allowed-tools` VALUE only — narrative mention of a tool in `description` / `when_to_use` is legitimate.
- Template neutrality (CLAUDE.local.md §25 / §15): the shipped skill body carries no internal SPEC IDs, REQ/AC tokens, audit citations, internal dates, or commit SHAs.
- Progressive disclosure: description ~100 tokens (Level 1); SKILL.md body ~5K tokens (Level 2); decision-tree tables + worked examples in a Level-3 reference file.

## §E. Self-Verification (plan-phase)

- SPEC ID passes the canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (verified via Bash: PASS).
- 12 canonical frontmatter fields present in spec.md; `created`/`updated` (not `_at`), `tags` comma-separated string.
- Out-of-Scope section present with `### Out of Scope —` H3 sub-headings and `-` bullets.

## §F. Milestones (ordered by decision-reversibility — highest-change-likelihood first)

### M1 — Skill interface design (highest change-likelihood; review focus)
The user-facing surface reviewers most need to check:
- Frontmatter `description` + `when_to_use` (Level-1 trigger text) — the words that decide when the skill loads.
- The 3-phase body (Preflight → Claim Triage → Validation).
- The 4-label taxonomy `{verified, unsupported, stale-suspected, needs-human}` + the ordered decision tree.
- The 3-section output contract (Input Scope Reviewed / Claim Assessments table / Boundary Notes) including the literal "no commands executed" certification.
- The HARD read-only boundaries (no exec, no patches, no code/security review).
Covers REQ-DCC-002, 004-012.

### M2 — Level-3 reference bundle (data-shape decision)
- Author `references/label-decision-tree.md` (or equivalent) carrying the ordered label decision tree, the `unsupported` reason set, atomic-decomposition worked examples, and the composite-claim splitting guidance. Keeps the SKILL.md body within the Level-2 budget.
- **Non-Go worked example (ENFORCEABLE — REQ-DCC-015 / AC-DCC-017):** the atomic-decomposition worked examples MUST include at least one NON-Go example (e.g. a Python / JS / Rust docs claim) so the shipped skill avoids Go template-bias — it ships to all 16 supported languages. (The Go-flavored scenarios in `acceptance.md` are internal SPEC artifacts and may remain Go; this binds the SHIPPED skill body/references only.) This was an unenforceable advisory MUST at v0.1.1; it now has a REQ home and an AC, so run-phase cannot silently skip it.
Covers REQ-DCC-003, 006, 007, 008, and the REQ-DCC-015 non-Go-example clause.

### M3 — Policy cross-reference wiring (integration decision)
- Add the cross-reference to `verification-claim-integrity.md` as the operationalized policy; ensure NO duplication of its text (a pointer + one-sentence framing only).
Covers REQ-DCC-013.

### M4 — Template-First placement + neutrality + clean-room authoring (mechanical, deferred)
- Create the skill in the template path first, then sync the local copy, then `make build` to recompile the embedded FS.
- Author clean-room: draft from this SPEC's functional-capability description ONLY, without consulting skillstead source text; add NO skillstead attribution (no footer provenance line, no `NOTICE` file/entry, no skillstead-TIED `license:` value) while DOING ship the house-convention `license: Apache-2.0` field. Record the clean-room PROCESS attestation in `progress.md` §E.2 (AC-DCC-016).
- Run the mechanical absence-invariant check (AC-DCC-013): vendor-token grep → 0; no `NOTICE` in either skill directory nor at the repository root; `grep -c '^license: Apache-2.0$'` → exactly 1 with no skillstead reference in the frontmatter. **This guard is NOT redundant with the existing CI guard**: `internal/template/internal_content_leak_test.go` `leakClasses` covers internal SPEC-IDs / REQ-AC tokens / audit citations / internal dates / memory paths — the vendor token `skillstead` is in none of them, so a leaked attribution footer would pass every existing repo check.
- Verify English body + template-neutrality CI guard clean.
Covers REQ-DCC-001, 014, 015, 016, 017.

## §G. Anti-Patterns to avoid

- Duplicating `verification-claim-integrity.md` content into the skill body (violates REQ-DCC-013 — reference, don't copy).
- Consulting skillstead source text while drafting the skill body (violates the clean-room process attestation AC-DCC-016 — draft from this SPEC's functional-capability description only).
- Treating a passing `grep -ril 'skillstead' → 0` as evidence the skill was IMPLEMENTED. It is an absence-invariant guard only; an empty directory also returns 0. Positive proof lives in the functional ACs (AC-DCC-001, 004-009).
- Adding a skillstead attribution footer line, a `NOTICE` file/entry, or a skillstead-TIED `license:` value (removed per the clean-room decision — none is added to the shipped files).
- Verifying the Write/Edit/Bash/AskUserQuestion prohibition with a whole-frontmatter grep. REQ-DCC-017 binds the `allowed-tools` value only, and a whole-frontmatter tool-token grep MEASURED 10 of 28 existing skills as false positives from `description` / `when_to_use` prose. Use the AC-DCC-002 exact-line match instead.
- OMITTING the house-convention `license: Apache-2.0` frontmatter field, or authoring a bare "no `license:` field present" absence check. The field is moai's own license declaration required by REQ-DCC-016; a bare-absence check would false-FAIL a correct implementation (this is the N1 defect fixed at v0.1.3).
- Adding Write/Edit/Bash to `allowed-tools` (violates the read-only contract).
- Leaking internal tokens (SPEC IDs, REQ tokens, dates, SHAs) into the shipped skill body.
- Authoring the local `.claude/skills/` copy first and forgetting the template source (Template-First violation → overwritten on next `moai update`).

## §H. Cross-References

- `.claude/rules/moai/core/verification-claim-integrity.md` — the policy this skill operationalizes.
- `.claude/rules/moai/development/skill-authoring.md` — frontmatter schema, namespace policy, progressive disclosure.
- `internal/template/CLAUDE.md` — Template-First rule + embedded FS.
- `CLAUDE.local.md` §15 / §25 — 16-language neutrality + internal-content isolation.
- Sibling Epic SPECs: SPEC-SKILLPORT-SVG-INFOGRAPHIC-001, SPEC-SKILLPORT-HUMANIZE-LEDGER-001.

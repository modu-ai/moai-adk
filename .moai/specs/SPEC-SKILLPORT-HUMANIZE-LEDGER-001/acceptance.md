# Acceptance Criteria — SPEC-SKILLPORT-HUMANIZE-LEDGER-001

## §D. Acceptance Criteria Matrix

| AC ID | REQ | Criterion | Verification |
|-------|-----|-----------|--------------|
| AC-HML-001 | REQ-HML-001 | Invariant Ledger step present | body defines a pre-edit ledger with the item taxonomy |
| AC-HML-002 | REQ-HML-002 | Ledger fidelity rule stated + supplied/inferred marking | body states never add/remove/narrow/broaden/strengthen/weaken; AND body requires each ledger item to be MARKED supplied or inferred, states that supplied items are hard-anchored while inferred-item removal is reported but not rolled back, and states the unmarked-defaults-to-supplied fail-safe |
| AC-HML-003 | REQ-HML-003 | Fast/Strict depth split; Fast keeps FULL taxonomy coverage | body: Fast = lightweight inline anchor list that still covers every canonical ledger taxonomy category — **facts** (with evidence boundaries); **identifiers** (commands, paths, URLs, status values, error codes, product names); **conditions / numbers / dates / versions / units / comparisons**; **exceptions / limitations / risks / uncertainty / approvals / rollback / next-actions**; Strict = explicit written boundary document; body states Fast is shorter in FORM, not narrower in CATEGORY COVERAGE. (Taxonomy reproduced verbatim from the canonical enumeration in spec.md §A.1 — all four categories must be present, including `comparisons` and `uncertainty`.) |
| AC-HML-004 | REQ-HML-004 | Delta Audit step present | body defines post-edit audit with survival-check list |
| AC-HML-005 | REQ-HML-005 | Rollback on SUPPLIED-item ledger violation | body: drift of any ledger item MARKED SUPPLIED (added/removed/narrowed/broadened/strengthened/weakened) rolls back that edit; body also states the complement — removal of an item marked INFERRED is reported in the audit output but does NOT trigger rollback. Both halves must be present: a body that says "any ledger item" without the inferred carve-out FAILS this AC (it would contradict §D.2 edge case 1). |
| AC-HML-006 | REQ-HML-006 | Audit feeds grading | body: ledger violation = meaning-distortion flag → Grade D |
| AC-HML-007 | REQ-HML-007 | S1/S2/S3 preserved | severity model present and unchanged (regression check) |
| AC-HML-008 | REQ-HML-008 | A/B/C/D tables preserved | both prose + copy grade tables present and unchanged |
| AC-HML-009 | REQ-HML-009 | Guardrails preserved | 30%/50% prose guard + copy fact-anchor guard present and unchanged |
| AC-HML-010 | REQ-HML-010 | 4-language routing preserved | ko/en/ja/zh modules + genre-module routing present and unchanged |
| AC-HML-011 | REQ-HML-011 | Additive integration | ledger/audit thread into existing "Anchor facts"/"Self-verify" steps; no step removed |
| AC-HML-012 | REQ-HML-012 | English content | enhancement prose in English |
| AC-HML-013 | REQ-HML-013 | Template neutrality (CI clean) | neutrality guard passes for the edited skill path |
| AC-HML-014 | REQ-HML-014 | Clean-room — MECHANICAL half (absence invariant + pre-existing-line preservation) | FOUR checks, ALL required. **Absence (new content must not add attribution):** (a) `grep -ril 'skillstead\|writing-quality-editor' internal/template/templates/.claude/skills/moai-domain-humanize/ .claude/skills/moai-domain-humanize/` → **0 matches**; (b) no NOTICE file added: `test ! -e internal/template/templates/.claude/skills/moai-domain-humanize/NOTICE && test ! -e .claude/skills/moai-domain-humanize/NOTICE` → exit 0. **PRESERVATION (pre-existing lines must survive byte-unchanged — do NOT assert their absence):** (c) the pre-existing frontmatter field `license: Apache-2.0` is still present and byte-identical: `grep -c '^license: Apache-2.0$' <template SKILL.md>` → **exactly 1**; (d) the pre-existing im-not-ai footer line is still present and byte-identical: `grep -c 'Category-catalogue structure inspired by the im-not-ai (Humanize KR) project.' <template SKILL.md>` → **exactly 1**. Confirm (c) and (d) additionally via `git diff` showing neither line in the changed-lines set. See the license-carve-out caveat below the matrix. |
| AC-HML-015 | REQ-HML-015 | Dual version bump — BOTH surfaces | frontmatter `metadata.version` reads `"1.3.0"` AND the SKILL.md trailing footer `Version:` line reads `Version: 1.3.0`; `metadata.updated` refreshed; neither surface is left at 1.2.0 |
| AC-HML-016 | REQ-HML-016 | Placement respects budget (default inline) | SKILL.md body within the Level-2 budget of ≤ ~5K tokens (measured); default inline, with Level-3 `modules/invariant-ledger.md` + a body pointer used only when the measured inline draft exceeds the budget |
| AC-HML-017 | REQ-HML-017 | Template-First edit + sync + embed (cross-cutting process check) | template source `internal/template/templates/.claude/skills/moai-domain-humanize/SKILL.md` edited FIRST; project-local mirror synced from it; `make build` succeeds; enhanced skill embedded (template + local mirror consistent) |
| AC-HML-018 | REQ-HML-014 | Clean-room — JUDGMENT half (process attestation) | The implementing agent records a clean-room PROCESS attestation in `progress.md` §E.2 stating that the Invariant Ledger and Delta Audit prose was drafted **from the technique description in this SPEC only** (original moai authoring), and that skillstead source text was NOT consulted while drafting. A reviewer separately confirms the added prose follows moai skill conventions and the existing humanize voice. This is a PROCESS attestation, NOT a textual-diff check: this SPEC deliberately records no skillstead source text, so a "no verbatim/near-verbatim wording" comparison has no reference corpus and is not executable. |

### License / footer carve-out caveat (binds AC-HML-014)

[CRITICAL] `moai-domain-humanize` is an EXISTING skill, so its clean-room check is ASYMMETRIC and differs from the two sibling SKILLPORT SPECs. The asymmetry is **ADD-vs-PRESERVE, not present-vs-absent**: all three Epic SPECs ship `license: Apache-2.0` per moai's house convention — the two sibling SPECs ADD the field to a newly-created skill, while this SPEC PRESERVES a field that already exists, byte-unchanged. **No sibling asserts a bare absence of a `license:` field.** (Both siblings previously did; those assertions were removed at their v0.1.3 because they contradicted the house convention — 18 of the 28 existing template skills carry `license: Apache-2.0`, and the repository LICENSE is Apache-2.0.) What remains asymmetric here:

- The `license: Apache-2.0` frontmatter field and the im-not-ai footer line **PRE-DATE this SPEC** and are unrelated to the enhancement. They are **preservation targets**, not absence targets.
- The correct assertion is *"no skillstead-tied license or footer was ADDED, and the two pre-existing lines are byte-unchanged"* — **NOT** *"no Apache-2.0 license is present"*. An AC written as a bare `license:` absence check would FAIL on the legitimate pre-existing field and is prohibited here.
- Only ONE footer element changes in this SPEC: the trailing `Version:` line (1.2.0 → 1.3.0, AC-HML-015).

### Absence-invariant caveat (binds AC-HML-014 checks a-b)

A passing `grep -ril 'skillstead' … → 0 matches` is an **absence-invariant guard only**. It proves no attribution token leaked; it is **NOT** evidence that the Invariant Ledger and Delta Audit were actually added to the skill. Positive proof lives in the functional ACs (AC-HML-001 through AC-HML-006) and the preservation-regression ACs (AC-HML-007 through AC-HML-011).

Why this guard is needed at all: the existing template-neutrality CI guard does **not** cover it. `internal/template/internal_content_leak_test.go` `leakClasses` enforces internal SPEC-ID prefixes, REQ/AC token prefixes, audit citations, internal dates, and memory/archive paths — the vendor token `skillstead` is in **none** of those classes, so a leaked attribution footer would pass every existing repo check.

## §D.1 Given-When-Then scenarios

### Scenario 1 — ledger catches fact drift (core path)
- **Given** a Strict-mode humanization of a report sentence "revenue rose 12% in Q3 2025, though the migration risk remains",
- **When** the skill builds the Invariant Ledger (records "12%", "Q3 2025", and the migration-risk caveat), rewrites, then runs the Delta Audit,
- **Then** a rewrite that drops the migration-risk caveat is caught by the Delta Audit as a ledger violation (limitation/risk survival failure), the edit is rolled back, and the output is flagged Grade D (meaning distortion) — demonstrating the anti-fact-drift layer.

### Scenario 2 — additive, existing machinery intact
- **Given** a Fast-mode prose humanization,
- **When** the skill runs,
- **Then** it applies the lightweight inline ledger AND the existing S1/S2/S3 detection, the 30% WARN / 50% HALT guard, and the A/B/C/D grading — none of which are altered by the enhancement.

### Scenario 3 — copy mode unaffected in structure
- **Given** a copy-mode headline rewrite,
- **When** the skill runs,
- **Then** the copy-mode fact-anchor preservation guard still governs (change-rate guard still does not apply), and the Delta Audit's identifier/number survival check reinforces — not replaces — the existing anchor guard.

## §D.2 Edge cases

- Ledger item is an inferred adjacent benefit (not a supplied fact): the ledger marks it as inferred; the audit does not treat its removal as a violation (only supplied facts are hard-anchored, per the REQ-HML-002 marking gate and the REQ-HML-005 supplied-scoped rollback). An UNMARKED item defaults to supplied — the fail-safe direction is preservation.
- Fast mode, short text: the inline anchor list suffices; a full written boundary document is not forced (REQ-HML-003).
- Body size after inline addition exceeds Level-2 budget: detail moves to `modules/invariant-ledger.md` (REQ-HML-016), SKILL.md keeps a pointer.

## §D.3 Definition of Done

- [ ] All 18 AC rows pass.
- [ ] Regression check confirms S1/S2/S3, A/B/C/D (both tables), 30%/50% + fact-anchor guards, and ko/en/ja/zh routing are unchanged.
- [ ] Version bumped 1.2.0 → 1.3.0 on BOTH surfaces (frontmatter `metadata.version` + the footer `Version:` line).
- [ ] Clean-room MECHANICAL half green (AC-HML-014): vendor-token grep → 0, no `NOTICE` added, AND the pre-existing im-not-ai footer line + `license: Apache-2.0` field each still present exactly once and byte-unchanged.
- [ ] Clean-room JUDGMENT half green (AC-HML-018): process attestation recorded in progress.md §E.2.
- [ ] Ledger taxonomy uses the canonical 4-category enumeration verbatim (incl. `comparisons` and `uncertainty`) wherever it appears in the shipped skill.
- [ ] `make build` succeeds; template source + local mirror consistent.
- [ ] Template-neutrality CI guard clean on the edited path.

## §D.4 Quality gate

- Additive-only, correctly scoped: a diff against v1.2.0 shows **NO deletion of the preserved machinery** — the S1/S2/S3 severity model, both A/B/C/D grade tables, the 30%/50% + fact-anchor guards, and the ko/en/ja/zh + genre-module routing (REQ-HML-007..010) are untouched, as are the pre-existing im-not-ai footer line and the pre-existing `license: Apache-2.0` field. No attribution line is added, and none is removed.
- "Additive" does NOT mean "the diff contains only added lines." REQ-HML-011 requires the ledger and audit to thread INTO existing workflow steps 2 ("Anchor facts") and 6 ("Self-verify / audit"), and in-place expansion of those two graft-point steps NECESSARILY produces deletion+addition diff line pairs. Such deletion lines at the graft points are **expected and permitted**. The prohibited deletions are: removal of a preserved-machinery element (REQ-HML-007..010), removal of an existing workflow step, and any step renumbering that silently regresses documented behavior. Scope the diff review accordingly — do not reject the graft-point rewrite as an additive-only violation.
- Progressive-disclosure budget respected (SKILL.md body ≤ ~5K tokens; Level-3 `modules/invariant-ledger.md` only when the measured inline draft exceeds the budget).

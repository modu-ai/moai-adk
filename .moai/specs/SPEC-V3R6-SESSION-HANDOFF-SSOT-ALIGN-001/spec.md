---
id: SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001
title: "session-handoff.md (SSOT) ↔ moai.md §8 (render surface) drift unification"
version: "0.1.0"
status: completed
created: 2026-06-17
updated: 2026-06-17
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/workflow/session-handoff.md + .claude/output-styles/moai/moai.md (local + template mirror)"
lifecycle: spec-anchored
tags: "session-handoff, ssot, render-surface, drift, doc-convention, byte-parity, gears"
era: V3R6
tier: S
---

# SPEC-V3R6-SESSION-HANDOFF-SSOT-ALIGN-001 — SSOT ↔ render drift unification

## HISTORY

- 2026-06-17 — plan-phase artifact set created (spec.md + plan.md + acceptance.md + research.md). Track 2 of the 3-track session-handoff defect analysis. Track 1 (D1/D2/D7/D8) closed under SPEC-SESSION-HANDOFF-ALIGN-001 (commit 82ea1b09f). This SPEC closes the SSOT↔render surface drift axis (track 2). Track 3 (D4 source_session_id dead-feature) is a separate code-level SPEC, out of scope here.
- 2026-06-17 (iter-2) — plan-auditor iter-1 FAIL (0.62, threshold 0.75) → iter-2 correction. 8 defects addressed: D1 BLOCKING (DELETE REQ/AC-SHA-004 — D5b false-defect from windowed-grep undercount; actual count is 9, label already correct); D2 BLOCKING (subsumed by D1 deletion, verified no surviving AC uses line-number-prefix extraction); D3 SHOULD-FIX (re-anchor AC to Localization Table section with baseline-absent phrase); D4 SHOULD-FIX (DROP REQ/AC-SHA-006 — doctrine already cleanly separates conversation_language from 16-prog-lang axis); D5 SHOULD-FIX (fix AC-SHA-003 inverted regex to match bare labels only); D6 MINOR (REQ-SHA-003 tightened — count qualifier does NOT satisfy concern-name qualification); D7 MINOR (all hard line-number windows removed from ACs; content-anchored); D8 MINOR (REQ-SHA-002 line citations noted as secondary attribution). Thesis downgraded from "prevention" to "mitigation + visibility" per plan-auditor falsification verdict. REQ count 9→7, AC count 9→7, no gaps, AC↔REQ coverage 100%.

## §A. Problem Statement / Background

The paste-ready resume / Session Handoff doctrine has a **canonical SSOT** at `.claude/rules/moai/workflow/session-handoff.md` and a **render surface** at `.claude/output-styles/moai/moai.md §8` (Response Templates → Session Handoff). Track 1 (SPEC-SESSION-HANDOFF-ALIGN-001) closed the drift *between the two TREES* (local vs `internal/template/templates/` mirror) plus 4 file-internal defects (D1/D2/D7/D8). Track 1 did **not** address drift *between the two FILES within one tree* — the SSOT (session-handoff.md) and the render surface (moai.md §8) carry duplicated content that has diverged in 3 places (D3/D5a/D6/D9). Because the duplication is manual and unmechanized, every future edit to either file can silently re-introduce drift; track 1's 4-defect closure already recurred on this axis.

This SPEC unifies the two surfaces on a single SSOT and codifies a doc-only **recurrence-mitigation + drift-visibility** mechanism (no Go code, no lint rule, no hook — those would be separate code SPECs and are explicitly excluded by the new-code=0 constraint).

> **Thesis-scope honesty (iter-2 correction, per plan-auditor falsification verdict)**: this SPEC delivers (a) 3 real drift fixes (D3/D5a/D6/D9) plus (b) a documentation convention (SSOT-pointer + self-check sentinel + "render-only, not canonical" marker) that makes future drift **more visible** to an editor who reads the sentinel, but does **not** mechanically prevent it. A future editor who changes one surface without reading the other surface's sentinel produces silent drift — no test fails, no CI breaks. The only **mechanical** catch is a deferred Go lint rule (follow-up code SPEC, explicitly out of scope here per the new-Go-code=0 constraint). The prior "prevention" wording over-promised; this iter-2 revision downgrades to "mitigation + visibility" wherever the thesis made the prevention claim. See plan.md §F.5 for the honest residual-risk note (confirmed honest per verification-claim-integrity §3.5).

### Track decomposition (3 tracks)

| Track | Axis | Defects | Owning SPEC | Status |
|-------|------|---------|-------------|--------|
| 1 | template↔local tree drift + file-internal D1/D2/D7/D8 | D1, D2, D7, D8 | SPEC-SESSION-HANDOFF-ALIGN-001 | completed (82ea1b09f) |
| **2** | **SSOT (session-handoff.md) ↔ render (moai.md §8)** | **D3, D5a (label collision), D6, D9** | **this SPEC** | **draft (iter-2)** |
| 3 | source_session_id 91% fallback dead-feature (code) | D4 | SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001 (future) | not started |

> **D5b WITHDRAWN (iter-2 correction)**: the original iter-1 SPEC claimed a label-vs-count drift defect ("D5b") at moai.md §8 L681 based on a windowed-grep measurement (`sed -n '683,696p'`) that excluded the first checkbox item at L682. The correct count is **9 items at L682–L690**, matching the label "all 9" exactly. The defect does not exist. REQ-SHA-004 / AC-SHA-004 (which codified this non-defect) are DELETED in this iter-2 revision, and subsequent REQs/ACs are renumbered. See research.md §A.5 for the honest mea culpa with 3 convergent re-measurements. D5a (the label-collision defect — three different concerns sharing the bare label "Pre-emit self-check") is REAL and remains in scope.

## §B. Related Work / Scope Boundary

### B.1 In-scope axis

SSOT↔render drift between TWO FILES within ONE tree:
- SSOT: `.claude/rules/moai/workflow/session-handoff.md` (23648B, both trees byte-identical)
- Render surface: `.claude/output-styles/moai/moai.md §8` (51018B, both trees byte-identical)

### B.2 Explicit non-overlap with 3 existing SPECs

| SPEC | Its scope | Boundary vs this SPEC |
|------|-----------|----------------------|
| SPEC-SESSION-HANDOFF-ALIGN-001 (completed) | template↔local tree drift (drift between the two TREES) + file-internal D1/D2/D7/D8 | **Orthogonal axis.** That SPEC closed tree-pair drift + 4 unrelated defects. This SPEC closes the *between-file* SSOT↔render axis. No overlap: its ACs verify tree byte-parity; my ACs verify SSOT↔render cross-reference integrity + tree byte-parity preservation. |
| SPEC-V3R6-SESSION-HANDOFF-AUTO-001 (completed) | Hook-driven auto-persistence of paste-ready resume to memory (Go code: `internal/hook`, `internal/hook/handoff`) | **Code SPEC, no overlap.** That SPEC modifies Go hook code. This SPEC modifies doctrine Markdown only (new Go code = 0). |
| SPEC-V3R6-SESSION-LEGACY-COVERAGE-001 (completed) | `internal/session` package test coverage 77.7%→≥85% (test-only) | **Test SPEC, no overlap.** That SPEC adds `*_test.go` under `internal/session`. This SPEC touches neither `internal/session` nor any Go file. |

### B.3 Explicit non-overlap with track 3 (D4)

D4 (source_session_id 91% fallback dead-feature diagnosis) is a **code-level** defect rooted in `internal/hook`, `internal/cli/session`, and the multi-session coordination layer. It is scoped to a separate Tier M SPEC (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001) and is **NOT absorbed** into this SPEC. The two tracks are independent and may run in parallel. This SPEC mentions `source_session_id` only as a *content* item that the SSOT↔render surfaces must agree on; it does not diagnose or repair the dead-feature.

## §C. Glossary

- **SSOT** — Single Source of Truth. For this doctrine: `.claude/rules/moai/workflow/session-handoff.md`. Other surfaces (moai.md §8) reference it; they do not re-define its content.
- **Render surface** — the file the orchestrator reads at output time to render a banner/template. For this doctrine: `.claude/output-styles/moai/moai.md §8`.
- **SSOT-pointer pattern** — a render-surface block reduced to a cross-reference ("see SSOT §X") rather than a duplicated copy of the SSOT content. Eliminates the duplicated content that drifts.
- **Self-check sentinel** — a numbered self-check item in BOTH the SSOT and the render surface that verifies the *other* surface's structural invariant (item count, table row count, vocabulary token) matches this surface's. A manual-but-attributable parity check.
- **Byte-parity** — the property that `local` and `internal/template/templates/` copies of the same file are byte-identical (enforced by existing `internal/template/...` mirror tests). This SPEC MUST preserve byte-parity for both modified files.

## §D. Requirements (GEARS)

> **Renumbering note (iter-2)**: the iter-1 SPEC had 9 REQs (REQ-SHA-001..009). iter-2 DELETES REQ-SHA-004 (D5b non-defect, see §A D5b WITHDRAWN note) and REQ-SHA-006 (D4 hypothetical non-conflation — the doctrine already cleanly separates the `conversation_language` ISO-639 axis from the 16-programming-language axis; research.md §A.6a confirms zero baseline mentions of the programming-language axis in the Localization Table region, so there is nothing to guard against). After deletion, the surviving REQs are renumbered REQ-SHA-001..007 with no gaps. ACs follow the same renumbering (acceptance.md §D).

### REQ-SHA-001 (D3 — retired `wave` filename token migration) [MUST]

The session-handoff.md SSOT **shall** replace the retired `wave` filename token `<wave>` and the `wave6` / `prev_wave` memory-filename examples with `sprint` / `round` semantics, per `.claude/rules/moai/development/sprint-round-naming.md` AP-SRN-004.

**Where** the `wave` token appears as a filename token or memory-filename example, the SSOT **shall** migrate it to `sprint` (multi-SPEC context) or `round` (within-SPEC context).

**When** a reader encounters the canonical filename pattern, the SSOT **shall** present `project_<sprint>_<spec>_<status>.md` (or `project_<round>_<spec>_<status>.md` for within-SPEC) with a concrete example using the new vocabulary.

### REQ-SHA-002 (D3 — generic-English-noun `wave` scope decision) [SHOULD]

The SSOT **shall** preserve the generic English noun `wave` where it describes a multi-SPEC scheduling concept in prose (e.g. "multi-SPEC waves" / "current wave" — line numbers approximate and cited only as secondary attribution; the AC anchors by content pattern, not line number), **while** the retired filename token and memory-filename examples are migrated per REQ-SHA-001.

**Where** the generic noun `wave` describes prose scheduling context (not a filename token), the SSOT **shall not** mechanically rewrite it, because AP-SRN-004 retires the `Wave` *terminology token* (filename/round-naming), not every English-noun occurrence. The prose noun carries reader meaning distinct from the filename token.

### REQ-SHA-003 (D5a — "Pre-emit self-check" label disambiguation) [MUST]

The SSOT and the render surface **shall** disambiguate the colliding label "Pre-emit self-check" by qualifying it with the surface's owned **concern-name** (not a count, not any other parenthetical):

- The SSOT (session-handoff.md) owns the **"paste-ready budget"** self-check (Block 2 ≤4 refs, Block 4 ≤200 chars, Block 5 single action, Block 6 ≤2 lines, no history/ceremony).
- The render surface (moai.md §8) owns the **"localization render"** self-check (read `conversation_language`, translate labels, preserve emoji, etc.).
- The render surface (moai.md §8) owns the **"session-handoff template completeness"** self-check (Block 1 ultrathink, Block 2 source_session_id, Block 4 ≤4 preconditions, cut-line markers present, etc.).

**When** either file references a self-check, the reference **shall** use the qualified concern-name, not the bare "Pre-emit self-check".

**Qualification definition (D6 clarification)**: the qualifier MUST be a **surface-owned concern name** (e.g. `(paste-ready budget)`, `(localization render)`, `(session-handoff template completeness)`). A parenthetical **count** (e.g. the current `(8 items)` at session-handoff.md) is a **numeric descriptor**, not a concern-name, and does NOT satisfy the qualification requirement — a count-only parenthetical leaves the label collision intact because the same `(N items)` pattern can appear at multiple collision sites. The implementer MUST replace count-only parentheticals with concern-name parentheticals at every collision site.

### REQ-SHA-004 (D6 — localization fallback rule codification) [MUST]

> *Renumbered from iter-1 REQ-SHA-005 after iter-1 REQ-SHA-004 (D5b) deletion.*

The SSOT (session-handoff.md Localization Table section) **shall** codify an explicit fallback rule: **When** `conversation_language` is a locale NOT in the explicit table (en / ko / ja / zh), the English (en) column is the canonical fallback skeleton, and the label translator **shall** render labels into the configured locale using the naturalization principle ("idiomatic phrasing a native reader would expect, not literal word-by-word translation").

**Where** the Localization Table covers only en/ko/ja/zh, the SSOT **shall** state the fallback rule **in the same section** using a phrase that does NOT appear anywhere in the baseline file (the AC verifies presence of a baseline-absent phrase, so it cannot pass vacuously at baseline).

### REQ-SHA-005 (D9 — consolidated anti-pattern index) [MUST]

> *Renumbered from iter-1 REQ-SHA-007.*

The SSOT (session-handoff.md) **shall** add a single consolidated anti-pattern index table at one canonical location, with one row per AP code (AP-D-001..005, AP-V-001..004, plus the general-hygiene items if they receive codes), each row carrying a forward link to the detail section that holds the domain context.

**When** a reader needs to find any anti-pattern by code, the index table **shall** be the single navigational entry point. The three detail sections (`## Anti-Patterns` general hygiene, `### Anti-pattern catalogue` AP-D, `### Anti-pattern` AP-V) **shall** remain in place (they carry domain context); the index is additive navigation, not a replacement.

### REQ-SHA-006 (recurrence-mitigation mechanism — SSOT-pointer + self-check sentinel) [MUST]

> *Renumbered from iter-1 REQ-SHA-008. Mechanism language downgraded from "prevention" to "mitigation + visibility" per the plan-auditor iter-1 falsification verdict (the mechanism is a documentation convention that surfaces drift to a reading editor, not a mechanical guarantee).*

The SSOT (session-handoff.md) and the render surface (moai.md §8 Session Handoff block) **shall** adopt the **SSOT-pointer + self-check sentinel** recurrence-mitigation mechanism (see plan.md §F for the full falsification analysis of the 4 candidate approaches and the recommendation):

- **SSOT-pointer**: the render surface's duplicated Session Handoff structural content (the 6-block skeleton, the cut-line marker spec, the Localization Table, the anti-pattern catalogue) is reduced to cross-references to the SSOT, plus a thin render-specific addition explicitly marked "render-only, not canonical".
- **Self-check sentinel**: each surface carries a numbered self-check item that verifies the *other* surface's structural invariant (Localization Table row count, self-check item count, filename-token vocabulary) matches this surface's. A manual-but-attributable parity check.

**When** a future editor modifies one surface, the self-check sentinel on the *other* surface **shall** surface the drift to any editor who reads it (the sentinel raises the visibility bar; it is NOT a mechanical guarantee — see plan.md §F.5 for the honest residual-risk note and §F.6 for the deferred mechanical lint-rule follow-up).

### REQ-SHA-007 (byte-parity preservation) [MUST]

> *Renumbered from iter-1 REQ-SHA-009.*

**Where** this SPEC modifies `.claude/rules/moai/workflow/session-handoff.md` and `.claude/output-styles/moai/moai.md`, the local copy and the `internal/template/templates/` mirror copy **shall** remain byte-identical after the change. The implementer **shall** apply every edit to both trees.

**While** byte-parity is a pre-existing invariant (enforced by `internal/template/...` mirror tests), this SPEC **shall not** regress it; the AC verifies byte-parity explicitly for both modified files.

## §E. Constraints

1. **New Go code = 0.** This is a doc-convention SPEC. No `.go` files, no lint rules under `internal/spec/`, no hook scripts under `.claude/hooks/`. If the recurrence-mitigation mechanism cannot be achieved doc-only, it is flagged as a deferred follow-up code SPEC rather than expanding scope. *(iter-2: "prevention" → "mitigation" per thesis downgrade.)*
2. **Byte-parity** for both modified files (local ↔ `internal/template/templates/`) MUST hold post-change.
3. **GEARS format** for all requirements (current MoAI standard; EARS legacy backward-compat is for pre-v3 SPECs only).
4. **ACs are mechanically verifiable** (grep-based / line-count-based / byte-parity-based). No prose-only ACs. This SPEC's whole thesis is that unverified claims drift; the ACs must be self-verifying.
5. **`era: V3R6`** explicit frontmatter override (avoids H-6 unclassified; the SPEC depends on lifecycle-sync-gate.md modern-era conventions and is a sibling of track 1 which also carries `era: V3R6`).

## §F. Test / Verification Strategy

All ACs are mechanically verifiable via `grep` / `wc -l` / `diff` / `cmp`. No test code is added. The acceptance.md file enumerates the exact command + expected output for each AC.

### Anti-pattern self-check for this SPEC

- [ ] No AC uses "the section is consistent" or "labels match" without naming the grep/diff command.
- [ ] Every drift claim in research.md cites a re-run grep with observed output (verification-claim-integrity §2 baseline-attribution). *(iter-2 ADDITION: every drift claim re-measured with a content-anchored grep, not a hard line-number window — the D5b false-defect entered via a windowed grep that excluded the disproof line.)*
- [ ] Byte-parity AC covers BOTH modified files, not just one.
- [ ] The recurrence-mitigation mechanism is defensible against the falsification question ("if a future editor changes one surface, will this catch the drift?") with an honest residual-risk note. *(iter-2 CORRECTION: the honest answer is "it surfaces drift to a reading editor, it does not mechanically catch it"; the residual-risk note in plan.md §F.5 carries this honestly.)*
- [ ] *(iter-2 ADDITION)* Every MUST AC FAILS at baseline (pre-fix) and PASSES only after the fix — no vacuous-pass ACs. The plan-auditor iter-1 report flagged 3 vacuous-pass ACs (D3/D4/D5); all are re-anchored in this iter-2 revision.
- [ ] *(iter-2 ADDITION)* No AC uses a hard line-number window (`sed -n 'NNN,MMMp'` or `awk 'NR>NNN && NR<MMM'`) as its primary gate; ACs anchor by content pattern, with line numbers permitted only as secondary attribution.

## §G. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| SSOT-pointer degrades render-time usability (orchestrator reads moai.md §8 at every output; a pointer forces an extra hop) | Medium | plan.md §F evaluates this trade-off; the hybrid (pointer + thin render-only additions) is the recommended primary, preserving render usability while eliminating duplicated canonical content. |
| Self-check sentinel is manual and can be ignored by a future editor | Medium | Honest residual risk — recorded in plan.md §F.5. The sentinel raises the visibility bar but is NOT a mechanical guarantee (per plan-auditor iter-1 falsification verdict); a code-level lint rule is the deferred follow-up (plan.md §F.6). |
| D5a label refactor (renaming "Pre-emit self-check" to qualified concern-names) may break grep-based ACs in sibling SPECs that cite the bare label | Low | grep verified: no sibling SPEC ACs cite the bare "Pre-emit self-check" label as a literal. The qualified concern-names are additive. |
| D9 index table adds a 4th AP catalogue surface, risking further fragmentation instead of consolidation | Low | The index is the single navigational entry; detail sections remain domain-context-only. plan.md §F requires the index to carry the forward links explicitly. |
| *(iter-2 NEW)* Over-promising "prevention" in the thesis language | Medium | Resolved in this iter-2 revision: §A thesis and REQ-SHA-006 body both downgrade "prevention" to "mitigation + visibility" per plan-auditor falsification. The residual-risk note (plan.md §F.5) is preserved as honest. |
| *(iter-2 NEW)* False-defect injection via hard line-number windows in ACs (the D5b mechanism) | Medium | Resolved in this iter-2 revision: all surviving ACs anchor by content pattern; §F anti-pattern self-check gains an explicit "no hard line-number windows" item. |

## §H. Exclusions (What NOT to Build)

### Out of Scope — explicit non-goals for this SPEC

- **No Go code.** No new `.go` files, no lint rules under `internal/spec/`, no hook scripts under `.claude/hooks/`. The recurrence-mitigation mechanism is doc-only (SSOT-pointer + self-check sentinel); a mechanical lint rule is explicitly deferred to a follow-up code SPEC (plan.md §F.6).
- **No track 3 (D4) absorption.** The source_session_id 91% fallback dead-feature diagnosis is a separate code-level SPEC (SPEC-V3R6-SESSION-ID-ATTRIBUTION-REPAIR-001). This SPEC mentions `source_session_id` only as a *content* item the two surfaces must agree on.
- **No moai.md §8 *Localization Contract* anti-pattern catalogues in scope for D9.** moai.md §8 L243/L291 catalogues are about localization violations, a separate concern. D9 is only about the 3 catalogues INSIDE session-handoff.md.
- **No prose-noun `wave` rewrite.** REQ-SHA-002 explicitly preserves the generic English noun `wave` in prose. Only the filename token `<wave>` and memory-filename examples are migrated (REQ-SHA-001).
- **No design.md.** Tier S — design is optional. The recurrence-mitigation mechanism design rationale lives in plan.md §F (falsification analysis of 4 candidate approaches), which is sufficient for Tier S scope.
- **No D5b axis (iter-2 correction).** The iter-1 SPEC's D5b "label-vs-count drift" axis is WITHDRAWN — the label and enumerated count already agree at 9. The corresponding REQ-SHA-004 / AC-SHA-004 are deleted; see §A D5b WITHDRAWN note and research.md §A.5.
- **No 16-programming-language axis guard (iter-2 correction).** The iter-1 SPEC's REQ-SHA-006 "non-conflation" axis is DROPPED — the doctrine already cleanly separates the `conversation_language` (ISO-639) axis from the 16-programming-language axis (research.md §A.6a confirms zero baseline conflation). There is nothing to guard against.

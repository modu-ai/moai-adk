---
id: SPEC-HUMANIZE-001
title: "Copy-Genre AI-Tell Detection Layer for moai-domain-humanize"
version: "0.3.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-10
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills/moai-domain-humanize"
lifecycle: spec-anchored
tags: "humanize, copy-genre, ai-tell, multilingual, skill"
tier: L
---

# SPEC-HUMANIZE-001 — Copy-Genre AI-Tell Detection Layer for `moai-domain-humanize`

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial plan-phase authoring (Tier L, 14 REQ / 16 AC) |
| 0.2.0 | 2026-07-09 | manager-spec | Plan-audit fix (FAIL 0.78 → revise in place): D1 AC-HUM-008 reachable-to-0 rewrite + commit-SHA class; D2 MIT NOTICE.md now in scope (REQ-HUM-015); D5 license reconciliation now in scope (REQ-HUM-016) + REQ-HUM-011 amended; D3 mechanical AC-006a/b/c; D4 Guard 2 concrete sample; D7 ENC-7 evidence footnote. Now 16 REQ / 19 AC checks (17 AC IDs; AC-006 = a/b/c). |
| 0.3.0 | 2026-07-09 | manager-spec | User decision at Implementation Kickoff gate: Korean module RE-AUTHORED as original work from the maintainer's own taxonomy (the source taxonomy carries zero im-not-ai lineage — verified by grep; the MIT claim was an inaccurate v1.0.0 self-description). MIT dependency dissolved: REQ-HUM-001 rewritten (full korean.md re-authoring incl. prose A–J); REQ-HUM-015 rewritten (attribution cleanup — courtesy credit, no NOTICE.md); REQ-HUM-016 simplified (license: Apache-2.0 unchanged); AC-HUM-016/017 replaced; REQ-HUM-011 re-amended. Second reversal recorded in §F. REQ/AC counts unchanged (16 REQ / 19 checks). |

## §A Purpose & Context

The shipped skill `moai-domain-humanize` (template-managed, `tier: optional-pack:design`, currently `version: 1.0.0`) detects and removes AI "tells" from **prose** across four languages (Korean, English, Japanese, Chinese) and rewrites toward human-authored style while preserving meaning. Its four language modules carry prose-only catalogues:

- `modules/korean.md` — categories A–J (A-1…A-19 etc.). Its v1.0.0 self-description "faithful port of the im-not-ai (Humanize KR) taxonomy" is **inaccurate**: the actual content lineage is the maintainer's own taxonomy (the claude.mo.ai.kr `general-humanize-korean` skill, which carries zero im-not-ai references — verified by grep). The inaccurate claim, not the content, created the apparent MIT encumbrance; this SPEC re-authors the module and removes the claim.
- `modules/english.md` — EN-A…EN-J, web-researched.
- `modules/japanese.md` — JA-01…JA-09, web-researched.
- `modules/chinese.md` — CN-A…CN-K, web-researched.

This SPEC extends the skill from **prose-only** to **prose + marketing-copy** humanization. The copy genre (marketing landing pages, headlines, CTAs, taglines, brand/founder storytelling, slide headlines) exhibits AI tells that are **structural and rhetorical, not lexical**: a landing-page headline can be free of *delve*/*tapestry* and still read unmistakably machine-written. The existing prose catalogues do not cover this genre.

WHY this matters: copy is the highest-visibility, highest-stakes text a user ships (a headline is read far more than a paragraph of body prose), and its AI-tell surface is orthogonal to the prose surface already covered. The reference source (the external `general-humanize-korean` skill — the **maintainer's own original taxonomy**, with no im-not-ai lineage) has evolved a mature Korean copy layer; this SPEC re-authors the Korean module (prose + copy) as an original work from that taxonomy, and authors three language-native copy layers (English, Japanese, Chinese) grounded in independent web research.

## §B Scope (user-approved)

1. `modules/korean.md` — **RE-AUTHOR the whole module as an original work** (v0.3.0 user decision): both the existing prose layer (A–J) and the new copy layer (A-20…A-25 copy translationese, L-1…L-8 storytelling slop, M-1…M-3 slide structural slop) are re-authored from the maintainer's own taxonomy (the claude.mo.ai.kr source — the maintainer's IP), referencing im-not-ai only as inspiration. The run phase REWRITES `modules/korean.md`, not merely appends to it. Pattern IDs and severities may align with the maintainer's taxonomy freely.
2. `modules/english.md` — add a **NEW** copy layer (ENC-1…ENC-9), authored from web evidence, NOT translated from Korean.
3. `modules/japanese.md` — add a **NEW** copy layer (JA-10…JA-14), web-authored.
4. `modules/chinese.md` — add a **NEW** copy layer (CN-L…CN-Q), web-authored.
5. `SKILL.md` — promote the copy-mode guard (fact-anchor preservation) and the dual prose/copy grading tables to the shared, all-language section; keep the existing language-routing structure.
6. `internal/template/catalog.yaml` — recompute the `moai-domain-humanize` hash and bump `version: 1.0.0` → `1.1.0`.

## §C Critical Design Finding — Three Non-Transfer Constraints

The research (see `research.md`) established that three Korean patterns **do NOT transfer** to the other languages. This is the load-bearing design constraint of the entire SPEC:

1. **Korean M-2 (predicate-less / particle-ending fragment headline) does not transfer to English.** English headlines and slide titles are natively fragmentary in fully human copy; fragment-ness is a high-false-positive signal in English, not a removable tell. The English copy layer MUST NOT include a standalone fragment-headline removable category.
2. **Japanese 体言止め (noun-ending) is a legitimate, prestigious native copywriting device.** A presence-based S1 (as Korean M-2 is) would systematically mis-flag competent human copy. The Japanese boundary is **ratio and intent, not presence** — hence JA-10 MUST be a frequency-gated S2. The Korean particle-ending half has no clean Japanese analog and MUST NOT be ported.
3. **Chinese 对偶 / 排比 are prized classical devices.** The boundary is **content-first vs template-first, not occurrence count**: skilled 排比 gives each member a distinct concrete fact; AI assembles symmetry first and dilutes information. This is the highest-false-positive area. The dash-contrast headline transfers only narrowly, because the full-width 破折号 —— (GB/T 15834) conventionally marks explanation and topic shift, never binary contrast.

[HARD] Anti-pattern (encoded as REQ-HUM-007): **mechanically translating the Korean copy layer into the other three modules is PROHIBITED** and would cause systematic false positives. Each of the three new modules MUST be a language-native catalogue backed by its own evidence.

## §D GEARS Requirements

### D.1 Copy-layer content (per language)

**REQ-HUM-001 (Korean — original re-authoring; rewritten v0.3.0).**
The run phase **shall** RE-AUTHOR `modules/korean.md` in full — the existing prose layer (categories A–J) AND the new copy layer (A-20…A-25 copy translationese, L-1…L-8 storytelling/brand-narrative slop, M-1…M-3 slide/presentation structural slop) — as an **original work** derived from the maintainer's own taxonomy (the claude.mo.ai.kr `general-humanize-korean` source, which is the maintainer's IP with zero im-not-ai lineage), referencing im-not-ai only as inspiration. The run phase rewrites the module file; it does not merely append. Pattern IDs and severity tiers MAY align with the maintainer's taxonomy freely (recommended alignment: A-20/A-21 and M-1…M-3 as S1; A-22…A-25, L-1…L-8 as S2, with L-1/L-3 carrying the phrase-S1/structure-S2 split). The rewritten module **shall not** claim to be a port of im-not-ai.

**REQ-HUM-002 (English — new, web-authored).**
The `modules/english.md` module **shall** carry a NEW copy-layer catalogue with the `ENC-` (English Copy) prefix (ENC-1…ENC-9), authored from independent web evidence and NOT translated from the Korean catalogue.

**REQ-HUM-003 (Japanese — new, web-authored).**
The `modules/japanese.md` module **shall** carry a NEW copy-layer catalogue continuing the `JA-` scheme (JA-10…JA-14), authored from independent web evidence.

**REQ-HUM-004 (Chinese — new, web-authored).**
The `modules/chinese.md` module **shall** carry a NEW copy-layer catalogue continuing the `CN-` scheme (CN-L…CN-Q), authored from independent web evidence.

### D.2 Shared SKILL.md promotion

**REQ-HUM-005 (copy-mode fact-anchor guard).**
The `SKILL.md` shared (all-language) section **shall** carry the copy-mode over-editing guard, in which fact-anchor preservation (numbers, dates, prices, proper nouns, plus the core promise/benefit preserved 100%) replaces the prose-mode change-rate guard, while expression and sentence structure MAY be rewritten in copy mode.

**REQ-HUM-006 (dual grading tables).**
The `SKILL.md` shared section **shall** present two grading tables — a prose-mode grade table (residual S1/S2 + change-rate) and a copy-mode grade table (residual S1 + fact-anchor loss + self-verification) — so that copy text is graded by fact-anchor integrity rather than by change-rate.

### D.3 Non-transfer & anti-pattern enforcement

**REQ-HUM-007 (non-transfer anti-pattern).**
The `english.md`, `japanese.md`, and `chinese.md` copy layers **shall not** mechanically translate the Korean copy patterns. Specifically: (a) `english.md` **shall not** include a standalone predicate-less-fragment removable category (Korean M-2 does not transfer); (b) `japanese.md` JA-10 (体言止め over-reliance) **shall** be a frequency-gated S2, never a presence-based S1, and the Korean particle-ending fragment tell **shall not** be ported; (c) `chinese.md` 对偶/排比 handling **shall** gate on content-first-vs-template-first, not on occurrence count, and the dash-contrast tell **shall** be scoped to the binary-contrast headline shape rather than to any em-dash.

### D.4 Cross-cutting authoring constraints

**REQ-HUM-008 (Template-First + byte-identity).**
Where the run phase edits any of the five skill files, it **shall** edit `internal/template/templates/.claude/skills/moai-domain-humanize/**` FIRST, then run `make build`, then synchronize the local `.claude/skills/moai-domain-humanize/**` copy, such that the two directory trees are byte-identical afterward.

**REQ-HUM-009 (§25 template neutrality).**
The ported skill content **shall** carry no internal SPEC IDs, no REQ/AC tokens, no internal work dates, no commit SHAs, no internal version annotations (e.g. `v2.6 신규`-style), and no `.moai/reports/research-*` paths, so that the CI guard `.github/workflows/template-neutrality-check.yaml` passes.

**REQ-HUM-010 (module authoring convention).**
Each module's copy layer **shall** present its instruction prose in English and its before/after rewrite examples in the target language only. Because the source taxonomy's instruction prose is Korean, the Korean copy layer **shall** be an English re-authoring of the instruction prose (not a verbatim Korean copy), while its before/after examples remain in Korean.

**REQ-HUM-011 (frontmatter preservation + attribution rewrite, scoped — re-amended v0.3.0).**
The run phase **shall** preserve the `SKILL.md` frontmatter fields `user-invocable: false`, `allowed-tools`, AND `license: Apache-2.0` unchanged (the v0.2.0 license-change exemption is REVOKED per the v0.3.0 user decision — see REQ-HUM-016). The following frontmatter fields are DELIBERATELY changed and EXEMPT from the preservation rule: `metadata.version` (→ `"1.1.0"`, REQ-HUM-013) and `updated` (run-phase ISO date). No other frontmatter field is added or removed. The trailing attribution line and each per-module "Source & License" block are NOT preserved verbatim — they are REWRITTEN per REQ-HUM-015 into a courtesy credit with no license claim and no NOTICE.md pointer.

**REQ-HUM-012 (evidence integrity).**
Every catalogued pattern in the English, Japanese, and Chinese copy layers **shall** trace to a verified source recorded in `research.md`. Unsourced hypotheses **shall** remain quarantined in a clearly-separated section and **shall not** appear in any module's main catalogue.

**REQ-HUM-013 (catalog metadata + routing table).**
The run phase **shall** recompute the `moai-domain-humanize` `hash` in `internal/template/catalog.yaml`, bump its `version` from `1.0.0` to `1.1.0`, bump the `SKILL.md` `metadata.version` to `"1.1.0"`, and update the SKILL.md language-routing table "Source basis" column to reflect the added copy layer per language.

**REQ-HUM-014 (retired-Python known limitation).**
Because the source skill's quantitative Python layer (`metrics.py`, `metrics_v2.py`, `humanize_html.py`, `baseline*.json`) is NOT ported, the change-rate guard becomes an LLM estimate rather than a reproducible measurement. The `SKILL.md` **shall** instruct conservative judgment near the 30%/50% change-rate thresholds (treat a borderline estimate as over-threshold), and the SPEC **shall** record the loss of reproducible before/after improvement percentages as a known limitation.

### D.5 Attribution & Licensing (rewritten v0.3.0 — MIT dependency dissolved per user decision)

**REQ-HUM-015 (attribution cleanup — courtesy credit, no NOTICE.md; rewritten v0.3.0).**
Because the re-authored Korean module (REQ-HUM-001) leaves NO MIT-ported content in the skill, no MIT permission-notice obligation exists and NO `NOTICE.md` file is created. Instead, the run phase **shall** remove or rewrite all five dangling `See NOTICE.md` pointers (the SKILL.md trailing attribution line + each of the four modules' "Source & License" blocks): each module's attribution block becomes a short **courtesy credit** — "category-catalogue structure inspired by the im-not-ai (Humanize KR) project" — carrying **no license claim, no copyright line, and no NOTICE.md pointer**. Inspiration and independently-authored structure do not attach MIT obligations. After the run phase, **zero** `NOTICE.md` references remain anywhere in the skill directory.

**REQ-HUM-016 (license field unchanged — no dangling license pointers; simplified v0.3.0).**
The `SKILL.md` frontmatter `license:` field **shall** remain `Apache-2.0` as-is (matching the repository LICENSE and all other template skills — no SPDX compound expression), and after the run phase the skill directory **shall** contain no dangling license pointers: no `NOTICE.md` reference, no `MIT License` claim, and no per-material license split. The courtesy credit (REQ-HUM-015) names im-not-ai as inspiration only.

## §E Known Limitations

- **Loss of reproducible metrics.** The retired Python layer computed character-level edit distance and per-metric before/after deltas. Without it, the 30%/50% change-rate guard and the "improvement %" figure become LLM estimates. Mitigation: REQ-HUM-014 (conservative judgment near thresholds).
- **Detector unreliability persists.** As in the prose layer, automated AI-text detectors remain unreliable across all four languages (especially CJK polite registers); the copy layer is a pattern-based editing aid, not a detection oracle.
- **English slide-structure evidence is thin.** Dedicated corpus research on English slide-title AI structure is sparse; the transferable slide claims are inherited from the general copy/headline sources, not a slide-specific study (recorded in `research.md`).

## §F Out of Scope

This section enumerates what this SPEC will NOT build. Each excluded item is routed to its correct home.

### Out of Scope — Quantitative Python layer

- `metrics.py`, `metrics_v2.py`, `humanize_html.py`, `baseline*.json`, and the source skill's `tests/` are NOT ported. The user decided the quantitative layer is retired in favor of LLM-applied heuristic guidance. The consequence is recorded as a known limitation in §E, not implemented here.

### Out of Scope — Korean-coupled peripheral skills

- The four peripheral Korean-coupled skills in the source repo (e.g. the general-ai-slop-reviewer chain and the strict-pipeline design notes) are deferred to a follow-up SPEC. Only the four-language `moai-domain-humanize` skill is in scope.

### Out of Scope — New Go code

- This SPEC adds NO Go code. The run phase is documentation authoring (5 skill markdown files) plus a `catalog.yaml` metadata edit and a `make build` regeneration. No `internal/`/`pkg/`/`cmd/` source is modified.

### Out of Scope — Detector engine / automated scoring

- No automated detector, no scoring engine, and no runtime metric are built. Detection remains a human/LLM pattern-matching activity guided by the catalogues.

### Out of Scope — EN/JA/ZH prose-layer changes

- The existing English / Japanese / Chinese prose catalogues (EN-A…EN-J, JA-01…JA-09, CN-A…CN-K) are NOT rewritten; their copy layers are additive. **Exception (v0.3.0)**: the Korean prose layer (A–J) IS re-authored as part of the full `korean.md` rewrite (REQ-HUM-001) — the Korean module is no longer covered by this exclusion.

### Reversal record (v0.2.0 → v0.3.0) — NOTICE.md: out of scope → in scope → MOOT

- **First reversal (v0.2.0)**: the v0.1.0 draft placed NOTICE.md creation OUT of scope; the plan-audit (D2) surfaced that shipping an apparently-MIT-sourced port around a dangling `See NOTICE.md` pointer worsened a licensing obligation, and the user approved bringing NOTICE.md creation IN scope.
- **Second reversal (v0.3.0)**: the user decided at the Implementation Kickoff gate to RE-AUTHOR the Korean module as an original work from the maintainer's own taxonomy (which has zero im-not-ai lineage — the MIT claim was an inaccurate v1.0.0 self-description). NOTICE.md is therefore **moot**: no MIT-ported content remains, so no MIT notice obligation exists. REQ-HUM-015 now mandates the inverse — removing all five `See NOTICE.md` pointers and replacing the attribution blocks with a license-free courtesy credit; REQ-HUM-016 keeps `license: Apache-2.0` unchanged. This note is the record of both reversals.

## §G Cross-References

- `research.md` — consolidated per-language evidence, the three non-transfer findings, verified source URLs, quarantined hypotheses.
- `design.md` — module-structure decision, ID-prefix rationale, copy-mode guard promotion, retired-Python consequence.
- `acceptance.md` — the AC matrix and verification surface.
- Reference source (the maintainer's own taxonomy — zero im-not-ai lineage): the external `general-humanize-korean` skill (`ai-tell-taxonomy.md` §A-20…A-25 / §L / §M, `SKILL.md` copy-mode guard + dual grading tables).
- `CLAUDE.local.md` §2 (Template-First), §15 (language neutrality), §25 (template internal-content isolation).

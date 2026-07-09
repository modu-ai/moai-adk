---
name: moai-domain-humanize
description: >
  AI text humanization and 윤문 (post-editing) specialist that detects and removes
  AI tells while preserving meaning, facts, and figures. Covers Korean, English,
  Japanese, and Chinese with a shared severity model (S1/S2/S3), quality grades
  (A/B/C/D), and 30%/50% over-editing guardrails. Use to make AI-generated text
  read as human-authored without changing what it says (de-ai, naturalness pass).

when_to_use: >
  Use for AI-text humanization and post-editing (윤문): detecting and
  removing AI tells across Korean, English, Japanese, and Chinese,
  applying the S1/S2/S3 severity model and quality grades while preserving
  meaning, facts, and figures.

license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob
user-invocable: false
metadata:
  version: "1.1.0"
  category: "domain"
  status: "active"
  updated: "2026-07-10"
  tags: "humanize, ai-tell, 윤문, post-edit, naturalness, multilingual, copy"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000
---

# moai-domain-humanize

Post-editing specialist that removes "AI tells" from generated text and rewrites it to read as human-authored, while preserving meaning. This is the **editing** counterpart to text generation: it does not write new content, it refines how existing content is said. Covers Korean, English, Japanese, and Chinese, across two genre surfaces: **prose** (columns, reports, blog posts, formal documents) and **marketing copy** (headlines, CTAs, landing pages, brand storytelling, slide titles). Each language module carries a prose catalogue and a copy-layer catalogue; the shared machinery below (severity model, dual grading, mode-specific guardrails) applies uniformly.

---

## Quick Reference

### Operating Principles (4)

1. **Meaning preservation is the top rule.** Facts, numbers, statistics, named entities, quotations, citations, and the author's stance/certainty stay intact. Any meaning drift forces a rollback. In copy mode, "meaning" is defined by the fact anchors plus the core promise/benefit — see the copy-mode guard below.
2. **Evidence-based edits only.** Every change must trace to a detected tell on a specific span. Stylistic "improvements" unconnected to a catalogued tell are themselves an over-editing signal and are forbidden.
3. **Genre and register preservation.** Humanize *within* the source register — academic stays academic, casual stays casual. Never push formal text into slang or vice versa. Copy and slide genres apply their own structural rules (noun-phrase title boundaries, appeal-vs-informational voice) defined in each module's copy layer.
4. **Over-editing prevention.** In prose mode, flag at >30% change (WARN) and halt at >50% change (forced stop / human review) — above 50% you are regenerating, not humanizing. In copy mode, the change-rate guard is REPLACED by the fact-anchor preservation guard (see Over-Editing Guardrails below).

### Genre Mode Selection (Prose vs Copy)

Two operating genres select which guardrail and grading table apply. Default from the text's genre; an explicit user instruction overrides.

| Mode | Genres | Over-editing guard | Grading table |
|------|--------|--------------------|---------------|
| **Prose mode** (default) | column, report, blog, formal/official document | Change-rate guard (WARN >30%, HALT >50%) | Prose-mode grades |
| **Copy mode** | marketing copy, headline, CTA, landing page, brand story, slides | Fact-anchor preservation guard | Copy-mode grades |

### Processing Mode Selection (Fast / Strict)

- **Fast mode** (default, up to ~5,000 chars): a single pass — detect, rewrite, self-verify against the meaning-preservation checklist.
- **Strict mode** (long or high-stakes text, or when requested): separate stages — detect → surgical rewrite → content-fidelity audit (facts/figures/stance unchanged) → naturalness review. Re-run a second pass when the result lands at Grade C.

### Output Contract

Return two things:

1. **The humanized text.**
2. **A short change report**: categories hit (with counts), the final quality grade (A/B/C/D), and — in prose mode — the estimated percent changed. When a guardrail fires, state it explicitly (prose mode: WARN at >30%, HALT at >50%; copy mode: any fact-anchor loss).

---

## Common Severity Model (shared by all 4 languages)

Each tell carries one severity tier. Detectors gate by occurrence count and overlap, because a single tell rarely proves AI authorship — confidence comes from clustering.

| Tier | Name | Rule |
|------|------|------|
| **S1** | Decisive | A single occurrence strongly confirms AI authorship → remove on first occurrence. |
| **S2** | Strong | Acceptable at 1–2 instances → remove at 3 or more. |
| **S3** | Weak | Problematic only when overlapping other tells → downgrade-only contributor. |

## Common Quality Grades (shared by all 4 languages — dual tables)

Graded **after** the rewrite. The genre mode selects the table: prose mode grades on residual tells plus change rate; copy mode grades on residual S1 plus fact-anchor integrity, with NO change-rate band.

### Prose-Mode Grade Table

Residual S1/S2 counts plus improvement % (= proportion of detected tells removed without introducing new ones).

| Grade | Criteria | Action |
|-------|----------|--------|
| **A** | 0 residual S1, ≤2 residual S2, ≥70% improvement | Pass — reads as human-authored |
| **B** | 0 residual S1, ≤4 residual S2, ≥50% improvement | Pass — minor polish remains |
| **C** | 1–2 residual S1, OR <50% improvement, OR over-edit WARN (>30%) | Trigger a second pass |
| **D** | ≥3 residual S1, OR over-edit HALT (>50%), OR meaning drift detected | Request human review; do not auto-ship |

### Copy-Mode Grade Table

Residual S1 (including the module's copy-layer S1 tells), fact-anchor integrity, and self-verification — no change-rate band, because a legitimate headline rewrite routinely changes most of its characters while preserving every anchor.

| Grade | Criteria | Action |
|-------|----------|--------|
| **A** | 0 residual S1, 0 fact-anchor loss, self-verification passed | Pass — ships as human copy |
| **B** | 0 residual S1, ≤1 conservative fact-anchor concern | Pass with an explicit note |
| **C** | 1 residual S1, OR self-verification partially failed | Trigger a second pass |
| **D** | 2+ residual S1, OR 2+ fact-anchor losses | Request human review; do not auto-ship |

Hard rule (both modes): any residual S1 caps the grade at C; any meaning-distortion flag forces D. S3 tells affect the grade only when ≥3 of them overlap and reinforce an S1/S2 finding.

### Over-Editing Guardrails (shared)

**Prose mode — change-rate guard.** Change rate = the proportion of the text altered; target band ~5–30%.

- **>30% changed → WARN.** Surface a caution and cap at Grade C until each edit is justified by a detected tell. Note: padding-removal legitimately shrinks text, so a length drop alone is not a violation — flag when meaning-bearing spans are altered.
- **>50% changed → HALT.** Stop and require human confirmation; revert to the last safe state.
- **Conservative judgment near the thresholds.** This skill carries no quantitative measurement layer, so the change rate is an LLM estimate, not a reproducible metric. Treat a borderline estimate as OVER the threshold: near ~30%, issue the WARN; near ~50%, HALT. Bias toward caution so an over-edit never slips through on an optimistic estimate. (Known limitation: without a computed metric, before/after improvement percentages are estimates as well — report them as such.)

**Copy mode — fact-anchor preservation guard (REPLACES the change-rate guard).** In copy mode, meaning invariance is anchored differently: numbers, dates, prices, proper nouns, and legal notation are preserved 100% character-intact, AND the core promise/benefit of the copy keeps its meaning — while expression and sentence structure MAY be rewritten freely. The change-rate guard does not apply, because copy humanization legitimately rewrites most of a headline; the guard that replaces it is absolute on anchors:

- **Any altered number, date, price, proper noun, or legal notation → rollback** of that edit.
- **Core promise/benefit drift → rollback.** The rewritten copy must promise the same thing to the same audience.
- **No invented specifics.** Replacing vague copy with concrete claims is only allowed when the concrete facts exist in the source or are supplied by the author.

### Meaning-Preservation Checklist (shared, all must hold)

1. Anchor facts first — fix the claims, numbers, names, dates, and certainty level before editing.
2. Edit at sentence/phrase level, not whole-document regeneration.
3. Add no new facts — never invent specifics to replace vagueness; simplify instead, or flag for the author.
4. Drop no load-bearing facts — removing an inflated wrapper must keep the substantive claim inside.
5. Preserve genuine certainty/hedging and technical terminology verbatim.
6. Final diff check — compare facts, tone, certainty, and examples against the original; revert any edit that drifts.

---

## Language Routing

Each target language has its own tell catalogue (categories, before/after examples in the target language, per-category severity). Load the module that matches the text being edited:

| Language | Module | Source basis |
|----------|--------|--------------|
| Korean (한국어) | `modules/korean.md` | Original catalogue — prose (10 categories A–J) + copy layer (A-20…A-25, L-1…L-8, M-1…M-3) |
| English | `modules/english.md` | Web-researched catalogue — prose (EN-A…EN-J) + copy layer (ENC-1…ENC-9) |
| Japanese (日本語) | `modules/japanese.md` | Web-researched catalogue — prose (JA-01…JA-09) + copy layer (JA-10…JA-14) |
| Chinese (中文) | `modules/chinese.md` | Web-researched catalogue — prose (CN-A…CN-K) + copy layer (CN-L…CN-Q) |

The Korean module is an original catalogue; the English, Japanese, and Chinese modules are independently web-researched catalogues built on the same architecture. Each module's copy layer is language-native — copy tells do NOT transfer mechanically between languages (English headlines are natively terse; Japanese 体言止め is prestigious craft gated by frequency, not presence; Chinese 对偶/排比 is judged content-first, not by count) — so never apply one language's copy rules to another. The common severity model and quality grades above apply uniformly to every module — the modules add only the language-specific tell categories, severities, and example rewrites.

For mixed-language text, detect the dominant language and route to its module; apply each module independently to its spans when the text is genuinely multilingual.

---

## Implementation Guide

### Workflow (per text)

1. **Identify language, genre mode, and processing mode.** Pick the module by dominant language; pick prose vs copy mode by genre (see Genre Mode Selection); pick Fast vs Strict by length / stakes.
2. **Anchor facts.** Record the numbers, names, dates, quotations, and stance that must not change (meaning-preservation checklist item 1). In copy mode these anchors are the guard itself.
3. **Detect tells.** Scan against the module's catalogue — the prose categories in prose mode, plus the module's Copy Layer categories in copy mode. Record each hit with its category ID, span, and severity. Count occurrences (S2/S3 gate on repetition).
4. **Rewrite surgically.** Edit only flagged spans. Replace each tell with a natural rendering in the same register. Do not touch unflagged text.
5. **Apply the mode's guardrail.** Prose mode: estimate the change rate and apply the change-rate guard (WARN >30%, HALT >50%, conservative near the thresholds). Copy mode: verify every fact anchor and the core promise/benefit instead.
6. **Self-verify (Fast) or audit + review (Strict).** Re-run the meaning-preservation checklist. In Strict mode, run the content-fidelity audit and naturalness review as separate stages.
7. **Grade.** Count residual S1/S2 and improvement %; assign A/B/C/D. Second pass on C; human review on D.
8. **Emit** the humanized text + change report.

### Detection note (shared across languages)

Automated AI-text detectors are unreliable across these four languages (notably weak on CJK polite registers, where they false-positive on correct formal writing). This skill is a **pattern-based editing tool**, not a detection oracle: rely on the catalogued tell categories and the clustering-based severity gates, not on a detector's verdict.

### Common pitfalls

- **Re-injecting AI-ness.** Rewriting AI text with a fresh full regeneration tends to add new tells. Favor surgical edits to flagged spans over wholesale rewriting.
- **Fabricating specifics.** When a module calls for concrete detail to replace vague filler and no real specifics exist, simplify or flag for the author — never invent.
- **Style drift.** "Cleaning up" beyond the flagged tells violates principle 2 and inflates the change rate toward the HALT threshold.
- **Mixing registers mid-document.** Keep one consistent register (and, for Japanese/Chinese, one consistent politeness/sentence-ending style) across the whole output.

---

## Works Well With

- `sync-auditor`: independent skeptical review. Use it to score whether the humanized output preserved meaning against the original and met the target grade.

---

Category-catalogue structure inspired by the im-not-ai (Humanize KR) project.

Version: 1.1.0

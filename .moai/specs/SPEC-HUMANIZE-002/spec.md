---
id: SPEC-HUMANIZE-002
title: "Port claude.mo.ai.kr slop-review skill knowledge into moai-domain-humanize v1.2.0 modules with 4-language expansion"
version: "0.1.0"
status: in-progress
created: 2026-07-10
updated: 2026-07-10
author: GOOS행님
priority: P2
phase: "v3.0.0"
module: "internal/template/templates/.claude/skills/moai-domain-humanize"
lifecycle: spec-anchored
tags: "humanize, slop-review, copy, multilingual, skill-port, template"
tier: M
era: V3R6
related_specs: [SPEC-HUMANIZE-001]
---

# SPEC-HUMANIZE-002 — Port claude.mo.ai.kr slop-review knowledge into `moai-domain-humanize` v1.2.0

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-10 | GOOS행님 (via manager-spec) | Initial plan-phase authoring (Tier M, 23 REQ / 24 AC checks). Porting shape + 4-language scope fixed by user decision at kickoff. |

---

## §A Context & Goal

### §A.1 Source inventory (read-only, outside this repo)

Four self-owned knowledge sources in `/Users/goos/MoAI/claude.mo.ai.kr/` carry AI-slop review knowledge not yet present in `moai-domain-humanize`:

| # | Source | Size | Carries |
|---|--------|------|---------|
| S1 | `plugins/moai-coworker/skills/general-ai-slop-reviewer/SKILL.md` | 306 lines | General post-generation slop review for ALL text deliverables: diagnosis checklist, 7 humanization repair strategies, 3-block review output format, meaning-preservation cautions |
| S2 | `plugins/moai-coworker/skills/general-cd-slop-check/SKILL.md` | 265 lines | Design-tool result-copy QA gate: EN/KO cliché formula dictionaries (near-always-slop + context-dependent), structural copy anti-patterns, 6-step review workflow, fix-proposal format (original / reason / 3 alternatives / preferred) |
| S3 | `plugins/moai-designer/skills/cd-slop-check/SKILL.md` | 279 lines | **Near-duplicate of S2** (adds a KO structural-slop table that itself cites this skill's `korean.md` as origin, plus an SSOT-deference note) |
| S4a | `plugins/moai-marketer/skills/marketing-landing-page/references/landing-page/copywriting-rules.md` | 56 lines | Landing copy anti-pattern repair table (vague-claim → concrete-mechanism), required copy element checklist, headline/subhead/body/CTA structural constraints, tone profile |
| S4b | `plugins/moai-marketer/skills/content-card-news/references/card-news/anti-ai-writing.md` | 140 lines | Korean card-news / short-form slide style guide: ending-monotony rules, habit-word tables, cover-length economy, one-thought-per-line, ending-slide rules, self-check checklist |

### §A.2 Target architecture (moai-domain-humanize v1.1.0, verified by SPEC-HUMANIZE-001)

The target skill ships a shared machinery — S1/S2/S3 severity model, dual A/B/C/D grade tables (prose-mode vs copy-mode), 30%/50% change-rate guardrails (prose) and the fact-anchor preservation guard (copy) — plus four language-native catalogue modules:

- `modules/korean.md` — prose A–J + copy layer A-20…A-25, L-1…L-8, M-1…M-3 (M-1 = dash-contrast headline, M-2 = fragment ending, M-3 = "X에서 Y로" transition formula)
- `modules/english.md` — prose EN-A…EN-J + copy layer ENC-1…ENC-9
- `modules/japanese.md` — prose JA-01…JA-09 + copy layer JA-10…JA-14
- `modules/chinese.md` — prose CN-A…CN-K + copy layer CN-L…CN-Q

Architectural invariant (v1.1.0): copy tells do NOT transfer mechanically between languages; each module's catalogue is language-native. The template tree (`internal/template/templates/.claude/skills/moai-domain-humanize/`) and the local tree (`.claude/skills/moai-domain-humanize/`) are byte-identical at plan baseline.

### §A.3 Porting shape (user-fixed decisions — treat as constraints)

1. **ABSORB, not new skill.** The knowledge is absorbed into `moai-domain-humanize` as exactly two new module files — `modules/copy-review.md` (QA-gate workflow + formula dictionaries) and `modules/design-copy.md` (display-copy genre rules). No new standalone skill; the template skill-directory count stays unchanged (plan-baseline: 28 directories at `39c74d777`). Catalog version bumps humanize `1.1.0 → 1.2.0`.
2. **4-language expansion is in scope.** The ported Korean-centric catalogues extend to KO/EN/JA/ZH following the v1.1.0 layered approach: KO original authorship, EN ported from the self-owned EN source plus web-grounding, JA/ZH independently web-grounded. Machine-translation of the Korean catalogue is prohibited and mechanically checked (language-native boundary constraints in the AC-HUM-006a/b/c tradition).

### §A.4 What is genuinely new vs. what deduplicates

New knowledge (not in v1.1.0): the review-only QA-gate operating mode (detect + propose, never auto-apply); the 6-step review pipeline with input classes; the fix-proposal format (original / reason / ≥3 alternatives / preferred); the review report template; the 7-strategy repair playbook; formula-level cliché dictionaries ("Reimagine your [X]", "Powered by AI", "Where X meets Y", unfounded-stat claims, 혁신적인/차세대/재정의하는 slot formulas); landing-page structural constraints and the vague-claim → concrete-mechanism repair table; short-form card/slide genre rules.

Deduplication anchors (illustrative, resolved exhaustively at run-phase): the S1/S3 KO structural-slop table cites this skill's own `korean.md` D/J + M-1/M-3 as origin — circular, reference only; "Unleash your [X]" / "Transform the way you [X]" / "Supercharge" → ENC-1; "No more X. Just Y." → ENC-2 family; "Fast. Smarter. Better." listing → ENC-3; colon-emphasis heading (JA) → JA-11; slot-fill landing templates (ZH) → CN-N; unfounded-stat claims (ZH) → CN-Q; KO habit words overlap `korean.md` prose categories. Ported entries already catalogued in v1.1.0 are cross-referenced, never re-defined.

### §A.5 Provenance & license (explicit statement)

The claude.mo.ai.kr source content is the maintainer's own intellectual property (same authorship as this repository). Direct reuse of its taxonomy, examples, and phrasing is therefore acceptable without external attribution or license import. Per the SPEC-HUMANIZE-001 precedent: the skill's license stays **Apache-2.0 single** — no MIT import, no NOTICE file, no compound license expression; the existing im-not-ai courtesy credit in `SKILL.md` remains untouched. Ported content is re-grounded as neutral prose: all plugin-namespace references, internal version-sync comments, and internal tracking tokens from the source files are stripped.

---

## §B Requirements (GEARS)

### §B.1 Absorption shape

- **REQ-H2-001** (Ubiquitous): The moai-domain-humanize skill shall absorb the S1–S4 source knowledge as exactly two new module files — `modules/copy-review.md` and `modules/design-copy.md` — with no new standalone skill and no files created outside the skill directory (template + local trees) other than `internal/template/catalog.yaml` edits; the template skill-directory count shall remain unchanged from the plan baseline (28 directories at `39c74d777`).
- **REQ-H2-002** (Ubiquitous): The two near-duplicate source skills (S2 `general-cd-slop-check` + S3 `cd-slop-check`) shall be consolidated into the single `modules/copy-review.md` — the union of their content appears exactly once; no file named after the source skills (e.g., `*slop*`) is created in the skill directory.

### §B.2 copy-review.md — QA-gate workflow module

- **REQ-H2-003** (Ubiquitous): `modules/copy-review.md` shall define a 6-stage review pipeline — (1) copy collection with input classes (single copy text, copy bundle, full exported document text, pasted chat extract), (2) language detection and per-language dictionary routing, (3) pattern matching against the formula dictionaries and the v1.1.0 catalogues, (4) context inference for context-dependent findings (industry, brand tone, audience), (5) fix proposal, (6) review report emission.
- **REQ-H2-004** (Ubiquitous): The module shall define a fix-proposal format carrying, per finding: the original span, the reason (named pattern/category), at least 3 alternative rewrites, and one preferred pick with a one-line justification; alternatives shall be concrete-fact-oriented and shall use explicit placeholders (e.g., `[number]`, `[persona]`) when no real fact is available in the source.
- **REQ-H2-005** (Ubiquitous): The module shall define a review report template: a summary table (findings by severity), per-finding sections in the fix-proposal format, and follow-up recommendations (including chaining into the skill's default rewrite flow).
- **REQ-H2-006** (Event-driven): **When** the skill is invoked as a post-generation QA gate on copy produced by another tool or workflow, the copy-review module shall operate in review-only mode — detect and propose without auto-applying any rewrite (user reviews before application) — distinct from the skill's default rewrite mode.
- **REQ-H2-007** (Ubiquitous): The module shall carry a repair-strategy playbook of exactly 7 named strategies with IDs `CRS-1`…`CRS-7` — voice concretization, rhythm variation, transition-word diet, list-to-prose conversion, specificity injection, opening rewrite, closing rewrite — each with at least one before/after example.
- **REQ-H2-008** (Ubiquitous): The module shall carry per-language formula dictionaries under dedicated headings naming the language in English (Korean / English / Japanese / Chinese), with entry IDs `CR-KO-N`, `CR-EN-N`, `CR-JA-N`, `CR-ZH-N` and minimum unique entry counts after deduplication: EN ≥ 8, KO ≥ 6, JA ≥ 4, ZH ≥ 4. Each entry carries: the formula pattern, a detection signal, a severity tier (S1/S2/S3), and a rewrite direction with a target-language example.
- **REQ-H2-009** (Ubiquitous): The source material's binary confidence vocabulary (near-always-slop vs. context-dependent) shall be remapped into the shared S1/S2/S3 severity model — near-always-slop formulas map to S1 or S2 per single-occurrence decisiveness; context-dependent expressions map to S3 or context-gated S2. The shipped modules shall not carry "Tier 1" / "Tier 2" as a parallel severity vocabulary.
- **REQ-H2-010** (Event-driven, dedup): **When** a candidate ported entry is already catalogued in a v1.1.0 language module (korean A-20…A-25 / L-1…L-8 / M-1…M-3 / prose A–J; ENC-1…ENC-9; JA-10…JA-14; CN-L…CN-Q), the new modules shall cross-reference the existing category ID instead of re-defining the entry — zero re-definition rows for existing IDs.

### §B.3 design-copy.md — display-copy genre module

- **REQ-H2-011** (Ubiquitous): `modules/design-copy.md` shall carry landing-page genre rules with IDs `DCG-N` (or per-language `DCG-{KO|EN|JA|ZH}-N` where language-specific): headline/subheadline/body/CTA structural constraints, the vague-claim → concrete-mechanism repair table (each anti-pattern paired with a repair direction that names a mechanism, metric, or evidence), the required copy element checklist, and tone-profile guidance.
- **REQ-H2-012** (Ubiquitous): The module shall carry short-form card/slide genre rules: cover-copy economy, one-thought-per-line body discipline, ending-slide rules (no formulaic reader-address closers), numeral style, and exclamation restraint.
- **REQ-H2-013** (Ubiquitous, language-native adaptation): Every genre rule with a language-dependent parameter (length limits, ending-form conventions, script/numeral conventions) shall carry per-language native adaptations for all 4 languages; a KO-native numeric limit (e.g., a Hangul character count) shall not be copied verbatim into the EN/JA/ZH adaptations — each language states its own native measure.

### §B.4 4-language expansion discipline

- **REQ-H2-014** (Ubiquitous): The 4-language content shall follow the v1.1.0 authorship layering — KO entries are original authorship (ported and reorganized from the self-owned KO sources); EN entries are ported from the self-owned EN source material plus web-grounded additions; JA and ZH entries are independently grounded, language-native entries. Machine-translation of the Korean catalogue into EN/JA/ZH is prohibited. Instruction prose is English; all before/after examples are in the target language and script.
- **REQ-H2-015** (State-driven): **While** a JA or ZH dictionary entry does not map to an existing v1.1.0 catalogued tell, the entry shall carry a grounding note tracing to the module's sources section — ungrounded hypotheses shall not enter the shipped dictionaries.
- **REQ-H2-016** (Ubiquitous): The new modules shall reuse the skill's shared machinery — the S1/S2/S3 severity model, the copy-mode grade table, and the fact-anchor preservation guard — and shall not define a parallel grading system; all fix proposals preserve fact anchors (numbers, dates, prices, proper nouns) and never invent specifics.

### §B.5 Integration & versioning

- **REQ-H2-017** (Capability gate): **Where** the text under review is display-surface copy (landing page, slide/card deck, design-tool result copy), the skill shall route to `modules/design-copy.md` in addition to the matching language module; **Where** the invocation is a post-generation QA-gate review, the skill shall route to `modules/copy-review.md`. `SKILL.md` shall carry this genre-module routing extension; the existing Language Routing table and the four language modules remain behaviorally unchanged.
- **REQ-H2-018** (Ubiquitous): `SKILL.md` `metadata.version` shall bump `1.1.0 → 1.2.0` (with `updated:` refreshed and the footer version line matched), and the `internal/template/catalog.yaml` humanize entry shall bump `version: 1.1.0 → 1.2.0` with its content hash regenerated through the existing generator (`make build` / gen-catalog-hashes) — hashes are derivatives, never hand-edited.

### §B.6 Isolation, neutrality, license

- **REQ-H2-019** (Ubiquitous): The four existing language modules (`korean.md`, `english.md`, `japanese.md`, `chinese.md`) shall remain byte-identical to the plan baseline in BOTH trees — deduplication is realized by referencing from the new modules into existing IDs, never by editing the v1.1.0-verified catalogues.
- **REQ-H2-020** (Ubiquitous): All changes shall be authored in the template tree first (`internal/template/templates/...`, Template-First rule) and synced to the local tree; after run-phase the two trees shall be byte-identical for the humanize skill directory.
- **REQ-H2-021** (Ubiquitous, neutrality): The shipped skill content shall carry zero occurrences of the 7 forbidden neutrality classes: (1) internal SPEC IDs, (2) REQ/AC tokens, (3) commit SHAs, (4) internal work dates in body content (SKILL.md frontmatter `created:`/`updated:` excepted), (5) source-plugin namespace references (`moai-writer:`, `moai-coworker:`, `moai-marketer:`, `moai-designer:`, `moai-officer:`, `moai-code:`), (6) source version-sync HTML comments (`3-point sync`, `plugin.json`, marketplace metadata), (7) internal report/research paths.
- **REQ-H2-022** (Ubiquitous, license): The skill's `license: Apache-2.0` frontmatter shall remain unchanged; zero `MIT License` tokens in the skill directory; the existing im-not-ai courtesy credit is preserved verbatim. The self-owned-IP direct-reuse rationale is recorded in §A.5 of this SPEC (no attribution text is added to the shipped skill for the claude.mo.ai.kr sources).
- **REQ-H2-023** (Unwanted behavior): The new modules shall not introduce programming-language bias — no programming language is named as primary or preferred anywhere in the new content. (Natural-language coverage of KO/EN/JA/ZH is the module's subject matter and is compliant with the 16-programming-language neutrality contract, which governs programming languages only.)

---

## §C Non-functional constraints

- **Zero Go source changes.** This SPEC is markdown + catalog.yaml only; the catalog hash is regenerated by the existing generator, not by new code.
- **Write scope whitelist** (run-phase): the 3 skill files (SKILL.md + 2 new modules) in each of the two trees, `internal/template/catalog.yaml`, and `.moai/specs/SPEC-HUMANIZE-002/*`. Everything else — including the 4 existing language modules — is PRESERVE.
- **Source repo read-only.** `/Users/goos/MoAI/claude.mo.ai.kr/` is never modified.
- **Verification claims** follow `.claude/rules/moai/core/verification-claim-integrity.md` — every AC PASS cites an executed command + verbatim output.

---

## §D Acceptance

The full AC matrix (24 checks: 16 mechanical / 3 hybrid / 5 manual), verification commands, and Given-When-Then scenarios live in `acceptance.md` (Tier M third artifact). Traceability: every REQ maps to ≥1 AC (see acceptance.md §D.5).

---

## §E Known limitations & recorded decisions

1. **New-entry placement decision.** All new pattern entries live in the two new genre modules; the four v1.1.0 language modules are byte-frozen (REQ-H2-019). The alternative — extending ENC-10+/JA-15+/CN-R+ series inside the language modules — was rejected to avoid regression risk on the 19-AC-verified v1.1.0 content. Consequence: a Korean copy review may load `korean.md` + `copy-review.md` (+ `design-copy.md` when display-surface); the routing extension (REQ-H2-017) makes this explicit.
2. **JA/ZH minimum counts are conservative (≥4 each).** If run-phase web-grounding cannot honestly reach the minimum, the implementer returns a blocker report proposing a threshold amendment — fabricating ungrounded entries to hit a count is prohibited (REQ-H2-015).
3. **No quantitative detector.** As in v1.1.0, the skill is a pattern-based LLM editing tool; formula "detection signals" are authoring guidance, not regex engines. Severity gating relies on the shared clustering rules.
4. **Skill-count baseline correction.** The kickoff brief cited a skill count of 18; the measured template baseline is 28 skill directories at `39c74d777`. ACs encode "unchanged vs. measured baseline", not a hard-coded figure.

---

## Out of Scope

The following are explicitly out of scope for SPEC-HUMANIZE-002:

### Out of Scope — New standalone skill or plugin surface

- No new skill directory, no plugin port, no marketplace artifact; the absorption target is exclusively the existing `moai-domain-humanize` skill.

### Out of Scope — Editing the four existing language modules

- `korean.md` / `english.md` / `japanese.md` / `chinese.md` stay byte-identical in both trees; any needed change there is a follow-up SPEC (blocker report first).

### Out of Scope — Go source and detection automation

- No Go code, no regex detection engine, no lint rule; catalog hash regeneration uses the existing generator only.

### Out of Scope — Source repository changes

- `/Users/goos/MoAI/claude.mo.ai.kr/` is read-only reference material; retiring or refactoring the source skills there is a separate concern in that repository.

### Out of Scope — Sibling skill multilingualization

- The SPEC-HUMANIZE-001 follow-up ("주변 스킬 4종 다국어화" P1 backlog) is a separate SPEC; this SPEC only extends moai-domain-humanize itself.

### Out of Scope — Programming-language rule expansion

- The 16 programming-language rules under `.claude/rules/moai/languages/` are untouched; this SPEC's language axis is natural language only.

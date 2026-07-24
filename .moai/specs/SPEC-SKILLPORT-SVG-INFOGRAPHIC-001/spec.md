---
id: SPEC-SKILLPORT-SVG-INFOGRAPHIC-001
title: "Clean-room author moai-domain-svg-infographic skill"
version: "0.1.3"
status: completed
created: 2026-07-24
updated: 2026-07-24
author: manager-spec
priority: Medium
phase: "v3.1.0 target"
module: "internal/template/templates/.claude/skills/moai-domain-svg-infographic"
lifecycle: spec-anchored
tier: M
tags: "skill, svg, infographic, diagram, clean-room, domain, epic-skillport"
---

# SPEC-SKILLPORT-SVG-INFOGRAPHIC-001 — moai-domain-svg-infographic

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-24 | manager-spec | Initial plan-phase draft. Functionally inspired by skillstead `svg-infographic` (an Apache-2.0 project), sibling to moai-domain-html-report. |
| 0.1.1 | 2026-07-24 | manager-spec | Plan-phase revision. (1) Clean-room pivot: the shipped skill (SKILL.md + references + Node scripts) borrows only the functional capability and is authored independently in moai's own voice/structure — no skillstead attribution (no provenance footer, no NOTICE, no skillstead-tied license); added the clean-room authoring requirement. (2) plan-audit fixes: deleted both residual clarification-gate markers and baked their decisions into prose (D-common), put an explicit Level-2 body token budget in AC-SVG-003 (D3), relabeled REQ-SVG-003 as Ubiquitous (D4), dropped the non-canonical `related_specs` field (D5). Added `tier: M`. |
| 0.1.3 | 2026-07-24 | manager-spec | Plan-phase iteration-3 audit delta fixes (no 4th audit — literal command/wording scopings). (N1, Epic-wide) `license:` follows the house convention: the skill now SHALL ship `license: Apache-2.0` (moai's own declaration — 18 of 28 existing template skills carry it, including the sibling `moai-domain-html-report`; repo LICENSE is Apache-2.0). REQ-SVG-016 gained the affirmative clause and AC-SVG-016(c) was rewritten from a bare `^license:` → 0-matches absence check (which would have false-FAILed a correct implementation) into a `grep -c '^license: Apache-2.0$'` → exactly-1 check plus a frontmatter-scoped skillstead-absence assertion. (N2) AC-SVG-018(d) rescoped: the unresolved frontmatter-block path placeholder and whole-frontmatter `AskUserQuestion\|Agent` grep — which MEASURED 10 of 28 existing skills as false positives from `description`/`when_to_use` prose — is replaced by a literal `allowed-tools`-scoped awk+grep command; REQ-SVG-017's prohibition is correspondingly bound to the `allowed-tools` value only. (N3) §A.1 file-layout provenance split: the `scripts/` + `references/` DIRECTORY names are attributed to `skill-authoring.md` (which documents them), while the script FILENAMES are stated as independently-chosen generic descriptors (that rule documents no filename convention). (N4) AC-SVG-014 gained the CJK-capability carve-out matching REQ-SVG-015's scope split, so a naive CJK-codepoint scan cannot false-FAIL the font-stack and Korean-label content AC-SVG-015 and §D.2 require. (N5) NOTICE scope made REQ↔AC-consistent: REQ-SVG-016 now names both surfaces (skill directories + repository root) and AC-SVG-016(b) asserts both. |
| 0.1.2 | 2026-07-24 | manager-spec | Plan-phase iteration-3 delta fixes. (D1) Resolved the §A.1 contradiction: concrete script PATHS moved OUT of the "borrowed functional properties" list (which now carries capability descriptions only), and a new sentence states that the script filenames and the `scripts/` + `references/` layout follow moai's OWN `skill-authoring.md` directory convention and are therefore not borrowed expression. (D3) Split AC-SVG-002 — it now verifies REQ-SVG-002 (Level-1 listing scope + MEASURED combined character count) only, and the new AC-SVG-018 carries full REQ-SVG-017 clause coverage including the previously-unverified `skills:` YAML-array sub-clause. (D4) REQ-SVG-002 restated: the 1,536-character cap is the COMBINED `description` + `when_to_use` platform ceiling, with the ~100-token Level-1 metadata budget as the design target inside it. (D5) REQ-SVG-013 relabelled `(Unwanted)`; the whole §B REQ block was swept for the same pure-negative-labelled-Ubiquitous class — REQ-SVG-013 was the only remaining instance. (D6) §D summary corrected (5 concepts had been listed for 4 ACs) and an explicit AC↔REQ diagonal-break note added. (P1-B) Mechanized the clean-room guarantee: AC-SVG-016 is now the executable absence-invariant half and the new AC-SVG-019 is the clean-room PROCESS attestation half. AC count 17 → 19. |

## §A. Overview (WHAT / WHY)

### A.1 What

Create a NEW skill, `moai-domain-svg-infographic`, that authors editable SVG architecture/flow infographics using a **layout-before-code** method (compute coordinates and grid arithmetic first, then write SVG), and renders them to 2× PNG via headless Chromium. It is functionally inspired by skillstead's `svg-infographic` skill (an Apache-2.0 project): only the functional CAPABILITY (layout-before-code editable SVG + Chromium PNG render + deterministic source lint + CJK-first wrapping) is borrowed. The SKILL.md body, references, and the bundled Node scripts are authored **clean-room** — independently, in moai's own voice and structure — and reproduce none of skillstead's wording, section structure, code, or file layout. Because only original expression is used, the shipped skill files carry NO skillstead attribution (no provenance footer, no `NOTICE` entry, no skillstead-tied `license:` value). This internal SPEC records the functional inspiration as design rationale; the shipped template files do not.

**House-convention `license:` field (NOT attribution).** The shipped skill DOES carry a `license: Apache-2.0` frontmatter field. This is moai's OWN license declaration following the existing catalog convention — 18 of the 28 current template skills already carry exactly this field, including the named sibling `moai-domain-html-report`, and the repository's own LICENSE is Apache-2.0. The field is unrelated to skillstead, is not a skillstead-tied license value, and does not constitute attribution of any kind. The clean-room prohibition above binds skillstead-TIED license values and skillstead attribution surfaces; it does not and must not be read as a prohibition on moai declaring its own license.

The borrowed functional properties — CAPABILITY DESCRIPTIONS ONLY, no file paths, no layout (the expression, including every filename and directory, is authored independently):
- **Layout-before-code**: numeric grid arithmetic and containment verification precede any SVG authoring — the main defense against render-fix loops.
- **CJK-first font stack** with a ~60% character budget for Korean line wrapping (Korean text budgets at ~60% of the Latin character count per line).
- **Canonical headless-Chromium rendering**: a single canonical renderer that discloses the exact browser executable and version and emits a 2× PNG with IHDR dimension verification.
- **Deterministic source linting**: a source-level lint pass separating hard errors (marker units, references, viewBox sanity) from low-confidence warnings.
- **Opt-in sketch preset** (hand-drawn / sketchnote feel) layered on computed layout.

File layout is NOT borrowed. Its two components have **different provenance bases**, stated separately so neither over-claims:

- **Directory layout (`scripts/` + `references/`) — moai's documented convention.** `.claude/rules/moai/development/skill-authoring.md` § Skill Directory Layout lists `scripts/` as the directory for executable utility scripts and `references/` as the multi-file reference directory. Both directory names are therefore mandated BY moai's own authoring rule. No existing moai skill currently ships a `scripts/` directory, so this SPEC cites the documented rule rather than in-catalog precedent.
- **Script filenames (`render.mjs`, `check-svg.mjs`) — independently chosen generic descriptors.** `skill-authoring.md` documents a directory table only; it prescribes NO filename convention, so the filenames are NOT attributable to that rule. They are generic functional descriptors chosen independently for this skill — `render` for the renderer, `check-svg` for the source linter, with the standard `.mjs` ES-module extension — and carry minimal expressive content under this SPEC's own idea/expression framing: a filename that merely names its own function is not protectable expression.

Together these keep REQ-SVG-016's file-layout prohibition consistent with the plan's M4 filename mandate: the directories follow moai's documented rule and the filenames are original generic choices. Neither is reproduced FROM skillstead.

### A.2 Why

The catalog has no editable static-diagram renderer. `moai-domain-html-report` renders reports (with a tier-gated mermaid CDN exception), and the docs-site renders inline mermaid — but neither produces a standalone, pixel-controlled SVG/PNG image for slides, email, social, or offline distribution, nor precise CJK wrapping. This skill fills the presentation/distribution static-diagram gap.

### A.3 CRITICAL positioning — complementary to mermaid, never a dual-maintenance replacement

This skill is a **presentation/distribution static-diagram renderer that is COMPLEMENTARY TO, NOT A REPLACEMENT FOR, the existing mermaid pipeline** (`moai-domain-html-report` + docs-site `foot.html` CDN mermaid). The SKILL.md MUST document explicit selection rules so the two never dual-maintain the same diagram:

- **Use mermaid for**: diagrams living inside markdown that change often; standard diagram types (flow/sequence/ER/state); anything synced across the 4 locales (text-label reuse).
- **Use svg-infographic for**: static images for slides/email/social/offline; freeform architecture infographics needing pixel control or precise CJK wrapping.

### A.4 Namespace and positioning

- Namespace: `moai-domain-*` (sibling to `moai-domain-html-report`).
- Runtime gate: requires Node 18+ and a headless Chromium browser for lint + PNG render; degrades gracefully when either is absent.

## §B. Requirements (GEARS)

### B.1 Skill existence and structure

- **REQ-SVG-001** (Ubiquitous): The skill shall exist at `internal/template/templates/.claude/skills/moai-domain-svg-infographic/SKILL.md` (template source of truth) with a synced project-local copy.
- **REQ-SVG-002** (Ubiquitous): The skill's Level-1 listing text — the `description` field **and** the `when_to_use` field **combined** — shall summarize the trigger and name the SVG-infographic / static-diagram scope, and shall fit within the platform's **1,536-character COMBINED listing cap** (`maxSkillDescriptionChars`; the cap applies to `description` + `when_to_use` together, not to `description` alone — see `.claude/rules/moai/development/skill-authoring.md`). Inside that hard ceiling, the authoring DESIGN TARGET is the ~100-token Level-1 metadata budget (≈ 400 characters combined). Both figures apply and are not in conflict: ~100 tokens is the target moai authors write to; 1,536 characters is the platform ceiling that must never be exceeded.
- **REQ-SVG-003** (Ubiquitous): The skill shall keep the SKILL.md body at Level-2 (~5K tokens) and place the heavy reference material (archetype layout skeletons, geometry/connector formulas, full icon set, sketch preset) in Level-3 reference files loaded on demand.

### B.2 Layout-before-code method

- **REQ-SVG-004** (Ubiquitous): The skill shall enforce a numeric layout pass — grid arithmetic and containment verification — before any SVG is authored.
- **REQ-SVG-005** (Ubiquitous): The skill shall derive icon/element centers from card geometry (e.g. `cy = card_y + card_h/2`) rather than hand-tuned per-language offsets.
- **REQ-SVG-006** (Ubiquitous): The skill shall apply a CJK-first font stack and budget Korean text at ~60% of the Latin character count per line, editing copy to fit before writing SVG.

### B.3 Rendering and lint (runtime-gated)

- **REQ-SVG-007** (Where Node 18+ and a headless Chromium browser are available): Where the runtime prerequisites are present, the skill shall render the SVG to 2× PNG via the canonical `scripts/render.mjs`, disclosing the exact browser executable and version and verifying IHDR dimensions.
- **REQ-SVG-008** (Where Node 18+ is available): Where Node is present, the skill shall lint the SVG source via `scripts/check-svg.mjs`, treating marker-units / reference / viewBox failures as hard errors and text-overflow / pill-fit as low-confidence warnings.
- **REQ-SVG-009** (When Node is unavailable): When Node 18+ is absent, the skill shall substitute a manual source checklist (no machine-linted label) and continue authoring the editable SVG.
- **REQ-SVG-010** (When no Chromium browser is available): When no headless Chromium is available, the skill shall deliver the SVG only and state the render limitation explicitly.
- **REQ-SVG-011** (Ubiquitous): The skill shall state the Node 18+ / headless-Chromium prerequisite clearly and note that neither is required to install, discover, or author the editable SVG — only to lint and render.

### B.4 Mermaid selection boundary (anti-dual-maintenance)

- **REQ-SVG-012** (Ubiquitous): The skill shall document explicit selection rules distinguishing when to use mermaid (markdown-embedded, frequently-changing, standard types, 4-locale-synced) versus svg-infographic (static distribution images, freeform architecture, pixel/CJK control), so the two pipelines never dual-maintain the same diagram.
- **REQ-SVG-013** (Unwanted): The skill shall NOT replace or deprecate the `moai-domain-html-report` mermaid path or the docs-site CDN mermaid; it is additive and complementary.

> §B label sweep (v0.1.2): the whole §B REQ block was swept for the "pure-negative statement labelled `(Ubiquitous)`" class, not point-fixed. REQ-SVG-013 was the only remaining instance and is relabelled above. REQ-SVG-015, REQ-SVG-016, and REQ-SVG-017 each lead with a positive `shall` obligation and carry a trailing negative clause, so `(Ubiquitous)` remains correct for them; REQ-SVG-009 and REQ-SVG-010 are correctly `(When ...)` event-driven forms.

### B.5 Portability, provenance, format

- **REQ-SVG-014** (Ubiquitous): The skill body, references, and all script comments shall be written in English. This binds PROSE and CODE COMMENTS only — the same scope split REQ-SVG-015 uses: CJK font-family names and Korean example labels are diagram-rendering capability (required by REQ-SVG-006), are expected to appear, and shall not be treated as a language violation.
- **REQ-SVG-015** (Ubiquitous): The skill body, references, and scripts shall remain template-neutral — free of internal moai-adk SPEC IDs, REQ/AC tokens, audit citations, internal dates, and commit SHAs — portable to all 16 supported languages. (CJK-support content — font-stack names, Korean character-budget guidance — is diagram-rendering capability, not an internal token, and is retained.)
- **REQ-SVG-016** (Ubiquitous): The skill body, references, and bundled scripts shall be authored clean-room — independently, in moai's own voice and structure — and shall NOT reproduce skillstead's wording, section structure, code, or file layout verbatim or near-verbatim. Only the functional capability is borrowed; the expression is original. The shipped skill files shall carry NO skillstead attribution — no provenance footer line, no `NOTICE` file (neither in the skill directories nor at the repository root, the two surfaces "no NOTICE file or entry" covers), and no skillstead-TIED `license:` frontmatter value. Conversely, the skill SHALL ship moai's own house-convention `license: Apache-2.0` frontmatter field; that field is moai's own license declaration (matching the repository LICENSE and the existing template-skill catalog), is unrelated to skillstead, and does not constitute attribution. An implementation that omits the `license: Apache-2.0` field violates this requirement.
- **REQ-SVG-017** (Ubiquitous): The skill frontmatter shall follow MoAI conventions — `allowed-tools` as a comma-separated string holding exactly `Read, Write, Edit, Grep, Glob, Bash` **in that order** (Bash required for the Node render/lint scripts; the order matches the sibling `moai-domain-html-report` and is pinned so AC-SVG-018(d) can verify by exact-line match), any `skills:` as a YAML array, all `metadata` values quoted — and the `allowed-tools` value shall NOT list AskUserQuestion or Agent. This prohibition binds the `allowed-tools` value ONLY: it does not bind narrative frontmatter fields (`description`, `when_to_use`), where naming a tool or an agent type in prose is legitimate and common across the existing catalog.

## §C. Exclusions

The following are **out of scope**. Routing rationale: the mermaid pipeline already owns markdown-embedded and locale-synced diagrams; provisioning runtimes is the user's environment concern; interactive/animated output is a different product.

### Out of Scope — replacing the mermaid pipeline
- svg-infographic does not migrate, replace, or deprecate any existing mermaid diagram in `moai-domain-html-report` or the docs-site; the selection rules keep the two complementary.
- Bulk conversion of existing markdown mermaid diagrams to SVG is not performed.

### Out of Scope — runtime provisioning
- The skill does not install Node, Chromium, or any package manager; it detects presence and degrades gracefully.
- Bundling a Chromium binary or an npm dependency tree into the template is not done (Node stdlib only).

### Out of Scope — interactive / animated output
- Animated SVG, interactive JS-driven diagrams, and live dashboards are not produced — output is a static SVG plus a 2× PNG raster.

### Out of Scope — docs-site design tokens
- The skill does not consume or modify the docs-site FROZEN `moai-brand.css` or Claude Warm Editorial tokens; its palette is self-contained.

## §D. Acceptance Criteria (summary)

Full scenarios in `acceptance.md`. Summary:

- AC-SVG-001..003: skill exists template-first + local-synced; Level-1 listing scope + MEASURED combined character count; progressive-disclosure structure (Level-3 references + scripts).
- AC-SVG-004..006: layout-before-code method, geometry-derived centers, CJK 60% budget documented.
- AC-SVG-007..011: render + lint recipes present; graceful degradation for missing Node / missing Chromium; runtime prerequisite stated as render/lint-only.
- AC-SVG-012..013: mermaid-vs-svg selection rules present; no replacement of the mermaid pipeline.
- AC-SVG-014..015: English body + scripts (with the CJK-capability carve-out — font-stack names and Korean sample labels are required capability, not a language violation); template neutrality (CI guard clean, CJK content retained).
- AC-SVG-016 + AC-SVG-019: clean-room guarantee, split into a MECHANICAL absence-invariant half (AC-SVG-016 — vendor-token grep, no NOTICE at either surface, house-convention `license: Apache-2.0` present exactly once with no skillstead reference in frontmatter) and a PROCESS attestation half (AC-SVG-019).
- AC-SVG-017: embedded-FS rebuild. AC-SVG-018: frontmatter-convention format (full REQ-SVG-017 clause coverage).

AC↔REQ numbering note: the AC-N ↔ REQ-N diagonal holds only through AC-SVG-016 and **breaks at 17**. AC-SVG-017 traces to **REQ-SVG-001** (embedded-FS rebuild of the skill this REQ places), AC-SVG-018 traces to REQ-SVG-017, and AC-SVG-019 is the second half of REQ-SVG-016. The `acceptance.md` §D matrix is the authoritative mapping — do not infer a REQ from an AC number.

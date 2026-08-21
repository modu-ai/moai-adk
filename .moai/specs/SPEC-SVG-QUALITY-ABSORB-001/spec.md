---
id: SPEC-SVG-QUALITY-ABSORB-001
title: "Absorb the diagram-design quality layer into moai-domain-svg-infographic"
version: 0.1.0
status: draft
created: 2026-08-22
updated: 2026-08-22
author: lane-7
priority: medium
phase: "v3.2"
module: .claude/skills/moai-domain-svg-infographic
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "skill, svg, diagram, accessibility, quality, absorption"
related_specs: []
---

# SPEC-SVG-QUALITY-ABSORB-001

## §1 Problem / Motivation

`moai-domain-svg-infographic` produces diagrams from numeric layout, but six
quality layers that make a diagram trustworthy are thin or missing. A survey of
`cathrynlavery/diagram-design` v2.6.1 (MIT) classified its 412 files into
absorb / reference / reject, and six items came back as immediate absorptions —
rules, not catalogue.

The distinction matters. The demo diagrams that motivated this work share a look
(warm paper ground, coral accent, serif titles, mono sub-labels, orthogonal
connectors) that is **not** produced by a type catalogue. It is produced by a
semantic-role skin, typographic role separation, and a one-accent discipline.
Absorbing 39 type specifications would fight this skill's mermaid routing and its
"one diagram, one home" rule; absorbing the rules behind the look does not.

One of the six is a plain gap rather than a thinness: **the skill has no
accessibility contract at all.** Measured: `aria`, `role=`, `<title>`, and
`<desc>` occur zero times across `SKILL.md`, all three references, and both
scripts. Every SVG this skill emits is currently unreadable to a screen reader.

## §2 Scope

**In Scope** — the six A-items, expressed as rules in the existing files:

| Item | Content | Lands in |
|---|---|---|
| A-1 | Connector rules, 6 enforced: orthogonal elbow `r=8`; 6–10px mask gap; no overlap, bridge/hop on cross; attachment fan `L·k/(N+1)` with ≥12px spacing; no routing behind a non-endpoint node; mask must not overlap a following node | `references/authoring.md` §2 |
| A-2 | Per-type complexity budget table (node ceilings by diagram type) | `SKILL.md` + `references/archetypes.md` |
| A-3 | Accessible-SVG contract: `role`, `aria-labelledby`, prefixed IDs, `<title>` as first child, `<desc>` describing content | `SKILL.md` + `references/authoring.md` |
| A-4 | 14 "AI slop" anti-patterns | `SKILL.md` § Red Flags |
| A-5 | Four output dials — format / size / detail / audience | `SKILL.md` § Frame |
| A-6 | Semantic-role skin tokens + light↔dark inversion rule | `references/authoring.md` §3 palette restructure |

**In Scope — the mermaid bypass exception list.** A short, evidence-gated list of
diagram types where the absorbed rules beat mermaid badly enough to justify
routing to this skill, **on the image-output path only**.

### Out of Scope — the 39-type catalogue

- Absorbing the per-type specifications would collide with this skill's mermaid
  routing and with its "one diagram, one home" rule.
- The look being reproduced comes from the skin and the discipline, not the
  catalogue, so the catalogue is not what makes the output good.

### Out of Scope — anything requiring an external asset at view time

- Google Fonts CDN is rejected outright: this skill's no-external-asset contract
  is what makes its output readable offline and in email.
- HTML variants, motion modes, and the terminal sub-skin stay out — the static
  SVG + PNG contract is unchanged.

### Out of Scope — the docs-site text-sync path

- Diagrams whose labels are kept in sync across the four locales stay on
  mermaid. The exception list may not touch that path, and nothing here changes
  how locale-synced text is authored.

### Out of Scope — deferred to sibling cards

- B-7 verifier geometry extension (`check-svg.mjs`), B-8 profiles, B-9
  drawio/mermaid importers. B-7 depends on A-1 landing first, since it is the
  executable form of these rules.

## §3 Requirements (GEARS)

Acceptance criteria live in `acceptance.md` (Tier M): the scope reaches ten files
once template mirrors are counted, and thirteen criteria exceed what Tier S
budgets — tiering up is the prescribed response to that, not trimming the
criteria until they fit.

### REQ-1 — Connector rules (A-1)

`authoring.md` §2 SHALL state the six connector rules as numeric constraints,
each with the formula or tolerance a checker could later assert.

### REQ-2 — Complexity budgets (A-2)

The skill SHALL carry a per-type node ceiling table, and the existing archetypes
SHALL each map to a budget.

### REQ-3 — Accessible-SVG contract (A-3)

Emitted SVG SHALL carry a resolving accessible name: `role`, `aria-labelledby`
pointing at a document-unique prefixed ID, `<title>` as the first child of
`<svg>`, and a `<desc>` describing the content rather than repeating the title.

### REQ-4 — AI-slop anti-patterns (A-4)

§ Red Flags SHALL carry the 14 anti-patterns, each phrased as an observable
symptom rather than a style preference.

### REQ-5 — Output dials (A-5)

The Frame section SHALL define four dials — format, size, detail, audience —
with the allowed values for each and the default the skill assumes when the
caller does not say.

### REQ-6 — Semantic-role skin tokens (A-6)

The palette section SHALL be restructured around semantic roles rather than raw
hex, with a light↔dark inversion rule.

### REQ-7 — Mermaid bypass exception list (evidence-gated)

The exception list SHALL name only types for which a sample comparison was
actually produced, and SHALL record, per type, the mermaid output and the
absorbed-rule output that justified it.

### REQ-8 — Distribution and attribution

Every changed file under `.claude/skills/` SHALL have its
`internal/template/templates/` mirror updated in the same commit, and the
absorbed rules SHALL carry attribution to `cathrynlavery/diagram-design` v2.6.1
(MIT). MIT into Apache-2.0 is compatible and carries an attribution duty;
omitting it is a licence defect, not a style miss.

## §4 Evidence

- `.moai/reports/diagram-design-absorption/survey.md` — the 412-file survey
- `.moai/reports/diagram-design-absorption/diagram-design-absorption-20260822.md`
  — the A/B/C/D classification this SPEC implements the A tier of
- Source clone: `/tmp/diagram-design` (v2.6.1, MIT, LICENSE present)
- A-3 absence measured this session: `aria|role=|<title>|<desc>` → 0 matches
  across `SKILL.md`, `references/*.md`, and `scripts/*.mjs`

## §5 Constraints / Non-Goals

- **The exception list is capped by evidence, not by ambition.** REQ-7 admits a
  type only after a sample pair exists. An empty list is an acceptable outcome.
- **No claim that absorbed rules beat mermaid generally.** The comparison is
  per-type and image-path-only; mermaid remains the default and wins ties.
- **The samples are judged by a human.** "Better" here is a design judgement the
  SPEC does not automate; what the SPEC requires is that the judgement be made
  against two artifacts that actually exist.
- Absorbed text is rewritten, not copied: the source is MIT and compatible, but
  the two skills differ in contract (no external assets, CJK-first budget), so a
  verbatim lift would import rules this skill contradicts.

## §6 HISTORY

- 0.1.0 (2026-08-22) — initial draft. Scope is the A tier of the absorption
  survey plus an evidence-gated mermaid bypass list.

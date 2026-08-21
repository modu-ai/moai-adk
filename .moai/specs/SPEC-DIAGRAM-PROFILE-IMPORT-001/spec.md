---
id: SPEC-DIAGRAM-PROFILE-IMPORT-001
title: "Absorb diagram-design profile persistence and drawio/mermaid importers as opt-in skill references"
version: "0.1.0"
status: draft
created: 2026-08-22
updated: 2026-08-22
author: manager-spec (card t167)
priority: P1
phase: "v3.1.3 target"
module: internal/template/templates/.claude/skills
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "template, skill, design-dna, svg-infographic, profile, importer, absorption"
related_specs: [SPEC-SVG-QUALITY-ABSORB-001]
depends_on: [SPEC-SVG-QUALITY-ABSORB-001]
---

# SPEC-DIAGRAM-PROFILE-IMPORT-001

## §1 Problem / Motivation

Two gaps remain after SPEC-SVG-QUALITY-ABSORB-001 (t165) landed the
diagram-design quality layer. Both come from the same upstream survey's B tier
(`cathrynlavery/diagram-design` v2.6.1, MIT — `.moai/reports/t165/upstream/`),
which that SPEC explicitly deferred to sibling cards:

**(a) No profile persistence in design-dna.** The skill extracts a Design DNA
JSON (Phase 2) and generates from it (Phase 3), but every extraction is
one-shot: there is no way to save a completed profile under a name and reuse it
in a later session, so a team re-derives the same reference design each time.
Measured this session: `grep -c "profile"` across
`moai-domain-design-dna/SKILL.md` + both references → **0 matches**. The
persistence concept is absent, not thin.

**(b) No migration path in svg-infographic.** The skill's Step 0 routing sends
existing diagrams to mermaid and its "one diagram, one home" rule forbids
shipping two forms — but there is no documented way to MIGRATE an existing
mermaid or draw.io diagram into this skill's numeric-layout authoring path.
Users with an existing diagram corpus have no on-ramp: measured the same way,
`grep -ci "drawio\|import"` across the skill → **0 matches**. Demand is real
(existing-diagram migration), and doing it ad hoc produces exactly the
dual-maintenance state Step 0 exists to prevent.

Both absorptions are **reference-shaped, not script-shaped**: the upstream
extractors are 78 KB of Python; this SPEC absorbs their structure and
discipline (IR extraction order, untrusted-source handling, fidelity ledger)
as documented procedure inside the existing skills, matching how t165 absorbed
rules rather than catalogue.

## §2 Scope

**In Scope — group (a), design-dna strengthening** (2 template files):

| Surface | Change |
|---|---|
| `moai-domain-design-dna/references/diagram-profiles.md` (new) | The profile mechanism: named DNA-JSON snapshots, project-scoped storage, marker-first resolution, slug grammar, metadata header, load-time schema check + backfill, save/load write safety |
| `moai-domain-design-dna/SKILL.md` | A short additive section pointing at the reference (Phase 2 can save; Phase 3 can start from the active profile) |

**In Scope — group (b), svg-infographic opt-in importers** (3 template files):

| Surface | Change |
|---|---|
| `moai-domain-svg-infographic/references/import-mermaid.md` (new) | Mermaid → IR → numeric-layout migration procedure, opt-in |
| `moai-domain-svg-infographic/references/import-drawio.md` (new) | draw.io → IR → numeric-layout migration procedure (4 container shapes), opt-in |
| `moai-domain-svg-infographic/SKILL.md` | Bundled-references table rows for both importers, marked opt-in; default flow untouched |

**In Scope — cross-cutting**: attribution, Template-First mirrors, catalog
regeneration, 16-language neutrality. The two groups are **separable**: either
milestone may land and close independently of the other.

The (a)/(b) groups carry no shared file. The only file another in-flight card
touches is `moai-domain-svg-infographic/SKILL.md`, whose t167 delta (table
rows) sits in a different section from t166's expected linting-section delta —
see §5 and plan.md §D for the non-overlap analysis.

### Out of Scope — shipped extractor scripts

- Porting or vendoring `drawio_extract.py` (31 KB) / `mermaid_extract.py`
  (47 KB). The importers are documented procedure — the reading agent parses
  the source directly; the upstream scripts inform the procedure's ORDER and
  DISCIPLINE, not its code.

### Out of Scope — icon normalization pipeline (B-10, deferred)

- Upstream's build-icons pipeline (vendored Tabler/Simple/logz/Devicon sets,
  ~102 SVGs, THIRD_PARTY_LICENSES.md bookkeeping) is **deferred**. The skill's
  existing 12-glyph set (`references/authoring.md` §5) is already 24×24 stroke
  `currentColor` — the normalization convention is natively present, and the
  no-external-asset contract (t165 REQ-9) plus license bookkeeping only pay
  off when imported diagrams genuinely need brand icons. Importer references
  instruct mapping source icons onto the existing set (or omitting them). A
  future card may adopt B-10 with the THIRD_PARTY notice if that need is
  demonstrated.

### Out of Scope — verifier changes (t166, in-flight on lane-6)

- `scripts/check-svg.mjs`, `scripts/render.mjs`, and `scripts/fixtures/` are
  NOT touched by this SPEC. t166 (lane-6) extends the checker with 3 geometry
  checks + self-test fixtures; this SPEC's importers are a CONSUMER of that
  checker, written against the quality rules both enforce (authoring.md §2.5).

### Out of Scope — default-flow and routing changes

- The six-step workflow, the Step 0 routing table, the four output dials, and
  both skills' `description`/`when_to_use` frontmatter are unchanged. The
  importers are a migration path, not part of the authoring mainline; profile
  persistence does not alter the analyze→generate flow when no profile exists.

## §3 Requirements (GEARS)

Acceptance criteria live in `acceptance.md` (Tier M: 15 criteria). Group (a) =
REQ-1..REQ-5, group (b) = REQ-6..REQ-12, cross-cutting = REQ-13..REQ-15.

### Group (a) — marker-first profile persistence (design-dna)

**REQ-1 (profile mechanism).** The design-dna skill shall document, in a new
`references/diagram-profiles.md`, a profile mechanism that persists a completed
Design DNA JSON as a named snapshot and lets a later generation start from that
snapshot instead of re-extracting the reference.

**REQ-2 (marker-first resolution).** **Where** a project carries a profile
marker at its root, the skill shall resolve the active profile marker-first —
the marker's slug wins over any other lookup — and shall read the referenced
snapshot **in place** rather than copying it into the skill's storage.
**Where** no marker exists, the skill shall not guess a profile; extraction
proceeds profile-less exactly as today. (Upstream ADR 0006 rationale, restated:
read-in-place avoids copy races and survives skill reinstallation.)

**REQ-3 (storage and slug grammar).** Profiles shall live in a project-scoped
directory outside the skill's own directory, addressed by a strict slug
(lowercase `a-z`, `0-9`, single hyphens; no card ids, no paths), and each
snapshot shall carry a small metadata header naming its origin reference and
schema version. (Exact directory: plan.md §F M1 records the decision and the
rejected alternatives — in-skill storage dies on skill update; home-global-only
is invisible to the repository.)

**REQ-4 (load-time validation).** **When** a profile is loaded, the skill shall
validate it against the three-dimension schema of `references/dna-schema.md`
and shall backfill missing optional fields with an explicit "not observed"
value before use; the skill shall not silently drop a dimension, and shall not
invent values the snapshot does not carry.

**REQ-5 (write safety).** **When** saving a profile whose slug already exists,
the skill shall ask before overwriting, and **When** a save completes, the skill
shall verify the write by re-reading the saved file rather than assuming it.

### Group (b) — drawio/mermaid importers (svg-infographic, opt-in)

**REQ-6 (opt-in importer references).** The svg-infographic skill shall
document two opt-in reference files — `references/import-mermaid.md` and
`references/import-drawio.md` — each describing the migration pipeline
source → intermediate representation (nodes, edges, labels, groups) → four
output dials → numeric layout pass → authored SVG → lint. The skill's default
authoring flow shall remain unchanged: the six-step workflow and the Step 0
routing table govern every non-import request exactly as before.

**REQ-7 (source untrusted).** The importer references shall treat source text
as untrusted data: the imported diagram is a starting sketch only, and source
coordinates, colors, fonts, theme, and layout shall never be carried into the
output. The migrated diagram shall be re-authored through the numeric layout
pass and validated against the same quality rules as a from-scratch diagram.

**REQ-8 (drawio container handling).** The drawio importer shall document
decoding before parsing — the four container shapes (plain XML,
base64-deflate-compressed `mxfile`, `.drawio.png`-embedded, and
`.drawio.svg`-embedded) — and shall state that compressed bytes are never read
as structure.

**REQ-9 (fidelity ledger).** **When** an import completes, the deliverable
shall include a fidelity ledger recording what was **merged, collapsed, or
dropped** relative to the source, and how any complexity-budget overrun was
resolved (zone past the band rule; split beyond the `detail` ceiling — the
`faithful ≤ 24 in labelled bands` / `balanced ≤ 12` / `simplified ≤ 7` budgets
of SKILL.md govern imports unchanged).

**REQ-10 (bulk replace — one home).** **When** an import delivers the migrated
diagram, the source mermaid block or drawio file shall be replaced or removed
**in the same change**; the migrated diagram and its source shall never
coexist. This extends the skill's existing "one diagram, one home" rule
(SKILL.md Step 0) to the migration path without exception.

**REQ-11 (verifier-compat by construction).** The importer references shall
state the import-relevant numeric constraints of `references/authoring.md`
§2.5 as authoring obligations on the imported diagram — elbow `r = 8`,
label-mask gap 6–10 units, attach-point spacing ≥ 12 units, and the paint-order
label-placement rule (C6) — so importer output satisfies both the t165
connector rules and the t166 geometry checks by construction; and the fidelity
ledger shall record a `check-svg.mjs` pass (zero errors) for every migrated
diagram. t166's exact diagnostic codes are not yet landed on this branch —
see §5.

**REQ-12 (discoverability, composing with routing).** SKILL.md's
bundled-references table shall carry one row per importer reference, marked
opt-in, and the importer entry point shall compose with — not contradict — the
Step 0 routing table: the routing table still decides where a diagram lives
going forward; the importers are the explicit migration path a caller invokes
when the deliverable is an image.

### Group (c) — cross-cutting

**REQ-13 (attribution).** Each new reference file shall carry an attribution
line naming `cathrynlavery/diagram-design` v2.6.1 (MIT) for the absorbed
patterns (marker-first profiles; importer pipeline + fidelity ledger).
Absorbed content is restated, not copied — this SPEC adds no vendored material,
so no THIRD_PARTY license file is required by it.

**REQ-14 (Template-First + catalog).** All edits shall land in
`internal/template/templates/` first, the local `.claude/skills/` mirrors shall
be updated in the same commit, and **Where** a skill's root `SKILL.md` changes,
the regenerated `internal/template/catalog.yaml` hash entry shall land in the
same commit. (Verified this session: the catalog hashes only the root
`SKILL.md` per skill — `internal/template/catalog_hash_norm.go` — so the
reference-file additions alone would not move the hash, but both groups also
touch SKILL.md, so regeneration is required either way.)

**REQ-15 (neutrality).** The new template files shall observe the 16-language
neutrality contract and the template-neutrality content classes: no SPEC IDs,
no card ids, no lane names, no moai-adi-internal references, and no internal
dates beyond the skill metadata's own version fields.

## §4 Evidence

All facts below were measured on this branch (HEAD = origin/release/v3.1.3
merged; t165's absorbed layers present, t166 NOT landed):

- `grep -c "profile"` over design-dna `SKILL.md` + `references/*.md` → 0
  matches (group-(a) absence).
- `grep -ci "drawio\|import"` over svg-infographic `SKILL.md` +
  `references/*.md` → 0 matches (group-(b) absence).
- `check-svg.mjs` (609 lines): diagnostic codes present are SVG001–SVG003,
  SVG010–SVG011, SVG020–SVG021, SVG030–SVG031, SVG040, SVG060–SVG064;
  `grep "paint-order\|attachment\|mask.*gap"` → 0 matches (t166's geometry
  checks absent on this branch — consistent with lane-6 in-flight).
- `authoring.md` §2.5 (line 152) carries the six mandatory connector rules
  with the exact numeric constraints REQ-11 restates: C1 `r = 8` (floor 6),
  C2 mask gap 6–10, C3 separation ≥ 12 + bridge `rHop = 5`, C4 fan
  `L·k/(N+1)` with spacing ≥ 12, C5 dashed-transit exception, C6
  mask-vs-later-node + paint order.
- `internal/template/catalog.yaml` entries for both skills hash the skill
  directory path; `internal/template/catalog_hash_norm.go` documents that
  skill-directory entries hash only the root `SKILL.md`/`skill.md`.
- Upstream patterns: `.moai/reports/t165/upstream/survey.md` (profiles.md row
  line 36 — marker-first resolution, slug grammar, schema check + default
  backfill; import rows lines 38–39, 98–99 — extract-first order, untrusted
  source, fidelity ledger, never carry source coordinates/colors/fonts;
  ADR 0006 line 90) and `absorption-verdict.md` (B-8/B-9/B-10
  classification). MIT license text: `UPSTREAM-LICENSE` in the same
  directory.
- t165's deferral of these items: `SPEC-SVG-QUALITY-ABSORB-001` §2 "Out of
  Scope — deferred to sibling cards".

## §5 Constraints / Non-Goals

- **t166 interface compatibility (lane-6, in-flight).** This SPEC writes the
  importer contract against the QUALITY RULES both sides enforce
  (authoring.md §2.5, verified present), NOT against t166's exact diagnostic
  surface — which is not on this branch. Any t166-specific interface fact
  (code numbers, flag names, fixture names) is marked "verify at run-phase
  against lane-6's landed code"; acceptance.md AC-VERIFY-001 is that
  verification.
- **File non-overlap with t166 (verified + stated).** t167 touches
  design-dna SKILL.md + references/, and svg-infographic SKILL.md (table rows
  only) + references/. t166 touches `scripts/check-svg.mjs` +
  `scripts/fixtures/`. No shared file except `moai-domain-svg-infographic/
  SKILL.md`, where t167's delta (bundled-references table) and t166's expected
  delta (Linting section) sit in different sections; integration resolves any
  conflict by re-applying the table rows. Run phase re-reads the file at M3
  start.
- **No default-flow change.** If a reader cannot tell from the default
  sections that anything changed, group (b) has succeeded; the importers live
  in references and are reachable only by explicit migration intent.
- **No verbatim upstream copy.** Restatement only (t165 §5 precedent): the two
  skills' contracts (no external assets, CJK-first budget, one-home rule)
  contradict several upstream defaults, so a verbatim lift would import rules
  this repo rejects.
- **Phase label.** `"v3.1.3 target"` per the card. (Noted for the
  orchestrator, not acted on here: sibling SPEC-SVG-QUALITY-ABSORB-001 carries
  `phase: "v3.2"` while its landed artifacts are inside origin/release/v3.1.3
  — a label inconsistency in the sibling card, reported in the final
  response.)
- **Icon normalization is decided, not open.** Deferred (§2 Out of Scope);
  importer references must state the map-to-existing-12-glyph rule so the
  deferral is actionable rather than silent.

## §6 HISTORY

- 0.1.0 (2026-08-22) — initial draft. Group (a) B-8 profile persistence,
  group (b) B-9 importers, cross-cutting attribution/template/neutrality;
  B-10 icon pipeline deferred with rationale. Depends on
  SPEC-SVG-QUALITY-ABSORB-001 (completed).

---
id: SPEC-DIAGRAM-PROFILE-IMPORT-001
title: "Absorb diagram-design profile persistence and drawio/mermaid importers as opt-in skill references"
version: "0.2.0"
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
Measured this session with a discriminating grep (the marker-profile
vocabulary, whole directory):
`grep -rniE "\.design-dna|profile[ -]?(marker|snapshot)|active profile marker" moai-domain-design-dna/`
→ **0 matches**. The persistence surface is absent, not thin. (The bare word
"profile" is NOT the discriminator: it appears 6 times in the instance sense —
a Design DNA profile — at SKILL.md:69/147/185, dna-schema.md:4/115,
effects-implementation.md:102; none names a save/load/marker mechanism.)

**(b) No migration path in svg-infographic.** The skill's Step 0 routing sends
existing diagrams to mermaid and its "one diagram, one home" rule forbids
shipping two forms — but there is no documented way to MIGRATE an existing
mermaid or draw.io diagram into this skill's numeric-layout authoring path.
Users with an existing diagram corpus have no on-ramp: measured this session,
`grep -rn "drawio" moai-domain-svg-infographic/` → **0 matches** (whole
directory, case-sensitive; `grep -rn "import-mermaid\|import-drawio"` over the
same tree → **0 matches**). (The substring "import" is NOT the discriminator:
`grep -ci "drawio\|import"` returns 2 — both the word "important", svg
SKILL.md:353 and authoring.md:352.) Demand is real (existing-diagram
migration), and doing it ad hoc produces exactly the dual-maintenance state
Step 0 exists to prevent.

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
| `moai-domain-svg-infographic/SKILL.md` | Bundled-references table rows for both importers, marked opt-in; plus ONE amended sentence in Step 0 (the no-migration default gains its caller-invoked exception, REQ-13) — the sole default-section change |

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
  both skills' `description`/`when_to_use` frontmatter are unchanged. The sole
  default-section edit is REQ-13's one-sentence amendment of the Step-0
  no-migration default. The importers are a migration path, not part of the
  authoring mainline; profile persistence does not alter the analyze→generate
  flow when no profile exists.

## §3 Requirements (GEARS)

Acceptance criteria live in `acceptance.md` (Tier M: 16 criteria). Group (a) =
REQ-1..REQ-5, group (b) = REQ-6..REQ-13, cross-cutting = REQ-14..REQ-16 — 16
requirements total, at the Tier M ceiling.

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
(SKILL.md Step 0) to the migration path without exception. Three boundaries
bind the obligation:

- **No auto-discovery.** The importer shall act only on a caller-named source;
  it shall never scan a project for diagrams to migrate.
- **Replacement requires caller intent.** The source is deleted because the
  caller invoked a migration, never as a side effect of reading the source.
- **Non-owned sources.** **Where** the source lives outside the caller's write
  scope (another repository's diagram used as a reference), the source text
  remains untrusted exactly as in REQ-7, the untouched source is left alone,
  and the import is recorded in the fidelity ledger as a **derivation**, not a
  migration — the one-home obligation scopes to diagrams the caller maintains.

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

**REQ-13 (the no-migration default and its one exception).** The svg-infographic
SKILL.md Step-0 sentence "Nothing here migrates, rewrites, or deprecates an
existing mermaid diagram" (SKILL.md:43-44) currently denies the migration path
REQ-6/REQ-12 add to the same file; the amendment shall reconcile them. At
run-phase that sentence — and only that sentence — shall be amended to state
that the no-migration default stands **except** for a caller-invoked import
through the opt-in importer references, in which case the one-home rule
(SKILL.md:65-67 — "either replace it outright … or leave it alone") governs
with REQ-10's same-change replacement. The sentence's other claims — additive
to mermaid, never a replacement, no diagram in both forms — remain standing;
the opt-in bundle-table row (REQ-12) is the pointer that marks the exception
at the site of the default.

### Group (c) — cross-cutting

**REQ-14 (attribution).** Each new reference file shall carry an attribution
line naming `cathrynlavery/diagram-design` v2.6.1 (MIT) for the absorbed
patterns (marker-first profiles; importer pipeline + fidelity ledger).
Absorbed content is restated, not copied — this SPEC adds no vendored material,
so no THIRD_PARTY license file is required by it.

**REQ-15 (Template-First + catalog).** All edits shall land in
`internal/template/templates/` first, the local `.claude/skills/` mirrors shall
be updated in the same commit, and **Where** a skill's root `SKILL.md` changes,
the regenerated `internal/template/catalog.yaml` hash entry shall land in the
same commit. (Verified this session: the catalog hashes only the root
`SKILL.md` per skill — `internal/template/catalog_hash_norm.go` — so the
reference-file additions alone would not move the hash, but both groups also
touch SKILL.md, so regeneration is required either way.)

**REQ-16 (neutrality).** The new template files shall observe the 16-language
neutrality contract and the template-neutrality content classes: no SPEC IDs,
no card ids, no lane names, no moai-adi-internal references, and no internal
dates beyond the skill metadata's own version fields.

## §4 Evidence

All facts below were measured on this branch (HEAD = origin/release/v3.1.3
merged; t165's absorbed layers present, t166 NOT landed):

- Group-(a) absence, discriminating grep (whole directory):
  `grep -rniE "\.design-dna|profile[ -]?(marker|snapshot)|active profile marker" moai-domain-design-dna/`
  → **0 matches**. The naive `grep -c "profile"` does NOT discriminate — it
  returns 6 instance-sense matches (SKILL.md:69/147/185, dna-schema.md:4/115,
  effects-implementation.md:102); re-measured this session, line numbers
  confirmed.
- Group-(b) absence, discriminating greps (whole directory):
  `grep -rn "drawio" moai-domain-svg-infographic/` → **0 matches**;
  `grep -rn "import-mermaid\|import-drawio" moai-domain-svg-infographic/` →
  **0 matches**. The naive `grep -ci "drawio\|import"` does NOT discriminate —
  it returns 2, both the word "important" (svg SKILL.md:353,
  authoring.md:352); re-measured this session.
- `check-svg.mjs` (609 lines): diagnostic codes present are SVG001–SVG003,
  SVG010–SVG011, SVG020–SVG021, SVG030–SVG031, SVG040, SVG050 (parser
  failure), SVG060–SVG064; `grep "paint-order\|attachment\|mask.*gap"` → 0
  matches (t166's geometry checks absent on this branch — consistent with
  lane-6 in-flight).
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
- **Default-flow change is confined to one sentence.** The only default-section
  edit group (b) makes is the REQ-13 amendment of the Step-0 no-migration
  sentence; the six-step workflow, the routing table, and the dials are
  untouched. Outside that sentence, a reader cannot tell from the default
  sections that anything changed — the importers live in references and are
  reachable only by explicit migration intent.
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

- 0.2.0 (2026-08-22) — plan-audit iter-1 fixes. D1: §1/§4 gap evidence
  restated with discriminating greps that actually return 0 (the naive
  `grep -c "profile"` / `grep -ci "drawio\|import"` return 6 and 2
  respectively — instance-sense and "important" false positives; both
  re-measured and recorded). D2: new REQ-13 reconciles the Step-0 no-migration
  sentence (SKILL.md:43-44) with the importer path via a caller-invoked
  exception; cross-cutting renumbered REQ-14..16. D3: SVG050 added to the §4
  enumeration. D5: REQ-10 gains the no-auto-discovery / caller-intent /
  non-owned-source boundaries. 16 REQ / 16 AC, both at the Tier M ceiling.
- 0.1.0 (2026-08-22) — initial draft. Group (a) B-8 profile persistence,
  group (b) B-9 importers, cross-cutting attribution/template/neutrality;
  B-10 icon pipeline deferred with rationale. Depends on
  SPEC-SVG-QUALITY-ABSORB-001 (completed).

---
id: SPEC-SKILL-GALLERY-BENCH-001
title: "SkillStead TypePack 9-Form Benchmark of moai-domain-svg-infographic + Producible-Diagram-Types Docs Emphasis"
version: "0.1.2"
status: completed
created: 2026-08-25
updated: 2026-08-25
author: manager-spec
priority: P2
phase: "v3.1.3"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "benchmark, svg-infographic, skillstead, typepack, docs-site, readme, 4-locale"
tier: M
---

# SPEC-SKILL-GALLERY-BENCH-001 — SkillStead TypePack 9-Form Benchmark + Docs Emphasis

## §A Overview

### Problem

The third-party skill catalog `github.com/kyungseo/skillstead` (Apache-2.0) publishes a
verified "TypePack" gallery for its `svg-infographic` skill: nine bounded diagram types,
each displayed with a receipt badge proving the artifact passed lint, layout, and
typography gates. Our counterpart skill, `moai-domain-svg-infographic`, documents four
layout archetypes (A1 architecture stack, A2 left-to-right flow, A3 side-by-side
comparison, A4 hierarchy tree) but has never been measured against an external catalog
of concrete diagram forms. Public surfaces (docs-site, README) state only that the skill
produces "editable SVG technical infographics" — they do not tell a user which diagram
kinds it can actually produce, and any such claim would currently be unverified.

### Goal

Measure, not assert: run the nine SkillStead TypePack forms as equivalent-form
generation briefs against `moai-domain-svg-infographic` at its current `main` state,
observe the skill's own deterministic gates per artifact (source lint + 2x PNG render),
classify every failure, and then — and only then — add a "producible diagram types"
emphasis to docs-site and README (4 locales, ko canonical) grounded exclusively in the
measured artifacts.

### Benchmark provenance (verified by the orchestrator, 2026-08-25)

- Source: `github.com/kyungseo/skillstead`, Apache-2.0; skill `svg-infographic` v0.11.0.
- Their pipeline: layout computation before drawing, structured SVG authoring, source
  lint, dimension-verified 2x PNG export, Korean/CJK-ready.
- The nine TypePacks (the benchmark set): `approval-gate`, `before-after`,
  `cards-kpi-grid`, `decision-matrix`, `layer-stack`, `nested-scope`, `process-flow`,
  `roadmap-timeline`, `topology-component`.
- "Equivalent form" reading (card wording: "각 형태(또는 동등 형태)"): an artifact
  qualifies when it communicates the same information structure as the named TypePack
  form, even where visual treatment differs (e.g. nested rectangles instead of
  concentric rings). The verdict record must name any such deviation (REQ-SGB-006).

### Our counterpart's observable gates (reused verbatim from the skill)

| Gate | Command | Pass condition |
|------|---------|----------------|
| G1 artifact | — | `artifacts/<form>.svg` exists under the evidence dir |
| G2 lint | `node .claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs <svg>` | exit 0, zero errors; every warning triaged and recorded |
| G3 render | `node .claude/skills/moai-domain-svg-infographic/scripts/render.mjs <svg> --out <png>` | exit 0; PNG `IHDR` dimensions match the 2x target; browser executable + version disclosed |

These are the skill's own delivery gates (`SKILL.md` sections "Linting the source" and
"Rendering and verifying the PNG"); the benchmark adds no new gate machinery.

## §B Terminology

| Term | Meaning |
|------|---------|
| TypePack form | One of the nine named SkillStead diagram types; the benchmark's unit of work |
| Equivalent form | An artifact expressing the same information structure as the TypePack form, with any deviation named |
| Producible | G1 + G2 + G3 all observed for the form's equivalent-form artifact |
| Partial | G1 observed; G2 or G3 observed only after degrading the form (dropped structural elements); deviation material |
| Not-producible | No artifact satisfying G1-G3 expressible without abandoning the form's information structure |
| Preset-gap | Failure class: a missing component preset / reference pattern a future skill addition could supply |
| Structural-limit | Failure class: the skill's numeric-layout method cannot express the form |
| Evidence dir | `.moai/reports/t272/` (worktree-relative, carried on branch `WT-skillstead-gallery`) |

## §C Scope

In scope:

1. The benchmark protocol: nine pinned briefs (§D.2), per-artifact gates G1-G3,
   verdict table, failure taxonomy.
2. The docs emphasis: "producible diagram types" content on docs-site and README,
   4 locales, ko canonical, grounded only in measured artifacts.
3. The verification pass: the oss-docs verify recipe over the touched surfaces.

Out of scope: see the `### Out of Scope` subsections in §G (required there for the
`OutOfScopeRule` lint convention).

## §D Requirements

### §D.1 Requirements (GEARS)

**REQ-SGB-001 (coverage)** — The benchmark run shall attempt an equivalent-form
artifact for each of the nine TypePack forms pinned in §D.2, with no form skipped,
deferred, or substituted with an unrelated topic.

**REQ-SGB-002 (pinned brief and reproducibility)** — Each artifact generation shall
start from the concrete brief pinned in §D.2 (topic, message) with the four output
dials settled before authoring, and the settled dial values shall be recorded in the
verdict row's settled-dials field (REQ-SGB-004) so a later regeneration reproduces
the artifact.

**REQ-SGB-003 (observable gates)** — Per form, the benchmark record shall capture the
three observable gates of §A: G1 SVG existence, G2 lint exit 0 with zero errors and
triaged warnings, G3 render exit 0 with verified 2x PNG dimensions and disclosed
browser executable + version.

**REQ-SGB-004 (verdict table)** — The run shall produce a verdict table at
`.moai/reports/t272/verdict.md` with one row per form carrying: form name, verdict
(`PRODUCIBLE` / `PARTIAL` / `NOT-PRODUCIBLE`), archetype used or `none`, gate results
G1-G3, the settled dial values (format, size, detail, audience — REQ-SGB-002), and
evidence file paths.

**REQ-SGB-005 (failure taxonomy)** — **When** a form's gate fails or only a partial
equivalent is achievable, the run record shall classify the gap as either `preset-gap`
(a missing component preset or reference pattern that a future skill addition could
supply) or `structural-limit` (the skill's numeric-layout method cannot express the
form), citing the observable evidence that justifies the classification.

**REQ-SGB-006 (deviation naming)** — **When** a produced artifact is equivalent but
not visually identical to the TypePack form, the verdict row shall name the deviation
in one sentence; an unnamed deviation is an incomplete verdict row.

**REQ-SGB-007 (evidence location)** — All benchmark artifacts, gate logs, and the
verdict table shall live under the worktree-relative evidence directory
`.moai/reports/t272/` on branch `WT-skillstead-gallery`; no evidence path outside this
directory shall be cited by the docs deliverable.

**REQ-SGB-008 (skill immutability)** — The benchmark shall not modify any file under
`.claude/skills/moai-domain-svg-infographic/` or under its template mirror
`internal/template/templates/.claude/skills/moai-domain-svg-infographic/`.

**REQ-SGB-009 (docs grounding)** — **While** the producible-diagram-types emphasis
section is being authored, every listed diagram type shall cite the evidence file path
of its measured artifact; the docs and README emphasis shall list `PRODUCIBLE` forms
only — `PARTIAL` and `NOT-PRODUCIBLE` outcomes are recorded exclusively in the card
evidence (`.moai/reports/t272/` verdict), never surfaced in user-facing docs.

**REQ-SGB-010 (docs 4-locale parity)** — The docs-site emphasis shall land in all four
locales (ko canonical; en/ja/zh derived in the ko→en→ja/zh chain) within the same PR,
per the oss-docs i18n rules.

**REQ-SGB-011 (README chain)** — The README emphasis shall be authored in
`README.ko.md` first and minimally derived into `README.md`, `README.ja.md`,
`README.zh.md`, preserving H2 heading parity and the language-switcher header contract
across all four files.

**REQ-SGB-012 (verify recipe)** — The run shall execute the `hns-oss-docs-verify`
recipe over the touched surfaces (warning-free hugo build + sitemap, URL blacklist,
Mermaid TD-only, 4-locale file-existence and section-count parity, README heading
parity, body-emoji scan, version-string sync) and record the per-check results.

**REQ-SGB-013 (environment preflight blocker)** — **Where** the benchmark environment
lacks Node 18+ or a headless Chromium-family browser, the run shall stop and return a
blocker report rather than delivering any unverified gate claim, honoring the skill's
own no-fabrication degradation contract.

### §D.2 Pinned brief table (the benchmark protocol)

All nine briefs use the skill's default dials — `format: svg+png`,
`size: doc-inline` (1200 wide), `detail: balanced` (<=12 nodes), `audience: mixed` —
so results are comparable across forms. A deviation forced by a form (e.g. a wider
canvas) is recorded in the verdict row, not silently applied. The "mapping hypothesis"
column is the archetype the skill is expected to route to; the benchmark tests it, it
does not assume it.

| # | TypePack form | Mapping hypothesis | Pinned brief (topic → message) |
|---|---------------|--------------------|--------------------------------|
| 1 | `approval-gate` | A2 flow + a distinct gate node | v3.1.4 release approval: PR draft → CI matrix green → code review → maintainer approval (the gate) → merge |
| 2 | `before-after` | two-panel composite (A1 x2 or A3) | agent-catalog redesign before/after: monolithic orchestrator vs 12-agent catalog, same three concerns on both panels |
| 3 | `cards-kpi-grid` | card grid (no direct archetype; A1 single-layer is nearest) | factory-lane KPI summary: 4 stat cards — cards completed, PR merge rate, CI pass rate, review turnaround |
| 4 | `decision-matrix` | A3 side-by-side comparison | database selection matrix: PostgreSQL / MongoDB / Redis x 4 criteria (data model, scaling, ops cost, familiarity) |
| 5 | `layer-stack` | A1 architecture stack (direct match) | moai-adk CLI architecture: cmd/moai → internal/cli → internal/{config,hook,spec} → internal/template |
| 6 | `nested-scope` | A4 hierarchy or A1 nested containers | trust-boundary nesting: organization → kanban lead session → lane worktree → card evidence |
| 7 | `process-flow` | A2 left-to-right flow (direct match) | kanban card lifecycle: backlog → plan → run → sync → done, with the sync gate marked |
| 8 | `roadmap-timeline` | none direct (time-phased bands) | SPEC 3-phase roadmap across an Epic: plan / run / sync milestones on a time axis |
| 9 | `topology-component` | A1/A2 composite component topology | web-console topology: browser → statusline/web console → moai CLI → tmux lanes / MCP broker |

Evidence layout under `.moai/reports/t272/`:

```
.moai/reports/t272/
├── verdict.md                 # REQ-SGB-004 verdict table
├── artifacts/<form>.svg       # G1, one per form
├── artifacts/<form>.png       # G3 output, one per form
└── logs/<form>-lint.txt       # G2 log — see the log-format contract below
    logs/<form>-render.txt     # G3 log — see the log-format contract below
```

**Log-format contract (single definition; every evidence log under `logs/`
follows it).** Each log is the command line, the tool's verbatim output, and an
explicit exit-status line, captured in one invocation:

```bash
{ node .claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs \
    .moai/reports/t272/artifacts/<form>.svg; echo "exit=$?"; } \
  > .moai/reports/t272/logs/<form>-lint.txt
```

The render log uses the same shape with `render.mjs <svg> --out <png>`. The
`exit=N` line is the machine-checkable pass signal AC-003 reads; a log missing any
of the three parts (command line, verbatim output, exit-status line) does not
satisfy G2/G3 evidence.

## §E Constraints

1. **No skill modification** (REQ-SGB-008): the benchmark measures the skill at its
   current `main` state. Preset additions suggested by the taxonomy are follow-up
   cards, never in-run edits.
2. **No fabricated gates** (REQ-SGB-013): the skill's degradation contract applies —
   a tool that did not run produces no verdict, only a named skip.
3. **oss-docs HARD rules** bind the docs deliverable: ko canonical chains (docs-site
   and README both), 4-locale same-PR, Mermaid TD-only in docs content, no body-text
   emoji (icon shortcode instead), URL blacklist (only `adk.mo.ai.kr` is valid;
   `docs.moai-ai.dev`, `adk.moai.com`, `adk.moai.kr` are forbidden), version SSOT
   from `hugo.toml`.
4. **Template-First boundary**: docs-site (`docs-site/`) and the README set (repo
   root) are repo surfaces, NOT template surfaces — no `internal/template/templates/`
   mirror is created or edited. The evidence dir `.moai/reports/t272/` is a
   local-only surface per CLAUDE.local.md §2 (never templated).
5. **SPEC language**: this SPEC and the verdict table are English; docs content is
   locale-specific (ko authored first). 
6. **No time estimates** in the plan; priority labels and phase ordering only.

## §F Success Criteria

- Nine verdict rows exist, each backed by observed gate evidence or a taxonomy
  classification with cited evidence.
- The docs emphasis lists exactly the `PRODUCIBLE` forms, each citing its artifact
  path; `PARTIAL` and `NOT-PRODUCIBLE` outcomes stay in the card evidence only.
- The verify recipe passes on the touched surfaces.
- `git diff` on the skill directory (local + template mirror) is empty.

## §G Exclusions

### Out of Scope — skill modification and preset authoring

- Adding or editing any archetype, reference, fixture, or script under
  `.claude/skills/moai-domain-svg-infographic/` or its template mirror.
- Authoring new component presets the taxonomy suggests; those are follow-up cards.

### Out of Scope — SkillStead parity features

- Implementing receipt badges, gallery generation, or a TypePack-style verified
  catalog as a product feature.
- Vendoring, porting, or importing any SkillStead code or assets (only the nine form
  definitions are borrowed as benchmark briefs).

### Out of Scope — unrelated docs surfaces

- Any docs-site page or README section other than the svg-infographic emphasis
  surface named in plan.md §F.
- Vercel binding or deployment configuration changes (redirects excepted, and none
  are expected: no page moves).

## §H History

| Date | Author | Change |
|------|--------|--------|
| 2026-08-25 | manager-spec | Initial draft (plan-phase, kanban card t272, Tier M) |
| 2026-08-25 | manager-spec | 0.1.1 — audit repairs D1 (settled-dials field in verdict-row schema + mechanical AC checks), D2 (single log-format contract), D3 (brief #5 `pkg/template` → `internal/template`) |
| 2026-08-25 | manager-spec | 0.1.2 — operator decisions at Implementation Kickoff Approval: phase target `v3.1.4 target` → `v3.1.3` (docs fold into pending v3.1.3 doc update; t204 release gate holds the tag); docs/README emphasis narrowed to PRODUCIBLE-only — PARTIAL/NOT-PRODUCIBLE recorded exclusively in card evidence |

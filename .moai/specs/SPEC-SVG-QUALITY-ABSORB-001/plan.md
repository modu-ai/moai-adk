# Plan — SPEC-SVG-QUALITY-ABSORB-001

> Ordered so the decisions most likely to change come first: the palette
> restructure and the exception list are the reviewable calls; the rest is
> rule-writing against a survey that already did the analysis.
>
> Tier M (3 artifacts). Tier S was the survey's estimate; the criterion count
> (14) exceeds Tier S's budget of 8, and tiering up is the prescribed response.
> Scope is eight files: four skill files plus their four mirrors.

## §A Context

Six quality layers move from a surveyed MIT source into an existing skill. Five
are additive text; one (A-6) restructures a section that other sections depend
on, and one (REQ-7) produces a routing carve-out that has to be earned with
samples rather than argued.

Target surfaces, measured this session:

| File | Bytes | What changes |
|---|---|---|
| `SKILL.md` | 14,190 | A-2 budgets, A-3 contract, A-4 Red Flags, A-5 dials |
| `references/authoring.md` | 12,560 | A-1 §2 connectors, A-3 skeleton, A-6 §3 palette |
| `references/archetypes.md` | 6,110 | A-2 per-archetype budgets |
| `scripts/check-svg.mjs` | 16,934 | AC-3b accessible-name check |
| template mirrors of the four above | — | Template-First (4 files) |

## §B Known Issues Going In

- **A-6 is a restructure, not an append.** §3 currently carries palette, type
  scale, typography, and focal discipline. Reorganising around semantic roles
  can silently drop the one-accent rule, which is why AC-6b exists as a separate
  criterion.
- **`SKILL.md` has a size ceiling to respect.** Four additions to a 14 KB file
  risk pushing the L2 body past its budget; A-2 and A-4 detail belong in the
  references with the SKILL.md side kept to the rule and a pointer.
- **REQ-7 can produce nothing.** If no sample pair shows a decisive difference,
  the list is empty and the routing table is unchanged. The plan must not
  presuppose entries.
- **The source is MIT, this repo is Apache-2.0.** Compatible, with an
  attribution duty — the risk is forgetting it, not licence conflict.

## §C Pre-Flight

1. `.moai/reports/t165/upstream/survey.md` readable **from this branch** — the
   rule text is derived from it. The survey originally lived in an untracked
   report directory that exists only in the primary checkout, so it is committed
   here; a pre-flight pointing at the untracked path would fail in the very tree
   the run executes in.
2. `.moai/reports/t165/upstream/UPSTREAM-LICENSE` present — the attribution basis
   (REQ-8), committed for the same reason. The `/tmp/diagram-design` clone is
   volatile and MUST NOT be a pre-flight dependency.
3. `node scripts/check-svg.mjs --help` (or equivalent) runs — AC-3b extends it.

## §D Constraints (Hard)

- No external asset at view time. Google Fonts is rejected; absorbed typography
  maps onto the existing language-aware stacks.
- Static SVG + PNG only. No HTML variants, no motion, no terminal skin.
- The exception list touches the image-output path only, never locale-synced text.
- Template-First: mirror + `make build` in the same commit.
- Rewrite, don't copy. Attribution recorded regardless.

## §E Design Decisions

### D1 — Absorb rules, refuse the catalogue

The 39 type specs are the bulk of the source and the least transferable part:
they presuppose a skill that owns every diagram type, while this one routes most
types to mermaid. Taking the skin, the connector geometry, and the budgets takes
what produced the look; taking the catalogue would take a different skill's
architecture along with it.

### D2 — A-3 is the priority item despite being listed third

The other five improve diagrams that already work. A-3 fixes output that is
currently invisible to assistive technology — a correctness gap, not a polish
gap. It is also the only item with an executable check (AC-3b), so it is the one
that can regress silently later.

### D3 — The exception list is gated on artifacts, not on argument

A routing carve-out is a durable decision: once a type bypasses mermaid, every
future diagram of that type is an image with the maintenance cost images carry.
Requiring a rendered pair per type makes the cost visible before the carve-out is
taken, and makes an empty list a legitimate answer.

### D4 — Budgets live in the references, pointers in SKILL.md

Progressive disclosure: the L2 body carries the rule ("every type has a node
ceiling; exceeding it means split or simplify") and the table lives at L3. This
keeps the always-listed metadata and the L2 body inside budget.

## §F Milestones

### M1 — A-6 palette restructure (`authoring.md` §3)

First because everything else references colour roles. Ends with AC-6a and AC-6b
both satisfied — the second guards what the restructure could drop.

### M2 — A-1 connector rules (`authoring.md` §2)

Six numeric rules into the existing connector section, phrased so B-7 can later
assert them mechanically.

### M3 — A-3 accessible-SVG contract (`SKILL.md` + `authoring.md` + `check-svg.mjs`)

Contract text, copyable skeleton, and the checker extension with both-direction
tests (AC-3b).

### M4 — A-2 budgets + A-4 anti-patterns + A-5 dials

Three text additions with no interdependency; grouped to avoid three passes over
the same two files.

### M5 — REQ-7 sample comparison

Per candidate type: render the mermaid form and the absorbed-rule form, store
both under `.moai/reports/t165/samples/`, and decide. Only decided types enter
the list.

### M6 — Template-First mirror + `make build` + verification

Mirror every changed skill file, rebuild, run the neutrality guard, and record
the attribution.

## §G Anti-Patterns (Avoid)

- Pasting source text verbatim. The two skills' contracts differ; a lift imports
  rules this one contradicts (external fonts, HTML variants).
- Letting the exception list grow by argument. No sample pair, no entry.
- Restructuring §3 and assuming the focal discipline came along — check it.
- Adding A-2's full table to `SKILL.md` and pushing the L2 body over budget.
- Treating a passing `check-svg.mjs` as proof the accessibility check works.
  Assert the failing direction too.

## §H Cross-References

- `.moai/reports/diagram-design-absorption/survey.md` — the source analysis
- `.moai/reports/diagram-design-absorption/diagram-design-absorption-20260822.md`
  — the A/B/C/D verdict this implements
- `.claude/rules/moai/development/skill-authoring.md` — progressive-disclosure
  budgets that bound D4

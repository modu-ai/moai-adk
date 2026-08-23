---
id: SPEC-SVG-GEOMETRY-CHECKS-001
title: "SVG connector-geometry checks in check-svg.mjs with bipolar self-test"
version: "0.3.0"
status: completed
created: 2026-08-22
updated: 2026-08-23
author: manager-spec
priority: P2
phase: "v3.1.3 target"
module: ".claude/skills/moai-domain-svg-infographic, internal/template/templates/.claude/skills/moai-domain-svg-infographic"
lifecycle: spec-anchored
tags: "svg, lint, geometry, connector-rules, self-test, template-mirror"
tier: M
---

# SPEC-SVG-GEOMETRY-CHECKS-001 — connector-geometry checks in check-svg.mjs

## HISTORY

| Date | Version | Change |
|---|---|---|
| 2026-08-22 | 0.1.0 | Initial plan-phase authoring (card t166). Five diagnostics (SVG070-SVG074), path-geometry reader, bipolar self-test runner, template mirror. |
| 2026-08-22 | 0.2.0 | Audit iteration 2. §B D7 records the C4 contradiction and the arrival-only ruling. `Q` subdivided; association window bounded at 16; `SVG071` narrowed to non-mask rects; `SVG074` count defined as transitively excluded; REQ-SGC-009's independence claim corrected. Requirement and AC counts unchanged at 16/16. |
| 2026-08-22 | 0.3.0 | Iteration-2 PASS (0.81) debt cleared. §B D8 accepts `SVG071`'s over-breadth; both-markers arrival case stated; empty-set header literal pinned with a count assertion; arrival semantics given must-flag twins; window invariant moved to K7; K3's bounded exception acknowledged; marker-less C4 gap disclosed in §A. Counts unchanged at 16/16. |

---

## §A Context

`scripts/check-svg.mjs` in the `moai-domain-svg-infographic` skill is a deterministic, render-free
source lint for hand-authored SVG infographics. It reads the file as text, keeps a light element
tree, and emits `file:line:column  level  code  message` with a stable code per finding. Its
present coverage is document structure (`SVG001`-`SVG050`), heuristic text fit (`SVG030`/`SVG031`,
warnings), viewBox overflow (`SVG040`, warning), and the accessible-SVG contract
(`SVG060`-`SVG064`).

The six mandatory connector rules in `references/authoring.md` §2.5 are settled prose. Three of
them carry a number a checker can assert, but nothing asserts it today: C2 (a label mask clears its
own stroke by 6-10 units), C4 (connectors on a shared box edge fan to attach points at least 12
units apart, 8 for a box under 120 units on that edge), and C6 (a label mask may not *partially*
overlap a node painted after it — fully inside is an allowed badge chip, fully outside is fine).
`SKILL.md` slop-symptom row 12 names all three breaches as failures, and `authoring.md` §8.3 states
the two existing fixture commands as manual prose, so today the only enforcement is a human reading
the source.

This SPEC mechanises those three rules inside the existing lint, and adds the self-test that keeps
the checks honest in both directions.

**What the C2 mechanisation does not reach.** `archetypes.md` A2 places branch labels at
`(arrowMidX, stageY - 14)` while its connector runs at `stageY + stageH/2` — roughly 99 units apart
at the documented `stageH = 170`. That is far outside the 16-unit association window of
REQ-SGC-001, so the archetype-standard label placement in the skill's most common flow archetype is
**not** checked by `SVG070` / `SVG073`. This is the accepted-false-negative posture of §B D3 rather
than a defect in the check, and it is recorded here so the SPEC does not overstate what lands.

**What the C4 mechanisation does not reach.** The arrival-only ruling of §B D7 has two blind spots,
both consequences of the ruling rather than gaps in it. Crowding on the **departure** side is never
reported — that is the ruling's stated cost. And `SVG072` is silent on any connector carrying
neither `marker-end` nor `marker-start`, because such a connector has no arrival point at all: an
undirected connector may crowd an edge freely. `SKILL.md` slop-symptom row 12 and `authoring.md`
§2.5 are annotated (REQ-SGC-016) to say C4 is machine-checked **on arrival points**, not in full,
so the claim that ships to users in the template mirror matches what the checker does.

### Why now

The prerequisite quality-layer card has landed: §2.5 is the settled rule text, with numbers, and it
is not re-opened here. This SPEC mechanises rules that already exist and imposes **no new authoring
obligation** — nothing an author must now write differently. It is not, however, strictly confined
to §2.5 as written: §B D7 resolves a contradiction inside C4, and §B D8 accepts a check that is
deliberately wider than C6. Both are recorded as binding decisions rather than left as implicit
drift, and neither changes what an author has to do.

---

## §B Scope decisions (binding; the run phase does not re-open these)

| # | Decision | Consequence |
|---|---|---|
| D1 | Classification uses **geometry and existing document shape** — sibling order (D2), marker attributes (D7) — but **no new authoring convention** (no `class="connector-label"`) | The authoring contract in `authoring.md` / `archetypes.md` is unchanged, and already-authored SVGs are covered without re-authoring |
| D2 | A connector-label mask candidate is a `<rect>` whose **immediately following sibling is a `<text>`**, and which is **not an attach target of any connector endpoint** | Nodes and containers fall out of the candidate set naturally |
| D3 | A mask wrapped so it has no adjacent `<text>` sibling is an accepted **false negative**; a **false positive is not acceptable** | Every ambiguity resolves toward silence |
| D4 | All checks live in the existing `lint()` of `check-svg.mjs` | No second script, no second parser, no second measurement path |
| D5 | Node 18+ standard library only; no package install, no network, no browser | Matches the script's existing contract |
| D6 | Three errors plus two warnings: `SVG070` `SVG071` `SVG072` errors, `SVG073` `SVG074` warnings | Fits the existing tier semantics (error = deterministic and structural) |
| D7 | **`SVG072` binds ARRIVAL points only** — see below | Two connectors leaving one point are permitted; two arriving at one point are the error |
| D8 | **`SVG071` is deliberately wider than C6** for non-mask, non-node rects — see below | A mask partially overlapping a later decorative divider or rule is reported, by intent |

### D7 — the C4 contradiction, and the ruling that resolves it

The premise that §2.5's rules are settled and internally consistent is **partially false**, and this
SPEC records that rather than papering over it. C4 states that no two connectors may meet a box at
the same point, but two of the skill's own formulas emit exactly that:

- `authoring.md` §2.3 fan-out — `for each target T: M A.right A.cy H trunkX V T.cy H (T.left - markerLen)`
  — every connector in the family departs from the identical point `(A.right, A.cy)`.
- `archetypes.md` A4 hierarchy tree — every child connector begins `M parent.cx nodeY(L)+nodeH`,
  one identical point on the parent's bottom edge.

Mechanising C4 literally would report `SVG072` on both, i.e. on one of the four archetypes that
`archetypes.md` says cover nearly every request — a false positive on a clean diagram, which K3
makes a defect.

**Ruling (operator): `SVG072` binds arrival points only.** An **arrival point** is an endpoint at
which a marker resolves — the final polyline point under `marker-end`, the first under
`marker-start`, both when the connector carries both markers, and none when it carries neither, in
which case the connector contributes nothing.

**Grounding.** C4's own formula `offset(k) = L*k/(N+1)` describes N connectors *dividing* an edge
between them, which is a statement about where they land. §2.3 states the shared trunk is
deliberate — "so the lines read as one branch rather than as several unrelated arrows" — so a
shared departure is the documented intent, not an oversight. The failure C4 guards against is two
arrowheads landing indistinguishably on one box, and that is a property of arrival.

**Consequences.** No document changes and no existing diagram breaks. The accepted cost is a false
negative: genuine crowding on the departure side goes unreported, and a connector carrying no marker
at all is never compared (disclosed in §A). Two further effects follow from the same ruling and are
recorded here so no other section claims otherwise — the `<defs>` arrowhead of `plan.md` §B4 carries
no marker attribute of its own, so it now has zero arrival points and is excluded a third time,
independently of B0 and of the `fill` guard; and a closed path (`z`) carrying **exactly one** of the
two markers contributes one arrival point rather than a zero-separation pair. That last consequence
is **not** universal: a closed path carrying **both** markers has two arrival points which coincide,
and `SVG072` reports it — correctly, since two arrowheads then land on the same spot. The
`c4-closed-both-markers` fixture pins that case as a must-flag rather than leaving the narrower
claim to be read as covering it.

### D8 — `SVG071` is deliberately wider than C6

C6 constrains a mask overlapping a **node**. `SVG071` as specified reports a mask partially
overlapping any later-painted rect that is not itself a mask candidate, which includes a decorative
divider, rule, or highlight — shapes C6 does not name.

Narrowing further is not available: separating a node from a decorative rect geometrically would
need the marking convention D1 declines, so the choice is between reporting the wider set and
dropping the check.

**Ruling: report the wider set.** What it catches is a real legibility fault whether or not C6 names
it — a label clipped by a later-painted rect renders as a fragment on that rect's border regardless
of what the rect means. The accepted cost is stated plainly: a diagram that *deliberately* overlaps
a connector label with a later decorative rule gets an error it must restructure around, and it has
no way to suppress it. `acceptance.md` § D.9 asserts this as expected behaviour rather than
tolerating it as drift.

---

## §C Requirements (GEARS)

> **Polarity note.** Each requirement below states its rule with the boundary that decides it, in
> one entry, rather than as a must-flag entry plus a must-not-flag twin. The per-value polarity
> cases (exactly 6 → silent, 5 → `SVG070`, and so on) live in `acceptance.md` § D.9, which is the
> mechanically-asserted edge-case table, not prose. Nothing is unasserted by being stated there.

### Diagnostics

- **REQ-SGC-001** (event-driven) — When a connector-label mask clears the stroke of its own
  connector by less than 6 units, the linter shall report `SVG070` at the error level, positioned
  at the mask rect; a mask that crosses or touches the stroke is the distance-0 case of this rule,
  and a clearance of 6 to 10 units inclusive shall produce no diagnostic. A mask associates to its
  own connector only within **16 units** — the C2 band's upper bound of 10 plus a tolerance of 6,
  6 being C2's own lower bound and so the smallest distance the rule treats as meaningful; a mask
  farther than 16 units from every connector associates to none and is not checked.
- **REQ-SGC-002** (event-driven) — When a connector-label mask clears its own stroke by more than
  10 units, the linter shall report `SVG073` at the warning level.
- **REQ-SGC-003** (event-driven) — When a connector-label mask **partially** overlaps a `<rect>`
  appearing later in document order, the linter shall report `SVG071` at the error level; a mask
  lying fully inside that rect (the badge chip) and a mask sharing no area with it shall produce no
  diagnostic. The later-rect set shall exclude every other label-mask candidate: a mask overlapping
  another mask is not an `SVG071`. It shall retain every other later-painted rect, including
  decorative dividers and rules that C6 does not name — deliberately wider than C6, per §B D8.
- **REQ-SGC-004** (event-driven) — When two connector **arrival** points bind to the same box edge
  with a separation below the floor, the linter shall report `SVG072` at the error level; coincident
  arrival points are the separation-zero case of this rule and shall not carry a separate code. An
  arrival point is an endpoint at which a marker resolves: the final polyline point under
  `marker-end`, the first under `marker-start`, **both** when the connector carries both markers,
  and none when it carries neither. Departure points shall not be compared (§B D7).
- **REQ-SGC-005** (state-driven) — While the box edge carrying the attach points is 120 units or
  longer, the `SVG072` floor shall be 12 units; while that edge is shorter than 120 units, the
  floor shall be 8 units.
- **REQ-SGC-006** (event-driven) — When one or more elements were excluded from the checks for
  carrying a `transform`, the linter shall emit exactly one aggregate `SVG074` warning per file,
  positioned at the root `<svg>` offset so its `file:line:column` is deterministic, and shall not
  emit one note per excluded element. The count it reports shall be the **transitively excluded**
  element count — every element the ancestor-walking `hasTransform` predicate excludes, not only
  those literally bearing the attribute — and the message shall state that count against the total
  candidate population, so a diagram wrapped in a single `<g transform>` discloses that all of its
  geometry went unverified rather than reporting one skipped element.
- **REQ-SGC-007** (ubiquitous) — Every diagnostic this SPEC adds shall carry the existing
  `file:line:column  level  code  message` shape, appear in `--json` output under the same
  `diagnostics` schema as existing codes, and respect `--strict`.

### Measurement path

- **REQ-SGC-008** (ubiquitous) — The checks shall be implemented inside the existing `lint()`
  function of `scripts/check-svg.mjs`, reusing its existing `tokenize` / `buildTree` / `walk` /
  `positionOf` / `num` / `hasTransform` / `report` helpers. The implementation — the linter and the
  self-test runner alike — shall introduce no second SVG parser, no second geometry model, and no
  dependency outside the Node 18 standard library: no package manifest, no lockfile, no installed
  module, no network call, and no browser.

### Candidate-set exclusions

- **REQ-SGC-009** (unwanted) — The following shall not enter the connector set, the box set, or the
  label-mask candidate set, and shall therefore yield no SVG07x diagnostic: (a) any element inside
  a non-rendered subtree — `<defs>`, `<marker>`, `<symbol>`, `<clipPath>`, `<mask>`, `<pattern>`;
  (b) any `<path>` carrying a `fill` other than `none`, the documented connector idiom being
  `fill="none"` plus a stroke; (c) any element carrying a `transform`, matching the existing
  `SVG040` behaviour. The three guards are **not** independent for every shape: the sketch-mode
  arrowhead documented in `sketch.md` §3 carries `fill="none"`, so guard (b) does not catch it and
  (a) is the only exclusion that does — which is why the `defs-marker-noise` fixture pins a stroked
  `fill="none"` arrowhead rather than a filled one, so (a) is the guard actually under test.

### Path-geometry reader

- **REQ-SGC-010** (ubiquitous) — The linter shall recover a connector's polyline from its `d`
  attribute for the documented idiom: `M` and `L` as points, `H` and `V` as axis-constrained
  points, `A` treated as a pass-through between its current point and its endpoint, `Z` emitting
  the subpath start, in both absolute and relative forms including implicit repeated coordinate
  pairs. A `Q` shall be **subdivided** — sampled at t = 0.25, 0.5, and 0.75 in addition to its
  endpoints, with the control point dropped as a vertex — because the control point of the
  documented elbow is the un-rounded corner, and emitting it biases the measured clearance inward
  by up to `0.25·r·√2` (≈ 2.83 units at the documented `r = 8`), which would report a mask authored
  at C2's documented minimum of 6 near a corner as an `SVG070` on a compliant diagram.
- **REQ-SGC-011** (event-driven) — When a `d` attribute carries a command the reader cannot fully
  interpret, the linter shall skip that connector silently, reporting no SVG07x diagnostic derived
  from it and neither guessing, approximating, nor partially evaluating the path to produce one.

### Attach-point detection

- **REQ-SGC-012** (ubiquitous) — Each connector endpoint shall bind to exactly one box edge — the
  nearest — and only where the endpoint lies within the documented `markerLen = 10` standoff of
  that edge and its projection falls within the edge's span. An endpoint binding to no edge shall
  contribute to no `SVG072` finding.

### Self-test

- **REQ-SGC-013** (ubiquitous) — The skill shall carry a self-test runner at
  `scripts/test-check-svg.mjs` which spawns the real `scripts/check-svg.mjs --json` as a child
  process for every fixture, prints one `PASS` or `FAIL` line per fixture, and exits 0 when every
  fixture matches and 1 otherwise.
- **REQ-SGC-014** (ubiquitous) — Every fixture shall declare its expectation in-file as a leading
  `<!-- expect: ... -->` comment naming the exact diagnostic code set, and the runner shall assert
  the **exact** emitted code set against that declaration rather than the process exit code alone.
  The empty set shall be written in exactly one form — the literal `<!-- expect: -->`, with no
  placeholder token between the colon and the closing `-->` — because that literal is what selects
  the clean-fixture population in AC-SGC-004; a fixture written `<!-- expect: {} -->` or
  `<!-- expect: none -->` would be silently omitted from that selection.
- **REQ-SGC-015** (ubiquitous) — The fixture set shall carry at least one must-flag and one
  must-not-flag fixture for each of `SVG070`, `SVG071`, `SVG072`, plus a must-flag fixture for
  `SVG073` and one for `SVG074`; every fixture shall otherwise satisfy the accessible-SVG contract
  and the structural checks, so the code under test is the only diagnostic it produces. The two
  pre-existing fixtures `a11y-present.svg` and `a11y-missing.svg` shall carry expectation headers
  and be run by the runner, so it subsumes the two manual commands of `authoring.md` §8.3.

### Documentation and mirror

- **REQ-SGC-016** (ubiquitous) — The change shall carry its documentation and its template mirror:
  the `check-svg.mjs` header comment documents the SVG07x tier; `SKILL.md` gains a script-table row
  for the runner, an updated fixtures row, and a slop-symptom row 12 stating which of the C2, C4,
  and C6 breaches it names are now machine-checked, with their codes and with the coverage bounds of
  §A — C4 on arrival points only, C2 within the association window; `authoring.md` §2.5 annotates
  C2, C4, and C6 with their codes and the same bounds, and §8.3 points at the runner; and every added or changed file is
  mirrored to `internal/template/templates/.claude/skills/moai-domain-svg-infographic/` followed by
  `make build`, with the mirrored files carrying no SPEC ID, card id, internal date, commit SHA, or
  other internal-development marker.

---

## §D Out of Scope

### Out of Scope — the other three connector rules

- C1 (orthogonal-only, rounded bends at `r = 8`) is not mechanised here.
- C3 (no two connectors sharing a path, ≥ 12 units of parallel separation, hop at crossings) is not
  mechanised here.
- C5 (dashed transit behind a non-endpoint box, no arrowhead on the intervening edge) is not
  mechanised here.

### Out of Scope — authoring-contract changes

- No new class, attribute, id convention, or marking scheme is introduced for connector labels.
- `archetypes.md` is not modified.
- The palette, type scale, icon set, and CJK text-budget sections of `authoring.md` are untouched.

### Out of Scope — rendering and tooling

- `scripts/render.mjs` is not modified; no rasterised or pixel-level verification is added.
- No headless browser, SVG rendering engine, or external SVG parser is introduced.
- No CI workflow is added or modified for the runner in this SPEC.

### Out of Scope — Go-side work

- No Go source change beyond what `make build` regenerates from the mirrored template tree.
- A local full Go suite (`go test ./...`) is out of scope on this machine; the full-suite verdict
  belongs to CI.

---

## §E Constraints

| # | Constraint |
|---|---|
| K1 | Node 18+ standard library only; no install, no network, no browser |
| K2 | Existing diagnostic codes, levels, message shape, and exit codes are unchanged |
| K3 | A false positive on a clean diagram is a defect; a false negative on an unusual authoring shape is accepted. **One bounded exception, acknowledged rather than implicit**: a non-label `<rect>`+`<text>` pair sitting 11-16 units from a connector is warned `SVG073`. Removing it needs the discriminator §B D1 declines, so it is bounded by the 16-unit window instead, fixtured as `chip-in-window`, and derived in `plan.md` §B. No other false positive is accepted |
| K4 | The template mirror stays content-neutral (no SPEC ID, card id, date, or SHA) |
| K5 | Verification uses named commands only; no local full Go suite |
| K6 | `SVG074` is a warning, so `--strict` now fails any diagram carrying a transformed element where it previously passed. This is an accepted behaviour change: a transformed element genuinely has unverified geometry, `--strict` is opt-in, and the default (non-strict) exit code is unaffected |
| K7 | The mask-association window shall never exceed the band it feeds by more than the stated tolerance. This governs future changes to the window rather than the linter's behaviour on any input, so it is a constraint here and not a requirement — no fixture can assert it (`plan.md` §B derives the current 10 + 6 = 16) |

---

## §F Cross-references

- `.claude/skills/moai-domain-svg-infographic/references/authoring.md` §2.5 — the six connector
  rules being mechanised, and the paint order the C6 check depends on
- `.claude/skills/moai-domain-svg-infographic/references/authoring.md` §2.1-2.4 — the connector
  idiom the path reader targets (`markerLen = 10`, `r = 8`, `rHop = 5`)
- `.claude/skills/moai-domain-svg-infographic/SKILL.md` — script table, fixtures row, slop-symptom
  row 12
- `.moai/specs/SPEC-SVG-GEOMETRY-CHECKS-001/plan.md` — milestones and technical approach
- `.moai/specs/SPEC-SVG-GEOMETRY-CHECKS-001/acceptance.md` — the AC matrix

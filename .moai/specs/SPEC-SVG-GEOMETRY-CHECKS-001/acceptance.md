# SPEC-SVG-GEOMETRY-CHECKS-001 — Acceptance Criteria

Path shorthand used below (project-root-relative). Paste this before running any command below; it is an assignment, not prose, so the `$S` / `$F` /
`$T` in every command resolve in this tree:

```bash
S=.claude/skills/moai-domain-svg-infographic
F=$S/scripts/fixtures
T=internal/template/templates/$S
```

Every AC is binary-testable: five carry an explicit `Command:` line, and the rest are verified by
running the linter or the runner on the fixture the criterion names. `go test ./...` is prohibited
on this machine; Go-side verification is limited to the template-embed packages `make build`
touches, with the full-suite verdict left to CI.

> **Where the per-case assertions live.** The numbered ACs below are the criteria a reviewer reads.
> The per-value boundary cases are asserted mechanically, not by prose: § D.9's edge-case table
> enumerates them, and every row of it is realised as a fixture whose `<!-- expect: … -->` header
> names the exact code set the runner asserts (AC-SGC-003). A case moved into that table is asserted
> more strictly than it was as prose, not less.

---

## §D AC Matrix

### D.1 — Runner and exact-code assertion

- **AC-SGC-001** — *Given* the fixture set and the runner, *When*
  `node $S/scripts/test-check-svg.mjs` is run, *Then* it prints one `PASS` or `FAIL` line per
  fixture and exits 0.
  Command: `node $S/scripts/test-check-svg.mjs; echo "rc=$?"`
- **AC-SGC-002** — *Given* a fixture whose declared expectation is deliberately mutated to a wrong
  code set in a scratch copy, *When* the runner is run, *Then* it prints `FAIL` for that fixture and
  exits 1 — the runner can actually fail.
- **AC-SGC-003** — *Given* every fixture, *When* the runner compares results, *Then* the comparison
  is against the **exact** emitted `code` set parsed from `--json`, not the exit code alone —
  verifiable by grepping the runner source for the set comparison and the absence of an
  exit-code-only branch.
  Command: `grep -n 'code' $S/scripts/test-check-svg.mjs`

### D.2 — No false positives on clean input

- **AC-SGC-004** — *Given* every fixture whose declared expectation is the empty set — including
  `a11y-present` (pre-existing), `c2-mask-clear`, `c2-mask-outer-elbow-corner-at-6`,
  `c6-mask-inside`, `c6-mask-outside`, `c6-mask-over-later-mask`, `c4-attach-short-edge-ok`,
  `c4-attach-spread`, `c4-fanout-shared-origin`, `c4-tree-stem`, `legend-chip-near-connector`,
  `badge-chip-near-connector`, `path-relative-form`, `defs-marker-noise`, `path-cubic-unreadable` —
  *When* the linter runs on each, *Then* each reports `0 errors, 0 warnings` and exits 0. The loop
  is driven off the literal empty-set header pinned by REQ-SGC-014, so it selects exactly the
  empty-set fixtures and no others; running it over every fixture would include the must-flag
  fixtures other ACs require to emit, and could never pass. **The selection count is asserted before
  the loop runs**: a zero-match selection would execute zero commands and exit 0, so this AC — which
  carries the SPEC's whole no-false-positive posture — could otherwise pass while checking nothing.
  Commands:
  `N=$(grep -l 'expect: *-->' $F/*.svg | wc -l); test "$N" -gt 0 && test "$N" -eq "$CLEAN_N"` —
  `CLEAN_N` is the clean-fixture count fixed when the M1 inventory is complete, and is updated with
  it; the `-gt 0` guard holds regardless of that value, so an empty selection fails here rather than
  passing silently below ·
  `for f in $(grep -l 'expect: *-->' $F/*.svg); do node $S/scripts/check-svg.mjs "$f"; done`
- **AC-SGC-005** — *Given* the pre-existing `$F/a11y-missing.svg`, *When* the linter runs on it,
  *Then* it exits 1 and emits exactly the code set declared in its expectation header.
- **AC-SGC-006** — *Given* `$F/defs-marker-noise.svg`, whose `<defs><marker>` carries the
  **sketch-mode arrowhead verbatim** — `<path d="M 0 1 L 12 5 L 0 9" fill="none" stroke="…"/>` per
  `sketch.md` §3, whose endpoints are 8 apart on `x = 0` and would bind to a background rect's left
  edge if admitted — plus a `<symbol>` and a `<clipPath>` each carrying a `<rect>`, *When* the
  linter runs, *Then* the emitted code set is exactly empty. The `fill="none"` is load-bearing: a
  filled arrowhead would be caught by REQ-SGC-009(b) and a B0 regression would pass unnoticed, so
  the fixture pins the one shape for which (a) is the only guard.

### D.3 — C2 (SVG070 / SVG073)

- **AC-SGC-007** — *Given* `$F/c2-mask-too-close.svg` (mask 4 units from its stroke) and
  `$F/c2-mask-too-close-hop.svg` (the same breach on a connector carrying an `A` crossing hop),
  *When* the linter runs on each, *Then* each emits exactly `{SVG070}` at error level. The hop
  fixture is REQ-SGC-010's positive assertion for `A`: a reader that failed to parse it would
  recover no polyline, associate no mask, and emit nothing — so the presence of the code, not its
  absence, is what proves the command was read. The distance-0 crossing case is asserted by its own
  § D.9 row and fixture.
- **AC-SGC-008** — *Given* `$F/c2-mask-too-far.svg` (mask 16 units from its stroke), *When* the
  linter runs, *Then* the emitted code set is exactly `{SVG073}` at warning level, and the process
  exits 0 without `--strict` and 1 with `--strict`.

### D.4 — C6 three-way discrimination (SVG071)

- **AC-SGC-009** — *Given* `$F/c6-mask-partial.svg` (mask overlapping a later-painted rect on one
  edge), *When* the linter runs, *Then* the emitted code set is exactly `{SVG071}`. The
  fully-inside (badge chip), fully-outside, and earlier-in-document-order cases are clean fixtures
  under AC-SGC-004 with their own § D.9 rows — the three-way discrimination is asserted by the
  contrast between this AC and those fixtures. `c6-mask-over-later-mask.svg`, where two label masks
  partially overlap each other, is clean under AC-SGC-004: C6 constrains overlap with a node, so
  mask-over-mask is outside it (REQ-SGC-003).

### D.5 — C4 floor branches (SVG072)

- **AC-SGC-010** — *Given* `$F/c4-attach-crowded.svg` (edge ≥ 120, **arrival** points 9 units apart,
  routed as elbows so each `d` carries a `Q` corner) and `$F/c4-attach-short-edge-bad.svg`
  (edge < 120, arrivals 6 units apart) and `$F/c4-attach-coincident.svg` (two arrivals at the same
  point), *When* the linter runs on each, *Then* each emits exactly `{SVG072}`. Three properties are
  asserted by this set: the 12-vs-8 branch, by the contrast between the first fixture and
  `c4-attach-short-edge-ok.svg` (edge < 120, the *same* 9-unit separation, clean under AC-SGC-004) —
  the same separation failing on a long edge and passing on a short one is what proves the branch is
  live; REQ-SGC-010's positive handling of `Q`, `H`, and `V`, since a reader that dropped the elbow
  would recover no arrival point and emit nothing; and REQ-SGC-004's arrival-only binding, by the
  contrast with `c4-fanout-shared-origin.svg` and `c4-tree-stem.svg` — N connectors departing from
  one identical point, clean under AC-SGC-004 (§B D7).
  Two further fixtures make the arrival semantics falsifiable rather than merely silent, since a
  clean result would otherwise be indistinguishable from a reader, a marker lookup, or `SVG072`
  itself never having been wired: `$F/c4-closed-both-markers.svg` — a closed `fill="none"` path
  carrying both `marker-start` and `marker-end`, whose two coincident arrivals bind one edge —
  emits exactly `{SVG072}`; and `$F/c4-markerless-pair.svg`, two connectors carrying no marker whose
  endpoints sit 4 apart on one edge, is clean under AC-SGC-004 while its twin
  `$F/c4-marker-pair-emits.svg` — the same geometry with `marker-end` added to both — emits exactly
  `{SVG072}`. The twin pair is what proves the silence is the marker rule and not a dead check.

### D.6 — Transform aggregate (SVG074)

- **AC-SGC-011** — *Given* `$F/transform-skipped.svg`, where three sibling elements carry
  `transform`, and `$F/transform-wrapper.svg`, a full diagram wrapped in a single
  `<g transform="translate(…)">`, *When* the linter runs on each, *Then* each emits exactly
  `{SVG074}` — one aggregate note per file, never one per element — positioned at the root `<svg>`
  offset, exiting 0 by default and 1 under `--strict` (the accepted constraint-K6 consequence,
  asserted rather than discovered). The wrapper fixture additionally pins the count semantics of
  REQ-SGC-006: because `hasTransform` walks ancestors, every element in it is excluded, so the
  message must report the transitively-excluded count against the total candidate population — a
  message reading "1 element skipped" while none of the diagram was checked fails this AC.

### D.7 — Single measurement path

- **AC-SGC-012** — *Given* the skill's `scripts/` directory, *When* it is inspected, *Then* the SVG
  parsing helpers appear only in `check-svg.mjs` (the runner spawns the CLI and re-implements no
  parsing), every import in both scripts resolves to a Node built-in (`node:fs`, `node:path`,
  `node:child_process`, `node:url`), and no `package.json`, lockfile, or `node_modules` has been
  added under `$S`.
  Commands: `grep -rn 'tokenize\|buildTree' $S/scripts/` · `grep -n '^import\|require(' $S/scripts/*.mjs` · `ls $S`

### D.8 — Docs, mirror, and build

- **AC-SGC-013** — *Given* the three documentation surfaces, *When* they are grepped, *Then*
  `check-svg.mjs`'s header comment documents the SVG07x tier, `SKILL.md` carries a script-table row
  for `scripts/test-check-svg.mjs` plus an updated fixtures row plus a slop-symptom row 12 naming
  the codes, and `authoring.md` §2.5 cites the code for each of C2 / C4 / C6 with §8.3 naming the
  runner.
  Commands: `sed -n '1,40p' $S/scripts/check-svg.mjs | grep -n 'SVG07'` · `grep -n 'test-check-svg\|SVG07' $S/SKILL.md` · `grep -n 'SVG07\|test-check-svg' $S/references/authoring.md`
- **AC-SGC-014** — *Given* both trees after `make build`, *When*
  `diff -rq --exclude=.moai $S $T` is run, *Then* it reports no differences and exits 0.
- **AC-SGC-015** — *Given* the mirrored tree, *When*
  `grep -rnE 'SPEC-[A-Z0-9-]+-[0-9]{3}|\bt166\b|2026-[0-9]{2}-[0-9]{2}|[0-9a-f]{9,40}' $T` is run
  over the files this SPEC added or changed, *Then* no SPEC ID, card id, internal date, or commit
  SHA appears.
- **AC-SGC-016** — *Given* the repository, *When* `make build` then
  `go test ./internal/template/...` are run, *Then* both exit 0. The full-suite verdict is CI's; a
  local `go test ./...` is not run.

---

## §D.9 Edge cases — each row is a fixture, asserted by its exact-code expectation header

Every row below is realised as a fixture (or as a case inside one) whose `<!-- expect: … -->` header
names the exact code set the runner asserts. This table is the boundary-value register the numbered
ACs above delegate to; it is not a prose annex.

Rows whose expected column reads as prose ("binds to that edge", "not a mask candidate", "12-unit
floor applies") are `{}` fixtures — the prose names the *mechanism* that makes them clean, and the
runner asserts the empty set. That is sufficient where a competing mechanism would produce a code,
and insufficient where the row asserts a silence, since silence is indistinguishable from a check
that was never wired. Every row of the second kind is therefore paired with a must-flag twin
differing in exactly the property under test — the marker-less pair and its `marker-end` twin, the
one-marker closed path and its both-markers twin.

| Case | Expected code set |
|---|---|
| Mask 5 units from its stroke | `{SVG070}` |
| Mask touching or crossing its stroke (distance 0) | `{SVG070}` |
| Mask at exactly 6 units | `{}` |
| Mask at exactly 10 units | `{}` |
| Mask at 11 units | `{SVG073}` |
| Attach separation exactly 12 on an edge ≥ 120 | `{}` |
| Attach separation 9 on an edge ≥ 120 | `{SVG072}` |
| Attach separation 9 on an edge < 120 | `{}` |
| Attach separation exactly 8 on an edge < 120 | `{}` |
| Attach separation 6 on an edge < 120 | `{SVG072}` |
| Edge length exactly 120 | 12-unit floor applies |
| Later-painted rect partially overlapping the mask | `{SVG071}` |
| Later-painted rect fully containing the mask (badge chip) | `{}` |
| Mask fully containing a later-painted rect | `{SVG071}` |
| Mask overlapping an *earlier*-painted rect (container / band) | `{}` |
| Endpoint 10 units off the edge (markerLen) | binds to that edge |
| Endpoint 11 units off the edge | binds to no edge, no `SVG072` contribution |
| Endpoint whose projection falls outside the edge span | binds to no edge |
| `<rect>` followed by `<text>` that is a connector attach target | not a mask candidate |
| `<rect>` with no adjacent `<text>` sibling | not a mask candidate (accepted false negative) |
| `<path>` / `<rect>` inside `<defs>`, `<marker>`, `<symbol>`, `<clipPath>`, `<mask>`, `<pattern>` | excluded from every candidate set |
| `<path>` with `fill` other than `none` | not a connector |
| Connector `d` using a `C` cubic the reader cannot interpret | `{}` — skipped silently |
| Three sibling elements carrying `transform` | `{SVG074}` — one note, not three |
| Whole diagram wrapped in one `<g transform>` | `{SVG074}` — count is the transitively-excluded population, not 1 |
| Document with zero transformed elements | no `SVG074` |
| Fan-out family: N connectors departing one identical point, arriving at distinct points | `{}` — arrival-only (§B D7) |
| Tree stem: N child connectors departing one identical point on a parent's bottom edge | `{}` — arrival-only (§B D7) |
| Two connectors arriving at the same point on one edge | `{SVG072}` |
| Two marker-less connectors whose endpoints sit 4 apart on one edge | `{}` — no arrival point |
| The same geometry with `marker-end` on both (the twin) | `{SVG072}` — proves the silence above is the marker rule, not a dead check |
| Closed path (`z`) carrying exactly one marker, first and last point coincident | `{}` — one arrival point, not a zero-separation pair |
| Closed path (`z`) carrying **both** markers, arrivals coincident | `{SVG072}` — two arrowheads on one spot |
| Mask on the convex side of a `Q` elbow at exactly 6 units | `{}` — the subdivision removes the ≈2.83-unit inward bias |
| Elbow connector (`Q` corner) whose two arrivals are 9 apart on a long edge | `{SVG072}` — proves `Q` was parsed |
| Hop connector (`A`) with a mask 4 units from its stroke | `{SVG070}` — proves `A` was parsed |
| Connector authored in relative form with implicit repeated pairs | `{}` — parsed, compliant |
| Mask 11-16 units from its connector | `{SVG073}` — inside the association window |
| Non-label `<rect>`+`<text>` pair (legend swatch, badge chip) 20 units from a connector | `{}` — outside the 16-unit window, associates to nothing |
| Non-label `<rect>`+`<text>` pair 11-16 units from a connector | `{SVG073}` — the residual exposure the window narrows but does not remove (plan.md §B) |
| Two label masks partially overlapping each other | `{}` — C6 constrains overlap with a node |
| Mask partially overlapping a later decorative divider or rule that is not a mask candidate | `{SVG071}` |

---

## §D.10 Quality gates

- No new runtime dependency; Node 18 stdlib only (AC-SGC-012).
- Existing diagnostics unchanged: `a11y-present.svg` stays clean, `a11y-missing.svg` keeps its code
  set (AC-SGC-004, AC-SGC-005).
- Template mirror parity and neutrality (AC-SGC-014, AC-SGC-015).
- `make build` green (AC-SGC-016); full-suite verdict deferred to CI.

---

## §D.11 Definition of Done

1. `node $S/scripts/test-check-svg.mjs` exits 0 with every fixture `PASS`.
2. Every AC in §D above has a recorded command and its verbatim output.
3. `diff -rq --exclude=.moai $S $T` is clean and `make build` exits 0.
4. Commits on this branch name card `t166`; evidence lives under `.moai/reports/t166/`.
5. CI is green on the pushed branch — the full-suite verdict.

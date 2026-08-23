# SPEC-SVG-GEOMETRY-CHECKS-001 — Progress

Card: t166 · Worktree: `.claude/worktrees/t166` · Branch: `WT-verify-geometry`

## §E.1 Plan-phase Audit-Ready Signal

- Artifacts authored: `spec.md`, `plan.md`, `acceptance.md`, `progress.md` (Tier M set).
- SPEC ID regex check executed as Bash: `PASS`.
- Frontmatter: all 12 canonical fields present; `status: draft`; `phase: "v3.1.3 target"`.
- Out of Scope: four `### Out of Scope — …` sub-headings with bullets (spec.md §D).
- Requirements: 16 REQ entries in GEARS notation (spec.md §C) — at the Tier M ceiling of 16.
- Acceptance criteria: 16 `AC-SGC-xxx` Given-When-Then entries — at the Tier M ceiling of 16 — plus the §D.9 boundary-value table
  (acceptance.md §D).
- Premises of the delegation prompt were re-verified against the tree: script length 609 lines,
  fixture pair present and behaving 0/1, mirror `diff -rq` clean. The "§2.5 rule text settled"
  premise proved **partially false** and is now recorded as spec.md §B D7 rather than assumed.
- Audit iteration 1: FAIL 0.74 (Tier M threshold 0.80), nine blocking findings. Iteration 2 applied
  all nine plus the routed optional findings, entirely as clauses of the existing 16 requirements.
- Audit iteration 2: **PASS 0.81**, no dimension regressing, with four blocking-class defects
  recorded as mandatory pre-run-phase debt (N1-N4) plus four optional (N5-N8). All eight are cleared
  in v0.3.0, again as sentence-level edits inside existing entries — counts unchanged at 16 REQ /
  16 AC, ids contiguous 001-016, no dangling cross-reference, no stale `§ D.8`.
  Every measurement either auditor cited was independently reproduced before the fix was applied.
- Debt status at run-phase entry: **cleared**. Remaining accepted costs are stated, not latent —
  K3's bounded 11-16 unit exception, the departure-side and marker-less C4 blind spots (§A), and
  the C2 association window's reach (§A).

## §E.2 Run-phase Evidence

### M1 — fixture harness (RED, expected)

Baseline for every measurement below: this run, this tree, branch `WT-verify-geometry`, HEAD
`0dcc945d7` before the M1 commit. Node `v22.14.0`. `check-svg.mjs` is **unmodified** in M1 — the
whole point of the milestone is that the must-flag fixtures cannot pass yet.

**Files touched.** `scripts/test-check-svg.mjs` (new), 40 new fixtures under `scripts/fixtures/`,
and the expectation headers backfilled onto the two pre-existing fixtures. Nothing else: the linter,
`render.mjs`, `SKILL.md`, `references/`, and the template mirror are untouched (M2 and M3 scope).

**Pre-flight, before any file was written.**

```
$ node --version
v22.14.0
$ node $S/scripts/check-svg.mjs $F/a11y-present.svg; echo "rc=$?"
0 errors, 0 warnings
rc=0
$ node $S/scripts/check-svg.mjs $F/a11y-missing.svg; echo "rc=$?"
... 4 errors, 0 warnings
rc=1
$ node $S/scripts/check-svg.mjs $F/a11y-missing.svg --json | grep -o '"code": "[A-Z0-9]*"' | sort -u
"code": "SVG060"
"code": "SVG061"
"code": "SVG062"
"code": "SVG063"
```

`a11y-missing.svg`'s expectation header carries that **measured** set — `SVG060 SVG061 SVG062
SVG063` — not the set inferred from its own prose comment or from the SPEC. `SVG064` is not in it.

**E1 — fixture inventory.** 42 fixtures. Every row's "current linter" column is the code set the
unmodified `check-svg.mjs` emits today, read from `--json`.

| Fixture | Declared header | Current linter emits | Intent |
|---|---|---|---|
| `a11y-present` | `<!-- expect: -->` | `{}` | clean (pre-existing, unchanged) |
| `a11y-missing` | `SVG060 SVG061 SVG062 SVG063` | `{SVG060, SVG061, SVG062, SVG063}` | pre-existing, measured set |
| `c2-mask-too-close` | `SVG070` | `{}` | **RED** |
| `c2-mask-touching` | `SVG070` | `{}` | **RED** |
| `c2-mask-too-close-hop` | `SVG070` | `{}` | **RED** |
| `c2-mask-at-11` | `SVG073` | `{}` | **RED** |
| `c2-mask-too-far` | `SVG073` | `{}` | **RED** |
| `chip-in-window` | `SVG073` | `{}` | **RED** |
| `c6-mask-partial` | `SVG071` | `{}` | **RED** |
| `c6-mask-contains-later` | `SVG071` | `{}` | **RED** |
| `c4-attach-crowded` | `SVG072` | `{}` | **RED** |
| `c4-attach-short-edge-bad` | `SVG072` | `{}` | **RED** |
| `c4-attach-coincident` | `SVG072` | `{}` | **RED** |
| `c4-edge-exactly-120` | `SVG072` | `{}` | **RED** |
| `c4-marker-pair-emits` | `SVG072` | `{}` | **RED** |
| `c4-closed-both-markers` | `SVG072` | `{}` | **RED** |
| `transform-skipped` | `SVG074` | `{}` | **RED** |
| `transform-wrapper` | `SVG074` | `{}` | **RED** |
| `c2-mask-clear` | `<!-- expect: -->` | `{}` | clean |
| `c2-mask-at-6` | `<!-- expect: -->` | `{}` | clean |
| `c2-mask-at-10` | `<!-- expect: -->` | `{}` | clean |
| `c2-mask-outer-elbow-corner-at-6` | `<!-- expect: -->` | `{}` | clean |
| `legend-chip-near-connector` | `<!-- expect: -->` | `{}` | clean |
| `badge-chip-near-connector` | `<!-- expect: -->` | `{}` | clean |
| `c6-mask-inside` | `<!-- expect: -->` | `{}` | clean |
| `c6-mask-outside` | `<!-- expect: -->` | `{}` | clean |
| `c6-mask-over-earlier-rect` | `<!-- expect: -->` | `{}` | clean |
| `c6-mask-over-later-mask` | `<!-- expect: -->` | `{}` | clean |
| `c4-attach-short-edge-ok` | `<!-- expect: -->` | `{}` | clean |
| `c4-attach-at-8` | `<!-- expect: -->` | `{}` | clean |
| `c4-attach-at-12` | `<!-- expect: -->` | `{}` | clean |
| `c4-attach-spread` | `<!-- expect: -->` | `{}` | clean |
| `c4-fanout-shared-origin` | `<!-- expect: -->` | `{}` | clean |
| `c4-tree-stem` | `<!-- expect: -->` | `{}` | clean |
| `c4-markerless-pair` | `<!-- expect: -->` | `{}` | clean |
| `attach-endpoint-11-off` | `<!-- expect: -->` | `{}` | clean |
| `attach-projection-outside-span` | `<!-- expect: -->` | `{}` | clean |
| `mask-no-adjacent-text` | `<!-- expect: -->` | `{}` | clean |
| `no-transform` | `<!-- expect: -->` | `{}` | clean |
| `path-relative-form` | `<!-- expect: -->` | `{}` | clean |
| `path-cubic-unreadable` | `<!-- expect: -->` | `{}` | clean |
| `defs-marker-noise` | `<!-- expect: -->` | `{}` | clean |

16 must-flag, 25 empty-set, 1 pre-existing non-empty set. **`CLEAN_N` = 25** — the value
AC-SGC-004's selection-count assertion is fixed to at the close of the M1 inventory.

Every fixture's geometry is stated in a second in-file comment naming the number from
`authoring.md` §2.5 it encodes and the boundary it sits on. Two are worth restating here because
they were computed rather than chosen: `c2-mask-outer-elbow-corner-at-6` places its mask corner at
`(142.5, 57.5)`, which is **6.36** units from the subdivided `Q` polyline and **3.54** from the
un-subdivided control-point polyline — clean under REQ-SGC-010's subdivision, `SVG070` without it.
`c4-attach-crowded` and `c4-attach-short-edge-ok` carry the **same** 9-unit arrival separation on a
120-unit and an 80-unit edge respectively, so the 12-vs-8 branch is asserted by the contrast rather
than by either fixture alone.

**E8 — RED evidence (load-bearing).** Verbatim, against the unmodified `check-svg.mjs`:

```
$ node .claude/skills/moai-domain-svg-infographic/scripts/test-check-svg.mjs; echo "rc=$?"
PASS  a11y-missing.svg  {SVG060, SVG061, SVG062, SVG063}
PASS  a11y-present.svg  {}
PASS  attach-endpoint-11-off.svg  {}
PASS  attach-projection-outside-span.svg  {}
PASS  badge-chip-near-connector.svg  {}
PASS  c2-mask-at-10.svg  {}
FAIL  c2-mask-at-11.svg  expected {SVG073} but got {}
PASS  c2-mask-at-6.svg  {}
PASS  c2-mask-clear.svg  {}
PASS  c2-mask-outer-elbow-corner-at-6.svg  {}
FAIL  c2-mask-too-close-hop.svg  expected {SVG070} but got {}
FAIL  c2-mask-too-close.svg  expected {SVG070} but got {}
FAIL  c2-mask-too-far.svg  expected {SVG073} but got {}
FAIL  c2-mask-touching.svg  expected {SVG070} but got {}
PASS  c4-attach-at-12.svg  {}
PASS  c4-attach-at-8.svg  {}
FAIL  c4-attach-coincident.svg  expected {SVG072} but got {}
FAIL  c4-attach-crowded.svg  expected {SVG072} but got {}
FAIL  c4-attach-short-edge-bad.svg  expected {SVG072} but got {}
PASS  c4-attach-short-edge-ok.svg  {}
PASS  c4-attach-spread.svg  {}
FAIL  c4-closed-both-markers.svg  expected {SVG072} but got {}
FAIL  c4-edge-exactly-120.svg  expected {SVG072} but got {}
PASS  c4-fanout-shared-origin.svg  {}
FAIL  c4-marker-pair-emits.svg  expected {SVG072} but got {}
PASS  c4-markerless-pair.svg  {}
PASS  c4-tree-stem.svg  {}
FAIL  c6-mask-contains-later.svg  expected {SVG071} but got {}
PASS  c6-mask-inside.svg  {}
PASS  c6-mask-outside.svg  {}
PASS  c6-mask-over-earlier-rect.svg  {}
PASS  c6-mask-over-later-mask.svg  {}
FAIL  c6-mask-partial.svg  expected {SVG071} but got {}
FAIL  chip-in-window.svg  expected {SVG073} but got {}
PASS  defs-marker-noise.svg  {}
PASS  legend-chip-near-connector.svg  {}
PASS  mask-no-adjacent-text.svg  {}
PASS  no-transform.svg  {}
PASS  path-cubic-unreadable.svg  {}
PASS  path-relative-form.svg  {}
FAIL  transform-skipped.svg  expected {SVG074} but got {}
FAIL  transform-wrapper.svg  expected {SVG074} but got {}
26/42 fixtures matched
rc=1
```

Exactly the 16 must-flag fixtures fail, each because the SVG07x tier does not exist yet. `rc=1` is
M1's deliverable, not a defect: a runner exiting 0 here would mean the fixtures assert nothing, or
that M2 was implemented early.

**Clean-fixture check (AC-SGC-004, run at M1 to catch fixture bugs early).** Every empty-set fixture
must already be clean today, since none of its expected codes exists yet.

```
$ grep -l 'expect: *-->' $F/*.svg | wc -l
      25
$ grep -l 'expect: *-->' $F/*.svg | xargs -n1 node $S/scripts/check-svg.mjs
0 errors, 0 warnings          (× 25 — every line identical, no other output)
```

No clean fixture emits anything, so no fixture bug was hidden by the RED signal.

**Runner self-test (AC-SGC-002) — the runner can actually fail.** Performed on a copy of `scripts/`
outside the repository (session scratch dir), with `c2-mask-clear.svg`'s header mutated from
`<!-- expect: -->` to `<!-- expect: SVG070 -->`:

```
$ node <scratch>/scripts-scratch/test-check-svg.mjs > <scratch>/mutated.txt; echo "rc=$?"
rc=1
$ grep -E "c2-mask-clear|matched" <scratch>/mutated.txt
FAIL  c2-mask-clear.svg  expected {SVG070} but got {}
25/42 fixtures matched
```

The mutation moved the count 26 → 25 and kept `rc=1`. Mutation not left behind — the repository copy
still reads `<!-- expect: -->` on line 1 and contains zero occurrences of `SVG070`; the scratch tree
was removed.

**AC-SGC-003 (exact-set comparison, not exit code).**

```
$ grep -n 'code' $S/scripts/test-check-svg.mjs
62:  return { codes: diagnostics.map((d) => d.code) };
65,66,69,70,91: set construction and comparison over those codes
17-19: "The child process exit code is never consulted"
```

The runner never reads the child's `status`; a JSON-parse failure is the only crash path and it is
reported as `FAIL`, not as a pass.

**AC-SGC-012 (single measurement path, Node stdlib only).**

```
$ grep -n '^import' $S/scripts/test-check-svg.mjs
node:fs, node:child_process, node:path, node:url        (all built-ins)
$ ls $S
.moai  SKILL.md  references  scripts                    (no package.json, no lockfile, no node_modules)
```

The runner spawns `check-svg.mjs --json` and re-implements no parsing: `tokenize` / `buildTree`
appear only in `check-svg.mjs`.

**Existing diagnostics unchanged.** After the header backfill, `a11y-present.svg` still reports
`0 errors, 0 warnings` and exits 0; `a11y-missing.svg` still exits 1 with the same four codes.

**Gaps (explicitly NOT observed at M1).**

- No SVG07x behaviour is verified, because none exists — the fixtures' *declared* code sets are
  assertions about M2, not measurements. M1 measures only that they are currently unmet and that no
  fixture trips an existing check.
- `AC-SGC-011`'s message criterion (the transitively-excluded count in the `SVG074` text) is not
  asserted here: the runner compares codes, and the count is read by a human at M2.
- No Go build, no `make build`, no template-mirror diff, no lint — nothing Go or mirror-side changed
  in M1 (M3 scope).
- Not pushed and no PR opened, per the delegation.

**Residual risk.**

- A must-flag fixture could fail at M2 for a reason other than the one it names (a second, unintended
  breach in the same file). Mitigated by the clean-fixture sweep above — every fixture is otherwise
  silent under the current linter — but not eliminated, since a *new* check could fire on geometry
  that today's checks ignore.
- `c2-mask-outer-elbow-corner-at-6`'s 6.36-unit clearance was computed against the polyline
  REQ-SGC-010 specifies (endpoints plus t = 0.25/0.5/0.75). An M2 reader that samples differently
  would measure a different number, and at 6.36 the margin over the 6-unit floor is 0.36.
- `transform-wrapper` and `transform-skipped` pin the `SVG074` code only. A per-element implementation
  emitting three notes would still produce the single code `SVG074` and pass the runner; only
  AC-SGC-011's human read of the message catches that.

### M2 — geometry engine and the five diagnostics (GREEN)

Baseline for every measurement below: this run, this tree, branch `WT-verify-geometry`, HEAD
`a4e3c8a33` (the M1 commit) before the M2 commit. Node `v22.14.0`.

**Files touched.** `scripts/check-svg.mjs` only (609 → 1020 lines) plus this `progress.md` block.
No fixture, no runner, no `render.mjs`, no `SKILL.md`, no `references/`, no template mirror — the
`git status --short` below is the evidence.

**Claim 1 — RED → GREEN.** The runner went from `26/42, rc=1` to `42/42, rc=0` with no fixture edit.

RED, verbatim, captured before any edit (16 must-flag fixtures reporting `got {}`; tail shown):

```
$ node $S/scripts/test-check-svg.mjs
...
FAIL  c2-mask-at-11.svg  expected {SVG073} but got {}
FAIL  c2-mask-too-close-hop.svg  expected {SVG070} but got {}
FAIL  c2-mask-too-close.svg  expected {SVG070} but got {}
FAIL  c2-mask-too-far.svg  expected {SVG073} but got {}
FAIL  c2-mask-touching.svg  expected {SVG070} but got {}
FAIL  c4-attach-coincident.svg  expected {SVG072} but got {}
FAIL  c4-attach-crowded.svg  expected {SVG072} but got {}
FAIL  c4-attach-short-edge-bad.svg  expected {SVG072} but got {}
FAIL  c4-closed-both-markers.svg  expected {SVG072} but got {}
FAIL  c4-edge-exactly-120.svg  expected {SVG072} but got {}
FAIL  c4-marker-pair-emits.svg  expected {SVG072} but got {}
FAIL  c6-mask-contains-later.svg  expected {SVG071} but got {}
FAIL  c6-mask-partial.svg  expected {SVG071} but got {}
FAIL  chip-in-window.svg  expected {SVG073} but got {}
FAIL  transform-skipped.svg  expected {SVG074} but got {}
FAIL  transform-wrapper.svg  expected {SVG074} but got {}
26/42 fixtures matched
rc=1
```

GREEN, verbatim, after the engine landed:

```
$ node $S/scripts/test-check-svg.mjs
PASS  a11y-missing.svg  {SVG060, SVG061, SVG062, SVG063}
PASS  a11y-present.svg  {}
PASS  attach-endpoint-11-off.svg  {}
PASS  attach-projection-outside-span.svg  {}
PASS  badge-chip-near-connector.svg  {}
PASS  c2-mask-at-10.svg  {}
PASS  c2-mask-at-11.svg  {SVG073}
PASS  c2-mask-at-6.svg  {}
PASS  c2-mask-clear.svg  {}
PASS  c2-mask-outer-elbow-corner-at-6.svg  {}
PASS  c2-mask-too-close-hop.svg  {SVG070}
PASS  c2-mask-too-close.svg  {SVG070}
PASS  c2-mask-too-far.svg  {SVG073}
PASS  c2-mask-touching.svg  {SVG070}
PASS  c4-attach-at-12.svg  {}
PASS  c4-attach-at-8.svg  {}
PASS  c4-attach-coincident.svg  {SVG072}
PASS  c4-attach-crowded.svg  {SVG072}
PASS  c4-attach-short-edge-bad.svg  {SVG072}
PASS  c4-attach-short-edge-ok.svg  {}
PASS  c4-attach-spread.svg  {}
PASS  c4-closed-both-markers.svg  {SVG072}
PASS  c4-edge-exactly-120.svg  {SVG072}
PASS  c4-fanout-shared-origin.svg  {}
PASS  c4-marker-pair-emits.svg  {SVG072}
PASS  c4-markerless-pair.svg  {}
PASS  c4-tree-stem.svg  {}
PASS  c6-mask-contains-later.svg  {SVG071}
PASS  c6-mask-inside.svg  {}
PASS  c6-mask-outside.svg  {}
PASS  c6-mask-over-earlier-rect.svg  {}
PASS  c6-mask-over-later-mask.svg  {}
PASS  c6-mask-partial.svg  {SVG071}
PASS  chip-in-window.svg  {SVG073}
PASS  defs-marker-noise.svg  {}
PASS  legend-chip-near-connector.svg  {}
PASS  mask-no-adjacent-text.svg  {}
PASS  no-transform.svg  {}
PASS  path-cubic-unreadable.svg  {}
PASS  path-relative-form.svg  {}
PASS  transform-skipped.svg  {SVG074}
PASS  transform-wrapper.svg  {SVG074}
42/42 fixtures matched
rc=0
```

**Claim 2 — existing behaviour unchanged.** `a11y-present.svg` still exits 0 clean;
`a11y-missing.svg` still exits 1 with the same four codes.

```
$ node $S/scripts/check-svg.mjs $F/a11y-present.svg; echo "rc=$?"
0 errors, 0 warnings
rc=0
$ node $S/scripts/check-svg.mjs $F/a11y-missing.svg --json | grep '"code"'
      "code": "SVG060",
      "code": "SVG062",
      "code": "SVG063",
      "code": "SVG061",
$ node $S/scripts/check-svg.mjs $F/a11y-missing.svg > /dev/null; echo "rc=$?"
rc=1
```

**Claim 3 — `--strict` is respected (REQ-SGC-007).** A warning-only fixture exits 0 by default and
1 under `--strict`.

```
$ node $S/scripts/check-svg.mjs $F/c2-mask-too-far.svg; echo "rc=$?"
...:14:3  warning  SVG073  label mask stands 16.0 units off its connector; C2 keeps it within 10
0 errors, 1 warning
rc=0
$ node $S/scripts/check-svg.mjs $F/c2-mask-too-far.svg --strict; echo "rc=$?"
... same line ...
rc=1
$ node $S/scripts/check-svg.mjs $F/transform-wrapper.svg --strict > /dev/null; echo "rc=$?"
rc=1
```

The last line is constraint K6 realised: a transformed diagram now fails `--strict` where it
previously passed.

**Claim 4 — the runner can still fail (AC-SGC-002).** A scratch copy of `scripts/` outside the
repository, with one expectation header mutated, fails that fixture and exits 1.

```
$ sed -i '' '1s|.*|<!-- expect: SVG099 -->|' $SCRATCH/mut/fixtures/c2-mask-too-close.svg
$ node $SCRATCH/mut/test-check-svg.mjs
FAIL  c2-mask-too-close.svg  expected {SVG099} but got {SVG070}
41/42 fixtures matched
rc=1
```

The scratch copies were removed afterwards; the working tree is clean apart from the untracked
report directory:

```
$ git status --short
 M .claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs
?? .moai/reports/t166/
```

**Claim 5 — the ruling fixtures, each quoted.**

- `c2-mask-outer-elbow-corner-at-6` → `0 errors, 0 warnings`. Load-bearing: a scratch copy with the
  `Q` subdivision replaced by the control point reports
  `error SVG070 label mask clears its connector by 3.5 units`, which is the ≈3.54 figure `plan.md`
  §C predicts. The subdivision is what keeps this compliant diagram silent.
- `c4-fanout-shared-origin` → `0 errors, 0 warnings`; `c4-tree-stem` → `0 errors, 0 warnings`.
  Arrival-only binding: three connectors leave one identical point in each, and only arrivals are
  grouped. The twin `c4-marker-pair-emits` → `{SVG072}` against `c4-markerless-pair` → `{}` proves
  the silence is the marker rule and not a dead check.
- `c6-mask-over-later-mask` → `0 errors, 0 warnings` — mask-over-mask is excluded from `SVG071`'s
  later-rect set.
- `chip-in-window` → `warning SVG073 label mask stands 13.0 units off its connector` — K3's
  accepted bounded exception, emitted exactly as `acceptance.md` §D.9 declares.

**Claim 6 — `SVG074` count semantics (AC-SGC-011's message criterion).**

```
$ node $S/scripts/check-svg.mjs $F/transform-wrapper.svg
...:3:1  warning  SVG074  6 of 6 candidate elements carry a transform and were skipped; their geometry is unverified
$ node $S/scripts/check-svg.mjs $F/transform-skipped.svg
...:3:1  warning  SVG074  3 of 4 candidate elements carry a transform and were skipped; their geometry is unverified
```

One note per file in both cases, at the root `<svg>` offset (line 3, column 1). The wrapper reports
the transitively-excluded population (6 of 6), not the single element carrying the attribute.

**Claim 7 — single measurement path (REQ-SGC-008 / AC-SGC-012).** `check-svg.mjs` still carries
exactly one import, `node:fs`; no second parser, no manifest, no lockfile, no `node_modules` were
added.

```
$ grep -n '^import' $S/scripts/check-svg.mjs
31:import { readFileSync } from 'node:fs';
$ wc -l $S/scripts/check-svg.mjs
    1020
```

**Deviation, declared rather than buried — the attach-target standoff.** `spec.md` §B D2 excludes a
rect from the mask-candidate set when it is "an attach target of any connector endpoint", and
`plan.md` §B4 illustrates that clause at the `markerLen = 10` standoff. Implemented literally at 10,
`attach-endpoint-11-off.svg` fails: its node box sits 11 units from two connector endpoints, is
therefore not an attach target, becomes a mask candidate, associates at 11 units — inside the
16-unit window — and is warned `SVG073`. Measured, in a scratch copy with the standoff set to
`MARKER_LEN`:

```
FAIL  attach-endpoint-11-off.svg  expected {} but got {SVG073}
41/42 fixtures matched
```

That is a node box reported as a label mask on clean geometry — the false-positive class K3 names as
a defect. The engine therefore runs D2's attach test at `MASK_WINDOW` (16) instead, keeping
REQ-SGC-012's projection-in-span requirement intact and leaving REQ-SGC-012's own `markerLen`
standoff for `SVG072` binding untouched at 10. The fixture is not edited. The accepted cost is a
narrow false negative in D3's posture: a genuine label mask sitting within 16 units of a connector
endpoint is not checked.

**Gaps (explicitly NOT observed at M2).**

- No documentation surface was touched — `check-svg.mjs`'s header comment still describes only the
  SVG001-SVG064 tiers, and `SKILL.md` / `authoring.md` carry no SVG07x annotation. That is M3 scope
  (REQ-SGC-016), so AC-SGC-013 is unverified here.
- No template mirror, no `make build`, no `diff -rq` parity, no neutrality grep — M3 scope, so
  AC-SGC-014 / AC-SGC-015 / AC-SGC-016 are unverified here.
- No Go test of any kind was run; `go test ./...` is prohibited on this machine.
- AC-SGC-004's clean-fixture loop was not run as its own named command with the `CLEAN_N` count
  assertion. The runner asserts the same fixtures' exact code sets, which is the stricter check, but
  the AC's literal command form is not evidenced here.
- Not pushed, no PR opened, per the delegation.

**Residual risk.**

- The attach-standoff deviation above changes the mask-candidate set for every diagram, not only the
  fixture that forced it. Its false-negative reach is bounded by the 16-unit window, but no fixture
  pins a genuine label mask inside that window near an endpoint, so the boundary is argued rather
  than measured.
- `SVG071` compares axis-aligned rect extents only. A rect with `rx` rounding, or a mask overlapping
  a non-`<rect>` shape painted later, is outside what the check sees — the wider-than-C6 posture of
  §B D8 is about which rects are in scope, not about non-rect shapes.
- The `A` command is a straight pass-through, so a large-radius arc's true sweep is never measured;
  a mask inside such an arc's bulge would be measured against the chord. The documented idiom is
  `rHop = 5`, where the error is under a unit, but nothing enforces that radius.
- The candidate population feeding `SVG074`'s denominator is `<path>` / `<rect>` / `<text>` outside
  non-rendered subtrees. A diagram whose geometry is carried by `<circle>` or `<line>` would report a
  denominator that understates what went unverified.

### M3 — documentation, template mirror, build

Baseline for every measurement below: this run, this tree, branch `WT-verify-geometry`, HEAD
`f8e60d3ed` (the M2 commit) before the M3 commit. Node `v22.14.0`, Go toolchain via `make build`.

**Files touched.** Three documentation surfaces on the local side — `scripts/check-svg.mjs` (header
comment only), `SKILL.md`, `references/authoring.md` — plus the template mirror of the whole skill
subtree and `internal/template/catalog.yaml` (regenerated by `make build`). No diagnostic
behaviour, code path, level, message, or position changed; no fixture and no runner edited.

**Claim 1 — the runner is still green after the documentation edits.**

```
$ node .claude/skills/moai-domain-svg-infographic/scripts/test-check-svg.mjs
...
PASS  transform-wrapper.svg  {SVG074}
42/42 fixtures matched
rc=0
```

Baseline-attribution: same command and same 42-fixture corpus as M2's GREEN measurement, re-run in
this tree after the three documentation edits landed. AC-SGC-001 / AC-SGC-002 unchanged.

**Claim 2 — AC-SGC-013, the three documentation surfaces.**

```
$ sed -n '1,40p' .claude/skills/moai-domain-svg-infographic/scripts/check-svg.mjs | grep -n 'SVG07'
29:// (SVG070-SVG072). The SVG07x tier mechanises three of the six connector rules
30:// in references/authoring.md section 2.5: SVG070 and SVG073 for a label mask
32:// mask within 16 units of a connector; SVG071 for a mask partially overlapping a
33:// later-painted rect; SVG072 for two connector ARRIVAL points closer than the
35:// connectors carrying neither marker, are never compared. SVG074 warns once per
```

```
$ grep -n 'test-check-svg\|SVG07' .claude/skills/moai-domain-svg-infographic/SKILL.md
297:| `scripts/test-check-svg.mjs` | Runs every fixture through the lint and asserts each one's exact diagnostic code set; exits non-zero on the first mismatch |
363:| 12 | Any breach of the six connector rules of `authoring.md` section 2.5 — … Three of the six are machine-checked by `check-svg.mjs`, each within a stated bound: C2 as `SVG070` (clearance under 6 units) and `SVG073` (over 10), but only for a mask lying within 16 units of a connector, so a label placed further out — archetype A2's branch labels among them — is not checked; C6 as `SVG071`; C4 as `SVG072`, but on **arrival** points only, so crowding on the departure side, and any connector carrying neither `marker-end` nor `marker-start`, go unreported. C1, C3, and C5 stay eye-only. |
```

(Row 12 is elided at `…` above for width; the full line is in the file.) The `scripts/fixtures/` row
was also rewritten from "Two SVGs that exercise the accessible-name check in both directions" to 42
SVGs spanning the accessible contract and the connector-geometry checks, and the following
sentence's now-stale "Both scripts" count was corrected to "Every script".

```
$ grep -n 'SVG07\|test-check-svg' .claude/skills/moai-domain-svg-infographic/references/authoring.md
180:Machine-checked as `SVG070` (clearance under 6 units, a mask that touches or
181:crosses the stroke being the zero case) and `SVG073` (clearance over 10, a
215:Machine-checked as `SVG072`, on **arrival** points only — an endpoint at which a
255:Machine-checked as `SVG071`, deliberately wider than this rule as written: the
571:node scripts/test-check-svg.mjs   # every fixture, exact code set asserted
```

§2.5's C2, C4, and C6 each gained one annotation paragraph; §8.3 replaced its two manual
`check-svg.mjs` fixture invocations with the single runner command. **Status: PASS.**

**Claim 3 — AC-SGC-014, mirror parity.**

```
$ diff -rq --exclude=.moai .claude/skills/moai-domain-svg-infographic \
        internal/template/templates/.claude/skills/moai-domain-svg-infographic
rc=0
```

No output, exit 0. Before the mirror the same command listed 43 divergences (2 modified fixtures, 40
fixtures plus the runner present only locally, and `check-svg.mjs` differing) — that list was the
mirror inventory, and it is now empty. **Status: PASS.**

**Claim 4 — AC-SGC-015, neutrality over the mirrored files this SPEC added or changed.**

```
$ grep -rnE 'SPEC-[A-Z0-9-]+-[0-9]{3}|\bt166\b|2026-[0-9]{2}-[0-9]{2}|[0-9a-f]{9,40}' \
    internal/template/templates/.claude/skills/moai-domain-svg-infographic/SKILL.md \
    internal/template/templates/.claude/skills/moai-domain-svg-infographic/references/authoring.md \
    internal/template/templates/.claude/skills/moai-domain-svg-infographic/scripts/
internal/template/templates/.claude/skills/moai-domain-svg-infographic/SKILL.md:23:  updated: "2026-07-24"
```

Zero SPEC IDs, zero card ids, zero commit SHAs. The single date match is the skill's own
`metadata.updated` frontmatter field, and it is **not introduced by this SPEC** — the same line
exists byte-identically in the mirror at the M2 commit (read out of HEAD, line 23, identical text).
It was deliberately left alone: bumping it would *introduce* a date rather than remove one, and
rewriting skill frontmatter is outside M3's scope. **Status: PASS, with the pre-existing frontmatter
date disclosed rather than suppressed.**

**Claim 5 — AC-SGC-016, build and the template-package tests.**

```
$ make build
...
catalog.yaml updated successfully (12899 bytes)
go build -ldflags "… -X …version.Commit=f8e60d3ed …" -o bin/moai ./cmd/moai
rc=0
```

`make build` regenerated `internal/template/catalog.yaml` (the per-skill content hashes); that file
is part of this commit, because an unmirrored catalog fails CI parity.

```
$ go test ./internal/template/...
ok  	github.com/modu-ai/moai-adk/internal/template	20.835s
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.790s
?   	github.com/modu-ai/moai-adk/internal/template/scripts	[no test files]
rc=0
```

**Status: PASS.** The full-suite verdict remains CI's; a local `go test ./...` was not run.

**Documentation accuracy against the SPEC's declared blind spots.** REQ-SGC-016 requires the shipped
claim to match what the checker actually does, and §A names three gaps. Each is stated, not rounded
up:

- **C4 departure side (D7's stated cost)** — `authoring.md` §2.5 C4: "on **arrival** points only — an
  endpoint at which a `marker-end` or `marker-start` resolves … Two connectors *departing* one point
  are the deliberate fan-out of §2.3 and are not reported, so genuine crowding on the departure side
  goes uncaught". `SKILL.md` row 12: "C4 as `SVG072`, but on **arrival** points only, so crowding on
  the departure side … go unreported".
- **C4 marker-less connectors** — `authoring.md`: "a connector carrying neither marker has no arrival
  point and is never compared." `SKILL.md`: "any connector carrying neither `marker-end` nor
  `marker-start`, go unreported."
- **C2 16-unit association window, which excludes A2's own documented label placement** —
  `authoring.md`: "The check associates a mask to a connector only within **16 units**, so a label
  placed further from its stroke than that — archetype A2's branch labels, which sit about 99 units
  off their connector, among them — associates to none and is not checked at all." `SKILL.md`: "only
  for a mask lying within 16 units of a connector, so a label placed further out — archetype A2's
  branch labels among them — is not checked."

No surface states that C2, C4, or C6 is checked without qualification. `SVG071`'s wider-than-C6
posture (§B D8) is likewise stated as wider rather than as matching C6.

**Gaps (explicitly NOT observed at M3).**

- No Go test outside `./internal/template/...` was run, and no lint (`golangci-lint`) was run — the
  full verdict is CI's, per the delegation and `plan.md` §F M3.
- The template-neutrality CI guard (`template-neutrality-check.yaml`) was not executed locally; only
  the AC-SGC-015 grep was. That guard's own content-class list may be broader than this regex.
- The mirrored `check-svg.mjs` was not executed from the mirror path; parity rests on `diff -rq`
  byte-identity rather than on a second run.
- `render.mjs` was not exercised; no PNG was produced at any milestone.
- Not pushed, no PR opened, per the delegation.

**Residual risk.**

- Documentation and checker can drift independently: nothing mechanically asserts that `SKILL.md`
  row 12 or `authoring.md` §2.5 still describes the codes the engine emits. A later change to a
  diagnostic's scope would leave these paragraphs silently wrong while the runner stayed green.
- The `SKILL.md` fixtures row now names a literal count (42). Adding a fixture without updating the
  row makes it stale; nothing checks the number in the prose.
- `catalog.yaml` hashes were regenerated in this tree. If another lane also regenerates it, the merge
  conflicts on a generated file and must be resolved by re-running `make build`, never by
  hand-picking hunks.
- The pre-existing `metadata.updated` frontmatter date will keep matching any date-shaped neutrality
  scan of this skill until a card that owns skill frontmatter decides what that field should say.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-08-23"
run_commit_sha: "<pending — this block is written into the commit it would name; backfill at sync>"
run_status: PASS
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: not_performed_worktree_isolated
l44_post_push_fetch: not_applicable_no_push
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: "make build rc=0, go test ./internal/template/... rc=0 (this tree)"
  linux_amd64: deferred_to_ci
  windows_amd64: deferred_to_ci
total_run_phase_files: 49
m1_to_mN_commit_strategy: "three commits, one per milestone — M1 fixtures + runner (RED), M2 engine (GREEN), M3 docs + mirror + build"
```

Notes bounding the block above:

- `ac_pass_count: 16` counts AC-SGC-001..016 across all three milestones; M3 itself verified
  AC-SGC-013..016 and re-verified AC-SGC-001 / AC-SGC-002 by re-running the runner. The per-AC
  evidence lives in the M1, M2, and M3 blocks of §E.2.
- `run_commit_sha` is a placeholder for the structural reason stated inline, not an omission.
- `cross_platform_build` is honest about scope: only darwin/arm64 was built and tested here.
  `plan.md` §F M3 scopes Go verification to the packages `make build` touches and defers the full
  verdict to CI.
- `total_run_phase_files: 49` = 1 runner + 40 new fixtures + 2 modified fixtures + `check-svg.mjs` +
  `SKILL.md` + `authoring.md` + `catalog.yaml` + `progress.md`, counting the skill subtree once (the
  mirror is a byte-identical copy of it).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Input parameters: tier M; scope ~5 files changed plus ~20 fixtures added; domain count 1 (one skill
subsystem: a Node lint script, its fixtures, its references, and the template mirror); file language
mix JavaScript plus markdown plus SVG; concurrency benefit LOW (coding-heavy, and M2 depends on M1's
fixtures existing).

| Mode | Selected | Rationale |
|---|---|---|
| `direct` | no | Not trivial — a geometry engine, a reader, and ~20 fixtures |
| `serial` | **yes** | Coding-heavy, single domain, milestones strictly ordered (M1 RED gates M2 GREEN) |
| `fanout` | no | 1 domain, not research-heavy; Anthropic's coding-task parallelism caveat applies |
| `sweep` | no | Not ≥ ~30 files and not one uniform mechanical transform rule |

Decision: `serial`

Justification: the work is one coding-heavy change to one script inside one skill, and its
milestones are dependency-ordered rather than parallel — M1's fixtures must exist and fail before
M2's engine can be shown to turn them green. Fan-out would add reconciliation cost with no
independent work to reconcile, and the coding-task parallelism caveat argues against it directly.
Progression mode: semi-autonomous per operator selection at the Implementation Kickoff Approval
gate — the orchestrator reads evidence at each milestone boundary; no goal is armed.

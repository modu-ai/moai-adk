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

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

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

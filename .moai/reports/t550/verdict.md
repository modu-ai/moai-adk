# t550 — lane verdict (GH #1631, SPEC-GATE-OXLINT-DETECT-001)

card: t550
branch: WT-js-linter-detect
base: d060e0d136f173f07353e46a709382d08ae25dd2 (local develop at dispatch)
commits: 7d28ddddf, d608847c2, a879d56ea
measured by: lane-5 orchestrator, independently of the implementing agent's own report

## Claim

`moai gate` now detects oxlint on the Node lint axis. An oxlint project ran zero lint
steps before the change and runs one after it. eslint, biome, and the linter-free
scaffold are unchanged. The JavaScript linter list names oxlint in both copies of
`javascript.md`, and the rebuilt binary carries the new sentence in its embedded template.

## Evidence — same instrument as the baseline, both sides

Instrument, identical to `baseline.md`: a binary built from this tree, seven fixtures
differing only in linter config file, `npx`/`npm`/`node` shadowed by exit-0 stubs on PATH
so nothing reaches the network, and the gate's own run summary read as the measurement.

Post-change binary: `go build -o <scratchpad>/moai-after ./cmd/moai` (exit 0).
Per fixture: `CLAUDE_PROJECT_DIR=<fx> PATH=<stub>:/usr/bin:/bin:/usr/sbin:/sbin <scratchpad>/moai-after gate`

| fixture | config file | BEFORE (lint steps) | AFTER (lint steps) | after: which row executed |
|---|---|---|---|---|
| eslintprj | `eslint.config.js` | 1 | 1 | eslint executed; biome, oxlint both skipped |
| biomeprj | `biome.json` | 1 | 1 | biome executed; eslint, oxlint both skipped |
| oxlintprj | `.oxlintrc.json` | **0** | **1** | **oxlint executed** |
| ox2 | `.oxlintrc.jsonc` | **0** | **1** | **oxlint executed** |
| ox3 | `oxlint.config.ts` | **0** | **1** | **oxlint executed**; eslint skipped |
| ox4 | `oxlint.config.mts` | **0** | **1** | **oxlint executed**; eslint skipped |
| bareprj (CONTROL) | none | 0 | **0** | all three skipped — unchanged |

Verbatim, the two rows that carry the verdict:

```
# oxlintprj, after
  - eslint: skipped — none of its config files exist in the project dire
  - biome: skipped — none of its config files exist in the project direc
  - oxlint: executed in 8ms

# bareprj (control), after
  - eslint: skipped — none of its config files exist in the project dire
  - biome: skipped — none of its config files exist in the project direc
  - oxlint: skipped — none of its config files exist in the project dire
```

The control is the row that matters as much as the fix: a Node project with no linter
config still runs zero lint steps. Had it gone to one, the change would have started
failing linter-free scaffolds, which is a regression rather than a success.

ox3 and ox4 carry a second duty: `oxlint.config.ts` / `oxlint.config.mts` differ from
eslint's `eslint.config.ts` / `eslint.config.mts` only in prefix, and the eslint row stays
skipped on both — the two entries do not claim each other's names.

## Evidence — Template-First, both axes

Source axis:

```
$ sed -n '20p' internal/template/templates/.claude/rules/moai/languages/javascript.md
- Linting: ESLint 9 flat config, Biome, oxlint
$ sed -n '20p' .claude/rules/moai/languages/javascript.md
- Linting: ESLint 9 flat config, Biome, oxlint
$ diff -q <the two paths>
DIFF_EXIT=0
```

Embed axis — the one a source-to-source comparison cannot reach:

```
$ ls -la bin/moai
-rwxr-xr-x@ 1 goos staff 70212386 Sep 10 14:30 bin/moai
$ strings bin/moai | grep -m2 "Linting: ESLint"
- Linting: ESLint 9 flat config, Biome, oxlint
```

`make build` actually ran and the rebuilt binary embeds the new sentence. Checking only
the two source copies would have passed even if the build had been skipped.

## Baseline-attribution

The BEFORE column is `.moai/reports/t550/baseline.md`, measured on this tree at
d060e0d13 before any edit. The AFTER column was measured in this run, on this tree, at
the branch tip carrying the three commits above, with the same fixtures, the same stubs,
and the same summary-reading instrument. No figure is carried over from another tree,
package, or point in time.

## Gaps — not observed

- Real `npx oxlint` / `npx biome` / `npx eslint` execution. The stubs exit 0
  unconditionally, so this measures step SELECTION, not linter verdicts. The AC-002 test
  covers the red-run case separately with a failing fake npx.
- A zero-config oxlint project. Out of scope by the lead's ruling and recorded as a
  surviving limitation in spec.md §5; oxlint runs without a config file
  (https://oxc.rs/docs/guide/usage/linter/config.html, accessed 2026-09-10).
- Co-resident linter configs (a project carrying two linters' config files). Explicitly
  out of scope; pre-existing behaviour preserved.
- Non-Node toolchains, other platforms than darwin, and the full test suite — CI's job.
- The implementing agent's mutant-probe and AC-matrix runs were reported to me, not
  re-executed by me; the seven-fixture re-measurement, the diff, and the embed check above
  are my own observations.

## Residual-risk

`npx oxlint` with no path argument relies on oxlint defaulting to the current directory.
The stub cannot confirm that default; it is taken from upstream documentation. If that
default ever changes, the step would run against the wrong scope while still reporting
executed.

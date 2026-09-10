# t550 — baseline measurement (GH #1631)

card: t550
tree: .claude/worktrees/t550
branch: WT-js-linter-detect
HEAD: d060e0d136f173f07353e46a709382d08ae25dd2 (= local develop at dispatch time)
measured: 2026-09-10

## Claim

On the tree above, `moai gate`'s Node lint axis detects **eslint and biome**, and detects
**no oxlint at all**. An oxlint project therefore runs zero lint steps while the gate exits 0.

The card text ("biome·oxlint 미지원") is **half stale**: biome landed on
`09561fd93 fix(hook): stop silent config-gated lint skip in moai gate (card t233)`,
whose source comment names issue #1631 directly. The residual gap is oxlint only.

## Evidence

Binary built from this tree: `go build -o <scratchpad>/moai-base ./cmd/moai` (exit 0).

Four fixtures, each `package.json` + `index.js`, differing only in linter config file.
`npx`/`npm`/`node` shadowed by exit-0 stubs on PATH so no network fetch occurs; the gate's
own run summary is the instrument.

Command per fixture:
`CLAUDE_PROJECT_DIR=<fx> PATH=<stub>:/usr/bin:/bin:/usr/sbin:/sbin <scratchpad>/moai-base gate`

| fixture | linter config present | eslint row | biome row | lint steps executed |
|---|---|---|---|---|
| eslintprj | `eslint.config.js` | **executed** — `npx eslint .` | skipped (config files absent) | 1 |
| biomeprj  | `biome.json`        | skipped (config files absent) | **executed** — `npx biome check .` | 1 |
| oxlintprj | `.oxlintrc.json`    | skipped (config files absent) | skipped (config files absent) | **0** |
| ox2       | `.oxlintrc.jsonc`   | skipped (config files absent) | skipped (config files absent) | **0** |
| ox3       | `oxlint.config.ts`  | skipped (config files absent) | skipped (config files absent) | **0** |
| ox4       | `oxlint.config.mts` | skipped (config files absent) | skipped (config files absent) | **0** |
| bareprj   | none (control)      | skipped (config files absent) | skipped (config files absent) | 0 |

All four of oxlint's own config filenames were measured, not just the first. A four-name
support claim resting on one measured name would leave three asserted-but-unmeasured;
the filename list is a cheap control, so it was spent. `oxlint.config.ts` and
`oxlint.config.mts` additionally confirm that the eslint entry does not accidentally
claim them — its list carries `eslint.config.ts` / `eslint.config.mts`, which differ only
in prefix.

Verbatim summary, oxlintprj:

```
quality gate steps (4 configured):
  - typecheck: skipped — no scripts.typecheck and no tsconfig.json; set gate.typecheck.command to enable one
  - eslint: skipped — none of its config files exist in the project directory (eslint.config.js, ...)
  - biome: skipped — none of its config files exist in the project directory (biome.json, biome.jsonc)
  - npm test: executed in 11ms at 2026-09-10T13:44:08.456+09:00 — npm test -- --passWithNoTests
```

Verbatim summary, biomeprj — the control that proves the instrument separates
"both pass" from "both skip":

```
  - eslint: skipped — none of its config files exist in the project directory (...)
  - biome: executed in 19ms at 2026-09-10T13:44:02.738+09:00 — npx biome check .
```

Source-tree absence check (`/usr/bin/grep`, not the shell's ugrep wrapper):

```
$ /usr/bin/grep -rn "oxlint" --include="*.go" internal/ pkg/ cmd/
(no output)
```

`biome` by contrast is present at `internal/hook/quality/gate.go:224`.

## Baseline-attribution

Every row above was produced by the command named in this section, run in this run,
against the tree at HEAD d060e0d13. No figure is carried over from another tree or run.

## Gaps

- Not observed: behaviour of a **zero-config** oxlint project (oxlint runs with no config
  file; https://oxc.rs/docs/guide/usage/linter/config.html — "Oxlint works out of the box").
  The fixtures pin the config-file axis only. Per the lead's ruling this stays out of
  scope: the population of zero-config oxlint users has never been measured, and widening
  detection for an unmeasured population would be speculation. Recorded, not silent.
- Not observed: real `npx oxlint` / `npx biome` execution — the stubs exit 0 unconditionally,
  so this measures step **selection**, not linter verdicts.
- Not observed: non-Node toolchains (Go/Python/Rust/...) — out of card scope.

## Residual-risk

The `bareprj` control shows the gate legitimately runs no lint step when a Node project
carries no linter config. Adding oxlint detection must not turn that control red — a
zero-config-oxlint discriminator resting on `package.json` alone would risk exactly that.

## Upstream fact (verified)

oxlint's own config discovery, per https://oxc.rs/docs/guide/usage/linter/config.html:
`.oxlintrc.json`, `.oxlintrc.jsonc`, `oxlint.config.ts`, `oxlint.config.mts`.

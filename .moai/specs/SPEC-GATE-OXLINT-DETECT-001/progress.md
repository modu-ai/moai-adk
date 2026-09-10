# Progress — SPEC-GATE-OXLINT-DETECT-001

Card: t550 · Issue: #1631 · Baseline tree: `d060e0d13`

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase artifacts authored 2026-09-10: `spec.md`, `plan.md`, `acceptance.md`, this file.
- Tier M (single-file behaviour change plus tests); no `design.md` / `research.md`.
- Baseline read from `.moai/reports/t550/baseline.md` (tree `d060e0d13`); no figure in the
  artifacts is derived elsewhere.
- SPEC ID regex check executed: `[[ "SPEC-GATE-OXLINT-DETECT-001" =~ ^SPEC(-[A-Z][A-Z0-9]*)+-[0-9]{3}$ ]]` → `PASS`.
- Gating axis **resolved** by lead ruling 2026-09-10: config-file gating only (`spec.md §3`).
  The zero-config oxlint gap is a knowing, cited out-of-scope clause (`spec.md §5`), and the
  card-premise correction is durable (`spec.md §3.1`).

## §E.2 Run-phase Evidence

Measured tree: **`d608847c2`** (tree object `797f289fb`), branch `WT-js-linter-detect`,
worktree `.claude/worktrees/t550`. `git status --porcelain --untracked-files=no` printed
nothing at measurement time, so every figure below was measured against exactly that tree.
Baseline for comparison: `.moai/reports/t550/baseline.md`, tree `d060e0d13`.

Verbatim outputs are persisted under `.moai/reports/t550/`; each row names the file that
holds the full run rather than quoting it twice.

### Pre-edit RED (observed, not inherited)

The seven tests were written and run **before** the `gate.go` edit. Evidence:
`.moai/reports/t550/red-pre-edit.txt` (tree `d060e0d13` + the new test file only).

```
$ go test ./internal/hook/quality/ -run 'TestNodeLintRunsOxlintStepOnOxlintProject|…' -v
--- FAIL: TestNodeLintRunsOxlintStepOnOxlintProject (0.13s)
    gate_oxlint_lint_test.go:81: no oxlint row in the run summary — the Node toolchain
        has no oxlint lint step; summary:
        quality gate steps (3 configured):
          - eslint: skipped — none of its config files exist in the project directory (…)
          - biome: skipped — none of its config files exist in the project directory (…)
--- FAIL: TestNodeLintOxlintViolationFailsGate (0.00s)
--- FAIL: TestNodeLintOxlintConfigFilenames (0.02s)
    --- FAIL: …/oxlint_config/.oxlintrc.json      --- FAIL: …/oxlint_config/.oxlintrc.jsonc
    --- FAIL: …/oxlint_config/oxlint.config.ts    --- FAIL: …/oxlint_config/oxlint.config.mts
--- FAIL: TestNodeLintOxlintStepShape (0.00s)
    gate_oxlint_lint_test.go:193: Node toolchain lintSteps has no oxlint entry: [{name:eslint …} {name:biome …}]
--- FAIL: TestNodeLintEslintProjectUnaffectedByOxlint (0.19s)
--- FAIL: TestNodeLintBiomeProjectUnaffectedByOxlint (0.17s)
--- FAIL: TestNodeLintLinterFreeScaffoldStillPasses (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/hook/quality	1.080s
```

RED for the stated reason: the Node `lintSteps` list had no oxlint entry, so no row was
recorded — not for anything the fixtures lacked.

### AC matrix

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-001 | **PASS** | `go test ./internal/hook/quality/ -run TestNodeLintRunsOxlintStepOnOxlintProject -v` | `--- PASS: TestNodeLintRunsOxlintStepOnOxlintProject (0.13s)` — the oxlint row reads `executed … — npx oxlint` |
| AC-002 | **PASS** | `go test … -run TestNodeLintOxlintViolationFailsGate -v` | `--- PASS: TestNodeLintOxlintViolationFailsGate (0.31s)` — gate fails, output contains `quality gate failed: oxlint` |
| AC-003 | **PASS** | `go test … -run TestNodeLintOxlintConfigFilenames -v` | **four** subtests reported by name: `--- PASS: …/oxlint_config/.oxlintrc.json`, `/.oxlintrc.jsonc`, `/oxlint.config.ts`, `/oxlint.config.mts` (`grep -c` over the persisted run = `4`). Swept set non-empty and complete |
| AC-004 | **PASS** | `go test … -run TestNodeLintOxlintStepShape -v` | `--- PASS: TestNodeLintOxlintStepShape (0.00s)` — binary `npx`, args `oxlint`, `optional` true, `configFiles` set-equal to the four names in both directions |
| AC-005 | **PASS** (control) | `go test … -run TestNodeLintEslintProjectUnaffectedByOxlint -v` | `--- PASS (0.17s)` — eslint `executed — npx eslint .`; biome and oxlint both `skipped` config-absent; executed lint steps = 1 |
| AC-006 | **PASS** (control) | `go test … -run TestNodeLintBiomeProjectUnaffectedByOxlint -v` | `--- PASS (0.17s)` — biome `executed — npx biome check .`; eslint and oxlint both `skipped` config-absent; executed lint steps = 1 |
| AC-007 | **PASS** (control) | `go test … -run TestNodeLintLinterFreeScaffoldStillPasses -v` | `--- PASS (0.01s)` — gate passes, executed lint steps = 0, three visible config-absent notices (eslint, biome, oxlint) |
| AC-008 | **PASS** | `go test ./internal/hook/quality/...` | `ok  	github.com/modu-ai/moai-adk/internal/hook/quality	18.271s`; the `-run` selection above reported 7 tests + 4 subtests by name, so the sweep is not empty. All three §D.9 mutants observed red (below) |
| AC-009 | **PASS** | `grep -n "Linting:" <template> <local>` and `diff <template> <local>` | both copies line 20 read `- Linting: ESLint 9 flat config, Biome, oxlint`; `diff` printed nothing, `diff-exit=0`. `git diff --stat` on each copy: `1 file changed, 1 insertion(+), 1 deletion(-)` — the 16-programming-language neutrality guard held |

Full run outputs: `.moai/reports/t550/red-pre-edit.txt` (RED),
`.moai/reports/t550/green-post-edit.txt` (GREEN).

### §D.9 mutant probes — all three run, all three observed red

| Mutant | `configFiles` under probe | Observed | Evidence |
|---|---|---|---|
| **M1** (ungated) | `[]` | AC-001 / AC-003 stayed green (4/4 subtests PASS) and **AC-005, AC-006, AC-007 all FAILED** — the ungated entry executed oxlint on the eslint, biome and linter-free fixtures. AC-004 failed too (len 0 ≠ 4). The controls are load-bearing, not decorative | `.moai/reports/t550/mutant-m1.txt` |
| **M2** (shrunk) | `[".oxlintrc.json"]` | AC-001 green; **AC-003 FAILED on the three remaining filenames** (`.oxlintrc.jsonc`, `oxlint.config.ts`, `oxlint.config.mts`) and **AC-004 FAILED on the missing side**: `missing config file name: .oxlintrc.jsonc (declared: [.oxlintrc.json])` ×3. Controls stayed green | `.moai/reports/t550/mutant-m2.txt` |
| **M3** (grown) | the four correct names **plus** `oxlint.config.js` | **Every behavioural AC stayed green** (AC-001, AC-002, AC-003 4/4, AC-005, AC-006, AC-007) and **AC-004 alone FAILED** on the growth side: `oxlint configFiles has 5 entries, want exactly 4` and `unexpected config file name: oxlint.config.js`. The §D.4 equality fix is therefore **demonstrated**, not asserted — M3 did not pass, so the check was written as equality | `.moai/reports/t550/mutant-m3.txt` |

Each mutant was applied to `gate.go`, measured, and reverted; the tree above carries the
correct four names.

### Post-change re-measurement — same instrument as the baseline

Recipe persisted at `.moai/reports/t550/remeasure.sh`; full output at
`.moai/reports/t550/remeasure-after.txt`. Identical to the baseline procedure: a binary
built from this tree (`go build -o $SP/moai-after ./cmd/moai`, exit 0), seven fixtures
differing only in their linter config file, `npx`/`npm`/`node` shadowed by exit-0 stubs on
`PATH` (nothing reaches the network), and the gate's own run summary read as the instrument.

| fixture | linter config | eslint | biome | oxlint | lint steps — **before** (`d060e0d13`) → **after** (`d608847c2`) | gate exit |
|---|---|---|---|---|---|---|
| eslintprj | `eslint.config.js` | executed | skipped | *(absent)* → skipped | 1 → **1** | 0 |
| biomeprj | `biome.json` | skipped | executed | *(absent)* → skipped | 1 → **1** | 0 |
| oxlintprj | `.oxlintrc.json` | skipped | skipped | *(absent)* → **executed — `npx oxlint`** | 0 → **1** | 0 |
| ox2 | `.oxlintrc.jsonc` | skipped | skipped | *(absent)* → **executed — `npx oxlint`** | 0 → **1** | 0 |
| ox3 | `oxlint.config.ts` | skipped | skipped | *(absent)* → **executed — `npx oxlint`** | 0 → **1** | 0 |
| ox4 | `oxlint.config.mts` | skipped | skipped | *(absent)* → **executed — `npx oxlint`** | 0 → **1** | 0 |
| bareprj | none (**CONTROL**) | skipped | skipped | *(absent)* → skipped | 0 → **0** | 0 |

The no-linter control is unchanged: zero lint steps, gate exit 0, and now three visible
config-absent notices instead of two. Adding oxlint did not turn a linter-free scaffold red.

The `ox3` / `ox4` rows carry the baseline's second finding forward: the eslint entry still
skips on `oxlint.config.ts` / `oxlint.config.mts` despite its own list carrying
`eslint.config.ts` / `eslint.config.mts`, one prefix away.

### Quality gates

| Check | Command | Output |
|---|---|---|
| Package suite | `go test ./internal/hook/quality/...` | `ok  	github.com/modu-ai/moai-adk/internal/hook/quality	18.271s` |
| Vet | `go vet ./internal/hook/quality/...` | no output, `vet-exit=0` |
| Lint | `golangci-lint run ./internal/hook/quality/...` | `0 issues.` |

Scope note: verification is package-scoped by dispatch ([HARD], loaded machine, six lanes).
`go test ./...` was **not** run locally and `internal/cli` was **not** touched or run — the
full-suite verdict belongs to CI.

### Decisions held unchanged

- Gating axis is config-files-only (`spec.md §3`). No `package.json` `devDependencies`
  discriminator was added in flight.
- The zero-config oxlint out-of-scope clause (`spec.md §5`) stands unedited.
- Template-First was applied per file, per `plan.md §E`: **not** to `gate.go` (no mirror, no
  `make build` obligation from it), **yes** to `javascript.md` (template source first →
  `make build` → local sync → `diff` exit 0).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-10
run_commit_sha: d608847c215007dfab0579f043359b12a9fb76c4
run_tree_sha: 797f289fb4a6656cb021e2e04349b8f24c94595f
run_status: complete
ac_pass_count: 9
ac_fail_count: 0
mutant_probes_run: 3
mutant_probes_observed_red: 3
preserve_list_post_run_count: 0
new_warnings_or_lints_introduced: 0
cross_platform_build:
  measured: false
  note: >-
    Not observed. Verification was package-scoped by dispatch; the
    GOOS=windows build is CI's. The execution-bearing tests skip on
    runtime.GOOS == "windows" (shell-script fake binaries), mirroring the
    sibling biome tests; AC-004 carries no such guard and runs everywhere.
total_run_phase_files: 4
m1_to_mN_commit_strategy: >-
  M1 (7d28ddddf) gate.go + gate_oxlint_lint_test.go; M1b (d608847c2)
  both javascript.md copies after make build; M2 this progress record.
  No push, no merge, no branch other than WT-js-linter-detect.
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

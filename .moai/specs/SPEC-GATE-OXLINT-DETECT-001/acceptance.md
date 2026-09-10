# Acceptance — SPEC-GATE-OXLINT-DETECT-001

> Harness: **standard**. Baseline tree: `d060e0d13` (card t550 worktree), measured 2026-09-10,
> evidence `.moai/reports/t550/baseline.md`. Every RED-now cell below is pinned to that SHA.
> All Go verification runs are package-scoped: `go test ./internal/hook/quality/...`.
> Full-suite verdicts belong to CI, never to a local run (CLAUDE.local.md §4).

## §D AC Matrix

| AC | Owner | REQ | Verification | Expected (baseline → after) |
|---|---|---|---|---|
| AC-001 | M1 | REQ-001·002 | oxlint fixture produces an `executed` oxlint row, command `npx oxlint` | no oxlint row → executed |
| AC-002 | M1 | REQ-003 | red oxlint run (fake npx exit 1) fails the gate, output names the step | gate passes → gate fails |
| AC-003 | M1 | REQ-002 | each of the four config filenames selects the step, **one fixture per name** | n/a → 4/4 measured |
| AC-004 | M1 | REQ-002 | step shape pinned against the `toolchains` table: binary, args, and `configFiles` **set-equal** to the four names (both directions) | absent entry → pinned |
| AC-005 | **control** | REQ-005 | eslint fixture: eslint executes, biome **and** oxlint both skip config-absent | 1 lint step → 1 lint step |
| AC-006 | **control** | REQ-005 | biome fixture: biome executes, eslint **and** oxlint both skip config-absent | 1 lint step → 1 lint step |
| AC-007 | **control** | REQ-004·006 | linter-free Node project: zero lint steps, three visible config-absent notices | 0 steps, 2 notices → 0 steps, 3 notices |
| AC-008 | M1 | REQ-007 | package test suite green, and all three mutant probes of §D.9 observed red | — |
| AC-009 | M1b | REQ-001 | javascript.md line 20 names oxlint in **both** copies, and the two copies are byte-identical (`diff` exit 0) | Biome only, 2 copies identical → +oxlint, 2 copies identical |

Controls (AC-005..007) are not optional rows: without them the change is
indistinguishable from "every lint step now runs on every Node project".

## §D.1 AC-001 — oxlint project runs an oxlint step (RED-now)

- **Given** a fixture directory containing `package.json` and `.oxlintrc.json` and **no**
  eslint or biome config, with `npx` shadowed by a fake binary exiting 0 — and, on tree
  `d060e0d13`, `/usr/bin/grep -rn "oxlint" --include="*.go" internal/ pkg/ cmd/` printed
  nothing (exit 1), so no oxlint step can exist to select.
- **When** the gate runs with lint as its only live axis (vet absent on Node, typecheck and
  ast-grep disabled, tests skipped), i.e. the `oxlintGate` helper mirroring `biomeGate`.
- **Then** `g.summary.recordFor("oxlint")` is non-nil, its `outcome` is `outcomeExecuted`,
  and its `command` is exactly `npx oxlint`. The gate passes.
- **RED reason:** red because the Node `lintSteps` list has no oxlint entry, so no row is
  recorded — not because of anything the fixture lacks.
- **Green path:** flipped by M1's `toolchains` edit adding the oxlint entry. Passing output
  becomes an `oxlint` row reading `executed … — npx oxlint`.

## §D.2 AC-002 — a red oxlint run fails the gate (RED-now)

- **Given** the same oxlint fixture, with the fake `npx` exiting **1** and printing a
  violation line on stderr.
- **When** the gate runs.
- **Then** `passed` is false and the output contains `quality gate failed: oxlint`.
- **RED reason:** red because the step is never selected, so the non-zero exit is never
  reached — the gate returns a pass on a red linter, which is issue #1631's reported symptom.
- **Green path:** M1's `toolchains` edit; the existing failure-propagation path then names
  the step, exactly as `TestNodeLintBiomeViolationFailsGate` pins for biome.

## §D.3 AC-003 — each config filename is exercised individually

- **Given** four fixtures, each carrying `package.json` plus exactly **one** of
  `.oxlintrc.json`, `.oxlintrc.jsonc`, `oxlint.config.ts`, `oxlint.config.mts`, and no other
  linter config; `npx` shadowed, exit 0.
- **When** the gate runs against each fixture in turn (a table-driven subtest, one case per
  filename).
- **Then** every one of the four cases records an `oxlint` row with outcome
  `outcomeExecuted`. A subtest that fails names the filename it was given.
- **Why individually — [HARD], not a preference:** REQ-002 claims four names are detected. A
  single fixture measures one of them and leaves the other three **asserted but unmeasured**,
  so a typo in any of the three passes unseen behind the fourth. One fixture per name, four
  fixtures, four observations. A filename list is a cheap control; it is spent here in full.
- **Swept-set check:** the subtest run must report four `oxlint_config/<filename>` cases by
  name. Three reported cases is a partial sweep, not a pass
  (`verification-completeness.md` §1.1).
- **RED reason:** red for the AC-001 reason (no oxlint entry exists). Measured per filename,
  not inferred: the baseline's `oxlintprj` / `ox2` / `ox3` / `ox4` fixtures (tree
  `d060e0d13`) each executed **0** lint steps.
- **Green path:** the same M1 edit, which must list all four names; all four subtests then
  report an executed oxlint row.
- **Second finding carried by `ox3` / `ox4`:** the eslint entry must still skip on
  `oxlint.config.ts` and `oxlint.config.mts`. Its own list carries `eslint.config.ts` /
  `eslint.config.mts`, which differ only in prefix, so these two cases double as a
  mis-claim control and each asserts an `eslint` row of `outcomeSkipped`.

## §D.4 AC-004 — step shape pinned against the toolchains table

- **Given** the package-level `toolchains` var, read through the existing `nodeToolchain(t)`
  helper.
- **When** the Node entry's `lintSteps` is searched for an element named `oxlint`.
- **Then** it exists; its `binary` is `npx`; `strings.Join(args, " ")` is `oxlint`;
  `optional` is true; and its `configFiles` is **set-EQUAL** to exactly the four names of
  §D.3 — `len(configFiles) == 4` **and** equal membership in **both** directions. The
  assertion reports each direction separately and distinguishably: every name of §D.3 absent
  from `configFiles` is reported as `missing config file name: <name>`, and every entry of
  `configFiles` outside §D.3 is reported as `unexpected config file name: <name>`. A missing
  name and an unexpected name therefore produce different failures, never one merged count.
- **Why equality and not containment:** REQ-002 states a closed norm — "No other filename is
  recognized" (`spec.md §2`). Containment constrains only the shrinking direction, so a
  fifth name added to the list (`oxlint.config.js`, which oxlint does not read) would leave
  AC-001 through AC-008 green while that norm is violated. Equality is what makes the
  growth direction observable; §D.9 mutant M3 is its demonstration.
- **Boundary:** this AC asserts the declaration; §D.1–§D.3 assert the behaviour. Neither
  substitutes for the other — a correct declaration that the selector never reads would pass
  this row and fail those, and a superset declaration passes every behavioural row while
  failing only this one.

## §D.5 AC-005 — CONTROL: eslint project is unchanged

- **Given** a fixture with `package.json` and `eslint.config.js` only; `npx` shadowed, exit 0.
- **When** the gate runs.
- **Then** the `eslint` row is `outcomeExecuted` with command `npx eslint .`; the `biome` row
  is `outcomeSkipped`; the `oxlint` row is `outcomeSkipped`; and the skip reason on each
  names the absent config files. Exactly one lint step executed.
- **Baseline (tree `d060e0d13`):** eslint executed, biome skipped, 1 lint step
  (`baseline.md` fixture `eslintprj`). This row must read the same after the change, with the
  oxlint skip added.
- **Failure meaning:** if oxlint executes here, the entry is not config-gated and the change
  has become "run every linter everywhere".

## §D.6 AC-006 — CONTROL: biome project is unchanged

- **Given** a fixture with `package.json` and `biome.json` only; `npx` shadowed, exit 0.
- **When** the gate runs.
- **Then** the `biome` row is `outcomeExecuted` with command `npx biome check .`; the
  `eslint` and `oxlint` rows are both `outcomeSkipped` with config-absent reasons. Exactly
  one lint step executed.
- **Baseline (tree `d060e0d13`):** biome executed, eslint skipped, 1 lint step
  (`baseline.md` fixture `biomeprj`).

## §D.7 AC-007 — CONTROL: linter-free scaffold still passes with visible notices

- **Given** a fixture with `package.json` only — no linter config of any kind; `npx`
  shadowed and present (so the optional-binary guard passes) but never invoked.
- **When** the gate runs.
- **Then** the gate **passes**; zero lint steps executed; and the run summary carries a
  visible config-absent skip notice naming each of `eslint`, `biome`, and `oxlint` — three
  notices, up from the two on the baseline tree.
- **Baseline (tree `d060e0d13`):** fixture `bareprj` — both entries skipped, 0 lint steps,
  gate exit 0.
- **Failure meaning:** a fail here means the change started blocking first commits and
  scaffolds. This is the row that guards the baseline's stated residual risk.

## §D.8 AC-008 — suite green and swept set non-empty

- **Given** the M1 edit and the new test file `internal/hook/quality/gate_oxlint_lint_test.go`.
- **When** `go test ./internal/hook/quality/...` runs.
- **Then** it exits 0, and its output is **not** an empty sweep — the run must report the new
  oxlint tests by name under `-run` selection, so a zero-match selector cannot be read as a
  pass (`verification-completeness.md` §1.1). The count of oxlint subtests observed is
  recorded in `progress.md §E.2` alongside the verbatim output.
- **Windows:** the fake-binary fixtures are shell scripts, so the execution-bearing tests
  skip on `runtime.GOOS == "windows"`, mirroring the biome tests. AC-004 carries no such
  guard and must run on every platform.

## §D.9b AC-009 — javascript.md linter list, both copies (RED-now)

- **Given** on tree `d060e0d13`, line 20 of BOTH copies reads
  `- Linting: ESLint 9 flat config, Biome` — measured, and the two copies `diff` clean
  (exit 0). oxlint is absent from both; Biome is already present in both.
- **When** after M1b: `grep -n "Linting:" internal/template/templates/.claude/rules/moai/languages/javascript.md .claude/rules/moai/languages/javascript.md`
  and `diff internal/template/templates/.claude/rules/moai/languages/javascript.md .claude/rules/moai/languages/javascript.md`.
- **Then** both grep hits name `oxlint` on line 20, and the `diff` exits 0 — the two copies
  are byte-identical again. `make build` was run between the template edit and the sync.
- **RED reason:** red because oxlint is absent from the line, on both copies. Not red for a
  parity reason — parity holds today and must still hold after.
- **Both directions matter:** an edit landing in only the template copy fails the `diff`; an
  edit landing in only the local copy fails the `diff` too **and** ships nothing to
  downstream users. Only the pair passes.
- **Scope guard:** `git diff --stat` on the two files shows exactly one changed line each.
  A larger diff means the neutrality guard (`plan.md §C.1`) was breached.

## §D.9 Mutant probe

Before adopting AC-001, AC-003, and AC-004, the run phase writes a mutant that would satisfy a shallower
criterion while violating the requirement, and confirms the AC catches it:

- **Mutant M1:** oxlint entry present but with an empty `configFiles` list (ungated). It
  satisfies AC-001 and AC-003 and **must fail AC-005, AC-006, and AC-007** — this is what
  makes the controls load-bearing rather than decorative.
- **Mutant M2 (shrinking direction):** oxlint entry declaring only `.oxlintrc.json`. It
  satisfies AC-001 and **must fail AC-003** on the three remaining filenames and AC-004 on
  the `missing config file name:` side of the equality check.
- **Mutant M3 (growth direction):** oxlint entry declaring the four correct names **plus**
  `oxlint.config.js` — a name absent from oxlint's documented discovery list. It satisfies
  AC-001, AC-003, and **every behavioural AC** (AC-002, AC-005, AC-006, AC-007 all stay
  green: no fixture in this SPEC carries an `oxlint.config.js` file, so the extra name
  changes no observed behaviour). It **must fail AC-004** on the `unexpected config file
  name: oxlint.config.js` side. If it does not fail AC-004, the equality check was not
  actually written as equality and the criterion is still too shallow to adopt — M3 is the
  direct control for the growth direction, and without it the §D.4 fix is asserted rather
  than demonstrated.

All three probes are run and their observed output recorded; a mutant that passes means the
AC is too shallow to adopt.

## §D.10 Definition of Done

- All nine ACs pass, each with the command and its verbatim output recorded in
  `progress.md §E.2`, attributed to the tree SHA they were measured on (REQ-007).
- All three mutant probes of §D.9 (M1, M2, M3) observed failing the ACs they are supposed to
  fail. M3 is not optional: dropping it closes the card on an asserted §D.4 fix rather than a
  demonstrated one.
- `go vet ./internal/hook/quality/...` and `golangci-lint run` clean on the touched package.
- **Template-First, split by file** (`plan.md §E`):
  - `internal/hook/quality/gate.go` — no template mirror added, none required.
  - `.claude/rules/moai/languages/javascript.md` — the template source
    (`internal/template/templates/…`) was edited **first**, `make build` was run, and the
    local copy was synced afterwards. `diff` of the two copies exits 0 (AC-009). A green
    suite with the two copies diverged is NOT done.
- `git diff --stat` shows exactly one changed line in each javascript.md copy — the
  16-programming-language neutrality guard held (`plan.md §C.1`).
- The gating axis stays as decided (`spec.md §3`, config files only). No `package.json`
  `devDependencies` discriminator was added in flight, and the zero-config out-of-scope
  clause (`spec.md §5`) still stands unedited at close.

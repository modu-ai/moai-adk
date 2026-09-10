# Plan — SPEC-GATE-OXLINT-DETECT-001

> Ordered by decision-reversibility: the decisions most likely to change lead, the
> mechanical steps close. Priority labels only — no time estimates.

## §A Context

`moai gate`'s Node toolchain runs a lint step only when one of that step's declared config
files exists. Two entries are declared (`eslint`, `biome`); oxlint is absent from the Go
sources entirely, so an oxlint project's lint axis runs nothing while the gate exits 0.
Measured on tree `d060e0d13` — `.moai/reports/t550/baseline.md`, the sole factual source for
this plan.

## §B Decisions that could still change (review these first)

### B.1 Gating axis — RESOLVED, config files only (Priority High)

Decided by the lead 2026-09-10 and recorded in `spec.md §3`: config-file gating, one
discriminator per linter entry. The `package.json` `devDependencies` axis is rejected for
now — the zero-config oxlint population has never been measured, and a `package.json`-based
discriminator is the shape most likely to turn the linter-free-scaffold control (AC-007) red.

Consequence run-phase must not quietly undo: a zero-config oxlint project still runs no lint
step after this fix. That is an explicit out-of-scope clause (`spec.md §5`), not an oversight
to fix in flight. Re-opening it is a lead decision, and the delta is one REQ-002 trigger
clause plus one acceptance row.

### B.2 The recognized filename set (Priority High)

Exactly four, taken from oxlint's own documented discovery order
(<https://oxc.rs/docs/guide/usage/linter/config.html>): `.oxlintrc.json`, `.oxlintrc.jsonc`,
`oxlint.config.ts`, `oxlint.config.mts`. Adding a name that oxlint does not read would make
the gate run oxlint on a project oxlint itself would ignore; omitting one leaves a real
oxlint project uncovered.

**The set is closed, and both directions are guarded.** REQ-002 says "No other filename is
recognized", so the check is set **equality**, never containment: AC-003 exercises each name
individually so no omission hides, and AC-004 asserts `configFiles` set-equal to exactly
these four — reporting missing names and unexpected names as distinguishable failures.
Containment would have constrained only the shrinking direction, leaving a fifth name free
to be added with the whole suite green. Mutant M3 (`acceptance.md §D.9`) demonstrates the
growth direction rather than asserting it.

### B.3 Command form — `npx oxlint` (Priority Medium)

Bare `npx oxlint` with no path argument, unlike `npx eslint .` and `npx biome check .`.
oxlint defaults to the current directory, so the bare form is the idiomatic invocation; the
run summary records the command string verbatim, and AC-001 asserts it exactly, so a change
of mind here is a one-line edit plus one assertion.

## §C Technical approach

One additive element in the Node entry of the `toolchains` var
(`internal/hook/quality/gate.go`), mirroring the biome entry that landed on `09561fd93` for
this same issue:

- `name: "oxlint"`, `binary: "npx"`, `args: []string{"oxlint"}`, `optional: true`,
  `configFiles:` the four names of §B.2.
- A short source comment naming issue #1631 and this SPEC, in the same register as the
  adjacent biome comment.

Nothing else in the file changes. The eslint and biome entries are not touched, and the
selection, skip-notice, and failure-propagation machinery is consumed as-is.

### §C.1 Documenting our own change — javascript.md line 20

The distributed language rule states the Node linter set and has not kept up:

```
- Linting: ESLint 9 flat config, Biome
```

[HARD] **Exactly ONE name is missing — oxlint.** Biome IS already listed; a reading that
both are absent is wrong, and adding Biome a second time would be the error that reading
produces. The edit adds `oxlint` to that one line and changes nothing else.

[HARD] **Programming-language neutrality (CLAUDE.local.md §15).** This file ships to every
downstream user, and the rule tree covers 16 programming languages equally. Change the JS
linter list to match reality and **nothing else**: no restructuring, no added examples, no
edit to any other language's section, and no edit to any other line of this file.

## §D Milestones

### M1 — oxlint entry + tests (Priority High)

1. Write `internal/hook/quality/gate_oxlint_lint_test.go`, mirroring
   `gate_biome_lint_test.go`: an `oxlintGate` helper (lint as the only live axis), reuse of
   `prependFakeNPX`, `writeFixture`, and `summary.recordFor`. Observe the AC-001/AC-002/AC-003
   tests **red** on the pre-edit tree and record the output — a criterion never seen red has
   proven nothing (`verification-completeness.md` §1.1).
2. Add the oxlint element to the Node `lintSteps` (§C).
3. Re-run; observe green. Record command + verbatim output + tree SHA.
4. Extend the control coverage of AC-005/AC-006/AC-007: the eslint and biome fixtures now
   assert an `oxlint` skip row as well, and the linter-free fixture asserts three visible
   config-absent notices instead of two.
5. Write AC-004 as a **set-equality** assertion over `configFiles` (both directions,
   distinguishable messages — `acceptance.md §D.4`), not a per-name membership check.
6. Run all three mutant probes of `acceptance.md §D.9` and record what each one failed. M3
   (four correct names plus `oxlint.config.js`) is the growth-direction control: it stays
   green on every behavioural AC and must fail AC-004 alone. An M3 that passes means the
   equality check was not written as equality.

### M1b — javascript.md linter list (Priority High)

Template-First order, non-negotiable (§E):

1. Edit `internal/template/templates/.claude/rules/moai/languages/javascript.md` line 20 —
   add `oxlint` to the linter list. Template source **first**.
2. Run `make build` to recompile the embedded template into the binary.
3. Sync the local copy `.claude/rules/moai/languages/javascript.md` to match.
4. Verify parity: `diff` of the two copies exits 0. A one-copy edit is drift, not a change.

### M2 — verification and close (Priority Medium)

- `go test ./internal/hook/quality/...`, `go vet ./internal/hook/quality/...`,
  `golangci-lint run` — package-scoped; the full-suite verdict is CI's, not a local run's
  (CLAUDE.local.md §4, §6).
- Record every AC's command and verbatim output in `progress.md §E.2` with the tree SHA.
- Confirm before close that the gating axis is still config-files-only (`spec.md §3`) and
  that the zero-config out-of-scope clause (`spec.md §5`) stands unedited.

## §E Template-First: applies to ONE of the two files

[HARD] This card touches two files and the Template-First rule (CLAUDE.local.md §2) answers
differently for each. Run phase MUST NOT apply one answer to both.

**`internal/hook/quality/gate.go` — Template-First does NOT apply.** It is Go source under
`internal/`, not a file under `internal/template/templates/`. There is no template mirror to
add and no `make build` obligation follows from editing it. Stated explicitly so run-phase
does not invent a mirror step.

**`.claude/rules/moai/languages/javascript.md` — Template-First DOES apply.** This file has
**two copies**, measured byte-identical on this tree (`diff` exit 0):
`internal/template/templates/.claude/rules/moai/languages/javascript.md` and
`.claude/rules/moai/languages/javascript.md`, each carrying the linter sentence at line 20.
The order is fixed: **edit the template source first → `make build` → then sync the local
copy.** Skipping `make build` leaves the template source edited while the embedded binary
still ships the old sentence, and landing the edit in only one copy is the drift this repo
keeps re-learning.

## §F Risks

| Risk | Mitigation |
|---|---|
| The entry ships ungated (empty `configFiles`) and every Node project starts invoking oxlint | AC-005/006/007 controls; mutant M1 of §D.9 is exactly this shape |
| A typo in one of four filenames hides behind the others | AC-003 exercises each name in its own subtest; mutant M2 |
| A test reaches real `npx` and fetches oxlint over the network | `prependFakeNPX` shadows `npx` in every execution-bearing test, as the biome tests do |
| The zero-config oxlint project stays uncovered | Knowingly accepted: lead decision in `spec.md §3`, explicit out-of-scope clause in `spec.md §5` with its upstream ground cited |
| Fake-binary fixtures are not executable on Windows | Execution-bearing tests skip on `runtime.GOOS == "windows"`, mirroring biome; AC-004 (table shape) runs everywhere |

## §G Anti-patterns to avoid

- Reading the card text "biome·oxlint 미지원" as current — biome landed on `09561fd93`
  (`spec.md §1.1`).
- Adding a template mirror or a `make build` step (§E).
- Landing the entry without the controls: green tests that only assert oxlint runs cannot
  tell the fix apart from "everything now runs".
- Adding a `package.json` `devDependencies` discriminator in-flight: the lead decided
  against it (`spec.md §3`), and the zero-config gap is a recorded out-of-scope clause.

## §H Cross-references

- `.moai/reports/t550/baseline.md` — the measured baseline (tree `d060e0d13`)
- `internal/hook/quality/gate.go` — the Node entry of the `toolchains` var
- `internal/hook/quality/gate_biome_lint_test.go` — the sibling test file to mirror (card t233)
- GitHub issue #1631; commit `09561fd93` (biome half of the same issue)
- `.claude/rules/moai/development/verification-completeness.md` §1.1, §2 — observed-failure
  completion and the two-cell RED-now discipline this plan's M1 follows

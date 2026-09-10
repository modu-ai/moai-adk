---
id: SPEC-GATE-OXLINT-DETECT-001
title: "moai gate Node lint axis — oxlint detection"
version: "0.1.0"
status: completed
created: 2026-09-10
updated: 2026-09-10
author: manager-spec (card t550)
priority: P1
phase: "v3.2.0"
module: "internal/hook/quality"
lifecycle: spec-anchored
tags: "quality-gate, node, lint, oxlint, issue-1631"
tier: M
issue_number: 1631
---

# SPEC — moai gate Node lint axis: oxlint detection

## HISTORY

- 2026-09-10: plan-phase authored on card t550. Baseline measured on tree `d060e0d13`
  (`.moai/reports/t550/baseline.md`); every factual claim below traces to that file.
- 2026-09-10: lead ruling applied. (a) The gating decision is **resolved** — Option A,
  config-file gating only; the former "Open decision" section became §3 Decision. (b) The
  zero-config oxlint gap is now an explicit out-of-scope clause with its upstream ground
  cited (§5). (c) The card-premise correction was promoted from an aside to a durable
  statement (§3.1), the lead having independently re-measured that #1631's biome axis was
  already closed by card t233 on `09561fd93`.
- 2026-09-10: plan-audit findings applied. **F1** — AC-004 was a containment check, which
  constrained only the shrinking direction and left REQ-002's closed norm ("No other filename
  is recognized") unguarded against an added fifth name; it is now a set-equality assertion,
  demonstrated by mutant M3 (`acceptance.md §D.4`, §D.9). **F4** — the Definition of Done
  still required "both" probes after M3 made three. **F5** — the reference-set guard below
  was added: the citation's access date plus the amendment rule for an upstream change.
- 2026-09-10: kickoff approved (operator, via the lead); three scope amendments. **F2** — the
  co-resident-config case is explicitly out of scope with its pre-existing behaviour preserved,
  and no follow-up card is raised (§5). **F6** — the §D.9 probe-scope clause now covers AC-004,
  the criterion M3 exists to validate. **F3** — documenting this change in
  `.claude/rules/moai/languages/javascript.md` is IN scope, which splits the Template-First
  claim: it does not apply to `gate.go` but DOES apply to javascript.md (two byte-identical
  copies, `make build` required) — §4, `plan.md §E`, AC-009.

## §1 Background and problem statement

`moai gate` selects its quality steps from a per-language toolchain table. The Node.js
toolchain (marker file `package.json`) carries a `lintSteps` list, and each entry is
config-gated: the step runs only when one of its declared config filenames exists in the
project directory. Absent every config file, the step is skipped with a visible
config-absent notice and the gate still exits 0.

That list currently holds exactly **two** entries — `eslint` and `biome` — both
`optional: true`, both `binary: "npx"` (`internal/hook/quality/gate.go`, the Node entry of
the `toolchains` var). A project whose linter is **oxlint** matches neither, so its lint
axis runs zero steps while the gate reports a pass.

Measured control matrix (tree `d060e0d13`, seven fixtures differing only in linter config,
`npx`/`npm`/`node` shadowed by exit-0 stubs — `.moai/reports/t550/baseline.md`):

| fixture | linter config | eslint | biome | lint steps executed |
|---|---|---|---|---|
| eslintprj | `eslint.config.js` | executed | skipped | 1 |
| biomeprj | `biome.json` | skipped | executed | 1 |
| oxlintprj | `.oxlintrc.json` | skipped | skipped | **0** |
| ox2 | `.oxlintrc.jsonc` | skipped | skipped | **0** |
| ox3 | `oxlint.config.ts` | skipped | skipped | **0** |
| ox4 | `oxlint.config.mts` | skipped | skipped | **0** |
| bareprj | none (**control**) | skipped | skipped | 0 |

**All four** of oxlint's config filenames were measured, not just the first — a four-name
claim resting on one measured name would leave three asserted-but-unmeasured. The `ox3` and
`ox4` rows carry a second finding: the eslint entry does **not** mis-claim
`oxlint.config.ts` / `oxlint.config.mts` despite its own list carrying
`eslint.config.ts` / `eslint.config.mts`, which differ only in prefix.

The `bareprj` row is the control: a Node project carrying no linter config legitimately
runs no lint step, and that must remain true after this change.

Source-tree absence confirmed on the same tree: `/usr/bin/grep -rn "oxlint" --include="*.go"
internal/ pkg/ cmd/` produced no output — oxlint has zero references in the Go sources.

### §1.1 The residual gap is oxlint alone

The `biomeprj` row above shows biome **executing**: issue #1631's biome axis is already
closed, and the residual gap is oxlint only. The card text that says otherwise is corrected
as a durable statement in **§3.1** — read that section, not this pointer, for the record.

## §2 GEARS requirements

- **REQ-001 (Ubiquitous):** The Node.js toolchain's lint axis shall recognize oxlint as a
  linter, so that a project whose only linter is oxlint has its lint axis actually run.

- **REQ-002 (When):** **When** a Node project directory contains any one of oxlint's own
  configuration files, the gate shall select an `oxlint` lint step and execute it as
  `npx oxlint`. The recognized filenames are exactly those oxlint itself discovers, per
  <https://oxc.rs/docs/guide/usage/linter/config.html> (accessed 2026-09-10):
  `.oxlintrc.json`, `.oxlintrc.jsonc`, `oxlint.config.ts`, `oxlint.config.mts`. No other
  filename is recognized.

  > **Reference-set guard (documentation, not a criterion).** Set equality measures whether
  > the implementation matches these four names; it cannot measure whether the four names are
  > still right, and that rests entirely on the citation above as read on its access date.
  > Equality sharpens the stake rather than softening it: should oxlint upstream add a fifth
  > discovery filename, AC-004 will actively **reject** it. If that happens, it is a REQ-002
  > amendment — update this list and the citation's access date — and never a test failure to
  > route around by loosening AC-004.

- **REQ-003 (When):** **When** the selected oxlint step exits non-zero, the gate shall fail
  and its failure output shall name the `oxlint` step, so the operator learns which axis
  went red.

- **REQ-004 (Where):** **Where** none of oxlint's configuration files is present, the
  oxlint step shall skip with the same visible config-absent notice every other config-gated
  lint entry emits, and shall not by itself fail the gate.

- **REQ-005 (Ubiquitous):** The oxlint entry shall be **additive**. The existing `eslint`
  and `biome` entries shall keep their present behaviour unchanged — their config gates,
  their commands, and their config-absent skip notices — so an eslint project never invokes
  oxlint and an oxlint project never invokes eslint or biome.

- **REQ-006 (Ubiquitous):** A Node project carrying no linter configuration at all shall
  continue to run zero lint steps and shall continue to emit a visible config-absent skip
  notice for each configured lint entry. The gate shall not begin failing linter-free
  scaffolds.

- **REQ-007 (Ubiquitous):** Every measurement recorded by this SPEC shall name the command
  that produced it, that command's observed output, and the tree it was measured on. No
  figure is carried over from another tree, package, or point in time.

## §3 Decision — config-file gating only

**Decided by the lead (2026-09-10): config-file gating, one discriminator per linter entry.**
REQ-002 is written against that decision, and no alternative remains open in this SPEC.

Three grounds, recorded so a later reader does not re-litigate the choice:

- **Consistency.** Every other linter entry in the `toolchains` table is gated on its own
  config files and nothing else. Adding a second discriminator to one entry would make
  oxlint the exception.
- **The speculative population was never measured.** The rejected alternative — also
  detecting `oxlint` in `package.json` `devDependencies` — would widen detection to cover
  zero-config oxlint projects, but the size of that population has never been measured on
  any tree. Widening on speculation was rejected; the risk it carries is concrete (a
  `package.json`-based discriminator is the shape most likely to turn the linter-free
  scaffold control red — baseline § Residual-risk).
- **It stays cheap later.** `nodeTypecheckStep` already reads `package.json` (it consults
  `scripts.typecheck`), so the `devDependencies` axis is not an alien capability. If it is
  ever wanted, the delta is one trigger clause in REQ-002 and one acceptance row.

## §3.1 Card-premise correction (durable statement)

Issue #1631 named two linters. **Its biome axis was already closed by card t233, on commit
`09561fd93`** (`fix(hook): stop silent config-gated lint skip in moai gate (card t233)`),
whose source comment names issue #1631 directly. **This SPEC covers oxlint alone.**

The correction is not an aside about wording: it fixes what the remaining work is. The card
text for t550 reads "biome·oxlint 미지원" (biome and oxlint unsupported), which is half
stale, and a reader taking it at face value would scope this SPEC to two linters and find
one of them already done. Two independent measurements agree: the `biomeprj` fixture of
`.moai/reports/t550/baseline.md` (tree `d060e0d13`) shows biome **executing**, and the lead
independently re-measured and confirmed the same.

## §4 Constraints

- One behaviour change (`gate.go`) plus tests, plus a one-line documentation edit that
  records it. No new dependency, no configuration surface.
- [HARD] **The Template-First rule (CLAUDE.local.md §2) answers differently for the two
  files, and one answer must not be applied to both** (full clause: `plan.md §E`):
  - `internal/hook/quality/gate.go` is Go source under `internal/`, **not** a file under
    `internal/template/templates/`. Template-First does **not** apply: no template mirror to
    update, no `make build` obligation. Stated so run-phase does not invent one.
  - `.claude/rules/moai/languages/javascript.md` has **two copies**, currently byte-identical,
    the second under `internal/template/templates/`. Template-First **does** apply: edit the
    template source first, `make build`, then sync the local copy, and leave the two copies
    byte-identical (AC-009).
- [HARD] The javascript.md edit is scoped to the one linter line. That file is a distributed
  template covering 16 programming languages equally (CLAUDE.local.md §15) — no
  restructuring, no added examples, no other language's section touched.
- Test fixtures shadow `npx` with a fake binary, as the sibling biome tests do; no test may
  reach a real `npx` (which would fetch oxlint over the network).
- Artifact language: English (code-adjacent convention).

## §5 Out of Scope

### Out of Scope — the zero-config oxlint project

- **A zero-config oxlint project still runs no lint step after this fix.** oxlint runs
  without any configuration file — "Oxlint works out of the box",
  <https://oxc.rs/docs/guide/usage/linter/config.html> — so such a project presents no
  config-file discriminator, and config-file gating (§3) cannot see it. This gap is closed
  **knowingly**, not overlooked: the decision to leave it, and its three grounds, are
  recorded in §3.
- Reading `package.json` `devDependencies` — the axis that would close it — was
  **considered and rejected by the lead** (§3), not left pending, and is therefore out of
  scope for this SPEC. Nothing here awaits a decision. Should the rejection ever be
  revisited, the delta is one trigger clause in REQ-002 plus one acceptance row.

### Out of Scope — co-resident linter configs

- A Node project carrying **two** linters' config files (an `eslint.config.js` and a
  `biome.json`, say — and after this change, plus an `.oxlintrc.json`) runs **two** lint
  steps, or three. This card does **not** specify that case, and its pre-existing behaviour
  is preserved unchanged: each entry is independently config-gated, so each one whose config
  is present runs.
- **Why no follow-up card is raised:** "unspecified" and "wrong" are different claims, and
  only the first is supported. Running both linters a project actually configured is a
  defensible reading of the config-gate design, and no evidence was produced that it is a
  defect. Raising a card would assert the second claim without a measurement behind it — an
  unobserved defect claim (`verification-claim-integrity.md` §1.1 surface 3). Should evidence
  of an actual problem appear later, that evidence is what justifies the card.

### Out of Scope — other linters and other toolchains

- Adding further Node linters (biome is already present; deno lint, quick-lint-js, xo and
  the rest are not surveyed here).
- Any non-Node toolchain (Go, Python, Rust, and the remaining entries of the `toolchains`
  table) — unchanged by this SPEC.

### Out of Scope — linter verdict fidelity

- Whether `npx oxlint` reports the right violations on a real project. This SPEC governs
  step **selection** and gate outcome propagation; the fixtures use exit-code stubs and
  assert nothing about oxlint's own analysis.

### Out of Scope — the config-absent notice mechanism

- The run-summary skip-notice machinery itself (already present on this tree, pinned by
  `TestSummaryDistinguishesAllFiveSkipPaths`) is consumed, not modified.

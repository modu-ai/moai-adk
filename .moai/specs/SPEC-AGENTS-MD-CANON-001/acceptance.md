# SPEC-AGENTS-MD-CANON-001 — Acceptance Criteria

Every criterion is binary-testable and names the command or artifact that decides it.
Given-When-Then throughout; the GEARS obligations live in `spec.md` §C.

## §A. Entry gate (pre-run)

**AC-AMC-001** — Given the run phase is about to start, When
`go test ./internal/config/ -run 'Budget|AlwaysLoaded'` runs on the entry tree, Then it exits 0 and
its output is recorded as the pre-diet baseline.

**AC-AMC-002** — Given the codex probe fixture, When the three commands in
`.moai/reports/t82/codex-probe.md` § 검증 재현 are re-run, Then the 32,768 B cut point and the
root-only load at repo-root invocation reproduce. (Reproducibility check, not a re-derivation: the
premises are already measured.)

## §B. M1 — classification and compressibility

**AC-AMC-003** — Given the 97 `[HARD]` clauses, When each is classified as Codex-relevant or
Claude-mechanism-only, Then a per-clause table exists with a byte figure per row and zero
unclassified rows, and the Codex-relevant subtotal is stated.

**AC-AMC-004** — Given a rewritten clause, When it is diffed against its original, Then subject,
modality (`shall` / `shall not` / `MUST` / `MUST NOT`), and binding scope are unchanged; any clause
failing this is reverted, not shipped.

**AC-AMC-005** — Given the pilot rewrite, When before/after bytes are tallied, Then the aggregate
compression ratio is stated as a number with its per-clause variance.

**AC-AMC-006** — Given the measured ratio applied to the Codex-relevant subtotal, When the
projection is compared against 24,576 B, Then the verdict (reachable / not reachable) is explicit
with the projected figure.

**AC-AMC-007** — Given the projection exceeds 24,576 B, When M1 concludes, Then a blocker report is
returned naming the shortfall in bytes and the trade against the 8,192 B headroom reserve, and no
file has been moved. (Negative-path criterion: M1 passes by halting correctly, not only by
succeeding.)

## §C. M2 — root contract layer

**AC-AMC-008** — Given the authored root `AGENTS.md`, When every Codex-relevant `[HARD]` clause
from M1's table is enumerated, Then each appears in the root document, and the trace table maps
clause → origin file with zero unmapped rows.

**AC-AMC-009** — Given the shipped `AGENTS.md`, When `wc -c` runs on it, Then the result is at or
below 24,576 B, and the headroom against 32,768 B is stated.

**AC-AMC-010** — Given the repository, When `find . -name AGENTS.md -not -path './.git/*'` runs,
Then exactly one result is returned — the repository root. (REQ-AMC-005: no nested documents.)

**AC-AMC-011** — Given the redistribution is complete, When the skills and lazy companions it
created are grepped for `[HARD]`, `MUST`, `MUST NOT`, and `shall`, Then no relocated obligation is
found — only rationale, procedure, examples, incident records, and cross-references.

## §D. M3 — import layer

**AC-AMC-012** — Given the edited `CLAUDE.md`, When a Claude Code session starts, Then the contract
text is present in context exactly once — no duplicate injection via both import and inline copy.

**AC-AMC-013** — Given the `@`-import chain, When it is resolved, Then every imported path exists
and resolves; an unresolvable import fails this criterion rather than warning.

**AC-AMC-014** — Given the full existing test suite for the affected packages, When it runs, Then
it is green, and no expected-behavior assertion was edited to accommodate the change.

## §E. M5 — guard and ratchet

**AC-AMC-015** — Given a test fixture that pushes `AGENTS.md` one byte past its ceiling, When the
guard runs, Then it **fails the build** — not warns — and its message names the measured byte
figure and the offending file. (REQ-AMC-009: advisory-only does not pass.)

**AC-AMC-016** — Given the byte guard's implementation, When it is read, Then it calls
`alwaysLoadedSurface()` (or the same enumeration helper) rather than re-globbing the rule tree.

**AC-AMC-017** — Given the post-diet integration branch, When
`go test ./internal/config/ -run 'Budget|AlwaysLoaded'` runs, Then it exits 0 with
`AlwaysLoadedTokenBudget` at or below 75,000, and the achieved token figure is quoted from that
run's output — not from a worktree-only measurement.

**AC-AMC-018** — Given the ratcheted constant, When its comment block is read, Then the "temporary
raise pending a separate card" note is retired and replaced by the ratchet's measured
justification.

## §F. M6 — distribution and disclosure

**AC-AMC-019** — Given every file this SPEC lands under `.claude/`, `.moai/`, or the repo root that
ships to users, When `internal/template/templates/` is checked, Then a mirror exists for each, and
`make build` has been run after the last mirror edit.

**AC-AMC-020** — Given the template mirrors, When the neutrality guard runs, Then it passes: no
SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs, macOS-biased absolute paths, or
`CLAUDE.local.md` references in any template copy.

**AC-AMC-021** — Given the shipped documentation, When it is grepped for `~/.codex/AGENTS.md`, Then
the warning is present and states all three facts: the global document joins the same chain, it is
consumed before the project document, and it narrows the project's available budget.

## §G. Edge cases

- A `[HARD]` clause binding Claude rendering only (the 75 in the output style) is classified out of
  the contract rather than copied into `AGENTS.md`.
- A clause appearing in two rule files is de-duplicated in `AGENTS.md` without either origin losing
  its trace-table row.
- A rule file that drops to zero `[HARD]` clauses after extraction still exists as a navigable stub
  with pointers, not an empty file.
- A clause inside one of the six Claude-mechanism-bound files that states a harness-generic
  principle is classified Codex-relevant, not excluded by file membership.

## §H. Definition of Done

- All AC in §A-§F pass, each with its command output recorded.
- Exactly one `AGENTS.md` exists, at the repository root, at or below 24,576 B.
- `spec.md` §E.2's ruling holds: the output style was not edited by this SPEC.
- No `[HARD]` clause left the always-loaded surface.
- The run-phase completion report uses the 5-section evidence format (Claim / Evidence /
  Baseline-attribution / Gaps / Residual-risk), with every byte and token figure attributed to a
  command run on the named tree.

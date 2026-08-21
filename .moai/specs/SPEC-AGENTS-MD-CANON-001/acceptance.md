# SPEC-AGENTS-MD-CANON-001 — Acceptance Criteria

Every criterion is binary-testable and names the command or artifact that decides it.
Given-When-Then throughout; the GEARS obligations live in `spec.md` §C.

## §A. Entry gate (pre-run)

**AC-AMC-001** — Given the run phase is about to start, When the four entry premises in
`spec.md` §D.1 are checked, Then each of P1, P2, P3, P4 has a recorded measurement with the
command that produced it, and no premise is recorded as "documented default" alone.

**AC-AMC-002** — Given the entry tree, When `go test ./internal/config/ -run 'Budget|AlwaysLoaded'`
runs, Then it exits 0, and its output is the recorded pre-diet baseline.

## §B. M1 — compressibility

**AC-AMC-003a** — Given the two pilot files (`kanban-dispatch.md`, `native-idiom-and-register.md`),
When their `[HARD]` clauses are rewritten, Then a per-clause before/after byte table exists and
the aggregate compression ratio is stated as a number.

**AC-AMC-003b** — Given the measured ratio, When it is extrapolated over the full 32,543 B
contract layer, Then the projected root-document size is stated in bytes and compared against the
P1-confirmed cap, and the comparison's verdict (reachable / not reachable) is explicit.

**AC-AMC-004** — Given a rewritten clause, When it is diffed against its original, Then subject,
modality (`shall` / `shall not` / `MUST` / `MUST NOT`), and binding scope are unchanged; any
clause failing this is reverted, not shipped.

**AC-AMC-005** — Given the projection exceeds the cap minus the stated nested reserve, When M1
concludes, Then a blocker report is returned naming the shortfall in bytes, and no file has been
moved. (Negative-path criterion: this SPEC passes M1 by halting correctly, not only by succeeding.)

## §C. M2/M3 — contract layer and directory map

**AC-AMC-006** — Given the authored root `AGENTS.md`, When every `[HARD]` clause from the 14
always-loaded rules and `CLAUDE.md` is enumerated, Then each one appears in the root `AGENTS.md`
or a nested `AGENTS.md`, and the trace table maps clause → origin file with zero unmapped rows.

**AC-AMC-007** — Given the shipped `AGENTS.md` set, When the deepest reachable merge chain is
summed with `wc -c`, Then the total is at or below the P1-confirmed `project_doc_max_bytes`.

**AC-AMC-008** — Given the redistribution is complete, When the skills and lazy companions it
created are grepped for `[HARD]`, `MUST`, `MUST NOT`, and `shall`, Then no relocated obligation is
found — only rationale, procedure, examples, incident records, and cross-references.

**AC-AMC-009** — Given P2's measured merge scope, When M3 concludes, Then either the nested
directory map exists with each file's owned directory and clause set named, or a recorded ruling
states that nested merging is unavailable and the nested leg is dropped. Silence on this point is
a failure.

## §D. M4 — import layer

**AC-AMC-010** — Given the edited `CLAUDE.md`, When a Claude Code session starts, Then the
contract text is present in context exactly once (no duplicate injection via both import and
inline copy).

**AC-AMC-011** — Given the `@`-import chain, When it is resolved, Then every imported path exists
and resolves; an unresolvable import fails this criterion rather than warning.

**AC-AMC-012** — Given the full existing test suite for the affected packages, When it runs, Then
it is green, and no expected-behavior assertion was edited to accommodate the change.

## §E. M6 — guard and ratchet

**AC-AMC-013** — Given a deliberate test fixture that pushes the merged chain one byte past the
cap, When the guard runs, Then it fails, and its message names the measured byte figure and the
offending file.

**AC-AMC-014** — Given the byte guard's implementation, When it is read, Then it calls
`alwaysLoadedSurface()` (or the same enumeration helper) rather than re-globbing the rule tree.

**AC-AMC-015** — Given the post-diet integration branch, When
`go test ./internal/config/ -run 'Budget|AlwaysLoaded'` runs, Then it exits 0 with
`AlwaysLoadedTokenBudget` at or below 75,000, and the achieved token figure is quoted from that
run's output — not from a worktree-only measurement.

**AC-AMC-016** — Given the ratcheted constant, When its comment block is read, Then the
"temporary raise pending a separate card" note is retired and replaced by the ratchet's
measured justification.

## §F. M7 — distribution

**AC-AMC-017** — Given every file this SPEC lands under `.claude/`, `.moai/`, or the repo root
that ships to users, When `internal/template/templates/` is checked, Then a corresponding mirror
exists for each, and `make build` has been run after the last mirror edit.

**AC-AMC-018** — Given the template mirrors, When the neutrality guard runs, Then it passes: no
SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs, macOS-biased absolute paths,
or `CLAUDE.local.md` references appear in any template copy.

## §G. Edge cases

- A `[HARD]` clause that binds Claude rendering only (the 75 in the output style) is correctly
  classified as out of scope rather than copied into `AGENTS.md`.
- A clause appearing in two rule files is de-duplicated in `AGENTS.md` without either origin
  losing its trace-table row.
- A rule file that drops to zero `[HARD]` clauses after extraction still exists as a navigable
  stub with pointers, not an empty file.

## §H. Definition of Done

- All AC in §A-§F pass, each with its command output recorded.
- `spec.md` §E.2's ruling on the output style holds: it was not edited by this SPEC.
- No `[HARD]` clause left the always-loaded surface.
- The 5-section evidence format (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)
  is used for the run-phase completion report, with every byte and token figure attributed to a
  command run on the named tree.

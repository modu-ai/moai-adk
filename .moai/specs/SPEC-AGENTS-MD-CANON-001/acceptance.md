# SPEC-AGENTS-MD-CANON-001 — Acceptance Criteria

Every criterion is binary-testable and names the command or artifact that decides it.
Given-When-Then throughout; the GEARS obligations live in `spec.md` §C.

> **Executability note.** `TestAlwaysLoadedTokenBudget` currently emits the token total only via
> `t.Errorf` on failure, so a passing run prints `ok …` and nothing else. M5 adds a `t.Logf` ahead
> of the over-budget check; every criterion below that quotes a token figure runs under
> `go test -v` and reads that logged line. A criterion demanding a figure a passing run does not
> print would require the run to pass and fail at once — that shape is what this revision removed.

## §A. Entry gate (pre-run)

**AC-AMC-001** — Given the entry tree, When
`go test -v ./internal/config/ -run 'Budget|AlwaysLoaded'` runs, Then it exits 0 and its output
carries the `always-loaded surface = N tokens` line; that N is recorded as the pre-diet baseline
alongside `git rev-parse --abbrev-ref HEAD` and `git rev-parse --short HEAD`.

**AC-AMC-002** — Given `.moai/reports/t82/probe-fixture.sh`, When it is executed on a clean
machine, Then it rebuilds the probe fixture from scratch and reports each recorded run against its
expected result — the 32,768 B cut at `MARK32670`, and nested-marker count 0 at repo-root
invocation. (The scratchpad path the earlier draft cited does not survive the session; this script
is the executable form.)

## §B. M1 — classification and compressibility

**AC-AMC-003** — Given the `[HARD]` marker set, When each marker is expanded to its **clause
block** — the marker line plus its continuation to the next clause or heading — and each block is
classified Codex-relevant or Claude-mechanism-only, Then a per-block table exists with a byte
figure per row, zero unclassified rows, and the Codex-relevant subtotal stated. Line-level grep
counts do not satisfy this criterion (`spec.md` §A.4: 15 of 93 markers are prose, 16 lead into
uncounted bodies).

**AC-AMC-004** — Given a rewritten clause, When it is diffed against its original, Then subject,
modality (`shall` / `shall not` / `MUST` / `MUST NOT`), and binding scope are unchanged; any clause
failing this is reverted, not shipped.

**AC-AMC-005** — Given the pilot rewrite, When before/after bytes are tallied, Then the aggregate
compression ratio is stated as a number with its per-clause variance.

**AC-AMC-006** — Given the measured ratio applied to the clause-block Codex-relevant subtotal, When
the projection is compared against 24,576 B, Then the verdict (reachable / not reachable) is
explicit with the projected figure, and the ceiling is re-derived against the clause-block figure
rather than the line proxy.

**AC-AMC-007** — Given the projection exceeds 24,576 B, When M1 concludes, Then a blocker report is
returned naming the shortfall in bytes and the trade against the 8,192 B reserve, and no file has
been moved. (Negative-path criterion: M1 passes by halting correctly, not only by succeeding.)

## §C. M2 — contract layer

**AC-AMC-008** — Given the authored root `AGENTS.md`, When every Codex-relevant clause block from
M1's table is enumerated, Then each appears in the root document, and the trace table maps clause →
origin file with zero unmapped rows.

**AC-AMC-009** — Given the shipped `AGENTS.md`, When `wc -c` runs on **both** the live root file
and its template mirror `internal/template/templates/AGENTS.md`, Then each is at or below 24,576 B
and the headroom against 32,768 B is stated for each. The ceiling binds both copies: the mirror is
what users receive, so a mirror over the ceiling truncates on their machines regardless of the live
file's size.

**AC-AMC-010** — Given the repository, When

```
git ls-files --full-name ':(top)*AGENTS.md' ':(exclude,top)internal/template/templates/*'
```

runs, Then the output is exactly `AGENTS.md` — one line, the repository root.

Three properties make this decidable where a filename count is not. **Tracked files only**, so card
worktrees under `.claude/worktrees/` and build output cannot contribute. **`:(top)` pathspec
magic**, so the result does not depend on the directory the criterion is run from. **The template
mirror explicitly excluded**, because REQ-AMC-015 requires it to exist.

A global `find . -name AGENTS.md` is not a valid substitute and must not be reintroduced: it counts
the mirror M6 creates — making `AC-AMC-010` and `REQ-AMC-015` unsatisfiable together — and it
varies with how many card worktrees happen to be live. Measured against the `CLAUDE.md` analogue in
this worktree: `find` returns **7**, while the command above returns the 6 live-tree files with the
mirror correctly excluded.

**AC-AMC-011** — Given the redistribution is complete, When the skills and lazy companions it
created are grepped for `[HARD]`, `MUST`, `MUST NOT`, and `shall`, Then no relocated obligation is
found — only rationale, procedure, examples, incident records, and cross-references.

**AC-AMC-012** — Given `spec.md` §D.6, When it is read, Then it states both revival conditions —
the evidence class (observed session-CWD habit, not asserted convenience) and the obligation to
lower the root ceiling by at least the nested document's size — and the nested-`CLAUDE.md`
asymmetry that motivates the asymmetric treatment. (Covers REQ-AMC-006, which binds this SPEC's
record; a future SPEC's compliance with the condition is that SPEC's criterion, not this one's.)

## §D. M3 — import layer

**AC-AMC-013** — Given the resolved import set (`CLAUDE.md` plus every `@`-imported file), When a
duplicate-line scan runs over it —
`cat CLAUDE.md AGENTS.md | grep -n '\[HARD\]' | sort -k2 | uniq -d -f1` — Then it returns no
output: no contract line appears in both the Claude-only layer and the imported `AGENTS.md`.

**AC-AMC-014** — Given the `@`-import chain, When each imported path is resolved, Then every path
exists; an unresolvable import fails this criterion rather than warning.

**AC-AMC-015** — Given the full existing test suite for the affected packages, When it runs, Then
it is green, and no expected-behavior assertion was edited to accommodate the change.

## §E. M5 — guard and ratchet

**AC-AMC-016** — The negative path has two dimensions, and only one of them needs new coverage.

**(a) Token-budget over-budget detection — already covered; do not write a second fixture.**
Given `TestAlwaysLoadedTokenBudget_OverBudgetFails`
(`internal/config/token_budget_guard_test.go`), When
`go test -v ./internal/config/ -run 'OverBudgetFails'` runs, Then both subtests pass: the
`over-budget` case plants an `AlwaysLoadedTokenBudget*4 + 4096` byte file in a temp tree and
asserts the guard fires; the `under-budget` case asserts it does not. Measured on this tree:
`--- PASS: …/over-budget`, `--- PASS: …/under-budget`, `ok … 0.561s`. A new fixture duplicating
this would be the second measurement path REQ-AMC-008 forbids.

**(b) Codex-cap byte-guard breach — genuinely new coverage.** Given a case that pushes `AGENTS.md`
one byte past its ceiling, When the guard runs, Then it **fails the build** — not warns — and its
message names the measured byte figure and the offending file (REQ-AMC-009). This case extends the
existing table-driven test in the same file and reuses the same path-resolution and repo-root
helpers; it does not introduce a parallel harness.

> Why (a) is cited rather than rebuilt: the existing test proves the guard *fires* when the surface
> is too large. It proves nothing about whether the constant was set from the achieved figure or
> simply parked at the ceiling — that gap is `AC-AMC-019`'s, and it is the only part of the ratchet
> still genuinely uncovered.

**AC-AMC-017** — Given the byte guard's implementation, When it is read, Then it calls
`alwaysLoadedSurface()` (or the same enumeration helper) rather than re-globbing the rule tree.

**AC-AMC-018** — Given the post-diet **integration branch** — the `release/vX.Y.Z` branch this card
merges into, carrying the merged sibling state — When
`go test -v ./internal/config/ -run 'Budget|AlwaysLoaded'` runs there, Then it exits 0, the logged
`always-loaded surface = N tokens` line is quoted, `AlwaysLoadedTokenBudget` is at or below 75,000,
and the evidence records `git rev-parse --abbrev-ref HEAD` and `git rev-list --count main..HEAD` so
the measured tree is identified rather than asserted.

**AC-AMC-019** — Given the achieved figure N from AC-AMC-018 and the headroom ratio stated in the
constant's comment, When `AlwaysLoadedTokenBudget` is compared against `N × (1 + ratio)`, Then the
two agree within ±1,000 tokens. (Without this, landing the surface at 60,000 and setting the
constant to 75,000 would satisfy the ceiling check verbatim while ratcheting nothing.)

**AC-AMC-020** — Given the ratcheted constant, When its comment block is read, Then the "temporary
raise pending a separate card" note is retired, replaced by the ratchet's measured justification,
and the headroom ratio AC-AMC-019 checks against is stated there.

## §F. M6 — distribution and disclosure

**AC-AMC-021** — Given every file this SPEC lands under `.claude/`, `.moai/`, or the repo root that
ships to users, When `internal/template/templates/` is checked, Then a mirror exists for each, and
`make build` has been run after the last mirror edit.

**AC-AMC-022** — Given the template mirrors, When the neutrality guard runs, Then it passes: no
SPEC IDs, REQ tokens, audit citations, internal dates, commit SHAs, macOS-biased absolute paths, or
`CLAUDE.local.md` references in any template copy.

**AC-AMC-023** — Given the shipped documentation, When it is grepped for `~/.codex/AGENTS.md`, Then
the warning is present and states all three facts: the global document joins the same chain, it is
consumed before the project document, and it narrows the project's available budget.

**AC-AMC-024** — Given `spec.md` §D.8 and REQ-AMC-018, When they are read, Then the cap-raise
position asserted by the SPEC is the measured one: project scope works only under
`trust_level = "trusted"`, the untrusted first session is the binding case at 32,768 B, and
non-application is silent. The retired "cannot ship" premise MAY appear only inside an explicit
correction that marks it false — a bare assertion of it anywhere fails this criterion.

## §G. Edge cases

- A `[HARD]` clause binding Claude rendering only (the 75 in the output style) is classified out of
  the contract rather than copied into `AGENTS.md`.
- A marker that is a prose mention rather than an obligation (15 of the 93) is classified out at
  M1, not counted as contract.
- A `[HARD]` lead line ending in `:` has its continuation body included in its clause block.
- A clause appearing in two rule files is de-duplicated in `AGENTS.md` without either origin losing
  its trace-table row.
- A rule file that drops to zero `[HARD]` clauses after extraction still exists as a navigable stub
  with pointers, not an empty file.
- A clause inside one of the six Claude-mechanism-bound files that states a harness-generic
  principle is classified Codex-relevant, not excluded by file membership.

## §H. Definition of Done

- All AC in §A-§F pass, each with its command output recorded.
- Exactly one live-tree `AGENTS.md`, at the repository root, at or below 24,576 B — and its
  template mirror likewise.
- `spec.md` §E.2's ruling holds: the output style was not edited by this SPEC.
- No `[HARD]` clause left the always-loaded surface.
- The run-phase completion report uses the 5-section evidence format (Claim / Evidence /
  Baseline-attribution / Gaps / Residual-risk), with every byte and token figure attributed to a
  command run on the named tree.

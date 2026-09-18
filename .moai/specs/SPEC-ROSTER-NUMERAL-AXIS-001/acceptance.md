# SPEC-ROSTER-NUMERAL-AXIS-001 — Acceptance Criteria

frontmatter-less sibling artifact. Verification scope is
`./internal/harness/rosterguard/...` plus any package actually touched; the full-suite verdict
belongs to CI.

## §A Scope of Verification

The layer is verified on four independent properties: it FIRES on an undeclared count claim
(AC-RNA-001), it discharges only on the D2 rule (AC-RNA-002), it selects the right numeral
(AC-RNA-003), and it can be shown to be alive rather than silent (AC-RNA-006 breadth printing,
AC-RNA-007 control probe, AC-RNA-008 anti-vacuity).

## §D AC Matrix

| AC | Requirement | Severity | Verification surface |
|----|-------------|----------|----------------------|
| AC-RNA-001 | REQ-RNA-001, REQ-RNA-002 | Must | Synthetic undeclared count claim fails the guard |
| AC-RNA-002 | REQ-RNA-004 | Must | Discharge table: `ClaimCount` row, exempt row, membership-only row |
| AC-RNA-003 | REQ-RNA-003 | Must | Nearest-numeral regression case |
| AC-RNA-004 | REQ-RNA-005 | Must | Empty-reason exempt row produces a finding |
| AC-RNA-005 | REQ-RNA-006 | Must | Word axis: synthetic positive control fires, live count 0 |
| AC-RNA-006 | REQ-RNA-008 | Must | Anchored grep over printed hit lines |
| AC-RNA-007 | REQ-RNA-010 | Must | Control probe observed to fire |
| AC-RNA-008 | REQ-RNA-009 | Must | Zero-hit run fails |
| AC-RNA-009 | REQ-RNA-007 | Must | Selector phrases produce no finding |
| AC-RNA-010 | REQ-RNA-011 | Must | Newly-reached stale sites carry `CountPattern` + `KnownStale` |
| AC-RNA-011 | spec.md §C (ordering) | Must | Baseline commit precedes the implementation commit |
| AC-RNA-012 | spec.md §C (PRESERVE) | Must | t922 tests pass unchanged |

### AC-RNA-001 — An undeclared count claim fails the guard

**Given** a file inside the swept tree carrying a numeral adjacent to a noun in the D1 class,
with no `Registry()` row and no numeral-exempt declaration for its path
**When** the numeral layer runs
**Then** the guard reports that path with a message naming the observed numeral, the matched
noun phrase, and the two ways to discharge it (register a `ClaimCount` row, or declare a
numeral exemption with a reason) — and the test FAILS.

### AC-RNA-002 — Discharge follows the D2 rule, and only it

**Given** three registry states for one path — (a) a row with `Claims` including `ClaimCount`,
(b) no row but a numeral-exempt declaration with a non-empty reason, (c) a row carrying
`ClaimMembership` only
**When** the layer evaluates a numeral hit in that path
**Then** (a) and (b) discharge the hit, and (c) does NOT — (c) produces a finding stating that a
membership registration does not discharge a count claim.

### AC-RNA-003 — Nearest preceding numeral, not leftmost

**Given** the literal input `CLAUDE.md 4 (13 retained agents)`
**When** the layer selects the numeral for the matched noun
**Then** the reported numeral is `13`, not `4`; a leftmost-numeral implementation fails this
case.

### AC-RNA-004 — An empty exempt reason is a finding, never a suppression

**Given** a numeral-exempt declaration whose reason is empty or whitespace-only
**When** the layer evaluates a hit in that path
**Then** the layer emits a finding reporting the declaration as incomplete, the hit is NOT
suppressed, and the test fails.

### AC-RNA-005 — The word axis is alive and its live population is 0

**Given** the word-axis regexp and a synthetic input of the form
`thirteen retained agents` that is NOT in the tree
**When** the word axis runs against (a) that synthetic input and (b) the live tree
**Then** (a) produces at least one hit — proving the regexp fires — and (b) produces a live hit
set of size 0 under the size-claim reading, with the 0 reported explicitly alongside (a) so a
dead regexp and an empty population are distinguishable.

### AC-RNA-006 — Breadth is printed, one line per hit, at column 0

**Given** the layer's hit set
**When** the package test runs
**Then** each hit is printed on its own line beginning at column 0 with a stable prefix
(`digit-axis hit ` / `word-axis hit `), such that
`go test -count=1 ./internal/harness/rosterguard/... -v | grep -c '^digit-axis hit '` returns
the digit-axis hit count — a `t.Logf`-based implementation fails this grep.

### AC-RNA-007 — The control probe fires, and is observed firing

**Given** the control probe's deliberately-wrong input
**When** the probe runs
**Then** every probe case produces the violation it declares; and the probe's execution is
observed in the same run (`go test … -run <probe> -v` shows it RUN, not `no tests to run`),
followed by a full-package run with no `-run` filter.

### AC-RNA-008 — A zero-hit run is a measurement failure

**Given** a tree that demonstrably carries count claims (the registry is non-empty by
construction)
**When** the numeral layer produces an empty hit set
**Then** the test FAILS with a message stating that zero hits means the walk, the noun class, or
the exclusion list is broken — not that the tree is clean.

### AC-RNA-009 — Selector phrasing produces no finding

**Given** the live phrases `one of the 11 retained agents` and `four are retained agents`
**When** the layer runs
**Then** neither produces a finding, and the neutralisation is exercised by an explicit test
case for each shape rather than being an accident of the window width.

### AC-RNA-010 — Newly-reached stale sites become enumerable, not exempt

**Given** the sites the layer newly reaches whose declared count disagrees with
`template.ProfileMatrixAgents()`
**When** run-phase completes
**Then** each carries a `Registry()` row with a `CountPattern` matching exactly once and a
`KnownStale` marker recording `DeclaredCount`, reason and follow-up; none is discharged by a
bare exemption; and the stale prose itself is unmodified (`git diff` empty on those content
lines).

### AC-RNA-011 — Baseline-first ordering is witnessed by the commit graph

**Given** the run-phase baseline artifact
**When** `git log --follow` is read over the branch
**Then** the baseline artifact's commit is an ANCESTOR of the first implementation commit — not
the same commit. A same-commit pair fails this criterion regardless of what the commit message
asserts.

### AC-RNA-012 — The t922 surface is preserved

**Given** `TestSweepFindsNoUndeclaredRosterListing`, `TestGuardFiresOnDeliberatelyWrongInput`,
`TestKnownStaleInventory` and the per-site checks
**When** the numeral layer lands
**Then** all pass without modification to their assertions, and `SweepUnreachable` semantics are
unchanged.

## §D.1 Edge Cases

- A path carrying BOTH a membership row and a separate count row — the count row discharges;
  the membership row alone would not.
- A numeral written with a thousands separator or attached punctuation (`13,`/`(13)`) adjacent
  to the noun.
- The noun inside a longer word (`retained agents-reference`) — the word boundary must prevent
  a match the same way `tabNoun` prevents `tab` inside `selectable`.
- A file that is both a template mirror and a local copy: two paths, two rows, no collapsing.
- rosterguard's own files describing the guard in prose — a measured false-positive class that
  must be handled by mechanism or by a reasoned exempt row, never silently.

## §D.2 Quality Gates

- `go test -count=1 ./internal/harness/rosterguard/...` green
- `golangci-lint run` clean on touched packages; `gofmt` clean; `go vet` clean
- Coverage ≥85% on the new layer's code
- Conventional commits; `card t930` named in every commit message on this branch

## §D.3 Definition of Done

- AC-RNA-001 … AC-RNA-012 all Must-severity PASS with cited in-run evidence (command + verbatim
  output) recorded in `progress.md` §E.2
- Plan-phase population figures re-derived in-run and cited as the run baseline
- Registry expansion complete, every new row carrying its own `CountPattern` and `KnownStale`
- Template-First pairs edited on both sides where touched

## §D.4 Traceability

REQ-RNA-001→AC-RNA-001 · REQ-RNA-002→AC-RNA-001 · REQ-RNA-003→AC-RNA-003 ·
REQ-RNA-004→AC-RNA-002 · REQ-RNA-005→AC-RNA-004 · REQ-RNA-006→AC-RNA-005 ·
REQ-RNA-007→AC-RNA-009 · REQ-RNA-008→AC-RNA-006 · REQ-RNA-009→AC-RNA-008 ·
REQ-RNA-010→AC-RNA-007 · REQ-RNA-011→AC-RNA-010 · spec.md §C→AC-RNA-011, AC-RNA-012

🗿 MoAI

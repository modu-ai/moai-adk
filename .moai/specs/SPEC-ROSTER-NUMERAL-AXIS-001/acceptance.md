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
| AC-RNA-013 | spec.md §D D3, REQ-RNA-004 | Must | Every hit accounted for; residual arithmetic closes |
| AC-RNA-014 | REQ-RNA-012 | Must | In-run re-derivation is the run baseline |
| AC-RNA-015 | spec.md §E (historical citations) | Must | Historical citations: no finding, no repair, reason if exempted |
| AC-RNA-016 | REQ-RNA-013 | Must | Layer scope = sweep exclusions + `.moai/research/` |

### AC-RNA-001 — An undeclared count claim fails the guard

**Given** a file inside the layer's scope (REQ-RNA-013) carrying a numeral adjacent to a noun in
the D1 class,
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

**Given** two inputs — (a) the LIVE string `CLAUDE.md §4 (the 13 retained agents` from
`internal/harness/delegationmap/types.go:73` (read in this tree; it is also the string
`axis.go`'s `CountPattern` doc comment cites as the reason the pattern is a regexp rather than a
line anchor), and (b) the synthetic variant `CLAUDE.md 4 (13 retained agents)`
**When** the layer selects the numeral for the matched noun
**Then** both report `13`, not `4`. (a) is the load-bearing case because it is the shape that
actually occurs — it differs from (b) in the section marker and the intervening word count, so
it exercises a different adjacency; (b) is kept as an additional row.

### AC-RNA-004 — An empty exempt reason is a finding, never a suppression

**Given** a numeral-exempt declaration whose reason is empty or whitespace-only
**When** the layer evaluates a hit in that path
**Then** the layer emits a finding reporting the declaration as incomplete, the hit is NOT
suppressed, and the test fails.

### AC-RNA-005 — The word axis is alive, and its reported population is 0 for a named reason

**Given** the word-axis regexp, a synthetic input of the form `thirteen retained agents` that is
NOT in the tree, and the neutralise → match → report → discharge order fixed by REQ-RNA-007
**When** the word axis runs against (a) that synthetic input and (b) the live tree
**Then** (a) produces at least one hit — proving the regexp fires — and (b) the live *reported*
word-axis hit set is empty **because** `.claude/agents/harness/workflow-specialist.md:52`
(`All four are retained agents.`) is removed by the selector mechanism BEFORE counting, not
because it was counted and then discharged by an exempt row. The run evidence names that file
and the neutralisation that removed it, so the 0 is pinned to an observed cause rather than an
expectation.

### AC-RNA-006 — Breadth is printed, one line per hit, at column 0

**Given** the layer's hit set and the INDEPENDENTLY derived population from AC-RNA-014 (the
in-run re-derivation, which does not consult the layer's own counter)
**When** the package test runs
**Then** both hold:

(a) each hit is printed on its own line beginning at column 0 with a stable prefix
(`digit-axis hit ` / `word-axis hit `), such that
`go test -count=1 ./internal/harness/rosterguard/... -v | grep -c '^digit-axis hit '` is
executable — a `t.Logf`-based implementation fails this grep because it indents and prefixes
with `file:line`; and

(b) the printed hit set EQUALS the union of (registered `ClaimCount` paths ∪ exempt paths ∪
mechanism-neutralised paths) exactly — no more and no less — and its size equals the AC-RNA-014
re-derived population. The equality is the assertion the reuse target already makes
(`docsTabAllowlist` asserts the hit set equals the allowlist exactly); comparing the printed
count to the implementation's own count is circular and a mutant printing only the first hit
passes it.

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

### AC-RNA-009 — Selector phrasing is neutralised before counting

**Given** the live phrases `one of the 11 retained agents` and `All four are retained agents.`
(the latter at `.claude/agents/harness/workflow-specialist.md:52`)
**When** the layer runs under the REQ-RNA-007 order (neutralise → match → report → discharge)
**Then** neither phrase reaches the match stage, so neither appears in the reported hit set of
AC-RNA-006 and neither is discharged by an exempt row; the neutralisation is exercised by an
explicit test case for each shape rather than being an accident of the window width. This
criterion and AC-RNA-005(b) refer to the SAME string under the SAME ordering and therefore
cannot hold opposite expectations of it.

### AC-RNA-010 — Newly-reached stale sites become enumerable, not exempt

**Given** the sites the layer newly reaches whose declared count disagrees with
`template.ProfileMatrixAgents()`
**When** run-phase completes
**Then** each carries a `Registry()` row with a `CountPattern` matching exactly once and a
`KnownStale` marker recording `DeclaredCount`, reason and follow-up; none is discharged by a
bare exemption; and the stale prose itself is unmodified (`git diff` empty on those content
lines).

### AC-RNA-011 — Baseline-first ordering is witnessed by the commit graph

**Given** the run-phase baseline artifact's commit `$BASE`, and the **first implementation
commit** `$IMPL` — defined mechanically as the first commit on this branch that modifies a file
under `internal/harness/rosterguard/`:

```
IMPL=$(git log --reverse --format=%H WT-numeral-roster-guard -- internal/harness/rosterguard | head -1)
```

**When** the ancestry is decided by a single invocation
**Then** `git merge-base --is-ancestor "$BASE" "$IMPL"` exits **0** and `$BASE != $IMPL`. A
same-commit pair fails this criterion regardless of what the commit message asserts — the commit
graph is the only sequencing witness (`verification-claim-integrity.md` §2.3).

### AC-RNA-012 — The t922 surface is preserved

**Given** `TestSweepFindsNoUndeclaredRosterListing`, `TestGuardFiresOnDeliberatelyWrongInput`,
`TestKnownStaleInventory` and the per-site checks
**When** the numeral layer lands
**Then** all pass without modification to their assertions, and `SweepUnreachable` semantics are
unchanged.

### AC-RNA-013 — Every hit is accounted for; the residual arithmetic closes

**Given** the in-run population of AC-RNA-014
**When** run-phase completes
**Then** the run evidence states the four quantities — total hits, hits discharged by a
pre-existing `ClaimCount` row, hits newly registered, hits neutralised by mechanism, and hits
exempted — and their sum EQUALS the total; no hit is left implicit, and the count of authored
rows a reviewer must read is reported alongside the derived-mirror count so the D3 saving is an
observed figure rather than the projection §D calls it.

### AC-RNA-014 — The run baseline is re-derived in-run

**Given** REQ-RNA-012 and the tree being changed
**When** run-phase begins
**Then** the digit-axis and word-axis populations and the residual arithmetic are produced by a
command run in THIS tree in THIS run, and the run evidence cites that command with its verbatim
output. Citing the plan-phase figures of spec.md §A as the baseline fails this criterion even
when the numbers happen to agree.

### AC-RNA-015 — Historical citations: no finding, no repair, reason where exempted

**Given** the live historical-citation sites — `the then-8-agent catalog` in
`.claude/agents/moai/manager-docs.md`, `.claude/agents/moai/manager-spec.md` and their
`internal/template/templates/` mirrors, and `17->8 agent catalog` in
`internal/template/*_test.go`
**When** the layer runs, by whichever route run-phase chooses (tense/selector mechanism, or a
per-path exempt declaration)
**Then** none produces a finding, none is repaired (`git diff` empty on those content lines),
and every one handled by exemption rather than by mechanism carries a non-empty per-path reason
a reviewer can disagree with. The mechanism choice stays deferred; this outcome does not.

### AC-RNA-016 — The layer's scope is stated and matches the sweep's

**Given** REQ-RNA-013
**When** the layer runs
**Then** its exclusion set is `sweepSkipPrefixes` ∪ `sweepSkipFiles` ∪ `{.moai/research/}`,
stated in ONE place in code and cited rather than re-listed; and
`.moai/research/anthropic-best-practices-2026-05-24.md` produces no hit — verified by asserting
the file is out of scope, not by an exempt row.

## §D.1 Edge Cases

- A path carrying BOTH a membership row and a separate count row — the count row discharges;
  the membership row alone would not.
- A numeral written with a thousands separator or attached punctuation (`13,`/`(13)`) adjacent
  to the noun.
- The noun inside a longer word (`retained agents-reference`) — the word boundary must prevent
  a match the same way `tabNoun` prevents `tab` inside `selectable`.
- A template mirror whose local counterpart was deleted or renamed — `mirrorOf()` must fail
  loudly rather than derive a row for a path that no longer has a source.
- A mirror whose content has diverged from its local counterpart: derivation must not assert a
  claim the mirror does not actually make.
- rosterguard's own files describing the guard in prose — a measured false-positive class that
  must be handled by mechanism or by a reasoned exempt row, never silently.

## §D.2 Quality Gates

- `go test -count=1 ./internal/harness/rosterguard/...` green
- `golangci-lint run` clean on touched packages; `gofmt` clean; `go vet` clean
- Coverage ≥85% on the new layer's code
- Conventional commits; `card t930` named in every commit message on this branch

## §D.3 Definition of Done

- AC-RNA-001 … AC-RNA-016 all Must-severity PASS with cited in-run evidence (command + verbatim
  output) recorded in `progress.md` §E.2
- Plan-phase population figures re-derived in-run and cited as the run baseline (AC-RNA-014)
- Residual arithmetic closed with no implicit hit (AC-RNA-013), and the observed authored-row
  count reported against the D3 projection
- Registry expansion complete, every new row carrying its own `CountPattern` and `KnownStale`,
  mirror rows derived via `mirrorOf()` rather than hand-copied
- Template-First pairs edited on both sides where touched

## §D.4 Traceability

REQ-RNA-001→AC-RNA-001 · REQ-RNA-002→AC-RNA-001 · REQ-RNA-003→AC-RNA-003 ·
REQ-RNA-004→AC-RNA-002 · REQ-RNA-005→AC-RNA-004 · REQ-RNA-006→AC-RNA-005 ·
REQ-RNA-007→AC-RNA-009, AC-RNA-005 · REQ-RNA-008→AC-RNA-006 · REQ-RNA-009→AC-RNA-008 ·
REQ-RNA-010→AC-RNA-007 · REQ-RNA-011→AC-RNA-010 · REQ-RNA-012→AC-RNA-014 ·
REQ-RNA-013→AC-RNA-016 · spec.md §C→AC-RNA-011, AC-RNA-012 · spec.md §D D3→AC-RNA-013 ·
spec.md §E (historical citations)→AC-RNA-015

🗿 MoAI

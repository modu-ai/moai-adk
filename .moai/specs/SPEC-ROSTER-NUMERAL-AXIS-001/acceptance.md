# SPEC-ROSTER-NUMERAL-AXIS-001 — Acceptance Criteria

frontmatter-less sibling artifact. Verification scope is
`./internal/harness/rosterguard/...` plus any package actually touched; the full-suite verdict
belongs to CI.

## §A Scope of Verification

The layer is verified on four independent properties: it FIRES on an undeclared count claim
(AC-RNA-001), it discharges only on the D2 rule (AC-RNA-002), it selects the right numeral
(AC-RNA-003), and it can be shown to be alive rather than silent (AC-RNA-006 breadth printing,
AC-RNA-007 control probe, AC-RNA-008 anti-vacuity).

**Two sets, named apart (REQ-RNA-008).** The **breadth set** is every post-neutralisation match,
discharged or not; the **finding set** is its undischarged subset, and is what fails the guard.
Neutralised phrases are in neither. Every criterion below says which set it means; a criterion
that said "hit set" without saying which would be unevaluable.

## §D AC Matrix

| AC | Requirement | Severity | Verification surface |
|----|-------------|----------|----------------------|
| AC-RNA-001 | REQ-RNA-001, REQ-RNA-002 | Must | Synthetic undeclared count claim fails the guard |
| AC-RNA-002 | REQ-RNA-004 | Must | Discharge table: `ClaimCount` row, exempt row, membership-only row |
| AC-RNA-003 | REQ-RNA-003 | Must | Nearest-numeral regression case (live string primary) |
| AC-RNA-004 | REQ-RNA-005 | Must | Empty-reason exempt row produces a finding |
| AC-RNA-005 | REQ-RNA-006, REQ-RNA-007 | Must | Selector neutralisation runs before matching; word axis alive, live breadth set empty |
| AC-RNA-006 | REQ-RNA-008 | Must | Breadth set printed at column 0 + set equality |
| AC-RNA-007 | REQ-RNA-010 | Must | Control probe observed to fire |
| AC-RNA-008 | REQ-RNA-009 | Must | Empty breadth set fails |
| AC-RNA-009 | REQ-RNA-011 | Must | Newly-reached stale sites carry `CountPattern` + `KnownStale` |
| AC-RNA-010 | spec.md §C (ordering) | Must | Baseline commit precedes the implementation commit |
| AC-RNA-011 | spec.md §C (PRESERVE) | Must | t922 tests pass unchanged |
| AC-RNA-012 | REQ-RNA-004 | Must | Every hit accounted for; residual arithmetic closes |
| AC-RNA-013 | REQ-RNA-012 | Must | In-run, layer-independent re-derivation is the run baseline |
| AC-RNA-014 | spec.md §E (historical citations) | Must | Historical citations: no finding, no repair, reason if exempted |
| AC-RNA-015 | REQ-RNA-013 | Must | Layer scope = sweep exclusions + `.moai/research/` |

15 criteria, under the Tier M ceiling of 16.

### AC-RNA-001 — An undeclared count claim fails the guard

**Given** a file inside the layer's scope (REQ-RNA-013) carrying a numeral adjacent to a noun in
the D1 class, with no `Registry()` row and no numeral-exempt declaration for its path
**When** the numeral layer runs
**Then** that path enters the **finding set**, and the guard fails with a message naming the
observed numeral, the matched noun phrase, and the two ways to discharge it (register a
`ClaimCount` row, or declare a numeral exemption with a reason).

### AC-RNA-002 — Discharge follows the D2 rule, and only it

**Given** three registry states for one path — (a) a row with `Claims` including `ClaimCount`,
(b) no row but a numeral-exempt declaration with a non-empty reason, (c) a row carrying
`ClaimMembership` only
**When** the layer evaluates a numeral hit in that path
**Then** (a) and (b) discharge the hit — it stays in the breadth set and does not enter the
finding set — and (c) does NOT: (c) produces a finding stating that a membership registration
does not discharge a count claim.

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
discharged, and the test fails.

### AC-RNA-005 — Neutralisation runs before matching; the word axis is alive and its live breadth set is empty

**Given** the REQ-RNA-007 order (neutralise → match → discharge → report), the live phrases
`one of the 11 retained agents` (selector) and `All four are retained agents.` (subset
predication, at `.claude/agents/harness/workflow-specialist.md:52`), and a synthetic word-axis
input of the form `thirteen retained agents` that is NOT in the tree
**When** the layer runs against the live tree and against the synthetic input
**Then** all three hold:

(a) the synthetic input produces at least one word-axis match — proving the regexp fires, so an
empty live result is distinguishable from a dead regexp;

(b) neither live phrase reaches the match stage, so neither enters the breadth set and neither is
discharged by an exempt row; each shape is exercised by its own explicit test case rather than
being an accident of the window width;

(c) consequently the live word-axis **breadth set** is empty, and the run evidence names
`workflow-specialist.md:52` and the neutralisation that removed it — so the empty result is
pinned to an observed cause rather than an expectation.

(b) and (c) are the same fact stated from two directions, which is why they are one criterion:
split across two Must criteria they contradicted each other.

### AC-RNA-006 — The breadth set is printed at column 0, and equals the discharged union

**Given** the layer's **breadth set** and the independently derived population from AC-RNA-013
**When** the package test runs on a green tree (finding set empty)
**Then** both hold:

(a) each breadth-set member is printed on its own line beginning at column 0 with a stable prefix
(`digit-axis hit ` / `word-axis hit `), such that
`go test -count=1 ./internal/harness/rosterguard/... -v | grep -c '^digit-axis hit '` is
executable — a `t.Logf`-based implementation fails this grep because it indents and prefixes
with `file:line`; and

(b) the printed breadth set EQUALS `registered ClaimCount paths ∪ exempt paths` exactly — no more
and no less — and its size equals the AC-RNA-013 re-derived population. Both terms of the RHS are
registry-derived and therefore independent of the layer; a layer-produced term (for instance the
neutralised paths) must NOT appear, because an over-eager neutraliser would then shrink both
sides together and the equality would hide the failure it exists to expose. Neutralised paths are
in neither set (REQ-RNA-008) and so belong on neither side. The equality is the assertion the
reuse target already makes (`docsTabAllowlist` asserts the hit set equals the allowlist exactly);
comparing the printed count to the implementation's own count is circular and a mutant printing
only the first hit passes it.

### AC-RNA-007 — The control probe fires, and is observed firing

**Given** the control probe's deliberately-wrong input
**When** the probe runs
**Then** every probe case produces the violation it declares; and the probe's execution is
observed in the same run (`go test … -run <probe> -v` shows it RUN, not `no tests to run`),
followed by a full-package run with no `-run` filter.

### AC-RNA-008 — An empty breadth set is a measurement failure

**Given** a tree that demonstrably carries count claims (the registry is non-empty by
construction)
**When** the numeral layer produces an empty **breadth set**
**Then** the test FAILS with a message stating that an empty breadth set means the walk, the noun
class, the neutraliser, or the exclusion list is broken — not that the tree is clean.

### AC-RNA-009 — Newly-reached stale sites become enumerable, not exempt

**Given** the sites the layer newly reaches whose declared count disagrees with
`template.ProfileMatrixAgents()`
**When** run-phase completes
**Then** each carries a `Registry()` row with a `CountPattern` matching exactly once and a
`KnownStale` marker recording `DeclaredCount`, reason and follow-up; none is discharged by a
bare exemption; and the stale prose itself is unmodified (`git diff` empty on those content
lines).

### AC-RNA-010 — Baseline-first ordering is witnessed by the commit graph

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

### AC-RNA-011 — The t922 surface is preserved

**Given** `TestSweepFindsNoUndeclaredRosterListing`, `TestGuardFiresOnDeliberatelyWrongInput`,
`TestKnownStaleInventory` and the per-site checks
**When** the numeral layer lands
**Then** all pass without modification to their assertions, and `SweepUnreachable` semantics are
unchanged.

### AC-RNA-012 — Every hit is accounted for; the residual arithmetic closes

**Given** the in-run population of AC-RNA-013
**When** run-phase completes
**Then** the run evidence states the quantities — total breadth-set size, hits discharged by a
pre-existing `ClaimCount` row, hits newly registered, hits exempted — and their sum EQUALS the
breadth-set size, with the finding set empty. No hit is left implicit, and the number of authored
registry rows a reviewer must read is reported as an observed figure against the plan-phase
residual of 46 (spec.md §A).

### AC-RNA-013 — The run baseline is re-derived in-run, independently of the layer

**Given** REQ-RNA-012 and the tree being changed
**When** run-phase begins
**Then** the digit-axis and word-axis populations and the residual arithmetic are produced by a
command run in THIS tree in THIS run, and the run evidence cites that command with its verbatim
output. The derivation MUST be independent of the layer's own counter: it reads `Registry()` and
the tree directly rather than consulting the layer's reported totals — that independence is what
makes AC-RNA-006(b)'s size comparison non-circular, so it is stated here, where the command is
specified. Citing the plan-phase figures of spec.md §A as the baseline fails this criterion even
when the numbers happen to agree.

> Parse-shape note: a text-level `Path:`/`Claims:` parse of `registry.go` undercounts, because
> `readmeSite()` builds four rows whose `Path` is a parameter with no string literal beside it.
> The re-derivation reads `Registry()` at runtime for this reason.

### AC-RNA-014 — Historical citations: no finding, no repair, reason where exempted

**Given** the live historical-citation sites — `the then-8-agent catalog` in
`.claude/agents/moai/manager-docs.md`, `.claude/agents/moai/manager-spec.md` and their
`internal/template/templates/` mirrors, and `17->8 agent catalog` in
`internal/template/*_test.go`
**When** the layer runs, by whichever route run-phase chooses (tense/selector mechanism, or a
per-path exempt declaration)
**Then** none enters the finding set, none is repaired (`git diff` empty on those content lines),
and every one handled by exemption rather than by mechanism carries a non-empty per-path reason
a reviewer can disagree with. The mechanism choice stays deferred; this outcome does not.

### AC-RNA-015 — The layer's scope is stated and matches the sweep's

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
- A template mirror and its local counterpart: two paths, two rows, no collapsing (D3 withdrawn,
  spec.md §D — separate registration is what detects a one-sided repair).
- rosterguard's own files describing the guard in prose — a measured false-positive class that
  must be handled by mechanism or by a reasoned exempt row, never silently.

## §D.2 Quality Gates

- `go test -count=1 ./internal/harness/rosterguard/...` green
- `golangci-lint run` clean on touched packages; `gofmt` clean; `go vet` clean
- Coverage ≥85% on the new layer's code
- Conventional commits; `card t930` named in every commit message on this branch

## §D.3 Definition of Done

- AC-RNA-001 … AC-RNA-015 all Must-severity PASS with cited in-run evidence (command + verbatim
  output) recorded in `progress.md` §E.2
- Plan-phase population figures re-derived in-run and cited as the run baseline (AC-RNA-013)
- Residual arithmetic closed with no implicit hit (AC-RNA-012), and the observed authored-row
  count reported against the plan-phase residual of 46
- Registry expansion complete, every new row carrying its own `CountPattern` and `KnownStale`;
  mirror rows are authored separately, not derived (D3 withdrawn)
- Template-First pairs edited on both sides where touched

## §D.4 Traceability

REQ-RNA-001→AC-RNA-001 · REQ-RNA-002→AC-RNA-001 · REQ-RNA-003→AC-RNA-003 ·
REQ-RNA-004→AC-RNA-002, AC-RNA-012 · REQ-RNA-005→AC-RNA-004 · REQ-RNA-006→AC-RNA-005 ·
REQ-RNA-007→AC-RNA-005 · REQ-RNA-008→AC-RNA-006 · REQ-RNA-009→AC-RNA-008 ·
REQ-RNA-010→AC-RNA-007 · REQ-RNA-011→AC-RNA-009 · REQ-RNA-012→AC-RNA-013 ·
REQ-RNA-013→AC-RNA-015 · spec.md §C→AC-RNA-010, AC-RNA-011 ·
spec.md §E (historical citations)→AC-RNA-014

🗿 MoAI

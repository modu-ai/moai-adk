# Progress — SPEC-HEADING-REQ-COLLECT-001

Card t894. Tier M. Plan-phase base: `9dcbc3dbe` (branch `WT-heading-req-coverage`).

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-09-18
tier: M
artifacts: spec.md, plan.md, acceptance.md (+ this progress.md)
requirements: 11 (REQ-HRC-001..011) — Tier M ceiling 16
acceptance_criteria: 14 (AC-HRC-001..011, AC-HRC-GATE-001..003) — Tier M ceiling 16
volume_measurement: taken BEFORE the design decision, per the card's [HARD] constraint — probe `internal/spec/heading_req_volume_probe_test.go` at `9dcbc3dbe`; variant B adopted (651 findings vs variant A's 1500, with ModalityMalformed 8 → 36)
scope_note: axis 1 of card t801, split by the lead; axis 2 is SPEC-SIBLING-MAPS-SHORTHAND-001 (completed). No file is touched by both.

## §E.2 Run-phase Evidence

### Pre-flight (plan.md §C) — measured BEFORE the first source edit

Run-phase base `<base>`: **`dcfad4805`** (branch `WT-heading-req-coverage`, tree clean).
Measuring instrument built FROM this tree and invoked BY PATH:
`go build -o <scratch>/moai-base ./cmd/moai` (exit 0).

**Corpus census, re-measured at `dcfad4805`** — not quoted from `spec.md` §A.3:

| Quantity | Value at `dcfad4805` | Command |
|---|---:|---|
| documents swept (probe's own sweep) | 1635 | `go test -tags heading_req_probe ./internal/spec/ -run TestHeadingREQVolume -v -count=1` |
| files carrying a domain-qualified `### REQ-…` heading | 218 | `grep -rlE '^### \*{0,2}REQ-[A-Z0-9]+(-[A-Z0-9]+)*-[0-9]+' .moai/specs --include='*.md' \| wc -l` |
| entries newly collected (both variants) | 1031 | probe, above |
| files gaining at least one entry | 125 | probe, above |
| `.md` files under `.moai/specs` (wider denominator) | 3697 | `find .moai/specs -name '*.md' -type f \| wc -l` |

Gap: the "files carrying heading-form REQs" figure has **no producing line in the
probe** at any SHA — the probe logs `documents swept`, `files gaining at least one
REQ entry`, `newly collected REQ entries` and the per-code table, and nothing that
yields it. The 218 above is therefore this run's own grep measurement, named with
its command, not a reproduction of §A.3's 218 by the same instrument. That the two
agree is a coincidence of two different probes, not a confirmation.

**Probe projection re-measured at `dcfad4805`** (`spec.md` §A.3 was `9dcbc3dbe`; the
corpus moved — `DuplicateREQID` went 0 → 1 and both totals rose):

| Finding code | A (heading title) | B (body paragraph) |
|---|---:|---:|
| `ModalityUnjudged` | 1023 | 145 |
| `CoverageIncomplete` | 473 | 473 |
| `ModalityMalformed` | 8 | 36 |
| `DuplicateREQID` | 1 | 1 |
| `LegacyEARSKeyword` | 0 | 0 |
| `InvalidREQID` | 0 | 0 |
| **TOTAL** | **1505** | **655** |
| top-15 concentration | 28% | 44% |

Variant-B ratio at this SHA: `ModalityUnjudged` 145 ÷ 1031 entries = **0.141**
(variant A: 1023 ÷ 1031 = **0.992**). AC-HRC-008's 0.50 threshold still sits
between them.

**Whole-corpus BEFORE per-code counts**, `<scratch>/moai-base spec lint --json`
at `dcfad4805`, exit 0, 3381 findings total:

| Finding code | BEFORE |
|---|---:|
| `CoverageIncomplete` | 2033 |
| `ModalityUnjudged` | 523 |
| `ModalityMalformed` | 180 |
| `LegacyEARSKeyword` | 48 |
| `InvalidREQID` | 6 |
| `DuplicateREQID` | 0 |
| **six-code subtotal** | **2790** |

Refused-tool Gap (`verification-claim-integrity.md` §3.1): the first attempt to
capture this file used a shell variable for the binary path and was refused by the
worktree-isolation guard ("command whose name is computed at runtime"). It was
re-issued with the literal absolute path and executed; nothing was inferred in
place of a measurement.

### M1 — implementation

Base HEAD at M1 entry: `dcfad4805` (tree clean apart from this file). All evidence
below was captured in this run, against this tree, inside the worktree
`.claude/worktrees/t894` on branch `WT-heading-req-coverage`.

**Files changed (M1 scope):**

- NEW `internal/spec/lint_req_heading.go` — `reqHeadingWidePattern`,
  `firstParagraphBelowHeading`, `parseREQsHeadingForm`
- NEW `internal/spec/lint_req_heading_test.go` — 8 unit tests
- `internal/spec/lint.go` — `REQSourceHeading` constant + `String()` case
- `internal/spec/lint_req_widen.go` — third source composed into the merge, plus
  the REQ-HRC-011 comment correction
- `internal/spec/lint_req_table.go` — REQ-HRC-011 comment correction only

**Naming note.** The live symbols are `parseREQsHeadingForm` /
`reqHeadingWidePattern` / `firstParagraphBelowHeading` rather than the probe's
`parseREQsHeading` / `headingREQPattern` / `firstParagraphAfter`. The probe lives
in the same package under a build tag, so reusing its names would make
`-tags heading_req_probe` fail to compile with duplicate declarations.

**TDD RED (E8).** Two RED observations were captured before any collector logic
existed. RED-1, `go test ./internal/spec/ -count=1 -run TestHeadingCollection`:

```
internal/spec/lint_req_heading_test.go:64:13: undefined: parseREQsHeadingForm
...
internal/spec/lint_req_heading_test.go:148:18: undefined: REQSourceHeading
FAIL	github.com/modu-ai/moai-adk/internal/spec [build failed]
```

RED-2, after adding only the `REQSourceHeading` constant and a stub collector
returning `nil`, so the assertions actually execute:

```
--- FAIL: TestHeadingCollection_CollectsLevel3Definitions (0.00s)
    lint_req_heading_test.go:66: collected 0 entries, want 2: []
--- FAIL: TestHeadingCollection_TextIsBodyParagraphNotHeadingTitle (0.00s)
--- FAIL: TestHeadingCollection_FallsBackToHeadingTextNeverEmpty (0.00s)
--- FAIL: TestHeadingCollection_SearchStopsAtNextHeadingOfAnyLevel (0.00s)
--- FAIL: TestHeadingCollection_EveryEntryIsWidened (0.00s)
--- FAIL: TestHeadingCollection_MergedInDocumentOrder (0.00s)
    lint_req_heading_test.go:178: collected [REQ-FIXH-020 REQ-FIXH-022], want [REQ-FIXH-020 REQ-FIXH-021 REQ-FIXH-022]
--- FAIL: TestHeadingCollection_SourceDoesNotDecideSeverity (0.00s)
```

One test, `TestHeadingCollection_IgnoresOtherHeadingLevels`, passed against the
`nil` stub — a negative criterion cannot go red on an empty collector. Its
meaningful run is the GREEN one, against the real pattern.

**GREEN.** `go test ./internal/spec/ -count=1 -v -run TestHeadingCollection` —
all 8 PASS (`ok github.com/modu-ai/moai-adk/internal/spec 0.274s`).

**M1-owned AC matrix.**

| AC | Status | Deciding command | Observed |
|---|---|---|---|
| AC-HRC-002 (REQ-HRC-004) | PASS | `go test -run TestHeadingCollection_TextIsBodyParagraphNotHeadingTitle` | PASS — `Text` equals the `**When** …` paragraph AND is asserted not equal to the heading title; both directions present |
| AC-HRC-003 (REQ-HRC-005) | PASS | `go test -run TestHeadingCollection_FallsBackToHeadingTextNeverEmpty` | PASS — entry produced, `Text` = `Ubiquitous — a title with no body`, non-empty |
| AC-HRC-004 (REQ-HRC-006) | PASS | `go test -run TestHeadingCollection_SearchStopsAtNextHeadingOfAnyLevel` | PASS — boundary is `##` (different level from the producing `###`); `Text` is the fallback and carries no substring of the following section |
| AC-HRC-006 (REQ-HRC-008) | PASS | `go test -run TestHeadingCollection_SourceDoesNotDecideSeverity` + `grep -n reqFindingSeverity internal/spec/lint.go` | PASS — severity distribution byte-identical when every entry is rotated through all three `Source` values; grep shows the call sites take `(REQEntry, Severity)` only, the `Source`-adjacent lines being the prohibition comments at 564/566/589 |
| AC-HRC-011 (REQ-HRC-011) | PASS | `grep -n 'four' internal/spec/lint_req_widen.go internal/spec/lint_req_table.go` | PASS — 4 hits, none claiming four findings/rules: two are this card's own quotation of the corrected text, one is `four properties` (L1 lexicon), one is `four-segment` (ID shape) |
| REQ-HRC-001 / 002 / 003 (unit level) | PASS | `go test -run TestHeadingCollection_CollectsLevel3Definitions`, `_IgnoresOtherHeadingLevels`, `_MergedInDocumentOrder` | PASS — 2 entries in document order; `##`/`####`/`#####` collect 0; merged list+heading order `020, 021, 022` with the list entry's `Text`/`Source` unperturbed |
| REQ-HRC-007 (unit level) | PASS | `go test -run TestHeadingCollection_EveryEntryIsWidened` | PASS — every entry `Widened: true`, `Source: REQSourceHeading` |

**REQ-HRC-011 corrected text.** `lint_req_widen.go` now reads: "doc.REQs feeds
SIX findings today — ModalityMalformed, ModalityUnjudged, LegacyEARSKeyword,
InvalidREQID, DuplicateREQID, CoverageIncomplete — emitted by three loops
(EARSModalityRule, REQIDUniquenessRule, CoverageRule)." `lint_req_table.go` now
reads: "The six finding codes that consume doc.REQs — ModalityMalformed,
ModalityUnjudged, LegacyEARSKeyword, InvalidREQID, DuplicateREQID,
CoverageIncomplete, emitted by three rules (EARSModalityRule,
REQIDUniquenessRule, CoverageRule)." The six were verified against the emission
sites before writing (`lint.go` EARSModalityRule ~838-888 emits three,
REQIDUniquenessRule ~1053-1065 emits two, CoverageRule ~1115 emits one) rather
than copied from the dispatch; code and dispatch agree.

**Quality gate.**

| Check | Command | Result |
|---|---|---|
| Package suite | `go test ./internal/spec/... -count=1` | `ok github.com/modu-ai/moai-adk/internal/spec 77.427s` |
| Coverage | `go test -cover ./internal/spec/... -count=1` | `coverage: 90.2% of statements` (≥ 85) |
| Vet | `go vet ./internal/spec/...` | exit 0, no output |
| Lint | `golangci-lint run --timeout=5m ./internal/spec/...` | `0 issues.` exit 0 |
| Build (host) | `go build ./...` | exit 0 |
| Build (windows) | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| Subagent boundary | `grep -rn AskUserQuestion internal/spec` | no output |
| Probe | `go test -tags heading_req_probe ./internal/spec/ -run TestHeadingREQVolume -count=1` | **FAIL** — see the blocker below |

**BLOCKER — AC-HRC-GATE-001's probe clause is unsatisfiable as written.** The
probe computes `fresh` as the heading entries the LIVE collector does not already
carry, and guards on it:

```
if v.newREQs == 0 {
    t.Fatalf("probe found zero heading-form REQs the live collector misses — either the corpus changed or the pattern is wrong")
}
```

Once M1 lands, that set is empty by construction, so the probe fatals on both
variants:

```
--- FAIL: TestHeadingREQVolume (1.95s)
    --- FAIL: TestHeadingREQVolume/A_heading_title_as_text (1.27s)
        heading_req_volume_probe_test.go:227: probe found zero heading-form REQs the live collector misses — either the corpus changed or the pattern is wrong
    --- FAIL: TestHeadingREQVolume/B_body_paragraph_as_text (0.68s)
```

The probe still BUILDS under its tag (the failure is a runtime `t.Fatalf`, not a
compile error), so the naming-collision half of the clause is satisfied; the
`exits 0` half is not, and cannot be without editing the instrument. This also
bears on AC-HRC-009, which asks for the probe's variant-B projections at the SAME
SHA as the M1 build — a differential instrument produces no projections once the
differential is zero. Resolving this requires a decision on the instrument or on
the two criteria, both of which are SPEC body content outside run-phase
ownership, so neither `heading_req_volume_probe_test.go` nor `acceptance.md` was
touched.

**Probe-vs-implementation divergence (named cause for the M2 reconciliation).**
The probe's `firstParagraphAfter` returns only the FIRST non-empty line below the
heading; this collector returns the whole run of consecutive non-empty lines
joined with single spaces, per plan.md §B ("the run of consecutive non-empty
lines beginning at the first non-empty line inside those bounds"). The plan is
binding on the run phase, so the implementation follows it. The two therefore
differ on multi-line paragraphs, and the per-code actuals are expected to diverge
from the probe's projections in that direction — more text reaching the judge
means some entries that the probe scored `ModalityUnjudged` will instead be
judged. This is a named cause, recorded here before M2 measures it.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

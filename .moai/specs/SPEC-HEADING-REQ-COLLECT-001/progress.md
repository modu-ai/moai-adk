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

### M1a — the instrument repair (lead ruling: option A, 2026-09-18)

The blocker above is resolved by repairing the INSTRUMENT, not the criteria.
`acceptance.md` is not opened: what broke is the probe, so the probe is where the
repair belongs. Lead ruling, recorded verbatim in the commit message of this
change.

**The change.** `heading_req_volume_probe_test.go` built its `existing` baseline
from `doc.REQs` — the live collector. Those two were the same set until the
heading source landed in the live collector; afterwards a `doc.REQs` baseline
makes `fresh` empty by construction and the probe measures its own subject. The
baseline is now the two pre-heading sources named directly:

```go
base := mergeREQsByLine(parseREQsWide(doc.Body), parseREQsTable(doc.Body))
```

**Why this is not "editing the instrument until it passes" — the control.** At
`dcfad4805` the live collector WAS exactly list+table, so the repaired probe must
reproduce that tree's pre-flight measurement (§E.2 above) EXACTLY. It does. Run at
`c461577eb` (the M1 build), `go test -tags heading_req_probe ./internal/spec/
-run TestHeadingREQVolume -v -count=1`, exit 0, `--- PASS`:

| Quantity | §E.2 pre-flight at `dcfad4805` | repaired probe at `c461577eb` |
|---|---:|---:|
| documents swept | 1635 | 1635 |
| files gaining at least one entry | 125 | 125 |
| newly collected entries | 1031 | 1031 |
| A — TOTAL / `ModalityUnjudged` | 1505 / 1023 | 1505 / 1023 |
| A — `CoverageIncomplete` / `ModalityMalformed` / `DuplicateREQID` | 473 / 8 / 1 | 473 / 8 / 1 |
| B — TOTAL / `ModalityUnjudged` | 655 / 145 | 655 / 145 |
| B — `CoverageIncomplete` / `ModalityMalformed` / `DuplicateREQID` | 473 / 36 / 1 | 473 / 36 / 1 |

Every row matches. The lead's condition was explicit: any one of the four
headline figures diverging means either the repair is wrong or the live collector
changed something beyond list+table, and those need different responses — "close
enough" would have destroyed the control this ruling rests on.

**The third fixed point.** The corpus is not byte-identical between `dcfad4805`
and `c461577eb` — this SPEC's own `spec.md` and `progress.md` changed in that
range (`git diff --name-only dcfad4805..HEAD -- .moai/specs`, 2 files; control:
7 files changed in the full range, so the range is non-empty). Their heading-form
contribution is what matters, and it is unchanged on both sides:
`grep -cE '^### \*{0,2}REQ-'` returns 2 for `spec.md` at `dcfad4805` and 2 at
`c461577eb`, 0 for `progress.md`. A difference in the table above would therefore
have been attributable to the probe edit alone.

**Consequence for the two criteria.** Both are now satisfiable AS WRITTEN, with no
amendment: AC-HRC-GATE-001's probe clause (`builds and exits 0`) passes at the M1
SHA, and AC-HRC-009's variant-B projections exist at the same SHA as the M1 build.

**Residual risk, named not fixed.** `spec.md`'s two heading-form matches are inside
FENCED CODE BLOCKS (the §A.1 and §A.2 illustrations). The heading collector does
not strip fences, so it collects them as definitions. This is a pre-existing
property of the collector family — `reqLineWidePattern` has the same blindness for
list-form definitions inside fences — and it is identical on both sides of this
measurement, so it does not affect the control above. It is NOT repaired here
(out of scope: this SPEC collects a heading shape, it does not add fence
awareness to the collector family) and is carried into M2's reconciliation as a
named cause for any projected-vs-actual difference on documents that illustrate
REQ syntax in code blocks.

### M2 — evidence (no source change)

Tree measured: **`4ff316bd8`**, `git status --short` empty at entry and exit. M2
changes no file under `internal/`; it writes `.moai/reports/t894/**` and this
section. Full evidence report, 5-section format:
`.moai/reports/t894/verdict.md`, **exported to the primary checkout** at
`/Users/goos/MoAI/moai-adk-go/.moai/reports/t894/` before being cited here
(`.moai/reports/*` is gitignored at `.gitignore:229`, so the in-worktree copy would
not resolve at audit time).

Instruments, each built FROM a named tree and invoked BY PATH (no PATH-resolved
`moai` decides any criterion here): `moai-base` (`dcfad4805`), `moai-m1`
(`4ff316bd8`, `go build` exit 0), `moai-mutant` (`4ff316bd8` + the one-line merge
removal, `go build` exit 0).

**M2-owned AC matrix.**

| AC | Status | Deciding command | Observed |
|---|---|---|---|
| AC-HRC-001 | PASS | `moai-m1 spec lint <fix>/SPEC-FIXH-001`; same with `moai-base` | GREEN: `CoverageIncomplete` names `REQ-FIXH-001`, exit 0, 11 warnings. RED-now at `<base>`: no line names it, exit 0, 10 warnings. The 10 `FrontmatterInvalid` witness lines (an unrelated rule) are present in BOTH runs |
| AC-HRC-005 | PASS | `moai-m1 spec lint <fix>/SPEC-FIXI-001` | `WARNING ModalityMalformed`, summary `0 error(s), 12 warning(s)`, **exit 0** — severity, error count and exit status all asserted. Control `SPEC-FIXICTL-001` (same statement, narrow list form, `Widened=false`) → `ERROR`, `1 error(s)`, **exit 1**, proving the demotion is the `Widened` axis and not `applyEraDemotion` |
| AC-HRC-007 | PASS | build-tagged dump (`//go:build reqdump`, **scratch trees only** — this tree gained no file), both runs over the SAME corpus | BEFORE 4497 entries, AFTER 5541, **removed-or-mutated = 0**, added = 1044, all `Source=heading` and `Widened=true`. Records carry `file\|line\|ID\|Source\|Widened\|Text`, so a mutation of any field would surface as a removal |
| AC-HRC-008 | PASS | `<build>/moai spec lint --json`, both sides, exit 0 | See the table below. Ratio **0.043** (< 0.50); total delta **+496** (> 0); **no negative delta** |
| AC-HRC-009 | PASS | probe at `4ff316bd8` (exit 0, PASS) + a scratch-only `reqdiag` cross-tab | Every projected-vs-actual difference reconciled to a measured cause; the diagnostic reproduces all six actual deltas **exactly** |
| AC-HRC-010 | PASS (RED as required) | `moai-mutant spec lint <fix>/SPEC-FIXH-001` | `CoverageIncomplete` naming `REQ-FIXH-001` **disappears**; all 10 witness lines **remain**. `diff -rq` shows exactly ONE differing file (`lint_req_widen.go`); the collector and its pattern are byte-unchanged. Mutant never committed |
| AC-HRC-GATE-001 | PASS | `go test ./internal/spec/... -count=1`; `-v -run TestHeadingCollection`; probe under its tag | exit 0 `ok … 91.158s`; all 8 `TestHeadingCollection_*` named and PASS; probe builds and **exits 0** |
| AC-HRC-GATE-002 | PASS | `go vet ./internal/spec/...` | exit 0, no output |
| AC-HRC-GATE-003 | PASS | `golangci-lint run --timeout=5m ./internal/spec/...`, HEAD and `<base>` | `0 issues.` at both — baseline measured in THIS run, nothing new |

**Whole-corpus BEFORE/AFTER (AC-HRC-008).** The CONTROL column is an addition beyond
the criterion: BEFORE and AFTER differ in binary AND tree, so the `<base>` binary was
also run against the HEAD tree to isolate the instrument.

| Finding code | BEFORE `moai-base`@`dcfad4805` | CONTROL `moai-base`@`4ff316bd8` | AFTER `moai-m1`@`4ff316bd8` | Δ |
|---|---:|---:|---:|---:|
| `CoverageIncomplete` | 2033 | 2033 | 2470 | +437 |
| `ModalityUnjudged` | 523 | 523 | 567 | +44 |
| `ModalityMalformed` | 180 | 180 | 181 | +1 |
| `LegacyEARSKeyword` | 48 | 48 | 48 | 0 |
| `InvalidREQID` | 6 | 6 | 6 | 0 |
| `DuplicateREQID` | 0 | 0 | 14 | +14 |
| **six-code subtotal** | **2790** | **2790** | **3286** | **+496** |

The CONTROL reproduces the `dcfad4805` counts **byte-for-byte on every code** (3381
findings total on both). Corpus movement between the two trees therefore contributes
**zero**, and the whole +496 is attributable to the collector change. BEFORE was
re-derived from `corpus-before.json` by `jq`, reproducing the hand-carried figures.

Ratio: 44 ÷ 1031 = **0.0427** (probe `fresh` denominator); 44 ÷ 1044 = 0.0421 (live,
all `.md`); 44 ÷ 1007 = 0.0437 (live, linter scope). Every denominator lands near
variant B's 0.14 and nowhere near variant A's 0.99.

**Corpus census re-measured at `4ff316bd8`** (NOT quoted from `spec.md` §A.3):
documents swept **1635**; files carrying a domain-qualified `### REQ-…` heading
**218** (`grep -rlE '^### \*{0,2}REQ-[A-Z0-9]+(-[A-Z0-9]+)*-[0-9]+' .moai/specs
--include='*.md' | wc -l` — the probe still emits no line producing this figure, so
it is this run's own grep, named with its command); entries newly collected **1031**
(probe `fresh`) / **1044** (live, all `.md`) / **1007** (live, linter scope); files
gaining an entry **125**; `.md` under `.moai/specs` **3697**; documents the linter's
REQ rules actually visit **874**.

**Reconciliation (AC-HRC-009) — three measured causes.**

| Code | Projected (probe B) | Actual Δ | Diff | Cause |
|---|---:|---:|---:|---|
| `CoverageIncomplete` | 473 | 437 | −36 | C1 scope (−37), C3 fresh-filter (+1) |
| `ModalityUnjudged` | 145 | 44 | −101 | C2 extractor (−65), C1 (−37), C3 (+1) |
| `ModalityMalformed` | 36 | 1 | −35 | C2 extractor |
| `DuplicateREQID` | 1 | 14 | +13 | C3 fresh-filter |
| `LegacyEARSKeyword` / `InvalidREQID` | 0 / 0 | 0 / 0 | 0 | — |
| **TOTAL** | **655** | **496** | **−159** | the three causes |

- **C1 — scope (dominant, and NOT anticipated in the M1 notes).** `discoverSPECs`
  (`lint.go:396`) globs `SPEC-*/spec.md` — one file per SPEC directory, one level
  deep — so the REQ-consuming rules visit **874** documents where the probe walks
  **1635**, and see **1007** heading entries where the collector produces 1044. The
  excess is `acceptance.md` / `plan.md` / nested / sample documents the rules never
  visit.
- **C2 — extractor.** Probe first-line vs shipped whole-paragraph-joined; the texts
  differ on **306** of 1044 entries (275 in scope). This is the cause named in advance
  in M1 — confirmed, **but the direction it predicted for `ModalityMalformed` was
  wrong**: M1 expected actual "possibly above" projection, and it is far below
  (36 → 1), because the full paragraph supplies the `SHALL` the truncated first line
  cut off, so probe-malformed entries are judged conforming. The direction predicted
  for `ModalityUnjudged` (below projection) was correct.
- **C3 — fresh-filter.** The probe drops heading entries whose ID already exists in
  list+table, so it can only count duplicates within its own fresh set. Measured:
  exactly **13** added entries collide with a pre-existing `(file, ID)` — precisely
  `1044 − 1031`, and precisely the `+13` `DuplicateREQID` discrepancy.
- The **fence-blindness** cause named in M1a explains **none** of the discrepancy:
  probe and collector are equally fence-blind, so it cancels on both sides. It remains
  a standing, unrepaired property (Residual-risk).

Applying the three causes, the scratch-only `reqdiag` cross-tab reproduces every
actual delta exactly: `ModalityUnjudged` 44, `ModalityMalformed` 1,
`CoverageIncomplete` 437, `DuplicateREQID` 13+1 = 14, `LegacyEARSKeyword` 0,
`InvalidREQID` 0 — total **496 = +496**. Nothing in the difference is unattributed.

**Gaps (M2).** `<scratch>` raw captures are machine-local and NOT exported — a known
loss, named rather than cited. `moai-base`'s provenance is hand-carried, not
re-verified by byte-identity. Two tool calls were refused by the worktree-isolation
guard (a compound `git archive` loop; a heredoc writing a Go file) and were re-issued
as plain commands / via the Write tool and **executed** — no fallback reading
substitutes for a measurement. My first control fixture was mis-designed (bold +
em-dash form, which `reqLinePattern` does not match, so it was not a control) and was
corrected after reading the pattern; the corrected run is the one quoted.
AC-HRC-002/003/004/006/011 were NOT re-measured in M2 — they are M1-owned and carried
forward on M1's evidence. No cross-platform build, coverage, or full-suite run was
re-observed in M2 (no source changed); `go test ./...` was deliberately not run.

### M2 addenda (lead-directed, second round) — measured at `4ff316bd8`

**A1 — AC-HRC-006's grep clause is a pre-existing wrong-reason red: Gap, not a pass.**
`grep -n '[S]ource' internal/spec/lint.go | grep 'reqFindingSeverity'` returns **one
row on both trees**, byte-identical, line 564 — and that line IS the prohibition
comment:

```
564:	// [HARD] Source MUST NOT reach reqFindingSeverity, or any other severity
```

Same single row at `dcfad4805` (archive extraction). The clause was unsatisfiable
before this card existed and M1 changed nothing about it — red at arrival, red after,
caused by content the work never touches (`verification-completeness.md` §2,
wrong-reason-red). The AC's own text calls the grep "a tripwire, not enforcement" and
names the flip test as load-bearing; M1 discharged that half
(`TestHeadingCollection_SourceDoesNotDecideSeverity`). **Load-bearing half PASS, grep
clause a Gap.** Repair needs a pattern excluding comment lines = an `acceptance.md`
amendment, NOT this card's to make → follow-up-card candidate.

*Correction to M1's figure, precisely.* This section previously wrote "the prohibition
comments at 564/566/589". That is **correct for the grep M1 ran** (`grep -n
reqFindingSeverity internal/spec/lint.go` — 8 rows at base, 9 at HEAD, M1 having added
589) and **wrong if read as the AC's grep**, which yields one row: 566 and 589 name
`reqFindingSeverity` but carry no `Source` token on the same line, so the pipe excludes
them.

**A2 — GATE-003 baseline.** Already measured in round 1 and quoted in the M2 gate
table: `0 issues.` at HEAD and `0 issues.` at `<base>`. Stated plainly as directed: the
base reports **0**, so NEW cannot exceed 0 — the criterion is met by a **zero floor**,
which is a different fact from "a change that could have introduced issues introduced
none". Both hold here; only the first is what was measured.

**A3 — guard refusal worth recording** (`verification-claim-integrity.md` §3.1). The
worktree guard refuses `grep 'Source' …` because the literal pattern spells the shell
builtin `source` — nothing to do with git. `[S]ource` passes and produced A1. Third and
fourth refusal classes this run (others: compound `git archive` loop; heredocs feeding
`python3`/`cat`); every one was re-issued in an accepted form and **executed**.

**A4 — M1's residual risk 3 measured at corpus scale: contributes ZERO.** The risk was
that joining a multi-line paragraph merges a bullet / table row / fence opener into the
judged statement. `ModalityMalformed` runs *below* projection, which is precisely where
a **false negative** would hide, so it was measured rather than dismissed:

```
--- SILENCED (probe malformed -> shipped not malformed): 35
---   of which the joined TAIL supplies a SHALL: 35
---   of which the joined TAIL is NON-PROSE (fence/table/bullet): 0
```

35 = exactly the 36 → 1 gap; **none** has a non-prose tail. Verified against the texts,
not just my shape classifier — e.g. `REQ-CDPG-001`, probe first line `While
configPathSeparator is pinned to '\\', when codexStaleSkillFinding judges a config`,
joined tail `entry declaring an absolute slash-form path, the test suite shall observe
…`. These are **one prose sentence wrapped across source lines** with the `shall` on
the continuation line: the joining **recovers** the requirement, it does not
contaminate it. (`REQEntry.Line` is 1-based into `doc.Body` with frontmatter stripped,
so it is NOT the file line — a first spot-check that `sed`'d it as one landed on the
wrong paragraph and was discarded.)

**A4a — a consequence A4 exposes: one clause of the shipped rationale is unsupported.**
`internal/spec/lint_req_heading.go`'s header justifies variant B partly by "raises
ModalityMalformed 8 → 36 — it finds MORE real defects while producing LESS noise". The
`8 → 36` are the PROBE's figures under its first-line extractor; the **shipped**
whole-paragraph extractor yields a real-linter delta of **+1**, because 35 of that 36
are truncation artifacts (A4). The "finds MORE real defects" half is therefore **not
reproduced by the collector that shipped**. This does NOT change the variant decision —
B is better supported than the comment claims (noise ratio 0.043 vs A's 0.99), and
declining to report 35 truncation artifacts is accuracy, not suppression — but the
comment's stated evidence is wrong at the shipped extractor, and a reader re-deriving
the decision from that sentence would be misled. Repair edits a source file, which M2
must not do → **flagged as an M3 / follow-up-card candidate** alongside A1.

### M2 addendum A5 — `DuplicateREQID` adjudicated (Residual-risk ④ closed)

**The adjudication is the LEAD's** — the lead opened all 14 cases; I did not, and do
not record that reading as mine. **Verdict: 14/14 are structural false positives;
"two different requirements sharing one ID" occurs 0 times.** Three shapes:

| Shape | n | Where |
|---|---:|---|
| heading is a SECTION TITLE for a list-defined REQ | 10 | `SPEC-GOAL-HTML-FLOW-001/spec.md` — `:57` `- **REQ-GHF-001** — <full statement>` (definition) vs `:93` `### REQ-GHF-001 — Dashboard renderer substrate` (its section heading) |
| heading is the DEFINITION, the list row a DISPOSITION | 3 | `SPEC-INTERNAL-TEST-004/spec.md` — `:40` `### REQ-GOLD-001 — 6 golden tests PASS (Ubiquitous)` vs `:125` `- **REQ-GOLD-001** (…): **MET** — via ce2a509dc.` |
| fenced-code illustration | 1 | this SPEC's `spec.md:69`/`:87`, the same `### REQ-ADV-001` line quoted twice |

**Independently re-measured by me**: the file distribution reproduces exactly from my
own `dump-added.txt` (10 + 3 collisions against pre-existing entries, plus 1
within-added repeat = 14, matching `dupAgainstExisting=13 + dupWithinFresh=1`); every
cited line reads as described and these are FILE lines that resolve correctly (checked
deliberately — the body-vs-file axis already misled me once this run, A4); and the
fence case is structural (fences at 68/70 and 86/90, the two heading lines at 69/87).

**Stated without softening**: these are **structural false positives created by the
collection method**, NOT pre-existing corpus state merely revealed. A definition list
plus a per-requirement section is normal, correct document structure, and heading
collection turns it into a duplicate. Describing the whole +496 as "existing corpus
state made visible" would be an overclaim and is not claimed here.

**Equally, not a failure of this card**: every affected code is advisory and nothing
gates, AC-HRC-008's two discriminators hold unchanged, and `spec.md` §D puts corpus
repair out of scope.

**Diagnosis → follow-up card (bundled with the fence blind spot)**: the heading axis
has **no section-title-vs-definition discriminator**. `SPEC-INTERNAL-TEST-004` is the
sharper evidence — there the heading is the definition and the list row the
disposition, so both directions occur in one corpus and a rule assuming either would
be wrong half the time. **Card t518 met this on the table axis** ("definition vs
disposition table") and resolved it with an L1 vocabulary discriminator; that is the
precedent, and the heading axis has no equivalent.

**Limit of the adjudication**: the lead examined the **14 `DuplicateREQID` cases
only**. How much of the same cause is mixed into `CoverageIncomplete` **+437** or
`ModalityUnjudged` **+44** was **not measured** by either of us. A 100% false-positive
rate in 14 cases says nothing about the remaining 482.

## §E.3 Run-phase Audit-Ready Signal

run_status: complete
run_complete_at: 2026-09-18
run_measurement_tree: 4ff316bd8
run_commit_range: dcfad4805 (base) → c461577eb (M1) → 4ff316bd8 (M1a) → this section
evidence_path: .moai/reports/t894/verdict.md
evidence_exported_to: /Users/goos/MoAI/moai-adk-go/.moai/reports/t894/ (primary checkout; `.moai/reports/*` is gitignored at `.gitignore:229`, so the in-worktree copy would not resolve at audit time — `cmp` byte-identical)
ac_pass_count: 13
ac_fail_count: 0
ac_gap_count: 1

**Run-phase is complete.** All acceptance criteria are discharged; none FAILED.

| AC | Owner | Status |
|---|---|---|
| AC-HRC-001 | M2 | PASS |
| AC-HRC-002 / 003 / 004 | M1 | PASS |
| AC-HRC-005 | M2 | PASS |
| AC-HRC-006 | M1 (flip test) / M2 (grep clause) | **PASS on the load-bearing half; grep clause a GAP** — see below |
| AC-HRC-007 / 008 / 009 / 010 | M2 | PASS |
| AC-HRC-011 | M1 | PASS |
| AC-HRC-GATE-001 / 002 / 003 | M2 | PASS |

**AC-HRC-006 disposition (lead-approved).** Its grep clause is a **pre-existing
wrong-reason red** (`verification-completeness.md` §2): `grep -n '[S]ource'
internal/spec/lint.go | grep 'reqFindingSeverity'` returns **one row on both trees**,
byte-identical, line 564 — and that line IS the prohibition comment. Red at arrival,
red after, caused by content this work never touches. Recorded as a **Gap**, not a
pass and not a failure of this card. The criterion's load-bearing half — the
`Source`-rotation severity-distribution test, which the AC's own text names as
load-bearing while calling the grep "a tripwire, not enforcement" — is **PASS**.

**Headline measurements** (tree `4ff316bd8`, instruments built FROM named trees and
invoked BY PATH): six-code corpus delta **+496** (2790 → 3286), no negative delta;
variant discriminator **0.043** (< 0.50); no-regression **removed-or-mutated = 0**
over 4497 → 5541 entries; mutant probe RED exactly where required, witness lines
intact; suite / probe / vet / lint all clean, lint baseline `0 issues.` at both HEAD
and `<base>`.

**Follow-up card candidates (4) — none is this card's to make.**

1. **Fence blind spot.** The collector family does not strip fenced code blocks, so a
   document illustrating REQ syntax contributes real entries and real findings. This
   SPEC's own `spec.md` is such a document. Pre-existing (`reqLineWidePattern` shares
   the blindness) and explicitly out of scope here.
2. **AC-HRC-006 grep-clause amendment.** Needs a pattern that excludes comment lines.
   This is an `acceptance.md` edit — SPEC body content, outside run-phase ownership.
3. **Heading-axis section-title-vs-definition discriminator (absent).** Root cause of
   the 14 adjudicated `DuplicateREQID` false positives; both directions occur in the
   corpus, so no one-sided assumption is safe. Precedent: card t518's L1 vocabulary
   discriminator on the table axis.
4. **Probe scope over-reporting.** The probe walks all 1635 `.md` under `.moai/specs`
   where the REQ-consuming rules visit only the 874 `SPEC-*/spec.md` that
   `discoverSPECs` globs, so any decision taken from a raw probe projection is ~37
   findings optimistic on this corpus unless it is scoped first. Measured as cause C1
   of the AC-HRC-009 reconciliation. The probe is an instrument, and amending it is
   outside run-phase ownership.

**A4a is NOT a follow-up candidate — it was repaired inside this card.** Per lead
adjudication ("correct it within this card, do not defer"), the `lint_req_heading.go`
header comment was re-attributed by `manager-develop` (m1) in commit **`5f5cc103d`**
("attribute the variant-B figures to their instruments", 1 file, +41/−17). The comment
now separates the PLAN-PHASE PROJECTION (the probe's `8 → 36` first-line figures, kept
as the actual decision inputs) from the SHIPPED EXTRACTOR measurement (delta +1, ratio
0.043), explains that 35 of the probe's 36 are truncation artifacts, and carries a
`[HARD] DO NOT RESTORE THE CLAIM THAT B "FINDS MORE REAL DEFECTS"` regression guard.
The variant decision is unchanged and is now better supported than the retired text
claimed.

Verified independently by me on the landed commit: the diff carries **zero non-comment
changed lines**, and the Go tree is byte-identical across the HEAD move that happened
during m1's work (`git diff --name-only faf3dcfbc 6d1fbf5fc` → `progress.md` only).
`go build` / `go vet` were re-measured by the lead; m1's own `go test
./internal/spec/... -count=1` (`ok … 87.589s`) is attributed to m1 and was not re-run
by the lead or by me.

**Process record — a second writer briefly existed in this worktree during run-phase.**
Recorded per `agent-common-protocol.md` § Background Agent Execution, which requires an
unexpected write on an actively-worked tree to be reported and recorded rather than
quietly absorbed. Facts only:

- **Cause: lane-orchestrator dispatch ordering.** The lead re-delegated the A4a comment
  repair to `manager-develop` (m1) while this agent (m2) was still executing the lead's
  previous instruction, putting two write-capable agents in one worktree at once. The
  cause was the dispatch order, **not** either agent writing out of turn.
- **Detected and stopped, both ends.** m2 observed `internal/spec/lint_req_heading.go`
  modified immediately after its own commit, did not revert it (work of unknown
  provenance), did not sweep it into its own commit (explicit pathspec), stopped further
  writes, and reported to the lead before recording anything. Independently, m1 noticed
  that HEAD had moved off its dispatched base `faf3dcfbc` to `6d1fbf5fc` and, instead of
  assuming, established with `git diff --name-only faf3dcfbc HEAD` that only
  `progress.md` differed — so the Go tree was byte-identical across the move and its own
  evidence remained attributable.
- **No loss.** The two agents' files did not overlap and neither commit absorbed the
  other's work; m1's edit landed as its own commit `5f5cc103d`. That the overlap was
  harmless was a property of this particular pairing, not of the process.

**Not touched** (sync-phase fields): `§E.4`, `sync_commit_sha`, and every frontmatter
field other than those M1 already set. Run-phase changed no SPEC body content.

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

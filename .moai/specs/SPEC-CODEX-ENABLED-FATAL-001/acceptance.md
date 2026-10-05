# SPEC-CODEX-ENABLED-FATAL-001 — Acceptance Criteria

16 criteria against the Tier M ceiling of 16. The measured-fixture criteria (AC-CEF-001..007) map
one-to-one onto the six rows of the acceptance lab (`.moai/reports/t508/codex-enabled-lab.md`), so
the guard's fixture set and the lab's fixture set are the same set.

**Document-level tree pin.** Every RED-now / baseline cell in this file was measured in this tree at
commit **`069795602`** — the HEAD after the branch absorbed develop. A cell citing no SHA is not a
baseline; a cell citing a different SHA is a carry-over (verification-claim-integrity §2).

**Control-row obligation.** Every criterion of the "a fatal shape is detected" family is paired with
an accepting-shape control in the same run. An absence-guard that also passes when the feature is
entirely missing asserts nothing; the control is what makes a mutant's failure informative.
AC-CEF-005, 006, 007 and 009 are those controls and are mandatory in every run of this suite.

**Fixture isolation.** Every fixture is created under `t.TempDir()` with `CODEX_HOME` pinned per
fixture. No criterion reads or writes the machine's real `~/.codex`.

---

## §D.0 The swept-count obligation (binds EVERY `-run`-based criterion)

[HARD] A Go `-run` selector that matches nothing **exits 0 and prints `PASS`**. Measured in this
tree at `069795602`, against a test name this SPEC's previous draft cited as if it existed:

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabled' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.673s [no tests to run]
$ echo $?
0
```

Six criteria of the previous draft (AC-CEF-002, 003, 004, 007, 008, 009) named tests that do not
exist. Each was therefore satisfiable **having run nothing**.

**The obligation.** Every criterion below whose Verify line uses `-run` carries this clause, and it
is part of the Then, not a footnote:

> **Swept-count.** The `-v` output contains a `--- PASS:` line for **each** test named by this
> criterion's selector. A run whose output contains `no tests to run` — or which produces fewer
> `--- PASS:` lines than the criterion names — is a **FAIL**, not a pass. The `ok … [no tests to
> run]` line is a zero-match signal, never a green one.

**Why not a `$`-anchored grep.** A `--- PASS:` line ends with ` (0.06s)`, so a pattern anchored at
`$` on the test name matches nothing and reads as a clean zero forever. Match the test name
followed by a space, or count `--- PASS:` occurrences.

---

## §D AC Matrix

| AC | Requirement | Shape under test | Expected | RED-now |
|---|---|---|---|---|
| AC-CEF-001 | REQ-CEF-003 | `enabled` line absent | fatal, names `enabled` | **observed FAIL** (§D.1) |
| AC-CEF-002 | REQ-CEF-004 | `enabled = 1` | fatal, names `enabled` | not observable — test absent |
| AC-CEF-003 | REQ-CEF-004 | `enabled = "true"` | fatal, names `enabled` | not observable — test absent |
| AC-CEF-004 | REQ-CEF-004 | `enabled = 'false'` | fatal, names `enabled` | not observable — test absent |
| AC-CEF-005 | REQ-CEF-005 | `enabled = true` | NOT fatal — **control** | **observed PASS** (baseline) |
| AC-CEF-006 | REQ-CEF-005 | `enabled = false` | NOT fatal — **control** | not observable — row absent |
| AC-CEF-007 | REQ-CEF-006 | `enabled = true`, `path` absent | NOT fatal — **scope control** | not observable — test absent |
| AC-CEF-008 | REQ-CEF-008 | fatal fixture, EXECUTED process | `moai doctor` exits 1 | not observable — test absent |
| AC-CEF-009 | REQ-CEF-009 | clean fixture, EXECUTED process | `moai doctor` exits 0 | not observable — test absent |
| AC-CEF-010 | REQ-CEF-001, 007 | parser + finding severity | tri-state widened; severity carried | not observable — test absent |
| AC-CEF-011 | REQ-CEF-010 | stale-path finding | trigger, grade, template unchanged | **observed PASS** (baseline) |
| AC-CEF-012 | REQ-CEF-002, 012 | read-only posture | no write under `CODEX_HOME` | not observable — helper absent |
| AC-CEF-013 | REQ-CEF-013 | reversed sites | rationale rewritten, not deleted | n/a — diff criterion |
| AC-CEF-014 | REQ-CEF-009 | non-`enabled` warn finding, EXECUTED | `CheckWarn`, exits 0 | not observable — test absent |
| AC-CEF-015 | REQ-CEF-011 | fatal finding text | names codex 0.153.4 | not observable — test absent |
| AC-CEF-016 | REQ-CEF-014 | severity zero value | zero value is advisory | not observable — test absent |

---

## §D.1 Criteria

### AC-CEF-001 — missing `enabled` key is fatal

**Given** a `[[skills.config]]` entry whose `path` resolves to an existing file and which declares
no `enabled` key,
**When** the Codex Wiring check runs against that config,
**Then** the check's status is `uikit.CheckFail`, its combined message-plus-detail text contains
the token `enabled`, and the **swept-count obligation (§D.0)** holds for the named test.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal' -count=1 -v`

**RED-now, `069795602`, exit 1:**

```
=== RUN   TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal
    doctor_codex_enabled_test.go:37: status = ok, want CheckFail — codex cannot start on this config: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
    doctor_codex_enabled_test.go:41: finding never names the `enabled` key: {Name:Codex Wiring Status:ok Message:wired and consistent (hooks valid, sidecar matches, moai on PATH, config canonical) Detail:}
--- FAIL: TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.774s
```

This is the existing RED guard in `internal/cli/doctor_codex_enabled_test.go:27`. It is the seed of
this suite, not a new criterion, and it is the ONE detection criterion whose RED was observed.

### AC-CEF-002 — integer `enabled` is fatal

**Given** an entry declaring `enabled = 1`,
**When** the Codex Wiring check runs,
**Then** the status is `uikit.CheckFail`, the finding text names `enabled`, and §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabled' -count=1 -v`

**RED-now: NOT OBSERVABLE.** `TestCheckCodexWiring_NonBooleanEnabled` does not exist at
`069795602`; the selector sweeps zero tests and exits 0 (§D.0, verbatim). The test is authored in
**M4**. Per `verification-completeness.md` §2.1 a criterion with no observed RED is **not
release-blocking**: AC-CEF-002 is classified a **regression guard** in §D.2, not MUST-PASS.

### AC-CEF-003 — double-quoted `enabled` is fatal

**Given** an entry declaring `enabled = "true"`,
**When** the Codex Wiring check runs,
**Then** the status is `uikit.CheckFail`, the finding text names `enabled`, and §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabled' -count=1 -v` —
the `enabled = "true"` row of that table. §D.0's per-row reading applies: the table's `-v` output
must carry a `--- PASS:` sub-test line for this row specifically, not only for the parent.

**RED-now: NOT OBSERVABLE** (test absent; authored in M4). Regression guard, per AC-CEF-002.

This is the criterion that carries the operator's B.1 reversal. Before this SPEC the same input
produced `SkillEnabledTrue` and a silent, healthy check.

### AC-CEF-004 — single-quoted `enabled` is fatal

**Given** an entry declaring `enabled = 'false'`,
**When** the Codex Wiring check runs,
**Then** the status is `uikit.CheckFail`, the finding text names `enabled`, and §D.0 holds.

Verify: as AC-CEF-003, the `enabled = 'false'` row.

**RED-now: NOT OBSERVABLE** (test absent; authored in M4). Regression guard.

### AC-CEF-005 — bare `true` stays non-fatal (CONTROL)

**Given** an entry with a resolving `path` and `enabled = true`,
**When** the Codex Wiring check runs,
**Then** the status is not `uikit.CheckFail`, **and** the check's detail text mentions the
fixture's entry — proving the entry was actually READ rather than the fixture degrading to the
codex-not-in-play informational skip, which also satisfies "not CheckFail". §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet' -count=1 -v`

**Control baseline — PARTIAL, `069795602`, exit 0:**

```
=== RUN   TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet
--- PASS: TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.697s
```

**What this PASS attributes, and what it does not.** Status-negative half observed at `069795602`;
positive-read half NOT observable at plan-time — authored in M4 item 3. The passing test's body
asserts only that the status is not `uikit.CheckFail`; it makes no assertion about the detail text,
so the observed PASS attributes the OLD, weaker criterion. Under
`verification-claim-integrity.md` §2 the positive-read half is **unattributed** until M4 item 3
lands, and §D.2's release-blocking classification of this criterion covers the status-negative half
ONLY. The Then above is not weakened by this: both halves must hold for AC-CEF-005 to be satisfied.

Mandatory. Without it, AC-CEF-001..004 pass equally well against a check that fails every config.
The positive-read clause is new: the control as previously written asserted only a negative, which
a skipped check satisfies.

### AC-CEF-006 — bare `false` stays non-fatal (CONTROL)

**Given** an entry with a resolving `path` and `enabled = false`,
**When** the Codex Wiring check runs,
**Then** the status is not `uikit.CheckFail`, the detail text mentions the fixture's entry, and
§D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet' -count=1 -v`
— the `enabled = false` row.

> **Selector note (was a defect).** The previous draft said "same command as AC-CEF-005
> (table-driven row)", but `TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet` at `069795602` is a
> **single-case test with no table**, so no such row exists to run. **M4 converts it into a table**
> with a `bare_true` and a `bare_false` case; this criterion's swept-count is then satisfied by the
> `--- PASS: …/bare_false` sub-test line. Until that conversion lands the selector sweeps the
> parent only and AC-CEF-006 is NOT satisfied by it.

**RED-now: NOT OBSERVABLE** — the `false` row does not exist at `069795602`. Regression guard.

Mandatory and distinct from AC-CEF-005: `false` is the shape 49/49 of the real machine's entries
declare. A fix that made `false` fatal would break every existing user while passing AC-CEF-005.

### AC-CEF-007 — `path`-absent stays out of scope (SCOPE CONTROL)

**Given** an entry declaring `enabled = true` and no `path` key,
**When** the Codex Wiring check runs,
**Then** the status is not `uikit.CheckFail`, the detail text mentions the fixture's entry (the
positive-read clause of AC-CEF-005 applies identically), and §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_PathAbsentIsNotFatal' -count=1 -v`

**RED-now: NOT OBSERVABLE** (test absent; authored in M4). Regression guard.

Mandatory as a scope control regardless of its release-blocking classification: codex accepts this
shape (lab row 6), and a fix that widened fatality to `path` would be a scope breach invisible to
AC-CEF-001..006.

### AC-CEF-008 — `moai doctor` exits 1 on a fatal fixture

**Given** a `CODEX_HOME` containing a config with a missing-`enabled` entry,
**When** the `moai doctor` **process** is executed against it,
**Then** the observed process exit status is 1, and §D.0 holds for the named test.

Verify: an executed process-level test — one that BUILDS the binary (`go test` with
`os/exec` against a `go build`-produced or `testscript`-driven binary, or `exec.Command(os.Args[0])`
with a re-exec guard) and reads the **real process exit status**, with `CODEX_HOME` pinned to the
fixture. Suggested name: `TestDoctorExitCode_CodexEnabledFatal`.

[HARD] **Forbidden satisfying forms.** This criterion is NOT satisfied by any of:

- calling `doctorExitStatus(1)` and asserting the returned `*exitCodeError` is non-nil;
- calling `doctorExitStatus(countFailedChecks(...))` on a synthesised check slice;
- asserting on `uikit.CheckFail` alone without a process.

Those forms are not hypothetical. `internal/cli/exitcode_contract_test.go:30,38` and
`internal/cli/binary_lag_test.go:102` already do exactly this in-package, so the previous draft's
"a plain in-package Go test" Verify line was satisfiable by copying an existing pattern that never
starts a process. The Then clause requires an **observed process exit status**; a function's return
value is a hypothesis about what the process would do.

**RED-now: NOT OBSERVABLE** (no process-level test exists at `069795602`; authored in M4).
Regression guard — but see §D.2: this is the criterion whose loss of release-blocking status is
most consequential, and M4 SHOULD author it early enough to restore an observed RED before close.

### AC-CEF-009 — `moai doctor` exits 0 on a clean fixture (CONTROL)

**Given** a `CODEX_HOME` whose entries all declare a bare boolean `enabled` and whose paths all
resolve,
**When** the `moai doctor` process is executed against it,
**Then** the observed process exit status is 0, and §D.0 holds.

Verify: the paired case of AC-CEF-008's process-level test. The same [HARD] forbidden-forms clause
applies.

**RED-now: NOT OBSERVABLE** (test absent; authored in M4). Regression guard.

Mandatory as a control: pairs with AC-CEF-008 so a doctor that exits 1 unconditionally cannot pass.
Note its limitation — this fixture takes the `len(problems) == 0 → CheckOK` branch and exercises no
warn finding at all, which is why AC-CEF-014 exists.

### AC-CEF-010 — the reading and severity axes exist

**Given** the changed parser and check,
**When** the package tests run,
**Then** the parser distinguishes absent / bare-boolean / declared-non-boolean as three separate
readings, and a check run producing one fatal and one advisory finding reports fatal status while
**still surfacing the advisory finding's text**.

Verify: `go test ./internal/codexwiring/... ./internal/cli/... -count=1` — a whole-package run, so
§D.0's zero-match hazard does not arise; instead the obligation is that the run is not empty
(`ok … [no test files]` on either package is a FAIL).

**RED-now: NOT OBSERVABLE** (the mixed-finding test does not exist at `069795602`; authored in M4).
Regression guard.

The mixed-finding case is the one that proves the severity axis was BUILT rather than the whole
check being flipped to fatal.

### AC-CEF-011 — the stale-path finding is unchanged

**Given** a config that triggers the existing stale-path finding and declares `enabled` as a bare
boolean throughout,
**When** the Codex Wiring check runs,
**Then** the finding fires on its current trigger conditions, at advisory grade, with its **message
template** byte-identical to the pre-change tree, and §D.0 holds.

Verify:
`go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported' -count=1 -v` plus
`go test ./internal/cli/ -run 'TestDoctorGolden_NoColor' -count=1 -v`
(`internal/cli/testdata/doctor-nocolor.golden` passing unmodified on this axis).

> **Selector note.** The previous draft's `-run 'TestCheckCodexWiring'` prefix-matches the RED guard
> `TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal`, so it cannot be green until M3 lands and
> its result says nothing about the stale-path finding specifically. The narrowed selector above is
> the one this criterion means.

**Baseline, `069795602`, exit 0:**

```
=== RUN   TestCheckCodexWiring_StaleHomeSkillsReported
--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.745s

=== RUN   TestDoctorGolden_NoColor
--- PASS: TestDoctorGolden_NoColor (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.961s
```

Both Verify commands are attributed: this criterion names two, and the baseline above quotes both
at `069795602`.

**Expected, permitted change — not a regression.** Per REQ-CEF-010's rescoping, the declared-split
COUNTS rendered by `internal/cli/doctor_codex.go:757` change as a consequence of REQ-CEF-001: a
quoted `"true"` entry leaves the `enabled` bucket. Exactly two surfaces are expected to move:

| Surface | Expected change |
|---|---|
| `internal/cli/doctor_codex.go:757` | rendered counts differ for a config containing a quoted `enabled`; the format string's literal text is preserved |
| `internal/cli/doctor_codex_test.go:302` | the live assertion `"(1 enabled, 0 disabled, 1 unspecified)"` no longer holds and must be restated |

Any OTHER change to this finding's trigger, grade, or template text is a regression.

### AC-CEF-012 — nothing writes under `CODEX_HOME`

**Given** any fixture in this suite,
**When** the full suite runs,
**Then** the fixture's `CODEX_HOME` config file's content hash is unchanged from before the run.

Verify: a hash-before / hash-after assertion inside the test helper, exercised by
`go test ./internal/cli/ ./internal/codexwiring/ -count=1` (whole-package; the non-empty obligation
of AC-CEF-010 applies).

**RED-now: NOT OBSERVABLE** (the helper's hash assertion does not exist at `069795602`; authored in
M4). Regression guard.

### AC-CEF-013 — the reversed sites carry a rewritten rationale

**Given** the eight sites enumerated in plan.md §B.3,
**When** the change lands,
**Then** each production-comment site carries a replacement rationale that answers **both** grounds
of the comment it replaces (spec.md §B.1) rather than having the previous rationale deleted, and
each assertion site carries its new expectation together with a comment naming the measurement that
reversed it.

Verify: read the diff. `git diff internal/codexwiring/skills.go internal/codexwiring/skills_test.go
internal/cli/doctor_codex_test.go` shows a replacement rationale at every comment site and no bare
expectation flip at any assertion site.

**RED-now: n/a** — this is a diff-inspection criterion, not a test selector; §D.0 does not apply.

This criterion exists because the reversal is the part of this change most likely to be read later
as an unexplained regression. A deleted rationale leaves the next reader with the old comment's
argument in the history and no answer to it.

### AC-CEF-014 — an advisory-only run stays warn and exits 0

**Given** a `CODEX_HOME` and project state producing **at least one non-`enabled` advisory
finding** — the hooks-whitelist-missing finding or the sidecar-hash divergence finding, either of
which reaches `problems = append(...)` at a site this SPEC does not modify — and no fatal finding,
**When** the Codex Wiring check runs AND the `moai doctor` process is executed against the same
fixture,
**Then** the check's status is `uikit.CheckWarn` (not `CheckOK`, not `CheckFail`), the advisory
finding's text is present in the message-plus-detail, the observed process exit status is 0, and
§D.0 holds.

Verify: a process-level test paired with AC-CEF-008's, plus the in-package status assertion.
Suggested name: `TestDoctorExitCode_CodexAdvisoryOnlyStaysZero`.

**RED-now: NOT OBSERVABLE** (test absent; authored in M4). Regression guard.

**Why this is not covered by AC-CEF-009.** AC-CEF-009's clean fixture produces `len(problems) == 0`
and takes the `CheckOK` branch, so it exercises **no warn finding at all**. A severity fold that
re-graded every warn to fatal would still pass AC-CEF-009 and still exit 0 there — and would exit 1
on every real advisory machine. This criterion is the one that catches it.

### AC-CEF-015 — the fatal finding names the measured codex version

**Given** any fixture that produces a fatal `enabled` finding,
**When** the Codex Wiring check runs,
**Then** the finding's message-plus-detail text contains the measured codex version string
(`0.153.4`) wherever it characterises what codex accepts, and contains no clause asserting the
behaviour holds for codex releases generally, and §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_FatalFindingNamesMeasuredVersion' -count=1 -v`

**RED-now: NOT OBSERVABLE** (test absent; authored in M4). Regression guard.

> **This criterion replaces a false traceability row.** The previous draft mapped REQ-CEF-011
> (finding text attributes behaviour to the measured version) onto AC-CEF-013 (reversed tests carry
> a rewritten rationale). Those are different subjects, so REQ-CEF-011 was uncovered and AC-CEF-013
> orphaned while §D.3's summary read "every requirement is covered". `moai spec lint`'s
> `CoverageIncomplete` rule counts references and cannot detect a subject mismatch — this is a
> reader-level obligation (spec.md §F).

### AC-CEF-016 — the advisory grade is the severity zero value

**Given** the severity type introduced in M1,
**When** a `codexFinding` is constructed by a composite literal that names no severity — the shape
all 13 pre-existing construction sites use (spec.md §B.2) —
**Then** its severity equals the advisory grade, and a check run containing only such findings
reports `uikit.CheckWarn`. §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCodexFindingZeroValueIsAdvisory' -count=1 -v`

**RED-now: NOT OBSERVABLE** (the severity type does not exist at `069795602`; authored in M1, test
in M4). Regression guard.

**Why this is distinct from AC-CEF-014.** AC-CEF-014 is satisfiable by explicitly setting the
advisory grade at all 13 sites — which leaves the invariant unpinned, so the 14th site added six
months from now silently re-grades to fatal and `moai doctor` starts exiting 1 on a machine nobody
changed. This criterion pins the zero value itself.

---

## §D.2 Severity classification

Classification follows `verification-completeness.md` §2.1: **a criterion with no observed RED at
the pinned tree is not release-blocking.** Ten criteria were MUST-PASS in the previous draft on the
strength of tests that do not exist; that is the classification this section corrects.

| Class | ACs | Failing means |
|---|---|---|
| **MUST-PASS (release-blocking)** — RED or baseline observed at `069795602` | AC-CEF-001 (RED observed), AC-CEF-005 (**control baseline PARTIAL** — status-negative half observed at `069795602`; positive-read half NOT observable at plan-time, authored in M4 item 3, so the release-blocking claim covers the status-negative half ONLY), AC-CEF-011 (baseline observed) | the SPEC is not delivered |
| **MUST-PASS (control)** — mandatory in every suite run regardless of RED eligibility | AC-CEF-005, 006, 007, 009 | the passing ACs are vacuous — the suite asserts nothing |
| **Regression guard** — no observed RED; authored during the run | AC-CEF-002, 003, 004, 006, 007, 008, 009, 010, 012, 014, 015, 016 | delivered but under-evidenced on that axis; blocks sync, not run |
| **Diff criterion** — inspected, not executed | AC-CEF-013 | the reversal lands unexplained; blocks sync |

**Consequence, stated rather than buried.** Only three criteria are release-blocking on evidence
observed at plan-time, and the exit-code contract (AC-CEF-008) is NOT among them. Two mitigations
apply and both are in the plan:

1. **M4 authors the process-level test first among its tasks**, so its RED is observed against the
   pre-M1 severity behaviour and AC-CEF-008 can be promoted to release-blocking within the run.
2. **The control obligation is independent of the severity class.** AC-CEF-005, 006, 007 and 009
   are mandatory in every run whatever their release-blocking status — a run that skips them
   produces a vacuous green regardless of what the matrix says.

---

## §D.3 Traceability

Every mapping below was checked by **subject**, not by reference count. `moai spec lint`'s
`CoverageIncomplete` rule counts references only.

| Requirement | Subject | Covering ACs |
|---|---|---|
| REQ-CEF-001 | parser distinguishes three cases | AC-CEF-010 |
| REQ-CEF-002 | parser writes nothing | AC-CEF-012 |
| REQ-CEF-003 | absent key is fatal | AC-CEF-001 |
| REQ-CEF-004 | non-boolean value is fatal | AC-CEF-002, 003, 004 |
| REQ-CEF-005 | bare boolean is not fatal | AC-CEF-005, 006 |
| REQ-CEF-006 | absent `path` is not fatal | AC-CEF-007 |
| REQ-CEF-007 | per-finding severity exists | AC-CEF-010 |
| REQ-CEF-008 | fatal finding ⇒ fail status, exit non-zero | AC-CEF-008 |
| REQ-CEF-009 | advisory-only ⇒ status and exit unchanged | AC-CEF-009, 014 |
| REQ-CEF-010 | stale-path trigger/grade/template preserved | AC-CEF-011 |
| REQ-CEF-011 | finding text names the measured version | AC-CEF-015 |
| REQ-CEF-012 | nothing rewrites the user config | AC-CEF-012 |
| REQ-CEF-013 | reversal carries a replacement rationale | AC-CEF-013 |
| REQ-CEF-014 | advisory is the severity zero value | AC-CEF-016 |

14 requirements, 16 criteria. Every requirement is covered by a criterion **about the same
subject**; no criterion is orphaned.

---

## §D.4 Explicit Gaps (must be carried into the run-phase report verbatim)

- **Only codex-cli 0.153.4 was measured.** Whether older codex releases tolerate an absent
  `enabled`, and in which release the requirement appeared, is UNMEASURED. The finding's wording
  must not claim version-independence (REQ-CEF-011, AC-CEF-015). A criterion asserting behaviour
  across codex versions cannot be written from the evidence this SPEC has.
- **REQ-CEF-004's class is an INDUCTION from three measured shapes.** Integer, double-quoted string
  and single-quoted string were probed; float, array, inline table, and bareword (`yes`) were not.
  Codex's uniform ``invalid type: … expected a boolean`` error text is the ground for generalising,
  and the generalisation remains an induction (spec.md §C).
- **Multi-entry reporting is unmeasured.** Whether codex reports the first offending entry or all
  of them was not probed. No AC depends on it; it would affect finding wording only.
- **`enabled` inside a multi-line string, or with a trailing comment,** was not probed against
  codex. The parser has its own handling for both and this SPEC does not change those paths.
- **The t506 prune verb was READ, not EXECUTED.** The claim "the prune cannot create the fatal
  shape" (spec.md §A) rests on reading `pruneCodexSkillEntries` and `judgeCodexSkillEntry` on
  develop; no fixture was run against it. A run-phase fixture — prune a config containing a
  fatal-shape entry, assert the entry survives and no new fatal shape is manufactured — is cheap
  and SHOULD be added rather than inheriting the reading.
- **Card t502, the other named writer, was not checked** for whether it has landed.
- **REQ-CEF-009's "unchanged" baseline is pinned to `069795602`.** "Unchanged from the current
  behaviour" without a tree pin is unattributable; the pin is what makes the comparison checkable.

---

## §D.5 Definition of Done

- [ ] **Swept-count (§D.0) satisfied on every `-run` criterion**: the `-v` output carries a
      `--- PASS:` line for each test the criterion names, and no run reports `no tests to run`. A
      criterion whose evidence is an `ok … [no tests to run]` line is recorded as a FAIL.
- [ ] All MUST-PASS ACs green, with the four control rows (AC-CEF-005, 006, 007, 009) green in the
      SAME run.
- [ ] AC-CEF-008 and AC-CEF-009 satisfied by an **observed process exit status**, cited verbatim —
      not by a `doctorExitStatus` return value (§D.1 forbidden forms).
- [ ] AC-CEF-014 green: an advisory-only fixture reports `CheckWarn` and the process exits 0.
- [ ] AC-CEF-016 green: the severity zero value is advisory.
- [ ] The pre-existing RED guard (`doctor_codex_enabled_test.go:27`) now passes, with its control
      (`:49`) still passing.
- [ ] All eight reversal sites (plan.md §B.3) accounted for in the diff, with AC-CEF-013's
      replacement-rationale obligation met at every comment site.
- [ ] `go test ./internal/cli/... ./internal/codexwiring/...` green; `go vet` clean on both.
- [ ] `internal/codexwiring/skills_extent_test.go` still green — the run's new tests must not
      collide with the extent pins t506 landed.
- [ ] The real `~/.codex` byte-identical before and after the full run (hash cited).
- [ ] The §D.4 gaps restated verbatim in the run-phase evidence, not silently dropped.

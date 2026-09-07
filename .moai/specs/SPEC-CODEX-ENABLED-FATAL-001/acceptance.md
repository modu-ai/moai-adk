# SPEC-CODEX-ENABLED-FATAL-001 — Acceptance Criteria

Every criterion below is binary-testable by the named command. The measured-fixture criteria
(AC-CEF-001..006) map one-to-one onto the six rows of the acceptance lab
(`.moai/reports/t508/codex-enabled-lab.md`), so the guard's fixture set and the lab's fixture set
are the same set.

**Control-row obligation.** Every criterion of the "a fatal shape is detected" family is paired
with an accepting-shape control in the same run. An absence-guard that also passes when the feature
is entirely missing asserts nothing; the control is what makes a mutant's failure informative.
AC-CEF-005 and AC-CEF-006 are those controls and are mandatory in every run of this suite.

**Fixture isolation.** Every fixture is created under `t.TempDir()` with `CODEX_HOME` pinned per
fixture. No criterion reads or writes the machine's real `~/.codex`.

---

## §D AC Matrix

| AC | Requirement | Shape under test | Expected |
|---|---|---|---|
| AC-CEF-001 | REQ-CEF-003 | `enabled` line absent | fatal, names `enabled` |
| AC-CEF-002 | REQ-CEF-004 | `enabled = 1` | fatal, names `enabled` |
| AC-CEF-003 | REQ-CEF-004 | `enabled = "true"` | fatal, names `enabled` |
| AC-CEF-004 | REQ-CEF-004 | `enabled = 'false'` | fatal, names `enabled` |
| AC-CEF-005 | REQ-CEF-005 | `enabled = true` | NOT fatal — **control** |
| AC-CEF-006 | REQ-CEF-005 | `enabled = false` | NOT fatal — **control** |
| AC-CEF-007 | REQ-CEF-006 | `enabled = true`, `path` absent | NOT fatal — **scope control** |
| AC-CEF-008 | REQ-CEF-008 | fatal fixture, whole process | `moai doctor` exits 1 |
| AC-CEF-009 | REQ-CEF-009 | clean fixture, whole process | `moai doctor` exits 0 |
| AC-CEF-010 | REQ-CEF-001, 007 | parser + finding severity | tri-state widened; severity carried |
| AC-CEF-011 | REQ-CEF-010 | stale-path finding | behaviour and wording unchanged |
| AC-CEF-012 | REQ-CEF-002, 012 | read-only posture | no write under `CODEX_HOME` |
| AC-CEF-013 | REQ-CEF-011 | reversed quoted-value tests | rationale rewritten, not deleted |

---

## §D.1 Criteria

### AC-CEF-001 — missing `enabled` key is fatal

**Given** a `[[skills.config]]` entry whose `path` resolves to an existing file and which declares
no `enabled` key,
**When** the Codex Wiring check runs against that config,
**Then** the check's status is `uikit.CheckFail` and its combined message-plus-detail text contains
the token `enabled`.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal' -count=1 -v`

This is the existing RED guard in `internal/cli/doctor_codex_enabled_test.go`, which fails against
the current tree. It is the seed of this suite, not a new criterion.

### AC-CEF-002 — integer `enabled` is fatal

**Given** an entry declaring `enabled = 1`,
**When** the Codex Wiring check runs,
**Then** the status is `uikit.CheckFail` and the finding text names `enabled`.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabled' -count=1 -v`

### AC-CEF-003 — double-quoted `enabled` is fatal

**Given** an entry declaring `enabled = "true"`,
**When** the Codex Wiring check runs,
**Then** the status is `uikit.CheckFail` and the finding text names `enabled`.

Verify: same command as AC-CEF-002 (table-driven row).

This is the criterion that carries the operator's B.1 reversal. Before this SPEC the same input
produced `SkillEnabledTrue` and a silent, healthy check.

### AC-CEF-004 — single-quoted `enabled` is fatal

**Given** an entry declaring `enabled = 'false'`,
**When** the Codex Wiring check runs,
**Then** the status is `uikit.CheckFail` and the finding text names `enabled`.

Verify: same command as AC-CEF-002 (table-driven row).

### AC-CEF-005 — bare `true` stays non-fatal (CONTROL)

**Given** an entry with a resolving `path` and `enabled = true`,
**When** the Codex Wiring check runs,
**Then** the status is not `uikit.CheckFail`.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet' -count=1 -v`

Mandatory. Without it, AC-CEF-001..004 pass equally well against a check that fails every config.

### AC-CEF-006 — bare `false` stays non-fatal (CONTROL)

**Given** an entry with a resolving `path` and `enabled = false`,
**When** the Codex Wiring check runs,
**Then** the status is not `uikit.CheckFail`.

Verify: same command as AC-CEF-005 (table-driven row).

Mandatory and distinct from AC-CEF-005: `false` is the shape 49/49 of the real machine's entries
declare. A fix that made `false` fatal would break every existing user while passing AC-CEF-005.

### AC-CEF-007 — `path`-absent stays out of scope (SCOPE CONTROL)

**Given** an entry declaring `enabled = true` and no `path` key,
**When** the Codex Wiring check runs,
**Then** the status is not `uikit.CheckFail`.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_PathAbsentIsNotFatal' -count=1 -v`

Mandatory. Codex accepts this shape (lab row 6); a fix that widened fatality to `path` would be a
scope breach invisible to AC-CEF-001..006.

### AC-CEF-008 — `moai doctor` exits 1 on a fatal fixture

**Given** a `CODEX_HOME` containing a config with a missing-`enabled` entry,
**When** the `moai doctor` process is EXECUTED against it,
**Then** the process exit code is 1.

Verify: an executed process-level test asserting the exit code, e.g.
`go test ./internal/cli/ -run 'TestDoctorExitCode_CodexEnabledFatal' -count=1 -v`

[HARD] This criterion must be satisfied by EXECUTING the command, not by reading
`doctorExitStatus`. The exit path was read from source during plan-phase and never run; a
source-read is a hypothesis about behaviour, not an observation of it.

### AC-CEF-009 — `moai doctor` exits 0 on a clean fixture (CONTROL)

**Given** a `CODEX_HOME` whose entries all declare a bare boolean `enabled`,
**When** the `moai doctor` process is EXECUTED against it,
**Then** the process exit code is 0.

Verify: same command as AC-CEF-008 (paired case).

Mandatory. Pairs with AC-CEF-008 so a doctor that exits 1 unconditionally cannot pass.

### AC-CEF-010 — the reading and severity axes exist

**Given** the changed parser and check,
**When** the package tests run,
**Then** the parser distinguishes absent / bare-boolean / declared-non-boolean as three separate
readings, and a check run producing one fatal and one advisory finding reports fatal status while
still surfacing the advisory finding's text.

Verify: `go test ./internal/codexwiring/... ./internal/cli/... -count=1`

The mixed-finding case is the one that proves the severity axis was BUILT rather than the whole
check being flipped to fatal.

### AC-CEF-011 — the stale-path finding is unchanged

**Given** a config that triggers the existing stale-path finding and declares `enabled` as a bare
boolean throughout,
**When** the Codex Wiring check runs,
**Then** the finding fires on its current trigger conditions, at advisory grade, with wording
byte-identical to the pre-change tree.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring' -count=1` plus the committed golden
fixtures (`internal/cli/testdata/doctor-nocolor.golden`) passing unmodified on this axis.

### AC-CEF-012 — nothing writes under `CODEX_HOME`

**Given** any fixture in this suite,
**When** the full suite runs,
**Then** the fixture's `CODEX_HOME` config file's content hash is unchanged from before the run.

Verify: a hash-before / hash-after assertion inside the test helper, exercised by
`go test ./internal/cli/ ./internal/codexwiring/ -count=1`.

### AC-CEF-013 — the reversed tests carry a rewritten rationale

**Given** `internal/codexwiring/skills_test.go:66` `TestParseSkillEntriesEnabledQuoted` and the
`internal/cli/doctor_codex_test.go:297` fixture comment ("quoted string, still true"),
**When** the change lands,
**Then** each carries its new expectation together with a comment stating why the previous
expectation was reversed — the measurement that falsified it — rather than having the old comment
merely deleted.

Verify: read the diff; `git diff` on those two files shows a replacement rationale, not a bare
expectation flip.

This criterion exists because the reversal is the part of this change most likely to be read later
as an unexplained regression. A deleted rationale leaves the next reader with the old comment's
argument in the history and no answer to it.

---

## §D.2 Severity

| Class | ACs | Failing means |
|---|---|---|
| MUST-PASS | AC-CEF-001..009, AC-CEF-012 | the SPEC is not delivered |
| MUST-PASS (control) | AC-CEF-005, 006, 007, 009 | the passing ACs are vacuous — the suite asserts nothing |
| SHOULD-PASS | AC-CEF-010, 011, 013 | delivered but under-evidenced; blocks sync, not run |

---

## §D.3 Traceability

| Requirement | Covering ACs |
|---|---|
| REQ-CEF-001 | AC-CEF-010 |
| REQ-CEF-002 | AC-CEF-012 |
| REQ-CEF-003 | AC-CEF-001 |
| REQ-CEF-004 | AC-CEF-002, 003, 004 |
| REQ-CEF-005 | AC-CEF-005, 006 |
| REQ-CEF-006 | AC-CEF-007 |
| REQ-CEF-007 | AC-CEF-010 |
| REQ-CEF-008 | AC-CEF-008 |
| REQ-CEF-009 | AC-CEF-009 |
| REQ-CEF-010 | AC-CEF-011 |
| REQ-CEF-011 | AC-CEF-013 |
| REQ-CEF-012 | AC-CEF-012 |

Every requirement is covered; no AC is orphaned.

---

## §D.4 Explicit Gaps (must be carried into the run-phase report)

- **Only codex-cli 0.153.4 was measured.** Whether older codex releases tolerate an absent
  `enabled`, and in which release the requirement appeared, is UNMEASURED. The finding's wording
  must not claim version-independence (REQ-CEF-011, AC-CEF-013). A criterion asserting behaviour
  across codex versions cannot be written from the evidence this SPEC has.
- **Multi-entry reporting is unmeasured.** Whether codex reports the first offending entry or all
  of them was not probed. No AC depends on it; it would affect finding wording only.
- **`enabled` inside a multi-line string, or with a trailing comment,** was not probed against
  codex. The parser has its own handling for both and this SPEC does not change those paths.

---

## §D.5 Definition of Done

- [ ] All MUST-PASS ACs green, with the control rows green in the SAME run.
- [ ] AC-CEF-008 satisfied by an EXECUTED process exit code, cited verbatim — not a source-read.
- [ ] The pre-existing RED guard (`doctor_codex_enabled_test.go`) now passes, with its control still
      passing.
- [ ] `go test ./internal/cli/... ./internal/codexwiring/...` green; `go vet` clean.
- [ ] The real `~/.codex` byte-identical before and after the full run (hash cited).
- [ ] The §D.4 gaps restated verbatim in the run-phase evidence, not silently dropped.

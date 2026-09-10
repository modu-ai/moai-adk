# SPEC-CODEX-STALE-SPLIT-FOURTH-001 — Acceptance Criteria

8 criteria against the Tier S ceiling of 8.

**Document-level tree pin.** Every baseline cell in this file was measured in this tree at commit
**`e0c904f58`** (= `origin/develop`, the tip carrying t508). A cell citing no SHA is not a baseline;
a cell citing a different SHA is a carry-over (verification-claim-integrity §2).

**Control-row obligation.** The core criterion (AC-SSF-001) asserts that a population moves INTO a
new bucket. Without controls, a change that moved *everything* into `non-boolean` would satisfy it.
AC-SSF-002 and AC-SSF-003 are those controls and are mandatory in every run of this suite.

**Fixture isolation.** Every fixture is created under `t.TempDir()` with `CODEX_HOME` pinned per
fixture. `t.Setenv("HOME", ...)` is prohibited. No criterion reads or writes the machine's real
`~/.codex` (AC-SSF-008).

---

## §D.0 The swept-count obligation (binds EVERY `-run`-based criterion)

[HARD] A Go `-run` selector that matches nothing **exits 0 and prints `PASS`**. Reused verbatim from
`SPEC-CODEX-ENABLED-FATAL-001` §D.0, where six criteria of a previous draft named tests that did not
exist and were each satisfiable having run nothing.

**The obligation.** Every criterion below whose Verify line uses `-run` carries this clause, and it
is part of the Then, not a footnote:

> **Swept-count.** The `-v` output contains a `--- PASS:` line for **each** test named by this
> criterion's selector. A run whose output contains `no tests to run` — or which produces fewer
> `--- PASS:` lines than the criterion names — is a **FAIL**, not a pass. The `ok … [no tests to
> run]` line is a zero-match signal, never a green one.

**Why not a `$`-anchored grep.** A `--- PASS:` line ends with ` (0.06s)`, so a pattern anchored at
`$` on the test name matches nothing and reads as a clean zero forever. Match the test name followed
by a space, or count `--- PASS:` occurrences.

---

## §D AC Matrix

| AC | Requirement | Shape under test | Expected | Baseline at `e0c904f58` |
|---|---|---|---|---|
| AC-SSF-001 | REQ-SSF-001, 002 | absent path + `enabled = yes` | counted `non-boolean` | not observable — test absent (§D.2) |
| AC-SSF-002 | REQ-SSF-003 | absent path + no `enabled` key — **control** | counted `unspecified` | observed PASS (baseline) |
| AC-SSF-003 | REQ-SSF-001 | absent path + bare `true` / `false` — **control** | counted `enabled` / `disabled` | observed PASS (baseline) |
| AC-SSF-004 | REQ-SSF-004 | advisory fires, zero non-boolean entries | `0 non-boolean` still rendered | observed FAIL (§D.1) |
| AC-SSF-005 | REQ-SSF-005 | stale-path finding trigger + grade | unchanged | observed PASS (baseline) |
| AC-SSF-006 | REQ-SSF-006, 001 | fatal finding untouched + no `default:` arm names a bucket | untouched / explicit `case` | n/a — diff criterion |
| AC-SSF-007 | REQ-SSF-008 | each changed assertion | carries a why-comment | n/a — diff criterion |
| AC-SSF-008 | spec.md §D | fixture isolation | no real `~/.codex` access | n/a — diff criterion |

---

## §D.1 Criteria

### AC-SSF-001 — a non-boolean `enabled` on an absent path is counted `non-boolean`

**Given** a `[[skills.config]]` entry whose `path` does not exist and which declares `enabled` with
a value that is not a bare TOML boolean,
**When** the Codex Wiring check runs against that config,
**Then** the stale-path advisory detail contains the ordered phrase `1 non-boolean`, contains
`0 unspecified`, and the **swept-count obligation (§D.0)** holds for the named test.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabledCountedSeparatelyInStaleSplit' -count=1 -v`

**Baseline: NOT OBSERVABLE.** The named test does not exist at `e0c904f58`; the selector sweeps zero
tests and exits 0 (§D.0, verbatim). The test is authored in **M2**. Per
`verification-completeness.md` §2.1 a criterion with no observed RED is **not release-blocking**:
AC-SSF-001 is classified a **regression guard** in §D.2, not MUST-PASS. The defect it guards IS
observed — at the CLI level, in research.md §1 — which is what makes the guard worth having.

### AC-SSF-002 — control: an absent `enabled` key stays `unspecified`

**Given** a `[[skills.config]]` entry whose `path` does not exist and which declares no `enabled`
key at all,
**When** the Codex Wiring check runs,
**Then** that entry is counted in the `unspecified` bucket and NOT in the `non-boolean` bucket, and
§D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately' -count=1 -v`

**Baseline at `e0c904f58`, exit 0 — the assertion string changes, the behaviour under test does
not:**

```
=== RUN   TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately
--- PASS: TestCheckCodexWiring_UnspecifiedEnabledReportedSeparately (0.00s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.946s
```

This test currently asserts `(0 enabled, 0 disabled, 2 unspecified)` for a fixture holding one
absent-key entry and one non-boolean entry. After the change it asserts
`(0 enabled, 0 disabled, 1 unspecified, 1 non-boolean)` — the absent-key entry stays put, which is
the control, and the non-boolean entry moves, which is AC-SSF-001.

**Without this control** a change that routed every non-`true`/`false` state into `non-boolean`
would satisfy AC-SSF-001 while destroying the meaning of `unspecified`.

### AC-SSF-003 — control: bare booleans still land in `enabled` / `disabled`

**Given** a config whose missing-path entries declare `enabled` as bare TOML booleans throughout,
**When** the Codex Wiring check runs,
**Then** those entries are counted in the `enabled` and `disabled` buckets exactly as before, with
`0 non-boolean`, and §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCodexSkillPath_AbsoluteExistingAndMissing' -count=1 -v`

**Swept-count for this criterion: TWO `--- PASS:` lines are required** (one per named test). One
line is a FAIL.

**Baseline at `e0c904f58`, exit 0:** both tests pass today, asserting
`(1 enabled, 2 disabled, 0 unspecified)` and `(0 enabled, 1 disabled, 0 unspecified)` respectively.
After the change they assert the same numbers with `, 0 non-boolean` appended.

**Without this control** a change that mis-wired the new `case` arm — catching `SkillEnabledFalse`
in `non-boolean`, say — would still satisfy AC-SSF-001 and AC-SSF-002.

### AC-SSF-004 — the fourth counter renders unconditionally

**Given** any config that fires the stale-path advisory and contains zero non-boolean entries,
**When** the Codex Wiring check runs,
**Then** the advisory detail still contains the literal `0 non-boolean` inside the parenthesis, and
the assertion at `internal/cli/doctor_codex_test.go:279` asserts the full ordered phrase
`(1 enabled, 2 disabled, 0 unspecified, 0 non-boolean)`, so this criterion's Verify command observes
the Then rather than merely coexisting with it, and §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported' -count=1 -v`

**Baseline at `e0c904f58`, observed RED — the four elements `verification-completeness.md` §2.1
requires of a release-blocking criterion (command, verbatim stdout, exit code, tree SHA):**

```
$ grep -c 'non-boolean' internal/cli/doctor_codex.go
0
exit=1
```

Tree `e0c904f58`. The token does not occur in the production file at all, so the fourth member
cannot be rendered and the phrase in the Then cannot be present. The measured current assertion at
`doctor_codex_test.go:279` is `(1 enabled, 2 disabled, 0 unspecified)` (research.md §4), which
likewise contains no `non-boolean` token.

This is the criterion that discriminates the chosen unconditional form (spec.md §B.1) from the
rejected conditional one — **once the assertion binding named in its Then is in place**. Without
that binding the discrimination is not automatic: under a consistently implemented conditional form
the line-279 assertion is never edited, keeps asserting the three-member string, and this
criterion's Verify command returns green while the Then is unmet. The criterion discriminates
because line 279 names the four-member phrase, not because its own text forbids the conditional
form.

### AC-SSF-005 — the finding's trigger and grade are unchanged

**Given** a config that triggers the stale-path finding,
**When** the Codex Wiring check runs,
**Then** the finding fires on its current trigger conditions, at advisory grade, retaining its
leading `%d with a path that no longer exists` count and its trailing
`remove the stale entries or restore the skill files` directive, and §D.0 holds.

Verify: `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported|TestCodexSkillPath_AbsoluteExistingAndMissing' -count=1 -v`

**Swept-count: TWO `--- PASS:` lines required.**

**Baseline at `e0c904f58`, exit 0.** Both tests already assert the leading count phrase
(`3 with a path that no longer exists`, `1 stale skill entry`) and the remove directive; those
assertions are **not** edited by this SPEC, which is what makes them evidence of preservation
rather than of the change.

### AC-SSF-006 — the fatal `enabled`-shape finding is untouched

**Given** the run-phase diff,
**When** it is reviewed,
**Then** `codexEnabledShapeFinding` and its `absent` / `nonBoolean` accounting
(`internal/cli/doctor_codex.go:762-768`) appear nowhere in the diff, and the bucketing switch at
`internal/cli/doctor_codex.go:864-871` carries an explicit `case codexwiring.SkillEnabledNonBoolean:`,
with no `default:` arm incrementing a named bucket (REQ-SSF-001 — no state reached by a `default:`
arm; the same diff hunk shows both, so no additional command is needed).

Verify: `git diff --unified=0 e0c904f58..HEAD -- internal/cli/doctor_codex.go` — no hunk touches the
`codexEnabledShapeFinding` body. A hunk that does is a FAIL, not a judgement call: changing it would
recreate the duplicate-naming defect (spec.md §D).

### AC-SSF-007 — each changed assertion says why it changed

**Given** the run-phase diff over `internal/cli/doctor_codex_test.go`,
**When** it is reviewed,
**Then** each of the three edited assertion sites (research.md §4: lines 279, 349, 943 at
`e0c904f58`) carries a comment naming `SPEC-CODEX-STALE-SPLIT-FOURTH-001` as the cause of the
changed string, and the line-349 test's existing "deliberately does NOT grow a fourth bucket"
paragraph is **rewritten with its reversal rationale, not deleted**.

Verify: `git diff e0c904f58..HEAD -- internal/cli/doctor_codex_test.go` — three edited assertions,
three explanatory comments, one rewritten (not removed) rationale paragraph. A bare string edit with
no comment is a FAIL: the whole point is that a later reader must not read this as a silent
regression fix.

### AC-SSF-008 — fixtures never touch the real `~/.codex`

**Given** every test fixture added or edited by this SPEC,
**When** the diff is reviewed and the suite is run,
**Then** each fixture is created under `t.TempDir()` with `CODEX_HOME` pinned to it, no test calls
`t.Setenv("HOME", ...)`, and no path under the invoking user's real `~/.codex` is read or written.

Verify: `git diff e0c904f58..HEAD -- internal/cli/` piped through a grep for `t.Setenv("HOME"` —
any match is a FAIL. Fixture construction goes through the existing `writeCodexHomeConfig` /
`stubCodexHome` helpers, which already satisfy this.

---

## §D.2 Severity classification

A failure in any row below blocks the run-phase close. The rows differ in **what evidence backs the
classification**, which the predecessor (`SPEC-CODEX-ENABLED-FATAL-001` §D.2) separated and this
draft had collapsed into one undifferentiated block:

| Class | ACs | Evidence backing the class |
|---|---|---|
| **MUST-PASS (release-blocking)** — RED or baseline observed at `e0c904f58` | AC-SSF-004 | observed RED, four-element cell in §D.1 (command, stdout, exit code, tree SHA) |
| **MUST-PASS (control)** — mandatory in every suite run regardless of RED eligibility | AC-SSF-002, AC-SSF-003, AC-SSF-005 | green-now baselines at `e0c904f58`; they cannot go RED before the change, and their job is to keep the passing criteria non-vacuous |
| **MUST-PASS (diff criterion)** — inspected, not executed | AC-SSF-006, AC-SSF-007, AC-SSF-008 | verified by reading the run-phase diff; no test run produces their verdict |

**Regression guard** (not release-blocking, per `verification-completeness.md` §2.1): AC-SSF-001 —
its test does not exist at `e0c904f58`, so no RED was observed at the Go-test level. The defect it
guards is observed at the CLI level (research.md §1), so the guard is worth authoring; but a
criterion whose RED was never seen cannot block a close on its own.

## §D.3 Traceability

| Requirement | Covered by |
|---|---|
| REQ-SSF-001 | AC-SSF-001, AC-SSF-003, AC-SSF-006 (the no-`default:`-arm clause) |
| REQ-SSF-002 | AC-SSF-001 |
| REQ-SSF-003 | AC-SSF-002 |
| REQ-SSF-004 | AC-SSF-004 |
| REQ-SSF-005 | AC-SSF-005 |
| REQ-SSF-006 | AC-SSF-006 |
| REQ-SSF-007 | spec.md §B.2 — a document criterion, verified by plan-audit reading, not by a test |
| REQ-SSF-008 | AC-SSF-007 |

## §D.4 Definition of Done

- All seven MUST-PASS criteria observed green in one run, each with its §D.0 swept-count satisfied.
- `go test ./internal/cli/ -count=1` green (the touched package; the full suite is CI's verdict per
  the lane-local verification rule).
- `go vet ./internal/cli/...` clean.
- No Go file outside `internal/cli/doctor_codex.go` and `internal/cli/doctor_codex_test.go` modified.
- AC-SSF-001's RED output recorded verbatim in `progress.md` §E.2, captured AFTER the M1 render
  change, with the command and its exit code.
- spec.md §B.2 supersession recorded (REQ-SSF-007).

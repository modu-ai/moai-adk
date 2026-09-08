# SPEC-CODEX-DOCTOR-PATH-GUARD-001 — Acceptance Criteria

Card t570 · Tier S · `cycle_type tdd`. Every AC is binary-testable. Line numbers are hints measured
2026-09-08 at `a4855f0b2`; symbols are the address. Each AC carries a **RED-now** cell (what was
observed failing, on which tree) and a **green-path** cell — a guard whose failure was never
observed on a known failing input is not adopted.

**Tense contract.** This file is authored at `status: draft`, before any run. Every RED-now cell
below states an **expectation to be discharged in run-phase**, never an observation already made.
No cell here may be read as evidence; the evidence lands in `progress.md` §E.2.

**Verification command for every AC below**, run sequentially (Go applies only the LAST `-run`
flag, so one selector per invocation):

```
go test ./internal/cli/... -run '<Selector>' -timeout 1200s -v
```

`-timeout 1200s` is not optional: the package has a measured ~615 s runtime and the 600 s default
fails it spuriously. `go test ./...` is not run locally.

---

### AC-CDPG-001 — The doctor stat target is the converted form (maps REQ-CDPG-001)

**Given** `configPathSeparator` pinned to `'\\'` via `overrideSeparator(t)` (restores through
`t.Cleanup`; the test is non-parallel) and `osStatFn` replaced by `stubStatRecording(t)`,
**When** `codexStaleSkillFinding()` judges a `CODEX_HOME` fixture config declaring the absolute
slash-form path `/Users/u/skills/probe/SKILL.md`, **Then** the recorder's captured argument equals
`\Users\u\skills\probe\SKILL.md` (= `fromConfigPath(declared, '\\')`) and does NOT equal
`/Users/u/skills/probe/SKILL.md`.

The fixture string is pinned rather than left as "some absolute path" so the two operands cannot
collapse onto each other: the declared form contains `/` and the expected form contains none, so the
inequality half of the assertion is non-vacuous by inspection.

- **RED-now**: with `internal/cli/doctor_codex.go` reverted to `statPath = e.Path`, the recorder
  will observe the DECLARED form and the assertion will fail. **Nothing has been executed yet** —
  this is a plan-phase SPEC at `status: draft`. AC-CDPG-004 carries the obligation to actually
  perform that revert during run-phase and to record the observed output there; until that record
  exists, this cell states an expectation, not an observation.
- **Green path**: M1 step 1, against the unreverted line.
- **Boundary**: asserts the stat TARGET only. No deletion branch is exercised; the fixture lives
  under `t.TempDir()` and the real `~/.codex/config.toml` is never read or written.

### AC-CDPG-002 — The extended-length family, with its reachability premise verified in the test (maps REQ-CDPG-002)

**Given** the same pinned separator and recorder, **When** the fixture declares
`//?/C:/Users/u/skills/probe/SKILL.md`, **Then** the recorder's captured argument is exactly
`\\?\C:\Users\u\skills\probe\SKILL.md`, **and** the same test separately asserts
`classifyCodexSkillPath("//?/C:/Users/u/skills/probe/SKILL.md") == codexPathAbsolute`.

- **Why this family and not another**: Windows accepts `/` inside ordinary absolute paths, so every
  other absolute family degrades gracefully under the unconverted form. Inside a `\\?\`
  extended-length prefix Windows performs no separator normalization, so the unconverted form does
  not resolve there — this is the one family on which the omission produces a false "missing path"
  verdict against a healthy registration.
- **Why the second assertion is mandatory**: the first assertion is only meaningful if the entry
  reaches the converting arm. That depends on `filepath.IsAbs` being true for this declaration on
  darwin. Measured this run (`go run` probe): `IsAbs("//?/C:/Users/u/skills/probe/SKILL.md") = true`.
  The AC requires the test to re-verify it rather than inherit this record, so a future change to
  `classifyCodexSkillPath` turns the guard red instead of silently making it vacuous.
- **RED-now**: shares AC-CDPG-001's revert (mutant M-1, to be executed under AC-CDPG-004) — under
  `statPath = e.Path` the recorder will observe the slash form and the assertion will fail. Not yet
  observed; plan-phase.
- **Green path**: M1 step 2.

### AC-CDPG-003 — Classification precedes conversion (ordering guard; maps REQ-CDPG-003)

**Given** `configPathSeparator` pinned to `'\\'`, **When** `codexStaleSkillFinding()` judges a
fixture declaring `C:/Users/u/SKILL.md` (on darwin: `IsAbs` false, no backslash in the declared
string), **Then** the returned finding's Detail contains the relative rendering
(`"relative entr"`, from `"%d relative %s (not checked: the resolution base is not observed)"`) and
does NOT contain `"oddly-formed"`.

- **Why this is the discriminator**: the forbidden convert-first order would rewrite the string to
  `C:\Users\u\SKILL.md` before classification, which contains a backslash and therefore classifies
  `codexPathOddlyFormed` — rendering the distinct `"%d oddly-formed %s (not checked: backslash or
  ~other-user shape)"` wording instead. The two orders are observable on this host through the
  Detail string alone.
- **[HARD] Explicitly NON-discriminating, recorded as such**: a zero-stat-call assertion. Both
  branches (`codexPathRelative` and `codexPathOddlyFormed`) `continue` BEFORE reaching `osStatFn`,
  so the recorder counts 0 under the correct order and under the forbidden one alike. The test may
  keep the assertion as a supporting check; it must not be presented as the discriminator, and the
  evidence record must carry this sentence. (This corrects the card dispatch, which proposed the
  stat count as the discriminator.)
- **Cells**: green-by-construction on the ordering axis (no convert-first code exists to revert).
  Its discriminating power is demonstrated by the AC-CDPG-004 mutant M-2 below, not claimed.
- **Contrast with t562**: `AC-CSRB-003` asserted a prune-side `SkipReason`; the doctor side exposes
  no per-entry verdict object, so the Detail string is the only reachable surface.

### AC-CDPG-004 — [HARD] The mutant obligation: to be executed and recorded both ways (maps REQ-CDPG-004)

**Given** the guard of AC-CDPG-001 has landed, **When** the maintainer reverts
`internal/cli/doctor_codex.go` to `statPath = e.Path`, runs
`go test ./internal/cli/... -run '<AC-CDPG-001 selector>' -timeout 1200s -v`, then undoes the revert
and runs the same command again, **Then** the evidence record carries BOTH invocations — command
and verbatim output — showing FAIL on the reverted tree and PASS on the restored tree.

- **Mutants to execute, at minimum**:
  - **M-1** `statPath = fromConfigPath(e.Path, configPathSeparator)` → `statPath = e.Path`.
    Expected: AC-CDPG-001 and AC-CDPG-002 both FAIL.
  - **M-2** In `classifyCodexSkillPath`'s caller, convert before classifying (classify
    `fromConfigPath(e.Path, configPathSeparator)` instead of `e.Path`). Expected: AC-CDPG-003 FAILS
    on the Detail string.
  - **M-3** Declare a second `type statRecorder struct{}` in the new test file. Expected:
    AC-CDPG-005's count predicate FAILS (prints `2`), and the package test binary fails to compile.
    This is the mutant that gives AC-CDPG-005 its discriminating power; it is a test-file-only
    mutation and so does not touch `internal/cli/doctor_codex.go`.
- **[HARD] Missed mutants are recorded, not omitted.** A mutant the guard does not catch draws the
  guard's boundary and is evidence about its reach. Reporting only the caught ones would make the
  record read as complete coverage.
- **[HARD] Restoration is verified, not assumed.** After each mutant, `git diff -- internal/cli/doctor_codex.go`
  must print empty before the card proceeds. A mutant left in the tree is a production change under
  a test-only card.
- **Isolation**: mutants touch a file another lane may be compiling. Confirm the tree is not shared
  mid-mutation, and keep each mutation window as short as one run.
- **Green path**: M1 steps 1 and 5.

### AC-CDPG-005 — No helper-name collision is introduced (maps REQ-CDPG-005)

**Given** the card's tests have landed, **When**
`/usr/bin/grep -rn 'type statRecorder' internal/cli/ | wc -l` and
`/usr/bin/grep -rn 'type pruneReadbackStatRecorder' internal/cli/ | wc -l` are run, **Then** each
prints `1` — the same count as the pre-change baseline.

- **Baseline, measured this run at `a4855f0b2`**: `type statRecorder` → exactly 1
  (`internal/cli/doctor_codex_stale_skill_test.go:331`). The AC asserts the count is UNCHANGED, so a
  second declaration turns it red regardless of which file introduces it.
- **Why a count and not a build**: a duplicate declaration does break `go build`/`go vet` on the
  test binary, but that failure names a compile error rather than the policy it violates. The count
  is the cheap, directly-readable form, and it is the exact shape of the collision recorded in-tree
  at `internal/cli/codex_skills_prune_readback_test.go:37-41`.
- **Cells**: green-by-construction (the card plans to declare no recorder type at all — `plan.md`
  §B-3). Discriminating power is demonstrated by mutant **M-3** under AC-CDPG-004, not claimed.
- **Supporting, non-discriminating**: `go vet ./internal/cli/...` exits 0.

### AC-CDPG-006 → discharged by constraint, not by criterion (maps REQ-CDPG-006)

REQ-CDPG-006 (no `t.Parallel()` in a test overriding `configPathSeparator` / `osStatFn` /
`codexUserHomeDir`; every override restored via `t.Cleanup`) is **deliberately NOT given an AC**.

- **Why**: its failure mode is a nondeterministic cross-test race, not a binary predicate. A
  passing run establishes nothing about it, so any AC written over it would be a guard whose green
  is uninformative — the vacuous-green shape this card exists to remove, reintroduced one level up.
- **How it is discharged instead**: `plan.md` §D-2 states it as a [HARD] constraint, `plan.md` §E
  item 1 requires the implementer to re-read the test source and confirm it, and the Definition of
  Done below carries the checkbox "Two new tests in `internal/cli`, non-parallel, all overrides
  restored via `t.Cleanup`". That checkbox is the named carrier for REQ-CDPG-006.
- **Reviewer instruction**: verify REQ-CDPG-006 by reading the two new test functions, not by
  reading a test result.

### AC-CDPG-007 — Test-only scope held (maps REQ-CDPG-007)

**Given** the card complete, **When**
`git diff $(git merge-base origin/develop HEAD) -- internal/cli/doctor_codex.go` runs, **Then** it
prints nothing, **and** `git diff --stat $(git merge-base origin/develop HEAD)` shows changed paths
confined to `internal/cli/*_test.go` and `.moai/`.

- **Why both halves**: the first pins the specific file the mutants temporarily modify (an
  unrestored mutant is the concrete failure mode); the second catches a production change made
  anywhere else in the package, which the first would miss entirely.
- **Cells**: green-by-construction — the card plans no production edit. Its discriminating power is
  demonstrated by the AC-CDPG-004 mutants, each of which makes the first half print a diff while it
  is applied; the AC therefore also serves as the mutant-restoration check.
- **On a genuine need for a production change**: REQ-CDPG-007 makes that a blocker report to the
  lead. Silently satisfying this AC by widening the SPEC is the prohibited path.

---

## Edge cases considered

| Case | Disposition |
|---|---|
| `configPathSeparator` left at `'/'` (host default) | `fromConfigPath` is the identity and no AC here can discriminate — this is exactly t562's AC-CSRB-006 ceiling. Every AC above pins the separator; an unpinned variant is not adopted. |
| Home-relative (`~/…`) declaration | Out of scope (`spec.md` §D): t562 REQ-CSRB-002 deliberately left that branch unconverted. |
| Empty `Path` | `continue`s before classification; unaffected, not asserted. |
| A native backslash absolute (`C:\…` on darwin) | `IsAbs` false, contains a backslash → oddly-formed. Not used as a fixture: it cannot distinguish the two orders. |

## Definition of Done

- [ ] Two new tests in `internal/cli`, non-parallel, all overrides restored via `t.Cleanup`.
      **← this checkbox is the named carrier for REQ-CDPG-006 (see AC-CDPG-006 above); verify it by
      reading the test source, not a test result.**
- [ ] `go test ./internal/cli/... -timeout 1200s` passes; rc and bounded tail recorded under `.moai/reports/t570/`.
- [ ] AC-CDPG-004's mutants M-1, M-2 and M-3 executed; FAIL and PASS outputs both recorded verbatim;
      any missed mutant recorded.
- [ ] AC-CDPG-007: `git diff $(git merge-base origin/develop HEAD) -- internal/cli/doctor_codex.go`
      prints empty, and `git diff --stat` against the same base shows only `internal/cli/*_test.go`
      and `.moai/`.
- [ ] AC-CDPG-005: `/usr/bin/grep -rn 'type statRecorder' internal/cli/ | wc -l` prints `1`
      (unchanged from the `a4855f0b2` baseline).
- [ ] The AC-CDPG-003 evidence record carries the non-discriminating-stat-count sentence verbatim.

### REQ-to-carrier map (every requirement has a named carrier)

| Requirement | Carrier |
|---|---|
| REQ-CDPG-001 | AC-CDPG-001 |
| REQ-CDPG-002 | AC-CDPG-002 |
| REQ-CDPG-003 | AC-CDPG-003 |
| REQ-CDPG-004 | AC-CDPG-004 |
| REQ-CDPG-005 | AC-CDPG-005 (+ mutant M-3) |
| REQ-CDPG-006 | **No AC by design** — `plan.md` §D-2 constraint + §E item 1 + DoD checkbox 1. Rationale: AC-CDPG-006 above. |
| REQ-CDPG-007 | AC-CDPG-007 |

# Acceptance Criteria — SPEC-SIBLING-MAPS-SHORTHAND-001

Every criterion below names the command whose output decides it. The measuring
instrument in each case is a build made FROM the run-phase tree and invoked BY
PATH — written `<build>/moai` below. A PATH-resolved `moai` does not satisfy any
criterion here: a stale installed build reports a clean pass byte-identically to a
fresh one.

`<fix>` abbreviates `.moai/reports/t801/repro/fixtures`.

## §A. The expansion

- AC-SMS-001 (maps REQ-SMS-001): Given fixture A at `<fix>/SPEC-FIXA-001/`, whose sibling `acceptance.md` carries the numeric-tail shorthand and whose `spec.md` leaves nothing genuinely uncovered, When `<build>/moai spec lint <fix>/SPEC-FIXA-001` runs, Then stdout carries **zero** `CoverageIncomplete` lines and **exactly one** `MissingExclusions` line, and exit is 0. The `MissingExclusions` line is the non-empty-sweep witness: a run printing nothing at all fails this criterion rather than passing it.

  - RED-now (tree `881aa4bb8`, same command): stdout carries `WARNING CoverageIncomplete … 8  REQ REQ-FIXA-002 is not referenced by any AC` plus the `MissingExclusions` line; `0 error(s), 2 warning(s)`; exit 0. Verbatim: `.moai/reports/t801/repro/lint-fixA-before.txt`.
  - Green path: flipped by M1 (the expander joining the sibling union). The passing output is the same two-line header plus the `MissingExclusions` line alone, `0 error(s), 1 warning(s)`.

- AC-SMS-002 (maps REQ-SMS-006): Given a NEW fixture E at `<fix>/SPEC-FIXE-001/` whose `spec.md` defines `REQ-FIXE-001` and `REQ-FIXE-002` and whose `acceptance.md` reads `- AC-FIXE-001 (maps 002): Given one, When two, Then three.` — a bare tail with no full id preceding it in its own section — When `<build>/moai spec lint <fix>/SPEC-FIXE-001` runs, Then stdout carries **exactly two** `CoverageIncomplete` lines, naming `REQ-FIXE-001` and `REQ-FIXE-002`. Both stay uncovered: no prefix is inferred from the `spec.md`, from a neighbouring line, or from the AC id.

- AC-SMS-003 (maps REQ-SMS-004, 006): Given a NEW fixture F at `<fix>/SPEC-FIXF-001/` whose `spec.md` defines `REQ-FIXF-001..004` and whose `acceptance.md` carries the single line

  ```
  - AC-FIXF-001 (maps REQ-FIXF-001, 002) — verifies REQ-FIXF-003, 004 are unreachable.
  ```

  When `<build>/moai spec lint <fix>/SPEC-FIXF-001` runs, Then stdout carries **exactly two** `CoverageIncomplete` lines, naming `REQ-FIXF-003` and `REQ-FIXF-004` — and no line naming `REQ-FIXF-001` or `REQ-FIXF-002`. This is the over-reach positive control: a line-scoped implementation counts `003` and `004` as covered and emits **zero** lines, which fails this criterion. A count of two with the wrong ids also fails.

  - RED-now (tree `881aa4bb8`, same command): **three** `CoverageIncomplete` lines, naming `REQ-FIXF-002`, `REQ-FIXF-003`, `REQ-FIXF-004` — `002` is the truncation defect, `003`/`004` are the genuine gaps. Red for the right reason: the `002` line is exactly what M1 removes, and the `003`/`004` lines must survive M1 unchanged.
  - Green path: flipped by M1. Passing output drops the `002` line only.

- AC-SMS-004 (maps REQ-SMS-005): Given a NEW fixture G at `<fix>/SPEC-FIXG-001/` whose `spec.md` defines `REQ-FIXG-001` and `REQ-FIXG-002` and whose `acceptance.md` carries a `maps` list whose element run is broken by a newline — the first line ending `(maps REQ-FIXG-001,` and the second line opening `002): Given one, When two, Then three.` — When `<build>/moai spec lint <fix>/SPEC-FIXG-001` runs, Then stdout carries **exactly one** `CoverageIncomplete` line, naming `REQ-FIXG-002`. An implementation whose locator separator admits `\n` expands across the break, emits zero lines, and fails this criterion.

## §B. Reuse and immutability

- AC-SMS-005 (maps REQ-SMS-002): Given the shipped M1 code, When the two commands below are run from the repository root, Then each prints exactly `1`. Any other value from either command fails this criterion.

  ```
  grep -rn 'numericTailPattern[[:space:]]*=' internal/spec/ | grep -v _test.go | wc -l
  grep -rnE 'MustCompile\(`[^`]*,[^`]*\[0-9\]' internal/spec/ | grep -v _test.go | wc -l
  ```

  Command 1 counts **named** definitions of `numericTailPattern` in non-test package source. Its deciding value is `1`: the reused rule exists, and it was not duplicated under its own name.

  Command 2 is a **tripwire for one duplication spelling**, not exhaustive enforcement. It counts compiled patterns in non-test package source written as a **backtick raw string whose body carries a comma and a literal `[0-9]` class** — the single most likely way a second rule would be written — whatever variable name it is bound to. Its deciding value is also `1`. It is name-blind where command 1 is name-bound, which is the one thing it adds.

  **What command 2 cannot see, stated plainly because the criterion must not overclaim.** Measured against six candidate second rules by plan-audit iter3 (`.moai/reports/t801/plan-audit-iter3.md` §E2), it catches one and misses five: `(\d+)`, `([[:digit:]]+)`, a double-quoted interpreted string literal, a `MustCompile(` split across two lines, and — the important one — a **non-regexp expander** built from `strings.Split` + `strconv.Atoi`. That last is the plainest instance of the duplication REQ-SMS-002 sentence 2 forbids, and no grep for `MustCompile` can ever reach it. Broadening the regex does not fix this: it is an arms race the non-regexp form wins by construction.

  The semantic "no second rule" property is therefore carried by **AC-SMS-010 (the positive call-site assertion)** plus code review — not by this grep. Constraint C2 (`spec.md` §E) rests on AC-SMS-010 first and on these two commands only as regression tripwires.

  - RED-now / green path (tree `881aa4bb8`): both commands print `1` today — `internal/spec/lint_coverage_sibling_table.go:56` is the single existing definition. Measured by plan-audit iter2 and re-measured by the lane and by plan-audit iter3 at the same HEAD. This criterion is therefore a **regression control**, not a RED→GREEN criterion: it is green before M1 and MUST stay green after, because REQ-SMS-002 requires reuse rather than a second rule.

- AC-SMS-010 (maps REQ-SMS-002): the positive call-site assertion — what REQ-SMS-002 literally requires ("by calling it or by calling a helper extracted from it"), and the criterion no mutant in the iter3 probe survives, because introducing a second expander necessarily removes or bypasses the call this criterion names.

  Given the shipped M1 code, in which the shared expansion is reached **either** by calling `cellREQIDs` directly **or** through a helper extracted from it and named `expandNumericTails` — the name is pinned here solely so the criterion is mechanically decidable, and pins a name, not a design; REQ-SMS-002 already admits both routes — When the three commands below are run from the repository root, Then each prints the stated value. Any other value from any command fails this criterion.

  ```
  grep -cE 'cellREQIDs|expandNumericTails' internal/spec/lint_coverage_sibling_maps.go
  grep -cE 'cellREQIDs|expandNumericTails' internal/spec/lint_coverage_sibling_table.go
  go test -run 'TestSiblingMapsExpansionSharesTheTableRule' -v ./internal/spec/ 2>&1 | grep -c '^--- PASS: TestSiblingMapsExpansionSharesTheTableRule'
  ```

  Command 1 (deciding value **≥ 1**) is the load-bearing one: the new `maps`-path expander file must *name* the shared symbol. A hand-rolled duplicate — iter3's candidate F, and every other candidate B-E — fails it by construction, because a file that implements its own expansion does not call the shared one.

  Command 2 (deciding value **≥ 1**) is the other half of "shared": the table path must still reach the same symbol. If M1 extracts a helper, the table file references `expandNumericTails`; if M1 calls `cellREQIDs` directly, the table file still defines and uses `cellREQIDs`. Together commands 1 and 2 assert *one* rule reached from *both* paths, which is the property command 1 alone cannot distinguish from two coincidentally-identical rules.

  Command 3 (deciding value **exactly `1`**) closes the "invokes the helper but discards its result" gap: the named test exercises the table path and the `maps` path on the same input through the same helper and asserts identical tail expansion. **The `--- PASS:` grep is load-bearing, not cosmetic** — `go test -run` on a name matching nothing exits **0** and prints `ok … [no tests to run]`, so the exit code cannot decide this criterion and an absent test would otherwise pass it vacuously (`verification-completeness.md` §1.1, empty-sweep). The named PASS line is what makes an absent test fail.

  **What AC-SMS-010 cannot see, stated for the same reason.** A second rule that *coexists* with a genuine call to the shared helper — dead code, or a duplicate reached on some other path — satisfies all three commands. No grep and no behavioural-equivalence test distinguishes that case; two rules that behave identically on the test's inputs pass command 3 by definition. That residual is carried by code review of the M1 diff, and is named here rather than left implicit.

  - RED-now (tree `881aa4bb8`, commands verbatim): command 1 prints **nothing on stdout and exits 2** (`internal/spec/lint_coverage_sibling_maps.go: No such file or directory`) — the file M1 creates does not exist yet. Command 3 prints **`0`** (exit 1), against a `go test` run that itself printed `ok … [no tests to run]` and exited 0 — the empty-sweep the grep exists to catch. Red for the stated reason: both are red because M1's artifacts are absent, not because of any pre-existing file this work does not touch. Command 2 already prints **`3`** and is the non-empty-sweep control proving the pattern matches real source rather than matching nothing.
  - Green path: flipped by M1 (the new expander file plus the shared-rule test). Passing output is command 1 ≥ 1, command 2 ≥ 1, command 3 exactly `1`.

- AC-SMS-006 (maps REQ-SMS-003): Given the run-phase diff, When `git diff 881aa4bb8..HEAD -- internal/spec/ears.go` runs, Then it prints nothing and exits 0. The base is pinned to the literal run-phase base SHA, which is an anchor at which the measurement is taken, not a moving ref.

## §C. Controls, mutant, and corpus delta

- AC-SMS-007 (maps REQ-SMS-007): Given fixture B at `<fix>/SPEC-FIXB-001/` — the control whose sibling maps every REQ with FULL ids — When `<build>/moai spec lint <fix>/SPEC-FIXB-001` runs before and after the change, Then both runs carry **zero** `CoverageIncomplete` lines and **exactly one** `MissingExclusions` line, `0 error(s), 1 warning(s)`, exit 0. The before value is recorded at `.moai/reports/t801/repro/lint-fixB-before.txt`; the after value must be identical byte-for-byte apart from any path prefix.

- AC-SMS-008 (maps REQ-SMS-007): Given a mutant produced on a scratch copy by deleting **only** the new expander's entry from `siblingAcceptanceCoveredREQIDs`'s union in `internal/spec/lint_coverage_sibling.go` — every other file, including the expander itself and its pattern, left byte-unchanged — When the mutant is built (`go build` exit 0, so the probe is not satisfied by a compile failure) and `<build-mutant>/moai spec lint <fix>/SPEC-FIXA-001` runs, Then the false `CoverageIncomplete … REQ REQ-FIXA-002` line REAPPEARS, verbatim, alongside the `MissingExclusions` witness line. The evidence report quotes both the mutant's stdout and the non-mutant stdout from AC-SMS-001; a probe recorded as "RED" without both outputs does not discharge this criterion. The mutant is reverted and never committed.

- AC-SMS-009 (maps REQ-SMS-007): Given the whole `.moai/specs` corpus in the run-phase tree (874 SPEC directories at `881aa4bb8`, the named sample), When the same whole-corpus lint command that produced `.moai/reports/t801/repro/corpus-before-counts.txt` is re-run with the M1 build and its `CoverageIncomplete` count is taken, Then the after count **equals** the recorded before count of `2018`, and the evidence report states the command, the build's tree SHA, and both counts side by side. A delta of zero asserted without the re-run fails this criterion; a non-zero delta is a finding to report and adjudicate, not a number to absorb.

  - Named exception, declared in advance — stated as measured, and measured to be unreachable. This SPEC's own directory uses the shorthand in its prose, **including in this sentence**, so the occurrence count moves every time the SPEC is edited and no fixed number is stated here; the corpus census in `spec.md` §A.3 excludes this directory for exactly that reason. What matters is not how many there are but that at least one is live rather than illustrative: `- AC-SMS-003 (maps REQ-SMS-004, 006)` is a **live mapping of this SPEC's own requirements**, not a fixture line. The exception's operative reasoning below (0 before, 0 after, delta exactly 0) does not depend on the count. Measured by plan-audit iter2 at tree `881aa4bb8`: `<build>/moai spec lint .moai/specs/SPEC-SIBLING-MAPS-SHORTHAND-001` emits `✓ No findings`, exit 0. This directory therefore contributes **0** `CoverageIncomplete` lines before the change, and the change can only ADD covered ids — hence only REMOVE `CoverageIncomplete` lines — so its contribution is 0 after as well. **The expected delta attributable to this directory is exactly 0, and this exception cannot fire.**
    - Verdict when it fires anyway: the criterion **FAILS**. A non-zero delta — from this directory or any other — is a finding to report and adjudicate, never a number this exception absorbs.
    - Attribution instrument: diff the after-run covered-id list against the recorded before list `.moai/reports/t801/repro/corpus-before-coverage-ids.txt` (sorted (file, REQ) pairs) and name the contributing directory for every differing line. An attribution asserted without that diff does not discharge this clause.

## §D. Quality gate

- AC-SMS-GATE-001: `go test ./internal/spec/...` exits 0, with the new tests present and named in the output, and the pre-existing tests unmodified.
- AC-SMS-GATE-002: `go vet ./internal/spec/...` exits 0 and prints nothing.
- AC-SMS-GATE-003: `golangci-lint run` over the changed files exits 0 and prints nothing new against the pre-existing baseline, which is measured in the same run-phase and quoted beside it.

## §E. Definition of Done

All of §A-§D pass; the M2 evidence report exists under `.moai/reports/t801/`
carrying the fixture outputs, the mutant's RED and non-RED stdout, and the
before/after corpus counts with their commands and the build's tree SHA.

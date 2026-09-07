# t508 / SPEC-CODEX-ENABLED-FATAL-001 — run-phase evidence

Card `t508` · worktree `.claude/worktrees/t508` · branch `WT-codex-enabled-guard`
Run started at HEAD **`91f82be72`**. Commits below are the run's own.

**Tree-pin note.** The SPEC's artifacts pin every plan-time baseline to `069795602`. This run's
tree is `91f82be72`, so nothing was carried: every RED and every baseline cited here was
**re-measured in this tree**, and the two plan-time cells that could be compared
(`TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal` FAIL, its control PASS) reproduced
verbatim.

Raw command output lives under `.moai/reports/t508/evidence/`.

---

## Commits

| SHA | Milestone | Subject |
|---|---|---|
| `656624180` | M4 item 1 (authored first, by design) | `test(t508): process-level moai doctor exit-code guard for the codex enabled fatal shape` |
| `203fc0221` | M1 | `feat(t508): per-finding severity axis on the Codex Wiring check, advisory-first` |
| `876a7a7e1` | M2 | `feat(t508): parser distinguishes a declared non-boolean enabled from an absent key` |
| `7d8803749` | M3 | `feat(t508): report an unusable codex enabled key at fatal grade` |
| `d4b0edfab` | M4 items 5, 8 | `test(t508): read-only fixture guard, and the prune verb EXECUTED rather than read` |

`spec.md` moved `draft → in-progress` on the first run-phase commit (`656624180`). No other SPEC
frontmatter or body content was touched.

---

## Claim

1. A `[[skills.config]]` entry whose `enabled` key is absent, or declared with a value that is not
   a bare TOML boolean, is reported by `moai doctor` at fatal grade, and the **process** exits 1.
2. An accepting shape (`enabled = true`, `enabled = false`), and an entry declaring no `path`, stay
   non-fatal — and the check is observed to have actually READ the config in each case.
3. The severity axis was BUILT, not selected: an advisory finding still surfaces its text in a run
   that also carries a fatal one, an advisory-only run stays `CheckWarn` and exits 0, and the
   advisory grade is the severity type's zero value.
4. The parser distinguishes three `enabled` readings (absent / bare boolean / declared
   non-boolean) and writes nothing.
5. The stale-path finding's trigger, grade and message template are unchanged; only its rendered
   counts move, as a consequence of (4).
6. Nothing in this run wrote to the machine's real `~/.codex`.

---

## Evidence — AC matrix

Every `-run` row satisfies the swept-count obligation (acceptance §D.0): the cited output carries a
`--- PASS:` line per named test or sub-test, and no row's evidence is an `ok … [no tests to run]`
line.

| AC | Status | Command | Observed output |
|---|---|---|---|
| AC-CEF-001 missing key is fatal | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal' -count=1 -v` | RED at `91f82be72`: `status = ok, want CheckFail`; GREEN after M3 — `--- PASS: TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal` |
| AC-CEF-002 integer is fatal | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabled' -count=1 -v` | RED: `status = ok, want CheckFail — codex refuses \`enabled = 1\``; GREEN — `--- PASS: TestCheckCodexWiring_NonBooleanEnabled/integer` |
| AC-CEF-003 double-quoted is fatal | **PASS** | same selector, `double_quoted_true` row | RED: `status = ok, want CheckFail — codex refuses \`enabled = "true"\``; GREEN — `--- PASS: …/double_quoted_true` |
| AC-CEF-004 single-quoted is fatal | **PASS** | same selector, `single_quoted_false` row | RED: `status = ok, want CheckFail — codex refuses \`enabled = 'false'\``; GREEN — `--- PASS: …/single_quoted_false` |
| AC-CEF-005 bare `true` non-fatal (CONTROL) | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet' -count=1 -v` | positive-read half RED at `876a7a7e1` (`the detail never names the fixture's config`); GREEN — `--- PASS: …/bare_true` |
| AC-CEF-006 bare `false` non-fatal (CONTROL) | **PASS** | same selector, `bare_false` row | row did not exist at `91f82be72`; created in M4 item 4 (single-case test → table); GREEN — `--- PASS: …/bare_false` |
| AC-CEF-007 `path`-absent out of scope (SCOPE CONTROL) | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_PathAbsentIsNotFatal' -count=1 -v` | RED: `positive-read clause: the detail never names the fixture's config`; GREEN — `--- PASS: TestCheckCodexWiring_PathAbsentIsNotFatal` |
| AC-CEF-008 process exits 1 on a fatal fixture | **PASS (release-blocking, promoted in-run)** | `go test ./internal/cli/ -run 'TestDoctorExitCode_Codex' -count=1 -v` | RED at `91f82be72`, **before the severity axis landed**: `doctor_exitcode_codex_test.go:103: moai doctor exit status = 0, want 1`; GREEN — `--- PASS: TestDoctorExitCode_CodexEnabledFatal (20.86s)` |
| AC-CEF-009 process exits 0 on a clean fixture (CONTROL) | **PASS** | same selector | `--- PASS: TestDoctorExitCode_CodexCleanStaysZero (18.15s)` — green at baseline AND after, which is what a control must be |
| AC-CEF-010 reading + severity axes exist | **PASS** | `go test ./internal/codexwiring/ -count=1` + `go test ./internal/cli/ -run 'TestCheckCodexWiring_MixedFindingSurfacesAdvisoryText\|TestParseSkillEntriesEnabledThreeWayReading'` | parser half: `--- PASS: TestParseSkillEntriesEnabledThreeWayReading` (11 sub-cases); severity half RED at `203fc0221` (`status = warn, want CheckFail`), GREEN after M3 |
| AC-CEF-011 stale-path finding unchanged | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported' -count=1 -v` and `-run 'TestDoctorGolden_NoColor' -count=1 -v` | `--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.02s)`, `--- PASS: TestDoctorGolden_NoColor (0.01s)`. Exactly the two expected surfaces moved (below) |
| AC-CEF-012 nothing writes under `CODEX_HOME` | **PASS** | `go test ./internal/cli/ -count=1` (guard registered by `writeCodexHomeConfig` for every fixture) + `go test ./internal/codexwiring/ -run 'TestParseSkillEntriesWritesNothing' -count=1 -v` | no fixture-hash failure in the package run; **mutant observed**: a temporary test writing to a fixture config produced `doctor_codex_test.go:175: the fixture config changed during the run` / `--- FAIL: TestMutantWritesFixtureConfig`, so the guard is not vacuous |
| AC-CEF-013 reversed sites carry a rewritten rationale | **PASS** | `git diff 91f82be72..HEAD -- internal/codexwiring/skills.go internal/codexwiring/skills_test.go internal/cli/doctor_codex_test.go` | all eight §B.3 sites accounted for (table below); every comment site carries a replacement rationale answering BOTH grounds; no assertion site is a bare expectation flip |
| AC-CEF-014 advisory-only stays warn, exits 0 | **PASS** | `go test ./internal/cli/ -run 'TestDoctorExitCode_CodexAdvisoryOnlyStaysZero' -count=1 -v` | `--- PASS: TestDoctorExitCode_CodexAdvisoryOnlyStaysZero (17.43s)` — in-package `CheckWarn` + advisory text present, process exit 0 |
| AC-CEF-015 fatal text names the measured version | **PASS** | `go test ./internal/cli/ -run 'TestCheckCodexWiring_FatalFindingNamesMeasuredVersion' -count=1 -v` | RED: `fixture did not produce a fatal finding, so there is no text to check`; GREEN — `--- PASS: TestCheckCodexWiring_FatalFindingNamesMeasuredVersion` |
| AC-CEF-016 advisory is the severity zero value | **PASS** | `go test ./internal/cli/ -run 'TestCodexFindingZeroValueIsAdvisory' -count=1 -v` | RED at `656624180` (compile failure — `undefined: codexSeverityAdvisory`); GREEN — `--- PASS: TestCodexFindingZeroValueIsAdvisory` |

### The eight reversal sites (AC-CEF-013)

| # | Site | Kind | Disposition |
|---|---|---|---|
| 1 | `skills.go` `type SkillEnabled` doc | comment | rewritten — "tri-state" → four-state; the "no observed default" claim replaced by the measurement that codex has no default, it refuses |
| 2 | `skills.go` `SkillEnabledUnspecified` doc | comment | rewritten — the "or one whose value this parser does not recognise" clause removed and its replacement named |
| 3 | `skills.go` leniency rationale | comment | **answered in place**, not deleted — the previous comment is quoted in full and both of its grounds replied to |
| 4 | `skills.go` `skillEnabledKeyRe` | code | narrowed to a bare boolean; `skillEnabledDeclRe` added for "the key is declared at all" |
| 5 | `skills_test.go` `TestParseSkillEntriesEnabledAbsentIsUnspecified` | comment | rewritten with the measurement. **Its assertion did NOT move** — see Deviations |
| 6 | `skills_test.go` `TestParseSkillEntriesEnabledQuoted` | assertion, 5 rows | all five moved to `SkillEnabledNonBoolean`, `enabled = yes` included; comment names the measurement that reversed them |
| 7 | `doctor_codex_test.go` fixture comment | comment | `"quoted string, still true"` → `"quoted string: codex rejects it"` |
| 8 | `doctor_codex_test.go` live assertion | assertion | `(1 enabled, 0 disabled, 1 unspecified)` → `(0 enabled, 0 disabled, 2 unspecified)`, with the reason recorded above it |

### The two expected surfaces (REQ-CEF-010 as rescoped)

| Surface | Change |
|---|---|
| `internal/cli/doctor_codex.go` declared-split `fmt.Sprintf` | rendered COUNTS differ for a config containing a quoted `enabled`; the format string's literal text is byte-identical |
| `internal/cli/doctor_codex_test.go` live assertion | restated per site 8 above |

No other change to that finding's trigger, grade, or template.

---

## Baseline-attribution

Every figure below was produced in **this run**, against **this tree**.

| Dimension | Command | Observed |
|---|---|---|
| Run-start HEAD | `git rev-parse --short HEAD` | `91f82be72` on `WT-codex-enabled-guard` |
| `internal/codexwiring` suite | `go test ./internal/codexwiring/ -count=1 -timeout 300s` | `ok github.com/modu-ai/moai-adk/internal/codexwiring 0.701s` |
| `internal/codexwiring` coverage | `go test ./internal/codexwiring/ -count=1 -cover` | `coverage: 89.6% of statements` (≥85%) |
| `internal/cli` suite | `go test ./internal/cli/ -count=1 -timeout 1800s` | `ok github.com/modu-ai/moai-adk/internal/cli 549.969s`, exit 0 (`evidence/m5-cli-package.txt`) |
| vet | `go vet ./internal/cli/ ./internal/codexwiring/` | exit 0, no output |
| host build | `go build ./...` | exit 0, no output |
| windows cross-build | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0, no output |
| subagent boundary | `grep -rn 'AskUserQuestion\|mcp__askuser' <the 7 files this run touched>` | exit 1 — no matches |
| t506 extent pins | included in the `internal/codexwiring` package run above | green |
| real `~/.codex` before | `shasum -a 256 ~/.codex/config.toml` | `c91a6b73e78d057f58ab460570e6c38481632fe3e6749e0b4704fba346469598` |
| real `~/.codex` after | same command, at run close | `c91a6b73e78d057f58ab460570e6c38481632fe3e6749e0b4704fba346469598` — byte-identical |
| template neutrality | `grep -rl 'skills\.config' internal/template/templates/` | exit 1, zero matches, against the non-vacuous control `grep -rl 'moai' …` → 396 files. **No template change; `make build` not required** |

---

## Gaps — what was explicitly NOT observed

The §D.4 gaps are carried verbatim, per the acceptance document's requirement, with this run's
disposition appended.

- **Only codex-cli 0.153.4 was measured.** Whether older codex releases tolerate an absent
  `enabled`, and in which release the requirement appeared, is UNMEASURED. *Disposition: unchanged.*
  The finding text names 0.153.4 and says "observed on that release only";
  `TestCheckCodexWiring_FatalFindingNamesMeasuredVersion` additionally rejects three
  version-generalising phrasings.
- **REQ-CEF-004's class is an INDUCTION from three measured shapes.** Integer, double-quoted string
  and single-quoted string were probed; float, array, inline table, and bareword (`yes`) were not
  probed against codex. *Disposition: unchanged, and now visible in the code.* The implementation
  treats "declared but not a bare boolean" as one class, so the unprobed members are handled by
  induction rather than by measurement. `enabled = yes` is pinned in the PARSER tests as
  NonBoolean, but its codex behaviour is still inferred from the uniform `expected a boolean`
  error text, not observed.
- **Multi-entry reporting is unmeasured.** Whether codex reports the first offending entry or all
  of them was not probed. *Disposition: unchanged.* The finding counts and reports all of them,
  which affects wording only, never the verdict.
- **`enabled` inside a multi-line string, or with a trailing comment, was not probed against
  codex.** *Disposition: unchanged.* The parser's own handling of both is untouched by this run;
  the trailing-comment case is pinned on the PARSER side only (`trailing_comment` row).
- **The t506 prune verb was READ, not EXECUTED.** *Disposition: **CLOSED**.*
  `TestPruneCodexSkillEntries_DoesNotManufactureFatalShape` executes `pruneCodexSkillEntries`
  against a config carrying all four shapes and observes the unusable count not rising.
- **Card t502, the other named writer, was not checked** for whether it has landed. *Disposition:
  unchanged — not checked in this run either.*
- **REQ-CEF-009's "unchanged" baseline is pinned to `069795602`.** *Disposition: re-pinned.* This
  run's comparisons are against `91f82be72`; the two plan-time cells that could be compared
  reproduced verbatim, so the two pins agree where they overlap. Cells that could NOT be compared
  (the ten "not observable — test absent" rows) were never measurements to begin with.

Gaps arising in this run:

- **`TestCheckCodexWiring_FatalSummaryStaysInsideWidthBand` passed at RED**, because with no
  finding the Message is short. It is a width guard rather than an absence guard, so a vacuous
  pre-implementation pass is expected — but it means the width band's own RED was never observed,
  and the guard's discriminating power is unproven. It would catch a summary that grew past 113
  runes; nothing here demonstrates that it does.
- **The `internal/cli` package suite is slow enough to need `-timeout 1800s`.** An earlier run with
  the default 600s budget hit the test-binary timeout and produced a panic dump; that was a
  timeout, not a failure, and is recorded here so it is not misread later.
- **Coverage was measured for `internal/codexwiring` only.** A `-cover` run over `internal/cli`
  costs a second full-package pass and was not taken; the package's pass/fail is observed, its
  coverage percentage is not.
- **`golangci-lint` was not run.** `go vet` was, and is clean on both packages. The lint verdict is
  CI's.

---

## Residual-risk — what could still be wrong despite the above

- **The class generalisation could be wrong at its edges.** A TOML value this parser sees as
  "declared non-boolean" that codex nonetheless accepts would be reported fatal on a working
  machine — a false fatal, and `moai doctor` would exit 1 where it should not. The three measured
  shapes make this unlikely for the common cases; a float or an inline table is unprobed.
- **The fatal grade reaches CI wrappers.** Any hook or wrapper reading `moai doctor`'s exit code
  will now fail on a machine carrying an unusable `enabled`. That is intended — the machine's codex
  is already broken — but it is a real behaviour change and will surface as "doctor started
  failing" before it surfaces as "codex was already dead".
- **The read-receipt note lengthens Detail on every machine with codex skill registrations.** It
  renders only under `--verbose` and never changes a status, but it is new text in an existing
  surface.
- **The prune-side disposition change is measured on the predicate, not on the verb end-to-end.**
  `judgeCodexSkillEntry` and `pruneCodexSkillEntries` were executed; `moai clean --codex-skills`
  as a process was not.
- **`~/.codex` on this machine is entirely `enabled = false`** (49/49), so no local run exercises
  the fatal path against real data. Everything fatal here is fixture-observed.

---

## Deviations from the SPEC / plan — reported rather than worked around

1. **plan.md §B.4 says `judgeCodexSkillEntry` "reads BOTH `Enabled` and `FirstUnrecognizedLine`".
   It does not read `Enabled` at all.** Measured: `grep -n 'Enabled' internal/cli/codex_skills_prune.go`
   returns nothing. The predicate reads `FirstUnrecognizedLine` and `Path` only. The plan's
   *conclusion* still holds — the disposition change travels through `FirstUnrecognizedLine`, and
   the direction is the safe one — but the stated mechanism was half wrong, and a reader checking
   the `Enabled` path would find nothing there.

2. **plan.md §B.3 site 5 expects the ASSERTION of
   `TestParseSkillEntriesEnabledAbsentIsUnspecified` to move ("Both the assertion and its comment
   move"). Only the comment moved.** An absent key is still read as `SkillEnabledUnspecified`, and
   deliberately so: the parser reports what is DECLARED, and nothing was declared. Moving that
   assertion would have required a "the key is absent AND that is fatal" parser state, which
   duplicates in the parser a judgment REQ-CEF-003 places in the doctor. The measurement the plan
   cites (codex refuses rather than defaults) is recorded in the rewritten comment; it changes what
   the state MEANS, not what the parser reads.

3. **plan.md §B.4 predicts the quoted-`enabled` disposition moves "from eligible to preserved",
   which is true only when the entry's path is MISSING.** With a resolving path the entry was
   preserved before and after — only the skip reason changes. The test pins the missing-path case,
   where the disposition genuinely moves.

4. **AC-CEF-005/006/007's positive-read clause required a production change the requirements do
   not mention.** On a clean config the check previously emitted nothing at all, so no detail text
   could name the fixture. A read-receipt note was added to `extraDetail` to satisfy it. This is
   within the SPEC's scope envelope (it is the `enabled` sub-check's own output, advisory, and
   renders only under `--verbose`), but it is a surface the requirements did not anticipate and is
   flagged here rather than left to be discovered in review.

5. **The declared-split message folds `SkillEnabledNonBoolean` into its "unspecified" count.**
   Adding a fourth bucket would change the message template that REQ-CEF-010 preserves, so the
   count moves and the wording does not. A quoted entry is therefore described as "unspecified" in
   that advisory line while the fatal finding names it precisely. This follows the SPEC's own
   instruction not to "fix the counts back", but the resulting wording is slightly imprecise and is
   worth a follow-up decision.

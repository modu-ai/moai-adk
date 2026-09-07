# SPEC-CODEX-ENABLED-FATAL-001 — Progress

card: t508 · branch `WT-codex-enabled-guard` · base `0b1e27877` · **current HEAD `069795602`**
(post-develop-absorb; carries card t506, `SPEC-CODEX-GHOST-SKILLS-PRUNE-001`)

## §E.1 Plan-phase Audit-Ready Signal

### Revision 2 — plan-audit FAIL repair (2026-09-07)

Plan-audit returned **FAIL 0.69** against the Tier S threshold 0.75, with nine blocking findings.
The operator chose **tier correction plus repair**, not scope-splitting. This revision addresses
D1-D12 and D15; D14 and D16 were withdrawn by the auditor after `moai spec lint` returned zero
findings for this SPEC.

- **Tier: S → M (D3).** A CORRECTION of a mis-declaration, recorded in spec.md HISTORY with its
  ground, and one that **RAISES** the bar: the PASS threshold moves 0.75 → 0.80. Ground: three
  layers (parser fidelity, a new severity axis, a detection surface), a reversal of a test-pinned
  decision across eight sites, and a change to `moai doctor`'s process exit-code contract.
- **Budget**: 14 requirements and 16 acceptance criteria against the Tier M ceilings of 16 / 16.
  The AC count is **at ceiling** — a further criterion requires either a Tier L correction or the
  merger of two existing ones.
- **Requirements**: 14 (REQ-CEF-001..014), GEARS notation. Added this revision: REQ-CEF-013
  (the reversal carries a replacement rationale) and REQ-CEF-014 (advisory is the severity zero
  value). Rescoped: REQ-CEF-010, from "wording" to "message template", because the previous wording
  forbade what M2 requires.
- **Acceptance criteria**: 16 (AC-CEF-001..016), each Given-When-Then with a named verify command
  and a swept-count obligation. Added this revision: AC-CEF-014 (advisory-only run stays warn and
  exits 0), AC-CEF-015 (finding text names the measured codex version), AC-CEF-016 (severity zero
  value is advisory). Re-mapped: AC-CEF-013, from REQ-CEF-011 to REQ-CEF-013.
- **SPEC ID check**: executed as Bash in revision 1, verbatim output `PASS`. The ID is unchanged in
  this revision.
- **Citations restamped against `069795602` (D12)**: fifteen `file:line` claims re-read in this
  tree; three moved (`skills.go` 32→34, ~60→125, `doctor.go` 137→142) and three sites the draft did
  not carry at all were added (`skills.go:29`, `doctor_codex_test.go:302`, `doctor_codex.go:757`).
  Table: research.md §9.

### RED-now evidence, measured at `069795602`

Executed in this plan-phase session, not carried from the earlier baseline:

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal' -count=1 -v
=== RUN   TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal
    doctor_codex_enabled_test.go:37: status = ok, want CheckFail — codex cannot start on this config: {...}
--- FAIL: TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.774s        exit 1

$ go test ./internal/cli/ -run 'TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet' -count=1 -v
--- PASS: TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.697s        exit 0

$ go test ./internal/cli/ -run 'TestCheckCodexWiring_StaleHomeSkillsReported' -count=1 -v
--- PASS: TestCheckCodexWiring_StaleHomeSkillsReported (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.745s        exit 0
```

**The absorb did not close the defect**: the mutant still fails and its control still passes on the
post-absorb HEAD.

### The zero-match measurement that motivated §D.0 (D1)

```
$ go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabled' -count=1 -v
testing: warning: no tests to run
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	0.673s [no tests to run]        exit 0
```

Six criteria of the previous draft named tests that do not exist and were therefore satisfiable
having run nothing. acceptance.md §D.0 is the clause that closes it; every `-run` criterion now
carries a swept-count obligation, and §D.5 restates it as a Definition-of-Done item.

### Severity re-classification (D2)

Per `verification-completeness.md` §2.1, a criterion with no observed RED is not release-blocking.
Only three criteria have observed evidence at `069795602` — AC-CEF-001 (RED), AC-CEF-005 (control
baseline), AC-CEF-011 (baseline) — so twelve criteria previously marked MUST-PASS are re-classified
as **regression guards** (acceptance.md §D.2), each naming the milestone that authors its test.

The consequence is stated rather than buried: the exit-code contract (AC-CEF-008) is NOT
release-blocking on plan-time evidence. plan.md §F M4 item 1 orders the process-level test FIRST so
its RED is observed against pre-M1 behaviour and the criterion can be promoted within the run.

### Other findings addressed

- **D4** — the false traceability row (REQ-CEF-011 → AC-CEF-013, different subjects) is closed by
  the new AC-CEF-015; §D.3 now records a subject per row, and spec.md §F notes that `moai spec
  lint`'s `CoverageIncomplete` counts references and cannot detect this.
- **D5** — the leniency comment is quoted in FULL (both grounds) in spec.md §B.1 and research.md §3;
  §B.1's singular "a decision whose stated ground turned out to be false" is corrected, and ground 2
  ("the declared intent is unambiguous") is answered: codex never reads the intent.
- **D6** — the reversal blast radius is enumerated at **eight** sites in three files (plan.md §B.3),
  with the three assertion sites distinguished from the five comment sites, including the five-row
  quoted table where the `enabled = yes` row also changes.
- **D7** — **13** `problems = append(` sites cited by line; the advisory grade is required to be the
  enum's zero value (REQ-CEF-014, AC-CEF-016); AC-CEF-014 pins a non-`enabled` warn finding to
  `CheckWarn` + exit 0; M1's deferred design question is answered in plan.md rather than left to a
  commit message.
- **D8** — AC-CEF-008's Verify names the process boundary and lists forbidden satisfying forms,
  citing `exitcode_contract_test.go:30,38` and `binary_lag_test.go:102` as the in-package pattern
  already in the tree.
- **D9** — REQ-CEF-010 rescoped to "message template"; the declared-split COUNT change is declared
  expected at exactly `doctor_codex.go:757` and `doctor_codex_test.go:302`.
- **D10** — the three shape controls gain a positive-read clause (the detail text must mention the
  fixture's entry), so a codex-not-in-play skip cannot satisfy them.
- **D11** — AC-CEF-006's selector is corrected: the named test is single-case at `069795602`, so M4
  converts it into a table and the criterion's swept-count is a named sub-test row.
- **D15** — REQ-CEF-004's generalisation is stated as an explicit induction from three measured
  shapes; REQ-CEF-009's "unchanged" baseline is pinned to `069795602`.

### Carried context

- **Evidence basis**: `.moai/reports/t508/codex-enabled-lab.md`, `red-baseline.md`, and (new this
  revision) `develop-t506-impact.md`.
- **t506 fold-in**: the priority argument moves from *latent* to *live* — a writing surface has
  landed that cannot create the fatal shape but will rewrite a config codex cannot load and report
  success. Scope does not widen to t506's code. Recorded as a Gap that the prune verb was READ, not
  executed; plan.md §F M4 item 8 proposes a run-phase fixture rather than inheriting the reading.
- **New surface**: `SkillEntry` gained `StartLine` / `EndLine` / `FirstUnrecognizedLine`, and
  `judgeCodexSkillEntry` is a second consumer of `Enabled` and `FirstUnrecognizedLine` under an
  `@MX:WARN` deletion guard. plan.md §B.4 makes checking against it a run-phase obligation, and
  `skills_extent_test.go` must stay green.
- **Template-First verdict**: unchanged — no template change expected. `grep -rl 'skills\.config'
  internal/template/templates/` exits 1 (zero matches) against a non-vacuous control
  (`grep -rl 'moai'` → 396 files); no `.go` files under templates.
- **Operator decisions**: B.1 (scope) and B.2 (severity) recorded with attribution to the
  2026-09-07 lane-3 session. Not re-opened. The tier correction and repair-vs-split choice were
  taken by the operator on the plan-audit FAIL.
- **Open clarifications**: none. Unmeasured items are recorded as Gaps (acceptance.md §D.4), not as
  `[NEEDS CLARIFICATION]` markers — none of them changes what gets built.

### Gaps in this plan-phase revision

- No Go source was modified in this revision; all measurements are read-only or test-selector runs.
- The three RED-now / baseline runs above are the only executed evidence. Thirteen criteria have no
  observed RED and say so per-criterion (acceptance.md §D.1) rather than being asserted green.
- ~~`moai spec lint` was not re-run after this revision's edits~~ — **now measured; see the
  subsection below.** The auditor's UNOBSERVED reading was a timing race (its output file was still
  0 bytes at every check), not a missing run.

### `moai spec lint` against the revised artifacts — measured at `069795602`

**Result.** Two independent invocations, both completed against the revised artifacts in this tree,
agree:

```
0 error(s), 4346 warning(s)
lint rc=0
grep -c 'ENABLED-FATAL' → 0
```

Zero findings name this SPEC. The auditor correctly reported the lint as UNOBSERVED and made no
zero-finding claim; this subsection supplies the measurement it could not take.

**Non-vacuity control — the engine reached this directory and had nothing to say.** A zero is only
evidence if the instrument was live. The same run emitted 4,346 warnings against OTHER SPECs,
including `CoverageIncomplete` and `SpecsDirMissingSpecFile` rows — the very rule families that
would fire on this SPEC's frontmatter, its REQ↔AC coverage, and its file set were any of them
violated. The zero is therefore a measured silence, not an unreached directory.

**[HARD] A clean lint is NOT verified traceability — the limitation is structural, not temporary.**
`CoverageIncomplete` counts whether a REQ is *referenced* by some AC. It cannot read whether the
referenced AC is about the **same subject**. The proof is in this SPEC's own history: **D4 — the
false `REQ-CEF-011 → AC-CEF-013` mapping — passed this same lint in the previous draft**, while
REQ-CEF-011 was uncovered and AC-CEF-013 orphaned. The engine has no capacity to detect that class
of defect and will not acquire one by being re-run.

The §D.3 **Subject** column therefore carries an obligation the linter is structurally incapable of
discharging. A future reader must not read `0 error(s)` as "traceability verified": the subject
check is a reader-level obligation (spec.md §F, acceptance.md §D.3), and this lint result attests
nothing about it.

_Status: draft. Awaiting re-audit against the Tier M threshold (0.80) and Implementation Kickoff
Approval._

## §E.2 Run-phase Evidence

Run started at HEAD **`91f82be72`**. Full evidence — commands, verbatim output, the eight-site
reversal table, the carried §D.4 gaps, and five recorded deviations from the plan — is at
`.moai/reports/t508/run-evidence.md`; raw command output under `.moai/reports/t508/evidence/`.

**Tree-pin note.** The plan artifacts pin every baseline to `069795602`; this run's tree is
`91f82be72`, so nothing was carried. Every RED and baseline below was re-measured here, and the two
plan-time cells that could be compared (the RED guard's FAIL and its control's PASS) reproduced
verbatim.

### Commits

| SHA | Milestone | Subject |
|---|---|---|
| `656624180` | M4 item 1 (authored first, by design) | process-level `moai doctor` exit-code guard |
| `203fc0221` | M1 | per-finding severity axis, advisory-first |
| `876a7a7e1` | M2 | parser distinguishes a declared non-boolean `enabled` |
| `7d8803749` | M3 | fatal-grade reporting of an unusable `enabled` key |
| `d4b0edfab` | M4 items 5, 8 | read-only fixture guard; prune verb executed |

### AC matrix

Every `-run` row satisfies the swept-count obligation (acceptance §D.0): a `--- PASS:` line per
named test or sub-test, and no row's evidence is an `ok … [no tests to run]` line.

| AC | Status | Verification command | Actual output |
|---|---|---|---|
| AC-CEF-001 | PASS | `go test ./internal/cli/ -run 'TestCheckCodexWiring_MissingEnabledKeyIsReportedFatal' -count=1 -v` | RED `status = ok, want CheckFail` → `--- PASS:` after M3 |
| AC-CEF-002 | PASS | `go test ./internal/cli/ -run 'TestCheckCodexWiring_NonBooleanEnabled' -count=1 -v` | RED `codex refuses \`enabled = 1\`` → `--- PASS: …/integer` |
| AC-CEF-003 | PASS | same selector | RED → `--- PASS: …/double_quoted_true` |
| AC-CEF-004 | PASS | same selector | RED → `--- PASS: …/single_quoted_false` |
| AC-CEF-005 (CONTROL) | PASS | `go test ./internal/cli/ -run 'TestCheckCodexWiring_DeclaredEnabledKeyStaysQuiet' -count=1 -v` | positive-read half RED → `--- PASS: …/bare_true` |
| AC-CEF-006 (CONTROL) | PASS | same selector | row created in M4 item 4 → `--- PASS: …/bare_false` |
| AC-CEF-007 (SCOPE CONTROL) | PASS | `go test ./internal/cli/ -run 'TestCheckCodexWiring_PathAbsentIsNotFatal' -count=1 -v` | RED → `--- PASS:` |
| AC-CEF-008 | PASS — **promoted to release-blocking in-run** | `go test ./internal/cli/ -run 'TestDoctorExitCode_Codex' -count=1 -v` | RED at `91f82be72`, pre-severity: `moai doctor exit status = 0, want 1` → `--- PASS: TestDoctorExitCode_CodexEnabledFatal (20.86s)` |
| AC-CEF-009 (CONTROL) | PASS | same selector | `--- PASS: TestDoctorExitCode_CodexCleanStaysZero (18.15s)` |
| AC-CEF-010 | PASS | `go test ./internal/codexwiring/ -count=1` + the mixed-finding selector | parser half `--- PASS: TestParseSkillEntriesEnabledThreeWayReading` (11 sub-cases); severity half RED → GREEN |
| AC-CEF-011 | PASS | `-run 'TestCheckCodexWiring_StaleHomeSkillsReported'` and `-run 'TestDoctorGolden_NoColor'` | both `--- PASS:`; only the two expected surfaces moved |
| AC-CEF-012 | PASS | package run + `-run 'TestParseSkillEntriesWritesNothing'` | no fixture-hash failure; **mutant observed** — a temporary writer produced `the fixture config changed during the run` / `--- FAIL:` |
| AC-CEF-013 | PASS | `git diff 91f82be72..HEAD -- internal/codexwiring/skills.go internal/codexwiring/skills_test.go internal/cli/doctor_codex_test.go` | all eight sites accounted for; every comment site answers BOTH grounds; no bare expectation flip |
| AC-CEF-014 | PASS | `-run 'TestDoctorExitCode_CodexAdvisoryOnlyStaysZero' -v` | `--- PASS: (17.43s)` — `CheckWarn`, advisory text present, exit 0 |
| AC-CEF-015 | PASS | `-run 'TestCheckCodexWiring_FatalFindingNamesMeasuredVersion' -v` | RED → `--- PASS:` |
| AC-CEF-016 | PASS | `-run 'TestCodexFindingZeroValueIsAdvisory' -v` | RED at `656624180` (`undefined: codexSeverityAdvisory`) → `--- PASS:` |

### Invariants

| Invariant | Command | Observed |
|---|---|---|
| Parser writes nothing (REQ-CEF-002) | `go test ./internal/codexwiring/ -run 'TestParseSkillEntriesWritesNothing' -count=1 -v` | `--- PASS:` — whole fixture directory hashed |
| No config repair (REQ-CEF-012) | `go test ./internal/cli/ -run 'TestPruneCodexSkillEntries_LeavesNoFileBehind' -count=1 -v` | `--- PASS:` |
| Real `~/.codex` untouched | `shasum -a 256 ~/.codex/config.toml`, before and after | `c91a6b73…69598` both times — byte-identical |
| t506 extent pins green | `go test ./internal/codexwiring/ -count=1` | `ok … 0.701s` (whole package) |
| No template change | `grep -rl 'skills\.config' internal/template/templates/` | exit 1, zero matches, control `grep -rl 'moai'` → 396 files |
| Full package suites | `go test ./internal/cli/ -count=1 -timeout 1800s` · `go test ./internal/codexwiring/ -count=1` | `ok … internal/cli 549.969s` exit 0 · `ok … internal/codexwiring 0.701s` |

### Deviations reported (not worked around)

Five, in full at `.moai/reports/t508/run-evidence.md` § Deviations. In brief: plan §B.4's claim
that `judgeCodexSkillEntry` reads `Enabled` is false (measured — zero matches); plan §B.3 site 5's
assertion correctly did NOT move; the quoted-`enabled` disposition moves only on a MISSING path;
AC-CEF-005/006/007's positive-read clause required a production read-receipt note the requirements
do not mention; and the declared-split message folds the new state into its "unspecified" count to
preserve the template REQ-CEF-010 protects.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-09-07
run_commit_sha: d4b0edfab
run_status: audit-ready
ac_pass_count: 16
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a — no push in this run (lane-local; the branch is unpushed by design)
l44_post_push_fetch: n/a — no push in this run
new_warnings_or_lints_introduced: 0 (go vet clean on both packages; golangci-lint not run — CI's verdict)
cross_platform_build:
  host: "go build ./... — exit 0"
  windows: "GOOS=windows GOARCH=amd64 go build ./... — exit 0"
total_run_phase_files: 10  # source files; measured `git diff --name-only 91f82be72..HEAD | grep -v '^.moai/' | wc -l`
m1_to_mN_commit_strategy: "one commit per milestone, M4 item 1 first so its RED preceded the severity axis"
```

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: 2026-09-07
sync_commit_sha: pending-backfill-sync
sync_status: audit-ready
b12_self_test_a: "grep -c 'CODEX-ENABLED-FATAL' CHANGELOG.md → 0 (pre-emission; no duplicate entry from a parallel BATCH-SYNC session)"
b12_self_test_b: "grep -oE 'AC-CEF-[0-9]+' .moai/specs/SPEC-CODEX-ENABLED-FATAL-001/acceptance.md | sort -u | wc -l → 16; CHANGELOG entry cites '16 acceptance criteria (AC-CEF-001..016), 16 PASS / 0 FAIL' — count matches"
b12_self_test_c: "every path claimed in the CHANGELOG entry verified via ls: .moai/specs/SPEC-CODEX-ENABLED-FATAL-001/spec.md, .moai/reports/t508/codex-enabled-lab.md, .moai/reports/t508/run-evidence.md — all exist"
changelog_entry_position: "CHANGELOG.md ### Fixed section, first entry (top of section, immediately after the '### Fixed' heading)"
frontmatter_status_transitions.spec_md: "in-progress -> completed (this commit; status: + updated: only, per manager-docs' allowed frontmatter scope)"
frontmatter_status_transitions.plan_md: "no status field (stateless on that axis, unchanged)"
frontmatter_status_transitions.acceptance_md: "no status field (stateless on that axis, unchanged)"
canary_compliance_check: n/a — this SPEC defines no forward-looking policy that its own sync tests
```

### CHANGELOG emission (§B12 self-test detail)

1. **Pre-emission grep** — `grep -c 'CODEX-ENABLED-FATAL' CHANGELOG.md` returned `0` before this
   commit's edit, confirmed by re-running the same command after the edit and observing exactly
   `1` (the entry this commit added). No duplicate-entry risk from a parallel BATCH-SYNC session.
2. **AC count match** — `acceptance.md` (SSOT, not `progress.md`) carries 16 distinct
   `AC-CEF-###` identifiers, all live (none is `[RETIRED]`/`[REF]`-marked). The CHANGELOG entry's
   text `16 acceptance criteria (AC-CEF-001..016), 16 PASS / 0 FAIL` states the same count, matching
   §E.3's `ac_pass_count: 16 / ac_fail_count: 0`.
3. **File path verification** — every file path named in the CHANGELOG entry (`spec.md`,
   `.moai/reports/t508/codex-enabled-lab.md`, `.moai/reports/t508/run-evidence.md`) exists,
   verified via `ls`.

### README / docs-site finding (not a silent scope decision)

- **README** — `grep -n "Codex Wiring" README.md` (all 4 locale files) → no match. README's doctor
  coverage is limited to the Home Disk Usage feature (line 145/734 of `README.md`); it never
  documents individual `moai doctor` checks at this granularity. No README edit was made — there
  is no existing surface this change extends.
- **docs-site** — `docs-site/content/en/cli-reference/doctor.md` DOES document individual doctor
  checks as dedicated callout sections (`## Home Disk Usage check {{< new-badge v3.1.1 >}}`,
  `## Hook Delivery check {{< new-badge v3.1.4 >}}`) — this is a real, non-vacuous documentation
  surface (control: `grep -n "Home Disk Usage" docs-site/content/en/cli-reference/doctor.md` → 1
  hit). `grep -rn "Codex Wiring" docs-site/content/en/` → 0 matches: the Codex Wiring check (this
  change included) has never had a docs-site callout, in any of the 4 locales. Adding one now would
  be a genuine docs-site addition, not a repair, and per `docs-site-i18n-rules.md` the 4-locale
  same-PR obligation applies — that is a larger unit of work than this sync task's scope. **This is
  reported as a blocker/finding for the orchestrator to route (a follow-up card), not done as a
  partial single-locale edit.**

### Gaps carried from run-phase (verbatim, per acceptance.md §D.4)

- Only codex-cli 0.153.4 was measured; whether older codex releases tolerate an absent `enabled`
  is unmeasured.
- REQ-CEF-004's "not a bare TOML boolean" class is an induction from three measured shapes
  (integer, double-quoted string, single-quoted string); float/array/inline-table/bareword forms
  were not probed.
- Multi-entry reporting order (first offending entry vs all) is unmeasured; no AC depends on it.
- `enabled` inside a multi-line string or with a trailing comment was not probed against codex.
- The t506 prune verb (`SPEC-CODEX-GHOST-SKILLS-PRUNE-001`) was EXECUTED in this run (§E.2 AC-CEF-012
  row cites `TestPruneCodexSkillEntries_LeavesNoFileBehind` PASS), closing the plan-phase gap that
  it had been read but not executed.
- Card t502, the other named writing surface, was not checked for whether it has landed.


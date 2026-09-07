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

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

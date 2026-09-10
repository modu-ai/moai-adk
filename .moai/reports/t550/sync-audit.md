# t550 — sync-phase audit (SPEC-GATE-OXLINT-DETECT-001, GH #1631)

card: t550
branch: WT-js-linter-detect
audited tip: 5b4007a93 (base d060e0d13 = local develop at dispatch)
auditor: sync-auditor, lane-5, independent of the implementing and sync agents
audited: 2026-09-10

**VERDICT: PASS** — no blocking finding. Six findings, all `optional`.

---

## Claim

The nine acceptance criteria hold on the audited tip as an *independent* observation, not
as a reading of `progress.md §E.2`. The oxlint entry is genuinely config-gated, the three
mutant probes each failed on the criterion they were written for, AC-004's set-equality
assertion compares two independently-authored sources (so it is not a tautology), AC-003
exercises four distinct paths rather than one loop reporting four names, and the
linter-free control is demonstrably capable of going red. Scope is confined to the gate
entry plus one line of `javascript.md` in both copies.

---

## Evidence

### 1. AC conformance — re-run, not read

```
$ go test ./internal/hook/quality/... -run 'Oxlint|oxlint' -v -count=1
--- PASS: TestNodeLintRunsOxlintStepOnOxlintProject (0.17s)          [AC-001]
--- PASS: TestNodeLintOxlintViolationFailsGate (0.34s)               [AC-002]
--- PASS: TestNodeLintOxlintConfigFilenames (0.53s)                  [AC-003]
    --- PASS: .../oxlint_config/.oxlintrc.json (0.14s)
    --- PASS: .../oxlint_config/.oxlintrc.jsonc (0.13s)
    --- PASS: .../oxlint_config/oxlint.config.ts (0.13s)
    --- PASS: .../oxlint_config/oxlint.config.mts (0.13s)
--- PASS: TestNodeLintOxlintStepShape (0.00s)                        [AC-004]
--- PASS: TestNodeLintEslintProjectUnaffectedByOxlint (0.12s)        [AC-005]
--- PASS: TestNodeLintBiomeProjectUnaffectedByOxlint (0.12s)         [AC-006]
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	1.985s
```

The `Oxlint|oxlint` selector returns six of the seven tests — **AC-007's test name carries
no oxlint token** — so it was run separately:

```
$ go test ./internal/hook/quality/... -run 'TestNodeLintLinterFreeScaffoldStillPasses' -v -count=1
--- PASS: TestNodeLintLinterFreeScaffoldStillPasses (0.00s)          [AC-007]
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	0.361s
```

```
$ go test ./internal/hook/quality/... -count=1                        [AC-008]
ok  	github.com/modu-ai/moai-adk/internal/hook/quality	16.081s   EXIT=0

$ go vet ./internal/hook/quality/...
(no output)  vet-exit=0

$ golangci-lint run ./internal/hook/quality/...
0 issues.
```

```
$ diff .claude/rules/moai/languages/javascript.md \
       internal/template/templates/.claude/rules/moai/languages/javascript.md
(no output)                                                           [AC-009]
$ shasum <both>
e1d4a63a55764e55911856fd9bc19887253b6208  .claude/rules/moai/languages/javascript.md
e1d4a63a55764e55911856fd9bc19887253b6208  internal/template/templates/.claude/rules/moai/languages/javascript.md
```

**Divergence against the §E.2 matrix: none in substance.** Every row I could re-execute
reproduced the claimed outcome. The only numeric differences are wall-clock timings
(§E.2 `18.271s` vs my `16.081s`), which measure the machine, not the code (F6).

### 2. Vacuous-pass re-check, on the code

- **AC-004 is not a tautology.** The expected set lives in the *test* file
  (`gate_oxlint_lint_test.go:29-34`, `oxlintConfigFilenames`); the checked set lives in
  the *production* table (`internal/hook/quality/gate.go:236-238`). Two independently
  authored literals, compared in both directions plus a length check. Confirmed live by
  the probes: M2 reddens the "missing" direction, M3 the "unexpected" direction.
- **AC-003 exercises four paths, not one.** Each iteration builds its own fixture with
  `filename` as the config file and asserts the oxlint row `executed`. M2 (list shrunk to
  `.oxlintrc.json`) reddened exactly the other three subtests while the first stayed
  green — a loop that reported four names while exercising one path could not produce
  that split.
- **Could the suite pass while detection fails?** No. AC-001/002 read the run summary of
  a real `QualityGate.Run` against a fixture directory, with `cmd.Dir` set to the project
  dir (`gate.go:1347`), so the `executed … npx oxlint` row is a real execution record.

### 3. Mutant-probe integrity (`mutant-m{1,2,3}.txt`)

Each recorded failure is an assertion message with a file:line that resolves in the
*current* test file — not a compile error, not a fixture failure:

```
$ sed -n '157p;206p;211p;216p;246p;248p;305p;310p' internal/hook/quality/gate_oxlint_lint_test.go
157: t.Errorf("config file %s: oxlint outcome = %q (reason %q), want %q",
206: t.Errorf("oxlint configFiles has %d entries, want exactly %d: %v",
211: t.Errorf("missing config file name: %s (declared: %v)", want, oxlint.configFiles)
216: t.Errorf("unexpected config file name: %s (oxlint does not read it; declared: %v)",
246: assertConfigAbsentSkip(t, g, "oxlint", out)
248: t.Errorf("executed lint steps = %d, want 1; summary:\n%s", n, out)
305: t.Errorf("executed lint steps = %d on a linter-free scaffold, want 0; summary:\n%s", n, out)
310: assertConfigAbsentSkip(t, g, label, out)
```

| Probe | Intended criterion | Observed failure | Judgement |
|---|---|---|---|
| M1 (`configFiles: []`) | the entry must be config-gated | AC-005/006/007 red with `oxlint: executed … npx oxlint` inside eslint/biome/bare fixtures; AC-004 red on `len 0 ≠ 4` | failed on the intended criterion; the run summary quoted in the output shows the oxlint row actually executing |
| M2 (`[".oxlintrc.json"]`) | each of the four names must select the step on its own | AC-003 subtests 2-4 red with reason `none of its config files exist … (.oxlintrc.json)`; AC-004 red on the missing side ×3 | failed on the intended criterion |
| M3 (four names + `oxlint.config.js`) | REQ-002's closed norm — the F1 repair | **every behavioural AC stayed green**, AC-004 alone red: `has 5 entries, want exactly 4` + `unexpected config file name: oxlint.config.js` | demonstrates the equality repair rather than asserting it; a containment check would have passed M3 |

`git status --porcelain` at audit time printed nothing, so the mutants were reverted.

### 4. Control integrity — AC-007 can go red

Two independent grounds, neither of which requires me to write to the tree:

1. **Observed red under M1**: `executed lint steps = 1 on a linter-free scaffold, want 0`,
   with the summary showing `oxlint: executed in 136ms`.
2. **The counter is demonstrably live**: `executedLintCount` returns `1` in AC-005 and
   AC-006 (both assert against `1`, and both pass), so its `0` in AC-007 is a real
   measurement over a non-empty step list — not an empty-set artifact.

AC-007 additionally asserts three named config-absent notices (up from two), so it also
pins the notice mechanism it consumes.

### 5. Evidence-claim integrity

- **The CHANGELOG's "eslint and biome unchanged" claim, flagged by the sync agent as
  measured one commit before the tip, does not matter.** The code it speaks about did not
  move:

  ```
  $ git diff --stat d608847c2..HEAD -- internal/hook/quality \
      .claude/rules/moai/languages/javascript.md \
      internal/template/templates/.claude/rules/moai/languages/javascript.md CHANGELOG.md
   CHANGELOG.md | 1 +
   1 file changed, 1 insertion(+)
  ```

  The only delta between the measured tree and the tip is the CHANGELOG bullet itself.
  Independently, I re-ran AC-005 and AC-006 **at the tip** and both pass. The claim is
  attributable.
- `.moai/reports/t550/remeasure-after.txt` carries all seven fixtures with the exact rows
  the §E.2 and `verdict.md` tables quote (`oxlint: executed … npx oxlint` on the four
  oxlint fixtures, `skipped` on eslint/biome/bare) — the tables are transcriptions of a
  persisted run, not summaries.
- `CHANGELOG.md:12` is the first `### Fixed` bullet, as `§E.4 changelog_entry_position`
  states.
- `acceptance.md` carries exactly `AC-001 … AC-009`, matching the CHANGELOG's "9
  acceptance criteria".
- `spec.md` frontmatter: `status: completed`; `§E.4 sync_commit_sha: e597a9ac1` is
  backfilled by `ff858eb72`, matching the declared D3 window.

### 6. Scope

- The full-branch diffstat touches production code in exactly two places:
  `internal/hook/quality/gate.go` (+12, one table entry) and
  `internal/hook/quality/gate_oxlint_lint_test.go` (new). Everything else is
  `javascript.md` ×2 (1 line each), `CHANGELOG.md` (+1), and SPEC/report artifacts.
- The two `javascript.md` copies are byte-identical (sha above).
- `spec.md §5` carries **both** required clauses, unedited: *Out of Scope — the
  zero-config oxlint project* (lines 185-197) and *Out of Scope — co-resident linter
  configs* (lines 198-211).
- No drive-by refactor of the eslint or biome entries; no other toolchain touched.

---

## Baseline-attribution

Every figure above was produced in this run, in this worktree
(`/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t550`), against tip `5b4007a93` on branch
`WT-js-linter-detect`, with `git status --porcelain` empty. The `d060e0d13` BEFORE column
is cited from `baseline.md` and is used only as the comparison baseline, never as a fresh
measurement. The mutant judgements are attributed to the recorded outputs plus the line
resolution shown in §3 — they are the lane's measurements, corroborated, not re-executed
by me (see Gaps).

---

## Gaps — explicitly NOT observed

- **I did not re-inject the mutants.** The dispatch restricts me to writing this report,
  and mutating a tracked file in a tree with live sibling agents is a write I declined to
  make. My M1/M2/M3 judgement rests on the recorded outputs plus file:line resolution
  against the current test file, not on my own re-execution.
- **`go test ./...` and `internal/cli` were not run** (dispatch [HARD]). Cross-package
  regression and the full-suite verdict remain CI's.
- **No non-darwin platform.** The six execution-bearing tests `t.Skip` on Windows; only
  AC-004 runs everywhere. The `GOOS=windows` build was not exercised.
- **No real `npx oxlint`.** Every fixture shadows `npx` with an exit-code stub, so step
  *selection* and outcome propagation are measured; oxlint's own analysis is not.
- **The zero-config oxlint case was not tested** — out of scope by ruling, and no test
  claims it.
- **`make build` / the embedded-template axis was not re-verified by me.** `verdict.md`
  records `strings bin/moai | grep "Linting: ESLint"` hitting the new sentence; I read
  that record but did not rebuild.
- **Co-resident config projects** (two linters' configs at once) were not measured; the
  SPEC declares the case unspecified with pre-existing behaviour preserved.

---

## Residual-risk

- **The four-name reference set can only be as right as its citation.** AC-004 measures
  the implementation against the test file's literal; neither measures oxlint upstream. An
  upstream fifth discovery filename would be actively *rejected* by AC-004, and the
  correct response is a REQ-002 amendment — the SPEC says so, but nothing mechanical
  enforces that it is not instead "fixed" by loosening the assertion.
- **`npx oxlint` carries no path argument**, relying on oxlint defaulting to the current
  directory. The stub cannot confirm that default; `cmd.Dir` is correct, so the risk is
  entirely upstream behaviour drift.
- **`npx <tool>` may fetch from the network** on a project that does not have the tool
  installed. This is the pre-existing eslint/biome pattern, inherited rather than
  introduced — but the oxlint entry widens the set of projects for which it can happen
  (F5).
- **A co-resident-config project now runs up to three lint steps.** Declared unspecified
  rather than wrong, on stated grounds; if it turns out to bother users, the evidence will
  arrive as a bug report rather than from this card.

---

## Findings

- **F1** [Low] [optional] `internal/hook/quality/gate_oxlint_lint_test.go` — AC-007's test
  is named `TestNodeLintLinterFreeScaffoldStillPasses`, the only one of the seven carrying
  no `oxlint` token. Any name-scoped sweep of "the oxlint tests" (`-run 'oxlint|Oxlint'`,
  as the dispatch itself specified) silently omits the **control** — the one test whose
  omission would hide a regression into "run every linter everywhere". Required fix
  (optional): rename to e.g. `TestNodeLintOxlintLinterFreeScaffoldStillPasses`, or note in
  the file header that the control must be run by package, never by name selector.
- **F2** [Low] [optional] `gate_oxlint_lint_test.go:54` — `containsString` hand-rolls
  `slices.Contains`, which the standard library provides and which this very package
  already uses (`internal/hook/quality/formatter.go:113`). Simplicity-ladder step 3.
  Required fix (optional): delete `containsString` and call `slices.Contains`.
- **F3** [Low] [optional] `gate_oxlint_lint_test.go:41-50` — `oxlintGate` is a verbatim
  copy of `biomeGate` (`gate_biome_lint_test.go:30-39`), same package, five identical
  assignments. Required fix (optional): call `biomeGate`, or rename it to a neutral
  `lintOnlyGate` shared by both files.
- **F4** [Info] [optional] `.moai/reports/t550/remeasure.sh` — the persisted recipe
  hardcodes an absolute worktree path and a session-specific scratchpad path, so no other
  actor can re-run it. The outputs it produced are persisted, so nothing is lost; the
  recipe is documentation rather than a reproducer. Required fix (optional): derive both
  from `git rev-parse --show-toplevel` and `mktemp -d`.
- **F5** [Low] [optional] `internal/hook/quality/gate.go:236` — `npx oxlint` on a project
  carrying an oxlint config but no installed oxlint will make `npx` fetch it. Pre-existing
  behaviour of the eslint and biome siblings, not introduced here, so it is out of this
  card's scope; recorded so it is not later "discovered" as a regression of this change.
- **F6** [Info] [optional] `progress.md §E.2` quality-gate row cites `ok … 18.271s`; I
  measured `16.081s` on the same tree. Wall-clock only — it measures the machine, not the
  code, and carries no verdict. No action.

An all-`optional` findings list does not convert a PASS into a FAIL.

---

## Dimension scores

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 96/100 | **PASS** | 9/9 ACs re-executed or re-derived by this auditor; `go test ./internal/hook/quality/... -count=1` → `ok … 16.081s`, exit 0; the four-fixture oxlint execution is confirmed end-to-end in `remeasure-after.txt` against a binary built from this tree |
| Security (25%) | 92/100 | **PASS** | No new trust boundary. The step runs under the existing `executeStep` path with `cmd.Dir` = project dir (`gate.go:1347`) and the package's env scrub (`step_git_env.go`); the only new external invocation is `npx oxlint`, inheriting the eslint/biome `npx` pattern verbatim (F5). No secret, no user input, no network reachable from the tests (`npx`/`npm`/`node` stubbed) |
| Craft (20%) | 88/100 | **PASS** | Seven tests + four subtests, all mutation-validated; three probes each red on their own criterion; RED observed before the edit (`red-pre-edit.txt` shows `no oxlint row … has no oxlint lint step`). Deductions: F1 (control invisible to a name-scoped sweep), F2 and F3 (stdlib and helper duplication) |
| Consistency (15%) | 90/100 | **PASS** | The entry mirrors its biome sibling field-for-field (`optional`, `configFiles`, `npx` binary) with a comment matching the file's density; Template-First honoured on `javascript.md` (both copies byte-identical, sha `e1d4a63a5…`); `golangci-lint run` → `0 issues.`; `go vet` exit 0. Deduction: `npx oxlint` omits the `.` path argument its two siblings carry — justified in the comment, but it is a local shape divergence |

Must-pass firewall (Functionality, Security): both PASS independently.

**Overall verdict: PASS.**

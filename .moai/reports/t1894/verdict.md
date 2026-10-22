# Run-phase verdict — card t1894, SPEC-CARD-MODE-JUDGMENT-001

Author: manager-develop (documents-only run; no Go source changed). Date: 2026-10-23.
Overall: **BLOCKED at M5 on AC-MJ-014.** M1-M3 are committed and every read-only criterion is green; one Go guard (`TestBindingLedgerIntegrity`) is red for a reason the plan could not see. Nothing was pushed, amended, or merged.

Tree measured: HEAD `770b5ae654` on `WT-fail-safe-serial`, `git status --short` empty (0 lines), base `28c7bb8d36`.
Run commits: M1 `2fb0be5ed8`, M2 `89eb51de2f`, M3 `770b5ae654`.

Path legend used below: `T:` = `internal/template/templates/.claude/` + `rules/moai/workflow/` (or `skills/…`), `D:` = the deployed twin under `.claude/`.

## 1. Claim

| # | Claim | Status |
|---|---|---|
| C1 | AC-MJ-001..013, -016..020 pass on HEAD `770b5ae654` (read-only cells of `acceptance.md`) | PASS |
| C2 | Each milestone commit carries template and deployed twins together (`pairs`) | PASS |
| C3 | The closed-world manifest is fully applied and nothing else changed | PASS |
| C4 | AC-MJ-014: the scoped Go guards stay green | **FAIL** — 1 test red (`TestBindingLedgerIntegrity`); 13 of 14 template guards and 6 of 6 `internal/cli` guards pass |
| C5 | AC-MJ-015 (lane acknowledgement) | not mine — recorded by the leader |

## 2. Evidence

Commands are the single-line cells of `.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/acceptance.md` §2 (ids `E-xxx`), executed unmodified with `sh -c`, `/usr/bin/grep` first on PATH. Exit status is the one the cell expects (a zero-hit `grep -c` exits 1).

### 2.1 RED before M1 (E8 equivalent for a documents-only run), HEAD `28c7bb8d36`, clean tree

- `E-001` printed `0` and `0`, exit 1.
- `E-006` printed `0 0 0 1` twice, exit 1. `E-007` printed `0 0 0` twice, exit 1. `E-009` printed `0 1 0` twice, exit 1.
- `E-002a` printed the `gtd.md:52` line, exit 1. `E-020a` first row `C template 0 2 0 0 0 deployed 0 2 0 0 0`, exit 1.
- Every other RED-now cell of the ledger reproduced its recorded value.

### 2.2 Acceptance matrix on HEAD `770b5ae654` (full ledger re-run on the committed tree)

| AC | Cells | Verbatim decisive output | Status |
|---|---|---|---|
| 001 | E-001 | `8` / `8`, exit 0 | PASS |
| 002 | E-002a, b, c, d | a: empty, exit 0; b: the base-commit line (control, exit 0); c: the six paths; d: `:1` / `:1` | PASS |
| 003 | E-003 | `T:factory-dispatch.md:1`, `D:factory-dispatch.md:1` | PASS |
| 004 | E-004a-e | a `:1`/`:1`; b `:0`/`:0`; c `:2`/`:2`; d `:1`/`:1`; e `:2`/`:2` | PASS |
| 005 | E-005a-d | a `:1`; b `:0`; c `:1`; d `:3` (both trees) | PASS |
| 006 | E-006 | `8 8 1 1` / `8 8 1 1`, exit 0 | PASS |
| 007 | E-007 | `4 4 1` / `4 4 1`, exit 0 | PASS |
| 008 | E-008a-f | a empty exit 0; b `core:1 factory-dispatch.md:2 auto-semantics.md:2 SKILL.md:1` (both trees); c `core:1`, `factory-dispatch.md:1`; d `auto-semantics.md:1`; e `2`/`2`; f `1`/`1` | PASS |
| 009 | E-009 | `1 1 0` / `1 1 0`, exit 0 | PASS |
| 010 | E-010a, b | a empty (exit 1 = no file listed); b `:1` / `:1` | PASS |
| 011 | E-011 (seven `cmp`) | all seven exit 0, empty output | PASS |
| 012 | E-012a | `OK 9997`, exit 0 (was `OK 9999`) | PASS |
| 013 | E-013a | `:0` on all eight template files | PASS |
| 016 | E-016 | `factory-dispatch-cards.md`, `factory-dispatch-detail.md`, `factory-dispatch-mechanics.md` (exactly the three homes) | PASS |
| 017 | E-017 | ` 1 file changed, 1 insertion(+), 1 deletion(-)` | PASS |
| 018 | E-018a-e | a `:1`/`:1`; b `:0` x4; c `:1` x4; d `:1`/`:1`; e `:1`/`:1` | PASS |
| 019 | E-019a, c | `:1` / `:1` each | PASS |
| 020 | E-020a, b, d | all sixteen rows `0 0 0 0 0`, exit 0; `unexpected-paths 0`; `same same` | PASS |
| 014 | Go guards (2.3) | `FAIL github.com/modu-ai/moai-adk/internal/template` | **FAIL** |

`E-020a` after M3, verbatim (template side; the deployed side prints the same row values):
`C 0 0 0 0 0`, `K 0 0 0 0 0`, `F 0 0 0 0 0`, `D 0 0 0 0 0`, `M 0 0 0 0 0`, `A 0 0 0 0 0`, `G 0 0 0 0 0`, `W 0 0 0 0 0`.

### 2.3 Commit-graph witnesses (AC-MJ-011 run-phase obligation)

- `closed-world-check.py pairs … 2fb0be5ed8:3` printed `2fb0be5ed8 pairs 3 unpaired 0 expected 3`, exit 0.
- `closed-world-check.py pairs … 89eb51de2f:5` printed `89eb51de2f pairs 5 unpaired 0 expected 5`, exit 0.
- `closed-world-check.py pairs … 770b5ae654:4` printed `770b5ae654 pairs 4 unpaired 0 expected 4`, exit 0.
- The three commits also carry the catalog hashes and ledger rows of what they change: `moai` (M1, M3), `moai-lane-watchdog` plus `ALB-0276` and `ALB-0257` (M2).

### 2.4 Go guards (AC-MJ-014), one invocation per package, `go test … -count=1 -v`

`internal/template`, exit 1. Top-level results: PASS for `TestDeployedAlwaysLoadedCharBudget`, `TestTemplateNoInternalContentLeak`, `TestLaneRecheckDoctrine`, `TestLaneWatchdogRenewsFirst`, `TestLaneWatchdogRecordsRefusalAndMeasures`, `TestFactoryDispatchBodyCarriesRenewIntervalLinkage`, `TestRoleCoreStubGate`, `TestReviewOwnership_CardReviewDoctrine`, `TestWorkflowRulePathsPinned`, `TestDeclaredRuleMirrorForks`, `TestRuleTemplateMirrorDrift`, `TestCatalogHashCoversSkillSubfiles`, `TestCardIDBaselineHasNoStaleEntries`; **FAIL** `TestBindingLedgerIntegrity (4.60s)`. Package line: `FAIL	github.com/modu-ai/moai-adk/internal/template	5.809s`. The deployed-surface log line for the stub reads `…/factory-dispatch-core.md 9997`.

The failing assertion, verbatim:
`binding_ledger_test.go:1140: ledger integrity: AC-ALB-021(2): binding fragment of row ALB-0257 appears in on-demand file .claude/rules/moai/workflow/factory-dispatch.md: "A reconciliation is named, not excepted: `/moai:todo --auto`"`
and the sweep line: `[AC-ALB-021(2)] fragment sweep: binding/normative qualifying lines=560 … violations=1 lines exempt (already in anchor copy)=22`.

`internal/cli`, exit 0: PASS for `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`, `TestAutoRankMarkerDisclosure`, `TestAutoPickDocDoctrine`, `TestAutoPickMirrorParity`, `TestTodoSkillDocumentsClassification`; package line `ok  	github.com/modu-ai/moai-adk/internal/cli	1.098s`.

Empty-sweep check: both logs contain 0 occurrences of `[no tests to run]` or `[no test files]`; the `internal/cli` log shows 179 `=== RUN` lines; all six named `internal/cli` tests and all fourteen named `internal/template` tests appear with a PASS/FAIL line.

### 2.5 Diagnosis of the red guard, and the throwaway experiment

Cause (read from `internal/template/binding_ledger_test.go`, `checkAC021FragmentRule` / `sweepFragmentLines`, lines 995-1119): every line of 40 or more UTF-16 units in a binding or normative row's `after_text` must appear as no line of an on-demand (`paths:`-scoped) rule file, unless that exact line already existed in that file's copy at the ledger's anchor commit (`a2a184ad3`). Row `ALB-0257` is the `--auto` reconciliation paragraph; its `after_text` is the K stub line (always-loaded). The plan has F carry the identical new paragraph (manifest F `+` row, byte-identical to the K row). F is on-demand, and its anchor copy holds only the old wording, so the exemption that used to cover the identical pair no longer covers it.

Experiment (throwaway, uncommitted, restored): appended one sentence to F's reconciliation line only, in both trees, so it no longer equals the K line. Result: `TestBindingLedgerIntegrity`, `TestRoleCoreStubGate`, `TestWorkflowRulePathsPinned`, `TestDeclaredRuleMirrorForks`, `TestRuleTemplateMirrorDrift` all PASS (`go exit=0`, `violations=0`), and the six `internal/cli` guards all PASS. Restoration check: `git status --short` empty and `git diff --quiet` exit 0.

## 3. Baseline-attribution

- Every cell above ran in this run against the committed tree `770b5ae654` (clean, 0 status lines); the pre-M1 values ran against `28c7bb8d36`. Per-milestone intermediate runs (after each apply, before each commit) agreed with the final run.
- Judging build: the Go results come from `go test` compiled from this tree; `gen-catalog-hashes.go` ran through `go run` from this tree; `closed-world-check.py` is the tree's own script. No installed `moai` build produced any cited measurement; the installed `moai` was used only for `moai slot` lease operations, so the tool-provenance second coordinate is not applicable beyond this statement.
- Go invocations ran under the `go-heavy` lease (acquired and released around each use, load 4.7-10.1 at acquisition, below the 32 limit).

## 4. Gaps (not observed)

- AC-MJ-015 (lane acknowledgement) is operational and not mine.
- `go build` for Windows, coverage, and lint were not run: no Go source changed, and the plan scopes M5 to the two `go test -run` invocations.
- The 13 other `internal/template` guards were observed on the tree that fails `TestBindingLedgerIntegrity`; their results are for that tree, and the only variant tree measured (the experiment) was reverted and is not what is committed.
- The full text of the `go test -v` logs and the intermediate ledger runs stayed in scratch outside the repository; only the lines above were carried here.
- Refused tool call: the worktree guard refused my first ledger-run invocation because it passed a shell variable as the script path (message: "a computed argument goes after the script or --"). I re-issued the same command with literal paths; no substitution of the measurement occurred.

## 5. Residual-risk

- The proposed amendment (section 6) was observed green on a scratch variant, not on a committed manifest, so the exact wording a manager-spec amendment chooses still needs its own run.
- Scratch outputs (`red-now`, per-milestone and `final` ledger runs, both Go logs) are machine-local and are named here as a known loss; they are not cited as evidence.
- Sync-audit residual stands: the manifest wording itself is read by the audit, not by any command.

## 6. Blocker report (E7)

**Blocker.** AC-MJ-014 cannot go green under the frozen manifest: `TestBindingLedgerIntegrity` fails AC-ALB-021(2) for row `ALB-0257` because manifest rows `K +` (reconciliation paragraph) and `F +` (the same paragraph, byte-identical) put one 40-plus-unit line of a binding row into an on-demand file with no anchor-copy exemption.

**Not mine to change.** The `F +` reconciliation row of `closed-world-manifest.txt`, `plan.md` §B M2 step 3, `design.md` §D-3 row S19 and REQ-MJ-017 (b) all state the clause edit as identical in F and K. Cells E-018b and E-018c would be unaffected by the proposal below (they read the opening clause only).

**Minimal proposed amendment for manager-spec (observed green on a scratch variant):** make F's reconciliation line differ from K's by one added trailing sentence, for example `The one further stated case is in `factory-dispatch-detail.md` § Delegated-dispatch promotion.`, and change the manifest `F +` row for the reconciliation paragraph (and the matching words in plan.md M2 step 3, design.md §D-3 S19, REQ-MJ-017 (b)) accordingly. Keep the opening `A reconciliation is named, not excepted:` so E-018b/E-018c stay as they are. Option not tested: set `ALB-0257.after_text` to a sub-line fragment of the K paragraph so no whole line is shared; this changes what the ledger row covers and needs the ledger owner's view.

**Work done and tree state.** M1, M2 and M3 are committed and green on every read-only cell; the tree is clean at `770b5ae654`. The remaining work is the manager-spec amendment, then one corrective edit of the F reconciliation line in both trees, then re-running `TestBindingLedgerIntegrity` and the manifest checks. Inference, not observed: the ledger rows and catalog hashes should not need to change, because only a rule file (F) changes and `ALB-0257.after_text` is the K line.

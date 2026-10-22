# Run-phase verdict — card t1894, SPEC-CARD-MODE-JUDGMENT-001

Author: manager-develop (documents-only run; no Go source changed). Date: 2026-10-23.
Overall: **PASS on every AC this phase owns** (AC-MJ-001..014, -016..020). The M5 blocker raised earlier is resolved (section 6). AC-MJ-015 is the leader's.

Tree measured: HEAD `05bcdc3a8f` on `WT-fail-safe-serial`, `git status --short` empty (0 lines), base `28c7bb8d36`. The evidence commit that carries this file follows `05bcdc3a8f` and touches only this file and `progress.md`.

Run commits:

| Commit | Content |
|---|---|
| `2fb0be5ed8` | M1 mode judgment and explicit issuance |
| `89eb51de2f` | M2 executing-card-only lease renewal and scoped cron |
| `770b5ae654` | M3 free lane and delegated-dispatch promotion |
| `0be1353bbd` | first evidence commit (M5 blocker record) |
| `8b9f610b91` | mid-run amendment by manager-spec, committed by the coordinator (manifest F row) |
| `05bcdc3a8f` | correction: F reconciliation line carries the home pointer |

Path legend: `T:` = `internal/template/templates/.claude/` + `rules/moai/workflow/` (or `skills/…`), `D:` = the deployed twin under `.claude/`.

## 1. Claim

| # | Claim | Status |
|---|---|---|
| C1 | AC-MJ-001..013 and -016..020 pass on HEAD `05bcdc3a8f` (read-only cells of `acceptance.md`) | PASS |
| C2 | Every document commit carries template and deployed twins together (`pairs`) | PASS |
| C3 | The closed-world manifest (amended, `8b9f610b91`) is fully applied and nothing else changed | PASS |
| C4 | AC-MJ-014: the scoped Go guards are green | PASS — 14 of 14 `internal/template`, 6 of 6 `internal/cli` |
| C5 | AC-MJ-015 (lane acknowledgement) | not mine — recorded by the leader |

## 2. Evidence

Commands are the single-line cells of `.moai/specs/SPEC-CARD-MODE-JUDGMENT-001/acceptance.md` §2 (ids `E-xxx`), executed unmodified with `sh -c`, `/usr/bin/grep` first on PATH. Exit status is the one the cell expects (a zero-hit `grep -c` exits 1).

### 2.1 RED before M1 (documents-only run, so this stands in for E8), HEAD `28c7bb8d36`, clean tree

- `E-001` printed `0` and `0`, exit 1.
- `E-006` printed `0 0 0 1` twice, exit 1. `E-007` printed `0 0 0` twice, exit 1. `E-009` printed `0 1 0` twice, exit 1.
- `E-002a` printed the `gtd.md:52` line, exit 1. `E-020a` first row `C template 0 2 0 0 0 deployed 0 2 0 0 0`, exit 1.
- Every other RED-now cell of the ledger reproduced its recorded value.

### 2.2 Acceptance matrix on HEAD `05bcdc3a8f` (full ledger run on the committed tree; identical to the run made just before the commit)

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
| 012 | E-012a | `OK 9997`, exit 0 (was `OK 9999`; unchanged by the correction) | PASS |
| 013 | E-013a | `:0` on all eight template files | PASS |
| 014 | Go guards (2.4) | `ok  	github.com/modu-ai/moai-adk/internal/template	6.155s`; `ok  	github.com/modu-ai/moai-adk/internal/cli	0.926s` | PASS |
| 016 | E-016 | `factory-dispatch-cards.md`, `factory-dispatch-detail.md`, `factory-dispatch-mechanics.md` (exactly the three homes) | PASS |
| 017 | E-017 | ` 1 file changed, 1 insertion(+), 1 deletion(-)` | PASS |
| 018 | E-018a-e | a `:1`/`:1`; b `:0` x4; c `:1` x4; d `:1`/`:1`; e `:1`/`:1` | PASS |
| 019 | E-019a, c | `:1` / `:1` each | PASS |
| 020 | E-020a, b, d, e, h | a all sixteen rows `0 0 0 0 0`, exit 0; b `unexpected-paths 0`; d `same same`; e (control) `DIFFERENT same`, exit 1; h (control) `C good 0 0 pair-move 44 44`, exit 1 | PASS |

`E-020a` after the correction, verbatim: `C template 0 0 0 0 0 deployed 0 0 0 0 0`, and the same `0 0 0 0 0` on both sides for `K`, `F`, `D`, `M`, `A`, `G`, `W`; exit 0.

### 2.3 Commit-graph witnesses (AC-MJ-011 run-phase obligation)

- `pairs … 2fb0be5ed8:3` printed `2fb0be5ed8 pairs 3 unpaired 0 expected 3`, exit 0.
- `pairs … 89eb51de2f:5` printed `89eb51de2f pairs 5 unpaired 0 expected 5`, exit 0.
- `pairs … 770b5ae654:4` printed `770b5ae654 pairs 4 unpaired 0 expected 4`, exit 0.
- `pairs … 05bcdc3a8f:1` printed `05bcdc3a8f pairs 1 unpaired 0 expected 1`, exit 0 (the correction commit changes the F pair only).
- The commits carry the catalog hashes and ledger rows of what they change: `moai` (M1, M3), `moai-lane-watchdog`, `ALB-0276` and `ALB-0257` (M2). The correction commit changes no catalog or ledger file.

### 2.4 Go guards (AC-MJ-014), one invocation per package, `go test … -count=1 -v`, run on the tree that `05bcdc3a8f` commits

`internal/template`, exit 0. Top-level lines, verbatim: `--- PASS: TestDeployedAlwaysLoadedCharBudget (0.07s)`, `--- PASS: TestBindingLedgerIntegrity (4.95s)`, `--- PASS: TestTemplateNoInternalContentLeak (0.71s)`, `--- PASS: TestLaneRecheckDoctrine (0.00s)`, `--- PASS: TestLaneWatchdogRenewsFirst (0.00s)`, `--- PASS: TestLaneWatchdogRecordsRefusalAndMeasures (0.00s)`, `--- PASS: TestFactoryDispatchBodyCarriesRenewIntervalLinkage (0.00s)`, `--- PASS: TestRoleCoreStubGate (0.00s)`, `--- PASS: TestReviewOwnership_CardReviewDoctrine (0.00s)`, `--- PASS: TestWorkflowRulePathsPinned (0.00s)`, `--- PASS: TestDeclaredRuleMirrorForks (0.00s)`, `--- PASS: TestRuleTemplateMirrorDrift (0.00s)`, `--- PASS: TestCatalogHashCoversSkillSubfiles (0.01s)`, `--- PASS: TestCardIDBaselineHasNoStaleEntries (0.02s)`, then `PASS` and `ok  	github.com/modu-ai/moai-adk/internal/template	6.155s`.

The check that had failed now reads: `[AC-ALB-021(2)] fragment sweep: binding/normative qualifying lines=560 on-demand rule files=100 .claude skill files=248 .agents skill files=17 violations=0 lines exempt (already in anchor copy)=22`.

`internal/cli`, exit 0. Top-level lines, verbatim: `--- PASS: TestAutoRankDoctrineAmendment (0.00s)`, `--- PASS: TestAutoRankMirrorParity (0.00s)`, `--- PASS: TestAutoRankMarkerDisclosure (0.00s)`, `--- PASS: TestAutoPickDocDoctrine (0.01s)`, `--- PASS: TestAutoPickMirrorParity (0.00s)`, `--- PASS: TestTodoSkillDocumentsClassification (0.00s)`, then `PASS` and `ok  	github.com/modu-ai/moai-adk/internal/cli	0.926s`.

Empty-sweep check: both logs contain 0 occurrences of `[no tests to run]` and of `[no test files]`; the `internal/cli` log shows 179 `=== RUN` lines; all 14 named template tests and all 6 named cli tests appear with a `--- PASS` line.

### 2.5 History of the earlier red (kept for the audit trail)

Before the amendment, on HEAD `770b5ae654`, `internal/template` exited 1: `--- FAIL: TestBindingLedgerIntegrity (4.60s)` with `binding_ledger_test.go:1140: ledger integrity: AC-ALB-021(2): binding fragment of row ALB-0257 appears in on-demand file .claude/rules/moai/workflow/factory-dispatch.md: "A reconciliation is named, not excepted: `/moai:todo --auto`"` and `violations=1`. Cause: the F reconciliation line was byte-identical to the K stub line that row `ALB-0257` covers, and F's anchor copy held only the old wording, so the anchor-copy exemption did not apply. A throwaway variant (one trailing sentence on F's line, reverted afterwards with `git diff --quiet` exit 0) had shown `violations=0` before the amendment was requested; the committed correction below reproduces that result.

## 3. Baseline-attribution

- Final matrix: this run, tree = committed HEAD `05bcdc3a8f`, `git status --short` empty. Pre-M1 values: HEAD `28c7bb8d36`. Intermediate per-milestone runs agreed with the final run.
- Go: `go test` compiled from this tree; catalog hashes through `go run` from this tree; `closed-world-check.py` is the tree's own script. No installed `moai` build produced any cited measurement; the installed `moai` served only `moai slot` lease operations, so no tool-provenance lag statement is owed.
- Go invocations ran under the `go-heavy` lease, acquired and released around each use; the one-minute load average read before each acquisition was 4.42 (correction run); earlier acquisitions read 4.74, 4.84 and 6.94, and the pre-flight read 10.09 — every reading under the 32 limit.

## 4. Gaps (not observed)

- AC-MJ-015 is operational and not mine.
- Windows build, coverage and lint were not run: no Go source changed, and the plan scopes M5 to the two `go test -run` invocations.
- Only the 14 + 6 named tests were run; the rest of both packages and `go test ./...` were deliberately not run (CI owns the repository-wide verdict, which is PENDING at report time).
- Refused tool calls: (1) the worktree guard refused my first ledger-run invocation, which passed a shell variable as the script path; (2) it refused a compound command that combined `git status`, a process substitution and two `diff` operands. I re-issued each as plain commands with literal paths and no measurement was substituted.
- Raw `go test -v` logs and intermediate ledger runs stayed in scratch outside the repository; only the lines above were carried here.

## 5. Residual-risk

- The manifest wording (amended row included) is read by the sync audit, not by any command.
- Heads-up for the sync phase, measured in this run: `acceptance.md` now contains 21 distinct `AC-…` identifiers (`grep -o -E 'AC-[A-Z]+-[0-9]+' … | sort -u | wc -l` printed 21) because the added edge case names `AC-ALB-021`; the SPEC's own count is still 20 (`AC-MJ-001..020`, the same command restricted to `AC-MJ-` printed 20). `acceptance.md` was not edited by this phase.
- Scratch outputs are machine-local and are a known loss; they are not cited as evidence.

## 6. Blocker report (E7) — RESOLVED

Earlier blocker: AC-MJ-014 could not go green under the frozen manifest because `TestBindingLedgerIntegrity` failed AC-ALB-021(2) on row `ALB-0257` (section 2.5).

Resolution: manager-spec amended the manifest F row under mid-run authority (`8b9f610b91`): F's reconciliation line ends with one extra pointer sentence and so differs from K's line. The correction `05bcdc3a8f` sets that line to the manifest row in both trees (line 48 in each), touching nothing else: the K stub, the ledger rows and `catalog.yaml` are unchanged, as expected for a rule file. Observed after the correction: `lines` `0 0 0 0 0` on all sixteen rows, E-012a `OK 9997`, `TestBindingLedgerIntegrity` PASS with `violations=0`, 14 of 14 and 6 of 6 guards PASS. No open blocker remains.

# SPEC-AUTONOMY-BATCH-GATE-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

plan_status: audit-ready
plan_complete_at: 2026-10-02
plan_audit_verdict: iteration 1 = FAIL (0.70, 2026-10-02, report `.moai/reports/t1344/plan-audit-iter1.md`, local-only); iteration 2 = FAIL (0.775, 2026-10-02, report `.moai/reports/t1344/plan-audit-iter2.md`, local-only); iteration 3 = FAIL (0.84 against the Tier L PASS threshold 0.85, 2026-10-02, report `.moai/reports/t1344/plan-audit-iter3.md`, local-only, audited_sha 4dec6281c). Iteration 3 reached the Tier L ceiling of 3 plan-audit iterations. The operator approved one further author revision (0.4.1: defects N16, N17, N18, N26 only) plus an auditor delta read — a ceiling extension decided by the operator, to be informed to the leader in the card completion report. The delta read has not yet run
plan_artifact_hash: pending (computed by the orchestrator after the last plan-phase edit; progress.md, spec-compact.md, and decision-index.md are not in the hashed set; design.md and research.md are, at Tier L)
tier: L (orchestrator ruling R1, 2026-10-02; counting rule and recount command in spec.md §A.5; 17 planned files)
artifacts: spec.md, plan.md, acceptance.md, design.md, research.md, spec-compact.md, decision-index.md, progress.md
open_decisions: decision-index.md Q1-Q3, Q6-Q10, Q12 and Q13 open (Q8 open and out of scope); Q4 and Q5 decided in scope (operator, Decision Point 1, 2026-10-02); Q11 POLICY-COVERED, recorded as the orchestrator's Tier L ruling
amended: version 0.4.1 on 2026-10-02 (narrow revision for plan-audit iteration 3 defects N16, N17, N18, N26; N19-N25 and N27 stay named debts) — 20 REQ, 17 AC, 45 guard anchors unchanged; version 0.4.0 (iteration 2 disposition N1-N15 and rulings R1-R4, plan.md §J) is commit 4dec6281c
recorded_by: manager-spec (card t1344); 0.4.1 edits authored over HEAD 4dec6281c on branch WT-batch-approval-gate; the commit that carries them is the orchestrator's record, not this line

## §E.2 Run-phase Evidence

Owner: manager-develop role (cycle_type tdd), milestones M1 and M2, card t1344. Started on branch WT-batch-approval-gate at tree 61801e166.

### M1 — guard test, RED first

New file `internal/template/batch_gate_summary_doctrine_test.go` (test only), top-level test `TestBatchGateSummaryDoctrine`. It cuts `### 9.2 The batch gate summary` out of the template copy of `auto-semantics.md` through `EmbeddedTemplates()` and checks the 45 anchors A01..A45 of plan.md section F. Subtests: `real_section` (logs `real_section violations=<n>`) and one subtest per anchor, each running a mutant of the real section and requiring that exactly that anchor is reported (logs `mutant <ID> rejected: reported=<ID>`; a mutant that does not change the body fails the subtest).

RED run (tree 61801e166, section absent):

- Command: `go test -count=1 -v -run '^TestBatchGateSummaryDoctrine$' ./internal/template/` with output redirected to a scratch file; exit code 1 (observed as its own step).
- Decisive lines, verbatim:

```text
=== RUN   TestBatchGateSummaryDoctrine
    batch_gate_summary_doctrine_test.go:246: section "### 9.2 The batch gate summary" not found in the template copy of .claude/rules/moai/workflow/auto-semantics.md
--- FAIL: TestBatchGateSummaryDoctrine (0.00s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/template	0.403s
```

- Classification per tdd-result-contract: EXPECTED_RED. The file compiled (a failure at line 246 is the intended assertion `section not found`, reached at run time), so this is not TOOL_FAILURE; it is not REGRESSION_FAILURE because the test is new.
- Provenance note: the matchers were first validated for coherence against a draft of the section placed temporarily in the template copy (47 `--- PASS` including the top level, 45 mutant lines), then that draft was removed and the RED above was re-run on the final test file before this commit. The first RED run (same decisive failure) preceded any draft.

M1 commit: 5936aeb32 (test file, spec.md frontmatter status draft to in-progress, this section).

### M2 — canonical section, GREEN

Section `### 9.2 The batch gate summary` written first in the template copy of `auto-semantics.md`, then the identical bytes copied to the live copy (`cmp` of the two files: exit 0). Also one pointer sentence at the end of 9.1 and one sentence in section 10. `make build` exit 0, regenerating no tracked file (git status afterwards listed only the two auto-semantics copies). Section size 5986 bytes (the file is `paths:`-scoped, not always-loaded; `TestAlwaysLoadedTokenBudget` headroom unchanged at 13373 tokens). Measured at tree 5936aeb32 plus the uncommitted M2 edits.

Guard test, `go test -count=1 -v -run '^TestBatchGateSummaryDoctrine$' ./internal/template/` redirected to a file, exit 0, ending `ok  	github.com/modu-ai/moai-adk/internal/template	0.493s`. AC-007 five readings, each a separate `grep -c` or `awk` on that file:

- `--- PASS` count 47 (one top level, `real_section`, 45 anchors); `--- FAIL` count 0.
- `=== RUN` count 47, equal to the `--- PASS` count.
- `[no tests to run]` count 0.
- Mutant lines `mutant <ID> rejected: reported=<ID>` with equal IDs: 45 (`awk '$2=="mutant" && $4=="rejected:" && $5=="reported="$3'`), and 45 lines contain `rejected: reported=`.
- `real_section violations=0` count 1.

AC-001: `grep -c -F '### 9.2 The batch gate summary'` prints 1 on the live copy and 1 on the template copy. Before the edit `git log --format=%h -S"### 9.2 The batch gate summary" -- internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` printed nothing.

Other commands, each its own run, exit 0:

- `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/`: `--- PASS: TestRuleTemplateMirrorDrift` (10 `--- PASS` lines, 10 `=== RUN`).
- `go test -count=1 -v -run '^TestTemplateNoInternalContentLeak$' ./internal/template/`: `--- PASS: TestTemplateNoInternalContentLeak`.
- `go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/`: `always-loaded surface = 64227 tokens (budget 77600, headroom 13373, 16 entries)` and `--- PASS: TestAlwaysLoadedTokenBudget`.
- `go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity)$' ./internal/cli/`: both `--- PASS`.
- `go test -count=1 -v -run '^(TestSubSkillLOCCeiling|TestEntryRouterLOCCeiling)$' ./internal/skills/`: both `--- PASS`.
- `go vet ./internal/template/`: exit 0.

AC-016 (a) pre-check: `grep -rlF "counter_refs=" ...` and the same for `searched=` over the rule, skill and hook sources (test files excluded) each print exactly the two auto-semantics copies.

### M4 — stale Kickoff wording alignment (milestones M4 and M5 run by a second manager-develop role spawn, card t1344)

Measured at tree 4e0a60bc1 plus the uncommitted M4 edits. Edited exactly two documents, each in the template copy first and then the live copy: `workflows/plan/spec-assembly.md` (7 lines rewritten as 7 lines) and `workflows/moai.md` (the pipeline gate 2 line and the Step 11.3 line, one line each; the two copies were edited separately with the same `old_string`, never copied onto each other).

AC-014 readings (each a separate command, run on both the live and the template path):

- `grep -c "stays MANDATORY"` and `grep -c "does NOT substitute for the gate"` on both `spec-assembly.md` copies: `0`, `0`. `grep -c "never bypasses it"` on both `moai.md` copies: `0`. `grep -ci "score-independent"` on all four files: `0` each.
- `grep -c "§9.2"`: `moai.md` live 2 and template 2, `spec-assembly.md` live 1 and template 1. `grep -c "§9.1"`: 2 in every one of the four files.
- `wc -l`: `spec-assembly.md` 597 and 597, `moai.md` live 284 and template 282.
- `diff` live `moai.md` against the template copy, hunk headers only: `215c215`, `245c245`, `253d252`, `283,284c282` (exactly the four pre-existing ones).
- Preserved: `[HARD] The Implementation Kickoff Approval` opening, `moai plan render-html` step, `Fail-open` paragraph, and the statement that the HTML report enriches the review surface and does not replace the gate, in `spec-assembly.md`; the merged-round sentence and the derived-completion-condition sentence in `moai.md`. The contract-mode blocks that follow in `moai.md` are untouched.
- Orphan scan: `git diff -U0 | grep -c "manager-tdd\|manager-ddd"` printed `0`.

`HUMAN GATE` label decision: kept unchanged on both lines. Reason: the rewritten sentences now qualify the gate as the operator form with the default autonomous transition of section 9.1, so the label no longer stands alone; renaming it would diverge from the same label in the sibling files the plan leaves unedited (decision-index Q12), and the requirement scope is the two stale phrases only.

Named tests, each its own command with output redirected to a file, exit code observed as its own field:

- `go test -count=1 -v -run '^(TestSpecAssembly_RewrittenToCLIPath|TestSpecAssembly_NoNewInternalTokens)$' ./internal/cli/` exit 0, 2 `--- PASS`, `ok  github.com/modu-ai/moai-adk/internal/cli 1.521s`.
- `go test -count=1 -v -run '^(TestSubSkillLOCCeiling|TestEntryRouterLOCCeiling)$' ./internal/skills/` exit 0, 2 `--- PASS`.
- `go test -count=1 -v -run '^(TestImplementationKickoffApprovalPreservedBeforeGoal|TestTemplateNoInternalContentLeak)$' ./internal/template/` exit 0, 2 `--- PASS`.
- Contract-mode block tests (`TestContractModeBlocksWellFormed`, `TestContractModeLocalTemplateParity`, `TestContractModeSSOTSections`, `TestContractModeLifecycleOrder`, `TestContractModeLifecycleEvidence`, `TestContractModeBlockCondition`, `TestContractModeAuditRetryBlocks`, `TestContractModeSyncBlocks`, `TestContractModeSigningBlocks`, plus `TestJevDoctrineAmendment` from the same file) exit 0, 10 `--- PASS`, no `--- FAIL`.

`make build` exit 0. It regenerated one tracked file, `internal/template/catalog.yaml` (one hash line); that file is part of the M4 commit.

Lexical-only checks: every grep above is a text match; whether the rewritten sentences agree with section 9.1 in meaning is a reading for the sync audit.

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

Owner: orchestrator (card t1344, lane session). Written 2026-10-02 over HEAD 81367083b on branch WT-batch-approval-gate, before the first run-phase spawn. This section supersedes two statements of §E.1 that went stale after it was written: "The delta read has not yet run" (it ran, see below) and "plan_artifact_hash: pending".

### Plan-audit evidence chain

- Iteration 1 FAIL 0.70, iteration 2 FAIL 0.775, iteration 3 FAIL 0.84 (Tier L PASS threshold 0.85), reports `.moai/reports/t1344/plan-audit-iter{1,2,3}.md` (local-only, gitignored).
- Delta read after iteration 3, audited commit 81367083b: verdict PASS-WITH-DEBT, report `.moai/reports/t1344/plan-audit-iter3-delta.md` (local-only). No must-pass criterion failed; no blocking defect remained. The delta read issued no score.
- Named debts carried into the run phase: N19, N20, N21, N22, N23, N24, N28, N29, N30 and M1 (author-fixable, text-level; the run phase may apply them while writing the doctrine). N25 and N27 are accepted by the operator (below).
- Ceiling extension: iteration 3 reached the Tier L ceiling of 3. The operator approved one further author revision plus an auditor delta read. This exceeds the ceiling by operator decision; the verdict record reads "iteration 3 exceeded the Tier L ceiling — operator approval". The leader is informed in the card completion report.

### Kickoff gate record (operator form, 2026-10-02)

The default autonomous Kickoff form (`.claude/rules/moai/workflow/auto-semantics.md` §9.1) was not available: the independent plan-audit verdict is PASS-WITH-DEBT, not PASS, and ten decision-index rows are open. The operator form ran as an orchestrator question after a findings report that included the strongest evidence against proceeding. Operator answers:

- Kickoff: enter the run phase.
- Debts N25 (log-only subtests and one-character baseline bodies can satisfy AC-007 and AC-008) and N27 (REQ-BGS-011 admits only plan-auditor verdicts, narrower than the audit cross of §9.1, on the fail-closed side): accepted as named debts.
- Progression mode: autonomous (no per-turn confirmation questions; questions are asked only at gates and for blockers). The `ac_converge` goal of `run.md` section 2 is NOT armed in this lane: its template condition requires `go test ./...` exit 0, which the lane verification rule (`.claude/rules/local/gitflow-lane-protocol.md` section 8, HARD) forbids running locally, and an armed goal blocks turn-end while this lane waits on background agents, which spins idle turns up to the ceiling. It may be armed later with a scoped condition if the operator asks.
- Preferences drained at this gate: tier L (orchestrator ruling, Q11); execution mode serial (below); PR strategy none per `.claude/rules/local/repo-local-pr-policy.md` (lanes merge into the develop integration worktree; no card pull request).
- Decision-index rows Q1-Q3, Q6-Q10, Q12, Q13 remain open and unanswered; the requirements carry the draft readings tagged in spec.md section B.8. A later answer that changes a requirement changes the plan-artifact hash and invalidates a cached audit verdict.

### Phase 1 Plan Audit Gate record

Verdict: BYPASSED by operator decision (2026-10-02). Reasons recorded: the run gate's lookup resolves `.moai/reports/plan-audit/<SPEC-ID>-review-N.md`, a directory the audit-artifact convention forbids, so the cache lookup misses; the final plan-phase verdict is PASS-WITH-DEBT at 0.84 and is not skip-eligible (PASS and score of at least 0.85 required). The operator chose to rely on the plan-phase evidence chain above instead of a fifth full audit. No fresh Phase 1 audit was run.

### Mode Selection

- Input parameters: tier L; scope 17 planned files counted by path (live and mirror separately); domains 3 (rule and skill documents, Go hook notice strings, tests); file language mix markdown and Go; concurrency benefit LOW (coding-heavy, ordered milestones); development_mode tdd (`.moai/config/sections/quality.yaml`); Agent Teams not requested.
- Mode evaluation: direct not selected (semantic change); serial selected; fanout not selected (coding-heavy and one writer per working tree); sweep not selected (not a uniform mechanical transform); agent-team not selected (not requested).
- Decision: serial
- Justification: milestones are ordered by dependency (M0, M1, M2, M4, M5, M3) and edit shared documents and one Go package; one write-capable agent runs per milestone in this tree, with the orchestrator verifying between milestones. Boundary case: 17 files and 3 domains exceed the fanout thresholds, resolved to serial by the coding-heavy tie-breaker of orchestration-mode-selection section B.2.

### Run-phase pre-flight and orchestration decisions (orchestrator, 2026-10-02)

Written over HEAD 9c18f2981 on branch WT-batch-approval-gate, before the first run-phase spawn. This block is the orchestrator's record; the run-phase evidence sections of the specialist stay empty until the specialist writes them.

- Pre-flight baseline (plan.md section C), measured in this run against this tree. Each command was run as its own invocation with output redirected to a scratch log; the decisive lines are quoted here because the scratch logs are machine-local.
  - `go test -count=1 -v -run '^TestAlwaysLoadedTokenBudget$' ./internal/config/` exit 0: `--- PASS: TestAlwaysLoadedTokenBudget`.
  - `go test -count=1 -v -run '^TestRuleTemplateMirrorDrift$' ./internal/template/` exit 0: `--- PASS: TestRuleTemplateMirrorDrift`.
  - `go test -count=1 -v -run '^(TestSubSkillLOCCeiling|TestEntryRouterLOCCeiling)$' ./internal/skills/` exit 0: both `--- PASS`.
  - `go test -count=1 -v -run '^(TestAutoRankDoctrineAmendment|TestAutoRankMirrorParity|TestSpecAssembly_RewrittenToCLIPath|TestSpecAssembly_NoNewInternalTokens)$' ./internal/cli/` exit 0: four `--- PASS`.
  - `go test -count=1 -v -run '^(TestImplementationKickoffApprovalPreservedBeforeGoal|TestTemplateNoInternalContentLeak|TestContractModeEmitterSites)$' ./internal/template/` exit 0: two `--- PASS` and `--- SKIP: TestContractModeEmitterSites`.
  - The leader notice baseline of plan.md (63 names, research.md R3.1 E17, package internal/hook) was not run in this block; it is measured at the start of M5, the milestone that touches those files.
- Gap, not a pass: `TestContractModeEmitterSites` skipped because `MOAI_GR_BASE` is unset (`contract_mode_guided_test.go:696`). Run once with `MOAI_GR_BASE=c50da9c2f8aa1227073bd77caa07ca1c75b8d81b` (the merge-base of this branch and develop) it exits 1: six unclassified Kickoff documents at that base (`moai-mcp-tools-catalogue.md`, `auto-semantics.md`, `contract-autonomy.md`, live and mirror), while the emitter-copy count was 20 as expected. Those documents entered develop after the classification table was written, so the failure is not attributable to this card, and the default invocation (no variable, as in CI) skips. This guard therefore gives no baseline signal for this card; the contract-mode block guards run in M4 are the ones that decide.
- develop moved: the local develop tip is now 802a72235 while this branch's merge-base is c50da9c2f. Absorbing it is the integration window's step (gitflow-lane-protocol section 8 and CLAUDE.local.md section 4.1), not a run-phase step.
- Delegation shape: each milestone is spawned as Agent(general-purpose) carrying the manager-develop role instructions, not as the manager-develop agent type. Reason, measured on 2026-10-01 (card t1318): the typed spawn lands in its own isolated worktree and every write into the card worktree is then refused. This is the documented substitute for card-worktree-pinned implementation; ownership boundaries (no edits to spec.md, plan.md, acceptance.md bodies) are written into each prompt.
- tasks.md is not generated. Planned files are the plan.md section F milestone lists and the spec-compact.md files-to-modify list; a tasks.md would join the plan-artifact hash set and is not one of the tier L artifacts. The drift guard compares actual changes to that list.

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

M4 commit: d196a9de4.

### M5 — leader notice sentence, RED first

Base tree d196a9de4 (M4 committed). Product Go changes are the four files `internal/hook/session_start_kanban_i18n.go`, `session_start_kanban.go`, `session_start_factory_i18n.go`, `session_start_factory.go`; test Go changes are the `TestKanbanLocalesCoverEveryField` field table of `session_start_kanban_i18n_test.go` (one entry appended, nothing weakened) and the new `session_start_leader_gate_notice_test.go` (top-level test `TestLeaderNoticeBatchGatePointer`). The new struct field is named `gateSummary` in both i18n tables instead of the plan's candidate name `leaderGateSummary`: a 17-character key would make gofmt realign the 13-field kanban struct and every locale block, turning a five-line change into a rewrite of unrelated lines.

Baseline before any Go edit, 63-name leader notice command of research.md R3.1 E17, output redirected to a file, run on tree 4e0a60bc1 (the M4 edits touch no Go): `--- PASS` 63, `--- FAIL` 0, `=== RUN` 63, final line `ok  github.com/modu-ai/moai-adk/internal/hook  15.181s`.

RED (new test written before any product change; tree d196a9de4 plus the new untracked test file): `go test -count=1 -v -run '^TestLeaderNoticeBatchGatePointer$' ./internal/hook/` redirected to a file, exit code 1 observed as its own step. Decisive lines, verbatim:

```text
    session_start_leader_gate_notice_test.go:178: real_grid violations=[omit-pointer omit-name-token]
    session_start_leader_gate_notice_test.go:190: kanban notice (en) violates [omit-pointer omit-name-token]; sentence line ""
    session_start_leader_gate_notice_test.go:229: mutant omit-pointer reported [omit-pointer omit-name-token], want exactly [omit-pointer]
--- FAIL: TestLeaderNoticeBatchGatePointer (8.72s)
    --- FAIL: TestLeaderNoticeBatchGatePointer/handler_level/kanban (1.19s)
    --- FAIL: TestLeaderNoticeBatchGatePointer/handler_level/factory (3.08s)
FAIL	github.com/modu-ai/moai-adk/internal/hook	9.668s
```

Classification per tdd-result-contract: EXPECTED_RED. The package compiled and `go vet ./internal/hook/` exited 0 first, so the failure is the intended assertion (pointer and name token absent from every rendered leader notice and from both handler channels), not TOOL_FAILURE. The same run showed the guards that must hold before the change already PASS (`mutant_omit-locale`, `lane_and_companion_lack_pointer`, `source_scan`).

GREEN: field `gateSummary` plus four locale values in each of the two i18n tables; kanban notice block (e) gets one `context = append(context, m.gateSummary)` line after the settings line; factory notice block (e) gets the field in its `strings.Join` list. The sentence is one sentence per locale carrying `batch gate summary` and the address `.claude/rules/moai/workflow/auto-semantics.md` followed by the section mark 9.2, and nothing else of the canonical content.

AC-015 readings on the final tree (each its own command, output redirected to a file):

- `go test -count=1 -v -run '^TestLeaderNoticeBatchGatePointer$' ./internal/hook/` exit 0: `--- PASS` 21, `--- FAIL` 0, `=== RUN` 21 (equal, counted the same way), `no tests to run` 0, mutant lines `mutant <name> rejected: reported=<name>` 6 (omit-locale, question-tool-name, omit-pointer, omit-name-token, card-id, one-sided-leader; each pair of names identical).
- `grep -c AskUserQuestion` over the four product files: 0 in each.
- `grep -cF 'auto-semantics.md` §9.2'` over the two i18n files: 4 and 4. `grep -cF 'batch gate summary'`: 4 and 4 (a first draft carried the name token in a struct comment and printed 5; the comments were reworded so the count is one per locale value).
- 63-name command re-run after the change: `--- PASS` 63, `--- FAIL` 0, `=== RUN` 63, `ok  github.com/modu-ai/moai-adk/internal/hook  13.911s`.
- `gofmt -l` over the six changed or new Go files: no output. `go vet ./internal/hook/` exit 0. `golangci-lint --version` printed v2.1.6 (the CI version), and `golangci-lint run --timeout=5m ./internal/hook/` exited 0.

Cost disclosure, measured in a throwaway test (deleted before the commit) by rendering the notices in the package: the sentence adds one line to a leader notice, 232 bytes in en (297 in ja, 248 in ko, 232 in zh — bytes, not characters). The en kanban leader notice renders 1768 bytes after and 1536 before (the before value is the after value minus the added line, a derivation, not a second render); the en factory leader notice renders 2557 after and 2325 before. Sessions that are not a leader session at startup pay 0 bytes: the field is read only inside the two leader notice builders, and the existing source gate leaves resume, clear, compact and fork without a notice.

Scope check: `git diff --numstat` over the four product files before the commit read `1 1` (factory.go), `5 0` (factory_i18n.go), `1 0` (kanban.go), `5 0` (kanban_i18n.go); the test table file reads `1 0`. `make build` exit 0 and regenerated no tracked file.

Lexical-only checks in M5: every token check of the new test is a substring or regular-expression match on the sentence line (a sentence with the pointer but the opposite meaning would pass, spec limit G-7); the forbidden-token checks run on the line that carries the name token, not on a parsed sentence; the card-id shape is a letter t followed by three to five digits.

### M3 — pointers (run last, third manager-develop role spawn, card t1344)

Base tree 193da0e05 (M0, M1, M2, M4, M5 committed, working tree clean at start). Edited exactly two documents, each in the template copy first and then the live copy with the same `old_string` (never copied onto each other): `kanban-dispatch.md` (the Boundaries "No gate bypass." bullet) and `workflows/run.md` (line 137). The preserved span of `kanban-dispatch.md` from "Promotion is the operator's act, always." through "The self-dispatch lane exception." was not touched, and the pre-existing live-versus-mirror difference at line 177 was left as it was (`diff` of the two copies after the edit prints only that one hunk, `177c177`).

Pointer wording. `kanban-dispatch.md`: the clause "keep-set gates (...) still require the operator" now continues ", and operator-form Kickoff rows that wait together are presented through the batch gate summary (`.claude/rules/moai/workflow/auto-semantics.md` §9.2)." `run.md`: the citation "(`.../auto-semantics.md` §9.1)" became "(`.../auto-semantics.md` §9.1 and §9.2)". Neither edit restates the 9.2 format or carries the canonical-only tokens.

"RED" reading, pre-edit tree 193da0e05 (observed before any edit): `grep -c "§9.2"` over the four pointer files printed `kanban-dispatch.md:0` (live and mirror), `run.md:0` (live and mirror). Pre-edit sizes: `wc -c` live 26807, mirror 26485 (equal to the merge-base blobs, so M3 starts from the unmodified baseline).

AC-016 readings, post-edit (each its own command; merge-base read at measurement time with `git merge-base develop HEAD` = c50da9c2f8aa1227073bd77caa07ca1c75b8d81b):

- (a) `grep -rlF "counter_refs=" .claude/rules .claude/skills internal/template/templates/.claude/rules internal/template/templates/.claude/skills internal/hook --exclude='*_test.go'` printed exactly the two `auto-semantics.md` copies; the same command with `searched=` printed the same two. Lexical.
- (b) `grep -c "§9.2"`: `kanban-dispatch.md` 1 (live) and 1 (mirror), `run.md` 1 (live) and 1 (mirror). The M4 and M5 pointer surfaces were measured in their own sections and were not re-read here. Lexical.
- (c) `git show c50da9c2f:<path> | wc -c` versus `wc -c`: live 26807 to 26959 (growth 152 bytes), mirror 26485 to 26637 (growth 152 bytes), both under 1000.
- (d) `git diff --numstat c50da9c2f..HEAD` after the commit: `1 1` for `run.md` live, `1 1` for `run.md` mirror (additions equal deletions, 199 lines unchanged), `1 1` for `kanban-dispatch.md` live and mirror (one added line each, at most 3). Before the commit the working-tree form read the same.
- Non-vacuity control: `git diff --name-only c50da9c2f..HEAD` printed 27 lines after the M3 commit (23 before it; at least one either way), so the readings above are measurements, not "unmeasurable". This reading is valid only before the card merges into develop: after the merge the merge-base becomes the card tip, the range empties, and (c) and (d) would pass vacuously.

`make build` exit 0 (no error output; the final line is the `go build` invocation). It regenerated one tracked file, `internal/template/catalog.yaml` (one hash line, the `moai` skill entry, derived from the `run.md` edit); it is part of the M3 commit.

Named tests, each its own command with output redirected to a scratch file, exit code observed as its own field, counts from `grep -c`:

- `TestAlwaysLoadedTokenBudget` (`./internal/config/`): exit 0, `always-loaded surface = 64265 tokens (budget 77600, headroom 13335, 16 entries)`; before M3 the M2 reading was 64227 tokens, headroom 13373 (so M3 costs 38 tokens). `--- PASS` 1, `--- FAIL` 0.
- `TestRuleTemplateMirrorDrift` (`./internal/template/`): exit 0, `--- PASS` 10 including subtests (1 top level), `=== RUN` 10, `--- FAIL` 0.
- `TestSubSkillLOCCeiling`, `TestEntryRouterLOCCeiling` (`./internal/skills/`): exit 0, `--- PASS` 2, `--- FAIL` 0.
- `TestAutoRankDoctrineAmendment`, `TestAutoRankMirrorParity`, `TestSpecAssembly_RewrittenToCLIPath`, `TestSpecAssembly_NoNewInternalTokens` (`./internal/cli/`): exit 0, `--- PASS` 4 top level (34 with subtests), `=== RUN` 34, `--- FAIL` 0.
- `TestImplementationKickoffApprovalPreservedBeforeGoal`, `TestTemplateNoInternalContentLeak`, `TestBatchGateSummaryDoctrine` (`./internal/template/`): exit 0, `--- PASS` 3 top level (49 with subtests), `=== RUN` 49, `--- FAIL` 0, `no tests to run` 0.
- Contract-mode block tests of `contract_mode_blocks_test.go` (`TestContractModeBlocksWellFormed`, `TestContractModeLocalTemplateParity`, `TestContractModeSSOTSections`, `TestContractModeLifecycleOrder`, `TestContractModeLifecycleEvidence`, `TestContractModeBlockCondition`, `TestContractModeAuditRetryBlocks`, `TestContractModeSyncBlocks`, `TestContractModeSigningBlocks`, `TestJevDoctrineAmendment`): exit 0, `--- PASS` 10 top level (36 with subtests), `=== RUN` 36, `--- FAIL` 0, `no tests to run` 0.

Lexical-only checks in M3: every pointer count and the canonical-token sweeps are text matches. Whether the two pointer sentences agree in meaning with section 9.2 is a reading for the sync audit.

M3 commit: 658cffa52.

### End checks (after M3; tree 658cffa52, branch WT-batch-approval-gate)

Each command was run as its own invocation.

- AC-008 V1: `ls .moai/specs/SPEC-AUTONOMY-BATCH-GATE-001/baseline-gate-rounds.md` printed the path, exit 0. V2: `git check-ignore -v <path>` printed nothing, exit 1. V3: the four `grep -c "^<Label>: [^ ]"` commands printed 1, 1, 1, 1 (Command, Observed output, Classification method, Limits). V4: `git log --format=%h --diff-filter=A -- <path>` printed `61801e166` (one line). V5: `git diff-tree --no-commit-id --name-only -r 61801e166` printed one line, the baseline path. V6: `git log --format=%h -- <path>` printed `61801e166` (exactly one line, equal to B). V7: `git log --format=%h -S"### 9.2 The batch gate summary" -- internal/template/templates/.claude/rules/moai/workflow/auto-semantics.md` printed `4e0a60bc1` (exactly one line, equal to D). V8: `git merge-base --is-ancestor 61801e166 4e0a60bc1` exit 0. The V6 to V8 ordering evidence is the commit graph itself (B 61801e166 before D 4e0a60bc1).
- AC-009: `git diff --name-only develop...HEAD -- '*.go'` printed exactly seven paths: `internal/hook/session_start_factory.go`, `internal/hook/session_start_factory_i18n.go`, `internal/hook/session_start_kanban.go`, `internal/hook/session_start_kanban_i18n.go`, `internal/hook/session_start_kanban_i18n_test.go`, `internal/hook/session_start_leader_gate_notice_test.go`, `internal/template/batch_gate_summary_doctrine_test.go`. `git diff --numstat develop...HEAD` over the four product files: `1 1` (factory.go), `5 0` (factory_i18n.go), `1 0` (kanban.go), `5 0` (kanban_i18n.go). The check that the product changes are limited to the notice sentence field and its merge line is a reading for the sync audit (spec limit G-8). The `develop...HEAD` merge-base at measurement time was c50da9c2f; valid only before the merge.
- AC-010: the guard commands are the M3 runs above (budget headroom 13335, which is at least 0). `wc -l` of `run.md`: 199 live, 199 mirror. The 63-name leader notice command was not re-run in the end check (no Go file changed after M5, where it measured 63 `--- PASS` and 0 `--- FAIL`).
- Tier recount (spec.md section A.5): `git diff --name-only develop...HEAD -- . ':(exclude).moai/specs' ':(exclude).moai/reports'` printed 18 lines against the enumerated 17. The 17 enumerated files are all present (10 rule and skill documents as 5 live and mirror pairs, the new template guard test, 4 product Go files, 2 hook tests). The one extra path is `internal/template/catalog.yaml`, a derived artifact that `make build` regenerates (`gen-catalog-hashes --all`; the `moai` skill hash changed because the `moai.md`, `spec-assembly.md` and `run.md` skill files changed); it is not in the plan's list and nothing was edited to force the count to 17.

## §E.3 Run-phase Audit-Ready Signal

run_status: audit-ready
run_complete_at: 2026-10-02
tier: L; mode: serial; development mode: tdd; operator-form Kickoff, Phase 1 plan-audit gate BYPASSED by operator decision (section F).

Run-phase commits (branch WT-batch-approval-gate, base develop merge-base c50da9c2f):

- M0 baseline of question-tool gate rounds: 61801e166
- M1 guard test, RED: 5936aeb32
- M2 canonical section 9.2, GREEN: 4e0a60bc1
- M4 stale Kickoff wording alignment: d196a9de4
- M5 leader notice pointer sentence: 193da0e05
- M3 pointers (kanban-dispatch, run): 658cffa52

The commit that carries this section and the end-check evidence above is the run-phase evidence commit and is not listed (a commit cannot cite its own hash).

Known debts carried: open decision-index rows Q1-Q3, Q6-Q10, Q12 and Q13 are implemented as draft readings (spec.md section B.8); the doctrine section is about twice the 3 KB target (5986 bytes); `TestContractModeEmitterSites` gives no baseline signal because it skips without its environment variable (and fails when run with `MOAI_GR_BASE` set, for documents not attributable to this card); the guard tests are lexical matchers with the limits G-4, G-7 and G-10 of spec.md (a body that keeps an anchor next to a contradicting sentence, a pointer sentence with the opposite meaning, and one-character baseline bodies can pass); the Tier recount reads 18, not the enumerated 17, because the derived artifact `internal/template/catalog.yaml` is regenerated by `make build`.

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

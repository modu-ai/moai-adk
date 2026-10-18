auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-TODO-UNIFY-001 (plan-audit iteration 2 of 3)

Reasoning context ignored per M1 Context Isolation. Read: the seven SPEC artifacts, `.moai/reports/t1641/plan-audit.md`, `.moai/reports/t1641/progress.md`, the tree at the audited commit, and the decision board. Tree clean, HEAD equals the audited commit before and after the audit (`git rev-parse HEAD` printed `3132f60c23afeb620415f9394c5eff45c4e03191`; `git status --short` printed nothing).

Result: FAIL. Aggregate 0.80 against the Tier L threshold 0.85 (up from 0.62, so no STOP). All nine must-pass criteria pass or are N/A, but seven blocking findings remain and the required codex backend returned `fail`. The revision closed the iteration-1 blockers; the new blockers sit mostly in the NEW work (the automatic receipt, the `spec` verb, the warning hook ordering), as the lead's brief expected.

verdict: FAIL
audited_sha: 3132f60c23afeb620415f9394c5eff45c4e03191
overall_score: 0.80
must_pass_failed: 0
blocking_count: 7
plan_artifact_hash: 966727e2c52676fcc89f1f282c136e8f25c310c4716c578cb0de929142216304
scope: delta
fix_scope: .moai/specs/SPEC-TODO-UNIFY-001/spec.md#REQ-TU-017; .moai/specs/SPEC-TODO-UNIFY-001/spec.md#REQ-TU-003; .moai/specs/SPEC-TODO-UNIFY-001/spec.md#REQ-TU-020; .moai/specs/SPEC-TODO-UNIFY-001/spec.md#REQ-TU-022-to-REQ-TU-023; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-002; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-004; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-014; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-016; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-021; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-025; .moai/specs/SPEC-TODO-UNIFY-001/design.md#D.2; .moai/specs/SPEC-TODO-UNIFY-001/design.md#D.5; .moai/specs/SPEC-TODO-UNIFY-001/design.md#D.7; .moai/specs/SPEC-TODO-UNIFY-001/plan.md#M2a; .moai/specs/SPEC-TODO-UNIFY-001/plan.md#M3b
defect_class: D16=design D17=design D18=ac-wording D19=design D20=design D21=design D22=traceability
reread_hunks: none
convergence_overall: fail
required_backend: codex fail

## Must-Pass Results

- [PASS] MP-1 REQ number consistency. Layer: requirement layer, `spec.md`. `grep -n -E '^\*\*REQ-TU-'` printed 25 definitions at spec.md:75, 77, 79, 81, 83, 85, 87, 89, 91, 93, 95, 97, 99, 101, 105, 107, 109, 111, 113, 115, 117, 119, 123, 125, 127. Python check: `REQ count 25 unique 25`, `REQ sequence contiguous: True`, three-digit padding. Cosmetic only: REQ-TU-025 (spec.md:119) sits before REQ-TU-022 to 024 (spec.md:123-127) (D28). AC set: `AC count 25`, `AC sequence contiguous: True`; the file order is non-monotonic by design, stated at acceptance.md:5.
- [PASS] MP-2 EARS/GEARS compliance. Layer judged: the 25 `REQ-TU-` entries of `spec.md` only; the Given-When-Then entries of `acceptance.md` are the verification layer and were graded under Group 4. Patterns read: Ubiquitous (001, 004, 005, 007, 008, 009, 011, 012, 013, 014, 018, 021, 024), Event-driven (002, 010, 015, 017, 019, 020, 022, 023, 025), Where (006), legacy Unwanted (003, 016), valid until 2026-11-22. REQ-TU-023 now reads "When the card already carries a different spec_id ..., the setter shall refuse" (spec.md:125): the iteration-1 `If ... shall` form is gone.
- [PASS] MP-3 YAML frontmatter validity. spec.md:2-24 carries all 12 canonical fields: `id`, quoted `title`, `version: "0.2.0"` (quoted), `status: draft`, `created`/`updated` 2026-10-18, `author`, `priority: P1`, `phase: "v3.2.0 target"`, `module`, `lifecycle: spec-anchored`, `tags` as a comma string; optional `tier: L` and `related_specs`. No snake_case alias. The four sibling artifacts carry no `status:` (decision-index.md:3 states it). `moai spec lint SPEC-TODO-UNIFY-001` with a binary built from this tree printed `0 error(s), 0 warning(s)` plus one INFO `OwnershipTransitionUnmeasured` (commit 63829fe20a has no `Authored-By-Agent` trailer); `moai spec lint --strict` exited 0 (`strict exit=0`).
- [N/A] MP-4 language neutrality: single-repository Go CLI scope, no multi-language tooling enumerated.
- [PASS] MP-5 D7 cross-SPEC reconciliation. Python port of the D7 verb (awk and sh not attempted, see Gaps) over the nine `SPEC-...` ids of spec.md: FACTORY-COMPLETION-RECOVERY-001 draft, FACTORY-RECORD-001 completed, FACTORY-RECORD-REPAIR-001 draft, GITHUB-FLOW-DEFAULT-001 in-progress, GTD-CANON-BODY-001 completed, KANBAN-TODO-CLI-001 in-progress, LOCAL-MAIN-FLOW-001 completed, TODO-LAND-AUTO-DONE-001 completed, and the SPEC itself draft. None retired, superseded, or archived; no `REVIEW:` and no `SHOULD:` line. See D23 (optional) for a related contradiction with a draft SPEC's requirement that the verb cannot see.
- [PASS] MP-6 D8 cross-platform discipline. Python port over the six artifacts printed `D8 <file> syscall-sections-without-tag: 0 []` for each; D8-4 auto-pass.
- [PASS] MP-7 clarification gate. `grep -rn -c 'NEEDS CLARIFICATION'` over the SPEC directory printed `progress.md:0`, `decision-index.md:0`, `spec.md:0`, `design.md:0`, `research.md:0`, `acceptance.md:0`, `plan.md:0`. Iteration 1 found nine markers in plan.md; none remain.
- [PASS] MP-8 RED-now cell re-execution. Release-blocking criteria are the class A ones: AC-TU-001, 005, 007, 008, 010, 017 and the wording clause of AC-TU-013. Every cited ledger command was re-executed on the current tree with its exit code observed (single `git grep` or `wc` invocations; none is a test runner, so the empty-selection rule does not apply). Results: E-01 exit 0, 73 paths (`wc -l` printed 73); E-02 exit 0, six lines (`card_transition.go:19, 157, 203, 207, 333, 334`); E-03 exit 0, six lines (`gtd.md:150, 161, 173, 185, 194, 211`); E-04 exit 1, empty; E-06 exit 1, empty; E-07 exit 1, empty; E-08 exit 0 (`gtd_compat_test.go:83`); E-11 exit 0 (`23780 AGENTS.md`); E-12 exit 0 (`fr_transition_test.go:166:	if len(accepted) != 71 || refused != 370 {`); E-13 exit 0, two lines; E-14 exit 0, eight lines; E-15 exit 1, empty; E-16 exit 0 (`AGENTS.md:1`, `AGENTS.md.tmpl:1`); E-17 exit 1, empty; E-18 exit 1, empty. E-09 reproduced on the installed build only (`Unknown command "reflect" for "moai todo".`; its exit code was not observed because the output was piped). E-10 was not re-run (Gap); AC-TU-001's RED stands without it through E-15 and E-08. The iteration-1 Gap on unobserved exit codes of empty `git grep` runs is closed: exits 1 were observed this run. Class B criteria record their RED at the owning milestone and are unadopted until then (acceptance.md:14), which is the §2.1 disposition. Boundary stated: re-execution confirms output, not that the grep measures the premise.
- [PASS] MP-9 cross-artifact ordering. The CN-4 verb was not executed (awk and sh not attempted); a Python port printed `CN4 COLLECTED: 9 milestones in plan order (M1 M2 M3 M4 M5 M6 M7 M8 M9), 0 exit bindings` and 22 candidates in acceptance.md (lines 17, 23, 28, 30, 37, 63, 64, 65, 69, 72, 74, 84, 89, 92, 100, 102, 111, 114, 120, 128, 136, 150). With 0 `Exit:` bindings no `CONFLICT:` line can be produced mechanically, so I read all 22. None orders a criterion against a milestone the plan schedules on the forbidden side: the ordering words concern cobra execution order (L30, L136), within-act order of the close (L69, L72, L100, L102, L114), the RED order inside M1 (L89, matches plan.md M1 RED list), a baseline pin (L120) and runner-output reading (L17). No confirmed ordering conflict. Nearest tension, not a conflict: D25.

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | minor ambiguity in one or two requirements | REQ-TU-017 fact (2) says the verdict file is "committed" without saying where it is read from (card worktree HEAD versus the merge) (spec.md:109, design.md:88); REQ-TU-003 says "flag validation" but the mechanism cannot honour it (D19) |
| Completeness | 0.90 | all sections and frontmatter present, between the 0.75 and 1.0 bands | HISTORY (spec.md:28), background and why (:35, :55), scope and tier (:59), requirements (:69), constraints (:129), three `### Out of Scope — <topic>` H3s with bullets (:139, :145, :152), design, research, decision-index; held below 1.0 because the setter's state set does not cover the cited wedge (D20) |
| Testability | 0.75 | one AC not satisfiable by a correct build, others need small rows | AC-TU-004's exact 32-verb set conflicts with the `spec` verb M3b registers (D18); AC-TU-002 lacks the required-flag row (D19); AC-TU-016 lacks stale-audit and tree-mismatch rows (D16, D17) |
| Traceability | 0.80 | every REQ has an AC, two clauses are not asserted | 25 of 25 REQs covered, 0 orphan ACs (matrix below); REQ-TU-020's non-`merged-local` clause and REQ-TU-017's log-row issuer have no AC row (D22) |

Aggregate (mean): (0.75 + 0.90 + 0.75 + 0.80) / 4 = 0.80. Tier L threshold 0.85 (spec.md:14 `tier: L`; spec-workflow.md § SPEC Complexity Tier). Below the threshold, blocking findings open, and the required backend failed: FAIL. The score moved up 0.18 from iteration 1, so the STOP clause does not fire.

## REQ to AC matrix (built by reading the 25 bold paragraphs against the 25 AC headers)

The Group 4 verb collects definitions only from list items, headings, or table rows opening with the id; this SPEC writes bold paragraphs, so the port printed `TRACE verb-form definitions collected from spec.md: 0`, which the real verb reports as a GAP. Neither that GAP nor `moai spec lint`'s silence is corroboration. The manual matrix, confirmed by a Python extraction of the `(REQ-...)` parentheses of each AC header (`UNCOVERED: []`, `ORPHAN: []`):

| REQ | AC | REQ | AC |
|-----|----|-----|----|
| REQ-TU-001 | AC-TU-003 | REQ-TU-014 | AC-TU-013 |
| REQ-TU-002 | AC-TU-001 | REQ-TU-015 | AC-TU-014 |
| REQ-TU-003 | AC-TU-002 | REQ-TU-016 | AC-TU-015 |
| REQ-TU-004 | AC-TU-004 | REQ-TU-017 | AC-TU-016 |
| REQ-TU-005 | AC-TU-005 | REQ-TU-018 | AC-TU-017 |
| REQ-TU-006 | AC-TU-006 | REQ-TU-019 | AC-TU-018 |
| REQ-TU-007 | AC-TU-007 | REQ-TU-020 | AC-TU-014, AC-TU-023 |
| REQ-TU-008 | AC-TU-008 | REQ-TU-021 | AC-TU-020 |
| REQ-TU-009 | AC-TU-009 | REQ-TU-022 | AC-TU-021, AC-TU-025 |
| REQ-TU-010 | AC-TU-009 | REQ-TU-023 | AC-TU-021 |
| REQ-TU-011 | AC-TU-010 | REQ-TU-024 | AC-TU-022 |
| REQ-TU-012 | AC-TU-011 | REQ-TU-025 | AC-TU-019 |
| REQ-TU-013 | AC-TU-012 | (all) | AC-TU-024 |

Result: 25 REQ, 25 AC, 0 uncovered, 0 orphan. Tier L ceilings 25 / 25 are met exactly (spec-workflow.md tier table). Two clauses inside covered REQs are not asserted by any AC row (D22).

## D1 to D15 re-verification

| ID | Result | Evidence (re-verified against the new text) |
|----|--------|---------------------------------------------|
| D1 markers | FIXED | `grep -rn -c 'NEEDS CLARIFICATION'` printed 0 for all seven files; plan.md:39-51 and decision-index.md Q1 to Q9 each carry a decision, evidence, rejected alternate, and a decision record. |
| D2 defaults contradicting the card | FIXED | The defaults are replaced: Q5, Q6, Q7 now satisfy the card's two completion criteria (spec.md:109-117; acceptance.md:96-102, DoD :156-160). |
| D3 no oracle | FIXED, with new gaps | REQ-TU-017, 019, 020, 025 (spec.md:109, 113, 115, 119) state concrete conditions; AC-TU-014, 016, 018, 019, 020, 023 have oracles and eight-row matrices. Every condition is testable as written. Whether the conditions are sufficient is D16 and D17. |
| D4 circular replacement path | FIXED | Edge T29 declared before T18 is removed (plan.md M1 RED order, design.md §D.6, spec.md:113); AC-TU-018 asserts the card reaches `done` with no push step in a no-remote repository (acceptance.md:87). `git grep -n -E '"T29"'` printed nothing, exit 1: T29 is absent today. |
| D5 two-store atomicity | FIXED for the close, PARTLY for the setter | Close: factory edge first, queue archive second, recovery stated (spec.md:115; design.md:93; AC-TU-023). Setter: order and recovery stated (spec.md:123; design.md:115-117; AC-TU-025). Residual: the recovery promise is unconditional but a lease interleaving defeats it (D21). |
| D6 independence claim | FIXED, with a new inconsistency | plan.md:21-35 replaces the claim with a graph; M3b lists `internal/cli/todo.go` (plan.md:113). New: M3b's `spec` verb versus AC-TU-004 (D18). |
| D7 warning ordering | PARTLY FIXED | Help, flag-parse error, argument validation, and the mistyped-verb guard now precede the hook: confirmed in cobra v1.10.2 `command.go` (`ParseFlags` :919, help :935, `ValidateArgs` :968, first `PersistentPreRunE` :985). NOT fixed: `ValidateRequiredFlags` is at :1007 and runs after the hook, so required-flag and flag-group refusals still print the warning (D19). |
| D8 REQ-TU-014 oracle | FIXED | AC-TU-013 now carries a wording clause (acceptance.md:127-129) with `TestAgentsMDTodoRowWording` created at RED; E-16 reproduces; ceiling `CodexContractByteCeiling = 24576` (token_budget_guard.go:101), baseline 23780. |
| D9 mutant and zero-match selectors | FIXED, with a new conflict | The expected verb set is fixed in the AC (acceptance.md:36-39); all 19 existing tests named in AC-TU-004, 006, 007, 011, 017, 018, 022 exist (see spot-check P8). The fixed set is itself wrong (D18). |
| D10 floors and host package | FIXED | Floors measured: 326 matching lines over 17 files reproduced exactly (2+2+42+42+8+36+194); the 5,000 swept-file floor sits below 8,827 text files and 263 binary files that I measured (9,090 in scope); host package `internal/template` named (acceptance.md:59). |
| D11 preservation guard | FIXED | AC-TU-011 (acceptance.md:120-123) adds behavioural checks. `gtd_answer_test.go:19` asserts the `.moai/gtd/answers/t128.md` path; `gtd_compat_test.go:166-170` drives `organize not-an-id` and `clarify ../../escape` refusals and `captured.ItemID` acceptance, so the id pattern is exercised. |
| D12 template-source claim | FIXED | plan.md §C item 7 and §G item 1 name `.moai/workflows/gtd.md` as deployed-only; always-loaded files are excluded as class E6 (design.md:54) and edited at M9 (plan.md:155-159). |
| D13 proxy REQ-cells | FIXED, with a nit | acceptance.md:25 and :64 state the proxy's flip condition; E-15 is the new proxy. Nit: D24. |
| D14 requirement wording | FIXED | REQ-TU-023 is event-detected form (spec.md:125); the t1621 overlap is recorded (research.md:215). Identifier names (T18, T29) persist inside REQ-TU-018/019 (D28, minor). |
| D15 milestone sizing | PARTLY FIXED | M3 split into M3a (230 lines) and M3b (200); M8 unit redefined as changed lines (plan.md:150); the docs-site page merge is dropped (spec.md:143). M2a (450) and M2b (520) still exceed the plan's own 400-line rule at estimate level (D27). |

## Premise spot-checks (own commands, this run)

Single-invocation, read-only, tree HEAD `3132f60c23`. "Confirmed" means my command printed the quoted output. Reading-only items are labelled.

| # | Premise (source) | My command | Observed | Result |
|---|------------------|-----------|----------|--------|
| P1 | Cobra v1.10.2 order (research.md §R.11, design.md:28) | `sed -n 915,1040p /Users/goos/go/pkg/mod/github.com/spf13/cobra@v1.10.2/command.go` and `grep -n` of the anchors | `919: err = c.ParseFlags(a)`, `935: return flag.ErrHelp`, `968: if err := c.ValidateArgs(argWoFlags)`, `985: if p.PersistentPreRunE != nil {`, `1007: if err := c.ValidateRequiredFlags(); err != nil {` | confirmed; the cited lines match. It also shows required-flag and flag-group validation after the pre-run hook (D19). Established by reading; no throwaway program was run. |
| P2 | Required flags exist on stage verbs | `grep -n -E 'MarkFlagRequired|MarkFlagsMutuallyExclusive' gtd.go gtd_answer.go todo*.go` | `gtd.go:98 MarkFlagRequired("event")`, `:121 ("disposition")`, `:158 ("class")`, `todo.go:356 MarkFlagsMutuallyExclusive("auto", "auto-lane", "auto-leader")` | confirmed (D19) |
| P3 | T29 absent; pin 71/370 unchanged | `git grep -n -E '"T29"' -- internal/homestate/card_transition.go` (exit 1, empty); read `fr_transition_test.go:127-183`, `frFullRequest` (:88-100) | no `ApprovalUUID` field in `frFullRequest`; `verifyTransitionApproval` refuses an empty uuid with `ErrApprovalCardMismatch`; final loop asserts `{merged-local, done}` and `{pushed, ci-green}` must be refused | confirmed by static reading, not run (go test forbidden). Adding T29 and removing T18 and the reservation keeps 71 and 370: the fixture request carries no uuid, so T29 refuses. REQ-TU-018 (spec.md:111) states the pin does not change. |
| P4 | `VerifyBinding` demands the leader role marker; the SPEC leaves it unchanged | read `leader_approval.go:116-142` | `if a.IssuerRole != ApprovalIssuerLeader { ... }`; owner label compared via `CanonicalApprovalOwnerLabel`; run, version, non-empty equal evidence hash checked | confirmed; the gate verifies binding only, so an `auto` receipt with role `leader` and the card's own `EvidenceSHA` passes by construction. All trust therefore rests on the three facts (D16, D17). |
| P5 | 32-verb set of AC-TU-004 | read `todo.go:342-349`, count the `AddCommand` constructors | 26 constructors (Add, List, Done, Undone, Next, Claim, Unpick, Edit, Move, Drop, Undrop, Merge, Hold, Unhold, Analyze, Relate, Unrelate, Why, Trace, PR, Landed, AutoDone, ExportJSON, History, Show, Triage) | the 26 holds; with six stage verbs 32, but M3b adds `spec` (plan.md:112): 33 (D18) |
| P6 | Verdict facts: the reader reads the card worktree | read `card_evidence_readers.go:96-130` (`readAuditVerdict`) | globs `<worktree>/.moai/reports/<card>/<phase>*.md`, newest file, `verdict:` and `audited_sha:` lines | confirmed; the file is read from a tree the lane writes (D16) |
| P7 | Re-measure callback binding | read `integration_remeasure.go:86-135` and `factory_card.go:2372-2378` | comment: "rec.Tree reads AS WRITTEN, never forced to the key"; the T16 callback is `ReadRemeasureRecord` then `ValidateRemeasureRecord`; `ValidateRemeasureRecord` checks non-empty fields, full-SHA shape, exit code, test count, but never compares `rec.Tree` to the key | confirmed by reading (D17); I did not reproduce codex's temporary-repo run |
| P8 | Named existing tests exist | `git grep -n -E 'func (<19 names>)\('` | all 19 printed with the cited lines (for example `gtd_compat_test.go:49`, `:124`, `:164`; `todo_test.go:121`, `:214`, `:1078`, `:1123`; `fr_transition_test.go:127`; `f7_gate_owner_admission_test.go:14`; `token_budget_guard_test.go:84`; `workflow_rule_paths_pinned_test.go:24`; `gtd_canonical_surface_test.go:10`) | confirmed |
| P9 | 17-names published-skill set (acceptance.md:50) | read `published_skills.go:29-36` | 3+3+4+3+3+1 = 17 names, `moai-gtd` and `moai-todo` among them | confirmed |
| P10 | 17 files / 300-token floor; 5,000-file floor | `git grep -c -E 'moai (todo|gtd) [a-z]' -- <17 files>`; `iter2_scope.py` | per-file counts sum to 326 over 17 files; `tracked 16151`, `in scope before binary filter 9090`, `text 8827 binary 263` | confirmed; floors 300 and 5,000 are sound |
| P11 | Lane holds no lease at `merged-local`; the retry path after the factory edge | read `card_record.go:66-68`, `card_transition.go:830-838` | `leaseHoldingStates` excludes `merged-local`; `OwnerLabel` is not cleared on a terminal transition | the factory-first retry is reachable: a repeat finds the row `done`, owner label intact, receipt re-stamped (`leader_approval` UPDATE at the approval-rebind step). The revision's recovery design holds for the close (reading, not run). |
| P12 | Setter state set versus the cited wedge | `moai decision read --scope card:t1641` (record `d-20261010T001732Z-d4e8`); read `card_picked.go:185` | the record's predicate: "a card of run tmnboq leased without a spec_id whose SPEC was authored after the lease"; the existing freeze text `its hints and pointers change only while picked` | the SPEC's setter refuses `leased` and later (spec.md:125; acceptance.md:106), the state of the observed wedge (D20) |
| P13 | Standing invariant the new issuance path touches | `git grep -n 'REQ-FCR-014' -- .moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md`; `leader_approval_gate.go:281-285` (`@MX:ANCHOR`) | spec.md:102 "No lane surface shall mint a receipt ..."; anchor reason "any second minting path would break the lane boundary this verb enforces"; `git grep 'FCR-014' -- .moai/specs/SPEC-TODO-UNIFY-001` printed nothing, exit 1 | the SPEC adds a lane-reachable minting path without naming the invariant it reverses (D23) |
| P14 | Lease acquisition needs no spec_id | read `card_transition.go:582-604` | `guardLeaseAcquire` checks registered worker and owner label only | confirmed; `assigned → leased` can proceed while a setter's half-write is pending (D21) |
| P15 | Time estimates and standing rules | `grep -n -i -E '[0-9]+ ?(days?|weeks?|hours?|minutes?)|as soon as possible'` over six artifacts | only false positives inside words such as "detail" and "T18"; none is a time estimate | no standing-rule violation: Template-First (REQ-TU-013), no test skip (AC-TU-024), no time estimates |

Judging builds (verification-claim-integrity §2.2). `moai spec lint` ran from a binary built from this tree by `go build -o <scratchpad>/moai-t1641-iter2 ./cmd/moai` at HEAD `3132f60c23` (plain build, build exit 0). `moai decision read`, `moai verify audit-plan`, `moai todo reflect` (E-09), and `moai todo show` ran on the installed build `moai-adk v3.2.0-rc.36-t1822c` (commit ee981a932, a strict ancestor of HEAD per codex's `build_lag`; the lag count was not re-measured here).

## Per-backend verdicts and receipts

- Audit plan: `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1641` printed `config_status: ok`, `cross_model_active: true`, `cross_model_required: true`, `enforced_required: ["codex"]`; claude gate off, glm gate off, codex gate required, all `explicit: true`.
- codex (required): `mcp__moai__codex_audit` with `project_root` set to this worktree, `mode: adversarial`, `target: baseBranch`. Returned `"verdict":"fail"` and `"audit_receipt":"rcpt-717ae7d0bd489fdeced99ba8"`, with four P1 and three P2 findings (titles quoted in Korean, translated here): P1 a committed PASS file treated as independent audit evidence, allowing the performer's self-approval (`design.md:88`); P1 an ancestor-only `audited_sha` check approves code added after the audit (`spec.md:109`); P1 the reused T16 callback does not bind the record content to the requested tree (`design.md:89`); P1 the TOCTOU between the three fact checks and the use of the approval is not closed (`design.md:83`); P2 a spec_id half-write followed by lease progress defeats the promised retry (`design.md:117`); P2 the warning hook runs before required-flag and flag-group validation (`design.md:28`); P2 the exact-32-verb criterion conflicts with the new `spec` verb (`acceptance.md:36`). I checked each against the tree: findings 1 to 3 and 5 to 7 are confirmed by my own reading (D16, D17, D21, D19, D18); finding 4 is real but is dominated by D16 and is carried as a required-fix clause of D16 (read the verdict from the merge commit's second-parent tree, not the card worktree HEAD). I did not reproduce codex's temporary-repository experiments; those parts rest on source reading.
- claude: gate off, not run. glm: gate off, not run.
- Convergence check: I wrote the digest of the single codex result to `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/207b7b17-aa81-4311-8e47-c01448800dee/scratchpad/audit-plan-result-iter2.json` (the lead asked for exactly one written report, so the digest is in the scratchpad rather than under `.moai/state`) and ran `moai verify audit-plan --project-root ... --result-file <that path>`. It printed `"convergence_check": {"ok": false, "unmet": [], "reason": "plan_source: \"\", want \"config\" (the server that produced the result did not read the configuration)"}`. This is not a gate result for the codex backend: my digest was reconstructed from `codex_audit`, which carries no `plan_source`, and `audit_multi` was not called (the lead named `codex_audit`). Read it as "the convergence check could not corroborate", not as an unmet backend (`unmet` is empty). The required leg itself is `fail`, which alone forbids a PASS.

## Defects Found

Severity vocabulary requested by the lead: BLOCKING, SHOULD-FIX, MINOR. Class per M6: blocking or optional. D1 to D15 are re-verified above; the list below continues the numbering for new or residual findings.

D16. spec.md:109; design.md:83, 88; acceptance.md:78-81 — Severity: BLOCKING — Class: blocking — Class code: design. Fact (2) of the automatic receipt is satisfiable with stale or self-authored evidence. The verdict file is read from the card's own worktree (`readAuditVerdict`, P6), which the lane writes, and the SPEC's only binding is "`audited_sha` is an ancestor of the integration branch": every older commit of the card satisfies that, so a PASS recorded at an early commit stays valid while later, unaudited code lands. The existing reader already binds `audited_sha` to the card's recorded `EvidenceSHA`, and the SPEC does not reuse that binding. AC-TU-016's rows (FAIL, uncommitted, non-ancestor) do not include "PASS at an ancestor, later change merged". Required fix: state in REQ-TU-017 fact (2) that the verdict file is read from the merge commit's second-parent tree (not the worktree HEAD, which also closes the check-to-use gap), that `audited_sha` equals the card's recorded evidence commit, and that the merged card tip equals that commit or differs from it by no tracked change (testable by tree equality); add the matching mutant rows to AC-TU-016 (PASS at an earlier ancestor with a later commit; a verdict file present only in the worktree). Prefer, if the leader's confirmed shape allows it, a fact the lane cannot author (a verdict that cites an audit receipt id the receipt store corroborates).

D17. spec.md:109 fact (3); design.md:89; acceptance.md:78-81 — Severity: BLOCKING — Class: blocking — Class code: design. "The re-measure record keyed to the merge commit's tree satisfies the gate the merge edge applied" inherits a verifier that never compares `rec.Tree` or `rec.Base` with the merge's tree and absorbed base (P7): a record stored under the right file name that names another tree validates. Required fix: add to fact (3) that the record's `Tree` equals the merge commit's tree and its `Base` equals the merge's first parent (or the stated absorbed base), and add a tree-mismatch row to AC-TU-016 (a record under the right key naming a different tree must issue nothing).

D18. acceptance.md:36-39; design.md:145; plan.md:112-113; research.md §D.3 table in design.md:46 — Severity: BLOCKING — Class: blocking — Class code: ac-wording. AC-TU-004 fixes the `todo` tree at exactly 26 + 6 = 32 verbs and the `gtd` tree at the same 32, while M3b registers the new `spec` verb on the same tree (plan.md:112; design.md §D.7; run order M2b, M3b, M4 at plan.md:33), giving 33. A correct build fails the AC; the mutant row ("the expected verb set is fixed in this criterion") makes it unsatisfiable. Design §D.3's S1 class also lists "the 26 plus the six stage verbs", omitting `spec`. Required fix: set the expected sets to the 33 names (26 present, six stage verbs, `spec`) in AC-TU-004 and design.md:46 and :145, and say that M3b lands before M4 so the count is read after both.

D19. spec.md:79; design.md:28; acceptance.md:28-30, 136 — Severity: BLOCKING — Class: blocking — Class code: design. REQ-TU-003 promises no warning for "an invocation refused by argument or flag validation". Required-flag and flag-group validation run at `command.go:1007` and `:1011`, after the persistent pre-run hook at `:985` that carries the warning (P1). Three of the six stage verbs the SPEC registers have required flags (`capture --event`, `clarify --disposition`, `organize --class`; P2), and `--auto`/`--auto-lane`/`--auto-leader` are mutually exclusive (`todo.go:356`). The existing test already feeds `capture x` as a refusal (`gtd_compat_test.go:166`). Those refusals will print the warning. AC-TU-002's three refusal rows (mistyped verb, missing required argument, unknown flag) omit them. Required fix: either print the warning from a point after required-flag and flag-group validation (for example in the command's run step), or narrow REQ-TU-003 to the cases the mechanism covers and state that required-flag and flag-group refusals carry the warning; add both refusal types as rows to AC-TU-002.

D20. spec.md:123, 125; acceptance.md:106; design.md:119; decision-index.md Q9 — Severity: SHOULD-FIX — Class: blocking — Class code: design. The SPEC's why (spec.md:57) and Q9's evidence cite the board record whose predicate is "a card ... leased without a spec_id whose SPEC was authored after the lease" (P12); kickoff reads the spec_id at `plan-audit → kickoff`, long after lease. REQ-TU-023 and AC-TU-021 refuse the setter for `leased` and later states, so the verb cannot repair the wedge it was written for; it only covers `picked` and `assigned` cards, where `next --spec` and `factory assign --spec` already work (the latter only while `picked`). The set-once rule already prevents the pointer swap that the freeze guards against. Required fix: admit the fill-if-empty case through the lease-holding states up to kickoff (and state the terminal-state refusal), or state explicitly that the leased wedge stays with the leader's database repair and drop the claim that the setter removes it from spec.md:57 and design.md §D.7; add the leased-card row to AC-TU-021.

D21. spec.md:123, 125; design.md:117; acceptance.md:114 — Severity: SHOULD-FIX — Class: blocking — Class code: design. REQ-TU-022 promises that a repeat with the same identifier "shall fill the empty carrier and succeed" after a half-write; REQ-TU-023 refuses any state other than `picked` or `assigned`. `guardLeaseAcquire` needs no spec_id (P14), so after a half-write the card can move to `leased` before the leader repeats, and the repeat is then refused with the factory carrier still empty. Required fix: resolved together with D20 (admitting the lease-holding states for a same-value repeat), or persist the recovery intent and block the lease edge while a half-written pair exists; add the lease-between-attempts interleaving to AC-TU-025.

D22. spec.md:109, 115; acceptance.md:69-72, 78-81 — Severity: SHOULD-FIX — Class: blocking — Class code: traceability. Two clauses of covered REQs have no AC row: REQ-TU-017's "the close shall write a log row naming the closing path and the receipt issuer" (AC-TU-014 asserts the closing path only; AC-TU-016 asserts the issuer on the receipt, not in the log), and REQ-TU-020's "a factory row in a state other than `merged-local` shall be left untouched and named in the log; its queue row closes only where the gate verifies a leader-issued receipt" (asserted only in plan.md M2b RED item 5 and the edge-case list at acceptance.md:142, not in an AC). This repeats the iteration-1 D8 pattern. Required fix: add one row each to AC-TU-014 (log row names path and issuer) and AC-TU-014 or AC-TU-016 (a `pushed` factory row stays `pushed`, is named in the log, and closes only under a leader-issued receipt).

D23. spec.md:109; design.md:91; `.moai/specs/SPEC-FACTORY-COMPLETION-RECOVERY-001/spec.md:102` — Severity: SHOULD-FIX — Class: optional — Class code: other. The automatic issuance path reachable from a lane's `todo done` reverses REQ-FCR-014 ("No lane surface shall mint a receipt") and the `@MX:ANCHOR` reason at `leader_approval_gate.go:281-285` ("any second minting path would break the lane boundary"). The SPEC cites REQ-FCR-010 and REQ-FCR-002 of that draft SPEC but never REQ-FCR-014, and plan.md M2a does not schedule updating the anchor. The leader's confirmation covers the design; the SPEC should name the reversal. Required fix: name REQ-FCR-014 in the HISTORY, REQ-TU-017 rationale, and plan §H; add the anchor-text update to M2a's file list.

D24. acceptance.md:25; research.md E-15; plan.md:130 — Severity: MINOR — Class: optional — Class code: ac-wording. E-15 greps only `internal/cli/gtd.go`, but M5 puts the warning in "a small warning helper" (plan.md:130); if the phrase lands there, the proxy never flips. Fix: widen the pathspec to `internal/cli` or name the file.

D25. plan.md:143-146, 148-152; acceptance.md:62-65 — Severity: MINOR — Class: optional — Class code: other. M7 lands `TestNoResidualGTDSpelling`, which stays red until M8 converts the 73 paths, and AC-TU-024 forbids a skip. Fix: state that M7 and M8 land together, or define the test's M7 corpus as the converted set only.

D26. design.md:103; `card_record.go:48-49` — Severity: MINOR — Class: optional — Class code: other. T29 admits decider `audit` (and `human`), but `DeciderAudit` is documented as "edge T8a only", REQ-TU-019 names no decider, and AC-TU-018 has no decider row, so a mutant that admits any decider passes. Fix: state the admitted deciders in REQ-TU-019 and add a decider row.

D27. plan.md:89-97 — Severity: MINOR — Class: optional — Class code: other. M2a (about 450 lines) and M2b (about 520) exceed the plan's own 400-line split rule at estimate level; each carries a split clause, so this is a scheduling nit.

D28. spec.md:111, 113, 119 — Severity: MINOR — Class: optional — Class code: format. REQ-TU-025 precedes REQ-TU-022 in the file; REQ-TU-018 and 019 name table identifiers (T18, T29) although design.md:3 says identifiers live in design.

## Regression Check

Defects from iteration 1:

- D1: RESOLVED (0 markers, seven files).
- D2: RESOLVED (defaults replaced, DoD consistent with the criteria).
- D3: RESOLVED as to oracles; the sufficiency of the conditions is the new D16 and D17.
- D4: RESOLVED (T29 declared before removal; AC-TU-018 asserts a terminal state).
- D5: RESOLVED for the close; PARTLY for the setter (D21 is the residual).
- D6: RESOLVED (graph); a new inconsistency D18 arises from M3b.
- D7: PARTLY RESOLVED (D19 is the residual).
- D8: RESOLVED.
- D9: RESOLVED as to selectors and mutants; the fixed set is itself wrong (D18).
- D10: RESOLVED. D11: RESOLVED. D12: RESOLVED. D13: RESOLVED (nit D24). D14: RESOLVED (nit D28).
- D15: PARTLY RESOLVED (D27).

Introduced by the revision (regressions in the sense of new defects in changed text): D16, D17, D18, D19, D20, D21, D22. No iteration-1 blocker reappears unchanged. The CN-4 ordering verb was re-run in full (as a Python port), and MP-9 was re-decided from it: no conflict.

## Recommendation

1. Fix D16 and D17 first (they are the trust model of the automatic receipt and are testable edits to REQ-TU-017 and AC-TU-016): bind fact (2) to the merge's second-parent tree and the card's recorded evidence commit, bind fact (3) to the merge tree and base, and add the three mutant rows.
2. Fix D18 (set the expected verb set to 33 names) and D19 (move the warning after required-flag validation or narrow REQ-TU-003, and add the rows to AC-TU-002).
3. Decide D20 and D21 together: either admit the fill-if-empty setter through the lease-holding states, or state that the leased wedge stays outside this SPEC and correct spec.md:57 and design.md §D.7; then add the interleaving row to AC-TU-025.
4. Close D22 with two AC rows; name REQ-FCR-014 (D23); the MINOR items D24 to D28 are the author's discretion.
5. Re-audit as iteration 3 (the last under the Tier L ceiling), scoped to the defect delta above plus the full must-pass set and the CN-4 ordering check.

## Plan-artifact hash

The runtime `InMemoryCache.ComputeHash` (internal/runtime/audit_cache.go) hashes, in order, acceptance.md, decision-index.md, design.md, plan.md, research.md, spec.md, tasks.md, skipping absent files, writing `%s:%d\x00` (name, length), the bytes, then one `\x00`. Present files and byte lengths: acceptance.md 28297, decision-index.md 21938, design.md 27423, plan.md 30040, research.md 27943, spec.md 22312; tasks.md absent; progress.md is not included.

- Implementation 1 (Python 3, `python3 -I iter2_hash.py <spec dir>`, scratchpad): `966727e2c52676fcc89f1f282c136e8f25c310c4716c578cb0de929142216304`
- Implementation 2 (Go, `go run main.go <spec dir>` in `<scratchpad>/hashgo`, a transcription of the runtime function without the receiver): `966727e2c52676fcc89f1f282c136e8f25c310c4716c578cb0de929142216304`
- Agreement: yes, identical. Files read from the clean working tree at HEAD `3132f60c23afeb620415f9394c5eff45c4e03191`.

## Gaps

- Refused tool (verification-claim-integrity §3.1): my first attempt to run the two Python port scripts with a shell variable in the path was refused by the worktree guard ("a computed argument ... cannot be shown not to be git"). I reran them with literal absolute paths and the results above are from those runs. Iteration 1 recorded that `sh` and `awk` were refused, so I did not attempt the awk verbs this iteration: the traceability, CN-4, D7 and D8 verbs ran as my own Python ports (a transcription, not the verbs' text). The traceability port collected 0 definitions because the SPEC's bold-paragraph form is not one of the three forms the verb reads; the matrix is manual.
- No `go test` and no heavy-slot use, per the brief. Every statement about test behaviour (the 71 and 370 pin, refusal behaviour, the factory-first retry, interleavings) is a static reading of source, not a run.
- Codex's temporary-repository counterexamples (self-authored PASS file, A-pass-B-lands, mismatched record, FAIL swap) were not reproduced by me; my confirmation of D16 and D17 is by reading `readAuditVerdict`, `ReadRemeasureRecord`, `ValidateRemeasureRecord`, and `VerifyBinding`.
- E-10 was not re-run; E-09 was reproduced on the installed (older) build only, with its exit code unobserved; the 516-commit lag of the installed `moai` was not re-measured.
- `mcp__moai__audit_multi` was not called; the convergence check ran on a digest I built from the single `codex_audit` result and reported `ok: false` for an empty `plan_source` (see above). It is not evidence about the codex backend either way.
- The leader's confirmation of Q5 is the coordinator's message, not a board record; I did not see the message, and the board holds no record that rules on Q1 to Q9 for this card (the one cited record, `d-20261010T001732Z-d4e8`, was read and bears on Q9 only).
- Not examined: the docs-site build, the board mod's JavaScript tests, whether each of the 67 conversion paths has a template source (plan §C item 7 does it at run time), whether `sh` exists on every CI runner, `REQ-FCR-010` current text beyond the cited line, and the claim that the deployed `.claude/` copies "follow the regenerated output" at M8b.
- Provenance of the tool measurements: see "Judging builds" above; the installed build lags the tree.

## Residual-risk

- After D16 and D17 are fixed, the three facts are still produced inside the lane's own worktree and the lane's own re-measure verb; they bind to the merged content, but they do not establish that an independent auditor ran. That is the leader-confirmed trust shape; the revocation of REQ-TU-025 is the compensating control, and its use depends on a leader noticing.
- In the local-main flow used by this repository (a remote exists, lanes merge to local `main`, `origin/main` advances by batch), REQ-TU-019 with a remote requires reachability from the landed ref, which resolves to `origin/<base>` (`prlink_landedref.go`, `LandedRefForWithLevel`): a `merged-local` card closes only after its batch lands. The doctrine line "a card's done does not wait for the batch" (gitflow-lane-protocol.md §4) is attributed to card t1621, so the ten-card live measurement at sync may count cards by batch landing, not by lane merge. Not a defect of this SPEC's own terms; a reader of the ten-card criterion should know it.
- `moai todo` and `moai gtd` share one tree, so every verb added later is also on the alias; the warning's removal statement is one constant, and its removal is outside this SPEC.
- The T29 `pushed` gap persists by design: a card whose factory row the leader already moved to `pushed` has no completion edge until card t1639 declares one (the SPEC says so at spec.md:148).
- The M7 and M8 conversion touches about 67 files in four locales; the class table is only as good as the allowlist, which is built at run time.

## Operational Notes (unverified)

- Measure whether required-flag refusals print the alias warning: after M5, run `<moai built from the tree> gtd capture x` against a temporary queue and capture stderr. Status: assumption (inferred from cobra `command.go:985` versus `:1007`).
- Measure D16: with a temporary repository, commit a PASS `sync-audit.md` naming an early ancestor, add one more commit to the card branch, merge, and run the three-fact check. Status: assumption (codex reported the same counterexample; I read the code, I did not run it).
- Measure the registered verb count after M3b and M4: list the `todo` command's subcommand names. Status: inferred (26 counted from `todo.go:342-349`, plus six stage verbs and `spec`).
- Re-measure the live ten-card criterion at sync with the exact listing command recorded in the sync record. Status: assumption.

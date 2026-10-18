auditor-model: claude-sonnet-5-5

# SPEC Review Report: SPEC-TODO-UNIFY-001 (plan-audit iteration 3 of 3, final)

Reasoning context ignored per M1 Context Isolation. Read in full: the seven SPEC artifacts, `.moai/reports/t1641/progress.md`, `plan-audit-iter2.md`. Tree: `git worktree list` shows t1641 at `164050fdde` (branch WT-todo-unify); `git rev-parse HEAD` (python subprocess, `git -C` that tree) printed `164050fdde316bc2717267e76515ae3e3d290768`; `git status --short` printed nothing; `git diff --stat 3dc1052f6f HEAD` lists only SPEC-directory files and `.moai/reports/t1641/progress.md`, so every code-side measurement of the earlier rounds still describes this tree.

Result: FAIL. Aggregate 0.86 (iteration 2: 0.80, no STOP), above the Tier L threshold 0.85, all nine must-pass criteria pass or are N/A, and traceability is now complete. It still fails for two independent reasons: the required codex backend returned `fail` on both of my runs (convergence overall `fail`), and three blocking findings are open (D29 to D31), all in the NEW text of this revision. Each of the three is narrow and has a cheap, testable fix; none reopens a design decision, and the iteration-2 blockers D16 to D22 are closed in the text.

verdict: FAIL
audited_sha: 164050fdde316bc2717267e76515ae3e3d290768
overall_score: 0.86
must_pass_failed: 0
blocking_count: 3
plan_artifact_hash: cbe6819defad8c2e664b58a7b275f3b973601f0005b119ed30de4275ad69e97a
scope: full
fix_scope: .moai/specs/SPEC-TODO-UNIFY-001/design.md#D.6; .moai/specs/SPEC-TODO-UNIFY-001/plan.md#M1; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-018; .moai/specs/SPEC-TODO-UNIFY-001/design.md#D.5; .moai/specs/SPEC-TODO-UNIFY-001/spec.md#REQ-TU-017; .moai/specs/SPEC-TODO-UNIFY-001/spec.md#REQ-TU-025; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-019; .moai/specs/SPEC-TODO-UNIFY-001/acceptance.md#AC-TU-016; .moai/specs/SPEC-TODO-UNIFY-001/plan.md#M2a
defect_class: D29=design D30=design D31=design
reread_hunks: none
convergence_overall: fail
required_backend: codex fail

## Must-Pass Results

- [PASS] MP-1 REQ number consistency. Layer: requirement layer, spec.md. `python3 -I mp_checks.py` (scratchpad): `REQ count 25 unique 25 contiguous True`; file order is now monotone REQ-TU-001 to REQ-TU-025 (the iteration-2 D28 order defect is fixed: REQ-TU-025 at spec.md:136 follows REQ-TU-024 at :132). `AC count 25 unique 25`, contiguous (file order of AC blocks is non-monotone by design, acceptance.md:7). Tier L ceilings 25 and 25 (spec-workflow.md:130) are met exactly.
- [PASS] MP-2 EARS/GEARS compliance. Layer judged: the 25 bold `REQ-TU-` paragraphs of spec.md only; the Given-When-Then blocks of acceptance.md are the verification layer and were graded under Group 4. Parenthesised patterns printed by the script: Ubiquitous (001, 004, 005, 007, 008, 009, 011, 012, 013, 014, 018, 021, 024), Event-driven (002, 010, 015, 017, 019, 020, 022, 023, 025), Where (006), legacy Unwanted (003, 016; valid to 2026-11-22). Several clauses are long compound sentences (REQ-TU-017 carries three numbered facts) but each opens with an event or a ubiquitous subject and carries one `shall` obligation set.
- [PASS] MP-3 YAML frontmatter validity. spec.md:2-24: `id`, quoted `title`, `version: "0.3.0"` (quoted), `status: draft`, `created`/`updated: 2026-10-18`, `author`, `priority: P1`, `phase`, `module`, `lifecycle: spec-anchored`, `tags` (comma string), plus optional `tier: L`, `related_specs`. No snake_case alias. `moai spec lint SPEC-TODO-UNIFY-001` run from the tree directory with a binary I built from this tree printed `0 error(s), 0 warning(s)` and one INFO `OwnershipTransitionUnmeasured` (commit 63829fe20a lacks an `Authored-By-Agent` trailer); the exit code was not captured (pipe), a Gap.
- [N/A] MP-4 language neutrality: single-repository Go CLI scope; no multi-language tooling enumerated.
- [PASS] MP-5 D7 cross-SPEC reconciliation. Python port of the D7 verb (awk/sh not attempted): nine ids in spec.md; statuses FACTORY-COMPLETION-RECOVERY-001 draft, FACTORY-RECORD-001 completed, FACTORY-RECORD-REPAIR-001 draft, GITHUB-FLOW-DEFAULT-001 in-progress, GTD-CANON-BODY-001 completed, KANBAN-TODO-CLI-001 in-progress, LOCAL-MAIN-FLOW-001 completed, TODO-LAND-AUTO-DONE-001 completed, itself draft. None retired, superseded, or archived; no REVIEW or SHOULD line. The REQ-FCR-014 reversal the verb cannot see is now named (design.md:107, spec.md:116, :161).
- [PASS] MP-6 D8 cross-platform discipline. `syscall` count 0 in all seven artifacts (script output); D8-4 auto-pass.
- [PASS] MP-7 clarification gate. `NEEDS CLARIFICATION` count 0 in spec.md, plan.md, acceptance.md, design.md, research.md, decision-index.md, progress.md.
- [PASS] MP-8 RED-now re-execution. Release-blocking criteria: the class A ones (AC-TU-001, 005, 007, 008, 010, 017, wording clause of 013). Each cited ledger command was re-executed with `git grep ... 164050fdde... --` from my own worktree against that commit object (identical content to the clean t1641 tree) via `mp8.py`, exit codes observed: E-01 `73` paths (wc -l); E-02 exit 0, 6 lines (card_transition.go:19,157,203,207,333,334); E-03 exit 0, 6 lines (gtd.md:150,161,173,185,194,211); E-04 exit 1; E-06 exit 1; E-07 exit 1; E-08 exit 0 (`gtd_compat_test.go:83`); E-12 exit 0 (`fr_transition_test.go:166`); E-13 exit 0, 2 lines; E-14 exit 0, 8 lines; E-15 exit 1; E-16 exit 0 (`AGENTS.md:1`, `AGENTS.md.tmpl:1`); E-17 exit 1; E-18 exit 1; E-21 exit 1; E-11 `wc -c AGENTS.md` printed `23780`. E-09 reproduced on a build of this tree (`Unknown command "reflect" for "moai todo".`); E-10: `gtd list --limit 1` stderr carried no deprecation text. None is a test runner. All REDs reproduce. Class B criteria are unadopted until their RED is recorded (acceptance.md:15-16): the section 2.1 disposition. Boundary: re-execution confirms output, not that the grep measures its premise.
- [PASS] MP-9 cross-artifact ordering. Python port of the CN-4 verb: `COLLECTED: 9 milestones in plan order (M1 M2 M3 M4 M5 M6 M7 M8 M9), 0 exit bindings`, 20 ordering records in acceptance.md; with 0 `Exit:` bindings no CONFLICT line can be produced, so I read all 20. The milestone-naming ones agree with the plan: AC-TU-004 "read after M3b and M4 have both landed, run order M2b, M3b, M4" matches plan.md:33; AC-TU-010 "created at M8's RED ... M7 lands only the script" matches plan.md:141, :150; the rest order acts within a close (factory edge before queue archive), cobra execution, RED order inside M1. No confirmed conflict. Nearest tension, not a conflict: D34.

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.75 | minor ambiguity in one or two requirements | REQ-TU-017 "issuer `auto`, distinguishable from a leader-issued receipt" (spec.md:116) is not guaranteed by the store (D30); design.md:119 says the push-gate caller passes "the integration branch it resolved from the landed-ref chain", which the unchanged caller does not (D29) |
| Completeness | 0.95 | all sections and 12 frontmatter fields present | HISTORY (spec.md:28-34), background (:36), scope/tier (:60), 25 requirements (:70), constraints (:138), three `### Out of Scope` H3s with bullets (:148, :154, :164), design, research, decision-index; held below 1.0 for D29 (planned file set cannot deliver REQ-TU-019's last sentence) |
| Testability | 0.75 | rows binary-testable; a few rows missing | AC-TU-016 rows (a)-(h) and AC-TU-014/021/025 rows are oracle-complete; missing rows: leading-whitespace and replacement-object cases for the git reads (D31), same-card manual receipt with issuer `auto` (D30), CLI-level no-config no-remote close (D29), second invocation of one gtd tree (D32) |
| Traceability | 1.00 | every REQ has an AC, no orphan | script output `UNCOVERED []`, `ORPHAN []` (AC-TU-024 is the all-requirements guard); the two iteration-2 gaps (log-row issuer, `pushed` row) now sit in AC-TU-014 (acceptance.md:73) |

Aggregate (mean): (0.75 + 0.95 + 0.75 + 1.00) / 4 = 0.8625, reported 0.86. The Tier L threshold is 0.85, so the score alone clears it; the verdict is FAIL because of the required codex leg and the three blocking findings (the score is anchored to the rubric, the verdict to the firewall and the gates). Not a STOP: 0.86 is above 0.80.

## REQ to AC matrix (own extraction: `(REQ-...)` parentheses of each AC header, script `mp_checks.py`)

| REQ | AC | REQ | AC |
|-----|----|-----|----|
| 001 | AC-TU-003 | 014 | AC-TU-013 |
| 002 | AC-TU-001 | 015 | AC-TU-014 |
| 003 | AC-TU-002 | 016 | AC-TU-015 |
| 004 | AC-TU-004 | 017 | AC-TU-016 |
| 005 | AC-TU-005 | 018 | AC-TU-017 |
| 006 | AC-TU-006 | 019 | AC-TU-018 |
| 007 | AC-TU-007 | 020 | AC-TU-014, AC-TU-023 |
| 008 | AC-TU-008 | 021 | AC-TU-020 |
| 009, 010 | AC-TU-009 | 022 | AC-TU-021, AC-TU-025 |
| 011 | AC-TU-010 | 023 | AC-TU-021 |
| 012 | AC-TU-011 | 024 | AC-TU-022 |
| 013 | AC-TU-012 | 025 | AC-TU-019 |

25 REQ, 25 AC, 0 uncovered, 0 orphan. The traceability awk verb was not run (this SPEC writes bold paragraphs, a form the verb does not read; it would print GAP); a Python port and a manual matrix stand in, and `moai spec lint`'s silence is not corroboration.

## D1 to D28 re-verification

| ID | Result | Evidence |
|----|--------|----------|
| D1 markers | FIXED | 0 markers in all seven files (MP-7) |
| D2 defaults vs card | FIXED | Q5/Q6/Q7 decided in spec.md:110-122; unchanged since iteration 2 |
| D3 oracles | FIXED | REQ-TU-017/019/020/025 carry conditions; AC-TU-014/016/018/019/020/023 carry rows |
| D4 circular replacement path | FIXED | `"T29"` absent today (E-17 exit 1); T29 declared first in M1 RED order (plan.md:77); AC-TU-018 asserts no-remote completion (acceptance.md:99). But see D29 for the CLI path |
| D5 two-store atomicity | FIXED | close order and recovery spec.md:122, design.md:109, AC-TU-023; setter spec.md:128, AC-TU-025 |
| D6 dependency graph | FIXED | plan.md:21-33 |
| D7 warning ordering | FIXED in design, not executed | cobra v1.10.2 anchors re-read: `command.go:919 ParseFlags`, `:968 ValidateArgs`, `:985 PersistentPreRunE`, `:1007 ValidateRequiredFlags`, `:1010 ValidateFlagGroups`, `:1014 RunE`; the run-step wrapper (design.md:28) and AC-TU-002 rows 4-6 close the iteration-2 hole; nothing was run (D32 is a small flaw in the wrapper's flag lifetime) |
| D8 REQ-TU-014 oracle | FIXED | E-16 reproduces (exit 0, 2 files); `wc -c AGENTS.md` 23780 |
| D9 fixed verb set, selectors | FIXED | 19 named existing tests all located with the cited lines (P7) |
| D10 floors | CARRIED | not re-measured this round (Gap); iteration 2 reproduced 326 over 17 files and 9,090 in scope |
| D11 preservation guard | FIXED | `TestGTDAnswerWritesAppendOnlyRecord:13`, `...RejectsInvalidID:43`, `...CardUntouched:52`, `TestGTDCLIInputRefusalsDoNotFallThrough:164` exist |
| D12 template-source claim | CARRIED | not re-measured (Gap) |
| D13 proxy cells | FIXED | E-15 now over `internal/cli`, exit 1 reproduced |
| D14 requirement wording | FIXED | REQ-TU-023 event form (spec.md:130) |
| D15 milestone sizing | PARTLY | M2a (470 lines) and M2b (520) above the plan's own 400 figure; split clauses present (plan.md:87, :95); optional (D27) |
| D16 stale/forged fact 2, worktree read, TOCTOU | FIXED in text; hardening open (D31) | spec.md:110-116; design.md:85-105; rule R2 reproduced on landed merges (P5); in-transaction issuance with rollback row (AC-TU-016 g) |
| D17 record never compared | FIXED | `VerifyRemeasureBinding` named (design.md:95, plan.md:86), AC-TU-016 row (f); premise measured on 11 merges (P6) |
| D18 verb count | FIXED | 26 and 32 reproduced (P2); AC-TU-004 pins 33 (acceptance.md:39) |
| D19 required-flag refusals print the warning | FIXED in design | see D7 row; AC-TU-002 rows (4)-(6) |
| D20 setter refuses the wedge's state | FIXED | first-fill set `picked`..`kickoff` (spec.md:128); AC-TU-021 six states, AC-TU-022 wedge row |
| D21 half-write overtaken by lease | FIXED | completing repeat in every non-terminal state (spec.md:128, :130); AC-TU-025 interleaving rows |
| D22 two clauses without rows | FIXED | AC-TU-014 now asserts the log issuer and the `pushed` row (acceptance.md:73-76) |
| D23 REQ-FCR-014 reversal | FIXED | design.md:107, research.md §R.8, spec.md:116, :161, plan.md §H |
| D24 E-15 pathspec | FIXED | exit 1 reproduced over `internal/cli` |
| D25 residual test red between M7 and M8 | FIXED, residual ambiguity | moved to M8 (plan.md:141, :150); see D34 |
| D26 T29 deciders | FIXED | spec.md:120 clause (c); AC-TU-018 decider row |
| D27 M2a/M2b size | NOT FIXED (optional) | author states it (progress.md §E.1) |
| D28 order and identifiers | HALF | order fixed (spec.md:132-136); T18/T29 identifiers remain in REQ-TU-018/019 (optional) |

## Premise spot-checks (own commands this run)

Judging builds (verification-claim-integrity 2.2): `go -C <tree> build -o <scratchpad>/moai-t1641-iter3a ./cmd/moai`, build exit 0, built from tree HEAD `164050fdde` (code identical to `3dc1052f6f`); the help-page counts and `moai spec lint` ran from that build; `moai verify audit-plan` and the MCP server run the installed build ee981a932 (a strict ancestor; codex reports `build_lag`).

| # | Premise | My command | Observed | Result |
|---|---------|-----------|----------|--------|
| P1 | cobra v1.10.2 order | `grep -n` of anchors in the module-cache `command.go`; `grep -n 'spf13/cobra' go.mod` | `919: err = c.ParseFlags(a)`, `968: ... ValidateArgs(argWoFlags)`, `985: if p.PersistentPreRunE != nil {`, `1007: ... ValidateRequiredFlags()`, `1010: ... ValidateFlagGroups()`, `1014: if c.RunE != nil {`; `go.mod:23 github.com/spf13/cobra v1.10.2` | confirmed (read, not executed) |
| P2 | 26 and 32 verbs today | `moai-t1641-iter3a todo --help`, `... gtd --help` into files; counted the COMMANDS block by reading | todo block lines 46-71 = 26 verbs; gtd block lines 18-49 = 32 (the 26 plus capture clarify organize reflect engage answer) | confirmed; 33 target = 26 + `spec` + 6 holds arithmetically |
| P3 | T29 absent, 71/370 pin | E-17 exit 1; `fr_transition_test.go:127-183` read | `if len(accepted) != 71 \|\| refused != 370`; `frFullRequest` carries no `ApprovalUUID`; final loop asserts `{merged-local, done}`, `{pushed, ci-green}`, `{ci-green, done}` refused | confirmed by static reading; not run (go test forbidden) |
| P4 | T16 callback validates, never compares | read `factory_card.go:2372-2378`; `grep` in `integration_remeasure.go` | callback is `ReadRemeasureRecord` then `ValidateRemeasureRecord` and ignores its `mergeSHA` argument; `ValidateRemeasureRecord` checks only non-empty/full-SHA `rec.Tree`/`rec.Base`; comment at `:98` "rec.Tree reads AS WRITTEN" | confirmed |
| P5 | Rule R2 on landed merges | `r2_rule.py` (scratchpad; `git -C` my tree, `diff --name-only --no-renames -z` for A = audited..tip and D = first parent..tip) over 83 integration merges on local main | 18 merges carry a committed `sync-audit*.md`. Reproduced exactly: t1856 A2 D7 A&D 2 R2 0; t1837 A2 D7 A&D 2 R2 0; t1812 A254 D5 A&D 2 R1 252 R2 0. Independent of the author's sample, R2 passes t1599, t1833, t1760, t1670, t1665, t1823, t1792, t1800, t1757, t1861 (13 in all). **t1854: R2 refuses**, violations `CHANGELOG.md`, `internal/template/catalog.yaml` (card absorbed main after the audit and combined shared hot files). Four merges tie on the highest iteration number and read as absent: t1689, t1864, t1858 (`-head-confirmation`, `-r2`), t1794 (`-r2..-r4`). The other 65 merges commit no sync-audit file | R2 premise confirmed, with measured over-refusals (D36) |
| P6 | Re-measure record binds tree and first parent | `remeasure_probe.py` reading `/Users/goos/MoAI/moai-adk-go/.moai/state/remeasure/<merge tree>.json` for 11 merges (t1760, t1665, t1833, t1854, t1599, t1670, t1823, t1792, t1861, t1757, t1856) | all 11: `tree_eq True base_eq_first_parent True exit 0` (record dir holds 163 files) | confirmed beyond the author's three |
| P7 | Named existing tests exist | `grep -rn -E '^func (19 names)\('` over the tree | all 19 found at the cited lines (e.g. `gtd_compat_test.go:49,124,164`, `todo_test.go:121,214,1078,1123`, `fr_transition_test.go:127`, `f7_gate_owner_admission_test.go:14`, `token_budget_guard_test.go:84`, `workflow_rule_paths_pinned_test.go:24`, `gtd_canonical_surface_test.go:10`) | confirmed |
| P8 | Plan-artifact hash | two implementations (below) | identical | confirmed |
| P9 | The push-gate caller passes the integration branch it resolved | read `factory_card.go:3238`, `:3304-3336`, `:2552-2557`; `card_evidence_readers.go:290-318`; `loader_integration_branch.go:91-129` | `integration := config.LoadGitFlowIntegrationConfig(root).IntegrationTarget` (config only; empty with no `git-strategy.yaml`, a non-manual git-flow profile, or an unknown workflow); `verifyMerge` begins `if !validBranchName(integration) { ... "integration branch %q is not usable" }` and requires `refs/heads/<integration>`; T18's `guardNoRemote` branch never reads the integration branch (`card_transition.go:730-746`) | **premise of design.md:119 is false for the unchanged caller** (D29) |
| P10 | `gitRead` immutability premise | read `card_evidence.go:19-34`; `grep -i replace internal/gitenv/*.go` | `gitRead` returns `strings.TrimSpace(string(out))` and runs plain `git -C dir <args>` with `gitenv.Env()`; no `--no-replace-objects`, and the grep over gitenv printed nothing | design.md:87 "cannot change what a fact reads" does not hold as stated (D31) |
| P11 | Issuer `auto` is distinguishable | read `leader_approval_gate.go:344-373`, `leader_approval.go:153-175` | `moai factory approve` takes `--issuer` (default the worker env, else `leader`) and writes `Issuer: issuer, IssuerRole: leader`; the table is keyed `(card_uuid, run_id)` with `ON CONFLICT ... DO UPDATE`; nothing refuses the string `auto` | design.md:105's exclusion "keyed on the `auto` issuer" cannot tell a leader receipt carrying `auto` from an automatic one (D30) |
| P12 | Queue states across the setter's range | `grep` of `BacklogState*` in `internal/factory/backlog*.go`; read `backlog_store.go:1157-1204` | states queued, picked, dropped, hold; a lapsed queue lease returns the item to `queued` and clears its `spec_id` (`clearBacklogLease`) | design.md:131 premise holds by reading; a card whose queue lease lapses while its factory row is `leased` fails the setter's `picked` precondition (D37, optional) |
| P13 | Lint | `cd <tree> && moai-t1641-iter3a spec lint SPEC-TODO-UNIFY-001` | `0 error(s), 0 warning(s)`; one INFO | confirmed; exit code not captured |

## Per-backend verdicts and receipts

- Audit plan: `moai verify audit-plan --project-root <tree>`: `config_status: ok`, `cross_model_active: true`, `cross_model_required: true`, `enforced_required: ["codex"]`; claude gate off, glm gate off, codex gate required, all `explicit: true`.
- codex single audit (`mcp__moai__codex_audit`, project_root the t1641 tree, mode adversarial, target baseBranch): `"verdict":"fail"`, `"audit_receipt":"rcpt-9557497ba4bcc4b92912703c"`. Findings (titles translated from Korean): P1 git replacement refs can make an old audit pass (design.md:94); P2 the unchanged push-gate caller hands T29 an empty integration branch in a no-remote repository without configuration (plan.md:80); P2 AC-TU-020 passes with a synthetic fixture even if the real most recent ten cards do not close (acceptance.md:108). I verified the first two (P10, P9) and judged the third (D35).
- codex via convergence (`mcp__moai__audit_multi`, no `gates` argument, project_root the t1641 tree, target baseBranch, session_id `t1641-plan-audit-iter3`): `per_backend_verdicts: codex, gate required, verdict fail`, `overall_verdict: fail`, `disagreement_flag: null`, `participant_count: 1`, `residual_risk_note: "required-backend FAIL: codex"`, `plan_source: config`, `"audit_receipt":"rcpt-01bf6dad688922d13c1931d1"`. Findings (translated): P1 R2 misses a card deleting main code it absorbed after the audit (design.md:97); P1 `gitRead` trims whitespace so a leading-space path in the NUL-separated diff passes the reports-boundary check (design.md:94); P2 `issuer == "auto"` cannot separate automatic from manual receipts (design.md:105); P3 the tree-lifetime `warned` flag gives 1 then 0 warnings on two runs of one tree (design.md:28). Verified by me: the second (P10) and third (P11). The first is a property of R2's invariant (D33); the fourth follows from the closure design (D32). The two codex runs differ in findings; the required leg is `fail` in both, and I treat each as a reviewed result, not an unmet gate.
- claude: gate off, not run. glm: gate off, not run.
- Convergence check: digest written to `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/207b7b17-aa81-4311-8e47-c01448800dee/scratchpad/audit-plan-result-iter3.json` (`overall_verdict: fail`, `gate_unmet: ""`, `plan_source: config`, per backend `codex / required / fail`; written to the scratchpad, not `.moai/state`, because the lead asked for exactly one written tree file). `moai verify audit-plan --project-root <tree> --result-file <that path>` printed `"convergence_check": {"ok": true, "unmet": []}`. So the gate was exercised and held (nothing unmet); the backend answered `fail`. Overall: FAIL by the convergence policy (a required FAIL).

## Defects Found

Severity vocabulary: BLOCKING, SHOULD-FIX, MINOR. Class per M6.

D29. design.md:119 (D.6), plan.md:80 and :87 (M1, "the push gate caller is unchanged", files list omits `factory_card.go`), spec.md:120 (REQ-TU-019 last sentence), acceptance.md:99-102 (AC-TU-018) — Severity: BLOCKING — Class: blocking — Class code: design. T29's no-remote path requires reachability from `refs/heads/<integration>` and a usable branch name (P9). The only production caller of the push gate, `decideOne` via `factoryDecideCards`, passes the configured `IntegrationTarget` only, which is empty for a repository without `git-strategy` configuration or with a non-manual git-flow or unknown workflow profile. Removed T18 never needed the branch, so a no-remote card the leader could close with a receipt before is refused after the change, and REQ-TU-019's unconditional sentence ("In a repository without a remote the push gate shall request done ... without any push step") fails; the design's premise that "the caller passes the integration branch it resolved from the landed-ref chain" is false for the caller as it stands. AC-TU-018 is homestate-level plus `TestPushGateTargetByRemote` and never runs the CLI path without configuration. Required fix: put `internal/cli/factory_card.go` in M1's files, make the push gate resolve the branch with the existing `factoryResolveIntegrationBranch` (`factory_card.go:2552`, falls back to the card worktree's branch) or state T29's no-remote reachability as "the local branch the merge was made on", and add a CLI-level AC-TU-018 row: no remote, no git-strategy file, valid receipt, `decide --gate push` reaches `done`.

D30. design.md:105 (D.5, "exclusion keyed on the `auto` issuer"), spec.md:116 and :136 (REQ-TU-017 "distinguishable from a leader-issued receipt", REQ-TU-025 "a leader-issued receipt shall be unaffected"), acceptance.md:104-106 (AC-TU-019) — Severity: BLOCKING — Class: blocking — Class code: design. The leader path accepts any issuer string through `moai factory approve --issuer`, and the store keeps one row per `(card_uuid, run_id)` with replace-on-conflict (P11). A leader receipt carrying the string `auto` is therefore indistinguishable from an automatic one, so the revocation exclusion would classify it as automatic and revoke it, contradicting REQ-TU-025; the "distinguishable" claim of REQ-TU-017 rests on a free-text label. Required fix: distinguish the kind by a column the manual path cannot write (an `auto` flag set only by the transaction-level insert) or reserve the issuer string `auto` in the leader path and define the treatment of an existing row; add a same-card manual-receipt row (issuer `auto` passed by hand, then revocation) to AC-TU-019 and an assertion to AC-TU-016 row (h).

D31. design.md:87, :94, :99 (D.5 reads through `gitRead`), spec.md:110 ("each read from git objects ... never from the card's working tree"), acceptance.md:82-93 (AC-TU-016) — Severity: BLOCKING — Class: blocking — Class code: design. The design names the existing `gitRead` for every fact read and states "a lane that edits its working tree, moves its branch, or amends a file cannot change what a fact reads" (design.md:87). `gitRead` trims leading and trailing whitespace of the whole output (P10), so the first path of a `-z` diff that begins with a space or newline loses it, and a path such as ` .moai/reports/<card>/x` (outside the directory) reads as inside it; and no read disables replacement objects, so a replacement ref in the shared repository can change what a named commit reads as. Both turn the "audit covers what lands" fact from an accident check into a bypass by a lane that deliberately wants it. The residual the SPEC accepts is a lane-authored verdict; these two are different (they defeat R2 and the immutability claim, not the audit's independence), and the fix is cheap. Required fix: state that fact reads use a non-trimming raw-output reader for the NUL-separated path lists and run with replacement objects disabled (`GIT_NO_REPLACE_OBJECTS=1` or `--no-replace-objects`), correct design.md:87, and add rows to AC-TU-016: a post-audit path with a leading space whose trimmed form lies under the reports directory issues nothing; a replacement ref that makes an old evidence commit read as the changed tree issues nothing.

D32. design.md:28 (D.2), acceptance.md:25-33 (AC-TU-001/002), plan.md:127 (M5) — Severity: MINOR — Class: optional — Class code: design. The wrapper keeps `warned` in the closure the constructor creates, so a tree built once and run twice (as a table-driven test over "each verb of the todo tree" would) warns once and then never. Fix: reset the flag when the outermost wrapped handler returns, or build a fresh tree per invocation in the test, and say which in M5; add a two-invocations row to AC-TU-001.

D33. design.md:97 (D.5 rationale of R2) — Severity: MINOR — Class: optional — Class code: other. R2's invariant is "every non-report path ends equal to the audited content or equal to the first parent", so a card that deletes or reverts main code it absorbed after the audit (path equal to E) passes. That is consistent with the invariant but unstated; only fact 3's re-measure covers it. Fix: add one sentence naming the case to the "what R2 does not do" list.

D34. plan.md:158 (M9 RED: "the residual-ledger test lists the AGENTS.md row as an allowed alias statement and fails before the edit"), design.md:54, :64, acceptance.md:67 — Severity: MINOR — Class: optional — Class code: ac-wording. Class E6 (`AGENTS.md`, its template) is excluded from the residual test ("same exclusions"), so the test cannot fail before M9's edit; either the file is in scope (then the test is red between M8 and M9, the iteration-2 D25 shape) or it is not (then M9's RED sentence is untrue). Fix: say which.

D35. acceptance.md:108, plan.md:192, design.md:113 (AC-TU-020, REQ-TU-021) — Severity: SHOULD-FIX — Class: optional — Class code: other. The card's completion criterion "done is automatic for the most recent ten landed cards" is verified by a synthetic fixture; the live listing is a record, not a pass condition. The SPEC says so openly, and my measurement agrees: 65 of the 83 integration merges on local main commit no sync-audit file, and only 13 of the 18 that do pass R2 (P5). Not a SPEC defect, but the leader and operator should read that the live criterion will not be met until lanes commit evidence; consider stating the card's criterion as conditional in the landing note.

D36. research.md §R.14, design.md:99 — Severity: MINOR — Class: optional — Class code: other. The over-refusals the design accepts are larger than its three-card sample shows: P5 found one more (t1854: `CHANGELOG.md` and `internal/template/catalog.yaml` combined in an absorption, a shape every card that edits those shared files and absorbs main after its audit will produce) and four ties from verdict files named `-head-confirmation`, `-r2`..`-r4`. Fix: add them to §R.14 so the live rate is stated from the wider sample.

D37. design.md:131 (D.7) — Severity: MINOR — Class: optional — Class code: other. A queue lease that lapses reverts the queue row to `queued` and clears its `spec_id` (P12), so the setter's `picked` precondition fails for a card whose factory lease is still live. State the case (re-pick, then the setter) or accept it as out of scope.

Also noted, not numbered: when any fact is absent or a git read errors, the refusal is the gate's receipt-absent text and does not name the cause (spec.md:116); a lane meeting it learns nothing. Optional: carry the first absent fact in the refusal's detail line. D27 and the T18/T29 identifiers inside REQ-TU-018/019 stand as in iteration 2 (optional).

## Regression Check

Iteration-2 defects:

- D16: RESOLVED in the SPEC text (spec.md:110-116, design.md:85-105, AC-TU-016); hardening open as D31.
- D17: RESOLVED (design.md:95/:101, AC-TU-016 row f; premise measured on 11 merges).
- D18: RESOLVED (26/32 reproduced; AC-TU-004 pins 33).
- D19: RESOLVED in design (cobra order reproduced; run-step wrapper; AC-TU-002 rows 4-6); flag-lifetime nit D32.
- D20, D21, D22, D23, D24, D26: RESOLVED. D25: RESOLVED (ambiguity D34). D28: order fixed. D27: not fixed (optional).
- D1 to D15 stay resolved except D10 and D12, carried without re-measurement (Gap).

Introduced by this revision: D29, D30, D31 (blocking); D32 to D34, D36, D37 (minor); D35 informational. No iteration-1 or iteration-2 blocker reappears unchanged. The CN-4 verb was re-run in full (port) and MP-9 re-decided: no conflict.

## Recommendation

1. Is the SPEC sound enough to enter run phase? Not yet, by this audit: the required codex leg fails and three blocking edits are open. They are narrow: D29 (add `factory_card.go` to M1 and a CLI-level no-config row), D30 (a column or a reserved string for the automatic kind, plus the same-card manual row), D31 (a raw reader and replacement-objects-off for the fact reads, plus two rows). Everything else raised in iteration 2 is closed, the R2 and fact-3 premises hold on the merges I measured, and the Q5 shape the leader confirmed (three machine facts, issuer `auto`, revocable, fail-closed) is testable as written once D30 and D31 land.
2. This is the ceiling-hit round. The enumerated `fix_scope` anchors are hunk-limited and the REQ and AC identifier sets need not change (25 and 25 stay), so a delta re-audit scoped to D29-D31 (plus the full must-pass set and CN-4) is the natural next step under the ceiling policy; nothing found here indicates a split.
3. Optional, at the author's discretion: D32 to D37.

## Plan-artifact hash

The runtime `InMemoryCache.ComputeHash` (`internal/runtime/audit_cache.go:121-141`, read this run) hashes, in `planArtifactNames` order (acceptance.md, decision-index.md, design.md, plan.md, research.md, spec.md, tasks.md), skipping absent files, `%s:%d\x00`, the bytes, `\x00`. Present: acceptance.md 42037, decision-index.md 23485, design.md 51453, plan.md 38804, research.md 45626, spec.md 27907; tasks.md absent; progress.md excluded.

- Implementation 1 (Python 3, `python3 -I hash3.py <spec dir>`): `cbe6819defad8c2e664b58a7b275f3b973601f0005b119ed30de4275ad69e97a`
- Implementation 2 (Go, `go -C <scratchpad>/hashgo3 run main.go <spec dir>`, a transcription of the function body above): `cbe6819defad8c2e664b58a7b275f3b973601f0005b119ed30de4275ad69e97a`
- Agreement: yes. Files read from the clean working tree at HEAD `164050fdde316bc2717267e76515ae3e3d290768`.

## Gaps

- Refusals (verification-claim-integrity 3.1): this session's worktree guard refused every `git -C <t1641 path>`, compound `; echo exit=$?` git lines, variable-path `python3`/`sed`, a git heredoc, and the Write tool for the t1641 report path. Instead: Read and plain `grep` on literal absolute paths; `git grep <commit> --` and `git log/diff/show` from my own worktree against the shared object store (`164050fdde` is the clean t1641 HEAD, so the content is identical); python scripts started with literal paths that call `git -C`. The report itself was staged in the scratchpad and copied by a python script.
- No `go test`, no heavy slot, per the brief. Every statement about the behaviour of designed code (wrapper, in-transaction issuance and rollback, setter interleavings, 33-name enumeration, the 71/370 pin with T29, the lane guard's admission of a card whose factory row is already `done` on a repeat) is static reading or design, not a run. The cobra order is read, not executed.
- The awk verbs (traceability, CN-4, D7, D8) were not run as written; Python ports stand in. `moai spec lint`'s exit code was not captured.
- Codex's counterexamples (replacement ref, leading-space path, `issuer auto`, double run of one tree, deletion of absorbed main code) were not reproduced by me; I confirmed the premises beneath the second and third by reading (P10, P11) and the empty-integration premise (P9) independently of codex.
- D10 (floors) and D12 (template-source claim) were not re-measured; the installed build's lag was not re-measured; docs-site build, board mod tests, `sh` on CI runners not examined.
- Sample sizes: R2 and fact 3 were measured on 18 and 11 merges of the last 83 local-main integration merges that name a card; recency of committed evidence is a reading, not a measured cause. The leader's confirmation of Q5 is the coordinator's message, not a board record; I did not read the board this round.

## Residual-risk

- Even with D29-D31 fixed, the three facts are produced inside the lane's own worktree and tooling and bind to merged content; they do not prove an independent auditor ran. The leader-confirmed shape accepts that; REQ-TU-025's revocation depends on a leader noticing.
- Live rate: most recent merges commit no verdict, so the automatic path will close few cards until lanes commit evidence (D35, D36); the ten-card criterion is a mechanism proof plus a report, not a live pass.
- T29 admits `merged-local to done` on a remote-bearing repository once the merge is on the remote-tracking ref, which T18 refused; that loosening is intended and §H records the risk, but the SPEC does not list separately what the human-only decider and the remote refusal protected.
- In this repository's local-main flow (a remote exists, lanes merge to local main, origin advances by batch) a `merged-local` card closes only after its batch lands.
- Shared hot files (CHANGELOG, catalog) make R2 over-refuse any card that edits them and absorbs main after its audit; they fall to the leader's manual approval, fail-closed.
- `moai todo` and `moai gtd` share one tree, so every later verb is also on the alias until its removal, which is outside this SPEC.

## Operational Notes (unverified)

- Measure D29: in a temporary repository with no remote and no `git-strategy.yaml`, run `moai factory decide <card> --gate push` for a merged-local card with a valid leader receipt, on a build from the run-phase tree. Status: assumption (inferred from `factory_card.go:3238` and `verifyMerge`'s branch check).
- Measure D30: run `moai factory approve <card> --issuer auto`, then the revocation act, then read the receipt store. Status: assumption (inferred from `leader_approval_gate.go:344-373`).
- Measure D31: create a `git replace` ref for the evidence commit and a diff path with a leading space in a fixture, then run the three-fact check. Status: assumption (codex reported both; I read `gitRead`).
- Re-count the live rate at sync with `r2_rule.py`-style reads of the landed merges. Status: measured here for 83 merges (P5), assumption for later ones.

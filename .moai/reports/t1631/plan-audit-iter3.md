auditor-model: claude-haiku-5-5

# Plan audit iteration 3 (final delta): SPEC-CAPACITY-RECLAIM-001 (factory card t1631)

## Verdict

FAIL. Overall score 0.56 against the Tier M threshold of 0.80. Must-pass failed: 1 (MP-8). Blocking findings: 11 (D1 to D11 below). The required codex backend returned a reviewed `fail` with receipt rcpt-cd4fc3e8fc2f1dee80f2bc72. The plan convergence check reports `ok: true` with no unmet gate, so this is a reviewed defect, not an unmet gate.

This is the delta round authorized under leader decision d-20261010T080627Z-987a. Audited commit ab9f5a6d3ff523f908ebd1855e3584b9d58bc327 on branch WT-capacity-reclaim; the working tree was clean (`git status --short` empty). Tier M. The iteration 2 verdict (`.moai/reports/t1631/plan-audit-iter2.md`) was read first.

Reasoning context from the lane was ignored per M1 Context Isolation. The lane's repair-round statuses were treated as claims and checked against the artifacts; each disagreement is listed in the status table below.

In short: B2, B3, B4, B5 and B6 are resolved in text. B8 is resolved as a design question (operator verdict 8a62 covers A11), but its implementation path is not. B1 and B7 remain open. The 8a62 reflection is incomplete for Q1, Q2, Q4 and A11: the decision text is carried into REQ-CAP-004, REQ-CAP-006, REQ-CAP-009 and REQ-CAP-011, but the plan, the residual-risk section and the verification layer still carry superseded or unverified behaviour.

## Evidence-bearing report

### Claim

- C1. The verdict is FAIL: 11 blocking findings and one must-pass failure (MP-8), with overall score 0.56.
- C2. B2, B3, B4, B5 and B6 are resolved in text. B1 and B7 are not resolved. B8 is reopened on its implementation path.
- C3. MP-8 fails for AC-CAP-009 (approved-branch RB cell whose cited command prints nothing) and for AC-CAP-014 (cited command does not run in the audit runner).
- C4. Disposal narrowing that verdict 8a62 rejects remains in plan M3, plan §D.4, plan §E, spec §F, spec A15 and AC-CAP-015 (D5).
- C5. AC-CAP-004's cited RED command reproduces and its premise is sound for the current template. This clears AC-CAP-004 under B1; codex finding P1 on this criterion is not adopted; the reason is given in the B1 row of the status table.
- C6. The approved SessionStart split (Q1), the report-only categories at SessionStart (REQ-CAP-008), the SessionStart and doctor warning surfaces (REQ-CAP-009) and the Q4 archive-before-delete guarantee (REQ-CAP-004) have no verifying criterion (D8, D9).
- C7. The SPEC's measurement that the compiled profile cap has one reader is false: `internal/cli/doctor_disk.go` reads it at lines 155 and 261 (D10).
- C8. Build 4f8aba061 is not a strict ancestor of HEAD ab9f5a6d3 (verified). Attribution is stated in Baseline-attribution.

### Evidence

Commands and their observed outputs are listed in "Commands run" below. The principal evidence is:

- E1. `moai verify audit-plan` for this worktree: `config_status: ok`, `cross_model_active: true`, `enforced_required: ["codex"]`.
- E2. `mcp__moai__audit_multi` (project_root = this worktree, target = SPEC directory, no `gates`): `overall_verdict: fail`; codex `required` `fail`; five findings; `audit_receipt: rcpt-cd4fc3e8fc2f1dee80f2bc72`. The raw return is quoted in "Raw audit_multi return". Its findings were checked independently against the artifacts; four are confirmed (D1, D5, D6, D7), and the AC-CAP-004 finding is not adopted (C5).
- E3. Re-execution of the cited RB and MS cells at ab9f5a6d3. EV-004, EV-011, EV-012, EV-013 and EV-005 reproduce their stated stdout. EV-014 does not run. The §D.0 command for AC-CAP-009 prints nothing; its unescaped form prints the line that the cell claims.
- E4. Template ignore rules: the template `.gitignore` contains no rule that can match `.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl`, or any `*.corrupt*` or archive path (see Commands run). The premise of AC-CAP-004 is therefore correct.
- E5. Code reads: `internal/hygiene/targets.go` lines 29 to 36 list eight classes (context-usage, goal, codex-stop-chain, codex-stop-cap, agent-stops, routing-pending, verify, todo); none is worktree, database, or run record. `internal/cli/doctor_disk.go` lines 155 and 261 read `config.DefaultProfileMaxBytes`.
- E6. Build provenance: `git merge-base --is-ancestor 4f8aba061 HEAD` exits 1; `git merge-base 4f8aba061 HEAD` prints 2aab5f797; `git rev-list --count HEAD..4f8aba061` is 1; `git rev-list --count 4f8aba061..HEAD` is 5.

### Baseline-attribution

- Tree measured: HEAD `ab9f5a6d3ff523f908ebd1855e3584b9d58bc327` (short `ab9f5a6d3`), branch `WT-capacity-reclaim`, clean. Measured in this run.
- Document pins: the EV ledger pins `5d36bf91b15c9869118298613df77209f8a3893b`. `git diff --name-only 5d36bf91b15c9869118298613df77209f8a3893b HEAD` lists only the five SPEC files (acceptance.md, decision-index.md, plan.md, progress.md, spec.md). The code paths of the cells are therefore unchanged between the pin and HEAD. Measured in this run.
- Artifact hash (lane-supplied): `e2282227ae0eebc0482768766f7e3d4172c687ee2e4747539c445d5a6d2cc07a`. Supplied by the lane from `mcp__moai__audit_cache` tool output on this tree and commit. The auditor did not replay the hash (the tool is not callable from this session) and does not independently reproduce it.
- Judging build: `moai version` prints `v3.2.0-rc.29   moai_cp/20261007_195927-158-g4f8aba061   built 2026-10-09T16:02:30Z`; the self-hosted MCP server reports build commit 4f8aba061.
- Provenance status per verification-claim-integrity §2.2: the build is not built from this tree (the first branch of §2.2 was not performed this round). The second branch is satisfied literally: the installed build commit 4f8aba061 is not a strict ancestor of HEAD ab9f5a6d3 (`git merge-base --is-ancestor` exits 1). The build is divergent, not a descendant: merge base 2aab5f797, one commit the tree lacks, five commits the build lacks. The build's one commit (`fix(web): unlock effort for Haiku 5.5 rows and de-duplicate agent rows`) touches `internal/cli/profile_setup_translations.go`, `internal/config/profile.go`, `internal/template/model_policy.go` and its test, and `internal/web/*`. None of those paths is measured by a RED cell of this SPEC.
- Attribution of the cross-model result: the codex result and the audit-plan configuration read are attributed to build 4f8aba061 beside tree HEAD ab9f5a6d3. The file findings rest on reads in this worktree at ab9f5a6d3, not on the build's output.
- Codex admission: codex returned a verdict (`fail`), with a receipt. Codex therefore admitted; no `required_backend_fail` line is written.

### Gaps

- G1. `mcp__moai__audit_cache` was not called; the plan_artifact_hash is lane-supplied (see Baseline-attribution).
- G2. The awk verbs (Group A traceability, Group B CN-4, Group C D7 discovery, Group D D8) were not run, because the auditor constraints for this round prohibit awk programs. Traceability, ordering and D7 were judged by hand from grep output (see "Traceability" and "Ordering judgement"). Their outcome is therefore a manual judgement and is labelled as such, not a mechanical pass.
- G3. AC-CAP-014 (EV-014) was not re-executed in a modified form. Its recorded command does not run in the audit runner, so under the MP-8 execution discipline it is not executed further and is not recorded as a pass (D2).
- G4. Exit status for the grep cells (EV-004, EV-011, EV-012, EV-013, EV-014) is not displayed by the audit runner for zero-match or match output. The recorded values are the grep convention consistent with the observed output; they are not independently observed (D4).
- G5. The go test for AC-CAP-004 was not re-run by the auditor. Codex reports `ok github.com/modu-ai/moai-adk/internal/template 2.681s` for the existing test. That observation is codex's, not the auditor's.
- G6. The AC-CAP-002 pipeline (heavy traversal) was not executed; it is N/A under the Q2 rejection in any case.
- G7. The first build-provenance branch (build from this tree, invoke by path) was not performed this round (see Baseline-attribution).
- G8. Pre-flight measurements (plan §C) were not run; this is a plan-phase audit.
- G9. Only the codex backend was active (claude and glm gates are `off` in the plan). No second opinion on the rubric scores.
- G10. The machine-state digest `.moai/state/audit-plan-result.json` was written, as the cross-model skill requires before `moai verify audit-plan --result-file`. It is runtime state, not SPEC content, and is disclosed here because the round's constraint named only the report file.

### Residual-risk

- R1. The judging build divergence (Baseline-attribution) means the codex and plan-check results are attributed to build 4f8aba061, not to a build of this tree. The file findings do not depend on the build.
- R2. The plan_artifact_hash is lane-supplied and not reproduced by the auditor.
- R3. Exit-status claims for the grep cells rest on grep conventions, not on runner output (G4). If exit status is required as a runner observation, the RB cells cannot satisfy it in this runner as written (D4).
- R4. Single backend (codex). Convergence is one reviewed verdict, not a cross-model agreement.
- R5. If the leader's ceiling policy adjudicates this final hit, every blocking finding carries a scoped fix anchor in `fix_scope`. Whether the CLI admission ladder then splits or holds is outside this report and is not asserted here.
- R6. The 8 GiB shipped default (REQ-CAP-009, A11) is not attributed to verdict text beyond the warning-only design (advisory A16). An operator who approved the warning-only design may not have approved the value.

## Must-pass results

| MP | Result | Evidence and layer |
|---|---|---|
| MP-1 REQ number consistency | PASS | 14 REQ definitions, REQ-CAP-001 to REQ-CAP-014, sequential, spec.md lines 76 to 98. Requirement layer. |
| MP-2 GEARS format compliance | PASS (with notes) | Requirement layer (spec.md). Each entry opens with a GEARS keyword. Notes: REQ-CAP-004 line 82 carries a second, declarative sentence; REQ-CAP-006 line 84 carries a second `Where` sentence and a declarative sentence; REQ-CAP-009 line 90 carries "no cap value shall be compiled into the code" (legacy negative form). Recorded as advisory A17, as in iteration 2. |
| MP-3 YAML frontmatter validity | PASS | spec.md lines 2 to 15: 12 canonical fields present; `version: "0.1.2"` quoted (line 4); `created`/`updated` as ISO dates (lines 6, 7); `tags` a string (line 13); no snake_case alias; `phase: "v3.2.0 target"` is a release label (line 10). Advisory: `related_specs` (line 15) is outside the optional-field list. |
| MP-4 Language neutrality | PASS (N/A-equivalent) | Go-scoped SPEC; no multi-language tool enumeration. |
| MP-5 D7 cross-SPEC reconciliation | PASS | Referenced SPEC-MOAI-HYGIENE-001, SPEC-WORKTREE-SWEEP-001 and SPEC-WORKTREE-REAPER-001 each read `status: completed` (line 5 of each). No retired, superseded or archived reference, so no BLOCKING. |
| MP-6 D8 cross-platform discipline | PASS | `grep -rn -i syscall .moai/specs/SPEC-CAPACITY-RECLAIM-001` prints nothing. |
| MP-7 clarification gate | PASS | `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CAPACITY-RECLAIM-001` prints nothing; no research.md (Tier M); plan.md carries no marker. |
| MP-8 RED-now cell re-execution | FAIL | D1: the AC-CAP-009 cell's cited command prints nothing and the cell has no exit element; the criterion is RB on the approved Q1 branch. D2: the AC-CAP-014 cited command does not run in the audit runner. |
| MP-9 cross-artifact ordering | PASS (manual judgement) | The mechanical CN-4 verb was not run (G2). The plan has no `Exit:` bindings (plan line 87 is the only occurrence of "Exit"), so the verb would bind no criterion. Candidates were judged by hand against the plan's milestone order (§A: M1, M0, M2, M3, M4, M5, M6). No criterion binds a milestone on the forbidden side. Advisory A4 and A5. |

## Category scores (rubric-anchored)

| Dimension | Score | Band | Evidence |
|---|---|---|---|
| Clarity | 0.50 | multiple requirements require interpretation | The plan contradicts REQ-CAP-011 (plan M3 lines 40 to 41; §D.4 line 112; A15 line 153), so an implementer can follow either. §D.1 (plan lines 93 to 97) contradicts M1 line 11. The G-EXIT paragraph (acceptance line 15) prescribes a wrapper the rules do not admit. |
| Completeness | 0.75 | all sections present; frontmatter complete; Out of Scope has H3 headings with bullets | Content gaps: last activity missing from REQ-CAP-010 (line 94) and plan M3 (line 39); the doctor path missing from plan M5 (lines 61 to 62); §A HISTORY has no v0.1.3 entry (lines 20 to 24). |
| Testability | 0.50 | several criteria do not yield a binary observed RED as written | AC-CAP-009 cell (acceptance line 101) prints nothing as written; AC-CAP-014 cell (line 250) does not run; the exit element is recorded as runner-captured for four cells the runner does not report (D4); SessionStart behaviour has no test (D9). |
| Traceability | 0.50 | 14 of 14 REQs and 16 of 16 ACs map; four REQ clauses have indirect or no verifying criterion | REQ-CAP-004 ordering clause (D8); REQ-CAP-006 and -008 SessionStart clauses (D9); REQ-CAP-009 template and SessionStart clauses (D9, D10); REQ-CAP-010 names no last-activity field though AC-CAP-012 checks it (D6). |

Overall score = mean of the four dimensions = (0.50 + 0.75 + 0.50 + 0.50) / 4 = 0.5625, reported as 0.56. Tier M threshold: 0.80. Not met.

## Defects found (blocking; counted in blocking_count = 11)

D1. AC-CAP-009 approved-branch RB cell — acceptance.md:L101, :L199 to :L204 — Operator verdict 8a62 approves Q1 (split default). Acceptance line 85 says the approved branch (`apply`) "is the one to verify", so the criterion is RB on the approved branch. The §D.0 cell's cited command, `grep -n -E 'mode: (report\|apply)' internal/template/templates/.moai/config/sections/workflow.yaml`, prints nothing at ab9f5a6d3, because the escaped pipe is a literal `|` in ERE. The cell carries no exit element and no EV entry. The §D.3 RED note (line 204) says the approved branch "is not observable until Q1 carries a verdict", which is stale. The unescaped form `grep -n -E 'mode: (report|apply)' …` prints `318:        mode: report              # report | apply (auto path only)`, so the approved-branch RED is observable now. Severity: critical — Class: blocking — Required fix: write a four-element cell for the approved branch in the EV ledger (command unescaped, verbatim output, runner exit status, tree SHA), classify it RB, and correct the RED note. Defect class: ac-wording.

D2. AC-CAP-014 RB cell does not run — acceptance.md:L54 to :L58 (EV-014), :L106 (§D.0), :L250 — The recorded command `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` returns `(eval):1: no matches found: --include=*.go` in the audit runner. The same runner failure occurred in iteration 2 (G3). Under the MP-8 execution discipline the command is not executed further and the criterion is not recorded as a pass. The RB classification has no runnable RED-now cell. Severity: critical — Class: blocking — Required fix: replace the cell with a command that runs in the runner and re-observe it on the tree, or demote AC-CAP-014 under verification-completeness §2.1 and say so. Defect class: ac-wording.

D3. AC-CAP-005 reclassified RB→MS on a premise the runner contradicts — acceptance.md:L97 (§D.0 says RB), :L148 (heading says MS), :L83 and :L13 (§0 says MS) — The three locations disagree. The MS basis is "a plain git read whose exit status the runner cannot capture (G-EXIT)" (line 13). The audit runner did surface a non-zero status for a plain git command in this session (`git merge-base --is-ancestor 4f8aba061 HEAD` reported "Exit code 1"). The re-executed command (`git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'`) reproduces its five paths. Under verification-completeness §2.1, a RED that can be re-executed does not take the undecidable disposition; the disposition for an unreproducible RED is RG, not MS. MS still gates the sync-phase close (line 13), so the reclassification keeps release gating without the four-element cell. Severity: major — Class: blocking — Required fix: restore AC-CAP-005 to RB with a four-element cell, or justify MS from a measured runner limit and reconcile §0, the heading and §D.0. Defect class: design.

D4. Exit element unsubstantiated; G-EXIT wrapper — acceptance.md:L15 (G-EXIT), :L20 to :L58 (EV-004, -011, -012, -013, -014 carry "exit: … (runner-captured)"), plan.md:L87 — The document labels the exit of EV-004, EV-012, EV-013 and EV-014 (empty stdout, recorded as 1) and EV-011 (matching stdout, recorded as 0) as "runner-captured". The plan itself states (plan line 87) that "the plan session could not capture the exit code of a command that prints nothing". In this session the runner displayed no status line for EV-004, EV-012 and EV-013 ("(Bash completed with no output)") and no status line for the EV-011 match. The G-EXIT paragraph prescribes `bash -c '<command>; echo EXIT=$?'`. That wrapper adds a `;` chain and a shell re-parse, so the cell command does not run unmodified; the auditor's constraints for this round prohibit `bash -c`; and G-EXIT's premise ("a plain git command surfaces only its stdout") is contradicted by the observed "Exit code 1". Severity: major — Class: blocking — Required fix: state the exit-capture rule in terms the runner can satisfy, record for each empty-output cell what the runner did and did not show, remove "runner-captured" wherever it is not observed, and withdraw the bash-c wrapper. Defect class: design.

D5. Disposal narrowing retained against verdict 8a62 (B7; Q2) — plan.md:L40 to :L41 (M3: unpushed and owner as PRESERVE predicates with causes `unpushed-commits`, `owner-unconfirmed`, `owner-undetermined`), plan.md:L112 (§D.4 predicate order), plan.md:L131 (§E risk row: "owner and unpushed predicates in the shared evaluation"), spec.md:L153 (A15: "a count above zero is an explicit PRESERVE predicate"), spec.md:L162 (§F mitigation), acceptance.md:L257 (AC-CAP-015 Then: "only the done candidate is removed"). Contradicted by REQ-CAP-011 (spec line 95: the marking "shall change neither the verdict of any worktree nor the set of worktrees that --yes removes"), §B.2 (line 68: owner, unpushed and last activity "change no verdict and no --yes removal"), A14 (line 152), the Q2 verdict (decision-index line 24: "this SPEC approves no disposal narrowing"), and AC-CAP-013 (acceptance line 239: W must stay DISPOSE). Following plan M3 would fail AC-CAP-013. Severity: major — Class: blocking — Required fix: delete PRESERVE semantics for unpushed and owner from plan M3, §D.4, §E and §F, rewrite A15 as an advisory cause, and make AC-CAP-015 consistent with REQ-CAP-011 (or remove the done-only narrowing from its approved branch). Defect class: design.

D6. Last activity missing from the requirement and the milestone (Q2) — spec.md:L94 (REQ-CAP-010 lists path, branch, owner label, owner state, remote-landing predicate result, unpushed-commit count), plan.md:L39 (M3 adds owner, owner_state and unpushed only), spec.md:L68 (§B.2 defines last activity), acceptance.md:L231 (AC-CAP-012 checks `last_activity` against `git log -1 --format=%cI`), decision-index.md:L24 (the Q2 verdict: the REQ-CAP-010 report "carries owner, branch, unpushed commit count, and last activity"). The verdict requires the field; the requirement, the plan and the AC disagree. §B.2 line 62 also says "use five terms", which overstates the REQ layer (advisory A1). Severity: major — Class: blocking — Required fix: add last activity to REQ-CAP-010 (report-only, with its source and its representation when the read fails) and to plan M3 (sweepVerdict, text and JSON). Defect class: traceability.

D7. Q4 default not carried through the technical approach — plan.md:L93 to :L97 (§D.1 specifies the usage-log rotation mechanics: declared directory, pass-lock order, serialization, displacement), plan.md:L134 to :L136 (§E rows: "The 30-day prune and the rotation interleave — Serialized on the prune-state lock (M1)"; "A displaced chunk discards events the prune has not archived"). Plan M1 (line 11) states that none of these steps is built under the Q4 default. Line 96 says the displaced chunk is discarded at the next rotation, so events the prune has not archived are lost, which contradicts REQ-CAP-004's archive-before-delete guarantee (spec line 82). Severity: major — Class: blocking — Required fix: delete §D.1's rotation mechanics and the §E rotation rows, or mark them "not built; alternate not approved". Keep the alternate only as an unapproved option in the decision index. Defect class: design.

D8. Q4 ordering guarantee has no verifying criterion on the default branch — spec.md:L82 (REQ-CAP-004 states the guarantee the verdict added), decision-index.md:L43 ("Ordering guarantee added by this verdict (REQ-CAP-004)"), acceptance.md:L178 to :L190 (AC-CAP-007 verifies rotation only on the approved branch; its default branch at line 189 checks only the sinks.go owner-exception grep). No criterion checks archive-before-rewrite under the default branch. Severity: major — Class: blocking — Required fix: add a default-branch check of the retention prune's archive-before-rewrite (a named test or fixture in the retention package), within the Tier M ceiling of 16 by merging or retiring a criterion if needed. Defect class: traceability.

D9. Approved SessionStart behaviour unverified (Q1, REQ-CAP-008, REQ-CAP-009) — spec.md:L84 (REQ-CAP-006: non-destructive categories apply at SessionStart; destructive categories and backups and releases stay report-only), spec.md:L89 (REQ-CAP-008: report-only at SessionStart), spec.md:L90 (REQ-CAP-009: warning from `moai doctor` or the SessionStart hook), acceptance.md:L130 to :L137 (AC-CAP-003, the verdict's "apply rows exist" criterion, runs the CLI `moai clean --audit-logs --apply`, which REQ-CAP-006 leaves unchanged), acceptance.md:L199 to :L204 (AC-CAP-009 checks only the template default value). The verdict's completion criterion is satisfied by the auto path's log rotation and cache apply (decision-index line 13), which no criterion exercises. No criterion checks that destructive categories stay report-only at SessionStart, that backups and releases stay report-only there, or the SessionStart and doctor warning surfaces. Severity: major — Class: blocking — Required fix: add a verifying criterion (within the Tier M ceiling) for the SessionStart auto path (apply rows from the auto path; report-only destructive, backup and release categories; warning surfaces for the profile cap), and point AC-CAP-003 at the auto path or add the auto-path check to AC-CAP-009. Defect class: traceability.

D10. Compiled profile cap is read by the doctor path, and the SPEC says it is not — spec.md:L57 ("The only reader of `DefaultProfileMaxBytes` is `clean_home.go:243`"), plan.md:L61 to :L63 (M5: warning from `moai doctor` and SessionStart; delete the constant "when the pre-flight grep shows no reader in internal"), acceptance.md:L223 (AC-CAP-011 green path: the grep prints nothing). Measured at ab9f5a6d3: `grep -rn -E 'DefaultProfileMaxBytes' internal` prints `internal/config/defaults.go:455`, `internal/cli/doctor_disk_test.go:183`, `internal/cli/clean_home.go:243`, `internal/cli/doctor_disk.go:155` (`if size > config.DefaultProfileMaxBytes`) and `internal/cli/doctor_disk.go:261` (`profiles above %s` with the compiled value). The measurement in spec line 57 is false at HEAD, and the SPEC's label "re-observed at 40cc1031c" cannot hold: the commits after that pin are documentation commits, and `git diff --name-only 5d36bf91b HEAD` lists only SPEC files. M5 does not name doctor_disk.go or its test, so M5's deletion condition cannot be met and the approved doctor warning (REQ-CAP-009, A11) has no implementation path. Severity: major — Class: blocking — Required fix: correct spec line 57; add the doctor comparisons at doctor_disk.go lines 155 and 261 and the fixture at doctor_disk_test.go line 183 to M5, and add them to AC-CAP-011's green path. Defect class: design.

D11. Residual-risk mitigation contradicts A11 (absent key) — spec.md:L166 (§F: "Users without `state.profile_max_bytes` lose profile-cap enforcement until they set it (A11). Mitigation: the notice names the key"), spec.md:L149 (A11: "An absent key produces no warning"), spec.md:L90 (REQ-CAP-009: "Where the key is absent, no profile-cap warning shall be emitted"). For the absent-key case, which is the one §F describes, no notice exists. The residual risk the operator accepts is therefore stated with a mitigation that does not operate. Severity: major — Class: blocking — Required fix: restate the mitigation for the absent-key case (no notice; new templates ship the key; existing profiles are not covered) and keep the notice claim only for present, invalid values. Defect class: other.

## Advisory findings (optional; not counted in blocking_count)

- A1. spec.md:L62 (§B.2): "REQ-CAP-010 to REQ-CAP-012 … use five terms". REQ-CAP-010 to -012 use four or fewer of the five terms (live session appears in none). Wording.
- A2. plan.md:L54 (M4): "carries an operator verdict" as the trigger. The same bullet excludes a rejection ("Under the default or a rejection, M4 does not run"). Reword to "approved verdict".
- A3. acceptance.md:L180 (AC-CAP-007 Given): "decision Q4 … carries an approved verdict". Under 8a62 the Q4 verdict approves the default. The default-branch text (line 189) resolves the branch; reword to "approved alternate".
- A4. acceptance.md:L166 and :L174 (AC-CAP-016): the SHAs of B0 and B1, and the output of commands run after B1 exists, cannot be recorded inside B1. Say the record is made in a follow-up commit within M2.
- A5. plan.md:L68 (M6): a pre-implementation duty ("before the first implementation commit") sits in the milestone the plan lists last. Pre-flight (plan line 87) is where the plan actually places the RED recording, so the pair can be followed. Relabel. This is the MP-9 candidate reviewed under "Ordering judgement".
- A6. acceptance.md:L145 and :L83: "The extended cases of AC-CAP-004 are written in M2, so their failing run is that pre-flight record." The RED run falls inside M2, before the template edit. Relabel. The extended cases' failing run is not yet observed; the premise cell is sound for the current template (C5).
- A7. acceptance.md:L85 and :L154 (AC-CAP-001 and -005 green path "holds only after a Q3 verdict") versus plan.md:L33 (the Q3 gate is on the push, not the local commit) and acceptance.md:L269 ("recorded locally"). Clarify that the local commit is allowed and the push is gated.
- A8. acceptance.md:L271 (§D.6): "Their verdict fields stay empty until the operator records them." Stale: Q1, Q2 and Q4 carry 8a62.
- A9. progress.md:L9 (§E.1): "Decisions Q1 to Q4 remain OPEN … The plan audit iteration 2 has not run." Stale.
- A10. decision-index.md:L3: "All four rows remain OPEN in revision v0.1.2." Stale for the current revision.
- A11. spec.md frontmatter `version: "0.1.2"` (line 4), while A11, A12 and A14 cite v0.1.3 (lines 149, 150, 152). The §A HISTORY (lines 20 to 24) has no v0.1.3 entry. Bookkeeping.
- A12. acceptance.md:L104 (§D.0 AC-CAP-012 row) carries `grep -n -i -E 'owner\|unpushed'` (escaped), while EV-012 (line 43) carries the unescaped form. The two carriers disagree; EV-012 is the correct one. The lane's statement that §D.0 "cites EV ids that §0 does not match" is accurate in substance: §D.0 cites no EV ids and carries its own command strings.
- A13. spec.md:L164 (§F owner-conflict row): "the requirement is gated by decision Q4." Under the default, REQ-CAP-004 does not reverse SPEC-MOAI-HYGIENE-001 §D. Stale wording.
- A14. decision-index Q3 and plan §C: the usage-log and lessons-inbox measurement recipe is operator-held; no action for the auditor. Noted for completeness.
- A15. Q1 (targets.go lines 29 to 36): no class is named worktree, database or run record (verified). The `todo` class (`ClassTodoRecord`) is a residue class only if pre-flight confirms it, as plan M1 (line 15) says. Confirm at pre-flight.
- A16. A11 and REQ-CAP-009 (lines 149, 90): the 8 GiB shipped default is not attributed by verdict text beyond the warning-only design. Confirm the value with the operator; the effect is warning-only.
- A17. MP-2 notes on multi-sentence requirement entries (REQ-CAP-004 line 82, REQ-CAP-006 line 84, REQ-CAP-009 line 90), as in iteration 2.
- A18. A17 (spec line 155) and plan §E (line 136) describe the Q4 alternate's displacement consequence. Keep them labelled "not built under the default".
- A19. The verdict's category names "worktrees, databases, run records" (REQ-CAP-006 line 84) do not match the registry's class names. Map them explicitly.
- A20. AC budget is at the Tier M ceiling of 16. The repairs for D8 and D9 may need a merge or a tier change; the author should decide which.
- A21. Build provenance: the judging build is a sibling commit that touches none of the measured paths, but it is not built from this tree. Before the run phase, build the tree and invoke the binary by path (verification-claim-integrity §2.2, first branch).
- A22. The lane's residual item on the G-EXIT paragraph is folded into D4.

## B1 to B8 and operator verdict 8a62: status

| Item | Iteration 2 finding | Lane claim | Auditor status | Basis |
|---|---|---|---|---|
| B1 (MP-8) | RB cells lack exit and verbatim stdout; AC-CAP-004 RED not reproduced | PARTIAL | NOT RESOLVED → D1, D2, D3, D4 | AC-CAP-004 cleared: its cited command (EV-004) reproduces (no output); the template contains no rule for the five state paths (lines checked: 116 to 303, plus archive, learning and gz searches), so the premise is sound. Codex P1 on this criterion is not adopted: MP-8 does not test whether a command measures its premise, and the premise holds here. AC-CAP-011, -012, -013 cells reproduce; their exit element is D4. |
| B2 (trigger wording) | plan.md:11, :21; REQ-CAP-004, -006 admit a rejection as approval | DONE | RESOLVED | REQ-CAP-006 (line 84) requires "the approved split verdict" and names "no verdict, the default, or a rejection" as the report-only branch; plan M1 line 13 says "approved verdict"; plan lines 11 and 21 carry no trigger wording. Residual wording: A2, A3. |
| B3 (AC-CAP-012 values) | key presence only | DONE | RESOLVED (text) | acceptance line 231 checks owner, owner_state, unpushed and last_activity values against fixtures and git. Depends on D6 for last_activity. |
| B4 (AC-CAP-014 mutant) | branch swap and size -1 pass | DONE | RESOLVED (text) | acceptance line 248 checks branch per path and integer size, and line 249 names the mutant. Its RB cell is D2. |
| B5 (10 GiB surface) | no confirmation surface | DONE | RESOLVED (surface) | decision-index line 23 carries the threshold, its source and "pending operator confirmation; not part of verdict 8a62". Confirmation is needed only if Q2 is later approved. |
| B6 (AC-CAP-016 RED) | GREEN from creation; RED unobservable | DONE | RESOLVED | EV-016a reproduces: `git merge-base 5d36bf91b… 2aab5f797…` prints 2aab5f797, not its first argument, so the predicate can fail. EV-016b reproduces. A4 is advisory. |
| B7 (disposal narrowing) | REQ-CAP-011 narrows `--yes` with no decision row | SPLIT | NOT RESOLVED → D5 | REQ-CAP-011 (line 95) is report-only, but plan M3 (lines 40 to 41), §D.4 (line 112), §E (line 131), §F (line 162), A15 (line 153) and AC-CAP-015 (line 257) still narrow `--yes`, and AC-CAP-013 (line 239) requires W to stay DISPOSE. |
| B8 (profile cap, product decision) | removal of the compiled cap is an unrecorded decision | DONE | Design question RESOLVED (A11 attributes the warning-only design to 8a62). Implementation path NOT RESOLVED → D10, D11 | doctor_disk.go lines 155 and 261 still read the compiled cap; spec line 57 says the opposite; §F line 166 misstates the absent-key case. |
| Q1 (8a62, split default) | — | DONE in text | Text RESOLVED. Verification NOT RESOLVED → D1, D9 | REQ-CAP-006 (line 84), plan M1 (lines 13 to 17), decision-index line 13 and §A reflect the split. targets.go verified (E5). AC-CAP-009 cell is D1; SessionStart coverage is D9. |
| Q2 (8a62, reject bulk; report contents) | — | SPLIT | NOT RESOLVED → D5, D6 | Rejection reflected: AC-CAP-002 and AC-CAP-015 are N/A (acceptance line 85); M4 does not run (plan line 54); size target surface (B5) present. Last activity missing (D6). Narrowing remains (D5). |
| Q4 (8a62, default stands; ordering guarantee) | — | DONE in text | NOT RESOLVED → D7, D8 | Default reflected in REQ-CAP-004 (line 82), plan M1 (line 11), §0 (line 87). §D.1 and §E still describe the unapproved rotation (D7). Ordering guarantee unverified (D8). A3 advisory. |
| A11 (8a62, profile cap warning) | — | DONE | NOT RESOLVED → D10, D11 | REQ-CAP-009 (line 90) and A11 (line 149) reflect the warning-only design and the absent-key rule. The doctor path (D10) and the residual-risk statement (D11) are not reflected. |
| Inbox (REQ-CAP-005) | — | DONE | RESOLVED | REQ-CAP-005 (line 83) unchanged; plan M1 (lines 19 to 21) keeps the inbox unrotated; untracking gated on Q3 (plan line 33). |

Lane claims that disagree with the artifacts: B7 SPLIT (narrowing is not outside the 8a62 reflection; it is D5); B8 DONE (implementation path open: D10, D11); Q1 DONE in text (verification open: D1, D9); Q4 DONE in text (D7, D8); A11 DONE (D10, D11). The lane's residual contradictions:

- §B.2 "five terms" and the last-activity bullet: A1 (wording) and D6 (substance).
- AC-CAP-012 checks last_activity but REQ-CAP-010 does not name it: D6.
- The Q2 verdict says REQ-CAP-010 carries last activity: D6.
- §0 G-EXIT prescribes a `bash -c` wrapper: D4.
- Plan M3, §D.4 and the §E risk row put unpushed and owner into PRESERVE predicates: D5.
- Plan M4 trigger "carries an operator verdict": A2.
- AC-CAP-007 Given "approved verdict": A3.
- AC-CAP-009 RED note says Q1 has no verdict: D1.
- AC-CAP-015 keeps the owner-gated narrowing: D5.
- §D.0 cites EV ids that §0 does not match: A12 (and D1 for the cell that has no EV entry).
- §D.6 says the verdicts are empty: A8.
- No AC verifies the SessionStart behaviour of REQ-CAP-006, -008 or -009; AC count at ceiling: D9 (counted; the lane classed it advisory, and the auditor does not agree because the approved behaviour is unverified).
- spec.md version 0.1.2 while A11, A12 and A14 cite v0.1.3: A11.
- progress §E.1 stale: A9.

## Traceability (manual; AC-4 and AC-5 equivalent)

COLLECTED: 14 REQ definitions (spec.md lines 76 to 98); 16 AC headings (acceptance.md lines 112 to 253). Manual mapping from acceptance.md and spec.md:

| REQ | Mapped AC(s) | Depth |
|---|---|---|
| REQ-CAP-001 | AC-CAP-004 | Premise sound (C5) |
| REQ-CAP-002, -003 | AC-CAP-001, -005, -006, -016 | Adequate |
| REQ-CAP-004 | AC-CAP-007 | Default-branch ordering guarantee unverified (D8) |
| REQ-CAP-005 | AC-CAP-008 | Adequate |
| REQ-CAP-006 | AC-CAP-009 | SessionStart clause unverified (D9) |
| REQ-CAP-007 | AC-CAP-003 | Exercises the CLI path, not the auto path (D9) |
| REQ-CAP-008 | AC-CAP-010 | SessionStart clause unverified (D9) |
| REQ-CAP-009 | AC-CAP-011 | Template default and surfaces unverified (D9, D10) |
| REQ-CAP-010 | AC-CAP-012 | Checks last_activity, which the REQ does not name (D6) |
| REQ-CAP-011 | AC-CAP-013 | Golden set adequate; contradicted by plan M3 (D5) |
| REQ-CAP-012 | AC-CAP-015 | Then-clause narrows (D5) |
| REQ-CAP-013 | AC-CAP-014 | Adequate; RB cell not runnable (D2) |
| REQ-CAP-014 | AC-CAP-002 | Conditional on Q2 (N/A) |

ORPHAN: none. UNCOVERED at the nominal level: none. Depth gaps are D6, D8, D9 and D10.

## Ordering judgement (CN-4 / MP-9; manual)

The mechanical verb was not run (G2). Milestone order read by hand from plan §A and the headings: M1, M0, M2, M3, M4, M5, M6. Ordering candidates (grep of "before", "after", "first", "prior to", "pre-change" across the SPEC directory) were reviewed. Each candidate was judged against the milestone order:

- plan.md:5 and :21 (M0 precedes M2) and plan.md:24 to :25 (B0 before B1): consistent.
- plan.md:26 (take B0 after M1 only if Q4 is approved): not applicable under the default.
- plan.md:30 to :31 (RED run before the template edit): consistent.
- plan.md:32 (commit B1 after B0): consistent.
- plan.md:33 (Q3 verdict before B1 is pushed): consistent with decision-index lines 30 to 32.
- plan.md:54 to :56 (M4 after M3 and the Q2 verdict; evidence row before removal): consistent.
- plan.md:63 (pre-flight confirms the home-tier seeding before M5): consistent.
- plan.md:68 (RED cells recorded before the first implementation commit): consistent with the pre-flight assignment at plan line 87 and the DoD gate at acceptance line 263; relabel only (A5).
- acceptance.md:145 and :83 (failing run at pre-flight for cases written in M2): relabel only (A6).
- acceptance.md:166 and :174 (record SHAs and output at the M2 commit): cannot hold for B1's own SHA; advisory (A4). Not a milestone-order conflict, because the record can sit in a follow-up commit within M2.
- decision-index.md:43 (ordering guarantee: deletion only after archive): consistent with the default, where the prune is unchanged.

Result: no criterion binds a milestone on the forbidden side. MP-9 passes on manual judgement, labelled as such.

## Commands run (with observed outputs)

- `date -u` → `Sat Oct 10 08:35:14 UTC 2026` (first); `Sat Oct 10 08:49:53 UTC 2026` (later).
- `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631`.
- `git rev-parse --short HEAD` → `ab9f5a6d3`; `git rev-parse HEAD` → `ab9f5a6d3ff523f908ebd1855e3584b9d58bc327`.
- `moai version` → `v3.2.0-rc.29   moai_cp/20261007_195927-158-g4f8aba061   built 2026-10-09T16:02:30Z`.
- `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631` → `config_status: ok`, `model: codex`, `cross_model_active: true`, `cross_model_required: true`, `enforced_required: ["codex"]`; backends claude off, codex required, glm off.
- `git merge-base --is-ancestor 4f8aba061 HEAD` → exit 1 (the runner displayed "Exit code 1").
- `git merge-base 4f8aba061 HEAD` → `2aab5f797b75983e132af451da68f69e3426557b`.
- `git rev-list --count HEAD..4f8aba061` → `1`; `git rev-list --count 4f8aba061..HEAD` → `5`.
- `git log --oneline -1 4f8aba061` → `4f8aba061 fix(web): unlock effort for Haiku 5.5 rows and de-duplicate agent rows`.
- `git show --name-only --format= 4f8aba061` → internal/cli/profile_setup_translations.go; internal/config/profile.go; internal/template/model_policy.go; internal/template/model_policy_test.go; internal/web/* (agent overrides, agentfm, assets, fieldsets, console test, testdata).
- `git status --short` → (empty).
- `git diff --name-only 5d36bf91b15c9869118298613df77209f8a3893b HEAD` → the five SPEC files only.
- `git ls-files -- .moai/harness/usage-log.jsonl .moai/lessons-inbox.jsonl '*.corrupt*'` → `.moai/harness/learning-history/archive/2026-07.jsonl.gz.corrupt`, `…/2026-08.jsonl.gz.corrupt`, `…/2026-09.jsonl.gz.corrupt`, `.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl` (EV-005 reproduces).
- `grep -n -E 'usage-log|lessons|corrupt' internal/template/templates/.gitignore` → (no output) (EV-004 reproduces; exit status not displayed).
- `grep -n -E 'moai|harness|jsonl|lessons|corrupt' internal/template/templates/.gitignore` → lines 116 to 303 (`.moai-backups/`, `.moai/**/credentials.yaml`, `.moai/**/*.credentials`, `.agents/skills/moai*` and exceptions, `.moai/status_line.sh`, `.moai/memory/`, `.moai/worktrees/`, `.moai/branches/`, `.moai/state/`, `.moai/cache/`, `.moai/logs/*`, `.moai/reports/*`, `.moai/evolution/telemetry/*.jsonl` (285), `.moai/observability/*.jsonl` (291), `.moai/project/graph/`); no rule names harness, lessons, a bare `*.jsonl` or `*.corrupt` pattern, or a path that matches the five state paths.
- `grep -n -E 'archive|learning|gz' internal/template/templates/.gitignore` → (no output).
- `grep -n -E 'func Test|check-ignore|no-index|TempDir' internal/template/gitignore_local_artifacts_test.go` → 29 (`func TestTemplateGitignoreIgnoresLocalArtifacts`), 46 (`repo := t.TempDir()`), 85 (`check-ignore -q --no-index`), 95 (`t.Fatalf`). No extended cases for the state paths yet.
- `grep -n -E 'DefaultProfileMaxBytes' internal/cli/clean_home.go` → `243: if profileBytes <= config.DefaultProfileMaxBytes {` (EV-011 reproduces).
- `grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go` → (no output) (EV-012 reproduces).
- `grep -n -i owner internal/cli/worktree/sweep.go` → (no output) (EV-013 reproduces).
- `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` (cell form, quoted glob) → `(eval):1: no matches found: --include=*.go` (EV-014 does not run). The unquoted form also failed with the same message.
- `git merge-base 5d36bf91b15c9869118298613df77209f8a3893b 2aab5f797b75983e132af451da68f69e3426557b` → `2aab5f797b75983e132af451da68f69e3426557b` (EV-016a reproduces; the first argument is not the result).
- `git merge-base 2aab5f797b75983e132af451da68f69e3426557b 5d36bf91b15c9869118298613df77209f8a3893b` → `2aab5f797b75983e132af451da68f69e3426557b` (EV-016b reproduces).
- `git show --name-only --format= 40cc1031c2b18781377f1bdf2c590cd39f08244e` → the five SPEC files (EV-016c reproduces).
- `grep -n -E 'mode: (report\|apply)' internal/template/templates/.moai/config/sections/workflow.yaml` (the AC-CAP-009 cell command exactly as cited in §D.0) → (no output).
- `grep -n -E 'mode: (report|apply)' internal/template/templates/.moai/config/sections/workflow.yaml` (unescaped) → `318:        mode: report              # report | apply (auto path only)`.
- `grep -n -E 'usage-log' internal/hygiene/sinks.go` → `91: "usage-log.jsonl":     "internal/harness/retention.go (30-day prune to monthly gz)",`.
- `grep -rn -E 'DefaultProfileMaxBytes' internal` → `internal/config/defaults.go:455`, `internal/cli/doctor_disk_test.go:183`, `internal/cli/clean_home.go:243`, `internal/cli/doctor_disk.go:155`, `internal/cli/doctor_disk.go:261`.
- `grep -rn -E 'TestSweepReportAdvisoryNoDisposalChange|TestHomeCleanKeepFloors|TestProfileCapWarnsWithoutDeleting|TestSweepBulkRunEvidenceFirst' internal` → (no output; the named tests are not yet written).
- `grep -rn -i owner internal/worktree internal/cli/worktree internal/core/git` → (no output) (A14's owner-field premise re-observed at HEAD).
- `grep -rn -i syscall .moai/specs/SPEC-CAPACITY-RECLAIM-001` → (no output).
- `grep -rn 'NEEDS CLARIFICATION' .moai/specs/SPEC-CAPACITY-RECLAIM-001` → (no output).
- `grep -n '^status:' .moai/specs/SPEC-MOAI-HYGIENE-001/spec.md` → `5:status: completed`; likewise `.moai/specs/SPEC-WORKTREE-SWEEP-001/spec.md` → `5:status: completed`; `.moai/specs/SPEC-WORKTREE-REAPER-001/spec.md` → `5:status: completed`.
- `grep -n -E '^- \*\*REQ-CAP-0' .moai/specs/SPEC-CAPACITY-RECLAIM-001/spec.md` → 14 lines, lines 76 to 98.
- `grep -n -E '^### AC-CAP-0' .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md` → 16 lines, lines 112 to 253.
- `grep -n -E 'Exit' .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md` → line 87 only.
- `grep -rn -i -E 'before|after|first|prior to|pre-change' .moai/specs/SPEC-CAPACITY-RECLAIM-001` → candidate set reviewed in "Ordering judgement".
- Audit digest: `.moai/state/audit-plan-result.json` written (overall_verdict fail; gate_unmet absent from the result, recorded as null; plan_source config; codex required fail). `moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631 --result-file /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631/.moai/state/audit-plan-result.json` → `convergence_check: {"ok": true, "unmet": []}`.

## Recommendation

FAIL. Revise the SPEC (manager-spec, via the orchestrator) to address D1 to D11. In order of dependency:

1. Remove the disposal narrowing from plan M3, §D.4, §E and §F, and from spec A15 and AC-CAP-015; keep unpushed and owner as advisory causes only, and keep AC-CAP-013's golden set (D5).
2. Add last activity to REQ-CAP-010 and to plan M3 (text and JSON), report-only (D6).
3. Rewrite the AC-CAP-009 RB cell as a four-element cell in the EV ledger with the unescaped command, its verbatim output, its runner exit status and the tree SHA; correct the RED note (D1).
4. Replace the AC-CAP-014 cell with a command that runs in the runner, re-observed on the tree, or demote the criterion under §2.1 (D2).
5. Restore AC-CAP-005 to RB with a four-element cell, or justify MS from a measured runner limit; make §0, the heading and §D.0 agree (D3).
6. State an exit-capture rule the runner can satisfy; remove "runner-captured" where it is not observed; withdraw the G-EXIT bash -c wrapper (D4).
7. Delete §D.1's rotation mechanics and the §E rotation rows, or mark them not built under the Q4 default (D7).
8. Add a default-branch criterion for the archive-before-delete guarantee in REQ-CAP-004 (D8).
9. Add verifying criteria for the SessionStart split, the report-only categories and the warning surfaces; point AC-CAP-003 at the auto path (D9).
10. Correct spec line 57; add doctor_disk.go lines 155 and 261 and doctor_disk_test.go line 183 to M5 and to AC-CAP-011's green path (D10).
11. Restate the §F absent-key mitigation (D11).
12. Work through the advisory list (A1 to A22) in the same revision.

This round is the final delta authorized by decision d-20261010T080627Z-987a; any further revision falls under harness.plan_audit_ceiling_policy and needs leader authorization. If a further revision is authorized, its delta must include each hunk named in fix_scope; the CN-4 ordering verb must run in full (its awk form was prohibited for this auditor and must be run by an auditor permitted to run it).

## Raw audit_multi return

> {"per_backend_verdicts":[{"backend":"codex","gate":"required","verdict":"fail","summary":"Verdict: fail\n- [P1] B7 未解决：M3 仍要求缩小处置集合，与 REQ-CAP-011 和 Q2 判定相冲突 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:40\n  M3:40–45 要求 unpushed、非 done owner 和不可读 store 返回 PRESERVE；§D.4:112、spec.md A15:153、AC-CAP-015:257 仍保留这些限制。AC-CAP-013:239 却要求无 owner 的 W 保持 DISPOSE，因此按计划实现会直接违反验收条件。\n  置信度：高。建议：统一改为报告字段和 advisory causes，删除所有未经批准的处置收窄指令；Q2 拒绝批量执行并不授权修改现有 --yes 行为。\n- [P1] Q1 已批准，但 AC-CAP-009 的 release-blocking RED 缺少四要素 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:204\n  此项没有记录退出码，也没有 EV 条目；§D.0:101 仍标为未批准分支的 guard。本次在 ab9f5a6d3 执行所列 grep，输出 `318:        mode: report              # report | apply (auto path only)`，退出码为 0，说明批准分支已有可观测 RED。\n  置信度：高。建议：记录命令、完整原始输出、退出码和树 SHA，并将表格状态更新为批准分支的 RB。\n- [P1] B1 仍未解决：AC-CAP-004 用 grep 无匹配代替实际 ignore 行为的 RED — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:145\n  文档明确承认未运行隔离的 check-ignore probe，却在 :83 宣布采用 EV-004。本次执行指定测试，结果为 `ok github.com/modu-ai/moai-adk/internal/template 2.681s`；现有测试尚无所列新增 cases。grep 退出 1 只能证明没有对应文本，不能证明 Git 未忽略目标路径。\n  置信度：高。建议：先观测扩展 cases 或等价隔离 probe 的实际失败，再采用 RB；同步修正 §D.0:96 的“未采用”与 :83 的“已采用”矛盾。\n- [P2] Q4 默认判定未贯穿技术方案，仍留下未批准的 usage-log 回转指令 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:95\n  M1:11 明确禁止新增目录、displacement 和锁顺序，§D.1:95–97 却仍描述这些实现，并允许未归档 chunk 在下一次回转时被丢弃；这也违背 REQ-CAP-004 的先归档后删除保证。\n  置信度：高。建议：删除当前分支不执行的旧方案，或明确标为未批准且不可执行的备选方案。\n- [P2] Q2 要求的 last_activity 未进入 REQ-CAP-010 和 M3 交付清单 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/spec.md:94\n  decision-index.md:24 和 AC-CAP-012:231 要求 last_activity，并定义了 `%cI` 比较；REQ-CAP-010 和 plan.md:39 却都遗漏该字段。按需求和里程碑实现会遗漏运营者要求的报告信息，并导致验收失败。\n  置信度：高。建议：把 last_activity、来源、读取失败时的表示方式，以及文本和 JSON 输出要求同步加入 REQ-CAP-010 与 M3。","findings":[{"severity":"P1","title":"B7 未解决：M3 仍要求缩小处置集合，与 REQ-CAP-011 和 Q2 判定相冲突 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:40","body":"B7 未解决：M3 仍要求缩小处置集合，与 REQ-CAP-011 和 Q2 判定相冲突 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:40\nM3:40–45 要求 unpushed、非 done owner 和不可读 store 返回 PRESERVE；§D.4:112、spec.md A15:153、AC-CAP-015:257 仍保留这些限制。AC-CAP-013:239 却要求无 owner 的 W 保持 DISPOSE，因此按计划实现会直接违反验收条件。\n置信度：高。建议：统一改为报告字段和 advisory causes，删除所有未经批准的处置收窄指令；Q2 拒绝批量执行并不授权修改现有 --yes 行为。","file":"","line":0,"confidence":0,"recommendation":""},{"severity":"P1","title":"Q1 已批准，但 AC-CAP-009 的 release-blocking RED 缺少四要素 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:204","body":"Q1 已批准，但 AC-CAP-009 的 release-blocking RED 缺少四要素 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:204\n此项没有记录退出码，也没有 EV 条目；§D.0:101 仍标为未批准分支的 guard。本次在 ab9f5a6d3 执行所列 grep，输出 `318:        mode: report              # report | apply (auto path only)`，退出码为 0，说明批准分支已有可观测 RED。\n置信度：高。建议：记录命令、完整原始输出、退出码和树 SHA，并将表格状态更新为批准分支的 RB。","file":"","line":0,"confidence":0,"recommendation":""},{"severity":"P1","title":"B1 仍未解决：AC-CAP-004 用 grep 无匹配代替实际 ignore 行为的 RED — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:145","body":"B1 仍未解决：AC-CAP-004 用 grep 无匹配代替实际 ignore 行为的 RED — .moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md:145\n文档明确承认未运行隔离的 check-ignore probe，却在 :83 宣布采用 EV-004。本次执行指定测试，结果为 `ok github.com/modu-ai/moai-adk/internal/template 2.681s`；现有测试尚无所列新增 cases。grep 退出 1 只能证明没有对应文本，不能证明 Git 未忽略目标路径。\n置信度：高。建议：先观测扩展 cases 或等价隔离 probe 的实际失败，再采用 RB；同步修正 §D.0:96 的“未采用”与 :83 的“已采用”矛盾。","file":"","line":0,"confidence":0,"recommendation":""},{"severity":"P2","title":"Q4 默认判定未贯穿技术方案，仍留下未批准的 usage-log 回转指令 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:95","body":"Q4 默认判定未贯穿技术方案，仍留下未批准的 usage-log 回转指令 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md:95\nM1:11 明确禁止新增目录、displacement 和锁顺序，§D.1:95–97 却仍描述这些实现，并允许未归档 chunk 在下一次回转时被丢弃；这也违背 REQ-CAP-004 的先归档后删除保证。\n置信度：高。建议：删除当前分支不执行的旧方案，或明确标为未批准且不可执行的备选方案。","file":"","line":0,"confidence":0,"recommendation":""},{"severity":"P2","title":"Q2 要求的 last_activity 未进入 REQ-CAP-010 和 M3 交付清单 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/spec.md:94","body":"Q2 要求的 last_activity 未进入 REQ-CAP-010 和 M3 交付清单 — .moai/specs/SPEC-CAPACITY-RECLAIM-001/spec.md:94\ndecision-index.md:24 和 AC-CAP-012:231 要求 last_activity，并定义了 `%cI` 比较；REQ-CAP-010 和 plan.md:39 却都遗漏该字段。按需求和里程碑实现会遗漏运营者要求的报告信息，并导致验收失败。\n置信度：高。建议：把 last_activity、来源、读取失败时的表示方式，以及文本和 JSON 输出要求同步加入 REQ-CAP-010 与 M3。","file":"","line":0,"confidence":0,"recommendation":""}],"next_steps":[]}],"overall_verdict":"fail","disagreement_flag":null,"participant_count":1,"residual_risk_note":"required-backend FAIL: codex","fail_open_backends":[],"audit_receipt":"rcpt-cd4fc3e8fc2f1dee80f2bc72","plan_source":"config","build_commit":"4f8aba061","tree_root":"/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1631"}

## Machine lines

verdict: FAIL
audited_sha: ab9f5a6d3ff523f908ebd1855e3584b9d58bc327
overall_score: 0.56
must_pass_failed: 1
blocking_count: 11
plan_artifact_hash: e2282227ae0eebc0482768766f7e3d4172c687ee2e4747539c445d5a6d2cc07a
scope: delta
fix_scope: spec.md#REQ-CAP-010; spec.md#A15; spec.md#REQ-CAP-004; spec.md#REQ-CAP-006; spec.md#REQ-CAP-008; spec.md#REQ-CAP-009; spec.md#§B; spec.md#A11; spec.md#§F; plan.md#M3; plan.md#M5; plan.md#§D.1; plan.md#§D.4; plan.md#§E; acceptance.md#§0; acceptance.md#AC-CAP-003; acceptance.md#AC-CAP-005; acceptance.md#AC-CAP-007; acceptance.md#AC-CAP-009; acceptance.md#AC-CAP-011; acceptance.md#AC-CAP-012; acceptance.md#AC-CAP-014; acceptance.md#AC-CAP-015
defect_class: D1=ac-wording; D2=ac-wording; D3=design; D4=design; D5=design; D6=traceability; D7=design; D8=traceability; D9=traceability; D10=design; D11=other
convergence_overall: fail
required_backend: codex fail

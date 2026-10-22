# SPEC-JEV-RESOLVED-MODEL-001 — Progress

Card: t1768 · Class C · Tier M · plan-phase artifacts authored 2026-10-15; repair round 1 applied the same day after plan-audit iteration 1 (FAIL, 0.65, audited commit `cfb5517962088b09936dfbb8a2ff5b80e63a8893`). Amends `SPEC-JEV-CORE-001` (v0.5.0) and `SPEC-JEV-OPTIN-MEASURE-001` (v0.2.0) in place.

## §E.1 Plan-phase Audit-Ready Signal

| Item | State |
|---|---|
| SPEC ID regex check | `PASS` (executed as a Bash command; verbatim output in plan.md §C) |
| SPEC ID collision | none (`ls -d .moai/specs/SPEC-JEV-RESOLVED-MODEL-001` → `No such file or directory` before the plan commit) |
| Tier | M — REQ 14 / ceiling 16; AC 16 / ceiling 16 |
| Artifact set | spec.md · plan.md · acceptance.md · progress.md |
| Requirements | 14 (REQ-JRM-001 … REQ-JRM-014) |
| Acceptance criteria | 16 (AC-JRM-001 … AC-JRM-016); AC-JRM-010 adopted (OQ-1 option A) |
| Clarification markers | none in the SPEC directory. OQ-1, OQ-2, and OQ-3 are decided (card t1768, board) and recorded in plan.md §G |
| Amendments applied | `SPEC-JEV-CORE-001` 0.4.0 → 0.5.0 (unchanged in the repair round) and `SPEC-JEV-OPTIN-MEASURE-001` 0.1.0 → 0.2.0 (the straddle clause of AC-JEVO-011 removed in the repair round). Each moves `completed → in-progress (amendment)`, with `amendment_of` set to itself and a `## Amendments` section |
| Predecessor | `SPEC-JEV-CORE-001`, `SPEC-JEV-OPTIN-MEASURE-001` (declared as `related_specs`, not `depends_on`; see plan.md D5) |
| Successor | card t1762 (dispatch order: t1687, t1768, t1762, t1767) |
| Vision revision | not triggered (`.moai/project/product.md` carries no statement of the pin or the alias) |
| Status transition | (none) → draft |
| plan_status | `pending-plan-audit`: repair round 1 is applied, and the plan-auditor re-audits the delta over the fix scope before the signal becomes audit-ready |

### Repair round 1 — status per audit item

| Audit item | Status | Where |
|---|---|---|
| D1 (open-question marker, change notice) | Decided: option A. Marker removed | plan §G; REQ-JRM-010; AC-JRM-010 |
| D2 (open-question marker, shadow terms) | Decided: option (a). Marker removed | plan §G; REQ-JRM-011; plan §D.4 |
| D3 (open-question marker, regression set) | Decided: option (b). Marker removed | plan §G; REQ-JRM-010; REQ-JRM-011 |
| D4 (AC-JRM-014 red-now cells) | Fixed. AC-JRM-014 is a guard with line-anchored GRN cells; EV-RED-13 and EV-RED-14 are kept as historical base-tree cells; AC-JRM-014 is out of the Definition of Done's red-to-green list | acceptance.md AC-JRM-014 and ledger |
| D5 (REQ-JRM-013 form) | Fixed. REQ-JRM-013 is one shall clause; the sync return is REQ-JRM-014, with AC-JRM-016 for the sync phase | spec.md §C.4 |
| D6 (ModelID allow-list) | Fixed. Explicit allow-list, and no attribution field receives the constant. Plan D1, the §D.3 Run step, and the M6 gate are aligned | acceptance.md AC-JRM-005; plan §B D1, §D.2, §E M6 |
| D7 (section title casing) | Fixed. `Jev Model Attribution` (Title Case, which matches the headings of the rule file) in REQ-JRM-011, AC-JRM-011, EV-RED-11, and plan §D.4 | spec.md; acceptance.md; plan §D.4 |
| D8 (AC-JRM-011 depth) | Fixed. Anchored content checks, with a positive control | acceptance.md AC-JRM-011 |
| D9 (non-discriminating fixture) | Fixed. `jev-9.8.7` in AC-JRM-002, 004, 007, and 009, and in every new fixture; a two-call case in AC-JRM-002 | acceptance.md; plan §D.6 and §D.7 |
| D10 (unadopted REQ-JRM-010) | Fixed. AC-JRM-010 is adopted with a RED-now cell and a green-path cell; the Definition of Done follows it | acceptance.md AC-JRM-010 and Definition of Done |
| D11 (live consumer omitted; selector; count) | Fixed. `internal/cli/todo_auto_rank_test.go` added (its fixture gains a versioned model, and it gains an unattributed case); `internal/cli/todo_auto_rank.go` gains the reason `jev-unattributed`; the §F selector is widened and anchored. Importer recount: 23 importers, of which 12 change; with the 5 non-importer files, 17 files change. The tier is not changed | plan §D.1, §D.6, §F |
| D12 (green-path cells absent) | Fixed. A green-path cell per release-blocking AC, naming the milestone and the `--- PASS:` line; the Definition of Done requires the line, not exit 0 alone | acceptance.md green-path cells and Definition of Done |
| D13 ("every usage record") | Fixed. The positive clause of REQ-JRM-005 is scoped to answered calls; "never the request alias" stays for every record | spec.md §C.1 |
| D14 (OPTIN straddle clause) | Fixed. The straddle clause is dropped from AC-JEVO-011. The OPTIN HISTORY row never named the straddle, so it needed no edit | SPEC-JEV-OPTIN-MEASURE-001 acceptance.md AC-JEVO-011 |
| D15 (§F misstates the open decisions) | Fixed. §F names the requirements each decision can change, and records the three decisions as settled | spec.md §F |
| D16 (AC-JRM-015 traces to no REQ) | Fixed (optional). Mapped to REQ-JRM-013 | acceptance.md AC-JRM-015 |
| D17 (unanchored amendment greps) | Fixed (optional). Line anchors. The prior_completed_sha key is a bullet of `## Amendments`, so its anchor is `^- prior_completed_sha: ` | acceptance.md AC-JRM-014 |
| D18 (sibling SPEC pin wording) | Not applied. It is outside the list for this round. Still open for the leader | — |
| D19 (lint warning not recorded) | Fixed (optional). The tree-built lint result, warning included, is recorded below | this section, lint record |
| D20 ("answered call"; REQ-JRM-001 scope) | Fixed (optional). "Answered call" is defined in the spec.md Definitions; REQ-JRM-001 is narrowed to the request's model | spec.md; REQ-JRM-001; AC-JRM-001 |

Decisions the author took in this round, for the leader to confirm: the reason `jev-unattributed` for an unattributed answer in the todo auto-rank surface (plan §G; the existing one-reason-per-condition invariant forbids sharing `jev-unreachable`), and the kickoff fallback `jev_call_failed` for an unattributed answer (plan §G; fail-closed).

### Tree-built lint record (AC-JRM-015, D19)

Tree: HEAD `cfb551796`, with the repair-round SPEC edits uncommitted. Binary: `go build -o <scratchpad>/moai-tree ./cmd/moai` (plain build, no ldflags), invoked by path. Command: `<scratchpad>/moai-tree spec lint .moai/specs/SPEC-JEV-RESOLVED-MODEL-001 .moai/specs/SPEC-JEV-CORE-001 .moai/specs/SPEC-JEV-OPTIN-MEASURE-001`. Verbatim output:

```text
SEVERITY  CODE                  FILE                                            LINE  MESSAGE
--------  ----                  ----                                            ----  -------
WARNING   StatusGitConsistency  .moai/specs/SPEC-JEV-OPTIN-MEASURE-001/spec.md  1     SPEC SPEC-JEV-OPTIN-MEASURE-001 frontmatter status 'in-progress' disagrees with git-implied status 'implemented'

0 error(s), 1 warning(s)
```

The warning is pre-existing: the plan audit's iteration-1 verdict reports the same row at `cfb551796`. It compares the OPTIN amendment's `in-progress` status with the git-implied status. Whether it clears at the sync phase is not measured in this round. AC-JRM-015's "no errors" holds: 0 errors.

## §E.2 Run-phase Evidence

Record format: Claim / Evidence / Baseline-attribution / Gaps / Residual-risk (`verification-claim-integrity.md` §3). The measurements below were taken by a read-only gates agent on tree `810ae985c5bd105b4245c22389904e05017cce9a` (M6). This section records only what that agent observed; nothing here is carried over from another tree.

### Claim

The run phase of SPEC-JEV-RESOLVED-MODEL-001 (card t1768) is complete at milestones M1 through M7: the build, format, vet, test, lint, audit, and AC sweeps listed under Evidence passed on the M6 tree, AC-JRM-001 through AC-JRM-015 are observed PASS, and AC-JRM-016 is a sync-phase criterion that was not observed.

### Evidence

Milestone commits (parent chain from `c985b6414`; full SHAs from `git -C <worktree> log --format=%H%x20%s -7`):

| Milestone | Commit | Scope |
|---|---|---|
| M1 | `da1a2e379e5507c78753697212dbf787419f04c8` | internal/jev data model and attribution |
| M2 | `2d86b88328ef60da587613970a3ec6e470717b48` | internal/jevmeasure measurement attribution |
| M3 | `09140cc18e697bad9eb497361a330ee63fd53579` | change notice (MCP notice test) |
| M4 | `d6cea20407a03dffba0fd871467ff6e3169abc5f` | model-policy.md Jev Model Attribution section and template mirror |
| M5 | `67e4c3d39518af59b8edd36647f47b66077d3053` | doctor, kickoff, auto-rank surfaces |
| M6 | `810ae985c5bd105b4245c22389904e05017cce9a` | comment sweep (pinned-model wording) |
| M7 | this commit | run-phase evidence and audit-ready signal (progress.md only) |

Gates, all exit 0 on tree `810ae985c`:

| Command | Observed result |
|---|---|
| `gofmt -l` on the plan §F file list | printed nothing |
| `go vet` of internal/jev, internal/jevmeasure, internal/contract/kickoff, internal/cli | printed nothing |
| `go test -count=1 -timeout=45m` internal/jev | `ok` 0.799s |
| `go test -count=1 -timeout=45m` internal/jevmeasure | `ok` 5.588s |
| `go test -count=1 -timeout=45m` internal/contract/kickoff | `ok` 254.525s |
| `go test -count=1 -timeout=45m` internal/contract/receipt | `ok` 9.078s |
| `go test -count=1 -timeout=45m` the plan §F anchored internal/cli selector (no session_worktree tests, per board f661) | `ok` 1344.263s |
| `TestRuleTemplateMirrorDrift` | `ok` 1.031s |
| `make build` | exit 0; `git status` unchanged afterwards |
| `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |

The same internal/cli selector hit the default 10-minute timeout once at M5 under machine load (~377); its green rests on the longer-timeout runs recorded above.

Lint and audit with a tree-built binary:

| Command | Observed result |
|---|---|
| `spec lint` on SPEC-JEV-RESOLVED-MODEL-001, SPEC-JEV-CORE-001, SPEC-JEV-OPTIN-MEASURE-001 | 0 errors, 1 warning (`StatusGitConsistency` on SPEC-JEV-OPTIN-MEASURE-001; pre-existing, an in-progress amendment, same row as the plan-phase record above) |
| `spec audit --filter-spec` for SPEC-JEV-CORE-001 and SPEC-JEV-OPTIN-MEASURE-001 | no drift finding except INFO `EraAutoDetected` |

AC sweep:

| AC | Status | Evidence |
|---|---|---|
| AC-JRM-001 through AC-JRM-015 | PASS | The evidence lines named in the acceptance.md green-path cells were observed |
| AC-JRM-005 | PASS | allow-list: 2 lines for `jev.ModelID` outside internal/jev (`measure.go:320`, `doctor_jev.go:114`); 4 lines inside `jev.go` (60, 72, 477, 595) |
| AC-JRM-011 | PASS | anchored greps print 1 on both copies; `does not ship` prints 0; `cmp` exit 0; mirror test PASS |
| AC-JRM-016 | not observed | sync-phase criterion; stays pending |

Clarification-marker check over the SPEC directory printed nothing (exit 1, as expected for a zero-hit grep).

### Baseline-attribution

Every row above was measured in this run, against tree `810ae985c5bd105b4245c22389904e05017cce9a`, by the commands named in its row. Lint and audit rows used a binary built from the tree and invoked by path; its VCS stamp is recorded under Gap 3.

### Gaps

1. The AC-JRM-011 positive control (a copy with the Version change bullet removed, expecting count 0) was refused by the worktree guard and was not run.
2. The swept test count of the plan §F internal/cli selector was not observed (the selector ran without `-v`); the AC-JRM-008 and AC-JRM-010 cli tests were run separately with `-v`.
3. The tree-built binary's VCS stamp (`478285ab4c`, `vcs.modified=true`) is a strict ancestor of HEAD, so the binary's stamp lags the tree (`go build` compiled the on-disk files).
4. The plan-audit iteration-3 verdict's Claude leg ran on `claude-haiku-5-5`, and the implementation agents inherited the lane session model (Sonnet 5.5 for the later milestones). This deviates from the lane rule; the leader accepted it (board d-20261015T092351Z-3721).
5. Coverage: internal/jev 96.0% and internal/jevmeasure 95.9% are subagent reports and were not re-measured.
6. Whole-package `./internal/cli` and session_worktree tests were not run (board f661, 14a2). The repository-wide test verdict belongs to CI on the integration branch and is PENDING at report time.
7. The T7 stage verb printed an error while applying the transition (command output and the recorded state disagree); the cause is unknown and was reported to the leader.

### Residual-risk

The live alias behaviour (`jev-latest` answered as `jev-1.13.0`) is the leader's observation of 2026-10-15; no endpoint call was made in this card. The green of the internal/cli selector depends on the 45-minute timeout runs, and its swept count was not observed (Gap 2).

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_status: audit-ready
run_complete_at: 2026-10-15
run_head_sha: 810ae985c5bd105b4245c22389904e05017cce9a
milestones_complete: M1-M7
ac_observed_pass: AC-JRM-001..AC-JRM-015
ac_pending_sync: AC-JRM-016
gaps: [1, 2, 3, 4, 5, 6, 7]   # numbered as in §E.2 Gaps
```

## §E.4 Sync-phase Audit-Ready Signal

Record format: Claim / Evidence / Baseline-attribution / Gaps / Residual-risk (`verification-claim-integrity.md` §3). Every observation below was made in the sync run, against the tree at HEAD `47fd09bb12c4801a17b52f84ac56af9d10b3e0ec` plus the uncommitted sync edits listed under Claim.

### Claim

The sync phase of SPEC-JEV-RESOLVED-MODEL-001 (card t1768) carries the single sync commit. That commit moves this SPEC's `spec.md` `status:` from `in-progress` to `completed` (the merged in-progress → implemented → completed transition) and returns SPEC-JEV-CORE-001 and SPEC-JEV-OPTIN-MEASURE-001 to `completed` by frontmatter-only changes (REQ-JRM-013, REQ-JRM-014). AC-JRM-016 is observed PASS. The CHANGELOG entry is added, and no `.moai/docs` file or README states the pinned model id or the alias as current behaviour.

### Evidence

| Check | Command | Observed result |
|---|---|---|
| Status of the three SPECs | `grep -n '^status: completed$'` on the three `spec.md` files | one line each: `SPEC-JEV-RESOLVED-MODEL-001/spec.md:5`, `SPEC-JEV-CORE-001/spec.md:5`, `SPEC-JEV-OPTIN-MEASURE-001/spec.md:5` |
| AC-JRM-016, CORE diff | `git diff --unified=0 HEAD -- .moai/specs/SPEC-JEV-CORE-001/spec.md` (and the same against plan commit `cfb551796`) | one hunk, `@@ -5 +5 @@`, inside the frontmatter: `-status: in-progress` / `+status: completed`. The `updated:` line was already `2026-10-15` and is unchanged |
| AC-JRM-016, OPTIN diff | the same two commands for `SPEC-JEV-OPTIN-MEASURE-001/spec.md` | one hunk, `@@ -5 +5 @@`, inside the frontmatter: `-status: in-progress` / `+status: completed`; `updated:` already `2026-10-15` |
| AC-JRM-014 guard | `grep -n '^amendment_of: \|^- prior_completed_sha: '` on both amended SPECs | CORE line 15 `amendment_of: SPEC-JEV-CORE-001`, line 34 `- prior_completed_sha: c8731f965`; OPTIN line 16 `amendment_of: SPEC-JEV-OPTIN-MEASURE-001`, line 32 `- prior_completed_sha: 52c460d30` |
| Lint (AC-JRM-015) | `<tree-built binary> spec lint` on the three SPEC directories | `No findings — all SPEC documents are valid` (the `StatusGitConsistency` warning on SPEC-JEV-OPTIN-MEASURE-001 recorded in §E.1 and §E.2 no longer appears) |
| Audit (AC-JRM-015) | `<tree-built binary> spec audit --base-dir <worktree> --filter-spec <ID>` for each of the three SPECs | one finding each, `[INFO] <ID> (V3R6) — EraAutoDetected`; no `SyncStatusDrift` and no other drift finding |
| CHANGELOG duplicate guard (B12) | `grep -c 'SPEC-JEV-RESOLVED-MODEL-001' CHANGELOG.md` | `0` before the edit, `1` after |
| AC count (B12) | the MOAI-AC-COUNTER awk over `acceptance.md` (Tier M) | `16` (`live=16 excluded=0 ambiguous=0`); requirement lines `^\*\*REQ-JRM-` count `14` |
| CHANGELOG file paths | `ls` of the six implementation paths and both `model-policy.md` copies | all exist |

### Documentation review (plan OQ-6)

| Surface | Command | Result |
|---|---|---|
| `.moai/docs/jev-local-operations.md`, `.moai/docs/kickoff-autonomy.md`, `.moai/docs/stale-card-premise-check.md` | `grep -rn -i 'jev-1.13.0\|jev-latest'` over the three files | no hit (exit 1); none states a pinned model id or the alias; none edited |
| README, docs-site, `.moai/docs` | `grep -rln -i 'jev-1.13.0' README.md docs-site .moai/docs` | no hit (exit 1); nothing to edit or report |

### Baseline-attribution

The status, diff, anchor, count, and grep rows were produced by the commands named in their rows, in this run. The lint and audit rows used a binary built from the tree by `go -C <worktree> build -o <scratchpad>/moai-tree-sync ./cmd/moai` and invoked by absolute path (`verification-claim-integrity.md` §2.2). Its VCS stamp is `vcs.revision=478285ab4c1fc893b020bb464d14cb69c8422523`, `vcs.modified=true`; `git merge-base --is-ancestor` of that revision against HEAD `47fd09bb1` exits 0, so the stamp is a strict ancestor and lags the tree. `go build` compiled the on-disk files, so the judging code is the tree's.

### Gaps

1. The tree-built binary's VCS stamp lags HEAD (see Baseline-attribution); the stamp was not corrected by a rebuild with ldflags.
2. `sync_commit_sha` was the placeholder `pending-backfill` in the sync commit itself (a commit cannot cite its own hash); the real SHA has been backfilled in a following commit.
3. The run-phase Gaps 1 to 7 in §E.2 were not re-observed in the sync phase and are carried by reference. In particular the repository-wide test verdict belongs to CI on the integration branch and was not observed here; no `go test` was run in the sync phase.
4. The MCP audit and sync-audit tools were not run; the sync-auditor verdict is owned by the orchestrator.
5. The CHANGELOG's description of the `notice` key on the MCP output rests on the SPEC's HISTORY row 0.1.2 (D21, `mcp_jev.go:114–116`); the sync phase did not re-run the MCP notice test.

### Residual-risk

The `completed` status of the two amended SPECs depends on the sync commit landing as one commit: a partial landing would leave an amended SPEC `completed` while its HISTORY row still reads as an amendment in flight. Until the backfill commit lands, `moai spec lint` may warn on the placeholder slot-format of `sync_commit_sha` in this SPEC's progress record, which the schema doctrine admits as a sanctioned intermediate state.

```yaml
sync_complete_at: 2026-10-15
sync_commit_sha: f009522d162ea5ce72905a54ccd22a1f2fbe47a4
sync_status: complete
b12_self_test_a: pass   # pre-emission grep: 0 before the edit
b12_self_test_b: pass   # AC count 16, requirement count 14
b12_self_test_c: pass   # all CHANGELOG paths exist
changelog_entry_position: "[Unreleased] > Added, first entry"
frontmatter_status_transitions:
  SPEC-JEV-RESOLVED-MODEL-001: in-progress -> completed
  SPEC-JEV-CORE-001: in-progress -> completed
  SPEC-JEV-OPTIN-MEASURE-001: in-progress -> completed
ac_observed_pass_sync: AC-JRM-016
```

## §G Override and Refusal Record

- 2026-10-15T09:33:28Z SPEC-JEV-RESOLVED-MODEL-001 ceiling-outcome outcome=pass-through reasons="plan-audit ceiling reached (round count 3 >= tier ceiling 2 + 1 delta rounds); the verdict is admission-clean and admits without a question (REQ-ACE-013)" evidence=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1768/.moai/reports/t1768/plan-audit-iter1.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1768/.moai/reports/t1768/plan-audit-iter2.md,/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1768/.moai/reports/t1768/plan-audit-iter3.md

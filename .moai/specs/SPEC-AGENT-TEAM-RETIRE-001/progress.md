# Progress — SPEC-AGENT-TEAM-RETIRE-001

> Canonical §E lifecycle skeleton. Plan-phase emits placeholder headings only;
> §E.2/§E.3 are populated by manager-develop (run-phase) and §E.4 by
> manager-docs (sync-phase).

## §E.1 Plan-phase Audit-Ready Signal

- SPEC ID regex self-check (executed Bash, verbatim output `PASS`):
  `decomposition: SPEC ✓ | AGENT ✓ | TEAM ✓ | RETIRE ✓ | 001 ✓ → PASS`
  (canonical `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- ID uniqueness: `.moai/specs/` grep — only `SPEC-V3R6-AGENT-TEAM-REBUILD-001`
  shares the token family (different ID); no collision.
- Frontmatter: 12 canonical fields present; `status: draft`; `priority: P1`;
  ISO `created`/`updated`; `tags` comma-separated string; `tier: L`;
  `era: V3R6` (explicit, avoids transient H-2 misclassification at plan-phase).
- Artifacts emitted (Tier L set + progress): spec.md, plan.md, acceptance.md,
  design.md, research.md, progress.md.
- Requirements: 22 REQ (GEARS — Ubiquitous / When / While / Where / shall-not
  all exercised); AC: 29 (100% AC→REQ coverage; 4 preservation STILL-EXISTS
  ACs; 4 GWT scenarios). v0.1.1: +REQ-ATR-022, +AC-ATR-027/028.
  v0.1.2: +AC-ATR-029 (D6 key-token sweep).
- plan-auditor iteration-1 (FAIL 0.86, blocking D1) corrections applied at
  v0.1.2 — all findings re-verified with live greps before editing:
  D1 AC-ATR-015 sweep re-rooted to explicit subtrees (bare `.claude/` root
  pulled in protected `worktrees/` + `agent-memory/` — mechanically
  unsatisfiable) + 4 extra dangling-ref files enumerated in M3;
  D2 fieldsets.templ 0-team-hit correction → schema_sections.go FieldDef rows
  (REQ-ATR-008 / AC-ATR-013 / research.md A.3 retraction);
  D3 `WorkflowTeamAutoSelection()` + types_test.go blocks added to M1
  (design.md §B blast-radius sentence corrected — "confined to tests" false);
  D4 REQ-ATR-011 expanded to workflows/** submodules, case-insensitive
  (research.md §E delta (4) retracted);
  D5 vacuous tokens replaced (`Mode 3 — RETIRED` baseline-0 tombstone;
  `Agent Teams 모드` baseline-1 run.md token);
  D6 AC-ATR-029 added; D7 15 funcs; D8 AC-ATR-020/021/025 grep hardening.
- Anchor verification: every removal/preservation target verified by executed
  command (research.md §A/§E); 6 deltas vs the task brief recorded in
  research.md §E (i18n key-family split, TeamAutoSelectionConfig extension,
  path corrections, sync.md non-reference, Phase 0 packages not-yet-existing).
- Clarifications: 0 open. Both plan-time [NEEDS CLARIFICATION] markers resolved
  by user decision (2026-07-11, orchestrator-relayed) at v0.1.1:
  (1) team/glm.md → migrate-essentials-then-delete (REQ-ATR-022 / AC-ATR-027 /
  design.md D9); (2) auto-select thresholds → prose-only SSOT, D8 adopted
  (REQ-ATR-010 extended / AC-ATR-028). Markers removed from plan.md §A.
- Status: plan-phase artifacts authored (status: draft). Commits (by the
  orchestrator): `8f7234a76` v0.1.0 (`docs(` prefix — StatusGitConsistency
  warning standing per the plan-commit feat( prefix lesson), `5ebed9241`
  v0.1.1 (`feat(` prefix). v0.1.2 audit-fix edits uncommitted at handback
  (agent instructed not to commit); `moai spec lint` result recorded in the
  plan-phase completion report.

## §E.2 Run-phase Evidence

### M0 — Phase 0 de-risk migration (REQ-ATR-001/002/003)

Commit: `feat(SPEC-AGENT-TEAM-RETIRE-001): M0 — migrate lockfile + taskledger` (SHA recorded at M1 entry below).

Files: NEW `internal/lockfile/{lockfile_unix.go,lockfile_windows.go,lockfile_unix_test.go}` (build-tag pair `//go:build !windows` / `//go:build windows` preserved; Windows cross-process limitation comment preserved verbatim); NEW `internal/cli/taskledger/{taskledger.go,taskledger_test.go}` (TeamTaskEntry / AppendTask / ClaimTask / TaskClaimer migrated; TestClaimTask + TestClaimTaskConcurrent migrated + 4 added error-path/round-trip tests); `internal/cli/team_spawn.go` thinned to delegating aliases; `internal/cli/{settings.go,glm_tools.go}` re-pointed `lockFile`/`unlockFile` → `lockfile.Lock`/`lockfile.Unlock` (2 additional non-test lock callers discovered at run-phase — not in the plan-time inventory, which only enumerated team_spawn symbols); `internal/cli/clifix_critical_repro_test.go` re-pointed to `taskledger.ClaimTask`; 3 lock files deleted from `internal/cli`.

MX tags: `@MX:ANCHOR` + `@MX:REASON` on `lockfile.Lock/Unlock` (fan_in 3: taskledger + settings.go + glm_tools.go); `@MX:ANCHOR` + `@MX:REASON` + `@MX:SPEC: SPEC-CLIFIX-CRITICAL-001` on `taskledger.ClaimTask`; `@MX:NOTE` on the Windows in-process-mutex limitation.

Exit-gate evidence (verbatim tails):

```
$ go build ./... ; GOOS=windows GOARCH=amd64 go build ./... ; go vet ./...
build exit=0 / winbuild exit=0 / vet exit=0
$ go test ./...
exit=0   (96 packages ok; full log: .moai/state/verify/atr-run/m0-test.log)
$ go test -cover ./internal/lockfile/ ./internal/cli/taskledger/
ok  github.com/modu-ai/moai-adk/internal/lockfile      0.524s  coverage: 100.0% of statements
ok  github.com/modu-ai/moai-adk/internal/cli/taskledger 0.990s  coverage: 92.7% of statements
$ go test -v -run 'TestClaimTaskAppend_Repro' ./internal/cli/
=== RUN   TestClaimTaskAppend_Repro
--- PASS: TestClaimTaskAppend_Repro (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	0.503s
$ go test -v -run 'TestClaimTask' ./internal/cli/taskledger/ | grep -c -- '--- PASS'
4
$ golangci-lint run --timeout=3m
0 issues.   (baseline: 0 issues — no NEW lint)
```

REQ-ATR-003 gate satisfied: whole-repo green BEFORE any deletion milestone.

### M1 — Phases 1-2 Go removal (REQ-ATR-004/005 + compile-coupled test edits)

M0 commit SHA: `d296f11cc` (precedes this M1 commit — AC-ATR-005 ordering evidence).

Removed: `internal/cli/team_spawn.go` + `team_spawn_test.go` (with the M0 lock-file
moves, all five `team_spawn*` files are now gone — AC-ATR-006);
`internal/config` `TeamConfig` / `RoleProfileEntry` / `TeamAutoSelectionConfig`
type decls, `WorkflowConfig.Team` field, `AutoSelectionLegacy` field
(compile-coupled — typed `TeamAutoSelectionConfig`), `defaults.go` `Team:` block
(GitStrategy `Team: ModeProfile` at defaults.go PRESERVED),
`workflow_accessors.go` `WorkflowTeamAutoSelection()`;
`internal/config/workflow_role_profiles_test.go` (file deleted);
`types_test.go` `TestTeamConfigStructShape` + `AutoSelectionLegacy` assertions +
`TeamAutoSelection` accessor case; `defaults_test.go` workflow-Team rows
(GitStrategy AC-GSS-005 rows untouched); `workflow_nested_test.go` team
assertions removed / partial-yaml fixture re-targeted to worktree keys /
`TestWorkflowConfigInconsistentRoleProfileKeys` deleted;
`internal/web/agent_settings_test.go` `TestRoleProfileEntryHasNoEffortField`
deleted (compile-coupled — referenced `config.RoleProfileEntry`; run-phase
discovery, not in the plan-time M1 enumeration).

agentlint repurpose (folded into M1 per plan.md §F M1 step 3 compile-necessity
decision): `workflow_lint.go` `validateRoleProfiles` + `writeHeavyRoles` removed;
NEW `validateModelRoutingProfiles` reusing `config.(*Config).ValidateModelRoutingProfiles`
(REQ-ATR-009); sentinel `ORC_WORKTREE_REQUIRED` → `MODEL_ROUTING_INVALID` in
`sentinels.go`; `workflow_lint_test.go` rewritten (6 tests: valid / absent-block /
invalid-model / invalid-perfTier / invalid-key + RunE violation-path asserting
`errLintViolations` + RunE clean-path). Exit-code contract (0/1/2/3) preserved.

Exit-gate evidence (verbatim tails):

```
$ go build ./... ; GOOS=windows GOARCH=amd64 go build ./... ; go vet ./...
build exit=0 / winbuild exit=0 / vet exit=0
$ go test ./...
exit=0   (96 packages ok, 0 FAIL; full log: .moai/state/verify/atr-run/m1-test.log)
$ golangci-lint run --timeout=3m
0 issues.   (baseline: 0 issues — no NEW lint)
```

AC evidence (M0/M1-scope ACs, run against the post-M1 tree):

```
AC-ATR-001  ls internal/lockfile/ → lockfile_unix.go lockfile_windows.go lockfile_unix_test.go
            grep -l 'go:build !windows' → lockfile_unix.go(+test); 'go:build windows' → lockfile_windows.go
            grep -rn "cross-process" internal/lockfile/ → lockfile_unix.go:7 (limitation comment preserved
            verbatim in lockfile_windows.go body)                                          PASS
AC-ATR-002  GOOS=windows GOARCH=amd64 go build ./... → exit=0                              PASS
AC-ATR-003  === RUN TestClaimTaskAppend_Repro / --- PASS / ok internal/cli 0.528s          PASS
AC-ATR-004  grep -c '--- PASS' (go test -v -run TestClaimTask ./internal/cli/taskledger/) → 4 (≥2)  PASS
AC-ATR-005  M0 d296f11cc precedes M1 commit (this commit); §E.2 M0 records go test exit 0  PASS
AC-ATR-006  ls internal/cli/team_spawn* → no matches, exit=1                               PASS
AC-ATR-007  grep -rn -E "TeamConfig|RoleProfileEntry|TeamAutoSelectionConfig" internal/config/ --include="*.go" | wc -l → 0   PASS
AC-ATR-008  grep -n 'yaml:"team"' internal/config/types.go → 1 hit at :153 — DISAMBIGUATION:
            survivor is GitStrategyConfig `Team ModeProfile` (PRESERVE per E2 edge case);
            grep -n 'Team: TeamConfig' internal/config/defaults.go | wc -l → 0             PASS
AC-ATR-009  grep -c "type RoleProfile struct" types.go → 1; "Sandbox string" → 1           PASS (preserved)
AC-ATR-010  grep -n "Team     ModeProfile" types.go | wc -l → 1; go test -run 'Defaults'
            ./internal/config/ → ok; grep -c "Team.Automation.AutoPush" defaults_test.go → 2  PASS (preserved)
AC-ATR-011  MergeTeamCheckpoints → 1; worktree team_launch/swarm_registry/handoff_guidance
            files present; teammateMode in glm.go/launcher.go → 12                         PASS (preserved)
AC-ATR-012  f.git_strategy.team.* count → 16 (== pre-flight baseline 16, unchanged);
            f.git_strategy.mode.opt.team → 4 (≥1)                                          PASS (preserved)
AC-ATR-014  grep -c "role_profiles" workflow_lint.go → 0; grep -c -i "ModelRouting" → 8;
            go test ./internal/cli/agentlint/... → ok (88.3% cover), incl. RunE
            violation-path test asserting errLintViolations                                PASS
AC-ATR-019  build trio exit 0 (partial — full re-verify due at M4); grep workflow_role_profiles_test → 0   PASS (partial)
```

Baselines recorded at pre-flight (for M2+ delta discipline): `f.workflow.team.` = 264
(removal target, untouched in M0/M1); `f.git_strategy.team.` = 16 (PRESERVE, unchanged).

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase — populated by manager-develop>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_

## Plan Audit Gate (Phase 0.5 input)

- plan_complete_at: 2026-07-11T00:00:00+09:00
- plan_status: audit-ready
- plan_audit: iter-2 PASS 0.89 (iter-1 FAIL 0.86 → D1-D8 fixed in v0.1.2, commit 254c695e5)
- skip_eligible: no (< 0.90)
- run-phase carry-over cautions: R1 (schema_sections.go:412 RawViewBlocks workflow.team.patterns), R2 (spec-workflow.md:107 Mode Dispatch row + CLAUDE.md §4/§15 manual sweep), R3 (delivery.md — only :393 "## Team Mode" is Agent-Teams scope; :285-338 are preserved GitStrategy team-profile prose)

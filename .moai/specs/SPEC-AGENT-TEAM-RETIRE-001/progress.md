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

M0 commit SHA: `d296f11cc`; M1 commit SHA: `995948330` (backfilled — a commit
cannot reference its own hash; AC-ATR-005 ordering: M0 precedes M1 in git log).

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

### M2 — Phase 3 web console team removal (REQ-ATR-008/009)

Commit: `feat(SPEC-AGENT-TEAM-RETIRE-001): M2 — remove Agent Teams web console surfaces`
(SHA in git log — a commit cannot reference its own hash). Phase 4 (agentlint
repurpose) was folded into M1 (AC-ATR-014 PASS); M2 covers the web console team
surfaces only.

Scope note: this milestone ran in an L1 worktree (`worktree-agent-a81cea6d62369b9b4`,
base `ea6117c2c`) autonomously materialized by the Claude Code runtime — the
orchestrator integrates the branch to main. The unrelated origin/main +3 commits
(SPEC-CLIFIX-CONCURRENCY-001 sync-phase) touch no M2 file (verified: empty
`git diff --stat HEAD..origin/main -- <M2 files>`); origin/main i18n still carried
264 `f.workflow.team` keys pre-M2 (no parallel-session duplication).

Removed (Go, `internal/settings` + `internal/web`):
- `schema_sections.go`: 5 `workflow.team` seam FieldDef rows (enabled / delegate_mode /
  max_teammates / default_model / require_plan_approval); `agentSettingsFields()`
  (구 surface (b) team.role_profiles — 7 profiles × {model,effort,isolation,mode});
  `RoleProfileNames()` + its only caller; `v4IsolationValues()` (dead after
  agentSettingsFields removal); the `workflow.team.patterns` RawViewBlocks entry
  (R1 auditor finding — the `:412` surface).
- `schema.go`: `SectionAgentSettings` const + its `AllSections()` entry.
- `schemaform.go`: `agent_settings` consoleTab + `SectionAgentSettings` schemaSectionMetas
  entry; `sec.workflow.desc` baseline reworded to drop "team".
- `schema_label.go`: the dead `role_profiles.<role>.<field>` label-expansion branch
  (REQ-ATR-008 named `schema.go` + `schema_label.go` consumers, both swept).
- `internal/web/assets/i18n.js`: all 66 `f.workflow.team.*` keys × 4 locales = 264
  lines removed; `sec.workflow.desc` reworded ×4 (team / 팀 / チーム / 团队 dropped).

Preserved (verified STILL-EXISTS): `f.git_strategy.team.*` (16) +
`f.git_strategy.mode.opt.team` (4) — GitStrategy namesake, unchanged; sub-agent
frontmatter surface (`agentfm.*`, `internal/web/agentfm.go`) untouched.

Test reconciliation: `TestRoleProfileEditRoutesThroughSeam` + `TestAgentSettingsEnumReject`
deleted (their subject surfaces removed); `TestAgentSettingsFourSurfacesRendered`
surface-(b) role_profiles assertions removed (surfaces a/c/d retained); NEW
`TestNoTeamRoleProfileRender` (absolute-absence guard, REQ-ATR-008); 5 seam-routing /
render tests re-targeted `workflow.team.max_teammates` / `workflow.team.patterns` →
surviving `workflow.token_budget.plan` / `harness.levels`. `agentfm` enum-reject
coverage retained by `TestAgentFMValidationAndAbsent`.

templ: NO regeneration — `fieldsets.templ` has no team fieldset (schema-driven since
WEB-CONSOLE-011 M2b; the 2 residual `SectionAgentSettings` refs are stale comments,
not executable); `git diff --stat --exit-code fieldsets_templ.go` clean.

Exit-gate evidence (verbatim tails):

```
$ go build ./... ; GOOS=windows GOARCH=amd64 go build ./... ; go vet ./...
build exit=0 / winbuild exit=0 / vet exit=0
$ go test ./...
exit=0   (0 FAIL, all packages ok; full log: /tmp/atr-m2/fulltest.log)
$ golangci-lint run --timeout=3m ./internal/settings/... ./internal/web/...
0 issues.   (baseline: 0 issues at pre-flight — no NEW lint)
$ node --check internal/web/assets/i18n.js
JS_OK (valid JS after 264-line removal; trailing comma before `}` is valid ES)
```

AC evidence (M2-scope, post-M2 tree):

```
AC-ATR-012  grep -c '"f\.git_strategy\.team\.' i18n.js → 16 (== pre-flight baseline);
            '"f\.git_strategy\.mode\.opt\.team"' → 4 (≥1)                              PASS (preserved)
AC-ATR-013  grep -c '"f\.workflow\.team\.' i18n.js → 0;
            grep -c '"workflow", "team"' schema_sections.go → 0;
            grep -rn "RoleProfileNames" internal/ --include="*.go" | wc -l → 0;
            templ conditional N/A (0 team markup in .templ; fieldsets_templ.go clean)  PASS
AC-ATR-019  build trio exit 0 + full go test ./... exit 0 (0 FAIL)                     PASS (partial — full re-verify at M4)
```

NEW-test ran (vacuous-green guard): `=== RUN TestNoTeamRoleProfileRender / --- PASS`.

Residual (for M3+ delta discipline): `sec.agent_settings.*` (×4) + `f.v4.*.opt.*`
i18n keys are now orphaned (no render reference after the section removal) — LEFT
in-place as harmless dead keys (minimal-scope choice; keeps `TestAgentFMWarnI18nParity`
green with zero test churn). A follow-up i18n cleanup or the sync phase MAY prune them.

### M3 — Phases 5-6-7 rules + skills + template mirror (REQ-ATR-010/011/012/013/022)

Commit: `feat(SPEC-AGENT-TEAM-RETIRE-001): M3 — remove team rules + skills, migrate glm.md, template mirror` (SHA in git log — a commit cannot reference its own hash). Ran in an L1 worktree (`worktree-agent-a4d53193547131922`, base `db4c08aa7`); orchestrator integrates to main.

Migrate-then-delete (REQ-ATR-022, ordering: migrate FIRST, delete in same commit): NEW `## CG Mode (Claude + GLM teammates)` section (mechanism/LLM-mode-detection/prereqs/tmux-env/error-recovery/cleanup) added to `.claude/rules/moai/core/glm-web-tooling.md` (both trees) — Agent-Teams orchestration prose NOT migrated. Then deleted all 6 `.claude/skills/moai/team/{plan,run,sync,debug,review,glm}.md` (both trees).

Deleted (both trees): `.claude/rules/moai/workflow/{team-protocol.md,team-pattern-cookbook.md}` (2 rules) + `.claude/skills/moai/team/` (6 skills) = 4+12 `git rm`.

Rule reframes (both trees): `orchestration-mode-selection.md` (Mode 3 catalog row + §B decision-tree branch + §C.1 gate + §C.2 + §E anti-pattern + §G.1 crosswalk → `Mode 3 — RETIRED` tombstone; §B.1 `auto_selection` pointer removed → prose-only SSOT, D8); `spec-workflow.md` (:107 Mode Dispatch auto-select row + `### Team Mode Methodology` + Phase 0.5 gate-entry team/run.md ref + entire `## Agent Teams Variant` section → RETIRED tombstone); `settings-management.md` (`## Agent Teams Settings` → RETIRED); `agent-authoring.md` (role-profile `Requires:` env line → retirement note); `model-policy.md` (:65 team-protocol/team/run.md cross-ref removed); `NOTICE.md` (team-examples cookbook attribution row de-referenced); `spec-frontmatter-schema.md` (:11 team/plan.md cross-ref removed); `worktree-integration.md` (`## Team Protocol` section removed + Version footer); `dynamic-workflows.md` (:136 team-protocol/cookbook cross-ref → orchestration-mode-selection §C.1). CLAUDE.md §4 (Dynamic Team Generation → RETIRED) + §15 (Agent Teams → RETIRED + CG Mode preserved), both trees.

Skill reframes (both trees, `.claude/skills/moai/workflows/`): removed all `--team` routing to deleted `team/*.md` + `## Team Mode` sections + `workflow.team.enabled` / `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` prereq refs across `run.md`, `moai.md`, `fix.md`, `mx.md`, `plan.md`, `review.md`, `sync.md`, `plan/spec-assembly.md`, `run/{mode-orchestration,context-loading,task-decomposition,phase-execution}.md`, `sync/delivery.md`. run.md `--mode` dispatch axis retained the `MODE_UNKNOWN` + `MODE_TEAM_UNAVAILABLE` sentinels (CI audit `agentless_audit_test.go`) while reframing `--mode team` as retired → emits `MODE_TEAM_UNAVAILABLE` → falls back to autopilot.

R3 CRITICAL (delivery.md): ONLY the `:393 ## Team Mode` Agent-Teams section (+ its `team/sync.md` ref) was deleted. The `:285-338` GitStrategy team ModeProfile prose (PR Ready Transition / draft→ready / approvals — defaults_test.go `Team.DraftPR` / `Team.RequiredReviews` preservation invariant) was PRESERVED. delivery.md is the sole documented GitStrategy `team mode` named-exception survivor in the workflows tree.

Template (Phase 7): removed `team:` block (auto_selection + role_profiles + patterns) from `internal/template/templates/.moai/config/sections/workflow.yaml`. `make build` regenerated the go:embed FS (catalog.yaml byte-identical → no hash cascade). Removal-coupled Go test reconciliation (necessary consequence of the deletions, NOT new logic): `skills_audit_test.go` team/run.md AC-WAG-05 case removed; `template_neutrality_audit_test.go` stale `team/run.md` allowlist entry removed.

Exit-gate evidence (verbatim tails):

```
$ go build ./... ; GOOS=windows GOARCH=amd64 go build ./... ; go vet ./...
build exit=0 / winbuild exit=0 / vet exit=0
$ go test ./...
exit=0   (all packages ok, 0 FAIL; full log: .moai/state/verify/atr-m3/1-gotest.log)
$ golangci-lint run --timeout=3m
0 issues.   (baseline: 0 issues at pre-flight — no NEW lint)
$ go test ./internal/template/...
ok  github.com/modu-ai/moai-adk/internal/template  (neutrality + leak + mirror-parity + split-namespace guards PASS)
```

AC evidence (M3-scope, post-M3 tree, both trees unless noted):

```
AC-ATR-015  team rules absent (per-tree [ -e ] loop → no RESIDUE); grep "Mode 3 — RETIRED"
            orchestration-mode-selection.md → 2 (>=1); dangling-ref sweep (5 .claude subtrees +
            template, --include=*.md, excl /team/) → 0                                      PASS
AC-ATR-016  team skills dir absent (per-tree [ -d ] → no RESIDUE); grep "Agent Teams 모드"
            run.md → 0; grep -i "## Team Mode" delivery.md → 0; case-insensitive "team mode"
            workflows/ sweep → only delivery.md (GitStrategy named exception, R3-preserved)  PASS
AC-ATR-017  make build exit 0; go test ./internal/template/... ok (neutrality/leak/mirror/split);
            grep "^    team:" template workflow.yaml → 0                                     PASS
            NOTE (deviation): local .moai/config/sections/workflow.yaml role_profiles/team block
            LEFT UNTOUCHED per the delegation instruction (§22 dev-local exempt) — AC-ATR-017's
            "role_profiles: local tree too → 0" sub-check is a documented deviation, not a defect;
            the DEPLOYED template (what users receive) has 0 role_profiles/team (verified by AC-ATR-018)
AC-ATR-018  bin/moai init sandbox → no team-protocol.md / no team-pattern-cookbook.md /
            no team skills dir / workflow.yaml team: block = 0 / CG token deployed = 1       PASS (resurrection-negative)
AC-ATR-027  grep "CG Mode (Claude + GLM" glm-web-tooling.md both trees → 1,1 (baseline 0 — non-vacuous);
            [ -e team/glm.md ] → removed (migrate-then-delete ordering: migration + deletion in same M3 commit)  PASS
AC-ATR-028  grep "auto_selection" orchestration-mode-selection.md both trees → 0,0;
            grep "≥ 3 domains" → 1,1 (prose threshold RETAINED as sole SSOT)                PASS
AC-ATR-029  key-token sweep (workflow.team.enabled|CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS) across the
            9 enumerated files both trees → 0; CLAUDE.md both trees → 0 (Mode 3 tombstone in
            orchestration-mode-selection.md is the sole allowed survivor, and it too carries 0)  PASS
```

Baselines recorded at pre-flight (delta discipline): CG token both trees = 0 (→ 1 after migration); `Mode 3 — RETIRED` = 0 (→ 2 after tombstone); `auto_selection` orchestration-mode-selection = 1 (→ 0 after D8 removal).

### M4 — Phase 8 final sweep + Scope B workflows (REQ-ATR-014..021)

Commit: `feat(SPEC-AGENT-TEAM-RETIRE-001): M4 — Scope B workflows (sync-audit-4dim + plan-research-fanout) + final sweep` (SHA in git log — a commit cannot reference its own hash). Ran in the L1 worktree `worktree-agent-a480091fd6e27c4e2`, base `0303e8c75` (== origin/main, 0/0 divergence verified pre-flight); orchestrator coordinates integration to main. FINAL milestone.

**Template-mirror decision: DO NOT MIRROR** (verified). `.claude/workflows/` is user-owned and NOT template-managed — `internal/template/templates/.claude/workflows/` does not exist and the template tree carries no `*.js` workflow files (the reference `codemaps-extract.js` is itself local-only, not mirrored). Aligns with design.md D7, spec §E "Out of Scope — Template-shipping the Scope B workflows", plan.md §D Forbidden, and dynamic-workflows.md § MoAI Integration Notes (user-owned `.claude/workflows/`). AC-018 fresh-init confirmed `.claude/workflows/` is NOT deployed by `moai init`. The template tree was NOT touched in M4, so the template-neutrality CI guard is not triggered.

Files (LOCAL tree only): NEW `.claude/workflows/sync-audit-4dim.js` (Context → 4 parallel read-only Judges → in-script harmonic-mean Verdict; INCOMPLETE-on-null-judge + zero-score guard; Tier M/L gate via args.tier; anti-patterns codified in header: no meta-judge / no LLM arithmetic / no Write-Edit judges); NEW `.claude/workflows/plan-research-fanout.js` (3-4 read-only lens Explorers, fixed-heading markdown + mandatory `### confidence_and_gaps`; Synthesizer marks cross-lens contradictions; `>= 2` null-lens abort → `insufficient_coverage`; lens hard-cap `slice(0, 4)`; no in-workflow writes — research.md written OUTSIDE by manager-spec/orchestrator). Both follow the `codemaps-extract.js` house style: `export const meta` + `phase()` markers + `agent()`/`parallel()` + `args` defaults + `.filter(Boolean)` null-filtering + determinism (no wall-clock / no random CALLS).

Whole-repo verification batch (verbatim tails; logs under `.moai/state/verify/atr-m4/`):

```
$ go build ./... ; GOOS=windows GOARCH=amd64 go build ./... ; go vet ./...
build exit=0 / winbuild exit=0 / vet exit=0
$ go test ./...
go test exit=0   (0 FAIL lines; scripts/i18n-validator ok last)
$ golangci-lint run --timeout=3m
0 issues.   (baseline: 0 issues at pre-flight — no NEW lint)
$ go test -cover ./internal/lockfile/ ./internal/cli/taskledger/
ok  internal/lockfile      coverage: 100.0% of statements
ok  internal/cli/taskledger coverage: 92.7% of statements
```

Node syntax checks (both Scope B scripts): `node --check .claude/workflows/sync-audit-4dim.js → exit 0`; `node --check .claude/workflows/plan-research-fanout.js → exit 0` (node v22.14.0).

Full M0-M4 AC matrix (all re-verified against the post-M4 tree this milestone unless marked carry-over):

```
AC-ATR-001  lockfile pkg (unix+windows+test); go:build !windows + windows tags; cross-process comment preserved   PASS (re-verified M4)
AC-ATR-002  GOOS=windows GOARCH=amd64 go build ./... → exit 0                                                       PASS (re-verified M4)
AC-ATR-003  go test -run TestClaimTaskAppend_Repro ./internal/cli/ → ok (migrated repro green)                       PASS (re-verified M4)
AC-ATR-004  go test -v -run TestClaimTask ./internal/cli/taskledger/ | grep -c -- '--- PASS' → 4 (≥2)                PASS (re-verified M4)
AC-ATR-005  git log --grep → M0 279d2f688 precedes M1 aff4a2537 (migration-before-deletion ordering)                 PASS (re-verified M4)
AC-ATR-006  ls internal/cli/team_spawn* → no matches                                                                  PASS (re-verified M4)
AC-ATR-007  grep TeamConfig|RoleProfileEntry|TeamAutoSelectionConfig internal/config/ → 0                             PASS (re-verified M4)
AC-ATR-008  yaml:"team" types.go → 1 (GitStrategy survivor, disambiguated); Team: TeamConfig defaults.go → 0         PASS (re-verified M4)
AC-ATR-009  type RoleProfile struct → 1; Sandbox string → 1 (PRESERVED)                                               PASS (re-verified M4)
AC-ATR-010  Team ModeProfile → 1; go test -run Defaults ./internal/config/ → ok; Team.Automation.AutoPush → 2       PASS (re-verified M4)
AC-ATR-011  MergeTeamCheckpoints → 1; worktree team_launch/swarm_registry/handoff_guidance present; teammateMode → 12  PASS (re-verified M4)
AC-ATR-012  f.git_strategy.team.* → 16 (== baseline, unchanged); f.git_strategy.mode.opt.team → 4 (PRESERVED)        PASS (re-verified M4)
AC-ATR-013  f.workflow.team.* → 0; RoleProfileNames → 0; "workflow","team" FieldDef → 0                              PASS (re-verified M4)
AC-ATR-014  role_profiles workflow_lint.go → 0; ModelRouting → 8; agentlint tests ok                                 PASS (re-verified M4)
AC-ATR-015  team rules absent (per-tree [ -e ]); Mode 3 — RETIRED tombstone → 2; dangling-ref sweep (5 subtrees + template) → 0   PASS (re-verified M4)
AC-ATR-016  team skills dir absent (per-tree); "Agent Teams 모드" run.md → 0; "## Team Mode" delivery.md → 0          PASS (re-verified M4)
AC-ATR-017  make build exit 0 (M3); go test ./internal/template/... ok; template workflow.yaml team: → 0            PASS-WITH-DEBT (deployed template clean per AC-018; local .moai/config/sections/workflow.yaml role_profiles left untouched per §22 dev-local exempt — M3 documented deviation, not a new M4 defect)
AC-ATR-018  bin/moai init sandbox → no team-protocol / no team skills dir / workflow.yaml team: 0 / CG token 1; .claude/workflows/ NOT deployed (user-owned)   PASS (re-verified M4 — fresh build+init)
AC-ATR-019  go build ./... exit 0 + GOOS=windows build exit 0 + go test ./... exit 0 (0 FAIL); workflow_role_profiles_test → 0   PASS (FULL re-verify at M4, closing the M1/M2 partial)
AC-ATR-020  node --check exit 0; title: → 3; Functionality|Security|Craft|Consistency → 12 (≥4); 0.85 → 2; 1[space]*/ → 3   PASS (M4)
AC-ATR-021  INCOMPLETE → 3; meta-judge → 2; agentType: 'Explore' → 5 (1 Context + 4 Judges, all read-only); label: 'judge: → 4   PASS (M4)
AC-ATR-022  args.tier → 3; "Tier S" (case-insensitive header gate) → 2                                               PASS (M4)
AC-ATR-023  node --check exit 0; lens tokens → 6 (≥4); confidence_and_gaps → 3; effort:'high' → 1; effort:'medium' → 1; "NONE found" → 7   PASS (M4)
AC-ATR-024  insufficient_coverage → 3; ">= 2" abort threshold → 3                                                    PASS (M4)
AC-ATR-025  slice(0, 4) → 2; effort:'xhigh' → 0; "research.md is written" → 1                                        PASS (M4)
AC-ATR-026  export const meta → 1 each; Date.now(/Math.random( CALLS → 0 both; label:  → 5 (sync-audit) / 2 (plan-research)   PASS (M4)
AC-ATR-027  CG Mode token glm-web-tooling.md both trees → 1,1; team/glm.md deleted (migrate-then-delete)             PASS (re-verified M4)
AC-ATR-028  auto_selection orchestration-mode-selection.md both trees → 0,0; "≥ 3 domains" prose → 1 (sole SSOT)     PASS (re-verified M4)
AC-ATR-029  key-token sweep (workflow.team.enabled|CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS) across 9 enumerated files both trees → 0   PASS (re-verified M4)
```

29/29 AC PASS (AC-017 PASS-WITH-DEBT — deployed-clean, local-tree deviation is a §22 dev-local exemption documented at M3, not a new M4 regression). No Go source, template, rule, or skill files were modified in M4 (only 2 new local `.claude/workflows/*.js` + progress.md), so the M0-M3 ACs are structurally guaranteed unchanged and were additionally re-verified above.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-11T00:00:00+09:00
run_commit_sha: pending-backfill-M4   # progress.md is IN the M4 commit; a commit cannot reference its own hash — orchestrator/next-commit backfills the real M4 SHA (D3 SHA-placeholder exemption, spec-frontmatter-schema.md)
run_status: complete
ac_pass_count: 29        # 28 clean PASS + AC-ATR-017 PASS-WITH-DEBT (deployed template clean; local-tree role_profiles §22 dev-local exemption)
ac_fail_count: 0
preserve_list_post_run_count: 7   # STILL-EXISTS at exit: config RoleProfile(Sandbox) / GitStrategy Team ModeProfile+tests / session MergeTeamCheckpoints / worktree team_launch+swarm+handoff / teammateMode glm.go+launcher.go / f.git_strategy.team.* (16) + mode.opt.team (4) / glm-web-tooling CG Mode
l44_pre_commit_fetch: worktree HEAD 0303e8c75 == origin/main (git rev-list --count --left-right → 0 0, verified pre-flight)
l44_post_push_fetch: n/a — agent does not push (orchestrator coordinates worktree→main integration per delegation instruction)
new_warnings_or_lints_introduced: 0   # golangci-lint 0 issues (== pre-flight baseline); go vet exit 0
cross_platform_build:
  darwin_amd64: exit 0
  windows_amd64: exit 0   # GOOS=windows GOARCH=amd64 go build ./...
total_run_phase_files:
  m4_new: 2               # .claude/workflows/{sync-audit-4dim.js, plan-research-fanout.js} — LOCAL only, NOT template-mirrored
  m4_progress: 1          # this progress.md §E.2/§E.3 update
  m0_to_m4_cumulative: 116  # git diff --stat 8f7234a76..HEAD -- internal/ .claude/rules/ .claude/skills/ internal/template/ (2587 ins / 6786 del)
m1_to_mN_commit_strategy: per-milestone separate commits (M0 279d2f688 → M1 aff4a2537 → M2 → M3 0303e8c75 → M4 this) with specific-path git add (never git add -A/-u); no --amend, no force-push; agent commits in the L1 worktree, orchestrator coordinates integration to main
scope_b_workflows:
  sync_audit_4dim: node --check exit 0; 3 phases (Context/Judge/Verdict); 5 read-only Explore agents (1+4); in-script harmonic mean n/Σ(1/sᵢ) with zero-score guard + INCOMPLETE branch; Tier M/L gate
  plan_research_fanout: node --check exit 0; 2 phases (Explore/Synthesize); 3-4 read-only lens Explorers (effort medium) + 1 Synthesizer (effort high); >=2 null-lens abort; lens cap slice(0,4); no in-workflow writes
template_mirror_decision: DO-NOT-MIRROR (.claude/workflows/ user-owned, not template-managed; template tree untouched; neutrality CI guard not triggered)
```

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — populated by manager-docs>_

## Plan Audit Gate (Phase 0.5 input)

- plan_complete_at: 2026-07-11T00:00:00+09:00
- plan_status: audit-ready
- plan_audit: iter-2 PASS 0.89 (iter-1 FAIL 0.86 → D1-D8 fixed in v0.1.2, commit 254c695e5)
- skip_eligible: no (< 0.90)
- run-phase carry-over cautions: R1 (schema_sections.go:412 RawViewBlocks workflow.team.patterns), R2 (spec-workflow.md:107 Mode Dispatch row + CLAUDE.md §4/§15 manual sweep), R3 (delivery.md — only :393 "## Team Mode" is Agent-Teams scope; :285-338 are preserved GitStrategy team-profile prose)

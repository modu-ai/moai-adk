# Plan — SPEC-AGENT-TEAM-RETIRE-001

> Implementation plan for the Agent Teams static-layer retirement + the two
> replacement dynamic workflows. Tier L. Milestones M0-M4 map the 10 removal
> phases + Scope B. All anchors below were verified against the working tree at
> plan authoring time (2026-07-11); line numbers cited are drift-prone — re-anchor
> by content token at run-phase (feedback: line-number drift asymmetry).

## §A. Context

- **Branch / baseline**: `main` (verify HEAD at run-phase pre-flight; do not
  assume the plan-time SHA).
- **SPEC artifacts**: `.moai/specs/SPEC-AGENT-TEAM-RETIRE-001/{spec,plan,acceptance,design,research,progress}.md`.
- **Verified inventory** (plan-time, full evidence in research.md §B):
  - `internal/cli/team_spawn.go` 15,888 B; 15 funcs; **0 non-test callers**.
  - `internal/cli/team_spawn_test.go` 20,379 B (13+ tests incl. `TestClaimTask`,
    `TestClaimTaskConcurrent`).
  - Lock pair: `team_spawn_lock_unix.go` (377 B, `lockFile`/`unlockFile` via
    `syscall.Flock`) + `team_spawn_lock_windows.go` (in-process per-path mutex
    fallback) + `team_spawn_lock_unix_test.go`.
  - `internal/config/types.go`: `WorkflowConfig.Team` field (yaml:"team",
    ~:328), `TeamConfig` (~:435), `RoleProfileEntry` (~:447),
    `TeamAutoSelectionConfig` (~:505); PRESERVE `RoleProfile` (~:474, Sandbox
    field) + GitStrategy `Team ModeProfile` (~:153).
  - `internal/config/defaults.go`: `Team: TeamConfig{...}` block (~:457-513);
    PRESERVE `Team: ModeProfile{...}` (~:353).
  - `internal/config/workflow_accessors.go` ~:35 `WorkflowTeamAutoSelection()`
    (non-test source consumer — auditor D3) + `internal/config/types_test.go`
    (~:349-429 `TestTeamConfigStructShape`, ~:511-521 accessor tests) — REMOVE.
  - Web: `internal/web/assets/i18n.js` (`f.workflow.team.*` family REMOVE /
    `f.git_strategy.team.*` + `f.git_strategy.mode.opt.team` PRESERVE);
    `internal/settings/schema_sections.go` `RoleProfileNames()` (~:54-56, consumed
    ~:337 + `schema.go` + `internal/web/schema_label.go`) + five `workflow.team`
    FieldDef rows (~:212-216). `internal/web/fieldsets.templ`: **0 team hits at
    plan-time** (schema-driven UI — auditor D2 verified); templ regen conditional
    on run-phase re-measure only.
  - agentlint: `internal/cli/agentlint/workflow_lint.go` `validateRoleProfiles`
    (~:70-86); repurpose target `internal/config/model_routing.go`
    `ValidateModelRoutingProfiles`.
  - Rules: `team-protocol.md` (7,764 B) + `team-pattern-cookbook.md` (14,757 B),
    both trees. Skills: `.claude/skills/moai/team/` 6 files, both trees.
    Workflow team refs: `workflows/{plan,run,fix,review,mx}.md` (sync.md: none).
  - Template config: `internal/template/templates/.moai/config/sections/workflow.yaml`
    `team:` block (~:30); local `.moai/config/sections/workflow.yaml`
    `role_profiles:` (~:101).
- **Resolved decisions** (user, 2026-07-11, orchestrator-relayed — no open
  clarifications remain):
  - **team/glm.md → migrate-essentials-then-delete**: glm.md stays in the
    Phase 6 deletion set; M3 gains a preceding migration step moving the
    essential CG Mode (GLM teammate) guidance (LLM mode detection,
    prerequisites, tmux env vars, error recovery) into
    `.claude/rules/moai/core/glm-web-tooling.md` (both trees) BEFORE deletion.
    Encoded as REQ-ATR-022 + AC-ATR-027 (grep token `CG Mode (Claude + GLM`,
    verified 0-count in the target at plan time).
  - **Auto-select thresholds → prose-only SSOT**: design.md D8 adopted.
    `orchestration-mode-selection.md` §B.1 prose becomes the sole SSOT; the
    "machine-readable source is workflow.yaml auto_selection" pointer sentence
    is removed with `TeamAutoSelectionConfig`. Encoded in REQ-ATR-010
    (extended) + AC-ATR-028.

## §B. Known Issues (filtered, Tier L)

- **B1 Cross-platform build tags**: the lock pair is build-tag split
  (`//go:build !windows` / `windows`). The `internal/lockfile` migration MUST
  keep the pair; verify `GOOS=windows GOARCH=amd64 go build ./...` after M0 and
  after M1 (Windows lock fallback references ClaimTask docs in comments).
- **B2 Cross-SPEC conflict pre-scan**: this SPEC partially reverses
  SPEC-V3R6-AGENT-TEAM-REBUILD-001 (static team layer). State the reversal
  explicitly in commits. SPEC-CLIFIX-CRITICAL-001's `TestClaimTaskAppend_Repro`
  MUST survive via the taskledger migration — deleting it is prohibited.
- **B5 CI 3-tier awareness**: spec-lint, golangci-lint, per-OS test jobs fail
  independently; template-neutrality-check.yaml triggers on
  `internal/template/templates/**` paths (M3 will trigger it).
- **B6 spec-lint heading convention**: retained rule edits must not orphan
  `### Out of Scope` H3 structures in touched SPEC files (not expected here).
- **B8/B10 working-tree hygiene + PRESERVE scope**: several unrelated untracked
  files exist at plan-time (`.moai/reports/*`, `.moai/lessons-inbox.jsonl`,
  other SPEC dirs). Commit specific paths only. Do not touch `.moai/state/`,
  parallel SPEC dirs, or the §D PRESERVE list.
- **Substring-collision hazard (this SPEC's #1 risk)**: "team" is a collision-rich
  token — `f.git_strategy.team.*`, GitStrategy `Team ModeProfile`,
  `MergeTeamCheckpoints`, `team_launch*`, `TaskLedger`("AppendTask" substring),
  `teammateMode`. EVERY removal grep must be prefix/word-anchored; a bare
  `grep team` sweep is prohibited (REQ-ATR-007; feedback: digit-boundary /
  substring-collision lessons).
- **Generated-file discipline**: `fieldsets_templ.go` is generated from
  `fieldsets.templ`. Edit the `.templ` source and regenerate; hand-editing only
  the `_templ.go` file will be reverted by the next `templ generate`.

## §C. Pre-flight Checklist (run before any change)

```bash
# 1. Baseline
git branch --show-current && git rev-parse HEAD
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
golangci-lint run --timeout=2m 2>&1 | tail -5   # lint baseline (NEW vs pre-existing)

# 2. Re-verify the 0-non-test-caller claim (must still hold)
grep -rn -E "InitTeamState|ArchiveTeamState|LoadRoleProfiles|ValidateSpawn|ValidateRoster|MailboxMessage|TaskClaimer" \
  internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "internal/cli/team_spawn"
# expected: only unrelated substring hits (session AppendTaskLedger); investigate any new hit

# 3. Measure removal baselines (counts recorded in progress.md §E.2 at run-phase)
grep -c '"f\.workflow\.team\.' internal/web/assets/i18n.js
grep -c '"f\.git_strategy\.team\.' internal/web/assets/i18n.js   # PRESERVE count — must be unchanged at exit

# 4. PRESERVE-target existence snapshot
ls internal/cli/worktree/team_launch* internal/cli/worktree/swarm_registry.go
grep -n "type RoleProfile struct" internal/config/types.go
grep -n "MergeTeamCheckpoints" internal/session/store.go

# 5. Retired/superseded conflict scan
grep -rn "Retired\|superseded" internal/cli/agentlint/ internal/config/ | head -5 || echo "no conflicts"
```

## §D. Constraints (DO NOT VIOLATE)

**PRESERVE list (REQ-ATR-006/007 — assert STILL EXISTS at exit)**:
- `internal/config/types.go` `RoleProfile` struct (Sandbox field) + `SecuritySandbox`
- `internal/config` GitStrategy `Manual/Personal/Team ModeProfile` fields + defaults + `defaults_test.go` AC-GSS-005 block
- `internal/session/store.go` `MergeTeamCheckpoints`, `AppendTaskLedger`, `TaskLedgerEntry`
- `internal/cli/worktree/` entire package (team_launch*, swarm_registry, handoff_guidance)
- `teammateMode` writes in `internal/cli/glm.go` / `launcher.go`
- `f.git_strategy.team.*` + `f.git_strategy.mode.opt.team` i18n keys
- `.claude/rules/moai/core/glm-web-tooling.md`, CG-mode CLI (`moai cg`)

**Forbidden**:
- `--no-verify`, `--amend`, force-push to main
- Bare-substring `team` grep-driven deletion (word/prefix anchors mandatory)
- Hand-editing `fieldsets_templ.go` without regenerating from `.templ`
- Deleting `TestClaimTaskAppend_Repro` or weakening its assertions
- Mirroring `.claude/workflows/*.js` into `internal/template/templates/`
- `Date.now()` / `Math.random()` CALLS in workflow script bodies

**Required**: Conventional Commits (`feat(SPEC-AGENT-TEAM-RETIRE-001): M{N} …`
for run-phase; plan-phase artifact commit uses `feat(` prefix per the
plan-commit-subject lesson), `🗿 MoAI` trailer, specific-path `git add`.

## §E. Self-Verification Deliverables

Per manager-develop-prompt-template §E (E1-E7), each milestone completion report
carries: E1 AC PASS/FAIL matrix (verbatim command output), E2 cross-platform
build results, E3 coverage for touched packages (`internal/lockfile`,
`internal/cli/taskledger` ≥85%), E4 subagent-boundary grep, E5 lint NEW-vs-baseline,
E6 commit SHAs + push state, E7 blocker reports (never AskUserQuestion).

## §F. Milestones

### M0 — Phase 0 de-risk migration (REQ-ATR-001/002/003)

1. Create `internal/lockfile` (exported `Lock(f *os.File) error` /
   `Unlock(f *os.File) error` or equivalent minimal API): move both build-tag
   variants + `team_spawn_lock_unix_test.go`; preserve the Windows limitation
   doc comment verbatim.
2. Create `internal/cli/taskledger`: move `ClaimTask`, `TaskClaimer`,
   `AppendTask`, `TeamTaskEntry` + whatever minimal helpers they transitively
   need; wire to `internal/lockfile`; migrate `TestClaimTask` /
   `TestClaimTaskConcurrent`; re-point `clifix_critical_repro_test.go`.
3. `team_spawn.go` temporarily delegates to the migrated symbols (thin aliases)
   so M0 lands green WITHOUT deleting anything.
4. **MX targets**: `@MX:ANCHOR` on the `internal/lockfile` exported API
   (invariant contract: advisory-lock semantics, fan-in from taskledger +
   future callers; requires `@MX:REASON`); `@MX:ANCHOR` on
   `taskledger.ClaimTask` (carries the SPEC-CLIFIX-CRITICAL-001 P0 fix;
   `@MX:SPEC: SPEC-CLIFIX-CRITICAL-001`); `@MX:NOTE` preserving the Windows
   in-process-mutex limitation explanation.
5. Exit gate (REQ-ATR-003): `go test ./...` exit 0 + windows cross-build exit 0.

### M1 — Phases 1-2: Go removal (REQ-ATR-004/005 + compile-coupled test edits)

1. Delete the five `internal/cli/team_spawn*` files.
2. Remove `TeamConfig` / `RoleProfileEntry` / `TeamAutoSelectionConfig` types,
   `WorkflowConfig.Team` field, `defaults.go` `Team:` block, AND the
   `WorkflowTeamAutoSelection()` accessor in `workflow_accessors.go`
   (non-test source — auditor D3).
3. Compile-coupled test reconciliation lands HERE by necessity:
   `defaults_test.go` TeamConfig rows (GitStrategy Team rows untouched),
   `workflow_nested_test.go` team assertions, `workflow_role_profiles_test.go`
   (delete — its subject type is gone), `types_test.go`
   `TestTeamConfigStructShape` (~:349-429) + accessor tests (~:511-521).
   Also `workflow_lint.go` compile break
   is deferred by stubbing? NO — M1 and M2's agentlint change must land in the
   same commit if `validateRoleProfiles` references `cfg.Team` (it does);
   sequence M1 → immediately M2 step 1, or fold the agentlint edit into M1.
   Decision: fold the `workflow_lint.go` role_profiles-check REMOVAL into M1
   (compile necessity); the model_routing REPLACEMENT is M2.
4. Exit: whole-repo build + test green on both OS targets.

### M2 — Phases 3-4: web console + agentlint repurpose (REQ-ATR-008/009)

1. i18n: remove `f.workflow.team.*` keys (all 4 locale blocks if the file is
   locale-partitioned — verify shape first); re-word `sec.workflow.desc`.
2. `schema_sections.go`: remove `RoleProfileNames()` + the five `workflow.team`
   FieldDef rows (~:212-216: enabled / delegate_mode / max_teammates /
   default_model / require_plan_approval); sweep consumers (`schema.go`,
   `schema_label.go`, `agent_settings_test.go`).
3. `fieldsets.templ`: re-measure for team markup (plan-time: **0 hits** —
   schema-driven UI, auditor D2). Only if the re-measure finds markup: edit
   the `.templ`, regenerate `fieldsets_templ.go`, verify no drift. Otherwise
   NO templ change in this SPEC.
4. agentlint: implement model_routing_profiles closed-set validation reusing
   `ValidateModelRoutingProfiles`; keep the `moai workflow lint` cobra `Use`
   and exit-code contract (non-zero on violation, 0 on clean).
5. Exit: web tests + settings tests green; `moai workflow lint` smoke on the
   local config exits 0.

### M3 — Phases 5-6-7: rules + skills + template mirror (REQ-ATR-010/011/012/013/022)

0. **Migrate-before-delete (REQ-ATR-022, precedes step 3)**: extract the
   essential CG Mode guidance from `.claude/skills/moai/team/glm.md` (LLM mode
   detection, prerequisites, tmux env vars, error recovery — NOT the Agent
   Teams orchestration prose) into a new
   `## CG Mode (Claude + GLM teammates)` section of
   `.claude/rules/moai/core/glm-web-tooling.md`, both trees; verify the
   `CG Mode (Claude + GLM` token lands (AC-ATR-027).
1. Delete `team-protocol.md` + `team-pattern-cookbook.md` (both trees).
2. Shrink `orchestration-mode-selection.md` Mode 3 (catalog row → retirement
   tombstone carrying the literal marker `Mode 3 — RETIRED` — the AC-ATR-015
   tombstone token, baseline 0 at plan-time; remove §C.1 gate detail; keep §G
   axis-warning intact); remove the §B.1 "machine-readable source is
   workflow.yaml auto_selection" pointer sentence — §B.1 prose thresholds
   become sole SSOT (D8 adopted, AC-ATR-028); clean `spec-workflow.md` Agent
   Teams variant sections; sweep dangling refs across (×2 trees each):
   dynamic-workflows.md cross-ref list, CLAUDE.md §15 pointer targets
   (template-managed: mirror edit), `model-policy.md` (~:65), `NOTICE.md`
   (~:23), `spec-frontmatter-schema.md` (~:11), `worktree-integration.md`
   (~:401, ~:405) — auditor D1 additions; PLUS the D6 key-token sweep
   (`workflow.team.enabled`, `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`) over the
   retained surfaces: `settings-management.md`, `agent-authoring.md`,
   `run/{mode-orchestration,context-loading}.md`,
   `workflows/{moai,fix,plan,mx,review}.md` (AC-ATR-029; the Mode 3 tombstone
   in orchestration-mode-selection.md MAY retain the tokens as historical
   context — it is the sole allowed survivor).
3. Delete `.claude/skills/moai/team/` (both trees); remove team routing refs
   from the `workflows/**` tree — top-level `{plan,run,fix,review,mx,moai}.md`
   + `sync.md` loading-table row + `sync/delivery.md` (frontmatter +
   `## Team Mode` section ~:285/:338/:393) +
   `run/{mode-orchestration,task-decomposition,context-loading,phase-execution}.md`
   + `plan/spec-assembly.md` (both trees; CASE-INSENSITIVE re-measure —
   auditor D4). CAUTION: run.md carries the `--mode` dispatch axis (`team`
   value, `MODE_TEAM_UNAVAILABLE` sentinel) and a CI sentinel audit pins
   `MODE_UNKNOWN` literals — reconcile the dispatch table + sentinel docs
   without breaking the sentinel CI audit (adjust the valid-mode set and
   sentinel docs together; the CI audit constraint is a blocker-report topic
   if the sentinel contract is ambiguous at run-time).
4. Remove `team:` block from template + local `workflow.yaml`.
5. `make build`; run neutrality + leak tests; both-tree absence loops
   (per-tree `[ -e ]` independent checks — the SUBCOMMAND-RETIRE-001 D7
   lesson: `grep -q "No such"` false-passes single-tree residue).
6. Exit: template CI guards green; `moai init` sandbox resurrection-negative.

### M4 — Phase 8 final sweep + Scope B workflows (REQ-ATR-014..021)

1. Whole-repo verification batch (tests, cross-build, lint, dangling-ref grep).
2. Author `.claude/workflows/sync-audit-4dim.js` per REQ-ATR-015/016/017 and
   design.md §C (phase structure, schemas, in-script harmonic mean with
   zero-score guard, INCOMPLETE branch, Tier M/L gate note).
3. Author `.claude/workflows/plan-research-fanout.js` per REQ-ATR-018/019/020
   and design.md §D (4-element lens prompts, fixed-heading markdown,
   `### confidence_and_gaps`, ≥2-null abort, no in-workflow writes).
4. `node --check` both scripts; structural greps per acceptance.md;
   optional dry-run smoke (Workflow launch is orchestrator-side, post-merge).
5. Exit: all 29 ACs PASS or documented PASS-WITH-DEBT.

## §G. Anti-Patterns (this SPEC)

- Deleting first, migrating later (Phase 0 inversion) — the CLIFIX repro test
  goes RED and the P0 regression guard is lost mid-flight.
- Bare `team` substring sweeps — deletes preserved GitStrategy/native-runtime
  surfaces (REQ-ATR-007).
- Treating the i18n 288-line plan-time `grep -c "team\."` figure as the removal
  target — it mixes PRESERVE and REMOVE families; re-measure with anchored
  patterns at run-phase (baseline-delta discipline).
- Adding a meta-judge or LLM arithmetic to sync-audit-4dim.js (REQ-ATR-016).
- Shipping the workflows into the template tree (Out of Scope §E).

## §H. Cross-References

- design.md — decision record D1-D9 (boundary, ordering, repurpose-vs-delete,
  script-side arithmetic, markdown-vs-schema, INCOMPLETE semantics, mirror
  strategy, auto-select threshold home [adopted], glm.md migrate-then-delete).
- research.md — verified inventory + external rationale sources.
- acceptance.md — AC matrix (SSOT), GWT scenarios, quality gates, DoD.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Tier L
  Section A-E delegation template applies to every milestone delegation.

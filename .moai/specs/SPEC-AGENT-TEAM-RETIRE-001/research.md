# Research — SPEC-AGENT-TEAM-RETIRE-001

> Verified codebase inventory + external rationale. Every claim below is backed
> by a command executed at plan authoring time (2026-07-11) against the working
> tree; line numbers are content-token-anchored hints and MUST be re-derived at
> run-phase (line-number drift lesson).

## §A. Removal Inventory (Scope A) — verified

### A.1 team_spawn file family (`internal/cli/`)

| File | Size | Disposition |
|------|------|-------------|
| `team_spawn.go` | 15,888 B | DELETE (after M0 migration) |
| `team_spawn_test.go` | 20,379 B | DELETE (ClaimTask tests migrate first) |
| `team_spawn_lock_unix.go` | 377 B | MIGRATE → `internal/lockfile` |
| `team_spawn_lock_windows.go` | 1,425 B | MIGRATE → `internal/lockfile` |
| `team_spawn_lock_unix_test.go` | 2,285 B | MIGRATE → `internal/lockfile` |

Symbol census (`grep -n "^func" team_spawn.go`): ValidateRoleProfile,
ValidateSpawn, ValidateRoster, ValidateMessage, NewMailboxMessage,
ParseMailboxMessage, SerializeMailboxMessage, InitTeamState, AppendTask,
ClaimTask, ArchiveTeamState, LoadRoleProfiles, buildWriteHeavySet,
NewTaskClaimer, TaskClaimer.Claim — plus a package-local `RoleProfile` type
(:56, DISTINCT from `internal/config.RoleProfile`).

**0-non-test-caller verification** (executed):

```
grep -rn -E "InitTeamState|ArchiveTeamState|LoadRoleProfiles|ValidateSpawn|ValidateRoster|MailboxMessage|AppendTask|TaskClaimer|ClaimTask" \
  internal/ cmd/ --include="*.go" | grep -v "_test.go" | grep -v "internal/cli/team_spawn"
→ only internal/session/store.go AppendTaskLedger hits (UNRELATED symbol — substring
  collision on "AppendTask"; session task-ledger, PRESERVE)
```

Test-side dependents: `internal/cli/clifix_critical_repro_test.go:92`
`TestClaimTaskAppend_Repro` (SPEC-CLIFIX-CRITICAL-001 P0 repro — MUST survive
via migration); `team_spawn_test.go` `TestClaimTask` / `TestClaimTaskConcurrent`.

Lock semantics: Unix = `syscall.Flock` LOCK_EX/LOCK_UN; Windows = in-process
per-path `sync.Mutex` map with a documented cross-process limitation comment
("Windows users run solo mode") — preserve verbatim per REQ-ATR-001.

### A.2 Config type family (`internal/config/`)

| Anchor | Content token | Disposition |
|--------|---------------|-------------|
| types.go ~:328 | `Team TeamConfig \`yaml:"team"\`` (WorkflowConfig) | DELETE |
| types.go ~:435 | `type TeamConfig struct` (AutoSelection, Enabled, MaxTeammates, DefaultModel, DelegateMode, RequirePlanApproval, RoleProfileKeys, RoleProfiles) | DELETE |
| types.go ~:447 | `type RoleProfileEntry struct` (Description/Isolation/Mode/Model) | DELETE |
| types.go ~:505 | `type TeamAutoSelectionConfig struct` (MinDomainsForTeam 3 / MinFilesForTeam 10 / MinComplexityScore) | DELETE (nested in TeamConfig; orphaned otherwise) — see design.md D8 threshold-home decision |
| defaults.go ~:457 | `Team: TeamConfig{...}` incl. RoleProfileKeys implementer/tester/reviewer + RoleProfiles map | DELETE |
| types.go ~:474 | `type RoleProfile struct { Sandbox string }` @MX:SPEC SPEC-V3R2-RT-003 | **PRESERVE** |
| types.go ~:153 | GitStrategy `Manual/Personal/Team ModeProfile` | **PRESERVE** |
| defaults.go ~:353 | GitStrategy `Team: ModeProfile{...}` | **PRESERVE** |

NOTE (discovered during verification, extends the brief): `TeamAutoSelectionConfig`
sits OUTSIDE the brief's cited :432-452 range but is TeamConfig-nested and
becomes dead on TeamConfig removal — included in scope (REQ-ATR-005).

### A.3 Web console surfaces

- `internal/web/assets/i18n.js`: **two "team" key families** —
  `f.workflow.team.*` (Agent Teams config → REMOVE) vs `f.git_strategy.team.*`
  + `f.git_strategy.mode.opt.team` (GitStrategy mode profile → PRESERVE).
  Plan-time `grep -c "team\."` = 288 lines is a MIXED figure spanning both
  families plus prose (`sec.workflow.desc`) — never use it as the removal
  target; measure `grep -c '"f\.workflow\.team\.'` at pre-flight.
- `internal/settings/schema_sections.go` :54-56 `RoleProfileNames()` (7-key
  role_profiles catalog) + consumer at :337; further consumers:
  `internal/settings/schema.go`, `internal/web/schema_label.go`,
  `internal/web/agent_settings_test.go`.
- `internal/web/fieldsets.templ` (SOURCE) + `fieldsets_templ.go` (GENERATED) —
  team fieldset removal edits the `.templ` and regenerates.

### A.4 agentlint (`internal/cli/agentlint/workflow_lint.go`)

Current: `validateRoleProfiles` (~:70) asserts
`role_profiles.{implementer,tester,designer}.isolation == "worktree"`; cobra
`Use: "lint"` (~:238). Repurpose target exists:
`internal/config/model_routing.go` `ValidateModelRoutingProfiles` (~:152)
validates `Workflow.ModelRoutingProfiles` against closed sets (perfTier, keys).

### A.5 Rules / skills / config (both trees)

- `.claude/rules/moai/workflow/team-protocol.md` (7,764 B) +
  `team-pattern-cookbook.md` (14,757 B) — present in BOTH trees (verified).
- `.claude/skills/moai/team/{plan,run,sync,debug,review,glm}.md` (6 files) —
  present in BOTH trees (verified).
- Team routing references: `workflows/{plan,run,fix,review,mx}.md` matched
  `team` at plan-time; **`workflows/sync.md` did NOT match** — Phase 6 scope is
  5 files unless run-phase re-measure differs.
- `internal/template/templates/.moai/config/sections/workflow.yaml` `team:`
  block (~:30, auto_selection thresholds); local
  `.moai/config/sections/workflow.yaml` `role_profiles:` (~:101).
- Rules cross-referencing team files (dangling-ref sweep targets):
  `dynamic-workflows.md` § Cross-references cites `team-protocol.md` +
  `team-pattern-cookbook.md`; `spec-workflow.md` carries the Agent Teams
  variant sections + `.claude/skills/moai/team/*` pointers; CLAUDE.md §15 +
  §4 (template-managed — mirror edits).

### A.6 Preservation census (Phase 9 — verified present)

- `internal/session/store.go` `MergeTeamCheckpoints` (tested at
  store_test.go:568), `AppendTaskLedger`/`TaskLedgerEntry` (:47-48, :355-356).
- `internal/cli/worktree/`: `team_launch_windows.go` (+ POSIX variant),
  `swarm_registry.go`, `handoff_guidance.go`, `new.go` — the `--team` P1-P4
  launch patterns (moai-workflow-worktree skill §5).
- `teammateMode` writers: `internal/cli/glm.go`
  (`ensureSettingsLocalJSON`/`injectGLMEnvForTeam` → `"tmux"`),
  `internal/cli/launcher.go` (`removeGLMEnv` → `""`) (CLAUDE.local.md §22.3).
- `internal/config/defaults_test.go` :249-298 GitStrategy Team ModeProfile
  assertions (AC-GSS-005) — PRESERVE; only TeamConfig rows are removal scope.

### A.7 Tests inventory (Phase 8)

- `internal/config/workflow_role_profiles_test.go` (5,926 B — subject type
  deleted → file deleted).
- `internal/config/workflow_nested_test.go` (6,961 B — team blocks only).
- `internal/config/defaults_test.go` — TeamConfig rows only.
- `internal/web/agent_settings_test.go` — role_profiles references.

## §B. Scope B Groundwork

- House-style exemplar: `.claude/workflows/codemaps-extract.js` (3,629 B)
  read in full — header doctrine comment (verdict scoping / determinism /
  read-only), `export const meta {name, description, phases}`, args-with-defaults
  (`(args && args.packages) || [...]`), `phase('...')` markers,
  `agent(prompt, {label: 'extract:${pkg}', phase, agentType: 'Explore',
  effort: 'low'})`, `parallel(...)`, `.filter(...)` null-filtering, no
  wall-clock/random calls, plain object return.
- Existing `.claude/workflows/` residents: `codemaps-extract.js`,
  `harness-release-update-run.js` (dev-only §21). The directory is user-owned,
  NOT template-managed (dynamic-workflows.md § MoAI Integration Notes) — Scope B
  scripts are local-only.
- Purpose-driven effort mapping (dynamic-workflows.md § Purpose-driven
  model+effort selection): read-only-extract → low/medium; synthesize → high;
  verify-judge → xhigh. sync-audit judges = verify-judge (xhigh); context
  extraction = read-only (medium — richer than codemaps' low because it parses
  acceptance.md semantics); plan-research explorers = read-only research
  (medium); synthesizer = synthesize (high).

## §C. Risks

| Risk | Severity | Mitigation |
|------|----------|------------|
| Substring-collision deletion ("team" token) | HIGH | Anchored greps only; REQ-ATR-007 preservation AC; per-family i18n counts |
| P0 repro test lost in migration gap | HIGH | M0-before-M1 gate (REQ-ATR-003); repro re-point verified by AC-ATR-003 vacuous-green guard |
| Windows build break (build-tag pair) | MED | AC-ATR-002 cross-build; B1 known-issue injection |
| `moai cg` doc routing loss (team/glm.md) | MED | [NEEDS CLARIFICATION] in plan.md §A — resolve pre-kickoff |
| Auto-select threshold SSOT dangling | LOW | design.md D8 + [NEEDS CLARIFICATION] in plan.md §A |
| templ regen version drift | MED | Edge case E6; pin regen tool; diff-scope review |
| Template CI guard trip (neutrality) | LOW | M3 runs guards locally before push |
| Workflow scripts drift from house style | LOW | AC-ATR-026 structural greps; codemaps exemplar |

## §D. External Rationale Sources (Scope B grounding)

1. **Anthropic engineering — "How we built our multi-agent research system"**
   (multi-agent research post; cited via `dynamic-workflows.md` § Routing
   Heuristic and `orchestration-mode-selection.md` §F "Anthropic multi-agent
   research engineering note"): multi-agent systems spend ~**15x more tokens**
   than single-agent chat — fan-out must be reserved for genuinely parallel
   high-value work; **coding is a poor fit for parallelism** ("most coding
   tasks involve fewer truly parallelizable tasks than research"); effective
   subagent prompts carry **4 elements** — objective, output format, tool
   guidance, task boundaries. → plan-research lens prompts (REQ-ATR-018),
   ≤4-lens cap, explorer effort medium (not xhigh).
2. **Anthropic engineering — "Building effective agents"**: the
   **sectioning** pattern (parallel subtasks with distinct responsibilities)
   grounds the 4-lens fan-out; the **voting** pattern (multiple independent
   perspectives on the same artifact) grounds the 4 independent judges — and
   its corollary that aggregation should be mechanical (script arithmetic),
   not another model opinion. → design.md D4 (no meta-judge).
3. **Claude Code workflows documentation** (`code.claude.com/docs/en/workflows`
   — URL verified in-repo via `dynamic-workflows.md` header citation): up to
   **16 concurrent agents** (1,000/run backstop); **resume caching** keys on
   deterministic script output (no wall-clock/random calls in body);
   **schema-forced output** available per agent call; workflow agents cannot
   prompt the user; `.claude/workflows/` saved-command mechanics. →
   REQ-ATR-021 determinism, judge schemas, read-only boundaries.
4. **In-repo prior art**: `codemaps-extract.js` falsification verdict
   ("augmentation, not extraction"; high-count justification only) — the
   Scope B scripts inherit its honesty conventions ("NONE found" valid;
   verdict-scoping header).

Note: sources 1-2 are cited by title; their canonical URLs were not re-fetched
in this session (WebFetch not exercised) — run-phase or sync-phase may pin
exact URLs. Source 3's URL appears verbatim in the repo rule file.

## §E. Verification Method Statement

All §A/§B claims were produced by direct commands (ls, grep, sed, file reads)
in this session — none are recalled from memory or assumed from the brief.
Deltas found against the brief: (1) i18n key family split (`f.workflow.team.*`
vs `f.git_strategy.team.*`) — brief's "~264 team.* keys" is a mixed count;
(2) `TeamAutoSelectionConfig` extends the brief's types.go range; (3) i18n
lives at `internal/web/assets/i18n.js` (brief implied web console pkg);
(4) `workflows/sync.md` carries no team reference (brief listed it);
(5) team_launch lives at `internal/cli/worktree/` (brief cited
`worktree/team_launch*.go`); (6) `internal/lockfile` and
`internal/cli/taskledger` do not exist yet — Phase 0 creates them.

# SPEC-AGENT-PROGRESS-PUSH-001 — Progress

## §E.1 Plan-phase Audit-Ready Signal

- **Tier**: L (37 REQ / 48 AC; ~28 files across two trees; 5 new + 2 amended constitution clauses; zero Go production logic)
- **Artifacts**: spec.md · plan.md · acceptance.md · research.md · design.md · progress.md (v0.2.0)
- **Open clarifications**: **0**. Both v0.1.0 markers resolved — roadmap trigger = any agent declaring `N` of 3 or more milestones (no new threshold); orchestrator continues independent READ-ONLY work while a background agent runs (read-only is HARD).
- **M0 gate**: **deleted**. The v0.1.0 blocking canary asked whether `SendMessage({to: "main"})` works from a foreground subagent; it was answered empirically (yes, on Claude Code v2.1.206). Replaced by a standing regression check (REQ-APP-037, M5).
- **Design change**: single channel → **dual channel**. Primary = `TaskCreate` / `TaskUpdate` / `TaskList` (officially documented). Secondary = `SendMessage({to: "main"})` (undocumented runtime behavior, best-effort). Rationale: research.md §B.2.
- **New scope**: background-default realignment. The runtime backgrounds subagents by default as of v2.1.198 (running: v2.1.206); no MoAI doctrine surface is aware of this. Four surfaces drifted; two constitution clauses amended in place.
- **Decision surfaced for review** (not a blocker): the two amended clauses carry inline `[ZONE:Frozen]` markers while the zone registry — the declared SSOT — records both as `zone: Evolvable`. Pre-existing contradiction. The amendment proceeds on the registry-SSOT ground **and** on the content ground (it replaces a backgrounding-based safeguard with a better-targeted concurrency-based one; it does not relax safety). See plan.md §C.1 and research.md §D.3.
- **Residual open measurement**: foreground push *delivery timing* — the probe showed the call succeeds, not when the user sees it. AC-APP-037b requires run-phase to measure, not infer.
- **Status**: draft — awaiting plan-audit.

## §E.2 Run-phase Evidence

### Implementation

Implemented by manager-develop (worktree commit `32cfaa7a0`), integrated into `main`
by orchestrator cherry-pick `2c086db6d` (stale-worktree cherry-pick pattern; target
files verified byte-identical to the shared-checkout baseline before integration).
`make build` re-run fresh in the shared checkout: catalog.yaml regeneration produced
no diff (the worktree embed was already correct). 29 files: 9 agents x 2 trees, SSOT
`progress-reporting-protocol.md` + template mirror, 5 doctrine-realignment surfaces,
zone-registry (live-only), rule_template_mirror_test.go allowlist, catalog.yaml.

### AC-APP-037b — dual-channel delivery, empirically observed (measured, not inferred)

The delivery observation is orchestrator-owned: only the main conversation can
adjudicate whether a `SendMessage({to:"main"})` push actually surfaced. Observed
this session on Claude Code runtime **v2.1.207** (`CLAUDE_CODE_EXECPATH` version dir):

- **Background subagent (`run_in_background: true`)**: a probe pushed 3 milestone
  messages via `SendMessage({to:"main"})`. Each call returned
  `{"success":true,"message":"Message queued for the main conversation's next turn."}`.
  All 3 surfaced in the main conversation as `<agent-message>` blocks. Task-notification
  final result confirmed `SendMessage result: SUCCESS (all 3 delivered)`.
- **Foreground subagent (`run_in_background: false`)**: a probe pushed 1 message via
  `SendMessage({to:"main"})`. Return `{"success":true,...}`; the `[FOREGROUND PROBE]`
  message surfaced in the main conversation. So the runtime schema's "background
  subagents only" note is NOT enforced at the tool-call layer for foreground either.
- **Delivery timing (the genuine open measurement, research.md §D.4 point 3)**: messages
  are "queued for the main conversation's next turn" and empirically surfaced at the
  orchestrator's next tool-call boundary while it stayed engaged. Corollary recorded in
  the SSOT: the orchestrator MUST NOT idle immediately after a background spawn, or the
  push drain is deferred until the next orchestrator turn.
- **Undocumented-channel caveat (verified)**: `to:"main"` is NOT in the official docs
  (tools-reference documents only agent-team teammate / agent-ID-or-name recipients);
  it exists only in the runtime tool schema. Hence the dual-channel design: the
  documented `Task*` channel is load-bearing; `SendMessage` is a best-effort immediacy
  layer that may break on a runtime update, at which point progress degrades from
  immediate to pull-visible rather than to an outage.

### Verification batch (shared checkout, integrated tree)

| Check | Result |
|-------|--------|
| `make build` | exit 0; catalog.yaml regen no-diff |
| `go test ./internal/template/...` | exit 0 (mirror-drift / neutrality / embed guards) |
| SendMessage in tools, 9 agents x 2 trees | 9/9 and 9/9 |
| Task* added to manager-design/plan-auditor/super-advisor/sync-auditor x 2 trees | all 1 |
| AC-APP-026 `MUST NOT perform Write/Edit` (live+tmpl) | 0, 0 |
| AC-APP-026b `run_in_background: false...conservative default` (live+tmpl) | 0, 0 |
| AC-APP-009a `AskUserQuestion(` in agent bodies | 0 (boundary preserved) |
| AC-APP-029b `moai constitution guard CONST-V3R2-020,044` | exit 0, no Frozen violation |
| AC-APP-029b `moai constitution validate` | 77 errors (== pre-existing baseline, no regression) |

## §E.3 Run-phase Audit-Ready Signal

- run_status: landed-on-main
- run_commit_sha: `3d4addc2f` (feat — dual-channel agent progress reporting + background-default realignment; integrated from worktree `32cfaa7a0` via orchestrator cherry-pick `2c086db6d`, then rebased onto the restored main as `3d4addc2f`; target files verified byte-identical to the shared-checkout baseline before integration per `feedback_stale_worktree_orchestrator_cherry_pick`)
- run_artifact_commit_sha: `af2484d8a` (docs — plan+run artifacts, draft→in-progress)
- ac_reverification_on_main: orchestrator-independently re-verified on the main integration tree — SendMessage in `tools:` 9/9 × 2 trees, `Task*` added to 4 agents (`manager-design`/`plan-auditor`/`super-advisor`/`sync-auditor`) × 2 trees, AC-APP-026/026b background-write-restriction literal 0/0 (live+tmpl), AC-APP-009a `AskUserQuestion(` in agent bodies 0 (boundary preserved), AC-APP-029b `moai constitution guard` exit 0, `make build` exit 0 (catalog.yaml regen no-diff), `go test ./internal/template/...` exit 0
- runtime_measured: Claude Code **v2.1.207** — `SendMessage({to:"main"})` delivery observed from BOTH background and foreground subagents (AC-APP-037b measured, not inferred); the runtime schema's "background subagents only" annotation is NOT enforced at the tool-call layer

## §E.4 Sync-phase Audit-Ready Signal

- sync_status: audit-ready
- sync_complete_at: 2026-07-13
- sync_commit_sha: `56ba822bc` (docs — sync-phase artifacts — 3-phase close; the real SHA, backfilled in this follow-up commit per the D3 self-referential-hazard exemption — a commit cannot reference its own hash)
- changelog_entry_added: true (standalone AGENT-PROGRESS-PUSH-001 bullet under `### Added`, immediately after the EVOLVE-003 bullet)
- frontmatter_transition: in-progress → completed (single sync commit carries the merged in-progress → implemented → completed transition per the 3-phase close convention; spec.md is the ONLY artifact with a `status:` frontmatter field — acceptance.md/plan.md/progress.md carry no frontmatter block in this SPEC, so the transition is a one-file edit)
- readme_changed: false (dual-channel progress-reporting user-facing narrative deferred to the pending v3.0 README redesign — `project_readme_v3_rc11_redesign_draft` memory; CHANGELOG carries user visibility in the meantime)
- docs_site_changed: false (same rationale; `agent-teams.md` SendMessage references are team-mode context, unchanged)
- ac_reverification_at_sync: orchestrator independently re-ran the §E.2 verification batch on the main integration tree BEFORE the sync commit (SendMessage 9/9 × 2 trees, `Task*` × 4 agents × 2 trees, AC-APP-026/026b/009a/029b, `moai spec audit` MUST-FIX 0)
- llm_yaml_excluded: the working-tree `team_mode: "" → glm` modification in `.moai/config/sections/llm.yaml` is GLM-backend runtime state (CLAUDE.local.md §22.3), NOT a SPEC artifact — excluded from the sync commit per scope discipline (`feedback_manager_docs_sync_phase_discipline`: no `git add -A` kitchen-sink)
- manager_docs_delegation: orchestrator-direct (GLM backend + EVOLVE-003 manager-docs autocompact-thrash precedent; sync scope is mechanical CHANGELOG + frontmatter + §E.3/§E.4 — orchestrator-allowed direct-execution territory per CLAUDE.md §4 "git operations / result synthesis")

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

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

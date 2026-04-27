---
name: moai-harness-learner
description: Harness learning subsystem coordinator. Surfaces Tier 4 auto-update proposals via AskUserQuestion and orchestrates Apply/Rollback flows. Triggers when harness learning proposals are pending or learning lifecycle management is needed.
triggers:
  - keyword: "harness apply"
  - keyword: "harness proposal"
  - keyword: "harness rollback"
  - keyword: "harness status"
  - keyword: "harness disable"
  - keyword: "learning proposal"
  - keyword: "tier 4"
  - keyword: "auto update proposal"
allowed-tools: Bash,Read,Write,Edit,AskUserQuestion
---

# moai-harness-learner

Coordinator skill for the Harness Learning Subsystem (SPEC-V3R3-HARNESS-LEARNING-001).
Surfaces Tier 4 auto-update proposals to the user via AskUserQuestion and orchestrates Apply/Rollback flows.

## Quick Reference

**Role**: Orchestrator-side bridge between CLI (`moai harness`) and AskUserQuestion.

**Key constraint** [HARD]: `moai harness apply` returns a JSON payload. This skill MUST receive that payload and surface it via `AskUserQuestion`. The CLI itself does NOT prompt the user.

**Common triggers**:
- `moai harness status` — check tier distribution and pending proposals
- `moai harness apply` — load next pending proposal (returns JSON payload)
- `moai harness rollback <date>` — restore snapshot
- `moai harness disable` — set learning.enabled: false

**Workflow**:
1. Run `moai harness status` to inspect state.
2. Run `moai harness apply` to get the proposal payload.
3. Surface payload via `AskUserQuestion` (approve / reject).
4. On approve: write approval to proposals dir and signal CLI to proceed.
5. On reject: remove proposal file (no changes applied).

---

## Implementation Guide

### Step 1: Status Check

```bash
moai harness status --project-root <project_root>
```

Output includes:
- `enabled` state
- Tier distribution (observation / heuristic / rule / auto_update)
- Rate limit window status
- Number of pending proposals

### Step 2: Fetch Proposal Payload

```bash
moai harness apply --project-root <project_root>
```

The command outputs a JSON block with:
- `id` — proposal identifier
- `target_path` — file to be modified
- `field_key` — `description` or `triggers`
- `new_value` — proposed new content
- `pattern_key` — what triggered this proposal
- `observation_count` — how many times this pattern was observed

### Step 3: Surface via AskUserQuestion

[HARD] This skill (not the CLI) calls AskUserQuestion. The CLI only provides the payload.

```
AskUserQuestion:
  question: "Harness 학습 자동 업데이트 제안 (proposal_id: <id>)\n\n대상: <target_path>\n필드: <field_key>\n새 값: <new_value>\n관찰 횟수: <observation_count>\n\n이 변경을 적용하시겠습니까?"
  options:
    - label: "승인 (권장)"
      description: "제안된 변경을 skill 파일에 적용합니다. 스냅샷이 먼저 생성됩니다."
      value: "approve"
      recommended: true
    - label: "거부"
      description: "이 제안을 건너뜁니다. proposal 파일이 삭제됩니다."
      value: "reject"
    - label: "자세히 보기"
      description: "대상 파일의 현재 내용을 확인한 후 결정합니다."
      value: "inspect"
    - label: "일시 정지"
      description: "지금은 결정하지 않습니다. proposal 파일은 유지됩니다."
      value: "defer"
```

### Step 4: On Approve

The skill applies the change by invoking the safety pipeline directly. Since the CLI `apply` only surfaces the payload (not executes), the actual write happens via the harness package's `Apply()` function, gated by the 5-Layer Safety Pipeline.

For the coordinator skill, the simplest flow is:
1. User selects "approve"
2. Write `approved: true` to `.moai/harness/proposals/<id>.decision`
3. Run `moai harness apply --execute` (if the CLI supports it) or call the harness API directly.

### Step 5: On Reject

1. Delete `.moai/harness/proposals/<id>.json`
2. Confirm deletion to user.

### Rollback Flow

```bash
# List available snapshots
ls .moai/harness/learning-history/snapshots/

# Rollback to a specific snapshot
moai harness rollback 2026-04-27T00-00-00.000000000Z --project-root <project_root>
```

### Disable Learning

```bash
moai harness disable --project-root <project_root>
```

Sets `learning.enabled: false` in `.moai/config/sections/harness.yaml`.
Comments and key ordering are preserved (YAML round-trip).

---

## Works Well With

- `moai-meta-harness` — generates the `my-harness-*` skills that are targets of auto-updates
- `moai-workflow-tdd` — TDD cycle generates events that feed into the observer
- `moai-foundation-quality` — quality gates run after auto-updates to validate correctness

## Safety Architecture Reference

The 5-Layer Safety Pipeline protects every auto-update:

| Layer | Guard | Action on violation |
|-------|-------|---------------------|
| L1 | Frozen Guard | Block — FROZEN paths are never modified |
| L2 | Canary Check | Block — if effectiveness drops >0.10 |
| L3 | Contradiction Detector | Block — if trigger conflicts arise |
| L4 | Rate Limiter | Block — max 3 per week, 24h cooldown |
| L5 | Human Oversight | Surface via AskUserQuestion (this skill) |

[HARD] L1 Frozen paths (never auto-modified at runtime):
- `.claude/agents/moai/**`
- `.claude/skills/moai-*/**`
- `.claude/rules/moai/**`
- `.moai/project/brand/**`

Only user-area skills (`.claude/skills/my-harness-*/`) are valid auto-update targets.

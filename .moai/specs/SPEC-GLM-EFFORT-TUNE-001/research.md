---
id: SPEC-GLM-EFFORT-TUNE-001
title: "GLM effort overlay configuration tune-up (P1/P2/P4) — source evidence"
version: "0.1.0"
status: completed
created: 2026-07-14
updated: 2026-07-14
author: manager-spec
priority: P2
phase: "v3.x config-tune"
module: "internal/template/glm_effort_overlay.go + .moai/config/sections/llm.yaml"
lifecycle: spec-anchored
tags: "glm, effort, overlay, config, template-mirror, reasoning-effort"
related_specs: [SPEC-MODEL-TIER-PLANTYPE-001]
---

# SPEC-GLM-EFFORT-TUNE-001 — research.md

This file cites line numbers from the actual source files read in this session. It does NOT paraphrase memory-summary prose (per `feedback_defect_claim_verification`). Every claim below is anchored to a verifiable line range.

## §A. The GLM effort overlay — implementation truth

### §A.1 The 3 canonical reasoning states

**File**: `internal/template/glm_effort_overlay.go` · **Lines 26-33**

```go
const (
    GLMStateThinkingOff = "thinking-off"        // line 28
    GLMStateReasoningHigh = "reasoning-high"    // line 30
    GLMStateReasoningMax = "reasoning-max"      // line 32
)
```

This is the **3-state reality**. The user proposal's "high/max" axis describes the two REASONING-ON tiers; the third reachable state (`thinking-off`) is the mechanical tier below them. P4 is a framing correction TOWARD this SSOT — the code constants are already correct.

### §A.2 The collapse function (5 → 3)

**File**: `internal/template/glm_effort_overlay.go` · **Lines 88-100** (`CollapseClaudeEffortToGLM`)

```go
switch effort {
case EffortLevelLow:
    return glmReasoningThinkingOff
case EffortLevelMedium, EffortLevelHigh:
    return glmReasoningHigh
case EffortLevelXHigh, EffortLevelMax:
    return glmReasoningMax
default:
    return glmReasoningMax   // totality clause — never under-reason
}
```

**Implication for P1**: removing `builder-harness` from the override set causes `ResolveGLMReasoning("builder-harness", "high")` to call `CollapseClaudeEffortToGLM("high")` which returns `glmReasoningHigh` (`reasoning-high`). This is the make-or-break behavioral claim of P1 (AC-GET-003).

### §A.3 The coding-max override set (the P1 target)

**File**: `internal/template/glm_effort_overlay.go` · **Lines 102-110**

```go
// glmCodingMaxOverrideAgents is the coding-max override set (REQ-MTP-028): the two
// code-producing retained agents that z.ai recommends reasoning_effort: max for
// coding tasks. manager-develop runs run-phase implementation; builder-harness
// generates dynamic specialists. No other retained agent is a code-producer, so
// no other agent is overridden. Named constant collection per §14.
var glmCodingMaxOverrideAgents = map[string]bool{
    "manager-develop": true,    // line 108
    "builder-harness": true,    // line 109  ← P1 removes THIS line
}
```

**The line 109 removal is the entire P1 production-code change.** The comment at lines 102-106 ("the two code-producing retained agents") must also be updated in the same edit (the set becomes one member).

### §A.4 The override application point

**File**: `internal/template/glm_effort_overlay.go` · **Lines 135-142** (`ResolveGLMReasoning`)

```go
func ResolveGLMReasoning(agentName, claudeEffort string) GLMReasoningState {
    if IsGLMCodingMaxOverrideAgent(agentName) {
        return glmReasoningMax
    }
    return CollapseClaudeEffortToGLM(claudeEffort)
}
```

This function is unchanged by P1 — it continues to consult the (now smaller) override set. The behavioral change comes from the set's contents, not the function's logic.

### §A.5 The session-global delivery (NOT affected by P1)

**File**: `internal/template/glm_effort_overlay.go` · **Lines 172-174** (`SessionGLMReasoningState`)

```go
func SessionGLMReasoningState() GLMReasoningState {
    return glmReasoningMax
}
```

This function unconditionally returns `glmReasoningMax`. It does NOT consult the override set. It is the session-global default for sub-agents and the empty-effort fallback. **P1 does NOT change this function** — the session-global default stays at `reasoning-max`. This is why the comment-only references in `internal/cli/glm.go:242` and `internal/cli/launcher.go:820` ("hardcoded coding-max default") are NOT affected by P1 (KI-2): they describe `SessionGLMReasoningState()`, not the override set.

## §B. The tierProfiles matrix — `builder-harness` is `high`, NOT `xhigh`

### §B.1 API plan_type profile

**File**: `internal/template/model_policy.go` · **Lines 295-308** (api rows)

```go
config.PlanTypeAPI: {
    "manager-spec":    {{"fable", "high"}, {"fable", "high"}, {"opus", "high"}},
    ...
    "manager-develop": {{"fable", "high"}, {"opus", "high"}, {"opus", "medium"}},
    "builder-harness": {{"opus", "high"}, {"opus", "medium"}, {"opus", "medium"}},  // line 303
    "e2e-specialist":  {{"opus", "high"}, {"opus", "medium"}, {"opus", "medium"}},
    ...
    "manager-git":     {{"sonnet", "low"}, {"sonnet", "low"}, {"sonnet", "low"}},   // line 306
    ...
}
```

### §B.2 Subscription plan_type profile

**File**: `internal/template/model_policy.go` · **Lines 309-322** (subscription rows)

```go
config.PlanTypeSubscription: {
    ...
    "manager-develop": {{"sonnet", "high"}, {"sonnet", "high"}, {"sonnet", "high"}},
    "builder-harness": {{"sonnet", "high"}, {"sonnet", "medium"}, {"sonnet", "medium"}},  // line 317
    ...
    "manager-git":     {{"sonnet", "low"}, {"sonnet", "low"}, {"sonnet", "low"}},         // line 320
    ...
}
```

**Implication for P1**: `builder-harness` effort is `high` (or `medium` at lower tiers) across BOTH plan_type profiles — NEVER `xhigh` or `max`. So removing the override places harness at `CollapseClaudeEffortToGLM("high") = reasoning-high` (at the high-effort tiers) or `CollapseClaudeEffortToGLM("medium") = reasoning-high` (at medium tiers, which converge). It NEVER falls below `reasoning-high` under P1.

**Implication for P4**: `manager-git` effort is `low` across all tiers → `CollapseClaudeEffortToGLM("low") = thinking-off`. This is the "mechanical tier" occupant named in P4's framing.

## §C. The agent-authoring Effort-Level Calibration Matrix

**File**: `.claude/rules/moai/development/agent-authoring.md` · **Lines 340-351**

| Agent | Default effort | Rationale (verbatim) |
|-------|----------------|----------------------|
| `manager-develop` | xhigh | run-phase implementation (coding/agentic) |
| `manager-git` | low | git operations, PR creation, Tier-L routing (fast bash execution; was high pre-No-Haiku) |
| `builder-harness` | **high** | **artifact scaffolding (agents/skills/plugins/hooks)** |
| `Explore` | medium | read-only codebase exploration |

**Implication for P1 design.md §A**: the rationale text "artifact scaffolding (agents/skills/plugins/hooks)" is the load-bearing role-classification evidence. `manager-develop`'s rationale "coding/agentic" is materially different — the two agents do not occupy the same role-class even though both touch code.

**Implication for P4**: `manager-git` rationale "fast bash execution" — this is the "mechanical work" classification that occupies the thinking-off tier.

## §D. The test that P1 must update

**File**: `internal/template/glm_effort_overlay_test.go` · **Lines 114-135**

```go
// TestGLMCodingMaxOverrideAgents_ExactlyTwo asserts the override set is EXACTLY
// {manager-develop, builder-harness} — no third member (REQ-MTP-028 / AC-MTP-030).
func TestGLMCodingMaxOverrideAgents_ExactlyTwo(t *testing.T) {
    got := GLMCodingMaxOverrideAgents()
    ...
    want := []string{"builder-harness", "manager-develop"}  // line 119
    if len(got) != 2 {                                       // line 121
        t.Fatalf("GLMCodingMaxOverrideAgents() has %d members %v, want exactly 2 %v", ...)
    }
    ...
    if !IsGLMCodingMaxOverrideAgent("manager-develop") || !IsGLMCodingMaxOverrideAgent("builder-harness") {
        t.Error("IsGLMCodingMaxOverrideAgent must be true for both override agents")
    }
    if IsGLMCodingMaxOverrideAgent("manager-spec") || IsGLMCodingMaxOverrideAgent("sync-auditor") {
        t.Error("IsGLMCodingMaxOverrideAgent must be false for non-override agents")
    }
}
```

**The M1 edit**: rename to `TestGLMCodingMaxOverrideAgents_ExactlyOne` (or equivalent); change `want` to `[]string{"manager-develop"}`; change the cardinality assertion to `len(got) == 1`; remove the `IsGLMCodingMaxOverrideAgent("builder-harness")` true-branch; ADD a behavioral assertion `ResolveGLMReasoning("builder-harness", "high").Name == "reasoning-high"` (AC-GET-003).

The earlier test at line 85 (`// override (AC-MTP-030): manager-develop and builder-harness resolve to ...`) also references both agents — M1 must verify whether that test asserts the override applies to BOTH or just enumerates them, and update accordingly.

## §E. The llm.yaml surfaces (P2 targets)

### §E.1 Local config

**File**: `.moai/config/sections/llm.yaml` · **Lines 1-24** (full file)

The local file exposes `glm.base_url` and `glm.models` (tier→GLM model map). It has NO reasoning_effort surface — confirmed by `grep -c 'reasoning_effort' = 0` (precondition 4 in plan.md §C).

### §E.2 Template mirror

**File**: `internal/template/templates/.moai/config/sections/llm.yaml` · **Lines 1-52** (full file)

The template mirror has the same `glm.base_url` + `glm.models` shape, plus a richer `context_windows` comment block (lines 36-50) that the local file lacks in full form. P2 adds the reasoning-effort exposure block to BOTH — Template-First (CLAUDE.local.md §2).

### §E.3 Mirror-parity test coverage (KI-4)

To be verified in M2: whether `internal/template/rule_template_mirror_test.go` (the byte-parity invariant for SSOT mirrors) actually covers `.moai/config/sections/llm.yaml`. If it does, both files must be edited to keep parity; if it does not, the edit is still made to both for consistency, but the CI guard does not bind. M2 records the test's coverage decision in this section.

## §F. Production callers of the override set (P1 blast radius)

**Command**: `grep -rn 'glmCodingMaxOverrideAgents\|GLMCodingMaxOverrideAgents\|IsGLMCodingMaxOverrideAgent' internal/ --include='*.go'` (run this session)

**Result**:
- `internal/template/glm_effort_overlay.go` — definitions + self-use (lines 102-135)
- `internal/template/glm_effort_overlay_test.go` — test (lines 85, 114-135)
- **NO other production caller.**

The comment-only references at `internal/cli/glm.go:242` and `internal/cli/launcher.go:820` use the phrase "hardcoded coding-max default" to describe `SessionGLMReasoningState()` — they do NOT reference the override set or its members. P1's blast radius is therefore confined to:
1. `internal/template/glm_effort_overlay.go` (the map definition + 3 doc comments)
2. `internal/template/glm_effort_overlay_test.go` (the test rename + assertion update + new behavioral assertion)

No CLI caller, no launcher caller, no other package imports the override set's membership.

## §G. z.ai reasoning_effort fact (memory-anchored, code-confirmed)

**Memory source**: `reference_glm_reasoning_effort_2026_07.md` — "GLM은 effort 미지원, thinking 토글 + reasoning_effort {high, max} 2단계".

**Code confirmation**: `glm_effort_overlay.go:35-51` defines the wire field constants:

```go
GLMReasoningEffortKey = "reasoning_effort"   // line 40
GLMThinkingKey = "thinking"                   // line 42
GLMReasoningEffortHigh = "high"               // line 44
GLMReasoningEffortMax = "max"                 // line 46
GLMThinkingEnabledVal = "enabled"             // line 49
GLMThinkingDisabledVal = "disabled"           // line 50
```

The memory summary "reasoning_effort {high, max} 2단계" is correct as far as the wire values go, but it is INCOMPLETE as a description of the reachable states — the thinking toggle (enabled/disabled) adds a third state (thinking-off) that the memory summary compresses away. P4's framing correction restores the 3-state model that the code constants already encode.

This is a worked example of why `feedback_defect_claim_verification` requires line-number citation rather than memory-summary prose: the memory summary was not wrong, but it was lossy in a way that hid the 3rd state.

## §H. Honesty caveat provenance (constraint C-3)

The honesty caveat ("implemented + wired; live validation pending") originates in SPEC-MODEL-TIER-PLANTYPE-001, carried forward via memory `project_model_tier_plantype_001_completed.md` and `project_model_tier_plantype_001_followups.md`. The followup list explicitly tracks "GLM wire 실증" (live validation) as backlog item #1 — that is the P3 this SPEC declares Out-of-Scope (spec.md §E).

The overlay's wire-effectiveness is therefore:
- **IMPLEMENTED**: the collapse + override + session-state logic exists and is unit-tested (`glm_effort_overlay_test.go`).
- **WIRED**: the overlay's output reaches the z.ai request path via `SessionGLMReasoningStateForEffort` → env-var / settings.local.json block (`internal/cli/glm.go`, `internal/cli/launcher.go`).
- **NOT LIVE-VALIDATED**: no end-to-end test exists that sends a request to z.ai and confirms the resolved reasoning_effort produces the expected model behavior. That validation is P3.

P2's exposure block wording MUST reflect this triple-state exactly. Words like "validated" or "guaranteed" violate the caveat (AC-GET-010).

## §I. Cross-references

- `verification-claim-integrity.md` §1.1 surface 3 — the policy basis for the honesty caveat
- `feedback_defect_claim_verification` — the methodological basis for this file's line-number discipline
- `feedback_ac_token_presence_not_reachability` — relevant to AC design: the P1 ACs verify BEHAVIORAL outcomes (`ResolveGLMReasoning(...).Name`), not just token presence in the map
- SPEC-MODEL-TIER-PLANTYPE-001 — the parent overlay SPEC
- design.md §A (P1 exclude-vs-include) · §B (P2 struct-field-vs-comments-only)

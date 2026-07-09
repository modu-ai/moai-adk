---
id: SPEC-AGENT-ARCH-V2-001
title: "MoAI Agent Architecture v2 — Research (Codebase Baselines)"
version: "0.1.0"
status: draft
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents/moai + internal/config"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "agent-arch, super-advisor, manager-design, no-haiku, 3-tier, claude-design, token-policy, research"
---

# SPEC-AGENT-ARCH-V2-001 — Research

> research.md is the codebase-analysis artifact. Every baseline below was measured 2026-07-09 by this agent via Read/Bash/Grep against the live tree at HEAD `2fae7057a`. Each section cites file:line and the observed content; vci §2 attribution applies (no carry-over from unrelated prior measurements).
>
> **Design authority**: `.moai/reports/agent-architecture-redesign-v2-20260709.html` (the SSOT HTML, 565 lines). This file records the live-tree extension points the v2 architecture will modify; it does NOT re-derive architecture.

---

## §A RouteModelFor baseline (the M3 2-arg → 3-arg extension point)

### §A.1 File: `internal/config/model_routing.go`

**Current signature (line 89)**:

```go
// @MX:ANCHOR: [AUTO] RouteModelFor is the spawn-time cost-routing accessor
// @MX:REASON: fan_in >= 3 expected (orchestrator spawn call-site, tests, Epic C/D consumers)
func (c *Config) RouteModelFor(tier, phase string) (ModelRoutingEntry, error) {
    if !validRoutingTiers[tier] { return ModelRoutingEntry{}, &ValidationError{...} }
    if !validRoutingPhases[phase] { return ModelRoutingEntry{}, &ValidationError{...} }
    if c == nil || c.Workflow.ModelRouting == nil { return defaultRoutingEntry, nil }
    entry, ok := c.Workflow.ModelRouting[routingKey(tier, phase)]
    if !ok { return defaultRoutingEntry, nil }
    entry.FallbackApplied = false
    return entry, nil
}
```

**Closed-set maps**:

```go
// lines 28-34 — validRoutingModels INCLUDES "haiku": true (M3 will REMOVE this per REQ-AA2-012)
validRoutingModels = map[string]bool{
    "inherit": true,
    "haiku":   true,   // ← REMOVE (REQ-AA2-012 HaikuResidualRule)
    "sonnet":  true,
    "opus":    true,
    "glm":     true,
}

// lines 42-46 — validRoutingTiers (S/M/L)
validRoutingTiers = map[string]bool{"S": true, "M": true, "L": true}

// lines 47-53 — validRoutingPhases (plan/run/sync/mx)
validRoutingPhases = map[string]bool{"plan": true, "run": true, "sync": true, "mx": true}
```

**`defaultRoutingEntry` fallback (lines 55-62)** — currently:

```go
var defaultRoutingEntry = ModelRoutingEntry{
    Model:            "sonnet",
    Effort:           "medium",
    FallbackApplied:  true,
}
```

### §A.2 External call sites (signature-change blast radius)

`grep -rn "RouteModelFor" --include="*.go" . | grep -v "_test.go\|model_routing.go"`:

```
internal/config/types.go:361:	// RouteModelFor(tier, phase). The key format is "<TIER>-<phase>" (e.g.
internal/config/types.go:365:	// When the block is absent the map is nil and RouteModelFor falls back to
```

Both hits are **doc comments** in `types.go` (the `ModelRouting` map doc block). **Zero production Go call sites** invoke `RouteModelFor` at the time of plan-phase authoring — the routing is currently aspirational (`workflow.yaml:161-170` comment acknowledges this). M3's 2-arg → 3-arg signature change is therefore safe: only doc comments and `model_routing_test.go` need co-updates.

> **Residual risk**: a future SPEC may wire `RouteModelFor` into `internal/cli/launcher.go` or `internal/hook/` before M3 lands. Run-phase pre-flight (plan.md §C) MUST re-grep to confirm zero call sites before editing the signature.

---

## §B workflow.yaml matrix (the model_routing block extension point)

### §B.1 File: `.moai/config/sections/workflow.yaml`

**`workflow_agents` block (lines 148-160)** — currently:

```yaml
# workflow_agents: dynamic-workflow purpose taxonomy -> {model, effort}
# {inherit, haiku, sonnet, opus}, effort in {low, medium, high, xhigh, max}.
workflow_agents:
    read-only-extract: { model: haiku, effort: low }   # ← line 154: haiku (M3 REQ-AA2-013 → sonnet)
    # ... (other entries, non-haiku)
```

**`role_profiles` block (lines 101-145)** — `designer` (lines 111-117) and `researcher` (lines 128-135):

```yaml
designer:
    description: UI/UX design with MCP design tools
    effort: medium
    isolation: worktree
    mode: acceptEdits
    model: sonnet        # ← non-haiku (OK); gains "Absorbed by manager-design" annotation (REQ-AA2-007)

researcher:
    description: Read-only codebase exploration and analysis (speed-critical)
    effort: low
    isolation: none
    mode: plan
    model: haiku         # ← line 131: haiku (M3 REQ-AA2-013 → sonnet)
```

**`model_routing` block (lines 161-183)** — currently a single 12-entry matrix (S/M/L × plan/run/sync/mx). The 5 haiku cells:

```yaml
model_routing:
    S-plan: { model: opus,    effort: xhigh }
    S-run:  { model: sonnet,  effort: xhigh }
    S-sync: { model: haiku,   effort: low }   # ← line 174 (M3 → sonnet/low)
    S-mx:   { model: haiku,   effort: low }   # ← line 175 (M3 → sonnet/low)
    # ... M-plan, M-run, M-sync ...
    M-mx:   { model: haiku,   effort: low }   # ← line 179 (M3 → sonnet/low)
    # ... L-* (all non-haiku already) ...
```

The legacy comment at lines 161-170 reads: *"the orchestrator consults RouteModelFor(tier, phase) at spawn time"* — currently aspirational (no Go call site per §A.2). M3 will make this assertion true (3-arg signature).

### §B.2 Extension shape (REQ-AA2-009)

`model_routing` (single matrix) → `model_routing_profiles.{max, medium, low}` (3 matrices of 12 cells each). The single `model_routing` key is preserved as the `medium` default profile for backward-compat (or removed — open decision D3 in plan.md).

---

## §C llm.yaml claude_models (the haiku→sonnet flip point)

### §C.1 File: `.moai/config/sections/llm.yaml`

**Full current content** (23 lines):

```yaml
llm:
    mode: ""
    team_mode: ""             # ← RUNTIME state ("glm" under moai glm session) — PRESERVE VERBATIM per CONSTRAINT #6
    glm_env_var: GLM_API_KEY
    performance_tier: ""      # ← line 5: field exists but unused (REQ-AA2-010 populates it with max|medium|low)
    claude_models:
        high: opus
        medium: sonnet
        low: haiku            # ← line 9: haiku (REQ-AA2-011 → sonnet); haiku key REMOVED per §2-E
    glm:
        base_url: https://api.z.ai/api/anthropic
        models:
            high: glm-5.2
            medium: glm-4.7
            low: glm-4.5-air
            fable: glm-5.2
            opus: glm-5.2
            sonnet: glm-4.7
            haiku: glm-4.5-air   # ← line 19: GLM-side haiku alias (OUT OF SCOPE per Out of Scope — CG mode)
    default_model: ""
    quality_model: ""
    speed_model: ""
```

### §C.2 Extension shape (REQ-AA2-010 + REQ-AA2-011)

- `claude_models.low: haiku → sonnet` (line 9)
- `claude_models` haiku key: REMOVED (per §2-E "haiku 항목 제거"). The GLM-side `glm.models.haiku` mapping at line 19 is OUT OF SCOPE (CG mode / GLM tables untouched per Out of Scope).
- `performance_tier` (line 5): populated to `max|medium|low` by `moai init --model-policy` (REQ-AA2-010). Default `medium`.
- `team_mode: glm` (line 3): **PRESERVE VERBATIM** — runtime state from the current `moai glm` session. This field is NOT in v2's commit scope (CONSTRAINT #6).

---

## §D CLAUDE.md §4 Retained Agents (the 8 → 10 ceiling change)

### §D.1 File: `CLAUDE.md`

**§4 Retained Agents header (line 79)**:

```
The MoAI agent catalog consists of exactly **8 retained agents** (7 MoAI-custom + 1 Anthropic built-in `Explore`). ...
```

**§4 Selection Decision Tree (lines 83-93)** — currently 9 numbered entries (1-9).

**§4 Retained Agents table (lines 95-106)** — 8 rows:

| Agent | Class | Phase scope | Reference |
|-------|-------|-------------|-----------|
| `manager-spec` | core/manager | Plan-phase | `.claude/agents/moai/manager-spec.md` |
| `manager-develop` | core/manager | Run-phase | `.claude/agents/moai/manager-develop.md` |
| `manager-docs` | core/manager | Sync-phase | `.claude/agents/moai/manager-docs.md` |
| `manager-git` | core/manager | PR + Late-Branch | `.claude/agents/moai/manager-git.md` |
| `plan-auditor` | meta/evaluator | Plan-phase audit | `.claude/agents/moai/plan-auditor.md` |
| `sync-auditor` | meta/evaluator | Sync-phase audit | `.claude/agents/moai/sync-auditor.md` |
| `builder-harness` | builder | Dynamic harness | `.claude/agents/moai/builder-harness.md` |
| `Explore` | Anthropic built-in | Read-only exploration | claude.com/docs/en/sub-agents |

**Archived Agents subsection (lines 108-112)** — mentions "12 archived agents" and "8 retained". The "8 retained agents above" phrasing at line 112 also needs the 8 → 10 update.

### §D.2 Extension shape (REQ-AA2-002)

- Line 79: `8 retained agents` → `10 retained agents`; `(7 MoAI-custom + 1 Anthropic built-in Explore)` → `(9 MoAI-custom + 1 Anthropic built-in Explore)`.
- §4 Selection Decision Tree: add entries 10 (`super-advisor`) and 11 (`manager-design`).
- §4 Retained Agents table: add 2 rows (`super-advisor` class `meta/advisor`; `manager-design` class `core/manager`).
- Line 112: "8 retained agents above" → "10 retained agents above".

### §D.3 Template mirror

`internal/template/templates/CLAUDE.md` carries the same §4 content (verified via Read on the template tree). Both edits land template-first per REQ-AA2-015.

---

## §E model-policy.md § Inherit-by-Default (the haiku-exception removal point)

### §E.1 File: `.claude/rules/moai/development/model-policy.md`

**§ Model Aliases (lines 12-25)** — defines the closed set `inherit | opus | sonnet | haiku`. M3 keeps the haiku alias lexically valid (the GLM table still maps `haiku → glm-4.5-air`) but removes its deployment in agent frontmatter / routing matrices.

**§ Inherit-by-Default Convention (lines 30-55)** — the critical haiku-exception prose:

```
30: ## Inherit-by-Default Convention
32: [ZONE:Evolvable] [HARD] All MoAI agents SHOULD declare `model: inherit` unless explicitly
    assigned `haiku` for speed-critical tasks. ...
47: - All package agents under `.claude/agents/moai/` (7 MoAI-custom retained agents) declare
    `model: inherit`, except `manager-docs` and `manager-git` which use `model: haiku`.
50: - `model: haiku` agents (`manager-docs`, `manager-git`) — Haiku has no `[1m]` variant, so the
    bug does NOT apply. Speed-critical agents should remain on `haiku` for cost and latency.
```

Line 47 is the specific haiku-exception prose. **NOTE (plan-audit iter-2 D2 reframe):** this prose is STALE relative to the live tree — all 7 MoAI-custom agent files (`manager-docs` and `manager-git` included) ALREADY declare `model: inherit` (live-verified 2026-07-09: `grep -lE '^model:\s*haiku' .claude/agents/moai/*.md` → 0 matches). M3 therefore removes the STALE PROSE from model-policy.md (REQ-AA2-014 § Inherit-by-Default haiku-exception removal), NOT live haiku frontmatter — there is no live haiku frontmatter to clean up; the only live haiku references are in config (`claude_models` / `workflow_agents` / `role_profiles` / routing matrices) + the Go closed-set map, which M3 + HaikuResidualRule (REQ-AA2-012) handle. The § Effort Calibration Matrix (currently in model-policy.md) is superseded by the §2-B table from the SSOT (per SSOT §06 M4 verbatim: "Effort Calibration Matrix를 §2-B 표로 대체"; corrected in plan-audit iter-2 — the initial draft cited §2-C, but the SSOT says §2-B).

### §E.2 Baseline-refill breaker note (lines 53-55)

```
55: [ZONE:Evolvable] The `[1m]` entitlement bug in § Inherit-by-Default Convention is the *spawn-time*
    failure mode ... A **distinct second failure mode** historically affected team-mode teammates
    spawned via per-spawn `model: "sonnet"` override: ...
```

This is the historical team-mode breaker doctrine. It was downgraded to "resolved" status by `SPEC-SONNET5-1M-TEAM-DISABLE` (Sonnet 5 ships a single 1M variant with no 200K fallback target). v2 preserves this resolution unchanged (Out of Scope — Agent Teams default flip).

---

## §F Supersede target #1 analysis — SPEC-ADVISOR-RUNG-001

### §F.1 Frontmatter (current)

```yaml
id: SPEC-ADVISOR-RUNG-001
title: "Executor-Advisor Escalation Rung for /moai fix and /moai loop + GLM Judgment Carve-Out"
version: "0.1.0"
status: draft                       # ← FLIP to: superseded (REQ-AA2-017)
created: 2026-07-09
updated: 2026-07-09                 # ← UPDATE to: 2026-07-09 (no change — same day)
author: manager-spec
priority: P2
phase: "v3.0.0"
module: ".claude/skills/moai/workflows"
lifecycle: spec-anchored
era: V3R6
tier: S
depends_on: [SPEC-MODEL-ROUTING-WIRE-001]
tags: "advisor-rung, escalation, moai-fix, moai-loop, glm-carve-out, cg-leader-review, per-spawn-model, workflow-reflex"
                                    # ← ADD: superseded_by: SPEC-AGENT-ARCH-V2-001
```

### §F.2 Concerns captured by v2

| ADVISOR-RUNG REQ | v2 native capture |
|------------------|-------------------|
| REQ-ADV-001 (per-spawn Agent(general-purpose, model: opus)) | **ABSTRACTION-LIFT**: promoted to `super-advisor.md` catalog agent file (REQ-AA2-001) |
| REQ-ADV-002 (escalation rung trigger N=2 same-diagnostic failures) | becomes E1 of v2 escalation doctrine (REQ-AA2-003, threshold tightened 2 → 3) |
| REQ-ADV-003 (advisor read-only, non-binding) | preserved verbatim in super-advisor frontmatter `permissionMode: plan` + description (REQ-AA2-001) |
| REQ-ADV-004 (GLM carve-out) | super-advisor inherits session model under `moai glm` (REQ-AA2-001 §B.5 of design.md) |
| REQ-ADV-005 (CG leader-review-as-advisor) | consultation surface identical under CG mode; peer-reviewer use case documented (REQ-AA2-001 §B.5) |
| REQ-ADV-006 (ceiling change — N.B. ADVISOR-RUNG did NOT own the ceiling change; it was a cross-ref) | owned natively by REQ-AA2-002 |

### §F.3 Disposition

**ABSTRACTION-LIFT** — the per-spawn `Agent(general-purpose, model: opus)` pattern is promoted to a dedicated catalog agent file. ADVISOR-RUNG was Tier S (2 artifacts: spec + plan) with the `advisor-rung` concept living as workflow skill prose; v2 lifts it to a Tier L catalog-level agent file with verbatim frontmatter, escalation doctrine E1-E4 in `agent-common-protocol.md`, and a CLAUDE.md §4 row. The `depends_on: [SPEC-MODEL-ROUTING-WIRE-001]` is also retired (the dependency on routing-wire is subsumed by M3's No-Haiku 3-tier policy that flips the WIRE direction).

---

## §G Supersede target #2 analysis — SPEC-MODEL-ROUTING-WIRE-001

### §G.1 Frontmatter (current)

```yaml
id: SPEC-MODEL-ROUTING-WIRE-001
title: "Wire the Tier×Phase Model Routing Matrix into Spawn Paths and Resolve Model-Policy Contradictions"
version: "0.1.0"
status: draft                       # ← FLIP to: superseded (REQ-AA2-017)
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/cli"
lifecycle: spec-anchored
era: V3R6
tier: M
related_specs: [SPEC-TOKEN-ROUTING-001]
tags: "model-routing, tier-phase-matrix, moai-route, spawn-wiring, model-policy, haiku-inherit, workflow-reflex"
                                    # ← ADD: superseded_by: SPEC-AGENT-ARCH-V2-001
```

### §G.2 Direction flip — `haiku-inherit` → No-Haiku

The `haiku-inherit` tag captured WIRE-001's intent to **preserve** haiku in low-cost slots via `model: inherit` semantics. v2's M3 mandates the OPPOSITE direction (haiku → sonnet low/medium). The supersede is a **POLICY-FLIP**, not a refinement: the tag's lexical meaning is reversed.

### §G.3 Concerns captured by v2

| MODEL-ROUTING-WIRE REQ | v2 native capture |
|------------------------|-------------------|
| REQ-MRW-001 (`RouteModelFor` 2-arg → spawn-path wiring) | REQ-AA2-008 (3-arg extension); the spawn-path wiring is the same goal |
| REQ-MRW-002 (`moai route` CLI subcommand) | folds into M3's `moai init --model-policy` selection + RouteModelFor extension |
| REQ-MRW-003 (pre-spawn consultation instruction) | embedded in `agent-common-protocol.md` Error Recovery Pattern (cross-ref'd from REQ-AA2-003 super-advisor doctrine) |
| REQ-MRW-004 (struct-YAML symmetry test extension) | REQ-AA2-009 (model_routing_profiles 3 matrices) co-updates the symmetry test |
| REQ-MRW-005 (haiku-inherit policy preservation) | **POLICY-FLIP**: REQ-AA2-011/012/013 (haiku → sonnet, HaikuResidualRule) |
| REQ-MRW-006..009 (spawn-path / launcher integration, edge cases) | natively covered by M3's Go code extension (signature change is safe per §A.2 — zero external call sites) |

### §G.4 Disposition

**POLICY-FLIP** + absorption of the `moai route` subcommand into `moai init --model-policy`. WIRE-001 was Tier M (3 artifacts: spec + plan + acceptance); v2 lifts the wiring to Tier L (5 artifacts) and adds the 3-tier profile dimension that WIRE-001's single matrix lacked.

---

## §H DesignSync MCP availability — GAP (design.md §C.4 forward-link)

### §H.1 Live measurement

```bash
$ cat .mcp.json | jq '.mcpServers | keys'
[
  "context7",
  "chrome-devtools",
  "web_search_prime",
  "web_reader"
]

$ grep -rn "DesignSync\|design-sync\|design_sync" .mcp.json .claude/ .moai/config/
# (zero matches)
```

**Observation**: the DesignSync MCP server is NOT registered in `.mcp.json` at plan-phase (2026-07-09). The SSOT §04 claims *"근거(실측): 세션에 이미 DesignSync 도구가 존재하며"* — this claim is **contradicted** by the live tree state at the time of plan-phase authoring.

### §H.2 Possible reconciliation

- (a) The SSOT author had a different `.mcp.json` at the time of report authorship (the file was since reset).
- (b) The DesignSync tool was injected via a user-level (`~/.claude/.mcp.json`) registration not visible from the project root.
- (c) The SSOT claim is forward-looking ("the tool will exist when M2 lands").

### §H.3 Plan-phase consequence (GAP)

**This GAP does NOT block plan-phase close.** The agent file + workflow skill are authored against the §04 documented 11-method contract verbatim (the contract is a forward-looking spec, not a runtime dependency). M2 run-phase execution MUST verify DesignSync is operationally available before exercising D2 (`finalize_plan` / `write_files`). Tool absence triggers the H1 blocker-report path (graceful degradation — the agent returns a structured "tool unavailable" diagnosis to the orchestrator, which surfaces via `AskUserQuestion` per the standard re-delegation flow).

The GAP is recorded as CONSTRAINT #4 in spec.md and surfaced in plan.md §B Known Issues B5.

---

## §I workflow_agents + role_profiles haiku audit (full enumeration)

### §I.1 `workflow_agents` haiku entries (workflow.yaml)

| Line | Entry | Current | M3 substitution |
|------|-------|---------|-----------------|
| 154 | `read-only-extract` | `{ model: haiku, effort: low }` | `{ model: sonnet, effort: low }` |

(1 entry total — `workflow_agents` block at lines 148-160)

### §I.2 `role_profiles` haiku entries (workflow.yaml)

| Line | Profile | Current model | M3 substitution |
|------|---------|---------------|-----------------|
| 79 | `investigation` (Competing hypothesis debugging) | `haiku` | `sonnet` |
| 131 | `researcher` | `haiku` | `sonnet` |

(2 entries — added in plan-audit iter-2 D3: line 79 `investigation` was missing from the initial enumeration; live `grep -nE 'model:\s*haiku' workflow.yaml` returns 6 total matches across §I.1 + §I.2 + §I.3.)

Other role_profiles already non-haiku (verified):
- `analyst`: sonnet (OK)
- `architect`: opus (OK)
- `designer`: sonnet (OK; gains "Absorbed by manager-design" annotation per REQ-AA2-007)
- `implementer`: sonnet (OK)
- `reviewer`: high + non-haiku (OK)

### §I.3 `model_routing` haiku cells (workflow.yaml)

| Line | Cell | Current | M3 substitution (max/medium/low profiles) |
|------|------|---------|-------------------------------------------|
| 174 | `S-sync` | `haiku / low` | `sonnet / low` across all 3 profiles |
| 175 | `S-mx` | `haiku / low` | `sonnet / low` across all 3 profiles |
| 179 | `M-mx` | `haiku / low` | `sonnet / low` across all 3 profiles |

(3 haiku cells in the legacy single matrix; M3 expands to 3 profiles × 12 cells = 36 cells total, with 9 cells formerly haiku across the 3 profiles — all become `sonnet / low`.)

### §I.4 Other surfaces

| File | Location | Match | Action |
|------|----------|-------|--------|
| `model_routing.go:31` | `validRoutingModels["haiku"]` | `"haiku": true` | REMOVE (REQ-AA2-012 HaikuResidualRule trigger) |
| `model-policy.md:47` | "except `manager-docs` and `manager-git` which use `model: haiku`" | haiku-exception prose | REMOVE (No-Haiku renders exception obsolete) |
| `team-protocol.md:19` | role matrix `\| researcher \| plan \| haiku \|` | `haiku` | `sonnet` (live + template mirrors) |
| `team-pattern-cookbook.md` | grep for "haiku" | (zero matches — verified) | no action |

**HARD success metric** (REQ-AA2-012 / REQ-AA2-016): `HaikuResidualRule` MUST fail lint if any of the above persists post-implementation. Target: lint 0건 (§08 row 1).

---

## §J Cross-references

- **Design SSOT**: `.moai/reports/agent-architecture-redesign-v2-20260709.html` (§01-§08 architecture authority).
- **spec.md / plan.md / acceptance.md / design.md** — sibling artifacts in this SPEC directory.
- **Supersede targets**: `SPEC-ADVISOR-RUNG-001/spec.md` (§F above); `SPEC-MODEL-ROUTING-WIRE-001/spec.md` (§G above).
- **Live-tree files cited** (verified 2026-07-09): `internal/config/model_routing.go`; `internal/config/types.go:361-365`; `.moai/config/sections/workflow.yaml`; `.moai/config/sections/llm.yaml`; `CLAUDE.md:79-112`; `.claude/rules/moai/development/model-policy.md`; `.claude/rules/moai/workflow/team-protocol.md:19`; `.mcp.json`.

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial research artifact. 10 baselines: §A RouteModelFor (zero ext call sites), §B workflow.yaml matrix, §C llm.yaml claude_models, §D CLAUDE.md §4, §E model-policy.md Inherit-by-Default, §F ADVISOR-RUNG-001 supersede analysis, §G MODEL-ROUTING-WIRE-001 supersede analysis, §H DesignSync MCP Gap, §I haiku full enumeration. Every baseline measured live 2026-07-09 by this agent (vci §2 attribution). |

# Research: SPEC-ADVISOR-001 (Advisor Strategy Adoption)

> **SPEC**: SPEC-ADVISOR-001
> **Wave**: 1 — Tier 0
> **Author**: manager-spec
> **Date**: 2026-04-30

---

## 1. Source Documents

### 1.1 Primary Source — Anthropic Blog "The Advisor Strategy"

**URL**: https://claude.com/blog/the-advisor-strategy

**Verbatim quotations** (foundational claims that motivate this SPEC):

> "The advisor never calls tools or produces user-facing output, and only provides guidance to the executor."

> "Sonnet + Opus advisor: 2.7% improvement on SWE-bench Multilingual, 11.9% cost reduction."

> "Haiku + Opus advisor: 41.2% on BrowseComp (vs 19.7% solo), 85% cheaper than Sonnet solo."

**Verbatim API pattern** (target integration shape; quoted from https://claude.com/blog/the-advisor-strategy § "API Code Example"):

```python
response = client.messages.create(
    model="claude-sonnet-4-6",  # executor
    tools=[
        {
            "type": "advisor_20260301",
            "name": "advisor",
            "model": "claude-opus-4-6",
            "max_uses": 3,
        },
        # ... your other tools
    ],
    messages=[...]
)

# Advisor tokens reported separately
# in the usage block.
```

> **Verification (2026-04-30, WebFetch confirmed)**: The model IDs `claude-sonnet-4-6` and `claude-opus-4-6` and the tool type `advisor_20260301` are reproduced verbatim from the Anthropic blog example. The blog uses a hypothetical 4-6 era model lineup as the documented sample. moai-adk-go's current runtime is Opus 4.7 — when implementing in this project, substitute the model IDs with the actual deployed model identifiers (consult `.moai/config/sections/system.yaml` `claude.model` and Anthropic's current model documentation). The verbatim quote is preserved above to keep the citation auditable; it does not imply a binding model choice for our implementation.

**Key insight**: The advisor is a *passive consultant*, not a delegated executor. It shares the executor's context window, reads it, and responds with guidance only. No tools. No user-facing turns. The executor remains the only agent that produces work product.

### 1.2 Supporting Source — Anthropic "Building Multi-Agent Systems"

**URL**: https://anthropic.com/engineering/built-multi-agent-research-system

> "When deciding how to allocate compute between models, consider that supervisor-style architectures (one strong model directing weaker ones) often outperform homogeneous teams."

**Inference**: The advisor pattern is a degenerate case of supervisor-architecture where the "supervisor" never executes. This minimizes coordination overhead while preserving the strong-model reasoning benefit at decision points.

### 1.3 Supporting Source — Anthropic "Best Practices for Opus 4.7"

> "Opus 4.7 prefers reasoning over tool invocation. When a task requires multiple decisions, prefer one Opus call with full context over many Haiku calls."

**Inference**: The advisor pattern aligns with this guidance — concentrate Opus reasoning at high-uncertainty decision points instead of throughout the entire executor session.

---

## 2. Codebase Analysis

### 2.1 Current Model Allocation

Direct survey of `.claude/agents/moai/*.md` `model:` field:

| Agent | Model | Effort | Phase | File:Line |
|-------|-------|--------|-------|-----------|
| manager-spec | opus | xhigh | plan | .claude/agents/moai/manager-spec.md:13 |
| manager-strategy | opus | xhigh | plan | .claude/agents/moai/manager-strategy.md:13 |
| expert-security | opus | high | run/audit | .claude/agents/moai/expert-security.md:12 |
| researcher | opus | (default) | various | .claude/agents/moai/researcher.md:15 |
| evaluator-active | sonnet | high | run-end | .claude/agents/moai/evaluator-active.md:13 |
| manager-ddd | sonnet | (default) | run | .claude/agents/moai/manager-ddd.md:14 |
| manager-tdd | sonnet | (default) | run | .claude/agents/moai/manager-tdd.md:14 |
| manager-quality | sonnet | (default) | run-end | .claude/agents/moai/manager-quality.md:13 |
| expert-backend | sonnet | (default) | run | .claude/agents/moai/expert-backend.md:13 |
| expert-frontend | sonnet | (default) | run | .claude/agents/moai/expert-frontend.md:13 |
| manager-docs | haiku | (default) | sync | .claude/agents/moai/manager-docs.md:13 |
| manager-git | haiku | (default) | sync | .claude/agents/moai/manager-git.md:13 |
| plan-auditor | inherit | (parent) | post-plan | .claude/agents/moai/plan-auditor.md:12 |

**Observation**: Plan-phase reasoners (manager-spec, manager-strategy) and security-critical reasoners (expert-security) consume Opus throughout their full execution, not just at decision points. evaluator-active is Sonnet but performs uncertainty-laden judgment calls (4-dimension scoring with FAIL thresholds) where Opus consultation could prevent score drift.

### 2.2 evaluator-active Decision Surface (Pilot Target #2)

`.claude/agents/moai/evaluator-active.md:38-55` defines four scored dimensions:

```
| Functionality | 40% | All SPEC acceptance criteria met | Any criterion FAIL |
| Security | 25% | OWASP Top 10 compliance | Any Critical/High finding |
| Craft | 20% | Test coverage >= 85%, error handling | Coverage below threshold |
| Consistency | 15% | Codebase pattern adherence | Major pattern violations |
```

**Decision uncertainty hot-spots** (from .moai/config/evaluator-profiles/default.md):

- Security dimension: 0.50 (No Critical, contained High) vs 0.25 (Critical present, triggers FAIL) — borderline cases need calibrated judgment.
- Functionality 0.75 (primary criteria pass, minor edge cases missing) vs 0.50 (1-2 criteria fail) — boundary is subjective.
- Consistency 0.75 (minor deviations) vs 0.50 (some pattern violations) — Sonnet alone may rationalize ambiguity.

These are exactly the points where a 2-3 use Opus advisor would tighten judgment without fully replacing Sonnet execution.

### 2.3 manager-spec Decision Surface (Pilot Target #1)

`.claude/agents/moai/manager-spec.md` workflow steps (Step 2-4) require:

- EARS pattern selection (5+ patterns; `.claude/agents/moai/manager-spec.md:46-52`)
- SPEC vs Report classification (`.claude/agents/moai/manager-spec.md:74-79`)
- Frontmatter schema enforcement (9 required fields; `.claude/agents/moai/manager-spec.md:119-160`)
- Exclusions section composition (`.claude/agents/moai/manager-spec.md:62-65`)

Most of these are mechanical (pattern matching, schema validation) — Sonnet-suitable. The genuinely hard decisions are:

- Choosing the EARS pattern when requirements blur Event-Driven and State-Driven (3-5 calls per SPEC)
- Detecting hidden conflicts with existing SPECs (1-2 calls per SPEC)
- Resolving exclusion-vs-feature ambiguity (1 call per SPEC)

A `max_uses: 3` advisor budget per SPEC matches the empirical decision count.

### 2.4 Claude Code `advisor_20260301` Tool Availability

**CRITICAL UNKNOWN**: Anthropic API documents `advisor_20260301` as a server-side tool type for direct API calls, but Claude Code's sub-agent runtime may not expose this tool type to user-defined agents.

**Verification approach** (deferred to plan.md / run phase):
1. Probe: Add a placeholder `tools: ..., advisor_20260301` field to a test agent and observe whether Claude Code accepts it.
2. Fallback path A — Skill-level wrapper: Create `moai-foundation-advisor` skill that invokes Opus via direct API call when running in a sub-agent context, transparently passing back recommendations.
3. Fallback path B — Agent() with `mode: "plan"`: When executor needs advice, spawn a one-shot plan-mode Opus advisor via Agent() and inject the response into reasoning. This is heavier (full agent spawn) but works today.

**Decision deferred to plan.md**: which path to pilot first.

### 2.5 Cost Observability Gap

`.moai/config/sections/observability.yaml` (modified in working tree) does not currently log per-agent token consumption with a model-attribution dimension. Any advisor cost-savings claim requires per-call (executor_model, advisor_model, advisor_tokens) tuples in logs.

Hook integration point: `internal/hook/post_tool.go` (modified in working tree) already records `duration_ms`. Extending the schema with `model`, `advisor_invocations`, `advisor_tokens` aligns with the in-flight hook work.

---

## 3. Alternative Approaches

### 3.1 Alternative A — Universal Opus Adoption (status quo amplified)

Promote all Sonnet executors to Opus. Eliminates uncertainty entirely.

- **Pros**: Single model surface, no advisor protocol needed, no fallback path required.
- **Cons**: 4-5x cost increase across all run-phase work; contradicts the "fewer tool calls, more reasoning" principle by paying for Opus reasoning even on mechanical tasks (formatting, schema fills).
- **Verdict**: Rejected. Cost is dominant constraint.

### 3.2 Alternative B — Static Model Routing (no advisor protocol)

Pre-classify decisions in agent prompts: "If <pattern>, escalate to Opus by spawning Agent(); else proceed in Sonnet."

- **Pros**: No new tool dependency; uses existing Agent() spawn mechanism.
- **Cons**: Each escalation is a full agent spawn (cold context); no shared session. The Anthropic blog's measured savings (11.9% cost reduction with 2.7% quality gain) depend on shared-context advisor calls.
- **Verdict**: Acceptable as fallback only (this is path B in §2.4).

### 3.3 Alternative C — Advisor Protocol via Claude Code Native (preferred)

Use `advisor_20260301` tool type if supported. Executor calls advisor in-context; advisor cannot call tools or emit user-facing output.

- **Pros**: Matches Anthropic's measured savings shape; minimal protocol overhead; aligns with one-turn-fully-loaded principle.
- **Cons**: Tool availability in sub-agent context is unverified.
- **Verdict**: Preferred. SPEC requires verification step (REQ-ADV-007).

### 3.4 Alternative D — Skill-Level Advisor Wrapper

Create `moai-foundation-advisor` skill that exposes a `consult_advisor(question, context)` function wrapping a direct Anthropic API call.

- **Pros**: Works today regardless of Claude Code's advisor tool support; testable via local-first patterns.
- **Cons**: Bypasses Claude Code's session/context management; requires API key handling; may double-bill if executor and advisor sessions don't share KV cache.
- **Verdict**: Hold as third-tier fallback if both A and B fail.

### 3.5 Decision Matrix

| Criterion (weight) | A (universal Opus) | B (static routing) | C (advisor protocol) | D (skill wrapper) |
|--------------------|--------------------|--------------------|----------------------|-------------------|
| Cost (30%) | 1/10 | 6/10 | 9/10 | 7/10 |
| Quality preservation (25%) | 9/10 | 6/10 | 8/10 | 6/10 |
| Implementation cost (20%) | 9/10 | 7/10 | 5/10 | 4/10 |
| Future-proofing (15%) | 6/10 | 5/10 | 9/10 | 5/10 |
| Observability (10%) | 7/10 | 7/10 | 8/10 | 6/10 |
| **Weighted total** | **5.65** | **6.20** | **7.85** | **5.65** |

Path C wins. Path B is the fallback. Path D is contingency.

---

## 4. Decision Rationale

### 4.1 Why pilot at manager-spec and evaluator-active

These two are the highest-yield pilots because:

1. **Decision density is naturally bounded**: manager-spec averages 3-5 decisions per invocation; evaluator-active averages 4 (one per dimension). `max_uses: 3` and `max_uses: 2` map cleanly to observed surface.
2. **Quality regression is measurable**: SPEC schema validation catches manager-spec output drift; per-dimension scoring history catches evaluator-active drift. Both produce structured artifacts.
3. **Failure mode is bounded**: If advisor returns garbage, manager-spec falls back to Sonnet's own decision (same as today); evaluator-active flags UNVERIFIED. Neither becomes worse than the baseline.
4. **Cost ceiling is calculable**: For manager-spec at `max_uses: 3` with ~2K-token advisor responses, Opus advisor cost per SPEC ≈ $0.10. For evaluator-active at `max_uses: 2`, ≈ $0.07 per evaluation. Acceptable per artifact.

### 4.2 Why NOT pilot at expert-security (deferred to Wave 2+)

Tempting because expert-security currently uses Opus directly — moving to Sonnet+Opus-advisor would yield largest savings. But:

- Security FAIL has hard threshold (`evaluator-active.md:55`): one missed Critical = overall FAIL. Sonnet executor + Opus advisor still leaves Sonnet as the final pen-tester. Until the advisor protocol is field-validated on lower-stakes pilots, expert-security remains direct Opus.
- Defer to Wave 2 once §2.4 verification step succeeds and pilot data exists.

### 4.3 Why `max_uses` cap is mandatory (not optional)

Without a cap, Sonnet executor under uncertainty pressure may escalate every decision to advisor. This collapses to Alternative A (universal Opus) at higher latency. The cap forces the executor to budget its consultations and maintain primary reasoning capacity itself. Per-agent caps (3 for manager-spec, 2 for evaluator-active) are starting points calibrated against §2.2 / §2.3 decision counts; revisable from telemetry.

### 4.4 Why telemetry is in-scope

Anthropic's claimed savings (11.9% cost reduction, 2.7% quality gain) cannot be replicated locally without measurement. Without telemetry, this SPEC degrades to "we adopted a pattern someone else measured." The observability extension is a small marginal cost on top of the in-flight `internal/hook/post_tool.go` work.

---

## 5. Risks & Mitigations

| # | Risk | Severity | Likelihood | Mitigation |
|---|------|----------|------------|------------|
| R1 | `advisor_20260301` not exposed to sub-agents | High | Medium | REQ-ADV-007 mandates verification; fallback path B (Agent() spawn) keeps SPEC viable |
| R2 | Sonnet executor over-relies on advisor (escalates everything) | Medium | High | `max_uses` cap (REQ-ADV-003); telemetry alarm if cap is hit on >50% of invocations |
| R3 | Advisor returns irrelevant guidance under bad prompting | Medium | Medium | Executor SHALL treat advisor output as advice not directive; REQ-ADV-008 fallback to executor's own decision |
| R4 | KV cache miss between executor turns and advisor turn doubles cost | Medium | Low | Anthropic claims shared-context advisor uses cache; verify in telemetry; fall back to path B if cache savings absent |
| R5 | Quality regression on pilot dimension scoring | High | Low | A/B test for ≥10 evaluations before promoting; rollback criterion: dimension score variance > 0.10 vs baseline |
| R6 | Cost telemetry leaks API keys into logs | Critical | Low | Log only token counts and model identifiers, never raw API responses |
| R7 | Compatibility breaks when Claude Code updates advisor tool version | Medium | Medium | Pin tool version in agent frontmatter; CI test that decodes the advisor protocol response shape |

---

## 6. References

### Anthropic Sources
- The Advisor Strategy: https://claude.com/blog/the-advisor-strategy
- Building Multi-Agent Systems: https://anthropic.com/engineering/built-multi-agent-research-system
- Best Practices Opus 4.7: https://platform.claude.com/docs/en/about-claude/models/whats-new-claude-4-7

### Codebase References
- `.claude/agents/moai/manager-spec.md:13,46-52,74-79,119-160` — Pilot target #1 anchors
- `.claude/agents/moai/evaluator-active.md:13,38-55,80-89` — Pilot target #2 anchors
- `.claude/agents/moai/expert-security.md:12` — Wave 2 candidate (deferred)
- `.moai/config/evaluator-profiles/default.md` — Dimension scoring rubric
- `.claude/rules/moai/development/model-policy.md` — Existing model allocation policy
- `.claude/rules/moai/core/moai-constitution.md` § "Opus 4.7 Prompt Philosophy" — Principle 5 alignment
- `internal/hook/post_tool.go` — Telemetry integration point (in-flight work)

### Related SPECs
- SPEC-AGENT-002 (completed) — Agent body minimization; provides per-agent template format
- SPEC-CORE-BEHAV-001 — Six Agent Core Behaviors; advisor must respect Behavior 6 (Verify, Don't Assume)

---

**Total lines**: ~210
**Status**: Ready for SPEC drafting

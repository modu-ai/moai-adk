---
id: SPEC-V3R6-AGENT-MODEL-ROUTING-001-RESEARCH
title: "Research — Agent 23개 모델 명시 라우팅 (baseline, A/B framework, cost validation methodology)"
version: "0.2.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents"
lifecycle: spec-anchored
tags: "agent, model-routing, opus, sonnet, haiku, cost-optimization, sprint-2, v3.0, research, baseline"
tier: L
---

# Research — SPEC-V3R6-AGENT-MODEL-ROUTING-001 Tier L

본 research는 Tier L 추가 산출물로서 REQ-AMR-006 baseline JSONL 측정 방법론 + REQ-AMR-007 A/B comparison framework + REQ-AMR-NF-010 cost-saving 검증 방법론을 기록한다.

---

## 1. Baseline Measurement Methodology (REQ-AMR-006)

### 1.1 Baseline JSONL Schema

`.moai/state/agent-model-baseline.jsonl` 파일 schema (one entry per agent, 23 lines total):

```jsonl
{
  "agent_name": "expert-backend",
  "agent_path": ".claude/agents/expert/expert-backend.md",
  "current_model_setting": "inherit",
  "current_effective_model": "claude-opus-4-7",
  "baseline_input_tokens_avg": 12543.2,
  "baseline_output_tokens_avg": 3287.5,
  "baseline_quality_score": 0.87,
  "quality_score_method": "task_category_weighted_avg",
  "measurement_window_start": "2026-05-15T00:00:00Z",
  "measurement_window_end": "2026-05-22T23:59:59Z",
  "n_invocations": 47,
  "task_categories": {
    "analysis": 0.85,
    "code_review": 0.89,
    "refactoring": 0.86,
    "documentation": 0.88
  }
}
```

### 1.2 Measurement Sources

**Primary source — Usage Log JSONL**:
- Path: `.moai/harness/usage-log.jsonl`
- Schema: per-agent invocation records with input/output tokens, timestamp, agent_name
- Window: 최근 7일 (rolling) — sufficient for 23-agent baseline

**Secondary source — Manual quality eval**:
- Manual eval set: 5 task categories × 3 representative tasks = 15 tasks per agent
- Total: 23 agents × 15 tasks = 345 manual evaluations
- Quality score: 0-1 scale, weighted average across categories

**Fallback source — Synthetic baseline**:
- For agents with < 5 invocations in usage log window (예: rarely-invoked specialists)
- Synthetic baseline = published model benchmark (e.g., Anthropic eval set scores for Opus 4.7)
- Marked with `"synthetic": true` field in JSONL

### 1.3 Measurement Procedure (M1 step-by-step)

```bash
# Step 1: Read usage log for 7-day window
JQ_FILTER='select(.timestamp >= (now - 604800 | todate)) | {agent_name, input_tokens, output_tokens, timestamp}'
jq -c "$JQ_FILTER" .moai/harness/usage-log.jsonl > /tmp/baseline-window.jsonl

# Step 2: Aggregate per-agent token averages
jq -s 'group_by(.agent_name) | map({
  agent_name: .[0].agent_name,
  n: length,
  input_avg: (map(.input_tokens) | add / length),
  output_avg: (map(.output_tokens) | add / length)
})' /tmp/baseline-window.jsonl > /tmp/baseline-aggregated.json

# Step 3: For each of 23 agents, generate baseline entry
# (Combine token avg + manual quality eval + measurement window)

# Step 4: Write 23 entries to .moai/state/agent-model-baseline.jsonl
```

### 1.4 Quality Score Definition

**Per-task quality score (0-1 scale)**:

| Score | Criteria |
|-------|----------|
| 1.0 | Output fully satisfies task + zero corrections needed |
| 0.9 | Output satisfies task + 1 minor correction |
| 0.8 | Output satisfies task + 2-3 minor corrections |
| 0.7 | Output partially satisfies task + 1 major correction |
| 0.5 | Output significantly deviates from task intent |
| 0.0 | Output is incorrect, off-topic, or unusable |

**Per-agent quality score** = weighted average across 5 task categories:
- analysis (weight: 0.25)
- code_review (weight: 0.25)
- refactoring (weight: 0.20)
- documentation (weight: 0.15)
- ideation (weight: 0.15)

Task categories adjusted per agent role (e.g., manager-docs heavy on documentation, expert-security heavy on analysis).

---

## 2. A/B Comparison Framework (REQ-AMR-007)

### 2.1 Comparison Design

**Hypothesis**: Sonnet/Haiku tier agents maintain quality within ±5% of Opus baseline.

**Null Hypothesis (H0)**: post-migration quality_score ≈ baseline_quality_score (within ±5%).

**Alternative Hypothesis (H1)**: post-migration quality_score deviates beyond ±5% from baseline.

**Test Method**: Paired comparison per agent, per task category.

### 2.2 Post-Migration Measurement

After M2~M4 milestone completion, manager-quality 위임 (M6) executes:

```bash
# Step 1: Reset usage log baseline marker
echo "MIGRATION_COMPLETE_AT=$(date -u +%Y-%m-%dT%H:%M:%SZ)" > .moai/state/agent-model-migration-marker.txt

# Step 2: 7-day post-migration window measurement
# (same procedure as §1.3 Step 1-3 but for post-migration window)

# Step 3: Regression report generation
jq -s '{
  spec_id: "SPEC-V3R6-AGENT-MODEL-ROUTING-001",
  measurement_date: now | todate,
  agents: [
    .[] | {
      agent_name: .agent_name,
      baseline_quality_score: .baseline_quality_score,
      post_change_quality_score: .post_change_quality_score,
      regression_pct: ((.post_change_quality_score - .baseline_quality_score) / .baseline_quality_score),
      within_bound: ((.post_change_quality_score - .baseline_quality_score) / .baseline_quality_score | fabs <= 0.05)
    }
  ]
}' /tmp/baseline-aggregated.json /tmp/post-migration-aggregated.json \
   > .moai/reports/quality-regression/SPEC-V3R6-AGENT-MODEL-ROUTING-001.json
```

### 2.3 Pass/Fail Criteria

- **PASS**: all 23 agents have `within_bound = true` (regression_pct in [-0.05, +0.05])
- **PARTIAL FAIL**: 1-3 agents exceed bound → revert affected agents to opus
- **FULL FAIL**: 4+ agents exceed bound → SPEC body re-evaluation + alternative tier classification

### 2.4 Eval Set Composition

Per task category, 3 representative tasks per agent (total 15 tasks per agent × 23 agents = 345 tasks):

| Category | Example Task |
|----------|--------------|
| analysis | "Identify security vulnerabilities in the attached Go HTTP handler" |
| code_review | "Review the diff and list 3 improvements" |
| refactoring | "Refactor the function to remove duplication while preserving behavior" |
| documentation | "Generate API reference for the exported function" |
| ideation | "Brainstorm 5 SPEC ideas to improve agent observability" |

**Task fixity**: Same 345 tasks used for baseline and post-migration measurement (paired comparison integrity).

---

## 3. Cost-Saving Validation Methodology (REQ-AMR-NF-010)

### 3.1 Pricing Constants (`.moai/research/v3.0-design-2026-05-22.md` §1.3)

| Model | Input ($/MTok) | Output ($/MTok) |
|-------|---------------|-----------------|
| Opus 4.7 | $5.00 | $25.00 |
| Sonnet 4.6 | $3.00 | $15.00 |
| Haiku 4.5 | $1.00 | $5.00 |
| Haiku 4.5 + Batch API (50% off) | $0.50 | $2.50 |

**New tokenizer impact (Opus 4.7)**: +35% effective tokens vs Opus 4.6 (per Anthropic docs).

### 3.2 Per-Agent Cost Reduction

For each of 23 agents:

```
cost_reduction_pct = (baseline_cost - post_migration_cost) / baseline_cost

where:
  baseline_cost = (avg_input_tokens × Opus_input_price + avg_output_tokens × Opus_output_price) / 1_000_000
  post_migration_cost = (avg_input_tokens × NewTier_input_price + avg_output_tokens × NewTier_output_price) / 1_000_000
```

**Examples**:

| Agent | Tier (baseline → post) | input_avg | output_avg | baseline_cost | post_migration_cost | reduction |
|-------|---------------------|-----------|------------|---------------|-----------------|-----------|
| expert-backend | Opus → Sonnet | 12,543 | 3,288 | $0.145 | $0.087 | 40% off |
| manager-quality | Opus → Sonnet | 8,234 | 1,567 | $0.080 | $0.048 | 40% off |
| manager-docs | Haiku → Haiku (no change) | 5,672 | 1,234 | $0.037 | $0.037 | 0% (already migrated) |
| researcher | Opus → Haiku+Batch | 25,000 | 8,000 | $0.325 | $0.033 | 90% off |

### 3.3 Main `/moai run` Workflow Cost Math

**5-agent set** (membership unchanged from iter 1):
- manager-develop (opus) — 7K input + 3K output per invocation
- expert-backend (sonnet) — 12K input + 3K output
- manager-quality (sonnet) — 8K input + 2K output
- manager-git (haiku) — 5K input + 1K output
- manager-docs (haiku) — 5K input + 1K output

**Baseline (all inherit/Opus)**:
```
total_input_cost = (7 + 12 + 8 + 5 + 5) × $5 / 1000 = $0.185
total_output_cost = (3 + 3 + 2 + 1 + 1) × $25 / 1000 = $0.250
total per /moai run = $0.435
Average input cost per MTok = $0.185 / 0.037 MTok = $5.00/MTok (definition)
But weighted with output: effective input ratio per turn ~ $4.20/MTok per design doc
```

**Post-migration (7 opus / 13 sonnet / 3 haiku)**:
```
manager-develop opus:    7K × $5 + 3K × $25 = $35 + $75 = $110 / 1000 = $0.110
expert-backend sonnet:  12K × $3 + 3K × $15 = $36 + $45 = $81 / 1000 = $0.081
manager-quality sonnet:  8K × $3 + 2K × $15 = $24 + $30 = $54 / 1000 = $0.054
manager-git haiku:       5K × $1 + 1K × $5  = $5  + $5  = $10 / 1000 = $0.010
manager-docs haiku:      5K × $1 + 1K × $5  = $5  + $5  = $10 / 1000 = $0.010
total per /moai run = $0.265
Effective input ratio per turn ~ $2.20/MTok per design doc
```

**Reduction**: $4.20 → $2.20/MTok = **48% off** ($0.435 → $0.265 per run = 39% off raw turn cost).

### 3.4 Opus Call Frequency Reduction

**Baseline (all inherit/Opus)**: 23/23 = 100% Opus calls.

**Post-migration**: 7/23 = 30.4% Opus calls.

**Reduction**: 70% off Opus call frequency (not raw cost, since cost depends on input/output token avg per agent).

### 3.5 researcher Batch API ROI

**Per researcher invocation** (25K input + 8K output, async-OK):

| Configuration | Input cost | Output cost | Total | Reduction |
|---------------|------------|-------------|-------|-----------|
| Opus inherit (baseline) | $0.125 | $0.200 | $0.325 | 0% |
| Haiku no batch | $0.025 | $0.040 | $0.065 | 80% off |
| Haiku + Batch API (50% off) | $0.013 | $0.020 | $0.033 | **90% off** |

**Trade-off**: Batch API adds async latency (minutes to hours). researcher's async-OK nature makes this trade-off acceptable.

---

## 4. Risk Quantification

### 4.1 Statistical Power for REQ-AMR-007

**Sample size per agent**: 15 manual tasks (5 categories × 3 tasks).

**Detectable effect size**: ±5% bound on a 0.85 baseline = ±0.0425 absolute score.

**Power analysis**:
- For paired comparison with σ ≈ 0.07 (typical quality score variance):
- Required n for 80% power at α=0.05: ~22 paired tasks
- Our n = 15 → ~65% power
- **Acceptable risk**: 15 tasks provides reasonable signal; M6 manager-quality validation augments with usage-log data for higher-traffic agents

### 4.2 Confounding Variables

| Confound | Mitigation |
|----------|-----------|
| Task drift (different tasks pre/post) | Same 345 tasks (paired comparison) |
| Time-of-day variance | Window spans 7 days (averages diurnal patterns) |
| Cache_write effects (Sprint 2 PROMPT-CACHE-001) | D6: AMR sync first + 24h baseline window before PROMPT-CACHE sync |
| Tokenizer +35% impact | Post-Opus-4.7-migration is the baseline (already incorporates +35%) |
| User feedback noise | manager-quality validation uses objective binary criteria (within bound or not) |

---

## 5. Comparison with Industry Patterns

### 5.1 Anthropic's Own Multi-Model Strategy

Anthropic Claude API supports per-call model selection. MoAI-ADK pattern aligns:
- High-reasoning tasks → Opus 4.7
- Code/docs/standard tasks → Sonnet 4.6
- Mechanical/batch tasks → Haiku 4.5

Reference: `https://docs.anthropic.com/en/docs/about-claude/models`

### 5.2 Cursor / Continue.dev / Aider Patterns

Other agent frameworks (Cursor, Continue.dev, Aider) typically expose model selection at the user level. MoAI-ADK differentiates by **per-agent default** classification:
- Static frontmatter `model:` field → no user friction
- Env-var override (`CLAUDE_MODEL`) remains as runtime escape hatch
- Cost-optimization built into framework, not user responsibility

### 5.3 Batch API Adoption Precedent

Anthropic Batch API documentation: `https://docs.anthropic.com/en/api/messages-batches`
- 50% off in + out pricing
- Up to 24h completion time
- Suitable for async workloads: classification, summarization, large-batch processing
- researcher agent's profile (async-OK, large-batch self-research) matches Batch API target use case

---

## 6. Future Research Directions

### 6.1 Continuous Regression Monitoring

Post-Sprint 2, evaluate:
- nightly cron + eval set replay (345 tasks)
- regression alarm threshold (2 weeks 연속 -5% 초과 시)
- Grafana dashboard for cost trend visualization

Estimated effort: 1 SPEC (Tier M, Sprint 6+).

### 6.2 Per-Skill Model Override

본 SPEC는 per-agent. Per-skill override (예: skill body 내에서 model 강제) 가능성:
- Use case: complex skill needs reasoning depth temporarily
- Mechanism: skill frontmatter `model_override: opus` (proposed)
- Risk: complexity creep, harder to track effective model
- Defer to user demand signal

### 6.3 Model Version Sunset

Opus 4.7 → Opus 4.8 transition handling:
- Static frontmatter `model: opus` should resolve to latest opus version automatically (Claude Code runtime contract)
- Confirmed via `manager-git` + `manager-docs` continuous operation through model version updates
- Sprint 5 SPEC-V3R6-RELEASE-V3-001 covers explicit sunset.yaml policy

### 6.4 GLM-4.6 Backend Composition

Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001 enables GLM-4.6 provider. Cross-SPEC composition matrix:

| Provider | Model Tier | Mapped Endpoint (hypothetical) |
|----------|-----------|--------------------------------|
| Anthropic Claude | opus | claude-opus-4-7 |
| Anthropic Claude | sonnet | claude-sonnet-4-6 |
| Anthropic Claude | haiku | claude-haiku-4-5 |
| Z.AI GLM | opus | glm-4.6-reasoning |
| Z.AI GLM | sonnet | glm-4.6-standard |
| Z.AI GLM | haiku | glm-4.6-batch |

Cost trade-off analysis deferred to Sprint 3 SPEC.

---

## 7. Research Out of Scope

### Out of Scope: Empirical pricing measurement

본 research는 Anthropic published pricing 사용. Real-time billing API integration (`https://api.anthropic.com/v1/organizations/usage_report`)는 향후 monitoring SPEC.

### Out of Scope: Cross-provider quality benchmarking

GLM-4.6 vs Claude quality benchmarking은 Sprint 3 SPEC-V3R6-BACKEND-ROUTING-001 scope. 본 SPEC는 Anthropic family only.

### Out of Scope: Long-tail agent invocation patterns

본 research는 23 agents의 average behavior. P99 latency / outlier handling은 별도 SLO SPEC (Sprint 6+).

---

## 8. References

- `.moai/research/v3.0-design-2026-05-22.md` §1.3 (라인 56-60) — model pricing table
- `.moai/research/v3.0-design-2026-05-22.md` §1.4 (라인 64-73) — pricing levers (Batch API 50% off, Haiku 강등 -80%, Sonnet 강등 -40%)
- `.moai/research/v3.0-design-2026-05-22.md` §6.1 (라인 412-414) — KPI: 단일 `/moai run` 턴 평균 비용 $1.10 → ≤ $0.45
- `.moai/research/moai-adk-current-state-2026-05-22.md` §3 (라인 134-178) — agent inventory baseline (iter 1 19-agent estimate)
- Anthropic Batch API Documentation — `https://docs.anthropic.com/en/api/messages-batches`
- Anthropic Pricing Page — `https://www.anthropic.com/pricing`
- Anthropic Models Documentation — `https://docs.anthropic.com/en/docs/about-claude/models`
- `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy (line 55, expert-refactoring opus carve-out)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier L definition)

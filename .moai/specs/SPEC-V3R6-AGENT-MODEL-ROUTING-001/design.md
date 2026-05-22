---
id: SPEC-V3R6-AGENT-MODEL-ROUTING-001-DESIGN
title: "Design — Agent 23개 모델 명시 라우팅 (architectural rationale, alternatives, decision log)"
version: "0.2.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents"
lifecycle: spec-anchored
tags: "agent, model-routing, opus, sonnet, haiku, cost-optimization, sprint-2, v3.0, design"
tier: L
---

# Design — SPEC-V3R6-AGENT-MODEL-ROUTING-001 Tier L

본 design은 Tier L 추가 산출물 (spec/plan/acceptance 외)로서 architectural rationale + alternatives + decision log + cross-Sprint 영향을 기록한다.

---

## 1. Architectural Context

### 1.1 4-Subdirectory Layout (iter 2 정정)

`.claude/agents/` 디렉토리는 4 subdirectory로 organize된 23 agent files를 hosting한다:

```
.claude/agents/
├── core/      (8 agents)   — Cross-cutting orchestration managers
├── expert/    (6 agents)   — Domain specialists (backend, frontend, security, etc.)
├── harness/   (4 agents)   — MoAI-ADK self-development specialists (NEW SPECC subdomain)
└── meta/      (5 agents)   — Meta-level agents (researcher, evaluator, builder, etc.)
```

**Why 4-subdirectory layout (vs flat `.claude/agents/moai/`)**:

| Dimension | Flat layout (iter 1 추정) | 4-subdir layout (iter 2 actual) |
|-----------|---------------------------|--------------------------------|
| Discoverability | 19 files mixed — hard to navigate | 4 categories — clear domain boundaries |
| Tier classification | Implicit (filename prefix) | Explicit (subdirectory) |
| Cross-domain scaling | Linear filename growth | Subdomain-scoped (e.g., harness/ for MoAI internal) |
| Glob complexity | `*.md` single glob | `{core,expert,harness,meta}/*.md` 4-subdir glob |
| Refactoring cost (rename) | Low (1 directory) | Medium (relative paths in glob) |

**iter 1 inventory drift root cause**: iter 1 spec.md inherited `.moai/research/v3.0-design-2026-05-22.md` §Layer 3 (라인 213-241)의 19-agent table without verifying actual filesystem. iter 2 BLOCKING B1 forced inventory re-verification via `find .claude/agents -name "*.md" -type f` → 23 agents in 4 subdirectories.

### 1.2 Agent Frontmatter Schema (vs SPEC Frontmatter)

본 SPEC가 변경하는 agent file frontmatter는 SPEC frontmatter (12-field canonical)와 다른 schema:

```yaml
---
name: <agent-name>          # required
description: <description>  # required
model: <tier>               # OPTIONAL — currently 21/23 inherit, 2/23 haiku
tools: <CSV string>         # optional, format per agent-authoring.md
disallowed-tools: <CSV>     # optional
---
```

본 SPEC scope: `model:` field 명시 (inherit → opus | sonnet | haiku). Other fields는 PRESERVE.

---

## 2. Tier Classification Rationale (23 agents)

### 2.1 Opus Tier (7 agents) — Reasoning-Intensive

Constitution `.claude/rules/moai/core/moai-constitution.md` § Opus 4.7 Prompt Philosophy 라인 55 verbatim:

> "Effort level selection: **reasoning-intensive agents** (manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, **expert-refactoring**) → `effort: xhigh` or `high`"

**7 opus tier agents** (constitution-aligned + manager-develop implementation rationale):

| Agent | Tier | Constitution alignment | Sub-rationale |
|-------|------|----------------------|----------------|
| `manager-spec` | opus | Reasoning-intensive (verbatim) | EARS spec authoring + traceability |
| `manager-strategy` | opus | Reasoning-intensive (verbatim) | Strategic analysis + architecture trade-off |
| `plan-auditor` | opus | Reasoning-intensive (verbatim) | Independent audit + bias prevention |
| `evaluator-active` | opus | Reasoning-intensive (verbatim) | 4-axis GAN Loop evaluation |
| `expert-security` | opus | Reasoning-intensive (verbatim) | OWASP analysis depth |
| `expert-refactoring` | opus | Reasoning-intensive (verbatim) — **iter 2 S2 carve-out** | AST 변환 reasoning depth |
| `manager-develop` | opus | Implementation-intensive (`effort: high` default) | Multi-file orchestration + DDD/TDD cycle delegation |

**Why manager-develop in opus (not sonnet)**:
- Multi-file orchestration requires deep planning across 5-30 files
- DDD/TDD cycle delegation involves sub-agent reasoning (which sub-agent for which task)
- Section A-E delegation template authoring requires reasoning about cross-cutting concerns
- Currently `inherit` (default Opus 4.7) — explicit opus assignment preserves baseline behavior

### 2.2 Sonnet Tier (13 agents) — Mid-Complexity / Code / Docs / Analysis

Sonnet 4.6 is optimized for code generation, documentation, and structured analysis. 13 agents fall into this tier:

| Agent | Tier | Rationale |
|-------|------|-----------|
| `manager-brain` | sonnet | 7-phase ideation — scaffolding work |
| `manager-project` | sonnet | Project documentation + structure (generation) |
| `manager-quality` | sonnet | Quality validation (mostly rule-based, TRUST 5 matrix) |
| `expert-backend` | sonnet | API design (constitution: implementation-intensive) |
| `expert-frontend` | sonnet | UI code (constitution: implementation-intensive) |
| `expert-devops` | sonnet | CI/CD patterns |
| `expert-performance` | sonnet | Profiling analysis |
| `builder-harness` | sonnet | YAML/Markdown generation |
| `claude-code-guide` | sonnet | Claude Code regression investigation |
| `cli-template-specialist` | sonnet (**iter 2 NEW**) | CLI scaffold + template integration |
| `hook-ci-specialist` | sonnet (**iter 2 NEW**) | Hook + CI workflow specialization |
| `quality-specialist` | sonnet (**iter 2 NEW**) | Quality gate + lint specialization |
| `workflow-specialist` | sonnet (**iter 2 NEW**) | Workflow orchestration specialist |

**4 harness specialists (iter 2 inventory drift fix)**: Domain-scoped specialists for MoAI-ADK self-development. Classified as sonnet because the harness work is mid-complexity code/config generation (not reasoning-intensive evaluation or strategy).

### 2.3 Haiku Tier (3 agents) — Mechanical / Async-OK

Haiku 4.5 is optimized for mechanical operations and async batch workloads:

| Agent | Tier | Rationale | Pre-existing? |
|-------|------|-----------|---------------|
| `manager-docs` | haiku | Documentation generation (README/API/Nextra) | YES (validated) |
| `manager-git` | haiku | Git workflow (Conventional Commits mechanical) | YES (validated) |
| `researcher` | haiku + batch_api | Self-research, async-OK, large-batch processing | NEW (was `inherit`) |

**Why researcher batch_api opt-in**:
- researcher is async-OK (worktree isolation experiments, self-research, Explore agent assistant)
- Batch API: 50% off in + out pricing for async workloads (`https://docs.anthropic.com/en/api/messages-batches`)
- Combined: Haiku $1/MTok × 50% Batch = $0.50/MTok input (vs current Opus inherit $5/MTok = 90% off)

---

## 3. Alternatives Considered

### Alternative A: Tier M scope-shrink (Rejected)

**Proposal**: Keep Tier M (3 artifacts), reduce scope to "opus 7 + others stay inherit". Defer sonnet 13 + haiku 3 to Sprint 5.

**Rejected because**:
1. 7-opus-only migration doesn't achieve the v3.0 KPI ($4.20 → $2.20/MTok 48% off). Cost benefit requires the full 23-tier migration.
2. Sprint 2 sibling SPEC (PROMPT-CACHE-001) depends on full per-agent model classification — partial migration creates ambiguity in cache_write break-even per agent.
3. Tier M ceiling is 5-15 files; this SPEC affects 23 agent files + 23 mirror + 4-locale docs = **50+ files** → Tier L mandate per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier.

### Alternative B: 6-opus / 11-sonnet / 6-haiku (Rejected — design.md §Layer 3 original)

**Proposal**: `.moai/research/v3.0-design-2026-05-22.md` §Layer 3 라인 213-241 의 19-agent table directly inherit:
- 6 opus: manager-develop, manager-spec, plan-auditor, manager-strategy, expert-security, evaluator-active
- 10 sonnet: manager-brain, manager-quality, manager-project, expert-frontend, expert-backend, expert-devops, expert-performance, expert-refactoring, builder-harness, claude-code-guide
- 3 haiku: manager-git, manager-docs, researcher

**Rejected because**:
1. **expert-refactoring sonnet contradicts constitution** (`moai-constitution.md` line 55 lists expert-refactoring in reasoning-intensive). iter 2 S2 BLOCKING resolution: constitution > design.md (constitution is authoritative source for effort tier alignment).
2. **Missing 4 harness specialists**: design.md table inventory was based on the 19-agent flat layout assumption. Actual filesystem has `harness/` subdirectory with 4 specialists.
3. Cost math contradiction in design.md §Layer 3: prose said "5/19 (74% off)" but table showed 6 opus rows — internal inconsistency.

### Alternative C: All-inherit preservation (Rejected)

**Proposal**: Do not migrate. Keep 21/23 = `model: inherit`.

**Rejected because**:
1. v3.0 environmental cost-savings goal (KPI: $4.20 → $2.20/MTok 48% off) cannot be achieved without explicit tier classification.
2. `manager-git` + `manager-docs` already validated `model: haiku` works in Claude Code runtime — extension to 21 agents is mechanical.
3. researcher Batch API 90% off ($5 → $0.50/MTok) is a free win for async workloads — no caller pattern change required.

### Alternative D: Migration tool automation (Deferred to future SPEC)

**Proposal**: Build a `moai agent model set <agent> <tier>` CLI command in Go (`internal/cli/agent_model.go`) for migration + future maintenance.

**Deferred because**:
1. Adds Go code change (REQ-AMR-NF-011 forbids).
2. Manual frontmatter edit via Edit tool (6 milestones M2~M4) is sufficient for 23 agents.
3. Future maintenance frequency unknown (model versions update slowly). Tool building is overhead.
4. Defer to `SPEC-V3R7-AGENT-MIGRATION-TOOL-001` (future Sprint 7 scope).

---

## 4. Decision Log

### Decision D1: Constitution > design.md for expert-refactoring opus assignment (iter 2 S2)

**Decision**: expert-refactoring is classified as **opus tier** (not sonnet).

**Rationale**:
- `.claude/rules/moai/core/moai-constitution.md` line 55 explicitly lists "expert-refactoring" in reasoning-intensive list ("manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, **expert-refactoring**").
- AST transformation requires reasoning depth (cross-file dependency analysis, behavior preservation, edge case identification).
- constitution > design doc authority for cross-cutting model tier alignment.

**iter 1 R-AMR-006 retired**: Iter 1 noted this as a Risk (Low/Low) but only documented it as risk; iter 2 promotes to Decision (resolved via constitution authority).

### Decision D2: Tier M → Tier L reclassification (iter 2 S3)

**Decision**: Tier L (5 artifacts: spec + plan + acceptance + design + research).

**Rationale**:
- 23 agent files + 23 template mirrors + 4-locale docs-site catalog + baseline JSONL + telemetry = **50+ files** affected.
- Tier M ceiling per `.claude/rules/moai/workflow/spec-workflow.md`: 5-15 files. Tier L threshold: > 15 files.
- Tier L PASS threshold 0.85 (vs Tier M 0.80) is appropriate for the architectural scope.

### Decision D3: batch_api key resilience (iter 2 S4)

**Decision**: AC-AMR-005 accepts exactly 1 of `batch_api: true` OR `use_batch_api: true` OR `invocation_mode: batch`.

**Rationale**:
- Claude Code SDK official documentation may use any of these key conventions.
- M1 orchestrator-direct WebFetch step verifies canonical key before manager-develop frontmatter edit.
- 3-key accept reduces lock-in to specific SDK version.

### Decision D4: Sprint-Round-Milestone naming SSOT alignment (iter 2 S1)

**Decision**: All artifacts use **Sprint** (multi-SPEC) + **Milestone** (M1~M6 within-SPEC). **Wave** terminology retired per `sprint-round-naming.md` SSOT v2.0.0 AP-SRN-004.

**Rationale**:
- design.md legacy "Wave 2" reference retained as verbatim quote with parenthetical clarification.
- Frontmatter tag: `sprint-2` (NOT `wave-2`).
- Body references: "Sprint 2" (multi-SPEC group), "M1~M6" (within-SPEC milestones).

### Decision D5: 5-agent main workflow cost math preservation

**Decision**: Main `/moai run` 5-agent set membership unchanged: manager-develop (opus) + expert-backend (sonnet) + manager-quality (sonnet) + manager-git (haiku) + manager-docs (haiku).

**Rationale**:
- 5-agent set was defined in iter 1 spec.md §1.4 for cost calculation.
- iter 2 23-agent inventory adds 4 harness specialists + adds expert-refactoring opus, but 5-agent main workflow set is unchanged.
- Cost target preserved: $4.20/MTok → $2.20/MTok = 48% off.

### Decision D6: Sync-phase ordering with PROMPT-CACHE-001 (R-AMR-004)

**Decision**: AGENT-MODEL-ROUTING-001 sync-PR merges first; PROMPT-CACHE-001 sync-PR follows after 24h baseline preservation window.

**Rationale**:
- Sonnet 1h cache_write break-even (+100% / -90% hit) differs from Opus.
- Pre-cache state baseline (REQ-AMR-006 JSONL) must reflect post-AMR-migration but pre-cache-activation state for clean A/B isolation.
- 24h window ensures REQ-AMR-007 regression validation completes before cache_write activation.

---

## 5. Cross-Sprint Impact

### 5.1 Sprint 2 Sibling SPECs

| SPEC | Relationship | Coordination |
|------|--------------|--------------|
| SPEC-V3R6-PROMPT-CACHE-001 | depends_on AGENT-MODEL-ROUTING-001 | sync-phase ordering (D6) |
| SPEC-V3R6-HOOK-OBSERVE-OPT-IN-001 | orthogonal | no coordination required |
| SPEC-V3R6-HOOK-ASYNC-EXPAND-001 | orthogonal | no coordination required |

### 5.2 Sprint 3 Forward Reference

| SPEC | Relationship | Note |
|------|--------------|------|
| SPEC-V3R6-BACKEND-ROUTING-001 | orthogonal (model tier vs provider) | Cross-SPEC composition (e.g., GLM + sonnet tier) is design.md §Layer 6 future work |

### 5.3 Future SPECs (Sprint 5-7)

- **Sprint 5 SPEC-V3R6-RELEASE-V3-001**: model retirement lifecycle (sunset.yaml) deferred.
- **Sprint 7 SPEC-V3R7-AGENT-MIGRATION-TOOL-001** (proposed): `moai agent model set <agent> <tier>` CLI automation.
- **Sprint 7 SPEC-V3R7-AGENT-EVAL-FRAMEWORK-001** (proposed): Go-based eval framework for REQ-AMR-007 continuous monitoring.

---

## 6. Layer 6 — Future Work (Cross-SPEC Composition)

### 6.1 GLM + Sonnet Composition

본 SPEC는 Anthropic Claude family 내 tier 분배만. SPEC-V3R6-BACKEND-ROUTING-001 (Sprint 3)가 GLM-4.6 provider 도입 시:

```yaml
# 가설적 frontmatter
model: sonnet           # AGENT-MODEL-ROUTING-001 본 SPEC
provider: glm-4.6       # BACKEND-ROUTING-001 추가 (Sprint 3)
```

두 SPECs의 정합성은 Sprint 3 SPEC body에서 명시:
- model tier (opus/sonnet/haiku) → Claude Code runtime이 provider 별로 매핑
- 예: GLM-4.6 + model: opus → Z.AI GLM 풀의 reasoning-intensive endpoint

### 6.2 Continuous Regression Monitoring

REQ-AMR-007 ±5% bound는 1-time post-migration validation. Continuous monitoring framework은 다음 단계:
- nightly cron + eval set replay
- baseline JSONL diff per week
- regression alarm threshold (예: 2 weeks 연속 -5% 초과 시)

---

## 7. Design Out of Scope

### Out of Scope: Provider abstraction layer

본 SPEC는 frontmatter `model:` 필드만. Anthropic / Z.AI / OpenAI provider abstraction (예: provider-agnostic model identifier)는 Sprint 3+ scope.

### Out of Scope: Per-skill model override

본 SPEC는 per-agent only. Per-skill `model:` override (예: skill body 내에서 model 강제)는 별도 architecture decision.

### Out of Scope: Cost monitoring dashboard

REQ-AMR-NF-010 cost reduction은 1-time verification (grep count). Real-time cost monitoring dashboard (예: Grafana + Anthropic usage API)는 Sprint 6+ scope.

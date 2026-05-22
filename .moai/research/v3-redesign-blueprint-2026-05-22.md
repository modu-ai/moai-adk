# MoAI-ADK Harness v3.0.0 Redesign Blueprint

> **Created**: 2026-05-22
> **Source**: 2 parallel research agents (industry best practices 2026 + MoAI-ADK baseline) + 71-issue audit
> **Status**: USER-CONFIRMED blueprint, ready for Wave 1 SPEC entry
> **Replaces**: v2.20.0-rc1 baseline → v3.0.0 major release

## 0. User Decisions (2026-05-22)

| Decision | Choice |
|---|---|
| Progression order | Wave 1 → 6 sequential |
| Agent slim | 23 → **11** (8 removed) |
| Skill slim | 37 → **9** + official standard re-verification + modules restructure + error audit |
| Folder layout | `agents/{core,expert,meta,harness}/` + `skills/moai-harness-*` series |
| GEARS migration | **추가 (2026-05-22 confirm)** — Wave 6 SPEC-V3R6-GEARS-MIGRATION-001 신설 |

## 1. Industry Findings (Top 10 — 2026-05 state)

1. Skills 3-tier Progressive Disclosure is dominant primitive (`agentskills.io` Dec 2025 open standard).
2. Hooks (26 events, exit 2 = block) = deterministic control layer; production deployments use 95+ hooks.
3. Hierarchical Supervisor = +12% accuracy at 5×–15× cost; cost-aware routing required.
4. 88% agent projects fail to reach production; 65% of failures are context drift (NOT architecture).
5. EARS → GEARS evolution (drops If/Then; clarifies where vs while for AI parsing). **v3 결정 (2026-05-22)**: 명시적 마이그레이션 SPEC 추가 — `SPEC-V3R6-GEARS-MIGRATION-001` (Wave 6).
6. Effective context cap = 8K–50K tokens regardless of 1M marketing.
7. Anthropic Claude Code Review (Mar 2026): 16→54% coverage using Agent Teams of specialists.
8. Model routing = biggest ROI lever (Opus orchestrator + Sonnet workers ≈ -40% cost).
9. Voyager + Reflexion remain canonical self-evolution patterns.
10. Constitution = living + canary validation; FROZEN/EVOLVABLE split aligns with 2026 best practice.

## 2. v3 Layer Architecture

```
Layer 1: ORCHESTRATION  (always loaded)
   .claude/skills/moai/SKILL.md           ≤500 LOC
   .claude/commands/moai/*.md             13 thin wrappers ≤20 LOC
   CLAUDE.md                              ~150 LOC

Layer 2: CORE WORKFLOW  (paths-triggered)
   .claude/agents/core/                   4 (spec/develop/docs/quality)
   .claude/skills/moai-core-*             3 (spec/project/worktree)
   .claude/rules/moai/core/               6 frozen-canonical rules

Layer 3: EXPERTISE  (paths/keywords-triggered)
   .claude/agents/expert/                 4 (backend/frontend/security/git)
   .claude/agents/meta/                   3 (builder-harness/evaluator-active/plan-auditor)
   .claude/skills/moai-foundation-core    TRUST 5 + delegation (cc/thinking/quality absorbed)
   .claude/rules/moai/{workflow,development}/  8 rules

Layer 4: HARNESS  (project-specific, self-evolving)
   .claude/agents/harness/                meta-harness-generated, 4+ per project
   .claude/skills/moai-harness-*          project-specific series
   .moai/harness/                         usage-log + proposals + snapshots (V3R4 5-Layer)
```

## 3. Taxonomy (Final)

### 3.1 Agents — 11 (was 23)

| Folder | Agent | Model | Effort | Isolation |
|---|---|---|---|---|
| core/ | manager-spec | inherit | xhigh | none |
| core/ | manager-develop | inherit | xhigh | worktree |
| core/ | manager-docs | haiku | medium | none |
| core/ | manager-quality | inherit | high | none |
| expert/ | expert-backend | inherit | high | worktree |
| expert/ | expert-frontend | inherit | high | worktree |
| expert/ | expert-security | inherit | xhigh | none |
| expert/ | manager-git | haiku | medium | none |
| meta/ | builder-harness | inherit | high | none |
| meta/ | evaluator-active | inherit | xhigh | none |
| meta/ | plan-auditor | inherit | high | none |

**Removed**: manager-brain (→ spec), manager-strategy (→ spec), manager-project (→ docs+meta-harness), expert-devops (→ harness), expert-performance (→ harness), expert-refactoring (→ develop), claude-code-guide (→ dev-only), researcher (→ harness mechanism)

### 3.2 Skills — 9 (was 37)

| Skill | Layer | Role |
|---|---|---|
| moai | L1 | Router (13 workflows + 4 team) |
| moai-foundation-core | L1 | TRUST 5 + SPEC + delegation (cc/thinking/quality absorbed) |
| moai-core-spec | L2 | SPEC lifecycle (plan + run + sync) |
| moai-core-project | L2 | Project doc + harness auto-init |
| moai-core-worktree | L2 | Worktree management |
| moai-meta-harness | L2 | Harness 6-Phase generator (upstream-aligned) |
| moai-harness-learner | L2 | Tier-4 surfacing (V3R4 5-Layer) |
| moai-design | L3 bundled | brand/copy/handoff/gan-loop unified |
| moai-ref-bundle | L3 bundled | api/react/owasp/git/testing-pyramid unified |

**Removed** (28): foundation-{cc,thinking,quality}, workflow-{loop,testing,ddd,tdd,ci-autofix,ci-watch,gan-loop,design-context,design-import}, domain-{ideation,research,design-handoff,brand-design,copywriting,backend,frontend,database}, design-system, ref-{api-patterns,git-workflow,owasp-checklist,react-patterns,testing-pyramid}, my-harness-* (4 renamed → moai-harness-*)

### 3.3 Commands — 13 (was 17 production)

Production: plan, run, sync, design, project, fix, loop, mx, review, gate, harness, feedback, codemaps
Removed: brain, clean, coverage, e2e, security (absorbed into options of remaining commands)
Dev-only (3): 97-release-update, 98-github, 99-release (preserved)

## 4. Self-Evolution v2 Mechanism

5-stage cycle integrating V3R4 + Voyager + Reflexion + AgentDevel release-engineering:

```
OBSERVE (PostToolUse hook)
   → usage-log.jsonl schema expansion:
     {agent_chosen, outcome_score, regression_detected, user_correction}

REFLECT (SubagentStop hook — verbal RL)
   → .moai/harness/reflections/<date>.md

PROMOTE (4-tier ladder)
   observation(1) → heuristic(3) → rule(5) → tier-4(10)

VALIDATE (5-Layer safety)
   L1 Frozen Guard / L2 Canary (last 5 days same project)
   L3 Contradiction Detector / L4 Rate Limit (3/week, 24h cooldown)
   L5 Human Oversight (AskUserQuestion)

APPLY / AUTO-ROLLBACK
   → snapshots/<ISO>/
   → 3-invocation monitor: score drop > 0.10 → auto rollback
```

V3R4 → v2 deltas (4 enhancements):
- outcome_score auto-calculation (PostToolUse hook from lint exit + test pass + AskUser frequency)
- Reflexion verbal critique stage (uses Claude native reasoning, no extra LLM call)
- Canary baseline: `last 3 projects` → `current project last 5 days`
- Auto-rollback trigger via 3-invocation drift monitor

## 5. /moai project + Harness Integration

```
Phase 1-5: existing (codebase analysis → product/structure/tech/codemaps)
Phase 6 [new]: Harness Socratic Interview (4 rounds AskUserQuestion)
   R1: auto-install yes/no
   R2: domain class (web/CLI/library/mobile)
   R3: tech stack verification
   R4: methodology (DDD/TDD/mixed)
Phase 7 [new]: meta-harness invocation (6-Phase upstream-aligned)
Phase 8 [new]: Preview + final user approval
Phase 9 [new]: Learning activation (harness.yaml: learning.enabled: true)
```

## 6. SPEC Decomposition — 19 SPECs (V3R6 prefix)

### Wave 1: Foundation Cleanup (Tier S, parallel-safe)
- SPEC-V3R6-CATALOG-SSOT-001
- SPEC-V3R6-HARNESS-LEARNER-FIX-001
- SPEC-V3R6-ABSORB-CLEANUP-001

### Wave 2: Folder Restructure (Tier S/M)
- SPEC-V3R6-HARNESS-RENAME-001 (S — `my-harness/` → `harness/`, `my-harness-*` → `moai-harness-*`, 65 refs)
- SPEC-V3R6-AGENT-FOLDER-SPLIT-001 (M — `.claude/agents/moai/` → `core/` + `expert/` + `meta/`)
- SPEC-V3R6-META-HARNESS-PATH-001 (S — meta-harness output paths + 6-Phase upstream alignment)

### Wave 3: Slim-down (Tier M/L)
- SPEC-V3R6-AGENT-SLIM-001 (M — 19→11 agents)
- **SPEC-V3R6-SKILL-SLIM-001 (Tier L)** — 37→9 + agentskills.io standard re-verification + modules restructure + error audit (user-requested expansion)
- SPEC-V3R6-COMMAND-SLIM-001 (S — 17→13 commands)
- SPEC-V3R6-MANAGER-DECOUPLE-001 (M — categorical lazy-load of skills)
- SPEC-V3R6-RULE-CONSOLIDATE-001 (M — 51→25 rules + CLAUDE.md split)

### Wave 4: Project + Harness Auto-Integration (Tier M/S)
- SPEC-V3R6-PROJECT-HARNESS-AUTO-001 (M)
- SPEC-V3R6-SESSION-MEMORY-AUTO-001 (S — SessionStart hook auto-inject)

### Wave 5: Self-Evolution v2 (Tier M/S)
- SPEC-V3R6-HARNESS-OBS-002 (M)
- SPEC-V3R6-HARNESS-REFLECT-001 (M)
- SPEC-V3R6-HARNESS-CANARY-V2-001 (M)
- SPEC-V3R6-HARNESS-AUTOROLLBACK-001 (S)

### Wave 6: Final Compliance + Release (Tier M)
- SPEC-V3R6-RULES-COMPLIANCE-001 (M — paths CSV, Korean→EN, zone-registry CI guard)
- SPEC-V3R6-GEARS-MIGRATION-001 (M — EARS → GEARS 키워드 마이그레이션, lint.go rule 갱신, 88개 SPEC 후보 변환 가이드, 4-locale docs-site 반영) **[추가 2026-05-22 — 사용자 결정 반영]**
- SPEC-V3R6-V3-CUTOVER-001 (M — v2.20.0-rc1 → v3.0.0 release manifest)

## 7. Dependency Graph

```
Wave 1 (parallel) → Wave 2 (sequential within) → Wave 3 (sequential) → Wave 4 → Wave 5 → Wave 6
```

## 8. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| 28 skill removal breaks user workflows | M | H | absorption mapping + migration script + v2.x LTS 6mo |
| 8 agent removal impacts 88 V3 SPECs | L | M | existing V3R2/R3/R5 SPECs archived but preserved |
| 65 file refs rename slip | M | M | grep-audit + CI guard + meta-harness verify |
| Phase 6-9 lengthens /moai project | L | L | "skip" option always offered |
| Self-evolution v2 spurious proposals | M | M | conservative thresholds + L3 contradiction + canary v2 |

## 9. Sources (research)

65+ URLs from May 2026 cited in research synthesis. Key references:
- Anthropic: equipping-agents-for-the-real-world-with-agent-skills
- Anthropic: effective-context-engineering-for-ai-agents
- agentskills.io open standard (Dec 2025)
- revfactory/harness (Apache 2.0, 6-Phase methodology)
- Voyager (arXiv:2305.16291)
- Reflexion (arXiv:2303.11366)
- AgentDevel (arXiv:2601.04620)
- ICLR 2026 RSI Workshop
- EARS official guide (Alistair Mavin)
- GEARS proposal (DEV 2026)

Full source list maintained in research agent transcript memory.

## 10. Wave 1 First-SPEC Recommendation

Suggested entry: **SPEC-V3R6-HARNESS-LEARNER-FIX-001** (highest criticality — HARD rule violation, Tier S, 1 file, ~50 LOC change).

Alternative: parallel plan-phase entry for all 3 Wave 1 SPECs (all Tier S, no dependencies). Late-Branch workflow applies — commits on main, branch at PR time.

---

**Next action**: User selects Wave 1 entry mode (sequential 1 SPEC OR parallel 3 SPECs).

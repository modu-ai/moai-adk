# SPEC-AGENCY-001: AI Agency Self-Evolving System

## Metadata
- ID: SPEC-AGENCY-001
- Title: AI Agency Self-Evolving System v3.2
- Status: approved
- Created: 2026-04-02
- Priority: high
- Complexity: 9/10
- Domains: agents, skills, commands, rules, configuration
- Estimated Files: 21+

## 1. Requirements (EARS Format)

### R-001: Agency Command Entry Point
When a user invokes `/agency` with a subcommand (brief, build, review, learn, evolve, resume, profile, sync-upstream, rollback, phase, config),
the system shall route to the corresponding agency workflow,
supporting both "just do it" (auto) and step-by-step modes.

### R-002: Independent Agent Definitions
When the agency system is initialized,
the system shall provide 6 independent agent definitions under `.claude/agents/agency/`:
- `planner.md` — Expands user request to full BRIEF (forked from manager-spec + manager-strategy)
- `copywriter.md` — Writes marketing/product copy (new, no moai equivalent)
- `designer.md` — Creates design systems and UI specs (new, no moai equivalent)
- `builder.md` — Implements code (forked from expert-frontend + expert-backend)
- `evaluator.md` — Tests with Playwright, scores quality (forked from evaluator-active)
- `learner.md` — Orchestrates all evolution (new, meta-evolution agent)

### R-003: Self-Evolving Skill Modules
When a skill module is created,
the system shall structure it with Static Zone (never auto-modify, Brand Context anchored) and Dynamic Zone (evolves via feedback),
so that each skill file contains versioned Rules, Anti-Patterns, and Heuristics with individual confidence scores.

The 5 base skill modules are:
- `agency-copywriting/SKILL.md` — Copy rules, tone, structure, anti-patterns
- `agency-design-system/SKILL.md` — Visual patterns, color, typography
- `agency-frontend-patterns/SKILL.md` — Tech stack, component patterns
- `agency-evaluation-criteria/SKILL.md` — Quality weights, pass/fail thresholds
- `agency-client-interview/SKILL.md` — Discovery questions, context gathering

### R-004: Dual Zone Architecture (Agents)
When an agent definition file is created,
the system shall separate it into FROZEN Zone (identity, safety_rails, ethical_boundaries — never auto-modified) and EVOLVABLE Zone (style guidelines, output patterns, evaluation weights — auto-modified by learner agent),
with safety_rails specifying max_evolution_rate, require_approval_for, rollback_window, and frozen_sections.

### R-005: Dual Zone Architecture (Skills)
When a skill module file is created,
the system shall separate it into Static Zone (Core Principles derived from Brand Context — manual change only) and Dynamic Zone (Rules with confidence scores, Anti-Patterns, Heuristics with weights),
where each rule tracks version, confidence, evidence count, and context.

### R-006: Brand Context as Constitution
When the agency system operates,
the system shall reference `.agency/context/` files (brand-voice.md, target-audience.md, visual-identity.md, tech-preferences.md, quality-standards.md) as immutable constitution,
so that no automatic evolution can violate Brand Context principles.

### R-007: Learnings Pipeline
When user feedback is collected,
the system shall append to `.agency/learnings/learnings.md` with timestamp, content, target module, and severity,
then detect patterns (3x→Heuristic, 5x→Rule, 10x+→High-confidence, 1x critical→Anti-Pattern immediate).

### R-008: Fork Manifest & Upstream Sync
When an agent or skill is forked from moai,
the system shall register it in `.agency/fork-manifest.yaml` with upstream file path, version_at_fork, hash_at_fork, divergence_score, and sync_policy (auto-propose/manual/never),
so that `moai update` can detect upstream changes and generate sync-report.md for user review.

### R-009: moai Skill Copy Mechanism
When the agency pipeline detects a project needs a moai skill (e.g., moai-lang-python),
the system shall ask the user whether to copy it to `agency-*` namespace for self-evolution,
copy the skill with name/metadata transformation, register in fork-manifest.yaml, and initialize with empty Dynamic Zone.

### R-010: GAN Loop (Builder <-> Evaluator)
When the build pipeline executes,
the system shall run Builder and Evaluator in a contract-based loop (max 5 iterations),
where the evaluation contract defines functional_criteria, priority_criteria with weights, and overall_pass_threshold,
and escalate to user if improvement < 5% after 3 iterations.

### R-011: Pipeline Adaptation Rules
When the system accumulates project history,
the system shall adapt the pipeline via 5 rules:
- Phase Skip: 3x consecutive no-edit approval → auto-approve
- Phase Merge: parallel phases always approved together → merge
- Iteration Adjust: max_iterations = ceil(avg_revisions * 1.5)
- Phase Reorder: 2x reverse execution detected → reorder proposal
- Phase Inject: 3x same additional request → new phase proposal

### R-012: Safety 5-Layer Architecture
When evolution is proposed,
the system shall enforce:
- Layer 1: Frozen Section Guard (hash integrity)
- Layer 2: Canary Validation (dry-run on last 3 sessions)
- Layer 3: Contradiction Detection (Brand Context + inter-rule)
- Layer 4: Rate Limiting (3/week per agent, 24h cooldown)
- Layer 5: Human Oversight (approval gates, rollback)

### R-013: BRIEF Document Format
When a project brief is created,
the system shall use Goal-Outcome format with sections: Project Goal, Target Audience, Brand & Tone, Content Requirements, Technical Constraints, Deliverables, Evaluation Criteria,
stored at `.agency/briefs/BRIEF-XXX/brief.md`.

### R-014: Agency Configuration
When the agency system is configured,
the system shall read `.agency/config.yaml` for pipeline settings, GAN loop parameters, evolution thresholds, and agent model assignments.

### R-015: Knowledge Graduation Protocol
When observations accumulate in learnings,
the system shall graduate them to skill/agent definitions when minimum_observations >= 5, minimum_confidence >= 0.80, consistency >= 4/5 recent, no contradiction with existing rules, and staleness_window <= 30 days.

## 2. Acceptance Criteria

### Structural
- [ ] AC-01: `.claude/agents/agency/` contains 6 agent .md files without `agency-` prefix
- [ ] AC-02: `.claude/skills/agency-*/SKILL.md` contains 5 skill modules with proper frontmatter
- [ ] AC-03: `.claude/commands/agency/agency.md` exists with proper routing
- [ ] AC-04: `.claude/rules/agency/constitution.md` exists with safety rules
- [ ] AC-05: `.agency/config.yaml` exists with all configuration sections
- [ ] AC-06: `.agency/fork-manifest.yaml` exists with fork tracking for 3 forked agents
- [ ] AC-07: `.agency/context/` contains 5 brand context template files
- [ ] AC-08: `.agency/templates/brief-template.md` exists with BRIEF format

### Agent Quality
- [ ] AC-09: Each agent has FROZEN Zone (identity, safety_rails, ethical_boundaries)
- [ ] AC-10: Each agent has EVOLVABLE Zone with [AUTO-LEARNED] placeholder pattern
- [ ] AC-11: Forked agents reference upstream in fork-manifest.yaml
- [ ] AC-12: Agent frontmatter follows Claude Code YAML spec (tools as CSV, skills as array)

### Skill Quality
- [ ] AC-13: Each skill has Static Zone with Core Principles
- [ ] AC-14: Each skill has Dynamic Zone with Rules/Anti-Patterns/Heuristics sections (initially empty)
- [ ] AC-15: Skill frontmatter includes metadata with quoted values
- [ ] AC-16: Skills define triggers.agents mapping to agency agents

### Configuration Quality
- [ ] AC-17: config.yaml includes pipeline, gan_loop, evolution, context, models sections
- [ ] AC-18: fork-manifest.yaml tracks 3 forked agents + 1 forked skill
- [ ] AC-19: constitution.md defines 5-layer safety architecture

## 3. File Manifest

| Priority | Path | Type | Description |
|----------|------|------|-------------|
| P0 | `.agency/config.yaml` | config | Agency global settings |
| P0 | `.agency/fork-manifest.yaml` | config | Upstream fork tracking |
| P0 | `.claude/rules/agency/constitution.md` | rule | Agency governance |
| P1 | `.claude/commands/agency/agency.md` | command | /agency entry point |
| P1 | `.agency/templates/brief-template.md` | template | BRIEF document template |
| P2 | `.claude/agents/agency/planner.md` | agent | fork: manager-spec |
| P2 | `.claude/agents/agency/copywriter.md` | agent | new |
| P2 | `.claude/agents/agency/designer.md` | agent | new |
| P2 | `.claude/agents/agency/builder.md` | agent | fork: expert-frontend |
| P2 | `.claude/agents/agency/evaluator.md` | agent | fork: evaluator-active |
| P2 | `.claude/agents/agency/learner.md` | agent | new (meta-evolution) |
| P3 | `.claude/skills/agency-copywriting/SKILL.md` | skill | Copy rules |
| P3 | `.claude/skills/agency-design-system/SKILL.md` | skill | Design patterns |
| P3 | `.claude/skills/agency-frontend-patterns/SKILL.md` | skill | Frontend patterns |
| P3 | `.claude/skills/agency-evaluation-criteria/SKILL.md` | skill | Quality criteria |
| P3 | `.claude/skills/agency-client-interview/SKILL.md` | skill | Interview questions |
| P4 | `.agency/context/brand-voice.md` | context | Brand tone template |
| P4 | `.agency/context/target-audience.md` | context | Customer persona template |
| P4 | `.agency/context/visual-identity.md` | context | Visual identity template |
| P4 | `.agency/context/tech-preferences.md` | context | Tech stack template |
| P4 | `.agency/context/quality-standards.md` | context | Quality standards template |

## 4. Architecture

### Pipeline Flow
```
User Request → Planner → Copywriter ─┐
                                      ├→ Builder → Evaluator (GAN Loop) → Learner
                         Designer ────┘
```

### Evolution Flow
```
User Feedback → learnings.md → Pattern Detection → rule-candidates.md
  → Validation Gate (Brand Context check) → Approval → Skill/Agent Dynamic Zone update
```

### Upstream Sync Flow
```
moai update → fork-manifest check → 3-way diff → sync-report.md → User approval
  FROZEN: upstream priority (security patches)
  EVOLVABLE: agency evolution preserved + improvements proposed
```

## 5. Design References

This SPEC is based on the AI Agency v3.2 design developed through:
- YouTube video analysis (Anthropic Harness, Agentic OS)
- 3-team brainstorming (Architecture, Evolution Mechanism, Adaptive Workflow)
- 4 design iterations (v1 → v2 → v3 → v3.1 → v3.2)

Key design documents referenced in this conversation:
- v3.2 Final Design: Agent naming, memory deletion, moai skill copy mechanism
- v3 Brainstorm Results: Self-evolving architecture, skill evolution pipeline, adaptive workflow
- GAN Loop Design: Contract-based evaluation, leniency prevention
- Safety Architecture: 5-layer protection, graduation protocol

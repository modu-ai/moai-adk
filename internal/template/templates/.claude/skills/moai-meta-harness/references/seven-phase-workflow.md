# 7-Phase Workflow — Detailed Walkthrough + 6 Architectural Patterns

Detailed per-phase activity, inputs, outputs, and architectural pattern selection for the meta-harness 7-Phase workflow.

## 6 Architectural Patterns (Upstream)

Selection guidance for Phase 3 (Synthesis):

| Pattern | Use case |
|---------|----------|
| Pipeline | Linear flow A → B → C, each agent transforms output of previous |
| Fan-out / Fan-in | One agent dispatches to N parallel workers, then aggregates |
| Expert Pool | Multiple agents compete or vote on best answer |
| Producer-Reviewer | One agent produces, another validates with veto power |
| Supervisor | Orchestrator coordinates worker agents with task assignment |
| Hierarchical Delegation | Multi-level tree (manager → leads → workers) |

## Phase 1 — Discovery

`manager-spec` conducts a 16-question Socratic interview (owned by the project-harness generation policy). The interview surfaces:

- Project domain (e.g., fintech, e-commerce, IoT)
- Primary languages and frameworks
- Team size and expertise level
- Quality, security, and performance priorities

Output: `answers.yaml` written to `.moai/harness/answers.yaml`.

## Phase 2 — Analysis

`manager-spec` and `manager-strategy` scan the repository:

- File structure patterns → infer domain boundaries
- Existing agents/skills → avoid duplication
- `go.mod`, `package.json`, `requirements.txt` → detect stack
- Test coverage baseline → set quality targets

Output: Analysis report passed to Phase 3.

## Phase 3 — Synthesis

`manager-spec` synthesizes the analysis into a SPEC document with EARS requirements. The SPEC specifies:

- Which of the 6 architectural patterns fits the domain
- Agent roles and their tool permissions
- Skill categories to generate
- Acceptance criteria for Phase 6 evaluation

## Phase 4 — Skeleton

This skill (`moai-meta-harness`) generates the harness skeleton:

1. Read `answers.yaml` and the SPEC document
2. Write `.moai/harness/main.md` — the harness entry point
3. Write extension files: `.moai/harness/agents.md`, `.moai/harness/skills.md`
4. Create agent definition stubs in `.claude/agents/harness/`

Agents involved: `builder-harness` (artifact_type=agent or artifact_type=skill) for artifact generation.

## Phase 5 — Customization

This skill fills the skeleton with domain-specific content:

1. Generate agent definitions (`.claude/agents/harness/*.md`) referencing existing MoAI agents: `manager-spec`, `manager-strategy`, `manager-develop` (`cycle_type=tdd` or `cycle_type=ddd` per `quality.yaml` `development_mode`), `manager-quality` (diagnostic-mode replaces former expert-debug), `manager-docs`, `manager-git`, `manager-brain`, `expert-backend`, `expert-frontend`, `expert-security`, `expert-refactoring`, `expert-performance`, `expert-devops`, `builder-harness` (use `artifact_type=agent|skill|plugin`), `sync-auditor`, `plan-auditor`.
2. Generate domain skills (`.claude/skills/harness-*/SKILL.md`) following the skill-authoring.md schema with `harness-*` prefix.
3. All artifacts are user-owned and never overwritten by `moai update`.

## Phase 6 — Evaluation

`sync-auditor` runs the Sprint Contract protocol (design constitution §11.5) against generated artifacts:

- Functionality: Do agents execute their stated purpose?
- Security: No credential leaks, safe tool permissions?
- Craft: YAML frontmatter valid, CSV allowed-tools, progressive disclosure?
- Consistency: Brand/domain alignment, naming conventions followed?

Pass threshold: 0.75 (configurable via `design.yaml pass_threshold`; FROZEN floor: 0.60 per design constitution §2).

## Phase 7 — Iteration

Owned by the harness-learning policy. Out of scope for this skill. The evolution mechanism captures scoring deltas from Phase 6 and feeds them back to Phase 4/5 on next harness run.

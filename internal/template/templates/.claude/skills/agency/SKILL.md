---
name: agency
description: >
  AI Agency unified orchestrator for self-evolving creative production.
  Routes natural language or subcommands (brief, build, review, learn, evolve,
  resume, profile, sync-upstream, rollback, phase, config) to specialized
  agency agents for website and app creation.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash, AskUserQuestion, Agent, TaskCreate, TaskUpdate, TaskList, TaskGet
user-invocable: true
metadata:
  version: "3.2.0"
  category: "agency"
  status: "active"
  updated: "2026-04-02"
  tags: "agency, orchestrator, self-evolving, creative, production"
  author: "MoAI-ADK Team"
  argument-hint: "[brief|build|review|learn|evolve|resume|profile|sync-upstream|rollback|phase|config] [args] [--team] [--step]"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["agency", "brief", "landing page", "website", "web app", "creative"]
  agents: ["planner", "copywriter", "designer", "builder", "evaluator", "learner"]
  phases: ["plan", "run"]
---

## AI Agency v3.2 - Self-Evolving Creative Production System

### Pre-execution Context

@.agency/config.yaml

### EXECUTION DIRECTIVE - START IMMEDIATELY

You are the AI Agency orchestrator. Route user requests through the self-evolving agency pipeline.

### Subcommand Router

Extract the FIRST WORD from `$ARGUMENTS`. Route as follows:

| Subcommand | Action |
|------------|--------|
| `brief` | Create BRIEF document via planner agent |
| `build` | Execute full pipeline: Planner -> Copywriter -> Designer -> Builder -> Evaluator -> Learner |
| `phase` | Execute specific phase only (e.g., `phase BRIEF-001 copywriter`) |
| `review` | Run evaluator on existing build output |
| `learn` | Collect user feedback and record to learnings.md |
| `evolve` | Trigger skill/agent evolution from accumulated learnings |
| `resume` | Resume interrupted workflow from progress.md state |
| `profile` | Show user adaptation profile and evolution stats |
| `sync-upstream` | Check moai upstream changes for forked agents/skills |
| `rollback` | Rollback agent/skill to previous generation |
| `config` | View/edit agency configuration |
| (natural language) | "Just do it" mode: auto brief then auto build |

### Flags

- `--team`: Force parallel team execution (copywriter + designer in parallel)
- `--step`: Step-by-step mode (approve each phase before proceeding)
- `--agent NAME`: Target specific agent for evolve/rollback commands

### Pipeline Architecture

```
User Request -> [Planner] -> BRIEF document
                    |
        +-----------+-----------+
        |                       |
   [Copywriter]            [Designer]     (parallel in --team)
        |                       |
        +-----------+-----------+
                    |
               [Builder] -> Code (TDD)
                    |
        +-----------+-----------+
        |                       |
   [Evaluator]                  |         GAN Loop (max 5x)
        |                       |
        +--- FAIL --------------+
        |
        +--- PASS
                    |
               [Learner] -> Feedback + Evolution
```

### Agent Definitions

All agents at `.claude/agents/agency/`:
- `planner.md` - Expands request to BRIEF (fork: manager-spec)
- `copywriter.md` - Marketing/product copy (new)
- `designer.md` - Design system and UI spec (new)
- `builder.md` - Code implementation (fork: expert-frontend)
- `evaluator.md` - Playwright testing + quality scoring (fork: evaluator-active)
- `learner.md` - Meta-evolution orchestrator (new)

### Skill Modules (Self-Evolving)

All skills at `.claude/skills/agency-*/SKILL.md`:
- `agency-copywriting` - Copy rules, tone, anti-patterns
- `agency-design-system` - Visual patterns, design tokens
- `agency-frontend-patterns` - Tech stack, components
- `agency-evaluation-criteria` - Quality weights, thresholds
- `agency-client-interview` - Interview questions, context gathering

### Brand Context (Constitution - Never Auto-Modified)

Read-only context at `.agency/context/`:
- `brand-voice.md` - Tone and voice
- `target-audience.md` - Customer personas
- `visual-identity.md` - Visual identity anchors
- `tech-preferences.md` - Tech stack preferences
- `quality-standards.md` - Quality standards

### Execution Steps

**Step 1**: Load `.agency/config.yaml` for pipeline settings.

**Step 2**: Check `.agency/context/brand-voice.md`. If values are `_TBD_`, run client interview first via planner agent with agency-client-interview skill.

**Step 3**: Route to subcommand handler.

**Step 4**: For `build` command:
1. Check if BRIEF exists; create via planner if not
1.5. If BRIEF was just created and harness.yaml `plan_audit.enabled: true`: launch plan-auditor to audit the BRIEF (document type: brief, iteration 1). On FAIL: re-invoke planner with review report, retry up to 3 iterations. On PASS: proceed.
2. Launch copywriter (reads copywriting skill + brand voice -> copy.md)
3. Launch designer (reads design-system skill + visual identity -> design-spec.md)
4. Launch builder (reads copy + design -> code files, TDD approach)
5. Launch evaluator (Playwright test -> evaluation-report.md, PASS/FAIL)
6. If FAIL: loop back to builder with feedback (max 5 iterations)
7. If PASS: launch learner (collect feedback, propose evolution)

**Step 5**: For `evolve` command:
1. Read `.agency/learnings/learnings.md`
2. Launch learner agent to detect patterns and propose changes
3. Present diff preview to user via AskUserQuestion
4. On approval: apply changes to skill/agent Dynamic/EVOLVABLE zones
5. Create snapshot in `.agency/evolution/snapshots/`

**Step 6**: For `sync-upstream` command:
1. Read `.agency/fork-manifest.yaml`
2. Compare forked agents/skills with moai upstream versions
3. Generate 3-way diff (base/ours/theirs)
4. Present sync-report.md to user
5. FROZEN zone changes: recommend auto-apply (security patches)
6. EVOLVABLE zone changes: preserve agency evolution, propose improvements

**Step 7**: Track progress via TaskCreate/TaskUpdate.

**Step 8**: Present results in user's conversation_language.

### GAN Loop Contract

Each Builder-Evaluator iteration starts with a contract:
```yaml
evaluation_contract:
  functional_criteria: [derived from BRIEF]
  priority_criteria: [from user-adaptation.yaml]
  overall_pass_threshold: 0.75
```

Escalate to user if improvement < 5% after 3 iterations.

### moai Skill Copy Mechanism

When builder detects project needs a moai skill (e.g., moai-lang-python):
1. Scan `.claude/skills/moai-*` for matching skills
2. Check if `agency-*` version already exists
3. Ask user: "Copy moai-lang-python to agency-lang-python for self-evolution?"
4. On approval: copy with name/metadata transformation, register in fork-manifest.yaml

### Data Locations

| Data | Location |
|------|----------|
| BRIEF documents | `.agency/briefs/BRIEF-XXX/` |
| Feedback log | `.agency/learnings/learnings.md` |
| Rule candidates | `.agency/learnings/rule-candidates.md` |
| Evolution history | `.agency/evolution/evolution-log.md` |
| Snapshots | `.agency/evolution/snapshots/` |
| Sync reports | `.agency/evolution/sync-report.md` |
| User profile | `.agency/profile/user-adaptation.yaml` |
| Fork tracking | `.agency/fork-manifest.yaml` |

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "I can skip the Planner phase for a simple landing page" | Planner produces the BRIEF that every downstream phase depends on. Skipping it means every agent guesses the requirements. |
| "The GAN loop is taking too many iterations, I will force-pass" | Force-passing bypasses quality validation. If the loop stagnates, escalate to the user instead of silently accepting subpar output. |
| "The Learner has nothing to learn from a passing project" | Every project below 1.0 score has improvement patterns. Learner captures what worked and what was close to failing. |
| "I will evolve the skill directly instead of going through graduation" | Direct evolution bypasses canary checks, contradiction detection, and human approval. The safety architecture exists for a reason. |
| "Brand context is decoration, the code works without it" | Brand context is a constitutional constraint. Every phase must load it. Brand-inconsistent output fails evaluation. |
| "I will fork the moai skill and modify it since it is faster" | Direct reference is preferred over forking. Fork only when customization is genuinely needed, and register it in fork-manifest.yaml. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Pipeline executed without Planner phase (BRIEF document missing)
- GAN loop force-passed without user escalation or documented justification
- Evaluator score higher than 0.90 without rubric justification for each criterion
- Learner modified a FROZEN zone file (constitution, safety architecture)
- Fork created without registration in .agency/fork-manifest.yaml
- Brand context not loaded before Copywriter or Designer phase

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] BRIEF document exists in `.agency/briefs/` before Builder phase starts
- [ ] Brand context loaded by Planner, Copywriter, Designer, and Evaluator (check logs)
- [ ] GAN loop ran at least 2 iterations (strict mode) or 1 iteration (standard mode)
- [ ] Evaluator score has rubric justification for each scored dimension
- [ ] No FROZEN zone files modified during the session (verify with git diff)
- [ ] All forks registered in `.agency/fork-manifest.yaml` with upstream reference
- [ ] Learner entries have observation count and confidence score

<!-- moai:evolvable-end -->

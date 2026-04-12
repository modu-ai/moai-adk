---
name: moai-workflow-research
description: >
  Self-research workflow for optimizing moai-adk components through
  binary eval experimentation loops. Adapted from autoresearch pattern
  with 5-layer safety architecture.
user-invocable: false
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "experimental"
  updated: "2026-04-09"
  tags: "research, eval, experiment, self-improvement, autoresearch"

progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

triggers:
  keywords: ["research", "eval", "experiment", "optimize skill", "improve agent"]
  agents: ["researcher"]
  phases: ["research"]
---

# Research Workflow

## Purpose

Optimize moai-adk components (skills, agents, rules, config) through iterative binary eval experimentation.

## Data Locations

| Data | Location |
|------|----------|
| Eval suites | `.moai/research/evals/{type}/{name}.eval.yaml` |
| Baselines | `.moai/research/baselines/{name}.baseline.json` |
| Experiments | `.moai/research/experiments/{name}/exp-NNN.json` |
| Changelogs | `.moai/research/experiments/{name}/changelog.md` |
| Observations | `.moai/research/observations/` |
| Dashboard | `moai research status` (CLI) |

## Eval Suite Schema

```yaml
target:
  path: .claude/skills/moai-lang-go/SKILL.md
  type: skill
test_inputs:
  - name: scenario-name
    prompt: "Test prompt for evaluation"
evals:
  - name: criterion-name
    question: "Binary yes/no question?"
    pass: "What yes looks like"
    fail: "What triggers no"
    weight: must_pass  # or nice_to_have
settings:
  runs_per_experiment: 3
  max_experiments: 20
  pass_threshold: 0.80
  target_score: 0.95
```

## Safety Layers

1. **FrozenGuard**: Constitution files cannot be modified
2. **Worktree Sandbox**: All experiments in isolated worktrees
3. **Canary Regression**: Proposed changes tested against baselines
4. **Rate Limiter**: Max experiments per session/week
5. **Human Approval**: Required before merging to main

<!-- moai:evolvable-start id="rationalizations" -->
## Common Rationalizations

| Rationalization | Reality |
|---|---|
| "The experiment results are obvious, I do not need a baseline" | Without a baseline, you cannot distinguish improvement from noise. Always measure before and after. |
| "This change is clearly better, I can skip the canary check" | Canary checks catch regressions the author cannot predict. Clear-to-author is not clear-to-system. |
| "I will merge the research branch now and validate later" | Unvalidated merges bypass every safety layer. Validate in the worktree before proposing a merge. |
| "The rate limiter is too restrictive for rapid experimentation" | Rate limits prevent runaway self-modification. If you need more experiments, batch them. |
| "I can modify the constitution to improve it" | Constitution files are FROZEN. Propose amendments through the human approval channel. |

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="red-flags" -->
## Red Flags

- Experiment branch merged without canary regression results
- Baseline measurement missing from the experiment report
- Multiple experiments run in the same session exceeding the rate limit
- Worktree not used for experimental changes (changes made in main workspace)
- Constitution or frozen-zone file modified during a research session

<!-- moai:evolvable-end -->

<!-- moai:evolvable-start id="verification" -->
## Verification

- [ ] Experiment conducted in an isolated worktree (show worktree path)
- [ ] Baseline measurement recorded before any changes
- [ ] Canary regression check passed (show before/after comparison)
- [ ] Rate limit not exceeded (check experiment count in session)
- [ ] Human approval obtained before merge proposal
- [ ] No frozen-zone files were modified (verify with git diff --name-only)

<!-- moai:evolvable-end -->

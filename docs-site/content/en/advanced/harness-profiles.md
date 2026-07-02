---
title: Harness Profiles & Evaluation
weight: 75
draft: false
---
# Harness Profiles and Evaluation System


An adaptive quality-validation system built on 3 harness levels and 4 evaluation dimensions.

## Overview

MoAI-ADK's Harness is a **3-tier adaptive quality-validation system**. It automatically
adjusts validation depth based on the SPEC's complexity. The sync-auditor agent performs
an independent, skeptical quality assessment using 4-dimension scoring.

## 3 Harness Levels

| Level | Description | When applied | sync-auditor |
|------|------|----------|-----------------|
| **minimal** | Fast validation | Simple changes (typos, config edits) | Can be skipped |
| **standard** | Standard quality validation | Most tasks | Optional |
| **thorough** | Full validation + TRUST 5 | Complex SPECs, large-scale changes | Required |

The harness level is determined automatically by the **Complexity Estimator** based on
the SPEC's scope.

## 4-Dimension Scoring

sync-auditor scores across 4 dimensions:

| Dimension | Description | Must-Pass by default |
|------|------|---------------|
| **Functionality** | Functional completeness — did it achieve its intended purpose | Yes |
| **Security** | Security — OWASP, authentication, authorization, input validation | Yes |
| **Craft** | Code quality — readability, structure, test coverage | No |
| **Consistency** | Consistency — adherence to project rules and code style | No |

### Score Range

Each dimension receives a score from 0.0 to 1.0.

### Rubric Anchors

Every evaluation criterion has a 4-tier rubric anchor:

| Score | Level | Meaning |
|------|------|------|
| 0.25 | Below standard | Basic requirements not met |
| 0.50 | Partial | Partially met, needs improvement |
| 0.75 | Meets standard | Mostly met, minor improvements needed |
| 1.00 | Excellent | Fully meets every criterion |

## Evaluation Profiles

4 profiles are provided under `.moai/config/evaluator-profiles/`:

| Profile | Description | Best for |
|--------|------|------------|
| `default.md` | Balanced default profile | Most tasks |
| `strict.md` | Strict criteria | Security-critical work |
| `lenient.md` | Lenient criteria | Prototyping |
| `frontend.md` | Frontend-specific | UI/UX work |

## Preventing Evaluator Bias (5 Mechanisms)

5 mechanisms work together to prevent evaluator leniency:

| # | Mechanism | Description |
|---|---------|------|
| 1 | **Rubric Anchoring** | Scores must be justified against the rubric |
| 2 | **Regression Baseline** | Detects an excessive score increase compared to prior projects |
| 3 | **Must-Pass Firewall** | Required criteria cannot be compensated for by scores in other dimensions |
| 4 | **Independent Re-evaluation** | Every 5th evaluation is independently re-scored (recalibrated when the deviation exceeds 0.10) |
| 5 | **Anti-Pattern Cross-Check** | Caps the affected dimension's score at 0.50 when a known anti-pattern is detected |

## Evaluator Memory Scope

The evaluator's judgment memory is **ephemeral per iteration**. In each iteration of
the GAN Loop, sync-auditor restarts with a fresh context, and the rationale from the
previous iteration's judgment is not included in the new prompt. Only the Sprint
Contract state persists across iterations.

## Configuration

Configure this in `.moai/config/sections/harness.yaml`:

```yaml
harness:
  level: auto              # auto | minimal | standard | thorough
  evaluator:
    memory_scope: per_iteration   # FROZEN — cannot be changed
    profiles:
      default: .moai/config/evaluator-profiles/default.md
      strict: .moai/config/evaluator-profiles/strict.md
    aggregation: min              # min | mean
    must_pass_dimensions:
      - Functionality
      - Security
```

## Related Documentation

- [Harness Engineering](/core-concepts/harness-engineering) — Overview of the harness concept
- [TRUST 5 Quality](/core-concepts/trust-5) — The 5 quality criteria
- [Constitution System](/core-concepts/constitution) — FROZEN/Evolvable rules
- GAN Loop — Design quality validation iteration (GAN Loop is an iterative validation pattern using an adversarial evaluator-discriminator loop for quality improvement)

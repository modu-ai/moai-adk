---
title: Harness Profiles and Evaluation System
weight: 75
draft: false
---

Applying the same depth of verification to every change wastes tokens, and flattening verification to a uniformly shallow level lets quality leak. MoAI-ADK's answer is **adaptive verification** — automatically adjusting verification depth to the complexity of the SPEC, and entrusting evaluation to an independent evaluator rather than the party that built the change.

## Overview

The MoAI-ADK harness is a **3-level adaptive quality verification system**. It automatically adjusts verification depth according to SPEC complexity, and the sync-auditor agent performs an independent, skeptical quality assessment with 4-dimension scoring. Completion is judged by scores and evidence, not by "it seems done."

## The 3 Harness Levels

| Level | Description | When applied | sync-auditor |
|------|------|----------|-----------------|
| **minimal** | Fast validation | Simple changes (typos, config edits) | Can be skipped |
| **standard** | Default quality checks | Most work | Optional |
| **thorough** | Full verification + TRUST 5 | Complex SPECs, large-scale changes | Required |

The harness level is determined automatically by the **Complexity Estimator** based on SPEC scope. Not running thorough verification on a typo fix — that in itself is tokenomics.

## 4-Dimension Scoring

The sync-auditor scores four dimensions.

| Dimension | Description | Must-Pass by default |
|------|------|---------------|
| **Functionality** | Functional completeness — does it achieve the intended purpose | Yes |
| **Security** | Security — OWASP, authentication, authorization, input validation | Yes |
| **Craft** | Code quality — readability, structure, test coverage | No |
| **Consistency** | Consistency — adherence to project rules and code style | No |

### Score Range

Each dimension receives a score from 0.0 to 1.0.

### Rubric Anchors

So that scores do not sway with the evaluator's mood, every evaluation criterion has 4-level rubric anchors.

| Score | Level | Meaning |
|------|------|------|
| 0.25 | Below bar | Basic requirements not met |
| 0.50 | Partial | Partially met, improvement needed |
| 0.75 | Met | Mostly met, minor improvements |
| 1.00 | Excellent | All criteria fully met |

## Evaluation Profiles

Four profiles are provided in `.moai/config/evaluator-profiles/`. You can change the strictness of the evaluation criteria to match the nature of the work.

| Profile | Description | Best suited for |
|--------|------|------------|
| `default.md` | Balanced default profile | Most work |
| `strict.md` | Strict criteria | Security-critical work |
| `lenient.md` | Lenient criteria | Prototyping |
| `frontend.md` | Frontend-specialized | UI/UX work |

## Evaluator Bias Prevention (5 Mechanisms)

Left unattended, LLM evaluators tend to drift toward leniency. Five mechanisms work together to structurally suppress this.

| # | Mechanism | Description |
|---|---------|------|
| 1 | **Rubric anchoring** | Every score requires a rubric justification |
| 2 | **Regression baseline** | Detects excessive score inflation relative to prior projects |
| 3 | **Must-Pass firewall** | Mandatory criteria cannot be compensated by scores in other areas |
| 4 | **Independent re-evaluation** | Independent re-evaluation every 5th run (recalibration when deviation > 0.10) |
| 5 | **Anti-pattern cross-check** | When a known anti-pattern is found, the affected dimension is capped at 0.50 |

## Evaluator Memory Scope

The evaluator's judgment memory is **transient per iteration**. In each iteration of the GAN Loop, the sync-auditor restarts with a fresh context, and the judgment rationale from the previous iteration is not included in the new prompt. Only the Sprint Contract state persists across iterations. This design prevents the evaluator from anchoring on its own prior judgments and scoring by inertia.

## Configuration

Configured in `.moai/config/sections/harness.yaml`.

```yaml
harness:
  default_profile: "default"        # default for SPECs without an evaluator_profile
  evaluator:
    memory_scope: per_iteration     # FROZEN — do not change
  mode_defaults:
    solo: auto                      # sub-agent mode: auto-detect
    team: auto                      # team mode: auto-detect
    cg: thorough                    # CG mode: always thorough
  auto_detection:
    enabled: true
    rules:
      minimal:
        conditions:
          - "file_count <= 3 AND single_domain"
      thorough:
        conditions:
          - "security_keywords OR payment_keywords present"
  escalation:
    enabled: true
    max_escalations: 2
  effort_mapping:
    minimal:  "low"
    standard: "medium"
    thorough: "high"
  levels:
    thorough:
      evaluator: true
```

## Related Documents

- [Harness Engineering](/en/core-concepts/harness-engineering) — harness concept overview
- [TRUST 5 Quality](/en/core-concepts/trust-5) — the five quality criteria
- [Constitution System](/en/core-concepts/constitution) — FROZEN/Evolvable rules
- GAN Loop — iterative design-quality verification (the GAN Loop is an adversarial evaluator-discriminator loop, an iterative verification pattern for quality improvement)

---
name: moai-workflow-gan-loop
description: >
  Builder-Evaluator GAN loop workflow for iterative design quality improvement.
  Implements Sprint Contract negotiation, 4-dimension scoring (Design Quality,
  Originality, Completeness, Functionality), stagnation detection, and
  escalation protocol. Reads parameters from design.yaml.
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read, Write, Edit, Grep, Glob, Bash
user-invocable: false
metadata:
  version: "1.0.0"
  category: "workflow"
  status: "active"
  updated: "2026-04-20"
  tags: "gan loop, builder evaluator, sprint contract, scoring, quality, iterative"
  related-skills: "moai-domain-brand-design, moai-domain-copywriting, moai-foundation-quality"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords: ["gan loop", "builder evaluator", "quality score", "pass threshold", "sprint contract", "iterative review", "design quality"]
  agents: ["evaluator-active", "expert-frontend"]
  phases: ["run"]
---

# moai-workflow-gan-loop

Implements the Builder-Evaluator GAN loop for iterative design quality improvement. Absorbed from agency constitution Section 11 and Section 12. Integrates Sprint Contract Protocol, 4-dimension scoring, stagnation detection, and Evaluator Leniency Prevention.

All loop parameters are read from `.moai/config/sections/design.yaml`. Do not hardcode thresholds.

---

## Quick Reference

### Loop Parameters (from design.yaml)

```
design.gan_loop:
  max_iterations: 5          # Maximum Builder-Evaluator cycles
  pass_threshold: 0.75       # Score >= this value to exit loop
  escalation_after: 3        # Escalate to user after N iterations without passing
  improvement_threshold: 0.05  # Minimum score delta per iteration
  strict_mode: false         # If true, each dimension must pass individually
  sprint_contract:
    enabled: true
    required_harness_levels: [thorough]
    optional_harness_levels: [standard]
    artifact_dir: ".moai/sprints"
    max_negotiation_rounds: 2
```

### 4-Dimension Scoring Weights

| Dimension | Weight | Description |
| --- | --- | --- |
| Design Quality | 30% | Visual consistency, brand token compliance, WCAG AA |
| Originality | 25% | Not generic, not AI-slop, unique brand expression |
| Completeness | 25% | All BRIEF sections present, copy matches contract |
| Functionality | 20% | Responsive, accessible, all interactions work |

Overall score = weighted average of all four dimensions.

Pass condition: `overall_score >= pass_threshold` AND (if `strict_mode: true`) each dimension score >= `pass_threshold`.

---

## Implementation Guide

### GAN Loop Execution Flow

**Phase 1: Sprint Contract (when required by harness level)**

Required when `harness_level == thorough`.
Optional when `harness_level == standard` and user opts in.
Skipped when `harness_level == minimal`.

Sprint Contract generation:
1. Evaluator analyzes the BRIEF document and current iteration scope.
2. Evaluator produces the Sprint Contract document:
   - `acceptance_checklist`: concrete, testable criteria for this iteration
   - `priority_dimension`: which of the 4 dimensions to focus on
   - `test_scenarios`: specific verification steps
   - `pass_conditions`: minimum score per criterion
3. Builder reviews the contract:
   - Accept: proceed with implementation
   - Request adjustment: propose alternatives (max `max_negotiation_rounds` rounds)
4. Contract is saved to `design.gan_loop.sprint_contract.artifact_dir/sprint-N.json`

Constraint: Evaluator must not score on criteria outside the Sprint Contract. Builder must not claim criteria as met without evidence.

**Phase 2: Builder Execution**

Builder implements based on:
- Accepted Sprint Contract (if present)
- BRIEF document
- Copy JSON from `moai-domain-copywriting`
- Design tokens from `moai-domain-brand-design` or `moai-workflow-design-import`

Builder outputs: code files, rendered previews (if Playwright available), implementation notes.

**Phase 3: Evaluator Scoring**

Evaluator scores against the 4 dimensions using the Evaluator Leniency Prevention mechanisms:

1. **Rubric Anchoring**: Score each dimension against the rubric (0.25 increments) with explicit justification. Scores without rubric reference are invalid.
2. **Evidence-Only Verdicts**: No PASS without concrete evidence (screenshot, test output, code reference).
3. **Anti-Pattern Cross-check**: Check known anti-patterns before finalizing. Any detected anti-pattern caps the relevant dimension score at 0.50.
4. **Must-Pass Firewall**: Copy integrity, mobile viewport, and WCAG AA are must-pass criteria. Failure in any must-pass = overall FAIL regardless of other scores.

Output: `evaluation-report-N.json` in `sprint_contract.artifact_dir`.

**Phase 4: Loop Decision**

```
if overall_score >= pass_threshold:
    EXIT LOOP → proceed to next phase
elif iteration >= max_iterations:
    ESCALATE → present failure report to user
elif stagnation_detected:
    ESCALATE → present stagnation options
else:
    ITERATE → pass feedback to Builder, increment N
```

**Phase 5: Iteration Feedback**

If looping back:
1. Evaluator generates targeted feedback per failed criterion.
2. Builder receives the feedback and previous Sprint Contract.
3. Previously passed criteria carry forward (no regression allowed).
4. New Sprint Contract is generated for failed criteria only.

---

### Stagnation Detection

Stagnation is detected when the score improvement between consecutive iterations is below `improvement_threshold` for 2 or more iterations.

Tracking:
- After each iteration, record `{iteration: N, score: X}` in the sprint artifact.
- Calculate `delta = score[N] - score[N-1]`.
- If `delta < improvement_threshold` for the last 2 iterations, flag stagnation.

When stagnation is detected, escalate to user via AskUserQuestion with three options:
1. Continue with current approach (Evaluator tries a different dimension focus)
2. Adjust criteria (user provides guidance or relaxes constraints)
3. Abort loop (accept current output as-is)

The escalation trigger at `escalation_after` iterations applies independently: if 3 iterations pass without a PASS score, escalate regardless of stagnation state.

---

### Evaluator Leniency Prevention Mechanisms

The following 5 mechanisms prevent score inflation and must be applied on every evaluation:

**Mechanism 1: Rubric Anchoring**

Score descriptions for each dimension:
- 0.25: Major defects, fails most criteria
- 0.50: Partial compliance, notable issues remain
- 0.75: Solid compliance, minor issues only
- 1.00: Full compliance, no issues found

Always state which rubric level applies and why before assigning a numeric score.

**Mechanism 2: Must-Pass Firewall**

The following conditions cause immediate FAIL regardless of other scores:
- Copy text differs from the original `copy.json` or BRIEF copy section
- AI slop detected: purple gradient (#8B5CF6-#6D28D9) as primary visual element with generic white cards
- Mobile viewport broken at 375px width (content overflow, unreadable text)
- Any interactive element returns 404 or broken state
- Lighthouse Accessibility < 80

**Mechanism 3: Anti-Pattern Penalty**

Known anti-patterns that cap dimension score at 0.50:
- Generic icon set without brand customization (Originality capped)
- Hard-coded spacing values outside the design token scale (Design Quality capped)
- Missing `alt` attributes on non-decorative images (Functionality capped)
- Section copy that does not match the contracted copy (Completeness capped)

**Mechanism 4: Evidence Requirement**

Each dimension score must cite specific evidence:
- Design Quality: Reference token file path and WCAG contrast ratio
- Originality: Describe what makes the design non-generic
- Completeness: List each BRIEF section and its implementation status
- Functionality: Reference test result or Playwright output

**Mechanism 5: Regression Baseline**

If a previous iteration passed a criterion, the current iteration must maintain that criterion. Regression from a previously passed criterion triggers an automatic score reduction in the relevant dimension.

---

### Sprint Contract Structure

Sprint Contract document format (`sprint-N.json`):

```json
{
  "sprint_id": "sprint-N",
  "iteration": N,
  "priority_dimension": "Design Quality | Originality | Completeness | Functionality",
  "acceptance_checklist": [
    {
      "id": "AC-01",
      "criterion": "Hero headline contrast ratio >= 4.5:1",
      "verification": "Check color pair with contrast calculator",
      "status": "pending | passed | failed"
    }
  ],
  "test_scenarios": [
    {
      "id": "TS-01",
      "description": "Mobile viewport renders without horizontal scroll",
      "tool": "Playwright | visual inspection",
      "command": "playwright test --viewport 375x667"
    }
  ],
  "pass_conditions": {
    "Design Quality": 0.75,
    "Originality": 0.70,
    "Completeness": 0.80,
    "Functionality": 0.75
  },
  "negotiation_history": [],
  "created_at": "ISO-8601"
}
```

---

## Advanced Patterns

### Strict Mode

When `strict_mode: true` in `design.yaml`:
- Each of the 4 dimension scores must individually meet `pass_threshold`.
- The weighted average alone is not sufficient.
- Minimum 2 iterations required even if the first iteration achieves a passing weighted average.
- Strict mode is recommended for client-facing deliverables.

### Independent Re-evaluation

Every 5th project triggers an independent re-evaluation:
- The same build is scored twice with independent prompts.
- If scores diverge by more than 0.10, a calibration warning is logged.
- Calibration results are stored in `sprint_contract.artifact_dir/calibration-log.json`.

### Playwright Integration

When claude-in-chrome MCP or Playwright is available, the Evaluator uses automated testing:
- Desktop screenshot (1280x720): full page
- Mobile screenshot (375x667): full page
- Interaction test: click all CTAs, verify no 404
- Accessibility scan: automated WCAG check

When testing tools are unavailable, fall back to static code analysis only, and note the limitation in the evaluation report.

---

## Works Well With

- `moai-domain-brand-design`: Provides design tokens that Evaluator validates in Design Quality dimension
- `moai-domain-copywriting`: Copy JSON is the reference for Completeness dimension
- `evaluator-active`: The GAN loop orchestrates evaluator-active for each scoring pass
- `moai-workflow-design-import`: Extracted tokens serve as the design reference baseline

---

Source: Absorbed from agency constitution (Section 11 GAN Loop Contract, Section 12 Evaluator Leniency Prevention) on 2026-04-20.
REQ coverage: REQ-SKILL-011, REQ-SKILL-012, REQ-SKILL-012a, REQ-SKILL-013, REQ-SKILL-014, REQ-CONST-004
Version: 1.0.0

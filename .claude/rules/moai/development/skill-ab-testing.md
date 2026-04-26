---
description: "A/B testing methodology for skill quality validation and iteration"
paths: ".claude/skills/**/SKILL.md,.claude/skills/**/*.md"
---

<!-- Source: revfactory/harness — Apache License 2.0 — see .claude/rules/moai/NOTICE.md -->

# Skill A/B Testing Methodology

Systematic approach to validating skill quality through comparative testing against baseline and iterative improvement.

## Core Principle

**With-Skill vs Baseline Comparison**: Every skill quality evaluation must compare performance with the skill loaded against performance without the skill under identical conditions.

---

## Test Framework

### Test Prompt Design

Create 2-3 test prompts that represent realistic use cases:

**Characteristics of Good Test Prompts**:
- Concrete and specific (avoid "analyze this")
- Natural language as users would write
- Cover core use case + edge case + optional complexity case
- Mix of explicit instructions and implicit context

**Bad Test Prompts**:
- "Generate code" (too vague)
- "Use this framework" (too prescriptive)

**Good Test Prompts**:
- "I have a React form with email + password fields. The password needs validation (min 8 chars, 1 uppercase, 1 number). Show me the hook code and tests."
- "This CSV has 3 columns: date, revenue, cost. Add a 4th column: profit margin %. Then sort by margin descending. Include handling for missing data."

### Test Execution: With-Skill vs Baseline

For each test prompt, execute **two parallel runs**:

| Dimension | With-Skill | Baseline |
|-----------|-----------|----------|
| **Prompt** | Same prompt + skill provided | Same prompt, no skill |
| **Output Path** | `_workspace/eval-{id}/with_skill/outputs/` | `_workspace/eval-{id}/without_skill/outputs/` |
| **Model** | Same model | Same model |
| **Constraints** | Same token budget, temperature | Same constraints |

### Metrics to Capture

**Token Usage**:
- `total_tokens`: Tokens consumed (lower is better for same quality)
- **Improvement**: (baseline_tokens - with_skill_tokens) / baseline_tokens

**Duration**:
- `duration_ms`: Wall-clock time (lower is better)
- **Speedup**: baseline_duration / with_skill_duration

**Quality Dimensions** (require manual evaluation or assertion-based scoring):

| Dimension | Measurement |
|-----------|-------------|
| **Correctness** | Does output solve the stated problem? (Boolean or %) |
| **Completeness** | Are all requirements met? (0-100%) |
| **Clarity** | Is code/response clear and well-organized? (0-100%) |
| **Efficiency** | Is approach optimal? (0-100%) |

---

## Assertion-Based Scoring

For skills with objective outputs (code generation, data extraction), define assertions:

**Structure**:
```
Expectation: "{human-readable expectation}"
Assertion: "{programmatic check}"
Evidence: "{what to look for in output}"
```

**Example Assertions**:
- `output.includes("useState") && output.includes("useEffect")` — React hooks present
- `fs.existsSync("output.csv") && csvHasRows("output.csv", 3)` — File created with data
- `jsAst.hasFunction("validateEmail") && hasReturnType(...)` — Function signature correct

**Scoring Rules**:
- Each assertion: 0 or 1 (or partial: 0.0-1.0)
- Score = assertions_passed / total_assertions
- **Improvement** = (with_skill_score - baseline_score) / baseline_score

---

## Statistical Significance

### Sample Size Guidance

| Difference Magnitude | Recommended Samples |
|---------------------|-------------------|
| ≥ 20% improvement | 2-3 samples (clear signal) |
| 10-20% improvement | 3-5 samples (moderate signal) |
| 5-10% improvement | 5-10 samples (weak signal, noise risk) |
| < 5% improvement | 10+ samples (ignore if < 2x total_tokens cost) |

### Variance Factors

Test prompt variance matters more than sample count. If your 2 test prompts exercise different skill aspects, they may be sufficient. If they're similar, results are less reliable.

**High-Variance Prompts** (good):
- Different domains (API design vs database design)
- Different skill uses (reference lookup vs advanced decision-making)
- Different complexity (simple vs complex scenario)

**Low-Variance Prompts** (less reliable):
- Similar domain (REST API v1 vs REST API v2)
- Similar complexity (both simple, both complex)

---

## Iteration Workflow

### Single Iteration Cycle

1. **Define Baselines**: Capture "with-skill" and "without-skill" outputs for 2-3 prompts
2. **Measure Metrics**: Token count, duration, assertion score
3. **Calculate Improvement**: (with_skill - baseline) / baseline for each metric
4. **Identify Gaps**: Which assertions failed? What prompt broke the skill?
5. **Refine Skill**: Improve description (better triggering), body (clearer instructions), schema
6. **Re-test**: Repeat with updated skill

### When to Iterate

**Continue iterating if**:
- Assertion pass rate < 80%
- With-skill score < baseline (regression)
- Improvement < 10% on primary metric (consider skill value)
- Skill fails on edge case prompt

**Stop iterating if**:
- 3+ iterations with < 5% improvement (diminishing returns)
- Improvement plateaued (no new gains for 2 cycles)
- Skill meets acceptance criteria (80%+ quality, 10%+ improvement)

### Iteration Cap

Set `max_iterations: 5` to prevent unbounded refinement. If quality target not met after 5 iterations, consider redesigning the skill or reconsidering whether it's valuable.

---

## Evaluation Checklist

**Before declaring skill complete**:

- [ ] 2-3 realistic test prompts defined
- [ ] With-skill and baseline runs executed for each prompt
- [ ] Token cost comparison calculated
- [ ] Quality assertions defined and scored (if applicable)
- [ ] With-skill improvement ≥ 10% on primary metric, ≥ 0% on all metrics (no regression)
- [ ] All edge case prompts pass
- [ ] Description accurately reflects trigger conditions
- [ ] Body structure follows progressive disclosure (Quick / Implementation / Advanced if large)
- [ ] Frontmatter complete (description, paths, metadata)

---

## Common Pitfalls

**Pitfall 1**: "Baseline that always passes"
- Assertion designed to pass with or without skill (non-discriminating)
- Fix: Baseline should have a lower score; skill should improve it measurably

**Pitfall 2**: "Token usage only metric"
- Token improvement doesn't guarantee quality improvement
- Fix: Always include quality metrics (assertions, manual evaluation)

**Pitfall 3**: "Single test prompt"
- One prompt may be lucky (skill happens to match well)
- Fix: Test 2-3 diverse prompts to validate generalization

**Pitfall 4**: "Ignoring model variance"
- Same prompt with different models/temperatures gives different results
- Fix: Keep model and constraints constant between with/baseline

**Pitfall 5**: "Unbounded iteration"
- Skill refinement never converges
- Fix: Set max_iterations and stop when improvement < threshold

---

## Reporting

For each skill iteration, document:

```markdown
## Iteration N Test Results

### Test Prompts
1. [Prompt A description]
2. [Prompt B description]

### Token Metrics
| Prompt | With-Skill | Baseline | Improvement |
|--------|-----------|----------|-------------|
| A | 5,234 | 6,102 | +14.2% |
| B | 3,891 | 4,012 | +2.9% |
| **Average** | — | — | **+8.5%** |

### Quality Metrics
| Assertion | With-Skill | Baseline | Status |
|-----------|-----------|----------|--------|
| Output is valid code | 100% | 75% | ✅ PASS |
| Includes all features | 100% | 50% | ✅ PASS |
| Well-formatted | 100% | 60% | ✅ PASS |
| **Average Score** | **100%** | **61.7%** | ✅ **+61.7%** |

### Conclusion
- Improvement target: ≥10% ✅ Met (8.5% token, +61.7% quality)
- All assertions passing ✅
- Ready for release
```

This methodology ensures skills provide measurable value before being marked complete.

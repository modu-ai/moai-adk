---
name: moai-foundation-thinking
description: >
  Structured thinking toolkit combining Critical Evaluation, Diverge-Converge
  Brainstorming, and Deep Questioning frameworks for creative problem-solving
  and rigorous analysis. Use when generating ideas, evaluating proposals,
  questioning assumptions, or exploring solution spaces systematically.
  Do NOT use for architecture decisions (use moai-foundation-philosopher instead)
  or code quality validation (use moai-foundation-quality instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob
user-invocable: false
metadata:
  version: "1.0.0"
  category: "foundation"
  status: "active"
  updated: "2026-02-10"
  modularized: "true"
  tags: "foundation, critical-thinking, brainstorming, ideation, evaluation, creative-thinking, diverge-converge"
  related-skills: "moai-foundation-philosopher"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords:
    - "brainstorm"
    - "ideation"
    - "creative"
    - "evaluate"
    - "critical thinking"
    - "diverge"
    - "converge"
    - "generate ideas"
    - "explore options"
    - "question"
    - "deep analysis"
    - "problem exploration"
    - "solution space"
    - "scoring"
    - "clustering"
    - "prioritize"
  agents:
    - "manager-strategy"
    - "manager-spec"
    - "team-analyst"
    - "team-architect"
    - "team-researcher"
  phases:
    - "plan"
---

# MoAI Foundation Thinking

Structured thinking toolkit for creative problem-solving and rigorous analysis. Integrates three complementary frameworks that cover the full spectrum from idea generation to critical evaluation.

Core Philosophy: Generate broadly, evaluate rigorously, question deeply. Creativity and criticism are complementary forces.

## Quick Reference

What is the Thinking Toolkit?

Three integrated frameworks for structured thinking:
- Critical Evaluation: Rigorous 7-step analysis to assess proposals and detect flaws
- Diverge-Converge: Systematic brainstorming from 20-50 raw ideas to 3-5 validated solutions
- Deep Questioning: 6-layer progressive inquiry to uncover hidden requirements and risks

When to Use Each Framework:

- Evaluating a proposal or recommendation: Critical Evaluation
- Generating solutions for an open-ended problem: Diverge-Converge
- Exploring an unfamiliar domain or unclear requirement: Deep Questioning
- Complex decisions: Combine all three (Question first, Generate second, Evaluate third)

Quick Access:
- Rigorous proposal assessment: [Critical Evaluation Module](modules/critical-evaluation.md)
- Creative solution generation: [Diverge-Converge Module](modules/diverge-converge.md)
- Progressive inquiry: [Deep Questioning Module](modules/deep-questioning.md)

---

## Implementation Guide

### Framework 1: Critical Evaluation

Purpose: Systematically assess proposals, claims, and recommendations to detect flaws before commitment.

Seven-Step Evaluation Process:

Step 1 - Restate: Reformulate the claim or proposal in your own words. Ensures genuine understanding before critique.

Step 2 - Assess Evidence: Examine supporting data. Is the evidence empirical, anecdotal, or assumed? What is the sample size and recency? Are there contradicting data points?

Step 3 - Detect Fallacies: Check for common reasoning errors. Appeal to authority without substance. False dichotomy (only two options presented). Hasty generalization from insufficient examples. Straw man misrepresentation of alternatives.

Step 4 - Expose Assumptions: Identify unstated premises. What must be true for this conclusion to hold? Which assumptions are testable? Which assumptions carry the highest risk if wrong?

Step 5 - Note Alternatives: For every claim, ask what else could explain the evidence. Generate at least two alternative interpretations. Consider the null hypothesis.

Step 6 - Check Contradictions: Look for internal inconsistencies. Do different parts of the proposal conflict? Are there contradictions with known facts or constraints?

Step 7 - Evaluate Burden of Proof: Determine if the evidence is proportional to the claim. Extraordinary claims require extraordinary evidence. Identify what additional evidence would strengthen or weaken the case.

Output Format:
- Evaluation Summary: Overall assessment (Strong, Moderate, Weak, Flawed)
- Key Strengths: What holds up under scrutiny
- Critical Gaps: What needs more evidence or revision
- Recommended Actions: Next steps to strengthen the proposal

WHY: Uncritical acceptance of proposals leads to preventable failures.
IMPACT: Structured evaluation catches 60-80% of flawed recommendations.

### Framework 2: Diverge-Converge Brainstorming

Purpose: Generate a broad solution space then systematically narrow to the best options.

Five-Phase Process:

Phase 1 - Gather Requirements: Define the problem space clearly. Identify stakeholders and success criteria. Set explicit constraints (budget, timeline, technology). Document "must-have" vs "nice-to-have" criteria.

Phase 2 - Diverge (Generate 20-50 Ideas): Quantity over quality during divergence. No criticism or filtering during generation. Include wild and unconventional ideas. Combine and build upon previous ideas. Use prompts: "What if we...", "How might we...", "What would happen if..."

Phase 3 - Cluster (Group into 4-8 Themes): Identify natural groupings among ideas. Name each cluster with a descriptive theme. Note which clusters have the most ideas (signals interest). Identify gaps where no ideas exist (potential blind spots).

Phase 4 - Converge (Score and Select): Rate each cluster against success criteria (1-10). Apply weighted scoring based on priority of criteria. Select top 3-5 candidates for deeper analysis. Document why rejected options were eliminated.

Phase 5 - Document and Validate: Write up selected solutions with rationale. Define validation experiments for top candidates. Identify risks and mitigation strategies. Plan implementation sequence.

Output Format:
- Problem Statement: Clear definition of what we are solving
- Idea Count: Total ideas generated and cluster distribution
- Top Candidates: 3-5 selected solutions with scores
- Validation Plan: How to test each candidate

WHY: Premature convergence on the first idea leaves better solutions undiscovered.
IMPACT: Teams using diverge-converge find 3x more viable solutions.

### Framework 3: Deep Questioning

Purpose: Progressively uncover hidden requirements, constraints, and risks through layered inquiry.

Six-Layer Progressive Inquiry:

Layer 1 - Surface Understanding: What is the stated goal or request? What does success look like? What are the obvious inputs and outputs? Verify: Can I explain this to someone else clearly?

Layer 2 - Problem Depth: Why does this problem exist? What is the root cause vs symptom? What has been tried before and why did it fail? What would happen if we did nothing?

Layer 3 - Context and Constraints: What are the technical constraints? What are the organizational or process constraints? What are the time and resource limitations? What external dependencies exist?

Layer 4 - User Perspective: Who are the actual end users? What is their current workflow? What pain points drive this request? What would they consider a disappointing solution?

Layer 5 - Solution Exploration: What are the boundary conditions? What edge cases could break the solution? What are the performance requirements? How will this integrate with existing systems?

Layer 6 - Validation and Risk: How will we know if the solution works? What could go wrong? What is the rollback strategy? What monitoring or alerting is needed?

Progressive Depth Indicators:
- Shallow: Only Layers 1-2 explored (common in quick tasks)
- Moderate: Layers 1-4 explored (sufficient for most features)
- Deep: All 6 layers explored (required for architecture decisions)
- Exhaustive: All layers with multiple iterations (critical systems)

Output Format:
- Understanding Level: Shallow, Moderate, Deep, or Exhaustive
- Key Discoveries: Insights from each explored layer
- Open Questions: Remaining unknowns requiring further investigation
- Risk Assessment: Identified risks by severity

WHY: Surface-level understanding leads to solutions that miss the real problem.
IMPACT: Deep questioning reduces requirement changes by 40-60%.

---

## Combined Workflow

For complex problems, use all three frameworks in sequence:

Step 1 - Deep Questioning: Explore the problem space (Layers 1-4 minimum)
Step 2 - Diverge-Converge: Generate and select solutions based on discoveries
Step 3 - Critical Evaluation: Rigorously assess the top candidates

Decision Complexity Guide:

Simple task (1-2 files): Skip thinking frameworks (direct implementation)
Feature addition: Deep Questioning (Layers 1-3) + brief evaluation
Design decision: Deep Questioning (full) + Diverge-Converge
Architecture change: All three frameworks in full

---

## Integration with MoAI Workflow

SPEC Phase (/moai plan):
- Apply Deep Questioning during requirements gathering
- Use Diverge-Converge for solution approach selection
- Apply Critical Evaluation to finalize SPEC document

Run Phase (/moai run):
- Use Critical Evaluation when reviewing implementation options
- Apply Deep Questioning when encountering unexpected complexity

Agent Teams:
- team-analyst: Primary user of Deep Questioning framework
- team-architect: Primary user of Critical Evaluation framework
- team-researcher: Uses all three for comprehensive analysis

---

## Works Well With

Agents:
- manager-strategy: Combined with Philosopher for full decision framework
- manager-spec: Deep Questioning during requirement analysis
- team-analyst: Primary consumer for plan phase analysis
- team-researcher: Comprehensive research methodology

Skills:
- moai-foundation-philosopher: Complementary (Philosopher = strategic decisions, Thinking = creative analysis)
- moai-foundation-core: Integration with SPEC workflow
- moai-workflow-spec: Requirement documentation support

Commands:
- /moai plan: Apply thinking frameworks during specification
- /moai run: Reference during implementation decisions

---

Module Deep Dives:
- [Critical Evaluation](modules/critical-evaluation.md)
- [Diverge-Converge](modules/diverge-converge.md)
- [Deep Questioning](modules/deep-questioning.md)

External Resources: [reference.md](references/reference.md)

Origin: Integrated from critical-thinking, brainstorm-diverge-converge, and ideation frameworks

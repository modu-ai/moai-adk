---
name: moai-foundation-philosopher
description: >
  Strategic thinking framework integrating First Principles Analysis, Stanford Design
  Thinking, and MIT Systems Engineering for deeper problem-solving.
  Use when performing architecture decisions, technology selection trade-offs,
  root cause analysis, cognitive bias detection, or first principles decomposition.
  Do NOT use for code quality validation (use moai-foundation-quality instead)
  or implementation workflows (use moai-workflow-ddd instead).
license: Apache-2.0
compatibility: Designed for Claude Code
allowed-tools: Read Grep Glob mcp__context7__resolve-library-id mcp__context7__get-library-docs
user-invocable: false
metadata:
  version: "1.1.0"
  category: "foundation"
  status: "active"
  updated: "2026-01-08"
  modularized: "true"
  tags: "foundation, strategic-thinking, first-principles, trade-off-analysis, cognitive-bias, decision-making"

# MoAI Extension: Progressive Disclosure
progressive_disclosure:
  enabled: true
  level1_tokens: 100
  level2_tokens: 5000

# MoAI Extension: Triggers
triggers:
  keywords:
    - "architecture"
    - "architecture decision"
    - "technology selection"
    - "trade-off"
    - "strategic"
    - "decision"
    - "analysis"
    - "design thinking"
    - "first principles"
    - "five whys"
    - "assumption"
    - "alternative"
    - "cognitive bias"
    - "root cause"
    - "framework selection"
    - "library selection"
    - "database selection"
    - "performance vs maintainability"
    - "breaking change"
  agents:
    - "manager-strategy"
    - "manager-spec"
    - "expert-backend"
    - "expert-frontend"
    - "expert-devops"
  phases:
    - "plan"
---

# MoAI Foundation Philosopher

Strategic thinking framework that promotes deeper analysis over quick calculations. Integrates three proven methodologies for systematic problem-solving.

Core Philosophy: Think deeply before acting. Question assumptions. Consider alternatives. Make trade-offs explicit. Check for cognitive biases.

## Quick Reference (30 seconds)

What is the Philosopher Framework?

A structured approach to complex decisions combining:
- First Principles Analysis: Break problems to fundamental truths
- Stanford Design Thinking: Divergent-convergent solution generation
- MIT Systems Engineering: Systematic risk assessment and validation

Five-Phase Thinking Process:
1. Assumption Audit: Surface and question what we take for granted
2. First Principles Decomposition: Break down to root causes
3. Alternative Generation: Create multiple solution options
4. Trade-off Analysis: Compare options systematically
5. Cognitive Bias Check: Verify thinking quality

When to Activate:
- Architecture decisions affecting 5+ files
- Technology selection (library, framework, database)
- Performance vs maintainability trade-offs
- Refactoring scope decisions
- Breaking changes consideration
- Any decision with significant long-term impact

Quick Access:
- Assumption questioning techniques: [Assumption Matrix Module](modules/assumption-matrix.md)
- Root cause analysis: [First Principles Module](modules/first-principles.md)
- Option comparison: [Trade-off Analysis Module](modules/trade-off-analysis.md)
- Bias prevention: [Cognitive Bias Module](modules/cognitive-bias.md)

---

## Implementation Guide (5 minutes)

### Phase 1: Assumption Audit

Purpose: Surface hidden assumptions before they become blind spots.

Five Critical Questions:
- What are we assuming to be true without evidence?
- What if this assumption turns out to be wrong?
- Is this a hard constraint or merely a preference?
- What evidence supports this assumption?
- Who else should validate this assumption?

Assumption Categories:
- Technical Assumptions: Technology capabilities, performance characteristics, compatibility
- Business Assumptions: User behavior, market conditions, budget availability
- Team Assumptions: Skill levels, availability, domain knowledge
- Timeline Assumptions: Delivery expectations, dependency schedules

Assumption Documentation Format:
- Assumption statement: Clear description of what is assumed
- Confidence level: High, Medium, or Low based on evidence
- Evidence basis: What supports this assumption
- Risk if wrong: Consequence if assumption proves false
- Validation method: How to verify before committing

WHY: Unexamined assumptions are the leading cause of project failures and rework.
IMPACT: Surfacing assumptions early prevents 40-60% of mid-project pivots.

### Phase 2: First Principles Decomposition

Purpose: Cut through complexity to find root causes and fundamental requirements.

The Five Whys Technique:
- Surface Problem: What the user or system observes
- First Why: Immediate cause analysis
- Second Why: Underlying cause investigation
- Third Why: Systemic driver identification
- Fourth Why: Organizational or process factor
- Fifth Why (Root Cause): Fundamental issue to adddess

Constraint Analysis:
- Hard Constraints: Non-negotiable (security, compliance, physics, budget)
- Soft Constraints: Negotiable preferences (timeline, feature scope, tooling)
- Self-Imposed Constraints: Assumptions disguised as requirements
- Degrees of Freedom: Areas where creative solutions are possible

Decomposition Questions:
- What is the actual goal behind this request?
- What problem are we really trying to solve?
- What would a solution look like if we had no constraints?
- What is the minimum viable solution?
- What can we eliminate while still achieving the goal?

WHY: Most problems are solved at the wrong level of abstraction.
IMPACT: First principles thinking reduces solution complexity by 30-50%.

### Phase 3: Alternative Generation

Purpose: Avoid premature convergence on suboptimal solutions.

Generation Rules:
- Minimum three distinct alternatives required
- Include at least one unconventional option
- Always include "do nothing" as baseline
- Consider short-term vs long-term implications
- Explore both incremental and transformative approaches

Alternative Categories:
- Conservative: Low risk, incremental improvement, familiar technology
- Balanced: Moderate risk, significant improvement, some innovation
- Aggressive: Higher risk, transformative change, cutting-edge approach
- Radical: Challenge fundamental assumptions, completely different approach

Creativity Techniques:
- Inversion: What would make this problem worse? Now do the opposite.
- Analogy: How do other domains solve similar problems?
- Constraint Removal: What if budget, time, or technology were unlimited?
- Simplification: What is the simplest possible solution?

WHY: The first solution is rarely the best solution.
IMPACT: Considering 3+ alternatives improves decision quality by 25%.

### Phase 4: Trade-off Analysis

Purpose: Make implicit trade-offs explicit and comparable.

Standard Evaluation Criteria:
- Performance: Speed, throughput, latency, resource usage
- Maintainability: Code clarity, documentation, team familiarity
- Implementation Cost: Development time, complexity, learning curve
- Risk Level: Technical risk, failure probability, rollback difficulty
- Scalability: Growth capacity, flexibility, future-proofing
- Security: Vulnerability surface, compliance, data protection

Weighted Scoring Method:
- Assign weights to criteria based on project priorities (total 100%)
- Rate each option 1-10 on each criterion
- Calculate weighted composite score
- Document reasoning for each score
- Identify score sensitivity to weight changes

Trade-off Documentation:
- What we gain: Primary benefits of chosen approach
- What we sacrifice: Explicit costs and limitations accepted
- Why acceptable: Rationale for accepting these trade-offs
- Mitigation plan: How to adddess downsides

WHY: Implicit trade-offs lead to regret and second-guessing.
IMPACT: Explicit trade-offs improve stakeholder alignment by 50%.

### Phase 5: Cognitive Bias Check

Purpose: Ensure recommendation quality by checking for common thinking errors.

Primary Biases to Monitor:
- Anchoring Bias: Over-reliance on first information encountered
- Confirmation Bias: Seeking evidence that supports existing beliefs
- Sunk Cost Fallacy: Continuing due to past investment
- Availability Heuristic: Overweighting recent or memorable events
- Overconfidence Bias: Excessive certainty in own judgment

Bias Detection Questions:
- Am I attached to this solution because I thought of it first?
- Have I actively sought evidence against my preference?
- Would I recommend this if starting fresh with no prior investment?
- Am I being influenced by recent experiences that may not apply?
- What would change my mind about this recommendation?

Mitigation Strategies:
- Pre-mortem: Imagine the decision failed; what went wrong?
- Devil's advocate: Argue against your own recommendation
- Outside view: What do base rates suggest about success?
- Disagreement seeking: Consult someone likely to challenge you
- Reversal test: If the opposite were proposed, what would you say?

WHY: Even experts fall prey to cognitive biases under time pressure.
IMPACT: Bias checking prevents 20-30% of flawed technical decisions.

---

## Advanced Implementation (10+ minutes)

### Integration with MoAI Workflow

SPEC Phase Integration:
- Apply Assumption Audit during /moai:1-plan
- Document assumptions in spec.md Problem Analysis section
- Include alternative approaches considered in plan.md
- Define validation criteria in acceptance.md

DDD Phase Integration:
- Use First Principles to identify core test scenarios
- Generate characterization test alternatives for legacy code
- Generate specification test alternatives for new features
- Apply Trade-off Analysis for test coverage decisions

Quality Phase Integration:
- Include Cognitive Bias Check in code review process
- Verify assumptions remain valid after implementation
- Document trade-offs accepted in final documentation

### Time Allocation Guidelines

Recommended effort distribution for complex decisions:
- Assumption Audit: 15% of analysis time
- First Principles Decomposition: 25% of analysis time
- Alternative Generation: 20% of analysis time
- Trade-off Analysis: 25% of analysis time
- Cognitive Bias Check: 15% of analysis time

Total Analysis vs Implementation:
- Simple decisions (1-2 files): 10% analysis, 90% implementation
- Medium decisions (3-10 files): 25% analysis, 75% implementation
- Complex decisions (10+ files): 40% analysis, 60% implementation
- Architecture decisions: 50% analysis, 50% implementation

### Decision Documentation Template

Strategic Decision Record:

Decision Title: Clear statement of what was decided

Context: Why this decision was needed

Assumptions Examined:
- Assumption 1 with confidence and validation status
- Assumption 2 with confidence and validation status

Root Cause Analysis:
- Surface problem identified
- Root cause determined through Five Whys

Alternatives Considered:
- Option A with pros, cons, and score
- Option B with pros, cons, and score
- Option C with pros, cons, and score

Trade-offs Accepted:
- What we gain with chosen approach
- What we sacrifice and why acceptable

Bias Check Completed:
- Confirmation of bias mitigation steps taken

Final Decision: Selected option with primary rationale

Success Criteria: How we will measure if decision was correct

Review Trigger: Conditions that would cause reconsideration

---

## Works Well With

Agents:
- manager-strategy: Primary consumer for SPEC analysis and planning
- expert-backend: Technology selection decisions
- expert-frontend: Architecture and framework choices
- expert-database: Schema design trade-offs
- manager-quality: Code review bias checking

Skills:
- moai-foundation-core: Integration with TRUST 5 and SPEC workflow
- moai-workflow-spec: Assumption documentation in SPEC format
- moai-domain-backend: Technology-specific trade-off criteria
- moai-domain-frontend: UI/UX decision frameworks

Commands:
- /moai:1-plan: Apply Philosopher Framework during specification
- /moai:2-run: Reference documented trade-offs during implementation

---

## Quick Decision Matrix

When to use which phase:

Simple Bug Fix: Skip Philosopher (direct implementation)
Feature Addition: Phases 1, 3, 4 (assumptions, alternatives, trade-offs)
Refactoring: Phases 1, 2, 4 (assumptions, root cause, trade-offs)
Technology Selection: All 5 phases (full analysis required)
Architecture Change: All 5 phases with extended documentation

---

Module Deep Dives:
- [Assumption Matrix](modules/assumption-matrix.md)
- [First Principles](modules/first-principles.md)
- [Trade-off Analysis](modules/trade-off-analysis.md)
- [Cognitive Bias](modules/cognitive-bias.md)

Examples: [examples.md](examples.md)
External Resources: [reference.md](reference.md)

Origin: Inspired by Claude Code Philosopher Ignition framework

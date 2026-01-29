# Trade-off Analysis Module

Deep dive into systematic option comparison and decision-making frameworks.

## Weighted Scoring Method

### Step 1: Define Evaluation Criteria

Standard criteria for technical decisions:

Performance Criteria:
- Response time and latency
- Throughput and capacity
- Resource efficiency (CPU, memory, network)
- Scalability under load

Quality Criteria:
- Code maintainability and readability
- Test coverage and testability
- Documentation completeness
- Error handling robustness

Cost Criteria:
- Implementation effort (person-days)
- Learning curve and training needs
- Operational cost (infrastructure, licensing)
- Technical debt introduced

Risk Criteria:
- Implementation complexity
- Dependency on external factors
- Failure modes and recovery
- Rollback difficulty

Strategic Criteria:
- Alignment with architecture vision
- Future flexibility and extensibility
- Team skill development
- Industry standard compliance

### Step 2: Assign Weights

Weight assignment process:

1. List all criteria relevant to the decision
2. Use AskUserQuestion to understand user priorities
3. Distribute 100% across criteria based on priority
4. Document rationale for weight assignments

Example Weight Distribution:

Performance-Critical Project:
- Performance: 35%
- Quality: 20%
- Cost: 15%
- Risk: 20%
- Strategic: 10%

Maintainability-Focused Project:
- Performance: 15%
- Quality: 35%
- Cost: 20%
- Risk: 15%
- Strategic: 15%

Rapid Delivery Project:
- Performance: 15%
- Quality: 15%
- Cost: 40%
- Risk: 20%
- Strategic: 10%

### Step 3: Score Options

Scoring guidelines (1-10 scale):

Score 9-10: Excellent, clearly superior
Score 7-8: Good, above average
Score 5-6: Adequate, meets requirements
Score 3-4: Below average, has concerns
Score 1-2: Poor, significant problems

Scoring requirements:
- Provide specific rationale for each score
- Reference evidence or experience
- Consider uncertainty in scores
- Be consistent across options

### Step 4: Calculate Composite Scores

Calculation method:
- Multiply each score by criterion weight
- Sum weighted scores for total
- Compare totals across options
- Analyze sensitivity to weight changes

## Trade-off Documentation

### Trade-off Record Format

For each significant trade-off:

Trade-off ID: T-001
Decision Context: What decision required this trade-off
What We Gain: Benefits of chosen approach
What We Sacrifice: Costs or limitations accepted
Why Acceptable: Rationale for accepting this trade-off
Mitigation Plan: Actions to reduce downside impact
Review Trigger: Conditions that would cause reconsideration

### Common Trade-off Patterns

Speed vs Quality:
- Faster delivery vs more thorough testing
- Quick fix vs proper solution
- MVP vs full feature set

Performance vs Maintainability:
- Optimized code vs readable code
- Custom solution vs standard library
- Inline logic vs abstraction layers

Flexibility vs Simplicity:
- Configurable vs hardcoded
- Generic vs specific
- Plugin architecture vs monolithic

Cost vs Capability:
- Build vs buy
- Open source vs commercial
- Cloud vs on-premise

## Integration with AskUserQuestion

When analyzing trade-offs:
- Use AskUserQuestion to confirm criterion weights match priorities
- Use AskUserQuestion to present options with scores
- Use AskUserQuestion to validate trade-off acceptability
- Use AskUserQuestion to explore sensitivity to different weights

Example AskUserQuestion for Trade-off Confirmation:

Question: Based on the trade-off analysis, Option B scores highest. It sacrifices some performance (score 6) for better maintainability (score 9). Is this trade-off acceptable?

Options:
- Accept Option B with acknowledged trade-off
- Prioritize performance over maintainability
- Request deeper analysis of performance impact
- Explore hybrid approach combining elements

---

Version: 1.0.0
Parent Skill: moai-foundation-philosopher

# Critical Evaluation Module

Detailed guide for the 7-step critical evaluation framework.

## Step-by-Step Process

### Step 1: Restate

Before evaluating, demonstrate understanding by restating the claim or proposal.

Restatement Techniques:
- Paraphrase in simpler language
- Identify the core assertion
- Separate facts from opinions
- Note the scope of the claim

Quality Check: If the original author would not recognize your restatement, you have misunderstood.

### Step 2: Assess Evidence

Evaluate the quality and relevance of supporting evidence.

Evidence Categories:
- Empirical Data: Measured, repeatable, verifiable (strongest)
- Case Studies: Specific examples with documented outcomes
- Expert Opinion: Authority-based but potentially biased
- Anecdotal: Personal experience (weakest standalone)
- Theoretical: Logically derived but untested

Assessment Questions:
- Is the evidence current or outdated?
- Is the sample size sufficient?
- Are there selection biases in the data?
- Has the evidence been peer-reviewed or independently verified?
- Are there contradicting data points being ignored?

### Step 3: Detect Fallacies

Common logical fallacies in technical proposals:

Appeal to Authority: "Google uses it, so it must be good."
Counter: Evaluate on merit, not on who uses it.

False Dichotomy: "We either use microservices or the system will not scale."
Counter: Identify intermediate options.

Hasty Generalization: "The last React project was fast, so React is always fast."
Counter: Examine broader evidence base.

Straw Man: "The alternative is a monolith with no modularity."
Counter: Represent alternatives fairly.

Sunk Cost: "We have invested too much to switch now."
Counter: Evaluate based on future value only.

Bandwagon: "Everyone is migrating to Kubernetes."
Counter: Evaluate fitness for specific context.

### Step 4: Expose Assumptions

Hidden Assumption Discovery Technique:

Reverse the conclusion: "If we did NOT adopt this approach, what would we lose?"

Challenge each "obvious" premise:
- "Users will prefer this interface" (Have we tested this?)
- "Performance will improve with caching" (What access patterns support this?)
- "The team can learn this technology quickly" (What is the actual learning curve?)

Assumption Risk Matrix:
- High Confidence + Low Impact: Accept and move on
- High Confidence + High Impact: Verify with data
- Low Confidence + Low Impact: Monitor casually
- Low Confidence + High Impact: Investigate immediately (danger zone)

### Step 5: Note Alternatives

For every conclusion, generate at least two alternatives:

Alternative Generation Prompts:
- What would a skeptic suggest instead?
- What if the opposite approach is better?
- What would a different industry do?
- What is the most conservative option?
- What is the most radical option?

Null Hypothesis: What happens if we do nothing? This establishes the baseline for comparison.

### Step 6: Check Contradictions

Internal Consistency Analysis:

Common contradiction patterns:
- Claiming simplicity while proposing complex architecture
- Promising speed while adding abstraction layers
- Emphasizing maintainability while using unfamiliar technology
- Stating scalability while using stateful design

Cross-Reference Checks:
- Do performance claims match architectural choices?
- Do timeline estimates match scope of work?
- Do team capabilities match technology requirements?

### Step 7: Evaluate Burden of Proof

Proportionality Principle: The strength of evidence should match the magnitude of the claim.

Evidence Levels for Technical Claims:
- Trivial claim (renaming a variable): No evidence needed
- Minor claim (using a utility library): Brief rationale sufficient
- Moderate claim (adopting a new pattern): Documented comparison required
- Major claim (architecture change): Proof of concept with benchmarks
- Critical claim (replacing core technology): Full evaluation with multiple stakeholders

---

## Evaluation Output Template

Claim Under Evaluation: [Clear statement]

Restatement: [Your understanding of the claim]

Evidence Assessment:
- Quality: [Strong / Moderate / Weak / Insufficient]
- Key Evidence: [List supporting data points]
- Missing Evidence: [What would strengthen the case]

Fallacies Detected:
- [Fallacy type]: [Specific instance]

Hidden Assumptions:
- [Assumption]: [Risk level] - [Validation status]

Alternatives Considered:
- [Alternative 1]: [Brief assessment]
- [Alternative 2]: [Brief assessment]

Contradictions Found:
- [Contradiction]: [Impact assessment]

Overall Verdict: [Strong / Moderate / Weak / Flawed]

Recommended Actions:
- [Action 1]
- [Action 2]

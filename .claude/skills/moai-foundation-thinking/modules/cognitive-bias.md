# Cognitive Bias Module

Deep dive into identifying and mitigating cognitive biases in technical decision-making.

## Primary Biases in Technical Decisions

### Anchoring Bias

Definition: Over-reliance on the first piece of information encountered.

How it manifests in development:
- First solution considered becomes the default
- Initial estimate becomes the baseline
- First technology researched gets preference
- Original requirements dominate even when changed

Detection questions:
- Is this solution preferred because it was first or because it is best?
- Would I reach the same conclusion if I encountered alternatives first?
- Am I adjusting from an initial value or evaluating independently?

Mitigation strategies:
- Deliberately consider alternatives before committing
- Generate options independently before comparison
- Seek input from people who started with different information
- Use structured evaluation frameworks

### Confirmation Bias

Definition: Seeking and interpreting information to confirm existing beliefs.

How it manifests in development:
- Searching for success stories of preferred technology
- Dismissing concerns about chosen approach
- Interpreting ambiguous data as supporting decision
- Selective attention to positive feedback

Detection questions:
- Am I seeking evidence that challenges my preference?
- Would I accept this evidence if it opposed my view?
- Have I genuinely considered why someone might disagree?

Mitigation strategies:
- Actively search for counterexamples
- Assign someone to argue the opposite position
- List reasons why the preferred option might fail
- Seek feedback from known skeptics

### Sunk Cost Fallacy

Definition: Continuing investment due to previously invested resources.

How it manifests in development:
- Continuing with failing approach due to time invested
- Maintaining legacy code due to past effort
- Persisting with tool due to training investment
- Defending decision due to public commitment

Detection questions:
- Would I start down this path if beginning fresh today?
- Am I continuing because of future value or past investment?
- What would I advise someone else in this situation?

Mitigation strategies:
- Evaluate decisions based on future costs and benefits only
- Set clear criteria for stopping before starting
- Regular checkpoint reviews with fresh perspective
- Separate decision from past investment explicitly

### Availability Heuristic

Definition: Overweighting information that comes easily to mind.

How it manifests in development:
- Recent project experience dominates thinking
- Memorable failures cause excessive caution
- Familiar technologies preferred over better alternatives
- Recent news influences technology assessment

Detection questions:
- Is this concern based on frequency or memorability?
- What does systematic data say vs anecdotal experience?
- Am I overweighting recent events?

Mitigation strategies:
- Seek base rate data for objective comparison
- Consider experiences beyond recent memory
- Consult documentation and research over recollection
- Balance personal experience with broader evidence

### Overconfidence Bias

Definition: Excessive confidence in own judgment and predictions.

How it manifests in development:
- Underestimating implementation complexity
- Overestimating ability to handle edge cases
- Certainty about user behavior predictions
- Confidence in estimates without uncertainty ranges

Detection questions:
- What is my track record on similar predictions?
- How would I feel if this turned out differently?
- What would need to be true for me to be wrong?

Mitigation strategies:
- Express estimates as ranges with confidence levels
- Review historical accuracy of predictions
- Seek external validation for important judgments
- Include explicit uncertainty in recommendations

## Bias Mitigation Techniques

### Pre-mortem Analysis

Process:
1. Imagine the decision has been implemented
2. Assume it has failed spectacularly
3. Write down all the reasons for failure
4. Evaluate which failure modes are most likely
5. Adddess highest-risk failure modes proactively

Benefits:
- Legitimizes dissent and concern-raising
- Surfaces risks that optimism might hide
- Creates concrete mitigation opportunities

### Devil's Advocate

Process:
1. Assign someone to argue against the preferred option
2. Require them to make the strongest possible case
3. Take the counterarguments seriously
4. Adjust decision if counterarguments reveal blind spots

Guidelines:
- Devil's advocate should prepare thoroughly
- Arguments should be substantive, not token
- Response should adddess arguments, not dismiss

### Outside View

Process:
1. Identify reference class of similar decisions
2. Research base rates for success and failure
3. Compare current situation to reference class
4. Adjust confidence based on base rates

Questions:
- How have similar decisions turned out historically?
- What is the typical success rate for this type of project?
- What makes this situation different from the reference class?

## Bias Check Checklist

Before finalizing any significant recommendation:

Anchoring Check:
- [ ] Considered alternatives before settling on first idea
- [ ] Evaluated options independently, not relative to anchor
- [ ] Sought perspectives from those with different starting points

Confirmation Check:
- [ ] Actively sought evidence against preferred option
- [ ] Genuinely considered counterarguments
- [ ] Consulted skeptics or critics

Sunk Cost Check:
- [ ] Evaluated based on future value only
- [ ] Considered what I would advise a newcomer
- [ ] Separated decision from past investments

Availability Check:
- [ ] Checked base rates beyond personal experience
- [ ] Considered data beyond recent events
- [ ] Balanced anecdotes with systematic evidence

Overconfidence Check:
- [ ] Expressed uncertainty in predictions
- [ ] Considered scenarios where I am wrong
- [ ] Reviewed historical accuracy of similar predictions

## Integration with Decision Process

Timing of bias checks:
- After generating initial recommendation
- Before presenting options to stakeholders
- When defending a position strongly
- When dismissing alternatives quickly

Documentation:
- Record which biases were checked
- Note any biases identified and how adddessed
- Document remaining uncertainty

---

Version: 1.0.0
Parent Skill: moai-foundation-philosopher

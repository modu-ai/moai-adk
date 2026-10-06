# Diverge-Converge Module

Detailed guide for the 5-phase brainstorming framework.

## Phase-by-Phase Guide

### Phase 1: Gather Requirements

Before generating ideas, establish a clear problem definition.

Requirement Gathering Checklist:
- Problem statement in one sentence
- Who is affected by this problem?
- What does success look like? (measurable criteria)
- What are the hard constraints? (non-negotiable)
- What are the soft constraints? (preferences)
- What resources are available? (budget, time, team)
- What has been tried before?

Problem Statement Template:
"We need to [achieve goal] for [stakeholders] within [constraints] because [motivation]."

Success Criteria Checklist:
- Must-Have: Features or qualities the solution requires
- Nice-to-Have: Features that add value but are not essential
- Must-Not-Have: Explicit exclusions or anti-patterns

### Phase 2: Diverge (Quantity over Quality)

Target: Generate 20-50 raw ideas without filtering.

Divergence Rules:
- No criticism during generation (defer judgment)
- Wild ideas are welcome (they spark practical ones)
- Build on others' ideas (combine and extend)
- Aim for quantity (quality comes later)
- Stay visual or brief (one sentence per idea)

Idea Generation Techniques:

Technique 1 - What If: Start every idea with "What if we..."
- What if we removed this constraint entirely?
- What if we solved the opposite problem?
- What if budget were unlimited?
- What if we had to ship in one day?

Technique 2 - Analogy Transfer: How do other domains solve this?
- How does nature solve this? (biomimicry)
- How does a different industry handle this?
- How would a competitor approach this?
- How did we solve something similar before?

Technique 3 - Constraint Inversion: Temporarily remove or reverse constraints.
- What if the data were read-only?
- What if we had no backward compatibility requirement?
- What if the user interface were voice-only?

Technique 4 - Combination: Merge two unrelated ideas.
- Take idea A's strength and idea B's simplicity
- Combine the cheapest approach with the most reliable

Idea Documentation Format:
- ID: Sequential number (1, 2, 3...)
- Title: 3-5 word description
- One Sentence: What it does and why
- Category: Technical, Process, Design, Hybrid

### Phase 3: Cluster (Identify Themes)

Target: Group ideas into 4-8 meaningful themes.

Clustering Process:
1. Read through all ideas once without categorizing
2. Identify the 2-3 most obvious groupings
3. Sort remaining ideas into existing or new groups
4. Merge groups that are too similar
5. Split groups that are too broad
6. Name each cluster with a descriptive theme

Cluster Quality Indicators:
- Good: 3-8 ideas per cluster, distinct theme
- Too broad: 10+ ideas (split into sub-themes)
- Too narrow: 1-2 ideas (merge with related cluster)
- Gap indicator: Important area with no ideas (explore further)

Cluster Documentation:
- Theme Name: Descriptive label
- Idea Count: Number of ideas in this cluster
- Key Insight: What unifies these ideas
- Representative Idea: Best example from the cluster

### Phase 4: Converge (Score and Select)

Target: Select top 3-5 candidates using weighted scoring.

Scoring Matrix Setup:

Default Criteria (adjust weights per project):
- Feasibility (25%): Can we actually build this?
- Impact (25%): How much value does it deliver?
- Effort (20%): Development cost and complexity
- Risk (15%): Probability and severity of failure
- Alignment (15%): Fit with existing architecture and goals

Scoring Scale (1-10):
- 1-2: Very poor / High risk / Major effort
- 3-4: Below average / Moderate-high concerns
- 5-6: Average / Acceptable trade-offs
- 7-8: Good / Minor concerns only
- 9-10: Excellent / Minimal downsides

Selection Process:
1. Score each cluster (not individual ideas) on all criteria
2. Calculate weighted scores
3. Rank clusters by total weighted score
4. Select top 3-5 for deeper analysis
5. Document why lower-ranked options were eliminated

Tie-Breaking Rules:
- Higher feasibility wins (prefer buildable solutions)
- Lower risk wins (prefer safer options)
- Higher alignment wins (prefer consistent approaches)

### Phase 5: Document and Validate

Target: Create actionable documentation for selected solutions.

Solution Documentation Template:

For each selected candidate:
- Solution Name: Clear, descriptive title
- Problem Addressed: Which requirements does this solve?
- Approach: High-level description of the solution
- Pros: Key advantages and strengths
- Cons: Known limitations and weaknesses
- Score: Weighted score from convergence phase
- Validation Method: How to test this solution
- Estimated Effort: Relative sizing (S/M/L/XL)
- Risks: Identified risks and mitigation strategies
- Dependencies: What this solution requires

Validation Experiments:
- Define minimum viable test for each candidate
- Identify what would prove the solution works
- Identify what would prove the solution fails
- Set time box for each validation experiment

---

## Diverge-Converge Output Template

Problem Statement: [One sentence]

Divergence Results:
- Total Ideas Generated: [count]
- Generation Techniques Used: [list]

Cluster Summary:
| Theme | Idea Count | Key Insight |
|-------|-----------|-------------|
| [Theme 1] | [count] | [insight] |
| [Theme 2] | [count] | [insight] |

Convergence Scores:
| Theme | Feasibility | Impact | Effort | Risk | Alignment | Weighted Total |
|-------|------------|--------|--------|------|-----------|---------------|
| [Theme 1] | [score] | [score] | [score] | [score] | [score] | [total] |

Selected Candidates:
1. [Top solution with brief rationale]
2. [Second solution with brief rationale]
3. [Third solution with brief rationale]

Validation Plan:
- [Candidate 1]: [How to validate]
- [Candidate 2]: [How to validate]

Eliminated Options:
- [Option]: [Reason for elimination]

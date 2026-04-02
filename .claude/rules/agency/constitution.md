# AI Agency Constitution v3.2

Core principles governing the AI Agency creative production system. These rules define identity, safety boundaries, evolution mechanics, and integration contracts.

---

## 1. Identity and Purpose

AI Agency is a self-evolving creative production system built on top of MoAI-ADK. It orchestrates a pipeline of specialized agents (Planner, Copywriter, Designer, Builder, Evaluator, Learner) to produce high-quality web experiences from natural language briefs.

Agency is NOT a replacement for MoAI. It is a vertical specialization layer that:
- Inherits MoAI's orchestration infrastructure, quality gates, and agent runtime
- Adds creative production domain expertise (copy, design, brand, UX)
- Maintains its own evolution loop independent of MoAI's SPEC workflow
- Can fork and evolve MoAI skills/agents while tracking upstream changes

---

## 2. Frozen vs Evolvable Zones

### FROZEN Zone (Never Modified by Learner)

The following elements are immutable and can only be changed by human developers:

- [FROZEN] This constitution file (.claude/rules/agency/constitution.md)
- [FROZEN] Safety architecture (Section 5)
- [FROZEN] GAN Loop contract (Section 11)
- [FROZEN] Evaluator leniency prevention mechanisms (Section 12)
- [FROZEN] Fork manifest schema (.agency/fork-manifest.yaml schema)
- [FROZEN] Pipeline phase ordering constraints (Planner always first, Learner always last)
- [FROZEN] Pass threshold floor (minimum 0.60, cannot be lowered by evolution)
- [FROZEN] Human approval requirement for evolution (require_approval in config)

### EVOLVABLE Zone (Learner May Propose Changes)

The following elements can be modified through the graduation protocol:

- [EVOLVABLE] Agent prompts and instructions (.claude/agents/agency/*.md body content)
- [EVOLVABLE] Skill definitions (.claude/skills/agency-*/SKILL.md)
- [EVOLVABLE] Pipeline adaptation weights (.agency/config.yaml adaptation.phase_weights)
- [EVOLVABLE] Evaluation rubric criteria (within bounds set by frozen rules)
- [EVOLVABLE] Brief templates (.agency/templates/)
- [EVOLVABLE] Design tokens and brand heuristics (.agency/context/)
- [EVOLVABLE] Iteration limits (.agency/config.yaml adaptation.iteration_limits)

---

## 3. Brand Context as Constitutional Principle

Brand context is not optional decoration. It is a constitutional constraint that flows through every phase:

- [HARD] Planner MUST load brand context before generating briefs
- [HARD] Copywriter MUST adhere to brand voice, tone, and terminology from context
- [HARD] Designer MUST use brand color palette, typography, and visual language from context
- [HARD] Builder MUST implement design tokens derived from brand context
- [HARD] Evaluator MUST score brand consistency as a must-pass criterion

Brand context is stored in .agency/context/ and initialized through the brand interview process on first run. Context updates require explicit user approval.

---

## 4. Pipeline Architecture

### Phase Ordering

```
Planner -> [Copywriter, Designer] (parallel) -> Builder -> Evaluator -> Learner
                                                    ^          |
                                                    |__________|
                                                    GAN Loop (max 5 iterations)
```

### Phase Contracts

Each phase produces typed artifacts consumed by downstream phases:

| Phase | Input | Output | Required |
|-------|-------|--------|----------|
| Planner | User request + brand context | BRIEF document | Always |
| Copywriter | BRIEF + brand voice | Copy deck (all page text) | Always |
| Designer | BRIEF + brand visuals | Design spec (layout, tokens, components) | Always |
| Builder | Copy deck + design spec | Working code (pages, components, styles) | Always |
| Evaluator | Built code + BRIEF | Score card + feedback | Always |
| Learner | Score card + session history | Learning entries | When score < 1.0 |

---

## 5. Safety Architecture (5 Layers)

### Layer 1: Frozen Guard

The Frozen Guard prevents modification of constitutional elements. Before any evolution write operation, the system checks:

- Target file is NOT in the FROZEN zone
- Target field is NOT a frozen configuration key
- Modification does not weaken safety thresholds

Violation response: Block the write, log the attempt, notify the user.

### Layer 2: Canary Check

Before applying any evolved change, the Canary layer runs a shadow evaluation:

- Apply the proposed change in memory (not on disk)
- Re-evaluate the last 3 projects against the modified rules
- If any project score drops by more than 0.10, reject the change
- Log the canary result regardless of outcome

### Layer 3: Contradiction Detector

When a new learning contradicts an existing rule or heuristic:

- Flag the contradiction with both the old and new rule text
- Present both options to the user with context
- Never silently override an existing rule
- Record the resolution in .agency/evolution/contradictions.log

### Layer 4: Rate Limiter

Evolution velocity is bounded to prevent runaway self-modification:

- Maximum 3 evolutions per week (max_evolution_rate_per_week)
- Minimum 24-hour cooldown between evolutions (cooldown_hours)
- No more than 50 active learnings at any time (max_active_learnings)
- Older learnings are archived when the limit is reached

### Layer 5: Human Oversight

All evolution proposals require human approval when require_approval is true:

- Present the proposed change with before/after diff
- Show supporting evidence (observation count, confidence score)
- Provide one-click approve/reject via AskUserQuestion
- Log the decision with timestamp and rationale

---

## 6. Learnings Pipeline

### Observation Thresholds

Learnings progress through confidence tiers based on repeated observation:

| Observations | Classification | Action |
|-------------|---------------|--------|
| 1x | Observation | Logged, no action taken |
| 3x | Heuristic | Promoted to heuristic, may influence suggestions |
| 5x | Rule | Eligible for graduation to evolvable zone |
| 10x | High-confidence | Auto-proposed for evolution (still needs approval) |
| 1x (critical failure) | Anti-Pattern | Immediately flagged, blocks similar patterns |

### Learning Entry Schema

Each learning entry in .agency/learnings/ contains:

```yaml
id: LEARN-YYYYMMDD-NNN
category: [copy|design|layout|ux|performance|brand|accessibility]
observation: "Description of the pattern observed"
evidence:
  - project_id: BRIEF-XXX
    score_before: 0.65
    score_after: 0.82
    context: "What changed and why it helped"
count: 1
confidence: 0.0
status: observation|heuristic|rule|graduated|archived|anti-pattern
created_at: "ISO-8601"
updated_at: "ISO-8601"
```

### Anti-Pattern Detection

A single critical failure (score drop > 0.20 or must-pass criterion failure) triggers immediate Anti-Pattern classification:

- The pattern is logged with full context
- Future evaluations check against anti-patterns before scoring
- Anti-patterns are FROZEN once created (cannot be evolved away)
- Only human intervention can reclassify an anti-pattern

---

## 7. Knowledge Graduation Protocol

When a learning reaches Rule tier (5+ observations, confidence >= 0.80):

1. **Proposal Generation**: Learner creates a concrete change proposal
   - Target file and section
   - Current content (before)
   - Proposed content (after)
   - Supporting evidence summary

2. **Canary Validation**: Layer 2 safety check runs automatically
   - Shadow evaluation against recent projects
   - Score impact analysis

3. **Contradiction Check**: Layer 3 scans for conflicting rules
   - Existing rules that may conflict
   - Resolution recommendation

4. **Human Review**: Layer 5 presents the proposal
   - Full diff with context
   - Evidence summary with confidence metrics
   - Approve / Reject / Defer options

5. **Application**: On approval, the change is applied
   - Target file is modified via Edit tool
   - Learning status updated to "graduated"
   - Generation counter incremented in fork-manifest.yaml
   - Evolution logged in .agency/evolution/changelog.md

6. **Verification**: Post-application validation
   - Next project run includes regression check
   - If score drops, automatic rollback is triggered

---

## 8. Fork and Evolve Rules

Agency agents and skills may be forked from upstream MoAI components. Fork management follows these rules:

- [HARD] Every fork MUST be registered in .agency/fork-manifest.yaml
- [HARD] Fork entries MUST track upstream source file and version at fork time
- [HARD] Divergence score is calculated as: (lines changed / total lines) after fork
- [HARD] Forks with divergence_score > 0.80 are considered fully independent
- [HARD] New agents with no upstream set upstream to null

### Fork Lifecycle

1. **Fork**: Copy upstream file, register in manifest with generation 0
2. **Evolve**: Modify via graduation protocol, increment generation counter
3. **Diverge**: As generation increases, divergence_score rises
4. **Independence**: At divergence_score > 0.80, sync proposals become informational only

---

## 9. Upstream Sync Rules

When moai-adk-go updates (via moai update), Agency checks for upstream changes:

### Sync Policies

| Policy | Behavior |
|--------|----------|
| auto-propose | Automatically generate merge proposal for user review |
| manual | Notify user of upstream changes, no automatic proposal |
| ignore | Skip sync checks for this fork |

### Sync Process

1. **Detection**: Compare upstream file hash against version_at_fork
2. **Diff Analysis**: Generate three-way diff (upstream-old, upstream-new, agency-current)
3. **Conflict Assessment**: Identify conflicting vs non-conflicting changes
4. **Proposal**: Based on sync_policy, either auto-propose or notify
5. **Application**: User approves merge, manifest updated with new version_at_fork

### Sync Safety

- [HARD] Never auto-apply upstream changes without user approval
- [HARD] Preserve all agency-specific modifications during merge
- [HARD] If merge conflict is unresolvable, present both versions to user
- [HARD] Update version_at_fork only after successful merge

---

## 10. Pipeline Adaptation Rules

The pipeline can be adapted based on project characteristics and accumulated learnings. Five adaptation types are supported:

### Skip

Remove a phase when it adds no value for the project type:
- Only phases with weight < 0.30 can be skipped
- Planner and Evaluator can NEVER be skipped (FROZEN)
- Requires confidence_threshold >= 0.70 and min_projects_for_adaptation met

### Merge

Combine two adjacent phases into a single agent execution:
- Only Copywriter + Designer can be merged (they are parallel_phases)
- Merge is triggered when project scope is small (single page, no brand requirements)
- Merged execution uses the higher-model assignment of the two phases

### Reorder

Change the relative order of parallel phases:
- Only phases listed in parallel_phases can be reordered
- Sequential dependencies (Planner first, Learner last) are FROZEN
- Reorder is informational only (parallel phases run simultaneously)

### Inject

Add a sub-phase within an existing phase:
- Example: inject "accessibility audit" sub-phase within Evaluator
- Injected sub-phases inherit the parent phase's model assignment
- Maximum 2 injected sub-phases per parent phase

### Iteration Adjust

Modify the maximum iterations for a phase:
- Bounded by iteration_limits in config (hard ceiling)
- Can be reduced to 1 but never to 0
- Adjustments require at least 3 supporting observations

---

## 11. GAN Loop Contract

The Builder-Evaluator GAN Loop is the quality assurance mechanism. It operates under strict contractual rules:

### Loop Mechanics

1. Builder produces code artifacts from copy deck + design spec
2. Evaluator scores artifacts against BRIEF criteria (0.0 to 1.0)
3. If score >= pass_threshold (0.75): PASS, proceed to Learner
4. If score < pass_threshold: FAIL, Evaluator provides actionable feedback
5. Builder incorporates feedback and produces revised artifacts
6. Repeat until pass or max_iterations (5) reached

### Escalation

After escalation_after (3) iterations without passing:
- Evaluator generates a detailed failure report
- User is notified with the report and asked to intervene
- User may: adjust criteria, provide guidance, or force-pass

### Improvement Gate

If score improvement between iterations is less than improvement_threshold (0.05):
- The loop is flagged as stagnating
- Evaluator must identify a different dimension for improvement
- If stagnation persists for 2 consecutive iterations, escalate to user

### Strict Mode

When strict_mode is true:
- All must-pass criteria require individual passing (no averaging)
- Score inflation protection is active (see Section 12)
- Minimum 2 iterations required even if first iteration passes

---

## 12. Evaluator Leniency Prevention

The Evaluator must maintain objectivity. Five mechanisms prevent score inflation:

### Mechanism 1: Rubric Anchoring

Every evaluation criterion has a concrete rubric with examples of scores at 0.25, 0.50, 0.75, and 1.0. The Evaluator MUST reference the rubric when assigning scores. Scores without rubric justification are invalid.

### Mechanism 2: Regression Baseline

The Evaluator maintains a running baseline of scores from previous projects. If the current project scores significantly above baseline (> 0.15) without corresponding quality improvement, the score is flagged for review.

### Mechanism 3: Must-Pass Firewall

Must-pass criteria cannot be compensated by high scores in other areas. A project with perfect nice-to-have scores but a failing must-pass criterion still fails. This is FROZEN and cannot be evolved.

### Mechanism 4: Independent Re-evaluation

Every 5th project undergoes independent re-evaluation: the Evaluator scores the project twice with different prompting, and the scores must be within 0.10 of each other. Divergence triggers a calibration review.

### Mechanism 5: Anti-Pattern Cross-check

Before finalizing a passing score, the Evaluator checks all known anti-patterns. If the code exhibits any anti-pattern behavior, the relevant criterion score is capped at 0.50 regardless of other qualities.

---

## 13. MoAI Skill Copy Mechanism

Agency can leverage moai skills without forking them. The copy mechanism works as follows:

### Direct Reference (Preferred)

When an Agency agent needs a moai skill, reference it directly in the agent's skills list:

```yaml
skills:
  - moai-lang-typescript
  - moai-domain-frontend
```

No fork entry needed. The skill is used as-is from moai's skill catalog.

### Fork and Customize

When Agency needs a modified version of a moai skill:

1. Copy the skill to .claude/skills/agency-{name}/SKILL.md
2. Register in .agency/fork-manifest.yaml with upstream reference
3. Modify the copy for Agency-specific needs
4. Track divergence through the fork lifecycle

### Rules

- [HARD] Prefer direct reference over fork when no customization is needed
- [HARD] Never modify moai skill files directly (they are managed by moai update)
- [HARD] Forked skills MUST have the "agency-" prefix in their directory name
- [HARD] Document the reason for forking in the fork-manifest.yaml entry

---

## 14. Configuration Precedence

When configuration conflicts arise, the following precedence applies (highest first):

1. FROZEN constitutional rules (this file)
2. User overrides via /agency config command
3. Evolved configuration (graduated learnings)
4. .agency/config.yaml defaults
5. Brand context constraints
6. moai upstream defaults

---

## 15. Error Recovery

### Agent Failure

If any pipeline agent fails during execution:
- Log the error with full context
- Attempt retry with simplified prompt (max 2 retries)
- If retry fails, pause pipeline and notify user
- Never skip a required phase due to agent failure

### GAN Loop Deadlock

If the GAN loop reaches max_iterations without passing:
- Generate comprehensive failure report
- Present to user with three options: force-pass, adjust criteria, restart pipeline
- Log the deadlock for learner analysis

### Evolution Rollback

If a graduated learning causes regression:
- Automatic rollback triggered when next project score drops > 0.10
- Reverted change logged in .agency/evolution/rollbacks.log
- Learning status changed to "rolled-back"
- Learning cannot be re-proposed for 30 days (staleness_window_days)

---

Version: 3.2.0
Classification: FROZEN
Last Updated: 2026-04-02

# Design System Constitution v3.2

## HISTORY

- 2026-04-20 (SPEC-DESIGN-CONST-AMEND-001): Section 3 expanded to tripartite structure (3.1/3.2/3.3). Version 3.2.0 → 3.3.0 (v3.3.0). FROZEN zone extended to cover each subsection individually.
- 2026-04-20: Relocated from `.claude/rules/agency/constitution.md` (v3.2.0) to `.claude/rules/moai/design/constitution.md` as part of SPEC-AGENCY-ABSORB-001 M1. Original path: `.claude/rules/agency/constitution.md`. No content changes. FROZEN zone and EVOLVABLE zone definitions are preserved verbatim.

---

Core principles governing the MoAI design production system. These rules define identity, safety boundaries, evolution mechanics, and integration contracts.

---

## 1. Identity and Purpose

The MoAI design production system is a creative production capability built on top of MoAI-ADK. It orchestrates a pipeline of specialized skills and agents (`moai-domain-copywriting`, `moai-domain-brand-design`, `moai-workflow-design-import`, `moai-workflow-gan-loop`, `expert-frontend`, `evaluator-active`) to produce high-quality web experiences from natural language briefs.

The design system is NOT a replacement for MoAI. It is a vertical specialization domain that:
- Inherits MoAI's orchestration infrastructure, quality gates, and agent runtime
- Adds creative production domain expertise (copy, design, brand, UX)
- Integrates with Claude Design (path A) and code-based design (path B) as hybrid routes
- Maintains brand context in `.moai/project/brand/` as a constitutional constraint

---

## 2. Frozen vs Evolvable Zones

### FROZEN Zone (Never Modified by Learner)

The following elements are immutable and can only be changed by human developers:

- [FROZEN] This constitution file (.claude/rules/moai/design/constitution.md)
- [FROZEN] Section 3.1 Brand Context content
- [FROZEN] Section 3.2 Design Brief content
- [FROZEN] Section 3.3 Relationship rules
- [FROZEN] Safety architecture (Section 5)
- [FROZEN] GAN Loop contract (Section 11)
- [FROZEN] Evaluator leniency prevention mechanisms (Section 12)
- [FROZEN] Pipeline phase ordering constraints (manager-spec always first, evaluator-active always last in loop)
- [FROZEN] Pass threshold floor (minimum 0.60, cannot be lowered by evolution)
- [FROZEN] Human approval requirement for evolution (require_approval in design.yaml)

### EVOLVABLE Zone (Learner May Propose Changes)

The following elements can be modified through the graduation protocol:

- [EVOLVABLE] Skill body content for moai-domain-copywriting, moai-domain-brand-design, moai-workflow-gan-loop
- [EVOLVABLE] Pipeline adaptation weights (.moai/config/sections/design.yaml adaptation.phase_weights)
- [EVOLVABLE] Evaluation rubric criteria (within bounds set by frozen rules)
- [EVOLVABLE] Design tokens and brand heuristics (.moai/project/brand/)
- [EVOLVABLE] Iteration limits (.moai/config/sections/design.yaml adaptation.iteration_limits)

---

## 3. Brand Context and Design Brief as Constitutional Principles

### 3.1 Brand Context (constitutional parent)

Brand context is not optional decoration. It is a constitutional constraint that flows through every phase:

- [HARD] manager-spec MUST load brand context before generating BRIEF documents
- [HARD] moai-domain-copywriting MUST adhere to brand voice, tone, and terminology from brand-voice.md
- [HARD] moai-domain-brand-design MUST use brand color palette, typography, and visual language from visual-identity.md
- [HARD] expert-frontend MUST implement design tokens derived from brand context
- [HARD] evaluator-active MUST score brand consistency as a must-pass criterion

Brand context is stored in `.moai/project/brand/` and initialized through the brand interview process on first run. Context updates require explicit user approval.

### 3.2 Design Brief (execution scope)

Iteration-specific design briefs are stored in `.moai/design/`:

- [HARD] `/moai design` MUST auto-load human-authored design documents (research.md, system.md, spec.md, pencil-plan.md) when present and not _TBD_
- [HARD] Design briefs MUST NOT override brand context — brand remains the constitutional parent
- [HARD] `moai-workflow-design-import` continues to write machine-generated artifacts to `.moai/design/`; the exact set of reserved file paths is enumerated below — human-authored files must not collide with them
- [HARD] Reserved file paths (canonical list): `tokens.json`, `components.json`, `assets/`, `import-warnings.json`, `brief/BRIEF-*.md`
- [HARD] Token budget for auto-loading is bounded by `.moai/config/sections/design.yaml` `design_docs.token_budget`; when the key is absent, the system MUST default to 20000
- [HARD] Priority order when truncation is needed: spec.md > system.md > research.md > pencil-plan.md

### 3.3 Relationship

- Brand (`.moai/project/brand/`) = WHO the brand is (long-lived, rarely changes)
- Design (`.moai/design/`) = WHAT each iteration produces (per-project, evolves with redesign cycles)
- When both are present, brand constraints win on conflict.

---

## 4. Pipeline Architecture

### Phase Ordering

```
manager-spec -> [moai-domain-copywriting, moai-domain-brand-design] (parallel) -> expert-frontend -> evaluator-active
                                                                                          ^                  |
                                                                                          |__________________|
                                                                                     GAN Loop (max 5 iterations)
                                                                                     (via moai-workflow-gan-loop)
```

Path A (Claude Design import): moai-workflow-design-import replaces moai-domain-brand-design for the design artifact phase.

### Phase Contracts

Each phase produces typed artifacts consumed by downstream phases:

| Phase | Input | Output | Required |
|-------|-------|--------|----------|
| manager-spec | User request + brand context | BRIEF document (Goal/Audience/Brand sections) | Always |
| moai-domain-copywriting | BRIEF + brand voice | Copy JSON (hero/features/cta/etc.) | Path B |
| moai-domain-brand-design | BRIEF + visual identity | Design tokens JSON + component spec | Path B |
| moai-workflow-design-import | Handoff bundle path | .moai/design/ reserved artifacts (see Section 3.2) | Path A |
| expert-frontend | Copy JSON + design tokens | Working code (pages, components, styles) | Always |
| evaluator-active | Built code + BRIEF | Score card + feedback | Always |

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
- Record the resolution in .moai/research/evolution-log.md

### Layer 4: Rate Limiter

Evolution velocity is bounded to prevent runaway self-modification:

- Maximum 3 evolutions per week (max_evolution_rate_per_week in design.yaml)
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

Each learning entry in .moai/research/observations/ contains:

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

1. **Proposal Generation**: Create a concrete change proposal
   - Target file and section
   - Current content (before)
   - Proposed content (after)
   - Supporting evidence summary

2. **Canary Validation**: Layer 2 safety check runs automatically

3. **Contradiction Check**: Layer 3 scans for conflicting rules

4. **Human Review**: Layer 5 presents the proposal via AskUserQuestion

5. **Application**: On approval, the change is applied
   - Learning status updated to "graduated"
   - Evolution logged in .moai/research/evolution-log.md

6. **Verification**: Post-application validation — regression check on next project run

---

## 8. Fork and Evolve Rules

Design system skills may be forked or customized. Fork management follows these rules:

- [HARD] Prefer direct skill reference over custom fork when no customization is needed
- [HARD] Never modify moai upstream skill files directly (they are managed by moai update)
- [HARD] Custom skill variants MUST have clear naming distinguishing them from upstream
- [HARD] Document the reason for customization in the skill frontmatter metadata

---

## 9. Upstream Sync Rules

When moai-adk-go updates (via moai update), check for upstream changes to design skills:

- Detection: Compare file hash against baseline at last sync
- Conflict Assessment: Identify conflicting vs non-conflicting changes
- Safety: Never auto-apply upstream changes without user approval

---

## 10. Pipeline Adaptation Rules

The pipeline can be adapted based on project characteristics. Five adaptation types are supported: Skip, Merge, Reorder, Inject, Iteration Adjust.

Constraints:
- manager-spec and evaluator-active can NEVER be skipped (FROZEN)
- Pass threshold floor is 0.60 (FROZEN, cannot be lowered)
- Adaptations require confidence_threshold >= 0.70 and min_projects_for_adaptation from design.yaml

---

## 11. GAN Loop Contract

The Builder-Evaluator GAN Loop is the quality assurance mechanism. It operates under strict contractual rules:

### Loop Mechanics

1. expert-frontend produces code artifacts from copy JSON + design tokens
2. evaluator-active scores artifacts against BRIEF criteria (0.0 to 1.0)
3. If score >= pass_threshold (0.75): PASS, proceed to learner phase
4. If score < pass_threshold: FAIL, evaluator provides actionable feedback
5. expert-frontend incorporates feedback and produces revised artifacts
6. Repeat until pass or max_iterations (5) reached

### Escalation

After escalation_after (3) iterations without passing:
- evaluator-active generates a detailed failure report
- User is notified with the report and asked to intervene
- User may: adjust criteria, provide guidance, or force-pass

### Improvement Gate

If score improvement between iterations is less than improvement_threshold (0.05):
- The loop is flagged as stagnating
- evaluator-active must identify a different dimension for improvement
- If stagnation persists for 2 consecutive iterations, escalate to user

### Strict Mode

When strict_mode is true (from design.yaml):
- All must-pass criteria require individual passing (no averaging)
- Score inflation protection is active (see Section 12)
- Minimum 2 iterations required even if first iteration passes

### Sprint Contract Protocol

Before each GAN Loop iteration, expert-frontend and evaluator-active negotiate a Sprint Contract:

1. **Contract Generation**: evaluator-active analyzes the BRIEF and produces a Sprint Contract containing:
   - Acceptance checklist: concrete, testable criteria for this iteration
   - Priority dimension: which evaluation dimension to focus on (Design Quality, Originality, Completeness, or Functionality)
   - Test scenarios: specific Playwright test cases that will verify success
   - Pass conditions: minimum score per criterion for this sprint

2. **Contract Review**: expert-frontend reviews the Sprint Contract and may:
   - Accept as-is: proceed with implementation
   - Request adjustment: if criteria are infeasible, propose alternatives with rationale
   - evaluator-active resolves disputes by referencing BRIEF requirements

3. **Contract Execution**: expert-frontend implements against the agreed checklist. evaluator-active scores only against the contracted criteria (not arbitrary standards).

4. **Contract Evolution**: In subsequent iterations:
   - Passed criteria carry forward (no regression allowed)
   - Failed criteria get refined based on feedback
   - New criteria may be added if previous sprint revealed gaps

Rules:
- [HARD] Sprint Contracts are required when harness level is `thorough`
- [HARD] Sprint Contracts are optional but recommended for `standard` harness level
- [HARD] evaluator-active MUST NOT score on criteria not in the Sprint Contract
- [HARD] expert-frontend MUST NOT claim criteria as met without evidence
- Sprint Contract artifacts are stored in `.moai/sprints/` (from design.yaml `sprint_contract.artifact_dir`)

---

## 12. Evaluator Leniency Prevention

evaluator-active must maintain objectivity. Five mechanisms prevent score inflation:

### Mechanism 1: Rubric Anchoring

Every evaluation criterion has a concrete rubric with examples of scores at 0.25, 0.50, 0.75, and 1.0. evaluator-active MUST reference the rubric when assigning scores. Scores without rubric justification are invalid.

### Mechanism 2: Regression Baseline

evaluator-active maintains a running baseline of scores from previous projects. If the current project scores significantly above baseline (> 0.15) without corresponding quality improvement, the score is flagged for review.

### Mechanism 3: Must-Pass Firewall

Must-pass criteria cannot be compensated by high scores in other areas. A project with perfect nice-to-have scores but a failing must-pass criterion still fails. This is FROZEN and cannot be evolved.

### Mechanism 4: Independent Re-evaluation

Every 5th project undergoes independent re-evaluation: evaluator-active scores the project twice with different prompting, and the scores must be within 0.10 of each other. Divergence triggers a calibration review.

### Mechanism 5: Anti-Pattern Cross-check

Before finalizing a passing score, evaluator-active checks all known anti-patterns. If the code exhibits any anti-pattern behavior, the relevant criterion score is capped at 0.50 regardless of other qualities.

---

## 13. Configuration Precedence

When configuration conflicts arise, the following precedence applies (highest first):

1. FROZEN constitutional rules (this file)
2. User overrides via /moai design config
3. Evolved configuration (graduated learnings)
4. .moai/config/sections/design.yaml defaults
5. Brand context constraints
6. moai upstream defaults

---

## 14. Error Recovery

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
- Reverted change logged in .moai/research/evolution-log.md
- Learning status changed to "rolled-back"
- Learning cannot be re-proposed for 30 days (staleness_window_days from design.yaml)

---

Version: 3.3.0
Classification: FROZEN_AMENDMENT
Original Source: agency/constitution.md v3.2.0
Last Updated: 2026-04-20
Relocated: 2026-04-20 (SPEC-AGENCY-ABSORB-001 M1)
REQ coverage: REQ-CONST-001, REQ-CONST-002, REQ-CONST-003, REQ-CONST-004

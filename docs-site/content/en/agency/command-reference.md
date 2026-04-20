---
title: Command Reference
weight: 50
draft: false
---

{{< callout type="info" >}}
All Agency commands start with `/agency`. Use natural language for quick results, or add flags for more control.
{{< /callout >}}

## Command Overview

| Command | Purpose | Speed | Control | Example |
|---------|---------|-------|---------|---------|
| `/agency` | Natural language mode | Auto | Minimal | `/agency Build a landing page for my AI startup` |
| `/agency brief` | Create new brief | 2 min | Guided | `/agency brief "SaaS onboarding flow"` |
| `/agency build` | Execute full pipeline | 15-30 min | Flags | `/agency build BRIEF-001 --team` |
| `/agency review` | Standalone quality check | 5 min | Focused | `/agency review BRIEF-001` |
| `/agency learn` | Collect feedback | 10 min | Interactive | `/agency learn` |
| `/agency evolve` | Trigger skill evolution | 5-15 min | Targeted | `/agency evolve --agent copywriter` |
| `/agency resume` | Continue interrupted build | Auto | Recovery | `/agency resume BRIEF-001` |
| `/agency profile` | View evolution status | 1 min | View-only | `/agency profile` |
| `/agency sync-upstream` | Check moai-adk updates | 2 min | View-only | `/agency sync-upstream` |
| `/agency rollback` | Revert skill generation | 2 min | Targeted | `/agency rollback copywriter gen-010` |
| `/agency config` | View/edit settings | Instant | Admin | `/agency config` |

---

## Natural Language Mode (Just Do It)

The simplest way to use Agency: describe what you want in natural language.

```
/agency Build a landing page for an AI research platform 
targeting academic institutions, emphasizing data collaboration 
and team productivity
```

Agency will:
1. Ask clarifying questions about brand, audience, and success metrics
2. Create a BRIEF automatically
3. Execute the full pipeline
4. Deploy live website
5. Show URL when complete

**Advantages:**
- Fastest path (2-3 questions, then automated)
- No need to learn flags
- Ideal for first-time users

**Disadvantages:**
- Less control over specifications
- Can't review BRIEF before building
- All-or-nothing execution

---

## `/agency brief` - Create New Brief

Create a detailed project specification from your description.

### Basic Usage

```
/agency brief "AI startup landing page for research labs"
```

### Interactive Flow

When you submit a brief, Agency asks clarifying questions:

**Question 1: Business Goals**
"What's the primary success metric for this project?"
- Options: Signup conversion, Trial signups, Lead generation, Brand awareness

**Question 2: Target Audience**
"Who's the main decision-maker visiting this site?"
- Options: Researcher, Lab Manager, University Administrator, Grant Manager

**Question 3: Brand Context**
"Do you have existing brand guidelines?"
- Options: Yes (provide link), No (create basic profile), Using template

**Question 4: Timeline & Budget**
"When do you need this live?"
- Options: ASAP (2 weeks), Standard (4 weeks), Flexible (8+ weeks)

### Output

Creates `.agency/briefs/BRIEF-001/spec.md` containing:

```markdown
# BRIEF-001: AI Research Platform Landing

## Goal
Increase trial signups from academic institutions by 15% within 90 days

## Outcome
Professional landing page deployed in 2 weeks, generating 
qualified leads for B2B sales team

## Context
Brand: Technical, trustworthy, innovation-focused
Audience: Research directors (PhD, 40-60 years old)
Technical: Next.js, PostgreSQL, Vercel deployment

## Requirements
1. Hero section with value proposition
2. Feature showcase (3-5 key benefits)
3. Social proof (customer testimonials)
4. Pricing/plan options
5. Contact form and demo CTA
6. Mobile-responsive design
7. SEO optimization

## Acceptance Criteria
- Mobile score: 90+
- Load time: < 2 seconds
- Lighthouse: 90+ across all metrics
- Accessibility: WCAG 2.1 AA
- Conversion: Form completion rate > 5%
```

---

## `/agency build` - Execute Pipeline

Run the full pipeline to transform a BRIEF into a live website.

### Basic Usage

```
/agency build BRIEF-001
```

Executes: Planner → Copywriter → Designer → Builder → Evaluator → Deploy

### Step-by-Step Mode

Review output after each phase before proceeding:

```
/agency build --step BRIEF-001
```

**Phase 1: Requirement Expansion**
Planner generates detailed specifications

**Review Output**
- Check requirements are complete
- Verify acceptance criteria are testable
- Approve or request modifications

**Phase 2: Creative**
Copywriter and Designer create content and UI specs (parallel)

**Review Output**
- Review copy alignment with brand
- Check design against brand palette
- Approve or request changes

**Phase 3: Implementation**
Builder implements code with 85%+ test coverage

**Review Output**
- Check implementation matches design
- Review code structure
- Verify tests are comprehensive

**Phase 4: Quality Check**
Evaluator runs automated tests

**Review Output**
- Check all criteria pass
- Review performance metrics
- Approve deployment

### Team Parallel Mode

Spawn multiple agents for faster execution:

```
/agency build --team BRIEF-001
```

Copywriter and Designer work in parallel (worktree isolation). Reduces execution time by 30-40% compared to sequential.

### Resume Interrupted Build

If build interrupted, resume from last completed phase:

```
/agency build BRIEF-001 --resume
```

---

## `/agency review` - Standalone Quality Check

Run only the evaluator phase without building.

```
/agency review BRIEF-001
```

Useful for:
- Testing code quality before deployment
- Checking if a previous build meets criteria
- Validating performance metrics
- Verifying accessibility compliance

### Output

Detailed quality report:

```
Quality Report for BRIEF-001
────────────────────────────

Design Quality:        92% ✓ (Target: 85%)
Code Quality:          88% ✓ (Target: 85%)
Performance:           94% ✓ (Target: 85%)
Accessibility:         90% ✓ (Target: 85%)

Overall Score: 91% ✓ PASS

Detailed Results:
─────────────────

Design Quality (92%)
  ✓ Brand alignment: 95%
  ✓ Visual hierarchy: 90%
  ✓ Spacing consistency: 91%
  ✓ Typography: 90%

Code Quality (88%)
  ✓ Test coverage: 87% (Target: 85%)
  ✓ Type safety: 92%
  ✓ Documentation: 84% (Warning: below target)
  ✓ Linting: 89%

Performance (94%)
  ✓ Load time: 1.2s (Target: < 2s)
  ✓ Lighthouse: 96
  ✓ Core Web Vitals: PASS
  ✓ Mobile score: 92

Accessibility (90%)
  ✓ WCAG 2.1 AA: PASS
  ✓ Keyboard navigation: PASS
  ✓ Screen reader: 89%
  ✓ Color contrast: PASS

Recommendations:
─────────────────
1. Increase documentation strings in 3 utility functions
2. Consider lazy-loading images below fold for better CLS
3. Add skip-to-content link for accessibility
```

---

## `/agency learn` - Feedback Collection

Collect feedback to teach the system about what works.

```
/agency learn
```

### Interactive Feedback Session

**Section 1: Overall Satisfaction**
"On a scale of 1-10, how happy are you with the result?"
- 1-4: Significant issues (triggers detailed improvement flow)
- 5-7: Good, with specific improvements
- 8-10: Excellent, capture what worked

**Section 2: What Worked Well**
"What did the system get right? Be specific."
- Examples: "The hero copy resonated with our audience", "Love the color palette", "Clean code structure"
- System captures patterns to replicate

**Section 3: What Needs Improvement**
"What should we change next time?"
- Examples: "Too much technical jargon in copy", "Colors don't match our brand", "Missing error handling"
- System identifies pain points to address

**Section 4: Design Feedback**
"How aligned is the design with your vision?"
- Layout: Too sparse / Just right / Too crowded
- Colors: Off brand / Close / Perfect
- Typography: Hard to read / Adequate / Excellent

**Section 5: Copy Feedback**
"How's the messaging resonating?"
- Tone: Matches brand (yes/no)
- Clarity: Clear vs Confusing
- CTAs: Compelling vs Generic

**Section 6: Technical Feedback**
"Any bugs or technical issues?"
- Performance: Any slow sections?
- Functionality: Anything broken?
- Mobile: Any layout issues?

### Output

Feedback stored in `.agency/feedback/BRIEF-001/feedback.md`:

```markdown
# Feedback for BRIEF-001

Overall Rating: 8/10
────────────────────

What Worked Well
────────────────
1. Hero copy perfectly captures our mission
2. Feature descriptions are benefit-focused (not feature-focused)
3. Color palette feels premium and trustworthy
4. Code is clean and well-documented
5. Mobile experience is flawless

Needs Improvement
─────────────────
1. Testimonials section feels generic (need research-specific quotes)
2. Pricing strategy unclear (should we show enterprise plan?)
3. Legal footer missing GDPR/CCPA privacy information
4. CTA button color should match brand blue, not default

Design Feedback
───────────────
- Layout: Perfect
- Colors: Need adjustment (one color)
- Typography: Excellent

Copy Feedback
─────────────
- Tone: Matches brand
- Clarity: Clear
- CTAs: Compelling

Technical Feedback
──────────────────
- No bugs found
- Mobile: Perfect
- Performance: Excellent
- Note: Contact form should redirect to Hubspot

Evolution Triggers
──────────────────
[System identifies]:
- Testimonials pattern: Need customer-specific quotes
- Legal compliance: GDPR/CCPA templates
- CTA optimization: Color choice matters
```

---

## `/agency evolve` - Trigger Skill Evolution

Manually trigger learning loop to improve agents.

```
/agency evolve
```

This analyzes ALL feedback across all projects and promotes high-confidence rules.

### Evolution by Agent

Evolve specific agents only:

```
/agency evolve --agent copywriter
```

Options: copywriter, designer, builder, evaluator, planner

### Evolution Dry-Run

Preview what would change without applying:

```
/agency evolve --dry-run
```

Shows:
- Rules ready for promotion
- Anti-patterns to block
- Confidence scores
- Next generation version numbers

### Output

Generates skill module updates:

```
Evolution Complete: Generation Updates
──────────────────────────────────────

Copywriter: v2.0 → v2.1
  New Rules:
    + Benefit-first messaging (confidence: 0.75)
    + Research institution persona patterns (0.72)
    + Video CTA for complex features (0.68)

Designer: v1.5 → v1.6
  New Rules:
    + Component spacing preferences (0.80)
    + Color palette customizations (0.71)
  Anti-Patterns:
    - Avoid generic stock photos (marked forbidden)

Builder: v2.0 → v2.1
  Improved:
    + Error boundary patterns (updated)
    + Database query optimization (refined)

Files Updated:
  .agency/skills/copywriter/gen-021/
  .agency/skills/designer/gen-016/
  .agency/skills/builder/gen-021/

Next Build will use new generations automatically.
```

---

## `/agency resume` - Continue Interrupted Work

Resume a build that was interrupted or failed.

```
/agency resume BRIEF-001
```

Picks up from the last completed phase:

```
Resume Status for BRIEF-001
───────────────────────────
Last Completed Phase: Designer (Phase 2)
Next Phase: Builder (Phase 3)
Time Elapsed: 5 minutes
Time Remaining: 10-15 minutes

Resume? [yes/no]
```

---

## `/agency profile` - View Evolution Status

See how Agency has learned from your projects.

```
/agency profile
```

### Output

```
AI Agency Evolution Profile
────────────────────────────

Projects Completed: 12
Feedback Sessions: 11
Active Rules: 18

Agent Generations
─────────────────
Copywriter:  v2.3 (11 rules, 8 active)
Designer:    v1.8 (6 rules, 5 active)
Builder:     v2.1 (7 rules, 6 active)
Evaluator:   v1.5 (3 weights, all active)
Planner:     v1.0 (no custom rules)

Learning Velocity
─────────────────
Projects/Week:         2.4
Rules Promoted/Month:  1.8
Confidence Gain:       +0.18/project
Anti-Patterns Found:   2

Skill Specialization
────────────────────
Most Confident Rules:
  1. Benefit-first messaging (0.92)
  2. Team collaboration CTAs (0.88)
  3. Research institution tone (0.85)

Custom Brand Voice
──────────────────
Copywriter has learned 3 custom patterns specific to
your brand vocabulary and audience preferences.

Next Evolution
──────────────
Ready for automatic evolution after next project.
Run: /agency evolve

Upstream Status
───────────────
moai-adk: Synchronized (v3.2.0)
Conflicts: 0
Last Sync: 2 days ago
Pending Updates: 0
```

---

## `/agency sync-upstream` - Check moai-adk Updates

Check for and merge updates from moai-adk core framework.

```
/agency sync-upstream
```

### Output

```
MoAI-ADK Upstream Check
───────────────────────

Current Version: moai-adk v3.2.0
Latest Available: moai-adk v3.2.1

Available Updates
─────────────────

1. moai-lang-python
   Changes: Type hint improvements, FastAPI patterns
   Conflicts: 0
   Recommendation: SAFE (no Agency modifications)

2. moai-workflow-testing
   Changes: New test patterns, coverage improvements
   Conflicts: 0
   Recommendation: SAFE (recommended)

3. moai-library-mermaid
   Changes: New diagram types, rendering optimization
   Conflicts: 0
   Recommendation: SAFE (recommended)

Summary
───────
3 updates available
0 conflicts detected
Recommendation: Apply all updates

Apply Updates? [yes/no]
```

---

## `/agency rollback` - Revert Skill Generation

Revert a skill module to a previous generation if evolution caused problems.

```
/agency rollback copywriter gen-010
```

### Rollback Criteria

Rollback when:
- New generation decreased quality (evaluator rejects more)
- New patterns contradict brand voice
- Anti-pattern incorrectly identified

### Output

```
Rollback: Copywriter gen-011 → gen-010
─────────────────────────────────────

Current: gen-011 (broken)
  Rules: 12 (9 working, 3 causing issues)
  Confidence Avg: 0.71
  Quality Impact: -4% (down from 92% to 88%)

Target: gen-010 (stable)
  Rules: 11 (all working)
  Confidence Avg: 0.75
  Quality Impact: +2% (was 92% when current)

Changes Discarded:
  - Over-aggressive benefit messaging (gen-011 only)
  - Removed testimonial strategy (gen-011 only)

Rollback Confirmed. Copywriter now using gen-010.
Next build will use gen-010 until next evolution.

Historical Note:
  Rollback recorded for learning: Why gen-011 failed?
  Next evolution will avoid similar patterns.
```

---

## `/agency config` - View/Edit Settings

View and modify Agency configuration.

```
/agency config
```

### Configuration Sections

**Brand Configuration**
```yaml
brand:
  name: "My Company"
  voice: "Professional yet approachable"
  colors:
    primary: "#0066CC"
    secondary: "#FF6B35"
  fonts:
    sans: "Inter"
    serif: "Merriweather"
```

**Project Settings**
```yaml
project:
  default_model: "sonnet"  # sonnet or opus
  max_iterations: 5
  deployment_target: "vercel"
  auto_deploy: true
```

**Evolution Settings**
```yaml
evolution:
  auto_promote: true        # Automatically promote rules at thresholds
  confidence_threshold: 0.7 # Minimum confidence to apply
  safety_override: false    # Require human approval for all changes
```

**Integrations**
```yaml
integrations:
  github: "org/repo"        # For deploying code
  hubspot: true             # For form submissions
  analytics: "posthog"      # Analytics provider
```

---

## GAN Loop Details

When evaluator rejects output (quality < 80%), learner agent analyzes failure and improves:

### GAN Loop Iterations

**Iteration 1:** Evaluator identifies issue (e.g., "Copy too technical for audience")

**Iteration 2:** Learner recommends improvement ("Simplify jargon, use analogies")

**Iteration 3:** Generator (Copywriter) revises based on feedback

**Iteration 4:** Evaluator re-tests

**Repeat** until quality passes or max iterations reached

### Iteration Limits

- Simple projects: 3 iterations max
- Standard projects: 5 iterations max (default)
- Complex projects: 7 iterations max

### Escalation

If max iterations exceeded without passing:

1. Log issue to `.agency/escalations/`
2. Send summary to user
3. Require manual intervention or project pause
4. Mark as anti-pattern for future evolution

---

## Configuration Reference

Key configuration locations:

| Path | Purpose | Frequency |
|------|---------|-----------|
| `.agency/config.yaml` | Main settings | Never (modify with `/agency config`) |
| `.agency/context/` | Brand constitution | Rarely (when brand changes) |
| `.agency/briefs/` | Project specs | Every project |
| `.agency/feedback/` | Learning data | Every project |
| `.agency/learnings/` | Pattern database | Auto-updated |
| `.agency/skills/` | Agent modules | Auto-evolving |

---

## Data Locations

| Location | Content | Access |
|----------|---------|--------|
| `.agency/briefs/BRIEF-XXX/` | Project specifications | Read |
| `.agency/feedback/BRIEF-XXX/` | User feedback | Read |
| `.agency/learnings/` | Pattern database | View only (modified by evolution) |
| `.agency/skills/` | Agent skill modules | View only (auto-evolving) |
| `.agency/escalations/` | Failed builds & issues | Read |
| `.agency/generations/` | Historical skill versions | Archive |
| `.agency/deployed/` | Live site URLs and configs | Read |

---

## Next Steps

Ready to start? Begin with `/agency brief "your project"` to create your first specification, then `/agency build BRIEF-001` to launch the pipeline.

Or explore the [Getting Started](/en/agency/getting-started) guide for a detailed walkthrough.

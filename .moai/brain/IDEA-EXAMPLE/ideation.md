# Idea: Habit-tracking application for knowledge workers
*Session: 2026-05-04*

> NOTE: This is a worked example showing the expected structure and quality of Phases 4 and 5 output.
> Replace with your actual ideation when running /moai brain with your own idea.

## Lean Canvas

### Problem
1. Habit tracking tools are disconnected from actual work outcomes — users complete habits but don't know if they're producing results
2. Context-switching to dedicated habit apps breaks deep work focus for knowledge workers
3. Individual accountability lacks social reinforcement; team-based commitment tools don't exist at the lightweight level

### Customer Segments
Primary: Knowledge workers (developers, designers, writers, researchers) at technology companies who care about professional development alongside personal habits — approximately 50M globally in English-speaking markets.

Early adopters: Remote workers who lost structured daily routines post-pandemic, actively seeking productivity improvement tools.

### Unique Value Proposition
The only habit tracker that connects your daily routines to measurable work outcomes and team commitments — so you know your habits are actually working, not just completed.

### Solution
1. Outcome-linked habit tracking: Each habit is connected to a measurable goal metric (lines of code, pages written, calls made) — weekly review shows correlation
2. Focus-mode integrations: Native integrations surface habit check-ins inside existing tools (Notion sidebar, browser extension) to eliminate context-switching
3. Accountability pods: Small groups (3-7 people) share habit streaks and commit to weekly review calls — no corporate HR overhead

### Channels
- Integration-as-distribution: Notion/Linear plugin reaches their user base directly
- Product Hunt / Indie Hacker community launch
- Team referral: When one team member adopts, they invite 2-6 colleagues

### Revenue Streams
- Individual: Freemium (5 habits free, $8/month for unlimited + outcomes tracking)
- Team: $12/user/month for accountability pods + manager dashboards
- Enterprise: Custom pricing for HR integration and SSO

### Cost Structure
- Infrastructure: Time-series database, notification delivery, integration maintenance
- Support: Onboarding for enterprise segment
- Marketing: Primarily product-led; low acquisition cost via integrations

### Key Metrics
- D30 habit retention rate (target: >40%; industry average ~30%)
- Habit-to-outcome correlation score (novel metric — measures whether tracked habits correlate with user-reported outcomes)
- Pod activation rate (what % of users form or join accountability pods)
- NPS at day 60 (target: >50)

### Unfair Advantage
- Outcome correlation dataset: Aggregate anonymized data showing which habits actually correlate with professional outcomes — no competitor has this
- Integration-first distribution: Being embedded in productivity tools creates switching cost and distribution moat

---

## Evaluation Report

### Strengths
- Outcome correlation is a genuinely novel value proposition — no direct competitor claims this
- Integration-as-distribution reduces customer acquisition cost significantly
- Team accountability pod fills a real gap between individual apps and enterprise HR tools
- Research confirms behavior tracking + social accountability is the most effective habit formation combination

### Weaknesses
- Outcome correlation requires significant user data collection to be meaningful — cold start problem
- Integration maintenance is high-cost (API changes from Notion, Linear, etc.)
- 30-day retention is the hard problem for all habit apps; the solution needs a stronger hook
- Enterprise HR integration is complex; better to defer to v0.2

### First Principles Validation

Breaking down to fundamentals:
1. **Why do habits fail?** Insufficient friction reduction, lack of meaningful feedback, absence of social commitment
2. **What does tracking actually need to deliver?** Evidence that the behavior is producing desired outcomes
3. **Why do integrations create distribution?** Users already live in their tools; adding friction of new app increases abandonment
4. **Can outcome correlation be bootstrapped?** Yes — start with user self-reported outcome ratings, then use aggregate data once scale achieved

First principles verdict: The core insight (habits without outcome feedback are weak) is sound. The technical approach (integration-first, outcome correlation) is defensible. Main risk is execution complexity of the integration layer.

### Verdict
Proceed with caveats:
- Scope v0.1 strictly to: individual tracking, basic outcome linking (self-reported), 1-2 integrations maximum
- Defer: accountability pods, enterprise HR, advanced correlation analytics
- Validate before build: Confirm outcome linking increases D30 retention through a landing page experiment or waitlist survey

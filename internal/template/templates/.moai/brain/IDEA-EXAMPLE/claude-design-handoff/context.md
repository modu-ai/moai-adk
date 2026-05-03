# Extended Context: HabitScope

> This file supplements prompt.md with additional context for your design session.
> It is NOT meant to be pasted into Claude Design — use prompt.md for that.

## Product Background

### Lean Canvas Summary

**Problem**:
1. Habit tracking disconnected from outcomes
2. Context-switching friction to separate apps
3. No lightweight team accountability

**Customer Segments**: Knowledge workers, primarily developers and remote workers aged 25-40

**Unique Value Proposition**: The only habit tracker that connects daily routines to measurable work outcomes

**Solution**: Outcome-linked tracking + focus-mode integrations + accountability pods (v0.2)

**Channels**: Integration-as-distribution (Notion/Linear plugins), community launch

**Revenue**: Freemium individual ($8/month), team ($12/user/month)

**Key Metrics**: D30 retention >40%, habit-to-outcome correlation score, pod activation rate, NPS >50 at day 60

**Unfair Advantage**: Outcome correlation dataset; integration-first distribution moat

## Roadmap Context

The v0.1 scope includes these capabilities (from the SPEC decomposition):
1. Core habit definition, scheduling, and completion tracking
2. Outcome metric definition and habit-to-outcome correlation calculation
3. Weekly review workflow with streak visualization and correlation display
4. Browser extension for focus-mode check-in integration
5. User authentication, onboarding flow, and account management
6. Reminder notification system with configurable timing and channels

The design brief (prompt.md) focuses on: the main dashboard and daily check-in interface (items 1, 2, 3 above).

Out of scope for v0.1 design: team/pod features, mobile native, admin dashboards.

## Research Findings Summary

Key market insights:
- Habit tracking market: 40M+ active users; 25% YoY growth; most users are at consumer tier
- Primary pain point: "tracking without insight" — 68% of knowledge workers want better habit tools
- Behavior science: visual progress tracking increases completion by 3x; habit formation takes median 66 days
- Competitor gap: No existing tool explicitly connects habits to professional outcomes
- Risk: 70%+ 30-day abandonment is the industry's hardest problem

Competitive landscape: Habitica (too gamified), Streaks (too simple), Fabulous (wellness-only), Notion templates (too unstructured)

## Brand Context

> No brand voice defined for this project yet. The prompt.md section 3 contains placeholder suggestions.
> To define brand context: populate .moai/project/brand/brand-voice.md with your brand guidelines.
> Then regenerate the handoff package with /moai brain --regenerate IDEA-EXAMPLE.

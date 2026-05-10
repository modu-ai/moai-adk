# Proposal: HabitScope — Outcome-Linked Habit Tracker for Knowledge Workers
*Generated: 2026-05-04 | Idea: IDEA-EXAMPLE*

> NOTE: This is a worked example showing the expected structure of Phase 6 output.
> Replace with your actual proposal when running /moai brain with your own idea.
> Key REQ-BRAIN-008 check: No tech-stack names appear in the requirements sections below.

## Product Summary

HabitScope is a habit tracking platform for knowledge workers that explicitly connects daily routines to measurable professional outcomes. Unlike general-purpose habit apps, it surfaces correlation between completed habits and user-defined work metrics, reducing the "tracking without insight" problem that causes most users to abandon habit apps within 30 days.

## Target User

Knowledge workers — developers, designers, writers, and researchers — at technology companies who actively invest in professional development. Early adopters are remote workers who lost structured routines and seek a tool that proves whether their habits are producing results.

## Core Problems Solved

1. Habit tracking disconnected from outcomes: users complete habits but cannot verify they produce desired results
2. Context-switching friction: switching to a separate app breaks deep work focus
3. Lack of social reinforcement at the team level: no lightweight accountability option between solo apps and corporate HR tools

## Proposed Solution

A cross-platform habit tracking application with three core capabilities:

1. **Outcome-linked habit tracking**: Users attach each habit to a self-reported outcome metric (e.g., "deep work hours", "articles published", "customer calls"). Weekly review surfaces correlation trends.

2. **Focus-mode check-ins**: Habit completion is surfaced inside existing productivity tools via lightweight integration (browser extension, sidebar widget) — no separate app required for daily check-ins.

3. **Accountability pods (v0.2+)**: Small groups (3-7 people) share habit streaks and weekly commitment. Deferred to v0.2 due to social coordination complexity.

### SPEC Decomposition Candidates

- SPEC-HABIT-001: Core habit definition, scheduling, and completion tracking data model
- SPEC-OUTCOME-001: Outcome metric definition and habit-to-outcome correlation calculation
- SPEC-REVIEW-001: Weekly review workflow with streak visualization and correlation display
- SPEC-INTEGRATION-001: Browser extension for focus-mode check-in integration
- SPEC-AUTH-001: User authentication, onboarding flow, and account management
- SPEC-NOTIFY-001: Reminder notification system with configurable timing and channels

## Recommended Execution Order

1. SPEC-AUTH-001 (foundation — all other SPECs depend on user identity)
2. SPEC-HABIT-001 (core data model — prerequisite for outcome and review SPECs)
3. SPEC-OUTCOME-001 (differentiating feature — validates core value proposition)
4. SPEC-REVIEW-001 (retention feature — weekly review is the primary engagement hook)
5. SPEC-NOTIFY-001 (adoption feature — reminders drive habit completion rates)
6. SPEC-INTEGRATION-001 (distribution feature — enables integration-as-distribution strategy)

## Out of Scope (v0.1)

- Accountability pods and team features (v0.2)
- Native mobile applications (web-first; mobile via PWA)
- Enterprise HR integration and SSO
- Advanced ML-based correlation analysis (start with simple Pearson correlation on self-reported data)
- Calendar sync (v0.2 integration expansion)
- Public API for third-party integrations

## Notes

- Outcome correlation requires sufficient user data to be meaningful — plan a day-30 review gate in the product to evaluate whether the correlation feature is working
- Integration-as-distribution strategy requires securing partnerships or building distribution with Notion/Linear ecosystems — validate channel fit early
- Concurrent invocations of /moai brain not supported in v0.1; avoid running two brain sessions simultaneously in the same project

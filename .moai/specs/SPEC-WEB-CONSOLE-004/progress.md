# Progress — SPEC-WEB-CONSOLE-004

## Plan Phase

- **Tier**: M (3 artifacts: spec.md + plan.md + acceptance.md)
- **cycle_type (run-phase)**: tdd (per `development_mode: tdd`)
- **status**: draft
- **Cohort**: web-console-v3 visual-restyle member (004); siblings 001 (모태) / 002 S1 (completed) / 003 S2a (in-progress)
- **Primary source**: `.moai/design/web-console-handoff/IMPLEMENTATION-HANDOFF.md` (formalized into SPEC)
- **REQ count**: 12 (REQ-WC4-001 .. REQ-WC4-012)
- **AC count**: 13 (AC-WC4-001 .. AC-WC4-013)
- **MUST-PASS invariant**: zero server-contract change (spec.md §3, REQ-WC4-009)

### Plan-phase artifacts created

- spec.md — 12 GEARS REQs + 13 AC index + §4 Exclusions (E.1 web-i18n/langpick→S3, E.2 CJK font→S3, E.3 nested config→S2b, E.4 .review aside, E.5 anti-patterns)
- plan.md — Tier M justification + §A Context + §B Known Issues + §C Pre-flight + §D Constraints + §E Self-Verification + §F Milestones (M1 token/font → M2 layout/components [server-contract gate] → M3 dark-mode → M4 inline-SVG icons → M5 a11y/closure) + §G Anti-Patterns + §H Cross-References
- acceptance.md — full Given-When-Then for all 13 AC + §D edge cases (EC-1..EC-10) + §E Definition of Done + §F forward-looking S3/S2b checks

plan_complete_at: 2026-06-03
plan_status: audit-ready

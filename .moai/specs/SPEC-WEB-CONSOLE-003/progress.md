# Progress — SPEC-WEB-CONSOLE-003

> S2a of the `web-console-v3` cohort. Tier M. FLAT/SHALLOW project-config parity (development_mode + git_convention.convention).

## Plan-phase signal

plan_complete_at: 2026-06-03
plan_status: audit-ready

## Tier & artifacts

- Tier: M (3 artifacts: spec.md + plan.md + acceptance.md)
- Justification: new project-config persistence path crossing internal/web → internal/config boundary; ~6-7 files; dual-editor (web + TUI) × validate/widget/persist/4-locale/read-on-render → 11 ACs.

## Scope decisions recorded

- Confirmed flat settings: `development_mode` ({ddd,tdd}, exported predicate `models.ValidDevelopmentModes()`), `git_convention.convention` ({auto,conventional-commits,angular,karma,custom}, `pkg/models` oneof SSOT; new exported `IsValidConvention` to be added in M1).
- NARROWED OUT: `llm.mode` (backend-switch toggle, only `""|glm`), `llm.default_model` (legacy enum-less string, no `validate` tag, no canonical enum to reuse). Recorded in spec.md §1.
- KEY design: project-config persistence via config-manager `LoadRaw`→`SetSection`→`Save` (new `app` seams), NOT `ProfilePreferences` (which has no slot for these). Bounded `SyncToProjectConfig`-pattern extension.

## REQ / AC counts

- REQs: 8 (REQ-WC3-001 .. REQ-WC3-008)
- ACs: 11 + closure (AC-WC3-001a/001b/002a/002b/003/004/005/006a/006b/007/008/009) + 6 edge cases (EC-1..EC-6)

## Next

- Phase 0.5 plan-auditor (Tier M PASS threshold 0.80) → GATE-2 → /moai run (cycle_type=tdd).

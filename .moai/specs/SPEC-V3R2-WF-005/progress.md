# SPEC-V3R2-WF-005 Progress

- plan_complete_at: 2026-05-03T18:43:25Z
- plan_status: audit-ready

## Artifacts

- spec.md — v0.2.0 (frontmatter 정합화 완료, 본문 보존; 15 REQs / 12 ACs unchanged)
- research.md — Phase 0.5 deep research (38 file:line citations; codebase-grounded scan of 17 files / ~50 `moai-lang-*` mentions + 13 files / 15 dead-skill-ID mentions)
- plan.md — Phase 1B implementation plan (5 milestones M1-M5; 16 file:line anchors; mx_plan with 6 tags; REQ↔AC traceability matrix 15/15 + 12/12)
- acceptance.md — Given/When/Then for 12 ACs (happy path + edge cases per AC; Go test scaffold names declared for `TestNoLangSkillDirectory`, `TestRelatedSkillsNoLangReference`, `TestLanguageNeutrality`)
- tasks.md — M1-M5 milestone breakdown (25 tasks T-WF005-01..25; owner roles aligned with TDD methodology)
- spec-compact.md — Auto-extracted REQ + AC + Files-to-modify + Exclusions reference

## Branch

- Branch: feature/SPEC-V3R2-WF-005-language-rules-boundary
- Mode: solo, no worktree (per user directive)
- Working directory: /Users/goos/MoAI/moai-adk-go
- Base: origin/main HEAD aa55780ce (Wave 6 plan PRs all merged: WF-004 #765, WF-003 #767, session-handoff #763)

## Key Plan Decisions

- Publication target: append new section to `.claude/rules/moai/development/skill-authoring.md` (NOT a separate principle file) — rationale in research.md §5 and plan.md §2 M2.
- BC scope: declaration-level (forward-looking forbid), not behavioral. Verified per-surface no-violation analysis in research.md §8. spec.md §10 BC 영향: 없음 (`breaking: false`, `bc_id: []`).
- TDD methodology: per `.moai/config/sections/quality.yaml development_mode`. M1 RED gate → M2/M3/M4/M5 GREEN gates → M5e REFACTOR + Trackable.
- Sentinel strings: `LANG_AS_SKILL_FORBIDDEN` (REQ-WF005-007), `DEAD_LANG_SKILL_REFERENCE` (REQ-WF005-013), `LANG_NEUTRALITY_VIOLATION` (REQ-WF005-014) — all NEW additions to the codebase.
- Cleanup surface: ~28 distinct files / ~65 mentions (research.md §3 + §4 aggregate). Decomposed per CLAUDE.md §1 Multi-File Decomposition HARD rule.

## Frontmatter Migration Verification

- Required fields present (9/9): `id`, `version`, `status`, `created_at`, `updated_at`, `author`, `priority`, `labels`, `issue_number` ✅
- Rejected aliases absent (0): `created`, `updated`, `spec_id`, `title:` H1-alias ✅
- `version` quoted string: `"0.2.0"` ✅
- `priority` enum: `P2` (bare uppercase, no descriptor) ✅
- `labels` YAML array: `[languages, rules, skills, boundary, paths-frontmatter, v3]` ✅
- `created_at` / `updated_at` ISO date: `2026-04-23` / `2026-05-04` ✅
- `issue_number: null` ✅

## Codebase Scan Summary (research.md grounded)

### `moai-lang-*` references found (17 files, ~50 mentions)

Top files by mention count:
1. `.claude/skills/moai/workflows/run.md` — 16 mentions (Phase 0.9 language detection mapping)
2. `.claude/skills/moai-foundation-core/modules/agents-reference.md` — 7 mentions (catalog table)
3. `.claude/skills/moai-framework-electron/SKILL.md` — 3 mentions (frontmatter + body)
4. `.claude/skills/moai-platform-chrome-extension/SKILL.md` — 3 mentions (frontmatter + body)
5. `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md` — 3 mentions
6. `.claude/skills/moai-platform-deployment/SKILL.md` — 2 mentions
7. `.claude/skills/moai-platform-auth/SKILL.md` — 2 mentions (frontmatter + body)
8. `.claude/skills/moai-workflow-loop/SKILL.md` — 2 mentions
9. `.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md` — 2 mentions
10. `.claude/skills/moai-foundation-cc/reference/skill-examples.md` — 2 mentions

Plus 7 additional files with 1 mention each, including 3 language rules with cross-language references (`scala.md`, `flutter.md` 2 lines, `cpp.md`).

### Other dead-skill-ID references (13 files, 15 mentions)

- `moai-essentials-debug`: 9 files (8 language rules + 1 sub-agent example)
- `moai-quality-testing`: 1 file (`kotlin.md`)
- `moai-quality-security`: 3 files (`moai-domain-backend`, `moai-workflow-project`, `flutter.md`)
- `moai-infra-docker`: 2 files (`kotlin.md`, `java.md`)

### Baseline confirmations

- All 16 language rules have `paths:` frontmatter (REQ-WF005-004 ✅ at baseline)
- No `moai-lang-*` directories exist in `.claude/skills/` (REQ-WF005-002 ✅ at baseline; the audit test will pass at GREEN immediately)
- `flutter.md` filename is canonical (REQ-WF005-010 ✅ at baseline)

## Next Phase

- Phase 0.5 Plan Audit Gate (plan-auditor) at `/moai run SPEC-V3R2-WF-005` entry — see `.claude/rules/moai/workflow/spec-workflow.md:172-204`.
- Implementation Methodology: TDD (per `.moai/config/sections/quality.yaml`).
- Run-phase command: `/moai run SPEC-V3R2-WF-005` (executed from `/Users/goos/MoAI/moai-adk-go` on branch `feature/SPEC-V3R2-WF-005-language-rules-boundary`).
- Post-implementation: `/moai sync SPEC-V3R2-WF-005` for documentation sync + PR creation.

## Plan-Audit-Ready Checklist Summary

All 15 criteria PASS per plan.md §8:

- C1: Frontmatter v0.2.0 (9 required fields)
- C2: HISTORY v0.2.0 entry
- C3: 15 EARS REQs across 5 categories (Ubiquitous 5, Event-Driven 3, State-Driven 2, Optional 2, Complex 3)
- C4: 12 ACs with 100% REQ mapping (15/15 REQ → AC traceability matrix in plan.md §1.4)
- C5: BC scope clarity (declaration-level forbid, `breaking: false`)
- C6: File:line anchors ≥10 (research.md: 38, plan.md: 16)
- C7: Exclusions section present (spec.md §1.2 Non-Goals + §2.2 Out of Scope; spec-compact.md §Exclusions)
- C8: TDD methodology declared
- C9: mx_plan section (6 tags / 5 files; 2 ANCHOR + 2 NOTE + 2 WARN)
- C10: Risk table with file-anchored mitigations (9 risks)
- C11: Solo mode path discipline (3 HARD rules, no worktree per user directive)
- C12: No implementation code in plan documents
- C13: Acceptance.md G/W/T format with edge cases (12 ACs covered)
- C14: tasks.md owner roles aligned with TDD (4 expert-backend / 4 manager-tdd / 17 manager-docs)
- C15: Cross-SPEC consistency (WF-001 completed; independent of WF-003/WF-004)

## Open Items for plan-auditor Review

- Confirm whether `moai-foundation-quality`, `moai-ref-testing-pyramid`, `moai-ref-owasp-checklist` exist as substitute targets for REQ-WF005-015 (M5b/M5c). If any are absent, the substitution may need to fall back to rule-path references or removal.
- Confirm that the new audit test file `internal/template/lang_boundary_audit_test.go` does not collide with any existing test in the package (verified absent at baseline; still worth a final check).
- Validate the regex set in plan.md §6.2 / research.md §6.2 for `TestLanguageNeutrality` — confirm it captures realistic primacy phrasings without false positives on legitimate per-language sections (e.g., `## Python Tooling` headings).

---

End of progress.md.

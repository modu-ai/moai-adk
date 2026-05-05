# SPEC-V3R2-WF-003 Progress

- plan_complete_at: 2026-05-04T12:00:00Z
- plan_status: audit-ready (iteration 2)

## Artifacts

- spec.md — v0.3.0 (frontmatter 정합화 완료, 본문 보존; 18 REQs / **17 ACs** — AC-16/17 추가 per iter 2 D1/D2)
- research.md — v0.1.0 unchanged (Phase 0.5 deep research; 38 file:line citations; with WF-004 sentinel cross-check verdict PASS)
- plan.md — v0.2.0 (Phase 1B implementation plan; M1-M5 + 25 file:line anchors + stacked-PR pre-merge hook + mx_plan with 8 tags; M4d scope expanded 2→4 mode-NA subcommands per iter 2 D4)
- acceptance.md — v0.2.0 (Given/When/Then for **17 ACs** + Traceability Matrix (REQ→AC) + reverse matrix (AC→REQ) per iter 2 D3)
- tasks.md — v0.2.0 (M1-M5 milestone breakdown; 19 tasks T-WF003-01..19 + 1 pre-merge hook T-WF003-PRE; T-WF003-18 scope expanded 2→4 skills per iter 2 D4)

## Worktree (Stacked)

- Branch: feature/SPEC-V3R2-WF-003
- Worktree path: /Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-003
- Base branch: feature/SPEC-V3R2-WF-004 (PR #765 OPEN, base HEAD 5ab409292)
- Pre-merge hook: PR base will transition from feature/SPEC-V3R2-WF-004 → main BEFORE PR #765 merges (CLAUDE.local.md §18.11 v2.14.0 case study)

## Key Plan Decisions

- Publication target for matrix extension: extend existing `## Subcommand Classification` section in `.claude/rules/moai/workflow/spec-workflow.md` (added by WF-004 M4) — NOT a parallel/new section. Single-source-of-truth coordination per research.md §5.
- BC-WF003 scope: additive flag — `breaking: false`, `bc_id: []`. Users not supplying `--mode` get exactly today's behavior. Verified per research.md §11 item 7.
- TDD methodology: per `.moai/config/sections/quality.yaml development_mode: tdd`. M1 RED gate → M2/M3/M4 GREEN gates → M5 REFACTOR + Trackable.
- Sentinel ownership and cross-spec verification:
  - WF-003 OWNS: `MODE_UNKNOWN` (REQ-WF003-010), `MODE_TEAM_UNAVAILABLE` (REQ-WF003-011)
  - SHARED: `MODE_PIPELINE_ONLY_UTILITY` (REQ-WF003-016 ↔ REQ-WF004-014) — verified byte-identical in research.md §4.1
  - WF-004 OWNS: `MODE_FLAG_IGNORED_FOR_UTILITY` (REQ-WF004-011) — preserved unchanged by WF-003
- Audit test architecture: extend WF-004's `internal/template/agentless_audit_test.go` (created in WF-004 M1) with 3 new test functions rather than create a parallel audit file. Single audit file = single coordination point.
- Path B1 (Figma) / B2 (Pencil) handling: NOT in `--mode` axis. Users wanting these paths supply NO `--mode` flag and receive the existing AskUserQuestion path selection. Per research.md §2.2.3 recommendation.
- WF-001 24-skill catalog stability: VERIFIED — all 4 mode targets (`moai-workflow-loop`, `moai-workflow-design-import`, `moai-domain-copywriting`, `moai-domain-brand-design`) are present and active per `.moai/specs/SPEC-V3R2-WF-001/spec.md:5,254,258,265,266`. NO new skills required.
- `default_mode` schema: NEW field added to `workflow.yaml` (skill-only consumption; Go-side loader deferred to SPEC-V3R2-MIG-003 per spec.md §9.2 blocks list).

## Next Phase

- Phase 0.5 Plan Audit Gate (plan-auditor) at `/moai run SPEC-V3R2-WF-003` entry — see `.claude/rules/moai/workflow/spec-workflow.md:172-204`.
- Implementation Methodology: TDD (per `.moai/config/sections/quality.yaml` `development_mode: tdd`).
- Run-phase command: `/moai run SPEC-V3R2-WF-003` (executed inside worktree at `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-003`).
- Pre-merge hook (orthogonal): switch PR base from `feature/SPEC-V3R2-WF-004` → `main` once PR #765 lands. Failure to do so = WF-003 PR auto-closes (CLAUDE.local.md §18.11).
- Post-implementation: `/moai sync SPEC-V3R2-WF-003` for documentation sync + PR creation.

## Plan-Audit-Ready Checklist Summary

All 16 criteria PASS per plan.md §8:

- C1: Frontmatter v0.2.0 (9 required fields)
- C2: HISTORY v0.2.0 entry
- C3: 18 EARS REQs across 5 categories (Ubiquitous 7, Event-Driven 4, State-Driven 2, Optional 2, Complex 3)
- C4: 17 ACs with 100% REQ mapping (AC-16/17 added per iter 2 D1/D2)
- C5: BC scope clarity (additive only — `breaking: false`, `bc_id: []`)
- C6: File:line anchors ≥10 (research.md: 38, plan.md: 25)
- C7: Exclusions section present (spec.md §1.2 + §2.2 = 14 exclusions total)
- C8: TDD methodology declared
- C9: mx_plan section (8 tags / 6 files)
- C10: Risk table with file-anchored mitigations (11 risks)
- C11: Worktree absolute path discipline (4 HARD rules including stacked-PR base transition)
- C12: No implementation code in plan documents
- C13: Acceptance.md G/W/T format with edge cases (15 ACs, 1+ edge cases per AC)
- C14: tasks.md owner roles aligned with TDD methodology (19 tasks + 1 hook)
- C15: Cross-SPEC consistency (WF-001 completed; WF-003 ↔ WF-004 sentinel match verified)
- C16: Stacked PR pre-merge hook documented (CLAUDE.local.md §18.11 cross-ref)

## Iteration History

### Iteration 1 → Iteration 2 (2026-05-04)

Iteration 1 audit verdict: **FAIL** (overall score 0.78). Audit report path: `.moai/reports/plan-audit/SPEC-V3R2-WF-003-review-1.md`. Three Major defects + one Minor defect blocked PASS:

- **D1 (Major)**: REQ-WF003-007 (subcommand × mode matrix publication) had no AC mapping — orphan REQ.
- **D2 (Major)**: REQ-WF003-015 (Optional, future mode-extension schema) had no AC mapping — orphan REQ.
- **D3 (Major)**: acceptance.md missing explicit per-REQ → AC traceability matrix (audit criterion C3 violation).
- **D4 (Minor)**: REQ-WF003-005 enumerates 4 subcommands (plan, sync, project, db) but plan/tasks only covered 2 (plan, sync).

Iteration 2 fixes applied:

- spec.md v0.2.0 → v0.3.0: Added AC-WF003-16 (REQ-007 matrix publication contract) and AC-WF003-17 (REQ-015 future-extension schema contract). Updated §10 traceability claim from "15 ACs" → "17 ACs, all REQ mapped (REQ-WF003-001..018 100%)". HISTORY row v0.3.0 added.
- acceptance.md v0.1.0 → v0.2.0: Added G/W/T scenarios for AC-16 (matrix publication) and AC-17 (additive future extension) per spec.md additions. Added `## Traceability Matrix (REQ → AC)` enumerating all 18 REQs to 17 ACs. Added reverse `## Traceability Matrix (AC → REQ)` for completeness. AC-WF003-09 footnote documents project/db skill-absence vs REQ-005 enumeration → resolved via matrix-publication contract (AC-16) instead of per-skill audit.
- plan.md v0.1.0 → v0.2.0: M4d scope expanded from 2 (plan/sync) → 4 (plan, sync, project, db) mode-NA implementation subcommands. §1.3 Deliverables table + §3.1 Anchor table updated with project.md / db.md additions. M1-M3 unchanged.
- tasks.md v0.1.0 → v0.2.0: T-WF003-18 scope expanded from 4 files (2 source + 2 template) → 8 files (4 source + 4 template) per M4d expansion. Aggregate Statistics + Cross-Reference Map (T-WF003-15 AC coverage adds AC-16; coverage summary updated 15→17 ACs).
- D5 (Advisory) deferred: REQ-WF003-012 `[mode-auto-downgrade]` info-log format not strictly specified in spec.md. Per iteration 2 plan: defer to v0.3.1 amendment or sync-phase docs PR (audit re-flag check).

Files modified in iteration 2: spec.md, acceptance.md, plan.md, tasks.md, progress.md (this file). research.md unchanged (research findings sound).

Iteration 2 audit-readiness verdict: **READY FOR RE-AUDIT**. All 4 blocking defects (D1-D4) addressed; D5 explicitly deferred per audit report's own iteration 2 path-to-PASS recommendation.

## Open Items for plan-auditor Review

- Should the `default_mode` schema extension in `workflow.yaml` include corresponding YAML schema validation in `internal/template/schema/`? Currently no Go-side loader is planned (deferred to SPEC-V3R2-MIG-003), but schema validation is a separate concern.
- Is the recommendation in research.md §2.2.3 (subsume Path B1/B2 under default invocation rather than creating new mode values) consistent with the master design intent at `docs/design/major-v3-master.md:L993`? Auditor may want to verify.
- Should the audit test for `/moai loop` alias documentation (`TestLoopAliasCrossReference`) also verify the literal phrase `--mode loop` appears in `run.md` (not just `loop.md`)? Current plan asserts the cross-reference only in `loop.md`.
- The info-log message format for `[mode-auto-downgrade]` (REQ-WF003-012) is recommended in research.md §6.3 but not strictly specified in spec.md. Should the auditor request a strict format or accept the recommendation as discretionary?

---

End of progress.md.

# SPEC-V3R5-STATUSLINE-V2145-001 Progress Tracker

## Status Summary

| Phase | Status | Owner | Notes |
|-------|--------|-------|-------|
| Plan | in-progress | manager-spec | Initial draft created 2026-05-20; awaiting plan-auditor verdict |
| Run (M1 Hotfix) | pending | TBD | Awaiting plan-auditor PASS + orchestrator dispatch |
| Run (M2 Feature) | pending | TBD | Depends on M1 mechanical completion (same PR) |
| Run (M3 Docs) | pending | TBD | Depends on M2 implementation (segment must exist before docs) |
| Sync | pending | manager-docs | Triggered after run-PR merge |

## Milestone Progress

### M1 — Disappearing Hotfix

- [ ] AC-SLV-001 — DEBUG_STATUSLINE default 0 in template + rendered
- [ ] AC-SLV-002 — Debug fork guarded by explicit opt-in
- [ ] AC-SLV-003 — Dead `echo ""` lines removed
- [ ] AC-SLV-004 — `statusLine.padding` documented in all 4 docs-site locales (cross-references M3)
- [ ] AC-SLV-005 — No user-specific absolute path in rendered wrapper

### M2 — PR Segment Addition

- [ ] AC-SLV-010 — `StdinData.PR` field exists with correct JSON tags
- [ ] AC-SLV-011 — `WorkspaceInfo.Repo` field exists
- [ ] AC-SLV-012 — PR segment default off
- [ ] AC-SLV-013 — PR segment render format `#<number> ⌥<state>`
- [ ] AC-SLV-014 — Review-state color coding for all 5 cases
- [ ] AC-SLV-015 — No segment emitted when PR absent
- [ ] AC-SLV-016 — `SegmentPR` constant + builder branch
- [ ] AC-SLV-017 — Coverage ≥85% on changed files

### M3 — docs-site 4-locale sync

- [ ] AC-SLV-020 — Korean canonical section exists
- [ ] AC-SLV-021 — 4-locale parity (ko/en/ja/zh)
- [ ] AC-SLV-022 — docs CI passes + URL blacklist clean

### Cross-Milestone

- [ ] AC-SLV-100 — Zero-regression on existing 14 segments
- [ ] AC-SLV-101 — golangci-lint baseline does not regress
- [ ] AC-SLV-102 — `make build` regenerates embedded templates successfully
- [ ] AC-SLV-103 — Conventional Commits format

## Commit Log

| Date | SHA | Branch | Author | Summary |
|------|-----|--------|--------|---------|
| 2026-05-20 | (pending) | plan/SPEC-V3R5-STATUSLINE-V2145-001 | manager-spec | feat(spec): SPEC-V3R5-STATUSLINE-V2145-001 plan-phase (spec.md + plan.md + acceptance.md + progress.md) |

## PR Links

| Phase | PR # | Status | Merged commit |
|-------|------|--------|---------------|
| Plan | TBD | not opened | — |
| Run | TBD | not opened | — |
| Sync | TBD | not opened | — |

## Open Issues / Blockers

None at plan-phase initialization. Open questions (OQ-1..OQ-4) resolved inline in plan.md §3.

## Re-planning Triggers

Per `.claude/rules/moai/workflow/spec-workflow.md` § Re-planning Gate, manager-develop MUST append iteration metrics here at the end of each run-phase iteration:

| Iteration | AC PASS count | NEW errors introduced | Stagnation flag |
|-----------|---------------|----------------------|-----------------|
| (none yet) | — | — | — |

Stagnation rule: if AC PASS count does not increase for 3 consecutive entries OR if NEW errors exceed errors fixed in a single iteration, manager-develop returns a structured re-planning report and the orchestrator initiates the AskUserQuestion gap-analysis flow.

## Lessons Capture

To be populated during run-phase per `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol. Anticipated lesson categories per plan.md §4 Risks:

- (pending) Cross-locale docs-site path discovery pattern (R5)
- (pending) v2.1.145 stdin schema adoption pattern for future Claude Code field additions (R4)
- (pending) Render-order test pattern for line-2 segments (R3)

## Handoff Notes

**For orchestrator**: This SPEC is plan-phase complete. Recommended next step:

1. Open plan-PR with branch `plan/SPEC-V3R5-STATUSLINE-V2145-001` containing the 4 artifact files.
2. Invoke plan-auditor for independent verification (orchestrator's call — not invoked by manager-spec per orchestrator-subagent boundary).
3. On plan-auditor PASS, merge plan-PR via admin squash.
4. Then `/moai run SPEC-V3R5-STATUSLINE-V2145-001` to dispatch manager-develop (cycle_type derived from `.moai/config/sections/quality.yaml` `development_mode`).

**For manager-develop (when dispatched)**: This is a 3-milestone bundle (M1 + M2 + M3) ideally shipped in a single run-PR. The plan.md §7 Phase Sequencing dictates the order: M1 Phase A→B, then M2 Phase A→B→C→D→E (RED-GREEN-REFACTOR per moai-workflow-tdd if development_mode=tdd), then M3 4-locale docs. Per the manager-develop-prompt-template.md, the orchestrator's spawn prompt MUST include Sections A-E with Known Issues B1 (cross-platform build tags — N/A here), B2 (cross-SPEC conflict scan — none expected), B3 (subagent boundary — N/A in statusline package), B4 (frontmatter canonical schema — applied), B5 (CI 3-tier), B6 (spec-lint h3 sub-section), B7 (observer.go N/A), B8 (working tree hygiene — preserve `.moai/harness/usage-log.jsonl`).

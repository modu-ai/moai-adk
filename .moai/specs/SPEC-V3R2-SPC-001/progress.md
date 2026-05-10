---
spec_id: SPEC-V3R2-SPC-001
plan_status: audit-ready
plan_complete_at: 2026-05-09
phase: plan
branch: feature/SPEC-V3R2-SPC-001-ears-hierarchical
base_branch: main
base_commit: 464366583
worktree_path: /Users/goos/.moai/worktrees/MoAI-ADK/spc-001-plan
---

# Progress — SPEC-V3R2-SPC-001

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-09 | manager-spec | Plan phase complete; status `audit-ready`. |

## Plan Phase Snapshot

- Plan author: manager-spec (Wave 9 root SPEC).
- Base commit: `464366583` (PR #808 — SPEC-V3R2-WF-005 language rules vs skills boundary codification).
- Branch: `feature/SPEC-V3R2-SPC-001-ears-hierarchical` cut from `origin/main`.
- Worktree: `/Users/goos/.moai/worktrees/MoAI-ADK/spc-001-plan`.

## Artifacts produced this PR

- `.moai/specs/SPEC-V3R2-SPC-001/spec.md` — pre-existing 232-line SPEC document; **not modified** in this PR.
- `.moai/specs/SPEC-V3R2-SPC-001/research.md` — new; 48 file:line anchors; gap analysis vs already-landed Go code.
- `.moai/specs/SPEC-V3R2-SPC-001/plan.md` — new; M1–M5 milestones with priority labels (no time estimates per agent-common-protocol).
- `.moai/specs/SPEC-V3R2-SPC-001/acceptance.md` — new; 17 ACs covering all 18 REQs (full traceability); self-demonstrates hierarchical schema on AC-SPC-001-01, -02, -09, -14.
- `.moai/specs/SPEC-V3R2-SPC-001/tasks.md` — new; 12 tasks (T-SPC001-01 through T-SPC001-12) with owner roles + dependency graph.
- `.moai/specs/SPEC-V3R2-SPC-001/spec-compact.md` — new; one-page REQ + AC + Files reference.
- `.moai/specs/SPEC-V3R2-SPC-001/progress.md` — this file.

## Key Findings (research.md highlights)

1. ~80% of SPC-001's runtime behaviour is already implemented in `internal/spec/`:
   - `Acceptance` struct, `MaxDepth = 3`, `GenerateChildID`, `Depth`, `InheritGiven`, `IsLeaf`, `CountLeaves`, `ValidateDepth`, `ExtractRequirementMappings`, `ValidateRequirementMappings` all in `ears.go` (165 LOC).
   - Parser auto-wrap (`parser.go:200-227`), child-attach (`parser.go:182-185`), `acceptance_format: flat` opt-out (`parser.go:117`) all live.
   - Errors `DuplicateAcceptanceID`, `MaxDepthExceeded`, `DanglingRequirementReference`, `MissingRequirementMapping` in `errors.go`.
   - REQ↔AC tree-aware coverage (`lint.go:394-403`).
   - `moai spec view` CLI command in `internal/cli/spec_view.go` with `--shape-trace` flag.

2. Plan therefore focuses on:
   - M2: 365-leaf perf benchmark + frontmatter edge cases.
   - M3: Documentation (spec-workflow.md amendment, SKILL.md update, zone-registry cross-link, --shape-trace audit).
   - M4: MIG-001 handoff (one cross-reference note).
   - M5: CON-002 amendment paperwork (Canary, HumanOversight evidence, MX tags).

3. 185 SPECs in production are 100% flat; auto-wrap covers read-path compatibility with zero source edits.

## Plan-Phase Self-Check

- [x] research.md ≥25 file:line anchors → 48 achieved.
- [x] All 18 REQs in spec.md §5 mapped to ≥1 AC in acceptance.md.
- [x] At least 3 ACs self-demonstrate hierarchical schema (AC-SPC-001-01, -02, -09, -14).
- [x] tasks.md uses T-SPC001-NN naming with owner role.
- [x] No spec.md modifications.
- [x] No time-based estimates (priority labels only, per agent-common-protocol §Time Estimation).
- [x] EARS modality FROZEN status confirmed (not touched).
- [x] CON-002 amendment requirement acknowledged (M5 task).

## Next Steps

1. Open PR `plan(spec): SPEC-V3R2-SPC-001 — EARS + hierarchical acceptance criteria` against `main`.
2. plan-auditor independent verification.
3. After plan-auditor PASS + admin merge, switch SPC-001 to run-phase: execute M2 → M3 → M5 → M4 per tasks.md dependency graph.
4. Update this progress.md to `plan_status: complete` after Canary evidence (T-SPC001-10) and HumanOversight approval (T-SPC001-12) land.

## Blocked-by

- None (plan phase). CON-001 is on main; zone-registry CONST-V3R2-001 entry exists.

## Blocks

- SPEC-V3R2-SPC-003 (SPEC linter) — depends on hierarchical schema being formally documented in spec-workflow.md (T-SPC001-06).
- SPEC-V3R2-HRN-002 (Sprint Contract) — references leaf-level AC IDs.
- SPEC-V3R2-HRN-003 (per-leaf evaluator scoring) — iterates `Acceptance.Children`.
- SPEC-V3R2-MIG-001 (cosmetic AC rewrite) — handoff note from T-SPC001-09.

End of progress.

## Run Phase Entry (2026-05-11)

### Wave A Complete (2026-05-11)
- wave: A
- milestone: M2 (TDD)
- agent: expert-backend
- tasks: T-03, T-04
- status: COMPLETE
- pr: #849
- commit: 6ac07bf81
- results:
  - BenchmarkParse365Leaves: 6.0ms (<500ms ✅)
  - AC-004/005/006/007/019: 5 tests passing
  - No regressions

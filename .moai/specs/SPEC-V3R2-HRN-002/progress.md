---
spec_id: SPEC-V3R2-HRN-002
plan_status: audit-ready
plan_complete_at: 2026-05-13
phase: plan
branch: plan/SPEC-V3R2-HRN-002
base_branch: main
base_commit: 07dabe011
worktree_path: n/a (plan-in-main)
---

# Progress — SPEC-V3R2-HRN-002

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-13 | manager-spec (HRN-002 plan author) | Plan phase complete; status `audit-ready`. |

## Plan Phase Snapshot

- Plan author: manager-spec (HRN-002 plan author, single session — no team mode).
- Base commit: `07dabe011` (PR #870 — SPEC-V3R2-SPC-001 M5 CON-002 amendment evidence + run-phase completion).
- Branch: `plan/SPEC-V3R2-HRN-002` cut from `origin/main` (plan-in-main doctrine per CLAUDE.local.md §18.12 + PR #822).
- Worktree: n/a — plan-in-main per spec-workflow.md Step 1a default.

## Artifacts produced this PR

- `.moai/specs/SPEC-V3R2-HRN-002/spec.md` — pre-existing 25KB SPEC document; **not modified** in this PR (authored 2026-04-23 by GOOS, Wave 4 round 2).
- `.moai/specs/SPEC-V3R2-HRN-002/research.md` — new; 50 file:line anchors; gap analysis vs as-is evaluator-memory surface; §11 D1–D10 decision block.
- `.moai/specs/SPEC-V3R2-HRN-002/plan.md` — new; M1–M5 milestones with priority labels (no time estimates per agent-common-protocol).
- `.moai/specs/SPEC-V3R2-HRN-002/acceptance.md` — new; 13 ACs covering all 19 REQs (full traceability); self-demonstrates hierarchical schema on AC-HRN-002-01, -04, -07 (the last with depth-2 grandchildren `.a.i`/`.a.ii`).
- `.moai/specs/SPEC-V3R2-HRN-002/tasks.md` — new; 12 tasks (T-HRN002-01 through T-HRN002-12) with owner roles + dependency graph.
- `.moai/specs/SPEC-V3R2-HRN-002/spec-compact.md` — new; one-page REQ + AC + Files reference.
- `.moai/specs/SPEC-V3R2-HRN-002/progress.md` — this file.

## Key Findings (research.md highlights)

1. **~95% greenfield work**. The amendment text, HarnessConfig extension, agent body cross-reference, SKILL.md iteration-handoff step, leak detection test, config keys, loader validator, CON-002 paperwork, and zone-registry CONST entry are all net-new. Only Sprint Contract durability is already in place (existing §11.4 text + SKILL.md JSON shape already encode passed/failed/refined/new status).

2. **Plan therefore focuses on**:
   - M2: Constitution amendment text + HarnessConfig.Evaluator.MemoryScope struct field (T-HRN002-03 + T-HRN002-04).
   - M3: Agent body cross-ref + SKILL.md Phase 4b iteration-handoff + prior-judgment leak detection test (T-HRN002-05 + T-HRN002-06 + T-HRN002-07).
   - M4: design.yaml + harness.yaml config keys + loader validator + 3 test fixtures (T-HRN002-08 + T-HRN002-09).
   - M5: CON-002 5-layer paperwork (mirror SPC-001 pattern) + dual evolution-log + zone-registry CONST entry + HumanOversight approval (T-HRN002-10 through T-HRN002-12).

3. **Decision D1**: Runner enforcement via SKILL.md text + Go integration test (leak detection in `internal/harness/evaluator_leak_test.go`); no new `internal/harness/gan_loop.go` module created. The orchestrator-level runner in moai-workflow-gan-loop SKILL.md is the actual integration point.

4. **Decision D3**: CON-002 paperwork mirrors SPC-001 pattern verbatim (PR #870 just exercised the 5-layer evidence cycle; HRN-002 reuses the format).

5. **Decision D5 + D7**: Dual evolution-log write to both core (`.moai/research/evolution-log.md`) AND design (`.moai/design/v3-research/evolution-log.md`) per the path disambiguation in research §R12. FROZEN value enforcement is symmetric: loader rejects + FrozenGuard rejects.

6. **R1 §9 Agent-as-a-Judge cite is load-bearing**. Zhuge et al. 2024 (arXiv:2410.10934) explicitly identifies cumulative evaluator memory as an anti-pattern; HRN-002's existence rests on this evidence. M5 AskUserQuestion options[].description bundles the citation + canary verdict + Principle 4 reference + research.md §8 gap table as Option 1.

## Plan-Phase Self-Check

- [x] research.md ≥25 file:line anchors → 50 achieved.
- [x] All 19 REQs in spec.md §5 mapped to ≥1 AC in acceptance.md.
- [x] At least 3 ACs self-demonstrate hierarchical schema (AC-HRN-002-01, AC-HRN-002-04, AC-HRN-002-07).
- [x] AC-HRN-002-07 demonstrates depth-2 nesting (`.a.i`, `.a.ii`) — exercises MaxDepth=3 boundary.
- [x] tasks.md uses T-HRN002-NN naming with owner role.
- [x] No spec.md modifications (read-only).
- [x] No time-based estimates (priority labels only, per agent-common-protocol §Time Estimation).
- [x] FROZEN-zone amendment compliance acknowledged (M5 task — CON-002 5-layer paperwork mirrors SPC-001 PR #870 pattern).
- [x] CON-002 amendment requirement acknowledged in M5 task list.
- [x] Decision block (research.md §11) records D1–D10 irreversible plan-time decisions.

## Next Steps

1. Open PR `plan(spec): SPEC-V3R2-HRN-002 — Evaluator Memory Scope Amendment plan artifacts` against `main`.
2. plan-auditor independent verification.
3. After plan-auditor PASS + admin merge, switch HRN-002 to run-phase: create fresh worktree via `moai worktree new SPEC-V3R2-HRN-002 --base origin/main`, then execute M2 → M3 → M4 → M5 per tasks.md dependency graph.
4. Update this progress.md to `plan_status: complete` after Canary evidence (T-HRN002-10) + CON-002 evidence (T-HRN002-12) + HumanOversight approval land.

## Blocked-by

- **None at plan phase**. Run-phase blockers:
  - CON-002 amendment protocol (live on main; SPC-001 PR #870 confirmed working pattern).
  - This SPEC is itself the next CON-002 use case after SPC-001 — sequencing matters per CON-002 §5 Layer 4 RateLimiter (3-amendment cap per v3.x cycle; HRN-002 is #2 after SPC-001).

## Blocks

- **SPEC-V3R2-HRN-003** (hierarchical per-leaf scoring) — depends on fresh-judgment semantics established by HRN-002 §11.4.1. HRN-003 cannot proceed until §11.4.1 lands and the leak detection test gates per-leaf scoring.
- **SPEC-V3R2-WF-003** (multi-mode router) — thorough harness mode uses Sprint Contract per HRN-002. HRN-002 ships the durable substrate + fresh-respawn protocol that thorough mode requires.
- Possibly **`.claude/agents/moai/evaluator-active.md`** body update (downstream of run-phase M3 task T-HRN002-05) — but this is HRN-002's own M3 task, not a downstream SPEC.

End of progress.

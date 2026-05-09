---
spec_id: SPEC-V3R2-ORC-001
plan_complete_at: 2026-05-09T14:00:00Z
plan_status: audit-ready
branch: feature/SPEC-V3R2-ORC-001-roster
base_commit: "464366583"
base_branch: origin/main
---

# Progress: SPEC-V3R2-ORC-001 — Plan Phase Complete

Plan-phase artefacts produced for SPEC-V3R2-ORC-001 (Agent roster
consolidation 22 → 17). Awaiting independent plan-auditor review before
proceeding to run phase.

---

## Status

- [x] Phase 0.5 — Codebase research (research.md)
- [x] Phase 1B — Implementation plan (plan.md)
- [x] Phase 1B — Acceptance criteria (acceptance.md)
- [x] Phase 1B — Tasks breakdown (tasks.md)
- [x] Phase 1B — Compact reference (spec-compact.md)
- [x] Plan complete; awaiting plan-auditor independent verification
- [ ] Plan-auditor approval received (pending)
- [ ] Run phase entry (deferred to next session per ORC-001 plan-auditor sign-off)

---

## Branch Information

- **Branch**: `feature/SPEC-V3R2-ORC-001-roster`
- **Base**: `origin/main` HEAD `464366583` (PR #808 — SPEC-V3R2-WF-005 language rules vs skills boundary codification)
- **Worktree**: `/Users/goos/.moai/worktrees/MoAI-ADK/orc-001-plan`
- **Repo**: `modu-ai/moai-adk`
- **Breaking**: `true` (BC-V3R2-005, BC-V3R2-009, BC-V3R2-016 declared in spec.md)

---

## Plan Artefact Inventory

| File | Purpose | Lines (approx) | Status |
|------|---------|----------------|--------|
| `spec.md` | Existing SPEC document (18 REQs, 12 ACs, 22→17 destiny table) | 335 | Pre-existing (NOT modified by this plan run) |
| `research.md` | Phase 0.5 codebase research with 30+ file:line citations | ~270 | NEW |
| `plan.md` | Phase 1B implementation plan (M1-M5 with mx_plan tags) | ~330 | NEW |
| `acceptance.md` | Given/When/Then ACs for all REQs + REQ↔AC traceability matrix | ~330 | NEW |
| `tasks.md` | 24 T-ORC001-NN tasks with owner roles and dependency graph | ~210 | NEW |
| `progress.md` | This file | ~80 | NEW |
| `spec-compact.md` | Compact reference (REQ + AC + Files-to-modify + Exclusions) | ~110 | NEW |

**Total artefacts**: 7 files (1 pre-existing + 6 new). Per spec.md
constraint, `spec.md` was NOT modified by this plan run; spec-level
amendments (post-R5 reconciliation, OQ resolutions) are deferred to the run
phase milestone M5.2.

---

## Key Plan Decisions

1. **Carry-over recognition**: Plan acknowledges that
   SPEC-V3R3-RETIRED-AGENT-001 and SPEC-V3R3-RETIRED-DDD-001 already created
   `manager-cycle.md` and retired `manager-ddd.md` / `manager-tdd.md`. M1
   verifies; no regression introduced.

2. **5 outstanding consolidations**: M2 retires `builder-agent`,
   `builder-skill`, `builder-plugin`, `expert-debug`, `expert-testing` to
   stubs and creates `builder-platform.md`.

3. **6 refactor passes**: M3 modifies manager-quality, manager-project,
   expert-backend, expert-performance, manager-git, plan-auditor.

4. **Template-First mandate**: All edits begin in
   `internal/template/templates/.claude/agents/moai/`; `make build`
   regenerates embedded.go and mirrors the local copy. CLAUDE.local.md §2
   strictly enforced.

5. **Deferred work**: ORC-002 (CI lint), ORC-003 (effort matrix), ORC-004
   (worktree MUST), MIG-001 (legacy SPEC rewriter), WF-003 (Pencil split)
   are explicitly NOT in scope here.

---

## Open Questions Captured

OQ-1 — Roster delta arithmetic (post-R5 additions)
OQ-2 — Stub frontmatter shape consistency
OQ-3 — Migration of legacy SPEC bodies (deferred to MIG-001)
OQ-4 — Concurrent SPEC interaction with ORC-002
OQ-5 — `claude-code-guide.md` inclusion in active count

All 5 OQs have **recommended resolutions** documented in research.md §4.
Auditor may concur or override; if override, M1.2 absorbs the new
resolution.

---

## Counts Summary

- REQs in spec.md: **17** (Ubiquitous 5, Event-Driven 3, State-Driven 3, Optional 3, Unwanted 3)
- ACs in acceptance.md: **17** (AC-01 through AC-17)
- Tasks in tasks.md: **24** (M1: 5, M2: 8, M3: 7, M4: 3, M5: 1)
- Cited file:line anchors in research.md: **30+** unique anchors across 13 files
- Milestones: **5** (M1: P0; M2: P0; M3: P1; M4: P1; M5: P0)

---

## Next Action

Open PR with title `plan: SPEC-V3R2-ORC-001 — Agent roster consolidation
(22 → 17)` requesting **plan-auditor independent verification (10
dimensions)**. plan-auditor should:

1. Validate REQ-AC traceability matrix completeness.
2. Verify EARS compliance on all 17 REQs.
3. Verify file:line citations in research.md are reachable from base
   `464366583`.
4. Inspect deferred-work boundary (ORC-002/003/004 dependencies declared
   correctly).
5. Confirm trigger-union snapshot capture sequence (T-04 must precede
   T-07..T-11).
6. Review OQ-1..5 recommended resolutions for soundness.

If plan-auditor returns PASS, run phase begins in a separate session per the
session-handoff protocol (see CLAUDE.md §17 + paste-ready resume message
generated post-merge).

---

## Blockers

**None at this time.** All inputs are present (existing spec.md, R5 audit,
master design doc, source agent bodies, template tree). No external
dependencies block plan-auditor review.

---

End of progress.md.

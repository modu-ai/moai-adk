---
spec_id: SPEC-V3R2-ORC-001
purpose: "Compact reference auto-extracted from spec.md + plan.md + acceptance.md"
created_at: 2026-05-09
---

# SPEC-V3R2-ORC-001 — Compact Reference

Quick-reference compact for SPEC-V3R2-ORC-001 (Agent roster consolidation 22
→ 17). For deep context, see `spec.md`, `plan.md`, `acceptance.md`,
`research.md`, and `tasks.md` in this directory.

---

## Goal

Consolidate 22 agents into 17 by merging manager-ddd+tdd into
manager-cycle (already done), retiring 3 builders into builder-platform,
folding expert-debug into manager-quality (Diagnostic Sub-Mode), splitting
expert-testing into manager-cycle (strategy) + expert-performance (load),
and scope-shrinking manager-project.

---

## Requirements (17 REQs)

| REQ-ID | Type | One-liner |
|--------|------|-----------|
| REQ-ORC-001-001 | Ubiquitous | Final roster: exactly 17 R5-baseline agents |
| REQ-ORC-001-002 | Ubiquitous | manager-cycle declares cycle_type, preserves DDD+TDD steps |
| REQ-ORC-001-003 | Ubiquitous | builder-platform declares artifact_type, retains 5-phase workflow |
| REQ-ORC-001-004 | Ubiquitous | manager-quality contains Diagnostic Sub-Mode section |
| REQ-ORC-001-005 | Ubiquitous | Template-First: every change in template tree first |
| REQ-ORC-001-006 | Event-Driven | 7 retired stubs at deprecated paths (status: retired) |
| REQ-ORC-001-007 | Event-Driven | MIG-001 rewrites legacy SPEC references |
| REQ-ORC-001-008 | Event-Driven | manager-project rejects out-of-scope writes via blocker |
| REQ-ORC-001-009 | State-Driven | Trigger keyword union preserved across merges |
| REQ-ORC-001-010 | State-Driven | expert-backend trigger row 12-15 EN tokens, no dups |
| REQ-ORC-001-011 | State-Driven | mcp__context7__* dropped from manager-git, manager-quality |
| REQ-ORC-001-012 | Optional | manager-quality optional memory: project |
| REQ-ORC-001-013 | Optional | plan-auditor gains memory: project |
| REQ-ORC-001-014 | Optional | expert-performance optional Write scope |
| REQ-ORC-001-015 | Unwanted | CI rejects advisor-only agents with write-side-effects |
| REQ-ORC-001-016 | Unwanted | CI fails on ORC_AGENT_DELETE_WITHOUT_STUB |
| REQ-ORC-001-017 | Unwanted | CI fails on trigger keyword drops |

---

## Acceptance Criteria (17 ACs)

| AC-ID | Verifies | Verification handle |
|-------|----------|---------------------|
| AC-ORC-001-01 | Final roster size = 17 | `grep -L "retired: true" template/agents/moai/*.md \| grep -v -E "expert-mobile\|manager-brain\|claude-code-guide" \| wc -l` |
| AC-ORC-001-02 | manager-cycle DDD+TDD preservation | `grep "cycle_type\|ANALYZE-PRESERVE-IMPROVE\|RED-GREEN-REFACTOR"` |
| AC-ORC-001-03 | builder-platform 7 artifact_types | `grep "artifact_type" + 5-phase scan` |
| AC-ORC-001-04 | manager-quality Diagnostic Sub-Mode | `grep "Diagnostic Sub-Mode"` |
| AC-ORC-001-05 | 7 retired stubs exist with retired: true | `grep -l "retired: true" \| wc -l = 7` |
| AC-ORC-001-06 | Template ↔ local byte-identical | `diff -r template local` exit 0 |
| AC-ORC-001-07 | MIG-001 rewrites legacy refs | (Deferred to MIG-001 integration) |
| AC-ORC-001-08 | manager-project blocker on out-of-scope | Body grep + manual spawn test |
| AC-ORC-001-09 | expert-backend EN row 12-15, no dups | `grep "EN:" \| tr , \\n \| wc -l` |
| AC-ORC-001-10 | Context7 absent from quality + git | `grep -c "context7" = 0` |
| AC-ORC-001-11 | plan-auditor memory: project | `grep "memory: project"` |
| AC-ORC-001-12 | Trigger union preserved | snapshot diff (see T-04) |
| AC-ORC-001-13 | Stub bodies redirect properly | `wc -l ≤ 50` per stub |
| AC-ORC-001-14 | manager-project allowlist enforced | grep + body inspection |
| AC-ORC-001-15 | builder-platform has Write tool (not advisor-only) | `grep "Write"` in tools |
| AC-ORC-001-16 | No orphan deletions | `git diff --name-status \| awk '$1=="D"'` empty |
| AC-ORC-001-17 | No silent trigger drops | (same as AC-12) |

---

## Files to Modify

### NEW (1 file)

- `internal/template/templates/.claude/agents/moai/builder-platform.md`

### MODIFIED — Body retired to stub (5 files)

- `internal/template/templates/.claude/agents/moai/builder-agent.md`
- `internal/template/templates/.claude/agents/moai/builder-skill.md`
- `internal/template/templates/.claude/agents/moai/builder-plugin.md`
- `internal/template/templates/.claude/agents/moai/expert-debug.md`
- `internal/template/templates/.claude/agents/moai/expert-testing.md`

### MODIFIED — Refactor passes (6 files)

- `internal/template/templates/.claude/agents/moai/manager-quality.md` (Diagnostic Sub-Mode + Context7 drop)
- `internal/template/templates/.claude/agents/moai/manager-project.md` (scope shrink)
- `internal/template/templates/.claude/agents/moai/expert-backend.md` (trigger dedup)
- `internal/template/templates/.claude/agents/moai/manager-git.md` (trigger trim)
- `internal/template/templates/.claude/agents/moai/expert-performance.md` (Write tool)
- `internal/template/templates/.claude/agents/moai/plan-auditor.md` (memory: project)

### MODIFIED — Downstream sync (~85 files)

- ~57 skill files in `.claude/skills/` and `internal/template/templates/.claude/skills/` (excluding `examples.md`)
- ~28 rule files in `.claude/rules/` and `internal/template/templates/.claude/rules/`
- `CLAUDE.md` (root — Agent Catalog L106-126, plus L62/L147/L378)
- `internal/template/templates/CLAUDE.md` (L136, L378)
- `internal/template/embedded.go` (REGENERATED by `make build`)

### NOT MODIFIED (intentionally)

- `spec.md` (preserved per task constraint; HISTORY amendments deferred to run phase M5.2)
- `expert-frontend.md` (Pencil split deferred to WF-003)
- `expert-mobile.md`, `manager-brain.md`, `claude-code-guide.md` (post-R5 KEEP)
- `evaluator-active.md` (per spec.md §10.1 KEEP)
- All Go code under `internal/`, `pkg/`, `cmd/` (no Go logic touched)

---

## Exclusions (What NOT to Build)

Per spec.md §1.2 Non-Goals + §2.2 Out of Scope:

1. ORC-002 CI lint implementation (AskUserQuestion text scrub)
2. ORC-003 effort field matrix population
3. ORC-004 isolation: worktree additions
4. MIG-001 v2→v3 user migrator implementation (this SPEC declares dependency only)
5. WF-003 expert-frontend Pencil split
6. Adding `manager-design` agent (P-A21 deferred)
7. Frontmatter schema changes (handled by V3-AGT-001)
8. Builder-platform Go-side dispatcher (CLI, not agent body)
9. evaluator-active / plan-auditor body changes beyond memory: project
10. Breaking changes outside BC-V3R2-005/006/009/016

---

## Dependencies

- **Blocked by**: SPEC-V3R2-CON-001 (constitution zone codification), SPEC-V3R2-ORC-002 (CI lint)
- **Blocks**: SPEC-V3R2-MIG-001, SPEC-V3R2-WF-003, SPEC-V3R2-HRN-001
- **Related**: SPEC-V3R2-CON-003, SPEC-V3-AGT-001 (legacy)
- **Carry-over from**: SPEC-V3R3-RETIRED-AGENT-001 (PR #776), SPEC-V3R3-RETIRED-DDD-001 (PR #781)

---

## Wave Position

Wave 9 root SPEC for V3R2 ORC series. Provides foundation for ORC-002~005
dependent chain. Parallel to SPC-001 (separate work).

---

End of spec-compact.md.

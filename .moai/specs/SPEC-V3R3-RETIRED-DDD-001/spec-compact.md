# SPEC-V3R3-RETIRED-DDD-001 — Compact Summary

> Token-budget-friendly summary of `spec.md` v0.3.0. Use this for context-loaded skill invocations.

## Identity
- **ID**: SPEC-V3R3-RETIRED-DDD-001
- **Title**: manager-ddd Retired Stub Standardization (Predecessor Pattern Replication)
- **Status**: draft (force-accept after iter 3, score 0.78, must-pass MP-1~4 PASS)
- **Priority**: P1
- **Predecessor**: SPEC-V3R3-RETIRED-AGENT-001 (PR #776, MERGED main `20d77d931`)
- **Issue**: #778
- **Branch**: `feature/SPEC-V3R3-RETIRED-DDD-001` (origin/main base)
- **Breaking**: false / `bc_id: []`

## Goal
manager-tdd retire 패턴 (predecessor SPEC-V3R3-RETIRED-AGENT-001) 을 **manager-ddd** 에 동일 적용. `manager-cycle` 을 unified DDD/TDD agent 로 정립.

## Scope (36 files total)
- **Cat A SUBSTITUTE (30)**: documentation cross-references `manager-ddd` → `manager-cycle` (or `manager-cycle with cycle_type=ddd`).
- **Cat B REWRITE (1)**: `manager-ddd.md` 7674 bytes → ~1.4KB retired stub (predecessor `manager-tdd.md` pattern).
- **Cat C UPDATE-WITH-ANNOTATION (2)**: `agent-hooks.md` table row preserved with retired marker; `agent-authoring.md` Manager Agents listing.
- **Ancillary (3)**: `agent_frontmatter_audit_test.go` row addition; `factory.go` verification; CHANGELOG.

## EARS Requirements (12 total)
- **Ubiquitous (4)**: REQ-RD-001 (retire stub), REQ-RD-002 (substitute), REQ-RD-003 (annotate), REQ-RD-004 (audit row).
- **Event-Driven (3)**: REQ-RD-005 (SubagentStart guard), REQ-RD-006 (make build), REQ-RD-007 (factory dispatch invariant).
- **State-Driven (3)**: REQ-RD-008 (body content), REQ-RD-009 (cross-references), REQ-RD-010 (audit single source of truth).
- **Optional (1)**: REQ-RD-011 (`moai agents list --retired` deferred per predecessor).
- **Unwanted (1, composite)**: REQ-RD-012 (`RETIREMENT_INCOMPLETE_manager-ddd` CI assertion + no silent acceptance + no documentation drift).

## Acceptance Criteria (4 ACs)
- **AC-RD-01 (Positive)**: Retire stub + 30 substitutions + embedded.go regenerated end-to-end.
- **AC-RD-02 (Edge)**: Cat C UPDATE-WITH-ANNOTATION preserves action-key dispatch; body matches predecessor format.
- **AC-RD-03 (Boundary)**: audit fails fast on any frontmatter regression.
- **AC-RD-04 (Negative)**: SubagentStart guard blocks manager-ddd spawn ≤500ms; factory ddd_handler dispatch intact.

REQ ↔ AC: 11/12 explicit (REQ-RD-011 deferred).

## Methodology
TDD: M1 RED (audit row) → M2 GREEN-1 (Cat B REWRITE) → M3 GREEN-2 (Cat A 30 SUBSTITUTE) → M4 GREEN-3 (Cat C 2 ANNOTATION) → M5 REFACTOR (CHANGELOG + final CI).

## Constraints
- Template-First HARD: `internal/template/templates/` mirror + `make build` 필수.
- 16-language neutrality.
- No drive-by refactor: Cat A files은 substitution scope 한정.
- Solo mode, no worktree.
- Predecessor-pattern fidelity: 새로운 패턴 도입 금지.

## Top Risks
- C18: Cat A enumeration baseline drift between iter 3 freeze and implementation. → M3.1 grep re-verification.
- Cat A `manager-ddd-pre-transformation` action key prefix collision → No `replace_all`; prose-only substitute.
- Cat C row 단순 제거로 factory dispatch orphan → 보존 + retired marker.
- predecessor merge stale → PR review에서 manager-cycle.md 변경 여부 확인.

## Plan-Audit History
- iter 1: FAIL 0.78 (8 defects D1-D8)
- iter 2: FAIL 0.74 (REGRESSION D-NEW-1~5)
- iter 3: FAIL 0.78 (recovery, MP-1~4 PASS, traceability dim 0.65)
- Manual textual sync (force-accept) — 4 fixes at spec.md:L170 + plan.md §3 + Risk row + C18 → grep 4-step verification PASS.

## Force-Accept Justification
MP-1 (REQ Number Consistency) / MP-2 (EARS Format) / MP-3 (Frontmatter) / MP-4 (Language Neutrality) — 4종 모두 PASS. score 0.78은 traceability dimension weighted average; cross-artifact reference manual sync 완료.

---

End of spec-compact.md

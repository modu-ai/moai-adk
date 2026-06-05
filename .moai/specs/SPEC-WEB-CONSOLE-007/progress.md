---
spec_id: SPEC-WEB-CONSOLE-007
status: draft
era: V3R6
tier: M
development_mode: tdd
deferred_to: SPEC-WEB-CONSOLE-008   # workflow/git-strategy/harness/llm nested editing (REQ-WC-012 boundary lift + new validators + sentinel retarget)
---

# SPEC-WEB-CONSOLE-007 — Progress

## §F.1 Plan-phase

- **Authored**: 2026-06-06 by manager-spec.
- **Artifacts**: spec.md (12-field frontmatter + GEARS REQ-WC7-001..014 + §F Exclusions), plan.md (Tier M 정당화 + M1..M6 + test-class per milestone), acceptance.md (AC-WC7-001..020 + traceability + MUST-PASS gate), design.md (nested serialization + curated inventory + per-field validation map + 신규 위젯 + write-seam load-modify-write + nested isolation), research.md (전체 file:line ground-truth).
- **Tier**: M (right-size 정당화 plan.md §F — 6 nested 필드는 작으나 2 신규 Templ 위젯 + write-seam 심화 + 검증 export seam + 4종 TDD 거동으로 Tier S 초과; 단일 패키지 + 신규 validator 0개 + 서버 계약 무변경으로 Tier L 미달).
- **Curated 편집 필드 인벤토리** (정확히 이것만, spec.md §E):
  - quality: `test_coverage_target`(int, 기존 0-100), `enforce_quality`(bool), `tdd_settings.min_coverage_per_commit`(int, 기존 0-100).
  - git_convention: `convention`(enum, 기존 유지) + `auto_detection.confidence_threshold`(float [0,1], 기존), `auto_detection.enabled`(bool), `custom.pattern`(string, 기존 custom-required).
- **CRITICAL SCOPE CONSTRAINT 준수**: 두 기존 검증기(validateQualityConfig/validateGitConventionConfig) 확장/export seam만; 신규 validator 함수 0개.
- **HARD invariants**: 8개 (spec.md §B). HARD-2(006 sentinel 무수정) + HARD-4(nested isolation 증명) 핵심.
- **Deferred → 008** (spec.md §F): workflow/git-strategy/harness/llm nested 편집(boundary lift + 신규 validator + sentinel retarget), partial-swap fragment, 동적 섹션 레지스트리.
- **SPEC ID self-check**: decomposition: SPEC ✓ | WEB ✓ | CONSOLE ✓ | 007 ✓ → PASS (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- **Plan-phase commit**: 926816abe

## §F.2 Plan Audit Gate
- (pending — plan-auditor 독립 감사 대기)

## §E.1 Run-phase
- (pending — GATE-2 사용자 승인 후 manager-develop cycle_type=tdd)

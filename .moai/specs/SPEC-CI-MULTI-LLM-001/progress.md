---
spec_id: SPEC-CI-MULTI-LLM-001
created: 2026-04-27
updated: 2026-04-27
plan_complete_at: 2026-04-27
plan_status: audit-ready
phase: plan
harness_level: standard
development_mode: tdd
---

# SPEC-CI-MULTI-LLM-001 — Progress

## Phase 1: Plan (Complete)

### Socratic Interview (2 rounds, 2026-04-27)

Round 1 — Architecture decisions (4 questions):
- Runner deployment model: 사용자 본인 머신 자동 설치
- CLI command structure: `moai github` 서브커맨드 그룹
- LLM trigger model: 코멘트 단수 + PR open 자동 패널
- Auth bootstrap automation: `moai github auth <llm>` 대화형

Round 2 — Scope and methodology (3 questions):
- Distribution scope: 사용자 프로젝트 템플릿 (`internal/template/templates/.github/workflows/`)
- v1.0 LLM coverage: 4개 전부 (Claude / Codex / Gemini / GLM)
- Next action: SPEC 우선 작성 → `/moai run`

Round 3 — Open Questions resolution (4 questions, 2026-04-27):
- OQ1: macOS arm64만 v1.0 (Linux v1.1 분리)
- OQ2: `moai init`에서 자동 호출 안 함
- OQ3: 단일 통합 코멘트 + LLM별 섹션 분리
- OQ4: 로컬 SessionStart 훅에서만 검증

### Artifacts

| File | Size | Status |
|------|------|--------|
| `spec.md` | 800 lines | audit-ready |
| `plan.md` | ~450 lines (manager-strategy 재작성) | audit-ready |
| `tasks.md` | ~200 lines (신규) | audit-ready |
| `acceptance.md` | ~450 lines (신규) | audit-ready |

### Plan Audit Pre-conditions

- [x] 31 REQ 매핑 완료 (REQ-CI-001~021 + REQ-SEC-001~010)
- [x] 30 task 4 wave 분할 완료
- [x] 11 risk 모두 task로 매핑
- [x] HARD 제약 13개 모두 SPEC 본문 명시
- [x] 4 Open Questions 모두 Resolved (spec.md §8)
- [x] EARS 5 패턴 모두 활용
- [x] 16-language 중립성 검증 (REQ-CI-012)
- [x] Template-First 검증 (REQ-CI-019)
- [x] Codex private 가드 (REQ-CI-007, REQ-SEC-001)
- [x] Claude SHA broken range 회피 (REQ-CI-018.1)

### Plan Audit Defect Remediation (2026-04-27)

- [x] acceptance.md 생성 완료 (60+ Given-When-Then criteria)
- [x] plan.md YAML frontmatter 수정 (status, version, author 업데이트)
- [x] EARS count mismatch 인지 및 plan.md §5에 수정본 명시
- [x] REQ-SEC-008 (Audit Log) → T-29에 매핑
- [x] REQ-SEC-009 (Token Expiry) → T-15~T-19 acceptance criteria에 명시
- [x] REQ-SEC-010 (External Contributor Guard) → T-15~T-20 acceptance criteria에 명시

## Phase 2: Run (Pending)

Trigger: `/moai run SPEC-CI-MULTI-LLM-001`

Pre-flight:
- Plan-Auditor (SPEC-WF-AUDIT-GATE-001) 자동 실행 예정
- Methodology: TDD (RED-GREEN-REFACTOR, quality.yaml `development_mode: tdd`)
- Min coverage per commit: 80% (`tdd_settings.min_coverage_per_commit`)

Wave plan (CLAUDE.local.md §16 stream stall 회피):
- Wave 1 (M1): T-01~T-06 (CLI skeleton + Runner domain)
- Wave 2 (M2): T-07~T-11 (LLM auth bootstrap)
- Wave 3 (M3): T-12~T-22 (Workflow templates + composite actions)
- Wave 4 (M4): T-23~T-30 (Integration + docs-site 4-locale)

### Plan Audit (2026-04-27)

- audit_verdict: FAIL_WARNED
- audit_report: .moai/reports/plan-audit/SPEC-CI-MULTI-LLM-001-2026-04-27.md
- audit_at: 2026-04-27T14:30:00Z
- grace_window: ACTIVE (D-5, expires 2026-05-02)
- blocking_defects: 3 (acceptance.md missing, YAML frontmatter violations, EARS count mismatch)
- plan_artifact_hash: computed from spec.md only (plan.md/acceptance.md/tasks.md not on disk)

## Phase 3: Sync (Pending)

Trigger: `/moai sync SPEC-CI-MULTI-LLM-001` (Wave 4 완료 후)

Sub-tasks:
- API documentation generation
- CHANGELOG.md entry
- docs-site 4-locale guides (ko/en/ja/zh)
- Conventional commits + PR creation

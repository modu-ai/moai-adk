# progress — SPEC-CODEMAPS-ACCURACY-001

## §E.1 Plan-phase Audit-Ready Signal

- SPEC-ID: SPEC-CODEMAPS-ACCURACY-001 · status: draft · Tier M · 2026-09-02
- 산출물: spec.md / plan.md / acceptance.md / progress.md (Tier M 4종 — spec.md §5 Tier 분류 근거)
- 사전검사: SPEC-ID 정규식 `PASS` (실행 출력 인용, plan §E); 카탈로그 충돌 0건 (`SPEC-CODEMAPS-ACCURACY-*` 부재 — 인접 ID SPEC-DWF-CODEMAPS-PILOT-001·SPEC-V3R6-DOCS-CODEMAPS-V3-001는 별개)
- 조사 완료: §1.1 부재 8개 분류표 완결 (P1–P8) · §1.2 ListActive 3지점 실API 대조 (registry.go 직독) · §1.3 생성기 입력 재발 경로 특정 (스킬 Phase 2 병합 입력 + Phase 4 비기계 검증) · 설계 결정 D1–D4 확정
- 병합 순서 제약 선언: run-phase는 M0(origin/develop 흡수 + 재측정, REQ-CMA-008)부터 — t432 병합 전 codemaps 편집 금지
- plan-phase는 `.moai/project/codemaps/`·`internal/` 무편집 · 커밋 없음 (레인 소관)

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

## §F Phase 4 Mode Selection

**Input parameters**: tier M · scope ~9 files (codemaps 2·3문서 + internal/graph + internal/cli + 테스트 + 스킬 로컬/템플릿 미러 2 + SPEC 산출물) · domain count 4 (Go 코드, 문서, 스킬 템플릿, 증거) · file language mix markdown+Go · concurrency benefit LOW (coding-heavy, 마일스톤 간 의존 M0→M2) · Agent Teams prereqs 미충족(명시 요청 없음)

**Mode evaluation**:

| Mode | 선택 | 근거 |
|------|------|------|
| direct | not selected | 다중 파일·Go 신규 코드 — trivial 아님 |
| serial | **selected** | coding-heavy (Anthropic coding-task caveat) + 마일스톤 순서 의존 + 문서 단일 작성자 |
| fanout | not selected | research-heavy 아님 — 병렬 이득 낮음 |
| sweep | not selected | ~30파일 기계변환 아님, 새 코드 포함 |

**Decision**: serial

**Justification**: 구현이 Go 신규 코드(M1)와 문서 수리(M2)·스킬 편집(M3)의 순서 의존 체인이라 병렬화 이득이 없고, Anthropic의 coding-task parallelism caveat에 따라 serial이 기본이다. 마일스톤당 manager-develop 1회 spawn, 병합 순서 제약(M0 게이트)이 상태를 공유하므로 단일 작성자가 옳다.

**Autonomy**: Implementation Kickoff Approval 통과(운영자가 본 레인 터미널에서 AskUserQuestion으로 직접 승인, 2026-09-02) · `ac_converge` goal 무장 후 **해제**(2026-09-02, 리드 지시 + t436 결함 — `moai goal` 산문 조건 오분류 계열. 직접 관측: 본 조건은 `conditions[0].type = model`로 정상 분류돼 t436 트리거(후행 `exits <N>` → mechanical)가 발화하지 않았고 turns 0 — 그러나 결함 축이 오늘 활발히 관리 중이므로 리드의 전 레인 무장 해제 방침을 따름) · **반자율 전환**: 마일스톤 경계마다 완료 보고 후 진행, 운영자 결정은 blocker 보고로 lead-1 경유

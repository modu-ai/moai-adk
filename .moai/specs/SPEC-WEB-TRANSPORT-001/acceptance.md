---
id: SPEC-WEB-TRANSPORT-001
title: "acceptance — validation-400 boosted-response transport discriminator"
version: "0.1.0"
created: 2026-09-22
author: manager-spec
---

## §A AC Matrix

| AC ID | Requirement | 검증 방법 | 이원적 판정 |
|-------|-------------|-----------|-------------|
| AC-TR400-001 | REQ-TR400-001 | httptest 특성화 테스트 | 테스트 PASS/FAIL |
| AC-TR400-002 | REQ-TR400-002 | 임베드 자산 기계 추출 검사 | 추출 성공+결과 기록 / deferred 기록 |
| AC-TR400-003 | REQ-TR400-003, REQ-TR400-004 | 판정 기록 판독 | 기록 존재+게이트 충족 |
| AC-TR400-004 | REQ-TR400-005 | git diff + 기록 판독 | 운용 코드 0변경 + 비인용 |
| AC-TR400-005 | 전체 | 영향 패키지 테스트 | exit 0 |

## §B Given-When-Then 시나리오

### AC-TR400-001 — 400 본문이 피드백을 운반한다 (D-400a)

- **Given** 설정 폼의 validator 가 하나 이상 실패하는 POST 페이로드
- **When** httptest 로 `POST /save` 를 서브밋한다
- **Then** 응답 상태는 400 이고, 본문은 "Validation failed — no changes were saved." 배너 문구를 포함하며, 실패한 필드의 오류 메시지 문자열을 최소 1개 포함한다
- **FAIL 시 분류** — 본문 미포함 관측은 REQ-TR400-003 진실표의 `measurement-invalid (서버 계약 발산)` 등급으로 분류된다 (판정이 아니라 재측정·에스컬레이션으로 닫힘 — 무분류 FAIL 없음)

### AC-TR400-002 — 임베드 htmx 의 4xx 스왑 계약이 측정된다 (D-400b)

- **Given** 커밋된 임베드 자산 `internal/web/assets/htmx.min.js` (htmx 2.0.4)
- **When** 기계 추출 검사가 자산 내부의 기본 응답 처리 테이블을 추출해 상태 400 에 대해 평가한다
- **Then** 평가 결과(swap true/false)와 매칭된 테이블 조각 **전문**이 증거로 기록된다. 추출이 불가능하면 결과 대신 `deferred-to-browser` 등급과 추출 시도의 실패 지점이 기록된다 — 어느 쪽이든 **측정 시도의 기록이 존재**하는 것이 합격이고, 무기록 불가 판정(추론 판정)은 불합격이다. 검사는 네트워크·브라우저 없이 Go 테스트 스위트 안에서 돈다.

### AC-TR400-003 — 판정 기록이 측정에 게이트돼 있다

- **Given** AC-TR400-001 과 AC-TR400-002 의 측정이 모두 관측됨
- **When** 판정 기록이 `.moai/reports/t1080/verdict.md` 에 작성된다
- **Then** 기록은 (1) REQ-TR400-003 사상표의 판정 등급(defect-present(contract-level) / defect-absent / measurement-invalid / deferred-to-browser 중 하나 — measurement-invalid 시 판정 대신 재측정·에스컬레이션 기록으로 적용), (2) REQ-TR400-004 의 신뢰도 등급, (3) 두 측정 출력의 전문 인용을 담는다. 측정 기록 없이 쓰인 판정은 불합격이다.

### AC-TR400-004 — 비소급 독립성

- **Given** run 단계가 완료됨
- **When** 변경 diff 와 판정 기록을 검사한다
- **Then** `internal/web` 운용 소스(`*_test.go` 제외)의 변경이 0 이고, SPEC-WEB-CONSOLE-017 산출물이 무변경이며, 판정 기록이 t1051 판정을 근거로 인용하지 않는다

### AC-TR400-005 — 품질 게이트

- **Given** 모든 구현이 완료됨
- **When** `unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED && go test ./internal/web/...` 를 실행한다
- **Then** exit 0 이고 신규 테스트가 포함된다. 모든 임시 디렉터는 `t.TempDir()` 아래에 있다

## §C 에지 케이스

1. **자산 추출 실패** — minified 리터럴 형태가 예상과 다르면 AC-TR400-002 는 deferred 기록으로 합격한다(REQ-TR400-003). 추출 실패를 테스트 FAIL 로 취급하면 판정 이연 경로가 죽는다.
2. **문서와 자산 불일치** — htmx 공식 문서 기본값이 자산 기본값과 다르면 자산값이 우선한다(plan §B.3). 판정 기록은 어느 쪽을 따랐는지 명시한다.
3. **배너 문구 변경** — "Validation failed — no changes were saved." 는 현재 코드의 문자열이다. 특성화 테스트는 이 문자열에 고정되며, 수리 카드가 문구를 바꾸면 테스트도 함께 갱신된다(본 SPEC 은 문구를 바꾸지 않는다).

## §D Definition of Done

- [ ] AC-TR400-001..005 전부 PASS (판정 기록에 전문 증거 인용)
- [ ] §E.2/§E.3 run-phase 증거가 progress.md 에 기록됨 (manager-develop 소관)
- [ ] 판정 기록이 `.moai/reports/t1080/verdict.md` 에 존재하고 카드 완료 보고가 그 경로를 인용함
- [ ] t1081 과의 분할 경계(spec.md §1.2)가 판정 기록에 한 줄 확인됨 (M4)

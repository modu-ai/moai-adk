# 진행 기록 — SPEC-DEPLOY-STEP-OPTIN-001

상태: draft (plan 단계 산출물 작성 완료, plan-audit 미실시). 이 파일의 §E.1 은 plan 담당(manager-spec)이 채웠고, §E.2·§E.3 은 manager-develop 이, §E.4 는 manager-docs 가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-19T12:47:02Z

- 의미: 감사를 받을 준비가 끝났다는 뜻이다. 1회차 plan-audit 의 FAIL(0.85) 개정 뒤 재제출 상태다. 2회차 감사 결과를 주장하지 않는다.
- 1회차 감사: FAIL 0.85, `.moai/reports/t1895/plan-audit.md`(HEAD 302b987142 고정). 차단 D1(REQ-DSO-016 ↔ AC-DSO-020(c) 모순), D2(배차 게이트가 MCP·`--auto-lane` 진입을 못 덮음), SHOULD-FIX D3·D4·D5, MINOR D6(무수정 — 감사 스스로 과잉 차단으로 기록).
- 개정 1(2026-10-19, spec v0.1.1): D1 을 쓰기+알림 방향으로 정렬(결정 Q15 신설, DECIDED — 운영자 상시 규칙 Jev 판정, 리더 run tn564n 중계, jev-1.13.0 p=0.89)하고 REQ-DSO-016 · AC-DSO-020(c) · AC-DSO-019 · design §7 · 시험 이름 다섯 표면을 한 규칙으로 맞췄다. D2 는 REQ-DSO-010 · design §2 · plan §C 의 게이트 지점을 공유 임대 이음매 `factoryNextLeaseOnceGated`(`factory_card.go:386`, 네 지점은 이 트리에서 직접 관측)로 확정하고 AC-DSO-009 에 MCP 진입·`--auto-lane` 진입 시험 두 개와 변이체를 더했다. D3 재범위 기록은 plan.md M1 과 spec §D.1 에, D4 순서·결과 사용 변이체는 AC-DSO-021 에, D5 는 완료 정의 (2) 를 research.md §6 대로 재진술했다.
- tier: L (spec.md · plan.md · acceptance.md · design.md · research.md, 그리고 progress.md 와 decision-index.md)
- 기준 트리: `8249bff8c3933d647b4f99e59048595f9808e83a` (브랜치 `WT-opt-in-moai`). plan 커밋: `e8e4050df5ba2bfe2dbe4d08626fcb83d1500c13`(research.md), `6859f444f44992962a580e42ebf3158a72d770c1`(spec.md · acceptance.md · decision-index.md), `8a1441e4cac1f53457df0bce3d5eb80b6651bbb7`(design.md · plan.md).
- plan_artifact_hash: `b41260a76ce97bd9d42862fa7d92c9a430e725cbf411aaef8f4e219ccacffbbd` (sha256, acceptance · decision-index · design · plan · research · spec 순, progress.md 제외, 파일명:길이 접두와 널 구분자 형식). 개정 1 상태의 값이다. 계산: `mcp__moai__audit_cache` op=compute_hash (moai MCP 서버 빌드 moai_cp/20261007_195927, 커밋 7705b365e2 — HEAD 302b987142 의 조상이나 6개 마크다운 파일의 순수 텍스트 해싱이라 조상 관계가 측정에 영향을 주지 않는다).
- 개수: 요구 18개(REQ-DSO-001–018), 인수 기준 22개(AC-DSO-001–022), 필요한 시험 최상위 49개(acceptance.md §D), 결정 행 15개(Q15 신설).
- lint: `moai spec lint SPEC-DEPLOY-STEP-OPTIN-001 --strict` → `✓ No findings — all SPEC documents are valid`, 종료 코드 0, 오류 0. 판정 빌드는 설치 CLI `v3.2.0-rc.37`(`moai_cp/20261007_195927-2046-g8249bff8c3`, 빌드 시각 2026-10-19T12:42:24Z, 빌드 커밋 `8249bff8c3` — 이 트리 HEAD 302b987142 의 조상이다. 조상 관계가 존재하므로 §2.2 대로 함께 명시한다: 측정은 SPEC 디렉터리의 텍스트 검증이라 빌드 지연이 결과에 영향을 주지 않는다).
- 열린 결정: 제품 수준 Q1–Q9, Q14 (10행) + 증거 대기 Q11 (1행). Q15 는 DECIDED(위 개정 1). 구현 수준 Q10, Q12, Q13 은 `DEFAULT-APPLIED 2026-10-19T06:05:24Z lane-4 manager-spec (plan)`. 막는 행은 decision-index.md 개요 아래에 마일스톤별로 적었다.
- 선행: run 은 t1882 착지 뒤에 시작한다. 착지한 트리에서 AC-DSO-006–008 의 RED-now 를 다시 측정하고 다시 고정한다(plan.md M1, §E).
- 측정 요약: research.md 증거 원장 L01–L63. 입력 메모와 달라진 전제는 research.md §3(t1882 미착지, "키 누락" 반증, 시험 이음매 이름 `factoryGH`, `yamlpatch` 는 디렉터리, 줄 번호 이동, automation 키는 읽는 코드가 없음, config 섹션은 템플릿-사본 비교의 예외).
- 미실행: `go test`, `go build`, 패키지 단위 도구. 자원 규칙(보드 8c1e) 측정: 1분 부하 14.75(한도 32 이하 통과), 가용 메모리 11%(한도 50% 미달) → Go 빌드·시험 금지 조건이라 문서 개정에 필요한 Go 실행을 하지 않았다. 모든 녹색 경로는 약속이다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

# 진행 기록 — SPEC-DEPLOY-STEP-OPTIN-001

상태: draft (plan 단계 산출물 작성 완료, plan-audit 미실시). 이 파일의 §E.1 은 plan 담당(manager-spec)이 채웠고, §E.2·§E.3 은 manager-develop 이, §E.4 는 manager-docs 가 채운다.

## §E.1 Plan-phase Audit-Ready Signal

audit_ready: true
plan_status: audit-ready
plan_complete_at: 2026-10-19T13:06:00Z

- 의미: 감사를 받을 준비가 끝났다는 뜻이다. 1회차 plan-audit FAIL(0.85) 개정 2건(개정 1: 감사 결함 수리, 개정 2: 제품 판정 기록 + 재범위)을 마친 재제출 상태다. 2회차 감사 결과를 주장하지 않는다.
- 1회차 감사: FAIL 0.85, `.moai/reports/t1895/plan-audit.md`(HEAD 302b987142 고정). 차단 D1(REQ-DSO-016 ↔ AC-DSO-020(c) 모순), D2(배차 게이트가 MCP·`--auto-lane` 진입을 못 덮음), SHOULD-FIX D3·D4·D5, MINOR D6(무수정 — 감사 스스로 과잉 차단으로 기록).
- 개정 1(2026-10-19T12:47Z, spec v0.1.1, 커밋 `62efc6d15a`): D1 을 쓰기+알림 방향으로 정렬(결정 Q15 신설, DECIDED — 운영자 상시 규칙 Jev 판정, 리더 run tn564n 중계, jev-1.13.0 p=0.89)하고 REQ-DSO-016 · AC-DSO-020(c) · AC-DSO-019 · design §7 · 시험 이름 다섯 표면을 한 규칙으로 맞췄다. D2 는 REQ-DSO-010 · design §2 · plan §C 의 게이트 지점을 공유 임대 이음매 `factoryNextLeaseOnceGated`(`factory_card.go:386`, 네 지점은 이 트리에서 직접 관측)로 확정하고 AC-DSO-009 에 MCP 진입·`--auto-lane` 진입 시험 두 개와 변이체를 더했다. D3 재범위 기록은 plan.md M1 과 spec §D.1 에, D4 순서·결과 사용 변이체는 AC-DSO-021 에, D5 는 완료 정의 (2) 를 research.md §6 대로 재진술했다.
- 개정 2(2026-10-19T13:06Z, spec v0.2.0): 제품 결정 행 Q1–Q9, Q14 의 판정을 운영자 상시 규칙(Jev, 불충분 시 Codex 교차; 리더 run tn564n 중계 2026-10-19)으로 decision-index 에 기록하고 SPEC 전체를 판정에 맞춰 스윕했다. Q4=(b): REQ-DSO-017 · AC-DSO-021 · 마일스톤 M7 을 빼 여섯 단계 SPEC 이 되었다(017·021 은 결번; AC-DSO-021 은 철회 노트로 자리를 표시). Q5=(a): 단계별 fallback 확정, AC-DSO-002 변이체 조건을 판정에 맞춰 해소. Q7=(c): `AutomationConfig`(`internal/config/types.go:58`)·템플릿 프로필·산문 4개 파일의 `auto_push`/`auto_pr` 제거를 M6 에 둠(이 저장소 운영자 파일은 보존). Q8=(a) OPERATOR-HELD(리더 거부권 대기) — M6 의 project 계층 이전, Q3=(a) 와 맞물려 `review_repair: auto` 포함. Q9=(b): 즉시 중지를 `.moai/state/` 정지 표식 설계로 교체(REQ-DSO-005 출처 도메인 확장, REQ-DSO-015 재작성, AC-DSO-019 재작성 + 경계 시험 `TestStopMarker_BoundariesStopWhileMarkerStands` 추가, design §3·§5·§7). Q11 은 열려 있다(t1882).
- tier: L (spec.md · plan.md · acceptance.md · design.md · research.md, 그리고 progress.md 와 decision-index.md)
- 기준 트리: `8249bff8c3933d647b4f99e59048595f9808e83a` (브랜치 `WT-opt-in-moai`). plan 커밋: `e8e4050df5ba2bfe2dbe4d08626fcb83d1500c13`(research.md), `6859f444f44992962a580e42ebf3158a72d770c1`(spec.md · acceptance.md · decision-index.md), `8a1441e4cac1f53457df0bce3d5eb80b6651bbb7`(design.md · plan.md), `62efc6d15a`(1회차 감사 개정 1).
- plan_artifact_hash: `4a2539e6ecb7d14bcf92a11c8cef03d756ea4c9e5612d67592e237225afa2bcc` (sha256, acceptance · decision-index · design · plan · research · spec 순, progress.md 제외, 파일명:길이 접두와 널 구분자 형식). 개정 2 상태의 값이다. 계산: `mcp__moai__audit_cache` op=compute_hash (moai MCP 서버 빌드 moai_cp/20261007_195927, 커밋 7705b365e2 — HEAD 의 조상이나 6개 마크다운 파일의 순수 텍스트 해싱이라 조상 관계가 측정에 영향을 주지 않는다).
- 개수: 요구 17개(REQ-DSO-017 결번), 인수 기준 21개(AC-DSO-021 결번), 필요한 시험 최상위 48개(acceptance.md §D), 결정 행 15개. 마일스톤 M1–M6.
- lint: `moai spec lint SPEC-DEPLOY-STEP-OPTIN-001 --strict` → `✓ No findings — all SPEC documents are valid`, 종료 코드 0, 오류 0(개정 1·2 각각에서 재실행해 확인). 판정 빌드는 설치 CLI `v3.2.0-rc.37`(`moai_cp/20261007_195927-2046-g8249bff8c3`, 빌드 시각 2026-10-19T12:42:24Z, 빌드 커밋 `8249bff8c3` — 이 트리 HEAD 의 조상이다. 조상 관계가 존재하므로 §2.2 대로 함께 명시한다: 측정은 SPEC 디렉터리의 텍스트 검증이라 빌드 지연이 결과에 영향을 주지 않는다).
- 결정 상태: DECIDED — Q1–Q9, Q14, Q15 (11행). DEFAULT-APPLIED — Q10, Q12, Q13. 열림 — Q11 (증거 대기, t1882). Q8 은 OPERATOR-HELD(리더 거부권 대기).
- 선행: run 은 t1882 착지 뒤에 시작한다. 착지한 트리에서 AC-DSO-006–008 의 RED-now 를 다시 측정하고 다시 고정한다(plan.md M1, §E).
- 측정 요약: research.md 증거 원장 L01–L63. 입력 메모와 달라진 전제는 research.md §3(t1882 미착지, "키 누락" 반증, 시험 이음매 이름 `factoryGH`, `yamlpatch` 는 디렉터리, 줄 번호 이동, automation 키는 읽는 코드가 없음, config 섹션은 템플릿-사본 비교의 예외). 개정 2가 새로 관측한 트리 사실: `AutomationConfig` 의 `AutoPR`·`AutoPush` 필드(`internal/config/types.go:58`), 대칭 시험 3개(`internal/config/git_strategy_nested_test.go`), `auto_push` 를 싣는 템플릿 파일 목록(git-strategy.yaml.tmpl, manager-git.md, manager-git.toml, spec-assembly.md, tab_schema.json — grep, 이 트리).
- 미실행: `go test`, `go build`, 패키지 단위 도구. 자원 규칙(보드 8c1e) 측정(개정 1 시점): 1분 부하 14.75(한도 32 이하 통과), 가용 메모리 11%(한도 50% 미달) → Go 빌드·시험 금지 조건이라 문서 개정에 필요한 Go 실행을 하지 않았다. 모든 녹색 경로는 약속이다.

## §E.2 Run-phase Evidence

_<pending run-phase>_

## §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase>_

---
id: SPEC-WEB-SETTINGS-SAVE-001
title: "acceptance — moai web 설정 저장 불가 수리 · 감사 pin 기본값 · 원격병합 가드"
version: "0.1.0"
created: 2026-10-01
updated: 2026-10-01
author: GOOS
---

# acceptance — SPEC-WEB-SETTINGS-SAVE-001

> 모든 AC 행은 기계 판정 가능해야 한다. 판정 명령 + 원문 출력은 `.moai/reports/t1393/verdict.md`로 반출한다.
> 재측정 단위(전 행 공통): `unset MOAI_AUTONOMY_TIER MOAI_CONFIG_SOURCE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID && go test -timeout 30m ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/settings/... ./internal/profile/...` — 명시 변수 나열 단일 호출(`unset MOAI_*` 글로브 폐기 — plan-audit D3 실측: 0-매치 환경에서 미실행). 나열 집합은 실행 레인의 생존 변수를 사전 재유도(`env | grep -o '^MOAI_[A-Z0-9_]*' | sort`).

## §D.0 Blocker 행의 RED-now/기준선 셀 처분 (verification-completeness.md §2.1)

- **오늘 측정 고정(기준선 셀)**: AC-WSS-005 · AC-WSS-014 — 수정 전 트리 기준선을 증거 원장(§D.8)에 4요소(명령 · 원문 stdout · exit code · tree SHA)로 고정했다(EV-001 · EV-002).
- **regression-guard 재분류(§2.1 undecidable disposition)**: AC-WSS-001 · AC-WSS-002 · AC-WSS-003 · AC-WSS-004 · AC-WSS-010 — 시작 관측이 M1 진단 산출물(운영자 실효 페이로드 기반 관측-RED 재현 테스트와 그 4요소 셀)에 의존해 plan-phase에서 재실행할 수 없으므로, release-blocking 자격을 잃고 **regression-guard로 분류**되며 pass로 기록되지 않는다. 각 행은 M1의 관측-RED 4요소 셀이 `.moai/reports/t1393/`에 착지하는 시점부터 release-blocking으로 활성화된다.

## §D AC Matrix

| AC | 스코프 | 요구사항 | 심각도 | 판정 |
|---|---|---|---|---|
| AC-WSS-001 | ① | REQ-WSS-101 | Blocker | 관측-RED 재현 증거 |
| AC-WSS-002 | ① | REQ-WSS-102 | Blocker | 토글 영속 테스트 GREEN |
| AC-WSS-003 | ① | REQ-WSS-103 | Blocker | 감사 필드 영속 테스트 GREEN |
| AC-WSS-004 | ① | REQ-WSS-104 | Blocker | 시늅 실패 배너 노출 테스트 GREEN |
| AC-WSS-005 | ① | REQ-WSS-105 · REQ-WSS-106 | Blocker | save_lossless + 저장 계약 기존 테스트 무수정 GREEN |
| AC-WSS-006 | ② | REQ-WSS-201 | Major | claude 기본값 단언 GREEN |
| AC-WSS-007 | ② | REQ-WSS-201 | Major | glm 핀 3중 정합 GREEN |
| AC-WSS-008 | ② | REQ-WSS-201 | Major | codex 불변 GREEN |
| AC-WSS-009 | ② | REQ-WSS-203 | Major | 관측-RED 증거 |
| AC-WSS-010 | ③ | REQ-WSS-301 · REQ-WSS-306 | Blocker | 미확증 거부 테스트 GREEN |
| AC-WSS-011 | ③ | REQ-WSS-304 | Major | 착지 경로 통과 테스트 GREEN |
| AC-WSS-012 | ③ | REQ-WSS-302 | Major | fail-open 보존 테스트 GREEN |
| AC-WSS-013 | ③ | REQ-WSS-303 | Major | merge-first 순서 보존 GREEN |
| AC-WSS-014 | 전체 | REQ-WSS-106 | Blocker | 소관 5패키지 전량 GREEN |
| AC-WSS-015 | ③ | REQ-WSS-305 | Major | cleanupSessionWorktreeFn 시늅 경유 관측 GREEN |
| AC-WSS-016 | ② | REQ-WSS-202 · REQ-WSS-204 | Major | 핀 우선순위·gates/enum 무변경 GREEN |

## §D.1 스코프 ① — 설정 저장 수리

- **AC-WSS-001 (관측-RED 재현 — regression-guard, §D.0)** Given 운영자 실효 환경에서 캡처한 실제 실패 요청 페이로드(전체 폼 POST — 교차 탭 hidden 입력 포함; `workflow.worktree.auto_create`·`auto_merge` 중첩 키 + `workflow.audit.*` 필드), When 그 페이로드 형태의 재현 테스트가 수정 전 트리(f130aa041)에서 실행되면, Then 디스크 workflow.yaml 무기록이 적색으로 관측되고 4요소 셀이 `.moai/reports/t1393/`에 착지한다(수정 후 동일 테스트 GREEN). **미재현 분기(명문화)**: 본 트리에서 재현되지 않으면 구현을 강행하지 않고 (a)/(b) 가설 재탐증 + 운영자 실행 바이너리 편차 진단으로 분기한다(decision-index Q2 — codex overlay 보고 "최소 제출은 정상 저장", 중간 신뢰도·미검증 Gap).
- **AC-WSS-002 (토글 영속)** Given 시드 프로젝트(workflow.yaml 존재, 주석·미모델링 키 포함), When POST /save가 worktree 토글 편집을 실어 보내면, Then `.moai/config/sections/workflow.yaml`의 `workflow.worktree.auto_create`·`auto_merge`가 제출값으로 기록되고 주석·미모델링 키가 보존된다.
- **AC-WSS-003 (감사 필드 영속)** Given 동일 시드, When POST /save가 `workflow.audit.*` 편집(예: gates.claude 값 변경)을 실어 보내면, Then 해당 키가 디스크에 기록된다.
- **AC-WSS-004 (무음 실패 부재)** Given 시늅 쓰기 실패 주입(쓰기 불가 디렉터리 등), When POST /save가 실행되면, Then 응답은 2xx 재렌더이고 오류 배너가 사용자에게 노출된다(SPEC-WEB-CONSOLE-017 계약 유지) — 상태코드 4xx/5xx로 사용자에게 무음인 경로가 새로 생기지 않는다.
- **AC-WSS-005 (무손실 회귀 0)** Given t1314의 무손실 테스트(`internal/settings/save_lossless_test.go` · `internal/settings/nested_test.go` · `internal/profile/sync_lossless_test.go`) + 저장 표면의 기존 계약 테스트(원자 거절, 중복 폼 값 거절 REQ-WWS-006, 닫힌 집합 검증, GLM/Jev 자격증명 시늅), When 스코프 ① 수리가 병합될 때, Then 해당 테스트들은 **무수정**으로 GREEN이다(REQ-WSS-105의 무손실 + REQ-WSS-106의 회귀 0). **판정 명령**: `unset <명시 10변수 — §D 전 행 공통> && go test -timeout 30m ./internal/settings/... ./internal/profile/... ./internal/config/...`. **수정 전 기준선**: 증거 원장 EV-002(§D.8).

## §D.2 스코프 ② — 감사 핀 기본값

- **AC-WSS-006 (claude 핀)** Given `NewDefaultWorkflowConfig()`, Then `Workflow.Audit.Claude == {claude-opus-5-5, high}`이다. 판정: internal/config 해당 단언 테스트 GREEN.
- **AC-WSS-007 (glm 핀 3중 정합)** Given 세 표면 — (1) `NewDefaultWorkflowConfig()`, (2) `internal/template/templates/.moai/config/sections/workflow.yaml` audit 블록, (3) GLM 감사 resolver 터미널 폴백 — Then 모두 `{glm-5.3, max}`이다.
- **AC-WSS-008 (codex 불변)** Given codex 핀, Then `{gpt-6.1-sol, high}`가 세 표면에서 유지된다(기존 테스트 무수정 GREEN — 검증만).
- **AC-WSS-009 (관측-RED)** Given 기존 기본값 단언 테스트(claude medium 픽스처 등), When 값 변경이 적용되기 전이면, Then 해당 테스트가 새 기대값에 대해 적색으로 관측됐다는 증거(출력)가 `.moai/reports/t1393/`에 있고, 갱신은 그 이후에 커밋된다.

## §D.3 스코프 ③ — 원격병합 가드

- **AC-WSS-010 (미확증 거부 — regression-guard, §D.0)** Given clean worktree + clean exit + 브랜치가 리모트에 있으나 채택 술어 미성립(푸시됐지만 `refs/remotes/origin/develop`에 미도달 — REQ-WSS-302 술어), When `cleanupSessionWorktree`가 실행되면, Then 제거가 거부되고 공지가 미확인 상태를 명명하며 워크트리가 보존된다.
- **AC-WSS-011 (착지 경로 통과)** Given 브랜치 팁이 `refs/remotes/origin/develop`에 도달 가능(`git merge-base --is-ancestor`) 또는 `git cherry refs/remotes/origin/develop <branch>` 공집합(스쿼시 병합 등가) + clean + clean exit, When 정리가 실행되면, Then 오늘과 동일하게 제거가 진행된다(공지 포함) — 추가 대화형 마찰 0.
- **AC-WSS-012 (fail-open 보존)** Given 병합 확인 검사의 이상(검사 오류 · 원격추적 통합 참조 부재 · 참조 소실 · 비정상 갱신 · 도달 판정 불신뢰 — REQ-WSS-302 ①②③), When 정리가 실행되면, Then 보존 + 공지가 발생하고 제거가 일어나지 않는다.
- **AC-WSS-013 (순서 보존)** Given init/web/profile 종료 경로, When 세션이 종료되면, Then `sessionExitAutoMerge`가 `cleanupSessionWorktree`(및 신규 가드)보다 먼저 실행된다 — 기존 순서 테스트 무수정 GREEN.
- **AC-WSS-015 (시늅 경유)** Given `cleanupSessionWorktreeFn` 시늅을 오버라이드하는 커맨드 수준 테스트, When 세 종료 경로(web.go:115 · init.go:466 · profile.go:80)에서 각각 실행되면, Then 원격병합 가드가 시늅을 경유해 동작함이 관측된다 — 가드가 호출부 우회 경로를 만들지 않는다.
- **AC-WSS-016 (우선순위·무변경)** Given 명시 핀을 가진 project workflow.yaml + 감사 gates 분포(claude/codex required, glm advisory) + `audit_model` enum + REQ-AMP-008 감사전용 경로, When 스코프 ② 변경이 적용되면, Then 명시 핀이 새 기본값을 이기고(기존 핀 우선순위 테스트 GREEN), gates/enum/감사전용 경로는 무변경이다(REQ-WSS-202·REQ-WSS-204 — 기존 테스트 무수정 GREEN).

## §D.4 전체 재측정

- **AC-WSS-014 (소관 패키지 전량)** Given 최종 트리, When 재측정 단위 명령(§D 전 행 공통 — 5패키지)이 실행되면, Then 전량 GREEN이고 원문 출력이 verdict.md에 반출된다. **수정 전 기준선**: 증거 원장 EV-001(§D.8). (로컬 전체 스위트 실행 금지 — 전 패키지 판정은 CI.)

## §D.5 간접 검증

- 템플릿 갱신의 분배 경로: M4에서 `make build` 후 `git status --short`에 재생성 드리프트 없음 + `go test ./internal/template/...` 해당 축 GREEN.
- 시늅 배선: `cleanupSessionWorktreeFn`을 경유하는 3호출부(web.go:115 · init.go:466 · profile.go:80)가 가드를 상속함을 grep/테스트로 확인.
- 변경 없음 확인: 감사 gates 분포·`audit_model` enum·REQ-AMP-008 경로(태스크 위임) 무변경 — REQ-WSS-204 단언 테스트 GREEN 유지.

## §D.6 품질 게이트

- TRUST 5: 신규/수정 코드 커버리지 기준 준수(패키지 85%), gofmt/goimports 통일, 입력 검증(웹 경계) 유지, Conventional Commits + 카드 id.
- `go vet` 변경 패키지 + `golangci-lint run`(CI 판 v2.1.6).
- M4에서만: `make build`(선행 emit-check 포함) 통과.

## §D.7 Definition of Done

1. AC-WSS-001..016 전 행 GREEN + 증거 반출(verdict.md).
2. 관측-RED 증거 2종(스코프 ① 재현, 스코프 ② 단언) 보존.
3. decision-index.md 미결 행의 운영자/리드 판정 반영 또는 명시적 보류 기록.
4. 소관 패키지 재측정 원문 출력이 verdict.md에 존재(§2 귀속: 명령 + 관측 출력).
5. sync 진입 조건: 위 1-4 충족 + progress.md §E.2·§E.3 run 증거 기록 완료.

## §D.8 증거 원장 (evidence ledger — §2.1 4요소 캐리어)

> MP-8/D1 수리: Blocker 행의 시작 관측을 4요소로 고정한다. 캐리어는 이 원장(펜스 대장)이고 표 셀은 id로 인용한다.

### EV-001 — AC-WSS-014 수정 전 기준선 (소관 5패키지)

- command (단일 호출): `go test -timeout 30m ./internal/web/... ./internal/cli/... ./internal/config/... ./internal/settings/... ./internal/profile/...`
- 실행 맥락(4요소 외 부기): 레인 정화 래퍼 `unset MOAI_AUTONOMY_TIER MOAI_CONFIG_SOURCE MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS MOAI_KANBAN_BACKEND MOAI_KANBAN_ID MOAI_KANBAN_SETTINGS_INJECTED MOAI_LAUNCH_PROVIDER MOAI_PROFILE_LEASE_TOKEN MOAI_SESSION_PID && <위 명령>` 을 1회 호출로 실행 — 래퍼는 명령 요소가 아니라 실행 맥락이다. 원문 전문: `.moai/reports/t1393/baseline-prefix-20261001.txt`
- verbatim stdout (꼬리): <PENDING-BID67HBR9>
- exit code: <PENDING-BID67HBR9>
- tree SHA: f130aa041bb90c81b235029afffa806a65e88480

### EV-002 — AC-WSS-005 수정 전 기준선 (무손실 대상 3패키지 행)

- command: EV-001과 동일 실행의 `./internal/config/...` · `./internal/settings/...` · `./internal/profile/...` 행
- verbatim stdout: <PENDING-BID67HBR9>
- exit code: <PENDING-BID67HBR9>
- tree SHA: f130aa041bb90c81b235029afffa806a65e88480

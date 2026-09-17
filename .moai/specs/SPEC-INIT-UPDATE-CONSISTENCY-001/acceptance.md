---
id: SPEC-INIT-UPDATE-CONSISTENCY-001
title: "acceptance criteria — init/update 정합성·문서 정리"
created: 2026-09-13
---

# acceptance.md — SPEC-INIT-UPDATE-CONSISTENCY-001

## §A 수용 원칙

모든 AC는 기계 판정 가능해야 한다(테스트 출력·grep 결과·렌더 스냅샷). "정리됐다" 류의 주관 표현은 AC 로 쓰지 않는다.

## §B 질 게이트

- 영향 패키지(`internal/cli`, `internal/config`, `internal/core/project`, `internal/web`, `internal/template`) 테스트 전건 통과.
- `GOOS=windows go build ./...` 성공.
- `golangci-lint run` 영향 패키지 0 error.

## §C 정의

- "페어": 같은 stripped-target 경로로 수렴하는 `.sh` + `.sh.tmpl` 템플릿 쌍.
- "백업 뿌리": `.moai-backups/`(config), `.moai/backups/update-<hyphenated-ISO>/`(namespace), `.moai/archive/skills/<drift>/`(archive-drift).

## §D AC 행렬

### AC-001 (REQ-ICU-001 / F8)

- **Given** 템플릿에서 `mode:` 키가 제거된 project.yaml.tmpl 이 임베드된 바이너리
- **When** `moai init` 이 신규 프로젝트를 초기화하면
- **Then** 생성된 `project.yaml` 에 `mode` 키가 없고, `--project-mode` 플래그는 존재하지 않으며(플래그 파싱 오류), 리포지토리 전체 grep 에서 project.yaml `mode` 키의 Go 리더가 0건이다.

### AC-002 (REQ-ICU-002 / F9)

- **Given** `internal/config/defaults.go` 의 `NewDefaultWorkflowConfig`
- **When** parity 테스트가 컴파일 기본과 템플릿 workflow.yaml 파싱값을 비교하면
- **Then** 양쪽 모두 `auto` 이고, `TestExecutionModeDefaultIsInSet` 은 계속 통과한다.

### AC-003 (REQ-ICU-003 / F12)

- **Given** `handle-agent-hook.sh` + `handle-agent-hook.sh.tmpl` 페어를 포함한 템플릿 목록
- **When** update가 managed 재배포 계수를 계산하면
- **Then** 페어는 1로 계수되고(4페어 기준 총계 4), outcome 요약의 "Updated N" 은 실제 배포 대상 수와 일치한다.

### AC-004 (REQ-ICU-004 / F13)

- **Given** config 백업과 namespace 백업이 모두 생성된 update 실행
- **When** outcome 요약이 렌더되면
- **Then** 생성된 백업 뿌리 전건이 각각의 경로와 함께 표기되고, 단일 뿌리 실행의 출력은 기존 계약 테스트와 동일하게 유지된다.

### AC-005 (REQ-ICU-005 / F14)

- **Given** `settings.AllFields()` 의 editable FieldDef 전건
- **When** parity 테스트가 렌더 표면 대응을 검사하면
- **Then** 모든 필드가 (a) 패널/전용 컴포넌트에 렌더되거나 (b) 근거 주석과 함께 exempt 목록에 명시되며, 양쪽 어디에도 없는 필드로 테스트가 실패하지 않는다.

### AC-006 (REQ-ICU-006 / F15)

- **Given** `evaluator-profiles/` 커스터마이징이 존재하는 프로젝트에서 update 실행
- **When** outcome 요약이 렌더되면
- **Then** 비-섹션 config 디렉터리가 병합 복원 대상이 아님을 밝히는 안내 행과 백업 경로가 표기된다.

### AC-007 (REQ-ICU-007 / F17, 기록)

- **Given** `internal/config/manager.go` 의 `Save`
- **When** godoc과 어노테이션을 grep 하면
- **Then** 6섹션 범위 선언 godoc + `@MX:DEBT` + `@MX:CEILING` + `@MX:UPGRADE` 가 각 1건 이상 존재한다.

### AC-008 (REQ 없음 — F16 소멸 회귀 가드)

- **Given** 재구성 질문 집합
- **When** `ReconfigureQuestions` 와 `InitQuestions` 의 ID 집합을 대조하면
- **Then** 기존 `AC-WIZ-012a` 테스트가 계속 통과한다(감사 선택 질문이 어느 집합에도 (재)도입되지 않았음을 간접 고정).

## §E 경계 케이스

- `.sh` 단독(페어 아님) 파일은 기존대로 1 계수 — dedupe 가 단독 파일을 건드리지 않는지 AC-003 테스트에 단독 케이스 포함.
- 백업 뿌리가 0개인 드라이런 — 요약에 백업 행이 새로 생기지 않는다.
- `project.yaml` 에 이미 `mode: team` 이 박힌 기존 프로젝트 — update 후 키 소멸이 데이터 손실이 아님을(리더 0건) 확인하는 문서 주석.

## §F 간접 검증

- F16 소멸은 신규 코드로 재증명하지 않는다 — 기존 `AC-WIZ-012a` 테스트의 존속이 간접 가드(AC-008)다.
- F13 분산 본체(record-only)는 `update_namespace_protect.go` 패키지 문서의 기존 기록으로 검증하며, run-phase 산출물을 만들지 않는다.

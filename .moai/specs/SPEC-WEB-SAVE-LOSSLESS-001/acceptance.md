# SPEC-WEB-SAVE-LOSSLESS-001 Acceptance Criteria

> 검증 계층 — 모든 항목은 Given-When-Then으로 기술되고 binary-testable하다.
> SEVERITY: S2(데이터 손실 — 무손실 계약 본체) / S3(가드·관측성).

## §D AC Matrix

| AC | 요구 | 시나리오 요지 | SEVERITY |
|---|---|---|---|
| AC-WSL-001 | REQ-WSL-001 | 무편집 Save → 전 섹션 파일 byte-identical | S2 |
| AC-WSL-002 | REQ-WSL-002 | 한 필드 편집 → 그 행만 변경 | S2 |
| AC-WSL-003 | REQ-WSL-003 | 미모델링 키 + 주석 영생존 | S2 |
| AC-WSL-004 | REQ-WSL-004 | 부재-키 빈 제출 무기록 (workflow.yaml audit pin) | S2 |
| AC-WSL-005 | REQ-WSL-005 | user 이름 편집 시 user 섹션 보존 | S2 |
| AC-WSL-006 | REQ-WSL-006, REQ-WSL-007, REQ-WSL-008 | 파일-부재/기록-실패 시 부분-저장 상태 판독 가능 + per-file 원자 쓰기 | S3 |
| AC-WSL-007 | REQ-WSL-010 | 손실-행위 단언 테스트 소멸 | S3 |
| AC-WSL-008 | REQ-WSL-009 | 선행 SPEC 게이트 회귀 없음 | S2 |

## §D.1 AC-WSL-001 — 무편집 Save는 모든 섹션에서 byte-identical (S2)

- **Given** 미모델링 키(`user.github_username`, `constitution.session_effort_default`)와 주석을 포함해 시드된 프로젝트(`.moai/config/sections/*.yaml` 전체 — user, language, quality, git-strategy, git-convention, llm typed 6종 + workflow, harness, ralph, feedback, observability, security, crosssession seam)에서,
- **When** `moai web` 설정 폼을 현재값 그대로 Save로 제출하면,
- **Then** 제출 직전과 직후의 모든 섹션 파일이 `git diff --stat`에서 빈 출력(또는 샤 해시 동일)이고, mtime도 불변이다.

변이: (a) 폼 생략형 비(非)브라우저 POST(키 없는 필드 = preserve), (b) bool hidden companion 미제출, (c) EmptySubmits 필드의 현재값 제출 — 모두 동일 Then.

## §D.2 AC-WSL-002 — 한 필드 편집은 그 필드의 행만 바꾼다 (S2)

- **Given** §D.1과 동일 시드에서,
- **When** 정확히 한 필드(예: `llm.glm.models.high`)의 값만 바꿔 Save하면,
- **Then** `git diff`가 llm.yaml의 그 값 행 1행만 보여주고, 다른 모든 파일과 llm.yaml의 나머지 행(주석·빈 줄·키 순서 포함)은 diff에 없다.

변이: git_strategy mode / quality DDD bool / workflow 감사 pin 각각 1회 — typed·seam 양 경로 모두.

## §D.3 AC-WSL-003 — 미모델링 키 + 주석 영생존 (S2)

- **Given** `user.github_username: example-user`, `constitution.session_effort_default: xhigh  # local note`, `llm.claude_models` 블록이 시드돼 있고,
- **When** (a) 무편집 Save, (b) 이 키들과 **다른** 필드의 실제 편집 Save를 각각 수행하면,
- **Then** 두 경우 모두 세 키와 그 주석이 제출 후 파일에 값·주석 그대로 존재한다(바이트 비교 또는 키+주석 행 원문 비교).

## §D.4 AC-WSL-004 — 부재-키 빈 제출은 기록하지 않는다 (S2)

- **Given** workflow.yaml에 `workflow.audit.claude.model` 키가 **부재**하고,
- **When** 해당 필드가 `""`(미설정)로 제출된 Save를 수행하면,
- **Then** workflow.yaml이 byte-identical이다 — `model: ""` 빈 키가 추가되지 않는다.

변이: (a) 키가 **존재**하는 상태에서 `""` 제출 → `""`로의 실제 변경은 기록됨(삭제 시맨틱 유지 — `TestCrossSessionEmptySubmitsRoundTrip` 계승), (b) 부재 + 실제값 제출 → upsert 기록됨.

## §D.5 AC-WSL-005 — user 섹션은 전체-교체되지 않는다 (S2)

- **Given** user.yaml에 `name` 외 키(github_username 등)가 있고,
- **When** 웹 폼에서 사용자 이름을 변경해 Save하면,
- **Then** user.yaml에서 `name` 행만 변경되고 나머지 키·주석이 생존한다.

## §D.6 AC-WSL-006 — 부분-저장 보증 (S3)

- **Given** 하나의 writer가 기록에 실패하도록 강제된 상태(예: 대상 디렉터리 비허가)에서,
- **When** Save가 실패 응답을 내면,
- **Then** (a) 폼 검증 실패·사전-플라이트 실패 케이스는 디스크 무변경이고, (b) 중간 실패 케이스의 오류 메시지는 어느 단계가 실패했는지 식별 가능하다(현행 `logSaveFailure` 단계명 계승), (c) 모든 성공 기록은 temp+rename 원자 쓰기다(REQ-WSL-007).

회귀 고정: `.moai/config/sections/mcp.yaml`이 없는 프로젝트에서의 Save는 HTTP 500이 아니라 정상 완결이다(SPEC-SEAM-GREENFIELD 착지 고정 — `e365c2d30`).

## §D.7 AC-WSL-007 — 손실-행위 단언 테스트 소멸 (S3)

- **Given** 구현 브랜치에서,
- **When** `internal/{settings,config,profile,web}`의 테스트를 점검하면,
- **Then** "무편집 Save가 파일을 재기록한다", "미모델링 키가 소실된다", "주석이 제거된다", "빈 키가 추가된다"를 긍정 단언하는 테스트는 0건이고, 각 손실 행위의 부정을 단언하는 테스트가 §D.1-§D.5에 대응해 존재한다.

## §D.8 AC-WSL-008 — 선행 게이트 회귀 없음 (S2)

- **When** 영향 패키지 4개(`internal/settings`, `internal/config`, `internal/profile`, `internal/web`)의 기존 테스트를 실행하면,
- **Then** SPEC-WEB-WRITE-SAFETY-001의 값-불변 게이트 테스트(`write_safety_test.go` — settings 13건, web 7건), SPEC-GITSTRATEGY-SAVE-ISOLATION-001 게이트, `TestCrossSessionEmptySubmitsRoundTrip`, `TestPatchFileValueInvariantPreservesBytes`가 전부 통과한다.

## §E Quality Gates

- 영향 패키지 4개 `go test` 통과 (전체 스위트는 CI — CLAUDE.local.md §4).
- `golangci-lint run` (CI 판 v2.1.6) 0 error.
- `moai spec lint SPEC-WEB-SAVE-LOSSLESS-001` 0 findings.
- C1 위반 grep: `internal/web` 신규 `yaml.Marshal`/`os.WriteFile` 0건.

## §F Definition of Done

- AC-WSL-001..008 전부 PASS (근거: 실측 출력 인용 — verification-claim-integrity §2).
- 이슈 재현 절차(GitHub #1731)를 본 트리에서 재실행해 3대 손실(user/quality/llm)·빈 키(workflow) 모두 미발생 실측.
- 회귀 테스트 전환 목록(§D.7)이 plan-audit에서 검증 가능한 형태로 남는다.

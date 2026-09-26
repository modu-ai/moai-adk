---
id: SPEC-CODEX-RESUME-SCOPE-001
document: plan
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
card: t1201
---

# Plan — SPEC-CODEX-RESUME-SCOPE-001

## §A 맥락

- 카드 t1201, GitHub 이슈 #1715. 소유 요구사항은 SPEC-CODEX-PHASE2-001 REQ-CX2-008.
- 전제는 이번 실행에서 측정했다(`spec.md` §A.2 탐침, §A.3 MCP 서버 해석 경로).
- Tier M: 변경 대상은 `internal/cli` 의 `codex_task.go`, `codex_jobs.go`, `mcp_server.go`, `codex_task_test.go`(와 필요하면 `codex_job_control` 계열 테스트) — 5파일 안팎, 테스트 포함 300~600 LOC 추정. 요구사항 9개·인수 기준 13개로 Tier M 상한(16/16) 안이다.
- 개발 방식: `quality.yaml` 의 development_mode 를 따른다. 기존 동작 보존(`TestCodexTask_ResumeLast*`)이 핵심이므로 RED 테스트(오선택 재현)를 먼저 둔다.

## §B 운영자 확인 항목 (Implementation Kickoff)

[NEEDS CLARIFICATION: 설계 선택지 확정 — 권고는 혼합안(A + B 모호성 폴백, `spec.md` §C.3). A 단독이면 REQ-CRS-006 과 AC-CRS-006·011 을 제외하고, B 단독이면 REQ-CRS-003·004·005·008 과 AC-CRS-001·003·004·008·009·010·012 를 제외한다. 선택지에 맞춰 REQ-CRS-002 의 "selected by" 목록을 줄인다.]

[NEEDS CLARIFICATION: `thread_id` 허용 범위 — 권고는 "이 프로젝트 레지스트리에 기록된 스레드만"(REQ-CRS-003). 대안은 임의 스레드 id 허용이며, 그 경우 codex 쪽 스레드 저장소(`CODEX_HOME`)에 있는 다른 프로젝트의 스레드까지 재개할 수 있게 된다.]

[NEEDS CLARIFICATION: 모호 판정 결과의 오류 표면 — 권고는 기존 도구 관례대로 구조화 JSON 결과(`status: "failed"`, `error_code: "resume_ambiguous"`)로 반환하고 MCP 전송 오류로 만들지 않는 것. 호출자 조치가 필요한 결함에 `IsError` 를 켜는 기존 두 분기(프롬프트 누락, 상태 디렉터리 쓰기 불가)와 맞춰 `IsError: true` 를 함께 켤지는 운영자 결정.]

## §C 사전 점검 (run 착수 전)

1. `git -C <worktree> rev-parse --short HEAD` 와 `git merge-base HEAD origin/develop` 로 기준을 기록한다.
2. `go test ./internal/cli -run 'TestCodexTask_ResumeLast' -count=1` 로 기존 두 테스트가 기준 트리에서 통과함을 기록한다(보존 기준선).
3. 탐침 `.moai/reports/t1201/premise-probe_test.go.txt` 를 RED 테스트로 옮겨 기준 트리에서 **실패**함을 관측한다(`thread/resume threadId` 가 `thr-6`).

## §D 제약

- fixture transport(`withCodexSession`, `codexTaskScript`, `sentParams`, `countSentMethod`)만 사용한다. LIVE codex 호출 0회.
- 새 레코드 필드는 JSON `omitempty` 로 추가해 기존 레코드 파일과 읽기 호환을 유지한다(키 없는 과거 레코드는 `work_key` 없음으로 읽힌다).
- 모호 판정·미기록 `thread_id`·잘못된 `work_key` 경로에서는 codex 프로세스를 띄우지 않는다(`codexLookPath` 이후, `openCodexSessionOn` 이전에 판정).
- 전체 스위트를 로컬에서 돌리지 않는다. `./internal/cli` 만 재고, 전 패키지 판정은 CI 에 맡긴다.

## §E 자기 검증 (run 종료 시)

- `go test ./internal/cli -run 'TestCodexTask_|TestCodexJob' -count=1` 전체 통과, 출력 원문 인용.
- `go vet ./internal/cli/...`, `golangci-lint run ./internal/cli/...` 무결.
- `acceptance.md` 의 AC 매트릭스 행마다 PASS/FAIL 과 해당 테스트 이름.

## §F 마일스톤 (결정이 바뀔 가능성이 큰 순서)

### M1 — 선택 규칙과 결과 모양 확정 (우선순위 High)

- Kickoff 에서 §B 세 항목을 확정하고, 제외되는 REQ/AC 를 `spec.md`·`acceptance.md` 에 반영한다(선택지에 따라 AC 수가 바뀌면 AC 스냅숏 재생성을 같은 커밋에 넣는다 — `.moai/docs/ac-count-baseline-refresh.md` §2 네 번째 행).
- 결과 구조의 새 필드를 정한다: `resume_basis`, `error_code`, `candidate_total`, `candidates[]`, `unused_selectors[]` 계열. 이름은 run 단계에서 기존 `CodexTaskResult` 명명 관례에 맞춘다.

### M2 — 레코드 모델: `work_key` 필드 (우선순위 High, A 채택 시)

- `CodexJobRecord` 와 `codexJobSpec` 에 `work_key` 를 더하고 백그라운드 생성 경로에서 기록한다. `codex_job_status` 가 그대로 노출한다.
- 입력 검증(REQ-CRS-005): 공백 제거 후 비어 있음, 128바이트 초과, 제어 문자 → 구조화 거절.

### M3 — 선택 함수 교체 (우선순위 High)

- `latestThreadID()` 를 "선택자 → 후보 집합 → 판정" 구조로 바꾼다: `thread_id` 조회, `work_key` 필터 후 최신, 선택자 없음이면 서로 다른 스레드 수로 분기(0 → 새 스레드, 1 → `sole_thread`, 2 이상 → `resume_ambiguous` 와 후보 10개).
- 읽을 수 없는 레코드를 건너뛰는 기존 fail-open 은 유지한다.

### M4 — 도구 스키마와 설명 (우선순위 Medium)

- `mcp_server.go` 의 `codex_task` 등록부에 채택된 선택자 입력을 추가하고 `resume_last` 설명을 REQ-CRS-009 에 맞춰 고친다.

### M5 — 테스트와 회귀 보존 (우선순위 Medium)

- AC-CRS-001~013 의 fixture 테스트를 `codex_task_test.go` 에 추가한다. 기존 `TestCodexTask_ResumeLastReusesRecordedThread`·`TestCodexTask_ResumeLastWithNoRecordedThread` 는 수정 없이 통과해야 한다.

## §G 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| B 채택 후 레코드 누적으로 `resume_last` 가 상시 모호 판정 | 단일 흐름 사용자의 체감 회귀 | 혼합안에서는 거부 결과가 선택자 값을 담아 한 번에 복구된다. 레코드 수명 관리는 범위 밖(후속 카드) |
| 기존 레코드에 `work_key` 가 없음 | 과거 작업은 `work_key` 로 찾을 수 없음 | 설계상 수용. `thread_id` 선택자로 재개 가능 |
| `updated_at` 동률 | 후보 순서 비결정 | 레코드 id 로 2차 정렬(REQ-CRS-006) |
| `request_summary` 가 후보 목록으로 다른 작업 항목에 노출 | 같은 프로젝트 안의 정보 노출 | 이미 레코드 생성 시 축약·마스킹된 값이며 `codex_job_status` 로도 읽을 수 있는 수준 |

## §H 교차 참조

- SPEC-CODEX-PHASE2-001 REQ-CX2-003(레코드 생성), REQ-CX2-007(sticky `sandboxPolicy` — 재개 경로 변경이 매 턴 명시 전송을 깨지 않아야 함), REQ-CX2-008(`resume_last`).
- `.claude/rules/moai/core/moai-mcp-tools.md` § `project_root` 입력 — `codex_task` 는 그 9개 도구 목록에 없다(§A.3).

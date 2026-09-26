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
- 전제는 이번 실행에서 측정했다(`spec.md` §A.2 탐침, §A.3 MCP 서버 해석 경로, §A.4 fixture 응답 id).
- Tier M: 변경 대상은 `internal/cli` 의 `codex_task.go`, `codex_jobs.go`, `mcp_server.go`, `codex_task_test.go`(와 필요하면 `codex_job_control` 계열 테스트) — 5파일 안팎. 요구사항 10개·인수 기준 15개로 Tier M 상한(16/16) 안이다.
- 개발 방식: `quality.yaml` 의 development_mode 를 따른다. 기존 동작 보존(`TestCodexTask_ResumeLast*`, `TestCodexTask_SandboxPolicyResetOnReusedThread`)이 핵심이므로 RED 테스트(오선택 재현)를 먼저 둔다.
- plan-audit iter-1 FAIL(`.moai/reports/t1201/plan-audit.md`) 의 D1~D8 과 선택 항목 D9~D13 을 v0.2.0 에서 수리했다.

## §B 운영자 결정 (Implementation Kickoff 에서 확정)

운영자가 질문 창 안에 답하지 않아, 세 항목 모두 권고 기본값으로 잠정 결정했다. Kickoff 에서 운영자가 확정하거나 뒤집는다. 뒤집으면 `spec.md` §C.5 의 유래 표를 출발점으로 REQ/AC 를 재개정하고, AC 수가 바뀌면 §F M1 의 스냅숏 규칙을 따른다.

1. **설계 선택지 — 혼합안(A 명시 선택자 + B 모호하면 거부 폴백, `sole_thread` 는 그대로).** Resolved — provisional default, to be confirmed at Kickoff. 규범 집합은 `spec.md` §D 의 REQ-CRS-001~010 전부다. A 단독·B 단독 매핑은 `spec.md` §C.5 에 역사적 근거로만 남는다.
2. **`thread_id` 허용 범위 — 이 프로젝트의 codex-jobs 레지스트리에 기록된 스레드만 허용하고, 그 밖의 id 는 거절한다.** Resolved — provisional default, to be confirmed at Kickoff. 반영: REQ-CRS-003(`thread_not_recorded` 거부, 레지스트리 밖 스레드 재개 금지). 기각된 대안은 임의 id 허용으로, 그 경우 `CODEX_HOME` 에 있는 다른 프로젝트의 스레드까지 재개할 수 있게 된다.
3. **거부 결과의 표면 — 구조화 JSON 결과만(`status: "failed"`, `error_code`, 모호 판정이면 `candidates`), `IsError` 는 켜지 않는다.** Resolved — provisional default, to be confirmed at Kickoff. 반영: REQ-CRS-010(모호 판정·미기록 `thread_id`·잘못된 `work_key` 세 거부 모두에 적용). 기각된 대안은 기존 두 분기(프롬프트 누락, 상태 디렉터리 쓰기 불가)처럼 `IsError: true` 를 함께 켜는 것이다.

## §C 사전 점검 (run 착수 전)

1. `git -C <worktree> rev-parse --short HEAD` 와 `git merge-base HEAD origin/develop` 로 기준을 기록한다.
2. `go test ./internal/cli -run 'TestCodexTask_ResumeLast|TestCodexTask_SandboxPolicyResetOnReusedThread' -count=1` 로 기존 세 테스트가 기준 트리에서 통과함을 기록한다(보존 기준선).
3. 탐침 `.moai/reports/t1201/premise-probe_test.go.txt` 를 RED 테스트로 옮겨 기준 트리에서 **실패**함을 관측한다(`thread/resume threadId` 가 `thr-6`).
4. Kickoff 에서 §B 세 결정이 확정됐는지 확인한다. 뒤집혔으면 run 을 시작하지 않고 SPEC 재개정을 먼저 요청한다.

## §D 제약

- fixture transport(`withCodexSession`, `codexTaskScript`, `sentParams`, `countSentMethod`)만 사용한다. LIVE codex 호출 0회. 요청 id 를 되돌려 주는 에코 fixture 는 테스트 도우미로 만든다(`acceptance.md` §A).
- 새 레코드 필드는 JSON `omitempty` 로 추가해 기존 레코드 파일과 읽기 호환을 유지한다(키 없는 과거 레코드는 `work_key` 없음으로 읽힌다).
- 모호 판정·미기록 `thread_id`·잘못된 `work_key` 경로에서는 codex 프로세스를 띄우지 않는다(`codexLookPath` 이후, `openCodexSessionOn` 이전에 판정).
- `resume_thread_id`·`resume_basis` 는 `thread/resume` 요청을 쓴 직후(응답 대기 전) 결과에 넣는다 — 재개 응답이 거부된 결과에도 남아야 하고, 요청을 보내기 전에 실패한 결과(세션 시작 실패, `initialize` 거부)에는 두 필드가 없어야 한다(REQ-CRS-001, AC-CRS-015).
- 전체 스위트를 로컬에서 돌리지 않는다. `./internal/cli` 만 재고, 전 패키지 판정은 CI 에 맡긴다.

## §E 자기 검증 (run 종료 시) 과 sync 체크리스트

- `go test ./internal/cli -run 'TestCodexTask_|TestCodexJob' -count=1` 전체 통과, 출력 원문 인용.
- `go vet ./internal/cli/...`, `golangci-lint run ./internal/cli/...` 무결.
- `acceptance.md` 의 AC 매트릭스 행마다 PASS/FAIL 과 해당 테스트 이름.
- sync 단계: SPEC-CODEX-PHASE2-001 의 HISTORY 에 "REQ-CX2-008 의 선택 절은 SPEC-CODEX-RESUME-SCOPE-001 이 좁힌다" 한 줄을 남긴다(본문 요구사항은 고치지 않는다, plan-audit D12).

## §F 마일스톤 (결정이 바뀔 가능성이 큰 순서)

### M1 — Kickoff 확정과 결과 모양 (우선순위 High)

- §B 세 잠정 결정을 운영자가 확정한다. 뒤집히면 `spec.md` §C.5 를 출발점으로 REQ/AC 를 재개정하고, AC 수가 스냅숏 기록 뒤 바뀌면 AC 스냅숏 재생성을 같은 커밋에 넣는다(`.moai/docs/ac-count-baseline-refresh.md` §2 네 번째 행).
- 결과 필드는 `spec.md` §D 가 정했다: `resume_thread_id`, `resume_basis`, `error_code`, `candidate_total`, `candidates[]`, `unused_selectors[]`. `CodexTaskResult` 에 `omitempty` 로 더한다.

### M2 — 레코드 모델: `work_key` 필드 (우선순위 High)

- `CodexJobRecord` 와 `codexJobSpec` 에 `work_key` 를 더하고 백그라운드 생성 경로에서 기록한다. `codex_job_status` 가 그대로 노출한다. `thread_id` 로 선택한 백그라운드 호출에서도 기록한다(REQ-CRS-005, 008).
- 입력 검증(REQ-CRS-005): 공백 제거 후 비어 있음, 128바이트 초과, 제어 문자 → `invalid_work_key` 거부.

### M3 — 선택 함수 교체 (우선순위 High)

- `latestThreadID()` 를 "선택자 → 후보 집합 → 판정" 구조로 바꾼다: `thread_id` 는 레지스트리 기록 여부 확인 후 선택(없으면 `thread_not_recorded`), `work_key` 는 그 키의 레코드 중 `updated_at` 최신(동률 시 레코드 id 사전순 최소), 선택자 없음이면 서로 다른 스레드 수로 분기(0 → 새 스레드, 1 → `sole_thread`, 2 이상 → `resume_ambiguous`).
- 후보는 스레드마다 대표 레코드(같은 `thread_id` 중 `updated_at` 최신, 동률 시 레코드 id 사전순 최소)로 집계하고, 대표 레코드의 `updated_at` 내림차순·레코드 id 오름차순으로 정렬해 앞 10개를 싣는다.
- 읽을 수 없는 레코드를 건너뛰는 기존 fail-open 은 유지한다.

### M4 — 도구 스키마와 설명 (우선순위 Medium)

- `mcp_server.go` 의 `codex_task` 등록부에 `thread_id`·`work_key` 입력을 추가하고 `resume_last` 설명을 REQ-CRS-009 에 맞춰 고친다.

### M5 — 테스트와 회귀 보존 (우선순위 Medium)

- AC-CRS-001~015 의 fixture 테스트를 `codex_task_test.go` 에 추가한다. 기존 `TestCodexTask_ResumeLastReusesRecordedThread`·`TestCodexTask_ResumeLastWithNoRecordedThread`·`TestCodexTask_SandboxPolicyResetOnReusedThread` 는 수정 없이 통과해야 한다.

## §G 위험

| 위험 | 영향 | 대응 |
|---|---|---|
| 레코드 누적으로 선택자 없는 `resume_last` 가 상시 모호 판정 | 단일 흐름 사용자의 체감 회귀(의도된 동작 변화, `spec.md` §C.3) | 거부 결과가 선택자 값을 담아 호출 한 번으로 복구된다. 레코드 수명 관리는 범위 밖(후속 카드) |
| 기존 레코드에 `work_key` 가 없음 | 과거 작업은 `work_key` 로 찾을 수 없음 | 설계상 수용. `thread_id` 선택자로 재개 가능 |
| `updated_at` 동률 | 선택·후보 순서 비결정 | 레코드 id 사전순으로 2차 정렬(REQ-CRS-004, 006) |
| 실제 codex 가 요청과 다른 스레드 id 를 돌려줌 | 호출자가 어느 스레드가 이어졌는지 오해 | `resume_thread_id`(보낸 id)와 `thread_id`(받은 id)를 따로 보고하고 `resumed_thread` 로 일치 여부를 알린다(REQ-CRS-001) |
| `request_summary` 가 후보 목록으로 다른 작업 항목에 노출 | 같은 프로젝트 안의 정보 노출 | 이미 레코드 생성 시 축약·마스킹된 값이며 `codex_job_status` 로도 읽을 수 있는 수준 |

## §H 교차 참조

- SPEC-CODEX-PHASE2-001 REQ-CX2-003(레코드 생성), REQ-CX2-007(sticky `sandboxPolicy` — 재개 경로 변경이 매 턴 명시 전송을 깨지 않아야 함, AC-CRS-005 가 보존 테스트로 지킴), REQ-CX2-008(`resume_last`).
- `.claude/rules/moai/core/moai-mcp-tools.md` § `project_root` 입력 — `codex_task` 는 그 9개 도구 목록에 없다(`spec.md` §A.3).

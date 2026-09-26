---
id: SPEC-CODEX-RESUME-SCOPE-001
document: acceptance
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
card: t1201
---

# Acceptance — SPEC-CODEX-RESUME-SCOPE-001

## §A 판정 규칙

- 모든 기준은 `internal/cli` 패키지의 fixture transport 테스트로 판정한다(`withCodexProjectDir`, `newCodexJobRegistry(...).create`, `withCodexSession`, `codexTaskScript`, `callCodexTask`, `structuredMap`, `sentParams`, `countSentMethod`). LIVE codex 호출은 0회다.
- **fixture 두 가지.** 측정(`spec.md` §A.4, `.moai/reports/t1201/fixture-probe.txt`): 기본 `codexTaskScript` 는 `thread/resume` 에 늘 `tid-fake` 를 돌려주므로 `turn/start` 의 `threadId` 와 결과의 `thread_id` 가 `tid-fake` 가 된다. 이를 아래처럼 나눠 쓴다.
  - **기본 fixture** — `codexTaskScript` 그대로. 요청 id 와 codex 응답 id 가 다른 경우다. 이 fixture 로 판정하는 기준은 **송신된** `thread/resume` 의 `threadId` 와 `resume_thread_id`·`resume_basis` 만 단언하고, `turn/start` 의 `threadId` 나 결과의 `thread_id` 를 요청 id 로 단언하지 않는다.
  - **에코 fixture** — `codexTaskScript` 의 모든 `tid-fake` 를 요청한 스레드 id 로 바꾼 변형(run 단계에서 스레드 id 를 인자로 받는 테스트 도우미로 만든다). 요청 id 와 응답 id 가 같은 경우이며, `turn/start` 의 `threadId`·결과의 `thread_id`·`resumed_thread: true` 를 단언할 수 있다. 측정상 이 변형에서 `turn/start="thr-5" result.thread_id=thr-5 resumed_thread=true` 가 관측됐다.
- "codex 프로세스를 띄우지 않는다"는 fixture 세션의 송신 기록(`sess.sent`)에 메시지가 0개임으로 판정한다.
- "거부 결과"는 REQ-CRS-010 의 모양이다: 결과의 `status` 가 `"failed"`, `error_code` 가 해당 코드, `CallToolResult.IsError` 가 false.
- 레코드를 심을 때 같은 `updated_at` 이 생기지 않도록 생성 사이에 간격을 두거나, 동률 정렬을 판정하는 기준에서는 동률을 의도적으로 만든다.
- 설계는 혼합안으로 잠정 확정됐으므로(`plan.md` §B) 아래 15개 기준 전부가 판정 대상이다.

### AC ↔ 요구사항 매핑

| AC | 요구사항 | fixture |
|---|---|---|
| AC-CRS-001 | REQ-CRS-002, 004 | 에코 |
| AC-CRS-002 | REQ-CRS-001 | 기본 (등록 id 가 `tid-fake` 라 요청=응답) |
| AC-CRS-003 | REQ-CRS-001, 003 | 에코 |
| AC-CRS-004 | REQ-CRS-003, 010 | 기본 (송신 0) |
| AC-CRS-005 | REQ-CRS-007 | 기존 테스트 |
| AC-CRS-006 | REQ-CRS-002, 006, 010 | 기본 (송신 0) |
| AC-CRS-007 | REQ-CRS-001, 007 | 기본 |
| AC-CRS-008 | REQ-CRS-004 | 기본 |
| AC-CRS-009 | REQ-CRS-005 | 기본 |
| AC-CRS-010 | REQ-CRS-005, 010 | 기본 (송신 0) |
| AC-CRS-011 | REQ-CRS-006 | 기본 (송신 0) |
| AC-CRS-012 | REQ-CRS-005, 008 | 에코 |
| AC-CRS-013 | REQ-CRS-009 | 없음 (`tools/list`) |
| AC-CRS-014 | REQ-CRS-004, 006 | 기본 + 에코 |
| AC-CRS-015 | REQ-CRS-001 | 기본 + 스레드 ack 오류 변형 |

## §B 인수 기준

### AC-CRS-001 — 이슈 재현: work_key 로 자기 카드 스레드를 재개한다 (REQ-CRS-002, 004)

- Given 레지스트리에 `thr-5`(`work_key: "card-A"`)가 먼저, `thr-6`(`work_key: "card-B"`)이 나중에 기록돼 있고, 세션은 `thr-5` 에코 fixture 다
- When `{"prompt": "continue card A", "resume_last": true, "work_key": "card-A"}` 로 호출한다
- Then 송신된 `thread/resume` 의 `threadId` 는 `"thr-5"` 이고, `thread/resume` 송신은 1회, `thread/start` 송신은 0회이며, 결과의 `resume_thread_id` 는 `"thr-5"`, `resume_basis` 는 `"work_key"` 다

### AC-CRS-002 — 재개 근거가 결과에 실린다 (REQ-CRS-001)

- Given 레지스트리에 스레드 `tid-fake` 하나만 기록돼 있고, 세션은 기본 fixture 다
- When `{"prompt": "continue", "resume_last": true}` 로 호출한다
- Then 결과의 `resume_thread_id` 와 `thread_id` 는 모두 `"tid-fake"`, `resumed_thread` 는 true, `resume_basis` 는 `"sole_thread"` 다

### AC-CRS-003 — thread_id 로 정확히 그 스레드를 재개한다 (REQ-CRS-001, 003)

- Given 레지스트리에 `thr-5` 와 `thr-6` 이 기록돼 있고(`thr-6` 이 최신), 세션은 `thr-5` 에코 fixture 다
- When `{"prompt": "p", "thread_id": "thr-5"}` 로 호출한다
- Then 송신된 `thread/resume` 의 `threadId` 는 `"thr-5"`, `turn/start` 의 `threadId` 도 `"thr-5"`, `thread/start` 송신은 0회이며, 결과의 `resume_thread_id` 는 `"thr-5"`, `resume_basis` 는 `"thread_id"` 다

### AC-CRS-004 — 기록되지 않은 thread_id 는 거절한다 (REQ-CRS-003, 010)

- Given 레지스트리에 `thr-5` 만 기록돼 있다
- When `{"prompt": "p", "thread_id": "thr-999"}` 로 호출한다
- Then 결과는 `status: "failed"`, `error_code: "thread_not_recorded"` 인 거부 결과이고 `IsError` 는 false 이며, 결과에 `"thr-999"` 가 명시되고, fixture 송신 기록은 0개다

### AC-CRS-005 — 기존 단일 흐름 테스트가 수정 없이 통과한다 (REQ-CRS-007)

- Given run 단계가 끝난 트리
- When `go test ./internal/cli -run 'TestCodexTask_ResumeLastReusesRecordedThread|TestCodexTask_ResumeLastWithNoRecordedThread|TestCodexTask_SandboxPolicyResetOnReusedThread' -count=1` 을 실행한다
- Then 세 테스트가 모두 `PASS` 하고, `git diff <기준 SHA> -- internal/cli/codex_task_test.go` 에서 세 테스트 함수 본문에 변경 행이 없다

### AC-CRS-006 — 선택자 없는 resume_last 는 모호하면 거절한다 (REQ-CRS-002, 006, 010)

- Given 레지스트리에 `thr-5`(`request_summary: "card A"`)와 나중에 `thr-6`(`"card B"`)이 기록돼 있다 — §A.2 탐침과 같은 모양
- When `{"prompt": "continue card A", "resume_last": true}` 로 호출한다
- Then fixture 송신 기록은 0개이고(`thread/resume` 도 `thread/start` 도 없음), 결과는 `status: "failed"`, `error_code: "resume_ambiguous"` 인 거부 결과이며 `IsError` 는 false, `candidate_total` 은 2, `candidates` 는 `[thr-6, thr-5]` 순이고 각 후보에 `thread_id`·`request_summary`·`updated_at` 이 있다

### AC-CRS-007 — 같은 스레드를 가진 레코드 여럿은 하나로 센다 (REQ-CRS-001, 007)

- Given 레지스트리에 `thread_id` 가 모두 `thr-5` 인 레코드 3건이 기록돼 있고, 세션은 기본 fixture 다
- When `{"prompt": "p", "resume_last": true}` 로 호출한다
- Then 모호 판정은 일어나지 않고, 송신된 `thread/resume` 의 `threadId` 는 `"thr-5"`, 결과의 `resume_thread_id` 는 `"thr-5"`, `resume_basis` 는 `"sole_thread"` 다

### AC-CRS-008 — 일치하는 work_key 가 없으면 새 스레드를 열고 말한다 (REQ-CRS-004)

- Given 레지스트리에 `thr-6`(`work_key: "card-B"`)만 기록돼 있다
- When `{"prompt": "p", "resume_last": true, "work_key": "card-A"}` 로 호출한다
- Then `thread/resume` 송신은 0회, `thread/start` 송신은 1회이고, 결과에 `resume_thread_id`·`resume_basis` 가 없으며, 결과의 note 에 `card-A` 에 기록된 스레드가 없다는 문장이 있다

### AC-CRS-009 — 백그라운드 작업 레코드에 work_key 가 기록된다 (REQ-CRS-005)

- Given 빈 레지스트리
- When `{"prompt": "p", "background": true, "work_key": "card-A"}` 로 호출하고 반환된 job id 로 `codex_job_status` 를 호출한다
- Then 레코드 JSON 파일과 `codex_job_status` 결과 모두에 `work_key` 가 `"card-A"` 로 있다

### AC-CRS-010 — 잘못된 work_key 는 codex 를 띄우기 전에 거절한다 (REQ-CRS-005, 010)

- Given 빈 레지스트리
- When `work_key` 를 각각 `"   "`, 129바이트 문자열, `"card\nA"` 로 주어 세 번 호출한다
- Then 세 호출 모두 `status: "failed"`, `error_code: "invalid_work_key"` 인 거부 결과이고 `IsError` 는 false 이며, 각 호출의 fixture 송신 기록은 0개다

### AC-CRS-011 — 후보 목록은 10개로 제한되고 결정적으로 정렬된다 (REQ-CRS-006)

- Given 레지스트리에 서로 다른 스레드 12개가 레코드 한 건씩으로 기록돼 있고, 그중 두 레코드는 `updated_at` 이 같다
- When `{"prompt": "p", "resume_last": true}` 로 호출한다
- Then `candidate_total` 은 12, `candidates` 길이는 10, 목록은 `updated_at` 내림차순이며 동률인 두 후보는 레코드 id 사전순 오름차순으로 놓인다. 같은 호출을 두 번 하면 두 후보 목록이 같다

### AC-CRS-012 — thread_id 가 다른 선택자보다 우선하고, 선택에 쓰지 않은 work_key 도 기록된다 (REQ-CRS-005, 008)

- Given 레지스트리에 `thr-5`(`work_key: "card-A"`)와 `thr-6`(`work_key: "card-B"`)이 기록돼 있고, 세션은 `thr-6` 에코 fixture 다
- When `{"prompt": "p", "thread_id": "thr-6", "resume_last": true, "work_key": "card-A"}` 로 포그라운드 호출을 한 번, 같은 입력에 `"background": true` 를 더해 한 번 더 호출한다
- Then 두 호출 모두 송신된 `thread/resume` 의 `threadId` 는 `"thr-6"`, `resume_basis` 는 `"thread_id"`, `unused_selectors` 는 정확히 `resume_last` 와 `work_key` 두 항목이다. 백그라운드 호출이 만든 새 레코드의 `work_key` 는 `"card-A"` 다

### AC-CRS-013 — 도구 스키마가 범위 규칙을 선언한다 (REQ-CRS-009)

- Given run 단계가 끝난 트리의 MCP 서버
- When 테스트가 `tools/list` 의 `codex_task` 항목을 읽는다
- Then `inputSchema.properties` 에 `thread_id` 와 `work_key` 가 문자열 타입으로 있고, `thread_id` 설명은 이 프로젝트 레지스트리에 기록된 스레드로 한정된다고, `work_key` 설명은 `resume_last` 를 그 키의 레코드로 한정한다고 적혀 있으며, `resume_last` 설명에 스레드가 둘 이상일 때 선택자 없는 호출이 거절된다고 적혀 있고 "most recently recorded codex thread for this project" 라는 무조건 문구는 없다

### AC-CRS-014 — 한 스레드가 여러 레코드에 걸칠 때 대표 레코드 값으로 후보와 선택을 정한다 (REQ-CRS-004, 006)

- Given 레지스트리에 다음 순서로 레코드 4건이 기록돼 있다: `thr-5`(`work_key: "card-A"`, `request_summary: "A first"`), `thr-6`(`work_key: "card-B"`, `"B first"`), `thr-5`(`work_key: "card-C"`, `"A resumed"`), `thr-5`(`work_key` 없음, `"A again"`)
- When (1) `{"prompt": "p", "resume_last": true}` 로 호출하고, (2) 에코 fixture(`thr-5`)로 `{"prompt": "p", "resume_last": true, "work_key": "card-A"}` 를 호출하고, (3) 에코 fixture(`thr-5`)로 `{"prompt": "p", "resume_last": true, "work_key": "card-C"}` 를 호출한다
- Then (1) 은 `resume_ambiguous` 거부이며 `candidate_total` 은 2, `candidates` 는 `[thr-5, thr-6]` 순이고, `thr-5` 후보의 `request_summary` 는 `"A again"`·`updated_at` 은 네 번째 레코드의 값이며 `work_key` 키가 없고, `thr-6` 후보의 `work_key` 는 `"card-B"` 다. (2) 와 (3) 은 모두 송신된 `thread/resume` 의 `threadId` 가 `"thr-5"` 이고 `resume_basis` 는 `"work_key"` 다

### AC-CRS-015 — codex 가 다른 id 를 돌려주거나 재개가 실패해도 보낸 id 와 근거를 보고한다 (REQ-CRS-001)

- Given 레지스트리에 `thr-5` 하나만 기록돼 있다
- When (1) 기본 fixture 로 `{"prompt": "p", "resume_last": true}` 를 호출하고, (2) 기본 fixture 의 스레드 ack 줄을 `{"id":2,"error":{"code":-32600,"message":"rejected by fake"}}` 로 바꾼 변형으로 같은 호출을 한다
- Then (1) 의 결과는 `resume_thread_id: "thr-5"`, `resume_basis: "sole_thread"`, `thread_id: "tid-fake"`, `resumed_thread: false` 다. (2) 의 결과는 `status: "failed"` 이고 `resume_thread_id: "thr-5"`, `resume_basis: "sole_thread"` 를 담는다

## §C 품질 게이트와 완료 정의

- `go test ./internal/cli -run 'TestCodexTask_|TestCodexJob' -count=1` 통과(출력 원문을 `progress.md` §E.2 에 인용).
- `go vet ./internal/cli/...` 와 `golangci-lint run ./internal/cli/...` 무결.
- AC-CRS-001~015 전부 PASS. Kickoff 에서 운영자가 잠정 결정을 뒤집어 기준이 제외되면, 제외된 기준과 사유를 `progress.md` 에 기록한다.
- 이 파일의 AC 수가 AC 스냅숏에 기록된 뒤 바뀌면 `.moai/reports/t338/ac-count-baseline.txt` 를 같은 커밋에서 재생성한다(`.moai/docs/ac-count-baseline-refresh.md` §2 네 번째 행).

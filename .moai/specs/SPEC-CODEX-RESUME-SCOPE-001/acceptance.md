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
- "codex 프로세스를 띄우지 않는다"는 fixture 세션의 송신 기록(`sess.sent`)에 메시지가 0개임으로 판정한다.
- 레코드를 심을 때 같은 `updated_at` 이 생기지 않도록 생성 사이에 간격을 두거나, 동률 정렬을 판정하는 기준에서는 동률을 의도적으로 만든다.
- 선택지 확정(`plan.md` §B)에 따라 제외되는 기준은 매핑표의 "선택지" 열로 정해진다.

### AC ↔ 요구사항 매핑

| AC | 요구사항 | 선택지 |
|---|---|---|
| AC-CRS-001 | REQ-CRS-002, 004 | A |
| AC-CRS-002 | REQ-CRS-001 | 공통 |
| AC-CRS-003 | REQ-CRS-001, 003 | A |
| AC-CRS-004 | REQ-CRS-003 | A |
| AC-CRS-005 | REQ-CRS-007 | 공통 |
| AC-CRS-006 | REQ-CRS-002, 006 | B |
| AC-CRS-007 | REQ-CRS-001, 007 | 공통 |
| AC-CRS-008 | REQ-CRS-004 | A |
| AC-CRS-009 | REQ-CRS-005 | A |
| AC-CRS-010 | REQ-CRS-005 | A |
| AC-CRS-011 | REQ-CRS-006 | B |
| AC-CRS-012 | REQ-CRS-008 | A |
| AC-CRS-013 | REQ-CRS-009 | 공통 |

## §B 인수 기준

### AC-CRS-001 — 이슈 재현: work_key 로 자기 카드 스레드를 재개한다 (REQ-CRS-002, 004)

- Given 레지스트리에 `thr-5`(`work_key: "card-A"`)가 먼저, `thr-6`(`work_key: "card-B"`)이 나중에 기록돼 있다
- When `{"prompt": "continue card A", "resume_last": true, "work_key": "card-A"}` 로 호출한다
- Then 송신된 `thread/resume` 의 `threadId` 는 `"thr-5"` 이고, `thread/resume` 송신은 1회, `thread/start` 송신은 0회이며, 결과의 재개 근거는 `work_key` 다

### AC-CRS-002 — 재개 근거가 결과에 실린다 (REQ-CRS-001)

- Given 레지스트리에 스레드 `tid-fake` 하나만 기록돼 있다
- When `{"prompt": "continue", "resume_last": true}` 로 호출한다
- Then 결과의 재개된 스레드 id 는 `"tid-fake"` 이고 재개 근거는 `sole_thread` 다

### AC-CRS-003 — thread_id 로 정확히 그 스레드를 재개한다 (REQ-CRS-001, 003)

- Given 레지스트리에 `thr-5` 와 `thr-6` 이 기록돼 있다(`thr-6` 이 최신)
- When `{"prompt": "p", "thread_id": "thr-5"}` 로 호출한다
- Then 송신된 `thread/resume` 의 `threadId` 는 `"thr-5"`, `turn/start` 의 `threadId` 도 `"thr-5"`, `thread/start` 송신은 0회이며 재개 근거는 `thread_id` 다

### AC-CRS-004 — 기록되지 않은 thread_id 는 거절한다 (REQ-CRS-003)

- Given 레지스트리에 `thr-5` 만 기록돼 있다
- When `{"prompt": "p", "thread_id": "thr-999"}` 로 호출한다
- Then 결과는 구조화 결과이며 `"thr-999"` 를 명시하고, fixture 송신 기록은 0개이며, 레지스트리의 레코드 수는 호출 전과 같다

### AC-CRS-005 — 기존 단일 흐름 테스트가 수정 없이 통과한다 (REQ-CRS-007)

- Given run 단계가 끝난 트리
- When `go test ./internal/cli -run 'TestCodexTask_ResumeLastReusesRecordedThread|TestCodexTask_ResumeLastWithNoRecordedThread' -count=1` 을 실행한다
- Then 두 테스트가 모두 `PASS` 하고, `git diff <기준 SHA> -- internal/cli/codex_task_test.go` 에서 두 테스트 함수 본문에 변경 행이 없다

### AC-CRS-006 — 선택자 없는 resume_last 는 모호하면 거절한다 (REQ-CRS-002, 006)

- Given 레지스트리에 `thr-5`(`request_summary: "card A"`)와 나중에 `thr-6`(`"card B"`)이 기록돼 있다 — §A.2 탐침과 같은 모양
- When `{"prompt": "continue card A", "resume_last": true}` 로 호출한다
- Then fixture 송신 기록은 0개이고(`thread/resume` 도 `thread/start` 도 없음), 결과의 오류 코드는 `resume_ambiguous`, 후보 총수는 2, 후보 목록은 `[thr-6, thr-5]` 순이며 각 후보에 `thread_id`·`request_summary`·`updated_at` 이 있다

### AC-CRS-007 — 같은 스레드를 가진 레코드 여럿은 하나로 센다 (REQ-CRS-001, 007)

- Given 레지스트리에 `thread_id` 가 모두 `thr-5` 인 레코드 3건이 기록돼 있다
- When `{"prompt": "p", "resume_last": true}` 로 호출한다
- Then 모호 판정은 일어나지 않고, 송신된 `thread/resume` 의 `threadId` 는 `"thr-5"`, 재개 근거는 `sole_thread` 다

### AC-CRS-008 — 일치하는 work_key 가 없으면 새 스레드를 열고 말한다 (REQ-CRS-004)

- Given 레지스트리에 `thr-6`(`work_key: "card-B"`)만 기록돼 있다
- When `{"prompt": "p", "resume_last": true, "work_key": "card-A"}` 로 호출한다
- Then `thread/resume` 송신은 0회, `thread/start` 송신은 1회이고, 결과의 note 에 `card-A` 에 기록된 스레드가 없다는 문장이 있다

### AC-CRS-009 — 백그라운드 작업 레코드에 work_key 가 기록된다 (REQ-CRS-005)

- Given 빈 레지스트리
- When `{"prompt": "p", "background": true, "work_key": "card-A"}` 로 호출하고 반환된 job id 로 `codex_job_status` 를 호출한다
- Then 레코드 JSON 파일과 `codex_job_status` 결과 모두에 `work_key` 가 `"card-A"` 로 있다

### AC-CRS-010 — 잘못된 work_key 는 codex 를 띄우기 전에 거절한다 (REQ-CRS-005)

- Given 빈 레지스트리
- When `work_key` 를 각각 `"   "`, 129바이트 문자열, `"card\nA"` 로 주어 세 번 호출한다
- Then 세 호출 모두 구조화 거절 결과이고, 각 호출의 fixture 송신 기록은 0개다

### AC-CRS-011 — 후보 목록은 10개로 제한되고 결정적으로 정렬된다 (REQ-CRS-006)

- Given 레지스트리에 서로 다른 스레드 12개가 기록돼 있고, 그중 두 레코드는 `updated_at` 이 같다
- When `{"prompt": "p", "resume_last": true}` 로 호출한다
- Then 후보 총수는 12, 후보 목록 길이는 10, 목록은 `updated_at` 내림차순이며 동률인 두 후보는 레코드 id 오름차순으로 놓인다. 같은 호출을 두 번 하면 두 후보 목록이 같다

### AC-CRS-012 — thread_id 가 다른 선택자보다 우선한다 (REQ-CRS-008)

- Given 레지스트리에 `thr-5`(`work_key: "card-A"`)와 `thr-6`(`work_key: "card-B"`)이 기록돼 있다
- When `{"prompt": "p", "thread_id": "thr-6", "resume_last": true, "work_key": "card-A"}` 로 호출한다
- Then 송신된 `thread/resume` 의 `threadId` 는 `"thr-6"` 이고, 결과는 `resume_last` 와 `work_key` 가 사용되지 않았다고 명시한다

### AC-CRS-013 — 도구 스키마가 범위 규칙을 선언한다 (REQ-CRS-009)

- Given run 단계가 끝난 트리의 MCP 서버
- When 테스트가 `tools/list` 의 `codex_task` 항목을 읽는다
- Then `inputSchema.properties` 에 채택된 선택자(`thread_id`, `work_key`)가 문자열 타입으로 있고, 각 설명이 비어 있지 않으며, `resume_last` 설명에 스레드가 둘 이상일 때의 동작이 적혀 있고 "most recently recorded codex thread for this project" 라는 무조건 문구는 없다

## §C 품질 게이트와 완료 정의

- `go test ./internal/cli -run 'TestCodexTask_|TestCodexJob' -count=1` 통과(출력 원문을 `progress.md` §E.2 에 인용).
- `go vet ./internal/cli/...` 와 `golangci-lint run ./internal/cli/...` 무결.
- 선택지 확정에 따라 남은 AC 전부 PASS. 제외된 AC 는 `progress.md` 에 제외 사유(운영자 결정)와 함께 기록.
- AC 수가 plan 단계와 달라지면 `.moai/reports/t338/ac-count-baseline.txt` 를 같은 커밋에서 재생성한다.

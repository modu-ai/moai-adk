---
id: SPEC-CODEX-RESUME-SCOPE-001
title: "codex_task resume_last must not resume another work item's thread — explicit selector plus refuse-on-ambiguity"
version: "0.2.0"
status: in-progress
created: 2026-09-26
updated: 2026-09-26
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "codex, mcp, codex_task, resume_last, thread, job-registry, worktree, kanban"
tier: M
card: t1201
issue_number: 1715
related_specs:
  - SPEC-CODEX-PHASE2-001  # owns REQ-CX2-008 (resume_last); this SPEC narrows its selection clause
---

# SPEC-CODEX-RESUME-SCOPE-001

## HISTORY

| 버전 | 날짜 | 내용 |
|---|---|---|
| 0.1.0 | 2026-09-26 | 카드 t1201 / 이슈 #1715 plan 초안. 전제는 fixture transport 탐침으로 이번 실행에서 측정했다(§A.2). 설계 선택지 A(명시 선택자)와 B(모호하면 거부)를 비교하고 A+B 혼합안을 **권고**했다(§C). |
| 0.2.0 | 2026-09-26 | plan-audit iter-1 FAIL(`.moai/reports/t1201/plan-audit.md`) 수리. 설계를 혼합안으로 **잠정 확정**했다(운영자가 질문 창 안에 답하지 않아 권고 기본값을 채택했고, Implementation Kickoff 에서 운영자가 확정하거나 뒤집는다 — `plan.md` §B). 선택지 표시(`[공통]`/`[A]`/`[B]`)를 요구사항에서 걷어내고 선택지별 유래는 §C.5 의 역사적 근거 표로 옮겼다. REQ-CRS-002 선택 근거에 REQ-CRS-007 을 넣고(D2), 결과 필드 `resume_thread_id`·`resume_basis` 의 의미와 codex 가 다른 id 를 돌려준 경우를 정했으며(D5), 한 스레드가 여러 레코드에 걸칠 때의 대표 레코드 규칙을 정했다(D6). REQ-CRS-008 을 "선택에 사용되지 않은 선택자" 로 좁히고(D8), 거부 결과의 표면을 REQ-CRS-010 으로 분리했다. |

## §A 배경과 전제

### A.1 현재 동작

`codex_task` 에 `resume_last: true` 를 주면 `codexJobRegistry.latestThreadID()`(`internal/cli/codex_jobs.go` 의 `latestThreadID` 함수)가 반환하는 스레드를 재개한다. 이 함수는 레지스트리 디렉터리(`<projectDir>/.moai/state/codex-jobs/`)의 모든 레코드 중 `thread_id` 가 비어 있지 않고 `updated_at` 이 가장 늦은 것 하나를 고른다. 호출자가 어떤 작업 항목을 이어 가려는지는 입력 어디에도 없으므로, 선택 기준은 **프로젝트 전체에서의 최신성** 하나뿐이다. 호출 지점은 `internal/cli/codex_task.go` 의 `handleCodexTask`(resume 분기), 도구 스키마는 `internal/cli/mcp_server.go` 의 `codex_task` 등록부다. 소유 요구사항은 SPEC-CODEX-PHASE2-001 의 REQ-CX2-008 이다:

> **While** `resume_last` is set, `codex_task` shall reuse the most recently recorded `threadId` for the project instead of opening a new thread; …

이 문언은 단일 작업 흐름을 전제한다. 병렬 작업 항목(칸반 카드 두 장이 각자 워크트리에서 같은 레지스트리를 공유하는 경우)에서는 카드 A를 이어 가려던 호출이 카드 B의 스레드를 재개한다(§A.2 에서 측정). 실제 codex 가 요청한 스레드를 그대로 돌려주면 그 결과가 `resumed_thread: true` 인 성공으로 보고된다는 점은 **코드상 추론**이다 — `ResumedThread = resumeThreadID != "" && session.threadID == resumeThreadID`(`codex_task.go`)에서 나온 것이고, 기본 fixture 는 어떤 요청에든 `tid-fake` 로 답해 성공 보고를 재현하지 못한다. 요청 id 를 되돌려 주는 fixture 변형에서는 `resumed_thread=true` 가 관측됐다(§A.4).

### A.2 측정된 전제 (이번 실행, fixture transport)

- 탐침 소스: `.moai/reports/t1201/premise-probe_test.go.txt` (패키지 `cli` 테스트로 실행)
- 관측 출력: `.moai/reports/t1201/premise-probe.txt`

설정: 레지스트리에 `thr-5`(`request_summary: "card A"`)를 먼저, 20ms 뒤 `thr-6`(`"card B"`)을 기록한 다음 `{"prompt": "continue card A", "resume_last": true}` 로 호출했다. 관측 출력(원문):

```
=== RUN   TestT1201Probe_ResumeLastPicksOtherWorkThread
    codex_task_t1201_probe_test.go:25: OBSERVED resumed_thread=false thread_id=tid-fake thread/resume threadId="thr-6" note=%!q(<nil>)
--- PASS: TestT1201Probe_ResumeLastPicksOtherWorkThread (0.02s)
PASS
```

codex 로 보낸 `thread/resume` 의 `threadId` 는 `"thr-6"` — 카드 B의 스레드다. 오선택을 알리는 note 는 없다(`<nil>`). `resumed_thread=false`, `thread_id=tid-fake` 는 fixture 스크립트가 어떤 요청에든 `tid-fake` 로 답하기 때문에 생긴 값으로, 판정 대상이 아니다. 판정 대상은 **송신된** `threadId` 다. `.moai/reports/` 는 gitignore 대상이라 두 파일은 커밋되지 않는다 — 경로는 primary checkout 기준이 아니라 이 워크트리 기준으로 존재한다.

### A.3 워크트리는 서로 다른 레지스트리를 갖는가 — 측정 결과: 갖지 않는다

`codex_task` 는 `projectDirResolver()`(= `resolveProjectDir`, `internal/cli/session.go`)로 레지스트리 위치를 정한다. 해석 순서는 ① 환경 변수 `CLAUDE_PROJECT_DIR`, ② MCP 서버 프로세스의 cwd 다. `codex_task` 에는 `spec_*` 도구와 달리 `project_root` 입력이 없다(`mcp_server.go` 등록부 확인).

이번 실행에서 이 세션에 도구를 제공하는 moai MCP 서버(서버 안내문이 밝힌 pid 82350)를 측정했다:

| 명령 | 관측 출력 |
|---|---|
| `lsof -a -p 82350 -d cwd -Fn` | `n/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1169` |
| `ps eww -p 82350 \| tr ' ' '\n' \| grep '^CLAUDE_PROJECT_DIR='` | `CLAUDE_PROJECT_DIR=/Users/goos/MoAI/moai-adk-go` |
| `ls /Users/goos/MoAI/moai-adk-go/.moai/state/codex-jobs` | `job-20260925T032604Z-4898c3e3.json` (1건) |

환경 변수가 우선하므로 해석 결과는 primary checkout 이고, 레지스트리는 `/Users/goos/MoAI/moai-adk-go/.moai/state/codex-jobs/` 하나다. 서버의 cwd 가 이 세션의 워크트리(`t1201`)가 아니라 **다른** 워크트리(`t1169`)라는 점도 관측됐다 — cwd 경로로 폴백하더라도 호출자의 워크트리를 가리킨다는 보장이 없다. 결론: 워크트리 분리는 레지스트리를 분리하지 않으며, 병렬 카드는 한 레지스트리를 공유한다. §A.2 의 오선택은 실제 운영 모양에서 성립한다.

측정의 한계: pid 82350 이 이 서브에이전트 세션과 다른 세션의 호출도 받는지는 재지 않았다. 결론(레지스트리가 워크트리마다 갈라지지 않는다)은 해석 순서와 환경 변수 값만으로 성립하므로 그 한계에 의존하지 않는다.

### A.4 fixture 가 돌려주는 스레드 id (이번 실행, plan-audit D4 대응)

- 관측 출력: `.moai/reports/t1201/fixture-probe.txt` (overlay 로 트리 무변경 실행)

레지스트리에 `thr-5` 한 건을 두고 `{"prompt": "p", "resume_last": true}` 로 호출했다. 관측 출력(원문):

```
OBSERVED stock methods=[initialize|thread/resume|turn/start] thread/resume="thr-5" turn/start="tid-fake" result.thread_id=tid-fake resumed_thread=false status=completed
OBSERVED echo-variant methods=[initialize|thread/resume|turn/start] thread/resume="thr-5" turn/start="thr-5" result.thread_id=thr-5 resumed_thread=true status=completed
```

`turn/start` 의 `threadId` 와 결과의 `thread_id` 는 **codex 가 `thread/resume` 에 돌려준 id** 를 따른다. 기본 `codexTaskScript` 는 늘 `tid-fake` 를 돌려주므로 "요청한 id 와 codex 가 돌려준 id 가 다른" 경우를, `tid-fake` 를 요청 id 로 바꾼 변형("echo-variant")은 "같은" 경우를 재현한다. 인수 기준은 이 둘을 구분해 쓴다(`acceptance.md` §A).

## §B 목표

`resume_last` 가 호출자가 의도하지 않은 작업 항목의 스레드를 **조용히** 재개하는 경로를 없앤다. 호출자는 어느 스레드를 이어 갈지 지정할 수 있어야 하고, 지정하지 않았는데 후보가 여럿이면 도구가 추측하지 않고 후보를 보여 주며 거절해야 한다. 스레드가 하나뿐인 기존 단일 흐름 사용은 바뀌지 않는다.

## §C 설계 선택지 비교와 결정

### C.1 선택지 A — 명시 선택자

`codex_task` 에 선택 입력 두 개를 더한다.

- `thread_id`: 재개할 스레드를 정확히 지정한다. 이 프로젝트 레지스트리에 기록된 스레드만 허용한다.
- `work_key`: 호출자가 정한 작업 키(예: 카드 id). 백그라운드 작업 레코드에 기록되고, `resume_last` 와 함께 주면 그 키를 가진 레코드 중 최신 스레드만 재개 후보가 된다.

장점: 칸반 레인처럼 자기 카드 id 를 아는 호출자는 한 번의 호출로 정확히 자기 스레드를 이어 간다. 선택이 입력으로 드러나 재현·감사가 쉽다.
단점: A **단독**이면 선택자를 주지 않는 기존 호출자(`resume_last` 만 쓰는 경우)는 여전히 최신성 선택(REQ-CX2-008)에 노출된다 — 이슈의 오선택 경로가 **닫히지 않고** 회피 수단만 생긴다. 이 때문에 A 단독은 채택하지 않았다(§C.3). 스키마가 넓어지고, 키를 기록하지 않은 과거 레코드는 `work_key` 로 찾을 수 없다.

### C.2 선택지 B — 모호하면 거부

선택자 없이 `resume_last` 를 받았을 때, 레지스트리에 서로 다른 스레드가 둘 이상 기록돼 있으면 재개도 새 스레드 개설도 하지 않고 후보 목록(`thread_id`, `request_summary`, `updated_at`)을 담은 구조화 오류를 반환한다.

장점: 오선택 경로 자체가 닫힌다. 스키마 변경이 없다(결과 필드만 는다). 구현이 작다.
단점: 레코드는 정리되지 않고 쌓이므로(REQ-CX2-003 이 백그라운드 작업마다 레코드를 만든다), 백그라운드 작업을 두 번 이상 돌린 프로젝트에서는 `resume_last` 가 **항상** 모호 판정으로 떨어진다 — 단일 흐름 사용자도 매번 거절당한다. 거절된 호출자가 후보 중 하나를 고를 수단이 없으므로(B 단독이면 입력이 없다), 거절 후 할 수 있는 일은 새 스레드를 여는 것뿐이다.

### C.3 혼합안 — A + B(모호성 폴백) [잠정 확정]

선택 순서: `thread_id` → `work_key` → 선택자 없음. 선택자가 없고 서로 다른 스레드가 둘 이상이면 B 의 거부를 적용하고, 거부 결과의 후보 목록에 각 후보의 `thread_id` 와 `work_key` 를 실어 호출자가 다음 호출에서 A 의 선택자로 바로 고를 수 있게 한다. 서로 다른 스레드가 정확히 하나면 지금처럼 재개하고(기존 테스트 보존), 없으면 지금처럼 새 스레드를 열고 그렇다고 말한다.

**근거.** A 단독은 이슈가 보고한 조용한 오선택을 선택자를 쓰지 않는 호출자에게 그대로 남기고, B 단독은 오선택은 막지만 거절된 호출자에게 고를 수단을 주지 않아 레코드가 둘만 쌓여도 `resume_last` 를 사실상 쓸 수 없게 만든다. 혼합안에서는 B 가 "추측하지 않는다"를 보장하고 A 가 "그럼 무엇을 고르나"에 답한다. 거부 결과가 다음 호출에 넣을 선택자 값을 그대로 담으므로 복구는 호출 한 번이다.

**호환성은 두 가지로 나눠 적는다.**

- 스키마 호환: 새 입력(`thread_id`, `work_key`)은 모두 선택 입력이고 새 결과 필드는 추가만 되므로, 기존 호출 모양은 그대로 유효하다.
- 의도된 동작 변화: 서로 다른 스레드가 둘 이상 기록된 레지스트리에서 선택자 없이 `resume_last` 를 쓰던 호출자는, 지금까지는 최신 스레드가 재개됐지만 이제 `resume_ambiguous` 거절을 받는다. 스레드가 0개·1개인 레지스트리의 동작은 바뀌지 않는다(REQ-CRS-007).

**결정 상태.** 운영자가 질문 창 안에 답하지 않아 권고 기본값을 채택했다. 이 결정과 `thread_id` 허용 범위·거부 결과 표면은 모두 잠정이며 Implementation Kickoff 에서 운영자가 확정하거나 뒤집는다(`plan.md` §B).

### C.4 검토했으나 이 SPEC 의 선택지로 두지 않은 대안 — 레지스트리를 워크트리별로 나누기

`codex_task` 에 `project_root` 입력을 더해 레지스트리를 호출자 워크트리 아래에 두면, 카드 하나당 워크트리 하나인 칸반 모양에서는 스레드가 자연히 분리된다. 그러나 (1) 같은 트리 안의 두 작업 항목은 여전히 구분하지 못하고, (2) 입력을 생략한 호출자는 §A.3 의 공유 레지스트리로 돌아가며, (3) 같은 값이 codex 턴의 `cwd`(현재 `projectDir` = primary checkout)까지 바꾸므로 작업 트리 위치라는 다른 축의 변경을 끌고 들어온다. 이 SPEC 은 선택 규칙만 다루고, 레지스트리·codex cwd 의 워크트리 한정은 범위 밖으로 둔다(§F).

### C.5 선택지별 유래 — 역사적 근거 표 (규범 아님)

아래 표는 v0.1.0 에서 선택지 확정 전에 요구사항을 어느 선택지에 걸었는지를 남긴 것이다. **규범 요구사항은 §D 의 혼합안 집합 전체**이며, 이 표로 요구사항을 빼지 않는다. 운영자가 Kickoff 에서 혼합안을 뒤집는 경우에만 재개정의 출발점으로 쓴다.

| 요구사항 | 유래 | A 단독이었다면 | B 단독이었다면 |
|---|---|---|---|
| REQ-CRS-001 | 공통 | 유지 (`resume_basis` 값에서 `sole_thread` 는 그대로) | 유지 (`resume_basis` 값은 `sole_thread` 뿐) |
| REQ-CRS-002 | 공통 + B | 최신성 금지 절이 빠지고 "선택자 없음 + 스레드 둘 이상" 에서 REQ-CX2-008 의 최신성 선택을 유지한다는 절이 필요했다(plan-audit D3) | 선택 근거 목록이 REQ-CRS-007 하나로 준다 |
| REQ-CRS-003·004·005·008 | A | 유지 | 제외 |
| REQ-CRS-006 | B | 제외 | 유지 (후보의 `work_key` 는 빠진다) |
| REQ-CRS-007·009 | 공통 | 유지 | 유지 |
| REQ-CRS-010 | 공통 (v0.2.0 신설) | REQ-CRS-006 참조 없이 유지 | REQ-CRS-003·005 참조 없이 유지 |

## §D 요구사항 (GEARS)

결과 필드 이름(`resume_thread_id`, `resume_basis`, `error_code`, `candidate_total`, `candidates`, `unused_selectors`)은 호출자가 읽는 관측 계약이므로 여기서 정한다. 기존 필드 `thread_id`(codex 가 돌려준 스레드 id)와 `resumed_thread` 의 의미는 바꾸지 않는다.

**용어 — 스레드의 대표 레코드.** 레지스트리에서 같은 `thread_id` 를 가진 레코드 집합 가운데 `updated_at` 이 가장 늦은 레코드이며, `updated_at` 이 같으면 레코드 id 가 사전순으로 가장 작은 레코드다. 재개된 스레드는 작업마다 새 레코드를 남기므로(`codex_task.go` 의 백그라운드 `registry.create`) 한 스레드가 여러 레코드에 걸치는 것은 정상 모양이다. 아래 요구사항에서 스레드의 `request_summary`·`updated_at`·`work_key` 는 모두 대표 레코드의 값이다.

### REQ-CRS-001 — The resume outcome names the sent thread and its basis

**When** `codex_task` sends `thread/resume`, the tool shall report in the result of that call `resume_thread_id`, set to the thread id it sent, and `resume_basis`, set to one of `thread_id`, `work_key`, or `sole_thread`, regardless of the call's final status. **When** the thread id codex returns differs from the one sent, the tool shall still report `resume_thread_id` and `resume_basis` as sent, shall report the returned id in `thread_id`, and shall set `resumed_thread` to false, as the existing result contract already defines. **When** `codex_task` does not send `thread/resume`, the result shall carry neither `resume_thread_id` nor `resume_basis`.

### REQ-CRS-002 — No resume of an unselected thread

The `codex_task` tool shall not send `thread/resume` for any thread other than the one selected by REQ-CRS-003, REQ-CRS-004, or REQ-CRS-007. Recency across the project registry shall not, on its own, select a thread when more than one distinct thread is recorded; that case is governed by REQ-CRS-006.

### REQ-CRS-003 — Explicit thread selection, limited to recorded threads

**When** the caller supplies `thread_id`, the `codex_task` tool shall resume exactly that thread with basis `thread_id`, provided at least one record in this project's job registry carries that `thread_id`. **When** the supplied `thread_id` is not carried by any record in this project's job registry, the tool shall neither resume a thread nor open a new one, shall start no codex process, and shall return a refusal (REQ-CRS-010) with error code `thread_not_recorded` that names the supplied `thread_id`. The tool shall not resume a thread that is recorded only outside this project's job registry.

### REQ-CRS-004 — Work-key scoped resume

**When** the caller supplies `work_key` together with `resume_last` and without `thread_id`, the `codex_task` tool shall consider only the records carrying that `work_key`, shall resume the `thread_id` of the record among them with the latest `updated_at` (ties broken by the lexicographically smallest record id), with basis `work_key`, and shall not consider records carrying a different `work_key` or none. **When** no record carries that `work_key`, the tool shall open a new thread and shall state in its result that no prior thread was recorded for that `work_key`.

### REQ-CRS-005 — The work key is recorded and bounded

**When** the caller supplies `work_key` to a background `codex_task`, the job record shall store it and `codex_job_status` shall return it, whichever selector decided the resumed thread. **When** a supplied `work_key` is empty after trimming, longer than 128 bytes, or contains a control character, the tool shall return a refusal (REQ-CRS-010) with error code `invalid_work_key` before starting any codex process.

### REQ-CRS-006 — Refuse on ambiguity

**When** `resume_last` is set without `thread_id` and without `work_key`, and the project's job registry records more than one distinct `thread_id`, the `codex_task` tool shall neither resume a thread nor open a new one, shall start no codex process, and shall return a refusal (REQ-CRS-010) with error code `resume_ambiguous`, `candidate_total` set to the number of distinct recorded threads, and `candidates` holding one entry per distinct thread. The candidates shall be ordered by their representative record's `updated_at` descending, ties broken by the lexicographically smallest representative record id first, and bounded to the first 10. Each candidate shall carry `thread_id` and its representative record's `request_summary` and `updated_at`, and shall carry that record's `work_key` when the representative record has one.

### REQ-CRS-007 — Single-thread and empty registries keep their behavior

**While** the project's job registry records exactly one distinct `thread_id`, a `resume_last` call without a selector shall resume that thread with basis `sole_thread`; **while** it records none, the call shall open a new thread and report that no prior thread was resumed, as REQ-CX2-008 already requires. Several records sharing one `thread_id` count as one distinct thread.

### REQ-CRS-008 — Selector precedence and unused-selector reporting

**When** `thread_id` is supplied together with `resume_last` or `work_key`, the `codex_task` tool shall select the thread by `thread_id` alone and shall list in `unused_selectors` every supplied selector input that did not decide the selection. A `work_key` listed there is unused for selection only: on a background call it is still recorded under REQ-CRS-005.

### REQ-CRS-009 — The published schema states the scoping

The `codex_task` tool schema shall declare `thread_id` and `work_key` with descriptions stating their scoping rules — `thread_id` limited to threads recorded in this project's job registry, `work_key` scoping `resume_last` to records carrying it — and the `resume_last` description shall state that a call without a selector is refused when more than one thread is recorded, replacing the unqualified "most recently recorded codex thread for this project" wording.

### REQ-CRS-010 — Refusals are structured results, not transport errors

The `codex_task` tool shall return every refusal defined by REQ-CRS-003, REQ-CRS-005, and REQ-CRS-006 as a structured JSON result with `status` set to `failed` and `error_code` set to the refusal's code, and shall not mark that result as an MCP tool error (`IsError` unset).

## §E 추적

| 요구사항 | 인수 기준 (acceptance.md) |
|---|---|
| REQ-CRS-001 | AC-CRS-002, AC-CRS-003, AC-CRS-007, AC-CRS-015 |
| REQ-CRS-002 | AC-CRS-001, AC-CRS-006 |
| REQ-CRS-003 | AC-CRS-003, AC-CRS-004 |
| REQ-CRS-004 | AC-CRS-001, AC-CRS-008, AC-CRS-014 |
| REQ-CRS-005 | AC-CRS-009, AC-CRS-010, AC-CRS-012 |
| REQ-CRS-006 | AC-CRS-006, AC-CRS-011, AC-CRS-014 |
| REQ-CRS-007 | AC-CRS-005, AC-CRS-007 |
| REQ-CRS-008 | AC-CRS-012 |
| REQ-CRS-009 | AC-CRS-013 |
| REQ-CRS-010 | AC-CRS-004, AC-CRS-006, AC-CRS-010 |

## §F 범위 밖

### Out of Scope — 레지스트리와 codex 작업 위치의 워크트리 한정

- `codex_task` 에 `project_root` 입력을 더하거나 레지스트리를 워크트리별로 나누는 일(§C.4). 후속 카드 후보로 남긴다.
- codex 턴의 `cwd` 가 호출자 워크트리가 아니라 `projectDirResolver()` 결과(측정상 primary checkout)라는 점. 쓰기 대상 트리의 문제이며 이 SPEC 의 선택 규칙과 독립이다.

### Out of Scope — 레지스트리 수명 관리

- 오래된 작업 레코드의 정리·만료·보존 정책. 모호 판정의 후보 수가 레코드 누적에 따라 늘어나는 것은 알려진 결과로 두고, 후보 목록 상한(10)으로만 다룬다.
- 포그라운드 `codex_task` 가 레코드를 남기지 않는 기존 동작(REQ-CX2-003). 포그라운드 작업의 스레드는 지금처럼 재개 후보가 아니다.

### Out of Scope — 다른 도구와 LIVE 검증

- `glm_task` 등 다른 작업 도구의 재개 의미.
- 실제 codex 바이너리에 대한 LIVE 호출. 모든 인수 기준은 fixture transport 테스트로 판정한다. 실제 codex 가 `thread/resume` 에 요청 id 를 그대로 돌려주는지는 이 SPEC 이 재지 않는다 — REQ-CRS-001 은 두 경우를 모두 정의한다.
- SPEC-CODEX-PHASE2-001 본문 수정. REQ-CX2-008 과의 관계는 이 SPEC 의 `related_specs` 와 sync 단계의 HISTORY 한 줄로 남긴다(`plan.md` §E).

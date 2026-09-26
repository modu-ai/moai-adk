---
id: SPEC-CODEX-RESUME-SCOPE-001
title: "codex_task resume_last must not resume another work item's thread — explicit selector plus refuse-on-ambiguity"
version: "0.1.0"
status: draft
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
| 0.1.0 | 2026-09-26 | 카드 t1201 / 이슈 #1715 plan 초안. 전제는 fixture transport 탐침으로 이번 실행에서 측정했다(§A.2). 설계 선택지 A(명시 선택자)와 B(모호하면 거부)를 비교하고 A+B 혼합안을 **권고**한다(§C). 선택지 확정은 Implementation Kickoff에서 운영자가 한다 — 요구사항마다 어느 선택지에 속하는지 표시해 두어, A만 또는 B만 고르더라도 빠지는 요구사항이 기계적으로 정해진다(§D). |

## §A 배경과 전제

### A.1 현재 동작

`codex_task` 에 `resume_last: true` 를 주면 `codexJobRegistry.latestThreadID()`(`internal/cli/codex_jobs.go` 의 `latestThreadID` 함수)가 반환하는 스레드를 재개한다. 이 함수는 레지스트리 디렉터리(`<projectDir>/.moai/state/codex-jobs/`)의 모든 레코드 중 `thread_id` 가 비어 있지 않고 `updated_at` 이 가장 늦은 것 하나를 고른다. 호출자가 어떤 작업 항목을 이어 가려는지는 입력 어디에도 없으므로, 선택 기준은 **프로젝트 전체에서의 최신성** 하나뿐이다. 호출 지점은 `internal/cli/codex_task.go` 의 `handleCodexTask`(resume 분기), 도구 스키마는 `internal/cli/mcp_server.go` 의 `codex_task` 등록부다. 소유 요구사항은 SPEC-CODEX-PHASE2-001 의 REQ-CX2-008 이다:

> **While** `resume_last` is set, `codex_task` shall reuse the most recently recorded `threadId` for the project instead of opening a new thread; …

이 문언은 단일 작업 흐름을 전제한다. 병렬 작업 항목(칸반 카드 두 장이 각자 워크트리에서 같은 레지스트리를 공유하는 경우)에서는 카드 A를 이어 가려던 호출이 카드 B의 스레드를 재개하고, 결과는 성공으로 보고된다.

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

## §B 목표

`resume_last` 가 호출자가 의도하지 않은 작업 항목의 스레드를 **조용히** 재개하는 경로를 없앤다. 호출자는 어느 스레드를 이어 갈지 지정할 수 있어야 하고, 지정하지 않았는데 후보가 여럿이면 도구가 추측하지 않고 후보를 보여 주며 거절해야 한다. 스레드가 하나뿐인 기존 단일 흐름 사용은 바뀌지 않는다.

## §C 설계 선택지 비교와 권고

### C.1 선택지 A — 명시 선택자

`codex_task` 에 선택 입력 두 개를 더한다.

- `thread_id`: 재개할 스레드를 정확히 지정한다. 이 프로젝트 레지스트리에 기록된 스레드만 허용한다.
- `work_key`: 호출자가 정한 작업 키(예: 카드 id). 백그라운드 작업 레코드에 기록되고, `resume_last` 와 함께 주면 그 키를 가진 레코드 중 최신 스레드만 재개 후보가 된다.

장점: 칸반 레인처럼 자기 카드 id 를 아는 호출자는 한 번의 호출로 정확히 자기 스레드를 이어 간다. 선택이 입력으로 드러나 재현·감사가 쉽다.
단점: 선택자를 주지 않는 기존 호출자(`resume_last` 만 쓰는 경우)는 여전히 최신성 선택에 노출된다 — A 단독으로는 이슈의 오선택 경로가 **닫히지 않고** 회피 수단만 생긴다. 스키마가 넓어지고, 키를 기록하지 않은 과거 레코드는 `work_key` 로 찾을 수 없다.

### C.2 선택지 B — 모호하면 거부

선택자 없이 `resume_last` 를 받았을 때, 레지스트리에 서로 다른 스레드가 둘 이상 기록돼 있으면 재개도 새 스레드 개설도 하지 않고 후보 목록(`thread_id`, `request_summary`, `updated_at`)을 담은 구조화 오류를 반환한다.

장점: 오선택 경로 자체가 닫힌다. 스키마 변경이 없다(결과 필드만 는다). 구현이 작다.
단점: 레코드는 정리되지 않고 쌓이므로(REQ-CX2-003 이 백그라운드 작업마다 레코드를 만든다), 백그라운드 작업을 두 번 이상 돌린 프로젝트에서는 `resume_last` 가 **항상** 모호 판정으로 떨어진다 — 단일 흐름 사용자도 매번 거절당한다. 거절된 호출자가 후보 중 하나를 고를 수단이 없으므로(B 단독이면 입력이 없다), 거절 후 할 수 있는 일은 새 스레드를 여는 것뿐이다.

### C.3 혼합안 — A + B(모호성 폴백) [권고]

선택 순서: `thread_id` → `work_key` → 선택자 없음. 선택자가 없고 서로 다른 스레드가 둘 이상이면 B 의 거부를 적용하고, 거부 결과의 후보 목록에 각 후보의 `thread_id` 와 `work_key` 를 실어 호출자가 다음 호출에서 A 의 선택자로 바로 고를 수 있게 한다. 서로 다른 스레드가 정확히 하나면 지금처럼 재개하고(기존 테스트 보존), 없으면 지금처럼 새 스레드를 열고 그렇다고 말한다.

**권고 근거.** A 단독은 이슈가 보고한 조용한 오선택을 선택자를 쓰지 않는 호출자에게 그대로 남기고, B 단독은 오선택은 막지만 거절된 호출자에게 고를 수단을 주지 않아 레코드가 둘만 쌓여도 `resume_last` 를 사실상 쓸 수 없게 만든다. 혼합안에서는 B 가 "추측하지 않는다"를 보장하고 A 가 "그럼 무엇을 고르나"에 답한다. 거부 결과가 다음 호출에 넣을 선택자 값을 그대로 담으므로 복구는 호출 한 번이다. 스레드가 하나뿐인 단일 흐름은 동작이 바뀌지 않는다. 비용은 스키마 필드 두 개와 레코드 필드 하나이며, 둘 다 선택 입력이라 기존 호출자는 깨지지 않는다.

이 권고는 결정이 아니다. 선택지 확정은 Implementation Kickoff 에서 운영자가 한다(`plan.md` §B 의 `[NEEDS CLARIFICATION]` 항목).

### C.4 검토했으나 이 SPEC 의 선택지로 두지 않은 대안 — 레지스트리를 워크트리별로 나누기

`codex_task` 에 `project_root` 입력을 더해 레지스트리를 호출자 워크트리 아래에 두면, 카드 하나당 워크트리 하나인 칸반 모양에서는 스레드가 자연히 분리된다. 그러나 (1) 같은 트리 안의 두 작업 항목은 여전히 구분하지 못하고, (2) 입력을 생략한 호출자는 §A.3 의 공유 레지스트리로 돌아가며, (3) 같은 값이 codex 턴의 `cwd`(현재 `projectDir` = primary checkout)까지 바꾸므로 작업 트리 위치라는 다른 축의 변경을 끌고 들어온다. 이 SPEC 은 선택 규칙만 다루고, 레지스트리·codex cwd 의 워크트리 한정은 범위 밖으로 둔다(§F).

## §D 요구사항 (GEARS)

각 요구사항 끝의 표시는 속한 선택지다. `[공통]` 은 어느 선택지든 남는다. A 만 채택하면 `[B]` 가, B 만 채택하면 `[A]` 가 빠진다.

### REQ-CRS-001 — The resume outcome names its basis `[공통]`

**When** `codex_task` resumes a recorded thread, the tool shall report in its result the resumed `thread_id` and the basis on which that thread was selected, as one of `thread_id`, `work_key`, or `sole_thread`.

### REQ-CRS-002 — No resume of an unselected thread `[공통]`

The `codex_task` tool shall not send `thread/resume` for any thread other than the one selected by REQ-CRS-003, REQ-CRS-004, or REQ-CRS-006; recency across the project registry shall not, on its own, select a thread when more than one distinct thread is recorded.

### REQ-CRS-003 — Explicit thread selection `[A]`

**Where** the caller supplies `thread_id`, the `codex_task` tool shall resume exactly that thread. **When** the supplied `thread_id` is not recorded in the project's job registry, the tool shall neither resume a thread nor open a new one, shall start no codex process, and shall return a structured result that names the unrecorded `thread_id`.

### REQ-CRS-004 — Work-key scoped resume `[A]`

**Where** the caller supplies `work_key` together with `resume_last`, the `codex_task` tool shall resume the most recently updated thread among the job records carrying that `work_key`, and shall not consider records carrying a different `work_key` or none. **When** no record carries that `work_key`, the tool shall open a new thread and shall state in its result that no prior thread was recorded for that `work_key`.

### REQ-CRS-005 — The work key is recorded and bounded `[A]`

**Where** the caller supplies `work_key` to a background `codex_task`, the job record shall store it and `codex_job_status` shall return it. **When** a supplied `work_key` is empty after trimming, longer than 128 bytes, or contains a control character, the tool shall reject the call with a structured result before starting any codex process.

### REQ-CRS-006 — Refuse on ambiguity `[B]`

**When** `resume_last` is set without `thread_id` and without `work_key`, and the project's job registry records more than one distinct `thread_id`, the `codex_task` tool shall neither resume a thread nor open a new one, shall start no codex process, and shall return a structured result carrying the error code `resume_ambiguous`, the total number of distinct candidate threads, and a candidate list ordered by `updated_at` descending (ties broken by record id), bounded to the 10 most recent, where each candidate carries `thread_id`, `request_summary`, `updated_at`, and `work_key` when recorded.

### REQ-CRS-007 — Single-thread and empty registries keep their behavior `[공통]`

**While** the project's job registry records exactly one distinct `thread_id`, a `resume_last` call without a selector shall resume that thread with basis `sole_thread`; **while** it records none, the call shall open a new thread and report that no prior thread was resumed, as REQ-CX2-008 already requires. Several records sharing one `thread_id` (a thread resumed across jobs) count as one distinct thread.

### REQ-CRS-008 — Selector precedence `[A]`

**When** `thread_id` is supplied together with `resume_last` or `work_key`, the `codex_task` tool shall select by `thread_id` alone and shall state in its result which supplied selectors were not used.

### REQ-CRS-009 — The published schema states the scoping `[공통]`

The `codex_task` tool schema shall declare every selector input this SPEC adopts with a description stating its scoping rule, and the `resume_last` description shall state what happens when more than one thread is recorded, replacing the unqualified "most recently recorded codex thread for this project" wording.

## §E 추적

| 요구사항 | 인수 기준 (acceptance.md) |
|---|---|
| REQ-CRS-001 | AC-CRS-002, AC-CRS-003, AC-CRS-007 |
| REQ-CRS-002 | AC-CRS-001, AC-CRS-006 |
| REQ-CRS-003 | AC-CRS-003, AC-CRS-004 |
| REQ-CRS-004 | AC-CRS-001, AC-CRS-008 |
| REQ-CRS-005 | AC-CRS-009, AC-CRS-010 |
| REQ-CRS-006 | AC-CRS-006, AC-CRS-011 |
| REQ-CRS-007 | AC-CRS-005, AC-CRS-007 |
| REQ-CRS-008 | AC-CRS-012 |
| REQ-CRS-009 | AC-CRS-013 |

## §F 범위 밖

### Out of Scope — 레지스트리와 codex 작업 위치의 워크트리 한정

- `codex_task` 에 `project_root` 입력을 더하거나 레지스트리를 워크트리별로 나누는 일(§C.4). 후속 카드 후보로 남긴다.
- codex 턴의 `cwd` 가 호출자 워크트리가 아니라 `projectDirResolver()` 결과(측정상 primary checkout)라는 점. 쓰기 대상 트리의 문제이며 이 SPEC 의 선택 규칙과 독립이다.

### Out of Scope — 레지스트리 수명 관리

- 오래된 작업 레코드의 정리·만료·보존 정책. 모호 판정의 후보 수가 레코드 누적에 따라 늘어나는 것은 알려진 결과로 두고, 후보 목록 상한(10)으로만 다룬다.
- 포그라운드 `codex_task` 가 레코드를 남기지 않는 기존 동작(REQ-CX2-003). 포그라운드 작업의 스레드는 지금처럼 `resume_last` 대상이 아니다.

### Out of Scope — 다른 도구와 LIVE 검증

- `glm_task` 등 다른 작업 도구의 재개 의미.
- 실제 codex 바이너리에 대한 LIVE 호출. 모든 인수 기준은 fixture transport 테스트로 판정한다.
- SPEC-CODEX-PHASE2-001 본문 수정. REQ-CX2-008 과의 관계는 이 SPEC 의 `related_specs` 와 sync 단계의 상호 참조로 남긴다.

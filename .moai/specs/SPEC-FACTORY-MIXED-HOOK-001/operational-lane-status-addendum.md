# 팩토리 운영 레인 상태 조회 계약 추가안

> 대상: `SPEC-FACTORY-MIXED-HOOK-001` / 카드 `t1074`
>
> 성격: 기존 구현 요청에 대한 한정 추가 계약
>
> 상태: 사용자 요청으로 추가된 normative 요구 — canonical `spec.md`, `plan.md`, `acceptance.md`가 이 문서의 OPS 요구·AC를 편입해 참조한다. 충돌 시 canonical 문서가 우선한다.

## 1. 문제와 관측 경계

사용자가 제공한 화면은 별도 프로젝트 `mo.ai.kr`에서 `moai codex -f` 리더와 `moai codex -f agent` 작업자 두 개를 실행한 뒤, 리더가 실제 작업자 레인의 상태를 조회하지 못한 사례다. 이 화면은 요구의 근거이며, 이 저장소에서 재실행한 검증 결과로 간주하지 않는다.

이번 문서 작성 시점의 코드 판독 기준선은 다음과 같다.

- `rg -n 'type Status|func \(s \*Store\) Status|factory_msg_status|factory_msg_list' internal/factorymsg/store.go internal/cli/mcp_factory_msg.go internal/cli/mcp_server.go`로 확인한 현재 `factory_msg_status`는 메시지 상태별 개수와 전달 경계만 반환한다.
- 같은 판독에서 `factory_msg_list`는 현재 세션 수신함의 메시지를 `Claim`하는 변경 작업이며, 레인 명부 조회가 아니다.
- `sed -n '250,375p' internal/factorymsg/store.go`로 확인한 `peers` 레지스트리는 안정적인 `slot`과 `project_key`, `run_id`, `backend`, `role`, `session_uuid`, `generation`, `pid`, `process_start`, `updated_at`을 이미 보유한다.
- `sed -n '1,80p' internal/hook/factory_messages.go`와 `sed -n '430,515p' internal/hook/session_start.go`로 확인한 등록 경로는 `SessionStart` 처리 중 `RegisterPeer`를 호출한다.

따라서 최소 구현은 새로운 명부 저장소를 만들지 않고 기존 `peers`와 기존 읽기 전용 `factory_msg_status` 표면을 확장하는 것이다.

## 2. 추가 운영 계약

### REQ-FMH-OPS-001 — canonical run 명부

팩토리 상태 조회가 호출되면 시스템은 호출 프로젝트의 canonical `project_key`와 명시된 active `run_id`가 모두 일치하는 현재 `peers` 행만 `lanes` 명부로 반환해야 한다. 다른 프로젝트나 다른 run의 행은 개수, 식별자, 상태 어느 형태로도 섞어 반환해서는 안 된다.

### REQ-FMH-OPS-002 — 논리 레인과 물리 세션 식별

각 명부 항목은 최소한 다음 값을 구조화된 필드로 반환해야 한다.

- 안정적인 논리 식별자: `slot`, `role`
- 실행 계열: `backend`
- 현재 물리 endpoint: `session_uuid`, `generation`
- 소유 프로세스: `pid`, 등록된 `process_start` fingerprint
- 관측 정보: 등록 갱신 시각, 상태 판정 시각

`slot`은 세션 재바인드 뒤에도 논리 레인 식별자로 유지되고, `session_uuid`와 `generation`은 현재 물리 endpoint를 나타내야 한다. 조회 구현이 이 값을 제목, 메시지 수, 또는 launcher의 별도 JSON 명부에서 재구성해서는 안 된다.

### REQ-FMH-OPS-003 — endpoint 생존 판정

시스템은 조회 시 등록된 `pid`를 실제 프로세스 식별 probe로 확인하고 등록된 `process_start`와 비교해야 한다. 반환하는 `endpoint_state`는 다음 의미를 가져야 한다.

| 값 | 판정 근거 |
|---|---|
| `live` | PID가 살아 있고 실제 fingerprint가 등록된 `process_start`와 정확히 일치한다. |
| `dead` | OS probe가 해당 PID의 부재를 확정한다. |
| `stale` | PID는 살아 있으나 fingerprint가 달라 PID 재사용 또는 이전 소유자의 바인딩임이 확정된다. |
| `unknown` | 권한, 플랫폼 또는 probe 오류 때문에 생존 여부를 확정할 수 없다. |

`dead`와 `stale`은 동일 상태로 접어서는 안 된다. 단순한 `updated_at` 경과만으로 장시간 살아 있는 세션을 `stale`로 판정해서도 안 된다. 조회 실패나 불확정 probe를 `dead`로 낮춰 표시해서는 안 된다.

### REQ-FMH-OPS-004 — 작업 상태의 진실성

각 명부 항목은 `task_state`를 `busy`, `idle`, `unknown` 중 하나로 반환해야 한다. `busy` 또는 `idle`은 명시적인 런타임 관측과 그 관측 시각·출처가 있을 때만 허용한다. 현재 레지스트리에 그런 관측이 없다면 반드시 `unknown`이어야 한다.

프로세스가 살아 있다는 사실, 미수신 메시지가 없다는 사실, 첫 프롬프트가 없다는 사실, 카드 배정 문자열만 존재한다는 사실은 각각 `idle`의 근거가 아니다. `endpoint_state`와 `task_state`는 별도 축으로 유지한다.

### REQ-FMH-OPS-005 — 읽기 전용 조회 표면

기존 MCP `factory_msg_status(run_id)`는 기존 메시지 집계를 유지하면서 `lanes` 명부를 함께 반환하는 canonical 운영 조회 표면이 되어야 한다. 이 호출은 메시지를 claim·acknowledge하거나 peer를 등록·갱신하거나 generation을 올리지 않아야 한다. `factory_msg_list`는 수신함 claim 용도 그대로이며 명부 조회 방법으로 안내해서는 안 된다.

별도 CLI를 추가하는 것은 필수가 아니다. 추후 CLI를 제공하더라도 동일한 저장소 조회와 동일한 상태 의미를 사용해야 하며 두 번째 상태 체계를 만들면 안 된다.

### REQ-FMH-OPS-006 — launcher 안내의 실존 표면 연결

팩토리 리더의 `SessionStart` 안내문은 해당 run의 레인 상태를 조회하는 실제 도구 이름과 필수 `run_id` 사용법을 알려야 한다. 안내문에 존재하지 않는 명령, 메시지를 claim하는 `factory_msg_list`, 향후 구현 예정 표면을 현재 기능처럼 적어서는 안 된다. 안내문 테스트는 번역 문자열 존재만이 아니라 등록된 MCP 도구와 handler가 실제 연결되어 있음을 함께 검증해야 한다.

### REQ-FMH-OPS-007 — 프롬프트 없는 등록

production 실증은 이번 tree에서 빌드한 동일 `moai` binary로 정확히 한 개의 `moai codex -f` 리더와 두 개의 `moai codex -f agent` 작업자 프로세스를 시작해야 한다. 세 세션은 첫 사용자/에이전트 프롬프트 전에 각 `SessionStart`가 전달한 실제 Codex owner PID와 실제 process-start fingerprint로 기존 `peers`에 등록되어야 하며, 이후 리더 세션의 MCP `factory_msg_status`가 바로 그 세 endpoint를 반환해야 한다. direct `codex`/`codex exec`, 외부 `MOAI_SESSION_PID` 주입, 직접 `RegisterPeer`, DB seed 또는 별도 owner 프로세스로 이 chain을 대체해서는 안 된다.

### REQ-FMH-OPS-008 — 실패와 불확정성 보존

레지스트리 조회 실패, active run 불일치, process probe 불확정은 정상 빈 명부로 위장해서는 안 된다. 구조화된 오류 또는 해당 항목의 `unknown`으로 보존해야 한다. 정렬은 `slot` 기준으로 결정적이어야 하며, 동일 canonical project/run에서 한 `slot`에는 현재 generation 하나만 보여야 한다.

## 3. 범위

### 포함

- 기존 run-scoped `peers`를 읽는 명부 API와 상태 자료형
- 기존 `factory_msg_status` 응답의 `lanes` 확장
- 실제 표면을 가리키는 팩토리 리더 안내
- 저장소 단위 계약, MCP 읽기 전용 계약, 실제 세션 통합 검증

### Out of Scope — 별도 스케줄러와 상태 추정

- 모델의 내부 추론, 토큰 생성 여부 또는 터미널 화면을 scraping하여 작업 상태를 추정하는 기능
- idle wake 또는 백그라운드 push 전달(t1075 범위)
- worktree 생성·`/cd`·세션 rebind lifecycle(t1082 범위)
- 신규 카드 발행, todo queue 수정, `mo.ai.kr` 프로젝트 수정
- 기존 `factory_msg_list`의 claim 의미 변경

## 4. 최소 구현 계획

### M1 — 기존 레지스트리의 읽기 모델

- `internal/factorymsg/store.go`: 기존 `Peer`/`peers`를 재사용하는 `LaneStatus` 읽기 모델과 결정적 명부 조회를 추가한다. `homestate.ProbeProcessIdentity` 결과와 fingerprint 비교를 한 곳에서 `live/dead/stale/unknown`으로 매핑한다.
- `internal/factorymsg/store_test.go`: project/run 격리, slot 정렬, generation 현재성, live/dead/stale/unknown 진실표, `task_state=unknown` 기본값을 검증한다.
- 조회를 위해 새 DB, launcher registry 또는 heartbeat 파일을 만들지 않는다. `updated_at`은 등록 시각으로만 표현하고 작업 활동 시각으로 오인하지 않는다.

### M2 — 읽기 전용 MCP 응답

- `internal/cli/mcp_factory_msg.go`: `handleFactoryMsgStatus`가 기존 메시지 집계와 `lanes`를 함께 반환하도록 한다.
- `internal/cli/mcp_server.go`: `factory_msg_status` 설명을 메시지 집계와 레인 명부 조회로 갱신하되 read-only annotation을 유지한다.
- `internal/cli/mcp_factory_msg_test.go`: 호출 전후 메시지 state/claim token, peer generation/`updated_at`, 행 수가 같음을 검증하여 무변경성을 입증한다. 다른 canonical project fixture의 session UUID가 응답에 나타나지 않는 것도 검증한다.

### M3 — SessionStart 등록과 안내 연결

- `internal/hook/factory_messages.go`: 기존 `registerFactoryHookPeer` 경로를 유지하고 프롬프트 없는 등록 계약을 고정한다.
- `internal/hook/session_start_factory.go`, `internal/hook/session_start_factory_i18n.go`: 리더 안내가 `factory_msg_status`와 현재 run ID를 사용하도록 한다.
- `internal/hook/factory_messages_test.go`, `internal/hook/session_start_factory_test.go`: `SessionStart` 입력만으로 세 레인이 등록되는지, 안내에 나온 tool이 MCP 등록표와 handler에 실제 존재하는지 검증한다.

### M4 — 설치 산출물로 별도 프로젝트 실증

- `internal/cli/factory_live_test.go`에 `TestFactoryLiveOperationalRosterBeforePrompt`와 `TestFactoryLiveOperationalLauncherChain`을 추가하고 cleanup-guaranteed 공통 production-launcher harness를 사용한다.
- 테스트는 저장소 밖 별도 임시 fixture 프로젝트를 만들고, 이번 tree에서 빌드·설치한 `moai` 산출물의 절대 경로와 SHA-256을 기록한 뒤 해당 binary를 가리키도록 MCP를 구성한다.
- harness는 그 binary의 production CLI를 통해 정확히 한 개의 `<built-moai> codex -f`와 두 개의 `<built-moai> codex -f agent`를 시작하고, 실제 process tree의 PID/PPID/argv를 기록해야 한다. direct `codex`/`codex exec`, 수동 `MOAI_SESSION_PID`, 직접 `RegisterPeer`, DB seed, 별도 sleep/owner 프로세스는 금지한다.
- 첫 모델 prompt를 보내지 않은 채 세 `SessionStart` 등록을 먼저 관측한다. 각 peer 행의 PID/process-start가 launcher가 만든 실제 Codex owner process를 OS에서 재측정한 값과 같아야 하며, 서로 다른 session UUID와 현재 generation을 가져야 한다.
- 프롬프트 전 등록 assertion이 끝난 뒤에만 리더 세션에 한 번 요청하여 그 리더가 새로 시작한 자기 MCP의 `factory_msg_status(run_id)`를 호출하게 한다. 반환값은 동일한 `lead`, `agent-1`, `agent-2` endpoint 세 개여야 하고 작업자에게는 계속 prompt를 보내지 않는다.
- 기존 MCP 프로세스를 재사용하지 않았음과 리더가 호출한 MCP의 PID/binary hash를 입증한다. 모든 자식 process tree는 테스트 cleanup에서 종료하고 잔존 PID가 없음을 확인한다.
- 각 항목의 backend/slot/session UUID/generation/PID/fingerprint를 대조한다. 두 작업자의 `task_state`는 근거가 없으므로 `unknown`이어야 한다.
- 별도의 두 번째 fixture 프로젝트에 의도적인 sentinel peer를 두고 첫 프로젝트 조회에 나타나지 않음을 검증한다.
- 최초 세 레인의 실세션 명부 assertion이 끝난 뒤 한 작업자 프로세스를 종료하여 `dead`를 검증한다. `stale`은 별도의 보조 행에 현재 살아 있는 실제 프로세스 PID와 의도적으로 불일치하는 등록 fingerprint를 넣고 실제 OS probe가 불일치를 발견하는 방식으로 검증한다. 이 보조 행은 최초 실세션 명부 assertion의 대체 증거가 아니며 가짜 process probe를 사용해서는 안 된다.

## 5. 이진 acceptance 계약

### AC-FMH-OPS-001 — 동일 run의 실세션 명부

Given 이번 tree의 built `moai`로 정확히 한 개의 `moai codex -f`와 두 개의 `moai codex -f agent`가 별도 fixture 프로젝트의 같은 active run에 시작되고 아직 어느 세션에도 prompt를 보내지 않았을 때, When 세 실제 Codex owner의 `SessionStart` 등록을 확인한 뒤 리더 세션의 MCP `factory_msg_status(run_id)`를 호출하면, Then `lanes`는 `lead`, `agent-1`, `agent-2` 세 항목만 반환하고 각 backend/session UUID/generation/PID/process-start가 실제 process tree와 일치한다.

### AC-FMH-OPS-002 — 상태 축과 불확정성

Given live/dead/PID-fingerprint 불일치/indeterminate endpoint가 준비되었을 때, When 명부를 조회하면, Then `endpoint_state`는 각각 `live/dead/stale/unknown`이며 작업 관측이 없는 모든 항목의 `task_state`는 `unknown`이다.

### AC-FMH-OPS-003 — 프로젝트 격리

Given 서로 다른 canonical project가 같은 run 문자열과 겹치는 slot 이름을 사용할 때, When 첫 프로젝트에서 조회하면, Then 두 번째 프로젝트의 session UUID, PID, 집계 또는 상태는 응답에 하나도 포함되지 않는다.

### AC-FMH-OPS-004 — 무변경 조회

Given 메시지와 peer 상태의 전후 snapshot이 가능할 때, When `factory_msg_status`를 반복 호출하면, Then 메시지 state·claim token·receipt, peer generation·`updated_at`, 모든 관련 행 수가 호출 전과 동일하다.

### AC-FMH-OPS-005 — 안내의 실행 가능성

Given 팩토리 리더 `SessionStart` 안내가 생성될 때, When 안내가 제시한 상태 조회 표면을 같은 run에서 호출하면, Then 등록된 MCP handler가 응답하며 `factory_msg_list`를 호출하거나 메시지를 claim하지 않는다.

### AC-FMH-OPS-006 — 설치 및 재시작 뒤 실제 통합 증거

Given 이번 tree의 설치 binary hash가 고정되고 기존 MCP가 종료되었을 때, When production argv인 한 개의 `moai codex -f`와 두 개의 `moai codex -f agent`만으로 세션을 시작하여 프롬프트 전 SessionStart owner 등록 후 리더 자신의 새 MCP에서 조회하면, Then 세 launcher argv/process tree, MCP PID/hash, 세 session UUID, 세 실제 owner PID/fingerprint, 동일 endpoint 조회, 다른 프로젝트 격리 및 전체 child cleanup이 실제 출력에 남고 테스트가 통과한다.

`t.Skip`, `NOT_RUN`, direct `codex`/`codex exec`, 외부 `MOAI_SESSION_PID`, 직접 `RegisterPeer`, 사전 seed한 `Peer`, 별도 owner process, 가짜 process probe, in-process handler만 호출한 fixture, 또는 목업 명부는 AC-FMH-OPS-001/005/006의 PASS 증거가 될 수 없다.

## 6. 실행 게이트

### 6.1 단위·MCP·hook 계약

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-lane-status-home GOCACHE=/tmp/t1074-lane-status-cache go test -json ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$' -count=1 -timeout=90s > /tmp/t1074-lane-status-unit.jsonl && jq -se '([.[] | select(.Action=="pass" and ((.Test // "") | test("^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$")))] | length)==4 and ([.[] | select(.Action=="skip" and ((.Test // "") | test("^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus)$")))] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0' /tmp/t1074-lane-status-unit.jsonl
```

성공 조건은 `jq` 출력 `true`와 종료 코드 0이다. 테스트 이름이 없어서 생기는 빈 성공은 허용하지 않는다.

### 6.2 production launcher·설치 binary·MCP 재시작·실세션 통합

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=operational-roster-before-prompt MOAI_HOME=/tmp/t1074-lane-live-prompt-home GOCACHE=/tmp/t1074-lane-live-prompt-cache go test -json ./internal/cli -run '^TestFactoryLiveOperationalRosterBeforePrompt$' -count=1 -timeout=240s > /tmp/t1074-lane-status-prompt-live.jsonl && jq -se '([.[] | select(.Action=="pass" and .Test=="TestFactoryLiveOperationalRosterBeforePrompt")] | length)==1 and ([.[] | select(.Action=="skip" and .Test=="TestFactoryLiveOperationalRosterBeforePrompt")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0 and any(.[]; (.Output // "") | contains("PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent")) and any(.[]; (.Output // "") | contains("SESSIONSTART_OWNER_MATCH_OK lead,agent-1,agent-2")) and any(.[]; (.Output // "") | contains("LEAD_MCP_ROSTER_OK lead,agent-1,agent-2")) and any(.[]; (.Output // "") | contains("NO_BYPASS_OK"))' /tmp/t1074-lane-status-prompt-live.jsonl

unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=operational-launcher-chain MOAI_HOME=/tmp/t1074-lane-live-chain-home GOCACHE=/tmp/t1074-lane-live-chain-cache go test -json ./internal/cli -run '^TestFactoryLiveOperationalLauncherChain$' -count=1 -timeout=240s > /tmp/t1074-lane-status-chain-live.jsonl && jq -se '([.[] | select(.Action=="pass" and .Test=="TestFactoryLiveOperationalLauncherChain")] | length)==1 and ([.[] | select(.Action=="skip" and .Test=="TestFactoryLiveOperationalLauncherChain")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0 and any(.[]; (.Output // "") | contains("PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent")) and any(.[]; (.Output // "") | contains("MCP_RESTART_OK")) and any(.[]; (.Output // "") | contains("REAL_SESSION_ROSTER_OK lead,agent-1,agent-2")) and any(.[]; (.Output // "") | contains("PROJECT_ISOLATION_OK")) and any(.[]; (.Output // "") | contains("CLEANUP_OK")) and any(.[]; (.Output // "") | contains("NO_BYPASS_OK"))' /tmp/t1074-lane-status-chain-live.jsonl
```

live 테스트는 실제 process tree, SessionStart peer row, lead-owned MCP 응답, binary hash, 프로젝트 격리, cleanup 및 우회 부재 assertion이 모두 끝난 뒤에만 위 sentinel을 출력해야 한다. process argv 로그는 built binary의 절대 경로를 포함해야 하며 sentinel의 축약 표기는 그 검증 결과다. 이 게이트는 현재 RED이며, 실행하지 않은 결과를 PASS로 기록해서는 안 된다.

## 7. 완료 증거 형식

구현 완료 보고에는 다음을 함께 남긴다.

- 설치 binary 절대 경로와 SHA-256
- built binary로 시작한 세 production argv와 PID/PPID process tree
- fixture 프로젝트 canonical key와 run ID(민감하지 않은 값만)
- 리더·작업자 두 개의 slot/backend/session UUID/generation/실제 Codex owner PID, 등록 fingerprint와 OS 재측정 fingerprint 일치 결과
- 리더 세션 MCP 재시작 전후 PID, 새 binary hash 및 그 MCP가 반환한 동일 세 endpoint
- direct `codex`/`codex exec`, 외부 `MOAI_SESSION_PID`, 직접 `RegisterPeer`, DB seed, 별도 owner process 미사용 assertion
- 조회 전후 DB 무변경 비교
- 다른 프로젝트 sentinel 부재 assertion
- 세 launcher process tree cleanup과 잔존 PID 부재
- §6 모든 게이트의 실제 종료 코드와 `jq` 출력

부분 구현, mock 통과, skip 또는 live 미실행은 명시적으로 `GAP` 또는 `NOT_RUN`으로 보고하며 완료로 간주하지 않는다.

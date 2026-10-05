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
- `sed -n '1,80p' internal/hook/factory_messages.go`와 `sed -n '430,515p' internal/hook/session_start.go`로 확인한 기존 hook session UUID 바인딩 경로는 `SessionStart` 처리 중 `RegisterPeer`를 호출한다. 그러나 공식 계약상 `SessionStart`는 startup/resume/clear/compact 수명주기 이벤트이며 정상 사용자 turn마다 재실행되는 이벤트가 아니다.

후속 실측은 신뢰 등록이 끝난 Codex `0.155.1` 픽스처에서 TUI를 30초 idle로 두어도 factory DB·peer가 생성되지 않았고, App Server에 `initialize`+`thread/start`만 보내 실제 thread ID를 받은 뒤 5초 대기해도 SessionStart sidecar·factory DB가 생성되지 않았음을 관측했다. 이어 실제 production launcher chain을 두 번 실행했을 때 모두 세 `launch-pending` 행과 정상 prompt의 실제 session은 관측됐지만 terminal 0은 재결합되지 않았다. 첫 시도의 owner PID 결함을 교정한 뒤에도 두 번째 시도가 같은 수명주기 경계에서 실패했으므로, 정상 prompt가 startup `SessionStart`를 다시 발생시킨다는 전제가 반증됐다. 이 결과는 해당 버전·픽스처의 기존 전제를 반증하지만, 모든 향후 Codex 버전의 보편 법칙으로 일반화하지 않는다.

따라서 최소 구현은 새로운 명부 저장소를 만들지 않고 기존 `peers`와 기존 읽기 전용 `factory_msg_status` 표면을 확장하되, launcher가 실제 child PID/process-start로 `launch-pending`을 등록하고 첫 정상·빈 값이 아닌 `UserPromptSubmit`이 inbox 처리 전에 실제 session UUID와 resolved owner PID/process-start로 `bound` 재결합하는 단계를 명시하는 것이다. startup `SessionStart`가 provisional 등록 뒤 실행되면 같은 결합을 조기에 시도할 수 있지만 정확성은 그 순서에 의존하지 않는다. 이미 같은 endpoint로 `bound`된 뒤의 prompt는 inbox를 계속 읽되 peer 행과 `updated_at`을 다시 쓰지 않는다.

## 2. 추가 운영 계약

### REQ-FMH-OPS-001 — canonical run 명부

팩토리 상태 조회가 호출되면 시스템은 호출 프로젝트의 canonical `project_key`와 명시된 active `run_id`가 모두 일치하는 현재 `peers` 행만 `lanes` 명부로 반환해야 한다. 다른 프로젝트나 다른 run의 행은 개수, 식별자, 상태 어느 형태로도 섞어 반환해서는 안 된다.

### REQ-FMH-OPS-002 — 논리 레인과 물리 세션 식별

각 명부 항목은 최소한 다음 값을 구조화된 필드로 반환해야 한다.

- 안정적인 논리 식별자: `slot`, `role`
- 실행 계열: `backend`
- endpoint 단계: `endpoint_phase` (`launch-pending` 또는 `bound`)
- 현재 물리 endpoint: `session_uuid`, `generation` (`launch-pending`에서 `session_uuid`는 미관측으로 비어 있어야 한다)
- 소유 프로세스: `pid`, 등록된 `process_start` fingerprint
- 관측 정보: 등록 갱신 시각, 상태 판정 시각

`slot`은 세션 재바인드 뒤에도 논리 레인 식별자로 유지되고, `generation`은 현재 물리 endpoint를 구분해야 한다. `session_uuid`는 hook이 실제로 관측한 endpoint를 `bound`한 단계에서만 필수다. 조회 구현이 미관측 UUID를 조작하거나 이 값을 제목, 메시지 수, 또는 launcher의 별도 JSON 명부에서 재구성해서는 안 된다.

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

### REQ-FMH-OPS-007 — capability-truth 초기 등록과 재결합

production 실증은 이번 tree에서 빌드한 동일 `moai` binary로 정확히 한 개의 `moai codex -f` 리더와 두 개의 `moai codex -f agent` 작업자 프로세스를 시작해야 한다. 첫 프롬프트 전에 launcher는 stable lane/run과 실제 child PID/process-start fingerprint를 근거로 세 `launch-pending` endpoint를 등록해야 하며 session UUID를 조작해서는 안 된다. startup `SessionStart`가 provisional 등록 뒤 실행되면 같은 검증·교체를 조기에 수행할 수 있지만 정확성은 그 순서에 의존해서는 안 된다. 각 레인의 첫 정상·빈 값이 아닌 `UserPromptSubmit`은 남아 있는 provisional 행을 inbox batch 처리 전에 관측된 실제 session UUID와 resolved owner PID/process-start로 원자적 `bound` 재결합해야 한다. 빈 문자열 또는 공백 전용 prompt는 bind하지 않아야 한다. 이미 같은 실제 endpoint로 `bound`된 후속 prompt는 inbox를 읽되 generation, peer 필드, `updated_at`을 다시 쓰지 않아야 한다. 빈/가짜 model turn, direct `codex`/`codex exec`, 외부 `MOAI_SESSION_PID` 주입, 직접 `RegisterPeer`, DB seed, 별도 owner 프로세스 또는 hook trust bypass로 이 chain을 대체해서는 안 된다.

### REQ-FMH-OPS-008 — 실패와 불확정성 보존

레지스트리 조회 실패, active run 불일치, process probe 불확정은 정상 빈 명부로 위장해서는 안 된다. 구조화된 오류 또는 해당 항목의 `unknown`으로 보존해야 한다. 정렬은 `slot` 기준으로 결정적이어야 하며, 동일 canonical project/run에서 한 `slot`에는 현재 generation 하나만 보여야 한다.

## 3. 범위

### 포함

- 기존 run-scoped `peers`를 읽는 명부 API와 상태 자료형
- launcher의 실제 프로세스 기반 `launch-pending` 등록, optional SessionStart 조기 bind, 첫 정상·빈 값이 아닌 UserPromptSubmit의 필수 원자적 `bound` 재결합과 bound steady-state 무쓰기
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

### M3 — launcher provisional 등록과 UserPromptSubmit 필수 재결합

- production launcher가 child 실행 후 실제 PID/process-start를 재측정하여 `launch-pending`을 등록하고, child 실행이 실패하면 provisional 행을 남기지 않도록 한다.
- `internal/hook/factory_messages.go`: 공통 bind 경로가 실제 hook session UUID와 provisional process identity를 검증해 원자적 `bound` 재결합을 수행하도록 하되, 이미 같은 endpoint가 bound이면 DB write 전에 no-op으로 종료하도록 확장한다.
- `internal/hook/session_start.go`와 `internal/hook/user_prompt_submit.go`: startup `SessionStart`는 provisional 행이 이미 있을 때 best-effort 조기 bind를 시도하고, 첫 정상·빈 값이 아닌 `UserPromptSubmit`은 inbox batch 전에 필수 fallback bind를 수행한다. empty/whitespace prompt는 bind하지 않으며, 후속 prompt는 bind write 없이 inbox 처리를 계속한다.
- `internal/hook/session_start_factory.go`, `internal/hook/session_start_factory_i18n.go`: 리더 안내가 `factory_msg_status`와 현재 run ID를 사용하도록 한다.
- launcher·hook 테스트는 `TestFactoryLauncherRegistersLaunchPendingPeers`, `TestFactorySessionStartRebindsLaunchPendingPeer`, `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer`, `TestFactoryBoundUserPromptSubmitDoesNotRewritePeer`로 provisional 상태에 session UUID가 없고 hook-bound 전달이 거부되는지, SessionStart 조기 경로가 순서 비의존적인 보조 경로인지, 첫 non-empty UserPromptSubmit이 inbox 전에 실제 endpoint를 결합하는지, 이미 bound인 동일 endpoint가 peer state를 쓰지 않는지 검증한다. 이 중 `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer`가 empty와 whitespace-only 입력 모두에 대해 peer가 `launch-pending`으로 유지되고 generation/identity/`updated_at`이 변하지 않음을 기계적으로 검증한다. 기존 안내 테스트는 tool이 MCP 등록표와 handler에 실제 존재하는지 검증한다.

### M4 — 설치 산출물로 별도 프로젝트 실증

- `internal/cli/factory_live_test.go`에 `TestFactoryLiveOperationalRosterBeforePrompt`와 `TestFactoryLiveOperationalLauncherChain`을 추가하고 cleanup-guaranteed 공통 production-launcher harness를 사용한다.
- 테스트는 저장소 밖 별도 임시 fixture 프로젝트를 만들고, 이번 tree에서 빌드·설치한 `moai` 산출물의 절대 경로와 SHA-256을 기록한 뒤 해당 binary를 가리키도록 MCP를 구성한다.
- harness는 그 binary의 production CLI를 통해 정확히 한 개의 `<built-moai> codex -f`와 두 개의 `<built-moai> codex -f agent`를 시작하고, 실제 process tree의 PID/PPID/argv를 기록해야 한다. direct `codex`/`codex exec`, 수동 `MOAI_SESSION_PID`, 직접 `RegisterPeer`, DB seed, 별도 sleep/owner 프로세스는 금지한다.
- 첫 model prompt를 보내지 않은 채 launcher가 등록한 세 `launch-pending` 행을 먼저 관측한다. 각 행의 PID/process-start가 launcher가 만든 실제 Codex owner process를 OS에서 재측정한 값과 같아야 하며 session UUID는 비어 있어야 한다.
- provisional assertion이 끝난 뒤에만 리더에게 상태 조회라는 명확한 정상·빈 값이 아닌 사용자 요청을 한 번 전달한다. 해당 `UserPromptSubmit`이 inbox 처리 전에 실제 session UUID와 owner identity로 재결합된 뒤 자기 MCP의 `factory_msg_status(run_id)`가 `bound` lead와 두 `launch-pending` worker를 같은 실제 endpoint와 함께 반환해야 한다.
- 이후 각 worker에게도 카드 인계 확인이라는 정상·빈 값이 아닌 첫 요청을 한 번씩 전달해 UserPromptSubmit fallback 재결합을 관측한다. 빈 문자열, 공백 전용, no-op, 꾸며낸 hook/model turn은 금지한다.
- 세 endpoint가 bound된 뒤 각 레인에 정상 후속 prompt를 한 번 더 전달하고, inbox 확인은 수행되지만 peer 행의 generation/identity/`updated_at` snapshot은 전후 동일함을 검증한다.
- 기존 MCP 프로세스를 재사용하지 않았음과 리더가 호출한 MCP의 PID/binary hash를 입증한다. 모든 자식 process tree는 테스트 cleanup에서 종료하고 잔존 PID가 없음을 확인한다.
- 각 항목의 backend/slot/session UUID/generation/PID/fingerprint를 대조한다. 두 작업자의 `task_state`는 근거가 없으므로 `unknown`이어야 한다.
- 별도의 두 번째 fixture 프로젝트에 의도적인 sentinel peer를 두고 첫 프로젝트 조회에 나타나지 않음을 검증한다.
- 최초 세 레인의 실세션 명부 assertion이 끝난 뒤 한 작업자 프로세스를 종료하여 `dead`를 검증한다. `stale`은 별도의 보조 행에 현재 살아 있는 실제 프로세스 PID와 의도적으로 불일치하는 등록 fingerprint를 넣고 실제 OS probe가 불일치를 발견하는 방식으로 검증한다. 이 보조 행은 최초 실세션 명부 assertion의 대체 증거가 아니며 가짜 process probe를 사용해서는 안 된다.

## 5. 이진 acceptance 계약

### AC-FMH-OPS-001 — 동일 run의 프롬프트 전 provisional 명부

Given 이번 tree의 built `moai`로 정확히 한 개의 `moai codex -f`와 두 개의 `moai codex -f agent`가 별도 fixture 프로젝트의 같은 active run에 시작되고 아직 어느 세션에도 prompt를 보내지 않았을 때, When 읽기 전용 명부를 조회하면, Then `lanes`는 `lead`, `agent-1`, `agent-2` 세 `launch-pending` 항목만 반환하고 각 backend/generation/PID/process-start가 실제 process tree와 일치하며 session UUID는 존재하지 않는다.

### AC-FMH-OPS-002 — 상태 축과 불확정성

Given live/dead/PID-fingerprint 불일치/indeterminate endpoint가 준비되었을 때, When 명부를 조회하면, Then `endpoint_state`는 각각 `live/dead/stale/unknown`이며 작업 관측이 없는 모든 항목의 `task_state`는 `unknown`이다.

### AC-FMH-OPS-003 — 프로젝트 격리

Given 서로 다른 canonical project가 같은 run 문자열과 겹치는 slot 이름을 사용할 때, When 첫 프로젝트에서 조회하면, Then 두 번째 프로젝트의 session UUID, PID, 집계 또는 상태는 응답에 하나도 포함되지 않는다.

### AC-FMH-OPS-004 — 무변경 조회

Given 메시지와 peer 상태의 전후 snapshot이 가능할 때, When `factory_msg_status`를 반복 호출하면, Then 메시지 state·claim token·receipt, peer generation·`updated_at`, 모든 관련 행 수가 호출 전과 동일하다.

### AC-FMH-OPS-005 — 첫 정상 turn 안내의 실행 가능성

Given startup `SessionStart` 안내가 존재하지만 리더가 여전히 `launch-pending`이고 첫 정상·빈 값이 아닌 사용자 요청이 제출될 때, When 해당 `UserPromptSubmit`이 안내가 제시한 상태 조회 표면을 같은 run에서 호출하기 전에 실행되면, Then 리더는 inbox 처리 전에 실제 session UUID와 resolved owner identity로 `bound`되고 등록된 MCP handler가 응답하며 `factory_msg_list`를 호출하거나 메시지를 claim하지 않는다.

### AC-FMH-OPS-006 — 설치 및 재시작 뒤 실제 통합 증거

Given 이번 tree의 설치 binary hash가 고정되고 기존 MCP가 종료되었을 때, When production argv인 한 개의 `moai codex -f`와 두 개의 `moai codex -f agent`만으로 세션을 시작하여 프롬프트 전 provisional 명부를 검증하고 각 레인에 한 번의 정상·빈 값이 아닌 첫 요청과 한 번의 정상 후속 요청을 전달한 뒤 리더 자신의 새 MCP에서 조회하면, Then 세 launcher argv/process tree, MCP PID/hash, 세 `launch-pending → bound` UserPromptSubmit 전이와 실제 session UUID, 세 실제 owner PID/fingerprint, bound 후 generation/identity/`updated_at` 무변경, 동일 endpoint 조회, 다른 프로젝트 격리 및 전체 child cleanup이 실제 출력에 남고 테스트가 통과한다.

AC-FMH-OPS-006은 §6.1 unit과 §6.2 LIVE가 모두 통과해야 충족된다. empty/whitespace no-bind는 `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer`의 unit 책임이고, LIVE는 프롬프트 전 pending, 각 레인의 첫 legitimate non-empty UserPromptSubmit bind-before-inbox, 후속 정상 prompt의 bound no-write를 책임진다. LIVE는 empty/whitespace 입력을 전송하거나 그 no-bind 동작을 증명하지 않는다. `t.Skip`, `NOT_RUN`, 가짜 model turn, direct `codex`/`codex exec`, 외부 `MOAI_SESSION_PID`, 직접 `RegisterPeer`, 사전 seed한 `Peer`, 별도 owner process, 가짜 process probe, hook trust bypass, in-process handler만 호출한 fixture, 또는 목업 명부는 AC-FMH-OPS-001/005/006의 PASS 증거가 될 수 없다.

## 6. 실행 게이트

### 6.1 단위·MCP·hook 계약

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1074-lane-status-home GOCACHE=/tmp/t1074-lane-status-cache go test -json ./internal/factorymsg ./internal/cli ./internal/hook -run '^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$' -count=1 -timeout=90s > /tmp/t1074-lane-status-unit.jsonl && jq -se '([.[] | select(.Action=="pass" and ((.Test // "") | test("^(TestFactoryLaneRosterStateTruth|TestFactoryLaneRosterProjectIsolation|TestFactoryMsgStatusReadOnlyRoster|TestFactoryLeadNoticeUsesOperationalStatus|TestFactoryLauncherRegistersLaunchPendingPeers|TestFactorySessionStartRebindsLaunchPendingPeer|TestFactoryUserPromptSubmitRebindsLaunchPendingPeer|TestFactoryBoundUserPromptSubmitDoesNotRewritePeer)$")))] | length)==8 and ([.[] | select(.Action=="skip")] | length)==0 and ([.[] | select(.Action=="fail")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0' /tmp/t1074-lane-status-unit.jsonl
```

성공 조건은 `jq` 출력 `true`와 종료 코드 0이다. 테스트 이름이 없어서 생기는 빈 성공은 허용하지 않는다. `TestFactoryUserPromptSubmitRebindsLaunchPendingPeer`는 non-empty bind-before-inbox와 함께 empty/whitespace-only no-bind를 담당하며, 이 입력 경계의 증거 책임은 §6.1에만 있다.

### 6.2 production launcher·설치 binary·MCP 재시작·실세션 통합

```bash
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=operational-roster-before-prompt MOAI_HOME=/tmp/t1074-lane-live-prompt-home GOCACHE=/tmp/t1074-lane-live-prompt-cache go test -json ./internal/cli -run '^TestFactoryLiveOperationalRosterBeforePrompt$' -count=1 -timeout=240s > /tmp/t1074-lane-status-prompt-live.jsonl && jq -se '([.[] | select(.Action=="pass" and .Test=="TestFactoryLiveOperationalRosterBeforePrompt")] | length)==1 and ([.[] | select(.Action=="skip")] | length)==0 and ([.[] | select(.Action=="fail")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0 and any(.[]; (.Output // "") | contains("PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent")) and any(.[]; (.Output // "") | contains("LAUNCH_PENDING_ROSTER_OK lead,agent-1,agent-2")) and any(.[]; (.Output // "") | contains("NO_SESSION_UUID_BEFORE_TURN_OK")) and any(.[]; (.Output // "") | contains("NO_BYPASS_OK"))' /tmp/t1074-lane-status-prompt-live.jsonl

unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_FACTORY_LIVE=1 MOAI_FACTORY_LIVE_CASE=operational-launcher-chain MOAI_HOME=/tmp/t1074-lane-live-chain-home GOCACHE=/tmp/t1074-lane-live-chain-cache go test -json ./internal/cli -run '^TestFactoryLiveOperationalLauncherChain$' -count=1 -timeout=240s > /tmp/t1074-lane-status-chain-live.jsonl && jq -se '([.[] | select(.Action=="pass" and .Test=="TestFactoryLiveOperationalLauncherChain")] | length)==1 and ([.[] | select(.Action=="skip")] | length)==0 and ([.[] | select(.Action=="fail")] | length)==0 and ([.[] | select((.Output // "") | contains("NOT_RUN"))] | length)==0 and any(.[]; (.Output // "") | contains("PRODUCTION_ARGV_OK moai codex -f | moai codex -f agent | moai codex -f agent")) and any(.[]; (.Output // "") | contains("MCP_RESTART_OK")) and any(.[]; (.Output // "") | contains("USERPROMPT_REBIND_OK lead,agent-1,agent-2")) and any(.[]; (.Output // "") | contains("BOUND_PROMPT_NO_REWRITE_OK lead,agent-1,agent-2")) and any(.[]; (.Output // "") | contains("BOUND_ROSTER_OK lead,agent-1,agent-2")) and any(.[]; (.Output // "") | contains("PROJECT_ISOLATION_OK")) and any(.[]; (.Output // "") | contains("CLEANUP_OK")) and any(.[]; (.Output // "") | contains("NO_SYNTHETIC_TURN_OR_BYPASS_OK"))' /tmp/t1074-lane-status-chain-live.jsonl
```

LIVE 테스트는 실제 process tree, 프롬프트 전 provisional peer row, 첫 정상·빈 값이 아닌 UserPromptSubmit의 inbox 전 재결합, bound 후 정상 prompt의 peer-state 무쓰기, lead-owned MCP 응답, binary hash, 프로젝트 격리, cleanup 및 가짜 turn·등록·신뢰 우회 부재 assertion이 모두 끝난 뒤에만 위 sentinel을 출력해야 한다. LIVE는 empty/whitespace 입력을 전송하거나 no-bind를 증명하지 않으며, 그 책임은 §6.1 unit gate에 있다. process argv 로그는 built binary의 절대 경로를 포함해야 하며 sentinel의 축약 표기는 그 검증 결과다. 이 게이트는 현재 RED이며, 실행하지 않은 결과를 PASS로 기록해서는 안 된다.

## 7. 완료 증거 형식

구현 완료 보고에는 다음을 함께 남긴다.

- 설치 binary 절대 경로와 SHA-256
- built binary로 시작한 세 production argv와 PID/PPID process tree
- fixture 프로젝트 canonical key와 run ID(민감하지 않은 값만)
- 리더·작업자 두 개의 slot/backend, 프롬프트 전 `launch-pending` generation/PID/fingerprint/UUID 미관측, 첫 정상·빈 값이 아닌 UserPromptSubmit 후 `bound` session UUID/generation/실제 Codex owner PID, 등록 fingerprint와 OS 재측정 fingerprint 일치, 후속 prompt 전후 peer generation/identity/`updated_at` 동일 결과
- 리더 세션 MCP 재시작 전후 PID, 새 binary hash 및 그 MCP가 반환한 동일 세 endpoint
- 가짜 model turn, direct `codex`/`codex exec`, 외부 `MOAI_SESSION_PID`, 직접 `RegisterPeer`, DB seed, 별도 owner process, hook trust bypass 미사용 assertion
- 조회 전후 DB 무변경 비교
- 다른 프로젝트 sentinel 부재 assertion
- 세 launcher process tree cleanup과 잔존 PID 부재
- §6 모든 게이트의 실제 종료 코드와 `jq` 출력

부분 구현, mock 통과, skip 또는 live 미실행은 명시적으로 `GAP` 또는 `NOT_RUN`으로 보고하며 완료로 간주하지 않는다.

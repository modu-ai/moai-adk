---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "design.md — 서버 요청 응답과 턴 단위 실패 격리 설계 결정 (F3·F4)"
version: "0.5.2"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# design.md — 설계 결정 D-1, D-2

> 상태축 없음(spec-frontmatter-schema.md § Artifact Statelessness). 생명주기는 `spec.md` 만 운반한다. 이 문서의 식별자(함수·필드 이름)는 설명을 위한 가칭이며 run 단계가 바꿀 수 있다. **관측 등급 표기**: "실측" = 이 plan 실행에서 명령을 돌려 본 것, "소스 판독" = 코드를 읽고 추론한 것.
>
> **개정 이력**: 0.1.0–0.3.0 은 F5(시그널·`Start`/`Close` 수명주기)를 포함했다. 0.4.0 은 운영자의 범위 분할 결정으로 **F5 를 카드 t1459 로 내보내고 F3·F4 만 남긴 개정**이다. 이전 D-3(시그널 설계, 수명 순서표)은 이 문서에서 완전히 빠졌고 git 이력(커밋 `95dfd85c8`, `820eff47f`, `951f2bfb6`)에서만 볼 수 있다. 이 문서는 그 내용을 어디에도 되풀이하지 않는다. 0.5.0 은 축소 범위 plan-audit 1차(FAIL 0.75, `bca1e0629`)의 개정으로, 로그 이음새의 동기화 규약·로그와 답장의 순서·쓰기 뮤텍스 범위·직전 턴 완료 프레임·상수 줄 번호만 바꿨다(D-1·D-2 의 설계 판단은 그대로). 0.5.2 는 완료된 SPEC 의 in-place 개정(독립 sync 감사 F1)으로, 정책표의 레거시 승인 두 줄(`applyPatchApproval`, `execCommandApproval`)의 응답 모양만 스키마에 맞게 고쳤고 정책 의미(거절·accept 없음·`abort` 없음)는 그대로다(spec.md HISTORY `### Amendments`).

## D-1 — 서버가 먼저 보내는 요청: 읽기 고루틴에서 최소 권한으로 답한다 (F3)

### 결정 1 — 프레임 분류와 쓰기 위치

현재 `read()` 는 `id != 0` 프레임을 전부 이벤트 채널에 올리고, `call()`/`waitTurn()` 은 `event.ID == id` 만 보므로 서버 요청이 (a) 버려지거나(실측: 응답 없음) (b) id가 겹치면 응답으로 오인되고(실측) (c) id가 문자열이면 읽기 루프가 끝난다(실측; `ID` 가 `int` 이기 때문).

**결정**: `read()` 가 프레임을 먼저 분류한다. 분류 기준은 필드 존재다 — `id` 와 `method` 가 모두 있으면 서버 요청, `method` 만 있으면 알림(지금처럼 `turn/started`·`turn/completed` 만 올리고 나머지는 버림), `id` 만 있으면 응답. `id` 는 원문(`json.RawMessage`)으로 받아 문자열·정수를 모두 담고, 응답 상관은 "정수이면서 대기 중인 클라이언트 id와 같음"으로만 한다. 서버 요청은 이벤트 채널에 올리지 않고 `read()` 고루틴에서 바로 답한다. 답장의 `id` 는 받은 원문 그대로 되돌린다.

**쓰기 위치와 이유**: gorilla/websocket은 동시 쓰기를 하나만 허용한다(`Close`·`WriteControl` 만 예외). 지금 쓰기 지점은 `call()` 과 `Start()` 의 `initialized` 알림이고, 둘 다 소유자의 메인 고루틴(핸드셰이크와 `DeliverTurn`)에서 직렬로 실행된다. 답장은 읽기 고루틴에서 쓰므로 쓰기 고루틴이 둘로 갈린다. 그래서 연결 쓰기를 뮤텍스 하나가 지키는 한 메서드로 모은다. **뮤텍스 범위**: 뮤텍스는 `WriteJSON` 호출 하나만 감싼다. `call()` 의 응답 대기(`select`)는 뮤텍스 밖이고, 읽기 고루틴은 턴 창 상태 잠금을 **풀고 난 뒤에** 답장을 쓰므로 두 잠금을 동시에 쥐는 곳이 없다(획득 순서 문제가 생기지 않는다). 읽기 고루틴이 답하는 이유는 둘이다.

1. 서버 요청은 턴 밖에서도 올 수 있다(인증 갱신, attestation). DeliverTurn 고루틴에서만 답하면 턴 사이에 온 요청은 다음 턴까지 응답이 없다.
2. 이벤트 채널(버퍼 32)로 DeliverTurn 고루틴에 넘기는 설계는 턴 사이에 소비자가 없어 버퍼가 차면 `read()` 가 막힌다. 서버는 우리 답을 기다리고 우리는 서버 프레임을 못 읽는 교착이 되고, 턴 타임아웃(10분)에서야 풀린다.

답장 쓰기가 실패하면 읽기 고루틴은 끝나고(연결이 죽은 것) 진행 중이던 DeliverTurn은 "connection closed"로 세션 치명 오류가 된다.

### 결정 2 — 응답 정책표 (스키마 실측: codex 0.160.0 `app-server generate-json-schema`)

기본값은 **거부 또는 아무것도 허가하지 않음**이다. 승인(accept 계열)은 하나도 쓰지 않는다 — 부모 D-5가 정한 MoAI 브로커 도구 사전 승인(`factoryMoAIMCPApprovalArgs`)이 이 계층의 유일한 권한 확대이고, 이 SPEC은 거기에 더하지 않는다.

| 요청 method | 응답 | 스키마 근거 | 선택 이유 |
|---|---|---|---|
| `item/commandExecution/requestApproval` | result `{"decision":"decline"}` | `CommandExecutionApprovalDecision`: `decline` = "agent will continue the turn", `cancel` = "turn will also be immediately interrupted" | 거부하되 턴은 이어 가야 모델이 승인된 MoAI 도구로 수신 확인(receipt)을 쓸 수 있다. `cancel` 은 턴을 끊어 F4의 턴 실패로 바꾸므로 쓰지 않는다. |
| `item/fileChange/requestApproval` | result `{"decision":"decline"}` | `FileChangeApprovalDecision` 동일 | 같은 이유 |
| `item/permissions/requestApproval` | result `{"permissions":{}}` | `GrantedPermissionProfile` 의 `fileSystem`·`network` 가 둘 다 선택·널 허용 → 빈 객체는 추가 허가 0. `scope`·`strictAutoReview` 는 생략 | 필수 필드 `permissions` 를 채우되 허가를 넓히지 않는다. |
| `mcpServer/elicitation/request` | result `{"action":"decline"}` | `McpServerElicitationAction` = accept/decline/cancel | 거부. 무인 세션에 답할 사람이 없다. |
| `item/tool/requestUserInput` | JSON-RPC error (code `-32000`) | `answers` 가 필수 맵이고 거부 표현이 스키마에 없다 | 빈 맵은 "사용자가 아무것도 답하지 않았다"는 주장이 되어 모델이 빈 답으로 진행할 수 있다. 오류는 도구 실패로 전달된다. |
| `item/tool/call` | result `{"contentItems":[{"type":"inputText","text":"…"}],"success":false}` | `DynamicToolCallResponse` 가 `contentItems`·`success` 필수 | 스키마가 정한 도구 실패 표현이다. 이 계층은 동적 도구를 등록하지 않으므로 정상 흐름에서는 오지 않는다. |
| `account/chatgptAuthTokens/refresh` | JSON-RPC error (`-32000`) | 응답이 `accessToken`·`chatgptAccountId` 필수 | 지어낼 수 없는 값이다. |
| `attestation/generate` | JSON-RPC error (`-32000`) | 응답이 `token` 필수 | 지어낼 수 없는 값이다. |
| `applyPatchApproval` | result `{"decision":{"denied":{"rejection":"<고정 문구 한 줄>"}}}` | 레거시 `ReviewDecision`(`ApplyPatchApprovalResponse`): `denied` 는 **객체** 변형이다 — 필수 키 `denied`, 그 값은 필수 문자열 `rejection` 을 가진 객체. 문자열 변형은 `approved`·`approved_for_session`·`approved_mcp_policy_amendment`·`timed_out`·`abort` 다. `denied` 객체는 "세션을 이어 가며 다른 방법을 시도"(우리가 원하는 의미), `abort` 는 턴을 끊는다 | `decline` 과 같은 이유로 `abort` 를 쓰지 않는다. `rejection` 문구는 코드의 상수 하나(`managedLegacyApprovalRejection`, `managedLegacyDeniedResult()` 가 만든다)이며 id·경로·비밀을 담지 않는다. 문자열 `"denied"` 는 스키마 위반이다(아래 정정 단락). |
| `execCommandApproval` | result `{"decision":{"denied":{"rejection":"<위와 같은 상수 문구>"}}}` | `ExecCommandApprovalResponse` 의 `ReviewDecision` 이 위와 같다 | 같음 |

**정정 (개정 0.5.2, 독립 sync 감사 F1)**: 0.1.0–0.5.1 의 이 두 줄은 `{"decision":"denied"}`(문자열)였고 codex 0.160.0 스키마에서 `ReviewDecision.denied` 가 객체 변형이라 스키마 위반이었다. 뿌리는 이 plan 산출물이다: plan 단계가 `ReviewDecision` 의 **열거 이름**(`denied`·`abort`)을 확인했을 뿐 **변형의 모양**(문자열인지 객체인지)을 확인하지 않았고, 세 번의 plan-audit 도 같은 이름 수준에서 읽었다. 그래서 이 표의 나머지 행도 응답 본문의 모양이 스키마에 맞는지를 `TestManagedServerRequestPolicyMatchesCodexSchema` 가 정책표 전체(10종과 미지 method 폴백)에 대해 검증한다 — 이름이 아니라 모양을 보는 가드다(acceptance.md AC-MH-001). 의미는 그대로다: 거절하고, accept 계열은 쓰지 않고, `abort` 도 쓰지 않는다.
| 그 밖에 `id` 를 가진 method | JSON-RPC error (`-32601`, method not found) | JSON-RPC 2.0 | 모르는 요청에 침묵하지 않는다. |

오류 응답의 `message` 는 영어이고 method 이름을 담는다(예: "managed Factory session cannot answer <method>: no operator is attached"). 정책표는 코드 안의 한 표 데이터이고 분기마다 판단을 흩지 않는다 — 표를 한 곳에서 읽고 한 곳에서 시험한다.

**운영자에게 보이는 흔적 (모든 응답 요청)**: 답한 서버 요청마다 로그 한 줄(영어)을 남긴다 — `Factory server request answered: <method> -> <decline|denied|empty|failed|error>` (이 토큰은 로그 어휘일 뿐이고 와이어 값의 모양을 뜻하지 않는다 — 레거시 두 종류의 `denied` 토큰은 위 표의 `denied` **객체** 응답에 대응한다). `mcpServer/elicitation/request` 줄은 이어서 `serverName=<name> turn=<귀속 턴 id|none> broker_declined=<k>` 를 담는다(`k` 는 그 요청을 처리한 뒤 열린 창에서 센 브로커 거부 수, 센 것이 없으면 0; 스키마 실측: `McpServerElicitationRequestParams` 의 필수 필드는 `serverName`·`threadId`, `turnId` 는 `string` 또는 `null`). 같은 한 곳(응답 정책표를 읽는 함수)에서 쓰므로 표와 로그가 갈라지지 않는다. **로그 줄은 답장을 쓰기 전에 쓴다** — 그래서 가짜 서버가 답장을 받은 시점에는 그 줄이 이미 있다. 그래도 시험의 모든 로그 단언은 상한 있는 폴링(5초, 10ms 간격)을 쓴다(순서가 바뀌는 구현 변경에도 시험이 구현 순서에 기대지 않게). 로그와 오류 `message` 에 들어가는 method 이름은 `%q` 로 인용해 출력한다(개행이 든 이름이 줄을 위조하지 못하게; 서버는 토큰 인증된 loopback 자식이라 위험은 낮다).

**로그 출력 이음새 (시험이 줄을 잡는 방법)**: 로그 줄과 F4 의 `Factory turn failed (…)` 줄은 모두 패키지 비공개 `atomic.Pointer[io.Writer]` 하나로 나간다(기본값 `os.Stderr`). `os.Stderr` 전역을 교체하지 않는다. **원자 포인터는 포인터 로드만 보호한다** — 가리키는 `io.Writer` 자체가 동기화돼 있지 않으면 쓰는 고루틴과 읽는 시험 고루틴 사이에 데이터 경합이 난다(plan-audit 가 `atomic.Pointer` + `bytes.Buffer` 로 `DATA RACE` 를 재현했다). 그래서 규약이 셋이다. (1) 시험이 꽂는 sink 는 쓰기와 스냅숏 조회가 **한 뮤텍스 아래** 있는 타입이어야 하고, 시험은 그 안의 원시 버퍼에 직접 접근하거나 `String()` 을 부르지 않는다(스냅숏 메서드만 쓴다). (2) 클라이언트를 시작한 시험은 정리에서 `shutdown()` 을 부르고 읽기 고루틴이 끝나기까지(`events` 채널이 닫힐 때까지) 기다린 **뒤에** 포인터를 복원한다(`t.Cleanup` 은 LIFO 이므로 포인터 복원을 먼저 등록하고 클라이언트 종료를 나중에 등록한다). 직전 시험의 고루틴이 다음 시험의 sink 에 쓰는 일이 없어야 줄 개수·`broker_declined=` 단언이 안정적이다. (3) 로그 대기가 있는 하위 케이스를 도는 모든 명령은 `-race` 를 포함한다(acceptance.md). 시험은 병렬로 돌리지 않는다. 이 이음새가 이 SPEC의 **유일한** 시험 이음새이며 프로덕션 비용은 포인터 로드 한 번이다.

### 결정 3 — MoAI 브로커 elicitation 거부는 그 턴의 턴 단위 실패다 (귀속은 읽기 고루틴의 스트림 순서로)

**목적(plan-audit D7)**: codex가 브로커 도구 승인을 elicitation으로 올리면 `decline` 은 도구 호출만 막고 턴은 `completed` 로 끝날 수 있다. 그러면 수신 확인이 안 써지고 claim은 lease(2분)마다 TTL까지 다시 배달되어 조용한 재배달 루프가 된다. 그 턴을 턴 단위 실패로 세면 연속 실패 상한(D-2)이 루프를 큰 소리의 정지로 바꾼다. 응답은 계속 `decline` 이며 이 규칙은 응답을 바꾸지 않는다.

**판별**: 거부한 요청의 `serverName` 이 MoAI 브로커 MCP 서버 이름과 같을 때만 센다. 그 이름은 같은 패키지의 기존 상수 `moaiMCPServerKey`(`mcp_server.go:57`, 값 `"moai"`)를 쓴다 — 리터럴을 새로 적지 않는다. 같은 값의 상수가 `moaiMCPServerName`(`:54`)에 따로 있고(`initialize` 의 서버 이름용) 둘은 지금 같은 `"moai"` 다 — 비교 기준은 `.mcp.json` 키인 `moaiMCPServerKey` 로 못 박고(codex config 의 `mcp_servers.<키>` 가 같은 키 공간이므로), 두 상수가 갈라지면 `TestManagedBrokerNameMatchesApprovalArgs` 가 붉어진다. 소유 App Server 명령행의 승인 인수 `mcp_servers.moai.*`(`managed_codex_factory.go:333-335`)는 PRESERVE 대상이라 고치지 않고, 두 이름이 같은 값임을 **시험이 고정한다**(AC-MH-006의 `TestManagedBrokerNameMatchesApprovalArgs`). codex가 요청의 `serverName` 에 config 키를 실제로 쓰는지는 **미관측**이다(아래 Gap).

**귀속 규칙 — 소비자가 아니라 읽기 고루틴이 정한다**: 소비자 쪽에서 계수기를 읽는 설계는 읽기 고루틴이 소비자보다 앞서 달릴 때(이벤트 채널 버퍼 32) 완료 프레임 **뒤에** 도착한 한가한 시간의 요청이 정상 완료된 턴을 실패로 만든다. 그래서 판정을 읽기 고루틴의 프레임 도착 순서 안에서 끝낸다.

- 읽기 고루틴이 상태(잠금 보호)를 쥔다: `open`(턴 창이 열렸는가), `turnID`(알려졌다면 그 id), `prevTurnID`(직전에 완료된 턴 id), `brokerDeclined`(창 안에서 센 브로커 거부 수).
- **창 열기·초기화**: `startTurn` 이 `turn/start` 를 **쓰기 전에** 호출하는 `armTurn` 이 `open=true`, `turnID=""`, `brokerDeclined=0` 으로 만든다. **`prevTurnID` 는 지우지 않는다**(직전 완료 턴 id 를 계속 기억한다). 초기화는 여기서만 한다.
- `turn/started` 프레임: 창이 열려 있고 `turnID` 가 비었으면 그 id를 `turnID` 로 기록.
- **브로커 elicitation 거부 시점의 귀속**: 요청의 `turnId`(`string` 또는 `null`)를 `t` 라 하자. (1) 창이 닫혀 있으면(**턴 사이 — 진행 중인 턴이 없음**) 로그만 남기고 어느 턴에도 세지 않는다. (2) 창이 열려 있고 `t` 가 있는데 **`t == prevTurnID`**(직전 완료 턴의 늦은 요청, `turn/started` 이전 구간 포함)이거나 **`turnID` 가 알려져 있고 `t != turnID`** 이면 세지 않는다. (3) 그 밖의 열린 창 안의 요청(`t` 가 null 이거나 `t` 가 현재 턴으로 읽히는 경우)은 `brokerDeclined` 를 올린다. JSON-RPC `id` 의 형태(정수·문자열)는 귀속과 무관하다.
- `turn/completed(X)` 프레임: 창이 열려 있고 `turnID` 가 비었거나 X와 같으면, 읽기 고루틴이 `brokerDeclined > 0` 이라는 **판정을 그 완료 이벤트에 실어** 소비자에게 올리고, `prevTurnID=X` 로 기록한 뒤 창을 닫는다. X가 다른 턴이면 판정 없이 올리고 창은 그대로 둔다. **`X == prevTurnID` 인 완료 프레임(직전 턴의 중복·지연 완료)은 창을 닫지 않고 판정도 싣지 않는다** — `armTurn` 직후 `turn/started` 이전에는 `turnID` 가 비어 있어 위 조건만으로는 새 턴의 창을 닫을 수 있기 때문이다(codex 가 한 턴에 완료 프레임을 한 번만 보내는지는 관측하지 못했다 — 그래서 방어만 두고 시험 행은 두지 않는다).
- 소비자는 완료 이벤트가 실어 온 판정만 읽는다. App Server가 그 턴을 `completed` 로 표시했어도 판정이 참이면 턴 단위 표식 오류(D-2의 `errManagedTurnFailed`)를 반환한다. 소비자 쪽에는 계수기도 리셋도 없다.

**이 규칙이 약속하는 것과 시험이 고정하는 것은 같은 문장이다**: "**창이 닫힌 때 도착한 요청, 직전에 완료된 턴 id 를 단 요청, 현재 턴으로 확정된 id 와 다른 id 를 단 요청은 어느 턴도 실패시키지 않는다. 열린 창 안의 나머지 요청은 그 턴을 정확히 한 번 실패시킨다(요청 수와 무관).**" AC-MH-006 의 하위 케이스 #11–#17 이 이 문장의 각 갈래를 하나씩 고정한다(acceptance.md §1.3).

**미관측 전제와 Gap (정정)**: 실제 codex가 각 거부·오류 응답을 받았을 때 모델이 어떻게 행동하는지, 그리고 요청의 `serverName` 에 config 키(`moai`)가 실제로 오는지는 관측하지 못했다. `mcpServer/elicitation/request` 를 거부하는 선택은 "부모 SPEC의 승인 스코핑(`approval_mode="approve"`)이 MoAI 브로커 도구의 승인창을 없앤다"는 전제에 기댄다. **이 전제는 관측된 사실이 아니라 미관측 전제다** — 부모 research.md(`:87`)가 "실제 Factory 승인 대화상자 유무는 다시 실세션에서 확인해야 한다 — live 게이트 테스트(M2)가 이 재확인을 소유"라고 남겼고, 부모의 `TestManagedCodexFactoryBrokerLive`(AC-MS-016)는 SKIP으로 통과했다. 부모의 AC-MS-012 가 확인하는 것은 인수 모양(`TestMoAIMCPApprovalArgsOnlyTargetMoAI`)뿐이다.

- 이 라이브 관측은 **run 진입 조건이 아니다**. 이름 붙은 Gap이며 `MOAI_FACTORY_LIVE_ROOT` 와 `MOAI_FACTORY_LIVE_RUN` 환경변수를 둔 환경에서 부모의 `TestManagedCodexFactoryBrokerLive` 로 돌릴 수 있다.
- 라이브를 돌린다면 볼 것: (1) 로그에 `Factory server request answered: mcpServer/elicitation/request … serverName=moai` 줄이 나타나는가, 그리고 `serverName` 값이 정말 `moai` 인가 — 나타나면 승인 스코핑이 브로커 elicitation을 없애지 못한다는 뜻이므로 이 정책표 줄을 재검토한다. (2) 수신 확인(`factory_msg_receipt`)이 실제로 써지는가 — 안 써지고 같은 `message_id` 가 lease 간격마다 다시 오면 위 조용한 루프다. (3) 명령·파일 승인 줄이 나타나는가 — 나타나면 모델이 승인되지 않은 동작을 시도한 것이다.

### 공시한 한계 (이 SPEC이 닫지 않음)

- **막힌 쓰기의 상한이 없다**: `call()` 은 `c.conn.WriteJSON(…)` 을 `select`(턴 컨텍스트·`done` 대기) **앞에서** 호출한다(`managed_codex_factory.go:212` 대 `:215`, 소스 판독). 그래서 서버가 읽지 않아 쓰기가 막히면 턴 타임아웃(10분)도 그것을 풀지 못한다. 쓰기 뮤텍스가 들어오면 읽기 고루틴의 답장 쓰기도 같은 뮤텍스에 걸려 함께 멈춘다. 이 한계의 상한은 "턴 타임아웃"이 아니라 **연결이 죽거나 세션이 닫힐 때까지**다(`Close` → `shutdown` 이 뮤텍스 없이 `conn.Close()` 를 부르므로 그때는 풀린다). 쓰기 데드라인 상수는 이 카드에서 더하지 않기로 판단했다: 상대는 우리가 소유한 loopback 자식이고 관측된 사례가 없다. 운영자 문서에 "상한 없음"으로 적고, 데드라인이 필요해지면 `defaults.go` 상수 하나와 REQ-MH-005 개정으로 닫는다(후속 후보). 이 단락은 소스 판독이며 실행으로 재현하지 않았다. `id` 필드가 없거나 원문이 JSON `null` 이면 "id 없음"으로 본다 — `method` 가 있으면 알림(지금처럼 대부분 버림), 없으면 버린다(`json.RawMessage` 가 리터럴 `null` 을 담아도 "id 가 있다"로 분류하지 않는다).

### 기각한 대안

- (a) 이벤트 채널로 DeliverTurn에 넘겨 거기서 답한다 — 위 이유 1·2.
- (b) 승인류에 `cancel`/`abort` 를 쓴다 — 턴을 즉시 끊어 모델이 수신 확인을 쓸 기회를 없앤다.
- (c) 사전 승인과 겹치는 요청은 accept 한다 — 요청 종류만으로는 어떤 도구 승인인지 안전하게 구별할 수 없다. 최소 권한이 기본이다.
- (d) 모르는 method를 무시한다 — 서버가 응답을 영원히 기다릴 수 있다(현재 결함의 본질).
- (e) 소비자 쪽 원자 계수기 + `startTurn` 리셋 — 읽기 고루틴이 소비자보다 앞서 달리면 완료 프레임 뒤의 요청이 정상 턴을 실패로 만든다(plan-audit N3). 판정을 완료 이벤트에 싣는 설계로 교체했다.
- (f) `turnID` 미확정 구간의 요청을 `turn/started` 까지 보류 — 보류 큐와 그 비움 규칙이 새로 필요하다. 직전 완료 턴 id 기억 하나(`prevTurnID`)로 같은 약속을 지킬 수 있다.

## D-2 — 오류 분류, 연속 실패 상한, claim 처분 (F4)

### 결정 1 — 분류는 "턴 단위로 명시한 것만 격리, 나머지는 세션 치명"

드라이버는 지금 `DeliverTurn` 오류를 모두 반환한다. **기본값을 그대로 둔다(실패에 닫힌 쪽)**: 오류는 소유자가 "이 턴만의 실패"라고 표시했을 때만 격리하고, 표시되지 않은 오류는 전부 세션 치명이다. 표시는 오류를 감싸는 한 가지 표식 오류(가칭 `errManagedTurnFailed`, `errors.Is` 로 판별)다. 소유자가 아래 세 곳에서만 표시한다.

| 실패 | 분류 | 근거 |
|---|---|---|
| 스트림 `result.is_error` (`pumpManagedStreamTurn`) | **턴 단위** | 세션이 살아 있고 결과 이벤트가 정상 도착했다. 일시적 429 같은 API 오류가 이 모양이다(감사서 §F4). |
| Codex 턴이 `completed` 가 아닌 상태로 종료(`turn ended as interrupted/failed`) | **턴 단위** | 서버가 `turn/completed` 를 정상 보고했다. 연결은 살아 있다. |
| 거부된 MoAI 브로커 `mcpServer/elicitation/request` 가 그 턴에 귀속된 Codex 턴(D-1 결정 3) — App Server가 그 턴을 `completed` 로 표시해도 | **턴 단위** | 브로커 도구가 막혀 수신 확인이 안 써졌을 가능성이 있는 턴이다. 턴 단위로 세어야 연속 실패 상한이 조용한 재배달 루프를 큰 소리의 정지로 바꾼다. 표식은 `turn/completed` 이벤트가 실어 온 판정으로 `waitTurn` 이 붙인다. |
| 우선 턴(priming) 실패 | **세션 치명** | 세션이 한 번도 쓸 수 있게 된 적이 없다. 기존 테스트 `TestManagedDriverFailureBranches` 가 이 동작을 고정한다. |
| 스트림 종료(`errManagedStreamClosed`), Codex "connection closed" | **세션 치명** | 소유한 자식이나 연결이 사라졌다. 다음 턴도 같은 이유로 실패한다. |
| 자식 stdin·WS 쓰기 실패, 스캐너 오류(한 줄 한도 초과) | **세션 치명** | 스트림 상태를 신뢰할 수 없다. |
| Codex 턴 타임아웃 (`DefaultManagedCodexTurnTimeout`) | **세션 치명** | 서버가 턴을 아직 돌리고 있을 수 있어 스레드가 비었는지 증명할 수 없다. 턴 단위로 올리려면 `turn/interrupt` 단계가 함께 필요한데 그것은 범위 밖이다. |
| 표시되지 않은 모든 오류(`turn/start` 의 JSON-RPC 오류 응답 포함) | **세션 치명** | 근거 없이 계속하지 않는다. 증거가 쌓이면 표에 한 줄 더하는 것으로 승격한다. |

**시험이 타임아웃을 만드는 방법(이음새 없음)**: `startTurn` 은 이미 컨텍스트를 인자로 받는다. 분류 시험은 `DeliverTurn` 대신 `startTurn` 에 50ms 마감 컨텍스트를 줘 `context.DeadlineExceeded` 를 얻고, 그 오류가 표식을 갖지 않음을 단언한다. 10분 상수를 줄이는 이음새를 만들지 않는다.

### 결정 2 — 연속 실패 상한 3, 기본값 파일에 둔다

`internal/config/defaults.go` 에 `DefaultManagedSessionMaxConsecutiveTurnFailures = 3` 을 둔다(같은 블록의 `DefaultManaged*` 상수 곁에 놓되 `DefaultManagedCodexTurnTimeout`(`:121`) **뒤**, 주석에 근거와 미측정 표기). **근거는 측정이 아니라 저장소의 관례**다: 헌장 Error Handling Protocol의 "Maximum 3 retries per operation". 실제 실패율 데이터는 없으므로 이 값은 UNMEASURED이고, 쿼터 게이트 기본값(`DefaultQuotaGate*`)과 같은 방식으로 그렇게 표기한다. 값이 틀렸다고 밝혀지면 상수 한 줄을 고친다.

의미: 연속 턴 단위 실패가 N번째에 닿으면 드라이버가 그 N번째 오류에 횟수를 붙여 반환한다. 성공한 턴(운영자 입력 턴 포함)이 횟수를 0으로 되돌린다. 우선 턴은 세지 않는다.

### 결정 3 — 운영자에게 보이는 기록

턴 단위 실패마다 로그 한 줄(영어)을 남긴다: `Factory turn failed (<k>/<N> consecutive): <cause>`. 출력은 D-1 의 로그 출력 이음새를 쓴다(stdout 은 모델 출력 몫이라 쓰지 않는다). claim 오류처럼 같은 문구를 한 번만 찍는 중복 제거는 하지 않는다 — 연속 실패는 N번이 끝이라 홍수가 되지 않고, 매번 새 정보(k)를 담는다.

### 결정 4 — claim된 메시지의 처분: 관리 계층은 아무것도 하지 않는다 (실측·소스 판독)

관측(`store.go`, 기준 트리 `7109e0900`):

1. 클라이언트가 쓸 수 있는 호출은 `Claim`·`ReadBody`·`RecordDisposition`·`Receipt` 이고 claim을 풀거나 되돌리는 호출은 없다. `messages` 테이블에 시도 횟수 열도 없다(`:389`).
2. 실패한 턴이 남긴 claim은 `claim_expires_at`(= claim 시각 + `DefaultManagedSessionClaimLease` 2분)이 지나면 `Claim` 의 후보 조건(`state='claimed' AND claim_expires_at<=now`, `:878`·`:901`)에 다시 들고, 재claim은 새 claim token을 주고 `disposition` 을 비운다(`:916`). 기존 테스트 `TestDispatchResultExactlyOnce/lost_receipt_redelivery` 가 기준 트리에서 PASS했다(실측).
3. 메시지는 `expires_at`(송신 시 TTL, 최대 7일)이 지나면 `Claim` 이 dead-letter(`ttl:expired`)로 보낸다(`:892-898`).
4. 실패한 턴이 일부 메시지에는 이미 receipt를 썼다면 그 행은 `acknowledged` 라 재배달되지 않는다.

**결정**: 턴 실패 때 관리 계층은 claim을 해제·확인·재주소하지 않는다(REQ-MH-009). 재배달은 lease 만료 뒤 브로커 정책이 한다. 이유: (a) 부모 REQ-MS-008이 `store.go` 무수정을 원칙으로 하고, 이 SPEC도 그것을 잇는다. (b) 해제 호출이 없는 API에서 관리 계층이 흉내 내면 상태를 두 곳에서 쥐게 된다.

**독 메시지(poison message)가 영원히 재배달될 수 있는가 — 정직한 답**:

- 살아 있는 세션 안에서, 그 메시지가 매번 턴을 실패시키고 사이에 성공 턴이 없으면: lease 간격(2분) 이상으로 벌어진 실패 3번 뒤에 세션이 끝난다. 영원하지 않다.
- 사이에 성공 턴(운영자 입력, 다른 메시지)이 끼면 횟수가 0으로 돌아가므로 **독 메시지는 lease 만료마다 TTL(최대 7일)까지 재배달될 수 있다.** 관리 계층에도 브로커에도 시도 상한이 없다. 이것이 이 SPEC 뒤에 남는 한계이고 docs·CHANGELOG에 그렇게 적는다(REQ-MH-010). 닫으려면 브로커에 시도 횟수와 nack/dead-letter 승격이 필요하며 그것은 별도 SPEC이다.
- 세션이 끝난 뒤: claim된 행은 그 엔드포인트(세션 UUID·세대) 앞으로 남는다. `Claim` 은 그 수신자만 호출하므로 TTL 정리도 그 수신자가 `Claim` 할 때만 돈다 — 새 세션은 새 세션 UUID라 인수하지 않는다는 것은 **소스 판독 추론**이고(핸드오프 해제 경로는 별개 메커니즘) 실측하지 않았다.

### 기각한 대안

- (a) 모든 턴 오류를 턴 단위로 본다 — 닫힌 스트림·죽은 연결에서 N번 더 쓰기를 시도하게 되고, 분류되지 않은 새 오류가 조용히 삼켜진다.
- (b) 타임아웃을 `turn/interrupt` 와 함께 턴 단위로 올린다 — 새 RPC가 범위 밖이고 인터럽트 응답까지 정의해야 한다.
- (c) 상한 없이 계속한다 — 결정적으로 실패하는 입력이 모델 호출을 무한히 태운다.
- (d) 실패 뒤 지수 백오프 — 상한 3이면 필요가 없다. 상한을 올리는 날 다시 본다.
- (e) 드라이버가 메시지 id별 시도 횟수를 센다 — 세션 사이에 사라지는 기억이고 브로커가 소유할 일이다.

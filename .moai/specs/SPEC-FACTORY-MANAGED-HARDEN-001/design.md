---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "design.md — 관리 세션 소유자 강건화 설계 결정"
version: "0.3.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# design.md — 설계 결정 D-1..D-3

> 상태축 없음(spec-frontmatter-schema.md § Artifact Statelessness). 생명주기는 `spec.md` 만 운반한다. 이 문서의 식별자(함수·필드·훅 이름)는 설명을 위한 가칭이며 run 단계가 바꿀 수 있다. **관측 등급 표기**: "실측" = 이 plan 실행에서 명령을 돌려 본 것, "소스 판독" = 코드를 읽고 추론한 것.
>
> **개정 이력**: 0.1.0 초안 · 0.2.0 plan-audit 1차 개정(D1–D10: Start/Close 수명주기 규칙 신설, elicitation 전제 정정과 탐지 수단) · 0.3.0 plan-audit 2차 개정(N1–N7: **수명 순서를 D-3.1 한 표로 못 박음**, `started` 게이트 모순 해소, 시그널 컨텍스트를 소유자 진입 첫 단계로, elicitation 귀속을 읽기 고루틴 스트림 순서로, 쓰기 한계 문장 정정).

## D-1 — 서버가 먼저 보내는 요청: 읽기 고루틴에서 최소 권한으로 답한다 (F3)

### 결정 1 — 프레임 분류와 쓰기 위치

현재 `read()` 는 `id != 0` 프레임을 전부 이벤트 채널에 올리고, `call()`/`waitTurn()` 은 `event.ID == id` 만 보므로 서버 요청이 (a) 버려지거나(실측: 응답 없음) (b) id가 겹치면 응답으로 오인되고(실측) (c) id가 문자열이면 읽기 루프가 끝난다(실측; `ID` 가 `int` 이기 때문).

**결정**: `read()` 가 프레임을 먼저 분류한다. 분류 기준은 필드 존재다 — `id` 와 `method` 가 모두 있으면 서버 요청, `method` 만 있으면 알림(지금처럼 `turn/started`·`turn/completed` 만 올리고 나머지는 버림), `id` 만 있으면 응답. `id` 는 원문(`json.RawMessage`)으로 받아 문자열·정수를 모두 담고, 응답 상관은 "정수이면서 대기 중인 클라이언트 id와 같음"으로만 한다. 서버 요청은 이벤트 채널에 올리지 않고 `read()` 고루틴에서 바로 답한다. 답장의 `id` 는 받은 원문 그대로 되돌린다.

**쓰기 위치와 이유**: gorilla/websocket은 동시 쓰기를 하나만 허용한다(`Close`·`WriteControl` 만 예외). 지금 쓰기 지점은 `call()` 과 `Start()` 의 `initialized` 알림이고, 둘 다 소유자의 메인 고루틴(핸드셰이크와 `DeliverTurn`)에서 직렬로 실행된다. 답장은 읽기 고루틴에서 쓰므로 쓰기 고루틴이 둘로 갈린다. 그래서 연결 쓰기를 뮤텍스 하나가 지키는 한 메서드로 모은다. 읽기 고루틴이 답하는 이유는 둘이다.

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
| `applyPatchApproval` | result `{"decision":"denied"}` | 레거시 `ReviewDecision`: `denied` = 거부하고 세션은 계속, `abort` = 사용자의 다음 명령까지 정지 | `decline` 과 같은 이유로 `abort` 를 쓰지 않는다. |
| `execCommandApproval` | result `{"decision":"denied"}` | 같음 | 같음 |
| 그 밖에 `id` 를 가진 method | JSON-RPC error (`-32601`, method not found) | JSON-RPC 2.0 | 모르는 요청에 침묵하지 않는다. |

오류 응답의 `message` 는 영어이고 method 이름을 담는다(예: "managed Factory session cannot answer <method>: no operator is attached"). 정책표는 코드 안의 한 표 데이터이고 분기마다 판단을 흩지 않는다 — 표를 한 곳에서 읽고 한 곳에서 시험한다.

**운영자에게 보이는 흔적 (모든 응답 요청)**: 답한 서버 요청마다 stderr에 한 줄(영어)을 남긴다 — `Factory server request answered: <method> -> <decline|denied|empty|failed|error>`. `mcpServer/elicitation/request` 줄은 `serverName` 과 **귀속된 턴 id(없으면 `none`)** 도 담는다(`… serverName=<name> turn=<id|none>`; 스키마 실측: `McpServerElicitationRequestParams` 의 필수 필드는 `serverName`·`threadId`, `turnId` 는 `string` 또는 `null`). 같은 한 곳(응답 정책표를 읽는 함수)에서 쓰므로 표와 로그가 갈라지지 않는다. AC-MH-001 의 가짜 서버 시험이 11개 하위 케이스 각각에서 그 줄을 단언한다.

### 결정 3 — MoAI 브로커 elicitation 거부는 그 턴의 턴 단위 실패다 (귀속은 읽기 고루틴의 스트림 순서로)

**목적(plan-audit D7)**: codex가 브로커 도구 승인을 elicitation으로 올리면 `decline` 은 도구 호출만 막고 턴은 `completed` 로 끝날 수 있다. 그러면 수신 확인이 안 써지고 claim은 lease(2분)마다 TTL까지 다시 배달되어 조용한 재배달 루프가 된다. 그 턴을 턴 단위 실패로 세면 연속 실패 상한(D-2)이 루프를 큰 소리의 정지로 바꾼다. 응답은 계속 `decline` 이며 이 규칙은 응답을 바꾸지 않는다.

**판별**: 거부한 요청의 `serverName` 이 MoAI 브로커 MCP 서버 이름과 같을 때만 센다. 그 이름은 같은 패키지의 기존 상수 `moaiMCPServerKey`(`mcp_server.go:56`, 값 `"moai"`)를 쓴다 — 리터럴을 새로 적지 않는다. 소유 App Server 명령행의 승인 인수 `mcp_servers.moai.*`(`managed_codex_factory.go:333-335`)는 PRESERVE 대상이라 고치지 않고, 두 이름이 같은 값임을 **시험이 고정한다**(AC-MH-006: 승인 인수 문자열이 `mcp_servers.` + 상수 + `.` 로 시작하는 접두를 가짐). codex가 요청의 `serverName` 에 config 키를 실제로 쓰는지는 **미관측**이다(아래 Gap).

**귀속 규칙 — 소비자가 아니라 읽기 고루틴이 정한다**: 이전 설계는 `startTurn` 이 계수기를 0으로 되돌리고 `turn/completed` 소비 시점에 읽었는데, 읽기 고루틴이 소비자보다 앞서 달리면(이벤트 채널 버퍼 32) 완료 프레임 **뒤에** 도착한 한가한 시간의 요청이 정상 완료된 턴을 실패로 만든다. 그래서 판정을 읽기 고루틴의 프레임 도착 순서 안에서 끝낸다.

- 읽기 고루틴이 상태(잠금 보호)를 쥔다: `open`(턴 창이 열렸는가), `turnID`(알려졌다면 그 id), `brokerDeclined`(창 안에서 거부한 브로커 elicitation 수).
- **창 열기·초기화**: `startTurn` 이 `turn/start` 를 **쓰기 전에** 호출하는 `armTurn` 이 `open=true`, `turnID=""`, `brokerDeclined=0` 으로 만든다. 초기화는 여기서만 한다(읽기 고루틴 상태, 쓰기보다 앞).
- `turn/started` 프레임: 창이 열려 있고 `turnID` 가 비었으면 그 id를 `turnID` 로 기록.
- 브로커 elicitation 거부 시점: 요청의 `turnId` 가 있으면 그것을, 없으면 열려 있는 창의 턴을 그 요청의 귀속 턴으로 본다. 창이 닫혀 있으면(**턴 사이 — 진행 중인 턴이 없음**) 로그만 남기고 어느 턴에도 세지 않는다. 창이 열려 있고 `turnId` 가 있는데 알려진 `turnID` 와 다르면 이전 턴의 늦은 요청이므로 세지 않는다. 그 밖에는 `brokerDeclined` 를 올린다.
- `turn/completed(X)` 프레임: 창이 열려 있고 `turnID` 가 비었거나 X와 같으면, 읽기 고루틴이 `brokerDeclined > 0` 이라는 **판정을 그 완료 이벤트에 실어** 소비자에게 올린 뒤 창을 닫는다. X가 다른 턴이면 판정 없이 올리고 창은 그대로 둔다.
- 소비자는 완료 이벤트가 실어 온 판정만 읽는다. App Server가 그 턴을 `completed` 로 표시했어도 판정이 참이면 턴 단위 표식 오류(D-2의 `errManagedTurnFailed`)를 반환한다. 소비자 쪽에는 계수기도 리셋도 없다.

이 규칙의 귀결(AC-MH-006 하위 케이스로 고정): (a) 턴 사이에 온 `moai` elicitation은 어느 턴도 실패시키지 않고 다음 정상 턴은 nil. (b) 거부된 턴 다음의 정상 턴은 nil(초기화는 `armTurn` 에서). (c) `turn/start` 응답 직후 `turn/started` 이전에 온, `turnId` 가 없는 요청도 열린 창에 귀속되어 그 턴을 실패시킨다. (d) 한 턴 안의 둘 이상은 정확히 한 번의 턴 단위 실패다(판정은 불리언).

**미관측 전제와 Gap (정정)**: 실제 codex가 각 거부·오류 응답을 받았을 때 모델이 어떻게 행동하는지, 그리고 요청의 `serverName` 에 config 키(`moai`)가 실제로 오는지는 관측하지 못했다. `mcpServer/elicitation/request` 를 거부하는 선택은 "부모 SPEC의 승인 스코핑(`approval_mode="approve"`)이 MoAI 브로커 도구의 승인창을 없앤다"는 전제에 기댄다. **이 전제는 관측된 사실이 아니라 미관측 전제다** — 부모 research.md(`:87`)가 "실제 Factory 승인 대화상자 유무는 다시 실세션에서 확인해야 한다 — live 게이트 테스트(M2)가 이 재확인을 소유"라고 남겼고, 부모의 `TestManagedCodexFactoryBrokerLive`(AC-MS-016)는 SKIP으로 통과했다. 부모의 AC-MS-012 가 확인하는 것은 인수 모양(`TestMoAIMCPApprovalArgsOnlyTargetMoAI`)뿐이다.

- 이 라이브 관측은 **run 진입 조건이 아니다**. 이름 붙은 Gap이며 `MOAI_FACTORY_LIVE_ROOT` 와 `MOAI_FACTORY_LIVE_RUN` 환경변수를 둔 환경에서 부모의 `TestManagedCodexFactoryBrokerLive` 로 돌릴 수 있다.
- 라이브를 돌린다면 볼 것: (1) stderr에 `Factory server request answered: mcpServer/elicitation/request … serverName=moai` 줄이 나타나는가, 그리고 `serverName` 값이 정말 `moai` 인가 — 나타나면 승인 스코핑이 브로커 elicitation을 없애지 못한다는 뜻이므로 이 정책표 줄을 재검토한다. (2) 수신 확인(`factory_msg_receipt`)이 실제로 써지는가 — 안 써지고 같은 `message_id` 가 lease 간격마다 다시 오면 위 조용한 루프다. (3) 명령·파일 승인 줄이 나타나는가 — 나타나면 모델이 승인되지 않은 동작을 시도한 것이다.

### 기각한 대안

- (a) 이벤트 채널로 DeliverTurn에 넘겨 거기서 답한다 — 위 이유 1·2.
- (b) 승인류에 `cancel`/`abort` 를 쓴다 — 턴을 즉시 끊어 모델이 수신 확인을 쓸 기회를 없앤다.
- (c) 사전 승인과 겹치는 요청은 accept 한다 — 요청 종류만으로는 어떤 도구 승인인지 안전하게 구별할 수 없다. 최소 권한이 기본이다.
- (d) 모르는 method를 무시한다 — 서버가 응답을 영원히 기다릴 수 있다(현재 결함의 본질).
- (e) 소비자 쪽 원자 계수기 + `startTurn` 리셋 — 읽기 고루틴이 소비자보다 앞서 달리면 완료 프레임 뒤의 요청이 정상 턴을 실패로 만든다(plan-audit N3). 판정을 완료 이벤트에 싣는 설계로 교체했다.

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
| 시그널(D-3) | 별개 — 실패가 아니라 중단 | 드라이버는 시그널 확인을 분류보다 먼저 한다. 중단으로 생긴 오류를 턴 실패로 세거나 로그에 남기지 않는다. |

### 결정 2 — 연속 실패 상한 3, 기본값 파일에 둔다

`internal/config/defaults.go` 에 `DefaultManagedSessionMaxConsecutiveTurnFailures = 3` 을 둔다(같은 블록의 `DefaultManaged*` 상수 곁, 주석에 근거와 미측정 표기). **근거는 측정이 아니라 저장소의 관례**다: 헌장 Error Handling Protocol의 "Maximum 3 retries per operation". 실제 실패율 데이터는 없으므로 이 값은 UNMEASURED이고, 쿼터 게이트 기본값(`DefaultQuotaGate*`)과 같은 방식으로 그렇게 표기한다. 값이 틀렸다고 밝혀지면 상수 한 줄을 고친다.

의미: 연속 턴 단위 실패가 N번째에 닿으면 드라이버가 그 N번째 오류에 횟수를 붙여 반환한다. 성공한 턴(운영자 입력 턴 포함)이 횟수를 0으로 되돌린다. 우선 턴은 세지 않는다.

### 결정 3 — 운영자에게 보이는 기록

턴 단위 실패마다 stderr에 한 줄(영어)을 남긴다: `Factory turn failed (<k>/<N> consecutive): <cause>`. stdout은 모델 출력 몫이라 쓰지 않는다. claim 오류처럼 같은 문구를 한 번만 찍는 중복 제거는 하지 않는다 — 연속 실패는 N번이 끝이라 홍수가 되지 않고, 매번 새 정보(k)를 담는다.

### 결정 4 — claim된 메시지의 처분: 관리 계층은 아무것도 하지 않는다 (실측·소스 판독)

관측(`store.go`, 기준 트리 `7109e0900`):

1. 클라이언트가 쓸 수 있는 호출은 `Claim`·`ReadBody`·`RecordDisposition`·`Receipt` 이고 claim을 풀거나 되돌리는 호출은 없다. `messages` 테이블에 시도 횟수 열도 없다(`:389`).
2. 실패한 턴이 남긴 claim은 `claim_expires_at`(= claim 시각 + `DefaultManagedSessionClaimLease` 2분)이 지나면 `Claim` 의 후보 조건(`state='claimed' AND claim_expires_at<=now`, `:878`·`:901`)에 다시 들고, 재claim은 새 claim token을 주고 `disposition` 을 비운다(`:916`). 기존 테스트 `TestDispatchResultExactlyOnce/lost_receipt_redelivery` 가 기준 트리에서 PASS했다(실측).
3. 메시지는 `expires_at`(송신 시 TTL, 최대 7일)이 지나면 `Claim` 이 dead-letter(`ttl:expired`)로 보낸다(`:892-898`).
4. 실패한 턴이 일부 메시지에는 이미 receipt를 썼다면 그 행은 `acknowledged` 라 재배달되지 않는다.

**결정**: 턴 실패 때 관리 계층은 claim을 해제·확인·재주소하지 않는다(REQ-MH-009). 재배달은 lease 만료 뒤 브로커 정책이 한다. 이유: (a) 부모 REQ-MS-008이 `store.go` 무수정을 원칙으로 하고, 이 SPEC도 그것을 잇는다. (b) 해제 호출이 없는 API에서 관리 계층이 흉내 내면 상태를 두 곳에서 쥐게 된다.

**독 메시지(poison message)가 영원히 재배달될 수 있는가 — 정직한 답**:

- 살아 있는 세션 안에서, 그 메시지가 매번 턴을 실패시키고 사이에 성공 턴이 없으면: lease 간격(2분) 이상으로 벌어진 실패 3번 뒤에 세션이 끝난다. 영원하지 않다.
- 사이에 성공 턴(운영자 입력, 다른 메시지)이 끼면 횟수가 0으로 돌아가므로 **독 메시지는 lease 만료마다 TTL(최대 7일)까지 재배달될 수 있다.** 관리 계층에도 브로커에도 시도 상한이 없다. 이것이 이 SPEC 뒤에 남는 한계이고 docs·CHANGELOG에 그렇게 적는다(REQ-MH-014). 닫으려면 브로커에 시도 횟수와 nack/dead-letter 승격이 필요하며 그것은 별도 SPEC이다.
- 세션이 끝난 뒤: claim된 행은 그 엔드포인트(세션 UUID·세대) 앞으로 남는다. `Claim` 은 그 수신자만 호출하므로 TTL 정리도 그 수신자가 `Claim` 할 때만 돈다 — 새 세션은 새 세션 UUID라 인수하지 않는다는 것은 **소스 판독 추론**이고(핸드오프 해제 경로는 별개 메커니즘) 실측하지 않았다.

### 기각한 대안

- (a) 모든 턴 오류를 턴 단위로 본다 — 닫힌 스트림·죽은 연결에서 N번 더 쓰기를 시도하게 되고, 분류되지 않은 새 오류가 조용히 삼켜진다.
- (b) 타임아웃을 `turn/interrupt` 와 함께 턴 단위로 올린다 — 새 RPC가 범위 밖이고 인터럽트 응답까지 정의해야 한다.
- (c) 상한 없이 계속한다 — 결정적으로 실패하는 입력이 모델 호출을 무한히 태운다.
- (d) 실패 뒤 지수 백오프 — 상한 3이면 필요가 없다. 상한을 올리는 날 다시 본다.
- (e) 드라이버가 메시지 id별 시도 횟수를 센다 — 세션 사이에 사라지는 기억이고 브로커가 소유할 일이다.

## D-3 — 시그널: 동시 안전 Close를 인터럽트로 쓰고 드라이버에 컨텍스트를 준다 (F5)

### 문제의 구조

진행 중인 `DeliverTurn` 은 두 소유자 모두에서 컨텍스트로 풀리지 않는 읽기 위에 있다. 스트림 소유자는 자식 stdout 파이프를 `bufio.Scanner` 로 읽는다(컨텍스트 인자가 없다). Codex 소유자는 이벤트 채널 select 에서 기다리는데 이 select 는 `ctx.Done()` 과 `done` 채널에 반응하지만 `DeliverTurn` 은 컨텍스트를 `context.Background()` 에서 만든다. 또 드라이버는 한가할 때 operator stdin 채널과 poll tick에서 기다리므로 `DeliverTurn` 밖이다.

### 대안 비교

| | A. `DeliverTurn(ctx, prompt)` 로 인터페이스 변경 | B. 동시 안전 `Close` 를 인터럽트로 쓰고, 드라이버 루프에만 `ctx` 추가 (**선택**) | C. 시그널 고루틴이 정리를 직접 하고 `os.Exit` |
|---|---|---|---|
| 스트림 소유자의 블로킹 읽기 | 컨텍스트로 풀리지 않는다. 읽기를 별도 고루틴으로 빼거나 자식을 죽여야 풀린다 — 결국 B의 kill이 필요하다 | 자식을 죽이면 stdout이 EOF를 내 읽기가 풀린다 | 풀 필요 없음(프로세스가 끝남) |
| Codex 소유자의 블로킹 | ctx 취소로 풀린다 | `Close` → `shutdown` 이 `done` 을 닫고 연결을 닫아 풀린다 | 풀 필요 없음 |
| 한가한 드라이버 | 인터페이스 변경과 별개로 드라이버에도 취소 경로가 필요하다 | 드라이버 `select` 에 `ctx.Done()` 추가 | 풀 필요 없음 |
| 수정 지점(실측: `grep -c` 기준 텍스트 출현 수) | `DeliverTurn(` 13곳(6개 파일) + 구현체 2종 + 가짜 `fakeManagedSession` 1종 + `driveManagedFactorySession(` 출현 10곳(5개 파일; 정의 1 + 호출 9) | `driveManagedFactorySession(` 출현 10곳(정의 1 + 호출 9: 비테스트 2·테스트 7, 5개 파일)에 `ctx` 첫 인자 추가, 인터페이스와 가짜 구현체는 무변경 | 정리를 한 번 더 구현 |
| 정리 경로 수 | 1(기존) | 1(기존) — `Close`·`defer` 그대로 | 2(`defer` 는 `os.Exit` 에서 돌지 않으므로 launch-pending 롤백·토큰 삭제를 다시 써야 한다) |
| 새 위험 | 인터페이스 확장이 이후 소유자 모두에 강제된다 | `Close` 가 `Start` 와 경합 없이 동시 호출 안전해야 한다(D-3.1) | 정리가 두 벌로 갈라진다 |

**선택 B.** 이유: (1) 스트림 소유자에서 A도 결국 자식을 죽여야 하므로 A는 B보다 작지 않다. (2) A는 인터페이스를 바꿔 가짜 구현체와 이후 소유자 전부에 파급되고, B는 드라이버 호출부 기계적 수정이 전부다. (3) B는 기존 정리 경로(`Close` + `defer`)를 그대로 쓰므로 시그널 때문에 정리가 갈라지지 않는다.

### B의 구성 요소 (가칭; 순서는 D-3.1이 정한다)

- **시그널 구독 도우미**: `signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)` 를 감싼 작은 도우미를 `internal/cli/launch_signals.go` 에 둔다. 첫 시그널이 오면 `stop()` 으로 기본 동작을 복원한다 — 정리가 멈췄을 때 두 번째 시그널이 프로세스를 끝낼 수 있게 하려는 것이다.
- **워처**: 세션이 만들어진 직후 부착하는 고루틴. `ctx.Done()` 에서 `session.Close()` 를 부른다. 컨텍스트가 이미 취소된 채로 부착되면 즉시 `Close` 한다.
- **드라이버**: `driveManagedFactorySession(ctx, …)`. 한가한 `select` 와 대기 `select` 에 `ctx.Done()` 을 추가하고, `DeliverTurn` 오류를 받으면 분류보다 먼저 `ctx.Err() != nil` 을 확인해 중단 오류(가칭 `errManagedInterrupted`)를 반환한다.
- **정리 컨텍스트 분리**: launch-pending 롤백은 이미 `context.Background()` 를 쓴다. 정리 호출이 취소된 시그널 컨텍스트를 물려받으면 롤백이 즉시 실패하므로, 정리에는 시그널 컨텍스트를 쓰지 않는다(AC-MH-010이 롤백 행 0건을 단언한다).
- **종료 형태**: 시그널로 끝난 런처는 `interrupted` 를 이름에 담은 중단 오류를 반환한다. 종료 코드는 cobra의 1이며 `128+시그널` 이 아니다 — 후속에서 필요하면 바꾼다.

### D-3.1 수명 순서표 — 이 문서에서 순서의 **유일한 정본**

다른 모든 산출물(spec.md, plan.md, acceptance.md)은 이 표의 번호(O1…O20)와 훅 이름을 가리킬 뿐 순서를 다시 적지 않는다. 용어: **L** = 세션마다 하나인 뮤텍스(Start·Close·DeliverTurn이 읽는 세션 필드 보호), `closed` = L이 지키는 플래그, **게시(published)** = 자원을 L 안에서, `closed` 가 거짓일 때만 세션 필드에 기록하는 것. `Close` 는 **게시된 자원만** 치운다. 게시 시점은 지금 코드에서 `s.started = true` 가 놓인 곳(`cmd.Start` 성공 직후)과 같으므로 게시 여부 = 지금의 `started` 와 동치다 — "`Close` 가 `started` 가 거짓이면 아무것도 안 한다"와 "`Close` 는 게시된 자원을 치운다"는 같은 문장이다(plan-audit N1이 지적한 모순의 해소).

**고속 지역 호출(`MkdirTemp`, 파일 쓰기, 엔드포인트 할당, `cmd.Start`)은 L 안에서** 하고, **느린 호출(준비 폴링, 다이얼, 핸드셰이크 RPC)은 L 밖에서** 한다. L 안의 구간은 `closed` 확인 → 부작용 → 게시가 한 임계구역이라 `Close` 가 그 사이에 끼어들 수 없다(끼어들려는 `Close` 는 L에서 기다린다). 그래서 닫힌 뒤 만들어진 자식이나 토큰 디렉터리는 존재할 수 없고, L 밖 단계마다 닫힘을 다시 확인하는 지점이 O12(연결 기록)에 하나 있다.

| # | 소유자 | 고루틴 | 단계 | L | 시험 훅 단계명 |
|---|---|---|---|---|---|
| O1 | 둘 다 | 소유자 메인 | **시그널 컨텍스트 생성 + 핸들러 설치** — 소유자 진입의 **첫 영속 단계**(launch-pending 등록, 인수 해석, 세션 생성보다 앞). 이 시점에 `defer stop()` 과 "모든 오류 반환을 컨텍스트가 취소돼 있으면 중단 오류로 매핑하는" `defer`(규칙 R-E)를 등록한다(가장 먼저 등록 = 가장 나중에 실행) | — | `handler-installed` |
| O2 | codex | 소유자 메인 | 인수 해석 + `newManagedCodexSession`(프로세스 없음). **세션 생성 직후 워처 부착** | — | `session-constructed` |
| O3 | 둘 다 | 소유자 메인 | `registerFactoryLaunchPending` + launch-pending 롤백 `defer` 등록 (codex 는 O2 뒤, stream 은 O1 바로 뒤) | — | `registered` |
| O4 | stream | 소유자 메인 | `newManagedStreamSession`(`exec.Cmd` 와 파이프 생성, 프로세스 없음). **세션 생성 직후 워처 부착**(O3 뒤) | — | `session-constructed` |
| O5 | 둘 다 | 소유자 메인 | `Start` 진입: **L 획득 → `closed` 확인.** 닫혀 있으면 아무것도 만들지 않고 닫힘 오류(`errManagedSessionClosed`)를 반환 | 획득 | `start-before-lock` (L 획득 **전**) |
| O6 | codex | 소유자 메인 | [L 안] `MkdirTemp`, 토큰 파일 쓰기, 엔드포인트 할당, 준비 컨텍스트와 그 취소 함수 생성. 하나라도 실패하면 오류 반환(지금 동작 그대로, F8 영역) | 유지 | `start-token-written` (토큰 파일 쓰기 직후) |
| O7 | 둘 다 | 소유자 메인 | [L 안] `cmd.Start()`. 실패하면 오류 반환(codex는 지금처럼 토큰 디렉터리 삭제) | 유지 | `start-child-spawned` (`cmd.Start` 성공 직후, 게시 **전**) |
| O8 | 둘 다 | 소유자 메인 | [같은 L 안] **게시**: `cmd`, (codex) `tokenDir`·준비 취소 함수, `started` 를 세션 필드에 기록하고 L 해제 | 해제 | `start-published` (L 해제 직후) |
| O9 | stream | 소유자 메인 | `Start` 가 nil 반환. 이후 `DeliverTurn` 은 L 안에서 `closed` 를 읽고 닫혀 있으면 오류 반환 | — | — |
| O10 | codex | 소유자 메인 | [L 밖] `/readyz` 폴링(준비 컨텍스트). 취소되면 닫힘 오류 반환 | — | — |
| O11 | codex | 소유자 메인 | [L 밖] WS 다이얼 | — | `start-dialed` (다이얼 성공 직후, **기록 전**; 훅이 연결을 받는다) |
| O12 | codex | 소유자 메인 | **L 획득 → `closed` 재확인.** 닫혀 있으면 L 해제 후 **`Start` 가 자기가 다이얼한 연결을 직접 닫고** 닫힘 오류 반환. 열려 있으면 연결을 기록(`client`)하고 L 해제, 읽기 고루틴 시작 | 획득·해제 | `start-conn-recorded` (기록 직후) |
| O13 | codex | 소유자 메인 | [L 밖] 핸드셰이크 RPC(`initialize`, `initialized`, `thread/start`, `thread/name/set`), 준비 컨텍스트로 제한. `Close` 가 연결을 닫으면 호출이 "connection closed" 로 실패 | — | — |
| O14 | codex | 소유자 메인 | `Start` 반환. **정규화**: O10–O13 에서 생긴 오류는 반환 시점에 세션이 닫혀 있으면 닫힘 오류로 돌려준다. 게시 이후 아무 단계도 없는 stream 은 정규화 없이 nil | — | — |
| O15 | 둘 다 | 소유자 메인 | 소유자 진입: `Start` 오류 반환 → 규칙 R-E 로 매핑 | — | — |
| O16 | 둘 다 | 소유자 메인 | 컨텍스트가 취소돼 있으면 중단 오류 반환. 아니면 (codex) `factorymsg.Open` + `BindLaunchPending`, (stream) `factorymsg.Open`. 이어서 우선 턴과 드라이버 루프. 우선 턴 성공 직후 훅 | — | `driver-running` |
| O17 | 둘 다 | 워처 또는 `defer` | `Close` 진입: **L 획득 → `closed` 확인.** 이미 세워져 있으면 L 해제 후 첫 정리가 끝날 때까지 기다렸다가 같은 결과를 반환 | 획득 | — |
| O18 | 둘 다 | 워처 또는 `defer` | [L 안] `closed=true`, **게시된 자원만** 스냅숏(`cmd`, `tokenDir`, `client`, 준비 취소 함수, `started`) 뜨고 L 해제 | 해제 | `close-snapshot` |
| O19 | 둘 다 | 워처 또는 `defer` | [L 밖, 호출당 아니라 **세션당 한 번만**] 준비 취소 함수 호출 · (codex) `client.shutdown`(`done` 닫기 + 연결 닫기) · 자식이 아직 안 끝났으면 kill + `Wait` · (codex) `tokenDir` 제거 · (stream) stdin 닫기, 게시된 자식이 없으면 생성자가 만든 파이프 양 끝 닫기 · 결과 저장 후 완료 신호 | — | `teardown` (정리 본문 시작; 시험이 횟수를 센다) |
| O20 | 둘 다 | 소유자 메인 | 소유자 진입 `defer` 역순: ① `session.Close()`(O17–O19, 이미 했으면 같은 결과) → ② launch-pending 롤백(**`context.Background()`**, 아직 bind 되지 않았을 때) → ③ 열렸던 store 닫기 → ④ 중단 매핑(R-E) → ⑤ `stop()` | — | — |

**규칙 R-E (소유자 진입의 모든 오류 반환)**: O1 이후 소유자 진입이 오류를 반환할 때 시그널 컨텍스트가 취소돼 있으면 그 오류는 항상 `interrupted` 를 이름에 담은 중단 오류로 바뀐다(원래 오류는 원인으로 남는다). `Start` 실패, 우선 턴 실패, 드라이버 반환 모두 같다. 매핑은 O20 ④ 한 곳이다.

**순서의 귀결 (표에서 읽어내는 것 — 새 규칙이 아님)**

| `Close` 가 도착하는 시점 | codex 소유자 | stream 소유자 |
|---|---|---|
| O5 이전(세션만 있고 `Start` 안 함, 또는 O5 의 L 획득 전) | `closed` 만 서고 정리할 게시된 자원이 없다 → nil. 이후 O5 에서 `Start` 거부. 아무것도 만들어지지 않는다 | 같다. 추가로 O19 가 생성자가 만든 파이프 양 끝을 닫는다 |
| O6–O8 도중(L 안) | `Close` 가 L 에서 기다린다. `Start` 가 O8 에서 게시하고 L 을 풀면 `Close` 가 게시된 자식·토큰 디렉터리를 치운다. 게시 뒤 `Start` 는 O10 에서 취소된 준비 컨텍스트를 보고 닫힘 오류 반환 | `Close` 가 L 에서 기다린다. O8 게시 뒤 `Close` 가 자식을 죽이고 `Wait`. `Start` 는 이미 게시했으므로 nil 반환; 이후 `DeliverTurn` 은 오류 |
| O8 이후 O10(준비 대기) 중 | 준비 취소 함수 호출로 폴링이 즉시 풀림 → 닫힘 오류. 자식·토큰 디렉터리는 `Close` 가 치움 | (해당 단계 없음) |
| O11 다이얼 성공 직후, O12 이전 | `Close` 가 게시된 자식·토큰 디렉터리를 치운다(`client` 는 아직 없어 연결은 대상 아님). O12 에서 `closed` 를 보고 `Start` 가 자기 연결을 닫고 닫힘 오류 반환 | (해당 단계 없음) |
| O12 이후 O13 중 | `Close` 가 기록된 `client` 를 `shutdown`. RPC 가 "connection closed" 로 실패 → O14 정규화로 닫힘 오류 | (해당 단계 없음) |
| O14 이후(`Start` 성공 뒤) | 게시된 자원 전부(연결·자식·토큰 디렉터리)를 한 번 치운다. 진행 중 `DeliverTurn` 은 "connection closed" | stdin 닫기 + kill + `Wait`. 진행 중 `DeliverTurn` 은 EOF |
| 두 번째 이후 `Close` | O17 에서 첫 정리가 끝날 때까지 기다리고 같은 결과 | 같다 |

**F8(t1410)과의 경계 — 정확히 무엇이 바뀌고 무엇이 안 바뀌는가**

- 바뀐다: `Start` 의 구조(O5 L 획득과 닫힘 확인, O6–O8 이 한 임계구역, O8 의 게시, O10 의 취소 연결, O12 의 닫힘 재확인), `Close` 의 단일 실행·동시 안전(O17–O19), 소유자 진입의 시그널 컨텍스트와 매핑(O1, O20).
- **바뀌지 않는다**: O6–O7 구간의 **오류 반환 분기**. 즉 세션이 **열려 있는데** `MkdirTemp` 뒤 토큰 파일 쓰기 또는 엔드포인트 할당이 실패하면 토큰 디렉터리가 남는 지금 동작(F8)은 그대로다 — 그 경우 아무것도 게시되지 않았으므로 deferred `Close` 는 지금처럼 아무것도 치우지 않는다(게시 ⇔ 지금의 `started`). t1410 이 그 분기에서 토큰 디렉터리를 지우는 것으로 F8 을 닫는다. `cmd.Start` 실패 때 토큰 디렉터리를 지우는 지금 코드도 그대로다. 핸드셰이크 예산 값(F9)과 `/readyz` 리디렉션 재검사(F13)도 바꾸지 않는다.
- 같은 함수 `Start()` 의 같은 줄 영역을 두 카드가 만지므로 병합 충돌 면적이 있다(t1410 이 이 카드 뒤에 실행). 이 카드는 그 구간을 마일스톤 M4 한 커밋에 몰아 둔다.
- 토큰 디렉터리는 O8 에서야 세션 필드에 기록된다(그 전에는 지역 변수, `Start` 소유). 그래서 "`Close` 가 기록 전의 토큰 디렉터리를 치우거나 못 치우는" 창이 없다.

### 시험 이음새 — 매개변수화된 훅 하나

`managedStepHook func(ctx context.Context, step string, res any)` 비공개 패키지 변수(프로덕션에서는 nil, 확인 비용은 nil 비교 한 번). 위 표의 "시험 훅 단계명" 열이 호출 지점이다. 훅은 **블로킹 가능**하며, `res` 는 그 단계가 가진 자원을 준다(`start-dialed` 에서는 `*websocket.Conn`). 시험은 같은 변수를 설정·복구하고 병렬로 돌리지 않는다. 시그널 재실행 도우미 프로세스는 이 훅으로 단계 도달을 stdout 한 줄(`ready <단계명>`)로 알리고, 지정한 단계에서는 `ctx.Done()` 까지 대기한다 — 부모 시험은 그 줄을 본 뒤에만 시그널을 보낸다(핸들러 설치 전에 신호가 가는 흔들림이 없다).

### 공시한 한계 (이 SPEC이 닫지 않음)

- **연결 쓰기 데드라인 없음, 막힌 쓰기의 상한도 없음**: `call()` 은 `c.conn.WriteJSON(…)` 을 `select`(턴 컨텍스트·`done` 대기) **앞에서** 호출한다(`managed_codex_factory.go:212` 대 `:215`, 소스 판독). 그래서 서버가 읽지 않아 쓰기가 막히면 턴 타임아웃(10분)도 그것을 풀지 못한다. 쓰기 뮤텍스가 들어오면 읽기 고루틴의 답장 쓰기도 같은 뮤텍스에 걸려 함께 멈춘다. 이 한계의 상한은 "턴 타임아웃"이 아니라 **`Close`(시그널) 또는 연결 종료가 올 때까지**다. 시그널 경로는 영향이 없다 — `Close` → `shutdown` 이 뮤텍스 없이 `conn.Close()` 를 부르므로 막힌 쓰기도 풀린다. 쓰기 데드라인 상수는 이 카드에서 더하지 않기로 판단했다: 상대는 우리가 소유한 loopback 자식이고 관측된 사례가 없으며, 이 한계는 운영자 문서에 "상한 없음"으로 적는다. 데드라인이 필요해지면 `defaults.go` 상수 하나와 REQ-MH-005 개정으로 닫는다(후속 후보). 이 단락은 소스 판독이며 실행으로 재현하지 않았다. `id: null` 프레임은 응답·요청 어느 쪽으로도 분류되지 않아 버려진다.
- **두 번째 시그널**: 첫 시그널에서 기본 동작을 복원하므로 정리 중 두 번째 시그널(SIGHUP 직후 SIGTERM 등)은 프로세스를 기본 동작으로 끝내 F5의 원래 상태를 재현할 수 있다. 정리가 멈췄을 때의 탈출구로 의도한 것이며 자동 시험으로 고정하지 않는다.
- **후손 프로세스**: 스트림 소유자의 `Kill` 은 직접 자식만 죽이고 후손은 남을 수 있다. AC-MH-010은 소유한 자식의 부재만 단언한다.
- **bind 된 행은 남는다**: O16 에서 `BindLaunchPending` 한 뒤 시그널로 끝나면 bound 행은 기존의 모든 종료와 같이 남는다(부모 edge: owner 소멸 → 스테일 엔드포인트). 이 카드가 바꾸지 않는다.

### `syscall` 참조와 부모 AC-MS-014

부모 AC-MS-014는 `grep -rn 'syscall\.' internal/cli/managed_*.go` 가 0행이어야 한다고 정한다. `SIGTERM`·`SIGHUP` 은 `syscall` 상수라 이 SPEC은 참조를 `managed_*` 가 아닌 파일(`launch_signals.go`)에 둔다. 이것은 **의도적 배치**이며 부모 AC의 목적(자식 프로세스 소유 모델, 프로세스 교체 exec·플랫폼 분기 없음, Windows 크로스 빌드)을 해치지 않는다는 판단에 기댄다 — 이 참조는 상수뿐이고 어떤 시스템 호출도 하지 않는다. 이것은 부모 REQ-MS-012 문면("zero `syscall` usage in newly added files")에 대한 명시적 예외이고, 같은 cross-platform exemption(EXCL-syscall) 선언이 spec.md C.3·§D 에 있다. 그 판단을 검증 가능하게 만들려고 두 가지를 AC에 건다: `managed_*.go` 의 0행이 유지된다(부모 AC-MS-014 grep 그대로), 그리고 `launch_signals.go` 의 `syscall.` 참조가 상수뿐이다(AC-MH-011).

**Windows 실측**: 별도 임시 모듈에서 `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)` 를 `GOOS=windows GOARCH=amd64 go build` 했을 때 exit 0이었다(go1.26.0). 기존 선례로 `internal/cli/mcp_server.go:123` 가 `os.Interrupt, syscall.SIGTERM` 을 같은 방식으로 쓴다. 참고로 `internal/cli/codex_job_control.go:88-95` 의 주석은 "`syscall.SIGTERM` does not exist on windows" 라고 쓰는데 위 실측과 어긋난다 — 이 SPEC은 그 주석을 건드리지 않는다. Windows에서는 런타임이 `os.Interrupt`(Ctrl-C/Ctrl-Break)만 전달하므로 SIGTERM·SIGHUP 구독은 컴파일되지만 발화하지 않는다. 콘솔 닫기 이벤트는 처리하지 않는다(범위 밖, 한계로 공시).

### 시험 설계 요약

순서의 관측은 acceptance.md 가 O번호·훅 이름으로 묶는다(AC-MH-010 시그널 종단, AC-MH-016 수명주기). 순서를 여기서 되풀이하지 않는다.

- 단위(AC-MH-009): 가짜 세션이 `DeliverTurn` 안에서 `Close` 가 불릴 때까지 막히게 하고, 컨텍스트 취소 → 워처의 `Close` → 반환을 단언한다(한가한 경우와 진행 중인 경우 둘 다).
- 종단 시그널(POSIX, AC-MH-010): 시험 바이너리를 `os/exec` 로 다시 기동해(재실행 선례: `TestManagedCodexFakeAppServer`) 소유자를 돌리고, 부모 시험이 지정 훅 단계의 `ready` 줄을 본 뒤 SIGTERM·SIGHUP·SIGINT를 `os.Process.Signal` 로 보낸다. **하네스 함정(실측)**: 셸의 `&` 로 띄운 백그라운드 프로세스는 SIGINT를 무시한 채 상속되어 `kill -INT` 가 듣지 않는다. 그래서 반드시 `os/exec` 로 띄운다. Windows에서는 `runtime.GOOS == "windows"` 로 건너뛰고(이유: POSIX 시그널 전달 시험) 빌드는 AC-MH-011이 증명한다.

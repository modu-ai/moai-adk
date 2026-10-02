---
id: SPEC-FACTORY-MANAGED-HARDEN-001
title: "design.md — 관리 세션 소유자 강건화 설계 결정"
version: "0.1.0"
created: 2026-10-03
updated: 2026-10-03
author: GOOS (manager-spec)
tier: M
---

# design.md — 설계 결정 D-1..D-3

> 상태축 없음(spec-frontmatter-schema.md § Artifact Statelessness). 생명주기는 `spec.md` 만 운반한다. 이 문서의 식별자(함수·필드 이름)는 설명을 위한 가칭이며 run 단계가 바꿀 수 있다. **관측 등급 표기**: "실측" = 이 plan 실행에서 명령을 돌려 본 것, "소스 판독" = 코드를 읽고 추론한 것.

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

**미관측(Gap)**: 실제 codex가 각 거부·오류 응답을 받았을 때 모델이 어떻게 행동하는지는 관측하지 못했다(라이브 게이트 미설정). 특히 `mcpServer/elicitation/request` 를 거부하는 선택은, codex가 MoAI MCP 도구 승인까지 이 요청으로 올리는 경우 브로커 도구 호출을 막을 수 있다. 부모 SPEC의 승인 스코핑(`approval_mode="approve"`)이 그 승인창을 없앤다는 부모 쪽 관측에 기대는 판단이며, 이 SPEC은 그 관측을 다시 하지 않는다. 라이브에서 수신 확인이 멈추는 것이 보이면 정책표의 해당 줄을 재검토한다.

### 기각한 대안

- (a) 이벤트 채널로 DeliverTurn에 넘겨 거기서 답한다 — 위 이유 1·2.
- (b) 승인류에 `cancel`/`abort` 를 쓴다 — 턴을 즉시 끊어 모델이 수신 확인을 쓸 기회를 없앤다.
- (c) 사전 승인과 겹치는 요청은 accept 한다 — 요청 종류만으로는 어떤 도구 승인인지 안전하게 구별할 수 없다(`mcpServer/elicitation/request` 의 `serverName` 으로 구별하려면 별도 설계가 필요하다). 최소 권한이 기본이다.
- (d) 모르는 method를 무시한다 — 서버가 응답을 영원히 기다릴 수 있다(현재 결함의 본질).

## D-2 — 오류 분류, 연속 실패 상한, claim 처분 (F4)

### 결정 1 — 분류는 "턴 단위로 명시한 것만 격리, 나머지는 세션 치명"

드라이버는 지금 `DeliverTurn` 오류를 모두 반환한다. **기본값을 그대로 둔다(실패에 닫힌 쪽)**: 오류는 소유자가 "이 턴만의 실패"라고 표시했을 때만 격리하고, 표시되지 않은 오류는 전부 세션 치명이다. 표시는 오류를 감싸는 한 가지 표식 오류(가칭 `errManagedTurnFailed`, `errors.Is` 로 판별)다. 두 소유자가 아래 두 곳에서만 표시한다.

| 실패 | 분류 | 근거 |
|---|---|---|
| 스트림 `result.is_error` (`pumpManagedStreamTurn`) | **턴 단위** | 세션이 살아 있고 결과 이벤트가 정상 도착했다. 일시적 429 같은 API 오류가 이 모양이다(감사서 §F4). |
| Codex 턴이 `completed` 가 아닌 상태로 종료(`turn ended as interrupted/failed`) | **턴 단위** | 서버가 `turn/completed` 를 정상 보고했다. 연결은 살아 있다. |
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
| 수정 지점(실측: `grep -c` 기준 텍스트 출현 수) | `DeliverTurn(` 13곳(6개 파일) + 구현체 2종 + 가짜 `fakeManagedSession` 1종 + `driveManagedFactorySession(` 10곳(5개 파일) | `driveManagedFactorySession(` 10곳(5개 파일)에 `ctx` 첫 인자 추가, 인터페이스와 가짜 구현체는 무변경 | 정리를 한 번 더 구현 |
| 정리 경로 수 | 1(기존) | 1(기존) — `Close`·`defer` 그대로 | 2(`defer` 는 `os.Exit` 에서 돌지 않으므로 launch-pending 롤백·토큰 삭제를 다시 써야 한다) |
| 새 위험 | 인터페이스 확장이 이후 소유자 모두에 강제된다 | `Close` 가 동시 호출 안전해야 한다(아래) | 정리가 두 벌로 갈라진다 |

**선택 B.** 이유: (1) 스트림 소유자에서 A도 결국 자식을 죽여야 하므로 A는 B보다 작지 않다. (2) A는 인터페이스를 바꿔 가짜 구현체와 이후 소유자 전부에 파급되고, B는 드라이버 호출부 기계적 수정이 전부다. (3) B는 기존 정리 경로(`Close` + `defer`)를 그대로 쓰므로 시그널 때문에 정리가 갈라지지 않는다.

### B의 구성 (가칭)

- **시그널 구독**: `signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)` 를 감싼 작은 도우미를 `internal/cli/launch_signals.go` 에 둔다. 첫 시그널이 오면 `stop()` 으로 기본 동작을 복원한다 — 정리가 멈췄을 때 두 번째 시그널이 프로세스를 끝낼 수 있게 하려는 것이다.
- **워처**: 소유자 진입(`runManagedFactoryStreamSession`, `runManagedFactoryCodex`)이 컨텍스트를 세션 생성 직후, `Start()` 이전에 만들고 워처 고루틴이 `ctx.Done()` 에서 `session.Close()` 를 부른다. 워처가 `Start()` 중에 도달해도 안전하려면 Codex 소유자의 `Start()` 가 준비 대기에 쓰는 컨텍스트를 이 컨텍스트에서 파생한다(필드로 주입; 인터페이스 불변). 그렇지 않으면 핸드셰이크 예산(10초) 동안 시그널이 기다린다.
- **드라이버**: `driveManagedFactorySession(ctx, …)`. 한가한 `select` 와 대기 `select` 에 `ctx.Done()` 을 추가하고, `DeliverTurn` 오류를 받으면 분류보다 먼저 `ctx.Err() != nil` 을 확인해 중단 오류(가칭 `errManagedInterrupted`)를 반환한다.
- **`Close` 의 동시 안전**: 지금 `closed` 는 보호되지 않은 불리언이고 `cmd.Wait()` 는 한 번만 부를 수 있다. 시그널 워처와 `defer` 가 겹치면 데이터 경합과 이중 `Wait` 이 생긴다. 두 소유자의 `Close` 를 한 번만 도는 정리(예: `sync.Once` 에 결과 오류를 저장)로 바꾸고 동시 호출자는 첫 정리가 끝날 때까지 기다려 같은 결과를 받는다.
- **정리 컨텍스트 분리**: launch-pending 롤백은 이미 `context.Background()` 를 쓴다. 정리 호출이 취소된 시그널 컨텍스트를 물려받으면 롤백이 즉시 실패하므로, 정리에는 시그널 컨텍스트를 쓰지 않는다(AC-MH-010이 롤백 행 0건을 단언한다).
- **종료 형태**: 시그널로 끝난 런처는 중단 오류(시그널이 정리를 일으켰음을 이름에 담은 오류)를 반환한다. 종료 코드는 cobra의 1이며 `128+시그널` 이 아니다 — 후속에서 필요하면 바꾼다.

### `syscall` 참조와 부모 AC-MS-014

부모 AC-MS-014는 `grep -rn 'syscall\.' internal/cli/managed_*.go` 가 0행이어야 한다고 정한다. `SIGTERM`·`SIGHUP` 은 `syscall` 상수라 이 SPEC은 참조를 `managed_*` 가 아닌 파일(`launch_signals.go`)에 둔다. 이것은 **의도적 배치**이며 부모 AC의 목적(자식 프로세스 소유 모델, 프로세스 교체 exec·플랫폼 분기 없음, Windows 크로스 빌드)을 해치지 않는다는 판단에 기댄다 — 이 참조는 상수뿐이고 어떤 시스템 호출도 하지 않는다. 그 판단을 검증 가능하게 만들려고 두 가지를 AC에 건다: `managed_*.go` 의 0행이 유지된다(부모 AC-MS-014 grep 그대로), 그리고 `launch_signals.go` 의 `syscall.` 참조가 상수뿐이다(AC-MH-011).

**Windows 실측**: 별도 임시 모듈에서 `signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)` 를 `GOOS=windows GOARCH=amd64 go build` 했을 때 exit 0이었다(go1.26.0). 기존 선례로 `internal/cli/mcp_server.go:123` 가 `os.Interrupt, syscall.SIGTERM` 을 같은 방식으로 쓴다. 참고로 `internal/cli/codex_job_control.go:88-95` 의 주석은 "`syscall.SIGTERM` does not exist on windows" 라고 쓰는데 위 실측과 어긋난다 — 이 SPEC은 그 주석을 건드리지 않는다. Windows에서는 런타임이 `os.Interrupt`(Ctrl-C/Ctrl-Break)만 전달하므로 SIGTERM·SIGHUP 구독은 컴파일되지만 발화하지 않는다. 콘솔 닫기 이벤트는 처리하지 않는다(범위 밖, 한계로 공시).

### 시그널 시험 설계

- 단위: 가짜 세션이 `DeliverTurn` 안에서 `Close` 가 불릴 때까지 막히게 하고, 컨텍스트 취소 → 워처의 `Close` → 반환을 단언한다(한가한 경우와 진행 중인 경우 둘 다).
- 종단(POSIX): 시험 바이너리를 `os/exec` 로 다시 기동해(재실행 선례: `TestManagedCodexFakeAppServer`) 소유자를 돌리고, 부모 시험이 SIGTERM·SIGHUP·SIGINT를 `os.Process.Signal` 로 보낸 뒤 5초 안 종료, 자식 프로세스 부재, 토큰 디렉터리 삭제, launch-pending 행 0건을 단언한다. **하네스 함정(실측)**: 셸의 `&` 로 띄운 백그라운드 프로세스는 SIGINT를 무시한 채 상속되어 `kill -INT` 가 듣지 않는다. 그래서 반드시 `os/exec` 로 띄운다. Windows에서는 `runtime.GOOS == "windows"` 로 건너뛰고(이유: POSIX 시그널 전달 시험) 빌드는 AC-MH-011이 증명한다.

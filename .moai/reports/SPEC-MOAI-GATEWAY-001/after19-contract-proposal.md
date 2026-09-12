# 19시 이후 관측에 따른 최소 계약 보강 제안

이 문서는 최초 제안과 후속 위임에 따른 반영 기록을 함께 보존한다. 최초 제안에서는 SPEC을 변경하지 않았고,
이후 부모가 M0·엄격한 SSE·실제 picker 사실 보강과 hash-only receipt 설계 후보의 본문 반영을 명시 재위임했다.
그 반영 결과는 끝의 후속 검증 기록에 있다. 출력 상한·Windows 의미 변경은 여전히 승인 대기다. 제품 코드는 소유하지
않으며 원래 목표인 Claude Code 안에서 네 GPT의 대화·도구·재개 사용을 줄이지 않는다.

## Claim

M0의 별도 세션 헤더 선택은 이미 승인된 측정 분기를 확정하는 사실 보강이다. 반면 구독 출력 상한 생략, no-tool reasoning 유실 처리, Windows 저장 성공의 내구성 수준은 관측만으로 확정할 수 없는 계약 문제다. SSE의 헤더 부재와 빈 completed.output은 제한된 구독 경로의 형식 차이로 명시하되, 형식 검증·충돌 거절·기존 음성 AC를 보존하는 보강을 제안한다.

### 1. 변경 위치와 결정의 성격

| 대상 | 현재 계약 | 최소 변경 | 성격 |
|---|---|---|---|
| 코어 REQ-MG-016, design §3.2, plan M0, AC-MG-021 | 운반 키를 M0로 결정, 정상 응답·refresh 확인까지 차단 | X-MoAI-Session-Token 확정과 두 관측의 증거 연결 | 승인된 측정 분기 확정; 제품 통합 PASS와 구분 |
| 코어 REQ-MG-012·013, design §4.3, AC-MG-007·008 | 오류를 성공으로 바꾸지 않고 index·ID·순서 보존 | 구독 SSE 헤더 부재 및 item.done 완료 자료의 수용 조건 명시 | 제한된 호환 형식 보강; malformed 허용 아님 |
| 코어 design §4.3, AC-MG-007·008, AUTH AC-GA-008 | max_tokens를 max_output_tokens로 전달 | 구독 경로만 별도 출력 정책 여부 결정 | **사용자 답변 대기 중인 의미 변경** |
| 코어 REQ-MG-013·015, design §4.3, AC-MG-009 | opaque는 PROBE ONLY, 전체 유실 binding 미결 | 가역 tool ID와 ordered envelope의 결합, no-tool 경계 별도 명시 | 후보 채택 및 미결 계약 해소가 필요; 즉시 활성화 불가 |
| AUTH REQ-GA-004, plan B.1·primitive 문단, AC-GA-004·005 | 파일·디렉터리 내구성 확인, atomicfile.Replace 지정 | 플랫폼별 성공 조건과 Windows 최종 교체 경로 명시 | 구현 경로 선택과 내구성 보장 변경을 구별 |
| PICKER REQ-GP-003·005·006·007 및 대응 AC | 재개·설정 격리·표시/요청 일치·실제 왕복 | s/Enter와 Default 관측을 근거에 추가 | 요구사항 완화 없음; 재시작·실계정 AC 유지 |

새 REQ·AC 번호를 발급하거나 기존 AC를 삭제할 필요는 없다. 해당 번호의 조건과 대조군을 보강하고 design·plan에 운반 방식과 남은 게이트를 명시한다.

### 2. M0: 그대로 옮길 수 있는 제안 문구

**design §3.2의 미결 운반 키 문단 대체안**

> 세션 접근 토큰은 별도 `X-MoAI-Session-Token` 헤더로 운반한다. 상속된 `ANTHROPIC_AUTH_TOKEN`은 기존 정리 단계에서 제거하고 세션 토큰을 그 키에 다시 넣지 않는다. gateway는 로컬 인증 후 이 별도 헤더를 제거하며 어떤 upstream에도 전달하지 않는다. Claude가 관리하는 OAuth Bearer와 허용 beta 헤더는 고정 Anthropic endpoint에만 전달한다. 자식 custom-header 구성에서 동일 세션 헤더의 중복·외부 값 충돌은 거절하고 임의 병합으로 인증을 모호하게 만들지 않는다.
>
> 2026-09-11의 Opus 5·Sonnet 5 정상 응답과 Opus 5의 직접 refresh POST 200→같은 본문 재전송 200 관측을 M0의 양성 근거로 사용한다. 종전 INCONCLUSIVE 결과를 덮어쓰지 않는다. 이 결정은 passthrough 구현의 선행 게이트를 해소하며 제품 gateway의 수명·stream·헤더 차단·도구·재개 수용을 대신하지 않는다. 갱신 credential을 MoAI가 복사하거나 관리하도록 인증 소유권을 변경하지 않는다.

**AC-MG-021 보강안**

> Given 양성 M0 근거와 제품 gateway가 있을 때, When Opus 5·Sonnet 5 요청 및 인증 실패 뒤 재요청을 실행하면, Then 로컬 세션 헤더의 검증과 upstream 수신 0, 허용 OAuth/beta 보존, 같은 본문의 정상 응답을 각각 판정한다. 직접 refresh의 발급 출처는 token POST 응답과 재사용 Bearer의 hash 상관관계로 판정하고 단순히 Bearer가 달라졌다는 사실로 대체하지 않는다. 잘못되거나 중복인 세션 헤더는 외부 송신 0이어야 한다.

REQ-MG-016 자체의 측정 게이트를 제거하지 않는다. CG·TEAMMATE 인계에는 선택된 키와 상속 AUTH_TOKEN 부재를 전달한다. 실제 product test에서 refresh를 매번 강제해야 한다는 새 요구를 만드는 대신 선행 refresh 증거와 통합 헤더 시험의 귀속을 분리한다.

### 3. 구독 SSE: 제한된 수용과 엄격한 종료

아래는 `auth-responses-runtime-observation.md`를 읽고 작성한 초안이다. 작성자는 해당 live 요청을 직접 실행하지 않았다. 보고서는 네 모델 함수 왕복의 완료, Content-Type 부재, 빈 terminal output을 구조화 결과로 기록한다. 이를 제품 통합 PASS로 확대하지 않는다.

**REQ-MG-012·013의 보강 문구**

> Where 검증된 GPT 구독 endpoint의 streaming 응답인 경우, the gateway shall Content-Type이 없는 HTTP 200 응답을 제한된 UTF-8 SSE 형식으로 검증할 수 있어야 한다. 헤더가 없다는 이유로 JSON·HTML·임의 바이트를 성공 응답으로 처리해서는 안 된다. 명시적으로 충돌하는 media type, 지원하지 않는 content encoding, 잘못된 UTF-8·JSON·이벤트 구조·초과 길이는 명시 오류로 끝내야 한다. 이 예외는 API 키 경로나 다른 provider의 media 계약으로 확장하지 않는다.
>
> When response.completed의 output 배열이 비어 있고 모든 출력 item이 앞선 완료 이벤트로 확정되어 있으면, the gateway shall 검증된 output_item.done의 item을 index 순서로 사용해야 한다. 각 item의 ID·type·index·완료 상태와 text/tool 인수는 앞선 added/delta/done 자료와 일치해야 한다. 서로 다른 item의 동일 index 또는 ID, 중복 완료·중복 terminal, 모순된 payload, 미완료 item, terminal 없는 EOF를 정상 완료로 판정해서는 안 된다.

**design §4.3의 정확한 우선순위 보강**

- `output_item.done`은 해당 item의 완료 자료다. `response.completed.output=[]`은 그 자료를 지우라는 뜻으로 해석하지 않는다. 반대로 빈 배열만으로 완료 item을 만들어서는 안 된다.
- `response.completed.output`이 비어 있지 않으면 이미 수신한 완료 item과 순서·ID·type·내용이 일치해야 한다. 어느 한쪽을 임의로 우선하여 충돌을 숨기지 않는다. 최종 snapshot만으로 앞선 delta 누락이나 미완료 블록을 보완해 성공시키는 새 예외를 만들지 않는다.
- item 완료만으로 전체 성공을 보내지 않는다. 전체 terminal과 기존 종료 표의 상태가 함께 유효해야 한다. 이미 전달한 byte 뒤 오류가 나면 재시도 없이 오류를 드러낸다.
- 무시할 수 있는 비출력 이벤트의 목록과 SSE comment/공백 처리는 명시적인 형식 규칙으로 둔다. 알 수 없는 출력 또는 종료 이벤트를 무시하여 잘림을 성공으로 바꾸지 않는다.
- 동시 tool call은 원래 output index 순서를 보존한다. 도착 순서를 전체 출력 순서로 대체하지 않는다. 공개 call ID의 의미는 다음 절의 가역 wire encoding과 구별한다.

**AC-MG-007·008에 추가할 대조군**

> Given 동일한 완료 text/tool 이벤트 본문에 SSE media type이 있는 응답과 없는 구독 응답이 있을 때, When 중계하면, Then 둘의 공개 출력·tool pair·순서·종료가 같다. 완료 item이 있는 빈 completed.output은 성공하고, 비어 있지 않은 일치 snapshot도 성공한다. header 없음+HTML/JSON, 충돌 media type, 중복 JSON field, ID/index 충돌, done 중복, 모순 snapshot, 미완료/잘린 이벤트, failed·다른 incomplete·취소의 각 변형은 성공 terminal 0이어야 한다. 기존 max-output 제한 대조군은 삭제하지 않는다.

HTTP 200 및 EOF 자체는 성공 조건이 아니다. nonstream 지원까지 이 조항에서 새로 선언하지 않는다.

### 4. 출력 상한: 답변 전 적용 금지

현재 design §4.3은 `max_tokens → max_output_tokens`를 명시한다. dispatch가 인계한 실제 Astra 응답 분류는 HTTP 400, `unsupported_parameter/max_output_tokens`다. 이 사실은 조용히 필드를 빼도 된다는 승인이 아니다.

사용자에게 이미 물은 결정의 영향은 다음처럼 표현한다.

> GPT 구독 경로에서 Claude의 요청별 생성 토큰 상한을 upstream에 같은 의미로 전달할 수 없으므로, 구독 서버의 출력 정책을 사용하도록 예외를 승인할 것인가? 승인한다면 사용자에게 해당 상한이 강제되지 않음을 명시하고 구독 경로에서만 필드를 생략한다. 입력 추정치·응답 바이트 제한·연결 취소를 생성 토큰 상한과 같다고 표시하지 않는다. API 키 경로의 기존 매핑과 max-output incomplete 종료 시험은 보존한다.

답변 전에는 현재 bound 계약을 유지한다. 승인된다면 REQ-MG-017의 provider별 지원 정책, design §4.3의 매핑 문장, AC-GA-008의 실제 지원 필드, AC-MG-007·008의 경로별 대조군을 함께 고쳐야 한다. 승인되지 않으면 실제로 지원되는 동등한 상한 수단을 입증할 때까지 이 경로를 미완료로 남긴다. 관련 목표나 네 모델을 제외하는 대안으로 닫지 않는다.

### 5. opaque와 공개 tool ID: 채택 가능한 부분과 남은 부분

**REQ-MG-013·015 보강 후보**

> Where 같은 OpenAI provider로 대화가 이어지는 경우, the gateway shall 반환받은 opaque item의 원래 순서와 바이트를 보존해야 한다. 공개 tool call과 함께 반환된 경우에는 공개 call ID를 가역적으로 운반하고 그 assistant 응답의 ordered opaque 묶음과 결합해야 한다. 다음 입력에서 결합이 깨지거나 필수 묶음의 전체·일부가 없으면 외부 송신 전에 명시 오류로 끝내야 한다. 다른 provider로 전환할 때에는 opaque를 전달하지 않되 공개 text와 완료된 tool pair의 순서 및 원래 ID 대응을 보존해야 한다.

**design §4.3의 최소 운반 방식 후보**

1. `redacted_thinking.data`에 버전·OpenAI provider·ordered item을 담은 MoAI envelope를 둔다. 이것을 Anthropic 서명 또는 공개 reasoning으로 해석하지 않는다. opaque item은 해독하지 않는다.
2. 도구 응답의 wire tool ID는 버전 있는 인코딩으로 원래 call ID와 해당 ordered envelope bytes의 SHA-256을 담는다. 현재 측정 후보는 `toolu_moai_v1_` 접두어와 base64url compact JSON의 `call_id`, `opaque_sha256`다. 원래 ID로 복원한 tool_use/tool_result 대응을 기존 ID 보존 AC의 기준으로 삼는다. 인코딩된 문자열 자체가 OpenAI call ID인 것처럼 송신하지 않는다.
3. decoder는 버전·provider·필수 필드·중복 필드·허용 문자·길이·item type·ID·hash·배열 순서를 검증한다. unmarked 일반 ID를 prefix 유사성만으로 해석하지 않는다. 정확한 예약 형식이 깨졌으면 거절한다. 서로 다른 assistant 묶음/요청의 hash나 ID를 교차 사용하지 않는다.
4. hash는 유실·변조의 일관성 검사다. 서명·비밀·공급자 인증이 아니다. 공격자가 envelope와 marker를 함께 임의로 바꾼 모든 경우를 이 hash 하나로 탐지한다고 주장하지 않는다.
5. 합성 68바이트 carrier와 172바이트 ID의 실제 Read 왕복은 형식 호환의 한 양성 대조군이다. 실제 encrypted item 크기·네 모델·resume·foreign strip·삭제 mutant는 별도 게이트다. 최대 길이는 한 번 관측한 172로 일반화하지 않고 제품 한도와 실제 client 수용 시험으로 확정한다.

**no-tool 응답의 정확한 경계**

tool call이 없는 응답에는 위 공개 tool ID 표식이 없다. 그러므로 envelope 블록 전체가 사라졌을 때 stateless decoder가 그 응답을 원래 opaque가 없던 응답과 구별한다는 주장을 하지 않는다. `auth-responses-runtime-observation.md`에서 실제 Luna no-tool 응답의 reasoning→message, encrypted 1400바이트, 원래 compact item bytes를 보존한 후속 요청의 HTTP 200을 읽었다. 삭제 mutant는 실행하지 않았으며 서버가 항목을 필수로 요구하는지는 미판정이다. 이 불확실성을 보존 의무를 없애는 이유로 사용하지 않는다.

no-tool 보존의 양성 근거가 있으므로 기존 대안 (b)의 대화 한정 hash receipt를 구체적인 최소 후보로 제안한다. 삭제 후 서버 200만으로 의미·추론 상태가 같거나 필수가 아니라고 판정하지 않는다. 임의의 합성 tool call을 만들어 no-tool 응답을 tool 응답으로 바꾸지 않는다.

필수 유실 감지를 유지하는 추가 설계 후보는 대화 한정 lineage다. 이 경우 다음 동작을 먼저 계약해야 한다.

> When opaque가 있는 no-tool 응답을 내보낸 경우, the gateway shall 그 대화에 필요한 opaque 묶음의 존재·순서 검증 정보를 해당 대화에만 귀속해야 한다. 같은 대화의 후속 요청·continue·resume에서는 그 요구와 실제 history를 검증하고 유실·불명확한 귀속은 송신 전에 거절해야 한다. 다른 대화·동시 요청·새 세션에 최근 상태를 전용해서는 안 된다. 기존 사용자 credential이나 전체 대화의 임의 복사로 귀속을 만들지 않는다.

이 후보는 아직 구현 지시가 아니다. 다만 어려움을 새 사용자 차단 질문으로 돌리지 않고 아래 방식으로 기존 대안 (b)의 구현·검증 조건을 제안한다. no-tool 유실 방지를 삭제하는 의미 변경은 포함하지 않는다.

**대화 한정 hash receipt의 구체 설계안**

1. **귀속 키.** launcher가 인증한 gateway 세션과 native 대화 UUID의 조합을 사용한다. 부모는 native request009의 `metadata.user_id` JSON 안에 `session_id`가 있음을 직접 읽었다고 인계했다. 임의 요청이 UUID를 주장하는 것만으로 다른 대화 receipt를 열 수 없으며, launcher가 새 대화 또는 명시 continue/resume으로 허용한 UUID에만 바인딩한다. 새 gateway 프로세스의 임의 세션 토큰 값 자체를 영속 키로 사용하지 않는다. 재개 시 같은 native UUID가 보존되는지는 진행 중인 실제 두 프로세스 probe의 수용 조건이다. UUID가 달라지는 재개 경로라면 launcher의 명시적 재개 대상과 연결된 alias만 허용하며 내용 유사성으로 자동 병합하지 않는다.
2. **공개 history 정규화.** 메모리 안에서만 공개 message의 순서·role·text Unicode 문자열·이미지 등 지원 content·완료 tool_use/tool_result의 원래 ID·이름·인수·결과·오류 flag를 구조화한다. JSON object key 순서와 공백, content 문자열과 같은 문자열의 단일 text block 표현만 정규화하며 block의 `cache_control`은 전송 힌트로서 digest에서 제외한다. 배열 순서·문자열·숫자 의미는 보존한다. tool ID는 엄격하게 검증한 뒤 원래 ID로 가역 복원하고 opaque/thinking 블록은 이 공개 projection에서 제외한다. top-level system/metadata/출력 정책은 공개 history digest에 넣지 않되 실제 upstream 요청의 처리 계약은 그대로 유지한다. message 안의 system과 native 알림 본문은 임의 삭제하지 않는다. 실제 클라이언트가 바꾸는 요소를 별도 측정하지 않은 상태에서 ‘안정화’를 이유로 광범위하게 제외하지 않는다. 아래 실제 resume raw 비교에서 이 최소 정규화 뒤 이전 5개 message prefix가 동일했다.
3. **prefix receipt.** 공개 history의 각 assistant 응답 경계까지의 정규화 바이트를 SHA-256으로 계산한다. receipt에는 버전, 대화 UUID의 digest, 공개 prefix digest, 앞선 공개 prefix digest, 그 응답의 ordered opaque envelope digest·item 수, provider 및 완료 상태만 저장한다. token·opaque 원문·공개 대화·tool 인수/결과는 저장하지 않는다. opaque 없는 응답의 receipt는 `opaque_required=false`로 명확히 구분한다. 요청 입력 prefix와 응답을 결합하므로 이전 대화와 무관한 같은 짧은 답변으로 일치시키지 않는다.
4. **송신 전 검사.** incoming history를 동일 규칙으로 계산하여 모든 해당 assistant prefix의 receipt와 실제 envelope를 비교한다. receipt가 opaque를 요구하는 prefix에서 envelope가 없거나 digest·순서·item 수가 다르면 송신 0이다. tool marker와 envelope를 함께 지워도 원래 ID로 정규화한 공개 prefix가 같으므로 receipt가 검출한다. no-tool 전체 envelope 삭제도 같은 원리다. 암호화 원문을 receipt에서 재생하거나 임의 복원하지 않는다.
5. **기록 시점과 실패.** 응답 item이 검증되어 prefix/envelope digest가 확정되면 전용 잠금 아래 receipt를 원자·durable publish하고 난 뒤 성공 terminal을 내보낸다. 그 이전에 일부 stream bytes를 보냈더라도 receipt 저장이 실패하면 성공 terminal을 보내지 않으며 재시도하지 않는다. 실패·취소·불완전 item은 완료 receipt를 만들지 않는다. 완료 receipt 뒤 client가 응답을 받지 못한 경우는 기록을 삭제하거나 다음 요청에 강제 삽입하지 않는다. 이는 수신 확인이 아닌 내보낼 완료 응답의 검증 기록이다.
6. **병렬·중복.** 같은 입력 prefix에서 병렬로 나온 서로 다른 공개 응답은 각각 별도의 prefix digest에 기록한다. 동일 공개 prefix로 서로 다른 opaque가 나온 경우 단일 최신 값으로 덮어쓰지 않고 유한 후보 집합으로 보존한다. 실제 envelope와 정확히 일치하는 후보가 있을 때만 수용하며, 없는 envelope·불명확한 후보·충돌은 명시 오류다. 동일 prefix와 동일 opaque digest의 중복 완료는 idempotent하다. 시간순 최신 응답이나 전역 last-provider로 선택하지 않는다. 이전 prefix에서 갈라지는 새로운 요청 자체는 다른 branch이므로 아직 소비되지 않은 병렬 응답을 강제로 끼우지 않는다.
7. **resume·receipt 상실.** receipt는 MoAI 전용 현재 사용자 private 경계에 대화별로 저장하며 gateway supervisor 종료 시 삭제하지 않는다. 새 대화 생성은 launcher가 명시적으로 초기 manifest를 만들 때만 가능하다. resume에서 directory/manifest/receipt가 없거나 손상되면 빈 새 저장소로 초기화하지 않고 송신 전에 거절한다. manifest의 receipt 수·세대·내용 digest로 부분 손실을 검증한다. 이전 snapshot 전체를 일관되게 되돌리는 악의적 rollback까지 hash만으로 막는다고 주장하지 않는다. 사용자 credential·Claude 전역 설정·원본 transcript에는 쓰지 않는다.
8. **compaction·history 편집의 한계.** 필수 opaque가 있는 기존 공개 prefix를 더 이상 재구성할 수 없는 history 축약·편집·알림 재배치·지원하지 않은 resume 변형은 성공으로 넘기지 않고 구체적인 재개 오류를 낸다. compaction을 우회하려고 기존 history를 임의 복사하지 않는다. 정상 native continue/resume에서 prefix가 안정적이라는 실제 양성 대조군을 반드시 먼저 확보한다. 이 조건이 안 맞으면 그 변형은 미완료이며 정상 재개 AC를 제거하지 않고 정규화/명시 branch 인계 규칙을 보강한다.
9. **외부 provider 전환.** 귀속 검사는 foreign strip 전에 한다. 확인된 OpenAI envelope만 제거하고 공개 pair는 원래 ID로 복원한다. receipt는 OpenAI 재복귀 시 필요한 검증 정보로 남으며 다른 provider에 보내지 않는다. foreign provider가 새로 만든 공개 응답도 해당 대화의 공개 prefix에 연결할 수 있지만 OpenAI opaque로 표시하지 않는다. 같은 provider/다른 모델은 provider 소유 item 처리와 실제 수용 시험으로 판정한다.
10. **유한 저장과 보안.** 별도 API/전역 대화 DB를 만들지 않고 대화별 private manifest 한 개의 원자 교체와 OS 프로세스 잠금 패턴을 재사용한다. 초기 구현 후보 한도는 대화별 receipt 4096개·직렬화 8MiB이며 초과 시 오래된 필수 receipt를 조용히 버리지 않고 명시 오류다. 수치와 retained-session 정리 정책은 run 전에 경계 시험과 함께 확정한다. cleanup은 명시 대화 폐기와 연결하고 자동 종료 정리와 혼동하지 않는다. private 권한의 hash 자료도 공개하지 않는다.

**결정적 시험과 실제 관측의 분담**

- 정규화 golden: object key/JSON 공백만 바뀐 양성군; 공개 text·배열·ID·도구 결과 변경 음성군; native 알림과 message 병합의 실제 raw 대조군. 정규화가 다른 공개 대화를 같은 digest로 만드는 뮤턴트를 거절한다.
- no-tool의 실제 encrypted envelope를 유지한 입력은 동일 digest로 복원되고, 전체 삭제·부분 삭제·순서 변경의 각 입력은 upstream 호출 0이다. provider가 삭제를 받아들이는지와 관계없이 gateway 보존 계약을 시험한다.
- 동일 prefix 병렬 응답 둘의 서로 다른 opaque, 같은 공개 답변/다른 opaque, 완전 중복 재전송, 세션 UUID 위조, 다른 대화 receipt 교환, terminal 전 저장 실패, 저장 후 client 취소를 결정적으로 시험한다.
- 종료 후 새 프로세스의 실제 native resume에서 UUID와 공개 prefix가 유지되고 원래 opaque/tool ID가 같은 요청으로 돌아와야 한다. receipt 전체/부분 삭제·manifest 손상은 외부 송신 0이어야 한다. compaction으로 귀속이 불명확한 대조군은 명시 오류이며 정상 resume의 PASS 대체물이 아니다.

이 방식은 존재하는 공개 history를 anchor로 하는 유실 탐지다. 공격자가 공개 history와 receipt 저장소를 함께 전면 교체하는 경우까지 인증하지 않는다. 기존 요구인 envelope 전체·부분 삭제, no-tool 누락, marker 동시 삭제는 공개 history가 남은 실제 mutant로 검증할 수 있다. 정확한 native 정규화와 재개 보존은 지금 수행 중인 probe의 결과를 받아 설계·감사에서 닫는다.

**AC-MG-009 유지·보강안**

> Given 실제 provider encrypted item을 가진 tool 응답과 no-tool 응답을 각각 준비했을 때, When 실제 TUI의 같은 provider 후속 turn·도구 왕복·continue·resume·다른 provider 전환을 수행하면, Then 필요한 원래 item과 순서·바이트가 보존되고 foreign 송신에는 없으며 공개 기록은 보존된다. tool marker만 남긴 전체 envelope 삭제, 부분 삭제, 순서 변경, malformed marker, 다른 assistant 묶음의 marker를 넣는 변형은 외부 송신 0이어야 한다. no-tool 전체 삭제 및 marker와 envelope의 동시 삭제는 검증 가능한 대화 귀속 계약의 범위에서 별도로 판정하며 단순 decoder PASS로 대체하지 않는다. 해당 귀속 계약과 실제 증거가 없으면 이 부분의 활성화 게이트는 닫힌 상태다.

이것은 기존 전체 carrier 유실 시험을 제거하는 조항이 아니다. 어떤 표식이 남는지에 따라 필요한 검증 장치가 다름을 명시한다. 전체 history를 다른 유효 history로 완전히 바꾸는 공격까지 방지한다는 새 인증 보장은 만들지 않는다.

### 6. Windows: 성공 의미를 먼저 정하고 primitive를 선택

**관측과 문서의 경계.** AUTH plan B.1은 디렉터리 내구성 확인 전 성공을 금지하고 primitive 문단은 `internal/atomicfile.Replace`를 지정한다. 현재 Windows primitive는 `os.Rename` 재시도이고 AUTH 후보는 직접 `MoveFileEx(REPLACE_EXISTING|WRITE_THROUGH)`다. 이 작성자가 두 코드 구간을 읽었다. Windows native 실행은 하지 않았다.

Microsoft의 [MoveFileExW 문서](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw)는 WRITE_THROUGH와 copy/delete flush를 설명한다. [FlushFileBuffers 문서](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-flushfilebuffers)는 파일 flush 및 관리자 권한을 요구하는 volume flush를 설명한다. 이 둘만으로 같은 volume rename 뒤 POSIX directory fsync와 동등한 전원 장애 보장을 도출하지 않는다.

**사용자 판단이 필요한 의미 변경안 — 승인 전 적용 금지**

> Where Windows인 경우, the 인증 계층 shall 같은 volume의 현재 사용자 전용 저장 경계 안에서 파일 쓰기·파일 flush·native write-through 교체·최종 파일과 부모 identity·권한·내용 readback이 모두 성공한 뒤에만 로컬 저장 완료를 표시해야 한다. 이 완료는 해당 OS 연산과 재개 가능한 완전한 상태를 확인한 것으로 정의한다. 전원 장애 뒤 디렉터리 엔트리 보존이 POSIX directory fsync와 동등하다는 보장은 별도 근거와 시험 없이는 하지 않는다. 그 더 강한 보장이 출시 조건으로 유지되면 Windows 활성화는 해당 조건이 검증될 때까지 완료로 표시하지 않는다.

이 문구를 채택하면 기존의 플랫폼 구분 없는 디렉터리 내구성 약속을 더 정확하지만 제한된 Windows 계약으로 바꾸는 것이므로 사용자에게 그 차이를 알려야 한다. GitHub CI에서 시험하자는 지시만으로 이 차이를 승인한 것으로 해석하지 않는다. 반대로 사용자에게 primitive 함수명을 고르게 할 필요는 없다.

**의미가 승인된 뒤 조정자가 선택할 가장 작은 구현 경로**

> POSIX canonical 교체에는 기존 atomicfile.Replace를 재사용한다. Windows AUTH canonical 교체에 한해 같은 디렉터리·같은 volume의 native write-through 교체를 허용한다. cross-volume copy와 fallback은 허용하지 않는다. 현재 atomicfile API에 전역 내구성 의미를 추가하거나 관련 없는 호출자를 바꾸지 않는다. AUTH writer가 파일 flush·경로/handle identity·ACL·readback·실패 정리를 함께 책임진다.

이 선택은 최소 범위 제안이다. 기존 atomicfile을 확장하는 선택이 필요하다면 영향받는 기존 호출자와 시험을 따로 조사한 뒤 범위를 승인해야 한다. 현재 primitive를 결함으로 단정하지 않는다.

**현재 사용자 전용 ACL 성공 조건 보강안**

> MoAI가 만드는 canonical 저장 루트·scratch 루트는 현재 사용자 SID가 소유하고 외부 권한 상속을 차단한 DACL을 가져야 한다. 부모 identity와 reparse 여부를 실제 handle에서 확인한다. MoAI가 직접 만드는 credential 상태 파일은 동일한 현재 사용자 접근 제한을 가져야 한다. broker가 생성한 파일은 inherited 표지만으로 거절하거나 허용하지 않고, 검증된 보호 부모에서 생성되었으며 현재 사용자 외 SID에 유효한 읽기·쓰기·삭제·권한 변경 허용이 없고 소유자·handle identity·부모 관계가 맞는지를 검증한 경우에만 읽는다. NULL/부재 DACL, 잘못된 소유자, 넓은 그룹 허용, reparse·경로 교체·잘못된 파일 형식은 거절한다. ACL 변경으로 이미 공개됐을 수 있는 broker 파일을 사후 안전한 것으로 바꾸지 않는다.

보호된 루트와 broker 자식의 상속 ACE는 서로 다른 검증 대상이다. 상속 ACL의 안전한 형태를 실제 Windows broker 대조군으로 확인하기 전 ‘상속이면 안전’ 또는 ‘protected 비트 없으면 전부 지원 불가’로 결론내리지 않는다. OS 관리자 권한에 의한 소유권 강탈까지 막는다는 의미로 현재 사용자 전용을 확대하지 않는다.

**AC-GA-004·005 보강안**

> Given Windows GitHub runner에서 소유자 전용 protected 루트, 안전한 inherited broker 파일, broad/NULL ACL·잘못된 소유자·reparse·부모 교체·flush/replace/readback 실패를 각각 준비했을 때, When 저장·읽기·로그인·refresh·logout을 수행하면, Then 정상 대조군만 완료되고 실패군은 성공 표시나 부분 credential 노출이 없어야 한다. 교체 이전 실패는 기존 canonical을 보존한다. 교체 후 확인 실패는 이전 파일이 남았다고 거짓 보고하지 않고 로컬 상태 변경 가능성을 오류에 구분하며 그 상태를 credential 송신에 사용하지 않는다. 프로세스 강제 종료의 각 경계 뒤에는 재열기 결과가 완전한 이전/새 상태 또는 명시 오류이고 부분 상태를 채택하지 않아야 한다. 별도 프로세스 잠금·취소·강제 종료 후 잠금 회수·늦은 세대 publish 거절도 판정한다.

실제 runner run ID·대상 HEAD·필수 test run/pass 이벤트·artifact를 근거로 남긴다. cross-compile·합성 checker의 PASS를 Windows 실행 PASS로 세지 않는다. 기존 9개 필수 시험을 삭제하거나 Skip으로 통과시키지 않는다. 전원 장애 시험과 프로세스 종료 시험을 구별한다. push·PR·dispatch의 별도 지시 경계는 유지한다.

### 7. 조정자에게 필요한 최소 결정 묶음

1. **이미 질문한 출력 상한 답변**을 먼저 반영한다. 답변이 오기 전 생략 구현은 승인하지 않는다.
2. **Windows 저장 완료의 의미**: 위 플랫폼별 OS 보장 수준을 받아들일지, 기존의 더 강한 디렉터리 내구성 증명을 계속 필수로 둘지 확인한다. 후자는 현재 증거만으로 완료되지 않는다. native 함수 선택 자체는 조정자가 범위 안에서 결정할 수 있다.
3. **no-tool은 기존 대안 (b)의 hash-only receipt를 조정자가 채택·검토할 구체 후보로 제안한다.** 위 설계를 독립 감사하고 진행 중인 native resume 관측을 결합하여 구현한다. 어려움을 이유로 사용자에게 목표 축소를 다시 묻지 않는다. 이는 도구 ID hash만으로 해결됐다고 보고하거나 no-tool 요구를 빼는 결정이 아니다. 현재 제안서만으로 안정적인 대화 귀속이 검증됐다는 전제는 없다.

M0 키 선택·SSE의 엄격한 호환 형식·s/Enter 관측은 위 사용자 결정과 독립적으로 본문 보강 준비가 가능하다. 실제 SPEC 변경은 부모의 재위임 이후에만 수행하고, 기존 구현에 대한 계약 수정과 아직 없는 reasoning/Windows 구현의 계획 감사를 분리한다.

## Evidence

작성 중 이 WT에서 실행한 기준 확인의 실제 출력:

```text
git rev-parse --short HEAD
81c1d58f9
git branch --show-current
WT-unified-gateway
moai session current
5c437366-d8cb-45cf-9261-a850f7127958
```

`sed`/`cat`으로 읽은 요구사항·계획·AC는 코어 REQ-MG-012·013·015·016·017, design §3.2·4.3, plan M0, AC-MG-007·008·009·021, AUTH REQ-GA-004·005와 plan B 및 AC-GA-004·005·008, PICKER REQ-GP-003·005·006·007 및 대응 AC다. 이 읽기는 실제 provider 시험이 아니다.

읽은 완료 보고서: [M0 정상/복구](m0-after19-observation.md), [직접 refresh](m0-refresh-transport-observation.md), [합성 carrier 실제 왕복](picker-reasoning-runtime-observation.md), [picker 후속](picker-followup-runtime-observation.md), [기존 binding 후보](reasoning-binding-candidate.md), [Windows 계약 차이](windows-store-contract-gap.md). 상세 명령·원문 출력·hash·관측 범위는 각 보고서에 있다. 선행 보고서의 수치를 이 작성자가 재실행한 시험으로 표시하지 않는다.

추가로 [native 재개 관측](picker-resume-runtime-observation.md)을 읽고 해당 `picker-resume-20260911T103342Z-9e003c97/raw/request-003.json`과 `request-004.json`을 Python `json.loads`로 직접 비교했다. 첫 경로를 raw 하위 디렉터리 없이 읽으려던 시도는 FileNotFoundError였고, `rg --files --hidden --no-ignore`로 실제 경로를 찾은 뒤 다시 읽었다. 실제 native 실행을 재수행하지 않았다.

검사 방식은 messages의 content 문자열을 단일 text block 배열로 바꾸고, thinking/redacted_thinking과 block의 cache_control만 공개 projection에서 제외한 뒤 `json.dumps(sort_keys=True,ensure_ascii=False,separators=(',',':'))`의 SHA-256을 계산하는 것이다. 첫 요청의 5개 messages와 재개 요청의 처음 5개를 비교했다. 원문 그대로는 index4의 system content가 배열/문자열로 달랐지만 단일 text 문자열은 동일했고, opaque assistant 경계까지의 원문 prefix 자체도 같았다. 출력은 다음과 같다.

```json
{
  "normalized_prefix_equal": true,
  "normalized_prefix_sha256": "fa2c3d9ddfa7b58bdc5d6f5b4e89793a7fc12555899cfdff90b00f46c288cbfd",
  "resumed_normalized_prefix_sha256": "fa2c3d9ddfa7b58bdc5d6f5b4e89793a7fc12555899cfdff90b00f46c288cbfd",
  "source_request003_sha256": "e5ef6a87b49ba353d58b2414f40163c6f4694b609e14957fd6a35045c2f76908",
  "source_request004_sha256": "92b8896c0ef8493deff8c3560e5af69f8c97a3baf08613d0bd0b258aa3b7d2c6"
}
```

같은 두 요청의 `metadata.user_id`를 JSON으로 읽어 session_id를 비교한 실제 출력은 `"metadata_session_equal": true`였다. native session UUID의 일치는 인증의 증거가 아니다.

## Baseline-attribution

2026-09-11, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, `WT-unified-gateway`, HEAD `81c1d58f9`. 부모 source_session_id는 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이며 위 session 출력은 이 하위 작성 세션의 값이다. 동시에 제품과 실계정 관측을 쓰는 다른 작성자가 있으므로 이 제안서만 소유한다.

## Gaps

- `auth-responses-runtime-observation.md`를 추가로 읽었다. 네 모델 direct 함수 왕복과 Luna no-tool opaque 후속의 구조화 증거는 있으나 실제 Claude와 upstream을 연결한 통합·삭제 mutant는 이 보고서의 관측이 아니다. Luna 보존 request SHA-256은 `3d6ae8372d5a671bb2e3d62bce1638229dcaf9afa25a48732c02591edd56be9c`이며 opaque 원문을 이 문서에 복사하지 않았다.
- 사용자 출력 상한 답변, Windows 내구성 의미 승인, no-tool 귀속 계약·실제 유실 변형은 미완료다.
- 실제 product factory·PICKER 설정/인증 양립·TEAMMATE 연결·네 GPT의 Claude Code 왕복·Windows CI는 이 문서의 완료 대상이 아니다.
- SPEC 파일 변경·lint·감사·구현 시험·원격 작업은 수행하지 않았다.

## Residual-risk

SSE의 새 수용 분기를 넓게 구현하면 기존 malformed 거절이 약해질 수 있다. tool hash를 인증으로 오인하거나 no-tool 전체 유실을 누락하면 reasoning 보존을 과장하게 된다. Windows post-replace 오류를 단순 rollback으로 설명하면 실제 디스크 상태와 기록이 어긋날 수 있다. 위 문구는 이러한 차이를 AC에 드러내기 위한 제안이며 실행 증거를 대신하지 않는다.

## 후속 위임 반영 및 검증 기록

부모의 명시 재위임에 따라 코어 5문서, AUTH 3문서, PICKER 3문서의 합계 11개를 수정했다. 각 spec.md의 HISTORY와
version을 갱신했으며 상태는 draft다. progress.md는 세 파일 모두 수정 전 hash와 같다. 소스 코드·commit·push·PR은 변경하지 않았다.

변경 파일:

- SPEC-MOAI-GATEWAY-001: spec.md, plan.md, acceptance.md, design.md, research.md
- SPEC-MOAI-GPT-AUTH-001: spec.md, plan.md, acceptance.md
- SPEC-MOAI-GATEWAY-PICKER-001: spec.md, plan.md, acceptance.md

SPEC ID의 실행한 Bash 정규식 결과:

```text
SPEC-MOAI-GATEWAY-001 PASS
SPEC-MOAI-GPT-AUTH-001 PASS
SPEC-MOAI-GATEWAY-PICKER-001 PASS
```

schema SSOT를 읽고 수정 전 12필드·각 상태/날짜/버전/phase를 Python으로 검사했다.

```text
SPEC-MOAI-GATEWAY-001 FRONTMATTER_PASS
SPEC-MOAI-GPT-AUTH-001 FRONTMATTER_PASS
SPEC-MOAI-GATEWAY-PICKER-001 FRONTMATTER_PASS
```

수정 전 저장한 각 REQ/AC 헤더 목록과 수정 후 목록을 Python으로 동등 비교한 실제 결과:

```text
SPEC-MOAI-GATEWAY-001 REQ 26 AC 25 IDS_UNCHANGED version "0.9.0" status draft
SPEC-MOAI-GPT-AUTH-001 REQ 10 AC 10 IDS_UNCHANGED version "0.2.0" status draft
SPEC-MOAI-GATEWAY-PICKER-001 REQ 9 AC 9 IDS_UNCHANGED version "0.2.0" status draft
changed_count 11
progress_hashes_unchanged True
```

위 헤더 계수는 묘비를 포함한다. 코어의 REQ-MG-007·020과 AC-MG-002에 `[RETIRED]`가 있는 것을 `rg`로 확인했으므로
활성 코어는 REQ 24/AC 24다. AUTH 10/10, PICKER 9/9의 번호와 순서는 그대로다.

아래 세 명령을 수정 전과 최종 수정 뒤 각각 실행했고 각 명령의 출력은 동일했다(exit 0).

```sh
/tmp/moai-gateway-working-20260911 spec lint SPEC-MOAI-GATEWAY-001
/tmp/moai-gateway-working-20260911 spec lint SPEC-MOAI-GPT-AUTH-001
/tmp/moai-gateway-working-20260911 spec lint SPEC-MOAI-GATEWAY-PICKER-001
```

```text
✓ No findings — all SPEC documents are valid
✓ No findings — all SPEC documents are valid
✓ No findings — all SPEC documents are valid
```

`git diff --check -- .moai/specs/SPEC-MOAI-GATEWAY-001 .moai/specs/SPEC-MOAI-GPT-AUTH-001 .moai/specs/SPEC-MOAI-GATEWAY-PICKER-001`
의 출력은 없고 exit 0이었다. 이 트리의 untracked 파일은 diff만으로 내용 검증되지 않으므로 별도로 UTF-8 readback과
변경 hash·식별자 비교를 수행했다. 최종 HEAD/branch 확인은 `81c1d58f9`, `WT-unified-gateway`였다.

**독립 감사 범위.** 아직 구현하지 않은 hash-only receipt·정규화·분기·재개·유실 처리에는 plan-auditor의 계획 감사를
요청한다. 이미 코드가 존재하는 SSE 형식 수용 및 M0 헤더 변경은 수정된 계약에 대한 sync-auditor의 구현 검증을
별도로 요구한다. 구조 lint가 두 감사를 대체하지 않는다. 출력 상한과 Windows 성공 의미는 plan에만 결정 대기로
표시하고 기존 생성 상한 매핑·AUTH B.1 디렉터리 내구성/primitive 요구는 변경하지 않았다.

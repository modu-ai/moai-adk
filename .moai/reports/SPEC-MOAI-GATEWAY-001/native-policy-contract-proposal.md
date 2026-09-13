# Native 요청 정책의 provider별 계약 제안

## Claim

실제 Claude 요청의 제목 schema와 adaptive/high/keep-all 정책을 보존하려면 입력 허용 키뿐 아니라 provider별 의미,
응답 블록 및 종료 검증을 함께 연결해야 한다. 이 문서의 **native 정책 후보는 보고서 단계**다. 승인된 receipt 설계나 제품 코드를 변경하지 않았고 실제 provider 요청·credential 접근을 수행하지 않았다. 작성 중 사용자가 별도로 승인한
구독 출력 정책과 Windows API 저장 기준만 부모의 재위임으로 SPEC에 반영했으며, 이 보고서의 나머지 정책 후보는
SPEC에 반영하지 않았다. 두 승인 변경의 파일·검증은 approved-cap-windows-contract-delta.md에 별도 기록한다.

현재 거절의 실행 근거는 [native-policy-readiness-observation.md](native-policy-readiness-observation.md)다.
원본 요청 003~008이 두 현재 검증 경계에서 모두 거절된 것은 해당 관측자의 overlay Go 시험 결과다. 이번 작성자는
raw003/004의 정책 필드와 현재 코드 구간을 읽었다. 첫 거절 이후의 모든 실패 원인을 시험했다고 주장하지 않는다.

### 1. 지금 진행 가능한 계약과 별도 결정의 구분

| 범위 | 최소 진행안 | 승인·증거 경계 |
|---|---|---|
| Anthropic native 요청 | 구조를 검증하고 관측된 adaptive/high/title schema/keep-all을 원래 의미로 전달 | 기존 OAuth·native 경험 보존 목표 안의 구현 보강. 제품 응답 thinking 처리와 실제 통합 시험 필요 |
| GPT high effort | `output_config.effort:high`를 `reasoning.effort:high`로 명시 매핑 | 네 모델 direct high 함수 수용 근거 있음. 공급자 간 토큰량·추론 과정 동등성은 보장하지 않음 |
| GPT title schema | 아래 정확한 schema를 `text.format` strict JSON schema로 옮김 | 공식 API/Codex 타입과 네 GPT의 정확한 title schema 실제 200 근거. native 변환·일반 schema는 별도 시험 |
| GPT keep-all | client history와 필요한 opaque를 전부 보존하고 provider 전용 context edit 필드는 보내지 않음 | 삭제 지시가 아닌 보존 의도를 구현하는 번역. 승인된 receipt 검사 뒤에만 가능; 자동 축약 없음 |
| GPT native metadata | session_id는 local 귀속 검사에만 사용, account/device 값은 foreign 요청에 싣지 않음 | transport metadata와 대화 내용을 분리하는 명시 정책 보강. 기존 무조건 metadata 복사 변경은 문서화·시험 필요 |
| 입력 계량 | schema·tools·system·history를 빠뜨리지 않는 명시 추정과 provider별 cap 근거 | 추정치를 실제 tokenizer/보장 상한으로 표시하지 않음. 숫자는 모델 이름에서 추측하지 않음 |
| 생성 출력 상한 | 승인된 고정 구독 경로만 outgoing max_output_tokens 생략, 입력 양의 정수 검증/API 키 매핑 유지 | 사용자 “구독 서버 출력 정책으로 진행” 승인. 바이트 제한을 생성 token cap으로 표시하지 않음 |
| Windows·receipt | Windows는 별도 승인된 API 저장 기준, receipt는 현재 계약/게이트 유지 | 사용자 “Windows API 기준으로 진행”의 좁은 AUTH 보강은 별도 delta; receipt 본문은 불변 |

정상 입력을 표준 방식으로 옮기고 잘못된 입력을 거절하는 작업은 기존 전체 구현 목표 안에서 진행할 수 있다.
승인된 구독 출력 정책 밖의 상한 생략, high의 임의 하향, schema 제약 삭제, 필수 reasoning 삭제는 이 제안으로 승인하지 않는다.
관측하지 않은 capability를 구현 편의를 위해 참으로 두는 것도 진행안이 아니다.

### 2. 실제 raw 정책과 가장 작은 입력 문법

작성자가 읽은 raw003의 형식은 아래와 같다. schema는 합성 native 제목 요청의 고정 구조이며 대화 본문은 싣지 않는다.

```json
{
  "output_config": {
    "effort": "high",
    "format": {
      "type": "json_schema",
      "schema": {
        "type": "object",
        "properties": {"title": {"type": "string"}},
        "required": ["title"],
        "additionalProperties": false
      }
    }
  }
}
```

raw004는 `thinking:{type:adaptive}`, `output_config:{effort:high}`,
`context_management:{edits:[{type:clear_thinking_20251015,keep:all}]}`다. raw003에는 thinking과 context_management
키가 **없다**. 이번 Python의 `.get()` 진단에 보인 null은 누락을 표시한 것이며 wire의 명시 null이 아니다.

최소 문법은 다음과 같이 엄격히 분리한다.

- `thinking`: 누락, 정확한 `{type:"adaptive"}`, 기존 `{type:"disabled"}`를 구별한다. null·문자열·배열·알 수 없는 키는
  허용하지 않는다. manual enabled/budget_tokens는 검증되지 않은 정책이므로 임의 adaptive로 바꾸지 않는다.
- `output_config`: 객체이며 검증된 effort·format 키만 받는다. effort는 문자열·모델/인증 경로의 선언된 지원 값이어야 한다.
  최초 필수 활성 값은 실제 관측된 high다. 다른 값은 공식 capability와 각 경로 계약에 추가한 뒤 사용한다.
- `format`: json_schema와 schema 객체를 요구한다. input_schema의 optional 도구 의미와 output format의 strict 의미를
  섞지 않는다. 출력 schema를 강제하려고 일반 도구의 optional 필드를 required로 바꾸지 않는다.
- `context_management`: 관측된 단일 keep-all edit를 정확히 인식한다. keep 수치·clear_tool_uses·복수/중복 edit·알 수 없는
  정책을 keep-all로 간주하지 않는다. 부재와 비어 있거나 잘못된 객체를 구별한다.
- 모든 층에서 기존 중복 JSON key·UTF-8·surrogate·깊이·바이트 한도·tool pair 검사를 유지한다. 알 수 없는 키를
  성공을 위해 조용히 버리는 일반 허용 모드는 만들지 않는다.

### 3. Anthropic native passthrough

**제안 조항 — REQ-MG-016과 기존 AC-MG-008·009·021의 보강**

> Where 검증된 Anthropic native 경로인 경우, the gateway shall 같은 provider의 정상 요청에서 thinking·effort·출력 schema·
> context 보존 정책 및 native reasoning history를 해당 provider 형식으로 보존해야 한다. 잘못된 형식과 미지원 정책은
> 외부 송신 전에 거절하고, 지원 필드 제거·effort 하향·thinking 비활성화로 성공을 만들어서는 안 된다.

구체적으로 native의 output_config와 context_management를 Responses 구조로 바꾸지 않는다. model은 catalog의 검증된
upstream ID만 사용하고, 허용된 beta/header는 기존 인증 경계를 따른다. exact native 모델 ID가 이미 같으면 불필요하게
전체 JSON을 다시 직렬화할 이유가 없다. 필요한 변환이 있어도 공개 문자열·schema·opaque/signature 문자열과 배열 순서를
보존한다. 부동소수점 재직렬화로 숫자 의미를 바꾸지 않는다.

현재 `nativeBlocks`는 text/tool_use/tool_result만, native stream의 block start는 그 함수를 통해 검사한다.
따라서 입력에 adaptive를 허용하는 것만으로 thinking 응답이 통과한다고 단정할 수 없다. 같은 provider에서 반환된
thinking/redacted_thinking 및 관련 delta/signature를 구조·크기·index·순서·완료 상태로 검증하고 그대로 전달하는
응답 보강도 포함한다. 이때 MoAI/OpenAI envelope를 Anthropic 서명으로 취급해서는 안 된다. foreign 전환과 그 envelope
제거는 기존 receipt 계약을 먼저 적용한다. native 서명의 진위 검사를 gateway가 대신 수행했다고 주장하지 않는다.

공식 [thinking/effort 설명](https://platform.claude.com/docs/en/build-with-claude/thinking-steering-and-cost)은 adaptive의
결정과 effort를 구별하고, [structured outputs](https://platform.claude.com/docs/en/build-with-claude/structured-outputs)는
`output_config.format`의 schema 표현을 설명한다. 같은 provider history의 block/순서 보존은
[preserved thinking](https://platform.claude.com/docs/en/build-with-claude/preserved-thinking)의 요구와도 맞는다.
이 문서들은 모든 구독 조합의 runtime PASS를 대신하지 않는다.

GLM이 같은 Messages 검증 함수를 쓰더라도 Anthropic 지원 값을 GLM capability로 자동 확장하지 않는다.
Z.AI의 검증된 정책만 별도 profile로 활성화하고, GLM의 기존 요청이 검증되는지 회귀 시험한다. native 검증 코드를 공유한다는
이유로 다른 provider의 adaptive·schema·beta가 지원된다고 선언하지 않는다.

### 4. GPT Responses 매핑

#### 4.1 Effort와 thinking

| Claude 입력 | Responses 후보 | 제약 |
|---|---|---|
| adaptive + high | reasoning.effort=high | adaptive를 token budget으로 해석하지 않음; receipt 경로가 준비되어야 함 |
| thinking 부재 + high (제목) | reasoning.effort=high | 부재를 disabled로 바꾸지 않음 |
| thinking 부재 + effort 부재 | reasoning effort 필드 부재 | provider 기본값 사용임을 명시; 근거 없이 high/low 주입하지 않음 |
| disabled | 검증된 비추론 모드가 있는 경로에서만 대응 | high와 충돌하면 오류. Astra에 none을 보내거나 low를 disabled와 같다고 표시하지 않음 |
| enabled/budget_tokens 또는 미지원 effort | 명시 정책 오류 | 수치 budget을 high로 추정 변환하지 않음 |

공식 [OpenAI reasoning 문서](https://developers.openai.com/api/docs/guides/reasoning)는 effort 값과 기본값이 모델마다
다르다고 설명한다. Astra의 none 미지원도 명시한다. 따라서 동일 high는 각 provider의 high 의도를 연결하는 명시 번역이며,
계산량·토큰·응답 품질의 수치 동등성 약속이 아니다. 실제 네 GPT의 high 함수 수용은
[직접 Responses 관측](auth-responses-runtime-observation.md)에 있다. provider가 요청마다 opaque를 반드시 반환한다는
전제는 두지 않는다. 반환되면 tool 유무와 관계없이 기존 보존/receipt 계약을 따른다.

#### 4.2 제목용 출력 schema

관측된 정확한 title schema의 최소 wire 매핑 후보:

```json
{
  "text": {
    "format": {
      "type": "json_schema",
      "name": "moai_native_output",
      "strict": true,
      "schema": {
        "type": "object",
        "properties": {"title": {"type": "string"}},
        "required": ["title"],
        "additionalProperties": false
      }
    }
  },
  "reasoning": {"effort": "high"}
}
```

고정 name은 응답 format의 transport 식별자이며 사용자의 schema를 재명명하거나 tool 이름 역매핑과 공유하지 않는다.
원래 schema는 변형하지 않는다. 이 title 구조는 required와 additionalProperties를 이미 갖추므로 optional 필드의 의미를
바꾸지 않고 strict schema 후보로 옮길 수 있다. 일반 schema의 unsupported keyword·optional 의미를 제거하여 title 형식에
억지로 맞추지 않는다. 지원하지 않은 형식은 오류로 남기고 공통 지원 subset을 명시적으로 늘린다.

공식 [Structured outputs](https://developers.openai.com/api/docs/guides/structured-outputs)는 Responses의 text.format에
name/schema/strict를 두는 형식과 제한된 JSON Schema subset을 설명한다. 공식 Codex 소스 `common.rs:185–206`도
TextFormat(type,strict,schema,name)과 TextControls(format)를 갖는다. 이것은 고정 구독 endpoint가 해당 요청을 수용했다는
증거는 아니지만, 작성 중 부모가 완료한 [title-policy-runtime-observation.md](title-policy-runtime-observation.md)를
추가로 읽었다. 네 GPT 각각 한 요청이 high·strict title schema로 HTTP 200/SSE 완료/유효 title JSON을 반환했다.
preflight의 name은 `moai_title_preflight`였고 이 제안의 고정 이름과 구별한다. 이는 정확한 작은 title schema의 구독 수용
근거이며 native 변환·모든 schema의 PASS는 아니다. 수용 실패 시 schema를 빼고 plain text로 재시도하지 않는다.

제목 응답의 정상 종료에는 공개 output_text를 합친 JSON이 원래 title schema와 일치해야 한다. title 누락·추가 속성·
비문자열·깨진 JSON·refusal·미완료는 정상 title 성공으로 표시하지 않는다. 이미 스트림 byte를 보냈다면 오류를 드러내고
성공 terminal을 덧붙이지 않는다. JSON 문자열을 임의로 고치거나 JSON이 아닌 답변에서 제목을 추출하여 성공시키지 않는다.
일반 output schema validator를 불필요하게 새로 만들기보다 기존 검증기를 먼저 찾고, 없으면 실제 지원할 subset을 명확히 한다.

#### 4.3 keep-all

정확한 `clear_thinking_20251015/keep:all`은 reasoning 삭제 요청이 아니다. 공식
[context editing](https://platform.claude.com/docs/en/build-with-claude/context-editing)은 all이 thinking 전체 보존임을
설명하며 이 정책은 Anthropic 서버에서 적용된다고 구분한다. GPT 번역은 그 Anthropic 전용 필드를 그대로 보내는 대신,
전체 공개 history와 필요한 OpenAI opaque를 순서대로 재구성하고 receipt 검사를 통과시킨다. 이 매핑은 **보존 의도에 대한
설계 해석**이며 서로 다른 provider의 서버 context 기능이 같다는 주장이 아니다.

구독 요청에는 검증되지 않은 context_management/truncation 옵션을 추가하지 않는다. `store:false`와 encrypted_content
include는 기존 경로를 유지한다. API 키 경로의 `truncation:disabled`도 그대로 둔다. keep-all 수용을 이유로 receipt의
유실·compaction/편집 실패를 무시하거나 context를 맞추려고 oldest messages/opaque를 버리지 않는다. 다른 clear 정책은
별도 의미와 실제 수용이 정의될 때까지 명시 오류다.

#### 4.4 Native metadata

raw metadata는 `user_id` 한 문자열이며 그 안에 account_uuid/device_id/session_id가 있다. session_id는 승인된 receipt
bootstrap과 일치하는지 **로컬에서 먼저** 검증한다. GPT request에는 이 Claude 계정·장치 식별 묶음을 통째로 보내지 않는
정책을 제안한다. metadata는 공개 message가 아니며 이 제거는 공개 대화 손실을 뜻하지 않는다. 다른 명시 metadata는
인증 경로별 허용 목록과 사용 목적이 확인된 경우에만 매핑하고, 알 수 없는 키를 임의 client_metadata로 바꾸지 않는다.

현재 translator는 문자열 metadata를 그대로 복사한다. 공식 Codex 요청 타입에는 `client_metadata`가 있으나 native
metadata와 동의어라는 근거는 없다. API 키의 metadata 지원과 구독 client_metadata를 혼동하지 않는다. 필요한 상관관계는
기존 local request context로 처리하고 사용자 ID를 hash하여 새 추적 ID를 외부에 발명하지 않는다. header/로그에도 해당
계정·장치 원문을 새로 넣지 않는다. 이 변경은 작은 명시 metadata 정책 보강으로 문서·wire 음성 시험을 함께 요구한다.

### 5. 문맥 입력 계량과 cap

현재 `EstimateInputTokens`는 messages/system/tools만 JSON 투영하여 `(RuneCount+3)/4`를 계산한다. 이는 정확 tokenizer도
보장 상한도 아니다. 새 output schema도 provider 입력 구성에 영향을 줄 수 있으므로 이 정책을 지원한 뒤 기존 투영만을
‘전체 정책 입력 계량’으로 표시하지 않는다.

구체 제안:

1. 원본 ingress 바이트 제한은 그대로 둔다. receipt·모델·정책 검증 뒤 실제 egress에 포함될 input/messages·instructions/system·
   tools·output schema를 계량 대상에 명시한다. native metadata나 effort 숫자를 임의 reasoning token으로 환산하지 않는다.
2. 현재 rune 추정은 `accuracy:estimate`를 유지하고 `json-runes-div4`의 대상 필드와 버전을 문서화한다. 기존 count_tokens와
   adapter가 다른 투영을 사용하면 각각의 이름/대상을 구별한다. 실제 provider tokenizer와 같다고 표시하지 않는다.
3. context cap은 exact model+auth method+endpoint별 검증 자료와 숫자의 의미를 catalog에 귀속한다. 총 context인지 입력 전용인지,
   출력 예약이 포함되는지를 기록한다. 공개 API model 문서의 숫자를 구독 계정에서 직접 측정한 값으로 표시하지 않는다.
4. 총 context와 실제 강제되는 출력 예약이 정의된 경로에서는 정수 overflow를 피하여 입력 허용 범위를 계산한다.
   승인된 구독 서버 출력 정책에서는 보장되지 않은 max_tokens 예약을 ‘강제된 출력 한도’로 설명하지 않는다. 요청이 상한을
   넘으면 로컬 오류이며, 추정이 낮아 실제 provider가 거절하면 그 context 오류를 명시 전달한다. silent truncation 없음.
5. 누락/0/잘못된 cap은 모델 이름이나 임의 큰 값으로 대체하지 않는다. 해당 capability 활성화 근거를 확보한다.
   바이트 cap·deadline·JSON 깊이/스키마 크기 제한은 별도의 자원 제한이며 생성 token cap과 다르다.

이 부분은 승인된 구독 출력 정책을 넘어 생성 상한 보장을 추가하지 않는다. 정확한 계량 수단이 없는 경우 추정과 provider-authoritative 오류의
한계를 문서에 남기며 ‘모든 oversized 입력을 사전에 완벽 차단’이라는 새 보장을 만들지 않는다. 기존 AC-MG-010의
고정 capability/알려진 초과 입력 대조군과 앞부분 무삭제 계약은 유지한다.

### 6. 기존 번호로 보강할 AC

| 기존 AC | Given / When | Then |
|---|---|---|
| AC-MG-021·008 | native title/main raw 정책과 정상 OAuth 대조군 / 제품 gateway로 전달 | 선택된 모델·schema·effort·adaptive·keep-all·허용 헤더 보존, native thinking/signature stream 정상 순서/종료 |
| AC-MG-004·007·008 | 실제 title schema와 high / GPT 변환 및 실제 네 모델 title 요청 | schema 원문 의미 유지, text.format/high 일치, 정상 JSON title 수용; schema 삭제·optional 강제·refusal/잘림 성공 변형 거절 |
| AC-MG-009 | adaptive/high·keep-all 및 이전 tool/no-tool opaque / 실제 후속·resume | 승인된 receipt 검사 후 원래 순서·바이트 보존; 누락·변조·foreign 전송은 기존 음성군 실패 |
| AC-MG-010 | schema/tools/system/history 각각을 크게 만든 경계 입력 / count와 adapter 검사 | 모든 선언된 입력 투영 반영, 초과/잘못된 cap 명시 오류, 문자열/앞부분 무삭제, 추정 표기 유지 |
| AC-MG-011 | 제목 요청과 본 대화·병렬 subagent의 서로 다른 모델/정책 / 동시 전달 | 요청별 model·effort·schema·receipt가 섞이지 않음; 제목 수신을 본 대화 성공으로 세지 않음 |
| AC-GA-008·010 | 구독/API 각각의 정책과 실제 네 모델 / 해당 경로 요청 | 구독과 API 형식·cap·metadata 차이를 명시, 타 과금/모델 fallback 0, native Claude 실제 왕복으로 최종 판정 |

모든 허용 정책에는 null·잘못된 타입·중복 필드·알 수 없는 enum/key·미지원 조합·길이 초과·다른 provider capability의
무송신 음성군을 둔다. disabled+high나 지원하지 않는 none을 낮은 effort로 바꾸는 뮤턴트도 거절한다. title용 effort 부재와
main adaptive 부재를 서로의 기본값으로 채우지 않는다. 기존 ordinary text/tool 및 종료 표의 대조군을 삭제하지 않는다.

### 7. 실행 순서와 인계

1. native/input·응답 policy 문법과 GPT wire 매핑의 로컬 양성/음성 대조군을 먼저 준비한다. 현재 raw003~008 거절 관측을
   새로운 입력 지원의 RED 근거에 연결하되 정책별 첫 실패 지점을 독립적으로 판정한다.
2. 부모의 완료된 title schema/high preflight를 제한된 양성 대조군으로 연결한다. 이 작성자는 호출하지 않았다.
   metadata 정책은 body key와 계정·장치 표지 부재로 판정한다. 응답 내용·token을 보고서에 복사하지 않는다.
3. 승인된 receipt 구현과 결합하여 keep-all·native same-provider thinking·foreign 전환의 실제 왕복을 검증한다.
4. context profile과 입력 투영 근거를 연결한 뒤 실제 Claude에서 title+main 요청, 네 GPT 회상·도구·stream·재개를 판정한다.

추가 사용자 질문을 만들어 정상 정책 구현을 멈출 이유는 없다. 출력 상한의 사용자 답변은 도착했으며 고정 구독
경로만 서버 출력 정책을 사용하도록 별도 반영했다. 나머지 정책 후보와 실제 통합 검증은 계속 남아 있다. schema stripping·effort downgrade·history 삭제를 선택해야만 한다면 그때는
관측과 구체 차이를 부모에게 반환하며 자동으로 요구를 줄이지 않는다.

## Evidence

실행한 기준 명령과 원문 출력:

```text
git rev-parse --short HEAD
81c1d58f9
git -C /tmp/openai-codex-audit.zbvNXB rev-parse --short HEAD
5a9eb14
```

Python으로 raw003/004를 json.loads한 후 정책과 metadata **키 이름만** 읽었다. 대화 내용·계정 ID 값은 출력하지 않았다.

```text
request 3 sha256 4406c0852222412720c83f5a44081e2e2c068b5f0fa77f013dc9fe8795c1082b
request 4 sha256 ad2ef38e71be87d715dc2a76bde98d960b25866d0912b33b2655168e30408cd2
metadata_keys: [user_id]
metadata_user_id_keys: [account_uuid, device_id, session_id]
```

정책 객체는 §2에 옮겼다. 원본 위치는 `picker-followup-20260911T102039Z-26558abd/raw`이며 부모 source_session_id의
ignored gateway-entry 아래다. request003~008의 현재 두 validator 거절은 readiness 보고서의 실제 Go overlay 출력에
귀속한다. 그 시험을 이번 작성자의 재실행으로 세지 않는다.

읽은 코드: `anthropic.go`의 nativeRequest/nativeBlocks/nativeBody.event·Send, `translate/request.go`의 허용 키/metadata/
max_tokens 매핑, `openai.go`의 cap·API truncation 분기, `input_estimate.go`의 투영, catalog Capabilities다.
처음 추정한 `input_measure.go` 경로는 없었으며 `rg --files`로 실제 `input_estimate.go`를 찾아 읽었다.
공식 Codex checkout의 `common.rs:185–206,259–283`를 직접 읽어 text/reasoning/metadata 관련 타입을 확인했다.

외부 근거는 본문에 직접 연결한 공식 Anthropic/OpenAI 문서를 web open/find로 읽었다. 일반 API 문서와 고정 구독
backend 실제 수용을 구분했다. 새 provider 호출·credential 파일 읽기·SPEC lint·제품 테스트는 수행하지 않았다.

## Baseline-attribution

2026-09-11, WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, HEAD `81c1d58f9`의 작업 중인 코드와
문서를 읽었다. 부모 source_session_id는 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이다. 다른 작성자가 receipt와 구현을
진행하므로 이 보고서만 소유하며 코드 상태가 이후 변할 수 있다. 이 제안은 새 SPEC ID·AC 번호를 만들지 않는다.

## Gaps

- 정확한 title schema의 네 GPT 구독 수용은 부모 보고서에서 확인했다. native adaptive/keep-all 통합·일반 schema·모든 effort 조합·metadata endpoint 수용·모델별 cap 활성화는 미완료다.
- native 응답 thinking 확장의 모든 이벤트 문법과 provider별 GLM 정책은 별도 명시·실행 확인이 필요하다.
- 공개 schema subset과 실제 title 구조의 매핑 후보는 있으나 공급자 간 품질/토큰 동등성을 측정하지 않았다.
- receipt 구현·감사 판정은 변경하지 않았다. 구독 출력 및 Windows API 저장 기준은 실제 도착한 두 승인만 별도 SPEC에 반영했다. 나머지 native 정책 후보는 보고서 단계다.

## Residual-risk

허용 키만 늘리면 정책 의미를 잃거나 첫 thinking 응답에서 다시 실패할 수 있다. 같은 enum 이름을 토큰량 동등성으로
설명하거나 일반 API 문서를 구독 backend의 지원 증거로 바꾸면 과장이다. title schema의 성공을 본 대화/도구 성공으로
세지 않고, 계량 추정·바이트 제한·출력 cap을 분리하여 기록해야 한다. 보고서의 제안을 승인된 제품 동작으로 표시하지 않는다.

## 후속 title 양성 근거의 귀속

부모의 title-policy-runtime-observation.md를 직접 읽었다. 2026-09-11 11:16:31~38 UTC의 네 응답 각각은 HTTP 200,
SSE completed, title:string만 가진 유효 JSON이다. 고정 구독 endpoint, high, strict schema, stream true/store false이며
승인된 정책에 따라 max_output_tokens를 생략했다. 정확히 네 요청이고 재시도하지 않았다. 보고서 원문의 종료 출력:

```text
--- PASS: TestTitlePreflightLive (9.82s)
PASS
ok  github.com/modu-ai/moai-adk/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/auth-live 10.131s
```

이 작성자는 actual 호출이나 응답 raw를 재실행/출력하지 않았다. 원시 파일 네 개의 0600·길이·SHA는 부모 보고서에
있다. 이 양성을 native→Responses 매핑 동등성·전체 JSON schema·context keep-all·계정의 모든 정책 지원으로 확대하지 않는다.

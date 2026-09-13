# Native 요청 정책 후보의 변경 준비 검토

Verdict: PASS — 기존 목표 안에서 SPEC 보강을 진행할 수 있는 후보
Scope: native-policy-contract-proposal.md의 명시된 최소 provider 정책과 AC 연결
Artifact type: 제안 보고서 검토; 전체 SPEC 계획 감사 또는 제품 활성화 판정 아님

Reasoning context ignored per M1 Context Isolation.

## Claim

이 후보는 기존의 “Claude Code 안에서 네 GPT의 실제 대화·도구·stream·재개” 목표를 구현하는 데 필요한 정책 매핑을 제시한다. **명시한 범위 그대로 SPEC에 옮기고 구현·검증하는 데 새 사용자 결정을 요구할 의미 변경을 확인하지 못했다.** high를 낮추거나 schema를 버리거나 history/opaque를 삭제하지 않으며, 별도 승인된 구독 출력 정책 밖의 생성 상한 생략을 허용하지 않는다.

후보의 PASS는 모든 정책이 이미 구현·수용되었다는 뜻이 아니다. SPEC 보강 시 provider별 지원 문법·응답 검증·cap 근거·기존 AC를 그대로 연결해야 한다. 일반 schema/모든 effort/GLM 지원을 자동으로 켜거나 관측되지 않은 cap을 발명하는 권한은 주지 않는다.

검토 전의 실패 가설: 입력 allowlist만 확장하고 native thinking 응답을 누락, adaptive를 임의 budget으로 치환, disabled를 low로 오인, title schema 제거/optional 강제, keep-all을 history 삭제로 구현, Claude account/device 식별자를 GPT로 복사, 도구/출력 schema를 빼고 입력 계량, title 200을 전체 대화 성공으로 세는 경우였다. 다음 본문에서 이 위험을 직접 대조했다.

## 정책별 판정

| 항목 | 변경 준비 판정 | 제안서 위치와 확인 |
|---|---|---|
| 엄격한 native 입력 문법 | PASS | :55–68. missing/null/type/unknown/중복을 구별하고 adaptive/disabled/manual budget을 동일 취급하지 않음 |
| Anthropic native | PASS | :72–91. 원래 형식·숫자·문자열·schema와 thinking/signature 응답 문법/순서/종료 검증을 함께 요구 |
| provider 분리 | PASS | :94–96. 공유 validator 때문에 GLM capability를 자동 활성화하지 않음 |
| GPT high/adaptive | PASS | :101–114. high→high의 명시 매핑, token/품질 동등성 주장 없음; disabled의 검증된 비추론 모드만 허용, 충돌/미지원 오류 |
| exact title schema | PASS | :118–155. 기존 required/additionalProperties 의미를 유지한 text.format strict schema; 일반 schema 강제 변형 금지; 응답 JSON과 정상 terminal 검사 |
| keep-all | PASS | :159–168. 보존 의도의 번역으로 명시, Anthropic 전용 edit를 GPT에 복사하지 않고 receipt 검증 뒤 전체 공개/opaque 순서 유지 |
| native metadata | PASS | :172–180. UUID는 local bootstrap 검증, account/device 묶음은 foreign egress에서 제외; 임의 client_metadata/hash 추적 ID 생성 금지 |
| 계량/cap | PASS, activation 근거 필요 | :184–204. 실제 egress 투영에 schema/tools/system/history 포함, estimate 표시, model+auth+endpoint 근거와 cap 의미 확인, 0/미확인 cap 임의 대체 금지 |
| AC 연결 | PASS | :208–219. 기존 AC-MG-004/007/008/009/010/011/021 및 GA-008/010의 양성/음성·네 모델·실제 통합으로 연결 |

[Anthropic context editing 문서](https://platform.claude.com/docs/en/build-with-claude/context-editing)의 keep 표는 `all`이 thinking 전체 보존임을 설명한다. 이를 GPT 서버의 같은 기능 지원으로 확장하지 않고 보존 의도의 번역이라고 제한한 :162–163은 타당하다. 이번에 공식 페이지를 직접 열어 해당 표를 확인했다.

[OpenAI Structured outputs 문서](https://developers.openai.com/api/docs/guides/structured-outputs)의 text.format/strict/schema 형식과 제한된 subset을 직접 읽었다. 제안은 작은 title schema의 의미를 그대로 옮기며 도구의 optional schema를 바꾸지 않는다. 공개 API 문서를 구독 endpoint의 실행 근거로 쓰지 않고 별도 네 모델 관측과 연결했다.

## 구체적인 차이와 남은 구현 조건

- exact title preflight의 format name은 `moai_title_preflight`, 후보는 `moai_native_output`이다(제안:149). 문서는 이 차이를 숨기지 않고 native 변환의 추가 검증을 남긴다. transport 이름 선택은 구현자가 결정할 수 있는 범위이며, 이름 변경 뒤 wire 대조와 실제 통합을 생략하는 근거가 아니다.
- “adaptive + high”의 high 매핑은 계산량 동일성의 약속이 아니다(:111–114). disabled/manual enabled/budget 조합을 조용히 high로 고치는 구현은 허용되지 않는다.
- 초기 지원은 실제 title schema와 명시한 policy 조합이다. 일반 schema validator 전체를 새로 만드는 요구가 아니며, 지원하지 않는 subset은 정책 오류다(:141–155). 단, 실제 native가 사용하는 필수 입력을 이 이유로 완료 범위에서 빼서는 안 된다.
- cap은 실제 model/auth/endpoint에 귀속한 근거가 있어야 한다(:194–199). 공개 shipped catalog 수치와 계정/runtime 실측치는 구별해야 한다. 본 검토는 숫자나 capability를 선택하거나 활성화하지 않는다.
- 응답 thinking/signature의 구체 event 문법은 구현 전에 해당 native profile에 고정하고 양성/음성 fixture로 검증해야 한다(:83–90, :217–219). 이 부분이 미구현이라는 사실을 사용자에게 임의 thinking 비활성화 결정을 다시 요구하는 사유로 삼지 않는다.

위 사항은 후보가 이미 요구하거나 미검증이라고 명시한 구현·검증 조건이다. 새 blocking 결함이나 추가 사용자 승인 항목으로 포장하지 않았다.

## Defects Found

이 후보의 명시 범위와 AC 연결에서 새로운 blocking 또는 optional 결함을 확인하지 못했다. 기존 receipt 게이트는 재개하지 않았다. readiness 관측의 validator 거절은 과거 실행 증거로 읽었고, 동시 작업 중인 현재 제품 코드의 실패를 이번 감사에서 재현했다고 주장하지 않는다.

## Evidence

이번 baseline 명령과 출력:

```text
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified branch --show-current
WT-unified-gateway
```

native-policy-contract-proposal.md 전체를 읽되 처음 잘린 :104–215 구간을 `nl -ba … | sed -n '104,215p'`로 다시 읽었다. native-policy-readiness-observation.md와 title-policy-runtime-observation.md 전체를 읽었다. 이 문서들의 Go/live 실행을 재실행하지 않았다.

이번에는 title 보고서가 가리키는 **보존 raw 응답 네 개를 오프라인으로 직접 읽어** 파일 SHA/길이/0600, SSE completed, 완료 message의 JSON `title:string` 단일 속성을 재계산했다. 응답/opaque 본문은 출력하지 않았다. 실행 코드:

```python
from pathlib import Path
import json,hashlib,stat
b=Path('/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai')
r=b/'reports/SPEC-MOAI-GATEWAY-001/title-policy-runtime-observation.md'
rows=json.loads(r.read_text().split('```json\n',1)[1].split('```',1)[0]);assert len(rows)==4
for row in rows:
 p=Path(row['file']); raw=p.read_bytes();assert hashlib.sha256(raw).hexdigest()==row['sha256'];assert len(raw)==row['bytes'];assert stat.S_IMODE(p.stat().st_mode)==0o600
 events=[json.loads(x[5:].strip()) for x in raw.decode('utf-8').splitlines() if x.startswith('data:') and x[5:].strip()!='[DONE]']
 completed=any(e.get('type')=='response.completed' for e in events)
 texts=[]
 for e in events:
  if e.get('type')=='response.output_item.done':
   it=e.get('item',{})
   if it.get('type')=='message':texts.extend(c['text'] for c in it.get('content',[]) if c.get('type')=='output_text')
 title=json.loads(''.join(texts));valid=isinstance(title,dict) and set(title)=={'title'} and isinstance(title['title'],str)
 assert completed and valid
 print(row['model'],'RAW_HASH_MATCH',True,'MODE0600',True,'COMPLETED',completed,'TITLE_SCHEMA',valid)
rawbase=b/'state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-followup-20260911T102039Z-26558abd/raw'
for i in (3,4):
 p=rawbase/f'request-{i:03}.json';q=json.loads(p.read_text()); print('request',i,'sha',hashlib.sha256(p.read_bytes()).hexdigest(),'policy_keys',[k for k in ('thinking','output_config','context_management') if k in q],'metadata_keys',list(q.get('metadata',{})))
```

`python3 - <<'PY'` 단일 invocation의 해당 원문 출력; exit 0:

```text
gpt-5.6-luna RAW_HASH_MATCH True MODE0600 True COMPLETED True TITLE_SCHEMA True
gpt-5.6-sol RAW_HASH_MATCH True MODE0600 True COMPLETED True TITLE_SCHEMA True
gpt-5.6-terra RAW_HASH_MATCH True MODE0600 True COMPLETED True TITLE_SCHEMA True
gpt-6-astra RAW_HASH_MATCH True MODE0600 True COMPLETED True TITLE_SCHEMA True
request 3 sha 4406c0852222412720c83f5a44081e2e2c068b5f0fa77f013dc9fe8795c1082b policy_keys ['output_config'] metadata_keys ['user_id']
request 4 sha ad2ef38e71be87d715dc2a76bde98d960b25866d0912b33b2655168e30408cd2 policy_keys ['thinking', 'output_config', 'context_management'] metadata_keys ['user_id']
```

이 검산은 보존 파일의 양성 구조와 출처 hash를 확인한다. HTTP 200 및 live 요청 횟수는 부모의 title-policy-runtime-observation.md에 귀속하고, 이 실행에서 HTTP 요청을 새로 보내 확인한 것으로 표시하지 않는다. 이 작은 판독기는 제품의 strict SSE/중복·순서·실패 음성 검증기를 대체하지 않는다.

같은 invocation에서 `hashlib.sha256`으로 측정한 후보 문서:

```text
native_proposal_sha c7421f79a38731cf0d88d4d5b75193af1dbd30b9d3233ef6e94de3a8736daa37
```

## Baseline-attribution

WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, HEAD/branch 위 출력, 대상 `.moai/reports/SPEC-MOAI-GATEWAY-001/native-policy-contract-proposal.md`. 부모 source_session_id `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`의 보존 관측을 오프라인 검산했다. SPEC·제품 코드·기존 receipt 판정·원격 상태를 변경하지 않았다. credential 저장 파일·keychain을 읽거나 실제 provider를 호출하지 않았다.

## Gaps

native→Responses 제품 변환, native thinking/signature 이벤트의 모든 변형, 실제 title+main 연속 실행, 모든 schema/effort/GLM 정책, cap 숫자 활성화·계정별 유효 한도, 실제 keep-all/opaque/foreign/resume 통합은 이 검토에서 실행하지 않았다. 공식 문서의 현재 형식과 보존 raw의 양성 구조는 제품 통합 증거가 아니다. 전체 SPEC MP-1~8 판정도 이 제안 검토의 범위가 아니다.

## Residual-risk

정책 allowlist만 늘리거나 opaque 반환을 항상 있다고 가정하면 실제 후속 요청이 실패할 수 있다. title의 작은 schema 성공을 일반 schema 지원으로 확대할 수 없다. rune 추정은 정확 tokenizer나 보장 상한이 아니며 provider의 명시 context 오류를 숨기지 않아야 한다. 서로 다른 provider의 high/keep-all을 토큰량·서버 기능의 동등성으로 표시하지 않는다.

## Recommendation

기존 전체 목표의 승인 범위에서 후보의 명시 문법·provider별 매핑·응답 검증·AC를 SPEC에 보강하고 구현·검증을 진행한다. 추가 사용자 결정 없이 진행 가능한 범위는 **의미를 유지하는 이 최소 번역**이다. 추후 측정에서 schema 삭제·effort 하향·필수 history/opaque 삭제 등의 요구 완화가 필요해진다면 그 구체 차이를 보고해야 하며, 이 판정으로 자동 승인된 것으로 처리하지 않는다.

# Receipt 설계 변경분 독립 계획 감사

Iteration: 1/3 (receipt 변경분의 첫 감사)
Verdict: FAIL
Overall Score: 0.6875
Scope: SPEC-MOAI-GATEWAY-001 0.9.0의 미구현 receipt/opaque 설계 변경분만

Reasoning context ignored per M1 Context Isolation.

## Claim

가역 tool ID와 대화별 hash receipt를 결합하는 방향은 no-tool 전체 envelope 삭제까지 시험할 수 있는 설계다. 그러나 현재 문서는 **대화 식별·native transcript 보존 수명**을 구현자가 결정해야 하는 상태이며, **정상 병렬 후보의 수용 여부가 design과 AC에서 상충**한다. 두 blocking 항목을 고친 뒤 그 변경분만 다시 감사해야 한다. 아직 receipt 구현에 대한 승인 판정은 내리지 않는다.

이번 판정은 전체 코어 SPEC 재감사나 기존 OAuth/SSE 구현 감사가 아니다. 출력 상한과 Windows 저장 의미는 위임에서 제외되었다. 코어 plan.md:224의 미해결 출력 상한 표식은 실제로 확인했으며 전체 SPEC의 MP-7 PASS로 처리하지 않는다.

읽기 전에 설정한 실패 가설은 다음과 같다: 요청 UUID만 믿는 다른 대화 접근, 새 프로세스에서 transcript 소실, 과도한 history 정규화, no-tool/marker 동시 삭제 누락, 같은 공개 답변의 opaque 덮어쓰기, receipt 부분 소실의 빈 저장소 취급, terminal 이후 publish, foreign opaque 전송, 무한 후보/저장, 정상 동시 요청을 오류로 규정한 AC. 아래 판정은 이 가설의 확인 범위만 반영한다.

## Defects Found

### RPA-1 — 재개할 native 대화의 사전 귀속과 보존 수명 미정

- 위치: design.md:411, :417, :420, :427; research.md:1036–1038; PICKER plan.md:92–93.
- Severity: major. Class: blocking. Confidence: high. Receipt 구현 착수 차단: yes.
- design:411은 launcher가 허용한 UUID만 연다고 정하지만, 새 대화의 UUID를 launcher가 어떤 시점에 확정하여 native 클라이언트와 일치시키는지, continue/선택형 resume/fork에서 대상 UUID를 외부 송신 전에 어떻게 확정하는지 정하지 않는다. :417/:420은 receipt만 종료 후 보존한다. 그 UUID의 native transcript/config 저장 경계를 종료 후 어디에 유지하며 다음 프로세스가 어떻게 다시 사용하는지 없다.
- 확인한 양성 관측은 **같은 config와 project를 두 프로세스 동안 유지**하고 처음에 `--session-id`, 두 번째에 정확한 `--resume UUID`를 지정한 경우다. `picker-resume-runtime-observation.md`의 Evidence와 Baseline-attribution이 이를 명시한다. 새 `picker-settings-source-observation.md`는 private config와 기존 secure-storage namespace의 OAuth 공존을 입증하지만, 끝의 Gaps/Residual-risk는 원래 프로젝트 세션과 resume 연결을 후속 계약으로 남긴다. 따라서 이 두 관측을 합쳐 임시 config가 정리된 뒤 정상 재개까지 입증했다고 읽을 수 없다.
- 영향: 구현자가 매 launch마다 private config를 만들고 종료 때 지우는 방법과, 대화별 native 저장 경계를 보존하는 방법을 모두 선택할 수 있다. 전자는 이번 양성 재개 대조군의 전제를 보존하지 않는다. 실제 제품에서 이 실패가 발생했다고 주장하는 것은 아니다.
- Required fix: **하나의 최소 lifecycle 표**로 new/exact resume/continue·선택형 resume/fork의 UUID 결정자·결정 시점·인증된 인계·저장 root 선택을 정한다. 현재 관측된 `--session-id`와 exact resume를 활용할 수 있다. native가 쓰는 transcript/config의 보존 주체, 원본 사용자 설정/credential과의 경계, 정상 종료와 명시 폐기의 차이를 명시한다. 지원하지 않은 변형은 완성된 것으로 표시하지 말고 기존 재개 AC를 유지한다. 새 프로세스의 retained config+receipt 양성, 잘못된 UUID·누락 manifest·정리된 transcript 음성을 같은 AC에 연결한다. 기존 credential/대화 원문을 MoAI receipt에 복제할 필요는 없다.

### RPA-2 — 병렬 후보와 빈 opaque의 수용 표가 상충

- 위치: design.md:413, :416; acceptance.md:169–175.
- Severity: major. Class: blocking. Confidence: high. Receipt 구현 착수 차단: yes.
- design:416은 동일 공개 prefix에 서로 다른 opaque가 있는 후보 집합을 보존하고, 실제 envelope가 정확한 후보와 일치하면 수용하도록 한다. 그러나 acceptance:174는 **“병렬 동일 공개 답변의 다른 opaque”를 각각 송신 0으로 판정**한다고 적어 정상 후보까지 거절하는 읽기를 만든다. 삭제/교차 혼합/미등록 후보라는 조건이 없다.
- 또한 design:413은 opaque 없는 응답에 `opaque_required=false`를 둔다. :416의 “없는 envelope … 명시 오류”에는 required=false의 유일한 정상 후보, required=true/false가 섞인 모호한 후보를 구분하는 조건이 없다. 실제 no-opaque 응답을 보존 대상으로 다루는 계약과 연결해야 한다.
- 영향: 정확한 후보를 통과시키는 구현과 모든 동일 공개 prefix 병렬 응답을 막는 구현이 서로 다른 AC 결과를 얻는다. 후자는 동시성을 위한 후보 집합 설계의 목적을 충족하지 않는다.
- Required fix: 같은 공개 prefix에 대해 최소한 (1) A/B opaque 각각 정확 일치 → 각각 정상 송신, (2) 다른 후보의 일부를 혼합/변조하거나 필수 envelope 삭제 → 송신 0, (3) 동일 opaque 중복 → idempotent, (4) 유일한 required=false 후보에서 envelope 없음 → 정상, (5) 빈 envelope로 후보가 모호한 경우 → 명시 오류를 구별한다. 후보 수/직렬화 총량 한도는 기존 4096개·8MiB 내에서 함께 계산하도록 정한다. AC의 정상 동시성 대조군을 삭제하지 않는다.

### RPA-A1 — receipt lookup miss와 명시 branch의 판정 표 보강

- 위치: design.md:414, :416, :418–419.
- Severity: minor. Class: optional. Confidence: medium. Receipt 구현 착수 차단: no.
- :418은 필수 prefix를 재구성할 수 없는 편집을 명시 오류로 정하므로 요구되는 결과 자체는 있다. 다만 :414의 “모든 해당 assistant prefix”가 lookup hit만 뜻하는 구현이 되지 않도록, 등록되지 않은 assistant prefix와 아직 소비되지 않은 응답을 포함하지 않는 정상 branch를 구별하는 짧은 판정 표가 있으면 명확하다. foreign 응답을 receipt chain에 연결하는 경우도 같은 표에 넣을 수 있다.
- 이 항목을 별도 상태 DB나 모든 transcript의 gateway 복사 요구로 확대하지 않는다. 실제 구현 결함은 관측하지 않았다.

## Category Scores

| Dimension | Score | Rubric band | Evidence |
|---|---:|---|---|
| Clarity | 0.50 | 여러 핵심 해석이 남음 | design:411의 귀속 수명, :413/:416의 빈 후보 조건 |
| Completeness | 0.75 | 제한된 설계 공백 | design:411–420의 10개 계약 중 native 저장 lifecycle 미정; 나머지 경계는 구체화됨 |
| Testability | 0.50 | 수용 결과 상충 | acceptance:174와 design:416의 동일 공개 답변 병렬 후보 |
| Traceability | 1.00 | 변경분 REQ/AC 직접 연결 | spec:520–523 → acceptance:169–175 → design:399–434 및 plan:219–222 |

평균 0.6875. 선택 사항 RPA-A1의 수나 분량을 FAIL 근거로 사용하지 않았다. blocking RPA-1/RPA-2의 정확성과 내부 일관성으로 FAIL이다.

## Must-Pass Results — 변경분 범위 표시

- MP-1/MP-2/MP-3/MP-4: 전체 코어 번호·GEARS·frontmatter·언어 중립성은 이 bounded receipt 감사에서 재판정하지 않았다. UNVERIFIED이며 전체 PASS로 전용할 수 없다. REQ-MG-015의 추가 요구층은 spec:520–523, 해당 검증층은 AC-MG-009:169–175로 구분해 읽었다.
- MP-5: 이 변경분의 PICKER 연계 문서를 읽었으나 전체 참조 SPEC 재고는 수행하지 않았다. 전체 D7은 UNVERIFIED.
- MP-6: Windows 의미와 기존 syscall 계약은 위임에서 제외되었다. 전체 D8은 UNVERIFIED이며 Windows PASS 주장이 아니다.
- MP-7: **전체 코어 기준 FAIL**. plan.md:224의 `[NEEDS CLARIFICATION: GPT 구독 요청별 출력 상한]`을 이 실행에서 확인했다. receipt 수정 요구로 전용하지 않았으며 사용자 답변 없이 해제하지 않는다.
- MP-8: receipt 변경 AC는 :175에서 활성화 BLOCKED이며 본 감사는 아직 구현하지 않은 설계의 수용 모순을 다룬다. 전체 release-blocking RED-now 재실행은 수행하지 않았다. 전체 MP-8은 UNVERIFIED. 아래 재계산은 raw 관측 검산이며 RED-now의 대체물이 아니다.

## 확인한 설계 경계

| 경계 | 변경분 판정 | 근거와 범위 |
|---|---|---|
| 최소 prefix 정규화 | PASS, 한 raw 양성군 한정 | design:412. 이번 Python 재계산에서 request002 입력 prefix가 003에 유지되고 003의 5 messages가 004에 유지됨 |
| no-tool 전체 삭제 탐지 계약 | PASS, 설계 의무 한정 | :413–414는 공개 prefix와 required envelope digest/item 수를 결합. 실제 삭제 시험은 미실행 |
| 부분/전체 receipt 소실 | PASS, 설계 의무 한정 | :417은 resume missing/corrupt를 빈 저장소로 초기화하지 않으며 manifest count/generation/content digest 확인 |
| 악의적 전체 rollback 한계 | PASS, 명시된 한계 | :417/:429는 일관된 과거 snapshot/공개 history+store 전면 교체를 hash 인증으로 주장하지 않음 |
| publish/terminal 실패 | PASS, 설계 의무 한정 | :396–397의 전체 terminal 검증과 :415의 durable publish-before-success, 실패/취소/불완전 receipt 금지 및 재시도 금지 |
| foreign strip | PASS, 설계 의무 한정 | :419는 귀속 검사 후 OpenAI envelope만 제거하고 공개 pair 원래 ID 복원; 실제 provider 전환은 미실행 |
| 유한 저장 | PASS, 설계 의무 한정 | :420은 4096개·8MiB 초과 오류와 필수 receipt 비자동폐기 명시. 정상/경계 runtime 미실행 |

이 표의 PASS는 **문서의 특정 경계가 명시되었다는 뜻**이다. 전체 설계, 구현 또는 실제 provider 수용 PASS가 아니다.

## Evidence

이 감사에서 직접 실행한 기준 명령과 원문 출력:

```text
$ moai session current
a28010b3-3a55-45d2-9a33-0fe61c2a5e44
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse --show-toplevel
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified branch --show-current
WT-unified-gateway
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified check-ignore .moai/reports/SPEC-MOAI-GATEWAY-001/receipt-plan-audit-iter1.md
[stdout empty; exit 1: ignored 경로 아님]
```

읽기는 `sed -n '374,475p' design.md`, `sed -n '1022,1095p' research.md`, `sed -n '145,190p' acceptance.md`, `sed -n '510,530p' spec.md`, `nl -ba plan.md`의 210–231행과 `rg -n 'receipt|manifest|UUID|CLAUDE_CONFIG_DIR|retained|transcript|대화.*폐기'`로 해당 코어 다섯 문서 전체 관련 위치를 대조했다. 경로는 모두 아래 Baseline-attribution의 SPEC 디렉터리다. 관측 보고서 네 개를 읽었으며 제안서의 작성자 판단/이유는 감사 판정에서 제외했다. after19-contract-proposal.md의 raw 비교 절(182–196행)만 별도 검산 근거로 사용했다.

실제로 실행한 raw 검산 코드(인증/header/본문 문자열 출력 없음):

```python
import json,pathlib,hashlib,copy
b=pathlib.Path('/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/state/verify/01a08e7b-6aa0-7361-ab7e-ea8da1f02228/gateway-entry/picker-resume-20260911T103342Z-9e003c97/raw')
r=[json.loads((b/f'request-{i:03}.json').read_text()) for i in (2,3,4)]
def norm(ms):
 out=copy.deepcopy(ms)
 for m in out:
  if isinstance(m['content'],str):m['content']=[{'type':'text','text':m['content']}]
  m['content']=[x for x in m['content'] if x.get('type') not in ('thinking','redacted_thinking')]
  for x in m['content']:x.pop('cache_control',None)
 return out
for i,x in zip((2,3,4),r):
 print('request',i,'roles',[m['role'] for m in x['messages']],'blocks',[[z.get('type') for z in m['content']] if isinstance(m['content'],list) else ['string'] for m in x['messages']])
a=norm(r[1]['messages']);c=norm(r[2]['messages'])
print('resume_prefix_equal',a==c[:len(a)])
print('resume_prefix_sha256',hashlib.sha256(json.dumps(a,sort_keys=True,ensure_ascii=False,separators=(',',':')).encode()).hexdigest())
x=norm(r[0]['messages'])
print('input002_len',len(x),'next003_prefix_equal',x==a[:len(x)])
for i,(u,v) in enumerate(zip(x,a)):
 if u!=v:print('changed_input_prefix_index',i,'role_before',u['role'],'role_after',v['role'])
print('metadata_uuid_equal',len(set(json.loads(q['metadata']['user_id'])['session_id'] for q in r))==1)
for i in (3,4):print('raw_sha',i,hashlib.sha256((b/f'request-{i:03}.json').read_bytes()).hexdigest())
```

`python3 - <<'PY'` 단일 invocation으로 위 코드를 실행한 원문 출력; exit 0:

```text
request 2 roles ['user', 'system'] blocks [['text', 'text', 'text', 'text', 'text'], ['text']]
request 3 roles ['user', 'system', 'assistant', 'user', 'system'] blocks [['text', 'text', 'text', 'text', 'text'], ['string'], ['redacted_thinking', 'tool_use'], ['tool_result'], ['text']]
request 4 roles ['user', 'system', 'assistant', 'user', 'system', 'assistant', 'user', 'system'] blocks [['text', 'text', 'text', 'text', 'text'], ['string'], ['redacted_thinking', 'tool_use'], ['tool_result'], ['string'], ['text'], ['string'], ['text']]
resume_prefix_equal True
resume_prefix_sha256 fa2c3d9ddfa7b58bdc5d6f5b4e89793a7fc12555899cfdff90b00f46c288cbfd
input002_len 2 next003_prefix_equal True
metadata_uuid_equal True
raw_sha 3 e5ef6a87b49ba353d58b2414f40163c6f4694b609e14957fd6a35045c2f76908
raw_sha 4 92b8896c0ef8493deff8c3560e5af69f8c97a3baf08613d0bd0b258aa3b7d2c6
```

이 raw 비교는 원래 marker ID를 그대로 둔다. **가역 ID decoder의 올바름을 증명하지 않는다.** object key·공백·문자열/단일 block·cache_control 최소 정규화로 관측된 prefix가 유지된다는 검산이다.

## Baseline-attribution

SPEC 디렉터리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/specs/SPEC-MOAI-GATEWAY-001`.
위 HEAD의 작업 중인 문서에 대한 SHA-256을 Python hashlib로 직접 측정했다:

```text
spec.md d3acee5b26190d61aa7e426cd21845be4134d27b4f91d8a64b61ed953b33260d
design.md eb70fe3d6e69f3901d66420450a6cd28ed8e3461f8c9a39c953061c18a014188
research.md a88081252492de6957297f0eeaff466ab8c1c7c82fc741695f91d314cf8815ee
acceptance.md 5ae103db444df2fcd2a333468afd8510e3897c5b4f5c4056bf049c4da5e814d8
plan.md 7a83644cdff02c75ce64d59f43ef49ce54c3f7cdbe443f0815af756e03b1436a
clarification_markers [(224, '- [NEEDS CLARIFICATION: GPT 구독 요청별 출력 상한] max_output_tokens 거절의 실측은 있으나 사용자 답변 전 기존')]
receipt_report_exists False
```

부모가 제공한 source_session_id는 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이며 raw 경로의 UUID다. 이번 감사 런타임이 반환한 child UUID와 구별한다. 제품 코드·SPEC·원격 상태를 수정하지 않았다. 새 감사 보고서 한 파일만 작성한다.

## Gaps

새 native/provider 실행, receipt 코드 실행, 모든 REQ/AC 전체 감사, native continue·fork·interactive resume, 실제 encrypted carrier의 Claude 왕복·삭제·교차 provider·디스크 실패·Windows runtime은 이번 감사에서 수행하지 않았다. 기존 관측 보고서의 runtime 수치를 이번 감사의 재실행 수치로 표현하지 않는다. private config+secure-storage 관측은 실제 OAuth 공존을 보여 주지만 retained transcript 재개의 통합 증거가 아니다.

## Residual-risk

hash-only 설계는 명시한 공개 prefix·private 저장 경계 안의 일관성 검사다. 원문 공개 history와 저장소를 함께 교체하는 악의적 rollback에 대한 인증 수단이 아니다. 두 blocking 수정 뒤에도 설계 PASS와 제품 활성화는 다르다. design:431–434의 실제 carrier/tool/no-tool/resume/foreign/mutant/네 GPT 후속 게이트는 그대로 남는다.

## Recommendation

1. RPA-1의 UUID·retained native storage lifecycle 표를 core/PICKER 소관에 일관되게 기록한다.
2. RPA-2의 정상 후보/삭제·혼합/빈 후보 수용 표를 design 및 AC-MG-009에 같은 결과로 반영한다.
3. 재감사는 위 두 항목과 그 수정으로 바뀐 기존 경계만 확인한다. RPA-A1은 선택 사항으로 남긴다. 출력 상한·Windows 의미의 사용자 결정을 이 감사가 대신하지 않는다.

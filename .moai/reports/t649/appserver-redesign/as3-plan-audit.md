# AS3 수명·권한·복구 변경분 독립 감사

Iteration: 2/3 (AS3-B1 수정분 재감사)
Verdict: PASS
Overall Score: 1.00 (범위 한정)

## Claim

SPEC-MOAI-GATEWAY-001 0.11.0의 AS-007~009 수명 계약은 설치된 App Server 규격에 대응할 수 있다. 최초 감사에서 발견한 인증 진입 계약 충돌 AS3-B1은 수정 문서의 명시적 경계 분리로 해소되었다. 수정분 및 기존 수명 계약의 관련 회귀에서 남은 차단 결함을 발견하지 못했다. 전체 SPEC 감사나 AS1/AS2 구현 감사는 수행하지 않았다.

Reasoning context ignored per M1 Context Isolation. 구현 제안은 계약의 부재를 메우는 근거로 쓰지 않았다. 읽기 전 실패 가능성으로 HTTP 종료와 turn 종료 혼동, 결과 교차 소비, 중복 실행, 프로세스 교체 뒤 RPC 재사용, 취소 전파 범위, 성공 terminal 오합성 및 기존 인증 진입 조건 충돌을 정했다.

## Defects Found

현재 미해결 결함: 0건. 아래 D1은 최초 감사에서 확인한 결함 이력이며 iteration 2에서 RESOLVED로 판정했다.

D1. AS3-B1 — spec.md:L752–761; design.md:L70–105 — 활성 REQ-MG-023은 모든 route가 provider credential seam에서 credential을 얻어야 한다고 규정한다. design §2.1의 GPT 구체 타입 부재는 GPT 요청의 명시 거절 조건이다. 반면 REQ-MG-017(spec.md:L579–583)은 App Server managed 인증을 채택하고 MoAI의 구독 토큰 읽기를 금지한다. design.md:L381의 대체 선언은 직접 Responses·opaque/receipt 설계를 가리키며 REQ-MG-023과 §2.1의 진입 조건을 정리하지 않는다. 계약 그대로라면 managed 인증이 있어도 credential 참조 부재로 거절하거나 의미 없는 CredentialRef를 만드는 두 구현으로 갈린다. Severity: major. Class: blocking. Required fix: managed route의 공식 인증 상태·계정/프로필 귀속을 검증하는 권한 경로와 credential route의 기존 seam을 분리하여 REQ-MG-023, 선택 시점 로컬 검증, design §2.1/ingress 및 관련 AC에 일관되게 반영한다. 기존 GLM/Anthropic credential 검사를 전역 우회해서는 안 된다.

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 1.00 | 1.00 | 수정 spec.md:L752–763의 직접 credential/App Server 권한 경계 구분 |
| Completeness | 1.00 | 1.00 | 수정 design.md:L70–77, L343–350 및 acceptance.md:L618–624의 권한 계약과 양성·음성 대조군 |
| Testability | 1.00 | 1.00 | acceptance.md:L569–579의 결과 귀속·재실행 금지·성공 terminal 0 |
| Traceability | 1.00 | 1.00 | AS-007~009가 REQ-MG-015/017을 명시 |

## 다른 AS3 점검

- PASS: spec.md:L552–556과 design.md:L438–443은 HTTP tool segment 종료 뒤 pending RPC를 유지하고 다음 결과를 기존 turn에 연결한다. 정상 HTTP 종료가 turn 취소라는 계약은 없다.
- PASS: acceptance.md:L569–571은 동시 대화 A/B 교환 시 서로의 call 소비 0과 동일 결과 재전송의 비재실행을 요구한다. 같은 인수로 생성한 별개의 정상 call을 중복 RPC라고 판정할 근거는 아니다.
- PASS: acceptance.md:L573–575는 실행 전·실행 후·결과 저장 후 crash를 구분하고 복구 불가능한 RPC의 명시 실패를 허용한다. 새 process에 옛 RPC 응답 주입이나 실행 여부가 불명확한 도구 재실행은 금지한다.
- PASS: acceptance.md:L577–579는 해당 turn만 interrupt하고 뒤늦은 결과의 다른 turn 진행을 금지한다. 설치 TurnInterruptParams.required는 threadId와 turnId를 제공한다.
- PASS: design.md:L448은 thread/turn/call ID, 반영 prefix digest, 결과 수신·소비 상태의 대화별 원자 저장을 요구한다. 구체 저장 순서와 queue 구현은 이번 계획 감사의 실행 증거가 아니다.

## Must-Pass 범위

MP-2의 AS3 요구층은 spec.md:L552의 Where 및 L555의 When/shall 구조로 확인했다. AC의 Given/When/Then은 검증층의 올바른 형식이다. MP-8은 검토한 AS3 시나리오에 release-blocking 분류와 RED-now 명령이 없어 N/A다. 나머지 전체 문서 must-pass는 이번 제한 감사의 범위 밖으로 UNVERIFIED이며 전체 SPEC PASS를 부여하지 않는다. D1은 평균으로 상쇄하지 않았으며 아래 수정 근거로만 해소했다.

## Evidence

실행 명령:

```python
python3 - <<'PY'
import json,pathlib,hashlib
r=pathlib.Path('/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified');s=r/'.moai/specs/SPEC-MOAI-GATEWAY-001';p=r/'.moai/reports/t649/appserver-redesign/probe/schema'
for f in ['DynamicToolCallParams.json','DynamicToolCallResponse.json','v2/TurnInterruptParams.json','v2/TurnCompletedNotification.json']:
 d=json.loads((p/f).read_text());print(f,'required='+','.join(d.get('required',[])))
for file,lo,hi in [('spec.md',752,764),('design.md',70,106),('acceptance.md',569,579)]:
 txt=(s/file).read_text();print('SHA256',file,hashlib.sha256(txt.encode()).hexdigest())
 for n,line in enumerate(txt.splitlines(),1):
  if lo<=n<=hi:print(f'{file}:{n}: {line}')
PY
```

Verbatim stdout에서 규격 부분, exit code 0:

```text
DynamicToolCallParams.json required=arguments,callId,threadId,tool,turnId
DynamicToolCallResponse.json required=contentItems,success
v2/TurnInterruptParams.json required=threadId,turnId
v2/TurnCompletedNotification.json required=threadId,turn
```

같은 실행에서 판독한 결함 문장:

```text
spec.md:754: 중립 credential 참조 seam을 통해서만 해석해야 하며, 그 seam이 해당 provider의 credential을
spec.md:755: 돌려주지 못하는 경우도 이 거절에 해당한다.
spec.md:761: - 모델의 route가 credential 참조 seam에서 credential을 얻지 못하면 HTTP 401 `authentication_error`로 답한다.
design.md:102: | GPT (PKCE / API key) | 형제 SPEC | 인터페이스만 존재. 구현 전까지 `ErrCredentialAbsent` → `REQ-MG-023`의 명시 거절 |
design.md:104: GPT 구체 타입이 없는 동안 GPT 모델 요청은 외부 송신 없이 명시 오류로 끝난다. 이것은 결함이
```

## Baseline-attribution

작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`. 문서가 미커밋 상태이므로 내용 SHA-256으로 감사 입력을 고정한다.

- spec.md: `b0eb66bfee1650041c232cd1deb576a5b865ee7f6ebd0e1545bd8738bb8d6310`
- design.md: `fe12dba7147cef3be64280b90eec10f978a4fd96ea08b4a125e44ef5837bc75f`
- acceptance.md: `71cd4e0da5a0b411e615121dfdc3d782f688e017e4d75715376032acb1d8d4ed`

규격은 probe/schema에 보존된 Codex 0.154.0 자료를 이번 실행에서 파싱했다. 이번 감사에서 schema를 새로 생성하지 않았다.

## Gaps

실제 HTTP 분리 왕복, queue overflow, 프로세스 강제 종료, 결과 저장 내구성, 취소 경쟁, API/managed auth 실제 계정 실행을 하지 않았다. 계획상의 시험 가능성과 runtime 안전성은 별개다. 전체 Tier L 문서 감사 및 기존 코드 검증은 범위 밖이다.

## Residual-risk

HTTP 완료와 사용자 취소를 같은 context 수명에 묶거나 pending 식별자를 저장하기 전에 Claude 실행이 시작되면 계약과 다른 동작이 생길 수 있다. 이 가능성은 코드 결함으로 판정하지 않았으며 AS-007~009의 구현 시험에서 확인해야 한다.

## Iteration history

1. FAIL, 0.8125: AS3-B1 인증 진입 계약 충돌 발견. 다음 감사는 이 결함의 수정과 관련 회귀만 확인한다.
2. PASS, 1.00: AS3-B1 RESOLVED. 초기 FAIL의 원문·입력 해시를 위에 보존하고 수정 근거를 아래에 기록한다.

## Regression Check — iteration 2

**AS3-B1 RESOLVED.** 수정 spec.md:L752–763은 직접 secret route의 CredentialRef 검증과 App Server의 관리 세션 권한 검증을 분리한다. 권한 항목은 profile·account·auth mode·generation·allowed model이며 누락·불일치를 명시 거절한다. design.md:L70–77과 L106–110은 App Server API-key 세션도 이 경계를 사용하며 가짜 CredentialRef, 빈 secret, 직접 Apply가 필요하지 않음을 명시한다. logout·계정/프로필 변경 시 무효화와 실제 turn 직전 세대 검사가 추가되었다.

선택 검증의 회귀도 확인했다. design.md:L339–350은 catalog 밖 ID 404를 먼저 판정하고 관리 권한 실패 401을 판정하며 생성·refresh·유료 호출 0을 유지한다. acceptance.md:L618–624의 양성 대조군과 필드별 변조·logout 세대 변경 음성군이 이 계약을 확인한다. 직접 API·GLM의 기존 credential 부재/세대 변경 시험을 유지하므로 관리 경계 도입이 전체 인증 우회를 허용하지 않는다.

AS-007~009의 HTTP 경계·결과 단일 소비·crash 비재실행·해당 turn 취소 조항도 재판독했으며 변경 전 계약이 유지된다. 코드를 실행해서 이 동작을 확인한 것은 아니다.

### 수정분 Evidence / Baseline-attribution

실행 명령:

```python
python3 - <<'PY'
from pathlib import Path
import hashlib
s=Path('/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/specs/SPEC-MOAI-GATEWAY-001')
texts={n:(s/n).read_text() for n in ['spec.md','design.md','acceptance.md','plan.md']}
assert 'App Server 경로는 secret 대신 관리 세션 권한을 검증해야 한다.' in texts['spec.md']
assert '실제 turn 직전에도 세대를 재검사한다.' in texts['design.md']
assert '토큰 파일을 읽거나 Apply를 호출하지 않는다.' in texts['design.md']
assert '직접 API·GLM credential 부재/세대 변경의 기존 음성군은 계속 통과해야 한다.' in texts['acceptance.md']
for n,t in texts.items():print(n,hashlib.sha256(t.encode()).hexdigest())
print('AS3_B1_CONTRACT_DELTA=PASS')
PY
```

Verbatim stdout, exit code 0:

```text
spec.md 6d54a5e9415c588f320283883a5b766f0cd35dea685bde78e7da4f7de3c71265
design.md cdb78efa528caa63dd4005bedb308a185035b8e258393df1561a509609c585ab
acceptance.md 35386461fb3a7b271384c8f4e6976b5adb296996fdc4d070cf2094a6bb6b2e8b
plan.md 7521f77145f35d8e482c4d53224c184536d216258fb2d4f8c2b6c591a8d2783e
AS3_B1_CONTRACT_DELTA=PASS
```

이 텍스트 검사는 변경 문장 존재와 감사 입력 해시의 기계 증거다. 의미 판정은 위의 문서 직접 대조로 내렸으며, runtime 성공을 뜻하지 않는다. Gaps와 Residual-risk는 최초 감사에 명시한 그대로 남는다.

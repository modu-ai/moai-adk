# AS2 혼합 도구 연결 변경분 독립 감사

Iteration: AS2 승인 변경분 1회
Verdict: PASS
Overall Score: 1.00 (아래 AS2 변경분에만 적용)

## Claim

SPEC-MOAI-GATEWAY-001 0.11.0의 **초기 완전 정의 도구 native 등록 + 후발 발견 도구 dispatcher** 계약과 AS-004의 여섯 시험군에서, 설치 규격으로 구현할 수 없게 만드는 모순이나 차단 결함을 발견하지 못했다. 이 판정은 t651 설계 변경분에 한정한다. 전체 SPEC 감사, 제품 동작 PASS, 카드 완료 또는 출시 승인을 뜻하지 않는다.

Reasoning context ignored per M1 Context Isolation. 작성자의 판단을 근거로 수용하지 않고 최종 문서와 저장된 기계 판독 자료를 대조했다. decision-record는 승인 범위 확인에만 사용했다.

읽기 전 점검할 실패 가능성을 정했다: 초기/후발 경로 혼동, 이름·예약 namespace 충돌, schema 없는 참조 승격, 변경된 schema의 기존 호출 적용, 불완전 인수 검증, 다른 대화 결과 소비, 재귀·우회 호출, 도구 등록을 위한 thread 재생성, 단일 dispatcher 실험을 hybrid 실증으로 오인하는 경우다.

## 변경분 판정

| 점검 | 판정 | 구체적 근거 |
|---|---|---|
| ToolSearch 유지와 초기/후발 구분 | PASS | spec.md:L593–597은 초기 도구 native 등록, 후발 도구만 dispatcher, 초기 도구 우회 금지, ToolSearch 유지와 별도 제품 검증을 명시한다. |
| 고정 도구 목록으로 구현 가능한 경로 | PASS | design.md:L407–409는 처음에 native 도구들과 dispatcher 하나를 함께 등록한다. 설치 schema의 ThreadStartParams에는 dynamicTools가 있고 TurnStartParams/ThreadResumeParams에는 없다. 후발 도구 추가 시 공식 목록 변경 기능을 요구하지 않는 설계다. |
| 이름 충돌·미등록 참조 | PASS | acceptance.md:L546–547은 충돌·중복·전체 schema 없는 참조의 실행 계수 0을 요구한다. design.md:L412–414는 typed 정의와 발견 결과를 함께 확인하고 기존 이름 덮어쓰기를 거절한다. |
| schema 고정·인수 검사 | PASS | design.md:L411–415는 대화별 digest/epoch와 pending snapshot, 중복 key 거절, dialect/$ref 범위 선언을 요구한다. acceptance.md:L548–550은 schema 변경과 required·타입·enum·중첩·additionalProperties·미지원 dialect/$ref를 각각 판정한다. |
| 결과 소유권·중복 소비 | PASS | acceptance.md:L551, L555–556은 대화·thread·turn·call 귀속 및 결과 단일 전달을 요구한다. 설치 DynamicToolCallParams.required에 arguments, callId, threadId, tool, turnId가 있어 결박 대상 식별자가 존재한다. |
| thread 유지·우회·재귀 | PASS | design.md:L409, L416과 acceptance.md:L552–553은 등록을 위한 thread/start·fork·resume 0 및 초기 native 도구의 dispatcher 우회·dispatcher 재귀 실행 0을 요구한다. |
| 검증 범위의 정직한 표시 | PASS | design.md:L420–427은 단일 dispatcher 가능성, hybrid 제품 실증, native 실행 격리를 분리한다. 저장된 probe는 실제로 ToolSearch도 dispatcher로 보냈으므로 이 구분이 필요하며 문서가 이를 지킨다. |

## Category Scores

| Dimension | Score | Rubric Band | Evidence |
|---|---:|---:|---|
| Clarity | 1.00 | 1.00 | spec.md:L593–597, 초기와 후발 경로의 배타적 구분 |
| Completeness | 1.00 | 1.00 | spec.md:L585–597, design.md:L407–416, AS2 요구·설계·검증 연결 |
| Testability | 1.00 | 1.00 | acceptance.md:L545–553, 여섯 군의 실행·소비·RPC 계수와 명시 오류 |
| Traceability | 1.00 | 1.00 | acceptance.md:L540, L555, L559가 모두 REQ-MG-015/017을 명시; 관련 hybrid 의무는 spec.md:L593–597 |

## Must-Pass 범위

- MP-2 변경 요구층: PASS. spec.md:L593의 `The adapter shall`과 L587–588의 `When ... the adapter shall`을 요구층에서 판정했다. Given/When/Then인 acceptance.md:L540 이하를 GEARS 결함으로 세지 않았다.
- MP-8 AS2: N/A. 검토한 AS2 시나리오는 `release-blocking`으로 분류되어 있지 않고 RED-now 명령을 싣지 않는다. 따라서 재현한 RED가 있다고 주장하지 않는다. 실행 게이트 자체는 acceptance.md:L519–522에 남아 있다.
- MP-1/3/4/5/6/7 및 전체 Tier L 문서 검토: **이번 변경분 감사의 범위 밖이며 UNVERIFIED**. 이를 전체 SPEC의 must-pass 통과로 해석하면 안 된다. 전면 감사의 PASS를 발급하지 않는다.

## Defects Found

No defects found within the bounded AS2 delta. 차단 결함 0건. 더 넓은 범위의 미검증을 이 변경분의 결함으로 만들어 점수를 낮추지 않았다.

## Evidence

실행한 명령:

```python
python3 - <<'PY'
import json,pathlib,re
r=pathlib.Path('/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified')
p=r/'.moai/reports/t649/appserver-redesign/probe'
a=(r/'.moai/specs/SPEC-MOAI-GATEWAY-001/acceptance.md').read_text()
b=a[a.index('AS-004의 여섯'):a.index('**AC-MG-007 / AS-005')]
assert re.findall(r'^([1-6])\. Given ',b,re.M)==list('123456')
print('AS004_NEGATIVE_GROUPS=1,2,3,4,5,6')
for name in ['ThreadStartParams','TurnStartParams','ThreadResumeParams']:
 d=json.loads((p/f'schema/v2/{name}.json').read_text())
 print(name, 'dynamicTools='+str('dynamicTools' in d['properties']))
d=json.loads((p/'schema/DynamicToolCallParams.json').read_text())
print('DYNAMIC_CALL_REQUIRED='+','.join(d['required']))
d=json.loads((p/'dispatcher-result.json').read_text())
print('PROBE_SUCCESS='+str(d['success']))
print('PROBE_ROUTES='+','.join(x['tool']+':'+x['arguments']['name'] for x in d['tool_calls']))
print('AS2_STRUCTURAL_CHECK=PASS')
PY
```

Verbatim stdout, exit code 0:

```text
AS004_NEGATIVE_GROUPS=1,2,3,4,5,6
ThreadStartParams dynamicTools=True
TurnStartParams dynamicTools=False
ThreadResumeParams dynamicTools=False
DYNAMIC_CALL_REQUIRED=arguments,callId,threadId,tool,turnId
PROBE_SUCCESS=True
PROBE_ROUTES=moai_claude_tool:ToolSearch,moai_claude_tool:fixture_echo
AS2_STRUCTURAL_CHECK=PASS
```

`git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified check-ignore .moai/reports/t649/appserver-redesign/hybrid-plan-audit.md` returned exit 1, stdout empty: 내보낼 경로가 gitignored 경로가 아님을 확인했다.

## Baseline-attribution

- 측정 작업 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`.
- 이번 실행의 `git rev-parse HEAD` stdout: `81c1d58f9cf7045594ee61d5e4ff380948ce9eba`.
- 문서는 미커밋 변경이므로 HEAD만으로 내용이 고정되지 않는다. 이번 읽기에서 SHA-256도 확인했다.

| 문서 | SHA-256 |
|---|---|
| spec.md | `09b49fb7d756e5cbfd72332181f83db1ae5bfd28e228da6b9e6211764d6e6c99` |
| design.md | `f8f90fa1cec48ba97b417b61a2b749c1943993f79c3ece8906aa36a94d181e7b` |
| acceptance.md | `3e439544751188d42e595188b3f4b764b06adfffaf799d0304e82c29b4b7bd7e` |
| plan.md | `ba4865ae63b2a17e18fee6d88e44fcd4b38fb135341f0739e2dea9ed2afdf8b4` |

- `moai session current` stdout: `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`.
- schema/probe는 기존 실행의 저장 자료를 이번 감사에서 파싱했다. 이 감사에서 Codex schema를 재생성하거나 실제 구독 호출을 다시 실행하지 않았다.

## Gaps

- 실제 Claude → MoAI hybrid → App Server → Claude 왕복은 이번 감사에서 실행하지 않았다.
- 여섯 군의 실행 가능한 테스트와 양성 대조군은 이 문서의 요구이며 이번 감사의 실행 결과가 아니다.
- validator의 구체 dialect와 $ref 지원 집합, 모델이 후발 schema 설명을 받아 올바른 인수를 만드는 성능, native 도구 실행 0은 구현 단계 검증 대상이다.
- 기존 probe의 ToolSearch도 dispatcher 경로였다. 해당 기록은 초기 ToolSearch native + 후발 도구 dispatcher 조합의 성공을 증명하지 않는다.
- 전체 Tier L SPEC의 5개 문서를 전수 감사하지 않았다. 현재 보고서로 전면 계획 감사나 기존 미해결 판정을 덮어쓸 수 없다.

## Residual-risk

규격상 연결 가능한 계획이라는 판정과 실제 제품의 도구 실행 안전성은 별개다. 문서대로 실행 전 검증·snapshot 고정·결과 단일 소비를 구현하지 못하면 같은 규격을 쓰더라도 잘못된 도구 실행이나 대화 간 혼입이 발생할 수 있다. 이 잔여 위험의 해소 판정은 AS-004~006 제품 시험 및 native 실행 격리 시험에 남긴다.

## Recommendation

AS2 변경분은 구현할 수 있다. 이 보고서의 한정 PASS를 카드 완료나 전체 SPEC PASS로 승격하지 말고, acceptance.md:L540–561에 적힌 제품 양성·음성 시험으로 구현 결과를 별도 판정한다.

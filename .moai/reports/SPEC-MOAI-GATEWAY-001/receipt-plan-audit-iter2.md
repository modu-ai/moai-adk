# Receipt 설계 변경분 독립 재감사

Iteration: 2/3 (receipt 변경분)
Verdict: PASS
Overall Score: 1.00
Scope: RPA-1·RPA-2 및 선택 RPA-A1의 수정과 그에 따른 경계만

Reasoning context ignored per M1 Context Isolation.

## Claim

**앞선 receipt 변경분 감사의 두 blocking 항목은 문서에서 해소되었다.** launcher의 사전 UUID 귀속과 retained native 저장 수명이 구체화되었고, 같은 공개 prefix의 정상 A/B 후보와 오류 후보의 수용 결과가 design과 AC에서 일치한다. 선택 사항 RPA-A1의 lookup miss/branch 표도 반영되었다. 이 변경분의 계획 감사 게이트는 해소할 수 있다.

이 PASS는 **미구현 receipt 설계 변경분**에 한정한다. 전체 코어/PICKER SPEC 감사, 구현 검증, native fork/continue runtime, 사용자 승인 대기 항목, reasoning 제품 활성화 또는 릴리스 PASS로 사용할 수 없다. 첫 감사에서 이미 확인한 전체 코어 MP-7의 출력 상한 clarification은 이 변경분의 범위 밖이며 해소 판정하지 않았다.

재감사 전 확인 대상으로 설정한 회귀 가설: UUID를 첫 요청에서 등록하도록 되돌림, retained transcript의 종료 시 삭제, fork 원문을 MoAI receipt로 복사, parent 갱신이 child snapshot을 덮어씀, family lease가 모든 gateway 동시 요청을 금지, 정상 A/B를 여전히 송신 0으로 요구, required=false 빈 후보 오판, lookup miss를 새 대화로 수용, 미실행 native fork를 완료로 표시. 아래 위치에서 각각 확인했다.

## Regression Check

| 이전 항목 | 판정 | 직접 확인한 본문 근거 |
|---|---|---|
| RPA-1 사전 UUID와 retained native 수명 | RESOLVED | core design.md:411, :424–465; core acceptance.md:528–533; PICKER plan.md:98–110, acceptance.md:62–70 |
| RPA-2 후보 수용 상충 | RESOLVED | core design.md:470–483, acceptance.md:514–526. 두 표 모두 8행이며 같은 순서의 PASS/PASS/FAIL/PASS/PASS/FAIL/PASS/FAIL |
| RPA-A1 lookup miss/정상 branch | RESOLVED, optional 반영 | core design.md:487–494, acceptance.md:535–537 |

### RPA-1 확인 결과

- **new:** design:433에서 launcher가 기동 전 UUID 및 초기 manifest를 생성하고 `--session-id`와 private bootstrap에 같은 UUID를 넣는다. 실패 시 child/송신 없음이 명시되었다.
- **exact resume:** :434에서 소유 사용자·프로젝트 인덱스와 manifest/transcript를 검증한 뒤 정확한 UUID만 인계한다. :411은 metadata로 root를 선택하거나 최초 UUID를 등록하는 동작을 금지한다.
- **continue/선택형 resume:** :435–436에서 launcher가 먼저 소유 인덱스의 UUID를 확정하고 native exact resume로 인계한다. 선택 취소·후보 부재·손상과 egress 허용 시점이 구분되어 있다.
- **fork:** :437–442에서 parent/child UUID 사전 확정, native 인수 조합의 preflight, child만 허용한 bootstrap, parent hash-only snapshot의 child 고정 및 개별 후보 한도 포함을 명시한다. parent 갱신/삭제가 child 귀속을 손상시키지 않으며, transcript owner/root/UUID 검증 전 완료 인덱스로 채택하지 않는다. **미확인 native 인수 조합은 오류·미완료이고 필수 시험은 유지**한다.
- **저장 자료:** :424–429는 Claude가 직접 쓰는 native settings/transcript와 MoAI의 UUID별 hash receipt를 구별한다. MoAI 인덱스는 family/UUID·프로젝트 식별·transcript 식별 경로·완료 순번만 보유한다. 원본 사용자 transcript·credential 복사나 opaque 원문의 receipt 영속화를 새로 허용하지 않는다.
- **수명/병렬:** :444–448은 같은 family의 lead 쓰기 경합을 lease/busy로 처리하고, 다른 family 및 한 lead의 병렬 gateway 요청을 구분한다. :452–454는 정상 종료/취소/crash 때 retained 자료를 보존하며, 명시 폐기 때도 살아 있는 sibling/fork의 공유 root를 삭제하지 않는다.
- **인증 namespace:** :456–465는 private config 변경 전 원래 secure-storage override를 빈 문자열까지 보존하고 원래 config/default namespace를 선택한다. credential은 native가 사용한다. 명시 빈 config, 다른 버전/플랫폼/환경은 preflight 게이트를 유지한다.
- **근거 경계:** research.md:1048–1057은 private config/OAuth 1회 관측과 동일 config의 exact resume가 별도 실행임을 명시한다. :1056 및 PICKER acceptance:69–70에서 합친 통합, continue/선택/fork가 여전히 실제 시험 대상이다. 새 native 실행 PASS를 문서 보정으로 만들어 내지 않았다.

### RPA-2 확인 결과

design:476–483과 acceptance:519–526의 8행을 전부 대조했다.

| 조건 | 두 문서의 일치 결과 |
|---|---|
| 등록된 required=true A/B 중 정확한 A | 정상 송신, A 복원 |
| 등록된 required=true A/B 중 정확한 B | 정상 송신, B 복원 |
| 혼합·변조·미등록 C·필수 envelope 전체 삭제 | 송신 0 |
| 동일 후보 재publish/재전송 | idempotent, 후보 증가 없이 정상 |
| required=false만 있고 envelope 없음 | 정상, opaque 추가 없음 |
| required=true/false 혼재, envelope 없음 | 모호성 오류·송신 0 |
| required=true/false 혼재, 정확한 A | 정상 |
| required=false만 있고 임의 envelope | 송신 0 |

acceptance:174의 옛 포괄 문구는 “병렬 후보의 혼합·변조·미등록 opaque”로 수정되었다. design:416과 acceptance:537은 prefix 수가 아니라 중복 제거된 **모든 개별 후보**에 4096개·8MiB를 적용한다. design:471–472의 PASS는 receipt 검사를 통과하여 송신할 수 있다는 뜻이며 provider의 응답 성공을 보장하지 않는다.

### RPA-A1 및 제한된 회귀 확인

design:489–494는 최초 user 입력/알려진 prefix의 새 user branch와 입력에 실제 존재하는 assistant lookup miss를 구분한다. 후자는 송신 0이며 새 branch로 등록하지 않는다. foreign 공개 응답에도 등록된 receipt를 요구하고, 필요 저장 자료의 손실·손상은 빈 초기화로 대체하지 않는다. acceptance:535–536에도 같은 조건이 연결되었다.

원래 경계인 최소 정규화(:412), private receipt의 자료 제한(:413), 필수 no-tool/marker 동시 삭제(:414), publish-before-success와 실패/취소 금지(:415), manifest 상실/rollback 한계(:417), compaction 불명확성의 오류(:418), 귀속 검사 후 foreign strip(:419), 유한 저장(:420)은 본문에 유지되어 있다. 첫 감사의 raw prefix 재계산을 다시 실행했다고 주장하지 않는다.

## Category Scores — 변경분만

| Dimension | Score | Rubric band | Evidence |
|---|---:|---|---|
| Clarity | 1.00 | 변경 항목의 해석이 명시됨 | design:431–454 lifecycle, :474–494 판정 표 |
| Completeness | 1.00 | 이전 차단 항목의 경계가 채워짐 | design:424–466; native 자료/receipt/namespace/lease/폐기 소관 |
| Testability | 1.00 | 해당 AC가 이진 결과로 대응 | acceptance:519–537; PICKER acceptance:62–70 |
| Traceability | 1.00 | 기존 번호에 직접 연결 | core acceptance의 기존 AC-MG-009 보강(:512); PICKER AC-GP-003·005(:60); core plan:531–535 |

이 점수는 두 SPEC 전체에 대한 새 점수가 아니다. 계획에서 preflight와 실제 native 수용 시험을 필수로 남긴 것은 계획 완결성과 runtime 성공을 구별한 것이며, 미실행을 runtime PASS로 채점한 것이 아니다.

## Must-Pass Results — 기존 전체 판정 보존

이번 위임은 앞선 RPA-1/RPA-2와 관련 회귀의 bounded delta다. 전체 MP-1~8를 재실행한 감사가 아니다. 첫 보고서에서 UNVERIFIED로 남긴 전체 구조/전수 추적/RED-now 판정은 그대로이며, 전체 MP-7의 미해결 출력 상한 게이트도 그대로다. Windows 계약은 범위 밖이다. **이 파일의 PASS를 core 전체 must-pass firewall 통과로 해석하면 안 된다.** 변경분에는 새 REQ/AC를 발급하거나 기존 검증 게이트를 제거하는 수정이 없으며, 요구층과 검증층을 혼동하여 점수를 주지 않았다.

## Defects Found

이전 blocking RPA-1·RPA-2는 해소되었다. 선택 RPA-A1도 반영되었다. 위 수정으로 바뀐 경계에서 새로운 blocking 또는 optional 결함을 확인하지 못했다. 이는 미검증 runtime의 무결함을 뜻하지 않는다.

## Evidence

이번 재감사에서 직접 실행한 명령과 원문 출력:

```text
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified branch --show-current
WT-unified-gateway
```

`receipt-plan-repair-iter1.md`를 읽어 변경 위치만 찾고, 작성자의 해소 주장·린트 결과를 판정 근거로 대체하지 않았다. `nl -ba`/`sed -n`으로 core design:399–528, acceptance:512–540, research:1046–1060, plan:527–539, 양쪽 spec HISTORY, PICKER plan:96–112 및 acceptance:60–72를 직접 읽었다. 첫 조회의 출력 일부가 잘려 design:450–503과 양쪽 AC를 별도 호출로 다시 읽었다.

다음 Python 코드를 `python3 - <<'PY'` 단일 invocation으로 실행했다. 텍스트 판정의 동일성과 감사 baseline을 기록하는 도구이며 실제 receipt 구현 시험이 아니다.

```python
from pathlib import Path
import hashlib
b=Path('/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified/.moai/specs')
d=(b/'SPEC-MOAI-GATEWAY-001/design.md').read_text(); a=(b/'SPEC-MOAI-GATEWAY-001/acceptance.md').read_text()
rows=lambda s,start,end:[x for x in s.split(start,1)[1].split(end,1)[0].splitlines() if x.startswith('|') and not x.startswith('|---')][1:]
dr=rows(d,'**RPA-2','**RPA-A1'); ar=rows(a,'## 0.9.0 receipt','Given launcher')
f=lambda r:['FAIL' if ('FAIL' in x.split('|')[-2] or '송신 0' in x.split('|')[-2]) else 'PASS' for x in r]
print('design_candidate_rows',len(dr),'acceptance_candidate_rows',len(ar))
print('design_outcomes',f(dr));print('acceptance_outcomes',f(ar));assert f(dr)==f(ar)==['PASS','PASS','FAIL','PASS','PASS','FAIL','PASS','FAIL']
print('OLD_RPA2_SENTENCE_ABSENT', '귀속, 병렬 동일 공개 답변의 다른 opaque' not in a)
print('fork_preflight_retained','해당 인수 조합/UUID 보존 미확인 시 오류·미완료' in d)
print('runtime_gap_retained','실제 continue/선택/fork·private-profile OAuth 통합이 미실행이면 해당 행은 미완료다.' in a)
print('lookup_miss_fail_closed','입력에 있는 assistant prefix가 lookup miss | 오류·송신 0' in d)
print('active_family_discard_rejected','실행 중 family는 거절한다.' in d)
for sid,ns in [('SPEC-MOAI-GATEWAY-001',['spec','design','research','plan','acceptance']),('SPEC-MOAI-GATEWAY-PICKER-001',['spec','plan','acceptance'])]:
 for n in ns:
  p=b/sid/(n+'.md');print(str(p.relative_to(b)),hashlib.sha256(p.read_bytes()).hexdigest())
```

실제 출력; exit 0:

```text
design_candidate_rows 8 acceptance_candidate_rows 8
design_outcomes ['PASS', 'PASS', 'FAIL', 'PASS', 'PASS', 'FAIL', 'PASS', 'FAIL']
acceptance_outcomes ['PASS', 'PASS', 'FAIL', 'PASS', 'PASS', 'FAIL', 'PASS', 'FAIL']
OLD_RPA2_SENTENCE_ABSENT True
fork_preflight_retained True
runtime_gap_retained True
lookup_miss_fail_closed True
active_family_discard_rejected True
SPEC-MOAI-GATEWAY-001/spec.md dd015deda8f094107de8a4d9956027db60c2d0f050ad52e1937ca1584cc4265e
SPEC-MOAI-GATEWAY-001/design.md 2f4d2220abb50d5245033490df59a34e1016a95d1a18d317a0a7a5f03c25259d
SPEC-MOAI-GATEWAY-001/research.md 07126bee16f4cb4deeaf5c286567dd635a205550c66cf01bf7028f18aec2e0de
SPEC-MOAI-GATEWAY-001/plan.md 5d70c75c6b4723c692b1f8fb76adb4326e2cc9edebea988542df4b9c34039943
SPEC-MOAI-GATEWAY-001/acceptance.md 6759674119783ce576a5a96a553347849e4a427ae335cbb79b953ec3e9b501e7
SPEC-MOAI-GATEWAY-PICKER-001/spec.md 2b75b1d01d418ed485563ba2a957b3866d55fed254dc39908634ed9df3c6ce00
SPEC-MOAI-GATEWAY-PICKER-001/plan.md 888a758c427b2c6e84349d78c56b746cf600650c5c3970e28700946c0dffb8c7
SPEC-MOAI-GATEWAY-PICKER-001/acceptance.md c13660f2458104026714d7ff339a5360420bef765459e550b1af7a08fc6b4cf2
```

## Baseline-attribution

대상 WT는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, branch/HEAD는 위 출력과 같다. core 0.9.0/PICKER 0.2.0의 작업 중 문서를 감사했으므로 HEAD만으로 문서 baseline을 대신하지 않고 위 8개 hash를 함께 기록한다. 부모 source_session_id는 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`, 이 감사자의 앞선 runtime UUID는 `a28010b3-3a55-45d2-9a33-0fe61c2a5e44`이다.

이 재감사는 새 보고서 한 파일만 작성한다. SPEC·제품 코드·기존 감사 파일·원격 상태를 수정하지 않았고, 완료된 작성자를 다시 호출하지 않았다.

## Gaps

receipt 구현은 이 감사의 입력이 아니며 실행하지 않았다. 새 native 실행, fork 인수 수용, continue/선택형 resume, retained config+실계정 OAuth+TUI 통합, 실제 opaque 삭제·혼합·cross-provider·디스크/락 실패, Windows runtime은 미실행이다. 문서에 요구된 namespace/UUID compatibility preflight와 실제 음성·양성 수용은 이후 작업으로 남는다. 출력 상한/Windows 승인 대기는 이번 판정으로 변하지 않는다.

## Residual-risk

native 내부 namespace 변수와 fork 인수 조합은 버전별 변화 가능성이 문서에 남아 있다. 실제 preflight 실패 시 그 경로는 여전히 미완료이며 대체 인계를 설계·검증해야 한다. 오류 반환만으로 정상 fork/재개 AC를 완료할 수 없다. hash receipt는 제한된 일관성 검사이며 공개 history와 저장소 전체를 함께 되돌리는 공격의 인증 수단이 아니다. 제품 활성화는 기존 실제 carrier·resume·foreign strip·유실/동시성·네 GPT 후속 게이트를 모두 통과해야 한다.

## Iteration History / Recommendation

- iter1: FAIL 0.6875 — blocking RPA-1/RPA-2, optional RPA-A1.
- iter2: PASS 1.00 — 위 세 항목의 문서 수정과 관련 회귀를 확인. 전체 코어 판정이나 제품 활성화로 확대하지 않는다.

승인된 범위에서 receipt/private lifecycle 구현과 문서에 남긴 compatibility preflight를 진행할 수 있다. 실제 native fork를 측정하기 전부터 지원 완료로 표시하지 말고, 구현 뒤 기존 양성·음성 AC로 수용을 판정한다.

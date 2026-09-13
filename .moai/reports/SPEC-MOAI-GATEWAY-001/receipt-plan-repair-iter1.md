# Receipt RPA-1·RPA-2 보정 및 변경분 재감사 인계

## Claim

부모의 명시 재위임에 따라 receipt 첫 감사의 RPA-1·RPA-2와 선택 RPA-A1을 문서에서 보정했다. 코어는 0.9.0,
PICKER는 0.2.0을 유지하며 각 HISTORY에 같은 버전의 보정을 기록했다. 구현·감사 PASS 또는 제품 활성화를 선언하지 않는다.

수정 위치와 재감사 항목:

| 항목 | 본문 위치 | 구체 보정 |
|---|---|---|
| RPA-1 | 코어 design.md:422, PICKER plan.md:96, acceptance.md:60 | new/exact resume/continue/선택형 resume/fork의 사전 UUID 결정·인증 bootstrap·retained root 표 |
| RPA-1 수명 | 같은 design 절 | family별 native config/transcript와 UUID별 receipt 보존, 정상 종료의 임시 overlay 정리와 명시 폐기 구분 |
| RPA-1 설정/인증 | 같은 design 절, research.md:1046 | 원래 secure-storage namespace 보존, 원본 설정/credential 복사 금지, 2.1.268 한정 근거와 compatibility preflight |
| RPA-2 | 코어 design.md:468, acceptance.md의 0.9.0 receipt 감사 보정 | 정확한 A/B 각각 정상, 변조/혼합/삭제 실패, 동일 중복 idempotent, required=false 단독 빈 입력 정상, true/false 혼재 빈 입력 실패 |
| RPA-A1 | 코어 design.md:485, 같은 AC 보정 | 최초 user·알려진 prefix branch 허용과 입력 assistant lookup miss 거절을 구별, foreign 공개 응답의 등록 필요 |

동일 family의 private native 설정은 실행 lease로 보호한다. 중복 lead는 명시 busy이며 다른 family의 동시 실행을 막지 않는다.
이것은 한 lead의 병렬 gateway 요청에 대한 직렬화가 아니다. 같은 공개 prefix의 A/B opaque 후보는 각각 수용한다.
fork는 같은 retained family 안에서 native가 새 transcript를 만들고 MoAI는 parent의 hash-only receipt snapshot만
child에 귀속한다. 원문 대화를 MoAI receipt에 복사하지 않는다. snapshot 후보도 4096개·8MiB 한도에 포함한다.

continue와 선택형 resume는 launcher가 소유 인덱스에서 먼저 UUID를 정하고 exact resume로 인계한다. request metadata는
그 UUID의 일치 검사만 한다. fork의 native 인수 조합과 정해진 child UUID의 사용은 아직 실제 관측 전이므로 preflight와
필수 runtime AC를 유지한다. 정상 종료 때 sole transcript를 삭제하는 구현은 기존 재개 양성 AC를 통과할 수 없도록 명시했다.

## Evidence

먼저 [receipt-plan-audit-iter1.md](receipt-plan-audit-iter1.md)와
[picker-settings-source-observation.md](picker-settings-source-observation.md)를 읽었다. 전자는 FAIL 0.6875와 두 blocking을
명시하고, 후자는 private config/원래 secure namespace의 Opus 5 한 요청 200 및 원본 여섯 설정 불변의 제한된 근거다.
해당 runtime을 다시 실행하지 않았으며 native exact resume와 한 프로세스의 OAuth 공존을 합쳐 제품 통합 PASS로 세지 않는다.

기준 확인의 실제 출력:

```text
git rev-parse --short HEAD
81c1d58f9
git branch --show-current
WT-unified-gateway
```

이번 수정 전 각 문서의 SHA-256을 저장하고 수정 후 UTF-8 readback·hash 비교·최초 REQ/AC 목록 비교를 Python으로 실행했다.

```text
changed_docs 8
.moai/specs/SPEC-MOAI-GATEWAY-001/research.md
.moai/specs/SPEC-MOAI-GATEWAY-001/plan.md
.moai/specs/SPEC-MOAI-GATEWAY-001/spec.md
.moai/specs/SPEC-MOAI-GATEWAY-001/design.md
.moai/specs/SPEC-MOAI-GATEWAY-001/acceptance.md
.moai/specs/SPEC-MOAI-GATEWAY-PICKER-001/plan.md
.moai/specs/SPEC-MOAI-GATEWAY-PICKER-001/spec.md
.moai/specs/SPEC-MOAI-GATEWAY-PICKER-001/acceptance.md
progress_unchanged True
SPEC-MOAI-GATEWAY-001 IDS_UNCHANGED "0.9.0" draft
SPEC-MOAI-GATEWAY-PICKER-001 IDS_UNCHANGED "0.2.0" draft
RPA2_OLD_CONTRADICTION_ABSENT True
```

마지막 줄은 지적된 옛 문장의 부재를 확인한 텍스트 검사이며 의미적 감사 PASS가 아니다.

```sh
/tmp/moai-gateway-working-20260911 spec lint SPEC-MOAI-GATEWAY-001
/tmp/moai-gateway-working-20260911 spec lint SPEC-MOAI-GATEWAY-PICKER-001
```

각 명령의 실제 출력, exit 0:

```text
✓ No findings — all SPEC documents are valid
✓ No findings — all SPEC documents are valid
```

`git diff --check -- .moai/specs/SPEC-MOAI-GATEWAY-001 .moai/specs/SPEC-MOAI-GATEWAY-PICKER-001`은 빈 출력, exit 0이었다.
untracked 문서의 내용을 이 명령만으로 검증했다고 주장하지 않으며 위 별도 readback을 수행했다.

## Baseline-attribution

2026-09-11, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, `WT-unified-gateway`, HEAD `81c1d58f9`.
부모 source_session_id는 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이다. manager-spec은 위 8문서와 이 보고서만 작성했다.
기존 AUTH 문서, 코드, progress, 감사 원문, 이전 관측 파일은 수정하지 않았다. commit·push·PR도 수행하지 않았다.

## Gaps

- 구현·독립 재감사·새 native 실행은 수행하지 않았다. 실제 continue/선택형 resume/fork와 retained profile/OAuth/TUI 통합은 미완료다.
- 원래 secure namespace의 모든 환경·버전·플랫폼 호환은 미검증이며 정확한 2.1.268 관측 밖은 preflight가 필요하다.
- 출력 상한 및 Windows 내구성의 사용자 답변 대기는 그대로다. `rg`로 core plan:224와 AUTH plan:112·115의 표식을 확인했다.
- 전체 코어/SSE/OAuth 구현 감사나 전체 SPEC MP-7 판정을 이 보정으로 통과시킬 수 없다.

## Residual-risk

family 실행 lease가 native 프로세스 수명과 제대로 연결되지 않으면 설정 경합 또는 거짓 busy가 생길 수 있다. fork 인수의
정확한 native UUID 결과는 실제 시험이 필요하다. 후보 집합의 정상 A/B와 required=false 혼재를 혼동하지 않도록 문서
수용 표와 AC를 함께 구현해야 한다. 이 위험을 회피하려고 정상 병렬 후보·fork·재개 AC를 삭제해서는 안 된다.

재감사 범위는 **RPA-1의 사전 귀속·retained native 수명 및 관련 lease/namespace/폐기 경계, RPA-2의 후보 수용 표,
RPA-A1 lookup miss/branch 분리와 이 보정이 바꾼 경계**로 한정한다. 출력 상한·Windows 의미 결정과 기존 구현 전체는
이 bounded plan 재감사에 포함하지 않는다. receipt 제품 활성화 게이트는 독립 감사·구현·실제 양성/음성 수용까지 닫혀 있다.

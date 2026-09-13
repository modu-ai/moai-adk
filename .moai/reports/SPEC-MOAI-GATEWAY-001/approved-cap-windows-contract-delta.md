# 승인된 구독 출력 정책·Windows API 저장 기준의 좁은 계약 보강

## Claim

부모가 전달한 두 명시 사용자 답변을 반영했다.

- “구독 서버 출력 정책으로 진행”
- “Windows API 기준으로 진행”

코어 0.9.0의 출력 상한 조항과 AUTH 0.2.0의 플랫폼별 저장 성공 조항만 보강했다. 새 native 정책 후보는
[native-policy-contract-proposal.md](native-policy-contract-proposal.md)에만 있으며 승인된 SPEC으로 옮기지 않았다.
receipt 본문의 전후 SHA-256은 같고 순서도 그대로다. 실제 구현·Windows CI·provider 통합 PASS는 선언하지 않는다.

### 출력 정책 인계

- AuthPKCE의 고정 구독 endpoint에서만 outgoing max_output_tokens를 생략한다.
- Claude 입력 max_tokens의 양의 정수·중복·허용 범위 검증은 유지한다. API 키의 max_output_tokens 매핑도 유지한다.
- 구독 서버 출력 정책 사용과 요청별 생성 token 상한 미강제를 명시한다. byte cap·취소·deadline을 생성 token과
  같은 제한이라고 설명하지 않는다. 응답 종료 표·재시도 금지·앞부분 무삭제 AC도 보존한다.
- API 키/다른 endpoint까지 생략하는 뮤턴트와 잘못된 입력의 송신 0 대조군을 기존 AC에 추가했다.

### Windows 저장 인계

- POSIX의 file fsync·rename·directory 내구성 요구는 그대로다.
- Windows AUTH에만 file flush → 같은 디렉터리/volume의 MoveFileEx(REPLACE_EXISTING|WRITE_THROUGH) →
  최종 owner/ACL·파일/부모 identity·내용 readback을 성공 조건으로 둔다. COPY_ALLOWED·cross-volume fallback 금지다.
- 기존 atomicfile.Replace 강제는 POSIX에 유지하고 Windows AUTH만 native 예외를 허용한다. 기존 primitive의 다른
  호출자나 전역 의미는 개편하지 않는다.
- protected current-user 루트와 broker 자식의 검증된 inherited ACL을 구별한다. NULL/broad ACL·잘못된 owner·reparse·
  부모 교체를 거절하며 사후 ACL 수정으로 공개 가능성이 있던 파일을 안전한 것으로 채택하지 않는다.
- 교체 전 실패는 이전 canonical 보존, 교체 후 readback 실패는 상태 변경 가능성 구분과 미확인 credential 송신 금지다.
- 프로세스 crash/잠금 회수는 실제 GitHub runner의 해당 HEAD·필수 run/pass·artifact로 판정한다. 전원 장애 뒤 POSIX
  directory fsync와 동등한 내구성으로 확대하지 않는다. 기존 필수 시험·세대·logout barrier 요구를 삭제하지 않는다.

## Evidence

이 위임에서 먼저 해당 core/AUTH의 현재 문장을 sed로 읽고 승인된 문구만 교체했다. 수정 전 문서별 SHA와 core design의
`선택된 설계 후보`부터 `## 5.` 직전까지 receipt 영역 SHA를 저장한 뒤 최종 비교했다. Python의 실제 출력:

```text
RECEIPT_REGION_SHA_UNCHANGED True
CHANGED_DOCS 7
.moai/specs/SPEC-MOAI-GATEWAY-001/plan.md
.moai/specs/SPEC-MOAI-GATEWAY-001/spec.md
.moai/specs/SPEC-MOAI-GATEWAY-001/design.md
.moai/specs/SPEC-MOAI-GATEWAY-001/acceptance.md
.moai/specs/SPEC-MOAI-GPT-AUTH-001/plan.md
.moai/specs/SPEC-MOAI-GPT-AUTH-001/spec.md
.moai/specs/SPEC-MOAI-GPT-AUTH-001/acceptance.md
SPEC-MOAI-GATEWAY-001 IDS_UNCHANGED "0.9.0" draft
SPEC-MOAI-GPT-AUTH-001 IDS_UNCHANGED "0.2.0" draft
PROGRESS_UNCHANGED True
```

REQ/AC의 식별자 목록과 순서를 최초 수정 전 목록과 기계적으로 비교했다. 활성 코어 24/24와 묘비 2/1, AUTH 10/10의
번호를 재사용하거나 제거하지 않았다. HISTORY는 같은 버전의 후속 결정으로 보강했다.

최종 실행한 명령:

```sh
/tmp/moai-gateway-working-20260911 spec lint SPEC-MOAI-GATEWAY-001
/tmp/moai-gateway-working-20260911 spec lint SPEC-MOAI-GPT-AUTH-001
```

각 명령의 원문 출력, exit 0:

```text
✓ No findings — all SPEC documents are valid
✓ No findings — all SPEC documents are valid
```

`rg -n`으로 직접 확인한 주요 위치:

| 변경 | 위치 |
|---|---|
| 구독 출력 HISTORY·REQ 보강 | core spec.md:29 및 REQ-MG-017 |
| 구독 매핑 | core design.md:377 |
| 결정 대기 표식 해소 | core plan.md:224 |
| 경로별 출력 AC | core acceptance.md:539 |
| Windows 성공 의미 | AUTH spec.md REQ-GA-004, plan.md:24 |
| primitive 예외 | AUTH plan.md:44 |
| 두 승인 기록·ACL | AUTH plan.md:119·123 이후 |
| Windows native AC | AUTH acceptance.md:67 |

구독 출력과 Windows 의미의 NEEDS CLARIFICATION 표식은 해당 core/AUTH plan에서 제거했다. 이전 감사/관측 보고서의
대기·FAIL 기록은 역사적 원문으로 보존했으며 새 완료로 덮어쓰지 않았다.

## Baseline-attribution

2026-09-11, `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, `WT-unified-gateway`, HEAD `81c1d58f9`.
부모 source_session_id는 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이다. 다른 작성자는 코드·probe를 소유한다.
이 작성자는 위 7문서, native 정책 제안 보고서와 본 delta만 작성했다. receipt 구현·코드·progress·commit·원격 변경 없음.

## Gaps

- 승인된 출력 분기의 코드·wire 음성/양성 시험과 사용자 표시 구현은 별도다. 아직 그 구현 PASS를 측정하지 않았다.
- Windows native Store 활성화·권한/잠금·crash·readback 실제 GitHub CI 증거는 아직 필요하다.
- 새 native adaptive/title/effort/keep-all/metadata 정책 후보는 미승인 SPEC 상태이며 이 두 승인에 포함하지 않는다.
- 구조 lint는 독립적인 의미/구현 감사를 대체하지 않는다. receipt 코드의 현재 상태도 이 작업에서 감사하지 않았다.

## Residual-risk 및 두 변경분 감사 범위

**출력 정책 delta:** 기존 translator/adapter 코드가 있는 범위이므로 sync-auditor가 고정 구독 AuthPKCE만 생략하는지,
양의 정수 입력 검증/API 키 매핑/terminal/byte·cancel 경계와 사용자 표시가 계약에 맞는지 검증한다. 출력 cap 생략을
일반 native 정책 허용이나 upstream truncation 보장으로 확대하지 않는다.

**Windows delta:** 새 Windows Store 통합 전에는 plan-auditor가 플랫폼별 성공 의미·native primitive 예외·post-replace
오류·protected/inherited ACL·CI 증거 기준의 일관성을 확인한다. 코드가 연결된 뒤에는 sync-auditor와 실제 GitHub native
시험으로 별도 수용한다. cross-compile이나 Win32 API 문서만으로 실행 PASS를 주지 않는다.

두 변경분 모두 receipt 계획 재감사를 다시 여는 근거가 아니며 receipt 활성화 게이트는 기존대로다. 그 외 native 정책
제안은 별도 승인·계약 보강·시험으로 처리한다.

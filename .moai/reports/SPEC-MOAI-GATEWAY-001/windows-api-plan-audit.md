# Windows AUTH API 저장 계약 변경분 감사

Iteration: 1 (승인된 Windows API delta)
Verdict: PASS
Overall Score: 1.00
Scope: SPEC-MOAI-GPT-AUTH-001 0.2.0의 Windows 저장 계약 변경분

Reasoning context ignored per M1 Context Isolation.

## Claim

사용자가 승인한 Windows API 기준을 반영한 AUTH spec/plan/acceptance의 해당 조항은 서로 일치한다. **새 Windows Store 통합을 진행할 수 있는 계획 계약**이며 실제 Windows 실행·Store 구현·전체 AUTH SPEC 또는 릴리스 PASS가 아니다. POSIX의 기존 내구성 요구를 변경하거나 Windows 전원 장애 보장을 덧붙이지 않는다.

감사 전의 회귀 가설은 POSIX까지 내구성 완화, Windows 예외의 전역 primitive 확장, cross-volume copy fallback, protected 루트와 inherited broker 파일의 혼동, 교체 후 실패를 무조건 rollback으로 보고, 미확인 credential의 송신, 합성 checker/누락 test를 native PASS로 취급하는 경우였다. 아래 각 경계에서 확인했다.

## 판정 근거

| 변경 경계 | 판정 | 직접 읽은 근거 |
|---|---|---|
| POSIX 유지 | PASS | AUTH plan.md:23–26, :43. POSIX write/fsync/rename/directory 내구성과 atomicfile.Replace 유지 |
| Windows 성공 조건 | PASS | spec.md:45–48, plan.md:24–27, acceptance.md:69–73. flush→교체→최종 owner/ACL·파일/부모 identity·내용 readback |
| Windows AUTH만 primitive 예외 | PASS | plan.md:43–46. 같은 디렉터리/volume MoveFileEx(REPLACE_EXISTING\|WRITE_THROUGH), COPY_ALLOWED/cross-volume 금지, 다른 atomicfile 호출자 변경 없음 |
| 실패 단계/송신 | PASS | plan.md:47–48, acceptance.md:72–73. 교체 전 기존 canonical 보존, 교체 뒤 상태 변경 가능성 표시, 미확인 credential 송신 0 |
| protected/inherited ACL | PASS | plan.md:127–131, acceptance.md:69–72. root/scratch current-user protected DACL과 broker 자식의 실제 owner/effective DACL/부모·handle identity를 구별 |
| 권한 실패의 사후 은폐 금지 | PASS | plan.md:129–130. NULL/부재·broad·잘못된 owner·reparse/경로 교체 거절, 공개 가능 파일의 사후 ACL 수정 채택 금지 |
| 세대/로그아웃 기존 계약 | PASS, 유지 확인 | spec.md:50–56, plan.md:28–41 및 :59–68, acceptance.md:75–77. 늦은 refresh와 tombstone 덮어쓰기 금지 및 프로세스 간 잠금 유지 |
| 실제 CI 조건 | PASS, 계약 한정 | plan.md:125, acceptance.md:64–65, :75–78. 해당 HEAD/run ID/필수 run-pass/artifact, missing/skip 불허, 프로세스 crash와 전원 장애 구별 |

AUTH plan:34의 일반적인 “취소·실패는 기존 상태를 보존한다”는 문장은 같은 plan:47–48과 acceptance:73의 **교체 전/후 명시 구분**에 따라 읽었다. 후자의 구체 계약을 무시하여 교체 후 실패에 자동 rollback을 구현할 수 있다는 허용으로 판정하지 않는다. 별도의 blocking 상충은 확인하지 못했다.

[Microsoft MoveFileExW 문서](https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw)의 flags와 반환 규칙을 이번 실행에서 읽었다. COPY_ALLOWED는 다른 volume에서 copy/delete 방식으로 움직이는 옵션이며, REPLACE_EXISTING과 WRITE_THROUGH가 별도로 정의되어 있다. 이 문서나 flag 이름만으로 POSIX directory fsync와 동등한 전원 장애 복구를 증명하지 않았다. 승인된 계약은 native API 실행·readback·실제 프로세스 crash 시험에 한정되어 있다.

## Defects Found

이 Windows 계약 변경분에서 blocking 또는 optional 결함을 확인하지 못했다. 아래 CI/구현 미실행은 아직 계획대로 수행할 작업이며, 이미 성공했다는 주장이 없으므로 설계 결함으로 바꾸지 않았다.

## Category Scores — 변경분만

| Dimension | Score | Evidence |
|---|---:|---|
| Clarity | 1.00 | plan:23–27, :43–48의 플랫폼/실패 단계 구별 |
| Completeness | 1.00 | plan:127–131 및 acceptance:69–78의 저장·권한·잠금·crash·증거 경계 |
| Testability | 1.00 | acceptance:69–78의 정상·실패·교체 이후 송신 0·missing/skip 대조군 |
| Traceability | 1.00 | spec REQ-GA-004·005(:44–50), acceptance의 기존 AC-GA-004·005 보강(:67) |

전체 AUTH MP-1~8 또는 다른 변경분을 다시 채점하지 않았다. 이 파일은 Windows 계약 변경분의 독립 계획 판정이며 전체 SPEC 감사 게이트를 덮어쓰지 않는다. 출력 상한 구현과 receipt 계획/코드는 범위 밖이다.

## Evidence

이 실행에서 직접 수행한 기준 명령과 원문 출력:

```text
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified rev-parse HEAD
81c1d58f9cf7045594ee61d5e4ff380948ce9eba
$ git -C /Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified branch --show-current
WT-unified-gateway
```

`nl -ba`로 AUTH spec.md 72행, plan.md 131행, acceptance.md 82행의 전체 문서를 직접 읽었다. approved-cap-windows-contract-delta.md는 변경 위치와 명시 사용자 결정의 전달 기록으로 읽었으며, 작성자의 lint/PASS 주장을 이번 실행으로 전용하지 않았다.

`cat scripts/ci-census/windows-gateway-required.json`의 현재 원문:

```json
[
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway/auth","Test":"TestWindowsPrivateStorageCandidate"},
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway/auth","Test":"TestWindowsPrivateStorageRejectsBroadACL"},
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway/auth","Test":"TestWindowsPrivateStorageRejectsWrongTypeAndRelative"},
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway/auth","Test":"TestWindowsLockCancelledBeforeAcquisition"},
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway/auth","Test":"TestWindowsLockContentionCancellationAndCrashRelease"},
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway","Test":"TestSupervisorChildStartsAndStops"},
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway","Test":"TestSupervisorRunnerLifetimeClosesPortAndOwnedOverlay"},
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway","Test":"TestSupervisorParentProcessDeath"},
  {"Package":"github.com/modu-ai/moai-adk/internal/gateway","Test":"TestSupervisorCancelledStopStillJoinsReaper"}
]
```

exit 0. 현재 9개 명시 test는 기존 private storage/lock/supervisor 경계다. windows-ci-preparation.md의 초기 7개라는 기록을 현재 manifest 수로 사용하지 않았다. **9개가 향후 Windows Store 통합의 모든 AC를 이미 검사한다는 주장은 하지 않는다.** 새 flush/replace/readback·safe inherited broker·postreplace 송신 0 등의 실제 시험은 acceptance:69–78에 따라 구현 후 CI에 연결해야 한다. 기존 test를 삭제하거나 skip하여 완료할 수 없다.

Python hashlib로 이번 문서 bytes를 측정한 원문 출력:

```text
spec a90b41166fbe6c5767c2fad9e73e63f9a62af38aef43921ef9620d7a9d8ef8c8
plan 0adffe4519a0fb5f72ace0bc463a1998b23eead2e452c55aa633a63903b4a0b0
acceptance 9834f0193bec480b11940120202fe97754159c568a5106f01b94d4f367c58a0a
```

측정 코드의 대상은 아래 SPEC 디렉터리이며 각각 `hashlib.sha256(p.read_bytes()).hexdigest()`를 실행했다. exit 0. 웹 읽기는 위 Microsoft URL의 flags/return 항목을 open으로 확인했으며 provider 요청·credential 접근·GitHub 실행은 없었다.

## Baseline-attribution

WT `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`, AUTH 경로 `.moai/specs/SPEC-MOAI-GPT-AUTH-001`, version 0.2.0, draft. HEAD만으로 미커밋 문서의 bytes를 대신하지 않고 위 hash를 함께 귀속한다. 부모 source_session_id는 `01a08e7b-6aa0-7361-ab7e-ea8da1f02228`이다. 이 감사는 새 보고서만 작성하며 SPEC·제품 코드·원격 상태를 변경하지 않는다.

## Gaps

Windows 실행·CI run ID/artifact·실제 ACL/LockFileEx/MoveFileEx/crash/readback 시험은 수행하지 않았다. 기존 구현 또는 CI checker의 기능 감사도 수행하지 않았다. 새 Store 통합 뒤 acceptance:69–78의 실제 test run/pass 증거가 필요하다. API 문서의 설명은 실행 증거가 아니다.

## Residual-risk

저장 API가 성공해도 이후 readback이 실패할 수 있으므로 단계별 상태 변경 가능성과 미확인 credential 송신 차단을 구현에서 유지해야 한다. crash 시험은 전원 장애 시험이 아니다. 사용자 승인은 Windows API 기준의 저장 계약을 선택한 것이며 실제 native 수용 완료를 대신하지 않는다.

## Recommendation

승인된 Windows AUTH 예외를 구현하고 기존 POSIX 및 다른 primitive 호출자는 유지한다. 통합 후 실제 GitHub Windows runner에서 기존 9개 gate와 새 Store AC를 함께 판정한다. 이 계획 PASS로 skipped/missing 결과 또는 오래된 HEAD의 실행을 수용하지 않는다.

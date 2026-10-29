# SPEC-AGENTIC-FACTORY-001 — 수용 기준

## §A 측정 규칙

- 수용 기준은 25개이며 Tier L 상한(25)과 같다. 모든 기준은 REQ 하나 이상에 연결된다.
- **시나리오 AC-AF-001 ~ AC-AF-007은 하나의 실제 Issue에서 처음부터 끝까지 측정한다.** 어떤 Issue를 쓸지는 OQ-1이며, 실제 Issue에 쓰는 동작은 운영자 승인 뒤에만 한다. 시나리오 시험은 빌드 태그 `af_e2e`와 환경 변수 `MOAI_AF_E2E_ISSUE=<번호>`가 있어야 실행된다. 환경 변수가 없으면 시험은 skip이 아니라 FAIL로 끝나야 한다(빈 실행이 PASS로 읽히지 않도록).
- 각 명령은 `-run '^<이름>$' -v`로 시험 하나를 고른다. 판정은 출력에 `--- PASS: <이름>`이 있고 `[no tests to run]` / `[no test files]`가 없을 때만 PASS다. 종료 코드 0만으로는 PASS가 아니다.
- 시험 이름과 경로는 계약 수준의 약속이며, 구현 단계에서 경로를 바꾸면 이 문서를 manager-spec이 같이 고친다.

## §B 증거 원장 (RED-now 기준점)

문서 수준 고정 SHA: `5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4`. 아래 기준점은 자기 SHA가 없는 모든 기준에 적용된다.

```
id: EV-RED-0
command: go test ./internal/agenticfactory/... -count=1
stdout:
pattern ./internal/agenticfactory/...: lstat ./internal/agenticfactory/: no such file or directory
FAIL	./internal/agenticfactory/... [setup failed]
exit_code: 1
tree_sha: 5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4
red_reason: 패키지가 없다 — 이 SPEC이 만들 대상 자체가 아직 없다.
```

이 기준점은 "대상이 없다"는 사실만 증명한다. 분류상 `TOOL_FAILURE`(셋업 실패)이지 `EXPECTED_RED`가 아니므로, run 단계는 각 기준의 이름 붙은 시험이 **의도한 단언에서** 실패하는 `EXPECTED_RED`를 먼저 기록해야 한다(`tdd-result-contract.md`). 그 기록 전에는 어느 기준도 채택된 것이 아니다. 이 기준점을 지금 다시 실행할 수 있으므로 release-blocking 자격은 유지한다.

## §C 수용 기준

### 시나리오 (하나의 실제 Issue, 처음부터 끝까지)

#### AC-AF-001 — 경쟁 claim (maps REQ-AF-006)
- Given 실제 Issue 하나가 `queued`이고 claim이 없다
- When 두 claimant가 동시에 claim을 요청한다
- Then 정확히 하나만 성공하고, 다른 하나는 종료 코드 10과 함께 쓰기가 거절되며, claim 기록은 하나다
- 검증: `MOAI_AF_E2E_ISSUE=<n> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenario01_CompetingClaim_ExactlyOneWins$' -count=1 -v`
- 초록 경로: M2(claim) 구현 뒤 PASS

#### AC-AF-002 — 중복 dispatch 거절 (maps REQ-AF-007)
- Given 같은 Issue에 살아 있는 attempt가 있다
- When 같은 Issue로 dispatch를 다시 보낸다
- Then 새 dispatch는 거절되고, 충돌한 기존 dispatch id가 기록된다. attempt 수는 늘지 않는다
- 검증: `MOAI_AF_E2E_ISSUE=<n> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenario02_DuplicateDispatch_Refused$' -count=1 -v`
- 초록 경로: M2

#### AC-AF-003 — ACK 유실 (maps REQ-AF-003)
- Given 명령이 `sent`이고 ACK가 오지 않았다
- When 재전송 판단 시점이 된다
- Then 받는 쪽 상태를 먼저 조회하고, `received` 이상이면 재전송하지 않는다. 전송 성공이나 idle 알림만으로는 `completed`가 되지 않는다
- 검증: `MOAI_AF_E2E_ISSUE=<n> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenario03_AckLoss_RequeryNotResend$' -count=1 -v`
- 초록 경로: M3(delivery)

#### AC-AF-004 — 중단 복구 (maps REQ-AF-009, REQ-AF-010)
- Given attempt가 진행 중이고 세션이 끊겼거나 lease가 만료됐다
- When 복구가 시작된다
- Then 실제 실행(PID 생존, 트리 HEAD·dirty, delivery 상태)을 먼저 조회한다. 강제 종료는 0건, 동시에 살아 있는 attempt는 언제나 1개 이하다
- 검증: `MOAI_AF_E2E_ISSUE=<n> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenario04_InterruptionRecovery_QueryFirstNoDuplicate$' -count=1 -v`
- 초록 경로: M2

#### AC-AF-005 — 감사 예산 소진 (maps REQ-AF-018)
- Given 감사 라운드가 `audit.max_rounds`에 닿았고 PASS가 없다
- When 병합 게이트가 판정한다
- Then Issue는 `blocked`이고 사유는 예산 소진이며, 병합 게이트 판정은 PASS가 아니다
- 검증: `MOAI_AF_E2E_ISSUE=<n> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenario05_AuditBudgetExhausted_Blocked$' -count=1 -v`
- 초록 경로: M4(gate)

#### AC-AF-006 — 병합 결과 미상 (maps REQ-AF-021)
- Given 병합 요청 뒤 응답이 끊겨 결과를 모른다
- When 병합 단계가 다시 판단한다
- Then 원격(PR 상태, `main` 포함 여부)을 먼저 조회하고, 병합 요청은 다시 보내지 않는다(병합 요청 횟수 1)
- 검증: `MOAI_AF_E2E_ISSUE=<n> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenario06_MergeOutcomeUnknown_QueryRemoteNoRemerge$' -count=1 -v`
- 초록 경로: M4

#### AC-AF-007 — 원격 반영 확인 → 종료 → 다음 dispatch (maps REQ-AF-021, REQ-AF-001)
- Given 병합이 원격 `main`에 반영됐다
- When ControlService가 반영을 확인한다
- Then 순서가 정확히 `done` 기록 → Issue 종료 → 다음 dispatch이며, 반영 확인 전에는 종료도 다음 dispatch도 없다
- 검증: `MOAI_AF_E2E_ISSUE=<n> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenario07_RemoteConfirmed_CloseThenNextDispatch$' -count=1 -v`
- 초록 경로: M4

### 상태·원장

#### AC-AF-008 — 전이표 강제 (maps REQ-AF-001, REQ-AF-002)
- Given §2.2 전이표
- When 표에 없는 간선(예: `implementing` → `done`), 권한 없는 역할의 전이, `blocked`에서 `return_stage`가 아닌 단계로의 복귀를 요청한다
- Then 셋 다 거절되고 상태는 그대로다. 표에 있는 간선은 모두 허용된다(양방향 확인). `blocked` 진입은 `design.md` §2.1a 집합의 값 하나와 근거·재개 조건이 있을 때만 허용되고, 집합 밖 값이나 근거 없는 진입은 거절되며, `blocked` 상태를 성공으로 보고하는 경로가 없다
- 검증: `go test ./internal/agenticfactory/state/ -run '^TestState_TransitionTable_RefusesUnlistedAndAcceptsListed$' -count=1 -v`

#### AC-AF-009 — 소유권 확인 실패 시 쓰기 거절 (maps REQ-AF-008)
- Given claim 없음 / holder 불일치 / 저장소 읽기 실패의 세 경우
- When writer가 쓰기를 시도한다
- Then 세 경우 모두 종료 코드 11로 거절되고, 확인 없는 쓰기로 넘어가지 않는다
- 검증: `go test ./internal/agenticfactory/claim/ -run '^TestClaim_OwnershipUnverified_WriteRefused$' -count=1 -v`

#### AC-AF-010 — 기한 지난 살아 있는 소유자 (maps REQ-AF-010)
- Given lease가 만료됐지만 소유 PID가 살아 있다(또는 PID를 모른다)
- When 다른 세션이 인계를 요청한다
- Then 인계는 거절되고 attempt는 `lapsed`가 된다. 사망 확정 또는 안전 정지 기록 뒤에만 epoch가 하나 오른 새 attempt가 생기며, 옛 트리·브랜치·SHA·dirty 목록이 새 attempt에 참조로 남는다. 인계 뒤 옛 epoch로 들어온 쓰기는 종료 코드 15로 거절된다(fencing)
- 검증: `go test ./internal/agenticfactory/claim/ -run '^TestClaim_ExpiredLeaseLiveOwner_TakeoverRefused$' -count=1 -v`

#### AC-AF-011 — 저장소와 거울의 불일치 조정 (maps REQ-AF-011)
- Given 효과 상태가 저장소와 Issue 거울에서 다르다 / 저장소가 없다
- When 조정이 돈다
- Then 첫 경우는 거울을 저장소 값으로 다시 쓴다. 둘째 경우는 `blocked`이며 거울에서 claim을 만들지 않는다
- 검증: `go test ./internal/agenticfactory/ledger/ -run '^TestLedger_Reconcile_StoreWinsAndMissingStoreBlocks$' -count=1 -v`

#### AC-AF-012 — 원장 필드 권한과 비밀 가림 (maps REQ-AF-004, REQ-AF-005)
- Given §4의 필드 표
- When 소유자가 아닌 역할이 필드를 쓰거나, 비밀 패턴이 담긴 값이 쓰기 경로에 들어온다
- Then 첫 경우는 거절, 둘째 경우는 가림 처리와 가림 기록이 남고 원문은 어디에도 없다. 관찰 읽기는 제어 권한 없이 성공한다
- 검증: `go test ./internal/agenticfactory/ledger/ -run '^TestLedger_FieldOwnershipAndSecretRedaction$' -count=1 -v`

### 정책

#### AC-AF-013 — Codex 미설정 시 호출 0건 (maps REQ-AF-013)
- Given `workflow.agentic_factory.codex.enabled`가 없다 / `false`다
- When 전체 사이클(claim → done)을 돈다
- Then Codex 호출 계수기가 두 경우 모두 0이다. `true`일 때는 계수기가 1 이상이어서 계수기 자체가 작동함을 보인다
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_CodexUnset_ZeroCalls$' -count=1 -v`

#### AC-AF-014 — GLM 제외와 환경 매트릭스의 정직한 차단 (maps REQ-AF-012, REQ-AF-015)
- Given 미설치·권한 거부·용량 부족·미설정 환경이 섞인 매트릭스
- When 매트릭스를 채운다
- Then 그 환경들은 모두 `BLOCKED(<원인>)`이고 `PASS`가 아니다. GLM 백엔드 호출은 0건이다
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_EnvMatrix_BlockedNotPass_GLMExcluded$' -count=1 -v`

#### AC-AF-015 — 인증 실패·429는 재시도 고리 없음 (maps REQ-AF-014)
- Given 공급자가 401 / 403 / 429를 돌려준다
- When 호출자가 처리한다
- Then 같은 요청 재시도 0회, 상태 보존, `blocked` + 정상 재개 조건 기록, 새 키·용량 요청 0건
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_AuthDenied429_NoRetryLoop$' -count=1 -v`

#### AC-AF-016 — 브라우저 선택 (maps REQ-AF-016)
- Given (a) 선택 없음, (b) `aside` 선택 + 설치됨, (c) `egolite` 선택 + 미설치, (d) 기존 선택 `aside` 상태에서 설정 다시 읽기
- When 브라우저 어댑터를 부른다
- Then (a) `chrome`만 호출, (b) `aside`만 호출, (c) `blocked(not-installed)` + 재개 조건이며 다른 브라우저 호출 0건, (d) 선택이 `aside`로 남는다. 추가로 (e) `egolite` 선택 + 설치됨 + 기능 점검 실패는 `blocked(connection-failed)`로 (c)와 구분된다
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_BrowserSelection_ExplicitWinsNoFallback$' -count=1 -v`

#### AC-AF-017 — 설정 범위 밖 값은 거절 (maps REQ-AF-017)
- Given §6의 각 키에 범위 밖 값과 잘못된 형식을 하나씩 넣는다
- When 설정을 읽는다
- Then 모두 거절되고 메시지에 키 경로·허용 범위·파일이 있다. 깎아 맞춘 값으로 실행되는 경우는 0건이다
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_ConfigOutOfRange_RefusedNotClamped$' -count=1 -v`

### 병합·완료

#### AC-AF-018 — 병합 게이트와 증거 분리 (maps REQ-AF-018, REQ-AF-020)
- Given 필수 검사 결과가 `skipped` / 누락 / 실패 / 취소 / 다른 HEAD / `PASS-WITH-DEBT`인 경우 각각, 그리고 모두 정상인 경우
- When 병합 게이트가 판정한다
- Then 앞의 여섯 경우는 모두 `blocked`, 마지막만 통과한다. 증거 기록 전후로 코드 후보 HEAD가 같다
- 검증: `go test ./internal/agenticfactory/gate/ -run '^TestGate_NonPassResultsBlocked_EvidenceOutsideCodeHead$' -count=1 -v`

#### AC-AF-019 — 보호 설정 불변 (maps REQ-AF-019)
- Given 측정한 `main` 보호 설정(research.md R6)
- When AgenticFactory 전체 사이클 전후로 보호 설정을 읽는다
- Then 필수 검사 목록, `enforce_admins`, `allow_force_pushes`, `allow_deletions`, `required_linear_history` 값이 같다. 관리자 우회 병합 0건
- 검증: `gh api repos/{owner}/{repo}/branches/main/protection --jq '[.required_status_checks.contexts, .enforce_admins.enabled, .allow_force_pushes.enabled, .allow_deletions.enabled, .required_linear_history.enabled]'` 를 사이클 전후 각 1회 실행해 두 출력이 같음을 확인. 측정 기준값: `[["Lint","Build (linux/amd64)","Analyze (Go) (go)"],true,false,false,false]`(2026-10-29, 이 세션)

#### AC-AF-020 — Codex 기본값 전환 순서 (maps REQ-AF-022)
- Given `plan.md` §F.3의 세 단계
- When 단계 기록을 검사한다
- Then "미설정 시 0건" 검증 기록(AC-AF-013 PASS)보다 앞선 기본값 변경이 없고, 대체 동작 검증 전에 제거된 Codex 기능이 없다
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_CodexMigrationOrder_GateBeforeDefaults$' -count=1 -v`

### 세션 구조·명령·호환 (명칭 미확정)

#### AC-AF-021 — 역할별 쓰기 권한 (maps REQ-AF-023)
- Given ControlService / coordinator / worker 역할
- When coordinator가 claim·lease·전이를 직접 쓰려 하고, worker가 다른 attempt에 쓰려 한다
- Then 둘 다 거절된다. ControlService를 통한 요청은 성공한다
- 검증: `go test ./internal/agenticfactory/state/ -run '^TestState_RoleAuthority_CoordinatorAndForeignWorkerRefused$' -count=1 -v`

#### AC-AF-022 — 명령 종료 코드와 관찰 전용 무변경 (maps REQ-AF-024)
- Given §8의 명령 표
- When 관찰 명령을 실행하고, 각 거절 상황을 일부러 만든다
- Then 관찰 명령 전후 상태 키(저장소 순번 + 트리 키)가 같다. 거절 상황마다 §8의 종료 코드가 나온다
- 검증: `go test ./internal/agenticfactory/state/ -run '^TestCLI_ObserveIsReadOnly_ExitCodesMatchContract$' -count=1 -v`

#### AC-AF-023 — 다시 붙기·여러 run·안전 재개 (maps REQ-AF-024)
- Given run 후보가 둘 이상이다 / 재개 대상 attempt의 소유자가 살아 있다 / 호출자 권한이 원래 attempt보다 좁다
- When id 없이 붙기, 재개, 재개 후 권한 확인을 한다
- Then 첫째는 종료 코드 12와 후보 목록, 둘째는 새 attempt 없음, 셋째는 재개 뒤 권한이 넓어지지 않음
- 검증: `go test ./internal/agenticfactory/state/ -run '^TestCLI_ReconnectAmbiguityAndSafeResume$' -count=1 -v`

#### AC-AF-024 — 호환 별칭과 세션 ID·권한 불변식 (maps REQ-AF-025)
- Given 기존 `moai cc -f` / `moai cc -l`, 그리고 기존 factory 기록에 있는 lane 주소(`lane-<n>`)·세션 ID·owner·handle
- When S1 단계 변경(새 명령 등록)을 적용한 빌드로 같은 인자를 주고, 같은 기록을 다시 읽는다
- Then 도움말·인자 해석·거절 문구(`-f`+`-l` 동시 입력, `--leader`+`--factory-run`, `worker-<n>`/`agent-<n>`)·종료 코드가 변경 전 빌드와 같고, lane 주소·세션 ID·owner·handle·권한 모드가 바이트 단위로 같으며, 실행 중인 세션을 재시작하지 않는다
- 검증: `go test ./internal/cli/ -run '^TestAgenticFactoryCompat_LegacyEntryUnchanged_SessionIDsAndPermissionsPreserved$' -count=1 -v`
- 참고: 이 시험은 `internal/cli`(launcher 소유 범위)에 들어가므로 launcher 담당 레인이 쓴다(design.md §10).

#### AC-AF-025 — 설계 입력 다섯 건의 재현 시험 (maps REQ-AF-001, REQ-AF-006, REQ-AF-009, REQ-AF-023)
- Given design.md §12의 다섯 시험
- When 각 시험을 실행한다
- Then 다섯 시험이 각각 PASS이고, 각 시험은 원래 결함의 증상이 대체 동작에서 재현되지 않음을 단언한다. 이 SPEC은 해당 Issue를 닫지 않는다
- 검증(다섯 번 각각 실행):
  - `go test ./internal/agenticfactory/claim/ -run '^TestClaim_DeadOwnerExpiredAttempt_ReissuedUnderSameClaim$' -count=1 -v` (#1938)
  - `go test ./internal/agenticfactory/state/ -run '^TestState_SpeclessIssue_TransitionsWithoutSpecHash$' -count=1 -v` (#1926)
  - `go test ./internal/agenticfactory/state/ -run '^TestState_OperatorDecision_RequiresOperatorEvidence_CoordinatorRefused$' -count=1 -v` (#1924)
  - `go test ./internal/agenticfactory/claim/ -run '^TestClaim_WorktreeIdentity_NormalizedPathPreventsMismatch$' -count=1 -v` (#1847)
  - `go test ./internal/agenticfactory/claim/ -run '^TestClaim_SameNameTree_NoDivergentTreeCreated$' -count=1 -v` (#1813)

## §D 변형(mutant) 점검 메모

- AC-AF-013은 미설정 0건만 단언하면 "계수기가 아무것도 세지 않는" 변형이 통과한다. 그래서 `true`일 때 1 이상을 함께 단언한다.
- AC-AF-008은 거절만 단언하면 "모두 거절" 변형이 통과한다. 그래서 표에 있는 간선의 허용도 함께 단언한다.
- AC-AF-019는 사이클 전후 비교라 "사이클을 돌리지 않은" 변형이 통과할 수 있다. 시나리오 AC-AF-007의 PASS 기록과 같은 Issue·같은 실행에서 측정해야 한다.

## §E 완료 정의

- 25개 기준 모두 이름 붙은 시험의 `EXPECTED_RED` 기록 → `PASS` 기록이 같은 HEAD 계열에서 있다.
- 시나리오 일곱 개는 같은 실제 Issue 하나에서 측정됐다.
- 독립 감사 무부채 PASS가 있다.
- 설계 입력 다섯 건은 열린 상태로 남고, 각 Issue에 대체 동작과 시험 이름이 연결돼 있다(처분은 별도).

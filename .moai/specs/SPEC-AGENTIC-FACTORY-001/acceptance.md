# SPEC-AGENTIC-FACTORY-001 — 수용 기준

## §A 측정 규칙

- 수용 기준은 25개이며 Tier L 상한(25)과 같다. 모든 기준은 REQ 하나 이상에 연결된다.
- 각 단위 시험 명령은 `-run '^<이름>$' -v`로 시험 하나를 고른다. 판정은 출력에 `--- PASS: <이름>`이 있고 `[no tests to run]` / `[no test files]`가 없을 때만 PASS다. 종료 코드 0만으로는 PASS가 아니다.
- 시험 이름과 경로는 계약 수준의 약속이며, 구현 단계에서 경로를 바꾸면 이 문서를 manager-spec이 같이 고친다.
- 명칭(`ControlService`, `Coordinator`, `Worker`, `run`, `attempt`, `slot`, `moai factory …`)은 모두 미확정이다. 기준은 이름이 아니라 권한 경계와 동작을 단언한다.

### A.1 시나리오 하네스 계약 (AC-AF-001 ~ AC-AF-007)

- **하나의 시험 묶음**: 시나리오 일곱 개는 `internal/agenticfactory/e2e/`의 `TestScenarioSuite` 하나 안에서 순서 있는 하위 시험 `01_…` ~ `07_…`로 돈다. 모두 같은 실제 Issue(`MOAI_AF_E2E_ISSUE=<번호>`)를 쓰고, 시나리오 7은 대기 중인 두 번째 실제 Issue(`MOAI_AF_E2E_NEXT_ISSUE=<번호>`)를 쓴다(OQ-1).
- **실행 명령(일곱 AC 공통)**: `MOAI_AF_E2E_ISSUE=<n> MOAI_AF_E2E_NEXT_ISSUE=<m> go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestScenarioSuite$' -count=1 -v`
- **판정**: 각 AC는 출력의 자기 줄 `--- PASS: TestScenarioSuite/0N_<이름> `(이름 뒤 공백 포함, Go 출력은 이어서 실행 시간을 적음)이 있을 때만 PASS다. 묶음 전체의 `ok`나 종료 코드 0만으로는 어느 AC도 PASS가 아니다.
- **선행 조건 검사**: 각 하위 시험은 시작할 때 앞 단계가 남긴 상태(같은 Issue의 claim·attempt·HEAD 기록)를 확인한다. 하위 시험 하나만 따로 실행하면(`-run 'TestScenarioSuite/03'`) 선행 조건 검사에서 **FAIL**로 끝나며 skip하지 않는다. 환경 변수가 없을 때도 FAIL이다. 빈 실행이 PASS로 읽히지 않게 하려는 것이다.
- **실제와 주입의 구분**:

| 하위 시험 | 실제로 하는 것 | 주입하는 것 |
|---|---|---|
| 01 경쟁 claim | 실제 claimant 두 개가 동시에 claim | 없음 |
| 02 중복 dispatch | 실제 dispatch 재전송 | 없음 |
| 03 ACK 유실 | 실제 명령 전송 | 전송 계층 경계에서 ACK만 버림 |
| 04 중단 복구 | Worker 프로세스를 정상 종료(강제 종료 아님) | 시계를 주입해 lease 만료를 만듦 |
| 05 감사 예산 소진 | 실제 감사 라운드 기록 | `audit.max_rounds=1` + 고정 FAIL 판정 파일(하네스가 감사 역할로 등록한 시험 세션이 등록 — 생산자 결속을 지킴). `blocked(budget-exhausted)` 확인 뒤, 운영자가 승인한 새 라운드를 재개 경로로 넣어 `return_stage`로 돌아가는 것까지 확인 |
| 06 병합 결과 미상 | 아래 "05→06 증거 생산 단계"를 실제로 거친 뒤 실제 병합 요청 | 병합 응답만 버림 |
| 07 원격 반영 → 완료 → 종료 → 다음 dispatch | 원격 반영 확인, 로컬 동기화, `done` 기록, Issue 종료, 두 번째 Issue dispatch | 배포 단계가 없는 프로젝트는 사유가 적힌 `deploy: N/A` 기록 |

- **05→06 증거 생산 단계(06의 선행 조건, 별도 AC 아님)**: 05에서 `return_stage`로 돌아온 attempt가 Issue #1948이 말하는 "구현→시험→리뷰→독립감사→양측 인계→정상 병합"의 앞 단계를 **실제로** 거친다. 주입은 없다.

| 단계 | 생산자(§11 결속) | 실제/주입 |
|---|---|---|
| 구현 커밋 + 후보 PR | Worker(현재 epoch) | 실제. Issue의 작업은 작은 실제 변경이며 후보 PR은 원격에 열린다 |
| 필수 검사 3종 | GitHub(원격 조회) | 실제 |
| 코드 리뷰·보안 리뷰 | 공급자 중립 리뷰 경로(design §11.1 (a) 리뷰 세션 또는 (b) 운영자 기록). Codex 호출 0건 | 실제 |
| 독립 감사 무부채 PASS | 감사 세션(writer와 다른 세션) | 실제 |
| QA 판정 PASS | QA 세션(`/moai gate`, §F.5) | 실제 |
| writer 인계 | Worker | 실제 |
| coordinator 인계 | Coordinator 세션 | 실제 |
| `verifying` → `ready` → `landing` 전이 | ControlService | 실제 |

  06은 시작할 때 위 증거가 모두 같은 HEAD로 등록돼 있는지 확인하고, 하나라도 없으면 FAIL로 끝난다(skip 아님).
- **분기·변형 사례의 격리 실행**: 한 하위 시험 안의 분기와 변형(AC-AF-002 (a)·(b), AC-AF-007 (c)~(e))은 실제 Issue의 진행 상태를 바꾸지 않도록 **격리 사본**에서 돈다. 하네스는 그 시점의 claim 저장소를 임시 디렉터리로 복사하고, GitHub 쓰기 대신 호출을 기록만 하는 원격 기록기를 붙이고, 로컬 동기화 변형은 임시 작업 트리 사본에서 만든다. 격리 사례는 실제 경로의 해당 단계 **직전**에 돌고(AC-AF-007의 변형은 실제 `done` 기록 전), 끝나면 사본을 버린다. 실제 경로의 저장소 순번은 격리 사례 전후로 같아야 한다.
- **재실행 규칙**: 묶음은 실제 Issue 한 쌍에 대해 **한 번만** 돈다. 하위 시험 하나라도 FAIL이면 그 실행은 FAIL로 기록되고, 다시 돌리려면 (a) 새 실제 Issue 한 쌍과 (b) 새 claim 저장소 디렉터리가 필요하다. 이전 실행의 claim 저장소는 지우지 않고 읽기 전용 증거로 남기며 새 실행에 다시 쓰지 않는다(01의 "claim 없음" 전제는 새 저장소로만 만족한다). 이전 실행의 Issue는 시험이 다시 열거나 닫지 않고 멈춘 상태 그대로 두며, 처분은 운영자가 한다. 06에서 실제 병합이 이미 일어났다면 그 병합은 되돌리지 않는다.
- **두 번째 Issue의 사후 상태**: 07이 두 번째 Issue를 dispatch해 `claimed`가 되는 것을 확인한 뒤, 하네스는 ControlService를 통해 그 Issue를 `on-hold`(사유 `scenario-suite-end`, `return_stage: claimed`)로 둔다. 07은 두 번째 Issue가 `on-hold`이고 attempt가 시작되지 않았음(`implementing` 아님)을 함께 단언한다. 처분은 운영자가 한다.
- **운영자 keep-set**: 06·07의 보호된 `main` 실제 병합과 Issue 종료는 운영자 승인 대상이다. 승인 기록이 없으면 06·07은 `BLOCKED(operator-ack-missing)`로 기록되고 PASS가 아니다. 01~05는 실제 Issue에 원장 댓글을 쓰므로 그것도 운영자 승인 뒤에만 실행한다. 실제 병합은 ControlService의 병합 단계로만 하며, 이 저장소의 `auto-merge.yml`(Codex 리뷰 읽기)에 기대지 않는다.
- **Codex**: 묶음 전체는 `workflow.agentic_factory.codex.enabled` 미설정으로 돈다. 묶음 끝에 Codex 호출 계수기가 0이어야 한다(REQ-AF-013).

## §B 증거 원장

### B.1 대상 부재 참고 관측 (채택 기준점이 아님)

```
id: EV-REF-0
command: go test ./internal/agenticfactory/... -count=1
stdout:
pattern ./internal/agenticfactory/...: lstat ./internal/agenticfactory/: no such file or directory
FAIL	./internal/agenticfactory/... [setup failed]
exit_code: 1
tree_sha: 5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4
classification: TOOL_FAILURE (setup failed) — EXPECTED_RED 아님
```

이 관측은 "이 SPEC이 만들 패키지가 아직 없다"는 사실만 보인다. `tdd-result-contract.md`에 따라 `TOOL_FAILURE`는 RED가 아니므로 어느 기준의 RED-now 칸도 되지 못한다.

**채택 상태(0.2.0)**: AC-AF-019를 뺀 24개 기준은 모두 **채택 전**이며 release-blocking이 아니다. run 단계가 각 기준의 이름 붙은 시험에서 **의도한 단언 때문에** 실패하는 `EXPECTED_RED`를 기록하면(명령, 원문 출력, 종료 코드, 트리 SHA 네 요소) 그때 RED-now 칸이 채워지고 채택된다. 각 기준의 "초록 경로" 줄이 짝을 이루는 두 번째 칸이다. RED가 올바른 이유인지는 각 기준의 "RED 이유" 줄로 판단한다 — 패키지 부재(`setup failed`)나 컴파일 오류는 올바른 이유가 아니다.

### B.2 AC-AF-019 기준값 (regression-guard)

```
id: EV-AC019-BASE
command: gh api repos/modu-ai/moai-adk/branches/main/protection --jq '[.required_status_checks.contexts, .enforce_admins.enabled, .allow_force_pushes.enabled, .allow_deletions.enabled, .required_linear_history.enabled]'
stdout:
[["Lint","Build (linux/amd64)","Analyze (Go) (go)"],true,false,false,false]
exit_code: 0
measured_at: 2026-10-29T04:08:11Z
tree_sha: e14b5bbac2eb2499cbaa3ae9434595001c91cb39 (관측 대상은 원격 보호 설정이며 트리와 무관. 측정 시점 표시용)
```

### B.3 AC-AF-024 기준값 출처 (golden 파일 전 단계의 소스 근거)

```
id: EV-AC024-SRC
command: grep -n -F 'are two entry tokens; a launch carries at most one' internal/cli/factory.go
stdout:
100:	entryTokenConflict    = "%s and %s are two entry tokens; a launch carries at most one"
exit_code: 0
tree_sha: e14b5bbac2eb2499cbaa3ae9434595001c91cb39
```

```
id: EV-AC024-GOLDEN-ABSENT
command: ls internal/cli/testdata/agentic_factory_compat
stdout:
ls: internal/cli/testdata/agentic_factory_compat: No such file or directory
exit_code: 1
tree_sha: e14b5bbac2eb2499cbaa3ae9434595001c91cb39
```

golden 파일 자체는 M6 첫 커밋에서 따로 만든다(design.md §9.6, VCI §2.3).

## §C 수용 기준

### 시나리오 (하나의 실제 Issue, 처음부터 끝까지 — §A.1 하네스)

#### AC-AF-001 — 경쟁 claim (maps REQ-AF-006)
- Given 실제 Issue 하나가 `queued`이고 claim이 없다
- When 두 claimant가 동시에 claim을 요청한다
- Then 정확히 하나만 성공하고, 다른 하나는 종료 코드 10과 함께 쓰기가 거절되며, claim 기록은 하나다
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/01_CompetingClaim_ExactlyOneWins `
- RED 이유: claim 저장소가 없어 둘 다 성공하거나 claim 기록이 0개다
- 초록 경로: M4 완료 뒤(시나리오 묶음은 claim·전달·원장·게이트가 모두 있어야 끝까지 돈다)

#### AC-AF-002 — 중복 dispatch 거절 (maps REQ-AF-007)
- Given (a) 같은 Issue에 살아 있는 attempt가 있다, (b) 살아 있는 attempt는 없지만 같은 Issue로 보낸 dispatch가 아직 `sent`이고 `received`가 되지 않았다
- When 같은 Issue로 dispatch를 다시 보낸다
- Then 두 경우 모두 새 dispatch는 거절되고, 충돌한 기존 attempt id 또는 dispatch id가 기록된다. attempt 수도 dispatch 수도 늘지 않는다
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/02_DuplicateDispatch_Refused `
- RED 이유: 중복 검사가 없어 dispatch 수가 2가 된다
- 초록 경로: M4 완료 뒤

#### AC-AF-003 — ACK 유실 (maps REQ-AF-003)
- Given 명령이 `sent`이고 ACK가 오지 않았다(전송 계층 경계에서 주입). 이어서 증거 다섯 가지를 하나씩 주입한다: ① 전송 결과(성공 응답), ② idle 알림, ③ 받는 쪽의 수신 확인 행, ④ attempt 시작 행 + 트리 HEAD 관측, ⑤ 결과 증거 파일(전체 HEAD SHA 포함)
- When 재전송 판단 시점이 되고, 각 증거가 들어올 때마다 전달 상태를 읽는다
- Then 받는 쪽 상태를 먼저 조회하고, `received` 이상이면 재전송하지 않는다(전송 횟수 1). 증거별로: ①·②만으로는 상태가 `sent`에 머문다(`received`·`executed`·`completed` 어느 것도 서지 않음). ③은 `received`만 세운다. ④는 `executed`만 세운다(③ 없이 ④가 오면 `received`도 함께 기록하되 `completed`는 아님). ⑤는 ControlService가 파일을 읽어 확인한 뒤에만 `completed`를 세운다. 어떤 증거도 자기 단계보다 높은 상태를 세우지 않는다
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/03_AckLoss_RequeryNotResend `
- RED 이유: 재조회 없이 재전송해 전송 횟수가 2가 된다
- 초록 경로: M4 완료 뒤

#### AC-AF-004 — 중단 복구 (maps REQ-AF-009, REQ-AF-010)
- Given attempt가 진행 중이고 Worker가 정상 종료됐으며 lease가 만료됐다(시계 주입)
- When 복구가 시작된다
- Then 실제 실행(활성 판정, 트리 HEAD·dirty, delivery 상태)을 먼저 조회한다. 강제 종료는 0건, 동시에 살아 있는 attempt는 언제나 1개 이하이며, 새 attempt의 epoch는 옛 attempt보다 1 크다
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/04_InterruptionRecovery_QueryFirstNoDuplicate `
- RED 이유: 조회 없이 새 attempt를 만들어 살아 있는 attempt가 2개가 된다
- 초록 경로: M4 완료 뒤

#### AC-AF-005 — 감사 예산 소진 (maps REQ-AF-018)
- Given `audit.max_rounds=1`이고 고정 FAIL 판정 파일로 한 라운드를 썼다
- When 병합 게이트가 판정하고, 이어서 운영자가 승인한 새 라운드가 재개 경로로 들어온다
- Then 먼저 Issue는 `blocked(budget-exhausted)`이고 병합 게이트 판정은 PASS가 아니다. 새 라운드 승인 기록이 들어온 뒤에만 Issue가 `return_stage`로 돌아가며, 승인 기록 없이는 돌아가지 않는다
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/05_AuditBudgetExhausted_Blocked `
- RED 이유: 예산 상한 검사가 없어 병합 게이트가 계속 감사를 요청하거나 통과시킨다
- 초록 경로: M4 완료 뒤

#### AC-AF-006 — 병합 결과 미상 (maps REQ-AF-021)
- Given 실제 병합 요청 뒤 응답을 버려 결과를 모른다(운영자 keep-set 승인 필요)
- When 병합 단계가 다시 판단한다
- Then 원격(PR 상태, `main` 포함 여부)을 먼저 조회하고, 병합 요청은 다시 보내지 않는다(병합 요청 횟수 1). 승인 기록이 없으면 `BLOCKED(operator-ack-missing)`이며 PASS가 아니다
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/06_MergeOutcomeUnknown_QueryRemoteNoRemerge `
- RED 이유: 결과 미상에서 병합을 다시 요청해 요청 횟수가 2가 된다
- 초록 경로: M4 완료 뒤

#### AC-AF-007 — 원격 반영 → 완료 → 종료 → 다음 dispatch (maps REQ-AF-002, REQ-AF-021)
- Given 병합이 원격 `main`에 반영됐다. 두 번째 실제 Issue가 `queued`로 대기 중이다
- When ControlService가 반영을 확인하고 완료를 기록한다
- Then (a) 로컬 동기화 기록(동기화 전 수정·스테이징·미추적 경로마다 내용 digest·인덱스 항목·파일 모드가 동기화 뒤에도 같음)과 배포 기록 또는 사유가 적힌 `deploy: N/A`가 있을 때만 `done`이 기록된다. (b) 순서는 정확히 `done` 기록 → Issue 종료 → 두 번째 Issue dispatch이며, 두 번째 Issue는 그 뒤 `on-hold`이고 attempt가 시작되지 않았다(§A.1). 격리 사본(§A.1)에서 실제 `done` 전에 도는 변형 세 가지는 모두 Issue가 `landing`에 남고 종료도 다음 dispatch도 없다: (c) 로컬 동기화 기록을 지움, (d) 미커밋 파일 하나의 **내용만** 다른 내용으로 덮어 경로 목록은 같지만 digest가 다름, (e) 배포 기록과 `deploy: N/A`를 둘 다 지움
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/07_RemoteConfirmed_CloseThenNextDispatch `
- RED 이유: 원격 반영만으로 `done`을 기록하거나, 경로 목록만 보아 (d)를 통과시키거나, 종료 전에 다음 dispatch를 한다
- 초록 경로: M4 완료 뒤

### 상태·원장

#### AC-AF-008 — 전이표 강제와 차단 기록 (maps REQ-AF-001)
- Given design.md §2.2 전이표와 §2.1a 차단 사유 집합
- When 표에 없는 간선(예: `implementing` → `done`), 권한 없는 역할의 전이, `blocked`에서 `return_stage`가 아닌 단계로의 복귀, 집합 밖 사유나 근거 없는 `blocked` 진입을 요청한다
- Then 모두 거절되고 상태는 그대로다. 표에 있는 간선은 모두 허용된다(양방향 확인). `blocked` 상태를 성공으로 보고하는 경로가 없다
- 검증: `go test ./internal/agenticfactory/state/ -run '^TestState_TransitionTable_RefusesUnlistedAndAcceptsListed$' -count=1 -v`
- RED 이유: 전이 엔진이 없어 표에 없는 간선이 받아들여진다
- 초록 경로: M1

#### AC-AF-009 — 소유권 확인 실패 시 쓰기 거절 (maps REQ-AF-008)
- Given claim 없음 / holder 불일치 / 저장소 읽기 실패의 세 경우
- When writer가 쓰기를 시도한다
- Then 세 경우 모두 종료 코드 11로 거절되고, 확인 없는 쓰기로 넘어가지 않는다
- 검증: `go test ./internal/agenticfactory/claim/ -run '^TestClaim_OwnershipUnverified_WriteRefused$' -count=1 -v`
- RED 이유: 확인 단계가 없어 쓰기가 성공한다
- 초록 경로: M2

#### AC-AF-010 — 기한 지난 소유자의 활성 판정과 인계 (maps REQ-AF-010)
- Given lease가 만료된 attempt와, 활성 판정 입력을 주입한 다섯 경우: (a) PID 살아 있음 + 시작 시각 일치, (b) 시작 시각 미기록, (c) 시작 시각 읽기 실패, (d) PID 살아 있음 + 시작 시각 불일치(PID 재사용), (e) PID 없음
- When 다른 세션이 인계를 요청한다
- Then (a)·(b)·(c)는 `live` 또는 `indeterminate`라 인계가 거절되고 attempt는 `lapsed`가 된다. (d)·(e)는 `dead`라 epoch가 하나 오른 새 attempt가 생기고, 옛 트리·브랜치·SHA·dirty 목록이 새 attempt에 참조로 남는다. 안전 정지 기록이 있는 경우도 인계된다. 인계 뒤 옛 epoch로 들어온 쓰기는 종료 코드 15로 거절된다(fencing)
- 검증: `go test ./internal/agenticfactory/claim/ -run '^TestClaim_ExpiredLeaseLiveOwner_TakeoverRefused$' -count=1 -v`
- RED 이유: 활성 판정이 없거나 PID만 보아 (d)를 살아 있다고 보거나 (b)를 죽었다고 본다
- 초록 경로: M2

#### AC-AF-011 — 저장소와 거울의 불일치 조정 (maps REQ-AF-011)
- Given 효과 상태가 저장소와 Issue 거울에서 다르다 / 저장소가 없다
- When 조정이 돈다
- Then 첫 경우는 거울을 저장소 값으로 다시 쓴다. 둘째 경우는 `blocked(store-unavailable)`이며 거울에서 claim을 만들지 않는다
- 검증: `go test ./internal/agenticfactory/ledger/ -run '^TestLedger_Reconcile_StoreWinsAndMissingStoreBlocks$' -count=1 -v`
- RED 이유: 조정기가 없어 거울 값이 그대로 남거나 거울에서 claim이 생긴다
- 초록 경로: M3

#### AC-AF-012 — 원장 필드 권한과 비밀 가림 (maps REQ-AF-004)
- Given design.md §4의 필드 표
- When 소유자가 아닌 역할이 필드를 쓰거나, 비밀 패턴이 담긴 값이 원장·전달 기록·로그 쓰기 경로에 들어온다
- Then 첫 경우는 거절, 둘째 경우는 세 경로 모두에서 가림 처리와 가림 기록이 남고 원문은 어디에도 없다. 관찰 읽기는 제어 권한 없이 성공한다
- 검증: `go test ./internal/agenticfactory/ledger/ -run '^TestLedger_FieldOwnershipAndSecretRedaction$' -count=1 -v`
- RED 이유: 필드 권한 검사·가림 처리가 없어 쓰기가 받아들여지고 원문이 남는다
- 초록 경로: M3

### 정책

#### AC-AF-013 — Codex 미설정 시 호출 0건 (maps REQ-AF-013)
- Given `workflow.agentic_factory.codex.enabled`가 없다 / `false`다. 리뷰는 공급자 중립 경로(design §11.1)로 채운다
- When 주입 시험 환경에서 전체 사이클(claim → done)을 돈다
- Then 두 경우 모두 사이클이 `done`에 닿고 Codex 호출 계수기가 0이다. `true`일 때는 Codex 리뷰 경로가 쓰여 계수기가 1 이상이어서 계수기 자체가 작동함을 보인다. 리뷰 증거가 공급자 중립 경로에도 Codex에도 없으면 병합이 `blocked(not-configured)`로 끝난다
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_CodexUnset_ZeroCalls$' -count=1 -v`
- RED 이유: 관문이 없어 계수기 연결이 없거나 미설정에서도 호출된다
- 초록 경로: M5

#### AC-AF-014 — GLM 제외와 환경 매트릭스의 단계별 실행 (maps REQ-AF-012, REQ-AF-015)
- Given (단위) 미설치·권한 거부·용량 부족·미설정 환경과, 여섯 단계가 모두 성공하는 환경을 섞은 주입 매트릭스. (실제) design.md §5.1의 실제 환경 목록(OQ-10)
- When 매트릭스를 채운다
- Then (단위) 막힌 환경은 모두 `BLOCKED(<원인>)`와 멈춘 단계를 기록하고 `PASS`가 아니다. 실행하지 않은 단계는 `not-run`이다. 여섯 단계 모두 성공한 환경만 `PASS`다. GLM 백엔드 호출은 0건이다. (실제) 적어도 하나의 환경이 여섯 단계(설치 확인, 기동, 작업 수신, 시험, 안전 보류·재개, 결과 회수)를 실제로 실행했고, 단계마다 명령·종료 코드·출력 일부가 기록돼 있다. 모든 환경이 `BLOCKED`면 실제 시험은 FAIL이다
- 검증(두 명령 각각):
  - `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_EnvMatrix_BlockedNotPass_GLMExcluded$' -count=1 -v`
  - `go test -tags af_e2e ./internal/agenticfactory/e2e/ -run '^TestEnvMatrix_Real_AtLeastOneFullyExecuted$' -count=1 -v`
- RED 이유: 매트릭스가 없거나 실행하지 않은 단계를 성공으로 적는다
- 초록 경로: M5

#### AC-AF-015 — 인증 실패·429는 재시도 고리 없음 (maps REQ-AF-014)
- Given (a) 실행 경로의 동작에 공급자가 401 / 403 / 429를 돌려준다, (b) Issue 원장 거울 쓰기만 429를 받는다
- When 호출자가 처리한다
- Then (a) 같은 요청 재시도 0회, 상태 보존, Issue가 `blocked`(`auth-failed` / `permission-denied` / `rate-limited`) + 정상 재개 조건 기록, 새 키·용량 요청 0건. (b) 재시도 0회, `mirror_state`가 `blocked(rate-limited)`이고 Issue의 `state`는 바뀌지 않으며 실행은 저장소 기준으로 계속된다
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_AuthDenied429_NoRetryLoop$' -count=1 -v`
- RED 이유: 재시도 고리가 있거나 거울 실패가 Issue 상태를 바꾼다
- 초록 경로: M5

#### AC-AF-016 — 브라우저 선택 모델 (maps REQ-AF-016)
- Given (a) 선택 없음, (b) `aside` 선택 + 설치됨, (c) `egolite` 선택 + 미설치, (d) `egolite` 선택 + 설치됨 + 기능 점검 실패, (e) 운영자가 세 선택지 중 `aside`를 골라 저장, (f) 선택 없음 상태에서 어댑터를 여러 번 부름
- When 브라우저 어댑터를 부르고, (e)는 설정을 다시 읽는다
- Then (a) `chrome`만 호출, (b) `aside`만 호출, (c) `blocked(not-installed)` + 재개 조건이며 다른 브라우저 호출 0건, (d) `blocked(connection-failed)`로 (c)와 구분, (e) 다시 읽은 `browser.selected`가 `aside`이고 선택지는 정확히 세 개, (f) 호출 뒤에도 `browser.selected`는 미설정 그대로다(어떤 경로도 대신 고르거나 `chrome`을 저장하지 않음)
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_BrowserSelection_ExplicitWinsNoFallback$' -count=1 -v`
- RED 이유: 선택 저장이 없거나, 미설치 시 다른 브라우저로 넘어가거나, 미선택에서 `chrome`을 저장한다
- 초록 경로: M5
- 범위 밖: 설정 화면(웹 UI) 표시(spec.md §I, OQ-18)

#### AC-AF-017 — 설정 범위 밖 값은 거절 (maps REQ-AF-017)
- Given design.md §6의 각 키에 범위 밖 값과 잘못된 형식을 하나씩 넣는다(`slots`는 하한 위반 `0`과 형식 위반만. 상한은 OQ-16 미정)
- When 설정을 읽는다
- Then 모두 거절되고 메시지에 키 경로·허용 범위·파일이 있다. 깎아 맞춘 값으로 실행되는 경우는 0건이다
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_ConfigOutOfRange_RefusedNotClamped$' -count=1 -v`
- RED 이유: 검증이 없어 범위 밖 값이 그대로 또는 깎여서 쓰인다
- 초록 경로: M5

### 병합·완료

#### AC-AF-018 — 병합 게이트와 증거 분리 (maps REQ-AF-018, REQ-AF-020)
- Given 모든 증거가 갖춰진 완전한 묶음 하나와, 거기서 하나씩 바꾼 사례들:
  - 증거 제거: 코드 리뷰 / 보안 리뷰 / 독립 감사 / QA 판정 / writer 인계 / coordinator 인계 각각 하나만 없앤 여섯 사례
  - 검사 결과: 필수 검사가 `skipped` / 누락 / 실패 / 취소인 네 사례
  - 판정 품질: 감사가 `PASS-WITH-DEBT`인 사례, 증거 하나의 판정 HEAD SHA가 후보 HEAD와 다른(stale) 사례
  - 판정값: 코드 리뷰 / 보안 리뷰 / 독립 감사 / QA 각각의 판정값이 `FAIL`, `BLOCKED`, 읽을 수 없는 값인 열두 사례(증거는 있고 허용된 생산자가 등록함)
  - 생산자: (i) writer attempt의 Worker 세션이 같은 HEAD로 코드 리뷰·보안 리뷰·감사 PASS·QA PASS·coordinator 인계를 모두 등록한 완전한 묶음, (ii) 나머지는 정상이고 감사 PASS만 Worker 세션이 등록, (iii) 나머지는 정상이고 coordinator 인계만 Worker 세션이 등록, (iv) 증거 파일 안에 생산자 역할을 "auditor"라고 스스로 적었지만 등록한 세션은 Worker인 감사 증거
- When 병합 게이트가 판정한다
- Then 완전한 묶음(허용된 생산자가 모두 `PASS`로 등록)만 통과하고, 나머지 스물여덟 사례는 모두 `blocked`이며 각 사례의 `block_reason`(`check-not-passed`, `debt-present`, `verdict-not-pass`, `producer-not-allowed` 등)이 기록된다. 증거 기록 전후로 코드 후보 HEAD가 같다
- 검증: `go test ./internal/agenticfactory/gate/ -run '^TestGate_NonPassResultsBlocked_EvidenceOutsideCodeHead$' -count=1 -v`
- RED 이유: 게이트가 없거나, 증거의 존재만 보고 판정값을 보지 않거나, SHA만 보고 생산자를 보지 않아 해당 사례가 통과한다
- 초록 경로: M4

#### AC-AF-019 — 보호 설정 불변 (maps REQ-AF-019) — regression-guard
- 분류: regression-guard. 기준값(§B.2 EV-AC019-BASE)은 지금 다시 잴 수 있지만 "변하지 않음"을 지키는 기준이라 이 SPEC의 작업으로 RED → GREEN이 되지 않는다. release-blocking 자격이 없고, 통과를 "PASS"로 기록하지 않고 "불변 확인"으로 기록한다
- Given §B.2의 측정 기준값
- When AgenticFactory 시나리오 묶음(§A.1)을 돌리기 직전과 직후에 같은 명령을 실행한다
- Then 두 출력이 §B.2 기준값과 바이트 단위로 같다. 관리자 우회 병합 0건
- 검증: §B.2 `command`를 시나리오 묶음 전후 각 1회 실행
- 실패 입력: 보호 설정 하나가 바뀐 상태(예: `required_linear_history`가 `true`)의 출력이면 비교가 실패한다. 실제 보호 설정은 바꾸지 않고, 비교 로직은 기록된 출력 문자열을 바꿔 확인한다

#### AC-AF-020 — Codex 기본값 전환 순서의 운영 기록 (maps REQ-AF-022)
- Given design.md §9.7의 `progress.md` `## §J Codex Migration Stage Record` 절
- When run 단계가 plan.md §F.3의 단계를 실행하며 행을 적은 뒤 기록을 검사한다
- Then (기록) 첫 명령 출력에 `| CX-1 |`, `| CX-2 |`, `| CX-3 |` 행이 하나씩 있고, 행마다 서로 다른 전체 `stage_sha`가 있으며, CX-2 행에 AC-AF-013 PASS 증거 경로와 판정 HEAD가 있다. (순서 — 커밋 그래프가 증인, VCI §2.3) CX-1 → CX-2, CX-2 → CX-3가 각각 엄격한 조상 관계이고(두 `merge-base` 명령 exit 0, 그리고 두 SHA가 다름), §J의 `run_base_sha`부터 CX-2 커밋까지 Codex 기본값 키를 바꾼 커밋이 없다(`git log` 출력이 비어 있음). 줄 순서만 맞고 커밋 그래프가 어긋나면 FAIL이다
- 검증(네 명령 각각, `<…>`는 §J에 적힌 값):
  - `grep -n '^| CX-' .moai/specs/SPEC-AGENTIC-FACTORY-001/progress.md`
  - `git merge-base --is-ancestor <CX-1 stage_sha> <CX-2 stage_sha>`
  - `git merge-base --is-ancestor <CX-2 stage_sha> <CX-3 stage_sha>`
  - `git log --format=%H -G 'review_gate|allow_write|ReviewGate|AllowWrite|model: codex|codex: required' <run_base_sha>..<CX-2 stage_sha> -- .moai/config/sections/workflow.yaml internal/config/defaults.go internal/template/templates/.moai/config/sections/workflow.yaml`
- RED 이유: 지금 progress.md에 §J 절이 없어 첫 명령 출력이 비고 종료 코드가 1이다(run 단계에서 기록). 변형: 기본값을 먼저 바꾼 커밋 뒤에 CX 행을 순서대로 적으면 넷째 명령 출력이 비지 않아 FAIL이다
- 초록 경로: M5 이후 §F.3 3단계 제안까지

### 세션 구조·명령·호환 (명칭 미확정)

#### AC-AF-021 — 역할별 쓰기 권한 (maps REQ-AF-023)
- Given ControlService / Coordinator / Worker 역할
- When Coordinator가 claim·lease·전이를 직접 쓰려 하고, Worker가 다른 attempt에 쓰려 한다
- Then 둘 다 거절된다. ControlService를 통한 요청은 성공한다
- 검증: `go test ./internal/agenticfactory/state/ -run '^TestState_RoleAuthority_CoordinatorAndForeignWorkerRefused$' -count=1 -v`
- RED 이유: 역할 검사가 없어 Coordinator 쓰기가 성공한다
- 초록 경로: M1

#### AC-AF-022 — 관찰 모니터 표시·무변경과 종료 코드 (maps REQ-AF-005, REQ-AF-024)
- Given design.md §8의 명령 표, §8.1의 표시 항목, 프로젝트 둘(각각 Issue 둘·owner 둘)의 주입 상태
- When 관찰 명령을 실행하고, 각 거절 상황을 일부러 만든다
- Then (a) 출력이 프로젝트별로 묶이고, 프로젝트·Issue·owner마다 §8.1 여덟 항목이 있으며 항목마다 원천과 읽은 시각이 표시된다. 한 프로젝트의 항목이 다른 프로젝트 묶음에 섞이지 않는다. (b) 관찰 명령 전후 상태 키(저장소 순번 + 트리 키)가 같다. (c) 관찰 경로에 쓰기 호출이 0건이다. (d) 거절 상황마다 §8의 종료 코드가 나온다
- 검증: `go test ./internal/agenticfactory/state/ -run '^TestCLI_ObserveIsReadOnly_ExitCodesMatchContract$' -count=1 -v`
- RED 이유: 모니터가 없거나 항목이 빠지거나 실행이 저장소 순번을 바꾼다
- 초록 경로: M6

#### AC-AF-023 — 다시 붙기·여러 run·안전 재개 (maps REQ-AF-024)
- Given run 후보가 둘 이상이다 / 재개 대상 attempt의 소유자 활성 판정이 `live`다 / 호출자 권한이 원래 attempt보다 좁다
- When id 없이 붙기, 재개, 재개 후 권한 확인을 한다
- Then 첫째는 종료 코드 12와 후보 목록, 둘째는 새 attempt 없음, 셋째는 재개 뒤 권한이 넓어지지 않음
- 검증: `go test ./internal/agenticfactory/state/ -run '^TestCLI_ReconnectAmbiguityAndSafeResume$' -count=1 -v`
- RED 이유: 명령이 없거나 후보 하나를 추측해 붙는다
- 초록 경로: M6

#### AC-AF-024 — 단계별 호환 별칭과 세션 ID·권한 불변식 (maps REQ-AF-025)
- Given design.md §9.6의 golden 파일(M6 첫 커밋, S1 변경보다 먼저 커밋)과 기존 factory 기록의 lane 주소(`lane-<n>`)·세션 ID·owner·handle
- When (S1) 새 명령을 등록한 빌드로 같은 인자를 준다. (S2) 폐기 예고를 켠 빌드로 같은 인자를 준다. 두 경우 모두 같은 factory 기록을 다시 읽는다
- Then (S1) 도움말·인자 해석·거절 문구(`-f`+`-l`, `--leader`+`--factory-run`, `--leader` 단독, `worker-<n>`/`agent-<n>`/`lead-<suffix>`)·stdout·stderr·종료 코드가 golden 파일과 바이트 단위로 같다. (S2) stderr에 대체 명령을 알리는 줄이 정확히 하나 더 있고, 그 줄을 뺀 나머지는 golden 파일과 바이트 단위로 같다. 두 경우 모두 lane 주소·세션 ID·owner·handle·권한 모드가 바이트 단위로 같고, 실행 중인 세션을 재시작하지 않는다
- 검증: `go test ./internal/cli/ -run '^TestAgenticFactoryCompat_LegacyEntryUnchanged_SessionIDsAndPermissionsPreserved$' -count=1 -v`
- 기준점: §B.3(현재 소스 문자열, golden 경로 부재). golden 파일은 M6 첫 커밋으로 만든다
- RED 이유: golden 대조 시험에서 golden 문자열 하나를 바꾼 변형(예: `two entry tokens`를 `2 entry tokens`로)을 넣으면 실패해야 한다. run 단계는 이 변형 실패를 `EXPECTED_RED`로 기록한다
- 초록 경로: M6
- 참고: 이 시험은 `internal/cli`(launcher 소유 범위)에 들어가므로 launcher 담당 레인이 쓴다(design.md §10)

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
- RED 이유: 각 시험이 원래 결함의 증상(행 셋, 해시 불일치 거절, 운영자 표식 부재, 경로 불일치, 다른 트리 생성)을 단언 위반으로 드러낸다
- 초록 경로: #1938·#1847·#1813은 M2, #1926·#1924는 M1

## §D 변형(mutant) 점검 메모

- AC-AF-013은 미설정 0건만 단언하면 "계수기가 아무것도 세지 않는" 변형이 통과한다. 그래서 `true`일 때 1 이상을 함께 단언한다.
- AC-AF-008은 거절만 단언하면 "모두 거절" 변형이 통과한다. 그래서 표에 있는 간선의 허용도 함께 단언한다.
- AC-AF-018은 증거를 하나만 빼면 "그 증거만 확인하는" 변형이 통과한다. 그래서 여섯 증거를 각각 따로 뺀다. "있기만 하면 통과" 변형은 판정값 열두 사례가, "SHA만 보는" 변형은 생산자 네 사례가 잡는다. "모두 막는" 변형은 완전한 묶음의 통과 단언이 잡는다.
- AC-AF-003은 `completed`만 보면 "전송 성공으로 `received`·`executed`를 세우는" 변형이 통과한다. 그래서 증거 다섯 가지 각각이 세울 수 있는 상태와 없는 상태를 모두 단언한다.
- AC-AF-007은 경로 목록만 보는 변형을 (d)가, 배포 증거를 보지 않는 변형을 (e)가 잡는다.
- AC-AF-020은 줄 순서만 보는 변형을 커밋 그래프 명령 셋이 잡는다.
- AC-AF-010은 PID만 보는 변형이 (a)·(e)를 맞히고 (d)를 틀린다. "모두 indeterminate" 변형은 (d)·(e)의 인계 단언이 잡는다.
- AC-AF-007은 순서만 보면 "로컬 동기화 없이 done" 변형이 통과한다. 그래서 동기화 기록을 지운 변형 실행 (c)를 둔다.
- AC-AF-016은 저장만 보면 "미선택에서 chrome 저장" 변형이 통과한다. 그래서 (f)를 둔다.
- AC-AF-019는 사이클 전후 비교라 "사이클을 돌리지 않은" 변형이 통과할 수 있다. 시나리오 묶음과 같은 실행에서 직전·직후에 측정해야 한다.
- AC-AF-024는 S1만 보면 "예고 줄 여러 개" 변형을 놓친다. 그래서 S2에서 정확히 한 줄을 단언한다.

## §E 완료 정의

- AC-AF-019를 뺀 24개 기준 모두 이름 붙은 시험의 `EXPECTED_RED` 기록 → `PASS` 기록이 같은 HEAD 계열에서 있다. AC-AF-019는 시나리오 묶음 전후 불변 확인 기록이 있다.
- 시나리오 일곱 개는 같은 실제 Issue 하나에서 `TestScenarioSuite` 한 번의 실행으로 측정됐다(07은 두 번째 Issue를 함께 씀). 실패한 실행 뒤의 재실행은 §A.1 재실행 규칙을 따랐다.
- **이 완료 정의를 채워도 Issue #1948 Phase 1의 "3선택 UI … 테스트" 중 사람이 보는 설정 화면 시험은 끝나지 않는다.** 그 항목은 OQ-18의 후속 작업(선행 조건: PR #1937·#1944 착지)이 끝나야 닫힌다(spec.md §A.3, §I).
- 독립 감사 무부채 PASS가 있다.
- 설계 입력 다섯 건은 열린 상태로 남고, 각 Issue에 대체 동작과 시험 이름이 연결돼 있다(처분은 별도).

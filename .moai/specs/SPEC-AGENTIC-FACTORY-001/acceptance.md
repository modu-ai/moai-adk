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
| 02 중복 dispatch | 실제 Issue에서는 분기 (a)(살아 있는 attempt가 있는 상태에서 같은 Issue로 실제 dispatch 재전송)만 돈다. 분기 (b)는 격리 사본에서 돈다(아래) | 없음 |
| 03 ACK 유실 | 실제 명령 전송 | 전송 계층 경계에서 ACK만 버림 |
| 04 중단 복구 | Worker 프로세스를 정상 종료(강제 종료 아님) | 시계를 주입해 lease 만료를 만듦 |
| 05 감사 예산 소진 | 실제 감사 라운드 기록 | `audit.max_rounds=1` + 고정 FAIL 판정 파일(시험 세션이 등록. 그 세션의 감사 역할은 하네스가 스스로 줄 수 없고, 운영자의 대화형 단말 확인으로 ControlService가 부여한 것이어야 한다 — design §11.2. 부여 확인이 없으면 05는 `BLOCKED(operator-ack-missing)`로 끝나며 PASS가 아니다). `blocked(budget-exhausted)` 확인 뒤, 운영자가 승인한 새 라운드를 재개 경로로 넣어 `return_stage`로 돌아가는 것까지 확인 |
| 06 병합 결과 미상 | 아래 "05→06 증거 생산 단계"를 실제로 거친 뒤 실제 병합 요청 | 병합 응답만 버림 |
| 07 원격 반영 → 완료 → 종료 → 다음 dispatch | 원격 반영 확인, 로컬 동기화, `done` 기록, Issue 종료, 두 번째 Issue dispatch | 배포 단계가 없는 프로젝트는 사유가 적힌 `deploy: N/A` 기록 |

- **05→06 증거 생산 단계(06의 선행 조건, 별도 AC 아님)**: 05에서 `return_stage`로 돌아온 attempt가 Issue #1948이 말하는 "구현→시험→리뷰→독립감사→양측 인계→정상 병합"의 앞 단계를 **실제로** 거친다. 주입은 없다.

| 단계 | 생산자(§11 결속) | 실제/주입 |
|---|---|---|
| 구현 커밋 + 후보 PR | Worker(현재 epoch) | 실제. Issue의 작업은 작은 실제 변경이며 후보 PR은 원격에 열린다 |
| 필수 검사 3종 | GitHub(원격 조회) | 실제 |
| 역할 부여 | ControlService(운영자의 대화형 단말 확인, design §11.2). 코드 리뷰·보안 리뷰·감사·QA·Coordinator 다섯 역할을 서로 다른 다섯 세션에 하나씩 준다. 감사 역할은 05가 쓴 세션의 것을 그대로 쓸 수 있다(같은 역할이므로 한 세션 한 역할에 어긋나지 않음). 어느 세션도 그 run의 Worker였던 적이 없다 | 실제. 운영자 개입 시점은 두 번이다: 05 시작 전(감사 역할 하나), 05가 끝난 뒤 구현 커밋 전(나머지 넷) |
| 코드 리뷰·보안 리뷰 | 공급자 중립 리뷰 경로(design §11.1 (a) 리뷰 세션 또는 (b) 운영자 기록). Codex 호출 0건 | 실제 |
| 독립 감사 무부채 PASS | 감사 역할을 부여받은 세션 | 실제 |
| QA 판정 PASS | QA 역할을 부여받은 세션(`/moai gate`, §F.5) | 실제 |
| writer 인계 | Worker | 실제 |
| coordinator 인계 | Coordinator 역할을 부여받은 세션 | 실제 |
| `verifying` → `ready` → `landing` 전이 | ControlService | 실제 |

  06은 시작할 때 위 증거가 모두 같은 HEAD로, 허용된 생산자 결속(design §11·§11.2)과 함께 등록돼 있는지 확인하고, 하나라도 없으면 FAIL로 끝난다(skip 아님). 세션 다섯 개를 띄우는 명령과 운영자 확인을 기다리는 대기 한도는 명칭과 함께 미확정이다(spec.md OQ-19). 하네스는 확인을 기다리는 동안 skip하지 않으며, 한도를 넘으면 `BLOCKED(operator-ack-missing)`로 기록한다.
- **분기·변형 사례의 격리 실행**: 실제 Issue에서 돌지 않는 분기와 변형(AC-AF-002 (b), AC-AF-007 (c)~(h))은 실제 Issue의 진행 상태를 바꾸지 않도록 **격리 사본**에서 돈다. 하네스는 그 시점의 claim 저장소를 임시 디렉터리로 복사하고, GitHub 쓰기 대신 호출을 기록만 하는 원격 기록기를 붙인다. 로컬 동기화 변형은 **독립 gitdir과 독립 인덱스를 가진 별도 클론**에서 만든다. 카드 작업 트리는 연결 작업 트리라서 디렉터리를 복사하면 `.git` 파일이 같은 gitdir(같은 인덱스·같은 HEAD ref)을 가리키기 때문이다. 그래서 디렉터리 복사와 `git worktree add`는 쓰지 않는다. 미커밋 변경은 실제 트리에서 읽어 클론에 같은 내용·모드로 다시 만든다. claim 사본 안의 attempt 작업 트리 경로는 그 클론으로 다시 묶는다. 격리 사례는 실제 경로의 해당 단계 **직전**에 돌고(AC-AF-007의 변형은 실제 `done` 기록 전), 끝나면 사본과 클론을 버린다.
- **격리 불변 단언(격리 사례마다 전후 비교, 하나라도 다르면 해당 하위 시험 FAIL)**: (1) 클론의 `git rev-parse --absolute-git-dir` 값이 실제 트리의 값과 다르다(시작 때 한 번). (2) 실제 트리의 `git rev-parse HEAD`, `git symbolic-ref HEAD`, `git status --porcelain=v2 -z` 출력이 전후로 바이트 단위로 같다. (3) 실제 트리 gitdir의 `index` 파일 SHA-256이 전후로 같다. (4) 실제 트리의 미커밋 경로마다 작업 트리 내용 SHA-256·인덱스 항목·파일 모드가 전후로 같다. (5) 실제 claim 저장소의 순번이 전후로 같다. 저장소 순번만 같은 것으로는 통과하지 않는다.
- **재실행 규칙**: 묶음은 실제 Issue 한 쌍에 대해 **한 번만** 돈다. 하위 시험 하나라도 FAIL이면 그 실행은 FAIL로 기록되고, 다시 돌리려면 (a) 새 실제 Issue 한 쌍과 (b) 새 claim 저장소 디렉터리가 필요하다. 이전 실행의 claim 저장소는 지우지 않고 읽기 전용 증거로 남기며 새 실행에 다시 쓰지 않는다(01의 "claim 없음" 전제는 새 저장소로만 만족한다). 이전 실행의 Issue는 시험이 다시 열거나 닫지 않고 멈춘 상태 그대로 두며, 처분은 운영자가 한다. 06에서 실제 병합이 이미 일어났다면 그 병합은 되돌리지 않는다.
- **두 번째 Issue의 사후 상태**: 07이 두 번째 Issue를 dispatch해 `claimed`가 되는 것을 확인한 뒤, 하네스는 ControlService를 통해 그 Issue를 `on-hold`(사유 `operator-hold`, design §2.1a. 요청 사유 원문 `scenario-suite-end`는 근거 칸에 적는다. `return_stage: claimed`)로 둔다. 이 보류 요청은 운영자 keep-set 승인(아래)에 포함된 하네스 요청이다. 07은 두 번째 Issue가 `on-hold`이고 attempt가 시작되지 않았음(`implementing` 아님)을 함께 단언한다. 처분은 운영자가 한다.
- **운영자 keep-set**: 06·07의 보호된 `main` 실제 병합과 Issue 종료는 운영자 승인 대상이다. 승인 기록이 없으면 06·07은 `BLOCKED(operator-ack-missing)`로 기록되고 PASS가 아니다. 01~05는 실제 Issue에 원장 댓글을 쓰므로 그것도 운영자 승인 뒤에만 실행한다. 실제 병합은 ControlService의 병합 단계로만 하며, 이 저장소의 `auto-merge.yml`(Codex 리뷰 읽기)에 기대지 않는다.
- **Codex**: 묶음 전체는 `workflow.agentic_factory.codex.enabled` 미설정으로 돈다. 묶음 끝에 Codex 호출 계수기가 0이어야 한다(REQ-AF-013). 단, 계수기가 실제로 세는지는 M5의 AC-AF-013 `true` 사례가 보인다. 그래서 M4 시점에 묶음을 돌리면 이 "0"은 아직 공허하다. 이 단언을 채택 근거로 쓰는 실행은 M5 뒤의 실행이다.

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

**채택 상태(0.4.0)**: AC-AF-019를 뺀 24개 기준은 모두 **채택 전**이며 release-blocking이 아니다. run 단계가 각 기준의 이름 붙은 시험에서 **의도한 단언 때문에** 실패하는 `EXPECTED_RED`를 기록하면(명령, 원문 출력, 종료 코드, 트리 SHA 네 요소) 그때 RED-now 칸이 채워지고 채택된다. 각 기준의 "초록 경로" 줄이 짝을 이루는 두 번째 칸이다. RED가 올바른 이유인지는 각 기준의 "RED 이유" 줄로 판단한다 — 패키지 부재(`setup failed`)나 컴파일 오류는 올바른 이유가 아니다.

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
- Then 정확히 하나만 성공하고, 다른 하나는 종료 코드 10과 함께 쓰기가 거절되며, claim 기록은 하나이고 그 holder가 성공한 claimant다. 이어서 진 claimant가 같은 Issue의 상태·원장 쓰기를 시도하면 종료 코드 11로 거절되고 저장소 순번이 바뀌지 않는다(단일 writer)
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
- Given 명령이 `sent`이고 ACK가 오지 않았다(전송 계층 경계에서 주입). 이어서 증거 여섯 가지를 하나씩 주입한다: ① 전송 결과(성공 응답), ② idle 알림, ③ 받는 쪽의 수신 확인 행, ④ attempt 시작 행 + 트리 HEAD 관측, ⑤ 결과 증거 파일(전체 HEAD SHA 포함), ⑥ 받는 쪽의 답장 메시지 "완료"
- When 재전송 판단 시점이 되고, 각 증거가 들어올 때마다 전달 상태를 읽는다
- Then 재전송 판단 전에 받는 쪽 상태 조회 기록이 1회 이상 있고 그 시각이 판단보다 앞선다. 조회 결과가 `received` 이상이면 재전송하지 않는다(전송 횟수 1). 증거별로(네 상태 칸 각각을 따로 읽는다): ①·②·⑥만으로는 `received`·`executed`·`completed` 어느 칸도 서지 않는다. ③은 `received` 칸만 세운다. ④는 `executed` 칸만 세운다. ③ 없이 ④만 들어온 시점에는 `executed` 칸만 서 있고 `received` 칸은 비어 있다(REQ-AF-003: 각 상태는 자기 원천에서만 선다). ⑤는 ControlService가 파일을 읽어 확인한 뒤에만 `completed` 칸을 세운다. 어떤 증거도 자기 칸이 아닌 칸을 세우지 않는다
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/03_AckLoss_RequeryNotResend `
- RED 이유: 재조회 없이 재전송해 전송 횟수가 2가 된다
- 초록 경로: M4 완료 뒤

#### AC-AF-004 — 중단 복구 (maps REQ-AF-009, REQ-AF-010)
- Given attempt가 진행 중이고 Worker가 정상 종료됐으며 lease가 만료됐다(시계 주입)
- When 복구가 시작된다
- Then 실제 실행 조회 기록 세 항목(활성 판정 결과, 트리 HEAD·dirty, delivery 상태)이 새 attempt 발급 기록보다 앞서 남는다. 강제 종료는 0건, 동시에 살아 있는 attempt는 언제나 1개 이하이며, 새 attempt의 epoch는 옛 attempt보다 1 크다
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
- Then 원격 조회 기록 두 가지(PR 상태, `main`의 HEAD 포함 여부)가 다음 판단보다 앞서 남고, 병합 요청은 다시 보내지 않는다(병합 요청 횟수 1). 조회 결과가 반영 확인이면 07로 넘어가고, 미반영이면 `landing`에 머문다. 승인 기록이 없으면 `BLOCKED(operator-ack-missing)`이며 PASS가 아니다
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/06_MergeOutcomeUnknown_QueryRemoteNoRemerge `
- RED 이유: 결과 미상에서 병합을 다시 요청해 요청 횟수가 2가 된다
- 초록 경로: M4 완료 뒤

#### AC-AF-007 — 원격 반영 → 완료 → 종료 → 다음 dispatch (maps REQ-AF-002, REQ-AF-021)
- Given 병합이 원격 `main`에 반영됐다. 두 번째 실제 Issue가 `queued`로 대기 중이다
- When ControlService가 반영을 확인하고 완료를 기록한다
- Then (a) 로컬 동기화 기록(동기화 전 수정·스테이징·미추적 경로마다 내용 digest·인덱스 항목·파일 모드가 동기화 뒤에도 같음)과 배포 기록 또는 사유가 적힌 `deploy: N/A`가 있을 때만 `done`이 기록된다. (b) 순서는 정확히 `done` 기록 → Issue 종료 → 두 번째 Issue dispatch이며, 두 번째 Issue는 그 뒤 `on-hold`이고 attempt가 시작되지 않았다(§A.1). 격리 사본(§A.1)에서 실제 `done` 전에 도는 변형 여섯 가지는 모두 Issue가 `landing`에 남고 종료도 다음 dispatch도 없으며, 각 변형이 (b) 쪽 판정에 쓰인 항목을 기록에 남긴다: (c) 로컬 동기화 기록을 지움. (d) 미커밋 파일의 **내용만** 다른 내용으로 덮어 경로 목록·모드·인덱스는 같고 digest만 다름 — 수정된 추적 파일 하나와 미추적 파일 하나에 대해 각각 따로 돈다. (e) 배포 기록과 `deploy: N/A`를 둘 다 지움. (f) **인덱스만** 바뀜: 스테이징된 경로 하나의 인덱스 항목을 다른 blob으로 바꾸고 작업 트리 내용과 모드는 그대로 둔다. (g) **파일 모드만** 바뀜: 미커밋 경로 하나의 실행 비트만 바꾸고 내용과 인덱스는 그대로 둔다. (h) 원격 기록기가 병합 응답은 성공으로 돌려주지만 직접 조회한 원격 `main`에는 그 HEAD가 없음(REQ-AF-002 (a): 직접 조회로만 확인)
- 검증: §A.1 명령, 판정 줄 `--- PASS: TestScenarioSuite/07_RemoteConfirmed_CloseThenNextDispatch `
- RED 이유: 원격 반영만으로 `done`을 기록하거나, 병합 응답을 원격 반영으로 읽어 (h)를 통과시키거나, 경로 목록만 보아 (d)를, 내용 digest만 보아 (f)·(g)를, 수정된 파일만 보아 미추적 (d)를 통과시키거나, 종료 전에 다음 dispatch를 한다
- 초록 경로: M4 완료 뒤

### 상태·원장

#### AC-AF-008 — 전이표 강제와 차단 기록 (maps REQ-AF-001)
- Given design.md §2.2 전이표와 §2.1a 차단 사유 집합
- When 표에 없는 간선(예: `implementing` → `done`), 권한 없는 역할의 전이, `blocked`에서 `return_stage`가 아닌 단계로의 복귀, 집합 밖 사유·근거 없는·재개 조건 없는 `blocked` 진입, 사유 `operator-hold`로 `blocked` 진입을 요청한다. 대조군으로 사유 `operator-hold`로 `on-hold` 진입을 요청한다
- Then 대조군을 뺀 요청은 모두 거절되고 상태는 그대로다. 대조군은 허용되고 `return_stage`가 기록된다. 표에 있는 간선은 모두 허용된다(양방향 확인). `blocked` 상태를 조회하는 경로(상태 조회·관찰 출력·종료 코드) 어디에도 성공 표시가 없다
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
- Then (a)·(b)·(c)는 `live` 또는 `indeterminate`라 인계가 거절되고 attempt는 `lapsed`가 된다. (d)·(e)는 `dead`라 epoch가 하나 오른 새 attempt가 생기고, 옛 트리·브랜치·SHA·dirty 목록이 새 attempt에 참조로 남고, 옛 트리는 디스크에 그대로 있으며 dirty 경로의 내용 digest가 인계 전후로 같다. 안전 정지 기록이 있는 경우도 인계된다. 인계 뒤 옛 epoch로 들어온 쓰기는 종료 코드 15로 거절된다(fencing)
- 검증: `go test ./internal/agenticfactory/claim/ -run '^TestClaim_ExpiredLeaseLiveOwner_TakeoverRefused$' -count=1 -v`
- RED 이유: 활성 판정이 없거나 PID만 보아 (d)를 살아 있다고 보거나 (b)를 죽었다고 본다
- 초록 경로: M2

#### AC-AF-011 — 저장소와 거울의 불일치 조정 (maps REQ-AF-011)
- Given (1) 효과 상태가 저장소와 Issue 거울에서 다르다, (2) 저장소가 없다, (3) 사람 소유 필드(우선순위)가 Issue와 저장소 사본에서 다르다
- When 조정이 돈다
- Then (1) 거울을 저장소 값으로 다시 쓰고, 다시 쓴 거울의 `store_seq`가 저장소 순번과 같다. (2) `blocked(store-unavailable)`이며 거울에서 claim을 만들지 않는다. (3) Issue 값이 그대로 남고 저장소 사본이 Issue 값으로 다시 읽힌다(Issue 쪽을 덮어쓰지 않음)
- 검증: `go test ./internal/agenticfactory/ledger/ -run '^TestLedger_Reconcile_StoreWinsAndMissingStoreBlocks$' -count=1 -v`
- RED 이유: 조정기가 없어 거울 값이 그대로 남거나 거울에서 claim이 생긴다
- 초록 경로: M3

#### AC-AF-012 — 원장 필드 권한과 비밀 가림 (maps REQ-AF-004)
- Given design.md §4의 필드 표
- When 소유자가 아닌 역할이 필드를 쓰거나, 소유자가 형식이 틀린 값(예: 40자 16진수가 아닌 `head_sha`, 열거 밖 `state`)을 쓰거나, 비밀 패턴이 담긴 값이 원장·전달 기록·로그 쓰기 경로에 들어온다
- Then 첫째·둘째 경우는 거절, 셋째 경우는 세 경로 모두에서 가림 처리와 가림 기록이 남고 원문은 어디에도 없다. 관찰 읽기는 제어 권한 없이 성공한다
- 검증: `go test ./internal/agenticfactory/ledger/ -run '^TestLedger_FieldOwnershipAndSecretRedaction$' -count=1 -v`
- RED 이유: 필드 권한 검사·가림 처리가 없어 쓰기가 받아들여지고 원문이 남는다
- 초록 경로: M3

### 정책

#### AC-AF-013 — Codex 미설정 시 호출 0건 (maps REQ-AF-013)
- Given `workflow.agentic_factory.codex.enabled`가 없다 / `false`다. 두 경우 모두 기존 키는 이 저장소의 로컬 값(`audit.model: codex`, `audit.gates.codex: required`, `codex.review_gate.enabled: true`)으로 둔다. 리뷰는 공급자 중립 경로(design §11.1)로 채운다. 계수기는 리뷰·감사·작업 어느 경로든 Codex 실행 파일과 Codex API 호출 경계 전체에서 센다
- When 주입 시험 환경에서 전체 사이클(claim → done)을 돈다
- Then 두 경우 모두 사이클이 `done`에 닿고 Codex 호출 계수기가 0이다. `true`일 때는 Codex 리뷰 경로가 쓰여 계수기가 1 이상이어서 계수기 자체가 작동함을 보인다. 리뷰 증거가 공급자 중립 경로에도 Codex에도 없으면(미설정) 병합이 `blocked(not-configured)`로 끝나고, 그때도 계수기는 0이다(대체 호출 없음)
- 검증: `go test ./internal/agenticfactory/policy/ -run '^TestPolicy_CodexUnset_ZeroCalls$' -count=1 -v`
- RED 이유: 관문이 없어 계수기 연결이 없거나 미설정에서도 호출된다
- 초록 경로: M5

#### AC-AF-014 — GLM 제외와 환경 매트릭스의 단계별 실행 (maps REQ-AF-012, REQ-AF-015)
- Given (단위) 미설치·권한 거부·용량 부족·미설정 환경과, 여섯 단계가 모두 성공하는 환경을 섞은 주입 매트릭스. (실제) design.md §5.1의 실제 환경 목록(OQ-10)
- When 매트릭스를 채운다
- Then (단위) 막힌 환경은 모두 `BLOCKED(<원인>)`와 멈춘 단계를 기록하고 `PASS`가 아니다. 실행하지 않은 단계는 `not-run`이다. 여섯 단계 모두 성공하고 단계마다 명령·종료 코드·출력 일부가 기록된 환경만 `PASS`다. 단계는 성공했지만 셋 중 하나라도 빈 환경은 `PASS`가 아니다. GLM 백엔드 호출은 0건이다. (실제) 적어도 하나의 환경이 여섯 단계(설치 확인, 기동, 작업 수신, 시험, 안전 보류·재개, 결과 회수)를 실제로 실행했고, 단계마다 명령·종료 코드·출력 일부가 기록돼 있다. 모든 환경이 `BLOCKED`면 실제 시험은 FAIL이다
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
- Given (a) 선택 없음, (b) `aside` 선택 + 설치됨, (c) `egolite` 선택 + 미설치, (d) `egolite` 선택 + 설치됨 + 기능 점검 실패, (e) 운영자가 세 선택지 중 `aside`를 골라 저장, (f) 선택 없음 상태에서 어댑터를 여러 번 부름, (g) `aside` 선택 + 미설치
- When 브라우저 어댑터를 부르고, (e)는 설정을 다시 읽는다
- Then (a) `chrome`만 호출, (b) `aside`만 호출, (c)·(g) `blocked(not-installed)` + 재개 조건이며 다른 브라우저 호출 0건이고 `browser.selected`는 각각 `egolite`·`aside` 그대로다, (d) `blocked(connection-failed)`로 (c)와 구분, (e) 다시 읽은 `browser.selected`가 `aside`이고 선택지는 정확히 세 개, (f) 호출 뒤에도 `browser.selected`는 미설정 그대로다(어떤 경로도 대신 고르거나 `chrome`을 저장하지 않음)
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
  - 생산자: (i) writer attempt의 Worker 세션이 같은 HEAD로 코드 리뷰·보안 리뷰·감사 PASS·QA PASS·coordinator 인계를 모두 등록한 완전한 묶음, (ii) 나머지는 정상이고 감사 PASS만 Worker 세션이 등록, (iii) 나머지는 정상이고 coordinator 인계만 Worker 세션이 등록, (iv) 증거 파일 안에 생산자 역할을 "auditor"라고 스스로 적었지만 등록한 세션은 Worker인 감사 증거, (v) 나머지는 정상이고 **코드 리뷰만** Worker 세션이 등록(Worker가 자기 코드를 리뷰), (vi) 나머지는 정상이고 **보안 리뷰만** Worker 세션이 등록, (vii) 나머지는 정상이고 **QA 판정만** Worker 세션이 등록, (viii) 나머지는 정상이고 코드 리뷰를 같은 run의 **이전 attempt에서 Worker였던 세션**이 등록(그 HEAD 계열의 코드를 쓴 세션이 리뷰)
- When 병합 게이트가 판정한다
- Then 완전한 묶음(허용된 생산자가 모두 `PASS`로 등록)만 통과하고, 나머지 서른두 사례(증거 제거 6 + 검사 결과 4 + 판정 품질 2 + 판정값 12 + 생산자 8)는 모두 `blocked`이며 각 사례의 `block_reason`(`check-not-passed`, `debt-present`, `verdict-not-pass`, `producer-not-allowed` 등)이 기록된다. 생산자 여덟 사례는 모두 `producer-not-allowed`다. 증거 기록 전후로 코드 후보 HEAD가 같다
- **역할 부여(DEC-AF-01, design §11.2)** — 둘째 시험. 대조군과 사례 열 가지. 사례마다 부여 결과, 그 세션 증거의 게이트 판정, `block_reason`을 단언한다:
  - 대조군: 운영자의 유효한 `interactive-tty` 확인으로 역할이 없는 등록 세션에 감사 역할을 부여 → 부여 성공, 그 세션의 감사 PASS 증거가 게이트에서 통과로 세어진다("모두 거절" 변형을 잡음)
  - (R1) Worker 세션이 자기에게 리뷰 역할을 요청하고, 자기가 저장소에 써 넣은 ack 파일을 붙인다 → 거절, 역할 없음
  - (R2) Worker 세션이 역할이 없는 다른 등록 세션 S2에 리뷰 역할을 요청한다 → 거절. S2의 리뷰 증거는 `producer-not-allowed`
  - (R3) 운영자의 유효한 확인이 있어도 현재 attempt의 Worker 세션에 감사 역할을 부여하려 한다 → 거절(역할 분리)
  - (R4) 운영자의 유효한 확인이 있어도 같은 run의 이전 attempt에서 Worker였던 세션에 감사 역할을 부여하려 한다 → 거절
  - (R5) 이미 리뷰 역할을 가진 세션에 유효한 확인으로 감사 역할을 더 부여하려 한다 → 거절(한 세션 한 역할)
  - (R6) 위조 신원: 리뷰 역할을 받은 S2의 session id를 다른 프로세스 P3가 복사해 증거를 등록한다 → 관측 pid 불일치로 거절, `producer-not-allowed`
  - (R7) PID 재사용·중복: S2가 끝난 뒤 같은 pid·다른 시작 시각의 프로세스가 S2의 session id로 증거를 등록한다 → `dead`로 거절. 살아 있는 S2의 session id로 다른 프로세스가 다시 등록을 요청한다 → 거절
  - (R8) 운영자 인증 없음 → 부여 없음, 게이트 `blocked(operator-ack-missing)`. 같은 묶음에서 만료(대상 세션 종료 뒤), 재사용(이미 쓴 ack로 두 번째 부여), 대상 불일치(ack의 session id나 역할이 다름), 방식 불일치(`method`가 `interactive-tty`가 아님) 네 경우도 각각 부여 없음
  - (R9) ControlService가 아닌 writer가 claim 저장소 파일에 `role_grant` 기록을 직접 써 넣는다 → 역할로 세지 않고 게이트 `producer-not-allowed`, 무결성 위반 기록이 남는다
  - (R10) 대조군에서 부여한 뒤 ControlService를 재시작한다 → 이전 부여는 이어받지 않아 그 세션의 증거는 세지 않고, 게이트는 `blocked(operator-ack-missing)`
- 검증(두 명령 각각):
  - `go test ./internal/agenticfactory/gate/ -run '^TestGate_NonPassResultsBlocked_EvidenceOutsideCodeHead$' -count=1 -v`
  - `go test ./internal/agenticfactory/gate/ -run '^TestGate_RoleGrant_ControlServiceOnly_OperatorAuthenticated_OneRolePerSession$' -count=1 -v`
- RED 이유: 게이트가 없거나, 증거의 존재만 보고 판정값을 보지 않거나, SHA만 보고 생산자를 보지 않거나, 감사·coordinator 인계의 생산자만 보고 코드 리뷰·보안 리뷰·QA 생산자를 보지 않거나, 현재 attempt의 Worker만 보고 이전 attempt의 Worker를 놓치거나, 저장소의 부여 기록을 그대로 믿거나, 호출자가 적어 보낸 session id를 그대로 믿어 해당 사례가 통과한다
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
- Then (기록) 첫 명령 출력에 `| CX-1 |`, `| CX-2 |`, `| CX-3 |` 행이 하나씩 있고, 행마다 서로 다른 전체 `stage_sha`가 있으며, CX-2 행에 AC-AF-013 PASS 증거 경로와 판정 HEAD가 있다. (순서 — 커밋 그래프가 증인, VCI §2.3) CX-1 → CX-2, CX-2 → CX-3가 각각 엄격한 조상 관계이고(두 `merge-base` 명령 exit 0, 그리고 두 SHA가 다름), §J의 `run_base_sha` 다음 커밋부터 CX-2 커밋까지(포함) **모든 커밋 각각에서** design.md §9.7의 Codex 기본값 키마다 (있음/없음, 값) 짝이 `run_base_sha`의 짝과 같다. 비교는 파일을 파싱한 값으로 하며 줄 단위 정규식이 아니다. 줄 순서만 맞고 커밋 그래프가 어긋나면 FAIL이다
- 검증(네 명령 각각, `<…>`는 §J에 적힌 값):
  - `grep -n '^| CX-' .moai/specs/SPEC-AGENTIC-FACTORY-001/progress.md`
  - `git merge-base --is-ancestor <CX-1 stage_sha> <CX-2 stage_sha>`
  - `git merge-base --is-ancestor <CX-2 stage_sha> <CX-3 stage_sha>`
  - `MOAI_AF_RUN_BASE=<run_base_sha> MOAI_AF_CX2=<CX-2 stage_sha> go test ./internal/agenticfactory/policy/ -run '^TestPolicy_CodexDefaultsUnchangedThroughCX2$' -count=1 -v` — 범위 안 커밋마다 `git show <커밋>:<경로>`로 다섯 파일(`.moai/config/sections/workflow.yaml`, `internal/template/templates/.moai/config/sections/workflow.yaml`, `internal/config/defaults.go`, `.moai/config/sections/llm.yaml`, `internal/template/templates/.moai/config/sections/llm.yaml`)을 읽어 키 값을 비교하고, 비교한 커밋 수와 키 수를 출력한다. 비교한 커밋 수가 0이면 FAIL이다(빈 범위는 통과가 아님). 환경 변수가 없으면 FAIL이다
- RED 이유: 지금 progress.md에 §J 절이 없어 첫 명령 출력이 비고 종료 코드가 1이다(run 단계에서 기록). 변형 셋이 넷째 명령을 실패시켜야 한다: (1) 키 줄은 그대로 두고 값 줄만 바꾼 커밋(`workflow.yaml` L351 `enabled: true` → `false`, `defaults.go` L1733 `Enabled: false,` → `true,` — 키와 값이 다른 줄이라 줄 정규식으로는 못 잡는 경우), (2) `llm.yaml`의 `llm.defaults.sessions.codex.lane.effort`만 바꾼 커밋, (3) CX-2 전에 바꿨다가 다시 되돌린 커밋 쌍(양 끝 비교로는 못 잡는 경우). run 단계는 이 변형 실패를 `EXPECTED_RED`로 기록한다
- 초록 경로: M5 이후 §F.3 3단계 제안까지

### 세션 구조·명령·호환 (명칭 미확정)

#### AC-AF-021 — 역할별 쓰기 권한 (maps REQ-AF-023)
- Given ControlService / Coordinator / Worker 역할
- When Coordinator가 claim·lease·활성 판정·조정 결과·재개 결정·전이를 각각 직접 쓰려 하고, Worker가 다른 attempt에 쓰려 하고, Worker가 자기 attempt의 작업 기록을 쓴다
- Then Coordinator의 여섯 쓰기와 Worker의 다른 attempt 쓰기는 모두 거절된다. Worker의 자기 attempt 쓰기와 ControlService를 통한 Coordinator 요청은 성공한다("모두 거절" 변형을 잡음)
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
- Then 첫째는 종료 코드 12와 후보 목록, 둘째는 새 attempt 없음, 셋째는 재개 뒤 권한이 넓어지지 않음. 세 경우 모두 실제 실행 조회 기록이 결정보다 앞서 남고 강제 종료는 0건이다
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

## §D 변형(mutant) 점검표 — 전 AC 한 번에 (0.4.0)

plan-audit iteration-3 권고 1에 따라 25개 AC 전부를 한 번에 훑었다. 방법은 이렇다. AC마다 연결된 REQ의 조항을 하나씩 떼어, 그 AC를 통과하면서 그 조항을 어기는 그럴듯한 잘못된 구현(변형)을 적는다. 그리고 AC의 어느 단언이 그 변형을 실패시키는지 적는다. "0.3.0 판정"은 0.3.0 문구 그대로였다면 그 변형이 살아남았는지를 뜻한다. 살아남은 것은 이번 판에서 단언을 더하거나 조였고, "보강" 칸에 그 위치를 적었다. 이 점검은 문서 위의 추론이다. 변형이 실제로 실패하는지는 run 단계가 각 시험의 `EXPECTED_RED`로 확인한다(§B.1).

| AC | REQ 조항 | 변형 | AC가 잡는 이유 | 0.3.0 판정 | 보강 |
|---|---|---|---|---|---|
| 001 | 006 동시 claim 중 정확히 하나 성공 | 나중 쓰기가 이김(둘 다 성공 응답) | 진 쪽 종료 코드 10 단언 | 처치 | — |
| 001 | 006 진 claimant의 쓰기 거절 | claim만 거절하고 이후 상태·원장 쓰기는 허용 | 진 쪽 쓰기 종료 코드 11 + 순번 불변 | **생존** | AC-AF-001 Then |
| 002 | 007 살아 있는 attempt 충돌 | 살아 있는 attempt만 보고 미전달 dispatch는 안 봄 | 분기 (b) | 처치 | — |
| 002 | 007 충돌 상대 기록 | 거절만 하고 기록 없음 | 충돌 id 기록 단언 | 처치 | — |
| 003 | 003 각 상태는 자기 원천에서만 | ④가 낮은 상태 `received`를 함께 세움 | ④만 온 시점에 `received` 칸이 비어 있음 | **모순**(AC가 이 변형을 요구) | AC-AF-003 Then(D7) |
| 003 | 003 답장·idle은 상태를 못 세움 | 답장 "완료"로 `completed`를 세움 | ⑥ 주입 | **생존** | AC-AF-003 ⑥ |
| 003 | 003 재전송 전 재조회 | 조회 없이 그냥 재전송 안 함 | 조회 기록이 판단보다 앞섬 | **생존**(전송 횟수 1만 봄) | AC-AF-003 Then |
| 004 | 009 실제 실행 먼저 조회 | 조회 없이 새 attempt 발급(정상 종료라 결과는 같음) | 조회 기록 세 항목이 발급보다 앞섬 | **생존** | AC-AF-004 Then |
| 004 | 009 강제 종료 금지, 중복 금지 | 새 attempt 전에 옛 프로세스 kill | 강제 종료 0건, 살아 있는 attempt ≤1 | 처치 | — |
| 004 | 010 epoch +1 | epoch 그대로 | 새 epoch = 옛 epoch + 1 | 처치 | — |
| 005 | 018 감사 예산 소진은 blocked | 예산을 넘겨 계속 감사 요청 | `blocked(budget-exhausted)` 단언 | 처치 | — |
| 005 | 001 `return_stage`로만 복귀 | 승인 없이 복귀 / 다른 단계로 복귀 | 승인 기록 뒤에만, `return_stage`로 | 처치 | — |
| 005 | 018 감사 생산자는 부여된 역할 | 하네스가 스스로 감사 역할 등록 | §A.1 05: 운영자 확인 부여만, 없으면 BLOCKED | **생존**(D1) | §A.1 05행 |
| 006 | 021 결과 미상이면 원격 먼저 조회 | 조회 없이 기다리기만 함(재병합도 없음) | 원격 조회 기록 두 가지 | **생존** | AC-AF-006 Then |
| 006 | 021 재병합 금지 | 응답이 없으니 다시 병합 요청 | 병합 요청 횟수 1 | 처치 | — |
| 007 | 002 (a) 원격 반영은 직접 조회로 | 병합 응답 성공을 반영으로 읽음 | 변형 (h) | **생존** | AC-AF-007 (h) |
| 007 | 002 (b) 내용 digest | 경로 목록만 비교 | 변형 (d) | 처치 | — |
| 007 | 002 (b) 미추적 경로 포함 | 수정된 추적 파일만 비교 | (d)를 미추적 파일에도 따로 돌림 | **생존** | AC-AF-007 (d) |
| 007 | 002 (b) 인덱스 항목 | 내용 digest만 비교 | 변형 (f) | **생존**(D5) | AC-AF-007 (f) |
| 007 | 002 (b) 파일 모드 | 내용 digest만 비교 | 변형 (g) | **생존**(D5) | AC-AF-007 (g) |
| 007 | 002 (c) 배포 또는 N/A | 배포 증거를 안 봄 | 변형 (e) | 처치 | — |
| 007 | 002 동기화 기록 필수 | 동기화 없이 `done` | 변형 (c) | 처치 | — |
| 007 | 021 `done` → 종료 → 다음 dispatch | 종료 전에 다음 dispatch | 순서 단언 (b) | 처치 | — |
| 007 | 002 Phase 2 이관은 `done` 아님 | 이관이 `done`을 기록 | 해당 없음: 이관 기능은 이 SPEC이 만들지 않음(spec.md §I) | 해당 없음 | — |
| §A.1 | 002 (b) 격리가 실제 트리를 안 건드림 | 연결 작업 트리를 디렉터리 복사(같은 gitdir·인덱스) | gitdir 다름 + HEAD·인덱스·경로 digest 불변 단언 | **생존**(D2, 순번만 봄) | §A.1 격리 불변 단언 |
| 008 | 001 표에 없는 간선 거절 | 모두 거절 | 표의 간선 허용(양방향) | 처치 | — |
| 008 | 001 `operator-hold`는 `on-hold` 전용 | `blocked`에도 `operator-hold` 허용 | `blocked` + `operator-hold` 거절 사례 | **생존**(D6 이후 새 조항) | AC-AF-008 When |
| 008 | 001 재개 조건 기록 | 재개 조건 없이 `blocked` 진입 허용 | 재개 조건 없는 진입 거절 | **생존** | AC-AF-008 When |
| 008 | 001 `blocked`를 성공으로 보고하지 않음 | 관찰 출력만 성공 표시 | 조회 경로 셋 모두 성공 표시 없음 | **생존**(대상 경로가 막연함) | AC-AF-008 Then |
| 009 | 008 확인 실패 시 쓰기 거절 | 세 경우 중 저장소 읽기 실패만 통과 | 세 경우 각각 종료 코드 11 | 처치 | — |
| 010 | 010 `live`·`indeterminate`면 인계 거절 | PID만 봄 | (d) PID 재사용, (b) 미기록 | 처치 | — |
| 010 | 010 옛 트리 보존 | 참조는 남기고 트리는 지움 | 옛 트리 디스크 존재 + dirty digest 불변 | **생존** | AC-AF-010 Then |
| 010 | 010 오래된 epoch 쓰기 거절 | fencing 없음 | 종료 코드 15 | 처치 | — |
| 011 | 011 저장소가 이김 | 거울 값을 유지 | (1) 거울 재작성 | 처치 | — |
| 011 | 011 거울에 저장소 순번 | 순번 없이 재작성 | (1) `store_seq` 일치 | **생존** | AC-AF-011 (1) |
| 011 | 011 사람 소유 필드는 Issue가 권위 | 저장소 사본으로 Issue 필드를 덮음 | (3) | **생존** | AC-AF-011 (3) |
| 011 | 011 저장소 없으면 거울에서 재구성 금지 | 거울에서 claim 생성 | (2) | 처치 | — |
| 012 | 004 필드 소유자 검사 | 모든 역할 허용 | 비소유자 쓰기 거절 | 처치 | — |
| 012 | 004 선언된 형식 | 형식 검사 없음 | 형식 틀린 값 거절 | **생존** | AC-AF-012 When |
| 012 | 004 비밀 가림 | 원장만 가리고 로그는 안 가림 | 세 경로 모두 | 처치 | — |
| 013 | 013 미설정이면 0건 | 계수기가 아무것도 안 셈 | `true`일 때 1 이상 | 처치 | — |
| 013 | 013 대체 호출 없음 | 리뷰 증거가 없으면 Codex로 대신 호출 | `not-configured` 사례에서도 계수기 0 | **생존** | AC-AF-013 Then |
| 013 | 013 기본 활성화 없음 | 기존 `audit.model: codex`를 따라 감사에 Codex 호출 | 로컬 값을 켠 상태로 0건 + 경계 전체 계수 | **생존** | AC-AF-013 Given |
| 014 | 015 실행·성공한 단계만 PASS | 단계 기록 없이 결과만 PASS | 명령·종료 코드·출력 일부 셋 다 있어야 PASS | **생존**(단위 시험) | AC-AF-014 Then |
| 014 | 012 GLM 호출 0건 | GLM으로 시험 실행 | GLM 호출 0건 단언 | 처치 | — |
| 015 | 014 재시도 고리 없음, 새 키 없음 | 429에 한 번 재시도 | 재시도 0회, 키 요청 0건 | 처치 | — |
| 015 | 014 거울 실패는 Issue 상태 불변 | 거울 429로 Issue를 `blocked` | (b) `state` 불변 | 처치 | — |
| 016 | 016 대신 고르지 않음 | 미선택에서 `chrome` 저장 | (f) | 처치 | — |
| 016 | 016 미설치면 다른 브라우저로 안 넘어감 | `egolite`만 막고 `aside` 실패는 `chrome`으로 넘어감 | (g) | **생존** | AC-AF-016 (g) |
| 016 | 016 기존 선택 보존 | 실패하면 선택을 지움 | (c)·(g) 뒤 `browser.selected` 그대로 | **생존** | AC-AF-016 Then |
| 017 | 017 깎지 않고 거절 | 상한으로 깎음 | 깎은 값 실행 0건 | 처치 | — |
| 018 | 018 판정값 `PASS`만 | 존재만 보고 통과 | 판정값 열두 사례 | 처치 | — |
| 018 | 018 생산자 — 코드 리뷰 | 감사·coordinator 인계 생산자만 검사 | (v) | **생존**(D3) | AC-AF-018 (v) |
| 018 | 018 생산자 — 보안 리뷰 | 같음 | (vi) | **생존**(D3) | AC-AF-018 (vi) |
| 018 | 018 생산자 — QA | 같음 | (vii) | **생존**(D3) | AC-AF-018 (vii) |
| 018 | 018 한 세션 한 역할(이전 attempt 포함) | 현재 attempt의 Worker만 비교 | (viii), (R4) | **생존** | AC-AF-018 (viii)·(R4) |
| 018 | 018 부여는 ControlService만 | 저장소의 `role_grant` 기록을 그대로 믿음 | (R9) | **생존**(D1) | AC-AF-018 (R9) |
| 018 | 018 Worker 자기·타인 부여 금지 | 등록 세션의 부여 요청 허용 | (R1)·(R2) | **생존**(D1) | AC-AF-018 (R1)·(R2) |
| 018 | 018 Worker·감사 겸직 금지 | 운영자 확인만 있으면 부여 | (R3)·(R5) | **생존**(D1) | AC-AF-018 (R3)·(R5) |
| 018 | 018 신원 결속 | 호출자가 적어 보낸 session id를 믿음 | (R6)·(R7) | **생존**(D1) | AC-AF-018 (R6)·(R7) |
| 018 | 018 운영자 인증·기본 부여 없음 | ack 존재만 확인(만료·재사용·대상·방식 무시) | (R8) 다섯 경우 | **생존**(D1) | AC-AF-018 (R8) |
| 018 | 018 재시작 뒤 부여 | 저장소에서 부여를 이어받음 | (R10) | **생존**(D1) | AC-AF-018 (R10) |
| 018 | 018 정상 부여는 통과 | 모든 부여 거절 | 대조군 | 처치(새 사례에 대조군을 함께 둠) | — |
| 018 | 020 증거는 코드 HEAD 밖 | 판정 파일을 후보 브랜치에 커밋 | 기록 전후 HEAD 같음 | 처치 | — |
| 018 | 020 stale 대조 | 이전 HEAD 증거로 통과 | stale 사례 | 처치 | — |
| 019 | 019 보호 설정 불변 | 시나리오를 돌리지 않고 비교 | 같은 실행의 직전·직후 측정 | 처치(regression-guard) | — |
| 020 | 022 순서는 커밋 그래프 | 행 순서만 맞춤 | `merge-base` 두 명령 | 처치 | — |
| 020 | 022 CX-2 전 기본값 불변 — 값만 바뀜 | 키와 다른 줄의 값만 바꿈 | 커밋마다 파싱 값 비교 | **생존**(D4, `-G`) | AC-AF-020 넷째 명령 |
| 020 | 022 CX-2 전 기본값 불변 — 출시 세션 기본값 | `llm.yaml`의 codex 세션 기본값 변경 | 대상 파일에 `llm.yaml` 두 벌 | **생존**(D4) | AC-AF-020, design §9.7 |
| 020 | 022 CX-2 전 기본값 불변 — 되돌린 변경 | 바꿨다가 되돌림 | 양 끝이 아닌 커밋마다 비교 | **생존** | AC-AF-020 넷째 명령 |
| 020 | 022 빈 범위 | 범위가 비어 비교 0건인데 통과 | 비교 커밋 수 0이면 FAIL | **생존** | AC-AF-020 넷째 명령 |
| 021 | 023 Coordinator는 제어 상태를 못 씀 | claim·lease·전이만 막고 활성 판정·조정·재개 결정은 허용 | 여섯 쓰기 각각 | **생존** | AC-AF-021 When |
| 021 | 023 Worker는 자기 attempt 안에서만 | Worker 쓰기를 모두 거절 | 자기 attempt 쓰기 성공 | **생존** | AC-AF-021 Then |
| 022 | 005 여덟 항목, 원천·시각 | 거울 상태 항목 누락 | 여덟 항목 단언 | 처치 | — |
| 022 | 024 관찰 전용 | 관찰이 순번을 올림 | (b) 상태 키 불변, (c) 쓰기 0건 | 처치 | — |
| 023 | 024 id 없이 추측 금지 | 첫 후보에 붙음 | 종료 코드 12 | 처치 | — |
| 023 | 024 실제 실행 먼저 조회, 강제 종료 금지 | 조회 없이 재개 / 재개 전 kill | 조회 기록 + 강제 종료 0건 | **생존** | AC-AF-023 Then |
| 024 | 025 S1 바이트 동일 | 도움말 한 줄 바뀜 | golden 비교 | 처치 | — |
| 024 | 025 S2 예고 정확히 한 줄 | 예고 여러 줄 | 정확히 한 줄 | 처치 | — |
| 024 | 025 S4 조건이 설 때만 제거 | 조건 없이 제거 | **잡지 못함** — S4 조건을 기계로 판정할 형식이 없다 | **생존** | 없음. OQ-20(남은 생존 변형) |
| 025 | 001·006·009·023 결함 다섯 건 | 결함 증상이 그대로 남음 | 다섯 시험 각각 증상 부재 단언 | 처치 | — |

**요약**: 훑은 변형 81개(AC 25개 전부 + §A.1 격리 규칙) 중 0.3.0 문구로 살아남거나 모순이던 것이 42개였다. 41개는 이번 판에서 단언을 더하거나 조여 잡았다. 1개(AC-AF-024의 S4)는 잡지 못해 OQ-20으로 남겼다. "해당 없음" 1개(Phase 2 이관)는 이 SPEC이 그 기능을 만들지 않아서다.

## §E 완료 정의

- AC-AF-019를 뺀 24개 기준 모두 이름 붙은 시험의 `EXPECTED_RED` 기록 → `PASS` 기록이 같은 HEAD 계열에서 있다. AC-AF-019는 시나리오 묶음 전후 불변 확인 기록이 있다.
- 시나리오 일곱 개는 같은 실제 Issue 하나에서 `TestScenarioSuite` 한 번의 실행으로 측정됐다(07은 두 번째 Issue를 함께 씀). 실패한 실행 뒤의 재실행은 §A.1 재실행 규칙을 따랐다.
- **이 완료 정의를 채워도 Issue #1948 Phase 1의 "3선택 UI … 테스트" 중 사람이 보는 설정 화면 시험은 끝나지 않는다.** 그 항목은 OQ-18의 후속 작업(선행 조건: PR #1937·#1944 착지)이 끝나야 닫힌다(spec.md §A.3, §I).
- 독립 감사 무부채 PASS가 있다.
- 설계 입력 다섯 건은 열린 상태로 남고, 각 Issue에 대체 동작과 시험 이름이 연결돼 있다(처분은 별도).

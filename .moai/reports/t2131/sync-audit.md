auditor-model: claude-opus-5-5
verdict: FAIL
score: 0.88
audited_sha: 2661e35534c2a5868d983f91f7d4482c39cb38d3

## 재감사 보고서 — 카드 t2131 (#1969 ② 2/3, internal/cli Windows 권한 비트 실패)

- 대상 커밋: `2661e35534c2a5868d983f91f7d4482c39cb38d3` (감사 시작과 기록 직전에 `git rev-parse HEAD`로 두 번 확인했고, `git status --short`는 두 번 모두 출력이 없었다)
- 브랜치: `WT-internal-cli-tests`, 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2131`
- 수리 차분: `git diff fe1ab887ae..HEAD`. Go 파일 2개(`filemode_portable_t2131_test.go`에 `weakModeRefusalExpected()` 6줄 추가, `goal_mission_test.go:300`에서 1줄 바꿈), progress.md 수정, 1회차 판정 파일 추적
- SPEC과 평가 프로필은 없어서 내장 기본 프로필을 썼다(Functionality·Security 필수 통과).

### 판정 요약

- **Claude 단독 코드 판정: PASS 수준(0.88).** 1회차의 차단 결함 F1·F2가 모두 해소됐다. 새로 생긴 차단 결함은 없다.
- **기록 판정: FAIL. 코드 결함이 아니라 미충족 게이트 때문이다.** 이 트리의 감사 계획(`moai verify audit-plan`)은 `"model": "codex"`, `"cross_model_required": true`, `"enforced_required": ["codex"]`이다. 리더 지시(Codex 미승인)에 따라 codex 계열 도구를 하나도 부르지 않았으므로 required 게이트 `codex`가 채워지지 않았다. 이 상태에서는 PASS를 낼 수 없다(필수 백엔드는 응답이 없으면 이름을 밝혀 실패로 닫는다). 필요한 조치는 코드 수정이 아니라 운영자 결정이다: codex 감사를 승인하거나 이 카드에 한해 게이트를 면제한다. 둘 중 하나가 정해지기 전에는 판정이 FAIL로 남는다.

### 차원별 점수 (Claude 단독)

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | Windows 모사 overlay로 지정 묶음을 돌렸다: `--- PASS` 46개, `--- FAIL` 0개, `--- SKIP` 0개, `ok … 11.034s`, `exit=0`. `TestAutoMissionTestsReceiptTrustBoundary`가 통과했다. 양성 대조(모사 Windows에서 게이트를 강제로 true로 둠)는 `goal_mission_test.go:301: weak receipt permissions accepted`로 실패해, overlay가 실제로 적용되고 있음을 보여 준다. 호스트 8개 테스트도 통과했다. |
| Security (25%) | 85/100 | PASS | 운영 코드는 이번 수리에서 바뀌지 않았다(차분은 테스트와 문서뿐이다). POSIX 쪽은 0644 거부 단계가 그대로 실행되고 통과한다(호스트 실행). Windows에서 모드로 강제하지 않는 것은 운영자 정책(2026-11-06)에 따른 것이며, 심링크·일반 파일·범위·내용 검사는 Windows에서도 계속 돈다. |
| Craft (20%) | 85/100 | PASS | `go vet` exit=0, `GOOS=windows go vet` exit=0, 변경된 Go 파일 9개에 `gofmt -l` 출력 없음. 게이트를 테스트 전체가 아니라 단계 하나에만 걸었다. progress.md의 거짓 주장도 바로잡았다. 다만 Evidence 절은 수리 후 실측을 싣지 않았다(N1, optional). |
| Consistency (15%) | 90/100 | PASS | `permMatchesOnHost`와 같은 파일에 같은 축(`runtime.GOOS`)으로 둔 헬퍼다. 기존 패턴(`clifix_concurrency_m3_test.go:70`)과 형제 카드 t2130의 정책과도 맞는다. |

조화평균 = 4 / (1/92 + 1/85 + 1/85 + 1/90) ≈ 87.9 → 0.88. 필수 통과 차원(Functionality·Security)은 둘 다 PASS다.

### 1회차 결함 처리 결과

| ID | 1회차 | 이번 결과 | 근거 |
|---|---|---|---|
| F1 [High, blocking] | 0644 영수증 거부 단계가 Windows에서 새로 실패(301행) | **해소** | 수정 overlay에서 PASS, 같은 overlay에 게이트만 true로 바꾼 대조에서는 301행 FAIL. POSIX 호스트에서는 거부 단계가 그대로 실행돼 PASS |
| F2 [Medium, blocking] | progress.md Gaps의 거짓 주장 | **해소** | Gaps에 "CORRECTION (audit round 1, F2)" 항목이 들어갔고, 내용(287행 헬퍼, 0644 단계, `weakModeRefusalExpected()` 게이트, 나머지 단계는 Windows에서도 실행)이 코드와 일치한다(`goal_mission_test.go:278-319` 직독) |
| F3 [Low, optional] | 192행의 t2130 의존 미기재 | **기재됨** | Gaps에 "inference, not run"이라는 표시와 함께 들어갔다. 표시가 정확하다(아래 Gap G3 참고) |
| F4 [Low, optional, outside] | `migrate_cg.go:260` 같은 검사 | **기재됨** | Gaps에 범위 밖으로 기록됐다. 판정에는 영향이 없다 |

반복 이력:

| 회차 | 대상 커밋 | 판정 | 점수 | 차단 사유 |
|---|---|---|---|---|
| 1 | `fe1ab887ae` | FAIL | 0.73 | F1(Windows 0644 단계 새 실패), F2(progress 거짓 주장), codex 게이트 미충족 |
| 2 (이번) | `2661e35534` | FAIL (게이트) / 코드 판정 PASS 수준 | 0.88 | codex required 게이트 미충족만 남음. 코드 차단 결함 0건. repair-born 차단 결함 없음 |

### 발견 사항 (이번 회차)

- **N1 [Low] [optional] [in-diff]** `.moai/reports/t2131/progress.md` Evidence — 1회차 F2의 요구("overlay 실행 명령과 출력으로 바꿔 쓸 것")와 달리, 수리 후 실측(모사 Windows 실행, 수리 후 호스트 실행)이 Evidence에 실려 있지 않다. Evidence에 남은 `go test` 결과(`ok … 2.602s`)는 수리 전 실행이다. 거짓 주장은 아니다: 수리 후 단계가 Windows에서 통과한다고 측정된 것처럼 쓰지 않았다. 확신도 높음.
  - 수정(선택): 이 보고서 Evidence 2·3의 명령과 출력을 인용하거나, 기존 실행이 수리 전 커밋 기준이라고 밝힌다.
- **N2 [Low] [optional] [in-diff]** `internal/cli/goal_mission_test.go:300` — Windows에서 이 단계를 건너뛸 때 `t.Log` 사유가 남지 않는다. 1회차 수정 지침에 들어 있던 항목이지만, 운영자 정책의 "사유 출력"은 테스트 전체를 건너뛸 때만 요구된다. 여기서는 나머지 단계가 계속 실행되므로 요구 사항 위반이 아니다.
- **N3 [Info] [optional] [outside]** `internal/cli/migrate_cg_test.go:221` (`public` 하위 테스트, `os.Chmod(first.Backup, 0644)`) — 약한 모드를 쓰는 부정 단계가 하나 더 있지만, 대상 운영 코드(`migrate_cg.go:260`)는 이 카드에서 바뀌지 않았다. 따라서 이 카드 때문에 Windows에서 새로 실패하는 단계가 아니다(F4와 같은 축).

질문 (2)의 결과: 이 카드가 바꾼 판정(`hostNoGroupOrOtherAccess`)을 쓰는 곳은 `goal.go:597` 하나뿐이다. 그 판정을 시험하는 테스트 호출 지점은 `goal_mission_test.go:292/296/300/306/313/316` 6곳(모두 `TestAutoMissionTestsReceiptTrustBoundary`)이다. 그중 약한 모드 단계는 300행 하나이고 이번에 게이트가 걸렸다. `internal/cli/*_test.go`에서 `os.Chmod(`을 전부 찾아 보면, 0644/0666 계열의 나머지 단계(`managed_codex_tui_repair_test.go:102`, `managed_codex_tui_logalias_test.go:21`, `migrate_cg_test.go:221`, `mode_migrate_test.go:51` 헬퍼)는 모두 이 카드가 바꾸지 않은 운영 코드를 시험한다. 이 카드 때문에 Windows에서 새로 실패할 단계는 더 없다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. F1이 해소됐다. Windows를 모사한 지정 묶음 전체가 통과하고, overlay가 적용되고 있음은 양성 대조로 확인된다.
2. POSIX 동작은 그대로다. 호스트에서 8개 테스트가 통과하고, 0644 거부 단계는 POSIX에서 계속 실행된다.
3. vet(호스트·windows)과 gofmt가 깨끗하다.
4. progress.md 정정 내용이 코드와 일치한다.
5. 이 카드 때문에 Windows에서 새로 실패하는 약한 모드 단계는 더 없다.

### Evidence
1. Windows 모사 overlay. `filemode_portable.go`의 `hostNoGroupOrOtherAccess`가 `noGroupOrOtherAccessOn("windows", perm)`을 부르게 했고, 테스트 헬퍼 `permMatchesOnHost`는 `"windows" == "windows" || perm == want`, `weakModeRefusalExpected`는 `"windows" != "windows"`로 바꿨다. 나머지 본문은 커밋된 것과 같다. overlay 파일은 세션 scratchpad에 두었다.
   `go test -overlay=<scratchpad>/t2131r2/overlay.json ./internal/cli -run 'TestAutoMission|TestMission|TestGoal|TestNoGroupOrOtherAccessOn|TestCleanHome_RetentionFromHomeTier|TestGlmKeyFlagSavesKeySameStorage|TestJevKeyFlagSavesCredential|TestOutboxLogAppendOnly0600|TestMemoryFold_VerbatimFiling' -count=1 -v`
   ```
   46            (grep -c '^--- PASS')
   0             (grep -c '^--- FAIL')
   0             (grep -c -- '--- SKIP')
   --- PASS: TestCleanHome_RetentionFromHomeTier (0.01s)
   --- PASS: TestOutboxLogAppendOnly0600 (0.01s)
   --- PASS: TestNoGroupOrOtherAccessOn (0.00s)
   --- PASS: TestMissionBlockedQuestionPersistedShapeUnchanged (0.00s)
   --- PASS: TestAutoMissionLifecycleApprovesSealsAndRunsPublishedOperation (1.12s)
   --- PASS: TestAutoMissionProductionGitCommitOwner (1.62s)
   --- PASS: TestAutoMissionProductionLocalDevelopMergeOwner (1.64s)
   --- PASS: TestAutoMissionTestsReceiptTrustBoundary (0.00s)
   --- PASS: TestAutoMissionLifecycleRefusalAndJSONBranches (0.00s)
   --- PASS: TestAutoMissionTamperedSealedContractBlocksBeforeQueueEffect (0.26s)
   --- PASS: TestAutoMissionSupervisorSplitWorktreesCommitMergeAndCompletion (2.50s)
   --- PASS: TestAutoMissionExternalDeliveryWithoutAuthoritativeProviderIsStableBlocked (0.25s)
   --- PASS: TestGoalMissionSupervisorArgumentRefusals (0.26s)
   --- PASS: TestGoalMissionOperationEarlyRefusalBranches (0.53s)
   --- PASS: TestGlmKeyFlagSavesKeySameStorage (0.00s)
   --- PASS: TestJevKeyFlagSavesCredential (0.00s)
   --- PASS: TestMemoryFold_VerbatimFiling (0.19s)
   PASS
   ok  	github.com/modu-ai/moai-adk/internal/cli	11.034s
   exit=0
   ```
   지정 정규식에 걸리지 않는 goal_mission 테스트 2개(CI 97번 행 포함)를 같은 overlay로 따로 돌렸다:
   `go test -overlay=… ./internal/cli -run 'TestGTDAutonomyEndToEndProductionCLIWithGovernanceReceipts|TestNewAutoMissionCommandTreatsTextAsData' -count=1 -v`
   ```
   --- PASS: TestNewAutoMissionCommandTreatsTextAsData (0.00s)
   --- PASS: TestGTDAutonomyEndToEndProductionCLIWithGovernanceReceipts (2.46s)
   ok  	github.com/modu-ai/moai-adk/internal/cli	3.673s
   ```
2. 양성 대조. 운영 코드는 모사 Windows로 두고, 테스트 헬퍼의 `weakModeRefusalExpected`만 `return true`(수리 전 동작)로 바꿨다:
   `go test -overlay=<scratchpad>/t2131r2/ctl/overlay.json ./internal/cli -run 'TestAutoMissionTestsReceiptTrustBoundary' -count=1 -v` → `exit=1`
   ```
       goal_mission_test.go:301: weak receipt permissions accepted
   --- FAIL: TestAutoMissionTestsReceiptTrustBoundary (0.02s)
   FAIL	github.com/modu-ai/moai-adk/internal/cli	1.565s
   ```
3. 호스트(darwin/arm64, 실제 POSIX): `go test ./internal/cli -run 'TestNoGroupOrOtherAccessOn|TestCleanHome_RetentionFromHomeTier|TestGlmKeyFlagSavesKeySameStorage|TestJevKeyFlagSavesCredential|TestOutboxLogAppendOnly0600|TestMissionBlockedQuestionPersistedShapeUnchanged|TestMemoryFold_VerbatimFiling|TestAutoMissionTestsReceiptTrustBoundary' -count=1 -v` → `exit=0`
   ```
   --- PASS: TestCleanHome_RetentionFromHomeTier (0.02s)
   --- PASS: TestOutboxLogAppendOnly0600 (0.00s)
   --- PASS: TestNoGroupOrOtherAccessOn (0.00s)
   --- PASS: TestMissionBlockedQuestionPersistedShapeUnchanged (0.00s)
   --- PASS: TestAutoMissionTestsReceiptTrustBoundary (0.00s)
   --- PASS: TestGlmKeyFlagSavesKeySameStorage (0.02s)
   --- PASS: TestJevKeyFlagSavesCredential (0.01s)
   --- PASS: TestMemoryFold_VerbatimFiling (0.26s)
   PASS
   ok  	github.com/modu-ai/moai-adk/internal/cli	2.314s
   ```
4. `go vet ./internal/cli` → `vet exit=0`. `GOOS=windows go vet ./internal/cli` → `winvet exit=0`. `gofmt -l internal/cli/{filemode_portable.go,filemode_portable_t2131_test.go,goal.go,goal_mission_test.go,clean_home_test.go,jev_key_test.go,feedback_participation_test.go,goal_blocked_question_regression_test.go,memory_fold_test.go}` → 출력 없음, `gofmt exit=0`.
5. 판정 소비자 검색: `grep -rn 'hostNoGroupOrOtherAccess\|noGroupOrOtherAccessOn' internal/` → 운영 코드에서 쓰는 곳은 `internal/cli/goal.go:597` 하나뿐이다. `grep -rn 'validateMissionTestsReceipt\|tests_receipt_unsafe' internal/cli` → 테스트 호출 지점은 `goal_mission_test.go:292,296,300,306,313,316`이다.
6. 감사 계획: `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2131` → `exit=0`, `"config_status": "ok"`, `"model": "codex"`, `"cross_model_active": true`, `"cross_model_required": true`, `"enforced_required": ["codex"]`.

### Baseline-attribution
모든 명령은 이번 실행에서 이 트리(HEAD `2661e35534`, clean)를 대상으로 돌렸다. 도구 체인은 `go version go1.27.3 darwin/arm64`다. `go test`/`go vet`는 트리 소스를 직접 빌드하므로, 판정한 빌드가 곧 이 트리다. overlay 파일은 `/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/98f002f8-4741-40c2-95ef-a06f9bdaac8b/scratchpad/t2131r2/` 아래에만 두었고, 추적 파일은 이 판정 파일 말고는 수정하지 않았다.

### Gaps
- **G1 미충족 게이트(PASS 차단):** required 게이트 `codex`. 리더 지시에 따라 `codex_audit`·`audit_multi`·`glm_audit`·`claude_audit`와 app-server 경로를 부르지 않았다. 감사 영수증은 발급되지 않았다(receipts=none). 이 판정은 Claude 단독 판정이다.
- **G2 실제 Windows 미실행:** overlay는 `internal/cli`의 `runtime.GOOS` 분기만 모사한다. 실제 Windows의 `os.Stat` 0666 보고, `os.Chmod`·`os.Symlink` 의미는 모사하지 않았다.
- **G3 t2130 의존성은 이 모사로 측정되지 않음:** 모사 실행에서 goal_mission 4행(74·75·76·97, 각각 123·192·250·489행)이 통과한 것은 `internal/mission`이 darwin에서 실제로 돌아 0600 파일을 받기 때문이다. `internal/mission`의 `Perm() != 0o600`(governance 영수증)과 lease 검사는 모사하지 않았다. 따라서 이 통과를 "t2130 없이도 Windows에서 통과한다"는 증거로 읽으면 안 된다. 이 행들의 Windows 통과 여부는 t2130 착지와 CI에 달려 있다. 이것은 공개된 의존성이고 이 카드의 결함이 아니다. progress.md의 "inference, not run" 표시는 이 상태를 정확히 나타낸다.
- 패키지 전체 스위트와 커버리지는 측정하지 않았다(지정 묶음만 돌렸다).
- `moai verify audit-plan`은 설치된 바이너리로 돌렸고, 그 빌드의 커밋을 트리 HEAD와 대조하지 않았다(§2.2). 결과는 게이트 상태 확인에만 썼다.
- 워크트리 격리 가드가 명령 3건을 거부했다(파이프가 붙은 호스트 `go test`, 변수가 든 `sed` overlay 생성, 복합 `git diff` 범위 계산). 대신 파일 리다이렉트 형태로 다시 돌렸고, overlay 파일은 Write 도구로 만들었다. 위 증거는 그 재실행 결과다. 카드 diff의 merge-base 통계는 끝내 다시 돌리지 않았다. 수리 범위는 `git diff fe1ab887ae..HEAD`로 확인했다.

### Residual-risk
- Windows에서 tests 영수증의 소유자 전용 속성은 모드로 강제되지 않는다(운영자 정책, t2130과 같다). POSIX 쪽은 변함없이 강제된다.
- t2130이 착지하기 전에는 Windows CI에서 goal_mission 4행이 계속 빨간불일 수 있다(G3).
- `weakModeRefusalExpected`는 테스트 쪽 정책 사본이다. 운영 판정(`noGroupOrOtherAccessOn`)의 Windows 분기가 나중에 바뀌어도 이 헬퍼는 함께 바뀌지 않는다. 이 차이는 POSIX CI에서 보이지 않고 Windows CI에서만 드러난다.

## 권고
1. 운영자 결정(G1): codex 감사를 승인하거나 이 카드에 한해 codex 게이트를 면제한다. 코드 수정은 필요 없다.
2. (선택) N1: progress.md Evidence에 수리 후 실측을 인용한다.
3. t2130 착지 뒤 Windows CI에서 goal_mission 4행이 초록인지 확인한다.

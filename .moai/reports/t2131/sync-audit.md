auditor-model: claude-opus-5-5
verdict: FAIL
score: 0.73
audited_sha: fe1ab887aece56810c6d0f142e5f23eb108552f8

## 감사 보고서 — 카드 t2131 (#1969 ② 2/3, internal/cli Windows 권한 비트 실패 11건)

- 대상 커밋: `fe1ab887aece56810c6d0f142e5f23eb108552f8` (`git rev-parse HEAD`로 확인, 트리 clean)
- 브랜치: `WT-internal-cli-tests`, 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2131`
- 연관 SPEC/평가 프로필: 없음. 내장 기본 프로필을 썼다(Functionality·Security 필수 통과).
- 반복 이력: 1회차(이번 감사)

### 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 55/100 | FAIL | Windows를 모사한 overlay 실행 결과 `goal_mission_test.go:301: weak receipt permissions accepted` / `--- FAIL: TestAutoMissionTestsReceiptTrustBoundary`. 카드가 고쳤다고 주장한 CI 행(77번, `goal_mission_test.go:317`)이 Windows에서는 여전히 실패하고, 실패 위치만 301행으로 옮겨 간다. |
| Security (25%) | 85/100 | PASS | POSIX 조건은 원래 조건과 같다(아래 F-증명). Windows에서 모드 비트 검사를 빼는 것은 운영자 정책(2026-11-06)과 일치한다. 심링크 거부·일반 파일 확인·범위 검사·내용 검사는 그대로 남아 있다. |
| Craft (20%) | 80/100 | PASS | `go vet` exit=0, `GOOS=windows go vet` exit=0, `gofmt -l` 출력 없음. 판정 함수 표 테스트가 변이 2종을 모두 잡았다. 다만 progress.md Gaps에 측정과 반대되는 주장이 있다(F2). |
| Consistency (15%) | 90/100 | PASS | 기존 패턴(`runtime.GOOS != "windows" && perm != …`, `clifix_concurrency_m3_test.go:70`)과 같은 축을 따른다. 형제 카드 t2130과 정책도 일치한다. |

필수 통과 방화벽: Functionality가 FAIL이므로 점수와 관계없이 전체 판정은 FAIL이다. 별도로 아래의 미충족 게이트도 PASS를 막는다.

### 발견 사항 (구조화된 결함 목록)

- **F1 [High] [blocking] [in-diff + 재현 증거]** `internal/cli/goal.go:597` (변경 행)와 `internal/cli/goal_mission_test.go:299-302` — 운영 코드가 Windows에서 약한 모드를 더는 거부하지 않는다. 그런데 `TestAutoMissionTestsReceiptTrustBoundary`는 0644 영수증(`passed`, head 일치, 범위 안)을 쓰고 거부를 기대하는 단계를 그대로 갖고 있다. Windows에서는 이 단계가 새로 실패한다(실패 위치 317행 → 301행). 즉 categorized.tsv 77번 행은 이 커밋으로 해소되지 않는다. 형제 감사(t2130)에서 이미 지적된 결함 유형이다. 확신도: 높음(overlay 재현).
  - 필요한 수정: 운영자 정책("Windows에서는 이식 가능한 부분만 검사, 남는 것이 없을 때만 사유를 출력하고 건너뜀")에 따라 299-302행의 약한 권한 단계만 `runtime.GOOS != "windows"`로 감싸고, Windows에서는 `t.Log`로 사유를 남긴다. 테스트 전체가 아니라 그 단계만 건너뛰어야 하고, 나머지(실패·stale·범위 밖·심링크·유효 영수증) 단계는 Windows에서도 그대로 실행돼야 한다. POSIX 단언은 바이트 단위로 유지한다. 수정 후 이 보고서의 overlay(아래 명령 3)로 다시 확인한다.
- **F2 [Medium] [blocking] [in-diff]** `.moai/reports/t2131/progress.md` Gaps 3번째 항목 — "goal_mission_test.go only chmods to a caller-provided mode at line 287 … checked, none needs gating"은 측정 결과와 반대다. 299행이 그 헬퍼에 0644를 넘긴다. 검증하지 않은 주장(verification-claim-integrity §1)이다.
  - 필요한 수정: F1 수정과 함께 이 항목을 실제 측정(overlay 실행 명령과 출력)으로 바꿔 쓴다.
- **F3 [Low] [optional] [in-diff]** `.moai/reports/t2131/progress.md` Evidence/Gaps — `goal_mission_test.go:192`(TestAutoMissionProductionGitCommitOwner, `tests_receipt_unsafe`)가 이 카드로 해소되는지 명시돼 있지 않다. `goal.go`에서 tests 영수증 검사(769행)는 governance 영수증 적재(809행, `mission.LoadGovernanceReceipts`)보다 먼저 실행된다. 그리고 `internal/mission/governance_receipt.go:207`이 `Perm() != 0o600`을 요구하므로, Windows에서는 다음 단계에서 `mission governance: receipt_unsafe`로 실패할 가능성이 높다(추론, 실측 아님, 확신도 중간). 이 행도 t2130 착지에 의존한다고 밝히는 편이 정확하다.
  - 필요한 수정: Gaps에 "192행은 이 카드로 tests 영수증 단계를 넘어가지만 governance 영수증 검사(t2130 범위)에서 다시 막힐 수 있음"을 적는다.
- **F4 [Low] [optional] [outside]** `internal/cli/migrate_cg.go:260` — 같은 `Perm()&0077 != 0` 검사가 남아 있다. 이번 11건 실패 목록에는 없으므로 판정에는 영향이 없다. 운영자에게 보이는 참고 사항이다.

### 운영자 정책 대조 — CI 11건 매핑 (categorized.tsv, POSIX_MODE_PERM, 패키지 internal/cli)

| 행 | 테스트 / 위치 | 이 커밋의 처리 | 감사 결과 |
|---|---|---|---|
| 15 | TestCleanHome_RetentionFromHomeTier `clean_home_test.go:272` | `permMatchesOnHost` | 처리됨 |
| 124 | TestGlmKeyFlagSavesKeySameStorage `jev_key_test.go:196` | `permMatchesOnHost` | 처리됨 |
| 160 | TestJevKeyFlagSavesCredential `jev_key_test.go:229` | `permMatchesOnHost` | 처리됨 |
| 147 | TestOutboxLogAppendOnly0600 `feedback_participation_test.go:111` | `permMatchesOnHost` | 처리됨 |
| 148 | TestMissionBlockedQuestionPersistedShapeUnchanged `goal_blocked_question_regression_test.go:79` | `permMatchesOnHost` | 처리됨 |
| 168 | TestMemoryFold_VerbatimFiling `memory_fold_test.go:750` | `permMatchesOnHost` | 처리됨 |
| 77 | TestAutoMissionTestsReceiptTrustBoundary `goal_mission_test.go:317` | 운영 코드 `hostNoGroupOrOtherAccess` | **해소 안 됨 — 301행에서 새로 실패 (F1)** |
| 75 | TestAutoMissionProductionGitCommitOwner `goal_mission_test.go:192` | 운영 코드 | tests 영수증 단계는 통과, 이후 t2130 의존 가능(F3, 미측정) |
| 74 | TestAutoMissionLifecycle… `goal_mission_test.go:123` | 없음 (internal/mission, t2130) | 의존성 공개됨 |
| 76 | TestAutoMissionProductionLocalDevelopMergeOwner `goal_mission_test.go:250` | 없음 (internal/mission `git_owner.go:156`, t2130) | 의존성 공개됨 |
| 97 | TestGTDAutonomyEndToEnd… `goal_mission_test.go:489` | 없음 (internal/mission, t2130) | 의존성 공개됨 |

이 밖에 테스트 쪽 POSIX 비트 단언 중 빠진 것은 없다. 1159번 행(`internal/cli/update`)은 다른 패키지라 범위 밖이다.

### 증명 — POSIX 조건 동치

원래 조건: `info.Mode().Perm()&0o077 != 0` → 거부. 새 조건: `!hostNoGroupOrOtherAccess(perm)` = `!noGroupOrOtherAccessOn(runtime.GOOS, perm)`. `goos != "windows"`이면 `!(perm&0o077 == 0)` ≡ `perm&0o077 != 0`로 같다. `||` 단락 순서(err → 심링크 → 일반 파일 → 권한)도 바뀌지 않았다. 아래 명령 1(호스트 darwin)에서 0644 거부 단계가 통과했고, 명령 4·5의 변이 실험이 이 조건을 테스트가 붙잡고 있음을 보여 준다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. 호스트(POSIX)에서 지정된 8개 테스트는 통과한다.
2. Windows에서는 `TestAutoMissionTestsReceiptTrustBoundary`가 301행에서 새로 실패한다(F1).
3. vet(호스트·windows)과 gofmt는 깨끗하다.
4. 판정 함수는 변이 테스트에 걸린다.

### Evidence
1. `go test ./internal/cli -run 'TestNoGroupOrOtherAccessOn|TestCleanHome_RetentionFromHomeTier|TestGlmKeyFlagSavesKeySameStorage|TestJevKeyFlagSavesCredential|TestOutboxLogAppendOnly0600|TestMissionBlockedQuestionPersistedShapeUnchanged|TestMemoryFold_VerbatimFiling|TestAutoMissionTestsReceiptTrustBoundary' -count=1 -v`
   ```
   --- PASS: TestCleanHome_RetentionFromHomeTier (0.03s)
   --- PASS: TestOutboxLogAppendOnly0600 (0.01s)
   --- PASS: TestNoGroupOrOtherAccessOn (0.00s)
   --- PASS: TestMissionBlockedQuestionPersistedShapeUnchanged (0.00s)
   --- PASS: TestAutoMissionTestsReceiptTrustBoundary (0.01s)
   --- PASS: TestGlmKeyFlagSavesKeySameStorage (0.01s)
   --- PASS: TestJevKeyFlagSavesCredential (0.00s)
   --- PASS: TestMemoryFold_VerbatimFiling (0.30s)
   PASS
   ok  	github.com/modu-ai/moai-adk/internal/cli	2.608s
   ```
2. `go vet ./internal/cli` → `vet exit=0`; `GOOS=windows go vet ./internal/cli` → `winvet exit=0`; `gofmt -l <변경된 Go 파일 8개>` → 출력 없음, `gofmt exit=0`.
3. Windows 모사(overlay: `hostNoGroupOrOtherAccess`가 `noGroupOrOtherAccessOn("windows", perm)`을 부르고, `permMatchesOnHost`가 항상 true를 돌려주도록 교체), 명령 1과 같은 `-run`:
   ```
   --- PASS: TestCleanHome_RetentionFromHomeTier (0.21s)
   --- PASS: TestOutboxLogAppendOnly0600 (0.06s)
   --- PASS: TestNoGroupOrOtherAccessOn (0.00s)
   --- PASS: TestMissionBlockedQuestionPersistedShapeUnchanged (0.05s)
       goal_mission_test.go:301: weak receipt permissions accepted
   --- FAIL: TestAutoMissionTestsReceiptTrustBoundary (0.06s)
   --- PASS: TestGlmKeyFlagSavesKeySameStorage (0.03s)
   --- PASS: TestJevKeyFlagSavesCredential (0.04s)
   --- PASS: TestMemoryFold_VerbatimFiling (1.12s)
   FAIL	github.com/modu-ai/moai-adk/internal/cli	4.127s
   ```
4. 변이 A(지시받은 변이: `if goos == "windows" || goos == "linux" { return true }`), `-run 'TestNoGroupOrOtherAccessOn|TestAutoMissionTestsReceiptTrustBoundary'`:
   ```
   filemode_portable_t2131_test.go:31: noGroupOrOtherAccessOn("linux", 640) = true, want false
   filemode_portable_t2131_test.go:31: noGroupOrOtherAccessOn("linux", 604) = true, want false
   --- FAIL: TestNoGroupOrOtherAccessOn (0.00s)
   --- PASS: TestAutoMissionTestsReceiptTrustBoundary (0.00s)
   ```
   변이가 죽었다(호스트가 darwin이라 운영 경로 테스트는 linux 변이의 영향을 받지 않는다).
5. 변이 B(모든 OS에서 `return true`):
   ```
   filemode_portable_t2131_test.go:31: noGroupOrOtherAccessOn("darwin", 666) = true, want false
   --- FAIL: TestNoGroupOrOtherAccessOn (0.00s)
       goal_mission_test.go:301: weak receipt permissions accepted
   --- FAIL: TestAutoMissionTestsReceiptTrustBoundary (0.00s)
   ```
6. `moai verify audit-plan --project-root <이 트리>` → `"config_status": "ok"`, `"model": "codex"`, `"cross_model_active": true`, `"enforced_required": ["codex"]`.
7. 보안 점검(grep): 변경된 운영 코드는 권한 판정 한 줄뿐이다. 새 입력 표면·비밀값·외부 명령은 없다. 심링크(`ModeSymlink`), 일반 파일(`IsRegular`), 범위(`filepath.Rel`), 내용(`status`/`head_sha`) 검사는 diff에서 손대지 않았다.

### Baseline-attribution
모든 명령은 이번 실행에서 이 트리(HEAD `fe1ab887ae`, clean)를 대상으로 돌렸다. 도구 체인은 `go version go1.27.3 darwin/arm64`이다. `go test`/`go vet`는 트리 소스를 직접 빌드하므로 판정한 빌드가 곧 트리다. overlay 파일은 세션 scratchpad에 두었고 추적 파일은 수정하지 않았다.

### Gaps
- **미충족 게이트(PASS 차단):** 이 트리의 감사 계획은 `codex` 게이트를 `required`로 둔다. 리더 지시(Codex 미승인)에 따라 `codex_audit`·`audit_multi`·`glm_audit`·`claude_audit`를 부르지 않았다. 이 판정은 Claude 단독 판정이고 codex 게이트는 미충족이다. 판정이 FAIL이라 결과에는 영향이 없지만, 수정 후 PASS를 내려면 이 게이트를 운영자가 처리해야 한다.
- 실제 Windows 호스트에서는 돌리지 않았다. Windows 동작은 overlay로 `runtime.GOOS` 분기만 모사한 것이다. 실제 Windows의 `os.Stat` 0666 보고, `os.Symlink` 권한, `os.Chmod` 의미는 모사하지 않았다. 다만 F1의 실패 지점(301행)은 이 차이와 무관하다. 모드 검사가 Windows에서 무조건 참이 되는 순간 0644 단계가 받아들여지기 때문이다.
- F3(192행이 governance 검사에서 다시 막힐 가능성)은 코드 순서와 `governance_receipt.go:207`을 읽고 추론한 것이다. 실행으로 확인하지는 않았다.
- `TestAutoMission*`/`TestGoal*` 전체 실행과 패키지 전체 스위트는 돌리지 않았다. 지정 테스트와 변이 대상만 실행했다.
- 커버리지는 측정하지 않았다(새 운영 코드는 26줄 파일 하나이고, 판정 함수의 두 분기는 표 테스트가 모두 친다).
- `moai verify audit-plan`은 설치된 바이너리로 실행했다. 그 빌드의 커밋을 트리 HEAD와 대조하지는 않았다(§2.2). 결과는 게이트 상태 확인에만 썼다.
- 워크트리 격리 가드가 명령 2건(파이프·런타임 변수를 포함한 복합 명령)을 거부했다. 같은 명령을 단순 형태로 나눠 다시 실행했고, 위 증거는 그 재실행 결과다.

### Residual-risk
- F1을 고친 뒤에도 TestAutoMissionTestsReceiptTrustBoundary가 Windows에서 심링크 단계(309-315행)를 지나는지는 CI가 판단한다(CI 실패가 317행이었으므로 Symlink 생성 자체는 CI에서 성공한 것으로 보인다).
- Windows에서 tests 영수증의 소유자 전용 속성은 모드로 강제되지 않는다(운영자 정책, t2130과 같다).
- goal_mission 실패 4건(74·75·76·97행)은 t2130이 착지하기 전에는 Windows CI에서 계속 빨간불일 수 있다.

## 권고
1. F1 수정: 299-302행 단계만 Windows에서 사유 로그와 함께 건너뛴다. POSIX 단언은 그대로 둔다.
2. F2·F3: progress.md의 Gaps를 실측으로 바로잡는다.
3. 재감사 범위는 F1-F3 차분과 codex 게이트 처리로 한정한다.

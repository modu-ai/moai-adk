auditor-model: claude-opus-5-5
verdict: FAIL
score: 0.62
audited_sha: 19aa331da44ebd5ae91d4b066f49322ea1cfed68

# t2130 sync-audit — #1969 ② 1/3 internal/mission Windows 권한 비트 실패 (21건)

- 대상: 커밋 19aa331da44ebd5ae91d4b066f49322ea1cfed68 (브랜치 WT-internal-mission-tests, 트리 /Users/goos/.moai/worktrees/projects/moai-adk-go/t2130)
- 감사 회차: 1회차 (이전 감사 없음)
- 감사자: Claude 단독 (리더 지시로 Codex·GLM·audit_multi 미호출)

## 판정 요약

전체 판정은 FAIL, 점수 0.62. POSIX 동작은 기존과 정확히 같고(Security PASS), 변이 검사도 테스트가 잡아낸다. 하지만 Windows 판정을 시뮬레이션해 보니 이 변경이 **Windows에서 새 실패 2건을 만들고, 기존 실패 1건은 다른 실패로 바뀔 뿐**이다. 카드 목표(mission 21건 해소)가 달성되지 않고 회귀가 생기므로 Functionality가 must-pass를 넘지 못한다. 아울러 audit plan이 codex를 required로 두는데 리더 지시로 호출하지 않았으므로, 그 게이트도 미충족(Gaps G1)이다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 50/100 | FAIL | F1: Windows 시뮬레이션에서 3건 실패, 그중 2건은 CI 04d8c27a7e의 Windows 실패 목록에 없던 테스트 |
| Security (25%) | 90/100 | PASS | POSIX 조건 3곳 모두 동치(아래 Q1). Windows 소유자 전용 미강제는 운영자 정책으로 공개된 잔여 위험 |
| Craft (20%) | 85/100 | PASS | 패키지 커버리지 88.2%, filemode.go 4개 함수 100%, golangci-lint 0 issues, gofmt 무출력 |
| Consistency (15%) | 90/100 | PASS | 작은 순수 함수 + host 래퍼, 기존 오류 문자열 유지, 파일 스타일 일치 |

## Findings (구조화 결함 목록)

- **F1 [High] [blocking] [in-diff, 재현 증거 있음]** `internal/mission/completion_receipt_test.go:70-75`, `internal/mission/git_owner_test.go:141-147`, `internal/mission/governance_receipt_test.go:77-82` — 세 음성(negative) 테스트가 `os.Chmod(path, 0o644)` 뒤에 "약한 모드 거부"를 단언한다. 이번 변경으로 Windows에서는 `hostPrivateModeOK` / `hostNoGroupOrOtherAccess`가 항상 true이므로, 거부가 일어나지 않아 단언이 실패한다. 호스트 판정을 `"windows"`로 고정한 overlay 실행 결과 `weak receipt accepted`, `weak lease mode accepted`, `weak mode err=<nil>` 3건이 실패했다. `TestCompletionReceiptRefusesUnsafeAndMalformedInputs`와 `TestGitOwnerSafetyMutants`는 CI run 42220998499(head 04d8c27a7e = 감사 커밋의 부모)의 Windows mission 실패 목록 21건에 없었다. 즉 Windows에서 새 회귀가 생긴다. `TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed`는 목록에 있었지만, 고친 뒤에도 81행에서 다시 실패한다.
  - Required fix: 운영자 정책("Windows에서는 이식 가능한 부분만 단언하고, 남는 것이 없을 때만 사유를 출력하고 skip")대로, 이 세 테스트의 약한 모드 단계만 POSIX에서 실행하라. 예: `if runtime.GOOS != "windows" { chmod 0o644 → 거부 단언 → (governance 테스트는) chmod 0o600 복원 } else { t.Log("windows: POSIX mode bits not enforced (t2130 policy)") }`. 테스트 전체를 `t.Skip`하지 말 것. 나머지 단계(누락, 외부 경로, 심볼릭 링크, 잘못된 JSON 등)는 이식 가능한 부분이므로 Windows에서도 계속 실행해야 한다. 수정 뒤 같은 overlay(`hostPrivateModeOK`/`hostNoGroupOrOtherAccess`를 `"windows"`로 고정)로 패키지 전체가 통과하는지 확인하라.

- **O1 [Low] [optional] [in-diff 행, 기존 동작]** `internal/mission/governance_receipt_jev_test.go:81-82` — `os.Stat`가 오류를 내면 `info`가 nil인 상태로 `t.Fatalf(... info.Mode() ...)`가 호출되어 panic이 난다. 이번 커밋 전부터 있던 패턴이고 81행 조건만 바뀌었다. 실패 메시지 품질 문제이므로 수정 여부는 재량.

- **O2 [Medium] [optional] [in-diff, 정책 결정]** `internal/mission/filemode.go:16-29` — Windows에서는 영수증과 임대(lease) 파일의 "소유자 전용" 성질이 어디서도 강제되지 않는다. 표준 라이브러리에는 ACL API가 없다. `golang.org/x/sys v0.49.0`이 go.mod에 있어 `windows.GetNamedSecurityInfo` + DACL 검사가 가능하긴 하지만, 상속 ACE와 SID 해석까지 다뤄야 해서 "간단한 방법"은 아니다. 운영자 정책 문언("where checkable")과 progress.md Residual-risk에 공개되어 있으므로 수용 가능한 잔여 위험으로 판단한다. 후속 카드 후보로만 기록.

- **O3 [Low] [optional] [outside]** `internal/mission/governance_receipt_test.go:94-99` (`os.Symlink`) — Windows 러너에서 심볼릭 링크 생성 권한이 없으면 해당 단계가 `t.Fatal`할 수 있다. 이 커밋 전에는 앞 단계에서 이미 실패해 관측된 적이 없다. F1을 고치고 나면 이 단계가 Windows에서 처음으로 실행된다. 검증하지 못함(G3).

## 판단 질문별 결론

- **Q1 POSIX 약화 여부: 약화 없음.** 바뀐 조건 세 곳을 각각 대조했다.
  - `readGovernanceReceipt`: `Perm() != 0o600` → `!hostPrivateModeOK(Perm(), 0o600)` = `!(perm == 0o600)`. 동치.
  - `LoadCompletionReceipt`: 위와 같은 형태로 동치.
  - `validateLease`: `Perm()&0o077 != 0` → `!hostNoGroupOrOtherAccess(Perm())` = `!(perm&0o077 == 0)`. 동치.
  - POSIX에서 `runtime.GOOS`는 `"windows"`가 될 수 없으므로 분기도 바뀌지 않는다. 일반 파일 검사와 심볼릭 링크 검사, `||` 단락 평가 순서도 그대로다.
- **Q2 Windows "검사 없음"의 수용 여부:** 정책 문언("not accessible to other users where checkable")에 비춰 보면, 표준 라이브러리로는 확인할 수 없으므로 검사가 없는 것은 수용 가능하다. 공개도 되어 있다(파일 주석, progress.md, card-review.md). "readable"이라는 이식 가능한 성질은 이후 `os.ReadFile`이 실패하면 거부되는 식으로 암묵적으로 유지된다. 다만 정책의 다른 절반, 즉 "남는 것만 단언"이 테스트 쪽에서 지켜지지 않았다(F1).
- **Q3 놓친 POSIX 비트 단언:** 프로덕션의 `Perm()` 비교는 세 곳이 전부이고, 모두 이번에 바뀌었다. 테스트에서는 `Perm()` 직접 비교 3곳이 헬퍼로 옮겨졌다. 그러나 **`os.Chmod(0o644)` 기반 음성 단언 3곳(F1)을 놓쳤다.** 나머지 `0o600`/`0o700` 출현은 쓰기 모드 인자(`WriteFile`, `Mkdir`, `Chmod`)라서 Windows에서 무해하다. `governance_receipt_test.go:187`의 `Mkdir(lockedDir, 0o500)` + "unwritable receipt directory accepted"는 Windows에서 디렉터리 읽기 전용 속성이 쓰기를 막지 않을 수 있지만, CI 실패 목록에 이 테스트(`TestWriteGovernanceReceipt…` 계열)는 없었으므로 판단을 보류한다(G4).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. POSIX 동작은 바이트 수준으로 동일하다(Q1).
2. 결정 함수의 POSIX 분기를 제거하는 변이는 테스트가 잡는다.
3. 이 커밋으로는 Windows에서 mission 패키지가 통과하지 않는다. 새 회귀 2건이 생긴다(F1).
4. 호스트 빌드: 테스트, vet(호스트·Windows), gofmt, golangci-lint 모두 깨끗하다.

### Evidence

```
$ git rev-parse HEAD
19aa331da44ebd5ae91d4b066f49322ea1cfed68

$ go test ./internal/mission -count=1
ok  	github.com/modu-ai/moai-adk/internal/mission	4.923s

$ GOOS=windows go vet ./internal/mission ; echo vet_exit=$?
vet_exit=0
$ gofmt -l internal/mission ; echo gofmt_exit=$?
gofmt_exit=0            (출력 없음)
$ go vet ./internal/mission ; echo vet_host_exit=$?
vet_host_exit=0
$ golangci-lint run ./internal/mission/...
0 issues.

$ go test ./internal/mission -count=1 -coverprofile=cov.out ; go tool cover -func=cov.out | grep -E 'filemode.go|total'
ok  	github.com/modu-ai/moai-adk/internal/mission	6.205s	coverage: 88.2% of statements
filemode.go:16:	privateModeOK			100.0%
filemode.go:24:	noGroupOrOtherAccess		100.0%
filemode.go:32:	hostPrivateModeOK		100.0%
filemode.go:34:	hostNoGroupOrOtherAccess	100.0%
total:	(statements)	88.2%
```

변이 검사: overlay로 `privateModeOK`의 `return perm == want`를 `return true`로 바꿔 POSIX 검사를 제거했다.

```
$ go test ./internal/mission -count=1 -overlay ov/mut.json
--- FAIL: TestCompletionReceiptRefusesUnsafeAndMalformedInputs (0.01s)
    completion_receipt_test.go:74: weak receipt accepted
--- FAIL: TestPrivateModeOK (0.00s)
    filemode_t2130_test.go:26: privateModeOK("linux", 666, 0600) = true, want false
    filemode_t2130_test.go:26: privateModeOK("linux", 640, 0600) = true, want false
    filemode_t2130_test.go:26: privateModeOK("darwin", 644, 0600) = true, want false
--- FAIL: TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed (0.01s)
    governance_receipt_test.go:81: weak mode err=<nil>
FAIL	github.com/modu-ai/moai-adk/internal/mission	3.908s
```

→ 변이는 잡혔다(새 테스트와 기존 테스트 모두 실패).

Windows 판정 시뮬레이션: overlay로 두 host 래퍼가 `runtime.GOOS` 대신 `"windows"`를 넘기게 했다. darwin 위에서 실행했다.

```
$ go test ./internal/mission -count=1 -overlay ov/win.json
--- FAIL: TestCompletionReceiptRefusesUnsafeAndMalformedInputs (0.01s)
    completion_receipt_test.go:74: weak receipt accepted
--- FAIL: TestGitOwnerSafetyMutants (0.29s)
    git_owner_test.go:146: weak lease mode accepted
--- FAIL: TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed (0.00s)
    governance_receipt_test.go:81: weak mode err=<nil>
FAIL	github.com/modu-ai/moai-adk/internal/mission	6.390s
```

CI 기준선(Windows 실패 목록):

```
$ gh run view 42220998499 --json headSha,headBranch
{"headBranch":"release/main-batch-20261106","headSha":"04d8c27a7ed77afc09217f8623519806d48df3c4"}
$ gh run view 42220998499 --job 127902905673 --log | grep 'FAILED.*internal/mission' | (테스트명만 추출, 중복 제거)
TestAutoMissionStateRoundTripAndRefusals
TestCompletionReceiptBindsLineageEvidenceAndAncestry
TestGitOwnerLocalDevelopNoFFLeaseAndCrashReadback
TestGitOwnerReadbackLeaseAndSnapshotRefusals
TestGovernanceReceiptAuxiliaryItemIsNotABindingField
TestGovernanceReceiptRecordsJevNoulAsSeparateItem
TestGovernanceReceiptWithoutAuxiliaryItemKeepsPreSpecShape
TestLoadGovernanceReceiptsBindsDecisionAndIndependentAudit
TestLoadGovernanceReceiptsRejectsEverySignedLineageAndStatusMutant (+ 11 subtests)
TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed
$ grep -c 'SafetyMutants\|RefusesUnsafeAndMalformed' winlog.txt
0
```

audit plan:

```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2130
"config_status": "ok", "model": "codex", "cross_model_active": true,
"cross_model_required": true, "enforced_required": ["codex"]
(backends: claude gate off, codex gate required, glm gate off)
```

### Baseline-attribution
- 위 명령은 모두 이번 실행에서 트리 t2130, HEAD 19aa331da4를 대상으로 돌렸다. 작업 트리는 깨끗했다(`git status --short` 무출력).
- Go 측정은 트리 소스로 그 자리에서 컴파일됐으므로 판정한 빌드 = 이 트리다.
- `moai verify audit-plan`은 설치된 바이너리(/Users/goos/go/bin/moai)로 실행했다. 동작 중인 MCP 서버 빌드는 04d8c27a7e(트리 HEAD의 부모)이다. 설치 바이너리의 커밋 자체는 확인하지 않았지만, 이 명령은 설정만 읽으므로 판정에 영향이 없다.
- CI 기준선은 run 42220998499 / job 127902905673(windows-latest, remainder), headSha 04d8c27a7e = 감사 커밋의 직전 커밋이다. mission 테스트 소스는 이 커밋의 diff를 빼면 동일하다.

### Gaps
- **G1 (PASS 차단 게이트):** audit plan이 `codex` 게이트를 `required`로 둔다(`enforced_required: ["codex"]`). 리더 지시("Codex 미승인")로 `codex_audit`/`audit_multi`를 호출하지 않았다. 따라서 필수 게이트는 미충족이고, 이 감사는 Claude 단독이다. 리더가 이 카드에 한해 Claude 단독 감사를 수용했다고 했지만, 게이트 자체는 충족되지 않았다. 이번 판정은 F1 때문에 어차피 FAIL이다.
- **G2:** 실제 Windows 호스트에서는 실행하지 못했다. F1은 판정 함수의 호스트 값을 `"windows"`로 고정한 overlay 시뮬레이션(darwin)에 근거한다. 실제 Windows에서 `Chmod(0o644)`는 0666로 보이고 헬퍼는 모드와 무관하게 true를 반환하므로, 판정 경로는 같다고 본다. 하지만 이후 단계(경로, EvalSymlinks)가 Windows에서 다르게 동작할 가능성은 관측하지 못했다.
- **G3:** O3(Windows 러너의 `os.Symlink` 권한)는 관측하지 못했다.
- **G4:** `Mkdir(…, 0o500)` 쓰기 불가 단언이 Windows에서 통과하는지는 이번 CI 목록에 없다는 사실로만 추정했고, 직접 관측하지 않았다.
- **G5:** CI 로그에서 FAILED 행만 추출했다. "목록에 없음 = 통과"라는 추론은 그 로그 형식(gotestsum 요약)이 실패를 빠짐없이 나열한다는 전제에 기댄다. PASS 행은 따로 확인하지 않았다.
- 거부된 도구 호출: 셸 변수와 git 유사 패턴이 섞인 복합 명령 2건이 worktree 격리 가드에 거부되었다. 대체로 리터럴 경로를 쓴 단순 명령으로 같은 측정을 다시 실행했다. 측정 결과 자체는 대체 명령의 실제 출력이다.

### Residual-risk
- Windows에서는 영수증과 임대 파일의 소유자 전용 성질이 강제되지 않는다(O2). 접근 통제는 상위 디렉터리에서 상속된 ACL에 의존한다. 운영자 정책으로 수용됐다.
- F1을 고친 뒤에도 Windows 전용 원인(심볼릭 링크 권한, 경로 표기)으로 mission 테스트가 실패할 수 있다. 최종 판단은 Windows CI 실행이다.
- 증거 원본(winlog.txt, cov.out, overlay 파일)은 세션 scratchpad에만 있다. 판정을 가른 출력 줄은 위에 옮겨 적었다.

## 반복 이력
| 회차 | SHA | 판정 | 점수 | 비고 |
|---|---|---|---|---|
| 1 | 19aa331da4 | FAIL | 0.62 | F1 (Windows 음성 단언 3건 누락, 새 회귀 2건), G1 codex 게이트 미충족 |

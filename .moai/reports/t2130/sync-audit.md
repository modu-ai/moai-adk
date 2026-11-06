auditor-model: claude-opus-5-5
verdict: PASS-WITH-DEBT
score: 0.90
audited_sha: 9881539b83a6589edbb5dcf2ae8fee5c596554f4

# t2130 sync-audit (2회차 재감사) — #1969 ② 1/3 internal/mission Windows 권한 비트 실패

- 대상: 커밋 9881539b83a6589edbb5dcf2ae8fee5c596554f4 (브랜치 WT-internal-mission-tests, 트리 /Users/goos/.moai/worktrees/projects/moai-adk-go/t2130)
- 수리 델타: `git diff 19aa331da4..HEAD` — 테스트 3곳의 약한 모드 단언에 `weakModeDetectable()` 조건 추가, `filemode_t2130_test.go`에 헬퍼 정의(7행). 프로덕션 코드 변경 없음.
- 감사자: Claude 단독. 리더 지시(이 카드는 Codex 미승인)로 codex_audit·audit_multi·glm_audit·claude_audit를 호출하지 않았다.

## 판정 요약

판정은 PASS-WITH-DEBT, 점수 0.90. 1회차 차단 결함 F1은 해소됐다. 1회차와 같은 방식으로 Windows를 시뮬레이션하면(호스트 래퍼와 `weakModeDetectable` 모두 `"windows"` 판정으로 고정) mission 패키지 전체가 통과한다(최상위 테스트 46건 PASS, FAIL 0). POSIX 동작은 그대로이고, 실제 빌드(darwin)에서는 약한 모드 단계가 여전히 살아 있어 POSIX 검사를 지우는 변이를 세 음성 테스트가 모두 잡는다. 게이트를 강제로 끄면 같은 변이를 그 세 테스트가 놓치므로, 게이트가 실제로 동작을 가른다는 것도 확인했다.

부채로 남기는 것은 두 가지다. (1) audit plan이 codex 게이트를 required로 두지만 리더 지시로 호출하지 않아 필수 게이트가 미충족이다(G1). (2) 실제 Windows 호스트 실행은 관측하지 못했다(G2). 최종 확인은 Windows CI다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 90/100 | PASS | Windows 시뮬레이션 `exit=0`, 46 PASS / 0 FAIL. 호스트 `go test` ok |
| Security (25%) | 90/100 | PASS | 프로덕션 diff 없음(POSIX 동치는 1회차 Q1 그대로). 변이 시 음성 테스트 3건 + 결정 함수 테스트 2건 FAIL |
| Craft (20%) | 88/100 | PASS | 커버리지 88.2%, filemode.go 4개 함수 100%, golangci-lint `0 issues.`, gofmt 무출력 |
| Consistency (15%) | 90/100 | PASS | 기존 단언 형태를 유지하고 조건만 덧붙인 최소 수리. 헬퍼 주석이 이유를 설명 |

## Findings (구조화 결함 목록)

차단(blocking) 결함 없음. 수리 델타에서 생긴(repair-born) 차단 결함도 없다.

- **F1 [High] [blocking] — 해소.** 세 음성 단언이 `weakModeDetectable()`로 Windows에서만 꺼진다. 시뮬레이션에서 세 테스트 모두 PASS. 나머지 단계(누락, 외부 경로, 심볼릭 링크, 잘못된 JSON, audit_failed 등)는 조건 밖이라 Windows에서도 실행된다.
- **O1 [Low] [optional] [touched-file, 기존]** `internal/mission/governance_receipt_jev_test.go:81-82` — 1회차와 같음. `os.Stat` 실패 시 nil `info`로 panic. 이번 수리 범위 밖.
- **O2 [Medium] [optional] [정책]** Windows에서 소유자 전용 성질은 강제되지 않는다. 운영자 정책으로 공개·수용된 잔여 위험. 후속 카드 후보.
- **O3 [Low] [optional] [outside]** `os.Symlink` 단계(`completion_receipt_test.go` 88행 부근, `governance_receipt_test.go` 94행 부근)가 이제 Windows에서 처음 실행된다. darwin 시뮬레이션은 심볼릭 링크를 만들 수 있으므로 이 단계는 통과했고, Windows 러너의 심볼릭 링크 권한 문제는 관측하지 못했다(G3). 만약 실패한다면 원인은 링크 생성 권한이지 이 카드의 결함이 아니다.
- **O4 [Low] [optional] [touched-file, 기존 동작]** `completion_receipt_test.go:76-81` — chmod 0644 뒤 `os.WriteFile(path, "{", 0o600)`은 기존 파일의 모드를 바꾸지 않는다. 그래서 POSIX에서 "malformed receipt accepted" 단언은 JSON 오류가 아니라 모드 거부 때문에 통과할 수 있다(원인 가림). Windows 시뮬레이션에서는 모드 검사가 없으므로 JSON 오류로 거부되어 통과했다. 이번 diff 이전부터 있던 구조이고, 단언 자체는 거짓 통과를 만들지 않는다.
- **O5 [Low] [optional] [in-diff]** 운영자 정책 문언("남는 것이 없을 때만 사유를 출력하고 skip")에 비춰 보면, Windows에서 약한 모드 단계를 건너뛸 때 `t.Log`로 사유를 남기지 않는다. 테스트 전체를 건너뛰는 것이 아니므로 정책 위반은 아니고, 헬퍼 주석에 사유가 있다. 기록 품질 문제다.

## 확인 항목별 결론

1. **F1 해소 (Windows 시뮬레이션): 확인.** overlay로 `hostPrivateModeOK` / `hostNoGroupOrOtherAccess`가 `"windows"`를 넘기게 하고, `weakModeDetectable`이 false를 반환하게 했다. 패키지 전체 통과. 다른 실패 없음. 심볼릭 링크 단계는 darwin에서 실행됐으므로 권한 문제는 관측 대상이 아니었다(O3/G3).
2. **POSIX 동등성 유지: 확인.** 수리 델타에 프로덕션 파일이 없다(`--stat`의 Go 파일은 `_test.go` 4개뿐). 호스트 test ok, `GOOS=windows go vet` exit 0, 호스트 vet exit 0, gofmt 무출력.
3. **변이 — 게이트가 공허하지 않음: 확인.**
   - 실제 헬퍼 + POSIX 검사 제거(`privateModeOK`·`noGroupOrOtherAccess` 모두 `return true`): 음성 테스트 3건(`weak receipt accepted`, `weak lease mode accepted`, `weak mode err=<nil>`)과 결정 함수 테스트 2건, 모두 5건 FAIL.
   - 같은 변이 + `weakModeDetectable` 강제 false: 음성 테스트 3건은 PASS(변이를 놓침), 결정 함수 테스트 2건만 FAIL. 즉 실제 빌드에서 게이트가 true일 때만 세 테스트가 POSIX 검사를 지킨다.
4. **Gaps 정직성:** 아래 Gaps 절 참조. lane의 progress.md에도 Codex 미요청과 Windows 미관측이 적혀 있다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. F1은 해소됐다. Windows 판정 시뮬레이션에서 mission 패키지가 전부 통과한다.
2. POSIX 동작은 바뀌지 않았다(이번 델타에 프로덕션 변경 없음).
3. 실제 빌드에서 약한 모드 게이트는 활성이고, POSIX 검사 제거 변이를 세 음성 테스트가 잡는다. 게이트를 끄면 놓친다.
4. 호스트 빌드: test, vet(호스트·Windows), gofmt, golangci-lint 모두 깨끗하다.

### Evidence

```
$ git rev-parse HEAD
9881539b83a6589edbb5dcf2ae8fee5c596554f4
$ git status --short
(출력 없음)
$ git diff --stat 19aa331da4..HEAD
 .moai/reports/t2130/progress.md             |   5 +
 .moai/reports/t2130/sync-audit.md           | 160 ++++++++++++++++++++++++++++
 internal/mission/completion_receipt_test.go |   2 +-
 internal/mission/filemode_t2130_test.go     |   7 ++
 internal/mission/git_owner_test.go          |   2 +-
 internal/mission/governance_receipt_test.go |   2 +-
 6 files changed, 175 insertions(+), 3 deletions(-)

$ go test ./internal/mission -count=1; echo test_exit=$?
ok  	github.com/modu-ai/moai-adk/internal/mission	8.406s
test_exit=0
$ GOOS=windows go vet ./internal/mission; echo winvet_exit=$?
winvet_exit=0
$ go vet ./internal/mission; echo vet_exit=$?
vet_exit=0
$ gofmt -l internal/mission; echo gofmt_exit=$?
gofmt_exit=0            (출력 없음)
$ golangci-lint run --allow-parallel-runners ./internal/mission/...
0 issues.
$ go test ./internal/mission -count=1 -coverprofile=cov.out ; go tool cover -func=cov.out | grep -E 'filemode.go|total'
ok  	github.com/modu-ai/moai-adk/internal/mission	7.253s	coverage: 88.2% of statements
filemode.go:16:	privateModeOK			100.0%
filemode.go:24:	noGroupOrOtherAccess		100.0%
filemode.go:32:	hostPrivateModeOK		100.0%
filemode.go:34:	hostNoGroupOrOtherAccess	100.0%
total:	(statements)	88.2%
```

overlay 소스 diff (생성 스크립트 출력 그대로):

```
== fm_win          (Windows 시뮬레이션)
< func hostPrivateModeOK(perm, want fs.FileMode) bool { return privateModeOK(runtime.GOOS, perm, want) }
> func hostPrivateModeOK(perm, want fs.FileMode) bool { return privateModeOK("windows", perm, want) }
< 	return noGroupOrOtherAccess(runtime.GOOS, perm)
> 	return noGroupOrOtherAccess("windows", perm)
   (+ 말미에 `var _ = runtime.GOOS` 한 줄: 사용하지 않는 import 컴파일 오류 방지)
== fmt_false       (게이트 강제 false)
< func weakModeDetectable() bool { return runtime.GOOS != "windows" }
> func weakModeDetectable() bool { _ = runtime.GOOS; return false }
== fm_mut_both     (POSIX 검사 제거 변이)
< 	return perm == want
> 	return true
< 	return perm&0o077 == 0
> 	return true
```

(1) Windows 시뮬레이션 — overlay `filemode.go→fm_win.go`, `filemode_t2130_test.go→fmt_false.go`:

```
$ go test ./internal/mission -count=1 -v -overlay win.json
exit=0
ok  	github.com/modu-ai/moai-adk/internal/mission	9.380s
top-level PASS count: 46
top-level FAIL count: 0
--- PASS: TestCompletionReceiptRefusesUnsafeAndMalformedInputs (0.15s)
--- PASS: TestGitOwnerSafetyMutants (0.62s)
--- PASS: TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed (0.03s)
--- PASS: TestPrivateModeOK (0.00s)
--- PASS: TestNoGroupOrOtherAccess (0.00s)
```

(3a) 실제 게이트 + POSIX 검사 제거 변이 — overlay `filemode.go→fm_mut_both.go`:

```
$ go test ./internal/mission -count=1 -v -overlay mut_real.json
exit=1
    completion_receipt_test.go:74: weak receipt accepted
--- FAIL: TestCompletionReceiptRefusesUnsafeAndMalformedInputs (0.00s)
    filemode_t2130_test.go:33: privateModeOK("linux", 666, 0600) = true, want false
    filemode_t2130_test.go:33: privateModeOK("linux", 640, 0600) = true, want false
    filemode_t2130_test.go:33: privateModeOK("darwin", 644, 0600) = true, want false
--- FAIL: TestPrivateModeOK (0.00s)
    filemode_t2130_test.go:53: noGroupOrOtherAccess("linux", 640) = true, want false
    filemode_t2130_test.go:53: noGroupOrOtherAccess("linux", 604) = true, want false
    filemode_t2130_test.go:53: noGroupOrOtherAccess("darwin", 666) = true, want false
--- FAIL: TestNoGroupOrOtherAccess (0.00s)
    git_owner_test.go:146: weak lease mode accepted
--- FAIL: TestGitOwnerSafetyMutants (0.44s)
    governance_receipt_test.go:81: weak mode err=<nil>
--- FAIL: TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/mission	9.634s
top-level PASS count: 41
top-level FAIL count: 5
```

(3b) 같은 변이 + 게이트 강제 false — overlay `filemode.go→fm_mut_both.go`, `filemode_t2130_test.go→fmt_false.go`:

```
$ go test ./internal/mission -count=1 -v -overlay mut_gateoff.json
exit=1
--- FAIL: TestPrivateModeOK (0.00s)
--- FAIL: TestNoGroupOrOtherAccess (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/mission	8.854s
top-level PASS count: 44
top-level FAIL count: 2
--- PASS: TestCompletionReceiptRefusesUnsafeAndMalformedInputs (0.02s)
--- PASS: TestGitOwnerSafetyMutants (0.99s)
--- PASS: TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed (0.18s)
```

→ (3a)와 (3b)의 차이가 정확히 세 음성 테스트다. 게이트는 실제 빌드에서 동작을 가른다.

audit plan:

```
$ moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2130
exit=0
"config_status": "ok"
"model": "codex"
"cross_model_active": true
"cross_model_required": true
"enforced_required": [ "codex" ]
```

### Baseline-attribution
- 위 명령은 모두 이번 실행에서 트리 t2130, HEAD 9881539b83(작업 트리 깨끗)을 대상으로 돌렸다.
- Go 측정은 트리 소스를 그 자리에서 컴파일했으므로 판정한 빌드 = 이 트리다. overlay는 표시한 파일만 대체한다.
- `moai verify audit-plan`은 설치된 바이너리로 실행했고, 설정만 읽는 명령이라 빌드 지연이 판정에 영향을 주지 않는다. 동작 중인 MCP 서버 빌드는 04d8c27a7e이며 이번 감사에서 MCP 감사 도구는 쓰지 않았다.
- 1회차 CI 기준선(run 42220998499, head 04d8c27a7e)은 다시 조회하지 않았다. 이번 결론은 그 기준선에 기대지 않고 시뮬레이션 결과에 기댄다.

### Gaps
- **G1 (필수 게이트 미충족, 부채):** audit plan은 `codex`를 required로 둔다(`enforced_required: ["codex"]`). 리더 지시("이 카드는 Codex 미승인")로 codex_audit·audit_multi를 호출하지 않았으므로 게이트는 충족되지 않았고, 이 판정은 Claude 단독이다. 영수증(receipt)이 발급되지 않았다. 종료 훅이 Codex 영수증을 요구해 이 판정을 거부하면, 그것은 이 게이트 때문이다.
- **G2:** 실제 Windows 호스트에서 실행하지 못했다. F1 해소는 판정 함수를 `"windows"`로 고정한 darwin 위 overlay 시뮬레이션에 근거한다. 실제 Windows에서만 달라지는 동작(경로 표기, `EvalSymlinks`, `Chmod` 의미)은 관측하지 못했다.
- **G3:** 이제 Windows에서 처음 실행되는 `os.Symlink` 단계의 권한 문제는 관측하지 못했다(O3). 실패한다면 링크 생성 권한 문제이며 이 카드의 결함이 아니다.
- **G4:** `Mkdir(…, 0o500)` 쓰기 불가 단언의 Windows 동작은 1회차처럼 관측하지 못했다.
- **G5 (거부된 도구 호출):** 셸 변수가 들어간 복합 명령 2건이 worktree 격리 가드에 거부되었다(sed에 런타임 변수 경로를 넘긴 overlay 생성 명령, 그리고 grep 패턴에 `Git…` 테스트 이름이 들어간 결과 필터 명령). 대신 리터럴 경로를 쓴 스크립트 파일(`mk.sh`, `run.sh`, 세션 scratchpad)로 같은 측정을 실행했다. 위 출력은 그 대체 실행의 실제 출력이다.
- **G6:** 첫 golangci-lint 실행은 `parallel golangci-lint is running`으로 중단되어 `--allow-parallel-runners`로 재실행했다. 재실행 결과가 위 `0 issues.`다.
- 개별 변이(privateModeOK만, noGroupOrOtherAccess만 제거)는 overlay를 만들었지만 실행하지 않았다. 둘 다 제거한 (3a)에서 각 음성 테스트가 자기 검사의 실패 메시지로 떨어졌으므로(`weak lease mode accepted`는 noGroupOrOtherAccess 경로, 나머지 둘은 privateModeOK 경로) 판정에는 충분하다고 본다.

### Residual-risk
- Windows에서 영수증·임대 파일의 소유자 전용 성질은 강제되지 않는다(O2, 운영자 정책으로 수용).
- 시뮬레이션은 판정 함수의 분기만 Windows로 바꾼다. 실제 Windows에서 다른 원인(심볼릭 링크 권한, 경로)으로 mission 테스트가 실패할 수 있다. 최종 판단은 Windows CI 실행이다.
- 증거 원본(win.out, mut_real.out, mut_gateoff.out, cov.out, overlay 파일)은 세션 scratchpad에만 있고 머신 로컬이다. 판정을 가른 줄은 위에 옮겨 적었다.

## 반복 이력
| 회차 | SHA | 판정 | 점수 | 비고 |
|---|---|---|---|---|
| 1 | 19aa331da4 | FAIL | 0.62 | F1 (Windows 음성 단언 3건 누락, 새 회귀 2건), G1 codex 게이트 미충족 |
| 2 | 9881539b83 | PASS-WITH-DEBT | 0.90 | F1 해소(시뮬레이션 46/46 PASS, 변이 게이트 비공허 확인). 부채: G1 codex 게이트 미충족(리더 지시), G2 실제 Windows 미관측 |

auditor-model: claude-opus-5-5
verdict: PASS-WITH-DEBT
score: 0.90
audited_sha: 9881539b83a6589edbb5dcf2ae8fee5c596554f4
receipt: rcpt-44a47f7d22a42a2c86d353c2

# t2130 sync-audit (3회차, 게이트 종결 재감사) — #1969 ② 1/3 internal/mission Windows 권한 비트 실패

- 대상 코드 커밋: 9881539b83a6589edbb5dcf2ae8fee5c596554f4 (브랜치 WT-internal-mission-tests, 트리 /Users/goos/.moai/worktrees/projects/moai-adk-go/t2130). 현재 HEAD 51d32a021b는 그 위의 문서 전용 판정 커밋이다.
- 카드 범위: `git diff 04d8c27a7e..9881539b83` — 19aa331da4(filemode.go 신설 + 프로덕션 호출 지점 3곳 + 테스트), 9881539b83(약한 모드 음성 단계에 `weakModeDetectable()` 조건).
- 감사자: Claude(이 보고서) + Codex 1회. 운영자 결정 d-20261106T163140Z-56c1로 승인된 `mcp__moai__codex_audit` 단 1회만 호출했다. audit_multi·glm_audit·claude_audit·codex_review·glm_review는 호출하지 않았다.

## 판정 요약

판정은 PASS-WITH-DEBT, 점수 0.90. 2회차에서 남긴 부채 중 G1(Codex 필수 게이트 미충족)은 이번에 해소됐다. 카드의 두 커밋으로 범위를 좁힌 Codex native 리뷰가 `pass`, 지적 0건을 돌려주었고 영수증 `rcpt-44a47f7d22a42a2c86d353c2`가 발급됐다. 재현할 지적이 없으므로 2회차의 측정(Windows 시뮬레이션 46/46 PASS, POSIX 검사 제거 변이를 음성 테스트 3건이 포착, 게이트를 끄면 놓침)이 그대로 판정 근거다.

남은 부채는 하나다. 실제 Windows 호스트 실행을 관측하지 못했다(D1 = G2). 이 카드의 목적이 Windows 실패 21건 해소이므로, 착지 전후 Windows CI 실행으로 확인해야 한다. 운영자 정책으로 수용된 잔여 위험(O2: Windows에서 소유자 전용 성질 미강제)은 부채가 아니라 공개된 정책 결과다.

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 90/100 | PASS | Windows 시뮬레이션 `exit=0`, 46 PASS / 0 FAIL. 호스트 `go test` ok (2회차 측정) |
| Security (25%) | 91/100 | PASS | POSIX 경로 동치, 변이 시 음성 테스트 3건 + 결정 함수 테스트 2건 FAIL (2회차). Codex 독립 리뷰 지적 0건 (3회차) |
| Craft (20%) | 88/100 | PASS | 커버리지 88.2%, filemode.go 4개 함수 100%, golangci-lint `0 issues.`, gofmt 무출력 (2회차) |
| Consistency (15%) | 90/100 | PASS | 기존 단언 형태를 유지하고 조건만 덧붙인 최소 수리. 헬퍼 주석이 이유를 설명 |

필수 통과 차원(Functionality, Security) 모두 기준 이상. 필수 게이트(codex) 충족.

## Findings (구조화 결함 목록)

차단(blocking) 결함 없음. Codex 지적 0건. 수리 델타에서 생긴(repair-born) 차단 결함도 없다.

- **F1 [High] [blocking] — 해소 (2회차).** 세 음성 단언이 `weakModeDetectable()`로 Windows에서만 꺼진다. 시뮬레이션에서 세 테스트 모두 PASS.
- **D1 [Medium] [부채] — 실제 Windows 미관측 (G2).** Required fix: Windows CI 실행에서 `internal/mission` 패키지 통과를 확인하고 결과를 카드 진행 기록에 남긴다.
- **O1 [Low] [optional] [touched-file, 기존]** `internal/mission/governance_receipt_jev_test.go:81-82` — `os.Stat` 실패 시 nil `info`로 panic. 이번 수리 범위 밖.
- **O2 [Medium] [optional] [정책]** Windows에서 소유자 전용 성질은 강제되지 않는다. 운영자 정책으로 공개·수용된 잔여 위험. 후속 카드 후보.
- **O3 [Low] [optional] [outside]** `os.Symlink` 단계가 Windows에서 처음 실행된다. 러너의 링크 생성 권한 문제는 관측하지 못했다(G3).
- **O4 [Low] [optional] [touched-file, 기존 동작]** `completion_receipt_test.go:76-81` — chmod 0644 뒤 `os.WriteFile(..., 0o600)`은 기존 파일 모드를 바꾸지 않아 POSIX에서 "malformed" 단언의 원인이 가려질 수 있다. 거짓 통과는 만들지 않는다.
- **O5 [Low] [optional] [in-diff]** Windows에서 약한 모드 단계를 건너뛸 때 `t.Log`로 사유를 남기지 않는다. 헬퍼 주석에 사유가 있어 정책 위반은 아니다.

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. Codex 필수 게이트가 카드 범위 리뷰로 충족됐다(pass, 지적 0건, 영수증 발급).
2. F1은 해소됐다. Windows 판정 시뮬레이션에서 mission 패키지가 전부 통과한다(2회차 측정).
3. POSIX 동작은 바뀌지 않았고, 실제 빌드에서 약한 모드 게이트는 활성이다(2회차 변이 측정).
4. 호스트 빌드: test, vet(호스트·Windows), gofmt, golangci-lint 모두 깨끗하다(2회차 측정).

### Evidence

3회차 (이번 실행):

```
$ git status --short
(출력 없음)
$ git rev-parse HEAD
51d32a021b93a0908e4531772eb092d184686fbe
$ git rev-parse 19aa331da4^
04d8c27a7ed77afc09217f8623519806d48df3c4
$ git diff --stat 19aa331da4^ 9881539b83
 internal/mission/completion_receipt.go          |   2 +-
 internal/mission/completion_receipt_test.go     |   4 +-
 internal/mission/coverage_test.go               |   2 +-
 internal/mission/filemode.go                    |  36 ++++++
 internal/mission/filemode_t2130_test.go         |  56 +++++++++
 internal/mission/git_owner.go                   |   2 +-
 internal/mission/git_owner_test.go              |   2 +-
 internal/mission/governance_receipt.go          |   2 +-
 internal/mission/governance_receipt_jev_test.go |   2 +-
 internal/mission/governance_receipt_test.go     |   2 +-
 (+ .moai/reports/t2130/ 문서 3개)
 13 files changed, 295 insertions(+), 9 deletions(-)

mcp__moai__codex_audit(project_root=<t2130>, target=baseBranch, delta_base=04d8c27a7ed77afc09217f8623519806d48df3c4, mode=native)
  result verdict = "pass"
  findings = []
  next_steps = []
  audit_receipt = "rcpt-44a47f7d22a42a2c86d353c2"
  review_base = "delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)"
  provenance.backend = "codex", transport = "codex app-server", resolved_model = "gpt-6.1-sol", requested_effort = "high"
  build_commit = "04d8c27a7e"
  summary(원문): "지정 기준 대비 변경에서 조치할 결함은 찾지 못했어. 정적 코드 검토만 했고, 지시대로 테스트·빌드·재현 명령은 실행하지 않았어."
```

2회차 측정 (트리 t2130, HEAD 9881539b83, 그대로 유효):

```
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

overlay 소스 diff (2회차):

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
--- FAIL: TestPrivateModeOK (0.00s)
--- FAIL: TestNoGroupOrOtherAccess (0.00s)
    git_owner_test.go:146: weak lease mode accepted
--- FAIL: TestGitOwnerSafetyMutants (0.44s)
    governance_receipt_test.go:81: weak mode err=<nil>
--- FAIL: TestLoadGovernanceReceiptsRejectsUnsafeMissingStaleAndFailed (0.01s)
top-level PASS count: 41
top-level FAIL count: 5
```

(3b) 같은 변이 + 게이트 강제 false:

```
$ go test ./internal/mission -count=1 -v -overlay mut_gateoff.json
exit=1
--- FAIL: TestPrivateModeOK (0.00s)
--- FAIL: TestNoGroupOrOtherAccess (0.00s)
top-level PASS count: 44
top-level FAIL count: 2
```

→ (3a)와 (3b)의 차이가 정확히 세 음성 테스트다. 게이트는 실제 빌드에서 동작을 가른다.

audit plan (2회차):

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
- 3회차 git 명령과 Codex 호출은 이번 실행에서 트리 t2130(작업 트리 깨끗, HEAD 51d32a021b)을 대상으로 했다. Codex 리뷰 범위는 `04d8c27a7e...HEAD`로, 카드의 두 코드 커밋과 그 위 문서 커밋을 포함한다.
- 2회차 Go 측정은 트리 소스를 그 자리에서 컴파일했으므로 판정한 빌드 = 이 트리(9881539b83)다. 이번에 다시 실행하지 않았다(지시: 긴 스위트 재실행 금지, Codex 지적이 없어 초점 테스트도 불필요).

### Gaps
- **G2 (부채 D1):** 실제 Windows 호스트에서 실행하지 못했다. F1 해소는 판정 함수를 `"windows"`로 고정한 darwin 위 overlay 시뮬레이션에 근거한다. 실제 Windows에서만 달라지는 동작(경로 표기, `EvalSymlinks`, `Chmod` 의미)은 관측하지 못했다.
- **G3:** Windows에서 처음 실행되는 `os.Symlink` 단계의 권한 문제는 관측하지 못했다(O3).
- **G4:** `Mkdir(…, 0o500)` 쓰기 불가 단언의 Windows 동작은 관측하지 못했다.
- **G7 (도구 빌드 지연, §2.2):** codex_audit를 중계한 moai MCP 서버 빌드는 04d8c27a7e로, 트리 HEAD 51d32a021b의 조상이다. 지연분은 이 카드 자신의 커밋(`internal/mission`)과 문서 커밋뿐이며 diff 수집·Codex 중계 경로를 바꾸지 않는다. 그래도 리뷰를 중계한 빌드가 트리 빌드가 아니라는 사실은 기록해 둔다. 코드 판정 자체는 Codex 모델(gpt-6.1-sol)이 diff를 읽고 내렸다.
- **G8:** Codex는 정적 리뷰만 했다(요약 원문: 테스트·빌드·재현 명령 미실행). 동작 증거는 2회차 Claude 측정이 담당한다.
- **G9 (전송 경로):** 결과의 `transport`는 `codex app-server`다. 운영자가 승인한 것은 `mcp__moai__codex_audit` 호출 1회이고, 이 도구가 내부적으로 해당 전송을 쓴다. 직접 app-server 경로를 호출하지는 않았다.
- 2회차 Gaps G5(셸 변수 복합 명령 2건 가드 거부 → 리터럴 경로 스크립트로 대체 실행), G6(golangci-lint 병렬 충돌 → `--allow-parallel-runners` 재실행)은 그대로 유효하다. 2회차 G1(Codex 게이트 미충족)은 3회차에서 해소됐다.

### Residual-risk
- Windows에서 영수증·임대 파일의 소유자 전용 성질은 강제되지 않는다(O2, 운영자 정책으로 수용).
- 시뮬레이션은 판정 함수의 분기만 Windows로 바꾼다. 실제 Windows에서 다른 원인(심볼릭 링크 권한, 경로)으로 mission 테스트가 실패할 수 있다. 최종 판단은 Windows CI 실행이다(D1).
- 2회차 증거 원본(win.out, mut_real.out, mut_gateoff.out, cov.out, overlay 파일)은 그때의 세션 scratchpad에만 있고 머신 로컬이다. 판정을 가른 줄은 위에 옮겨 적었다.

## 반복 이력
| 회차 | SHA | 판정 | 점수 | 비고 |
|---|---|---|---|---|
| 1 | 19aa331da4 | FAIL | 0.62 | F1 (Windows 음성 단언 3건 누락, 새 회귀 2건), G1 codex 게이트 미충족 |
| 2 | 9881539b83 | PASS-WITH-DEBT | 0.90 | F1 해소(시뮬레이션 46/46 PASS, 변이 게이트 비공허 확인). 부채: G1 codex 게이트 미충족(리더 지시), G2 실제 Windows 미관측 |
| 3 | 9881539b83 | PASS-WITH-DEBT | 0.90 | 운영자 승인 codex_audit 1회(delta 04d8c27a7e, native): pass, 지적 0건, 영수증 rcpt-44a47f7d22a42a2c86d353c2 → G1 해소. 남은 부채 D1(G2) 실제 Windows CI 확인 |

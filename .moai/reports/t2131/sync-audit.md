auditor-model: claude-opus-5-5
verdict: PASS
score: 0.88
audited_sha: 2661e35534c2a5868d983f91f7d4482c39cb38d3
audit_receipt: rcpt-1f9615d82d2136a86d0b74ab

## 게이트 마감 재감사 보고서 — 카드 t2131 (#1969 ② 2/3, internal/cli Windows 권한 비트 실패)

- 대상 코드 커밋: `2661e35534c2a5868d983f91f7d4482c39cb38d3`
- 현재 HEAD: `940d181cdc9650489c2d6c90c7ccc9fc07083298`. 대상 커밋 위에 문서 커밋이 하나 있다. `git diff --stat 2661e35534 HEAD`의 출력은 `.moai/reports/t2131/sync-audit.md | 183 ++++++++++++++++++++------------------`, `1 file changed`이다. 코드는 바뀌지 않았다.
- 감사를 시작할 때 `git status --short`의 출력은 비어 있었다.
- 브랜치: `WT-internal-cli-tests`, 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2131`
- SPEC과 평가 프로필이 없어 내장 기본 프로필을 썼다. 필수 통과 차원은 Functionality와 Security다.
- 이번 회차의 새 작업은 운영자가 승인한 `mcp__moai__codex_audit` 1회 호출이다(결정 `d-20261106T163140Z-56c1`). 긴 테스트 묶음은 다시 돌리지 않았다. 2회차 실측은 코드가 바뀌지 않았으므로 그대로 유효하며, 아래 Evidence에 원문대로 옮겼다.

### 판정 요약

- **PASS, 0.88.** 2회차에서 FAIL로 기록한 유일한 사유는 required 게이트 `codex`가 미충족이었다는 것이다. 이번 회차에서 그 게이트가 충족됐다. Codex native 리뷰는 카드 변경분(`04d8c27a7e...HEAD`)을 대상으로 `verdict: pass`를 내고 지적 사항 0건을 돌려줬다. 영수증 id는 `rcpt-1f9615d82d2136a86d0b74ab`이다.
- 차단 결함은 0건이다. 남은 지적 N1~N3은 모두 optional이다. 형제 카드 t2130에 대한 의존성(goal_mission 4행)은 공개된 의존성이며, 이 카드의 결함이 아니다. 따라서 PASS-WITH-DEBT가 아니라 PASS로 판정한다. 이 의존성은 Residual-risk에 남겼다.

### 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 92/100 | PASS | Windows를 모사한 overlay로 지정 묶음을 돌린 결과는 `--- PASS` 46개, `--- FAIL` 0개, `--- SKIP` 0개, `ok … 11.034s`, `exit=0`이다(2회차, 같은 코드). 양성 대조는 `goal_mission_test.go:301: weak receipt permissions accepted`로 실패했다. 호스트 POSIX에서도 8개 테스트가 통과했다. Codex 리뷰는 지적 사항 0건이다. |
| Security (25%) | 85/100 | PASS | 운영 코드는 `goal.go` 1줄과 `filemode_portable.go`만 바뀌었다. POSIX에서는 0644 거부를 그대로 강제한다. Windows에서 모드를 강제하지 않는 것은 운영자 정책(2026-11-06)이다. 심링크, 일반 파일, 범위, 내용 검사는 Windows에서도 계속 돈다. Codex 리뷰에서 보안 지적은 없었다. |
| Craft (20%) | 85/100 | PASS | `go vet` exit=0, `GOOS=windows go vet` exit=0이고, 변경된 Go 파일 9개에 대한 `gofmt -l` 출력은 비어 있다(2회차). 감점 사유는 N1(optional)이다. |
| Consistency (15%) | 90/100 | PASS | 새 헬퍼는 `permMatchesOnHost`와 같은 파일에 있고 같은 기준(`runtime.GOOS`)을 쓴다. 기존 패턴(`clifix_concurrency_m3_test.go:70`)과도, 형제 카드 t2130의 정책과도 맞는다. |

조화평균은 4 / (1/92 + 1/85 + 1/85 + 1/90) ≈ 87.9이며, 0.88로 기록한다. 필수 통과 차원인 Functionality와 Security는 둘 다 PASS다.

### Codex 감사 결과 (이번 회차, 1회 호출)

호출: `mcp__moai__codex_audit(project_root=/Users/goos/.moai/worktrees/projects/moai-adk-go/t2131, target=baseBranch, delta_base=04d8c27a7ed77afc09217f8623519806d48df3c4, mode=native)`

delta 기준은 `fe1ab887ae`의 부모다. `git rev-parse fe1ab887ae^`의 출력은 `04d8c27a7ed77afc09217f8623519806d48df3c4`이고, `git merge-base --is-ancestor`의 결과는 `ancestor`다. 이 범위에는 카드 커밋 `fe1ab887ae`와 `2661e35534`가 들어가고, 문서 커밋 `940d181cdc`도 함께 들어간다.

결과 원문(요약 필드):
```
"verdict":"pass","summary":"Verdict: pass","findings":[],"next_steps":[]
"transport":"codex app-server","resolved_model":"gpt-6.1-sol","requested_effort":"high","tool_surface":"codex_audit","error_code":""
"audit_receipt":"rcpt-1f9615d82d2136a86d0b74ab"
"review_base":"delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)"
"build_commit":"04d8c27a7e"
```

stale_server 거부는 없었다. 재시도도 하지 않았다. Codex가 지적한 사항이 없어 재현할 대상도 없다.

### 결함 처리 결과 (누적)

| ID | 처음 나온 회차 | 현재 상태 | 근거 |
|---|---|---|---|
| F1 [High, blocking] | 1 | **해소** | 수정한 overlay에서는 PASS다. 게이트만 true로 바꾼 대조에서는 301행이 FAIL이다. POSIX에서는 거부 단계가 그대로 실행돼 통과한다. |
| F2 [Medium, blocking] | 1 | **해소** | progress.md의 정정 항목이 코드와 일치한다. |
| F3 [Low, optional] | 1 | **기록됨** | t2130 의존성이 "inference, not run"으로 표시돼 있다. |
| F4 [Low, optional, outside] | 1 | **기록됨** | 범위 밖이며 판정에 영향이 없다. |
| G1 codex 게이트 미충족 | 1·2 | **충족** | 영수증 `rcpt-1f9615d82d2136a86d0b74ab`, Codex 판정 pass, 지적 0건 |

### 발견 사항 (미해결, 모두 optional)

- **N1 [Low] [optional] [in-diff]** `.moai/reports/t2131/progress.md`의 Evidence 절에는 수리 후 실측이 없다. 이 보고서 Evidence 1~3을 인용하면 된다. 거짓 주장은 아니다.
- **N2 [Low] [optional] [in-diff]** `internal/cli/goal_mission_test.go:300`은 Windows에서 이 단계를 건너뛸 때 `t.Log`로 사유를 남기지 않는다. 테스트의 나머지 단계는 계속 실행되므로 정책 위반은 아니다.
- **N3 [Info] [optional] [outside]** `internal/cli/migrate_cg_test.go:221`에도 약한 모드를 쓰는 부정 단계가 있다. 다만 이 단계가 시험하는 운영 코드는 이 카드에서 바뀌지 않았으므로, 이 카드 때문에 새로 생긴 실패는 아니다.

### 반복 이력

| 회차 | 대상 커밋 | 판정 | 점수 | 차단 사유 |
|---|---|---|---|---|
| 1 | `fe1ab887ae` | FAIL | 0.73 | F1(Windows 0644 단계 새 실패), F2(progress 거짓 주장), codex 게이트 미충족 |
| 2 | `2661e35534` | FAIL (게이트) / 코드 판정 PASS 수준 | 0.88 | codex required 게이트 미충족만 남음. 코드 차단 결함 0건. 수리에서 생긴 차단 결함 없음 |
| 3 (이번, 게이트 마감) | `2661e35534` | PASS | 0.88 | 없음. codex 게이트 충족(`rcpt-1f9615d82d2136a86d0b74ab`, pass, 지적 0건) |

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. required 게이트 `codex`가 충족됐다. Codex native 리뷰가 카드 변경분에 pass를 냈고 지적은 0건이다.
2. F1이 해소됐다. Windows를 모사한 지정 묶음이 전부 통과한다. overlay가 실제로 적용됐다는 것은 양성 대조로 확인했다.
3. POSIX 동작은 바뀌지 않았다.
4. vet(호스트와 windows)과 gofmt 결과가 깨끗하다.
5. 이 카드 때문에 Windows에서 새로 실패하는 약한 모드 단계는 더 없다.

### Evidence
0. (이번 회차) `git status --short`의 출력은 비어 있다. `git rev-parse HEAD`의 출력은 `940d181cdc9650489c2d6c90c7ccc9fc07083298`이다. `git diff --stat 2661e35534 HEAD`는 `.moai/reports/t2131/sync-audit.md` 1개 파일만 보여 준다. Codex 결과 원문은 위 "Codex 감사 결과" 절에 있다.
1. (2회차, 같은 코드) Windows 모사 overlay를 썼다. `hostNoGroupOrOtherAccess`가 `noGroupOrOtherAccessOn("windows", perm)`을 부르게 하고, `permMatchesOnHost`는 `"windows" == "windows" || perm == want`로, `weakModeRefusalExpected`는 `"windows" != "windows"`로 바꿨다.
   `go test -overlay=<scratchpad>/t2131r2/overlay.json ./internal/cli -run 'TestAutoMission|TestMission|TestGoal|TestNoGroupOrOtherAccessOn|TestCleanHome_RetentionFromHomeTier|TestGlmKeyFlagSavesKeySameStorage|TestJevKeyFlagSavesCredential|TestOutboxLogAppendOnly0600|TestMemoryFold_VerbatimFiling' -count=1 -v`
   ```
   46            (grep -c '^--- PASS')
   0             (grep -c '^--- FAIL')
   0             (grep -c -- '--- SKIP')
   --- PASS: TestAutoMissionTestsReceiptTrustBoundary (0.00s)
   PASS
   ok  	github.com/modu-ai/moai-adk/internal/cli	11.034s
   exit=0
   ```
   같은 overlay로 다음 명령도 돌렸다. `go test -overlay=… ./internal/cli -run 'TestGTDAutonomyEndToEndProductionCLIWithGovernanceReceipts|TestNewAutoMissionCommandTreatsTextAsData' -count=1 -v`
   ```
   --- PASS: TestNewAutoMissionCommandTreatsTextAsData (0.00s)
   --- PASS: TestGTDAutonomyEndToEndProductionCLIWithGovernanceReceipts (2.46s)
   ok  	github.com/modu-ai/moai-adk/internal/cli	3.673s
   ```
2. (2회차) 양성 대조. `weakModeRefusalExpected`만 `return true`로 바꿨다. 결과는 `exit=1`이다.
   ```
       goal_mission_test.go:301: weak receipt permissions accepted
   --- FAIL: TestAutoMissionTestsReceiptTrustBoundary (0.02s)
   FAIL	github.com/modu-ai/moai-adk/internal/cli	1.565s
   ```
3. (2회차) 호스트 darwin/arm64에서 8개 테스트를 지정해 돌린 결과는 `exit=0`이다.
   ```
   --- PASS: TestNoGroupOrOtherAccessOn (0.00s)
   --- PASS: TestAutoMissionTestsReceiptTrustBoundary (0.00s)
   PASS
   ok  	github.com/modu-ai/moai-adk/internal/cli	2.314s
   ```
4. (2회차) `go vet ./internal/cli`는 `vet exit=0`, `GOOS=windows go vet ./internal/cli`는 `winvet exit=0`이다. 변경된 Go 파일 9개에 대한 `gofmt -l` 출력은 비어 있다.
5. (2회차) 운영 코드에서 판정을 쓰는 곳은 `internal/cli/goal.go:597` 하나뿐이다. 테스트에서 호출하는 곳은 `goal_mission_test.go:292,296,300,306,313,316`이다.
6. (2회차) `moai verify audit-plan --project-root /Users/goos/.moai/worktrees/projects/moai-adk-go/t2131`의 출력은 `"config_status": "ok"`, `"model": "codex"`, `"cross_model_required": true`, `"enforced_required": ["codex"]`이다.

### Baseline-attribution
Evidence 0은 이번 실행에서 이 트리(HEAD `940d181cdc`, clean)를 대상으로 얻었다. Evidence 1~6은 2회차에 HEAD `2661e35534`를 대상으로 얻었다. 그 뒤의 차이는 판정 파일 1개뿐이므로, 대상 코드는 같다. 도구 체인은 `go version go1.27.3 darwin/arm64`이고, `go test`와 `go vet`는 트리 소스를 직접 빌드한다. Codex 리뷰는 서버가 `04d8c27a7e...HEAD` 범위로 리뷰했다고 보고했다.

### Gaps
- **Codex 판정은 Codex의 보고다.** 리뷰가 실제로 읽은 diff는 도구 결과에 담겨 오지 않는다. 범위는 `review_base` 필드로만 확인할 수 있다.
- **운반 경로:** 도구가 보고한 transport는 `codex app-server`다. 이것은 승인된 `mcp__moai__codex_audit` 도구가 내부에서 쓰는 경로이고, 이 감사가 app-server를 직접 호출하지는 않았다.
- **MCP 서버 빌드 지연:** `build_lag`에 따르면 서버 바이너리는 `04d8c27a7`에서 빌드됐고, 이 커밋은 트리 HEAD `940d181cd`의 조상이다. Codex 리뷰는 git diff를 읽으므로 리뷰 내용에는 영향이 없다고 판단한다. 다만 영수증 발급 로직을 포함한 서버 쪽 코드는 HEAD보다 오래된 것이다.
- 실제 Windows에서는 실행하지 않았다. overlay는 `internal/cli`의 `runtime.GOOS` 분기만 모사한다.
- t2130 의존성은 이 모사로 측정되지 않았다. goal_mission의 4행(123·192·250·489)이 모사 실행에서 통과한 것은 `internal/mission`이 darwin에서 돌기 때문이다.
- 패키지 전체 스위트와 커버리지는 측정하지 않았다. 지시에 따라 긴 테스트 묶음은 다시 돌리지 않았다.
- 2회차에서 워크트리 격리 가드가 명령 3건을 거부했다. 다른 형태로 다시 돌려 얻은 결과가 Evidence 1~4다.

### Residual-risk
- Windows에서는 tests 영수증이 소유자 전용인지 모드로 강제하지 않는다. 이것은 운영자 정책이며 t2130도 같다.
- t2130이 착지하기 전에는 Windows CI에서 goal_mission의 4행이 계속 빨간불일 수 있다.
- `weakModeRefusalExpected`는 운영 판정 정책을 테스트 쪽에 복사한 것이다. 운영 판정이 나중에 바뀌면 이 차이는 Windows CI에서만 드러난다.

## 권고
1. 차단 사항은 없다. 카드는 감사 단계를 떠날 수 있다.
2. (선택) N1: progress.md의 Evidence 절에 수리 후 실측을 인용한다.
3. t2130이 착지한 뒤 Windows CI에서 goal_mission의 4행이 초록인지 확인한다.

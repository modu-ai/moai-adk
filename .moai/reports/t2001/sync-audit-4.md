auditor-model: claude-opus-5-5

verdict: PASS-WITH-DEBT
audited_sha: b7c4e23b8a1c55e51bcd38531031f6421bde2efb

# t2001 sync 감사 4회차 (D1 수리 확인) — PR #1955 (WT-cc-bash-monitor)

- 카드: t2001 (Class B, SPEC 없음) · 승인 기록: d-20261103T013038Z-0563 (D1 한 줄 lint 수리 + 재푸시 + 확인 감사 1회, FAIL이면 카드 정지)
- 고정 대상: `git rev-parse HEAD` → `b7c4e23b8a1c55e51bcd38531031f6421bde2efb`. 감사 시작 때와 이 파일을 쓰기 직전에 두 번 확인했고 같았다. 작업 트리는 두 번 모두 깨끗했다(`git status --short` 빈 출력).
- 원격 PR head: `gh pr view 1955 --json headRefOid` → `b7c4e23b8a1c55e51bcd38531031f6421bde2efb` (draft, OPEN, MERGEABLE). 시작과 끝 두 번 읽었고 고정 SHA와 같다.
- 종합: **PASS-WITH-DEBT**. D1은 닫혔다. 바뀐 소스는 정확히 한 줄이고 De Morgan 법칙으로 동작이 같다. 로컬 `golangci-lint`는 `0 issues.`, 새 head의 CI `Lint`는 `pass`, 필수 체크 세 개 모두 `pass`, 필수 Codex gate도 `pass`다. 새 blocking 결함은 없다. 다만 3회차의 optional 항목(D2·D3·D4·D5·N4·F7·F8)은 이번 범위 밖이라 그대로 열려 있고, 새 Info 항목 D6(Windows 전용 잠금 테스트를 CI가 돌리지 않음)을 추가한다. 이 열린 항목들 때문에 PASS가 아니라 PASS-WITH-DEBT로 둔다.

## 요약 판정

| 항목 | 상태 |
|---|---|
| D1 (staticcheck QF1001, CI Lint 실패) | **CLOSED** — 한 줄 수리, 로컬 `0 issues.`, CI `Lint` pass @ b7c4e23b8a (run 41560649278) |
| D2 (stamp 재독 TOCTOU) | **OPEN** (Low, optional, 코드 변경 없음) |
| D3 (장치 하나만 지운 변이체가 테스트를 통과) | **OPEN** (Low, optional, 테스트 변경 없음) |
| D4 (문서의 본문 "바이트 단위" 과장) | **OPEN** (Low, optional, 문서 변경 없음) |
| D5 (CHANGELOG 이중 "or") | **OPEN** (Info, optional) |
| D6 | **NEW** (Info, optional) — 아래 결함 목록 |
| N1 (링크·ignore로 stamp 안정성 우회) | **CLOSED** (유지) — 회귀 테스트 두 개 이번 head에서 다시 PASS |
| N2 · N3 | **CLOSED** (유지, 이번 diff에서 바뀌지 않음) |
| N4 (임계값을 PR이 바꾸는 경우) | **OPEN** (Info, optional) |
| F1 (Windows 잠금 영역) | **CLOSED** (유지) — `GOOS=windows go vet` exit 0. 실행 기록 부재는 D6·Gaps |
| F2 (stamp 되감기로 inherited) | **CLOSED** (유지) — `TestGraphCheckPolicy_RewoundStampBlocks` PASS |
| F3 · F4 · F5 · F6 | **CLOSED** (유지, 이번 diff에서 바뀌지 않음) |
| F7 (inbox 이중 rotation 경합) | **OPEN** (Low, optional) |
| F8 (base=`HEAD^1` 전제) | **OPEN** (Info, optional) |

결함 수(열린 것): Critical 0 · High 0 · Medium 0 · Low 4(D2, D3, D4, F7) · Info 4(D5, D6, N4, F8). blocking 0.

## 결함 목록 (structured defect-list)

- **D1** [Medium] [blocking] `internal/graph/check.go:892` — **CLOSED**. 조건식이 `if err != nil || (!d.IsDir() && !d.Type().IsRegular()) {`로 바뀌었다(Claim A). `!(A || B)` ≡ `!A && !B`이므로 모든 입력에서 같은 값이다. 단락 평가 순서도 같다. `err != nil`이 먼저 평가되므로 `d`가 nil일 수 있는 오류 경로에서 `d.IsDir()`를 부르지 않는 점은 수리 전과 같다.
- **D2** [Low] [optional] `internal/graph/check.go:870` vs `:408` — OPEN. 3회차 설명 그대로다. 이번 diff는 이 영역을 바꾸지 않았다.
- **D3** [Low] [optional] `internal/cli/graph_check_policy_test.go:371`, `:415` — OPEN. 3회차 그대로다.
- **D4** [Low] [optional] `CHANGELOG.md:21`, `docs-site/content/{en,ko,ja,zh}/cli-reference/graph.md:76`, `internal/graph/check.go:99-104`, `internal/graph/check_policy.go` — OPEN. 3회차 그대로다.
- **D5** [Info] [optional] `CHANGELOG.md:21` — OPEN.
- **D6** [Info · 관측] [optional] `internal/lockfile/lockfile_windows_test.go:22` (`TestWindowsLockAllowsDataIOWhileExcludingContenders`)
  - 이 테스트는 파일명 `_windows_test.go` 때문에 Windows에서만 컴파일된다. 그런데 Windows에서 도는 CI job `Multi-OS Test (windows-latest)`는 `go test -count=1 -json -timeout 25m ./internal/hook/... ./internal/cli/worktree/...`만 실행한다(Claim F, 3회차 run 41555114270 로그). `internal/lockfile`은 그 목록에 없다. 그래서 "Windows CI가 끝나면 F1 실행 결과를 확인할 수 있다"는 3회차 권고 4번의 전제는 성립하지 않는다. job이 성공해도 이 테스트가 돈 증거가 되지 않는다.
  - 판정을 바꾸지 않는 이유: F1은 2·3회차에서 교차 vet과 소스 검토로 닫혔고, 이 PR이 처음 넣은 문제가 아니라 CI 범위의 문제다. 확신도 중간(다른 Windows job이 lockfile을 돌리는지는 이 run 하나만 봤다).
  - Required fix(선택): Windows job의 패키지 목록에 `./internal/lockfile/...`를 추가하거나, Windows 실행을 수동으로 한 번 기록한다.
- **N4** [Info] [optional] OPEN — 2회차 설명 그대로.
- **F7** [Low] [optional] OPEN — `internal/hook/failure_observer.go`, `inbox_lifecycle.go`는 이번 diff에 없다.
- **F8** [Info] [optional] OPEN — base는 여전히 `HEAD^1`(`check.go:822`, 이번 diff로 줄 번호도 바뀌지 않음).

## Codex leg (audit_multi, 원문 복사)

호출: `mcp__moai__audit_multi(project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001, target=baseBranch, delta_base=2fad4a8de6426f559d1c6d915c0b9e6a2c424e03, focus=<D1 수리 확인>)`. `gates`와 `card_id`는 넘기지 않았다. delta_base는 3회차 감사가 읽은 커밋이고, 승인 범위(D1 한 줄 수리의 확인 감사)에 맞춰 이미 감사된 범위의 결과가 다시 올라오지 않도록 썼다. 아래 Residual-risk에 그 한계를 적었다.

| 항목 | 값 (원문) |
|---|---|
| overall_verdict | `pass` |
| gate_unmet | 결과에 필드 없음 |
| plan_source | `config` |
| per_backend_verdicts | `codex` · gate `required` · verdict `pass` · summary `Verdict: pass` · findings `[]` |
| disagreement_flag | `null` |
| participant_count | `1` |
| fail_open_backends | `[]` |
| residual_risk_note | `delta_base 2fad4a8de6426f559d1c6d915c0b9e6a2c424e03: no claude or glm leg produced a verdict on it; codex was scoped in prose only (not mechanically bounded); no second-review record and no session state written` |
| audit_receipt | `rcpt-97d9474d1cf5df8ce8ae335a` |
| build_commit | `b1806498fe` |
| tree_root | `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001` |

이 감사와 Codex의 판정은 같다(둘 다 D1 수리에 결함 없음). 이 판정은 오케스트레이터 검증이 끝나기 전에는 최종이 아니다. 오케스트레이터가 위 digest 항목(`overall_verdict`, `gate_unmet`, `plan_source`, `per_backend_verdicts[].backend/gate/verdict`)을 `<toplevel>/.moai/state/audit-plan-result.json`에 새로 쓰고 `moai verify audit-plan --project-root <toplevel> --result-file <그 경로>`를 돌려야 한다.

## Audit plan

```
$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001
"config_status": "ok"
"model": "codex"
"model_source": "config"
backends: claude gate off · codex gate required · glm gate off (source config.gates, explicit true)
"cross_model_active": true
"cross_model_required": true
"enforced_required": [ "codex" ]
```

## 차원별 점수

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 86/100 | PASS | 수리는 De Morgan 동치(Claim A). `go test ./internal/cli/ -run GraphCheck` 26개 전부 PASS, 그중 N1 회귀 `SymlinkedStampTargetRewoundBlocks`·`IgnoredStampBlocks`와 F2 `RewoundStampBlocks` PASS. `go test ./internal/graph/` ok(Claim B) |
| Security (25%) | 85/100 | PASS | 이번 diff에 새 입력·신뢰 경계 없음. 수리 뒤에도 링크·특수 파일이면 `ok=false`(fail closed)이고 오류면 먼저 걸린다. D2는 3회차와 같은 Low |
| Craft (20%) | 82/100 | PASS | `golangci-lint run ./internal/graph/` → `0 issues.`, `./internal/cli/ ./internal/lockfile/` → `0 issues.`, gofmt 빈 출력, vet 호스트·Windows exit 0, CI `Lint` pass. 감점: D3 장치별 테스트 부족, coverage 미측정 |
| Consistency (15%) | 80/100 | PASS-WITH-DEBT | 수리 커밋 메시지가 카드 id·D1·Conventional Commit 형식을 지킨다. D4·D5 문서 정밀도 문제는 그대로 |

Must-pass(Functionality, Security)는 기준을 넘는다. blocking 결함 0, 필수 Codex gate `pass`, 필수 CI 체크 세 개 `pass`. 열린 optional 부채(D2~D6, N4, F7, F8)를 기록하므로 → **PASS-WITH-DEBT**.

## Claim / Evidence

**Claim A — 범위: 소스는 한 줄만 바뀌었고 동작은 같다**
```
$ git rev-parse HEAD
b7c4e23b8a1c55e51bcd38531031f6421bde2efb
$ git log --oneline 2fad4a8de6..HEAD
b7c4e23b8a [t2001] fix(graph): apply De Morgan to the codemaps tree walk condition so Lint passes (D1, card t2001)
3a08ec7e23 [t2001] docs(audit): record the round-3 fix-confirm sync audit verdict FAIL at 2fad4a8de6 (card t2001)
$ git diff 2fad4a8de6 HEAD --stat
 .moai/reports/t2001/sync-audit-3.md | 329 ++++++++++++++++++++++++++++++++++++
 internal/graph/check.go             |   2 +-
 2 files changed, 330 insertions(+), 1 deletion(-)
$ git diff 2fad4a8de6 HEAD -- internal
@@ -889,7 +889,7 @@ func regularCodemapsTree(root string) bool {
-		if err != nil || !(d.IsDir() || d.Type().IsRegular()) {
+		if err != nil || (!d.IsDir() && !d.Type().IsRegular()) {
```
`internal/` 밖에서 바뀐 것은 3회차 판정 파일 하나뿐이다. 수리 뒤 `check.go:886-899`를 직접 읽었고, 바뀐 줄 외에는 3회차가 본 함수와 같다.

**Claim B — 범위 테스트 (macOS, 이 트리, HEAD b7c4e23b8a)**
```
$ uptime
10:36  up 2 days,  6:09, 26 users, load averages: 40.59 32.48 25.84
$ go test ./internal/cli/ -run GraphCheck -count=1 -v
(26개 테스트 전부 --- PASS, FAIL 0) 그중:
--- PASS: TestGraphCheckPolicy_RewoundStampBlocks (1.84s)
--- PASS: TestGraphCheckPolicy_UntouchedStampStillWarns (2.62s)
--- PASS: TestGraphCheckPolicy_CrossedByChangeBlocks (1.81s)
--- PASS: TestGraphCheckPolicy_BaseOverLinePrZeroWarns (1.87s)
--- PASS: TestGraphCheckPolicy_InheritedPullRequestWarns (1.72s)
--- PASS: TestGraphCheckPolicy_SymlinkedStampTargetRewoundBlocks (1.07s)
--- PASS: TestGraphCheckPolicy_IgnoredStampBlocks (0.86s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	31.631s
$ go test ./internal/graph/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/graph	49.777s
```

**Claim C — 정적 검사**
```
$ golangci-lint run ./internal/graph/
0 issues.
$ golangci-lint run ./internal/cli/ ./internal/lockfile/
0 issues.
$ golangci-lint version
golangci-lint has version v2.1.6 built with go1.26.8 from (unknown, modified: ?, mod sum: "h1:LXqShFfAGM5BDzEOWD2SL1IzJAgUOqES/HRBsfKjI+w=") on (unknown)
$ gofmt -l internal/graph internal/cli/graph_check_policy_test.go internal/cli/graph_check.go internal/lockfile internal/config
(빈 출력)
$ go vet ./internal/graph/ ./internal/cli/ ./internal/lockfile/
(빈 출력, exit 0)
$ env GOOS=windows go vet ./internal/graph/ ./internal/cli/ ./internal/lockfile/
(빈 출력, exit 0)
```

**Claim D — CI: Lint가 새 head에서 통과 (읽기만, 재실행·취소 없음)**
```
$ gh run list --branch WT-cc-bash-monitor --limit 8 --json headSha,name,status,conclusion,databaseId
{"databaseId":41560649278,"headSha":"b7c4e23b8a1c55e51bcd38531031f6421bde2efb","name":"CI","status":"in_progress"}
{"databaseId":41560649230,"headSha":"b7c4e23b8a1c55e51bcd38531031f6421bde2efb","name":"PR Multi-OS Gate","status":"in_progress"}
{"databaseId":41560649231,"headSha":"b7c4e23b8a1c55e51bcd38531031f6421bde2efb","name":"CodeQL","conclusion":"success"}
{"databaseId":41560649233,"headSha":"b7c4e23b8a1c55e51bcd38531031f6421bde2efb","name":"Graph Freshness","conclusion":"success"}
(Release PR Multi-OS Verification, Template Neutrality Check, docs i18n parity check, lsel-leak-guard: 모두 같은 head, success)
$ gh api repos/modu-ai/moai-adk/branches/main/protection/required_status_checks --jq .contexts
["Lint","Build (linux/amd64)","Analyze (Go) (go)"]
$ gh pr checks 1955 --required
Analyze (Go) (go)	pass	1m57s	https://github.com/modu-ai/moai-adk/actions/runs/41560649231/job/125670281054
Build (linux/amd64)	pass	2m25s	https://github.com/modu-ai/moai-adk/actions/runs/41560649278/job/125670288693
Lint	pass	4m2s	https://github.com/modu-ai/moai-adk/actions/runs/41560649278/job/125670288557
```
`Lint` job은 run 41560649278 소속이고, 그 run의 headSha는 `b7c4e23b8a1c55e51bcd38531031f6421bde2efb`다(`gh run view 41560649278 --json headSha` 확인). 따라서 Lint 통과는 새 head에 대한 결과다. 같은 시각 `gh pr checks 1955` 전체 목록에서 나머지 job은 다음 상태였다.
```
$ gh run view 41560649278 ... (success·skipped가 아닌 job)
Race Test (internal/cli 1/2/3/4 of 4), Race Test (rosterguard), Race Test (remainder packages),
Test (browser fire guard), Test (ubuntu-latest): 모두 in_progress
$ gh run view 41560649230 ...
Multi-OS Test (windows-latest): in_progress · Multi-OS Test (macos-latest): in_progress
```

**Claim E — 판정 도구 빌드 provenance (§2.2)**
```
$ moai version
 v3.2.0-rc.39   v3.2.0-1041-gdad3603de1   built 2026-11-02T19:17:51Z
$ git merge-base --is-ancestor dad3603de1 HEAD
(exit 1)
$ git merge-base --is-ancestor b1806498fe HEAD
(exit 1)
```

**Claim F — Windows CI 범위 (D6)**
```
$ gh run view 41555114270 --json headSha,status,conclusion,jobs   (3회차 head)
{"conclusion":"success","headSha":"2fad4a8de6426f559d1c6d915c0b9e6a2c424e03", ... "Multi-OS Test (windows-latest)" success ...}
$ gh run view 41555114270 --job 125652179796 --log | grep -E "TestWindowsLock|internal/lockfile" | head -20
(빈 출력)
$ gh run view 41555114270 --job 125652179796 --log | grep -E "lockfile|go test" | head -20
Multi-OS Test (windows-latest)	Run go test	... ##[group]Run go test -count=1 -json -timeout 25m ./internal/hook/... ./internal/cli/worktree/...
$ grep -rn "func TestWindowsLockAllowsDataIOWhileExcludingContenders" internal
internal/lockfile/lockfile_windows_test.go:22:func TestWindowsLockAllowsDataIOWhileExcludingContenders(t *testing.T) {
```

## Baseline-attribution

모든 명령은 이 실행에서 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001`, HEAD `b7c4e23b8a1c55e51bcd38531031f6421bde2efb`를 대상으로 돌렸다. `go test`, `go vet`, `gofmt`, `golangci-lint`는 트리 소스를 직접 컴파일하므로 판정 빌드가 곧 트리 HEAD다. `moai verify audit-plan`은 설치 빌드 `dad3603de1`, Codex leg는 MCP 서버 빌드 `b1806498fe`가 판정했다. 둘 다 HEAD의 조상이 아니다(`git merge-base --is-ancestor` exit 1). 즉 이 트리에서 만든 빌드가 아니다. 다만 둘 다 판정 대상 Go 코드를 실행하지 않는다. audit-plan은 설정만 읽고, Codex는 diff와 소스를 직접 읽는다. CI 결과는 head `b7c4e23b8a` 기준 run 41560649278, 41560649230, 41560649231이다. D6 근거는 3회차 head `2fad4a8de6`의 run 41555114270이고, Windows job의 패키지 목록이 새 head에서 바뀌었는지는 워크플로를 바꾼 커밋이 없다는 점(Claim A, `.github/` 변경 없음)으로만 이어 붙였다.

## Gaps (관측하지 않은 것)

- **Windows에서 새 잠금 테스트가 실행된 기록은 여전히 없다.** 게다가 Windows CI job은 `internal/lockfile`을 아예 돌리지 않는다(D6). `GOOS=windows go vet`(테스트 파일 컴파일 포함) exit 0과 `Build (windows/amd64)` pass만 확인했다.
- **coverage는 측정하지 않았다**(부하 평균 40 수준이라 범위 테스트만 돌렸다).
- 새 head의 CI 중 필수가 아닌 job은 이 감사 시점에 끝나지 않았다. `Test (ubuntu-latest)`, Race Test 7개, `Test (browser fire guard)`, `Multi-OS Test (windows-latest)`, `Multi-OS Test (macos-latest)`는 `in_progress`였다. 진행 중은 통과가 아니다. 필수 체크 세 개만 완료·pass를 관측했다.
- Codex leg는 delta 범위(`2fad4a8de6..HEAD`)만 보았다. 결과 메모대로 이 범위는 프롬프트로만 전달되었고 기계적으로 묶이지 않았다. PR 전체 diff에 대한 Codex 판정은 3회차(`fail`, D2·D4·D1)가 마지막이다.
- D2·D3·D4는 이번에 다시 재현하지 않았다. 해당 코드와 문서가 바뀌지 않았으므로(Claim A) 3회차 관측을 그대로 이어 쓴다.
- 2회차에서 본 hook 범위 테스트는 이번 diff가 그 파일을 건드리지 않아 다시 돌리지 않았다.
- D6에서 "다른 Windows job도 lockfile을 돌리지 않는다"는 판단은 run 41555114270의 job 하나만 보고 내렸다. `Test (${{ matrix.os }})` job은 skipping이라 범위를 확인하지 않았다.
- 도구 거부: 이번 실행에서 거부된 명령은 없었다. 파이프를 쓴 `gh run view --log | grep` 명령 두 개도 실행되었다.

## Residual-risk

- Codex를 PR 전체 diff로 다시 돌리면 D2·D4를 P2로 다시 보고해 `fail`이 날 수 있다. 이번 `pass`는 D1 수리 delta에 대한 판정이다. D2·D4는 이 감사와 3회차가 optional로 분류했지만, Codex gate는 그 분류를 따르지 않는다.
- D6 때문에 Windows `LockFileEx` 경로는 실제 Windows에서 한 번도 실행되지 않은 채 병합될 수 있다. 컴파일과 vet만 확인된 상태다.
- 안정성 판정은 여전히 "PR이 판정 코드와 임계값을 바꾸지 않았다"는 전제에 기댄다(N4).
- 정책상 inherited drift가 PR에서 warn으로 통과하면 병합 뒤 main push 실행은 block한다. 누군가 codemaps를 다시 만들 때까지 main이 빨간색으로 남을 수 있다(R208이 의도한 trade-off).
- 아직 끝나지 않은 CI job(Race·ubuntu·Multi-OS)이 이후 실패할 수 있다. 이번 diff는 조건식 한 줄이라 그럴 가능성은 낮지만 관측하지는 않았다.

## Iteration history

- 1회차 (`sync-audit-1.md`): FAIL @ b2387e7123 — blocking F1·F2, Codex required gate `fail` (rcpt-9ac33515965037b2a13e2e8c).
- 2회차 (`sync-audit-2.md`): FAIL @ 7a946cfca2 — F1·F2·F3-F6 CLOSED, NEW N1(blocking)·N2·N3·N4, Codex `fail` (rcpt-b5d843df9399068fbf1d6832).
- 3회차 (`sync-audit-3.md`): FAIL @ 2fad4a8de6 — N1·N2·N3 CLOSED, NEW D1(blocking, 필수 `Lint` 실패)·D2·D3·D4·D5, Codex `fail` (rcpt-ebc3d97cd2fdfecd56edf95b).
- 4회차 (이 파일): PASS-WITH-DEBT @ b7c4e23b8a — D1 CLOSED, NEW D6(Info, optional), D2~D5·N4·F7·F8 OPEN 유지, Codex `pass` (delta, rcpt-97d9474d1cf5df8ce8ae335a).

## 권고 (모두 optional, 이 카드의 병합을 막지 않음)

1. D6: Windows CI job 패키지 목록에 `./internal/lockfile/...`를 넣거나 Windows 실행을 한 번 수동으로 기록한다.
2. D2: `checkCodemaps`가 읽은 stamp 바이트를 안정성 함수로 넘긴다(Codex 3회차 P2-1).
3. D4·D5: 문서 문장을 실제 보장에 맞추고 CHANGELOG의 이중 "or"를 지운다. D3은 장치별 단위 테스트로 보강한다.
4. 병합 전에 남은 CI job(Race·ubuntu·Multi-OS)의 완료를 읽는다.

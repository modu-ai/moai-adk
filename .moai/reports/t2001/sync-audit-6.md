auditor-model: claude-opus-5-5

verdict: PASS-WITH-DEBT
audited_sha: 608fee9e24076e531544fc4134fd27fcab48320f

# t2001 sync 감사 6회차 (최종, G1·G2·G3 수리 확인) — WT-cc-bash-monitor

- 카드: t2001 (Class B, SPEC 없음). Claude 단독 감사, Codex·GLM·audit_multi 호출 없음(유료 호출 미승인, Codex gate advisory).
- 고정 대상: `git rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001`, `git branch --show-current` → `WT-cc-bash-monitor`, `git rev-parse HEAD` → `608fee9e24076e531544fc4134fd27fcab48320f`. 감사 시작과 이 파일을 쓰기 직전에 두 번 확인했고 같았다. `git status --short`는 두 번 모두 빈 출력. 프로브와 RED 재현은 `go test -overlay`로 scratchpad 파일만 끼워 넣었고 트리는 고치지 않았다.
- 종합: **PASS-WITH-DEBT, 점수 90/100**. 5회차의 유일한 blocking 결함 G1이 닫혔다. 겹침 사례(Value=6, Contribution=2, BaseCount=6, threshold=5, inherited)에서 문장이 이제 "the base alone carries 6 >= threshold 5, this change added 2"를 찍는다. 수리 전 코드에 새 테스트를 끼우면 세 하위 테스트가 모두 실패하는 것도 직접 재현했다. G2·G3 CHANGELOG 문구는 코드와 맞는다. `internal/cli`·`internal/graph`의 비테스트 코드에는 `Value - Contribution` 뺄셈이 더 남아 있지 않다. 남은 부채는 범위 밖 optional 항목(D2-D6, N4, F7, F8, G4, G5)이고, 이번 회차에 새로 생긴 blocking 결함은 없다.

## 이전 항목 상태

| 항목 | 상태 |
|---|---|
| **G1** (stderr 문장의 Value−Contribution) | **CLOSED** — Claim B, C, D |
| **G2** (CHANGELOG "releases its lock explicitly") | **CLOSED** — Claim E |
| **G3** (CHANGELOG offset 0) | **CLOSED** — Claim E |
| G4 (`migrate_lock_windows.go:12-14`의 낡은 주석) | OPEN (Info, optional, outside, main 소유) — 그대로 남아 있음 |
| G5 (dirty-generation 경로의 `anchor-moved` 이름) | OPEN (Info, optional) — 이번 diff 밖 |
| D2, D3, D4, D5, D6, N4, F7, F8 | OPEN (optional) — 범위 밖, 지시대로 다시 보지 않음 |

## 결함 목록 (structured defect-list)

새 blocking 결함 없음. 아래는 모두 optional이다.

- **H1** [Info] [optional] [in-diff] `internal/cli/graph_check.go:312-319` — `BaseCount == nil`일 때의 inherited·crossed-by-change 문장 분기는 실제 실행에서는 닿지 않는다. `driftOrigin`(`internal/graph/check_policy.go:127-128`)이 `BaseCount == nil`을 이미 `anchor-moved`로 보내기 때문이다. 5회차 G1 Required fix가 "숫자를 지어내지 않게" 하라고 명시한 방어 분기라서 수리 요구는 없다. 직접 만든 `LayerReport`(테스트, 또는 정책을 거치지 않은 미래 호출자)에서만 쓰인다. 확신도: 높음.
- **H2** [Info] [optional] [touched-file] `internal/cli/graph_check.go:345-347` — `writeInheritedDriftWarnings`의 `::warning` 줄은 "inherited: the base alone carries the count (%s)" 안에 `driftOriginSentence`를 넣어서 "the base alone carries"가 두 번 나온다. 숫자는 이제 맞으므로 정확성 문제는 아니고 문장 중복일 뿐이다. 수리 요구 없음. 확신도: 높음.

## 차원별 점수

| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 92/100 | PASS | 프로브: `PROBE inherited => inherited — the base alone carries 6 >= threshold 5, this change added 2`. RED 재현(수리 전 코드 overlay): `--- FAIL: TestDriftOriginSentence_UsesMeasuredBaseCount` 하위 3개 모두 FAIL. 현재 코드: `go test -count=1 -run 'DriftOriginSentence\|GraphCheck' -v ./internal/cli/` → `ok ... 134.954s`, `--- PASS` 30건, `--- FAIL` 0건. `go test -count=1 ./internal/graph/` → `ok ... 233.494s` |
| Security (25%) | 92/100 | PASS | 새 입력 경계 없음. 바뀐 코드는 이미 측정된 `*int` 필드를 `fmt.Sprintf`로 찍는 것뿐이고 nil 역참조는 nil 검사 뒤에서만 일어난다(`:312`, `:317`). CHANGELOG는 문서 |
| Craft (20%) | 84/100 | PASS | `go vet ./internal/cli/ ./internal/graph/` exit 0, `GOOS=windows go vet` 같은 범위 exit 0, `gofmt -l internal/cli internal/graph` 빈 출력. golangci-lint는 timeout으로 미완료(exit 4, `Timeout exceeded`), 끊기기 전 보고 6건은 main과 같은 테스트 파일의 errcheck(카드 밖) — Claim F. 새 테스트는 RED를 실제로 잡는다(Claim C). 감점: lint 미완료, 종단(빌드된 `moai graph check`) stderr 테스트 없음 |
| Consistency (15%) | 88/100 | PASS | 함수 주석이 새 동작을 설명한다(:300-305). CHANGELOG 문구가 `lockfile_windows.go:20-47`과 `withMigrationLock`(:124-137)에 맞는다. `docs-site/content/en/cli-reference/graph.md:76`의 "`base_count`, not derived by subtraction" 설명과 이제 CLI 문장도 일치. 커밋 제목은 `[t2001] fix(cli): ...` 형식. 감점: G4 낡은 주석(main 소유) 그대로 |

가중 합 0.4×92 + 0.25×92 + 0.2×84 + 0.15×88 = 36.8 + 23.0 + 16.8 + 13.2 = 89.8 → **90/100**. Must-pass(Functionality, Security) 둘 다 기준 이상. blocking 결함 0건, optional 부채가 남아 있으므로 → **PASS-WITH-DEBT**.

## Claim / Evidence

**Claim A — 5회차 기록 이후 바뀐 것은 이 커밋 하나, 파일 세 개뿐이다**

```
$ git diff --stat 138d081fb1 HEAD
 CHANGELOG.md                                     |  2 +-
 internal/cli/graph_check.go                      | 19 ++++++--
 internal/cli/graph_check_origin_sentence_test.go | 58 ++++++++++++++++++++++++
 3 files changed, 73 insertions(+), 6 deletions(-)
```

**Claim B — 겹침 사례에서 문장은 측정된 BaseCount를 찍는다 (G1 닫힘)**

`zz_audit_probe_test.go`를 scratchpad에 두고 `-overlay`로 `internal/cli`에 끼워 `driftOriginSentence`를 직접 불렀다.

```
$ go test -count=1 -overlay .../scratchpad/overlay_probe.json -run 'ZZAuditProbe' -v ./internal/cli/
    zz_audit_probe_test.go:19: PROBE inherited => inherited — the base alone carries 6 >= threshold 5, this change added 2
    zz_audit_probe_test.go:19: PROBE crossed-by-change => crossed by this change — the base carried 4 < threshold 5, this change added 3 and took the count over the line
    zz_audit_probe_test.go:19: PROBE inherited => inherited — threshold 5, this change added 2 (the base's own count was not measured)
    zz_audit_probe_test.go:19: PROBE crossed-by-change => crossed by this change — threshold 5, this change added 2 and took the count over the line (the base's own count was not measured)
    zz_audit_probe_test.go:19: PROBE originated => originated by this change — it alone is 5 >= threshold 5
--- PASS: TestZZAuditProbeOriginSentence (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.797s
```

입력 순서: (1) Value=6 C=2 B=6 inherited, (2) Value=6 C=3 B=4 crossed, (3) B=nil inherited, (4) B=nil crossed, (5) C=5 B=1 originated. 1번이 5회차 G1의 겹침 사례이며 이제 "carries 6"이다.

**Claim C — 새 테스트는 수리 전 코드에서 실제로 실패한다 (구현자의 RED 보고를 독립 재현)**

`git show 138d081fb1:internal/cli/graph_check.go`(수리 전 원본 파일 전체)를 scratchpad로 꺼내 overlay로 바꿔 끼웠다.

```
$ go test -count=1 -overlay .../scratchpad/overlay_red.json -run 'DriftOriginSentence|ZZAuditProbe' -v ./internal/cli/
    zz_audit_probe_test.go:19: PROBE inherited => inherited — the base alone carries 4 >= threshold 5, this change added 2
    zz_audit_probe_test.go:19: PROBE crossed-by-change => crossed by this change — the base carried 3 < threshold 5, this change added 3 and took the count over the line
    ...
    graph_check_origin_sentence_test.go:25: sentence must state the measured base 6, got: "inherited — the base alone carries 4 >= threshold 5, this change added 2"
    graph_check_origin_sentence_test.go:28: sentence must not print the subtracted figure 4, got: "inherited — the base alone carries 4 >= threshold 5, this change added 2"
    graph_check_origin_sentence_test.go:51: inherited: no base figure may be printed without a measurement, got: "inherited — the base alone carries 4 >= threshold 5, this change added 2"
    graph_check_origin_sentence_test.go:51: crossed-by-change: no base figure may be printed without a measurement, got: "crossed by this change — the base carried 4 < threshold 5, this change added 2 and took the count over the line"
    graph_check_origin_sentence_test.go:39: sentence must state the measured base 4 < 5, got: "crossed by this change — the base carried 3 < threshold 5, this change added 3 and took the count over the line"
--- FAIL: TestDriftOriginSentence_UsesMeasuredBaseCount (0.00s)
    --- FAIL: TestDriftOriginSentence_UsesMeasuredBaseCount/inherited_with_overlap_prints_the_measured_base (0.00s)
    --- FAIL: TestDriftOriginSentence_UsesMeasuredBaseCount/nil_BaseCount_prints_no_base_number (0.00s)
    --- FAIL: TestDriftOriginSentence_UsesMeasuredBaseCount/crossed-by-change_stays_truthful (0.00s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.416s
red exit=1
```

**Claim D — 비테스트 코드에 Value−Contribution 뺄셈이 더 없다**

```
$ grep -rn "Contribution\|BaseCount" --include='*.go' internal/cli internal/graph | grep -v '_test.go'
internal/cli/graph_check.go:222:	case l.Contribution != nil:
internal/cli/graph_check.go:223:		base := l.ContributionBase
internal/cli/graph_check.go:228:			*l.Contribution, base, driftOriginSentence(l))
...
internal/cli/graph_check.go:315:		return fmt.Sprintf("inherited — the base alone carries %d >= threshold %d, this change added %d", *l.BaseCount, l.Threshold, *l.Contribution)
internal/cli/graph_check.go:320:		return fmt.Sprintf("crossed by this change — the base carried %d < threshold %d, ...", *l.BaseCount, l.Threshold, *l.Contribution)
internal/cli/graph_check.go:343:		if l.Contribution != nil {
internal/cli/graph_check.go:344:			contribution = fmt.Sprintf("%d", *l.Contribution)
internal/graph/check.go:131:	Contribution *int `json:"contribution,omitempty"`
internal/graph/check.go:145:	BaseCount *int `json:"base_count,omitempty"`
internal/graph/check.go:858:	out, err := gitOutput(projectRoot, append([]string{"diff", "--name-only", anchor, rep.ContributionBase, "--"}, roots...)...)
internal/graph/check_policy.go:119:	if l.Contribution == nil {
internal/graph/check_policy.go:123:	case *l.Contribution >= threshold:
internal/graph/check_policy.go:127:	case l.BaseCount == nil:
internal/graph/check_policy.go:129:	case *l.BaseCount >= threshold:
$ grep -rn "Value *-\|Value-\|- \*.*Contribution" --include='*.go' internal/cli internal/graph | grep -v _test.go
internal/graph/check_policy.go:38:// anchor and the first parent; not Value - Contribution, since the two change
internal/graph/check_policy.go:106:// It is not Value - Contribution: Value is the union of the base's and the
```

(전체 grep 출력 중 주석·필드 선언 줄은 일부 생략. 뺄셈 grep의 두 건은 모두 "뺄셈을 쓰지 않는다"는 주석이다.) 각 히트를 읽었다: `:222-228`(`writeLayerAttribution`)은 `ContributionBase` SHA 9자를 찍고 base 숫자는 `driftOriginSentence`에 맡긴다. `:343-344`(`writeInheritedDriftWarnings`)는 Contribution 자체만 찍는다. `check.go`는 측정·필드, `check_policy.go`는 `BaseCount`로 판정한다. base 숫자를 뺄셈으로 만드는 곳은 남아 있지 않다.

**Claim E — G2·G3 CHANGELOG 문구가 코드와 맞다**

```
$ grep -c "releases its lock explicitly" CHANGELOG.md
0
$ grep -c "one byte at offset 0xFFFFFFFF00000000" CHANGELOG.md
1
$ grep -c "one byte at offset 0 " CHANGELOG.md
0
$ grep -n -i "offset\|0xFFFF\|LockFileEx\|UnlockFileEx" internal/lockfile/*windows*.go
internal/lockfile/lockfile_windows.go:20:// The locked region is ONE BYTE FAR BEYOND ANY FILE DATA (offset
internal/lockfile/lockfile_windows.go:21:// 0xFFFFFFFF00000000), not byte 0. Win32 byte-range locks are mandatory, so a
internal/lockfile/lockfile_windows.go:41:// lockOffsetHigh puts the one-byte lock region at offset 0xFFFFFFFF00000000.
internal/lockfile/lockfile_windows.go:42:const lockOffsetHigh = 0xFFFFFFFF
internal/lockfile/lockfile_windows.go:47:	return &windows.Overlapped{OffsetHigh: lockOffsetHigh}
internal/lockfile/lockfile_windows.go:52:	return windows.LockFileEx(windows.Handle(f.Fd()), flags, 0, 1, 0, lockOverlapped())
internal/lockfile/lockfile_windows.go:76:	err := windows.UnlockFileEx(windows.Handle(f.Fd()), 0, 1, 0, lockOverlapped())
```

새 CHANGELOG 문장: "a `LockFileEx` byte-range lock on one byte at offset 0xFFFFFFFF00000000 (past any data region, so data I/O stays allowed while contenders are excluded)" — 코드 주석 :20-26과 같은 설명이다. 마이그레이션 쪽은 "the worktree migration rebinds the card by its host-form path instead of git's slash form"만 남았다. 코드 대조:

- `internal/cli/worktree/migrate.go:124-137` `withMigrationLock`: `defer func() { _ = f.Close() }()` 다음 `tryMigrationLock(f)`; 명시적 Unlock 호출 없음. 그러므로 빠진 절("releases its lock explicitly")은 빠지는 것이 맞다.
- `migrate_lock_other.go`(`!windows`)는 `lockfile.TryLock(f)`, `migrate_lock_windows.go`는 `windows.LockFileEx(..., 0, 1, 0, &overlapped)`를 직접 부른다(핸들 닫힘으로 해제).
- `migrate.go:1727-1735` `rebindCardForPath`: `oldPath = filepath.FromSlash(oldPath)`, `newPath = filepath.FromSlash(newPath)` — "host-form path instead of git's slash form"과 맞다.

**Claim F — 집중 회귀와 정적 검사**

```
$ go test -count=1 -run 'DriftOriginSentence|GraphCheck' -v ./internal/cli/ > .../t_cli.log 2>&1; echo "cli exit=$?"
cli exit=0
--- PASS: TestDriftOriginSentence_UsesMeasuredBaseCount (0.00s)
    --- PASS: TestDriftOriginSentence_UsesMeasuredBaseCount/inherited_with_overlap_prints_the_measured_base (0.00s)
    --- PASS: TestDriftOriginSentence_UsesMeasuredBaseCount/nil_BaseCount_prints_no_base_number (0.00s)
    --- PASS: TestDriftOriginSentence_UsesMeasuredBaseCount/crossed-by-change_stays_truthful (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	134.954s
(grep -c "--- PASS" → 30, grep -c "--- FAIL" → 0)

$ go test -count=1 ./internal/graph/      → graph exit=0
ok  	github.com/modu-ai/moai-adk/internal/graph	233.494s
$ go vet ./internal/cli/ ./internal/graph/                 → vet exit=0
$ GOOS=windows go vet ./internal/cli/ ./internal/graph/    → win vet exit=0
$ gofmt -l internal/cli/graph_check.go internal/cli/graph_check_origin_sentence_test.go internal/cli internal/graph
(빈 출력)                                                   → gofmt exit=0
$ golangci-lint run ./internal/graph/... ./internal/cli/ > .../lint.log 2>&1; echo "lint exit=$?"
lint exit=4
(lint.log, 21줄, 앞 2건)
internal/cli/factory_aff_s2_lease_t2013_test.go:73:8: Error return value is not checked (errcheck)
internal/cli/factory_lease_bound_test.go:77:4: Error return value is not checked (errcheck)
(뒤 4건)
internal/cli/factory_leaselost_t2013_test.go:87:8: Error return value is not checked (errcheck)
internal/cli/mcp_server_test.go:557:8: Error return value is not checked (errcheck)
internal/cli/mcp_server_test.go:560:4: Error return value is not checked (errcheck)
internal/cli/migrate_home_state_test.go:343:8: Error return value is not checked (errcheck)
6 issues:
* errcheck: 6
level=error msg="Timeout exceeded: try increasing it by passing --timeout option"

$ git diff --stat b25adf4fca HEAD -- internal/cli/factory_aff_s2_lease_t2013_test.go internal/cli/factory_lease_bound_test.go internal/cli/factory_leaselost_t2013_test.go internal/cli/mcp_server_test.go internal/cli/migrate_home_state_test.go
(빈 출력)
```

golangci-lint는 설정 `run.timeout: 5m`(`.golangci.yml:21`)을 넘겨 **끝까지 돌지 못했다**(exit 4, 부하 load average 120-300 아래 약 21분 경과). 끝나기 전에 보고된 6건은 모두 테스트 파일의 errcheck이고, 다섯 파일 모두 흡수한 main `b25adf4fca`와 바이트 단위로 같다(위 diff 빈 출력). 즉 카드 diff 밖(`outside`)이며 판정에 넣지 않는다. 바뀐 두 파일(`graph_check.go`, `graph_check_origin_sentence_test.go`)에 대한 보고는 없었지만, 실행이 중간에 끊겼으므로 "lint clean"으로 주장하지 않는다(Gaps 참조).

**Claim G — 문서와 일치**

`docs-site/content/en/cli-reference/graph.md:76`: "`inherited` (the base alone — its own count, measured directly between the anchor and the first parent and reported as `base_count`, not derived by subtraction — ...)". 이제 stderr 문장도 같은 `base_count`를 찍으므로 JSON·문서·stderr가 한 숫자를 말한다. `grep -rln "value - contribution\|Value - Contribution\|value minus contribution\|Value minus Contribution" --include='*.md' .`(reports 제외) → 빈 출력.

## Baseline-attribution

모든 명령은 이번 실행에서 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001`, HEAD `608fee9e24076e531544fc4134fd27fcab48320f`에 대해 돌렸다. `go test`·`go vet`·`gofmt`·`golangci-lint`는 트리 소스를 직접 컴파일하므로 판정 빌드가 곧 트리 HEAD다(`go version` → `go1.27.3 darwin/arm64`; golangci-lint 2.15.0, 설치본이지만 트리 소스를 읽는 외부 도구라 §2.2의 자기 빌드 지연 문제가 없다). 이번에는 빌드 캐시·IO 오류가 나지 않아 전용 `GOCACHE`로 다시 돌린 명령은 없다. Claim B·C는 `-overlay`로 scratchpad 파일을 끼운 빌드이고, Claim C는 `138d081fb1`의 원본 `graph_check.go` 전체를 썼다(부분 되돌림 아님). 실행 중 머신 부하는 load average 약 124-130(`uptime`)이었다.

## Gaps

- 거부된 도구 호출: worktree 격리 가드가 복합 셸 명령 2건을 거부했다(①`cd ... && go test ... | grep ...` 파이프 묶음, ②`cat > ... <<EOF` + `printf` + `git -C` 묶음). 둘 다 단순 명령과 Write 도구로 나눠 다시 실행했다. 소스 읽기로 대체한 측정은 없다.
- golangci-lint가 끝까지 돌지 않았다: `./internal/graph/... ./internal/cli/` 실행이 설정 timeout 5m을 넘겨 exit 4(`Timeout exceeded`)로 끝났다. 끊기기 전에 보고된 6건은 카드 밖(main과 동일한 테스트 파일). 바뀐 `graph_check.go`·새 테스트 파일은 lint가 깨끗하다고 관측하지 못했고, 대신 `go vet`(호스트·Windows)과 `gofmt -l`이 깨끗하다는 것만 관측했다. `internal/graph`는 이번 커밋에서 바뀌지 않았다(Claim A). 다른 세션의 부하(load average 120-300) 때문에 `--timeout`을 늘려 다시 돌리지 않았다.
- 빌드된 `moai graph check`를 겹침 픽스처에 돌려 stderr를 받는 종단 관측은 하지 않았다. `writeLayerAttribution`(:227-228)과 `writeInheritedDriftWarnings`(:345-347)가 이 함수 결과를 그대로 찍는 것은 코드 읽기로 확인했다.
- `internal/cli` 전체 스위트, `go build ./...`, `internal/lockfile` 테스트, 커버리지 측정은 돌리지 않았다(지시 범위·부하). `internal/cli` 커버리지는 이번 회차에 측정하지 않았다.
- Windows는 vet(컴파일)만 했고 실제 Windows 실행은 없다(D6 그대로).
- `moai verify audit-plan`은 이번 회차에 돌리지 않았다. 리더 지시로 Codex/GLM/audit_multi가 금지되었고 Claude 단독 경로가 고정되어 있어 계획 결과가 호출할 백엔드를 바꿀 수 없었다. Codex 영수증 없음.
- ko·ja·zh 문서는 이번에 다시 읽지 않았다(이번 커밋이 문서를 건드리지 않음, Claim A).
- 원격 CI·PR head는 읽지 않았다(로컬 AF-Flow 착지 전 재감사).

## Residual-risk

- 종단 CLI 테스트가 없어서, 나중에 다른 출력 함수가 base 숫자를 따로 계산하면 stderr와 JSON이 다시 어긋날 수 있다. 지금은 숫자를 찍는 곳이 `driftOriginSentence` 한 곳뿐이라 위험은 낮다.
- BaseCount 정확성은 여전히 `HEAD^1`이 base라는 전제(F8)와 anchor 안정성(`CodemapsInputsStable`)에 기댄다. 5회차와 같다.
- Windows `LockFileEx` 잠금 해제 시점이 핸들 닫힘 뒤 시스템 자원에 달렸다는 이론적 위험(5회차 Residual-risk)은 그대로이고 main 설계에 속한다.
- 실행 중 부하가 매우 높았다(load average 120+). 타이밍에 민감한 테스트가 이번에는 통과했지만 부하가 결과를 가렸을 가능성은 배제하지 못한다.

## 반복 기록

- 1회차 (`sync-audit-1.md`): FAIL.
- 2회차 (`sync-audit-2.md`): FAIL @ 7a946cfca2.
- 3회차 (`sync-audit-3.md`): FAIL @ 2fad4a8de6 — D1 blocking.
- 4회차 (`sync-audit-4.md`): PASS-WITH-DEBT @ b7c4e23b8a — D1 CLOSED, D6 NEW.
- 5회차 (`sync-audit-5.md`): FAIL @ e60f0e02b0 — G1 blocking, G2·G3·G4·G5 optional.
- 6회차 (이 파일): PASS-WITH-DEBT @ 608fee9e24 — G1·G2·G3 CLOSED(G1은 RED 독립 재현), 새 blocking 0건, optional H1·H2(Info). repair-born blocking 0건 → 수리 중단 규칙 카운트 0.

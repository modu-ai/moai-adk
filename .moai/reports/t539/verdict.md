# 판정 — SPEC-CTX-BLIND-DOUBLE-001 (카드 t539, run-phase)

> 트리 `.claude/worktrees/t539` · 브랜치 `WT-ctx-blind-double` · 기준 `dbc1f7125`
> 5절 형식(Claim / Evidence / Baseline-attribution / Gaps / Residual-risk).
> 인용한 로그 경로는 전부 `.moai/reports/t539/` 아래 실재한다.
>
> 적용한 정책 규칙(policy-rule 인용, `verification-completeness.md` §7):
> `verification-completeness.md` §1.1(공허한 초록) · §2(두 셀 채택) · §2.1(undecidable disposition) ·
> `verification-claim-integrity.md` §1(관측하지 않은 주장 금지) · §2(baseline 귀속).

---

## 1. Claim — 이 카드가 주장하는 것

1. **M1**: 판별식(a)(b)(c)(d)과 스윕 산출물이 이 트리에서 재현된다. 스캐너는 `internal/` 로
   승격되지 않았다. — 단, 이 3건은 **regression-guard 이며 통과로 기록하지 않는다.**
2. **M2**: GLM audit 경로에 취소 가드가 없었고, 그 원인은 프로덕션이 아니라 **대역**이었다.
   가드를 세웠고, 채택 근거는 커버리지가 아니라 **뮤턴트 삼단**(생존 → 수리 → 사망)이다.
   프로덕션 코드는 0줄 바뀌었다.
3. **M3**: 다섯 후보군 84건 전부가 근거를 갖는다 — **뮤턴트 판정 81건 + seam 부재 3건**,
   판정 공백 0건. **측정만 했고 수리하지 않았다** — 생존 47 · 검출 34 · seam 부재 3.

   > **정정 (sync-audit F1·F2)**: 최초 작성본은 "미측정 0건 · 생존 49 · seam 부재 1" 이라고
   > 적었다. 그 진술은 **거짓이었다.** 씨앗 5a 는 호출자가 없는 죽은 함수에 들어가 아무것도
   > 재지 못했고(F1), 씨앗 4b 의 merge 근거는 패키지 경계를 넘지 못했다(F2). 도달하지 못한
   > 뮤턴트의 `ok` 는 생존이 아니라 **무판정**이다. 75·76행은 감사자가 살아 있는 seam
   > `:122`·`:158` 에서 도달을 증명한 뒤 재측정해 같은 결론(생존)을 세웠고, 73·74행은
   > seam 부재로 재분류했다. 상세: §2 M3 표 · §4 Gap 1b/1c.
4. **전역**: 모든 `-run` 인용에 `-list` 건수가 붙어 있고 0건 셀렉터 근거는 없다.
   뮤턴트 잔재 0. 부재 주장은 전부 `/usr/bin/grep`. 커밋 제목 전부에 `t539`.

주장하지 **않는** 것: 84건 각각이 자기 자신의 뮤턴트를 가졌다는 것(§4 Gaps 참조).

---

## 2. Evidence — 명령과 그 축자 출력

### M1 (로그: `m1-rerun.log`)

```
$ go run .moai/reports/t539/ctxsweep/main.go internal > /tmp/t539-sweep-test-rerun.tsv
exit=0 · stderr 무출력 · wc -l = 442

$ cmp /tmp/t539-sweep-test-rerun.tsv .moai/reports/t539/sweep-test.tsv
cmp_exit=0        ← 바이트 동일

$ awk -F'\t' '{c[$1"/"$6]++} END {for (k in c) print k, c[k]}' /tmp/t539-sweep-test-rerun.tsv | sort
funclit/ 196 · funclit/no 52 · funclit/yes 12
method/ 127  · method/no 26  · method/yes 29

$ /usr/bin/grep -rl ctxsweep internal/
grep_exit=1       ← 무출력 = 스캐너 부재
```

### M2 (로그: `m2-mutant.log`)

뮤턴트: `internal/cli/mcp_glm.go:305`
`http.NewRequestWithContext(ctx, …)` → `http.NewRequest(…)`

**창 1 — 생존 (AC-CBD-004, RED-now)**  가드 없음:
```
$ go test ./internal/cli/ -run GLM     -count=1 -timeout 1800s → ok  ...internal/cli  11.610s   (-list 223)
$ go test ./internal/cli/ -run Audit   -count=1 -timeout 1800s → ok  ...internal/cli  31.061s   (-list  99)
$ go test ./internal/cli/ -run Converg -count=1 -timeout 1800s → ok  ...internal/cli   1.314s   (-list  32)
$ git diff --stat  (되돌림 후) → 무출력
```

**창 2 — 사망 (AC-CBD-005, E8 축자 RED)**  가드 있음, 같은 뮤턴트:
```
$ go test ./internal/cli/ -run TestGLMAudit_CancelledContext_IsNotSwallowed -count=1 -timeout 1800s
--- FAIL: TestGLMAudit_CancelledContext_IsNotSwallowed (0.45s)
    mcp_glm_audit_ctx_test.go:89: a cancelled audit returned the canned success verdict "pass" — the request reached the transport detached from its context (check http.NewRequestWithContext at mcp_glm.go:305)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.468s

$ go test ./internal/cli/ -run GLM   -count=1 -timeout 1800s → FAIL (같은 테스트)  9.024s
$ go test ./internal/cli/ -run Audit -count=1 -timeout 1800s → FAIL (같은 테스트) 20.540s
$ git diff --stat  (되돌림 후) → 무출력
```

**비회귀 (AC-CBD-006/007)**  깨끗한 트리:
```
$ go test ./internal/cli/ -run GLM         -count=1 -timeout 1800s → ok   8.837s   (-list 225 ≥ 223)
$ go test ./internal/cli/ -run Audit       -count=1 -timeout 1800s → ok  20.907s   (-list 101 ≥  99)
$ go test ./internal/cli/ -run Converg     -count=1 -timeout 1800s → ok   0.980s   (-list  32 =  32)
$ go test ./internal/cli/ -run TestGLMTask -count=1 -timeout 1800s → ok   1.276s   (-list  16)
$ go test ./internal/cli/ -run GLM -race   -count=1 -timeout 1800s → ok  11.937s
$ git diff dbc1f7125..HEAD -- internal/cli/mcp_glm_test.go | wc -l → 0   ← stubGLMDoer 눈먼 채 유지
$ git diff --stat -- internal/cli/glm_task_bg_context_test.go        → 무출력
```

`-list` 출력은 파일로 보존됐다 — 건수가 본문 숫자가 아니라 파일에서 재유도된다:
```
$ /usr/bin/grep -c '^Test' .moai/reports/t539/list-GLM.log      → 225
$ /usr/bin/grep -c '^Test' .moai/reports/t539/list-Audit.log    → 101
$ /usr/bin/grep -c '^Test' .moai/reports/t539/list-Converg.log  →  32
$ /usr/bin/grep -c '^Test' .moai/reports/t539/list-GLM-baseline.log     → 223
$ /usr/bin/grep -c '^Test' .moai/reports/t539/list-Audit-baseline.log   →  99
$ /usr/bin/grep -c '^Test' .moai/reports/t539/list-Converg-baseline.log →  32
```

### M3 (로그: `m3-update.log` · `m3-statusline.log` · `m3-lsp.log` · `m3-template-deployer.log` · `m3-gh-client.log` · 귀속표 `m3-attribution.md`)

닫힌 집합 확정:
```
$ /usr/bin/grep -E '^- `internal/[^`]+_test\.go:[0-9]+` `' .moai/specs/SPEC-CTX-BLIND-DOUBLE-001/plan.md | wc -l
84
```

| 후보군 | 씨앗(주입 지점) | `-list` | 결과 | 판정 |
|---|---|---:|---|---|
| 1 update | `update/orchestrator.go:55` `updater.Download(ctx→Background)` | 88 | `ok 6.334s` | **생존** |
| 1 update | (씨앗 없음 — `IsUpdateAvailable(string)` 에 ctx 파라미터 부재) | — | — | seam 부재 / 옳게 눈멂 |
| 2 statusline | `statusline/builder.go:362` `gitProvider.CollectGitStatus` | 318 | `ok 23.423s` | **생존** |
| 3a LSP agg | `lsp/aggregator/aggregator.go:175` `client.GetDiagnostics(qCtx→Background)` | 19 | `FAIL 0.495s` ×2 | **검출** |
| 3b LSP core | `lsp/core/manager.go:366` `c.Start(ctx→Background)` | 97 | `ok 0.305s` | **생존** |
| 3c LSP transport | `lsp/transport/request.go:53` `t.Call(ctx→Background)` | 32 | `FAIL 180.201s` (행) | **검출** |
| 4a template | `core/project/initializer.go:434` `deployer.Deploy` | 118 | `ok 1.609s` | **생존** |
| 4b template | `cli/mirror_notice.go:37,:42` `Deploy`/`DeployWithResult` | 161·32 | `ok` ×2 | **생존** (10행) |
| 4b′ merge | (seam 부재 — `merge.go:306` 에 ctx 파라미터 없음) | 51* | — | **측정 불가** (2행, F2) |
| ~~5a gh~~ | ~~`cli/branch_protection.go:65` `gh.Run`~~ | ~~5~~ | ~~`ok 0.862s`~~ | **무효 — 무도달 (F1)** |
| 5a′ gh | `cli/branch_protection.go:122` `gh.Run` (`PreflightGh`) | 2 | `ok 0.887s` (도달 증명 후) | **생존** |
| 5a″ gh | `cli/branch_protection.go:158` `gh.RunWithStdin` (`ApplyBranchProtection`) | 5 | `ok 0.688s` (도달 증명 후) | **생존** |
| 5b gh | `github/pr_reviewer.go:110` `gh.PRView` | 136 | `ok 0.244s` | **생존** |
| 5c gh | `guardstate/evaluate.go:180` `q.RunsForSubject` | 47 | `ok 0.228s` | **생존** |

\* 51 은 merge 패키지가 초록임을 확인한 수치일 뿐 73·74행에 대해 **아무것도 주장하지 않는다**
(그 수치를 씨앗 4b 의 근거로 나열한 것이 F2 의 결함이었다). 로그: `list-merge.log` · `m3-merge.log`.

**씨앗 5a 무효 판정의 근거 (F1 [High])** — 뮤턴트는 실행되지 않았다:
```
$ /usr/bin/grep -rn 'DiscoverOwnerRepo' --include='*.go' .
./internal/cli/branch_protection.go:62:   // 주석
./internal/cli/branch_protection.go:64:   func DiscoverOwnerRepo(...)
                                          ← 정의뿐, 호출자 0건

$ sed -i '' '65s|^\t|\tpanic("REACHPROBE"); |' internal/cli/branch_protection.go
$ go test ./internal/cli/ -run BranchProtection -count=1 -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/cli	0.846s     ← panic 미발생 = 미도달
```
**보정 재측정 — 도달성을 먼저 세우고 그 위에서 뮤턴트** (감사자 실행):
```
$ sed -i '' '122s|^\t|\tpanic("REACH122"); |' … ; go test -run PreflightGh
--- FAIL: TestPreflightGh_Authed (0.00s)
panic: REACH122 [recovered, repanicked]
	...cli.PreflightGh(...)  internal/cli/branch_protection.go:122          ← 도달
$ sed -i '' '158s|^\t|\tpanic("REACH158"); |' … ; go test -run BranchProtection
--- FAIL: TestApplyBranchProtection_Success (0.00s)
panic: REACH158 [recovered, repanicked]
	...cli.ApplyBranchProtection(...)  internal/cli/branch_protection.go:158 ← 도달

# 도달이 선 뒤 뮤턴트 (gh.Run / gh.RunWithStdin 의 ctx → context.Background())
$ go test ./internal/cli/ -run PreflightGh      -count=1 -timeout 600s → ok 0.887s  (-list 2)
$ go test ./internal/cli/ -run BranchProtection -count=1 -timeout 600s → ok 0.688s  (-list 5)
```
로그: `sync-audit-5a-reach-preflight.log` · `sync-audit-5a-reach158.log` ·
`sync-audit-5a-live-preflight.log` · `sync-audit-5a-live-apply.log` ·
`sync-audit-m3-5a.log` · `sync-audit-m3-5a-reach.log`.
결론(생존)은 같으나 **그것을 세운 증거가 달라졌다** — 원래 증거는 공허했다.

**F2 seam 부재의 근거** — package merge 안에 잴 대상이 없다:
```
$ /usr/bin/grep -rn 'Deploy\|Deployer' internal/cli/update/merge/*.go | /usr/bin/grep -v '_test.go'
internal/cli/update/merge/merge.go:306:func AnalyzeMergeChanges(deployer template.Deployer, projectRoot string) …
                                        ← ctx 파라미터 없음, ListTemplates() 만 부른다
$ /usr/bin/grep -n '\.Deploy(\|\.ValidateAll(' internal/cli/update/merge/merge.go internal/cli/update/merge/base.go
grep_exit=1     ← 프로덕션 호출 0건
$ /usr/bin/grep -n '\.Deploy(\|\.ValidateAll(' internal/cli/update/merge/*_test.go
grep_exit=1     ← 이 패키지 자신의 테스트에서도 0건
```

**도달성 프로브 전수** (감사자, `sync-audit-reach-*.log`) — 생존 씨앗 8자리 중 5a 만 미도달:
1 update `orchestrator.go:55` 도달 · 2 statusline `builder.go:362` 도달 ·
3b lsp/core `manager.go:366` 도달(×5) · 4a project `initializer.go:434` 도달 ·
4b `mirror_notice.go:42` 도달(`-run Mirror`) · `:37` 도달(`-run Update`) ·
5b github `pr_reviewer.go:110` 도달 · 5c guardstate `evaluate.go:180` 도달 ·
**5a `branch_protection.go:65` 미도달 → 무효**.

검출된 두 자리의 축자 출력:
```
# 3a
--- FAIL: TestGetDiagnostics_timeout_returnsEmptyWhenNoCache (0.20s)
    aggregator_test.go:553: timeout (no cache): got 1 diagnostics, want 0
--- FAIL: TestGetDiagnostics_timeout_returnsCachedOnSlowUpstream (0.30s)
    aggregator_test.go:530: got message "fresh", want "stale cached" (stale cache)

# 3c — 테스트가 반환하지 않아 패키지 타임아웃으로 패닉
...transport.CallWithTimeout(...)  request.go:53 +0x94
...transport_test.TestCallWithTimeout_DeadlineExceeded(...)  request_test.go:116 +0x90
FAIL	github.com/modu-ai/moai-adk/internal/lsp/transport	180.201s
```

### 전역

```
$ git diff --stat                          → 무출력 (뮤턴트 잔재 0, AC-CBD-012)
$ git status --short                       → 무출력
$ git diff --stat dbc1f7125..HEAD -- 'internal/***.go' ':(exclude)internal/**/*_test.go'
  → 무출력                                  ← 프로덕션 변경 0줄
$ git log --format=%s dbc1f7125..HEAD
  test(t539): M2 guard the GLM audit path against a dead context (SPEC-CTX-BLIND-DOUBLE-001)
  test(t539): M1 pin the discriminator and sweep artifacts (SPEC-CTX-BLIND-DOUBLE-001)
$ go build ./...                           → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
$ go vet ./internal/{cli,update,statusline,lsp,core/project,github,guardstate}/...  → exit 0
$ golangci-lint run --timeout=8m           → "0 issues."   (기준선도 "0 issues." → NEW 0건)
```

---

## 3. Baseline-attribution — 무엇에 대고 쟀는가

| 측정 | 기준 | 이 실행에서 관측 |
|---|---|---|
| 스윕 집계 | `sweep-test.tsv` (plan 단계 산출물) | `cmp` exit 0 — 바이트 동일 |
| `-list` 223/99/32 | plan 본문 숫자 (파일 없음, D8 간극) | `list-*-baseline.log` 로 **이 실행에서 파일화**, `grep -c '^Test'` 로 재유도 |
| M2 뮤턴트 생존 | plan 단계 `m2-run-*.log` (`ok` 한 줄) | 이 트리·이 HEAD 에서 3 셀렉터 재측정, 다시 생존 |
| 린트 NEW 판정 | `lint-baseline.log` (`0 issues.`, HEAD `dbc1f7125`) | `lint-final.log` (`0 issues.`) → NEW 0건 |
| 84건 열거 | `plan.md` §F M3 | 엄격 정규식 `wc -l` = 84 (이 실행에서 재계수) |

모든 수치는 이 트리(`dbc1f7125` → `a9a50254e`)에서 이번에 실행한 명령의 출력이다.
다른 트리·다른 시점의 수를 옮겨 적은 항목은 없다.

---

## 4. Gaps — 관측하지 **않은** 것

1. **84건 개별 뮤턴트는 측정하지 않았다.** 씨앗은 12개이고, 나머지 구성원은 같은 seam 뒤에
   있다는 이유로 **귀속**됐다. "84건 전부에 판정이 귀속된다"(AC-CBD-008 의 문언)는 참이지만,
   "84건 전부가 자기 뮤턴트를 가졌다"는 **거짓**이다. 상세는 `m3-attribution.md` §Gaps.

1b. **[F1 정정] 씨앗 5a 는 아무것도 재지 못했고, 그 사실을 run-phase 는 놓쳤다.**
   `branch_protection.go:65` 는 호출자 0건인 죽은 함수 안이라 `-run BranchProtection` 이
   그 줄을 실행할 수 없었다. 그런데도 `ok 0.862s` 를 **생존**으로 기록했다 — 도달하지 못한
   뮤턴트의 초록과 진짜 생존은 출력이 같아서(`ok`) 구별되지 않으며, 구별하는 유일한 관측은
   도달성 프로브인데 M3 절차에 그 단계가 없었다. `-list 5` 가 씨앗 중 최소값이었던 것이
   신호였으나 읽지 못했다. 감사자가 살아 있는 seam `:122`·`:158` 에서 도달을 증명하고
   재측정해 같은 결론(생존)을 세웠다. **결론은 유지되고 증거는 교체됐다.**

1c. **[F2 정정] 귀속표 73·74행은 씨앗 4b 에 귀속될 수 없었다.**
   `deployWithMirrorNotice` 는 package `cli` 의 비공개 함수이고 `merge_test.go` 는
   package `merge` 라, merge 테스트 바이너리는 그 seam 을 컴파일조차 하지 않는다. 근거로
   나열했던 `ok 0.321s (-list 51)` 는 그 뮤턴트에 대해 아무것도 재지 않았다. package merge
   안에 대체 seam 을 찾았으나 없다(`AnalyzeMergeChanges` 는 ctx 파라미터가 없고
   `.Deploy(`/`.ValidateAll(` 호출은 프로덕션·테스트 모두 0건) → **seam 부재 / 측정 불가**로
   재분류했다. 로그 `m3-merge.log`.

1d. **[절차 결함] M3 절차에 도달성 확인 단계가 없었다.** 1b 는 그 구멍의 결과다. 정정된 절차는
   "주입할 줄에 `panic` 프로브 → 같은 셀렉터가 실패해야 도달 → 되돌림 → 그 뒤에야 뮤턴트" 이며
   `m3-attribution.md` § 절차 정정과 `m3-gh-client.log` § 절차 정정에 [HARD] 로 적었다.
   AC-CBD-011 의 `-list > 0` 공허성 가드는 이 층을 지키지 못한다(후속 카드 요청 F8).
2. **후보군 2 의 형제 seam 2개**(`builder.go:373` updateProvider, `:388` usageProvider)는
   개별 측정하지 않았다. gitProvider seam 1건만 실측이다.
3. **후보군 5 의 `mockGHClient` 6 메서드 중 5개**(`PRCreate`/`PRMerge`/`PRChecks`/`Push`/
   `IsAuthenticated`)는 뮤턴트가 지나가지 않았다. `PRView` 만 실측이다.
4. **전체 스위트를 로컬에서 돌리지 않았다** (`go test ./...` 금지, CLAUDE.local.md §4·§6).
   전 패키지 판정은 CI 몫이다. 이 카드가 돌린 것은 건드린 패키지뿐이다.
5. ~~**루트 `internal/cli` 커버리지 미관측.**~~ **[F5 로 닫힘]** 감사자가 측정했다:
   `go test -cover ./internal/cli/ -count=1 -timeout 1800s` → `ok ... 552.001s coverage: 81.1%
   of statements` (`sync-audit-cover-cli.log`). 프로필 임계 85% 미달이나 **이 카드가 만든
   회귀가 아니다** — 프로덕션 문장 0줄 변경 + 테스트 순증이므로 분모가 같고
   `coverage(HEAD) ≥ coverage(dbc1f7125)` 가 기계적으로 성립한다. 임계를 올리는 일은 이 카드의
   범위가 아니며, 물려받은 baseline 을 이 카드의 결함으로 계상하지 않는다(`AGENTS.md` §1). (§6 E3)
6. **`-run 'GLM|Audit|Converg'` 합성 셀렉터 형태로는 돌리지 않았다.** 이 세션의 Bash 가드가
   따옴표 안 정규식 교대(alternation)를 거부해 세 셀렉터를 개별 실행했다. 세 실행의 합집합은
   합성 셀렉터의 상위집합이므로 판정은 보존되지만, **명령 문자열 자체는 AC 문언과 다르다.**
7. **plan §C 의 "`git diff --stat` 무출력" 전제가 진입 시점에 성립하지 않았다.**
   `progress.md` 에 오케스트레이터가 쓴 §Phase 4 Mode Selection 블록이 미커밋 상태였다.
   코드 트리는 깨끗했고, 해당 파일은 이 에이전트 소관 산출물이라 M1 커밋에 포함했다.

---

## 5. Residual-risk — 관측했는데도 틀릴 수 있는 것

1. **생존 49건은 "가려진 결함이 있다"가 아니라 "가려진다면 보이지 않는다"이다.** 뮤턴트 생존은
   테스트 집합의 감지력 부재를 세울 뿐, 그 자리에 실제 결함이 있다고 말하지 않는다.
   후속 카드는 이 구분을 지고 출발해야 한다.
2. **검출된 두 자리(3a·3c)의 CLOSE 근거는 "위층이 본다"이고, 그 위층이 나중에 사라질 수 있다.**
   aggregator 의 서킷 브레이커나 transport 의 `isContextErr` 가 제거·우회되면 (d) 축이 참으로
   뒤집히고 31 + 3 건이 다시 후보가 된다. CLOSE 는 현재 트리에 대한 판정이다.
3. **M2 가드는 `ctxAwareGLMDoer` 의 충실도에 의존한다.** 이 대역은 `req.Context().Err()` 만
   본다 — 실제 `*http.Client` 의 취소 의미론(전송 중 중단, 부분 응답)까지 모델링하지 않는다.
   더 미묘한 컨텍스트 결함은 이 가드도 통과할 수 있다.
4. **`stubGLMDoer` 를 눈먼 채로 둔 결정은 되돌릴 수 있는 부채다.** audit 경로의 컨텍스트 축은
   이제 덮였지만, 같은 스텁을 쓰는 30여 개 테스트는 여전히 죽은 요청과 산 요청을 구별하지
   못한다. 그 축의 새 결함은 이 카드가 세운 가드 **바깥**에서 생길 수 있다.
5. **3c 의 검출은 단언이 아니라 행(hang)으로 성립했다.** 180초 패키지 타임아웃이 없었다면
   판정이 나오지 않는다 — 이 검출은 CI 타임아웃 설정에 의존한다.

---

## 6. §E 자기 검증 (E1–E8)

### E1 — AC PASS/FAIL 매트릭스

| AC | 상태 | 검증 명령 | 관측 출력 |
|---|---|---|---|
| AC-CBD-001 | **regression-guard** (통과로 기록 안 함) | `/usr/bin/grep -n 판별식 research.md` | `92:## 4. 판별식 — 네 축` · `101:수리 대상 = (c) ∧ (d)` · `99:` (d) 근거 `codex_task.go:124` |
| AC-CBD-002 | **regression-guard** | `go run …/ctxsweep/main.go internal` + `cmp` + `awk` | `cmp_exit=0` · 127/26/29 · 196/52/12 · `grep -rl ctxsweep internal/` exit 1 |
| AC-CBD-003 | **regression-guard** | `/usr/bin/grep -n '옳게 눈멂\|teeGLMDoer' research.md` | `89:` goal 러너 참조 0건 · `182/183:` · `57:` 위임 예외 |
| AC-CBD-004 | PASS | 뮤턴트 주입 후 3 셀렉터 | `ok 11.610s` / `ok 31.061s` / `ok 1.314s` (생존) · `-list` 223/99/32 |
| AC-CBD-005 | PASS | 가드 후 같은 뮤턴트 | `--- FAIL: TestGLMAudit_CancelledContext_IsNotSwallowed (0.45s)` · 되돌림 후 `ok` |
| AC-CBD-006 | PASS | `-run TestGLMTask` | `ok 1.276s` · `-list 16` · 기존 가드 파일 diff 0 |
| AC-CBD-007 | PASS | 3 셀렉터 + `-list` 파일 보존 | `ok`×3 · 225/101/32 (감소 없음) · `mcp_glm_test.go` diff 0줄 |
| AC-CBD-008 | **PASS-WITH-DEBT** | 열거 계수 + 귀속표 대조 | `wc -l = 84` · **뮤턴트 판정 81 + seam 부재 3 = 84**, 판정 공백 0건 · `m3-attribution.md` · `m3-merge.log`. 최초의 "미측정 0건" 진술은 F1(무도달 씨앗)·F2(패키지 경계)로 정정됨 |
| AC-CBD-009 | PASS | 생존 후보 수리 diff 부재 + 후속 카드 요청 | 프로덕션 diff 0 · 후속 카드 요청 §7 에 6항목 |
| AC-CBD-010 | PASS | 검출 후보 수리 diff 부재 + 판정 기록 | 3a·3c CLOSE ("위층이 본다") · `m3-lsp.log` |
| AC-CBD-011 | PASS | 모든 `-run` 인용에 `-list` 동반 | 이 문서의 모든 셀렉터 인용에 건수 기재 · 0건 셀렉터 근거 0 |
| AC-CBD-012 | PASS | `git diff --stat` / `git status --short` | 둘 다 무출력 · 커밋 2건 어디에도 뮤턴트 없음 |
| AC-CBD-013 | PASS | 부재 주장의 근거 명령 | 전부 `/usr/bin/grep` 절대경로 (셸 맨 `grep` 근거 0건) |
| AC-CBD-014 | PASS | 채택 근거 형태 | M2 = 생존→수리→사망 삼단 · 커버리지 단독 채택 0건 |
| AC-CBD-015 | PASS | `git log --format=%s dbc1f7125..HEAD` | 커밋 제목 2건 모두 `t539` 포함 · 인용 로그 전부 실재 |

### E2 — 크로스 플랫폼 빌드
```
$ go build ./...                           → exit 0
$ GOOS=windows GOARCH=amd64 go build ./... → exit 0
```
주의: 크로스 빌드는 테스트 파일을 컴파일하지 않는다. 이 카드의 산출물이 테스트 파일뿐이므로
windows 빌드는 **이 변경에 대해 아무것도 주장하지 않는다** — 회귀 부재의 방증일 뿐이다.

### E3 — 커버리지 (기록만, 게이트 아님)

```
$ go test -cover ./internal/cli/... -count=1 -timeout 1800s
(로컬 예산 600s 초과 → 백그라운드 이관, exit code 0, FAIL 0건)

ok  .../internal/cli/update/deploy   6.456s  coverage: 91.6% of statements
ok  .../internal/cli/update/merge    5.886s  coverage: 92.1% of statements
ok  .../internal/cli/update/plan     6.700s  coverage: 95.0% of statements
ok  .../internal/cli/update/report   4.583s  coverage: 92.9% of statements
ok  .../internal/cli/wizard          7.768s  coverage: 92.0% of statements
ok  .../internal/cli/worktree        9.672s  coverage: 87.1% of statements
```

run-phase 는 루트 `internal/cli` 수치를 관측하지 못했다(백그라운드 출력이 꼬리만 보존).
**[F5] sync-audit 이 그 간극을 닫았다**:
```
$ go test -cover ./internal/cli/ -count=1 -timeout 1800s
ok  	github.com/modu-ai/moai-adk/internal/cli	552.001s	coverage: 81.1% of statements
   (sync-audit-cover-cli.log)
```
프로필 임계 85% 미달 → sync-audit 의 Craft 차원 FAIL(75/100). 다만 **이 카드가 만든 회귀가
아니다**: 프로덕션 문장 0줄 변경 + 테스트 124줄 순증이므로 분모가 동일하고
`coverage(HEAD) ≥ coverage(dbc1f7125)` 가 성립한다. 물려받은 baseline 을 카드의 결함으로
계상하지 않는다(`AGENTS.md` §1 baseline 귀속).

**커버리지는 여전히 이 카드의 채택 근거가 아니다**(AC-CBD-014, plan §G 첫 항목: "커버리지를
채택 증거로 쓰기" 가 명시적 안티패턴). 채택은 §2 의 뮤턴트 삼단이 판정했다. 이 수치는
기록이지 게이트가 아니다.

### E4 — 서브에이전트 경계 grep
```
$ /usr/bin/grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_glm_audit_ctx_test.go
grep_exit=1   ← 무출력
```
패키지 전역 스캔의 히트는 전부 `internal/cli/testdata/agent_lint/*.md` 픽스처(마크다운, Go 아님,
이 카드 이전부터 존재)다. 신규 코드 히트 0건.

### E5 — 린트 (NEW vs 기준선)
```
기준선 (HEAD dbc1f7125): lint-baseline.log → "0 issues."
최종   (HEAD a9a50254e): lint-final.log    → "0 issues."
NEW: 0건
```

### E6 — 커밋 / push 상태
```
6fe7a4141  test(t539): M1 pin the discriminator and sweep artifacts (SPEC-CTX-BLIND-DOUBLE-001)
a9a50254e  test(t539): M2 guard the GLM audit path against a dead context (SPEC-CTX-BLIND-DOUBLE-001)
f83ed0c04  test(t539): M3 measure the five candidate groups with mutants, repair none (SPEC-CTX-BLIND-DOUBLE-001)
73a588054  (오케스트레이터 검증 배치)
+ sync-audit F1~F4 정정 커밋 (§8)

$ git log --format=%s dbc1f7125..HEAD | /usr/bin/grep -cv 't539'   → 0   (누락 0건)
```
**push 없음.** 레인 규율상 `WT-*` 브랜치는 push 하지 않으며 통합은 리드 소관이다.

### E7 — Blocker
없음. 운영자 결정 D0("측정만 이 카드, 수리는 후속 카드")이 plan 단계에서 이미 내려져 있어
M3 의 생존 뮤턴트가 재위임 사유가 되지 않았다.

### E8 — 축자 RED 출력 (GREEN 이전)
```
--- FAIL: TestGLMAudit_CancelledContext_IsNotSwallowed (0.45s)
    mcp_glm_audit_ctx_test.go:89: a cancelled audit returned the canned success verdict "pass" — the request reached the transport detached from its context (check http.NewRequestWithContext at mcp_glm.go:305)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.468s
```
이 카드의 RED 는 통상적 TDD 의 "구현 이전"이 아니라 **뮤턴트 주입 상태**에서 성립한다.
프로덕션은 처음부터 옳았고 없던 것은 가드이므로, 가드의 실패를 볼 수 있는 유일한 트리가
뮤턴트 트리다. 창 1 의 생존(`ok`×3)이 "가드 부재" 관측이고, 위 FAIL 이 "가드 작동" 관측이다.

---

## 7. 후속 카드 요청 (AC-CBD-009)

생존한 자리마다 한 항목. **이 카드는 어느 것도 수리하지 않았다.**

| # | 대상 패키지 | 생존 뮤턴트 주입 지점 | 구성원 | 로그 |
|---|---|---|---|---|
| F1 | `internal/update` | `update/orchestrator.go:55` `o.updater.Download(ctx, info)` | 3 | `m3-update.log` |
| F2 | `internal/statusline` | `statusline/builder.go:362` `b.gitProvider.CollectGitStatus(ctx)` (+ 미측정 형제 seam `:373` `:388`) | 4 | `m3-statusline.log` |
| F3 | `internal/lsp/core` | `lsp/core/manager.go:366` `c.Start(ctx)` | 12 | `m3-lsp.log` |
| F4 | `internal/core/project` + `internal/cli` | `core/project/initializer.go:434` · `cli/mirror_notice.go:37,:42` | 18 | `m3-template-deployer.log` |
| F5 | `internal/cli` + `internal/github` + `internal/guardstate` | `cli/branch_protection.go:122` (`PreflightGh`) · `:158` (`ApplyBranchProtection`) · `github/pr_reviewer.go:110` · `guardstate/evaluate.go:180` | 10 | `m3-gh-client.log` · `sync-audit-5a-live-*.log` |
| F6 | (설계 관찰, 대역 아님) | `internal/update/checker.go:325` `IsUpdateAvailable(current string)` 가 컨텍스트를 받지 않고 내부에서 `CheckLatest(context.Background())` 를 부른다 — 호출자의 컨텍스트가 이 경로에 도달할 수 없다 | — | `m3-update.log` 씨앗 B |
| **F7** | `internal/cli` (죽은 코드) | `branch_protection.go:64` `DiscoverOwnerRepo` 는 저장소 전체에 **호출자가 0건**인 공개 함수다. 공개 심볼이라 컴파일러가 잡지 않는다. `/moai clean` 또는 별도 카드 대상 | — | `sync-audit.md` F9 · `m3-gh-client.log` 씨앗 5a |
| **F8** | (절차/AC 개선) | 뮤턴트 절차에 **도달성 전제**를 AC 문언으로 넣는다 — *"생존 판정은 같은 줄에 `panic` 프로브를 넣었을 때 같은 셀렉터가 실패함을 보인 뒤에만 채택한다."* AC-CBD-011 의 `-list > 0` 은 이 층을 지키지 못한다(씨앗 5a 가 `-list 5` 로 문언을 충족하며 아무것도 재지 못했다) | — | `sync-audit.md` F1·F6 |
| **F9** | `internal/cli/update/merge` (측정 불가 2건) | 귀속표 73·74행(`mockDeployer.Deploy`/`.ValidateAll`)은 package merge 안에 소비 seam 이 없어 이 카드에서 잴 수 없었다. 이 두 대역이 지켜야 할 컨텍스트 축이 **있어야 하는지**(즉 `AnalyzeMergeChanges` 가 ctx 를 받아야 하는지)는 설계 질문이며 별도 카드 소관 | 2 | `m3-merge.log` |

**생존 47 의 해석 — 도달성 구분(리드 요청, 2026-09-08).** 생존 47 은 전부 **도달이 증명된 씨앗**의
판정이다: 정정 후 씨앗 12 개 모두 같은 셀렉터에서 `panic` 프로브가 발화한 뒤 변이했고
(`m3-attribution.md` 씨앗별 프로브 기록, 5a′/5a″ 는 `sync-audit-5a-reach-*.log`), 무도달로
드러난 5a 는 집계에서 **빠졌다**(무판정, 생존 아님). 따라서 47 은 "실행조차 안 된 것을 포함한
상한"이 아니라 "실행됐고 살아남은 것"의 수다. 단, 47 은 **구성원** 수이고 뮤턴트를 직접 맞은 것은
**씨앗 12 자리**다 — 구성원은 seam 귀속으로 판정을 물려받는다(§4 Gaps). 실행되지 않아 판정이 없는
자리는 seam 부재 3 행뿐이며 별도로 센다.

각 후속 카드의 권장 진입 형태: 이 카드의 M2 를 그대로 반복하되, **0단계로 도달성 프로브를
먼저 돌린다**(F8 이 강제하는 단계 — 이 카드의 씨앗 5a 가 그것 없이 무판정을 생존으로 적었다).
도달이 선 뒤 해당 seam 에 뮤턴트를 넣어 **생존을 재확인**하고, 충실한 대역으로 가드를 세운 뒤,
같은 뮤턴트가 **죽는 것**을 보인다.
F2 는 형제 seam 2개(`:373` `:388`)의 개별 측정부터 시작한다(이 카드는 대표성을 측정하지 않았다).
F5 는 `mockGHClient` 6 메서드 중 `PRView` 만 실측이므로 나머지 5개의 개별 측정이 남아 있다.

---

## 8. 커밋

| SHA | 제목 |
|---|---|
| `6fe7a4141` | `test(t539): M1 pin the discriminator and sweep artifacts (SPEC-CTX-BLIND-DOUBLE-001)` |
| `a9a50254e` | `test(t539): M2 guard the GLM audit path against a dead context (SPEC-CTX-BLIND-DOUBLE-001)` |
| `f83ed0c04` | `test(t539): M3 measure the five candidate groups with mutants, repair none (SPEC-CTX-BLIND-DOUBLE-001)` |
| `73a588054` | 오케스트레이터 검증 배치 (run-phase 증거 재확인) |
| (이 커밋) | `docs(spec): t539 sync-audit F1–F4 repairs — reachability-corrected gh seam, merge rows measured, line/SHA fixes (SPEC-CTX-BLIND-DOUBLE-001)` |

run-phase 종단 커밋은 `f83ed0c04` 다(`progress.md` §E.3 `run_commit_sha`). `73a588054` 는
그 뒤의 검증 배치이고, 이 커밋은 sync-audit 정정이다.

push 하지 않았다.

# sync-audit 판정 — SPEC-CTX-BLIND-DOUBLE-001 (카드 t539)

> 트리 `.claude/worktrees/t539` · 브랜치 `WT-ctx-blind-double` · HEAD `73a588054` · 기준 `dbc1f7125`
> 평가 프로필 `default` (`.moai/config/evaluator-profiles/default.md`), 평면 가중 백분율 방식
> (`harness.yaml` 에 `evaluator_mode: hierarchical` 없음). must-pass = Functionality + Security.
> 감사자는 run-phase 기록을 읽기만 하지 않았다 — 아래 판정의 근거는 **이 감사에서 직접 재실행한
> 명령의 출력**이며, 재실행하지 않은 항목은 Gaps 로 분리했다.

---

## 종합 판정

**PASS-WITH-DEBT · 83.8 / 100**

M2 가드의 채택 근거인 뮤턴트 삼단(생존 → 수리 → 사망)을 감사자가 처음부터 끝까지 재현했고,
그중 run-phase 가 로그 순서로만 주장할 수 있었던 **생존 단계까지 기계적으로 다시 세웠다.**
프로덕션 변경 0줄, 위생 무결, 추적성 성립.

그러나 M3 의 씨앗 11개 중 **1.5개가 아무것도 재지 못한 뮤턴트**였다. 씨앗 5a 는 호출자가
저장소 전체에 하나도 없는 죽은 함수에 들어갔고, 씨앗 4b 의 세 갈래 근거 중 하나는 패키지 경계를
넘지 못하는 측정이었다. AC-CBD-008 의 문언("미측정 0건")은 실질에서 **2건 거짓**이다.
이 SPEC 이 §4 판별식에서 스스로 "(d) 는 판독으로 추정하되 판정은 뮤턴트로 확정한다" 고 못박은
바로 그 축에서, 판정을 내리지 못하는 뮤턴트가 판정으로 기록됐다.

부채가 남지만 산출물을 되돌릴 사유는 아니다. 가드는 건전하고, 84건 중 80건의 귀속은 건전하며,
어긋난 2건은 감사자가 살아 있는 seam 에서 다시 재어 같은 결론을 세웠고, 남은 2건은 기록 정정과
1회 재측정으로 닫힌다. 그 4건을 **blocking 정정 항목**으로 아래에 명시한다.

---

## 차원 점수

| 차원 | 점수 | 판정 | 증거 (이 감사에서 실행한 명령의 축자 출력) |
|---|---:|---|---|
| Functionality (40%) | 80/100 | PASS-WITH-DEBT | `go test ./internal/cli/ -run TestGLMAudit_ -count=1 -v -timeout 600s` → 가드 제거 + 뮤턴트: `ok 6.777s`/`ok 16.942s`/`ok 0.885s` (생존) · 가드 복원 + 뮤턴트: `--- FAIL: TestGLMAudit_CancelledContext_IsNotSwallowed (0.36s)` 단 1건, 거울상은 `--- PASS` · 되돌림 후 `git diff --stat` 무출력. 열거 대조: `wc -l = 84`, 귀속표와 **양방향 차집합 0**. 결함: 씨앗 5a 무도달(F1), 4b merge 귀속 무효(F2) |
| Security (25%) | 95/100 | PASS | `/usr/bin/grep -n 'AskUserQuestion\|mcp__askuser' internal/cli/mcp_glm_audit_ctx_test.go` → `grep_exit=1` (무출력) · 비밀값 스캔 `grep -nE 'sk-\|api[_-]?key *= *"[^"]{16,}\|token *= *"[^"]{16,}\|Bearer [A-Za-z0-9]{16,}'` → `grep_exit=1` · 실제 키 리터럴은 `"test-glm-key"` 2회(가짜) · `git diff --name-only dbc1f7125..HEAD -- '*.go'` → `internal/cli/mcp_glm_audit_ctx_test.go` 단 1건(테스트 전용, 프로덕션 표면 증가 0) |
| Craft (20%) | 75/100 | FAIL (임계) | `go test -cover ./internal/cli/ -count=1 -timeout 1800s` → `ok ... 552.001s coverage: 81.1% of statements` — 프로필 임계 85% 미달. `golangci-lint run --timeout=10m` → `0 issues.` · `gofmt -l` 무출력 · `go vet` 7개 패키지 exit 0 |
| Consistency (15%) | 92/100 | PASS | `git log --format=%s dbc1f7125..HEAD` 4건 전부 `t539` 포함(누락 계수 `0`) · `git status --short` 추적 파일 변경 0 · 인용 증거 경로 전수 resolve · `git diff dbc1f7125..HEAD -- internal/cli/mcp_glm_test.go \| wc -l` → `0`, `glm_task_bg_context_test.go` → `0` |

**가중 조화평균** = 1 / (0.40/80 + 0.25/95 + 0.20/75 + 0.15/92)
= 1 / (0.005000 + 0.002632 + 0.002667 + 0.001630) = 1 / 0.011929 = **83.83**

### must-pass 방화벽

| 차원 | must-pass | 임계 | 관측 | 결과 |
|---|---|---|---:|---|
| Functionality | **예** | 모든 AC 충족 | 80 | **통과** (AC-CBD-008 은 문언 충족 · 실질 2/84 부채 — 아래 판독 근거) |
| Security | **예** | Critical/High 0건 | 95 | **통과** (어느 등급의 발견도 0건) |
| Craft | 아니오 | 커버리지 ≥ 85% | 75 | **미달** — 그러나 must-pass 아님. 프로필이 전체 FAIL 로 승격시키는 차원은 Security 하나뿐("Security FAIL = Overall FAIL")이며 Craft 는 명시적으로 제외된다 |
| Consistency | 아니오 | 주요 패턴 위반 없음 | 92 | 통과 |

**Craft 미달을 이 카드에 귀속하지 않는 이유(baseline 무결성)**: 81.1% 는 `internal/cli` 패키지
전체의 성질이고, 이 카드는 프로덕션 문장을 **0줄** 바꿨다(`git diff --name-only dbc1f7125..HEAD -- '*.go'`
→ 테스트 파일 1개, 124줄 삽입 · 0줄 삭제). 분모가 동일하고 테스트는 순증했으므로
`coverage(HEAD) ≥ coverage(dbc1f7125)` 가 기계적으로 성립한다 — 즉 임계 미달은 **이 카드 이전부터
있던 조건**이며 이 카드가 만든 회귀가 아니다. 물려받은 baseline 을 카드의 결함으로 계상하는 것은
`AGENTS.md` §1 의 baseline 귀속 위반이므로 하지 않는다. 다만 프로필의 하드 임계는 문언 그대로
적용해 차원 판정을 FAIL 로 적는다.

### AC-CBD-008 에 어느 판독을 적용했는가 (요구된 명시)

AC 문언은 **귀속**을 요구한다 — "84건 전부가 어느 후보군의 뮤턴트 판정 아래 귀속되어 미측정으로
남은 항목이 0건". 구성원 단위 개별 뮤턴트를 요구하지 않는다. run-phase 는 이 구분을 스스로
드러냈다(`verdict.md` §4 Gap 1, `m3-attribution.md` §Gaps). 따라서 "84건 전부가 자기 뮤턴트를
가졌는가" 로 읽어 FAIL 을 매기지 않는다 — **귀속 판독**을 적용한다.

그러나 귀속은 "판정에 귀속"이지 "아무것에나 귀속"이 아니다. 뮤턴트가 **도달하지 못한 자리**는
생존도 사망도 아닌 **무판정**이고, 무판정에 귀속된 행은 미측정 행이다. 이 좁은 의미에서
AC-CBD-008 은 4행(75·76·73·74)에서 어긋났고, 그중 75·76 은 감사자가 살아 있는 seam 에서 다시 재어
같은 결론(생존)을 세웠으므로 실질 잔여는 **2행**이다. 84 중 2 = 2.4%.

---

## 발견 사항

> 발견 단계는 필터링하지 않는다 — 확신이 낮거나 경미한 것도 전부 적는다(프로필 § Finding-Stage Reporting).
> `blocking` = 정확성 또는 SPEC 이 실제로 기술한 요구를 건드리는 것. `optional` = 그 밖의 전부.

### F1 [High] [blocking] · 확신 100% — 씨앗 5a 는 죽은 함수에 들어갔다 (공허한 생존)

`internal/cli/branch_protection.go:65` (`DiscoverOwnerRepo` 내부).

이 함수는 **저장소 전체에 호출자가 없다**. 따라서 `-run BranchProtection` 이 이 줄을 실행할 수
없고, 뮤턴트가 살아남은 것은 대역이 눈멀어서가 아니라 **뮤턴트가 실행되지 않아서**다.
`-list 5` 는 M3 씨앗 중 최소값이었고, 그 자체가 신호였다.

재현:
```bash
/usr/bin/grep -rn 'DiscoverOwnerRepo' --include='*.go' .
#  → 정의 2줄(branch_protection.go:62 주석, :64 시그니처)뿐. 호출자 0건.

sed -i '' '65s|^\t|\tpanic("REACHPROBE"); |' internal/cli/branch_protection.go
go test ./internal/cli/ -run BranchProtection -count=1 -timeout 600s
#  → ok  github.com/modu-ai/moai-adk/internal/cli  0.846s   ← panic 미발생 = 미도달
git checkout -- internal/cli/branch_protection.go
```

**감사자가 수행한 보정 재측정** — `stubGhClient` 는 살아 있는 seam 두 곳으로 실제 사용된다:
```bash
# 도달성 먼저 (개별 프로브 — :122 가 :158 을 가리므로 분리 실행)
sed -i '' '122s|^\t|\tpanic("REACH122"); |' internal/cli/branch_protection.go
go test ./internal/cli/ -run PreflightGh -count=1 -timeout 600s   # → REACH122 발생, FAIL
git checkout -- internal/cli/branch_protection.go
sed -i '' '158s|^\t|\tpanic("REACH158"); |' internal/cli/branch_protection.go
go test ./internal/cli/ -run BranchProtection -count=1 -timeout 600s  # → REACH158 발생, FAIL
git checkout -- internal/cli/branch_protection.go

# 도달성이 선 뒤 뮤턴트
sed -i '' '122s|gh.Run(ctx, "auth", "status")|gh.Run(context.Background(), "auth", "status")|' internal/cli/branch_protection.go
sed -i '' '158s|gh.RunWithStdin(ctx, json|gh.RunWithStdin(context.Background(), json|' internal/cli/branch_protection.go
go test ./internal/cli/ -run PreflightGh       -count=1 -timeout 600s  # → ok 0.887s  (-list 2)
go test ./internal/cli/ -run BranchProtection  -count=1 -timeout 600s  # → ok 0.688s  (-list 5)
git checkout -- internal/cli/branch_protection.go
```
두 seam 모두 도달 확인 후 **생존**. 즉 75·76 행의 *결론*은 옳다 — 옳지 않았던 것은 그 결론을 세운
증거다. 로그: `sync-audit-5a-live-preflight.log` · `sync-audit-5a-live-apply.log` ·
`sync-audit-5a-reach-preflight.log` · `sync-audit-5a-reach158.log` · `sync-audit-m3-5a.log` ·
`sync-audit-m3-5a-reach.log`.

**필요한 수정**: `m3-gh-client.log` 씨앗 5a 절, `verdict.md` §2 M3 표·§7 F5, `m3-attribution.md`
75~76행, `progress.md` §E.2 M3 표의 주입 지점을 `branch_protection.go:65` 에서
`:122`(`PreflightGh`) 및 `:158`(`ApplyBranchProtection`) 으로 교체하고, 위 보정 재측정의 출력과
도달성 프로브를 근거로 인용한다. `DiscoverOwnerRepo` 가 죽은 코드라는 사실은 별개의 관찰이므로
후속 카드 요청에 부기한다.

### F2 [Medium] [blocking] · 확신 100% — 씨앗 4b 의 merge 패키지 귀속은 패키지 경계를 넘지 못한다

`m3-attribution.md` 73~74행: `internal/cli/update/merge/merge_test.go:621` · `:626`
(`mockDeployer.Deploy` / `.ValidateAll`) 이 씨앗 4b (`internal/cli/mirror_notice.go:37,:42`) 에 귀속돼 있다.

`deployWithMirrorNotice` 는 **package `cli` 의 비공개 함수**이고 `merge_test.go` 는 **package `merge`**
다. merge 테스트 바이너리는 그 함수를 컴파일하지도 실행하지도 못한다. 따라서
`m3-template-deployer.log` 가 근거로 든 `go test ./internal/cli/update/merge/ → ok 0.321s (-list 51)`
는 그 뮤턴트에 대해 **아무것도 재지 않았다**.

재현:
```bash
head -1 internal/cli/mirror_notice.go              # → package cli
head -1 internal/cli/update/merge/merge_test.go    # → package merge
/usr/bin/grep -rn 'deployWithMirrorNotice' --include='*.go' internal/
#  → 정의 1 + 호출 2 (update_clean_install.go:455, update_template_sync.go:341) — 전부 package cli
```

씨앗 4b 의 나머지 근거는 건전하다 — `:37` 과 `:42` 는 둘 다 도달성이 확인됐다:
```bash
sed -i '' '42s|^\t|\tpanic("REACHPROBE42"); |' internal/cli/mirror_notice.go
go test ./internal/cli/ -run Mirror -count=1 -timeout 600s   # → REACHPROBE42 발생 (FAIL)
git checkout -- internal/cli/mirror_notice.go
sed -i '' '37s|^\t\t|\t\tpanic("REACHPROBE37"); |' internal/cli/mirror_notice.go
go test ./internal/cli/ -run Update -count=1 -timeout 900s   # → REACHPROBE37 발생 (FAIL)
git checkout -- internal/cli/mirror_notice.go
```
로그: `sync-audit-reach-4b-mirror.log` · `sync-audit-reach-4b-update.log`.
따라서 후보군 4 의 20건 중 **18건은 건전하고 2건(73·74)만 무효**다.

**필요한 수정**: 73~74행을 (a) `internal/cli/update/merge` 안의 실제 소비 seam 에 뮤턴트를 넣어
재측정하거나, (b) **미측정으로 재분류**하고 후속 카드 요청에 한 항목으로 올린다. 어느 쪽이든
`verdict.md` §1·§2 와 `m3-attribution.md` 머리말의 **"미측정 0건" 문구를 정정**해야 한다.

### F3 [Low] [blocking] · 확신 100% — 씨앗 3a 의 인용 줄이 1 어긋난다

`internal/lsp/aggregator/aggregator.go:174` 로 인용됐으나 실제 seam 은 **175행**이다
(174행은 `var getErr error`).

```bash
sed -n '174p;175p' internal/lsp/aggregator/aggregator.go
# 174:			var getErr error
# 175:			diags, getErr = client.GetDiagnostics(qCtx, path)
```

인용된 before/after 텍스트가 유일해 재현 자체는 막히지 않는다 — 감사자는 175행에서 재현했고
로그가 인용한 두 실패 메시지가 **축자 일치**했다:
```bash
sed -i '' '175s|client.GetDiagnostics(qCtx, path)|client.GetDiagnostics(context.Background(), path)|' internal/lsp/aggregator/aggregator.go
go test ./internal/lsp/aggregator/... -count=1 -timeout 180s
# --- FAIL: TestGetDiagnostics_timeout_returnsEmptyWhenNoCache (0.20s)
#     aggregator_test.go:553: timeout (no cache): got 1 diagnostics, want 0
# --- FAIL: TestGetDiagnostics_timeout_returnsCachedOnSlowUpstream (0.30s)
#     aggregator_test.go:530: got message "fresh", want "stale cached" (stale cache)
git checkout -- internal/lsp/aggregator/aggregator.go
```
로그: `sync-audit-m3-3a.log`.

**필요한 수정**: `174` → `175` 를 네 곳에서 고친다 — `m3-lsp.log:5`, `verdict.md:108`,
`m3-attribution.md:58`, `progress.md:72`.

### F4 [Low] [blocking] · 확신 100% — 판정 문서의 커밋 표가 낡았다

`verdict.md` §6 E6 은 커밋 2건만 적고 M3 를 `(+ M3 커밋 — 아래 §8)` 로 넘기며, §8 표의 M3 행은
SHA 자리가 `(M3)` 로 비어 있다. 4번째 커밋 `73a588054` 는 어디에도 없다. `progress.md` §E.3 의
`run_commit_sha: pending-backfill-run` 도 미기입이다.

AC-CBD-015 의 실질은 성립한다(감사자 재확인: 4건 전부 `t539` 포함, 누락 계수 `0`):
```bash
git log --format=%s dbc1f7125..HEAD | /usr/bin/grep -cv 't539'   # → 0
```
성립하지 않는 것은 **문서 자신의 커밋 표**다.

**필요한 수정**: `verdict.md` §8 에 `f83ed0c04` (M3) 와 `73a588054` (오케스트레이터 검증 배치) 를
채우고, `progress.md` §E.3 `run_commit_sha` 를 `f83ed0c04` 로 확정한다.

### F5 [Low] [optional] · 확신 100% — E3 커버리지 간극을 감사자가 닫았다: 81.1%

run-phase 는 루트 `internal/cli` 수치를 관측하지 못했다고 Gap 5 로 남겼다. 감사자가 측정했다:
```bash
go test -cover ./internal/cli/ -count=1 -timeout 1800s
# ok  github.com/modu-ai/moai-adk/internal/cli  552.001s  coverage: 81.1% of statements
```
로그: `sync-audit-cover-cli.log`. 프로필 임계 85% 미달이나 위 § must-pass 방화벽에 적은 대로
이 카드가 만든 회귀가 아니다(프로덕션 문장 0줄 변경, 테스트 순증).

**권장(선택)**: `verdict.md` §6 E3 과 §4 Gap 5 를 "미관측"에서 "81.1% (sync-audit 에서 측정)" 로
갱신한다. 수치를 올리는 일은 이 카드의 범위가 아니다.

### F6 [Low] [optional] · 확신 90% — AC-CBD-011 은 도달성을 지키지 않는다

AC-CBD-011 은 `-list > 0` 만 요구한다. 씨앗 5a 는 `-list 5` 로 이 AC 를 **문언대로 충족하면서**
아무것도 재지 못했다(F1). 공허성 가드가 한 층 아래에서 뚫린다.

**권장(선택)**: 후속 SPEC 에 뮤턴트 절차의 전제 조건을 한 줄 추가한다 — *"생존 판정은 같은 줄에
`panic` 프로브를 넣었을 때 같은 셀렉터가 실패하는 것을 보인 뒤에만 채택한다."* 이 카드가 §4 에서
"(d) 는 판독으로 추정하되 판정은 뮤턴트로 확정한다" 고 세운 원칙의 자연스러운 다음 칸이다.

### F7 [Info] [optional] · 확신 100% — AC-CBD-007 문언 이탈은 판정을 보존한다

AC 는 `-run 'GLM|Audit|Converg'` 를 지명하나 실행은 세 번의 개별 호출이었다(셸 가드가 따옴표 안
정규식 교대를 거부). Go 의 `-run` 은 최상위 테스트명에 정규식을 앵커 없이 적용하므로 교대식이
매칭하는 집합은 세 셀렉터가 매칭하는 집합들의 **합집합과 정확히 같다** — 상위집합이 아니라 동일.
감사자가 세 셀렉터를 전부 재실행해 같은 결과를 얻었다.

**판정: PASS-with-note.** 정정 불필요. `verdict.md` §4 Gap 6 이 이미 이탈을 스스로 적어 뒀다.

### F8 [Info] [optional] · 확신 70% — `list-TestGLMTask.log` 는 본문에서 경로로 인용되지 않는다

`verdict.md` 는 `-list 16` 이라는 수치만 적고 그 수치가 재유도되는 파일 경로
(`.moai/reports/t539/list-TestGLMTask.log`, 실재)는 다른 여섯 로그와 달리 명시하지 않는다.
다른 셀렉터에는 전부 붙어 있으므로 일관성 흠결이다. 감사자가 재유도했다:
`/usr/bin/grep -c '^Test' .moai/reports/t539/list-TestGLMTask.log` → `16`.

### F9 [Info] [optional] · 확신 100% — `DiscoverOwnerRepo` 는 죽은 공개 코드다

`internal/cli/branch_protection.go:64`. F1 의 프로브가 부수적으로 드러냈다. 공개 심볼이라
컴파일러가 잡지 않는다. `/moai clean` 또는 별도 후속 카드의 대상.

---

## 5절 증거

### Claim — 이 감사가 주장하는 것

1. M2 가드의 채택 근거인 뮤턴트 삼단이 이 트리에서 **전 단계 재현**된다. 특히 run-phase 가 로그
   순서로만 주장할 수 있었던 **생존 단계를 기계적으로 다시 세웠다.**
2. 프로덕션 변경은 0줄이며, 그 주장을 세운 pathspec 이 공허하지 않다(양성 대조 통과).
3. `plan.md` §F 의 84건 열거는 닫혀 있고 귀속표와 정확히 일치하며, 84건 전부가 `sweep-test.tsv` 의
   진짜 후보 행(method · consults ≠ yes)이다.
4. M3 씨앗 11개 중 **9개는 도달성이 증명됐고 1개(5a)는 무도달, 1개(4b)는 세 근거 중 하나가 무효**다.
5. 빌드·크로스빌드·vet·lint·gofmt 는 이 감사에서 재실행해 전부 깨끗하다.
6. 루트 `internal/cli` 커버리지는 81.1% 이며, 이는 이 카드가 만든 회귀가 아니다.

### Evidence — 명령과 그 축자 출력

**M2 삼단 전 구간 재현** (`sync-audit-m2-clean.log` · `sync-audit-survive-{GLM,Audit,Converg}.log` · `sync-audit-m2-mutant.log`)
```
# 1) 깨끗한 트리
$ go test ./internal/cli/ -run TestGLMAudit_ -count=1 -v -timeout 600s
ok  	github.com/modu-ai/moai-adk/internal/cli	4.004s        (-list TestGLMAudit_ = 12)

# 2) 생존 단계 재구성 — 가드 파일을 치우고 같은 뮤턴트 주입
$ mv internal/cli/mcp_glm_audit_ctx_test.go /tmp/t539-guard-backup.go
$ sed -i '' '305s|http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))|http.NewRequest(http.MethodPost, url, bytes.NewReader(body))|' internal/cli/mcp_glm.go
$ git status --short | grep -v '^??'
 M internal/cli/mcp_glm.go
 D internal/cli/mcp_glm_audit_ctx_test.go
$ go test ./internal/cli/ -list GLM   -count=1 | grep -c '^Test'  → 223   ← 기준선과 정확히 일치
$ go test ./internal/cli/ -list Audit -count=1 | grep -c '^Test'  →  99   ← 기준선과 정확히 일치
$ go test ./internal/cli/ -run GLM     -count=1 -timeout 600s → ok  ...internal/cli   6.777s
$ go test ./internal/cli/ -run Audit   -count=1 -timeout 600s → ok  ...internal/cli  16.942s
$ go test ./internal/cli/ -run Converg -count=1 -timeout 600s → ok  ...internal/cli   0.885s
                                                                 ↑ 생존. 223+99+32 건이 못 본다.

# 3) 사망 단계 — 가드 복원, 같은 뮤턴트
$ go test ./internal/cli/ -run TestGLMAudit_ -count=1 -v -timeout 600s
--- FAIL: TestGLMAudit_CancelledContext_IsNotSwallowed (0.36s)
    mcp_glm_audit_ctx_test.go:89: a cancelled audit returned the canned success verdict "pass" — the request reached the transport detached from its context (check http.NewRequestWithContext at mcp_glm.go:305)
--- PASS: TestGLMAudit_LiveContext_StillReachesTheDoer (0.36s)     ← 거울상은 살아 있다
FAIL	github.com/modu-ai/moai-adk/internal/cli	3.947s
   (12건 중 정확히 1건만 FAIL — 뮤턴트가 무차별로 깨뜨린 것이 아니라 컨텍스트 축만 잡았다)

# 4) 되돌림
$ git diff --stat   → 무출력       $ git status --short (추적)  → 무출력
```

**프로덕션 변경 0줄 + 양성 대조** (pathspec 이 공허하지 않음을 먼저 세운다)
```
$ git diff --stat dbc1f7125..HEAD -- 'internal/*.go' ':(exclude)*_test.go'   → 무출력
$ git diff --stat dbc1f7125..HEAD -- 'internal/*.go'                          ← 대조군
 internal/cli/mcp_glm_audit_ctx_test.go | 124 +++++++++++++++++++++++++++++++++
 1 file changed, 124 insertions(+)                     ← pathspec 은 작동한다
$ git diff --name-only dbc1f7125..HEAD -- '*.go'
internal/cli/mcp_glm_audit_ctx_test.go                  ← Go 파일은 이 하나뿐, 테스트 전용
```

**84건 닫힌 집합 3중 대조**
```
$ /usr/bin/grep -cE '^- `internal/[^`]+_test\.go:[0-9]+` `' .moai/specs/SPEC-CTX-BLIND-DOUBLE-001/plan.md  → 84
$ /usr/bin/grep -c '^- `internal/' .moai/specs/SPEC-CTX-BLIND-DOUBLE-001/plan.md                          → 85 (초과 1건은 file:line 형태 아님)
$ comm -23 plan84 attr84  → 무출력        $ comm -13 plan84 attr84  → 무출력   ← 양방향 차집합 0
$ sort plan84 | uniq -d   → 무출력                                            ← 중복 0
$ awk -F'\t' '$1=="method" && $6!="yes" {print $2}' sweep-test.tsv | sort -u | wc -l  → 153
$ comm -23 plan84rel sweepcand  → 무출력      ← 84건 전부가 진짜 스윕 후보 행
```

**M1 스윕 독립 재실행** (`AC-CBD-002` regression-guard)
```
$ go run .moai/reports/t539/ctxsweep/main.go internal > /tmp/t539-audit-sweep.tsv
exit=0 · stderr 무출력 · wc -l = 442
$ cmp /tmp/t539-audit-sweep.tsv .moai/reports/t539/sweep-test.tsv   → cmp_exit=0 (바이트 동일)
$ awk -F'\t' '{c[$1"/"$6]++} END {for (k in c) print k, c[k]}' | sort
funclit/ 196 · funclit/no 52 · funclit/yes 12 · method/ 127 · method/no 26 · method/yes 29
$ /usr/bin/grep -rl ctxsweep internal/   → grep_exit=1 (스캐너 부재)
```

**M3 도달성 프로브 — 생존 씨앗 8자리 전수** (`sync-audit-reach-*.log`)

| 씨앗 | 줄 | 프로브 결과 | 판정의 건전성 |
|---|---|---|---|
| 1 update | `orchestrator.go:55` | `REACHPROBE` 발생 → FAIL 0.561s | 도달 · 생존 유효 |
| 2 statusline | `builder.go:362` | `REACHPROBE` 발생 → FAIL 0.560s | 도달 · 생존 유효 |
| 3b lsp/core | `manager.go:366` | `REACHPROBE` ×5 → FAIL 0.378s | 도달 · 생존 유효 |
| 4a project | `initializer.go:434` | `REACHPROBE` 발생 → FAIL 0.507s | 도달 · 생존 유효 |
| 4b cli | `mirror_notice.go:42` | `REACHPROBE42` (`-run Mirror`) | 도달 · 유효 |
| 4b cli | `mirror_notice.go:37` | `REACHPROBE37` (`-run Update`) | 도달 · 유효 |
| 5b github | `pr_reviewer.go:110` | `REACHPROBE` 발생 → FAIL 0.674s | 도달 · 생존 유효 |
| 5c guardstate | `evaluate.go:180` | `REACHPROBE` 발생 → FAIL 0.339s | 도달 · 생존 유효 |
| **5a cli** | **`branch_protection.go:65`** | **무발생 → `ok 0.846s`** | **미도달 · 판정 공허 (F1)** |

**M3 검출 씨앗 2자리 재현**
```
# 3a — 175행(인용은 174, F3)
--- FAIL: TestGetDiagnostics_timeout_returnsEmptyWhenNoCache (0.20s)
    aggregator_test.go:553: timeout (no cache): got 1 diagnostics, want 0
--- FAIL: TestGetDiagnostics_timeout_returnsCachedOnSlowUpstream (0.30s)
    aggregator_test.go:530: got message "fresh", want "stale cached" (stale cached)
FAIL	github.com/modu-ai/moai-adk/internal/lsp/aggregator	0.813s   ← 로그와 축자 일치

# 3c — 행(hang) 이 검출인지, 외부 kill 인지 판별 (기준선 0.453s)
$ go test ./internal/lsp/transport/... -count=1 -timeout 90s
panic: test timed out after 1m30s
		TestCallWithTimeout_DeadlineExceeded (1m30s)
	.../internal/lsp/transport/request.go:53 +0x94
github.com/modu-ai/moai-adk/internal/lsp/transport_test.TestCallWithTimeout_DeadlineExceeded(...)
FAIL	github.com/modu-ai/moai-adk/internal/lsp/transport	90.338s
   → Go 테스트 런타임 자신의 타임아웃 패닉이고 정지 지점이 request.go:53 이다. 외부 kill 아님.
     180s 가 아닌 90s 로도 같은 결론이 서므로 판정이 특정 타임아웃 값에 매이지도 않는다.
```

**품질 게이트 재실행**
```
$ go build ./...                              → build_exit=0
$ GOOS=windows GOARCH=amd64 go build ./...    → win_build_exit=0
$ gofmt -l internal/cli/mcp_glm_audit_ctx_test.go                 → 무출력
$ go vet ./internal/cli/... ./internal/update/... ./internal/statusline/... \
         ./internal/lsp/... ./internal/core/project/... ./internal/github/... \
         ./internal/guardstate/...                                → vet_exit=0
$ golangci-lint run --timeout=10m             → 0 issues.   (기준선도 0 issues. → NEW 0건)
$ go test ./internal/cli/ -run TestGLMTask -count=1 -timeout 600s → ok 1.388s  (-list 16)
```

**대역 소유권 주장 검증** — `verdict.md` 가 근거로 든 "31개 호출 지점 · 7개 파일" 이 정확한지
```
$ /usr/bin/grep -rn 'stubGLMDoer' internal/cli/ | wc -l        → 33
$ /usr/bin/grep -c 'stubGLMDoer' internal/cli/mcp_glm_audit_ctx_test.go  → 2 (신규 파일의 주석)
   33 - 2 = 31, 파일 8 - 1(신규) = 7.   ← 주장과 정확히 일치
```

### Baseline-attribution — 무엇에 대고 쟀는가

| 측정 | 기준 | 이 감사에서 관측 |
|---|---|---|
| M2 생존 셀렉터 건수 | `list-{GLM,Audit,Converg}-baseline.log` (223/99/32) | 가드 파일을 치운 상태에서 `-list` 재실행 → **223 / 99** 정확히 재현. 기준선이 "가드 없는 트리"의 것이라는 귀속이 이로써 독립적으로 성립 |
| M2 사망 | `orch-m2-mutant-recheck.log` (오케스트레이터 배치) | 같은 뮤턴트 · 같은 실패 메시지, 이 감사에서 재현 (`sync-audit-m2-mutant.log`) |
| 스윕 산출물 | `sweep-test.tsv` (plan 단계) | `cmp_exit=0` — 바이트 동일 |
| 84건 열거 | `plan.md` §F M3 | 엄격 정규식 `wc -l` = 84, 귀속표와 양방향 차집합 0 |
| 3a 실패 메시지 | `m3-lsp.log` | 175행 재주입, 두 메시지 축자 일치 |
| 3c 행(hang) | `m3-lsp.log` (`180.201s`) | 90s 상한으로 재현 (`90.338s`), 정지 지점 동일 |
| 린트 NEW | `lint-baseline.log` (`0 issues.`) | `sync-audit-lint.log` (`0 issues.`) → NEW 0건 |
| 커버리지 | 없음 (run-phase Gap 5) | 이 감사에서 **처음 측정**: `81.1%` |

모든 수치는 트리 `.claude/worktrees/t539`, HEAD `73a588054` 에서 이번에 실행한 명령의 출력이다.
다른 트리·다른 시점의 수를 옮겨 적은 항목은 없다.

### Gaps — 이 감사가 관측하지 **않은** 것

1. **씨앗 11개 중 2개(1B seam 부재 · 3a/3c 는 검출로 이미 도달 증명)를 제외한 도달성 프로브를
   전수 돌렸으나, 84 구성원 각각의 도달성은 재지 않았다.** 프로브는 seam 단위다.
2. **F2 의 정정안 (a)** — `internal/cli/update/merge` 안의 실제 소비 seam 을 찾아 뮤턴트를 넣는 일 —
   을 수행하지 않았다. 감사자의 소관은 판정이지 수리가 아니다. 73·74행은 **미측정 상태 그대로** 남는다.
3. **전체 스위트를 돌리지 않았다.** `go test ./...` 는 로컬 금지(`CLAUDE.local.md` §4·§6)이며 전 패키지
   판정은 CI 몫이다. 돌린 것은 이 카드가 건드리거나 이 감사가 뮤턴트를 넣은 패키지뿐이다.
4. **`-race` 를 재실행하지 않았다.** run-phase 의 `m2-race-GLM.log` (`ok 11.937s`) 를 읽었을 뿐 다시 재지 않았다.
5. **커버리지 baseline(`dbc1f7125` 시점)을 직접 측정하지 않았다.** 워크트리에서 브랜치 상태를 바꾸는 것이
   금지돼 있어(`main-checkout-branch-guard.md`) 대신 **연역**했다 — 프로덕션 문장 집합이 동일하고
   테스트가 순증이므로 `coverage(HEAD) ≥ coverage(base)`. 이것은 측정이 아니라 추론이며 그렇게 표시한다.
6. **`research.md` 의 (a)(b) 축 분류 전수를 재검증하지 않았다.** (d) 축 근거 `codex_task.go:124`
   (`case <-ctx.Done():`) 한 곳만 확인했다.
7. **cross-model 2차 의견을 받지 않았다.** `audit_model` 미설정 → 단일 백엔드(Claude) 감사다.

### Residual-risk — 관측했는데도 틀릴 수 있는 것

1. **도달성 프로브는 "그 줄이 실행된다"만 세운다.** 실행되더라도 그 실행이 판정에 필요한 경로를
   지나갔는지는 별개다 — 특히 `-run Mirror` 에서 `:37` 이 아니라 `:42` 만 도달한 것처럼, 한 seam 의
   도달이 형제 분기의 도달을 뜻하지 않는다. 이 감사는 그 두 분기를 분리해 쟀지만 다른 씨앗에서는
   분리하지 않았다.
2. **생존 49건은 "가려진 결함이 있다"가 아니라 "가려진다면 보이지 않는다"이다.** run-phase 의
   Residual-risk 1 을 그대로 승계한다. 후속 카드는 이 구분을 지고 출발해야 한다.
3. **F1 의 보정 재측정이 세운 것은 `:122`·`:158` 두 seam 뿐이다.** `mockGHClient` 6개 메서드(77~82행)와
   `fakeQuerier` 2개(83~84행)의 귀속은 여전히 씨앗 5b·5c 한 자리씩에 기대며, 그 대표성은
   측정되지 않았다(run-phase Gap 3 승계).
4. **3a·3c 의 CLOSE 근거는 "위층이 본다"이고 그 위층은 나중에 사라질 수 있다.** aggregator 의
   서킷 브레이커나 transport 의 `isContextErr` 가 제거되면 (d) 축이 뒤집혀 34건이 다시 후보가 된다.
5. **M2 가드는 `ctxAwareGLMDoer` 의 충실도에 의존한다.** 이 대역은 `req.Context().Err()` 만 본다 —
   전송 중 취소나 부분 응답 같은 실제 `*http.Client` 의 미묘한 취소 의미론은 모델링하지 않는다.
6. **커버리지 81.1% 는 단일 실행값이다.** 병렬 부하나 캐시 상태에 따라 소수점이 흔들릴 수 있으나
   85% 임계와의 간격(3.9%p)이 그 흔들림보다 크므로 판정은 뒤집히지 않는다.

---

## sync 커밋 전에 오케스트레이터가 해야 할 일

sync 커밋 전에 **네 건의 기록 정정**이 blocking 이다. 첫째, F1 — `branch_protection.go:65` 를
주입 지점으로 적은 네 문서(`m3-gh-client.log`, `verdict.md` §2 표·§7 F5, `m3-attribution.md` 75~76행,
`progress.md` §E.2 M3 표)를 살아 있는 두 seam `:122`·`:158` 로 교체하고, 이 보고서에 실린 도달성
프로브와 보정 재측정 출력을 근거로 인용한다. 둘째, F2 — `m3-attribution.md` 73~74행의 merge 패키지
귀속이 패키지 경계 때문에 무효임을 기록하고, 그 두 행을 미측정으로 재분류하거나 merge 내부 seam 에서
재측정한 뒤, `verdict.md` 와 `m3-attribution.md` 의 **"미측정 0건" 문구를 실제 수(2건)에 맞게 정정**한다.
셋째, F3 — `aggregator.go:174` 를 `:175` 로 네 곳에서 고친다. 넷째, F4 — `verdict.md` §8 커밋 표에
`f83ed0c04` 와 `73a588054` 를 채우고 `progress.md` §E.3 의 `run_commit_sha` 를 확정한다.
그 뒤 선택 항목으로, F5 의 커버리지 81.1% 를 `verdict.md` §6 E3 과 §4 Gap 5 에 반영하고, F1 이 부수적으로
드러낸 죽은 코드 `DiscoverOwnerRepo`(F9)와 F6 의 도달성 전제 조항을 후속 카드 요청에 각각 한 항목으로
추가한다. 정정은 전부 문서 편집이며 **프로덕션 코드도 테스트 코드도 건드리지 않는다** — 이 카드의
불변식(프로덕션 0줄, M2 밖 수리 없음)은 정정 뒤에도 그대로 유지되어야 한다. 감사자는 트리를 원상태로
되돌려 두었다(`git diff --stat` 무출력, `git status --short` 추적 파일 변경 0). 이 감사가 생성한
untracked 로그 19개는 `.moai/reports/t539/sync-audit-*.log` 이며 증거로 커밋하든 버리든 무방하다.

---

## 감사 로그 (이 감사가 생성한 증거)

| 파일 | 내용 |
|---|---|
| `sync-audit-m2-clean.log` | 깨끗한 트리 가드 12건 PASS |
| `sync-audit-m2-mutant.log` | 가드 있는 상태 뮤턴트 → 1건만 FAIL, 거울상 PASS |
| `sync-audit-survive-{GLM,Audit,Converg}.log` | **생존 단계 재구성** (가드 제거 + 뮤턴트) |
| `sync-audit-m3-5a.log` · `sync-audit-m3-5a-reach.log` | 씨앗 5a 재현 + 무도달 증명 (F1) |
| `sync-audit-5a-live-{preflight,apply}.log` | F1 보정 재측정 — 살아 있는 두 seam 생존 |
| `sync-audit-5a-reach-{preflight,apply}.log` · `sync-audit-5a-reach158.log` | 보정 재측정의 도달성 증명 |
| `sync-audit-reach-{1-update,2-statusline,3b-lspcore,4a-project,5b-github,5c-guardstate}.log` | 생존 씨앗 도달성 프로브 |
| `sync-audit-reach-4b-{mirror,update}.log` | 씨앗 4b 두 분기 도달성 |
| `sync-audit-m3-3a.log` · `sync-audit-m3-3c.log` | 검출 씨앗 2자리 재현 |
| `sync-audit-lint.log` · `sync-audit-glmtask.log` · `sync-audit-cover-cli.log` | 품질 게이트 + 커버리지 81.1% |

---

_판정자: sync-auditor · 2026-09-08 · 단일 백엔드(Claude) 감사 · 프로필 `default` · 평면 가중 백분율_

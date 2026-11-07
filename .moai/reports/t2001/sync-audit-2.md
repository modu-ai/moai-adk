auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 7a946cfca29350476330355a20cf32595ce8825e

# t2001 sync 감사 2회차 (수리 확인) — PR #1955 (WT-cc-bash-monitor)

- 카드: t2001 (Class B, SPEC 없음) · 승인 기록: d-20261103T000424Z-c158 (이 카드의 수리 확인 1라운드)
- 고정 대상: `git rev-parse HEAD` → `7a946cfca29350476330355a20cf32595ce8825e` (감사 시작 시, 작성 직전 두 번 확인했고 일치). 작업 트리 깨끗함(`git status --short` 빈 출력).
- 비교 기준: `git merge-base HEAD origin/main` → `5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4`. 이전 감사 지점 `b2387e7123` 이후 수리 커밋 5개(4cd7b364a8, dba972a8be, 0b4e7483e6, b951841d85, 7a946cfca2).
- 판정 도구 빌드: `go test`/`go vet`/`golangci-lint`는 트리 소스를 직접 컴파일하므로 트리 HEAD 7a946cfca2 기준이다. `moai verify audit-plan`은 설치 CLI `v3.2.0-rc.39 v3.2.0-1041-gdad3603de1`, Codex leg는 MCP 서버 `build_commit: b1806498fe`가 판정했다. 두 커밋 모두 `git merge-base --is-ancestor <c> HEAD`가 exit 1이라 이 트리에서 만든 빌드가 아니다(§2.2, Gaps).
- 종합: **FAIL**. 1회차 blocking 결함 F1·F2가 지적한 경로는 닫혔다. 하지만 필수 Codex gate가 `fail`이고(새 P2), 이 감사도 그 P2를 같은 계열의 blocking 결함(N1)으로 본다.

## 요약 판정

| 항목 | 상태 |
|---|---|
| F1 (Windows 잠금이 데이터 읽기를 막음) | **CLOSED** — 소스와 Windows 교차 vet·테스트 컴파일로 확인. Windows 실행은 관측하지 못함 |
| F2 (stamp/본문 되감기로 inherited·warn) | **CLOSED** (단일 PR 경로). 남은 변형은 N1 |
| F3 (주석·설정의 옛 규칙) | **CLOSED** (지적한 다섯 곳). 같은 종류의 잔여는 N2 |
| F4 (migrate.go mutex 주석) | **CLOSED** |
| F5 (crossproc 테스트의 Windows leg 설명) | **CLOSED** |
| F6 (Monitor 테스트 과장 문구) | **CLOSED** |
| F7 (inbox 이중 rotation 경합) | **OPEN** (optional, 코드 변경 없음) |
| F8 (base=`HEAD^1` 전제) | **OPEN** (Info, 코드 변경 없음) |
| N1 | **NEW** Medium · blocking |
| N2, N3 | **NEW** Low · optional |
| N4 | **NEW** Info · optional |

결함 수: Medium 1(N1) · Low 2(N2, N3) · Info 1(N4), 그리고 이전 회차에서 열린 채 남은 F7(Low)·F8(Info).

## Codex leg (audit_multi, 원문 복사)

호출: `mcp__moai__audit_multi(project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001, target=baseBranch, focus=…)`. `gates` 인자와 `card_id`는 넘기지 않았다(지시된 단일 파일 외에는 기록하지 않기 위해서다).

| 항목 | 값 (원문) |
|---|---|
| overall_verdict | `fail` |
| gate_unmet | 결과에 필드 없음. `residual_risk_note: "required-backend FAIL: codex"` |
| plan_source | `config` |
| per_backend_verdicts | `codex` · gate `required` · verdict `fail` |
| disagreement_flag | `null` |
| participant_count | `1` |
| fail_open_backends | `[]` |
| audit_receipt | `rcpt-b5d843df9399068fbf1d6832` |
| build_commit | `b1806498fe` |
| tree_root | `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001` |

Codex summary 원문:

```
Verdict: fail
- [P2] 실제 stamp가 바뀌어도 `codemaps_inputs_stable=true`로 판정하여 차단을 우회한다 — internal/graph/check.go:842
  - 신뢰도: 높음. 디렉터리 내부의 Git 경로 변경만 검사하므로, 바깥 파일을 가리키는 provenance 심볼릭 링크의 대상 변경과 ignore된 provenance 변경을 놓친다.
  - `7a946cfca2`의 임시 overlay 시험에서 두 경우 모두 재현했다. 기준 트리는 `fresh/value=0`이었으나 stamp를 되감자 `stale/value=41/contribution=0`을 `inherited/warn`으로 분류했고 `Blocking()`은 false였다.
  - 권고: 실제 읽는 stamp/body의 바이트가 first parent와 같은지 검증하고, 동일성을 증명할 수 없는 입력은 차단하라.
  - 검증 한계: Windows는 교차 컴파일만 수행했다. 독립 감사 런처는 exit 1로 실패하여 별도 판정이 없다.
- [P3] CLI 문서와 정책 머리말에 수정 전 판정 조건이 남아 있다 — docs-site/content/en/cli-reference/graph.md:76
  - 신뢰도: 높음. 네 언어 문서 모두 `anchor-moved`와 stamp/body 동일성 조건을 누락했다. `internal/graph/check_policy.go:11`도 여전히 누적 개수와 기여 개수 조건만 설명한다.
  - 권고: 네 언어 문서와 정책 머리말에 변경되었거나 측정하지 못한 입력은 차단된다는 조건을 반영하라.
```

이 판정은 오케스트레이터 검증이 끝나기 전에는 최종이 아니다. 오케스트레이터가 위 digest 항목을 `<toplevel>/.moai/state/audit-plan-result.json`에 새로 쓰고 `moai verify audit-plan --project-root <toplevel> --result-file <그 경로>`를 돌려야 한다. 필수 backend가 `fail`이므로 그 검증도 PASS를 낼 수 없다.

## Audit plan

```
$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001
plan-exit=0
"config_status": "ok"
"model": "codex"
"cross_model_active": true
"cross_model_required": true
"enforced_required": [ "codex" ]
```

## 결함 목록 (structured defect-list)

- **N1** [Medium · 소스로 확인, 런타임 재현은 Codex 보고만 · 발생 가능성 낮음] [blocking] `internal/graph/check.go:840-852` (`codemapsInputsStable`) vs `internal/graph/check.go:403-404`
  - 안정성 판정은 `git diff --name-only <base> -- .moai/project/codemaps/`와 `ls-files --others --exclude-standard`로 **경로 목록**만 본다. 그런데 판정에 실제로 쓰는 stamp는 `os.ReadFile(.moai/project/codemaps/provenance.json)`이 읽는 **바이트**이고, `os.ReadFile`은 심볼릭 링크를 따라간다. git은 심볼릭 링크를 링크 문자열 blob으로만 기록한다. 그래서 base에서 provenance.json이 디렉터리 밖 파일을 가리키는 링크라면, PR이 그 대상 파일만 되감아도 디렉터리 diff는 비고 `codemaps_inputs_stable=true`가 된다. 이후는 1회차 F2와 같은 흐름이다(창이 넓어져 Value는 커지고 Contribution은 `.go`만 세므로 0 → `inherited`·`warn`). 필드 주석(`check.go:96-101`의 "byte-for-byte what the first parent carried")이 내건 전제가 이 입력에서 거짓이 된다.
  - 도달 조건: 지금 이 저장소의 provenance는 일반 파일이다(`git ls-files -s` → `100644`). 따라서 이 경로가 열리려면 provenance를 링크로 바꾸는 별도 변경이 먼저 병합되어야 한다. 그 변경 자체는 codemaps 변경이라 stale 상태에서는 `anchor-moved`로 막힌다. Codex가 함께 보고한 "ignore된 provenance" 변형은 provenance가 추적되는 CI 체크아웃에서는 도달할 수 없고 로컬 실행에서만 생긴다.
  - blocking으로 분류한 이유: F2와 같은 결함 계열이고(명시한 불변식이 깨짐), 필수 공동 감사자가 독립적으로 재현했다고 보고했으며, 고치는 비용이 작다.
  - Required fix: `Lstat`로 provenance.json(필요하면 본문 파일들까지)이 일반 파일인지 확인하고 아니면 안정성을 false로 둔다. 또는 `checkCodemaps`가 읽은 바이트를 `git show <base>:.moai/project/codemaps/provenance.json`과 비교하고, 비교할 수 없으면 nil로 둔다. 링크로 된 provenance에 대한 회귀 테스트를 추가한다.

- **N2** [Low] [optional] 코드 주석에 옛 규칙이나 불완전한 열거가 남아 있다(1회차 F3 목록에서 빠진 곳).
  - `internal/graph/check_policy.go:10-13` 패키지 머리말: "the change under judgment contributed fewer described-worthy files than the threshold"는 c295d363ff 이전 규칙이다. base 25 + 변경 16은 이 문장대로면 warn이지만 코드는 block한다(`TestGraphCheckPolicy_CrossedByChangeBlocks` PASS).
  - `internal/graph/check_policy.go:59-72` `ClassifyVerdicts` 문서: stamp/본문 안정성 조건과 `anchor-moved`가 빠져 있다.
  - `internal/graph/check.go:92-94` `DriftOrigin` 필드 주석: 값을 세 개(inherited, crossed-by-change, originated)로만 열거한다.
  - Required fix: 세 곳을 4분류(anchor-moved 포함) 규칙으로 고친다.

- **N3** [Low] [optional] 사용자 문서가 새 분류를 반영하지 않는다. `CHANGELOG.md:21`과 `docs-site/content/{en,ko,ja,zh}/cli-reference/graph.md`(en은 `:76`)는 `drift_origin`을 세 값으로 닫아 열거하고, 다운그레이드 조건에 "변경이 codemaps 본문·stamp를 건드리지 않았을 것"이 없다. 새 JSON 필드 `codemaps_inputs_stable`도 문서에 없다(`grep -rn "anchor-moved\|codemaps_inputs_stable" CHANGELOG.md docs-site/content/*/cli-reference/graph.md` → 일치 없음, exit 1). 남아 있는 문장 중 거짓이 된 것은 "세 값 중 하나"라는 열거뿐이다. "inherited만 다운그레이드한다"는 여전히 참이다. Required fix: CHANGELOG 항목과 네 언어 문서에 `anchor-moved`, 안정성 조건, 새 필드를 추가한다.

- **N4** [Info · 추론, 미측정] [optional] 임계값은 판정 대상 트리의 `gate.yaml`에서 읽는다. PR이 `codemaps_changed_files`를 base의 실제 개수보다 낮추면, base는 자기 임계값으로는 fresh였는데도 `Value - C ≥ 새 임계값`이 되어 `inherited`·`warn`으로 분류된다. 게이트를 조이는 변경이 스스로를 block하지 않고 warn을 받는 정도라 해는 작다. 또 pull_request 워크플로는 PR head의 코드를 빌드하므로(`graph-freshness.yml` "Build binary from PR head") 이 게이트는 원래 악의적인 PR 작성자를 막는 경계가 아니다. 기록만 남긴다.

- **F7** [Low · 추론] [optional] OPEN — `internal/hook/failure_observer.go`, `internal/hook/inbox_lifecycle.go`는 b2387e7123 이후 바뀌지 않았다(수리 diff stat에 없음). 1회차 설명이 그대로 유효하다.

- **F8** [Info] [optional] OPEN — base는 여전히 `HEAD^1`이다(`check.go:818`). 배포된 `graph-freshness.yml`은 기본 merge ref를 체크아웃하므로 CI에서는 이 경로가 열리지 않는다.

## 범위별 판단

### A. 수리 커밋

**F2 — 닫힘 판단의 근거.** `driftOrigin`(`check_policy.go:106-120`)은 이제 `originated`를 먼저 보고, 다음으로 `CodemapsInputsStable`이 nil이거나 false면 `anchor-moved`로 분류하며, 그 뒤에야 산술로 판정한다. 안정성 값은 `attachContribution`이 base를 확인한 뒤에만 설정하고(`check.go:831`), git이 실패하면 nil을 돌려준다(`check.go:842-849`). base가 없으면 Contribution도 nil이라 출처가 비고 block한다. 단위 테스트가 false와 nil 두 경우를 고정한다(`TestClassifyVerdicts_MovedOrUnknownAnchorBlocks`).

반례를 찾아본 범위: content anchor의 입력은 provenance.json(stamp, roots, dirty)과 codemaps 본문, 그리고 stamp S의 역사뿐이다(`resolveContentAnchor` `check.go:331-369`, 경로는 `check.go:397`와 `codemapsBodyPathspec` `:228-231`에 같은 문자열로 하드코딩되어 있다). 디렉터리가 base와 같으면 anchor도 base의 anchor와 같다. 그때 Value와 Contribution은 같은 술어(`mx.IsDescribedWorthy`)와 같은 합집합(diff + untracked)으로 세어지므로, 경로 집합에 대해 |D(a,w)| ≤ |D(a,b)| + |D(b,w)|가 성립한다. `--name-only`의 rename 표시는 목적지 경로만 내보내므로 이 부등식을 깨지 않는다. 따라서 `inherited`이면 base 자신의 개수도 임계값 이상이다(같은 임계값과 같은 판정 코드라는 전제 아래). 이 전제가 깨지는 입력은 N1(링크된 stamp)과 N4(임계값 변경)뿐이었다. CI의 bootstrap 단계(`mx scan` → `.moai/state/`, `graph build` → gitignore된 `.moai/project/graph/`)는 codemaps 디렉터리에 아무것도 쓰지 않는다(`grep` 결과, 아래 Evidence). 그래서 정상적인 CI 실행에서 안정성이 늘 false가 되어 R208 warn이 죽어 버리는 일은 생기지 않는다.

**fixture가 실제 CI 모양과 맞는지.** `driftFixture`는 이제 base 쪽 커밋에 codemaps를 추적 파일로 넣는다(`commitCodemaps`). 재-stamp가 같은 바이트를 쓰므로 `TestGraphCheckPolicy_UntouchedStampStillWarns`가 `"codemaps_inputs_stable": true`와 exit 0을 확인한다. `rewoundStampFixture`는 base fresh 상태에서 41개 파일을 추가한 뒤, docs-only 변경과 함께 stamp를 c0으로 되감는다. 이 경우 exit 1과 `anchor-moved`가 나온다(PASS). merge 커밋 대신 선형 커밋으로 `HEAD^1`을 만들지만 판정 코드가 보는 것은 `HEAD^1`과 작업 트리뿐이라 구조상 같다.

**저자가 "base의 개수를 다시 계산"하는 대신 "창이 움직이면 fail closed"를 고른 것에 대해.** 타당하다고 본다. 새 분기는 R208의 완화를 줄일 뿐이고, block은 R208 이전 main의 기본 동작이다. 과잉 차단은 "base가 이미 stale인데 PR이 codemaps 문서나 stamp를 손으로 고친" 경우에만 생긴다. 정식 재생성(`moai codemaps` + stamp)은 보통 codemaps를 fresh로 만들어 분류 자체가 일어나지 않는다. 손으로 고친 문서에 대한 해결책은 재생성이다. 다만 판정의 정확도는 N1이 보여 주듯 "경로가 같다 = 바이트가 같다"라는 근사에 달려 있다.

**F1 — 닫힘 판단의 근거.** `lockOverlapped()`(`lockfile_windows.go:199-201`)가 `OffsetHigh: 0xFFFFFFFF`, Offset 0을 돌려준다. `Lock`과 `TryLock`은 `lockRegion`을 거쳐 이 값을 쓰고(`:206`), `Unlock`도 같은 helper와 같은 길이 `(1, 0)`를 쓴다(`:215`). 세 경로의 영역이 같다. Microsoft 문서에 따르면 `LockFileEx`는 파일 끝을 넘는 영역도 잠글 수 있다. 이 감사는 그 문서를 다시 가져와 확인하지 않았다(Gaps). 같은 저장소의 `internal/sessionmsg/lock_windows.go:62-88`도 높은 offset 영역 패턴을 이미 쓴다. 호출부 전체(`grep`, 아래 Evidence)를 보면 taskledger(`taskledger.go:76-91`, 데이터 파일 잠금 후 별도 핸들로 읽음)와 retention(`retention.go:137,175`)만 데이터 파일을 잠그고, 나머지는 모두 전용 `.lock`/sidecar 파일을 잠근다. 1회차 목록에 없던 `internal/escalation/lock.go`(base와 동일, `escalation-diff=0`)도 전용 `files.Lock`을 잠근다. byte-0의 강제 차단에 기대는 호출부는 없다. main의 Windows 구현은 프로세스 내부 mutex였으므로(`git show 5c3f0ebf27:…lockfile_windows.go` → `sync.Mutex`, `map[`) 애초에 강제 차단에 기댈 수가 없었다. 새 Windows 전용 테스트는 읽어 보면 맞다. 다른 핸들의 `os.ReadFile`과 append가 성공해야 하고, 두 번째 핸들의 `TryLock`은 `errLockHeld`, `Unlock` 뒤에는 성공해야 한다. 하지만 **Windows에서 실행한 사람은 없다.**

**F3-F6.** 지적한 다섯 곳(`check.go:99-108`→현재 `:105-116`, `gate.yaml` 두 사본, `types.go`, `defaults.go`)은 3분류와 안정성 조건을 반영한다. 두 `gate.yaml`에서 PR이 추가한 블록은 바이트가 같다(`added-block-cmp=0`). 파일 전체의 차이(62행 `enabled`, 70행 `rules_dir`)는 이 PR 이전부터 있던 것이다. `migrate.go`, `lockfile_crossproc_test.go`, `monitor_wireformat_test.go` 주석은 이제 코드와 맞는다.

### B. 이전 범위 재확인
- registry 테스트 파일은 base와 바이트가 같다(`m5-base-vs-HEAD=0`).
- `lockfile_contract_test.go`와 `lockfile_unix.go`의 PR 변경은 주석뿐이다(diff 확인). assertion을 지운 곳은 없다.
- inbox cap 경로와 Monitor-as-Bash 회귀 테스트는 수리 커밋에서 주석 외에 바뀌지 않았고, 범위 테스트가 모두 PASS다.
- CHANGELOG와 docs-site: 1회차에 맞았던 진술(플래그, 환경변수 fallback, 기본 `warn`, exit 코드)은 여전히 맞다. 새 분류가 빠진 점은 N3에 적었다.

### C. 그 밖의 diff
- 보안: 새 입력 경계는 없다. 이벤트는 완전일치 비교이고, 정책 값은 화이트리스트(warn|block) 밖이면 exit 2다. N1과 N4는 게이트 정합성 문제다. PR head 코드를 빌드하는 워크플로라 권한 상승으로 이어지지는 않는다.

## 차원별 점수

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 75/100 | FAIL | F1·F2의 원래 경로는 닫혔다(테스트 PASS, 아래 Claim C). 하지만 N1(링크된 stamp로 F2 불변식 우회, blocking)이 남았다 |
| Security (25%) | 85/100 | PASS | 신뢰 경계 변화 없음. 이벤트 완전일치, 정책 화이트리스트, exit 2 |
| Craft (20%) | 78/100 | PASS-WITH-DEBT | gofmt 출력 없음, `go vet` 호스트와 Windows exit 0, golangci-lint 0 issues, 회귀 테스트 추가. 다만 Windows 테스트는 실행되지 않았고 coverage는 측정하지 않았다 |
| Consistency (15%) | 70/100 | PASS-WITH-DEBT | N2(주석 3곳), N3(CHANGELOG와 docs-site 4개 언어) |

Must-pass(Functionality, Security): Functionality가 N1 때문에 기준에 못 미친다. 필수 Codex gate도 `fail`이므로 → **FAIL**.

## Claim / Evidence

**Claim A — 고정 SHA와 범위**
```
$ git rev-parse HEAD
7a946cfca29350476330355a20cf32595ce8825e
$ git merge-base HEAD origin/main
5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4
$ git log --oneline b2387e7123..HEAD
7a946cfca2 [t2001] docs(comments): make stale comments and config text match the code (F3-F6, card t2001)
b951841d85 [t2001] fix(lockfile): lock a region past any file data on Windows so the lock stays advisory (F1, card t2001)
0b4e7483e6 [t2001] test(lockfile): pin that the Windows lock excludes lockers without blocking data I/O (F1, card t2001)
dba972a8be [t2001] fix(graph): count a drift as inherited only while the change left the codemaps stamp and body alone (F2, card t2001)
4cd7b364a8 [t2001] test(graph): pin that a change which moves the codemaps stamp is never "inherited" (F2, card t2001)
$ git diff --stat b2387e7123..HEAD | tail -1
 14 files changed, 324 insertions(+), 49 deletions(-)
$ git diff --stat 5c3f0ebf27..HEAD | tail -1
 31 files changed, 1381 insertions(+), 101 deletions(-)
```

**Claim B — 범위 테스트 (macOS, 이 트리)**
```
$ uptime
 9:07  up 2 days,  4:41, 26 users, load averages: 120.03 72.16 50.45
$ go test ./internal/graph/ ./internal/lockfile/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/graph	74.946s
ok  	github.com/modu-ai/moai-adk/internal/lockfile	0.573s
$ go test ./internal/cli/ -run 'GraphCheck' -count=1 -v   (파일로 리다이렉트)
exit=0 · '--- PASS' 24줄 · '--- (FAIL|SKIP)' 0줄
--- PASS: TestGraphCheckPolicy_RewoundStampBlocks (1.29s)
--- PASS: TestGraphCheckPolicy_UntouchedStampStillWarns (1.43s)
--- PASS: TestGraphCheckPolicy_CrossedByChangeBlocks (1.39s)
--- PASS: TestGraphCheckPolicy_BaseOverLinePrZeroWarns (1.60s)
ok  	github.com/modu-ai/moai-adk/internal/cli	27.892s
$ go test ./internal/hook/ -run 'Inbox|WorktreeRegistry|ShellTool|MonitorCommand' -count=1 -v | grep -E '^--- |^ok|^FAIL|^PASS'
--- PASS: TestInboxCap_RotatesWhenPathSizeLagsBehindHandle (0.00s)
--- PASS: TestInboxCap_ConcurrentAppendsCrossCapBoundary (0.08s)
--- PASS: TestMonitorCommandUnderBashMatcherIsGuarded (0.00s)
--- PASS: TestWorktreeRegistry_TwoWritersAcrossProcessesLoseNeitherEntry (0.23s)
--- PASS: TestWorktreeRegistry_HeldLockPastLoweredBoundLapses (0.14s)
(그 밖의 PASS 18줄, FAIL 0줄)
PASS
ok  	github.com/modu-ai/moai-adk/internal/hook	3.673s
```

**Claim C — 정적 검사**
```
$ gofmt -l internal/graph internal/lockfile internal/cli/graph_check.go internal/cli/graph_check_policy_test.go internal/cli/worktree/migrate.go internal/config internal/hook
(빈 출력)
$ go vet ./internal/graph/ ./internal/lockfile/ ./internal/cli/ ./internal/hook/ ./internal/config/
vet=0
$ GOOS=windows go vet ./internal/lockfile/ ./internal/hook/ ./internal/cli/worktree/ ./internal/graph/
winvet=0
$ GOOS=windows go vet ./internal/lockfile/ ./internal/escalation/ ./internal/cli/taskledger/ ./internal/harness/
winvet=0
$ GOOS=windows go test -c -o /dev/null ./internal/lockfile/
wintestcompile=0
$ golangci-lint run ./internal/graph/ ./internal/lockfile/
0 issues.
$ GOOS=windows golangci-lint run ./internal/lockfile/
0 issues.
```

**Claim D — 테스트 약화 없음, 설정 사본 일치**
```
$ git diff --quiet 5c3f0ebf27 HEAD -- internal/hook/worktree_registry_m5_test.go; echo $?
m5-base-vs-HEAD=0
$ cmp <PR이 추가한 template gate.yaml 줄> <PR이 추가한 live gate.yaml 줄>
added-block-cmp=0   (12줄)
$ diff .moai/config/sections/gate.yaml internal/template/templates/.moai/config/sections/gate.yaml
62c62  enabled: true / enabled: false
70c70  rules_dir: ".moai/astgrep-rules" / ".moai/config/astgrep-rules"   (PR 이전부터 있던 차이)
```

**Claim E — F1 소스 근거**
```
$ grep -rn "lockfile\.\(Lock\|TryLock\|Unlock\)" internal cmd pkg | grep -v _test.go
internal/contract/receipt/store.go:209/212  (store.lock)
internal/template/settings.go:118/121       (lockF)
internal/cli/settings.go:81/84              (lockF)
internal/cli/glm_tools.go:645/650           (lockF)
internal/cli/worktree/migrate.go:133/139    (TryLock/Unlock)
internal/cli/taskledger/taskledger.go:82/86 (tasklist.md 데이터 파일)
internal/harness/retention.go:175/178       (stamp 상태 파일)
internal/escalation/lock.go:34/38/51        (files.Lock, detector.go:126)
internal/hook/worktree_registry.go:224/226  (lockPath)
internal/hook/agentmemory.go:463/466        (lockF)
$ git show 5c3f0ebf27:internal/lockfile/lockfile_windows.go | grep -n "sync.Mutex\|map\["
12:var fileLocksMu sync.Mutex
27:var fileLocks = map[string]*sync.Mutex{}
```

**Claim F — N1 소스 근거와 CI 모양**
- `check.go:403-404` `pvPath := filepath.Join(dir, "provenance.json")` → `os.ReadFile(pvPath)` (링크를 따라감)
- `check.go:842,846` `git diff --name-only <base> -- .moai/project/codemaps/`, `git ls-files --others --exclude-standard -- <dir>` (경로만 비교)
- `git ls-files -s .moai/project/codemaps/provenance.json` → `100644` (현재 일반 파일)
- `graph-freshness.yml:42-112`: checkout(`fetch-depth: 0`, 기본 merge ref) → stamp 도달성 가드 → `go build -o ./bin/moai ./cmd/moai` → `moai mx scan --quiet` + `moai graph build` → `./bin/moai graph check`. 출력 위치: `check.go:613` `.moai/state/<sidecar>`, `meta.go:16` `edges.meta.json`은 gitignore된 `.moai/project/graph/`(`.gitignore:345`). 둘 다 codemaps 디렉터리 밖이다.

## Baseline-attribution

위 명령은 모두 이 실행에서 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001`, HEAD `7a946cfca29350476330355a20cf32595ce8825e`를 대상으로 돌렸다. Go 도구 결과는 트리 소스를 컴파일한 것이다. `moai verify audit-plan`은 설치 빌드 `dad3603de1`, Codex leg는 MCP 서버 빌드 `b1806498fe`가 판정했다. 둘 다 HEAD의 조상이 아니므로(`--is-ancestor` exit 1) 트리 HEAD에서 만든 빌드가 아니다. Codex leg가 판정한 것은 Go 코드를 직접 읽고 돌린 결과라고 보고되어 있다(overlay 시험).

## Gaps (관측하지 않은 것)

- **Windows 실행은 전혀 관측하지 못했다.** 새 잠금 코드와 `TestWindowsLockAllowsDataIOWhileExcludingContenders`는 Windows에서 실행된 적이 없다. Windows runner가 없고, b2387e7123 이후 커밋은 푸시되지 않아 CI도 보지 못했다(원격 PR head는 a81a31875b). 확인한 것은 `GOOS=windows go vet`, 테스트 바이너리 교차 컴파일, 소스 읽기뿐이다. "LockFileEx가 EOF 너머 영역을 허용한다"는 Microsoft 문서 내용은 이번 실행에서 다시 가져와 확인하지 않았고, 같은 저장소 sessionmsg의 선례와 1회차 근거에 기댔다.
- **coverage는 측정하지 않았다**(부하 120 상태라 범위 테스트만 돌렸다).
- **N1의 런타임 재현은 Codex 보고뿐이다.** 이 감사는 소스로 경로를 확인했다. 직접 재현하려던 scratch 저장소 시연은 worktree 격리 가드가 거부했다(아래). Codex가 말한 overlay 시험 산출물은 트리에 없다.
- **도구 거부 (§3.1)**: worktree 격리 가드가 Bash 2건을 거부했다. (1) gofmt·vet·lint·`diff <(git show …)`를 묶은 복합 명령은 단순 명령으로 나눠 같은 검사를 다시 돌렸다. `gate.yaml` 비교는 `git show` 대신 작업 트리 파일의 `cmp`/`diff`로 대신했는데, 작업 트리가 깨끗하므로 HEAD와 같다. (2) scratchpad에서 새 git 저장소를 만들어 링크된 provenance를 시연하려던 명령은 가드를 우회하지 않고 생략했다. N1은 소스 읽기와 Codex 보고로만 판단했다.
- `golangci-lint`는 graph·lockfile만 돌렸다. hook·cli·config는 돌리지 않았다.
- docs-site ja/zh/ko 본문은 `grep`으로 `anchor-moved` 부재만 확인했고 문장 전체는 읽지 않았다(en만 읽음).
- Monitor-as-Bash 런타임 관측(Claim 1·2)은 다시 측정하지 않았다.
- N4는 소스 추론이고 fixture로 재현하지 않았다.

## Residual-risk

- N1을 고치더라도 안정성 판정은 "PR이 codemaps 입력을 바꾸지 않았다"는 사실에만 기댄다. 판정 코드 자체(술어, 기본 roots, 임계값)를 PR이 바꾸면 base의 실제 판정과 달라질 수 있다(N4). pull_request 워크플로가 PR head를 빌드하는 한 피할 수 없는 성질이다.
- 정책상 inherited drift가 PR에서 warn으로 통과하면, 병합 뒤 main push 실행은 block한다. 누군가 codemaps를 다시 만들 때까지 main이 빨간색으로 남을 수 있다(R208이 의도한 trade-off).
- Windows `LockFileEx`는 핸들 단위다. 다른 핸들로 Unlock하면 `ERROR_NOT_LOCKED`→nil이라 조용히 아무것도 풀지 않는다. 지금 호출부는 모두 같은 핸들로 Unlock하지만, 이후 호출부가 이 가정을 깰 수 있다. 이전 빌드(프로세스 mutex, 또는 PR head a81a31875b의 byte-0 영역)와 새 빌드가 같은 잠금 파일에서 섞여 돌면 서로 배제하지 못한다(`lockfile_windows.go:33-35` COMPATIBILITY 주석).
- F7 경합은 테스트로 확인하지 않았다.
- 세션 scratchpad(`/private/tmp/claude-501/…/scratchpad/`)에 남긴 출력(`cli.txt`, `plan.json`)은 이 파일에 옮긴 줄 외에는 인용 근거가 아니며, 세션이 끝나면 사라진다.

## Iteration history

- 1회차 (`sync-audit-1.md`): FAIL @ b2387e7123 — blocking F1·F2, Codex required gate `fail` (rcpt-9ac33515965037b2a13e2e8c).
- 2회차 (이 파일): FAIL @ 7a946cfca2 — F1·F2 원래 경로 CLOSED, F3-F6 CLOSED, F7·F8 OPEN. NEW: N1(blocking, Codex P2와 같은 내용), N2·N3(Codex P3와 겹침), N4. Codex required gate `fail` (rcpt-b5d843df9399068fbf1d6832).

## 권고

1. N1: `checkCodemaps`가 읽는 provenance(와 본문)가 일반 파일인지 `Lstat`로 확인하거나 base blob과 바이트를 비교한다. 비교할 수 없으면 nil로 둔다. 회귀 테스트를 하나 추가한다.
2. N2·N3: 정책 머리말, `ClassifyVerdicts` 문서, `DriftOrigin` 주석, CHANGELOG, docs-site 네 언어에 `anchor-moved`와 안정성 조건을 넣는다(Codex P3와 같은 범위).
3. 푸시한 뒤 Windows leg(릴리스 multi-OS 또는 workflow_dispatch)에서 `go test ./internal/lockfile/ ./internal/cli/taskledger/ ./internal/harness/`를 실제로 관측한다.
4. 재감사는 N1-N3 delta로 범위를 좁힌다.

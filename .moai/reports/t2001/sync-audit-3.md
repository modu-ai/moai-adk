auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 2fad4a8de6426f559d1c6d915c0b9e6a2c424e03

# t2001 sync 감사 3회차 (수리 확인) — PR #1955 (WT-cc-bash-monitor)

- 카드: t2001 (Class B, SPEC 없음) · 승인 기록: d-20261103T002644Z-9259 (수리 뒤 확인 감사 1회, FAIL이면 카드 정지)
- 고정 대상: `git rev-parse HEAD` → `2fad4a8de6426f559d1c6d915c0b9e6a2c424e03`. 감사 시작 때와 이 파일을 쓰기 직전에 두 번 확인했고 같았다. 작업 트리는 두 번 모두 깨끗했다(`git status --short` 빈 출력).
- 원격 PR head: `gh pr view 1955 --json headRefOid` → `2fad4a8de6426f559d1c6d915c0b9e6a2c424e03` (draft, OPEN, MERGEABLE). 고정 SHA와 같다.
- 비교 기준: `git merge-base HEAD origin/main` → `5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4`. 이전 판정 커밋 `5bfd590da5` 이후 수리 커밋 4개(4d434b4d78, e3805e8766, 64e4a8eb66, 2fad4a8de6).
- 종합: **FAIL**. N1의 두 우회 경로(링크된 stamp 대상 되감기, ignore된 stamp)는 실제로 막혔고, RED 커밋에서 두 테스트가 실패하는 것도 직접 관측했다. 그러나 수리 코드가 **필수 CI 검사 `Lint`를 깨뜨렸다**(D1, staticcheck QF1001, 로컬과 CI 모두 관측). 필수 Codex gate도 `fail`이다.

## 요약 판정

| 항목 | 상태 |
|---|---|
| N1 (링크·ignore로 stamp 안정성 우회) | **CLOSED** — 보고된 두 경로는 막힘(RED 실패·GREEN 통과 관측). 재독(TOCTOU) 변형은 D2로 따로 기록 |
| N2 (주석의 옛 규칙) | **CLOSED** |
| N3 (CHANGELOG·docs-site 4개 언어) | **CLOSED** — 정밀도 문제는 D4·D5로 따로 기록 |
| N4 (임계값을 PR이 바꾸는 경우) | **OPEN** (Info, 코드 변경 없음) |
| F1 (Windows 잠금 영역) | **CLOSED** — 수리 커밋은 lockfile을 건드리지 않았고 Windows 교차 vet exit 0. Windows 실행은 CI에서 아직 진행 중(Gap) |
| F2 (stamp 되감기로 inherited) | **CLOSED** — `TestGraphCheckPolicy_RewoundStampBlocks` PASS |
| F3 · F4 · F5 · F6 | **CLOSED** (수리 커밋에서 바뀌지 않음) |
| F7 (inbox 이중 rotation 경합) | **OPEN** (Low, optional, 코드 변경 없음) |
| F8 (base=`HEAD^1` 전제) | **OPEN** (Info, `check.go:822` 그대로) |
| D1 | **NEW** Medium · blocking |
| D2, D3, D4 | **NEW** Low · optional |
| D5 | **NEW** Info · optional |

결함 수(이번 회차 기준 열린 것): Medium 1(D1) · Low 4(D2, D3, D4, F7) · Info 3(D5, N4, F8).

## 결함 목록 (structured defect-list)

- **D1** [Medium · 관측] [blocking] `internal/graph/check.go:892` (`regularCodemapsTree`)
  - 새 조건식 `if err != nil || !(d.IsDir() || d.Type().IsRegular())`가 staticcheck QF1001에 걸린다. `golangci-lint run ./internal/graph/`가 로컬에서 exit 1이고, CI의 `Lint` 체크(run 41555114444, head 2fad4a8de6)가 같은 위치 annotation으로 `fail`이다. `Lint`는 main 보호 규칙의 필수 체크다(`required_status_checks.contexts` → `["Lint","Build (linux/amd64)","Analyze (Go) (go)"]`). 따라서 지금 head로는 병합할 수 없다. 2회차 감사가 같은 패키지에서 `0 issues.`를 기록한 것은 이 코드가 생기기 전이었다.
  - Required fix: `if err != nil || (!d.IsDir() && !d.Type().IsRegular()) {`로 바꾸고 `golangci-lint run ./internal/graph/`가 `0 issues.`인지 확인한다. 동작은 바뀌지 않는다.

- **D2** [Low · 소스로 확인, 런타임 재현은 Codex 보고만] [optional] `internal/graph/check.go:870` vs `internal/graph/check.go:408`
  - 안정성 검사는 stamp를 **다시 읽어서** base blob과 비교한다. 하지만 측정에 쓰인 바이트는 `checkCodemaps`가 408행에서 처음 읽어 파싱한 것이다. 두 읽기 사이에 다른 프로세스가 stamp를 base 내용으로 되돌리면, 되감긴 stamp로 계산한 Value가 `codemaps_inputs_stable=true`·`inherited`를 받는다. 필드 주석과 함수 주석이 내건 "checkCodemaps가 읽는 stamp의 바이트를 비교한다"는 전제가 엄밀히는 성립하지 않는다.
  - 도달 조건: 검사 도중 stamp를 쓰는 동시 작성자가 있어야 한다. CI의 graph-freshness job은 단일 프로세스이고, 로컬 실행은 보통 이벤트가 없어 어차피 block한다. 게다가 pull_request 워크플로는 PR head 코드를 빌드하므로 악의적인 작성자를 막는 경계가 아니다(N4와 같은 성질). 그래서 optional로 둔다.
  - Required fix(선택): `checkCodemaps`가 읽은 `data`를 `attachContribution` → `codemapsInputsStable`로 넘겨 그 바이트를 base blob과 비교한다. 두 번째 읽기는 없앤다.

- **D3** [Low · 관측] [optional] `internal/cli/graph_check_policy_test.go:371`, `:415`
  - 두 새 테스트는 방어 장치 세 개(링크 walk, ignore 목록, stamp 바이트 비교) 중 **어느 하나만 지워서는 실패하지 않는다**. 스크래치 사본에서 두 변이체를 직접 돌렸다. M1(walk와 ignore 검사를 끔, 바이트 비교 유지)과 M2(바이트 비교를 끔, walk와 ignore 검사 유지) 모두에서 두 테스트가 PASS였다. 링크 경우는 링크 문자열 blob과 대상 바이트가 달라서, ignore 경우는 base에 stamp blob이 없어서 바이트 비교만으로도 걸리고, 반대로 walk와 ignore 검사만으로도 걸린다. 묶음 전체는 의미가 있다(RED 커밋에서 둘 다 실패, 아래 Claim C). 다만 나중에 누가 장치 하나를 "중복"이라며 지워도 테스트는 잡지 못한다.
  - Required fix(선택): `codemapsInputsStable`을 직접 부르는 단위 테스트를 장치별로 하나씩 둔다. 예를 들어 링크된 본문 파일(바이트 비교는 통과)과 ignore된 본문 파일(바이트 비교는 통과)이 각각 nil을 돌려받는지 고정한다.

- **D4** [Low · 소스 추론] [optional] `CHANGELOG.md:21`, `docs-site/content/{en,ko,ja,zh}/cli-reference/graph.md:76`, `internal/graph/check.go:99-104`, `internal/graph/check_policy.go:11-12`, `:71-73`
  - 문서와 주석은 "stamp와 본문이 첫 부모와 바이트 단위로 같을 때만 true"라고 쓴다. 코드가 바이트로 비교하는 것은 stamp뿐이다. 본문은 `git diff --name-only`로 비교하므로 git이 보는 내용(eol 정규화, index의 assume-unchanged·skip-worktree 비트 반영 뒤)이 같다는 것만 증명한다. Codex는 이 차이로 본문 변경이 `inherited`·exit 0을 받는 것을 재현했다고 보고했다(P2).
  - 이 감사가 판정을 바꾸지 않는 이유: content anchor 계산(`resolveContentAnchor`, `check.go:335-373`)도 본문을 같은 git 관점으로만 본다. 그래서 git이 보는 본문이 base와 같으면 anchor도 base의 anchor와 같고, Value − Contribution은 그대로 base의 개수다. 본문 바이트는 이 산술에 들어가지 않는다. index 비트는 PR이 CI 체크아웃에 심을 수 없는 로컬 상태다. 결국 남는 문제는 문서가 실제 보장보다 강하게 말한다는 점뿐이다. 이 판단은 소스 추론이고 런타임으로 확인하지 않았다(Gaps).
  - Required fix(선택): "본문은 git이 보는 내용이 같고, stamp는 바이트가 같을 때"로 문장을 고치거나, 본문도 `git ls-tree`/`cat-file`로 바이트 비교하도록 코드를 강화한다.

- **D5** [Info · 관측] [optional] `CHANGELOG.md:21` — 열거가 "`crossed-by-change` (…), or `originated` (…), or `anchor-moved` (…)"로 "or"가 두 번 나온다. en docs-site(`graph.md:76`)는 바르게 고쳐져 있다. Required fix: 앞의 "or"를 지운다.

- **N4** [Info] [optional] OPEN — 2회차 설명 그대로. 수리 커밋은 임계값 읽기 경로를 바꾸지 않았다.
- **F7** [Low] [optional] OPEN — `internal/hook/failure_observer.go`, `inbox_lifecycle.go`는 수리 diff에 없다(`git diff --stat 5bfd590da5 HEAD`).
- **F8** [Info] [optional] OPEN — base는 여전히 `HEAD^1`이다(`check.go:822`).

## N1 우회 시도 (요청 목록별)

| 시도 | 결과 | 근거 |
|---|---|---|
| codemaps 디렉터리 자체가 심볼릭 링크 | 막힘(nil) | `regularCodemapsTree`가 `os.Lstat(root)`로 디렉터리인지 본다(`check.go:887`). `filepath.Join`이 끝의 `/`를 지우므로 링크 자체가 Lstat 대상이 된다 |
| stamp나 본문 파일이 링크(대상이 디렉터리 밖) | 막힘(nil) | `filepath.WalkDir`는 링크를 따라가지 않고 `d.Type()`이 `ModeSymlink`이라 regular가 아니다(`:891-896`). 게다가 stamp 링크는 바이트 비교에서도 걸린다(D3 변이체 M1 PASS로 관측) |
| `.moai`·`.moai/project`가 링크 | 막힘(false 또는 nil) | git은 링크 아래 경로를 추적하지 못한다. base에서 실제 파일이었다면 작업 트리에서 삭제로 보여 `tracked` 목록이 비지 않는다. 링크가 base에도 있었다면 `git show base:<stamp>`가 실패해 nil이다(소스 추론) |
| git-ignore된 stamp | 막힘(nil) | ignore 목록 검사(`:865-867`)와 base blob 부재(`:874-876`) 두 겹. RED에서 실패, GREEN에서 통과 관측 |
| 정규 파일이 아닌 것(FIFO, 장치) | 막힘(nil) | walk 단계에서 regular도 dir도 아니면 false. 단, stamp 자리의 FIFO는 그 전에 `checkCodemaps`의 `os.ReadFile`(`:408`)에서 멈출 수 있다. 이 부분은 PR 이전부터 있던 코드이고 git이 FIFO를 추적하지 못하므로 CI에서는 생기지 않는다 |
| 추적 내용과 작업 트리 불일치 | false | `git diff --name-only <base> -- dir`는 base 트리와 작업 트리를 비교한다. stamp는 따로 바이트 비교 |
| stamp 삭제 | layer absent → block | `checkCodemaps`가 먼저 absent를 돌려준다(`:408-412`). 안정성 함수도 읽기 실패로 nil |
| 추적되지 않은 본문 파일 | false | untracked 목록(`:861`)이 비지 않는다 |
| CRLF | 과잉 차단 쪽으로만 작동 | provenance.json은 `text`·`eol: lf` 속성(`git check-attr -a` 출력)이라 Windows 체크아웃에서도 LF다. 바이트가 달라지면 false가 되어 block할 뿐 통과하지 않는다. 본문 CRLF는 D4 |
| 모드만 바뀜 | false(block) | 모드 변경은 `git diff --name-only`에 나온다. 과잉 차단이지만 실제 변경이다 |
| base에 codemaps가 없음 | nil(block) | `git show base:<stamp>` 실패 |
| first parent 없음 | 출처 미기록 → block | `attachContribution`이 Contribution을 nil로 두고 안정성 함수를 부르지 않는다(`:822-827`) |
| git 실패 | nil(block) | 모든 `gitOutput` 오류가 nil을 돌려준다 |
| 측정 도중 stamp 재작성 | **통과 가능** | D2 (TOCTOU, Low) |
| index 비트(assume-unchanged/skip-worktree) | 본문은 통과, stamp는 바이트 비교로 막힘 | D4 |

과잉 차단 판단: 정상 CI 실행은 체크아웃(`text eol=lf`) → `mx scan`·`graph build`(출력은 `.moai/state/`와 gitignore된 `.moai/project/graph/`, 2회차 Claim F) → `graph check` 순서다. 이 저장소의 codemaps는 전부 `100644` 일반 파일이고(`git ls-files -s`), 이 트리에서 ignore된 파일도 없다(`git ls-files --others --ignored --exclude-standard -- .moai/project/codemaps/` 빈 출력). 손대지 않은 stamp는 여전히 warn이다(`TestGraphCheckPolicy_UntouchedStampStillWarns`, `BaseOverLinePrZeroWarns`, `InheritedPullRequestWarns` PASS). 추가 비용은 git 프로세스 2개와 codemaps 디렉터리(지금 8개 파일) walk 한 번이라 상수에 가깝다. `attachContribution`은 fresh 경로에서도 불리므로(`:457`) 매 실행마다 이 비용이 들지만 작다.

## Codex leg (audit_multi, 원문 복사)

호출: `mcp__moai__audit_multi(project_root=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001, target=baseBranch, focus=…)`. `gates` 인자와 `card_id`는 넘기지 않았다(지시된 파일 하나 외에는 기록하지 않기 위해서다).

| 항목 | 값 (원문) |
|---|---|
| overall_verdict | `fail` |
| gate_unmet | 결과에 필드 없음. `residual_risk_note: "required-backend FAIL: codex"` |
| plan_source | `config` |
| per_backend_verdicts | `codex` · gate `required` · verdict `fail` |
| disagreement_flag | `null` |
| participant_count | `1` |
| fail_open_backends | `[]` |
| audit_receipt | `rcpt-ebc3d97cd2fdfecd56edf95b` |
| build_commit | `b1806498fe` |
| tree_root | `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001` |

Codex summary 원문:

```
Verdict: fail
- [P2] stamp를 재독해 실제 판정에 사용한 바이트와 다른 내용을 검증할 수 있음 — internal/graph/check.go:870
  신뢰도: 높음. 최초 읽기 이후 stamp를 첫 부모 내용으로 복원하는 경쟁 조건을 재현하자, 되감긴 stamp로 계산한 결과가 `codemaps_inputs_stable=true`, `inherited`, CLI 종료 코드 0으로 통과했어.
  권고: `checkCodemaps`가 최초 읽은 바이트를 전달해 부모 blob과 비교하고, 검사 중 입력 변경도 차단해야 해.
- [P2] Git diff만으로 본문의 바이트 동일성을 증명할 수 없음 — internal/graph/check.go:857
  신뢰도: 높음. `assume-unchanged`·`skip-worktree`를 설정한 본문 변경과 CRLF 정규화에서 안정성이 `true`였어. 본문 변경은 CLI에서도 `inherited`와 종료 코드 0으로 통과했어. CHANGELOG와 네 로케일 문서의 바이트 동일성 설명도 실제 보장보다 강해.
  권고: 본문 파일 집합과 실제 바이트를 첫 부모 tree/blob에 직접 대조하고, 해당 우회 회귀시험을 추가해야 해.
- [P2] 새 조건식이 저장소의 staticcheck 검사를 실패시킴 — internal/graph/check.go:892
  신뢰도: 높음. `golangci-lint run --timeout=2m ./internal/graph/...`가 `QF1001: could apply De Morgan's law (staticcheck)`로 종료 코드 1을 반환했어.
  권고: 부정 조건을 `!d.IsDir() && !d.Type().IsRegular()`로 정리하고 같은 검사를 다시 실행해야 해.
```

이 감사와 Codex의 대응: P2-3 = D1(동의, blocking). P2-1 = D2(내용 동의, 도달 조건 때문에 이 감사는 Low·optional로 본다). P2-2 = D4(문서가 과장한다는 점은 동의, 기능 우회라는 판단에는 동의하지 않는다. 이유는 D4에 적었다). 판정 방향은 같다(둘 다 fail).

이 판정은 오케스트레이터 검증이 끝나기 전에는 최종이 아니다. 오케스트레이터가 위 digest 항목을 `<toplevel>/.moai/state/audit-plan-result.json`에 새로 쓰고 `moai verify audit-plan --project-root <toplevel> --result-file <그 경로>`를 돌려야 한다. 필수 backend가 `fail`이므로 그 검증도 PASS를 낼 수 없다.

## Audit plan

```
$ moai verify audit-plan --project-root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001
"config_status": "ok"
"model": "codex"
backends: claude gate off · codex gate required · glm gate off (source config.gates, explicit)
"cross_model_active": true
"cross_model_required": true
"enforced_required": [ "codex" ]
```

## 차원별 점수

| Dimension | Score | Verdict | Evidence |
|---|---|---|---|
| Functionality (40%) | 82/100 | PASS | N1 두 경로 차단을 RED 실패·GREEN 통과로 관측(Claim B, C). F2 회귀 없음. 잔여 D2(동시 작성자 필요)는 CI 경로에서 도달하지 않는다 |
| Security (25%) | 85/100 | PASS | 새 신뢰 경계 없음. 안정성 판정은 실패·불확실 모두 nil(fail closed). PR head를 빌드하는 워크플로라 게이트는 원래 악의적 작성자 경계가 아니다 |
| Craft (20%) | 55/100 | FAIL | `golangci-lint run ./internal/graph/` exit 1, QF1001(D1). 필수 CI `Lint` fail. 장치별 테스트 고정 부족(D3). gofmt 빈 출력, vet 호스트·Windows exit 0. coverage 미측정 |
| Consistency (15%) | 80/100 | PASS-WITH-DEBT | N2·N3 반영, 네 언어 문서가 같은 내용을 자연스러운 문장으로 담음. D4(본문 "바이트" 과장), D5(이중 "or") |

Must-pass(Functionality, Security)는 기준을 넘는다. 그러나 blocking 결함 D1이 필수 병합 검사를 막고, 필수 Codex gate가 `fail`이므로 → **FAIL**.

## Claim / Evidence

**Claim A — 고정 SHA와 범위**
```
$ git rev-parse HEAD
2fad4a8de6426f559d1c6d915c0b9e6a2c424e03
$ git merge-base HEAD origin/main
5c3f0ebf27584d2c45deb2d00f9d69a4eea932c4
$ git log --oneline 5bfd590da5..HEAD
2fad4a8de6 [t2001] docs(graph): document anchor-moved, the stability condition and codemaps_inputs_stable in CHANGELOG and four-locale graph pages (N3, card t2001)
64e4a8eb66 [t2001] docs(comments): state the real inherited-warn rule in the policy header, ClassifyVerdicts and DriftOrigin comments (N2, card t2001)
e3805e8766 [t2001] fix(graph): prove codemaps inputs equal the first parent byte for byte, and fail closed on links and ignored files (N1, card t2001)
4d434b4d78 [t2001] test(graph): pin that a rewound symlinked stamp target or an ignored stamp blocks as anchor-moved (N1, card t2001)
$ git diff --stat 5bfd590da5 HEAD
 CHANGELOG.md                                |   2 +-
 docs-site/content/en/cli-reference/graph.md |   2 +-
 docs-site/content/ja/cli-reference/graph.md |   2 +-
 docs-site/content/ko/cli-reference/graph.md |   2 +-
 docs-site/content/zh/cli-reference/graph.md |   2 +-
 internal/cli/graph_check_policy_test.go     | 102 ++++++++++++++++++++++++++++
 internal/graph/check.go                     |  59 ++++++++++++++--
 internal/graph/check_policy.go              |  19 ++++--
 8 files changed, 173 insertions(+), 17 deletions(-)
$ git diff --stat 5c3f0ebf27 HEAD | tail -1
 32 files changed, 1803 insertions(+), 101 deletions(-)
```

**Claim B — 범위 테스트 (macOS, 이 트리, HEAD 2fad4a8de6)**
```
$ uptime
 9:48  up 2 days,  5:21, 26 users, load averages: 41.92 37.05 37.76
$ go test ./internal/cli/ -run GraphCheck -count=1 -v
(26개 테스트 전부 --- PASS, FAIL 0) 그중:
--- PASS: TestGraphCheckPolicy_RewoundStampBlocks (1.88s)
--- PASS: TestGraphCheckPolicy_UntouchedStampStillWarns (1.72s)
--- PASS: TestGraphCheckPolicy_CrossedByChangeBlocks (1.73s)
--- PASS: TestGraphCheckPolicy_BaseOverLinePrZeroWarns (2.68s)
--- PASS: TestGraphCheckPolicy_InheritedPullRequestWarns (2.07s)
--- PASS: TestGraphCheckPolicy_SymlinkedStampTargetRewoundBlocks (1.10s)
--- PASS: TestGraphCheckPolicy_IgnoredStampBlocks (1.65s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	41.719s
$ go test ./internal/graph/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/graph	61.937s
$ go test ./internal/config/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	10.598s
```

**Claim C — 새 테스트가 RED 커밋에서 실제로 실패함 (비공허성)**

방법: `git archive -o <scratchpad>/red.tar 4d434b4d78` → `tar -xf … -C <scratchpad>/red` → `go -C <scratchpad>/red test …`. 저장소, stash, worktree는 건드리지 않았다.
```
$ go -C …/scratchpad/red test ./internal/cli/ -run 'SymlinkedStampTargetRewoundBlocks|IgnoredStampBlocks|UntouchedStampStillWarns' -count=1
--- FAIL: TestGraphCheckPolicy_SymlinkedStampTargetRewoundBlocks (0.97s)
    graph_check_policy_test.go:376: a rewound symlinked stamp must exit 1, got 0
        "verdict_class": "warn", "drift_origin": "inherited", "codemaps_inputs_stable": true, "contribution": 0, "value": 41
--- FAIL: TestGraphCheckPolicy_IgnoredStampBlocks (1.01s)
    graph_check_policy_test.go:420: an ignored stamp must exit 1, got 0
        "verdict_class": "warn", "drift_origin": "inherited", "codemaps_inputs_stable": true, "contribution": 0, "value": 41
FAIL	github.com/modu-ai/moai-adk/internal/cli	4.581s
```
(`UntouchedStampStillWarns`는 RED에서도 PASS.) 이 출력은 N1이 수리 전 코드에서 실제로 inherited·exit 0으로 통과했다는 첫 직접 관측이기도 하다.

**Claim D — 장치별 변이체 (D3)**

방법: `git archive -o <scratchpad>/head.tar 2fad4a8de6` → `<scratchpad>/mut`에 풀고, **스크래치 사본의** `check.go`만 고쳤다.
```
M1 (walk·ignore 검사 끔, 바이트 비교 유지)
$ go -C …/scratchpad/mut test ./internal/cli/ -run 'SymlinkedStampTargetRewoundBlocks|IgnoredStampBlocks|UntouchedStampStillWarns' -count=1 -v
--- PASS: TestGraphCheckPolicy_UntouchedStampStillWarns (1.72s)
--- PASS: TestGraphCheckPolicy_SymlinkedStampTargetRewoundBlocks (1.52s)
--- PASS: TestGraphCheckPolicy_IgnoredStampBlocks (1.00s)
ok  	github.com/modu-ai/moai-adk/internal/cli	5.730s

M2 (walk·ignore 검사 유지, git show와 바이트 비교 제거)
$ go -C …/scratchpad/mut test ./internal/cli/ -run 'SymlinkedStampTargetRewoundBlocks|IgnoredStampBlocks|UntouchedStampStillWarns|RewoundStampBlocks' -count=1 -v
--- PASS: TestGraphCheckPolicy_RewoundStampBlocks (1.67s)
--- PASS: TestGraphCheckPolicy_UntouchedStampStillWarns (1.17s)
--- PASS: TestGraphCheckPolicy_SymlinkedStampTargetRewoundBlocks (1.33s)
--- PASS: TestGraphCheckPolicy_IgnoredStampBlocks (1.32s)
ok  	github.com/modu-ai/moai-adk/internal/cli	7.216s
```

**Claim E — 정적 검사**
```
$ gofmt -l internal/graph internal/cli/graph_check_policy_test.go internal/cli/graph_check.go internal/lockfile internal/config
(빈 출력)
$ go vet ./internal/graph/ ./internal/cli/ ./internal/lockfile/ ./internal/config/
(빈 출력, exit 0)
$ env GOOS=windows go vet ./internal/graph/ ./internal/cli/ ./internal/lockfile/
(빈 출력, exit 0)
$ golangci-lint run ./internal/graph/
internal/graph/check.go:892:20: QF1001: could apply De Morgan's law (staticcheck)
		if err != nil || !(d.IsDir() || d.Type().IsRegular()) {
		                 ^
1 issues:
* staticcheck: 1
(exit 1)
$ golangci-lint version
golangci-lint has version v2.1.6 built with go1.26.8 …
```

**Claim F — CI 상태 (읽기만, 재실행·취소 없음)**
```
$ gh run view 41555114444 --json headSha,status,conclusion,event
{"conclusion":"","event":"pull_request","headSha":"2fad4a8de6426f559d1c6d915c0b9e6a2c424e03","status":"in_progress"}
$ gh pr checks 1955   (exit 1, 발췌)
Lint	fail	4m6s	https://github.com/modu-ai/moai-adk/actions/runs/41555114444/job/125652130399
Multi-OS Test (macos-latest)	pass	6m28s
Multi-OS Test (windows-latest)	pending	0
Test (ubuntu-latest)	pending	0
Race Test (internal/cli 1 of 4)	pending	0
Race Test (internal/cli 3 of 4)	pending	0
Race Test (remainder packages)	pending	0
Test (browser fire guard)	pending	0
Build (windows/amd64)	pass	2m55s
Build (linux/amd64)	pass	2m49s
Analyze (Go) (go)	pass	2m5s
graph-freshness	pass	1m28s
Race Test (internal/cli 2 of 4)	pass	10m51s
Race Test (internal/cli 4 of 4)	pass	10m29s
Validate 4-locale parity	pass	29s
Template Neutrality Audit	pass	37s
$ gh api repos/modu-ai/moai-adk/check-runs/125652130399/annotations   (발췌)
"path":"internal/graph/check.go","start_line":892,"start_column":20,"annotation_level":"failure","message":"QF1001: could apply De Morgan's law (staticcheck)"
$ gh run view 41555114444 --job 125652130399 --log-failed
run 41555114444 is still in progress; logs will be available when it is complete
$ gh api repos/modu-ai/moai-adk/branches/main/protection/required_status_checks --jq .contexts
["Lint","Build (linux/amd64)","Analyze (Go) (go)"]
$ gh run list --branch WT-cc-bash-monitor --limit 5
completed success … Release PR Multi-OS Verification … 41555114343
completed success … docs i18n parity check … 41555114305
in_progress       … PR Multi-OS Gate … 41555114270   (Windows 잠금 테스트를 돌릴 job)
completed success … Template Neutrality Check … 41555114275
in_progress       … CI … 41555114444
```

**Claim G — 설정 사본과 주석·문서**
```
$ diff .moai/config/sections/gate.yaml internal/template/templates/.moai/config/sections/gate.yaml
62c62  enabled: true / enabled: false
70c70  rules_dir: ".moai/astgrep-rules" / ".moai/config/astgrep-rules"
```
두 줄 모두 PR 이전부터 있던 차이이고, `graph_freshness` 블록(live `:80-` 이하)은 두 사본이 같다. 수리 커밋은 gate.yaml을 건드리지 않았다. 주석(`check_policy.go:8-17`, `:47-50`, `:62-79`, `:103-112`, `check.go:93-104`, `:838-851`)은 `driftOrigin`의 실제 순서(originated → 안정성 nil/false면 anchor-moved → 산술)와 맞는다. 네 언어 문서는 `anchor-moved`, `codemaps_inputs_stable`, warn 조건 네 가지를 같은 구성으로 담고, 식별자는 원문 그대로이며 번역투 없이 읽힌다(ko·ja·zh 본문을 diff로 전부 읽음).

**Claim H — 판정 도구 빌드 provenance (§2.2)**
```
$ moai version
 v3.2.0-rc.39   v3.2.0-1041-gdad3603de1   built 2026-11-02T19:17:51Z
$ git merge-base --is-ancestor dad3603de1 HEAD
(exit 1)
$ git merge-base --is-ancestor b1806498fe HEAD
(exit 1)
```

## Baseline-attribution

모든 명령은 이 실행에서 `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t2001`, HEAD `2fad4a8de6426f559d1c6d915c0b9e6a2c424e03`를 대상으로 돌렸다. `go test`, `go vet`, `gofmt`, `golangci-lint`는 트리 소스를 직접 컴파일하므로 판정 빌드가 곧 트리 HEAD다. RED 관측은 `4d434b4d78`의 `git archive` 사본, 변이체 관측은 `2fad4a8de6`의 `git archive` 사본을 고친 것이다. `moai verify audit-plan`은 설치 빌드 `dad3603de1`, Codex leg는 MCP 서버 빌드 `b1806498fe`가 판정했다. 둘 다 HEAD의 조상이 아니므로(`--is-ancestor` exit 1) 이 트리에서 만든 빌드가 아니다. Codex는 Go 코드를 직접 읽고 돌렸다고 보고했다. CI 결과는 head `2fad4a8de6` 기준 run 41555114444, 41555114270이다.

## Gaps (관측하지 않은 것)

- **Windows 실행은 아직 관측하지 못했다.** `Multi-OS Test (windows-latest)`(run 41555114270)와 `Test (ubuntu-latest)`는 이 감사 시점에 `pending`이었다. 진행 중인 run은 통과가 아니다. `Build (windows/amd64)` pass와 `GOOS=windows go vet` exit 0만 확인했다. 따라서 F1의 Windows 잠금 코드와 `TestWindowsLockAllowsDataIOWhileExcludingContenders`는 여전히 Windows에서 실행된 기록이 없다.
- **coverage는 측정하지 않았다**(부하가 높아 범위 테스트만 돌렸다).
- `Lint` job 로그 본문은 run이 진행 중이라 읽지 못했다(`logs will be available when it is complete`). 실패 원인은 check-run annotation과 로컬 재현으로 확인했다.
- **D2(재독 경쟁)와 D4(index 비트·CRLF 본문)는 직접 재현하지 않았다.** 소스 읽기와 Codex 보고에 기댔다. D4에서 "anchor도 git 관점이라 산술이 유지된다"는 판단은 소스 추론이다.
- `.moai/project/codemaps/*.md` 본문 파일의 eol 속성은 확인하지 않았다(provenance.json만 `git check-attr -a`로 확인).
- `golangci-lint`는 `internal/graph`만 돌렸다. CI annotation이 그 한 곳만 가리키지만 `internal/cli` 테스트 파일은 로컬 lint를 돌리지 않았다.
- 2회차에서 본 hook·lockfile 범위 테스트는 수리 커밋이 그 파일을 건드리지 않아 다시 돌리지 않았다. `GOOS=windows go vet ./internal/lockfile/`만 다시 돌렸다.
- 도구 거부: 이번 실행에서 거부된 명령은 없었다.

## Residual-risk

- D1만 고치면 병합 검사는 풀릴 가능성이 높지만, 필수 Codex gate가 D2·D4를 P2로 계속 보고하면 다음 감사도 Codex leg에서 `fail`이 날 수 있다. D2는 수정 비용이 작다(읽은 바이트를 넘기기).
- 안정성 판정은 여전히 "PR이 판정 코드와 임계값을 바꾸지 않았다"는 전제에 기댄다(N4). PR head를 빌드하는 워크플로에서는 피할 수 없다.
- 정책상 inherited drift가 PR에서 warn으로 통과하면 병합 뒤 main push 실행은 block한다. 누군가 codemaps를 다시 만들 때까지 main이 빨간색으로 남을 수 있다(R208이 의도한 trade-off).
- Windows `LockFileEx` 관련 잔여 위험(핸들 단위 Unlock, 옛 빌드와 섞인 실행)은 2회차 그대로다.
- 스크래치패드(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/2dca407b-ed1e-41af-b4a9-a4696467b9ee/scratchpad/red`, `…/mut`, `red.tar`, `head.tar`)는 이 파일에 옮긴 줄 외에는 인용 근거가 아니며 세션이 끝나면 사라진다.

## Iteration history

- 1회차 (`sync-audit-1.md`): FAIL @ b2387e7123 — blocking F1·F2, Codex required gate `fail` (rcpt-9ac33515965037b2a13e2e8c).
- 2회차 (`sync-audit-2.md`): FAIL @ 7a946cfca2 — F1·F2·F3-F6 CLOSED, NEW N1(blocking)·N2·N3·N4, Codex `fail` (rcpt-b5d843df9399068fbf1d6832).
- 3회차 (이 파일): FAIL @ 2fad4a8de6 — N1·N2·N3 CLOSED, NEW D1(blocking, 필수 `Lint` 실패)·D2·D3·D4·D5, Codex `fail` (rcpt-ebc3d97cd2fdfecd56edf95b).

## 권고

1. D1: `check.go:892`를 `if err != nil || (!d.IsDir() && !d.Type().IsRegular()) {`로 고치고 `golangci-lint run ./internal/graph/` → `0 issues.`를 확인한다.
2. D2: `checkCodemaps`가 읽은 stamp 바이트를 안정성 함수로 넘겨 그 바이트를 비교한다(Codex P2-1 해소).
3. D4: 문서·주석을 실제 보장에 맞추거나 본문도 바이트 비교한다(Codex P2-2 해소). D3·D5는 같은 커밋에서 정리할 수 있다.
4. 수정 뒤 `Multi-OS Test (windows-latest)` 완료를 기다려 F1의 Windows 실행 결과를 확인한다.

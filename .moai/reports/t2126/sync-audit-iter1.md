auditor-model: claude-opus-5-5

verdict: PASS
audited_sha: 8ecd4bfcc0c0d058c8b0a6a2244e867b7a89e33b

# t2126 sync 감사 판정 (iteration 1)

- 카드: t2126 (Class B, SPEC 없음) — PR #1969 Codex P2, checkpoint `GitRemote.git`의 GIT_DIR 상속
- 트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2126`, 브랜치 `WT-pr-codex-repair-3`
- 기준: `04d8c27a7e` → 감사 대상 HEAD `8ecd4bfcc0c0d058c8b0a6a2244e867b7a89e33b`
- 판정: **PASS**, 점수 **0.93**
- Codex(required) 1회: verdict `pass`, findings 0건, receipt `rcpt-ee212793db05707dcbc2a1da`

## 평가 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 96/100 | PASS | 재현 테스트가 변이(수정 줄 제거)에서 FAIL, HEAD에서 PASS — 아래 증거 E2/E3 |
| Security (25%) | 90/100 | PASS | 수정이 저장소 위치 변수 11종을 제거하는 기존 `gitenv.Env()`를 재사용. 같은 결함 형태의 형제 호출 1곳(`ctxmodel.GitVerifier.Lookup`)이 범위 밖에 남음 — F1 (outside, non-blocking) |
| Craft (20%) | 92/100 | PASS | checkpoint 패키지 커버리지 90.6%, gofmt 깨끗, vet exit 0 |
| Consistency (15%) | 95/100 | PASS | 같은 저장소의 `mergeq/gitcli.go`가 이미 쓰는 `cmd.Env = gitenv.Env()` 패턴과 동일 |

Must-pass(Functionality, Security) 모두 통과. 가중 합 ≈ 0.93.

## 결함 목록

- **F1** [Medium, 신뢰도 high] [optional] [outside] `internal/agenticfactory/ctxmodel/ctxmodel.go:310` — `GitVerifier.Lookup`이 `exec.Command("git", "cat-file", "blob", ...)`에 `cmd.Dir`만 지정하고 `cmd.Env`를 두지 않아 호출자의 GIT_DIR/GIT_OBJECT_DIRECTORY를 그대로 상속한다. 상속된 GIT_DIR이 있으면 다른 저장소의 객체로 "존재/해시"를 답할 수 있다(이번 카드와 같은 읽기 형태 결함). 소스 읽기로 확인, 재현 실행은 하지 않음(가설 수준 결함 주장이 아님: 코드상 Env 미지정은 관측 사실, 악용 경로 재현은 미실행). Required fix(후속 카드): `cmd.Env = gitenv.Env()` 추가 + 동형 재현 테스트. 이 카드 diff 밖이라 판정에 영향 없음, 리더가 운영자에게 표시.
- **F2** [Low, 신뢰도 medium] [optional] [outside] `internal/agenticfactory/effect/internal/procrun/procrun.go:77` — `cmd.Env = append(append(os.Environ(), s.Base...), s.Env...)`로 임의 바이너리를 실행한다. 그 자식이 git을 부르면 같은 상속이 생길 수 있다. 호출처가 git을 실행하는지는 확인하지 않음(가설). Required fix: 호출처 확인 후 필요 시 `gitenv.Env()` 기반으로.
- **F3** [Low, 신뢰도 low] [optional] [outside] `internal/agenticfactory/browser/resolve.go:140` — 브라우저 탐지 probe가 부모 env를 상속한다. git이 아닌 브라우저 실행 파일 대상이라 실질 위험은 낮다고 판단(가설).
- **F4** [Info] [optional] [in-diff] 재현 테스트는 `Head`(ls-remote) 경로에서 `ErrNotAcked`로 끝나므로 `Contains`(fetch + merge-base) 경로는 직접 행사하지 않는다. 두 경로 모두 같은 `g.git()`을 거치므로 수정 한 줄이 양쪽을 덮는 것은 코드상 확인됨. 추가 테스트는 필수 아님.

blocking 결함 없음.

## 증거 (Claim / Evidence / Baseline-attribution / Gaps / Residual-risk)

### Claim
1. 수정(954a2e42a3)은 `GitRemote.git`에 `c.Env = gitenv.Env()` 한 줄 + import만 추가한 최소 변경이다.
2. `gitenv.Env()`는 GIT_DIR을 제거한다.
3. 재현 테스트는 수정이 없으면 실패하고, 수정이 있으면 통과한다.
4. 대상 패키지 테스트와 vet이 통과한다.

### Evidence
- E0 트리 확인: `git rev-parse --show-toplevel` → `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2126`; `git rev-parse HEAD` → `8ecd4bfcc0c0d058c8b0a6a2244e867b7a89e33b`; `git status --short` → 빈 출력(감사 시작·종료 모두).
- E1 diff: `git diff 04d8c27a7e HEAD --stat`
  ```
   .moai/reports/t2126/progress.md                    | 22 +++++++
   internal/agenticfactory/checkpoint/checkpoint.go   |  4 ++
   .../checkpoint/gitremote_env_test.go               | 69 ++++++++++++++++++++++
   3 files changed, 95 insertions(+)
  ```
  production 변경: import `github.com/modu-ai/moai-adk/internal/gitenv` + 주석 2줄 + `c.Env = gitenv.Env()`.
  `internal/gitenv/gitenv.go:53-65` `RepoScopingVars`에 `"GIT_DIR"` 포함; `Env()` = `Scrub(os.Environ())`(122-124), `Scrub`은 항상 non-nil 반환(72-75) — nil Env로 상속이 되살아나는 경로 없음.
- E2 변이 프로브(**실행됨**): `git archive` HEAD → 스크래치 추출, 316행 `c.Env = gitenv.Env()`를 `_ = gitenv.Env // MUTANT: fix line reverted`로 교체(카드 트리 미수정).
  `go test -C <scratch>/mutant -count=1 -run TestConfirmRemoteIgnoresInheritedGitDir ./internal/agenticfactory/checkpoint` → exit 1
  ```
  --- FAIL: TestConfirmRemoteIgnoresInheritedGitDir (1.33s)
      gitremote_env_test.go:64: ConfirmRemote err = <nil>, want ErrNotAcked (the checkpoint repository's origin lacks the branch)
  FAIL
  FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/checkpoint	1.696s
  ```
- E3 HEAD 단일 테스트(**실행됨**): `go test -count=1 -v -run TestConfirmRemoteIgnoresInheritedGitDir ./internal/agenticfactory/checkpoint` → exit 0
  ```
  --- PASS: TestConfirmRemoteIgnoresInheritedGitDir (1.23s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/checkpoint	1.555s
  ```
- E4 `go test -count=1 ./internal/agenticfactory/checkpoint ./internal/gitenv` → exit 0
  ```
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/checkpoint	9.687s
  ok  	github.com/modu-ai/moai-adk/internal/gitenv	13.420s
  ```
- E5 `go vet ./internal/agenticfactory/...` → `vet_exit=0`, 출력 없음.
- E6 `go test -count=1 -cover ./internal/agenticfactory/checkpoint` → `ok ... coverage: 90.6% of statements`.
- E7 `gofmt -l` 변경 두 파일 → 출력 없음, exit 0.
- E8 형제 탐색: `grep -rn --include='*.go' -E 'exec\.Command(Context)?\(' internal/agenticfactory | grep -v _test.go`
  ```
  internal/agenticfactory/checkpoint/checkpoint.go:312:	c := exec.Command("git", args...)
  internal/agenticfactory/effect/internal/procrun/procrun.go:75:	cmd := exec.CommandContext(rctx, s.Bin, s.Args...)
  internal/agenticfactory/browser/resolve.go:140:	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
  internal/agenticfactory/ctxmodel/ctxmodel.go:310:	cmd := exec.Command("git", "cat-file", "blob", commit+":"+path)
  internal/agenticfactory/mergeq/gitcli.go:91:	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
  ```
  `mergeq/gitcli.go:96`은 `cmd.Env = append(gitenv.Env(), ...)` — 이미 안전. `ctxmodel.go:310`은 Env 미지정 → F1.
- E9 감사 계획: `moai verify audit-plan --project-root <toplevel>` → `config_status: ok`, `model: codex`, codex `gate: required`, claude/glm `off`, `cross_model_active: true`, `enforced_required: ["codex"]`.
- E10 Codex: `mcp__moai__codex_audit(mode=native, target=baseBranch, delta_base=04d8c27a7e, project_root=<toplevel>)` → `verdict: pass`, `findings: []`, `audit_receipt: rcpt-ee212793db05707dcbc2a1da`, `resolved_model: gpt-6.1-sol`, `review_base: delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)`.

### Baseline-attribution
- 모든 go 명령은 이 감사 실행에서 카드 트리(HEAD 8ecd4bfcc0) 또는 그 HEAD의 `git archive` 사본(변이 1줄만 다름)에 대해 실행.
- 도구 출처(§2.2): `moai` 바이너리와 MCP 서버는 commit `04d8c27a7e`(기준 커밋, HEAD의 조상)로 빌드됨 — codex_audit 결과의 `build_lag`가 이를 명시. 이 카드의 변경은 audit-plan/codex_audit 코드 경로를 건드리지 않으므로(diff는 checkpoint 패키지뿐) 판정 도구 자체의 동작에는 영향이 없다고 판단하나, 엄밀히는 "기준 커밋 빌드가 판정"이다. go test/vet 결과는 트리 소스를 직접 컴파일했으므로 lag 대상 아님.

### Gaps
- 작업 트리 격리 가드가 명령 두 건을 거부함(실행 전 거부): (a) test와 vet을 `${pipestatus}`로 묶은 복합 명령, (b) 셸 변수 `$S`를 쓰는 sed 명령. 둘 다 단순 명령으로 분리해 재실행했고 결과는 위 E2/E4/E5. 대체는 동일 명령의 분리 실행이며 소스 읽기 대체는 없음.
- F1~F3은 코드 읽기로만 확인; 상속 GIT_DIR로 실제 오답이 나는지 재현 실행은 하지 않음.
- 전체 `go test ./internal/agenticfactory/...`는 이번 감사에서 재실행하지 않음(progress.md의 lane 측 기록은 관측하지 않은 이월 값이라 인용하지 않음).
- Windows 대소문자 접기 경로는 이 호스트(darwin)에서 `scrub(…, true)` 단위 테스트로만 간접 확인(E4 gitenv ok), Windows 실행 없음.
- Codex 호출 provenance의 `transport`가 `codex app-server`로 표기됨. 감사자는 app-server를 직접 호출하지 않았고 운영자가 허용한 MCP `codex_audit` 1회만 호출했으나, MCP 서버 내부 전송이 app-server임을 그대로 기록함. 운영자 규칙(d-20261106T064700Z-de2d)의 "no app-server" 문구가 이 내부 전송까지 금하는지는 리더 판단 사항.
- Codex 요약은 "테스트·빌드·재현 명령은 실행하지 않았다"고 명시 — Codex 판정은 정적 검토만임.

### Residual-risk
- 같은 프로세스에서 `gitenv` 목록 밖의 변수(예: `GIT_CONFIG_PARAMETERS`의 `remote.origin.url` 주입, `GIT_CONFIG_GLOBAL`이 가리키는 설정의 `url.<base>.insteadOf`)로도 원격 해석을 바꿀 수 있다. gitenv는 동작·설정 변수를 의도적으로 유지하므로(gitenv.go:45-52) 이 카드 범위 밖이며, ACK 위조 경로로서의 실재성은 검증하지 않음.
- F1의 ctxmodel 형제 결함은 남아 있음(후속 카드 필요).

## 반복 이력
- iter1 (이 파일): PASS 0.93 @8ecd4bfcc0, Codex pass receipt `rcpt-ee212793db05707dcbc2a1da`.

auditor-model: claude-opus-5-5

verdict: FAIL
audited_sha: 5e02a7a780c3605fc3ad7064f795e611f1095765

## Evaluation Report
SPEC: t2136 (Class B, SPEC 없음 — t2126 후속)
Overall Verdict: FAIL — 사유는 코드가 아니라 **Codex required 게이트 미충족(fail-closed)** 하나뿐입니다. Claude 단독 코드 평가는 PASS 수준(점수 0.94, blocking finding 0건)입니다.
판정 정정 경위: 처음에는 PASS로 기록했으나, SubagentStop 훅이 `AUDIT_RECEIPT_VIOLATION`(영수증 없는 PASS)으로 거부했습니다. 리더 지시상 이 카드에서 Codex는 승인되지 않았으므로 Codex를 호출해 게이트를 채우지 않았고, 대신 판정을 게이트 기준(FAIL)으로 바로잡았습니다. PASS로 바꾸려면 (a) Codex 감사를 승인받아 실행한 뒤 영수증을 인용하거나, (b) 운영자가 게이트 예외를 기록해야 합니다. 코드 수정은 필요 없습니다.
감사 범위: `04d8c27a7e..5e02a7a780` (3커밋: RED 8dd053e9dc → 수정 6645242e80 → 기록 5e02a7a780)

Codex 게이트 관련 주의: 감사 계획은 codex를 `required`로 보고하지만 이 카드에서 Codex 사용은 승인되지 않았습니다. 따라서 cross-model 수렴 게이트는 **충족되지 않았고(fail-closed)**, 아래 PASS는 Claude 단독 판정입니다. 게이트 처분은 리더·운영자 몫입니다(Gaps G1).

### Dimension Scores
| Dimension | Score | Verdict | Evidence |
|-----------|-------|---------|----------|
| Functionality (40%) | 97/100 | PASS | 변이체(수정 줄 제거) `mutant exit=1` / `--- FAIL: TestGitVerifierLookupIgnoresInheritedGitDir (1.42s)` / `gitverifier_env_test.go:51: Lookup found the blob (sha256 f3552474efba84643730245ed27003d7f865a896cb25ba3c340dffc5d6f5b9aa) in another repository through an inherited GIT_DIR`; HEAD `head exit=0` / `--- PASS: TestGitVerifierLookupIgnoresInheritedGitDir (0.53s)` / `ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel	0.808s` |
| Security (25%) | 95/100 | PASS | 수정은 신뢰 경계(상속 환경변수) 축소 방향. `internal/gitenv/gitenv.go:54-64` RepoScopingVars에 `GIT_DIR`, `GIT_COMMON_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE`, `GIT_OBJECT_DIRECTORY`, `GIT_ALTERNATE_OBJECT_DIRECTORIES` 등 포함; `Env()`=`Scrub(os.Environ())`(122-124). 새 입력 표면 없음, 비밀값 없음 |
| Craft (20%) | 90/100 | PASS | `go test -count=1 ./internal/agenticfactory/ctxmodel ./internal/gitenv` → `ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel	1.106s` / `ok  	github.com/modu-ai/moai-adk/internal/gitenv	6.031s`; `go vet ./internal/agenticfactory/ctxmodel` → `vet exit=0`; `gofmt -l` → 출력 없음 `gofmt exit=0`. 커버리지 수치는 측정 안 함(G3) |
| Consistency (15%) | 95/100 | PASS | 동일 패턴 기존 사용처: `internal/agenticfactory/mergeq/gitcli.go:96` `cmd.Env = append(gitenv.Env(),`; 패키지 TestMain이 이미 `gitenv.ScrubProcess()` 호출(`ctxmodel/main_test.go:16`). 수정은 import 1줄 + 주석 2줄 + 대입 1줄 |

Must-pass(Functionality, Security) 모두 통과.

### Findings (structured defect-list)
- F1 [Low, 확신 높음] [optional] [outside] `internal/agenticfactory/checkpoint/checkpoint.go:311` — `GitRemote.git`이 `exec.Command("git", args...)` + `c.Dir = g.Dir`만 쓰고 `Env`를 지정하지 않음. 같은 결함 계열(상속 GIT_DIR이 `Dir`을 무시하게 만듦). 이 카드 diff 밖이므로 판정에 영향 없음. Required fix: 별도 후속 카드에서 `c.Env = gitenv.Env()` + 재현 테스트(본 카드와 같은 방식).
- F2 [Info, 확신 낮음] [optional] [outside] `internal/agenticfactory/effect/internal/procrun/procrun.go:75`(`exec.CommandContext(rctx, s.Bin, s.Args...)`), `internal/agenticfactory/browser/resolve.go:140`(`exec.CommandContext(ctx, name, args...)`) — 임의 바이너리 실행 경로. git을 직접 실행하는지, 자식이 git을 부르는지는 확인 안 함(`gitenv.Env` 문서의 "arbitrary process that may run git of its own" 범주 후보). Required fix: 해당 없음 — 필요 시 리더가 조사 카드로 분류.
- F3 [Low, 확신 낮음] [optional] [in-diff] `gitverifier_env_test.go:19` — 픽스처가 `GIT_CONFIG_GLOBAL=/dev/null`, `GIT_CONFIG_SYSTEM=/dev/null`을 씀. Windows CI에서의 동작은 관측하지 않음(Git for Windows는 `/dev/null`을 특수 처리하는 것으로 알려져 있으나 이 실행에서 검증 안 함). Required fix: 없음 — Windows CI 결과로 확인.
- F4 [Info] [optional] [outside, 기존 동작] `ctxmodel.go:318-322` — `Lookup`은 모든 `*exec.ExitError`를 "not found"(nil 오류)로 처리. 따라서 `Dir`이 저장소가 아니어도 not-found로 보임. 이번 diff가 바꾼 동작이 아니며 수정으로 악화되지도 않음. 테스트의 "found=false, err=nil" 기대가 이 성질에 기대고 있다는 점만 기록.

blocking finding 없음.

### 정합성·최소성 판단
- diff 전체(`git diff 04d8c27a7e HEAD -- internal`): 프로덕션 변경은 `ctxmodel.go`의 import 1줄과 `cmd.Env = gitenv.Env()` 1줄(+주석 2줄)뿐. t2126과 같은 한 줄 패턴. 범위 밖 수정 없음.
- 커밋 순서: `git log --oneline 04d8c27a7e..HEAD`에서 RED(8dd053e9dc)가 수정(6645242e80)보다 먼저 — 재현 우선 순서를 커밋 그래프가 증언함(§2.3 충족).
- 테스트 설계: 두 개의 실제 저장소(clone은 커밋 없음, other는 있음) + `t.Setenv("GIT_DIR", other/.git)`. 변이체에서 실제로 실패하고 HEAD에서 통과하므로 수정 줄을 판별함. `t.Setenv` 사용으로 병렬 아님(주석에 명시). 픽스처 git은 Setenv 이전에 실행되고 TestMain이 프로세스 환경을 이미 정리하므로 픽스처 자체가 외부 저장소에 쓰는 위험 없음.

### Recommendations
- F1(`checkpoint.go:311`)을 같은 패턴의 후속 카드로 발행할지 리더가 결정.
- Codex required 게이트 미충족을 리더가 처분(승인 후 실행 또는 게이트 예외 기록).

---

## Evidence-bearing report

### Claim
1. `GitVerifier.Lookup`은 수정 전 상속 GIT_DIR이 가리키는 다른 저장소에서 blob을 찾았다.
2. `cmd.Env = gitenv.Env()`가 그것을 막는다.
3. 새 테스트는 수정 줄을 판별한다(없으면 실패, 있으면 통과).
4. 대상 패키지 테스트·vet·gofmt가 깨끗하다.
5. `gitenv.Env()`는 GIT_DIR을 제거한다.

### Evidence
- 변이체 probe — **실행함(executed)**:
  - `git -C <tree> archive -o <scratchpad>/t2136-head.tar HEAD` → 출력 없음(성공)
  - `tar -xf <scratchpad>/t2136-head.tar -C <scratchpad>/t2136-mutant` → 출력 없음
  - 스크래치 사본에서만 `ctxmodel.go:315` `cmd.Env = gitenv.Env()` → `_ = gitenv.Env // mutant: fix line reverted`로 교체(작업 트리는 건드리지 않음)
  - `go test -C <scratchpad>/t2136-mutant -count=1 -run IgnoresInherited ./internal/agenticfactory/ctxmodel`:
    ```
    mutant exit=1
    --- FAIL: TestGitVerifierLookupIgnoresInheritedGitDir (1.42s)
        gitverifier_env_test.go:51: Lookup found the blob (sha256 f3552474efba84643730245ed27003d7f865a896cb25ba3c340dffc5d6f5b9aa) in another repository through an inherited GIT_DIR
    FAIL
    FAIL	github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel	1.849s
    ```
- HEAD 대상 테스트 — 실행함: `go test -C <tree> -count=1 -run IgnoresInherited -v ./internal/agenticfactory/ctxmodel`:
  ```
  head exit=0
  === RUN   TestGitVerifierLookupIgnoresInheritedGitDir
  --- PASS: TestGitVerifierLookupIgnoresInheritedGitDir (0.53s)
  PASS
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel	0.808s
  ```
- 패키지 테스트 — 실행함: `go test -C <tree> -count=1 ./internal/agenticfactory/ctxmodel ./internal/gitenv`:
  ```
  ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/ctxmodel	1.106s
  ok  	github.com/modu-ai/moai-adk/internal/gitenv	6.031s
  ```
- vet/format — 실행함: `go -C <tree> vet ./internal/agenticfactory/ctxmodel` → `vet exit=0`; `gofmt -l internal/agenticfactory/ctxmodel` → 출력 없음, `gofmt exit=0`.
- GIT_DIR 제거 — 소스 읽기(derived, 단 gitenv 패키지 테스트 `ok`로 보강): `gitenv.go:54` `"GIT_DIR",`가 RepoScopingVars에 있고, `scrub`(91-114)는 해당 이름을 건너뜀, `Env()`(122-124)는 `Scrub(os.Environ())`. 변이체/HEAD 대비가 이 경로를 행동으로 증명함(executed).
- 잔존 미수정 git exec 목록: `grep -rn 'exec\.Command' internal/agenticfactory --include='*.go' | grep -v _test.go` →
  ```
  internal/agenticfactory/checkpoint/checkpoint.go:311:	c := exec.Command("git", args...)
  internal/agenticfactory/effect/internal/procrun/procrun.go:75:	cmd := exec.CommandContext(rctx, s.Bin, s.Args...)
  internal/agenticfactory/browser/resolve.go:140:	out, err := exec.CommandContext(ctx, name, args...).CombinedOutput()
  internal/agenticfactory/ctxmodel/ctxmodel.go:311:	cmd := exec.Command("git", "cat-file", "blob", commit+":"+path)
  internal/agenticfactory/mergeq/gitcli.go:91:	cmd := exec.CommandContext(ctx, "git", append([]string{"-C", dir}, args...)...)
  ```
  `checkpoint.go:300-321`을 읽어 Env 미지정 확인(F1); `mergeq/gitcli.go:96`은 `cmd.Env = append(gitenv.Env(),`로 이미 정리됨.
- 감사 계획: `moai verify audit-plan --project-root <tree>` → `"config_status": "ok"`, `"model": "codex"`, codex `"gate": "required"`, `"cross_model_required": true`, `"enforced_required": ["codex"]`.

### Baseline-attribution
- 트리: `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2136` (`git rev-parse --show-toplevel`로 확인), HEAD `5e02a7a780c3605fc3ad7064f795e611f1095765`, `git status --short` 깨끗함 — 감사 시작과 끝에 동일하게 재확인.
- 비교 기준: base `04d8c27a7e`. 변이체는 HEAD를 `git archive`한 사본에서 수정 줄만 되돌린 것.
- Go 툴체인: 트리 소스로 직접 컴파일(`go test`/`go vet`), 설치된 moai 바이너리는 판정에 쓰지 않음. 감사 계획 조회만 설치 바이너리(`/Users/goos/go/bin/moai`, v3.2.1-rc.2, MCP 서버 빌드 커밋 04d8c27a7e = base)로 수행 — 계획 조회용이며 코드 판정 근거 아님(§2.2).

### Gaps
- G1 **Codex required 게이트 미충족(fail-closed)**: 감사 계획이 codex를 required로 강제하지만 이 카드에서 Codex가 승인되지 않아 `codex_audit`/`audit_multi`를 호출하지 않음. 영수증 없음(`receipts=none`). 수렴 게이트 기준으로는 미통과이며 처분은 리더·운영자 몫.
- G2 워크트리 격리 가드가 `cd <dir> && ... go test -run TestGitVerifier...` 형태와 `git archive | tar` 파이프 형태를 "git 명령이 너무 복잡함"으로 거부함(3회). 대신 `git -C <tree> archive -o <file> HEAD`, `tar -xf`, `go test -C <dir> -run IgnoresInherited`로 나눠 실행. 측정 자체는 대체 형태로 실제 수행했으며 추론으로 바꾼 부분은 없음. `-run` 패턴은 `IgnoresInherited`로 같은 단일 테스트를 선택(출력의 테스트 이름으로 확인).
- G3 커버리지(`go test -cover`)와 `golangci-lint`는 실행하지 않음.
- G4 `internal/agenticfactory` 전체 트리와 Windows/Linux 크로스 빌드는 실행하지 않음(F3 Windows 동작 미관측).
- G5 RED 커밋 8dd053e9dc 자체를 체크아웃해 실행하지는 않음; 그 대신 HEAD 사본에서 수정 줄만 되돌린 변이체로 동등한 조건을 실행함(import는 `_ = gitenv.Env`로 유지 — 동작상 차이 없음은 derived).
- G6 F2의 두 실행 경로가 git을 부르는지 조사하지 않음.

### Residual-risk
- 같은 결함 계열이 `checkpoint.go:311`(F1)에 남아 있어, 훅 문맥에서 `GitRemote`가 다른 저장소를 읽을 수 있음(이 카드 범위 밖).
- `Lookup`이 모든 git 종료 오류를 not-found로 삼키므로(F4), 환경 문제로 인한 실패가 "없음"으로 보고되는 성질은 그대로임.
- 테스트 출력 원본은 스크래치(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/912f2297-cbe5-4eff-adfd-454c218a698a/scratchpad/mutant.out`, `head.out`)에만 있음 — 판정에 필요한 줄은 이 파일에 옮겨 적었고, 스크래치 파일 자체는 인용 대상이 아님(유실 가능).

### Iteration history
- iter1 (이 파일): 처음 PASS 0.94 @5e02a7a780(Claude 단독)으로 기록 → SubagentStop 훅 `AUDIT_RECEIPT_VIOLATION` 거부 → Codex 미승인으로 호출하지 않고 판정을 FAIL(게이트 미충족 사유만)로 바로잡음. 코드 평가 0.94와 findings는 그대로.

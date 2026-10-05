# t532 — glm_task 백그라운드 잡이 요청 스코프 컨텍스트를 물려받아 즉시 죽는다

- **카드**: t532 (t514 파생, Class B, Tier S~M)
- **브랜치**: `WT-glm-bg-context`
- **트리**: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t532`
- **기준 커밋**: 로컬 develop tip `6a46c0edb` (`git merge develop --ff-only` 로 fast-forward, 병합 커밋 없음)

---

## Claim

1. `glm_task` 의 `background=true` 잡은 요청 스코프 컨텍스트에서 파생돼, MCP 호스트가 핸들러 반환 직후 그 컨텍스트를 끝내면 **밀리초 단위로 죽는다.** 카드가 인용한 프로브 `status="failed" after 5.248208ms` 와 같은 결함이다.
2. 수리는 `context.WithCancel(context.WithoutCancel(ctx))` 이다 — `WithoutCancel` 로 요청의 종료로부터 분리하되, 그 위에 `WithCancel` 을 얹어 `glmLiveJobs` 가 저장하는 취소 핸들을 살려 둔다.
3. **취소 경로는 수리 후에도 동작한다.** 이 카드의 주된 위험(수리가 취소를 깬다)에 대한 가드가 붙었고, 그 가드는 공허하지 않다 — 잘못된 수리를 실제로 잡는다.
4. 결함이 기존 스위트에 걸리지 않은 이유는 **두 가지**이며 둘 다 이 카드에서 닫혔다.

---

## Evidence

### E1 — 결함 위치 (수리 전, `6a46c0edb`)

```
$ grep -n 'WithCancel\|WithoutCancel' internal/cli/glm_task.go
215:	jobCtx, cancel := context.WithCancel(ctx)
241:		cancel() // release the WithCancel context; idempotent once fired
```

`ctx` 는 `handleGLMTask(ctx context.Context, …)` 의 요청 스코프 컨텍스트다.

### E2 — RED: 재현 (수리 전)

```
$ go test ./internal/cli/ -run 'TestGLMTask_BackgroundJobSurvivesRequestContextEnd|TestGLMTask_DetachedBackgroundJobIsStillCancellable' -count=1 -v -timeout 600s
=== RUN   TestGLMTask_BackgroundJobSurvivesRequestContextEnd
    glm_task_bg_context_test.go:108: job ended "failed" after 6.210875ms with error "z.ai request failed: context canceled" — the request context ending killed it
--- FAIL: TestGLMTask_BackgroundJobSurvivesRequestContextEnd (0.01s)
=== RUN   TestGLMTask_DetachedBackgroundJobIsStillCancellable
    glm_task_bg_context_test.go:146: job went terminal ("failed", error "z.ai request failed: context canceled") while its call was still blocked — the request context ending killed it
--- FAIL: TestGLMTask_DetachedBackgroundJobIsStillCancellable (0.01s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	0.885s
```

실측 `6.210875ms` 는 카드가 인용한 프로브 `5.248208ms` 와 같은 자릿수다 — 같은 결함이라는 근거.

### E3 — GREEN: 수리 후

```
$ go test ./internal/cli/ -run 'TestGLMTask_BackgroundJobSurvivesRequestContextEnd|TestGLMTask_DetachedBackgroundJobIsStillCancellable' -count=1 -v -timeout 600s
--- PASS: TestGLMTask_BackgroundJobSurvivesRequestContextEnd (0.01s)
--- PASS: TestGLMTask_DetachedBackgroundJobIsStillCancellable (0.33s)
ok  	github.com/modu-ai/moai-adk/internal/cli	1.296s
```

### E4 — 뮤턴트: 잘못된 수리를 가드가 잡는가

주입한 뮤턴트 = codex 한 줄 수리의 순진한 복사. `go runGLMBackgroundJob(jobCtx, …)` 를 `go runGLMBackgroundJob(context.WithoutCancel(ctx), …)` 로 바꿔, `glmLiveJobs` 는 잡에 더 이상 닿지 않는 cancel 함수를 들고 있게 만든 상태.

```
$ go test ./internal/cli/ -run 'TestGLMTask_BackgroundJobSurvivesRequestContextEnd|TestGLMTask_DetachedBackgroundJobIsStillCancellable|TestGLMJobCancel_CancelsRunningJob' -count=1 -v -timeout 600s
--- FAIL: TestGLMJobCancel_CancelsRunningJob (10.02s)
--- PASS: TestGLMTask_BackgroundJobSurvivesRequestContextEnd (0.01s)
    glm_task_bg_context_test.go:180: the blocked call was never aborted by glm_job_cancel — the job's context no longer reaches it
--- FAIL: TestGLMTask_DetachedBackgroundJobIsStillCancellable (8.33s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	19.246s
```

세 줄이 각각 다른 것을 말한다:

- direction 1 이 **PASS** — 방향 1만으로는 잘못된 수리를 통과시킨다. 가드 두 방향이 왜 필요한지의 직접 증거.
- direction 2 가 **FAIL** — 이 카드가 추가한 가드가 뮤턴트를 잡는다. 공허하지 않다.
- 기존 `TestGLMJobCancel_CancelsRunningJob` 도 **FAIL** — 아래 Gaps 참조.

뮤턴트는 되돌렸다(E5·E6 은 되돌린 트리에서 측정).

### E5 — 전체 패키지 스위트 + 정적 검사

```
$ gofmt -l internal/cli/glm_task.go internal/cli/glm_task_bg_context_test.go
(무출력 — 두 파일 모두 포맷 준수)

$ go vet ./internal/cli/
vet rc=0

$ go test ./internal/cli/ -count=1 -timeout 900s
ok  	github.com/modu-ai/moai-adk/internal/cli	517.048s
```

### E6 — 레이스 검출기 (goroutine·context 를 건드리는 변경이므로)

```
$ go test ./internal/cli/ -run 'TestGLM' -count=1 -race -timeout 900s
ok  	github.com/modu-ai/moai-adk/internal/cli	9.416s
```

### E7 — 변경 범위

```
$ git status --short
 M internal/cli/glm_task.go
?? internal/cli/glm_task_bg_context_test.go

$ git diff --stat
 internal/cli/glm_task.go | 14 +++++++++++---
 1 file changed, 11 insertions(+), 3 deletions(-)
```

프로덕션 변경은 한 줄(`WithCancel(ctx)` → `WithCancel(WithoutCancel(ctx))`)이고 나머지 10줄은 왜 이 모양이어야 하는지를 적은 주석이다.

---

## Baseline-attribution

모든 측정은 이 트리(`.claude/worktrees/t532`), 이 실행에서 나왔다.

| 측정 | 트리 상태 | HEAD |
|---|---|---|
| E1·E2 (RED) | 수리 전, 워킹트리 = develop tip + 신규 테스트 파일만 | `6a46c0edb` |
| E3 (GREEN) | 수리 적용 | `6a46c0edb` + 미커밋 수리 |
| E4 (뮤턴트) | 수리 + 뮤턴트 주입 | `6a46c0edb` + 미커밋 |
| E5·E6 | 뮤턴트 되돌린 최종 상태 | `6a46c0edb` + 미커밋 수리 |

`6a46c0edb` 는 `git merge develop --ff-only` 로 도달한 로컬 develop tip 이며, 리드가 배차문에 적은 값과 일치한다.

---

## Gaps — 관측하지 않은 것

1. **기존 `TestGLMJobCancel_CancelsRunningJob` 도 뮤턴트를 잡는다** (E4). 즉 이 카드가 추가한 direction-2 가드는 취소 회귀의 **유일한** 가드가 아니다. 과대주장하지 않는다. 다만 기존 테스트는 요청 컨텍스트가 살아 있는 모양(`context.Background()`)에서만 취소를 재는 반면, 새 가드는 **요청 컨텍스트가 끝난 분리 상태**에서 잰다 — 수리가 실제로 만드는 그 모양이다. 두 테스트는 겹치되 같지 않다.
2. **실제 z.ai 엔드포인트로는 검증하지 않았다.** 전 구간이 주입된 doer 를 쓴다(네트워크·키 불요는 이 패키지의 기존 규율). 따라서 "실운영 GLM 잡이 이제 완주한다"는 **주장하지 않는다** — 주장하는 것은 요청 컨텍스트 종료가 더 이상 잡을 죽이지 않는다는 것뿐이다.
3. **결함 A(오보고)는 범위 밖이며 손대지 않았다.** GLM 은 `z.ai request failed: context canceled` 라고 정직하게 말하므로(E2 에서 그대로 관측됨) codex 쪽 문구 수리가 필요 없다. 카드가 명시한 범위 한정을 그대로 지켰다.
4. **`internal/cli` 밖 패키지는 재지 않았다.** 변경이 그 패키지에 갇혀 있다. 전 패키지 판정은 CI 몫.
5. **다른 `WithCancel(ctx)` 배경-잡 지점을 전수 스윕하지 않았다.** 카드 범위가 GLM 뿌리 결함 하나로 한정돼 있다. 같은 축의 후속(`exec.Command` 환경 누출 스윕)이 이미 t516 에서 리드에게 올라가 결재 대기 중이다.

---

## Residual-risk

1. **`WithoutCancel` 은 값은 옮기고 취소만 버린다.** 요청 컨텍스트에 실린 값 중 *수명이 요청에 묶인* 것(예: 요청 종료 시 닫히는 리소스 핸들)이 있다면, 잡이 그 사후를 참조하게 된다. 현재 이 경로가 컨텍스트에서 읽는 값은 MCP progress token 뿐이고, 백그라운드 잡은 `callGLMTask(…, nil)` 로 토큰을 **명시적으로 넘기지 않는다**(`glm_task.go:250`) — 그래서 지금은 무해하다. 나중에 이 경로가 컨텍스트 값을 더 읽게 되면 이 전제가 깨진다.
2. **direction-2 가드의 settle 창(300ms)은 시간 기반이다.** 극도로 부하가 걸린 머신에서는 goroutine 이 300ms 안에 시작조차 못 해 이론상 흔들릴 수 있다. 다만 판정은 "종료되지 않았음"이라 느린 머신에서는 오히려 더 쉽게 만족된다 — 위양성(가짜 실패)보다 위음성 쪽으로 기울어 있다.
3. **`ctxAwareGLMDoer` 는 실제 `http.Client` 의 컨텍스트 동작을 최소한으로만 모델한다** (시작 시점 1회 검사). 실제 클라이언트는 전송 도중 취소도 잡는다. 이 카드가 재는 성질에는 충분하지만, 전송 중 취소를 재는 테스트를 나중에 쓴다면 이 doer 로는 부족하다.
4. **잡은 여전히 서버 프로세스와 함께 죽는다.** `runGLMBackgroundJob` 의 기존 주석이 명시한 설계이고 이 수리가 바꾸지 않았다. 요청보다 오래 사는 것이지 프로세스보다 오래 사는 것이 아니다.

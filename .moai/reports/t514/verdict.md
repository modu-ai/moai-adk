# t514 — codex_task 즉시 실패를 "10m0s timeout" 으로 오보고 (GH #1687)

- 카드: t514 (Tier S~M, Class B — plan 생략, 원인 규명은 run 소관)
- 브랜치: `WT-codex-task-failcause`
- 트리: `.claude/worktrees/t514`

---

## Claim

1. 제보된 증상은 **두 개의 별개 결함**이 한 문구로 합쳐진 것이다.
   - **A(오보고)**: `runCodexTaskTurn` 이 파생 컨텍스트의 `Done` 만 보고 자기 바운드를
     무조건 이름 붙였다. 호출자 컨텍스트가 끝나서 중단된 턴도 "10분 바운드를 다 썼다"고
     보고했다.
   - **B(즉시 실패의 뿌리)**: background job 이 **요청 스코프 컨텍스트**를 그대로 물려받았다.
     MCP 호스트는 핸들러가 반환하는 순간 그 컨텍스트를 끝내는데, `background=true` 는 즉시
     반환한다. 그래서 모든 background job 이 생성 직후 밀리초 안에 죽었다.
2. 수리는 문구 교체가 아니라 **원인 구분**이다: 우리 바운드가 발화한 경우와 호출자가 끝낸
   경우가 서로 다른 문구·다른 next step 을 싣는다.
3. 회귀는 **2방향** 모두 있다. 진짜 바운드 만료 → 타임아웃 문구, 호출자 종료 → 취소 문구.
4. 취소 경로(`codex_job_cancel`)는 이 변경에 영향받지 않는다.

## Evidence

### E1. 재현 (수리 **전**, RED) — 제보 증상 그대로

명령:
```
go test ./internal/cli/ -run 'TestCodexTaskTurn_CallerCancelIsNotReportedAsTimeout|TestCodexTaskTurn_RealBoundExpiryStillReportsTimeout|TestCodexTask_BackgroundJobSurvivesRequestContextEnd' -count=1 -v
```

출력(발췌, 축약 없음):
```
=== RUN   TestCodexTaskTurn_CallerCancelIsNotReportedAsTimeout
    codex_task_failcause_test.go:78: turn ended after 50.650958ms (bound was 10m0s)
    codex_task_failcause_test.go:81: a caller-cancelled turn that ended in 50.650958ms is reported as a bound expiry: "codex_task turn timed out after 10m0s (the bound codex_task imposes on its own turns); the turn was abandoned and the session torn down"
    codex_task_failcause_test.go:84: summary must name the actual cause (caller cancellation); got "codex_task turn timed out after 10m0s (the bound codex_task imposes on its own turns); the turn was abandoned and the session torn down"
--- FAIL: TestCodexTaskTurn_CallerCancelIsNotReportedAsTimeout (0.05s)
=== RUN   TestCodexTaskTurn_RealBoundExpiryStillReportsTimeout
--- PASS: TestCodexTaskTurn_RealBoundExpiryStillReportsTimeout (0.12s)
=== RUN   TestCodexTask_BackgroundJobSurvivesRequestContextEnd
    codex_task_failcause_test.go:144: background job ended "failed" with error "codex_task turn timed out after 10m0s (the bound codex_task imposes on its own turns); the turn was abandoned and the session torn down"; the request context ending must not kill it
--- FAIL: TestCodexTask_BackgroundJobSurvivesRequestContextEnd (0.01s)
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.115s
```

**측정된 경과 시간**: 카드가 [HARD] 로 요구한 증거다.
- 호출자 취소 경로: **50.65 ms** — 보고된 바운드 `10m0s` 의 약 1/11,800.
- background job: **0.01 s** 안에 `failed` 로 착지하면서 `10m0s timeout` 문구를 실었다.
  제보의 "~300ms 안에 실패" 와 같은 축이며, 제보 문구와 **완전히 동일한 문자열**이다.

### E2. 수리 후 (GREEN)

```
=== RUN   TestCodexTaskTurn_CallerCancelIsNotReportedAsTimeout
    codex_task_failcause_test.go:78: turn ended after 51.286083ms (bound was 10m0s)
--- PASS: TestCodexTaskTurn_CallerCancelIsNotReportedAsTimeout (0.05s)
=== RUN   TestCodexTaskTurn_RealBoundExpiryStillReportsTimeout
--- PASS: TestCodexTaskTurn_RealBoundExpiryStillReportsTimeout (0.12s)
=== RUN   TestCodexTask_BackgroundJobSurvivesRequestContextEnd
--- PASS: TestCodexTask_BackgroundJobSurvivesRequestContextEnd (0.01s)
PASS
ok  	github.com/modu-ai/moai-adk/internal/cli	1.560s
```

### E3. 뮤턴트 — direction-2 가드가 공허하지 않음

direction-2(진짜 타임아웃)는 수리 **전에도** 초록이었으므로, RED-now 로는 채택 판정할 수
없다. 판별식을 "항상 호출자 종료"로 뒤집는 뮤턴트를 넣고 측정했다:

뮤턴트: `if parentErr := parent.Err(); parentErr != nil {` → `if parentErr := parent.Err(); true { ...; parentErr = context.Canceled`

```
--- FAIL: TestCodexTaskTurn_RealBoundExpiryStillReportsTimeout (0.12s)
    codex_task_failcause_test.go:108: a real bound expiry must still be reported as a timeout; got "codex_task turn was cancelled by the caller, well before the 120ms bound codex_task imposes on its own turns; the turn was abandoned and the session torn down"
FAIL
```

가드가 뮤턴트를 잡았다. 뮤턴트는 되돌렸고 재확인 초록:
```
go test ./internal/cli/ -run 'TestCodexTaskTurn_|TestCodexTask_' -count=1
ok  	github.com/modu-ai/moai-adk/internal/cli	1.853s
```

### E4. 취소 경로가 깨지지 않는 근거 (코드 판독)

`context.WithoutCancel` 은 컨텍스트 취소로 job 을 죽이는 경로를 없앤다. 그 경로가 실제
취소 기능이 쓰는 경로인지 확인했다:

```
grep -n "codexLiveJobSessions\|sendTurnInterrupt\|context\|cancel" internal/cli/codex_job_control.go
209:	stored, live := codexLiveJobSessions.Load(rec.ID)
236:	} else if err := session.sendTurnInterrupt(rec.ThreadID, rec.TurnID); err != nil {
238:			" The cancel fell back to terminating the codex process this server spawned for the job."
```

`handleCodexJobCancel` 의 시그니처는 `(_ context.Context, ...)` — 컨텍스트를 **아예 쓰지
않는다**. 취소는 `turn/interrupt` 전송 + 프로세스 종료로 이뤄지므로 이 변경과 무관하다.

### E5. 건드린 패키지 전체

```
go vet ./internal/cli/...        → exit 0, 무출력
go test ./internal/cli/ -count=1 → ok  github.com/modu-ai/moai-adk/internal/cli  478.125s
```

## Baseline-attribution

- 모든 수치는 이 트리(`.claude/worktrees/t514`, 브랜치 `WT-codex-task-failcause`)에서
  이번 실행으로 측정했다. 다른 패키지·다른 시점에서 옮겨온 값은 없다.
- 재현(E1)은 수리 커밋 **이전** 워킹트리에서, GREEN(E2)은 수리 적용 **직후** 같은 트리에서
  측정했다.
- 전체 패키지(E5)는 뮤턴트를 되돌린 뒤 측정했다.

## Gaps — 관측하지 **않은** 것

1. **실제 codex 바이너리로는 재현하지 않았다.** 재현은 주입된 페이크 세션(정지형 전송)
   위에서 이뤄졌다. 제보자의 ~300ms 와 본 측정의 0.01s 는 자릿수가 다르며, 실환경에서는
   프로세스 spawn·핸드셰이크 비용이 더해질 것이다. 문구·경과시간의 **정성적 일치**는
   보였으나, 제보자 환경의 정확한 ms 를 재현했다고 주장하지 않는다.
2. **제보자 환경에서 실제로 무엇이 요청 컨텍스트를 끝냈는지**는 확인하지 않았다. B 가
   충분조건임은 보였으나, 제보 사례가 B 였다는 것은 **추론**이다 — 다른 즉시 실패 원인이
   같은 문구로 흡수됐을 가능성도 A 의 정의상 남아 있다(그래서 A 를 별도로 고쳤다).
3. **다른 OS/아키텍처**에서는 측정하지 않았다(darwin/arm64 only). 매트릭스 판정은 CI 몫.
4. **`internal/cli` 밖의 패키지**는 돌리지 않았다. 변경 반경이 그 안이므로 CI 가 전수 판정.
5. **codex_audit / review 게이트 경로**는 별도로 측정하지 않았다. 이 경로는
   `runCodexTaskTurn` 을 타지 않고 자체 드라이버를 쓴다.

## Residual-risk — 관측한 것에도 불구하고 틀릴 수 있는 것

1. `context.WithoutCancel` 은 취소를 버리므로, **서버 종료 시** background job 이 컨텍스트로
   중단되던 (있었다면) 경로도 사라진다. 다만 이 job 은 원래도 서버 프로세스 안의
   goroutine 이고 "서버가 죽으면 job 도 죽는다"가 명시된 설계라(파일 헤더 주석), 실질
   변화는 없다고 판단했다 — 그러나 이것은 **판단**이지 측정이 아니다.
2. 취소 문구는 `cancelled` 라는 단어에 회귀 테스트가 걸려 있다. 문구를 나중에 다듬을 때
   그 단어를 빼면 가드가 조용히 약해진다(테스트는 여전히 통과하지 않고 실패하므로 시끄럽게
   깨지긴 한다).
3. 부모 컨텍스트가 **우리 바운드와 거의 동시에** 끝나는 경합에서는 어느 쪽이 먼저인지에
   따라 문구가 갈릴 수 있다. 두 원인이 실제로 동시 발생한 경우이므로 어느 문구도 거짓이
   아니지만, 결정적이지는 않다.

---

## 함께 발견된 것 — 카드 범위 **밖**, 리드에 후속 카드 요청

`glm_task` 가 **같은 뿌리 결함 B** 를 갖고 있다. 추측이 아니라 실측했다(임시 프로브,
측정 후 삭제):

```
handleGLMTask(ctx, background=true) → cancel() 직후
PROBE RESULT: status="failed" after 5.248208ms (bound was 30s) error="z.ai request failed: context canceled"
```

- 위치: `internal/cli/glm_task.go:214` — `jobCtx, cancel := context.WithCancel(ctx)` 가
  요청 컨텍스트에서 파생된다.
- 차이: GLM 은 **오보고하지 않는다**(`context canceled` 라고 정직하게 말한다). 즉 결함 A 는
  codex 전용, 결함 B 는 두 표면 공통이다.
- **이 카드에서 고치지 않았다.** 카드 범위는 `codex_task` 이고, 큐의 생산자는 리드다.
- 주의: GLM 은 `glmLiveJobs` 에 **cancel 함수를 저장해 취소에 쓴다**(codex 와 달리 컨텍스트가
  취소 경로다). 따라서 `WithoutCancel` 을 그대로 적용하면 취소가 깨진다 — 수리는
  `context.WithoutCancel(ctx)` 위에 새 `WithCancel` 을 얹는 모양이어야 한다. codex 의
  일줄 수리를 복사하면 안 된다.

---

## 변경 파일

| 파일 | 성격 |
|---|---|
| `internal/cli/codex_task.go` | 수리 — 원인 구분(A) + background 컨텍스트 분리(B) |
| `internal/cli/codex_task_failcause_test.go` | 회귀 2방향 + 뿌리 결함 가드 (신규) |
| `.moai/reports/t514/verdict.md` | 이 문서 |

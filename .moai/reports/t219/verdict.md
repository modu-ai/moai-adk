# t219 — `TestGLMTask_BackgroundLifecycle` 라이브맵 경합 (Class B)

카드: t219 · 브랜치 `WT-glm-livemap-race` · base `origin/main cd0cee1b8`
발견 경로: PR #1625(t214)의 Race Test 실패 → 귀속 부정 → 별도 카드

## 1. Claim

테스트가 **생산자가 보장하지 않는 순서를 단언**하고 있었습니다. 단언을 폴링으로 바꿨습니다.
생산 코드는 건드리지 않았습니다.

## 2. Evidence

### 2-A. 순서는 정확히 반대로 보장돼 있습니다

```go
// internal/cli/glm_task.go — runGLMBackgroundJob
defer func() {
    glmLiveJobs.Delete(jobID)          // ② 라이브맵 삭제
    cancel()
}()
...
registry.updateUnlessCancelled(jobID, func(r *GLMJobRecord) {
    r.Status = glmJobStatusCompleted   // ① 터미널 레코드 기록
    r.Output = output
})
```

`defer`는 본문이 끝난 뒤에 돕니다. 따라서 **레코드가 `completed`가 되는 시점이 라이브맵 삭제보다
반드시 앞섭니다.** 그런데 테스트는 레지스트리를 폴링하다 `completed`를 **처음 보는 그 순간**
라이브맵이 이미 비었다고 단언했습니다:

```go
// 수정 전 — glm_task_test.go
if rec.Status == glmJobStatusCompleted && rec.Output == "background done" {
    if _, live := glmLiveJobs.Load(jobID); live {
        t.Error("live-map entry must be gone after the job ended")
    }
```

**보장된 적이 없고, 창이 좁아 통과해온 것입니다.**

### 2-B. 재현 — 그리고 부하 의존성 실측

```
사전 수정 트리, 머신 load ≈ 29:
  go test -race -count=200 -run TestGLMTask_BackgroundLifecycle ./internal/cli/
  → --- FAIL: TestGLMTask_BackgroundLifecycle (0.23s)
       glm_task_test.go:404: live-map entry must be gone after the job ended
     FAIL  46.072s

사전 수정 트리, 머신 load ≈ 10 (같은 코드, 같은 명령):
  count=200 → ok 45.977s
  count=600 → ok 17.763s   (600 RUN / 600 PASS 로 확인 — 캐시 아님)
```

**같은 코드가 부하에 따라 갈립니다.** count=600이 count=200보다 빨랐던 것(17.7s vs 46s)이
그 자체로 증거입니다 — 반복 횟수가 아니라 **머신 부하가 반복당 시간을 8배 바꿨고**, 창이
벌어지는 것도 같은 축입니다. 이것이 **로컬에서는 잘 통과하는데 CI 러너에서 터지는** 이유이고,
PR #1625가 정확히 그렇게 실패했습니다.

### 2-C. 수정과 검증

기존 관용구를 그대로 씁니다 — 같은 파일의 `waitForGLMJobToStop`(5초 기한 폴링, 기한 초과 시
`t.Errorf`)이 이미 이 join을 위해 존재했고, 다른 3개 지점이 쓰고 있습니다.

```go
// 수정 후
if rec.Status == glmJobStatusCompleted && rec.Output == "background done" {
    waitForGLMJobToStop(t, jobID)   // 단언은 유지 — 기한 내 안 사라지면 실패
    return
}
```

```
수정 후:
  count=200 → ok  49.894s
  count=500 → ok 112.985s
  count=600, TestGLMTask 계열 전체 (-run TestGLMTask) → ok 153.022s
```

## 3. Baseline-attribution

- 트리: `WT-glm-livemap-race`, base `origin/main cd0cee1b8`. **어느 PR에도 스택하지 않았습니다**(독립).
- 사전 수정 상태 재현은 `git checkout -- internal/cli/glm_task_test.go`로 원본을 되돌린
  실제 트리에서 측정했습니다(수정본은 스크래치에 보관 후 복원).
- 부하는 `uptime`으로 확인: 재현 시점 load 29.10 → 통과 시점 9.90.

## 4. Gaps (미검증)

- **수정 후 통과 횟수는 부재의 근거가 아닙니다.** §2-B가 보여주듯 이 결함은 부하 의존적이라,
  낮은 부하에서의 통과는 사전 수정 코드에서도 나옵니다. 통과 회차를 근거로 삼으면
  **사전 수정 코드도 "고쳐졌다"고 말하게 됩니다.**

  **근거는 통계가 아니라 구조입니다**: 수정 후 단언은 순서에 의존하지 않습니다. 즉시 표본을
  뜨는 대신 5초 기한으로 폴링하므로, 삭제가 레코드 기록 뒤에 오든 앞에 오든 통과합니다.
  실패 양태가 **줄어든 게 아니라 제거**됐습니다.
- **CI 러너에서의 재확인 미완.** 이 트리에서 재현한 부하 조건은 로컬 실측이고, 실제 실패는
  CI에서 났습니다. CI가 최종 판정입니다.
- **`internal/cli` 전체 수트 로컬 미실행** (단독 336초 실측, [HARD] 금지). CI 몫.

## 5. Residual-risk

- **같은 형태가 다른 곳에 남아 있을 수 있습니다.** "터미널 레코드 = 정리 완료"라는 가정은
  codex 잡 계열(`codex_task.go` / `codex_job_control.go`)이 같은 defer 구조를 쓰고 있으면
  그대로 재현됩니다. 이번 카드는 실패한 한 지점만 고쳤고, 계열 전수 점검은 하지 않았습니다.
- **생산자를 고치지 않은 것은 의도입니다.** `glm_task.go`의 주석이 defer의 목적을 명시합니다 —
  "the live-map entry is removed and the cancel context released on EVERY exit path, so a
  terminal record is never left with a live entry the cancel path could still address."
  삭제를 레코드 기록보다 앞으로 옮기면 그 보장(모든 종료 경로 커버)이 깨집니다.
  **테스트의 가정이 틀린 것이지 생산 코드의 순서가 틀린 게 아닙니다.**
- **`waitForGLMJobToStop`이 t.Cleanup에서 한 번 더 불립니다.** 이미 사라진 뒤라 즉시 반환하므로
  무해하지만, 중복 호출이라는 점은 기록해 둡니다.

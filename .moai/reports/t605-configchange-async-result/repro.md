# t605 — 재현 재확인 기록 (착수 전)

- 카드: t605 · [hooks 감사 2026-09-11 · H10 · P2] 단발성 ConfigChange 프로세스가 비동기 검증 결과를 남기지 못함
- 기준 트리: worktree `.claude/worktrees/t605`, 브랜치 `WT-configchange-async-result`, HEAD `eabce74448e094dd1a4393044138a04b2016af99` (= origin/develop)
- 대상: `internal/hook/config_change.go:73-79` (goroutine 기동), `internal/cli/hook.go:267-353` (dispatch 후 즉시 반환)

## Claim

1. H10 은 최신 develop 에서 **재현된다**. CLI 프로세스 5/5 에서 거부 경고가 관측되지 않았다.
2. 다만 **원인이 카드에 적힌 것 하나가 아니라 둘**이다. 어느 하나만 고쳐도 완료 조건("실제 CLI 프로세스가 종료된 뒤 기대한 결과가 남는지")은 성립하지 않는다.
   - C1 — 비동기 goroutine 이 프로세스 종료 전에 검증 단계에 **도달조차 못 한다** (0/5).
   - C2 — `moai hook` 경로는 slog 레코드를 **무조건 폐기**한다(`io.Discard`). 따라서 C1 을 고쳐도 경고는 여전히 아무 데도 남지 않는다.

## Evidence

### E1 — 원본 바이너리 재현 (5/5 거부 경고 없음)

명령: `bin/moai hook config-change < payload.json` (payload 의 `config_file_path` 는 잘못된 YAML 파일)

```
exit=0
--stdout--
{}
--stderr--
(0 bytes)
```

`MOAI_LOG_LEVEL=debug` 를 걸어도 stderr 는 `0` 바이트였다. 동기적으로 먼저 찍히는
`slog.Info("config file changed", ...)` 조차 나오지 않는다 — 이것이 C2 의 직접 증거다.

### E2 — C2 의 코드 근거

`internal/cli/logging.go:49-63` `resolveLoggingDecision`:

```go
if isHookCommand(args) {
    return loggingDecision{dest: io.Discard, level: defaultLogLevel}
}
```

주석이 명시한다 — "That carve-out is unconditional — MOAI_LOG_LEVEL does not re-open it."

### E3 — C1 의 계측 측정 (probe build, 이후 원상복구)

`runReload` 의 검증 직전/거부 분기에 파일 마커 쓰기를 임시 삽입해 `/tmp/t605-repro/moai-probe` 로 빌드.

- CLI 프로세스 5회 실행 → `marker-reached-validate` **0/5** (검증 단계 미도달)
- 대조군: `go test ./internal/hook/ -run TestConfigChange -count=1` (WaitGroup 을 조인하는 in-process 경로)
  → `ok  github.com/modu-ai/moai-adk/internal/hook  1.074s`, 마커 **2/2 생성**
  (`marker-reached-validate` 1 byte, `marker-rejected` 66 bytes)

대조군이 같은 계측으로 마커를 남겼으므로 계측 자체는 멀쩡하며, 갈리는 변수는 "goroutine 을 기다렸는가" 하나다.

측정 후 `internal/hook/config_change.go` 는 원본으로 복구했고 작업 트리 상태는 깨끗하다.

### E4 — 프로덕션 배선 확인

`ConfigChange` 는 실제로 배선돼 있다(죽은 경로가 아니다):
`.claude/settings.json:293` 및 `internal/template/templates/.claude/settings.json.tmpl:293`.

## Baseline-attribution

모든 측정은 이 런에서, 이 트리(HEAD `eabce7444`)에서 수행했다. 감사 보고서의
`probes-async.json` 수치는 인용하지 않았고 재측정으로 대체했다.

## Gaps

- 형제 핸들러 3본(`notification.go`, `task_created.go`, `file_changed.go`)이 **같은 모양**을 갖는다는 것은 코드 읽기로만 확인했고, 각각을 CLI 로 재현하지는 않았다. 카드 범위 밖이라 측정하지 않았다.
- 거부 결과를 **어디에** 남길지(내구 로그 경로)는 아직 정하지 않았다. 선례는 있다:
  `internal/hook/branch_guard.go:47` (`.moai/logs/branch-guard-audit.log`),
  `internal/hook/session_guard.go:15` (`.moai/logs/preedit-session-guard.log`).
- 실제 Claude Code 런타임이 이 훅을 발화시키는 상황은 재현하지 않았다(합성 payload 로만 측정).

## Residual-risk

- C1 은 타이밍 의존이므로 부하가 낮은 머신에서는 간헐적으로 도달할 수 있다. 0/5 는 "항상 실패"가 아니라 "이 트리·이 머신에서 5회 모두 실패"다.
- `registry.Shutdown()` 이 이미 flush barrier 로 쓰이고 있어(`internal/cli/hook.go:328`) 여기에 핸들러 대기를 얹는 것이 자연스러워 보이지만, 이는 아직 **설계 제안**이며 검증된 바 없다.

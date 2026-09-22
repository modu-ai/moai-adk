# t605 — 수리 판정

- 카드: t605 · [hooks 감사 2026-09-11 · H10 · P2] 단발성 ConfigChange 프로세스가 비동기 검증 결과를 남기지 못함
- 트리: worktree `.claude/worktrees/t605`, 브랜치 `WT-configchange-async-result`, 기준 HEAD `eabce74448e094dd1a4393044138a04b2016af99` (= origin/develop)
- 착수 전 재현 기록: `.moai/reports/t605-configchange-async-result/repro.md`
- 범위 판정: 리드 2026-09-12 — (a) C2 를 이 카드에 포함, (b) 형제 핸들러 3본은 별도 카드

> **증거 경로가 배차와 다르다 — 자진 보고.** 배차는 `evidence: .moai/reports/t605/verdict.md` 였으나,
> 그 경로에는 **번호가 충돌하는 옛 t605 카드**(goal 동사 무장, 브랜치 `WT-goal-verb-arm`)의 착지 기록이
> 이미 있었다. 나는 그것을 인지하지 못한 채 `verdict.md` 를 덮어썼고, 커밋 직전 `git status` 가 신규(`??`)가
> 아니라 수정(`M`)으로 잡아낸 덕에 발견해 `git restore` 로 원본을 되돌렸다(추적 파일이라 유실 없음).
> 두 카드가 같은 번호를 쓰는 이상 한 디렉터리를 공유할 수 없어, 이 카드의 증거는
> `.moai/reports/t605-configchange-async-result/` 로 옮겼다. 카드 id 는 경로에 남아 traceability 를 유지한다.
> 배차가 지정한 경로에서 벗어난 것이므로 리드의 판정을 받는다.

## Claim

1. C1 — 단발성 훅 프로세스가 종료 전에 핸들러의 비동기 작업을 **제한된 범위 안에서 기다린다**.
2. C2 — 그 작업의 결과가 프로세스 종료 **후에도 남는다**. 거부군·정상군 양쪽 모두.
3. 카드의 완료 조건("핸들러 단위 WaitGroup 시험뿐 아니라 실제 CLI 프로세스가 종료된 뒤 기대한 결과가 남는지")을 실제 CLI 프로세스 측정으로 충족한다.

## Evidence

### E1 — 실제 CLI 프로세스, 실패 재현군 (카드 완료 조건의 절반)

각 실행은 **별도 프로세스**다. 바이너리: `bin/moai` (`make build`, exit=0).

```
bin/moai hook config-change < /tmp/t605-verify/payload-bad.json
→ stdout {}, exit=0   (5회 각각)
```

종료 후 `/tmp/t605-verify/.moai/logs/config-change-audit.log`:

```
[2026-09-12T08:27:11Z] session=fail-grp path=/tmp/t605-verify/bad.yaml source=project result=rejected detail="invalid YAML: yaml: line 1: did not find expected ',' or ']'"
[2026-09-12T08:27:13Z] session=fail-grp path=/tmp/t605-verify/bad.yaml source=project result=rejected detail="invalid YAML: yaml: line 1: did not find expected ',' or ']'"
[2026-09-12T08:27:16Z] session=fail-grp path=/tmp/t605-verify/bad.yaml source=project result=rejected detail="invalid YAML: yaml: line 1: did not find expected ',' or ']'"
[2026-09-12T08:27:19Z] session=fail-grp path=/tmp/t605-verify/bad.yaml source=project result=rejected detail="invalid YAML: yaml: line 1: did not find expected ',' or ']'"
[2026-09-12T08:27:21Z] session=fail-grp path=/tmp/t605-verify/bad.yaml source=project result=rejected detail="invalid YAML: yaml: line 1: did not find expected ',' or ']'"
```

`grep -c 'result=rejected'` → **5**. 수리 전 같은 축의 측정은 **0/5** 였다(repro.md E3).

### E2 — 실제 CLI 프로세스, 정상 대조군 (나머지 절반)

```
bin/moai hook config-change < /tmp/t605-verify/payload-good.json
→ stdout {}, exit=0   (5회 각각)
```

```
[2026-09-12T08:27:31Z] session=ctrl-grp path=/tmp/t605-verify/good.yaml source=project result=reloaded detail="fallback"
[2026-09-12T08:27:34Z] session=ctrl-grp path=/tmp/t605-verify/good.yaml source=project result=reloaded detail="fallback"
[2026-09-12T08:27:37Z] session=ctrl-grp path=/tmp/t605-verify/good.yaml source=project result=reloaded detail="fallback"
[2026-09-12T08:27:40Z] session=ctrl-grp path=/tmp/t605-verify/good.yaml source=project result=reloaded detail="fallback"
[2026-09-12T08:27:43Z] session=ctrl-grp path=/tmp/t605-verify/good.yaml source=project result=reloaded detail="fallback"
```

`grep -c 'result=reloaded'` → **5**. 파일 전체 **10** 줄.

대조군을 기록하는 이유는 장식이 아니다. 거부만 기록하면 "설정이 멀쩡했다"와 "검사가 아예 안 돌았다"가 같은 침묵으로 나타나며, 그 모호함이 이 카드가 없애려던 바로 그 결함이다.

### E3 — 단위 테스트

```
go test ./internal/hook/ -run TestRegistryShutdownJoinsConfigChangeAsyncWork -v
--- PASS: TestRegistryShutdownJoinsConfigChangeAsyncWork (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/hook	1.031s

go test ./internal/hook/ -run TestConfigChangeAuditRecordsBothOutcomes -v
--- PASS: TestConfigChangeAuditRecordsBothOutcomes (0.00s)
    --- PASS: TestConfigChangeAuditRecordsBothOutcomes/valid_config_records_reloaded (0.02s)
    --- PASS: TestConfigChangeAuditRecordsBothOutcomes/invalid_config_records_rejected (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/hook	0.569s

go test ./internal/hook/ -run TestConfigChangeJoinAsync -v
--- PASS: TestConfigChangeJoinAsyncReturnsOnCompletion (0.00s)
--- PASS: TestConfigChangeJoinAsyncBoundsTheWait (0.02s)
ok  	github.com/modu-ai/moai-adk/internal/hook	0.602s

go test ./internal/hook/ -run TestRegistryShutdownTolerantOfPlainHandlers -v
--- PASS: TestRegistryShutdownTolerantOfPlainHandlers (0.00s)
ok  	github.com/modu-ai/moai-adk/internal/hook	0.650s

go test ./internal/config/ -count=1
ok  	github.com/modu-ai/moai-adk/internal/config	9.108s
```

**테스트 설계 판단 (리드 지시로 명시 기록).** `TestRegistryShutdownJoinsConfigChangeAsyncWork` 는 `testutil.WaitForAsync` 를 **일부러 부르지 않는다.** 테스트 쪽에서 핸들러의 WaitGroup 을 조인하면 이미 수리 전에도 통과하던 대조 조건을 재현하는 셈이고, 정작 실패하던 프로덕션 조건(조인해 줄 주체가 없는 단발성 프로세스)은 건드리지 못한다. 이 테스트는 `Dispatch` → `Shutdown` 만으로 결과가 디스크에 남는지를 본다 — CLI 가 실제로 하는 일과 같은 순서다. 양성 대조로 `TestConfigChangeJoinAsyncReturnsOnCompletion` 을 두어, 항상 타임아웃을 보고하는 `joinAsync` 가 통과해 버리는 경우를 배제했다.

### E4 — lint

```
golangci-lint run ./internal/hook/... ./internal/config/...
0 issues.
exit=0
```

### E5 — 선행 결함 귀속 (제 변경 소관 아님)

`go test ./internal/hook/` 전체에서 `TestSessionStart_DeferredScanDoesNotBlockReturn` 1건이 실패한다(500ms 예산 대비 605ms). **제 변경 탓이 아님을 기계적으로 갈랐다**: 수정 3본을 `git restore` 로 기준 상태로 되돌리고 새 테스트 파일을 치운 뒤 같은 테스트를 돌리면 **기준에서도 726ms 로 동일하게 실패**한다.

```
(수정 적용)  session_start_parallel_test.go:97: Handle blocked 605.787417ms ... FAIL
(기준 복원)  session_start_parallel_test.go:97: Handle blocked 726.423833ms ... FAIL
```

측정 시점 머신 부하 15~35 구간의 latency 단정으로, 별도 소관이다. 측정 후 파일은 전부 복구했다.

## Baseline-attribution

모든 측정은 이 런에서, 이 트리에서 수행했다. 비교 기준인 "수리 전 0/5" 는 같은 트리·같은 머신에서 착수 전에 직접 잰 값이며(repro.md E3), 감사 보고서의 `probes-async.json` 수치를 인용하지 않았다. E5 의 기준 측정은 `git restore` 로 만든 실제 기준 상태에서 수행했다.

## 변경 내역

| 파일 | 내용 |
|---|---|
| `internal/config/defaults.go` | `DefaultHookAsyncJoinTimeout = 2 * time.Second` 추가. §14 임계값 SSOT |
| `internal/hook/config_change.go` | `joinAsync(timeout)` 추가(C1). `.moai/logs/config-change-audit.log` 에 거부·정상 양쪽 기록(C2) |
| `internal/hook/registry.go` | `Shutdown()` 이 등록 핸들러 중 `asyncJoiner` 를 먼저 조인한 뒤 trace writer flush. `asyncJoiner` 비공개 인터페이스 추가 |
| `internal/hook/config_change_teardown_test.go` | 신규, 테스트 5본 |

**새 호출 지점을 만들지 않았다.** CLI 가 이미 `defer` 하던 `registry.Shutdown()`(`internal/cli/hook.go:328`)에 두 번째 비동기 축을 얹었다. 조인이 flush 보다 먼저 오는 이유는 핸들러가 나가는 길에 내보낸 것이 아래 drain 앞에 놓이게 하기 위해서다.

**설계 판단 — 중복 제거 자료구조를 스스로 걷어냄 (리드 지시로 명시 기록).** 조인 대상을 고르는 루프에 `map[Handler]struct{}` 중복 제거를 넣었다가 제거했다. `Register` 가 `handler.EventType()` 으로 버킷을 정하므로 한 인스턴스는 한 버킷에만 들어가 중복이 **구조적으로 불가능**하고, 남는 것은 비교 불가 동적 타입이 들어올 때의 런타임 패닉 위험뿐이다. 못 버는 복잡도라 걷어냈다.

## Gaps

- 형제 핸들러 3본(`notification.go`, `task_created.go`, `file_changed.go`)은 **건드리지 않았다**. 리드 판정대로 별도 카드 소관이며, 이 카드에서 CLI 재현을 하지 않았다. 다만 `asyncJoiner` 인터페이스는 이미 일반적이라, 형제 카드는 각 핸들러에 `joinAsync` 를 붙이는 것만으로 이 조인 경로에 올라탈 수 있다.
- 실제 Claude Code 런타임이 `ConfigChange` 를 발화시키는 상황은 재현하지 않았다. 합성 payload 로만 측정했다. 배선 자체는 `.claude/settings.json:293` 에 살아 있음을 확인했다.
- `result=reloaded` 의 `detail=fallback` 은 `mgr == nil` 인 폴백 경로다. `mgr` 가 주입된 RT-005 경로(`detail=rt005-manager`)는 단위 테스트로도 CLI 로도 밟지 않았다 — 현재 `NewConfigChangeHandler()` 가 `mgr` 를 주입하지 않아 프로덕션에서 도달하지 않는 가지다.
- `-race` 를 돌리지 않았다. 부하 지침(직렬 1건씩)과 전체 수트 1회가 320초인 점을 감안해 CI 판정에 맡긴다.

## Residual-risk

- 조인 예산 2초는 **측정이 아니라 선택**이다. flush 예산과 같은 값을 택했고(둘 다 "단발성 훅 프로세스가 종료 시 얼마나 머물러도 되는가"를 재는 같은 축), 실제 작업량(20ms 디바운스 + 파일 읽기 + YAML 파싱)은 그보다 몇 자릿수 아래다. 예산을 넘기면 경고만 남기고 나가므로 세션이 멈추지는 않는다.
- 훅 프로세스의 종료가 이제 비동기 작업만큼 늦어진다. 측정된 정상 경로는 디바운스 20ms + I/O 로, 훅 래퍼 타임아웃(수십 초) 대비 무시할 수준이지만 0 은 아니다.
- 감사 로그는 append-only 이며 회전 장치가 없다. `.moai/logs/` 의 형제 로그들과 같은 성질이고 `prune_logs.go` 가 `trace-*.jsonl` 를 정리하는 것과는 별개 축이다. 설정 파일이 비정상적으로 자주 바뀌는 환경에서는 자라며, 현재 설계상 그렇다.
- E5 의 선행 실패는 이 카드가 닫지 않는다. 통합 전 develop 병합 트리에서도 같은 부하 민감성이 남아 있을 수 있다.

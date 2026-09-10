# t435 latency verdict — 동기 전환의 세션 시작 지연 실측

운영자 판정(동기 전환, 조건부)의 [HARD] 선행 실측: **구현 전 측정, 측정 커밋과 수리 커밋 분리,
전후 같은 트리·같은 조건, 반복 측정**. 본 커밋은 측정 증거만 담는다 — 수리 코드는 커밋에 없다
(브랜치 `ae36fcc40`의 실측 보고 위에 측정 증거만 추가).

## Claim

1. 동기 전환 시 edges refresh 가 세션 시작에 그대로 더해진다 — 그 단독 소요는 **중앙 6.01s
   (n=10, 5.47–6.72s)** 이고 query-time update budget 기본값(2000ms, `internal/config/defaults.go:29`)
   의 **약 3배**다.
2. stale 세션 시작( edges 가 stale 인 상태 — refresh 가 발동하는 유일한 상태)의 프로세스 수명은
   현행(비동기) 중앙 **0.23s** → 동기 시뮬 중앙 **7.11s** (+6.9s).
3. fresh 세션(edges 산출물 신선 — refresh 스킵)의 동기 전환 비용은 0 에 수렴해야 하지만 본
   측정의 sync fresh 중앙 1.36s 는 **edges 가 아니라 측정 스위치의 부수**(아래 Gaps)다 —
   edges-only 전환의 fresh 비용은 0 이다.

## 측정 설계

- 트리: `/tmp/t435-bench` — 본 워크트리(HEAD `ae36fcc40`) 소스 전체 복사(184MB, bin/.git/.moai
  산출물 제외), edges 산출물 없는 상태에서 시작 — 실제 규모 소스로 refresh 가 실작업을 한다.
- 바이너리: 본 워크트리에서 벤치 전용 패치 2건을 얹어 빌드한 `bin/moai-bench` — (a)
  `MOAI_BENCH_SYNC=1` env 로 deferred 경로를 동기화(세션 시작이 deferred 완료를 기다림), (b)
  edges refresh 소요를 stderr 에 항상 출력. **두 패치는 커밋하지 않았고 측정 후 완전 원복**
  (`git diff` 무출력으로 검증) — diff 내용은 아래 Bench-patch 절.
- 교차(interleave): 10 회차 × {async, sync} × {stale, fresh} — 모드 순서 편향 제거. stale 회차는
  매번 edges.jsonl+meta 삭제로 refresh 발동 조건 유지, fresh 회차는 직전 stale 회차가 만든
  산출물 유지.
- 반복: 각 조합 n=10, `/usr/bin/time -p real` 수집 + refresh duration probe.

## 결과 (`latency-results.csv` 전체 수록)

| 조합 | n | min | median | max |
|---|---|---|---|---|
| async(현행) stale | 10 | 0.19s | **0.23s** | 2.18s(첫 회차 워밍업 1건) |
| async fresh | 10 | 0.20s | 0.22s | 0.42s |
| sync(시뮬) stale | 10 | 6.51s | **7.11s** | 8.64s |
| sync fresh | 10 | 1.19s | 1.36s | 1.60s |
| **edges refresh 단독** | 10 | 5.47s | **6.01s** | 6.72s |

## 판정 재료 (수치만 — 수용 여부는 운영자)

- stale 세션 시작이 **+6.9s** (0.23s → 7.11s, 중앙). refresh 단독으로 봐도 **+6.0s**.
- 시스템이 스스로 정한 update budget(2000ms)의 3배 — "예산 내" 기준으로도 3배 초과.
- fresh 세션의 비용은 0 — 지연은 refresh 가 발동하는 세션에만 발생.
- 착지율 0/60(선행 실측)과 합치면: 지금은 "착지 0"이고 동기 전환은 "항상 착지, 대가 +6s".

## Gaps

- **sync fresh 중앙 1.36s 의 해석 주의**: 본 측정의 스위치는 edges 만이 아니라 **advisory
  스캔 전체**를 동기화했다(test seam 경로). 그 1.36s − async 0.22s ≈ 1.1s 는 advisory 동기화
  비용이지 edges-only 전환의 비용이 아니다. 운영자 제안은 "edges 갱신을 지연 goroutine 에서
  빼 동기 경로로"이므로 fresh 비용은 0 이 맞다 — 본 수치를 edges-only 로 읽지 말 것.
- 프로세스 수명은 직접 실행 기준 — Claude Code 의 실제 훅 수거 타이밍은 미측정(선행 실측과
  같은 한계).
- 단일 머신(darwin)·백그라운드 부하 통제 안 된 상태의 분산 — 다만 교차 배치라 모드 간 비교는
  유효.

## Bench-patch (커밋 안 함 — 기록용)

1. `internal/hook/session_start.go` `deferredScansAsyncEnabled()`: env `MOAI_BENCH_SYNC=1`
   → `return false` (동기 경로 시뮬레이션 — 기존 test seam 경로와 동일 로직).
2. `internal/cli/graph_refresh_cli.go` `deferredEdgesRefresh()`: 성공 시 stderr 에
   `[bench] edges refresh took <duration>` 항상 출력.

## Residual-risk

- refresh 소요가 트리 규모·디스크 캐시에 비례 — 본 수치는 이 리포(184MB 소스) 기준이며 더
  큰 사용자 리포에서는 더 길어진다 (BuildWithCodeLayers 가 소스 전체를 훑음).
- stale 조건의 빈도는 edges 변화율이 결정 — 실사용에서 stale 세션의 비중은 미측정.

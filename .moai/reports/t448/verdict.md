# t448 verdict — 지연 goroutine 부수효과 5축 판정 (run-phase 초안)

카드: t448 (Class B — plan 생략, run → sync). 트리: `WT-deferred-side-effects`
@ develop `5a8449859` (ff 기반; 측정 시점 기준 — 흡수 시 재판독). 측정 근거:
`.moai/reports/t448/probe-results.md` (`1b78bed0e` + 정정 `0d95497e7`).

**상태: 초안** — edges 처방(1안)은 리드 방향 확정(2026-09-03), 구현 시퀀싱((a) 지금 /
(b) t216 sync 착지 후)은 리드가 t216 착지를 보고 결정. 코드 미접촉 상태.

## Claim

1. **edges refresh (session_start 지연 goroutine)**: 유실 실측 0/60(t435, 현장 L1
   워크트리, 산출물 파일 기준, `2660bcd09`) + 실규모 지연 비용 6.01s(t435 벤치, 184MB,
   budget 3배). 두 실측이 합쳐 "지연 dispatch"는 **유실되거나(현장) 비용이 과대(동기화
   시)** — 어느 쪽도 수용 불가. 처방은 **1안 (M4 패턴 이전)**: 훅 goroutine 에서
   edges refresh 를 제거하고, edges 를 필요로 하는 소비자가 동기 지불.
2. **file_changed**: die-at-exit 유실 **메커니즘 재현** — 직접 CLI 발화 0/10 + 인라인
   대조군 통과. 단 발화는 `moai hook file-changed` **직접 호출**(matcher 우회)이므로
   **배포판 필드 노출 여부는 미결**(CC 런타임의 matcher 해석 — 정규식이면
   `envkeys.go` 형 파일 매칭 가능, t454 런타임 측정 축). 코드 변경 없음 — 판정은
   t454 측정 후.
3. **config_change**: 유실 개념 **성립 안 함** — 산출물 비내구(reload 는 in-memory,
   slog 는 훅 경로 io.Discard). 완료해도 남는 것이 없는 축. 비동기 유지의 정당성 자체가
   판정 대상(리드·운영자 큐).
4. **notification / task_created**: 유실 재현 **안 됨** — io.Discard 싱크 관측(게이트
   전개방 + 올바른 cwd, stderr 0B, debug 포함). 유일한 내구 산출물인 레지스트리 트레이스는
   goroutine 밖 **동기** 경로로 착지(146B 관측). "해당 없음" 기록. 주석 정정 후보는
   task_created.go:8-9 의 "JSONL append"(실현되지 않는 조건의 서술 — 훅 경로에 JSONL
   slog 핸들러 설치 코드 부재) 한정.
5. 배차문 `file_changed 0/5` 는 리드 생성 미귀속 수치로 폐기(리드 시인). 본 카드가 그
   축의 첫 실측.

## Evidence

`probe-results.md` E1-E7 참조 (본 문서의 축별 숫자는 전부 거기서 인용):
file_changed 0/10 + 테스트 대조군 PASS, config_change 2런 파일 0개, notification/
task_created 게이트 전개방 stderr 0B + 트레이스 146B 착지, matcher 정규식 재확인 grep.

## Baseline-attribution

측정: `WT-deferred-side-effects` @ `5a8449859`, `bin/moai` commit 스탬프 동일(darwin).
t435 근거(0/60, 6.01s): t435 워크트리 `ae36fcc40`·`4cbf8ce1b`, 본 카드 미재측정.
t454 근거(matcher 정규식): lane-5 실측 + 리드 재측정 + 본인 sanity grep — 인용으로 귀속.

## 1안 구현 범위 추정 (리드 착수 확인용 — 코드 미변경)

**철거 대상 (훅 쪽 + seam)**:

| 위치 | 내용 |
|---|---|
| `internal/hook/session_start.go:48-73` | `DeferredEdgesRefresh` 타입·필드·`WithDeferredEdgesRefresh` 옵션 |
| `internal/hook/session_start.go:~345-356` | `edgesStale` 스냅샷 계산 (Handle 내) |
| `internal/hook/session_start.go:~405` | 인라인(테스트) 경로 dispatch |
| `internal/hook/session_start.go:~739-757` | `spawnDeferredAdvisoryScans` 의 `edgesStale` 파라미터 + goroutine 내 dispatch |
| `internal/hook/session_start.go:762-780` | `runDeferredEdgesRefresh` + @MX:NOTE |
| `internal/cli/deps.go:224-226` | `WithDeferredEdgesRefresh(deferredEdgesRefresh)` 배선 |
| `internal/cli/graph_refresh_cli.go:92-111` | `deferredEdgesRefresh` wrapper (유일 호출자가 위 배선) |
| 관련 주석 | edgesStale·tear-down 서술 주석들(t216 정정본 포함) — dispatch 소멸로 무의미화 |

**보존 (소비자 쪽 — 이미 존재, 신규 공사 불필요)**:

| 위치 | 역할 |
|---|---|
| `internal/cli/graph.go:158` | 쿼리 경로가 이미 `edgesRefreshNeeded` → `refreshEdgesArtifact` 호출 — 소비자 지불 **live** |
| `internal/cli/graph_refresh_cli.go:37-90` | `edgesRefreshNeeded`·`refreshEdgesArtifact`·budget/overrun helpers |
| `internal/graph/codequery.go:261` | 아티팩트 부재 시 "run 'moai graph build'" 안내 |
| `moai graph build` 동사 | 소비자가 명시적으로 지불하는 경로 |

**테스트**: session_start 의 edgesStale 경로 테스트 + graph_refresh_cli 의
`deferredEdgesRefresh` 테스트 + deps 테스트 제거·수정. t216(같은 블록, `mxScanNeeded`
제거)과 병합 순서에 따라 재해소 — lane-1 의 verbatim 최종형 보유.

**후속 후보 (본 카드 범위 밖, 별도 판정)**: `graph.EdgesSourcesMoved`(기본 아티팩트
변형)가 훅 스냅샷 제거 후 참조 0 이 되는지 — 되면 별도 정리. SPEC-GRAPH-REPORT-001
REQ-GR-010(지연 refresh) 메커니즘 철회의 SPEC 본문·HISTORY 정리는 sync 단계 소관.

## Gaps

- edges 축 재측정 없음(t435 근거 인용 — 본 카드 측정 아님).
- file_changed 배포판 노출 여부 미결(t454 런타임 측정 대기 — 본 카드 손대지 않음,
  리드 지정).
- 구현 범위 추정은 코드판독 기반 — 착수 시점에 리드가 재확인하기로 함.
- (a)/(b) 시퀀싱 미결 — t216 sync 착지 후 리드 판정.

## Residual-risk

- 1안 착수 시 `session_start.go` 는 t216 과 동일 블록 — lane-1 이 보존해둔 edges
  refresh 블록이 제거 대상이 되는 것이 lane-1 판단을 뒤집는 변경이나, 0/60 + 6.01s
  실측이 그 명시적 근거(리드 방향 확정).
- config_change 비동기 유지 문제와 notification/task_created 주석은 본 카드 미처리 —
  리드·운영자 큐에 남음.
- 훅 dispatch 제거 후 신규 워크트리에서 edges 아티팩트는 최초 `moai graph` 사용 때까지
  부재 — M4 의 MX 와 같은 수용된 트레이드오프(소비자 지불).

# t435 verdict — SessionStart 지연 goroutine 의 edges refresh 착지율 실측

카드: SessionStart 지연 goroutine 의 내구 부수효과 3번째 인스턴스 — **[HARD] 착지율 실측이 첫
작업, 결함 단정 금지**. 트리: `WT-deferred-edges-rate` @ develop `2660bcd09` (t435 작업 시작
시점 기준 — 창 절차 후 develop 은 e85c55fa9 로 전진, 본 카드 브랜치는 그 이전), darwin.

## 판정 요약

M4 논거("SessionStart 지연 goroutine 은 CLI 프로세스가 Handle 반환과 함께 죽으므로 거기서
dispatch 한 내구 부수효과는 착지를 신뢰할 수 없다")의 edges refresh 적용을 **실측으로
확인**했다. 현장 착지 0/60, 재현 실험에서는 착지-성공(probe2)과 착지-실패(probe4)가 같은
조건에서 갈렸다 — **배선·시임은 정상이고, 변수는 프로세스 수명과 goroutine 순서의 경합**.
결함 여부의 판정(의도/수리 방향)은 이 실측을 받은 리드·운영자 소관으로 남긴다.

## Claim

1. 현장: 이 머신 L1 워크트리 60개 중 edges refresh 산출물(edges.jsonl + edges.meta.json)은
   **0곳** — 한 번도 착지한 적 없다.
2. 대조: 같은 방식의 MX 인덱스 실측은 60개 중 4곳 — M4 의 MX 실측("153개 중 2곳")과 같은
   양상, 측정 방법의 유효성 교차 확인.
3. 실험: 빈 프로젝트에서 hook session-start 실행 시 배선(deferredEdgesRefresh 시임)·시임은
   정상 — 3부수효과(advisory·MX·edges)가 모두 착지한 실행(probe2)이 있다. 그러나 동일
   조건의 다른 실행(probe4)에서는 **MX 는 착지하고 edges 는 유실**됐다.
4. 메커니즘: goroutine 내 순서가 advisory → MX cold-start → edges refresh(session_start.go:751-757)
   이고 join bound 는 250ms(session_start.go:1584) — **edges 는 프로세스 수명 경합에서
   순서상 가장 뒤에 서서 착지율이 MX 보다도 낮은 구조**다.

## Evidence

| # | 명령 | 관측 출력 |
|---|------|----------|
| E1 | `find …/.claude/worktrees -path "*/.moai/project/graph/edges.jsonl"` (+meta) | 워크트리 60개(L1) + 7개(L2) 중 **0개**; primary `.moai/project/graph/` 디렉터리 부재 |
| E2 | `find …/.claude/worktrees -path "*/.moai/state/mx-index.json"` | **4/60** — 대조군(M4 양상 재현); primary mx-index.json 456KB 존재 |
| E3 | 시임 배선 `internal/cli/deps.go:226` | `hook.WithDeferredEdgesRefresh(deferredEdgesRefresh)` — nil 갈래 배제 |
| E4 (probe2) | `bin/moai hook session-start < input`(빈 프로젝트 /tmp/t435-probe2) | rc=0; 종료 후 `mx-index.json` 347B + `edges.jsonl` 0B + `edges.meta.json` 462B — **3부수효과 모두 착지** |
| E5 (probe4) | 동일 재실행(/tmp/t435-probe4, /usr/bin/time) | 프로세스 수명 **real 0.28s**(join bound 250ms 경계); `mx-index.json` 착지, **edges.jsonl·meta 부재** — advisory 키 0건(join 실패); MX 착지/edges 유실 갈림 |
| E6 | goroutine 순서 판독 `internal/hook/session_start.go:749-757` | `resultCh <- advisory` → `if mxScanNeeded { runMXColdStartScan }` → `if edgesStale { h.runDeferredEdgesRefresh }` |

## Baseline-attribution

E1-E3: 본 워크트리(HEAD `2660bcd09`, t435 커밋 전)에서 실측. E4-E5: 본 워크트리에서 빌드한
`bin/moai`(LDFLAGS commit `2660bcd09`)로 /tmp/t435-probe{2,4} 프로젝트에서 실행 — 측정
도구는 측정 트리에서 빌드(도구 귀속). 두 실행의 차이는 코드가 아니라 타이밍 경합 — 같은
입력, 같은 바이너리.

## Gaps

- **실제 Claude Code 훅 환경에서의 프로세스 수명 분포는 미측정** — Claude Code 가 훅
  프로세스 stdout 을 수거한 뒤 얼마나 빨리 수거하는지가 착지율의 나머지 변수인데, probe 의
  0.28s 는 직접 실행 수명일 뿐이다. 현장 0/60 이 그 간접 근거.
- edgesStale 이 true 였던 세션의 과거 재구성은 불가 — 사이드카 부재 ⇒ moved=true(graph/meta.go:163-168)
  이므로 산출물이 없는 모든 워크트리에서 세션 시작마다 edgesStale=true 였다는 추론만 가능
  (probe 조건과 일치).
- join bound 초과 여부의 로그 증거 — slog.Debug 는 기본 레벨에서 미출력, stderr 0바이트.
- t216 M4 의 원실측 기록(153개 워크트리)은 리포트에서 미발견 — 배차문 인용을 근원으로 사용.

## Residual-risk

- **주석 379-380행의 낙관 서술**: "The goroutine continues to completion in the background
  (durable side effects still land)" — probe4 가 이 보장의 부재를 실측했다. 그러나 이 주석이
  lane-1 정정 3곳에 포함되는지 본 카드는 확인하지 않았고, 고치는 것은 결함 단정이 아니라
  사실 정정의 영역이라 리드 판정을 받아 처리하는 게 안전하다. 66행의 "fire-and-forget at
  process lifetime"(@MX:NOTE)은 lane-1 정정 서술로 보존 확인.
- 본 실측은 darwin 단일 머신·단일 실행 반복 기반 — CI windows/macos 의 수명 특성은 미측정.
- **t263 과의 겹침 판정**: t263 은 `internal/hook/file_changed.go:110-118` 의 die-at-exit
  (증분 인덱스 경로) — 본 카드의 표면(session_start.go 지연 goroutine)과 **다른 경로**로
  겹치지 않는다. 같은 계열(프로세스 수명 vs 내구 부수효과)의 별개 인스턴스로 보고.

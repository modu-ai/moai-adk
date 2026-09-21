# t483 verdict — t448 잔여 항목 전수 재측정

카드: t483 (Class B 계열 — 확인 선행, 확인 결과에 따라 run). 트리: `WT-t448-residue`
@ develop `a825183dd` (로컬 develop tip ff 흡수 후 측정). 배차: lead-1, 2026-09-04.

## Claim

1. **「t448 잔여 5건」의 출처는 리드 자체 열거다.** 리드 기록
   `card-status-20260902.md:8074` — "t448 이 남긴 후속 5건 미발행(**edges 1안 구현** ·
   **(a)/(b) 시퀀싱** · **config_change 비동기 정당성** · **notification/task_created
   주석 정정** · **file_changed 배포판 노출**)". 개수 5는 이 열거와 일치한다.
   단 t448 판정서 §Gaps 자체는 **4건**이다 — 5건은 Gaps 4건에 판정 본문의 처방
   (task_created 주석)을 합친 운영 큐 목록이다. 배차문의 「본문에 항목 목록이 없다」는
   카드 본문 기준으로는 참이지만, 리드 기록에는 목록이 있었다.
2. **5건 전수 재측정 결과 — 해소 3 · 운영자 상신 1 · 실작업 1:**

| # | 리드 기록 항목 | 재측정 결과 (tip `a825183dd`) | 처분 |
|---|---|---|---|
| ② | (a)/(b) 시퀀싱 | **사건으로 해소** — t216 착지 (`bac2cf15b`·`52d3f7a49` Merge WT-hook-wiring-drift), `mxScan` 참조 tip에서 0, 대상 파일 충돌 카드 0 | 닫음 — (b) 전제 충족, (a) 지금 |
| ④ | notification/task_created 주석 정정 | **해소 + 오기** — `file_changed.go` 주석은 t454 가 정정·착지 (`f759964b3`); 형제 3종(task_created·config_change·notification)도 t454 형제 스윕으로 정정; 「task_created.go:8-9 JSONL append」는 오기 — 현재 8-9행은 듀얼게이트 조건부 디스패치 설명, 102-104행은 "JSONL append 는 slog 핸들러 소관"이라는 올바른 귀속 | 닫음 — 리드·메모리 기록의 해당 항목 정정 요청 |
| ⑤ | file_changed 배포판 노출 | **측정은 해소** — t454 addendum D1 (4-slot 런타임 프로브): FileChanged 발화 도메인은 `.env`/`.envrc`/`.gitignore` 한정·matcher 무관, 소스 확장자 파일 불발화 → lane-8 의 0-노출 결론 유지(사유 갱신). matcher 확대는 작동 불능 판명 | 닫음 — 단, 좁아진 운영자 결정 1건 상신: **retire vs keep-dormant** |
| ③ | config_change 비동기 정당성 | **미해결** — t454 ⑵ 가 전제를 강화 확인 (프로덕션 생성자 `NewConfigChangeHandler()` 가 mgr 미배선 — tip `config_change.go:37-39` 재확인, reload 는 무내구, slog 는 `io.Discard`). 제거는 SPEC-V3R6-HOOK-ASYNC-EXPAND-001 AC-HAE-003 의 계약 변경 | **운영자 상신** (t454 D2 — 레인이 열 수 없는 게이트) |
| ① | edges 1안 구현 | **미구현** — tip 에 `DeferredEdgesRefresh` 전 경로 존재 (session_start.go:51-75 · 340-354 · 392-393 · 701-757, deps.go:226, graph_refresh_cli.go:92-111) | **본 카드 run 으로 집행** |

3. **① 착수 조건 전부 충족**: 리드 방향 확정(2026-09-03, 1안) · (b) 전제(t216 착지)
   충족 · 미병합 11개 브랜치(`git for-each-ref --no-merged=develop` 전수) 대상 4파일
   무접촉 · t448 verdict 철거 테이블을 tip 에서 행번호 재확인 · 소비자-지불 경로 live
   (`graph.go:158-159` → `refreshEdgesArtifact`, 단일 재빌드 경로).
4. **t448 판정서 §Gaps 4건도 전부 소멸**: G1(edges 재측정) = ① 구현으로 축 자체 소멸,
   t435 인용(0/60·6.01s)이 계속 유효 근거 / G2 = ⑤ 로 해소 / G3(범위 재확인) = 본
   카드가 tip 에서 수행 / G4 = ② 로 해소.
5. **병행 소견**: 철거 후 기본-아티팩트 변형 `EdgesSourcesMoved()`(무인자)의 마지막
   호출자(`session_start.go:342`)가 사라져 참조 0 — t448 verdict 가 예고한 「후속
   후보」 성립. `EdgesSourcesMovedFor` 는 쿼리 경로(`graph_refresh_cli.go:45`)와
   `fanin_edge.go:93` 에서 생존. **별도 정리 — 본 카드 범위 밖**(t448 verdict 명시).

## Evidence

- ② t216 착지: `git log --oneline -300 | grep t216` → `84fc6161d`·`0de28ebfd`·
  `bac2cf15b`/`52d3f7a49` (Merge branch 'WT-hook-wiring-drift' into develop);
  `grep -n mxScan internal/hook/session_start.go` → 0 히트.
- ④ t454 착지: `.moai/reports/t454/verdict.md`(⑶a 정정·⑶b 오기 판정) +
  `addendum-runtime-matcher.md` 직독; `task_created.go:102-104` 직독 — "The slog
  handler is responsible for any JSONL append; this handler only invokes slog".
- ⑤ t454 addendum: 4-slot(`.*` 캐치올 대조군 포함) 프로브표 — `envkeys.go` 포함 .go
  파일 전부 불발화, 편집 착지 검증 포함. 측정 주체 lane-5, 2026-09-03.
- ③ `sed -n '30,45p' internal/hook/config_change.go` → `NewConfigChangeHandler()` returns
  `&configChangeHandler{}` (mgr 필드 세팅 없음).
- ① `grep -n "DeferredEdgesRefresh\|edgesStale" internal/hook/session_start.go` →
  51/54-59/61/71/340-354/392-393/718-743/749 히트; `grep -rn deferredEdgesRefresh
  internal/cli/` → deps.go:226 · graph_refresh_cli.go:92-111 · 테스트 3파일.
- 충돌 검사: 미병합 11개 브랜치 각각 `git diff --name-only develop...<b> -- <4 files>`
  → 전부 0 히트.

## Baseline-attribution

- 코드 존재/부재 판정 전부: 이 워크트리 `WT-t448-residue` @ `a825183dd`, 이번 실행.
- t216/t454 착지 판정: 로컬 develop 이력의 병합 커밋 + 착지본 리포트 직독 (재실행 아님).
- t435 근거(0/60·6.01s): t448 verdict 경유 인용 — 본 카드 미재측정 (G1 대로).
- t454 런타임 프로브: 착지본 인용, 재실행 안 함 — 측정 주체 lane-5 명시.

## Gaps

- t448 `probe-results.md`(E1-E7) 본문 미재판독 — `verdict.md` 경유 인용만.
- config_change D2 · file_changed D1'(retire vs dormant) 운영자 결정 — 레인이 열 수
  없는 게이트, 상신으로 마감.
- ① 구현 결과의 검증(테스트·grep 0 참조)은 run 완료 후 `run-evidence.md` 가 담당 —
  본 문서는 착수 전 상태만 증명.
- `EdgesSourcesMoved()` 무인자 변형의 참조-0 후속 정리 — 본 카드 미수행.

## Residual-risk

- 1안 착수는 lane-1(t448) 이 보존해둔 블록의 제거다 — 0/60+6.01s 실측과 리드 방향
  확정이 근거이며, 창 병합 시 리드가 이 판단을 재차 뒤집을 여지는 이 보고서가 닫지
  않는다(t448 Residual-risk 계승).
- FileChanged 발화 도메인은 CC 런타임의 비계약 동작 — keep-dormant 선택 시 잠복
  상태가 계속되고 상류 확대를 알리는 신호가 없다(t454 Residual-risk 계승).
- 병합 순서에 따라 t216 과 같은 파일 블록이지만 t216 은 이미 착지 — 병합 충돌 위험은
  측정 시점 기준 0. 창 시점까지 새 카드가 착지하면 재흡수 필요.

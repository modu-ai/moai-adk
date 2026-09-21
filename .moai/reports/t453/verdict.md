# t453 판정서 — TestAlwaysLoadedTokenBudget 적색 (always-loaded 표면 포화 트립)

- 카드: t453 (Class B — plan 생략, run → sync)
- 브랜치: `WT-token-budget` @ `400f37eb9` (로컬 develop 팁; origin/develop `e79c010b8` +27)
- 작업일: 2026-09-03 · 측정 환경: 레인 워크트리 로컬 실행 (CI 판정 별도)

---

## Claim (주장)

1. `TestAlwaysLoadedTokenBudget` 적색은 **예고된 포화 트립**이다 — 가드 상수 주석의
   @MX:DEBT가 "이 표면은 포화 상태 … 다음에 always-loaded 파일을 늘리는 카드는 이
   가드에 부딪힌다"고 명시한 바로 그 발화다.
2. 보정 커밋 `b9efb3626`(t421, 2026-09-01) 이후 표면 성장 **+810 토큰은 전부 착지된
   카드의 교리 조항**이며 전수 귀속됐다. 지방(축소 후보) 없음, 측정 결함 없음.
3. 판정: **상한 76,400 → 77,200 보정 인상**(소모 조항 전수 열거 부착) + **대형
   always-loaded 룰 다이어트를 구속 출구로 못박음**. 측정 대상 변경·문서 축소 두
   대안은 측정으로 기각했다.

## Evidence (증거 — 실행한 명령 + 그 출력)

**재현 [로컬 실측, `400f37eb9`, 이번 실행]**

```
go test -count=1 -run 'TestAlwaysLoadedTokenBudget$' ./internal/config/ -v
  → always-loaded surface = 76939 tokens (budget 76400, headroom -539, 17 entries)
  → --- FAIL: TestAlwaysLoadedTokenBudget
```

**표면 파일별 기여 (가드와 동일 메서드 — 같은 패키지 내부 함수 재사용 임시 덤프, 커밋 안 함)**

| 토큰 | 바이트 | 파일 |
|---:|---:|---|
| 16,563 | 66,255 | `.claude/output-styles/moai/moai.md` (표면의 21.5% — 다이어트 1순위) |
| 8,567 | 34,268 | `workflow/kanban-dispatch.md` |
| 6,684 | 26,739 | `core/agent-common-protocol.md` |
| 6,329 | 25,317 | `core/verification-claim-integrity.md` |
| 5,455 | 21,822 | `core/askuser-protocol.md` |
| 5,299 | 21,197 | `workflow/session-handoff.md` |
| 4,941 | 19,766 | `CLAUDE.md` |
| 4,126 | 16,506 | `core/moai-constitution.md` |
| 4,083 | 16,332 | `workflow/cross-session-messaging.md` |
| 3,693 | 14,774 | `AGENTS.md` |
| 2,220 | 8,882 | `workflow/context-window-management.md` |
| 1,745 | 6,980 | `workflow/cache-aware-execution.md` |
| 1,650 | 6,600 | `workflow/goal-directive.md` |
| 1,600 | 6,403 | `core/moai-mcp-tools.md` |
| 1,598 | 6,395 | `workflow/main-checkout-branch-guard.md` |
| 1,398 | 5,595 | `workflow/skill-routing.md` |
| 988 | 3,952 | `core/native-idiom-and-register.md` |
| **76,939** | | **17항목 = paths: 없는 룰 14 + 고정 슬롯 3** |

**보정 시점 대비 델타 [b9efb3626 → 400f37eb9, `git ls-tree -r -l` blob 크기 실측]**

| 파일 | 바이트 | 토큰 | 착지 카드 |
|---|---:|---:|---|
| `workflow/kanban-dispatch.md` | +1,468 | +367 | t224(레인 spawn 권한 조항, 33,203→34,268 B) + t386(감사 산출물 컨벤션, 32,800→33,203 B) |
| `core/agent-common-protocol.md` | +770 | +192 | t224 (경계 분류 문단) |
| `AGENTS.md` | +545 | +136 | t196 역량 결속표 — **직전 상향(t421)이 예상했던 바로 그것** |
| `core/moai-constitution.md` | +380 | +95 | t224 (오케스트레이터 룰 불릿) |
| `core/moai-mcp-tools.md` | +83 | +20 | t236 (graph_shortest_path 카탈로그 갱신) |
| 합계 | +3,246 | **+810** | |

검산: 보정 시점 표면 = 76,939 − 810 = **76,129** — t421 주석 기록 실측값과 정확히 일치
(t196 미착지 시점). 예측 여유 135가 소진되고 −539가 된 것.

**측정 대상 변경(③) 기각 근거**: 17항목이 가드가 정의한 배포 표면(no-`paths:` 룰 14 +
3 고정 슬롯)과 정확히 일치하고, t368 배너 로고 건 같은 주입되지 않는 외래물의 계수가
없다(단방향 신실성 — 열거 밖 주입물인 `CLAUDE.local.md`(추적됨, 49,597 B)는 가드
설계상 부외이며 이번 창(b9efb3626↔400f37eb9)에서 불변이라 귀속 산술에 영향 없음).

**문서 축소(②) 기각 근거**: 유일한 de-dup 후보였던 t224의 agent-common-protocol.md
재진술은 `git show 02cf8ec39` 로 확인한 결과 **5표면 의도 설계**(kanban-dispatch 본문 /
detail 불릿 / agent-common-protocol 경계 분류 / moai-constitution 구속 / manager-lead
절제 — 각 표면이 서로 다른 역할). 되돌리면 착지 설계 재소송. 이번 성장분 자체에 지방 없음.

**창 대기 브랜치 표면 추가분 0 [8개 브랜치명 존재 확인(`git for-each-ref`) 후
`git diff develop...<branch> --stat -- .claude/rules CLAUDE.md AGENTS.md
.claude/output-styles` 전부 빈 출력]** — 이 상향이 창 재측정을 재트립시키지 않는다.

**수리 후 그린 [로컬 실측, 이번 실행]**

```
go test -count=1 -run 'TestAlwaysLoadedTokenBudget$' ./internal/config/ -v
  → always-loaded surface = 76939 tokens (budget 77200, headroom 261, 17 entries)
go test ./internal/config/ -count=1  → ok (패키지 전체)
```

## Baseline-attribution (baseline 귀속)

- 판정·측정 전부 **레인 워크트리 로컬 실행**, 트리 `400f37eb9`(WT-token-budget),
  측정일 2026-09-03. 도구: 리포 자체 `go test`(internal/config) — 별도 빌드 불필요.
- 표면 델타의 두 좌표: 트리 `b9efb3626`(t421 보정) ↔ `400f37eb9`(develop 팁).
- 독립 교차 관측: lane-9(t447 사전측정)·lane-13(t449 재측정)이 base `5a8449859` 귀속으로
  동일 적색 확인(배차문 인용; 본 레인은 `400f37eb9`에서 독립 재현).
- CI 판정: develop 병합·push는 리드 소관 — 병합 후 develop push CI가 최종 판정.

## Gaps (미검증)

- 창 대기 8장의 카드↔브랜치 매핑 중 t300·t345·t348·t353·t339는 기억 의존(브랜치명 존재은
  확인, SHA 대조는 안 함). 빈 델타 판정의 구속 검사는 창 안 rev-list 재측정(확립 절차)이
  담당 — 본 판정서의 "추가분 0"은 보조 신호다.
- 템플릿 미러(`internal/template/templates/**`)의 동일 룰 트리는 재지 않았다 — 가드가
  재포 트리만 재는 것이 설계이며, 배포판 예산은 본 카드 소관 밖.
- moai-constitution-detail.md 등 paths-scoped 파일의 성장(+2/-1 등)은 표면 밖이라 미계수.

## Residual-risk (잔여 위험)

- 상향 4회째(74,317 초안 → 76,000 → 76,210 → 76,400 → 77,200)는 DEBT 규율을
  마모시킬 수 있다. 완화: 상수 주석에 "다음 트립에 자동 정당성은 없다"를 명시하고
  @MX:UPGRADE에 측정 기반 다이어트 표적(moai.md 16.5K tok 등 4개)을 못박았다.
- 다이어트 카드는 아직 큐에 없다 — 본 판정서가 착수 근거(다음 트립 시 자동 발화).
  리드가 오퍼레이터 큐에 카드화하는 것을 권고한다.
- 인상 후 여유 261 tok(0.34%) — 단일 조항평균분. 다음 always-loaded 조항 하나면
  재트립하며, 그것이 가드의 기능이다.

## 권고 (리드 → 오퍼레이터)

1. 본 카드 develop 병합 후 창 재측정 진행 — 8장 전부 표면 추가분 0이라 순번 영향 없음.
2. "대형 always-loaded 룰 다이어트(stub + lazy loading)" 카드를 큐에 올릴 것 — t421이
   두 차례 명시적으로 연기한 근본 해결이며, 다음 트립이 그 착수 근거다.

---

## sync-audit 기록 (2026-09-03)

- **판정: PASS-WITH-DEBT — 92.5/100** (Functionality 88 · Security 100 · Craft 90 ·
  Consistency 95). 독립 감사자(sync-auditor, opus)가 전수 재실측 — 판정 문서 수치를
  신뢰하지 않고 blob을 다시 재 측정. 상세: `sync-audit-verdict.md` (같은 디렉터리).
- 감사 발견 F1 [Medium] 수리 완료: 카드별 분할을 눈대정 추정(~+900B/~+570B)으로 적어
  실측과 어긋났던 것을 blob 실측값으로 정정 — t386 +100 tok(32,800→33,203 B),
  t224 +554 tok(kanban +267 / agent-common-protocol +192 / moai-constitution +95).
  총계 +810과 파일 귀속은 처음부터 정확했음(감사자 확인).
- F2 수리 완료: "17항목 = 주입 집합 일치" 서술을 단방향(배포 표면 일치 + 외래물
  계수 없음)으로 한정하고 CLAUDE.local.md 부외 경계를 명시. F3 수리 완료: DEBT 체인
  표기에 76,210 단계 복원. F4(여유 261 tok 선택)는 체인 관례 일치로 유지.
- 감사자 확인 사항: 테스트 파일 diff 0행(미약화), 11개 변경 중 6개가 `paths:` 보유로
  표면 밖(귀속 완전성), 최소성 — 축소 대안은 실행 불가(+192 < 539, 착지 설계 되돌림)
  이므로 인상이 가장 작은 정직한 수리.


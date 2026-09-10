# moai-adk-go 프로젝트 진행 상황 (마크다운 쌍둥이 — 에이전트용 축약본)

> 측정 기준: 흡수 트리 `647cfeae7` (local develop `3084f1071` + t476 브랜치 2커밋 병합)
> 측정 시각: 2026-09-06 · 측정자: lane-15 (t476)
> 전체 서술·도해는 `progress-report-20260906.html` (human-facing) 참조.

## 핵심 수치

| 지표 | 값 |
|---|---|
| SPEC 누적 완료 | 574 (implemented 142 별도) — `spec.md` status 필드 기준 총 ~784 |
| 활성 카드 | 20장 (queued 13 · picked 7) |
| 미병합 잔량 (로컬 develop 기준, 워크트리 보유) | 9 브랜치 · 62커밋 |
| CI 적색 6원인 | 수리 3 (codemaps·t474×2) · 미수리 3 |
| 코드 규모 | non-test 1,097 파일 / test 1,775 / 235,201 LOC / 137 패키지 |
| 병합 속도 (주간, develop 머지 커밋) | W33 3 · W34 150 · W35 126 · W36 409 (금주 진행 중) |
| 미푸시 develop 적체 | 185커밋 (origin/develop = 25a3212a9 무변동, 리드 일괄 push 대기) |
| 등록 워크트리 | 80 |

## CI 적색 6원인 — 흡수 트리 재검증

| # | 원인 (25a3212a9 CI) | 흡수 트리 상태 |
|---|---|---|
| 1 | Graph Freshness — codemaps 144 | **수리** — t476, fresh 10/40 (t478 수리 게이트 기준) |
| 2 | SPEC Lint — status 필드 ×2 | **미수리** — SPEC-CODEX-E2E-MEASURE-001 plan/acceptance:5 |
| 3 | CI/Lint — errcheck | **미수리** — internal/template/catalog_tree_hash.go:60 fmt.Fprintf |
| 4 | Race — TestAlwaysLoadedTokenBudget | **미수리** — internal/config 이동, overflow 123 동일 |
| 5 | Race — TestStatusAheadBehindFromHeader | **수리** — t474, internal/core/git에서 통과 |
| 6 | Race — TestStatusBranchHeaderShapes | **수리** — t474, 〃 |

## t475 흡수 종결 판정 (중간)

- t476이 t475(codemaps stale 144)를 흡수 — 흡수 트리에서 수리 게이트 재측정 결과 codemaps 행 `value=10 threshold=40 verdict=fresh`.
- 스탬프만 찍었더라면 수리 게이트가 value>0+stale로 잡았을 것이므로, 재생성이 진짜였음이 확인됨.
- **최종 종결은 병합 트리에서의 재측정** — 창에서 develop 병합 후 동일 판별식(codemaps 행) 재측정 필요.
- mx-index(18)·edges(2) stale은 이 카드 소관 밖(별도 카드).

## 게이트 출력 (그대로)

```
codemaps  metric=described-source-diff     value=10  threshold=40  verdict=fresh
mx-index  metric=inventory-content-diff    value=18  threshold=1   verdict=stale
edges     metric=source-fingerprint-mismatch value=2  threshold=0   verdict=stale
citations metric=positive-cited-path-absence value=0 threshold=0   verdict=fresh
```

## 최근 병합 (측정분)

| 커밋 | 내용 |
|---|---|
| 3084f1071 | develop tip — WT-settings-origin |
| e883493e5 | t479 — binary-lag allowlist key-shape guard |
| 251b8c950 | t478 — codemaps 게이트 수리(재생성/스탬프 구별) |
| 5f9f61d65 | WT-t410-followups |
| becf0be68 | t480 — evidence export |

## 남은 일 (carryover)

1. 미푸시 185커밋 — 리드 일괄 push (CI 판정은 push 후 생성)
2. 소관자 없는 적색 3건 — SPEC Lint status 2줄 / errcheck 1건 / 예산 overflow 123 (항상-적재 표면 77,723 / 예산 77,600, 17항목)
3. t475 최종 종결 — 병합 트리 재측정 (창에서)
4. mx-index(18)·edges(2) stale — 별도 카드 필요
5. 작업 트리 사고 기록 — lane-1 커밋 후 03:43에 codemaps 5본이 CLI 템플릿 출력으로 덮임 + 판정서 2건 삭제(미커밋) → git restore로 복원, 손실 없음

## Gaps

- 워크트리 80개의 세션 생사 미측정 (브랜치 보유 = 기계적 교집합)
- SPEC status 필드 존재만 셈, 서술 정확성 미검증
- W33 이전 주차 병합 수 미조회
- 제품 버전 표기 미확정 (태그 상태 특이 — `git describe` 결과 "list")

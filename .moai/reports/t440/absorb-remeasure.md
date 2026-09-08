# t440 흡수·재측정 기록 — 창 병합 전 (lane-13, 2026-09-03)

카드: t440 · 브랜치 `WT-delivery-notice-docs` · 판정 트리 `3dc389d09` (창 직전 최종).

## 배경

창 순번 6번 배정에 따른 사전 흡수·재측정. 1차 흡수에서 `origin/develop`(400f37eb9,
뒤진 쪽)을 기준으로 삼는 오류가 있었고, 리드 정정으로 **로컬 develop 팁**을 기준으로
재흡수했다. 판정은 최종 기준 트리에서만 유효하다(1차 재측정 기록은 기각).

## 병합 이력

| 순서 | 대상 | 결과 | 트리 |
|---|---|---|---|
| 병합 1 | origin/develop 400f37eb9 | ort 충돌 0 (기준 오류 — 기각) | c9de6307e |
| 병합 2 | **로컬 develop 3bdd5a803** | ort 충돌 0, 183 files +9936 | **3dc389d09** |

전 단계 `git status --porcelain` 0행 확인 후 병합. 최종 상태: **behind 0 / ahead 4**
(t440 본체 63ea8693a + 병합 커밋 2 + 본 증거 커밋), porcelain 0행.

## 유입 델타 직독 (merge-base 400f37eb9..3bdd5a803)

- docs-site: `cli-reference/{profile,worktree}.md` ×4로케일 = 8파일 (t297분) —
  **moai-sync.md 무접촉** (커밋 63ea8693a와 내용 교집합 0)
- .go 51개 (internal/cli·core·config·profile·statusline·web 등) — 카드 검증 표면
  (hugo 빌드·패리티)과 무의존. 규율 ①의 `go list -deps` 분석 축이 카드에 없는 케이스.

주의: `c9de6307e..3bdd5a803` 2점 diff에는 moai-sync.md ×4가 보이나, 이는 내 커밋이
develop에 없어서 생기는 역방향 비대칭이다. 유입 판정은 merge-base 기준 3점이 정확하다.

## 재측정 — 원 verdict 동일 레시피, 트리 3dc389d09에서 8/8 PASS

| 검증 | 관측 |
|---|---|
| moai-sync.md 섹션 패리티 | ko 46 · en 46 · ja 46 · zh 46 (원판정 46×4 불변) |
| 전체 트리 랫chet | now 54 / base 54페이지 · 신규 이격 0건 · 수렴 0건 |
| hugo build | rc=0 · WARN/ERROR 0건 · 4289 ms |
| sitemap | docs-site/public/sitemap.xml 존재 |
| URL 블랙리스트 | rc=1 (0매치) |
| Mermaid 방향 | rc=1 (0매치) |
| 본문 이모지 (편집 4파일) | rc=1 (0매치) |

Verdict(`verdict.md`)의 8항목 표와 동일 레시피·동일 결과. hugo 원로그:
`.moai/state/verify/lane13-t440-remeasure/hugo-build-v2.log` (스크래치 — 착지 시 소멸,
본 파일이 인용 대상).

## Gaps

- Go 테스트 재실행 없음 — 카드 델타에 .go 0개, 증거 표면이 문서 축인 것이 diff로 확인됨.
  창 병합 트리에서 Go 축 전체 재판정은 리드 소관(규율 ②).
- 최종 창 병합 트리(3dc389d09 + 창 내 선행 카드 병합분)에서의 재판정은 창 시점 수행.

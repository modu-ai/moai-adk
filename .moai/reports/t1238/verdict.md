# t1238 판정서 — codemaps 재생성 (described-source-diff)

- 카드: t1238 · 브랜치 `WT-codemaps-refresh3` (`WT-codemaps-refresh` 는 병합 끝난 t1132 가 점유) · 기준 로컬 develop `38148d891`
- 커밋: `4a05fd3d6` (codemaps 본문) · `16d5fa916` (provenance 재스탬프)
- PR 교차확인: no-link (리드 지정)

## 주장
재생성 후 `codemaps described-source-diff` 가 40 미만이고 `citations positive-cited-path-absence` 가 0 이다.

## 증거
측정 명령은 두 번 모두 `moai graph check` (출력은 세션 스크래치패드로만 리다이렉트, `.moai/reports/` 밑으로 보내지 않음).

재생성 전 (HEAD `38148d891`):
```
codemaps  metric=described-source-diff value=77 threshold=40 verdict=stale
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
  measured from: 65cc8df48 (last-body-change)
```

재생성 후 (HEAD `16d5fa916`):
```
codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
```

Fold/Omission 판정 검사 (codemaps.md Phase 4 스크립트): `COLLECTED: fold=37 omission=29`, UNCOVERED·FOLD-PROSE·GAP 0줄.

## 기준 귀속
두 측정 모두 이 워크트리(`.claude/worktrees/t1238`)에서 이번 실행에 잰 값이다. 카드 본문의 43 은 이전 develop 기준 값이고, 이 트리에서는 그 사이 병합분이 더해져 77 이었다.

## 변경
`.moai/project/codemaps/` 4개 파일 (+33/−13): modules.md (`internal/contract` 계열 신설 행, cli 클러스터 수치·설명 갱신), dependencies.md (contract fan-in 행), entry-points.md (`moai contract verify|show|sign`), overview.md (레이어 표). provenance.json 재스탬프.

## 미검증
- `moai graph check` 전체 exit 는 1 이다. 원인은 mx-index·edges 층의 `absent`(새 워크트리에는 미추적 파생 산출물이 없음)이며 재생성 전과 동일하다. 이 카드 범위 밖이다.
- 수정만 된 파일 일부(config defaults 등, rosterguard 는 기존 omission)는 문서가 원래 서술 단위로 삼지 않아 본문에 추가하지 않았다.

## 잔여 위험
- develop 에 다른 카드가 병합되면 창에서 흡수할 때 값이 다시 올라간다. 병합 트리에서 재측정해야 한다.
- 문서 서술의 정확성은 인용 경로 존재 검사(citations)까지만 기계 확인됐다.

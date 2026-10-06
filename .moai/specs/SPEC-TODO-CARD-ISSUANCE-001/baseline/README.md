# M0 기준선 재현 절차 (SPEC-TODO-CARD-ISSUANCE-001, card t1454)

이 디렉터리의 `baseline.md` 는 2026-10-05 측정이다. 다시 재는 절차:

## 0. 고정점

- 측정 시점 카드 브랜치 끝(card-head): `4bc76ceb3`
- 흡수한 develop 팁(git figure 의 기준 커밋): `4964d0796`
- 큐 스냅숏: 라이브 DB `/Users/goos/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` 를 복사해 고정
  (M0 측정분 sha256 `0169ff35…7d6a` — 전체값은 baseline.md 머리)
- 측정 기간: `--since=2026-08-13 --until=<측정 시각 +09:00>`

## 1. 스냅숏과 원자료 TSV

```
cp /Users/goos/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db <scratch>/queue-snapshot.db
git log --first-parent --since=2026-08-13 --until=<NOW+09:00> --format='%H%x09%P%x09%cI%x09%s' 4964d0796 > <scratch>/fp.tsv
git log --since=2026-08-13 --until=<NOW+09:00> --format='%h%x09%cI%x09%an%x09%x09%s' 4964d0796 > <scratch>/allc.tsv
```

큰 TSV(`fp.tsv`·`allc.tsv`)와 `commits.json`·`delivered.json` 은 추적하지 않는다 — 위 절차로 다시 만든다.

## 2. 스크립트 실행

`scripts/` 를 scratch 로 복사하고 **경로 상수만** 바꾼다(M0 실행은 tracked 사본을 손대지 않는다):

- `lib.py` 의 `DB` → 스냅숏 사본의 read-only URI(`file:<스냅숏 경로>?mode=ro`)
- `t1454_*.py` 다섯 파일의 `DB` → 스냅숏 사본 경로
- `03_git_collect.py` 의 `G` → git 히스토리를 읽을 워크트리 아무 경로(같은 저장소)
- `OUT` 은 스크립트 자신의 디렉터리라, fp.tsv·allc.tsv 는 스크립트 사본 디렉터리에 둔다

실행 순서:

```
01_inventory.py 02_origin_graph.py 03_git_collect.py 04_granularity.py \
05_dups_stale_lanes.py 06_followup_overlap.py \
t1454_jaccard.py t1454_issuance_sim.py t1454_idf_sim.py t1454_recall.py t1454_paths.py
```

`delivered.json` 은 카드 세션의 inline 산출물이라 번호 스크립트에 없다 — `commits.json` 에서
카드 id 첫 값으로 aggregate해 만든다: `{cid: {t: 마지막 커밋 시각, lines: Σ(add+del),
files: 전체 경로, prim: 최다 변경 경로}}`. 이 모양은 원본 scratch 산출물
(`…/061c4c0e…/scratchpad/cards/delivered.json`, 964장)과 같다.

## 3. 기대 헤드라인 (M0 측정값 — baseline.md 행과 같다)

- fp.tsv 1,357행, 카드 부착 1,239(대체 146) · 큐 1,231행(live 129 + archived 1,102)
- GB02: first-parent 커밋 가진 카드 1,045 중 스냅숏 존재 948 · GB15 허브 상위: catalog.yaml 121(20)
- SB01: 기계 near 구간 [0.80,1.0) 0장 · SB07: 깊이 3 공유 쌍 0
- 개수 figure: WT-* 422 · 워크트리 149 · SPEC 디렉터리 1,050 · `card:` SPEC 36 (측정 시각 값)

## 4. 허브 목록 재측정

`hub-files.txt` 머리의 명령 형태로 다시 센다. 컷은 `baseline.md` 임계값 표(72시간 창 최대 겹침 ≥ 10).
개명 이력이 생기면 두 이름의 이력을 합치거나 `--follow` 를 쓰고 그 방식을 머리 주석에 적는다(현재 5경로는 개명 없음).

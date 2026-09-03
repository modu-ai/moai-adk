# t460 판정 — `WT-lane-spawn-authority@8b6c30f77` 조상 밖 흡수 커밋

- **Date**: 2026-09-03
- **Card**: t460 (Class B — plan 생략, run → sync)
- **Worktree**: `.claude/worktrees/t460` · branch `WT-absorb-verdict`
- **Measured against**: `origin/develop = 400f37eb92bfeb9fcef23373d7ba8bf91d1fc764` (본 세션 `git fetch origin develop` 직후 재측정 — 리드 전달값을 신뢰하지 않고 자체 측정)

## Claim

`WT-lane-spawn-authority` 브랜치 팁 `8b6c30f77`은 develop 계보 밖에 걸린 흡수 커밋이지만, **고유 콘텐츠가 0** 이다. 따라서 브랜치는 **삭제 가능**하며, 삭제로 잃는 것은 도달 불가능해지는 계보 커밋 객체 1개뿐이다 (콘텐츠 손실 없음). 원인은 창 절차의 순서 역전 — develop 병합이 흡수보다 먼저 실행됐다.

## Evidence

### 1. 조상 관계 (fetch 직후 자체 측정)

```
origin/develop tip: 400f37eb92bfeb9fcef23373d7ba8bf91d1fc764
9dde832d8 IN origin/develop: YES
8b6c30f77 IN origin/develop: NO
8b6c30f77 IN local develop:  NO
```

(`git merge-base --is-ancestor` 각각 실행, 종료코드 판독)

### 2. 흡수 커밋의 구조

```
8b6c30f77 | parents: 02cf8ec39c170df514a01d1177a92dabf5f2d129 9dde832d83ae8414a7c9e3f7351f4a76e2843a6c | 2026-09-02 23:50:40 +0900 | Merge branch 'develop' into WT-lane-spawn-authority
9dde832d8 | parents: 48c35a4d4eb88e2272c5b1fce6a37d4c6e83024b 02cf8ec39c170df514a01d1177a92dabf5f2d129 | 2026-09-02 23:50:18 +0900 | Merge branch 'WT-lane-spawn-authority' into develop (card t224)
```

- develop 병합 `9dde832d8` (23:50:18) → **22초 뒤** 흡수 `8b6c30f77` (23:50:40).
- 흡수의 parent1 = `02cf8ec39`(카드 브랜치 팁) — 이 커밋은 이미 `9dde832d8`의 parent2다. 즉 흡수는 "이미 develop에 합쳐진 자기 브랜치"에 develop을 다시 얹은 역산 병합.
- `9dde832d8..8b6c30f77` 사이 커밋: 흡수 커밋 1개뿐.

### 3. 트리 동등성 — 고유 콘텐츠 0의 결정적 증거

```
tree(8b6c30f77) = 614fae83140a21bd696aabeaf4b9eaf5ebf170c5
tree(9dde832d8) = 614fae83140a21bd696aabeaf4b9eaf5ebf170c5   ← 바이트 동일
tree(02cf8ec39) = be7831ee2886c55ad215a230b5f6579b960cd072

$ git diff 9dde832d8 8b6c30f77 --stat   → 출력 없음 (exit 0)
```

흡수 병합의 merge-base가 `02cf8ec39`(develop 병합의 조상)이므로 병합 결과 트리가 develop 쪽 트리와 동일해진 것 — 충돌 해결분조차 없다.

### 4. 삭제 시 잃는 것의 전수 목록

```
commits in 8b6c30f77 not in origin/develop: 1
8b6c30f77 Merge branch 'develop' into WT-lane-spawn-authority   ← 이 1개뿐
```

`8b6c30f77`에 도달하는 ref는 `WT-lane-spawn-authority` 하나 — 브랜치 삭제 시 이 커밋 객체만 도달 불가능해진다. 그 트리는 develop 계보 안의 `9dde832d8` 트리와 동일하므로 **콘텐츠 손실 0**.

## Judgment

| # | 판정 항목 | 결론 |
|---|---|---|
| 1 | 조상 밖 커밋의 내용 | 흡수 커밋 1개, 트리가 `9dde832d8`와 바이트 동일(`614fae83…`) — **고유 콘텐츠 0** |
| 2 | 별도 착지 vs 폐기 | **폐기(삭제 가능)**. 별도 착지할 변경 없음. t458(lane-7)의 잔재 정리에서 `WT-lane-spawn-authority` **삭제 제외를 해제**해도 된다 |
| 3 | 재발 방지 | 아래 절차 수정 참조 |

### 판정 3 — 재발 방지: 창 절차의 어디를 고쳐야 하는가

**원인**: lane-7이 t224 창을 **선병합 후 역산 흡수** 순서로 실행 — develop 병합(`9dde832d8`, 23:50:18)을 먼저 하고, 그 22초 뒤 카드 워크트리에서 흡수(`git merge develop` → `8b6c30f77`, 23:50:40)를 했다. 카드 브랜치가 develop에 합쳐진 **후에** 브랜치에 얹히는 커밋은 정의상 develop 계보 밖에 놓인다.

**고칠 것** — 독트린 자체(`CLAUDE.local.md` §4.1: acquire → 흡수 → 재측정 → develop 병합)는 이미 올바른 순서를 규정하므로, 새 독트린이 아니라 **감지와 규율**이 답이다:

1. **레인 규율**: 흡수(`git merge origin/develop`)는 develop 병합 **이전**, 같은 창 hold 안에서 끝낸다. 자기 카드의 develop 병합이 이미 착지한 뒤에는 카드 브랜치에 그 어떤 커밋도 얹지 않는다 — 흡수를 빠뜨렸다는 것을 뒤늦게 깨달으면 리드에 보고하고, 이미 병합된 브랜치에 절대 역산 흡수하지 않는다.
2. **리드 감지(기계적)**: 카드 done 승인 전 레인이 보고한 브랜치 팁에 대해 `git merge-base --is-ancestor <팁> origin/develop` 검사. 이 사례에서는 흡수가 병합 22초 뒤라 병합 시점 검사로는 못 잡았을 것이므로, **잔재 정리 시점**(t458류)의 조상 검사가 실질 방어선이다 — 그리고 조상 밖으로 밝혀진 브랜치는 조상 검사만으로 삭제 판정하지 않고, **트리 비교**(이 판정서 §3 레시피)를 한 뒤 삭제한다. 조상 밖 = 손실 위험 이 아니라, 조상 밖 + 트리 동일 = 안전 삭제다.
3. **건전한 쌍둥이 대조**: 같은 날 t444의 흡수 `86862826a`("Merge branch 'refs/heads/develop' into WT-doctor-freshness-reds")는 develop 병합 `5107bbfff` **이전**에 이뤄져 develop 계보 안에 있다. 올바른 순서의 실례로 기록해 둔다.

## Baseline-attribution

- 측정 시각: 2026-09-03, 본 세션(lane-6, source_session_id: a6521ac3-8da8-4371-beb5-c0f899976adb)
- 측정 트리: `.claude/worktrees/t460` @ `WT-absorb-verdict` (본 커밋)
- 판독 ref: `origin/develop = 400f37eb92bfeb9fcef23373d7ba8bf91d1fc764` (측정 직전 `git fetch origin develop`로 갱신)
- 모든 SHA는 40자 전문으로 기록 (이동 ref 고정: `9dde832d83ae8414a7c9e3f7351f4a76e2843a6c` / `8b6c30f77afcdf1b62f62492a3b9d64f964ffbdd` / `02cf8ec39c170df514a01d1177a92dabf5f2d129` / 트리 `614fae83140a21bd696aabeaf4b9eaf5ebf170c5`)

## Gaps

- lane-7 레인 본인에게 당시 창 실행 순서의 사유를 직접 확인하지 않았다 — git 기록(커밋 subject·부모 구조·22초 시차)이 순서 역전을 완전히 재구성하므로 확인 불요로 판단. 리드가 lane-7 컨텍스트를 원하면 별도 문의 가능.
- `8b6c30f77` 객체의 실제 gc 회수 시점은 관찰하지 않았다 (객체 도달성만 측정 — gc는 git 내부 스케줄).

## Residual-risk

- 브랜치 삭제 전 누군가 `8b6c30f77` SHA를 어딘가에 인용해 두었다면 그 인용이 도달 불가능 객체를 가리키게 된다 — 본 판정서가 40자 전문을 보존하므로 기록 복원은 가능.
- 향후 develop에 force-push 등 계보 재작성이 있으면 본 판정의 조상 관계가 무효화될 수 있다 (현재 원격은 merge-only 운용이라 위험 낮음).

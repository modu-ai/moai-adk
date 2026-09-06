# t478 — 임계 40 적정성 분석 (운영자 상신용)

측정 트리: `.claude/worktrees/t478`, HEAD `456665e8d`
측정 대상 이력: 로컬 `refs/heads/develop` (`7896f30d7`)
측정 일자: 2026-09-04 · 전부 **로컬 측정** (CI 아님)

## Claim

임계 **40 은 이 리포의 병합 리듬에 대해 과민하지 않다.** 리드가 인용한 `144` 는
단일 push 의 델타가 아니라 **마지막 스탬프 이후 누적치**이며, 그 대부분은
**공백만 바뀐 gofmt 병합 한 건**에서 나왔다. 조정 후보는 임계값이 아니라
**"무엇을 드리프트로 셀 것인가" 라는 술어** 쪽이다.

## Evidence

### E1 — 병합당 described-worthy `.go` 파일 수 분포 (develop first-parent 최근 30)

명령:

```
git log --first-parent --diff-merges=first-parent -30 --format='C %h %s' --name-only -- internal cmd pkg
```

집계:

```
n=30   median=2   max=154
임계 40 이상인 병합: 1 / 30
```

상위 6건:

```
 154  d41ba7479 Merge card t457 (WT-gofmt-drift) — gofmt entire tree
  20  3345cc015 Merge branch 'WT-precommit-gate-scope'
  19  e969dc07d Merge branch 'WT-project-continuation' (card t191)
  13  52d3f7a49 Merge branch 'WT-hook-wiring-drift'
  13  456665e8d chore(t478): absorb develop tip
  10  d102892c0 Merge branch 'WT-goal-prose-arm' (card t436)
```

정상 카드 병합은 한 자리~20 대이고, 임계를 넘긴 것은 전수 포맷팅 1건뿐이다.

### E2 — 그 1건은 실질 변경이 아니다

```
$ git diff -w --shortstat d41ba7479^1 d41ba7479 -- internal cmd pkg
 31 files changed, 61 insertions(+), 22 deletions(-)

$ git diff    --shortstat d41ba7479^1 d41ba7479 -- internal cmd pkg
 154 files changed, 801 insertions(+), 762 deletions(-)
```

154건 중 **123건이 공백만 바뀐 파일**이다. 코드맵 산문이 낡았는지와 무관한 변화다.

### E3 — 현재 누적치

```
$ ./bin/moai graph check
codemaps  metric=described-source-diff value=146 threshold=40 verdict=stale
```

스탬프 커밋은 `ad272be20`(2026-09-02). `146` 은 그 시점 이후 누적이며 E2 의 123건을 포함한다.

## Baseline-attribution

E1/E2 는 이 워크트리(HEAD `456665e8d`)의 `refs/heads/develop`(`7896f30d7`) 이력에서
위에 적힌 명령을 그대로 실행해 얻은 출력이다. E3 은 이 트리의 소스로 빌드한
`./bin/moai`(`go build -o ./bin/moai ./cmd/moai`, HEAD `456665e8d`)의 출력이며
설치본(`~/go/bin/moai`)이 아니다.

## 결론 (운영자 결정 사항)

1. **임계 40 유지 권고.** 최근 30 병합 중 29건이 20 이하이므로, "대량 병합마다
   발화한다" 는 전제는 측정으로 성립하지 않는다.
2. **별도 카드 후보** — `described-source-diff` 술어가 공백만 바뀐 파일을 드리프트로
   센다. `git diff -w` 기준 또는 gofmt-정규화 비교로 바꾸면 E2 의 123건이 빠진다.
   이 SPEC 범위 밖이며, 임계값을 건드리지 않고 신호 대 잡음을 개선하는 방향이다.

## Gaps

- 30 병합 창 **밖**(그 이전 이력)은 측정하지 않았다.
- E1 의 awk 필터는 `\.go$` 이므로 `mx.IsDescribedWorthy` 의 전체 술어와 1:1 대응하지
  않는다 — E3 의 `146` 만이 바이너리가 낸 정식 값이고, E1 은 분포 파악용 근사다.
- CI 상의 값은 재지 않았다.

## Residual-risk

포맷팅 무시 술어로 바꾸면 들여쓰기로만 표현되는 블록 구조 변화를 `-w` 가 흡수해
진짜 드리프트를 놓칠 여지가 생긴다. 그래서 이 문서는 임계 유지 + 술어 변경은
별도 카드 분리를 권고한다.

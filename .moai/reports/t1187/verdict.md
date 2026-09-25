# t1187 판정 — codemaps described-source 갱신

card: t1187 · branch `WT-codemaps-source-refresh` · worktree `.claude/worktrees/t1187`

## Claim

`origin/develop` `a8a9b9376`의 기술 변경을 코드맵 본문 다섯 파일에 반영했다.
`bd71c59e4` 스탬프 이후 변한 설명 대상 소스 41개를 본문 커밋
`c6a9fcaf6`으로 흡수한 뒤 그 커밋을 스탬프했다. codemaps 층은 fresh다.

## Evidence

- `git diff --name-only bd71c59e4 HEAD -- internal cmd pkg | wc -l` → `126`.
  테스트·fixture를 포함한 전체 파일 수이며 게이트 수치로 대체하지 않았다.
- 이 워크트리에서 `go build -o /tmp/moai-t1187-graph ./cmd/moai` → exit 0.
- 수정 전 `/tmp/moai-t1187-graph graph check --json` →
  `codemaps metric=described-source-diff value=41 threshold=40 verdict=stale`.
- 본문 다섯 파일 커밋 `c6a9fcaf6` 직후 같은 검사 → `value=41 verdict=stale`.
  본문을 바꿨다는 사실만으로 새 스탬프가 생기지 않았다.
- `/tmp/moai-t1187-graph graph stamp codemaps --commit c6a9fcaf6` →
  `OK: stamped .../provenance.json`, 기록된 `commit_sha=c6a9fcaf6f5e28828417fff65bbb9ddcee7ce215`.
- 스탬프 뒤 `/tmp/moai-t1187-graph graph check --json` →
  `codemaps value=0 threshold=40 verdict=fresh`, `citations value=0 verdict=fresh`.
  전체 명령 exit 1은 `mx-index`·`edges`가 이 새 워크트리에서 `absent`이기 때문이다.
- 현재 트리에서 `find internal cmd pkg` → 비테스트 Go `1307`, 테스트 Go `2302`;
  `go list ./...` → `152` 패키지; `go list -f` 기반 내부 import → 패키지 엣지
  `386`, 최상위 엣지 `241`; `find internal/template/templates -type f` → `591`.
- `git diff --check` → exit 0.

## Baseline-attribution

모든 측정은 `.claude/worktrees/t1187`의 `origin/develop` fast-forward 결과
`a8a9b9376`과 본문 커밋 `c6a9fcaf6`에서 이 실행에 수행했다. 설치된 `moai`
바이너리는 쓰지 않고 이 트리에서 빌드한 `/tmp/moai-t1187-graph`를 썼다.

## Gaps

- `mx-index`와 `edges.jsonl`은 새 워크트리에서 absent이며 갱신하거나 fresh로 판정하지 않았다.
- Windows 런타임이나 실제 Factory 배차 동작은 이 문서 카드의 검증 대상이 아니다.
- 원격 push·develop 병합·CI는 이 레인에서 하지 않았다. 병합은 스탬프 대상
  `c6a9fcaf6`이 이력에 남는 방식이어야 한다. squash로 버리면 비교 불가가 된다.

## Residual-risk

문서의 과거 판정 서술은 그대로 남긴 곳이 있다. 이번에 측정한 최신 수치와
신규 실행 경로는 갱신했지만, 41개 소스 파일의 모든 분기를 개별 사례로
실행하지는 않았다. `graph check`의 fresh는 내용의 의미 정확도 전체를 증명하지 않는다.

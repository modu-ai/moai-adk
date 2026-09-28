# t1295 — codemaps 갱신 판정

## Claim

기준 `fdc5361c3` 이후 달라진 구조 가운데 MoAI L1 워크트리 경로와 Factory 역할 전환을 다섯 codemap에 반영했다. 이 워크트리의 codemaps freshness는 `fresh`다.

## Evidence

측정 트리: `WT-codemaps-freshness`, 시작 HEAD `cee197917b83aff5b04761e13d5a6328179cfee1`.

```text
$ git diff --name-only fdc5361c3 HEAD -- cmd internal pkg | rg '^((cmd|internal|pkg)/).+\.go$' | rg -v '_test\.go$' | wc -l
      81
$ find internal cmd pkg -name '*.go' -not -name '*_test.go' | wc -l
    1419
$ find internal cmd pkg -name '*_test.go' | wc -l
    2544
$ go list ./... | wc -l
     164
$ find internal/template/templates -type f | wc -l
     598
$ git diff --check
(출력 없음, exit 0)
$ moai graph stamp codemaps --commit cee197917 --root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1295
OK: stamped /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1295/.moai/project/codemaps/provenance.json
provenance: tree=/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1295 commit=cee197917b83
$ moai graph check --json --root /Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1295
{
  "tree_root": "/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1295",
  "layers": [
    {
      "layer": "codemaps",
      "metric": "described-source-diff",
      "value": 0,
      "threshold": 40,
      "verdict": "fresh",
      "contribution": 0,
      "contribution_base": "f26b6997040026bb798b923191922fd48d67ad91",
      "content_anchor": "cee197917b83aff5b04761e13d5a6328179cfee1",
      "content_anchor_source": "working-tree-differs-from-stamp"
    },
    {
      "layer": "mx-index",
      "metric": "inventory-content-diff",
      "value": 0,
      "threshold": 1,
      "verdict": "absent",
      "reason": "mx-index absent (untracked runtime artifact — fresh worktree state)"
    },
    {
      "layer": "edges",
      "metric": "source-fingerprint-mismatch",
      "value": 0,
      "threshold": 0,
      "verdict": "absent",
      "reason": "edges.jsonl absent (untracked derived artifact — fresh worktree state)"
    },
    {
      "layer": "citations",
      "metric": "positive-cited-path-absence",
      "value": 0,
      "threshold": 0,
      "verdict": "fresh"
    }
  ]
}
(전체 명령 exit 1; stderr의 absent 진단은 Gaps에 기록)
```

`go list -deps -json ./...` 결과를 파싱해 프로젝트 패키지 164개, 내부 import 452개, 최상위로 접은 엣지 278개를 셌다. 최상위 fan-out은 `internal/cli` 70, `internal/hook` 39다.

최신 로컬 `develop` `948a59b6c`를 흡수한 뒤 같은 `graph check`에서 codemaps 계층은 `fresh`, 값 11, 문턱 40이었다(전체 exit 1: 새 워크트리의 `mx-index`·`edges` 부재).

## Baseline-attribution

위 수치와 freshness는 이 워크트리의 `cee197917` 소스에서 이번 실행에 측정했다. 스탬프가 가리키는 커밋은 브랜치 전용 커밋이 아니라 현재 `develop`의 조상이다. 과거 문서의 상세 fan-in/fan-out 표는 해당 판의 기록이라고 명시했다.

## Gaps

`mx-index`와 `edges`는 새 워크트리에 런타임 산출물이 없어 `absent`이며 전체 `graph check`는 exit 1이다. 이 카드에서는 두 계층을 생성하거나 신선하다고 주장하지 않는다. 이 보고서 작성 시점에는 새 병합 커밋의 CI 판정과 큐 완료 처리가 없다.

## Residual-risk

다른 카드의 병합으로 described source가 다시 바뀌면 재측정과 스탬프 갱신이 필요할 수 있다. 이전 판의 상세 패키지 판단은 이번 부분 갱신에서 전수 재검증하지 않았다.

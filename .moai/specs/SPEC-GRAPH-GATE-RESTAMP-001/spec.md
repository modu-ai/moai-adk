---
id: SPEC-GRAPH-GATE-RESTAMP-001
title: "codemaps 게이트의 맨손 재스탬프 위조-초록 차단 — 측정 기점을 본문 최종 변경 커밋으로 이동"
version: "0.1.3"
status: completed
created: 2026-09-04
updated: 2026-09-06
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/graph"
lifecycle: spec-anchored
tags: "graph, codemaps, freshness, gate, anchor, restamp, false-green"
era: V3R6
tier: M
related_specs: [SPEC-GRAPH-FRESHNESS-CADENCE-001, SPEC-V3R6-GRAPH-FRESHNESS-001, SPEC-V3R6-GRAPH-FRESHNESS-002, SPEC-STAMP-REACHABILITY-001]
---

# SPEC-GRAPH-GATE-RESTAMP-001 — codemaps 게이트 재스탬프 위조-초록 차단

## HISTORY

| Version | Date | Change | Author |
|---------|------|--------|--------|
| 0.1.3 | 2026-09-04 | **§B.4와 §D.1의 거짓 문장 2건을 정정**(manager-develop이 하류에서 발견, 레인 결정). ① §B.4는 "모든 기존 codemaps 층 픽스처가 본문을 미추적으로 남긴다"고 했으나 거짓 — 내부 헬퍼 `writeCodemapsProvenanceBlock`은 본문을 **아예 쓰지 않으며**, 이를 직접 호출하는 픽스처는 본문이 없다. ② §D.1은 본문-부재 상태가 선행 "codemaps directory missing" 분기에서 이미 처분된다고 했으나 거짓 — 그 분기는 `os.Stat(dir).IsDir()`이라 `provenance.json`만 든 디렉터리는 통과한다. 본문-부재는 **도달 가능**하고 규칙 C로 떨어진다. 결정: **규칙 C를 C1/C2로 분리**한다 — C1(본문 부재) = absent + **시스템 오류 없음**(exit 1), C2(앵커 해석 불가) = absent + 시스템 오류(exit 2, 불변). C1은 형제 층 `checkCitations`의 `docs == 0` 처분과 일치시키는 **정정**이지 새 발명이 아니다. AC-4를 C1 형태로 재작성하고 C2의 도달 불가성을 명시, AC-5에 픽스처 1건 편집을 허가했다. | manager-spec |
| 0.1.2 | 2026-09-04 | 규칙 A의 두 번째 구멍 종결(레인 결정). 기존 픽스처의 codemaps 본문이 **미추적**이라는 실측(§B.4)에 따라, 규칙 A의 프로브를 `git diff`와 `git ls-files --others --exclude-standard`의 **합집합**으로 정정했다 — `gitDiffNameList`가 described roots에 이미 쓰는 관용구를 재사용한다. 정정 전이라면 기존 codemaps 층 테스트가 전부 규칙 C(absent)로 떨어져 AC-5가 구조적으로 실패했을 것이다. 규칙 B/C의 로직은 불변이며, 정정 이후 규칙 C가 거의 도달 불가능해지는 사정을 §D.1에 명시했다. `acceptance.md`에 AC-7(미추적 본문 픽스처 회귀 잠금)을 신설했다. | manager-spec |
| 0.1.1 | 2026-09-04 | §D.2의 미해소 항목 종결. 레인이 규칙 A 프로브의 오명명 지적을 수용해 결정했다: 규칙 A의 로직은 배차문 그대로 두고, `content_anchor_source` 토큰만 `working-tree-differs-from-stamp`로 개명한다. §D.1 표, §C REQ-GGR-004, §G Q2를 그에 맞춰 갱신하고 `[NEEDS CLARIFICATION]` 표식을 제거했다. 이 수리가 닫지 **않는** 잔여 위험 1건(본문에 사소한 미커밋 편집을 넣고 스탬프하면 규칙 A로 통과)을 §I에 신설해 명시적으로 범위 밖으로 선언했다. | manager-spec |
| 0.1.0 | 2026-09-04 | 카드 t478 plan-phase 최초 저작. §B의 모든 수치는 워크트리 `.claude/worktrees/t478`(브랜치 `WT-graph-gate-restamp`, HEAD `456665e8d`)에서 직접 측정했다. 배차문이 제시한 앵커 규칙 A/B/C를 그대로 채택하되, 규칙 A의 프로브가 이 트리에서 실제로 발화하는 것을 관측해 §D.2에 판정으로 기록하고 `content_anchor_source` 토큰 명명을 운영자 결정 항목으로 남겼다. | manager-spec |

## §A. 문제 진술

`moai graph check`의 codemaps 층은 `described-source-diff` 지표로 판정한다. 정의는
`internal/graph/check.go`의 `checkCodemaps`에 있으며, **`.moai/project/codemaps/provenance.json`의
`commit_sha`가 지목한 커밋**과 작업 트리 사이에서 described roots(`internal`, `cmd`, `pkg`) 아래
described-worthy 파일이 몇 개나 달라졌는지를 센다. 40 이상이면 stale이다.

한편 `moai graph stamp codemaps`(`internal/cli/graph_stamp.go` → `mx.StampCodemaps`)는 provenance
**만** 쓴다 — tree root, HEAD(또는 `--commit`) sha, described roots, dirty/content fingerprint. codemaps
**본문**(`.moai/project/codemaps/*.md`)은 한 글자도 읽지 않는다.

따라서 본문을 그대로 둔 채 스탬프만 다시 찍으면 측정 창이 0으로 초기화되고, 낡은 산문 위에서 게이트가
초록이 된다. **게이트를 통과하는 가장 값싼 경로가 곧 위조-초록이다.** 오늘 이 경로를 막는 것은 카드
t475 본문에 적힌 산문 주의 한 줄뿐이며, 그 카드를 읽지 않은 유지자가 스탬프를 찍으면 게이트는 통과한다.

## §B. 측정된 베이스라인

모든 수치는 워크트리 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t478`, 브랜치
`WT-graph-gate-restamp`, HEAD `456665e8d`에서 2026-09-04에 측정했다. 각 행은 그 값을 만든 명령을 함께
적는다.

### B.1 — 맨손 재스탬프가 실제로 게이트를 뒤집는다 (배차 측정, 이 트리)

| 단계 | 명령 | 관측 |
|---|---|---|
| 1 | `./bin/moai graph check` | codemaps `metric=described-source-diff` `value=146` `threshold=40` `verdict=stale` |
| 2 | `./bin/moai graph stamp codemaps` (본문 미편집) | rc=0, `commit=456665e8d59e` |
| 3 | `./bin/moai graph check` | codemaps `value=0` `threshold=40` `verdict=fresh` |

이후 `provenance.json`은 추적본으로 되돌렸고 `git status --porcelain .moai/project/codemaps/`는 깨끗하다.

### B.2 — 앵커 규칙 B의 보수성 (이 트리에서 재측정)

스탬프된 sha `S = ad272be20abff9e4f3b1b363fce3e48dac4c5132`(현재 추적본 provenance의 `commit_sha`,
`dirty=false`, `described_roots=[internal, cmd, pkg]`)에 대해:

| 명령 | 출력 |
|---|---|
| `git log -1 --format=%H S -- .moai/project/codemaps/ ':(exclude).moai/project/codemaps/provenance.json'` | `2f28bc394718c2cb49a6dc96577d12c6da35b05d` |
| `git merge-base --is-ancestor 2f28bc394 ad272be20` | rc=0 (조상 성립) |
| `git diff --name-only ad272be20 -- internal cmd pkg \| wc -l` | `405` |
| `git diff --name-only 2f28bc394 -- internal cmd pkg \| wc -l` | `570` |

앵커를 S에서 본문 최종 변경 커밋으로 옮기면 값은 405 → 570으로 **커진다**. 즉 이 이동은 게이트를 더
붉게만 만들 뿐 결코 더 초록으로 만들지 않는다. (405/570은 `IsDescribedWorthy` 필터를 거치지 않은 raw
`git diff` 카운트다. 게이트가 세는 값은 이보다 작거나 같으며, 두 앵커의 **대소 관계**가 이 절의 주장이다.)

### B.3 — 규칙 A의 프로브가 이 트리에서 발화한다

`git diff --name-only ad272be20 -- .moai/project/codemaps/ ':(exclude).moai/project/codemaps/provenance.json'`
는 비어 있지 않다 — 6개 본문 파일(`data-flow.md`, `dependencies.md`, `docs-truth.md`,
`entry-points.md`, `modules.md`, `overview.md`)을 출력한다. 그러나 이 트리에는 codemaps 본문에 대한
**미커밋 변경이 없다**. 차이는 S 이후 HEAD가 전진하면서 들어온 **커밋된** 본문 변경에서 온다. §D.2 참조.

### B.4 — 기존 픽스처의 codemaps 본문은 두 가지 모양이며, 둘 다 규칙 A/B를 통과하지 못한다

`internal/graph/check_test.go`에는 헬퍼가 **두 겹**이고 본문 처리가 서로 다르다.

| 헬퍼 | 본문 처리 | 픽스처가 갖는 상태 |
|---|---|---|
| `writeCodemapsProvenance` (바깥) | base 커밋 **이후** `.moai/project/codemaps/modules.md`를 `os.WriteFile`로 쓰고 커밋하지 않음 | 본문 존재, **미추적** |
| `writeCodemapsProvenanceBlock` (안쪽, 바깥 헬퍼가 감싸는 것) | 본문을 **아예 쓰지 않음** — `provenance.json`만 marshal해서 쓴다 | **본문 부재** |

안쪽 헬퍼를 직접 호출하는 픽스처(`check_absent_test.go` 등)는 codemaps 디렉터리에 `provenance.json`
하나만 갖는다. 두 모양 모두 정정 전 규칙 A(추적 diff 단독)로는 발화하지 않고, 규칙 B의
`git log -1 S -- <본문 pathspec>`도 비어 있다 — 그 픽스처의 어떤 커밋도 codemaps 디렉터리를 건드린 적이
없기 때문이다. 따라서 둘 다 규칙 C로 떨어진다.

미추적-본문 모양은 §D.1의 합집합 정정(REQ-GGR-004a)이 해결한다. **본문-부재 모양은 합집합으로도 해결되지
않는다** — 셀 본문 자체가 없다. 이것이 규칙 C를 C1/C2로 분리해야 하는 이유다(§D.1).

### B.5 — 레인 실측: 정정 전 구현에서 기존 테스트 2건이 같은 원인으로 실패한다

HEAD `2649fe296`의 구현에 대해 레인이 측정했다. 두 실패의 원인은 동일하다 — 규칙 C가
**시스템 오류**를 반환한다는 것이다.

```
$ go test ./internal/graph/ -run TestCheckFreshness_DescribedRootsScopeFidelity -count=1
--- FAIL: TestCheckFreshness_DescribedRootsScopeFidelity (0.52s)
    check_regression_lock_test.go:80: CheckFreshness: codemaps stamp d3fb238fc81b not comparable in this checkout: no commit in the stamped history touches the codemaps body

$ go test ./internal/cli/ -run TestGraphCheckCmd_AbsentExitsOne -count=1
--- FAIL: TestGraphCheckCmd_AbsentExitsOne (0.46s)
    graph_check_test.go:271: absent layers must exit 1, got graph check: codemaps stamp c369bcc3faed not comparable in this checkout: no commit in the stamped history touches the codemaps body
```

두 실패 중 어느 쪽도 **absent 판정이 틀렸다**고 말하지 않는다. 둘 다 **시스템 오류가 틀렸다**고 말한다.
`AbsentExitsOne`은 exit 1을 단언하는데 exit 2 오류 경로를 받았다. 이 구분 — 확정된 관측 대 실패한 측정 —
이 정확히 1/2 종료코드 계약이 이미 싣고 있는 구분이다.

## §C. 요구사항 (GEARS)

- **REQ-GGR-001** (Ubiquitous). codemaps 층은 `described-source-diff`의 측정 기점으로 스탬프된 커밋이
  아니라 **codemaps 본문이 마지막으로 실제 변경된 지점**을 사용해야 한다(shall).
- **REQ-GGR-002** (Ubiquitous). 본문 pathspec은 `.moai/project/codemaps/`에서
  `.moai/project/codemaps/provenance.json`을 제외한 것으로 정의된다(shall).
  git 형식: `-- .moai/project/codemaps/ ':(exclude).moai/project/codemaps/provenance.json'`.
- **REQ-GGR-003** (Where, capability gate). `Where` provenance가 clean 스탬프 경로일 때(`pv.Dirty` 거짓,
  `pv.CommitSHA` 비어 있지 않음), codemaps 층은 앵커 해석 절차(§D.1)를 수행해야 한다(shall).
- **REQ-GGR-004** (When, event-driven). `When` 작업 트리의 codemaps 본문이 S 시점의 본문과 동일하지
  않다고 관측되면 — **추적 차이(`git diff --name-only S -- <본문 pathspec>`)와 미추적 파일
  (`git ls-files --others --exclude-standard -- <본문 pathspec>`)의 합집합이 비어 있지 않으면** — 층은
  앵커를 스탬프된 sha S로 두고 `content_anchor_source=working-tree-differs-from-stamp`를 보고해야 한다(shall).
- **REQ-GGR-004a** (Ubiquitous). 규칙 A의 합집합 프로브는 `gitDiffNameList`가 described roots에 이미
  쓰는 관용구(diff + `ls-files --others`)를 재사용해야 한다(shall) — 두 번째 관용구를 만들지 않는다.
  추적 쪽만 보는 필터는 미추적 파일을 놓치며, 그 누락이야말로 이 지표가 막으려는 잡음이다.
- **REQ-GGR-005** (When, event-driven). `When` 규칙 A가 발화하지 않고 S의 이력 안에서 본문을 건드린
  커밋이 발견되면, 층은 그 커밋을 앵커로 삼고 `content_anchor_source=last-body-change`를 보고해야
  한다(shall).
- **REQ-GGR-006** (When, event-detected). `When` 앵커가 해석되지 않고 codemaps 본문 집합이 비어 있음이
  관측되면(C1), 층은 `verdict=absent`와 본문 부재를 명시하는 reason을 보고해야 하며, 시스템 오류를
  반환해서는 안 된다(shall not) — exit 1. 형제 층 `checkCitations`의 `docs == 0` 처분과 동일하다.
- **REQ-GGR-006a** (When, event-detected). `When` 본문은 존재하나 앵커가 해석되지 않으면(C2, 이력 절단),
  층은 `verdict=absent`를 보고하고 시스템 오류를 함께 반환해야 한다(shall) — exit 2. 기존 not-comparable
  경로와 동일한 처분이다.
- **REQ-GGR-007** (Unwanted). 층은 앵커를 해석할 수 없는 상태를 `fresh`로 보고해서는 안 된다(shall not).
  이는 C1·C2 양쪽에 구속된다 — **어떤 앵커 해석 결과도 결코 `fresh`를 낳지 않는다**는 것이 지켜야 할
  성질이고, C1은 그 성질의 도달 가능한 증인이다(AC-4).
- **REQ-GGR-008** (Ubiquitous). 앵커 walk는 S 자신의 이력으로 제한되어야 한다(shall). 그 제한이 앵커가
  S의 조상임을 보장하며, 값이 더 보수적인 방향(더 붉은 쪽)으로만 움직임을 보장한다(§B.2).
- **REQ-GGR-009** (Ubiquitous). `LayerReport`는 `content_anchor`(sha)와 `content_anchor_source`(§D.1의
  토큰 중 하나) 두 필드를 **보고 전용**으로 실어야 한다(shall). 두 필드는 게이팅에 쓰이지 않는다.
- **REQ-GGR-010** (Ubiquitous). 지표 토큰 `described-source-diff`는 유지되어야 한다(shall) — 정의만
  좁아진다. 이 토큰은 `.github/workflows/graph-freshness.yml`과 선행 SPEC들이 인용한다.
- **REQ-GGR-011** (Ubiquitous). `pv.Dirty` fingerprint 경로, mx-index 층, edges 층, citations 층의 동작은
  변경되지 않아야 한다(shall not change).
- **REQ-GGR-012** (Ubiquitous). 임계값 40은 이 작업으로 변경되지 않아야 한다(shall not change). §G 참조.

## §D. 판정 기록

### D.1 앵커 해석 절차 (채택)

clean 스탬프 경로에서만 동작한다. 스탬프된 sha를 S라 한다.

| 규칙 | 조건 | 앵커 | `content_anchor_source` |
|---|---|---|---|
| A | 작업 트리 본문이 S 시점 본문과 다름 — `git diff --name-only S -- <본문 pathspec>`과 `git ls-files --others --exclude-standard -- <본문 pathspec>`의 **합집합**이 비어 있지 않음 | S | `working-tree-differs-from-stamp` |
| B | A가 아니고 `git log -1 --format=%H S -- <본문 pathspec>`이 비어 있지 않음 | 그 커밋 | `last-body-change` |
| C1 | 그 외이고 **본문 집합이 비어 있음** — 작업 트리에도 S에도 본문이 없다(디렉터리에 `provenance.json`만) | 없음 | — (`verdict=absent`, **시스템 오류 없음** → exit 1) |
| C2 | 그 외이고 본문은 존재하나 앵커가 해석되지 않음(이력 절단) | 없음 | — (`verdict=absent` + **시스템 오류** → exit 2) |

규칙 A가 표현하는 의미: **작업 트리의 본문이 S 시점의 본문과 바이트 동일하지 않다면 본문은 최소한
S만큼은 새롭고, 따라서 S에서 재는 것이 옳으며 지나치게 관대할 수 없다.** 추적 차이만 보면 미추적
본문이 세어지지 않으므로, 합집합이 아니라 diff 단독이면 이 의미가 성립하지 않는다(§B.4).

합집합 형태는 `gitDiffNameList`가 described roots에 이미 쓰는 것과 **같은 모양**이며 이유도 같다 —
추적 쪽만 거르면 미추적 파일이 세어지지 않고, 그 누락이 바로 이 지표가 막으려는 잡음이다(REQ-GGR-004a).

규칙 B의 walk를 S 이력으로 제한하는 것이 하중을 받는 부분이다 — 그 제한이 앵커의 조상성을 보장하고,
따라서 이 변경이 게이트를 더 초록으로 만들 수 없음을 보장한다(§B.2 실측: 405 → 570, 조상성 rc=0).

**규칙 C는 두 갈래로 갈린다 — 확정된 관측(C1)과 실패한 측정(C2).** 종전 이 자리에는 "본문-부재는 선행
`codemaps directory missing` 분기가 이미 absent로 처분하므로 규칙 C는 사실상 도달 불가"라고 적혀 있었다.
**거짓이다.** 그 분기는 `os.Stat(dir); !info.IsDir()`(`check.go`)이므로 `provenance.json` 하나만 든
디렉터리는 통과한다. 본문-부재는 도달 가능하며, §B.4가 보여주듯 기존 픽스처가 실제로 그 상태다.

- **C1 — 본문 부재.** 기술하는 대상이 아예 없으므로 신선도는 판정 불가다. 그러나 이것은 **확정된 관측**이지
  실패한 측정이 아니다 → `verdict=absent`, 본문 부재를 명시하는 reason, **시스템 오류 없음**. exit 1.
- **C2 — 앵커 해석 불가.** 본문은 있는데 앵커가 해석되지 않는 경우(이력 절단) → `verdict=absent` +
  시스템 오류. exit 2. 배차 원안 그대로다.

**선례 — C1은 발명이 아니라 형제 층과의 정합이다.** 같은 파일군의 citations 층이 doc-less 상태를 이미
이렇게 처분한다(`internal/graph/check_citations.go` `checkCitations`, `docs == 0` 분기):

```go
	if docs == 0 {
		rep.Verdict = VerdictAbsent
		rep.Reason = "no codemaps documents to check"
		return rep
	}
```

absent이고 오류가 없다. C1은 codemaps 층을 citations 층과 **일치**시키며, C1이 없는 상태가 오히려 두 층의
분기(divergence)다.

### D.2 규칙 A 프로브의 오명명 — 라벨 개명으로 종결 (결정 완료)

`git diff S -- <paths>`는 S의 트리와 **작업 트리**를 비교하므로, S 이후에 본문을 바꾼 **커밋**이 있어도
발화한다. 즉 이 프로브가 실제로 재는 것은 "미커밋 재생성"이 아니라 "작업 트리가 스탬프와 다름"이다.
이 트리가 정확히 그 상태다(§B.3): 미커밋 본문 변경은 0인데 프로브는 본문 6개 파일을 출력한다.

**결정 (레인, 2026-09-04). 운영자 결정 사항이 아니며 재개하지 않는다.**

1. **규칙 A의 로직은 배차문 그대로 둔다.** 그 분기의 앵커는 S이고, S는 참된 본문 최종 변경 커밋의
   조상이거나 그 자신이므로 값은 더 보수적인 방향으로만 움직인다. 안전성 변화 없음.
2. **규칙 A의 `content_anchor_source` 토큰을 `working-tree-differs-from-stamp`로 개명한다.** 프로브가
   실제로 잰 것을 정직하게 기술하는 이름이다. 나머지 두 토큰은 그대로다 — 규칙 B는 `last-body-change`,
   규칙 C는 앵커를 아예 보고하지 않는다 — C1은 absent(오류 없음), C2는 absent + 시스템 오류.

개명의 근거는 위 측정이다. 이 SPEC이 `content_anchor` / `content_anchor_source` 두 필드를 새로 싣는
이유가 "앵커를 주장 대신 읽히게 한다"인 이상, `working-tree-regeneration`이라는 거짓 라벨은 그 목적에
정면으로 반한다. 로직 변경이 0이므로 개명의 비용은 없다.

채택하지 않은 대안: **규칙 A의 프로브를 미커밋 전용으로 좁히기**(`git diff` + `git diff --cached` +
`git ls-files --others`). 로직 변경이며 배차 범위 밖이고, 현행도 이미 안전하므로 안전성 이득이 없다.

### D.3 지표 토큰을 유지하고 정의만 좁히는 선택

`described-source-diff`를 새 토큰으로 갈아치우면 `.github/workflows/graph-freshness.yml`과 이 토큰을
인용한 선행 SPEC들이 일제히 스테일해진다. 정의만 좁히고 토큰을 유지하면, 무엇이 달라졌는지는 새로 실리는
`content_anchor` / `content_anchor_source` 두 필드가 말해 준다 — 이 편이 인용면을 흔들지 않으면서도
변화가 보고에 드러난다.

## §E. 범위 밖 (Out of Scope)

### Out of Scope — 스탬프 쓰기 경로
- `internal/cli/graph_stamp.go` 및 `mx.StampCodemaps`는 이 작업에서 변경하지 않는다. 스탬프가 본문을
  읽게 만드는 설계는 별도 사안이다.
- `internal/mx/provenance.go`의 provenance 스키마는 변경하지 않는다. 새 두 필드는 `LayerReport`(보고
  구조체)에만 실린다.

### Out of Scope — 임계값 재산정
- 임계값 40의 적정성은 §G의 운영자 결정 항목으로만 남긴다. 이 SPEC의 구현은 임계값을 바꾸지 않는다.

### Out of Scope — 다른 층과 dirty 경로
- `pv.Dirty` fingerprint 경로, mx-index / edges / citations 층은 손대지 않는다. AC-5가 이를 잠근다.

### Out of Scope — CI 워크플로와 codemaps 본문 정확도
- `.github/workflows/graph-freshness.yml`은 변경하지 않는다.
- codemaps 산문 자체가 코드를 정확히 기술하는가는 이 SPEC이 다루지 않는다. 이 SPEC은 "낡은 산문 위에서
  초록이 나오는 경로"만 닫는다.

## §F. 성공 기준

1. 본문을 그대로 둔 맨손 재스탬프 이후에도 codemaps 층이 stale로 판정된다(AC-1).
2. 진짜 재생성(미커밋/커밋 양쪽)은 여전히 fresh로 통과한다(AC-2).
3. 규칙 B가 고른 앵커가 스탬프 sha의 조상임이 기계적으로 확인된다(AC-3).
4. 측정 불가 상태가 absent + 시스템 오류로 남는다(AC-4).
5. 기존 `internal/graph` 테스트가 그대로 통과한다(AC-5).

## §G. 열린 질문 — 운영자 결정 사항

### Q1. 임계값 40이 이 저장소의 병합 리듬에 맞는가

**결정 주체: 운영자. 이 SPEC의 구현 범위 밖이며, 이 작업으로 임계값은 바뀌지 않는다(REQ-GGR-012).**

근거 자료: 2026-09-03 develop push는 커밋 167개를 한 번에 실었고 그때의 `described-source-diff`는
**144**로 임계값의 약 3.6배였다. 큰 통합마다 게이트가 발화한다면 그 신호는 신호가 아니라 잡음이다.

이 SPEC이 하는 일은 임계값 조정이 아니라 **측정 기점의 정정**이며, §B.2가 보여주듯 기점 정정은 값을
키우는 방향으로 작동한다. 따라서 Q1은 이 작업 이후 오히려 더 날카로워진다 — 운영자가 별도 카드로 판단할
사안이다.

### Q2. `content_anchor_source` 규칙 A 토큰 명명 — **종결됨 (열린 질문 아님)**

레인이 2026-09-04에 결정했다: 규칙 A의 토큰은 `working-tree-differs-from-stamp`. 로직은 불변.
근거와 측정은 §D.2. 이 항목은 운영자 판단 대상이 아니며, 기록으로만 남긴다.

## §I. 잔여 위험 (이 수리가 닫지 않는 것)

**본문에 사소한 미커밋 편집을 넣고 스탬프하면 규칙 A로 통과한다.** 유지자가 codemaps 본문에 어떤
사소한 미커밋 편집이라도 — 공백 문자 하나라도 — 넣은 뒤 스탬프를 찍으면, 규칙 A의 프로브가 발화해
앵커가 S가 되고 게이트는 초록이 된다.

이것은 설계의 결함이 아니라 **범위 밖**이다. 두 가지 이유로 그렇다.

- 이 SPEC이 닫는 것은 **가장 값싼 경로**다. 맨손 재스탬프는 아무 의도 없이도 밟게 되는 경로이므로
  닫는다. 위 우회는 게이트를 통과시키려는 **고의적 행위**이며, 고의를 막는 것은 게이트의 일이 아니다.
- 닫으려면 "이 편집이 실질적인가"를 판정하는 술어가 필요한데, 게이트는 그런 술어를 갖고 있지 않고
  이 작업으로 도입하지도 않는다.

## §H. 교차 참조

- `internal/graph/check.go` — `checkCodemaps`, `gitDiffNameList`, `LayerReport`
- `.github/workflows/graph-freshness.yml` — `described-source-diff` 인용면
- SPEC-GRAPH-FRESHNESS-CADENCE-001 — described-worthy 술어와 임계값 재산정 이력
- SPEC-STAMP-REACHABILITY-001 — `--commit` merge-base 앵커
- 카드 t475 — 오늘 이 위조-초록을 막고 있는 유일한 산문 주의

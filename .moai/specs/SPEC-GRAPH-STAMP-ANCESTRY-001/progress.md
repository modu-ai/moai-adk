# SPEC-GRAPH-STAMP-ANCESTRY-001 진행 기록

## §E.1 Plan-phase Audit-Ready Signal

- card: `t688`
- plan_status: `audit-ready`
- plan_complete_at: `2026-09-13T23:31:05+0900`
- worktree: `.claude/worktrees/t688`
- branch: `WT-graph-stamp-freshness`
- baseline_head: `30d505ac4`
- tier: `M`
- artifacts: `spec.md` (v0.2.0, `status: draft`) · `plan.md` · `acceptance.md` · `progress.md` — Tier M 산출물 3종 + 이 진행 기록이 모두 존재한다. 종전 이 절이 "`spec.md`만 생성됨"이라고 적고 있었으나 그 서술은 `b4626c042` 시점 이후로 거짓이었다.
- next_owner: `plan-auditor`
- next_action: 독립 plan-audit를 실행한다. Tier M PASS 임계는 `0.80`이며, `acceptance.md AC-GSA-011`이 네 렌즈 교차 모순 검사를 감사 항목으로 지정한다.

### 이번 보강에서 메운 것

산출물 네 개를 함께 읽고 교차 불일치만 겨냥해 고쳤다. 재작성은 하지 않았다.

- **라벨 충돌** — `spec.md §B.1`의 목표가 `M1~M3`이고 `plan.md §F`의 마일스톤이 `M1~M5`여서, `M3`이 두 문서에서 서로 다른 일을 가리켰다. 목표를 `G1~G3`으로 분리하고 두 축이 1:1이 아님을 명시했다. `REQ-GSA-011`도 마일스톤 이름을 걷어냈다 — 요구사항 층이 계획 층 식별자를 참조하고 있었다.
- **규범 표의 누락 행** — `spec.md §D` 판정 순서표에 본문 부재(C1) 행이 없었다. `REQ-GSA-003`이 막으려는 오독이 정확히 "`VerdictAbsent` 문자열만 보고 C1과 비조상을 합치는 것"인데, 그 구분이 규범 표에서는 검증할 수 없었다. C1 행과 `system error` 열을 추가했고, 종료코드만으로는 갈라지지 않는다는 사실(C1과 reachable stale이 둘 다 exit 1)을 적었다.
- **생성기 5문서 미명시** — `docs-truth.md` 제외가 세 문서에 걸쳐 있었는데 다섯이 무엇인지는 어디에도 없었다. `.claude/skills/moai/workflows/codemaps.md` § Output Files를 읽어 `overview`/`modules`/`dependencies`/`entry-points`/`data-flow`로 확정하고 `spec.md §B.1 G3`·`plan.md §A.3`·`§F M4`에 명시했다. `fold-judgments.txt`와 `provenance.json`도 생성기 밖임을 함께 적었다.
- **제외 범위의 소재** — warning-only 완화 제외가 이 진행 기록의 메모에만 있었다. 제외 범위는 SPEC이 소유해야 진행 기록과 함께 사라지지 않으므로 `spec.md §G`로 올렸다.
- **문서 수준 트리 핀** — `acceptance.md §A`가 단일 핀 `7097e6e21`을 제시했지만 `RED-GSA-008`은 `b4626c042`에서 측정됐다. 핀 우선순위(자기 핀이 이긴다)를 명시해 두 값의 차이가 드리프트로 읽히지 않게 했다.

### 보존된 두 결함 분리

이 SPEC의 핵심은 두 적색 원인을 서로 다른 실패 종류로 유지하는 것이며, 보강 과정에서 흐리지 않았다.

- **스탬프 도달 불가** — 저장된 clean `commit_sha`가 checkout `HEAD`의 조상이 아니다. 구조적/system error, exit 2, 숫자 freshness 없음.
- **freshness 값 초과** — 스탬프는 조상이지만 `described-source-diff`가 임계 40 이상이다. 진짜 stale이며, 재생성과 도달 가능한 재스탬핑이 필요하다. exit 1, 숫자와 귀속 유지.

warning-only 모드와 자동·맨손 재스탬프는 양쪽 모두에서 제외다.

### 확인된 계획 근거

- 읽기 전용 조사 4개 렌즈가 producer/consumer 경로, release merge 위상, 기존 SPEC 계약, 회귀 위험을 각각 조사했다. 합성 기록은 `plan.md §B`가 대신한다.
- 현재 기준에서 develop 스탬프 `f7b4919541f10b6415173b6bc9fb7192e8824450`은 `origin/develop`의 조상이지만 `origin/main`의 조상은 아니다.
- 기준 커밋 `7097e6e21`의 CI 근거는 `value=254`, `threshold=40`, `verdict=stale`, exit 1이다. 카드의 `value=224`는 같은 기준에서 재현되지 않았으므로 인수 기준의 고정값으로 쓰지 않는다.

### Evidence

이 보강 turn에서 실제로 실행한 명령과 원문 출력이다.

```text
command: moai spec lint SPEC-GRAPH-STAMP-ANCESTRY-001 --strict
tree_sha: 30d505ac4 (보강 편집 후, 커밋 전)
exit_code: 0
stdout:
0 error(s), 0 warning(s)
INFO  OwnershipTransitionUnmeasured  spec.md:1  transition "(none)" → "draft" expected owner
      "manager-spec" but commit b4626c0421d3b21564fcf7ba08967889af52a278 has no
      Authored-By-Agent trailer — ownership transition unmeasured
```

```text
command: moai spec audit --filter-spec SPEC-GRAPH-STAMP-ANCESTRY-001
tree_sha: 30d505ac4
exit_code: 0
stdout:
Total SPECs: 1 / Grandfathered: 0 / Modern-era clean: 1 / Drift findings: 0
No drift findings — all modern-era SPECs clean.
```

`OwnershipTransitionUnmeasured`는 INFO이며 `--strict`의 error/warning 집계에 들어가지 않는다. 이 보강 커밋은 `Authored-By-Agent: manager-spec` trailer를 담으므로, 다음 전이 측정부터는 귀속이 성립한다. 선행 커밋 `b4626c042`의 미측정 상태는 소급 수정하지 않는다.

- RED 원장 스크립트: `.moai/reports/t688/red-{ancestry,cli-unreachable,history-topologies}.sh`, `red-push-guard.py`
- 중단 기록: `.moai/reports/t688/gateway-502-recurrence.md`, `gateway-400-20260913T133738Z.md`, `gateway-400-20260913T141256Z.md`

### Gaps

이 보강 turn이 **관측하지 않은** 것.

- RED 원장 4건(`RED-GSA-001/002/006/008`)을 재실행하지 않았다. `acceptance.md §D`에 담긴 출력은 선행 turn의 관측이며, 이 turn은 그 기록의 존재만 확인했다.
- `BASE-GSA-001`(`moai graph check --root .`)을 재측정하지 않았다. `value=254`는 `7097e6e21`에 귀속된 값이고 현재 트리 `30d505ac4`에서 다시 재지 않았다.
- Go 테스트, `go vet`, `golangci-lint`를 실행하지 않았다. plan 단계이므로 코드 변경이 없다.
- `plan-auditor` 독립 감사를 실행하지 않았다 — 이것이 `next_action`이다.
- 두 gateway 중단(502 1건, 400 2건)의 하위 원인은 이 카드에서 진단하지 않았다. `t707`/`t708` 소관이다.
- push하지 않았다. 원격 반영은 리드의 일괄 소관이다.

### Residual risk

산출물은 자기 정합성이 올라갔지만 독립 감사를 통과한 적은 없다. 특히 `AC-GSA-011`의 네 렌즈 교차 모순 검사는 정의상 저자가 스스로 만족시켰다고 주장할 수 없는 항목이다 — `plan-auditor`가 각 행의 반증 조건을 원본 인용에 대조해야 성립한다. 또한 `value=254`가 현재 트리에서 재현되는지 확인하지 않았으므로, run 시작 시 `plan.md §C` 실행 전 점검이 이 값을 다시 재는 것이 전제다.

## §E.2 Run-phase Evidence

### §E.2.0 실행 전 재측정 (`plan.md §C`)

Implementation Kickoff Approval 뒤, 코드 편집 **전에** 워크트리 `.claude/worktrees/t688`, `HEAD 8275c82a5f706e4cc0830a6c093bc1d988358dd3`, 브랜치 `WT-graph-stamp-freshness`에서 여섯 항목을 다시 쟀다. plan 단계 수치는 `7097e6e21`/`b4626c042`에 귀속된 관측이며 아래가 run 시점 기준선이다.

**1. HEAD / branch**

```text
command: git rev-parse HEAD; git branch --show-current
exit_code: 0
stdout:
8275c82a5f706e4cc0830a6c093bc1d988358dd3
WT-graph-stamp-freshness
```

**2. tracked provenance 스탬프의 객체 존재와 조상성**

```text
command: python3 -c "... json.load('.moai/project/codemaps/provenance.json')"
stamp: f7b4919541f10b6415173b6bc9fb7192e8824450
dirty: false
described_roots: internal, cmd, pkg

command: git cat-file -e f7b4919541f10b6415173b6bc9fb7192e8824450^{commit}   → exit 0 (무출력)
command: git merge-base --is-ancestor f7b491954 HEAD                          → exit 0 (무출력)
command: git merge-base --is-ancestor f7b491954 origin/develop                → exit 0 (무출력)
command: git merge-base --is-ancestor f7b491954 origin/main                   → exit 1 (무출력)
```

plan `§A.2`의 관측과 같은 위상이다. 이 워크트리의 스탬프는 `HEAD`의 조상이므로, 이 트리 자체는 이 SPEC이 새로 닫는 비조상 경로에 해당하지 않는다 — 비조상 판정은 합성 fixture로만 검증한다.

**3. `moai graph check --root .` codemaps 행과 종료코드**

```text
command: moai graph check --root .
tree_sha: 8275c82a5
exit_code: 1
selected_stdout:
codemaps  metric=described-source-diff value=254 threshold=40 verdict=stale
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent  (mx-index absent (untracked runtime artifact — fresh worktree state))
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent  (edges.jsonl absent (untracked derived artifact — fresh worktree state))
citations metric=positive-cited-path-absence value=1 threshold=0 verdict=stale  (1 positively-cited path(s) absent from the tree (first cited in dependencies.md))
```

`value=254 threshold=40`은 `BASE-GSA-001`(`7097e6e21`)과 같은 값으로 이 트리에서도 재현됐다. 가정하지 않고 다시 쟀다. 다만 codemaps 외에 **citations도 `value=1 verdict=stale`**이며(`dependencies.md`가 인용한 `internal/cli/huh_theme.go`가 트리에 없다), mx-index와 edges는 워크트리에 untracked 산출물이 없어 `absent`다. plan `§A.2` 재시도 관측과 같은 구도이므로, M4 종결은 codemaps 행만으로 전체-green을 주장할 수 없다.

**4. 테스트 선택자가 실제로 테스트를 실행하는지**

RED 4건 재실행이 그 자체로 이 항목의 증거다(아래 5·`§E.2.1`). `internal/graph` 선택자는 1개 테스트와 3개 서브테스트를, `internal/cli` 선택자는 1개 테스트를 실제로 실행했다. `[no tests to run]`은 없었다.

**5. RED 원장 4건의 현재 트리 재현**

```text
command: bash .moai/reports/t688/red-ancestry.sh
tree_sha: 8275c82a5
exit_code: 1
stdout_stderr:
--- FAIL: TestT688ExistingNonAncestorStampIsUnmeasured (1.02s)
    t688_red_ancestry_test.go:24: existing non-ancestor stamp must be freshness-unmeasured with a system error; got verdict="fresh" value=1 threshold=40 content_anchor="d9029beaf2be3e6c70b6813e9715594ca8f5d085"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/graph	1.545s
classification: EXPECTED_RED (RED-GSA-001 재현)
```

```text
command: bash .moai/reports/t688/red-cli-unreachable.sh
tree_sha: 8275c82a5
exit_code: 1
stdout_stderr:
--- FAIL: TestT688CLIExistingNonAncestorStampExitsTwoWithRecovery (1.09s)
    t688_red_ancestry_test.go:38: existing non-ancestor stamp must exit 2; err=graph freshness check failed (stale or absent layer) stdout="codemaps  metric=described-source-diff value=1 threshold=40 verdict=fresh\n..." stderr="graph check: layer edges verdict=stale value=1 threshold=0 — source set(s) moved: codemaps\n"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/cli	1.953s
classification: EXPECTED_RED (RED-GSA-002 재현)
```

```text
command: bash .moai/reports/t688/red-history-topologies.sh
tree_sha: 8275c82a5
exit_code: 1
stdout_stderr:
--- FAIL: TestT688MergeSquashRebaseLikeTopologies (3.69s)
    --- FAIL: .../squash_retains_object_but_drops_ancestry (1.12s)
        t688_red_history_test.go:67: object-present non-ancestor stamp must be freshness-unmeasured; got verdict="fresh" value=0 anchor="7ecbb2b135fa62ee6fc9808c309810bede814a73"
    --- FAIL: .../rebase-like_rewrite_retains_object_but_drops_ancestry (1.30s)
        t688_red_history_test.go:78: object-present non-ancestor stamp must be freshness-unmeasured; got verdict="fresh" value=1 anchor="ec12109ed5c697a077a3e295a14f374acbf8197e"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/graph	4.206s
classification: EXPECTED_RED (RED-GSA-008 재현). merge 행은 통과했고 squash·rebase-like 두 행만 의도한 assertion에서 실패했다.
```

```text
command: python3 .moai/reports/t688/red-push-guard.py
tree_sha: 8275c82a5
exit_code: 1
stdout:
push_branch_has_target_head=false
push_branch_exits_before_ancestry=true
classification: EXPECTED_RED (RED-GSA-006 재현)
```

**6. workflow의 세 대상 선택과 관련 여섯 SPEC 상태**

`.github/workflows/graph-freshness.yml` reachability step 현재 상태: ordinary PR는 `TARGET="origin/${GITHUB_BASE_REF}"`, `release/*` head는 `TARGET="HEAD"`(merge preview), push는 `GITHUB_BASE_REF` 공백 분기에서 `echo "guard: push event (no base ref) — object presence verified only"` 뒤 **`exit 0`**으로 조상성 검사 전에 빠져나간다. 이것이 `REQ-GSA-008`이 닫는 틈이다.

```text
command: grep -m1 '^status:' .moai/specs/<SPEC>/spec.md  (6건)
SPEC-V3R6-GRAPH-FRESHNESS-001: status: completed
SPEC-V3R6-GRAPH-FRESHNESS-002: status: completed
SPEC-GRAPH-FRESHNESS-CADENCE-001: status: completed
SPEC-STAMP-REACHABILITY-001: status: completed
SPEC-GRAPH-GATE-RESTAMP-001: status: completed
SPEC-CODEMAPS-REFRESH-002: status: completed
```

### §E.2.1 마일스톤별 RED→GREEN

모든 명령은 워크트리 `.claude/worktrees/t688`, 브랜치 `WT-graph-stamp-freshness`에서 실행했다. 각 항목은 명령·원문 출력·종료코드·측정 트리를 함께 담는다.

**M1 — 조상성 선판정과 계보 fixture (커밋 `b6e84fe5c`)**

RED 원장 4건은 `§E.2.0`의 5번에 있다. 그 assertion들을 영구 테스트로 편입한 뒤 같은 트리에서 다시 재면:

```text
command: go test -count=1 -run 'TestCheckCodemaps_(ExistingNonAncestorStampIsUnmeasured|MergeSquashRebaseLikeTopologies|StampAtHeadIsComparable|BodyAbsentWinsOverNonAncestorStamp|UnresolvableStampKeepsNotComparableReason)' ./internal/graph
tree_sha: 8275c82a5 (구현 전)
exit_code: 1
stdout_stderr:
--- FAIL: TestCheckCodemaps_ExistingNonAncestorStampIsUnmeasured (0.98s)
    t688_ancestry_test.go:43: existing non-ancestor stamp must be freshness-unmeasured with a system error; got verdict="fresh" value=1 threshold=40 content_anchor="b239a6622d83c08b8239d5ab4068ec232c5a4ce6"
--- FAIL: TestCheckCodemaps_MergeSquashRebaseLikeTopologies (3.90s)
    --- FAIL: .../squash_retains_object_but_drops_ancestry (1.16s)
        t688_ancestry_test.go:81: object-present non-ancestor stamp must be freshness-unmeasured; got verdict="fresh" value=0 anchor="50b86adee8160f5146bddd5501e0063585305426"
    --- FAIL: .../rebase-like_rewrite_retains_object_but_drops_ancestry (1.44s)
        t688_ancestry_test.go:92: object-present non-ancestor stamp must be freshness-unmeasured; got verdict="fresh" value=1 anchor="0dc90966baa847a216aa4f578865de24cc56616d"
FAIL
FAIL	github.com/modu-ai/moai-adk/internal/graph	7.608s
classification: EXPECTED_RED — 2개 신규 assertion만 실패했고, 회귀 잠금 3건(StampAtHead / BodyAbsent / Unresolvable)은 구현 전에도 통과해 잠금 역할을 한다.
evidence_path: .moai/reports/t688/run/m1-red.log
```

```text
command: (같은 선택자, 구현 후) go test -count=1 -run '...' -v ./internal/graph
tree_sha: b6e84fe5c
exit_code: 0
stdout:
--- PASS: TestCheckCodemaps_ExistingNonAncestorStampIsUnmeasured (1.01s)
--- PASS: TestCheckCodemaps_MergeSquashRebaseLikeTopologies (3.77s)
    --- PASS: .../merge_commit_preserves_ancestry
    --- PASS: .../squash_retains_object_but_drops_ancestry
    --- PASS: .../rebase-like_rewrite_retains_object_but_drops_ancestry
--- PASS: TestCheckCodemaps_StampAtHeadIsComparable (0.76s)
--- PASS: TestCheckCodemaps_BodyAbsentWinsOverNonAncestorStamp (0.83s)
--- PASS: TestCheckCodemaps_UnresolvableStampKeepsNotComparableReason (0.50s)
ok  	github.com/modu-ai/moai-adk/internal/graph	7.481s
evidence_path: .moai/reports/t688/run/m1-green.log
```

**M2 — push 가드와 CLI 경계 (커밋 `c613b7c6b`)**

워크플로 가드는 텍스트 스캔이 아니라 **step의 셸 본문을 추출해 실제로 실행**한다. 세 이벤트 형태 각각에 대해, 잘못된 대상을 골랐다면 반대 판정이 나오는 fixture를 붙였다.

```text
command: go test -count=1 -run 'TestGraphFreshnessReachabilityGuard_TargetSelection' ./internal/graph
tree_sha: b6e84fe5c (워크플로 수정 전)
exit_code: 1
stdout:
--- FAIL: TestGraphFreshnessReachabilityGuard_TargetSelection (4.69s)
    --- FAIL: .../push_rejects_an_object-present_non-ancestor_stamp (0.76s)
classification: EXPECTED_RED — push 행 하나만 실패했다. 나머지 다섯 행(push 조상 수용 / ordinary PR base 판정 / release PR merge-preview / anchorless skip / 객체 부재 3형태)은 수정 전에도 통과해 회귀 잠금이며, 특히 release 행이 수정 전에 통과한다는 사실이 `set -euo pipefail` 아래 release 분기가 원래 도달 가능했음을 보인다.
evidence_path: .moai/reports/t688/run/m2-workflow-red.log
```

```text
command: python3 .moai/reports/t688/red-push-guard.py
tree_sha: c613b7c6b
exit_code: 0
stdout:
push_branch_has_target_head=true
push_branch_exits_before_ancestry=false
```

```text
command: go test -count=1 -run 'TestGraphFreshnessReachabilityGuard_TargetSelection' -v ./internal/graph
tree_sha: c613b7c6b
exit_code: 0
stdout:
--- PASS: TestGraphFreshnessReachabilityGuard_TargetSelection (4.49s)
    --- PASS: .../push_rejects_an_object-present_non-ancestor_stamp (0.76s)
    --- PASS: .../push_accepts_an_ancestor_stamp (0.57s)
    --- PASS: .../ordinary_PR_judges_base_ancestry,_not_HEAD (0.63s)
    --- PASS: .../release_PR_judges_the_merge-preview_HEAD (0.65s)
    --- PASS: .../anchorless_provenance_skips_with_a_reason (0.45s)
    --- PASS: .../missing_object_fails_on_every_event_shape (1.42s)
ok  	github.com/modu-ai/moai-adk/internal/graph	4.930s
evidence_path: .moai/reports/t688/run/m2-workflow-green.log
```

CLI 경계:

```text
command: go test -count=1 -run 'TestGraphCheckCmd_(ExistingNonAncestorStampExitsTwoWithRecovery|UnresolvableStampOmitsRegenerationAdvice)' -v ./internal/cli
tree_sha: b6e84fe5c (CLI 수정 전)
exit_code: 1
stdout_stderr:
    t688_graph_check_ancestry_test.go:48: stderr missing recovery token "unreachable": graph check: system error: codemaps stamp 710fa422f1a5 not comparable in this checkout: stamped commit is not an ancestor of HEAD
    t688_graph_check_ancestry_test.go:48: stderr missing recovery token "freshness unmeasured": (동상)
    t688_graph_check_ancestry_test.go:48: stderr missing recovery token "regenerate": (동상)
--- FAIL: TestGraphCheckCmd_ExistingNonAncestorStampExitsTwoWithRecovery (1.05s)
--- PASS: TestGraphCheckCmd_UnresolvableStampOmitsRegenerationAdvice (0.63s)
classification: EXPECTED_RED — exit 2는 이미 성립했고 복구 토큰 3종이 없었다. 두 번째 테스트가 수정 전에 통과하는 것은 mutant 방향(해석 불가 스탬프에 재생성 안내를 붙이는 구현)을 잠근다.
evidence_path: .moai/reports/t688/run/m2-cli-red.log
```

```text
command: (같은 선택자, 수정 후)
tree_sha: c613b7c6b
exit_code: 0
stdout:
--- PASS: TestGraphCheckCmd_ExistingNonAncestorStampExitsTwoWithRecovery (0.98s)
--- PASS: TestGraphCheckCmd_UnresolvableStampOmitsRegenerationAdvice (0.60s)
ok  	github.com/modu-ai/moai-adk/internal/cli	2.550s
evidence_path: .moai/reports/t688/run/m2-cli-green.log
```

**판정 순서의 수정 — 기존 계약이 이긴 지점.** 최초 구현은 `plan.md §D.1`의 문자 그대로 본문 존재(0)를 객체 해석(1)보다 앞에 두었고, 그 결과 `TestGraphCheckCmd_NotComparableExitsTwo`(본문 없음 + 해석 불가 스탬프)가 exit 2에서 exit 1로 바뀌어 실패했다.

```text
command: go test -count=1 -run 'Graph|graph' ./internal/cli
tree_sha: (M2 중간 상태)
exit_code: 1
stdout:
--- FAIL: TestGraphCheckCmd_NotComparableExitsTwo (0.43s)
    graph_check_test.go:208: unresolvable stamp must exit 2 (system error), got graph freshness check failed (stale or absent layer)
```

`spec.md §D` 1행이 C1을 "기존 계약 그대로, 이 SPEC이 바꾸지 않는다"로 규정하므로, 해석 불가(2행)가 본문 부재(1행)보다 앞선다는 **선행 순서를 보존**하는 쪽으로 고쳤다. 최종 순서는 해석 → 본문 존재 → 조상성이며, 이 SPEC이 새로 넣은 조상성 행에 대해서만 C1이 앞선다. 수정 후 같은 선택자는 exit 0이다(`m2-cli-graph-selector2.log`).

**M3 — 기존 freshness 계약 회귀 잠금**

```text
command: go test -count=1 -v -run '^(TestCheckCodemaps_BareRestampStaysStale|TestCheckCodemaps_CommittedRegenerationIsFresh|TestCheckCodemaps_UncommittedRegenerationIsFresh|TestCheckCodemaps_BodyAbsentIsAbsentWithoutError|TestCheckCodemaps_DirtyPathCarriesNoAnchor|TestCheckCodemaps_Contribution|TestCheckCodemaps_DrivingPaths|TestCheckCodemaps_Absent|TestCheckCodemaps_RuleBAnchorIsAncestorOfStamp|TestCheckCodemaps_UntrackedBodyResolvesViaRuleA|TestCheckFreshness_AllFresh|TestCheckFreshness_AgedLayerFails|TestCheckFreshness_ThresholdBoundaryOneBelowFresh|TestCheckFreshness_RevertedChurnCountsZero|TestCheckFreshness_DescribedRootsScopeFidelity|TestCheckFreshness_DirtyGenerationAnchor|TestCheckFreshness_NotComparableIsSystemError|TestCheckFreshness_NoProvenanceIsNotFresh)$' ./internal/graph
tree_sha: c613b7c6b
exit_code: 0
measured: `--- PASS` 18줄, `--- FAIL` 0줄, `no tests to run` 0회
tail: ok  github.com/modu-ai/moai-adk/internal/graph  22.430s
evidence_path: .moai/reports/t688/run/m3-named-regression.log
```

```text
command: go test -count=1 -v -run '^(TestGraphCheckCmd_.*|TestGraphStampCmd_.*)$' ./internal/cli
tree_sha: c613b7c6b
exit_code: 0
measured: `--- PASS` 18줄, `--- FAIL` 0줄, `no tests to run` 0회
tail: ok  github.com/modu-ai/moai-adk/internal/cli  13.050s
evidence_path: .moai/reports/t688/run/m3-cli-named.log
```

선택자가 지정한 이름 수와 실제 실행 수가 각각 18로 일치한다 — 0건 선택이 아니다.

**M4 — 진짜 재생성과 도달 가능한 스탬핑 (커밋 `eced8eb38`)**

재생성은 트리에서 다시 잰 사실만 반영했다. 확인한 드리프트:

| 대상 | 문서에 있던 값 | 트리에서 잰 값 |
|---|---|---|
| `github.com/charmbracelet/huh` (v1) | direct require, `internal/cli`가 사용 | `go.mod`에 없음, import 0건 |
| `github.com/google/uuid` | 표에 없음 | direct require v1.6.0 |
| `internal/cli/huh_theme.go` | v1·v2를 한 파일에서 import | 파일 자체가 없음 (`internal/cli/theme.go` 13줄 래퍼가 대체, huh import 0) |
| lipgloss v1 비테스트 사용처 | 언급 없음 | 6개 패키지 12개 파일 |
| bubbles / validator / runewidth / mvdan.cc-sh 버전 | v2.1.1 / v10.30.3 / v0.0.28 / v3.13.1 | v2.2.1 / v10.30.4 / v0.0.29 / v3.14.0 |
| `golang.org/x/net` 테스트 사용처 | 5곳 | 6곳 |
| `internal/graph` 비테스트 파일 | 15 | 14 |

`huh_theme.go`는 삭제된 경로이므로 blockquote(negative citation) 형태로 기록했다. 이것이 citations 층을 stale에서 fresh로 되돌린 변경이다.

```text
command: git diff --stat -- '.moai/project/codemaps/' ':(exclude).moai/project/codemaps/provenance.json'
tree_sha: c613b7c6b + 미커밋 재생성
exit_code: 0
stdout:
 .moai/project/codemaps/data-flow.md    | 41 ++++++++++++++++++++++++++++++++++
 .moai/project/codemaps/dependencies.md | 37 +++++++++++++++++++-----------
 .moai/project/codemaps/entry-points.md | 24 ++++++++++++++++++++
 .moai/project/codemaps/modules.md      |  3 ++-
 .moai/project/codemaps/overview.md     | 21 +++++++++++++++--
 5 files changed, 110 insertions(+), 16 deletions(-)
classification: 본문 변경 증거. bare restamp가 아니다 — provenance.json만 바뀐 변경이 아니라 생성기 5문서가 모두 바뀌었다.
evidence_path: .moai/reports/t688/run/m4-body-diff.log
```

Phase 4 fold/omission 판정 검사:

```text
command: (codemaps.md § Fold and Omission Judgment Check 의 awk 스크립트)
tree_sha: c613b7c6b + 재생성
exit_code: 0
stdout:
COLLECTED: fold=5 omission=15
classification: PASS — UNCOVERED / FOLD-PROSE / GAP 줄이 하나도 없다. 재생성이 fold 단위에 산문을 주지 않았고 omission 단위의 서술을 지우지 않았다.
evidence_path: .moai/reports/t688/run/m4-fold-check.log
```

스탬핑은 `moai graph stamp codemaps --commit "$(git merge-base HEAD origin/develop)"` — 명령 자신의 권고 레시피(branch-local HEAD 금지, squash 생존)를 따랐다. 결과 스탬프는 `7097e6e214195c45e65cab5fa565b19ca4514c4e`.

**종결 관측 — 서로 독립된 두 명령, 같은 트리**

```text
tree_sha: eced8eb38bec75eef147fc145c49e53499f8d9c5
tree_object: a9ad8e24666f2fde6c6cc6d297bf409155c6ba11
judging_build: eced8eb38 (go build -o ./bin/moai ./cmd/moai — 이 트리에서 빌드한 바이너리로 판정. 설치본 v3.2.0-rc.10(commit 7a7a08f20)은 이 변경 이전 빌드라 사용하지 않았다)
stamp: 7097e6e214195c45e65cab5fa565b19ca4514c4e

관측 1 — 조상성 (git 단독, moai 미개입)
command: git merge-base --is-ancestor 7097e6e214195c45e65cab5fa565b19ca4514c4e HEAD
exit_code: 0 (무출력)
command: git cat-file -e 7097e6e214195c45e65cab5fa565b19ca4514c4e^{commit}
exit_code: 0 (무출력)

관측 2 — freshness (이 트리에서 빌드한 moai)
command: ./bin/moai graph check --root .
stdout:
codemaps  metric=described-source-diff value=2 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=absent  (mx-index absent (untracked runtime artifact — fresh worktree state))
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=absent  (edges.jsonl absent (untracked derived artifact — fresh worktree state))
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
exit_code: 1  (codemaps는 fresh; mx-index·edges의 absent 때문)
evidence_path: .moai/reports/t688/run/m4-closure.log
```

두 관측은 서로 독립이다 — 조상성은 git만 쓰고 freshness 값을 읽지 않으며, freshness는 값만 읽고 조상성을 보고하지 않는다. `value=2 < 40`이 `spec.md §F` 4번의 수치 조건이고, `exit 0` 조상성이 그 짝이다. anchor는 `working-tree-differs-from-stamp`로 스탬프 자신이며(`--json` 확인), 값 2는 merge-base 이후 바뀐 described-worthy 비테스트 파일 2개(`internal/graph/check.go`, `internal/cli/graph_check.go`)다. 나머지 3개는 `_test.go`라 described-worthy 필터가 제외한다.

**mx-index·edges의 absent 처분과 CI 등가 부트스트랩.** 두 층은 워크트리에 없는 untracked 런타임 산출물이라 fresh-worktree 상태에서 absent다. 워크플로가 check 전에 수행하는 부트스트랩(`moai mx scan --quiet` + `moai graph build`)을 같은 트리에서 실행하면 네 층이 모두 fresh이고 종료코드가 0이 된다.

```text
command: ./bin/moai graph build && ./bin/moai graph check --root .
tree_sha: eced8eb38 (부트스트랩 산출물은 untracked, 커밋 대상 아님)
exit_code: 0
stdout:
codemaps  metric=described-source-diff value=2 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=fresh
citations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh
evidence_path: .moai/reports/t688/run/m4-closure-bootstrapped.log
```

한 가지 관측 순서 주의: 첫 시도에서 edges가 `value=1 stale (source set(s) moved: reports)`였다. `graph build` 이후 내가 `.moai/reports/` 아래 증거 로그를 쓴 것이 fingerprint를 움직였기 때문이며, 로그를 쓴 뒤 build를 다시 돌리고 check 하면 위와 같이 fresh다. 코드 결함이 아니라 측정 순서의 산물이므로 기록해 둔다.

**M5 — 범위 검증**

영향 패키지 전량:

```text
command: go test -count=1 ./internal/graph/...
tree_sha: eced8eb38
exit_code: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/graph	43.530s
ok  	github.com/modu-ai/moai-adk/internal/graph/symbol	0.876s
evidence_path: .moai/reports/t688/run/m2-graph-pkg2.log
```

```text
command: go test -count=1 ./internal/cli/...
tree_sha: eced8eb38
exit_code: 1
관측: 루트 `internal/cli` 패키지가 `panic: test timed out after 10m0s`로 끊겼다. `--- FAIL` 0건이며 하위 패키지 18개는 전부 `ok`다. 이 패키지는 go test 기본 10분 타임아웃보다 오래 걸린다 — 실패가 아니라 시간 초과다.
evidence_path: .moai/reports/t688/run/m5-cli-full.log
```

```text
command: go test -count=1 -timeout 45m ./internal/cli
tree_sha: eced8eb38
exit_code: 0
stdout:
ok  	github.com/modu-ai/moai-adk/internal/cli	1480.218s
관측: `--- FAIL` 0건. 위 타임아웃이 시간 문제였음을 같은 트리에서 확인했다.
evidence_path: .moai/reports/t688/run/m5-cli-root-45m.log
```

```text
command: ./bin/moai spec lint SPEC-GRAPH-STAMP-ANCESTRY-001 --strict
tree_sha: eced8eb38
exit_code: 0
stdout:
✓ No findings — all SPEC documents are valid
evidence_path: .moai/reports/t688/run/m5-spec-lint.log
```

```text
command: ./bin/moai spec audit --filter-spec SPEC-GRAPH-STAMP-ANCESTRY-001
tree_sha: eced8eb38
exit_code: 0
stdout:
Total SPECs: 1 / Grandfathered: 0 / Modern-era clean: 1 / Drift findings: 0
No drift findings — all modern-era SPECs clean.
evidence_path: .moai/reports/t688/run/m5-spec-audit.log
```

```text
command: go vet ./internal/graph/... ./internal/cli/...
tree_sha: eced8eb38
exit_code: 0
stdout: (무출력)
evidence_path: .moai/reports/t688/run/m5-vet.log
```

```text
command: golangci-lint run ./internal/graph/... ./internal/cli/...
tree_sha: eced8eb38
exit_code: 1
summary: 37 issues (errcheck 36, staticcheck 1)
attribution: 이 카드가 바꾼 파일과의 교집합 0. 지적된 파일은 gateway_*.go / gpt_auth.go / migrate_cg.go 10개이며, 이 카드의 변경 파일 목록(`m5-changed-files.log`, 43개)에 하나도 들어 있지 않다.
command: comm -12 <(git diff --name-only "$(git merge-base HEAD origin/develop)" HEAD | sort) <(lint 출력에서 뽑은 파일 목록) | wc -l  →  0
evidence_path: .moai/reports/t688/run/m5-lint.log, m5-changed-files.log
```

범위: 변경 파일 43개 전부가 이 SPEC의 선언 범위 안이다 — 워크플로 1, codemaps 6, SPEC 산출물 4, 증거 파일 24, 구현·테스트 5(추가 3, 수정 2), 그리고 plan 단계가 남긴 파일들. `.moai/state/**` 변경 0건, 다른 SPEC 디렉터리 변경 0건.

### §E.2.2 AC PASS/FAIL 매트릭스

| AC | 분류 | 검증 명령 | 실제 출력 | 상태 |
|---|---|---|---|---|
| AC-GSA-001 | release-blocking | `go test -run TestCheckCodemaps_ExistingNonAncestorStampIsUnmeasured ./internal/graph` | `--- PASS` (1.01s) | PASS |
| AC-GSA-002 | release-blocking | `go test -run TestGraphCheckCmd_ExistingNonAncestorStampExitsTwoWithRecovery ./internal/cli` | `--- PASS` (0.98s); exit 2 + 토큰 4종 + 숫자 행 부재 | PASS |
| AC-GSA-003 | release-blocking | 위 두 테스트의 `assertUnmeasured` / stdout·stderr 단언 | `--- PASS`; Value·ContentAnchor·ContentAnchorSource·Contribution·DrivingPaths·DrivingPathsOmitted 전부 비어 있음 | PASS |
| AC-GSA-004 | regression guard | `go test -run '^(TestCheckFreshness_AgedLayerFails\|TestCheckCodemaps_DrivingPaths\|TestCheckCodemaps_Contribution)$' ./internal/graph` (M3 묶음에 포함) | `--- PASS` 3건 | PASS |
| AC-GSA-005 | regression guard | `TestCheckFreshness_AllFresh`, `TestCheckFreshness_ThresholdBoundaryOneBelowFresh`, `TestCheckCodemaps_StampAtHeadIsComparable` | `--- PASS` 3건 | PASS |
| AC-GSA-006 | release-blocking | `go test -run TestGraphFreshnessReachabilityGuard_TargetSelection ./internal/graph` + `python3 red-push-guard.py` | 서브테스트 6/6 PASS; probe `has_target_head=true`, `exits_before_ancestry=false`, exit 0 | PASS |
| AC-GSA-007 | regression guard | 같은 테스트의 `ordinary PR judges base ancestry` / `release PR judges the merge-preview HEAD` 두 행 | 둘 다 PASS; 같은 fixture에서 head ref만 바꿔 대상 선택이 갈리는 것을 보임 | PASS |
| AC-GSA-008 | release-blocking | `go test -run TestCheckCodemaps_MergeSquashRebaseLikeTopologies ./internal/graph` | 3개 서브테스트 PASS; 세 계보 모두 객체 존재, merge만 조상 | PASS |
| AC-GSA-009 | release-blocking closure | `git diff --stat`(본문) + `git merge-base --is-ancestor`(exit 0) + `./bin/moai graph check`(value=2 fresh) | 본문 5문서 110+/16− 변경; 조상성 exit 0; `value=2 threshold=40 verdict=fresh` | PASS |
| AC-GSA-010 | regression guard | M3의 두 이름 지정 묶음 | `--- PASS` 18 + 18, FAIL 0, `no tests to run` 0 | PASS |
| AC-GSA-011 | plan-audit gate | plan-auditor 독립 감사 (run 단계 소관 아님) | plan-audit PASS 0.974 — `.moai/reports/plan-audit/SPEC-GRAPH-STAMP-ANCESTRY-001-plan-audit.md` | PASS (plan 단계) |
| AC-GSA-012 | scope/quality closure | `go vet`(exit 0), `golangci-lint`(신규 0), 변경 파일 범위 검사 | vet 무출력; lint 37건 전부 카드 밖 파일; `.moai/state/**` 변경 0 | PASS |

새 dependency·verdict enum·provenance 필드·임계값·predicate 변경은 없다. 자동 재스탬프도 warning-only 완화도 추가하지 않았다.

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: "2026-09-14T01:05:00+0900"
run_commit_sha: "pending-backfill-run"   # M5 증거 커밋 SHA — 이 커밋 자신의 해시라 착지 후 backfill
run_status: PASS
ac_pass_count: 12
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: "git fetch 미실행 — 워크트리 로컬 작업이며 push하지 않는다. 원격 대조는 리드의 일괄 push 소관"
l44_post_push_fetch: "해당 없음 — 이 카드는 push하지 않는다"
new_warnings_or_lints_introduced: 0
cross_platform_build:
  darwin_arm64: "go build ./internal/... exit 0"
  windows_amd64: "미실행 — Gaps 참조"
total_run_phase_files: 5   # 구현·테스트 (추가 3, 수정 2). codemaps 6 + SPEC 2 + 증거 24는 별도
m1_to_mN_commit_strategy: "마일스톤당 1커밋 — M1 b6e84fe5c, M2 c613b7c6b, M4 eced8eb38, M5(이 커밋). M3는 검증 전용이라 코드 변경이 없어 증거를 M4 커밋에 실었다"
```

- card: `t688`
- worktree: `.claude/worktrees/t688`
- branch: `WT-graph-stamp-freshness`
- 미푸시 커밋: 리드 보고 시점 기준 아래 § 완료 보고 참조
- next_owner: `manager-docs` (sync-phase) — 리드 배차 뒤
- next_action: sync는 병합 **전에** 이 워크트리 안에서 끝낸다(run만 닫고 병합하면 SPEC이 `in-progress`로 develop에 올라간다)

### Gaps — 이 run 단계가 관측하지 않은 것

- **windows/amd64 크로스 빌드를 실행하지 않았다.** `go build ./internal/...`(darwin/arm64) exit 0만 관측했다. 이 카드의 Go 변경은 `os/exec`·`errors`·`fmt`만 쓰고 syscall이나 빌드 태그를 건드리지 않지만, 그것은 추론이지 측정이 아니다. 판정은 CI 매트릭스 몫이다.
- **워크플로 가드 테스트는 `bash`·`jq`·`git`을 요구한다.** 셋 중 하나라도 없으면 `t.Fatalf`로 실패한다(skip이 아니다 — 건너뛴 초록은 공허하다). 로컬에서는 셋 다 있었고(`jq-1.8.1`, GNU bash 3.2.57), 우분투 러너에도 있지만 **CI에서 실제로 실행되는 것은 아직 관측하지 않았다**.
- **실제 GitHub 이벤트로 가드를 검증하지 않았다.** 세 이벤트 형태는 합성 로컬 이력에서 step 셸 본문을 실행해 재현했다. 실제 push·PR·release PR 실행은 `spec.md §G`의 제외 범위다.
- **`golangci-lint`의 37건을 merge-base에서 다시 재지 않았다.** 카드 변경 파일과의 교집합이 0이라는 것은 측정했지만, 그 37건이 merge-base에서도 같은 수였는지는 재지 않았다. 귀속 근거는 "이 카드가 그 파일들을 건드리지 않았다"이지 "이전에도 있었다"가 아니다.
- **`moai graph check`의 부트스트랩 초록은 untracked 산출물에 의존한다.** `mx scan`·`graph build`가 만든 파일은 커밋 대상이 아니므로, 다른 트리에서 그대로 재현하려면 같은 부트스트랩을 먼저 돌려야 한다.
- **push·PR·머지를 수행하지 않았다.** 원격 반영과 develop 병합은 리드 소관이다.
- **issue #1661에 아무것도 쓰지 않았다.** `spec.md §G` 제외 범위다.

### Residual risk

- **[측정됨] 통합 뒤 codemaps는 다시 stale이 된다 — 그리고 이것은 이 카드가 고칠 수 있는 것이 아니다.** 스탬프 `7097e6e21`은 `origin/develop`의 조상이므로(exit 0) 조상성은 병합 뒤에도 성립한다. 그러나 값은 다르다:

  ```text
  command: git diff --name-only 7097e6e214195c45e65cab5fa565b19ca4514c4e origin/develop -- internal cmd pkg | grep -v '_test\.go$' | wc -l
  tree_sha: dfe04c9b8
  exit_code: 0
  stdout: 58
  command: git rev-list --count --left-right origin/develop...HEAD
  stdout: 112	9
  ```

  `origin/develop`이 fork 지점 이후 112커밋 앞서 있고 그 사이 described-worthy 비테스트 파일 58개가 바뀌었다. 이 브랜치가 develop을 흡수하면 `described-source-diff`는 약 60이 되어 임계 40을 다시 넘는다. **이 브랜치에서 도달 가능한 어떤 커밋에 스탬핑해도 마찬가지다** — 그 58개 변경은 이 브랜치의 이력 밖에 있다.

  이것은 이 SPEC이 만든 결함이 아니라 이 SPEC이 이름 붙인 두 번째 적색 원인(freshness 값 초과)이며, 그 복구 절차도 이 SPEC이 규정한 그대로다. 통합 창에서 레인이 develop을 흡수한 뒤 **병합 트리에서 재측정**하면 값이 40을 넘는 것이 드러날 것이고, 그때 흡수된 develop tip으로 재스탬핑하면 anchor가 그리로 옮겨가 값이 다시 2로 떨어진다(본문 재생성은 이 카드에서 이미 진짜로 수행했으므로 bare restamp가 아니다). 이 카드는 병합하지 않으므로 그 단계를 실행하지 않았고, 리드 판정 사항으로 남긴다.
- **codemaps 부분 재측정의 경계.** 다섯 문서 중 이번에 다시 잰 것은 각 문서 상단의 `**부분 재측정**` 줄이 명시한 절뿐이다. 나머지 서술은 `e7bd89ee3` 시점 값이며, 트리에 문서가 이름조차 대지 않는 패키지가 47개 남아 있다(`internal/core/git`은 의도된 fold라 제외). 그 공백은 codemaps 전면 리프레시 소관이지 이 카드의 범위가 아니다.
- **조상성 실패를 fail-closed로 처리한다.** `merge-base --is-ancestor`가 비조상 이외의 이유(읽을 수 없는 이력 등)로 non-zero를 내도 unreachable로 분류한다. 안전한 방향이지만, 그런 상태에서는 reason이 원인을 정확히 지목하지 못한다.
- **`internal/cli` 루트 패키지가 기본 타임아웃을 넘는다(1480s).** 이 카드가 만든 상태는 아니지만, 기본값으로 도는 어떤 소비자든 이 패키지에서 timeout panic을 본다.

## §E.4 Sync-phase Audit-Ready Signal

```yaml
sync_complete_at: "2026-09-14T01:20:00+0900"
sync_commit_sha: 5ea0bac5fe43e7feb4bb872f675bf69f6e9b3714
sync_status: PASS
b12_self_test_a: "grep -c 'SPEC-GRAPH-STAMP-ANCESTRY-001' CHANGELOG.md → 0 (삽입 전). 중복 없음"
b12_self_test_b: "acceptance.md AC 고유 식별자 12개 (AC-GSA-001~012) = CHANGELOG 항목이 적은 12개. 일치"
b12_self_test_c: "CHANGELOG가 이름 대는 경로 7개 전부 ls 확인 — MISSING 0"
changelog_entry_position: "[Unreleased] › Changed 선두"
frontmatter_status_transitions:
  spec_md: "in-progress → implemented (updated: 2026-09-13 → 2026-09-14)"
  plan_md: "해당 없음 — frontmatter 블록이 없다"
  acceptance_md: "해당 없음 — frontmatter 블록이 없다"
  progress_md: "해당 없음 — frontmatter 블록이 없다"
canary_compliance_check:
  applicable: false
  reason: "이 SPEC은 자기 sync가 시험하는 전방적 정책을 정의하지 않는다"
codemaps_provenance_ancestry:
  command: "git merge-base --is-ancestor 7097e6e214195c45e65cab5fa565b19ca4514c4e HEAD"
  exit_code: 0
  note: "M4에서 이미 재생성했으므로 재스탬프하지 않았다"
spec_lint: "moai spec lint SPEC-GRAPH-STAMP-ANCESTRY-001 --strict → exit 0, '✓ No findings'"
spec_audit: "mcp spec_audit → total_specs=1, modern_era_clean=1, drift_findings=[]"
```

### 문서 동기화 판정

- **README — 변경 없음.** `moai graph check`의 종료 코드나 graph freshness 워크플로를 설명하는 절이 README 4개 로케일 어느 쪽에도 없다.
- **docs-site — 변경 없음, 단 한 줄을 간극으로 남긴다.** `docs-site/content/{ko,en,ja,zh}/cli-reference/graph.md`가 종료 코드 0/1/2와 CI 가드를 설명한다. 종료 코드 서술(약 70행)은 이 변경 뒤에도 그대로 참이다 — 도달 불가 스탬프는 시스템 오류(2)로 닫힌다. CI 가드 서술(약 98행)은 가드가 "풀 리퀘스트의 베이스 브랜치"에 대해 조상성을 검사한다고만 적는다. PR 경로에 대해서는 여전히 참이지만, 이제 push 이벤트도 `HEAD` 조상성을 판정하므로 **범위를 좁게 적고 있다**. 거짓 진술이 아니라 새 문장을 더해야 메워지는 부족이므로, 이번 sync 범위("틀린 것만 고친다")에서는 고치지 않고 후속 카드 후보로 남긴다. 고치려면 4개 로케일을 같은 변경에서 함께 손봐야 한다.
- **codemaps — 재생성하지 않았다.** M4에서 이미 수행됐고, 다시 찍으면 값이 움직인다. 조상성만 확인했다(위 yaml).

### Gaps — 이 sync 단계가 관측하지 않은 것

- 테스트를 다시 돌리지 않았다. AC 12건은 전부 §E.2.2의 run 단계 측정을 귀속시킨 것이며, 이 커밋은 문서만 바꾸므로 재측정 근거가 아니다.
- `go test ./...`을 돌리지 않았다(배차 제약).
- hugo 빌드를 돌리지 않았다 — docs-site를 건드리지 않았기 때문이다.
- push도 PR도 하지 않았다. 원격 반영과 develop 병합은 리드 소관이다.
- issue #1661에 아무것도 쓰지 않았다.

### Residual risk

- `sync_commit_sha`가 플레이스홀더로 남는다. backfill을 빠뜨리면 이 신호는 자기 커밋을 가리키지 못한다.
- docs-site 약 98행의 범위 부족을 고치지 않고 남겼다. push 실행에서 exit 2를 맞은 독자가 그 문서를 읽으면 가드가 PR 전용이라고 오해할 수 있다.
- run 단계가 기록한 잔여 위험(통합 뒤 codemaps가 다시 stale이 된다, windows 빌드 미측정, 워크플로 가드의 실제 이벤트 미검증)은 이 sync가 해소하지 않았다. §E.3 Residual risk가 그대로 유효하다.
mx_commit_sha: (this commit)

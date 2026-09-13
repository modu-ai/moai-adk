# SPEC-GRAPH-STAMP-ANCESTRY-001 인수 기준

## §A. 적용 범위와 판정 규칙

- Tier: M
- 요구사항: REQ-GSA-001~REQ-GSA-012
- 기준 RED 트리: `7097e6e214195c45e65cab5fa565b19ca4514c4e`
- release-blocking 기준은 아래 RED 원장에 명령·원문 출력·종료코드·트리 SHA가 모두 있는 항목뿐이다.
- 현재 트리에서 다시 실행할 수 없는 과거 수치나 사건은 regression guard로만 분류하며 PASS로 기록하지 않는다.
- selector가 0개 테스트를 실행하거나 `[no tests to run]`을 출력하면 해당 AC는 FAIL이다.

## §B. 요구사항 추적표

| AC | 추적 REQ | 분류 | 완료를 뒤집는 마일스톤 |
|---|---|---|---|
| AC-GSA-001 | 001, 002 | release-blocking | M1 |
| AC-GSA-002 | 002, 003, 004, 005 | release-blocking | M2 |
| AC-GSA-003 | 003, 004 | release-blocking | M1/M2 |
| AC-GSA-004 | 006 | regression guard | M3 |
| AC-GSA-005 | 007 | regression guard | M3 |
| AC-GSA-006 | 008 | release-blocking | M2 |
| AC-GSA-007 | 009 | regression guard | M2 |
| AC-GSA-008 | 001, 002, 010 | release-blocking | M1 |
| AC-GSA-009 | 005, 011 | release-blocking closure | M4 |
| AC-GSA-010 | 012 | regression guard | M3 |
| AC-GSA-011 | 001~012 | plan-audit contradiction gate | plan-auditor |
| AC-GSA-012 | 011, 012 | scope/quality closure | M5 |

## §C. Given-When-Then 시나리오

### AC-GSA-001 — 객체는 존재하지만 비조상인 checker 결과

**Given** clean `commit_sha` 객체가 fixture 저장소에는 존재하지만 현재 checkout `HEAD`의 조상이 아니고 codemaps 본문도 존재한다.  
**When** codemaps checker가 freshness를 평가한다.  
**Then** report는 호환용 `VerdictAbsent`를 운반하고, reason은 `unreachable stamp`와 `freshness unmeasured`를 구분 가능하게 담으며, non-nil system error를 반환한다. 숫자 freshness 계산은 실행되지 않는다.

- RED-now: `RED-GSA-001`
- GREEN path: M1이 ancestry precheck와 영구 단위 테스트를 추가한 뒤 같은 명령은 테스트 1개를 실행해 `ok .../internal/graph`와 exit 0을 낸다.
- mutant probe: `git cat-file`만 성공하면 freshness 계산을 계속하는 구현은 RED fixture에서 error가 nil이므로 실패해야 한다.

### AC-GSA-002 — CLI exit 2와 복구 안내

**Given** AC-GSA-001과 같은 비조상 스탬프가 CLI fixture에 있다.  
**When** `graph check` command가 실행된다.  
**Then** CLI는 exit 2를 반환하고 stderr에 `unreachable`, `freshness unmeasured`, `regenerate`, `stamp` 의미를 모두 담으며, stdout에 `metric=described-source-diff value=` 숫자 row를 출력하지 않는다.

- RED-now: `RED-GSA-002`
- GREEN path: M2가 CLI 경계와 영구 테스트를 고친 뒤 같은 명령은 테스트 1개를 실행해 `ok .../internal/cli`와 exit 0을 낸다.
- mutant probe: exit 2만 반환하고 genuine regeneration 안내 또는 reachable stamp 안내 중 하나를 빼는 구현은 토큰 검증에서 실패해야 한다.

### AC-GSA-003 — 미측정 필드 비생성

**Given** checker가 비조상 clean stamp를 감지했다.  
**When** report와 CLI text/JSON surface를 읽는다.  
**Then** `Value`, threshold 비교 결과, `ContentAnchor`, `ContentAnchorSource`, `Contribution`, `ContributionBase`, `DrivingPaths`, `DrivingPathsOmitted`는 측정 결과로 제시되지 않고, `VerdictAbsent`는 reason + non-nil error와 함께 호환 운반체로만 해석된다.

- RED-now: `RED-GSA-001`은 현재 `verdict="fresh" value=1 ... content_anchor=<sha>`를 관측해 이 기준을 정당한 이유로 실패한다.
- GREEN path: M1 checker 테스트와 M2 CLI JSON/text 테스트가 필드 부재를 각각 검증한다.
- mutant probe: verdict만 absent로 바꾸고 value/anchor를 남기는 구현은 필드 assertion에서 실패해야 한다.

### AC-GSA-004 — reachable stale 숫자 계약

**Given** clean stamp가 현재 `HEAD`의 조상이고 `described-source-diff`가 threshold 40 이상이다.  
**When** checker와 CLI가 실행된다.  
**Then** codemaps row는 숫자 value, threshold, content-anchor, contribution/driving-path 귀속과 `stale`을 보고하고 CLI는 exit 1을 반환한다.

- 현재 재현 관측: `BASE-GSA-001`의 `value=254 threshold=40 verdict=stale`, exit 1.
- GREEN path: M3가 `TestCheckFreshness_AgedLayerFails`, stale attribution tests와 신규 ancestry tests를 같은 영향 범위에서 실행한다.
- mutant probe: 모든 ancestry 성공을 fresh로 단락하는 구현은 threshold fixture에서 실패해야 한다.

### AC-GSA-005 — reachable fresh 숫자 계약

**Given** clean stamp가 현재 `HEAD`의 조상이고 `described-source-diff`가 40 미만이며 다른 층도 fresh다.  
**When** checker와 CLI가 실행된다.  
**Then** codemaps row는 숫자 value와 `fresh`를 보고하고 전체 CLI는 exit 0을 반환한다.

- 분류: 기존 동작 regression guard. 현재 이 plan 단계에서는 새 PASS를 주장하지 않는다.
- GREEN path: M3가 `TestCheckFreshness_AllFresh`, `TestGraphCheckCmd_AllFreshExitZero`, one-below threshold test를 실제 실행하고 테스트 수와 원문 출력을 기록한다.
- mutant probe: ancestry 검사가 자기 자신인 stamp도 거부하는 구현은 fresh fixture에서 실패해야 한다.

### AC-GSA-006 — push `HEAD` ancestry guard

**Given** workflow가 push event를 처리하고 provenance에 clean `commit_sha`가 있으며 객체는 다른 ref를 통해 checkout에 남아 있지만 push `HEAD`의 조상은 아니다.  
**When** reachability step의 push 분기를 평가한다.  
**Then** push target은 `HEAD`이고 ancestry 검사 전에 성공 종료하지 않으며, 비조상은 job failure로 도달한다.

- RED-now: `RED-GSA-006`
- GREEN path: M2가 workflow를 고친 뒤 같은 probe가 `push_branch_has_target_head=true`, `push_branch_exits_before_ancestry=false`, exit 0을 낸다.
- mutant probe: `TARGET="HEAD"`를 추가하되 그 앞의 `exit 0`을 남기는 구현은 두 번째 assertion에서 실패해야 한다.

### AC-GSA-007 — PR target 선택 보존

**Given** graph-freshness가 pull-request event를 처리한다.  
**When** head ref가 ordinary branch 또는 `release/*`인 두 경우를 평가한다.  
**Then** ordinary PR는 `origin/<base_ref>`, `release/*`는 checkout된 merge-preview `HEAD`를 target으로 사용한다.

- 분류: 기존 동작 regression guard.
- GREEN path: M2 workflow fixture/static test가 두 분기를 모두 실행하고 선택된 target 문자열을 비교한다.
- mutant probe: 모든 PR를 `HEAD`로 완화하거나 모든 PR를 base ancestry로 통일하는 두 구현이 각각 한 fixture에서 실패해야 한다.

### AC-GSA-008 — merge/squash/rebase-like 계보 표

**Given** 같은 원본 stamp commit을 포함하는 제어된 로컬 이력을 merge-commit, squash, rebase-like rewritten 형태로 각각 만든다.  
**When** 객체 존재, `HEAD` 조상성, checker 처분을 평가한다.  
**Then** 세 경우 모두 객체는 존재하고, merge-commit에서만 원 stamp가 조상이며 정상 freshness 경로에 진입한다. squash와 rebase-like에서는 원 객체가 존재해도 비조상이고 `VerdictAbsent` 운반체 + non-nil system error로 미측정된다.

- RED-now: `RED-GSA-008`
- GREEN path: M1이 표 전체를 영구 테스트로 편입한 뒤 같은 명령은 `ok .../internal/graph`, exit 0을 낸다.
- mutant probe: object existence를 ancestry로 대신하는 구현은 squash/rebase-like 두 행에서 실패해야 한다. 모든 비선형 이력을 거부하는 구현은 merge 행에서 실패해야 한다.

### AC-GSA-009 — genuine regeneration과 reachable stamping 종결

**Given** M1/M2 변경 뒤 codemaps가 stale이다.  
**When** 생성기 계약의 5문서를 현재 트리에 맞게 실제 재생성하고, checkout에서 도달 가능한 commit으로 스탬핑한 뒤 서로 독립된 ancestry 검사와 freshness 검사를 실행한다.  
**Then** 본문 변경이 증거에 나타나고 stamp ancestry는 exit 0이며 `described-source-diff < 40`이다. provenance만 바꾼 bare restamp는 이 AC를 만족하지 않는다.

- 분류: release-blocking closure. plan 시점에는 재생성 전이므로 PASS를 주장하지 않는다.
- GREEN path: M4가 재생성 전후 body diff, stamp SHA, `merge-base --is-ancestor` 출력/exit, graph check codemaps row/exit를 같은 tree SHA에 귀속해 기록한다.
- mutant probe: body diff가 비어 있는데 stamp만 바꾼 결과는 bare-restamp 회귀 검사 또는 body-diff 조건에서 실패해야 한다.

### AC-GSA-010 — 기존 freshness 계약 보존

**Given** 기존 fixture가 bare restamp, committed/uncommitted genuine regeneration, body absent C1, dirty fingerprint, reverted churn, described-root scope, one-below/exact-threshold, contribution/driving paths를 각각 구성한다.  
**When** 영향받는 `internal/graph`와 `internal/cli` 테스트를 실행한다.  
**Then** 모든 fixture가 기존 의미로 통과하고, 실제 실행 테스트 수가 0이 아니다.

- 분류: regression guard.
- GREEN path: M3가 이름 지정 검사를 먼저 실행한 뒤 영향 package 전체를 실행하고 원문 출력을 기록한다.
- mutant probe: ancestry precheck를 dirty path에도 적용하거나 body-absent C1을 exit 2로 바꾸는 구현은 기존 테스트에서 실패해야 한다.

### AC-GSA-011 — 네 렌즈 교차 모순 검사

**Given** `plan.md §B`가 생산자/소비자, release 위상, 관련 SPEC, 회귀 위험 네 렌즈를 근거 위치·결정·반증 조건으로 열거한다.  
**When** plan-auditor가 `spec.md`, `plan.md`, `acceptance.md`와 열거된 source/related SPEC을 함께 읽는다.  
**Then** 다음 네 문장이 동시에 참이어야 한다: (1) ancestry는 freshness보다 먼저다, (2) 비조상은 미측정 exit 2이고 reachable stale은 숫자 exit 1이다, (3) ordinary/release/push target 선택은 각각 base/merge-preview HEAD/push HEAD다, (4) 복구는 genuine regeneration + reachable stamp이고 bare/automatic restamp가 아니다. 하나라도 반대 문구가 있으면 FAIL이다.

- 분류: plan-audit contradiction gate.
- 현재 근거: `plan.md §B`의 4행과 `§B.1`의 4개 결론.
- GREEN path: plan-auditor가 각 행의 반증 조건을 source citation에 대조하고 결과를 카드 plan-audit report에 남긴다.
- mutant probe: `VerdictAbsent`를 실제 body absence라고 단정하거나 release PR를 base ancestry로 바꾸는 문장을 넣으면 감사에서 실패해야 한다.

### AC-GSA-012 — 범위와 품질 종결

**Given** M1~M4 변경과 M5 증거가 준비되었다.  
**When** SPEC lint, 영향 package test/vet/lint, workflow probes, diff scope 검사를 실행한다.  
**Then** REQ-GSA-001~012가 모두 AC에 추적되고, 새 dependency/schema/verdict/threshold/predicate/자동 restamp가 없으며 `.moai/state/**`와 무관 파일 변경이 없다. 로컬 full suite는 실행하지 않는다.

- 분류: release closure.
- GREEN path: M5가 명령·출력·exit/tree SHA를 기록하고 origin/develop CI를 전체 suite 판정으로 남긴다.
- mutant probe: provenance schema나 threshold 40을 바꾸는 diff는 scope 검사에서 실패해야 한다.

## §D. RED-now 증거 원장

### RED-GSA-001

```text
command: bash .moai/reports/t688/red-ancestry.sh
tree_sha: 7097e6e214195c45e65cab5fa565b19ca4514c4e
exit_code: 1
stdout_stderr:
--- FAIL: TestT688ExistingNonAncestorStampIsUnmeasured (1.07s)
    t688_red_ancestry_test.go:24: existing non-ancestor stamp must be freshness-unmeasured with a system error; got verdict="fresh" value=1 threshold=40 content_anchor="a0c2d580b0df62c1ff9ced549f303d349b1e9a38"
FAIL
FAIL github.com/modu-ai/moai-adk/internal/graph 1.642s
FAIL
classification: EXPECTED_RED — intended assertion ran and failed because ancestry precheck is absent.
```

### RED-GSA-002

```text
command: bash .moai/reports/t688/red-cli-unreachable.sh
tree_sha: 7097e6e214195c45e65cab5fa565b19ca4514c4e
exit_code: 1
stdout_stderr:
2026/09/13 22:41:27 WARN config sections directory not found, using defaults path=/var/folders/kt/nq2q81cn4gx3y41r7x47ggmr0000gn/T/TestT688CLIExistingNonAncestorStampExitsTwoWithRecovery3422650000/001/.moai/config/sections
--- FAIL: TestT688CLIExistingNonAncestorStampExitsTwoWithRecovery (1.19s)
    t688_red_ancestry_test.go:38: existing non-ancestor stamp must exit 2; err=graph freshness check failed (stale or absent layer) stdout="codemaps  metric=described-source-diff value=1 threshold=40 verdict=fresh\nmx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh\nedges     metric=source-fingerprint-mismatch value=1 threshold=0 verdict=stale  (source set(s) moved: codemaps)\ncitations metric=positive-cited-path-absence value=0 threshold=0 verdict=fresh\n" stderr="graph check: layer edges verdict=stale value=1 threshold=0 — source set(s) moved: codemaps\n"
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 2.106s
FAIL
classification: EXPECTED_RED — intended exit-2 assertion ran; current path returned the stale/absent exit-1 carrier instead.
```

### RED-GSA-006

```text
command: python3 .moai/reports/t688/red-push-guard.py
tree_sha: 7097e6e214195c45e65cab5fa565b19ca4514c4e
exit_code: 1
stdout:
push_branch_has_target_head=false
push_branch_exits_before_ancestry=true
classification: EXPECTED_RED — both assertions identify the current push early-success gap.
```

### RED-GSA-008

```text
command: .moai/reports/t688/red-history-topologies.sh
tree_sha: 7097e6e214195c45e65cab5fa565b19ca4514c4e
exit_code: pending-plan-probe
stdout_stderr: pending-plan-probe
classification: must be executed before adoption; if the intended assertion is not reached, classify TOOL_FAILURE rather than EXPECTED_RED.
```

## §E. 기준선 및 과거 관측

### BASE-GSA-001 — 현재 freshness 기준선

```text
command: moai graph check --root .
tree_sha: 7097e6e214195c45e65cab5fa565b19ca4514c4e
exit_code: 1
selected_stdout:
codemaps  metric=described-source-diff value=254 threshold=40 verdict=stale
selected_stderr:
graph check: layer codemaps verdict=stale value=254 threshold=40 —
  measured from: f7b491954 (working-tree-differs-from-stamp)
classification: current reproducible baseline, not an implementation PASS.
```

### REG-GSA-224 — 재현되지 않은 과거 값

```text
historical_value: 224
current_tree_reproduction: not reproduced
replacement_measurement: BASE-GSA-001 value=254
classification: regression guard only; MUST NOT be used as an exact AC value or current PASS.
```

## §F. 경계 및 오류 사례

1. commit 객체 자체가 없음: 기존 not-comparable system error/exit 2를 유지하고 ancestry 비교를 시도해 fresh/stale을 만들지 않는다.
2. 객체는 존재하지만 비조상: 이 SPEC의 신규 미측정 경로다.
3. body가 실제로 없음(C1): 기존 `VerdictAbsent` + nil error/exit 1을 유지한다.
4. dirty fingerprint: commit ancestry를 적용하지 않는다.
5. stamp가 `HEAD` 자신: ancestor-or-self이므로 정상 freshness 계산에 진입한다.
6. threshold 39/40: 각각 fresh/stale이다.
7. release merge preview: head history가 stamp를 포함하면 통과한다.
8. squash/rebase-like: 원 객체가 checkout에 남아도 `HEAD` 조상이 아니면 실패한다.
9. workflow의 anchorless provenance: 기존 skip-with-reason 계약을 유지한다.
10. test selector가 0건: green이 아니라 검증 실패다.

## §G. 품질 게이트와 Definition of Done

- [ ] `moai spec lint SPEC-GRAPH-STAMP-ANCESTRY-001 --strict`가 exit 0이다.
- [ ] Tier M artifact set은 `spec.md`, `plan.md`, `acceptance.md`이며 `progress.md` lifecycle skeleton이 있다.
- [ ] AC-GSA-001/002/006/008의 RED가 의도한 assertion에서 관측되었고 GREEN 뒤 같은 명령이 exit 0이다.
- [ ] AC-GSA-004/005/007/010 regression guard가 실제 테스트를 1개 이상 실행하고 통과한다.
- [ ] merge/squash/rebase-like 세 fixture가 모두 실행된다.
- [ ] push/ordinary PR/release PR 세 target 경로가 모두 실행된다.
- [ ] genuine codemaps 5문서 재생성과 reachable stamping이 별도 증거로 남는다.
- [ ] `described-source-diff < 40`과 stamp ancestry exit 0이 같은 tree SHA에 귀속된다.
- [ ] bare/automatic restamp, schema/verdict/threshold/predicate 변경이 없다.
- [ ] 영향 package test/vet/lint가 통과하고 origin/develop CI가 전체 suite를 판정한다.
- [ ] `progress.md §E.2`와 `§E.3`은 manager-develop이 run evidence로 채우며 plan 단계에서 선기입하지 않는다.

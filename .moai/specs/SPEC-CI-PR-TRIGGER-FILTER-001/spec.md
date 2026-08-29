---
id: SPEC-CI-PR-TRIGGER-FILTER-001
title: "graph-freshness.yml pull_request 트리거의 무필터 생략 교정 — 발화 집합을 선언으로 고정"
version: "0.1.0"
status: draft
created: 2026-08-29
updated: 2026-08-29
author: manager-spec
priority: P3
phase: "v3.1.4 target"
module: ".github/workflows"
lifecycle: spec-first
tags: "ci, github-actions, trigger, git-flow, graph-freshness"
tier: S
era: V3R6
related_specs: [SPEC-GRAPH-FRESHNESS-CADENCE-001, SPEC-GITFLOW-DOCTRINE-ALIGN-001]
---

# SPEC: `Graph Freshness` 워크플로 `pull_request` 트리거 필터 명시화

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-08-29 | manager-spec | 최초 작성 — 카드 t294. 근거는 `.moai/reports/t294/investigation.md`(측정 트리 `.claude/worktrees/t294`, HEAD `51daada00`) + 본 세션에서 추가로 실행한 실측 5건(§1.3·§1.4) |

---

## 1. 배경

### 1.1 현재 선언 (측정 트리 기준)

`.github/workflows/graph-freshness.yml` L3-6:

```yaml
on:
  pull_request:
  push:
    branches: [main, develop]
```

`pull_request:`에는 `branches:`도 `paths:`도 없다. 즉 **base 브랜치가 무엇이든** 모든 PR에서 발화한다.

### 1.2 판정 — 의도가 아니라 누락이다

조사 보고서 §2.2(`git blame -L 1,8`)의 귀속:

- `6786c3fa4` (2026-08-26) 파일 최초 작성. 당시 PR base는 `main` 하나뿐이라 무필터 `pull_request:`와
  `branches: [main]`은 **동치**였다.
- `11216d13f` (2026-08-27) git-flow 전환. 커밋 메시지가 스스로 범위를 못박는다 —
  *"6 CI workflows: push trigger branches [main] -> [main, develop]"*. `push:` 줄만 손댔고,
  6개 워크플로 어디에서도 `pull_request` 블록을 건드리지 않았다.

무필터를 **유지하겠다는 판단이 어디에도 기록돼 있지 않고**, 그 문장이 쓰인 시점에는 develop이 존재하지도
않았다. 따라서 조치 방향은 "무필터를 의도로 문서화"가 아니라 **필터를 명시**하는 쪽이다.

### 1.3 [정정] 카드가 서술한 이중 발화는 이 필터로 닫히지 않는다

카드는 무필터를 develop→main PR의 **이중 발화**(같은 head 커밋에 push 런 1개 + pull_request 런 1개)의
원인으로 읽는다. 그러나 GitHub Actions에서 `pull_request`의 `branches:` 필터는 PR의 **base** 브랜치를
거른다 — head가 아니다. 본 세션 WebFetch로 GitHub 공식 문서를 확인했다(2026-08-29,
`docs.github.com/en/actions/reference/workflows-and-actions/events-that-trigger-workflows`):
*"You can use the `branches` or `branches-ignore` filter to configure your workflow to only run on pull
requests that **target** specific branches."*

develop→main PR의 base는 `main`이므로, `branches:`에 `main`이 있는 한 그 PR의 `pull_request` 런은
그대로 발화한다. 동시에 develop에 커밋이 얹히면 `push` 런도 발화하고, `concurrency.group`이
`${{ github.workflow }}-${{ github.ref }}`(L13)라 두 런의 ref(`refs/heads/develop` vs
`refs/pull/N/merge`)가 갈려 서로를 취소하지 못한다.

**따라서 이중 발화의 원인은 누락된 `branches:` 필터가 아니라 ref-키 concurrency 그룹이다.** 본 SPEC은
이중 발화를 닫는다고 주장하지 않으며(REQ-PTF-004), 그 처리는 §4의 범위 밖 항목 + kickoff 결정 축 B로
올린다. 이 정정이 없으면 "필터를 넣었으니 이중 발화가 사라졌다"는 미관측 주장이 그대로 남는다.

### 1.4 무필터가 실제로 넓히는 집합 (본 세션 실측)

무필터가 오늘 만들어내는 초과 발화는 "base가 `main`도 `develop`도 아닌 PR"이다. 최근 PR 100건의 base 분포:

```console
$ gh pr list --state all --limit 100 --json baseRefName \
    --jq '[.[].baseRefName] | group_by(.) | map({b: .[0], n: length})'
[{"b":"develop","n":1},{"b":"main","n":99}]
```

- base `develop` 1건 = PR #1677.
- 조사 보고서 §2.4가 관측한 유일한 `pull_request` 런(head `fix/heavy-gate-nested-toolchain`)의 base는
  본 세션 실측 결과 `main`이다
  (`gh pr list --state all --limit 200 --search 'head:fix/heavy-gate-nested-toolchain' --json number,baseRefName,headRefName`
  → `[{"baseRefName":"main","headRefName":"fix/heavy-gate-nested-toolchain","number":1681}]`).

즉 `branches: [main, develop]`로 좁히면 **오늘 기준 제거되는 발화는 0건**이고, 앞으로 생길 수 있는
비-통합-브랜치 base PR(예: `release/*`로 향하는 PR, feature→feature 스택 PR)에서만 발화가 사라진다.
`[main]`만 쓰면 PR #1677 같은 develop-base PR의 병합 전 검사가 사라지므로, 채택 형태는 `[main, develop]`다
— push 트리거가 이미 선언한 git-flow 통합 브랜치 집합과 동일하게 맞추는 것이기도 하다.

### 1.5 선례 — 같은 자리를 이미 다룬 두 워크플로

`spec-lint.yml`(L3-15)과 `docs-i18n-check.yml`은 git-flow 전환 때 push에 develop을 넣으면서 이유를 주석으로
고정했다. 예: *"develop is the integration branch under git-flow: lane work merges into it directly, with no
card PR to carry the pull_request trigger above."* 두 곳 모두 `pull_request` 필터는 손대지 않았다
(`paths:`로만 좁혀져 있다). 본 SPEC이 채택하는 형태는 **그 주석 관행은 따르되, 필터는 명시**하는 쪽이다 —
graph-freshness는 `paths:`조차 없어 노출 폭이 그 둘보다 넓기 때문이다(§4 결정 축 A 참조).

---

## 2. 요구사항 (GEARS)

**REQ-PTF-001** — The `Graph Freshness` workflow's `pull_request` trigger shall declare an explicit
`branches:` filter naming the git-flow integration branches `main` and `develop`, so that the trigger's
firing set is stated in the declaration rather than inherited from an absent filter.

> 값 선택 근거는 §1.4(실측): `[main, develop]`는 오늘의 발화 집합에서 아무것도 제거하지 않으면서
> push 트리거가 이미 선언한 브랜치 집합과 일치한다. `[main]`은 PR #1677 형태(develop-base PR)의
> 병합 전 검사를 잃는다.

**REQ-PTF-002** — A comment adjacent to the `pull_request` trigger shall record why the filter is
explicit — that the unfiltered form was authored when `main` was the only pull-request base, and that
git-flow silently widened the same statement — following the in-repo commenting precedent of
`spec-lint.yml`.

**REQ-PTF-003** — Where a pull request targets `main` or `develop`, or a push lands on `main` or
`develop`, the workflow shall continue to run exactly as it does today; the change shall remove firing
only for pull requests whose base is neither branch.

**REQ-PTF-004** — The workflow's `concurrency` block shall remain unchanged, and this SPEC shall not
claim to close the duplicate-run behaviour on a develop-head pull request: because the `branches:`
filter matches the base branch and not the head (§1.3), that duplication survives the change and is
recorded as an open defect rather than as a fixed one.

**REQ-PTF-005** — While `develop` is frozen and no pull request whose head is `develop` can be opened,
verification shall rest on parser-level assertions over the trigger declaration, and the
execution-based observation shall be registered as a deferred check rather than asserted as performed.

---

## 3. 수락 기준 (Tier S — 인라인)

| REQ | 피복 AC |
|---|---|
| REQ-PTF-001 | AC-PTF-001 |
| REQ-PTF-002 | AC-PTF-002 |
| REQ-PTF-003 | AC-PTF-003 |
| REQ-PTF-004 | AC-PTF-004 |
| REQ-PTF-005 | AC-PTF-006 |
| (비회귀) | AC-PTF-005 |

미피복 REQ 0건, 고아 AC 0건.

**AC-PTF-001** — Given the edited workflow, When
`yq -o=json '.on.pull_request.branches' .github/workflows/graph-freshness.yml` runs,
Then the output is exactly `["main","develop"]`.
*RED-now (본 세션 실측, HEAD `51daada00`)*: 같은 명령이 `null`을 출력한다 — 필터가 없다.

**AC-PTF-002** — Given the edited workflow, When
`grep -n -B6 '^  pull_request:' .github/workflows/graph-freshness.yml` runs,
Then at least one `#` comment line appears in the printed context above the `pull_request:` key, and
that comment names both the authoring-time equivalence (`main` was the only base) and git-flow as what
widened it.
*RED-now*: 같은 명령의 출력에 주석 0줄 — 파일의 유일한 주석(`# Cancel in-progress runs …`, L11)은
`concurrency` 블록 소속이라 이 문맥에 나타나지 않는다.

**AC-PTF-003** (two-cell — 보존 축) —
- (a) `yq -o=json '.on.push' .github/workflows/graph-freshness.yml` → `{"branches":["main","develop"]}`
  — 변경 전후 동일.
- (b) `git diff --stat origin/develop -- .github/workflows/` → 정확히 1 파일(`graph-freshness.yml`),
  그리고 `git diff origin/develop -- .github/workflows/graph-freshness.yml`의 모든 hunk가 `on:` 블록
  (L3-6) 안에 든다 — `jobs:` / `permissions:` / `concurrency:` hunk 0건.

**AC-PTF-004** — Given the edited workflow, When
`yq '.concurrency.group' .github/workflows/graph-freshness.yml` runs, Then the output is
`${{ github.workflow }}-${{ github.ref }}` — 변경 전(본 세션 실측, L13)과 바이트 동일 —
and `yq '.concurrency.cancel-in-progress' …` → `true`, 역시 동일.

**AC-PTF-005** (비회귀) — Given the edited workflow, When
`actionlint .github/workflows/graph-freshness.yml` runs, Then rc=0 and stdout is empty.
*Baseline (본 세션 실측, 변경 전)*: 같은 명령 rc=0, 출력 없음.

**AC-PTF-006** — Given `develop` is frozen so no develop-head pull request can be opened, When run-phase
reports, Then `progress.md` §E.2 carries exactly one deferred-observation row marked `DEFERRED-OBS`
naming `gh run list --workflow=graph-freshness.yml --limit 30 --json event,headSha,headBranch` as the
check to run once develop unfreezes, and no §E row asserts that an execution-based observation was
performed.
Verified by: `grep -c 'DEFERRED-OBS' .moai/specs/SPEC-CI-PR-TRIGGER-FILTER-001/progress.md` → `1`.

---

## 4. 범위 밖 (exclusions)

### Out of Scope — paths-scoped 무필터 3종 (kickoff 결정 축 A)
- `lsel-leak-guard.yaml` · `spec-lint.yml` · `docs-i18n-check.yml` 도 `pull_request`에 `branches:`가
  없다(조사 §4 표 — 카드가 `lsel-leak-guard.yaml`에 `branches: [main]`이 있다고 적은 것은 반증됐다).
- 본 SPEC은 이 셋을 건드리지 않는다. 셋 다 `paths:`로 좁혀져 있어 노출 폭이 graph-freshness(무조건 발화)보다
  작고, 한 카드에서 4개 워크플로를 함께 바꾸면 diff 귀속이 흐려진다.
- 함께 정규화할지는 **운영자 결정**이며 kickoff에 올린다(권고: 좁은 범위 — graph-freshness 1개).

### Out of Scope — 이중 발화 제거 (kickoff 결정 축 B)
- develop-head PR의 중복 런은 `concurrency.group`을 head SHA 키
  (`${{ github.event.pull_request.head.sha || github.sha }}`)로 바꾸면 닫히지만, 그러면 push 런이
  브랜치 보호가 요구하는 PR 런을 취소해 체크가 `cancelled`로 남을 수 있다 — 미측정 위험이다.
- 본 SPEC은 concurrency 블록을 건드리지 않는다(REQ-PTF-004). 채택 여부는 kickoff 결정.

### Out of Scope — push 트리거
- `push: branches: [main, develop]`은 그대로 둔다. develop은 카드 PR 없이 직접 병합되는 통합 브랜치라
  push 트리거가 유일한 발화 경로다(선례 주석, §1.5).

### Out of Scope — 실행 기반 회귀 관측
- develop을 head로 하는 PR을 실제로 여는 관측은 develop 동결로 불가하고, 카드 브랜치 push는 CI-inert다
  (lane-4 실측, 조사 §5). 본 카드는 관측을 수행하지 않고 유예 항목으로만 등록한다(REQ-PTF-005).

### Out of Scope — graph-freshness 잡 내용과 실행 주기
- 잡 스텝, 검사 로직, 실행 주기(cadence)는 손대지 않는다 — SPEC-GRAPH-FRESHNESS-CADENCE-001 소관.

### Out of Scope — `paths:` 필터 도입
- graph-freshness에 `paths:`를 붙여 발화를 더 좁히는 것은 검사 성격(그래프 신선도는 어떤 변경으로도
  낡을 수 있다)에 대한 별개 판단이다. 본 SPEC은 base 필터만 다룬다.

---

## 5. 제약

- **영향 파일 1개**: `.github/workflows/graph-freshness.yml`. 변경 라인은 `on:` 블록 내부로 한정
  (필터 2줄 + 주석). 이 범위를 넘으면 설계를 다시 본다.
- **템플릿 미러 없음(재확인 대상)**: `.github/workflows/`는 `internal/template/templates/` 관리 대상이
  아니라고 보나, run-phase에서 `ls internal/template/templates/.github` 로 직접 확인한 뒤 진행한다.
- **카드 브랜치는 CI-inert**: 이 변경의 판정 주체는 develop 통합 후의 CI다. 레인 로컬 검증은
  `actionlint` + `yq` 파서 단언까지가 전부이며, 그 이상을 실행 증거로 주장하지 않는다.
- **결정 축 2개가 kickoff 전에 닫혀야 한다**(§4 축 A·B). 닫히기 전 run-phase 진입 금지.

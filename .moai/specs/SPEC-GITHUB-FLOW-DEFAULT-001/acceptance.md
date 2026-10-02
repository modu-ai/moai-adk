# SPEC-GITHUB-FLOW-DEFAULT-001 — 인수 기준

> 문서 수준 트리 핀: 카드 트리 `4bf547bcad7c155b1e91485921569db709ec3ac2`. 이 핀은 자체 핀이 없는 모든 기준에 적용되고, 기준 수준 핀이 있으면 그것이 이긴다. 브랜치 이름은 핀으로 쓰지 않는다.

## §A 측정 규약과 분류

- **RED-now 셀**은 네 요소를 함께 담는다 — 명령(읽기 전용 단일 호출), 그 명령의 verbatim stdout(원본 파일 바이트), 종료 코드(별도 필드), 트리 SHA. 네 요소는 아래 **증거 원장**(§B)에 항목 ID 로 기록하고 기준 표가 ID 를 인용한다. 표 셀은 셸 메타문자를 훼손하므로 원장이 운반체다.
- **녹색 경로 셀**은 어느 마일스톤이 기준을 뒤집는지와 통과 출력의 모양을 적는다.
- **분류**: `release-blocking`(절체 변경 묶음의 병합 게이트)과 `regression-guard`. RED 를 이 트리에서 다시 실행할 수 없는 기준은 `regression-guard` 로 분류되고 통과로 기록되지 않는다.
- **돌연변이 프로브**: 요구를 어기면서 기준을 만족하는 변이를 쓸 수 있으면 그 기준은 너무 얕다. 기준마다 죽여야 할 변이와 그것을 죽이는 입력을 적는다. 무효 입력만 보는 기준은 전부-일치 변이를, 유효 입력만 보는 기준은 전부-불일치 변이를 통과시키므로 양방향 입력을 함께 둔다.
- **빈 스윕 금지**: 모든 시험 선택은 `-run` 을 앵커로 고정하고 실행된 시험 수를 확인한다. `[no tests to run]`·`no test files` 는 통과가 아니다.
- **연속 발화**: 가드 시험은 `SKIP` 으로 끝날 수 있으면 완료가 아니다. 통과 출력은 `--- PASS` 줄과 방문 수를 포함해야 하고 `--- SKIP` 이 없어야 한다.

## §B 증거 원장 (RED-now 4요소)

모든 항목의 트리는 `4bf547bcad7c155b1e91485921569db709ec3ac2` 다. 종료 코드는 같은 명령을 `; echo "exit=$?"` 를 붙여 따로 관측한 값이다(셀의 명령에는 붙이지 않는다).

```
E-01
command: git grep -n "landingBaseBranch = " -- internal/cli/worktree/done.go
stdout:
internal/cli/worktree/done.go:281:const landingBaseBranch = "develop"
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-02
command: git grep -n "\"origin/develop\"" -- internal/cli/worktree/sweep.go
stdout:
internal/cli/worktree/sweep.go:224:	cmd.Flags().String("base", "origin/develop", "Remote integration base the landing check compares against (default diverges from clean --stale's origin/main)")
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-03
command: git grep -n "\.DevelopBranch" -- internal/cli/factory_card.go internal/cli/integration.go internal/cli/factory_merge.go internal/cli/session_worktree_automerge.go internal/closure/pushcheck.go internal/cli/contract_pushcheck.go
stdout:
internal/cli/factory_card.go:1064:	if branch := strings.TrimSpace(config.LoadGitFlowIntegrationConfig(root).DevelopBranch); branch != "" {
internal/cli/factory_merge.go:93:				developRef = config.LoadGitFlowIntegrationConfig(integrationLockRoot()).DevelopBranch
internal/cli/integration.go:353:			branch, wt, source := resolveIntegrationTarget(branchFlag, gitFlow.DevelopBranch)
internal/cli/session_worktree_automerge.go:176:	if !gitFlow.IsGitFlow() || gitFlow.DevelopBranch == "" {
internal/cli/session_worktree_automerge.go:183:	develop := gitFlow.DevelopBranch
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-04
command: git grep -c "prerelease" -- .goreleaser.yml
stdout:
exit: 1
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-05
command: git grep -n "Must be on 'main'" -- scripts/release.sh
stdout:
scripts/release.sh:128:        die "Must be on 'main' branch (current: $CURRENT_BRANCH). Use --hotfix for hotfix branches."
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-06
command: git grep -n -E "check (5|6) \(" -- .github/workflows/release.yml
stdout:
.github/workflows/release.yml:107:            fail "check 5 (CHANGELOG): CHANGELOG.md at ${TAG_COMMIT} has no '## [${VERSION_NO_V}]' or '## [v${VERSION_NO_V}]' section."
.github/workflows/release.yml:115:            fail "check 6 (version SSOT): ${SSOT_PATH} at ${TAG_COMMIT} has version='${ACTUAL}', expected '${TAG}'."
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-07
command: git grep -n -w "develop" -- AGENTS.md
stdout:
AGENTS.md:76:`workflow.branch_guard.deny_commits_on`), card work flows on develop-based worktrees, and
AGENTS.md:145:**Start a new card in a new worktree from local `develop`.** A Claude Code session exits its
AGENTS.md:147:tree's HEAD equals the local `develop` tip before editing; never reuse the previous card's tree.
AGENTS.md:149:worktree. When complete, merge the card branch into local `develop` through the serial integration
AGENTS.md:150:window; the factory leader pushes `develop` after the local merges.
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-08
command: git grep -n -E "^[[:space:]]+(worktree_base_branch|workflow|develop_branch): " -- .moai/config/sections/git-strategy.yaml
stdout:
.moai/config/sections/git-strategy.yaml:5:    worktree_base_branch: develop
.moai/config/sections/git-strategy.yaml:9:        workflow: git-flow
.moai/config/sections/git-strategy.yaml:16:        develop_branch: develop
.moai/config/sections/git-strategy.yaml:43:        workflow: github-flow
.moai/config/sections/git-strategy.yaml:70:        workflow: github-flow
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-09
command: git grep -n -i -E "workflow_dispatch|Release Verify|release-pr-multi-os" -- scripts/release.sh
stdout:
exit: 1
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-10
command: git grep -c -F "NO L2/L3 worktree at this step" -- .claude/rules/moai/workflow/spec-workflow.md
stdout:
.claude/rules/moai/workflow/spec-workflow.md:1
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-11
command: git rev-parse 4755c5e506225ba90b7a303c5763fa303c699492^{tree} 284e09c44023598affe486f17701717ca173e6ca^{tree}
stdout:
d57fdeecf4b8849da70a6693bbff19ad77ad5932
4d8a7f46189a3db1c6cad1bb6715e76e90b00b15
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-12
command: git rev-list --count 284e09c44023598affe486f17701717ca173e6ca..4bf547bcad7c155b1e91485921569db709ec3ac2
stdout:
122
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-13
command: git grep -l "TestGitHubFlowSweep" -- internal/template
stdout:
exit: 1
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-14
command: git grep -n "\"merge\", \"--no-ff\"" -- internal/cli/factory_card.go
stdout:
internal/cli/factory_card.go:1157:	merge := exec.Command("git", "merge", "--no-ff", "-m",
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-15
command: git grep -c -E "cherry|patch-id|is-ancestor|gh pr" -- internal/cli/worktree/done.go
stdout:
exit: 1
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-16
command: git grep -n -i -E "gh pr create|pr-open" -- internal/cli/factory_card.go internal/homestate/card_transition.go
stdout:
exit: 1
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-17
command: git grep -c -w "develop" -- docs-site/content/ko/workflow-commands/moai-sync.md docs-site/content/en/workflow-commands/moai-sync.md docs-site/content/ja/workflow-commands/moai-sync.md docs-site/content/zh/workflow-commands/moai-sync.md
stdout:
docs-site/content/en/workflow-commands/moai-sync.md:6
docs-site/content/ja/workflow-commands/moai-sync.md:5
docs-site/content/ko/workflow-commands/moai-sync.md:5
docs-site/content/zh/workflow-commands/moai-sync.md:5
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-18
command: git grep -l -w "develop" -- README.md README.ko.md README.ja.md README.zh.md
stdout:
README.ja.md
README.ko.md
README.md
README.zh.md
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-19
command: git grep -n -w "develop" -- .coderabbit.yaml
stdout:
.coderabbit.yaml:69:      - develop
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-20
command: git grep -n "develop:refs/remotes/origin/develop" -- .github/workflows/spec-lint.yml
stdout:
.github/workflows/spec-lint.yml:70:        run: git fetch origin main:refs/remotes/origin/main develop:refs/remotes/origin/develop
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-21
command: git grep -n -E "branches: \[main" -- .github/workflows/ci.yml
stdout:
.github/workflows/ci.yml:18:    branches: [main, develop]
.github/workflows/ci.yml:20:    branches: [main]  # main으로 향하는 모든 PR에서 CI 실행
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-22
command: git rev-list --count --left-right 284e09c44023598affe486f17701717ca173e6ca...4755c5e506225ba90b7a303c5763fa303c699492
stdout:
7614	1
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-23
command: git merge-tree --write-tree --name-only 284e09c44023598affe486f17701717ca173e6ca 4755c5e506225ba90b7a303c5763fa303c699492
stdout:
992f57f42aee6c3f9a39754c3e0095eaadf9d44b
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-24
command: git grep -n "Step 4 (cleanup) applies to" -- .claude/rules/moai/core/zone-registry.md
stdout:
.claude/rules/moai/core/zone-registry.md:850:  clause: "Step 4 (cleanup) applies to **Route B only**. It MUST happen ONLY after BOTH run AND sync PRs are merged"
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

```
E-25
command: git grep -c "Step 3.3.5" -- .claude/skills/moai/workflows/sync/
stdout:
.claude/skills/moai/workflows/sync/delivery.md:1
.claude/skills/moai/workflows/sync/quality-gates-context.md:1
exit: 0
tree: 4bf547bcad7c155b1e91485921569db709ec3ac2
```

## §C 인수 기준

### AC-GFD-001 — 기준 해석은 구성에서 나온다 (maps REQ-GFD-001)

- **분류**: release-blocking (M1 종료 게이트)
- **Given** 구성 둘 — `workflow: git-flow` + `develop_branch: develop`, `workflow: github-flow`. 그리고 해석표의 나머지 행(gitlab-flow·release-flow)을 담은 구성.
- **When** 카드 전달 경로의 기준 해석(통합 창 획득·병합 준비 점검·세션 종료 정리·push 사전 점검·카드 diff 범위)을 각 구성에서 실행하면
- **Then** git-flow 구성은 현행 출력(develop)과 같고, github-flow 구성은 `main` 을 해석하며, 표의 모든 행에서 해석 결과가 통합 목표 해석기의 결과와 같다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-03 (종료 코드 0, 해석 지점 5곳이 `DevelopBranch` 를 읽는다) |
| RED 인 이유 | `DevelopBranch` 는 manual 프로필+git-flow 일 때만 채워지므로 github-flow 구성에서는 이 지점이 빈 기준을 본다. 기준이 이 지점들을 통합 목표 해석기로 옮길 때까지 읽기가 남는다 |
| 돌연변이 프로브 | (1) 읽기를 리터럴 `"main"` 으로 바꾼 변이는 github-flow 행은 통과하고 git-flow 특성화 행에서 죽는다. (2) `DevelopBranch` 가 비면 `"main"` 으로 대체하는 변이는 두 행을 모두 통과하지만 gitlab-flow 행(해석 결과가 `main` 이 아니다)에서 죽는다. 그래서 시험은 해석표의 **모든 행**을 표 기반으로 돈다 |
| 녹색 경로 | M1 이 읽기를 통합 목표 해석기로 옮기고 git-flow 특성화 시험을 먼저 고정한다 |
| 통과 출력 | `go test -v -run '^(TestDeliveryBaseResolvesFromIntegrationTarget|TestDeliveryBaseGitFlowUnchanged)$' ./internal/cli/ ./internal/config/` 가 두 시험 각각 4행 이상의 하위 시험 `--- PASS`, 방문 시험 수 ≥ 2. E-03 의 명령은 종료 코드 1 과 빈 출력 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-002 — squash 병합을 알아보는 착지 판정 (maps REQ-GFD-002)

- **분류**: release-blocking (M2 종료 게이트)
- **Given** 격리 저장소 픽스처 여섯 — F1 커밋 1개 카드를 squash 병합, F2 커밋 2개 카드를 squash 병합, F3 병합 커밋으로 병합, F4 병합되지 않은 카드, F5 squash 병합 뒤 main 이 같은 파일을 다시 바꾼 경우(누적 patch-id 불일치)에 PR 상태 MERGED 를 알려 주는 `gh` 대역, F6 확정할 수 없는 경우(가져오기 실패)
- **When** 착지 판정(`moai worktree done` 의 착지 점검, `moai worktree sweep`, 세션 종료 정리)을 각 픽스처에서 실행하면
- **Then** F1·F2·F3·F5 는 "착지", F4·F6 은 "보존"이다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-15 (종료 코드 1, `done.go` 에 `cherry`·`patch-id`·`is-ancestor`·`gh pr` 가 없다). 보조 관측: `research.md` §5 — 커밋 2개 카드를 squash 병합하면 조상 판정은 종료 코드 1, `git cherry` 는 두 커밋 모두 `+` |
| RED 인 이유 | 현 `done.go` 는 `...` 범위의 우측 개수 0 만 보므로 squash 병합 카드는 영원히 미착지다 |
| 돌연변이 프로브 | `git cherry` 만 쓰는 변이는 F1 만 통과하고 F2 에서 죽는다. 조상 판정만 쓰는 변이는 F3 만 통과한다. "항상 착지" 변이는 F4 에서 죽는다. "항상 보존" 변이는 F1~F3·F5 에서 죽는다 |
| 녹색 경로 | M2 가 세 층(조상, 누적 patch-id, PR 병합 상태)을 `worktree done`·`sweep`·세션 정리가 공유하는 하나의 판정으로 모은다 |
| 통과 출력 | `go test -v -run '^TestLandingPredicateSquashSafe$' ./internal/cli/...` 에서 F1~F6 하위 시험 6개 `--- PASS`, 방문 수 6 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-003 — 기본값에 리터럴 develop 이 없다 (maps REQ-GFD-003)

- **분류**: release-blocking (M1 종료 게이트)
- **Given** 구성 둘(git-flow, github-flow)
- **When** `moai worktree sweep --help` 의 `--base` 기본값, `moai worktree new` 의 기준, 카드 diff 범위 계산의 기준 브랜치를 각 구성에서 읽으면
- **Then** git-flow 는 `origin/develop`·`develop`, github-flow 는 `origin/main`·`main` 이다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-01 과 E-02 (둘 다 종료 코드 0, 리터럴 `develop` 상수와 기본값) |
| RED 인 이유 | 상수와 기본값이 구성과 무관하게 `develop` 이다 |
| 돌연변이 프로브 | 리터럴을 `main` 으로만 바꾼 변이는 github-flow 행을 통과하고 git-flow 행에서 죽는다. 구성 값을 읽되 빈 값이면 `develop` 으로 되돌리는 변이는 github-flow 행에서 죽는다 |
| 녹색 경로 | M1 |
| 통과 출력 | E-01·E-02 명령이 종료 코드 1 과 빈 출력, `go test -v -run '^TestBaseDefaultsFollowIntegrationTarget$' ./internal/cli/...` 하위 시험 ≥ 2 `--- PASS` |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-004 — 카드 PR 전달 간선 (maps REQ-GFD-004)

- **분류**: release-blocking (M2 종료 게이트)
- **Given** 카드 브랜치를 가진 격리 저장소와 bare 원격, 그리고 `gh pr create`·`gh pr view` 에 응답하는 `gh` 대역, `workflow: github-flow` 구성의 merge-ready 카드
- **When** `moai factory complete` 를 실행하면
- **Then** 카드 브랜치가 원격으로 push 되고, 대상이 `main` 인 PR 이 열리며, `gh pr view` 대역이 MERGED 를 돌려주기 전에는 카드가 병합 완료로 기록되지 않고, 돌려준 뒤에 기록된다. `git merge` 는 한 번도 호출되지 않고 통합 창 잠금 파일은 만들어지지 않는다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-16 (종료 코드 1, PR 간선 부재) 과 E-14 (종료 코드 0, `factory_card.go:1157` 의 로컬 `git merge --no-ff`) |
| RED 인 이유 | 현 경로는 로컬 병합만 있고 PR 간선이 없다 |
| 돌연변이 프로브 | PR 개설 직후 병합으로 기록하는 변이는 "MERGED 전에는 미기록" 단언에서 죽는다. 로컬 병합으로 처리하는 변이는 "`git merge` 미호출" 단언에서 죽는다. 창을 잡는 변이는 잠금 파일 단언에서 죽는다. PR 개설 없이 push 만 하는 변이는 `gh pr create` 호출 단언에서 죽는다 |
| 녹색 경로 | M2 |
| 통과 출력 | `go test -v -run '^TestFactoryCompleteGitHubFlowPR$' ./internal/cli/...` 하위 시험 ≥ 4 `--- PASS`. E-16 명령은 종료 코드 0 과 적중 줄 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-005 — PR 전 병합 준비 점검 (maps REQ-GFD-005)

- **분류**: release-blocking (M2 종료 게이트)
- **Given** 픽스처 넷 — 충돌이 있는 카드, 충돌 없는 카드, sync 감사 PASS 기록이 없는 카드, 트리 항등이 깨진 카드
- **When** PR 개설 직전의 병합 준비 점검을 실행하면
- **Then** 충돌·감사 누락·트리 항등 실패 카드는 PR 을 열지 않고 종료 코드가 0 이 아니며 사유를 출력한다. 충돌 없는 카드는 점검을 통과한다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-03 의 `factory_merge.go:93` 줄 (종료 코드 0, 준비 점검의 대상 브랜치가 `DevelopBranch` 에서 나온다) |
| RED 인 이유 | github-flow 구성에서 `DevelopBranch` 가 비어 현 점검은 대상 없이 거부하거나 건너뛴다(입력 보고서의 읽기 결과) |
| 돌연변이 프로브 | 충돌 검사를 건너뛰는 변이는 충돌 픽스처에서 죽는다. 점검을 현재 브랜치 대비로 하는 변이는 충돌 픽스처에서 죽는다. 항상 실패시키는 변이는 충돌 없는 픽스처에서 죽는다 |
| 녹색 경로 | M2 |
| 통과 출력 | `go test -v -run '^TestMergeReadinessBeforePR$' ./internal/cli/... ./internal/factorylane/...` 하위 시험 4개 `--- PASS`, 방문 수 4 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-006 — 창은 선행 조건이 아니고 슬롯은 남는다 (maps REQ-GFD-006)

- **분류**: release-blocking (M2 종료 게이트)
- **Given** github-flow 구성의 merge-ready 카드와 git-flow 구성의 merge-ready 카드
- **When** 각 구성에서 `moai factory complete` 를 실행하고, 별도로 `moai slot acquire`·`release` 를 실행하면
- **Then** github-flow 에서는 `moai integration status` 가 어느 시점에도 보유자를 보이지 않고, git-flow 는 현행 창 동작이 보존되며, 슬롯 획득과 반납은 두 구성 모두에서 종료 코드 0 이다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-14 (종료 코드 0, 로컬 병합 호출이 창 아래에서 실행되는 유일한 병합 경로) |
| RED 인 이유 | 현 `complete` 는 github-flow 에서도 창과 로컬 병합을 전제한다 |
| 돌연변이 프로브 | github-flow 에서 창을 잡고 곧 놓는 변이는 "어느 시점에도 보유자 없음" 단언에서 죽는다. 두 구성 모두에서 창을 없애는 변이는 git-flow 특성화에서 죽는다 |
| 녹색 경로 | M2 |
| 통과 출력 | `go test -v -run '^(TestFactoryCompleteNoWindowGitHubFlow|TestFactoryCompleteWindowGitFlowUnchanged)$' ./internal/cli/...` 각 `--- PASS`, 기존 슬롯 시험 묶음이 변경 없이 통과(방문 수 기록) |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-007 — main HEAD 에서만 태그 (maps REQ-GFD-007)

- **분류**: release-blocking (M3 종료 게이트)
- **Given** bare 원격을 가진 격리 저장소 픽스처 다섯 — (a) `origin/main` 팁에 놓인 detached HEAD, (b) 더 오래된 커밋의 detached HEAD, (c) `main` 브랜치가 `origin/main` 과 같음, (d) `main` 이 아닌 일반 브랜치, (e) `origin/main` 보다 앞선 HEAD
- **When** `scripts/release.sh v9.9.9-rc.1 --dry-run` 을 각 픽스처에서 실행하면
- **Then** (a)(c) 는 브랜치 검사와 동기 검사를 통과하고, (b)(d)(e) 는 종료 코드 1 로 거부한다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-05 (종료 코드 0, 브랜치 이름이 `main` 이어야 한다는 거부). 보조 관측: `research.md` §8 — 카드 트리에서 `--dry-run` 이 종료 코드 1 과 `Must be on 'main' branch` 로 거부, 격리 저장소 detached HEAD 의 `rev-parse --abbrev-ref HEAD` 는 `HEAD` |
| RED 인 이유 | 하네스가 쓰는 detached 워크트리는 브랜치 이름이 `HEAD` 라서 거부된다 |
| 돌연변이 프로브 | 모든 detached HEAD 를 허용하는 변이는 (b) 에서 죽는다. 동기 검사를 건너뛰는 변이는 (e) 에서 죽는다. 브랜치 이름 검사를 그대로 두는 현행은 (a) 에서 죽는다 |
| 녹색 경로 | M3 |
| 통과 출력 | `go test -v -run '^TestReleaseScriptHeadAdmission$' ./scripts/...` 또는 동등 위치에서 하위 시험 5개 `--- PASS` |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-008 — rc 태그의 출처 검증 (maps REQ-GFD-008)

- **분류**: release-blocking (M3 종료 게이트)
- **Given** 출처 검증을 시험 가능한 위치로 옮긴 뒤 격리 저장소에 만든 태그 픽스처 — (1) 올바른 트레일러와 main 조상 커밋을 가진 rc 태그(CHANGELOG 절·`system.yaml` 버전 없음), (2) 트레일러 커밋이 틀린 rc 태그, (3) main 조상이 아닌 커밋의 rc 태그, (4) CHANGELOG 절이 없는 정식 태그, (5) `system.yaml` 버전이 태그와 다른 정식 태그
- **When** 출처 검증을 각 태그에 실행하면
- **Then** (1) 만 통과하고 (2)~(5) 는 각자의 검사 이름을 대며 실패한다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-06 (종료 코드 0, 검사 5·6 이 모든 태그에 적용된다) |
| RED 인 이유 | 접미사 구분이 없어 rc 태그가 CHANGELOG 절 없이는 거부된다 |
| 돌연변이 프로브 | rc 에서 모든 검사를 건너뛰는 변이는 (2)(3) 에서 죽는다. 정식 태그 검사를 약화하는 변이는 (4)(5) 에서 죽는다 |
| 녹색 경로 | M3 |
| 통과 출력 | `go test -v -run '^TestReleaseProvenanceRcRule$' ./internal/...` 하위 시험 5개 `--- PASS` |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-009 — prerelease 표시 (maps REQ-GFD-009)

- **분류**: release-blocking (M3 종료 게이트)
- **Given** `.goreleaser.yml`
- **When** YAML 경로 `release.prerelease` 를 읽으면
- **Then** 값이 `auto` 다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-04 (종료 코드 1, `prerelease` 문자열이 없다) |
| RED 인 이유 | 설정 키가 없다 |
| 돌연변이 프로브 | 키를 다른 중첩 위치에 둔 변이는 문자열 grep 을 통과하지만 YAML 경로 단언에서 죽는다. 값을 `true` 로 둔 변이는 정식 태그도 prerelease 로 만들어 값 단언에서 죽는다 |
| 녹색 경로 | M3 |
| 통과 출력 | `go test -v -run '^TestGoreleaserPrereleaseAuto$' ./internal/...` 하위 시험 `--- PASS`. `goreleaser check` 는 도구가 설치된 환경에서만 종료 코드 0 을 기록하고, 설치되지 않았으면 Gap 으로 적는다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-010 — 태그 전 3-OS 매트릭스 (maps REQ-GFD-010)

- **분류**: release-blocking (M3 종료 게이트)
- **Given** `gh` 대역 다섯 — (a) 대상 SHA 에 성공한 `Release Verify` 런, (b) 런 없음, (c) 다른 SHA 에 성공한 런, (d) 대상 SHA 에 실패한 런, (e) 대상 SHA 에 진행 중인 런
- **When** `scripts/release.sh v9.9.9-rc.1 --dry-run` 이 게이트 검증 단계에 이르면
- **Then** (a) 만 통과하고 나머지는 종료 코드 1 로 태그 전에 거부한다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-09 (종료 코드 1, 스크립트에 매트릭스 게이트 참조가 없다) |
| RED 인 이유 | 현 스크립트는 결합 상태 API 한 번만 본다 |
| 돌연변이 프로브 | 아무 SHA 의 성공 런이나 받는 변이는 (c) 에서 죽는다. 결론을 보지 않는 변이는 (d)(e) 에서 죽는다. 게이트를 건너뛰는 변이는 (b) 에서 죽는다 |
| 녹색 경로 | M3 (옵션 B 채택 시) |
| 통과 출력 | `go test -v -run '^TestReleaseScriptMatrixGate$' ./scripts/...` 또는 동등 위치에서 하위 시험 5개 `--- PASS` |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-011 — 낡은 develop 서술 잔존 0 (maps REQ-GFD-011)

- **분류**: release-blocking (M4 종료 게이트, 절체 변경 묶음의 병합 조건)
- **Given** 절체 변경 묶음이 적용된 트리와 `internal/template` 의 스윕 가드 시험(무장 상태)
- **When** 가드를 실행하면
- **Then** 범위 표면(`AGENTS.md`·`AGENTS.local.md`·`CLAUDE.md`·`.claude/rules/**`·`.claude/agents/**`·`.claude/skills/**`·`.claude/output-styles/**`·README 4종·docs-site 4로케일·`.moai/docs/**`·템플릿 사본)에 살아 있는 develop 기준 서술이 0 이고 방문 수가 바닥값 이상이다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-07 (종료 코드 0, `AGENTS.md` 에 5줄)과 E-18 (종료 코드 0, README 4종 모두 적중) |
| RED 인 이유 | 서술이 살아 있다 |
| 돌연변이 프로브 | 가드의 자체 시험(AC-GFD-021 의 픽스처 양방향)이 죽인다. 추가로 서술을 지우지 않고 허용 목록에만 올리는 변이는 허용 항목마다 `Why`·`Literal` 이 필요하다는 검사에서 죽는다 |
| 녹색 경로 | M4 (t1448 착지 뒤 규칙 편집 포함) |
| 통과 출력 | `go test -v -run '^TestGitHubFlowSweepGuard$' ./internal/template/` 가 `--- PASS` 와 `visited=<N>` 줄을 내고 `--- SKIP` 이 없다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-012 — 전달 경로 서술과 4로케일 (maps REQ-GFD-012)

- **분류**: release-blocking (M4 종료 게이트)
- **Given** docs-site `workflow-commands/moai-sync.md` 4로케일과 README 4종
- **When** 부정 스윕(살아 있는 develop 서술)과 긍정 앵커(카드 브랜치·push·PR·CI 재측정·GitHub 병합 경로 서술)를 각 파일에 적용하면
- **Then** 8개 파일 모두 부정 스윕 0, 긍정 앵커 ≥ 1 이다. 로케일 간 문장 수 일치는 단언하지 않는다. zh 의 추가 live 문장은 M4 의 zh 담당이 해소한다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-17 (종료 코드 0, `moai-sync.md` 의 `develop` 단어 적중이 en 6·ja 5·ko 5·zh 5) |
| RED 인 이유 | 4로케일이 develop 서술을 갖고 있고 적중 수도 로케일마다 다르다 |
| 돌연변이 프로브 | develop 문장만 지우고 새 경로를 서술하지 않는 변이는 긍정 앵커에서 죽는다. 새 경로 문장만 덧붙이고 낡은 문장을 남기는 변이는 부정 스윕에서 죽는다. 한 로케일만 고치는 변이는 8개 파일 단언에서 죽는다 |
| 녹색 경로 | M4 |
| 통과 출력 | 로케일별 부정·긍정 단언이 모두 `--- PASS`, 방문 파일 수 8 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-013 — Frozen 조항의 정규 개정 (maps REQ-GFD-013)

- **분류**: release-blocking (M4 종료 게이트)
- **Given** `zone-registry.md` 의 `CONST-V3R5-027`·`-028` 과 `worktree-integration.md` 의 Frozen 두 줄(per-step applicability, disposal contract)
- **When** 문언을 바꾸는 개정을 `moai constitution amend` 로 수행하면
- **Then** 각 개정이 5단 게이트(FrozenGuard, Canary, ContradictionDetector, RateLimiter, HumanOversight)를 통과해 종료 코드 0 으로 끝나고 그 출력이 판정서에 남으며, 등재 문언과 본문이 일치한다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-24 (종료 코드 0, `CONST-V3R5-028` 문언)과 E-10 (종료 코드 0, `spec-workflow.md` 의 plan 단계 금지 문장 1회) |
| RED 인 이유 | 두 clause 가 개정 전 문언을 담고 있다 |
| 돌연변이 프로브 | 레지스트리를 손으로 고친 변이는 문장 일치 grep 을 통과하지만 amend 실행 출력이 없어서 이 기준에서 죽는다. 본문만 고치고 레지스트리를 두는 변이는 일치 단언에서 죽는다 |
| 녹색 경로 | M4 (t1448 착지 뒤, 이 개정은 SPEC-LATE-BRANCH-REDESIGN-001 의 B3 를 흡수) |
| 통과 출력 | amend 출력 3~4건(종료 코드 0, 5단 통과), 등재 clause 에 이전 문언의 `NO L2/L3 worktree at this step`·`BOTH run AND sync PRs` 가 0, 본문과 등재의 새 문언이 같다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-014 — 상시로드 증가의 진술과 측정 (maps REQ-GFD-014)

- **분류**: regression-guard (측정 기준선이 단일 호출이 아니므로 RED-now 셀 요건을 채우지 못해 통과로 기록하지 않는다)
- **Given** M4 착수 시점의 상시로드 기준선(`research.md` §10 — 16개·285,543 바이트, `CLAUDE.md` 15,658·`AGENTS.md` 20,252 바이트를 이 트리에서 측정)
- **When** M4 의 편집 중 상시로드 파일을 1,000 바이트 넘게 늘리는 편집이 있으면
- **Then** 그 변경 설명이 `rule-authoring.md` 의 진술(파일 바이트, 비호출 세션이 치르는 비용, `paths:` 범위 우선 검토)을 담고, 전후 개수·바이트가 보고된다.

| 셀 | 내용 |
|---|---|
| 회귀 방지 신호 | 편집마다 증가량 측정, 1,000 바이트 초과분 목록과 진술 대조 |
| 녹색 경로 | M4 판정서에 전후 측정과 진술 목록 |
| 한계 | 호스트가 `InstructionsLoaded` 관측을 노출하지 않으면 그 사실을 Gap 으로 보고한다. 바이트 절감은 도달성 확인 없이 성능 결과가 아니다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-015 — 로컬·템플릿 사본 동반 (maps REQ-GFD-015)

- **분류**: regression-guard (기존 정합 시험이 감시하는 불변 조건)
- **Given** M1~M5 에서 로컬과 템플릿 양쪽에 존재하는 파일을 바꾼 변경
- **When** `make agents-emit-check` 와 `go test -run '^(TestCatalogHashParity|TestManifestHashFormat)$' ./internal/spec/... ./internal/template/...` 를 실행하면
- **Then** 두 축 모두 종료 코드 0 이고 선택한 시험 수가 0 이 아니다. 한 축의 초록을 다른 축의 근거로 쓰지 않는다.

| 셀 | 내용 |
|---|---|
| 회귀 방지 신호 | 두 정합 시험이 이미 존재하며 변경이 두 사본의 방향을 어긋나게 하면 적색이 된다 |
| 한계 | 임베드 축(`make build` 가 싣는 내용)은 리더가 배치 종료 때 빌드하므로 이 SPEC 에서 Gap 이며 통과로 기록하지 않는다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-016 — 저장소 설정과 CI 는 절체 시점에 함께 (maps REQ-GFD-016)

- **분류**: release-blocking (M5 종료 게이트)
- **Given** 절체 변경 묶음
- **When** 두 방향을 모두 단언하면 — (사전) 절체 이전에 develop 에 병합되는 커밋 집합(M1~M3)이 아래 파일을 건드리지 않는다, (사후) 묶음이 적용된 트리에서 아래 값이 github-flow 기준이다
- **Then** 사전 방향은 변경 파일 목록에 `.moai/config/sections/git-strategy.yaml`·`.moai/config/sections/workflow.yaml`·`.coderabbit.yaml`·`.github/workflows/*` 가 없고, 사후 방향은 `worktree_base_branch: main`, `workflow: github-flow`, `.coderabbit.yaml` 에 `develop` 없음, `spec-lint.yml` 이 `develop` 을 가져오지 않음, 워크플로 push 트리거에 `develop` 없음이다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-08 (종료 코드 0, 저장소 값이 git-flow), E-19 (종료 코드 0), E-20 (종료 코드 0), E-21 (종료 코드 0, `branches: [main, develop]`) |
| RED 인 이유 | 네 파일이 develop 기준 값을 담고 있다 |
| 돌연변이 프로브 | 구성을 일찍 바꾼 변이는 사후 단언은 통과하지만 사전 단언에서 죽는다. 구성만 바꾸고 CodeRabbit 은 두는 변이는 사후 단언에서 죽는다. `spec-lint.yml` 의 가져오기를 두고 develop 만 삭제하는 시점 변이는 develop 삭제 직전 사전 점검에서 죽는다 |
| 녹색 경로 | M5, 적용은 절체 경계 |
| 통과 출력 | 사후: E-19·E-20 명령이 종료 코드 1 과 빈 출력, E-21 명령이 `branches: [main]` 줄만 출력(`develop` 없음), E-08 명령이 `worktree_base_branch: main`·`workflow: github-flow` 값을 출력. 사전: M1~M3 병합 구간의 변경 파일 목록에 위 파일이 없다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-017 — t810 범위의 흡수 (maps REQ-GFD-017)

- **분류**: release-blocking (M4 종료 게이트)
- **Given** t810 의 REQ-LBR-001..006 과 이 SPEC 의 문서 마일스톤
- **When** 여섯 의도를 하나씩 이 SPEC 의 인수 기준·측정에 대응시키면
- **Then** 대응표가 판정서에 있고, 다음이 모두 참이다 — `spec-workflow.md` 에서 "opens a PR per phase"·"NO L2/L3 worktree at this step"·"BOTH run AND sync PRs" 가 0, `worktree-integration.md` 에서 "BOTH run PR AND sync PR" 가 0, `delivery.md`·`quality-gates-context.md` 에서 `Step 3.3.5` 가 0, 새 서술(SPEC 당 PR 1개·plan 단계 워크트리 진입·단일 PR 폐기 조건)이 각 파일에 있고, 두 사본이 같은 방향이다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-10 (종료 코드 0), E-24 (종료 코드 0), E-25 (종료 코드 0, `Step 3.3.5` 정의와 참조 각 1) |
| RED 인 이유 | 세 문서가 두-PR 전제·plan 단계 금지·Step 3.3.5 를 담고 있다 |
| 돌연변이 프로브 | 옛 문장만 지우고 새 규칙을 쓰지 않는 변이는 긍정 앵커에서 죽는다. 한 사본만 고친 변이는 사본 일치 단언에서 죽는다 |
| 녹색 경로 | M4 (AC-GFD-013 의 amend 와 함께) |
| 통과 출력 | 위 grep 이 해당 경로에서 종료 코드 1, 긍정 앵커 grep 이 종료 코드 0 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-018 — main 과 develop 의 수렴 리허설 (maps REQ-GFD-018)

- **분류**: release-blocking (M6 종료 게이트)
- **Given** 스크래치 클론에서 develop 팁 `284e09c44023598affe486f17701717ca173e6ca` 와 main 팁 `4755c5e506225ba90b7a303c5763fa303c699492` 로 만든 두 로컬 브랜치
- **When** develop 이 main 을 흡수하고 그 결과를 main 에 병합 커밋으로 병합하면
- **Then** 병합 후 main 의 트리가 병합 직전 develop 팁의 트리와 같고, develop 팁이 main 의 조상이며, 병합 커밋을 되돌린 main 의 트리는 되돌리기 전 main 팁의 트리 `d57fdeecf4b8849da70a6693bbff19ad77ad5932` 와 같다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-11 (종료 코드 0, 두 팁의 트리가 다르다)과 E-22 (종료 코드 0, 발산 `7614	1`). 보조: E-23 (종료 코드 0, 흡수 병합이 충돌 없는 트리 `992f57f42aee6c3f9a39754c3e0095eaadf9d44b` 를 낸다) |
| RED 인 이유 | 두 트리가 다르고 main 에만 있는 커밋이 있어 빨리감기가 불가능하다 |
| 돌연변이 프로브 | squash 로 수렴하는 변이는 트리 항등은 통과하지만 "develop 팁이 main 의 조상" 에서 죽는다. 흡수 없이 병합하는 변이는 트리 항등에서 죽는다. 되돌리기를 시연하지 않는 변이는 되돌린 트리 단언에서 죽는다 |
| 녹색 경로 | M6 (스크래치 클론, GitHub 는 건드리지 않는다) |
| 통과 출력 | 세 비교 명령이 각각 같은 값을 출력하고 `git merge-base --is-ancestor` 가 종료 코드 0 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (원격 팁은 위 SHA 로 고정) |

### AC-GFD-019 — 배치 경계 사전 점검 (maps REQ-GFD-019)

- **분류**: release-blocking (M6 종료 게이트)
- **Given** 읽기 전용 사전 점검 스크립트와 부정 픽스처 넷 — 미푸시 develop 커밋 1개, 살아 있는 통합 창 또는 슬롯 보유자, develop 에 병합되지 않은 picked 카드 브랜치, 활성 레인 세션
- **When** 점검을 실행하면
- **Then** 모든 조건이 충족된 픽스처만 종료 코드 0 이고, 부정 픽스처는 각자 위반한 조건을 이름으로 대며 종료 코드가 0 이 아니다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-12 (종료 코드 0, 카드 트리가 develop 팁보다 122 커밋 앞서 있어 "푸시 완료" 조건이 지금 충족되지 않는다) |
| RED 인 이유 | 로컬 develop 에 미푸시분이 있어 지금은 경계가 아니다 |
| 돌연변이 프로브 | 발산만 보는 변이는 창·카드·세션 픽스처에서 죽는다. 항상 0 을 내는 변이는 부정 픽스처에서 죽는다 |
| 녹색 경로 | M6 |
| 통과 출력 | `go test -v -run '^TestCutoverPrecheck$' ./internal/...` 또는 동등 위치에서 하위 시험 5개 `--- PASS` |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-020 — 외부 공유 시스템 변경의 확인 경계 (maps REQ-GFD-020)

- **분류**: regression-guard (부재(하지 않았다)는 이 트리에서 다시 실행할 수 없는 관측이므로 통과로 기록하지 않는다)
- **Given** 이 SPEC 의 카드 작업이 끝난 시점
- **When** sync 감사가 `gh api repos/modu-ai/moai-adk/branches/main/protection`·`gh api repos/modu-ai/moai-adk/rulesets`·`gh api repos/modu-ai/moai-adk --jq .default_branch` 를 `research.md` §2 의 관측과 대조하면
- **Then** 값이 같거나, 다르면 리더의 확인 기록이 있다.

| 셀 | 내용 |
|---|---|
| 회귀 방지 신호 | 보호 규칙 `strict: false`·필수 체크 5개·`enforce_admins: true`, 규칙셋 1개, 기본 브랜치 `main`, develop 비보호 |
| 한계 | 값이 같다는 것은 "바뀐 증거가 없다"이지 "바꾸지 않았다"의 증명이 아니다. 확인 기록 파일 위치는 plan.md §E |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-021 — 스윕 가드는 실패를 관측해야 한다 (maps REQ-GFD-021)

- **분류**: release-blocking (M4 종료 게이트)
- **Given** 가드의 픽스처 시험(트리와 무관, 사전 병합 가능)과 아래 양방향 입력

| 픽스처 | 입력 | 기대 |
|---|---|---|
| F-red-1 | 살아 있는 문장 "카드 브랜치는 develop 에서 만든다" | 위반 |
| F-red-2 | CJK 인접 "develop에서 분기한다" | 위반 |
| F-red-3 | 표지 없는 펜스 셸 줄 `git switch -c x origin/develop` | 위반 |
| F-green-1 | 폐기 표지 줄 "RETIRED: card branches fork from develop" | 통과 |
| F-green-2 | 식별자 `manager-develop` | 통과 |
| F-green-3 | 디렉터리명 `.claude/worktrees/develop` | 통과 |
| F-green-4 | 동사 "to develop a feature" | 통과 |
| F-green-5 | 스탬프 `develop @ 0cca34439` | 통과 |
| F-green-6 | 구조 식별자 `--develop-worktree` | 통과(별도 차선에서 추적) |
| F-empty | 방문 파일 0 | 가드 실패(빈 스윕) |

- **When** `go test -v -run '^(TestGitHubFlowSweepFixtures|TestGitHubFlowSweepGuard)$' ./internal/template/` 를 실행하면
- **Then** 픽스처 10행 모두 기대대로이고 가드 본체는 방문 수를 출력하며 `--- SKIP` 이 없다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-13 (종료 코드 1, 가드 시험 부재) |
| RED 인 이유 | 가드가 없다. 시험 선택 `-run` 은 존재하지 않는 시험에서도 종료 코드 0 과 `[no tests to run]` 을 내므로 이 기준은 시험 이름이 아니라 방문 수와 `--- PASS` 줄을 요구한다 |
| 돌연변이 프로브 | 전부-일치 변이(항상 위반)는 F-green 행에서 죽고, 전부-불일치 변이(항상 통과)는 F-red 행에서 죽으며, 빈 스윕을 통과로 읽는 변이는 F-empty 에서 죽는다. CJK 를 `\b` 로 처리하는 변이는 F-red-2 에서 죽는다 |
| 녹색 경로 | M4 (픽스처 시험은 PRE-CUTOVER-SAFE 로 먼저 병합 가능, 트리 단언은 절체 시점에 무장) |
| 통과 출력 | 하위 시험 10행 `--- PASS`, 가드 본체 `visited=<N>` 출력과 `--- SKIP` 없음 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-022 — 병합 큐 판정과 t1452 조정의 기록 (maps REQ-GFD-022)

- **분류**: regression-guard (문서 존재와 후속 카드 기록은 이 트리에서 RED 를 다시 실행하는 형태가 아니다)
- **Given** `design.md` §D-3 의 병합 큐 증거 표와 §D-11 의 t1452 처분
- **When** sync 감사가 `progress.md` 의 후속 카드 id 와 `moai todo list` 를 대조하면
- **Then** 병합 큐 측정 후속 카드가 큐에 있고, t1452 의 (a)(b) 가 은퇴·(c) 가 유효로 분류돼 있다.

| 셀 | 내용 |
|---|---|
| 회귀 방지 신호 | 표의 미측정 행(청구 분, CodeRabbit 한도 이력, 의미 충돌 빈도)이 모두 후속 카드의 측정 항목에 있다 |
| 한계 | 병합 큐의 효과는 이 SPEC 에서 측정하지 않는다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

## §D 추적표

| AC | 대응 REQ | 분류 | 마일스톤 |
|---|---|---|---|
| AC-GFD-001 | REQ-GFD-001 | release-blocking | M1 |
| AC-GFD-002 | REQ-GFD-002 | release-blocking | M2 |
| AC-GFD-003 | REQ-GFD-003 | release-blocking | M1 |
| AC-GFD-004 | REQ-GFD-004 | release-blocking | M2 |
| AC-GFD-005 | REQ-GFD-005 | release-blocking | M2 |
| AC-GFD-006 | REQ-GFD-006 | release-blocking | M2 |
| AC-GFD-007 | REQ-GFD-007 | release-blocking | M3 |
| AC-GFD-008 | REQ-GFD-008 | release-blocking | M3 |
| AC-GFD-009 | REQ-GFD-009 | release-blocking | M3 |
| AC-GFD-010 | REQ-GFD-010 | release-blocking | M3 |
| AC-GFD-011 | REQ-GFD-011 | release-blocking | M4 |
| AC-GFD-012 | REQ-GFD-012 | release-blocking | M4 |
| AC-GFD-013 | REQ-GFD-013 | release-blocking | M4 |
| AC-GFD-014 | REQ-GFD-014 | regression-guard | M4 |
| AC-GFD-015 | REQ-GFD-015 | regression-guard | M1~M5 |
| AC-GFD-016 | REQ-GFD-016 | release-blocking | M5 |
| AC-GFD-017 | REQ-GFD-017 | release-blocking | M4 |
| AC-GFD-018 | REQ-GFD-018 | release-blocking | M6 |
| AC-GFD-019 | REQ-GFD-019 | release-blocking | M6 |
| AC-GFD-020 | REQ-GFD-020 | regression-guard | M6 |
| AC-GFD-021 | REQ-GFD-021 | release-blocking | M4 |
| AC-GFD-022 | REQ-GFD-022 | regression-guard | M2 |

## §E 이 SPEC 이 통과로 기록하지 않는 것

| 항목 | 이유 |
|---|---|
| 임베드 축 (`make build` 가 싣는 내용) | 리더가 배치 종료 때 빌드하므로 이 SPEC 의 관측이 아니다 |
| 병합 큐의 효과 | 판정(보류)만 하고 측정은 후속 카드 |
| GitHub 쪽 PR 생성·대형 diff 병합 | 스크래치 클론 리허설은 git 층만 시연한다. GitHub 층은 리더·운영자가 절체 경계에서 관측한다 |
| rc 태그의 실제 릴리스 | 태그는 불변 규칙셋으로 되돌릴 수 없어 이 카드에서 만들지 않는다 |
| 상시로드 비용의 토큰 환산 | 호스트가 관측을 노출하지 않으면 바이트만 보고한다 |

---

🗿 MoAI

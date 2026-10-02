# SPEC-GITHUB-FLOW-DEFAULT-001 — 인수 기준

> 문서 수준 트리 핀: 카드 트리 `4bf547bcad7c155b1e91485921569db709ec3ac2`. 이 핀은 자체 핀이 없는 모든 기준에 적용되고, 기준 수준 핀이 있으면 그것이 이긴다. 브랜치 이름은 핀으로 쓰지 않는다. 개정 0.1.1 에서 더하거나 바꾼 원장 항목(E-26 이후)과 측정 행(M-1~M-7)은 계획 커밋 `855563dba79da74528e0f01560efe80f1016cc11` 에서 측정했다 — 그 항목을 인용하는 기준은 그 핀을 자기 핀으로 적는다. 두 트리의 차이는 SPEC 디렉터리의 6개 파일뿐이다(`git diff --stat 4bf547bcad7c155b1e91485921569db709ec3ac2 855563dba79da74528e0f01560efe80f1016cc11` 이 6개 파일을 나열했다). **개정 0.1.2 에서 더하거나 다시 잰 원장 항목(E-36 이후)과 측정 행(M-1·M-3·M-6·M-9)은 계획 개정 2 커밋 `6c2277295d9ddeaa92e83c0225910445f83cc1c2` 에서 측정했다** — 이 커밋과 위 두 트리의 차이도 SPEC 디렉터리의 6개 파일뿐이다(`git diff --name-only 855563dba79da74528e0f01560efe80f1016cc11 6c2277295d9ddeaa92e83c0225910445f83cc1c2` 가 같은 6개 파일을 나열했다). 그 항목을 인용하는 기준은 그 핀을 자기 핀으로 적는다. **개정 0.1.3 에서 더한 원장 항목(E-46~E-55)과 측정 행(M-11)은 카드 트리 HEAD `eda61419564296892b41d0e7608c8b136e52c04c` 에서 이 회차에 쟀다** — 이 트리는 run-phase M1 커밋과 develop 흡수를 담고 있어 앞 세 트리와 같다고 주장하지 않으며, 그 항목을 인용하는 기준(AC-GFD-024)은 이 핀을 자기 핀으로 적는다.

## §A 측정 규약과 분류

- **RED-now 셀**은 네 요소를 함께 담는다 — 명령(읽기 전용 단일 호출), 그 명령의 verbatim stdout(원본 파일 바이트), 종료 코드(별도 필드), 트리 SHA. 네 요소는 아래 **증거 원장**(§B)에 항목 ID 로 기록하고 기준 표가 ID 를 인용한다. 표 셀은 셸 메타문자를 훼손하므로 원장이 운반체다.
- **녹색 경로 셀**은 어느 마일스톤이 기준을 뒤집는지와 통과 출력의 모양을 적는다.
- **분류**: `release-blocking`(절체 변경 묶음의 병합 게이트)과 `regression-guard`. RED 를 이 트리에서 다시 실행할 수 없는 기준은 `regression-guard` 로 분류되고 통과로 기록되지 않는다.
- **돌연변이 프로브**: 요구를 어기면서 기준을 만족하는 변이를 쓸 수 있으면 그 기준은 너무 얕다. 기준마다 죽여야 할 변이와 그것을 죽이는 입력을 적는다. 무효 입력만 보는 기준은 전부-일치 변이를, 유효 입력만 보는 기준은 전부-불일치 변이를 통과시키므로 양방향 입력을 함께 둔다.
- **빈 스윕 금지**: 모든 시험 선택은 `-run` 을 앵커로 고정하고 실행된 시험 수를 확인한다. `[no tests to run]`·`no test files` 는 통과가 아니다.
- **연속 발화**: 가드 시험은 `SKIP` 으로 끝날 수 있으면 완료가 아니다. 통과 출력은 `--- PASS` 줄과 방문 수를 포함해야 하고 `--- SKIP` 이 없어야 한다.

## §B 증거 원장 (RED-now 4요소)

E-01~E-25 의 트리는 `4bf547bcad7c155b1e91485921569db709ec3ac2` 이고 E-26~E-35 는 `855563dba79da74528e0f01560efe80f1016cc11`, E-36 이후는 항목마다 적은 `6c2277295d9ddeaa92e83c0225910445f83cc1c2` 이다. 종료 코드는 같은 명령을 `; echo "exit=$?"` 를 붙여 따로 관측한 값이다(셀의 명령에는 붙이지 않는다). 1회차 감사가 E-01~E-25 를 이 트리에서 모두 다시 실행해 기록과 같음을 확인했으므로 개정에서 바꾸지 않았다(E-03 의 역할만 맥락 행으로 바뀐다). **측정 행 M-n 은 RED 셀이 아니라 방법·출력·날짜를 적은 맥락 행이다** — 일부는 파이프나 스크래치 프로그램을 쓰므로 RED-now 의 단일 호출 형태가 아니다.

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
note: 맥락 행이다. AC-GFD-001·005 의 RED 는 이 다섯 줄 중 카드 전달의 기준을 읽는 두 줄(E-28)이다. integration.go:353 과 session_worktree_automerge.go:176,183 은 github-flow 에서 휴면이 맞는 git-flow 전용 경로이며 통합 목표 해석기로 옮기지 않는다(design D-1)
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

### 개정 0.1.1 에서 더한 원장 항목 (트리 `855563dba79da74528e0f01560efe80f1016cc11`)

```
E-26
command: git grep -l TestCutoverPrecheck -- internal scripts
stdout:
exit: 1
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
E-27
command: git grep -l -i -E "cutover-rehearsal|TestCutoverRehearsal" -- internal scripts
stdout:
exit: 1
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
E-28
command: git grep -n "\.DevelopBranch" -- internal/cli/factory_card.go internal/cli/factory_merge.go
stdout:
internal/cli/factory_card.go:1064:	if branch := strings.TrimSpace(config.LoadGitFlowIntegrationConfig(root).DevelopBranch); branch != "" {
internal/cli/factory_merge.go:93:				developRef = config.LoadGitFlowIntegrationConfig(integrationLockRoot()).DevelopBranch
exit: 0
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
E-29
command: git grep -n -o -E 'cardBaseBranch +[=] +"develop"|MergeTarget: "develop"' -- internal/cli/codex_review_scope.go internal/cli/goal.go
stdout:
internal/cli/codex_review_scope.go:55:cardBaseBranch   = "develop"
internal/cli/goal.go:326:MergeTarget: "develop"
exit: 0
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
E-30
command: git grep -n -E "^[[:space:]]+(worktree_base_branch|workflow): " -- internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl
stdout:
internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl:12:  worktree_base_branch: ""
internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl:18:    workflow: github-flow
internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl:56:    workflow: github-flow
internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl:92:    workflow: github-flow
exit: 0
tree: 855563dba79da74528e0f01560efe80f1016cc11
note: "배포 기본값은 이미 github-flow" 라는 서술(spec.md §A.2)의 증거는 이 항목이다. E-08 은 이 저장소의 파일이지 배포 템플릿이 아니다
```

```
E-31
command: git grep -n -F -e "--require-matrix-run" -- scripts/release.sh .claude/agents/harness/hns-release-specialist.md
stdout:
exit: 1
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
E-32
command: git grep -l -F "TestCutoverRunbookShape" -- internal scripts
stdout:
exit: 1
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
E-33
command: moai constitution list --file .claude/rules/moai/workflow/worktree-integration.md
stdout:
No entries.
exit: 0
tree: 855563dba79da74528e0f01560efe80f1016cc11
build: 이 트리 HEAD 에서 빌드한 moai (설치본 v3.2.0-rc.25 는 이 HEAD 의 엄격한 조상 802a72235 라서 쓰지 않았다). 설치본으로도 같은 출력("No entries.")을 관측했다
```

```
E-34
command: git grep -n -E "^- id: CONST-V3R5-02[78]$" -- .claude/rules/moai/core/zone-registry.md
stdout:
.claude/rules/moai/core/zone-registry.md:837:- id: CONST-V3R5-027
.claude/rules/moai/core/zone-registry.md:845:- id: CONST-V3R5-028
exit: 0
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
E-35
command: moai constitution validate
stdout:
constitution validate: OK — no drift or violations detected (97 of 101 entries checked)

  4 retired entry/entries skipped ([SUPERSEDED …] marker); re-check them with --strict
exit: 0
tree: 855563dba79da74528e0f01560efe80f1016cc11
build: 이 트리 HEAD 에서 빌드한 moai. 이 행은 개정 전 기준선이며 RED 가 아니다(AC-GFD-013 의 녹색 경로가 개정 뒤 같은 종료 코드 0 을 요구한다)
```

### 개정 0.1.2 에서 더한 원장 항목 (트리 `6c2277295d9ddeaa92e83c0225910445f83cc1c2`)

이 항목의 `moai` 호출은 이 트리 HEAD 에서 `go build -o <스크래치>/bin/moai ./cmd/moai`(종료 코드 0)로 새로 빌드한 바이너리를 경로로 불렀다 — 설치본(v3.2.0-rc.25, 커밋 802a72235)은 이 HEAD 의 엄격한 조상이라 쓰지 않았다. ldflags 없이 빌드해 `moai version` 이 `v3.1.3`·커밋 `none` 을 찍으므로 빌드 커밋은 빌드한 트리 HEAD 로 귀속한다.

```
E-36
command: git grep -n -F "NO worktree at this step" -- .claude/rules/moai/workflow/spec-workflow.md
stdout:
.claude/rules/moai/workflow/spec-workflow.md:164:[ZONE:Frozen] [HARD] Execute in main checkout. NO worktree at this step. See § SPEC Phase Discipline (Step 1).
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
```

```
E-37
command: git grep -n -F "NO worktree at this step" -- internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md:164:[ZONE:Frozen] [HARD] Execute in main checkout. NO worktree at this step. See § SPEC Phase Discipline (Step 1).
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
```

```
E-38
command: git grep -n "ZONE:Frozen" -- .claude/rules/moai/workflow/spec-workflow.md
stdout:
.claude/rules/moai/workflow/spec-workflow.md:23:[ZONE:Frozen] [HARD] Every MoAI SPEC follows the three-phase lifecycle (plan → run → sync). How each phase transition is *triggered* depends on the **route** the SPEC takes. There are exactly TWO routes, and the route is determined by Tier (per § SPEC Complexity Tier) and the explicit `--pr` flag:
.claude/rules/moai/workflow/spec-workflow.md:49:[ZONE:Frozen] [HARD] Step ordering rules:
.claude/rules/moai/workflow/spec-workflow.md:164:[ZONE:Frozen] [HARD] Execute in main checkout. NO worktree at this step. See § SPEC Phase Discipline (Step 1).
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
```

```
E-39
command: git grep -n "ZONE:Frozen" -- internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
stdout:
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md:23:[ZONE:Frozen] [HARD] Every MoAI SPEC follows the three-phase lifecycle (plan → run → sync). How each phase transition is *triggered* depends on the **route** the SPEC takes. There are exactly TWO routes, and the route is determined by Tier (per § SPEC Complexity Tier) and the explicit `--pr` flag:
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md:49:[ZONE:Frozen] [HARD] Step ordering rules:
internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md:164:[ZONE:Frozen] [HARD] Execute in main checkout. NO worktree at this step. See § SPEC Phase Discipline (Step 1).
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
```

```
E-40
command: moai constitution list --file .claude/rules/moai/workflow/spec-workflow.md
stdout:
ID                  Zone        File                                                Clause                                  
----------------------------------------------------------------------------------------------------------------------------
CONST-V3R2-001      Frozen      .claude/rules/moai/workflow/spec-workflow.md        Create comprehensive specification us...
CONST-V3R5-027      Frozen      .claude/rules/moai/workflow/spec-workflow.md        Step 1 (plan) MUST execute in main ch...
CONST-V3R5-028      Frozen      .claude/rules/moai/workflow/spec-workflow.md        Step 4 (cleanup) applies to **Route B...

Total 3 entries
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
note: 이 세 clause 의 본문 줄은 CONST-V3R2-001 = 166 줄(`Create comprehensive specification using EARS format.`), CONST-V3R5-027 = 50 줄, CONST-V3R5-028 = 53 줄이다. 어느 clause 도 줄 23·49·164 의 표지 줄 자체가 아니다
```

```
E-41
command: moai constitution list --file internal/template/templates/.claude/rules/moai/workflow/spec-workflow.md
stdout:
No entries.
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
note: 레지스트리는 로컬 경로만 안다. 템플릿 사본의 Frozen 표지 줄은 모두 미등재로 읽히며 AC-GFD-015 의 사본 정합 시험이 로컬과 같게 지킨다
```

```
E-42
command: git grep -n -F "verify-release-provenance" -- .github/workflows/release.yml
stdout:
exit: 1
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
```

```
E-43
command: git diff --name-only 4bf547bcad7c155b1e91485921569db709ec3ac2 6c2277295d9ddeaa92e83c0225910445f83cc1c2 -- AGENTS.md AGENTS.local.md CLAUDE.md README.md README.ko.md README.ja.md README.zh.md .claude docs-site/content .moai/docs internal/template/templates .moai/project/codemaps .moai/config .github .coderabbit.yaml
stdout:
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
note: 가드 행(도착 시점에 이미 참)이다 — AC-GFD-016 의 사전 방향이 금지하는 산문 표면·설정·CI 경로가 카드 트리 `4bf547bca…` 에서 이 HEAD 까지 바뀌지 않았다. 이 기준은 경로 목록이 이 명령의 상위집합이어야 한다
```

```
E-44
command: git grep -n -F "all 7 checks passed" -- .github/workflows/release.yml
stdout:
.github/workflows/release.yml:123:          echo "RELEASE_PROVENANCE_GATE: all 7 checks passed for ${TAG} (${TAG_COMMIT})."
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
note: 변경 전 인라인 단계의 성공 문구다. 접미사 없는 태그의 동치 시험이 이 문구를 기준으로 쓴다(design D-6 의 배선 계약)
```

```
E-45
command: git grep -n -w "Develop" -- AGENTS.md AGENTS.local.md CLAUDE.md README.md README.ko.md README.ja.md README.zh.md .claude/rules .claude/agents .claude/skills .claude/output-styles docs-site/content .moai/docs internal/template/templates .moai/project/codemaps scripts
stdout:
.claude/skills/moai-foundation-cc/reference/claude-code-skills-official.md:282:### Develop Skills Iteratively with Claude
.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:667:1. Develop systematic troubleshooting approach
.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:1284:1. Develop comprehensive security improvement plan
.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md:957:- Solution Templates: Develop reusable solution templates
.claude/skills/moai-workflow-worktree/modules/parallel-advanced.md:23:Develop feature flags across multiple worktrees for comprehensive testing.
.claude/skills/moai-workflow-worktree/modules/parallel-advanced.md:259:Develop in Worktree:
docs-site/content/en/worktree/examples.md:261:# Develop in the Worktree
docs-site/content/en/worktree/examples.md:284:# Develop in the Worktree
docs-site/content/en/worktree/faq.md:570:        A2[Develop]
docs-site/content/en/worktree/faq.md:576:        B2[Develop]
internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-skills-official.md:294:### Develop Skills Iteratively with Claude
internal/template/templates/.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:667:1. Develop systematic troubleshooting approach
internal/template/templates/.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-examples.md:1284:1. Develop comprehensive security improvement plan
internal/template/templates/.claude/skills/moai-foundation-cc/reference/sub-agents/sub-agent-formatting-guide.md:957:- Solution Templates: Develop reusable solution templates
internal/template/templates/.claude/skills/moai-workflow-worktree/modules/parallel-advanced.md:23:Develop feature flags across multiple worktrees for comprehensive testing.
internal/template/templates/.claude/skills/moai-workflow-worktree/modules/parallel-advanced.md:259:Develop in Worktree:
exit: 0
tree: 6c2277295d9ddeaa92e83c0225910445f83cc1c2
note: 맥락 행이다 — 대문자 `Develop` 은 동사·제목·다이어그램 라벨 열여섯 줄뿐이고 브랜치 서술은 없다(design D-9 의 대소문자 한계의 근거)
```

### 개정 0.1.3 에서 더한 원장 항목 (트리 `eda61419564296892b41d0e7608c8b136e52c04c`)

개정 A-1(spec.md `## Amendments`)이 더한 항목이다. 아래 명령은 모두 이 트리에서 읽기 전용 단일 호출로 한 번씩 실행했고 종료 코드는 호출 직후 따로 관측한 값이다(셀의 명령에는 붙이지 않는다). E-46·E-47 이 AC-GFD-024 의 RED 셀이고 나머지는 범위 경계와 동결 표면을 보이는 맥락 행·가드 행이다.

```
E-46
command: git grep -n -F '"develop"' -- internal/contract/projection_mission.go
stdout:
internal/contract/projection_mission.go:82:	missionMergeTarget        = "develop"
exit: 0
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 이 줄의 상수가 같은 파일 222 줄에서 `MergeTarget: missionMergeTarget` 으로 투영된 미션 계약의 병합 목표가 된다(`git grep -n -F missionMergeTarget -- internal/contract/projection_mission.go` 가 82·222 두 줄을 냈다)
```

```
E-47
command: git grep -l -F 'TestProjectionMergeTarget' -- internal
stdout:
exit: 1
tree: eda61419564296892b41d0e7608c8b136e52c04c
```

```
E-48
command: git grep -n -F 'ProjectToMission(' -- '*.go' ':!*_test.go'
stdout:
internal/contract/projection_mission.go:173:func ProjectToMission(c *Contract) (mission.MissionContract, error) {
exit: 0
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 맥락 행이다 — 비테스트 호출자가 없다(정의 한 줄뿐). 호출 지점은 시험 파일 하나의 네 곳이다(`git grep -c -F 'ProjectToMission(' -- internal/contract/projection_mission_test.go` 가 `internal/contract/projection_mission_test.go:4`, 종료 코드 0). 그래서 인자를 더해도 운영 호출 경로가 움직이지 않는다
```

```
E-49
command: git grep -n -F '.MergeTarget' -- '*.go' ':!*_test.go'
stdout:
internal/mission/contract.go:58:		strings.TrimSpace(c.MergeTarget) != "" && c.ResourceLimits.MaxOperations > 0 && c.ResourceLimits.MaxRetries >= 0 &&
exit: 0
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 맥락 행이다 — 비테스트 코드에서 `MergeTarget` 을 읽는 곳은 봉인 전 완결성 검사의 비어 있지 않음 확인 한 줄뿐이다. 값은 계약 해시에 들어가지만(`internal/mission/contract.go` 의 `SealMissionContract` 가 정규화한 계약 JSON 을 해시한다) 어떤 브랜치 이름과도 비교되지 않는다. 구성 쪽 생산자는 둘이다 — `internal/cli/goal.go` 의 approve(M1 이 `IntegrationTarget` 으로 옮겼다, 원장 E-29)와 이 투영. 서로의 값을 받지 않는 별개의 생산 지점이다
```

```
E-50
command: git grep -n -F 'branch != "develop"' -- internal/mission/git_owner.go
stdout:
internal/mission/git_owner.go:201:		if branch != "develop" || !strings.HasPrefix(o.Effect.WorktreeBranch, "WT-") || o.Effect.CardSHA == "" || o.Effect.BaseSHA == "" {
exit: 0
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 맥락 행이다 — 이 줄은 `ActionLocalMerge`(동결된 행동 식별자 `local_develop_merge`)의 로컬 병합 경계이며 현재 브랜치가 리터럴 `develop` 이 아니면 `merge boundary invalid` 로 거부한다. 읽기만 했고 실행해 관측하지 않았다. 범위 밖이다(design D-30)
```

```
E-51
command: git grep -n 'develop' -- internal/contract/testdata/mission_surface_baseline.txt
stdout:
internal/contract/testdata/mission_surface_baseline.txt:45:	ActionLocalMerge    Action = "local_develop_merge"
exit: 0
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 가드 행(도착 시점에 이미 참)이다 — 고정된 스냅숏에서 `develop` 이 든 줄은 동결된 행동 식별자 값 한 줄뿐이고, `MergeTarget` 은 줄 347 에 필드 선언(`MergeTarget        string         `json:"merge_target"``)으로만 있다. 그래서 투영의 병합 목표 값 변경은 이 스냅숏을 바꾸지 않는다(`go doc -all ./internal/mission` 은 선언과 문서 주석만 담는다). 변경 뒤에도 이 출력은 바이트 단위로 같아야 한다
```

```
E-52
command: git grep -n -F 'moai-adk/internal/config"' -- internal/contract
stdout:
exit: 1
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 가드 행이다 — 계약 코어는 `internal/config` 를 import 하지 않는다(`internal/contract/projection_mission.go` 가 import 하는 내부 패키지는 `internal/mission` 하나다, `git grep -n 'moai-adk/internal/' -- internal/contract/projection_mission.go` 가 `projection_mission.go:10` 의 `internal/mission` 한 줄만 냈고 종료 코드 0 이었다). 구성은 호출자가 읽어 인자로 넘긴다 — design D-30 의 인자 방식의 근거
```

```
E-53
command: git grep -n -F -e 'local-merge-develop' -e 'push-develop' -- internal/contract/rules.go internal/contract/projection_mission.go
stdout:
internal/contract/projection_mission.go:76:	"local-merge-develop": mission.ActionLocalMerge,
internal/contract/rules.go:13:const ActionPushDevelop = "push-develop"
internal/contract/rules.go:16:	allowedActions   = []string{"commit", "worktree", "local-merge-develop", ActionPushDevelop}
exit: 0
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 가드 행이다 — 동결된 계약 쪽 행동 식별자의 사용 줄 셋이다. 변경 뒤에도 줄 문언은 같아야 한다(줄 번호는 편집이 옮기므로 인용은 문언으로 한다)
```

```
E-54
command: git diff --name-only 4bf547bcad7c155b1e91485921569db709ec3ac2 eda61419564296892b41d0e7608c8b136e52c04c -- internal/mission internal/contract
stdout:
exit: 0
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 가드 행이다 — 카드 트리에서 이 HEAD 까지(M1 커밋과 develop 흡수 포함) `internal/mission` 과 `internal/contract`(스냅숏 포함)의 어느 파일도 바뀌지 않았다. 빈 출력이 종료 코드 0 과 함께 나왔으므로 명령 형태가 차이를 보이는지는 E-55 가 양성 대조로 확인한다
```

```
E-55
command: git diff --name-only 4bf547bcad7c155b1e91485921569db709ec3ac2 eda61419564296892b41d0e7608c8b136e52c04c -- internal/cli/goal.go
stdout:
internal/cli/goal.go
exit: 0
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: E-54 의 양성 대조다 — 같은 형태의 명령이 M1 이 고친 파일은 보인다
```

### 측정 행 (맥락 행 — RED 셀이 아니다, 날짜와 방법을 적는다)

개정 0.1.2 는 CI 행을 **변하지 않는 식별자로 고정한 값**으로 바꿨다(design D-28). 같은 `gh run list` 명령이 호출마다 다른 창을 돌려주는 것이 관측됐으므로 목록 호출의 출력은 인용하지 않는다.

```
M-1  (2026-10-02, 실행 id 열두 개로 고정 — 각 id 를 하나씩 다시 읽었다. 트리 6c2277295d9ddeaa92e83c0225910445f83cc1c2 와 무관한 GitHub 쪽 값)
command: gh run view <id> --json databaseId,conclusion,createdAt,startedAt,updatedAt,headSha,workflowName,event,headBranch   (열두 id 에 각각 한 번씩)
stdout (id 별 conclusion · createdAt · startedAt→updatedAt 초 · headSha 앞 9자, 전부 workflowName=CI event=push headBranch=develop):
36989864262  failure    2026-10-02T09:27:17Z  1046  284e09c44
36981574917  failure    2026-10-02T08:00:12Z  1223  802a72235
36969555332  failure    2026-10-02T05:34:37Z  1413  c50da9c2f
36894416006  failure    2026-10-01T16:45:47Z  1484  6789de39b
36892129460  cancelled  2026-10-01T16:27:27Z  -     2bb8829de
36890528950  cancelled  2026-10-01T16:14:47Z  -     e83aa47df
36887673351  cancelled  2026-10-01T15:52:37Z  -     89e3164af
36886926130  cancelled  2026-10-01T15:46:49Z  -     a301529bb
36883300773  failure    2026-10-01T15:19:08Z  1517  f28cce310
36882533500  cancelled  2026-10-01T15:13:14Z  -     a9f43a6fc
36876598542  failure    2026-10-01T14:28:27Z  1551  a466f5585
36870179597  failure    2026-10-01T13:38:32Z  1443  678ca0c27
derived: failure 7·cancelled 5·success 0, createdAt 2026-10-01T13:38:32Z ~ 2026-10-02T09:27:17Z, failure 초 평균 1382 (= 9677 / 7)
note: 열두 id 는 `.moai/reports/t1453/m0-develop-ci-red.md` 의 표에서 왔고 그 파일은 인용 대상이 아니다 — 같은 열두 id 를 research.md §4.1 이 표로 옮겼다. 첫 id 의 headSha 는 develop 팁 핀 284e09c44023598affe486f17701717ca173e6ca 와 같다
```

```
M-2  (인용하지 않음 — 0.1.2 가 뺐다)
note: 0.1.1 의 main push CI 200건 목록 표본(success 146·cancelled 43·failure 10·startup_failure 1, success 평균 438 초)은 200개 실행을 id 로 고정하지 못했고 감사 회차에 다른 결과(112/45/43)와 HTTP 502 를 돌려주었다. 이 SPEC 의 어떤 기준도 이 수에 기대지 않는다 — research.md §9 Gaps
```

```
M-3  (2026-10-02, PR 번호 스무 개로 고정 — 스무 PR 을 번호별로 읽었다)
command: gh api graphql -f query=<스무 PR 을 번호별 별칭 `pullRequest(number:N){number baseRefName state mergedAt}` 로 읽는 질의 한 번>
stdout (번호:baseRefName:state):
1695:main:MERGED 1699:develop:MERGED 1702:main:MERGED 1713:develop:MERGED 1719:develop:MERGED 1720:develop:MERGED 1721:develop:MERGED 1723:develop:MERGED 1724:develop:MERGED 1725:develop:MERGED 1726:develop:MERGED 1727:develop:MERGED 1728:develop:MERGED 1729:develop:MERGED 1740:main:MERGED 1741:develop:MERGED 1742:develop:MERGED 1743:develop:MERGED 1744:develop:MERGED 1745:develop:MERGED
derived: develop 17·main 3(1695·1702·1740), 스무 건 모두 MERGED
note: 스무 번호는 같은 날 `gh pr list --state merged --limit 20 --json number,baseRefName,mergedAt` 한 번이 낸 번호 목록이며 목록 호출 자체는 인용하지 않는다(번호는 불변이고 병합된 PR 의 baseRefName 은 병합 뒤 바뀌지 않는다)
```

```
M-4  (파일 수 — 스크래치 스크립트 `git ls-files | wc -l`, 이 트리에서 한 번)
stdout:
7  AGENTS.md AGENTS.local.md CLAUDE.md README.md README.ko.md README.ja.md README.zh.md
113  .claude/rules/*.md
22  .claude/agents/*.md
255  .claude/skills/*.md
3  .claude/output-styles/*.md
620  docs-site/content/*.md
37  .moai/docs/*.md
408  internal/template/templates/*.md internal/template/templates/*.md.tmpl
6  .moai/project/codemaps/*.md
54  scripts/*.sh scripts/*.py
1525  TOTAL (unique)
```

```
M-5  (개정 0.1.1 의 스윕 패턴 프로브 — 약한 적중을 개수로만 보고하던 21행 설계의 기록이다. 개정 0.1.2 의 설계(약한 적중도 위반)와 25행은 M-9 가 대체한다. 스크래치 Go 프로그램 `go -C <스크래치>/probe run .`, 소스는 저장소 밖에 있고 커밋하지 않았다)
stdout (픽스처 20행 + 빈 스윕 규칙 줄, 그리고 `AGENTS.md` 를 입력으로 한 실행):
OK  F-red-1    want_violation=true got_violation=true (P-B token)
OK  F-red-2    want_violation=true got_violation=true (P-D cjk-adjacent)
OK  F-red-3    want_violation=true got_violation=true (P-A fenced)
OK  F-red-4    want_violation=true got_violation=true (P-A origin/refs)
OK  F-red-5    want_violation=true got_violation=true (P-B token)
OK  F-red-6    want_violation=true got_violation=true (P-B token)
OK  F-red-7    want_violation=true got_violation=true (P-A origin/refs)
OK  F-red-8    want_violation=true got_violation=true (P-D cjk-adjacent)
OK  F-red-9    want_violation=true got_violation=true (P-B token)
OK  F-red-10   want_violation=true got_violation=true (P-B token)
OK  F-red-11   want_violation=true got_violation=true (P-B token)
OK  F-green-1  want_violation=false got_violation=false (line-marker)
OK  F-green-9  want_violation=false got_violation=false ()
OK  F-green-2  want_violation=false got_violation=false ()
OK  F-green-3  want_violation=false got_violation=false ()
OK  F-green-4  want_violation=false got_violation=false (verb)
OK  F-green-5  want_violation=false got_violation=false (stamp)
OK  F-green-6  want_violation=false got_violation=false ()
OK  F-green-7  want_violation=false got_violation=false ()
OK  F-green-8  want_violation=false got_violation=false (section-marker)
F-empty    visited=0 lines=0 -> guard must FAIL (design rule, not exercised here)
fixtures=20 bad=0
AGENTS.md 입력:
76: `workflow.branch_guard.deny_commits_on`), card work flows on develop-based worktrees, and [P-B token]
145: **Start a new card in a new worktree from local `develop`.** A Claude Code session exits its [P-A backtick]
147: tree's HEAD equals the local `develop` tip before editing; never reuse the previous card's tree. [P-A backtick]
149: worktree. When complete, merge the card branch into local `develop` through the serial integration [P-A backtick]
150: window; the factory leader pushes `develop` after the local merges. [P-A backtick]
flagged=5 of 332 lines
note: 프로브가 원장 E-07 의 5줄(76·145·147·149·150)을 모두 잡는다. 1회차 설계의 패턴은 76 줄(`develop-based`)을 놓쳤고 이번 프로브의 첫 실행에서 그 구멍이 관측됐다(4줄만 적중) — 오른쪽 경계에 `-based` 등을 더해 고쳤다
```

```
M-6  (2026-10-02, develop 팁 SHA 로 고정 — 트리 6c2277295d9ddeaa92e83c0225910445f83cc1c2)
command: git rev-list --count --first-parent --since=2026-09-25T00:00:00Z --until=2026-10-02T00:00:00Z 284e09c44023598affe486f17701717ca173e6ca
stdout:
248
exit: 0
note: 0.1.1 은 같은 명령을 움직이는 `origin/develop` 에 걸었다. 이 회차에 SHA 로 고정한 형태가 `248` 을 냈고 `origin/develop`(그 사이 `d1c84a19cf317a903703051e8f6533378ab6e830` 으로 움직였다)에 건 같은 명령도 `248` 이었다. 같은 창의 모든 커밋 수(`1856`)는 0.1.1 의 기록이며 이 회차에 다시 재지 않았다 — 어떤 결정도 이 수에 기대지 않는다
```

```
M-7
command: git grep -l -w develop -- .github/workflows
stdout:
.github/workflows/ci.yml
.github/workflows/codeql.yml
.github/workflows/docs-i18n-check.yml
.github/workflows/graph-freshness.yml
.github/workflows/judgment-first-consistency.yaml
.github/workflows/lsel-leak-guard.yaml
.github/workflows/spec-lint.yml
.github/workflows/template-neutrality-check.yaml
.github/workflows/test-install.yml
.github/workflows/workflow-parse-guard.yaml
exit: 0
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
M-8  (AC-GFD-015 의 기준선 — 개정 전 이 트리의 네 mirror guard)
command: go test -v -count=1 -run '^(TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestLateBranchTemplateMirror|TestSanitizedPairParity)$' ./internal/template/
stdout (`--- PASS` 줄만 옮겼다, 전체 출력은 스크래치에 있다):
--- PASS: TestDeclaredRuleMirrorForks (0.00s)
--- PASS: TestLateBranchTemplateMirror (0.00s)
--- PASS: TestRuleTemplateMirrorDrift (0.00s)
--- PASS: TestSanitizedPairParity (0.00s)
exit: 0 (환경변수 정리 뒤 한 번의 복합 호출 `unset MOAI_KANBAN … && go test …`, 마지막 줄 `ok  	github.com/modu-ai/moai-adk/internal/template	0.494s`)
tree: 855563dba79da74528e0f01560efe80f1016cc11
```

```
M-9  (개정 0.1.2 설계 — 약한 적중도 허용 목록 외에는 위반 — 의 스크래치 프로브. 소스는 저장소 밖 스크래치(`probe3`)에 있고 커밋하지 않았다. 호출: `go -C <스크래치>/probe3 run .` (픽스처), `go -C <스크래치>/probe3 run . <표면 파일 목록> <트리 루트>` (표면 스캔). 트리 6c2277295d9ddeaa92e83c0225910445f83cc1c2)
stdout (픽스처 25행 — 허용 목록은 픽스처마다 따로 준다):
OK  F-red-1    want_violation=true got_violation=true (strong)
OK  F-red-2    want_violation=true got_violation=true (strong)
OK  F-red-3    want_violation=true got_violation=true (strong)
OK  F-red-4    want_violation=true got_violation=true (strong)
OK  F-red-5    want_violation=true got_violation=true (strong)
OK  F-red-6    want_violation=true got_violation=true (strong)
OK  F-red-7    want_violation=true got_violation=true (strong)
OK  F-red-8    want_violation=true got_violation=true (strong)
OK  F-red-9    want_violation=true got_violation=true (strong)
OK  F-red-10   want_violation=true got_violation=true (strong)
OK  F-red-11   want_violation=true got_violation=true (strong)
OK  F-red-12   want_violation=true got_violation=true (weak)
OK  F-red-13   want_violation=true got_violation=true (weak)
OK  F-red-14   want_violation=true got_violation=true (weak)
OK  F-green-1  want_violation=false got_violation=false ()
OK  F-green-2  want_violation=false got_violation=false ()
OK  F-green-3  want_violation=false got_violation=false ()
OK  F-green-4  want_violation=false got_violation=false ()
OK  F-green-5  want_violation=false got_violation=false ()
OK  F-green-6  want_violation=false got_violation=false ()
OK  F-green-7  want_violation=false got_violation=false ()
OK  F-green-8  want_violation=false got_violation=false ()
OK  F-green-9  want_violation=false got_violation=false ()
OK  F-green-10 want_violation=false got_violation=false ()
OK  F-empty    visited=0 guard_fails=true
fixtures=25 bad=0
stdout (회피 프로브 — 0.1.1 설계에서는 넷 모두 "약한 적중=녹색"이었다):
"Work starts from develop."                  violation=true (weak)
"Lane PRs go to develop."                    violation=true (weak)
"Commit on develop."                         violation=true (weak)
"Compare with the develop tip."              violation=true (weak)
"Develop branch is the integration line."    violation=false (nomatch)
stdout (표면 스캔 요약 — 표면 파일 목록은 아래 열 곳의 `git ls-files` 를 합쳐 중복 제거한 1523개):
files=1523 lines=328055 strong_violations=286 weak(now violations unless allow-listed)=31 exempt=9 weakFiles=25
note: 위반 286·약한 적중 31·표지 면제 9 는 2회차 감사가 같은 알고리즘을 따로 구현해 낸 수와 같다(감사 보고서 `plan-audit-iter2.md`). 약한 적중 31줄에 대한 제 분류(의견이며 측정이 아니다): 살아 있는 기준 서술 16줄 — `goal.md` "the leased local develop"(로컬·템플릿 2), `scripts/jev/triage.py`·`test_triage.py` "ancestor of develop"(2), `scripts/ac-baseline/check-armed.sh`·`install-hook.sh` "until … absorbs develop"(2), `.claude/rules/local/gitflow-lane-protocol.md` 1, `.moai/docs/` 9; 동사·키워드·중괄호 8줄(`Create and develop foundation SPEC`, `[…, "develop", …]` 키워드 목록 둘, `manager-{spec,develop,docs}`, 각 로컬·템플릿 쌍); 이동 참조 예시 2줄(`git fetch origin develop`); git-flow 옵션 서술 2줄(`manager-git.md` 의 `develop_direct`, 쌍); docs-site 3줄은 읽지 않았다. 스캔의 섹션·표지·펜스 처리는 제 재구현이라 실제 가드의 값이 아니다
```

```
M-10  (AC-GFD-008 의 접미사 없는 태그 동치 기준선 — 변경 전 인라인 출처 검증 단계를 스크래치 저장소 픽스처 셋에서 실행한 관측. 트리 6c2277295d9ddeaa92e83c0225910445f83cc1c2, `release.yml` 의 단계 본문(줄 59~123)을 `sed` 로 잘라 `TAG=<tag> bash inline.sh` 로 실행, 두 `git fetch` 줄은 bare 원격을 둔 스크래치 저장소에서 그대로 돌았다)
stdout (잘라 낸 첫 줄은 YAML 의 `run: |` 가 섞인 프로브 잡음이라 `command not found` 한 줄이 났고 검사에는 영향이 없다):
=== v9.9.9   (CHANGELOG 절 `## [9.9.9]`, system.yaml `version: v9.9.9`, 트레일러 올바름)
  RELEASE_PROVENANCE_GATE: all 7 checks passed for v9.9.9 (6df0e4f0ed751b7eb2c666712b10e8f784e73a44).
  exit=0
=== v9.9.8   (CHANGELOG 절 없음)
  ::error::RELEASE_PROVENANCE_GATE: check 5 (CHANGELOG): CHANGELOG.md at f4880fe70bc54e0dcbf24677a9625d94f920aeb3 has no '## [9.9.8]' or '## [v9.9.8]' section.
  Releases must be produced by the /harness:release maintainer harness.
  exit=1
=== v9.9.7   (system.yaml `version: v9.9.6`)
  ::error::RELEASE_PROVENANCE_GATE: check 6 (version SSOT): .moai/config/sections/system.yaml at d05f30336a33cedf56200301d961c87ca7eb944b has version='v9.9.6', expected 'v9.9.7'.
  Releases must be produced by the /harness:release maintainer harness.
  exit=1
note: 이것이 AC-GFD-008 의 (6)·(4)·(5) 가 변경 후 스크립트에서 같아야 하는 판정 줄과 종료 코드의 기준선이다(커밋 SHA 는 픽스처마다 다르다). 이 관측은 스크래치 저장소에서 한 번 한 것이고 소스 파일은 커밋하지 않았다
```

```
M-11  (AC-GFD-024 의 기준선 — 변경 전 이 트리에서 계약→미션 투영 시험이 초록이고 스냅숏 비교(limb 5)가 포함돼 있음을 한 번 관측. 트리 eda61419564296892b41d0e7608c8b136e52c04c)
command: go test -count=1 -v -run '^TestContractProjectsOntoMissionValidator$' ./internal/contract/
stdout (`--- PASS` 줄과 마지막 줄만 옮겼다, 전체 출력은 스크래치에 있다):
--- PASS: TestContractProjectsOntoMissionValidator (0.15s)
    --- PASS: TestContractProjectsOntoMissionValidator/limb_1_signed-valid_contract_is_not_refused_incomplete_contract (0.00s)
    --- PASS: TestContractProjectsOntoMissionValidator/limb_2_five_refusal_reasons_are_produced_by_mission.ValidateMissionDecision (0.00s)
    --- PASS: TestContractProjectsOntoMissionValidator/limb_3_slash-star-star_translates_to_a_prefix_that_neither_narrows_nor_widens (0.00s)
    --- PASS: TestContractProjectsOntoMissionValidator/limb_4_inner_wildcard_and_unmapped_fields_fail_closed_naming_the_field;_card_does_not_trigger (0.00s)
    --- PASS: TestContractProjectsOntoMissionValidator/limb_5_internal/mission_exported_surface_unchanged_from_the_M4-start_baseline (0.14s)
ok  	github.com/modu-ai/moai-adk/internal/contract	0.308s
exit: 0 (환경변수 정리 뒤 한 번의 복합 호출 `unset MOAI_KANBAN … && go test …`; limb 2 의 하위 시험 다섯은 위에서 줄였다)
tree: eda61419564296892b41d0e7608c8b136e52c04c
note: 이 시험은 `ProjectToMission` 을 한 인자로 부르는 현행 시그니처를 쓰므로 M2a 가 그 호출 지점 넷을 고친 뒤에도 같은 이름·같은 limb 다섯으로 초록이어야 한다 — 시험의 호출 지점 편집은 그 변경 하나로 한정한다. 이 관측은 도착 시점의 가드 행이다
```

## §C 인수 기준

### AC-GFD-001 — 기준 해석은 구성에서 나온다 (maps REQ-GFD-001)

- **분류**: release-blocking (M1 종료 게이트)
- **Given** 구성 둘 — `workflow: git-flow` + `develop_branch: develop`, `workflow: github-flow`. 그리고 해석표의 나머지 행(gitlab-flow·release-flow)을 담은 구성.
- **When** 카드 전달 경로의 **전달 기준** 해석(병합 준비 점검과 `factory complete` 가 대상 브랜치를 정하는 두 지점, 그리고 카드 diff 범위·push 사전 점검)을 각 구성에서 실행하면, 그리고 세션 종료 자동 병합을 github-flow 구성에서 실행하면
- **Then** git-flow 구성은 현행 출력(develop)과 같고, github-flow 구성은 `main` 을 해석하며, 표의 모든 행에서 해석 결과가 통합 목표 해석기의 결과와 같다. **세션 종료 자동 병합은 github-flow 구성에서 휴면이다** — `git merge` 호출 0, 통합 창 기록 쓰기 0.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-28 (종료 코드 0, 전달 기준을 읽는 두 지점 `factory_card.go:1064`·`factory_merge.go:93` 이 `DevelopBranch` 를 읽는다). 맥락: 원장 E-03 의 나머지 세 줄은 github-flow 에서 휴면이 맞는 git-flow 전용 경로라 이 기준의 RED 가 아니다 |
| RED 인 이유 | `DevelopBranch` 는 manual 프로필+git-flow 일 때만 채워지므로 github-flow 구성에서는 두 지점이 빈 기준을 본다(`factory_merge.go:93` 은 비면 "no integration branch configured … the merge target would be a guess" 로 거부한다 — 1회차 감사가 `factory_merge.go:95` 에서 확인). 기준이 이 두 지점을 통합 목표 해석기로 옮길 때까지 읽기가 남는다 |
| 돌연변이 프로브 | (1) 읽기를 리터럴 `"main"` 으로 바꾼 변이는 github-flow 행은 통과하고 git-flow 특성화 행에서 죽는다. (2) `DevelopBranch` 가 비면 `"main"` 으로 대체하는 변이는 두 행을 모두 통과하지만 gitlab-flow 행(해석 결과가 `main` 이 아니다)에서 죽는다. 그래서 시험은 해석표의 **모든 행**을 표 기반으로 돈다. (3) 같은 동작의 래퍼 접근자로 E-28 만 비우는 변이는 grep 을 통과하지만 표 기반 시험의 github-flow 행(`main`)에서 죽는다 — grep 은 형태를 보고 시험이 동작을 본다. (4) E-03 을 비우려고 세션 종료 자동 병합의 읽기도 통합 목표로 옮기는 변이는 E-28 을 건드리지 않으므로 grep 은 모르지만, github-flow 에서 main 으로의 로컬 `git merge` 가 켜지므로 휴면 양성 시험(`git merge` 호출 0)에서 죽는다. 휴면 시험은 도착 시점에 이미 초록인 **가드 행**이고 그 적색은 이 변이다 |
| 녹색 경로 | M1 이 두 읽기를 통합 목표 해석기로 옮기고 git-flow 특성화 시험을 먼저 고정한다. 세션 종료 자동 병합의 읽기는 옮기지 않는다 |
| 통과 출력 | `go test -v -run '^(TestDeliveryBaseResolvesFromIntegrationTarget|TestDeliveryBaseGitFlowUnchanged)$' ./internal/cli/ ./internal/config/` 가 두 시험 각각 4행 이상의 하위 시험 `--- PASS`, 그리고 `go test -v -run '^TestSessionExitAutoMergeInertUnderGitHubFlow$' ./internal/cli/` 가 `--- PASS`, 방문 시험 수 ≥ 3. E-28 의 명령은 종료 코드 1 과 빈 출력 |
| 트리 핀 | `855563dba79da74528e0f01560efe80f1016cc11` (E-28) |

### AC-GFD-002 — squash 병합을 알아보는 착지 판정 (maps REQ-GFD-002)

- **분류**: release-blocking (M2 종료 게이트)
- **Given** 격리 저장소 픽스처 아홉 — F1 커밋 1개 카드를 squash 병합, F2 커밋 2개 카드를 squash 병합, F3 병합 커밋으로 병합, F4 병합되지 않은 카드, F5 squash 병합 뒤 main 이 같은 파일을 다시 바꾼 경우(누적 patch-id 불일치)에 PR 상태 MERGED·`headRefOid` 가 로컬 팁과 같음·병합 커밋이 main 의 조상임을 알려 주는 `gh` 대역, F6 확정할 수 없는 경우(가져오기 실패), **F7 `gh` 대역은 PR 상태 MERGED 를 알려 주지만 로컬 카드 팁이 PR 의 `headRefOid` 와 다르다(병합 뒤 로컬 커밋이 더 있다)**, **F8 `gh` 대역이 제한 시간 10 초를 넘겨 응답하지 않는다**, **F9 `gh` 대역은 MERGED·`headRefOid` 가 로컬 팁과 같음을 알리지만 병합 커밋이 아직 로컬 통합 ref 에서 도달되지 않는다(가져오기 지연 — 세 번째 조건만 거짓)**
- **When** 착지 판정(`moai worktree done` 의 착지 점검, `moai worktree sweep`, 세션 종료 정리)을 각 픽스처에서 실행하면
- **Then** `done`·`sweep` 은 F1·F2·F3·F5 를 "착지", F4·F6·F7·F8·F9 를 "보존"으로 답한다(REQ-GFD-002 의 첫 문장 — 세 층). **세션 종료 정리 경로(REQ-GFD-002 의 둘째 문장)는 `gh` 를 한 번도 호출하지 않는다** — 1·2층만 쓰고, F1·F2·F3 은 "착지"로, 3층이 있어야 확정되는 F5 는 "보존"으로 답한다(그 트리는 다음 `sweep` 의 3층이 판정한다). 층 2 는 비교 커밋 수 상한 500 을 넘으면 "답하지 못함"이다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-15 (종료 코드 1, `done.go` 에 `cherry`·`patch-id`·`is-ancestor`·`gh pr` 가 없다). 보조 관측: `research.md` §5 — 커밋 2개 카드를 squash 병합하면 조상 판정은 종료 코드 1, `git cherry` 는 두 커밋 모두 `+` |
| RED 인 이유 | 현 `done.go` 는 `...` 범위의 우측 개수 0 만 보므로 squash 병합 카드는 영원히 미착지다 |
| 돌연변이 프로브 | `git cherry` 만 쓰는 변이는 F1 만 통과하고 F2 에서 죽는다. 조상 판정만 쓰는 변이는 F3 만 통과한다. "항상 착지" 변이는 F4 에서 죽는다. "항상 보존" 변이는 F1~F3·F5 에서 죽는다. **병합된 PR 이 있으면 무조건 착지로 읽는 변이는 F7 에서 죽는다**(병합 뒤 연장된 브랜치를 지우는 것이 그 변이의 결과다). 3층에서 `gh` 시간 초과를 착지로 읽는 변이는 F8 에서 죽는다. **3층이 병합 커밋의 통합 ref 조상 조건을 빼고 MERGED 와 `headRefOid` 일치만 보는 변이는 F9 에서 죽는다**(F5 의 대역은 조상 조건이 참이라 이 변이를 통과시킨다). 세션 종료 경로에 `gh` 호출을 넣는 변이는 호출 횟수 단언에서 죽고, 세션 종료 경로가 3층까지 도는 변이(REQ-GFD-002 를 한 문장으로 읽은 구현)도 F5 의 "보존" 단언과 호출 횟수 단언에서 죽는다 |
| 녹색 경로 | M2 가 세 층(조상, 누적 patch-id, PR 병합 상태)을 `worktree done`·`sweep` 이 공유하는 하나의 판정으로 모으고, 세션 정리는 1·2층만 쓴다 |
| 통과 출력 | `go test -v -run '^TestLandingPredicateSquashSafe$' ./internal/cli/worktree/` 에서 F1~F9 하위 시험 9개 `--- PASS`, 방문 수 9. `go test -v -run '^TestSessionExitLandingNoNetwork$' ./internal/cli/` 가 `--- PASS`(`gh` 대역 호출 0) |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-15) |

### AC-GFD-003 — 기본값에 리터럴 develop 이 없다 (maps REQ-GFD-003)

- **분류**: release-blocking (M1 종료 게이트)
- **Given** 구성 둘(git-flow, github-flow)
- **When** `moai worktree sweep --help` 의 `--base` 기본값, `moai worktree done` 의 착지 기준 브랜치, 카드 diff 범위 계산의 기준(`cardBaseBranch`), 목표 계약의 `MergeTarget` 을 각 구성에서 읽으면
- **Then** git-flow 는 `origin/develop`·`develop`, github-flow 는 `origin/main`·`main` 이다. (`moai worktree new` 는 이미 `config.LoadWorktreeBaseBranch` 를 읽어 `session_worktree.go:235` 에서 구성 값을 따르므로 도착 시점에 이미 초록이다 — 이 기준의 RED 가 아니라 **특성화 행**으로 두 구성에서 그대로 초록임을 확인한다.)

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-01 과 E-02 와 E-29 (모두 종료 코드 0, 리터럴 `develop` 상수·기본값 넷: `landingBaseBranch`, `--base` 기본값, `cardBaseBranch`, `MergeTarget`) |
| RED 인 이유 | 상수와 기본값이 구성과 무관하게 `develop` 이다 |
| 돌연변이 프로브 | 리터럴을 `main` 으로만 바꾼 변이는 github-flow 행을 통과하고 git-flow 행에서 죽는다. 구성 값을 읽되 빈 값이면 `develop` 으로 되돌리는 변이는 github-flow 행에서 죽는다. `goal.go` 의 `MergeTarget` 만 건드리지 않는 변이는 E-29 가 한 줄 남으므로 grep 에서 죽는다 |
| 녹색 경로 | M1 |
| 통과 출력 | E-01·E-02·E-29 명령이 종료 코드 1 과 빈 출력, `go test -v -run '^TestBaseDefaultsFollowIntegrationTarget$' ./internal/cli/ ./internal/cli/worktree/` 하위 시험 ≥ 2 `--- PASS` (선택한 시험 수가 0 이 아님을 `--- PASS` 줄로 확인) |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-01·E-02), `855563dba79da74528e0f01560efe80f1016cc11` (E-29) |

### AC-GFD-004 — 카드 PR 전달 간선 (maps REQ-GFD-004)

- **분류**: release-blocking (M2 종료 게이트)
- **Given** 카드 브랜치를 가진 격리 저장소와 bare 원격, 그리고 `gh pr create`·`gh pr view` 에 응답하는 `gh` 대역, `workflow: github-flow` 구성의 merge-ready 카드
- **When** `moai factory complete` 를 실행하면
- **Then** 카드 브랜치가 원격으로 push 되고, 대상이 `main` 인 PR 이 열리며, `gh pr view` 대역이 MERGED 를 돌려주기 전에는 카드가 병합 완료로 기록되지 않고, 돌려준 뒤에 기록된다. `git merge` 는 한 번도 호출되지 않고 통합 창 잠금 파일은 만들어지지 않는다. 새 병합 완료 상태는 선행 카드 통과 검사(`homestate/card_picked.go:195,205`)를 만족해 그 카드를 선행으로 둔 뒤 카드가 막히지 않고, 원격이 설정돼 있어도 `card_transition.go:532-537` 의 거부를 만나지 않고 `done` 에 닿는다(design D-4 의 소비자 세 지점).

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-16 (종료 코드 1, PR 간선 부재) 과 E-14 (종료 코드 0, `factory_card.go:1157` 의 로컬 `git merge --no-ff`) |
| RED 인 이유 | 현 경로는 로컬 병합만 있고 PR 간선이 없다 |
| 돌연변이 프로브 | PR 개설 직후 병합으로 기록하는 변이는 "MERGED 전에는 미기록" 단언에서 죽는다. 로컬 병합으로 처리하는 변이는 "`git merge` 미호출" 단언에서 죽는다. 창을 잡는 변이는 잠금 파일 단언에서 죽는다. PR 개설 없이 push 만 하는 변이는 `gh pr create` 호출 단언에서 죽는다. 새 상태를 선행 통과 목록(`card_picked.go:195`)에 넣지 않는 변이는 선행 통과 하위 시험에서, 새 상태에서 `done` 으로 가는 경로를 `card_transition.go:532` 거부에 걸리게 두는 변이는 `done` 도달 하위 시험에서 죽는다 |
| 녹색 경로 | M2 |
| 통과 출력 | `go test -v -run '^TestFactoryCompleteGitHubFlowPR$' ./internal/cli/ ./internal/homestate/` 하위 시험 ≥ 6 `--- PASS`(전달 간선 4 + 선행 통과 1 + `done` 도달 1). E-16 명령은 종료 코드 0 과 적중 줄 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-005 — PR 전 병합 준비 점검 (maps REQ-GFD-005)

- **분류**: release-blocking (M2 종료 게이트)
- **Given** 픽스처 넷 — 충돌이 있는 카드, 충돌 없는 카드, sync 감사 PASS 기록이 없는 카드, 트리 항등이 깨진 카드
- **When** PR 개설 직전의 병합 준비 점검을 실행하면
- **Then** 충돌·감사 누락·트리 항등 실패 카드는 PR 을 열지 않고 종료 코드가 0 이 아니며 사유를 출력한다. 충돌 없는 카드는 점검을 통과한다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-28 의 `factory_merge.go:93` 줄 (종료 코드 0, 준비 점검의 대상 브랜치가 `DevelopBranch` 에서 나온다) |
| RED 인 이유 | github-flow 구성에서 `DevelopBranch` 가 비어 현 점검은 대상 없이 거부하거나 건너뛴다(입력 보고서의 읽기 결과) |
| 돌연변이 프로브 | 충돌 검사를 건너뛰는 변이는 충돌 픽스처에서 죽는다. 점검을 현재 브랜치 대비로 하는 변이는 충돌 픽스처에서 죽는다. 항상 실패시키는 변이는 충돌 없는 픽스처에서 죽는다 |
| 녹색 경로 | M2 |
| 통과 출력 | `go test -v -run '^TestMergeReadinessBeforePR$' ./internal/cli/ ./internal/factorylane/` 하위 시험 4개 `--- PASS`, 방문 수 4 |
| 트리 핀 | `855563dba79da74528e0f01560efe80f1016cc11` (E-28) |

### AC-GFD-006 — 창은 선행 조건이 아니고 슬롯은 남는다 (maps REQ-GFD-006)

- **분류**: release-blocking (M2 종료 게이트)
- **Given** github-flow 구성의 merge-ready 카드와 git-flow 구성의 merge-ready 카드
- **When** 각 구성에서 `moai factory complete` 를 실행하고, 별도로 `moai slot acquire`·`release` 를 실행하면
- **Then** github-flow 에서는 `moai integration status` 가 어느 시점에도 보유자를 보이지 않고, git-flow 는 현행 창 동작이 보존되며, 슬롯 획득과 반납은 두 구성 모두에서 종료 코드 0 이다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-14 (종료 코드 0, 로컬 `git merge --no-ff` 호출의 존재) |
| RED 인 이유 | 현 `complete` 는 github-flow 에서도 로컬 병합을 하며 그 호출이 통합 창을 잡은 뒤에만 도는 것은 입력 보고서가 `factory_card.go` 를 읽은 결과다(E-14 는 호출의 존재만 보인다) |
| 돌연변이 프로브 | github-flow 에서 창을 잡고 곧 놓는 변이는 "어느 시점에도 보유자 없음" 단언에서 죽는다. 두 구성 모두에서 창을 없애는 변이는 git-flow 특성화에서 죽는다 |
| 녹색 경로 | M2 |
| 통과 출력 | `go test -v -run '^(TestFactoryCompleteNoWindowGitHubFlow|TestFactoryCompleteWindowGitFlowUnchanged)$' ./internal/cli/` 각 `--- PASS`, 기존 슬롯 시험 `go test -v -run '^(TestSlotCLI_AcquireJSONFields|TestSlotCLI_ForceReportsDisplaced|TestSlotCLI_HeldAndBusyCarryDistinctExitCodes|TestSlotCLI_ReleaseRoundTripAndRefusals)$' ./internal/cli/` 네 개가 변경 없이 `--- PASS`(실행된 시험 수 4) |
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
| 통과 출력 | `go test -v -run '^TestReleaseScriptHeadAdmission$' ./internal/template/` 에서 하위 시험 5개 `--- PASS`(`scripts/` 최상위는 Go 패키지가 아니므로 스크립트 시험은 이미 `release.yml` 을 읽고 `bash` 를 실행하는 시험이 있는 `internal/template` 에 둔다 — design D-6) |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-008 — rc 태그의 출처 검증 (maps REQ-GFD-008)

- **분류**: release-blocking (M3 종료 게이트)
- **Given** 출처 검증을 시험 가능한 위치(`scripts/verify-release-provenance.sh <tag>`)로 옮긴 뒤 격리 저장소에 만든 태그 픽스처 여섯 — (1) 올바른 트레일러와 main 조상 커밋을 가진 rc 태그(CHANGELOG 절·`system.yaml` 버전 없음), (2) 트레일러 커밋이 틀린 rc 태그, (3) main 조상이 아닌 커밋의 rc 태그, (4) CHANGELOG 절이 없는 정식(접미사 없는) 태그, (5) `system.yaml` 버전이 태그와 다른 정식 태그, **(6) 일곱 검사를 모두 만족하는 정식 태그**. 그리고 출하되는 관문: `.github/workflows/release.yml` 의 `verify-provenance` 잡
- **When** 출처 검증을 각 태그에 실행하고, 워크플로 파일을 읽으면
- **Then** (1)(6) 만 통과하고 (2)~(5) 는 각자의 검사 이름을 대며 실패한다. **(6) 의 통과 줄은 `RELEASE_PROVENANCE_GATE: all 7 checks passed for <tag> (<commit>)`(변경 전 인라인 단계의 문구, 원장 E-44)이고, (4)(5) 의 실패 줄은 변경 전 인라인 단계가 찍던 `check 5 (CHANGELOG)`·`check 6 (version SSOT)` 이름을 그대로 쓴다 — 접미사 없는 태그의 일곱 검사는 바뀌지 않는다.** 그리고 **워크플로 배선이 선다**: (W1) `git grep -n -E "check (5|6) \(" -- .github/workflows/release.yml` 가 종료 코드 1·빈 출력(인라인 검사 5·6 이 워크플로에 없다), (W2) `git grep -n -F "verify-release-provenance" -- .github/workflows/release.yml` 가 종료 코드 0 과 호출 줄 하나 이상(워크플로가 스크립트를 부른다), (W3) 시험이 `verify-provenance` 잡의 `run` 본문에서 호출 줄 `bash scripts/verify-release-provenance.sh "${TAG}"` 을 뽑아 **시험이 스크립트를 직접 부르는 인자(`<tag>` 하나)와 같은 인자**임을 단언하고, 같은 호출 줄을 `TAG` 만 치환해 (1)~(6) 의 격리 저장소에서 재생해 직접 호출과 같은 종료 코드·같은 판정 줄을 얻는다. 네트워크가 필요한 두 `git fetch` 줄과 `env: TAG: ${{ github.ref_name }}` 는 워크플로에 남아 호출 줄 앞에 있다(순서 단언).

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-06 (종료 코드 0, 인라인 검사 5·6 이 워크플로에 있고 모든 태그에 적용된다 — W1 의 적색)과 E-42 (종료 코드 1, 워크플로가 스크립트를 부르지 않는다 — W2 의 적색), 그리고 E-44 (종료 코드 0, 변경 전 성공 문구 — (6) 과 W3 의 기준). 맥락 행 M-10 이 접미사 없는 세 픽스처((6)·(4)·(5) 모양)에서 변경 전 인라인 단계가 낸 판정 줄과 종료 코드의 기준선이다 |
| RED 인 이유 | 접미사 구분이 없어 rc 태그가 CHANGELOG 절 없이는 거부되고, 검증 로직이 시험할 수 없는 인라인 셸에만 있어 스크립트도 호출도 없다. 스크립트 시험만 있으면 이 RED 의 둘째 이유(워크플로 배선 부재)가 초록으로 가려진다 |
| 돌연변이 프로브 | rc 에서 모든 검사를 건너뛰는 변이는 (2)(3) 에서 죽는다. 정식 태그 검사를 약화하는 변이는 (4)(5) 에서 죽는다. **스크립트를 추가하고 그 시험만 초록으로 만든 채 `release.yml` 을 그대로 두는 변이는 W2 에서 죽는다**(E-06 이 종료 코드 0 으로 남아 W1 도 죽는다). **스크립트를 부르면서 인라인 검사 5·6 을 남기는 변이는 W1 에서 죽는다**(rc 태그가 여전히 5·6 에서 거부된다). **워크플로가 스크립트에 다른 인자(예: 커밋 SHA 나 빈 인자)를 넘기는 변이는 W3 에서 죽는다.** **접미사 없는 태그의 일곱 검사를 rc 와 같이 약화하는 변이는 (4)(5) 에서, 통과 문구를 바꾸는 변이는 (6) 의 문구 단언에서 죽는다.** 호출 줄이 `git fetch` 앞에 오는 변이는 순서 단언에서 죽는다 |
| 녹색 경로 | M3 (스크립트, 워크플로 배선, 인라인 검사 본문 삭제를 한 마일스톤에서). 접미사 없는 태그의 동작이 같으므로 PRE-CUTOVER-SAFE 다(AC-GFD-016 의 허용 목록에 `release.yml` 이 있다) |
| 통과 출력 | `go test -v -run '^(TestReleaseProvenanceRcRule|TestReleaseWorkflowInvokesProvenanceScript)$' ./internal/template/` 에서 `TestReleaseProvenanceRcRule` 하위 시험 6개와 `TestReleaseWorkflowInvokesProvenanceScript`(W3 와 순서 단언) 각 `--- PASS`(대상은 `scripts/verify-release-provenance.sh`, `--- SKIP` 없음, 실행된 시험 수 ≥ 7). 그리고 W1 명령이 종료 코드 1 과 빈 출력, W2 명령이 종료 코드 0 과 호출 줄 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-06), `6c2277295d9ddeaa92e83c0225910445f83cc1c2` (E-42·E-44) |

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
| 통과 출력 | `go test -v -run '^TestGoreleaserPrereleaseAuto$' ./internal/template/` 하위 시험 `--- PASS`. `goreleaser check` 는 도구가 설치된 환경에서만 종료 코드 0 을 기록하고, 설치되지 않았으면 Gap 으로 적는다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` |

### AC-GFD-010 — 태그 전 3-OS 매트릭스 (maps REQ-GFD-010)

- **분류**: release-blocking (M3 종료 게이트)
- **Given** `gh` 대역 일곱 — 워크플로 `Release PR Multi-OS Verification` 의 `workflow_dispatch` 실행에 대해 (a) 대상 SHA 에 세 OS 레그(`Release Verify (ubuntu-latest)`·`(macos-latest)`·`(windows-latest)`)가 모두 성공한 런, (b) 런 없음, (c) 다른 SHA 에 성공한 런, (d) 대상 SHA 에 실패한 런, (e) 대상 SHA 에 진행 중인 런, (f) 런 전체는 success 인데 레그 하나가 skipped 인 런(`release-pr-multi-os.yml` 의 `go_code` 게이트가 PR 이벤트 런에서 만들 수 있는 모양 — `workflow_dispatch` 에서는 레그가 항상 돌지만 스크립트는 런 수준 결론이 아니라 레그를 읽어야 한다), (g) 대상 SHA 에서 `workflow_dispatch` 가 아니라 `pull_request` 이벤트로 돈 성공 런
- **When** `scripts/release.sh v9.9.9-rc.1 --dry-run` 을 **두 스위치 상태에서** 각각 실행하면 — 옵션 없음, `--require-matrix-run`
- **Then** 옵션 없음에서는 (a)~(g) 모두 매트릭스 게이트를 보지 않고 현행 동작과 같다(`gh run list` 호출 0). `--require-matrix-run` 에서는 (a) 만 통과하고 (b)~(g) 는 종료 코드 1 로 태그 전에 거부하며 거부 사유가 어느 조건인지 이름으로 적힌다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-09 (종료 코드 1, 스크립트에 매트릭스 게이트 참조가 없다)와 E-31 (종료 코드 1, `--require-matrix-run` 옵션이 스크립트·하네스에 없다) |
| RED 인 이유 | 현 스크립트는 결합 상태 API 한 번만 보고 새 옵션이 없다 |
| 돌연변이 프로브 | 아무 SHA 의 성공 런이나 받는 변이는 (c) 에서 죽는다. 결론을 보지 않는 변이는 (d)(e) 에서 죽는다. 게이트를 건너뛰는 변이는 (b) 에서 죽는다. 런 수준 결론만 보는 변이는 (f) 에서 죽고, 이벤트를 보지 않는 변이는 (g) 에서 죽는다. **옵션 없이도 게이트를 켜는 변이는 옵션 없음 상태의 `gh run list` 호출 0 단언에서, 옵션이 있어도 게이트를 못 켜는 변이는 `--require-matrix-run` 상태의 (b) 에서 죽는다** |
| 녹색 경로 | M3 (옵션 B, 스위치 기본 꺼짐). 하네스가 옵션을 넘기도록 바꾸는 켜기는 M5 이고 AC-GFD-016 의 사후 방향이 그 존재를 단언한다 |
| 통과 출력 | `go test -v -run '^TestReleaseScriptMatrixGate$' ./internal/template/` 에서 두 상태 × 일곱 대역 하위 시험 14개 `--- PASS` |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-09), `855563dba79da74528e0f01560efe80f1016cc11` (E-31) |

### AC-GFD-011 — 낡은 develop 서술 잔존 0 (maps REQ-GFD-011)

- **분류**: release-blocking (**M5 종료 게이트** — 가드는 M5 가 바꾸는 `git-strategy.yaml` 값으로 무장하고 M5 범위의 `AGENTS.local.md`·`.claude/rules/local/**`·`.moai/docs/*doctrine*` 까지 훑으므로 M4 와 M5 가 한 묶음으로 적용된 트리에서 관측한다. 가드의 픽스처 자가 시험은 AC-GFD-021 이 M4 종료 게이트로 따로 지킨다. 절체 변경 묶음의 병합 조건)
- **Given** 절체 변경 묶음이 적용된 트리와 `internal/template` 의 스윕 가드 시험(무장 상태)
- **When** 가드를 실행하면
- **Then** 범위 표면(design D-9 의 하위 트리 열 곳 — `AGENTS.md`·`AGENTS.local.md`·`CLAUDE.md`·README 4종, `.claude/rules/**`·`.claude/agents/**`·`.claude/skills/**`·`.claude/output-styles/**`, docs-site 4로케일, `.moai/docs/**`, 템플릿 사본, `.moai/project/codemaps/*.md`, `scripts/` 의 `*.sh`·`*.py`)에 살아 있는 develop 기준 서술이 0 이고 — **강한 적중과 약한 적중(`Work starts from develop.` 처럼 분기 말 토큰이 없는 줄) 모두 줄 전체 허용 항목에 오르지 않은 것이 0 이다(D-27)** —, 전체 방문 수가 **1372 이상**이며 하위 트리마다 design D-9 표의 바닥값(6·101·19·229·2·558·33·367·5·48) 이상이다. 허용 목록 항목은 **약한 적중 허용을 포함해 전체 40개 이하**이고 각 항목의 `Literal` 은 그 파일의 한 줄 전체와 일치하며 항목마다 `Why` 가 있다. (바닥값 리터럴은 M4 착수 시 같은 `git ls-files` 계수로 다시 잰 값 ⌊0.9×수⌋ 로 갱신하고 판정서에 차이를 적는다.) **한계(Limits)**: H1 제목의 표지는 문서 전체를 침묵시키고, 날짜를 곁들인 표지를 살아 있는 줄에 붙이는 변이와 대문자 `Develop`·`DEVELOP` 서술은 가드가 보지 못한다 — 리뷰 소관이며 가드는 표지 면제 줄 수와 허용 목록 사용 수를 출력해 증가를 눈에 띄게 한다(design D-9 의 남는 한계, 원장 E-45).

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-07 (종료 코드 0, `AGENTS.md` 에 5줄)과 E-18 (종료 코드 0, README 4종 모두 적중) |
| RED 인 이유 | 서술이 살아 있다 |
| 돌연변이 프로브 | 가드의 자체 시험(AC-GFD-021 의 픽스처 양방향)이 죽인다. 추가로 **서술을 지우지 않고 허용 목록에만 올리는 변이**: `{File: "AGENTS.md", Literal: "develop"}` 같은 맨 토큰 항목은 `Literal` 이 한 줄 전체와 일치하지 않아 거부되고, 줄 전체 항목은 항목마다 `Why` 가 필요하며 전체 40개 상한과 낡은 항목 거부가 항목 수를 묶는다. **한 파일만 걷는 가드**(방문 수 바닥값 1 짜리)는 전체 1372·하위 트리별 바닥값에서 죽는다. **하위 트리 하나를 통째로 빼는 변이**(예: codemaps 나 `scripts/`)는 그 하위 트리의 바닥값에서 죽는다. **줄에 `(legacy)` 를 덧붙여 살아 있는 문장을 가리는 변이**는 `legacy` 가 표지가 아니므로 위반으로 남고(F-red-9), **줄 표지 `RETIRED` 를 날짜 없이 붙이는 변이**는 날짜·식별자 동반 요건에서 죽는다. 날짜를 곁들인 표지를 살아 있는 문장에 붙이는 변이는 막지 못한다 — Residual-risk, 리뷰 소관이며 가드는 표지 면제 줄 수를 출력해 증가를 눈에 띄게 한다. **약한 적중을 보고만 하고 위반으로 세지 않는 변이(0.1.1 설계)는 토큰 없는 살아 있는 문장(`Work starts from develop.`·`Lane PRs go to develop.`)에서 죽는다 — AC-GFD-021 의 F-red-12·13 이 그 입력이고 원장 M-9 가 이 입력을 가드가 적색으로 읽음을 관측했다. 허용 항목의 `Literal` 을 줄 일부와 대조하는 변이는 F-red-14 에서 죽는다** |
| 녹색 경로 | M5 (M4 의 t1448 착지 뒤 규칙 편집과 가드 본체, M5 의 설정 값 변경이 한 묶음으로 적용된 트리에서 무장 실행) |
| 통과 출력 | `go test -v -run '^TestGitHubFlowSweepGuard$' ./internal/template/` 가 `--- PASS` 와 `visited=<N>` 줄(N ≥ 1372)·하위 트리별 방문 수 줄·**허용 목록으로 통과한 약한 적중 수와 위반으로 남은 약한 적중 수(0) 줄**을 내고 `--- SKIP` 이 없다. 판정서에 무장 실행 시점의 약한 적중 줄 목록(파일:줄)을 적는다 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-07·E-18), `855563dba79da74528e0f01560efe80f1016cc11` (M-4 의 바닥값 근거), `6c2277295d9ddeaa92e83c0225910445f83cc1c2` (M-9, E-45) |

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
- **Given** `zone-registry.md` 에 등재된 `CONST-V3R5-027`·`-028`(원장 E-34). 그리고 **이 SPEC 이 건드리는 `[ZONE:Frozen]` 줄의 전수**(design D-26 보충, 원장 E-38~E-41): `spec-workflow.md`(로컬·템플릿 사본 같은 세 줄) — Route A/B 문단 표지 줄(`Every MoAI SPEC follows the three-phase lifecycle …`)은 **미등재**, 단계 순서 머리줄(`Step ordering rules:`)은 머리줄 자체가 **미등재**이고 그 아래 `Step 1 (plan) MUST execute in main checkout on BOTH routes. NO L2/L3 worktree at this step` 가 **등재 `CONST-V3R5-027`**, `Step 4 (cleanup) applies to **Route B only**…` 가 **등재 `CONST-V3R5-028`**, plan 단계 표지 줄(`Execute in main checkout. NO worktree at this step.`)은 **미등재**이며 그 아래 `Create comprehensive specification using EARS format.`(`CONST-V3R2-001` 의 clause)는 **편집하지 않는다**. `moai constitution list --file …spec-workflow.md` 는 등재 셋(`CONST-V3R2-001`·`-027`·`-028`)을 내고 템플릿 사본 경로는 `No entries.`(원장 E-40·E-41). `worktree-integration.md` 의 `[ZONE:Frozen]` 표지 두 줄(`:390` per-step applicability, `:401` disposal contract)도 **등재되어 있지 않다**(원장 E-33: `moai constitution list --file …worktree-integration.md` 가 `No entries.`)
- **When** 레인이 개정 입력(조항마다 `--rule`·`--before`·`--after`·`--evidence`)과 `--dry-run` 제안을 `.moai/reports/t1453/constitution-amend-inputs.md` 에 준비하고, **운영자가** `moai constitution amend` 를 **정확히 두 번**(`CONST-V3R5-027` 한 번, `CONST-V3R5-028` 한 번) 실행하면 — 5단째 인간 승인은 대화형 Y/N 이라 에이전트가 만들어 낼 수 없다(`internal/constitution/human_oversight.go:42-55`, 1회차 감사 관측)
- **Then** (1) 두 실행이 각각 종료 코드 0 이고 그 출력이 판정서에 남는다. (2) 개정 뒤 `moai constitution validate` 가 종료 코드 0 이다(원장 E-35 가 개정 전 기준선). (3) 등재 문언과 `spec-workflow.md` 본문의 새 문언이 같다. (4) `worktree-integration.md:390,401` 과 `spec-workflow.md` 의 **미등재 표지 줄 둘**(Route A/B 문단, plan 단계 `NO worktree at this step`)은 `amend` 가 아니라 **일반 편집**으로 바뀌고 개정된 `spec-workflow.md` 등재 문언과 같은 방향이며 로컬·템플릿 사본이 같다 — 미등재 줄을 지키는 장치는 `amend` 가 아니라 AC-GFD-017 의 0-grep 과 AC-GFD-015 의 사본 정합 시험이다. (5) M4 가 편집하는 파일마다 `moai constitution list --file <경로>` 의 출력이 판정서에 있고 등재 줄은 `amend` 로만, 미등재 줄은 일반 편집으로 처리됐다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-24 (종료 코드 0, `CONST-V3R5-028` 문언)과 E-10 (종료 코드 0, `spec-workflow.md` 의 plan 단계 금지 문장 1회) |
| RED 인 이유 | 두 clause 가 개정 전 문언을 담고 있다 |
| 돌연변이 프로브 | 레지스트리를 손으로 고친 변이는 문장 일치 grep 을 통과하지만 `amend` 실행 출력(두 건)이 없어서 이 기준에서 죽는다. 본문만 고치고 레지스트리를 두는 변이는 `validate` 와 일치 단언에서 죽는다. 등재되지 않은 줄을 `amend` 로 고치려는 변이는 `--rule` 이 등재 ID 만 받으므로(`internal/cli/constitution.go:513` 의 `registry.Get`, "rule … not found") 실행이 거부된다. 개정 후 한 clause 만 새 문언인 변이는 일치 단언에서 죽는다. **`CONST-V3R5-027`·`-028` 은 개정하고 미등재 plan 단계 표지 줄(`NO worktree at this step`)을 그대로 두는 변이는 `amend`·`validate`·일치 단언을 모두 통과하지만 AC-GFD-017 의 0-grep 에서 죽는다**(원장 E-36·E-37 이 그 줄을 종료 코드 0 으로 보인다). 등재 clause 를 건드리는 `Create comprehensive specification using EARS format.` 편집은 `validate` 에서 죽는다 |
| 녹색 경로 | M4 (t1448 착지 뒤, 이 개정은 SPEC-LATE-BRANCH-REDESIGN-001 의 B3 를 흡수) |
| 통과 출력 | 운영자 실행 `amend` 출력 **2건**(각 종료 코드 0, 5단 통과), `moai constitution validate` 종료 코드 0, 등재 clause 와 `spec-workflow.md` 에서 `NO L2/L3 worktree at this step`·`BOTH run AND sync PRs` 가 0(`git grep -c -F …` 종료 코드 1), 본문과 등재의 새 문언이 같다. 레인이 `--dry-run` 을 돌릴 수 있었으면 그 출력, 대화형 층 때문에 돌릴 수 없었으면 그 사실을 판정서의 Gap 으로 적는다 |
| 트리 핀 | `855563dba79da74528e0f01560efe80f1016cc11` (E-33·E-34·E-35), `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-24·E-10) |

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
- **When** 실제 사본 정합 시험 네 개를 앵커 고정 `-run` 으로 실행하면 — `go test -v -run '^(TestRuleTemplateMirrorDrift|TestDeclaredRuleMirrorForks|TestLateBranchTemplateMirror|TestSanitizedPairParity)$' ./internal/template/`
- **Then** 종료 코드 0 이고 네 시험이 각각 `--- PASS` 줄을 낸다(실행된 시험 수 4, 0 이 아니다). 한 줄이라도 `--- SKIP` 이나 `[no tests to run]` 이면 통과가 아니다. (1회차 계획이 골랐던 `TestCatalogHashParity`·`TestManifestHashFormat`·`make agents-emit-check` 는 카탈로그 해시 형식과 `.codex` TOML 의 정합을 보며 로컬 규칙·스킬과 템플릿 사본의 정합을 비교하지 않으므로 이 요구(REQ-GFD-015)의 증거가 아니다. 템플릿 본문을 고치면 카탈로그 해시가 따라 바뀌어야 하는 별도 저장소 절차는 이 기준의 밖이다.)

| 셀 | 내용 |
|---|---|
| 회귀 방지 신호 | 네 시험이 이미 존재하고 이 트리에서 초록이다(원장 M-8: 네 시험 `--- PASS`, 종료 코드 0). 변경이 로컬 사본과 템플릿 사본의 방향을 어긋나게 하면 적색이 된다 |
| 돌연변이 프로브 | 로컬 `.claude/rules/**` 사본만 고치는 변이는 `TestRuleTemplateMirrorDrift` 가 잡도록 설계된 입력이다. **이 입력의 적색은 이번 계획 회차에 관측하지 못했다**(저장소 파일을 바꿀 수 없는 계획 단계라 Gap) — 런 단계가 M1 병합 전에 한 사본만 바꿔 이 시험이 적색이 됨을 관측해 판정서에 남긴다 |
| 한계 | 임베드 축(`make build` 가 싣는 내용)은 리더가 배치 종료 때 빌드하므로 이 SPEC 에서 Gap 이며 통과로 기록하지 않는다 |
| 트리 핀 | `855563dba79da74528e0f01560efe80f1016cc11` (M-8) |

### AC-GFD-016 — 저장소 설정과 CI 는 절체 시점에 함께 (maps REQ-GFD-016)

- **분류**: release-blocking (M5 종료 게이트)
- **Given** 절체 변경 묶음
- **When** 두 방향을 모두 단언하면 — (사전) 절체 경계 **이전에 develop 에 병합되는 모든 커밋**(M1~M3 와 M4 의 가드 픽스처 커밋과 M6 의 추가 산출물 포함, 런북 2단계의 묶음 병합 직전까지)의 변경 파일 목록이 아래 열거한 파일을 건드리지 않는다, (사후) 묶음이 적용된 트리에서 아래 값이 github-flow 기준이다
- **Then** 세 갈래를 모두 만족한다.
  - **허용 목록**(M3 가 바꿀 수 있다): `.github/workflows/release.yml`(태그 push 에서만 도는 파일이고 접미사 없는 태그의 동작이 같다), `.goreleaser.yml`, `scripts/` 아래 파일(`release.sh`·`verify-release-provenance.sh`·`cutover-precheck.sh`·`cutover-rehearsal.sh`), `internal/**` 의 코드·시험.
  - **사전 방향의 금지 목록**(열거, 절체 경계 이전의 어떤 병합도 건드리지 않는다): `.moai/config/sections/git-strategy.yaml`·`.moai/config/sections/workflow.yaml`·`.coderabbit.yaml`, 워크플로 `ci.yml`·`codeql.yml`·`graph-freshness.yml`·`judgment-first-consistency.yaml`·`lsel-leak-guard.yaml`·`template-neutrality-check.yaml`·`test-install.yml`·`workflow-parse-guard.yaml`·`docs-i18n-check.yml`·`spec-lint.yml`, `.claude/agents/harness/hns-release-specialist.md`, `AGENTS.local.md`, **그리고 CUTOVER-TIME 산문 표면 — design D-9 하위 트리 1~9 의 추적 `*.md`·`*.md.tmpl` 파일(`AGENTS.md`·`CLAUDE.md`·README 4종, `.claude/rules/**`·`agents/**`·`skills/**`·`output-styles/**`, `docs-site/content/**`, `.moai/docs/**`, `internal/template/templates/**`, `.moai/project/codemaps/*.md`)**. 스윕 가드의 시험 파일(`internal/template/*_test.go`)은 `.md` 가 아니라 이 금지에 걸리지 않으므로 문서 묶음 중 경계 이전에 병합할 수 있는 유일한 산출물이다(가드 픽스처 시험, PRE-CUTOVER-SAFE).
  - **사후 방향**(묶음이 적용된 트리): `worktree_base_branch: main`, `workflow: github-flow`, `.coderabbit.yaml` 에 `develop` 없음, `spec-lint.yml` 이 `develop` 을 가져오지 않음, 릴리스 하네스 본문이 `--require-matrix-run` 을 넘김(`git grep -n -F -e "--require-matrix-run" -- .claude/agents/harness/hns-release-specialist.md` 종료 코드 0). **그리고 워크플로 push 트리거에는 `develop` 이 그대로 남는다**(`ci.yml:18 branches: [main, develop]` 불변 — develop 팁 CI 가 절체 내내 관측되도록, design D-17). 이 마지막 단언은 묶음이 push 트리거를 일찍 건드리는 변이를 죽인다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-08 (종료 코드 0, 저장소 값이 git-flow), E-19 (종료 코드 0), E-20 (종료 코드 0), E-31 (종료 코드 1, 하네스에 옵션 호출이 없다). 사후 단언의 push 트리거 불변 행은 이미 참이다(E-21, 종료 코드 0, `branches: [main, develop]`) — **가드 행**이며 그 적색은 "묶음이 push 트리거를 바꾼" 변이다 |
| RED 인 이유 | 세 파일이 develop 기준 값을 담고 있고 하네스에 옵션 호출이 없다 |
| 돌연변이 프로브 | 구성을 일찍 바꾼 변이는 사후 단언은 통과하지만 사전 단언에서 죽는다(변경 파일 목록에 `git-strategy.yaml` 이 든다). 구성만 바꾸고 CodeRabbit 은 두는 변이는 사후 단언에서 죽는다. `spec-lint.yml` 의 가져오기를 두는 변이는 사후 단언(E-20 명령 종료 코드 1)에서 죽는다. 가드 픽스처 커밋에 `.coderabbit.yaml` 을 끼워 넣는 변이는 사전 방향이 **모든** 사전 병합 커밋을 보므로 죽는다. **M4 의 산문 묶음을 절체 묶음에서 떼어 일찍 병합하는 변이(산문은 github-flow 를 말하고 `git-strategy.yaml` 은 git-flow 를 말하는 어긋남)는 사전 방향의 산문 표면 금지에서 죽는다**(그 시점 가드는 무장 전이라 이 단언이 유일한 방어다 — design D-9). 도착 시점의 사전 단언은 가드 행으로 이미 참이다(원장 E-43: 카드 트리 `4bf547bca…` 에서 이 HEAD 까지 금지 경로가 바뀌지 않았다, 종료 코드 0·빈 출력). 하네스가 옵션을 넘기지 않는 변이는 `--require-matrix-run` 단언에서 죽는다. 묶음이 push 트리거에서 `develop` 을 빼는 변이는 E-21 불변 단언에서 죽는다 |
| 녹색 경로 | M5, 적용은 런북 2단계 |
| 통과 출력 | 사후: E-19·E-20 명령이 종료 코드 1 과 빈 출력, E-21 명령이 `branches: [main, develop]` 과 `branches: [main]` 두 줄을 그대로 출력, E-08 명령이 `worktree_base_branch: main`·`workflow: github-flow` 값을 출력, E-31 의 하네스 한정 형태가 종료 코드 0. 사전: 사전 병합 커밋 범위의 `git diff --name-only <절체 전 기준>..<2단계 직전 develop>` 가 금지 목록과 교집합 0 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-08·E-19·E-20·E-21), `855563dba79da74528e0f01560efe80f1016cc11` (E-31), `6c2277295d9ddeaa92e83c0225910445f83cc1c2` (E-43) |

### AC-GFD-017 — t810 범위의 흡수 (maps REQ-GFD-017)

- **분류**: release-blocking (M4 종료 게이트)
- **Given** t810 의 REQ-LBR-001..006 과 이 SPEC 의 문서 마일스톤
- **When** 여섯 의도를 하나씩 이 SPEC 의 인수 기준·측정에 대응시키면
- **Then** 대응표가 판정서에 있고, 다음이 모두 참이다 — `spec-workflow.md` 에서 "opens a PR per phase"·"NO L2/L3 worktree at this step"·**"NO worktree at this step"**(plan 단계 미등재 표지 줄, design D-26 보충)·"BOTH run AND sync PRs" 가 0, `worktree-integration.md` 에서 "BOTH run PR AND sync PR" 가 0, `delivery.md`·`quality-gates-context.md` 에서 `Step 3.3.5` 가 0, 새 서술(SPEC 당 PR 1개·plan 단계 워크트리 진입·단일 PR 폐기 조건)이 각 파일에 있고, 두 사본이 같은 방향이다. **`spec-workflow.md` 의 로컬 사본과 템플릿 사본 양쪽에서 위 0-grep 이 모두 성립한다**(`git grep -c -F "NO worktree at this step" -- <두 경로>` 종료 코드 1). 등재되지 않은 Frozen 줄은 일반 편집이며 개정된 등재 문언(AC-GFD-013)과 같은 방향이어야 한다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-10 (종료 코드 0), E-24 (종료 코드 0), E-25 (종료 코드 0, `Step 3.3.5` 정의와 참조 각 1), **E-36 (종료 코드 0, 로컬 사본 164 줄 `NO worktree at this step`)과 E-37 (종료 코드 0, 템플릿 사본 같은 줄)** |
| RED 인 이유 | 세 문서가 두-PR 전제·plan 단계 금지·Step 3.3.5 를 담고 있고, `spec-workflow.md` 는 등재되지 않은 plan 단계 금지 문장을 두 사본에 각각 담고 있다 |
| 돌연변이 프로브 | 옛 문장만 지우고 새 규칙을 쓰지 않는 변이는 긍정 앵커에서 죽는다. 한 사본만 고친 변이는 사본 일치 단언에서 죽는다. **`CONST-V3R5-027`·`-028` 을 `amend` 하고 Route A/B 문단(`opens a PR per phase`)이나 plan 단계 표지 줄 `NO worktree at this step` 를 남기는 변이는 `NO worktree at this step` 0-grep(로컬·템플릿 각각)과 `opens a PR per phase` 0-grep 에서 죽는다 — AC-GFD-013 의 `amend`·`validate` 는 미등재 줄을 보지 못하므로 이 0-grep 이 그 줄의 유일한 지킴이다** |
| 녹색 경로 | M4 (AC-GFD-013 의 amend 와 함께) |
| 통과 출력 | 위 grep 이 해당 경로에서 종료 코드 1(`NO worktree at this step` 은 로컬·템플릿 두 경로 각각), 긍정 앵커 grep 이 종료 코드 0 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-10·E-24·E-25), `6c2277295d9ddeaa92e83c0225910445f83cc1c2` (E-36·E-37) |

### AC-GFD-018 — main 과 develop 의 수렴 리허설 (maps REQ-GFD-018)

- **분류**: release-blocking (M6 종료 게이트)
- **Given** 스크래치 클론에서 develop 팁 `284e09c44023598affe486f17701717ca173e6ca` 와 main 팁 `4755c5e506225ba90b7a303c5763fa303c699492` 로 만든 두 로컬 브랜치
- **When** develop 이 main 을 흡수하고 그 결과를 main 에 병합 커밋으로 병합하면
- **Then** 병합 후 main 의 트리가 병합 직전 develop 팁의 트리와 같고, develop 팁이 main 의 조상이며, 병합 커밋을 되돌린 main 의 트리는 되돌리기 전 main 팁의 트리 `d57fdeecf4b8849da70a6693bbff19ad77ad5932` 와 같다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-27 (종료 코드 1, 수렴 리허설 스크립트·시험이 없다). 맥락 행(RED 가 아니다): E-11 (종료 코드 0, 두 팁의 트리가 다르다), E-22 (종료 코드 0, 발산 `7614	1`), E-23 (종료 코드 0, 흡수 병합이 충돌 없는 트리 `992f57f42aee6c3f9a39754c3e0095eaadf9d44b` 를 낸다) |
| RED 인 이유 | 이 기준이 요구하는 **산출물**(`scripts/cutover-rehearsal.sh` 와 그 실행 출력)이 없다. E-11·E-22 는 저장소의 사실이며 M6 가 어떤 일을 해도 바뀌지 않는다 — 리허설이 있든 없든 같은 값을 읽으므로(verification-completeness §2 의 도착 시점 적색·이후 적색) 그 값을 RED 로 쓰면 이 작업을 재지 못한다. 그래서 RED 는 산출물의 부재다 |
| 돌연변이 프로브 | squash 로 수렴하는 변이는 트리 항등은 통과하지만 "develop 팁이 main 의 조상" 에서 죽는다. 흡수 없이 병합하는 변이는 트리 항등에서 죽는다. 되돌리기를 시연하지 않는 변이는 되돌린 트리 단언에서 죽는다 |
| 녹색 경로 | M6 (스크래치 클론, GitHub 는 건드리지 않는다). E-27 명령이 `scripts/cutover-rehearsal.sh` 의 경로를 출력(종료 코드 0) |
| 통과 출력 | 세 비교 명령이 각각 같은 값을 출력하고 `git merge-base --is-ancestor` 가 종료 코드 0 |
| 트리 핀 | `855563dba79da74528e0f01560efe80f1016cc11` (E-27), `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-11·E-22·E-23, 원격 팁은 위 SHA 로 고정) |

### AC-GFD-019 — 배치 경계 사전 점검 (maps REQ-GFD-019)

- **분류**: release-blocking (M6 종료 게이트)
- **Given** 읽기 전용 사전 점검 스크립트 `scripts/cutover-precheck.sh` 와 픽스처 여섯 — 부정 넷: (1) 미푸시 develop 커밋 1개, (2) 살아 있는 통합 창 또는 슬롯 보유자, (3) 병합되지 않았거나 병합됐어도 push 되지 않은 picked 카드 브랜치(팁이 origin/develop 의 조상이 아니다), (4) 활성 레인 세션. 긍정 둘: (5) 모든 조건이 충족된 픽스처, (6) 이 카드 자신의 브랜치 팁이 origin/develop 의 조상이 아니지만 `--exclude-card t1453` 을 받은 픽스처(제외가 출력에 이름으로 남는다 — design D-25)
- **When** 런북 3단계(M4·M5 묶음의 병합·push 뒤, 수렴 병합 앞)에서 점검을 실행하면
- **Then** (5)(6) 만 종료 코드 0 이고, (1)~(4) 는 각자 위반한 조건을 이름으로 대며 종료 코드가 0 이 아니다. (6) 에서 `--exclude-card` 없이 같은 픽스처를 돌리면 (3) 으로 실패한다(제외는 명시된 한 카드뿐이다).

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-26 (종료 코드 1, 점검 시험·스크립트가 없다). 맥락 행(RED 가 아니다): E-12 (종료 코드 0, 카드 트리가 develop 팁보다 122 커밋 앞서 있어 "푸시 완료" 조건이 지금 충족되지 않는다) |
| RED 인 이유 | 이 기준이 요구하는 **산출물**(`scripts/cutover-precheck.sh` 와 `TestCutoverPrecheck`)이 없다. E-12 의 122 는 저장소 사실이고 M6 가 점검을 만들어도 같은 값이다(점검의 부정 픽스처가 바로 이 상태를 모사한다) — 그래서 RED 는 산출물의 부재다 |
| 돌연변이 프로브 | 발산만 보는 변이는 창·카드·세션 픽스처에서 죽는다. 항상 0 을 내는 변이는 부정 픽스처에서 죽는다. 모든 picked 카드를 제외하는 변이는 (6) 의 `--exclude-card` 없는 형태와 (3) 에서 죽는다. 이 카드를 암묵적으로 제외하는 변이는 `--exclude-card` 없는 (6) 에서 죽는다 |
| 녹색 경로 | M6 |
| 통과 출력 | `go test -v -run '^TestCutoverPrecheck$' ./internal/template/` 에서 하위 시험 6개 `--- PASS`, 방문 수 6. E-26 명령이 시험 파일 경로를 출력(종료 코드 0) |
| 트리 핀 | `855563dba79da74528e0f01560efe80f1016cc11` (E-26), `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-12) |

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
| F-red-4 | 백틱 없는 맨 `origin/develop` 산문 "then merge into origin/develop when green" | 위반 |
| F-red-5 | 동사 제외와 구별: "push to develop" | 위반 |
| F-red-6 | "Check out develop, pull, delete the local branch" | 위반 |
| F-red-7 | 접두 ref "compare against refs/remotes/origin/develop" | 위반 |
| F-red-8 | CJK 앞 이웃 "develop을 기준으로 만든다" | 위반 |
| F-red-9 | 표지 변이 "카드 브랜치는 develop 에서 만든다 (legacy)" | 위반(`legacy` 는 표지가 아니다) |
| F-red-10 | 제목 표지의 효력 한계: `## RETIRED 2026-10-02 old chain` 절 안의 줄은 면제, **그 다음 같은 수준 제목** `## Current rules` 아래 "카드는 develop 에 병합한다" | 위반 |
| F-red-11 | 하이픈 접미 "card work flows on develop-based worktrees" (`AGENTS.md:76` 형) | 위반 |
| F-red-12 | 분기 말 토큰이 없는 살아 있는 문장 "Work starts from develop." (약한 적중, 허용 목록 없음) | 위반 |
| F-red-13 | 토큰 없는 "Lane PRs go to develop." (동사 제외에 걸리지 않는다 — `to develop` 뒤에 한정사가 없다) | 위반 |
| F-red-14 | 약한 적중 "Compare with the develop tip." 에 허용 목록이 줄의 **일부** `develop tip` 만 담은 항목을 가진 경우 | 위반(`Literal` 은 줄 전체여야 한다) |
| F-green-1 | 폐기 표지 줄 "RETIRED 2026-10-02: card branches fork from develop" | 통과 |
| F-green-2 | 식별자 `manager-develop` | 통과 |
| F-green-3 | 디렉터리명 `.claude/worktrees/develop` | 통과 |
| F-green-4 | 동사 "to develop a feature" | 통과 |
| F-green-5 | 스탬프 `develop @ 0cca34439` | 통과 |
| F-green-6 | 구조 식별자 `--develop-worktree` | 통과(별도 차선에서 추적) |
| F-green-7 | 구조 식별자 `push-develop`·`push_develop` | 통과(별도 차선에서 추적) |
| F-green-8 | 표지 있는 제목 절과 그 하위 제목 아래의 develop 줄들, 다음 같은 수준 제목 앞까지 | 통과 |
| F-green-9 | 식별자형 이웃 "the develop-worktree helper and develop-notes" | 통과 |
| F-green-10 | 약한 적중 "Compare with the develop tip." 에 허용 목록이 **줄 전체** `Compare with the develop tip.`(+ `Why`)를 담은 항목을 가진 경우 | 통과 |
| F-empty | 방문 파일 0 | 가드 실패(빈 스윕) |

- **When** `go test -v -run '^(TestGitHubFlowSweepFixtures|TestGitHubFlowSweepGuard)$' ./internal/template/` 를 실행하면
- **Then** 픽스처 **25행**(F-red 14·F-green 10·F-empty 1) 모두 기대대로이고 가드 본체는 `visited=<N>` 과 하위 트리별 방문 수를 출력하며 `--- SKIP` 이 없다. 가드 본체의 바닥값은 AC-GFD-011 의 숫자(전체 1372, 하위 트리별 6·101·19·229·2·558·33·367·5·48)다. 이 픽스처 행들은 이번 계획 회차의 스크래치 프로브가 같은 패턴과 약한 적중 승격 설계로 이미 관측했다(원장 M-9: 25행 `bad=0`, F-empty 는 방문 0 이면 가드가 실패한다는 규칙을 프로브가 직접 단언했다 — 실제 가드의 `--- FAIL` 형태는 아직 없다).

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-13 (종료 코드 1, 가드 시험 부재) |
| RED 인 이유 | 가드가 없다. 시험 선택 `-run` 은 존재하지 않는 시험에서도 종료 코드 0 과 `[no tests to run]` 을 내므로 이 기준은 시험 이름이 아니라 방문 수와 `--- PASS` 줄을 요구한다 |
| 돌연변이 프로브 | 전부-일치 변이(항상 위반)는 F-green 행에서 죽고, 전부-불일치 변이(항상 통과)는 F-red 행에서 죽으며, 빈 스윕을 통과로 읽는 변이는 F-empty 에서 죽는다. CJK 를 `\b` 로 처리하는 변이는 F-red-2·F-red-8 에서 죽는다. 왼쪽 경계가 `/` 이웃을 제외하는 1회차 설계의 변이는 F-red-3·F-red-4·F-red-7 에서 죽는다. 오른쪽 경계가 모든 하이픈 이웃을 제외하는 1회차 설계의 변이는 F-red-11 에서 죽는다. 동사 제외를 `to develop` 전체로 넓히는 변이는 F-red-5 에서 죽는다. `to develop` 제외를 없애는 변이는 F-green-4 에서 죽는다. 제목 표지의 효력을 문서 끝까지 늘리는 변이는 F-red-10 에서 죽는다. **약한 적중을 개수로만 보고하는 변이(0.1.1 설계)는 F-red-12·F-red-13 에서 죽는다. 허용 항목의 `Literal` 을 줄 일부와 대조하는 변이는 F-red-14 에서, 허용 항목이 있는데도 위반으로 읽는 변이는 F-green-10 에서 죽는다.** 허용 목록 전체를 무시하는 변이는 F-green-10 에서, 맨 토큰 항목을 받아들이는 변이는 F-red-14 에서 죽는다 |
| 녹색 경로 | M4 (픽스처 시험은 PRE-CUTOVER-SAFE 로 먼저 병합 가능, 트리 단언은 절체 시점에 무장 — 무장 실행의 종료 증거는 AC-GFD-011 의 M5) |
| 통과 출력 | 하위 시험 25행 `--- PASS`(F-red 14·F-green 10·F-empty 1), 가드 본체 `visited=<N>` 출력(N ≥ 1372)과 `--- SKIP` 없음 |
| 트리 핀 | `4bf547bcad7c155b1e91485921569db709ec3ac2` (E-13), `855563dba79da74528e0f01560efe80f1016cc11` (M-5), `6c2277295d9ddeaa92e83c0225910445f83cc1c2` (M-9) |

### AC-GFD-022 — 병합 큐 판정과 t1452 조정의 기록 (maps REQ-GFD-022)

- **분류**: regression-guard (문서 존재와 리더 소관의 후속 기록은 이 트리에서 RED 를 다시 실행하는 형태가 아니다)
- **Given** `design.md` §D-3 의 병합 큐 증거 표(재현된 행만)와 §D-11 의 t1452 처분
- **When** sync 감사가 `progress.md` §G 와 `research.md` §9 를 읽으면
- **Then** (1) 증거 표의 미측정 행(청구 분, CodeRabbit 한도 이력, 의미 충돌 빈도)이 후속 측정 항목으로 `research.md` §9 에 있다. (2) 병합 큐 측정 카드를 발행하라는 **권고가 `progress.md` §G 에 기록**돼 있다. 카드의 발행은 리더의 일(`kanban-dispatch.md` — 리더가 큐의 유일한 생산자)이므로 **리더가 카드를 발행하면 그 id 를 같은 곳에 적는다** — id 가 아직 없다는 것은 이 기준의 실패가 아니다. (3) t1452 의 (a)(b) 가 은퇴·(c) 가 유효로 분류돼 있다.

| 셀 | 내용 |
|---|---|
| 회귀 방지 신호 | 표의 미측정 행이 모두 `research.md` §9 의 후속 측정 항목에 있다 |
| 한계 | 병합 큐의 효과는 이 SPEC 에서 측정하지 않는다. 이 기준은 M2 의 종료 증거가 아니라 M6(판정서 작성) 시점에 읽는다 — 이 카드가 만들 수 없는 행위(카드 발행)에 의존하지 않도록 한 개정 |
| 트리 핀 | `855563dba79da74528e0f01560efe80f1016cc11` |

### AC-GFD-023 — 런북 내용: 레인 재기동 순서와 단계 표지 (maps REQ-GFD-019)

- **분류**: release-blocking (M6 종료 게이트)
- **Given** M6 의 런북 문서 `.moai/specs/SPEC-GITHUB-FLOW-DEFAULT-001/cutover-runbook.md` 와 그 표를 읽는 시험 `TestCutoverRunbookShape`
- **When** 시험이 런북의 단계 표를 파싱하면
- **Then** (1) 단계가 이 순서로 있다 — 선행 조건 확인(origin/develop 팁 CI), 레인 정지·`/clear`·정리, 묶음 병합과 push, 사전 점검(AC-GFD-019), develop 의 main 흡수, 수렴 PR 병합, 트리 항등·조상 확인, 레인 재기동, 첫 카드 관측, develop 퇴역 단계. (2) 모든 단계 행에 실행 주체 열(**정확히 한 값** — 카드·리더·운영자 중 하나)과 외부 공유 시스템 여부 열(`예`·`아니오` 로 시작)이 있다. 한 단계에 두 주체가 필요하면 접미 번호 행으로 나뉘어 있다(design D-29: `5a` 리더·`5b` 운영자, `9a` 운영자·`9b` 리더). (3) 운영자 수행 단계(수렴 PR 병합 `5b`, `Release PR Multi-OS Gate` 필수 체크 제거·develop 보호·삭제 `9a`)는 모두 주체가 운영자이고 외부 공유 시스템이 예다. (4) 퇴역 뒤 점검 문구 — `git grep -n -w develop -- .github/workflows` 가 종료 코드 1 이어야 한다는 것과 `gh run list --workflow=CI --branch develop --limit 3` 의 가장 새 `createdAt` 이 전진하지 않아야 한다는 것 — 이 있다. (5) 이 카드 자신의 면제(design D-25)와 점검이 2단계 병합 뒤 4단계 앞에서 돈다는 순서가 있다.

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-32 (종료 코드 1, 런북 형태 시험이 없다) |
| RED 인 이유 | 이 기준이 요구하는 **산출물**(런북 문서와 `TestCutoverRunbookShape`)이 없다 |
| 돌연변이 프로브 | 단계 순서를 바꾸는 변이(레인 재기동을 수렴 앞에 두는 것)는 순서 단언에서 죽는다. 주체 열이나 외부 공유 시스템 열이 빠진 행은 (2) 에서 죽는다. **두 값을 한 셀에 담은 행(설계 표 0.1.1 의 `운영자·리더`, `운영자(…), 리더(…)` 처럼 접미 번호 없이 합친 행)은 (2) 의 "정확히 한 값" 단언에서 죽는다.** 운영자 단계의 주체를 카드로 적는 변이는 (3) 에서 죽는다. 퇴역 뒤 점검 문구가 없는 변이는 (4) 에서 죽는다. 점검을 4단계 뒤에 두는 변이는 (5) 에서 죽는다 |
| 녹색 경로 | M6 |
| 통과 출력 | `go test -v -run '^TestCutoverRunbookShape$' ./internal/template/` 하위 시험 5개(위 (1)~(5)) `--- PASS`. E-32 명령이 시험 파일 경로를 출력(종료 코드 0) |
| 트리 핀 | `855563dba79da74528e0f01560efe80f1016cc11` (E-32) |

### AC-GFD-024 — 계약에서 미션으로의 투영이 쓰는 병합 목표는 구성에서 나온다 (maps REQ-GFD-023)

- **분류**: release-blocking (M2a 종료 게이트)
- **Given** 서명 유효 계약 픽스처(`signedValidMissionFixture`)와 병합 목표 인자 행 다섯 — `develop`(git-flow 구성의 통합 목표), `main`(github-flow), `staging`(gitlab-flow 처럼 `main` 도 `develop` 도 아닌 값), 빈 문자열(목표 없음), 공백뿐인 문자열
- **When** 투영 `ProjectToMission(contract, mergeTarget)` 을 각 행에서 실행하고 비어 있지 않은 행은 `mission.SealMissionContract` 로 봉인하면
- **Then** (1) 비어 있지 않은 행은 투영된 계약의 `MergeTarget` 이 인자와 같다(`develop`→`develop`, `main`→`main`, `staging`→`staging`) — 어떤 행에서도 인자와 다른 리터럴이 나오지 않는다. (2) 빈 행과 공백뿐인 행은 `errors.Is(err, ErrNotProjectable)` 가 참이고 메시지가 `merge_target` 을 이름으로 대며 쓸 수 있는 미션 계약을 돌려주지 않는다 — `develop` 대체값도 `main` 기본값도 없다. (3) git-flow 행(`develop`)의 봉인 해시가 변경 전 투영이 같은 픽스처에 냈던 해시 리터럴과 같다(변경 전 트리에서 특성화 시험이 먼저 고정한다 — 아래 녹색 경로) — 투영 결과가 바이트 단위로 현행과 같다는 증명이 이것이다. (4) 동결 표면이 그대로다 — 스냅숏 비교(`TestContractProjectsOntoMissionValidator` 의 limb 5)가 초록이고, 원장 E-51·E-53 의 명령이 변경 전과 같은 출력(E-53 은 줄 문언 기준)을 내며, 스냅숏 파일과 `internal/mission` 의 어느 파일도 편집되지 않았다. (5) 계약 코어는 `internal/config` 를 import 하지 않는다(원장 E-52 가 종료 코드 1 과 빈 출력). 이 기준은 스냅숏 파일을 **편집하지 않는다** — 편집이 필요해 보이면 그 편집은 이 기준이 허용하는 일이 아니라 마일스톤을 멈추고 리더에게 올리는 사유다(design D-30)

| 셀 | 내용 |
|---|---|
| RED-now | 원장 E-46 (종료 코드 0, 투영의 병합 목표가 리터럴 `develop` 상수 한 줄 — 82 줄) 과 E-47 (종료 코드 1, 이 기준을 관측하는 `TestProjectionMergeTarget…` 시험이 없다). 맥락 행: E-48(비테스트 호출자 없음, 호출 지점은 시험의 네 곳), E-49(`MergeTarget` 은 비어 있지 않음만 검사되고 어떤 브랜치와도 비교되지 않는다), M-11(변경 전 투영 시험과 스냅숏 비교가 초록) |
| RED 인 이유 | 투영의 병합 목표가 구성과 무관하게 리터럴 `develop` 이라 `main`·`staging` 행의 기대와 빈 목표 거부가 거짓이고(E-46), 인자를 바꿔 보는 시험도 아직 없다(E-47). 이 두 사실은 이 마일스톤 자신의 편집(상수 한 줄을 지우고 인자를 받는 시그니처와 그 시험을 쓰는 일)만이 바꾼다 — 투영과 무관한 파일을 고쳐서 뒤집히는 적색이 아니다. 불가능한 기준도 아니다: 인자를 받는 투영은 E-46 을 종료 코드 1 로 뒤집고 위 행 다섯을 모두 통과시킨다 |
| 돌연변이 프로브 | (1) 인자가 비었을 때 `develop` 으로 되돌리는 변이는 E-46 의 `"develop"` 리터럴이 남아 종료 코드 0 이라 grep 에서 죽고, 리터럴을 쪼개 grep 을 피하는 변이(`"dev"+"elop"`)는 빈 행·공백 행의 거부 단언(`ErrNotProjectable`, `merge_target`)에서 죽는다. (2) 리터럴을 `"main"` 으로만 바꾼 변이는 `develop`·`staging` 행에서 죽는다. (3) 인자를 받지만 값에 쓰지 않는 변이는 `main`·`staging` 행에서 죽는다. (4) 빈 목표를 거부하되 `ErrNotProjectable` 로 감싸지 않거나 필드 이름을 대지 않는 변이는 (2) 의 `errors.Is`·부분 문자열 단언에서 죽는다. (5) 공백뿐인 목표를 걸러 내지 않는 변이는 공백 행에서 죽는다(봉인 단계의 `incomplete_contract` 가 아니라 투영에서 거부돼야 한다). (6) 계약 코어가 `internal/config` 를 직접 읽게 하는 변이는 E-52 가 종료 코드 0 이 돼 죽는다 — 그 변이는 인자 행에 영향이 없으므로 grep 만이 잡는 변이이고 시험이 보는 것은 값이다. (7) 동결된 행동 식별자를 개명하는 변이(`local-merge-develop`→다른 이름, `local_develop_merge` 변경)는 E-53 의 줄 문언 비교와 limb 5(스냅숏 45 줄)에서 죽는다. (8) 스냅숏을 손으로 다시 생성해 미션 표면 변경을 숨기는 변이는 `git diff --name-only 4bf547bca… <M2a 종료 트리> -- internal/mission internal/contract/testdata` 가 비어 있어야 한다는 범위 가드에서 죽는다(E-54 가 도착 시점에 빈 출력, E-55 가 같은 형태의 양성 대조). (9) `internal/mission/git_owner.go:201` 의 `branch != "develop"` 까지 `main` 으로 옮기는 변이는 같은 범위 가드에서 죽는다 — 그 변이는 github-flow 에서 `main` 으로의 로컬 병합을 켜므로 REQ-GFD-004·006 과도 충돌한다. 양방향 입력(거부 행 둘, 통과 행 셋)을 함께 두므로 전부-통과 변이는 빈·공백 행에서, 전부-거부 변이는 `develop`·`main`·`staging` 행에서 죽는다 |
| 녹색 경로 | M2a. 순서: (1) 변경 전 트리에서 git-flow 행의 봉인 해시를 특성화 시험으로 고정하고(리터럴 `develop` 과 같은 값이므로 도착 시점에 초록), (2) 다섯 행의 표 기반 RED 를 쓰고(상수가 있는 동안 `main`·`staging` 행과 거부 행이 적색), (3) `ProjectToMission` 이 병합 목표를 인자로 받게 하고 상수를 지운다, (4) 시험의 호출 지점 네 곳을 같은 변경으로 고친다 |
| 통과 출력 | `go test -v -run '^TestProjectionMergeTargetFollowsIntegrationTarget$' ./internal/contract/` 가 하위 시험 6개(통과 행 셋·거부 행 둘·git-flow 해시 동일) `--- PASS`, 방문 수 6, `--- SKIP` 없음. `go test -v -run '^TestContractProjectsOntoMissionValidator$' ./internal/contract/` 가 limb 다섯 `--- PASS`(M-11 과 같은 모양). E-46 명령은 종료 코드 1 과 빈 출력, E-52 는 종료 코드 1 과 빈 출력, E-51 은 변경 전과 같은 한 줄 |
| 트리 핀 | `eda61419564296892b41d0e7608c8b136e52c04c` (E-46~E-55·M-11) |

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
| AC-GFD-011 | REQ-GFD-011 | release-blocking | M5 (M4+M5 묶음의 무장 실행) |
| AC-GFD-012 | REQ-GFD-012 | release-blocking | M4 |
| AC-GFD-013 | REQ-GFD-013 | release-blocking | M4 |
| AC-GFD-014 | REQ-GFD-014 | regression-guard | M4 |
| AC-GFD-015 | REQ-GFD-015 | regression-guard | M5 (M1~M5 의 편집마다 돌리고 종료 증거는 M5) |
| AC-GFD-016 | REQ-GFD-016 | release-blocking | M5 |
| AC-GFD-017 | REQ-GFD-017 | release-blocking | M4 |
| AC-GFD-018 | REQ-GFD-018 | release-blocking | M6 |
| AC-GFD-019 | REQ-GFD-019 | release-blocking | M6 |
| AC-GFD-020 | REQ-GFD-020 | regression-guard | M6 |
| AC-GFD-021 | REQ-GFD-021 | release-blocking | M4 |
| AC-GFD-022 | REQ-GFD-022 | regression-guard | M6 |
| AC-GFD-023 | REQ-GFD-019 | release-blocking | M6 |
| AC-GFD-024 | REQ-GFD-023 | release-blocking | M2a |

## §E 이 SPEC 이 통과로 기록하지 않는 것

| 항목 | 이유 |
|---|---|
| 임베드 축 (`make build` 가 싣는 내용) | 리더가 배치 종료 때 빌드하므로 이 SPEC 의 관측이 아니다 |
| 병합 큐의 효과 | 판정(보류)만 하고 측정은 후속 카드 |
| main push CI 의 벽시계·성공률, `gh run list`·`gh pr list` 목록 호출의 출력 | 같은 명령이 호출마다 다른 창을 돌려주는 것이 이 세션에서 관측됐다(research.md §4 "목록 호출 불안정"). 실행 id·PR 번호로 고정하지 못한 수치는 측정값으로 인용하지 않고 research.md §9 Gaps 에 둔다(design D-28) |
| GitHub 쪽 PR 생성·대형 diff 병합 | 스크래치 클론 리허설은 git 층만 시연한다. GitHub 층은 리더·운영자가 절체 경계에서 관측한다 |
| 운영자가 수행하는 `constitution amend` 의 실제 실행 | 5단째 인간 승인이 대화형이라 레인이 실행하지 않는다. 레인은 입력과 `--dry-run`(가능하면)과 사후 `validate` 만 관측한다 |
| main 필수 체크 목록에서 `Release PR Multi-OS Gate` 제거 | 운영자 수행 단계(외부 공유 시스템). 레인은 런북에 적을 뿐이고 이 SPEC 은 제거 후 상태를 관측하지 않는다 |
| 병합 큐 후속 측정 카드의 id | 카드 발행은 리더의 일이다. 이 SPEC 은 권고를 `progress.md` 에 기록하고 id 는 발행 뒤에 적힌다(AC-GFD-022) |
| 계약→미션 투영의 운영 호출 경로, `internal/mission/git_owner.go:201` 의 실행 동작 | 투영에 비테스트 호출자가 없고(원장 E-48) `git_owner.go:201` 은 읽기만 했다(원장 E-50). 이 SPEC 은 투영의 값과 동결 표면을 단위 수준에서 판정하고, 투영이 운영에 연결될 때의 호출 쪽 구성 읽기는 그 연결을 만드는 카드의 몫이다(design D-30) |
| 설치된 `moai` 로 낸 판정 | 설치본(v3.2.0-rc.25, 커밋 802a72235)은 이 트리의 엄격한 조상이라 판정에 쓰지 않았고 이 트리에서 빌드한 `moai` 를 경로로 호출했다 |
| rc 태그의 실제 릴리스 | 태그는 불변 규칙셋으로 되돌릴 수 없어 이 카드에서 만들지 않는다 |
| 상시로드 비용의 토큰 환산 | 호스트가 관측을 노출하지 않으면 바이트만 보고한다 |

---

🗿 MoAI

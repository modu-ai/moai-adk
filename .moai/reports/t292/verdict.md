# t292 M1 판정 — main codemaps 고아 스탬프, 전용 수리가 필요한가

- 카드: t292
- 브랜치: `WT-stamp-orphan-repair`
- 트리: `.claude/worktrees/t292` (base `origin/develop` = `6310dbf28`)
- 측정 시각 기준 refs: `origin/main` = `48239c7dc`, `origin/develop` = `6310dbf28`

## 판정

**(a) 다음 릴리스 머지가 자연 해소하므로 전용 수리는 불필요 — 카드 철회 권고.**

여기서 멈추고 보고합니다. M2 수리에 착수하지 않았습니다. 카드 철회·변경은 운영자 몫입니다.

---

## Claim

1. 카드가 겨눈 고아 스탬프는 **실재한다**. `origin/main`의 `.moai/project/codemaps/provenance.json`이 기록한 `a995e58fa`는 `origin/main`에서 도달 불가다.
2. `origin/develop`은 **깨끗하다**. 스탬프 `d2fcecc8b`는 `origin/develop`에서 도달 가능하다.
3. `origin/main`은 `origin/develop`의 **엄격한 조상**이며, main이 앞선 커밋은 **0개**다. 따라서 develop→main 통합에서 이 파일이 충돌하는 것은 **구조적으로 불가능**하고, 통합 직후 main의 스탬프는 `d2fcecc8b`가 되어 고아가 소멸한다.
4. 릴리스 브랜치는 **`develop`에서 딴다**([HARD] 독트린). 그래서 릴리스 PR 자신은 초록이고, 그 머지가 곧 수리다.
5. 카드가 적은 피해("main 파생 신규 브랜치의 레드 상속")는 **이미 소멸했다**. 카드 단위 PR이 폐지되고 레인이 develop에서 분기하기 때문이다.
6. 릴리스 전까지 main에서 분기하는 경로는 **hotfix 하나뿐**이고, 그 경로에서도 graph-freshness는 **머지를 막지 못한다** — required status check 목록에 없다.
7. 카드가 지정한 목표 팁 `26c5a7d54`는 낡았다. 현재 main은 `48239c7dc`다.

## Evidence

리드 값을 이월하지 않고 직접 재측정했습니다.

**(1) main 스탬프와 도달 불가**

```
$ git show origin/main:.moai/project/codemaps/provenance.json
  "tree_root": "/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t289",
  "commit_sha": "a995e58fa69b40f3d76a76c0a5ca56d0594fb528",
  "dirty": false,
  "generated_at": "2026-08-26T18:27:30Z"

$ git merge-base --is-ancestor a995e58fa69b40f3d76a76c0a5ca56d0594fb528 origin/main
rc=1
```

부가 관측 — 이 커밋은 **전역 고아가 아니다**. 객체로 존재하고 `origin/develop`에서는 도달 가능하다. 고아인 것은 오직 main 기준이다:

```
$ git cat-file -t a995e58fa69b40f3d76a76c0a5ca56d0594fb528
commit
$ git branch -r --contains a995e58fa69b40f3d76a76c0a5ca56d0594fb528
  origin/develop
  origin/WT-glm-flash-default
  origin/WT-statusline-cost
```

**(2) develop 스탬프와 도달 가능**

```
$ git show origin/develop:.moai/project/codemaps/provenance.json
  "commit_sha": "d2fcecc8b40d1cb388efc7ed19b6e354b1bff8ab",
  "generated_at": "2026-08-27T11:14:53Z"

$ git merge-base --is-ancestor d2fcecc8b40d1cb388efc7ed19b6e354b1bff8ab origin/develop
rc=0
$ git merge-base --is-ancestor d2fcecc8b40d1cb388efc7ed19b6e354b1bff8ab origin/main
rc=1
```

**(3) 조상 관계와 divergence**

```
$ git merge-base --is-ancestor origin/main origin/develop
rc=0
$ git rev-list --count --left-right origin/main...origin/develop
0	173
```

좌항 0 = main이 develop에 없는 커밋을 **하나도** 가지지 않음. main은 이 파일을 마지막 공통 상태 이후 건드린 적이 없으므로, develop→main 통합에서 `provenance.json`은 develop 쪽 값으로 무조건 결정된다. 3-way 머지가 충돌을 낼 여지가 없다.

**(4) 릴리스 경로의 분기점**

```
$ git show origin/develop:.moai/docs/git-workflow-doctrine.md | grep -n "release/"
51:- [HARD] Release branch `release/vX.Y.Z` is cut **from `develop`**, never from `main`.
55:- [HARD] **Hotfix is the single exception**: `hotfix/vX.Y.Z-*` is cut **from `main`**
101:| release/* → main | **merge commit** | `gh pr merge N --merge` |
```

릴리스 브랜치가 develop에서 갈라지므로 그 head 트리의 스탬프는 `d2fcecc8b`이고, head 이력 안에서 도달 가능하다 — 릴리스 PR의 graph-freshness는 초록이다. 그 PR이 머지되는 순간 main의 스탬프도 `d2fcecc8b`로 바뀌고, 동시에 그 커밋이 main에서 도달 가능해진다.

**(5) 카드가 적은 피해의 소멸**

```
$ git show origin/develop:.moai/docs/git-workflow-doctrine.md | grep -n "카드 단위 PR"
113: [HARD] POLICY CHANGE (2026-08-27) — 카드 단위 PR 폐지. (…) 카드 워크트리는
     develop에서 분기하고 (…) origin/develop의 CI가 통합을 판정한다
```

실측으로도 develop push의 graph-freshness는 최근 5회 연속 초록이다:

```
$ gh run list --workflow=graph-freshness.yml --limit 12 --json conclusion,event,headBranch
develop push × 5 연속 success (11:20 / 11:43 / 11:53 / 13:19 / 13:22 Z)
```

**(6) 잔여 노출과 그 무해성**

graph-freshness 트리거는 `pull_request` (base 무관) + `push: [main, develop]`이다. 그러므로 main-push는 지금 실제로 레드다:

```
$ gh run list --workflow=graph-freshness.yml --branch main --limit 8
2026-08-26T18:56 failure / 18:30 failure / 18:20 failure / 18:06 success …
```

릴리스 전까지 main에서 분기하는 유일한 경로는 hotfix([HARD] 단일 예외)이고, 그 PR은 main 파생이므로 같은 레드를 상속한다. 그러나 그 레드는 **머지를 막지 못한다**:

```
$ gh api repos/modu-ai/moai-adk/branches/main/protection --jq .required_status_checks.contexts
["Test (ubuntu-latest)", "Lint", "Build (linux/amd64)", "Analyze (Go) (go)", "Release PR Multi-OS Gate"]
```

`Graph Freshness`는 required 목록에 없다. hotfix PR에서도 권고 성격의 레드 한 줄이 뜰 뿐, 24h SLA 경로가 차단되지 않는다.

**(7) 카드 목표 팁의 노후화**

```
$ git rev-parse origin/main
48239c7dc7428c8751a04f6321887c2d36123884
```

카드 본문의 `26c5a7d54`는 현재 main보다 2커밋 뒤다.

## Baseline-attribution

- 모든 측정은 이 세션에서 실행했다. 트리는 `.claude/worktrees/t292`, HEAD `6310dbf28`.
- 원격 참조는 측정 직전 `git fetch origin main develop`로 갱신했다: `origin/main` = `48239c7dc`, `origin/develop` = `6310dbf28`.
- 리드 메시지가 제시한 네 값은 근거로 인용하지 않았다. 위 rc/출력은 전부 이 실행의 관측이며, 결과적으로 리드 값과 일치했다.
- CI 판정은 `gh run list` / `gh api` 응답이다. `Graph Freshness` 워크플로 정의는 `origin/develop` 판(`push: [main, develop]`)을 읽었다 — primary 체크아웃의 사본은 main 판(`push: [main]`)이라 트리거 축이 다르다.

## Gaps

관측하지 **않은** 것:

- **다음 릴리스의 실제 시점**을 확인하지 않았다. 릴리스가 임박한지, 몇 주 뒤인지 측정 근거가 없다. 판정 (a)는 "릴리스가 언젠가 온다"에만 의존하고 시점에는 의존하지 않지만, main-push 레드가 유지되는 기간의 길이는 미지수다.
- **릴리스 머지를 실제로 시뮬레이션하지 않았다.** 충돌 불가는 divergence `0 173`에서 연역했을 뿐, 시험 머지를 돌려 확인하지는 않았다.
- **`moai graph check`를 main 트리에 대해 실행하지 않았다.** main-push 레드의 원인이 이 스탬프라는 것은 CI 실패 시각·워크플로 정의·도달불가 rc의 정황 일치이며, 실패 로그 본문을 열어 원인 문자열을 확인하지는 않았다.
- **hotfix 경로를 실제로 시험하지 않았다.** hotfix PR이 레드를 상속한다는 것은 트리거 정의와 분기점 독트린에서 연역했다.
- **`26c5a7d54` 이후 main에 `provenance.json`을 건드린 커밋이 있는지** 파일 단위 로그로 훑지 않았다. divergence 좌항 0이 이를 함의하므로 중복 측정으로 판단했다.

## Residual-risk

- **릴리스가 hotfix보다 늦게 오는 순서**라면, hotfix PR 작성자가 레드 한 줄을 보고 원인을 찾느라 시간을 쓴다. 차단은 아니지만 긴급 경로에서의 인지 비용이다. 이는 수리가 아니라 **기록**으로 더 싸게 막을 수 있다 — 알려진 무해 레드라는 메모 한 줄.
- **`Graph Freshness`가 나중에 required check로 승격되면** 판정 (a)의 근거 하나가 뒤집힌다. 승격은 branch protection 변경이므로 관측 가능한 사건이며, 그때 카드를 되살리면 된다.
- 전용 수리를 지금 진행할 경우의 반대 위험: main에 커밋하려면 PR이 필요하고(`enforce_admins: true`), 그 PR이 만드는 main 커밋은 다음 릴리스 머지 때 **어차피 develop 값으로 덮인다**. 즉 수리 작업물의 수명이 릴리스까지로 한정된다.

## 권고

카드 t292 철회. 대신 다음 둘 중 하나를 운영자 판단에 올립니다.

1. 아무것도 하지 않고 다음 릴리스에 맡긴다 (권장). 비용 0.
2. main-push의 graph-freshness 레드가 "릴리스 전까지 알려진 무해 상태"임을 한 줄로 기록해, hotfix 작성자가 원인 추적에 시간을 쓰지 않게 한다.

`Graph Freshness`가 required check로 승격되면 그때 재평가가 필요합니다.

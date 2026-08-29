# t294 조사 — graph-freshness.yml `pull_request` 무필터 트리거

- 카드: t294 (Tier S, Class A~B)
- 측정 트리: `.claude/worktrees/t294`, 브랜치 `WT-freshness-trigger`, HEAD `51daada00` (origin/develop 조상 확인됨)
- 단계: plan (run 미진입 — Implementation Kickoff Approval 별도)

> **정정 이력**: §2.3은 최초 작성 후 정정됐다. `pull_request`의 `branches:`가 base를 거른다는 사실을
> 확인하면서 "무필터 → 이중 발화" 인과가 반증됐다. §1의 의도/누락 판정은 영향받지 않는다.

## 1. Claim — 무필터는 의도가 아니라 누락이다

`graph-freshness.yml`의 `pull_request:` 무필터는 "모든 base 브랜치의 PR에서 돌린다"는 선택이 아니라,
**작성 시점에 base가 main 하나뿐이라 필터를 쓸 이유가 없었던 생략**이다. develop이 생기면서 같은 문장의
의미가 조용히 넓어졌다.

## 2. Evidence

### 2.1 현재 트리거 (측정 트리 기준)

`.github/workflows/graph-freshness.yml` L3-6:

```yaml
on:
  pull_request:
  push:
    branches: [main, develop]
```

`pull_request:`에는 `branches:`도 `paths:`도 없다.

### 2.2 provenance — 무필터가 develop보다 하루 먼저다

`git blame -L 1,8 -- .github/workflows/graph-freshness.yml`:

```
6786c3fa42 (Goos Kim 2026-08-26 4)   pull_request:
6786c3fa42 (Goos Kim 2026-08-26 5)   push:
11216d13f6 (t        2026-08-27 6)     branches: [main, develop]
```

- `6786c3fa4` (2026-08-26): 파일 최초 작성. 당시 개발 모델은 GitHub Flow이고 PR base는 main 하나뿐 —
  무필터 `pull_request:`와 `branches: [main]`은 그때 **동치**였다.
- `11216d13f` (2026-08-27): git-flow 전환. 커밋 메시지가 범위를 스스로 못박는다 —
  *"6 CI workflows: push trigger branches [main] -> [main, develop]"*. 즉 **push 줄만** 손댔고
  `pull_request` 블록은 6개 워크플로 어디에서도 건드리지 않았다(`git show 11216d13f -- .github/workflows/`
  전체 diff에서 `pull_request` 인접 변경 0건).

무필터를 유지하겠다는 판단이 어디에도 기록돼 있지 않고, 그 문장이 쓰인 시점에는 develop이 존재하지도 않았다.
따라서 **누락** 쪽으로 판정한다.

### 2.3 [정정] 이중 발화의 원인은 무필터가 **아니다**

최초 작성분은 "무필터라서 develop head PR에서 push·PR이 겹친다"고 적었으나, 이는 틀렸다.
`pull_request`의 `branches:`는 PR의 **base** 브랜치를 거른다 — head가 아니다. 이 저장소 실측으로 양방향 확인했다.

- `gh run list --workflow=ci.yml --event=pull_request --limit 5` → head가 `WT-glm-settings-persist`,
  `WT-main-stamp-repair` 등인 PR에서 CI가 **발화했다**. `ci.yml`의 필터는 `branches: [main]`이므로,
  필터가 head를 걸렀다면 이 런들은 존재할 수 없다 → base를 거른 것이다.
- `gh pr checks 1677` (유일한 base=develop PR, head `WT-statusline-cost`) → `graph-freshness`·`lsel-leak-guard`는
  발화했고 **CI는 목록에 없다**. `[main]` 필터가 base=develop을 걸러냈다.

따라서 develop→main PR은 base가 main이라 `[main]`이든 `[main, develop]`이든 **어느 필터에서도 발화한다**.
같은 head(develop)에 push 런이 함께 남는 것은 `concurrency.group`이 `github.ref` 키(push=`refs/heads/develop`,
PR=`refs/pull/N/merge`)라 서로 취소하지 못하기 때문이며, **필터를 명시해도 닫히지 않는다.**
게다가 이 중복은 graph-freshness 고유 문제가 아니다 — push에 develop이 들어간 6개 워크플로 전부가 같은 모양이다.

무필터가 실제로 만드는 차이는 따로 있다: **base가 main도 develop도 아닌 PR에서까지 무조건 발화한다**는 것.
오늘 그런 PR은 0건이고, base=develop PR은 1건(#1677)이라 `[main, develop]`은 현재 발화를 하나도 줄이지 않는 대신
develop-base PR의 사전 검사 커버리지를 유지한다.

### 2.4 실제 발화 관측 — 아직 이중 발화 없음

`gh run list --workflow=graph-freshness.yml --limit 30`: 30건 중 29건이 `push`/`develop`,
1건이 `pull_request`/`fix/heavy-gate-nested-toolchain`. head가 develop인 PR 런은 0건 —
카드 서술("지금은 무해, develop PR 한 번이면 드러남")과 일치한다.

## 3. Baseline-attribution

위 세 측정 모두 이 조사 턴에서, 워크트리 `.claude/worktrees/t294`(HEAD `51daada00`)에 대고 실행했다.
`gh run list`만 원격 상태(조회 시점 2026-08-29)를 읽는다.

## 4. Gaps — 카드의 범위 주장이 부분적으로 거짓이다

카드는 *"다른 5개 워크플로(ci·codeql·lsel-leak-guard·template-neutrality-check·test-install)는
pull_request에 branches: [main]이 명시돼 있어 대상 아님"* 이라고 적었다. **훑은 범위는 5개가 아니라
`.github/workflows/` 전체 18개 파일**(`grep -n -A6 '^  pull_request' .github/workflows/*.yml *.yaml`)이고,
결과는 다음과 같다.

| 워크플로 | `pull_request` 브랜치 필터 | 판정 |
|---|---|---|
| graph-freshness.yml | 없음 (paths도 없음) | **대상 — 카드 본건** |
| lsel-leak-guard.yaml | **없음** (`paths:`만) | **카드 주장 반증** |
| spec-lint.yml | 없음 (`paths:`만) | 카드 미열거 |
| docs-i18n-check.yml | 없음 (`paths:`만) | 카드 미열거 |
| ci.yml | `branches: [main]` | 대상 아님 |
| codeql.yml | `branches: [main]` | 대상 아님 |
| test-install.yml | `branches: [main]` | 대상 아님 |
| template-neutrality-check.yaml | `branches: [main]` | 대상 아님 |
| release-pr-multi-os.yml | `branches: [main]` | 대상 아님 |
| spec-status-auto-sync.yml | `types: [closed]`만 | 별개 성격 |
| claude.yml / community.yml | `pull_request_review*` / `pull_request_target` | 별개 이벤트 |
| auto-merge / label-sync / release-drafter* / release / review-quality-gate | `pull_request` 트리거 없음 | 대상 아님 |

즉 `lsel-leak-guard.yaml`은 `branches: [main]`을 갖고 있지 않다 — 카드의 5개 열거 중 1개가 틀렸다.
다만 세 워크플로(lsel-leak-guard·spec-lint·docs-i18n-check)는 `paths:`로 좁혀져 있어 노출 폭이
graph-freshness(무조건 발화)보다 작다.

### 4.1 선례 — 같은 자리를 이미 다룬 두 워크플로

`spec-lint.yml`과 `docs-i18n-check.yml`은 git-flow 전환 뒤 push에 develop을 넣으면서 그 이유를
주석으로 고정했다(예: *"develop is the integration branch under git-flow: lane work merges into it
directly, with no card PR to carry the pull_request trigger"*). 두 곳 모두 **`pull_request` 필터는
손대지 않았다** — 이 카드의 조치 형태를 정할 때 참고할 기존 선례다.

## 5. 미검증 / 잔여 위험

- **실행 기반 회귀 확인 불가**: 유일한 실행 관측은 develop을 head로 하는 PR을 실제로 여는 것인데 develop이
  동결 중이라 열 수 없다. 카드 브랜치 push는 CI-inert(lane-4 실측)라 대체 관측이 되지 않는다.
  → plan에서는 파서 기반 트리거 해석으로 잡고, 실행 관측 요구는 kickoff 때 올린다.
- **범위 확장 판단은 운영자 몫**: 카드 본건은 graph-freshness 1개다. 위 표에서 새로 드러난 3개
  (paths-scoped 무필터)를 같은 카드에서 함께 고칠지는 범위 결정이므로 이 조사에서 정하지 않는다.
- `e4eb15ea4`(t291 M2)가 같은 파일을 수정했으나 트리거 블록은 건드리지 않았다 — 이 카드와 충돌 없음.

# t315 / D7a — 배포 기본값 동작 변화 실측

측정 트리: `.claude/worktrees/t315`, 브랜치 `WT-release-notes-gitflow`, HEAD `b7462203a`(로컬 develop 흡수 후). 측정일 2026-09-02.

기준선은 **실제로 설치돼 있는 최신 공개 릴리스 `v3.1.2`**(2026-08-21)다. `v3.1.3`은 Draft
상태라 배포된 적이 없으므로 기준선이 될 수 없다.

---

## 1. 배포 기본값 — v3.1.2에서 실제로 나가는 값

| 측정 대상 | 명령 | 관측 |
|---|---|---|
| 사어 키 기본값 | `git show v3.1.2:internal/template/templates/.moai/config/sections/system.yaml.tmpl \| grep -n spec_git_workflow` | `31:  spec_git_workflow: main_direct` |
| 정본 키 기본값 | `git show v3.1.2:internal/template/templates/.moai/config/sections/git-strategy.yaml.tmpl \| grep -n 'workflow:'` | `13/45/81: workflow: github-flow` (3개 모드 전부) |
| 그때 delivery가 읽던 키 | `git show v3.1.2:…/sync/delivery.md \| grep -n spec_git_workflow` | `17: Read github.spec_git_workflow from .moai/config/sections/system.yaml` |

**여기가 결정적이다.** v3.1.2 시점에도 정본 키는 이미 `github-flow`를 싣고 나갔지만,
delivery는 그 키를 읽지 않고 사어 키(`main_direct`)를 읽었다. 즉 배포본 안에서 두 값이
서로 모순된 채 공존했고, 실효값은 사어 키 쪽이었다.

t303이 읽는 키를 정본으로 바꾸면서, **설정 파일을 한 글자도 안 고친 사용자의 실효값이
`main_direct` → `github-flow`로 바뀐다.** 사용자가 뭘 한 게 아니라 읽는 쪽이 바뀐 것이다.

현재 트리에서 사어 키는 템플릿 설정에서 제거됐다(`grep -rn spec_git_workflow
internal/template/templates/.moai/config/sections/system.yaml.tmpl` → rc=1). 남은 1건은
delivery.md의 폴백 센티넬뿐이며, 그 폴백은 **정본 키가 부재할 때만** 발화한다. 배포
git-strategy.yaml이 정본 키를 싣고 나가므로 일반 설치에서는 폴백이 발화하지 않는다.

## 2. 동작 대조 — 무엇이 실제로 달라지는가

카드 본문은 "PR을 만들지 않던 기본값이 이제 PR을 만드는 쪽으로 뒤집힌다"고 적었다. 재보니
**참이지만 조건부**다. 아래가 정확한 형태다.

**v3.1.2 `main_direct` 전문** (`delivery.md:250-255`)

```
##### Strategy: main_direct

All commits go directly to main, no PRs:
1. Push to main: `git push origin main`
2. Display push confirmation
3. No PR created regardless of branch name
```

`regardless of branch name` — 어느 브랜치에 있든 무조건 main으로 밀고 PR을 안 만든다.

**현재 `github-flow`** (`delivery.md:236-266`)

| 상황 | v3.1.2 (`main_direct`) | 현재 (`github-flow`) | 변화 |
|---|---|---|---|
| main 브랜치, Tier S/M, `--pr` 없음 | main으로 직접 push, PR 없음 | main으로 직접 push, PR 없음 | **없음** |
| main 아닌 브랜치 | (브랜치 무시하고) main으로 push, PR 없음 | 그 브랜치를 push하고 **PR 생성** | **있음** |
| 워크트리 컨텍스트 | 위와 같음 | 워크트리 브랜치 push + **PR 생성** | **있음** |
| 브랜치가 어느 경로에도 안 맞음(detached HEAD 등) | main으로 push (무조건 경로라 그대로 나감) | **중단** — PR도 push도 하지 않고 보고 | **있음** |

### 정정: 카드가 놓친 두 번째 변화

표의 마지막 행이 카드 본문에 없다. 종전에는 브랜치 상태가 뭐든 `main_direct`가 무조건
push했으므로 sync가 어떻게든 완주했다. 지금은 경로에 안 맞으면 **명시적으로 멈춘다**.
사용자 입장에서 "PR이 하나 더 생긴다"보다 "되던 sync가 이제 멈춘다"가 더 크게 체감될 수
있다 — 릴리스노트가 이 축을 빠뜨리면 안 된다.

또한 첫 행이 말해주듯 **모든 사용자가 PR을 받게 되는 것은 아니다.** 혼자 main에서 Tier
S/M SPEC만 돌리는 사용자(= `main_direct`를 쓰던 전형적 사용자)는 아무 변화도 겪지 않는다.
릴리스노트에 "이제 PR을 만듭니다"라고만 적으면 그쪽이 오히려 부정확하다.

## 3. 릴리스 대상 불일치 — D7a가 겨눈 릴리스가 아니다

카드와 t303 감사는 이 의무를 **v3.2.0**에 걸었다. 실측은 다르다.

| 측정 | 명령 | 관측 |
|---|---|---|
| t303 착지 위치 | `git merge-base --is-ancestor 7ed6edb3e origin/main` | rc=1 — main에 없음 |
| 릴리스 브랜치 포함 여부 | `git merge-base --is-ancestor 7ed6edb3e origin/release/v3.1.4` | **rc=0 — 포함됨** |
| 릴리스 PR | `gh pr view 1685` | `release/v3.1.4 → main`, state=**OPEN** |
| 릴리스 브랜치 버전 | `git show origin/release/v3.1.4:pkg/version/version.go` | `Version = "v3.1.4"` |
| CHANGELOG 구간 | `git show origin/release/v3.1.4:CHANGELOG.md \| grep -n '^## \['` | `[3.1.4] - 2026-08-31`은 10행, t303 항목은 **200행 = 그 구간 안** |

즉 이 동작 변화는 **v3.2.0이 아니라 v3.1.4(패치)에 이미 실려 있고**, 그 릴리스 PR은 지금
열려 있다. 두 가지가 따라온다.

1. **D7a는 지금 급하다.** #1685가 이대로 머지되면 v3.1.2 → v3.1.4로 올린 사용자가 §2의
   변화를 예고 없이 만난다 — D7a가 막으려던 바로 그 상황이다. "v3.2.0 준비 시점에 처리"
   라는 카드의 시점 서술은 대상 릴리스가 어긋나 있어 그대로 따르면 늦는다.
2. **SemVer 축이 따로 열린다.** 설정을 안 고친 사용자의 배포 동작이 바뀌는 것은 patch가
   아니라 minor 급 변화다. v3.1.4로 낼지 v3.2.0으로 번호를 올릴지는 릴리스 판단이라 이
   레인이 정하지 않고 리드에 보고한다.

## 4. 현재 CHANGELOG 항목이 D7a를 충족하는가 — 아니다

`[3.1.4]` 구간 200행의 t303 항목은 키 통일을 상세히 기록한다: 정본 키, 두 값 도메인
`{github-flow, git-flow}`, 미매칭 값의 명시적 중단, `WT-*` 라우팅, AC 12건.

하지만 **사용자가 겪는 문장이 없다.** "설정을 고치지 않은 프로젝트의 sync가 종전과 다르게
동작한다"는 진술도, §2 표의 네 행 중 어느 것도 그 항목에 없다. 있는 것은 구현 서술이고,
D7a가 요구하는 것은 동작 고지다. 개발자용 기록으로는 충분하고 릴리스노트로는 미달이다.

따라서 **D7a는 미충족**이다 — 항목이 없어서가 아니라, 있는 항목이 다른 질문에 답하고 있어서다.

## Gaps

- **G1** 이 레인은 `release/v3.1.4` 브랜치를 편집하지 않았고 push도 하지 않았다. §5의
  문안을 그 브랜치에 반영하는 것은 릴리스 소관이다.
- **G2** docs-site 4로케일에 이 동작 변화를 실을지 재지 않았다. t303 sync 기록은 "docs-site
  변경 없음 — 사어 키와 정본 경로가 docs-site에 나타나지 않는다"고 적었는데, 그건 *키*가
  없다는 뜻이지 *동작 고지*가 불필요하다는 뜻은 아니다. 별도 판정이 필요하다.
- **G3** 실제 사용자 프로젝트에서 upgrade를 실행해 §2 표를 end-to-end로 재현하지 않았다.
  표는 배포되는 두 tree의 `delivery.md` 원문 대조에서 나왔다.
- **G4** `git_strategy.mode` 기본값이 어느 모드인지는 확인했으나(3개 모드 모두 `github-flow`)
  모드별로 갈리는 다른 축(`main_branch` 등)은 이 카드 범위 밖이라 재지 않았다.

## Residual-risk

- **R1** §2 표는 `delivery.md`가 **지시하는** 동작이다. 이 파일은 에이전트가 읽는 지시문이지
  기계 코드가 아니므로, 실제 수행은 에이전트 해석에 달려 있다. 표를 "코드가 이렇게 분기한다"로
  읽으면 안 된다.
- **R2** `moai update`를 돌리지 않은 사용자는 아무 변화도 겪지 않는다(설정과 스킬이 함께 옛
  상태로 남는다). 변화는 update 시점에 발생하므로, 릴리스노트만이 유일한 고지 창구다.
- **R3** G2가 열려 있다 — docs-site에 실을지 판정되지 않았다.

---

## 5. D7a 이관 문안 (릴리스노트에 그대로 넣을 수 있는 형태)

아래는 §2 실측만 담았고 추정을 넣지 않았다. `[3.1.4]`(또는 번호가 올라간다면 그 구간)의
`### Changed` 성격 항목으로 넣는 것을 전제로 쓴 문안이다.

```markdown
### Changed — delivery behavior for projects that never edited their git config

The sync-phase delivery step now resolves its strategy from
`git_strategy.{mode}.workflow` instead of the retired
`github.spec_git_workflow`. Both keys shipped in previous versions, and they
disagreed: the shipped default for the retired key was `main_direct` while the
shipped default for the canonical key was, and remains, `github-flow`. Because
the retired key was the one actually read, **a project that never edited either
file was running `main_direct` and now runs `github-flow`.** No configuration
change on your side caused this; the key being read changed.

What that changes, precisely:

- **On your main branch, for a Tier S/M SPEC without `--pr` — nothing changes.**
  The sync commit is still pushed directly to the main branch and no pull
  request is created. If this describes how you work, you will see no
  difference.
- **On any branch other than main, a pull request is now created.** Previously
  `main_direct` pushed to the main branch and opened no pull request
  "regardless of branch name". Now the current branch is pushed and a pull
  request is opened against the main branch. This also covers worktree
  delivery.
- **A branch state matching no route now stops the delivery step.** Previously
  the unconditional `main_direct` push completed from any branch state,
  including a detached HEAD. Delivery now reports the branch state and creates
  no pull request and pushes nothing. **A sync that used to finish may now
  halt** — this is deliberate: the previous behavior pushed to the main branch
  from a state nobody had chosen.

To keep the previous behavior on non-main branches, set
`git_strategy.{mode}.workflow` explicitly for your active mode in
`.moai/config/sections/git-strategy.yaml`. Note that the value axis is now the
two-member domain `{github-flow, git-flow}` — `main_direct` is no longer a
strategy value; direct-to-main delivery is the `github-flow` main-branch route
described above.
```

## 6. D6 — 소진하지 않음 (카드가 지시한 정상 경로)

카드는 "릴리스 준비 때 이 카드를 열어 D7a만 소진하고 D6는 남기는 것이 정상 경로"라고
못박았다. 그대로 따랐다. 다음 카드가 처음부터 다시 찾지 않도록 좌표만 확정해 남긴다.

| 항목 | 실측 |
|---|---|
| 센티넬 위치(배포 템플릿) | `internal/template/templates/.claude/skills/moai/workflows/sync/delivery.md:33` |
| 센티넬 위치(로컬 미러) | `.claude/skills/moai/workflows/sync/delivery.md:33` |
| 각 파일 내 히트 수 | 1건씩 (`grep -c spec_git_workflow` → 1) |
| 쌍 관계 | **바이트 동일 쌍이 아니다.** `cmp` rc=1 (char 12891, line 275). SPEC v0.2.0 감사-D2가 이 쌍을 "중립화 미러(로컬 전용 내용 보존)"로 판정했으므로 차이는 정상이며, 결함으로 보고할 것이 아니다 |
| 제거 시한 | 본문에 `removed in v3.3.0` 명시 |
| 발화 조건 | 정본 키가 **부재**하고 사어 키가 **존재**할 때만. 배포 git-strategy.yaml이 정본 키를 실으므로 일반 설치에서는 발화하지 않는다 |

v3.3.0 카드를 세울 때의 범위 초안: 위 두 좌표의 폴백 절과 매핑 표를 제거하고, 제거 후에도
정본 키 부재가 "미매칭 값 = 중단"으로 올바르게 떨어지는지 회귀로 단언할 것. 대표 mutant —
배포 템플릿만 고치고 로컬 미러를 남겨 두 지시문이 갈리게 만드는 구현.

카드 발행은 운영자·리드 소관이므로 이 레인은 세우지 않았다.

# t472 — 카드 전제 재검 (plan 착수 전)

측정 트리: `.claude/worktrees/t472`, HEAD `7835148d3` (= `origin/develop`).
측정자: lane-3. 카드 본문의 리드 실측을 인용하지 않고 이 트리에서 다시 쟀다.

## 결론 먼저

카드의 진단 — **"이 리포의 설정 공란: git-strategy.yaml:7 `worktree_base_branch: ''`"** — 은
이 트리에서 **거짓**이다. 그러나 카드가 보고한 **증상**(착지 판정이 `origin/main` 을 묻는다)은
**참**이다. 원인이 다르고, 원인이 다르면 축 A 의 처방이 통째로 무효가 된다.

## 측정

### 대조군 — 커밋 메시지 카드 id 관례는 살아 있다

| 형태 | 명령 | 결과 |
|---|---|---|
| `(tNNN)` | `git log refs/remotes/origin/develop --format=%s \| grep -cE '\(t[0-9]+\)'` | 963 |
| `(card tNNN)` | 같은 파이프, `grep -cE '\(card t[0-9]+\)'` | 214 |

카드가 적은 952 와 다르다(963). 리드 측정 이후 커밋이 늘어난 것으로 보이며, 어느 쪽도
틀리지 않았다 — 이동하는 ref 를 잰 값이라 시점이 다르면 값이 다르다. 이 표의 값은
`7835148d3` 에 대한 것이다.

괄호 없는 맨 형태는 대조군에 넣지 않았다. `t461` 로 시험하면 맨 형태 0 건, 괄호 형태 11 건이
나와 맨 형태가 귀속을 잡지 못하는 사례가 실제로 있고, 반대로 참조를 귀속으로 오독하는
사례도 카드가 이미 둘 적어 두었다.

### 설정 값 — 통합 브랜치에는 이미 `develop` 이 들어 있다

    git show HEAD:.moai/config/sections/git-strategy.yaml | grep -n worktree_base_branch
    5:    worktree_base_branch: develop

워킹트리도 클린이다(`git status --short` 무출력). 값이 들어온 시점은 `a9c61cf56`
(2026-08-27, "add git_strategy.worktree_base_branch schema key and neutral default")이고,
그 커밋 이후 이 파일을 건드린 커밋은 없다.

### 그런데 판정은 여전히 `origin/main` 을 묻는다

이 트리에서 빌드한 바이너리로 잰 값이다(설치본이 아니다 — 아래 도구 출처 절 참조):

    moai todo done --help
      `--require-landed` refuses unless a commit on origin/main names the card.

    moai todo pr
      t237  landed   queued  ...

`landed` 가 실제로 나오므로 기제는 죽어 있지 않다. 다만 묻는 ref 가 `origin/main` 이다.

### 원인 — 읽는 트리가 다르다

`internal/cli/todo.go:81-90` 이 그 이유를 주석으로 이미 적어 두었다. 착지 ref 는
`resolveTodoQueueRoot()` 로 해석되고, 그것은 워크트리가 아니라 **primary 체크아웃**이다 —
"큐와 통합 브랜치는 하나의 저장소의 성질이지, 명령이 어느 워크트리에서 돌든의 성질이
아니다" 라는 근거로 그렇게 설계돼 있다.

그리고 primary 체크아웃은 `main` 에 있다:

    cat .git/HEAD            → ref: refs/heads/main
    cat .git/refs/heads/main → 7ad9f8534

primary 의 같은 파일은 `worktree_base_branch: ""` 다(7행). 이것은 스테일도 아니고
`moai update` 되돌림이라 단정할 수도 없다 — **값을 넣은 커밋이 `main` 계열에 없다**:

    git merge-base --is-ancestor a9c61cf56 refs/remotes/origin/main → rc=1
    git show refs/remotes/origin/main:.moai/config/sections/git-strategy.yaml
      | grep worktree_base_branch → 무출력 (키 자체가 없다)

대조군: 더 오래된 `11216d13f` 도 `origin/main` 조상이 아니다(rc=1). 즉 위 음성은 ref 를
잘못 짚어 나온 것이 아니라 `origin/main` 이 실제로 develop 계열보다 한참 뒤에 있어서 나온다.

**정리하면**: 키는 develop 브랜치 안에 있고, 그 값을 읽는 주체는 `main` 에 파킹된
체크아웃이다. 그래서 값은 릴리스로 develop 이 main 에 닿기 전까지 원리상 발효될 수 없다.

## 축 A 재판정 — "한 줄 설정" 이 아니다

카드는 축 A 를 "worktree_base_branch 를 develop 로 설정. 한 줄이고 기제 변경 0" 으로 적었다.
그 편집은 이 브랜치에서 **무동작**이다 — 값이 이미 `develop` 이다. 커밋해도 판정은 그대로
`origin/main` 을 묻는다. 읽는 트리가 `main` 이기 때문이다.

남는 선택지는 둘이고, 성격이 다르다.

1. **릴리스 경로** — develop 이 main 에 닿으면 저절로 발효된다. 할 일이 없는 대신, 언제
   고쳐지는지가 릴리스 일정에 묶이고 그때까지 증상이 계속된다.
2. **해석 경로** — 착지 ref 를 "primary 체크아웃이 체크아웃한 브랜치의 설정" 에서 읽는 것이
   옳은가를 다시 판정한다. 이것은 축 B(빈 값의 조용한 폴백)와 같은 뿌리이고, 한 줄 설정이
   아니라 설계 결정이다.

축 A 를 원문대로 두면 "한 줄이면 끝난다" 는 잘못된 기대를 만든다.

## 내가 틀렸던 것 하나 (기록)

`moai todo pr t461` 이 `queue is empty` 를 내는 것을 보고 저장소가 갈렸다고 판단했다.
`.moai/` 아래 `backlog.db` 가 네 벌 있고 그중 둘이 0바이트라 그럴듯했다. 틀렸다 —
`todo pr` 은 **라이브 큐**만 훑고 `t461` 은 이미 아카이브된 카드다. 인자 없이 돌리면
정상적으로 전 카드를 낸다. 보고 전에 실행으로 확인해 잡았다.

0바이트 사본 둘(`.moai/backlog.db`, `.moai/state/backlog.db`, 둘 다 09-02 07:32)은 그대로
남아 있다. 살아 있는 저장소는 `.moai/state/todo/backlog.db`(368KB, 09-03 15:19)다. 이 둘이
무해한 잔재인지 언젠가 읽히는 함정인지는 **확인하지 않았다** — 미검증 항목이다.

## 도구 출처 (VCI §2.2)

설치본 `~/go/bin/moai` 는 `v3.2.0-rc.0`, 커밋 `e79c010b8` 로 이 트리 HEAD 보다
**190 커밋 뒤**다. 그리고 그 사이 판정 대상 코드가 바뀌었다 — `internal/cli/todo.go`,
`internal/kanban/prlink_landed.go`, `internal/config` 를 건드린 커밋이 18 개 있으며 그중
`78795be32` 는 todo help 문안 자체를 고쳤다.

따라서 위 help 측정은 설치본이 아니라 **이 트리에서 빌드한 바이너리**로 다시 쟀다
(`go build -o <scratch>/moai-t472 ./cmd/moai`, rc=0). 두 빌드가 같은 답을 냈다는 사실은
별도로 적어 둔다 — 결론은 바뀌지 않았으나, 안 재고 넘어갔으면 근거가 없었다.

## 축 A 선행 실측 — 이 키를 읽는 소비자 4곳

카드가 [HARD] 로 요구한 "다른 소비자 영향 실측" 이다. `LoadWorktreeBaseBranch` 호출부를
전수로 셌다(테스트 제외).

| 소비자 | 위치 | 하는 일 |
|---|---|---|
| 카드 워크트리 base | `internal/cli/session_worktree.go:215` | 카드 워크트리를 어느 브랜치에서 팔지 결정 |
| doctor 점검 | `internal/cli/doctor_worktree_base.go:43` | 설정값과 `refs/remotes/origin/HEAD` 불일치를 보고 |
| 착지 ref | `internal/kanban/prlink_landed.go:75` | 이 카드가 다루는 축 |
| SessionStart 정렬 | `internal/hook/worktree_base_branch.go:156` | **`refs/remotes/origin/HEAD` 를 설정값에 맞춰 재정렬한다**(:130 의 notice 문안) |

명시적 비소비자도 하나 있다 — `internal/cli/mcp_review_material.go:116` 이 "이 키를 여기서
읽지 않는다" 를 주석으로 못박아 두었다.

네 번째가 중요하다. 이 키는 조회 전용이 아니라 **저장소의 공유 ref 를 쓰는 경로를 가진다**.
따라서 축 A 의 "기제 변경 0" 은 두 번째 이유로도 성립하지 않는다. 다만 지금 이 브랜치에서
값은 이미 `develop` 이므로, develop 기반 워크트리에서 도는 네 소비자는 이미 `develop` 을
본다. `""` 를 보는 것은 `main` 에 파킹된 primary 뿐이다.

## 미검증 / 잔여 위험

- primary 의 `worktree_base_branch: ""` 줄이 어디서 왔는지 **확정하지 않았다**.
  `origin/main` 에는 키가 아예 없으므로 커밋에서 온 값이 아니다. `moai update` 가
  `.moai/config` 를 통째로 재배포하며 템플릿 기본값을 쓴 결과라는 가설이 유력하나
  (CLAUDE.local.md §2.3 이 그 동작을 명시한다), 이 실행을 관측하지는 못했다.
- 축 A 가 건드리는 값의 **다른 소비자**를 아직 세지 않았다. 이름이 landed 전용이 아니므로
  워크트리 생성 base 등 다른 경로가 같은 키를 읽는지 확인이 필요하다 — 카드가 [HARD] 로
  요구한 선행 실측이며 아직 미이행이다.
- 축 C·D·E 는 이 재검에서 건드리지 않았다. 카드의 서술을 그대로 신뢰하지 말고 각각 따로
  재야 한다.
- 착지-그러나-picked 적체의 현재 규모를 이 트리에서 다시 세지 않았다. 카드의 18 장은
  리드 측정이며 그 뒤 리드가 수동 스윕을 돌려 수가 이미 바뀌었다.

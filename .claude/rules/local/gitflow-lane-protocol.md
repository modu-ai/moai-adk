---
description: "git-flow lane protocol (repo-local) — card worktrees branch from develop, lanes merge into a single develop integration worktree, origin/develop is the CI verdict surface, rc builds are cut from develop, release/vX.Y.Z is the only path to main"
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,.claude/rules/local/repo-local-pr-policy.md"
---

# git-flow Lane Protocol (moai-adk-go, local-only)

> **2026-08-27, 운영자 지시.** 이 리포는 GitHub Flow에서 git-flow로 전환했다. 2026-08-26 백로그 카드 t281이 정한 "로컬 전용·일회용 `develop`, 원격 push 금지"를 **의도적으로 뒤집는다**. 모델과 근거는 `CLAUDE.local.md` §4.1, 여기는 레인과 리드가 그대로 따르는 **운영 규칙**이다.
>
> 로컬 전용 파일이다(형제: `ci-watch-protocol.md`, `repo-local-pr-policy.md`, `lifecycle-sync-gate.md`). `internal/template/templates/`에 미러하지 않는다.

---

## 1. 분기점 — 카드 워크트리는 `develop`에서

[HARD] 카드 워크트리는 **`develop`에서** 만든다. `main`이 아니다.

[HARD] **로컬 `main`은 동기화만 하는 참조점이고, 아무도 거기서 분기하지 않는다.** 이 모델에서 `main`을 갱신하는 유일한 경로는 릴리스 PR이며, 로컬 `main`이 `origin/main`보다 뒤처져 있어도 작업에는 지장이 없다 — `develop`이 `origin/main`을 포함하기 때문이다. 따라서 로컬 `main`을 앞당기지 못하는 상황(예: 공유 체크아웃의 미커밋 작업과 충돌)은 작업을 막는 사유가 아니다. 상태줄의 `↓N` 표시는 그 사실의 반영일 뿐이다.

- 런처를 경유한다: `moai cc -w <card-id>` 또는 현재 세션에서 `EnterWorktree(<card-id>)`. **맨손 `git worktree add` 금지** — git은 아는데 MoAI는 모르는 트리가 생겨 `done`/`clean`/`recover`가 닫을 대상이 없어진다.
- 생성 직후 브랜치를 제자리에서 개명한다: `git branch -m WT-<slug>`. slug은 카드가 **하는 일**에서 뽑고(소문자 `a-z0-9-`, 토큰 3개 이하, 24자 이하), **카드 id를 넣지 않는다**. 워크트리 디렉터리는 카드 id를 유지한다(`.claude/worktrees/<card-id>`).
- 새 카드는 새 워크트리다. 이전 카드 트리에 앵커돼 있으면 `ExitWorktree`로 primary 체크아웃에 돌아온 뒤 만든다 — 안 그러면 새 카드 작업이 옛 카드 브랜치에 얹힌다.
- **추적성 운반체 3종은 그대로다**: dispatch의 `card:` 필드, 브랜치 위 **모든** 커밋 메시지 안의 카드 id, 증거 경로(`.moai/reports/<card-id>/…`). 브랜치 이름은 더 이상 카드를 식별하지 않으므로 셋 중 무엇도 생략하지 않는다.

## 2. 통합 면 — `develop` 워크트리는 하나뿐

[HARD] `develop`은 **정확히 한 워크트리**에만 체크아웃된다 — 리드가 배치 시작 시 provisioning 하는 통합 워크트리 `.claude/worktrees/develop`. 레인은 자기 트리에 `develop`을 체크아웃하지 않는다(git은 한 브랜치를 한 워크트리에만 내준다).

[HARD] **병합 절차의 정본은 여기가 아니다.** `WT-*` 브랜치의 통합 절차 — 병합 창 조율, 통합 워크트리 진입, no-ff 병합, 통합 브랜치 push(강제 push 금지), 퇴장 — 는 배포되는 `.claude/skills/moai/workflows/sync/delivery.md` Step 3.2 의 `git-flow` 전략 `WT-*` 경로가 소유한다. 레인은 그 절차를 그대로 따르고, 이 문서는 아래 리포 고유 사항만 덧붙인다. 절차를 여기에 다시 적지 않는다 — 두 벌이 되는 순간 갈라진다.

리포 고유 사항:

- 통합 워크트리 경로는 `.claude/worktrees/develop` 이고, 리드가 배치 시작 시 provisioning 한다.
- 통합 브랜치는 `develop` 이며 push 대상은 `origin/develop` 이다(§4의 CI 판정 면).
- 병합을 마치고 자기 카드 작업이 남아 있으면 `EnterWorktree(<card-id>)` 로 재진입한다 — `ExitWorktree` 는 primary 체크아웃으로 돌아가지 자기 트리로 돌아가지 않는다.

[HARD] 자기 워크트리 안에서 `git -C .claude/worktrees/develop merge …` 로 **원격 조작하는 것은 worktree-session 가드가 거부한다**. 들어가는 것이 유일한 인가 경로다.

## 3. 직렬화 — 병합 창은 한 번에 한 레인

[HARD] 기존 메커니즘을 그대로 쓴다. 새로 만들지 않는다.

> **[HARD] 창의 생존 판정은 세션 프로세스에 묶여 있다 — 카드 t298.** `acquire`가 남기는 pid는 그 명령을 실행한 짧은 CLI 프로세스가 아니라 **그것을 실행한 세션**의 것이다. 그래서 창은 `acquire`가 반환한 뒤에도 계속 held로 읽히고, 풀리는 길은 셋뿐이다 — 홀더 세션이 죽거나, 홀더가 스스로 `release` 하거나, 다른 레인이 기록을 남기는 `--force`로 가져가거나.
> 소유자를 판별하지 못한 채 잡힌 창은 pid 0으로 기록되고 **살아 있는 것으로** 읽힌다. 확실하지 않을 때 창을 비우는 쪽이 두 레인이 함께 머지하는 사고로 이어지므로, 판정은 늘 "살아 있다" 쪽으로 기운다.
> **수정 이전에 잡힌 창은 여전히 인수 가능하게 읽힌다** — 옛 기록에는 세션 앵커가 없다. 업그레이드 시점에 창을 쥐고 있던 레인은 `moai integration acquire`를 한 번 더 실행해 재획득한다.
> 이 기록이 레인을 기계적으로 갈라놓지는 않는다. `acquire` 자체가 읽고-고치고-쓰는 과정을 갈라 세우지 않으므로, 같은 순간에 두 레인이 잡으러 들어오면 둘 다 잡았다고 믿을 수 있다. 이것은 조율 신호이지 권한 경계가 아니며, **리드 공지가 여전히 첫 번째 층**이고 이 기록은 그 아래 기계 층이다.

```bash
moai integration acquire --name <lane>   # 통합 워크트리에 들어가기 전
moai integration status                  # 누가 쥐고 있는지
moai integration release                 # 완료 보고를 보낸 뒤
```

살아 있는 보유자의 창을 뺏으려면 `--force`가 필요하고, 무엇을 밀어냈는지 기록된다 — 의도적이어야 하며 조용해서는 안 된다.

[HARD] **`MERGE_HEAD`가 비어 있는 것은 필요조건일 뿐 결코 충분조건이 아니다.** `git rev-parse -q --verify MERGE_HEAD` 가 아무것도 찍지 않는 상태는 다른 레인이 해결 중일 때도 똑같이 나타난다 — `git merge --abort` 와 재시도 사이, 혹은 아직 아무것도 stage 하지 않은 시점. 이 침묵을 "트리가 비었다"로 읽는 것이 정확히 두 레인이 겹치는 경로다. 겹침은 한쪽이 커밋할 때까지 보이지 않는다. 프로브는 **마지막** 확인이지 첫 확인이 아니다.

[HARD] **병합 커밋 직전과 push 직전에 `HEAD`를 다시 읽는다.** 턴 앞에서 읽어둔 값도, 세션 시작 시 보고된 브랜치도 쓰지 않는다.

```bash
git rev-parse --short HEAD
git branch --show-current
```

가정과 다르면 진행하지 말고 발산을 보고한다.

## 4. Push — `develop`을 원격에 올린다

병합 후 레인이 직접 올린다: `git push origin develop`.

거부되면 다른 레인이 먼저 올린 것이다: `git fetch` → 가져온 `develop`을 통합 → 다시 push. **절대 force 하지 않는다.**

**원격 CI(`origin/develop`)가 통합 판정의 주체다.** 로컬 통과는 조기 신호일 뿐이다 — 깨끗한 환경도, darwin/windows 매트릭스도 아니다.

## 5. 충돌 — 변경을 소유한 레인의 몫

병합이 일으킨 충돌은 그 병합을 하는 레인이 해결한다. 해결할 수 없는 충돌(다른 레인이 이미 합친 변경과의 의미적 충돌)은 **리드에게 blocker 보고**다. 강제 병합이 아니다.

## 6. 병합 이후 — 레인은 카드를 스스로 고르지 않는다

- 병합·push를 마치면 `ExitWorktree`로 primary 체크아웃에 돌아와, **리드가 다음 카드를 dispatch 할 때까지 기다린다.** 레인이 큐에서 카드를 집지 않는다.
- [HARD] **카드 워크트리는 작업이 `origin/develop`에 올라간 뒤에야 폐기한다.** 그전까지 그 트리가 작업의 유일한 사본이다. L1 트리(`.claude/worktrees/…`)는 `moai worktree done`의 대상이 아니다 — 세션 종료 keep/remove 프롬프트나 `git worktree unlock` + `git worktree remove`로 닫는다.

## 7. 리드 — 읽어서 판정한다

- [HARD] 리드는 레인의 **답장이 아니라 증거를 읽고** 카드를 전진시킨다. 답장은 관측이 아니라 주장이고, 라우팅이 보장되지도 않는다.
- 병합 판정(무엇이 `develop`에 들어갔는가)은 리드의 것이다. 작업을 만든 레인에게 자기 결과를 판정하게 하지 않는다.
- 증거 파일이 없거나 읽히지 않거나 낡았으면 **gap**이다 — 카드는 그대로 두고 이유를 보고한다.
- 병합이 확인되면 다음 카드를 **지금 비어 있는 레인**에 dispatch 한다.

## 8. 검증은 레인-로컬

[HARD] 자기 변경이 영향 줄 수 있는 테스트만 돌리고, push 후 `origin/develop` CI가 전체 스위트를 돌리게 한다.

- **`go test ./...` 를 로컬에서 돌리지 않는다.** 레인 여럿이 동시에 돌려 load 413까지 치솟고 머신을 마비시킨 사고가 있다(2026-08-15).
- **백그라운드 부하를 만들지 않는다.** 경합이 필요한 검증이라면 부하는 정리 보장이 있어야 한다 — 테스트 프레임워크 cleanup 훅에 등록된 kill이거나, 밖에서 프로세스를 묶는 `timeout` 래퍼. 뒤에 붙인 `kill`은 정리가 아니다(도달하지 못하는 줄이다).

## 9. rc 빌드 — 운영자 요청 시, 통합 워크트리에서

```bash
# .claude/worktrees/develop 안에서
make build VERSION=vX.Y.Z-rc.N
rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai
~/go/bin/moai version; echo $?          # exit 0 이어야 한다
```

- [HARD] `rm -f` 를 생략한 맨 `cp` 덮어쓰기는 다음 호출에서 **exit 137(SIGKILL)** 을 낸 전례가 있다 — inode를 갈아끼우는 clean 재설치여야 한다. (`make install` 도 동등하다.)
- [HARD] **맨손 `go install ./cmd/moai` 금지.** Makefile의 `LDFLAGS`를 안 실어서 버전/커밋/날짜가 컴파일 기본값으로 박히고, 그러면 binary lag 검증 자체가 불가능해진다.
- 137이 나오면 `rm -f` + `cp` 를 다시 한다.

## 10. 릴리스 — 레인의 범위 밖

로컬 rc 시험을 통과하면 `release/vX.Y.Z`를 **`develop`에서** 분기하고, 이후는 `release` 하네스가 맡는다. `main`은 그 릴리스 PR로만 갱신된다(브랜치 보호 `enforce_admins: true`). 레인은 릴리스 브랜치를 만들지도, main PR을 내지도 않는다.

---

## Cross-references

- `.claude/skills/moai/workflows/sync/delivery.md` Step 3.2 — `WT-*` 통합 절차의 정본(배포되는 스킬이 소유; 이 문서는 리포 고유 사항만)
- `CLAUDE.local.md` §4.1 — 이 모델의 근거와 전환 기록(t281 역전, 2026-08-14 기각 사유의 현재 상태)
- `.claude/rules/local/repo-local-pr-policy.md` — `main` 브랜치 보호와 PR 필수 규정
- `.moai/docs/git-workflow-doctrine.md` (§18) / `.moai/docs/git-local-workflow-doctrine.md` (§23) — 상위 독트린(일부 초과분(superseded), 상단 주석 참조)
- `.moai/config/sections/git-strategy.yaml` — `git_strategy.manual.workflow: git-flow` + develop/release 키(`moai update` 후 재적용 필요, `CLAUDE.local.md` §2.3)

---

Classification: Local-only operational rule (dev-only, never mirrored to the distributed template).

---
description: "git-flow lane protocol (repo-local) — card worktrees branch from local main, lanes land through the integration verb on local main, origin/main is the CI verdict surface, batch releases reach origin/main only through a release PR"
paths: ".moai/specs/**,.claude/skills/moai/workflows/run.md,.claude/skills/moai/workflows/sync.md,.claude/rules/local/repo-local-pr-policy.md,AGENTS.local.md,internal/cli/**,internal/hook/**"
---

# git-flow Lane Protocol (moai-adk-go, local-only)

> **2026-08-27, 운영자 지시.** 이 리포는 GitHub Flow에서 git-flow로 전환했다. 2026-08-26 백로그 카드 t281이 정한 "로컬 전용·일회용 `develop`, 원격 push 금지"를 **의도적으로 뒤집는다**. 모델과 근거는 `AGENTS.local.md` §4.1, 여기는 레인과 팩토리 리더가 그대로 따르는 **운영 규칙**이다.
>
> 로컬 전용 파일이다(형제: `ci-watch-protocol.md`, `repo-local-pr-policy.md`, `lifecycle-sync-gate.md`). `internal/template/templates/`에 미러하지 않는다.

---

## 1. 분기점 — 카드 워크트리는 로컬 `main`에서

[HARD] 카드 워크트리는 **로컬 `main`에서** 만든다. `origin/main`이 아니다(§4.0 §1).

[HARD] **로컬 `main`은 카드 기저이자 착지 면이다** (§4.0 §1·§2). `origin/main`은 릴리스 배치로만 전진하므로 로컬 `main`이 앞서 있는 것이 정상이다. 상태줄의 `↓N`은 배치가 `origin/main`에 올라갔는데 재동기화가 아직 없다는 뜻이며, 그때는 §4.0 §4의 재동기화 대상이다.

[HARD] **로컬 `main`에서 커밋하는 것은 설정이 막는다 (SPEC-MAIN-COMMIT-BAN-001, 카드 t1337).** `workflow.branch_guard.deny_commits_on: [main]`이 primary 체크아웃의 `git commit` / `git revert` / `git cherry-pick`을 거부하며, SPEC-LOCAL-MAIN-FLOW-001 Q1 결정(유지)으로 그 값은 그대로다. 카드 커밋은 로컬 `main`에서 분기한 카드 워크트리에서만 만들고, 병합만 §4.0 §2의 도구 경로가 한다. 잔여물 처분 절차는 `.moai/docs/gitflow-integration-chain.md`의 역사 기록이다.

- 하네스에 맞는 경로로 들어간다:
  - Claude Code 레인: `moai cc -w <card-id>` 또는 현재 세션의 `EnterWorktree(<card-id>)`.
  - Codex 레인: 기존 트리에 새 세션으로 들어갈 때 `moai codex -w <card-id>`.
  - Codex의 `-l` 레인: 감독 런처가 카드 워크트리를 고르고 `codex -C <워크트리 절대경로>`로 자식 세션을 시작한다. 자식 세션은 해당 트리의 `AGENTS.local.md`를 읽고 작업하며 `moai cc -w`를 호출하지 않는다. `moai codex` 런처로 시작한 세션은 이 파일을 `developer_instructions`에 싣는다.
  - 트리가 없으면 `moai worktree new <card-id>`로 먼저 만든다. **맨손 `git worktree add` 금지** — git은 아는데 MoAI는 모르는 트리가 생겨 `done`/`clean`/`recover`가 닫을 대상이 없어진다.
- 생성 직후 카드 트리의 `HEAD`와 로컬 `main`의 커밋이 같은지 확인한다. 다르면 작업을 시작하지 않고 분기 기준을 바로잡는다.
- 생성 직후 브랜치를 제자리에서 개명한다: `git branch -m WT-<slug>`. slug은 카드가 **하는 일**에서 뽑고(소문자 `a-z0-9-`, 토큰 3개 이하, 24자 이하), **카드 id를 넣지 않는다**. 워크트리 디렉터리는 카드 id를 유지한다(`.claude/worktrees/<card-id>`).
- 새 카드는 새 워크트리다. 이전 카드 트리에 앵커돼 있으면 `ExitWorktree`로 primary 체크아웃에 돌아온 뒤 만든다 — 안 그러면 새 카드 작업이 옛 카드 브랜치에 얹힌다.
- **추적성 운반체 3종은 그대로다**: dispatch의 `card:` 필드, 브랜치 위 **모든** 커밋 메시지 안의 카드 id, 증거 경로(`.moai/reports/<card-id>/…`). 브랜치 이름은 더 이상 카드를 식별하지 않으므로 셋 중 무엇도 생략하지 않는다.
- [HARD] 레인 세션은 카드 워크트리 안에서 시작해 그 안에 머문다 — 세션 도중 카드 워크트리에서 다른 카드 워크트리로의 이동은 금지며, 불가피하게 이동할 때는 이동 뒤 `/clear` 를 정확히 1회 실행한다. 이 금지의 정본은 배포 규칙 `.claude/rules/moai/workflow/factory-dispatch.md`의 Isolation 절이며, 여기에 다시 적지 않는다. primary 체크아웃 복귀(`ExitWorktree`, §6)는 이 금지의 적용 밖이다.

## 2. 통합 면 — primary 체크아웃(`main`)

[HARD] 착지 면은 **primary 체크아웃**이다(HEAD가 `main`을 가리킬 때, §4.0 §2). `main`을 잡은 별도 통합 워크트리가 있으면 그것이 면이다. 레인은 자기 카드 트리에 `main`을 체크아웃하지 않는다(git은 한 브랜치를 한 워크트리에만 내준다).

[HARD] **병합 절차의 정본은 여기가 아니다.** 로컬 `main`으로의 병합 정본은 `AGENTS.local.md` §4.0 §2(도구 경로)다. 배포 스킬 `.claude/skills/moai/workflows/sync/delivery.md` Step 3.2 의 `WT-*` 경로는 그 정본과 충돌하는 범위에서 이 리포에 적용하지 않는다. 절차를 여기에 다시 적지 않는다 — 두 벌이 되는 순간 갈라진다.

[HARD] **리포 로컬 예외 — delivery.md Step 3.2 제6단계(통합 브랜치 push)는 이 리포에서 EXCLUDED다(2026-09-02 운영자 지시).** 배포 스킬의 WT-* 절차가 push 단계를 안내하더라도 이 리포의 레인은 그 단계를 수행하지 않는다 — 배치 릴리스는 리더의 일괄 행위다(§4, §7). 배포 스킬 파일은 건드리지 않으며, 레인에게는 이 리포의 로컬 답이 우선한다.

리포 고유 사항:

- `main`을 잡은 별도 통합 워크트리가 없을 때 면은 primary다. provisioning 절차는 없다 — develop 시절의 `.claude/worktrees/develop`과 `moai cc -w develop --branch develop`(2026-09-02)은 역사다.
- 통합 브랜치는 `main`이다. 배치는 `release/main-batch-YYYYMMDD`를 거쳐 `origin/main`에 착지하며(§4.0 §6), 그 CI가 판정 면이다(§4).
- 병합을 마치고 자기 카드 작업이 남아 있으면 `EnterWorktree(<card-id>)` 로 재진입한다 — `ExitWorktree` 는 primary 체크아웃으로 돌아가지 자기 트리로 돌아가지 않는다.

[HARD] 자기 워크트리 안에서 `git -C <primary 경로> merge …` 로 **병합하는 것은 worktree-session 가드가 거부한다**. 로컬 `main`으로의 병합은 도구 경로(`moai integration merge --card`)로만 한다(§4.0 §2).

## 3. 직렬화 — 병합 창은 한 번에 한 레인

[HARD] 기존 메커니즘을 그대로 쓴다. 새로 만들지 않는다.

> **[HARD] 창의 생존 판정은 세션 프로세스에 묶여 있다 — 카드 t298.** `acquire`가 남기는 pid는 그 명령을 실행한 짧은 CLI 프로세스가 아니라 **그것을 실행한 세션**의 것이다. 그래서 창은 `acquire`가 반환한 뒤에도 계속 held로 읽히고, 풀리는 길은 셋뿐이다 — 홀더 세션이 죽거나, 홀더가 스스로 `release` 하거나, 다른 레인이 기록을 남기는 `--force`로 가져가거나.
> 소유자를 판별하지 못한 채 잡힌 창은 pid 0으로 기록되고 **살아 있는 것으로** 읽힌다. 확실하지 않을 때 창을 비우는 쪽이 두 레인이 함께 머지하는 사고로 이어지므로, 판정은 늘 "살아 있다" 쪽으로 기운다.
> **수정 이전에 잡힌 창은 여전히 인수 가능하게 읽힌다** — 옛 기록에는 세션 앵커가 없다. 업그레이드 시점에 창을 쥐고 있던 레인은 `moai integration acquire --name <lane> --card <card-id>`를 한 번 더 실행해 재획득한다.
> 이 기록이 레인을 기계적으로 갈라놓지는 않는다. `acquire` 자체가 읽고-고치고-쓰는 과정을 갈라 세우지 않으므로, 같은 순간에 두 레인이 잡으러 들어오면 둘 다 잡았다고 믿을 수 있다. 이것은 조율 신호이지 권한 경계가 아니며, 이 기록이 기계 층이다. open 정책에서 대기열 맨 앞으로 승격되어 창을 쥔 레인은 리더 지명 없이 병합한다.

```bash
moai integration acquire --name <lane> --card <card-id>   # 통합 워크트리에 들어가기 전
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

## 4. Push — 리더 일괄, 레인은 하지 않는다

**레인은 `main`을 push하지 않는다.** 이 규칙은 2026-09-02 운영자 지시(당시 대상은 develop)의 현행 형태다. 레인의 공개 소관은 로컬 병합에서 끝난다 — 병합 커밋 SHA(카드 id 포함)를 완료 보고로 리더에게 전달하는 것까지다. 배치 릴리스는 리더가 **일괄**로 수행한다: 완료 보고에서 병합 SHA를 모아 배치를 닫고 §4.0 §6의 절차로 `release/main-batch-YYYYMMDD`에 push한다.

**리더의 원격 착지 검증.** 배치 PR이 병합된 뒤 리더는 `git fetch origin main` + `git rev-parse origin/main`으로 원격이 움직였는지 확인하고, 그 뒤에야 워크트리 폐기를 승인한다. 카드 done은 배치를 기다리지 않는다(§4.0 §5; 구현은 t1621).

**원격 CI(`origin/main`)가 통합 판정의 주체다.** 로컬 통과는 조기 신호일 뿐이다 — 깨끗한 환경도, darwin/windows 매트릭스도 아니다.

**배치 닫기 트리거 — 임계값은 설정이 운반한다(SPEC-LEAD-AUTOPUSH-001).** 배치를 닫을 시점은 리더의 재량이 아니라 계수 트리거가 정한다: `git rev-list --count origin/main..main`이 `git_strategy.manual.lead_push_threshold`에 닿으면 배치를 닫고 §4.0 §6의 절차를 시작한다. 값의 원천은 `.moai/config/sections/git-strategy.yaml`이며 이 문서는 그 수를 다시 쓰지 않는다 — 키와 계수 명령만 명명한다. 트리거는 병합 창이 닫힌 뒤에만 평가한다 — 창(`moai integration acquire`/`release`) 안에서 push하지 않는다.

**초록 조건부 — 배치는 green일 때만 다시 연다.** 두 층이 있다. (i) 카드별 사전 게이트: 통합 창의 병합 트리 재측정을 통과하지 못한 병합은 배치에 넣지 않는다(지난 3건의 적색은 전부 이 재측정이 push 전에 잡았다). (ii) 배치 단위 보류: 마지막으로 착지한 배치의 `origin/main` CI가 red면 다음 배치를 보류한다 — 수리되어 green으로 돌아온 뒤에야 트리거가 다시 유효하다. 초록 조건부가 계수 트리거보다 우선한다. 그 CI 판정이 아직 없으면(첫 배치 등) red가 아니므로 보류 사유가 아니다.

**임계값 0(또는 키 부재)은 오류가 아니라 비활성이다.** `lead_push_threshold`가 0이면 트리거는 꺼진 것으로, 배치 닫기 시점은 리더의 판단으로 돌아간다 — 오류 조건으로 취급하지 않는다.

## 5. 충돌 — 변경을 소유한 레인의 몫

병합이 일으킨 충돌은 그 병합을 하는 레인이 해결한다. 해결할 수 없는 충돌(다른 레인이 이미 합친 변경과의 의미적 충돌)은 **리더에게 blocker 보고**다. 강제 병합이 아니다.

## 6. 병합 이후 — 레인은 카드를 스스로 고르지 않는다

- 로컬 병합을 마치고 병합 SHA를 리더에게 보고하면 `ExitWorktree`로 primary 체크아웃에 돌아와, **리더가 다음 카드를 dispatch 할 때까지 기다린다.** 레인이 큐에서 카드를 집지 않는다. 그 대기는 열린 대기가 아니라 **명시적 대기(explicit wait)** 다 — 디스크에 대기 기록(reason + 대상 + 재확인 지점)을 남기는 대기이고, 각성(awaken) 턴에서는 카드 작업 재개 전에 레인 스톨 워치독을 먼저 실행한다 — Skill("moai-lane-watchdog")을 호출해 따른다. 워치독은 진행 3채널(HEAD SHA · 증거 mtime · 통합 창 상태)을 재측정해 원인을 분류하고, 판단이 필요하면 결정 사다리(디스크 증거 → 결정 보드 → 감사 교차 → jev_ask → 리드 채팅 최후)로 답장 없이 푼다 — 답장 부재를 진행 부재로 읽지 않는다. 이 대기·사다리 독트린의 정본은 `.claude/rules/moai/workflow/auto-semantics.md`이고, 워치독의 queue-readonly 경계는 큐 수준 픽업 필터를 소유하는 SPEC-RELATION-PICKUP-FILTER-001과의 조합으로 성립한다(레인은 여전히 카드를 고르지 않는다 — 위 금지와 같은 경계다).
- **self-dispatch lane 예외 — 카드 임대.** self-dispatch 팩토리 run의 레인은 `moai factory next`로 대기 중인 다음 카드를 임대할 수 있다(레인이 수행하는 유일한 promotion). 이 예외를 제외한 큐 변경(`add`, `drop`, `done`, `edit` 등)과 `moai contract sign`은 레인에게 금지된다.
- **self-dispatch lane 예외 — 병합 창.** Claude self-dispatch 레인은 리더에게 창을 요청하지 않고 `moai integration merge --card`를 부르는 `moai factory complete`의 통합 절차로 스스로 통합 창을 잡고 자기 카드를 로컬 `main`에 병합한다(위 첫 번째 항목의 「리더에게 병합을 요청한다」를 이 레인에서 대체한다). 레인은 main에 손으로 git merge 하지 않고 moai integration merge --card 또는 그것을 부르는 moai factory complete 로만 병합한다. Codex 레인은 예외가 아니다 — merge-ready에서 정지한다(REQ-SD-025). 두 예외 모두 위 금지(그 외 큐 변경 + `moai contract sign`)를 바꾸지 않는다.
- [HARD] **카드 워크트리는 그 병합을 실은 배치가 `origin/main`에 착지한 뒤에야 폐기한다(§4.0 §5).** 그전까지 그 트리가 작업의 유일한 사본이다. 원격 착지는 리더의 배치 릴리스가 만든다(§4, §7). L1 트리(`.claude/worktrees/…`)는 `moai worktree done`의 대상이 아니다 — 세션 종료 keep/remove 프롬프트나 `git worktree unlock` + `git worktree remove`로 닫는다.

## 7. 리더 — 읽어서 판정한다

- [HARD] 리더는 레인의 **답장이 아니라 증거를 읽고** 카드를 전진시킨다. 답장은 관측이 아니라 주장이고, 라우팅이 보장되지도 않는다.
- 병합 판정(무엇이 로컬 `main`에 들어갔는가)은 리더의 것이다. 작업을 만든 레인에게 자기 결과를 판정하게 하지 않는다.
- 증거 파일이 없거나 읽히지 않거나 낡았으면 **gap**이다 — 카드는 그대로 두고 이유를 보고한다.
- 병합이 확인되면 다음 카드를 **지금 비어 있는 레인**에 dispatch 한다.
- **배치 릴리스는 리더의 일괄 소관이다.** 레인 완료 보고에서 카드 id와 로컬 병합 SHA를 모은다 → 배치를 닫을 시점은 §4의 임계 트리거와 초록 조건부가 정한다(리더 재량 단독이 아니다) → §4.0 §6의 절차로 `release/main-batch-YYYYMMDD`에 push하고 PR을 병합한다 → `git fetch`와 `git rev-parse origin/main`으로 원격 착지를 검증한다 → 그 뒤에야 워크트리 폐기를 승인한다.

## 8. 검증은 레인-로컬

**[SUPERSEDED by §4.0]** [HARD] 자기 변경이 영향 줄 수 있는 테스트만 돌리고, push 후 `origin/develop` CI가 전체 스위트를 돌리게 한다.

- **`go test ./...` 를 로컬에서 돌리지 않는다.** 레인 여럿이 동시에 돌려 load 413까지 치솟고 머신을 마비시킨 사고가 있다(2026-08-15).
- **백그라운드 부하를 만들지 않는다.** 경합이 필요한 검증이라면 부하는 정리 보장이 있어야 한다 — 테스트 프레임워크 cleanup 훅에 등록된 kill이거나, 밖에서 프로세스를 묶는 `timeout` 래퍼. 뒤에 붙인 `kill`은 정리가 아니다(도달하지 못하는 줄이다).
- **무거운 실행이 겹칠 자리에서는 자원 임대를 먼저 잡는다.** `moai slot acquire --resource <이름> --max-duration <상한>` → 실행 → `moai slot release --resource <이름>`. 보유자는 `moai slot status`로 읽는다. 통합 창(`moai integration`)과 기록·락·설정 키가 **분리돼 있다** — 병합 대기와 무거운 실행 대기가 서로를 막지 않게 하려는 것이다. 상한은 보유자 자신의 선언이라, 해제를 잊어도 상한까지만 묶인다(그 뒤에는 다른 레인이 `--force` 없이 인수한다). 매칭 명령을 거부하는 PreToolUse 가드는 **선택형이고 기본 꺼짐**(`workflow.slot_lease.enabled`)이며, 세 동작은 그 값과 무관하게 돈다. 표면 전체는 `.claude/rules/moai/workflow/resource-slot-lease.md`.
  - 거부는 숫자로 온다: 살아 있는 보유자가 상한 안에 있는 동안 다른 세션의 획득은 **exit 3**으로 침묵 거부되고 아무도 대체되지 않는다(2026-09-14 실측, `t774-repro-demo`) — 리더에게 묻는 대신 `moai slot status --resource <이름>` 한 번으로 지금 순번의 주인을 확인한다.
  - "무거운 실행"의 예 — 최소한 `internal/cli` 전체 스위트(2026-09-10 침입: 세 레인이 동시에 돌려 load 8~21에서 판정이 뒤집혔다 — 이것이 표면이 없을 때 침범이 실제로 난다는 통제군 기록이며, 동시 스위트 재실행으로 재시연하지 않는다), 그리고 `internal/factory`·`internal/hook` 같은 분 단위 패키지 스위트.
  - 이 규율이 리더의 기억이 아니라 디스크 위에 있어야 하는 이유가 바로 그 두 차례의 침입이다: 1차는 통지 누락, 2차는 통지가 도달해도 레인이 순번을 조회할 방법이 없었다 — 선언과 통지와 조회 가능한 절차는 각각 다른 일이다.
- **[SUPERSEDED by §4.0]** **[HARD] 「이 카드가 무엇을 바꿨는가」는 흡수한 ref 와의 merge-base 부터 잰다 — 리터럴 base SHA 로 재지 않는다.** "Go 변경 없음", "템플릿 변경 없음", "이 경로만" 같은 범위 판정식의 왼쪽 끝은 읽는 시점에 `CARD_BASE=$(git merge-base develop HEAD)` 로 다시 구하고, 값을 핀하지 않는다. 대조군은 `git diff --name-only "$CARD_BASE"..HEAD | wc -l`(1 이상이어야 함), 프로브는 같은 범위에 pathspec 을 붙인 형태다. 대조군이 0 이면 "변경 없음"이 아니라 "측정 불가"로 보고한다.
  - **[SUPERSEDED by §4.0]** 이유: 흡수하는 순간 리터럴 핀 범위에 다른 카드의 커밋이 들어온다. 로컬 develop 이 원격보다 앞서 있으면 `origin/develop` 기준 merge-base 도 흡수 전 분기점에 머물러 같은 오탐을 낸다. 실측(2026-09-10, `.moai/reports/t543/verdict.md`): 로컬 develop 을 흡수한 뒤 리터럴 핀과 `origin/develop` 기준은 모두 Go 51개를 냈고, `develop` 기준만 카드 자기 기여(파일 4, Go 0)를 냈다. 이 재현이 이 규율의 대조군이다.
  - **[SUPERSEDED by §4.0]** 원칙은 "흡수한 바로 그 ref"다. 이 저장소 절차의 흡수 대상은 로컬 `develop`(§11, `AGENTS.local.md` §4.1)이라 기본값이 `develop` 이다. 원격 develop 을 흡수하는 절차라면 ref 는 `origin/develop` 이 된다. develop 이 흡수 뒤 더 앞서가도 merge-base 는 마지막으로 흡수한 develop 커밋에 머물러 계속 옳다.
  - **[SUPERSEDED by §4.0]** 한계 — **병합 뒤에는 쓸 수 없다.** 카드가 develop 에 병합되면 merge-base 가 카드 tip 자신이 되어 범위가 비고, 판정식은 공허하게 통과한다(실측: 이미 병합된 카드 브랜치에서 빈 출력, `.moai/reports/t543/repro/limit-a-post-merge-all.txt`). 범위 판정식은 병합 전 평가 전용이다. 병합 뒤 근거는 병합 트리와 카드 브랜치 트리의 동일성으로 대신한다.
- **[HARD] 「이건 누구 것인가」는 커밋 열거로만 답한다 — 트리 비교로 답하지 않는다.** 정본형은 하나다: `git log <내 HEAD>..<상대 ref> -- <경로>`. 그리고 **두 번째 피연산자는 양성 대조 경로**다 — 같은 형태를 **남의 커밋이 확실히 있는 경로**에 한 번 더 걸어 발화를 본다. 그 두 번째 실행이 비면 첫 번째의 0행은 부재가 아니라 **미측정**이다. 대조를 따로 기억해야 할 규율로 두지 않는 이유는, 그렇게 둬 봤더니 물지 않았기 때문이다(2026-09-20 같은 날 세 카드가 전부 대조를 언급하고도 전부 틀렸다 — `t950` 은 9회 언급).
  - `git diff` 는 **어떤 철자든 트리 비교**라 이 질문을 답하지 않는다. 실측: 같은 피연산자 쌍에서 `git diff A..B` 와 `git diff A B` 는 **출력이 바이트 동일**하고, 같은 `A..B` 가 `git log` 에서는 반대 답을 낸다(log 0행 / diff 1행). 세 점 `git diff A...B` 만 `log A..B` 와 일치한다.
  - 왼쪽 끝은 **`HEAD`** 다 — 「내가 이미 가진 전부」라는 뜻이다. merge-base 나 핀한 SHA 를 놓으면 **내 앞 카드의 커밋이 상대 쪽 접촉으로 돌아온다**(실측 2건, 둘 다 `git merge-base --is-ancestor <커밋> <내 tip>` exit 0 = 내 것).
  - 바로 위 규율과 **묻는 질문이 다르다**: 저쪽은 「이 카드가 무엇을 바꿨는가」(자기 기여), 이쪽은 「이건 누구 것인가」(귀속). 병합 전 평가 전용이라는 경계는 같다.
  - 측정 출처: `.moai/reports/t1014/verdict.md` (재현 13건 동반).

## 9. rc 빌드 — 운영자 요청 시, 로컬 main 트리에서

```bash
# 로컬 main의 HEAD를 가진 트리(primary 체크아웃) 안에서
# macOS의 /usr/bin 개발 도구가 Xcode 라이선스 exit 69이면 설치된 CLT를 우선한다.
if [ -x /Library/Developer/CommandLineTools/usr/bin/make ]; then
  PATH="/Library/Developer/CommandLineTools/usr/bin:$PATH"
  export PATH
fi
make build VERSION=vX.Y.Z-rc.N
git status --short                      # 빌드 뒤 추적 파일 변경이 없어야 한다 (AGENTS.local.md §4.0 §3)
rm -f ~/go/bin/moai && cp bin/moai ~/go/bin/moai
sh scripts/verify-local-install.sh       # byte 동일성 + 설치본 version SHA/exit 0
```

- [HARD] `rm -f` 를 생략한 맨 `cp` 덮어쓰기는 다음 호출에서 **exit 137(SIGKILL)** 을 낸 전례가 있다 — inode를 갈아끼우는 clean 재설치여야 한다. (`make install` 도 동등하다.)
- [HARD] **맨손 `go install ./cmd/moai` 금지.** Makefile의 `LDFLAGS`를 안 실어서 버전/커밋/날짜가 컴파일 기본값으로 박히고, 그러면 binary lag 검증 자체가 불가능해진다.
- [HARD] 설치본 검증에 macOS `strings`를 쓰지 않는다. Xcode 라이선스 동의 여부가 제품 binary 판정을 중단시키기 때문이다. `scripts/verify-local-install.sh`는 `cmp`로 `bin/moai`와 설치본의 byte 동일성을 확인하고, 설치본의 `version`을 직접 실행해 측정 시점 HEAD의 short SHA와 exit 0을 함께 관측한다. 기본 `git`이 불가하면 macOS Command Line Tools의 `git`을 자동으로 재시도하며, 둘 다 불가하면 세 번째 인자로 기대 short SHA를 받아야 한다. `make verify-local-install`은 같은 script를 호출하는 편의 alias다.
- [HARD] macOS에서 `/usr/bin/make` 또는 `/usr/bin/git`이 Xcode 라이선스 exit 69를 내면 위 전처리로 설치된 Command Line Tools 경로를 PATH 선두에 둔다. `sudo xcodebuild -license`를 자동 실행하거나 검증 실패를 `|| true`로 숨기지 않는다.
- 137이 나오면 `rm -f` + `cp` 를 다시 한다.

## 10. 릴리스 — 레인의 범위 밖

**[SUPERSEDED — pending OQ-7]** 로컬 rc 시험을 통과하면 `release/vX.Y.Z`를 **`develop`에서** 분기하고, 이후는 `release` 하네스가 맡는다. `origin/main`은 릴리스 PR로만 갱신된다 — 배치 릴리스(`release/main-batch-YYYYMMDD`)도 포함한다(§4.0 §6; 브랜치 보호 `enforce_admins: true`). 레인은 릴리스 브랜치를 만들지도, main PR을 내지도 않는다.

## 11. main 갱신 — 배치 착지 후 로컬 main 재동기화 (SPEC-RC-TESTBED-001)

릴리스 배치가 `origin/main`에 착지하면 로컬 `main`은 뒤처질 수 있다(§4.0 §4). 언제 갱신할지의 **판정 기준**과 갱신하는 **경로**를 이 절이 정한다. rc.N 번호 정책(몇 번째 후보인가)은 `.moai/docs/version-management.md`의 Local RC Numbering 절이 소유한다.

[HARD] **판정 기준은 `origin/main`과의 ref 비교이고, 병합 모양을 전제하지 않는다.**

```bash
git fetch origin main
git rev-list --count --left-right origin/main...main
# "0 0" = 같다. "0 M" = 로컬에만 있는 착지분(배치 전) → 갱신 대상 아님
# "N 0"(N≥1) = origin/main 쪽에만 있는 커밋 → fast-forward 대상(§4.0 §4)
# "N M"(둘 다 ≥1) = 갈라짐 → 도구가 거부한다(운영자 결정 Q2 (i))
```

- `git branch --merged`는 **충분한 기준이 아니다.** `--merged`는 브랜치 헤드의 도달성만 보고, squash 병합으로 들어온 변경은 커밋 그래프에 헤드를 남기지 않아 못 본다 — reachability blindness의 사례가 `SPEC-WORKTREE-SQUASH-MERGE-001` 이다. 이 리포의 로컬 main 착지는 `--no-ff`(§4.0 §2의 도구 경로)로 규정되어 있지만, 그래서 *"--merged가 비어 있다"*를 이 레인의 확립된 사실로 단언하지도 않는다 — 판정식이 병합 모양과 무관해야 하므로, 위 rev-list 비교를 쓴다.
- 기준은 언제나 `origin/main`이다. 갱신 대상인 로컬 `main`도, 카드 워크트리의 브랜치 상태도 기준이 아니다.

[HARD] **갱신 경로는 도구 경로 하나뿐이다 — 통합 창 안의 fast-forward 전용 병합(§4.0 §4).**

- 운영자 터미널에서 직접 하지 않는다. 도구가 통합 창을 잡고 HEAD가 `main`을 가리키며 primary가 깨끗한지 확인한 뒤 fast-forward 전용으로 수행한다(§4.0 §4). 명령 이름은 실행 단계에서 정해지므로 여기 적지 않는다.
- primary 체크아웃에서 `git branch` / `git checkout` 으로 갱신하지 않는다. BranchGuard(`workflow.yaml` `branch_guard.enabled`, 로컬 opt-in)가 해당 패턴을 차단하며, 면제 경로 두 개(`claude --agent` 직접 실행, `MOAI_BRANCH_GUARD_EXEMPT=1`)는 **도구로 spawn된 서브에이전트에서 도달 불가능하다.** 운영자 터미널 sentinel 경로를 서브에이전트에 권장하지 않는다 — 도달 불가능한 면제에 기대는 절차는 의존 가능한 절차가 아니다.
- 자기 워크트리 안에서 `git -C <primary 경로> …` 로 조작하는 것은 §2와 같은 이유로 worktree-session 가드가 거부한다. **도구 경로가 유일한 인가 경로다.**
- 갱신 창 조율은 §3의 메커니즘을 그대로 쓴다(병합 창과 갱신이 겹치지 않게 한다). 갱신 후의 rc 빌드·clean 재설치는 §9 런북이 소유한다 — 여기에 다시 적지 않는다.

---

## Cross-references

- `.claude/skills/moai/workflows/sync/delivery.md` Step 3.2 — `WT-*` 통합 절차의 정본(배포되는 스킬이 소유; 이 문서는 리포 고유 사항만)
- `AGENTS.local.md` §4.1 — 이 모델의 근거와 전환 기록(t281 역전, 2026-08-14 기각 사유의 현재 상태)
- `.claude/rules/local/repo-local-pr-policy.md` — `main` 브랜치 보호와 PR 필수 규정
- `.moai/docs/git-workflow-doctrine.md` (§18) / `.moai/docs/git-local-workflow-doctrine.md` (§23) — 상위 독트린(일부 초과분(superseded), 상단 주석 참조)
- **[SUPERSEDED by §4.0]** `.moai/config/sections/git-strategy.yaml` — `git_strategy.manual.workflow: git-flow` + develop/release 키(`moai update` 후 재적용 필요, `AGENTS.local.md` §2.3)

---

Classification: Local-only operational rule (dev-only, never mirrored to the distributed template).

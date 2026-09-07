# t395 — `backlog.json` 직독 표면 전수 조사

트리: `.claude/worktrees/t395` · 브랜치 `WT-stale-backlog-json` @ `ad272be20`
`premise-verdict.md` 의 후속 — "읽는 쪽이 스테일임을 알 수 있어야 한다"(카드 범위 2)의 대상 목록.

## Claim

피해는 lane-10 한 번의 사고가 아니다. **저장소의 문서와 스킬이 아직 `backlog.json` 을 큐라고
지목하고 있고**, 그중 하나는 코드로 그 파일을 읽는다. 즉 스테일 사본을 읽으라고 시키는 지시가
저장소 안에 남아 있다.

## Evidence

### R1 — 스킬이 `backlog.json` 을 감시한다 (실동작 결함)

`.claude/skills/moai-kanban-foreman/SKILL.md:95` — 무인 foreman 반복의 "Queue watch" 가 거는
persistent Monitor:

```sh
f=.moai/state/todo/backlog.json
last=init
while true; do
  if [ -f "$f" ]; then cur=$(cksum "$f"); else cur=missing; fi
  if [ "$cur" != "$last" ]; then
    [ "$last" != init ] && echo "backlog changed"
    last=$cur
  fi
  sleep 5
done
```

마이그레이션된 프로젝트에서 이 감시는 두 경우 모두 **아무것도 잡지 못한다**:

- 파일이 없으면 `cur=missing` 이 영원히 고정 — 큐가 아무리 바뀌어도 변화 신호가 안 난다.
- 파일이 있으면(현재 이 저장소가 그렇다) 그건 정본이 아닌 08/31 스냅샷이라 **다시는 안 바뀐다** —
  역시 신호가 안 난다.

어느 쪽이든 조용히 통과한다. 이 저장소가 반복해 본 공허한 초록과 같은 모양이다.
정본은 `backlog.db` 이므로 감시 대상이 틀렸다.

### R2 — 문서가 `backlog.json` 을 상태 저장 위치로 단언한다

- `.claude/skills/moai/SKILL.md:170` — "State: `.moai/state/todo/backlog.json` — project-local,
  not committed, atomic writes."
- `.claude/skills/moai/workflows/todo.md:17` — "State lives at `.moai/state/todo/backlog.json`
  of the PRIMARY checkout"
- `.claude/skills/moai/workflows/todo.md:21` — 홈 폴백도 `~/.moai/todo/<project-key>/backlog.json`

셋 다 SQLite 전환 이후 사실이 아니다. 이것이 사람과 에이전트가 그 파일을 직독하게 만드는 지시다.

### R3 — 정확한 문서는 따로 있고, 이미 이 상태를 "예외" 로만 다룬다

`.moai/docs/todo-queue-storage.md:20` — `backlog.json` 은 "Present only if you exported one,
or if the queue has not been moved onto the database yet." 옳은 서술이지만, R2 의 세 곳과
정면으로 어긋난다. 읽는 쪽은 어느 문서를 먼저 만나느냐로 갈린다.

### R4 — 이미 있는 공시 장치 (재사용 후보)

`internal/kanban/backlog_archive_vouch.go:46-58` `InspectBacklogArchiveVouch` 는 **어느 저장소가
읽기를 답하는지**를 이미 판별해 이름으로 돌려준다(`BacklogStoreSQLite` / `BacklogStoreLegacyJSON`
/ `BacklogStoreNone`). 읽기 전용이고 마이그레이션·DDL·락 없음.
`internal/cli/todo_history.go:84,99` 가 이걸 써서 stderr 로 저장소 정체를 밝힌다 —
`absent` 가 권위 있는 답으로 오독되지 않게 하려는, **이 카드와 같은 종류의 방어**다.

즉 "읽는 쪽에 정체를 알린다"는 패턴이 이 코드베이스에 이미 있다. 새로 만들 필요가 없다.

### R5 — 같은 오지시가 배포 템플릿에도 그대로 있다

> **정정 (plan-audit D7, 2026-09-02).** 이 절의 최초 인용은 결론을 뒷받침하지 못했다.
> `grep -rn 'state/todo/backlog\.json' internal/template/templates/` 의 히트 4건을 그대로
> "결함 4곳"으로 옮겨 적었는데, 그중 `todo-queue-storage.md:55` 는 `export-json` 이
> canonical 경로에 쓴다는 **정상 서술**이고, 정작 결함인 `todo.md:21`(홈 폴백)은 경로 모양이
> 달라 그 패턴에 안 걸린다. 결론("결함 지점이 템플릿에도 미러돼 있다")은 참이지만,
> 근거는 아래로 교체한다.

결함 지점(로컬 ↔ 템플릿, 각각 실측):

| 지점 | 로컬 | 템플릿 | 성격 |
|---|---|---|---|
| foreman 감시 대상 | `.claude/skills/moai-kanban-foreman/SKILL.md:95` | 동일 경로 :95 | R1 — 죽은 감시 |
| 상태 위치 단언 | `.claude/skills/moai/SKILL.md:170` | 동일 경로 :170 | R2 |
| 상태 위치 단언(primary) | `.claude/skills/moai/workflows/todo.md:17` | 동일 경로 :17 | R2 |
| 상태 위치 단언(홈 폴백) | `.claude/skills/moai/workflows/todo.md:21` | 동일 경로 :21 | R2 |

정상이라 **건드리지 않는** 지점(대조군):

- `.moai/docs/todo-queue-storage.md:20` — "Present only if you exported one … or if the queue
  has not been moved onto the database yet." 옳다.
- `.moai/docs/todo-queue-storage.md:55` — `moai todo export-json` 이 canonical 경로에 쓴다는
  서술. 다운그레이드 경로의 설계된 동작이므로 옳다.

R1(죽은 foreman 감시)과 R2(틀린 상태 위치 단언)가 **배포판 사용자에게도 나간다.**
수리 범위는 로컬 `.claude/` 만이 아니라 템플릿 미러까지다(CLAUDE.local.md §2 Template-First).

**[HARD] 완전성 검사에 쓸 패턴 주의.** `state/todo/backlog\.json` 은 홈 폴백 줄(`:21`)을
구조적으로 못 본다 — 경로 모양이 `~/.moai/todo/<project-key>/backlog.json` 이라 매칭되지 않는다.
이 패턴을 완전성 기준으로 쓰면 `:21` 이 안 고쳐진 채로도 통과한다(plan-audit D5).

## Baseline-attribution

전부 이번 실행에서 워크트리 `ad272be20` 트리에 대해 `grep -rn` 으로 측정했다.
R1 의 동작 결론은 인용한 쉘 코드를 읽고 도출한 것이며, foreman 반복을 실제로 돌려 관측하지는 않았다.

## Gaps (미검증)

- R1 의 "신호가 안 난다"를 **실행으로 재현하지 않았다** — 코드 판독 결과다. 재현은 plan-phase 의
  acceptance 로 넘길 항목이다.
- ~~템플릿 사본 여부 미측정~~ → **닫음. 네 곳 모두 템플릿에도 있다** (아래 R5).
- 사람이 직독하는 경로(에이전트 프롬프트, 과거 리포트)는 셀 수 없으므로 목록에 넣지 않았다.

## Residual-risk

- R1 은 t395 와 **독립적으로 고쳐야 할 결함**일 수 있다(무인 foreman 이 큐 변화를 못 잡는다).
  이 카드에 접을지 별도 카드로 낼지는 리드 판단 사항 — 스스로 넓히지 않았다.
- R4 의 재사용이 맞더라도, 공시는 도구를 쓰는 읽기만 보호한다. `cat backlog.json` 하는 사람은
  여전히 못 막는다 — 파일 자체가 스스로 정체를 말하지 않는 한.
